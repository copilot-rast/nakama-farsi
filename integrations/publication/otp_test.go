package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
)

func TestRandomDigits(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		c, err := randomDigits(6)
		if err != nil || len(c) != 6 || strings.Trim(c, "0123456789") != "" {
			t.Fatalf("randomDigits = %q, %v", c, err)
		}
		seen[c] = true
	}
	if len(seen) < 190 {
		t.Fatalf("codes repeat too often: %d unique of 200", len(seen))
	}
}

func TestOTPHashIsKeyedAndBound(t *testing.T) {
	s := testService()
	h := s.otpHash("id1", "+989121234567", "12345")
	if len(h) != 64 || strings.Contains(h, "12345") {
		t.Fatalf("hash = %q", h)
	}
	if h == s.otpHash("id2", "+989121234567", "12345") || h == s.otpHash("id1", "+989121234568", "12345") || h == s.otpHash("id1", "+989121234567", "12346") {
		t.Fatal("hash must depend on request id, phone and code")
	}
	other, err := loadSettings(map[string]string{"ECONOMY_SECRET_KEY": "another-secret-0123456789abcdef012345"})
	if err != nil {
		t.Fatal(err)
	}
	s2 := newService(nil, nil, testLogger{}, other)
	if h == s2.otpHash("id1", "+989121234567", "12345") {
		t.Fatal("hash must depend on the server secret")
	}
}

func TestOTPText(t *testing.T) {
	got := otpText("کد {app}: {code} ({minutes} دقیقه)", "48213", "ستاره‌چین", 121)
	if got != "کد ستاره‌چین: 48213 (3 دقیقه)" {
		t.Fatalf("otpText = %q", got)
	}
	if otpText("Code {code}", "1", "", 60) != "Code 1" {
		t.Fatal("plain template")
	}
}

func TestLoginKey(t *testing.T) {
	a, _ := newLoginKey()
	b, _ := newLoginKey()
	if a == b || !strings.HasPrefix(a, otpLoginKeyPrefix) || len(a) != len(otpLoginKeyPrefix)+40 {
		t.Fatalf("login keys %q %q", a, b)
	}
}

type otpLogger struct {
	runtime.Logger
	mu    sync.Mutex
	lines []string
}

func (l *otpLogger) Info(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, fmt.Sprintf(format, v...))
}
func (l *otpLogger) Warn(format string, v ...interface{})  { l.Info(format, v...) }
func (l *otpLogger) Error(format string, v ...interface{}) { l.Info(format, v...) }
func (l *otpLogger) Debug(string, ...interface{})          {}

var sandboxLine = regexp.MustCompile(`OTP sandbox code (\d+) for (\+\d+)`)

func (l *otpLogger) lastCode(phone string) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := len(l.lines) - 1; i >= 0; i-- {
		if m := sandboxLine.FindStringSubmatch(l.lines[i]); m != nil && m[2] == phone {
			return m[1]
		}
	}
	return ""
}

type otpNakama struct {
	fakeNakama
	h *harness
}

func (n *otpNakama) AuthenticateCustom(ctx context.Context, id, username string, create bool) (string, string, bool, error) {
	uid := uuid.Must(uuid.NewV4()).String()
	if _, err := n.h.s.db.ExecContext(ctx, `INSERT INTO users (id, username, custom_id) VALUES ($1,$2,$3)`, uid, "p"+strings.ReplaceAll(uid[:13], "-", ""), id); err != nil {
		return "", "", false, err
	}
	return uid, "", true, nil
}

func (n *otpNakama) LinkCustom(ctx context.Context, userID, customID string) error {
	_, err := n.h.s.db.ExecContext(ctx, `UPDATE users SET custom_id=$2 WHERE id=$1`, userID, customID)
	return err
}

func (n *otpNakama) AuthenticateTokenGenerate(userID, username string, exp int64, vars map[string]string) (string, int64, error) {
	return "token-" + userID, 1, nil
}

func otpHarness(t *testing.T) (*harness, *otpLogger) {
	h := newHarness(t)
	for _, q := range []string{
		`TRUNCATE publication_otp_code, publication_phone_account`,
		`DELETE FROM cluster_rate_limit WHERE key LIKE 'otp:%'`,
		`INSERT INTO publication_otp_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING`,
		`UPDATE publication_otp_settings SET is_enabled=true, sandbox=true, provider_id=NULL, code_length=5, ttl_seconds=120, max_attempts=3, resend_seconds=60, phone_hourly=5, phone_daily=10, ip_hourly=30,
			message='{"fa":"کد: {code}","en":"Code: {code}"}'`,
		`UPDATE base_country SET is_active = true`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	log := &otpLogger{}
	h.s.logger = log
	h.s.nk = &otpNakama{h: h}
	return h, log
}

func callerCtx(userID, ip string) context.Context {
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, userID)
	return context.WithValue(ctx, runtime.RUNTIME_CTX_CLIENT_IP, ip)
}

func appCode(err error) string {
	if app, ok := asApp(err); ok {
		return app.Code
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func TestOTPLoginFlowPostgres(t *testing.T) {
	h, log := otpHarness(t)
	guest := h.user(nil, "", "fa")
	ctx := callerCtx(guest, "10.0.0.1")
	if _, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"12"}`); appCode(err) != "INVALID_PHONE" {
		t.Fatalf("bad phone = %v", err)
	}
	out, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"۰۹۱۲۱۲۳۴۵۶۷"}`)
	if err != nil {
		t.Fatal(err)
	}
	res := out.(map[string]any)
	if res["phone"] != "+989121234567" || res["sandbox"] != true || res["codeLength"] != 5 || res["resendAfter"] != 60 {
		t.Fatalf("request = %v", res)
	}
	code := log.lastCode("+989121234567")
	if len(code) != 5 {
		t.Fatalf("sandbox code not logged: %v", log.lines)
	}
	var stored string
	if err := h.s.db.QueryRowContext(h.ctx, `SELECT code_hash FROM publication_otp_code WHERE phone='+989121234567'`).Scan(&stored); err != nil || strings.Contains(stored, code) || len(stored) != 64 {
		t.Fatalf("code must be stored hashed: %q %v", stored, err)
	}
	if _, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"09121234567"}`); appCode(err) != "OTP_TOO_SOON" {
		t.Fatalf("resend inside the cooldown = %v", err)
	}
	wrong := "00000"
	if wrong == code {
		wrong = "11111"
	}
	if _, err := h.s.rpcOTPVerify(ctx, guest, `{"phone":"09121234567","code":"`+wrong+`"}`); appCode(err) != "OTP_INVALID_CODE" {
		t.Fatalf("wrong code = %v", err)
	}
	var attempts int
	_ = h.s.db.QueryRowContext(h.ctx, `SELECT attempts FROM publication_otp_code WHERE phone='+989121234567'`).Scan(&attempts)
	if attempts != 1 {
		t.Fatalf("a wrong code must be counted and kept, attempts = %d", attempts)
	}
	out, err = h.s.rpcOTPVerify(ctx, guest, `{"phone":"+98 912 123 4567","code":"`+code+`"}`)
	if err != nil {
		t.Fatal(err)
	}
	login := out.(map[string]any)
	if login["created"] != true || login["linked"] != false || login["userId"] == guest || !strings.HasPrefix(login["loginKey"].(string), "ph_") || login["token"] != "token-"+login["userId"].(string) {
		t.Fatalf("verify = %v", login)
	}
	uid := login["userId"].(string)
	var phone, custom string
	if err := h.s.db.QueryRowContext(h.ctx, `SELECT metadata->>'phone', custom_id FROM users WHERE id=$1`, uid).Scan(&phone, &custom); err != nil || phone != "+989121234567" || custom != login["loginKey"] {
		t.Fatalf("account = %q %q %v", phone, custom, err)
	}
	if _, err := h.s.rpcOTPVerify(ctx, guest, `{"phone":"09121234567","code":"`+code+`"}`); appCode(err) != "OTP_NOT_FOUND" {
		t.Fatalf("a used code must not work twice: %v", err)
	}

	if _, err := h.s.db.ExecContext(h.ctx, `DELETE FROM cluster_rate_limit WHERE key LIKE 'otp:resend:%'`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"09121234567"}`); err != nil {
		t.Fatal(err)
	}
	again, err := h.s.rpcOTPVerify(ctx, guest, `{"phone":"09121234567","code":"`+log.lastCode("+989121234567")+`"}`)
	if err != nil {
		t.Fatal(err)
	}
	if again.(map[string]any)["userId"] != uid || again.(map[string]any)["created"] != false || again.(map[string]any)["loginKey"] != login["loginKey"] {
		t.Fatalf("second login must reach the same account: %v", again)
	}
	me, _ := h.s.rpcPhone(callerCtx(uid, ""), uid, `{}`)
	if me.(map[string]any)["phone"] != "+989121234567" {
		t.Fatalf("phone = %v", me)
	}
}

func TestOTPAttemptsExpiryAndLinkPostgres(t *testing.T) {
	h, log := otpHarness(t)
	guest := h.user(nil, "", "en")
	ctx := callerCtx(guest, "10.0.0.2")
	if _, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"09351112233"}`); err != nil {
		t.Fatal(err)
	}
	code := log.lastCode("+989351112233")
	bad := "9" + code[1:]
	if bad == code {
		bad = "8" + code[1:]
	}
	for i, want := range []string{"OTP_INVALID_CODE", "OTP_INVALID_CODE", "OTP_TOO_MANY_ATTEMPTS", "OTP_TOO_MANY_ATTEMPTS"} {
		if _, err := h.s.rpcOTPVerify(ctx, guest, `{"phone":"09351112233","code":"`+bad+`"}`); appCode(err) != want {
			t.Fatalf("attempt %d = %v, want %s", i+1, err, want)
		}
	}
	if _, err := h.s.rpcOTPVerify(ctx, guest, `{"phone":"09351112233","code":"`+code+`"}`); appCode(err) != "OTP_TOO_MANY_ATTEMPTS" {
		t.Fatalf("the right code after the limit = %v", err)
	}

	if _, err := h.s.db.ExecContext(h.ctx, `DELETE FROM cluster_rate_limit WHERE key LIKE 'otp:resend:%'`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"09351112233","lang":"en"}`); err != nil {
		t.Fatal(err)
	}
	code = log.lastCode("+989351112233")
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_otp_code SET expires_at = now() - interval '1 second' WHERE status='pending'`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.rpcOTPVerify(ctx, guest, `{"phone":"09351112233","code":"`+code+`"}`); appCode(err) != "OTP_EXPIRED" {
		t.Fatalf("expired code = %v", err)
	}

	if _, err := h.s.db.ExecContext(h.ctx, `DELETE FROM cluster_rate_limit WHERE key LIKE 'otp:resend:%'`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"09351112233"}`); err != nil {
		t.Fatal(err)
	}
	out, err := h.s.rpcOTPVerify(ctx, guest, `{"phone":"09351112233","code":"`+log.lastCode("+989351112233")+`","link":true}`)
	if err != nil {
		t.Fatal(err)
	}
	res := out.(map[string]any)
	if res["userId"] != guest || res["linked"] != true || res["created"] != false || res["loginKey"] == "" {
		t.Fatalf("link = %v", res)
	}
}

func TestOTPRateLimitsAndCountriesPostgres(t *testing.T) {
	h, _ := otpHarness(t)
	guest := h.user(nil, "", "fa")
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_otp_settings SET resend_seconds=10, phone_hourly=2, phone_daily=2`); err != nil {
		t.Fatal(err)
	}
	ctx := callerCtx(guest, "10.0.0.3")
	for i := 0; i < 2; i++ {
		if _, err := h.s.db.ExecContext(h.ctx, `DELETE FROM cluster_rate_limit WHERE key LIKE 'otp:resend:%'`); err != nil {
			t.Fatal(err)
		}
		if _, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"09191110000"}`); err != nil {
			t.Fatalf("request %d: %v", i+1, err)
		}
	}
	if _, err := h.s.db.ExecContext(h.ctx, `DELETE FROM cluster_rate_limit WHERE key LIKE 'otp:resend:%'`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"09191110000"}`); appCode(err) != "OTP_RATE_LIMITED" {
		t.Fatalf("third request in an hour = %v", err)
	}
	var replaced int
	_ = h.s.db.QueryRowContext(h.ctx, `SELECT count(*) FROM publication_otp_code WHERE phone='+989191110000' AND status='replaced'`).Scan(&replaced)
	if replaced != 1 {
		t.Fatalf("a new code must replace the old one, replaced = %d", replaced)
	}

	if _, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"1512345678","country":"DE"}`); err != nil {
		t.Fatalf("German number with country: %v", err)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE base_country SET is_active=false WHERE code='DE'`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"+4915123456789"}`); appCode(err) != "PHONE_COUNTRY_NOT_ALLOWED" {
		t.Fatalf("disabled country = %v", err)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_otp_settings SET is_enabled=false`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"09191110001"}`); appCode(err) != "OTP_DISABLED" {
		t.Fatalf("disabled = %v", err)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_otp_settings SET is_enabled=true, sandbox=false`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.rpcOTPRequest(ctx, guest, `{"phone":"09191110002"}`); appCode(err) != "OTP_UNAVAILABLE" {
		t.Fatalf("live without a provider = %v", err)
	}
}

func TestOTPLiveSendUsesProviderPostgres(t *testing.T) {
	h, _ := otpHarness(t)
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		got = r.Form.Get("receptor") + "|" + r.Form.Get("message")
		_, _ = w.Write([]byte(`{"return":{"status":200,"message":"ok"},"entries":[{"messageid":7,"status":1,"statustext":"queued"}]}`))
	}))
	defer srv.Close()
	h.s.baseURLs["kavenegar"] = srv.URL
	h.mustAPI(http.StatusCreated, http.MethodPost, "/providers", map[string]any{"channel": "sms", "kind": "kavenegar", "isActive": true, "isDefault": true, "sandbox": false, "config": map[string]any{"apiKey": "k", "sender": "1000"}})
	code, out := h.api(http.MethodPut, "/otp/settings", map[string]any{"isEnabled": true, "sandbox": false, "codeLength": 6, "ttlSeconds": 300, "maxAttempts": 3, "resendSeconds": 30,
		"phoneHourly": 5, "phoneDaily": 10, "ipHourly": 30, "message": map[string]any{"fa": "کد شما {code}", "en": "Code {code}"}})
	if code != http.StatusOK || out["sandbox"] != false || out["codeLength"] != float64(6) {
		t.Fatalf("settings = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPut, "/otp/settings", map[string]any{"isEnabled": true, "codeLength": 6, "ttlSeconds": 300, "maxAttempts": 3, "resendSeconds": 30,
		"phoneHourly": 5, "phoneDaily": 10, "ipHourly": 30, "message": map[string]any{"fa": "بدون کد"}}); code != http.StatusBadRequest || out["error"] != "OTP_MESSAGE_NEEDS_CODE" {
		t.Fatalf("message without {code} = %d %v", code, out)
	}
	guest := h.user(nil, "", "fa")
	res, err := h.s.rpcOTPRequest(callerCtx(guest, "10.0.0.4"), guest, `{"phone":"09121230000"}`)
	if err != nil {
		t.Fatal(err)
	}
	if res.(map[string]any)["sandbox"] != false || !strings.HasPrefix(got, "09121230000|کد شما ") || len([]rune(strings.TrimPrefix(got, "09121230000|کد شما "))) != 6 {
		t.Fatalf("live send = %q %v", got, res)
	}
	code, list := h.api(http.MethodGet, "/otp?q=0912123", nil)
	items, _ := list["items"].([]any)
	if code != http.StatusOK || len(items) != 1 || items[0].(map[string]any)["provider"] != "kavenegar" || items[0].(map[string]any)["status"] != "pending" {
		t.Fatalf("requests = %d %v", code, list)
	}
	if strings.Contains(fmt.Sprint(list), strings.TrimPrefix(got, "09121230000|کد شما ")) {
		t.Fatal("the console must never show the code")
	}
}

func TestOTPPatternSendAndProviderTestPostgres(t *testing.T) {
	h, _ := otpHarness(t)
	var path string
	var form url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		path, form = r.URL.Path, r.PostForm
		_, _ = w.Write([]byte(`{"return":{"status":200,"message":"ok"},"entries":[{"messageid":9,"status":5,"statustext":"sent"}]}`))
	}))
	defer srv.Close()
	h.s.baseURLs["kavenegar"] = srv.URL
	created := h.mustAPI(http.StatusCreated, http.MethodPost, "/providers", map[string]any{"channel": "sms", "kind": "kavenegar", "isActive": true, "isDefault": true, "sandbox": true,
		"config": map[string]any{"apiKey": "k", "otpTemplate": "login", "otpToken2": "minutes"}})
	id := created["id"].(string)
	if created["config"].(map[string]any)["otpTemplate"] != "login" {
		t.Fatalf("created = %v", created)
	}
	code, out := h.api(http.MethodPost, "/providers/"+id+"/test", map[string]any{"to": "09121230001", "otp": true})
	if code != http.StatusOK || out["sandbox"] != true || out["mode"] != "pattern (kavenegar template login)" || path != "" {
		t.Fatalf("sandbox otp test = %d %v (path %q)", code, out, path)
	}
	h.mustAPI(http.StatusOK, http.MethodPut, "/providers/"+id, map[string]any{"sandbox": false})
	code, out = h.api(http.MethodPost, "/providers/"+id+"/test", map[string]any{"to": "09121230001", "otp": true})
	if code != http.StatusOK || out["ok"] != true || path != "/v1/k/verify/lookup.json" || form.Get("template") != "login" || form.Get("token") != out["code"] || form.Get("token2") != "2" {
		t.Fatalf("live otp test = %d %v %s %v", code, out, path, form)
	}
	code, out = h.api(http.MethodPost, "/providers/"+id+"/test", map[string]any{"to": "09121230001", "message": "سلام"})
	if code != http.StatusOK || out["ok"] != true || path != "/v1/k/sms/send.json" || form.Get("message") != "سلام" {
		t.Fatalf("plain test = %d %v %s %v", code, out, path, form)
	}
	h.mustAPI(http.StatusOK, http.MethodPut, "/otp/settings", map[string]any{"isEnabled": true, "sandbox": false, "codeLength": 5, "ttlSeconds": 120, "maxAttempts": 3, "resendSeconds": 30,
		"phoneHourly": 5, "phoneDaily": 10, "ipHourly": 30, "message": map[string]any{"fa": "کد شما {code}"}})
	guest := h.user(nil, "", "fa")
	if _, err := h.s.rpcOTPRequest(callerCtx(guest, "10.0.0.5"), guest, `{"phone":"09121230002"}`); err != nil {
		t.Fatal(err)
	}
	if path != "/v1/k/verify/lookup.json" || form.Get("receptor") != "09121230002" || len(form.Get("token")) != 5 || form.Has("message") {
		t.Fatalf("login code = %s %v", path, form)
	}
}
