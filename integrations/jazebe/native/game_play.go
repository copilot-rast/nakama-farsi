package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gamePlayInput struct {
	PlanetID string `json:"planetId"`
	GameID   string `json:"gameId"`
	Paid     bool   `json:"paid"`
	Box      *int   `json:"box"`
	Hits     *int   `json:"hits"`
}

type playableGame struct {
	ID, PlanetID, Name, Engine, EntryModel, Status, PlanetStatus                              string
	PointCost, FreeIndividualCount, FreeGroupCount, PrizeMin, PrizeMax, RewardXp, DurationSec int
	FreeIndividualPeriod, FreeGroupPeriod, LotteryID                                          sql.NullString
	TemplateActive                                                                            bool
}

func loadPlayableGame(ctx context.Context, tx *sql.Tx, planetID, gameID string) (playableGame, error) {
	var g playableGame
	var own, group sql.NullInt64
	err := tx.QueryRowContext(ctx, `SELECT g.id,g.planet_id,g.name,COALESCE(t.engine,g.config->>'engine','box'),g.entry_model,g.status,
		g.point_cost,g.free_individual_count,g.free_group_count,g.prize_ticket_min,g.prize_ticket_max,g.reward_xp,
		g.free_individual_period,g.free_group_period,g.target_lottery_id,
		COALESCE((g.config->>'durationSec')::integer,(t.default_config->>'durationSec')::integer,10),
		COALESCE(t.is_active AND t.suspended_at IS NULL,true),p.status
		FROM jazebe_game g JOIN jazebe_planet p ON p.id=g.planet_id LEFT JOIN jazebe_game_template t ON t.id=g.template_id
		WHERE g.id=$1 AND g.planet_id=$2`, gameID, planetID).Scan(&g.ID, &g.PlanetID, &g.Name, &g.Engine, &g.EntryModel, &g.Status,
		&g.PointCost, &own, &group, &g.PrizeMin, &g.PrizeMax, &g.RewardXp, &g.FreeIndividualPeriod, &g.FreeGroupPeriod, &g.LotteryID, &g.DurationSec, &g.TemplateActive, &g.PlanetStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return g, status.Error(codes.NotFound, "game not found")
	}
	if err != nil {
		return g, err
	}
	if own.Valid {
		g.FreeIndividualCount = int(own.Int64)
	}
	if group.Valid {
		g.FreeGroupCount = int(group.Int64)
	}
	return g, nil
}

func gameConfigPlayable(g playableGame) error {
	if g.Status != "active" || !g.TemplateActive || g.PlanetStatus != "active" {
		return status.Error(codes.FailedPrecondition, "game not active")
	}
	if g.Engine != "box" && g.Engine != "tap" {
		return status.Error(codes.FailedPrecondition, "game engine unsupported")
	}
	if g.PrizeMax < g.PrizeMin || g.PrizeMin < 0 || g.DurationSec < 3 || g.DurationSec > 600 {
		return status.Error(codes.FailedPrecondition, "invalid game configuration")
	}
	if err := gameConfigValid(g); err != nil {
		return status.Error(codes.FailedPrecondition, "invalid game configuration")
	}
	return nil
}

func gameQuotaKey(period sql.NullString, now time.Time) (string, error) {
	p := "daily"
	if period.Valid {
		p = period.String
	}
	return missionPeriodKey("periodic", p, now)
}

func gameQuotaUsed(ctx context.Context, tx *sql.Tx, g playableGame, user, group string, now time.Time) (int, int, string, string, error) {
	userKey, err := gameQuotaKey(g.FreeIndividualPeriod, now)
	if err != nil {
		return 0, 0, "", "", err
	}
	groupKey, err := gameQuotaKey(g.FreeGroupPeriod, now)
	if err != nil {
		return 0, 0, "", "", err
	}
	var own, shared int
	if err = tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT used FROM jazebe_game_quota WHERE game_id=$1 AND scope='user' AND user_id=$2 AND period_key=$3),0)`, g.ID, user, userKey).Scan(&own); err != nil {
		return 0, 0, "", "", err
	}
	if group != "" {
		if err = tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT used FROM jazebe_game_quota WHERE game_id=$1 AND scope='group' AND group_id=$2 AND period_key=$3),0)`, g.ID, group, groupKey).Scan(&shared); err != nil {
			return 0, 0, "", "", err
		}
	}
	return own, shared, userKey, groupKey, nil
}

func gameBalance(ctx context.Context, tx *sql.Tx, planetID, user string) (int64, error) {
	var balance int64
	err := tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT balance FROM jazebe_point_wallet WHERE planet_id=$1 AND user_id=$2),0)`, planetID, user).Scan(&balance)
	return balance, err
}

func gameStart(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in gamePlayInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	gameID, err := requiredUUID(in.GameID)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	// Serialize starts per game/player across all Nakama nodes.
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, "jazebe_game:"+gameID+":"+user); err != nil {
		return "", err
	}
	g, err := loadPlayableGame(ctx, tx, planetID, gameID)
	if err != nil {
		return "", err
	}
	if err = gameConfigPlayable(g); err != nil {
		return "", err
	}
	var groupID string
	err = tx.QueryRowContext(ctx, `SELECT gm.group_id FROM jazebe_group_membership gm JOIN jazebe_group gr ON gr.id=gm.group_id
		WHERE gm.planet_id=$1 AND gm.user_id=$2 AND gm.status='active' AND gr.status='active'`, planetID, user).Scan(&groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.PermissionDenied, "active group membership required")
	}
	if err != nil {
		return "", err
	}
	if g.PrizeMax > 0 {
		if !g.LotteryID.Valid {
			return "", status.Error(codes.FailedPrecondition, "game lottery required")
		}
		var onSale bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_lottery WHERE id=$1 AND planet_id=$2 AND status='active' AND starts_at<=$3 AND draw_at>$3)`, g.LotteryID.String, planetID, now).Scan(&onSale); err != nil {
			return "", err
		}
		if !onSale {
			return "", status.Error(codes.FailedPrecondition, "game lottery closed")
		}
	}
	var recent int
	if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_game_session WHERE game_id=$1 AND user_id=$2 AND started_at>=$3`, gameID, user, now.Add(-time.Minute)).Scan(&recent); err != nil {
		return "", err
	}
	if recent >= gameMaxRoundsPerMinute {
		return "", status.Error(codes.ResourceExhausted, "too many game rounds")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_game_session SET status='abandoned',finished_at=$3 WHERE game_id=$1 AND user_id=$2 AND status='open'`, gameID, user, now); err != nil {
		return "", err
	}
	own, shared, userKey, groupKey, err := gameQuotaUsed(ctx, tx, g, user, groupID, now)
	if err != nil {
		return "", err
	}
	var dailyCap sql.NullInt64
	if err = tx.QueryRowContext(ctx, `SELECT free_game_daily_cap FROM jazebe_planet_settings WHERE planet_id=$1`, planetID).Scan(&dailyCap); err != nil {
		return "", err
	}
	if dailyCap.Valid {
		var used int64
		if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_game_session WHERE planet_id=$1 AND user_id=$2
			AND entry_type IN ('free_individual','free_group') AND (started_at AT TIME ZONE 'Asia/Tehran')::date=($3::timestamptz AT TIME ZONE 'Asia/Tehran')::date`, planetID, user, now).Scan(&used); err != nil {
			return "", err
		}
		if used >= dailyCap.Int64 {
			own = g.FreeIndividualCount
			shared = g.FreeGroupCount
		}
	}
	entry := ""
	cost := 0
	scope := ""
	period := ""
	freePreferred := g.EntryModel == "free" || !in.Paid
	if g.EntryModel != "cost" && freePreferred && own < g.FreeIndividualCount {
		entry = "free_individual"
		scope = "user"
		period = userKey
	}
	if entry == "" && g.EntryModel != "cost" && freePreferred && shared < g.FreeGroupCount {
		entry = "free_group"
		scope = "group"
		period = groupKey
	}
	if entry == "" && g.EntryModel == "free" {
		return "", status.Error(codes.FailedPrecondition, "free game quota exhausted")
	}
	if entry == "" {
		if g.PointCost <= 0 {
			return "", status.Error(codes.FailedPrecondition, "game point cost required")
		}
		entry = "paid"
		cost = g.PointCost
	}
	var boxes [3]int
	if g.Engine == "box" {
		boxes, err = gameRollBoxes(g.PrizeMin, g.PrizeMax)
		if err != nil {
			return "", err
		}
	}
	secret, _ := json.Marshal(map[string]any{"engine": g.Engine, "boxes": boxes, "durationSec": g.DurationSec})
	id := uuid.Must(uuid.NewV4()).String()
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_game_session(id,planet_id,game_id,user_id,group_id,entry_type,point_cost,result,started_at)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9)`, id, planetID, gameID, user, groupID, entry, cost, string(secret), now); err != nil {
		return "", err
	}
	var balance int64
	if entry == "paid" {
		movement, e := applyPointMovement(ctx, tx, planetID, user, "debit", int64(cost), "game", id, g.Name, now)
		if errors.Is(e, errInsufficientPoints) {
			return "", status.Error(codes.FailedPrecondition, "insufficient points")
		}
		if e != nil {
			return "", e
		}
		balance = movement.Balance
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_game_session SET point_transaction_id=$2 WHERE id=$1`, id, movement.TransactionID); err != nil {
			return "", err
		}
	} else {
		target := user
		if scope == "group" {
			target = groupID
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_game_quota(id,planet_id,game_id,scope,user_id,group_id,period_key,used)
			VALUES($1,$2,$3,$4,CASE WHEN $4='user' THEN $5::uuid ELSE NULL END,CASE WHEN $4='group' THEN $5::uuid ELSE NULL END,$6,0)
			ON CONFLICT DO NOTHING`, uuid.Must(uuid.NewV4()).String(), planetID, gameID, scope, target, period)
		if err != nil {
			return "", err
		}
		limit := g.FreeIndividualCount
		if scope == "group" {
			limit = g.FreeGroupCount
		}
		result, e := tx.ExecContext(ctx, `UPDATE jazebe_game_quota SET used=used+1 WHERE game_id=$1 AND scope=$2 AND period_key=$3
			AND (($2='user' AND user_id=$4) OR ($2='group' AND group_id=$4)) AND used<$5`, gameID, scope, period, target, limit)
		if e != nil {
			return "", e
		}
		affected, e := result.RowsAffected()
		if e != nil {
			return "", e
		}
		if affected != 1 {
			return "", status.Error(codes.FailedPrecondition, "free game quota exhausted")
		}
		balance, err = gameBalance(ctx, tx, planetID, user)
		if err != nil {
			return "", err
		}
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	left := g.FreeIndividualCount - own
	if entry == "free_individual" {
		left--
	}
	if left < 0 {
		left = 0
	}
	if left == 0 && g.FreeGroupCount > 0 {
		left = g.FreeGroupCount - shared
		if entry == "free_group" {
			left--
		}
		if left < 0 {
			left = 0
		}
	}
	return encode(map[string]any{"balance": balance, "freePlaysLeft": left, "minRoundMs": gameMinRoundMs})
}

type gameRound struct {
	ID, GroupID, EntryType string
	Cost                   int
	Started                time.Time
	Secret                 []byte
}

func gameOpenRound(ctx context.Context, tx *sql.Tx, planetID, gameID, user string) (gameRound, error) {
	var r gameRound
	err := tx.QueryRowContext(ctx, `SELECT id,group_id,entry_type,point_cost,started_at,result FROM jazebe_game_session
		WHERE planet_id=$1 AND game_id=$2 AND user_id=$3 AND status='open' FOR UPDATE`, planetID, gameID, user).
		Scan(&r.ID, &r.GroupID, &r.EntryType, &r.Cost, &r.Started, &r.Secret)
	if errors.Is(err, sql.ErrNoRows) {
		return r, status.Error(codes.NotFound, "open game round not found")
	}
	return r, err
}

func gameSettleEmpty(ctx context.Context, tx *sql.Tx, id string, now time.Time, reason string) error {
	_, err := tx.ExecContext(ctx, `UPDATE jazebe_game_session SET status='finished',finished_at=$2,
		result=COALESCE(result,'{}'::jsonb)||jsonb_build_object('rejected',$3::text) WHERE id=$1 AND status='open'`, id, now, reason)
	return err
}

func gameRefund(ctx context.Context, db *sql.DB, planetID, user, gameID string, now time.Time, reason string) (int64, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	r, err := gameOpenRound(ctx, tx, planetID, gameID, user)
	if err != nil {
		return 0, err
	}
	if err = gameSettleEmpty(ctx, tx, r.ID, now, reason); err != nil {
		return 0, err
	}
	var balance int64
	if r.EntryType == "paid" && r.Cost > 0 {
		movement, e := applyPointMovement(ctx, tx, planetID, user, "credit", int64(r.Cost), "game_refund", r.ID, "Game prize unavailable", now)
		if e != nil {
			return 0, e
		}
		balance = movement.Balance
	} else {
		balance, err = gameBalance(ctx, tx, planetID, user)
		if err != nil {
			return 0, err
		}
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return balance, nil
}

func gameClaim(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in gamePlayInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	planetID, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	gameID, err := requiredUUID(in.GameID)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	g, err := loadPlayableGame(ctx, tx, planetID, gameID)
	if err != nil {
		return "", err
	}
	r, err := gameOpenRound(ctx, tx, planetID, gameID, user)
	if err != nil {
		return "", err
	}
	elapsed := int(now.Sub(r.Started).Milliseconds())
	if elapsed < gameMinRoundMs {
		return "", status.Error(codes.InvalidArgument, "game round too fast")
	}
	if elapsed > gameMaxRoundMs {
		if err = gameSettleEmpty(ctx, tx, r.ID, now, "expired"); err != nil {
			return "", err
		}
		if err = tx.Commit(); err != nil {
			return "", err
		}
		return "", status.Error(codes.FailedPrecondition, "game round expired")
	}
	var secret struct {
		Engine      string `json:"engine"`
		Boxes       [3]int `json:"boxes"`
		DurationSec int    `json:"durationSec"`
	}
	if err = json.Unmarshal(r.Secret, &secret); err != nil {
		return "", err
	}
	box, hits := 0, 0
	switch secret.Engine {
	case "box":
		if in.Box == nil {
			return "", status.Error(codes.InvalidArgument, "box required")
		}
		box = *in.Box
	case "tap":
		if in.Hits == nil {
			return "", status.Error(codes.InvalidArgument, "hits required")
		}
		hits = *in.Hits
	default:
		return "", status.Error(codes.FailedPrecondition, "unsupported stored round engine")
	}
	score, tickets, outcomeErr := gameRoundOutcome(secret.Engine, secret.Boxes, box, hits, secret.DurationSec, elapsed, g.PrizeMin, g.PrizeMax)
	if outcomeErr != nil {
		if err = gameSettleEmpty(ctx, tx, r.ID, now, "impossible"); err != nil {
			return "", err
		}
		if err = tx.Commit(); err != nil {
			return "", err
		}
		return "", status.Error(codes.InvalidArgument, "impossible game result")
	}
	resultJSON, _ := json.Marshal(map[string]any{"engine": secret.Engine, "boxes": secret.Boxes, "durationSec": secret.DurationSec, "box": in.Box, "hits": in.Hits, "score": score, "tickets": tickets})
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_game_session SET status='finished',finished_at=$2,score=$3,
		result=$4::jsonb,awarded_tickets=$5,awarded_xp=$6 WHERE id=$1`, r.ID, now, score, string(resultJSON), tickets, g.RewardXp); err != nil {
		return "", err
	}
	if tickets > 0 {
		if !g.LotteryID.Valid {
			_ = tx.Rollback()
			balance, refundErr := gameRefund(ctx, db, planetID, user, gameID, now, "lottery_missing")
			if refundErr != nil {
				return "", refundErr
			}
			return encode(map[string]any{"tickets": 0, "balance": balance, "refunded": true, "boxes": secret.Boxes})
		}
		_, issueErr := issueLotteryTickets(ctx, tx, ticketIssue{LotteryID: g.LotteryID.String, PlanetID: planetID, GroupID: r.GroupID, UserID: user,
			Source: "game", SourceID: r.ID, CreatorType: "member", Count: tickets}, now)
		if issueErr != nil {
			// Roll back the attempted settlement before the independent refund transaction.
			_ = tx.Rollback()
			code := status.Code(issueErr)
			if code != codes.FailedPrecondition && code != codes.ResourceExhausted {
				return "", issueErr
			}
			balance, refundErr := gameRefund(ctx, db, planetID, user, gameID, now, code.String())
			if refundErr != nil {
				return "", refundErr
			}
			return encode(map[string]any{"tickets": 0, "balance": balance, "refunded": true, "boxes": secret.Boxes})
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,actor_user_id,message_key,payload)
			VALUES($1,$2,$3,'game',$4,'game.played',jsonb_build_object('gameId',$5::uuid,'sessionId',$6::uuid,'tickets',$7))`,
			uuid.Must(uuid.NewV4()).String(), planetID, r.GroupID, user, gameID, r.ID, tickets); err != nil {
			return "", err
		}
	}
	if err = gameAwardXP(ctx, tx, planetID, user, r.ID, g.RewardXp, now); err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_planet_stats(planet_id,user_id) VALUES($1,$2) ON CONFLICT(planet_id,user_id) DO NOTHING`, planetID, user); err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_user_planet_stats SET games_played=games_played+1,tickets_created=tickets_created+$3,last_activity_at=$4
		WHERE planet_id=$1 AND user_id=$2`, planetID, user, tickets, now); err != nil {
		return "", err
	}
	var inserted bool
	err = tx.QueryRowContext(ctx, `INSERT INTO jazebe_planet_membership(planet_id,user_id,join_path,last_activity_at)
		VALUES($1,$2,'first_activity',$3) ON CONFLICT(planet_id,user_id) DO NOTHING RETURNING true`, planetID, user, now).Scan(&inserted)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if inserted {
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_planet SET fans_count=fans_count+1 WHERE id=$1`, planetID); err != nil {
			return "", err
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET last_activity_at=$3 WHERE planet_id=$1 AND user_id=$2`, planetID, user, now); err != nil {
		return "", err
	}
	if err = enqueueActivityXP(ctx, tx, planetID, user, r.ID, "game"); err != nil {
		return "", err
	}
	balance, err := gameBalance(ctx, tx, planetID, user)
	if err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	gameMedalProgression(ctx, logger, db, planetID, user, r.ID)
	return encode(map[string]any{"tickets": tickets, "balance": balance, "boxes": secret.Boxes})
}

func gameAwardXP(ctx context.Context, tx *sql.Tx, planetID, user, roundID string, amount int, now time.Time) error {
	if amount <= 0 {
		return nil
	}
	var applied bool
	err := tx.QueryRowContext(ctx, `INSERT INTO jazebe_xp_record(id,planet_id,user_id,amount,source,source_id)
		VALUES($1,$2,$3,$4,'game',$5) ON CONFLICT(planet_id,user_id,source,source_id) DO NOTHING RETURNING true`,
		uuid.Must(uuid.NewV4()).String(), planetID, user, amount, roundID).Scan(&applied)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_planet_stats(planet_id,user_id) VALUES($1,$2) ON CONFLICT(planet_id,user_id) DO NOTHING`, planetID, user); err != nil {
		return err
	}
	var total int64
	var current int
	if err = tx.QueryRowContext(ctx, `UPDATE jazebe_user_planet_stats SET total_xp=total_xp+$3,last_activity_at=$4
		WHERE planet_id=$1 AND user_id=$2 RETURNING total_xp,level`, planetID, user, amount, now).Scan(&total, &current); err != nil {
		return err
	}
	var thresholds []byte
	if err = tx.QueryRowContext(ctx, `SELECT xp_level_thresholds FROM jazebe_planet_settings WHERE planet_id=$1`, planetID).Scan(&thresholds); err != nil {
		return err
	}
	level, key := levelAt(total, thresholds, current)
	if level > current {
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_user_planet_stats SET level=$3,level_key=$4 WHERE planet_id=$1 AND user_id=$2`, planetID, user, level, key)
	}
	return err
}
