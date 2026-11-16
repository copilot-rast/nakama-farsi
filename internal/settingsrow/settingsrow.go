package settingsrow

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"sync"
	"time"
)

var tablePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type Store[T any] struct {
	Table    string
	Defaults func() T
	TTL      time.Duration

	mu      sync.Mutex
	cached  T
	fetched time.Time
}

type Meta struct {
	UpdatedBy string `json:"updatedBy,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

func New[T any](table string, defaults func() T) *Store[T] {
	if !tablePattern.MatchString(table) {
		panic("settingsrow: invalid table name " + table)
	}
	return &Store[T]{Table: table, Defaults: defaults, TTL: 10 * time.Second}
}

func (s *Store[T]) Invalidate() {
	s.mu.Lock()
	s.fetched = time.Time{}
	s.mu.Unlock()
}

func (s *Store[T]) Load(ctx context.Context, db *sql.DB) (T, error) {
	if db == nil {
		return s.Defaults(), nil
	}
	s.mu.Lock()
	if !s.fetched.IsZero() && time.Since(s.fetched) < s.TTL {
		out := s.cached
		s.mu.Unlock()
		return out, nil
	}
	s.mu.Unlock()
	out, _, err := s.Read(ctx, db)
	if err != nil {
		return s.Defaults(), err
	}
	s.mu.Lock()
	s.cached, s.fetched = out, time.Now()
	s.mu.Unlock()
	return out, nil
}

func (s *Store[T]) Read(ctx context.Context, db *sql.DB) (T, Meta, error) {
	out := s.Defaults()
	var raw []byte
	var by string
	var at time.Time
	err := db.QueryRowContext(ctx, `SELECT value, updated_by, updated_at FROM `+s.Table+` WHERE id=1`).Scan(&raw, &by, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return out, Meta{}, nil
	}
	if err != nil {
		return out, Meta{}, err
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			return s.Defaults(), Meta{}, err
		}
	}
	m := Meta{UpdatedBy: by}
	if by != "" {
		m.UpdatedAt = at.UTC().Format(time.RFC3339)
	}
	return out, m, nil
}

func (s *Store[T]) Save(ctx context.Context, db *sql.DB, v T, by string) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `INSERT INTO `+s.Table+` (id, value, updated_by, updated_at) VALUES (1, $1, $2, now())
		ON CONFLICT (id) DO UPDATE SET value=EXCLUDED.value, updated_by=EXCLUDED.updated_by, updated_at=now()`, raw, by)
	s.Invalidate()
	return err
}
