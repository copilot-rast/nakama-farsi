package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
)

const (
	maxEventDepth    = 6
	maxEventMetadata = 2048
	defaultTimezone  = "Asia/Tehran"
)

var (
	eventKeyPattern        = regexp.MustCompile(`^[a-z][a-z0-9_.]{1,47}$`)
	builtinEventKeyPattern = regexp.MustCompile(`^[A-Z][A-Za-z]{1,31}:[A-Z][A-Za-z]{1,31}$`)
)

func validEventKey(key string) bool {
	return eventKeyPattern.MatchString(key) || builtinEventKeyPattern.MatchString(key)
}

type eventSettings struct {
	Timezone        string `json:"timezone"`
	WeekStart       int    `json:"weekStart"`
	RetentionDays   int    `json:"retentionDays"`
	PlayerRateLimit int    `json:"playerRateLimit"`
	// RelationChainDepth is how many meta relations may trigger one another
	// in a row; 0 means a relation's targets never fire further relations.
	RelationChainDepth int `json:"relationChainDepth"`
	// RelationPrizeMissing decides what a relation does when its prize group
	// has nothing to draw: complete without a prize, or wait and retry.
	RelationPrizeMissing string `json:"relationPrizeMissing"`
	UpdatedAt            string `json:"updatedAt,omitempty"`
	loc                  *time.Location
}

func loadEventSettings(ctx context.Context, q queryer) (eventSettings, error) {
	st := eventSettings{Timezone: defaultTimezone, WeekStart: 6, RetentionDays: 30, PlayerRateLimit: 120, RelationChainDepth: 8, RelationPrizeMissing: "complete"}
	var updated time.Time
	err := q.QueryRowContext(ctx, `SELECT timezone, week_start, retention_days, player_rate_limit, relation_chain_depth, relation_prize_missing, updated_at FROM economy_event_settings WHERE id=1`).
		Scan(&st.Timezone, &st.WeekStart, &st.RetentionDays, &st.PlayerRateLimit, &st.RelationChainDepth, &st.RelationPrizeMissing, &updated)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return st, err
	}
	if err == nil {
		st.UpdatedAt = updated.UTC().Format(time.RFC3339)
	}
	loc, lerr := time.LoadLocation(st.Timezone)
	if lerr != nil {
		loc, _ = time.LoadLocation(defaultTimezone)
	}
	st.loc = loc
	return st, nil
}

func (st eventSettings) clock() eventClock {
	loc := st.loc
	if loc == nil {
		loc = time.UTC
	}
	return eventClock{loc: loc, weekStart: time.Weekday(st.WeekStart)}
}

type eventClock struct {
	loc       *time.Location
	weekStart time.Weekday
}

var epochDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)

func (c eventClock) civil(t time.Time) time.Time {
	y, m, d := t.In(c.loc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func (c eventClock) periodStart(period string, t time.Time) time.Time {
	switch period {
	case "day":
		return c.civil(t)
	case "week":
		day := c.civil(t)
		back := (int(t.In(c.loc).Weekday()) - int(c.weekStart) + 7) % 7
		return day.AddDate(0, 0, -back)
	}
	return epochDate
}

func (c eventClock) instant(civil time.Time) time.Time {
	return time.Date(civil.Year(), civil.Month(), civil.Day(), 0, 0, 0, 0, c.loc)
}

func (c eventClock) periodEnd(period string, t time.Time) *time.Time {
	var next time.Time
	switch period {
	case "day":
		next = c.instant(c.periodStart(period, t).AddDate(0, 0, 1))
	case "week":
		next = c.instant(c.periodStart(period, t).AddDate(0, 0, 7))
	default:
		return nil
	}
	return &next
}

func floorDiv(a, b int64) int64 {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

func (c eventClock) dayIndex(t time.Time) int64 {
	return floorDiv(c.civil(t).Unix(), 86400)
}

func (c eventClock) streakIndex(t time.Time, periodDays int) int64 {
	if periodDays < 1 {
		periodDays = 1
	}
	return floorDiv(c.dayIndex(t), int64(periodDays))
}

func (c eventClock) streakPeriodStart(idx int64, periodDays int) time.Time {
	if periodDays < 1 {
		periodDays = 1
	}
	return c.instant(epochDate.AddDate(0, 0, int(idx*int64(periodDays))))
}

func dateKey(d time.Time) string {
	return d.Format("2006-01-02")
}

type eventType struct {
	ID            string
	Key           string
	ClientAllowed bool
	Min, Max      *float64
	RateLimit     int
	Active        bool
}

func loadEventType(ctx context.Context, q queryer, key string) (eventType, error) {
	var et eventType
	var min, max sql.NullFloat64
	err := q.QueryRowContext(ctx, `SELECT id::text, key, client_allowed, value_min::float8, value_max::float8, rate_limit, is_active FROM economy_event_type WHERE key=$1`, key).
		Scan(&et.ID, &et.Key, &et.ClientAllowed, &min, &max, &et.RateLimit, &et.Active)
	if errors.Is(err, sql.ErrNoRows) {
		return et, notFound("EVENT_TYPE_NOT_FOUND")
	}
	if err != nil {
		return et, err
	}
	if min.Valid {
		et.Min = &min.Float64
	}
	if max.Valid {
		et.Max = &max.Float64
	}
	return et, nil
}

type eventInput struct {
	UserID   string
	Key      string
	Value    *float64
	Metadata map[string]any
	Source   string
	depth    int
}

type ruleProgress struct {
	RuleID      string     `json:"ruleId"`
	Key         string     `json:"key"`
	Period      string     `json:"period"`
	PeriodStart string     `json:"periodStart"`
	PeriodEnd   *time.Time `json:"periodEndsAt"`
	Measure     string     `json:"measure"`
	Value       float64    `json:"value"`
	Threshold   float64    `json:"threshold"`
	Reached     bool       `json:"reached"`
	NewClaims   int        `json:"newClaims"`
}

type streakChange struct {
	StreakID      string `json:"streakId"`
	Key           string `json:"key"`
	Count         int    `json:"count"`
	Best          int    `json:"best"`
	Advanced      bool   `json:"advanced"`
	Reset         bool   `json:"reset"`
	Floored       bool   `json:"floored"`
	Floor         int    `json:"floor"`
	NextMilestone int    `json:"nextMilestone,omitempty"`
}

type levelChange struct {
	TableID string `json:"tableId"`
	Key     string `json:"key"`
	From    int    `json:"from"`
	To      int    `json:"to"`
}

type claimRef struct {
	ClaimID  string `json:"claimId"`
	Kind     string `json:"kind"`
	PrizeID  string `json:"prizeId"`
	Mode     string `json:"grantMode"`
	Status   string `json:"status"`
	PrizeKey string `json:"prizeKey,omitempty"`
}

type eventOutcome struct {
	EventID      string              `json:"eventId"`
	Key          string              `json:"key"`
	Value        float64             `json:"value"`
	Rules        []ruleProgress      `json:"rules"`
	Metas        map[string]float64  `json:"metas"`
	Streaks      []streakChange      `json:"streaks"`
	Levels       []levelChange       `json:"levels"`
	Claims       []claimRef          `json:"claims"`
	Tasks        []taskChange        `json:"tasks"`
	Achievements []achievementChange `json:"achievements"`
	Relations    []relationChange    `json:"relations"`
	metaIDs      []string
	referrals    []string
	achPending   []achievementChange
}

func newOutcome() *eventOutcome {
	return &eventOutcome{Rules: []ruleProgress{}, Metas: map[string]float64{}, Streaks: []streakChange{}, Levels: []levelChange{}, Claims: []claimRef{}, Tasks: []taskChange{}, Achievements: []achievementChange{}, Relations: []relationChange{}}
}

func (o *eventOutcome) merge(other *eventOutcome) {
	if other == nil {
		return
	}
	o.Rules = append(o.Rules, other.Rules...)
	for k, v := range other.Metas {
		o.Metas[k] = v
	}
	o.Streaks = append(o.Streaks, other.Streaks...)
	o.Levels = append(o.Levels, other.Levels...)
	o.Claims = append(o.Claims, other.Claims...)
	o.Tasks = append(o.Tasks, other.Tasks...)
	o.Achievements = append(o.Achievements, other.Achievements...)
	o.Relations = append(o.Relations, other.Relations...)
	o.referrals = append(o.referrals, other.referrals...)
}

func formatAmount(v float64) string {
	return strconv.FormatFloat(math.Round(v*10000)/10000, 'f', 4, 64)
}

func cleanEventMetadata(in map[string]any) (string, error) {
	if in == nil {
		return "{}", nil
	}
	raw, err := json.Marshal(in)
	if err != nil {
		return "", badRequest("INVALID_METADATA", "metadata", "metadata must be a JSON object")
	}
	if len(raw) > maxEventMetadata {
		return "", badRequest("METADATA_TOO_LARGE", "metadata", "metadata must be at most 2 KB")
	}
	return string(raw), nil
}

func (s *service) recordEvent(ctx context.Context, in eventInput) (*eventOutcome, error) {
	if in.depth > maxEventDepth {
		return newOutcome(), nil
	}
	if _, err := uuid.FromString(in.UserID); err != nil {
		return nil, badRequest("INVALID_ID", "userId", "user id must be a UUID")
	}
	key := strings.TrimSpace(in.Key)
	if !validEventKey(key) {
		return nil, notFound("EVENT_TYPE_NOT_FOUND")
	}
	et, err := loadEventType(ctx, s.db, key)
	if err != nil {
		return nil, err
	}
	if !et.Active {
		return nil, notFound("EVENT_TYPE_NOT_FOUND")
	}
	if in.Source == "client" && !et.ClientAllowed {
		return nil, fail(http.StatusForbidden, "EVENT_NOT_ALLOWED", "key", "this event can only be recorded by the server")
	}
	value := 1.0
	if in.Value != nil {
		value = *in.Value
	}
	if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) >= 1e15 {
		return nil, badRequest("INVALID_VALUE", "value", "value must be a finite number")
	}
	value = math.Round(value*10000) / 10000
	if et.Min != nil && value < *et.Min {
		return nil, badRequest("VALUE_OUT_OF_RANGE", "value", "value is below the allowed minimum")
	}
	if et.Max != nil && value > *et.Max {
		return nil, badRequest("VALUE_OUT_OF_RANGE", "value", "value is above the allowed maximum")
	}
	meta, err := cleanEventMetadata(in.Metadata)
	if err != nil {
		return nil, err
	}
	settings, err := loadEventSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if in.Source == "client" {
		d, err := clusterkit.Allow(ctx, s.db, "ev:"+in.UserID+":"+et.ID, int64(et.RateLimit), time.Minute)
		if err != nil {
			return nil, err
		}
		if d.Allowed {
			d, err = clusterkit.Allow(ctx, s.db, "ev:"+in.UserID, int64(settings.PlayerRateLimit), time.Minute)
			if err != nil {
				return nil, err
			}
		}
		if !d.Allowed {
			return nil, fail(http.StatusTooManyRequests, "TOO_MANY_EVENTS", "", "too many events, try again in a minute")
		}
	}
	now := s.now()
	clk := settings.clock()
	out := newOutcome()
	out.Key, out.Value = et.Key, value
	out.EventID = uuid.Must(uuid.NewV4()).String()
	var tags map[string]bool
	userTags := func() (map[string]bool, error) {
		if tags != nil {
			return tags, nil
		}
		t, err := s.userTagIDs(ctx, in.UserID)
		tags = t
		return t, err
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_player_event (id, user_id, type_id, value, metadata, source, created_at) VALUES ($1,$2,$3,$4::numeric,$5::jsonb,$6,$7)`,
			out.EventID, in.UserID, et.ID, formatAmount(value), meta, in.Source, now); err != nil {
			if isForeignKeyViolation(err) {
				return notFound("USER_NOT_FOUND")
			}
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_player_event_total AS t (user_id, type_id, count, sum, max, first_at, last_at) VALUES ($1,$2,1,$3::numeric,$3::numeric,$4,$4)
			ON CONFLICT (user_id, type_id) DO UPDATE SET count=t.count+1, sum=t.sum+EXCLUDED.sum, max=GREATEST(t.max, EXCLUDED.max), last_at=EXCLUDED.last_at`,
			in.UserID, et.ID, formatAmount(value), now); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_player_event_day AS d (user_id, type_id, day, count, sum, max) VALUES ($1,$2,$3::date,1,$4::numeric,$4::numeric)
			ON CONFLICT (user_id, type_id, day) DO UPDATE SET count=d.count+1, sum=d.sum+EXCLUDED.sum, max=GREATEST(d.max, EXCLUDED.max)`,
			in.UserID, et.ID, dateKey(clk.civil(now)), formatAmount(value)); err != nil {
			return err
		}
		if err := s.applyEventRules(ctx, tx, in.UserID, et.ID, value, now, clk, userTags, out); err != nil {
			return err
		}
		if err := s.applyTasks(ctx, tx, in.UserID, et.ID, value, now, clk, userTags, out); err != nil {
			return err
		}
		if err := s.applyAchievements(ctx, tx, in.UserID, et.ID, value, now, out); err != nil {
			return err
		}
		if err := s.applyMetaRules(ctx, tx, in.UserID, et.ID, value, now, userTags, out); err != nil {
			return err
		}
		if err := s.applyStreaks(ctx, tx, in.UserID, et.ID, now, clk, out); err != nil {
			return err
		}
		return s.applyReferralGoal(ctx, tx, in.UserID, et.ID, out)
	})
	if err != nil {
		return nil, err
	}
	s.settleTasks(ctx, in.UserID, out)
	s.settleOutcome(ctx, in.UserID, out, in.depth)
	if len(out.referrals) > 0 {
		s.noteReferralDone(ctx, out.referrals)
		out.referrals = nil
	}
	return out, nil
}

func (s *service) settleOutcome(ctx context.Context, userID string, out *eventOutcome, depth int) {
	if len(out.achPending) > 0 {
		pending := out.achPending
		out.achPending = nil
		s.settleAchievements(ctx, userID, pending, depth, out)
	}
	for i := range out.Claims {
		c := &out.Claims[i]
		if c.Mode != "auto" || c.Status != "ready" {
			continue
		}
		extra, err := s.grantClaim(ctx, c.ClaimID, depth+1)
		if err != nil {
			c.Status = "failed"
			s.logger.Warn("Economy event reward %s not granted yet: %v", c.ClaimID, err)
			continue
		}
		c.Status = "granted"
		out.merge(extra)
	}
	if len(out.metaIDs) > 0 {
		ids := out.metaIDs
		out.metaIDs = nil
		extra, err := s.evaluateLevels(ctx, userID, ids, depth+1)
		if err != nil {
			s.logger.Warn("Economy level check failed for %s: %v", userID, err)
		}
		out.merge(extra)
	}
}

func (s *service) recordServerEvent(ctx context.Context, userID, key string, value float64, metadata map[string]any) {
	v := value
	if _, err := s.recordEvent(ctx, eventInput{UserID: userID, Key: key, Value: &v, Metadata: metadata, Source: "server"}); err != nil {
		if app, ok := asApp(err); ok && app.Code == "EVENT_TYPE_NOT_FOUND" {
			return
		}
		s.logger.Warn("Economy event %s for %s not recorded: %v", key, userID, err)
	}
}

func (s *service) playerEventRecorder() playerevents.Recorder {
	return func(ctx context.Context, ev playerevents.Event) (playerevents.Result, error) {
		v := ev.Value
		out, err := s.recordEvent(ctx, eventInput{UserID: ev.UserID, Key: ev.Key, Value: &v, Metadata: ev.Metadata, Source: "server"})
		if err != nil {
			if app, ok := asApp(err); ok {
				return nil, errors.New(app.Code)
			}
			return nil, err
		}
		raw, _ := json.Marshal(out)
		res := playerevents.Result{}
		_ = json.Unmarshal(raw, &res)
		return res, nil
	}
}

func audienceAllows(audience []string, tags func() (map[string]bool, error)) (bool, error) {
	if len(audience) == 0 {
		return true, nil
	}
	have, err := tags()
	if err != nil {
		return false, err
	}
	for _, id := range audience {
		if have[id] {
			return true, nil
		}
	}
	return false, nil
}

func parseTextArray(raw string) []string {
	raw = strings.Trim(raw, "{}")
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	for i := range parts {
		parts[i] = strings.Trim(parts[i], `"`)
	}
	return parts
}
