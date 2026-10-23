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

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Source: planet-admin.controller.ts / planet-admin.service.ts category CRUD
// and planet-brand.controller.ts "Active categories for the planet form".
var planetCategoryKey = regexp.MustCompile(`^[a-z0-9_]{2,64}$`)

type planetCategory struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	TitleFa   string    `json:"titleFa"`
	TitleEn   *string   `json:"titleEn"`
	SortOrder int       `json:"sortOrder"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

const planetCategoryColumns = `id,key,title_fa,title_en,sort_order,is_active,created_at,updated_at`

func scanPlanetCategory(row interface{ Scan(...any) error }) (planetCategory, error) {
	var c planetCategory
	var titleEn sql.NullString
	err := row.Scan(&c.ID, &c.Key, &c.TitleFa, &titleEn, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if titleEn.Valid {
		c.TitleEn = &titleEn.String
	}
	return c, err
}

func managerialOperator(ctx context.Context) (string, error) {
	actor, err := missionOperator(ctx)
	if err != nil {
		return "", err
	}
	if !isOperator(actor, os.Getenv("JAZEBE_MANAGERIAL_OPERATOR_USER_IDS")) && !isOperator(actor, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "managerial operator role required")
	}
	return actor, nil
}

func validCategoryTitle(title string) bool {
	return utf8.ValidString(title) && utf8.RuneCountInString(title) >= 1 && utf8.RuneCountInString(title) <= 128
}

// brandPlanetCategories lists active categories for any active brand administrator.
func brandPlanetCategories(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, _ string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var admin bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM jazebe_brand_admin a JOIN jazebe_brand b ON b.id=a.brand_id
		WHERE a.user_id=$1 AND a.is_active AND b.is_active)`, user).Scan(&admin); err != nil {
		return "", status.Error(codes.Internal, "brand authorization unavailable")
	}
	if !admin {
		return "", status.Error(codes.PermissionDenied, "active brand administrator required")
	}
	rows, err := db.QueryContext(ctx, `SELECT `+planetCategoryColumns+` FROM jazebe_planet_category
		WHERE is_active AND deleted_at IS NULL ORDER BY sort_order,title_fa`)
	if err != nil {
		return "", status.Error(codes.Internal, "categories unavailable")
	}
	defer rows.Close()
	items := []planetCategory{}
	for rows.Next() {
		c, err := scanPlanetCategory(rows)
		if err != nil {
			return "", status.Error(codes.Internal, "categories unavailable")
		}
		items = append(items, c)
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "categories unavailable")
	}
	return encode(map[string]any{"items": items})
}

func operatorPlanetCategoryList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	var in struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.Page < 0 || in.Page > 100000 || in.Limit < 0 || in.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid category page")
	}
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	var total int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_planet_category WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return "", status.Error(codes.Internal, "categories unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT `+planetCategoryColumns+` FROM jazebe_planet_category WHERE deleted_at IS NULL
		ORDER BY sort_order,title_fa LIMIT $1 OFFSET $2`, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", status.Error(codes.Internal, "categories unavailable")
	}
	defer rows.Close()
	items := []planetCategory{}
	for rows.Next() {
		c, err := scanPlanetCategory(rows)
		if err != nil {
			return "", status.Error(codes.Internal, "categories unavailable")
		}
		items = append(items, c)
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "categories unavailable")
	}
	return encode(map[string]any{"items": items, "total": total, "page": in.Page, "limit": in.Limit})
}

func operatorPlanetCategoryCreate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := managerialOperator(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		Key       string  `json:"key"`
		TitleFa   string  `json:"titleFa"`
		TitleEn   *string `json:"titleEn"`
		SortOrder int     `json:"sortOrder"`
		IsActive  *bool   `json:"isActive"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	in.Key = strings.TrimSpace(in.Key)
	in.TitleFa = strings.TrimSpace(in.TitleFa)
	if !planetCategoryKey.MatchString(in.Key) {
		return "", status.Error(codes.InvalidArgument, "key must be snake_case")
	}
	if !validCategoryTitle(in.TitleFa) {
		return "", status.Error(codes.InvalidArgument, "invalid Persian title")
	}
	var titleEn any
	if in.TitleEn != nil {
		trimmed := strings.TrimSpace(*in.TitleEn)
		if !utf8.ValidString(trimmed) || utf8.RuneCountInString(trimmed) > 128 {
			return "", status.Error(codes.InvalidArgument, "invalid English title")
		}
		titleEn = trimmed
	}
	active := in.IsActive == nil || *in.IsActive
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "category creation unavailable")
	}
	defer tx.Rollback()
	c, err := scanPlanetCategory(tx.QueryRowContext(ctx, `INSERT INTO jazebe_planet_category(id,key,title_fa,title_en,sort_order,is_active,created_by)
		VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING `+planetCategoryColumns,
		uuid.Must(uuid.NewV4()).String(), in.Key, in.TitleFa, titleEn, in.SortOrder, active, actor))
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == pgerrcode.UniqueViolation {
			return "", status.Error(codes.AlreadyExists, "category key already exists")
		}
		return "", status.Error(codes.Internal, "category creation failed")
	}
	if err := planetCategoryAudit(ctx, tx, "planet_category.create", actor, c.ID, map[string]any{"key": c.Key}); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "category creation failed")
	}
	return encode(c)
}

func operatorPlanetCategoryUpdate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := managerialOperator(ctx)
	if err != nil {
		return "", err
	}
	var in map[string]json.RawMessage
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	var id string
	if json.Unmarshal(in["id"], &id) != nil {
		return "", status.Error(codes.InvalidArgument, "category id required")
	}
	if id, err = requiredUUID(id); err != nil {
		return "", err
	}
	delete(in, "id")
	sets := []string{}
	args := []any{id}
	changed := []string{}
	for key, raw := range in {
		var value any
		switch key {
		case "titleFa":
			var s string
			if json.Unmarshal(raw, &s) != nil || !validCategoryTitle(strings.TrimSpace(s)) {
				return "", status.Error(codes.InvalidArgument, "invalid Persian title")
			}
			value = strings.TrimSpace(s)
		case "titleEn":
			var s *string
			if json.Unmarshal(raw, &s) != nil {
				return "", status.Error(codes.InvalidArgument, "invalid English title")
			}
			if s != nil {
				trimmed := strings.TrimSpace(*s)
				if !utf8.ValidString(trimmed) || utf8.RuneCountInString(trimmed) > 128 {
					return "", status.Error(codes.InvalidArgument, "invalid English title")
				}
				value = trimmed
			}
		case "sortOrder":
			var n int
			if json.Unmarshal(raw, &n) != nil {
				return "", status.Error(codes.InvalidArgument, "invalid sort order")
			}
			value = n
		case "isActive":
			var b bool
			if json.Unmarshal(raw, &b) != nil {
				return "", status.Error(codes.InvalidArgument, "invalid active flag")
			}
			value = b
		default:
			return "", status.Error(codes.InvalidArgument, "unsupported category field: "+key)
		}
		column := map[string]string{"titleFa": "title_fa", "titleEn": "title_en", "sortOrder": "sort_order", "isActive": "is_active"}[key]
		args = append(args, value)
		sets = append(sets, column+"=$"+stringInt(len(args)))
		changed = append(changed, key)
	}
	args = append(args, actor)
	sets = append(sets, "updated_by=$"+stringInt(len(args)), "updated_at=now()")
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "category update unavailable")
	}
	defer tx.Rollback()
	// Column names come from the fixed map above, never from the payload.
	c, err := scanPlanetCategory(tx.QueryRowContext(ctx, `UPDATE jazebe_planet_category SET `+strings.Join(sets, ",")+
		` WHERE id=$1 AND deleted_at IS NULL RETURNING `+planetCategoryColumns, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "category not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "category update failed")
	}
	if err := planetCategoryAudit(ctx, tx, "planet_category.update", actor, c.ID, map[string]any{"key": c.Key, "fields": changed}); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "category update failed")
	}
	return encode(c)
}

// operatorPlanetCategoryDelete soft-deletes like the source: the row stays so
// planets that reference it keep a valid foreign key.
func operatorPlanetCategoryDelete(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := managerialOperator(ctx)
	if err != nil {
		return "", err
	}
	var in struct {
		ID     string `json:"id"`
		Reason string `json:"reason"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if in.ID, err = requiredUUID(in.ID); err != nil {
		return "", err
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if !utf8.ValidString(in.Reason) || utf8.RuneCountInString(in.Reason) < 3 || utf8.RuneCountInString(in.Reason) > 512 {
		return "", status.Error(codes.InvalidArgument, "reason must be 3 to 512 characters")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "category delete unavailable")
	}
	defer tx.Rollback()
	c, err := scanPlanetCategory(tx.QueryRowContext(ctx, `UPDATE jazebe_planet_category SET is_active=false,deleted_at=now(),
		deleted_by=$2,updated_at=now() WHERE id=$1 AND deleted_at IS NULL RETURNING `+planetCategoryColumns, in.ID, actor))
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "category not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "category delete failed")
	}
	if err := planetCategoryAudit(ctx, tx, "planet_category.delete", actor, c.ID, map[string]any{"key": c.Key, "reason": in.Reason}); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "category delete failed")
	}
	return encode(c)
}

func planetCategoryAudit(ctx context.Context, tx *sql.Tx, action, actor, id string, params map[string]any) error {
	raw, err := json.Marshal(params)
	if err != nil {
		return status.Error(codes.Internal, "category audit failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,params)
		VALUES($1,'operator',$2,'planet_category',$3,$4::jsonb)`, action, actor, id, string(raw)); err != nil {
		return status.Error(codes.Internal, "category audit failed")
	}
	return nil
}
