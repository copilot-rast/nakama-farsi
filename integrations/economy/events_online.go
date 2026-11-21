package main

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/restrict"
)

const (
	onlineEventKey = "online"
	onlineTick     = 30 * time.Second
)

type onlineTracker struct {
	mu       sync.Mutex
	sessions map[string]string
	counts   map[string]int
	sent     map[string]string
}

var onlineInit sync.Mutex

func (s *service) tracker() *onlineTracker {
	onlineInit.Lock()
	defer onlineInit.Unlock()
	if s.online == nil {
		s.online = newOnlineTracker()
	}
	return s.online
}

func newOnlineTracker() *onlineTracker {
	return &onlineTracker{sessions: map[string]string{}, counts: map[string]int{}, sent: map[string]string{}}
}

func (t *onlineTracker) add(sessionID, userID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.sessions[sessionID]; ok {
		return
	}
	t.sessions[sessionID] = userID
	t.counts[userID]++
}

func (t *onlineTracker) remove(sessionID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	userID, ok := t.sessions[sessionID]
	if !ok {
		return
	}
	delete(t.sessions, sessionID)
	t.counts[userID]--
	if t.counts[userID] <= 0 {
		delete(t.counts, userID)
		delete(t.sent, userID)
	}
}

func (t *onlineTracker) due(userID, mark string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sent[userID] == mark {
		return false
	}
	t.sent[userID] = mark
	return true
}

func (t *onlineTracker) users() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]string, 0, len(t.counts))
	for id := range t.counts {
		out = append(out, id)
	}
	return out
}

func (s *service) onlineMark(ctx context.Context, now time.Time) (string, bool) {
	settings, err := loadEventSettings(ctx, s.db)
	if err != nil {
		return "", false
	}
	rows, err := s.db.QueryContext(ctx, `SELECT st.day_zone, st.day_start_minute, st.period_days FROM economy_streak st JOIN economy_event_type et ON et.id=st.event_type_id
		WHERE et.key=$1 AND et.is_active AND st.is_active ORDER BY st.id`, onlineEventKey)
	if err != nil {
		return "", false
	}
	defer rows.Close()
	clk := settings.clock()
	parts := []string{dateKey(clk.civil(now))}
	for rows.Next() {
		var zone string
		var minute, days int
		if err := rows.Scan(&zone, &minute, &days); err != nil {
			return "", false
		}
		parts = append(parts, strconv.FormatInt(clk.withDay(zone, minute).streakIndex(now, days), 10))
	}
	if rows.Err() != nil {
		return "", false
	}
	return strings.Join(parts, "|"), true
}

func (s *service) recordOnline(ctx context.Context, userID, mark, why string) bool {
	if !s.tracker().due(userID, mark) {
		return false
	}
	if r, err := restrict.Active(ctx, s.db, userID, restrict.KindSuspend); err == nil && r != nil {
		return false
	}
	s.recordServerEvent(ctx, userID, onlineEventKey, 1, map[string]any{"reason": why})
	return true
}

func (s *service) onSessionStart(ctx context.Context, _ runtime.Logger, _ *api.Event) {
	userID, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	sessionID, _ := ctx.Value(runtime.RUNTIME_CTX_SESSION_ID).(string)
	if userID == "" || sessionID == "" {
		return
	}
	s.tracker().add(sessionID, userID)
	mark, ok := s.onlineMark(ctx, s.now())
	if !ok {
		return
	}
	s.recordOnline(ctx, userID, mark, "connect")
}

func (s *service) onSessionEnd(ctx context.Context, _ runtime.Logger, _ *api.Event) {
	sessionID, _ := ctx.Value(runtime.RUNTIME_CTX_SESSION_ID).(string)
	if sessionID != "" {
		s.tracker().remove(sessionID)
	}
}

func (s *service) onlineLoop(ctx context.Context) {
	ticker := time.NewTicker(onlineTick)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		s.onlineSweep(ctx)
	}
}

func (s *service) onlineSweep(ctx context.Context) int {
	users := s.tracker().users()
	if len(users) == 0 {
		return 0
	}
	mark, ok := s.onlineMark(ctx, s.now())
	if !ok {
		return 0
	}
	n := 0
	for _, id := range users {
		if ctx.Err() != nil {
			return n
		}
		if s.recordOnline(ctx, id, mark, "new_day") {
			n++
		}
	}
	return n
}
