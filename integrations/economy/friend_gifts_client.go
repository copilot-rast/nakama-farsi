package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func (s *service) friendGiftRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_friend_gift_options": s.rpcFriendGiftOptions,
		"economy_friend_gift_send":    s.rpcFriendGiftSend,
		"economy_friend_gifts":        s.rpcFriendGifts,
		"economy_friend_gift_accept":  s.rpcFriendGiftRespond("accept"),
		"economy_friend_gift_decline": s.rpcFriendGiftRespond("decline"),
		"economy_friend_gift_cancel":  s.rpcFriendGiftRespond("cancel"),
	}
}

const friendGiftJSON = `jsonb_build_object('id',g.id,'senderId',g.sender_id,'senderName',COALESCE(NULLIF(su.display_name,''),su.username),'senderUsername',su.username,
	'recipientId',g.recipient_id,'recipientName',COALESCE(NULLIF(ru.display_name,''),ru.username),'recipientUsername',ru.username,
	'source',g.source,'kind',g.kind,'itemId',g.item_id,'itemKey',g.item_key,'itemName',i.name,'itemIcon',i.icon_url,'rarity',i.rarity,
	'currencyKey',g.currency_key,'currencyName',c.name,'currencyIcon',c.icon_url,'amount',g.amount,
	'productId',g.product_id,'productKey',p.key,'productName',p.name,'productIcon',p.icon_url,'prizeItems',z.items,
	'paymentId',g.payment_id,'method',g.method,'paymentStatus',pay.status,'finalToman',pay.final_toman,'coinCurrency',pay.coin_currency,'coinAmount',pay.coin_amount,
	'message',g.message,'messageFlagged',g.message_flagged,'delivery',g.delivery,'status',g.status,'outcome',g.outcome,'error',g.error,'requestId',g.request_id,'day',g.day,
	'expiresAt',g.expires_at,'deliveredAt',g.delivered_at,'respondedAt',g.responded_at,'reversedAt',g.reversed_at,'reversedBy',g.reversed_by,'reverseNote',g.reverse_note,
	'itemState',g.item_state,'createdAt',g.created_at,'updatedAt',g.updated_at)`

const friendGiftFrom = ` FROM economy_friend_gift g LEFT JOIN users su ON su.id=g.sender_id LEFT JOIN users ru ON ru.id=g.recipient_id
	LEFT JOIN economy_item i ON i.id=g.item_id LEFT JOIN economy_currency c ON g.kind='currency' AND c.key=g.currency_key
	LEFT JOIN economy_product p ON p.id=g.product_id LEFT JOIN economy_prize z ON z.id=g.prize_id LEFT JOIN economy_payment pay ON pay.id=g.payment_id`

func resolveText(langs *localized.Set, lang string, v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	t := localized.Text{}
	for k, x := range m {
		if str, ok := x.(string); ok {
			t[k] = str
		}
	}
	return langs.Resolve(t, lang)
}

// clientGift turns an admin gift row into what a player sees: no prices for
// the recipient, no moderation fields, names in the player's language.
func clientGift(raw map[string]any, userID string, langs *localized.Set, lang string, now time.Time) map[string]any {
	box := "sent"
	if raw["recipientId"] == userID {
		box = "received"
	}
	name := ""
	icon := ""
	switch raw["kind"] {
	case "item":
		name, icon = resolveText(langs, lang, raw["itemName"]), str(raw["itemIcon"])
	case "currency":
		name, icon = resolveText(langs, lang, raw["currencyName"]), str(raw["currencyIcon"])
	case "product":
		name, icon = resolveText(langs, lang, raw["productName"]), str(raw["productIcon"])
	}
	out := map[string]any{
		"id": raw["id"], "box": box, "kind": raw["kind"], "source": raw["source"], "status": raw["status"], "outcome": raw["outcome"], "delivery": raw["delivery"],
		"name": name, "iconUrl": icon, "itemKey": raw["itemKey"], "currencyKey": raw["currencyKey"], "amount": raw["amount"], "productKey": raw["productKey"],
		"message": raw["message"], "createdAt": raw["createdAt"], "expiresAt": raw["expiresAt"], "deliveredAt": raw["deliveredAt"], "respondedAt": raw["respondedAt"],
		"from": map[string]any{"id": raw["senderId"], "username": raw["senderUsername"], "displayName": raw["senderName"]},
		"to":   map[string]any{"id": raw["recipientId"], "username": raw["recipientUsername"], "displayName": raw["recipientName"]},
	}
	if raw["kind"] == "product" {
		out["items"] = raw["prizeItems"]
	}
	if box == "sent" {
		out["method"], out["paymentId"], out["paymentStatus"] = raw["method"], raw["paymentId"], raw["paymentStatus"]
	}
	if raw["status"] == "pending" {
		if at, ok := raw["expiresAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339Nano, at); err == nil {
				out["secondsLeft"] = secondsLeft(now, &t)
			}
		}
	}
	return out
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func (s *service) giftRow(ctx context.Context, id string) (map[string]any, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT `+friendGiftJSON+`::text`+friendGiftFrom+` WHERE g.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("GIFT_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	return out, json.Unmarshal([]byte(raw), &out)
}

func (s *service) giftClientView(ctx context.Context, userID, id, want string) (map[string]any, error) {
	raw, err := s.giftRow(ctx, id)
	if err != nil {
		return nil, err
	}
	if raw["senderId"] != userID && raw["recipientId"] != userID {
		return nil, notFound("GIFT_NOT_FOUND")
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, want)
	if err != nil {
		return nil, err
	}
	return clientGift(raw, userID, langs, lang, s.now()), nil
}

func (s *service) rpcFriendGiftSend(ctx context.Context, userID, payload string) (any, error) {
	var in giftSendInput
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	return s.sendFriendGift(ctx, userID, in)
}

func (s *service) rpcFriendGiftRespond(action string) clientRPC {
	return func(ctx context.Context, userID, payload string) (any, error) {
		var in struct {
			GiftID string `json:"giftId"`
			Lang   string `json:"lang"`
		}
		if err := decodeStrict(payload, &in); err != nil {
			return nil, err
		}
		g, err := s.respondGift(ctx, userID, in.GiftID, action)
		if err != nil {
			return nil, err
		}
		view, err := s.giftClientView(ctx, userID, g.ID, in.Lang)
		if err != nil {
			return nil, err
		}
		return map[string]any{"gift": view}, nil
	}
}

func (s *service) rpcFriendGifts(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Box    string `json:"box"`
		Status string `json:"status"`
		Limit  int    `json:"limit"`
		Cursor string `json:"cursor"`
		Lang   string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if in.Box == "" {
		in.Box = "received"
	}
	if err := oneOf("box", "INVALID_BOX", in.Box, "received", "sent", "all"); err != nil {
		return nil, err
	}
	if in.Status != "" {
		if err := oneOf("status", "INVALID_STATUS", in.Status, "pending", "delivered", "declined", "expired", "cancelled", "reversed", "failed", "paying"); err != nil {
			return nil, err
		}
	}
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	before := s.now().Add(time.Hour)
	beforeID := "ffffffff-ffff-ffff-ffff-ffffffffffff"
	if in.Cursor != "" {
		raw, err := base64.RawURLEncoding.DecodeString(in.Cursor)
		parts := strings.SplitN(string(raw), "|", 2)
		if err != nil || len(parts) != 2 {
			return nil, badRequest("INVALID_CURSOR", "cursor", "invalid cursor")
		}
		t, err := time.Parse(time.RFC3339Nano, parts[0])
		if _, idErr := uuid.FromString(parts[1]); err != nil || idErr != nil {
			return nil, badRequest("INVALID_CURSOR", "cursor", "invalid cursor")
		}
		before, beforeID = t, parts[1]
	}
	who := map[string]string{"received": "g.recipient_id=$1 AND g.status NOT IN ('sending','paying','failed')", "sent": "g.sender_id=$1", "all": "(g.sender_id=$1 OR (g.recipient_id=$1 AND g.status NOT IN ('sending','paying','failed')))"}[in.Box]
	rows, err := s.db.QueryContext(ctx, `SELECT `+friendGiftJSON+`::text, g.created_at, g.id::text`+friendGiftFrom+`
		WHERE `+who+` AND ($2='' OR g.status=$2) AND (g.created_at < $3 OR (g.created_at = $3 AND g.id < $4::uuid)) ORDER BY g.created_at DESC, g.id DESC LIMIT $5`,
		userID, in.Status, before, beforeID, in.Limit+1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	next := ""
	var lastAt time.Time
	var lastID string
	now := s.now()
	for rows.Next() {
		var raw, id string
		var at time.Time
		if err := rows.Scan(&raw, &at, &id); err != nil {
			return nil, err
		}
		if len(items) == in.Limit {
			next = base64.RawURLEncoding.EncodeToString([]byte(lastAt.UTC().Format(time.RFC3339Nano) + "|" + lastID))
			break
		}
		m := map[string]any{}
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			return nil, err
		}
		items = append(items, clientGift(m, userID, langs, lang, now))
		lastAt, lastID = at, id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var pending int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM economy_friend_gift WHERE recipient_id=$1 AND status='pending'`, userID).Scan(&pending)
	return map[string]any{"lang": lang, "items": items, "cursor": next, "pendingCount": pending}, nil
}

// rpcFriendGiftOptions lists what the player can give one friend right now,
// marking what the friend already has and why something cannot be sent.
func (s *service) rpcFriendGiftOptions(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		FriendID string `json:"friendId"`
		Lang     string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	st, err := loadFriendGiftSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	friend := strings.TrimSpace(in.FriendID)
	if _, err := uuid.FromString(friend); err != nil || friend == userID {
		return nil, badRequest("RECIPIENT_NOT_FOUND", "friendId", "pick a friend")
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	rel, err := s.giftRelation(ctx, s.db, st, userID, friend)
	if err != nil {
		return nil, err
	}
	if rel.RecipientGone {
		return nil, notFound("RECIPIENT_NOT_FOUND")
	}
	now := s.now()
	reason := ""
	if !st.Enabled {
		reason = "FRIEND_GIFTS_OFF"
	} else {
		reason = relationBlock(st, rel, now)
	}
	lim, err := s.giftLimitsFor(ctx, s.db, st, userID, friend)
	if err != nil {
		return nil, err
	}
	day := s.giftDay(ctx)
	use, err := giftUsageFor(ctx, s.db, userID, friend, day)
	if err != nil {
		return nil, err
	}
	if reason == "" && !lim.CanSend {
		reason = "GIFT_SENDING_BLOCKED"
	}
	if reason == "" {
		reason = limitBlock(lim, use)
	}
	var fname, fuser string
	_ = s.db.QueryRowContext(ctx, `SELECT username, COALESCE(display_name,'') FROM users WHERE id=$1`, friend).Scan(&fuser, &fname)

	// inventory items
	items := []map[string]any{}
	if st.InventoryEnabled {
		defs, err := loadItems(ctx, s.db, `WHERE t.is_active AND EXISTS (SELECT 1 FROM economy_user_item u WHERE u.item_id=t.id AND u.user_id=$1)`, userID)
		if err != nil {
			return nil, err
		}
		mine, err := ownedItems(ctx, s.db, userID)
		if err != nil {
			return nil, err
		}
		theirs, err := s.ownedItemKeys(ctx, friend)
		if err != nil {
			return nil, err
		}
		pending := map[string]bool{}
		rows, err := s.db.QueryContext(ctx, `SELECT item_id::text FROM economy_friend_gift WHERE recipient_id=$1 AND status='pending' AND item_id IS NOT NULL`, friend)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id string
			if rows.Scan(&id) == nil {
				pending[id] = true
			}
		}
		rows.Close()
		lc := newLevelCache(s.db, friend)
		for _, d := range defs {
			o := mine[d.ID]
			why := ""
			switch {
			case d.IsStarter:
				why = "STARTER_ITEM"
			case !st.itemGiftable(d.Giftable, d.TypeID):
				why = "ITEM_NOT_GIFTABLE"
			case o.Equipped && st.Equipped != "unequip":
				why = "ITEM_EQUIPPED"
			case o.Source == "gift" && st.Regift == "block":
				why = "REGIFT_BLOCKED"
			case theirs[d.Key]:
				why = "RECIPIENT_OWNS_ITEM"
			case pending[d.ID]:
				why = "GIFT_ALREADY_PENDING"
			case st.ItemRequirements != "ignore" && d.needs() != "" && !theirs[d.needs()]:
				why = "RECIPIENT_REQUIRES_ITEM"
			case st.ItemRequirements != "ignore" && d.unlock() > 0:
				if lv, err := lc.level(ctx, d.table()); err == nil && lv < d.unlock() {
					why = "RECIPIENT_LEVEL_TOO_LOW"
				}
			}
			if d.IsStarter {
				continue
			}
			v := itemPublic(d, langs, lang)
			v["upgradeLevel"], v["equipped"] = o.UpgradeLevel, o.Equipped
			v["recipientOwns"] = theirs[d.Key]
			v["giftable"] = why == ""
			v["reason"] = why
			items = append(items, v)
		}
	}

	// currencies
	currencies := []map[string]any{}
	if st.CurrencyEnabled && len(st.Currencies) > 0 {
		wallet, err := s.userWallet(ctx, userID)
		if err != nil {
			return nil, err
		}
		for _, c := range st.Currencies {
			var name, icon string
			if err := s.db.QueryRowContext(ctx, `SELECT name::text, icon_url FROM economy_currency WHERE key=$1`, c.CurrencyKey).Scan(&name, &icon); err != nil {
				continue
			}
			currencies = append(currencies, map[string]any{"key": c.CurrencyKey, "name": langs.Resolve(localized.Parse(name), lang), "iconUrl": icon, "min": c.Min, "max": c.Max, "balance": wallet[c.CurrencyKey]})
		}
	}

	// shop products
	products := []map[string]any{}
	if st.ShopEnabled {
		methods, err := s.activeMethods(ctx)
		if err != nil {
			return nil, err
		}
		tags, err := s.userTagIDs(ctx, userID)
		if err != nil {
			return nil, err
		}
		rows, err := s.db.QueryContext(ctx, `SELECT `+productJSON+`::text FROM economy_product t WHERE t.is_active AND NOT t.is_subscription ORDER BY t.sort, t.created_at`)
		if err != nil {
			return nil, err
		}
		var list []productView
		for rows.Next() {
			var raw string
			if err := rows.Scan(&raw); err != nil {
				rows.Close()
				return nil, err
			}
			var p productView
			if err := json.Unmarshal([]byte(raw), &p); err != nil {
				rows.Close()
				return nil, err
			}
			list = append(list, p)
		}
		rows.Close()
		for _, p := range list {
			if !p.visibleTo(tags) || !st.productGiftable(p.Giftable) {
				continue
			}
			avail := []string{}
			sku := map[string]string{}
			for _, m := range p.PaymentMethods {
				if _, ok := methods[m]; ok && st.methodAllowed(m) {
					avail = append(avail, m)
					if isStore(m) {
						sku[m] = p.SKU[m]
					}
				}
			}
			if len(avail) == 0 {
				continue
			}
			prize, err := loadPrize(ctx, s.db, p.PrizeID)
			if err != nil {
				return nil, err
			}
			h, err := s.productHoldingFor(ctx, s.db, friend, p, prize)
			if err != nil {
				return nil, err
			}
			why := productOwnBlock(st, h)
			if why == "PRODUCT_NOT_GIFTABLE" {
				continue
			}
			if why == "" && (st.ProductRules == "recipient" || st.ProductRules == "both") {
				if err := s.checkLevelRules(ctx, friend, p, "", "RECIPIENT_LEVEL_TOO_LOW", "RECIPIENT_LEVEL_TOO_HIGH"); err != nil {
					if app, ok := asApp(err); ok {
						why = app.Code
					} else {
						return nil, err
					}
				}
			}
			if why == "" && (st.ProductRules == "sender" || st.ProductRules == "both") {
				if err := s.checkLevelRules(ctx, userID, p, "", "LEVEL_TOO_LOW", "LEVEL_TOO_HIGH"); err != nil {
					if app, ok := asApp(err); ok {
						why = app.Code
					} else {
						return nil, err
					}
				}
			}
			var final any
			if p.PriceToman != nil {
				final = discounted(*p.PriceToman, p.DiscountPercent)
			}
			products = append(products, map[string]any{
				"id": p.ID, "key": p.Key, "name": langs.Resolve(p.Name, lang), "description": langs.Resolve(p.Description, lang), "iconUrl": p.IconURL, "coverUrl": p.CoverURL,
				"priceToman": p.PriceToman, "finalToman": final, "coinPrice": p.CoinPrice, "discountPercent": p.DiscountPercent, "paymentMethods": avail, "sku": sku,
				"items": prizeItemKeys(prize.Items), "recipientOwns": why == "RECIPIENT_OWNS_PRODUCT", "giftable": why == "", "reason": why,
			})
		}
	}
	sender := map[string]any{"used": use.Sender, "limit": lim.Sender, "remaining": remaining(lim.Sender, use.Sender), "bonus": lim.SenderBonus}
	recv := map[string]any{"used": use.Recipient, "limit": lim.Recipient, "remaining": remaining(lim.Recipient, use.Recipient), "bonus": lim.RecipientBonus}
	pair := map[string]any{"used": use.Pair, "limit": lim.Pair, "remaining": remaining(lim.Pair, use.Pair)}
	return map[string]any{
		"lang": lang, "friend": map[string]any{"id": friend, "username": fuser, "displayName": fname, "isFriend": rel.Mine == 0 && rel.Theirs == 0},
		"canSend": reason == "", "reason": reason, "day": day,
		"limits":   map[string]any{"sender": sender, "recipient": recv, "pair": pair},
		"delivery": st.Delivery, "expiryHours": st.ExpiryHours, "messageEnabled": st.MessageEnabled, "messageMaxLength": st.MessageMaxLength,
		"items": items, "currencies": currencies, "products": products, "discountCodes": st.DiscountCodes,
	}, nil
}
