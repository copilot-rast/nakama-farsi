package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const groupChatStreamMode uint8 = 100
const groupChatStreamLabel = "jazebe-group-chat"

// An authenticated Nakama session joins the custom stream only after a
// database membership check. The same stream address is used for deliveries.
func groupChatSubscribe(ctx context.Context, _ runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	session, ok := ctx.Value(runtime.RUNTIME_CTX_SESSION_ID).(string)
	if !ok || session == "" {
		return "", status.Error(codes.FailedPrecondition, "active realtime session required")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "group stream unavailable")
	}
	defer tx.Rollback()
	if _, err := chatMember(ctx, tx, ref, user, true); err != nil {
		return "", err
	}
	joined, err := nk.StreamUserJoin(groupChatStreamMode, ref.GroupID, ref.PlanetID, groupChatStreamLabel, user, session, false, true, "")
	if err != nil {
		return "", status.Error(codes.Internal, "group stream join failed")
	}
	if err := tx.Commit(); err != nil {
		_ = nk.StreamUserLeave(groupChatStreamMode, ref.GroupID, ref.PlanetID, groupChatStreamLabel, user, session)
		return "", status.Error(codes.Internal, "group stream unavailable")
	}
	return encode(map[string]any{"joined": joined, "mode": groupChatStreamMode, "subject": ref.GroupID, "subcontext": ref.PlanetID, "label": groupChatStreamLabel})
}

func groupChatUnsubscribe(ctx context.Context, _ runtime.Logger, _ *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	ref, err := readGroupRef(payload)
	if err != nil {
		return "", err
	}
	session, ok := ctx.Value(runtime.RUNTIME_CTX_SESSION_ID).(string)
	if !ok || session == "" {
		return "", status.Error(codes.FailedPrecondition, "active realtime session required")
	}
	if err := nk.StreamUserLeave(groupChatStreamMode, ref.GroupID, ref.PlanetID, groupChatStreamLabel, user, session); err != nil {
		return "", status.Error(codes.Internal, "group stream leave failed")
	}
	return encode(map[string]any{"left": true})
}

func groupChatBroadcast(ctx context.Context, db *sql.DB, nk runtime.NakamaModule, ref groupRef, event string, data any) error {
	// Filter the stream's current presences by committed membership. A delayed
	// eviction must not deliver a later message to someone already removed.
	var groupActive bool
	if err := db.QueryRowContext(ctx, `SELECT g.status='active' AND p.status='active' AND b.is_active
		FROM jazebe_group g JOIN jazebe_planet p ON p.id=g.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		WHERE g.id=$1 AND g.planet_id=$2`, ref.GroupID, ref.PlanetID).Scan(&groupActive); err != nil || !groupActive {
		return err
	}
	rows, err := db.QueryContext(ctx, `SELECT m.user_id FROM jazebe_group_membership m
		LEFT JOIN jazebe_planet_membership pm ON pm.planet_id=m.planet_id AND pm.user_id=m.user_id
		WHERE m.group_id=$1 AND m.status='active' AND COALESCE(pm.is_active,true)`, ref.GroupID)
	if err != nil {
		return err
	}
	allowed := make(map[string]struct{})
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		allowed[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	current, err := nk.StreamUserList(groupChatStreamMode, ref.GroupID, ref.PlanetID, groupChatStreamLabel, true, true)
	if err != nil {
		return err
	}
	presences := make([]runtime.Presence, 0, len(current))
	for _, presence := range current {
		if _, ok := allowed[presence.GetUserId()]; ok {
			presences = append(presences, presence)
		}
	}
	if len(presences) == 0 {
		return nil
	}
	message, _ := json.Marshal(map[string]any{"event": event, "data": data})
	return nk.StreamSend(groupChatStreamMode, ref.GroupID, ref.PlanetID, groupChatStreamLabel, string(message), presences, true)
}

func groupChatKickUser(nk runtime.NakamaModule, ref groupRef, user string) error {
	presences, err := nk.StreamUserList(groupChatStreamMode, ref.GroupID, ref.PlanetID, groupChatStreamLabel, true, true)
	if err != nil {
		return err
	}
	for _, p := range presences {
		if p.GetUserId() == user {
			if err := nk.StreamUserKick(groupChatStreamMode, ref.GroupID, ref.PlanetID, groupChatStreamLabel, p); err != nil {
				return err
			}
		}
	}
	return nil
}
