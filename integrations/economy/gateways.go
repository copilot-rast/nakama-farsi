package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	zarinpalLiveAPI      = "https://api.zarinpal.com/pg/v4/payment"
	zarinpalSandboxAPI   = "https://sandbox.zarinpal.com/pg/v4/payment"
	zarinpalLiveStart    = "https://www.zarinpal.com/pg/StartPay"
	zarinpalSandboxStart = "https://sandbox.zarinpal.com/pg/StartPay"
	zibalBase            = "https://gateway.zibal.ir"
	bazaarBase           = "https://pardakht.cafebazaar.ir/devapi/v2"
	myketBase            = "https://developer.myket.ir"
)

const (
	zarinpalSandboxMerchant = "00000000-0000-0000-0000-000000000000"
	zibalSandboxMerchant    = "zibal"
)

type outcome int

const (
	outcomeUnknown outcome = iota
	outcomePaid
	outcomePaidNotVerified
	outcomePending
	outcomeFailed
)

type exchange struct {
	URL        string
	Request    any
	Response   any
	HTTPStatus *int
	Duration   time.Duration
}

type gatewayStart struct {
	OK        bool
	Reference string
	Redirect  string
	Error     string
	Exchange  exchange
}

type gatewayCheck struct {
	Outcome  outcome
	Amount   *int64
	RefID    string
	Card     string
	Code     *int64
	Status   *int64
	Error    string
	Exchange exchange
}

func (s *service) doJSON(ctx context.Context, method, target string, header map[string]string, body any, logged any) (map[string]any, exchange, error) {
	ex := exchange{URL: target, Request: logged}
	var reader io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return nil, ex, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	started := time.Now()
	resp, err := s.client.Do(req)
	ex.Duration = time.Since(started)
	if err != nil {
		ex.Response = map[string]any{"error": scrubError(err)}
		return nil, ex, err
	}
	defer resp.Body.Close()
	status := resp.StatusCode
	ex.HTTPStatus = &status
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var parsed map[string]any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&parsed); err != nil {
		text := string(raw)
		if len(text) > 2000 {
			text = text[:2000]
		}
		ex.Response = map[string]any{"raw": text}
		return nil, ex, nil
	}
	ex.Response = parsed
	return parsed, ex, nil
}

func scrubError(err error) string {
	var uerr *url.Error
	if errors.As(err, &uerr) {
		return uerr.Op + ": " + uerr.Err.Error()
	}
	return err.Error()
}

func section(m map[string]any, name string) map[string]any {
	if m == nil {
		return nil
	}
	if v, ok := m[name].(map[string]any); ok {
		return v
	}
	return nil
}

func numberField(m map[string]any, name string) *int64 {
	if m == nil {
		return nil
	}
	switch v := m[name].(type) {
	case json.Number:
		if n, err := v.Int64(); err == nil {
			return &n
		}
		if f, err := v.Float64(); err == nil {
			n := int64(f)
			return &n
		}
	case string:
		if n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64); err == nil {
			return &n
		}
	case float64:
		n := int64(v)
		return &n
	}
	return nil
}

func textField(m map[string]any, name string) string {
	if m == nil {
		return ""
	}
	switch v := m[name].(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	}
	return ""
}

func zarinpalAPI(sandbox bool) string {
	if sandbox {
		return zarinpalSandboxAPI
	}
	return zarinpalLiveAPI
}

func zarinpalMerchant(cfg methodConfig) string {
	if m := cfg.str("merchantId"); m != "" {
		return m
	}
	return zarinpalSandboxMerchant
}

func zibalMerchant(cfg methodConfig) string {
	if cfg.flag("sandbox") {
		return zibalSandboxMerchant
	}
	return cfg.str("merchant")
}

func maskedMerchant(sandbox bool, value string) string {
	if sandbox {
		return value
	}
	return secretMask
}

func (s *service) zarinpalRequest(ctx context.Context, cfg methodConfig, amountRial int64, callback, description string) gatewayStart {
	sandbox := cfg.flag("sandbox")
	merchant := zarinpalMerchant(cfg)
	body := map[string]any{"merchant_id": merchant, "amount": amountRial, "currency": "IRR", "callback_url": callback, "description": description}
	logged := map[string]any{"merchant_id": maskedMerchant(sandbox, merchant), "amount": amountRial, "currency": "IRR", "callback_url": maskTicketURL(callback), "description": description}
	reply, ex, err := s.doJSON(ctx, http.MethodPost, zarinpalAPI(sandbox)+"/request.json", nil, body, logged)
	out := gatewayStart{Exchange: ex}
	if err != nil {
		out.Error = "GATEWAY_UNAVAILABLE"
		return out
	}
	data := section(reply, "data")
	code := numberField(data, "code")
	authority := textField(data, "authority")
	if code != nil && *code == 100 && authority != "" {
		out.OK = true
		out.Reference = authority
		start := zarinpalLiveStart
		if sandbox {
			start = zarinpalSandboxStart
		}
		out.Redirect = start + "/" + url.PathEscape(authority)
		return out
	}
	out.Error = zarinpalError(reply)
	return out
}

func zarinpalError(reply map[string]any) string {
	errs := section(reply, "errors")
	if code := numberField(errs, "code"); code != nil {
		return fmt.Sprintf("ZARINPAL_%d", *code)
	}
	if code := numberField(section(reply, "data"), "code"); code != nil {
		return fmt.Sprintf("ZARINPAL_%d", *code)
	}
	return "GATEWAY_REJECTED"
}

func zarinpalVerifyOutcome(code *int64) outcome {
	if code == nil {
		return outcomeUnknown
	}
	switch *code {
	case 100, 101:
		return outcomePaid
	case -50, -51, -52, -53, -54, -55:
		return outcomeFailed
	}
	return outcomeUnknown
}

func (s *service) zarinpalVerify(ctx context.Context, cfg methodConfig, amountRial int64, authority string) gatewayCheck {
	sandbox := cfg.flag("sandbox")
	merchant := zarinpalMerchant(cfg)
	body := map[string]any{"merchant_id": merchant, "amount": amountRial, "authority": authority}
	logged := map[string]any{"merchant_id": maskedMerchant(sandbox, merchant), "amount": amountRial, "authority": authority}
	reply, ex, err := s.doJSON(ctx, http.MethodPost, zarinpalAPI(sandbox)+"/verify.json", nil, body, logged)
	out := gatewayCheck{Exchange: ex}
	if err != nil {
		out.Error = "GATEWAY_UNAVAILABLE"
		return out
	}
	data := section(reply, "data")
	code := numberField(data, "code")
	if code == nil {
		code = numberField(section(reply, "errors"), "code")
	}
	out.Code = code
	out.Outcome = zarinpalVerifyOutcome(code)
	out.RefID = textField(data, "ref_id")
	out.Card = textField(data, "card_pan")
	if a := numberField(data, "amount"); a != nil {
		out.Amount = a
	} else if out.Outcome == outcomePaid {
		bound := amountRial
		out.Amount = &bound
	}
	if out.Outcome != outcomePaid {
		out.Error = zarinpalError(reply)
	}
	return out
}

func (s *service) zibalPost(ctx context.Context, cfg methodConfig, path string, body map[string]any) (map[string]any, exchange, error) {
	logged := map[string]any{}
	for k, v := range body {
		logged[k] = v
	}
	logged["merchant"] = maskedMerchant(cfg.flag("sandbox"), zibalMerchant(cfg))
	if cb, ok := logged["callbackUrl"].(string); ok {
		logged["callbackUrl"] = maskTicketURL(cb)
	}
	return s.doJSON(ctx, http.MethodPost, strings.TrimRight(zibalBase, "/")+path, nil, body, logged)
}

func zibalTrack(trackID string) any {
	if n, err := strconv.ParseInt(trackID, 10, 64); err == nil {
		return n
	}
	return trackID
}

func (s *service) zibalRequest(ctx context.Context, cfg methodConfig, amountRial int64, callback, description, orderID string) gatewayStart {
	body := map[string]any{"merchant": zibalMerchant(cfg), "amount": amountRial, "callbackUrl": callback, "description": description, "orderId": orderID}
	reply, ex, err := s.zibalPost(ctx, cfg, "/v1/request", body)
	out := gatewayStart{Exchange: ex}
	if err != nil {
		out.Error = "GATEWAY_UNAVAILABLE"
		return out
	}
	result := numberField(reply, "result")
	track := textField(reply, "trackId")
	if result != nil && *result == 100 && track != "" {
		out.OK = true
		out.Reference = track
		out.Redirect = strings.TrimRight(zibalBase, "/") + "/start/" + url.PathEscape(track)
		return out
	}
	if result != nil {
		out.Error = fmt.Sprintf("ZIBAL_%d", *result)
	} else {
		out.Error = "GATEWAY_REJECTED"
	}
	return out
}

func zibalVerifyOutcome(result *int64) outcome {
	if result == nil {
		return outcomeUnknown
	}
	switch *result {
	case 100, 201:
		return outcomePaid
	case 202:
		return outcomeFailed
	}
	return outcomeUnknown
}

func zibalInquiryOutcome(result, status *int64) outcome {
	if result == nil || *result != 100 || status == nil {
		return outcomeUnknown
	}
	switch st := *status; {
	case st == 1:
		return outcomePaid
	case st == 2:
		return outcomePaidNotVerified
	case st == -1:
		return outcomePending
	case st >= 3 && st <= 12, st == 15, st == 16, st == 18:
		return outcomeFailed
	}
	return outcomeUnknown
}

func readZibal(reply map[string]any, ex exchange, oc outcome) gatewayCheck {
	out := gatewayCheck{Exchange: ex, Outcome: oc}
	out.Code = numberField(reply, "result")
	out.Status = numberField(reply, "status")
	out.Amount = numberField(reply, "amount")
	out.RefID = textField(reply, "refNumber")
	if card := textField(reply, "cardNumber"); len(card) > 1 {
		out.Card = card
	}
	if oc != outcomePaid && oc != outcomePaidNotVerified && out.Code != nil {
		out.Error = fmt.Sprintf("ZIBAL_%d", *out.Code)
	}
	return out
}

func (s *service) zibalVerify(ctx context.Context, cfg methodConfig, trackID string) gatewayCheck {
	reply, ex, err := s.zibalPost(ctx, cfg, "/v1/verify", map[string]any{"merchant": zibalMerchant(cfg), "trackId": zibalTrack(trackID)})
	if err != nil {
		return gatewayCheck{Exchange: ex, Error: "GATEWAY_UNAVAILABLE"}
	}
	return readZibal(reply, ex, zibalVerifyOutcome(numberField(reply, "result")))
}

func (s *service) zibalInquire(ctx context.Context, cfg methodConfig, trackID string) gatewayCheck {
	reply, ex, err := s.zibalPost(ctx, cfg, "/v1/inquiry", map[string]any{"merchant": zibalMerchant(cfg), "trackId": zibalTrack(trackID)})
	if err != nil {
		return gatewayCheck{Exchange: ex, Error: "GATEWAY_UNAVAILABLE"}
	}
	return readZibal(reply, ex, zibalInquiryOutcome(numberField(reply, "result"), numberField(reply, "status")))
}

type storeClass int

const (
	storeAccepted storeClass = iota
	storeRejected
	storeConfigRejected
	storeUnavailable
)

func classifyStoreReply(status int, errorField string) storeClass {
	switch {
	case status >= 200 && status < 300:
		return storeAccepted
	case status == 401 || status == 403:
		return storeConfigRejected
	case status == 404 && errorField == "invalid_value":
		return storeConfigRejected
	case status == 400 || status == 404 || status == 410:
		return storeRejected
	}
	return storeUnavailable
}

type storeCheck struct {
	Class       storeClass
	Valid       bool
	State       string
	Payload     string
	ValidUntil  *int64
	LinkedToken string
	Exchange    exchange
}

func storeSegment(value string) (string, error) {
	if value == "" || strings.Trim(value, ".") == "" {
		return "", badRequest("INVALID_TOKEN", "token", "token, sku and package name must not be empty or dots only")
	}
	return url.PathEscape(value), nil
}

func sandboxPayload(token string) string {
	if !strings.HasPrefix(token, "sandbox-") {
		return ""
	}
	body := strings.TrimPrefix(token, "sandbox-")
	split := strings.LastIndex(body, "-")
	if split <= 0 || split == len(body)-1 {
		return ""
	}
	return body[:split]
}

func (s *service) validateStore(ctx context.Context, method string, cfg methodConfig, sku, token string, subscription bool) (storeCheck, error) {
	if cfg.flag("sandbox") {
		payload := sandboxPayload(token)
		check := storeCheck{Class: storeAccepted, Valid: payload != "", Payload: payload, State: "SANDBOX"}
		if payload == "" {
			check.Class = storeRejected
			check.State = "SANDBOX_TOKEN_INVALID"
		}
		if subscription && payload != "" {
			until := s.now().Add(30 * 24 * time.Hour).UnixMilli()
			check.ValidUntil = &until
		}
		check.Exchange = exchange{URL: "sandbox://" + method, Request: map[string]any{"sku": sku, "token": token}, Response: map[string]any{"sandbox": true, "developerPayload": payload, "valid": check.Valid}}
		return check, nil
	}
	pkg, err := storeSegment(cfg.str("packageName"))
	if err != nil {
		return storeCheck{}, err
	}
	skuSeg, err := storeSegment(sku)
	if err != nil {
		return storeCheck{}, err
	}
	tokenSeg, err := storeSegment(token)
	if err != nil {
		return storeCheck{}, err
	}
	var target, verb string
	var body any
	header := map[string]string{}
	logged := map[string]any{"sku": sku, "token": token}
	switch method {
	case "bazaar":
		header["CAFEBAZAAR-PISHKHAN-API-SECRET"] = cfg.str("apiSecret")
		verb = http.MethodGet
		if subscription {
			target = strings.TrimRight(bazaarBase, "/") + "/api/applications/" + pkg + "/subscriptions/" + skuSeg + "/purchases/" + tokenSeg
		} else {
			target = strings.TrimRight(bazaarBase, "/") + "/api/validate/" + pkg + "/inapp/" + skuSeg + "/purchases/" + tokenSeg + "/"
		}
	case "myket":
		if subscription {
			return storeCheck{}, badRequest("SUBSCRIPTION_NOT_SUPPORTED", "method", "Myket subscriptions are not supported")
		}
		header["X-Access-Token"] = cfg.str("accessToken")
		verb = http.MethodPost
		target = strings.TrimRight(myketBase, "/") + "/api/partners/applications/" + pkg + "/purchases/products/" + skuSeg + "/verify"
		body = map[string]any{"tokenId": token}
	default:
		return storeCheck{}, badRequest("INVALID_METHOD", "method", "not a store method")
	}
	reply, ex, err := s.doJSON(ctx, verb, target, header, body, logged)
	check := storeCheck{Exchange: ex}
	if err != nil || ex.HTTPStatus == nil {
		check.Class = storeUnavailable
		return check, nil
	}
	check.Class = classifyStoreReply(*ex.HTTPStatus, textField(reply, "error"))
	if check.Class == storeAccepted && reply == nil {
		check.Class = storeUnavailable
	}
	if check.Class != storeAccepted {
		return check, nil
	}
	if subscription {
		check.ValidUntil = numberField(reply, "validUntilTimestampMsec")
		check.LinkedToken = textField(reply, "linkedSubscriptionToken")
		check.Valid = check.ValidUntil != nil && *check.ValidUntil >= s.now().UnixMilli()
		check.State = "EXPIRED"
		if check.Valid {
			check.State = "ACTIVE"
		}
		check.Payload = textField(reply, "developerPayload")
		return check, nil
	}
	state := numberField(reply, "purchaseState")
	check.Valid = state != nil && *state == 0
	check.Payload = textField(reply, "developerPayload")
	switch {
	case check.Valid:
		check.State = "VALID"
	case state != nil && *state == 1:
		check.State = "CANCELED"
	case state != nil && *state == 2:
		check.State = "REFUNDED"
	default:
		check.State = "INVALID"
	}
	return check, nil
}

func maskTicketURL(raw string) string {
	i := strings.LastIndex(raw, "/callback/")
	if i < 0 {
		return raw
	}
	return raw[:i] + "/callback/" + secretMask
}
