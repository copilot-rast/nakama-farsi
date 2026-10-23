package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type digitalProductInput struct {
	ticketProductInput
	Digital struct {
		AssetKind string `json:"assetKind"`
		Value     string `json:"value"`
	} `json:"digital"`
}

func validDigitalPayload(kind, value string) bool {
	if kind != "code" && kind != "link" && kind != "file" {
		return false
	}
	value = strings.TrimSpace(value)
	return utf8.ValidString(value) && utf8.RuneCountInString(value) >= 1 && utf8.RuneCountInString(value) <= 2048
}

func digitalProductCreate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in digitalProductInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	if err = validateTicketProduct(in.ticketProductInput); err != nil {
		return "", err
	}
	if !validDigitalPayload(in.Digital.AssetKind, in.Digital.Value) {
		return "", status.Error(codes.InvalidArgument, "invalid digital delivery")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "product creation unavailable")
	}
	defer tx.Rollback()
	var allowed bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet p JOIN jazebe_active_brand_admin a ON a.brand_id=p.brand_id WHERE p.id=$1 AND a.user_id=$2 AND a.is_active AND a.role IN ('main','marketing'))`, in.PlanetID, actor).Scan(&allowed)
	if err != nil {
		return "", status.Error(codes.Internal, "product authorization unavailable")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "brand economy administrator required")
	}
	id := uuid.Must(uuid.NewV4()).String()
	now := time.Now().UTC()
	config, _ := json.Marshal(map[string]any{"assetKind": in.Digital.AssetKind, "deliveryPayload": map[string]any{"value": strings.TrimSpace(in.Digital.Value)}})
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_shop_product(id,planet_id,name,short_description,long_description,image_asset,type,price_points,stock,terms,delivery_note,type_config,status,freshness_rank,price_updated_at) VALUES($1,$2,$3,$4,$5,$6,'digital',$7,$8,$9,$10,$11::jsonb,'active',$12,$13)`, id, in.PlanetID, strings.TrimSpace(in.Name), nullableText(in.ShortDescription), nullableText(in.LongDescription), nullableText(in.ImageAsset), in.PricePoints, in.Stock, nullableText(in.Terms), nullableText(in.DeliveryNote), string(config), now.Unix(), now)
	if err != nil {
		logger.Error("Digital product insert: %v", err)
		return "", status.Error(codes.Internal, "product creation failed")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params) VALUES('shop.product.create','brand_admin',$1,'shop_product',$2,$3,jsonb_build_object('type','digital','assetKind',$4::text))`, actor, id, in.PlanetID, in.Digital.AssetKind)
	if err != nil {
		return "", status.Error(codes.Internal, "product audit failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "product creation failed")
	}
	var brandID string
	if err = db.QueryRowContext(ctx, `SELECT brand_id FROM jazebe_planet WHERE id=$1`, in.PlanetID).Scan(&brandID); err != nil {
		return "", status.Error(codes.Internal, "product response unavailable")
	}
	row, err := shopBrandProductRow(ctx, db, in.PlanetID, id, brandID)
	if err != nil {
		return "", err
	}
	return encode(row)
}
