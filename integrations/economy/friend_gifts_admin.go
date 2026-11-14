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

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/activity"
)

func friendGiftResources() []resource {
	return []resource{{
		name:    "friend-gift-rules",
		table:   "economy_friend_gift_rule",
		json:    friendGiftRuleJSON,
		search:  []string{"t.name"},
		filters: map[string]string{"isActive": "t.is_active = ?::boolean"},
		save:    saveFriendGiftRule,
		remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
			return nil
		},
	}}
}

const friendGiftRuleJSON = `jsonb_build_object('id',t.id,'name',t.name,'tagIds',to_jsonb(t.tag_ids),'priority',t.priority,'perSenderDaily',t.per_sender_daily,
	'perRecipientDaily',t.per_recipient_daily,'perPairDaily',t.per_pair_daily,'canSend',t.can_send,'isActive',t.is_active,'createdAt',t.created_at,'updatedAt',t.updated_at,
	'players',(SELECT count(DISTINCT ut.user_id) FROM base_user_tag ut WHERE ut.tag_id::text = ANY(t.tag_ids)))`

func saveFriendGiftRule(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Name              string   `json:"name"`
		TagIDs            []string `json:"tagIds"`
		Priority          int      `json:"priority"`
		PerSenderDaily    *int     `json:"perSenderDaily"`
		PerRecipientDaily *int     `json:"perRecipientDaily"`
		PerPairDaily      *int     `json:"perPairDaily"`
		CanSend           *bool    `json:"canSend"`
		IsActive          *bool    `json:"isActive"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	name := strings.TrimSpace(in.Name)
	if n := len([]rune(name)); n == 0 || n > 100 {
		return badRequest("INVALID_NAME", "name", "1 to 100 letters")
	}
	if in.Priority < -1000 || in.Priority > 1000 {
		return badRequest("INVALID_PRIORITY", "priority", "between -1000 and 1000")
	}
	for field, v := range map[string]*int{"perSenderDaily": in.PerSenderDaily, "perRecipientDaily": in.PerRecipientDaily, "perPairDaily": in.PerPairDaily} {
		if v != nil && (*v < 0 || *v > 10000) {
			return badRequest("INVALID_LIMIT", field, "between 0 and 10000")
		}
	}
	tags, err := checkUserTags(ctx, tx, in.TagIDs)
	if err != nil {
		var app *appError
		if errors.As(err, &app) {
			app.Field = "tagIds"
		}
		return err
	}
	if len(tags) == 0 {
		return badRequest("TAGS_REQUIRED", "tagIds", "pick at least one player tag")
	}
	canSend := in.CanSend == nil || *in.CanSend
	active := in.IsActive == nil || *in.IsActive
	return execSave(ctx, tx, "economy_friend_gift_rule", id, create,
		[]string{"name", "tag_ids", "priority", "per_sender_daily", "per_recipient_daily", "per_pair_daily", "can_send", "is_active"},
		[]any{name, tags, in.Priority, in.PerSenderDaily, in.PerRecipientDaily, in.PerPairDaily, canSend, active})
}

func (s *service) friendGiftConsoleRoutes() []consoleRoute {
	return []consoleRoute{
		{consolePrefix + "/friend-gift-settings", http.MethodGet, s.consoleFriendGiftSettingsGet},
		{consolePrefix + "/friend-gift-settings", http.MethodPut, s.consoleFriendGiftSettingsPut},
		{consolePrefix + "/friend-gift-report", http.MethodGet, s.consoleFriendGiftReport},
		{consolePrefix + "/friend-gifts", http.MethodGet, s.consoleFriendGiftList},
		{consolePrefix + "/friend-gifts/{id}", http.MethodGet, s.consoleFriendGiftGet},
		{consolePrefix + "/friend-gifts/{id}/reverse", http.MethodPost, s.consoleFriendGiftReverse},
		{consolePrefix + "/users/{userId}/friend-gifts", http.MethodGet, s.consoleUserFriendGifts},
	}
}

func (s *service) consoleFriendGiftSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := loadFriendGiftSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"settings": st, "day": s.giftDay(r.Context())})
}

func validateFriendGiftSettings(ctx context.Context, q queryer, st *friendGiftSettings) error {
	ranges := []struct {
		field    string
		v, lo, h int
	}{
		{"minFriendDays", st.MinFriendDays, 0, 3650}, {"minMutualFriends", st.MinMutualFriends, 0, 1000}, {"minSenderLevel", st.MinSenderLevel, 0, 10000},
		{"minSenderAccountDays", st.MinSenderAccountDays, 0, 3650}, {"minRecipientAccountDays", st.MinRecipientAccountDays, 0, 3650},
		{"perSenderDaily", st.PerSenderDaily, 0, 10000}, {"perRecipientDaily", st.PerRecipientDaily, 0, 10000}, {"perPairDaily", st.PerPairDaily, 0, 10000},
		{"expiryHours", st.ExpiryHours, 1, 8760}, {"messageMaxLength", st.MessageMaxLength, 1, 500},
	}
	for _, c := range ranges {
		if c.v < c.lo || c.v > c.h {
			return badRequest("INVALID_SETTING", c.field, "between "+strconv.Itoa(c.lo)+" and "+strconv.Itoa(c.h))
		}
	}
	policies := []struct {
		field, v string
		allowed  []string
	}{
		{"recipients", st.Recipients, []string{"friends", "any"}},
		{"equipped", st.Equipped, []string{"block", "unequip"}},
		{"upgradeLevel", st.UpgradeLevel, []string{"keep", "reset"}},
		{"regift", st.Regift, []string{"allow", "block"}},
		{"itemRequirements", st.ItemRequirements, []string{"check", "ignore"}},
		{"stackableProducts", st.StackableProducts, []string{"allow", "block"}},
		{"partialOwned", st.PartialOwned, []string{"block", "allow"}},
		{"productRules", st.ProductRules, []string{"recipient", "sender", "both", "none"}},
		{"matchPolicy", st.MatchPolicy, []string{"priority", "generous"}},
		{"delivery", st.Delivery, []string{"instant", "accept"}},
		{"onExpiry", st.OnExpiry, []string{"return", "accept"}},
		{"coinReturn", st.CoinReturn, []string{"refund", "sender", "none"}},
		{"moneyReturn", st.MoneyReturn, []string{"sender", "refund_due", "none"}},
		{"messageFilter", st.MessageFilter, []string{"rules", "block", "off"}},
		{"notifyChannel", st.NotifyChannel, []string{"none", "in_app", "push_in_app"}},
	}
	for _, p := range policies {
		if err := oneOf(p.field, "INVALID_POLICY", p.v, p.allowed...); err != nil {
			return err
		}
	}
	st.LevelTableID = strings.TrimSpace(st.LevelTableID)
	if st.LevelTableID != "" {
		var ok bool
		if _, err := uuid.FromString(st.LevelTableID); err == nil {
			if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_level_table WHERE id=$1)`, st.LevelTableID).Scan(&ok); err != nil {
				return err
			}
		}
		if !ok {
			return badRequest("LEVEL_TABLE_NOT_FOUND", "levelTableId", "level table does not exist")
		}
	}
	if len(st.ItemTypes) > 100 {
		return badRequest("INVALID_SETTING", "itemTypes", "at most 100 rows")
	}
	seen := map[string]bool{}
	for i, t := range st.ItemTypes {
		field := "itemTypes[" + strconv.Itoa(i) + "].typeId"
		var ok bool
		if _, err := uuid.FromString(t.TypeID); err == nil {
			if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM base_category WHERE id=$1 AND key_type='item_type')`, t.TypeID).Scan(&ok); err != nil {
				return err
			}
		}
		if !ok {
			return badRequest("ITEM_TYPE_NOT_FOUND", field, "pick an item type")
		}
		if seen[t.TypeID] {
			return badRequest("DUPLICATE_TYPE", field, "each item type once")
		}
		seen[t.TypeID] = true
	}
	if len(st.Currencies) > 20 {
		return badRequest("INVALID_SETTING", "currencies", "at most 20 rows")
	}
	seen = map[string]bool{}
	for i, c := range st.Currencies {
		field := "currencies[" + strconv.Itoa(i) + "]"
		var ok bool
		if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, c.CurrencyKey).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("CURRENCY_NOT_FOUND", field+".currencyKey", "currency does not exist")
		}
		if seen[c.CurrencyKey] {
			return badRequest("DUPLICATE_CURRENCY", field+".currencyKey", "each currency once")
		}
		seen[c.CurrencyKey] = true
		if c.Min < 0 || c.Min > 1_000_000_000 {
			return badRequest("INVALID_AMOUNT", field+".min", "between 0 and 1000000000")
		}
		if c.Max < 0 || c.Max > 1_000_000_000 || (c.Max > 0 && c.Max < c.Min) {
			return badRequest("INVALID_AMOUNT", field+".max", "0 (no limit) or at least the minimum")
		}
	}
	seen = map[string]bool{}
	for _, m := range st.Methods {
		if err := oneOf("methods", "INVALID_POLICY", m, giftMethodKinds...); err != nil {
			return err
		}
		if seen[m] {
			return badRequest("INVALID_POLICY", "methods", "each method once")
		}
		seen[m] = true
	}
	for field, key := range map[string]*string{"sendBonusMeta": &st.SendBonusMeta, "receiveBonusMeta": &st.ReceiveBonusMeta} {
		*key = strings.TrimSpace(*key)
		if *key == "" {
			continue
		}
		var ok bool
		if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_meta WHERE key=$1)`, *key).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("META_NOT_FOUND", field, "no meta with this key")
		}
	}
	return nil
}

func (s *service) consoleFriendGiftSettingsPut(w http.ResponseWriter, r *http.Request) {
	st, err := loadFriendGiftSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := readBody(r, &st); err != nil {
		s.writeError(w, err)
		return
	}
	if st.ItemTypes == nil {
		st.ItemTypes = []giftItemType{}
	}
	if st.Currencies == nil {
		st.Currencies = []giftCurrency{}
	}
	if st.Methods == nil {
		st.Methods = []string{}
	}
	if err := validateFriendGiftSettings(r.Context(), s.db, &st); err != nil {
		s.writeError(w, err)
		return
	}
	st.UpdatedAt = ""
	raw, _ := json.Marshal(st)
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO economy_friend_gift_settings (id, settings, updated_at) VALUES (1,$1,now()) ON CONFLICT (id) DO UPDATE SET settings=EXCLUDED.settings, updated_at=now()`, string(raw)); err != nil {
		s.writeError(w, err)
		return
	}
	s.consoleFriendGiftSettingsGet(w, r)
}

func (s *service) consoleFriendGiftList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	q := r.URL.Query()
	where := &sqlWhere{}
	for _, name := range []string{"userId", "senderId", "recipientId", "itemId", "productId"} {
		v, err := uuidParam(r, name)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if v == "" {
			continue
		}
		switch name {
		case "userId":
			n := where.next()
			where.args = append(where.args, v)
			where.parts = append(where.parts, "(g.sender_id = "+n+"::uuid OR g.recipient_id = "+n+"::uuid)")
		case "senderId":
			where.add("g.sender_id = ?::uuid", v)
		case "recipientId":
			where.add("g.recipient_id = ?::uuid", v)
		case "itemId":
			where.add("g.item_id = ?::uuid", v)
		case "productId":
			where.add("g.product_id = ?::uuid", v)
		}
	}
	filters := []struct {
		name, col string
		allowed   []string
	}{
		{"status", "g.status", []string{"sending", "paying", "pending", "delivered", "declined", "expired", "cancelled", "reversed", "failed"}},
		{"kind", "g.kind", []string{"item", "currency", "product"}},
		{"source", "g.source", []string{"inventory", "shop"}},
		{"outcome", "g.outcome", []string{"delivered", "returned", "refunded", "sender_got", "refund_due", "kept"}},
	}
	for _, f := range filters {
		if v := q.Get(f.name); v != "" {
			if err := oneOf(f.name, "INVALID_FILTER", v, f.allowed...); err != nil {
				s.writeError(w, err)
				return
			}
			where.add(f.col+" = ?", v)
		}
	}
	if q.Get("flagged") == "true" {
		where.add("g.message_flagged = ?::boolean", "true")
	}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(COALESCE(su.username,'') ILIKE "+n+" OR COALESCE(ru.username,'') ILIKE "+n+" OR g.item_key ILIKE "+n+" OR g.currency_key ILIKE "+n+" OR COALESCE(p.key,'') ILIKE "+n+" OR g.message ILIKE "+n+")")
	}
	s.listRaw(w, r, `SELECT count(*)`+friendGiftFrom, `SELECT `+friendGiftJSON+`::text`+friendGiftFrom, where, ` ORDER BY g.created_at DESC, g.id DESC`, p)
}

func (s *service) consoleFriendGiftGet(w http.ResponseWriter, r *http.Request) {
	row, err := s.giftRow(r.Context(), pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	if id, _ := row["id"].(string); id != "" {
		var events []json.RawMessage
		rows, err := s.db.QueryContext(r.Context(), `SELECT jsonb_build_object('userId',user_id,'action',action,'actor',actor,'data',data,'createdAt',created_at)::text FROM base_activity WHERE ref_id=$1 ORDER BY created_at, id LIMIT 50`, id)
		if err == nil {
			for rows.Next() {
				var raw string
				if rows.Scan(&raw) == nil {
					events = append(events, json.RawMessage(raw))
				}
			}
			rows.Close()
		}
		if events == nil {
			events = []json.RawMessage{}
		}
		row["trail"] = events
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *service) consoleFriendGiftReverse(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ReturnTo string `json:"returnTo"`
		Note     string `json:"note"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if in.ReturnTo == "" {
		in.ReturnTo = "sender"
	}
	if err := oneOf("returnTo", "INVALID_POLICY", in.ReturnTo, "sender", "refund", "none"); err != nil {
		s.writeError(w, err)
		return
	}
	note := strings.TrimSpace(in.Note)
	if len([]rune(note)) > 500 {
		s.writeError(w, badRequest("TEXT_TOO_LONG", "note", "at most 500 letters"))
		return
	}
	if err := s.reverseFriendGift(r.Context(), pathVar(r, "id"), consoleName(r), in.ReturnTo, note); err != nil {
		s.writeError(w, err)
		return
	}
	s.consoleFriendGiftGet(w, r)
}

// reverseFriendGift lets an admin undo a gift: a pending gift is cancelled, a
// delivered one is taken back from the recipient. What was taken goes to the
// sender, is refunded (coin purchases) or is kept by nobody.
func (s *service) reverseFriendGift(ctx context.Context, id, admin, returnTo, note string) error {
	st, err := loadFriendGiftSettings(ctx, s.db)
	if err != nil {
		return err
	}
	g, err := loadFriendGift(ctx, s.db, id, false)
	if err != nil {
		return err
	}
	delivered := g.Status == "delivered" && g.Outcome == "delivered"
	if !delivered && g.Status != "pending" {
		return conflict("GIFT_NOT_REVERSIBLE", "only pending or delivered gifts can be reversed")
	}
	if returnTo == "refund" && !(g.Kind == "product" && g.Method == "in_app") {
		return badRequest("INVALID_POLICY", "returnTo", "only coin purchases can be refunded")
	}
	var userPrize string
	var rewards []prizeItem
	if delivered && g.Kind != "item" {
		take := map[string]int64{}
		if g.Kind == "currency" {
			take[g.CurrencyKey] = -g.Amount
		} else {
			var raw string
			err := s.db.QueryRowContext(ctx, `SELECT id::text, rewards::text FROM economy_user_prize WHERE source_type='friend_gift' AND source_id=$1 AND claimed`, g.ID).Scan(&userPrize, &raw)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			_ = json.Unmarshal([]byte(raw), &rewards)
			for _, it := range rewards {
				if it.Type == "currency" {
					if v, ok := int64FromJSON(it.Value); ok && v > 0 {
						take[it.Key] -= v
					}
				}
			}
		}
		if len(take) > 0 {
			err := s.inTx(ctx, func(tx *sql.Tx) error {
				return s.walletChange(ctx, tx, g.RecipientID, giftGrantID(g, "reverse"), take, map[string]any{"reason": "friend_gift_reverse", "giftId": g.ID, "admin": admin})
			})
			if err != nil {
				return err
			}
		}
	}
	outcome := ""
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		cur, err := loadFriendGift(ctx, tx, id, true)
		if err != nil {
			return err
		}
		if cur.Status != g.Status || cur.Outcome != g.Outcome {
			return conflict("GIFT_CHANGED", "the gift changed meanwhile; reload and try again")
		}
		if delivered && g.Kind == "item" {
			res, err := tx.ExecContext(ctx, `DELETE FROM economy_user_item WHERE user_id=$1 AND item_id=$2 AND source='gift' AND source_ref=$3`, g.RecipientID, g.ItemID.String, g.ID)
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n == 0 {
				return conflict("ITEM_GONE", "the recipient no longer has this item")
			}
		}
		if delivered && g.Kind == "product" && userPrize != "" {
			if _, err := tx.ExecContext(ctx, `DELETE FROM economy_user_item WHERE user_id=$1 AND source='prize' AND source_ref=$2`, g.RecipientID, userPrize); err != nil {
				return err
			}
			for _, it := range rewards {
				if it.Type == "meta" && it.Action != "set" {
					if _, err := tx.ExecContext(ctx, `UPDATE economy_user_meta um SET value=um.value - $3::numeric, updated_at=now() FROM economy_meta m WHERE m.id=um.meta_id AND m.key=$2 AND um.user_id=$1`,
						g.RecipientID, it.Key, string(it.Value)); err != nil {
						return err
					}
				}
			}
			if _, err := tx.ExecContext(ctx, `UPDATE economy_user_prize SET once_key=NULL, updated_at=now() WHERE id=$1`, userPrize); err != nil {
				return err
			}
		}
		if g.Kind == "item" {
			if returnTo == "sender" {
				if err := restoreGiftItem(ctx, tx, g); err != nil {
					return err
				}
				outcome = "returned"
			} else {
				outcome = "kept"
			}
		}
		_, err = tx.ExecContext(ctx, `UPDATE economy_friend_gift SET status='reversed', outcome=$2, reversed_at=now(), reversed_by=$3, reverse_note=$4, updated_at=now() WHERE id=$1`,
			g.ID, outcome, admin, note)
		return err
	})
	if err != nil {
		return err
	}
	g.Status = "reversed"
	if g.Kind != "item" {
		action := "kept"
		switch {
		case returnTo == "sender" && g.Kind == "currency":
			action = "returned"
		case returnTo == "sender":
			action = "sender_got"
		case returnTo == "refund":
			action = "refunded"
		}
		if action != "kept" {
			if outcome, err = s.giveBack(ctx, st, g, action); err != nil {
				_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET error=$2, updated_at=now() WHERE id=$1`, g.ID, truncate(err.Error(), 500))
				return err
			}
		} else {
			outcome = "kept"
		}
		_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET outcome=$2, error=NULL, updated_at=now() WHERE id=$1`, g.ID, outcome)
	}
	data := map[string]any{"admin": admin, "returnTo": returnTo, "outcome": outcome, "note": note}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: g.RecipientID, Kind: activity.KindPrize, Action: "gift_reversed", RefID: g.ID, Actor: "admin", Data: data})
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: g.SenderID, Kind: activity.KindPrize, Action: "gift_reversed", RefID: g.ID, Actor: "admin", Data: data})
	s.giftSocket(ctx, st, g.RecipientID, "reversed", g)
	s.giftSocket(ctx, st, g.SenderID, "reversed", g)
	return nil
}

func (s *service) consoleFriendGiftReport(w http.ResponseWriter, r *http.Request) {
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
	from := s.now().Add(-time.Duration(days) * 24 * time.Hour)
	statuses := map[string]int64{}
	rows, err := s.db.QueryContext(ctx, `SELECT status, count(*) FROM economy_friend_gift WHERE created_at >= $1 GROUP BY status`, from)
	if err != nil {
		s.writeError(w, err)
		return
	}
	for rows.Next() {
		var st string
		var n int64
		if rows.Scan(&st, &n) == nil {
			statuses[st] = n
		}
	}
	rows.Close()
	kinds := map[string]int64{}
	rows, err = s.db.QueryContext(ctx, `SELECT kind, count(*) FROM economy_friend_gift WHERE created_at >= $1 AND `+giftCounted+` GROUP BY kind`, from)
	if err != nil {
		s.writeError(w, err)
		return
	}
	for rows.Next() {
		var k string
		var n int64
		if rows.Scan(&k, &n) == nil {
			kinds[k] = n
		}
	}
	rows.Close()
	var sent, senders, recipients, revenue int64
	_ = s.db.QueryRowContext(ctx, `SELECT count(*), count(DISTINCT sender_id), count(DISTINCT recipient_id) FROM economy_friend_gift WHERE created_at >= $1 AND `+giftCounted+` AND status NOT IN ('sending','paying')`, from).Scan(&sent, &senders, &recipients)
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(sum(p.final_toman),0) FROM economy_payment p JOIN economy_friend_gift g ON g.id=p.friend_gift_id WHERE g.created_at >= $1 AND p.status='success' AND p.method<>'in_app'`, from).Scan(&revenue)
	coins := map[string]int64{}
	var coinRaw string
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(jsonb_object_agg(c, a),'{}')::text FROM (SELECT p.coin_currency c, sum(p.coin_amount) a FROM economy_payment p JOIN economy_friend_gift g ON g.id=p.friend_gift_id
		WHERE g.created_at >= $1 AND p.status='success' AND p.method='in_app' GROUP BY p.coin_currency) z`, from).Scan(&coinRaw)
	_ = json.Unmarshal([]byte(coinRaw), &coins)
	transferred := map[string]int64{}
	var trRaw string
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(jsonb_object_agg(currency_key, a),'{}')::text FROM (SELECT currency_key, sum(amount) a FROM economy_friend_gift
		WHERE created_at >= $1 AND kind='currency' AND status='delivered' GROUP BY currency_key) z`, from).Scan(&trRaw)
	_ = json.Unmarshal([]byte(trRaw), &transferred)
	top := func(query string) []json.RawMessage {
		out := []json.RawMessage{}
		rows, err := s.db.QueryContext(ctx, query, from)
		if err != nil {
			return out
		}
		defer rows.Close()
		for rows.Next() {
			var raw string
			if rows.Scan(&raw) == nil {
				out = append(out, json.RawMessage(raw))
			}
		}
		return out
	}
	topSenders := top(`SELECT jsonb_build_object('userId',g.sender_id,'username',u.username,'displayName',u.display_name,'count',count(*))::text FROM economy_friend_gift g LEFT JOIN users u ON u.id=g.sender_id
		WHERE g.created_at >= $1 AND g.` + giftCounted + ` AND g.status NOT IN ('sending','paying') GROUP BY g.sender_id, u.username, u.display_name ORDER BY count(*) DESC, u.username LIMIT 10`)
	topRecipients := top(`SELECT jsonb_build_object('userId',g.recipient_id,'username',u.username,'displayName',u.display_name,'count',count(*))::text FROM economy_friend_gift g LEFT JOIN users u ON u.id=g.recipient_id
		WHERE g.created_at >= $1 AND g.status='delivered' GROUP BY g.recipient_id, u.username, u.display_name ORDER BY count(*) DESC, u.username LIMIT 10`)
	topThings := top(`SELECT jsonb_build_object('kind',kind,'key',key,'name',name,'count',n)::text FROM (
		SELECT g.kind, COALESCE(NULLIF(g.item_key,''),NULLIF(g.currency_key,''),p.key) AS key, COALESCE(i.name,c.name,p.name) AS name, count(*) AS n
		FROM economy_friend_gift g LEFT JOIN economy_item i ON i.id=g.item_id LEFT JOIN economy_currency c ON g.kind='currency' AND c.key=g.currency_key LEFT JOIN economy_product p ON p.id=g.product_id
		WHERE g.created_at >= $1 AND g.status='delivered' GROUP BY 1, 2, 3) z ORDER BY n DESC, key LIMIT 10`)
	daily := top(`SELECT jsonb_build_object('day',day,'sent',count(*) FILTER (WHERE status NOT IN ('failed','cancelled','sending','paying')),'delivered',count(*) FILTER (WHERE status='delivered'))::text
		FROM economy_friend_gift WHERE created_at >= $1 GROUP BY day ORDER BY day`)
	writeJSON(w, http.StatusOK, map[string]any{"days": days, "from": from.UTC(), "statuses": statuses, "kinds": kinds,
		"totals":     map[string]any{"sent": sent, "senders": senders, "recipients": recipients, "revenueToman": revenue, "coins": coins, "transferred": transferred, "delivered": statuses["delivered"], "pending": statuses["pending"], "reversed": statuses["reversed"]},
		"topSenders": topSenders, "topRecipients": topRecipients, "topThings": topThings, "daily": daily})
}

func (s *service) consoleUserFriendGifts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := s.consoleUser(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	st, err := loadFriendGiftSettings(ctx, s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	day := s.giftDay(ctx)
	var sent, received, pendingIn, sentToday, receivedToday int64
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE sender_id=$1 AND `+giftCounted+` AND status NOT IN ('sending','paying')), count(*) FILTER (WHERE recipient_id=$1 AND status='delivered'),
		count(*) FILTER (WHERE recipient_id=$1 AND status='pending'), count(*) FILTER (WHERE sender_id=$1 AND day=$2::date AND `+giftCounted+`), count(*) FILTER (WHERE recipient_id=$1 AND day=$2::date AND `+giftCounted+`)
		FROM economy_friend_gift WHERE sender_id=$1 OR recipient_id=$1`, userID, day).Scan(&sent, &received, &pendingIn, &sentToday, &receivedToday)
	lim, err := s.giftLimitsFor(ctx, s.db, st, userID, "")
	if err != nil {
		s.writeError(w, err)
		return
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+friendGiftJSON+`::text`+friendGiftFrom+` WHERE g.sender_id=$1 OR g.recipient_id=$1 ORDER BY g.created_at DESC, g.id DESC LIMIT 10`, userID)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	recent := []json.RawMessage{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			s.writeError(w, err)
			return
		}
		recent = append(recent, json.RawMessage(raw))
	}
	writeJSON(w, http.StatusOK, map[string]any{"sent": sent, "received": received, "pendingIn": pendingIn, "sentToday": sentToday, "receivedToday": receivedToday,
		"limits": map[string]any{"sender": lim.Sender, "senderBonus": lim.SenderBonus, "canSend": lim.CanSend, "ruleId": lim.RuleID, "ruleName": lim.RuleName}, "day": day, "enabled": st.Enabled, "recent": recent})
}
