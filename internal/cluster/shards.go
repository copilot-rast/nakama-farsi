package cluster

import (
	"context"
	"hash/fnv"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Shards struct {
	rdb  redis.UniversalClient
	keys Keys
	n    int
	ttl  time.Duration
}

func NewShards(rdb redis.UniversalClient, keys Keys, n int, ttl time.Duration) *Shards {
	if n < 1 {
		n = 1
	}
	if ttl <= 0 {
		ttl = 6 * time.Second
	}
	return &Shards{rdb: rdb, keys: keys, n: n, ttl: ttl}
}

func (s *Shards) Count() int {
	return s.n
}

func (s *Shards) TTL() time.Duration {
	return s.ttl
}

func (s *Shards) OwnerKey(i int) string {
	return s.keys.Prefix + ":mm:{mm" + strconv.Itoa(i) + "}:own"
}

func (s *Shards) TicketsKey(i int) string {
	return s.keys.Prefix + ":mm:{mm" + strconv.Itoa(i) + "}:t"
}

func (s *Shards) ShardOf(key string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum32() % uint32(s.n))
}

func Preferred(shard int, alive []string) string {
	best := ""
	var bestScore uint64
	for _, name := range alive {
		h := fnv.New64a()
		_, _ = h.Write([]byte(name))
		_, _ = h.Write([]byte{'#'})
		_, _ = h.Write([]byte(strconv.Itoa(shard)))
		score := mix64(h.Sum64())
		if best == "" || score > bestScore || (score == bestScore && name < best) {
			best, bestScore = name, score
		}
	}
	return best
}

func OwnerName(value string) string {
	if i := strings.IndexByte(value, '/'); i >= 0 {
		return value[:i]
	}
	return value
}

var renewShardScript = redis.NewScript(`
local v = redis.call('GET', KEYS[1])
if v == ARGV[1] then
  redis.call('PEXPIRE', KEYS[1], ARGV[2])
  return 1
end
if not v then
  redis.call('SET', KEYS[1], ARGV[1], 'PX', ARGV[2])
  return 2
end
return 0
`)

var releaseShardScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('DEL', KEYS[1])
end
return 0
`)

var putTicketScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) ~= ARGV[1] then
  return 0
end
redis.call('HSET', KEYS[2], ARGV[2], ARGV[3])
return 1
`)

var commitScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) ~= ARGV[1] then
  return {0}
end
local missing = {}
for i = 2, #ARGV do
  if redis.call('HEXISTS', KEYS[2], ARGV[i]) == 0 then
    table.insert(missing, ARGV[i])
  end
end
if #missing > 0 then
  return {2, missing}
end
for i = 2, #ARGV do
  redis.call('HDEL', KEYS[2], ARGV[i])
end
return {1}
`)

func (s *Shards) Acquire(ctx context.Context, i int, token string) (bool, error) {
	return s.rdb.SetNX(ctx, s.OwnerKey(i), token, s.ttl).Result()
}

const (
	RenewLost      = 0
	RenewOK        = 1
	RenewReclaimed = 2
)

func (s *Shards) Renew(ctx context.Context, i int, token string) (int, error) {
	return renewShardScript.Run(ctx, s.rdb, []string{s.OwnerKey(i)}, token, s.ttl.Milliseconds()).Int()
}

func (s *Shards) Release(ctx context.Context, i int, token string) error {
	return releaseShardScript.Run(ctx, s.rdb, []string{s.OwnerKey(i)}, token).Err()
}

func (s *Shards) Owners(ctx context.Context) ([]string, error) {
	pipe := s.rdb.Pipeline()
	cmds := make([]*redis.StringCmd, s.n)
	for i := range cmds {
		cmds[i] = pipe.Get(ctx, s.OwnerKey(i))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}
	out := make([]string, s.n)
	for i, c := range cmds {
		if v, err := c.Result(); err == nil {
			out[i] = OwnerName(v)
		}
	}
	return out, nil
}

func (s *Shards) PutTicket(ctx context.Context, i int, token, ticket, value string) (bool, error) {
	n, err := putTicketScript.Run(ctx, s.rdb, []string{s.OwnerKey(i), s.TicketsKey(i)}, token, ticket, value).Int()
	return n == 1, err
}

func (s *Shards) PutTickets(ctx context.Context, i int, tickets map[string]string) error {
	if len(tickets) == 0 {
		return nil
	}
	return s.rdb.HSet(ctx, s.TicketsKey(i), tickets).Err()
}

type CommitResult int

const (
	CommitLost CommitResult = iota
	CommitOK
	CommitMissing
)

func (s *Shards) Commit(ctx context.Context, i int, token string, tickets []string) (CommitResult, []string, error) {
	args := make([]any, 0, len(tickets)+1)
	args = append(args, token)
	for _, t := range tickets {
		args = append(args, t)
	}
	res, err := commitScript.Run(ctx, s.rdb, []string{s.OwnerKey(i), s.TicketsKey(i)}, args...).Slice()
	if err != nil {
		return CommitLost, nil, err
	}
	code, _ := res[0].(int64)
	switch code {
	case 1:
		return CommitOK, nil, nil
	case 2:
		var missing []string
		if len(res) > 1 {
			missing = stringsOf(res[1], false)
		}
		return CommitMissing, missing, nil
	}
	return CommitLost, nil, nil
}

func (s *Shards) DeleteTickets(ctx context.Context, i int, tickets ...string) error {
	if len(tickets) == 0 {
		return nil
	}
	return s.rdb.HDel(ctx, s.TicketsKey(i), tickets...).Err()
}

func (s *Shards) DeleteTicketsEverywhere(ctx context.Context, tickets ...string) error {
	if len(tickets) == 0 {
		return nil
	}
	pipe := s.rdb.Pipeline()
	for i := 0; i < s.n; i++ {
		pipe.HDel(ctx, s.TicketsKey(i), tickets...)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Shards) Tickets(ctx context.Context, i int, tickets ...string) (map[string]string, error) {
	if len(tickets) == 0 {
		return s.rdb.HGetAll(ctx, s.TicketsKey(i)).Result()
	}
	vals, err := s.rdb.HMGet(ctx, s.TicketsKey(i), tickets...).Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(tickets))
	for j, v := range vals {
		if str, ok := v.(string); ok {
			out[tickets[j]] = str
		}
	}
	return out, nil
}

func mix64(x uint64) uint64 {
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return x
}
