package inapp

import (
	"context"
	"errors"
	"sync"
)

type Action struct {
	Type  string            `json:"type"`
	Value string            `json:"value"`
	Label map[string]string `json:"label,omitempty"`
}

type Message struct {
	UserIDs   []string
	Title     map[string]string
	Body      map[string]string
	Action    *Action
	CreatedBy string
}

type Sender func(ctx context.Context, m Message) (string, error)

var (
	ErrUnavailable = errors.New("in-app notifications are not available: the publication plugin is not loaded")
	mu             sync.RWMutex
	sender         Sender
)

func SetSender(s Sender) {
	mu.Lock()
	defer mu.Unlock()
	sender = s
}

func Available() bool {
	mu.RLock()
	defer mu.RUnlock()
	return sender != nil
}

func Send(ctx context.Context, m Message) (string, error) {
	mu.RLock()
	s := sender
	mu.RUnlock()
	if s == nil {
		return "", ErrUnavailable
	}
	return s(ctx, m)
}
