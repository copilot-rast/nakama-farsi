package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func (s *service) clientRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_catalog":         s.rpcCatalog,
		"economy_balances":        s.rpcBalances,
		"economy_purchase_start":  s.rpcPurchaseStart,
		"economy_purchase_submit": s.rpcPurchaseSubmit,
		"economy_purchase_review": s.rpcPurchaseReview,
		"economy_purchase_status": s.rpcPurchaseStatus,
		"economy_store_config":    s.rpcStoreConfig,
		"economy_gift_redeem":     s.rpcGiftRedeem,
		"economy_payments_mine":   s.rpcPaymentsMine,
	}
}

func (s *service) rpcCatalog(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		CategoryKey string `json:"categoryKey"`
		Lang        string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	methods, err := s.activeMethods(ctx)
	if err != nil {
		return nil, err
	}
	currencies := []map[string]any{}
	rows, err := s.db.QueryContext(ctx, `SELECT t.key, t.name::text, t.type, t.icon_url, `+taxonomyKeysSQL("currency")+` FROM economy_currency t ORDER BY t.sort, t.key`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var key, name, typ, icon, catKey, tagsRaw string
		if err := rows.Scan(&key, &name, &typ, &icon, &catKey, &tagsRaw); err != nil {
			rows.Close()
			return nil, err
		}
		currencies = append(currencies, map[string]any{"key": key, "name": langs.Resolve(localized.Parse(name), lang), "type": typ, "iconUrl": icon, "categoryKey": catKey, "tags": tagKeyList(tagsRaw)})
	}
	rows.Close()

	categories := []map[string]any{}
	categoryKeys := map[string]string{}
	rows, err = s.db.QueryContext(ctx, `SELECT id::text, key, name::text, description::text, icon_url, sort FROM base_category WHERE key_type='product' ORDER BY sort, key`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, key, name, desc, icon string
		var sort int
		if err := rows.Scan(&id, &key, &name, &desc, &icon, &sort); err != nil {
			rows.Close()
			return nil, err
		}
		categoryKeys[id] = key
		categories = append(categories, map[string]any{"key": key, "name": langs.Resolve(localized.Parse(name), lang), "description": langs.Resolve(localized.Parse(desc), lang), "iconUrl": icon, "sort": sort})
	}
	rows.Close()

	tagKeys := map[string]string{}
	rows, err = s.db.QueryContext(ctx, `SELECT id::text, key FROM base_tag WHERE key_type IN ('product','general')`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, key string
		if err := rows.Scan(&id, &key); err != nil {
			rows.Close()
			return nil, err
		}
		tagKeys[id] = key
	}
	rows.Close()

	userTags, err := s.userTagIDs(ctx, userID)
	if err != nil {
		return nil, err
	}
	counts, err := s.purchaseCounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	ownedKeys, err := s.ownedItemKeys(ctx, userID)
	if err != nil {
		return nil, err
	}
	lc := newLevelCache(s.db, userID)
	disc := s.userDiscounts(ctx, s.db, userID)
	prizes := map[string]prizeRow{}
	products := []map[string]any{}
	rows, err = s.db.QueryContext(ctx, `SELECT `+productJSON+`::text FROM economy_product t WHERE t.is_active ORDER BY t.sort, t.created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var p productView
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			return nil, err
		}
		if !p.visibleTo(userTags) {
			continue
		}
		categoryKey := ""
		if p.CategoryID != nil {
			categoryKey = categoryKeys[*p.CategoryID]
		}
		if in.CategoryKey != "" && categoryKey != in.CategoryKey {
			continue
		}
		available := []string{}
		sku := map[string]string{}
		for _, m := range p.PaymentMethods {
			if _, ok := methods[m]; ok {
				available = append(available, m)
				if isStore(m) {
					sku[m] = p.SKU[m]
				}
			}
		}
		if len(available) == 0 {
			continue
		}
		prize, ok := prizes[p.PrizeID]
		if !ok {
			if prize, err = loadPrize(ctx, s.db, p.PrizeID); err != nil {
				return nil, err
			}
			prizes[p.PrizeID] = prize
		}
		st, err := s.productStateFor(ctx, p, prize, lc, counts, ownedKeys)
		if err != nil {
			return nil, err
		}
		if st.OverMax || (p.IsOneTime && st.Count > 0) {
			continue
		}
		tags := []string{}
		for _, id := range p.TagIDs {
			if k, ok := tagKeys[id]; ok {
				tags = append(tags, k)
			}
		}
		ownPercent := p.DiscountPercent
		storeFinal := any(nil)
		if p.PriceToman != nil {
			storeFinal = discounted(*p.PriceToman, ownPercent)
		}
		p = p.withTagDiscount(disc)
		var final any
		if p.PriceToman != nil {
			final = discounted(*p.PriceToman, p.DiscountPercent)
			if p.TagFinalToman != nil {
				final = *p.TagFinalToman
			}
		}
		var coinOriginal any
		if p.CoinOriginal != nil {
			coinOriginal = p.CoinOriginal
		}
		products = append(products, map[string]any{
			"id": p.ID, "key": p.Key, "name": langs.Resolve(p.Name, lang), "description": langs.Resolve(p.Description, lang),
			"iconUrl": p.IconURL, "coverUrl": p.CoverURL, "priceToman": p.PriceToman, "finalToman": final, "coinPrice": p.CoinPrice,
			"discountPercent": p.DiscountPercent, "tagDiscountPercent": p.TagDiscountPercent, "coinOriginal": coinOriginal, "storeFinalToman": storeFinal,
			"paymentMethods": available, "sku": sku, "isSubscription": p.IsSubscription,
			"categoryKey": categoryKey, "tags": tags, "minLevel": p.MinLevel, "maxLevel": p.MaxLevel, "exclusive": len(p.AudienceTagIDs) > 0,
			"isOneTime": p.IsOneTime, "owned": st.Owned, "levelLocked": st.LevelLocked, "bought": st.Count > 0, "purchaseCount": st.Count,
			"requiresKey": st.Missing, "purchaseMessage": langs.Resolve(p.PurchaseMessage, lang), "items": prizeItemKeys(prize.Items),
		})
	}
	return map[string]any{"lang": lang, "currencies": currencies, "categories": categories, "products": products}, rows.Err()
}

func (s *service) userWallet(ctx context.Context, userID string) (map[string]int64, error) {
	account, err := s.nk.AccountGetId(ctx, userID)
	if err != nil {
		return nil, err
	}
	wallet := map[string]int64{}
	if account.Wallet != "" {
		if err := json.Unmarshal([]byte(account.Wallet), &wallet); err != nil {
			return nil, err
		}
	}
	return wallet, nil
}

func (s *service) userMetas(ctx context.Context, userID string, includeServer bool) (map[string]json.Number, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT m.key, COALESCE(um.value, m.default_value)::text FROM economy_meta m
		LEFT JOIN economy_user_meta um ON um.meta_id=m.id AND um.user_id=$1 WHERE $2 OR NOT m.server_only ORDER BY m.key`, userID, includeServer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]json.Number{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		if strings.Contains(value, ".") {
			value = strings.TrimRight(strings.TrimRight(value, "0"), ".")
		}
		out[key] = json.Number(value)
	}
	return out, rows.Err()
}

func (s *service) rpcBalances(ctx context.Context, userID, payload string) (any, error) {
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	s.grantPendingRewards(ctx, userID)
	wallet, err := s.userWallet(ctx, userID)
	if err != nil {
		return nil, err
	}
	metas, err := s.userMetas(ctx, userID, false)
	if err != nil {
		return nil, err
	}
	items := []map[string]any{}
	cursor := ""
	for i := 0; i < 10; i++ {
		objects, next, err := s.nk.StorageList(ctx, "", userID, "economy_items", 100, cursor)
		if err != nil {
			return nil, err
		}
		for _, o := range objects {
			var value any
			_ = json.Unmarshal([]byte(o.Value), &value)
			items = append(items, map[string]any{"key": o.Key, "value": value, "updatedAt": o.UpdateTime.AsTime()})
		}
		if next == "" {
			break
		}
		cursor = next
	}
	return map[string]any{"wallet": wallet, "metas": metas, "items": items}, nil
}

type giftRow struct {
	ID           string
	Type         string
	PrizeID      sql.NullString
	Percent      sql.NullInt64
	PerUserLimit int
	MaxTotal     sql.NullInt64
	ValidFrom    sql.NullTime
	ExpireAt     sql.NullTime
	Active       bool
}

func (s *service) lockGift(ctx context.Context, tx *sql.Tx, userID, code string) (giftRow, error) {
	var g giftRow
	code = strings.ToUpper(strings.TrimSpace(code))
	if !giftCodePattern.MatchString(code) {
		return g, notFound("GIFT_NOT_FOUND")
	}
	err := tx.QueryRowContext(ctx, `SELECT id::text, type, prize_id::text, discount_percent, per_user_limit, max_total_uses, valid_from, expire_at, is_active
		FROM economy_gift WHERE code=$1 FOR UPDATE`, code).Scan(&g.ID, &g.Type, &g.PrizeID, &g.Percent, &g.PerUserLimit, &g.MaxTotal, &g.ValidFrom, &g.ExpireAt, &g.Active)
	if errors.Is(err, sql.ErrNoRows) {
		return g, notFound("GIFT_NOT_FOUND")
	}
	if err != nil {
		return g, err
	}
	now := s.now()
	if !g.Active {
		return g, conflict("GIFT_INACTIVE", "this code is not active")
	}
	if g.ValidFrom.Valid && now.Before(g.ValidFrom.Time) {
		return g, conflict("GIFT_NOT_STARTED", "this code is not valid yet")
	}
	if g.ExpireAt.Valid && !now.Before(g.ExpireAt.Time) {
		return g, conflict("GIFT_EXPIRED", "this code has expired")
	}
	var mine, total int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE user_id=$2), count(*) FROM economy_gift_use WHERE gift_id=$1 AND status IN ('reserved','used')`, g.ID, userID).Scan(&mine, &total); err != nil {
		return g, err
	}
	if mine >= g.PerUserLimit {
		return g, conflict("GIFT_ALREADY_USED", "you have already used this code")
	}
	if g.MaxTotal.Valid && int64(total) >= g.MaxTotal.Int64 {
		return g, conflict("GIFT_EXHAUSTED", "this code has no uses left")
	}
	return g, nil
}

func (s *service) rpcGiftRedeem(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Code string `json:"code"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	useID := uuid.Must(uuid.NewV4()).String()
	var prizeID string
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		g, err := s.lockGift(ctx, tx, userID, in.Code)
		if err != nil {
			return err
		}
		if g.Type != "reward" || !g.PrizeID.Valid {
			return conflict("GIFT_NOT_REDEEMABLE", "this is a discount code; use it at checkout")
		}
		prizeID = g.PrizeID.String
		prize, err := loadPrize(ctx, tx, prizeID)
		if err != nil {
			return err
		}
		if !prize.Repeatable {
			claimed, err := prizeClaimedByUser(ctx, tx, userID, prizeID)
			if err != nil {
				return err
			}
			if claimed {
				return conflict("PRIZE_ALREADY_CLAIMED", "this prize can be received only once")
			}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_gift_use (id, gift_id, user_id, status) VALUES ($1,$2,$3,'used')`, useID, g.ID, userID); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE economy_gift SET total_used=total_used+1, updated_at=now() WHERE id=$1`, g.ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	rewards, err := s.grant(ctx, grantRequest{UserID: userID, PrizeID: prizeID, SourceType: "gift", SourceID: useID, Reason: "gift"})
	if err != nil {
		_ = s.inTx(ctx, func(tx *sql.Tx) error {
			var giftID string
			if err := tx.QueryRowContext(ctx, `UPDATE economy_gift_use SET status='released', updated_at=now() WHERE id=$1 AND status='used' RETURNING gift_id::text`, useID).Scan(&giftID); err != nil {
				return err
			}
			_, err := tx.ExecContext(ctx, `UPDATE economy_gift SET total_used=GREATEST(total_used-1,0), updated_at=now() WHERE id=$1`, giftID)
			return err
		})
		return nil, err
	}
	s.recordServerEvent(ctx, userID, "gift_redeemed", 1, map[string]any{"code": strings.ToUpper(strings.TrimSpace(in.Code))})
	activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindPrize, Action: "gift_redeemed", Actor: "player", Data: map[string]any{"code": strings.ToUpper(strings.TrimSpace(in.Code))}}))
	out := map[string]any{"rewards": rewards}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, "")
	if err != nil {
		return out, nil
	}
	var name, catKey, tagsRaw string
	if err := s.db.QueryRowContext(ctx, `SELECT t.name::text, `+taxonomyKeysSQL("gift")+` FROM economy_gift t JOIN economy_gift_use u ON u.gift_id=t.id WHERE u.id=$1`, useID).Scan(&name, &catKey, &tagsRaw); err == nil {
		out["gift"] = map[string]any{"name": langs.Resolve(localized.Parse(name), lang), "categoryKey": catKey, "tags": tagKeyList(tagsRaw), "lang": lang}
	}
	return out, nil
}

func taxonomyKeysSQL(keyType string) string {
	return `COALESCE((SELECT c.key FROM base_category c WHERE c.id=t.category_id),''), COALESCE((SELECT jsonb_agg(g.key ORDER BY g.sort, g.key) FROM base_tag g WHERE g.id::text = ANY(t.tag_ids) AND g.key_type IN ('` + keyType + `','general')),'[]'::jsonb)::text`
}

func tagKeyList(raw string) []string {
	out := []string{}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

type purchaseInput struct {
	ProductID string `json:"productId"`
	OfferID   string `json:"offerId"`
	Method    string `json:"method"`
	GiftCode  string `json:"giftCode"`
	Lang      string `json:"lang"`
}

func (s *service) rpcPurchaseStart(ctx context.Context, userID, payload string) (any, error) {
	var in purchaseInput
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	return s.purchaseStart(ctx, userID, in)
}

func (s *service) purchaseStart(ctx context.Context, userID string, in purchaseInput) (any, error) {
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	var product productView
	var err error
	allowGift := true
	if strings.TrimSpace(in.OfferID) != "" {
		var offer userOffer
		var st offerSettings
		if offer, st, err = s.offerForPurchase(ctx, userID, in.OfferID); err != nil {
			return nil, err
		}
		if in.ProductID != "" && in.ProductID != offer.Tpl.ProductID {
			return nil, badRequest("OFFER_PRODUCT_MISMATCH", "productId", "this offer sells another product")
		}
		if product, err = s.loadProduct(ctx, s.db, offer.Tpl.ProductID); err != nil {
			return nil, err
		}
		if !product.IsActive && st.InactiveProducts == "skip" {
			return nil, notFound("OFFER_NOT_FOUND")
		}
		allowed := false
		for _, m := range offerMethods(product, offer.price(), nil) {
			if m == in.Method {
				allowed = true
			}
		}
		if !validMethod(in.Method) || !allowed {
			return nil, badRequest("METHOD_NOT_ALLOWED", "method", "this offer cannot be bought with this method")
		}
		product = product.withOffer(offer)
		allowGift = st.GiftCodes
	} else {
		if product, err = s.loadProduct(ctx, s.db, in.ProductID); err != nil {
			return nil, err
		}
		if !product.IsActive {
			return nil, notFound("PRODUCT_NOT_FOUND")
		}
		if len(product.AudienceTagIDs) > 0 {
			tags, err := s.userTagIDs(ctx, userID)
			if err != nil {
				return nil, err
			}
			if !product.visibleTo(tags) {
				return nil, notFound("PRODUCT_NOT_FOUND")
			}
		}
		if !validMethod(in.Method) || !product.allows(in.Method) {
			return nil, badRequest("METHOD_NOT_ALLOWED", "method", "this product cannot be bought with this method")
		}
		if !isStore(in.Method) && product.giftID == "" {
			product = product.withTagDiscount(s.userDiscounts(ctx, s.db, userID))
		}
	}
	pm, err := s.loadMethod(ctx, s.db, in.Method)
	if err != nil {
		return nil, err
	}
	if !pm.usable() {
		return nil, fail(http.StatusServiceUnavailable, "METHOD_UNAVAILABLE", "method", "this payment method is not available")
	}
	prize, err := loadPrize(ctx, s.db, product.PrizeID)
	if err != nil {
		return nil, err
	}
	if err := s.checkProductRules(ctx, userID, product, prize); err != nil {
		return nil, err
	}
	if !prize.Repeatable {
		claimed, err := prizeClaimedByUser(ctx, s.db, userID, prize.ID)
		if err != nil {
			return nil, err
		}
		if claimed {
			return nil, conflict("PRIZE_ALREADY_CLAIMED", "you already own this")
		}
	}
	if strings.TrimSpace(in.GiftCode) != "" && (!isGateway(in.Method) || !allowGift) {
		if !allowGift {
			return nil, badRequest("GIFT_NOT_APPLICABLE", "giftCode", "discount codes do not work on offers")
		}
		return nil, badRequest("GIFT_NOT_APPLICABLE", "giftCode", "discount codes work only with online gateways")
	}
	switch {
	case in.Method == "in_app":
		return s.startInApp(ctx, userID, product, in.Lang)
	case isStore(in.Method):
		return s.startStore(ctx, userID, product, pm)
	default:
		return s.startGateway(ctx, userID, product, pm, in.GiftCode)
	}
}

func (s *service) startInApp(ctx context.Context, userID string, product productView, lang string) (any, error) {
	if product.CoinPrice == nil {
		return nil, badRequest("METHOD_NOT_ALLOWED", "method", "product has no coin price")
	}
	wallet, err := s.userWallet(ctx, userID)
	if err != nil {
		return nil, err
	}
	if wallet[product.CoinPrice.CurrencyKey] < product.CoinPrice.Amount {
		return nil, fail(http.StatusPaymentRequired, "INSUFFICIENT_BALANCE", "", "not enough "+product.CoinPrice.CurrencyKey)
	}
	id := uuid.Must(uuid.NewV4()).String()
	if product.offer != nil {
		if err := s.reserveOffer(ctx, product.offer.ID); err != nil {
			return nil, err
		}
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO economy_payment (id, user_id, product_id, method, status, coin_currency, coin_amount, discount_percent, once_key, user_offer_id)
		VALUES ($1,$2,$3,'in_app','pending',$4,$5,$7,$6,$8)`, id, userID, product.ID, product.CoinPrice.CurrencyKey, product.CoinPrice.Amount, productOnceKey(product, userID), offerPercent(product), product.offerID()); err != nil {
		s.releaseOffer(ctx, product)
		if c, unique := isUniqueViolation(err); unique && c == "economy_payment_once_idx" {
			return nil, conflict("PRODUCT_ALREADY_BOUGHT", "this product can be bought only once")
		}
		return nil, err
	}
	rewards, err := s.grant(ctx, grantRequest{
		UserID: userID, PrizeID: product.PrizeID, SourceType: "payment", SourceID: id, Reason: "purchase", CheckItems: product.CheckContents,
		Debit: map[string]int64{product.CoinPrice.CurrencyKey: product.CoinPrice.Amount},
		After: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, `UPDATE economy_payment SET status='success', outcome='paid', reward_granted=true, reward_error=NULL, verified_at=now(), updated_at=now() WHERE id=$1 AND status='pending'`, id)
			return err
		},
	})
	if err != nil {
		code := err.Error()
		if app, ok := asApp(err); ok {
			code = app.Code
		}
		charged, cerr := s.inAppCharged(ctx, userID, id)
		if cerr == nil && charged {
			_, _ = s.db.ExecContext(ctx, `UPDATE economy_payment SET status='success', outcome='paid', reward_granted=false, reward_error=$2, verified_at=now(), updated_at=now() WHERE id=$1 AND status='pending'`, id, truncate(code, 500))
			return map[string]any{"paymentId": id, "status": "success", "rewardGranted": false, "error": "REWARD_PENDING"}, nil
		}
		if cerr == nil {
			if closed, _ := s.closePayment(ctx, id, "failed", code); closed {
				s.releaseOffer(ctx, product)
			}
			_, _ = s.db.ExecContext(ctx, `DELETE FROM economy_user_prize WHERE source_type='payment' AND source_id=$1 AND NOT claimed`, id)
		} else {
			_, _ = s.db.ExecContext(ctx, `UPDATE economy_payment SET reward_error=$2, updated_at=now() WHERE id=$1`, id, truncate(code, 500))
		}
		return nil, err
	}
	s.recordServerEvent(ctx, userID, "purchase_completed", float64(product.CoinPrice.Amount), map[string]any{"paymentId": id, "method": "in_app", "productKey": product.Key, "currency": product.CoinPrice.CurrencyKey})
	s.offerBoughtEvent(ctx, id)
	activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindPurchase, Action: "completed", RefID: id, Actor: "player", Data: map[string]any{"method": "in_app", "productKey": product.Key, "coins": product.CoinPrice.Amount, "currency": product.CoinPrice.CurrencyKey}}))
	return map[string]any{"paymentId": id, "status": "success", "rewards": rewards, "purchaseMessage": s.purchaseMessage(ctx, userID, product, lang)}, nil
}

func (s *service) inAppCharged(ctx context.Context, userID, paymentID string) (bool, error) {
	var charged bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_user_prize p JOIN wallet_ledger l ON l.user_id=p.user_id AND l.metadata->>'grantId'=p.id::text
		WHERE p.source_type='payment' AND p.source_id=$2 AND p.user_id=$1)`, userID, paymentID).Scan(&charged)
	return charged, err
}

func (s *service) purchaseMessage(ctx context.Context, userID string, p productView, want string) string {
	if len(p.PurchaseMessage) == 0 {
		return ""
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, want)
	if err != nil {
		return ""
	}
	return langs.Resolve(p.PurchaseMessage, lang)
}

func (s *service) startStore(ctx context.Context, userID string, product productView, pm paymentMethod) (any, error) {
	sku := product.SKU[pm.Method]
	if sku == "" {
		return nil, badRequest("SKU_REQUIRED", "method", "product has no sku for this store")
	}
	id := uuid.Must(uuid.NewV4()).String()
	var price int64
	if product.PriceToman != nil {
		price = *product.PriceToman
	}
	final := discounted(price, product.DiscountPercent)
	if product.offer != nil && product.offer.Final != nil {
		final = *product.offer.Final
	}
	once := paymentOnceKey(product, userID)
	if err := s.freeOnceKey(ctx, once); err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO economy_payment (id, user_id, product_id, method, sandbox, status, original_toman, final_toman, amount_rial, discount_percent, user_offer_id, friend_gift_id, once_key)
		VALUES ($1,$2,$3,$4,$5,'pending',$6,$7,$8,$9,$10,$11,$12)`, id, userID, product.ID, pm.Method, pm.Config.flag("sandbox"), price, final, final*10, product.DiscountPercent, product.offerID(), product.friendGiftID(), once); err != nil {
		return nil, onceInsertError(err)
	}
	return map[string]any{"paymentId": id, "status": "store", "sku": sku, "developerPayload": id}, nil
}

func (s *service) startGateway(ctx context.Context, userID string, product productView, pm paymentMethod, giftCode string) (any, error) {
	if product.PriceToman == nil {
		return nil, badRequest("METHOD_NOT_ALLOWED", "method", "product has no Toman price")
	}
	original := *product.PriceToman
	final := discounted(original, product.DiscountPercent)
	if product.TagFinalToman != nil {
		final = *product.TagFinalToman
	}
	if product.offer != nil && product.offer.Final != nil {
		final = *product.offer.Final
	}
	percent := product.DiscountPercent
	id := uuid.Must(uuid.NewV4()).String()
	sandbox := pm.Config.flag("sandbox")
	once := paymentOnceKey(product, userID)
	if err := s.freeOnceKey(ctx, once); err != nil {
		return nil, err
	}
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var giftID any
		if strings.TrimSpace(giftCode) != "" {
			g, err := s.lockGift(ctx, tx, userID, giftCode)
			if err != nil {
				return err
			}
			if g.Type != "discount" || !g.Percent.Valid {
				return badRequest("GIFT_NOT_APPLICABLE", "giftCode", "this code is not a discount code")
			}
			final = discounted(final, int(g.Percent.Int64))
			percent = 100 - int((100-int64(percent))*(100-g.Percent.Int64)/100)
			giftID = g.ID
		}
		if final != 0 && final < 100 {
			return badRequest("PRICE_TOO_LOW", "giftCode", "the final price must be 0 or at least 100 Toman")
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_payment (id, user_id, product_id, method, sandbox, status, original_toman, final_toman, amount_rial, gift_id, discount_percent, user_offer_id, friend_gift_id, once_key)
			VALUES ($1,$2,$3,$4,$5,'pending',$6,$7,$8,$9,$10,$11,$12,$13)`, id, userID, product.ID, pm.Method, sandbox, original, final, final*10, giftID, percent, product.offerID(), product.friendGiftID(), once); err != nil {
			return onceInsertError(err)
		}
		if giftID != nil {
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_gift_use (id, gift_id, user_id, payment_id, status) VALUES ($1,$2,$3,$4,'reserved')`,
				uuid.Must(uuid.NewV4()).String(), giftID, userID, id); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	p, err := s.loadPayment(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if final == 0 {
		won, err := s.markSuccess(ctx, p, successFields{Outcome: "free"})
		if err != nil {
			return nil, err
		}
		var rewards []prizeItem
		if won {
			if rewards, err = s.grantPayment(ctx, id); err != nil {
				return nil, err
			}
		}
		return map[string]any{"paymentId": id, "status": "success", "rewards": rewards}, nil
	}
	description := s.purchaseDescription(ctx, product.Name)
	var start gatewayStart
	if pm.Method == "zarinpal" {
		start = s.zarinpalRequest(ctx, pm.Config, final*10, s.callbackURL(id), description)
	} else {
		start = s.zibalRequest(ctx, pm.Config, final*10, s.callbackURL(id), description, id)
	}
	s.logExchange(ctx, id, "request", start.Exchange)
	if !start.OK {
		_, _ = s.closePayment(ctx, id, "failed", start.Error)
		return nil, fail(http.StatusBadGateway, "GATEWAY_REQUEST_FAILED", "", start.Error)
	}
	column := "authority"
	if pm.Method == "zibal" {
		column = "track_id"
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE economy_payment SET `+column+`=$2, updated_at=now() WHERE id=$1 AND status='pending'`, id, start.Reference); err != nil {
		_, _ = s.closePayment(ctx, id, "failed", "REFERENCE_NOT_SAVED")
		return nil, err
	}
	return map[string]any{"paymentId": id, "status": "redirect", "redirectUrl": start.Redirect}, nil
}

func (s *service) rpcPurchaseSubmit(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		PaymentID string `json:"paymentId"`
		Token     string `json:"token"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	token := strings.TrimSpace(in.Token)
	if token == "" || len(token) > 1024 {
		return nil, badRequest("INVALID_TOKEN", "token", "token is required")
	}
	p, err := s.loadPayment(ctx, s.db, in.PaymentID)
	if err != nil {
		return nil, err
	}
	if p.UserID != userID {
		return nil, notFound("PAYMENT_NOT_FOUND")
	}
	if !isStore(p.Method) {
		return nil, badRequest("INVALID_METHOD", "paymentId", "not a store payment")
	}
	product, err := s.loadProduct(ctx, s.db, p.ProductID)
	if err != nil {
		return nil, err
	}
	if product, err = s.paymentProduct(ctx, p.ID, product); err != nil {
		return nil, err
	}
	return s.settleStore(ctx, p, product, token)
}

func (s *service) rpcPurchaseReview(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Method string `json:"method"`
		SKU    string `json:"sku"`
		Token  string `json:"token"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if !isStore(in.Method) {
		return nil, badRequest("INVALID_METHOD", "method", "method must be bazaar or myket")
	}
	token := strings.TrimSpace(in.Token)
	if token == "" || len(token) > 1024 || !skuPattern.MatchString(in.SKU) {
		return nil, badRequest("INVALID_TOKEN", "token", "sku and token are required")
	}
	var productID string
	viaOffer := false
	err := s.db.QueryRowContext(ctx, `SELECT id::text FROM economy_product WHERE sku->>$1 = $2 AND $1 = ANY(payment_methods) ORDER BY is_active DESC, created_at DESC LIMIT 1`, in.Method, in.SKU).Scan(&productID)
	if errors.Is(err, sql.ErrNoRows) {
		err = s.db.QueryRowContext(ctx, `SELECT t.product_id::text FROM economy_user_offer u JOIN economy_offer t ON t.id=u.offer_id WHERE u.user_id=$3 AND u.sku->>$1 = $2 ORDER BY u.created_at DESC LIMIT 1`, in.Method, in.SKU, userID).Scan(&productID)
		viaOffer = err == nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("PRODUCT_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	product, err := s.loadProduct(ctx, s.db, productID)
	if err != nil {
		return nil, err
	}
	if viaOffer {
		sku := map[string]string{}
		for k, v := range product.SKU {
			sku[k] = v
		}
		sku[in.Method] = in.SKU
		product.SKU = sku
	}
	pm, err := s.loadMethod(ctx, s.db, in.Method)
	if err != nil {
		return nil, err
	}
	if !methodConfigured(in.Method, pm.Config) {
		return nil, fail(http.StatusServiceUnavailable, "METHOD_UNAVAILABLE", "method", "this payment method is not available")
	}
	var existing string
	err = s.db.QueryRowContext(ctx, `SELECT id::text FROM economy_payment WHERE method=$1 AND (store_token=$2 OR store_token LIKE $3) ORDER BY created_at DESC LIMIT 1`, in.Method, token, likePrefix(token)+"#%").Scan(&existing)
	if err == nil {
		p, err := s.loadPayment(ctx, s.db, existing)
		if err != nil {
			return nil, err
		}
		if p.UserID != userID {
			return nil, conflict("TOKEN_ALREADY_USED", "this purchase belongs to another account")
		}
		return s.storeResult(ctx, p)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	check, err := s.storeCheck(ctx, pm, pm.Config.flag("sandbox"), product, token)
	if err != nil {
		return nil, err
	}
	var p payment
	if check.Payload != "" {
		p, err = s.loadPayment(ctx, s.db, check.Payload)
		if err != nil {
			return nil, conflict("DEVELOPER_PAYLOAD_MISMATCH", "the purchase does not belong to a known payment")
		}
	} else if product.IsSubscription {
		var id string
		err = s.db.QueryRowContext(ctx, `SELECT id::text FROM economy_payment WHERE user_id=$1 AND product_id=$2 AND method=$3 AND status IN ('pending','expired') ORDER BY created_at DESC LIMIT 1`, userID, product.ID, in.Method).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notFound("PAYMENT_NOT_FOUND")
		}
		if err != nil {
			return nil, err
		}
		if p, err = s.loadPayment(ctx, s.db, id); err != nil {
			return nil, err
		}
	} else {
		return nil, conflict("DEVELOPER_PAYLOAD_MISMATCH", "the purchase has no developer payload")
	}
	if p.UserID != userID || p.ProductID != product.ID || p.Method != in.Method {
		return nil, conflict("DEVELOPER_PAYLOAD_MISMATCH", "the purchase does not belong to this account")
	}
	s.logExchange(ctx, p.ID, "store_validate", check.Exchange)
	if p.Sandbox != pm.Config.flag("sandbox") && p.Status != "success" {
		return nil, conflict("PAYMENT_CLOSED", "this payment was started in another store mode")
	}
	if err := s.expiredAllowed(ctx, p, pm.Config.flag("sandbox")); err != nil {
		return nil, err
	}
	return s.finishStore(ctx, p, product, token, check)
}

func likePrefix(v string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(v)
}

func (s *service) storeCheck(ctx context.Context, pm paymentMethod, sandbox bool, product productView, token string) (storeCheck, error) {
	cfg := methodConfig{}
	for k, v := range pm.Config {
		cfg[k] = v
	}
	cfg["sandbox"] = sandbox
	check, err := s.validateStore(ctx, pm.Method, cfg, product.SKU[pm.Method], token, product.IsSubscription)
	if err != nil {
		return check, err
	}
	switch check.Class {
	case storeConfigRejected:
		return check, fail(http.StatusServiceUnavailable, "STORE_CONFIGURATION_REJECTED", "", "the store rejected the server credentials")
	case storeUnavailable:
		return check, fail(http.StatusServiceUnavailable, "STORE_UNAVAILABLE", "", "the store could not be reached; try again")
	case storeRejected:
		return check, conflict("PURCHASE_INVALID", "the store does not know this purchase")
	}
	if !check.Valid {
		return check, conflict("PURCHASE_INVALID", "purchase state: "+check.State)
	}
	return check, nil
}

func (s *service) settleStore(ctx context.Context, p payment, product productView, token string) (any, error) {
	if p.Status == "success" {
		if p.StoreToken.Valid && (p.StoreToken.String == token || strings.HasPrefix(p.StoreToken.String, token+"#")) {
			return s.storeResult(ctx, p)
		}
		return nil, conflict("PAYMENT_ALREADY_SETTLED", "this payment is already complete")
	}
	if p.Status == "failed" {
		return nil, conflict("PAYMENT_CLOSED", "this payment has failed; start a new one")
	}
	pm, err := s.loadMethod(ctx, s.db, p.Method)
	if err != nil {
		return nil, err
	}
	sandbox := pm.Config.flag("sandbox")
	if p.Sandbox && !sandbox {
		_, _ = s.closePayment(ctx, p.ID, "failed", "SANDBOX_OFF")
		return nil, conflict("PAYMENT_CLOSED", "this test payment was closed when test mode was turned off")
	}
	if !p.Sandbox && sandbox {
		s.notePending(ctx, p.ID, "STORE_MODE_CHANGED")
		return nil, fail(http.StatusServiceUnavailable, "STORE_MODE_CHANGED", "", "the store is in test mode; this payment is checked when test mode is off")
	}
	if err := s.expiredAllowed(ctx, p, sandbox); err != nil {
		return nil, err
	}
	if !methodConfigured(p.Method, pm.Config) && !sandbox {
		return nil, fail(http.StatusServiceUnavailable, "METHOD_UNAVAILABLE", "", "this payment method is not available")
	}
	check, err := s.storeCheck(ctx, pm, sandbox, product, token)
	s.logExchange(ctx, p.ID, "store_validate", check.Exchange)
	if err != nil {
		if app, ok := asApp(err); ok {
			switch app.Code {
			case "STORE_CONFIGURATION_REJECTED", "STORE_UNAVAILABLE":
				s.notePending(ctx, p.ID, app.Code)
			case "PURCHASE_INVALID":
				_, _ = s.closePayment(ctx, p.ID, "failed", "PURCHASE_INVALID")
			}
		}
		return nil, err
	}
	if check.Payload != "" && check.Payload != p.ID {
		return nil, conflict("DEVELOPER_PAYLOAD_MISMATCH", "the purchase does not belong to this payment")
	}
	if check.Payload == "" && !product.IsSubscription {
		return nil, conflict("DEVELOPER_PAYLOAD_MISMATCH", "the purchase has no developer payload")
	}
	return s.finishStore(ctx, p, product, token, check)
}

func (s *service) finishStore(ctx context.Context, p payment, product productView, token string, check storeCheck) (any, error) {
	key := token
	if product.IsSubscription {
		if check.ValidUntil == nil {
			return nil, conflict("TOKEN_NOT_VERIFIED", "the store did not return the subscription period")
		}
		base := token
		if check.LinkedToken != "" {
			base = check.LinkedToken
		}
		var other bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_payment WHERE method=$1 AND store_token LIKE $2 AND user_id<>$3)`, p.Method, likePrefix(base)+"#%", p.UserID).Scan(&other); err != nil {
			return nil, err
		}
		if other {
			return nil, conflict("SUBSCRIPTION_OWNED_BY_ANOTHER_USER", "this subscription is linked to another account")
		}
		key = base + "#" + jsonInt(*check.ValidUntil)
	}
	won, err := s.markSuccess(ctx, p, successFields{StoreToken: key, Outcome: "paid", Allowed: []string{"pending", "expired"}})
	if _, unique := isUniqueViolation(err); unique {
		return nil, conflict("TOKEN_ALREADY_USED", "this purchase token was already used")
	}
	if err != nil {
		return nil, err
	}
	if !won {
		fresh, err := s.loadPayment(ctx, s.db, p.ID)
		if err != nil {
			return nil, err
		}
		if fresh.Status == "success" && fresh.StoreToken.String == key {
			return s.storeResult(ctx, fresh)
		}
		return nil, conflict("PAYMENT_CLOSED", "payment state changed: "+fresh.Status)
	}
	rewards, err := s.grantPayment(ctx, p.ID)
	if err != nil {
		code := "REWARD_PENDING"
		if app, ok := asApp(err); ok && (app.Code == "PRODUCT_ALREADY_BOUGHT" || app.Code == "PRIZE_ALREADY_OWNED") {
			code = app.Code
		}
		return map[string]any{"paymentId": p.ID, "status": "success", "rewardGranted": false, "error": code}, nil
	}
	return map[string]any{"paymentId": p.ID, "status": "success", "rewardGranted": true, "rewards": rewards}, nil
}

func (s *service) storeResult(ctx context.Context, p payment) (any, error) {
	out := map[string]any{"paymentId": p.ID, "status": p.Status, "rewardGranted": p.RewardGranted}
	if p.Status == "success" && !p.RewardGranted {
		out["error"] = "REWARD_PENDING"
		if p.RewardError.Valid && paidConflictCode(p.RewardError.String) {
			out["error"] = p.RewardError.String
		}
	}
	if p.RewardGranted {
		var raw string
		if err := s.db.QueryRowContext(ctx, `SELECT rewards::text FROM economy_user_prize WHERE source_type='payment' AND source_id=$1 AND claimed`, p.ID).Scan(&raw); err == nil {
			var items []prizeItem
			_ = json.Unmarshal([]byte(raw), &items)
			out["rewards"] = items
		}
	}
	return out, nil
}

func (s *service) rpcPurchaseStatus(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		PaymentID string `json:"paymentId"`
		Lang      string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	p, err := s.loadPayment(ctx, s.db, in.PaymentID)
	if err != nil {
		return nil, err
	}
	if p.UserID != userID {
		return nil, notFound("PAYMENT_NOT_FOUND")
	}
	if p.Status == "success" && !p.RewardGranted {
		s.grantPendingRewards(ctx, userID)
		if p, err = s.loadPayment(ctx, s.db, p.ID); err != nil {
			return nil, err
		}
	}
	out := map[string]any{"paymentId": p.ID, "status": p.Status, "rewardGranted": p.RewardGranted}
	if p.Error.Valid && p.Error.String != "" {
		out["error"] = p.Error.String
	} else if p.RewardError.Valid && p.RewardError.String != "" {
		out["error"] = p.RewardError.String
	}
	if giftID := s.paymentGiftID(ctx, p.ID); giftID != "" {
		var gs string
		_ = s.db.QueryRowContext(ctx, `SELECT status FROM economy_friend_gift WHERE id=$1`, giftID).Scan(&gs)
		out["friendGiftId"], out["friendGiftStatus"] = giftID, gs
	}
	if p.Status == "success" && p.RewardGranted {
		if product, err := s.loadProduct(ctx, s.db, p.ProductID); err == nil {
			if msg := s.purchaseMessage(ctx, userID, product, in.Lang); msg != "" {
				out["purchaseMessage"] = msg
			}
		}
	}
	return out, nil
}

func (s *service) rpcStoreConfig(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Method string `json:"method"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if !isStore(in.Method) {
		return nil, badRequest("INVALID_METHOD", "method", "method must be bazaar or myket")
	}
	pm, err := s.loadMethod(ctx, s.db, in.Method)
	if err != nil {
		return nil, err
	}
	if !pm.usable() {
		return nil, fail(http.StatusServiceUnavailable, "METHOD_UNAVAILABLE", "method", "this payment method is not available")
	}
	return map[string]any{"method": in.Method, "sandbox": pm.Config.flag("sandbox"), "rsaPublicKey": pm.Config.str("rsaPublicKey"), "packageName": pm.Config.str("packageName")}, nil
}

func (s *service) purchaseDescription(ctx context.Context, name i18n) string {
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		return "خرید " + name["fa"]
	}
	return "خرید " + langs.Resolve(name, "fa")
}

func (s *service) rpcPaymentsMine(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Limit  int    `json:"limit"`
		Cursor string `json:"cursor"`
		Lang   string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	if in.Limit <= 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}
	before := time.Now().Add(time.Hour)
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
	rows, err := s.db.QueryContext(ctx, `SELECT jsonb_build_object('paymentId',p.id,'productId',p.product_id,'productKey',pr.key,'productName',pr.name,'method',p.method,'status',p.status,
		'finalToman',p.final_toman,'coinCurrency',p.coin_currency,'coinAmount',p.coin_amount,'rewardGranted',p.reward_granted,'refId',p.ref_id,'friendGiftId',p.friend_gift_id,'createdAt',p.created_at)::text, p.created_at, p.id::text
		FROM economy_payment p JOIN economy_product pr ON pr.id=p.product_id
		WHERE p.user_id=$1 AND (p.created_at < $2 OR (p.created_at = $2 AND p.id < $3::uuid)) ORDER BY p.created_at DESC, p.id DESC LIMIT $4`, userID, before, beforeID, in.Limit+1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	var lastAt time.Time
	var lastID string
	next := ""
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
		item := map[string]any{}
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			return nil, err
		}
		names := i18n{}
		if m, ok := item["productName"].(map[string]any); ok {
			for k, v := range m {
				if str, ok := v.(string); ok {
					names[k] = str
				}
			}
		}
		item["productName"] = langs.Resolve(names, lang)
		items = append(items, item)
		lastAt, lastID = at, id
	}
	return map[string]any{"lang": lang, "items": items, "cursor": next}, rows.Err()
}
