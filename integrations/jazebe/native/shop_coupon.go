package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const couponAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
const couponDefaultFormat = "####-####"

var bigCouponAlphabetLen = big.NewInt(int64(len(couponAlphabet)))

type couponProductInput struct {
	ticketProductInput
	Coupon struct {
		CodeFormat string  `json:"codeFormat"`
		ExpiresAt  *string `json:"expiresAt"`
		UsageTerms string  `json:"usageTerms"`
	} `json:"coupon"`
}

func couponFormat(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = couponDefaultFormat
	}
	if len([]rune(value)) < 4 || len([]rune(value)) > 48 || !strings.ContainsRune(value, '#') {
		return "", status.Error(codes.InvalidArgument, "invalid coupon code format")
	}
	return value, nil
}

func couponExpiry(raw *string) (*time.Time, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	at, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(*raw))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid coupon expiration")
	}
	return &at, nil
}

func renderCouponCode(format string) (string, error) {
	var out strings.Builder
	for _, r := range format {
		if r != '#' {
			out.WriteRune(r)
			continue
		}
		index, err := rand.Int(rand.Reader, bigCouponAlphabetLen)
		if err != nil {
			return "", err
		}
		out.WriteByte(couponAlphabet[index.Int64()])
	}
	return out.String(), nil
}

func couponProductCreate(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in couponProductInput
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
	format, err := couponFormat(in.Coupon.CodeFormat)
	if err != nil {
		return "", err
	}
	expiry, err := couponExpiry(in.Coupon.ExpiresAt)
	if err != nil {
		return "", err
	}
	if len([]rune(in.Coupon.UsageTerms)) > 1024 {
		return "", status.Error(codes.InvalidArgument, "coupon usage terms too long")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "product creation unavailable")
	}
	defer tx.Rollback()
	var allowed bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_planet p JOIN jazebe_brand_admin a ON a.brand_id=p.brand_id WHERE p.id=$1 AND a.user_id=$2 AND a.is_active AND a.role IN ('main','marketing'))`, in.PlanetID, actor).Scan(&allowed)
	if err != nil {
		return "", status.Error(codes.Internal, "product authorization unavailable")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "brand economy administrator required")
	}
	id := uuid.Must(uuid.NewV4()).String()
	now := time.Now().UTC()
	config, _ := json.Marshal(map[string]any{"codeFormat": format, "expiresAt": expiry, "usageTerms": nullableText(in.Coupon.UsageTerms)})
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_shop_product(id,planet_id,name,short_description,long_description,image_asset,type,price_points,stock,terms,delivery_note,type_config,status,freshness_rank,price_updated_at) VALUES($1,$2,$3,$4,$5,$6,'coupon',$7,$8,$9,$10,$11::jsonb,'active',$12,$13)`, id, in.PlanetID, strings.TrimSpace(in.Name), nullableText(in.ShortDescription), nullableText(in.LongDescription), nullableText(in.ImageAsset), in.PricePoints, in.Stock, nullableText(in.Terms), nullableText(in.DeliveryNote), string(config), now.Unix(), now)
	if err != nil {
		logger.Error("Coupon product insert: %v", err)
		return "", status.Error(codes.Internal, "product creation failed")
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params) VALUES('shop.product.create','brand_admin',$1,'shop_product',$2,$3,jsonb_build_object('type','coupon'))`, actor, id, in.PlanetID)
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

type couponGenerateInput struct {
	PlanetID   string          `json:"planetId"`
	ProductID  string          `json:"productId"`
	Count      int             `json:"count"`
	CodeFormat *string         `json:"codeFormat"`
	ExpiresAt  json.RawMessage `json:"expiresAt"`
}

func couponGenerate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var in couponGenerateInput
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
	if in.Count < 1 || in.Count > 1000 {
		return "", status.Error(codes.InvalidArgument, "coupon batch must contain 1 to 1000 codes")
	}
	var kind, state string
	var config []byte
	var allowed bool
	err = db.QueryRowContext(ctx, `SELECT p.type,p.status,p.type_config,EXISTS(SELECT 1 FROM jazebe_planet q JOIN jazebe_brand_admin a ON a.brand_id=q.brand_id WHERE q.id=p.planet_id AND a.user_id=$3 AND a.is_active AND a.role IN ('main','marketing')) FROM jazebe_shop_product p WHERE p.id=$1 AND p.planet_id=$2`, in.ProductID, in.PlanetID, actor).Scan(&kind, &state, &config, &allowed)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "product not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "product lookup failed")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "brand economy administrator required")
	}
	if kind != "coupon" || state == "suspended" {
		return "", status.Error(codes.FailedPrecondition, "coupon product unavailable")
	}
	var saved struct {
		CodeFormat string  `json:"codeFormat"`
		ExpiresAt  *string `json:"expiresAt"`
	}
	if json.Unmarshal(config, &saved) != nil {
		return "", status.Error(codes.Internal, "coupon config invalid")
	}
	format := saved.CodeFormat
	if in.CodeFormat != nil && strings.TrimSpace(*in.CodeFormat) != "" {
		format = *in.CodeFormat
	}
	format, err = couponFormat(format)
	if err != nil {
		return "", err
	}
	rawExpiry := saved.ExpiresAt
	if in.ExpiresAt != nil {
		if string(in.ExpiresAt) == "null" {
			rawExpiry = nil
		} else {
			var value string
			if err = json.Unmarshal(in.ExpiresAt, &value); err != nil {
				return "", status.Error(codes.InvalidArgument, "invalid coupon expiration")
			}
			rawExpiry = &value
		}
	}
	expiry, err := couponExpiry(rawExpiry)
	if err != nil {
		return "", err
	}
	generatedCodes := make([]string, 0, in.Count)
	for attempts := 0; len(generatedCodes) < in.Count && attempts < in.Count*20; attempts++ {
		code, err := renderCouponCode(format)
		if err != nil {
			return "", status.Error(codes.Internal, "coupon randomness unavailable")
		}
		_, err = db.ExecContext(ctx, `INSERT INTO jazebe_coupon_code(id,planet_id,shop_product_id,code,expires_at) VALUES($1,$2,$3,$4,$5)`, uuid.Must(uuid.NewV4()).String(), in.PlanetID, in.ProductID, code, expiry)
		if err != nil {
			var pg *pgconn.PgError
			if errors.As(err, &pg) && pg.Code == pgerrcode.UniqueViolation {
				continue
			}
			return "", status.Error(codes.Internal, "coupon generation failed")
		}
		generatedCodes = append(generatedCodes, code)
	}
	var available, issued, expired int
	err = db.QueryRowContext(ctx, `SELECT count(*) FILTER(WHERE status='available' AND (expires_at IS NULL OR expires_at>now())),count(*) FILTER(WHERE status='issued'),count(*) FILTER(WHERE status='expired' OR status='available' AND expires_at<=now()) FROM jazebe_coupon_code WHERE shop_product_id=$1 AND deleted_at IS NULL`, in.ProductID).Scan(&available, &issued, &expired)
	if err != nil {
		return "", status.Error(codes.Internal, "coupon inventory unavailable")
	}
	return encode(map[string]any{"generated": len(generatedCodes), "codes": generatedCodes, "inventory": map[string]int{"available": available, "issued": issued, "expired": expired}})
}
