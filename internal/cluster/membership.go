package cluster

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type NodeInfo struct {
	Name    string  `json:"name"`
	Inc     string  `json:"inc"`
	Addr    string  `json:"addr"`
	Hash    [6]byte `json:"-"`
	Started int64   `json:"started"`
}

func (n NodeInfo) Field() string {
	return n.Name + "/" + n.Inc
}

func NodeHash(name string) [6]byte {
	sum := sha1.Sum([]byte(name))
	var h [6]byte
	copy(h[:], sum[:6])
	return h
}

func NewIncarnation() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

type MembershipEvents struct {
	Up           func(NodeInfo)
	Down         func(NodeInfo)
	Dead         func(NodeInfo)
	Epoch        func()
	Reincarnated func(old, current string)
	RedisUp      func()
}

type Membership struct {
	logger *zap.Logger
	rdb    redis.UniversalClient
	keys   Keys
	every  time.Duration
	ttl    time.Duration
	events MembershipEvents

	mu      sync.RWMutex
	self    NodeInfo
	alive   map[string]NodeInfo
	byHash  map[[6]byte]string
	epoch   string
	okSince time.Time
	redisOK bool

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

var heartbeatScript = redis.NewScript(`
if redis.call('EXISTS', KEYS[3]) == 1 then
  return 2
end
local v = redis.call('GET', KEYS[1])
if v and v ~= ARGV[1] then
  return 0
end
redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2])
redis.call('HSET', KEYS[2], ARGV[3], ARGV[4])
return 1
`)

var finishReapScript = redis.NewScript(`
redis.call('HDEL', KEYS[1], ARGV[1])
redis.call('SET', KEYS[2], '1', 'PX', ARGV[2])
if redis.call('GET', KEYS[3]) == ARGV[3] then
  redis.call('DEL', KEYS[3])
end
local v = redis.call('GET', KEYS[4])
if v == ARGV[4] then
  redis.call('DEL', KEYS[4])
end
return 1
`)

var nodeLeaveScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  redis.call('DEL', KEYS[1])
end
redis.call('HDEL', KEYS[2], ARGV[2])
return 1
`)

func NewMembership(logger *zap.Logger, rdb redis.UniversalClient, keys Keys, self NodeInfo, every, ttl time.Duration, events MembershipEvents) *Membership {
	if every <= 0 {
		every = time.Second
	}
	if ttl <= every {
		ttl = 5 * every
	}
	self.Hash = NodeHash(self.Name)
	if self.Started == 0 {
		self.Started = time.Now().UnixMilli()
	}
	m := &Membership{
		logger: logger,
		rdb:    rdb,
		keys:   keys,
		every:  every,
		ttl:    ttl,
		events: events,
		self:   self,
		alive:  map[string]NodeInfo{self.Name: self},
		byHash: map[[6]byte]string{self.Hash: self.Name},
		done:   make(chan struct{}),
	}
	m.ctx, m.cancel = context.WithCancel(context.Background())
	return m
}

func (m *Membership) Self() NodeInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.self
}

func (m *Membership) Start(ctx context.Context) error {
	if _, err := m.beat(ctx); err != nil {
		return err
	}
	m.mu.Lock()
	m.okSince = time.Now()
	m.redisOK = true
	m.mu.Unlock()
	m.watch(ctx)
	go m.loop()
	return nil
}

func (m *Membership) loop() {
	defer close(m.done)
	ticker := time.NewTicker(m.every)
	defer ticker.Stop()
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
		}
		ctx, cancel := context.WithTimeout(m.ctx, m.ttl)
		changed, err := m.beat(ctx)
		if err != nil {
			m.mu.Lock()
			if m.redisOK {
				m.logger.Error("Cluster heartbeat failed", zap.Error(err))
			}
			m.redisOK = false
			m.okSince = time.Time{}
			m.mu.Unlock()
			cancel()
			continue
		}
		m.mu.Lock()
		recovered := !m.redisOK
		if recovered {
			m.okSince = time.Now()
			m.redisOK = true
			m.logger.Warn("Cluster heartbeat recovered")
		}
		m.mu.Unlock()
		if recovered && m.events.RedisUp != nil {
			m.events.RedisUp()
		}
		if changed && m.events.Epoch != nil {
			m.events.Epoch()
		}
		m.watch(ctx)
		cancel()
	}
}

func (m *Membership) beat(ctx context.Context) (bool, error) {
	m.mu.RLock()
	self := m.self
	m.mu.RUnlock()
	payload, _ := json.Marshal(self)
	res, err := heartbeatScript.Run(ctx, m.rdb, []string{m.keys.Heartbeat(self.Name), m.keys.Nodes(), m.keys.Dead(self.Field())}, self.Inc, m.ttl.Milliseconds(), self.Field(), string(payload)).Int()
	if err != nil {
		return false, err
	}
	if res != 1 {
		old := self.Inc
		m.mu.Lock()
		m.self.Inc = NewIncarnation()
		self = m.self
		m.alive[self.Name] = self
		m.mu.Unlock()
		if res == 0 {
			m.logger.Error("Cluster node name taken over by another incarnation, registering again", zap.String("node", self.Name), zap.String("old", old), zap.String("inc", self.Inc))
		} else {
			m.logger.Error("Cluster node was declared dead and reaped, registering again", zap.String("node", self.Name), zap.String("old", old), zap.String("inc", self.Inc))
		}
		payload, _ = json.Marshal(self)
		if _, err := m.rdb.Set(ctx, m.keys.Heartbeat(self.Name), self.Inc, m.ttl).Result(); err != nil {
			return false, err
		}
		if err := m.rdb.HSet(ctx, m.keys.Nodes(), self.Field(), string(payload)).Err(); err != nil {
			return false, err
		}
		if m.events.Reincarnated != nil {
			m.events.Reincarnated(old, self.Inc)
		}
	}
	epoch, err := m.rdb.Get(ctx, m.keys.Epoch()).Result()
	if errors.Is(err, redis.Nil) {
		fresh := NewIncarnation()
		if ok, err := m.rdb.SetNX(ctx, m.keys.Epoch(), fresh, 0).Result(); err != nil {
			return false, err
		} else if ok {
			epoch = fresh
		} else if epoch, err = m.rdb.Get(ctx, m.keys.Epoch()).Result(); err != nil {
			return false, err
		}
	} else if err != nil {
		return false, err
	}
	m.mu.Lock()
	changed := m.epoch != "" && m.epoch != epoch
	m.epoch = epoch
	m.mu.Unlock()
	return changed, nil
}

func (m *Membership) watch(ctx context.Context) {
	entries, err := m.rdb.HGetAll(ctx, m.keys.Nodes()).Result()
	if err != nil {
		return
	}
	infos := make([]NodeInfo, 0, len(entries))
	names := make([]string, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for field, raw := range entries {
		var info NodeInfo
		if err := json.Unmarshal([]byte(raw), &info); err != nil || info.Name == "" {
			continue
		}
		if info.Field() != field {
			continue
		}
		info.Hash = NodeHash(info.Name)
		infos = append(infos, info)
		if _, ok := seen[info.Name]; !ok {
			seen[info.Name] = struct{}{}
			names = append(names, info.Name)
		}
	}
	current := make(map[string]string, len(names))
	if len(names) > 0 {
		keys := make([]string, len(names))
		for i, n := range names {
			keys[i] = m.keys.Heartbeat(n)
		}
		vals, err := m.rdb.MGet(ctx, keys...).Result()
		if err != nil {
			return
		}
		for i, v := range vals {
			if s, ok := v.(string); ok {
				current[names[i]] = s
			}
		}
	}

	m.mu.Lock()
	self := m.self
	judge := m.redisOK && !m.okSince.IsZero() && time.Since(m.okSince) >= m.ttl
	alive := make(map[string]NodeInfo, len(infos))
	var dead []NodeInfo
	for _, info := range infos {
		if info.Name == self.Name {
			if info.Inc != self.Inc {
				dead = append(dead, info)
			}
			continue
		}
		if current[info.Name] == info.Inc {
			alive[info.Name] = info
		} else {
			dead = append(dead, info)
		}
	}
	alive[self.Name] = self
	var ups, downs []NodeInfo
	for name, info := range alive {
		if prev, ok := m.alive[name]; !ok || prev.Inc != info.Inc || prev.Addr != info.Addr {
			if ok && name != self.Name {
				downs = append(downs, prev)
			}
			if name != self.Name {
				ups = append(ups, info)
			}
		}
	}
	for name, prev := range m.alive {
		if _, ok := alive[name]; !ok {
			if judge {
				downs = append(downs, prev)
			} else {
				alive[name] = prev
			}
		}
	}
	m.alive = alive
	m.byHash = make(map[[6]byte]string, len(alive))
	for name, info := range alive {
		m.byHash[info.Hash] = name
	}
	m.mu.Unlock()

	sort.Slice(downs, func(i, j int) bool { return downs[i].Name < downs[j].Name })
	for _, d := range downs {
		if m.events.Down != nil {
			m.events.Down(d)
		}
	}
	for _, u := range ups {
		if m.events.Up != nil {
			m.events.Up(u)
		}
	}
	if judge && m.events.Dead != nil {
		for _, d := range dead {
			m.events.Dead(d)
		}
	}
}

func (m *Membership) Alive() []NodeInfo {
	m.mu.RLock()
	out := make([]NodeInfo, 0, len(m.alive))
	for _, info := range m.alive {
		out = append(out, info)
	}
	m.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (m *Membership) AliveNames() []string {
	m.mu.RLock()
	out := make([]string, 0, len(m.alive))
	for name := range m.alive {
		out = append(out, name)
	}
	m.mu.RUnlock()
	sort.Strings(out)
	return out
}

func (m *Membership) IsAlive(name string) bool {
	m.mu.RLock()
	_, ok := m.alive[name]
	m.mu.RUnlock()
	return ok
}

func (m *Membership) Info(name string) (NodeInfo, bool) {
	m.mu.RLock()
	info, ok := m.alive[name]
	m.mu.RUnlock()
	return info, ok
}

func (m *Membership) NodeByHash(h [6]byte) (string, bool) {
	m.mu.RLock()
	name, ok := m.byHash[h]
	m.mu.RUnlock()
	return name, ok
}

func (m *Membership) TryReap(ctx context.Context, info NodeInfo, holder string) (bool, error) {
	return m.rdb.SetNX(ctx, m.keys.Reap(info.Field()), holder, 30*time.Second).Result()
}

func (m *Membership) RenewReap(ctx context.Context, info NodeInfo) {
	m.rdb.PExpire(ctx, m.keys.Reap(info.Field()), 30*time.Second)
}

func (m *Membership) FinishReap(ctx context.Context, info NodeInfo, holder string) error {
	return finishReapScript.Run(ctx, m.rdb, []string{m.keys.Nodes(), m.keys.Dead(info.Field()), m.keys.Reap(info.Field()), m.keys.Heartbeat(info.Name)}, info.Field(), (10 * time.Minute).Milliseconds(), holder, info.Inc).Err()
}

func (m *Membership) Stop(ctx context.Context) {
	m.cancel()
	<-m.done
	self := m.Self()
	if err := nodeLeaveScript.Run(ctx, m.rdb, []string{m.keys.Heartbeat(self.Name), m.keys.Nodes()}, self.Inc, self.Field()).Err(); err != nil {
		m.logger.Warn("Cluster leave failed", zap.Error(err))
	}
}

func SplitField(field string) (string, string) {
	i := strings.LastIndexByte(field, '/')
	if i < 0 {
		return field, ""
	}
	return field[:i], field[i+1:]
}

func (m *Membership) KillForTest() {
	m.cancel()
	<-m.done
}
