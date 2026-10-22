package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type lotteryDeliveryInput struct {
	PlanetID  string          `json:"planetId"`
	LotteryID string          `json:"lotteryId"`
	WinnerID  string          `json:"winnerId"`
	Status    string          `json:"status"`
	Note      json.RawMessage `json:"note"`
}

func lotteryBrandWinners(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in lotteryRef
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	in.LotteryID, err = requiredUUID(in.LotteryID)
	if err != nil {
		return "", err
	}
	if err = lotteryBrandAccess(ctx, db, in.PlanetID, actor); err != nil {
		return "", err
	}
	list, err := lotteryWinnerRows(ctx, db, in.PlanetID, in.LotteryID)
	if err != nil {
		return "", err
	}
	return encode(list)
}

func lotteryWinnerRows(ctx context.Context, db *sql.DB, planetID, lotteryID string) ([]map[string]any, error) {
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_lottery WHERE id=$1 AND planet_id=$2)`, lotteryID, planetID).Scan(&exists); err != nil {
		return nil, status.Error(codes.Internal, "lottery lookup failed")
	}
	if !exists {
		return nil, status.Error(codes.NotFound, "lottery not found")
	}
	rows, err := db.QueryContext(ctx, `SELECT w.id,p.id,p.priority,p.title,p.type,p.value_rial::text,g.id,g.name,t.serial,w.delivery_status,w.delivery_note,w.recorded_at FROM jazebe_lottery_winner w JOIN jazebe_lottery_prize p ON p.id=w.prize_id JOIN jazebe_group g ON g.id=w.group_id JOIN jazebe_lottery_ticket t ON t.id=w.ticket_id WHERE w.lottery_id=$1 AND w.planet_id=$2 ORDER BY p.priority`, lotteryID, planetID)
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery winners unavailable")
	}
	type winner struct {
		ID, PrizeID, PrizeTitle, PrizeType, Value, GroupID, GroupName, Serial, State string
		Priority                                                                     int
		Note                                                                         sql.NullString
		Recorded                                                                     time.Time
	}
	winners := []winner{}
	for rows.Next() {
		var w winner
		if err = rows.Scan(&w.ID, &w.PrizeID, &w.Priority, &w.PrizeTitle, &w.PrizeType, &w.Value, &w.GroupID, &w.GroupName, &w.Serial, &w.State, &w.Note, &w.Recorded); err != nil {
			rows.Close()
			return nil, status.Error(codes.Internal, "lottery winners unavailable")
		}
		winners = append(winners, w)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, status.Error(codes.Internal, "lottery winners unavailable")
	}
	result := make([]map[string]any, 0, len(winners))
	for _, w := range winners {
		shareRows, err := db.QueryContext(ctx, `SELECT s.user_id,COALESCE(NULLIF(trim(u.display_name),''),'کاربر'),s.individual_ticket_count,s.share_percent::text,s.share_value_rial::text,s.payout_status,s.blocked_by_kyc FROM jazebe_lottery_winner_share s JOIN users u ON u.id=s.user_id WHERE s.winner_id=$1 ORDER BY s.individual_ticket_count DESC,s.user_id`, w.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, "winner shares unavailable")
		}
		shares := []map[string]any{}
		for shareRows.Next() {
			var id, name, percent, value, payout string
			var count int
			var blocked bool
			if err = shareRows.Scan(&id, &name, &count, &percent, &value, &payout, &blocked); err != nil {
				shareRows.Close()
				return nil, status.Error(codes.Internal, "winner shares unavailable")
			}
			pct, err := strconv.ParseFloat(percent, 64)
			if err != nil {
				shareRows.Close()
				return nil, status.Error(codes.Internal, "winner share percentage invalid")
			}
			shares = append(shares, map[string]any{"userId": id, "name": name, "tickets": count, "percent": pct, "valueRial": value, "payoutStatus": payout, "blockedByKyc": blocked})
		}
		err = shareRows.Err()
		shareRows.Close()
		if err != nil {
			return nil, status.Error(codes.Internal, "winner shares unavailable")
		}
		result = append(result, map[string]any{"id": w.ID, "prize": map[string]any{"id": w.PrizeID, "priority": w.Priority, "title": w.PrizeTitle, "type": w.PrizeType, "valueRial": w.Value}, "group": map[string]any{"id": w.GroupID, "name": w.GroupName}, "ticketSerial": w.Serial, "deliveryStatus": w.State, "deliveryNote": nullString(w.Note), "recordedAt": w.Recorded.Format(time.RFC3339Nano), "shares": shares})
	}
	return result, nil
}

func lotteryDeliveryTransition(from, to string) bool {
	return from == "pending" && to == "contacted" || from == "contacted" && (to == "delivered" || to == "rejected")
}

func lotteryBrandDelivery(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in lotteryDeliveryInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	in.LotteryID, err = requiredUUID(in.LotteryID)
	if err != nil {
		return "", err
	}
	in.WinnerID, err = requiredUUID(in.WinnerID)
	if err != nil {
		return "", err
	}
	if in.Status != "contacted" && in.Status != "delivered" && in.Status != "rejected" {
		return "", status.Error(codes.InvalidArgument, "invalid delivery status")
	}
	var note any
	if in.Note != nil {
		if string(in.Note) != "null" {
			var s string
			if json.Unmarshal(in.Note, &s) != nil {
				return "", status.Error(codes.InvalidArgument, "invalid delivery note")
			}
			s = strings.TrimSpace(s)
			if !utf8.ValidString(s) || utf8.RuneCountInString(s) > 512 {
				return "", status.Error(codes.InvalidArgument, "delivery note too long")
			}
			if s != "" {
				note = s
			}
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "delivery unavailable")
	}
	defer tx.Rollback()
	var from, prizeTitle, lotteryTitle string
	var allowed bool
	err = tx.QueryRowContext(ctx, `SELECT w.delivery_status,p.title,l.title,EXISTS(SELECT 1 FROM jazebe_planet pl JOIN jazebe_brand_admin a ON a.brand_id=pl.brand_id WHERE pl.id=l.planet_id AND a.user_id=$4 AND a.is_active AND a.role IN ('main','marketing')) FROM jazebe_lottery_winner w JOIN jazebe_lottery l ON l.id=w.lottery_id JOIN jazebe_lottery_prize p ON p.id=w.prize_id WHERE w.id=$1 AND w.lottery_id=$2 AND w.planet_id=$3 FOR UPDATE OF w`, in.WinnerID, in.LotteryID, in.PlanetID, actor).Scan(&from, &prizeTitle, &lotteryTitle, &allowed)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "winner not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "winner lookup failed")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "brand economy role required")
	}
	if !lotteryDeliveryTransition(from, in.Status) {
		return "", status.Error(codes.FailedPrecondition, "invalid delivery transition")
	}
	if in.Status == "delivered" {
		var pending int
		if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_lottery_winner_share WHERE winner_id=$1 AND payout_status='pending_verification'`, in.WinnerID).Scan(&pending); err != nil {
			return "", status.Error(codes.Internal, "winner KYC lookup failed")
		}
		if pending > 0 {
			return "", status.Error(codes.FailedPrecondition, "prize shares pending verification")
		}
	}
	if in.Note == nil {
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery_winner SET delivery_status=$2 WHERE id=$1`, in.WinnerID, in.Status)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery_winner SET delivery_status=$2,delivery_note=$3 WHERE id=$1`, in.WinnerID, in.Status, note)
	}
	if err != nil {
		return "", status.Error(codes.Internal, "delivery update failed")
	}
	if in.Status == "delivered" {
		if _, err = tx.ExecContext(ctx, `UPDATE jazebe_lottery_winner_share SET payout_status='paid' WHERE winner_id=$1 AND payout_status='payable'`, in.WinnerID); err != nil {
			return "", status.Error(codes.Internal, "share settlement failed")
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT user_id FROM jazebe_lottery_winner_share WHERE winner_id=$1`, in.WinnerID)
	if err != nil {
		return "", status.Error(codes.Internal, "winner holders unavailable")
	}
	users := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "winner holders unavailable")
		}
		users = append(users, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "winner holders unavailable")
	}
	now := time.Now().UTC()
	label := map[string]string{"contacted": "تماس انجام شده", "delivered": "تحویل شده", "rejected": "رد شده"}[in.Status]
	for _, id := range users {
		if err = persistJazebeNotification(ctx, tx, id, "lottery-delivery:"+in.WinnerID+":"+in.Status, "وضعیت تحویل جایزه", map[string]any{"category": "system", "actionLink": "/raffle?id=" + in.LotteryID, "body": "تحویل جایزه «" + prizeTitle + "» از قرعه‌کشی «" + lotteryTitle + "»: " + label}, now); err != nil {
			return "", status.Error(codes.Internal, "delivery notice failed")
		}
	}
	params, _ := json.Marshal(map[string]any{"winnerId": in.WinnerID, "from": from, "to": in.Status, "note": note})
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params) VALUES('lottery.delivery','brand_admin',$1,'lottery',$2,$3,$4::jsonb)`, actor, in.LotteryID, in.PlanetID, string(params)); err != nil {
		return "", status.Error(codes.Internal, "delivery audit failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "delivery commit failed")
	}
	list, err := lotteryWinnerRows(ctx, db, in.PlanetID, in.LotteryID)
	if err != nil {
		return "", err
	}
	for _, w := range list {
		if w["id"] == in.WinnerID {
			return encode(w)
		}
	}
	return "", status.Error(codes.Internal, "updated winner unavailable")
}
