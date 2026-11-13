package gamecatalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Definition struct {
	Key      string            `json:"key"`
	Name     map[string]string `json:"name,omitempty"`
	Plugin   string            `json:"plugin,omitempty"`
	Realtime bool              `json:"realtime"`
	Schema   json.RawMessage   `json:"schema,omitempty"`
}

type Fee struct {
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
}

type Room struct {
	Key          string            `json:"key"`
	Name         map[string]string `json:"name,omitempty"`
	EntryFee     *Fee              `json:"entryFee,omitempty"`
	PrizeID      string            `json:"prizeId,omitempty"`
	PrizeGroupID string            `json:"prizeGroupId,omitempty"`
	MinLevel     int               `json:"minLevel,omitempty"`
}

type Game struct {
	ID            string            `json:"id"`
	Key           string            `json:"key"`
	Active        bool              `json:"isActive"`
	Name          map[string]string `json:"name,omitempty"`
	MinPlayers    int               `json:"minPlayers"`
	MaxPlayers    int               `json:"maxPlayers"`
	TurnSeconds   int               `json:"turnSeconds,omitempty"`
	MatchSeconds  int               `json:"matchSeconds,omitempty"`
	BotsAllowed   bool              `json:"botsAllowed"`
	LeaderboardID string            `json:"leaderboardId,omitempty"`
	Rooms         []Room            `json:"rooms"`
	Settings      map[string]any    `json:"settings"`
	Version       int64             `json:"version"`
}

func (g *Game) Room(key string) (*Room, bool) {
	if g == nil {
		return nil, false
	}
	for i := range g.Rooms {
		if g.Rooms[i].Key == key {
			return &g.Rooms[i], true
		}
	}
	return nil, false
}

func (g *Game) Int(name string, fallback int) int {
	if g == nil {
		return fallback
	}
	switch v := g.Settings[name].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		if n, err := v.Int64(); err == nil {
			return int(n)
		}
	}
	return fallback
}

func (g *Game) Bool(name string, fallback bool) bool {
	if g == nil {
		return fallback
	}
	if v, ok := g.Settings[name].(bool); ok {
		return v
	}
	return fallback
}

func (g *Game) String(name, fallback string) string {
	if g == nil {
		return fallback
	}
	if v, ok := g.Settings[name].(string); ok {
		return v
	}
	return fallback
}

type Entry struct {
	MatchID string
	Game    string
	Room    string
	UserID  string
}

type EntryResult struct {
	Status   string `json:"status"`
	Charged  bool   `json:"charged"`
	Fee      *Fee   `json:"fee,omitempty"`
	Players  int    `json:"players"`
	Rejoined bool   `json:"rejoined"`
}

type Payout struct {
	UserID  string           `json:"userId"`
	PrizeID string           `json:"prizeId,omitempty"`
	Rewards []map[string]any `json:"rewards"`
	Error   string           `json:"error,omitempty"`
}

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Code + ": " + e.Message
	}
	return e.Code
}

func Code(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

type Provider interface {
	Game(ctx context.Context, key string) (*Game, error)
	Enter(ctx context.Context, e Entry) (EntryResult, error)
	Leave(ctx context.Context, matchID, userID string) (bool, error)
	Start(ctx context.Context, matchID string) error
	Cancel(ctx context.Context, matchID string) (int, error)
	Finish(ctx context.Context, matchID string, winners []string) ([]Payout, error)
}

var (
	ErrUnavailable = errors.New("game settings are not available: the economy plugin is not loaded")
	ErrNotFound    = &Error{Code: "GAME_NOT_FOUND", Message: "this game is not configured in the console"}
	ErrInvalidKey  = errors.New("a game key is 2 to 64 lower-case letters, digits or underscores and starts with a letter")
	keyPattern     = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)
	camelPattern   = regexp.MustCompile(`^[a-z][a-zA-Z0-9_]{0,63}$`)
	mu             sync.RWMutex
	defs           = map[string]Definition{}
	provider       Provider
)

func ValidKey(key string) bool {
	return keyPattern.MatchString(key)
}

func Register(d Definition) error {
	d.Key = strings.TrimSpace(d.Key)
	if !ValidKey(d.Key) {
		return fmt.Errorf("game %q: %w", d.Key, ErrInvalidKey)
	}
	if len(d.Schema) > 0 {
		if _, err := ParseSchema(d.Schema); err != nil {
			return fmt.Errorf("game %s schema: %w", d.Key, err)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	defs[d.Key] = d
	return nil
}

func Unregister(key string) {
	mu.Lock()
	defer mu.Unlock()
	delete(defs, key)
}

func Registered() []Definition {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Definition, 0, len(defs))
	for _, d := range defs {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func Lookup(key string) (Definition, bool) {
	mu.RLock()
	defer mu.RUnlock()
	d, ok := defs[key]
	return d, ok
}

func SetProvider(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	provider = p
}

func current() (Provider, error) {
	mu.RLock()
	defer mu.RUnlock()
	if provider == nil {
		return nil, ErrUnavailable
	}
	return provider, nil
}

func Settings(ctx context.Context, key string) (*Game, error) {
	p, err := current()
	if err != nil {
		return nil, err
	}
	return p.Game(ctx, key)
}

func Enter(ctx context.Context, e Entry) (EntryResult, error) {
	p, err := current()
	if err != nil {
		return EntryResult{}, err
	}
	return p.Enter(ctx, e)
}

func Leave(ctx context.Context, matchID, userID string) (bool, error) {
	p, err := current()
	if err != nil {
		return false, err
	}
	return p.Leave(ctx, matchID, userID)
}

func Start(ctx context.Context, matchID string) error {
	p, err := current()
	if err != nil {
		return err
	}
	return p.Start(ctx, matchID)
}

func Cancel(ctx context.Context, matchID string) (int, error) {
	p, err := current()
	if err != nil {
		return 0, err
	}
	return p.Cancel(ctx, matchID)
}

func Finish(ctx context.Context, matchID string, winners []string) ([]Payout, error) {
	p, err := current()
	if err != nil {
		return nil, err
	}
	return p.Finish(ctx, matchID, winners)
}
