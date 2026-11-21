package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/events"
	imports "github.com/heroiclabs/nakama/v3/integrations/insights/internal/importer"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/store"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const connectorLease = 5 * time.Minute

type connectorView struct {
	ID           string          `json:"id"`
	Kind         string          `json:"kind"`
	Name         string          `json:"name"`
	Enabled      bool            `json:"enabled"`
	Config       json.RawMessage `json:"config"`
	HasSecret    bool            `json:"hasSecret"`
	EveryMinutes int             `json:"everyMinutes"`
	Cursor       json.RawMessage `json:"cursor"`
	NextRunAt    time.Time       `json:"nextRunAt"`
	Running      bool            `json:"running"`
	LastRunAt    *time.Time      `json:"lastRunAt"`
	LastStatus   string          `json:"lastStatus"`
	LastError    string          `json:"lastError"`
	LastCount    int64           `json:"lastCount"`
	UpdatedAt    time.Time       `json:"updatedAt"`
}

const connectorCols = `id::text, kind, name, enabled, config, secret <> '', every_minutes, cursor, next_run_at, coalesce(claimed_until > now(), false), last_run_at, last_status, last_error, last_count, updated_at`

func scanConnector(sc interface{ Scan(...any) error }) (connectorView, error) {
	var c connectorView
	var last sql.NullTime
	var cfg, cur []byte
	err := sc.Scan(&c.ID, &c.Kind, &c.Name, &c.Enabled, &cfg, &c.HasSecret, &c.EveryMinutes, &cur, &c.NextRunAt, &c.Running, &last, &c.LastStatus, &c.LastError, &c.LastCount, &c.UpdatedAt)
	if last.Valid {
		c.LastRunAt = &last.Time
	}
	c.Config, c.Cursor = cfg, cur
	return c, err
}

func (s *service) connectorList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+connectorCols+` FROM insights_connector ORDER BY name LIMIT 200`)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []connectorView{}
	for rows.Next() {
		c, err := scanConnector(rows)
		if err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, c)
	}
	kinds := []string{}
	for k := range imports.All {
		kinds = append(kinds, k)
	}
	sortStrings(kinds)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "kinds": kinds, "secretKey": s.secretKey != nil})
}

func (s *service) connectorSave(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Kind         string          `json:"kind"`
		Name         string          `json:"name"`
		Enabled      bool            `json:"enabled"`
		Config       json.RawMessage `json:"config"`
		Secret       *string         `json:"secret"`
		EveryMinutes int             `json:"everyMinutes"`
		ResetCursor  bool            `json:"resetCursor"`
	}
	if err := pluginkit.ReadBody(r, &in, 256*1024); err != nil {
		s.fail(w, err)
		return
	}
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var oldSecret, oldKind string
	if id != "" {
		if _, err := uuid.FromString(id); err != nil {
			s.fail(w, pluginkit.Missing("NOT_FOUND"))
			return
		}
		if err := s.db.QueryRowContext(ctx, `SELECT kind, secret FROM insights_connector WHERE id=$1`, id).Scan(&oldKind, &oldSecret); errors.Is(err, sql.ErrNoRows) {
			s.fail(w, pluginkit.Missing("NOT_FOUND"))
			return
		} else if err != nil {
			s.fail(w, err)
			return
		}
		if in.Kind == "" {
			in.Kind = oldKind
		}
	}
	imp, ok := imports.All[in.Kind]
	if !ok || (oldKind != "" && oldKind != in.Kind) {
		s.fail(w, pluginkit.Bad("INVALID_KIND", "kind", ""))
		return
	}
	if in.Name = cleanLabel(in.Name, 120); in.Name == "" {
		s.fail(w, pluginkit.Bad("REQUIRED", "name", ""))
		return
	}
	if in.EveryMinutes == 0 {
		in.EveryMinutes = 60
	}
	if in.EveryMinutes < 5 || in.EveryMinutes > 7*24*60 {
		s.fail(w, pluginkit.Bad("VALUE_OUT_OF_RANGE", "everyMinutes", ""))
		return
	}
	sealed := oldSecret
	plain := ""
	if in.Secret != nil {
		plain = *in.Secret
		if plain == "" {
			sealed = ""
		} else {
			if s.secretKey == nil {
				s.fail(w, pluginkit.Bad("SECRET_KEY_MISSING", "secret", "ECONOMY_SECRET_KEY is not set"))
				return
			}
			var err error
			if sealed, err = connectorBox.Seal(s.secretKey, []byte(plain)); err != nil {
				s.fail(w, err)
				return
			}
		}
	}
	if len(in.Config) == 0 {
		in.Config = json.RawMessage(`{}`)
	}
	cfg, err := imp.Check(in.Config, plain)
	if err != nil {
		var ce *imports.ConfigError
		if errors.As(err, &ce) {
			s.fail(w, pluginkit.Bad(ce.Code, ce.Field, ""))
			return
		}
		s.fail(w, err)
		return
	}
	if in.Enabled && sealed == "" {
		s.fail(w, pluginkit.Bad("REQUIRED", "secret", ""))
		return
	}
	if id == "" {
		id = uuid.Must(uuid.NewV4()).String()
		_, err = s.db.ExecContext(ctx, `INSERT INTO insights_connector (id, kind, name, enabled, config, secret, every_minutes) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			id, in.Kind, in.Name, in.Enabled, cfg, sealed, in.EveryMinutes)
	} else {
		reset := ""
		if in.ResetCursor {
			reset = `, cursor='{}'`
		}
		_, err = s.db.ExecContext(ctx, `UPDATE insights_connector SET name=$2, enabled=$3, config=$4, secret=$5, every_minutes=$6, updated_at=now()`+reset+` WHERE id=$1`,
			id, in.Name, in.Enabled, cfg, sealed, in.EveryMinutes)
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	c, err := scanConnector(s.db.QueryRowContext(ctx, `SELECT `+connectorCols+` FROM insights_connector WHERE id=$1`, id))
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, c)
}

func (s *service) connectorDelete(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM insights_connector WHERE id=$1`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) connectorRunNow(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `UPDATE insights_connector SET next_run_at=now() WHERE id=$1`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	select {
	case s.importWake <- struct{}{}:
	default:
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"queued": true})
}

func (s *service) connectorRuns(w http.ResponseWriter, r *http.Request) {
	id := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT id::text, started_at, finished_at, status, read, written, skipped, error, cursor_from, cursor_to FROM insights_import_run WHERE connector_id=$1 ORDER BY started_at DESC LIMIT 50`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var rid, status, msg string
		var started time.Time
		var finished sql.NullTime
		var read, written, skipped int64
		var from, to []byte
		if err := rows.Scan(&rid, &started, &finished, &status, &read, &written, &skipped, &msg, &from, &to); err != nil {
			s.fail(w, err)
			return
		}
		it := map[string]any{"id": rid, "startedAt": started, "finishedAt": nil, "status": status, "read": read, "written": written, "skipped": skipped, "error": msg,
			"cursorFrom": json.RawMessage(from), "cursorTo": json.RawMessage(to)}
		if finished.Valid {
			it["finishedAt"] = finished.Time
		}
		items = append(items, it)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *service) connectorLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-s.importWake:
		}
		for i := 0; i < 5; i++ {
			ran, err := s.runDueConnector(ctx, holder)
			if err != nil && s.logger != nil {
				s.logger.Warn("Insights import failed: %v", err)
			}
			if !ran {
				break
			}
		}
	}
}

func (s *service) runDueConnector(ctx context.Context, holder string) (bool, error) {
	var id, kind string
	var cfg, cursor []byte
	var sealed string
	var every int
	token := holder + ":" + claimToken()
	err := s.db.QueryRowContext(ctx, `UPDATE insights_connector SET claimed_until = now() + $2::int * INTERVAL '1 second', claimed_by = $1
		WHERE id = (SELECT id FROM insights_connector WHERE enabled AND next_run_at <= now() AND (claimed_until IS NULL OR claimed_until < now()) ORDER BY next_run_at LIMIT 1 FOR UPDATE SKIP LOCKED)
		RETURNING id::text, kind, config, secret, cursor, every_minutes`, token, int(connectorLease.Seconds())).Scan(&id, &kind, &cfg, &sealed, &cursor, &every)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, s.runConnector(ctx, id, kind, cfg, sealed, cursor, every, token)
}

func (s *service) runConnector(ctx context.Context, id, kind string, cfg []byte, sealed string, cursor []byte, every int, token string) error {
	runID := uuid.Must(uuid.NewV4()).String()
	_, _ = s.db.ExecContext(ctx, `INSERT INTO insights_import_run (id, connector_id, cursor_from) VALUES ($1,$2,$3)`, runID, id, cursor)
	var written int64
	finish := func(out imports.Outcome, runErr error) error {
		status, msg := "ok", ""
		if runErr != nil {
			status, msg = "error", runErr.Error()
			if len(msg) > 1000 {
				msg = msg[:1000]
			}
		}
		next := "now() + $4::int * INTERVAL '1 minute'"
		if out.More && runErr == nil {
			next = "now() + INTERVAL '5 seconds'"
		} else if runErr != nil {
			next = "now() + LEAST($4::int, 15) * INTERVAL '1 minute'"
		}
		cur := out.Cursor
		if len(cur) == 0 {
			cur = cursor
		}
		_, err := s.db.ExecContext(ctx, `UPDATE insights_connector SET cursor=$3, next_run_at=`+next+`, claimed_until=NULL, claimed_by='', last_run_at=now(), last_status=$5, last_error=$6, last_count=$7
			WHERE id=$1 AND claimed_by=$2`, id, token, cur, every, status, msg, written)
		_, _ = s.db.ExecContext(ctx, `UPDATE insights_import_run SET finished_at=now(), status=$2, read=$3, written=$4, skipped=$5, error=$6, cursor_to=$7 WHERE id=$1`,
			runID, status, out.Read, written, out.Skipped, msg, cur)
		if err != nil {
			return err
		}
		return runErr
	}
	imp, ok := imports.All[kind]
	if !ok {
		return finish(imports.Outcome{}, errors.New("unknown importer kind "+kind))
	}
	secret := ""
	if sealed != "" {
		if s.secretKey == nil {
			return finish(imports.Outcome{}, errors.New("ECONOMY_SECRET_KEY is not set"))
		}
		b, err := connectorBox.Open(s.secretKey, sealed)
		if err != nil {
			return finish(imports.Outcome{}, err)
		}
		secret = string(b)
	}
	st := s.settings(ctx)
	loc := st.Location()
	pol := policy(st.Events)
	pol.UnknownEvents, pol.UnknownTagGroups, pol.UnknownTagValues = events.PolicyAccept, events.PolicyAccept, events.PolicyFlag
	cats := map[string]*events.Catalog{}
	source := "import:" + kind
	env := imports.Env{HTTP: s.httpClient, Now: s.now().UTC(), Config: cfg, Secret: secret, Cursor: cursor, Max: 200000,
		Emit: func(ctx context.Context, items []imports.Item) error {
			cat, ok := cats[""]
			if !ok {
				c, err := s.catalog(ctx, "")
				if err != nil {
					return err
				}
				cats[""], cat = c, c
			}
			now := s.now().UTC()
			list := make([]store.Event, 0, len(items))
			for i, it := range items {
				p, res := events.Prepare(i, it.Raw, cat, pol, now)
				if res.Status != events.StatusStored && res.Status != events.StatusFlagged {
					continue
				}
				at := it.At
				p.ClientTime = &at
				list = append(list, buildEvent(p, eventID("i", id, it.Key), "", it.UserID, it.SessionID, cleanLabel(it.Platform, 32), cleanLabel(it.AppVersion, 32), source, now, loc))
			}
			for len(list) > 0 {
				n := len(list)
				if n > 500 {
					n = 500
				}
				if err := enqueue(ctx, s.db, "import", list[:n]); err != nil {
					return err
				}
				written += int64(n)
				list = list[n:]
			}
			s.kick()
			return nil
		},
		Renew: func(ctx context.Context) error {
			res, err := s.db.ExecContext(ctx, `UPDATE insights_connector SET claimed_until = now() + $3::int * INTERVAL '1 second' WHERE id=$1 AND claimed_by=$2`, id, token, int(connectorLease.Seconds()))
			if err != nil {
				return err
			}
			if n, _ := res.RowsAffected(); n == 0 {
				return errors.New("import lease lost")
			}
			return nil
		},
	}
	rctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	out, err := imp.Pull(rctx, env)
	return finish(out, err)
}

func sortStrings(v []string) {
	for i := 1; i < len(v); i++ {
		for j := i; j > 0 && strings.Compare(v[j-1], v[j]) > 0; j-- {
			v[j-1], v[j] = v[j], v[j-1]
		}
	}
}
