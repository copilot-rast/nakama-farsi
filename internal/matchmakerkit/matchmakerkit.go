package matchmakerkit

import (
	"errors"
	"sync"
)

type Remover func(sessionID, ticket string) error

var (
	mu      sync.RWMutex
	remover Remover

	ErrUnavailable = errors.New("matchmaker not available")
)

func Set(fn Remover) {
	mu.Lock()
	remover = fn
	mu.Unlock()
}

func Remove(sessionID, ticket string) error {
	mu.RLock()
	fn := remover
	mu.RUnlock()
	if fn == nil {
		return ErrUnavailable
	}
	return fn(sessionID, ticket)
}
