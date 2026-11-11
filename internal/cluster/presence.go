package cluster

import (
	"context"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

const Sep = "\x1f"

type StreamRef struct {
	Tag    string
	Suffix string
}

type PresenceOp struct {
	Ref          StreamRef
	Leave        bool
	Field        string
	Meta         string
	Node         string
	Followers    bool
	WantSnapshot bool
	Index        string
	IndexMember  string
}

type PresenceResult struct {
	Changed   bool
	Version   int64
	Nodes     []string
	Followers []string
	Snapshot  map[string]string
	OldMeta   string
	Err       error
}

type PresenceStore struct {
	rdb  redis.UniversalClient
	keys Keys
}

func NewPresenceStore(rdb redis.UniversalClient, keys Keys) *PresenceStore {
	return &PresenceStore{rdb: rdb, keys: keys}
}

func (s *PresenceStore) P(r StreamRef) string {
	return s.keys.Prefix + ":{" + r.Tag + "}:p" + r.Suffix
}

func (s *PresenceStore) N(r StreamRef) string {
	return s.keys.Prefix + ":{" + r.Tag + "}:n" + r.Suffix
}

func (s *PresenceStore) F(r StreamRef) string {
	return s.keys.Prefix + ":{" + r.Tag + "}:f"
}

func (s *PresenceStore) FollowersKey(userTag string) string {
	return s.keys.Prefix + ":{" + userTag + "}:f"
}

var versionLua = `
local t = redis.call('TIME')
local now = tonumber(t[1]) * 1000000 + tonumber(t[2])
local prev = tonumber(redis.call('HGET', KEYS[2], '_v') or '0')
local v = now
if prev >= v then v = prev + 1 end
redis.call('HSET', KEYS[2], '_v', string.format('%.0f', v))
`

var joinScript = redis.NewScript(`
redis.call('PERSIST', KEYS[2])
local added = redis.call('HSET', KEYS[1], ARGV[1], ARGV[2])
if added == 1 then redis.call('HINCRBY', KEYS[2], ARGV[3], 1) end
` + versionLua + `
local nodes = redis.call('HKEYS', KEYS[2])
local fol = {}
if ARGV[4] == '1' then fol = redis.call('HKEYS', KEYS[3]) end
local snap = {}
if ARGV[5] == '1' then snap = redis.call('HGETALL', KEYS[1]) end
return {added, string.format('%.0f', v), nodes, fol, snap}
`)

var leaveScript = redis.NewScript(`
local old = redis.call('HGET', KEYS[1], ARGV[1])
if not old then
  return {0, '0', {}, {}, ''}
end
redis.call('HDEL', KEYS[1], ARGV[1])
local c = redis.call('HINCRBY', KEYS[2], ARGV[3], -1)
if c <= 0 then redis.call('HDEL', KEYS[2], ARGV[3]) end
` + versionLua + `
local nodes = redis.call('HKEYS', KEYS[2])
if redis.call('HLEN', KEYS[1]) == 0 then
  redis.call('DEL', KEYS[1])
  redis.call('PEXPIRE', KEYS[2], 600000)
else
  redis.call('PERSIST', KEYS[2])
end
local fol = {}
if ARGV[4] == '1' then fol = redis.call('HKEYS', KEYS[3]) end
return {1, string.format('%.0f', v), nodes, fol, old}
`)

var snapshotScript = redis.NewScript(`
return {redis.call('HGET', KEYS[2], '_v') or '0', redis.call('HGETALL', KEYS[1])}
`)

func flag(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (s *PresenceStore) LoadScripts(ctx context.Context) error {
	for _, sc := range []*redis.Script{joinScript, leaveScript, snapshotScript} {
		if err := sc.Load(ctx, s.rdb).Err(); err != nil {
			return err
		}
	}
	return nil
}

func anyNoScript(cmds []*redis.Cmd) bool {
	for _, c := range cmds {
		if c != nil && isNoScript(c.Err()) {
			return true
		}
	}
	return false
}

func (s *PresenceStore) Apply(ctx context.Context, ops []PresenceOp) ([]PresenceResult, error) {
	return s.apply(ctx, ops, true)
}

func (s *PresenceStore) apply(ctx context.Context, ops []PresenceOp, retry bool) ([]PresenceResult, error) {
	if len(ops) == 0 {
		return nil, nil
	}
	pipe := s.rdb.Pipeline()
	cmds := make([]*redis.Cmd, len(ops))
	for i, op := range ops {
		keys := []string{s.P(op.Ref), s.N(op.Ref), s.F(op.Ref)}
		if op.Leave {
			cmds[i] = leaveScript.Run(ctx, pipe, keys, op.Field, "", op.Node, flag(op.Followers))
			if op.Index != "" {
				pipe.SRem(ctx, op.Index, op.IndexMember)
			}
		} else {
			cmds[i] = joinScript.Run(ctx, pipe, keys, op.Field, op.Meta, op.Node, flag(op.Followers), flag(op.WantSnapshot))
			if op.Index != "" {
				pipe.SAdd(ctx, op.Index, op.IndexMember)
			}
		}
	}
	_, err := pipe.Exec(ctx)
	if retry && anyNoScript(cmds) {
		if lerr := s.LoadScripts(ctx); lerr == nil {
			return s.apply(ctx, ops, false)
		}
	}
	results := make([]PresenceResult, len(ops))
	for i, cmd := range cmds {
		v, cerr := cmd.Slice()
		if cerr != nil {
			results[i].Err = cerr
			if err == nil {
				err = cerr
			}
			continue
		}
		results[i] = parsePresenceResult(v, !ops[i].Leave)
	}
	return results, err
}

func isNoScript(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT")
}

func parsePresenceResult(v []any, join bool) PresenceResult {
	var r PresenceResult
	if len(v) < 5 {
		return r
	}
	n, _ := v[0].(int64)
	r.Changed = n == 1
	if vs, ok := v[1].(string); ok {
		r.Version, _ = strconv.ParseInt(vs, 10, 64)
	}
	r.Nodes = stringsOf(v[2], true)
	r.Followers = stringsOf(v[3], false)
	if join {
		if snap, ok := v[4].([]any); ok && len(snap) > 0 {
			r.Snapshot = make(map[string]string, len(snap)/2)
			for i := 0; i+1 < len(snap); i += 2 {
				k, _ := snap[i].(string)
				val, _ := snap[i+1].(string)
				r.Snapshot[k] = val
			}
		}
	} else {
		r.OldMeta, _ = v[4].(string)
	}
	return r
}

func stringsOf(v any, dropVersion bool) []string {
	list, _ := v.([]any)
	out := make([]string, 0, len(list))
	for _, x := range list {
		if s, ok := x.(string); ok {
			if dropVersion && s == "_v" {
				continue
			}
			out = append(out, s)
		}
	}
	return out
}

type Snapshot struct {
	Version int64
	Entries map[string]string
}

func (s *PresenceStore) Snapshots(ctx context.Context, refs []StreamRef) ([]Snapshot, error) {
	pipe := s.rdb.Pipeline()
	cmds := make([]*redis.Cmd, len(refs))
	for i, r := range refs {
		cmds[i] = snapshotScript.Run(ctx, pipe, []string{s.P(r), s.N(r)})
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		if anyNoScript(cmds) {
			if lerr := s.LoadScripts(ctx); lerr == nil {
				return s.Snapshots(ctx, refs)
			}
		}
		return nil, err
	}
	out := make([]Snapshot, len(refs))
	for i, cmd := range cmds {
		v, err := cmd.Slice()
		if err != nil || len(v) < 2 {
			continue
		}
		vs, _ := v[0].(string)
		out[i].Version, _ = strconv.ParseInt(vs, 10, 64)
		list, _ := v[1].([]any)
		out[i].Entries = make(map[string]string, len(list)/2)
		for j := 0; j+1 < len(list); j += 2 {
			k, _ := list[j].(string)
			val, _ := list[j+1].(string)
			out[i].Entries[k] = val
		}
	}
	return out, nil
}

func (s *PresenceStore) List(ctx context.Context, refs []StreamRef) ([]map[string]string, error) {
	pipe := s.rdb.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(refs))
	for i, r := range refs {
		cmds[i] = pipe.HGetAll(ctx, s.P(r))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}
	out := make([]map[string]string, len(refs))
	for i, c := range cmds {
		out[i], _ = c.Result()
	}
	return out, nil
}

func (s *PresenceStore) Fields(ctx context.Context, refs []StreamRef) ([][]string, error) {
	pipe := s.rdb.Pipeline()
	cmds := make([]*redis.StringSliceCmd, len(refs))
	for i, r := range refs {
		cmds[i] = pipe.HKeys(ctx, s.P(r))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}
	out := make([][]string, len(refs))
	for i, c := range cmds {
		out[i], _ = c.Result()
	}
	return out, nil
}

func (s *PresenceStore) Count(ctx context.Context, r StreamRef) (int64, error) {
	return s.rdb.HLen(ctx, s.P(r)).Result()
}

func (s *PresenceStore) Exists(ctx context.Context, refs []StreamRef) ([]bool, error) {
	pipe := s.rdb.Pipeline()
	cmds := make([]*redis.IntCmd, len(refs))
	for i, r := range refs {
		cmds[i] = pipe.Exists(ctx, s.P(r))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}
	out := make([]bool, len(refs))
	for i, c := range cmds {
		n, _ := c.Result()
		out[i] = n > 0
	}
	return out, nil
}

func (s *PresenceStore) Nodes(ctx context.Context, r StreamRef) ([]string, error) {
	keys, err := s.rdb.HKeys(ctx, s.N(r)).Result()
	if err != nil {
		return nil, err
	}
	out := keys[:0]
	for _, k := range keys {
		if k != "_v" {
			out = append(out, k)
		}
	}
	return out, nil
}

func (s *PresenceStore) Follow(ctx context.Context, node, index string, added, removed []string) error {
	if len(added) == 0 && len(removed) == 0 {
		return nil
	}
	pipe := s.rdb.Pipeline()
	for _, u := range added {
		pipe.HSet(ctx, s.FollowersKey("u:"+u), node, "1")
	}
	for _, u := range removed {
		pipe.HDel(ctx, s.FollowersKey("u:"+u), node)
	}
	if index != "" {
		if len(added) > 0 {
			members := make([]any, len(added))
			for i, u := range added {
				members[i] = u
			}
			pipe.SAdd(ctx, index, members...)
		}
		if len(removed) > 0 {
			members := make([]any, len(removed))
			for i, u := range removed {
				members[i] = u
			}
			pipe.SRem(ctx, index, members...)
		}
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *PresenceStore) ScanSet(ctx context.Context, key string, batch int64, fn func([]string) error) error {
	var cursor uint64
	for {
		members, next, err := s.rdb.SScan(ctx, key, cursor, "", batch).Result()
		if err != nil {
			return err
		}
		if len(members) > 0 {
			if err := fn(members); err != nil {
				return err
			}
		}
		if next == 0 {
			return nil
		}
		cursor = next
	}
}

func (s *PresenceStore) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	pipe := s.rdb.Pipeline()
	for _, k := range keys {
		pipe.Del(ctx, k)
	}
	_, err := pipe.Exec(ctx)
	return err
}
