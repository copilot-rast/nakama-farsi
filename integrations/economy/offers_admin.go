package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/activity"
)

const offerAdminJSON = `(` + offerJSON + ` || jsonb_build_object('productKey',(SELECT p.key FROM economy_product p WHERE p.id=t.product_id),'productName',(SELECT p.name FROM economy_product p WHERE p.id=t.product_id),
	'productActive',(SELECT p.is_active FROM economy_product p WHERE p.id=t.product_id),
	'made',(SELECT count(*) FROM economy_user_offer u WHERE u.offer_id=t.id),'bought',(SELECT COALESCE(sum(u.bought),0) FROM economy_user_offer u WHERE u.offer_id=t.id)))`

func offerResources() []resource {
	mk := func(name, kind string) resource {
		return resource{
			name:   name,
			table:  "economy_offer",
			json:   offerAdminJSON,
			scope:  "t.kind = '" + kind + "'",
			search: []string{"t.key", translatedSearch("t.name")},
			filters: map[string]string{
				"isActive":  "t.is_active = ?::boolean",
				"productId": "t.product_id::text = ?",
			},
			save: func(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
				return saveOffer(ctx, tx, id, create, kind, raw)
			},
			remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
				return refuseIfUsed(ctx, tx, id, `SELECT EXISTS (SELECT 1 FROM economy_user_offer WHERE offer_id=$1 AND bought > 0)`)
			},
		}
	}
	return []resource{mk("offers", "daily"), mk("entry-offers", "entry")}
}

func saveOffer(ctx context.Context, tx *sql.Tx, id string, create bool, kind string, raw []byte) error {
	var in struct {
		Key             string            `json:"key"`
		Name            i18n              `json:"name"`
		Description     i18n              `json:"description"`
		Badge           i18n              `json:"badge"`
		ProductID       string            `json:"productId"`
		PriceMode       string            `json:"priceMode"`
		DiscountPercent *int              `json:"discountPercent"`
		PriceToman      *int64            `json:"priceToman"`
		CoinPrice       *coinPrice        `json:"coinPrice"`
		SKU             map[string]string `json:"sku"`
		PurchaseLimit   *int              `json:"purchaseLimit"`
		Weight          *int              `json:"weight"`
		AudienceTagIDs  []string          `json:"audienceTagIds"`
		ExcludeTagIDs   []string          `json:"excludeTagIds"`
		MinLevel        int               `json:"minLevel"`
		MaxLevel        *int              `json:"maxLevel"`
		LevelTableID    string            `json:"levelTableId"`
		MetaConditions  []struct {
			MetaID   string          `json:"metaId"`
			Operator string          `json:"operator"`
			Value    json.RawMessage `json:"value"`
		} `json:"metaConditions"`
		MinAccountDays  *int            `json:"minAccountDays"`
		MaxAccountDays  *int            `json:"maxAccountDays"`
		TriggerType     string          `json:"triggerType"`
		TriggerValue    json.RawMessage `json:"triggerValue"`
		TriggerRef      string          `json:"triggerRef"`
		DurationMinutes *int            `json:"durationMinutes"`
		Repeat          string          `json:"repeat"`
		CooldownHours   *int            `json:"cooldownHours"`
		MaxTriggers     *int            `json:"maxTriggers"`
		StartsAt        string          `json:"startsAt"`
		EndsAt          string          `json:"endsAt"`
		IsActive        *bool           `json:"isActive"`
		Sort            int             `json:"sort"`
		Metadata        json.RawMessage `json:"metadata"`
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
	if err := keepKey(ctx, tx, "economy_offer", id, create, key); err != nil {
		return err
	}
	name, err := cleanI18n(langs, "name", in.Name, false, 128)
	if err != nil {
		return err
	}
	desc, err := cleanI18n(langs, "description", in.Description, false, 1000)
	if err != nil {
		return err
	}
	badge, err := cleanI18n(langs, "badge", in.Badge, false, 32)
	if err != nil {
		return err
	}
	product, err := refExists(ctx, tx, "economy_product", "productId", "PRODUCT_NOT_FOUND", in.ProductID, true)
	if err != nil {
		return err
	}
	mode := in.PriceMode
	if mode == "" {
		mode = "discount"
	}
	if err := oneOf("priceMode", "INVALID_PRICE_MODE", mode, "discount", "price"); err != nil {
		return err
	}
	pct := 0
	if in.DiscountPercent != nil {
		pct = *in.DiscountPercent
	}
	if pct < 0 || pct > 100 {
		return badRequest("INVALID_DISCOUNT", "discountPercent", "between 0 and 100")
	}
	var priceToman *int64
	var coin any
	if mode == "price" {
		pct = 0
		if in.PriceToman != nil {
			if *in.PriceToman < 0 {
				return badRequest("INVALID_PRICE", "priceToman", "must be zero or more")
			}
			if *in.PriceToman > 0 && *in.PriceToman < 100 {
				return badRequest("INVALID_PRICE", "priceToman", "0 or at least 100 Toman")
			}
			priceToman = in.PriceToman
		}
		if in.CoinPrice != nil && (in.CoinPrice.CurrencyKey != "" || in.CoinPrice.Amount != 0) {
			if in.CoinPrice.Amount < 0 {
				return badRequest("INVALID_PRICE", "coinPrice.amount", "must be zero or more")
			}
			var ok bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, in.CoinPrice.CurrencyKey).Scan(&ok); err != nil {
				return err
			}
			if !ok {
				return badRequest("CURRENCY_NOT_FOUND", "coinPrice.currencyKey", "currency does not exist")
			}
			b, _ := json.Marshal(in.CoinPrice)
			coin = string(b)
		}
		if priceToman == nil && coin == nil {
			return badRequest("PRICE_REQUIRED", "priceToman", "set a special Toman or coin price")
		}
	}
	sku := map[string]string{}
	for k, v := range in.SKU {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if !isStore(k) {
			return badRequest("INVALID_SKU", "sku", "only bazaar and myket")
		}
		if !skuPattern.MatchString(v) {
			return badRequest("INVALID_SKU", "sku."+k, "letters, digits, dot, dash and underscore")
		}
		sku[k] = v
	}
	limit := 1
	if in.PurchaseLimit != nil {
		limit = *in.PurchaseLimit
	}
	if limit < 1 || limit > 1000 {
		return badRequest("INVALID_LIMIT", "purchaseLimit", "between 1 and 1000")
	}
	weight := 10
	if in.Weight != nil {
		weight = *in.Weight
	}
	if weight < 0 || weight > 1000000 {
		return badRequest("INVALID_WEIGHT", "weight", "between 0 and 1000000")
	}
	audience, err := checkUserTags(ctx, tx, in.AudienceTagIDs)
	if err != nil {
		return err
	}
	exclude, err := checkUserTags(ctx, tx, in.ExcludeTagIDs)
	if err != nil {
		if app, ok := asApp(err); ok {
			app.Field = "excludeTagIds"
		}
		return err
	}
	if in.MinLevel < 0 {
		return badRequest("INVALID_LEVEL", "minLevel", "must be zero or more")
	}
	if in.MaxLevel != nil && (*in.MaxLevel < 0 || *in.MaxLevel < in.MinLevel) {
		return badRequest("INVALID_LEVEL", "maxLevel", "must be at least the minimum level")
	}
	table, err := refExists(ctx, tx, "economy_level_table", "levelTableId", "LEVEL_TABLE_NOT_FOUND", in.LevelTableID, false)
	if err != nil {
		return err
	}
	if len(in.MetaConditions) > 10 {
		return badRequest("TOO_MANY_CONDITIONS", "metaConditions", "at most 10 meta conditions")
	}
	conds := []offerMetaCond{}
	for i, c := range in.MetaConditions {
		field := "metaConditions[" + strconv.Itoa(i) + "]"
		ref, err := refExists(ctx, tx, "economy_meta", field+".metaId", "META_NOT_FOUND", c.MetaID, true)
		if err != nil {
			return err
		}
		op := c.Operator
		if op == "" {
			op = "gte"
		}
		if err := oneOf(field+".operator", "INVALID_OPERATOR", op, "gt", "gte", "lt", "lte", "eq"); err != nil {
			return err
		}
		v, err := optionalNumeric(field+".value", c.Value)
		if err != nil {
			return err
		}
		if v == nil {
			return badRequest("INVALID_VALUE", field+".value", "a value is required")
		}
		conds = append(conds, offerMetaCond{MetaID: *ref, Operator: op, Value: numericFloat(*v)})
	}
	for _, f := range []struct {
		name string
		v    *int
	}{{"minAccountDays", in.MinAccountDays}, {"maxAccountDays", in.MaxAccountDays}} {
		if f.v != nil && (*f.v < 0 || *f.v > 36500) {
			return badRequest("INVALID_DAYS", f.name, "between 0 and 36500")
		}
	}
	if in.MinAccountDays != nil && in.MaxAccountDays != nil && *in.MaxAccountDays < *in.MinAccountDays {
		return badRequest("INVALID_DAYS", "maxAccountDays", "must be at least the minimum")
	}
	trigger, triggerRef := "first_entry", ""
	var triggerValue *string
	duration, cooldown, repeat := 1440, 168, "once"
	var maxTriggers *int
	if kind == "entry" {
		if in.TriggerType != "" {
			trigger = in.TriggerType
		}
		if err := oneOf("triggerType", "INVALID_TRIGGER", trigger, "first_entry", "level", "meta", "tag", "account_days", "lapsed"); err != nil {
			return err
		}
		if triggerValue, err = optionalNumeric("triggerValue", in.TriggerValue); err != nil {
			return err
		}
		switch trigger {
		case "level", "meta", "account_days", "lapsed":
			if triggerValue == nil {
				return badRequest("INVALID_VALUE", "triggerValue", "a value is required")
			}
			if trigger != "meta" && numericFloat(*triggerValue) < 0 {
				return badRequest("INVALID_VALUE", "triggerValue", "must be zero or more")
			}
		default:
			triggerValue = nil
		}
		switch trigger {
		case "level":
			ref, err := refExists(ctx, tx, "economy_level_table", "triggerRef", "LEVEL_TABLE_NOT_FOUND", in.TriggerRef, false)
			if err != nil {
				return err
			}
			if ref != nil {
				triggerRef = *ref
			}
		case "meta":
			ref, err := refExists(ctx, tx, "economy_meta", "triggerRef", "META_NOT_FOUND", in.TriggerRef, true)
			if err != nil {
				return err
			}
			triggerRef = *ref
		case "tag":
			tags, err := checkUserTags(ctx, tx, []string{in.TriggerRef})
			if err != nil || len(tags) == 0 || strings.TrimSpace(in.TriggerRef) == "" {
				return badRequest("TAG_NOT_FOUND", "triggerRef", "pick a player tag")
			}
			triggerRef = tags[0]
		}
		if in.DurationMinutes != nil {
			duration = *in.DurationMinutes
		}
		if duration < 1 || duration > 525600 {
			return badRequest("INVALID_DURATION", "durationMinutes", "between 1 minute and 365 days")
		}
		if in.Repeat != "" {
			repeat = in.Repeat
		}
		if err := oneOf("repeat", "INVALID_REPEAT", repeat, "once", "cooldown"); err != nil {
			return err
		}
		if in.CooldownHours != nil {
			cooldown = *in.CooldownHours
		}
		if cooldown < 0 || cooldown > 87600 {
			return badRequest("INVALID_COOLDOWN", "cooldownHours", "between 0 and 87600 hours")
		}
		if repeat == "cooldown" && in.MaxTriggers != nil {
			if *in.MaxTriggers < 1 || *in.MaxTriggers > 1000000 {
				return badRequest("INVALID_MAX_TRIGGERS", "maxTriggers", "between 1 and 1000000")
			}
			maxTriggers = in.MaxTriggers
		}
	}
	starts, ends, err := cleanWindow(in.StartsAt, in.EndsAt)
	if err != nil {
		return err
	}
	meta, err := cleanObject("metadata", in.Metadata)
	if err != nil {
		return err
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	skuJSON, _ := json.Marshal(sku)
	condJSON, _ := json.Marshal(conds)
	return execSave(ctx, tx, "economy_offer", id, create,
		[]string{"key", "kind", "name", "description", "badge", "product_id", "price_mode", "discount_percent", "price_toman", "coin_price", "sku", "purchase_limit", "weight",
			"audience_tag_ids", "exclude_tag_ids", "min_level", "max_level", "level_table_id", "meta_conditions", "min_account_days", "max_account_days",
			"trigger_type", "trigger_value", "trigger_ref", "duration_minutes", "repeat", "cooldown_hours", "max_triggers", "starts_at", "ends_at", "is_active", "sort", "metadata"},
		[]any{key, kind, i18nJSON(name), i18nJSON(desc), i18nJSON(badge), *product, mode, pct, priceToman, coin, string(skuJSON), limit, weight,
			audience, exclude, in.MinLevel, in.MaxLevel, table, string(condJSON), in.MinAccountDays, in.MaxAccountDays,
			trigger, triggerValue, triggerRef, duration, repeat, cooldown, maxTriggers, starts, ends, active, in.Sort, meta})
}

func (s *service) offerConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/offer-settings", http.MethodGet, s.consoleOfferSettingsGet},
		{consolePrefix + "/offer-settings", http.MethodPut, s.consoleOfferSettingsPut},
		{consolePrefix + "/offer-preview", http.MethodGet, s.consoleOfferPreview},
		{consolePrefix + "/offer-report", http.MethodGet, s.consoleOfferReport},
		{consolePrefix + "/users/{userId}/offers", http.MethodGet, s.consoleUserOffers},
		{consolePrefix + "/users/{userId}/offers/reset-daily", http.MethodPost, s.consoleUserOffersReset},
		{consolePrefix + "/users/{userId}/offers/{id}/expire", http.MethodPost, s.consoleUserOfferExpire},
		{consolePrefix + "/users/{userId}/entry-offers/{offerId}/reset", http.MethodPost, s.consoleUserEntryReset},
	}
}

func (s *service) consoleOfferSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := loadOfferSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	start, end := st.offerPeriod(s.now())
	writeJSON(w, http.StatusOK, map[string]any{"settings": st, "periodStart": start.UTC(), "periodEnd": end.UTC()})
}

func validateOfferSettings(ctx context.Context, q queryer, st *offerSettings) error {
	st.Timezone = strings.TrimSpace(st.Timezone)
	if _, err := time.LoadLocation(st.Timezone); err != nil || st.Timezone == "" || strings.EqualFold(st.Timezone, "local") {
		return badRequest("INVALID_TIMEZONE", "timezone", "use a time zone name such as Asia/Tehran or UTC")
	}
	checks := []struct {
		field    string
		v, lo, h int
	}{
		{"resetHour", st.ResetHour, 0, 23}, {"resetMinute", st.ResetMinute, 0, 59}, {"periodDays", st.PeriodDays, 1, 30}, {"count", st.Count, 0, 20},
		{"noRepeatDays", st.NoRepeatDays, 0, 365}, {"boughtCooldownDays", st.BoughtCooldownDays, -1, 3650}, {"rerollLimit", st.RerollLimit, 0, 100},
		{"entryMaxActive", st.EntryMaxActive, 1, 20},
	}
	for _, c := range checks {
		if c.v < c.lo || c.v > c.h {
			return badRequest("INVALID_SETTING", c.field, "between "+strconv.Itoa(c.lo)+" and "+strconv.Itoa(c.h))
		}
	}
	if err := oneOf("weighting", "INVALID_POLICY", st.Weighting, "weighted", "uniform", "priority"); err != nil {
		return err
	}
	if err := oneOf("inactiveProducts", "INVALID_POLICY", st.InactiveProducts, "allow", "skip"); err != nil {
		return err
	}
	if err := oneOf("entryTimerStart", "INVALID_POLICY", st.EntryTimerStart, "list", "seen", "trigger"); err != nil {
		return err
	}
	if len(st.CountByTag) > 20 {
		return badRequest("INVALID_SETTING", "countByTag", "at most 20 rows")
	}
	seen := map[string]bool{}
	for i, tc := range st.CountByTag {
		field := "countByTag[" + strconv.Itoa(i) + "]"
		var ok bool
		if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_tag WHERE id::text=$1 AND key_type='user')`, tc.TagID).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("TAG_NOT_FOUND", field+".tagId", "pick a player tag")
		}
		if seen[tc.TagID] {
			return badRequest("DUPLICATE_TAG", field+".tagId", "each tag once")
		}
		seen[tc.TagID] = true
		if tc.Count < 0 || tc.Count > 20 {
			return badRequest("INVALID_SETTING", field+".count", "between 0 and 20")
		}
	}
	if st.RerollCost != nil && (st.RerollCost.CurrencyKey == "" && st.RerollCost.Amount == 0) {
		st.RerollCost = nil
	}
	if st.RerollCost != nil {
		var ok bool
		if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, st.RerollCost.CurrencyKey).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("CURRENCY_NOT_FOUND", "rerollCost.currencyKey", "currency does not exist")
		}
		if st.RerollCost.Amount < 1 || st.RerollCost.Amount > 1000000000 {
			return badRequest("INVALID_SETTING", "rerollCost.amount", "between 1 and 1000000000")
		}
	}
	if st.RerollCostStep < 0 || st.RerollCostStep > 1000000000 {
		return badRequest("INVALID_SETTING", "rerollCostStep", "between 0 and 1000000000")
	}
	return nil
}

func (s *service) consoleOfferSettingsPut(w http.ResponseWriter, r *http.Request) {
	st, err := loadOfferSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := readBody(r, &st); err != nil {
		s.writeError(w, err)
		return
	}
	if st.CountByTag == nil {
		st.CountByTag = []offerTagCount{}
	}
	if err := validateOfferSettings(r.Context(), s.db, &st); err != nil {
		s.writeError(w, err)
		return
	}
	st.UpdatedAt = ""
	raw, _ := json.Marshal(st)
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO economy_offer_settings (id, settings, updated_at) VALUES (1,$1,now()) ON CONFLICT (id) DO UPDATE SET settings=EXCLUDED.settings, updated_at=now()`, string(raw)); err != nil {
		s.writeError(w, err)
		return
	}
	s.consoleOfferSettingsGet(w, r)
}

var offerReasonOrder = map[string]int{"picked": 0, "recent": 1, "": 2}

// consoleOfferPreview shows what a player would get now: the current daily
// set, a fresh draw for this period with the reason each template was left
// out, and the state of every entry offer.
func (s *service) consoleOfferPreview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := uuidParam(r, "userId")
	if err == nil && userID == "" {
		err = badRequest("REQUIRED", "userId", "pick a player")
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	st, err := loadOfferSettings(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	oc, err := s.newOfferCtx(ctx, userID, st)
	if err != nil {
		s.writeError(w, err)
		return
	}
	start, end := st.offerPeriod(oc.now)
	d, err := oc.drawDaily(start, 0, nil, nil)
	if err != nil {
		s.writeError(w, err)
		return
	}
	pickIdx := map[string]int{}
	for i, id := range d.Picks {
		pickIdx[id] = i
	}
	draw := []map[string]any{}
	for id, o := range d.Pool {
		reason := d.Reasons[id]
		name := o.Name
		if len(name) == 0 {
			if p, ok := oc.products[o.ProductID]; ok {
				name = p.Name
			}
		}
		row := map[string]any{"id": id, "key": o.Key, "name": name, "weight": o.Weight, "reason": reason, "picked": reason == "picked", "isActive": o.IsActive}
		if i, ok := pickIdx[id]; ok {
			row["slot"] = i
		}
		draw = append(draw, row)
	}
	sortRows(draw, func(a, b map[string]any) bool {
		ra, rb := a["reason"].(string), b["reason"].(string)
		oa, ok := offerReasonOrder[ra]
		if !ok {
			oa = 3
		}
		ob, ok := offerReasonOrder[rb]
		if !ok {
			ob = 3
		}
		if oa != ob {
			return oa < ob
		}
		if ra == "picked" {
			return a["slot"].(int) < b["slot"].(int)
		}
		return a["key"].(string) < b["key"].(string)
	})
	entries, err := oc.entryStates(false)
	if err != nil {
		s.writeError(w, err)
		return
	}
	entry := []map[string]any{}
	for _, e := range entries {
		name := e.tpl.Name
		if len(name) == 0 {
			if p, ok := oc.products[e.tpl.ProductID]; ok {
				name = p.Name
			}
		}
		entry = append(entry, map[string]any{"id": e.tpl.ID, "key": e.tpl.Key, "name": name, "triggerType": e.tpl.TriggerType, "holds": e.holds, "due": e.due && st.EntryEnabled,
			"reason": e.reason, "count": e.hist.Count, "open": e.hist.Open, "repeat": e.tpl.Repeat, "isActive": e.tpl.IsActive})
	}
	current, err := loadUserOffers(ctx, s.db, `WHERE u.user_id=$1 AND u.kind='daily' AND u.period_start=$2 ORDER BY u.slot, u.created_at`, userID, start)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var username string
	_ = s.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id=$1`, userID).Scan(&username)
	writeJSON(w, http.StatusOK, map[string]any{"userId": userID, "username": username, "periodStart": start.UTC(), "periodEnd": end.UTC(), "count": d.Count,
		"dailyEnabled": st.DailyEnabled, "entryEnabled": st.EntryEnabled, "draw": draw, "entry": entry, "current": adminOfferRows(current, oc.now)})
}

func sortRows(rows []map[string]any, less func(a, b map[string]any) bool) {
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && less(rows[j], rows[j-1]); j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}
}

func adminOfferRows(list []userOffer, now time.Time) []map[string]any {
	out := []map[string]any{}
	for _, u := range list {
		state := u.Status
		switch {
		case u.Status != "active":
		case u.soldOut():
			state = "sold_out"
		case u.expired(now):
			state = "expired"
		case u.ExpiresAt == nil:
			state = "waiting"
		}
		out = append(out, map[string]any{"id": u.ID, "offerId": u.OfferID, "key": u.Tpl.Key, "name": u.displayName(), "kind": u.Kind, "slot": u.Slot, "status": u.Status, "state": state,
			"periodStart": u.PeriodStart, "originalToman": u.OriginalToman, "finalToman": u.FinalToman, "coinCurrency": u.CoinCurrency, "coinAmount": u.CoinAmount,
			"discountPercent": u.DiscountPercent, "purchaseLimit": u.PurchaseLimit, "bought": u.Bought, "reroll": u.Reroll, "forgotten": u.Forgotten,
			"triggeredAt": u.TriggeredAt, "shownAt": u.ShownAt, "expiresAt": u.ExpiresAt, "secondsLeft": secondsLeft(now, u.ExpiresAt), "lastBoughtAt": u.LastBoughtAt})
	}
	return out
}

// consoleOfferReport counts per offer template, for offers made in the last
// `days` days: made, shown, players who bought, purchases, conversion and
// what the purchases brought in.
func (s *service) consoleOfferReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	days := 30
	if v := strings.TrimSpace(r.URL.Query().Get("days")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 3650 {
			s.writeError(w, badRequest("INVALID_FILTER", "days", "between 1 and 3650"))
			return
		}
		days = n
	}
	kind := r.URL.Query().Get("kind")
	if kind != "" && kind != "daily" && kind != "entry" {
		s.writeError(w, badRequest("INVALID_FILTER", "kind", "daily or entry"))
		return
	}
	from := s.now().Add(-time.Duration(days) * 24 * time.Hour)
	rows, err := s.db.QueryContext(ctx, `SELECT t.id::text, t.key, t.kind, (CASE WHEN t.name = '{}'::jsonb THEN (SELECT p.name FROM economy_product p WHERE p.id=t.product_id) ELSE t.name END)::text, t.is_active,
			count(u.id), count(u.shown_at), count(*) FILTER (WHERE u.bought > 0), COALESCE(sum(u.bought),0), count(DISTINCT u.user_id) FILTER (WHERE u.shown_at IS NOT NULL),
			COALESCE((SELECT sum(p.final_toman) FROM economy_payment p JOIN economy_user_offer x ON x.id=p.user_offer_id WHERE x.offer_id=t.id AND p.status='success' AND p.method<>'in_app' AND x.created_at >= $1),0),
			COALESCE((SELECT jsonb_object_agg(c, a) FROM (SELECT p.coin_currency c, sum(p.coin_amount) a FROM economy_payment p JOIN economy_user_offer x ON x.id=p.user_offer_id
				WHERE x.offer_id=t.id AND p.status='success' AND p.method='in_app' AND x.created_at >= $1 GROUP BY p.coin_currency) z),'{}')::text
		FROM economy_offer t LEFT JOIN economy_user_offer u ON u.offer_id=t.id AND u.created_at >= $1
		WHERE ($2 = '' OR t.kind = $2)
		GROUP BY t.id ORDER BY t.kind, t.sort, t.key`, from, kind)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	var tMade, tShown, tBuyers, tPurchases, tRevenue int64
	coins := map[string]int64{}
	for rows.Next() {
		var id, key, k, name, coinRaw string
		var active bool
		var made, shown, buyers, purchases, players, revenue int64
		if err := rows.Scan(&id, &key, &k, &name, &active, &made, &shown, &buyers, &purchases, &players, &revenue, &coinRaw); err != nil {
			s.writeError(w, err)
			return
		}
		c := map[string]int64{}
		_ = json.Unmarshal([]byte(coinRaw), &c)
		for cur, a := range c {
			coins[cur] += a
		}
		var conv float64
		if shown > 0 {
			conv = float64(buyers) / float64(shown)
		}
		items = append(items, map[string]any{"id": id, "key": key, "kind": k, "name": json.RawMessage(name), "isActive": active, "made": made, "shown": shown, "players": players,
			"buyers": buyers, "purchases": purchases, "conversion": conv, "revenueToman": revenue, "coins": c})
		tMade += made
		tShown += shown
		tBuyers += buyers
		tPurchases += purchases
		tRevenue += revenue
	}
	var rerolls int64
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(sum(rerolls),0) FROM economy_user_offer_period WHERE period_start >= $1`, from).Scan(&rerolls)
	var conv float64
	if tShown > 0 {
		conv = float64(tBuyers) / float64(tShown)
	}
	writeJSON(w, http.StatusOK, map[string]any{"days": days, "from": from.UTC(), "items": items,
		"totals": map[string]any{"made": tMade, "shown": tShown, "buyers": tBuyers, "purchases": tPurchases, "conversion": conv, "revenueToman": tRevenue, "coins": coins, "rerolls": rerolls}})
}

func (s *service) consoleUserOffers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	st, err := loadOfferSettings(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	now := s.now()
	start, end := st.offerPeriod(now)
	daily, err := loadUserOffers(ctx, s.db, `WHERE u.user_id=$1 AND u.kind='daily' AND u.period_start=$2 ORDER BY u.slot, u.created_at`, userID, start)
	if err != nil {
		s.writeError(w, err)
		return
	}
	entry, err := loadUserOffers(ctx, s.db, `WHERE u.user_id=$1 AND u.kind='entry' ORDER BY u.created_at DESC LIMIT 50`, userID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var rerolls int
	_ = s.db.QueryRowContext(ctx, `SELECT rerolls FROM economy_user_offer_period WHERE user_id=$1 AND period_start=$2`, userID, start).Scan(&rerolls)
	var shown, bought int64
	_ = s.db.QueryRowContext(ctx, `SELECT count(shown_at), count(*) FILTER (WHERE bought > 0) FROM economy_user_offer WHERE user_id=$1`, userID).Scan(&shown, &bought)
	writeJSON(w, http.StatusOK, map[string]any{"periodStart": start.UTC(), "periodEnd": end.UTC(), "rerolls": rerolls, "dailyEnabled": st.DailyEnabled, "entryEnabled": st.EntryEnabled,
		"daily": adminOfferRows(daily, now), "entry": adminOfferRows(entry, now), "shown": shown, "bought": bought})
}

// consoleUserOffersReset closes the player's daily offers of this period so
// the next list draws again (with the pool and settings as they are now).
func (s *service) consoleUserOffersReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	st, err := loadOfferSettings(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	start, _ := st.offerPeriod(s.now())
	var closed int64
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `UPDATE economy_user_offer SET status='closed', updated_at=now() WHERE user_id=$1 AND kind='daily' AND period_start=$2 AND status='active'`, userID, start)
		if err != nil {
			return err
		}
		closed, _ = res.RowsAffected()
		_, err = tx.ExecContext(ctx, `DELETE FROM economy_user_offer_period WHERE user_id=$1 AND period_start=$2`, userID, start)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindPurchase, Action: "offers_reset", RefID: periodKey(start), Actor: "admin", Data: map[string]any{"closed": closed}})
	writeJSON(w, http.StatusOK, map[string]any{"closed": closed})
}

func (s *service) consoleUserOfferExpire(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	id, err := cleanUUID("id", pathVar(r, "id"), true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var key string
	err = s.db.QueryRowContext(ctx, `UPDATE economy_user_offer u SET status='closed', expires_at=LEAST(COALESCE(u.expires_at, $3), $3), updated_at=now() FROM economy_offer o
		WHERE o.id=u.offer_id AND u.id=$1 AND u.user_id=$2 AND u.status='active' RETURNING o.key`, *id, userID, s.now()).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		s.writeError(w, notFound("OFFER_NOT_FOUND"))
		return
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindPurchase, Action: "offer_closed", RefID: key, Actor: "admin", Data: map[string]any{"offer": key}})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// consoleUserEntryReset forgets the player's history with an entry offer so
// its trigger can fire again; open copies close. Purchases stay in reports.
func (s *service) consoleUserEntryReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	offerID, err := cleanUUID("offerId", pathVar(r, "offerId"), true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var key string
	if err := s.db.QueryRowContext(ctx, `SELECT key FROM economy_offer WHERE id=$1 AND kind='entry'`, *offerID).Scan(&key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.writeError(w, notFound("NOT_FOUND"))
			return
		}
		s.writeError(w, err)
		return
	}
	res, err := s.db.ExecContext(ctx, `UPDATE economy_user_offer SET forgotten=true, status=CASE WHEN status='active' THEN 'closed' ELSE status END, updated_at=now() WHERE user_id=$1 AND offer_id=$2 AND NOT forgotten`, userID, *offerID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	n, _ := res.RowsAffected()
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindPurchase, Action: "entry_offer_reset", RefID: key, Actor: "admin", Data: map[string]any{"offer": key}})
	writeJSON(w, http.StatusOK, map[string]any{"reset": n})
}
