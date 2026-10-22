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

type shopAdminInput struct {
	ID       string `json:"id"`
	PlanetID string `json:"planetId"`
	Status   string `json:"status"`
	Search   string `json:"search"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
	Reason   string `json:"reason"`
}

func shopAdminRow(ctx context.Context, db *sql.DB, id string) (map[string]any, error) {
	var planetID, planetName, brandID, brandName string
	err := db.QueryRowContext(ctx, `SELECT p.id,p.name,b.id,b.name FROM jazebe_shop_product s JOIN jazebe_planet p ON p.id=s.planet_id JOIN jazebe_brand b ON b.id=p.brand_id WHERE s.id=$1`, id).Scan(&planetID, &planetName, &brandID, &brandName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.NotFound, "shop product not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "shop product unavailable")
	}
	row, err := shopBrandProductRow(ctx, db, planetID, id, brandID)
	if err != nil {
		return nil, err
	}
	row["planet"] = map[string]string{"id": planetID, "name": planetName}
	row["brand"] = map[string]string{"productId": brandID, "tradeName": brandName}
	return row, nil
}

func shopAdminGet(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	var in shopAdminInput
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	id, err := requiredUUID(in.ID)
	if err != nil {
		return "", err
	}
	row, err := shopAdminRow(ctx, db, id)
	if err != nil {
		return "", err
	}
	return encode(row)
}

func shopAdminList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := missionOperator(ctx); err != nil {
		return "", err
	}
	var in shopAdminInput
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
		return "", status.Error(codes.InvalidArgument, "invalid product search")
	}
	switch in.Status {
	case "", "active", "paused", "archived", "suspended":
	default:
		return "", status.Error(codes.InvalidArgument, "invalid product status")
	}
	if in.PlanetID != "" {
		var err error
		in.PlanetID, err = requiredUUID(in.PlanetID)
		if err != nil {
			return "", err
		}
	}
	search := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(strings.TrimSpace(in.Search))
	var total int
	err := db.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_shop_product WHERE ($1::uuid IS NULL OR planet_id=$1) AND ($2::text IS NULL OR status=$2) AND ($3::text IS NULL OR name ILIKE $3 ESCAPE E'\\')`, optionalText(in.PlanetID), optionalText(in.Status), optionalSearch(search)).Scan(&total)
	if err != nil {
		return "", status.Error(codes.Internal, "shop product count unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT id FROM jazebe_shop_product WHERE ($1::uuid IS NULL OR planet_id=$1) AND ($2::text IS NULL OR status=$2) AND ($3::text IS NULL OR name ILIKE $3 ESCAPE E'\\') ORDER BY created_at DESC,id DESC LIMIT $4 OFFSET $5`, optionalText(in.PlanetID), optionalText(in.Status), optionalSearch(search), in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return "", status.Error(codes.Internal, "shop products unavailable")
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "shop products unavailable")
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "shop products unavailable")
	}
	items := []map[string]any{}
	for _, id := range ids {
		row, err := shopAdminRow(ctx, db, id)
		if err != nil {
			return "", err
		}
		items = append(items, row)
	}
	return encode(map[string]any{"code": "SUCCESS", "data": items, "meta": map[string]int{"total": total, "page": in.Page, "limit": in.Limit, "pages": (total + in.Limit - 1) / in.Limit}})
}

func shopAdminReason(reason string, required bool) (string, error) {
	reason = strings.TrimSpace(reason)
	if !utf8.ValidString(reason) || utf8.RuneCountInString(reason) > 512 || required && utf8.RuneCountInString(reason) < 3 {
		return "", status.Error(codes.InvalidArgument, "invalid moderation reason")
	}
	return reason, nil
}

func shopAdminSuspend(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return shopAdminTransition(ctx, db, payload, true)
}
func shopAdminUnsuspend(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	return shopAdminTransition(ctx, db, payload, false)
}

func shopAdminTransition(ctx context.Context, db *sql.DB, payload string, suspend bool) (string, error) {
	actor, err := missionOperator(ctx)
	if err != nil {
		return "", err
	}
	var in shopAdminInput
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
		return "", status.Error(codes.Internal, "shop moderation unavailable")
	}
	defer tx.Rollback()
	var planetID, brandID, state, name string
	err = tx.QueryRowContext(ctx, `SELECT s.planet_id,p.brand_id,s.status,s.name FROM jazebe_shop_product s JOIN jazebe_planet p ON p.id=s.planet_id WHERE s.id=$1 FOR UPDATE OF s`, in.ID).Scan(&planetID, &brandID, &state, &name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "shop product not found")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "shop moderation unavailable")
	}
	next, action, title, body := "paused", "shop.product.unsuspend", "رفع تعلیق کالا", "تعلیق کالای «"+name+"» برداشته شد؛ کالا در حالت «مخفی» است و می‌توانی دوباره فعالش کنی."
	params := map[string]any{"reason": nullableText(in.Reason), "name": name}
	if suspend {
		if state == "suspended" {
			return "", status.Error(codes.FailedPrecondition, "product already suspended")
		}
		next = "suspended"
		action = "shop.product.suspend"
		title = "تعلیق کالا توسط اپراتور"
		body = "کالای «" + name + "» توسط اپراتور جاذبه از فروشگاه تعلیق شد. دلیل: " + in.Reason
		params["previousStatus"] = state
	} else if state != "suspended" {
		return "", status.Error(codes.FailedPrecondition, "product is not suspended")
	}
	if _, err = tx.ExecContext(ctx, `UPDATE jazebe_shop_product SET status=$2,updated_at=now() WHERE id=$1`, in.ID, next); err != nil {
		return "", status.Error(codes.Internal, "product moderation failed")
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", status.Error(codes.Internal, "moderation audit failed")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO jazebe_audit_log(action,actor_type,actor_id,target_type,target_id,planet_id,params) VALUES($1,'operator',$2,'shop_product',$3,$4,$5::jsonb)`, action, actor, in.ID, planetID, string(paramsJSON)); err != nil {
		return "", status.Error(codes.Internal, "moderation audit failed")
	}
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT user_id FROM jazebe_brand_admin WHERE brand_id=$1 AND is_active`, brandID)
	if err != nil {
		return "", status.Error(codes.Internal, "brand administrators unavailable")
	}
	users := []string{}
	for rows.Next() {
		var user string
		if err = rows.Scan(&user); err != nil {
			rows.Close()
			return "", status.Error(codes.Internal, "brand administrators unavailable")
		}
		users = append(users, user)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", status.Error(codes.Internal, "brand administrators unavailable")
	}
	event := uuid.Must(uuid.NewV4()).String()
	now := time.Now().UTC()
	for _, user := range users {
		if err = persistJazebeNotification(ctx, tx, user, "shop-moderation:"+event, in.ID, map[string]any{"title": title, "body": body, "category": "system", "priority": "urgent", "subjectType": "shop_product", "subjectId": in.ID}, now); err != nil {
			return "", status.Error(codes.Internal, "brand notification failed")
		}
	}
	if err = tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "product moderation failed")
	}
	row, err := shopAdminRow(ctx, db, in.ID)
	if err != nil {
		return "", err
	}
	return encode(row)
}
