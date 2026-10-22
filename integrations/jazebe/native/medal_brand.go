package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type medalBrandInput struct {
	PlanetID        string          `json:"planetId"`
	ID              string          `json:"id"`
	MedalLibraryID  string          `json:"medalLibraryId"`
	ConditionParams json.RawMessage `json:"conditionParams"`
	RewardPoints    *int            `json:"rewardPoints"`
	RewardTickets   *int            `json:"rewardTickets"`
	Status          string          `json:"status"`
}

func medalBrandAccess(ctx context.Context, db *sql.DB, in medalBrandInput) (string, string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", "", err
	}
	planet, err := requiredUUID(in.PlanetID)
	if err != nil {
		return "", "", err
	}
	brand, err := shopBrandAccess(ctx, db, planet, actor, true)
	return actor, brand, err
}

func medalBrandRow(ctx context.Context, db *sql.DB, planetID, medalID string) (map[string]any, error) {
	var libraryID, key, name, theme, kind, state string
	var description, icon sql.NullString
	var schema, params []byte
	var points, tickets, acquisitions int
	var active bool
	var created time.Time
	err := db.QueryRowContext(ctx, `SELECT pm.medal_library_id,l.key,l.name,l.description,l.icon,l.theme_category,
		l.condition_type,l.condition_param_schema,pm.condition_params,pm.reward_points,pm.reward_tickets,
		pm.status,l.is_active,pm.created_at,(SELECT count(*) FROM jazebe_medal_acquisition a WHERE a.planet_medal_id=pm.id)
		FROM jazebe_planet_medal pm JOIN jazebe_medal_library l ON l.id=pm.medal_library_id
		WHERE pm.planet_id=$1 AND pm.id=$2`, planetID, medalID).Scan(&libraryID, &key, &name, &description, &icon, &theme, &kind, &schema, &params, &points, &tickets, &state, &active, &created, &acquisitions)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "planet medal not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "planet medal unavailable")
	}
	return map[string]any{"id": medalID, "medalLibraryId": libraryID, "key": key, "name": name,
		"description": nullableString(description), "icon": nullableString(icon), "themeCategory": theme,
		"conditionType": kind, "conditionParamSchema": json.RawMessage(schema), "conditionParams": json.RawMessage(params),
		"rewardPoints": points, "rewardTickets": tickets, "status": state, "libraryActive": active,
		"acquisitions": acquisitions, "createdAt": created}, nil
}

func nullableString(value sql.NullString) any {
	if value.Valid {
		return value.String
	}
	return nil
}

func medalBrandList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	var in medalBrandInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if _, _, err := medalBrandAccess(ctx, db, in); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT id FROM jazebe_planet_medal WHERE planet_id=$1 ORDER BY created_at,id`, in.PlanetID)
	if err != nil {
		return "", status.Error(codes.Internal, "planet medals unavailable")
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return "", err
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", err
	}
	result := []map[string]any{}
	for _, id := range ids {
		row, rowErr := medalBrandRow(ctx, db, in.PlanetID, id)
		if rowErr != nil {
			return "", rowErr
		}
		result = append(result, row)
	}
	return encode(result)
}

func medalBrandMeta(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	var in medalBrandInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if _, _, err := medalBrandAccess(ctx, db, in); err != nil {
		return "", err
	}
	policy, err := loadPlatformPolicy(ctx, db)
	if err != nil {
		return "", err
	}
	var name string
	if err = db.QueryRowContext(ctx, `SELECT point_name FROM jazebe_planet WHERE id=$1`, in.PlanetID).Scan(&name); err != nil {
		return "", err
	}
	return encode(map[string]any{"pointName": name, "rewardPointsMax": policy.MissionRewardPointMax, "rewardTicketsMax": policy.MissionRewardTicketMax})
}

func medalBrandLibrary(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	var in medalBrandInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if _, _, err := medalBrandAccess(ctx, db, in); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT l.id,l.key,l.name,l.description,l.icon,l.theme_category,l.condition_type,
		l.condition_param_schema,l.version,EXISTS(SELECT 1 FROM jazebe_planet_medal pm WHERE pm.planet_id=$1 AND pm.medal_library_id=l.id)
		FROM jazebe_medal_library l WHERE l.is_active ORDER BY l.created_at,l.id`, in.PlanetID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	result := []map[string]any{}
	for rows.Next() {
		var id, key, name, theme, kind string
		var desc, icon sql.NullString
		var schema []byte
		var version int
		var selected bool
		if err = rows.Scan(&id, &key, &name, &desc, &icon, &theme, &kind, &schema, &version, &selected); err != nil {
			return "", err
		}
		result = append(result, map[string]any{"id": id, "key": key, "name": name, "description": nullableString(desc), "icon": nullableString(icon),
			"themeCategory": theme, "conditionType": kind, "conditionParamSchema": json.RawMessage(schema), "version": version, "selected": selected})
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	return encode(result)
}

func medalBrandValidateRewards(ctx context.Context, db *sql.DB, points, tickets int) error {
	policy, err := loadPlatformPolicy(ctx, db)
	if err != nil {
		return err
	}
	if points < 0 || points > policy.MissionRewardPointMax || tickets < 0 || tickets > policy.MissionRewardTicketMax || tickets > 100000 {
		return status.Error(codes.InvalidArgument, "medal reward outside platform policy")
	}
	return nil
}

func medalBrandAudit(ctx context.Context, tx *sql.Tx, actor, brand, planet, medal, action string) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,brand_id,planet_id)
		VALUES($1,'brand_admin',$2,'planet_medal',$3,$4,$5)`, action, actor, medal, brand, planet)
	return err
}

func medalBrandAdd(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	var in medalBrandInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	actor, brand, err := medalBrandAccess(ctx, db, in)
	if err != nil {
		return "", err
	}
	libraryID, err := requiredUUID(in.MedalLibraryID)
	if err != nil {
		return "", err
	}
	points, tickets := 0, 0
	if in.RewardPoints != nil {
		points = *in.RewardPoints
	}
	if in.RewardTickets != nil {
		tickets = *in.RewardTickets
	}
	if err = medalBrandValidateRewards(ctx, db, points, tickets); err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var schema []byte
	var active bool
	var kind string
	err = tx.QueryRowContext(ctx, `SELECT condition_param_schema,is_active,condition_type FROM jazebe_medal_library WHERE id=$1`, libraryID).Scan(&schema, &active, &kind)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "medal library entry not found")
	}
	if err != nil {
		return "", err
	}
	if !active {
		return "", status.Error(codes.FailedPrecondition, "medal library entry inactive")
	}
	if kind == "activity_count" {
		var rule map[string]json.RawMessage
		if json.Unmarshal(schema, &rule) != nil {
			return "", status.Error(codes.Internal, "medal schema invalid")
		}
		var activity string
		if json.Unmarshal(rule["activity"], &activity) != nil {
			return "", status.Error(codes.Internal, "medal activity invalid")
		}
	}
	params, issues, err := validateMedalParams(schema, in.ConditionParams)
	if err != nil || len(issues) > 0 {
		return "", status.Error(codes.InvalidArgument, "invalid medal condition parameters")
	}
	encoded, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	id := uuid.Must(uuid.NewV4()).String()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_planet_medal(id,planet_id,medal_library_id,condition_params,reward_points,reward_tickets)
		VALUES($1,$2,$3,$4::jsonb,$5,$6)`, id, in.PlanetID, libraryID, string(encoded), points, tickets)
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "medal already on planet")
		}
		return "", err
	}
	if err = queueMedalBackfill(ctx, tx, in.PlanetID, id); err != nil {
		return "", err
	}
	if err = medalBrandAudit(ctx, tx, actor, brand, in.PlanetID, id, "planet_medal.add"); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	row, err := medalBrandRow(ctx, db, in.PlanetID, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func medalBrandUpdate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	var in medalBrandInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	actor, brand, err := medalBrandAccess(ctx, db, in)
	if err != nil {
		return "", err
	}
	id, err := requiredUUID(in.ID)
	if err != nil {
		return "", err
	}
	if in.ConditionParams == nil && in.RewardPoints == nil && in.RewardTickets == nil {
		return "", status.Error(codes.InvalidArgument, "no medal changes")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var schema, params []byte
	var points, tickets int
	err = tx.QueryRowContext(ctx, `SELECT l.condition_param_schema,pm.condition_params,pm.reward_points,pm.reward_tickets
		FROM jazebe_planet_medal pm JOIN jazebe_medal_library l ON l.id=pm.medal_library_id
		WHERE pm.id=$1 AND pm.planet_id=$2 FOR UPDATE OF pm`, id, in.PlanetID).Scan(&schema, &params, &points, &tickets)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet medal not found")
	}
	if err != nil {
		return "", err
	}
	if in.ConditionParams != nil {
		clean, issues, e := validateMedalParams(schema, in.ConditionParams)
		if e != nil || len(issues) > 0 {
			return "", status.Error(codes.InvalidArgument, "invalid medal condition parameters")
		}
		params, e = json.Marshal(clean)
		if e != nil {
			return "", e
		}
	}
	if in.RewardPoints != nil {
		points = *in.RewardPoints
	}
	if in.RewardTickets != nil {
		tickets = *in.RewardTickets
	}
	if err = medalBrandValidateRewards(ctx, db, points, tickets); err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `UPDATE jazebe_planet_medal SET condition_params=$2::jsonb,reward_points=$3,reward_tickets=$4,updated_at=now() WHERE id=$1`, id, string(params), points, tickets)
	if err != nil {
		return "", err
	}
	if in.ConditionParams != nil {
		if err = queueMedalBackfill(ctx, tx, in.PlanetID, id); err != nil {
			return "", err
		}
	}
	if err = medalBrandAudit(ctx, tx, actor, brand, in.PlanetID, id, "planet_medal.update"); err != nil {
		return "", err
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	row, err := medalBrandRow(ctx, db, in.PlanetID, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func medalBrandStatus(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	var in medalBrandInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	actor, brand, err := medalBrandAccess(ctx, db, in)
	if err != nil {
		return "", err
	}
	id, err := requiredUUID(in.ID)
	if err != nil {
		return "", err
	}
	if in.Status != "active" && in.Status != "paused" {
		return "", status.Error(codes.InvalidArgument, "invalid medal status")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var old string
	err = tx.QueryRowContext(ctx, `SELECT status FROM jazebe_planet_medal WHERE id=$1 AND planet_id=$2 FOR UPDATE`, id, in.PlanetID).Scan(&old)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "planet medal not found")
	}
	if err != nil {
		return "", err
	}
	if old != in.Status {
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_planet_medal SET status=$2,updated_at=now() WHERE id=$1`, id, in.Status)
		if err != nil {
			return "", err
		}
		if in.Status == "active" {
			if err = queueMedalBackfill(ctx, tx, in.PlanetID, id); err != nil {
				return "", err
			}
		}
		if err = medalBrandAudit(ctx, tx, actor, brand, in.PlanetID, id, "planet_medal.status"); err != nil {
			return "", err
		}
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	row, err := medalBrandRow(ctx, db, in.PlanetID, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}
