package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func groupRequestQueue(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	var role string
	err = db.QueryRowContext(ctx, `SELECT m.role FROM jazebe_group g JOIN jazebe_group_membership m ON m.group_id=g.id
		WHERE g.id=$1 AND g.planet_id=$2 AND m.user_id=$3 AND m.status='active'`, ref.GroupID, ref.PlanetID, actor).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.PermissionDenied, "group manager required")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group request queue unavailable")
	}
	if role != "creator" && role != "admin" {
		return "", status.Error(codes.PermissionDenied, "group manager required")
	}
	return groupRequestRows(ctx, db, ref)
}

func groupRequestRows(ctx context.Context, db *sql.DB, ref groupRef) (string, error) {
	rows, err := db.QueryContext(ctx, `SELECT r.id,r.user_id,coalesce(nullif(trim(u.display_name),''),'کاربر'),r.requested_at
		FROM jazebe_group_join_request r JOIN users u ON u.id=r.user_id
		WHERE r.group_id=$1 AND r.planet_id=$2 AND r.status='pending'
		ORDER BY r.requested_at ASC,r.id ASC`, ref.GroupID, ref.PlanetID)
	if err != nil {
		return "", status.Error(codes.Internal, "group request queue unavailable")
	}
	defer rows.Close()
	now := time.Now().UTC()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, user, name string
		var requested time.Time
		if err := rows.Scan(&id, &user, &name, &requested); err != nil {
			return "", status.Error(codes.Internal, "group request queue unavailable")
		}
		name = strings.TrimSpace(name)
		items = append(items, map[string]any{"id": id, "userId": user, "name": name, "initial": groupInitial(name),
			"color": groupMemberColor(user), "date": groupRelativeFa(requested, now), "requestedAt": requested.UTC()})
	}
	if err := rows.Err(); err != nil {
		return "", status.Error(codes.Internal, "group request queue unavailable")
	}
	return encode(map[string]any{"requests": items})
}
