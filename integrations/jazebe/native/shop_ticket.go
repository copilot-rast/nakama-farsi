package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ticketProductInput struct {
	PlanetID         string `json:"planetId"`
	Name             string `json:"name"`
	ShortDescription string `json:"shortDescription"`
	LongDescription  string `json:"longDescription"`
	ImageAsset       string `json:"imageAsset"`
	PricePoints      int    `json:"pricePoints"`
	Stock            *int   `json:"stock"`
	Terms            string `json:"terms"`
	DeliveryNote     string `json:"deliveryNote"`
	Ticket           struct {
		LotteryID string `json:"lotteryId"`
	} `json:"ticket"`
}

func validateTicketProduct(in ticketProductInput) error {
	if !utf8.ValidString(in.Name) || utf8.RuneCountInString(strings.TrimSpace(in.Name)) < 1 || utf8.RuneCountInString(in.Name) > 128 ||
		!utf8.ValidString(in.ShortDescription) || utf8.RuneCountInString(in.ShortDescription) > 512 ||
		!utf8.ValidString(in.LongDescription) || utf8.RuneCountInString(in.LongDescription) > 4000 ||
		!utf8.ValidString(in.ImageAsset) || utf8.RuneCountInString(in.ImageAsset) > 512 ||
		!utf8.ValidString(in.Terms) || utf8.RuneCountInString(in.Terms) > 2000 ||
		!utf8.ValidString(in.DeliveryNote) || utf8.RuneCountInString(in.DeliveryNote) > 512 ||
		in.PricePoints < 1 || in.PricePoints > 1_000_000_000 || in.Stock != nil && (*in.Stock < 0 || *in.Stock > 1_000_000_000) {
		return status.Error(codes.InvalidArgument, "invalid ticket product")
	}
	return nil
}

func shopTicketCreate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in ticketProductInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.PlanetID, err = requiredUUID(in.PlanetID)
	if err != nil {
		return "", err
	}
	in.Ticket.LotteryID, err = requiredUUID(in.Ticket.LotteryID)
	if err != nil {
		return "", err
	}
	if err = validateTicketProduct(in); err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "product creation unavailable")
	}
	defer tx.Rollback()
	var min, max int
	var allowed bool
	err = tx.QueryRowContext(ctx, `SELECT s.ticket_price_min,s.ticket_price_max,EXISTS(SELECT 1 FROM jazebe_brand_admin a WHERE a.brand_id=p.brand_id AND a.user_id=$2 AND a.is_active AND a.role IN ('main','marketing')) FROM jazebe_planet p JOIN jazebe_planet_settings s ON s.planet_id=p.id WHERE p.id=$1`, in.PlanetID, actor).Scan(&min, &max, &allowed)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "product policy unavailable")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	if in.PricePoints < min || in.PricePoints > max {
		return "", status.Error(codes.FailedPrecondition, "ticket price outside planet policy")
	}
	var lotteryStatus string
	err = tx.QueryRowContext(ctx, `SELECT status FROM jazebe_lottery WHERE id=$1 AND planet_id=$2`, in.Ticket.LotteryID, in.PlanetID).Scan(&lotteryStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.InvalidArgument, "lottery not found on planet")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "lottery lookup failed")
	}
	if lotteryStatus != "draft" && lotteryStatus != "active" {
		return "", status.Error(codes.FailedPrecondition, "lottery cannot be sold")
	}
	id := uuid.Must(uuid.NewV4()).String()
	config, _ := json.Marshal(map[string]string{"lotteryId": in.Ticket.LotteryID})
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_shop_product(id,planet_id,name,short_description,long_description,image_asset,type,price_points,stock,terms,delivery_note,type_config,status,freshness_rank,price_updated_at) VALUES($1,$2,$3,$4,$5,$6,'ticket',$7,$8,$9,$10,$11::jsonb,'active',$12,$13)`, id, in.PlanetID, strings.TrimSpace(in.Name), nullableText(in.ShortDescription), nullableText(in.LongDescription), nullableText(in.ImageAsset), in.PricePoints, in.Stock, nullableText(in.Terms), nullableText(in.DeliveryNote), string(config), time.Now().Unix(), time.Now().UTC())
	if err != nil {
		logger.Error("Ticket product insert: %v", err)
		return "", status.Error(codes.Internal, "product creation failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params) VALUES('shop.product.create','brand_admin',$1,'shop_product',$2,$3,jsonb_build_object('type','ticket','lotteryId',$4::uuid))`, actor, id, in.PlanetID, in.Ticket.LotteryID); err != nil {
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
