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

type StartCheck struct {
	UserID     string
	StageKey   string
	ChapterKey string
	ItemKey    string
	Server     bool
}

type Result struct {
	UserID         string
	AttemptID      string
	StageKey       string
	ChapterKey     string
	ItemKey        string
	Success        bool
	ElapsedSeconds float64
	HealthPercent  *float64
	Fields         map[string]any
	Replayed       bool
}

type (
	StartGate func(ctx context.Context, c StartCheck) error

	ResultObserver func(ctx context.Context, r Result) (any, error)

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
	gates     = map[string]StartGate{}
	observers = map[string]ResultObserver{}
)

func SetStartGate(name string, fn StartGate) {
	mu.Lock()
	defer mu.Unlock()
	if fn == nil {
		delete(gates, name)
		return
	}
	gates[name] = fn
}

func SetResultObserver(name string, fn ResultObserver) {
	mu.Lock()
	defer mu.Unlock()
	if fn == nil {
		delete(observers, name)
		return
	}
	observers[name] = fn
}

func sortedNames[T any](m map[string]T) ([]string, []T) {
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	fns := make([]T, len(names))
	for i, n := range names {
		fns[i] = m[n]
	}
	return names, fns
}

func CheckStart(ctx context.Context, c StartCheck) error {
	mu.RLock()
	_, fns := sortedNames(gates)
	mu.RUnlock()
	for _, fn := range fns {
		if err := fn(ctx, c); err != nil {
			return err
		}
	}
	return nil
}

func HasResultObservers() bool {
	mu.RLock()
	defer mu.RUnlock()
	return len(observers) > 0
}

func Observe(ctx context.Context, r Result) (map[string]any, map[string]error) {
	mu.RLock()
	names, fns := sortedNames(observers)
	mu.RUnlock()
	out, errs := map[string]any{}, map[string]error{}
	for i, fn := range fns {
		v, err := fn(ctx, r)
		if err != nil {
			errs[names[i]] = err
			continue
		}
		if v != nil {
			out[names[i]] = v
		}
	}
	return out, errs
}

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
