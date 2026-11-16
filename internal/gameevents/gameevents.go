package gameevents

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
)

type Event struct {
	ID      string
	UserID  string
	Name    string
	Context string
	Fields  map[string]any
	Source  string
	Typed   []Typed
}

type Typed struct {
	Type     string         `json:"type"`
	ID       string         `json:"id,omitempty"`
	Value    *float64       `json:"value,omitempty"`
	Status   string         `json:"status,omitempty"`
	P1       string         `json:"p1,omitempty"`
	P2       string         `json:"p2,omitempty"`
	P3       string         `json:"p3,omitempty"`
	Score    *int64         `json:"score,omitempty"`
	Flow     string         `json:"flow,omitempty"`
	Currency string         `json:"currency,omitempty"`
	Amount   *float64       `json:"amount,omitempty"`
	ItemType string         `json:"itemType,omitempty"`
	Item     string         `json:"item,omitempty"`
	CartType string         `json:"cartType,omitempty"`
	Store    string         `json:"store,omitempty"`
	Receipt  string         `json:"receipt,omitempty"`
	Severity string         `json:"severity,omitempty"`
	Message  string         `json:"message,omitempty"`
	Context  map[string]any `json:"context,omitempty"`
}

func Progression(status, p1, p2, p3 string, score *int64) Typed {
	return Typed{Type: "progression", Status: status, P1: p1, P2: p2, P3: p3, Score: score}
}

func Resource(flow, currency string, amount float64, itemType, item string) Typed {
	return Typed{Type: "resource", Flow: flow, Currency: currency, Amount: &amount, ItemType: itemType, Item: item}
}

func Business(currency string, amount int64, itemType, item, cartType, store, receipt string) Typed {
	a := float64(amount)
	return Typed{Type: "business", Currency: currency, Amount: &a, ItemType: itemType, Item: item, CartType: cartType, Store: store, Receipt: receipt}
}

type Subscriber func(ctx context.Context, ev Event) (any, error)

var (
	ErrInvalid = errors.New("a game event needs an id, a user id and a name or a typed event")

	mu   sync.RWMutex
	subs = map[string]Subscriber{}
)

func Subscribe(owner string, fn Subscriber) {
	mu.Lock()
	defer mu.Unlock()
	if fn == nil {
		delete(subs, owner)
		return
	}
	subs[owner] = fn
}

func Has() bool {
	mu.RLock()
	defer mu.RUnlock()
	return len(subs) > 0
}

func Publish(ctx context.Context, ev Event) (map[string]any, map[string]error) {
	ev.ID, ev.UserID, ev.Name, ev.Source = strings.TrimSpace(ev.ID), strings.TrimSpace(ev.UserID), strings.TrimSpace(ev.Name), strings.TrimSpace(ev.Source)
	if ev.ID == "" || ev.UserID == "" || (ev.Name == "" && len(ev.Typed) == 0) {
		return nil, map[string]error{"": ErrInvalid}
	}
	mu.RLock()
	owners := make([]string, 0, len(subs))
	for o := range subs {
		owners = append(owners, o)
	}
	sort.Strings(owners)
	list := make([]Subscriber, len(owners))
	for i, o := range owners {
		list[i] = subs[o]
	}
	mu.RUnlock()
	answers, errs := map[string]any{}, map[string]error{}
	for i, fn := range list {
		out, err := fn(ctx, ev)
		if err != nil {
			errs[owners[i]] = err
			continue
		}
		if out != nil {
			answers[owners[i]] = out
		}
	}
	return answers, errs
}
