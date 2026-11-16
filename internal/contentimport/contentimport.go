package contentimport

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type Importer func(ctx context.Context, doc []byte, by string) (map[string]any, error)

var (
	ErrUnknown = errors.New("contentimport: no importer registered under this name")

	mu        sync.RWMutex
	importers = map[string]Importer{}
)

func Register(name string, fn Importer) {
	mu.Lock()
	defer mu.Unlock()
	if fn == nil {
		delete(importers, name)
		return
	}
	importers[name] = fn
}

func Names() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(importers))
	for n := range importers {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func get(name string) Importer {
	mu.RLock()
	defer mu.RUnlock()
	return importers[name]
}

func Import(ctx context.Context, name string, doc []byte, by string) (map[string]any, error) {
	fn := get(name)
	if fn == nil {
		return nil, ErrUnknown
	}
	return fn(ctx, doc, by)
}

func Await(ctx context.Context, name string, wait time.Duration) bool {
	deadline := time.Now().Add(wait)
	for {
		if get(name) != nil {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(200 * time.Millisecond):
		}
	}
}
