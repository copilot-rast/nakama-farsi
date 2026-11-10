package restrict

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	KindSuspend = "suspend"
	KindMute    = "mute"
	KindWarn    = "warn"

	CodeSuspended = "ACCOUNT_SUSPENDED"
	CodeMuted     = "CHAT_MUTED"

	NotifySuspended = 1102
	NotifyMuted     = 1103
	NotifyWarned    = 1104
	NotifyLifted    = 1105

	grpcPermissionDenied = 7
)

type Restriction struct {
	ID     string
	UserID string
	Kind   string
	EndsAt time.Time
	Reason localized.Text
}

type RowQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func Active(ctx context.Context, q RowQueryer, userID, kind string) (*Restriction, error) {
	if userID == "" {
		return nil, nil
	}
	r := &Restriction{UserID: userID, Kind: kind}
	var reason string
	err := q.QueryRowContext(ctx, `SELECT id::text, ends_at, reason::text FROM base_restriction
		WHERE user_id = $1::uuid AND kind = $2 AND status = 'active' AND ends_at > now() ORDER BY ends_at DESC LIMIT 1`, userID, kind).Scan(&r.ID, &r.EndsAt, &reason)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.Reason = localized.Parse(reason)
	return r, nil
}

type Notice struct {
	Code   string `json:"code"`
	Kind   string `json:"kind"`
	Until  string `json:"until"`
	Reason string `json:"reason"`
	Lang   string `json:"lang"`
}

func (r *Restriction) Notice(ctx context.Context, db localized.DB, requested string) Notice {
	n := Notice{Kind: r.Kind, Until: r.EndsAt.UTC().Format(time.RFC3339)}
	switch r.Kind {
	case KindSuspend:
		n.Code = CodeSuspended
	case KindMute:
		n.Code = CodeMuted
	}
	langs, lang, err := localized.ForUser(ctx, db, r.UserID, requested)
	if err == nil {
		n.Lang = lang
		n.Reason = langs.Resolve(r.Reason, lang)
	}
	return n
}

func (n Notice) JSON() string {
	b, _ := json.Marshal(n)
	return string(b)
}

func (n Notice) Error() error {
	return runtime.NewError(n.JSON(), grpcPermissionDenied)
}

func (n Notice) Map() map[string]any {
	return map[string]any{"code": n.Code, "kind": n.Kind, "until": n.Until, "reason": n.Reason, "lang": n.Lang}
}

func RefuseSuspended(ctx context.Context, db localized.DB, userID, requested string) error {
	r, err := Active(ctx, db, userID, KindSuspend)
	if err != nil || r == nil {
		return err
	}
	return r.Notice(ctx, db, requested).Error()
}

func TokenUserID(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		UID string `json:"uid"`
	}
	if json.Unmarshal(raw, &claims) != nil {
		return ""
	}
	return claims.UID
}
