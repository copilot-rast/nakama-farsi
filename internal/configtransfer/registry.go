package configtransfer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"sync"
)

const (
	Format        = "nakama-config"
	FormatVersion = 1

	StatusNew          = "new"
	StatusChanged      = "changed"
	StatusSame         = "same"
	StatusMissingRefs  = "missing_refs"
	StatusSecretNeeded = "secret_needed"
	StatusSkipped      = "skipped"
	StatusSwitchedOff  = "switched_off"
)

var (
	keyPattern   = regexp.MustCompile(`^[a-z][a-z0-9_]{0,47}$`)
	identPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

	ErrFormat  = errors.New("not a config transfer file")
	ErrVersion = errors.New("config transfer file version is not supported")
)

type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Ref struct {
	Column string
	Path   string
	Table  string
	Key    []string
	By     string
}

type Table struct {
	Name   string
	Key    []string
	Refs   []Ref
	Skip   []string
	Secret []string
	Active string
	Where  string
}

type Custom interface {
	Export(ctx context.Context, q Querier) ([]FileTable, error)
	Apply(ctx context.Context, q Querier, in []FileTable, pick func(table, key string) bool, by string) ([]RowResult, error)
}

type Area struct {
	Key    string
	Owner  string
	Title  map[string]string
	Note   map[string]string
	Order  int
	Tables []Table
	Custom Custom
	After  func(ctx context.Context, res AreaResult) []string
}

type AreaInfo struct {
	Key       string            `json:"key"`
	Owner     string            `json:"owner"`
	Title     map[string]string `json:"title"`
	Note      map[string]string `json:"note,omitempty"`
	Order     int               `json:"order"`
	Tables    []string          `json:"tables"`
	SwitchOff bool              `json:"switchOff"`
}

var (
	mu    sync.RWMutex
	areas = map[string]Area{}
)

func Register(list ...Area) error {
	mu.Lock()
	defer mu.Unlock()
	for _, a := range list {
		if err := validArea(a); err != nil {
			return err
		}
	}
	for _, a := range list {
		areas[a.Key] = a
	}
	return nil
}

func Unregister(keys ...string) {
	mu.Lock()
	defer mu.Unlock()
	for _, k := range keys {
		delete(areas, k)
	}
}

func Lookup(key string) (Area, bool) {
	mu.RLock()
	defer mu.RUnlock()
	a, ok := areas[key]
	return a, ok
}

func Areas() []Area {
	mu.RLock()
	out := make([]Area, 0, len(areas))
	for _, a := range areas {
		out = append(out, a)
	}
	mu.RUnlock()
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Key < out[j].Key
	})
	return out
}

func List() []AreaInfo {
	list := Areas()
	out := make([]AreaInfo, 0, len(list))
	for _, a := range list {
		info := AreaInfo{Key: a.Key, Owner: a.Owner, Title: a.Title, Note: a.Note, Order: a.Order, Tables: []string{}}
		for _, t := range a.Tables {
			info.Tables = append(info.Tables, t.Name)
			if t.Active != "" {
				info.SwitchOff = true
			}
		}
		out = append(out, info)
	}
	return out
}

func validArea(a Area) error {
	if !keyPattern.MatchString(a.Key) {
		return fmt.Errorf("configtransfer: invalid area key %q", a.Key)
	}
	if a.Custom == nil && len(a.Tables) == 0 {
		return fmt.Errorf("configtransfer: area %s has no tables", a.Key)
	}
	for _, t := range a.Tables {
		if !identPattern.MatchString(t.Name) {
			return fmt.Errorf("configtransfer: area %s: invalid table %q", a.Key, t.Name)
		}
		if len(t.Key) == 0 {
			return fmt.Errorf("configtransfer: area %s: table %s has no key", a.Key, t.Name)
		}
		names := append(append(append(append([]string{}, t.Key...), t.Skip...), t.Secret...), t.Active)
		for _, r := range t.Refs {
			names = append(names, r.Column, r.Table, r.By)
			names = append(names, r.Key...)
		}
		for _, n := range names {
			if n != "" && !identPattern.MatchString(n) {
				return fmt.Errorf("configtransfer: area %s: invalid name %q", a.Key, n)
			}
		}
		for _, r := range t.Refs {
			if r.Column == "" || r.Table == "" {
				return fmt.Errorf("configtransfer: area %s: ref without column or table", a.Key)
			}
			if _, err := parsePath(r.Path); err != nil {
				return fmt.Errorf("configtransfer: area %s: %w", a.Key, err)
			}
		}
	}
	return nil
}

func (r Ref) keyCols() []string {
	if len(r.Key) == 0 {
		return []string{"key"}
	}
	return r.Key
}

func (r Ref) byCol() string {
	if r.By == "" {
		return "id"
	}
	return r.By
}
