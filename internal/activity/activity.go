package activity

import (
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	KindSignIn      = "signin"
	KindPurchase    = "purchase"
	KindPrize       = "prize"
	KindEvent       = "event"
	KindRestriction = "restriction"
	KindReport      = "report"
	KindTicket      = "ticket"
	KindChat        = "chat"
	KindReferral    = "referral"
	KindConfig      = "config"
	KindItem        = "item"
)

var Kinds = []string{KindSignIn, KindPurchase, KindPrize, KindEvent, KindRestriction, KindReport, KindTicket, KindChat, KindReferral, KindConfig, KindItem}

type Entry struct {
	UserID   string
	Kind     string
	Action   string
	RefID    string
	Data     map[string]any
	Actor    string
	IP       string
	Country  string
	Platform string
}

type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func ValidKind(kind string) bool {
	for _, k := range Kinds {
		if k == kind {
			return true
		}
	}
	return false
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	r := []rune(s)
	for len(string(r)) > n {
		r = r[:len(r)-1]
	}
	return string(r)
}

func Record(ctx context.Context, db Execer, e Entry) error {
	if _, err := uuid.FromString(e.UserID); err != nil {
		return err
	}
	if !ValidKind(e.Kind) || strings.TrimSpace(e.Action) == "" {
		return nil
	}
	data := []byte("{}")
	if len(e.Data) > 0 {
		raw, err := json.Marshal(e.Data)
		if err == nil && len(raw) <= 4096 {
			data = raw
		}
	}
	_, err := db.ExecContext(ctx, `INSERT INTO base_activity (user_id, kind, action, ref_id, data, actor, ip, country, platform) VALUES ($1,$2,$3,$4,$5::jsonb,$6,$7,$8,$9)`,
		e.UserID, e.Kind, clip(e.Action, 48), clip(e.RefID, 64), string(data), clip(e.Actor, 128), clip(e.IP, 64), strings.ToUpper(clip(e.Country, 2)), clip(e.Platform, 16))
	return err
}

func Safe(ctx context.Context, db Execer, logger runtime.Logger, e Entry) {
	if err := Record(ctx, db, e); err != nil && logger != nil {
		logger.Warn("Activity %s/%s for %s not recorded: %v", e.Kind, e.Action, e.UserID, err)
	}
}

var countryHeaders = []string{"cf-ipcountry", "x-country-code", "x-geo-country", "cloudfront-viewer-country"}

func ClientIP(ctx context.Context) string {
	ip, _ := ctx.Value(runtime.RUNTIME_CTX_CLIENT_IP).(string)
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	return strings.TrimSpace(ip)
}

func headers(ctx context.Context) map[string][]string {
	h, _ := ctx.Value(runtime.RUNTIME_CTX_HEADERS).(map[string][]string)
	return h
}

func CountryFromHeaders(h map[string][]string) string {
	for k, v := range h {
		for _, name := range countryHeaders {
			if strings.EqualFold(k, name) && len(v) > 0 {
				c := strings.ToUpper(strings.TrimSpace(v[0]))
				if len(c) == 2 && c != "XX" && c != "T1" {
					return c
				}
			}
		}
	}
	return ""
}

func PlatformFromHeaders(h map[string][]string) string {
	for k, v := range h {
		if strings.EqualFold(k, "x-client-platform") && len(v) > 0 {
			p := strings.ToLower(strings.TrimSpace(v[0]))
			switch p {
			case "android", "ios", "web", "windows", "macos", "linux":
				return p
			}
		}
	}
	for k, v := range h {
		if !strings.EqualFold(k, "user-agent") || len(v) == 0 {
			continue
		}
		ua := strings.ToLower(v[0])
		switch {
		case strings.Contains(ua, "android"):
			return "android"
		case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
			return "ios"
		case strings.Contains(ua, "mozilla"):
			return "web"
		}
	}
	return ""
}

func FromContext(ctx context.Context, e Entry) Entry {
	h := headers(ctx)
	if e.IP == "" {
		e.IP = ClientIP(ctx)
	}
	if e.Country == "" {
		e.Country = CountryFromHeaders(h)
	}
	if e.Platform == "" {
		e.Platform = PlatformFromHeaders(h)
	}
	return e
}

func CountryFromRequest(r *http.Request) string {
	return CountryFromHeaders(r.Header)
}
