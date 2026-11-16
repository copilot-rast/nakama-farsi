package playerevents

import (
	"context"
	"errors"
	"strings"
	"sync"
)

type Event struct {
	UserID   string
	Key      string
	Value    float64
	Metadata map[string]any
}

type Result map[string]any

type Recorder func(ctx context.Context, ev Event) (Result, error)

var (
	ErrUnavailable = errors.New("player events are not available: the economy plugin is not loaded")
	ErrInvalid     = errors.New("player event needs a user id and a key")
	mu             sync.RWMutex
	recorder       Recorder
)

func SetRecorder(r Recorder) {
	mu.Lock()
	defer mu.Unlock()
	recorder = r
}

func Available() bool {
	mu.RLock()
	defer mu.RUnlock()
	return recorder != nil
}

func Record(ctx context.Context, ev Event) (Result, error) {
	ev.UserID = strings.TrimSpace(ev.UserID)
	ev.Key = strings.TrimSpace(ev.Key)
	if ev.UserID == "" || ev.Key == "" {
		return nil, ErrInvalid
	}
	mu.RLock()
	r := recorder
	mu.RUnlock()
	if r == nil {
		return nil, ErrUnavailable
	}
	return r(ctx, ev)
}

type Recorded struct {
	ID       string
	UserID   string
	Key      string
	Value    float64
	Source   string
	Metadata map[string]any
}

type Observer func(ctx context.Context, ev Recorded)

var (
	obsMu     sync.RWMutex
	observers = map[string]Observer{}
)

func Observe(owner string, fn Observer) {
	obsMu.Lock()
	defer obsMu.Unlock()
	if fn == nil {
		delete(observers, owner)
		return
	}
	observers[owner] = fn
}

func Notify(ctx context.Context, ev Recorded) {
	obsMu.RLock()
	list := make([]Observer, 0, len(observers))
	for _, fn := range observers {
		list = append(list, fn)
	}
	obsMu.RUnlock()
	for _, fn := range list {
		fn(ctx, ev)
	}
}
