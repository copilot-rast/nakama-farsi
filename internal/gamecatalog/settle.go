package gamecatalog

import (
	"context"
	"encoding/json"
	"time"
)

const (
	ResultWon      = "won"
	ResultLost     = "lost"
	ResultForfeit  = "forfeit"
	ResultTimeout  = "timeout"
	ResultRefunded = "refunded"
)

func ValidResult(s string) bool {
	switch s {
	case ResultWon, ResultLost, ResultForfeit, ResultTimeout, ResultRefunded:
		return true
	}
	return false
}

type Result struct {
	UserID string `json:"userId"`
	Rank   int    `json:"rank"`
	Status string `json:"status"`
}

type Settlement struct {
	MatchID  string          `json:"matchId"`
	Results  []Result        `json:"results"`
	Reason   string          `json:"reason,omitempty"`
	ForcedBy string          `json:"forcedBy,omitempty"`
	Note     string          `json:"note,omitempty"`
	Summary  json.RawMessage `json:"summary,omitempty"`
}

type ActiveMatch struct {
	MatchID   string    `json:"matchId"`
	Game      string    `json:"game"`
	Room      string    `json:"room"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type BotQuery struct {
	Game     string
	Room     string
	Count    int
	MinLevel int
	Exclude  []string
	Fee      *Fee
}

type Bot struct {
	UserID   string            `json:"userId"`
	Username string            `json:"username"`
	Name     string            `json:"name"`
	Avatar   string            `json:"avatarUrl"`
	Level    int               `json:"level"`
	Template string            `json:"template"`
	Look     map[string]string `json:"look"`
}

type Action struct {
	UserID string         `json:"userId,omitempty"`
	Kind   string         `json:"kind"`
	Data   map[string]any `json:"data,omitempty"`
	At     time.Time      `json:"at"`
}

func Settle(ctx context.Context, st Settlement) ([]Payout, error) {
	p, err := current()
	if err != nil {
		return nil, err
	}
	return p.Settle(ctx, st)
}

func Void(ctx context.Context, matchID, reason, by string) (int, error) {
	p, err := current()
	if err != nil {
		return 0, err
	}
	return p.Void(ctx, matchID, reason, by)
}

func Release(ctx context.Context, matchID, userID string) error {
	p, err := current()
	if err != nil {
		return err
	}
	return p.Release(ctx, matchID, userID)
}

func Active(ctx context.Context, userID string) (*ActiveMatch, error) {
	p, err := current()
	if err != nil {
		return nil, err
	}
	return p.Active(ctx, userID)
}

func Bots(ctx context.Context, q BotQuery) ([]Bot, error) {
	p, err := current()
	if err != nil {
		return nil, err
	}
	return p.Bots(ctx, q)
}

func Log(ctx context.Context, matchID string, actions []Action) error {
	if len(actions) == 0 {
		return nil
	}
	p, err := current()
	if err != nil {
		return err
	}
	return p.Log(ctx, matchID, actions)
}

func PotPayout(fees int64, percent int) int64 {
	if fees <= 0 || percent <= 0 {
		return 0
	}
	if percent > 100 {
		percent = 100
	}
	return fees * int64(percent) / 100
}
