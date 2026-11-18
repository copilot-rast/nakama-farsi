package contenthooks

import (
	"context"
	"errors"
	"sync"
)

type CatalogReader func(ctx context.Context, platform, appVersion string) (map[string]any, error)

var (
	ErrUnavailable = errors.New("content packs are not available: the content plugin is not loaded")

	mu     sync.RWMutex
	reader CatalogReader
)

func SetCatalog(fn CatalogReader) {
	mu.Lock()
	defer mu.Unlock()
	reader = fn
}

func Available() bool {
	mu.RLock()
	defer mu.RUnlock()
	return reader != nil
}

func Catalog(ctx context.Context, platform, appVersion string) (map[string]any, error) {
	mu.RLock()
	fn := reader
	mu.RUnlock()
	if fn == nil {
		return nil, ErrUnavailable
	}
	return fn(ctx, platform, appVersion)
}
