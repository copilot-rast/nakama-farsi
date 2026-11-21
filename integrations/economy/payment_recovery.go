package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

type paymentSettings struct {
	RewardRetry       bool   `json:"rewardRetry"`
	RetryMinutes      int    `json:"retryMinutes"`
	MaxAttempts       int    `json:"maxAttempts"`
	GrantOnNextCall   bool   `json:"grantOnNextCall"`
	PaidConflict      string `json:"paidConflict"`
	OncePending       string `json:"oncePending"`
	OnceHoldMinutes   int    `json:"onceHoldMinutes"`
	ExpiredReceipts   string `json:"expiredReceipts"`
	CloseSandboxOnOff bool   `json:"closeSandboxOnOff"`

	LiveAcceptsTestPurchases bool   `json:"liveAcceptsTestPurchases"`
	AppStorePayload          string `json:"appStorePayload"`
	StoreConsumed            string `json:"storeConsumed"`
	GoogleAcknowledge        bool   `json:"googleAcknowledge"`
	StripeToleranceSeconds   int    `json:"stripeToleranceSeconds"`
	PendingMinutes           int    `json:"pendingMinutes"`
	UpdatedAt                string `json:"updatedAt,omitempty"`
}

const consoleUsernameHeader = "X-Nakama-Console-Username"

func defaultPaymentSettings() paymentSettings {
	return paymentSettings{
		RewardRetry: true, RetryMinutes: 2, MaxAttempts: 20, GrantOnNextCall: true,
		PaidConflict: "refund", OncePending: "hold", OnceHoldMinutes: 15,
		ExpiredReceipts: "verified", CloseSandboxOnOff: true,
		AppStorePayload: "optional", StoreConsumed: "accept", GoogleAcknowledge: true, StripeToleranceSeconds: 300,
	}
}

func loadPaymentSettings(ctx context.Context, q queryer) (paymentSettings, error) {
	st := defaultPaymentSettings()
	var raw string
	var updated time.Time
	err := q.QueryRowContext(ctx, `SELECT settings::text, updated_at FROM economy_payment_settings WHERE id=1`).Scan(&raw, &updated)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return st, err
	}
	if err == nil {
		_ = json.Unmarshal([]byte(raw), &st)
		st.UpdatedAt = updated.UTC().Format(time.RFC3339)
	}
	return st, nil
}

func (s *service) pendingTTL(ctx context.Context) time.Duration {
	st, err := loadPaymentSettings(ctx, s.db)
	if err == nil && st.PendingMinutes > 0 {
		return time.Duration(st.PendingMinutes) * time.Minute
	}
	return s.cfg.pendingTTL
}

func validatePaymentSettings(st paymentSettings) error {
	switch {
	case st.RetryMinutes < 1 || st.RetryMinutes > 1440:
		return badRequest("INVALID_SETTING", "retryMinutes", "must be 1 to 1440")
	case st.MaxAttempts < 1 || st.MaxAttempts > 1000:
		return badRequest("INVALID_SETTING", "maxAttempts", "must be 1 to 1000")
	case st.PaidConflict != "refund" && st.PaidConflict != "keep":
		return badRequest("INVALID_SETTING", "paidConflict", "must be refund or keep")
	case st.OncePending != "hold" && st.OncePending != "replace":
		return badRequest("INVALID_SETTING", "oncePending", "must be hold or replace")
	case st.OnceHoldMinutes < 1 || st.OnceHoldMinutes > 1440:
		return badRequest("INVALID_SETTING", "onceHoldMinutes", "must be 1 to 1440")
	case st.ExpiredReceipts != "verified" && st.ExpiredReceipts != "refuse":
		return badRequest("INVALID_SETTING", "expiredReceipts", "must be verified or refuse")
	case st.AppStorePayload != "optional" && st.AppStorePayload != "required":
		return badRequest("INVALID_SETTING", "appStorePayload", "must be optional or required")
	case st.StoreConsumed != "accept" && st.StoreConsumed != "refuse":
		return badRequest("INVALID_SETTING", "storeConsumed", "must be accept or refuse")
	case st.StripeToleranceSeconds < 30 || st.StripeToleranceSeconds > 3600:
		return badRequest("INVALID_SETTING", "stripeToleranceSeconds", "must be 30 to 3600")
	case st.PendingMinutes < 0 || st.PendingMinutes > 10080:
		return badRequest("INVALID_SETTING", "pendingMinutes", "must be 0 to 10080")
	}
	return nil
}

func (s *service) consolePaymentSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := loadPaymentSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"settings": st})
}

func (s *service) consolePaymentSettingsPut(w http.ResponseWriter, r *http.Request) {
	st, err := loadPaymentSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := readBody(r, &st); err != nil {
		s.writeError(w, err)
		return
	}
	if err := validatePaymentSettings(st); err != nil {
		s.writeError(w, err)
		return
	}
	st.UpdatedAt = ""
	raw, _ := json.Marshal(st)
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO economy_payment_settings (id, settings, updated_at) VALUES (1,$1,now()) ON CONFLICT (id) DO UPDATE SET settings=EXCLUDED.settings, updated_at=now()`, string(raw)); err != nil {
		s.writeError(w, err)
		return
	}
	s.consolePaymentSettingsGet(w, r)
}

func paidConflictCode(code string) bool {
	return code == "PRODUCT_ALREADY_BOUGHT" || code == "PRIZE_ALREADY_OWNED"
}

func (s *service) notePaidConflict(ctx context.Context, id, code string) {
	st, err := loadPaymentSettings(ctx, s.db)
	if err != nil {
		st = defaultPaymentSettings()
	}
	status := "needed"
	if st.PaidConflict == "keep" {
		status = "kept"
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE economy_payment SET refund_status=$2, reward_error=$3, updated_at=now() WHERE id=$1 AND status='success' AND NOT reward_granted AND refund_status IS NULL`, id, status, code); err != nil {
		s.logger.Warn("Economy payment %s paid conflict not recorded: %v", id, err)
		return
	}
	s.logger.Warn("Economy payment %s was paid but cannot deliver (%s); marked %s for the admin", id, code, status)
}

const rewardRetryBatchSize = 20

func (s *service) claimRewardRetry(ctx context.Context, st paymentSettings) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `UPDATE economy_payment SET grant_attempts=grant_attempts+1,
		grant_retry_at=now() + make_interval(mins => $1::int * LEAST(power(2, LEAST(grant_attempts, 6))::int, 60))
		WHERE id = (SELECT id FROM economy_payment WHERE status='success' AND NOT reward_granted AND refund_status IS NULL
			AND COALESCE(outcome,'')<>'duplicate' AND friend_gift_id IS NULL AND grant_attempts < $2
			AND (grant_retry_at IS NULL OR grant_retry_at <= now()) AND COALESCE(verified_at, updated_at) < now() - interval '30 seconds'
			ORDER BY grant_retry_at NULLS FIRST, verified_at LIMIT 1 FOR UPDATE SKIP LOCKED)
		AND NOT reward_granted RETURNING id::text`, st.RetryMinutes, st.MaxAttempts).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (s *service) rewardRetryBatch(ctx context.Context) (int, error) {
	st, err := loadPaymentSettings(ctx, s.db)
	if err != nil {
		return 0, err
	}
	if !st.RewardRetry {
		return 0, nil
	}
	n := 0
	for n < rewardRetryBatchSize {
		id, err := s.claimRewardRetry(ctx, st)
		if err != nil {
			return n, err
		}
		if id == "" {
			break
		}
		n++
		if _, err := s.grantPayment(ctx, id); err != nil {
			s.logger.Warn("Economy reward retry of %s failed: %v", id, err)
		} else {
			s.logger.Info("Economy reward retry granted payment %s", id)
		}
	}
	return n, nil
}

func (s *service) grantPendingRewards(ctx context.Context, userID string) {
	st, err := loadPaymentSettings(ctx, s.db)
	if err != nil || !st.GrantOnNextCall {
		return
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text FROM economy_payment WHERE user_id=$1 AND status='success' AND NOT reward_granted AND refund_status IS NULL
		AND COALESCE(outcome,'')<>'duplicate' AND friend_gift_id IS NULL AND grant_attempts < $2
		AND (grant_retry_at IS NULL OR grant_retry_at <= now()) ORDER BY verified_at LIMIT 3`, userID, st.MaxAttempts)
	if err != nil {
		return
	}
	var ids []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		if _, err := s.db.ExecContext(ctx, `UPDATE economy_payment SET grant_attempts=grant_attempts+1, grant_retry_at=now() + make_interval(mins => $2::int) WHERE id=$1 AND NOT reward_granted`, id, st.RetryMinutes); err != nil {
			continue
		}
		if _, err := s.grantPayment(ctx, id); err != nil {
			s.logger.Warn("Economy reward for %s still pending: %v", id, err)
		}
	}
}

func (s *service) consolePaymentRefund(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	var in struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	in.Note = strings.TrimSpace(in.Note)
	if in.Status != "refunded" && in.Status != "kept" && in.Status != "needed" {
		s.writeError(w, badRequest("INVALID_STATUS", "status", "must be refunded, kept or needed"))
		return
	}
	if in.Status == "refunded" && in.Note == "" {
		s.writeError(w, badRequest("NOTE_REQUIRED", "note", "write how the money was returned"))
		return
	}
	if utf8.RuneCountInString(in.Note) > 1000 {
		s.writeError(w, badRequest("NOTE_TOO_LONG", "note", "at most 1000 characters"))
		return
	}
	p, err := s.loadPayment(r.Context(), s.db, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if p.Status != "success" || p.RewardGranted {
		s.writeError(w, conflict("REFUND_NOT_APPLICABLE", "only paid payments without a reward can be marked"))
		return
	}
	if _, err := s.db.ExecContext(r.Context(), `UPDATE economy_payment SET refund_status=$2, refund_note=NULLIF($3,''), refund_by=NULLIF($4,''),
		refund_at=CASE WHEN $2='refunded' THEN now() ELSE refund_at END, updated_at=now() WHERE id=$1`, id, in.Status, in.Note, strings.TrimSpace(r.Header.Get(consoleUsernameHeader))); err != nil {
		s.writeError(w, err)
		return
	}
	s.consolePaymentGet(w, r)
}

func (s *service) closeSandboxPending(ctx context.Context, method string) int {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text FROM economy_payment WHERE method=$1 AND sandbox AND status='pending'`, method)
	if err != nil {
		return 0
	}
	var ids []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	n := 0
	for _, id := range ids {
		if ok, _ := s.closePayment(ctx, id, "failed", "SANDBOX_OFF"); ok {
			n++
		}
	}
	return n
}

func paymentOnceKey(p productView, userID string) any {
	if p.friendGiftID() != nil {
		return nil
	}
	return productOnceKey(p, userID)
}

func (s *service) freeOnceKey(ctx context.Context, key any) error {
	if key == nil {
		return nil
	}
	var id, status string
	var created time.Time
	err := s.db.QueryRowContext(ctx, `SELECT id::text, status, created_at FROM economy_payment WHERE once_key=$1`, key).Scan(&id, &status, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if status != "pending" {
		return conflict("PRODUCT_ALREADY_BOUGHT", "this product can be bought only once")
	}
	st, err := loadPaymentSettings(ctx, s.db)
	if err != nil {
		return err
	}
	if st.OncePending == "hold" && created.After(s.now().Add(-time.Duration(st.OnceHoldMinutes)*time.Minute)) {
		return conflict("PAYMENT_IN_PROGRESS", "a payment for this product is still open: "+id)
	}
	if _, err := s.closePayment(ctx, id, "expired", "REPLACED"); err != nil {
		return err
	}
	return nil
}

func onceInsertError(err error) error {
	if c, unique := isUniqueViolation(err); unique && c == "economy_payment_once_idx" {
		return conflict("PAYMENT_IN_PROGRESS", "a payment for this product is already open")
	}
	return err
}

func (s *service) expiredAllowed(ctx context.Context, p payment, sandbox bool) error {
	if p.Status != "expired" {
		return nil
	}
	if sandbox || p.Sandbox {
		return conflict("PAYMENT_EXPIRED", "this payment has expired; start a new one")
	}
	st, err := loadPaymentSettings(ctx, s.db)
	if err != nil {
		return err
	}
	if st.ExpiredReceipts == "refuse" {
		return conflict("PAYMENT_EXPIRED", "this payment has expired; start a new one")
	}
	return nil
}
