package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
