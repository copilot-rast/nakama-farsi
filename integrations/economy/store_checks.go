package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

const storeTestPlaceholderSKU = "nakama_credentials_test"

func storeErrorClip(v string) string {
	v = strings.TrimSpace(v)
	if len(v) > 300 {
		v = v[:300]
	}
	return v
}

func storeClassName(c storeClass) string {
	switch c {
	case storeAccepted:
		return "accepted"
	case storeRejected:
		return "rejected"
	case storeConfigRejected:
		return "config_rejected"
	}
	return "unavailable"
}

func storeCheckRecord(check storeCheck, source string, at time.Time) map[string]any {
	out := map[string]any{
		"at":               at.UTC().Format(time.RFC3339),
		"source":           source,
		"class":            storeClassName(check.Class),
		"reason":           check.Reason,
		"storeError":       storeErrorClip(check.StoreError),
		"storeDescription": storeErrorClip(check.StoreErrorText),
		"durationMs":       int(check.Exchange.Duration / time.Millisecond),
	}
	if check.Exchange.HTTPStatus != nil {
		out["httpStatus"] = *check.Exchange.HTTPStatus
	}
	return out
}

func storeCallMade(check storeCheck) bool {
	return check.Exchange.HTTPStatus != nil || check.Reason == reasonNetwork || (check.Class == storeConfigRejected && check.Reason == reasonCredentials)
}

func (s *service) noteStoreCheck(ctx context.Context, method string, check storeCheck, source string) {
	if s.db == nil || !storeCallMade(check) || strings.HasPrefix(check.Exchange.URL, "sandbox://") {
		return
	}
	var err error
	switch check.Class {
	case storeAccepted, storeRejected:
		_, err = s.db.ExecContext(ctx, `UPDATE economy_payment_method SET store_last_ok_at=now() WHERE method=$1`, method)
	default:
		raw, _ := json.Marshal(storeCheckRecord(check, source, s.now()))
		_, err = s.db.ExecContext(ctx, `UPDATE economy_payment_method SET store_last_failure=$2::jsonb WHERE method=$1`, method, string(raw))
	}
	if err != nil {
		s.logger.Warn("Economy store check note failed: %v", err)
	}
}

func (s *service) storeHealth(ctx context.Context, method string) map[string]any {
	out := map[string]any{"lastFailure": nil, "lastOkAt": nil, "lastTest": nil}
	if !isStore(method) {
		return out
	}
	var failure, test sql.NullString
	var okAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `SELECT store_last_failure::text, store_last_ok_at, store_last_test::text FROM economy_payment_method WHERE method=$1`, method).Scan(&failure, &okAt, &test)
	if err != nil {
		return out
	}
	if failure.Valid {
		out["lastFailure"] = json.RawMessage(failure.String)
	}
	if okAt.Valid {
		out["lastOkAt"] = okAt.Time.UTC().Format(time.RFC3339)
	}
	if test.Valid {
		out["lastTest"] = json.RawMessage(test.String)
	}
	var stuck int64
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM economy_payment p WHERE p.method=$1 AND p.status IN ('pending','expired') AND p.created_at > now() - interval '30 days'
		AND EXISTS (SELECT 1 FROM economy_payment_log l WHERE l.payment_id=p.id AND l.action='store_validate' AND COALESCE(l.request->>'token', l.request->>'transactionId', '') <> '')`, method).Scan(&stuck)
	out["stuckPayments"] = stuck
	return out
}

func storeTestResult(check storeCheck) string {
	switch check.Class {
	case storeAccepted, storeRejected:
		return "credentials_ok"
	case storeConfigRejected:
		switch check.Reason {
		case reasonPackage:
			return "package_not_found"
		case reasonProduct:
			return "product_not_found"
		case reasonUnknownValue:
			return "config_rejected"
		}
		return "credentials_rejected"
	}
	return "unreachable"
}

func storeTestToken(method string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	if method == "apple" {
		n := "2"
		for _, c := range b {
			n += string(rune('0' + int(c)%10))
		}
		return n + "000000000"
	}
	return "nakama-credentials-test-" + hex.EncodeToString(b)
}

func (s *service) storeTestSKU(ctx context.Context, method, given string) (string, string) {
	if given = strings.TrimSpace(given); given != "" {
		return given, "given"
	}
	var sku string
	err := s.db.QueryRowContext(ctx, `SELECT sku->>$1 FROM economy_product WHERE COALESCE(sku->>$1,'') <> '' AND $1 = ANY(payment_methods) ORDER BY is_active DESC, created_at LIMIT 1`, method).Scan(&sku)
	if err == nil && sku != "" {
		return sku, "product"
	}
	return storeTestPlaceholderSKU, "placeholder"
}

func storeLiveConfigured(method string, cfg methodConfig) bool {
	live := methodConfig{}
	for k, v := range cfg {
		live[k] = v
	}
	live[sandboxField(method)] = false
	return methodConfigured(method, live)
}

func (s *service) consoleMethodTest(w http.ResponseWriter, r *http.Request) {
	method := pathVar(r, "method")
	if !isStore(method) {
		s.writeError(w, badRequest("INVALID_METHOD", "method", "only store methods can be tested"))
		return
	}
	var in struct {
		SKU string `json:"sku"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if in.SKU != "" && !skuPattern.MatchString(in.SKU) {
		s.writeError(w, badRequest("INVALID_SKU", "sku", "not a valid product id"))
		return
	}
	pm, err := s.loadMethod(r.Context(), s.db, method)
	if err != nil {
		s.writeError(w, err)
		return
	}
	sku, from := s.storeTestSKU(r.Context(), method, in.SKU)
	out := map[string]any{"method": method, "sku": sku, "skuSource": from, "testedAt": s.now().UTC().Format(time.RFC3339)}
	cfg := methodConfig{}
	for k, v := range pm.Config {
		cfg[k] = v
	}
	if !isAppStore(method) {
		cfg["sandbox"] = false
	}
	switch {
	case !storeLiveConfigured(method, pm.Config):
		out["result"] = "not_configured"
	case method == "apple" && !appleAPIKeys(cfg):
		out["result"] = "not_testable"
	default:
		check, err := s.validateStore(r.Context(), method, cfg, sku, storeTestToken(method), false)
		if err != nil {
			s.writeError(w, err)
			return
		}
		for k, v := range storeCheckRecord(check, "test", s.now()) {
			if k != "at" && k != "source" {
				out[k] = v
			}
		}
		out["result"] = storeTestResult(check)
	}
	raw, _ := json.Marshal(out)
	if _, err := s.db.ExecContext(r.Context(), `INSERT INTO economy_payment_method (method) VALUES ($1) ON CONFLICT (method) DO NOTHING`, method); err != nil {
		s.writeError(w, err)
		return
	}
	q := `UPDATE economy_payment_method SET store_last_test=$2::jsonb WHERE method=$1`
	if out["result"] == "credentials_ok" {
		q = `UPDATE economy_payment_method SET store_last_test=$2::jsonb, store_last_ok_at=now() WHERE method=$1`
	}
	if _, err := s.db.ExecContext(r.Context(), q, method, string(raw)); err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) lastStoreToken(ctx context.Context, paymentID string) string {
	var token string
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(NULLIF(request->>'token',''), request->>'transactionId', '') FROM economy_payment_log
		WHERE payment_id=$1 AND action='store_validate' AND COALESCE(NULLIF(request->>'token',''), request->>'transactionId', '') <> ''
		ORDER BY created_at DESC, id DESC LIMIT 1`, paymentID).Scan(&token)
	if err != nil {
		return ""
	}
	return token
}

func (s *service) recheckStore(ctx context.Context, id string) (map[string]any, error) {
	p, err := s.loadPayment(ctx, s.db, id)
	if err != nil {
		return nil, err
	}
	if !isStore(p.Method) {
		return nil, badRequest("INVALID_METHOD", "paymentId", "not a store payment")
	}
	if p.Status != "pending" && p.Status != "expired" {
		return nil, conflict("PAYMENT_NOT_PENDING", "only pending or expired store payments can be checked again")
	}
	token := s.lastStoreToken(ctx, p.ID)
	if token == "" {
		return nil, conflict("STORE_TOKEN_UNKNOWN", "the game never sent a purchase token for this payment")
	}
	product, err := s.loadProduct(ctx, s.db, p.ProductID)
	if err != nil {
		return nil, err
	}
	if product, err = s.paymentProduct(ctx, p.ID, product); err != nil {
		return nil, err
	}
	out := map[string]any{"paymentId": p.ID}
	res, err := s.settleStore(ctx, p, product, token)
	if err != nil {
		app, ok := asApp(err)
		if !ok {
			return nil, err
		}
		out["error"] = app.Code
	} else if m, ok := res.(map[string]any); ok {
		for k, v := range m {
			out[k] = v
		}
	}
	out["status"] = s.currentStatus(ctx, p.ID)
	return out, nil
}

func (s *service) consolePaymentRecheck(w http.ResponseWriter, r *http.Request) {
	res, err := s.recheckStore(r.Context(), pathVar(r, "id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = notFound("PAYMENT_NOT_FOUND")
		}
		s.writeError(w, err)
		return
	}
	detail, err := s.paymentDetail(r, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	detail["recheck"] = res
	writeJSON(w, http.StatusOK, detail)
}

const storeRecheckLimit = 100

func (s *service) consoleMethodRecheck(w http.ResponseWriter, r *http.Request) {
	method := pathVar(r, "method")
	if !isStore(method) {
		s.writeError(w, badRequest("INVALID_METHOD", "method", "only store methods can be checked again"))
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT p.id::text FROM economy_payment p WHERE p.method=$1 AND p.status IN ('pending','expired') AND p.created_at > now() - interval '30 days'
		AND EXISTS (SELECT 1 FROM economy_payment_log l WHERE l.payment_id=p.id AND l.action='store_validate' AND COALESCE(l.request->>'token', l.request->>'transactionId', '') <> '')
		ORDER BY p.created_at LIMIT $2`, method, storeRecheckLimit)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			s.writeError(w, err)
			return
		}
		ids = append(ids, id)
	}
	rows.Close()
	counts := map[string]int{"checked": 0, "delivered": 0, "stillPending": 0, "closed": 0}
	items := []map[string]any{}
	for _, id := range ids {
		res, err := s.recheckStore(r.Context(), id)
		if err != nil {
			app, ok := asApp(err)
			if !ok {
				s.writeError(w, err)
				return
			}
			res = map[string]any{"paymentId": id, "error": app.Code, "status": s.currentStatus(r.Context(), id)}
		}
		counts["checked"]++
		switch res["status"] {
		case "success":
			counts["delivered"]++
		case "pending", "expired":
			counts["stillPending"]++
		default:
			counts["closed"]++
		}
		items = append(items, res)
	}
	writeJSON(w, http.StatusOK, map[string]any{"method": method, "checked": counts["checked"], "delivered": counts["delivered"], "stillPending": counts["stillPending"], "closed": counts["closed"], "items": items})
}
