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
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var gameTemplateKeyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)

type gameOperatorInput struct {
	ID            string  `json:"id"`
	Key           string  `json:"key"`
	Name          *string `json:"name"`
	Engine        *string `json:"engine"`
	Description   *string `json:"description"`
	Icon          *string `json:"icon"`
	DurationSec   *int    `json:"durationSec"`
	IsActive      *bool   `json:"isActive"`
	AvailableOnly bool    `json:"availableOnly"`
	PlanetID      string  `json:"planetId"`
	Status        string  `json:"status"`
	Reason        string  `json:"reason"`
	Search        string  `json:"search"`
	Page          int     `json:"page"`
	Limit         int     `json:"limit"`
}

func gameManager(ctx context.Context) (string, error) {
	actor, err := missionOperator(ctx)
	if err != nil {
		return "", err
	}
	if !isOperator(actor, os.Getenv("JAZEBE_MANAGERIAL_OPERATOR_USER_IDS")) && !isOperator(actor, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) {
		return "", status.Error(codes.PermissionDenied, "managerial operator role required")
	}
	return actor, nil
}

func gameOperatorPage(in *gameOperatorInput) error {
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	if in.Page < 1 || in.Page > 100000 || in.Limit < 1 || in.Limit > 100 || !utf8.ValidString(in.Search) || utf8.RuneCountInString(in.Search) > 128 {
		return status.Error(codes.InvalidArgument, "invalid game pagination")
	}
	return nil
}

const gameTemplateSelect = `SELECT jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'engine',t.engine,
	'description',t.description,'icon',t.icon,'defaultConfig',t.default_config,
	'isActive',t.is_active,'suspendedAt',t.suspended_at,'suspendReason',t.suspend_reason,
	'createdAt',t.created_at,'updatedAt',t.updated_at,
	'usage',jsonb_build_object('planets',(SELECT count(*) FROM jazebe_game g WHERE g.template_id=t.id),
	'active',(SELECT count(*) FROM jazebe_game g WHERE g.template_id=t.id AND g.status='active'))) FROM jazebe_game_template t`

func gameTemplateRow(ctx context.Context, db *sql.DB, id string) (map[string]any, error) {
	var raw []byte
	err := db.QueryRowContext(ctx, gameTemplateSelect+` WHERE t.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "game template not found")
	}
	if err != nil {
		return nil, err
	}
	var row map[string]any
	err = json.Unmarshal(raw, &row)
	return row, err
}

func gameTemplateGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	var in gameOperatorInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	id, err := requiredUUID(in.ID)
	if err != nil {
		return "", err
	}
	row, err := gameTemplateRow(ctx, db, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func gameTemplateList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	var in gameOperatorInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if err := gameOperatorPage(&in); err != nil {
		return "", err
	}
	search := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(strings.TrimSpace(in.Search))
	var total int
	err := db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_game_template t WHERE ($1::text IS NULL OR t.name ILIKE $1 ESCAPE E'\\') AND (NOT $2 OR t.is_active)`, optionalSearch(search), in.AvailableOnly).Scan(&total)
	if err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, gameTemplateSelect+` WHERE ($1::text IS NULL OR t.name ILIKE $1 ESCAPE E'\\') AND (NOT $2 OR t.is_active) ORDER BY t.created_at DESC,t.id DESC LIMIT $3 OFFSET $4`, optionalSearch(search), in.AvailableOnly, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", err
	}
	items := []map[string]any{}
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			break
		}
		var item map[string]any
		err = json.Unmarshal(raw, &item)
		if err != nil {
			break
		}
		items = append(items, item)
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		return "", err
	}
	return encode(map[string]any{"code": "SUCCESS", "data": items, "meta": map[string]int{"total": total, "page": in.Page, "limit": in.Limit, "pages": (total + in.Limit - 1) / in.Limit}})
}

func validateGameTemplate(in gameOperatorInput, create bool) error {
	if create && !gameTemplateKeyPattern.MatchString(strings.TrimSpace(in.Key)) {
		return status.Error(codes.InvalidArgument, "invalid template key")
	}
	for _, field := range []struct {
		value    *string
		min, max int
		name     string
	}{{in.Name, 2, 128, "name"}, {in.Engine, 2, 32, "engine"}, {in.Description, 0, 4000, "description"}, {in.Icon, 0, 64, "icon"}} {
		if field.value == nil {
			if create && (field.name == "name" || field.name == "engine") {
				return status.Errorf(codes.InvalidArgument, "%s required", field.name)
			}
			continue
		}
		v := strings.TrimSpace(*field.value)
		if !utf8.ValidString(v) || utf8.RuneCountInString(v) < field.min || utf8.RuneCountInString(v) > field.max {
			return status.Errorf(codes.InvalidArgument, "invalid %s", field.name)
		}
		*field.value = v
	}
	if in.DurationSec != nil && (*in.DurationSec < 3 || *in.DurationSec > 600) {
		return status.Error(codes.InvalidArgument, "invalid duration")
	}
	return nil
}

func gameTemplateCreate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := gameManager(ctx); err != nil {
		return "", err
	}
	var in gameOperatorInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if err := validateGameTemplate(in, true); err != nil {
		return "", err
	}
	id := uuid.Must(uuid.NewV4()).String()
	duration := map[string]any{}
	if in.DurationSec != nil {
		duration["durationSec"] = *in.DurationSec
	}
	config, _ := json.Marshal(duration)
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	_, err := db.ExecContext(ctx, `INSERT INTO jazebe_game_template(id,key,name,engine,description,icon,default_config,is_active) VALUES($1,$2,$3,$4,$5,$6,$7::jsonb,$8)`, id, strings.TrimSpace(in.Key), *in.Name, *in.Engine, nullableGameText(in.Description), nullableGameText(in.Icon), string(config), active)
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			return "", status.Error(codes.AlreadyExists, "game template key already exists")
		}
		return "", err
	}
	row, err := gameTemplateRow(ctx, db, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func gameTemplateUpdate(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := gameManager(ctx); err != nil {
		return "", err
	}
	var in gameOperatorInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	id, err := requiredUUID(in.ID)
	if err != nil {
		return "", err
	}
	if err = validateGameTemplate(in, false); err != nil {
		return "", err
	}
	var supplied map[string]json.RawMessage
	if err = json.Unmarshal([]byte(payload), &supplied); err != nil {
		return "", status.Error(codes.InvalidArgument, "invalid template update")
	}
	_, descriptionSet := supplied["description"]
	_, iconSet := supplied["icon"]
	res, err := db.ExecContext(ctx, `UPDATE jazebe_game_template SET name=COALESCE($2,name),engine=COALESCE($3,engine),
		description=CASE WHEN $4::boolean THEN $5 ELSE description END,icon=CASE WHEN $6::boolean THEN $7 ELSE icon END,
		default_config=CASE WHEN $8::integer IS NULL THEN default_config ELSE jsonb_set(default_config,'{durationSec}',to_jsonb($8::integer)) END,
		is_active=COALESCE($9,is_active),updated_at=now() WHERE id=$1`, id, nullableGameText(in.Name), nullableGameText(in.Engine), descriptionSet, nullableGameText(in.Description), iconSet, nullableGameText(in.Icon), in.DurationSec, in.IsActive)
	if err != nil {
		return "", err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return "", status.Error(codes.NotFound, "game template not found")
	}
	row, err := gameTemplateRow(ctx, db, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func nullableGameText(v *string) any {
	if v == nil {
		return nil
	}
	return *v
}

func gameAdminRow(ctx context.Context, db *sql.DB, id string) (map[string]any, error) {
	var planetID, planetName, brandID, brandName string
	err := db.QueryRowContext(ctx, `SELECT p.id,p.name,b.id,b.name FROM jazebe_game g JOIN jazebe_planet p ON p.id=g.planet_id JOIN jazebe_brand b ON b.id=p.brand_id WHERE g.id=$1`, id).Scan(&planetID, &planetName, &brandID, &brandName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "game not found")
	}
	if err != nil {
		return nil, err
	}
	row, err := brandGameRow(ctx, db, planetID, id)
	if err != nil {
		return nil, err
	}
	row["planet"] = map[string]string{"id": planetID, "name": planetName}
	row["brand"] = map[string]string{"productId": brandID, "tradeName": brandName}
	return row, nil
}

func gameAdminGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	var in gameOperatorInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	id, err := requiredUUID(in.ID)
	if err != nil {
		return "", err
	}
	row, err := gameAdminRow(ctx, db, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func gameAdminList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	var in gameOperatorInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	if err := gameOperatorPage(&in); err != nil {
		return "", err
	}
	if in.PlanetID != "" {
		var err error
		in.PlanetID, err = requiredUUID(in.PlanetID)
		if err != nil {
			return "", err
		}
	}
	if in.Status != "" && in.Status != "active" && in.Status != "paused" && in.Status != "suspended" {
		return "", status.Error(codes.InvalidArgument, "invalid game status")
	}
	search := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(strings.TrimSpace(in.Search))
	filter := ` FROM jazebe_game g WHERE ($1::uuid IS NULL OR g.planet_id=$1) AND ($2::text IS NULL OR g.status=$2) AND ($3::text IS NULL OR g.name ILIKE $3 ESCAPE E'\\')`
	args := []any{optionalText(in.PlanetID), optionalText(in.Status), optionalSearch(search)}
	var total int
	if err := db.QueryRowContext(ctx, `SELECT count(*)`+filter, args...).Scan(&total); err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT g.id`+filter+` ORDER BY g.created_at DESC,g.id DESC LIMIT $4 OFFSET $5`, append(args, in.Limit, (in.Page-1)*in.Limit)...)
	if err != nil {
		return "", err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			break
		}
		ids = append(ids, id)
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		return "", err
	}
	items := []map[string]any{}
	for _, id := range ids {
		row, e := gameAdminRow(ctx, db, id)
		if e != nil {
			return "", e
		}
		items = append(items, row)
	}
	return encode(map[string]any{"code": "SUCCESS", "data": items, "meta": map[string]int{"total": total, "page": in.Page, "limit": in.Limit, "pages": (total + in.Limit - 1) / in.Limit}})
}

func gameTemplateSuspend(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return gameOperatorTransition(ctx, db, payload, true, true)
}
func gameTemplateUnsuspend(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return gameOperatorTransition(ctx, db, payload, true, false)
}
func gameAdminSuspend(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return gameOperatorTransition(ctx, db, payload, false, true)
}
func gameAdminUnsuspend(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return gameOperatorTransition(ctx, db, payload, false, false)
}

func gameOperatorTransition(ctx context.Context, db *sql.DB, payload string, template, suspend bool) (string, error) {
	actor, err := missionOperator(ctx)
	if err != nil {
		return "", err
	}
	if template {
		actor, err = gameManager(ctx)
		if err != nil {
			return "", err
		}
	}
	var in gameOperatorInput
	if err = decode(payload, &in); err != nil {
		return "", err
	}
	in.ID, err = requiredUUID(in.ID)
	if err != nil {
		return "", err
	}
	in.Reason, err = shopAdminReason(in.Reason, suspend)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	var action, target, name, planetID, brandID, old string
	if template {
		var suspended sql.NullTime
		err = tx.QueryRowContext(ctx, `SELECT name,suspended_at FROM jazebe_game_template WHERE id=$1 FOR UPDATE`, in.ID).Scan(&name, &suspended)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.NotFound, "game template not found")
		}
		if err != nil {
			return "", err
		}
		if suspended.Valid == suspend {
			return "", status.Error(codes.FailedPrecondition, "invalid template suspension state")
		}
		action, target = "game_template.unsuspend", "game_template"
		if suspend {
			action = "game_template.suspend"
		}
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_game_template SET suspended_at=CASE WHEN $2 THEN now() ELSE NULL END,suspend_reason=CASE WHEN $2 THEN $3 ELSE NULL END,updated_at=now() WHERE id=$1`, in.ID, suspend, in.Reason)
	} else {
		err = tx.QueryRowContext(ctx, `SELECT g.name,g.status,p.id,p.brand_id FROM jazebe_game g JOIN jazebe_planet p ON p.id=g.planet_id WHERE g.id=$1 FOR UPDATE OF g`, in.ID).Scan(&name, &old, &planetID, &brandID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.NotFound, "game not found")
		}
		if err != nil {
			return "", err
		}
		if (old == "suspended") == suspend {
			return "", status.Error(codes.FailedPrecondition, "invalid game suspension state")
		}
		action, target = "game.unsuspend", "game"
		next := "paused"
		if suspend {
			action = "game.suspend"
			next = "suspended"
		}
		_, err = tx.ExecContext(ctx, `UPDATE jazebe_game SET status=$2,updated_at=now() WHERE id=$1`, in.ID, next)
	}
	if err != nil {
		return "", err
	}
	params, _ := json.Marshal(map[string]any{"reason": nullableText(in.Reason), "previousStatus": old, "name": name})
	_, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params) VALUES($1,'operator',$2,$3,$4,$5::uuid,$6::jsonb)`, action, actor, target, in.ID, optionalText(planetID), string(params))
	if err != nil {
		return "", err
	}
	query := `SELECT DISTINCT a.user_id FROM jazebe_brand_admin a WHERE a.is_active AND a.brand_id=$1`
	args := []any{brandID}
	if template {
		query = `SELECT DISTINCT a.user_id FROM jazebe_brand_admin a JOIN jazebe_planet p ON p.brand_id=a.brand_id JOIN jazebe_game g ON g.planet_id=p.id WHERE a.is_active AND g.template_id=$1`
		args = []any{in.ID}
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return "", err
	}
	users := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			break
		}
		users = append(users, id)
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		return "", err
	}
	event := uuid.Must(uuid.NewV4()).String()
	title := "Game suspension lifted"
	body := "Game " + name + " is paused and may be activated by the brand."
	if suspend {
		title = "Game suspended by operator"
		body = "Game " + name + " was suspended. Reason: " + in.Reason
	}
	for _, user := range users {
		if err = persistJazebeNotification(ctx, tx, user, "game-moderation:"+event, in.ID, map[string]any{"title": title, "body": body, "category": "system", "priority": "urgent", "subjectType": target, "subjectId": in.ID}, time.Now().UTC()); err != nil {
			return "", err
		}
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	if template {
		row, e := gameTemplateRow(ctx, db, in.ID)
		if e != nil {
			return "", e
		}
		return encode(row)
	}
	row, e := gameAdminRow(ctx, db, in.ID)
	if e != nil {
		return "", e
	}
	return encode(row)
}
