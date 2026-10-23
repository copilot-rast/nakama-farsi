package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const lotteryDrawAlgorithm = "hmac-sha256/rejection/serial-asc/v1"

type lotteryDrawInput struct {
	PlanetID  string `json:"planetId"`
	LotteryID string `json:"lotteryId"`
}
type lotteryDrawPrize struct {
	ID, Title, Kind, Value string
	Priority               int
}

// The privileged manual path uses the same flip-and-resume discipline as the
// source worker. The seed is committed before the winner transaction; a retry
// resumes from that committed seed and snapshot time.
func lotteryBrandDraw(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in lotteryDrawInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.LotteryID, err = requiredUUID(in.LotteryID)
	if err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	var allowed bool
	err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet p JOIN jazebe_brand_admin a ON a.brand_id=p.brand_id WHERE p.id=$1 AND a.user_id=$2 AND a.is_active AND a.role IN ('main','marketing'))`, in.PlanetID, actor).Scan(&allowed)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery authorization failed")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "brand economy role required")
	}
	now := time.Now().UTC()
	seed, err := newLotteryDrawSeed()
	if err != nil {
		return "", status.Error(codes.Internal, "draw seed unavailable")
	}
	meta, _ := json.Marshal(map[string]any{"algorithm": lotteryDrawAlgorithm, "snapshotAt": now.Format(time.RFC3339Nano)})
	res, err := db.ExecContext(ctx, `UPDATE jazebe_lottery SET status='drawing',draw_seed=COALESCE(draw_seed,$2),draw_meta=$3::jsonb,updated_at=$4 WHERE id=$1 AND status='active' AND draw_at<=$4 AND planet_id=$5`, in.LotteryID, seed, string(meta), now, in.PlanetID)
	if err != nil {
		return "", status.Error(codes.Internal, "draw flip failed")
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return "", status.Error(codes.Internal, "draw flip failed")
	}
	if affected == 0 {
		var state string
		err = db.QueryRowContext(ctx, `SELECT status FROM jazebe_lottery WHERE id=$1 AND planet_id=$2`, in.LotteryID, in.PlanetID).Scan(&state)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.NotFound, "lottery not found")
		}
		if err != nil {
			return "", status.Error(codes.Internal, "lottery lookup failed")
		}
		if state != "drawing" {
			return "", status.Error(codes.FailedPrecondition, "lottery not drawable")
		}
	}
	return runLotteryDraw(ctx, db, in.LotteryID, actor, "brand_admin", "manual", now)
}

func newLotteryDrawSeed() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func runLotteryDraw(ctx context.Context, db *sql.DB, lotteryID, actor, actorType, trigger string, now time.Time) (string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "draw unavailable")
	}
	defer tx.Rollback()
	var planetID, title, state, seed string
	var drawAt time.Time
	var rawMeta []byte
	err = tx.QueryRowContext(ctx, `SELECT planet_id,title,status,COALESCE(draw_seed,''),draw_meta,draw_at FROM jazebe_lottery WHERE id=$1 FOR UPDATE`, lotteryID).Scan(&planetID, &title, &state, &seed, &rawMeta, &drawAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "lottery not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "lottery lookup failed")
	}
	if state != "drawing" || seed == "" {
		return "", status.Error(codes.FailedPrecondition, "lottery not drawable")
	}
	var previous struct {
		SnapshotAt string `json:"snapshotAt"`
		Trigger    string `json:"trigger"`
	}
	if json.Unmarshal(rawMeta, &previous) != nil {
		return "", status.Error(codes.Internal, "draw metadata invalid")
	}
	snapshot, err := time.Parse(time.RFC3339Nano, previous.SnapshotAt)
	if err != nil {
		snapshot = drawAt
	}
	if previous.Trigger != "" {
		trigger = previous.Trigger
	}
	prizes, err := loadLotteryDrawPrizes(ctx, tx, lotteryID)
	if err != nil {
		return "", err
	}
	tickets, err := loadLotteryDrawTickets(ctx, tx, lotteryID, snapshot)
	if err != nil {
		return "", err
	}
	purePrizes := make([]drawPrize, len(prizes))
	for i, p := range prizes {
		purePrizes[i] = drawPrize{p.ID, p.Priority}
	}
	outcome, err := drawLottery(seed, tickets, purePrizes)
	if err != nil {
		return "", status.Error(codes.Internal, "draw computation failed")
	}
	byPrize := map[string]lotteryDrawPrize{}
	for _, p := range prizes {
		byPrize[p.ID] = p
	}
	rate, err := lotteryPrizeRate()
	if err != nil {
		return "", err
	}
	for _, pick := range outcome.Picks {
		p := byPrize[pick.PrizeID]
		winnerID := uuid.Must(uuid.NewV4()).String()
		_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_lottery_winner(id,planet_id,lottery_id,prize_id,ticket_id,group_id,recorded_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, winnerID, planetID, lotteryID, p.ID, pick.TicketID, pick.GroupID, now)
		if err != nil {
			return "", status.Error(codes.Internal, "winner persistence failed")
		}
		members, err := lotteryShareMembers(ctx, tx, lotteryID, pick.GroupID, snapshot)
		if err != nil {
			return "", err
		}
		value, ok := new(big.Int).SetString(p.Value, 10)
		if !ok {
			return "", status.Error(codes.Internal, "invalid prize value")
		}
		shares, err := splitLotteryPrize(value, members)
		if err != nil {
			return "", status.Error(codes.Internal, "share computation failed")
		}
		allPaid := p.Kind == "planet_points"
		for _, s := range shares {
			shareID := uuid.Must(uuid.NewV4()).String()
			verified, err := lotteryUserVerified(ctx, tx, s.UserID)
			if err != nil {
				return "", err
			}
			payout := "pending_verification"
			if verified {
				payout = "payable"
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_lottery_winner_share(id,planet_id,winner_id,user_id,individual_ticket_count,share_percent,share_value_rial,payout_status,blocked_by_kyc) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, shareID, planetID, winnerID, s.UserID, s.IndividualTicketCount, s.SharePercent, s.ShareValueRial, payout, !verified)
			if err != nil {
				return "", status.Error(codes.Internal, "winner share persistence failed")
			}
			if err = persistJazebeNotification(ctx, tx, s.UserID, "lottery-win:"+winnerID, "گروه تو برنده قرعه‌کشی شد", map[string]any{"category": "critical", "priority": "urgent", "actionLink": "/raffle?id=" + lotteryID, "planetId": planetID, "lotteryId": lotteryID, "winnerId": winnerID, "prize": p.Title, "shareRial": s.ShareValueRial, "blockedByKyc": !verified}, now); err != nil {
				return "", status.Error(codes.Internal, "winner notice failed")
			}
			if p.Kind == "planet_points" {
				if !verified {
					allPaid = false
					continue
				}
				shareRial, valid := new(big.Int).SetString(s.ShareValueRial, 10)
				if !valid || shareRial.Sign() < 0 {
					return "", status.Error(codes.Internal, "invalid share value")
				}
				points := new(big.Int).Quo(shareRial, big.NewInt(rate))
				if !points.IsInt64() {
					return "", status.Error(codes.Internal, "point prize overflow")
				}
				if points.Sign() > 0 {
					if _, err = applyPointMovement(ctx, tx, planetID, s.UserID, "credit", points.Int64(), "lottery_prize", shareID, "Lottery prize: "+title, now); err != nil {
						return "", status.Error(codes.Internal, "prize payout failed")
					}
				}
				if _, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery_winner_share SET payout_status='paid' WHERE id=$1`, shareID); err != nil {
					return "", status.Error(codes.Internal, "prize payout status failed")
				}
				if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,target_type,target_id,planet_id,params) VALUES('lottery.payout','system','lottery',$1,$2,jsonb_build_object('winnerId',$3::uuid,'shareId',$4::uuid,'userId',$5::uuid,'valueRial',$6::text,'points',$7::text,'rialPerPoint',$8::bigint))`, lotteryID, planetID, winnerID, shareID, s.UserID, s.ShareValueRial, points.String(), rate); err != nil {
					return "", status.Error(codes.Internal, "prize payout audit failed")
				}
			}
		}
		if allPaid {
			if _, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery_winner SET delivery_status='delivered' WHERE id=$1`, winnerID); err != nil {
				return "", status.Error(codes.Internal, "winner delivery failed")
			}
		}
	}
	finalMeta, _ := json.Marshal(map[string]any{"algorithm": lotteryDrawAlgorithm, "snapshotAt": snapshot.Format(time.RFC3339Nano), "trigger": trigger, "ticketCount": outcome.TicketCount, "ticketDigest": outcome.TicketDigest, "picks": outcome.Picks, "unawardedPrizeIds": outcome.UnawardedPrizeIDs})
	_, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery SET status='finished',drawn_at=$2,draw_meta=$3::jsonb,updated_at=$2 WHERE id=$1`, lotteryID, now, string(finalMeta))
	if err != nil {
		return "", status.Error(codes.Internal, "draw finish failed")
	}
	_, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery_group_stat s SET rank=r.rank,updated_at=$2 FROM (SELECT group_id,RANK() OVER (ORDER BY ticket_count DESC) AS rank FROM jazebe_lottery_group_stat WHERE lottery_id=$1) r WHERE s.lottery_id=$1 AND s.group_id=r.group_id`, lotteryID, now)
	if err != nil {
		return "", status.Error(codes.Internal, "draw rank failed")
	}
	params, _ := json.Marshal(map[string]any{"seed": seed, "algorithm": lotteryDrawAlgorithm, "trigger": trigger, "snapshotAt": snapshot.Format(time.RFC3339Nano), "ticketCount": outcome.TicketCount, "ticketDigest": outcome.TicketDigest, "picks": outcome.Picks, "unawardedPrizeIds": outcome.UnawardedPrizeIDs})
	var actorID any
	if actor != "" {
		actorID = actor
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params) VALUES('lottery.draw',$5,$1,'lottery',$2,$3,$4::jsonb)`, actorID, lotteryID, planetID, string(params), actorType)
	if err != nil {
		return "", status.Error(codes.Internal, "draw audit failed")
	}
	if err = announceLotteryDraw(ctx, tx, planetID, lotteryID, title, outcome.Picks, now); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "draw commit failed")
	}
	return encode(map[string]any{"lotteryId": lotteryID, "drawnAt": now.Format(time.RFC3339Nano), "seed": seed, "outcome": outcome})
}

func announceLotteryDraw(ctx context.Context, tx *sql.Tx, planetID, lotteryID, title string, picks []drawPick, now time.Time) error {
	groups := make([]string, 0, len(picks))
	seen := map[string]bool{}
	for _, pick := range picks {
		if !seen[pick.GroupID] {
			seen[pick.GroupID] = true
			var name string
			if err := tx.QueryRowContext(ctx, `SELECT name FROM jazebe_group WHERE id=$1`, pick.GroupID).Scan(&name); err != nil {
				return status.Error(codes.Internal, "winning group lookup failed")
			}
			groups = append(groups, name)
		}
		payload, _ := json.Marshal(map[string]any{"lotteryId": lotteryID, "serial": pick.Serial, "prizeId": pick.PrizeID})
		if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_group_activity(id,planet_id,group_id,kind,message_key,payload,created_at) VALUES($1,$2,$3,'prize','lottery.won',$4::jsonb,$5)`, uuid.Must(uuid.NewV4()).String(), planetID, pick.GroupID, string(payload), now); err != nil {
			return status.Error(codes.Internal, "prize activity failed")
		}
	}
	line := strings.Join(groups, "، ")
	if line == "" {
		line = "بدون برنده"
	}
	rows, err := tx.QueryContext(ctx, `SELECT user_id FROM jazebe_planet_membership WHERE planet_id=$1 AND is_active`, planetID)
	if err != nil {
		return status.Error(codes.Internal, "raffle audience unavailable")
	}
	users := []string{}
	for rows.Next() {
		var user string
		if err = rows.Scan(&user); err != nil {
			rows.Close()
			return status.Error(codes.Internal, "raffle audience unavailable")
		}
		users = append(users, user)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return status.Error(codes.Internal, "raffle audience unavailable")
	}
	for _, user := range users {
		if err = persistJazebeNotification(ctx, tx, user, "lottery-result:"+lotteryID, "نتیجه قرعه‌کشی اعلام شد", map[string]any{"category": "system", "priority": "normal", "actionLink": "/raffles", "planetId": planetID, "lotteryId": lotteryID, "body": "قرعه‌کشی «" + title + "» اجرا شد. گروه‌های برنده: " + line}, now); err != nil {
			return status.Error(codes.Internal, "raffle result notice failed")
		}
	}
	return nil
}

func loadLotteryDrawPrizes(ctx context.Context, tx *sql.Tx, id string) ([]lotteryDrawPrize, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,title,type,value_rial::text,priority FROM jazebe_lottery_prize WHERE lottery_id=$1 AND deleted_at IS NULL ORDER BY priority`, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "draw prizes unavailable")
	}
	defer rows.Close()
	list := []lotteryDrawPrize{}
	for rows.Next() {
		var p lotteryDrawPrize
		if err = rows.Scan(&p.ID, &p.Title, &p.Kind, &p.Value, &p.Priority); err != nil {
			return nil, status.Error(codes.Internal, "draw prizes unavailable")
		}
		list = append(list, p)
	}
	if rows.Err() != nil {
		return nil, status.Error(codes.Internal, "draw prizes unavailable")
	}
	return list, nil
}
func loadLotteryDrawTickets(ctx context.Context, tx *sql.Tx, id string, at time.Time) ([]drawTicket, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,serial,group_id FROM jazebe_lottery_ticket WHERE lottery_id=$1 AND status='active' AND created_at<=$2`, id, at)
	if err != nil {
		return nil, status.Error(codes.Internal, "draw tickets unavailable")
	}
	defer rows.Close()
	list := []drawTicket{}
	for rows.Next() {
		var t drawTicket
		if err = rows.Scan(&t.ID, &t.Serial, &t.GroupID); err != nil {
			return nil, status.Error(codes.Internal, "draw tickets unavailable")
		}
		list = append(list, t)
	}
	if rows.Err() != nil {
		return nil, status.Error(codes.Internal, "draw tickets unavailable")
	}
	return list, nil
}
func lotteryShareMembers(ctx context.Context, tx *sql.Tx, lotteryID, groupID string, at time.Time) ([]shareInput, error) {
	rows, err := tx.QueryContext(ctx, `WITH members AS (SELECT user_id FROM jazebe_group_membership WHERE group_id=$2 AND joined_at<=$3 AND (status='active' OR left_at>$3) UNION SELECT creator_user_id FROM jazebe_lottery_ticket WHERE lottery_id=$1 AND group_id=$2 AND creator_type='member' AND status='active' AND created_at<=$3) SELECT m.user_id,COUNT(t.id)::integer FROM members m LEFT JOIN jazebe_lottery_ticket t ON t.lottery_id=$1 AND t.group_id=$2 AND t.creator_user_id=m.user_id AND t.creator_type='member' AND t.status='active' AND t.created_at<=$3 GROUP BY m.user_id`, lotteryID, groupID, at)
	if err != nil {
		return nil, status.Error(codes.Internal, "share snapshot unavailable")
	}
	defer rows.Close()
	list := []shareInput{}
	for rows.Next() {
		var s shareInput
		if err = rows.Scan(&s.UserID, &s.MemberTickets); err != nil {
			return nil, status.Error(codes.Internal, "share snapshot unavailable")
		}
		list = append(list, s)
	}
	if rows.Err() != nil {
		return nil, status.Error(codes.Internal, "share snapshot unavailable")
	}
	return list, nil
}
func lotteryUserVerified(ctx context.Context, tx *sql.Tx, userID string) (bool, error) {
	var ok bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_user_verification WHERE user_id=$1 AND status='verified')`, userID).Scan(&ok)
	if err != nil {
		return false, status.Error(codes.Internal, "verification lookup failed")
	}
	return ok, nil
}
func lotteryPrizeRate() (int64, error) {
	s := os.Getenv("LOTTERY_PRIZE_RIAL_PER_POINT")
	if s == "" {
		return 1, nil
	}
	n, e := strconv.ParseInt(s, 10, 64)
	if e != nil || n < 1 {
		return 0, status.Error(codes.FailedPrecondition, "invalid LOTTERY_PRIZE_RIAL_PER_POINT")
	}
	return n, nil
}
