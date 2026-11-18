package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/heroiclabs/nakama/v3/integrations/cup/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/cuphooks"
)

var errDuplicate = errors.New("cup result already counted")

func (s *service) report(ctx context.Context, r cuphooks.Result) error {
	var cupID string
	err := s.db.QueryRowContext(ctx, `SELECT cup_id::text FROM cup_game WHERE game_match_id=$1`, r.GameMatchID).Scan(&cupID)
	if errors.Is(err, sql.ErrNoRows) && uuidPattern.MatchString(r.MatchID) && r.GameNo > 0 {
		err = s.db.QueryRowContext(ctx, `SELECT cup_id::text FROM cup_game WHERE match_id=$1::uuid AND game_no=$2`, r.MatchID, r.GameNo).Scan(&cupID)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return apierr.NotFound("CUP_GAME_NOT_FOUND")
	}
	if err != nil {
		return err
	}
	_, err = s.withCup(ctx, cupID, false, func(tx *sql.Tx, b *bracket) error {
		var g *gameRow
		for _, list := range b.games {
			for _, x := range list {
				if x.GameMatchID == r.GameMatchID || (x.MatchID == r.MatchID && x.GameNo == r.GameNo && r.GameNo > 0) {
					g = x
				}
			}
		}
		if g == nil {
			return apierr.NotFound("CUP_GAME_NOT_FOUND")
		}
		if g.Status == "finished" || g.Status == "void" {
			return errDuplicate
		}
		if g.GameMatchID == "" {
			g.GameMatchID = r.GameMatchID
		}
		m := b.matchByID(g.MatchID)
		if m == nil {
			return apierr.NotFound("CUP_MATCH_NOT_FOUND")
		}
		if b.cup.Status != "running" {
			g.Status, g.Reason, g.FinishedAt = "void", "cup_"+b.cup.Status, ptr(b.now)
			b.newGame = append(b.newGame, g)
			return nil
		}
		b.applyResult(m, g, r, s.rnd)
		b.advance()
		return nil
	})
	if errors.Is(err, errDuplicate) {
		return nil
	}
	return err
}

func (s *service) rpcReport(ctx context.Context, userID, payload string) (any, error) {
	if userID != "" {
		return nil, apierr.Forbidden("CUP_SERVER_ONLY")
	}
	var in struct {
		GameMatchID string         `json:"gameMatchId"`
		MatchID     string         `json:"matchId"`
		GameNo      int            `json:"gameNo"`
		WinnerID    string         `json:"winnerId"`
		Draw        bool           `json:"draw"`
		Void        bool           `json:"void"`
		Present     []string       `json:"present"`
		Score       map[string]int `json:"score"`
		Reason      string         `json:"reason"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	if in.GameMatchID == "" {
		return nil, apierr.BadRequest("CUP_GAME_ID_REQUIRED").WithField("gameMatchId")
	}
	err := s.report(ctx, cuphooks.Result{GameMatchID: in.GameMatchID, MatchID: in.MatchID, GameNo: in.GameNo, Winner: in.WinnerID, Draw: in.Draw, Void: in.Void,
		Present: in.Present, Score: in.Score, Reason: in.Reason})
	if err != nil {
		return nil, err
	}
	return map[string]any{"accepted": true}, nil
}

func (b *bracket) activeMatchOf(user string) *matchRow {
	var best *matchRow
	for _, m := range b.matches {
		if (m.A == user || m.B == user) && !done(m.Status) && m.Status != stPending {
			if best == nil || m.Round < best.Round {
				best = m
			}
		}
	}
	return best
}

func (s *service) ready(ctx context.Context, userID, cupID string) (map[string]any, error) {
	var out map[string]any
	_, err := s.withCup(ctx, cupID, false, func(tx *sql.Tx, b *bracket) error {
		if b.cup.Status != "running" {
			return apierr.Conflict("CUP_NOT_RUNNING")
		}
		p := b.players[userID]
		if p == nil || p.Seed == 0 {
			return apierr.Forbidden("CUP_NOT_A_PLAYER")
		}
		m := b.activeMatchOf(userID)
		if m == nil {
			if p.Status == "playing" {
				return apierr.Conflict("CUP_WAITING_FOR_OPPONENT")
			}
			return apierr.Conflict("CUP_NO_MATCH")
		}
		mode := b.cup.Rules.Timing.Mode
		switch m.Status {
		case stWaiting:
			return apierr.Conflict("CUP_CHECK_IN_NOT_OPEN").With("opensAt", iso(m.OpensAt))
		case stPlaying:
			out = map[string]any{"matchId": m.ID, "status": m.Status, "gameNo": m.Games, "gameMatchId": m.GameMatchID, "launched": false}
			return nil
		}
		opp := b.other(m, userID)
		wasReady := (userID == m.A && b.fresh(m.ReadyA)) || (userID == m.B && b.fresh(m.ReadyB))
		now := ptr(b.now)
		if userID == m.A {
			m.ReadyA, m.ShowedA = now, true
		} else {
			m.ReadyB, m.ShowedB = now, true
		}
		b.touch(m)
		checkIn := mode == settings.ModeLive && m.Games == 0
		if !wasReady {
			b.notify([]string{opp}, "cup_opponent_ready", map[string]any{"matchId": m.ID, "round": m.Round, "opponentId": userID, "checkIn": checkIn}, false, false)
		}
		oppReady := (opp == m.A && b.fresh(m.ReadyA)) || (opp == m.B && b.fresh(m.ReadyB))
		switch {
		case checkIn:
			if m.StartsAt != nil && !b.now.Before(*m.StartsAt) {
				b.tickMatch(m, s.rnd)
			}
		case oppReady:
			b.launch(m)
		}
		b.advance()
		out = map[string]any{"matchId": m.ID, "status": m.Status, "gameNo": m.Games, "gameMatchId": "", "launched": m.Status == stPlaying}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if id, _ := out["matchId"].(string); id != "" {
		var gid string
		_ = s.db.QueryRowContext(ctx, `SELECT game_match_id FROM cup_match WHERE id=$1::uuid`, id).Scan(&gid)
		out["gameMatchId"] = gid
	}
	return out, nil
}

func (s *service) rpcReady(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		CupID string `json:"cupId"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	if !uuidPattern.MatchString(in.CupID) {
		return nil, apierr.BadRequest("CUP_ID_REQUIRED").WithField("cupId")
	}
	out, err := s.ready(ctx, userID, in.CupID)
	if err != nil {
		return nil, err
	}
	return readyJSON(out), nil
}

func (s *service) rpcStart(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		CupID string `json:"cupId"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	if !uuidPattern.MatchString(in.CupID) {
		return nil, apierr.BadRequest("CUP_ID_REQUIRED").WithField("cupId")
	}
	out, err := s.ready(ctx, userID, in.CupID)
	if err != nil {
		return nil, err
	}
	return readyJSON(out), nil
}

func readyJSON(out map[string]any) map[string]any {
	matchID, _ := out["matchId"].(string)
	status, _ := out["status"].(string)
	gameNo, _ := out["gameNo"].(int)
	gameMatchID, _ := out["gameMatchId"].(string)
	launched, _ := out["launched"].(bool)
	return map[string]any{"matchId": matchID, "status": status, "gameNo": gameNo, "gameMatchId": gameMatchID, "launched": launched}
}
