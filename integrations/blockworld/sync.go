package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	syncLease   = "blockworld-moderation-sync"
	refundLease = "blockworld-purchase-refund"
)

type restrictionChange struct {
	sourceKey string
	userID    string
	kind      string
	state     string
	until     int64
	reason    string
}

func (s *service) syncLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.syncEvery)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	tick := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		tick++
		ok, err := clusterkit.AcquireLease(ctx, s.db, syncLease, holder, 10*time.Second)
		if err == nil && ok {
			if _, err := s.syncRestrictions(ctx); err != nil && ctx.Err() == nil {
				s.logger.Warn("BlockWorld moderation sync failed: %v", err)
			}
		}
		if tick%40 == 0 {
			if ok, err := clusterkit.AcquireLease(ctx, s.db, refundLease, holder, time.Minute); err == nil && ok {
				if err := s.refundOrphans(ctx); err != nil && ctx.Err() == nil {
					s.logger.Warn("BlockWorld purchase refund sweep failed: %v", err)
				}
			}
		}
	}
}

func (s *service) pendingRestrictions(ctx context.Context) ([]restrictionChange, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT x.source_key, x.user_id, x.kind, x.state, x.until_ms, x.reason FROM (
			SELECT 'r:' || r.id::text AS source_key, r.user_id::text AS user_id, r.kind,
				CASE WHEN r.status = 'active' AND r.ends_at > now() THEN 'active' ELSE 'ended' END AS state,
				(EXTRACT(EPOCH FROM r.ends_at) * 1000)::bigint AS until_ms, r.reason::text AS reason
			FROM base_restriction r JOIN blockworld_link l ON l.user_id = r.user_id
			WHERE r.kind IN ('suspend', 'mute') AND r.created_at > now() - INTERVAL '400 days'
			UNION ALL
			SELECT 'ban:' || u.id::text, u.id::text, 'ban',
				CASE WHEN u.disable_time > '1970-01-01 00:00:00+00' THEN 'active' ELSE 'ended' END,
				0::bigint, '{}'
			FROM users u JOIN blockworld_link l ON l.user_id = u.id
		) x LEFT JOIN blockworld_restriction_sync s ON s.source_key = x.source_key
		WHERE (s.source_key IS NULL AND x.state = 'active') OR (s.source_key IS NOT NULL AND s.state <> x.state)
		LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []restrictionChange
	for rows.Next() {
		var c restrictionChange
		if err := rows.Scan(&c.sourceKey, &c.userID, &c.kind, &c.state, &c.until, &c.reason); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func reasonJSON(raw string) string {
	text := localized.Parse(raw)
	clean := map[string]string{}
	for k, v := range text {
		if v = strings.TrimSpace(v); v != "" {
			if r := []rune(v); len(r) > 120 {
				v = string(r[:120])
			}
			clean[k] = v
		}
	}
	if len(clean) == 0 {
		return ""
	}
	b, _ := json.Marshal(clean)
	return string(b)
}

func (s *service) syncRestrictions(ctx context.Context) (int, error) {
	changes, err := s.pendingRestrictions(ctx)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, c := range changes {
		reason := reasonJSON(c.reason)
		payload, _ := json.Marshal(map[string]any{"key": c.sourceKey, "kind": c.kind, "active": c.state == "active", "untilMs": c.until, "reason": reason})
		err := s.inTx(ctx, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, `INSERT INTO blockworld_restriction_sync (source_key, user_id, state) VALUES ($1, $2::uuid, $3)
				ON CONFLICT (source_key) DO UPDATE SET state = EXCLUDED.state, synced_at = now()`, c.sourceKey, c.userID, c.state); err != nil {
				return err
			}
			_, err := tx.ExecContext(ctx, `INSERT INTO blockworld_outbox (user_id, kind, dedupe_key, payload) VALUES ($1::uuid, 'restrict', $2, $3::jsonb) ON CONFLICT (dedupe_key) DO NOTHING`,
				c.userID, c.sourceKey+":"+c.state+":"+time.Now().UTC().Format("20060102T150405.000000"), string(payload))
			return err
		})
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

func (s *service) refundOrphans(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT l.user_id::text, l.metadata->>'sourceId', COALESCE(-(l.changeset->>$1)::bigint, 0) FROM wallet_ledger l
		WHERE l.metadata->>'module' = 'blockworld' AND l.metadata->>'reason' = 'block_purchase' AND l.create_time < now() - INTERVAL '2 minutes' AND l.create_time > now() - INTERVAL '7 days'
		AND NOT EXISTS (SELECT 1 FROM blockworld_purchase p WHERE p.id::text = l.metadata->>'sourceId')
		AND NOT EXISTS (SELECT 1 FROM wallet_ledger r WHERE r.user_id = l.user_id AND r.metadata->>'blockworldKey' = 'refund:' || (l.metadata->>'sourceId'))
		LIMIT 50`, s.cfg.currency)
	if err != nil {
		return err
	}
	type orphan struct {
		userID, purchaseID string
		amount             int64
	}
	var list []orphan
	for rows.Next() {
		var o orphan
		if err := rows.Scan(&o.userID, &o.purchaseID, &o.amount); err != nil {
			rows.Close()
			return err
		}
		list = append(list, o)
	}
	rows.Close()
	for _, o := range list {
		if o.amount <= 0 {
			continue
		}
		if _, _, err := s.nk.WalletUpdate(ctx, o.userID, map[string]int64{s.cfg.currency: o.amount}, map[string]any{
			"module": "blockworld", "reason": "purchase_refund", "sourceType": "blockworld_purchase", "sourceId": o.purchaseID, "blockworldKey": "refund:" + o.purchaseID,
		}, true); err != nil {
			return err
		}
		s.logger.Warn("BlockWorld refunded %d %s for purchase %s that never committed", o.amount, s.cfg.currency, o.purchaseID)
	}
	return nil
}
