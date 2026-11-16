package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/chatfilter"
	"github.com/heroiclabs/nakama/v3/internal/inapp"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

type friendGift struct {
	ID          string
	SenderID    string
	RecipientID string
	Source      string
	Kind        string
	ItemID      sql.NullString
	ItemKey     string
	CurrencyKey string
	Amount      int64
	ProductID   sql.NullString
	PrizeID     sql.NullString
	PaymentID   sql.NullString
	Method      string
	ItemState   string
	Message     string
	Delivery    string
	Status      string
	Outcome     string
	RequestID   string
	ExpiresAt   sql.NullTime
	CreatedAt   time.Time
}

const friendGiftCols = `id::text, sender_id::text, recipient_id::text, source, kind, item_id::text, item_key, currency_key, amount, product_id::text, prize_id::text, payment_id::text,
	method, item_state::text, message, delivery, status, outcome, request_id, expires_at, created_at`

func scanFriendGift(row interface{ Scan(...any) error }) (friendGift, error) {
	var g friendGift
	err := row.Scan(&g.ID, &g.SenderID, &g.RecipientID, &g.Source, &g.Kind, &g.ItemID, &g.ItemKey, &g.CurrencyKey, &g.Amount, &g.ProductID, &g.PrizeID, &g.PaymentID,
		&g.Method, &g.ItemState, &g.Message, &g.Delivery, &g.Status, &g.Outcome, &g.RequestID, &g.ExpiresAt, &g.CreatedAt)
	return g, err
}

func loadFriendGift(ctx context.Context, q queryer, id string, lock bool) (friendGift, error) {
	if _, err := uuid.FromString(strings.TrimSpace(id)); err != nil {
		return friendGift{}, notFound("GIFT_NOT_FOUND")
	}
	query := `SELECT ` + friendGiftCols + ` FROM economy_friend_gift WHERE id=$1`
	if lock {
		query += ` FOR UPDATE`
	}
	g, err := scanFriendGift(q.QueryRowContext(ctx, query, strings.TrimSpace(id)))
	if errors.Is(err, sql.ErrNoRows) {
		return g, notFound("GIFT_NOT_FOUND")
	}
	return g, err
}

// giftDay is the project day (event time zone) a gift counts against.
func (s *service) giftDay(ctx context.Context) string {
	zone := "Asia/Tehran"
	_ = s.db.QueryRowContext(ctx, `SELECT timezone FROM economy_event_settings WHERE id=1`).Scan(&zone)
	return s.now().In(loadZone(zone)).Format("2006-01-02")
}

// lockGiftPair serialises gift sends that touch the same sender or recipient
// so limits and "already owns" checks cannot race.
func lockGiftPair(ctx context.Context, tx *sql.Tx, a, b string) error {
	ids := []string{a, b}
	if ids[1] < ids[0] {
		ids[0], ids[1] = ids[1], ids[0]
	}
	for _, id := range ids {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('economy.friend_gift:' || $1))`, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) giftRelation(ctx context.Context, q queryer, st friendGiftSettings, sender, recipient string) (giftRelation, error) {
	r := giftRelation{Mine: -1, Theirs: -1}
	var rCreated, sCreated, disabled time.Time
	var since sql.NullTime
	err := q.QueryRowContext(ctx, `SELECT r.create_time, r.disable_time, s.create_time, COALESCE(mine.state,-1), COALESCE(theirs.state,-1), CASE WHEN mine.state=0 THEN mine.update_time END
		FROM users r JOIN users s ON s.id=$1
		LEFT JOIN user_edge mine ON mine.source_id=$1 AND mine.destination_id=r.id
		LEFT JOIN user_edge theirs ON theirs.source_id=r.id AND theirs.destination_id=$1
		WHERE r.id=$2 AND r.id <> '00000000-0000-0000-0000-000000000000'::uuid`, sender, recipient).Scan(&rCreated, &disabled, &sCreated, &r.Mine, &r.Theirs, &since)
	if errors.Is(err, sql.ErrNoRows) {
		r.RecipientGone = true
		return r, nil
	}
	if err != nil {
		return r, err
	}
	if disabled.After(time.Unix(0, 0)) {
		r.RecipientGone = true
	}
	now := s.now()
	r.SenderAge, r.RecipientAge = now.Sub(sCreated), now.Sub(rCreated)
	if since.Valid {
		t := since.Time
		r.FriendsSince = &t
	}
	if st.MinMutualFriends > 0 {
		if err := q.QueryRowContext(ctx, `SELECT count(*) FROM user_edge a JOIN user_edge b ON b.destination_id=a.destination_id AND b.source_id=$2 AND b.state=0
			WHERE a.source_id=$1 AND a.state=0`, sender, recipient).Scan(&r.Mutual); err != nil {
			return r, err
		}
	}
	if st.MinSenderLevel > 0 {
		lv, err := newLevelCache(q, sender).level(ctx, st.LevelTableID)
		if err != nil {
			return r, err
		}
		r.SenderLevel = lv
	}
	return r, nil
}

func (s *service) giftRules(ctx context.Context, q rowsQueryer, userID string) ([]giftRule, error) {
	rows, err := q.QueryContext(ctx, `SELECT r.id::text, r.name, r.per_sender_daily, r.per_recipient_daily, r.per_pair_daily, r.can_send FROM economy_friend_gift_rule r
		WHERE r.is_active AND EXISTS (SELECT 1 FROM base_user_tag ut WHERE ut.user_id=$1::uuid AND ut.tag_id::text = ANY(r.tag_ids) AND (ut.expires_at IS NULL OR ut.expires_at > now()))
		ORDER BY r.priority DESC, r.created_at, r.id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []giftRule{}
	for rows.Next() {
		var r giftRule
		if err := rows.Scan(&r.ID, &r.Name, &r.Sender, &r.Recipient, &r.Pair, &r.CanSend); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *service) giftMetaBonus(ctx context.Context, q queryer, userID, key string) (int, error) {
	if key == "" {
		return 0, nil
	}
	var v sql.NullFloat64
	err := q.QueryRowContext(ctx, `SELECT COALESCE(um.value, m.default_value)::float8 FROM economy_meta m LEFT JOIN economy_user_meta um ON um.meta_id=m.id AND um.user_id=$1::uuid WHERE m.key=$2`, userID, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !v.Valid) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if v.Float64 <= 0 || math.IsNaN(v.Float64) {
		return 0, nil
	}
	return int(math.Min(math.Floor(v.Float64), 100000)), nil
}

// giftLimitsFor works out the limits of one sender-recipient pair: the
// sender's tag rules give the sender and pair limits and whether they may
// send at all, the recipient's tag rules give the receive limit, and the
// bonus metas add purchasable extra quota.
func (s *service) giftLimitsFor(ctx context.Context, q interface {
	queryer
	rowsQueryer
}, st friendGiftSettings, sender, recipient string) (giftLimits, error) {
	base := giftLimits{Sender: st.PerSenderDaily, Recipient: st.PerRecipientDaily, Pair: st.PerPairDaily}
	sr, err := s.giftRules(ctx, q, sender)
	if err != nil {
		return base, err
	}
	out := applyGiftRules(base, sr, st.MatchPolicy)
	if recipient != "" {
		rr, err := s.giftRules(ctx, q, recipient)
		if err != nil {
			return out, err
		}
		out.Recipient = applyGiftRules(base, rr, st.MatchPolicy).Recipient
	}
	if out.Sender > 0 {
		if out.SenderBonus, err = s.giftMetaBonus(ctx, q, sender, st.SendBonusMeta); err != nil {
			return out, err
		}
		out.Sender += out.SenderBonus
	}
	if out.Recipient > 0 && recipient != "" {
		if out.RecipientBonus, err = s.giftMetaBonus(ctx, q, recipient, st.ReceiveBonusMeta); err != nil {
			return out, err
		}
		out.Recipient += out.RecipientBonus
	}
	return out, nil
}

// countedStatuses are the gift states that use up a daily limit: everything
// except gifts that never left the sender.
const giftCounted = `status NOT IN ('failed','cancelled')`

func giftUsageFor(ctx context.Context, q queryer, sender, recipient, day string) (giftUsage, error) {
	var u giftUsage
	err := q.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE sender_id=$1), count(*) FILTER (WHERE recipient_id=$2), count(*) FILTER (WHERE sender_id=$1 AND recipient_id=$2)
		FROM economy_friend_gift WHERE day=$3::date AND (sender_id=$1 OR recipient_id=$2) AND `+giftCounted, sender, recipient, day).Scan(&u.Sender, &u.Recipient, &u.Pair)
	return u, err
}

func (s *service) chatWordFilter(ctx context.Context, st friendGiftSettings) (*chatfilter.Filter, error) {
	if st.MessageFilter == "off" {
		return nil, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, word, action, match FROM base_chat_word WHERE is_active ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rules := []chatfilter.Rule{}
	for rows.Next() {
		var r chatfilter.Rule
		if err := rows.Scan(&r.ID, &r.Word, &r.Action, &r.Match); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return chatfilter.Compile(rules), rows.Err()
}

type giftSendInput struct {
	FriendID    string `json:"friendId"`
	ItemKey     string `json:"itemKey"`
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
	ProductID   string `json:"productId"`
	Method      string `json:"method"`
	GiftCode    string `json:"giftCode"`
	Message     string `json:"message"`
	RequestID   string `json:"requestId"`
	Lang        string `json:"lang"`
}

func (in giftSendInput) kind() string {
	n := 0
	k := ""
	if strings.TrimSpace(in.ItemKey) != "" {
		n, k = n+1, "item"
	}
	if strings.TrimSpace(in.CurrencyKey) != "" {
		n, k = n+1, "currency"
	}
	if strings.TrimSpace(in.ProductID) != "" {
		n, k = n+1, "product"
	}
	if n != 1 {
		return ""
	}
	return k
}

func (in giftSendInput) fingerprint() string {
	return strings.Join([]string{in.FriendID, in.kind(), strings.TrimSpace(in.ItemKey), strings.TrimSpace(in.CurrencyKey), jsonInt(in.Amount), strings.TrimSpace(in.ProductID), in.Method}, "|")
}

type itemHolding struct {
	Source       string `json:"source"`
	SourceRef    string `json:"sourceRef"`
	UpgradeLevel int    `json:"upgradeLevel"`
	Equipped     bool   `json:"equipped"`
	Slot         string `json:"slot"`
}

// checkGiftItem checks one inventory item for a gift and returns its
// definition and the sender's copy. The sender's row stays locked.
func (s *service) checkGiftItem(ctx context.Context, tx *sql.Tx, st friendGiftSettings, sender, recipient, key string) (itemDef, itemHolding, error) {
	var h itemHolding
	if !st.InventoryEnabled {
		return itemDef{}, h, conflict("INVENTORY_GIFTS_OFF", "gifts from the inventory are turned off")
	}
	d, err := loadItemByKey(ctx, tx, strings.TrimSpace(key))
	if err != nil {
		return d, h, err
	}
	if !d.IsActive {
		return d, h, notFound("ITEM_NOT_FOUND")
	}
	if d.IsStarter {
		return d, h, conflict("STARTER_ITEM", "starter items belong to every player")
	}
	if !st.itemGiftable(d.Giftable, d.TypeID) {
		return d, h, conflict("ITEM_NOT_GIFTABLE", "this item cannot be given as a gift")
	}
	err = tx.QueryRowContext(ctx, `SELECT source, source_ref, upgrade_level, equipped, slot FROM economy_user_item WHERE user_id=$1 AND item_id=$2 FOR UPDATE`, sender, d.ID).
		Scan(&h.Source, &h.SourceRef, &h.UpgradeLevel, &h.Equipped, &h.Slot)
	if errors.Is(err, sql.ErrNoRows) {
		return d, h, conflict("ITEM_NOT_OWNED", "you do not own this item")
	}
	if err != nil {
		return d, h, err
	}
	if h.Equipped && st.Equipped != "unequip" {
		return d, h, conflict("ITEM_EQUIPPED", "take this item off before giving it away")
	}
	if h.Source == "gift" && st.Regift == "block" {
		return d, h, conflict("REGIFT_BLOCKED", "an item you got as a gift cannot be given again")
	}
	if err := s.checkRecipientItem(ctx, tx, st, recipient, d); err != nil {
		return d, h, err
	}
	return d, h, nil
}

// checkRecipientItem refuses an item the recipient owns, is already getting,
// or (when the admin checks requirements) cannot hold yet.
func (s *service) checkRecipientItem(ctx context.Context, q queryer, st friendGiftSettings, recipient string, d itemDef) error {
	has, err := owns(ctx, q, recipient, d.Key)
	if err != nil {
		return err
	}
	if has {
		return conflict("RECIPIENT_OWNS_ITEM", "your friend already has this item")
	}
	var pending bool
	if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_friend_gift WHERE recipient_id=$1 AND item_id=$2 AND status='pending')`, recipient, d.ID).Scan(&pending); err != nil {
		return err
	}
	if pending {
		return conflict("GIFT_ALREADY_PENDING", "your friend is already getting this item")
	}
	if st.ItemRequirements != "ignore" {
		if need := d.needs(); need != "" {
			ok, err := owns(ctx, q, recipient, need)
			if err != nil {
				return err
			}
			if !ok {
				return conflict("RECIPIENT_REQUIRES_ITEM", "your friend must own "+need+" first")
			}
		}
		if d.unlock() > 0 {
			lv, err := newLevelCache(q, recipient).level(ctx, d.table())
			if err != nil {
				return err
			}
			if lv < d.unlock() {
				return conflict("RECIPIENT_LEVEL_TOO_LOW", "your friend's level is too low for this item")
			}
		}
	}
	return nil
}

// productHoldingFor reads what the recipient already has of a product.
func (s *service) productHoldingFor(ctx context.Context, q queryer, recipient string, p productView, prize prizeRow) (productHolding, error) {
	var h productHolding
	keys := prizeItemKeys(prize.Items)
	h.ItemKeys = len(keys)
	if len(keys) > 0 {
		if err := q.QueryRowContext(ctx, `SELECT count(*) FROM economy_user_item u JOIN economy_item i ON i.id=u.item_id WHERE u.user_id=$1 AND i.key = ANY($2)`, recipient, keys).Scan(&h.OwnedKeys); err != nil {
			return h, err
		}
	}
	if !prize.Repeatable {
		claimed, err := prizeClaimedByUser(ctx, q, recipient, prize.ID)
		if err != nil {
			return h, err
		}
		h.PrizeClaimed = claimed
	}
	if p.IsOneTime {
		if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_payment WHERE user_id=$1 AND product_id=$2 AND status='success' AND friend_gift_id IS NULL AND COALESCE(outcome,'') <> 'duplicate')
			OR EXISTS (SELECT 1 FROM economy_friend_gift WHERE recipient_id=$1 AND product_id=$2 AND status IN ('delivered','pending','paying'))`, recipient, p.ID).Scan(&h.OneTimeHad); err != nil {
			return h, err
		}
	}
	if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_friend_gift WHERE recipient_id=$1 AND product_id=$2 AND status IN ('pending','paying'))`, recipient, p.ID).Scan(&h.PendingSame); err != nil {
		return h, err
	}
	return h, nil
}

// checkGiftProduct checks a shop product and payment method for a gift.
func (s *service) checkGiftProduct(ctx context.Context, st friendGiftSettings, sender, recipient, productID, method string) (productView, prizeRow, paymentMethod, error) {
	var prize prizeRow
	var pm paymentMethod
	if !st.ShopEnabled {
		return productView{}, prize, pm, conflict("SHOP_GIFTS_OFF", "buying gifts from the shop is turned off")
	}
	product, err := s.loadProduct(ctx, s.db, strings.TrimSpace(productID))
	if err != nil {
		return product, prize, pm, err
	}
	if !product.IsActive {
		return product, prize, pm, notFound("PRODUCT_NOT_FOUND")
	}
	senderTags, err := s.userTagIDs(ctx, sender)
	if err != nil {
		return product, prize, pm, err
	}
	if !product.visibleTo(senderTags) {
		return product, prize, pm, notFound("PRODUCT_NOT_FOUND")
	}
	if product.IsSubscription || !st.productGiftable(product.Giftable) {
		return product, prize, pm, conflict("PRODUCT_NOT_GIFTABLE", "this product cannot be bought as a gift")
	}
	if !validMethod(method) || !product.allows(method) || !st.methodAllowed(method) {
		return product, prize, pm, badRequest("METHOD_NOT_ALLOWED", "method", "this gift cannot be bought with this method")
	}
	if pm, err = s.loadMethod(ctx, s.db, method); err != nil {
		return product, prize, pm, err
	}
	if !pm.usable() {
		return product, prize, pm, fail(http.StatusServiceUnavailable, "METHOD_UNAVAILABLE", "method", "this payment method is not available")
	}
	if prize, err = loadPrize(ctx, s.db, product.PrizeID); err != nil {
		return product, prize, pm, err
	}
	if st.ProductRules == "sender" || st.ProductRules == "both" {
		if err := s.checkLevelRules(ctx, sender, product, "", "LEVEL_TOO_LOW", "LEVEL_TOO_HIGH"); err != nil {
			return product, prize, pm, err
		}
	}
	if st.ProductRules == "recipient" || st.ProductRules == "both" {
		if err := s.checkLevelRules(ctx, recipient, product, "RECIPIENT_", "RECIPIENT_LEVEL_TOO_LOW", "RECIPIENT_LEVEL_TOO_HIGH"); err != nil {
			return product, prize, pm, err
		}
	}
	return product, prize, pm, nil
}

func (s *service) checkLevelRules(ctx context.Context, userID string, p productView, prefix, low, high string) error {
	if p.MinLevel == 0 && p.MaxLevel == nil {
		return nil
	}
	table := ""
	if p.LevelTableID != nil {
		table = *p.LevelTableID
	}
	lv, err := newLevelCache(s.db, userID).level(ctx, table)
	if err != nil {
		return err
	}
	if lv < p.MinLevel {
		return conflict(low, "level too low for this product")
	}
	if p.MaxLevel != nil && lv > *p.MaxLevel {
		return conflict(high, "level too high for this product")
	}
	return nil
}

// sendFriendGift validates and starts one gift. Inventory items move inside
// the send transaction; currency moves right after it through the wallet;
// shop gifts start a purchase paid by the sender.
func (s *service) sendFriendGift(ctx context.Context, sender string, in giftSendInput) (map[string]any, error) {
	st, err := loadFriendGiftSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if !st.Enabled {
		return nil, conflict("FRIEND_GIFTS_OFF", "friend gifts are turned off")
	}
	in.FriendID = strings.TrimSpace(in.FriendID)
	if _, err := uuid.FromString(in.FriendID); err != nil {
		return nil, badRequest("RECIPIENT_NOT_FOUND", "friendId", "pick a friend")
	}
	if in.FriendID == sender {
		return nil, badRequest("GIFT_SELF", "friendId", "you cannot send a gift to yourself")
	}
	kind := in.kind()
	if kind == "" {
		return nil, badRequest("GIFT_SOURCE_REQUIRED", "", "send exactly one of itemKey, currencyKey or productId")
	}
	in.RequestID = strings.TrimSpace(in.RequestID)
	if len(in.RequestID) > 64 {
		return nil, badRequest("INVALID_REQUEST_ID", "requestId", "at most 64 characters")
	}
	if in.RequestID != "" {
		if out, done, err := s.replayGift(ctx, sender, in); done || err != nil {
			return out, err
		}
	}
	filter, err := s.chatWordFilter(ctx, st)
	if err != nil {
		return nil, err
	}
	msg, flagged, err := cleanGiftMessage(st, in.Message, filter)
	if err != nil {
		return nil, err
	}
	if err := s.ensureDefaults(ctx, sender); err != nil {
		return nil, err
	}
	var exists bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id=$1 AND id <> '00000000-0000-0000-0000-000000000000'::uuid)`, in.FriendID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, notFound("RECIPIENT_NOT_FOUND")
	}
	if err := s.ensureStarters(ctx, in.FriendID); err != nil {
		return nil, err
	}
	var product productView
	var prize prizeRow
	var pm paymentMethod
	if kind == "product" {
		if product, prize, pm, err = s.checkGiftProduct(ctx, st, sender, in.FriendID, in.ProductID, in.Method); err != nil {
			return nil, err
		}
		if strings.TrimSpace(in.GiftCode) != "" && (!st.DiscountCodes || !isGateway(in.Method)) {
			return nil, badRequest("GIFT_NOT_APPLICABLE", "giftCode", "discount codes do not work here")
		}
	}
	now := s.now()
	day := s.giftDay(ctx)
	id := uuid.Must(uuid.NewV4()).String()
	g := friendGift{ID: id, SenderID: sender, RecipientID: in.FriendID, Kind: kind, Source: "inventory", Delivery: st.Delivery, Message: msg, RequestID: in.RequestID, CreatedAt: now}
	var expires any
	if st.Delivery == "accept" {
		t := now.Add(time.Duration(st.ExpiryHours) * time.Hour)
		expires = t
		g.ExpiresAt = sql.NullTime{Time: t, Valid: true}
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if err := lockGiftPair(ctx, tx, sender, in.FriendID); err != nil {
			return err
		}
		rel, err := s.giftRelation(ctx, tx, st, sender, in.FriendID)
		if err != nil {
			return err
		}
		if code := relationBlock(st, rel, now); code != "" {
			if code == "RECIPIENT_NOT_FOUND" {
				return notFound(code)
			}
			return conflict(code, "this friend cannot get a gift from you")
		}
		lim, err := s.giftLimitsFor(ctx, tx, st, sender, in.FriendID)
		if err != nil {
			return err
		}
		if !lim.CanSend {
			return conflict("GIFT_SENDING_BLOCKED", "you cannot send gifts")
		}
		use, err := giftUsageFor(ctx, tx, sender, in.FriendID, day)
		if err != nil {
			return err
		}
		if code := limitBlock(lim, use); code != "" {
			return fail(http.StatusTooManyRequests, code, "", "daily gift limit reached")
		}
		status, outcome := "sending", ""
		switch kind {
		case "item":
			d, hold, err := s.checkGiftItem(ctx, tx, st, sender, in.FriendID, in.ItemKey)
			if err != nil {
				return err
			}
			g.ItemID = sql.NullString{String: d.ID, Valid: true}
			g.ItemKey = d.Key
			raw, _ := json.Marshal(hold)
			g.ItemState = string(raw)
			if _, err := tx.ExecContext(ctx, `DELETE FROM economy_user_item WHERE user_id=$1 AND item_id=$2`, sender, d.ID); err != nil {
				return err
			}
			if st.Delivery == "instant" {
				if err := s.giveGiftItem(ctx, tx, st, g, hold); err != nil {
					return err
				}
				status, outcome = "delivered", "delivered"
			} else {
				status = "pending"
			}
		case "currency":
			c, ok := st.currency(strings.TrimSpace(in.CurrencyKey))
			if !ok {
				return conflict("CURRENCY_NOT_GIFTABLE", "this currency cannot be given as a gift")
			}
			if in.Amount < 1 || (c.Min > 0 && in.Amount < c.Min) || (c.Max > 0 && in.Amount > c.Max) {
				return badRequest("INVALID_AMOUNT", "amount", "the amount is outside the allowed range")
			}
			g.CurrencyKey, g.Amount = c.CurrencyKey, in.Amount
		case "product":
			h, err := s.productHoldingFor(ctx, tx, in.FriendID, product, prize)
			if err != nil {
				return err
			}
			if code := productOwnBlock(st, h); code != "" {
				return conflict(code, "your friend cannot get this product")
			}
			g.Source = "shop"
			g.ProductID = sql.NullString{String: product.ID, Valid: true}
			g.PrizeID = sql.NullString{String: prize.ID, Valid: true}
			g.Method = in.Method
			status = "paying"
		}
		g.Status = status
		if g.ItemState == "" {
			g.ItemState = "{}"
		}
		var delivered any
		if status == "delivered" {
			delivered = now
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO economy_friend_gift (id, sender_id, recipient_id, source, kind, item_id, item_key, currency_key, amount, product_id, prize_id, method, item_state,
			message, message_flagged, delivery, status, outcome, request_id, day, expires_at, delivered_at, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20::date,$21,$22,$23,$23)`,
			g.ID, sender, in.FriendID, g.Source, kind, nullStr(g.ItemID), g.ItemKey, g.CurrencyKey, max(g.Amount, 1), nullStr(g.ProductID), nullStr(g.PrizeID), g.Method, g.ItemState,
			msg, flagged, st.Delivery, status, outcome, in.RequestID, day, expires, delivered, now)
		return err
	})
	if c, unique := isUniqueViolation(err); unique && c == "economy_friend_gift_request_idx" {
		if out, done, rerr := s.replayGift(ctx, sender, in); done || rerr != nil {
			return out, rerr
		}
	}
	if err != nil {
		return nil, err
	}
	switch kind {
	case "item":
		if g.Status == "delivered" {
			s.giftDelivered(ctx, st, g)
		} else {
			s.giftArrived(ctx, st, g)
		}
		s.giftSentEvent(ctx, g)
	case "currency":
		if err := s.moveGiftCurrency(ctx, st, &g); err != nil {
			return nil, err
		}
	case "product":
		pay, err := s.startGiftPurchase(ctx, st, g, product, pm, in)
		if err != nil {
			return nil, err
		}
		view, verr := s.giftClientView(ctx, sender, g.ID, in.Lang)
		if verr != nil {
			return nil, verr
		}
		out := map[string]any{"gift": view}
		for k, v := range pay {
			out[k] = v
		}
		return out, nil
	}
	view, err := s.giftClientView(ctx, sender, g.ID, in.Lang)
	if err != nil {
		return nil, err
	}
	return map[string]any{"gift": view}, nil
}

func nullStr(v sql.NullString) any {
	if !v.Valid {
		return nil
	}
	return v.String
}

func (s *service) replayGift(ctx context.Context, sender string, in giftSendInput) (map[string]any, bool, error) {
	var id, friend, kind, itemKey, cur, method string
	var amount int64
	var product sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT id::text, recipient_id::text, kind, item_key, currency_key, amount, product_id::text, method FROM economy_friend_gift WHERE sender_id=$1 AND request_id=$2`, sender, in.RequestID).
		Scan(&id, &friend, &kind, &itemKey, &cur, &amount, &product, &method)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, true, err
	}
	same := friend == in.FriendID && kind == in.kind()
	switch kind {
	case "item":
		same = same && itemKey == strings.TrimSpace(in.ItemKey)
	case "currency":
		same = same && cur == strings.TrimSpace(in.CurrencyKey) && amount == in.Amount
	case "product":
		same = same && product.String == strings.TrimSpace(in.ProductID) && method == in.Method
	}
	if !same {
		return nil, true, conflict("REQUEST_ID_REUSED", "this requestId was used for another gift")
	}
	view, err := s.giftClientView(ctx, sender, id, in.Lang)
	if err != nil {
		return nil, true, err
	}
	return map[string]any{"gift": view, "replayed": true}, true, nil
}

// giveGiftItem puts a gifted item in the recipient's inventory.
func (s *service) giveGiftItem(ctx context.Context, tx *sql.Tx, st friendGiftSettings, g friendGift, hold itemHolding) error {
	level := hold.UpgradeLevel
	if st.UpgradeLevel == "reset" {
		level = 0
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO economy_user_item (user_id, item_id, source, source_ref, upgrade_level) VALUES ($1,$2,'gift',$3,$4) ON CONFLICT (user_id, item_id) DO NOTHING`,
		g.RecipientID, g.ItemID.String, g.ID, level)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return conflict("RECIPIENT_OWNS_ITEM", "your friend already has this item")
	}
	return nil
}

// restoreGiftItem gives a held or taken-back item to the sender as they had it.
func restoreGiftItem(ctx context.Context, tx *sql.Tx, g friendGift) error {
	var hold itemHolding
	_ = json.Unmarshal([]byte(g.ItemState), &hold)
	if hold.Source == "" {
		hold.Source = "gift"
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO economy_user_item (user_id, item_id, source, source_ref, upgrade_level) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (user_id, item_id) DO NOTHING`,
		g.SenderID, g.ItemID.String, hold.Source, hold.SourceRef, hold.UpgradeLevel)
	return err
}

func giftGrantID(g friendGift, part string) string {
	return "friend_gift:" + g.ID + ":" + part
}

// moveGiftCurrency moves gifted currency through the wallet ledger: from the
// sender to the recipient, or into escrow for a gift that must be accepted.
func (s *service) moveGiftCurrency(ctx context.Context, st friendGiftSettings, g *friendGift) error {
	md := map[string]any{"module": "economy", "reason": "friend_gift", "giftId": g.ID, "grantId": giftGrantID(*g, "send")}
	updates := []*runtime.WalletUpdate{{UserID: g.SenderID, Changeset: map[string]int64{g.CurrencyKey: -g.Amount}, Metadata: md}}
	next := "pending"
	if st.Delivery == "instant" {
		in := map[string]any{"module": "economy", "reason": "friend_gift", "giftId": g.ID, "grantId": giftGrantID(*g, "in")}
		updates = append(updates, &runtime.WalletUpdate{UserID: g.RecipientID, Changeset: map[string]int64{g.CurrencyKey: g.Amount}, Metadata: in})
		next = "delivered"
	}
	if _, _, err := s.nk.MultiUpdate(ctx, nil, nil, nil, updates, true); err != nil {
		code := "WALLET_FAILED"
		var neg *runtime.WalletNegativeError
		if errors.As(err, &neg) {
			code = "INSUFFICIENT_BALANCE"
		}
		_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET status='failed', error=$2, updated_at=now() WHERE id=$1 AND status='sending'`, g.ID, code)
		if code == "INSUFFICIENT_BALANCE" {
			return fail(http.StatusPaymentRequired, code, "amount", "not enough "+g.CurrencyKey)
		}
		return err
	}
	outcome := ""
	if next == "delivered" {
		outcome = "delivered"
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET status=$2, outcome=$3, delivered_at=CASE WHEN $2='delivered' THEN now() END, updated_at=now() WHERE id=$1 AND status='sending'`, g.ID, next, outcome); err != nil {
		return err
	}
	g.Status = next
	if next == "delivered" {
		s.giftDelivered(ctx, st, *g)
	} else {
		s.giftArrived(ctx, st, *g)
	}
	s.giftSentEvent(ctx, *g)
	return nil
}

// startGiftPurchase starts the sender's payment for a shop gift.
func (s *service) startGiftPurchase(ctx context.Context, st friendGiftSettings, g friendGift, product productView, pm paymentMethod, in giftSendInput) (map[string]any, error) {
	product.giftID = g.ID
	var out any
	var err error
	switch {
	case in.Method == "in_app":
		out, err = s.startGiftInApp(ctx, g, product)
	case isStore(in.Method):
		out, err = s.startStore(ctx, g.SenderID, product, pm)
	default:
		code := ""
		if st.DiscountCodes {
			code = in.GiftCode
		}
		out, err = s.startGateway(ctx, g.SenderID, product, pm, code)
	}
	if err != nil {
		code := err.Error()
		if app, ok := asApp(err); ok {
			code = app.Code
		}
		_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET status='failed', error=$2, updated_at=now() WHERE id=$1 AND status='paying'`, g.ID, truncate(code, 500))
		return nil, err
	}
	res, _ := out.(map[string]any)
	if pid, ok := res["paymentId"].(string); ok {
		_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET payment_id=$2, updated_at=now() WHERE id=$1 AND payment_id IS NULL`, g.ID, pid)
	}
	delete(res, "rewards")
	return res, nil
}

// startGiftInApp charges the sender's coins and hands the gift on.
func (s *service) startGiftInApp(ctx context.Context, g friendGift, product productView) (any, error) {
	if product.CoinPrice == nil || product.CoinPrice.CurrencyKey == "" {
		return nil, badRequest("METHOD_NOT_ALLOWED", "method", "product has no coin price")
	}
	id := uuid.Must(uuid.NewV4()).String()
	if _, err := s.db.ExecContext(ctx, `INSERT INTO economy_payment (id, user_id, product_id, method, status, coin_currency, coin_amount, friend_gift_id) VALUES ($1,$2,$3,'in_app','pending',$4,$5,$6)`,
		id, g.SenderID, product.ID, product.CoinPrice.CurrencyKey, product.CoinPrice.Amount, g.ID); err != nil {
		return nil, err
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET payment_id=$2, updated_at=now() WHERE id=$1`, g.ID, id)
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		return s.walletChange(ctx, tx, g.SenderID, id, map[string]int64{product.CoinPrice.CurrencyKey: -product.CoinPrice.Amount}, map[string]any{"reason": "friend_gift", "paymentId": id, "giftId": g.ID})
	})
	if err != nil {
		code := err.Error()
		if app, ok := asApp(err); ok {
			code = app.Code
		}
		_, _ = s.closePayment(ctx, id, "failed", code)
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE economy_payment SET status='success', outcome='paid', verified_at=now(), updated_at=now() WHERE id=$1 AND status='pending'`, id); err != nil {
		return nil, err
	}
	s.recordServerEvent(ctx, g.SenderID, "purchase_completed", float64(product.CoinPrice.Amount), map[string]any{"paymentId": id, "method": "in_app", "productKey": product.Key, "currency": product.CoinPrice.CurrencyKey, "friendGiftId": g.ID})
	if _, err := s.grantPayment(ctx, id); err != nil {
		s.logger.Warn("Economy friend gift %s paid but not handed on yet: %v", g.ID, err)
	}
	return map[string]any{"paymentId": id, "status": "success"}, nil
}

func (s *service) paymentGiftID(ctx context.Context, paymentID string) string {
	var id sql.NullString
	_ = s.db.QueryRowContext(ctx, `SELECT friend_gift_id::text FROM economy_payment WHERE id=$1`, paymentID).Scan(&id)
	return id.String
}

// friendGiftPaid runs when a gift purchase is paid: the product goes to the
// recipient now, or waits for them to accept it.
func (s *service) friendGiftPaid(ctx context.Context, p payment, product productView, giftID string) ([]prizeItem, error) {
	st, err := loadFriendGiftSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	var g friendGift
	first := false
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if g, err = loadFriendGift(ctx, tx, giftID, true); err != nil {
			return err
		}
		if g.Status == "paying" || (g.Status == "failed" && g.Outcome == "") {
			first = true
			next := "delivered"
			var expires any
			if g.Delivery == "accept" {
				next = "pending"
				t := s.now().Add(time.Duration(st.ExpiryHours) * time.Hour)
				expires = t
				g.ExpiresAt = sql.NullTime{Time: t, Valid: true}
			}
			g.Status = next
			if _, err := tx.ExecContext(ctx, `UPDATE economy_friend_gift SET status=$2, payment_id=$3, error=NULL, expires_at=COALESCE($4, expires_at), updated_at=now() WHERE id=$1`, g.ID, next, p.ID, expires); err != nil {
				return err
			}
		}
		if g.Status != "delivered" || g.Outcome != "" {
			_, err := tx.ExecContext(ctx, `UPDATE economy_payment SET reward_granted=true, reward_error=NULL, updated_at=now() WHERE id=$1`, p.ID)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if first {
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: p.UserID, Kind: activity.KindPurchase, Action: "completed", RefID: p.ID, Actor: "player", Data: map[string]any{"method": p.Method, "productKey": product.Key, "toman": p.FinalToman, "friendGiftId": g.ID}})
		s.giftSentEvent(ctx, g)
	}
	if g.Status == "pending" {
		if first {
			s.giftArrived(ctx, st, g)
		}
		return []prizeItem{}, nil
	}
	if g.Status == "delivered" && g.Outcome == "" {
		if err := s.deliverGiftProduct(ctx, st, g, &p.ID); err != nil {
			return nil, err
		}
	}
	return []prizeItem{}, nil
}

// deliverGiftProduct grants the product's prize to the recipient. A recipient
// who got the contents meanwhile makes the gift go back per the settings.
func (s *service) deliverGiftProduct(ctx context.Context, st friendGiftSettings, g friendGift, paymentID *string) error {
	check := true
	if g.ProductID.Valid {
		if p, err := s.loadProduct(ctx, s.db, g.ProductID.String); err == nil {
			check = p.CheckContents
		}
	}
	_, err := s.grant(ctx, grantRequest{UserID: g.RecipientID, PrizeID: g.PrizeID.String, SourceType: "friend_gift", SourceID: g.ID, Reason: "friend_gift", CheckItems: check,
		After: func(ctx context.Context, tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, `UPDATE economy_friend_gift SET outcome='delivered', delivered_at=COALESCE(delivered_at, now()), error=NULL, updated_at=now() WHERE id=$1`, g.ID); err != nil {
				return err
			}
			if paymentID != nil {
				_, err := tx.ExecContext(ctx, `UPDATE economy_payment SET reward_granted=true, reward_error=NULL, updated_at=now() WHERE id=$1`, *paymentID)
				return err
			}
			return nil
		}})
	if err != nil {
		code := err.Error()
		if app, ok := asApp(err); ok {
			code = app.Code
		}
		if code == "PRIZE_ALREADY_CLAIMED" || code == "PRIZE_ALREADY_OWNED" {
			_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET status='declined', error='RECIPIENT_OWNS_PRODUCT', responded_at=now(), updated_at=now() WHERE id=$1 AND status='delivered' AND outcome=''`, g.ID)
			g.Status = "declined"
			if paymentID != nil {
				_, _ = s.db.ExecContext(ctx, `UPDATE economy_payment SET reward_granted=true, reward_error=NULL, updated_at=now() WHERE id=$1`, *paymentID)
			}
			s.settleGiftReturn(ctx, st, g)
			return conflict("RECIPIENT_OWNS_PRODUCT", "your friend already has this product")
		}
		_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET error=$2, updated_at=now() WHERE id=$1`, g.ID, truncate(code, 500))
		return err
	}
	s.giftDelivered(ctx, st, g)
	return nil
}

// respondGift handles accept, decline (recipient) and cancel (sender).
func (s *service) respondGift(ctx context.Context, userID, giftID, action string) (friendGift, error) {
	st, err := loadFriendGiftSettings(ctx, s.db)
	if err != nil {
		return friendGift{}, err
	}
	var g friendGift
	var ownErr error
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if g, err = loadFriendGift(ctx, tx, giftID, true); err != nil {
			return err
		}
		mine := (action == "cancel" && g.SenderID == userID) || (action != "cancel" && g.RecipientID == userID)
		if !mine {
			return notFound("GIFT_NOT_FOUND")
		}
		if action == "cancel" && !st.SenderCancel {
			return conflict("GIFT_CANCEL_OFF", "sent gifts cannot be taken back")
		}
		if g.Status != "pending" {
			return conflict("GIFT_NOT_PENDING", "this gift is not waiting for an answer")
		}
		if g.ExpiresAt.Valid && !g.ExpiresAt.Time.After(s.now()) {
			return conflict("GIFT_EXPIRED", "this gift has expired")
		}
		if action != "accept" {
			next := map[string]string{"decline": "declined", "cancel": "cancelled"}[action]
			g.Status = next
			if g.Kind == "item" {
				if err := restoreGiftItem(ctx, tx, g); err != nil {
					return err
				}
				g.Outcome = "returned"
			}
			_, err := tx.ExecContext(ctx, `UPDATE economy_friend_gift SET status=$2, outcome=$3, responded_at=now(), updated_at=now() WHERE id=$1`, g.ID, next, g.Outcome)
			return err
		}
		ownErr, err = s.acceptInTx(ctx, tx, st, &g)
		return err
	})
	if err != nil {
		return g, err
	}
	s.afterResponse(ctx, st, g, action)
	if ownErr != nil {
		return g, ownErr
	}
	return g, nil
}

// acceptInTx delivers a pending gift. An item the recipient got meanwhile
// sends the gift back; the returned error is reported after commit.
func (s *service) acceptInTx(ctx context.Context, tx *sql.Tx, st friendGiftSettings, g *friendGift) (error, error) {
	if g.Kind == "item" {
		var hold itemHolding
		_ = json.Unmarshal([]byte(g.ItemState), &hold)
		if err := s.giveGiftItem(ctx, tx, st, *g, hold); err != nil {
			if app, ok := asApp(err); ok && app.Code == "RECIPIENT_OWNS_ITEM" {
				if err := restoreGiftItem(ctx, tx, *g); err != nil {
					return nil, err
				}
				g.Status, g.Outcome = "declined", "returned"
				_, err := tx.ExecContext(ctx, `UPDATE economy_friend_gift SET status='declined', outcome='returned', error='RECIPIENT_OWNS_ITEM', responded_at=now(), updated_at=now() WHERE id=$1`, g.ID)
				return conflict("RECIPIENT_OWNS_ITEM", "you already have this item; it went back to the sender"), err
			}
			return nil, err
		}
		g.Status, g.Outcome = "delivered", "delivered"
		_, err := tx.ExecContext(ctx, `UPDATE economy_friend_gift SET status='delivered', outcome='delivered', delivered_at=now(), responded_at=now(), updated_at=now() WHERE id=$1`, g.ID)
		return nil, err
	}
	g.Status, g.Outcome = "delivered", ""
	_, err := tx.ExecContext(ctx, `UPDATE economy_friend_gift SET status='delivered', outcome='', responded_at=now(), updated_at=now() WHERE id=$1`, g.ID)
	return nil, err
}

// afterResponse finishes the wallet or prize side and tells both players.
func (s *service) afterResponse(ctx context.Context, st friendGiftSettings, g friendGift, action string) {
	switch g.Status {
	case "delivered":
		if g.Outcome == "" {
			if err := s.settleDelivery(ctx, st, g); err != nil {
				s.logger.Warn("Economy friend gift %s accepted but not delivered yet: %v", g.ID, err)
				return
			}
		} else {
			s.giftDelivered(ctx, st, g)
		}
		s.notifySender(ctx, st, g, "accepted")
	case "declined", "cancelled", "expired":
		if g.Outcome == "" {
			s.settleGiftReturn(ctx, st, g)
		} else {
			s.giftReturnedNotice(ctx, st, g, g.Outcome)
		}
	}
}

// settleDelivery hands an accepted (or auto-accepted) currency or product gift
// to the recipient. It is idempotent and retried by the sweep.
func (s *service) settleDelivery(ctx context.Context, st friendGiftSettings, g friendGift) error {
	switch g.Kind {
	case "currency":
		err := s.inTx(ctx, func(tx *sql.Tx) error {
			return s.walletChange(ctx, tx, g.RecipientID, giftGrantID(g, "in"), map[string]int64{g.CurrencyKey: g.Amount}, map[string]any{"reason": "friend_gift", "giftId": g.ID})
		})
		if err != nil {
			_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET error=$2, updated_at=now() WHERE id=$1`, g.ID, truncate(err.Error(), 500))
			return err
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET outcome='delivered', delivered_at=COALESCE(delivered_at, now()), error=NULL, updated_at=now() WHERE id=$1 AND outcome=''`, g.ID); err != nil {
			return err
		}
		s.giftDelivered(ctx, st, g)
		return nil
	case "product":
		return s.deliverGiftProduct(ctx, st, g, nil)
	}
	return nil
}

// settleGiftReturn gives back a declined, expired, cancelled or refused
// currency or product gift per the settings. Idempotent; retried by the sweep.
func (s *service) settleGiftReturn(ctx context.Context, st friendGiftSettings, g friendGift) {
	outcome, err := s.giveBack(ctx, st, g, returnAction(st, g.Kind, g.Method))
	if err != nil {
		s.logger.Warn("Economy friend gift %s could not go back yet: %v", g.ID, err)
		_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET error=$2, updated_at=now() WHERE id=$1`, g.ID, truncate(err.Error(), 500))
		return
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET outcome=$2, updated_at=now() WHERE id=$1 AND outcome=''`, g.ID, outcome); err != nil {
		return
	}
	s.giftReturnedNotice(ctx, st, g, outcome)
}

// giveBack performs one return action and reports what really happened: a
// product the sender cannot receive turns into a refund (coins) or a refund
// to do by hand (money).
func (s *service) giveBack(ctx context.Context, st friendGiftSettings, g friendGift, action string) (string, error) {
	switch g.Kind {
	case "item":
		return "returned", s.inTx(ctx, func(tx *sql.Tx) error { return restoreGiftItem(ctx, tx, g) })
	case "currency":
		return "returned", s.inTx(ctx, func(tx *sql.Tx) error {
			return s.walletChange(ctx, tx, g.SenderID, giftGrantID(g, "back"), map[string]int64{g.CurrencyKey: g.Amount}, map[string]any{"reason": "friend_gift_return", "giftId": g.ID})
		})
	}
	switch action {
	case "refunded":
		var cur sql.NullString
		var amount sql.NullInt64
		if g.PaymentID.Valid {
			_ = s.db.QueryRowContext(ctx, `SELECT coin_currency, coin_amount FROM economy_payment WHERE id=$1 AND method='in_app' AND status='success'`, g.PaymentID.String).Scan(&cur, &amount)
		}
		if !cur.Valid || !amount.Valid || amount.Int64 <= 0 {
			return "refund_due", nil
		}
		return "refunded", s.inTx(ctx, func(tx *sql.Tx) error {
			return s.walletChange(ctx, tx, g.SenderID, giftGrantID(g, "refund"), map[string]int64{cur.String: amount.Int64}, map[string]any{"reason": "friend_gift_refund", "giftId": g.ID})
		})
	case "sender_got":
		_, err := s.grant(ctx, grantRequest{UserID: g.SenderID, PrizeID: g.PrizeID.String, SourceType: "friend_gift", SourceID: g.ID + ":sender", Reason: "friend_gift_return"})
		if err != nil {
			if app, ok := asApp(err); ok && (app.Code == "PRIZE_ALREADY_CLAIMED" || app.Code == "PRIZE_ALREADY_OWNED") {
				if g.Method == "in_app" {
					return s.giveBack(ctx, st, g, "refunded")
				}
				return "refund_due", nil
			}
			return "", err
		}
		return "sender_got", nil
	case "refund_due":
		return "refund_due", nil
	}
	return "kept", nil
}

// expireFriendGifts ends pending gifts past their time, finishes deliveries
// and returns that did not complete, and settles currency sends cut off
// between the ledger and the gift row.
func (s *service) expireFriendGifts(ctx context.Context, n int) {
	st, err := loadFriendGiftSettings(ctx, s.db)
	if err != nil {
		return
	}
	for i := 0; i < n; i++ {
		var g friendGift
		claimed := false
		err := s.inTx(ctx, func(tx *sql.Tx) error {
			var err error
			g, err = scanFriendGift(tx.QueryRowContext(ctx, `SELECT `+friendGiftCols+` FROM economy_friend_gift WHERE status='pending' AND expires_at <= $1 ORDER BY expires_at LIMIT 1 FOR UPDATE SKIP LOCKED`, s.now()))
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			if err != nil {
				return err
			}
			claimed = true
			if st.OnExpiry == "accept" {
				_, err := s.acceptInTx(ctx, tx, st, &g)
				return err
			}
			g.Status = "expired"
			if g.Kind == "item" {
				if err := restoreGiftItem(ctx, tx, g); err != nil {
					return err
				}
				g.Outcome = "returned"
			}
			_, err = tx.ExecContext(ctx, `UPDATE economy_friend_gift SET status='expired', outcome=$2, responded_at=now(), updated_at=now() WHERE id=$1`, g.ID, g.Outcome)
			return err
		})
		if err != nil {
			s.logger.Warn("Economy friend gift expiry failed: %v", err)
			return
		}
		if !claimed {
			break
		}
		s.afterResponse(ctx, st, g, "expire")
	}
	s.retryFriendGifts(ctx, st, n)
}

func (s *service) retryFriendGifts(ctx context.Context, st friendGiftSettings, n int) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+friendGiftCols+` FROM economy_friend_gift
		WHERE ((outcome='' AND status IN ('delivered','declined','expired','cancelled') AND kind <> 'item') OR status='sending') AND updated_at < $1
		ORDER BY updated_at LIMIT $2`, s.now().Add(-time.Minute), n)
	if err != nil {
		return
	}
	var list []friendGift
	for rows.Next() {
		g, err := scanFriendGift(rows)
		if err != nil {
			break
		}
		list = append(list, g)
	}
	rows.Close()
	for _, g := range list {
		_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET updated_at=now() WHERE id=$1`, g.ID)
		switch g.Status {
		case "sending":
			s.settleSending(ctx, st, g)
		case "delivered":
			if g.Kind == "product" && g.PaymentID.Valid && g.Delivery == "instant" {
				if _, err := s.grantPayment(ctx, g.PaymentID.String); err != nil {
					s.logger.Warn("Economy friend gift %s retry failed: %v", g.ID, err)
				}
			} else if err := s.settleDelivery(ctx, st, g); err != nil {
				s.logger.Warn("Economy friend gift %s retry failed: %v", g.ID, err)
			}
		default:
			s.settleGiftReturn(ctx, st, g)
		}
	}
}

// settleSending finishes a currency send whose wallet step may or may not
// have run.
func (s *service) settleSending(ctx context.Context, st friendGiftSettings, g friendGift) {
	var applied bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM wallet_ledger WHERE user_id=$1 AND metadata->>'grantId'=$2)`, g.SenderID, giftGrantID(g, "send")).Scan(&applied); err != nil {
		return
	}
	if !applied {
		_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET status='failed', error='WALLET_NOT_APPLIED', updated_at=now() WHERE id=$1 AND status='sending'`, g.ID)
		return
	}
	next, outcome := "pending", ""
	if g.Delivery == "instant" {
		next, outcome = "delivered", "delivered"
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE economy_friend_gift SET status=$2, outcome=$3, delivered_at=CASE WHEN $2='delivered' THEN now() END, updated_at=now() WHERE id=$1 AND status='sending'`, g.ID, next, outcome)
}

// --- events, activity and notifications ---

func (s *service) giftWhat(ctx context.Context, g friendGift) (string, map[string]any) {
	md := map[string]any{"giftId": g.ID, "kind": g.Kind, "source": g.Source, "senderId": g.SenderID, "recipientId": g.RecipientID}
	key := ""
	switch g.Kind {
	case "item":
		key = g.ItemKey
		md["itemKey"] = g.ItemKey
	case "currency":
		key = g.CurrencyKey
		md["currency"], md["amount"] = g.CurrencyKey, g.Amount
	case "product":
		var pk string
		_ = s.db.QueryRowContext(ctx, `SELECT key FROM economy_product WHERE id=$1`, g.ProductID.String).Scan(&pk)
		key = pk
		md["productKey"], md["method"] = pk, g.Method
	}
	md["key"] = key
	return key, md
}

func (s *service) giftSentEvent(ctx context.Context, g friendGift) {
	key, md := s.giftWhat(ctx, g)
	s.recordServerEvent(ctx, g.SenderID, giftSentEvent, float64(max(g.Amount, 1)), md)
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: g.SenderID, Kind: activity.KindPrize, Action: "gift_sent", RefID: g.ID, Actor: "player", Data: map[string]any{"what": key, "kind": g.Kind, "to": g.RecipientID}})
}

func (s *service) giftDelivered(ctx context.Context, st friendGiftSettings, g friendGift) {
	key, md := s.giftWhat(ctx, g)
	s.recordServerEvent(ctx, g.RecipientID, giftReceivedEvent, float64(max(g.Amount, 1)), md)
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: g.RecipientID, Kind: activity.KindPrize, Action: "gift_received", RefID: g.ID, Actor: "player", Data: map[string]any{"what": key, "kind": g.Kind, "from": g.SenderID}})
	if g.Delivery == "instant" {
		s.giftArrived(ctx, st, g)
	}
}

var giftTexts = map[string]localized.Text{
	"arrived.title":  {"fa": "یک هدیه از {name}", "en": "A gift from {name}", "ar": "هدية من {name}"},
	"arrived.body":   {"fa": "{name} برایت «{what}» فرستاد.", "en": "{name} sent you “{what}”.", "ar": "أرسل لك {name} «{what}»."},
	"pending.body":   {"fa": "{name} برایت «{what}» فرستاد. برای گرفتنش، هدیه را قبول کن.", "en": "{name} sent you “{what}”. Accept it to receive it.", "ar": "أرسل لك {name} «{what}». اقبلها لتحصل عليها."},
	"returned.title": {"fa": "هدیه‌ات برگشت", "en": "Your gift came back", "ar": "عادت هديتك"},
	"returned.body":  {"fa": "هدیهٔ «{what}» به {name} نرسید و به تو برگشت.", "en": "Your gift “{what}” to {name} was not delivered and came back to you.", "ar": "لم تصل هديتك «{what}» إلى {name} وعادت إليك."},
	"refunded.body":  {"fa": "هدیهٔ «{what}» به {name} نرسید و پولش به حسابت برگشت.", "en": "Your gift “{what}” to {name} was not delivered; you got your coins back.", "ar": "لم تصل هديتك «{what}» إلى {name} واستعدت عملاتك."},
	"open.label":     {"fa": "دیدن هدیه", "en": "Open gift", "ar": "عرض الهدية"},
}

func (s *service) giftNames(ctx context.Context, g friendGift, other string) (map[string]string, string) {
	var name string
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(NULLIF(display_name,''), username) FROM users WHERE id=$1`, other).Scan(&name)
	var raw string
	switch g.Kind {
	case "item":
		_ = s.db.QueryRowContext(ctx, `SELECT name::text FROM economy_item WHERE id=$1`, g.ItemID.String).Scan(&raw)
	case "currency":
		_ = s.db.QueryRowContext(ctx, `SELECT name::text FROM economy_currency WHERE key=$1`, g.CurrencyKey).Scan(&raw)
	case "product":
		_ = s.db.QueryRowContext(ctx, `SELECT name::text FROM economy_product WHERE id=$1`, g.ProductID.String).Scan(&raw)
	}
	names := localized.Parse(raw)
	out := map[string]string{}
	for _, lang := range []string{"fa", "en", "ar"} {
		v := names[lang]
		if v == "" {
			v = names["fa"]
		}
		if v == "" {
			v = names["en"]
		}
		if v == "" {
			v = g.ItemKey + g.CurrencyKey
		}
		if g.Kind == "currency" {
			v = jsonInt(g.Amount) + " " + v
		}
		out[lang] = v
	}
	return out, name
}

func giftFill(t localized.Text, name string, what map[string]string) map[string]string {
	out := map[string]string{}
	for lang, v := range t {
		w := what[lang]
		if w == "" {
			w = what["en"]
		}
		out[lang] = strings.NewReplacer("{name}", name, "{what}", w).Replace(v)
	}
	return out
}

func (s *service) giftInbox(ctx context.Context, st friendGiftSettings, to string, title, body map[string]string, g friendGift) {
	if st.NotifyChannel == "none" || !inapp.Available() {
		return
	}
	m := inapp.Message{UserIDs: []string{to}, Title: title, Body: body, Channel: st.NotifyChannel,
		Action: &inapp.Action{Type: "goto", Value: "gifts/" + g.ID, Label: giftTexts["open.label"]}, CreatedBy: "economy:friend_gift"}
	if _, err := inapp.Send(ctx, m); err != nil && m.Channel != "in_app" {
		m.Channel = "in_app"
		_, err = inapp.Send(ctx, m)
		if err != nil {
			s.logger.Warn("Economy friend gift %s notice not sent: %v", g.ID, err)
		}
	} else if err != nil {
		s.logger.Warn("Economy friend gift %s notice not sent: %v", g.ID, err)
	}
}

func (s *service) giftSocket(ctx context.Context, st friendGiftSettings, to, event string, g friendGift) {
	if !st.SocketNotify {
		return
	}
	content := map[string]any{"type": "friend_gift", "event": event, "giftId": g.ID, "status": g.Status, "kind": g.Kind, "senderId": g.SenderID, "recipientId": g.RecipientID}
	if err := s.nk.NotificationSend(ctx, to, "friend_gift", content, giftNotifyCode, "", false); err != nil {
		s.logger.Debug("Economy friend gift socket notice failed: %v", err)
	}
}

// giftArrived tells the recipient a gift is there (delivered or waiting).
func (s *service) giftArrived(ctx context.Context, st friendGiftSettings, g friendGift) {
	what, name := s.giftNames(ctx, g, g.SenderID)
	body := giftTexts["arrived.body"]
	event := "received"
	if g.Status == "pending" {
		body = giftTexts["pending.body"]
		event = "pending"
	}
	if g.Message != "" {
		body = appendText(body, "\n«"+g.Message+"»")
	}
	s.giftInbox(ctx, st, g.RecipientID, giftFill(giftTexts["arrived.title"], name, what), giftFill(body, name, what), g)
	s.giftSocket(ctx, st, g.RecipientID, event, g)
}

func appendText(t localized.Text, suffix string) localized.Text {
	out := localized.Text{}
	for k, v := range t {
		out[k] = v + suffix
	}
	return out
}

func (s *service) notifySender(ctx context.Context, st friendGiftSettings, g friendGift, event string) {
	if st.NotifySender {
		s.giftSocket(ctx, st, g.SenderID, event, g)
	}
}

func (s *service) giftReturnedNotice(ctx context.Context, st friendGiftSettings, g friendGift, outcome string) {
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: g.SenderID, Kind: activity.KindPrize, Action: "gift_returned", RefID: g.ID, Actor: "system", Data: map[string]any{"status": g.Status, "outcome": outcome}})
	if !st.NotifySender {
		return
	}
	s.giftSocket(ctx, st, g.SenderID, g.Status, g)
	if outcome != "returned" && outcome != "refunded" && outcome != "sender_got" {
		return
	}
	what, name := s.giftNames(ctx, g, g.RecipientID)
	body := giftTexts["returned.body"]
	if outcome == "refunded" {
		body = giftTexts["refunded.body"]
	}
	s.giftInbox(ctx, st, g.SenderID, giftFill(giftTexts["returned.title"], name, what), giftFill(body, name, what), g)
}
