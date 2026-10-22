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

type shopSalesInput struct {
	PlanetID  string `json:"planetId"`
	ProductID string `json:"productId"`
	From      string `json:"from"`
	To        string `json:"to"`
	Type      string `json:"type"`
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
}

func shopSalesFilters(in shopSalesInput) (sql.NullTime, sql.NullTime, sql.NullString, error) {
	var from, to sql.NullTime
	var kind sql.NullString
	for _, item := range []struct {
		raw string
		out *sql.NullTime
	}{{in.From, &from}, {in.To, &to}} {
		if item.raw == "" {
			continue
		}
		value, err := time.Parse(time.RFC3339Nano, item.raw)
		if err != nil {
			value, err = time.Parse("2006-01-02", item.raw)
		}
		if err != nil {
			return from, to, kind, status.Error(codes.InvalidArgument, "invalid shop date range")
		}
		*item.out = sql.NullTime{Time: value, Valid: true}
	}
	if from.Valid && to.Valid && !from.Time.Before(to.Time) {
		return from, to, kind, status.Error(codes.InvalidArgument, "shop range end must follow start")
	}
	if in.Type != "" {
		switch in.Type {
		case "ticket", "coupon", "digital", "physical":
			kind = sql.NullString{String: in.Type, Valid: true}
		default:
			return from, to, kind, status.Error(codes.InvalidArgument, "invalid shop product type")
		}
	}
	return from, to, kind, nil
}

func shopBrandSales(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in shopSalesInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	from, to, kind, err := shopSalesFilters(in)
	if err != nil {
		return "", err
	}
	if _, err = shopBrandAccess(ctx, db, in.PlanetID, actor, false); err != nil {
		return "", err
	}
	var orders, buyers int
	var points int64
	err = db.QueryRowContext(ctx, `SELECT count(*),COALESCE(sum(paid_points),0),count(DISTINCT user_id) FROM jazebe_order WHERE planet_id=$1 AND ($2::timestamptz IS NULL OR purchased_at>=$2) AND ($3::timestamptz IS NULL OR purchased_at<$3) AND ($4::text IS NULL OR product_type_snapshot=$4)`, in.PlanetID, from, to, kind).Scan(&orders, &points, &buyers)
	if err != nil {
		return "", status.Error(codes.Internal, "shop sales unavailable")
	}
	topRows, err := db.QueryContext(ctx, `SELECT shop_product_id,MAX(product_name_snapshot),MAX(product_type_snapshot),count(*),COALESCE(sum(paid_points),0) FROM jazebe_order WHERE planet_id=$1 AND ($2::timestamptz IS NULL OR purchased_at>=$2) AND ($3::timestamptz IS NULL OR purchased_at<$3) AND ($4::text IS NULL OR product_type_snapshot=$4) GROUP BY shop_product_id ORDER BY count(*) DESC,sum(paid_points) DESC LIMIT 10`, in.PlanetID, from, to, kind)
	if err != nil {
		return "", status.Error(codes.Internal, "shop sales unavailable")
	}
	top := []map[string]any{}
	for topRows.Next() {
		var id, name, typ string
		var count int
		var sum int64
		if err = topRows.Scan(&id, &name, &typ, &count, &sum); err != nil {
			topRows.Close()
			return "", status.Error(codes.Internal, "shop sales unavailable")
		}
		top = append(top, map[string]any{"productId": id, "name": name, "type": typ, "orders": count, "points": sum})
	}
	err = topRows.Err()
	topRows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "shop sales unavailable")
	}
	typeRows, err := db.QueryContext(ctx, `SELECT product_type_snapshot,count(*),COALESCE(sum(paid_points),0) FROM jazebe_order WHERE planet_id=$1 AND ($2::timestamptz IS NULL OR purchased_at>=$2) AND ($3::timestamptz IS NULL OR purchased_at<$3) AND ($4::text IS NULL OR product_type_snapshot=$4) GROUP BY product_type_snapshot`, in.PlanetID, from, to, kind)
	if err != nil {
		return "", status.Error(codes.Internal, "shop sales unavailable")
	}
	byType := []map[string]any{}
	for typeRows.Next() {
		var typ string
		var count int
		var sum int64
		if err = typeRows.Scan(&typ, &count, &sum); err != nil {
			typeRows.Close()
			return "", status.Error(codes.Internal, "shop sales unavailable")
		}
		byType = append(byType, map[string]any{"type": typ, "orders": count, "points": sum})
	}
	err = typeRows.Err()
	typeRows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "shop sales unavailable")
	}
	var fromValue, toValue any
	if in.From != "" {
		fromValue = in.From
	}
	if in.To != "" {
		toValue = in.To
	}
	return encode(map[string]any{"range": map[string]any{"from": fromValue, "to": toValue}, "totals": map[string]any{"orders": orders, "points": points, "buyers": buyers, "cancelRate": 0}, "topProducts": top, "byType": byType})
}

func shopBrandOrders(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return shopOrders(ctx, db, payload, false)
}
func shopBrandProductOrders(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return shopOrders(ctx, db, payload, true)
}

func shopOrders(ctx context.Context, db *sql.DB, payload string, byProduct bool) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in shopSalesInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	if byProduct {
		in.ProductID, err = requiredUUID(in.ProductID)
		if err != nil {
			return "", err
		}
	}
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	if in.Page < 1 || in.Page > 1000000 || in.Limit < 1 || in.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid order pagination")
	}
	from, to, kind, err := shopSalesFilters(in)
	if err != nil {
		return "", err
	}
	brandID, err := shopBrandAccess(ctx, db, in.PlanetID, actor, false)
	if err != nil {
		return "", err
	}
	var productFilter sql.NullString
	if byProduct {
		var exists bool
		err = db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_shop_product WHERE id=$1 AND planet_id=$2)`, in.ProductID, in.PlanetID).Scan(&exists)
		if err != nil {
			return "", status.Error(codes.Internal, "product lookup failed")
		}
		if !exists {
			return "", status.Error(codes.NotFound, "shop product not found")
		}
		productFilter = sql.NullString{String: in.ProductID, Valid: true}
	}
	const where = ` WHERE planet_id=$1 AND ($2::timestamptz IS NULL OR purchased_at>=$2) AND ($3::timestamptz IS NULL OR purchased_at<$3) AND ($4::text IS NULL OR product_type_snapshot=$4) AND ($5::uuid IS NULL OR shop_product_id=$5)`
	var total int
	err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_order`+where, in.PlanetID, from, to, kind, productFilter).Scan(&total)
	if err != nil {
		return "", status.Error(codes.Internal, "shop order count unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT id,user_id,shop_product_id,product_name_snapshot,product_type_snapshot,paid_points,status,delivery,point_transaction_id,purchased_at FROM jazebe_order`+where+` ORDER BY purchased_at DESC,id DESC LIMIT $6 OFFSET $7`, in.PlanetID, from, to, kind, productFilter, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", status.Error(codes.Internal, "shop orders unavailable")
	}
	items := []map[string]any{}
	for rows.Next() {
		var id, user, product, name, typ, state, txID string
		var paid int
		var delivery []byte
		var purchased time.Time
		if err = rows.Scan(&id, &user, &product, &name, &typ, &paid, &state, &delivery, &txID, &purchased); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "shop orders unavailable")
		}
		var detail any
		if len(delivery) > 0 {
			if err = json.Unmarshal(delivery, &detail); err != nil {
				rows.Close()
				return "", status.Error(codes.Internal, "shop order delivery invalid")
			}
		}
		items = append(items, map[string]any{"id": id, "productId": brandID, "planetId": in.PlanetID, "userId": user, "shopProductId": product, "productNameSnapshot": name, "productTypeSnapshot": typ, "paidPoints": paid, "status": state, "delivery": detail, "pointTransactionId": txID, "purchasedAt": purchased.Format(time.RFC3339Nano)})
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "shop orders unavailable")
	}
	return encode(map[string]any{"code": "SUCCESS", "data": items, "meta": map[string]int{"total": total, "page": in.Page, "limit": in.Limit, "pages": (total + in.Limit - 1) / in.Limit}})
}
