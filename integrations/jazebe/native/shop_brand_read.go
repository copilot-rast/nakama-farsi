package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type shopBrandInput struct {
	PlanetID  string `json:"planetId"`
	ProductID string `json:"productId"`
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	Search    string `json:"search"`
}

func validateShopBrandList(in *shopBrandInput) error {
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	if in.Page < 1 || in.Limit < 1 || in.Limit > 100 || in.Page > 1000000 {
		return status.Error(codes.InvalidArgument, "invalid shop pagination")
	}
	switch in.Status {
	case "", "active", "paused", "archived", "suspended":
	default:
		return status.Error(codes.InvalidArgument, "invalid product status")
	}
	switch in.Type {
	case "", "ticket", "coupon", "digital", "physical":
	default:
		return status.Error(codes.InvalidArgument, "invalid product type")
	}
	if len([]rune(in.Search)) > 128 {
		return status.Error(codes.InvalidArgument, "product search too long")
	}
	return nil
}

func shopBrandAccess(ctx context.Context, db *sql.DB, planetID, actor string, write bool) (string, error) {
	var brandID string
	var role sql.NullString
	err := db.QueryRowContext(ctx, `SELECT p.brand_id::text,(SELECT a.role FROM jazebe_brand_admin a WHERE a.brand_id=p.brand_id AND a.user_id=$2 AND a.is_active LIMIT 1) FROM jazebe_planet p WHERE p.id=$1`, planetID, actor).Scan(&brandID, &role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "shop authorization unavailable")
	}
	if !role.Valid {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	if write && role.String != "main" && role.String != "marketing" {
		return "", status.Error(codes.PermissionDenied, "brand economy role required")
	}
	return brandID, nil
}

func shopBrandProductRow(ctx context.Context, db *sql.DB, planetID, productID, brandID string) (map[string]any, error) {
	var name, kind, state string
	var short, long, image, terms, note sql.NullString
	var stock sql.NullInt64
	var price, sold int
	var config []byte
	var priced sql.NullTime
	var created, updated time.Time
	err := db.QueryRowContext(ctx, `SELECT name,type,status,short_description,long_description,image_asset,terms,delivery_note,stock,price_points,sold_count,type_config,price_updated_at,created_at,updated_at FROM jazebe_shop_product WHERE id=$1 AND planet_id=$2`, productID, planetID).Scan(&name, &kind, &state, &short, &long, &image, &terms, &note, &stock, &price, &sold, &config, &priced, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "shop product not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "shop product unavailable")
	}
	var details map[string]any
	if len(config) > 0 && string(config) != "null" {
		if err = json.Unmarshal(config, &details); err != nil {
			return nil, status.Error(codes.Internal, "shop product configuration invalid")
		}
	}
	if details == nil {
		details = map[string]any{}
	}
	var ticket, coupon, digital, suspension any
	switch kind {
	case "ticket":
		lotteryID, _ := details["lotteryId"].(string)
		var lottery any
		if lotteryID != "" {
			var title, statusValue string
			var draw time.Time
			err = db.QueryRowContext(ctx, `SELECT title,status,draw_at FROM jazebe_lottery WHERE id::text=$1 AND planet_id=$2`, lotteryID, planetID).Scan(&title, &statusValue, &draw)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, status.Error(codes.Internal, "lottery option unavailable")
			}
			if err == nil {
				lottery = map[string]any{"id": lotteryID, "title": title, "status": statusValue, "drawAt": draw.Format(time.RFC3339Nano)}
			}
		}
		ticket = map[string]any{"lotteryId": lotteryID, "lottery": lottery}
	case "coupon":
		var available, issued, expired int
		err = db.QueryRowContext(ctx, `SELECT count(*) FILTER(WHERE status='available' AND (expires_at IS NULL OR expires_at>now())),count(*) FILTER(WHERE status='issued'),count(*) FILTER(WHERE status='expired' OR status='available' AND expires_at<=now()) FROM jazebe_coupon_code WHERE shop_product_id=$1 AND deleted_at IS NULL`, productID).Scan(&available, &issued, &expired)
		if err != nil {
			return nil, status.Error(codes.Internal, "coupon inventory unavailable")
		}
		format := details["codeFormat"]
		if format == nil {
			format = couponDefaultFormat
		}
		coupon = map[string]any{"codeFormat": format, "expiresAt": details["expiresAt"], "usageTerms": details["usageTerms"], "inventory": map[string]int{"available": available, "issued": issued, "expired": expired}}
	case "digital":
		digital = map[string]any{"assetKind": details["assetKind"], "deliveryPayload": details["deliveryPayload"]}
	}
	if state == "suspended" {
		var params []byte
		var at time.Time
		err = db.QueryRowContext(ctx, `SELECT params,created_at FROM jazebe_audit_log WHERE target_type='shop_product' AND target_id=$1 AND action='shop.product.suspend' ORDER BY created_at DESC LIMIT 1`, productID).Scan(&params, &at)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.Internal, "product suspension unavailable")
		}
		var reason any
		var atValue any
		if err == nil {
			var meta map[string]any
			if json.Unmarshal(params, &meta) == nil {
				reason = meta["reason"]
			}
			atValue = at.Format(time.RFC3339Nano)
		}
		suspension = map[string]any{"reason": reason, "at": atValue}
	}
	var pricedValue any
	if priced.Valid {
		pricedValue = priced.Time.Format(time.RFC3339Nano)
	}
	return map[string]any{"id": productID, "planetId": planetID, "productId": brandID, "name": name, "shortDescription": nullString(short), "longDescription": nullString(long), "imageAsset": nullString(image), "type": kind, "pricePoints": price, "stock": nullInt(stock), "soldCount": sold, "terms": nullString(terms), "deliveryNote": nullString(note), "status": state, "priceUpdatedAt": pricedValue, "createdAt": created.Format(time.RFC3339Nano), "updatedAt": updated.Format(time.RFC3339Nano), "ticket": ticket, "coupon": coupon, "digital": digital, "suspension": suspension}, nil
}

func shopBrandGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	brandID, err := shopBrandAccess(ctx, db, in.PlanetID, actor, false)
	if err != nil {
		return "", err
	}
	row, err := shopBrandProductRow(ctx, db, in.PlanetID, in.ProductID, brandID)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func shopBrandList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	if err = validateShopBrandList(&in); err != nil {
		return "", err
	}
	brandID, err := shopBrandAccess(ctx, db, in.PlanetID, actor, false)
	if err != nil {
		return "", err
	}
	search := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(in.Search)
	var total int
	err = db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_shop_product WHERE planet_id=$1 AND ($2::text IS NULL OR status=$2) AND ($3::text IS NULL OR type=$3) AND ($4::text IS NULL OR name ILIKE $4 ESCAPE E'\\')`, in.PlanetID, optionalText(in.Status), optionalText(in.Type), optionalSearch(search)).Scan(&total)
	if err != nil {
		return "", status.Error(codes.Internal, "shop product count unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT id FROM jazebe_shop_product WHERE planet_id=$1 AND ($2::text IS NULL OR status=$2) AND ($3::text IS NULL OR type=$3) AND ($4::text IS NULL OR name ILIKE $4 ESCAPE E'\\') ORDER BY created_at DESC,id DESC LIMIT $5 OFFSET $6`, in.PlanetID, optionalText(in.Status), optionalText(in.Type), optionalSearch(search), in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", status.Error(codes.Internal, "shop products unavailable")
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "shop products unavailable")
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "shop products unavailable")
	}
	items := []map[string]any{}
	for _, id := range ids {
		row, err := shopBrandProductRow(ctx, db, in.PlanetID, id, brandID)
		if err != nil {
			return "", err
		}
		items = append(items, row)
	}
	pages := (total + in.Limit - 1) / in.Limit
	return encode(map[string]any{"code": "SUCCESS", "data": items, "meta": map[string]int{"total": total, "page": in.Page, "limit": in.Limit, "pages": pages}})
}

func optionalText(value string) any {
	if value == "" {
		return nil
	}
	return value
}
func optionalSearch(value string) any {
	if value == "" {
		return nil
	}
	return "%" + value + "%"
}
