package main

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type lotteryFinishedInput struct {
	PlanetID string `json:"planetId"`
	Range    string `json:"range"`
}

func lotteryRangeMax(value string) (int, error) {
	switch value {
	case "", "all":
		return 999, nil
	case "m1":
		return 0, nil
	case "m3":
		return 2, nil
	case "y":
		return 11, nil
	}
	return 0, status.Error(codes.InvalidArgument, "invalid raffle range")
}

func lotteryClientFinished(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in lotteryFinishedInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	max, err := lotteryRangeMax(in.Range)
	if err != nil {
		return "", err
	}
	if err = lotteryClientGuard(ctx, db, in.PlanetID, actor); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT id,title,COALESCE(banner_asset,''),status,COALESCE(drawn_at,cancelled_at,draw_at) FROM jazebe_lottery WHERE planet_id=$1 AND status IN ('finished','cancelled') ORDER BY COALESCE(drawn_at,cancelled_at,draw_at) DESC,id DESC`, in.PlanetID)
	if err != nil {
		return "", status.Error(codes.Internal, "finished raffles unavailable")
	}
	type item struct {
		ID, Title, Banner, State string
		At                       time.Time
	}
	items := []item{}
	for rows.Next() {
		var v item
		if err = rows.Scan(&v.ID, &v.Title, &v.Banner, &v.State, &v.At); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "finished raffles unavailable")
		}
		items = append(items, v)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "finished raffles unavailable")
	}
	now := time.Now().UTC()
	out := []map[string]any{}
	for _, v := range items {
		months := lotteryMonthsAgo(v.At, now)
		if months > max {
			continue
		}
		var prizeCount int
		var firstKind, firstImage sql.NullString
		if err = db.QueryRowContext(ctx, `SELECT count(*),min(type) FILTER(WHERE priority=(SELECT min(priority) FROM jazebe_lottery_prize WHERE lottery_id=$1 AND deleted_at IS NULL)),min(image_asset) FILTER(WHERE priority=(SELECT min(priority) FROM jazebe_lottery_prize WHERE lottery_id=$1 AND deleted_at IS NULL)) FROM jazebe_lottery_prize WHERE lottery_id=$1 AND deleted_at IS NULL`, v.ID).Scan(&prizeCount, &firstKind, &firstImage); err != nil {
			return "", status.Error(codes.Internal, "raffle prize summary failed")
		}
		image := strings.TrimSpace(v.Banner)
		if image == "" {
			image = lotteryPrizeImage(firstKind.String, firstImage.String)
		}
		winnerRows, err := db.QueryContext(ctx, `SELECT g.name,EXISTS(SELECT 1 FROM jazebe_lottery_winner_share s WHERE s.winner_id=w.id AND s.user_id=$2) FROM jazebe_lottery_winner w JOIN jazebe_group g ON g.id=w.group_id JOIN jazebe_lottery_prize p ON p.id=w.prize_id WHERE w.lottery_id=$1 ORDER BY p.priority`, v.ID, actor)
		if err != nil {
			return "", status.Error(codes.Internal, "raffle winners unavailable")
		}
		names := []string{}
		seen := map[string]bool{}
		var mine string
		for winnerRows.Next() {
			var name string
			var own bool
			if err = winnerRows.Scan(&name, &own); err != nil {
				winnerRows.Close()
				return "", status.Error(codes.Internal, "raffle winners unavailable")
			}
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
			if own && mine == "" {
				mine = name
			}
		}
		err = winnerRows.Err()
		winnerRows.Close()
		if err != nil {
			return "", status.Error(codes.Internal, "raffle winners unavailable")
		}
		card := map[string]any{"id": v.ID, "title": v.Title, "image": image, "date": lotteryDateFa(v.At), "monthsAgo": months, "prizes": prizeCount, "winners": names, "status": v.State}
		if mine != "" {
			card["mine"] = mine
		}
		out = append(out, card)
	}
	return encode(out)
}
