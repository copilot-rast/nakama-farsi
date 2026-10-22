package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Source Store_purchase equivalent. Type-specific RPCs remain for existing clients.
// An already-committed order controls replay even after a product is archived.
func shopPurchase(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	if err = shopPurchaseRate(ctx, db, user); err != nil {
		return "", err
	}
	ctx = context.WithValue(ctx, shopRateContextKey{}, true)
	var in ticketPurchaseInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in, err = validateTicketPurchase(in)
	if err != nil {
		return "", err
	}
	if err = lotteryClientGuard(ctx, db, in.PlanetID, user); err != nil {
		return "", err
	}
	orderID := purchaseOrderID(user, in.ProductID, in.IdempotencyKey)
	var kind string
	err = db.QueryRowContext(ctx, `SELECT product_type_snapshot FROM jazebe_order WHERE id=$1 AND planet_id=$2 AND user_id=$3`, orderID, in.PlanetID, user).Scan(&kind)
	if errors.Is(err, sql.ErrNoRows) {
		err = db.QueryRowContext(ctx, `SELECT type FROM jazebe_shop_product WHERE id=$1 AND planet_id=$2`, in.ProductID, in.PlanetID).Scan(&kind)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.NotFound, "product not found")
		}
	}
	if err != nil {
		return "", status.Error(codes.Internal, "product lookup failed")
	}
	switch kind {
	case "ticket":
		return shopTicketPurchase(ctx, logger, db, nk, payload)
	case "coupon":
		return couponPurchase(ctx, logger, db, nk, payload)
	case "digital":
		return digitalPurchase(ctx, logger, db, nk, payload)
	default:
		return "", status.Error(codes.FailedPrecondition, "product type unsupported")
	}
}
