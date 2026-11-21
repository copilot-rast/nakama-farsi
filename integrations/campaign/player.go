package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/campaign/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/campaignhooks"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func owned(ctx context.Context, stageKey string) any {
	if owner, ref := campaignhooks.OwnerOf(ctx, stageKey); owner != "" {
		return map[string]string{"owner": owner, "ref": ref}
	}
	return nil
}

func iso(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}

func (s *service) zone(cfg settings.All) *time.Location {
	loc, err := time.LoadLocation(cfg.Rewards.DayZone)
	if err != nil {
		return time.UTC
	}
	return loc
}

func text(langs *localized.Set, lang string, m map[string]string) string {
	if langs == nil {
		return m[lang]
	}
	return langs.Resolve(localized.Text(m), lang)
}

func rewardList(cur map[string]int64) []map[string]any {
	out := []map[string]any{}
	for _, k := range sortedKeys(cur) {
		if cur[k] > 0 {
			out = append(out, map[string]any{"currencyKey": k, "amount": cur[k]})
		}
	}
	return out
}

func sortedKeys(m map[string]int64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

func (s *service) multipliers(cfg settings.All, st *stage) []float64 {
	if st.OwnMultiplier && len(st.Multipliers) > 0 {
		return st.Multipliers
	}
	if st.Mode == settings.ModeRepeatable {
		return cfg.Stars.RepeatableMultipliers
	}
	return cfg.Stars.Multipliers
}

func (s *service) starRules(cfg settings.All, st *stage) []settings.StarRule {
	if st.OwnRules {
		return st.StarRules
	}
	if st.StarTemplate != "" {
		if t := cfg.Stars.Template(st.StarTemplate); t != nil {
			return t.Rules
		}
	}
	return cfg.Stars.Rules
}

func minSecondsOf(cfg settings.All, st *stage) float64 {
	if st.MinSeconds != nil {
		return *st.MinSeconds
	}
	return cfg.Attempts.DefaultMinSeconds
}

func replayable(st *stage, r *userStage) bool {
	return !st.NoReplay || r.Wins == 0
}

func (s *service) nextLevelXP(ctx context.Context, p *player, level int) any {
	var v sql.NullFloat64
	err := s.db.QueryRowContext(ctx, `SELECT s.threshold::float8 FROM economy_level_step s WHERE s.level=$2 AND s.table_id = COALESCE(
		(SELECT id FROM economy_level_table WHERE key=NULLIF($1,'')), (SELECT id FROM economy_level_table WHERE is_active AND NULLIF($1,'') IS NULL ORDER BY sort, created_at LIMIT 1))`,
		p.cfg.XP.LevelTableKey, level+1).Scan(&v)
	if err != nil || !v.Valid {
		return nil
	}
	return int64(v.Float64)
}

func (s *service) xpLevel(ctx context.Context, p *player) (int64, int) {
	var xp int64
	_ = s.db.QueryRowContext(ctx, `SELECT xp FROM campaign_user WHERE user_id=$1`, p.userID).Scan(&xp)
	return xp, p.level(ctx, p.cfg.XP.LevelTableKey)
}

func (s *service) rpcChapters(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang    string `json:"lang"`
		Chapter string `json:"chapter"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	cat, err := loadCatalog(ctx, s.db, true)
	if err != nil {
		return nil, err
	}
	p, err := s.loadPlayer(ctx, s.db, userID, cfg, cat)
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	now, loc := s.now(), s.zone(cfg)
	chapters := []map[string]any{}
	for _, ch := range cat.chapters {
		if in.Chapter != "" && ch.Key != in.Chapter {
			continue
		}
		locks := p.unmet(ctx, p.chapterRules(ch))
		if len(locks) > 0 && !cfg.Unlock.ShowLocked {
			continue
		}
		earned, max, pct := p.chapterStars(ch)
		var paid []any
		var raw []byte
		if err := s.db.QueryRowContext(ctx, `SELECT milestones FROM campaign_user_chapter WHERE user_id=$1 AND chapter_id=$2`, userID, ch.ID).Scan(&raw); err == nil {
			unmarshal(raw, &paid)
		}
		milestones := []map[string]any{}
		for _, m := range ch.Milestones {
			milestones = append(milestones, map[string]any{"id": milestoneID(m), "percent": m.Percent, "rewards": rewardList(m.Currencies), "prizeKey": m.PrizeKey, "prizeGroupKey": m.PrizeGroupKey,
				"claimed": milestonePaid(paid, m), "reached": reachedPercent(pct, m.Percent)})
		}
		resolveText := func(m map[string]string) string { return text(langs, lang, m) }
		stages := []map[string]any{}
		for _, st := range cat.stagesOf(ch.ID) {
			slocks := append(append([]Rule{}, locks...), p.unmet(ctx, p.stageRules(st))...)
			if len(slocks) > 0 && !cfg.Unlock.ShowLocked {
				continue
			}
			r := p.row(st.ID)
			g := p.group(st.group())
			stages = append(stages, map[string]any{
				"key": st.Key, "name": text(langs, lang, st.Name), "description": text(langs, lang, st.Description), "iconUrl": st.IconURL, "address": st.Address,
				"sort": st.Sort, "mode": st.Mode, "unlocked": len(slocks) == 0, "locks": slocks, "bestStars": r.BestStars, "completions": r.Completions, "wins": r.Wins,
				"bestSeconds": r.BestSeconds, "rewards": rewardList(st.Rewards.Currencies), "xp": st.XP, "tierPrizes": st.TierPrizes, "starMultipliers": s.multipliers(cfg, st),
				"claimed": rewardList(r.Claimed), "claimedXp": r.ClaimedXP, "timeLimitSeconds": st.TimeLimit, "rewardGroup": st.RewardGroup,
				"cooldownSeconds": st.Cooldown, "dailyRewardLimit": st.DailyLimit, "nextRewardAt": iso(nextRewardAt(g, st.Cooldown, now)),
				"rewardsLeftToday": rewardsLeft(g, st.DailyLimit, now, loc), "metadata": st.Metadata, "completedBy": owned(ctx, st.Key),
				"starRules": s.starRules(cfg, st), "starTemplate": st.StarTemplate, "minSeconds": minSecondsOf(cfg, st), "noReplayAfterWin": st.NoReplay, "replayable": replayable(st, r),
				"settings": settings.Resolve(cfg.Fields.Stage, settings.ScopeOf(ch.Key, ch.Metadata), st.Settings, resolveText),
			})
		}
		chapters = append(chapters, map[string]any{"key": ch.Key, "name": text(langs, lang, ch.Name), "description": text(langs, lang, ch.Description), "iconUrl": ch.IconURL,
			"address": ch.Address, "sort": ch.Sort, "unlocked": len(locks) == 0, "locks": locks, "stars": earned, "maxStars": max, "starPercent": math.Round(pct*100) / 100,
			"milestones": milestones, "stages": stages, "metadata": ch.Metadata, "settings": settings.Resolve(cfg.Fields.Chapter, settings.ScopeOf(ch.Key, ch.Metadata), ch.Settings, resolveText)})
	}
	rides := []map[string]any{}
	where := `is_active`
	if !cfg.Freeride.TableEnabled {
		where = `false`
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+freerideCols+` FROM campaign_freeride WHERE `+where+` ORDER BY sort, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		f, err := scanFreeride(rows)
		if err != nil {
			return nil, err
		}
		g := p.group(f.group())
		rides = append(rides, map[string]any{"key": f.Key, "name": text(langs, lang, f.Name), "description": text(langs, lang, f.Description), "iconUrl": f.IconURL,
			"rewards": rewardList(f.Rewards.Currencies), "prizeKey": f.Rewards.PrizeKey, "prizeGroupKey": f.Rewards.PrizeGroupKey, "xp": f.XP, "minLevel": f.MinLevel,
			"rewardGroup": f.RewardGroup, "cooldownSeconds": f.Cooldown, "dailyRewardLimit": f.DailyLimit,
			"nextRewardAt": iso(nextRewardAt(g, f.Cooldown, now)), "rewardsLeftToday": rewardsLeft(g, f.DailyLimit, now, loc)})
	}
	xp, level := s.xpLevel(ctx, p)
	return map[string]any{"lang": lang, "serverTime": now.UTC().Format(time.RFC3339), "xp": xp, "level": level, "nextLevelXp": s.nextLevelXP(ctx, p, level), "maxStars": cfg.Stars.MaxStars, "chapters": chapters, "freeride": rides}, nil
}

func (s *service) rpcStart(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		StageKey string `json:"stageKey"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	return s.startWith(ctx, userID, in.StageKey, false, itemParam(payload, s.settings(ctx).Items))
}

func (s *service) start(ctx context.Context, userID, stageKey string, server bool) (any, error) {
	return s.startWith(ctx, userID, stageKey, server, "")
}

func (s *service) startWith(ctx context.Context, userID, stageKey string, server bool, itemKey string) (any, error) {
	stageKey = strings.TrimSpace(stageKey)
	if stageKey == "" {
		return nil, apierr.BadRequest("STAGE_KEY_REQUIRED").WithField("stageKey")
	}
	cfg := s.settings(ctx)
	cat, err := loadCatalog(ctx, s.db, true)
	if err != nil {
		return nil, err
	}
	st := cat.byKey[stageKey]
	if st == nil {
		return nil, apierr.NotFound("STAGE_NOT_FOUND").WithField("stageKey")
	}
	p, err := s.loadPlayer(ctx, s.db, userID, cfg, cat)
	if err != nil {
		return nil, err
	}
	if locks := p.stageLocks(ctx, st); len(locks) > 0 {
		return nil, apierr.Forbidden("STAGE_LOCKED").With("locks", locks)
	}
	if !replayable(st, p.row(st.ID)) {
		return nil, apierr.Forbidden("STAGE_ALREADY_WON").WithField("stageKey")
	}
	item, err := s.checkStageItem(ctx, userID, st, cfg.Items, itemKey)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	id := uuid.Must(uuid.NewV4()).String()
	expires := now.Add(time.Duration(cfg.Attempts.Hours) * time.Hour)
	var cost *playCost
	var regen *economyhooks.Regen
	var rawCost []byte
	if currency, amount := stageCost(cfg, st); amount > 0 {
		cost, regen, err = s.takeCost(ctx, userID, id, currency, amount)
		if err != nil {
			return nil, err
		}
		rawCost, _ = json.Marshal(cost)
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if cfg.Attempts.CloseOthers && (!server || cfg.Attempts.ServerClosesOthers) {
			if _, err := tx.ExecContext(ctx, `UPDATE campaign_attempt SET status='closed', closed_at=$2 WHERE user_id=$1 AND status='open'`, userID, now); err != nil {
				return err
			}
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO campaign_attempt (id, user_id, stage_id, started_at, expires_at, play_cost) VALUES ($1,$2,$3,$4,$5,$6)`, id, userID, st.ID, now, expires, rawCost)
		return err
	})
	if err != nil {
		if cost != nil && cost.Spent > 0 {
			if rerr := s.giveBack(ctx, userID, id, &refundRef{CurrencyKey: cost.CurrencyKey, Amount: cost.Spent, Via: cost.Via}); rerr != nil {
				s.logger.Error("Campaign play cost of %s not returned after a failed start: %v", id, rerr)
			}
		}
		return nil, err
	}
	if gameevents.Has() {
		if _, errs := gameevents.Publish(ctx, gameevents.Event{ID: "start:" + id, UserID: userID, Source: "campaign", Context: st.Key,
			Typed: []gameevents.Typed{gameevents.Progression("Start", st.ChapterKey, st.Key, "", nil)}}); len(errs) > 0 {
			s.logger.Debug("Campaign start event not recorded: %v", errs)
		}
	}
	out := map[string]any{"attemptId": id, "stageKey": st.Key, "startedAt": now.Format(time.RFC3339Nano), "expiresAt": expires.Format(time.RFC3339), "timeLimitSeconds": st.TimeLimit, "playCost": nil}
	if item != nil {
		out["item"] = item
	}
	if cost != nil {
		out["playCost"] = cost
		if regen != nil {
			out["regen"] = regen
		}
	}
	return out, nil
}

func (s *service) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	var err error
	for try := 1; try <= 3; try++ {
		var tx *sql.Tx
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if err = fn(tx); err != nil {
			_ = tx.Rollback()
		} else if err = tx.Commit(); err == nil {
			return nil
		}
		if !retryable(err) {
			return err
		}
	}
	return apierr.Conflict("STAGE_BUSY").Retry(1)
}

type completeIn struct {
	AttemptID     string         `json:"attemptId"`
	StageKey      string         `json:"stageKey"`
	Success       bool           `json:"success"`
	ElapsedSecond *float64       `json:"elapsedSeconds"`
	HealthPercent *float64       `json:"healthPercent"`
	Fields        map[string]any `json:"fields"`
	Lang          string         `json:"lang"`
}

func (s *service) rpcComplete(ctx context.Context, userID, payload string) (any, error) {
	var in completeIn
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	var key string
	if _, err := uuid.FromString(strings.TrimSpace(in.AttemptID)); err == nil {
		_ = s.db.QueryRowContext(ctx, `SELECT s.key FROM campaign_attempt a JOIN campaign_stage s ON s.id=a.stage_id WHERE a.id=$1 AND a.user_id=$2`, strings.TrimSpace(in.AttemptID), userID).Scan(&key)
	}
	if owner, ref := campaignhooks.OwnerOf(ctx, key); owner != "" {
		return nil, apierr.Forbidden("STAGE_SERVER_OWNED").With("owner", owner).With("ref", ref)
	}
	return s.complete(ctx, userID, in)
}

func (s *service) hookStart(ctx context.Context, userID, stageKey string) (string, error) {
	out, err := s.start(ctx, userID, stageKey, true)
	if err != nil {
		return "", err
	}
	id, _ := out.(map[string]any)["attemptId"].(string)
	return id, nil
}

func (s *service) hookComplete(ctx context.Context, userID string, c campaignhooks.Completion) (map[string]any, error) {
	elapsed := c.ElapsedSeconds
	out, err := s.complete(ctx, userID, completeIn{AttemptID: c.AttemptID, StageKey: c.StageKey, Success: c.Success, ElapsedSecond: &elapsed, Fields: c.Fields})
	if err != nil {
		return nil, err
	}
	m, _ := out.(map[string]any)
	return m, nil
}

type plan struct {
	GrantID    string           `json:"grantId"`
	Currencies map[string]int64 `json:"currencies"`
	Prizes     []prizeRef       `json:"prizes"`
	MetaKey    string           `json:"metaKey"`
	Reason     string           `json:"reason"`
	SourceID   string           `json:"sourceId"`
	Event      *eventRef        `json:"event,omitempty"`
	Refund     *refundRef       `json:"refund,omitempty"`
}

type prizeRef struct {
	PrizeKey      string `json:"prizeKey,omitempty"`
	PrizeGroupKey string `json:"prizeGroupKey,omitempty"`
	SourceID      string `json:"sourceId"`
	For           string `json:"for"`
	Paid          bool   `json:"paid"`
	Error         string `json:"error,omitempty"`
}

type eventRef struct {
	Economy   string             `json:"economy"`
	Analytics string             `json:"analytics"`
	Context   string             `json:"context"`
	Value     float64            `json:"value"`
	Fields    map[string]any     `json:"fields"`
	Sent      bool               `json:"sent"`
	Typed     []gameevents.Typed `json:"typed,omitempty"`
}

var reservedFields = map[string]bool{"stars": true, "success": true, "elapsedSeconds": true, "healthPercent": true, "level": true, "timeSavedPercent": true, "stageKey": true, "chapterKey": true, "mode": true}

func (s *service) reject(ctx context.Context, attemptID, userID, code string) {
	_, _ = s.db.ExecContext(ctx, `UPDATE campaign_attempt SET rejected=rejected+1, last_reject=$3 WHERE id=$1 AND user_id=$2`, attemptID, userID, code)
}

func (s *service) complete(ctx context.Context, userID string, in completeIn) (any, error) {
	in.AttemptID, in.StageKey = strings.TrimSpace(in.AttemptID), strings.TrimSpace(in.StageKey)
	if in.AttemptID == "" {
		return nil, apierr.BadRequest("ATTEMPT_REQUIRED").WithField("attemptId")
	}
	if _, err := uuid.FromString(in.AttemptID); err != nil {
		return nil, apierr.NotFound("ATTEMPT_NOT_FOUND").WithField("attemptId")
	}
	if in.ElapsedSecond == nil || *in.ElapsedSecond < 0 || math.IsNaN(*in.ElapsedSecond) || *in.ElapsedSecond > 7*24*3600 {
		return nil, apierr.BadRequest("INVALID_RESULT").WithField("elapsedSeconds")
	}
	if in.HealthPercent != nil && (*in.HealthPercent < 0 || *in.HealthPercent > 100) {
		return nil, apierr.BadRequest("INVALID_RESULT").WithField("healthPercent")
	}
	if len(in.Fields) > 40 {
		return nil, apierr.BadRequest("TOO_MANY_FIELDS").WithField("fields")
	}
	numbers, extra := map[string]float64{}, map[string]any{}
	for k, v := range in.Fields {
		if !settings.FieldPattern.MatchString(k) || reservedFields[k] {
			continue
		}
		switch x := v.(type) {
		case float64:
			numbers[k], extra[k] = x, x
		case bool:
			extra[k] = x
			if x {
				numbers[k] = 1
			} else {
				numbers[k] = 0
			}
		case string:
			if len(x) <= 64 {
				extra[k] = x
			}
		}
	}
	cfg := s.settings(ctx)
	var done any
	var status string
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		done, status = nil, ""
		var owner, stageID, stageKey string
		var started, expires time.Time
		var stored, rawCost []byte
		err := tx.QueryRowContext(ctx, `SELECT a.user_id::text, a.stage_id::text, s.key, a.started_at, a.expires_at, a.status, a.result, a.play_cost FROM campaign_attempt a JOIN campaign_stage s ON s.id=a.stage_id WHERE a.id=$1 FOR UPDATE OF a`,
			in.AttemptID).Scan(&owner, &stageID, &stageKey, &started, &expires, &status, &stored, &rawCost)
		if errors.Is(err, sql.ErrNoRows) || err == nil && (owner != userID || in.StageKey != "" && in.StageKey != stageKey) {
			return apierr.NotFound("ATTEMPT_NOT_FOUND").WithField("attemptId")
		}
		if err != nil {
			return err
		}
		now := s.now().UTC()
		switch status {
		case "done", "paying":
			return nil
		case "closed":
			return apierr.Conflict("ATTEMPT_CLOSED")
		}
		if now.After(expires) {
			return apierr.Conflict("ATTEMPT_EXPIRED")
		}
		cat, err := loadCatalog(ctx, tx, true)
		if err != nil {
			return err
		}
		st := cat.byKey[stageKey]
		if st == nil {
			return apierr.NotFound("STAGE_NOT_FOUND").WithField("stageKey")
		}
		grace := float64(cfg.Attempts.GraceSeconds)
		serverSeconds := now.Sub(started).Seconds()
		elapsed := *in.ElapsedSecond
		minSeconds := minSecondsOf(cfg, st)
		switch {
		case elapsed > serverSeconds+grace:
			return apierr.BadRequest("INVALID_RESULT").WithField("elapsedSeconds").With("serverSeconds", math.Round(serverSeconds*100)/100)
		case in.Success && st.TimeLimit != nil && elapsed > *st.TimeLimit+grace:
			return apierr.BadRequest("INVALID_RESULT").WithField("elapsedSeconds").With("timeLimitSeconds", *st.TimeLimit)
		case in.Success && minSeconds > 0 && (elapsed < minSeconds || serverSeconds+grace < minSeconds):
			return apierr.BadRequest("STAGE_FINISHED_TOO_FAST").WithField("elapsedSeconds").With("minSeconds", minSeconds)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO campaign_user_stage (user_id, stage_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, userID, st.ID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `SELECT 1 FROM campaign_user_stage WHERE user_id=$1 AND stage_id=$2 FOR UPDATE`, userID, st.ID); err != nil {
			return err
		}
		p, err := s.loadPlayer(ctx, tx, userID, cfg, cat)
		if err != nil {
			return err
		}
		if locks := p.stageLocks(ctx, st); len(locks) > 0 {
			return apierr.Forbidden("STAGE_LOCKED").With("locks", locks)
		}
		if !replayable(st, p.row(st.ID)) {
			return apierr.Forbidden("STAGE_ALREADY_WON").WithField("stageKey")
		}
		res, pl, err := s.score(ctx, tx, p, st, in, numbers, extra, now, serverSeconds)
		if err != nil {
			return err
		}
		pl.GrantID = "campaign:" + in.AttemptID
		pl.Refund = refundFor(cfg, in.Success, rawCost)
		res["attemptId"] = in.AttemptID
		res["playCostRefund"] = nil
		rawRes, _ := json.Marshal(res)
		rawPlan, _ := json.Marshal(pl)
		stars, _ := res["stars"].(int)
		_, err = tx.ExecContext(ctx, `UPDATE campaign_attempt SET status='paying', completed_at=$2, closed_at=$2, success=$3, stars=$4, result=$5, plan=$6 WHERE id=$1`,
			in.AttemptID, now, in.Success, stars, rawRes, rawPlan)
		status = "paying"
		return err
	})
	if err != nil {
		if e, ok := apierr.As(err); ok && (e.Code == "INVALID_RESULT" || e.Code == "STAGE_FINISHED_TOO_FAST") {
			s.reject(ctx, in.AttemptID, userID, e.Code)
		}
		return nil, err
	}
	replayed := status == "done"
	done, err = s.finishAttempt(ctx, in.AttemptID)
	if err != nil {
		return nil, err
	}
	if m, ok := done.(map[string]any); ok && replayed {
		m["replayed"] = true
	}
	return done, nil
}

func (s *service) score(ctx context.Context, tx *sql.Tx, p *player, st *stage, in completeIn, numbers map[string]float64, extra map[string]any, now time.Time, serverSeconds float64) (map[string]any, plan, error) {
	cfg := p.cfg
	run := runResult{Success: in.Success, Elapsed: *in.ElapsedSecond, Health: in.HealthPercent, Fields: numbers}
	vals := resultValues(run, st.TimeLimit)
	rules := s.starRules(cfg, st)
	missingPass := cfg.Stars.MissingField == settings.MissingPass
	stars := computeStars(in.Success, vals, rules, cfg.Stars.MaxStars, cfg.Stars.MinOnSuccess, missingPass)
	row := p.row(st.ID)
	prevBest, prevWins := row.BestStars, row.Wins
	pay := map[string]int64{}
	var prizes []prizeRef
	xpGained := 0
	limited := ""
	if stars > 0 {
		mult := multiplier(stars, s.multipliers(cfg, st))
		repeat := st.Mode == settings.ModeRepeatable || cfg.Rewards.ReplayMode == settings.ReplayFull
		if st.Cooldown > 0 || st.DailyLimit > 0 {
			var err error
			if limited, err = s.takeGroup(ctx, tx, p, st.group(), st.Cooldown, st.DailyLimit, now); err != nil {
				return nil, plan{}, err
			}
		}
		switch {
		case limited != "":
			if st.XPWhenLimited {
				xpGained = int(scaled(int64(st.XP), mult))
			}
		case repeat:
			for k, v := range st.Rewards.Currencies {
				if n := scaled(v, mult); n > 0 {
					pay[k] = n
				}
			}
			xpGained = int(scaled(int64(st.XP), mult))
			for _, t := range st.TierPrizes {
				if t.Stars <= stars {
					prizes = append(prizes, prizeRef{PrizeKey: t.PrizeKey, PrizeGroupKey: t.PrizeGroupKey, SourceID: in.AttemptID + ":t" + itoa(t.Stars), For: "tier"})
				}
			}
		case cfg.Rewards.ReplayMode == settings.ReplayNone && prevWins > 0:
		default:
			for k, v := range st.Rewards.Currencies {
				if n := payout(v, mult, row.Claimed[k]); n > 0 {
					pay[k] = n
					row.Claimed[k] += n
				}
			}
			if n := int(payout(int64(st.XP), mult, int64(row.ClaimedXP))); n > 0 {
				xpGained = n
				row.ClaimedXP += n
			}
			for _, t := range st.TierPrizes {
				if t.Stars > row.TiersPaid && t.Stars <= stars {
					prizes = append(prizes, prizeRef{PrizeKey: t.PrizeKey, PrizeGroupKey: t.PrizeGroupKey, SourceID: p.userID + ":" + st.ID + ":t" + itoa(t.Stars), For: "tier"})
				}
			}
			if stars > row.TiersPaid {
				row.TiersPaid = stars
			}
		}
	}
	row.Completions++
	if in.Success {
		row.Wins++
		if row.BestSeconds == nil || *in.ElapsedSecond < *row.BestSeconds {
			v := *in.ElapsedSecond
			row.BestSeconds = &v
		}
	}
	if stars > row.BestStars {
		row.BestStars = stars
	}
	p.rows[st.ID] = row
	claimed, _ := json.Marshal(row.Claimed)
	if _, err := tx.ExecContext(ctx, `UPDATE campaign_user_stage SET best_stars=$3, claimed=$4, claimed_xp=$5, tiers_paid=$6, completions=$7, wins=$8, best_seconds=$9,
		first_win_at=CASE WHEN $10 AND first_win_at IS NULL THEN $11 ELSE first_win_at END, last_completed_at=$11, updated_at=$11 WHERE user_id=$1 AND stage_id=$2`,
		p.userID, st.ID, row.BestStars, claimed, row.ClaimedXP, row.TiersPaid, row.Completions, row.Wins, row.BestSeconds, in.Success, now); err != nil {
		return nil, plan{}, err
	}
	milestonePay := map[string]int64{}
	reached := []float64{}
	reachedIDs := []string{}
	ch := p.cat.chapter[st.ChapterID]
	if ch != nil && len(ch.Milestones) > 0 && counted(st.Mode, cfg.Rewards.MilestoneModes) && row.BestStars > prevBest {
		if _, err := tx.ExecContext(ctx, `INSERT INTO campaign_user_chapter (user_id, chapter_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, p.userID, ch.ID); err != nil {
			return nil, plan{}, err
		}
		var raw []byte
		if err := tx.QueryRowContext(ctx, `SELECT milestones FROM campaign_user_chapter WHERE user_id=$1 AND chapter_id=$2 FOR UPDATE`, p.userID, ch.ID).Scan(&raw); err != nil {
			return nil, plan{}, err
		}
		var paid []any
		unmarshal(raw, &paid)
		_, _, pct := p.chapterStars(ch)
		for _, m := range ch.Milestones {
			if milestonePaid(paid, m) || !reachedPercent(pct, m.Percent) {
				continue
			}
			paid = append(paid, paidEntry(m))
			reached = append(reached, m.Percent)
			reachedIDs = append(reachedIDs, milestoneID(m))
			for k, v := range m.Currencies {
				if v > 0 {
					milestonePay[k] += v
				}
			}
			if m.PrizeKey != "" || m.PrizeGroupKey != "" {
				prizes = append(prizes, prizeRef{PrizeKey: m.PrizeKey, PrizeGroupKey: m.PrizeGroupKey, SourceID: p.userID + ":" + ch.ID + ":m" + milestoneID(m), For: "milestone"})
			}
		}
		if len(reached) > 0 {
			raw, _ := json.Marshal(paid)
			if _, err := tx.ExecContext(ctx, `UPDATE campaign_user_chapter SET milestones=$3 WHERE user_id=$1 AND chapter_id=$2`, p.userID, ch.ID, raw); err != nil {
				return nil, plan{}, err
			}
		}
	}
	var xpTotal int64
	if err := s.addXP(ctx, tx, p.userID, cfg.XP.MetaKey, xpGained, &xpTotal); err != nil {
		return nil, plan{}, err
	}
	total := map[string]int64{}
	for k, v := range pay {
		total[k] += v
	}
	for k, v := range milestonePay {
		total[k] += v
	}
	pl := plan{Currencies: total, Prizes: prizes, MetaKey: cfg.XP.MetaKey, Reason: "stage_complete", SourceID: st.Key}
	fields := map[string]any{"stars": stars, "success": in.Success, "elapsedSeconds": *in.ElapsedSecond, "stageKey": st.Key, "chapterKey": st.ChapterKey, "mode": st.Mode,
		"timeSavedPercent": math.Round(vals["timeSavedPercent"]*100) / 100, "xp": xpGained}
	if in.HealthPercent != nil {
		fields["healthPercent"] = *in.HealthPercent
	}
	for k, v := range extra {
		fields[k] = v
	}
	ev := &eventRef{Context: st.Key, Value: float64(stars), Fields: fields, Typed: stageTyped(st.ChapterKey, st.Key, in.Success, stars, total)}
	if in.Success || cfg.Events.OnFailure {
		ev.Economy, ev.Analytics = cfg.Events.EconomyEvent, cfg.Events.AnalyticsEvent
	}
	pl.Event = ev
	g := p.group(st.group())
	res := map[string]any{
		"stageKey": st.Key, "chapterKey": st.ChapterKey, "mode": st.Mode, "success": in.Success, "stars": stars, "bestStars": row.BestStars, "previousBestStars": prevBest,
		"rewards": rewardList(pay), "milestoneRewards": rewardList(milestonePay), "milestones": reached, "milestoneIds": reachedIDs, "xpGained": xpGained, "xp": xpTotal,
		"criteria": criteriaTiers(in.Success, vals, rules, cfg.Stars.MaxStars, cfg.Stars.MinOnSuccess, missingPass), "starRules": rules, "minSeconds": minSecondsOf(cfg, st),
		"timeSavedPercent": math.Round(vals["timeSavedPercent"]*100) / 100, "replayable": replayable(st, row),
		"rewardLimited": nil, "nextRewardAt": iso(nextRewardAt(g, st.Cooldown, now)), "rewardsLeftToday": rewardsLeft(g, st.DailyLimit, now, s.zone(cfg)),
		"serverSeconds": math.Round(serverSeconds*100) / 100, "replayed": false,
	}
	if limited != "" {
		res["rewardLimited"] = limited
	}
	return res, pl, nil
}

func (s *service) takeGroup(ctx context.Context, tx *sql.Tx, p *player, key string, cooldown, daily int, now time.Time) (string, error) {
	if _, err := tx.ExecContext(ctx, `INSERT INTO campaign_reward_group (user_id, group_key) VALUES ($1,$2) ON CONFLICT DO NOTHING`, p.userID, key); err != nil {
		return "", err
	}
	var st groupState
	var at sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT last_at, day, count FROM campaign_reward_group WHERE user_id=$1 AND group_key=$2 FOR UPDATE`, p.userID, key).Scan(&at, &st.Day, &st.Count); err != nil {
		return "", err
	}
	if at.Valid {
		v := at.Time
		st.LastAt = &v
	}
	loc := s.zone(p.cfg)
	if why := limitCheck(st, cooldown, daily, now, loc); why != "" {
		p.groups[key] = st
		return why, nil
	}
	day := dayKey(now, loc)
	if st.Day != day {
		st.Day, st.Count = day, 0
	}
	st.Count++
	st.LastAt = &now
	p.groups[key] = st
	_, err := tx.ExecContext(ctx, `UPDATE campaign_reward_group SET last_at=$3, day=$4, count=$5 WHERE user_id=$1 AND group_key=$2`, p.userID, key, now, st.Day, st.Count)
	return "", err
}

func (s *service) addXP(ctx context.Context, tx *sql.Tx, userID, metaKey string, xp int, total *int64) error {
	if xp > 0 {
		if _, err := tx.ExecContext(ctx, `INSERT INTO campaign_user (user_id, xp) VALUES ($1,$2) ON CONFLICT (user_id) DO UPDATE SET xp=campaign_user.xp+EXCLUDED.xp, updated_at=now()`, userID, xp); err != nil {
			return err
		}
		if metaKey != "" {
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_user_meta (user_id, meta_id, value) SELECT $1, id, default_value + $3::numeric FROM economy_meta WHERE key=$2
				ON CONFLICT (user_id, meta_id) DO UPDATE SET value=economy_user_meta.value + $3::numeric, updated_at=now()`, userID, metaKey, xp); err != nil {
				return err
			}
		}
	}
	return tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT xp FROM campaign_user WHERE user_id=$1),0)`, userID).Scan(total)
}

func itoa(n int) string { return ftoa(float64(n)) }

func ftoa(f float64) string {
	b, _ := json.Marshal(f)
	return string(b)
}

func stageTyped(chapter, stage string, success bool, stars int, paid map[string]int64) []gameevents.Typed {
	status := "Complete"
	if !success {
		status = "Fail"
	}
	score := int64(stars)
	out := []gameevents.Typed{gameevents.Progression(status, chapter, stage, "", &score)}
	keys := make([]string, 0, len(paid))
	for k, v := range paid {
		if v > 0 {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		out = append(out, gameevents.Resource("Source", k, float64(paid[k]), "Stage", stage))
	}
	return out
}
