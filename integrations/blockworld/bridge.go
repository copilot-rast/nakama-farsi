package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
)

const (
	eventGameWon = "blockworld_game_won"
	flagPrefix   = "blockworld.feature."
)

var (
	eventKeyRe     = regexp.MustCompile(`^[a-z0-9:_\-]{3,160}$`)
	legacyRe       = regexp.MustCompile(`^[0-9a-f]{64}$`)
	errBadEvent    = runtime.NewError("INVALID_EVENT", 3)
	errBadUser     = runtime.NewError("INVALID_USER", 3)
	errUnavailable = runtime.NewError("EVENTS_UNAVAILABLE", 14)
)

type action struct {
	ID      int64           `json:"id"`
	UserID  string          `json:"userId"`
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload"`
	Attempt int             `json:"attempt"`
}

func (s *service) rpcBridgePull(ctx context.Context, _ string, payload string) (any, error) {
	var in struct {
		Limit int `json:"limit"`
	}
	if err := decode(payload, &in); err != nil {
		return nil, err
	}
	if in.Limit < 1 || in.Limit > 200 {
		in.Limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `UPDATE blockworld_outbox o SET delivered_at = now(), attempts = o.attempts + 1
		WHERE o.id IN (SELECT id FROM blockworld_outbox WHERE acked_at IS NULL AND (delivered_at IS NULL OR delivered_at < now() - $2::INT8 * INTERVAL '1 millisecond')
			ORDER BY id LIMIT $1 FOR UPDATE SKIP LOCKED)
		RETURNING o.id, o.user_id::text, o.kind, o.payload::text, o.attempts`, in.Limit, s.cfg.redelivery.Milliseconds())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []action{}
	for rows.Next() {
		var a action
		var raw string
		if err := rows.Scan(&a.ID, &a.UserID, &a.Kind, &raw, &a.Attempt); err != nil {
			return nil, err
		}
		a.Payload = json.RawMessage(raw)
		items = append(items, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortActions(items)
	return map[string]any{"items": items}, nil
}

func sortActions(items []action) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].ID < items[j-1].ID; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}

func (s *service) rpcBridgeAck(ctx context.Context, _ string, payload string) (any, error) {
	var in struct {
		IDs []int64 `json:"ids"`
	}
	if err := decode(payload, &in); err != nil {
		return nil, err
	}
	if len(in.IDs) == 0 {
		return map[string]any{"acked": 0}, nil
	}
	if len(in.IDs) > 500 {
		return nil, runtime.NewError("TOO_MANY_IDS", 3)
	}
	res, err := s.db.ExecContext(ctx, `UPDATE blockworld_outbox SET acked_at = now() WHERE id = ANY($1::bigint[]) AND acked_at IS NULL`, int64Array(in.IDs))
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	return map[string]any{"acked": n}, nil
}

func int64Array(ids []int64) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprint(id)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

type report struct {
	EventKey string          `json:"eventKey"`
	Kind     string          `json:"kind"`
	UserID   string          `json:"userId"`
	Data     json.RawMessage `json:"data"`
}

func (s *service) rpcBridgeReport(ctx context.Context, _ string, payload string) (any, error) {
	var in report
	if err := decode(payload, &in); err != nil {
		return nil, err
	}
	if !eventKeyRe.MatchString(in.EventKey) {
		return nil, errBadEvent
	}
	if _, err := uuid.FromString(in.UserID); err != nil {
		return nil, errBadUser
	}
	switch in.Kind {
	case "game_won":
		return s.reportGameWon(ctx, in)
	case "wallet_migration":
		return s.reportMigration(ctx, in)
	}
	return nil, errBadEvent
}

func (s *service) userExists(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, userID string) (bool, error) {
	var ok bool
	err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1::uuid)`, userID).Scan(&ok)
	return ok, err
}

func (s *service) reportGameWon(ctx context.Context, in report) (any, error) {
	var data struct {
		GameKind  int    `json:"gameKind"`
		GameName  string `json:"gameName"`
		ResultID  string `json:"resultId"`
		SessionID string `json:"sessionId"`
		WorldID   string `json:"worldId"`
	}
	if err := json.Unmarshal(in.Data, &data); err != nil {
		return nil, errBadEvent
	}
	if !playerevents.Available() {
		return nil, errUnavailable
	}
	ok, err := s.userExists(ctx, s.db, in.UserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		if _, err := s.db.ExecContext(ctx, `INSERT INTO blockworld_world_event (event_key, kind, user_id, status, payload) VALUES ($1, 'game_won', NULL, 'skipped', $2::jsonb) ON CONFLICT DO NOTHING`, in.EventKey, string(in.Data)); err != nil {
			return nil, err
		}
		return map[string]any{"status": "skipped", "reason": "USER_NOT_FOUND"}, nil
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO blockworld_world_event (event_key, kind, user_id, status, payload, created_at, updated_at) VALUES ($1, 'game_won', $2::uuid, 'recording', $3::jsonb, now(), now() - INTERVAL '1 hour') ON CONFLICT DO NOTHING`, in.EventKey, in.UserID, string(in.Data)); err != nil {
		return nil, err
	}
	var status string
	err = s.db.QueryRowContext(ctx, `UPDATE blockworld_world_event SET updated_at = now() WHERE event_key = $1 AND status = 'recording' AND updated_at < now() - INTERVAL '2 minutes' RETURNING status`, in.EventKey).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		if err := s.db.QueryRowContext(ctx, `SELECT status FROM blockworld_world_event WHERE event_key = $1`, in.EventKey).Scan(&status); err != nil {
			return nil, err
		}
		if status == "recording" {
			return nil, runtime.NewError("EVENT_IN_PROGRESS", 10)
		}
		return map[string]any{"status": status, "duplicate": true}, nil
	}
	if err != nil {
		return nil, err
	}
	bg := context.WithoutCancel(ctx)
	var already bool
	if err := s.db.QueryRowContext(bg, `SELECT EXISTS (SELECT 1 FROM economy_player_event e JOIN economy_event_type t ON t.id = e.type_id
		WHERE e.user_id = $1::uuid AND t.key = $2 AND e.metadata->>'eventKey' = $3)`, in.UserID, eventGameWon, in.EventKey).Scan(&already); err != nil {
		return nil, err
	}
	recorded := map[string]any{"recovered": already}
	if !already {
		res, err := playerevents.Record(bg, playerevents.Event{UserID: in.UserID, Key: eventGameWon, Value: 1, Metadata: map[string]any{
			"eventKey": in.EventKey, "gameKind": data.GameKind, "game": data.GameName, "resultId": data.ResultID, "worldId": data.WorldID,
		}})
		if err != nil {
			_, _ = s.db.ExecContext(bg, `UPDATE blockworld_world_event SET updated_at = now() - INTERVAL '1 hour' WHERE event_key = $1 AND status = 'recording'`, in.EventKey)
			return nil, err
		}
		recorded = summarize(res)
	}
	raw, _ := json.Marshal(recorded)
	if _, err := s.db.ExecContext(bg, `UPDATE blockworld_world_event SET status = 'done', result = $2::jsonb, updated_at = now() WHERE event_key = $1`, in.EventKey, string(raw)); err != nil {
		return nil, err
	}
	return map[string]any{"status": "done", "duplicate": false, "result": recorded}, nil
}

func summarize(res playerevents.Result) map[string]any {
	out := map[string]any{}
	for _, k := range []string{"eventId", "claims", "rules"} {
		if v, ok := res[k]; ok {
			out[k] = v
		}
	}
	return out
}

func (s *service) reportMigration(ctx context.Context, in report) (any, error) {
	var data struct {
		Legacy  string `json:"legacyIdentity"`
		Coins   int64  `json:"coins"`
		Blocks  []int  `json:"blocks"`
		WorldID string `json:"worldId"`
	}
	if err := json.Unmarshal(in.Data, &data); err != nil {
		return nil, errBadEvent
	}
	data.Legacy = strings.ToLower(strings.TrimPrefix(data.Legacy, "0x"))
	if !legacyRe.MatchString(data.Legacy) || data.Coins < 0 || data.Coins > 1_000_000_000 || len(data.Blocks) > 256 {
		return nil, errBadEvent
	}
	migrationKey := "migration:" + data.Legacy
	ok, err := s.userExists(ctx, s.db, in.UserID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errUserNotFound
	}
	duplicate := false
	owner := in.UserID
	coins := data.Coins
	unlocked := []int{}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `INSERT INTO blockworld_migration (legacy_identity, user_id, coins, blocks) VALUES ($1, $2::uuid, $3, $4::int[]) ON CONFLICT DO NOTHING`,
			data.Legacy, in.UserID, data.Coins, intArray(data.Blocks))
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			duplicate = true
			return tx.QueryRowContext(ctx, `SELECT user_id::text, coins FROM blockworld_migration WHERE legacy_identity = $1`, data.Legacy).Scan(&owner, &coins)
		}
		for _, b := range data.Blocks {
			var known bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM blockworld_block WHERE id = $1)`, b).Scan(&known); err != nil {
				return err
			}
			if !known {
				continue
			}
			r, err := tx.ExecContext(ctx, `INSERT INTO blockworld_purchase (id, user_id, block_id, price, source, source_ref) VALUES ($1::uuid, $2::uuid, $3, 0, 'migration', $4) ON CONFLICT DO NOTHING`,
				uuid.Must(uuid.NewV4()).String(), in.UserID, b, migrationKey)
			if err != nil {
				return err
			}
			if n, _ := r.RowsAffected(); n == 0 {
				continue
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO blockworld_outbox (user_id, kind, dedupe_key, payload) VALUES ($1::uuid, 'unlock', $2, jsonb_build_object('block', $3::int, 'source', 'migration')) ON CONFLICT DO NOTHING`,
				in.UserID, unlockKey(in.UserID, b), b); err != nil {
				return err
			}
			unlocked = append(unlocked, b)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO blockworld_world_event (event_key, kind, user_id, status, payload) VALUES ($1, 'wallet_migration', $2::uuid, 'done', $3::jsonb) ON CONFLICT DO NOTHING`, in.EventKey, in.UserID, string(in.Data))
		return err
	})
	if err != nil {
		return nil, err
	}
	bg := context.WithoutCancel(ctx)
	var credited bool
	if err := s.db.QueryRowContext(bg, `SELECT EXISTS (SELECT 1 FROM wallet_ledger WHERE user_id = $1::uuid AND metadata->>'blockworldKey' = $2)`, owner, migrationKey).Scan(&credited); err != nil {
		return nil, err
	}
	if coins > 0 && !credited {
		if _, _, err := s.nk.WalletUpdate(bg, owner, map[string]int64{s.cfg.currency: coins}, map[string]any{
			"module": "blockworld", "reason": "wallet_migration", "sourceType": "blockworld_migration", "sourceId": data.Legacy, "blockworldKey": migrationKey,
		}, true); err != nil {
			return nil, err
		}
	}
	if !duplicate {
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: in.UserID, Kind: activity.KindPurchase, Action: "blockworld/migration", RefID: data.Legacy,
			Data: map[string]any{"coins": data.Coins, "blocks": data.Blocks}})
	}
	return map[string]any{"status": "done", "duplicate": duplicate, "userId": owner, "coins": coins, "unlocked": unlocked}, nil
}

func intArray(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprint(id)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func (s *service) rpcBridgeFlags(ctx context.Context, _ string, payload string) (any, error) {
	var in struct{}
	if err := decode(payload, &in); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT key, value::text, version FROM base_config_key WHERE is_active AND type = 'bool' AND key LIKE $1 ORDER BY key`, flagPrefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	flags := map[string]bool{}
	versions := map[string]int{}
	for rows.Next() {
		var key, raw string
		var version int
		if err := rows.Scan(&key, &raw, &version); err != nil {
			return nil, err
		}
		var on bool
		if json.Unmarshal([]byte(raw), &on) != nil {
			continue
		}
		name := strings.TrimPrefix(key, flagPrefix)
		flags[name] = on
		versions[name] = version
	}
	return map[string]any{"flags": flags, "versions": versions}, rows.Err()
}
