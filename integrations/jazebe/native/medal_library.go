package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var medalKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

type medalLibraryInput struct {
	ID                   string          `json:"id"`
	Key                  string          `json:"key"`
	Name                 string          `json:"name"`
	Description          *string         `json:"description"`
	Icon                 *string         `json:"icon"`
	ThemeCategory        string          `json:"themeCategory"`
	ConditionType        string          `json:"conditionType"`
	ConditionParamSchema json.RawMessage `json:"conditionParamSchema"`
	Reason               string          `json:"reason"`
	Search               string          `json:"search"`
	Active               *bool           `json:"active"`
	Page                 int             `json:"page"`
	Limit                int             `json:"limit"`
}

type medalLibraryRow struct {
	ID                   string          `json:"id"`
	Key                  string          `json:"key"`
	Name                 string          `json:"name"`
	Description          *string         `json:"description"`
	Icon                 *string         `json:"icon"`
	ThemeCategory        string          `json:"themeCategory"`
	ConditionType        string          `json:"conditionType"`
	ConditionParamSchema json.RawMessage `json:"conditionParamSchema"`
	IsActive             bool            `json:"isActive"`
	Version              int             `json:"version"`
	CreatedAt            time.Time       `json:"createdAt"`
	UpdatedAt            time.Time       `json:"updatedAt"`
	Usage                struct {
		Planets      int `json:"planets"`
		Acquisitions int `json:"acquisitions"`
	} `json:"usage"`
}

func medalOperator(ctx context.Context, super bool) (string, error) {
	actor, err := missionOperator(ctx)
	if err != nil {
		return "", err
	}
	if super && !isOperator(actor, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "super operator role required")
	}
	return actor, nil
}

func medalSchemaProblems(kind string, raw json.RawMessage) []string {
	problems := []string{}
	var schema map[string]json.RawMessage
	if len(raw) == 0 || json.Unmarshal(raw, &schema) != nil || schema == nil {
		return []string{"conditionParamSchema: object required"}
	}
	var field string
	switch kind {
	case "activity_count":
		field = "count"
	case "duration":
		field = "days"
	case "special_event":
		field = "event"
	default:
		return []string{"conditionType"}
	}
	if kind == "special_event" {
		var v string
		if json.Unmarshal(schema[field], &v) != nil || v != "string" {
			problems = append(problems, "event: must be string")
		}
	} else {
		var bounds struct {
			Min int `json:"min"`
			Max int `json:"max"`
		}
		if json.Unmarshal(schema[field], &bounds) != nil || bounds.Min < 1 || bounds.Max < bounds.Min {
			problems = append(problems, field+": invalid integer range")
		}
	}
	if kind == "activity_count" {
		var activity string
		_ = json.Unmarshal(schema["activity"], &activity)
		switch activity {
		case "purchase", "mission", "game", "referral", "group_members":
		default:
			problems = append(problems, "activity: invalid")
		}
	}
	if kind == "duration" && len(schema["unit"]) != 0 {
		var unit string
		if json.Unmarshal(schema["unit"], &unit) != nil || unit != "day" {
			problems = append(problems, "unit: only day")
		}
	}
	return problems
}

func validateMedalLibraryInput(in medalLibraryInput, creation bool) (medalLibraryInput, error) {
	in.Key = strings.TrimSpace(in.Key)
	in.Name = strings.TrimSpace(in.Name)
	if creation && (len(in.Key) < 2 || len(in.Key) > 64 || !medalKeyPattern.MatchString(in.Key)) {
		return in, status.Error(codes.InvalidArgument, "invalid medal key")
	}
	if creation || in.Name != "" {
		if !utf8.ValidString(in.Name) || utf8.RuneCountInString(in.Name) < 1 || utf8.RuneCountInString(in.Name) > 128 {
			return in, status.Error(codes.InvalidArgument, "invalid medal name")
		}
	}
	if in.Description != nil {
		v := strings.TrimSpace(*in.Description)
		if !utf8.ValidString(v) || utf8.RuneCountInString(v) > 2000 {
			return in, status.Error(codes.InvalidArgument, "invalid description")
		}
		in.Description = &v
	}
	if in.Icon != nil {
		v := strings.TrimSpace(*in.Icon)
		if !utf8.ValidString(v) || utf8.RuneCountInString(v) > 64 {
			return in, status.Error(codes.InvalidArgument, "invalid icon")
		}
		in.Icon = &v
	}
	if creation || in.ThemeCategory != "" {
		switch in.ThemeCategory {
		case "participation", "purchase", "social", "progression":
		default:
			return in, status.Error(codes.InvalidArgument, "invalid theme category")
		}
	}
	if creation || in.ConditionType != "" {
		switch in.ConditionType {
		case "activity_count", "duration", "special_event":
		default:
			return in, status.Error(codes.InvalidArgument, "invalid condition type")
		}
	}
	return in, nil
}

const medalLibrarySelect = `SELECT l.id,l.key,l.name,l.description,l.icon,l.theme_category,l.condition_type,l.condition_param_schema,l.is_active,l.version,l.created_at,l.updated_at,
	(SELECT count(*)::int FROM jazebe_planet_medal pm WHERE pm.medal_library_id=l.id),
	(SELECT count(*)::int FROM jazebe_medal_acquisition a JOIN jazebe_planet_medal pm ON pm.id=a.planet_medal_id WHERE pm.medal_library_id=l.id)
	FROM jazebe_medal_library l`

func scanMedalLibrary(scanner interface{ Scan(...any) error }) (medalLibraryRow, error) {
	var row medalLibraryRow
	var desc, icon sql.NullString
	err := scanner.Scan(&row.ID, &row.Key, &row.Name, &desc, &icon, &row.ThemeCategory, &row.ConditionType, &row.ConditionParamSchema, &row.IsActive, &row.Version, &row.CreatedAt, &row.UpdatedAt, &row.Usage.Planets, &row.Usage.Acquisitions)
	if desc.Valid {
		row.Description = &desc.String
	}
	if icon.Valid {
		row.Icon = &icon.String
	}
	return row, err
}

func medalLibraryGetRow(ctx context.Context, db *sql.DB, id string) (medalLibraryRow, error) {
	row, err := scanMedalLibrary(db.QueryRowContext(ctx, medalLibrarySelect+` WHERE l.id=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return row, status.Error(codes.NotFound, "medal library entry not found")
	}
	if err != nil {
		return row, status.Error(codes.Internal, "medal library unavailable")
	}
	return row, nil
}

func medalLibraryGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
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
	row, err := medalLibraryGetRow(ctx, db, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func medalLibraryMeta(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, _ string) (string, error) {
	if _, err := medalOperator(ctx, false); err != nil {
		return "", err
	}
	return encode(map[string]any{"themeCategories": []string{"participation", "purchase", "social", "progression"}, "conditionTypes": []any{
		map[string]any{"type": "activity_count", "fields": []any{map[string]any{"name": "activity", "kind": "fixed", "options": []string{"purchase", "mission", "game", "referral", "group_members"}}, map[string]string{"name": "count", "kind": "range"}}},
		map[string]any{"type": "duration", "fields": []any{map[string]any{"name": "unit", "kind": "fixed", "options": []string{"day"}}, map[string]string{"name": "days", "kind": "range"}}},
		map[string]any{"type": "special_event", "fields": []any{map[string]string{"name": "event", "kind": "string"}}},
	}})
}

func medalLibraryList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := medalOperator(ctx, false); err != nil {
		return "", err
	}
	var in medalLibraryInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	if in.Page < 1 || in.Page > 1000000 || in.Limit < 1 || in.Limit > 100 || !utf8.ValidString(in.Search) || utf8.RuneCountInString(in.Search) > 128 {
		return "", status.Error(codes.InvalidArgument, "invalid medal library query")
	}
	search := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(strings.TrimSpace(in.Search))
	var total int
	err := db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_medal_library l WHERE ($1::boolean IS NULL OR l.is_active=$1) AND ($2::text IS NULL OR l.name ILIKE $2 ESCAPE E'\\' OR l.key ILIKE $2 ESCAPE E'\\')`, in.Active, optionalSearch(search)).Scan(&total)
	if err != nil {
		return "", status.Error(codes.Internal, "medal library unavailable")
	}
	rows, err := db.QueryContext(ctx, medalLibrarySelect+` WHERE ($1::boolean IS NULL OR l.is_active=$1) AND ($2::text IS NULL OR l.name ILIKE $2 ESCAPE E'\\' OR l.key ILIKE $2 ESCAPE E'\\') ORDER BY l.created_at,l.id LIMIT $3 OFFSET $4`, in.Active, optionalSearch(search), in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", status.Error(codes.Internal, "medal library unavailable")
	}
	items := []medalLibraryRow{}
	for rows.Next() {
		item, e := scanMedalLibrary(rows)
		if e != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "medal library unavailable")
		}
		items = append(items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "medal library unavailable")
	}
	return encode(map[string]any{"code": "SUCCESS", "data": items, "meta": map[string]int{"total": total, "page": in.Page, "limit": in.Limit, "pages": (total + in.Limit - 1) / in.Limit}})
}
