package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Operator identity, directory and audit-log explorer (source
// operator-access.service.ts and audit-log-explorer.service.ts).
//
// Adaptation: the source keeps operator levels in platform_admin rows. In
// Nakama-Farsi the authority is the deployment allowlists
// JAZEBE_OPERATOR_USER_IDS (normal), JAZEBE_MANAGERIAL_OPERATOR_USER_IDS and
// JAZEBE_SUPER_OPERATOR_USER_IDS, read on every request by the checks; these
// RPCs report that same authority and do not change it.

var operatorCapabilities = map[string]string{
	"auditLog": "super", "largeAdjustment": "super", "kycEditVerified": "super", "dataPurge": "super",
	"medalLibrary": "super", "operatorLevels": "super", "operations": "super",
	"brands": "managerial", "planetLifecycle": "managerial", "catalogs": "managerial", "platformPolicy": "managerial",
	"broadcasts": "managerial", "groupDisband": "managerial", "kycDecision": "managerial",
	"moderation": "normal", "support": "normal",
}

var operatorRank = map[string]int{"normal": 1, "managerial": 2, "super": 3}

// operatorLevelOf returns the level the allowlists grant, or "" for a
// non-operator. Elevated levels require the base operator allowlist too, as
// every operator check does.
func operatorLevelOf(user string) string {
	if !isOperator(user, os.Getenv("JAZEBE_OPERATOR_USER_IDS")) {
		return ""
	}
	if isOperator(user, os.Getenv("JAZEBE_SUPER_OPERATOR_USER_IDS")) {
		return "super"
	}
	if isOperator(user, os.Getenv("JAZEBE_MANAGERIAL_OPERATOR_USER_IDS")) {
		return "managerial"
	}
	return "normal"
}

func operatorMe(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, _ string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	level := operatorLevelOf(user)
	if level == "" {
		return "", status.Error(codes.PermissionDenied, "platform operator required")
	}
	can := map[string]bool{}
	for capability, need := range operatorCapabilities {
		can[capability] = operatorRank[level] >= operatorRank[need]
	}
	return encode(map[string]any{"id": user, "userId": user, "level": level, "can": can, "levelSource": "deployment"})
}

func operatorDirectory(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, _ string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	if operatorLevelOf(user) != "super" {
		return "", status.Error(codes.PermissionDenied, "super operator required")
	}
	ids := map[string]bool{}
	for _, id := range strings.Split(os.Getenv("JAZEBE_OPERATOR_USER_IDS"), ",") {
		if id = strings.TrimSpace(id); id != "" {
			if _, err := requiredUUID(id); err == nil {
				ids[id] = true
			}
		}
	}
	names, err := displayNames(ctx, db, ids)
	if err != nil {
		return "", err
	}
	list := make([]string, 0, len(ids))
	for id := range ids {
		list = append(list, id)
	}
	sort.Strings(list)
	items := []map[string]any{}
	for _, id := range list {
		var name any
		if n, ok := names[id]; ok {
			name = n
		}
		items = append(items, map[string]any{"id": id, "userId": id, "level": operatorLevelOf(id), "isActive": true, "phone": nil, "name": name})
	}
	return encode(map[string]any{"items": items, "levelSource": "deployment"})
}

// operatorAuditSearch is the super-only audit explorer: actor, target, action
// (exact, dotted prefix, or "prefix." / "prefix*"), brand, result and time
// range, newest first.
func operatorAuditSearch(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	if operatorLevelOf(user) != "super" {
		return "", status.Error(codes.PermissionDenied, "super operator required")
	}
	var in struct {
		ActorType  string `json:"actorType"`
		ActorID    string `json:"actorId"`
		TargetType string `json:"targetType"`
		TargetID   string `json:"targetId"`
		Action     string `json:"action"`
		BrandID    string `json:"brandId"`
		PlanetID   string `json:"planetId"`
		Result     string `json:"result"`
		From       string `json:"from"`
		To         string `json:"to"`
		Page       int    `json:"page"`
		Limit      int    `json:"limit"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	switch in.ActorType {
	case "", "user", "brand_admin", "operator", "system":
	default:
		return "", status.Error(codes.InvalidArgument, "invalid actorType")
	}
	for _, id := range []*string{&in.ActorID, &in.TargetID, &in.BrandID, &in.PlanetID} {
		if *id != "" {
			if *id, err = requiredUUID(*id); err != nil {
				return "", err
			}
		}
	}
	if len(in.TargetType) > 64 || len(in.Action) > 64 || len(in.Result) > 16 {
		return "", status.Error(codes.InvalidArgument, "filter too long")
	}
	var from, to any
	for _, pair := range []struct {
		raw string
		out *any
	}{{in.From, &from}, {in.To, &to}} {
		if pair.raw != "" {
			at, err := time.Parse(time.RFC3339, pair.raw)
			if err != nil {
				return "", status.Error(codes.InvalidArgument, "invalid time range")
			}
			*pair.out = at
		}
	}
	if in.Page < 0 || in.Page > 100000 || in.Limit < 0 || in.Limit > 100 {
		return "", status.Error(codes.InvalidArgument, "invalid page")
	}
	if in.Page == 0 {
		in.Page = 1
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	action := strings.TrimSpace(in.Action)
	escape := strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`)
	exact, prefix := "", ""
	switch {
	case action == "":
	case strings.HasSuffix(action, ".") || strings.HasSuffix(action, "*"):
		prefix = escape.Replace(strings.TrimSuffix(action, "*")) + "%"
	default:
		exact, prefix = action, escape.Replace(action)+".%"
	}
	where := ` FROM jazebe_audit_log a WHERE ($1::text='' OR a.actor_type=$1::text) AND ($2::uuid IS NULL OR a.actor_id=$2::uuid)
		AND ($3::text='' OR a.target_type=$3::text) AND ($4::uuid IS NULL OR a.target_id=$4::uuid)
		AND ($5::text='' AND $6::text='' OR a.action=$5::text OR a.action LIKE $6::text)
		AND ($7::uuid IS NULL OR a.brand_id=$7::uuid) AND ($8::uuid IS NULL OR a.planet_id=$8::uuid)
		AND ($9::text='' OR a.result=$9::text)
		AND ($10::timestamptz IS NULL OR a.created_at>=$10::timestamptz) AND ($11::timestamptz IS NULL OR a.created_at<$11::timestamptz)`
	args := []any{in.ActorType, nullUUID(in.ActorID), in.TargetType, nullUUID(in.TargetID), exact, prefix,
		nullUUID(in.BrandID), nullUUID(in.PlanetID), in.Result, from, to}
	var total int
	if err := db.QueryRowContext(ctx, `SELECT count(*)`+where, args...).Scan(&total); err != nil {
		return "", status.Error(codes.Internal, "audit log unavailable")
	}
	rows, err := db.QueryContext(ctx, `SELECT a.id,a.created_at,a.action,a.actor_type,a.actor_id::text,
		COALESCE(NULLIF(trim(u.display_name),''),NULL),a.target_type,a.target_id::text,a.brand_id::text,b.name,a.planet_id::text,a.params,a.result`+
		strings.Replace(where, " FROM jazebe_audit_log a ", " FROM jazebe_audit_log a LEFT JOIN users u ON u.id=a.actor_id LEFT JOIN jazebe_brand b ON b.id=a.brand_id ", 1)+
		` ORDER BY a.created_at DESC, a.id DESC LIMIT $12 OFFSET $13`, append(args, in.Limit, (in.Page-1)*in.Limit)...)
	if err != nil {
		return "", status.Error(codes.Internal, "audit log unavailable")
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, act, actorType, targetType, result string
		var created time.Time
		var actor, actorName, target, brand, brandName, planet sql.NullString
		var params []byte
		if err := rows.Scan(&id, &created, &act, &actorType, &actor, &actorName, &targetType, &target, &brand, &brandName, &planet, &params, &result); err != nil {
			return "", status.Error(codes.Internal, "audit log unavailable")
		}
		var p any
		_ = json.Unmarshal(params, &p)
		items = append(items, map[string]any{"id": id, "createdAt": created, "action": act, "actorType": actorType, "actorId": nullable(actor),
			"actorName": nullable(actorName), "targetType": targetType, "targetId": nullable(target), "brandId": nullable(brand),
			"brand": nullable(brandName), "planetId": nullable(planet), "params": p, "result": result})
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "audit log unavailable")
	}
	return encode(map[string]any{"items": items, "total": total, "page": in.Page, "limit": in.Limit})
}
