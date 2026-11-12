package main

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/inapp"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

var achievementClaimNS = uuid.Must(uuid.FromString("8b6f3f0e-2c4d-4f4a-9a1e-3d7c5b2a9e44"))

const maxAchievementTiers = 10

type achievementTier struct {
	ID        string
	Tier      int
	Threshold float64
	Name      string
	IconURL   string
	PrizeID   string
	GrantMode string
}

type achievementDef struct {
	ID            string
	Key           string
	Name          string
	Description   string
	Hint          string
	IconURL       string
	LockedIconURL string
	CategoryID    string
	TagIDs        []string
	Rarity        string
	Points        int
	Hidden        bool
	Source        string
	EventTypeID   string
	Metric        string
	ValueMin      *float64
	ValueMax      *float64
	TaskID        string
	GroupID       string
	Notify        bool
	Sort          int
	Active        bool
	Tiers         []achievementTier
}

const achievementColumns = `a.id::text, a.key, a.name::text, a.description::text, a.hint::text, a.icon_url, a.locked_icon_url, COALESCE(a.category_id::text,''), a.tag_ids::text,
	a.rarity, a.points, a.is_hidden, a.source, COALESCE(a.event_type_id::text,''), a.metric, a.value_min::float8, a.value_max::float8, COALESCE(a.task_id::text,''),
	COALESCE(a.task_group_id::text,''), a.notify, a.sort, a.is_active`

func loadAchievements(ctx context.Context, q dbRunner, where string, args ...any) ([]*achievementDef, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+achievementColumns+` FROM economy_achievement a `+where+` ORDER BY a.sort, a.created_at, a.id`, args...)
	if err != nil {
		return nil, err
	}
	var out []*achievementDef
	byID := map[string]*achievementDef{}
	for rows.Next() {
		var a achievementDef
		var tags string
		var vmin, vmax sql.NullFloat64
		if err := rows.Scan(&a.ID, &a.Key, &a.Name, &a.Description, &a.Hint, &a.IconURL, &a.LockedIconURL, &a.CategoryID, &tags, &a.Rarity, &a.Points, &a.Hidden, &a.Source,
			&a.EventTypeID, &a.Metric, &vmin, &vmax, &a.TaskID, &a.GroupID, &a.Notify, &a.Sort, &a.Active); err != nil {
			rows.Close()
			return nil, err
		}
		a.TagIDs = parseTextArray(tags)
		if vmin.Valid {
			a.ValueMin = &vmin.Float64
		}
		if vmax.Valid {
			a.ValueMax = &vmax.Float64
		}
		out = append(out, &a)
		byID[a.ID] = &a
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return out, nil
	}
	ids := make([]string, 0, len(out))
	for _, a := range out {
		ids = append(ids, a.ID)
	}
	trows, err := q.QueryContext(ctx, `SELECT achievement_id::text, id::text, tier, threshold::float8, name::text, icon_url, COALESCE(prize_id::text,''), grant_mode
		FROM economy_achievement_tier WHERE achievement_id::text = ANY($1) ORDER BY achievement_id, tier`, ids)
	if err != nil {
		return nil, err
	}
	defer trows.Close()
	for trows.Next() {
		var aid string
		var t achievementTier
		if err := trows.Scan(&aid, &t.ID, &t.Tier, &t.Threshold, &t.Name, &t.IconURL, &t.PrizeID, &t.GrantMode); err != nil {
			return nil, err
		}
		if a := byID[aid]; a != nil {
			a.Tiers = append(a.Tiers, t)
		}
	}
	return out, trows.Err()
}

func (s *service) findAchievement(ctx context.Context, idOrKey string) (*achievementDef, error) {
	idOrKey = strings.TrimSpace(idOrKey)
	var list []*achievementDef
	var err error
	if _, uerr := uuid.FromString(idOrKey); uerr == nil {
		list, err = loadAchievements(ctx, s.db, `WHERE a.id=$1`, idOrKey)
	} else if idOrKey != "" {
		list, err = loadAchievements(ctx, s.db, `WHERE a.key=$1`, idOrKey)
	}
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, notFound("ACHIEVEMENT_NOT_FOUND")
	}
	return list[0], nil
}

func (a *achievementDef) maxTier() int {
	if len(a.Tiers) == 0 {
		return 1
	}
	return a.Tiers[len(a.Tiers)-1].Tier
}

func (a *achievementDef) tier(n int) *achievementTier {
	for i := range a.Tiers {
		if a.Tiers[i].Tier == n {
			return &a.Tiers[i]
		}
	}
	return nil
}

func (a *achievementDef) reached(value float64) int {
	n := 0
	for _, t := range a.Tiers {
		if value+1e-9 >= t.Threshold {
			n = t.Tier
		}
	}
	return n
}

func (a *achievementDef) valueAllowed(v float64) bool {
	if a.ValueMin != nil && v < *a.ValueMin {
		return false
	}
	if a.ValueMax != nil && v > *a.ValueMax {
		return false
	}
	return true
}

func achievementClaimID(userID, tierID string) string {
	return uuid.NewV5(achievementClaimNS, userID+":"+tierID).String()
}

type achievementChange struct {
	AchievementID string  `json:"achievementId"`
	Key           string  `json:"key"`
	From          int     `json:"from"`
	To            int     `json:"to"`
	Value         float64 `json:"value"`
	Unlocked      bool    `json:"unlocked"`
	Source        string  `json:"source"`
	def           *achievementDef
}

func raiseAchievement(ctx context.Context, q dbRunner, a *achievementDef, userID string, to int, source, ref string, now time.Time) (int, bool, error) {
	if _, err := q.ExecContext(ctx, `INSERT INTO economy_user_achievement (achievement_id, user_id, created_at, updated_at) VALUES ($1,$2,$3,$3) ON CONFLICT DO NOTHING`, a.ID, userID, now); err != nil {
		return 0, false, err
	}
	var from int
	if err := q.QueryRowContext(ctx, `SELECT tier FROM economy_user_achievement WHERE achievement_id=$1 AND user_id=$2 FOR UPDATE`, a.ID, userID).Scan(&from); err != nil {
		return 0, false, err
	}
	if from >= to {
		return from, false, nil
	}
	if _, err := q.ExecContext(ctx, `UPDATE economy_user_achievement SET tier=$3, unlocked_at=COALESCE(unlocked_at,$4), tier_at=$4,
		source=CASE WHEN tier=0 THEN $5 ELSE source END, source_ref=CASE WHEN tier=0 THEN $6 ELSE source_ref END, updated_at=$4
		WHERE achievement_id=$1 AND user_id=$2`, a.ID, userID, to, now, source, truncate(ref, 200)); err != nil {
		return 0, false, err
	}
	return from, true, nil
}

func (s *service) applyAchievements(ctx context.Context, tx *sql.Tx, userID, typeID string, value float64, now time.Time, out *eventOutcome) error {
	list, err := loadAchievements(ctx, tx, `WHERE a.event_type_id=$1 AND a.source='event' AND a.is_active`, typeID)
	if err != nil || len(list) == 0 {
		return err
	}
	for _, a := range list {
		if !a.valueAllowed(value) {
			continue
		}
		var count int64
		var sum, max float64
		var tier int
		if err := tx.QueryRowContext(ctx, `INSERT INTO economy_user_achievement AS p (achievement_id, user_id, count, sum, max, created_at, updated_at) VALUES ($1,$2,1,$3::numeric,$3::numeric,$4,$4)
			ON CONFLICT (achievement_id, user_id) DO UPDATE SET count=p.count+1, sum=p.sum+EXCLUDED.sum, max=GREATEST(p.max, EXCLUDED.max), updated_at=EXCLUDED.updated_at
			RETURNING count, sum::float8, COALESCE(max,0)::float8, tier`, a.ID, userID, formatAmount(value), now).Scan(&count, &sum, &max, &tier); err != nil {
			return err
		}
		measured := taskMeasure(a.Metric, count, sum, max)
		reached := a.reached(measured)
		if reached <= tier {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_user_achievement SET tier=$3, unlocked_at=COALESCE(unlocked_at,$4), tier_at=$4,
			source=CASE WHEN tier=0 THEN 'event' ELSE source END, source_ref=CASE WHEN tier=0 THEN $5 ELSE source_ref END WHERE achievement_id=$1 AND user_id=$2`,
			a.ID, userID, reached, now, out.EventID); err != nil {
			return err
		}
		ch := achievementChange{AchievementID: a.ID, Key: a.Key, From: tier, To: reached, Value: measured, Unlocked: tier == 0, Source: "event", def: a}
		out.Achievements = append(out.Achievements, ch)
		out.achPending = append(out.achPending, ch)
	}
	return nil
}

func (s *service) applyTaskAchievements(ctx context.Context, q dbRunner, userID string, t *taskDef, scope string, now time.Time, out *eventOutcome) error {
	list, err := loadAchievements(ctx, q, `WHERE a.is_active AND a.source='task' AND (a.task_id=$1 OR a.task_group_id::text=$2)`, t.ID, t.GroupID)
	if err != nil || len(list) == 0 {
		return err
	}
	groupDone := false
	checked := false
	for _, a := range list {
		if a.GroupID != "" {
			if !checked {
				checked = true
				var total, done int64
				if err := q.QueryRowContext(ctx, `SELECT count(*), count(ut.task_id) FROM economy_task m
					LEFT JOIN economy_user_task ut ON ut.task_id=m.id AND ut.user_id=$2 AND ut.scope=$3 AND ut.status<>'active'
					WHERE m.group_id=$1 AND m.parent_id IS NULL AND m.is_active`, a.GroupID, userID, scope).Scan(&total, &done); err != nil {
					return err
				}
				groupDone = total > 0 && done >= total
			}
			if !groupDone {
				continue
			}
		}
		from, changed, err := raiseAchievement(ctx, q, a, userID, 1, "task", t.Key, now)
		if err != nil {
			return err
		}
		if changed {
			ch := achievementChange{AchievementID: a.ID, Key: a.Key, From: from, To: 1, Value: 1, Unlocked: from == 0, Source: "task", def: a}
			out.Achievements = append(out.Achievements, ch)
			out.achPending = append(out.achPending, ch)
		}
	}
	return nil
}

func (s *service) grantAchievementItems(ctx context.Context, tx *sql.Tx, userID string, keys []string, rowID string) ([]achievementChange, error) {
	var out []achievementChange
	now := s.now()
	for _, key := range keys {
		list, err := loadAchievements(ctx, tx, `WHERE a.key=$1`, key)
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			continue
		}
		a := list[0]
		from, changed, err := raiseAchievement(ctx, tx, a, userID, 1, "prize", rowID, now)
		if err != nil {
			return nil, err
		}
		if changed {
			out = append(out, achievementChange{AchievementID: a.ID, Key: a.Key, From: from, To: 1, Value: 1, Unlocked: from == 0, Source: "prize", def: a})
		}
	}
	return out, nil
}

type eventDepthKey struct{}

func withEventDepth(ctx context.Context, depth int) context.Context {
	return context.WithValue(ctx, eventDepthKey{}, depth)
}

func eventDepth(ctx context.Context) int {
	if d, ok := ctx.Value(eventDepthKey{}).(int); ok {
		return d
	}
	return 1
}

func (s *service) settleAchievements(ctx context.Context, userID string, changes []achievementChange, depth int, out *eventOutcome) {
	for _, ch := range changes {
		a := ch.def
		if a == nil {
			continue
		}
		for _, t := range a.Tiers {
			if t.Tier <= ch.From || t.Tier > ch.To || t.PrizeID == "" || t.GrantMode != "auto" {
				continue
			}
			ref, err := s.ensureAutoTierClaim(ctx, userID, a, t)
			if err != nil {
				s.logger.Warn("Economy achievement reward %s/%d for %s not created: %v", a.Key, t.Tier, userID, err)
				continue
			}
			if ref == nil {
				continue
			}
			if depth > maxEventDepth {
				if out != nil {
					out.Claims = append(out.Claims, *ref)
				}
				continue
			}
			extra, err := s.grantClaim(ctx, ref.ClaimID, depth+1)
			if err != nil {
				ref.Status = "failed"
				s.logger.Warn("Economy achievement reward %s not granted yet: %v", ref.ClaimID, err)
			} else {
				ref.Status = "granted"
			}
			if out != nil {
				out.Claims = append(out.Claims, *ref)
				out.merge(extra)
			}
		}
		action := "tier"
		if ch.Unlocked {
			action = "unlocked"
		}
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindAchievement, Action: action, RefID: a.Key, Actor: actorFor(ch.Source),
			Data: map[string]any{"achievement": a.Key, "from": ch.From, "tier": ch.To, "source": ch.Source}})
		if ch.Unlocked && depth <= maxEventDepth {
			v := float64(a.Points)
			extra, err := s.recordEvent(ctx, eventInput{UserID: userID, Key: "achievement_unlocked", Value: &v, Metadata: map[string]any{"key": a.Key}, Source: "server", depth: depth + 1})
			if err != nil {
				if app, ok := asApp(err); !ok || app.Code != "EVENT_TYPE_NOT_FOUND" {
					s.logger.Warn("Economy event achievement_unlocked for %s not recorded: %v", userID, err)
				}
			} else if out != nil {
				out.merge(extra)
			}
		}
		if a.Notify {
			s.notifyAchievement(ctx, userID, a, ch)
		}
	}
}

func actorFor(source string) string {
	if source == "admin" {
		return "admin"
	}
	return "system"
}

func (s *service) ensureAutoTierClaim(ctx context.Context, userID string, a *achievementDef, t achievementTier) (*claimRef, error) {
	id := achievementClaimID(userID, t.ID)
	ref, err := insertClaimWithID(ctx, s.db, id, userID, "achievement", a.ID, t.ID, "t"+strconv.Itoa(t.Tier), t.PrizeID, "auto", float64(t.Tier))
	if err != nil || ref != nil {
		return ref, err
	}
	var status, mode string
	err = s.db.QueryRowContext(ctx, `SELECT status, grant_mode FROM economy_reward_claim WHERE id=$1`, id).Scan(&status, &mode)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil || status == "granted" || mode != "auto" {
		return nil, err
	}
	return &claimRef{ClaimID: id, Kind: "achievement", PrizeID: t.PrizeID, Mode: "auto", Status: status}, nil
}

var achievementNoticeTexts = map[string]map[string]string{
	"title":        {"fa": "دستاورد تازه", "en": "New achievement", "ar": "إنجاز جديد"},
	"body.unlock":  {"fa": "دستاورد «{name}» باز شد.", "en": "You unlocked “{name}”.", "ar": "فتحت إنجاز «{name}»."},
	"body.tier":    {"fa": "به ردهٔ {tier} از «{name}» رسیدی.", "en": "You reached tier {tier} of “{name}”.", "ar": "وصلت إلى المستوى {tier} من «{name}»."},
	"action.label": {"fa": "دیدن دستاوردها", "en": "Open achievements", "ar": "عرض الإنجازات"},
}

func (s *service) notifyAchievement(ctx context.Context, userID string, a *achievementDef, ch achievementChange) {
	if !inapp.Available() {
		return
	}
	names := localized.Parse(a.Name)
	key := "body.tier"
	if ch.Unlocked {
		key = "body.unlock"
	}
	body := map[string]string{}
	for lang, tpl := range achievementNoticeTexts[key] {
		name := names[lang]
		if name == "" {
			name = names["fa"]
		}
		if name == "" {
			name = a.Key
		}
		tier := strconv.Itoa(ch.To)
		if lang == "fa" {
			tier = faDigits(tier)
		}
		body[lang] = strings.NewReplacer("{name}", name, "{tier}", tier).Replace(tpl)
	}
	if _, err := inapp.Send(ctx, inapp.Message{UserIDs: []string{userID}, Title: achievementNoticeTexts["title"], Body: body,
		Action: &inapp.Action{Type: "goto", Value: "achievements", Label: achievementNoticeTexts["action.label"]}, CreatedBy: "economy:achievement"}); err != nil {
		s.logger.Warn("Economy achievement notice for %s not sent: %v", userID, err)
	}
}

type tierClaim struct {
	Tier    int
	ClaimID string
	PrizeID string
	Rewards []prizeItem
}

func (s *service) userAchievementTier(ctx context.Context, userID, achievementID string) (int, error) {
	var tier int
	err := s.db.QueryRowContext(ctx, `SELECT tier FROM economy_user_achievement WHERE achievement_id=$1 AND user_id=$2`, achievementID, userID).Scan(&tier)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return tier, err
}

func (s *service) claimAchievement(ctx context.Context, userID string, a *achievementDef, tierNum int) ([]tierClaim, *eventOutcome, error) {
	current, err := s.userAchievementTier(ctx, userID, a.ID)
	if err != nil {
		return nil, nil, err
	}
	if current == 0 {
		return nil, nil, conflict("ACHIEVEMENT_LOCKED", "this achievement is not unlocked yet")
	}
	var todo []achievementTier
	if tierNum > 0 {
		t := a.tier(tierNum)
		if t == nil {
			return nil, nil, notFound("TIER_NOT_FOUND")
		}
		if tierNum > current {
			return nil, nil, conflict("TIER_NOT_REACHED", "this tier is not reached yet")
		}
		if t.PrizeID == "" {
			return nil, nil, conflict("TIER_HAS_NO_PRIZE", "this tier has no prize")
		}
		todo = append(todo, *t)
	} else {
		for _, t := range a.Tiers {
			if t.Tier <= current && t.PrizeID != "" {
				todo = append(todo, t)
			}
		}
	}
	extra := newOutcome()
	var done []tierClaim
	already := 0
	for _, t := range todo {
		res, err := s.claimTier(ctx, userID, a, t)
		if err != nil {
			if app, ok := asApp(err); ok && app.Code == "ALREADY_CLAIMED" {
				already++
				if tierNum > 0 {
					return nil, nil, err
				}
				continue
			}
			if len(done) > 0 {
				break
			}
			return nil, nil, err
		}
		done = append(done, *res)
	}
	if len(done) == 0 {
		if already > 0 {
			return nil, nil, conflict("ALREADY_CLAIMED", "this reward was already received")
		}
		return nil, nil, conflict("NOTHING_TO_CLAIM", "there is no prize to collect")
	}
	var metaKeys []string
	for _, d := range done {
		for _, item := range d.Rewards {
			if item.Type == "meta" {
				metaKeys = append(metaKeys, item.Key)
			}
		}
	}
	if len(metaKeys) > 0 {
		if ids, err := s.metaIDsByKey(ctx, metaKeys); err == nil {
			lv, err := s.evaluateLevels(ctx, userID, ids, 2)
			if err != nil {
				s.logger.Warn("Economy level check after achievement reward failed: %v", err)
			}
			extra.merge(lv)
		}
	}
	return done, extra, nil
}

func (s *service) claimTier(ctx context.Context, userID string, a *achievementDef, t achievementTier) (*tierClaim, error) {
	cid := achievementClaimID(userID, t.ID)
	var status string
	err := s.db.QueryRowContext(ctx, `SELECT status FROM economy_reward_claim WHERE id=$1`, cid).Scan(&status)
	createdHere := false
	if errors.Is(err, sql.ErrNoRows) {
		ref, err := insertClaimWithID(ctx, s.db, cid, userID, "achievement", a.ID, t.ID, "t"+strconv.Itoa(t.Tier), t.PrizeID, t.GrantMode, float64(t.Tier))
		if err != nil {
			return nil, err
		}
		createdHere = ref != nil
	} else if err != nil {
		return nil, err
	} else if status == "granted" {
		return nil, conflict("ALREADY_CLAIMED", "this reward was already received")
	}
	var res tierClaim
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var prizeID, st string
		err := tx.QueryRowContext(ctx, `SELECT prize_id::text, status FROM economy_reward_claim WHERE id=$1 AND user_id=$2 FOR UPDATE`, cid, userID).Scan(&prizeID, &st)
		if errors.Is(err, sql.ErrNoRows) {
			return conflict("NOTHING_TO_CLAIM", "there is no prize to collect")
		}
		if err != nil {
			return err
		}
		if st == "granted" {
			return conflict("ALREADY_CLAIMED", "this reward was already received")
		}
		rewards, err := s.grant(withEventDepth(ctx, 2), grantRequest{UserID: userID, PrizeID: prizeID, SourceType: "achievement", SourceID: cid, Reason: "achievement_reward"})
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_reward_claim SET status='granted', error=NULL, granted_at=now(), attempted_at=now(), attempts=attempts+1 WHERE id=$1`, cid); err != nil {
			return err
		}
		if rewards == nil {
			rewards = []prizeItem{}
		}
		res = tierClaim{Tier: t.Tier, ClaimID: cid, PrizeID: prizeID, Rewards: rewards}
		return nil
	})
	if err != nil {
		if app, ok := asApp(err); ok && (app.Code == "ALREADY_CLAIMED" || app.Code == "NOTHING_TO_CLAIM") {
			return nil, err
		}
		msg := err.Error()
		if app, ok := asApp(err); ok {
			msg = app.Code
		}
		if createdHere && t.GrantMode == "claim" {
			_, _ = s.db.ExecContext(ctx, `DELETE FROM economy_reward_claim WHERE id=$1 AND status<>'granted' AND grant_mode='claim'`, cid)
		} else {
			_, _ = s.db.ExecContext(ctx, `UPDATE economy_reward_claim SET status='failed', error=$2, attempted_at=now() WHERE id=$1 AND status<>'granted'`, cid, truncate(msg, 500))
		}
		return nil, err
	}
	return &res, nil
}

func (s *service) adminGiveAchievement(ctx context.Context, userID string, a *achievementDef, tier int, adminID string) (*eventOutcome, error) {
	if tier < 1 {
		tier = 1
	}
	if a.tier(tier) == nil {
		return nil, notFound("TIER_NOT_FOUND")
	}
	out := newOutcome()
	var changes []achievementChange
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id=$1)`, userID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return notFound("USER_NOT_FOUND")
		}
		from, changed, err := raiseAchievement(ctx, tx, a, userID, tier, "admin", adminID, s.now())
		if err != nil {
			return err
		}
		if !changed {
			return conflict("ALREADY_UNLOCKED", "the player already has this achievement at this tier")
		}
		changes = append(changes, achievementChange{AchievementID: a.ID, Key: a.Key, From: from, To: tier, Value: 1, Unlocked: from == 0, Source: "admin", def: a})
		return nil
	})
	if err != nil {
		return nil, err
	}
	out.Achievements = append(out.Achievements, changes...)
	s.settleAchievements(ctx, userID, changes, 1, out)
	return out, nil
}

func (s *service) adminRevokeAchievement(ctx context.Context, userID string, a *achievementDef, adminID string) error {
	var tier int
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, `DELETE FROM economy_user_achievement WHERE achievement_id=$1 AND user_id=$2 RETURNING tier`, a.ID, userID).Scan(&tier)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && tier == 0) {
			return conflict("NOT_UNLOCKED", "the player does not have this achievement")
		}
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `DELETE FROM economy_reward_claim WHERE kind='achievement' AND owner_id=$1 AND user_id=$2 AND status<>'granted'`, a.ID, userID)
		return err
	})
	if err != nil {
		return err
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindAchievement, Action: "revoked", RefID: a.Key, Actor: "admin", Data: map[string]any{"achievement": a.Key, "tier": tier, "admin": adminID}})
	return nil
}

func (s *service) healAchievementClaims(ctx context.Context, userID string, defs []*achievementDef, tiers map[string]int, claims map[string]achievementClaimRow) bool {
	healed := false
	for _, a := range defs {
		if !a.Active {
			continue
		}
		cur := tiers[a.ID]
		var missing []achievementTier
		for _, t := range a.Tiers {
			if t.Tier <= cur && t.PrizeID != "" && t.GrantMode == "auto" {
				if _, ok := claims[t.ID]; !ok {
					missing = append(missing, t)
				}
			}
		}
		if len(missing) == 0 {
			continue
		}
		healed = true
		for _, t := range missing {
			ref, err := s.ensureAutoTierClaim(ctx, userID, a, t)
			if err != nil || ref == nil {
				continue
			}
			if _, err := s.grantClaim(ctx, ref.ClaimID, 2); err != nil {
				s.logger.Warn("Economy achievement reward %s not granted yet: %v", ref.ClaimID, err)
			}
		}
	}
	return healed
}

type achievementClaimRow struct {
	ID     string
	Status string
	Mode   string
}

func (s *service) achievementClaims(ctx context.Context, userID string) (map[string]achievementClaimRow, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT ref_id::text, id::text, status, grant_mode FROM economy_reward_claim WHERE user_id=$1 AND kind='achievement'`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]achievementClaimRow{}
	for rows.Next() {
		var ref string
		var c achievementClaimRow
		if err := rows.Scan(&ref, &c.ID, &c.Status, &c.Mode); err != nil {
			return nil, err
		}
		out[ref] = c
	}
	return out, rows.Err()
}

type userAchievementRow struct {
	Count      int64
	Sum        float64
	Max        float64
	Tier       int
	UnlockedAt sql.NullTime
	TierAt     sql.NullTime
	Source     string
	SourceRef  string
}

func (s *service) userAchievementRows(ctx context.Context, userID string) (map[string]userAchievementRow, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT achievement_id::text, count, sum::float8, COALESCE(max,0)::float8, tier, unlocked_at, tier_at, source, source_ref FROM economy_user_achievement WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]userAchievementRow{}
	for rows.Next() {
		var id string
		var r userAchievementRow
		if err := rows.Scan(&id, &r.Count, &r.Sum, &r.Max, &r.Tier, &r.UnlockedAt, &r.TierAt, &r.Source, &r.SourceRef); err != nil {
			return nil, err
		}
		out[id] = r
	}
	return out, rows.Err()
}

func (s *service) sweepAchievements(ctx context.Context) {
	_, _ = s.db.ExecContext(ctx, `DELETE FROM economy_user_achievement WHERE user_id IN (SELECT DISTINCT u.user_id FROM economy_user_achievement u LEFT JOIN users x ON x.id=u.user_id WHERE x.id IS NULL LIMIT 200)`)
}

func faDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return '۰' + (r - '0')
		}
		return r
	}, s)
}

func sortAchievements(list []*achievementDef) {
	sort.SliceStable(list, func(i, j int) bool { return list[i].Sort < list[j].Sort })
}
