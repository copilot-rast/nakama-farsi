package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const KindComms = "comms"

func init() {
	Register(Spec{
		Kind: KindComms,
		Fields: []Field{
			{Name: "url", Kind: "url", Required: true},
			{Name: "publicUrl", Kind: "url"},
			{Name: "serverKey", Kind: "text", Secret: true, Required: true},
			{Name: "appId", Kind: "text"},
			{Name: "tokenIssuer", Kind: "text"},
			{Name: "tokenKeyId", Kind: "text"},
			{Name: "tokenSecret", Kind: "text", Secret: true},
		},
		Check: checkUserTokenFields,
		New: func(cfg Config, client *http.Client) (Provider, error) {
			p := &comms{base: strings.TrimRight(cfg.Get("url"), "/"), public: strings.TrimRight(cfg.Get("publicUrl"), "/"), key: cfg.Get("serverKey"), client: client}
			if p.base == "" || p.key == "" {
				return nil, &ProviderError{Code: "VOICE_NOT_CONFIGURED"}
			}
			if signer, err := NewUserTokenSigner(cfg); err == nil {
				p.user = &signer
			}
			if p.public == "" {
				p.public = p.base
			}
			return p, nil
		},
	})
}

type comms struct {
	base   string
	public string
	key    string
	user   *UserTokenSigner
	client *http.Client
}

func PublicURL(cfg Config) string {
	if v := strings.TrimRight(cfg.Get("publicUrl"), "/"); v != "" {
		return v
	}
	return strings.TrimRight(cfg.Get("url"), "/")
}

type commsError struct {
	status int
	code   string
}

func (e *commsError) Error() string { return "comms " + e.code }

func (c *comms) call(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.key)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.client.Do(req)
	if err != nil {
		return errors.Join(ErrUnavailable, err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		if out != nil && len(raw) > 0 {
			return json.Unmarshal(raw, out)
		}
		return nil
	}
	var e struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(raw, &e)
	return &commsError{status: res.StatusCode, code: e.Error.Code}
}

func roomPath(room string) string { return "/v1/server/game-rooms/" + url.PathEscape(room) }

func (c *comms) open(ctx context.Context, req Request) error {
	players := req.Players
	seen := false
	for _, p := range players {
		if p.ID == req.Identity {
			seen = true
		}
	}
	if !seen {
		players = append(players, Player{ID: req.Identity, Name: req.Name})
	}
	body := map[string]any{"match_id": req.Room, "players": players, "video": false}
	if req.TTL > 0 {
		body["ttl_seconds"] = int(req.TTL.Seconds())
	}
	err := c.call(ctx, http.MethodPost, "/v1/server/game-rooms", body, nil)
	var ce *commsError
	if errors.As(err, &ce) && ce.status == http.StatusBadRequest && body["ttl_seconds"] != nil {
		delete(body, "ttl_seconds")
		err = c.call(ctx, http.MethodPost, "/v1/server/game-rooms", body, nil)
	}
	return err
}

func (c *comms) token(ctx context.Context, req Request) (string, time.Time, error) {
	var out struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	err := c.call(ctx, http.MethodPost, roomPath(req.Room)+"/players/"+url.PathEscape(req.Identity)+"/token", nil, &out)
	return out.Token, out.ExpiresAt, err
}

func (c *comms) Token(ctx context.Context, req Request) (Grant, error) {
	tok, exp, err := c.token(ctx, req)
	var ce *commsError
	if errors.As(err, &ce) {
		switch {
		case ce.status == http.StatusNotFound || ce.code == "room_closed":
			if oerr := c.open(ctx, req); oerr != nil {
				return Grant{}, mapComms(oerr)
			}
			tok, exp, err = c.token(ctx, req)
		case ce.code == "not_listed":
			if perr := c.call(ctx, http.MethodPut, roomPath(req.Room)+"/players/"+url.PathEscape(req.Identity), map[string]any{"name": req.Name}, nil); perr != nil {
				return Grant{}, mapComms(perr)
			}
			tok, exp, err = c.token(ctx, req)
		}
	}
	if err != nil {
		return Grant{}, mapComms(err)
	}
	if tok == "" {
		return Grant{}, &ProviderError{Code: "VOICE_PROVIDER_ERROR", Detail: "empty token"}
	}
	if exp.IsZero() {
		exp = time.Now().Add(10 * time.Minute)
	}
	return Grant{Provider: KindComms, URL: c.public, Token: tok, Room: req.Room, Identity: req.Identity, ExpiresAt: exp.UTC()}, nil
}

func mapComms(err error) error {
	var ce *commsError
	if errors.As(err, &ce) {
		switch {
		case ce.status == http.StatusUnauthorized || ce.status == http.StatusForbidden:
			return &ProviderError{Code: "VOICE_BAD_CREDENTIALS", Status: ce.status}
		case ce.code == "room_full":
			return &ProviderError{Code: "VOICE_ROOM_FULL", Status: ce.status}
		case ce.status == http.StatusTooManyRequests:
			return &ProviderError{Code: "VOICE_PROVIDER_BUSY", Status: ce.status}
		case ce.status == http.StatusServiceUnavailable:
			return &ProviderError{Code: "VOICE_PROVIDER_UNAVAILABLE", Status: ce.status, Detail: ce.code}
		}
		return &ProviderError{Code: "VOICE_PROVIDER_ERROR", Status: ce.status, Detail: ce.code}
	}
	return err
}

func (c *comms) Test(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+"/healthz", nil)
	if err != nil {
		return err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return errors.Join(ErrUnavailable, err)
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 1<<16))
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return &ProviderError{Code: "VOICE_PROVIDER_UNAVAILABLE", Status: res.StatusCode}
	}
	err = c.call(ctx, http.MethodGet, roomPath("nakama-test-"+randomID()), nil, nil)
	var ce *commsError
	if errors.As(err, &ce) && ce.status == http.StatusNotFound {
		err = nil
	}
	if err != nil {
		return mapComms(err)
	}
	if c.user == nil {
		return nil
	}
	token, _, err := c.user.Sign(UserTokenTestSubject, "Nakama connection test", time.Minute, time.Now())
	if err != nil {
		return err
	}
	return ExchangeUserToken(ctx, c.client, c.base, token)
}
