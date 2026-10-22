package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func instantPurchaseReplay(ctx context.Context, db *sql.DB, in ticketPurchaseInput, user, orderID, expectedKind string) (string, bool, error) {
	var productID, name, kind string
	var paid int
	var delivery []byte
	var purchased time.Time
	var balance int64
	err := db.QueryRowContext(ctx, `SELECT o.shop_product_id,o.product_name_snapshot,o.product_type_snapshot,o.paid_points,o.delivery,o.purchased_at,COALESCE(w.balance,0) FROM jazebe_order o LEFT JOIN jazebe_point_wallet w ON w.planet_id=o.planet_id AND w.user_id=o.user_id WHERE o.id=$1 AND o.planet_id=$2 AND o.user_id=$3`, orderID, in.PlanetID, user).Scan(&productID, &name, &kind, &paid, &delivery, &purchased, &balance)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, status.Error(codes.Internal, "purchase replay failed")
	}
	if kind != expectedKind {
		return "", false, status.Error(codes.Internal, "order type mismatch")
	}
	record, err := shopPurchaseRecord(orderID, productID, name, kind, paid, purchased, delivery, "", time.Now().UTC())
	if err != nil {
		return "", false, status.Error(codes.Internal, "order delivery invalid")
	}
	out, err := encode(map[string]any{"record": record, "balance": balance})
	return out, true, err
}

func couponPurchase(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return instantPurchase(ctx, logger, db, payload, "coupon")
}

func digitalPurchase(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return instantPurchase(ctx, logger, db, payload, "digital")
}

func instantPurchase(ctx context.Context, logger runtime.Logger, db *sql.DB, payload, expectedKind string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	if !shopRateAlreadyChecked(ctx) {
		if err = shopPurchaseRate(ctx, db, user); err != nil {
			return "", err
		}
	}
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
	if out, found, err := instantPurchaseReplay(ctx, db, in, user, orderID, expectedKind); err != nil || found {
		return out, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "purchase unavailable")
	}
	defer tx.Rollback()
	var name, kind, state string
	var price int
	var config []byte
	err = tx.QueryRowContext(ctx, `SELECT name,type,status,price_points,type_config FROM jazebe_shop_product WHERE id=$1 AND planet_id=$2`, in.ProductID, in.PlanetID).Scan(&name, &kind, &state, &price, &config)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "product not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "product lookup failed")
	}
	if kind != expectedKind {
		return "", status.Error(codes.FailedPrecondition, "product delivery type unavailable")
	}
	if state != "active" {
		return "", status.Error(codes.FailedPrecondition, "product not active")
	}
	var saved struct {
		ExpiresAt       *string `json:"expiresAt"`
		AssetKind       string  `json:"assetKind"`
		DeliveryPayload struct {
			Value string `json:"value"`
		} `json:"deliveryPayload"`
	}
	if json.Unmarshal(config, &saved) != nil {
		return "", status.Error(codes.FailedPrecondition, "product configuration invalid")
	}
	var configExpiry *time.Time
	if expectedKind == "coupon" {
		configExpiry, err = couponExpiry(saved.ExpiresAt)
		if err != nil {
			return "", status.Error(codes.FailedPrecondition, "coupon expiration invalid")
		}
	}
	if expectedKind == "digital" && !validDigitalPayload(saved.AssetKind, saved.DeliveryPayload.Value) {
		return "", status.Error(codes.FailedPrecondition, "digital configuration invalid")
	}
	now := time.Now().UTC()
	movement, err := applyPointMovement(ctx, tx, in.PlanetID, user, "debit", int64(price), "purchase", orderID, name, now)
	if errors.Is(err, errInsufficientPoints) {
		return "", status.Error(codes.FailedPrecondition, "insufficient points")
	}
	if err != nil {
		logger.Error("Shop point debit: %v", err)
		return "", status.Error(codes.Internal, "purchase failed")
	}
	if !movement.Applied {
		_ = tx.Rollback()
		out, found, err := instantPurchaseReplay(ctx, db, in, user, orderID, expectedKind)
		if err != nil {
			return "", err
		}
		if !found {
			return "", status.Error(codes.Internal, "purchase replay missing order")
		}
		return out, nil
	}
	result, err := tx.ExecContext(ctx, `UPDATE jazebe_shop_product SET stock=CASE WHEN stock IS NULL THEN NULL ELSE stock-1 END,sold_count=sold_count+1,updated_at=now() WHERE id=$1 AND status='active' AND (stock IS NULL OR stock>0)`, in.ProductID)
	if err != nil {
		return "", status.Error(codes.Internal, "product stock update failed")
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return "", status.Error(codes.ResourceExhausted, "product out of stock")
	}
	var delivery []byte
	if expectedKind == "coupon" {
		var code string
		var codeExpiry sql.NullTime
		err = tx.QueryRowContext(ctx, `UPDATE jazebe_coupon_code SET status='issued',issued_to_user_id=$2,issued_at=$3,updated_at=$3 WHERE id=(SELECT id FROM jazebe_coupon_code WHERE shop_product_id=$1 AND planet_id=$4 AND status='available' AND deleted_at IS NULL AND (expires_at IS NULL OR expires_at>$3) ORDER BY created_at,id LIMIT 1 FOR UPDATE SKIP LOCKED) RETURNING code,expires_at`, in.ProductID, user, now, in.PlanetID).Scan(&code, &codeExpiry)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.ResourceExhausted, "coupon inventory exhausted")
		}
		if err != nil {
			return "", status.Error(codes.Internal, "coupon delivery failed")
		}
		expires := configExpiry
		if codeExpiry.Valid {
			expires = &codeExpiry.Time
		}
		var expiryValue any
		if expires != nil {
			expiryValue = expires.UTC().Format(time.RFC3339Nano)
		}
		delivery, _ = json.Marshal(map[string]any{"couponCode": code, "couponExpiresAt": expiryValue})
	} else {
		delivery, _ = json.Marshal(map[string]any{"digitalPayload": map[string]any{"assetKind": saved.AssetKind, "value": saved.DeliveryPayload.Value}})
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_order(id,planet_id,user_id,shop_product_id,product_name_snapshot,product_type_snapshot,paid_points,status,delivery,point_transaction_id,purchased_at) VALUES($1,$2,$3,$4,$5,$6,$7,'delivered',$8::jsonb,$9,$10)`, orderID, in.PlanetID, user, in.ProductID, name, expectedKind, price, string(delivery), movement.TransactionID, now)
	if err != nil {
		return "", status.Error(codes.Internal, "order record failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_user_planet_stats(planet_id,user_id) VALUES($1,$2) ON CONFLICT(planet_id,user_id) DO NOTHING`, in.PlanetID, user); err != nil {
		return "", status.Error(codes.Internal, "purchase stats failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_user_planet_stats SET last_activity_at=$3 WHERE planet_id=$1 AND user_id=$2`, in.PlanetID, user, now); err != nil {
		return "", status.Error(codes.Internal, "purchase stats failed")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_planet_membership SET last_activity_at=$3 WHERE planet_id=$1 AND user_id=$2`, in.PlanetID, user, now); err != nil {
		return "", status.Error(codes.Internal, "purchase membership activity failed")
	}
	if err = enqueueActivityXP(ctx, tx, in.PlanetID, user, orderID, "purchase"); err != nil {
		return "", status.Error(codes.Internal, "purchase progression event failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "purchase failed")
	}
	shopPurchaseProgression(ctx, logger, db, in.PlanetID, user, orderID)
	record, err := shopPurchaseRecord(orderID, in.ProductID, name, expectedKind, price, now, delivery, "", now)
	if err != nil {
		return "", status.Error(codes.Internal, "purchase record invalid")
	}
	return encode(map[string]any{"record": record, "balance": movement.Balance})
}
