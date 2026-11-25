package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"hash/fnv"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
	"github.com/heroiclabs/nakama/v3/internal/stealthbots"
)

const (
	botGrowSource = "stealth-grow"
	claimMaxAge   = 2 * time.Hour
)

func newRand() *rand.Rand {
	return rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
}

func seededRand(id string) *rand.Rand {
	h := fnv.New64a()
	_, _ = h.Write([]byte(id))
	v := h.Sum64()
	return rand.New(rand.NewPCG(v, v^0x9e3779b97f4a7c15))
}

func (s *service) growLook(ctx context.Context, tx *sql.Tx, level, maxLevel int, st stealthbots.Settings, r *rand.Rand) map[string]string {
	var raw string
	if tx.QueryRowContext(ctx, `SELECT look::text FROM economy_bot WHERE enabled AND look <> '{}'::jsonb ORDER BY random() LIMIT 1`).Scan(&raw) != nil {
		return map[string]string{}
	}
	slots := map[string]string{}
	_ = json.Unmarshal([]byte(raw), &slots)
	return stealthbots.LevelLook(ctx, tx, slots, level, maxLevel, st, r)
}

func (s *service) poolSize(ctx context.Context, q *sql.Tx) int {
	var n int
	_ = q.QueryRowContext(ctx, `SELECT count(*) FROM economy_bot WHERE enabled`).Scan(&n)
	return n
}

func (s *service) growBots(ctx context.Context, q gamecatalog.BotQuery, st stealthbots.Settings, need int) []gamecatalog.Bot {
	if need <= 0 || !st.Enabled || !st.PoolBots || !st.AutoGrow || st.PoolMax <= 0 {
		return nil
	}
	r := newRand()
	maxLevel := stealthbots.MaxLevel(ctx, s.db)
	out := []gamecatalog.Bot{}
	seeds := map[string]stealthbots.CardStats{}
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('economy-bot-grow'))`); err != nil {
			return err
		}
		room := st.PoolMax - s.poolSize(ctx, tx)
		if room <= 0 {
			return nil
		}
		if need > room {
			need = room
		}
		wallet := map[string]int64{}
		if q.Fee != nil && q.Fee.Amount > 0 && q.Fee.CurrencyKey != "" {
			wallet[q.Fee.CurrencyKey] = q.Fee.Amount * 20
		}
		walletJSON, _ := json.Marshal(wallet)
		for len(out) < need {
			level := min(stealthbots.GrowLevel(r, q.Level, q.LevelRange, q.MinLevel), 1000)
			created, updated := stealthbots.Age(r)
			id := uuid.Must(uuid.NewV4()).String()
			name, display := "", ""
			for try := 0; try < 8 && name == ""; try++ {
				u, d := stealthbots.Identity(r)
				if try >= 4 {
					u += string(rune('a'+r.IntN(26))) + string(rune('0'+r.IntN(10)))
				}
				res, err := tx.ExecContext(ctx, `INSERT INTO users (id, username, display_name, metadata, wallet, create_time, update_time)
					VALUES ($1,$2,NULLIF($3,''),'{}',$4::jsonb, now() - $5 * interval '1 second', now() - $6 * interval '1 second') ON CONFLICT (username) DO NOTHING`,
					id, u, d, string(walletJSON), int64(created.Seconds()), int64(updated.Seconds()))
				if err != nil {
					return err
				}
				if n, _ := res.RowsAffected(); n == 1 {
					name, display = u, d
				}
			}
			if name == "" {
				break
			}
			look := s.growLook(ctx, tx, level, maxLevel, st, r)
			lookJSON, _ := json.Marshal(look)
			if _, err := tx.ExecContext(ctx, `INSERT INTO economy_bot (user_id, enabled, level, template, rooms, look, note, source, updated_by) VALUES ($1,true,$2,'','[]',$3::jsonb,'',$4,'stealth')`,
				id, level, string(lookJSON), botGrowSource); err != nil {
				return err
			}
			if len(wallet) > 0 {
				meta, _ := json.Marshal(map[string]any{"module": "economy", "reason": "bot_wallet", "sourceType": "bot", "grantId": "bot-create:" + id})
				if _, err := tx.ExecContext(ctx, `INSERT INTO wallet_ledger (id, user_id, changeset, metadata) VALUES (gen_random_uuid(), $1, $2, $3)`, id, string(walletJSON), string(meta)); err != nil {
					return err
				}
			}
			if st.GrowStats && q.Game != "" {
				cs := stealthbots.SeedStats(r, level)
				seeds[id] = cs
				if _, err := tx.ExecContext(ctx, `INSERT INTO economy_game_stats (user_id, game_key, games, wins, losses, playtime_seconds, last_played_at, updated_at)
					VALUES ($1,$2,$3,$4,$5,$6, now() - $7 * interval '1 hour', now()) ON CONFLICT (user_id, game_key) DO NOTHING`,
					id, q.Game, cs.Games, cs.Wins, cs.Losses, cs.PlaytimeSeconds, cs.LastPlayedHours); err != nil {
					return err
				}
			}
			if q.MatchID != "" {
				if _, err := tx.ExecContext(ctx, `INSERT INTO economy_bot_lease AS l (user_id, match_id, game_key, busy, leased_at) VALUES ($1::uuid,$2,$3,true,now())
					ON CONFLICT (user_id) DO UPDATE SET match_id=EXCLUDED.match_id, game_key=EXCLUDED.game_key, busy=true, leased_at=now()`, id, q.MatchID, q.Game); err != nil {
					return err
				}
			}
			b := gamecatalog.Bot{UserID: id, Username: name, Name: display, Level: level, Look: look}
			if b.Name == "" {
				b.Name = name
			}
			out = append(out, b)
		}
		return nil
	})
	if err != nil {
		s.logger.Warn("Economy stealth pool grow for %s failed: %v", q.Game, err)
		return nil
	}
	if len(out) > 0 {
		s.logger.Info("Economy stealth pool grew by %d for %s", len(out), q.Game)
	}
	s.seedTrophies(ctx, seeds)
	return out
}

func (s *service) seedTrophies(ctx context.Context, seeds map[string]stealthbots.CardStats) {
	if len(seeds) == 0 {
		return
	}
	var ok bool
	if s.db.QueryRowContext(ctx, `SELECT to_regclass('trophy_player') IS NOT NULL`).Scan(&ok) != nil || !ok {
		return
	}
	for id, cs := range seeds {
		draws := cs.Games / 12
		if _, err := s.db.ExecContext(ctx, `INSERT INTO trophy_player (user_id, trophies, best, season_best, wins, losses, draws) VALUES ($1,$2,$3,$2,$4,$5,$6) ON CONFLICT (user_id) DO NOTHING`,
			id, cs.Trophies, cs.Best, cs.Wins, max(0, cs.Losses-draws), draws); err != nil {
			s.logger.Debug("Economy stealth trophy seed for %s skipped: %v", id, err)
			return
		}
	}
}

func (p gameProvider) KeepGhost(ctx context.Context, game string, b gamecatalog.Bot) error {
	return p.s.keepGhost(ctx, game, b)
}

func (s *service) keepGhost(ctx context.Context, game string, b gamecatalog.Bot) error {
	if _, err := uuid.FromString(b.UserID); err != nil {
		return badRequest("INVALID_ID", "userId", "must be a UUID")
	}
	look := b.Look
	if look == nil {
		look = map[string]string{}
	}
	raw, _ := json.Marshal(look)
	name := b.Name
	if name == b.Username {
		name = ""
	}
	created, _ := stealthbots.Age(seededRand(b.UserID))
	_, err := s.db.ExecContext(ctx, `INSERT INTO economy_bot_ghost (user_id, username, display_name, avatar_url, level, look, game_key, create_time)
		VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7, now() - $8 * interval '1 second') ON CONFLICT (user_id) DO UPDATE SET level=EXCLUDED.level, look=EXCLUDED.look, created_at=now()`,
		b.UserID, truncate(b.Username, 128), truncate(name, 255), truncate(b.Avatar, 512), max(1, b.Level), string(raw), game, int64(created.Seconds()))
	return err
}

func (s *service) ghostCard(ctx context.Context, viewer, target, game, table, lang string) (map[string]any, error) {
	var username, display, avatar, look, ghostGame string
	var level int
	err := s.db.QueryRowContext(ctx, `SELECT username, display_name, avatar_url, level, look::text, game_key FROM economy_bot_ghost WHERE user_id::text=$1`, target).Scan(&username, &display, &avatar, &level, &look, &ghostGame)
	if err != nil {
		return nil, notFound("USER_NOT_FOUND")
	}
	langs, l, err := s.gameLangs(ctx, viewer, lang)
	if err != nil {
		return nil, err
	}
	slots := map[string]string{}
	_ = json.Unmarshal([]byte(look), &slots)
	equipped := s.equippedView(ctx, slots, langs, l)
	r := seededRand(target)
	cs := stealthbots.SeedStats(r, level)
	key := game
	if key == "" {
		key = ghostGame
	}
	last := time.Now().UTC().Add(-time.Duration(cs.LastPlayedHours%3) * time.Hour).Format("2006-01-02T15:04:05Z")
	v := gameStatsView{Game: key, Games: cs.Games, Wins: cs.Wins, Losses: cs.Losses, PlaytimeSeconds: cs.PlaytimeSeconds, Winnings: map[string]int64{}, LastPlayedAt: &last}
	v.WinRate, v.LossRate = rate(v.Wins, v.Games), rate(v.Losses, v.Games)
	stats := []gameStatsView{}
	if key != "" {
		stats = append(stats, v)
	}
	if display == "" {
		display = username
	}
	card := map[string]any{"userId": target, "username": username, "displayName": display, "avatarUrl": avatar, "equipped": equipped,
		"level": s.levelView(ctx, target, table, level), "stats": stats, "lang": l}
	if game != "" && len(stats) == 1 {
		card["gameStats"] = stats[0]
	}
	return card, nil
}

func (p gameProvider) OneMatch(ctx context.Context, userID, game string, create func() (string, error)) (string, bool, error) {
	return p.s.oneMatch(ctx, userID, game, create)
}

func (s *service) oneMatch(ctx context.Context, userID, game string, create func() (string, error)) (string, bool, error) {
	st := s.stealthOf(ctx, game)
	if _, err := uuid.FromString(userID); err != nil || !st.Enabled || !st.OneLiveMatch {
		id, err := create()
		return id, err == nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('economy-claim:' || $1 || ':' || $2))`, userID, game); err != nil {
		return "", false, err
	}
	var matchID string
	var at time.Time
	err = tx.QueryRowContext(ctx, `SELECT match_id, created_at FROM economy_game_claim WHERE user_id=$1 AND game_key=$2`, userID, game).Scan(&matchID, &at)
	if err == nil && s.matchLive(ctx, matchID, userID, at) {
		return matchID, false, tx.Commit()
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", false, err
	}
	id, err := create()
	if err != nil {
		return "", false, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO economy_game_claim (user_id, game_key, match_id, created_at) VALUES ($1,$2,$3,now())
		ON CONFLICT (user_id, game_key) DO UPDATE SET match_id=EXCLUDED.match_id, created_at=now()`, userID, game, id); err != nil {
		return id, true, err
	}
	return id, true, tx.Commit()
}

func (s *service) liveClaim(ctx context.Context, userID, game string) string {
	var matchID string
	var at time.Time
	if s.db.QueryRowContext(ctx, `SELECT match_id, created_at FROM economy_game_claim WHERE user_id::text=$1 AND game_key=$2`, userID, game).Scan(&matchID, &at) != nil {
		return ""
	}
	if s.matchLive(ctx, matchID, userID, at) {
		return matchID
	}
	return ""
}

func (s *service) matchLive(ctx context.Context, matchID, userID string, at time.Time) bool {
	if matchID == "" || time.Since(at) > claimMaxAge || s.nk == nil {
		return false
	}
	m, err := s.nk.MatchGet(ctx, matchID)
	if err != nil || m == nil {
		return false
	}
	var status string
	if s.db.QueryRowContext(ctx, `SELECT status FROM economy_game_match WHERE id=$1`, matchID).Scan(&status) == nil && (status == "finished" || status == "cancelled") {
		return false
	}
	var entry string
	if s.db.QueryRowContext(ctx, `SELECT status FROM economy_game_entry WHERE match_id=$1 AND user_id::text=$2`, matchID, userID).Scan(&entry) == nil && (entry == "left" || entry == "refunded") {
		return false
	}
	return true
}

func (p gameProvider) RecordStats(ctx context.Context, matchID string, results []gamecatalog.Result) error {
	return providerError(p.s.recordStats(ctx, matchID, results))
}

func statsStatus(v string) bool {
	switch v {
	case gamecatalog.ResultWon, gamecatalog.ResultLost, gamecatalog.ResultDraw, gamecatalog.ResultForfeit, gamecatalog.ResultTimeout:
		return true
	}
	return false
}

func (s *service) recordStats(ctx context.Context, matchID string, results []gamecatalog.Result) error {
	matchID, err := cleanMatchID(matchID)
	if err != nil {
		return err
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		var game string
		var started, ended sql.NullTime
		err := tx.QueryRowContext(ctx, `UPDATE economy_game_match SET stats_done=true WHERE id=$1 AND NOT stats_done AND status <> 'cancelled' RETURNING game_key, started_at, ended_at`, matchID).Scan(&game, &started, &ended)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		st := s.stealthOf(ctx, game)
		played := int64(0)
		if started.Valid {
			end := s.now()
			if ended.Valid {
				end = ended.Time
			}
			played = max(0, int64(end.Sub(started.Time).Seconds()))
		}
		seen := map[string]bool{}
		for _, r := range results {
			if _, err := uuid.FromString(r.UserID); err != nil || seen[r.UserID] || !statsStatus(strings.TrimSpace(r.Status)) {
				continue
			}
			seen[r.UserID] = true
			var bot bool
			err := tx.QueryRowContext(ctx, `SELECT bot FROM economy_game_entry WHERE match_id=$1 AND user_id=$2 AND status IN ('paid','free')`, matchID, r.UserID).Scan(&bot)
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil {
				return err
			}
			if bot && !(st.Enabled && st.Progress) {
				continue
			}
			if err := s.addGameStats(ctx, tx, r.UserID, game, strings.TrimSpace(r.Status), played, "", 0); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *service) sweepStealth(ctx context.Context) {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM economy_bot_ghost g WHERE g.created_at < now() - make_interval(days => COALESCE((SELECT GREATEST(1, LEAST(3650, COALESCE((e.stealth->>'ghostDays')::int, 90))) FROM economy_game e WHERE e.key=g.game_key), 90))`); err != nil {
		s.logger.Warn("Economy ghost sweep failed: %v", err)
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM economy_game_claim WHERE created_at < now() - interval '1 day'`); err != nil {
		s.logger.Warn("Economy claim sweep failed: %v", err)
	}
}
