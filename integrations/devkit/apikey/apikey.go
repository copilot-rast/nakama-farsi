package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/internal/settings"
)

const (
	Prefix     = "nfk_"
	HeaderName = "X-Api-Key"
)

var (
	ScopePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,39}:[a-z][a-z0-9-]{0,39}$`)
	grantPattern = regexp.MustCompile(`^(\*|[a-z][a-z0-9-]{0,39}:(\*|[a-z][a-z0-9-]{0,39}))$`)
	keyPattern   = regexp.MustCompile(`^nfk_([a-z0-9]{12})_([A-Za-z0-9_-]{43})$`)
)

type Key struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type ctxKey struct{}

func FromContext(ctx context.Context) *Key {
	k, _ := ctx.Value(ctxKey{}).(*Key)
	return k
}

func Generate() (full, lookup string, hash []byte, err error) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	raw := make([]byte, 12+32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", nil, err
	}
	b := make([]byte, 12)
	for i := range b {
		b[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	lookup = string(b)
	full = Prefix + lookup + "_" + base64.RawURLEncoding.EncodeToString(raw[12:])
	return full, lookup, Hash(full), nil
}

func Hash(full string) []byte {
	sum := sha256.Sum256([]byte(full))
	return sum[:]
}

func Parse(full string) (string, bool) {
	m := keyPattern.FindStringSubmatch(strings.TrimSpace(full))
	if m == nil {
		return "", false
	}
	return m[1], true
}

func ValidGrant(scope string, allowWildcards bool) bool {
	if allowWildcards {
		return grantPattern.MatchString(scope)
	}
	return ScopePattern.MatchString(scope)
}

func Allows(granted []string, need string, allowWildcards bool) bool {
	if need == "" {
		return true
	}
	res := need
	if i := strings.IndexByte(need, ':'); i > 0 {
		res = need[:i]
	}
	for _, g := range granted {
		if g == need {
			return true
		}
		if allowWildcards && (g == "*" || g == res+":*") {
			return true
		}
	}
	return false
}

func FromRequest(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get(HeaderName)); v != "" {
		return v
	}
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	for _, p := range []string{"Bearer ", "ApiKey "} {
		if strings.HasPrefix(auth, p) && strings.HasPrefix(strings.TrimSpace(auth[len(p):]), Prefix) {
			return strings.TrimSpace(auth[len(p):])
		}
	}
	return ""
}

func ClientIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			return strings.TrimSpace(strings.Split(xff, ",")[0])
		}
		if xr := strings.TrimSpace(r.Header.Get("X-Real-Ip")); xr != "" {
			return xr
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func ipAllowed(list []string, ip string) bool {
	if len(list) == 0 {
		return true
	}
	addr := net.ParseIP(ip)
	for _, entry := range list {
		if strings.Contains(entry, "/") {
			if _, n, err := net.ParseCIDR(entry); err == nil && addr != nil && n.Contains(addr) {
				return true
			}
		} else if entry == ip {
			return true
		}
	}
	return false
}

func Verify(ctx context.Context, db *sql.DB, full, need, ip string) (*Key, error) {
	if strings.TrimSpace(full) == "" {
		return nil, apierr.Unauthorized("API_KEY_REQUIRED")
	}
	lookup, ok := Parse(full)
	if !ok {
		return nil, apierr.Unauthorized("API_KEY_INVALID")
	}
	cfg, _ := settings.Load(ctx, db)
	var k Key
	var hash []byte
	var scopesRaw, ipsRaw string
	var expires, revoked, lastUsed sql.NullTime
	err := db.QueryRowContext(ctx, `SELECT id::text, name, secret_hash, array_to_json(scopes)::text, array_to_json(allowed_ips)::text, expires_at, revoked_at, last_used_at FROM devkit_api_key WHERE lookup=$1`, lookup).
		Scan(&k.ID, &k.Name, &hash, &scopesRaw, &ipsRaw, &expires, &revoked, &lastUsed)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apierr.Unauthorized("API_KEY_INVALID")
	}
	if err != nil {
		return nil, err
	}
	scopes, ips := []string{}, []string{}
	_ = json.Unmarshal([]byte(scopesRaw), &scopes)
	_ = json.Unmarshal([]byte(ipsRaw), &ips)
	if subtle.ConstantTimeCompare(hash, Hash(strings.TrimSpace(full))) != 1 {
		return nil, apierr.Unauthorized("API_KEY_INVALID")
	}
	now := time.Now()
	if revoked.Valid {
		return nil, apierr.Unauthorized("API_KEY_REVOKED")
	}
	if expires.Valid && !expires.Time.After(now) {
		return nil, apierr.Unauthorized("API_KEY_EXPIRED").With("expiredAt", expires.Time.UTC().Format(time.RFC3339))
	}
	if !ipAllowed(ips, ip) {
		return nil, apierr.Forbidden("API_KEY_IP_DENIED").With("ip", ip)
	}
	if !Allows(scopes, need, cfg.APIKeys.AllowWildcards) {
		return nil, apierr.Forbidden("API_KEY_SCOPE_DENIED").With("scope", need)
	}
	k.Scopes = scopes
	if expires.Valid {
		t := expires.Time
		k.ExpiresAt = &t
	}
	every := time.Duration(cfg.APIKeys.LastUsedEverySeconds) * time.Second
	if !lastUsed.Valid || now.Sub(lastUsed.Time) >= every {
		_, _ = db.ExecContext(ctx, `UPDATE devkit_api_key SET last_used_at=now(), last_used_ip=$2, use_count=use_count+1 WHERE id=$1`, k.ID, ip)
	} else {
		_, _ = db.ExecContext(ctx, `UPDATE devkit_api_key SET use_count=use_count+1 WHERE id=$1`, k.ID)
	}
	return &k, nil
}

func Require(db *sql.DB, scope string, next func(w http.ResponseWriter, r *http.Request, key *Key)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, _ := settings.Load(r.Context(), db)
		key, err := Verify(r.Context(), db, FromRequest(r), scope, ClientIP(r, cfg.APIKeys.TrustProxyHeaders))
		if err != nil {
			apierr.WriteHTTP(w, r, db, err)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, key)), key)
	}
}

func RequireRPC(ctx context.Context, db *sql.DB, scope string) (*Key, error) {
	headers, _ := ctx.Value(runtime.RUNTIME_CTX_HEADERS).(map[string][]string)
	r := &http.Request{Header: http.Header{}}
	for k, v := range headers {
		for _, x := range v {
			r.Header.Add(k, x)
		}
	}
	ip, _ := ctx.Value(runtime.RUNTIME_CTX_CLIENT_IP).(string)
	return Verify(ctx, db, FromRequest(r), scope, ip)
}

func RegisterScopes(ctx context.Context, db *sql.DB, source string, scopes map[string]string) error {
	for name, desc := range scopes {
		if !ScopePattern.MatchString(name) {
			return errors.New("invalid scope " + name)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO devkit_api_scope (name, description, source) VALUES ($1,$2,$3) ON CONFLICT (name) DO NOTHING`, name, desc, source); err != nil {
			return err
		}
	}
	return nil
}
