package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func couponInventory(ctx context.Context, db *sql.DB, productID string) (map[string]int, error) {
	var available, issued, expired int
	err := db.QueryRowContext(ctx, `SELECT count(*) FILTER(WHERE status='available' AND (expires_at IS NULL OR expires_at>now())),count(*) FILTER(WHERE status='issued'),count(*) FILTER(WHERE status='expired' OR status='available' AND expires_at<=now()) FROM jazebe_coupon_code WHERE shop_product_id=$1 AND deleted_at IS NULL`, productID).Scan(&available, &issued, &expired)
	if err != nil {
		return nil, status.Error(codes.Internal, "coupon inventory unavailable")
	}
	return map[string]int{"available": available, "issued": issued, "expired": expired}, nil
}

func shopBrandCouponInventory(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	if _, err = shopBrandAccess(ctx, db, in.PlanetID, actor, false); err != nil {
		return "", err
	}
	var present bool
	if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_shop_product WHERE id=$1 AND planet_id=$2)`, in.ProductID, in.PlanetID).Scan(&present); err != nil {
		return "", status.Error(codes.Internal, "product lookup failed")
	}
	if !present {
		return "", status.Error(codes.NotFound, "shop product not found")
	}
	inventory, err := couponInventory(ctx, db, in.ProductID)
	if err != nil {
		return "", err
	}
	return encode(inventory)
}

func shopBrandCouponList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	if in.Page < 1 || in.Page > 1000000 || in.Limit < 1 || in.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid coupon pagination")
	}
	switch in.Status {
	case "", "available", "issued", "expired":
	default:
		return "", status.Error(codes.InvalidArgument, "invalid coupon status")
	}
	brandID, err := shopBrandAccess(ctx, db, in.PlanetID, actor, false)
	if err != nil {
		return "", err
	}
	var productExists bool
	if err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_shop_product WHERE id=$1 AND planet_id=$2)`, in.ProductID, in.PlanetID).Scan(&productExists); err != nil {
		return "", status.Error(codes.Internal, "product lookup failed")
	}
	if !productExists {
		return "", status.Error(codes.NotFound, "shop product not found")
	}
	var total int
	err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_coupon_code WHERE shop_product_id=$1 AND deleted_at IS NULL AND ($2::text IS NULL OR status=$2)`, in.ProductID, optionalText(in.Status)).Scan(&total)
	if err != nil {
		return "", status.Error(codes.Internal, "coupon count unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT id,code,status,issued_to_user_id,issued_at,expires_at,created_at,updated_at FROM jazebe_coupon_code WHERE shop_product_id=$1 AND deleted_at IS NULL AND ($2::text IS NULL OR status=$2) ORDER BY created_at DESC,id DESC LIMIT $3 OFFSET $4`, in.ProductID, optionalText(in.Status), in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", status.Error(codes.Internal, "coupon list unavailable")
	}
	items := []map[string]any{}
	for rows.Next() {
		var id, code, state string
		var issuedTo sql.NullString
		var issued, expires sql.NullTime
		var created, updated time.Time
		if err = rows.Scan(&id, &code, &state, &issuedTo, &issued, &expires, &created, &updated); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "coupon list unavailable")
		}
		var userValue, issuedValue, expiresValue any
		if issuedTo.Valid {
			userValue = issuedTo.String
		}
		if issued.Valid {
			issuedValue = issued.Time.Format(time.RFC3339Nano)
		}
		if expires.Valid {
			expiresValue = expires.Time.Format(time.RFC3339Nano)
		}
		items = append(items, map[string]any{"id": id, "productId": brandID, "planetId": in.PlanetID, "shopProductId": in.ProductID, "code": code, "status": state, "issuedToUserId": userValue, "issuedAt": issuedValue, "expiresAt": expiresValue, "createdAt": created.Format(time.RFC3339Nano), "updatedAt": updated.Format(time.RFC3339Nano)})
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "coupon list unavailable")
	}
	inventory, err := couponInventory(ctx, db, in.ProductID)
	if err != nil {
		return "", err
	}
	return encode(map[string]any{"code": "SUCCESS", "data": items, "meta": map[string]int{"total": total, "page": in.Page, "limit": in.Limit, "pages": (total + in.Limit - 1) / in.Limit}, "inventory": inventory})
}
