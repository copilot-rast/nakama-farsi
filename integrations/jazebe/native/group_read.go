package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func requireActiveGroupViewer(ctx context.Context, db *sql.DB, ref groupRef, actor string, creatorOnly bool) error {
	var role string
	err := db.QueryRowContext(ctx, `SELECT m.role FROM jazebe_group_membership m
		JOIN jazebe_group g ON g.id=m.group_id JOIN jazebe_planet p ON p.id=g.planet_id
		JOIN jazebe_brand b ON b.id=p.brand_id
		LEFT JOIN jazebe_planet_membership pm ON pm.planet_id=p.id AND pm.user_id=m.user_id
		WHERE m.group_id=$1 AND m.planet_id=$2 AND m.user_id=$3 AND m.status='active'
		AND g.status<>'disbanded' AND p.status='active' AND b.is_active AND COALESCE(pm.is_active,true)`, ref.GroupID, ref.PlanetID, actor).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return status.Error(codes.PermissionDenied, "active group membership required")
	}
	if err != nil {
		return status.Error(codes.Internal, "group access unavailable")
	}
	if creatorOnly && role != "creator" {
		return status.Error(codes.PermissionDenied, "group creator required")
	}
	return nil
}

func groupMembers(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	if err := requireActiveGroupViewer(ctx, db, ref, actor, false); err != nil {
		return "", err
	}
	return groupMemberRows(ctx, db, ref)
}

func groupMemberRows(ctx context.Context, db *sql.DB, ref groupRef) (string, error) {
	rows, err := db.QueryContext(ctx, `SELECT m.user_id,m.role,m.joined_at,COALESCE(NULLIF(trim(u.display_name),''),'کاربر ناشناس')
		FROM jazebe_group_membership m JOIN users u ON u.id=m.user_id
		WHERE m.group_id=$1 AND m.planet_id=$2 AND m.status='active'
		ORDER BY m.joined_at ASC,m.id ASC LIMIT 100`, ref.GroupID, ref.PlanetID)
	if err != nil {
		return "", status.Error(codes.Internal, "group members unavailable")
	}
	defer rows.Close()
	type member struct {
		ID       string    `json:"id"`
		Name     string    `json:"name"`
		Initial  string    `json:"initial"`
		Color    string    `json:"color"`
		Role     string    `json:"role"`
		JoinedAt time.Time `json:"joinedAt"`
	}
	members := make([]member, 0)
	for rows.Next() {
		var m member
		if err := rows.Scan(&m.ID, &m.Role, &m.JoinedAt, &m.Name); err != nil {
			return "", status.Error(codes.Internal, "group members unavailable")
		}
		m.Initial = groupInitial(m.Name)
		m.Color = groupMemberColor(m.ID)
		members = append(members, m)
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "group members unavailable")
	}
	// The source takes the oldest hundred seats, then puts creator and admins first.
	for i := 1; i < len(members); i++ {
		for j := i; j > 0 && memberRolePriority(members[j].Role) < memberRolePriority(members[j-1].Role); j-- {
			members[j], members[j-1] = members[j-1], members[j]
		}
	}
	return encode(map[string]any{"members": members})
}

func memberRolePriority(role string) int {
	switch role {
	case "creator":
		return 0
	case "admin":
		return 1
	default:
		return 2
	}
}

func groupActivity(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	if err := requireActiveGroupViewer(ctx, db, ref, actor, false); err != nil {
		return "", err
	}
	return groupActivityRows(ctx, db, ref)
}

func groupActivityRows(ctx context.Context, db *sql.DB, ref groupRef) (string, error) {
	rows, err := db.QueryContext(ctx, `SELECT id,kind,message_key,payload,created_at FROM jazebe_group_activity
		WHERE group_id=$1 AND planet_id=$2 ORDER BY created_at DESC,id DESC LIMIT 50`, ref.GroupID, ref.PlanetID)
	if err != nil {
		return "", status.Error(codes.Internal, "group activity unavailable")
	}
	defer rows.Close()
	now := time.Now().UTC()
	activity := make([]map[string]any, 0)
	for rows.Next() {
		var id, kind, key string
		var raw []byte
		var at time.Time
		if err := rows.Scan(&id, &kind, &key, &raw, &at); err != nil {
			return "", status.Error(codes.Internal, "group activity unavailable")
		}
		var data map[string]any
		if len(raw) > 0 && json.Unmarshal(raw, &data) != nil {
			return "", status.Error(codes.Internal, "group activity invalid")
		}
		text := key
		if rendered, ok := data["text"].(string); ok {
			text = rendered
		}
		activity = append(activity, map[string]any{"id": id, "kind": groupFeedKind(kind), "text": text, "time": groupRelativeFa(at, now), "at": at})
	}
	if rows.Err() != nil {
		return "", status.Error(codes.Internal, "group activity unavailable")
	}
	return encode(map[string]any{"activity": activity})
}

func groupFeedKind(kind string) string {
	switch kind {
	case "ticket", "game", "prize":
		return kind
	case "mission":
		return "game"
	default:
		return "join"
	}
}

func groupMyRequest(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	actor, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	input.PlanetID, err = requiredUUID(input.PlanetID)
	if err != nil {
		return "", err
	}
	var id, name string
	var requested time.Time
	err = db.QueryRowContext(ctx, `SELECT g.id,g.name,r.requested_at FROM jazebe_group_join_request r
		JOIN jazebe_group g ON g.id=r.group_id JOIN jazebe_planet p ON p.id=g.planet_id
		JOIN jazebe_brand b ON b.id=p.brand_id LEFT JOIN jazebe_planet_membership pm ON pm.planet_id=p.id AND pm.user_id=r.user_id
		WHERE r.planet_id=$1 AND r.user_id=$2 AND r.status='pending' AND p.status='active'
		AND b.is_active AND COALESCE(pm.is_active,true)`, input.PlanetID, actor).Scan(&id, &name, &requested)
	if errors.Is(err, sql.ErrNoRows) {
		return encode(map[string]any{"request": nil})
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group request unavailable")
	}
	return encode(map[string]any{"request": map[string]any{"groupId": id, "groupName": name, "date": groupRelativeFa(requested, time.Now().UTC()), "requestedAt": requested}})
}
