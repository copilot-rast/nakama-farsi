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

func patchShopText(patch map[string]json.RawMessage, key string, current *sql.NullString, required bool) error {
	raw, present := patch[key]
	if !present {
		return nil
	}
	if string(raw) == "null" {
		if required {
			return status.Error(codes.InvalidArgument, "product name required")
		}
		*current = sql.NullString{}
		return nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return status.Error(codes.InvalidArgument, "invalid product text")
	}
	value = strings.TrimSpace(value)
	if value == "" && !required {
		*current = sql.NullString{}
		return nil
	}
	*current = sql.NullString{String: value, Valid: true}
	return nil
}

func patchShopInt(patch map[string]json.RawMessage, key string, current *sql.NullInt64, required bool) error {
	raw, present := patch[key]
	if !present {
		return nil
	}
	if string(raw) == "null" {
		if required {
			return status.Error(codes.InvalidArgument, "product price required")
		}
		*current = sql.NullInt64{}
		return nil
	}
	var value int64
	if err := json.Unmarshal(raw, &value); err != nil || value < 0 || value > 1000000000 || required && value == 0 {
		return status.Error(codes.InvalidArgument, "invalid product number")
	}
	*current = sql.NullInt64{Int64: value, Valid: true}
	return nil
}

func shopBrandUpdate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	var patch map[string]json.RawMessage
	if err = json.Unmarshal([]byte(payload), &patch); err != nil {
		return "", status.Error(codes.InvalidArgument, "invalid product update")
	}
	brandID, err := shopBrandAccess(ctx, db, in.PlanetID, actor, true)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "product update unavailable")
	}
	defer tx.Rollback()
	var name, short, long, image, terms, note sql.NullString
	var price, stock sql.NullInt64
	var kind, state string
	var config []byte
	var priced sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT name,short_description,long_description,image_asset,terms,delivery_note,price_points,stock,type,status,type_config,price_updated_at FROM jazebe_shop_product WHERE id=$1 AND planet_id=$2 FOR UPDATE`, in.ProductID, in.PlanetID).Scan(&name, &short, &long, &image, &terms, &note, &price, &stock, &kind, &state, &config, &priced)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "shop product not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "product lookup failed")
	}
	if state == "suspended" {
		return "", status.Error(codes.PermissionDenied, "operator suspended product")
	}
	if err = patchShopText(patch, "name", &name, true); err != nil {
		return "", err
	}
	for _, field := range []struct {
		key   string
		value *sql.NullString
	}{{"shortDescription", &short}, {"longDescription", &long}, {"imageAsset", &image}, {"terms", &terms}, {"deliveryNote", &note}} {
		if err = patchShopText(patch, field.key, field.value, false); err != nil {
			return "", err
		}
	}
	previousPrice := price.Int64
	if err = patchShopInt(patch, "pricePoints", &price, true); err != nil {
		return "", err
	}
	if err = patchShopInt(patch, "stock", &stock, false); err != nil {
		return "", err
	}
	validation := ticketProductInput{Name: name.String, ShortDescription: short.String, LongDescription: long.String, ImageAsset: image.String, PricePoints: int(price.Int64), Stock: nil, Terms: terms.String, DeliveryNote: note.String}
	if stock.Valid {
		v := int(stock.Int64)
		validation.Stock = &v
	}
	if err = validateTicketProduct(validation); err != nil {
		return "", err
	}
	if kind == "ticket" && price.Int64 != previousPrice {
		var min, max int
		err = tx.QueryRowContext(ctx, `SELECT COALESCE((SELECT ticket_price_min FROM jazebe_planet_settings WHERE planet_id=$1),1),COALESCE((SELECT ticket_price_max FROM jazebe_planet_settings WHERE planet_id=$1),1000)`, in.PlanetID).Scan(&min, &max)
		if err != nil {
			return "", status.Error(codes.Internal, "ticket price policy unavailable")
		}
		if int(price.Int64) < min || int(price.Int64) > max {
			return "", status.Error(codes.FailedPrecondition, "ticket price outside planet policy")
		}
	}
	var details map[string]any
	if len(config) > 0 && string(config) != "null" {
		if json.Unmarshal(config, &details) != nil {
			return "", status.Error(codes.Internal, "product configuration invalid")
		}
	}
	if details == nil {
		details = map[string]any{}
	}
	if raw, ok := patch[kind]; ok && string(raw) != "null" {
		var changed map[string]any
		if json.Unmarshal(raw, &changed) != nil {
			return "", status.Error(codes.InvalidArgument, "invalid product configuration")
		}
		switch kind {
		case "ticket":
			if rawID, present := changed["lotteryId"]; present {
				if rawID == nil { break }
				id, ok := rawID.(string)
				if !ok {
					return "", status.Error(codes.InvalidArgument, "invalid ticket lottery")
				}
				id, err = requiredUUID(id)
				if err != nil {
					return "", err
				}
				var state string
				err = tx.QueryRowContext(ctx, `SELECT status FROM jazebe_lottery WHERE id=$1 AND planet_id=$2`, id, in.PlanetID).Scan(&state)
				if err != nil || state != "draft" && state != "active" {
					return "", status.Error(codes.FailedPrecondition, "ticket lottery unavailable")
				}
				details["lotteryId"] = id
			}
		case "coupon":
			for _, key := range []string{"codeFormat", "expiresAt", "usageTerms"} {
				if value, ok := changed[key]; ok {
					details[key] = value
				}
			}
			if raw, ok := details["codeFormat"].(string); ok {
				format, err := couponFormat(raw)
				if err != nil {
					return "", err
				}
				details["codeFormat"] = format
			}
			if value, ok := details["expiresAt"].(string); ok {
				if _, err = couponExpiry(&value); err != nil {
					return "", err
				}
			}
			if value, ok := details["usageTerms"].(string); ok && len([]rune(value)) > 1024 {
				return "", status.Error(codes.InvalidArgument, "coupon usage terms too long")
			}
		case "digital":
			if value, ok := changed["assetKind"]; ok {
				details["assetKind"] = value
			}
			if value, ok := changed["value"]; ok {
				details["deliveryPayload"] = map[string]any{"value": value}
			}
			asset, _ := details["assetKind"].(string)
			delivery, _ := details["deliveryPayload"].(map[string]any)
			value, _ := delivery["value"].(string)
			if !validDigitalPayload(asset, value) {
				return "", status.Error(codes.InvalidArgument, "invalid digital delivery")
			}
		}
	}
	config, err = json.Marshal(details)
	if err != nil {
		return "", status.Error(codes.Internal, "product configuration invalid")
	}
	var pricedValue any = priced
	if price.Int64 != previousPrice {
		pricedValue = time.Now().UTC()
	}
	_, err = tx.ExecContext(ctx, `UPDATE jazebe_shop_product SET name=$3,short_description=$4,long_description=$5,image_asset=$6,terms=$7,delivery_note=$8,price_points=$9,stock=$10,type_config=$11::jsonb,price_updated_at=$12,updated_at=now() WHERE id=$1 AND planet_id=$2`, in.ProductID, in.PlanetID, name.String, short, long, image, terms, note, price.Int64, stock, string(config), pricedValue)
	if err != nil {
		return "", status.Error(codes.Internal, "product update failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "product update failed")
	}
	row, err := shopBrandProductRow(ctx, db, in.PlanetID, in.ProductID, brandID)
	if err != nil {
		return "", err
	}
	return encode(row)
}
