package insightsapi

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Tag struct {
	Group string `json:"tagGroup"`
	ID    string `json:"id"`
}

type Event struct {
	Key        string
	UserID     string
	SessionID  string
	Game       string
	Name       string
	Path       string
	Tags       []Tag
	Value      *float64
	Time       time.Time
	Fields     map[string]any
	Platform   string
	AppVersion string
	Source     string
}

type Result struct {
	Accepted int      `json:"accepted"`
	Flagged  int      `json:"flagged"`
	Rejected int      `json:"rejected"`
	Codes    []string `json:"codes,omitempty"`
}

type Sink func(ctx context.Context, events []Event) (Result, error)

var (
	ErrNoSink = errors.New("insights: no analytics sink registered")

	mu   sync.RWMutex
	sink Sink
)

func Register(fn Sink) {
	mu.Lock()
	sink = fn
	mu.Unlock()
}

func Enabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return sink != nil
}

func Send(ctx context.Context, events ...Event) (Result, error) {
	if len(events) == 0 {
		return Result{}, nil
	}
	mu.RLock()
	fn := sink
	mu.RUnlock()
	if fn == nil {
		return Result{}, ErrNoSink
	}
	return fn(ctx, events)
}

func Num(v float64) *float64 { return &v }

func Path(path string, value float64) Event { return Event{Path: path, Value: &value} }

func Tags(pairs ...string) []Tag {
	out := make([]Tag, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, Tag{Group: pairs[i], ID: pairs[i+1]})
	}
	return out
}
