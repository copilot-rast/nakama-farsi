package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

type coreBoards interface {
	Create(ctx context.Context, id, sortOrder, operator, schedule string, metadata map[string]any) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*api.Leaderboard, error)
	Write(ctx context.Context, id, owner, username string, score, subscore int64, override *int) (*api.LeaderboardRecord, error)
	RecordDelete(ctx context.Context, id, owner string) error
	List(ctx context.Context, id string, owners []string, limit int, cursor string) ([]*api.LeaderboardRecord, []*api.LeaderboardRecord, string, error)
	Haystack(ctx context.Context, id, owner string, limit int) ([]*api.LeaderboardRecord, error)
}

type nkBoards struct{ nk runtime.NakamaModule }

func (c nkBoards) Create(ctx context.Context, id, sortOrder, operator, schedule string, metadata map[string]any) error {
	return c.nk.LeaderboardCreate(ctx, id, true, sortOrder, operator, schedule, metadata, true)
}

func (c nkBoards) Delete(ctx context.Context, id string) error {
	return c.nk.LeaderboardDelete(ctx, id)
}

func (c nkBoards) Get(ctx context.Context, id string) (*api.Leaderboard, error) {
	list, err := c.nk.LeaderboardsGetId(ctx, []string{id})
	if err != nil || len(list) == 0 {
		return nil, err
	}
	return list[0], nil
}

func (c nkBoards) Write(ctx context.Context, id, owner, username string, score, subscore int64, override *int) (*api.LeaderboardRecord, error) {
	return c.nk.LeaderboardRecordWrite(ctx, id, owner, username, score, subscore, nil, override)
}

func (c nkBoards) RecordDelete(ctx context.Context, id, owner string) error {
	return c.nk.LeaderboardRecordDelete(ctx, id, owner)
}

func (c nkBoards) List(ctx context.Context, id string, owners []string, limit int, cursor string) ([]*api.LeaderboardRecord, []*api.LeaderboardRecord, string, error) {
	recs, own, next, _, err := c.nk.LeaderboardRecordsList(ctx, id, owners, limit, cursor, 0)
	return recs, own, next, err
}

func (c nkBoards) Haystack(ctx context.Context, id, owner string, limit int) ([]*api.LeaderboardRecord, error) {
	list, err := c.nk.LeaderboardRecordsHaystack(ctx, id, owner, limit, "", 0)
	if err != nil || list == nil {
		return nil, err
	}
	return list.Records, nil
}

func expiry(cron string, now time.Time) time.Time {
	_, end := periodBounds(cron, now)
	if end == nil {
		return time.Unix(0, 0).UTC()
	}
	return *end
}

func countPlayers(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, key string, exp time.Time) int {
	var n int
	_ = q.QueryRowContext(ctx, `SELECT count(*) FROM leaderboard_record WHERE leaderboard_id=$1 AND expiry_time=$2`, key, exp).Scan(&n)
	return n
}

func coreOrder(sortOrder string) string {
	if sortOrder == "asc" {
		return "score ASC, subscore ASC, owner_id ASC"
	}
	return "score DESC, subscore DESC, owner_id DESC"
}

func rankOf(ctx context.Context, db *sql.DB, key, sortOrder string, exp time.Time, userID string) (int, int64, int64, bool) {
	var score, sub int64
	if err := db.QueryRowContext(ctx, `SELECT score, subscore FROM leaderboard_record WHERE leaderboard_id=$1 AND expiry_time=$2 AND owner_id=$3`, key, exp, userID).Scan(&score, &sub); err != nil {
		return 0, 0, 0, false
	}
	cmp := `(score, subscore, owner_id) > ($3, $4, $5::uuid)`
	if sortOrder == "asc" {
		cmp = `(score, subscore, owner_id) < ($3, $4, $5::uuid)`
	}
	var ahead int
	_ = db.QueryRowContext(ctx, `SELECT count(*) FROM leaderboard_record WHERE leaderboard_id=$1 AND expiry_time=$2 AND `+cmp, key, exp, score, sub, userID).Scan(&ahead)
	return ahead + 1, score, sub, true
}
