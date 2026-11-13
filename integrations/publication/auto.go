package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

var (
	autoTriggers      = []string{"event", "streak_at_risk", "chance_ready", "inactive"}
	autoChannels      = []string{"push", "in_app", "push_in_app"}
	autoFieldPattern  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,39}$`)
	autoVarPattern    = regexp.MustCompile(`\{([A-Za-z][A-Za-z0-9_]{0,39})\}`)
	autoEventOverlap  = 2 * time.Minute
	autoFireBatch     = 500
	autoEventBatch    = 5000
	autoReadyLookback = 24 * time.Hour
)

type autoRule struct {
	ID         string
	Key        string
	Name       string
	Trigger    string
	EventKey   string
	MatchField string
	MatchValue string
	Params     autoParams
	Channel    string
	CategoryID string
	Title      i18n
	Body       i18n
	Image      string
	Action     json.RawMessage
	Cooldown   int
	QuietStart *int
	QuietEnd   *int
	CreatedAt  time.Time
}

type autoParams struct {
	StreakKey   string `json:"streakKey,omitempty"`
	HoursBefore int    `json:"hoursBefore,omitempty"`
	GameKey     string `json:"gameKey,omitempty"`
	Days        int    `json:"days,omitempty"`
}

const autoRuleColumns = `id::text, key, name, trigger, event_key, match_field, match_value, params::text, channel, category_id::text, title::text, body::text, image_url, COALESCE(action::text,'null'),
	cooldown_minutes, quiet_start, quiet_end, created_at`

func scanAutoRule(row rowScanner) (autoRule, error) {
	var r autoRule
	var params, title, body, act string
	var qs, qe sql.NullInt64
	if err := row.Scan(&r.ID, &r.Key, &r.Name, &r.Trigger, &r.EventKey, &r.MatchField, &r.MatchValue, &params, &r.Channel, &r.CategoryID, &title, &body, &r.Image, &act,
		&r.Cooldown, &qs, &qe, &r.CreatedAt); err != nil {
		return r, err
	}
	_ = json.Unmarshal([]byte(params), &r.Params)
	r.Title, r.Body = parseI18n(title), parseI18n(body)
	r.Action = json.RawMessage(act)
	r.QuietStart, r.QuietEnd = nullInt(qs), nullInt(qe)
	return r, nil
}

func (s *service) activeAutoRules(ctx context.Context) (map[string]autoRule, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+autoRuleColumns+` FROM publication_auto_rule WHERE is_active ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]autoRule{}
	for rows.Next() {
		r, err := scanAutoRule(rows)
		if err != nil {
			return nil, err
		}
		out[r.ID] = r
	}
	return out, rows.Err()
}

func renderVars(text string, vars map[string]string) string {
	if len(vars) == 0 || !strings.Contains(text, "{") {
		return text
	}
	return autoVarPattern.ReplaceAllStringFunc(text, func(m string) string {
		if v, ok := vars[m[1:len(m)-1]]; ok {
			return v
		}
		return m
	})
}

func scalarVars(meta map[string]any, into map[string]string) {
	for k, v := range meta {
		if !autoFieldPattern.MatchString(k) || len(into) >= 30 {
			continue
		}
		switch x := v.(type) {
		case string:
			into[k] = truncate(x, 80)
		case float64:
			into[k] = strconv.FormatFloat(x, 'f', -1, 64)
		case bool:
			into[k] = strconv.FormatBool(x)
		}
	}
}

func inQuiet(r autoRule, local time.Time) bool {
	if r.QuietStart == nil || r.QuietEnd == nil {
		return false
	}
	m := local.Hour()*60 + local.Minute()
	if *r.QuietStart < *r.QuietEnd {
		return m >= *r.QuietStart && m < *r.QuietEnd
	}
	return m >= *r.QuietStart || m < *r.QuietEnd
}

func quietEnds(r autoRule, now time.Time, zone *time.Location) time.Time {
	local := now.In(zone)
	end := time.Date(local.Year(), local.Month(), local.Day(), *r.QuietEnd/60, *r.QuietEnd%60, 0, 0, zone)
	if !end.After(local) {
		end = end.AddDate(0, 0, 1)
	}
	return end
}

func (s *service) autoLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.autoEvery)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ok, err := clusterkit.AcquireLease(ctx, s.db, "publication-auto", holder, s.cfg.autoEvery-time.Second)
			if err != nil || !ok {
				continue
			}
			if err := s.runAuto(ctx); err != nil && ctx.Err() == nil {
				s.logger.Error("Publication automatic notifications failed: %v", err)
			}
		}
	}
}

func (s *service) runAuto(ctx context.Context) error {
	rules, err := s.activeAutoRules(ctx)
	if err != nil {
		return err
	}
	if len(rules) > 0 {
		if err := s.scanAutoTriggers(ctx, rules); err != nil {
			return err
		}
	}
	for i := 0; i < 20; i++ {
		n, err := s.deliverFires(ctx)
		if err != nil {
			return err
		}
		if n < autoFireBatch {
			break
		}
	}
	return nil
}

func (s *service) scanAutoTriggers(ctx context.Context, rules map[string]autoRule) error {
	now := s.now()
	zone := projectZone(ctx, s.db)
	byTrigger := map[string][]autoRule{}
	for _, r := range rules {
		byTrigger[r.Trigger] = append(byTrigger[r.Trigger], r)
	}
	if len(byTrigger["event"]) > 0 {
		if err := s.scanAutoEvents(ctx); err != nil {
			return err
		}
	}
	if len(byTrigger["streak_at_risk"])+len(byTrigger["chance_ready"])+len(byTrigger["inactive"]) == 0 || (s.timeScan != nil && !s.timeScan.due(now, time.Minute)) {
		return nil
	}
	for _, r := range byTrigger["streak_at_risk"] {
		if err := s.scanStreakRisk(ctx, r, now, zone); err != nil {
			return err
		}
	}
	for _, r := range byTrigger["chance_ready"] {
		if err := s.scanChanceReady(ctx, r, now); err != nil {
			return err
		}
	}
	for _, r := range byTrigger["inactive"] {
		if err := s.scanInactive(ctx, r, now); err != nil {
			return err
		}
	}
	return nil
}

type fireRow struct {
	RuleID, UserID, Occurrence string
	Vars                       map[string]string
}

func (s *service) insertFires(ctx context.Context, fires []fireRow) error {
	if len(fires) == 0 {
		return nil
	}
	type rec struct {
		R string            `json:"r"`
		U string            `json:"u"`
		O string            `json:"o"`
		V map[string]string `json:"v"`
	}
	recs := make([]rec, 0, len(fires))
	for _, f := range fires {
		recs = append(recs, rec{R: f.RuleID, U: f.UserID, O: f.Occurrence, V: f.Vars})
	}
	raw, _ := json.Marshal(recs)
	_, err := s.db.ExecContext(ctx, `INSERT INTO publication_auto_fire (rule_id, user_id, occurrence, vars)
		SELECT x.r, x.u, x.o, COALESCE(x.v, '{}'::jsonb) FROM jsonb_to_recordset($1::jsonb) AS x(r uuid, u uuid, o text, v jsonb)
		WHERE EXISTS (SELECT 1 FROM users WHERE id = x.u) ON CONFLICT DO NOTHING`, string(raw))
	return err
}

func (s *service) scanAutoEvents(ctx context.Context) error {
	var cursor time.Time
	if err := s.db.QueryRowContext(ctx, `SELECT event_cursor FROM publication_auto_state WHERE id=1`).Scan(&cursor); err != nil {
		return err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT r.id::text, e.id::text, e.user_id::text, e.value::float8, e.metadata::text, e.created_at
		FROM economy_player_event e JOIN economy_event_type t ON t.id=e.type_id
		JOIN publication_auto_rule r ON r.is_active AND r.trigger='event' AND r.event_key=t.key
		WHERE e.created_at > $1 AND e.created_at >= r.created_at
			AND (r.match_field = '' OR e.metadata->>r.match_field = r.match_value)
		ORDER BY e.created_at, e.id LIMIT $2`, cursor.Add(-autoEventOverlap), autoEventBatch)
	if err != nil {
		return err
	}
	fires := []fireRow{}
	last := cursor
	for rows.Next() {
		var ruleID, eventID, userID, meta string
		var value float64
		var at time.Time
		if err := rows.Scan(&ruleID, &eventID, &userID, &value, &meta, &at); err != nil {
			rows.Close()
			return err
		}
		vars := map[string]string{}
		var md map[string]any
		_ = json.Unmarshal([]byte(meta), &md)
		scalarVars(md, vars)
		vars["value"] = strconv.FormatFloat(value, 'f', -1, 64)
		fires = append(fires, fireRow{RuleID: ruleID, UserID: userID, Occurrence: "ev:" + eventID, Vars: vars})
		if at.After(last) {
			last = at
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if err := s.insertFires(ctx, fires); err != nil {
		return err
	}
	if len(fires) >= autoEventBatch {
		_, err = s.db.ExecContext(ctx, `UPDATE publication_auto_state SET event_cursor=$1 + $2::interval, updated_at=now() WHERE id=1`, last, autoEventOverlap.String())
		return err
	}
	_, err = s.db.ExecContext(ctx, `UPDATE publication_auto_state SET event_cursor=GREATEST(event_cursor, $1), updated_at=now() WHERE id=1`, last)
	return err
}

func civilDayIndex(t time.Time, zone *time.Location) int64 {
	y, m, d := t.In(zone).Date()
	return floorDivide(time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix(), 86400)
}

func floorDivide(a, b int64) int64 {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

func (s *service) scanStreakRisk(ctx context.Context, r autoRule, now time.Time, zone *time.Location) error {
	hours := r.Params.HoursBefore
	if hours < 1 || hours > 23 {
		hours = 4
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, key, name::text, period_days, grace_periods FROM economy_streak WHERE is_active AND ($1 = '' OR key = $1)`, r.Params.StreakKey)
	if err != nil {
		return err
	}
	type streak struct {
		id, key, name string
		days, grace   int
	}
	list := []streak{}
	for rows.Next() {
		var st streak
		if err := rows.Scan(&st.id, &st.key, &st.name, &st.days, &st.grace); err != nil {
			rows.Close()
			return err
		}
		list = append(list, st)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	set, err := localized.Load(ctx, s.db)
	if err != nil {
		return err
	}
	for _, st := range list {
		days := int64(st.days)
		if days < 1 {
			days = 1
		}
		idx := floorDivide(civilDayIndex(now, zone), days)
		endDay := (idx + 1) * days
		ends := time.Date(1970, 1, 1, 0, 0, 0, 0, zone).AddDate(0, 0, int(endDay))
		if now.Before(ends.Add(-time.Duration(hours) * time.Hour)) {
			continue
		}
		urows, err := s.db.QueryContext(ctx, `SELECT user_id::text, count FROM economy_user_streak WHERE streak_id=$1 AND count > 0 AND last_period IS NOT NULL
			AND last_period < $2 AND last_period = $2 - 1 - GREATEST($3 - grace_used, 0) LIMIT 10000`, st.id, idx, st.grace)
		if err != nil {
			return err
		}
		fires := []fireRow{}
		name := set.Resolve(parseI18n(st.name), set.Default)
		for urows.Next() {
			var uid string
			var count int
			if err := urows.Scan(&uid, &count); err != nil {
				urows.Close()
				return err
			}
			fires = append(fires, fireRow{RuleID: r.ID, UserID: uid, Occurrence: "streak:" + st.key + ":" + strconv.FormatInt(idx, 10),
				Vars: map[string]string{"streak": name, "count": strconv.Itoa(count), "hours": strconv.Itoa(hours)}})
		}
		urows.Close()
		if err := urows.Err(); err != nil {
			return err
		}
		if err := s.insertFires(ctx, fires); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) scanChanceReady(ctx context.Context, r autoRule, now time.Time) error {
	rows, err := s.db.QueryContext(ctx, `SELECT p.user_id::text, g.key, g.name::text, max(p.created_at)
		FROM economy_chance_play p JOIN economy_chance_game g ON g.id=p.game_id
		WHERE p.mode='free' AND g.is_active AND g.interval_time IS NOT NULL AND ($1 = '' OR g.key = $1)
			AND p.created_at > $2::timestamptz - make_interval(mins => g.interval_time) - $3::interval
		GROUP BY p.user_id, g.key, g.name::text, g.interval_time
		HAVING max(p.created_at) + make_interval(mins => g.interval_time) <= $2::timestamptz
		LIMIT 10000`, r.Params.GameKey, now, autoReadyLookback.String())
	if err != nil {
		return err
	}
	set, err := localized.Load(ctx, s.db)
	if err != nil {
		rows.Close()
		return err
	}
	fires := []fireRow{}
	for rows.Next() {
		var uid, key, name string
		var last time.Time
		if err := rows.Scan(&uid, &key, &name, &last); err != nil {
			rows.Close()
			return err
		}
		fires = append(fires, fireRow{RuleID: r.ID, UserID: uid, Occurrence: "chance:" + key + ":" + strconv.FormatInt(last.UnixMilli(), 10),
			Vars: map[string]string{"game": set.Resolve(parseI18n(name), set.Default)}})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	return s.insertFires(ctx, fires)
}

func (s *service) scanInactive(ctx context.Context, r autoRule, now time.Time) error {
	days := r.Params.Days
	if days < 1 || days > 365 {
		days = 3
	}
	rows, err := s.db.QueryContext(ctx, `WITH lt AS (SELECT id FROM economy_event_type WHERE key='login')
		SELECT u.id::text, GREATEST(u.create_time, COALESCE(tot.last_at, u.create_time)) AS last
		FROM users u LEFT JOIN economy_player_event_total tot ON tot.user_id=u.id AND tot.type_id=(SELECT id FROM lt)
		WHERE u.id <> '`+systemUserID+`'::uuid AND u.disable_time = '1970-01-01 00:00:00+00'
			AND GREATEST(u.create_time, COALESCE(tot.last_at, u.create_time)) <= $1::timestamptz - make_interval(days => $2)
			AND GREATEST(u.create_time, COALESCE(tot.last_at, u.create_time)) > $1::timestamptz - make_interval(days => $2 + 2)
		LIMIT 10000`, now, days)
	if err != nil {
		return err
	}
	fires := []fireRow{}
	for rows.Next() {
		var uid string
		var last time.Time
		if err := rows.Scan(&uid, &last); err != nil {
			rows.Close()
			return err
		}
		fires = append(fires, fireRow{RuleID: r.ID, UserID: uid, Occurrence: "inactive:" + strconv.Itoa(days) + ":" + last.UTC().Format("20060102T150405"),
			Vars: map[string]string{"days": strconv.Itoa(days)}})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	return s.insertFires(ctx, fires)
}

type pendingFire struct {
	RuleID, UserID, Occurrence string
	Vars                       map[string]string
}

func (s *service) deliverFires(ctx context.Context) (int, error) {
	zone := projectZone(ctx, s.db)
	now := s.now()
	var handled int
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `SELECT rule_id::text, user_id::text, occurrence, vars::text FROM publication_auto_fire
			WHERE status='pending' AND due_at <= now() ORDER BY due_at, rule_id, user_id LIMIT $1 FOR UPDATE SKIP LOCKED`, autoFireBatch)
		if err != nil {
			return err
		}
		byRule := map[string][]pendingFire{}
		order := []string{}
		for rows.Next() {
			var f pendingFire
			var vars string
			if err := rows.Scan(&f.RuleID, &f.UserID, &f.Occurrence, &vars); err != nil {
				rows.Close()
				return err
			}
			_ = json.Unmarshal([]byte(vars), &f.Vars)
			if _, ok := byRule[f.RuleID]; !ok {
				order = append(order, f.RuleID)
			}
			byRule[f.RuleID] = append(byRule[f.RuleID], f)
			handled++
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		for _, ruleID := range order {
			if err := s.deliverRuleFires(ctx, tx, ruleID, byRule[ruleID], now, zone); err != nil {
				return err
			}
		}
		return nil
	})
	return handled, err
}

func (s *service) markFires(ctx context.Context, tx *sql.Tx, ruleID string, fires []pendingFire, status, reason string, notificationID *string) error {
	if len(fires) == 0 {
		return nil
	}
	type key struct {
		U string `json:"u"`
		O string `json:"o"`
	}
	keys := make([]key, 0, len(fires))
	for _, f := range fires {
		keys = append(keys, key{U: f.UserID, O: f.Occurrence})
	}
	raw, _ := json.Marshal(keys)
	_, err := tx.ExecContext(ctx, `UPDATE publication_auto_fire f SET status=$2, reason=$3, notification_id=$4, done_at=now()
		FROM jsonb_to_recordset($5::jsonb) AS x(u uuid, o text) WHERE f.rule_id=$1 AND f.user_id=x.u AND f.occurrence=x.o AND f.status='pending'`, ruleID, status, reason, notificationID, string(raw))
	return err
}

func (s *service) deliverRuleFires(ctx context.Context, tx *sql.Tx, ruleID string, fires []pendingFire, now time.Time, zone *time.Location) error {
	r, err := scanAutoRule(tx.QueryRowContext(ctx, `SELECT `+autoRuleColumns+` FROM publication_auto_rule WHERE id=$1 AND is_active`, ruleID))
	if errors.Is(err, sql.ErrNoRows) {
		return s.markFires(ctx, tx, ruleID, fires, "skipped", "rule_off", nil)
	}
	if err != nil {
		return err
	}
	if inQuiet(r, now.In(zone)) {
		raw, _ := json.Marshal(fireKeys(fires))
		_, err := tx.ExecContext(ctx, `UPDATE publication_auto_fire f SET due_at=$2 FROM jsonb_to_recordset($3::jsonb) AS x(u uuid, o text)
			WHERE f.rule_id=$1 AND f.user_id=x.u AND f.occurrence=x.o AND f.status='pending'`, ruleID, quietEnds(r, now, zone), string(raw))
		return err
	}
	ids := make([]string, 0, len(fires))
	for _, f := range fires {
		ids = append(ids, f.UserID)
	}
	recent := map[string]bool{}
	if r.Cooldown > 0 {
		rows, err := tx.QueryContext(ctx, `SELECT DISTINCT user_id::text FROM publication_auto_fire WHERE rule_id=$1 AND status='sent' AND user_id = ANY($2::uuid[]) AND done_at > now() - make_interval(mins => $3)`,
			ruleID, uuidArray(ids), r.Cooldown)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return err
			}
			recent[id] = true
		}
		rows.Close()
	}
	out, err := s.optedOut(ctx, r.CategoryID, ids)
	if err != nil {
		return err
	}
	var cooled, opted, send []pendingFire
	picked := map[string]bool{}
	for _, f := range fires {
		switch {
		case out[f.UserID]:
			opted = append(opted, f)
		case recent[f.UserID] || (picked[f.UserID] && r.Cooldown > 0):
			cooled = append(cooled, f)
		default:
			picked[f.UserID] = true
			send = append(send, f)
		}
	}
	if err := s.markFires(ctx, tx, ruleID, opted, "skipped", "opted_out", nil); err != nil {
		return err
	}
	if err := s.markFires(ctx, tx, ruleID, cooled, "skipped", "cooldown", nil); err != nil {
		return err
	}
	if len(send) == 0 {
		return nil
	}
	names, err := playerNames(ctx, tx, ids)
	if err != nil {
		return err
	}
	vars := map[string]map[string]string{}
	users := make([]string, 0, len(send))
	for _, f := range send {
		v := map[string]string{}
		for k, x := range f.Vars {
			v[k] = x
		}
		if _, ok := v["player"]; !ok {
			v["player"] = names[f.UserID]
		}
		vars[f.UserID] = v
		users = append(users, f.UserID)
	}
	in := notificationInput{Channel: r.Channel, Title: r.Title, Body: r.Body, ImageURL: r.Image, CategoryID: r.CategoryID, Target: target{Mode: "users", UserIDs: users}}
	if len(r.Action) > 0 && string(r.Action) != "null" {
		var a action
		if json.Unmarshal(r.Action, &a) == nil && a.Type != "" {
			in.Action = &a
		}
	}
	id, err := s.createNotificationTx(ctx, tx, in, "auto:"+r.Key, notifyOpts{kind: "auto", ruleID: r.ID, vars: vars, trusted: true})
	if err != nil {
		return err
	}
	return s.markFires(ctx, tx, ruleID, send, "sent", "", &id)
}

func fireKeys(fires []pendingFire) []map[string]string {
	out := make([]map[string]string, 0, len(fires))
	for _, f := range fires {
		out = append(out, map[string]string{"u": f.UserID, "o": f.Occurrence})
	}
	return out
}

func playerNames(ctx context.Context, q rowsQueryer, ids []string) (map[string]string, error) {
	rows, err := q.QueryContext(ctx, `SELECT id::text, COALESCE(NULLIF(display_name,''), username) FROM users WHERE id = ANY($1::uuid[])`, uuidArray(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		out[id] = name
	}
	return out, rows.Err()
}

const autoRuleJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'trigger',t.trigger,'eventKey',t.event_key,'matchField',t.match_field,'matchValue',t.match_value,
	'params',t.params,'channel',t.channel,'categoryId',t.category_id,'categoryKey',c.key,'categoryName',c.name,'title',t.title,'body',t.body,'imageUrl',t.image_url,'action',t.action,
	'cooldownMinutes',t.cooldown_minutes,
	'quietStart',CASE WHEN t.quiet_start IS NULL THEN NULL ELSE lpad((t.quiet_start/60)::text,2,'0')||':'||lpad((t.quiet_start%60)::text,2,'0') END,
	'quietEnd',CASE WHEN t.quiet_end IS NULL THEN NULL ELSE lpad((t.quiet_end/60)::text,2,'0')||':'||lpad((t.quiet_end%60)::text,2,'0') END,
	'isActive',t.is_active,
	'sent7',(SELECT count(*) FROM publication_auto_fire f WHERE f.rule_id=t.id AND f.status='sent' AND f.created_at > now() - interval '7 days'),
	'skipped7',(SELECT count(*) FROM publication_auto_fire f WHERE f.rule_id=t.id AND f.status='skipped' AND f.created_at > now() - interval '7 days'),
	'pending',(SELECT count(*) FROM publication_auto_fire f WHERE f.rule_id=t.id AND f.status='pending'),
	'opened7',COALESCE((SELECT sum(n.opened_count) FROM publication_notification n WHERE n.rule_id=t.id AND n.created_at > now() - interval '7 days'),0),
	'lastSentAt',(SELECT max(f.done_at) FROM publication_auto_fire f WHERE f.rule_id=t.id AND f.status='sent'),
	'createdAt',t.created_at,'updatedAt',t.updated_at)`

func autoRuleResource() resource {
	return resource{
		name:    "auto-rules",
		table:   "publication_auto_rule",
		json:    autoRuleJSON,
		from:    ` FROM publication_auto_rule t JOIN publication_notify_category c ON c.id=t.category_id`,
		search:  []string{"t.key", "t.name", translatedSearch("t.title"), "t.event_key"},
		filters: map[string]string{"isActive": "t.is_active = ?::boolean", "trigger": "t.trigger = ?", "categoryId": "t.category_id = ?::uuid", "channel": "t.channel = ?"},
		order:   "t.key, t.id",
		save:    saveAutoRule,
	}
}

func saveAutoRule(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key             string          `json:"key"`
		Name            string          `json:"name"`
		Trigger         string          `json:"trigger"`
		EventKey        string          `json:"eventKey"`
		MatchField      string          `json:"matchField"`
		MatchValue      string          `json:"matchValue"`
		Params          autoParams      `json:"params"`
		Channel         string          `json:"channel"`
		CategoryID      string          `json:"categoryId"`
		Title           i18n            `json:"title"`
		Body            i18n            `json:"body"`
		ImageURL        string          `json:"imageUrl"`
		Action          json.RawMessage `json:"action"`
		CooldownMinutes *int            `json:"cooldownMinutes"`
		QuietStart      string          `json:"quietStart"`
		QuietEnd        string          `json:"quietEnd"`
		IsActive        *bool           `json:"isActive"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := localized.Load(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanBannerKey(ctx, tx, "publication_auto_rule", id, create, in.Key)
	if err != nil {
		return err
	}
	name := strings.TrimSpace(in.Name)
	if len([]rune(name)) > 120 {
		return badRequest("TEXT_TOO_LONG", "name", "name is too long (max 120)")
	}
	if !containsStr(autoTriggers, in.Trigger) {
		return badRequest("INVALID_TRIGGER", "trigger", "trigger must be event, streak_at_risk, chance_ready or inactive")
	}
	p := autoParams{}
	eventKey, matchField, matchValue := "", "", ""
	switch in.Trigger {
	case "event":
		eventKey = strings.TrimSpace(in.EventKey)
		var ok bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_event_type WHERE key=$1)`, eventKey).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("EVENT_TYPE_NOT_FOUND", "eventKey", "event type does not exist")
		}
		matchField, matchValue = strings.TrimSpace(in.MatchField), strings.TrimSpace(in.MatchValue)
		if matchField != "" && !autoFieldPattern.MatchString(matchField) {
			return badRequest("INVALID_MATCH_FIELD", "matchField", "use a metadata field name like entityId")
		}
		if matchField == "" {
			matchValue = ""
		} else if matchValue == "" || len([]rune(matchValue)) > 100 {
			return badRequest("INVALID_MATCH_VALUE", "matchValue", "give the value to match (max 100)")
		}
	case "streak_at_risk":
		p.StreakKey = strings.TrimSpace(in.Params.StreakKey)
		if p.StreakKey != "" {
			var ok bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_streak WHERE key=$1)`, p.StreakKey).Scan(&ok); err != nil {
				return err
			}
			if !ok {
				return badRequest("STREAK_NOT_FOUND", "params.streakKey", "streak does not exist")
			}
		}
		p.HoursBefore = in.Params.HoursBefore
		if p.HoursBefore == 0 {
			p.HoursBefore = 4
		}
		if p.HoursBefore < 1 || p.HoursBefore > 23 {
			return badRequest("INVALID_HOURS_BEFORE", "params.hoursBefore", "between 1 and 23 hours")
		}
	case "chance_ready":
		p.GameKey = strings.TrimSpace(in.Params.GameKey)
		if p.GameKey != "" {
			var ok bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_chance_game WHERE key=$1 AND interval_time IS NOT NULL)`, p.GameKey).Scan(&ok); err != nil {
				return err
			}
			if !ok {
				return badRequest("CHANCE_GAME_NOT_FOUND", "params.gameKey", "chance game does not exist or has no free turn")
			}
		}
	case "inactive":
		p.Days = in.Params.Days
		if p.Days == 0 {
			p.Days = 3
		}
		if p.Days < 1 || p.Days > 365 {
			return badRequest("INVALID_DAYS", "params.days", "between 1 and 365 days")
		}
	}
	paramsJSON, _ := json.Marshal(p)
	if in.Channel == "" {
		in.Channel = "push_in_app"
	}
	if !containsStr(autoChannels, in.Channel) {
		return badRequest("INVALID_CHANNEL", "channel", "channel must be push, in_app or push_in_app")
	}
	cat, err := cleanCategoryID(ctx, tx, in.CategoryID, true)
	if err != nil {
		return err
	}
	title, err := cleanI18n(langs, "title", in.Title, true, 120)
	if err != nil {
		return err
	}
	body, err := cleanI18n(langs, "body", in.Body, true, 500)
	if err != nil {
		return err
	}
	image, err := cleanWebURL("imageUrl", in.ImageURL)
	if err != nil {
		if app, ok := asApp(err); ok && app.Code == "INVALID_URL" && strings.HasPrefix(strings.TrimSpace(in.ImageURL), "/") {
			image, err = cleanURL("imageUrl", in.ImageURL)
		}
		if err != nil {
			return err
		}
	}
	var actionJSON *string
	if len(in.Action) > 0 && string(in.Action) != "null" {
		var a action
		if err := json.Unmarshal(in.Action, &a); err != nil {
			return badRequest("INVALID_ACTION", "action", "action must be an object")
		}
		if a.Type != "" && a.Type != "none" {
			clean, err := cleanCampaignAction(ctx, tx, langs, "action", a)
			if err != nil {
				return err
			}
			b, _ := json.Marshal(clean)
			v := string(b)
			actionJSON = &v
		}
	}
	cooldown := 1440
	if in.CooldownMinutes != nil {
		cooldown = *in.CooldownMinutes
	}
	if cooldown < 0 || cooldown > 525600 {
		return badRequest("INVALID_COOLDOWN", "cooldownMinutes", "between 0 and 525600 minutes")
	}
	qs, err := parseMinutes("quietStart", in.QuietStart, false)
	if err != nil {
		return err
	}
	qe, err := parseMinutes("quietEnd", in.QuietEnd, false)
	if err != nil {
		return err
	}
	if (qs == nil) != (qe == nil) {
		field := "quietEnd"
		if qs == nil {
			field = "quietStart"
		}
		return badRequest("INVALID_QUIET_HOURS", field, "give both a start and an end time")
	}
	if qs != nil && *qs == *qe {
		return badRequest("INVALID_QUIET_HOURS", "quietEnd", "start and end cannot be the same")
	}
	return execSave(ctx, tx, "publication_auto_rule", id, create,
		[]string{"key", "name", "trigger", "event_key", "match_field", "match_value", "params", "channel", "category_id", "title", "body", "image_url", "action", "cooldown_minutes", "quiet_start", "quiet_end", "is_active"},
		[]any{key, name, in.Trigger, eventKey, matchField, matchValue, string(paramsJSON), in.Channel, *cat, i18nJSON(title), i18nJSON(body), image, actionJSON, cooldown, qs, qe, boolOr(in.IsActive, true)})
}

func cleanCategoryID(ctx context.Context, q queryer, raw string, required bool) (*string, error) {
	id, err := cleanUUID("categoryId", raw, required)
	if err != nil || id == nil {
		if err != nil {
			if app, ok := asApp(err); ok && app.Code == "REQUIRED" {
				return nil, badRequest("CATEGORY_REQUIRED", "categoryId", "pick a notification category")
			}
		}
		return id, err
	}
	var ok bool
	if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_notify_category WHERE id=$1)`, *id).Scan(&ok); err != nil {
		return nil, err
	}
	if !ok {
		return nil, badRequest("CATEGORY_NOT_FOUND", "categoryId", "notification category does not exist")
	}
	return id, nil
}

func (s *service) consoleAutoFires(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	p := readPage(r)
	where := &sqlWhere{args: []any{id}, parts: []string{"f.rule_id = $1"}}
	if v := r.URL.Query().Get("status"); v != "" {
		if v != "pending" && v != "sent" && v != "skipped" {
			s.writeError(w, badRequest("INVALID_FILTER", "status", "pending, sent or skipped"))
			return
		}
		where.add("f.status = ?", v)
	}
	var total int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM publication_auto_fire f`+where.sql(), where.args...).Scan(&total); err != nil {
		s.writeError(w, err)
		return
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(r.Context(), `SELECT jsonb_build_object('userId',f.user_id,'username',u.username,'displayName',u.display_name,'occurrence',f.occurrence,'status',f.status,
		'reason',f.reason,'vars',f.vars,'dueAt',f.due_at,'notificationId',f.notification_id,'createdAt',f.created_at,'doneAt',f.done_at)::text
		FROM publication_auto_fire f LEFT JOIN users u ON u.id=f.user_id`+where.sql()+` ORDER BY f.created_at DESC, f.user_id LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2),
		append(append([]any{}, where.args...), p.Limit, p.Offset)...)
	if err != nil {
		s.writeError(w, err)
		return
	}
	items, err := rawRows(rows)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": p.Page, "limit": p.Limit})
}

func (s *service) consoleAutoTest(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID string            `json:"userId"`
		Vars   map[string]string `json:"vars"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	rule, err := scanAutoRule(s.db.QueryRowContext(r.Context(), `SELECT `+autoRuleColumns+` FROM publication_auto_rule WHERE id::text=$1`, pathVar(r, "id")))
	if errors.Is(err, sql.ErrNoRows) {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	uid, err := cleanUUID("userId", in.UserID, true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	names, err := playerNames(r.Context(), s.db, []string{*uid})
	if err != nil {
		s.writeError(w, err)
		return
	}
	if names[*uid] == "" {
		s.writeError(w, badRequest("USER_NOT_FOUND", "userId", "player does not exist"))
		return
	}
	vars := map[string]string{"player": names[*uid], "value": "1", "days": "3", "hours": "4", "count": "5", "streak": "…", "game": "…"}
	for k, v := range in.Vars {
		if autoFieldPattern.MatchString(k) && len(vars) < 40 {
			vars[k] = truncate(v, 80)
		}
	}
	notif := notificationInput{Channel: rule.Channel, Title: rule.Title, Body: rule.Body, ImageURL: rule.Image, CategoryID: rule.CategoryID, Target: target{Mode: "users", UserIDs: []string{*uid}}}
	if len(rule.Action) > 0 && string(rule.Action) != "null" {
		var a action
		if json.Unmarshal(rule.Action, &a) == nil && a.Type != "" {
			notif.Action = &a
		}
	}
	var id string
	err = s.inTx(r.Context(), func(tx *sql.Tx) error {
		var err error
		id, err = s.createNotificationTx(r.Context(), tx, notif, consoleUser(r), notifyOpts{kind: "test", ruleID: rule.ID, vars: map[string]map[string]string{*uid: vars}, trusted: true})
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	out, err := s.notificationDetail(r.Context(), id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}
