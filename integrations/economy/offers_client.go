package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func (s *service) offerRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"economy_offers":        s.rpcOffers,
		"economy_offer_buy":     s.rpcOfferBuy,
		"economy_offers_reroll": s.rpcOffersReroll,
		"economy_offer_seen":    s.rpcOfferSeen,
	}
}

func (s *service) noteOfferLogin(ctx context.Context, userID string) {
	if _, err := s.db.ExecContext(ctx, `INSERT INTO economy_user_offer_state AS o (user_id, last_login_at, logins) VALUES ($1, $2, 1)
		ON CONFLICT (user_id) DO UPDATE SET prev_login_at=o.last_login_at, last_login_at=EXCLUDED.last_login_at, logins=o.logins+1, updated_at=now()`, userID, s.now()); err != nil {
		s.logger.Warn("Economy offer login note for %s failed: %v", userID, err)
	}
}

func (s *service) offerGameFinished(ctx context.Context, userID string) {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM economy_offer WHERE kind='entry' AND trigger_type='game' AND is_active`).Scan(&n); err != nil || n == 0 {
		return
	}
	st, err := loadOfferSettings(ctx, s.db)
	if err != nil || !st.EntryEnabled {
		return
	}
	oc, err := s.newOfferCtx(ctx, userID, st)
	if err == nil {
		_, err = oc.entryStates(true)
	}
	if err != nil {
		s.logger.Warn("Economy entry offers after a finished match for %s failed: %v", userID, err)
	}
}

func offerPercent(p productView) int {
	if p.offer == nil {
		return 0
	}
	return p.DiscountPercent
}

func (s *service) reserveOffer(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE economy_user_offer SET bought=bought+1, last_bought_at=now(), updated_at=now() WHERE id=$1 AND bought < purchase_limit`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return conflict("OFFER_SOLD_OUT", "you already bought this offer as many times as it allows")
	}
	return nil
}

func (s *service) releaseOffer(ctx context.Context, p productView) {
	if p.offer == nil {
		return
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE economy_user_offer SET bought=GREATEST(bought-1,0), updated_at=now() WHERE id=$1`, p.offer.ID)
}

func (s *service) offerBoughtEvent(ctx context.Context, paymentID string) {
	var userID, key, kind, offerID string
	var toman int64
	err := s.db.QueryRowContext(ctx, `SELECT p.user_id::text, t.key, u.kind, u.id::text, p.final_toman FROM economy_payment p JOIN economy_user_offer u ON u.id=p.user_offer_id JOIN economy_offer t ON t.id=u.offer_id WHERE p.id=$1`, paymentID).
		Scan(&userID, &key, &kind, &offerID, &toman)
	if err != nil {
		return
	}
	s.recordServerEvent(ctx, userID, offerBoughtEvent, float64(toman), map[string]any{"key": key, "kind": kind, "offerId": offerID, "paymentId": paymentID})
}

func (s *service) paymentProduct(ctx context.Context, paymentID string, p productView) (productView, error) {
	var offerID sql.NullString
	if err := s.db.QueryRowContext(ctx, `SELECT user_offer_id::text FROM economy_payment WHERE id=$1`, paymentID).Scan(&offerID); err != nil || !offerID.Valid {
		return p, nil
	}
	list, err := loadUserOffers(ctx, s.db, `WHERE u.id=$1`, offerID.String)
	if err != nil || len(list) == 0 {
		return p, err
	}
	return p.withOffer(list[0]), nil
}

func (s *service) offerForPurchase(ctx context.Context, userID, id string) (userOffer, offerSettings, error) {
	st, err := loadOfferSettings(ctx, s.db)
	if err != nil {
		return userOffer{}, st, err
	}
	if _, err := uuid.FromString(strings.TrimSpace(id)); err != nil {
		return userOffer{}, st, notFound("OFFER_NOT_FOUND")
	}
	list, err := loadUserOffers(ctx, s.db, `WHERE u.id=$1 AND u.user_id=$2`, strings.TrimSpace(id), userID)
	if err != nil {
		return userOffer{}, st, err
	}
	if len(list) == 0 {
		return userOffer{}, st, notFound("OFFER_NOT_FOUND")
	}
	u := list[0]
	now := s.now()
	if u.Status != "active" || !u.Tpl.live(now) || (u.Kind == "daily" && !st.DailyEnabled) || (u.Kind == "entry" && !st.EntryEnabled) {
		return u, st, notFound("OFFER_NOT_FOUND")
	}
	if u.expired(now) {
		return u, st, conflict("OFFER_EXPIRED", "this offer has ended")
	}
	if u.soldOut() {
		return u, st, conflict("OFFER_SOLD_OUT", "you already bought this offer as many times as it allows")
	}
	if u.Kind == "entry" && u.ExpiresAt == nil {
		oc := &offerCtx{s: s, ctx: ctx, userID: userID, now: now}
		if err := oc.startEntryTimers([]string{u.ID}); err != nil {
			return u, st, err
		}
	}
	return u, st, nil
}

type offerListOut struct {
	lang   string
	langs  *localized.Set
	oc     *offerCtx
	start  time.Time
	end    time.Time
	reroll int
}

func (s *service) offerItem(o *offerListOut, u userOffer) (map[string]any, error) {
	now := o.oc.now
	p, prize, err := o.oc.product(u.Tpl.ProductID)
	if err != nil {
		return nil, err
	}
	price := u.price()
	var coin any
	if price.CoinCurrency != "" && price.CoinAmount != nil {
		coin = map[string]any{"currencyKey": price.CoinCurrency, "amount": *price.CoinAmount}
	}
	methods := offerMethods(p, price, o.oc.methods)
	sku := map[string]string{}
	for _, m := range methods {
		if isStore(m) {
			sku[m] = price.SKU[m]
		}
	}
	name := o.langs.Resolve(u.Tpl.Name, o.lang)
	if name == "" {
		name = o.langs.Resolve(p.Name, o.lang)
	}
	desc := o.langs.Resolve(u.Tpl.Description, o.lang)
	if desc == "" {
		desc = o.langs.Resolve(p.Description, o.lang)
	}
	item := map[string]any{
		"offerId": u.ID, "key": u.Tpl.Key, "kind": u.Kind, "name": name, "description": desc, "badge": o.langs.Resolve(u.Tpl.Badge, o.lang),
		"product":       map[string]any{"id": p.ID, "key": p.Key, "name": o.langs.Resolve(p.Name, o.lang), "iconUrl": p.IconURL, "coverUrl": p.CoverURL, "items": prizeItemKeys(prize.Items)},
		"originalToman": price.OriginalToman, "finalToman": price.FinalToman, "coinPrice": coin, "coinOriginal": price.CoinOriginal, "discountPercent": price.DiscountPercent,
		"paymentMethods": methods, "sku": sku, "purchaseLimit": u.PurchaseLimit, "bought": u.Bought, "remaining": max(0, u.PurchaseLimit-u.Bought), "soldOut": u.soldOut(),
		"started": u.ExpiresAt != nil, "expiresAt": u.ExpiresAt, "secondsLeft": secondsLeft(now, u.ExpiresAt), "slot": u.Slot,
	}
	if u.ShownAt != nil {
		item["shownAt"] = u.ShownAt
	}
	if u.Kind == "entry" {
		item["durationMinutes"] = u.Tpl.DurationMinutes
	}
	return item, nil
}

func (s *service) offerList(ctx context.Context, userID, want string, oc *offerCtx) (map[string]any, error) {
	langs, lang, err := localized.ForUser(ctx, s.db, userID, want)
	if err != nil {
		return nil, err
	}
	st := oc.st
	start, end, rerolls, err := oc.ensureDaily()
	if err != nil {
		return nil, err
	}
	if st.EntryEnabled {
		if _, err := oc.entryStates(true); err != nil {
			return nil, err
		}
		if st.EntryTimerStart == "list" {
			if err := oc.startEntryTimers(nil); err != nil {
				return nil, err
			}
		}
	}
	out := &offerListOut{lang: lang, langs: langs, oc: oc, start: start, end: end, reroll: rerolls}
	now := oc.now
	daily := []map[string]any{}
	if st.DailyEnabled {
		list, err := loadUserOffers(ctx, s.db, `WHERE u.user_id=$1 AND u.kind='daily' AND u.period_start=$2 AND u.status='active'`, userID, start)
		if err != nil {
			return nil, err
		}
		sortOffersBySlot(list)
		unseen := []string{}
		for _, u := range list {
			if !u.Tpl.live(now) || (u.soldOut() && !st.ShowSoldOut) {
				continue
			}
			item, err := s.offerItem(out, u)
			if err != nil {
				return nil, err
			}
			daily = append(daily, item)
			if u.ShownAt == nil {
				unseen = append(unseen, u.ID)
			}
		}
		if len(unseen) > 0 {
			if _, err := s.db.ExecContext(ctx, `UPDATE economy_user_offer SET shown_at=$2, updated_at=now() WHERE id::text = ANY($1) AND shown_at IS NULL`, unseen, now); err != nil {
				return nil, err
			}
		}
	}
	entry := []map[string]any{}
	if st.EntryEnabled {
		list, err := loadUserOffers(ctx, s.db, `WHERE u.user_id=$1 AND u.kind='entry' AND u.status='active' AND (u.expires_at IS NULL OR u.expires_at > $2) ORDER BY t.sort, u.created_at`, userID, now)
		if err != nil {
			return nil, err
		}
		for _, u := range list {
			if !u.Tpl.live(now) || (u.soldOut() && (!st.ShowSoldOut || u.ExpiresAt == nil)) {
				continue
			}
			item, err := s.offerItem(out, u)
			if err != nil {
				return nil, err
			}
			entry = append(entry, item)
		}
	}
	reroll := map[string]any{"enabled": st.DailyEnabled && st.RerollEnabled, "used": rerolls, "limit": st.RerollLimit}
	if st.RerollLimit > 0 {
		reroll["left"] = max(0, st.RerollLimit-rerolls)
	}
	if cost := st.rerollPrice(rerolls); cost != nil {
		reroll["cost"] = cost
	} else {
		reroll["cost"] = nil
	}
	return map[string]any{
		"lang": lang, "now": now.UTC(),
		"daily": map[string]any{"enabled": st.DailyEnabled, "periodStart": start.UTC(), "periodEnd": end.UTC(), "secondsLeft": secondsLeft(now, &end), "items": daily, "reroll": reroll},
		"entry": map[string]any{"enabled": st.EntryEnabled, "timerStart": st.EntryTimerStart, "items": entry},
	}, nil
}

func (s *service) rpcOffers(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	st, err := loadOfferSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	oc, err := s.newOfferCtx(ctx, userID, st)
	if err != nil {
		return nil, err
	}
	return s.offerList(ctx, userID, in.Lang, oc)
}

func (s *service) rpcOfferBuy(ctx context.Context, userID, payload string) (any, error) {
	var in purchaseInput
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.OfferID) == "" {
		return nil, badRequest("REQUIRED", "offerId", "offerId is required")
	}
	return s.purchaseStart(ctx, userID, in)
}

func (s *service) rpcOfferSeen(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		OfferID string `json:"offerId"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if _, err := uuid.FromString(strings.TrimSpace(in.OfferID)); err != nil {
		return nil, notFound("OFFER_NOT_FOUND")
	}
	oc := &offerCtx{s: s, ctx: ctx, userID: userID, now: s.now()}
	list, err := loadUserOffers(ctx, s.db, `WHERE u.id=$1 AND u.user_id=$2`, in.OfferID, userID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, notFound("OFFER_NOT_FOUND")
	}
	u := list[0]
	if u.Kind == "entry" {
		if err := oc.startEntryTimers([]string{u.ID}); err != nil {
			return nil, err
		}
	} else if _, err := s.db.ExecContext(ctx, `UPDATE economy_user_offer SET shown_at=COALESCE(shown_at,$2), updated_at=now() WHERE id=$1`, u.ID, oc.now); err != nil {
		return nil, err
	}
	list, err = loadUserOffers(ctx, s.db, `WHERE u.id=$1`, u.ID)
	if err != nil || len(list) == 0 {
		return nil, err
	}
	return map[string]any{"offerId": u.ID, "expiresAt": list[0].ExpiresAt, "secondsLeft": secondsLeft(oc.now, list[0].ExpiresAt)}, nil
}

func (s *service) rpcOffersReroll(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	st, err := loadOfferSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if !st.DailyEnabled || !st.RerollEnabled {
		return nil, conflict("REROLL_DISABLED", "offers cannot be rerolled")
	}
	if err := s.ensureDefaults(ctx, userID); err != nil {
		return nil, err
	}
	oc, err := s.newOfferCtx(ctx, userID, st)
	if err != nil {
		return nil, err
	}
	start, end, _, err := oc.ensureDaily()
	if err != nil {
		return nil, err
	}
	var done int
	if err := s.db.QueryRowContext(ctx, `SELECT rerolls FROM economy_user_offer_period WHERE user_id=$1 AND period_start=$2`, userID, start).Scan(&done); err != nil {
		return nil, err
	}
	if st.RerollLimit > 0 && done >= st.RerollLimit {
		return nil, conflict("REROLL_LIMIT", "no rerolls left today")
	}
	paid := st.rerollPrice(done)
	if paid != nil {
		if _, _, err := s.nk.WalletUpdate(ctx, userID, map[string]int64{paid.CurrencyKey: -paid.Amount}, map[string]any{"module": "economy", "reason": "offer_reroll", "period": periodKey(start)}, true); err != nil {
			var neg *runtime.WalletNegativeError
			if errors.As(err, &neg) {
				return nil, fail(http.StatusPaymentRequired, "INSUFFICIENT_BALANCE", "", "not enough "+paid.CurrencyKey)
			}
			return nil, err
		}
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		var locked int
		if err := tx.QueryRowContext(ctx, `SELECT rerolls FROM economy_user_offer_period WHERE user_id=$1 AND period_start=$2 FOR UPDATE`, userID, start).Scan(&locked); err != nil {
			return err
		}
		if locked != done {
			return conflict("REROLL_BUSY", "another reroll just happened; try again")
		}
		current, err := loadUserOffers(ctx, tx, `WHERE u.user_id=$1 AND u.kind='daily' AND u.period_start=$2 AND u.status='active'`, userID, start)
		if err != nil {
			return err
		}
		keep := []string{}
		avoid := map[string]bool{}
		replace := []string{}
		for _, u := range current {
			if u.Bought > 0 {
				keep = append(keep, u.OfferID)
				continue
			}
			replace = append(replace, u.ID)
			if st.RerollAvoidCurrent {
				avoid[u.OfferID] = true
			}
		}
		if len(replace) == 0 {
			return conflict("NOTHING_TO_REROLL", "every offer today is already bought")
		}
		d, err := oc.drawDaily(start, done+1, keep, avoid)
		if err != nil {
			return err
		}
		if len(d.Picks) == 0 {
			return conflict("NO_OFFERS_LEFT", "there are no other offers to draw")
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_user_offer SET status='replaced', updated_at=now() WHERE id::text = ANY($1)`, replace); err != nil {
			return err
		}
		slot := len(current)
		for _, id := range d.Picks {
			if _, err := oc.insertOffer(tx, d.Pool[id], &start, slot, done+1, &end, false); err != nil {
				return err
			}
			slot++
		}
		_, err = tx.ExecContext(ctx, `UPDATE economy_user_offer_period SET rerolls=rerolls+1, updated_at=now() WHERE user_id=$1 AND period_start=$2`, userID, start)
		return err
	})
	if err != nil {
		if paid != nil {
			if _, _, rerr := s.nk.WalletUpdate(ctx, userID, map[string]int64{paid.CurrencyKey: paid.Amount}, map[string]any{"module": "economy", "reason": "offer_reroll_refund"}, true); rerr != nil {
				s.logger.Error("Economy offer reroll refund for %s failed: %v", userID, rerr)
			}
		}
		return nil, err
	}
	data := map[string]any{"rerolls": done + 1}
	if paid != nil {
		data["cost"] = paid.Amount
		data["currency"] = paid.CurrencyKey
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindPurchase, Action: "offers_rerolled", RefID: periodKey(start), Actor: "player", Data: data})
	fresh, err := s.newOfferCtx(ctx, userID, st)
	if err != nil {
		return nil, err
	}
	out, err := s.offerList(ctx, userID, in.Lang, fresh)
	if err != nil {
		return nil, err
	}
	out["paid"] = paid
	return out, nil
}
