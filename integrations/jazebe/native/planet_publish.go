package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type publishFields struct {
	name, shortDescription, pointName        string
	categoryID, brandColor, iconAsset, glyph sql.NullString
}

func publishMissing(p publishFields, activeMission bool) []string {
	missing := make([]string, 0, 7)
	if strings.TrimSpace(p.name) == "" {
		missing = append(missing, "name")
	}
	if strings.TrimSpace(p.shortDescription) == "" {
		missing = append(missing, "short_description")
	}
	if !p.categoryID.Valid {
		missing = append(missing, "category")
	}
	if strings.TrimSpace(p.pointName) == "" {
		missing = append(missing, "point_name")
	}
	if !p.brandColor.Valid || p.brandColor.String == "" {
		missing = append(missing, "brand_color")
	}
	if (!p.iconAsset.Valid || p.iconAsset.String == "") && (!p.glyph.Valid || p.glyph.String == "") {
		missing = append(missing, "icon_or_glyph")
	}
	if !activeMission {
		missing = append(missing, "active_mission")
	}
	return missing
}

func readPublishState(ctx context.Context, tx *sql.Tx, ownerID, planetID string) (string, []string, error) {
	var fields publishFields
	var state string
	err := tx.QueryRowContext(ctx, `SELECT p.status,p.name,COALESCE(p.short_description,''),p.category_id,
		p.point_name,p.brand_color,p.icon_asset,p.glyph FROM jazebe_planet p
		JOIN jazebe_active_brand_admin a ON a.brand_id=p.brand_id AND a.user_id=$2
			AND a.is_active AND a.role IN ('main','technical')
		WHERE p.id=$1 FOR UPDATE OF p`,
		planetID, ownerID).Scan(&state, &fields.name, &fields.shortDescription, &fields.categoryID,
		&fields.pointName, &fields.brandColor, &fields.iconAsset, &fields.glyph)
	if err != nil {
		return "", nil, err
	}
	var activeMission bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_mission WHERE planet_id=$1 AND status='active')`, planetID).Scan(&activeMission); err != nil {
		return "", nil, err
	}
	return state, publishMissing(fields, activeMission), nil
}

func planetPublishCheck(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	ownerID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	id, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "publish check failed")
	}
	defer tx.Rollback()
	state, missing, err := readPublishState(ctx, tx, ownerID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		logger.Error("Jazebe publish check failed: %v", err)
		return "", status.Error(codes.Internal, "publish check failed")
	}
	return encode(map[string]any{"planetId": id, "status": state, "missing": missing})
}

func planetRequestPublish(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	ownerID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	id, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "publish request failed")
	}
	defer tx.Rollback()
	state, missing, err := readPublishState(ctx, tx, ownerID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet not found")
	}
	if err != nil {
		logger.Error("Jazebe publish request lookup failed: %v", err)
		return "", status.Error(codes.Internal, "publish request failed")
	}
	if state != "setup" {
		return "", status.Error(codes.FailedPrecondition, "planet is not in setup")
	}
	if len(missing) != 0 {
		return "", status.Error(codes.FailedPrecondition, "planet setup is incomplete: "+strings.Join(missing, ","))
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_planet SET status='pending_review' WHERE id=$1`, id); err != nil {
		logger.Error("Jazebe publish request failed: %v", err)
		return "", status.Error(codes.Internal, "publish request failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log
		(action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES ('planet.publish_request','brand_admin',$1,'planet',$2,$2,jsonb_build_object('from','setup','to','pending_review'))`,
		ownerID, id); err != nil {
		logger.Error("Jazebe publish request audit failed: %v", err)
		return "", status.Error(codes.Internal, "publish request failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "publish request failed")
	}
	return encode(map[string]any{"planetId": id, "status": "pending_review"})
}

func planetWithdrawPublish(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	ownerID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	id, err := requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	result, err := db.ExecContext(ctx, `UPDATE jazebe_planet p SET status='setup' FROM jazebe_active_brand_admin a
		WHERE p.id=$1 AND p.brand_id=a.brand_id AND a.user_id=$2 AND a.is_active
		AND a.role IN ('main','technical') AND p.status='pending_review'`, id, ownerID)
	if err != nil {
		logger.Error("Jazebe publish withdrawal failed: %v", err)
		return "", status.Error(codes.Internal, "publish withdrawal failed")
	}
	n, err := result.RowsAffected()
	if err != nil || n == 0 {
		return "", status.Error(codes.FailedPrecondition, "planet is not pending review or not owned by caller")
	}
	return encode(map[string]any{"planetId": id, "status": "setup"})
}
