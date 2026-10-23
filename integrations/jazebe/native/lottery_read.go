package main

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type lotteryListInput struct {
	PlanetID string `json:"planetId"`
	Status   string `json:"status"`
	Search   string `json:"search"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
}

func lotteryBrandMeta(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	var name string
	var min, max int
	var allowed bool
	err = db.QueryRowContext(ctx, `SELECT p.point_name,s.ticket_price_min,s.ticket_price_max,EXISTS(SELECT 1 FROM jazebe_active_brand_admin a WHERE a.brand_id=p.brand_id AND a.user_id=$2 AND a.is_active) FROM jazebe_planet p JOIN jazebe_planet_settings s ON s.planet_id=p.id WHERE p.id=$1`, in.PlanetID, actor).Scan(&name, &min, &max, &allowed)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "lottery metadata unavailable")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	return encode(map[string]any{"pointName": name, "ticketPriceMin": min, "ticketPriceMax": max, "prizeTypes": []string{"cash", "brand_credit", "planet_points", "discount_coupon"}})
}

func lotteryBrandAccess(ctx context.Context, db *sql.DB, planetID, actor string) error {
	var allowed bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet p JOIN jazebe_active_brand_admin a ON a.brand_id=p.brand_id WHERE p.id=$1 AND a.user_id=$2 AND a.is_active)`, planetID, actor).Scan(&allowed)
	if err != nil {
		return status.Error(codes.Internal, "lottery authorization failed")
	}
	if !allowed {
		return status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	return nil
}

func lotteryBrandList(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in lotteryListInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	if err = lotteryBrandAccess(ctx, db, in.PlanetID, actor); err != nil {
		return "", err
	}
	if in.Status != "" {
		switch in.Status {
		case "draft", "active", "drawing", "finished", "cancelled":
		default:
			return "", status.Error(codes.InvalidArgument, "invalid lottery status")
		}
	}
	in.Search = strings.TrimSpace(in.Search)
	if len(in.Search) > 128 {
		return "", status.Error(codes.InvalidArgument, "search too long")
	}
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	if in.Page < 1 || in.Limit < 1 || in.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid page")
	}
	var total int
	err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_lottery WHERE planet_id=$1 AND ($2='' OR status=$2) AND ($3='' OR title ILIKE '%'||$3||'%')`, in.PlanetID, in.Status, in.Search).Scan(&total)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery list failed")
	}
	rows, err := db.QueryContext(ctx, `SELECT id FROM jazebe_lottery WHERE planet_id=$1 AND ($2='' OR status=$2) AND ($3='' OR title ILIKE '%'||$3||'%') ORDER BY draw_at DESC,id DESC LIMIT $4 OFFSET $5`, in.PlanetID, in.Status, in.Search, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery list failed")
	}
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "lottery list failed")
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "lottery list failed")
	}
	list := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		row, err := lotteryBrandRow(ctx, db, in.PlanetID, id)
		if err != nil {
			return "", err
		}
		list = append(list, row)
	}
	return encode(map[string]any{"items": list, "total": total, "page": in.Page, "limit": in.Limit})
}

func lotteryBrandGet(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	row, err := lotteryBrandRow(ctx, db, in.PlanetID, in.LotteryID)
	if err != nil {
		return "", err
	}
	return encode(row)
}
func nullString(v sql.NullString) any {
	if !v.Valid {
		return nil
	}
	return v.String
}
func nullInt(v sql.NullInt64) any {
	if !v.Valid {
		return nil
	}
	return v.Int64
}

// Pool statistics are read from the same counters maintained by ticket issuance.
func lotteryBrandPool(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	var state string
	var total int
	var draw any
	err = db.QueryRowContext(ctx, `SELECT status,total_ticket_count,draw_at FROM jazebe_lottery WHERE id=$1 AND planet_id=$2`, in.LotteryID, in.PlanetID).Scan(&state, &total, &draw)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "lottery not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "lottery pool unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT s.group_id,g.name,s.ticket_count,s.individual_ticket_count,s.collective_ticket_count,s.rank FROM jazebe_lottery_group_stat s JOIN jazebe_group g ON g.id=s.group_id WHERE s.lottery_id=$1 AND s.ticket_count>0 ORDER BY s.ticket_count DESC,s.group_id`, in.LotteryID)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery pool unavailable")
	}
	defer rows.Close()
	groups := make([]map[string]any, 0)
	for rows.Next() {
		var groupID, name string
		var count, individual, collective int
		var rank sql.NullInt64
		if err = rows.Scan(&groupID, &name, &count, &individual, &collective, &rank); err != nil {
			return "", status.Error(codes.Internal, "lottery pool unavailable")
		}
		chance := 0.0
		if total > 0 {
			chance = math.Round(float64(count)/float64(total)*10000) / 100
		}
		groups = append(groups, map[string]any{"groupId": groupID, "name": name, "tickets": count, "individual": individual, "collective": collective, "chancePercent": chance, "rank": nullInt(rank)})
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "lottery pool unavailable")
	}
	return encode(map[string]any{"lotteryId": in.LotteryID, "status": state, "drawAt": draw, "totalTicketCount": total, "groups": groups})
}
