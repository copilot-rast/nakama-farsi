package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Brand administrators can pause, withdraw to draft, or end an active
// mission. Source protects the last simple solo point mission from removal.
func missionSetStatus(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID  string `json:"planetId"`
		MissionID string `json:"missionId"`
		Status    string `json:"status"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	ref := missionRef{PlanetID: input.PlanetID, MissionID: input.MissionID}
	ref.PlanetID, err = requiredUUID(ref.PlanetID)
	if err != nil {
		return "", err
	}
	ref.MissionID, err = requiredUUID(ref.MissionID)
	if err != nil {
		return "", err
	}
	if input.Status != "active" && input.Status != "draft" && input.Status != "paused" && input.Status != "expired" {
		return "", status.Error(codes.InvalidArgument, "invalid mission status")
	}
	if input.Status == "active" {
		var kind string
		err = db.QueryRowContext(ctx, `SELECT functional_type FROM jazebe_mission WHERE id=$1 AND planet_id=$2`, ref.MissionID, ref.PlanetID).Scan(&kind)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.NotFound, "mission not found")
		}
		if err != nil {
			return "", status.Error(codes.Internal, "mission lookup failed")
		}
		switch kind {
		case "media_view":
			return mediaMissionActivate(ctx, logger, db, nk, `{"planetId":"`+ref.PlanetID+`","missionId":"`+ref.MissionID+`"}`)
		case "code_entry":
			return codeMissionActivate(ctx, logger, db, nk, `{"planetId":"`+ref.PlanetID+`","missionId":"`+ref.MissionID+`"}`)
		case "content_creation":
			return contentMissionActivate(ctx, logger, db, nk, `{"planetId":"`+ref.PlanetID+`","missionId":"`+ref.MissionID+`"}`)
		default:
			return "", status.Error(codes.FailedPrecondition, "mission action unavailable")
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "mission transition failed")
	}
	defer tx.Rollback()
	// Two administrators must not both pause the last two simple missions.
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(774103,hashtext($1))`, ref.PlanetID); err != nil {
		return "", status.Error(codes.Internal, "mission transition failed")
	}
	var current, target string
	var points, tickets int
	err = tx.QueryRowContext(ctx, `SELECT status,target_type,reward_points,reward_tickets FROM jazebe_mission
		WHERE id=$1 AND planet_id=$2 FOR UPDATE`, ref.MissionID, ref.PlanetID).Scan(&current, &target, &points, &tickets)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "mission not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "mission transition failed")
	}
	var allowed bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_planet p JOIN jazebe_brand_admin a ON a.brand_id=p.brand_id
		WHERE p.id=$1 AND a.user_id=$2 AND a.is_active)`, ref.PlanetID, actor).Scan(&allowed); err != nil {
		return "", status.Error(codes.Internal, "mission authorization failed")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	if current == "suspended" {
		return "", status.Error(codes.FailedPrecondition, "suspended mission is controlled by operator")
	}
	if current == "active" && target == "solo" && points > 0 && tickets == 0 {
		var other bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_mission WHERE planet_id=$1 AND id<>$2
			AND status='active' AND target_type='solo' AND reward_points>0 AND reward_tickets=0)`, ref.PlanetID, ref.MissionID).Scan(&other); err != nil {
			return "", status.Error(codes.Internal, "mission transition failed")
		}
		if !other {
			return "", status.Error(codes.FailedPrecondition, "last solo point mission must remain active")
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE jazebe_mission SET status=$2 WHERE id=$1`, ref.MissionID, input.Status); err != nil {
		return "", status.Error(codes.Internal, "mission transition failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log (action,actor_type,actor_id,target_type,target_id,planet_id,params)
		VALUES ('mission.status','brand_admin',$1,'mission',$2,$3,jsonb_build_object('from',$4::text,'to',$5::text))`, actor, ref.MissionID, ref.PlanetID, current, input.Status); err != nil {
		return "", status.Error(codes.Internal, "mission transition audit failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "mission transition failed")
	}
	return encode(map[string]any{"missionId": ref.MissionID, "status": input.Status})
}
