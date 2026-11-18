package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

type payment struct {
	ID            string
	UserID        string
	ProductID     string
	Method        string
	Sandbox       bool
	Status        string
	OriginalToman int64
	FinalToman    int64
	AmountRial    int64
	GiftID        sql.NullString
	Authority     sql.NullString
	TrackID       sql.NullString
	StoreToken    sql.NullString
	Error         sql.NullString
	RewardGranted bool
	RewardError   sql.NullString
	CreatedAt     time.Time
	Outcome       string
	Currency      sql.NullString
	AmountMinor   sql.NullInt64
}

const paymentColumns = `id::text, user_id::text, product_id::text, method, sandbox, status, original_toman, final_toman, amount_rial, gift_id::text, authority, track_id, store_token, error, reward_granted, reward_error, created_at, COALESCE(outcome,''), currency, amount_minor`

func scanPayment(row interface{ Scan(...any) error }) (payment, error) {
	var p payment
	err := row.Scan(&p.ID, &p.UserID, &p.ProductID, &p.Method, &p.Sandbox, &p.Status, &p.OriginalToman, &p.FinalToman, &p.AmountRial, &p.GiftID, &p.Authority, &p.TrackID, &p.StoreToken, &p.Error, &p.RewardGranted, &p.RewardError, &p.CreatedAt, &p.Outcome, &p.Currency, &p.AmountMinor)
	return p, err
}

func (s *service) loadPayment(ctx context.Context, q queryer, id string) (payment, error) {
	if _, err := uuid.FromString(id); err != nil {
		return payment{}, notFound("PAYMENT_NOT_FOUND")
	}
	p, err := scanPayment(q.QueryRowContext(ctx, `SELECT `+paymentColumns+` FROM economy_payment WHERE id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return p, notFound("PAYMENT_NOT_FOUND")
	}
	return p, err
}

type productView struct {
	ID                 string            `json:"id"`
	Key                string            `json:"key"`
	Name               i18n              `json:"name"`
	Description        i18n              `json:"description"`
	IsActive           bool              `json:"isActive"`
	CategoryID         *string           `json:"categoryId"`
	TagIDs             []string          `json:"tagIds"`
	AudienceTagIDs     []string          `json:"audienceTagIds"`
	PriceToman         *int64            `json:"priceToman"`
	PriceMinor         *int64            `json:"priceMinor"`
	PriceCurrency      string            `json:"priceCurrency"`
	CoinPrice          *coinPrice        `json:"coinPrice"`
	DiscountPercent    int               `json:"discountPercent"`
	PaymentMethods     []string          `json:"paymentMethods"`
	SKU                map[string]string `json:"sku"`
	IsSubscription     bool              `json:"isSubscription"`
	PrizeID            string            `json:"prizeId"`
	IconURL            string            `json:"iconUrl"`
	CoverURL           string            `json:"coverUrl"`
	MinLevel           int               `json:"minLevel"`
	MaxLevel           *int              `json:"maxLevel"`
	LevelTableID       *string           `json:"levelTableId"`
	IsOneTime          bool              `json:"isOneTime"`
	CheckContents      bool              `json:"checkContents"`
	Giftable           *bool             `json:"giftable"`
	PurchaseMessage    i18n              `json:"purchaseMessage"`
	TagDiscountPercent int               `json:"-"`
	CoinOriginal       *coinPrice        `json:"-"`
	TagFinalToman      *int64            `json:"-"`
	offer              *offerRef
	giftID             string
}

func (p productView) friendGiftID() any {
	if p.giftID == "" {
		return nil
	}
	return p.giftID
}

func (p productView) allows(method string) bool {
	for _, m := range p.PaymentMethods {
		if m == method {
			return true
		}
	}
	return false
}

func (s *service) loadProduct(ctx context.Context, q queryer, id string) (productView, error) {
	var p productView
	if _, err := uuid.FromString(id); err != nil {
		return p, notFound("PRODUCT_NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+productJSON+`::text FROM economy_product t WHERE t.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return p, notFound("PRODUCT_NOT_FOUND")
	}
	if err != nil {
		return p, err
	}
	return p, json.Unmarshal([]byte(raw), &p)
}

func (s *service) logExchange(ctx context.Context, paymentID, action string, ex exchange) {
	req, _ := json.Marshal(ex.Request)
	resp, _ := json.Marshal(ex.Response)
	if ex.Request == nil {
		req = []byte("{}")
	}
	if ex.Response == nil {
		resp = []byte("{}")
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO economy_payment_log (id, payment_id, action, url, request, response, http_status, duration_ms) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		uuid.Must(uuid.NewV4()).String(), paymentID, action, ex.URL, string(req), string(resp), ex.HTTPStatus, int(ex.Duration/time.Millisecond)); err != nil {
		s.logger.Warn("Economy payment log write failed: %v", err)
	}
}

type successFields struct {
	RefID      string
	Card       string
	StoreToken string
	Outcome    string
	Allowed    []string
}

func (s *service) markSuccess(ctx context.Context, p payment, f successFields) (bool, error) {
	allowed := f.Allowed
	if len(allowed) == 0 {
		allowed = []string{"pending"}
	}
	won := false
	update := func(duplicate bool) error {
		won = false
		return s.inTx(ctx, func(tx *sql.Tx) error {
			once := `CASE WHEN friend_gift_id IS NULL AND (SELECT pr.is_one_time FROM economy_product pr WHERE pr.id=product_id) THEN user_id::text||':'||product_id::text ELSE NULL END`
			outcome, rewardErr := f.Outcome, ""
			if duplicate {
				once, outcome, rewardErr = `NULL`, "duplicate", "PRODUCT_ALREADY_BOUGHT"
			}
			res, err := tx.ExecContext(ctx, `UPDATE economy_payment SET status='success', ref_id=COALESCE(NULLIF($2,''), ref_id), card_mask=COALESCE(NULLIF($3,''), card_mask),
				store_token=COALESCE(NULLIF($4,''), store_token), outcome=$5, error=NULL, verified_at=now(), updated_at=now(), once_key=COALESCE(once_key, `+once+`),
				reward_error=NULLIF($7,'')
				WHERE id=$1 AND status = ANY($6)`, p.ID, f.RefID, f.Card, f.StoreToken, outcome, allowed, rewardErr)
			if err != nil {
				return err
			}
			n, _ := res.RowsAffected()
			if n == 0 {
				return nil
			}
			won = true
			return s.consumeGift(ctx, tx, p.ID)
		})
	}
	err := update(false)
	if c, unique := isUniqueViolation(err); unique && c == "economy_payment_once_idx" {
		err = update(true)
		if err == nil && won {
			s.logger.Warn("Economy payment %s paid a one-time product the player already bought; kept as duplicate for a refund", p.ID)
		}
	}
	return won, err
}

func (s *service) consumeGift(ctx context.Context, tx *sql.Tx, paymentID string) error {
	var giftID string
	err := tx.QueryRowContext(ctx, `UPDATE economy_gift_use SET status='used', updated_at=now() WHERE payment_id=$1 AND status IN ('reserved','released') RETURNING gift_id::text`, paymentID).Scan(&giftID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE economy_gift SET total_used = total_used + 1, updated_at=now() WHERE id=$1`, giftID)
	return err
}

func (s *service) closePayment(ctx context.Context, id, status, code string) (bool, error) {
	closed := false
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `UPDATE economy_payment SET status=$2, error=$3, outcome=$2, once_key=NULL, updated_at=now() WHERE id=$1 AND status='pending'`, id, status, truncate(code, 500))
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return nil
		}
		closed = true
		if _, err := tx.ExecContext(ctx, `UPDATE economy_friend_gift SET status='failed', error=$2, updated_at=now()
			WHERE id=(SELECT friend_gift_id FROM economy_payment WHERE id=$1) AND status='paying'`, id, truncate(code, 500)); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE economy_gift_use SET status='released', updated_at=now() WHERE payment_id=$1 AND status='reserved'`, id)
		return err
	})
	return closed, err
}

func (s *service) notePending(ctx context.Context, id, code string) {
	_, _ = s.db.ExecContext(ctx, `UPDATE economy_payment SET error=$2, updated_at=now() WHERE id=$1 AND status='pending'`, id, truncate(code, 500))
}

func (s *service) grantPayment(ctx context.Context, id string) ([]prizeItem, error) {
	p, err := s.loadPayment(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if p.Status != "success" {
		return nil, conflict("PAYMENT_NOT_SUCCESSFUL", "payment is not successful")
	}
	if p.Outcome == "duplicate" {
		s.notePaidConflict(ctx, p.ID, "PRODUCT_ALREADY_BOUGHT")
		return nil, conflict("PRODUCT_ALREADY_BOUGHT", "this one-time product was already delivered by another payment; refund this one")
	}
	product, err := s.loadProduct(ctx, s.db, p.ProductID)
	if err != nil {
		return nil, err
	}
	if giftID := s.paymentGiftID(ctx, p.ID); giftID != "" {
		return s.friendGiftPaid(ctx, p, product, giftID)
	}
	if p.Method == "in_app" {
		charged, err := s.inAppCharged(ctx, p.UserID, p.ID)
		if err != nil {
			return nil, err
		}
		if !charged {
			return nil, conflict("PAYMENT_NOT_CHARGED", "this coin payment took no coins, so it gives no reward")
		}
	}
	first := false
	rewards, err := s.grant(ctx, grantRequest{
		UserID: p.UserID, PrizeID: product.PrizeID, SourceType: "payment", SourceID: p.ID, Reason: "purchase", CheckItems: product.CheckContents,
		After: func(ctx context.Context, tx *sql.Tx) error {
			res, err := tx.ExecContext(ctx, `UPDATE economy_payment SET reward_granted=true, reward_error=NULL, updated_at=now() WHERE id=$1 AND NOT reward_granted`, p.ID)
			if err != nil {
				return err
			}
			n, _ := res.RowsAffected()
			first = n > 0
			if first {
				_, err = tx.ExecContext(ctx, `UPDATE economy_user_offer SET bought=bought+1, last_bought_at=now(), updated_at=now()
					WHERE id=(SELECT user_offer_id FROM economy_payment WHERE id=$1 AND method<>'in_app')`, p.ID)
			}
			return err
		},
	})
	if err != nil {
		code := err.Error()
		if app, ok := asApp(err); ok {
			code = app.Code
		}
		_, _ = s.db.ExecContext(ctx, `UPDATE economy_payment SET reward_error=$2, updated_at=now() WHERE id=$1 AND reward_granted=false`, p.ID, truncate(code, 500))
		if paidConflictCode(code) && p.Method != "in_app" {
			s.notePaidConflict(ctx, p.ID, code)
		}
		return nil, err
	}
	if first {
		s.recordServerEvent(ctx, p.UserID, "purchase_completed", float64(p.FinalToman), map[string]any{"paymentId": p.ID, "method": p.Method, "productKey": product.Key})
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: p.UserID, Kind: activity.KindPurchase, Action: "completed", RefID: p.ID, Actor: "player", Data: map[string]any{"method": p.Method, "productKey": product.Key, "toman": p.FinalToman}})
		s.offerBoughtEvent(ctx, p.ID)
	}
	return rewards, nil
}

func (s *service) callbackURL(paymentID string) string {
	return s.cfg.publicBase + "/v2/economy/payments/callback/" + signTicket(s.cfg.ticketKey, paymentID, s.now().Add(ticketLifetime))
}

func (s *service) checkGateway(ctx context.Context, p payment, callbackOK *bool, reconciling bool) (string, error) {
	if p.Status != "pending" {
		return p.Status, nil
	}
	pm, err := s.loadMethod(ctx, s.db, p.Method)
	if err != nil {
		return p.Status, err
	}
	if methodSandbox(p.Method, pm.Config) != p.Sandbox || !methodConfigured(p.Method, pm.Config) {
		s.logger.Warn("Economy payment %s cannot be checked: its gateway settings changed", p.ID)
		return p.Status, nil
	}
	var check gatewayCheck
	switch p.Method {
	case "zarinpal":
		if !p.Authority.Valid {
			return p.Status, nil
		}
		check = s.zarinpalVerify(ctx, pm.Config, p.AmountRial, p.Authority.String)
		s.logExchange(ctx, p.ID, "verify", check.Exchange)
		if reconciling && check.Outcome == outcomeFailed {
			check.Outcome = outcomePending
		}
	case "zibal":
		if !p.TrackID.Valid {
			return p.Status, nil
		}
		if callbackOK == nil || !*callbackOK {
			inquiry := s.zibalInquire(ctx, pm.Config, p.TrackID.String)
			s.logExchange(ctx, p.ID, "inquiry", inquiry.Exchange)
			switch inquiry.Outcome {
			case outcomeFailed:
				_, err := s.closePayment(ctx, p.ID, "failed", inquiry.Error)
				return s.currentStatus(ctx, p.ID), err
			case outcomePaid, outcomePaidNotVerified:
			default:
				return p.Status, nil
			}
		}
		check = s.zibalVerify(ctx, pm.Config, p.TrackID.String)
		s.logExchange(ctx, p.ID, "verify", check.Exchange)
		if check.Outcome == outcomePaid && check.Amount == nil {
			details := s.zibalInquire(ctx, pm.Config, p.TrackID.String)
			s.logExchange(ctx, p.ID, "inquiry", details.Exchange)
			check.Amount = details.Amount
			if check.RefID == "" {
				check.RefID = details.RefID
			}
			if check.Card == "" {
				check.Card = details.Card
			}
		}
	default:
		return p.Status, nil
	}
	switch check.Outcome {
	case outcomePaid:
		if check.Amount == nil {
			s.notePending(ctx, p.ID, "AMOUNT_UNKNOWN")
			return p.Status, nil
		}
		if *check.Amount != p.AmountRial {
			_, err := s.closePayment(ctx, p.ID, "failed", "AMOUNT_MISMATCH")
			s.logger.Warn("Economy payment %s captured %d rial but expected %d", p.ID, *check.Amount, p.AmountRial)
			return s.currentStatus(ctx, p.ID), err
		}
		won, err := s.markSuccess(ctx, p, successFields{RefID: check.RefID, Card: check.Card, Outcome: "paid"})
		if err != nil {
			return p.Status, err
		}
		if won {
			if _, err := s.grantPayment(ctx, p.ID); err != nil {
				s.logger.Warn("Economy grant for payment %s failed: %v", p.ID, err)
			}
		}
		return s.currentStatus(ctx, p.ID), nil
	case outcomeFailed:
		code := check.Error
		if code == "" {
			code = "NOT_VERIFIED"
		}
		_, err := s.closePayment(ctx, p.ID, "failed", code)
		return s.currentStatus(ctx, p.ID), err
	default:
		if check.Error != "" {
			s.notePending(ctx, p.ID, check.Error)
		}
		return p.Status, nil
	}
}

func (s *service) currentStatus(ctx context.Context, id string) string {
	var status string
	_ = s.db.QueryRowContext(ctx, `SELECT status FROM economy_payment WHERE id=$1`, id).Scan(&status)
	return status
}

func (s *service) callbackHTTP(w http.ResponseWriter, r *http.Request) {
	paymentID, ok := readTicket(s.cfg.ticketKey, pathVar(r, "ticket"), s.now())
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_TICKET"})
		return
	}
	ctx := r.Context()
	p, err := s.loadPayment(ctx, s.db, paymentID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "PAYMENT_NOT_FOUND"})
		return
	}
	_ = r.ParseForm()
	query := map[string]any{}
	for k, v := range r.Form {
		if len(v) > 0 {
			query[k] = truncate(v[0], 256)
		}
	}
	matched := false
	var callbackOK *bool
	switch p.Method {
	case "zarinpal":
		authority := strings.TrimSpace(firstNonEmpty(r.Form.Get("Authority"), r.Form.Get("authority")))
		matched = p.Authority.Valid && authority != "" && authority == p.Authority.String
		ok := strings.EqualFold(firstNonEmpty(r.Form.Get("Status"), r.Form.Get("status")), "OK")
		callbackOK = &ok
	case "zibal":
		track := strings.TrimSpace(firstNonEmpty(r.Form.Get("trackId"), r.Form.Get("trackid")))
		matched = p.TrackID.Valid && track != "" && track == p.TrackID.String
		ok := r.Form.Get("success") == "1"
		callbackOK = &ok
	}
	query["matched"] = matched
	s.logExchange(ctx, p.ID, "callback", exchange{URL: "/v2/economy/payments/callback/" + secretMask, Request: query, Response: map[string]any{"status": p.Status}})
	status := p.Status
	if matched {
		if p.Method == "zarinpal" {
			callbackOK = nil
		}
		status, err = s.checkGateway(ctx, p, callbackOK, false)
		if err != nil {
			s.logger.Error("Economy callback check failed for %s: %v", p.ID, err)
		}
	}
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, s.redirectTarget(status == "success", p.ID), http.StatusFound)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

const reconcileBatchSize = 20

func (s *service) reconcileLoop(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n, err := s.reconcileBatch(ctx); err != nil {
				s.logger.Warn("Economy reconciler failed: %v", err)
			} else if n > 0 {
				s.logger.Info("Economy reconciler checked %d payments", n)
			}
			if n, err := s.rewardRetryBatch(ctx); err != nil {
				s.logger.Warn("Economy reward retry failed: %v", err)
			} else if n > 0 {
				s.logger.Info("Economy reward retry tried %d payments", n)
			}
		}
	}
}

func (s *service) reconcileBatch(ctx context.Context) (int, error) {
	n := 0
	for n < reconcileBatchSize {
		id, err := s.claimForReconcile(ctx)
		if err != nil {
			return n, err
		}
		if id == "" {
			break
		}
		n++
		if _, err := s.reconcileClaimed(ctx, id); err != nil {
			s.logger.Warn("Economy reconcile of %s failed: %v", id, err)
		}
	}
	return n, nil
}

func (s *service) claimForReconcile(ctx context.Context) (string, error) {
	now := s.now()
	var id string
	err := s.db.QueryRowContext(ctx, `UPDATE economy_payment SET inquiry_count=inquiry_count+1, inquired_at=now()
		WHERE id = (SELECT id FROM economy_payment WHERE status='pending' AND created_at < $1 AND (inquired_at IS NULL OR inquired_at < $2)
			ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED)
		AND status='pending' RETURNING id::text`, now.Add(-3*time.Minute), now.Add(-time.Minute)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (s *service) reconcileOne(ctx context.Context, id string) (string, error) {
	res, err := s.db.ExecContext(ctx, `UPDATE economy_payment SET inquiry_count=inquiry_count+1, inquired_at=now() WHERE id=$1 AND status='pending'`, id)
	if err != nil {
		return "", err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return s.currentStatus(ctx, id), nil
	}
	return s.reconcileClaimed(ctx, id)
}

func (s *service) reconcileClaimed(ctx context.Context, id string) (string, error) {
	p, err := s.loadPayment(ctx, s.db, id)
	if err != nil {
		return "", err
	}
	if isGateway(p.Method) {
		yes := true
		var hint *bool
		if p.Method == "zibal" {
			pm, err := s.loadMethod(ctx, s.db, p.Method)
			if err == nil && p.TrackID.Valid && pm.Config.flag("sandbox") == p.Sandbox {
				inquiry := s.zibalInquire(ctx, pm.Config, p.TrackID.String)
				s.logExchange(ctx, p.ID, "inquiry", inquiry.Exchange)
				switch inquiry.Outcome {
				case outcomePaid, outcomePaidNotVerified:
					hint = &yes
				case outcomeFailed:
					if _, err := s.closePayment(ctx, p.ID, "failed", inquiry.Error); err != nil {
						return "", err
					}
					return s.currentStatus(ctx, p.ID), nil
				}
			}
			if hint == nil {
				return s.expireIfOld(ctx, p)
			}
		}
		status, err := s.checkGateway(ctx, p, hint, true)
		if err != nil {
			return status, err
		}
		if status != "pending" {
			return status, nil
		}
	}
	return s.expireIfOld(ctx, p)
}

func (s *service) expireIfOld(ctx context.Context, p payment) (string, error) {
	if p.CreatedAt.After(s.now().Add(-s.cfg.pendingTTL)) {
		return "pending", nil
	}
	if _, err := s.closePayment(ctx, p.ID, "expired", "TIMED_OUT"); err != nil {
		return "", err
	}
	return s.currentStatus(ctx, p.ID), nil
}
