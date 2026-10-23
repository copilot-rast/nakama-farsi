package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func brandActions(role string) []string {
	switch role {
	case "main":
		return []string{"planets", "economy", "content", "broadcasts", "reports", "moderation", "team", "integrations"}
	case "technical":
		return []string{"planets", "content", "reports", "integrations"}
	case "marketing":
		return []string{"economy", "content", "broadcasts", "reports", "moderation"}
	default:
		return []string{}
	}
}

// Source BrandTeam_list is visible only to the active main administrator.
func brandTeamList(ctx context.Context, logger runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	callerID, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		BrandID string `json:"brandId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	brandID, err := requiredUUID(input.BrandID)
	if err != nil {
		return "", err
	}
	var allowed bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM jazebe_active_brand_admin
		WHERE brand_id=$1 AND user_id=$2 AND role='main' AND is_active)`, brandID, callerID).Scan(&allowed); err != nil {
		logger.Error("Jazebe brand team access check failed: %v", err)
		return "", status.Error(codes.Internal, "brand team unavailable")
	}
	if !allowed {
		return "", status.Error(codes.PermissionDenied, "active main brand administrator required")
	}
	rows, err := db.QueryContext(ctx, `SELECT a.user_id,a.full_name,a.role,a.is_active,
		COALESCE(u.display_name,''),a.created_at FROM jazebe_brand_admin a
		JOIN users u ON u.id=a.user_id WHERE a.brand_id=$1 ORDER BY a.created_at ASC`, brandID)
	if err != nil {
		logger.Error("Jazebe brand team list failed: %v", err)
		return "", status.Error(codes.Internal, "brand team unavailable")
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, name, role, displayName string
		var active bool
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &role, &active, &displayName, &createdAt); err != nil {
			return "", status.Error(codes.Internal, "brand team unavailable")
		}
		items = append(items, map[string]any{"id": id, "fullName": name, "role": role,
			"isActive": active, "phone": nil, "displayName": displayName,
			"isMe": id == callerID, "actions": brandActions(role), "createdAt": createdAt})
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "brand team unavailable")
	}
	return encode(map[string]any{"members": items})
}
