package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type shopPurchasesInput struct {
	PlanetID string `json:"planetId"`
	Type     string `json:"type"`
	Range    string `json:"range"`
}

func shopPurchaseFilters(in shopPurchasesInput, now time.Time) (sql.NullTime, error) {
	switch in.Type {
	case "", "all", "ticket", "coupon", "digital":
	default:
		return sql.NullTime{}, status.Error(codes.InvalidArgument, "invalid purchase type")
	}
	switch in.Range {
	case "", "all":
		return sql.NullTime{}, nil
	case "month":
		return sql.NullTime{Time: now.Add(-30 * 24 * time.Hour), Valid: true}, nil
	case "quarter":
		return sql.NullTime{Time: now.Add(-120 * 24 * time.Hour), Valid: true}, nil
	default:
		return sql.NullTime{}, status.Error(codes.InvalidArgument, "invalid purchase range")
	}
}

func shopPurchaseRecord(id, productID, name, kind string, paid int, purchased time.Time, delivery []byte, lotteryTitle string, now time.Time) (map[string]any, error) {
	var detail map[string]any
	if len(delivery) != 0 && string(delivery) != "null" {
		if err := json.Unmarshal(delivery, &detail); err != nil {
			return nil, err
		}
	}
	if detail == nil {
		detail = map[string]any{}
	}
	record := map[string]any{"id": id, "pid": productID, "name": name, "type": kind, "paid": paid, "when": shopWhenFa(purchased, now), "mo": lotteryMonthsAgo(purchased, now)}
	switch kind {
	case "ticket":
		record["code"] = nil
		record["raffle"] = lotteryTitle
		if lotteryID, ok := detail["lotteryId"].(string); ok && lotteryID != "" {
			record["raffleId"] = lotteryID
		}
	case "coupon":
		record["code"] = detail["couponCode"]
		if raw, ok := detail["couponExpiresAt"].(string); ok && raw != "" {
			expires, err := time.Parse(time.RFC3339Nano, raw)
			if err != nil {
				return nil, err
			}
			record["expiry"] = lotteryDateFa(expires)
			record["expired"] = !expires.After(now)
		}
	case "digital":
		var value any
		if payload, ok := detail["digitalPayload"].(map[string]any); ok {
			value = payload["value"]
		}
		record["code"] = value
	default:
		return nil, status.Error(codes.Internal, "unknown order product type")
	}
	return record, nil
}

// Source Page-23: only the buyer can retrieve archived delivery payloads.
func shopPurchases(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in shopPurchasesInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	since, err := shopPurchaseFilters(in, now)
	if err != nil {
		return "", err
	}
	if err = lotteryClientGuard(ctx, db, in.PlanetID, actor); err != nil {
		return "", err
	}
	filterType := sql.NullString{}
	if in.Type != "" && in.Type != "all" {
		filterType = sql.NullString{String: in.Type, Valid: true}
	}
	rows, err := db.QueryContext(ctx, `SELECT o.id,o.shop_product_id,o.product_name_snapshot,o.product_type_snapshot,o.paid_points,o.purchased_at,o.delivery,COALESCE(l.title,'') FROM jazebe_order o LEFT JOIN jazebe_lottery l ON l.id::text=o.delivery->>'lotteryId' AND l.planet_id=o.planet_id WHERE o.planet_id=$1 AND o.user_id=$2 AND ($3::text IS NULL OR o.product_type_snapshot=$3) AND ($4::timestamptz IS NULL OR o.purchased_at >= $4) ORDER BY o.purchased_at DESC,o.id DESC LIMIT 500`, in.PlanetID, actor, filterType, since)
	if err != nil {
		return "", status.Error(codes.Internal, "purchase history unavailable")
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, productID, name, kind, title string
		var paid int
		var purchased time.Time
		var delivery []byte
		if err = rows.Scan(&id, &productID, &name, &kind, &paid, &purchased, &delivery, &title); err != nil {
			return "", status.Error(codes.Internal, "purchase history unavailable")
		}
		record, err := shopPurchaseRecord(id, productID, name, kind, paid, purchased, delivery, title, now)
		if err != nil {
			return "", status.Error(codes.Internal, "purchase archive invalid")
		}
		out = append(out, record)
	}
	if err = rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "purchase history unavailable")
	}
	return encode(out)
}
