package main

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Source AdminSupportService.constellation joins each current planet membership
// to its wallet, progression and active group. This uses Nakama user IDs and
// the existing Jazebe tables without introducing a second account identity.
func operatorUserConstellation(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	if _, err := operatorReportActor(ctx); err != nil {
		return "", err
	}
	var in struct {
		UserID string `json:"userId"`
	}
	if err := decode(payload, &in); err != nil {
		return "", err
	}
	id, err := requiredUUID(in.UserID)
	if err != nil {
		return "", err
	}
	var exists int
	err = db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE id=$1`, id).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.NotFound, "account not found")
	}
	if err != nil {
		return "", err
	}
	rows, err := db.QueryContext(ctx, `SELECT m.planet_id,p.name,p.status,p.point_name,b.name,m.join_path,
		m.joined_at,m.is_active,m.sort_order,m.last_activity_at,COALESCE(w.balance,0),
		COALESCE(s.total_xp,0),COALESCE(s.level,1),COALESCE(s.medals_count,0),
		g.id,g.name,g.status,gm.role
		FROM jazebe_planet_membership m
		JOIN jazebe_planet p ON p.id=m.planet_id
		JOIN jazebe_brand b ON b.id=p.brand_id
		LEFT JOIN jazebe_point_wallet w ON w.planet_id=m.planet_id AND w.user_id=m.user_id
		LEFT JOIN jazebe_user_planet_stats s ON s.planet_id=m.planet_id AND s.user_id=m.user_id
		LEFT JOIN jazebe_group_membership gm ON gm.planet_id=m.planet_id AND gm.user_id=m.user_id AND gm.status='active'
		LEFT JOIN jazebe_group g ON g.id=gm.group_id
		WHERE m.user_id=$1
		ORDER BY m.sort_order ASC NULLS LAST,m.last_activity_at DESC NULLS LAST,m.joined_at DESC
		LIMIT 501`, id)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var planetID, planetName, planetStatus, pointName, brand, joinPath string
		var joinedAt time.Time
		var active bool
		var sortOrder sql.NullInt64
		var lastActivity sql.NullTime
		var balance, totalXP int64
		var level, medals int
		var groupID, groupName, groupStatus, groupRole sql.NullString
		if err = rows.Scan(&planetID, &planetName, &planetStatus, &pointName, &brand, &joinPath,
			&joinedAt, &active, &sortOrder, &lastActivity, &balance, &totalXP, &level, &medals,
			&groupID, &groupName, &groupStatus, &groupRole); err != nil {
			return "", err
		}
		var order, activity, group any
		if sortOrder.Valid {
			order = sortOrder.Int64
		}
		if lastActivity.Valid {
			activity = lastActivity.Time
		}
		if groupID.Valid {
			group = map[string]any{"id": groupID.String, "name": groupName.String,
				"role": groupRole.String, "status": groupStatus.String}
		}
		items = append(items, map[string]any{
			"planetId": planetID, "planetName": planetName, "planetStatus": planetStatus,
			"brand": brand, "joinPath": joinPath, "joinedAt": joinedAt,
			"membershipActive": active, "sortOrder": order, "lastActivityAt": activity,
			"balance": balance, "pointName": pointName, "totalXp": totalXP,
			"level": level, "medals": medals, "group": group,
		})
	}
	if err = rows.Err(); err != nil {
		return "", err
	}
	if len(items) > 500 {
		return "", status.Error(codes.ResourceExhausted, "too many memberships for support response")
	}
	return encode(items)
}
