package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/events"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/store"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/insightsapi"
)

var eventNamespace = uuid.Must(uuid.FromString("6f0b6d2e-3b8a-4c39-9a51-6c1f2b7d9e40"))

type sendIn struct {
	BatchID    string       `json:"batchId"`
	SessionID  string       `json:"sessionId"`
	Platform   string       `json:"platform"`
	AppVersion string       `json:"appVersion"`
	Game       string       `json:"game"`
	Events     []events.Raw `json:"events"`
	Lang       string       `json:"lang"`
}

type sendOut struct {
	BatchID   string          `json:"batchId"`
	Duplicate bool            `json:"duplicate"`
	Received  int             `json:"received"`
	Accepted  int             `json:"accepted"`
	Flagged   int             `json:"flagged"`
	Rejected  int             `json:"rejected"`
	Dropped   int             `json:"dropped"`
	Results   []events.Result `json:"results"`
}

func policy(cfg settings.Events) events.Policy {
	return events.Policy{MaxTags: cfg.MaxTags, MaxNameLength: cfg.MaxNameLength, MaxFields: cfg.MaxFields, MaxStringLength: cfg.MaxStringLength,
		ValueMin: cfg.ValueMin, ValueMax: cfg.ValueMax, PastHours: cfg.PastHours, FutureMinutes: cfg.FutureMinutes,
		UnknownEvents: cfg.UnknownEvents, UnknownTagGroups: cfg.UnknownTagGroups, UnknownTagValues: cfg.UnknownTagValues, ValueOutOfRange: cfg.ValueOutOfRange}
}

func eventID(parts ...string) string {
	return uuid.NewV5(eventNamespace, strings.Join(parts, "\x1f")).String()
}

func dayOf(t time.Time, loc *time.Location) string { return t.In(loc).Format(time.DateOnly) }

func (s *service) rpcSend(ctx context.Context, userID, payload string) (any, error) {
	var in sendIn
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	return s.send(ctx, userID, in)
}

func allowN(ctx context.Context, db *sql.DB, key string, n, limit int64, window time.Duration) (bool, time.Duration, error) {
	var count, left int64
	err := db.QueryRowContext(ctx, `INSERT INTO cluster_rate_limit (key, window_start, count, expires_at) VALUES ($1, now(), $3, now() + $2::INT8 * INTERVAL '1 millisecond')
		ON CONFLICT (key) DO UPDATE SET
			window_start = CASE WHEN cluster_rate_limit.expires_at <= now() THEN now() ELSE cluster_rate_limit.window_start END,
			count = CASE WHEN cluster_rate_limit.expires_at <= now() THEN $3 ELSE cluster_rate_limit.count + $3 END,
			expires_at = CASE WHEN cluster_rate_limit.expires_at <= now() THEN EXCLUDED.expires_at ELSE cluster_rate_limit.expires_at END
		RETURNING count, GREATEST(0, CEIL(EXTRACT(EPOCH FROM (expires_at - now())) * 1000))::INT8`, key, window.Milliseconds(), n).Scan(&count, &left)
	if err != nil {
		return false, 0, err
	}
	return count <= limit, time.Duration(left) * time.Millisecond, nil
}

func (s *service) send(ctx context.Context, userID string, in sendIn) (*sendOut, error) {
	cfg := s.settings(ctx)
	ev := cfg.Events
	if !ev.Enabled {
		return nil, apierr.Unavailable("ANALYTICS_DISABLED")
	}
	in.BatchID = strings.TrimSpace(in.BatchID)
	switch {
	case in.BatchID == "":
		return nil, apierr.BadRequest("ANALYTICS_BATCH_ID_REQUIRED").WithField("batchId")
	case !events.BatchIDPattern.MatchString(in.BatchID):
		return nil, apierr.BadRequest("ANALYTICS_INVALID_BATCH_ID").WithField("batchId")
	case len(in.Events) == 0:
		return nil, apierr.BadRequest("ANALYTICS_BATCH_EMPTY").WithField("events")
	case len(in.Events) > ev.MaxBatch:
		return nil, apierr.BadRequest("ANALYTICS_BATCH_TOO_LARGE").WithField("events").With("max", ev.MaxBatch)
	case in.SessionID != "" && !events.SessionPattern.MatchString(in.SessionID):
		return nil, apierr.BadRequest("ANALYTICS_INVALID_SESSION").WithField("sessionId")
	case !events.GamePattern.MatchString(in.Game):
		return nil, apierr.BadRequest("ANALYTICS_INVALID_APP_INFO").WithField("game")
	}
	if out, err := s.receipt(ctx, s.db, userID, in.BatchID); err != nil || out != nil {
		return out, err
	}
	if s.pending.Load() >= int64(cfg.Storage.MaxPendingEvents) {
		return nil, apierr.Unavailable("ANALYTICS_BACKLOG_FULL").Retry(30)
	}
	if ok, wait, err := allowN(ctx, s.db, "ins:b:"+userID, 1, int64(ev.BatchesPerMinute), time.Minute); err != nil {
		return nil, err
	} else if !ok {
		return nil, apierr.TooManyRequests("ANALYTICS_RATE_LIMITED", int(math.Ceil(wait.Seconds())))
	}
	if ok, wait, err := allowN(ctx, s.db, "ins:e:"+userID, int64(len(in.Events)), int64(ev.EventsPerMinute), time.Minute); err != nil {
		return nil, err
	} else if !ok {
		return nil, apierr.TooManyRequests("ANALYTICS_RATE_LIMITED", int(math.Ceil(wait.Seconds())))
	}
	cat, err := s.catalog(ctx, in.Game)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()
	loc := cfg.Location()
	platform, version := cleanLabel(strings.ToLower(in.Platform), 32), cleanLabel(in.AppVersion, 32)
	out := &sendOut{BatchID: in.BatchID, Received: len(in.Events), Results: make([]events.Result, 0, len(in.Events))}
	list := make([]store.Event, 0, len(in.Events))
	pol := policy(ev)
	for i, raw := range in.Events {
		p, res := events.Prepare(i, raw, cat, pol, now)
		out.Results = append(out.Results, res)
		switch res.Status {
		case events.StatusStored, events.StatusFlagged:
			if res.Status == events.StatusFlagged {
				out.Flagged++
			} else {
				out.Accepted++
			}
			list = append(list, buildEvent(p, eventID("c", userID, in.BatchID, strconv.Itoa(i)), in.Game, userID, in.SessionID, platform, version, "client", now, loc))
		case events.StatusDropped:
			out.Dropped++
		default:
			out.Rejected++
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	raw, _ := json.Marshal(out)
	res, err := tx.ExecContext(ctx, `INSERT INTO insights_batch (user_id, batch_id, events, result) VALUES ($1,$2,$3,$4) ON CONFLICT (user_id, batch_id) DO NOTHING`, userID, in.BatchID, len(list), raw)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		_ = tx.Rollback()
		dup, err := s.receipt(ctx, s.db, userID, in.BatchID)
		if err == nil && dup == nil {
			err = apierr.Conflict("ANALYTICS_BUSY").Retry(1)
		}
		return dup, err
	}
	if err := enqueue(ctx, tx, "client", list); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.kick()
	return out, nil
}

func buildEvent(p events.Prepared, id, game, userID, sessionID, platform, version, source string, now time.Time, loc *time.Location) store.Event {
	at := now
	if p.ClientTime != nil {
		at = *p.ClientTime
	}
	return store.Event{ID: id, Game: game, Name: p.Name, Value: p.Value, Tags: p.Tags, Fields: p.Fields, UserID: userID, SessionID: sessionID,
		At: at, Day: dayOf(at, loc), ServerAt: now, Platform: platform, AppVersion: version, Source: source, Flagged: p.Flagged, FlagReason: p.Reason()}
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func enqueue(ctx context.Context, q execer, source string, list []store.Event) error {
	if len(list) == 0 {
		return nil
	}
	raw, err := json.Marshal(list)
	if err != nil {
		return err
	}
	_, err = q.ExecContext(ctx, `INSERT INTO insights_outbox (source, events, count) VALUES ($1, $2, $3)`, source, raw, len(list))
	return err
}

func (s *service) receipt(ctx context.Context, q interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}, userID, batchID string) (*sendOut, error) {
	var raw []byte
	err := q.QueryRowContext(ctx, `SELECT result FROM insights_batch WHERE user_id=$1 AND batch_id=$2`, userID, batchID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out sendOut
	if err := json.Unmarshal(raw, &out); err != nil || out.BatchID == "" {
		return nil, apierr.Conflict("ANALYTICS_BUSY").Retry(1)
	}
	out.Duplicate = true
	return &out, nil
}

func (s *service) serverSink(ctx context.Context, list []insightsapi.Event) (insightsapi.Result, error) {
	cfg := s.settings(ctx)
	var out insightsapi.Result
	if !cfg.Events.Enabled {
		out.Rejected = len(list)
		out.Codes = []string{"ANALYTICS_DISABLED"}
		return out, nil
	}
	if s.pending.Load() >= 2*int64(cfg.Storage.MaxPendingEvents) {
		return out, apierr.Unavailable("ANALYTICS_BACKLOG_FULL")
	}
	now := s.now().UTC()
	loc := cfg.Location()
	pol := policy(cfg.Events)
	cats := map[string]*events.Catalog{}
	stored := make([]store.Event, 0, len(list))
	for i, e := range list {
		if !events.GamePattern.MatchString(e.Game) {
			out.Rejected++
			out.Codes = append(out.Codes, "ANALYTICS_INVALID_APP_INFO")
			continue
		}
		cat, ok := cats[e.Game]
		if !ok {
			c, err := s.catalog(ctx, e.Game)
			if err != nil {
				return out, err
			}
			cats[e.Game], cat = c, c
		}
		raw := events.Raw{Name: e.Name, Path: e.Path, Fields: e.Fields}
		if e.Value != nil {
			raw.Value = *e.Value
		}
		if len(e.Tags) > 0 {
			raw.Tags, _ = json.Marshal(e.Tags)
		}
		if !e.Time.IsZero() {
			raw.Time = e.Time.UTC().Format(time.RFC3339Nano)
		}
		p, res := events.Prepare(i, raw, cat, pol, now)
		if res.Status != events.StatusStored && res.Status != events.StatusFlagged {
			out.Rejected++
			out.Codes = append(out.Codes, res.Code)
			continue
		}
		if res.Status == events.StatusFlagged {
			out.Flagged++
		} else {
			out.Accepted++
		}
		id := eventID("s", uuid.Must(uuid.NewV4()).String())
		if k := strings.TrimSpace(e.Key); k != "" {
			id = eventID("k", k)
		}
		source := strings.TrimSpace(e.Source)
		if source == "" {
			source = "server"
		}
		stored = append(stored, buildEvent(p, id, e.Game, strings.TrimSpace(e.UserID), e.SessionID, cleanLabel(strings.ToLower(e.Platform), 32), cleanLabel(e.AppVersion, 32), cleanLabel(source, 32), now, loc))
	}
	if err := enqueue(ctx, s.db, "server", stored); err != nil {
		return out, err
	}
	s.kick()
	return out, nil
}

func (s *service) gameEvent(ctx context.Context, ev gameevents.Event) (any, error) {
	if !s.settings(ctx).Events.GameEvents {
		return nil, nil
	}
	list := []insightsapi.Event{}
	src := ev.Source
	if src == "" {
		src = "game_event"
	}
	if ev.Name != "" && events.ValidName(ev.Name, 128) {
		e := insightsapi.Event{Key: "ge:" + ev.ID, UserID: ev.UserID, Name: ev.Name, Source: src}
		if c := strings.TrimSpace(ev.Context); c != "" && events.ValidTagID(c) {
			e.Tags = []insightsapi.Tag{{Group: "context", ID: c}}
		}
		fields := map[string]any{}
		for k, v := range ev.Fields {
			if e.Value == nil && (k == "value" || k == "amount" || k == "score") {
				if n, ok := events.Number(v); ok && n != nil {
					e.Value = n
					continue
				}
			}
			switch v.(type) {
			case string, float64, bool, int, int64:
				fields[k] = v
			}
		}
		if len(fields) > 0 && len(fields) <= 16 {
			e.Fields = fields
		}
		list = append(list, e)
	}
	for i, t := range ev.Typed {
		raw := events.Raw{Type: t.Type, ID: t.ID, Status: t.Status, P1: t.P1, P2: t.P2, P3: t.P3, Flow: t.Flow, Currency: t.Currency, ItemType: t.ItemType, Item: t.Item,
			CartType: t.CartType, Store: t.Store, Severity: t.Severity, Message: t.Message}
		if t.Value != nil {
			raw.Value = *t.Value
		}
		if t.Score != nil {
			raw.Score = float64(*t.Score)
		}
		if t.Amount != nil {
			raw.Amount = *t.Amount
		}
		p, res := events.Prepare(i, raw, nil, events.DefaultPolicy(), s.now())
		if res.Status == events.StatusRejected {
			continue
		}
		tags := make([]insightsapi.Tag, 0, len(p.Tags))
		for g, id := range p.Tags {
			tags = append(tags, insightsapi.Tag{Group: g, ID: id})
		}
		list = append(list, insightsapi.Event{Key: "ge:" + ev.ID + ":t" + strconv.Itoa(i), UserID: ev.UserID, Name: p.Name, Tags: tags, Value: p.Value, Fields: p.Fields, Source: src})
	}
	if len(list) == 0 {
		return nil, nil
	}
	_, err := s.serverSink(ctx, list)
	return nil, err
}

func (s *service) sessionEvent(ctx context.Context, userID, game, sessionID, name, platform, version string, seconds *int64) {
	if !s.settings(ctx).Events.SessionEvents {
		return
	}
	if !events.GamePattern.MatchString(game) {
		game = ""
	}
	e := insightsapi.Event{Key: "sess:" + userID + ":" + sessionID + ":" + name, UserID: userID, Game: game, SessionID: sessionID, Name: name, Platform: platform, AppVersion: version, Source: "server"}
	if seconds != nil {
		e.Value = insightsapi.Num(float64(*seconds))
	}
	if _, err := s.serverSink(ctx, []insightsapi.Event{e}); err != nil && s.logger != nil {
		s.logger.Debug("Insights session event %s for %s not queued: %v", name, userID, err)
	}
}
