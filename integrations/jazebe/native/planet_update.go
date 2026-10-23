package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

var planetUpdateColumns = map[string]string{
	"name": "name", "slug": "slug", "latinName": "latin_name",
	"categoryId": "category_id", "shortDescription": "short_description",
	"longDescription": "long_description", "iconAsset": "icon_asset",
	"logoAsset": "logo_asset", "coverAsset": "cover_asset",
	"brandColor": "brand_color", "glyph": "glyph", "pointName": "point_name",
	"deploymentMode": "deployment_mode",
}

func planetUpdateValue(key string, raw json.RawMessage) (any, error) {
	if string(raw) == "null" {
		switch key {
		case "name", "slug", "pointName", "deploymentMode":
			return nil, status.Error(codes.InvalidArgument, key+" cannot be null")
		default:
			return nil, nil
		}
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil || !utf8.ValidString(value) {
		return nil, status.Error(codes.InvalidArgument, "invalid "+key)
	}
	value = strings.TrimSpace(value)
	switch key {
	case "name":
		return planetName(value)
	case "slug":
		return validatedSlug(value)
	case "categoryId":
		if value == "" {
			return nil, nil
		}
		return requiredUUID(value)
	case "brandColor":
		if value == "" {
			return nil, nil
		}
		if !hexColor.MatchString(value) {
			return nil, status.Error(codes.InvalidArgument, "brandColor must be #RRGGBB")
		}
	case "deploymentMode":
		if value != "cloud" && value != "on_premise" {
			return nil, status.Error(codes.InvalidArgument, "invalid deployment mode")
		}
	case "pointName", "glyph":
		if utf8.RuneCountInString(value) > 32 || (value == "" && key == "pointName") {
			return nil, status.Error(codes.InvalidArgument, "invalid "+key)
		}
	case "latinName":
		if utf8.RuneCountInString(value) > 64 {
			return nil, status.Error(codes.InvalidArgument, "invalid Latin name")
		}
	case "shortDescription", "iconAsset", "logoAsset", "coverAsset":
		if utf8.RuneCountInString(value) > 512 {
			return nil, status.Error(codes.InvalidArgument, "invalid "+key)
		}
	case "longDescription":
		if utf8.RuneCountInString(value) > 4000 {
			return nil, status.Error(codes.InvalidArgument, "invalid long description")
		}
	}
	if value == "" {
		return nil, nil
	}
	return value, nil
}

// Only role-scoped, declared identity fields are writable; status and domain
// remain separate privileged lifecycle operations.
func brandPlanetUpdate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	ownerID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input map[string]json.RawMessage
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	rawID, ok := input["planetId"]
	if !ok {
		return "", status.Error(codes.InvalidArgument, "planetId is required")
	}
	var idValue string
	if err := json.Unmarshal(rawID, &idValue); err != nil {
		return "", status.Error(codes.InvalidArgument, "invalid planetId")
	}
	planetID, err := requiredUUID(idValue)
	if err != nil {
		return "", err
	}
	delete(input, "planetId")
	if len(input) == 0 {
		return "", status.Error(codes.InvalidArgument, "at least one field is required")
	}
	args := []any{}
	assignments := make([]string, 0, len(input))
	var newName, newCategory any
	for key, raw := range input {
		column, allowed := planetUpdateColumns[key]
		if !allowed {
			return "", status.Error(codes.InvalidArgument, "unknown planet field: "+key)
		}
		value, err := planetUpdateValue(key, raw)
		if err != nil {
			return "", err
		}
		if key == "name" {
			newName = value
		}
		if key == "categoryId" {
			newCategory = value
		}
		args = append(args, value)
		assignments = append(assignments, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "planet update failed")
	}
	defer tx.Rollback()
	var brandID, planetStatus string
	err = tx.QueryRowContext(ctx, `SELECT p.brand_id, p.status FROM jazebe_planet p
		JOIN jazebe_active_brand_admin a ON a.brand_id = p.brand_id AND a.user_id=$2
			AND a.is_active AND a.role IN ('main','technical')
		WHERE p.id = $1 FOR UPDATE OF p`, planetID, ownerID).Scan(&brandID, &planetStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		logger.Error("Jazebe planet update lookup failed: %v", err)
		return "", status.Error(codes.Internal, "planet update failed")
	}
	if planetStatus == "archived" {
		return "", status.Error(codes.PermissionDenied, "archived planet cannot be changed")
	}
	if newCategory != nil {
		var active bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet_category WHERE id = $1 AND is_active)`, newCategory).Scan(&active); err != nil {
			return "", status.Error(codes.Internal, "planet update failed")
		}
		if !active {
			return "", status.Error(codes.NotFound, "active planet category not found")
		}
	}
	if newName != nil {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(774102, hashtext($1))`, brandID); err != nil {
			return "", status.Error(codes.Internal, "planet update failed")
		}
		var duplicate bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet WHERE brand_id = $1 AND id <> $2 AND lower(name) = lower($3))`, brandID, planetID, newName).Scan(&duplicate); err != nil {
			return "", status.Error(codes.Internal, "planet update failed")
		}
		if duplicate {
			return "", status.Error(codes.AlreadyExists, "planet name already exists in this brand")
		}
	}
	args = append(args, planetID)
	query := fmt.Sprintf("UPDATE jazebe_planet SET %s WHERE id = $%d", strings.Join(assignments, ", "), len(args))
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "planet name or slug already exists")
		}
		logger.Error("Jazebe planet update failed: %v", err)
		return "", status.Error(codes.Internal, "planet update failed")
	}
	fields := make([]string, 0, len(input))
	for key := range input {
		fields = append(fields, key)
	}
	sort.Strings(fields)
	fieldJSON, _ := json.Marshal(fields)
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log
		(action,actor_type,actor_id,target_type,target_id,brand_id,planet_id,params)
		VALUES ('planet.update','brand_admin',$1,'planet',$2,$3,$2,jsonb_build_object('fields',$4::jsonb))`,
		ownerID, planetID, brandID, string(fieldJSON)); err != nil {
		logger.Error("Jazebe planet update audit failed: %v", err)
		return "", status.Error(codes.Internal, "planet update failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "planet update failed")
	}
	return brandPlanetGet(ctx, logger, db, nk, fmt.Sprintf(`{"planetId":%q}`, planetID))
}
