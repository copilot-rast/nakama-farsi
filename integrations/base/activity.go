package main

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/hookchain"
	"github.com/heroiclabs/nakama/v3/internal/restrict"
)

const activityJSON = `jsonb_build_object('id',a.id,'userId',a.user_id,'username',u.username,'displayName',u.display_name,'kind',a.kind,'action',a.action,'refId',a.ref_id,'data',a.data,
	'actor',a.actor,'ip',a.ip,'country',a.country,'platform',a.platform,'createdAt',a.created_at)`

func activityWhere(r *http.Request) (*sqlWhere, error) {
	q := r.URL.Query()
	where := &sqlWhere{}
	if v := strings.TrimSpace(q.Get("userId")); v != "" {
		if _, err := uuid.FromString(v); err != nil {
			return nil, badRequest("INVALID_ID", "userId", "player id must be a UUID")
		}
		where.add("a.user_id = ?::uuid", v)
	}
	if v := strings.TrimSpace(q.Get("kind")); v != "" {
		kinds := []string{}
		for _, k := range strings.Split(v, ",") {
			if activity.ValidKind(k) {
				kinds = append(kinds, k)
			}
		}
		if len(kinds) > 0 {
			where.add("a.kind = ANY(?)", kinds)
		}
	}
	if v := strings.TrimSpace(q.Get("action")); v != "" {
		where.add("a.action = ?", v)
	}
	if v := strings.TrimSpace(q.Get("country")); len(v) == 2 {
		where.add("a.country = ?", strings.ToUpper(v))
	}
	if v := strings.TrimSpace(q.Get("platform")); v != "" {
		where.add("a.platform = ?", v)
	}
	for _, b := range []struct{ name, op string }{{"from", ">="}, {"to", "<"}} {
		if v := strings.TrimSpace(q.Get(b.name)); v != "" {
			at, err := time.Parse(time.RFC3339, v)
			if err != nil {
				if d, derr := time.Parse("2006-01-02", v); derr == nil {
					at = d
					if b.name == "to" {
						at = at.AddDate(0, 0, 1)
					}
				} else {
					return nil, badRequest("INVALID_DATE", b.name, "use a date like 2026-11-12")
				}
			}
			where.add("a.created_at "+b.op+" ?", at)
		}
	}
	if v := strings.TrimSpace(q.Get("q")); v != "" {
		n := where.next()
		where.args = append(where.args, likePattern(v))
		where.parts = append(where.parts, "(u.username ILIKE "+n+" OR a.ref_id ILIKE "+n+" OR a.ip ILIKE "+n+" OR a.action ILIKE "+n+" OR a.data::text ILIKE "+n+" OR a.actor ILIKE "+n+")")
	}
	return where, nil
}

func (s *service) activityList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	p.Q = ""
	where, err := activityWhere(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	out, err := s.listRaw(r, ` FROM base_activity a JOIN users u ON u.id=a.user_id`, activityJSON, "a.created_at DESC, a.id DESC", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) activitySummary(w http.ResponseWriter, r *http.Request) {
	where, err := activityWhere(r)
	if err != nil {
		s.writeError(w, err)
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT a.kind, count(*) FROM base_activity a JOIN users u ON u.id=a.user_id`+where.sql()+` GROUP BY a.kind`, where.args...)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	counts := map[string]int64{}
	for _, k := range activity.Kinds {
		counts[k] = 0
	}
	for rows.Next() {
		var k string
		var n int64
		if rows.Scan(&k, &n) == nil {
			counts[k] = n
		}
	}
	st, err := loadSupportSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"kinds": activity.Kinds, "counts": counts, "retentionDays": st.ActivityRetentionDays})
}

var playerActivityKinds = []string{activity.KindSignIn, activity.KindPurchase, activity.KindPrize, activity.KindEvent, activity.KindTicket, activity.KindChat, activity.KindReferral, activity.KindRestriction}

func (s *service) rpcMyActivity(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Limit  int    `json:"limit"`
		Kind   string `json:"kind"`
		Before int64  `json:"before"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if in.Limit < 1 || in.Limit > 50 {
		in.Limit = 20
	}
	kinds := playerActivityKinds
	if in.Kind != "" {
		if !oneOfList(in.Kind, playerActivityKinds) {
			return nil, badRequest("INVALID_KIND", "kind", "unknown activity kind")
		}
		kinds = []string{in.Kind}
	}
	args := []any{userID, kinds, in.Limit + 1}
	cond := ""
	if in.Before > 0 {
		cond = ` AND id < $4`
		args = append(args, in.Before)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, kind, action, ref_id, data::text, platform, created_at FROM base_activity WHERE user_id=$1 AND kind = ANY($2)`+cond+` ORDER BY id DESC LIMIT $3`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id int64
		var kind, action, ref, data, platform string
		var at time.Time
		if err := rows.Scan(&id, &kind, &action, &ref, &data, &platform, &at); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"id": id, "kind": kind, "action": action, "refId": ref, "data": rawJSON(data), "platform": platform, "createdAt": at.UTC().Format(time.RFC3339)})
	}
	var next any
	if len(items) > in.Limit {
		items = items[:in.Limit]
		next = items[len(items)-1]["id"]
	}
	return map[string]any{"items": items, "before": next}, rows.Err()
}

func (s *service) sweepActivity(ctx context.Context) (int64, error) {
	st, err := loadSupportSettings(ctx, s.db)
	if err != nil {
		return 0, err
	}
	var total int64
	for i := 0; i < 20; i++ {
		res, err := s.db.ExecContext(ctx, `DELETE FROM base_activity WHERE id IN (SELECT id FROM base_activity WHERE created_at < now() - make_interval(days => $1) ORDER BY id LIMIT 5000)`, st.ActivityRetentionDays)
		if err != nil {
			return total, err
		}
		n, _ := res.RowsAffected()
		total += n
		if n < 5000 {
			break
		}
	}
	return total, nil
}

func (s *service) signIn(ctx context.Context, session *api.Session, method string) {
	if session == nil {
		return
	}
	uid := restrict.TokenUserID(session.Token)
	if uid == "" {
		return
	}
	e := activity.FromContext(ctx, activity.Entry{UserID: uid, Kind: activity.KindSignIn, Action: method, Actor: "player", Data: map[string]any{"created": session.Created}})
	if e.Country == "" {
		e.Data["accountCountry"] = s.userCountry(ctx, uid)
	}
	s.record(ctx, e)
}

func (s *service) registerActivityHooks(initializer runtime.Initializer) error {
	steps := []func() error{
		func() error {
			return initializer.RegisterAfterAuthenticateDevice(hookchain.After("AuthenticateDevice", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateDeviceRequest) error {
				s.signIn(ctx, out, "device")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateCustom(hookchain.After("AuthenticateCustom", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateCustomRequest) error {
				s.signIn(ctx, out, "custom")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateEmail(hookchain.After("AuthenticateEmail", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateEmailRequest) error {
				s.signIn(ctx, out, "email")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateGoogle(hookchain.After("AuthenticateGoogle", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateGoogleRequest) error {
				s.signIn(ctx, out, "google")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateApple(hookchain.After("AuthenticateApple", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateAppleRequest) error {
				s.signIn(ctx, out, "apple")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateFacebook(hookchain.After("AuthenticateFacebook", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateFacebookRequest) error {
				s.signIn(ctx, out, "facebook")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateSteam(hookchain.After("AuthenticateSteam", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateSteamRequest) error {
				s.signIn(ctx, out, "steam")
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateGameCenter(hookchain.After("AuthenticateGameCenter", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateGameCenterRequest) error {
				s.signIn(ctx, out, "gamecenter")
				return nil
			}))
		},
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}
