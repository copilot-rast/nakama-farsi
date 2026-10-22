package main

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type lotteryClientInput struct {
	PlanetID string `json:"planetId"`
}

func lotteryClientGuard(ctx context.Context, db *sql.DB, planetID, actor string) error {
	var planetStatus string
	var brandActive, memberAllowed bool
	err := db.QueryRowContext(ctx, `SELECT p.status,b.is_active,COALESCE((SELECT m.is_active FROM jazebe_planet_membership m WHERE m.planet_id=p.id AND m.user_id=$2),true) FROM jazebe_planet p JOIN jazebe_brand b ON b.id=p.brand_id WHERE p.id=$1`, planetID, actor).Scan(&planetStatus, &brandActive, &memberAllowed)
	if errors.Is(err, sql.ErrNoRows) {
		return status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		return status.Error(codes.Internal, "planet lookup failed")
	}
	if planetStatus != "active" || !brandActive {
		return status.Error(codes.PermissionDenied, "planet unavailable")
	}
	if !memberAllowed {
		return status.Error(codes.PermissionDenied, "planet membership suspended")
	}
	return nil
}

// The active raffle tab is public to authenticated users on an active planet,
// including users who have not joined a group. Suspended fans cannot read it.
func lotteryClientActive(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in lotteryClientInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	if err = lotteryClientGuard(ctx, db, in.PlanetID, actor); err != nil {
		return "", err
	}
	var groupID sql.NullString
	err = db.QueryRowContext(ctx, `SELECT group_id FROM jazebe_group_membership WHERE planet_id=$1 AND user_id=$2 AND status='active' LIMIT 1`, in.PlanetID, actor).Scan(&groupID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "group lookup failed")
	}
	rows, err := db.QueryContext(ctx, `SELECT l.id,l.title,COALESCE(l.banner_asset,''),l.status,l.draw_at,COALESCE(s.ticket_count,0) FROM jazebe_lottery l LEFT JOIN jazebe_lottery_group_stat s ON s.lottery_id=l.id AND s.group_id=$2::uuid WHERE l.planet_id=$1 AND l.status IN ('active','drawing') ORDER BY l.draw_at ASC,l.id ASC`, in.PlanetID, nullString(groupID))
	if err != nil {
		return "", status.Error(codes.Internal, "raffle list unavailable")
	}
	type card struct {
		ID, Title, Banner, State string
		Draw                     time.Time
		GroupTickets             int
	}
	cards := []card{}
	for rows.Next() {
		var c card
		if err = rows.Scan(&c.ID, &c.Title, &c.Banner, &c.State, &c.Draw, &c.GroupTickets); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "raffle list unavailable")
		}
		cards = append(cards, c)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "raffle list unavailable")
	}
	result := make([]map[string]any, 0, len(cards))
	now := time.Now().UTC()
	for _, c := range cards {
		pr, err := db.QueryContext(ctx, `SELECT type,COALESCE(image_asset,'') FROM jazebe_lottery_prize WHERE lottery_id=$1 AND deleted_at IS NULL ORDER BY priority`, c.ID)
		if err != nil {
			return "", status.Error(codes.Internal, "raffle prizes unavailable")
		}
		types := []string{}
		image := strings.TrimSpace(c.Banner)
		seen := map[string]bool{}
		prizeCount := 0
		for pr.Next() {
			var kind, asset string
			if err = pr.Scan(&kind, &asset); err != nil {
				pr.Close()
				return "", status.Error(codes.Internal, "raffle prizes unavailable")
			}
			prizeCount++
			if !seen[kind] {
				seen[kind] = true
				types = append(types, lotteryPrizeTypeFa(kind))
			}
			if image == "" {
				image = lotteryPrizeImage(kind, asset)
			}
		}
		err = pr.Err()
		pr.Close()
		if err != nil {
			return "", status.Error(codes.Internal, "raffle prizes unavailable")
		}
		if image == "" {
			image = "prize-gift"
		}
		result = append(result, map[string]any{"id": c.ID, "title": c.Title, "image": image, "prizes": prizeCount, "ptype": strings.Join(types, "، "), "groupTickets": c.GroupTickets, "hoursLeft": lotteryHoursUntil(c.Draw, now), "status": c.State, "drawAt": c.Draw.Format(time.RFC3339Nano)})
	}
	return encode(result)
}

func lotteryHoursUntil(draw, now time.Time) int {
	return int(math.Max(0, math.Ceil(draw.Sub(now).Hours())))
}
func lotteryPrizeTypeFa(kind string) string {
	switch kind {
	case "cash":
		return "جایزه نقدی"
	case "brand_credit":
		return "اعتبار برند"
	case "planet_points":
		return "امتیاز سیاره"
	case "orbit":
		return "مدار"
	case "discount_coupon":
		return "کوپن خرید"
	}
	return ""
}
func lotteryPrizeImage(kind, asset string) string {
	if s := strings.TrimSpace(asset); s != "" {
		return s
	}
	if kind == "discount_coupon" {
		return "prize-percent"
	}
	return "prize-gift"
}
