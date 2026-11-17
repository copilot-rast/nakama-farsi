package cuphooks

import (
	"context"
	"errors"
	"sort"
	"sync"
)

type Launch struct {
	CupID       string
	MatchID     string
	GameNo      int
	Game        string
	Players     []string
	Options     map[string]any
	Rated       bool
	Draw        string
	Final       bool
	JoinSeconds int
}

type Launched struct {
	GameMatchID string
	Info        map[string]any
}

type Result struct {
	GameMatchID string
	MatchID     string
	GameNo      int
	Winner      string
	Draw        bool
	Void        bool
	Present     []string
	Score       map[string]int
	Reason      string
}

type Game struct {
	Key    string
	Name   map[string]string
	Launch func(ctx context.Context, l Launch) (Launched, error)
}

type Reporter func(ctx context.Context, r Result) error

var (
	ErrUnavailable = errors.New("cups are not available: the cup plugin is not loaded")
	ErrNoGame      = errors.New("no game registered for this cup")

	mu       sync.RWMutex
	games    = map[string]Game{}
	reporter Reporter
)

func Register(g Game) {
	if g.Key == "" || g.Launch == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	games[g.Key] = g
}

func Unregister(key string) {
	mu.Lock()
	defer mu.Unlock()
	delete(games, key)
}

func Lookup(key string) (Game, bool) {
	mu.RLock()
	defer mu.RUnlock()
	g, ok := games[key]
	return g, ok
}

func Games() []Game {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Game, 0, len(games))
	for _, g := range games {
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func SetReporter(fn Reporter) {
	mu.Lock()
	defer mu.Unlock()
	reporter = fn
}

func Available() bool {
	mu.RLock()
	defer mu.RUnlock()
	return reporter != nil
}

func Report(ctx context.Context, r Result) error {
	mu.RLock()
	fn := reporter
	mu.RUnlock()
	if fn == nil {
		return ErrUnavailable
	}
	if r.GameMatchID == "" {
		return errors.New("a cup result needs the game match id")
	}
	return fn(ctx, r)
}
