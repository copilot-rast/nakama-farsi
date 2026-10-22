package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func brandProductStatusAllowed(current, next string) error {
	if next != "active" && next != "paused" && next != "archived" {
		return status.Error(codes.InvalidArgument, "invalid brand product status")
	}
	if current == "suspended" {
		return status.Error(codes.PermissionDenied, "operator suspended product")
	}
	if current == "archived" && next != "archived" {
		return status.Error(codes.FailedPrecondition, "archived product cannot be restored")
	}
	return nil
}

func shopBrandSetStatus(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in shopBrandInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	in.ProductID, err = requiredUUID(in.ProductID)
	if err != nil {
		return "", err
	}
	if err = brandProductStatusAllowed("active", in.Status); err != nil {
		return "", err
	}
	brandID, err := shopBrandAccess(ctx, db, in.PlanetID, actor, true)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "product status unavailable")
	}
	defer tx.Rollback()
	var current string
	err = tx.QueryRowContext(ctx, `SELECT status FROM jazebe_shop_product WHERE id=$1 AND planet_id=$2 FOR UPDATE`, in.ProductID, in.PlanetID).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "shop product not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "product status unavailable")
	}
	if err = brandProductStatusAllowed(current, in.Status); err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_shop_product SET status=$3,updated_at=now() WHERE id=$1 AND planet_id=$2`, in.ProductID, in.PlanetID, in.Status); err != nil {
		return "", status.Error(codes.Internal, "product status update failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "product status update failed")
	}
	row, err := shopBrandProductRow(ctx, db, in.PlanetID, in.ProductID, brandID)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func shopBrandLotteryOptions(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in shopBrandInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	if _, err = shopBrandAccess(ctx, db, in.PlanetID, actor, false); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT id,title,status,draw_at,ticket_price_points,per_user_ticket_cap,per_group_ticket_cap FROM jazebe_lottery WHERE planet_id=$1 AND status IN ('draft','active') ORDER BY draw_at ASC`, in.PlanetID)
	if err != nil {
		return "", status.Error(codes.Internal, "lottery options unavailable")
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, title, state string
		var draw time.Time
		var price int
		var userCap, groupCap sql.NullInt64
		if err = rows.Scan(&id, &title, &state, &draw, &price, &userCap, &groupCap); err != nil {
			return "", status.Error(codes.Internal, "lottery options unavailable")
		}
		out = append(out, map[string]any{"id": id, "title": title, "status": state, "drawAt": draw.Format(time.RFC3339Nano), "ticketPricePoints": price, "perUserTicketCap": nullInt(userCap), "perGroupTicketCap": nullInt(groupCap)})
	}
	if err = rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "lottery options unavailable")
	}
	return encode(out)
}
