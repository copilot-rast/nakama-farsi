package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"net/http"
	"time"

	"github.com/gofrs/uuid/v5"
)

const (
	chanceScale        = 10000
	chanceTotal        = 100 * chanceScale
	chanceEventKey     = "ChanceGame:Played"
	maxChanceRows      = 50
	maxGroupMembers    = 100
	maxChanceInterval  = 525600
	chanceRecoverAfter = 2 * time.Minute
)

type chanceRow struct {
	Chance       float64 `json:"chance"`
	PrizeID      *string `json:"prizeId"`
	PrizeGroupID *string `json:"prizeGroupId"`
}

type chancePrice struct {
	CurrencyID string `json:"currencyId"`
	Amount     int64  `json:"amount"`
}

type chanceGame struct {
	ID          string
	Key         string
	Name        i18n
	Description i18n
	Metadata    json.RawMessage
	PrizeType   string
	Rows        []chanceRow
	Interval    *int
	Price       *chancePrice
	Active      bool
	Sort        int
}

type chancePick struct {
	Roll    int
	Slot    int
	PrizeID string
	GroupID string
}

const chanceGameColumns = `id::text, key, name::text, description::text, metadata::text, prize_type, prize::text, interval_time, price::text, is_active, sort`

type scanner interface {
	Scan(dest ...any) error
}

func scanChanceGame(row scanner) (*chanceGame, error) {
	var g chanceGame
	var name, desc, meta, rows string
	var price sql.NullString
	var interval sql.NullInt64
	if err := row.Scan(&g.ID, &g.Key, &name, &desc, &meta, &g.PrizeType, &rows, &interval, &price, &g.Active, &g.Sort); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(name), &g.Name)
	_ = json.Unmarshal([]byte(desc), &g.Description)
	g.Metadata = json.RawMessage(meta)
	if err := json.Unmarshal([]byte(rows), &g.Rows); err != nil {
		return nil, err
	}
	if interval.Valid {
		v := int(interval.Int64)
		g.Interval = &v
	}
	if price.Valid && price.String != "" && price.String != "null" {
		var p chancePrice
		if err := json.Unmarshal([]byte(price.String), &p); err == nil && p.CurrencyID != "" && p.Amount > 0 {
			g.Price = &p
		}
	}
	return &g, nil
}

func loadChanceGame(ctx context.Context, q queryer, id string, lock string) (*chanceGame, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("CHANCE_GAME_NOT_FOUND")
	}
	g, err := scanChanceGame(q.QueryRowContext(ctx, `SELECT `+chanceGameColumns+` FROM economy_chance_game WHERE id=$1`+lock, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("CHANCE_GAME_NOT_FOUND")
	}
	return g, err
}

func loadChanceGames(ctx context.Context, q rowsQueryer, where string, args ...any) ([]*chanceGame, error) {
	rows, err := q.QueryContext(ctx, `SELECT `+chanceGameColumns+` FROM economy_chance_game`+where+` ORDER BY sort, created_at, id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*chanceGame
	for rows.Next() {
		g, err := scanChanceGame(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func scaledChance(v float64) int {
	return int(math.Round(v * chanceScale))
}

func secureInt(n int) (int, error) {
	if n <= 0 {
		return 0, errors.New("secure random range must be positive")
	}
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}
	return int(v.Int64()), nil
}

func pickSlot(rows []chanceRow, roll int) int {
	cum := 0
	for i, r := range rows {
		cum += scaledChance(r.Chance)
		if roll < cum {
			return i
		}
	}
	return -1
}

type groupMember struct {
	PrizeID string
	Weight  int
}

func loadGroupMembers(ctx context.Context, q rowsQueryer, groupID string) ([]groupMember, error) {
	rows, err := q.QueryContext(ctx, `SELECT prize_id::text, weight FROM economy_prize_group_member WHERE group_id=$1 ORDER BY sort, prize_id`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []groupMember
	for rows.Next() {
		var m groupMember
		if err := rows.Scan(&m.PrizeID, &m.Weight); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func drawGroupMember(members []groupMember) (string, error) {
	total := 0
	for _, m := range members {
		total += m.Weight
	}
	if total <= 0 {
		return "", badRequest("PRIZE_GROUP_EMPTY", "prizeGroupId", "the prize group has no prizes")
	}
	r, err := secureInt(total)
	if err != nil {
		return "", err
	}
	for _, m := range members {
		if r < m.Weight {
			return m.PrizeID, nil
		}
		r -= m.Weight
	}
	return members[len(members)-1].PrizeID, nil
}

func (s *service) drawFromPrizeGroup(ctx context.Context, q rowsQueryer, groupID string) (string, error) {
	members, err := loadGroupMembers(ctx, q, groupID)
	if err != nil {
		return "", err
	}
	return drawGroupMember(members)
}

func (s *service) pickChance(ctx context.Context, tx *sql.Tx, g *chanceGame, userID string) (chancePick, error) {
	p, err := s.drawChance(ctx, tx, g)
	if err != nil || p.PrizeID == "" {
		return p, err
	}
	var repeatable bool
	if err := tx.QueryRowContext(ctx, `SELECT is_repeatable FROM economy_prize WHERE id=$1`, p.PrizeID).Scan(&repeatable); err != nil {
		return p, err
	}
	if !repeatable {
		owned, err := prizeClaimedByUser(ctx, tx, userID, p.PrizeID)
		if err != nil {
			return p, err
		}
		if owned {
			p.PrizeID = ""
		}
	}
	return p, nil
}

func (s *service) drawChance(ctx context.Context, tx *sql.Tx, g *chanceGame) (chancePick, error) {
	roll, err := secureInt(chanceTotal)
	if err != nil {
		return chancePick{}, err
	}
	p := chancePick{Roll: roll, Slot: pickSlot(g.Rows, roll)}
	if p.Slot < 0 {
		return p, nil
	}
	row := g.Rows[p.Slot]
	if g.PrizeType == "prizeGroup" {
		if row.PrizeGroupID == nil || *row.PrizeGroupID == "" {
			return p, nil
		}
		p.GroupID = *row.PrizeGroupID
		p.PrizeID, err = s.drawFromPrizeGroup(ctx, tx, p.GroupID)
		return p, err
	}
	if row.PrizeID != nil {
		p.PrizeID = *row.PrizeID
	}
	return p, nil
}

func chanceWait(g *chanceGame, last sql.NullTime, now time.Time) time.Duration {
	if g.Interval == nil || !last.Valid {
		return 0
	}
	left := time.Duration(*g.Interval)*time.Minute - now.Sub(last.Time)
	if left < 0 {
		return 0
	}
	return left
}

func waitSeconds(d time.Duration) int64 {
	if d <= 0 {
		return 0
	}
	return int64(math.Ceil(d.Seconds()))
}

func lastFreePlay(ctx context.Context, q queryer, gameID, userID string) (sql.NullTime, error) {
	var last sql.NullTime
	err := q.QueryRowContext(ctx, `SELECT max(created_at) FROM economy_chance_play WHERE game_id=$1 AND user_id=$2 AND mode='free'`, gameID, userID).Scan(&last)
	return last, err
}

func waitingForTurn(left time.Duration) *appError {
	e := fail(http.StatusForbidden, "WAITING_FOR_TURN", "", "wait for the next free turn")
	e.Meta = map[string]any{"nextPlayTime": waitSeconds(left)}
	return e
}

type chancePlayRow struct {
	ID, GameID, UserID, Mode, Status string
	Roll                             int
	Slot                             sql.NullInt64
	PrizeID, GroupID, ClaimID        sql.NullString
	Price                            *coinPrice
	CreatedAt                        time.Time
}

const chancePlayColumns = `id::text, game_id::text, user_id::text, mode, status, roll, slot, prize_id::text, prize_group_id::text, claim_id::text, price::text, created_at`

func scanChancePlay(row scanner) (*chancePlayRow, error) {
	var p chancePlayRow
	var price sql.NullString
	if err := row.Scan(&p.ID, &p.GameID, &p.UserID, &p.Mode, &p.Status, &p.Roll, &p.Slot, &p.PrizeID, &p.GroupID, &p.ClaimID, &price, &p.CreatedAt); err != nil {
		return nil, err
	}
	if price.Valid && price.String != "" && price.String != "null" {
		var cp coinPrice
		if json.Unmarshal([]byte(price.String), &cp) == nil && cp.CurrencyKey != "" {
			p.Price = &cp
		}
	}
	return &p, nil
}

type chanceResult struct {
	Play    *chancePlayRow
	Game    *chanceGame
	Rewards []prizeItem
	Wait    time.Duration
}

func (s *service) chanceCurrencyKey(ctx context.Context, q queryer, id string) (string, error) {
	var key string
	err := q.QueryRowContext(ctx, `SELECT key FROM economy_currency WHERE id::text=$1`, id).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", badRequest("TICKET_PLAY_IS_NOT_SUPPORT", "", "this game's ticket currency no longer exists")
	}
	return key, err
}

func notEnough(key string, amount, balance int64) *appError {
	e := conflict("CURRENCY_IS_NOT_ENOUGH", "not enough "+key)
	e.Meta = map[string]any{"currencyKey": key, "amount": amount, "balance": balance}
	return e
}

func (s *service) playChance(ctx context.Context, userID, gameID, mode string) (*chanceResult, error) {
	g, err := loadChanceGame(ctx, s.db, gameID, "")
	if err != nil {
		return nil, err
	}
	if !g.Active {
		return nil, notFound("CHANCE_GAME_NOT_FOUND")
	}
	var price *coinPrice
	switch mode {
	case "free":
		if g.Interval == nil {
			return nil, badRequest("INTERVAL_TIME_IS_NULL", "", "this game has no free turn")
		}
	case "ticket":
		if g.Price == nil {
			return nil, badRequest("TICKET_PLAY_IS_NOT_SUPPORT", "", "this game cannot be played with a ticket")
		}
		key, err := s.chanceCurrencyKey(ctx, s.db, g.Price.CurrencyID)
		if err != nil {
			return nil, err
		}
		price = &coinPrice{CurrencyKey: key, Amount: g.Price.Amount}
		wallet, err := s.userWallet(ctx, userID)
		if err != nil {
			return nil, err
		}
		if wallet[key] < price.Amount {
			return nil, notEnough(key, price.Amount, wallet[key])
		}
	}
	now := s.now().UTC()
	play := &chancePlayRow{ID: uuid.Must(uuid.NewV4()).String(), GameID: g.ID, UserID: userID, Mode: mode, Price: price, CreatedAt: now}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_chance_state (game_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, g.ID, userID); err != nil {
			return err
		}
		var plays int64
		if err := tx.QueryRowContext(ctx, `SELECT plays FROM economy_chance_state WHERE game_id=$1 AND user_id=$2 FOR UPDATE`, g.ID, userID).Scan(&plays); err != nil {
			return err
		}
		cur, err := loadChanceGame(ctx, tx, g.ID, " FOR SHARE")
		if err != nil {
			return err
		}
		if !cur.Active {
			return notFound("CHANCE_GAME_NOT_FOUND")
		}
		g = cur
		if mode == "free" {
			if g.Interval == nil {
				return badRequest("INTERVAL_TIME_IS_NULL", "", "this game has no free turn")
			}
			last, err := lastFreePlay(ctx, tx, g.ID, userID)
			if err != nil {
				return err
			}
			if left := chanceWait(g, last, now); left > 0 {
				return waitingForTurn(left)
			}
		} else {
			if g.Price == nil {
				return badRequest("TICKET_PLAY_IS_NOT_SUPPORT", "", "this game cannot be played with a ticket")
			}
			key, err := s.chanceCurrencyKey(ctx, tx, g.Price.CurrencyID)
			if err != nil {
				return err
			}
			price = &coinPrice{CurrencyKey: key, Amount: g.Price.Amount}
			play.Price = price
		}
		pick, err := s.pickChance(ctx, tx, g, userID)
		if err != nil {
			return err
		}
		play.Roll = pick.Roll
		if pick.Slot >= 0 {
			play.Slot = sql.NullInt64{Int64: int64(pick.Slot), Valid: true}
		}
		if pick.GroupID != "" {
			play.GroupID = sql.NullString{String: pick.GroupID, Valid: true}
		}
		play.Status = "done"
		if pick.PrizeID != "" {
			play.PrizeID = sql.NullString{String: pick.PrizeID, Valid: true}
			claim, err := insertClaim(ctx, tx, userID, "chance", g.ID, play.ID, "play", pick.PrizeID, "auto", 1)
			if err != nil {
				if isForeignKeyViolation(err) {
					return notFound("USER_NOT_FOUND")
				}
				return err
			}
			if claim == nil {
				return conflict("PLAY_CONFLICT", "this play was already recorded")
			}
			play.ClaimID = sql.NullString{String: claim.ClaimID, Valid: true}
			play.Status = "pending"
		} else if mode == "ticket" {
			play.Status = "pending"
		}
		var priceJSON any
		if price != nil {
			raw, _ := json.Marshal(price)
			priceJSON = string(raw)
		}
		var finished any
		if play.Status == "done" {
			finished = now
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO economy_chance_play (id, game_id, user_id, mode, status, roll, slot, prize_id, prize_group_id, claim_id, price, created_at, finished_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12,$13)`,
			play.ID, g.ID, userID, mode, play.Status, play.Roll, play.Slot, play.PrizeID, play.GroupID, play.ClaimID, priceJSON, now, finished); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE economy_chance_state SET plays=plays+1, last_play_at=$3 WHERE game_id=$1 AND user_id=$2`, g.ID, userID, now)
		return err
	})
	if err != nil {
		return nil, err
	}
	res := &chanceResult{Play: play, Game: g, Rewards: []prizeItem{}}
	if play.Status == "pending" {
		rewards, err := s.finishChancePlay(ctx, play)
		if err != nil {
			if s.cancelChancePlay(ctx, play.ID) {
				if app, ok := asApp(err); ok && app.Code == "INSUFFICIENT_BALANCE" && price != nil {
					wallet, _ := s.userWallet(ctx, userID)
					return nil, notEnough(price.CurrencyKey, price.Amount, wallet[price.CurrencyKey])
				}
			}
			return nil, err
		}
		res.Rewards = rewards
	}
	s.recordChanceEvent(ctx, play)
	if g.Interval != nil {
		last, err := lastFreePlay(ctx, s.db, g.ID, userID)
		if err == nil {
			res.Wait = chanceWait(g, last, s.now().UTC())
		}
	}
	return res, nil
}

func (s *service) recordChanceEvent(ctx context.Context, play *chancePlayRow) {
	res, err := s.db.ExecContext(ctx, `UPDATE economy_chance_play SET event_at=now() WHERE id=$1 AND status='done' AND event_at IS NULL`, play.ID)
	if err != nil {
		s.logger.Warn("Economy chance event for %s not marked: %v", play.ID, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return
	}
	s.recordServerEvent(ctx, play.UserID, chanceEventKey, 1, map[string]any{"entityId": play.GameID, "mode": play.Mode, "playId": play.ID})
}

func (s *service) finishChancePlay(ctx context.Context, play *chancePlayRow) ([]prizeItem, error) {
	if play.ClaimID.Valid {
		if _, err := s.grantClaim(ctx, play.ClaimID.String, 1); err != nil {
			return nil, err
		}
		var status string
		if err := s.db.QueryRowContext(ctx, `SELECT status FROM economy_chance_play WHERE id=$1`, play.ID).Scan(&status); err != nil {
			return nil, err
		}
		rewards, err := s.claimRewards(ctx, play.ClaimID.String)
		if err != nil {
			return nil, err
		}
		play.Status = status
		return rewards, nil
	}
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM economy_chance_play WHERE id=$1 FOR UPDATE`, play.ID).Scan(&status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return notFound("PLAY_NOT_FOUND")
			}
			return err
		}
		if status == "done" {
			return nil
		}
		if play.Price != nil {
			if err := s.walletChange(ctx, tx, play.UserID, play.ID, map[string]int64{play.Price.CurrencyKey: -play.Price.Amount}, map[string]any{
				"reason": "chance_ticket", "sourceType": "chance", "sourceId": play.ID, "part": "debit",
			}); err != nil {
				return err
			}
		}
		_, err := tx.ExecContext(ctx, `UPDATE economy_chance_play SET status='done', error=NULL, finished_at=now() WHERE id=$1`, play.ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	play.Status = "done"
	return []prizeItem{}, nil
}

func (s *service) chanceGrantExtras(ctx context.Context, claimID string, req *grantRequest) error {
	var playID, userID string
	var price sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT id::text, user_id::text, price::text FROM economy_chance_play WHERE claim_id=$1`, claimID).Scan(&playID, &userID, &price)
	if errors.Is(err, sql.ErrNoRows) {
		return notFound("PLAY_NOT_FOUND")
	}
	if err != nil {
		return err
	}
	if price.Valid && price.String != "" && price.String != "null" {
		var cp coinPrice
		if json.Unmarshal([]byte(price.String), &cp) == nil && cp.CurrencyKey != "" && cp.Amount > 0 {
			req.Debit = map[string]int64{cp.CurrencyKey: cp.Amount}
		}
	}
	req.Reason = "chance_game"
	req.After = func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `UPDATE economy_chance_play SET status='done', error=NULL, finished_at=now() WHERE id=$1 AND status='pending'`, playID)
		return err
	}
	return nil
}

func (s *service) chanceWalletTouched(ctx context.Context, q queryer, userID string, ids ...string) (bool, error) {
	var found bool
	err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM wallet_ledger WHERE user_id=$1 AND metadata->>'sourceType'='chance' AND metadata->>'sourceId' = ANY($2))`, userID, ids).Scan(&found)
	return found, err
}

func (s *service) cancelChancePlay(ctx context.Context, playID string) bool {
	cancelled := false
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		play, err := scanChancePlay(tx.QueryRowContext(ctx, `SELECT `+chancePlayColumns+` FROM economy_chance_play WHERE id=$1 FOR UPDATE`, playID))
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if play.Status != "pending" {
			return nil
		}
		ids := []string{play.ID}
		if play.ClaimID.Valid {
			ids = append(ids, play.ClaimID.String)
			var claimed bool
			err := tx.QueryRowContext(ctx, `SELECT claimed FROM economy_user_prize WHERE source_type='chance' AND source_id=$1 FOR UPDATE`, play.ClaimID.String).Scan(&claimed)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			if claimed {
				return nil
			}
		}
		touched, err := s.chanceWalletTouched(ctx, tx, play.UserID, ids...)
		if err != nil || touched {
			return err
		}
		if play.ClaimID.Valid {
			if _, err := tx.ExecContext(ctx, `DELETE FROM economy_user_prize WHERE source_type='chance' AND source_id=$1 AND NOT claimed`, play.ClaimID.String); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM economy_reward_claim WHERE id=$1 AND status<>'granted'`, play.ClaimID.String); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM economy_chance_play WHERE id=$1`, play.ID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE economy_chance_state SET plays=GREATEST(plays-1,0) WHERE game_id=$1 AND user_id=$2`, play.GameID, play.UserID); err != nil {
			return err
		}
		cancelled = true
		return nil
	})
	if err != nil {
		s.logger.Warn("Economy chance play %s not cancelled: %v", playID, err)
		return false
	}
	return cancelled
}

func (s *service) recoverChancePlays(ctx context.Context, limit int) int {
	done := 0
	for i := 0; i < limit; i++ {
		var id string
		err := s.db.QueryRowContext(ctx, `UPDATE economy_chance_play SET attempts=attempts+1, attempted_at=now() WHERE id = (
				SELECT id FROM economy_chance_play WHERE status='pending' AND attempts < 8 AND created_at < $1
					AND (attempted_at IS NULL OR attempted_at < $1)
				ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED)
			RETURNING id::text`, s.now().UTC().Add(-chanceRecoverAfter)).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return done
		}
		if err != nil {
			s.logger.Warn("Economy chance recovery claim failed: %v", err)
			return done
		}
		play, err := scanChancePlay(s.db.QueryRowContext(ctx, `SELECT `+chancePlayColumns+` FROM economy_chance_play WHERE id=$1`, id))
		if err != nil {
			continue
		}
		_, err = s.finishChancePlay(ctx, play)
		if err != nil {
			code := err.Error()
			if app, ok := asApp(err); ok {
				code = app.Code
			}
			if code == "INSUFFICIENT_BALANCE" || code == "PRIZE_ALREADY_CLAIMED" || code == "PRIZE_ALREADY_OWNED" || code == "PRIZE_NOT_FOUND" {
				s.cancelChancePlay(ctx, play.ID)
				continue
			}
			_, _ = s.db.ExecContext(ctx, `UPDATE economy_chance_play SET error=$2 WHERE id=$1`, play.ID, truncate(code, 500))
			continue
		}
		s.recordChanceEvent(ctx, play)
		done++
	}
	return done
}

func (s *service) sweepChance(ctx context.Context) {
	for _, q := range []string{
		`DELETE FROM economy_chance_state WHERE user_id IN (SELECT DISTINCT c.user_id FROM economy_chance_state c WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = c.user_id) LIMIT 200)`,
		`DELETE FROM economy_chance_play WHERE user_id IN (SELECT DISTINCT c.user_id FROM economy_chance_play c WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = c.user_id) LIMIT 200)`,
	} {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			s.logger.Warn("Economy chance sweep failed: %v", err)
			return
		}
	}
}
