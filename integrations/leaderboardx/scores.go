package main

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
)

var boardCache struct {
	sync.Mutex
	at   time.Time
	list []*board
}

func invalidateBoards() {
	boardCache.Lock()
	boardCache.at = time.Time{}
	boardCache.Unlock()
}

func (s *service) activeBoards(ctx context.Context) []*board {
	boardCache.Lock()
	defer boardCache.Unlock()
	if time.Since(boardCache.at) < 10*time.Second {
		return boardCache.list
	}
	list, err := s.boards(ctx, "active")
	if err != nil {
		return boardCache.list
	}
	boardCache.list, boardCache.at = list, time.Now()
	return list
}

type writeResult struct {
	Delta     int64
	Outcome   string
	Score     int64
	Subscore  int64
	Rank      int64
	Duplicate bool
}

func (s *service) addScore(ctx context.Context, b *board, userID, source, sourceID string, value float64, subscore int64) (*writeResult, error) {
	now := s.now()
	if st := b.status(now); st != "active" {
		return nil, apierr.Conflict("LEADERBOARD_NOT_ACTIVE").With("status", st)
	}
	delta := scoreDelta(b.Source, value)
	if source == "client" && b.Source.Calc != "count" {
		delta = int64(math.Round(value))
	}
	if delta < 0 || subscore < 0 {
		return nil, apierr.BadRequest("LEADERBOARD_INVALID_SCORE").WithField("score")
	}
	if delta == 0 && (b.Operator == "incr" || b.Operator == "decr") && subscore == 0 {
		return &writeResult{Outcome: "zero"}, nil
	}
	delta, outcome := clampWrite(b.Limits, delta)
	if outcome == "too_high" {
		s.logRefused(ctx, b, userID, source, value, outcome)
		return nil, apierr.BadRequest("LEADERBOARD_SCORE_TOO_HIGH").WithField("score").With("max", b.Limits.MaxPerWrite)
	}
	if b.Limits.PerMinute > 0 && source != "admin" {
		d, err := clusterkit.Allow(ctx, s.db, "lbx:"+b.ID+":"+userID, int64(b.Limits.PerMinute), time.Minute)
		if err != nil {
			return nil, err
		}
		if !d.Allowed {
			s.logRefused(ctx, b, userID, source, value, "rate_limited")
			return nil, apierr.TooManyRequests("LEADERBOARD_RATE_LIMITED", 60)
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.ExecContext(ctx, `INSERT INTO leaderboardx_score_log (board_id, source_id, user_id, source, value, delta, outcome) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`,
		b.ID, sourceID, userID, source, strconv.FormatFloat(value, 'f', -1, 64), delta, outcome)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		out := &writeResult{Duplicate: true}
		if _, own, _, err := s.core.List(ctx, b.Key, []string{userID}, 1, ""); err == nil && len(own) > 0 {
			out.Score, out.Subscore, out.Rank = own[0].Score, own[0].Subscore, own[0].Rank
		}
		return out, nil
	}
	var username string
	_ = tx.QueryRowContext(ctx, `SELECT COALESCE(username,'') FROM users WHERE id=$1`, userID).Scan(&username)
	rec, err := s.core.Write(ctx, b.Key, userID, username, delta, subscore, nil)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	out := &writeResult{Delta: delta, Outcome: outcome}
	if rec != nil {
		out.Score, out.Subscore, out.Rank = rec.Score, rec.Subscore, rec.Rank
	}
	return out, nil
}

func (s *service) logRefused(ctx context.Context, b *board, userID, source string, value float64, why string) {
	_, _ = s.db.ExecContext(ctx, `INSERT INTO leaderboardx_score_log (board_id, source_id, user_id, source, value, delta, outcome) VALUES ($1,$2,$3,$4,$5,0,$6) ON CONFLICT DO NOTHING`,
		b.ID, "refused:"+uuid.Must(uuid.NewV4()).String(), userID, source, strconv.FormatFloat(value, 'f', -1, 64), why)
}

func (s *service) onEconomyEvent(ctx context.Context, ev playerevents.Recorded) {
	for _, b := range s.activeBoards(ctx) {
		if b.Source.Type != "economy_event" || b.Source.Event != ev.Key {
			continue
		}
		if _, err := s.addScore(ctx, b, ev.UserID, "economy_event", "ev:"+ev.ID, ev.Value, 0); err != nil {
			s.logger.Debug("Leaderboard %s: economy event %s not counted: %v", b.Key, ev.ID, err)
		}
	}
}

func (s *service) onGameEvent(ctx context.Context, ev gameevents.Event) (any, error) {
	if ev.Name == "" || strings.HasPrefix(ev.ID, owner+":") {
		return nil, nil
	}
	for _, b := range s.activeBoards(ctx) {
		if b.Source.Type != "game_event" || b.Source.Event != ev.Name || (b.Source.Context != "" && b.Source.Context != ev.Context) {
			continue
		}
		value := 1.0
		if b.Source.Calc == "value" && b.Source.Field != "" {
			v, ok := number(ev.Fields[b.Source.Field])
			if !ok {
				continue
			}
			value = v
		}
		if _, err := s.addScore(ctx, b, ev.UserID, "game_event", "ge:"+ev.ID, value, 0); err != nil {
			s.logger.Debug("Leaderboard %s: game event %s not counted: %v", b.Key, ev.ID, err)
		}
	}
	return nil, nil
}

func number(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case int32:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case bool:
		if x {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

func (s *service) rpcSubmit(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key       string   `json:"key"`
		Score     *float64 `json:"score"`
		Subscore  float64  `json:"subscore"`
		RequestID string   `json:"requestId"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	key := strings.TrimSpace(in.Key)
	if key == "" {
		return nil, apierr.BadRequest("LEADERBOARD_KEY_REQUIRED").WithField("key")
	}
	b, err := s.boardByKey(ctx, s.db, key)
	if errors.Is(err, errNoBoard) || (err == nil && !b.Active) {
		return nil, apierr.NotFound("LEADERBOARD_NOT_FOUND").WithField("key")
	}
	if err != nil {
		return nil, err
	}
	if !b.Limits.ClientSubmit {
		return nil, apierr.Forbidden("LEADERBOARD_SUBMIT_NOT_ALLOWED").WithField("key")
	}
	if in.Score == nil || math.IsNaN(*in.Score) || math.IsInf(*in.Score, 0) || *in.Score < 0 || *in.Score > 1e15 || in.Subscore < 0 || in.Subscore > 1e15 {
		return nil, apierr.BadRequest("LEADERBOARD_INVALID_SCORE").WithField("score")
	}
	req := strings.TrimSpace(in.RequestID)
	if req == "" {
		if cfg.Scores.RequireRequestID {
			return nil, apierr.BadRequest("REQUEST_ID_REQUIRED").WithField("requestId")
		}
		req = uuid.Must(uuid.NewV4()).String()
	}
	if len(req) > 64 {
		return nil, apierr.BadRequest("INVALID_REQUEST_ID").WithField("requestId")
	}
	if cfg.Scores.PlayerPerMinute > 0 {
		d, err := clusterkit.Allow(ctx, s.db, "lbx:p:"+userID, int64(cfg.Scores.PlayerPerMinute), time.Minute)
		if err != nil {
			return nil, err
		}
		if !d.Allowed {
			return nil, apierr.TooManyRequests("LEADERBOARD_RATE_LIMITED", 60)
		}
	}
	if b.Limits.MaxPerWrite <= 0 && cfg.Scores.ClientMaxPerWrite > 0 {
		capped := *b
		capped.Limits.MaxPerWrite = cfg.Scores.ClientMaxPerWrite
		b = &capped
	}
	res, err := s.addScore(ctx, b, userID, "client", "c:"+userID+":"+req, *in.Score, int64(math.Round(in.Subscore)))
	if err != nil {
		return nil, err
	}
	out := map[string]any{"key": b.Key, "requestId": req, "accepted": !res.Duplicate && res.Outcome != "zero", "clamped": res.Outcome == "clamped", "replayed": res.Duplicate,
		"written": res.Delta, "score": res.Score, "subscore": res.Subscore, "rank": res.Rank, "tier": nil}
	if res.Rank > 0 {
		total := 0
		if b.TierBasis == "percent" {
			total = countPlayers(ctx, s.db, b.Key, expiry(b.ResetSchedule, s.now()))
		}
		if idx := tierIndex(b, int(res.Rank), res.Score, total); idx >= 0 {
			out["tier"] = b.Tiers[idx].Key
		}
	}
	return out, nil
}
