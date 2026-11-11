package cluster

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Invalidation struct {
	Op         string
	UserIDs    []string
	SessionExp int64
	SessionID  string
	RefreshExp int64
	RefreshID  string
	Origin     string
	At         time.Time
}

type InvalidationLog struct {
	logger *zap.Logger
	rdb    redis.UniversalClient
	key    string
	keep   time.Duration
	origin string
}

func NewInvalidationLog(logger *zap.Logger, rdb redis.UniversalClient, key string, keep time.Duration, origin string) *InvalidationLog {
	if keep < time.Minute {
		keep = time.Minute
	}
	return &InvalidationLog{logger: logger, rdb: rdb, key: key, keep: keep, origin: origin}
}

func (l *InvalidationLog) Append(ctx context.Context, inv Invalidation) error {
	minID := strconv.FormatInt(time.Now().Add(-l.keep).UnixMilli(), 10)
	values := map[string]any{
		"op":   inv.Op,
		"o":    l.origin,
		"se":   inv.SessionExp,
		"sid":  inv.SessionID,
		"re":   inv.RefreshExp,
		"rid":  inv.RefreshID,
		"uids": joinIDs(inv.UserIDs),
	}
	return l.rdb.XAdd(ctx, &redis.XAddArgs{Stream: l.key, MinID: minID, Approx: true, Values: values}).Err()
}

func joinIDs(ids []string) string {
	out := ""
	for i, id := range ids {
		if i > 0 {
			out += ","
		}
		out += id
	}
	return out
}

func splitIDs(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if i > start {
				out = append(out, s[start:i])
			}
			start = i + 1
		}
	}
	return out
}

func parseInvalidation(msg redis.XMessage) Invalidation {
	str := func(k string) string {
		v, _ := msg.Values[k].(string)
		return v
	}
	num := func(k string) int64 {
		n, _ := strconv.ParseInt(str(k), 10, 64)
		return n
	}
	inv := Invalidation{Op: str("op"), Origin: str("o"), UserIDs: splitIDs(str("uids")), SessionExp: num("se"), SessionID: str("sid"), RefreshExp: num("re"), RefreshID: str("rid")}
	if i := strings.IndexByte(msg.ID, '-'); i > 0 {
		if ms, err := strconv.ParseInt(msg.ID[:i], 10, 64); err == nil {
			inv.At = time.UnixMilli(ms)
		}
	}
	return inv
}

func (l *InvalidationLog) Replay(ctx context.Context, apply func(Invalidation)) string {
	start := strconv.FormatInt(time.Now().Add(-l.keep).UnixMilli(), 10)
	last := "0"
	if msgs, err := l.rdb.XRange(ctx, l.key, start, "+").Result(); err == nil {
		for _, m := range msgs {
			apply(parseInvalidation(m))
			last = m.ID
		}
	} else {
		l.logger.Warn("Session invalidation replay failed", zap.Error(err))
	}
	if last == "0" {
		if id, err := l.lastID(ctx); err == nil && id != "" {
			last = id
		}
	}
	return last
}

func (l *InvalidationLog) Run(ctx context.Context, apply func(Invalidation)) {
	l.Tail(ctx, l.Replay(ctx, apply), apply)
}

func (l *InvalidationLog) Tail(ctx context.Context, last string, apply func(Invalidation)) {
	for ctx.Err() == nil {
		res, err := l.rdb.XRead(ctx, &redis.XReadArgs{Streams: []string{l.key, last}, Block: 5 * time.Second, Count: 500}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			if !sleepCtx(ctx, 500*time.Millisecond) {
				return
			}
			continue
		}
		for _, s := range res {
			for _, m := range s.Messages {
				apply(parseInvalidation(m))
				last = m.ID
			}
		}
	}
}

func (l *InvalidationLog) lastID(ctx context.Context) (string, error) {
	msgs, err := l.rdb.XRevRangeN(ctx, l.key, "+", "-", 1).Result()
	if err != nil || len(msgs) == 0 {
		return "0", err
	}
	return msgs[0].ID, nil
}
