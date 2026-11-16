package main

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/settings"
)

var sessionIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,64}$`)

type sessionIn struct {
	Event      string   `json:"event"`
	SessionID  string   `json:"sessionId"`
	AppVersion string   `json:"appVersion"`
	Platform   string   `json:"platform"`
	Device     string   `json:"device"`
	OS         string   `json:"os"`
	Seconds    *float64 `json:"seconds"`
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func cleanLabel(v string, max int) string {
	v = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, strings.TrimSpace(v))
	if r := []rune(v); len(r) > max {
		v = string(r[:max])
	}
	return v
}

func (s *service) rpcSession(ctx context.Context, userID, payload string) (any, error) {
	var in sessionIn
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	if cfg.Sessions.Source != settings.SourceClient {
		return map[string]any{"sessionId": in.SessionID, "ignored": true, "source": cfg.Sessions.Source}, nil
	}
	return s.session(ctx, userID, in, settings.SourceClient)
}

func (s *service) session(ctx context.Context, userID string, in sessionIn, source string) (map[string]any, error) {
	cfg := s.settings(ctx)
	in.Event = strings.ToLower(strings.TrimSpace(in.Event))
	in.SessionID = strings.TrimSpace(in.SessionID)
	switch {
	case in.Event != "start" && in.Event != "ping" && in.Event != "end":
		return nil, apierr.BadRequest("ANALYTICS_INVALID_SESSION").WithField("event")
	case !sessionIDPattern.MatchString(in.SessionID):
		return nil, apierr.BadRequest("ANALYTICS_INVALID_SESSION").WithField("sessionId")
	case in.Seconds != nil && (math.IsNaN(*in.Seconds) || *in.Seconds < 0):
		return nil, apierr.BadRequest("ANALYTICS_INVALID_SESSION").WithField("seconds")
	}
	in.AppVersion, in.Platform = cleanLabel(in.AppVersion, 32), cleanLabel(in.Platform, 32)
	in.Device, in.OS = cleanLabel(in.Device, 64), cleanLabel(in.OS, 64)
	now := s.now().UTC()
	day := now.In(cfg.Location()).Format(time.DateOnly)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var id string
	var started time.Time
	var ended sql.NullTime
	var secs sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT id::text, started_at, ended_at, seconds FROM insights_session WHERE user_id=$1 AND client_id=$2 FOR UPDATE`, userID, in.SessionID).Scan(&id, &started, &ended, &secs)
	created := false
	switch {
	case errors.Is(err, sql.ErrNoRows):
		var recent int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM insights_session WHERE user_id=$1 AND started_at > $2`, userID, now.Add(-time.Hour)).Scan(&recent); err != nil {
			return nil, err
		}
		if recent >= cfg.Sessions.StartsPerHour {
			return nil, apierr.TooManyRequests("ANALYTICS_RATE_LIMITED", 60).With("limit", cfg.Sessions.StartsPerHour)
		}
		if cfg.Sessions.CloseOtherOpen {
			if err := s.closeOpen(ctx, tx, userID, cfg); err != nil {
				return nil, err
			}
		}
		started = now
		if in.Event == "end" && in.Seconds != nil {
			started = now.Add(-time.Duration(math.Min(*in.Seconds, float64(cfg.Sessions.MaxSeconds)) * float64(time.Second)))
		}
		id = uuid.Must(uuid.NewV4()).String()
		if _, err := tx.ExecContext(ctx, `INSERT INTO insights_session (id, user_id, client_id, source, day, started_at, last_seen_at, app_version, platform, device, os)
			VALUES ($1,$2,$3,$4,$5::date,$6,$7,$8,$9,$10,$11)`, id, userID, in.SessionID, source, day, started, now, in.AppVersion, in.Platform, in.Device, in.OS); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO insights_player (user_id, first_day, first_at, last_day, app_version, platform) VALUES ($1,$2::date,$3,$2::date,$4,$5)
			ON CONFLICT (user_id) DO UPDATE SET last_day=GREATEST(insights_player.last_day, EXCLUDED.last_day),
			app_version=CASE WHEN EXCLUDED.app_version<>'' THEN EXCLUDED.app_version ELSE insights_player.app_version END,
			platform=CASE WHEN EXCLUDED.platform<>'' THEN EXCLUDED.platform ELSE insights_player.platform END`, userID, day, now, in.AppVersion, in.Platform); err != nil {
			return nil, err
		}
		if err := touchDay(ctx, tx, day, userID, 1, in.AppVersion, in.Platform); err != nil {
			return nil, err
		}
		created = true
	case err != nil:
		return nil, err
	case ended.Valid:
		out := map[string]any{"sessionId": in.SessionID, "status": "ended", "duplicate": true, "startedAt": started}
		if secs.Valid {
			out["seconds"] = secs.Int64
		}
		return out, tx.Commit()
	default:
		ping := 0
		if in.Event == "ping" {
			ping = 1
		}
		if _, err := tx.ExecContext(ctx, `UPDATE insights_session SET last_seen_at=GREATEST(last_seen_at,$2), pings=pings+$3,
			app_version=CASE WHEN app_version='' THEN $4 ELSE app_version END, platform=CASE WHEN platform='' THEN $5 ELSE platform END,
			device=CASE WHEN device='' THEN $6 ELSE device END, os=CASE WHEN os='' THEN $7 ELSE os END WHERE id=$1`, id, now, ping, in.AppVersion, in.Platform, in.Device, in.OS); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE insights_player SET last_day=GREATEST(last_day,$2::date) WHERE user_id=$1`, userID, day); err != nil {
			return nil, err
		}
		if err := touchDay(ctx, tx, day, userID, 0, in.AppVersion, in.Platform); err != nil {
			return nil, err
		}
	}
	out := map[string]any{"sessionId": in.SessionID, "status": "open", "created": created, "startedAt": started,
		"pingSeconds": cfg.Sessions.PingSeconds, "timeoutSeconds": cfg.Sessions.TimeoutSeconds}
	if in.Event == "end" {
		n, err := s.closeSession(ctx, tx, id, now, in.Seconds, cfg)
		if err != nil {
			return nil, err
		}
		out["status"], out["seconds"] = "ended", n
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func touchDay(ctx context.Context, q queryer, day, userID string, sessions int, version, platform string) error {
	_, err := q.ExecContext(ctx, `INSERT INTO insights_player_day (day, user_id, sessions, app_version, platform) VALUES ($1::date,$2,$3,$4,$5)
		ON CONFLICT (day, user_id) DO UPDATE SET sessions=insights_player_day.sessions+EXCLUDED.sessions,
		app_version=CASE WHEN EXCLUDED.app_version<>'' THEN EXCLUDED.app_version ELSE insights_player_day.app_version END,
		platform=CASE WHEN EXCLUDED.platform<>'' THEN EXCLUDED.platform ELSE insights_player_day.platform END`, day, userID, sessions, version, platform)
	return err
}

func (s *service) closeSession(ctx context.Context, q queryer, id string, end time.Time, client *float64, cfg settings.All) (int64, error) {
	var started, last time.Time
	if err := q.QueryRowContext(ctx, `SELECT started_at, last_seen_at FROM insights_session WHERE id=$1 AND ended_at IS NULL`, id).Scan(&started, &last); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	if end.Before(last) {
		end = last
	}
	wall := end.Sub(started).Seconds()
	secs := wall
	if client != nil && *client <= wall+60 {
		secs = *client
	}
	secs = math.Max(0, math.Min(secs, float64(cfg.Sessions.MaxSeconds)))
	n := int64(math.Round(secs))
	var userID, day string
	if err := q.QueryRowContext(ctx, `UPDATE insights_session SET ended_at=$2, last_seen_at=GREATEST(last_seen_at,$2), seconds=$3 WHERE id=$1 AND ended_at IS NULL RETURNING user_id::text, day::text`, id, end, n).Scan(&userID, &day); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	_, err := q.ExecContext(ctx, `UPDATE insights_player_day SET seconds=seconds+$3 WHERE day=$1::date AND user_id=$2`, day, userID, n)
	return n, err
}

func (s *service) closeOpen(ctx context.Context, q queryer, userID string, cfg settings.All) error {
	rows, err := q.QueryContext(ctx, `SELECT id::text, last_seen_at FROM insights_session WHERE user_id=$1 AND ended_at IS NULL FOR UPDATE`, userID)
	if err != nil {
		return err
	}
	type open struct {
		id   string
		last time.Time
	}
	list := []open{}
	for rows.Next() {
		var o open
		if err := rows.Scan(&o.id, &o.last); err != nil {
			rows.Close()
			return err
		}
		list = append(list, o)
	}
	rows.Close()
	for _, o := range list {
		if _, err := s.closeSession(ctx, q, o.id, o.last, nil, cfg); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) closeIdle(ctx context.Context, cfg settings.All) (int, error) {
	cut := s.now().UTC().Add(-time.Duration(cfg.Sessions.TimeoutSeconds) * time.Second)
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, last_seen_at FROM insights_session WHERE ended_at IS NULL AND last_seen_at < $1 ORDER BY last_seen_at LIMIT 2000`, cut)
	if err != nil {
		return 0, err
	}
	type open struct {
		id   string
		last time.Time
	}
	list := []open{}
	for rows.Next() {
		var o open
		if rows.Scan(&o.id, &o.last) == nil {
			list = append(list, o)
		}
	}
	rows.Close()
	closed := 0
	for _, o := range list {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return closed, err
		}
		if _, err := tx.ExecContext(ctx, `SELECT 1 FROM insights_session WHERE id=$1 FOR UPDATE`, o.id); err != nil {
			_ = tx.Rollback()
			return closed, err
		}
		if _, err := s.closeSession(ctx, tx, o.id, o.last, nil, cfg); err != nil {
			_ = tx.Rollback()
			return closed, err
		}
		if err := tx.Commit(); err != nil {
			return closed, err
		}
		closed++
	}
	return closed, nil
}
