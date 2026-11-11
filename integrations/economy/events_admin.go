package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func eventResources() []resource {
	return []resource{
		{
			name:    "event-types",
			table:   "economy_event_type",
			json:    `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'clientAllowed',t.client_allowed,'valueMin',t.value_min,'valueMax',t.value_max,'rateLimit',t.rate_limit,'isActive',t.is_active,'isSystem',t.is_system,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at,'usedBy',(SELECT count(*) FROM economy_event_rule r WHERE r.event_type_id=t.id)+(SELECT count(*) FROM economy_meta_rule r WHERE r.event_type_id=t.id)+(SELECT count(*) FROM economy_streak r WHERE r.event_type_id=t.id))`,
			search:  []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "clientAllowed": "t.client_allowed = ?::boolean"},
			save:    saveEventType,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				var system bool
				if err := tx.QueryRowContext(ctx, `SELECT is_system FROM economy_event_type WHERE id=$1`, id).Scan(&system); err != nil {
					return err
				}
				if system {
					return conflict("SYSTEM_EVENT", "built-in events cannot be deleted; turn them off instead")
				}
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_event_rule WHERE event_type_id=$1) OR EXISTS (SELECT 1 FROM economy_meta_rule WHERE event_type_id=$1)
					OR EXISTS (SELECT 1 FROM economy_streak WHERE event_type_id=$1) OR EXISTS (SELECT 1 FROM economy_player_event_total WHERE type_id=$1)`)
			},
		},
		{
			name:    "event-rules",
			table:   "economy_event_rule",
			json:    `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'eventTypeId',t.event_type_id,'measure',t.measure,'period',t.period,'threshold',t.threshold,'repeat',t.repeat,'maxClaims',t.max_claims,'prizeId',t.prize_id,'grantMode',t.grant_mode,'audienceTagIds',to_jsonb(t.audience_tag_ids),'startsAt',t.starts_at,'endsAt',t.ends_at,'isActive',t.is_active,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at,'claims',(SELECT count(*) FROM economy_reward_claim c WHERE c.kind='event' AND c.owner_id=t.id))`,
			search:  []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "eventTypeId": "t.event_type_id::text = ?", "prizeId": "t.prize_id::text = ?"},
			save:    saveEventRule,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_reward_claim WHERE kind='event' AND owner_id=$1)`)
			},
		},
		{
			name:    "meta-rules",
			table:   "economy_meta_rule",
			json:    `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'eventTypeId',t.event_type_id,'metaId',t.meta_id,'action',t.action,'valueSource',t.value_source,'amount',t.amount,'audienceTagIds',to_jsonb(t.audience_tag_ids),'startsAt',t.starts_at,'endsAt',t.ends_at,'isActive',t.is_active,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at)`,
			search:  []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "eventTypeId": "t.event_type_id::text = ?", "metaId": "t.meta_id::text = ?"},
			save:    saveMetaRule,
			remove:  func(s *service, ctx context.Context, tx *sql.Tx, id string) error { return nil },
		},
		{
			name:    "level-tables",
			table:   "economy_level_table",
			json:    `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'metaId',t.meta_id,'levelMetaId',t.level_meta_id,'isActive',t.is_active,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at,'steps',(SELECT COALESCE(jsonb_agg(jsonb_build_object('id',s.id,'level',s.level,'threshold',s.threshold,'name',s.name,'prizeId',s.prize_id,'grantMode',s.grant_mode) ORDER BY s.level),'[]'::jsonb) FROM economy_level_step s WHERE s.table_id=t.id),'players',(SELECT count(*) FROM economy_user_level u WHERE u.table_id=t.id AND u.level > 0))`,
			search:  []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "metaId": "t.meta_id::text = ?"},
			save:    saveLevelTable,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_reward_claim WHERE kind='level' AND owner_id=$1)`)
			},
		},
		{
			name:    "streaks",
			table:   "economy_streak",
			json:    `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'eventTypeId',t.event_type_id,'periodDays',t.period_days,'gracePeriods',t.grace_periods,'missAction',t.miss_action,'cycleLength',t.cycle_length,'isActive',t.is_active,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at,'rewards',(SELECT COALESCE(jsonb_agg(jsonb_build_object('id',r.id,'day',r.day,'prizeId',r.prize_id,'grantMode',r.grant_mode) ORDER BY r.day),'[]'::jsonb) FROM economy_streak_reward r WHERE r.streak_id=t.id),'players',(SELECT count(*) FROM economy_user_streak u WHERE u.streak_id=t.id AND u.count > 0))`,
			search:  []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "eventTypeId": "t.event_type_id::text = ?"},
			save:    saveStreak,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_reward_claim WHERE kind='streak' AND owner_id=$1)`)
			},
		},
	}
}

func optionalNumeric(field string, raw json.RawMessage) (*string, error) {
	if len(raw) == 0 || string(raw) == "null" || strings.TrimSpace(string(raw)) == `""` {
		return nil, nil
	}
	v, ok := numericFromJSON(raw)
	if !ok {
		return nil, badRequest("INVALID_VALUE", field, "must be a number with at most 4 decimals")
	}
	return &v, nil
}

func numericFloat(v string) float64 {
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func keepKey(ctx context.Context, tx *sql.Tx, table, id string, create bool, key string) error {
	if create {
		return nil
	}
	var old string
	if err := tx.QueryRowContext(ctx, `SELECT key FROM `+table+` WHERE id=$1`, id).Scan(&old); err == nil && old != key {
		return badRequest("KEY_IMMUTABLE", "key", "the key is used by games and cannot change")
	}
	return nil
}

func refExists(ctx context.Context, tx *sql.Tx, table, field, code, id string, required bool) (*string, error) {
	ref, err := cleanUUID(field, id, required)
	if err != nil || ref == nil {
		return ref, err
	}
	var ok bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM `+table+` WHERE id=$1)`, *ref).Scan(&ok); err != nil {
		return nil, err
	}
	if !ok {
		return nil, badRequest(code, field, "the selected record does not exist")
	}
	return ref, nil
}

func cleanWindow(starts, ends string) (*time.Time, *time.Time, error) {
	a, err := cleanTime("startsAt", starts)
	if err != nil {
		return nil, nil, err
	}
	b, err := cleanTime("endsAt", ends)
	if err != nil {
		return nil, nil, err
	}
	if a != nil && b != nil && !a.Before(*b) {
		return nil, nil, badRequest("INVALID_WINDOW", "endsAt", "the end must be after the start")
	}
	return a, b, nil
}

func oneOf(field, code, v string, allowed ...string) error {
	for _, a := range allowed {
		if v == a {
			return nil
		}
	}
	return badRequest(code, field, "must be one of "+strings.Join(allowed, ", "))
}

func saveEventType(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key           string          `json:"key"`
		Name          i18n            `json:"name"`
		Description   i18n            `json:"description"`
		ClientAllowed bool            `json:"clientAllowed"`
		ValueMin      json.RawMessage `json:"valueMin"`
		ValueMax      json.RawMessage `json:"valueMax"`
		RateLimit     int             `json:"rateLimit"`
		IsActive      bool            `json:"isActive"`
		Sort          int             `json:"sort"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := loadLangs(ctx, tx)
	if err != nil {
		return err
	}
	key := strings.TrimSpace(in.Key)
	if !eventKeyPattern.MatchString(key) {
		return badRequest("INVALID_EVENT_KEY", "key", "use 2 to 48 lower-case letters, digits, _ or ., starting with a letter")
	}
	if err := keepKey(ctx, tx, "economy_event_type", id, create, key); err != nil {
		return err
	}
	if !create {
		var system bool
		if err := tx.QueryRowContext(ctx, `SELECT is_system FROM economy_event_type WHERE id=$1`, id).Scan(&system); err == nil && system && in.ClientAllowed {
			return badRequest("SYSTEM_EVENT_SERVER_ONLY", "clientAllowed", "built-in events are recorded only by the server")
		}
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 128)
	if err != nil {
		return err
	}
	desc, err := cleanI18n(langs, "description", in.Description, false, 1000)
	if err != nil {
		return err
	}
	min, err := optionalNumeric("valueMin", in.ValueMin)
	if err != nil {
		return err
	}
	max, err := optionalNumeric("valueMax", in.ValueMax)
	if err != nil {
		return err
	}
	if min != nil && max != nil && numericFloat(*min) > numericFloat(*max) {
		return badRequest("INVALID_RANGE", "valueMax", "the largest value must not be smaller than the smallest")
	}
	if in.RateLimit == 0 {
		in.RateLimit = 30
	}
	if in.RateLimit < 1 || in.RateLimit > 100000 {
		return badRequest("INVALID_RATE_LIMIT", "rateLimit", "between 1 and 100000 per minute")
	}
	return execSave(ctx, tx, "economy_event_type", id, create, []string{"key", "name", "description", "client_allowed", "value_min", "value_max", "rate_limit", "is_active", "sort"},
		[]any{key, i18nJSON(name), i18nJSON(desc), in.ClientAllowed, min, max, in.RateLimit, in.IsActive, in.Sort})
}

func saveEventRule(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key            string          `json:"key"`
		Name           i18n            `json:"name"`
		Description    i18n            `json:"description"`
		EventTypeID    string          `json:"eventTypeId"`
		Measure        string          `json:"measure"`
		Period         string          `json:"period"`
		Threshold      json.RawMessage `json:"threshold"`
		Repeat         string          `json:"repeat"`
		MaxClaims      *int            `json:"maxClaims"`
		PrizeID        string          `json:"prizeId"`
		GrantMode      string          `json:"grantMode"`
		AudienceTagIDs []string        `json:"audienceTagIds"`
		StartsAt       string          `json:"startsAt"`
		EndsAt         string          `json:"endsAt"`
		IsActive       bool            `json:"isActive"`
		Sort           int             `json:"sort"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := loadLangs(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	if err := keepKey(ctx, tx, "economy_event_rule", id, create, key); err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 128)
	if err != nil {
		return err
	}
	desc, err := cleanI18n(langs, "description", in.Description, false, 1000)
	if err != nil {
		return err
	}
	typeID, err := refExists(ctx, tx, "economy_event_type", "eventTypeId", "EVENT_TYPE_NOT_FOUND", in.EventTypeID, true)
	if err != nil {
		return err
	}
	if err := oneOf("measure", "INVALID_MEASURE", in.Measure, "count", "sum", "max"); err != nil {
		return err
	}
	if err := oneOf("period", "INVALID_PERIOD", in.Period, "total", "day", "week"); err != nil {
		return err
	}
	if err := oneOf("repeat", "INVALID_REPEAT", in.Repeat, "once", "period", "every"); err != nil {
		return err
	}
	if in.Repeat == "period" && in.Period == "total" {
		return badRequest("REPEAT_NEEDS_PERIOD", "repeat", "once per day or week needs a day or week window")
	}
	if in.Repeat == "every" && in.Measure == "max" {
		return badRequest("EVERY_NEEDS_TOTAL", "repeat", "every N works with a count or a sum, not with the highest value")
	}
	threshold, err := optionalNumeric("threshold", in.Threshold)
	if err != nil {
		return err
	}
	if threshold == nil || numericFloat(*threshold) <= 0 {
		return badRequest("INVALID_THRESHOLD", "threshold", "the target must be greater than zero")
	}
	if in.MaxClaims != nil && *in.MaxClaims < 1 {
		return badRequest("INVALID_MAX_CLAIMS", "maxClaims", "at least 1, or empty for no limit")
	}
	prizeID, err := refExists(ctx, tx, "economy_prize", "prizeId", "PRIZE_NOT_FOUND", in.PrizeID, true)
	if err != nil {
		return err
	}
	if err := oneOf("grantMode", "INVALID_GRANT_MODE", in.GrantMode, "auto", "claim"); err != nil {
		return err
	}
	audience, err := checkUserTags(ctx, tx, in.AudienceTagIDs)
	if err != nil {
		return err
	}
	starts, ends, err := cleanWindow(in.StartsAt, in.EndsAt)
	if err != nil {
		return err
	}
	return execSave(ctx, tx, "economy_event_rule", id, create, []string{"key", "name", "description", "event_type_id", "measure", "period", "threshold", "repeat", "max_claims", "prize_id", "grant_mode", "audience_tag_ids", "starts_at", "ends_at", "is_active", "sort"},
		[]any{key, i18nJSON(name), i18nJSON(desc), typeID, in.Measure, in.Period, *threshold, in.Repeat, in.MaxClaims, prizeID, in.GrantMode, audience, starts, ends, in.IsActive, in.Sort})
}

func saveMetaRule(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key            string          `json:"key"`
		Name           i18n            `json:"name"`
		EventTypeID    string          `json:"eventTypeId"`
		MetaID         string          `json:"metaId"`
		Action         string          `json:"action"`
		ValueSource    string          `json:"valueSource"`
		Amount         json.RawMessage `json:"amount"`
		AudienceTagIDs []string        `json:"audienceTagIds"`
		StartsAt       string          `json:"startsAt"`
		EndsAt         string          `json:"endsAt"`
		IsActive       bool            `json:"isActive"`
		Sort           int             `json:"sort"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := loadLangs(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 128)
	if err != nil {
		return err
	}
	typeID, err := refExists(ctx, tx, "economy_event_type", "eventTypeId", "EVENT_TYPE_NOT_FOUND", in.EventTypeID, true)
	if err != nil {
		return err
	}
	metaID, err := refExists(ctx, tx, "economy_meta", "metaId", "META_NOT_FOUND", in.MetaID, true)
	if err != nil {
		return err
	}
	if err := oneOf("action", "INVALID_META_ACTION", in.Action, "add", "set", "max"); err != nil {
		return err
	}
	if err := oneOf("valueSource", "INVALID_VALUE_SOURCE", in.ValueSource, "fixed", "event"); err != nil {
		return err
	}
	amount, err := optionalNumeric("amount", in.Amount)
	if err != nil {
		return err
	}
	if amount == nil {
		return badRequest("INVALID_VALUE", "amount", "an amount is required")
	}
	audience, err := checkUserTags(ctx, tx, in.AudienceTagIDs)
	if err != nil {
		return err
	}
	starts, ends, err := cleanWindow(in.StartsAt, in.EndsAt)
	if err != nil {
		return err
	}
	return execSave(ctx, tx, "economy_meta_rule", id, create, []string{"key", "name", "event_type_id", "meta_id", "action", "value_source", "amount", "audience_tag_ids", "starts_at", "ends_at", "is_active", "sort"},
		[]any{key, i18nJSON(name), typeID, metaID, in.Action, in.ValueSource, *amount, audience, starts, ends, in.IsActive, in.Sort})
}

func saveLevelTable(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key         string `json:"key"`
		Name        i18n   `json:"name"`
		Description i18n   `json:"description"`
		MetaID      string `json:"metaId"`
		LevelMetaID string `json:"levelMetaId"`
		IsActive    bool   `json:"isActive"`
		Sort        int    `json:"sort"`
		Steps       []struct {
			Level     int             `json:"level"`
			Threshold json.RawMessage `json:"threshold"`
			Name      i18n            `json:"name"`
			PrizeID   string          `json:"prizeId"`
			GrantMode string          `json:"grantMode"`
		} `json:"steps"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := loadLangs(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	if err := keepKey(ctx, tx, "economy_level_table", id, create, key); err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 128)
	if err != nil {
		return err
	}
	desc, err := cleanI18n(langs, "description", in.Description, false, 1000)
	if err != nil {
		return err
	}
	metaID, err := refExists(ctx, tx, "economy_meta", "metaId", "META_NOT_FOUND", in.MetaID, true)
	if err != nil {
		return err
	}
	levelMetaID, err := refExists(ctx, tx, "economy_meta", "levelMetaId", "META_NOT_FOUND", in.LevelMetaID, false)
	if err != nil {
		return err
	}
	if levelMetaID != nil && *levelMetaID == *metaID {
		return badRequest("SAME_META", "levelMetaId", "the level must be written to a different meta")
	}
	if len(in.Steps) == 0 {
		return badRequest("STEPS_REQUIRED", "steps", "add at least one level")
	}
	if len(in.Steps) > 200 {
		return badRequest("TOO_MANY_STEPS", "steps", "at most 200 levels")
	}
	type step struct {
		level           int
		threshold, mode string
		name            i18n
		prize           *string
	}
	steps := make([]step, 0, len(in.Steps))
	seen := map[int]bool{}
	for i, st := range in.Steps {
		field := "steps[" + strconv.Itoa(i) + "]"
		if st.Level < 1 {
			return badRequest("INVALID_LEVEL", field+".level", "levels start at 1")
		}
		if seen[st.Level] {
			return badRequest("DUPLICATE_LEVEL", field+".level", "each level once")
		}
		seen[st.Level] = true
		th, err := optionalNumeric(field+".threshold", st.Threshold)
		if err != nil {
			return err
		}
		if th == nil || numericFloat(*th) < 0 {
			return badRequest("INVALID_THRESHOLD", field+".threshold", "the target cannot be negative")
		}
		nm, err := cleanI18n(langs, field+".name", st.Name, false, 64)
		if err != nil {
			return err
		}
		prize, err := refExists(ctx, tx, "economy_prize", field+".prizeId", "PRIZE_NOT_FOUND", st.PrizeID, false)
		if err != nil {
			return err
		}
		mode := st.GrantMode
		if mode == "" {
			mode = "claim"
		}
		if err := oneOf(field+".grantMode", "INVALID_GRANT_MODE", mode, "auto", "claim"); err != nil {
			return err
		}
		steps = append(steps, step{level: st.Level, threshold: *th, mode: mode, name: nm, prize: prize})
	}
	sort.Slice(steps, func(i, j int) bool { return steps[i].level < steps[j].level })
	for i := 1; i < len(steps); i++ {
		if numericFloat(steps[i].threshold) < numericFloat(steps[i-1].threshold) {
			return badRequest("THRESHOLDS_NOT_ORDERED", "steps", "a higher level needs the same or a larger target")
		}
	}
	if err := execSave(ctx, tx, "economy_level_table", id, create, []string{"key", "name", "description", "meta_id", "level_meta_id", "is_active", "sort"},
		[]any{key, i18nJSON(name), i18nJSON(desc), metaID, levelMetaID, in.IsActive, in.Sort}); err != nil {
		return err
	}
	levels := []int{}
	for _, st := range steps {
		levels = append(levels, st.level)
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_level_step (id, table_id, level, threshold, name, prize_id, grant_mode) VALUES ($1,$2,$3,$4::numeric,$5::jsonb,$6,$7)
			ON CONFLICT (table_id, level) DO UPDATE SET threshold=EXCLUDED.threshold, name=EXCLUDED.name, prize_id=EXCLUDED.prize_id, grant_mode=EXCLUDED.grant_mode`,
			uuid.Must(uuid.NewV4()).String(), id, st.level, st.threshold, i18nJSON(st.name), st.prize, st.mode); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM economy_level_step WHERE table_id=$1 AND NOT (level = ANY($2))`, id, levels)
	return err
}

func saveStreak(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key          string `json:"key"`
		Name         i18n   `json:"name"`
		Description  i18n   `json:"description"`
		EventTypeID  string `json:"eventTypeId"`
		PeriodDays   int    `json:"periodDays"`
		GracePeriods int    `json:"gracePeriods"`
		MissAction   string `json:"missAction"`
		CycleLength  *int   `json:"cycleLength"`
		IsActive     bool   `json:"isActive"`
		Sort         int    `json:"sort"`
		Rewards      []struct {
			Day       int    `json:"day"`
			PrizeID   string `json:"prizeId"`
			GrantMode string `json:"grantMode"`
		} `json:"rewards"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := loadLangs(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	if err := keepKey(ctx, tx, "economy_streak", id, create, key); err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 128)
	if err != nil {
		return err
	}
	desc, err := cleanI18n(langs, "description", in.Description, false, 1000)
	if err != nil {
		return err
	}
	typeID, err := refExists(ctx, tx, "economy_event_type", "eventTypeId", "EVENT_TYPE_NOT_FOUND", in.EventTypeID, true)
	if err != nil {
		return err
	}
	if in.PeriodDays == 0 {
		in.PeriodDays = 1
	}
	if in.PeriodDays < 1 || in.PeriodDays > 30 {
		return badRequest("INVALID_PERIOD_DAYS", "periodDays", "between 1 and 30 days")
	}
	if in.GracePeriods < 0 || in.GracePeriods > 30 {
		return badRequest("INVALID_GRACE", "gracePeriods", "between 0 and 30")
	}
	if in.MissAction == "" {
		in.MissAction = "reset"
	}
	if err := oneOf("missAction", "INVALID_MISS_ACTION", in.MissAction, "reset", "step_back"); err != nil {
		return err
	}
	if in.CycleLength != nil && (*in.CycleLength < 1 || *in.CycleLength > 365) {
		return badRequest("INVALID_CYCLE", "cycleLength", "between 1 and 365 days, or empty")
	}
	if len(in.Rewards) > 365 {
		return badRequest("TOO_MANY_REWARDS", "rewards", "at most 365 reward days")
	}
	seen := map[int]bool{}
	days := []int{}
	type reward struct {
		day   int
		prize *string
		mode  string
	}
	var rewards []reward
	for i, r := range in.Rewards {
		field := "rewards[" + strconv.Itoa(i) + "]"
		if r.Day < 1 || r.Day > 3650 {
			return badRequest("INVALID_DAY", field+".day", "days start at 1")
		}
		if in.CycleLength != nil && r.Day > *in.CycleLength {
			return badRequest("DAY_OUTSIDE_CYCLE", field+".day", "the day is after the end of the cycle")
		}
		if seen[r.Day] {
			return badRequest("DUPLICATE_DAY", field+".day", "each day once")
		}
		seen[r.Day] = true
		prize, err := refExists(ctx, tx, "economy_prize", field+".prizeId", "PRIZE_NOT_FOUND", r.PrizeID, true)
		if err != nil {
			return err
		}
		mode := r.GrantMode
		if mode == "" {
			mode = "claim"
		}
		if err := oneOf(field+".grantMode", "INVALID_GRANT_MODE", mode, "auto", "claim"); err != nil {
			return err
		}
		days = append(days, r.Day)
		rewards = append(rewards, reward{day: r.Day, prize: prize, mode: mode})
	}
	if err := execSave(ctx, tx, "economy_streak", id, create, []string{"key", "name", "description", "event_type_id", "period_days", "grace_periods", "miss_action", "cycle_length", "is_active", "sort"},
		[]any{key, i18nJSON(name), i18nJSON(desc), typeID, in.PeriodDays, in.GracePeriods, in.MissAction, in.CycleLength, in.IsActive, in.Sort}); err != nil {
		return err
	}
	for _, r := range rewards {
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_streak_reward (id, streak_id, day, prize_id, grant_mode) VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (streak_id, day) DO UPDATE SET prize_id=EXCLUDED.prize_id, grant_mode=EXCLUDED.grant_mode`, uuid.Must(uuid.NewV4()).String(), id, r.day, r.prize, r.mode); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM economy_streak_reward WHERE streak_id=$1 AND NOT (day = ANY($2))`, id, days)
	return err
}

func (s *service) eventConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/event-settings", http.MethodGet, s.consoleEventSettingsGet},
		{consolePrefix + "/event-settings", http.MethodPut, s.consoleEventSettingsPut},
		{consolePrefix + "/event-log", http.MethodGet, s.consoleEventLog},
		{consolePrefix + "/event-claims", http.MethodGet, s.consoleEventClaims},
		{consolePrefix + "/event-claims/{id}/retry", http.MethodPost, s.consoleEventClaimRetry},
		{consolePrefix + "/users/{userId}/progress", http.MethodGet, s.consoleUserProgress},
		{consolePrefix + "/users/{userId}/events", http.MethodPost, s.consoleUserEvent},
		{consolePrefix + "/users/{userId}/streaks/{streakId}/reset", http.MethodPost, s.consoleStreakReset},
	}
}

func (s *service) consoleEventSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := loadEventSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *service) consoleEventSettingsPut(w http.ResponseWriter, r *http.Request) {
	var in eventSettings
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	in.Timezone = strings.TrimSpace(in.Timezone)
	if _, err := time.LoadLocation(in.Timezone); err != nil || in.Timezone == "" || strings.EqualFold(in.Timezone, "local") {
		s.writeError(w, badRequest("INVALID_TIMEZONE", "timezone", "use a time zone name such as Asia/Tehran or UTC"))
		return
	}
	if in.WeekStart < 0 || in.WeekStart > 6 {
		s.writeError(w, badRequest("INVALID_WEEK_START", "weekStart", "0 (Sunday) to 6 (Saturday)"))
		return
	}
	if in.RetentionDays < 1 || in.RetentionDays > 3650 {
		s.writeError(w, badRequest("INVALID_RETENTION", "retentionDays", "between 1 and 3650 days"))
		return
	}
	if in.PlayerRateLimit < 1 || in.PlayerRateLimit > 100000 {
		s.writeError(w, badRequest("INVALID_RATE_LIMIT", "playerRateLimit", "between 1 and 100000 per minute"))
		return
	}
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO economy_event_settings (id, timezone, week_start, retention_days, player_rate_limit, updated_at) VALUES (1,$1,$2,$3,$4,now())
		ON CONFLICT (id) DO UPDATE SET timezone=EXCLUDED.timezone, week_start=EXCLUDED.week_start, retention_days=EXCLUDED.retention_days, player_rate_limit=EXCLUDED.player_rate_limit, updated_at=now()`,
		in.Timezone, in.WeekStart, in.RetentionDays, in.PlayerRateLimit); err != nil {
		s.writeError(w, err)
		return
	}
	s.consoleEventSettingsGet(w, r)
}

func uuidParam(r *http.Request, name string) (string, error) {
	v := strings.TrimSpace(r.URL.Query().Get(name))
	if v == "" {
		return "", nil
	}
	if _, err := uuid.FromString(v); err != nil {
		return "", badRequest("INVALID_ID", name, "must be a UUID")
	}
	return v, nil
}

func (s *service) listRaw(w http.ResponseWriter, r *http.Request, countSQL, selectSQL string, where *sqlWhere, order string, p pageQuery) {
	var total int64
	if err := s.db.QueryRowContext(r.Context(), countSQL+where.sql(), where.args...).Scan(&total); err != nil {
		s.writeError(w, err)
		return
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(r.Context(), selectSQL+where.sql()+order+` LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2), append(append([]any{}, where.args...), p.Limit, p.Offset)...)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	items := []json.RawMessage{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			s.writeError(w, err)
			return
		}
		items = append(items, json.RawMessage(raw))
	}
	if err := rows.Err(); err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": p.Page, "limit": p.Limit})
}

func (s *service) consoleEventLog(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	q := r.URL.Query()
	where := &sqlWhere{}
	for _, name := range []string{"userId", "typeId"} {
		v, err := uuidParam(r, name)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if v != "" {
			if name == "userId" {
				where.add("e.user_id = ?::uuid", v)
			} else {
				where.add("e.type_id = ?::uuid", v)
			}
		}
	}
	if v := q.Get("source"); v != "" {
		if err := oneOf("source", "INVALID_FILTER", v, "client", "server", "admin"); err != nil {
			s.writeError(w, err)
			return
		}
		where.add("e.source = ?", v)
	}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(COALESCE(u.username,'') ILIKE "+n+" OR et.key ILIKE "+n+")")
	}
	from := ` FROM economy_player_event e JOIN economy_event_type et ON et.id=e.type_id LEFT JOIN users u ON u.id=e.user_id`
	s.listRaw(w, r, `SELECT count(*)`+from, `SELECT jsonb_build_object('id',e.id,'userId',e.user_id,'username',u.username,'displayName',u.display_name,'typeId',e.type_id,'key',et.key,'name',et.name,'value',e.value,'metadata',e.metadata,'source',e.source,'createdAt',e.created_at)::text`+from,
		where, ` ORDER BY e.created_at DESC, e.id DESC`, p)
}

func (s *service) consoleEventClaims(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	q := r.URL.Query()
	where := &sqlWhere{}
	for _, name := range []string{"userId", "ownerId"} {
		v, err := uuidParam(r, name)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if v != "" {
			where.add("c."+map[string]string{"userId": "user_id", "ownerId": "owner_id"}[name]+" = ?::uuid", v)
		}
	}
	if v := q.Get("kind"); v != "" {
		where.add("c.kind = ?", v)
	}
	if v := q.Get("status"); v != "" {
		where.add("c.status = ?", v)
	}
	from := ` FROM economy_reward_claim c JOIN economy_prize z ON z.id=c.prize_id LEFT JOIN users u ON u.id=c.user_id
		LEFT JOIN economy_event_rule er ON c.kind='event' AND er.id=c.owner_id LEFT JOIN economy_level_table lt ON c.kind='level' AND lt.id=c.owner_id
		LEFT JOIN economy_level_step ls ON c.kind='level' AND ls.id=c.ref_id LEFT JOIN economy_streak st ON c.kind='streak' AND st.id=c.owner_id`
	s.listRaw(w, r, `SELECT count(*)`+from, `SELECT jsonb_build_object('id',c.id,'userId',c.user_id,'username',u.username,'kind',c.kind,'ownerId',c.owner_id,'sourceKey',COALESCE(er.key,lt.key,st.key),
		'sourceName',COALESCE(er.name,lt.name,st.name),'level',ls.level,'scope',c.scope,'prizeId',c.prize_id,'prizeKey',z.key,'prizeName',z.name,'grantMode',c.grant_mode,'status',c.status,'value',c.value,
		'error',c.error,'attempts',c.attempts,'createdAt',c.created_at,'grantedAt',c.granted_at)::text`+from, where, ` ORDER BY c.created_at DESC, c.id DESC`, p)
}

func (s *service) consoleEventClaimRetry(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("CLAIM_NOT_FOUND"))
		return
	}
	var status string
	if err := s.db.QueryRowContext(r.Context(), `SELECT status FROM economy_reward_claim WHERE id=$1`, id).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.writeError(w, notFound("CLAIM_NOT_FOUND"))
			return
		}
		s.writeError(w, err)
		return
	}
	if status == "granted" {
		s.writeError(w, conflict("ALREADY_CLAIMED", "this reward was already received"))
		return
	}
	if _, err := s.grantClaim(r.Context(), id, 1); err != nil {
		s.writeError(w, err)
		return
	}
	rewards, err := s.claimRewards(r.Context(), id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rewards": rewards})
}

func (s *service) consoleUser(r *http.Request) (string, error) {
	userID := pathVar(r, "userId")
	if _, err := uuid.FromString(userID); err != nil {
		return "", notFound("USER_NOT_FOUND")
	}
	var ok bool
	if err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM users WHERE id=$1)`, userID).Scan(&ok); err != nil {
		return "", err
	}
	if !ok {
		return "", notFound("USER_NOT_FOUND")
	}
	return userID, nil
}

func (s *service) consoleUserProgress(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	lang := langs.Choose(r.URL.Query().Get("lang"))
	rows, err := s.db.QueryContext(ctx, `SELECT jsonb_build_object('typeId',t.type_id,'key',et.key,'name',et.name,'count',t.count,'sum',t.sum,'max',t.max,'firstAt',t.first_at,'lastAt',t.last_at)::text
		FROM economy_player_event_total t JOIN economy_event_type et ON et.id=t.type_id WHERE t.user_id=$1 ORDER BY t.last_at DESC LIMIT 50`, userID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	totals := []json.RawMessage{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			rows.Close()
			s.writeError(w, err)
			return
		}
		totals = append(totals, json.RawMessage(raw))
	}
	rows.Close()
	streaks, err := s.streakViews(ctx, userID, langs, lang, false)
	if err != nil {
		s.writeError(w, err)
		return
	}
	levels, err := s.levelViews(ctx, userID, langs, lang, false)
	if err != nil {
		s.writeError(w, err)
		return
	}
	claims, err := s.claimViews(ctx, userID, "", 20, langs, lang)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var ready int64
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM economy_reward_claim WHERE user_id=$1 AND status<>'granted'`, userID).Scan(&ready); err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"lang": lang, "totals": totals, "streaks": streaks, "levels": levels, "claims": claims, "readyClaims": ready})
}

func (s *service) consoleUserEvent(w http.ResponseWriter, r *http.Request) {
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var in struct {
		Key      string         `json:"key"`
		Value    *float64       `json:"value"`
		Metadata map[string]any `json:"metadata"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	out, err := s.recordEvent(r.Context(), eventInput{UserID: userID, Key: in.Key, Value: in.Value, Metadata: in.Metadata, Source: "admin"})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) consoleStreakReset(w http.ResponseWriter, r *http.Request) {
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	streakID := pathVar(r, "streakId")
	if _, err := uuid.FromString(streakID); err != nil {
		s.writeError(w, notFound("STREAK_NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `UPDATE economy_user_streak SET run=run+1, count=0, grace_used=0, last_period=NULL, started_at=now(), updated_at=now() WHERE streak_id=$1 AND user_id=$2`, streakID, userID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.writeError(w, notFound("STREAK_NOT_FOUND"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}
