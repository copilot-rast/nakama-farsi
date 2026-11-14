package voice

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const KindLiveKit = "livekit"

func init() {
	Register(Spec{
		Kind: KindLiveKit,
		Fields: []Field{
			{Name: "url", Kind: "wsurl", Required: true},
			{Name: "apiKey", Kind: "text", Required: true},
			{Name: "apiSecret", Kind: "text", Secret: true, Required: true},
		},
		New: func(cfg Config, client *http.Client) (Provider, error) {
			p := &liveKit{url: cfg.Get("url"), key: cfg.Get("apiKey"), secret: cfg.Get("apiSecret"), client: client, now: time.Now}
			if p.url == "" || p.key == "" || p.secret == "" {
				return nil, &ProviderError{Code: "VOICE_NOT_CONFIGURED"}
			}
			return p, nil
		},
	})
}

type liveKit struct {
	url    string
	key    string
	secret string
	client *http.Client
	now    func() time.Time
}

type liveKitVideo struct {
	RoomJoin       bool   `json:"roomJoin,omitempty"`
	RoomList       bool   `json:"roomList,omitempty"`
	Room           string `json:"room,omitempty"`
	CanPublish     *bool  `json:"canPublish,omitempty"`
	CanSubscribe   *bool  `json:"canSubscribe,omitempty"`
	CanPublishData *bool  `json:"canPublishData,omitempty"`
}

type liveKitClaims struct {
	Issuer    string       `json:"iss"`
	Subject   string       `json:"sub,omitempty"`
	Name      string       `json:"name,omitempty"`
	NotBefore int64        `json:"nbf"`
	IssuedAt  int64        `json:"iat"`
	Expires   int64        `json:"exp"`
	ID        string       `json:"jti"`
	Video     liveKitVideo `json:"video"`
	Metadata  string       `json:"metadata,omitempty"`
}

func boolPtr(b bool) *bool { return &b }

func signHS256(claims any, secret string) (string, error) {
	head := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := head + "." + base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func randomID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (p *liveKit) Token(_ context.Context, req Request) (Grant, error) {
	ttl := req.TTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	now := p.now()
	exp := now.Add(ttl)
	meta := ""
	if req.GameKey != "" {
		b, _ := json.Marshal(map[string]string{"gameKey": req.GameKey})
		meta = string(b)
	}
	claims := liveKitClaims{
		Issuer: p.key, Subject: req.Identity, Name: req.Name, NotBefore: now.Add(-10 * time.Second).Unix(), IssuedAt: now.Unix(), Expires: exp.Unix(), ID: randomID(),
		Video:    liveKitVideo{RoomJoin: true, Room: req.Room, CanPublish: boolPtr(req.CanPublish), CanSubscribe: boolPtr(true), CanPublishData: boolPtr(false)},
		Metadata: meta,
	}
	token, err := signHS256(claims, p.secret)
	if err != nil {
		return Grant{}, err
	}
	return Grant{Provider: KindLiveKit, URL: p.clientURL(), Token: token, Room: req.Room, Identity: req.Identity, ExpiresAt: time.Unix(exp.Unix(), 0).UTC()}, nil
}

func (p *liveKit) clientURL() string {
	u, err := url.Parse(p.url)
	if err != nil {
		return p.url
	}
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}
	return strings.TrimRight(u.String(), "/")
}

func (p *liveKit) apiURL() string {
	u, err := url.Parse(p.url)
	if err != nil {
		return p.url
	}
	switch u.Scheme {
	case "ws":
		u.Scheme = "http"
	case "wss":
		u.Scheme = "https"
	}
	return strings.TrimRight(u.String(), "/")
}

func (p *liveKit) Test(ctx context.Context) error {
	now := p.now()
	token, err := signHS256(liveKitClaims{Issuer: p.key, NotBefore: now.Add(-10 * time.Second).Unix(), IssuedAt: now.Unix(), Expires: now.Add(time.Minute).Unix(), ID: randomID(), Video: liveKitVideo{RoomList: true}}, p.secret)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL()+"/twirp/livekit.RoomService/ListRooms", bytes.NewReader([]byte("{}")))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := p.client.Do(req)
	if err != nil {
		return errors.Join(ErrUnavailable, err)
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 1<<16))
	switch {
	case res.StatusCode == http.StatusOK:
		return nil
	case res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden:
		return &ProviderError{Code: "VOICE_BAD_CREDENTIALS", Status: res.StatusCode}
	default:
		return &ProviderError{Code: "VOICE_PROVIDER_ERROR", Status: res.StatusCode}
	}
}
