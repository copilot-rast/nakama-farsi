package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
)

type outMessage struct {
	To             string
	Subject        string
	Text           string
	HTML           string
	Lang           string
	IdempotencyKey string
}

type deliveryError struct {
	msg       string
	permanent bool
}

func (e *deliveryError) Error() string { return e.msg }

func permanentErr(format string, args ...any) error {
	return &deliveryError{msg: fmt.Sprintf(format, args...), permanent: true}
}

func transientErr(format string, args ...any) error {
	return &deliveryError{msg: fmt.Sprintf(format, args...)}
}

func isPermanent(err error) bool {
	var d *deliveryError
	return errors.As(err, &d) && d.permanent
}

func (s *service) deliver(ctx context.Context, p provider, msg outMessage) (string, error) {
	switch p.Kind {
	case "kavenegar":
		return s.sendKavenegar(ctx, p.Config, msg)
	case "smsir":
		return s.sendSmsir(ctx, p.Config, msg)
	case "twilio":
		return s.sendTwilio(ctx, p.Config, msg)
	case "smtp":
		return s.sendSMTP(ctx, p.Config, msg)
	case "resend":
		return s.sendResend(ctx, p.Config, msg)
	}
	return "", permanentErr("unknown provider kind %q", p.Kind)
}

func (s *service) doHTTP(req *http.Request, name string) (int, []byte, error) {
	resp, err := s.client.Do(req)
	if err != nil {
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		return 0, nil, transientErr("%s request failed: %v", name, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	return resp.StatusCode, body, nil
}

func transientStatus(code int) bool {
	return code == http.StatusTooManyRequests || code >= 500
}

func (s *service) sendKavenegar(ctx context.Context, cfg providerConfig, msg outMessage) (string, error) {
	local, ok := iranLocal(msg.To)
	if !ok {
		return "", permanentErr("kavenegar only sends to Iranian mobile numbers")
	}
	form := url.Values{"receptor": {local}, "message": {msg.Text}}
	if sender := cfg.str("sender"); sender != "" {
		form.Set("sender", sender)
	}
	endpoint := s.baseURLs["kavenegar"] + "/v1/" + url.PathEscape(cfg.str("apiKey")) + "/sms/send.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", permanentErr("kavenegar request could not be built")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	code, body, err := s.doHTTP(req, "kavenegar")
	if err != nil {
		return "", err
	}
	var out struct {
		Return struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		} `json:"return"`
		Entries []struct {
			MessageID  int64  `json:"messageid"`
			Status     int    `json:"status"`
			StatusText string `json:"statustext"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(body, &out); err != nil || out.Return.Status == 0 {
		if transientStatus(code) {
			return "", transientErr("kavenegar HTTP %d", code)
		}
		return "", transientErr("kavenegar HTTP %d with an unreadable reply", code)
	}
	if out.Return.Status != 200 {
		reason := fmt.Sprintf("kavenegar %d: %s", out.Return.Status, out.Return.Message)
		switch out.Return.Status {
		case 409, 429, 451:
			return "", transientErr("%s", reason)
		}
		return "", permanentErr("%s", reason)
	}
	if len(out.Entries) == 0 {
		return "kavenegar 200", nil
	}
	return fmt.Sprintf("kavenegar messageid=%d status=%d %s", out.Entries[0].MessageID, out.Entries[0].Status, out.Entries[0].StatusText), nil
}

func (s *service) sendSmsir(ctx context.Context, cfg providerConfig, msg outMessage) (string, error) {
	local, ok := iranLocal(msg.To)
	if !ok {
		return "", permanentErr("sms.ir only sends to Iranian mobile numbers")
	}
	line, err := strconv.ParseInt(cfg.str("lineNumber"), 10, 64)
	if err != nil {
		return "", permanentErr("sms.ir lineNumber is not a number")
	}
	payload, _ := json.Marshal(map[string]any{"lineNumber": line, "messageText": msg.Text, "mobiles": []string{local}})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURLs["smsir"]+"/v1/send/bulk", bytes.NewReader(payload))
	if err != nil {
		return "", permanentErr("sms.ir request could not be built")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-KEY", cfg.str("apiKey"))
	code, body, err := s.doHTTP(req, "sms.ir")
	if err != nil {
		return "", err
	}
	var out struct {
		Status  *int   `json:"status"`
		Message string `json:"message"`
		Data    *struct {
			PackID     string   `json:"packId"`
			MessageIDs []*int64 `json:"messageIds"`
			Cost       float64  `json:"cost"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil || out.Status == nil {
		if transientStatus(code) {
			return "", transientErr("sms.ir HTTP %d", code)
		}
		if code == http.StatusUnauthorized {
			return "", permanentErr("sms.ir HTTP 401: invalid API key")
		}
		return "", transientErr("sms.ir HTTP %d with an unreadable reply", code)
	}
	if *out.Status != 1 {
		reason := fmt.Sprintf("sms.ir %d: %s", *out.Status, out.Message)
		if *out.Status == 0 || transientStatus(code) {
			return "", transientErr("%s", reason)
		}
		return "", permanentErr("%s", reason)
	}
	if out.Data == nil || len(out.Data.MessageIDs) == 0 || out.Data.MessageIDs[0] == nil {
		return "", permanentErr("sms.ir rejected the number or text (no message id)")
	}
	if *out.Data.MessageIDs[0] == 0 {
		return "", permanentErr("sms.ir: the number is on the blacklist")
	}
	return fmt.Sprintf("sms.ir packId=%s messageId=%d cost=%g", out.Data.PackID, *out.Data.MessageIDs[0], out.Data.Cost), nil
}

func (s *service) sendTwilio(ctx context.Context, cfg providerConfig, msg outMessage) (string, error) {
	sid := cfg.str("accountSid")
	form := url.Values{"To": {msg.To}, "Body": {msg.Text}}
	from := cfg.str("from")
	if strings.HasPrefix(from, "MG") {
		form.Set("MessagingServiceSid", from)
	} else {
		form.Set("From", from)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURLs["twilio"]+"/2010-04-01/Accounts/"+url.PathEscape(sid)+"/Messages.json", strings.NewReader(form.Encode()))
	if err != nil {
		return "", permanentErr("twilio request could not be built")
	}
	req.SetBasicAuth(sid, cfg.str("authToken"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	code, body, err := s.doHTTP(req, "twilio")
	if err != nil {
		return "", err
	}
	var out struct {
		Sid     string `json:"sid"`
		Status  any    `json:"status"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &out)
	if code >= 200 && code < 300 && out.Sid != "" {
		return fmt.Sprintf("twilio sid=%s status=%v", out.Sid, out.Status), nil
	}
	reason := fmt.Sprintf("twilio HTTP %d", code)
	if out.Code != 0 || out.Message != "" {
		reason = fmt.Sprintf("twilio HTTP %d code=%d: %s", code, out.Code, out.Message)
	}
	if transientStatus(code) || (code >= 200 && code < 300) {
		return "", transientErr("%s", reason)
	}
	return "", permanentErr("%s", reason)
}

func fromHeader(cfg providerConfig) string {
	addr := mail.Address{Name: cfg.str("fromName"), Address: cfg.str("from")}
	return addr.String()
}

func (s *service) sendResend(ctx context.Context, cfg providerConfig, msg outMessage) (string, error) {
	payload, _ := json.Marshal(map[string]any{"from": fromHeader(cfg), "to": []string{msg.To}, "subject": msg.Subject, "html": msg.HTML, "text": msg.Text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURLs["resend"]+"/emails", bytes.NewReader(payload))
	if err != nil {
		return "", permanentErr("resend request could not be built")
	}
	req.Header.Set("Authorization", "Bearer "+cfg.str("apiKey"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "nakama-publication/1.0")
	if msg.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", msg.IdempotencyKey)
	}
	code, body, err := s.doHTTP(req, "resend")
	if err != nil {
		return "", err
	}
	var out struct {
		ID         string `json:"id"`
		StatusCode int    `json:"statusCode"`
		Name       string `json:"name"`
		Message    string `json:"message"`
	}
	_ = json.Unmarshal(body, &out)
	if code >= 200 && code < 300 && out.ID != "" {
		return "resend id=" + out.ID, nil
	}
	reason := fmt.Sprintf("resend HTTP %d", code)
	if out.Name != "" || out.Message != "" {
		reason = fmt.Sprintf("resend HTTP %d %s: %s", code, out.Name, out.Message)
	}
	if transientStatus(code) || code == http.StatusConflict || (code >= 200 && code < 300) {
		return "", transientErr("%s", reason)
	}
	return "", permanentErr("%s", reason)
}
