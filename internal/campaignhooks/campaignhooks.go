package campaignhooks

import (
	"context"
	"errors"
	"sort"
	"sync"
)

type Completion struct {
	AttemptID      string
	StageKey       string
	Success        bool
	ElapsedSeconds float64
	Fields         map[string]any
}

type (
	Starter func(ctx context.Context, userID, stageKey string) (attemptID string, err error)

	Completer func(ctx context.Context, userID string, c Completion) (map[string]any, error)

	Owner func(ctx context.Context, stageKey string) (ref string, owned bool)
)

var (
	ErrUnavailable = errors.New("campaign stages are not available: the campaign plugin is not loaded")

	mu        sync.RWMutex
	starter   Starter
	completer Completer
	owners    = map[string]Owner{}
)

func SetCampaign(start Starter, complete Completer) {
	mu.Lock()
	defer mu.Unlock()
	starter, completer = start, complete
}

func SetOwner(name string, fn Owner) {
	mu.Lock()
	defer mu.Unlock()
	if fn == nil {
		delete(owners, name)
		return
	}
	owners[name] = fn
}

func Available() bool {
	mu.RLock()
	defer mu.RUnlock()
	return starter != nil && completer != nil
}

func Start(ctx context.Context, userID, stageKey string) (string, error) {
	mu.RLock()
	fn := starter
	mu.RUnlock()
	if fn == nil {
		return "", ErrUnavailable
	}
	return fn(ctx, userID, stageKey)
}

func Complete(ctx context.Context, userID string, c Completion) (map[string]any, error) {
	mu.RLock()
	fn := completer
	mu.RUnlock()
	if fn == nil {
		return nil, ErrUnavailable
	}
	return fn(ctx, userID, c)
}

func OwnerOf(ctx context.Context, stageKey string) (owner, ref string) {
	mu.RLock()
	names := make([]string, 0, len(owners))
	for n := range owners {
		names = append(names, n)
	}
	sort.Strings(names)
	fns := make([]Owner, len(names))
	for i, n := range names {
		fns[i] = owners[n]
	}
	mu.RUnlock()
	for i, fn := range fns {
		if r, ok := fn(ctx, stageKey); ok {
			return names[i], r
		}
	}
	return "", ""
}
