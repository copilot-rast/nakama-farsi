package main

import (
	"context"
	"database/sql"
	"math/big"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Called inside the transaction that changes a user's verification status to
// verified. Row locks, the point-ledger source key and share state make a
// retry safe. The KYC provider and verification workflow are separate work.
func releaseVerifiedLotteryShares(ctx context.Context, tx *sql.Tx, userID string, now time.Time) (int, error) {
	var verified bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_user_verification WHERE user_id=$1 AND status='verified')`, userID).Scan(&verified); err != nil {
		return 0, status.Error(codes.Internal, "verification lookup failed")
	}
	if !verified {
		return 0, status.Error(codes.FailedPrecondition, "user not KYC verified")
	}
	rows, err := tx.QueryContext(ctx, `SELECT s.id,s.winner_id,s.planet_id,s.share_value_rial::text,s.payout_status,w.lottery_id,p.title,p.type,l.title FROM jazebe_lottery_winner_share s JOIN jazebe_lottery_winner w ON w.id=s.winner_id JOIN jazebe_lottery_prize p ON p.id=w.prize_id JOIN jazebe_lottery l ON l.id=w.lottery_id WHERE s.user_id=$1 AND s.blocked_by_kyc FOR UPDATE OF s`, userID)
	if err != nil {
		return 0, status.Error(codes.Internal, "blocked shares unavailable")
	}
	type blocked struct{ ID, WinnerID, PlanetID, Value, Payout, LotteryID, PrizeTitle, PrizeKind, LotteryTitle string }
	list := []blocked{}
	for rows.Next() {
		var s blocked
		if err = rows.Scan(&s.ID, &s.WinnerID, &s.PlanetID, &s.Value, &s.Payout, &s.LotteryID, &s.PrizeTitle, &s.PrizeKind, &s.LotteryTitle); err != nil {
			rows.Close()
			return 0, status.Error(codes.Internal, "blocked shares unavailable")
		}
		list = append(list, s)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, status.Error(codes.Internal, "blocked shares unavailable")
	}
	rate, err := lotteryPrizeRate()
	if err != nil {
		return 0, err
	}
	for _, s := range list {
		payout := s.Payout
		if payout == "pending_verification" {
			payout = "payable"
		}
		if s.PrizeKind == "planet_points" && payout == "payable" {
			value, ok := new(big.Int).SetString(s.Value, 10)
			if !ok || value.Sign() < 0 {
				return 0, status.Error(codes.Internal, "share value invalid")
			}
			points := new(big.Int).Quo(value, big.NewInt(rate))
			if !points.IsInt64() {
				return 0, status.Error(codes.Internal, "point prize overflow")
			}
			if points.Sign() > 0 {
				if _, err = applyPointMovement(ctx, tx, s.PlanetID, userID, "credit", points.Int64(), "lottery_prize", s.ID, "Lottery prize: "+s.LotteryTitle, now); err != nil {
					return 0, status.Error(codes.Internal, "KYC prize payout failed")
				}
			}
			payout = "paid"
			if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,target_type,target_id,planet_id,params) VALUES('lottery.payout','system','lottery',$1,$2,jsonb_build_object('winnerId',$3::uuid,'shareId',$4::uuid,'userId',$5::uuid,'valueRial',$6::text,'points',$7::text,'rialPerPoint',$8))`, s.LotteryID, s.PlanetID, s.WinnerID, s.ID, userID, s.Value, points.String(), rate); err != nil {
				return 0, status.Error(codes.Internal, "KYC payout audit failed")
			}
		}
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery_winner_share SET blocked_by_kyc=false,payout_status=$2 WHERE id=$1`, s.ID, payout); err != nil {
			return 0, status.Error(codes.Internal, "share release failed")
		}
		if s.PrizeKind == "planet_points" {
			if _, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery_winner w SET delivery_status='delivered',delivery_note=COALESCE(delivery_note,'پرداخت خودکار به کیف پول') WHERE w.id=$1 AND w.delivery_status<>'delivered' AND NOT EXISTS(SELECT 1 FROM jazebe_lottery_winner_share s WHERE s.winner_id=w.id AND s.payout_status IN ('payable','pending_verification'))`, s.WinnerID); err != nil {
				return 0, status.Error(codes.Internal, "winner delivery failed")
			}
		}
		if err = persistJazebeNotification(ctx, tx, userID, "lottery-kyc-release:"+s.ID, "سهم جایزه‌ات آزاد شد", map[string]any{"category": "critical", "priority": "urgent", "actionLink": "/raffle?id=" + s.LotteryID, "planetId": s.PlanetID, "lotteryId": s.LotteryID, "winnerId": s.WinnerID, "shareId": s.ID, "prizeTitle": s.PrizeTitle, "shareRial": s.Value}, now); err != nil {
			return 0, status.Error(codes.Internal, "share release notice failed")
		}
	}
	return len(list), nil
}
