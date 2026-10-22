package main

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type lotteryDetailPrize struct {
	ID, Title, Kind, Value, Image string
	Priority                      int
}
type lotteryDetailStat struct {
	GroupID, Name                   string
	Tickets, Individual, Collective int
}
type lotteryDetailWinner struct {
	ID, GroupID, GroupName, PrizeID, PrizeTitle, Serial, Delivery string
	Priority                                                      int
	Mine                                                          bool
	Tickets                                                       int
	Percent                                                       float64
	Value, Payout                                                 string
	Blocked                                                       bool
}
type lotteryDetailInput struct {
	PlanetID  string `json:"planetId"`
	LotteryID string `json:"lotteryId"`
}

func lotteryLifecycle(state string, inGroup bool) string {
	if state == "drawing" || state == "finished" || state == "cancelled" {
		return state
	}
	if inGroup {
		return "active"
	}
	return "nogroup"
}

func lotteryClientDetail(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in lotteryDetailInput
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
	if err = lotteryClientGuard(ctx, db, in.PlanetID, actor); err != nil {
		return "", err
	}
	var title, state string
	var draw time.Time
	var total int
	var drawn sql.NullTime
	var seed, reason sql.NullString
	err = db.QueryRowContext(ctx, `SELECT title,status,draw_at,total_ticket_count,drawn_at,draw_seed,cancel_reason FROM jazebe_lottery WHERE id=$1 AND planet_id=$2`, in.LotteryID, in.PlanetID).Scan(&title, &state, &draw, &total, &drawn, &seed, &reason)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "lottery not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "raffle detail unavailable")
	}
	if state == "draft" {
		return "", status.Error(codes.NotFound, "lottery not found")
	}
	prizes, err := loadLotteryDetailPrizes(ctx, db, in.LotteryID)
	if err != nil {
		return "", err
	}
	stats, err := loadLotteryDetailStats(ctx, db, in.LotteryID)
	if err != nil {
		return "", err
	}
	winners, err := loadLotteryDetailWinners(ctx, db, in.LotteryID, actor)
	if err != nil {
		return "", err
	}
	var activeGroup sql.NullString
	err = db.QueryRowContext(ctx, `SELECT group_id FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active' LIMIT 1`, in.PlanetID, actor).Scan(&activeGroup)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "group lookup failed")
	}
	var own *lotteryDetailWinner
	groupID := activeGroup.String
	for i := range winners {
		if winners[i].Mine {
			own = &winners[i]
			groupID = winners[i].GroupID
			break
		}
	}
	groupName := ""
	var groupStat *lotteryDetailStat
	for i := range stats {
		if stats[i].GroupID == groupID {
			groupStat = &stats[i]
			groupName = stats[i].Name
			break
		}
	}
	if groupName == "" && own != nil {
		groupName = own.GroupName
	}
	groupTickets, individual, collective := 0, 0, 0
	if groupStat != nil {
		groupTickets = groupStat.Tickets
		individual = groupStat.Individual
		collective = groupStat.Collective
	}
	prizeViews := make([]map[string]any, 0, len(prizes))
	for _, p := range prizes {
		value, _ := strconv.ParseFloat(p.Value, 64)
		prizeViews = append(prizeViews, map[string]any{"rank": p.Priority, "name": p.Title, "type": lotteryPrizeTypeFa(p.Kind), "value": lotteryRialFa(p.Value), "valNum": value, "image": lotteryPrizeImage(p.Kind, p.Image)})
	}
	participants := []map[string]any{}
	for _, s := range stats {
		if s.Tickets <= 0 {
			continue
		}
		v := map[string]any{"name": s.Name, "tickets": s.Tickets}
		if s.GroupID == groupID {
			v["me"] = true
		}
		participants = append(participants, v)
	}
	winnerViews := []map[string]any{}
	winningTickets := []string{}
	drawTickets := []string{}
	for _, w := range winners {
		v := map[string]any{"rank": w.Priority, "group": w.GroupName, "prize": w.PrizeTitle}
		if w.Mine {
			v["me"] = true
		}
		winnerViews = append(winnerViews, v)
		winningTickets = append(winningTickets, w.Serial)
		if w.GroupID == groupID {
			drawTickets = append(drawTickets, w.Serial)
		}
	}
	members := []map[string]any{}
	if groupID != "" {
		winnerID := ""
		if own != nil {
			winnerID = own.ID
		}
		members, err = lotteryDetailMembers(ctx, db, in.LotteryID, groupID, winnerID, actor)
		if err != nil {
			return "", err
		}
	}
	steps := []string{"در حال پیگیری", "تماس انجام شده", "تحویل شده"}
	var step, tracking, myShare any
	if own != nil {
		if own.Delivery == "rejected" {
			steps[2] = "رد شده"
		}
		n := 0
		if own.Delivery == "contacted" {
			n = 1
		} else if own.Delivery == "delivered" || own.Delivery == "rejected" {
			n = 2
		}
		step = n
		tracking = lotteryTrackingCode(own.ID)
		myShare = map[string]any{"tickets": own.Tickets, "percent": own.Percent, "valueRial": own.Value, "payoutStatus": own.Payout, "blockedByKyc": own.Blocked}
	}
	var drawnAt, drawSeed, cancelReason any
	if drawn.Valid {
		drawnAt = lotteryDrawInstantFa(drawn.Time)
	}
	if state == "finished" && seed.Valid {
		drawSeed = seed.String
	}
	if state == "cancelled" && reason.Valid {
		cancelReason = reason.String
	}
	chance := 0.0
	if total > 0 {
		chance = float64(groupTickets) / float64(total) * 100
	}
	return encode(map[string]any{"id": in.LotteryID, "title": title, "status": lotteryLifecycle(state, activeGroup.Valid || own != nil), "drawAt": draw.Format(time.RFC3339Nano), "groupName": groupName, "groupAchievement": collective, "individualTotal": individual, "groupTotal": groupTickets, "planetTotal": total, "groupChance": chance, "prizes": prizeViews, "members": members, "participants": participants, "winners": winnerViews, "deliverySteps": steps, "deliveryStep": step, "trackingCode": tracking, "drawTickets": drawTickets, "winningTickets": winningTickets, "drawnAt": drawnAt, "drawSeed": drawSeed, "myShare": myShare, "cancelReason": cancelReason})
}

func loadLotteryDetailPrizes(ctx context.Context, db *sql.DB, id string) ([]lotteryDetailPrize, error) {
	rows, err := db.QueryContext(ctx, `SELECT id,title,type,value_rial::text,priority,COALESCE(image_asset,'') FROM jazebe_lottery_prize WHERE lottery_id=$1 AND deleted_at IS NULL ORDER BY priority`, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "raffle prizes unavailable")
	}
	defer rows.Close()
	out := []lotteryDetailPrize{}
	for rows.Next() {
		var p lotteryDetailPrize
		if err = rows.Scan(&p.ID, &p.Title, &p.Kind, &p.Value, &p.Priority, &p.Image); err != nil {
			return nil, status.Error(codes.Internal, "raffle prizes unavailable")
		}
		out = append(out, p)
	}
	if rows.Err() != nil {
		return nil, status.Error(codes.Internal, "raffle prizes unavailable")
	}
	return out, nil
}
func loadLotteryDetailStats(ctx context.Context, db *sql.DB, id string) ([]lotteryDetailStat, error) {
	rows, err := db.QueryContext(ctx, `SELECT s.group_id,g.name,s.ticket_count,s.individual_ticket_count,s.collective_ticket_count FROM jazebe_lottery_group_stat s JOIN jazebe_group g ON g.id=s.group_id WHERE s.lottery_id=$1 ORDER BY s.ticket_count DESC,s.group_id`, id)
	if err != nil {
		return nil, status.Error(codes.Internal, "raffle pool unavailable")
	}
	defer rows.Close()
	out := []lotteryDetailStat{}
	for rows.Next() {
		var s lotteryDetailStat
		if err = rows.Scan(&s.GroupID, &s.Name, &s.Tickets, &s.Individual, &s.Collective); err != nil {
			return nil, status.Error(codes.Internal, "raffle pool unavailable")
		}
		out = append(out, s)
	}
	if rows.Err() != nil {
		return nil, status.Error(codes.Internal, "raffle pool unavailable")
	}
	return out, nil
}
func loadLotteryDetailWinners(ctx context.Context, db *sql.DB, id, user string) ([]lotteryDetailWinner, error) {
	rows, err := db.QueryContext(ctx, `SELECT w.id,w.group_id,g.name,w.prize_id,p.title,t.serial,w.delivery_status,p.priority,s.user_id IS NOT NULL,COALESCE(s.individual_ticket_count,0),COALESCE(s.share_percent,0)::text,COALESCE(s.share_value_rial,0)::text,COALESCE(s.payout_status,''),COALESCE(s.blocked_by_kyc,false) FROM jazebe_lottery_winner w JOIN jazebe_group g ON g.id=w.group_id JOIN jazebe_lottery_prize p ON p.id=w.prize_id JOIN jazebe_lottery_ticket t ON t.id=w.ticket_id LEFT JOIN jazebe_lottery_winner_share s ON s.winner_id=w.id AND s.user_id=$2 WHERE w.lottery_id=$1 ORDER BY p.priority`, id, user)
	if err != nil {
		return nil, status.Error(codes.Internal, "raffle winners unavailable")
	}
	defer rows.Close()
	out := []lotteryDetailWinner{}
	for rows.Next() {
		var w lotteryDetailWinner
		var percent string
		if err = rows.Scan(&w.ID, &w.GroupID, &w.GroupName, &w.PrizeID, &w.PrizeTitle, &w.Serial, &w.Delivery, &w.Priority, &w.Mine, &w.Tickets, &percent, &w.Value, &w.Payout, &w.Blocked); err != nil {
			return nil, status.Error(codes.Internal, "raffle winners unavailable")
		}
		w.Percent, err = strconv.ParseFloat(percent, 64)
		if err != nil {
			return nil, status.Error(codes.Internal, "raffle share invalid")
		}
		out = append(out, w)
	}
	if rows.Err() != nil {
		return nil, status.Error(codes.Internal, "raffle winners unavailable")
	}
	return out, nil
}

func lotteryDetailMembers(ctx context.Context, db *sql.DB, lotteryID, groupID, winnerID, userID string) ([]map[string]any, error) {
	var rows *sql.Rows
	var err error
	if winnerID != "" {
		rows, err = db.QueryContext(ctx, `SELECT s.user_id,COALESCE(NULLIF(trim(u.display_name),''),'کاربر'),s.individual_ticket_count FROM jazebe_lottery_winner_share s JOIN users u ON u.id=s.user_id WHERE s.winner_id=$1`, winnerID)
	} else {
		rows, err = db.QueryContext(ctx, `WITH members AS (SELECT user_id FROM jazebe_group_membership WHERE group_id=$2 AND status='active' UNION SELECT creator_user_id FROM jazebe_lottery_ticket WHERE lottery_id=$1 AND group_id=$2 AND creator_type='member' AND status='active') SELECT m.user_id,COALESCE(NULLIF(trim(u.display_name),''),'کاربر'),COUNT(t.id)::integer FROM members m JOIN users u ON u.id=m.user_id LEFT JOIN jazebe_lottery_ticket t ON t.lottery_id=$1 AND t.group_id=$2 AND t.creator_type='member' AND t.status='active' AND t.creator_user_id=m.user_id GROUP BY m.user_id,u.display_name`, lotteryID, groupID)
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "raffle members unavailable")
	}
	type member struct {
		ID, Name string
		Tickets  int
	}
	found := []member{}
	for rows.Next() {
		var m member
		if err = rows.Scan(&m.ID, &m.Name, &m.Tickets); err != nil {
			rows.Close()
			return nil, status.Error(codes.Internal, "raffle members unavailable")
		}
		found = append(found, m)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, status.Error(codes.Internal, "raffle members unavailable")
	}
	sort.Slice(found, func(i, j int) bool {
		if found[i].ID == userID {
			return true
		}
		if found[j].ID == userID {
			return false
		}
		if found[i].Tickets != found[j].Tickets {
			return found[i].Tickets > found[j].Tickets
		}
		return found[i].ID < found[j].ID
	})
	out := make([]map[string]any, 0, len(found))
	for _, m := range found {
		v := map[string]any{"name": m.Name, "tickets": m.Tickets, "color": groupMemberColor(m.ID)}
		if m.ID == userID {
			v["name"] = "تو"
			v["me"] = true
		}
		out = append(out, v)
	}
	return out, nil
}
