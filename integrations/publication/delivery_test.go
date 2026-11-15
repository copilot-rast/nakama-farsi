package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
)

type testLogger struct{ runtime.Logger }

func (testLogger) Info(string, ...interface{})  {}
func (testLogger) Warn(string, ...interface{})  {}
func (testLogger) Error(string, ...interface{}) {}
func (testLogger) Debug(string, ...interface{}) {}

func testService() *service {
	cfg, err := loadSettings(map[string]string{"ECONOMY_SECRET_KEY": "test-secret-0123456789abcdef0123456789"})
	if err != nil {
		panic(err)
	}
	cfg.retryBackoff = time.Second
	return newService(nil, nil, testLogger{}, cfg)
}

type captured struct {
	method, path, auth, ctype, apiKey, idem, ua string
	form                                        url.Values
	body                                        map[string]any
}

func fakeAPI(t *testing.T, status int, reply string) (*httptest.Server, *captured) {
	t.Helper()
	c := &captured{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.method, c.path, c.auth, c.ctype = r.Method, r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("Content-Type")
		c.apiKey, c.idem, c.ua = r.Header.Get("X-API-KEY"), r.Header.Get("Idempotency-Key"), r.Header.Get("User-Agent")
		raw, _ := io.ReadAll(r.Body)
		if strings.HasPrefix(c.ctype, "application/json") {
			_ = json.Unmarshal(raw, &c.body)
		} else {
			c.form, _ = url.ParseQuery(string(raw))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(reply))
	}))
	t.Cleanup(srv.Close)
	return srv, c
}

func TestKavenegar(t *testing.T) {
	s := testService()
	cfg := providerConfig{"apiKey": "KEY/1", "sender": "10004346"}
	msg := outMessage{To: "+989121234567", Text: "سلام"}
	srv, c := fakeAPI(t, 200, `{"return":{"status":200,"message":"تایید شد"},"entries":[{"messageid":8792343,"message":"سلام","status":1,"statustext":"در صف ارسال","sender":"10004346","receptor":"09121234567","date":1356619709,"cost":120}]}`)
	s.baseURLs["kavenegar"] = srv.URL
	reply, err := s.deliver(context.Background(), provider{Kind: "kavenegar", Config: cfg}, msg)
	if err != nil {
		t.Fatal(err)
	}
	if c.method != http.MethodPost || c.path != "/v1/KEY/1/sms/send.json" || c.form.Get("receptor") != "09121234567" || c.form.Get("sender") != "10004346" || c.form.Get("message") != "سلام" {
		t.Fatalf("request = %+v", c)
	}
	if !strings.Contains(reply, "8792343") {
		t.Fatalf("reply = %s", reply)
	}
	srv2, _ := fakeAPI(t, 403, `{"return":{"status":403,"message":"کد شناسائی API-Key معتبر نمی‌باشد"},"entries":null}`)
	s.baseURLs["kavenegar"] = srv2.URL
	if _, err := s.deliver(context.Background(), provider{Kind: "kavenegar", Config: cfg}, msg); err == nil || !isPermanent(err) {
		t.Fatalf("403 must be permanent: %v", err)
	}
	srv3, _ := fakeAPI(t, 409, `{"return":{"status":409,"message":"server busy"},"entries":null}`)
	s.baseURLs["kavenegar"] = srv3.URL
	if _, err := s.deliver(context.Background(), provider{Kind: "kavenegar", Config: cfg}, msg); err == nil || isPermanent(err) {
		t.Fatalf("409 must be retried: %v", err)
	}
	if _, err := s.deliver(context.Background(), provider{Kind: "kavenegar", Config: cfg}, outMessage{To: "+14155550100", Text: "x"}); err == nil || !isPermanent(err) {
		t.Fatalf("foreign number must fail permanently: %v", err)
	}
	s.baseURLs["kavenegar"] = "http://127.0.0.1:1"
	_, err = s.deliver(context.Background(), provider{Kind: "kavenegar", Config: providerConfig{"apiKey": "SECRETKEY123"}}, msg)
	if err == nil || isPermanent(err) || strings.Contains(err.Error(), "SECRETKEY123") {
		t.Fatalf("network error must be transient and must not leak the key: %v", err)
	}
}

func TestSmsir(t *testing.T) {
	s := testService()
	cfg := providerConfig{"apiKey": "smsir-key", "lineNumber": "30007732"}
	msg := outMessage{To: "+989121234567", Text: "hello"}
	srv, c := fakeAPI(t, 200, `{"status":1,"message":"موفق","data":{"packId":"8b5f1d7e-1111-2222-3333-444455556666","messageIds":[86522023],"cost":1.0}}`)
	s.baseURLs["smsir"] = srv.URL
	reply, err := s.deliver(context.Background(), provider{Kind: "smsir", Config: cfg}, msg)
	if err != nil {
		t.Fatal(err)
	}
	if c.path != "/v1/send/bulk" || c.apiKey != "smsir-key" || c.body["lineNumber"] != float64(30007732) || c.body["messageText"] != "hello" {
		t.Fatalf("request = %+v", c)
	}
	if mobiles, _ := c.body["mobiles"].([]any); len(mobiles) != 1 || mobiles[0] != "09121234567" {
		t.Fatalf("mobiles = %v", c.body["mobiles"])
	}
	if !strings.Contains(reply, "86522023") {
		t.Fatalf("reply = %s", reply)
	}
	for _, tc := range []struct {
		status    int
		body      string
		permanent bool
	}{
		{400, `{"status":101,"message":"شماره خط نامعتبر میباشد","data":null}`, true},
		{200, `{"status":1,"message":"موفق","data":{"packId":"x","messageIds":[null],"cost":0}}`, true},
		{200, `{"status":1,"message":"موفق","data":{"packId":"x","messageIds":[0],"cost":0}}`, true},
		{401, `{"status":10,"message":"کلید وب سرویس نامعتبر است","data":null}`, true},
		{429, `too many`, false},
		{500, `{"status":0,"message":"system error","data":null}`, false},
	} {
		srv, _ := fakeAPI(t, tc.status, tc.body)
		s.baseURLs["smsir"] = srv.URL
		_, err := s.deliver(context.Background(), provider{Kind: "smsir", Config: cfg}, msg)
		if err == nil || isPermanent(err) != tc.permanent {
			t.Errorf("%d %s: err=%v permanent=%v", tc.status, tc.body, err, isPermanent(err))
		}
	}
}

func TestTwilio(t *testing.T) {
	s := testService()
	cfg := providerConfig{"accountSid": "ACxxxx", "authToken": "tok", "from": "+14155550100"}
	msg := outMessage{To: "+989121234567", Text: "hi"}
	srv, c := fakeAPI(t, 201, `{"sid":"SM123","status":"queued","error_code":null}`)
	s.baseURLs["twilio"] = srv.URL
	if _, err := s.deliver(context.Background(), provider{Kind: "twilio", Config: cfg}, msg); err != nil {
		t.Fatal(err)
	}
	user, pass, _ := (&http.Request{Header: http.Header{"Authorization": {c.auth}}}).BasicAuth()
	if c.path != "/2010-04-01/Accounts/ACxxxx/Messages.json" || user != "ACxxxx" || pass != "tok" || c.form.Get("To") != "+989121234567" || c.form.Get("From") != "+14155550100" || c.form.Get("Body") != "hi" {
		t.Fatalf("request = %+v", c)
	}
	srv2, _ := fakeAPI(t, 400, `{"code":21211,"message":"The 'To' number is not a valid phone number.","more_info":"https://www.twilio.com/docs/errors/21211","status":400}`)
	s.baseURLs["twilio"] = srv2.URL
	_, err := s.deliver(context.Background(), provider{Kind: "twilio", Config: cfg}, msg)
	if err == nil || !isPermanent(err) || !strings.Contains(err.Error(), "21211") {
		t.Fatalf("400 must be permanent with the code: %v", err)
	}
	srv3, _ := fakeAPI(t, 503, `{"code":20503,"message":"Service unavailable","status":503}`)
	s.baseURLs["twilio"] = srv3.URL
	if _, err := s.deliver(context.Background(), provider{Kind: "twilio", Config: cfg}, msg); err == nil || isPermanent(err) {
		t.Fatalf("503 must be retried: %v", err)
	}
	srv4, c4 := fakeAPI(t, 201, `{"sid":"SM9","status":"accepted"}`)
	s.baseURLs["twilio"] = srv4.URL
	cfg["from"] = "MG123"
	if _, err := s.deliver(context.Background(), provider{Kind: "twilio", Config: cfg}, msg); err != nil || c4.form.Get("MessagingServiceSid") != "MG123" || c4.form.Get("From") != "" {
		t.Fatalf("messaging service: %v %+v", err, c4.form)
	}
}

func TestResend(t *testing.T) {
	s := testService()
	cfg := providerConfig{"apiKey": "re_123", "from": "news@example.com", "fromName": "خبرنامه"}
	msg := outMessage{To: "u@example.com", Subject: "S", HTML: "<p>x</p>", Text: "x", IdempotencyKey: "n1/u1"}
	srv, c := fakeAPI(t, 200, `{"id":"49a3999c-0ce1-4ea6-ab68-afcd6dc2e794"}`)
	s.baseURLs["resend"] = srv.URL
	reply, err := s.deliver(context.Background(), provider{Kind: "resend", Config: cfg}, msg)
	if err != nil || !strings.Contains(reply, "49a3999c") {
		t.Fatalf("reply=%s err=%v", reply, err)
	}
	to, _ := c.body["to"].([]any)
	if c.path != "/emails" || c.auth != "Bearer re_123" || c.idem != "n1/u1" || c.ua == "" || len(to) != 1 || to[0] != "u@example.com" ||
		c.body["subject"] != "S" || c.body["html"] != "<p>x</p>" || c.body["text"] != "x" || !strings.HasSuffix(c.body["from"].(string), "<news@example.com>") {
		t.Fatalf("request = %+v", c)
	}
	srv2, _ := fakeAPI(t, 422, `{"statusCode":422,"name":"validation_error","message":"Invalid from field."}`)
	s.baseURLs["resend"] = srv2.URL
	if _, err := s.deliver(context.Background(), provider{Kind: "resend", Config: cfg}, msg); err == nil || !isPermanent(err) || !strings.Contains(err.Error(), "validation_error") {
		t.Fatalf("422 must be permanent: %v", err)
	}
	srv3, _ := fakeAPI(t, 429, `{"statusCode":429,"name":"rate_limit_exceeded","message":"Too many requests"}`)
	s.baseURLs["resend"] = srv3.URL
	if _, err := s.deliver(context.Background(), provider{Kind: "resend", Config: cfg}, msg); err == nil || isPermanent(err) {
		t.Fatalf("429 must be retried: %v", err)
	}
}

func TestKavenegarOTPPattern(t *testing.T) {
	s := testService()
	cfg := providerConfig{"apiKey": "KEY", "sender": "10004346", "otpTemplate": "rakhtkan-login", "otpToken2": "minutes", "otpToken3": "skip", "otpToken10": "app"}
	otp := &otpValues{Code: "48213", Minutes: "2", App: "رخت کن بازی"}
	msg := outMessage{To: "+989121234567", Text: "کد ورود شما: 48213", OTP: otp}
	srv, c := fakeAPI(t, 200, `{"return":{"status":200,"message":"تایید شد"},"entries":[{"messageid":8792343,"message":"کد: 48213","status":5,"statustext":"ارسال به مخابرات","sender":"10004346","receptor":"09121234567","date":1356619709,"cost":120}]}`)
	s.baseURLs["kavenegar"] = srv.URL
	reply, err := s.deliver(context.Background(), provider{Kind: "kavenegar", Config: cfg}, msg)
	if err != nil {
		t.Fatal(err)
	}
	if c.method != http.MethodPost || c.path != "/v1/KEY/verify/lookup.json" {
		t.Fatalf("request = %s %s", c.method, c.path)
	}
	want := url.Values{"receptor": {"09121234567"}, "template": {"rakhtkan-login"}, "token": {"48213"}, "token2": {"2"}, "token10": {"رخت کن بازی"}}
	if c.form.Encode() != want.Encode() {
		t.Fatalf("form = %v, want %v", c.form, want)
	}
	if !strings.Contains(reply, "8792343") {
		t.Fatalf("reply = %s", reply)
	}

	// app in token2 cannot carry spaces: words are joined with a zero-width non-joiner.
	cfg2 := providerConfig{"apiKey": "KEY", "otpTemplate": "login", "otpToken2": "app"}
	if _, err := s.deliver(context.Background(), provider{Kind: "kavenegar", Config: cfg2}, msg); err != nil {
		t.Fatal(err)
	}
	if got := c.form.Get("token2"); got != "رخت‌کن‌بازی" || c.form.Has("token3") || c.form.Has("token10") {
		t.Fatalf("token2 = %q form = %v", got, c.form)
	}

	// Without OTP values (campaigns) the template is ignored and plain send is used.
	if _, err := s.deliver(context.Background(), provider{Kind: "kavenegar", Config: cfg}, outMessage{To: msg.To, Text: "خبر تازه"}); err != nil {
		t.Fatal(err)
	}
	if c.path != "/v1/KEY/sms/send.json" || c.form.Get("message") != "خبر تازه" || c.form.Has("template") {
		t.Fatalf("plain send = %s %v", c.path, c.form)
	}
	// Without a template an OTP goes out as plain text, as before.
	if _, err := s.deliver(context.Background(), provider{Kind: "kavenegar", Config: providerConfig{"apiKey": "KEY"}}, msg); err != nil {
		t.Fatal(err)
	}
	if c.path != "/v1/KEY/sms/send.json" || c.form.Get("message") != msg.Text {
		t.Fatalf("otp without template = %s %v", c.path, c.form)
	}

	for _, tc := range []struct {
		status, ret int
		permanent   bool
	}{
		{200, 424, true}, // template not found or not approved
		{200, 431, true}, // invalid token
		{200, 418, true}, // no credit
		{200, 409, false},
		{200, 451, false},
		{502, 0, false},
	} {
		body := `{"return":{"status":` + strconv.Itoa(tc.ret) + `,"message":"x"},"entries":null}`
		if tc.ret == 0 {
			body = `bad gateway`
		}
		srv, _ := fakeAPI(t, tc.status, body)
		s.baseURLs["kavenegar"] = srv.URL
		_, err := s.deliver(context.Background(), provider{Kind: "kavenegar", Config: cfg}, msg)
		if err == nil || isPermanent(err) != tc.permanent {
			t.Errorf("%d/%d: err=%v permanent=%v", tc.status, tc.ret, err, isPermanent(err))
		}
	}
	if _, err := s.deliver(context.Background(), provider{Kind: "kavenegar", Config: cfg}, outMessage{To: "+14155550100", OTP: otp}); err == nil || !isPermanent(err) {
		t.Fatalf("foreign number must fail permanently: %v", err)
	}
}

func TestKavenegarToken(t *testing.T) {
	if got := kavenegarToken("token10", "a b c d e f g"); strings.Count(got, " ") != 5 {
		t.Fatalf("token10 = %q", got)
	}
	if got := kavenegarToken("token", strings.Repeat("۱", 150)); len([]rune(got)) != 100 {
		t.Fatalf("token length = %d", len([]rune(got)))
	}
}

func TestSmsirOTPPattern(t *testing.T) {
	s := testService()
	cfg := providerConfig{"apiKey": "smsir-key", "lineNumber": "30007732", "otpTemplateId": "123456", "otpCodeParam": "Code", "otpAppParam": "APP"}
	otp := &otpValues{Code: "48213", Minutes: "2", App: "یک نام خیلی خیلی بلند برای این بازی"}
	msg := outMessage{To: "+989121234567", Text: "code 48213", OTP: otp}
	srv, c := fakeAPI(t, 200, `{"status":1,"message":"موفق","data":{"messageId":89545112,"cost":1.0}}`)
	s.baseURLs["smsir"] = srv.URL
	reply, err := s.deliver(context.Background(), provider{Kind: "smsir", Config: cfg}, msg)
	if err != nil {
		t.Fatal(err)
	}
	if c.method != http.MethodPost || c.path != "/v1/send/verify" || c.apiKey != "smsir-key" || c.body["mobile"] != "09121234567" || c.body["templateId"] != float64(123456) {
		t.Fatalf("request = %+v", c)
	}
	params, _ := c.body["parameters"].([]any)
	if len(params) != 2 {
		t.Fatalf("parameters = %v", c.body["parameters"])
	}
	p0, _ := params[0].(map[string]any)
	p1, _ := params[1].(map[string]any)
	if p0["name"] != "Code" || p0["value"] != "48213" || p1["name"] != "APP" || len([]rune(p1["value"].(string))) != 25 {
		t.Fatalf("parameters = %v", params)
	}
	if !strings.Contains(reply, "89545112") {
		t.Fatalf("reply = %s", reply)
	}

	// Default code variable is CODE; the minutes variable is sent when named.
	if _, err := s.deliver(context.Background(), provider{Kind: "smsir", Config: providerConfig{"apiKey": "k", "lineNumber": "1", "otpTemplateId": "7", "otpMinutesParam": "MIN"}}, msg); err != nil {
		t.Fatal(err)
	}
	params, _ = c.body["parameters"].([]any)
	if len(params) != 2 || params[0].(map[string]any)["name"] != "CODE" || params[1].(map[string]any)["value"] != "2" {
		t.Fatalf("parameters = %v", params)
	}

	// Without OTP values or without a template, the plain bulk send is used.
	srvBulk, cb := fakeAPI(t, 200, `{"status":1,"message":"موفق","data":{"packId":"p","messageIds":[1],"cost":1.0}}`)
	s.baseURLs["smsir"] = srvBulk.URL
	if _, err := s.deliver(context.Background(), provider{Kind: "smsir", Config: cfg}, outMessage{To: msg.To, Text: "news"}); err != nil || cb.path != "/v1/send/bulk" {
		t.Fatalf("campaign send = %s %v", cb.path, err)
	}
	if _, err := s.deliver(context.Background(), provider{Kind: "smsir", Config: providerConfig{"apiKey": "k", "lineNumber": "1"}}, msg); err != nil || cb.path != "/v1/send/bulk" || cb.body["messageText"] != msg.Text {
		t.Fatalf("otp without template = %s %v %v", cb.path, cb.body, err)
	}

	for _, tc := range []struct {
		status    int
		body      string
		permanent bool
	}{
		{400, `{"status":113,"message":"قالب یافت نشد","data":null}`, true},
		{400, `{"status":114,"message":"طول مقدار پارامتر بیش از 25 کاراکتر","data":null}`, true},
		{401, `{"status":10,"message":"کلید وب سرویس نامعتبر است","data":null}`, true},
		{200, `{"status":1,"message":"موفق","data":null}`, true},
		{429, `{"status":20,"message":"limit","data":null}`, false},
		{500, `{"status":0,"message":"system error","data":null}`, false},
	} {
		srv, _ := fakeAPI(t, tc.status, tc.body)
		s.baseURLs["smsir"] = srv.URL
		_, err := s.deliver(context.Background(), provider{Kind: "smsir", Config: cfg}, msg)
		if err == nil || isPermanent(err) != tc.permanent {
			t.Errorf("%d %s: err=%v permanent=%v", tc.status, tc.body, err, isPermanent(err))
		}
	}
}

func TestOTPMode(t *testing.T) {
	cases := map[string]provider{
		"pattern (kavenegar template login)": {Kind: "kavenegar", Config: providerConfig{"otpTemplate": "login"}},
		"pattern (sms.ir template 77)":       {Kind: "smsir", Config: providerConfig{"otpTemplateId": "77"}},
		"plain text":                         {Kind: "twilio", Config: providerConfig{"otpTemplate": "x"}},
	}
	for want, p := range cases {
		if got := otpMode(p); got != want {
			t.Errorf("%s: %s", p.Kind, got)
		}
	}
	if got := otpMode(provider{Kind: "kavenegar", Config: providerConfig{}}); got != "plain text" {
		t.Errorf("kavenegar without template: %s", got)
	}
}
