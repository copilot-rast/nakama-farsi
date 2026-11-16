package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
)

const chanceGameJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'metadata',t.metadata,'prizeType',t.prize_type,'prize',t.prize,
	'intervalTime',t.interval_time,'price',t.price,'isActive',t.is_active,'sort',t.sort,'kind',t.kind,'dailyLimit',t.daily_limit,'totalLimit',t.total_limit,'showOdds',t.show_odds,'createdAt',t.created_at,'updatedAt',t.updated_at,
	'plays',(SELECT count(*) FROM economy_chance_play p WHERE p.game_id=t.id),
	'players',(SELECT count(*) FROM economy_chance_state c WHERE c.game_id=t.id),
	'wins',(SELECT count(*) FROM economy_chance_play p WHERE p.game_id=t.id AND p.prize_id IS NOT NULL),
	'lastPlayAt',(SELECT max(p.created_at) FROM economy_chance_play p WHERE p.game_id=t.id))`

const prizeGroupJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at,
	'members',(SELECT COALESCE(jsonb_agg(jsonb_build_object('prizeId',m.prize_id,'weight',m.weight,'sort',m.sort) ORDER BY m.sort, m.prize_id),'[]'::jsonb) FROM economy_prize_group_member m WHERE m.group_id=t.id),
	'games',(SELECT count(*) FROM economy_chance_game g WHERE g.prize @> jsonb_build_array(jsonb_build_object('prizeGroupId',t.id::text))))`

func chanceResources() []resource {
	return []resource{
		{
			name:   "chance-games",
			table:  "economy_chance_game",
			json:   chanceGameJSON,
			search: []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "prizeType": "t.prize_type = ?", "kind": "t.kind = ?",
				"mode": "(CASE ? WHEN 'free' THEN t.interval_time IS NOT NULL WHEN 'ticket' THEN t.price IS NOT NULL ELSE true END)"},
			save: saveChanceGame,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_chance_play WHERE game_id=$1)`)
			},
		},
		{
			name:    "prize-groups",
			table:   "economy_prize_group",
			json:    prizeGroupJSON,
			search:  []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{"prizeId": "EXISTS (SELECT 1 FROM economy_prize_group_member m WHERE m.group_id=t.id AND m.prize_id::text = ?)"},
			save:    savePrizeGroup,
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_chance_game g WHERE g.prize @> jsonb_build_array(jsonb_build_object('prizeGroupId',$1::text)))
					OR EXISTS (SELECT 1 FROM economy_chance_play WHERE prize_group_id::text=$1::text) OR EXISTS (SELECT 1 FROM economy_meta_relation WHERE prize_group_id::text=$1::text)`)
			},
		},
	}
}

func chanceTotalOf(rows []chanceRow) float64 {
	sum := 0
	for _, r := range rows {
		sum += scaledChance(r.Chance)
	}
	return float64(sum) / chanceScale
}

func saveChanceGame(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key         string          `json:"key"`
		Name        i18n            `json:"name"`
		Description i18n            `json:"description"`
		Metadata    json.RawMessage `json:"metadata"`
		PrizeType   string          `json:"prizeType"`
		Prize       []struct {
			Chance       json.RawMessage `json:"chance"`
			PrizeID      string          `json:"prizeId"`
			PrizeGroupID string          `json:"prizeGroupId"`
		} `json:"prize"`
		IntervalTime json.RawMessage `json:"intervalTime"`
		Price        *struct {
			CurrencyID string          `json:"currencyId"`
			Amount     json.RawMessage `json:"amount"`
		} `json:"price"`
		IsActive   bool   `json:"isActive"`
		Sort       int    `json:"sort"`
		Kind       string `json:"kind"`
		DailyLimit *int   `json:"dailyLimit"`
		TotalLimit *int   `json:"totalLimit"`
		ShowOdds   *bool  `json:"showOdds"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	if in.Kind == "" {
		in.Kind = "wheel"
	}
	if err := oneOf("kind", "INVALID_KIND", in.Kind, "wheel", "chest"); err != nil {
		return err
	}
	if in.DailyLimit != nil && (*in.DailyLimit < 1 || *in.DailyLimit > 100000) {
		return badRequest("INVALID_LIMIT", "dailyLimit", "from 1 to 100000, or empty for no limit")
	}
	if in.TotalLimit != nil && (*in.TotalLimit < 1 || *in.TotalLimit > 10000000) {
		return badRequest("INVALID_LIMIT", "totalLimit", "from 1 to 10000000, or empty for no limit")
	}
	showOdds := in.ShowOdds == nil || *in.ShowOdds
	langs, err := loadLangs(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanKey("key", in.Key)
	if err != nil {
		return err
	}
	if err := keepKey(ctx, tx, "economy_chance_game", id, create, key); err != nil {
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
	meta, err := cleanObject("metadata", in.Metadata)
	if err != nil {
		return err
	}
	if in.PrizeType == "" {
		in.PrizeType = "prize"
	}
	if err := oneOf("prizeType", "INVALID_PRIZE_TYPE", in.PrizeType, "prize", "prizeGroup"); err != nil {
		return err
	}
	if len(in.Prize) == 0 {
		return badRequest("PRIZE_LIST_REQUIRED", "prize", "add at least one row with its chance")
	}
	if len(in.Prize) > maxChanceRows {
		return badRequest("TOO_MANY_PRIZES", "prize", "at most 50 rows")
	}
	rows := make([]chanceRow, 0, len(in.Prize))
	for i, r := range in.Prize {
		field := "prize[" + strconv.Itoa(i) + "]"
		var chance float64
		if err := json.Unmarshal(r.Chance, &chance); err != nil || math.IsNaN(chance) || chance < 0 || chance > 100 {
			return badRequest("INVALID_PRIZE_CHANCE", field, "chance must be a number from 0 to 100")
		}
		chance = math.Round(chance*chanceScale) / chanceScale
		row := chanceRow{Chance: chance}
		prizeID, groupID := strings.TrimSpace(r.PrizeID), strings.TrimSpace(r.PrizeGroupID)
		if in.PrizeType == "prize" && groupID != "" || in.PrizeType == "prizeGroup" && prizeID != "" {
			return badRequest("PRIZE_TYPE_MISMATCH", field, "use prizeId rows for prizeType prize and prizeGroupId rows for prizeType prizeGroup")
		}
		if prizeID != "" {
			ref, err := refExists(ctx, tx, "economy_prize", field, "PRIZE_NOT_FOUND", prizeID, false)
			if err != nil {
				return err
			}
			row.PrizeID = ref
		}
		if groupID != "" {
			ref, err := refExists(ctx, tx, "economy_prize_group", field, "PRIZE_GROUP_NOT_FOUND", groupID, false)
			if err != nil {
				return err
			}
			row.PrizeGroupID = ref
		}
		rows = append(rows, row)
	}
	if total := chanceTotalOf(rows); scaledChance(total) != chanceTotal {
		e := badRequest("TOTAL_CHANCE_MUST_BE_100", "prize", "the chances must add up to exactly 100")
		e.Meta = map[string]any{"totalChance": total}
		return e
	}
	var interval *int
	if len(in.IntervalTime) > 0 && string(in.IntervalTime) != "null" {
		var v float64
		if err := json.Unmarshal(in.IntervalTime, &v); err != nil || v != math.Trunc(v) || v < 1 || v > maxChanceInterval {
			return badRequest("INVALID_INTERVAL_TIME", "intervalTime", "whole minutes from 1 to 525600")
		}
		n := int(v)
		interval = &n
	}
	var price any
	if in.Price != nil && (strings.TrimSpace(in.Price.CurrencyID) != "" || (len(in.Price.Amount) > 0 && string(in.Price.Amount) != "null")) {
		if strings.TrimSpace(in.Price.CurrencyID) == "" || len(in.Price.Amount) == 0 || string(in.Price.Amount) == "null" {
			return badRequest("INVALID_PRICE_PAYLOAD", "price", "give both the currency and the amount, or neither")
		}
		cur, err := refExists(ctx, tx, "economy_currency", "price.currencyId", "CURRENCY_NOT_FOUND", in.Price.CurrencyID, true)
		if err != nil {
			return err
		}
		var amount float64
		if err := json.Unmarshal(in.Price.Amount, &amount); err != nil || amount != math.Trunc(amount) || amount < 1 || amount > 1e12 {
			return badRequest("INVALID_PRICE_AMOUNT", "price.amount", "a whole amount of 1 or more")
		}
		raw, _ := json.Marshal(chancePrice{CurrencyID: *cur, Amount: int64(amount)})
		price = string(raw)
	}
	if interval == nil && price == nil {
		return badRequest("INTERVAL_OR_PRICE_REQUIRED", "intervalTime", "set a free turn interval, a ticket price, or both")
	}
	rowsJSON, _ := json.Marshal(rows)
	return execSave(ctx, tx, "economy_chance_game", id, create,
		[]string{"key", "name", "description", "metadata", "prize_type", "prize", "interval_time", "price", "is_active", "sort", "kind", "daily_limit", "total_limit", "show_odds"},
		[]any{key, i18nJSON(name), i18nJSON(desc), meta, in.PrizeType, string(rowsJSON), interval, price, in.IsActive, in.Sort, in.Kind, in.DailyLimit, in.TotalLimit, showOdds})
}

func savePrizeGroup(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key         string `json:"key"`
		Name        i18n   `json:"name"`
		Description i18n   `json:"description"`
		Sort        int    `json:"sort"`
		Members     []struct {
			PrizeID string          `json:"prizeId"`
			Weight  json.RawMessage `json:"weight"`
		} `json:"members"`
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
	if err := keepKey(ctx, tx, "economy_prize_group", id, create, key); err != nil {
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
	if len(in.Members) == 0 {
		return badRequest("PRIZE_GROUP_EMPTY", "members", "add at least one prize")
	}
	if len(in.Members) > maxGroupMembers {
		return badRequest("TOO_MANY_PRIZES", "members", "at most 100 prizes")
	}
	type member struct {
		prize  string
		weight int
	}
	members := make([]member, 0, len(in.Members))
	seen := map[string]bool{}
	for i, m := range in.Members {
		field := "members[" + strconv.Itoa(i) + "]"
		ref, err := refExists(ctx, tx, "economy_prize", field, "PRIZE_NOT_FOUND", m.PrizeID, true)
		if err != nil {
			return err
		}
		if seen[*ref] {
			return badRequest("DUPLICATE_PRIZE", field, "each prize can be in the group once")
		}
		seen[*ref] = true
		weight := 1.0
		if len(m.Weight) > 0 && string(m.Weight) != "null" {
			if err := json.Unmarshal(m.Weight, &weight); err != nil || weight != math.Trunc(weight) || weight < 1 || weight > 1000000 {
				return badRequest("INVALID_WEIGHT", field, "a whole weight from 1 to 1000000")
			}
		}
		members = append(members, member{prize: *ref, weight: int(weight)})
	}
	if err := execSave(ctx, tx, "economy_prize_group", id, create, []string{"key", "name", "description", "sort"}, []any{key, i18nJSON(name), i18nJSON(desc), in.Sort}); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM economy_prize_group_member WHERE group_id=$1`, id); err != nil {
		return err
	}
	for i, m := range members {
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_prize_group_member (group_id, prize_id, weight, sort) VALUES ($1,$2,$3,$4)`, id, m.prize, m.weight, i); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) chanceConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/chance-plays", http.MethodGet, s.consoleChancePlays},
		{consolePrefix + "/users/{userId}/chance-plays", http.MethodGet, s.consoleUserChancePlays},
	}
}

const chancePlayAdminJSON = `jsonb_build_object('id',p.id,'gameId',p.game_id,'gameKey',g.key,'gameName',g.name,'userId',p.user_id,'username',u.username,'mode',p.mode,'status',p.status,
	'roll',p.roll::numeric/10000,'slot',p.slot,'prizeId',p.prize_id,'prizeKey',z.key,'prizeName',z.name,'prizeGroupId',p.prize_group_id,'groupKey',pg.key,'groupName',pg.name,
	'claimId',p.claim_id,'price',p.price,'error',p.error,'createdAt',p.created_at,'finishedAt',p.finished_at)::text`

const chancePlayAdminFrom = ` FROM economy_chance_play p JOIN economy_chance_game g ON g.id=p.game_id LEFT JOIN users u ON u.id=p.user_id
	LEFT JOIN economy_prize z ON z.id=p.prize_id LEFT JOIN economy_prize_group pg ON pg.id=p.prize_group_id`

func (s *service) chancePlayFilters(r *http.Request, where *sqlWhere) error {
	q := r.URL.Query()
	for _, name := range []string{"gameId", "userId", "prizeId"} {
		v := strings.TrimSpace(q.Get(name))
		if v == "" {
			continue
		}
		if _, err := uuid.FromString(v); err != nil {
			return badRequest("INVALID_ID", name, "must be a UUID")
		}
		where.add("p."+map[string]string{"gameId": "game_id", "userId": "user_id", "prizeId": "prize_id"}[name]+" = ?::uuid", v)
	}
	if v := q.Get("mode"); v != "" {
		if v != "free" && v != "ticket" {
			return badRequest("INVALID_FILTER", "mode", "free or ticket")
		}
		where.add("p.mode = ?", v)
	}
	switch q.Get("won") {
	case "true":
		where.add("p.prize_id IS NOT NULL")
	case "false":
		where.add("p.prize_id IS NULL")
	case "":
	default:
		return badRequest("INVALID_FILTER", "won", "must be true or false")
	}
	if v := q.Get("status"); v != "" {
		if v != "pending" && v != "done" {
			return badRequest("INVALID_FILTER", "status", "pending or done")
		}
		where.add("p.status = ?", v)
	}
	if v := strings.TrimSpace(q.Get("q")); v != "" {
		n := where.next()
		where.args = append(where.args, likePattern(v))
		where.parts = append(where.parts, "(COALESCE(u.username,'') ILIKE "+n+" OR g.key ILIKE "+n+" OR COALESCE(z.key,'') ILIKE "+n+")")
	}
	return nil
}

func (s *service) consoleChancePlays(w http.ResponseWriter, r *http.Request) {
	where := &sqlWhere{}
	if err := s.chancePlayFilters(r, where); err != nil {
		s.writeError(w, err)
		return
	}
	s.listRaw(w, r, `SELECT count(*)`+chancePlayAdminFrom, `SELECT `+chancePlayAdminJSON+chancePlayAdminFrom, where, ` ORDER BY p.created_at DESC, p.id DESC`, readPage(r))
}

func (s *service) consoleUserChancePlays(w http.ResponseWriter, r *http.Request) {
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	where := &sqlWhere{}
	where.add("p.user_id = ?::uuid", userID)
	if err := s.chancePlayFilters(r, where); err != nil {
		s.writeError(w, err)
		return
	}
	s.listRaw(w, r, `SELECT count(*)`+chancePlayAdminFrom, `SELECT `+chancePlayAdminJSON+chancePlayAdminFrom, where, ` ORDER BY p.created_at DESC, p.id DESC`, readPage(r))
}
