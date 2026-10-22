package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func medalLibraryAudit(ctx context.Context, tx *sql.Tx, action, actor, id string, version int, key string, changes any, reason string) error {
	params := map[string]any{"version": version, "key": key}
	if changes != nil {
		field := "changes"
		if action == "medal_library.create" {
			field = "snapshot"
		}
		params[field] = changes
	}
	if reason != "" {
		params["reason"] = reason
	}
	raw, _ := json.Marshal(params)
	_, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,params)
		VALUES($1,'operator',$2,'medal_library',$3,$4::jsonb)`, action, actor, id, string(raw))
	return err
}

func medalLibraryCreate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := medalOperator(ctx, true)
	if err != nil {
		return "", err
	}
	var in medalLibraryInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in, err = validateMedalLibraryInput(in, true)
	if err != nil {
		return "", err
	}
	if problems := medalSchemaProblems(in.ConditionType, in.ConditionParamSchema); len(problems) > 0 {
		return "", status.Error(codes.InvalidArgument, "invalid medal schema: "+strings.Join(problems, ", "))
	}
	id := uuid.Must(uuid.NewV4()).String()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "medal library unavailable")
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_medal_library(id,key,name,description,icon,theme_category,condition_type,condition_param_schema,created_by)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9)`, id, in.Key, in.Name, in.Description, in.Icon, in.ThemeCategory, in.ConditionType, string(in.ConditionParamSchema), actor)
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "medal key already used")
		}
		return "", status.Error(codes.Internal, "medal creation failed")
	}
	if err = medalLibraryAudit(ctx, tx, "medal_library.create", actor, id, 1, in.Key,
		map[string]any{"name": in.Name, "description": in.Description, "icon": in.Icon, "themeCategory": in.ThemeCategory, "conditionType": in.ConditionType, "conditionParamSchema": json.RawMessage(in.ConditionParamSchema), "isActive": true}, ""); err != nil {
		return "", status.Error(codes.Internal, "medal audit failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "medal creation failed")
	}
	row, err := medalLibraryGetRow(ctx, db, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func medalLibraryUpdate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := medalOperator(ctx, true)
	if err != nil {
		return "", err
	}
	var fields map[string]json.RawMessage
	if err = decode(payload, &fields); err != nil {
		return "", err
	}
	for key := range fields {
		switch key {
		case "id", "name", "description", "icon", "themeCategory", "conditionType", "conditionParamSchema":
		default:
			return "", status.Error(codes.InvalidArgument, "unknown medal field")
		}
	}
	var id string
	if err = json.Unmarshal(fields["id"], &id); err != nil {
		return "", status.Error(codes.InvalidArgument, "medal id required")
	}
	id, err = requiredUUID(id)
	if err != nil {
		return "", err
	}
	if _, ok := fields["key"]; ok {
		return "", status.Error(codes.InvalidArgument, "medal key cannot change")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "medal library unavailable")
	}
	defer tx.Rollback()
	old, err := scanMedalLibrary(tx.QueryRowContext(ctx, medalLibrarySelect+` WHERE l.id=$1 FOR UPDATE OF l`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "medal library entry not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "medal library unavailable")
	}
	next := old
	if raw, ok := fields["name"]; ok {
		if err = json.Unmarshal(raw, &next.Name); err != nil {
			return "", status.Error(codes.InvalidArgument, "invalid name")
		}
		next.Name = strings.TrimSpace(next.Name)
	}
	if raw, ok := fields["description"]; ok {
		if err = json.Unmarshal(raw, &next.Description); err != nil {
			return "", status.Error(codes.InvalidArgument, "invalid description")
		}
	}
	if raw, ok := fields["icon"]; ok {
		if err = json.Unmarshal(raw, &next.Icon); err != nil {
			return "", status.Error(codes.InvalidArgument, "invalid icon")
		}
	}
	if raw, ok := fields["themeCategory"]; ok {
		if err = json.Unmarshal(raw, &next.ThemeCategory); err != nil {
			return "", status.Error(codes.InvalidArgument, "invalid theme category")
		}
	}
	if raw, ok := fields["conditionType"]; ok {
		if err = json.Unmarshal(raw, &next.ConditionType); err != nil {
			return "", status.Error(codes.InvalidArgument, "invalid condition type")
		}
	}
	if raw, ok := fields["conditionParamSchema"]; ok {
		next.ConditionParamSchema = raw
	}
	validated := medalLibraryInput{Key: old.Key, Name: next.Name, Description: next.Description, Icon: next.Icon, ThemeCategory: next.ThemeCategory, ConditionType: next.ConditionType, ConditionParamSchema: next.ConditionParamSchema}
	validated, err = validateMedalLibraryInput(validated, true)
	if err != nil {
		return "", err
	}
	if problems := medalSchemaProblems(validated.ConditionType, validated.ConditionParamSchema); len(problems) > 0 {
		return "", status.Error(codes.InvalidArgument, "invalid medal schema: "+strings.Join(problems, ", "))
	}
	next.Name = validated.Name
	next.Description = validated.Description
	next.Icon = validated.Icon
	changes := map[string]any{}
	var oldSchema, newSchema any
	_ = json.Unmarshal(old.ConditionParamSchema, &oldSchema)
	_ = json.Unmarshal(next.ConditionParamSchema, &newSchema)
	for _, field := range []struct {
		name          string
		before, after any
	}{
		{"name", old.Name, next.Name}, {"description", old.Description, next.Description}, {"icon", old.Icon, next.Icon},
		{"themeCategory", old.ThemeCategory, next.ThemeCategory}, {"conditionType", old.ConditionType, next.ConditionType},
		{"conditionParamSchema", oldSchema, newSchema},
	} {
		if !reflect.DeepEqual(field.before, field.after) {
			changes[field.name] = map[string]any{"from": field.before, "to": field.after}
		}
	}
	if len(changes) == 0 {
		return encode(old)
	}
	_, err = tx.ExecContext(ctx, `UPDATE jazebe_medal_library SET name=$2,description=$3,icon=$4,theme_category=$5,condition_type=$6,
		condition_param_schema=$7::jsonb,version=version+1,updated_by=$8,updated_at=now() WHERE id=$1`, id, next.Name, next.Description, next.Icon, next.ThemeCategory, next.ConditionType, string(next.ConditionParamSchema), actor)
	if err != nil {
		return "", status.Error(codes.Internal, "medal update failed")
	}
	if err = medalLibraryAudit(ctx, tx, "medal_library.update", actor, id, old.Version+1, old.Key, changes, ""); err != nil {
		return "", status.Error(codes.Internal, "medal audit failed")
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "medal update failed")
	}
	row, err := medalLibraryGetRow(ctx, db, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func medalLibrarySetActive(ctx context.Context, db *sql.DB, payload string, active bool) (string, error) {
	actor, err := medalOperator(ctx, true)
	if err != nil {
		return "", err
	}
	var in medalLibraryInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.ID, err = requiredUUID(in.ID)
	if err != nil {
		return "", err
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if !active && (!utf8.ValidString(in.Reason) || utf8.RuneCountInString(in.Reason) < 3 || utf8.RuneCountInString(in.Reason) > 512) {
		return "", status.Error(codes.InvalidArgument, "deactivation reason required")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "medal library unavailable")
	}
	defer tx.Rollback()
	var key string
	var was bool
	var version int
	err = tx.QueryRowContext(ctx, `SELECT key,is_active,version FROM jazebe_medal_library WHERE id=$1 FOR UPDATE`, in.ID).Scan(&key, &was, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "medal library entry not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "medal library unavailable")
	}
	if was != active {
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_medal_library SET is_active=$2,version=version+1,updated_by=$3,updated_at=now() WHERE id=$1`, in.ID, active, actor)
		if err != nil {
			return "", status.Error(codes.Internal, "medal status update failed")
		}
		action := "medal_library.deactivate"
		if active {
			action = "medal_library.activate"
		}
		if err = medalLibraryAudit(ctx, tx, action, actor, in.ID, version+1, key, map[string]any{"isActive": map[string]any{"from": was, "to": active}}, in.Reason); err != nil {
			return "", status.Error(codes.Internal, "medal audit failed")
		}
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "medal status update failed")
	}
	row, err := medalLibraryGetRow(ctx, db, in.ID)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func medalLibraryActivate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return medalLibrarySetActive(ctx, db, payload, true)
}
func medalLibraryDeactivate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return medalLibrarySetActive(ctx, db, payload, false)
}

func medalLibraryHistory(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := medalOperator(ctx, false); err != nil {
		return "", err
	}
	var in medalLibraryInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	id, err := requiredUUID(in.ID)
	if err != nil {
		return "", err
	}
	if _, err = medalLibraryGetRow(ctx, db, id); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT id,action,actor_id,params,created_at FROM jazebe_audit_log
		WHERE target_type='medal_library' AND target_id=$1 ORDER BY created_at DESC,id DESC LIMIT 200`, id)
	if err != nil {
		return "", status.Error(codes.Internal, "medal history unavailable")
	}
	items := []map[string]any{}
	for rows.Next() {
		var seq int64
		var action string
		var actor sql.NullString
		var raw []byte
		var at any
		if err = rows.Scan(&seq, &action, &actor, &raw, &at); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "medal history unavailable")
		}
		var params struct {
			Version *int `json:"version"`
			Changes any  `json:"changes"`
		}
		if json.Unmarshal(raw, &params) != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "medal history unavailable")
		}
		var actorID any
		if actor.Valid {
			actorID = actor.String
		}
		items = append(items, map[string]any{"id": seq, "action": action, "version": params.Version, "actorId": actorID, "changes": params.Changes, "createdAt": at})
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "medal history unavailable")
	}
	return encode(items)
}
