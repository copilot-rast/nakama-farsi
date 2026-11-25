package stealthbots

import (
	"context"
	"database/sql"
	"hash/fnv"
	"strconv"
	"sync"
	"time"

	"github.com/heroiclabs/nakama-common/api"
)

type DB interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

const idleWindow = 20 * time.Minute

var pool struct {
	sync.Mutex
	at      time.Time
	on      bool
	percent int
}

func Reset() {
	pool.Lock()
	pool.at = time.Time{}
	pool.Unlock()
}

func poolState(ctx context.Context, db DB) (bool, int) {
	pool.Lock()
	defer pool.Unlock()
	if !pool.at.IsZero() && time.Since(pool.at) < 30*time.Second {
		return pool.on, pool.percent
	}
	var on bool
	var pct int
	err := db.QueryRowContext(ctx, `SELECT COALESCE(bool_or(COALESCE((stealth->>'enabled')::boolean,false)),false),
		COALESCE(max(COALESCE((stealth->>'onlineIdlePercent')::int, 4)) FILTER (WHERE COALESCE((stealth->>'enabled')::boolean,false)), 0) FROM economy_game`).Scan(&on, &pct)
	if err != nil {
		on, pct = false, 0
	}
	pool.at, pool.on, pool.percent = time.Now(), on, pct
	return on, pct
}

func Active(ctx context.Context, db DB) bool {
	on, _ := poolState(ctx, db)
	return on
}

func idleOnline(id string, now time.Time, pct int) bool {
	if pct <= 0 {
		return false
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(id + ":" + strconv.FormatInt(now.Unix()/int64(idleWindow/time.Second), 10)))
	return int(h.Sum32()%100) < pct
}

func Online(ctx context.Context, db DB, ids []string) map[string]bool {
	out := map[string]bool{}
	if len(ids) == 0 {
		return out
	}
	on, pct := poolState(ctx, db)
	if !on {
		return out
	}
	rows, err := db.QueryContext(ctx, `SELECT b.user_id::text, COALESCE(l.busy,false), COALESCE(l.online_until > now(), false)
		FROM economy_bot b LEFT JOIN economy_bot_lease l ON l.user_id=b.user_id WHERE b.user_id::text = ANY($1::text[]) AND b.enabled`, ids)
	if err != nil {
		return out
	}
	defer rows.Close()
	now := time.Now()
	for rows.Next() {
		var id string
		var busy, recent bool
		if rows.Scan(&id, &busy, &recent) == nil {
			out[id] = busy || recent || idleOnline(id, now, pct)
		}
	}
	return out
}

func MarkUsers(ctx context.Context, db DB, users []*api.User) {
	ids := []string{}
	for _, u := range users {
		if u != nil && !u.Online {
			ids = append(ids, u.Id)
		}
	}
	if len(ids) == 0 {
		return
	}
	on := Online(ctx, db, ids)
	for _, u := range users {
		if u != nil && on[u.Id] {
			u.Online = true
		}
	}
}

func MarkFriends(ctx context.Context, db DB, friends []*api.Friend) {
	users := make([]*api.User, 0, len(friends))
	for _, f := range friends {
		if f != nil && f.User != nil {
			users = append(users, f.User)
		}
	}
	MarkUsers(ctx, db, users)
}

func IsBot(ctx context.Context, db DB, userID string) bool {
	var ok bool
	_ = db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_bot WHERE user_id::text=$1)`, userID).Scan(&ok)
	return ok
}

func GuardWins(ctx context.Context, db DB, userID, game string) int {
	var n int
	_ = db.QueryRowContext(ctx, `SELECT COALESCE((SELECT wins FROM economy_bot_guard WHERE user_id::text=$1 AND game_key=$2 AND day=current_date),0)`, userID, game).Scan(&n)
	return n
}

func GuardAdd(ctx context.Context, db DB, userID, game string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO economy_bot_guard AS g (user_id, game_key, day, wins) VALUES ($1::uuid, $2, current_date, 1)
		ON CONFLICT (user_id, game_key, day) DO UPDATE SET wins=g.wins+1`, userID, game)
	return err
}

func (s Settings) BoostFor(ctx context.Context, db DB, userID, game string) int {
	if !s.Enabled || !s.Guard {
		return 0
	}
	return s.Boost(GuardWins(ctx, db, userID, game))
}
