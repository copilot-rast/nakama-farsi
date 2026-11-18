package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

var stripeAPI = "https://api.stripe.com"

const (
	defaultStripeCurrency = "usd"
	stripeWebhookPath     = "/v2/economy/stripe/webhook"
	stripeBodyLimit       = 1 << 20
)

func nullText(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func stripeCurrency(cfg methodConfig) string {
	if c := strings.ToLower(cfg.str("currency")); c != "" {
		return c
	}
	return defaultStripeCurrency
}

func stripePriceFits(p productView, cfg methodConfig) bool {
	return p.PriceMinor != nil && *p.PriceMinor > 0 && strings.EqualFold(p.PriceCurrency, stripeCurrency(cfg))
}

func stripeKeyMode(key string) string {
	for _, prefix := range []string{"sk_", "rk_", "pk_"} {
		if strings.HasPrefix(key, prefix+"test_") {
			return "test"
		}
		if strings.HasPrefix(key, prefix+"live_") {
			return "live"
		}
	}
	return ""
}

func checkStripeConfig(cfg methodConfig) error {
	want := "live"
	if cfg.flag("testMode") {
		want = "test"
	}
	if key := cfg.str("secretKey"); key != "" {
		if !strings.HasPrefix(key, "sk_") && !strings.HasPrefix(key, "rk_") {
			return badRequest("INVALID_VALUE", "secretKey", "a Stripe secret key starts with sk_ or rk_")
		}
		if stripeKeyMode(key) != want {
			return badRequest("STRIPE_KEY_MODE", "secretKey", "test mode needs an sk_test_ key, live mode an sk_live_ key")
		}
	}
	if key := cfg.str("publishableKey"); key != "" {
		if !strings.HasPrefix(key, "pk_") {
			return badRequest("INVALID_VALUE", "publishableKey", "a Stripe publishable key starts with pk_")
		}
		if stripeKeyMode(key) != want {
			return badRequest("STRIPE_KEY_MODE", "publishableKey", "test mode needs a pk_test_ key, live mode a pk_live_ key")
		}
	}
	if key := cfg.str("webhookSecret"); key != "" && !strings.HasPrefix(key, "whsec_") {
		return badRequest("INVALID_VALUE", "webhookSecret", "a Stripe webhook signing secret starts with whsec_")
	}
	if c := cfg.str("currency"); c != "" && !priceCurrencyPattern.MatchString(strings.ToLower(c)) {
		return badRequest("INVALID_CURRENCY", "currency", "use a three-letter ISO currency code such as usd")
	}
	if c := cfg.str("currency"); c != "" {
		cfg["currency"] = strings.ToLower(c)
	}
	return nil
}

func (s *service) stripeSessionLifetime() time.Duration {
	d := s.cfg.pendingTTL
	if d < 31*time.Minute {
		d = 31 * time.Minute
	}
	if d > 23*time.Hour {
		d = 23 * time.Hour
	}
	return d
}

func (s *service) startStripe(ctx context.Context, userID string, product productView, pm paymentMethod, lang string) (any, error) {
	if product.offer != nil || product.giftID != "" {
		return nil, badRequest("METHOD_NOT_ALLOWED", "method", "stripe sells products only at their own price")
	}
	if !stripePriceFits(product, pm.Config) {
		return nil, badRequest("METHOD_NOT_ALLOWED", "method", "product has no price in "+stripeCurrency(pm.Config))
	}
	testMode := pm.Config.flag("testMode")
	if checkStripeConfig(pm.Config) != nil {
		return nil, fail(http.StatusServiceUnavailable, "METHOD_UNAVAILABLE", "method", "the Stripe keys do not match the test mode setting")
	}
	currency := stripeCurrency(pm.Config)
	amount := discounted(*product.PriceMinor, product.DiscountPercent)
	if amount < 1 {
		return nil, badRequest("PRICE_TOO_LOW", "method", "the final price must be at least 1")
	}
	id := uuid.Must(uuid.NewV4()).String()
	once := paymentOnceKey(product, userID)
	if err := s.freeOnceKey(ctx, once); err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO economy_payment (id, user_id, product_id, method, sandbox, status, original_toman, final_toman, amount_rial, discount_percent, once_key, currency, amount_minor)
		VALUES ($1,$2,$3,'stripe',$4,'pending',0,0,0,$5,$6,$7,$8)`, id, userID, product.ID, testMode, product.DiscountPercent, once, currency, amount); err != nil {
		return nil, onceInsertError(err)
	}
	name := product.Key
	if langs, l, err := localized.ForUser(ctx, s.db, userID, lang); err == nil {
		if v := strings.TrimSpace(langs.Resolve(product.Name, l)); v != "" {
			name = v
		}
	}
	ticket := signTicket(s.cfg.ticketKey, id, s.now().Add(ticketLifetime))
	back := s.cfg.publicBase + "/v2/economy/stripe/return/" + ticket
	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("client_reference_id", id)
	form.Set("success_url", back+"?result=success")
	form.Set("cancel_url", back+"?result=cancel")
	form.Set("line_items[0][quantity]", "1")
	form.Set("line_items[0][price_data][currency]", currency)
	form.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(amount, 10))
	form.Set("line_items[0][price_data][product_data][name]", name)
	form.Set("metadata[paymentId]", id)
	form.Set("metadata[productKey]", product.Key)
	form.Set("payment_intent_data[metadata][paymentId]", id)
	form.Set("expires_at", strconv.FormatInt(s.now().Add(s.stripeSessionLifetime()).Unix(), 10))
	logged := map[string]any{}
	for k, v := range form {
		logged[k] = v[0]
	}
	logged["success_url"] = s.cfg.publicBase + "/v2/economy/stripe/return/" + secretMask + "?result=success"
	logged["cancel_url"] = s.cfg.publicBase + "/v2/economy/stripe/return/" + secretMask + "?result=cancel"
	header := map[string]string{"Authorization": "Bearer " + pm.Config.str("secretKey"), "Idempotency-Key": "economy-" + id}
	reply, ex, err := s.doRequest(ctx, http.MethodPost, strings.TrimRight(stripeAPI, "/")+"/v1/checkout/sessions", header, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()), logged)
	s.logExchange(ctx, id, "request", ex)
	session := textField(reply, "id")
	redirect := textField(reply, "url")
	if err != nil || ex.HTTPStatus == nil || *ex.HTTPStatus != http.StatusOK || session == "" || redirect == "" {
		code := "GATEWAY_UNAVAILABLE"
		if e := section(reply, "error"); e != nil {
			code = "STRIPE_" + strings.ToUpper(firstNonEmpty(textField(e, "code"), textField(e, "type"), "rejected"))
		}
		_, _ = s.closePayment(ctx, id, "failed", code)
		return nil, fail(http.StatusBadGateway, "GATEWAY_REQUEST_FAILED", "", code)
	}
	if live, ok := reply["livemode"].(bool); ok && live == testMode {
		_, _ = s.closePayment(ctx, id, "failed", "STRIPE_MODE_MISMATCH")
		return nil, fail(http.StatusBadGateway, "GATEWAY_REQUEST_FAILED", "", "STRIPE_MODE_MISMATCH")
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE economy_payment SET authority=$2, updated_at=now() WHERE id=$1 AND status='pending'`, id, session); err != nil {
		_, _ = s.closePayment(ctx, id, "failed", "REFERENCE_NOT_SAVED")
		return nil, err
	}
	out := map[string]any{"paymentId": id, "status": "redirect", "redirectUrl": redirect, "sessionId": session, "currency": currency, "amountMinor": amount}
	if pk := pm.Config.str("publishableKey"); pk != "" {
		out["publishableKey"] = pk
	}
	return out, nil
}

func verifyStripeSignature(header string, body []byte, secret string, now time.Time, tolerance time.Duration) string {
	var ts int64
	var sigs []string
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			ts, _ = strconv.ParseInt(kv[1], 10, 64)
		case "v1":
			sigs = append(sigs, kv[1])
		}
	}
	if ts <= 0 || len(sigs) == 0 {
		return "INVALID_SIGNATURE"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10) + "."))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	match := false
	for _, sig := range sigs {
		if hmac.Equal([]byte(expected), []byte(sig)) {
			match = true
		}
	}
	if !match {
		return "INVALID_SIGNATURE"
	}
	age := now.Sub(time.Unix(ts, 0))
	if age < 0 {
		age = -age
	}
	if age > tolerance {
		return "SIGNATURE_TOO_OLD"
	}
	return ""
}

type stripeEvent struct {
	ID       string
	Type     string
	Livemode bool
	Object   map[string]any
}

func (s *service) stripeWebhookHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	raw, err := io.ReadAll(io.LimitReader(r.Body, stripeBodyLimit+1))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "BODY_UNREADABLE"})
		return
	}
	if len(raw) > stripeBodyLimit {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "BODY_TOO_LARGE"})
		return
	}
	pm, err := s.loadMethod(ctx, s.db, "stripe")
	if err != nil || pm.Config.str("webhookSecret") == "" {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "METHOD_UNAVAILABLE"})
		return
	}
	st := s.paySettings(ctx)
	if code := verifyStripeSignature(r.Header.Get("Stripe-Signature"), raw, pm.Config.str("webhookSecret"), s.now(), time.Duration(st.StripeToleranceSeconds)*time.Second); code != "" {
		s.logger.Warn("Economy Stripe webhook refused: %s", code)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": code})
		return
	}
	var body map[string]any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_JSON"})
		return
	}
	ev := stripeEvent{ID: textField(body, "id"), Type: textField(body, "type"), Object: section(section(body, "data"), "object")}
	ev.Livemode, _ = body["livemode"].(bool)
	if ev.ID == "" || ev.Type == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_EVENT"})
		return
	}
	if ev.Livemode == pm.Config.flag("testMode") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "MODE_MISMATCH"})
		return
	}
	var seen bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_payment_webhook WHERE method='stripe' AND event_id=$1)`, ev.ID).Scan(&seen); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL"})
		return
	}
	if seen {
		writeJSON(w, http.StatusOK, map[string]any{"received": true, "duplicate": true})
		return
	}
	result, paymentID, err := s.stripeHandle(ctx, ev)
	if err != nil {
		s.logger.Error("Economy Stripe webhook %s failed: %v", ev.ID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "INTERNAL"})
		return
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO economy_payment_webhook (method, event_id, type, payment_id, result) VALUES ('stripe',$1,$2,$3,$4) ON CONFLICT DO NOTHING`, ev.ID, ev.Type, nullText(paymentID), result); err != nil {
		s.logger.Warn("Economy Stripe webhook %s not recorded: %v", ev.ID, err)
	}
	writeJSON(w, http.StatusOK, map[string]any{"received": true, "result": result})
}

func (s *service) stripeHandle(ctx context.Context, ev stripeEvent) (string, string, error) {
	switch ev.Type {
	case "checkout.session.completed", "checkout.session.async_payment_succeeded", "checkout.session.async_payment_failed", "checkout.session.expired":
	default:
		return "ignored", "", nil
	}
	obj := ev.Object
	pid := firstNonEmpty(textField(obj, "client_reference_id"), textField(section(obj, "metadata"), "paymentId"))
	p, err := s.loadPayment(ctx, s.db, pid)
	if err != nil {
		if app, ok := asApp(err); ok && app.Code == "PAYMENT_NOT_FOUND" {
			return "unknown_payment", "", nil
		}
		return "", "", err
	}
	if p.Method != "stripe" {
		return "unknown_payment", "", nil
	}
	session := textField(obj, "id")
	if p.Authority.Valid && p.Authority.String != session {
		return "session_mismatch", p.ID, nil
	}
	amount := numberField(obj, "amount_total")
	currency := strings.ToLower(textField(obj, "currency"))
	s.logExchange(ctx, p.ID, "webhook", exchange{URL: stripeWebhookPath, Request: map[string]any{"eventId": ev.ID, "type": ev.Type, "sessionId": session, "paymentStatus": textField(obj, "payment_status"), "amountTotal": amount, "currency": currency, "paymentIntent": textField(obj, "payment_intent")}, Response: map[string]any{"status": p.Status}})
	switch ev.Type {
	case "checkout.session.async_payment_failed":
		_, err := s.closePayment(ctx, p.ID, "failed", "ASYNC_PAYMENT_FAILED")
		return "failed", p.ID, err
	case "checkout.session.expired":
		_, err := s.closePayment(ctx, p.ID, "expired", "CHECKOUT_EXPIRED")
		return "expired", p.ID, err
	}
	if textField(obj, "payment_status") != "paid" {
		if ev.Type == "checkout.session.completed" {
			s.notePending(ctx, p.ID, "ASYNC_PAYMENT_PENDING")
			return "pending", p.ID, nil
		}
		return "not_paid", p.ID, nil
	}
	if amount == nil || !p.AmountMinor.Valid || *amount != p.AmountMinor.Int64 || currency != p.Currency.String {
		s.logger.Warn("Economy Stripe payment %s paid %v %s but expected %d %s", p.ID, amount, currency, p.AmountMinor.Int64, p.Currency.String)
		_, err := s.closePayment(ctx, p.ID, "failed", "AMOUNT_MISMATCH")
		return "amount_mismatch", p.ID, err
	}
	won, err := s.markSuccess(ctx, p, successFields{RefID: textField(obj, "payment_intent"), Outcome: "paid", Allowed: []string{"pending", "expired", "failed"}})
	if err != nil {
		return "", p.ID, err
	}
	if won {
		if _, err := s.grantPayment(ctx, p.ID); err != nil {
			s.logger.Warn("Economy grant for Stripe payment %s failed: %v", p.ID, err)
		}
		return "paid", p.ID, nil
	}
	return "already_settled", p.ID, nil
}

func (s *service) stripeReturnHTTP(w http.ResponseWriter, r *http.Request) {
	paymentID, ok := readTicket(s.cfg.ticketKey, pathVar(r, "ticket"), s.now())
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "INVALID_TICKET"})
		return
	}
	p, err := s.loadPayment(r.Context(), s.db, paymentID)
	if err != nil || p.Method != "stripe" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "PAYMENT_NOT_FOUND"})
		return
	}
	result := r.URL.Query().Get("result")
	s.logExchange(r.Context(), p.ID, "callback", exchange{URL: "/v2/economy/stripe/return/" + secretMask, Request: map[string]any{"result": truncate(result, 32)}, Response: map[string]any{"status": p.Status}})
	success := result == "success" && p.Status != "failed" && p.Status != "expired"
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, s.redirectTarget(success, p.ID), http.StatusFound)
}
