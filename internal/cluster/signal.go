package cluster

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Signals interface {
	Signal(topic string)
	Wait(ctx context.Context, topic string, max time.Duration) bool
	Enabled() bool
}

type sleepSignals struct{}

func (sleepSignals) Signal(string) {}

func (sleepSignals) Wait(ctx context.Context, _ string, max time.Duration) bool {
	timer := time.NewTimer(max)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
	return false
}

func (sleepSignals) Enabled() bool { return false }

var (
	signalsMu sync.RWMutex
	signals   Signals = sleepSignals{}
)

func SetSignals(s Signals) {
	signalsMu.Lock()
	signals = s
	signalsMu.Unlock()
}

func current() Signals {
	signalsMu.RLock()
	defer signalsMu.RUnlock()
	return signals
}

func Signal(topic string) {
	current().Signal(topic)
}

func WaitSignal(ctx context.Context, topic string, max time.Duration) bool {
	return current().Wait(ctx, topic, max)
}

func SignalsEnabled() bool {
	return current().Enabled()
}

type redisSignals struct {
	logger *zap.Logger
	rdb    redis.UniversalClient
	keys   Keys
	ctx    context.Context

	mu      sync.Mutex
	waiters map[string]map[chan struct{}]struct{}
}

func NewRedisSignals(ctx context.Context, logger *zap.Logger, rdb redis.UniversalClient, keys Keys) Signals {
	s := &redisSignals{logger: logger, rdb: rdb, keys: keys, ctx: ctx, waiters: make(map[string]map[chan struct{}]struct{})}
	go s.run()
	return s
}

func (s *redisSignals) Enabled() bool { return true }

func (s *redisSignals) run() {
	prefix := s.keys.Signal("")
	for s.ctx.Err() == nil {
		sub := s.rdb.PSubscribe(s.ctx, s.keys.SignalPattern())
		ch := sub.Channel()
		for msg := range ch {
			s.wake(strings.TrimPrefix(msg.Channel, prefix))
		}
		_ = sub.Close()
		sleepCtx(s.ctx, time.Second)
	}
}

func (s *redisSignals) wake(topic string) {
	s.mu.Lock()
	for ch := range s.waiters[topic] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	s.mu.Unlock()
}

func (s *redisSignals) Signal(topic string) {
	ctx, cancel := context.WithTimeout(s.ctx, 2*time.Second)
	defer cancel()
	if err := s.rdb.Publish(ctx, s.keys.Signal(topic), "1").Err(); err != nil {
		s.logger.Warn("Cluster signal could not be published", zap.String("topic", topic), zap.Error(err))
	}
}

func (s *redisSignals) Wait(ctx context.Context, topic string, max time.Duration) bool {
	ch := make(chan struct{}, 1)
	s.mu.Lock()
	set, ok := s.waiters[topic]
	if !ok {
		set = make(map[chan struct{}]struct{})
		s.waiters[topic] = set
	}
	set[ch] = struct{}{}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(set, ch)
		if len(set) == 0 {
			delete(s.waiters, topic)
		}
		s.mu.Unlock()
	}()
	timer := time.NewTimer(max)
	defer timer.Stop()
	select {
	case <-ch:
		return true
	case <-ctx.Done():
	case <-timer.C:
	case <-s.ctx.Done():
	}
	return false
}
