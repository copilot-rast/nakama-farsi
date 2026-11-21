package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func (s *service) eventRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_event_record":    s.rpcEventRecord,
		"economy_events_progress": s.rpcEventsProgress,
		"economy_levels":          s.rpcLevels,
		"economy_streaks":         s.rpcStreaks,
		"economy_reward_claim":    s.rpcRewardClaim,
		"economy_rewards_mine":    s.rpcRewardsMine,
		"economy_boosters":        s.rpcBoosters,
	}
}

type prizeName struct {
	Key, Name, Message string
}

func (s *service) prizeNames(ctx context.Context, langs *localized.Set, lang string, ids []string) (map[string]prizeName, error) {
	out := map[string]prizeName{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, key, name::text, message::text FROM economy_prize WHERE id::text = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id, key, name, msg string
		if err := rows.Scan(&id, &key, &name, &msg); err != nil {
			return nil, err
		}
		out[id] = prizeName{Key: key, Name: langs.Resolve(localized.Parse(name), lang), Message: langs.Resolve(localized.Parse(msg), lang)}
	}
	return out, rows.Err()
}

func prizeJSON(p prizeName, id string) map[string]any {
	return map[string]any{"id": id, "key": p.Key, "name": p.Name, "message": p.Message}
}

func (s *service) rpcEventRecord(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key      string         `json:"key"`
		Value    *float64       `json:"value"`
		Metadata map[string]any `json:"metadata"`
		Lang     string         `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	out, err := s.recordEvent(ctx, eventInput{UserID: userID, Key: in.Key, Value: in.Value, Metadata: in.Metadata, Source: "client"})
	if err != nil {
		return nil, err
	}
	return s.describeOutcome(ctx, userID, in.Lang, out)
}

func (s *service) describeOutcome(ctx context.Context, userID, lang string, out *eventOutcome) (map[string]any, error) {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, lang)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for _, c := range out.Claims {
		ids = append(ids, c.PrizeID)
	}
	names, err := s.prizeNames(ctx, langs, lang, ids)
	if err != nil {
		return nil, err
	}
	claims := []map[string]any{}
	for _, c := range out.Claims {
		claims = append(claims, map[string]any{"claimId": c.ClaimID, "kind": c.Kind, "grantMode": c.Mode, "status": c.Status, "prize": prizeJSON(names[c.PrizeID], c.PrizeID)})
	}
	visible, err := s.visibleMetaKeys(ctx)
	if err != nil {
		return nil, err
	}
	metas := map[string]float64{}
	for k, v := range out.Metas {
		if visible[k] {
			metas[k] = v
		}
	}
	return map[string]any{"lang": lang, "eventId": out.EventID, "key": out.Key, "value": out.Value, "rules": out.Rules, "metas": metas, "streaks": out.Streaks, "levels": out.Levels, "relations": out.Relations, "claims": claims}, nil
}

func (s *service) visibleMetaKeys(ctx context.Context) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key FROM economy_meta WHERE NOT server_only`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		out[k] = true
	}
	return out, rows.Err()
}

type claimRow struct {
	ID, Kind, OwnerID, RefID, Scope, PrizeID, Mode, Status string
	Value                                                  float64
	Error                                                  sql.NullString
	CreatedAt                                              time.Time
	GrantedAt                                              sql.NullTime
	ExpiresAt                                              sql.NullTime
}

func (c *claimRow) expired(now time.Time) bool {
	return c.Status != "granted" && c.ExpiresAt.Valid && !c.ExpiresAt.Time.After(now)
}

func (s *service) userClaims(ctx context.Context, userID, kind string, owners []string) ([]claimRow, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, kind, owner_id::text, ref_id::text, scope, prize_id::text, grant_mode, status, COALESCE(value,0)::float8, error, created_at, granted_at, expires_at
		FROM economy_reward_claim WHERE user_id=$1 AND kind=$2 AND owner_id::text = ANY($3) ORDER BY created_at`, userID, kind, owners)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []claimRow
	for rows.Next() {
		var c claimRow
		if err := rows.Scan(&c.ID, &c.Kind, &c.OwnerID, &c.RefID, &c.Scope, &c.PrizeID, &c.Mode, &c.Status, &c.Value, &c.Error, &c.CreatedAt, &c.GrantedAt, &c.ExpiresAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *service) rpcEventsProgress(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	settings, err := loadEventSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	clk := settings.clock()
	now := s.now()
	tags, err := s.userTagIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT r.id::text, r.key, r.name::text, r.description::text, r.measure, r.period, r.threshold::float8, r.repeat, COALESCE(r.max_claims,0), r.grant_mode,
			r.audience_tag_ids::text, r.ends_at, et.key, et.name::text, r.prize_id::text
		FROM economy_event_rule r JOIN economy_event_type et ON et.id=r.event_type_id
		WHERE r.is_active AND et.is_active AND (r.starts_at IS NULL OR r.starts_at <= $1) AND (r.ends_at IS NULL OR r.ends_at > $1)
		ORDER BY r.sort, r.created_at, r.id`, now)
	if err != nil {
		return nil, err
	}
	type item struct {
		def                             ruleDef
		name, desc, eventKey, eventName string
		endsAt                          sql.NullTime
	}
	var list []item
	for rows.Next() {
		var it item
		var audience string
		if err := rows.Scan(&it.def.ID, &it.def.Key, &it.name, &it.desc, &it.def.Measure, &it.def.Period, &it.def.Threshold, &it.def.Repeat, &it.def.MaxClaims, &it.def.GrantMode,
			&audience, &it.endsAt, &it.eventKey, &it.eventName, &it.def.PrizeID); err != nil {
			rows.Close()
			return nil, err
		}
		it.def.Audience = parseTextArray(audience)
		if ok, _ := audienceAllows(it.def.Audience, func() (map[string]bool, error) { return tags, nil }); !ok {
			continue
		}
		list = append(list, it)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ruleIDs := []string{}
	prizeIDs := []string{}
	for _, it := range list {
		ruleIDs = append(ruleIDs, it.def.ID)
		prizeIDs = append(prizeIDs, it.def.PrizeID)
	}
	type prog struct {
		count    int64
		sum, max float64
		steps    int
	}
	progress := map[string]prog{}
	if len(ruleIDs) > 0 {
		prow, err := s.db.QueryContext(ctx, `SELECT rule_id::text, period_start::text, count, sum::float8, COALESCE(max,0)::float8, steps FROM economy_event_rule_progress WHERE user_id=$1 AND rule_id::text = ANY($2)`, userID, ruleIDs)
		if err != nil {
			return nil, err
		}
		for prow.Next() {
			var rid, start string
			var p prog
			if err := prow.Scan(&rid, &start, &p.count, &p.sum, &p.max, &p.steps); err != nil {
				prow.Close()
				return nil, err
			}
			progress[rid+"|"+start[:10]] = p
		}
		prow.Close()
	}
	claims, err := s.userClaims(ctx, userID, "event", ruleIDs)
	if err != nil {
		return nil, err
	}
	names, err := s.prizeNames(ctx, langs, lang, prizeIDs)
	if err != nil {
		return nil, err
	}
	items := []map[string]any{}
	ready := []map[string]any{}
	for _, it := range list {
		start := dateKey(clk.periodStart(it.def.Period, now))
		p := progress[it.def.ID+"|"+start]
		current := ruleMeasure(it.def.Measure, p.count, p.sum, p.max)
		total := 0
		claimable := []string{}
		for _, c := range claims {
			if c.RefID != it.def.ID {
				continue
			}
			total++
			if c.Status != "granted" && c.Mode == "claim" {
				claimable = append(claimable, c.ID)
				ready = append(ready, map[string]any{"claimId": c.ID, "kind": "event", "source": it.def.Key, "sourceName": langs.Resolve(localized.Parse(it.name), lang), "prize": prizeJSON(names[c.PrizeID], c.PrizeID), "createdAt": c.CreatedAt})
			}
		}
		done := false
		switch it.def.Repeat {
		case "once":
			done = total > 0
		case "period":
			done = p.steps > 0
		}
		if it.def.MaxClaims > 0 && total >= it.def.MaxClaims {
			done = true
		}
		target := it.def.Threshold
		if it.def.Repeat == "every" {
			target = it.def.Threshold * float64(p.steps+1)
		}
		row := map[string]any{
			"key": it.def.Key, "name": langs.Resolve(localized.Parse(it.name), lang), "description": langs.Resolve(localized.Parse(it.desc), lang),
			"eventKey": it.eventKey, "eventName": langs.Resolve(localized.Parse(it.eventName), lang),
			"measure": it.def.Measure, "period": it.def.Period, "repeat": it.def.Repeat, "threshold": it.def.Threshold, "target": target,
			"current": current, "claims": total, "maxClaims": it.def.MaxClaims, "grantMode": it.def.GrantMode, "claimable": claimable, "done": done,
			"periodEndsAt": clk.periodEnd(it.def.Period, now), "prize": prizeJSON(names[it.def.PrizeID], it.def.PrizeID),
		}
		if it.endsAt.Valid {
			row["endsAt"] = it.endsAt.Time
		}
		items = append(items, row)
	}
	return map[string]any{"lang": lang, "timezone": settings.Timezone, "items": items, "ready": ready}, nil
}

func (s *service) rpcLevels(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	items, err := s.levelViews(ctx, userID, langs, lang, true)
	if err != nil {
		return nil, err
	}
	return map[string]any{"lang": lang, "items": items}, nil
}

func (s *service) levelViews(ctx context.Context, userID string, langs *localized.Set, lang string, publicOnly bool) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT t.id::text, t.key, t.name::text, t.description::text, m.key, m.name::text, m.server_only,
			COALESCE((SELECT value FROM economy_user_meta WHERE user_id=$1 AND meta_id=m.id), m.default_value)::float8,
			COALESCE((SELECT level FROM economy_user_level WHERE user_id=$1 AND table_id=t.id), 0)
		FROM economy_level_table t JOIN economy_meta m ON m.id=t.meta_id WHERE t.is_active ORDER BY t.sort, t.created_at, t.id`, userID)
	if err != nil {
		return nil, err
	}
	type table struct {
		id, key, name, desc, metaKey, metaName string
		serverOnly                             bool
		value                                  float64
		level                                  int
	}
	var tables []table
	for rows.Next() {
		var t table
		if err := rows.Scan(&t.id, &t.key, &t.name, &t.desc, &t.metaKey, &t.metaName, &t.serverOnly, &t.value, &t.level); err != nil {
			rows.Close()
			return nil, err
		}
		if publicOnly && t.serverOnly {
			continue
		}
		tables = append(tables, t)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ids := []string{}
	for _, t := range tables {
		ids = append(ids, t.id)
	}
	claims, err := s.userClaims(ctx, userID, "level", ids)
	if err != nil {
		return nil, err
	}
	out := []map[string]any{}
	for _, t := range tables {
		srows, err := s.db.QueryContext(ctx, `SELECT id::text, level, threshold::float8, name::text, COALESCE(prize_id::text,''), grant_mode FROM economy_level_step WHERE table_id=$1 ORDER BY level`, t.id)
		if err != nil {
			return nil, err
		}
		type step struct {
			id, name, prize, mode string
			level                 int
			threshold             float64
		}
		var steps []step
		prizeIDs := []string{}
		for srows.Next() {
			var st step
			if err := srows.Scan(&st.id, &st.level, &st.threshold, &st.name, &st.prize, &st.mode); err != nil {
				srows.Close()
				return nil, err
			}
			steps = append(steps, st)
			if st.prize != "" {
				prizeIDs = append(prizeIDs, st.prize)
			}
		}
		srows.Close()
		names, err := s.prizeNames(ctx, langs, lang, prizeIDs)
		if err != nil {
			return nil, err
		}
		if t.level == 0 {
			for _, st := range steps {
				if st.threshold <= 0 && st.level > t.level {
					t.level = st.level
				}
			}
		}
		list := []map[string]any{}
		var next map[string]any
		var curTitle string
		curThreshold := 0.0
		for _, st := range steps {
			status := "locked"
			var claimID string
			if st.level <= t.level {
				status = "reached"
				for _, c := range claims {
					if c.RefID == st.id {
						claimID = c.ID
						if c.Status == "granted" {
							status = "claimed"
						} else if c.Mode == "claim" {
							status = "claimable"
						} else {
							status = "pending"
						}
					}
				}
			}
			row := map[string]any{"level": st.level, "threshold": st.threshold, "title": langs.Resolve(localized.Parse(st.name), lang), "grantMode": st.mode, "status": status}
			if claimID != "" {
				row["claimId"] = claimID
			}
			if st.prize != "" {
				row["prize"] = prizeJSON(names[st.prize], st.prize)
			}
			if st.level == t.level {
				curTitle = row["title"].(string)
				curThreshold = st.threshold
			}
			if st.level > t.level && next == nil {
				next = row
			}
			list = append(list, row)
		}
		item := map[string]any{"key": t.key, "name": langs.Resolve(localized.Parse(t.name), lang), "description": langs.Resolve(localized.Parse(t.desc), lang),
			"metaKey": t.metaKey, "metaName": langs.Resolve(localized.Parse(t.metaName), lang), "value": t.value, "level": t.level, "title": curTitle,
			"levelThreshold": curThreshold, "next": next, "steps": list}
		out = append(out, item)
	}
	return out, nil
}

func (s *service) rpcStreaks(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	items, err := s.streakViews(ctx, userID, langs, lang, true)
	if err != nil {
		return nil, err
	}
	return map[string]any{"lang": lang, "items": items}, nil
}

func (s *service) streakViews(ctx context.Context, userID string, langs *localized.Set, lang string, activeOnly bool) ([]map[string]any, error) {
	settings, err := loadEventSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	clk := settings.clock()
	now := s.now()
	q := `SELECT st.id::text, st.key, st.name::text, st.description::text, st.period_days, st.grace_periods, st.miss_action, COALESCE(st.cycle_length,0), st.is_active, et.key, st.repay_after_fall, st.week_length, st.infinity, st.infinity_currency, st.infinity_min, st.infinity_max,
			st.day_zone, st.day_start_minute, st.signup_delay_days, st.claim_window, st.infinity_mode, COALESCE(st.infinity_prize_id::text,''),
			COALESCE(u.count,0), COALESCE(u.best,0), COALESCE(u.run,1), COALESCE(u.grace_used,0), COALESCE(u.peak,0), COALESCE(u.falls,0), u.last_period, u.updated_at
		FROM economy_streak st JOIN economy_event_type et ON et.id=st.event_type_id LEFT JOIN economy_user_streak u ON u.streak_id=st.id AND u.user_id=$1`
	if activeOnly {
		q += ` WHERE st.is_active`
	}
	rows, err := s.db.QueryContext(ctx, q+` ORDER BY st.sort, st.created_at, st.id`, userID)
	if err != nil {
		return nil, err
	}
	type streak struct {
		id, key, name, desc, miss, eventKey string
		days, grace, cycle                  int
		active, repay                       bool
		week                                int
		inf                                 bool
		infCur                              string
		infMin, infMax                      int64
		zone, window, infMode, infPrize     string
		dayStart, delay                     int
		st                                  streakState
		updated                             sql.NullTime
	}
	var list []streak
	for rows.Next() {
		var sk streak
		var last sql.NullInt64
		if err := rows.Scan(&sk.id, &sk.key, &sk.name, &sk.desc, &sk.days, &sk.grace, &sk.miss, &sk.cycle, &sk.active, &sk.eventKey, &sk.repay, &sk.week, &sk.inf, &sk.infCur, &sk.infMin, &sk.infMax,
			&sk.zone, &sk.dayStart, &sk.delay, &sk.window, &sk.infMode, &sk.infPrize,
			&sk.st.Count, &sk.st.Best, &sk.st.Run, &sk.st.GraceUsed, &sk.st.Peak, &sk.st.Falls, &last, &sk.updated); err != nil {
			rows.Close()
			return nil, err
		}
		if last.Valid {
			v := last.Int64
			sk.st.Last = &v
		}
		list = append(list, sk)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ids := []string{}
	for _, sk := range list {
		ids = append(ids, sk.id)
	}
	claims, err := s.userClaims(ctx, userID, "streak", ids)
	if err != nil {
		return nil, err
	}
	milestones, err := loadStreakMilestones(ctx, s.db, ids)
	if err != nil {
		return nil, err
	}
	extras, err := loadStreakExtras(ctx, s.db, ids)
	if err != nil {
		return nil, err
	}
	out := []map[string]any{}
	for _, sk := range list {
		rrows, err := s.db.QueryContext(ctx, `SELECT id::text, day, prize_id::text, grant_mode FROM economy_streak_reward WHERE streak_id=$1 ORDER BY day`, sk.id)
		if err != nil {
			return nil, err
		}
		type reward struct {
			id, prize, mode string
			day             int
		}
		var rewards []reward
		prizeIDs := []string{}
		for rrows.Next() {
			var r reward
			if err := rrows.Scan(&r.id, &r.day, &r.prize, &r.mode); err != nil {
				rrows.Close()
				return nil, err
			}
			rewards = append(rewards, r)
			prizeIDs = append(prizeIDs, r.prize)
		}
		rrows.Close()
		ms := milestones[sk.id]
		for _, m := range ms {
			if m.prize != "" {
				prizeIDs = append(prizeIDs, m.prize)
			}
		}
		for _, x := range extras[sk.id] {
			prizeIDs = append(prizeIDs, x.prize)
		}
		names, err := s.prizeNames(ctx, langs, lang, prizeIDs)
		if err != nil {
			return nil, err
		}
		sclk := clk.withDay(sk.zone, sk.dayStart)
		idx := sclk.streakIndex(now, sk.days)
		rule := streakRule{Grace: sk.grace, MissAction: sk.miss, Milestones: milestoneDays(ms), WeekLength: sk.week, Infinity: streakInfinityOn(sk.inf, sk.infMode, sk.infCur, sk.infPrize), Track: sk.cycle}
		broken := streakBroken(sk.st, idx, rule)
		count := sk.st.Count
		run := sk.st.Run
		peak := sk.st.Peak
		falls := sk.st.Falls
		floored := false
		if broken {
			missed, left := streakMissed(sk.st, idx, rule)
			c, fl := streakFallCount(sk.st, missed-left, rule)
			count, floored = c, fl
			if sk.miss == "step_back" || sk.miss == "week_start" || fl {
				falls++
			} else {
				run++
				peak = 0
				falls = 0
			}
		}
		floor := streakFloor(peak, rule.Milestones)
		counted := !broken && sk.st.Last != nil && *sk.st.Last == idx
		graceLeft := sk.grace - sk.st.GraceUsed
		if broken || graceLeft < 0 {
			graceLeft = sk.grace
			if sk.miss == "step_back" && broken {
				graceLeft = 0
			}
		}
		var breaksAt *time.Time
		if count > 0 && sk.st.Last != nil && !broken {
			b := sclk.streakPeriodStart(*sk.st.Last+int64(2+graceLeft), sk.days)
			breaksAt = &b
		}
		nextAt := sclk.streakPeriodStart(idx+1, sk.days)
		length := sk.cycle
		if length == 0 {
			for _, r := range rewards {
				if r.day > length {
					length = r.day
				}
			}
		}
		base := 0
		if sk.cycle > 0 {
			if counted && count > 0 {
				base = ((count - 1) / sk.cycle) * sk.cycle
			} else {
				base = (count / sk.cycle) * sk.cycle
			}
		}
		byDay := map[int]reward{}
		for _, r := range rewards {
			byDay[r.day] = r
		}
		days := []map[string]any{}
		claimable := []string{}
		for d := 1; d <= length; d++ {
			pos := base + d
			if sk.cycle == 0 {
				pos = d
			}
			status := "locked"
			if pos <= count {
				status = "done"
			} else if pos == count+1 && !counted {
				status = "today"
			}
			row := map[string]any{"day": d, "status": status}
			if r, ok := byDay[d]; ok {
				row["prize"] = prizeJSON(names[r.prize], r.prize)
				row["grantMode"] = r.mode
				scopes := streakDayScopes(run, falls, pos, sk.repay)
				var found *claimRow
				for _, scope := range scopes {
					for i := range claims {
						if found == nil && claims[i].RefID == r.id && claims[i].Scope == scope {
							found = &claims[i]
						}
					}
				}
				if c := found; c != nil {
					row["claimId"] = c.ID
					row["status"] = claimStatusWord(c, s.now())
				}
			}
			days = append(days, row)
		}
		for i := range claims {
			c := &claims[i]
			if c.OwnerID == sk.id && c.Status != "granted" && c.Mode == "claim" && !c.expired(now) {
				claimable = append(claimable, c.ID)
			}
		}
		mlist := []map[string]any{}
		for _, m := range ms {
			row := map[string]any{"day": m.day, "reached": m.day <= peak, "safe": m.day == floor && floor > 0}
			status := "locked"
			if m.day <= peak {
				status = "reached"
			}
			if m.prize != "" {
				row["prize"] = prizeJSON(names[m.prize], m.prize)
				row["grantMode"] = m.mode
				for i := range claims {
					c := &claims[i]
					if c.RefID == sk.id && c.Scope == streakMilestoneScope(m.day) {
						row["claimId"] = c.ID
						status = claimStatusWord(c, s.now())
					}
				}
			}
			row["status"] = status
			mlist = append(mlist, row)
		}
		var nextMilestone any
		if n := streakNextMilestone(peak, rule.Milestones); n > 0 {
			nextMilestone = n
		}
		xlist := []map[string]any{}
		for _, x := range extras[sk.id] {
			row := map[string]any{"day": x.day, "prize": prizeJSON(names[x.prize], x.prize), "grantMode": x.mode, "status": "locked"}
			for i := range claims {
				c := &claims[i]
				if c.RefID == x.id {
					row["claimId"] = c.ID
					row["status"] = claimStatusWord(c, s.now())
				}
			}
			xlist = append(xlist, row)
		}
		phase := "track"
		if rule.Infinity && sk.cycle > 0 && count > sk.cycle {
			phase = "infinity"
		}
		var infinity any
		if rule.Infinity {
			inf := map[string]any{"currencyKey": sk.infCur, "min": sk.infMin, "max": sk.infMax, "mode": sk.infMode}
			if sk.infMode == "claim" {
				inf["prizeId"] = sk.infPrize
				for i := range claims {
					c := &claims[i]
					if c.RefID == sk.id && strings.HasPrefix(c.Scope, "inf:") && sk.st.Last != nil && *sk.st.Last == idx && c.Scope == "inf:r"+strconv.Itoa(sk.st.Run)+":"+strconv.Itoa(sk.st.Count) {
						inf["claimId"], inf["status"] = c.ID, claimStatusWord(c, now)
					}
				}
			}
			infinity = inf
		}
		item := map[string]any{"key": sk.key, "milestones": mlist, "floor": floor, "nextMilestone": nextMilestone, "floored": floored, "peak": peak, "repayAfterFall": sk.repay, "name": langs.Resolve(localized.Parse(sk.name), lang), "description": langs.Resolve(localized.Parse(sk.desc), lang),
			"eventKey": sk.eventKey, "periodDays": sk.days, "count": count, "best": sk.st.Best, "run": run, "countedThisPeriod": counted, "broken": broken,
			"graceLeft": graceLeft, "gracePeriods": sk.grace, "missAction": sk.miss, "cycleLength": sk.cycle, "nextPeriodAt": nextAt, "breaksAt": breaksAt,
			"days": days, "claimable": claimable, "isActive": sk.active, "claimWindow": sk.window, "signupDelayDays": sk.delay, "dayZone": sk.zone, "dayStartMinute": sk.dayStart, "extras": xlist, "phase": phase, "infinity": infinity, "weekLength": sk.week}
		if !activeOnly {
			item["id"] = sk.id
			if sk.updated.Valid {
				item["updatedAt"] = sk.updated.Time
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *service) rpcRewardClaim(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ClaimID string `json:"claimId"`
		Lang    string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if _, err := uuid.FromString(in.ClaimID); err != nil {
		return nil, notFound("CLAIM_NOT_FOUND")
	}
	if taskOwner, isTask, err := s.taskClaimOwner(ctx, in.ClaimID); err != nil {
		return nil, err
	} else if isTask {
		if taskOwner != userID {
			return nil, notFound("CLAIM_NOT_FOUND")
		}
		res, err := s.claimTask(ctx, userID, "", in.ClaimID, "player")
		if err != nil {
			return nil, err
		}
		return s.taskClaimResponse(ctx, userID, in.Lang, res)
	}
	var owner, status, prizeID, kind string
	var expires sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT user_id::text, status, prize_id::text, kind, expires_at FROM economy_reward_claim WHERE id=$1`, in.ClaimID).Scan(&owner, &status, &prizeID, &kind, &expires)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && owner != userID) {
		return nil, notFound("CLAIM_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	if status == "granted" {
		return nil, conflict("ALREADY_CLAIMED", "this reward was already received")
	}
	if expires.Valid && !expires.Time.After(s.now()) {
		e := fail(http.StatusGone, "CLAIM_EXPIRED", "", "this reward could only be received until the day ended")
		e.Meta = map[string]any{"expiredAt": expires.Time.UTC().Format(time.RFC3339)}
		return nil, e
	}
	extra, err := s.grantClaim(ctx, in.ClaimID, 1)
	if err != nil {
		return nil, err
	}
	rewards, err := s.claimRewards(ctx, in.ClaimID)
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	names, err := s.prizeNames(ctx, langs, lang, []string{prizeID})
	if err != nil {
		return nil, err
	}
	res := map[string]any{"lang": lang, "claimId": in.ClaimID, "kind": kind, "rewards": rewards, "prize": prizeJSON(names[prizeID], prizeID), "levels": []levelChange{}}
	if extra != nil {
		res["levels"] = extra.Levels
	}
	return res, nil
}

func (s *service) rpcRewardsMine(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Status string `json:"status"`
		Limit  int    `json:"limit"`
		Lang   string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if in.Limit < 1 || in.Limit > 100 {
		in.Limit = 30
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	items, err := s.claimViews(ctx, userID, in.Status, in.Limit, langs, lang)
	if err != nil {
		return nil, err
	}
	return map[string]any{"lang": lang, "items": items}, nil
}

func (s *service) claimViews(ctx context.Context, userID, status string, limit int, langs *localized.Set, lang string) ([]map[string]any, error) {
	where := &sqlWhere{}
	where.add("c.user_id = ?::uuid", userID)
	if status == "ready" {
		where.add("c.status <> 'granted'")
		where.add("(c.expires_at IS NULL OR c.expires_at > now())")
	} else if status != "" {
		where.add("c.status = ?", status)
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(ctx, `SELECT c.id::text, c.kind, c.scope, c.grant_mode, c.status, COALESCE(c.value,0)::float8, COALESCE(c.error,''), c.created_at, c.granted_at, c.prize_id::text, c.expires_at,
			COALESCE(r.key, lt.key, st.key, tk.key, ac.key, cg.key, mr.key, ''), COALESCE(r.name, lt.name, st.name, tk.name, ac.name, cg.name, mr.name, '{}'::jsonb)::text, COALESCE(ls.level, at.tier, 0)
		FROM economy_reward_claim c
		LEFT JOIN economy_event_rule r ON c.kind='event' AND r.id=c.owner_id
		LEFT JOIN economy_level_table lt ON c.kind='level' AND lt.id=c.owner_id
		LEFT JOIN economy_level_step ls ON c.kind='level' AND ls.id=c.ref_id
		LEFT JOIN economy_streak st ON c.kind='streak' AND st.id=c.owner_id
		LEFT JOIN economy_task tk ON c.kind='task' AND tk.id=c.owner_id
		LEFT JOIN economy_achievement ac ON c.kind='achievement' AND ac.id=c.owner_id
		LEFT JOIN economy_achievement_tier at ON c.kind='achievement' AND at.id=c.ref_id
		LEFT JOIN economy_chance_game cg ON c.kind='chance' AND cg.id=c.owner_id
		LEFT JOIN economy_meta_relation mr ON c.kind='relation' AND mr.id=c.owner_id`+where.sql()+` ORDER BY c.created_at DESC, c.id DESC LIMIT $`+itoa(n+1), append(where.args, limit)...)
	if err != nil {
		return nil, err
	}
	type row struct {
		id, kind, scope, mode, status, errText, prize, srcKey, srcName string
		value                                                          float64
		created                                                        time.Time
		granted, expires                                               sql.NullTime
		level                                                          int
	}
	var list []row
	prizeIDs := []string{}
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.kind, &r.scope, &r.mode, &r.status, &r.value, &r.errText, &r.created, &r.granted, &r.prize, &r.expires, &r.srcKey, &r.srcName, &r.level); err != nil {
			rows.Close()
			return nil, err
		}
		list = append(list, r)
		prizeIDs = append(prizeIDs, r.prize)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	names, err := s.prizeNames(ctx, langs, lang, prizeIDs)
	if err != nil {
		return nil, err
	}
	out := []map[string]any{}
	for _, r := range list {
		item := map[string]any{"claimId": r.id, "kind": r.kind, "scope": r.scope, "grantMode": r.mode, "status": r.status, "value": r.value,
			"source": r.srcKey, "sourceName": langs.Resolve(localized.Parse(r.srcName), lang), "prize": prizeJSON(names[r.prize], r.prize), "createdAt": r.created}
		if r.granted.Valid {
			item["grantedAt"] = r.granted.Time
		}
		if r.expires.Valid {
			item["expiresAt"] = r.expires.Time
			if r.status != "granted" && !r.expires.Time.After(s.now()) {
				item["status"] = "expired"
			}
		}
		if r.errText != "" {
			item["error"] = r.errText
		}
		if r.level > 0 {
			item["level"] = r.level
		}
		out = append(out, item)
	}
	return out, nil
}

func claimStatusWord(c *claimRow, now time.Time) string {
	if c.Status == "granted" {
		return "claimed"
	}
	if c.expired(now) {
		return "expired"
	}
	if c.Mode == "claim" {
		return "claimable"
	}
	return "pending"
}
