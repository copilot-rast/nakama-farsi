package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/heroiclabs/nakama/v3/internal/activity"
)

const maxRelationTargets = 20

func metaRelationResources() []resource {
	return []resource{
		{
			name:  "meta-relations",
			table: "economy_meta_relation",
			json: `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'metaId',t.meta_id,'operator',t.operator,'value',t.value,'targets',t.targets,
				'prizeType',t.prize_type,'prizeId',t.prize_id,'prizeGroupId',t.prize_group_id,'grantMode',t.grant_mode,'repeat',t.repeat,'maxFires',t.max_fires,
				'audienceTagIds',to_jsonb(t.audience_tag_ids),'clientVisible',t.client_visible,'startsAt',t.starts_at,'endsAt',t.ends_at,'isActive',t.is_active,'sort',t.sort,
				'createdAt',t.created_at,'updatedAt',t.updated_at,
				'players',(SELECT count(*) FROM economy_user_meta_relation u WHERE u.relation_id=t.id AND u.fires > 0),
				'claims',(SELECT count(*) FROM economy_reward_claim c WHERE c.kind='relation' AND c.owner_id=t.id))`,
			search: []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{
				"isActive":     "t.is_active = ?::boolean",
				"metaId":       "t.meta_id::text = ?",
				"targetMetaId": "t.targets @> jsonb_build_array(jsonb_build_object('metaId', ?::text))",
				"prizeId":      "t.prize_id::text = ?",
			},
			save: saveMetaRelation,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_reward_claim WHERE kind='relation' AND owner_id=$1)`)
			},
		},
	}
}

func saveMetaRelation(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key         string          `json:"key"`
		Name        i18n            `json:"name"`
		Description i18n            `json:"description"`
		MetaID      string          `json:"metaId"`
		Operator    string          `json:"operator"`
		Value       json.RawMessage `json:"value"`
		Targets     []struct {
			MetaID string          `json:"metaId"`
			Action string          `json:"action"`
			Amount json.RawMessage `json:"amount"`
		} `json:"targets"`
		PrizeType      string   `json:"prizeType"`
		PrizeID        string   `json:"prizeId"`
		PrizeGroupID   string   `json:"prizeGroupId"`
		GrantMode      string   `json:"grantMode"`
		Repeat         string   `json:"repeat"`
		MaxFires       *int     `json:"maxFires"`
		AudienceTagIDs []string `json:"audienceTagIds"`
		ClientVisible  *bool    `json:"clientVisible"`
		StartsAt       string   `json:"startsAt"`
		EndsAt         string   `json:"endsAt"`
		IsActive       bool     `json:"isActive"`
		Sort           int      `json:"sort"`
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
	if err := keepKey(ctx, tx, "economy_meta_relation", id, create, key); err != nil {
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
	if err := oneOf("operator", "INVALID_OPERATOR", in.Operator, "gt", "gte", "lt", "lte", "eq"); err != nil {
		return err
	}
	value, err := optionalNumeric("value", in.Value)
	if err != nil {
		return err
	}
	if value == nil {
		return badRequest("INVALID_VALUE", "value", "a value is required")
	}
	if len(in.Targets) == 0 && in.PrizeType == "none" {
		return badRequest("RELATION_NEEDS_RESULT", "targets", "add a target meta or a prize")
	}
	if len(in.Targets) > maxRelationTargets {
		return badRequest("TOO_MANY_TARGETS", "targets", "at most 20 target metas")
	}
	type target struct {
		MetaID string          `json:"metaId"`
		Action string          `json:"action"`
		Amount json.RawMessage `json:"amount"`
	}
	targets := make([]target, 0, len(in.Targets))
	seen := map[string]bool{}
	for i, t := range in.Targets {
		field := "targets[" + strconv.Itoa(i) + "]"
		ref, err := refExists(ctx, tx, "economy_meta", field+".metaId", "META_NOT_FOUND", t.MetaID, true)
		if err != nil {
			return err
		}
		action := t.Action
		if action == "" {
			action = "add"
		}
		if err := oneOf(field+".action", "INVALID_META_ACTION", action, "add", "set"); err != nil {
			return err
		}
		if seen[*ref+action] {
			return badRequest("DUPLICATE_TARGET", field+".metaId", "each target meta once per action")
		}
		seen[*ref+action] = true
		amount, err := optionalNumeric(field+".amount", t.Amount)
		if err != nil {
			return err
		}
		if amount == nil {
			return badRequest("INVALID_VALUE", field+".amount", "an amount is required")
		}
		if action == "add" && numericFloat(*amount) == 0 {
			return badRequest("ZERO_AMOUNT", field+".amount", "adding zero changes nothing")
		}
		targets = append(targets, target{MetaID: *ref, Action: action, Amount: json.RawMessage(*amount)})
	}
	prizeType := in.PrizeType
	if prizeType == "" {
		prizeType = "none"
	}
	if err := oneOf("prizeType", "INVALID_PRIZE_TYPE", prizeType, "none", "prize", "prizeGroup"); err != nil {
		return err
	}
	var prizeID, groupID *string
	switch prizeType {
	case "prize":
		if prizeID, err = refExists(ctx, tx, "economy_prize", "prizeId", "PRIZE_NOT_FOUND", in.PrizeID, true); err != nil {
			return err
		}
	case "prizeGroup":
		if groupID, err = refExists(ctx, tx, "economy_prize_group", "prizeGroupId", "PRIZE_GROUP_NOT_FOUND", in.PrizeGroupID, true); err != nil {
			return err
		}
	}
	mode := in.GrantMode
	if mode == "" {
		mode = "auto"
	}
	if err := oneOf("grantMode", "INVALID_GRANT_MODE", mode, "auto", "claim"); err != nil {
		return err
	}
	repeat := in.Repeat
	if repeat == "" {
		repeat = "once"
	}
	if err := oneOf("repeat", "INVALID_REPEAT", repeat, "once", "rearm", "every"); err != nil {
		return err
	}
	var maxFires *int
	if repeat != "once" && in.MaxFires != nil {
		if *in.MaxFires < 1 || *in.MaxFires > 1000000 {
			return badRequest("INVALID_MAX_FIRES", "maxFires", "between 1 and 1000000")
		}
		maxFires = in.MaxFires
	}
	audience, err := checkUserTags(ctx, tx, in.AudienceTagIDs)
	if err != nil {
		return err
	}
	starts, ends, err := cleanWindow(in.StartsAt, in.EndsAt)
	if err != nil {
		return err
	}
	visible := true
	if in.ClientVisible != nil {
		visible = *in.ClientVisible
	}
	rawTargets, _ := json.Marshal(targets)
	return execSave(ctx, tx, "economy_meta_relation", id, create,
		[]string{"key", "name", "description", "meta_id", "operator", "value", "targets", "prize_type", "prize_id", "prize_group_id", "grant_mode", "repeat", "max_fires", "audience_tag_ids", "client_visible", "starts_at", "ends_at", "is_active", "sort"},
		[]any{key, i18nJSON(name), i18nJSON(desc), metaID, in.Operator, *value, string(rawTargets), prizeType, prizeID, groupID, mode, repeat, maxFires, audience, visible, starts, ends, in.IsActive, in.Sort})
}

func (s *service) metaRelationConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/users/{userId}/metas", http.MethodPost, s.consoleUserMetaChange},
		{consolePrefix + "/users/{userId}/meta-relations", http.MethodGet, s.consoleUserRelations},
		{consolePrefix + "/users/{userId}/meta-relations/{relationId}", http.MethodDelete, s.consoleUserRelationReset},
	}
}

// consoleUserMetaChange lets an admin add to or set a player's meta; the
// change runs the meta relations and level tables like any other change.
func (s *service) consoleUserMetaChange(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var in struct {
		MetaID string          `json:"metaId"`
		Action string          `json:"action"`
		Value  json.RawMessage `json:"value"`
		Note   string          `json:"note"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if err := oneOf("action", "INVALID_META_ACTION", in.Action, "add", "set"); err != nil {
		s.writeError(w, err)
		return
	}
	amount, err := optionalNumeric("value", in.Value)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if amount == nil {
		s.writeError(w, badRequest("INVALID_VALUE", "value", "a value is required"))
		return
	}
	metaID, err := cleanUUID("metaId", in.MetaID, true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var key string
	var before float64
	if err := s.db.QueryRowContext(ctx, `SELECT m.key, COALESCE(u.value, m.default_value)::float8 FROM economy_meta m LEFT JOIN economy_user_meta u ON u.meta_id=m.id AND u.user_id=$2 WHERE m.id=$1`, *metaID, userID).Scan(&key, &before); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.writeError(w, badRequest("META_NOT_FOUND", "metaId", "the selected record does not exist"))
			return
		}
		s.writeError(w, err)
		return
	}
	var after float64
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		return applyMetaByID(ctx, tx, userID, *metaID, in.Action, *amount, &after)
	}); err != nil {
		s.writeError(w, err)
		return
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindEvent, Action: "meta_changed", RefID: key, Actor: "admin", Data: map[string]any{"meta": key, "action": in.Action, "from": before, "to": after, "note": truncate(in.Note, 300)}})
	out, err := s.evaluateLevels(ctx, userID, []string{*metaID}, 1)
	if err != nil {
		s.logger.Warn("Economy meta checks after admin change for %s failed: %v", userID, err)
	}
	if out == nil {
		out = newOutcome()
	}
	out.Metas[key] = after
	for k := range out.Metas {
		var v float64
		if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(u.value, m.default_value)::float8 FROM economy_meta m LEFT JOIN economy_user_meta u ON u.meta_id=m.id AND u.user_id=$2 WHERE m.key=$1`, k, userID).Scan(&v); err == nil {
			out.Metas[k] = v
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"key": key, "from": before, "value": out.Metas[key], "metas": out.Metas, "relations": out.Relations, "levels": out.Levels, "claims": out.Claims})
}

func applyMetaByID(ctx context.Context, tx *sql.Tx, userID, metaID, action, amount string, after *float64) error {
	q := `INSERT INTO economy_user_meta AS u (user_id, meta_id, value) SELECT $1, id, default_value + $3::numeric FROM economy_meta WHERE id=$2
		ON CONFLICT (user_id, meta_id) DO UPDATE SET value=u.value + $3::numeric, updated_at=now() RETURNING u.value::float8`
	if action == "set" {
		q = `INSERT INTO economy_user_meta AS u (user_id, meta_id, value) SELECT $1, id, $3::numeric FROM economy_meta WHERE id=$2
			ON CONFLICT (user_id, meta_id) DO UPDATE SET value=EXCLUDED.value, updated_at=now() RETURNING u.value::float8`
	}
	err := tx.QueryRowContext(ctx, q, userID, metaID, amount).Scan(after)
	if errors.Is(err, sql.ErrNoRows) {
		return badRequest("META_NOT_FOUND", "metaId", "the selected record does not exist")
	}
	return err
}

func (s *service) consoleUserRelations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	rows, err := s.db.QueryContext(ctx, `SELECT jsonb_build_object('relationId',r.id,'key',r.key,'name',r.name,'metaId',r.meta_id,'operator',r.operator,'value',r.value,'repeat',r.repeat,
			'fires',u.fires,'armed',u.armed,'firstDoneAt',u.first_done_at,'lastDoneAt',u.last_done_at,
			'claims',(SELECT COALESCE(jsonb_agg(jsonb_build_object('claimId',c.id,'status',c.status,'grantMode',c.grant_mode,'prizeId',c.prize_id) ORDER BY c.created_at),'[]'::jsonb)
				FROM economy_reward_claim c WHERE c.kind='relation' AND c.owner_id=r.id AND c.user_id=u.user_id))::text
		FROM economy_user_meta_relation u JOIN economy_meta_relation r ON r.id=u.relation_id WHERE u.user_id=$1 AND u.fires > 0 ORDER BY u.last_done_at DESC LIMIT 100`, userID)
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
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// consoleUserRelationReset forgets that a player completed a relation so it
// can fire again; prizes already received stay.
func (s *service) consoleUserRelationReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	relID, err := cleanUUID("relationId", pathVar(r, "relationId"), true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var key string
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if err := tx.QueryRowContext(ctx, `SELECT key FROM economy_meta_relation WHERE id=$1`, *relID).Scan(&key); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return notFound("NOT_FOUND")
			}
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM economy_reward_claim WHERE kind='relation' AND owner_id=$1 AND user_id=$2 AND status<>'granted' AND grant_mode='claim'`, *relID, userID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `DELETE FROM economy_user_meta_relation WHERE relation_id=$1 AND user_id=$2`, *relID, userID)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindEvent, Action: "meta_relation_reset", RefID: key, Actor: "admin", Data: map[string]any{"relation": key}})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
