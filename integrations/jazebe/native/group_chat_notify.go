package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

// Nakama notification rows have no read timestamp; clients clear them by
// deletion. A stable per-user/group ID refreshes one outstanding notice.
func groupChatNotifyAbsent(ctx context.Context, db *sql.DB, nk runtime.NakamaModule, ref groupRef, senderID, senderName, text string) error {
	presences, err := nk.StreamUserList(groupChatStreamMode, ref.GroupID, ref.PlanetID, groupChatStreamLabel, true, true)
	if err != nil {
		return err
	}
	present := make(map[string]struct{}, len(presences))
	for _, p := range presences {
		present[p.GetUserId()] = struct{}{}
	}
	rows, err := db.QueryContext(ctx, `SELECT m.user_id,g.name FROM jazebe_group_membership m JOIN jazebe_group g ON g.id=m.group_id
		JOIN jazebe_planet p ON p.id=g.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		LEFT JOIN jazebe_planet_membership pm ON pm.planet_id=m.planet_id AND pm.user_id=m.user_id
		WHERE m.group_id=$1 AND m.status='active' AND g.status='active' AND p.status='active'
		AND b.is_active AND COALESCE(pm.is_active,true)`, ref.GroupID)
	if err != nil {
		return err
	}
	type recipient struct{ id, name string }
	absent := make([]recipient, 0)
	for rows.Next() {
		var r recipient
		if err := rows.Scan(&r.id, &r.name); err != nil {
			rows.Close()
			return err
		}
		if r.id != senderID {
			if _, ok := present[r.id]; !ok {
				absent = append(absent, r)
			}
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	preview := []rune(text)
	if len(preview) > 80 {
		preview = preview[:80]
	}
	for _, r := range absent {
		content, _ := json.Marshal(map[string]any{"category": "group_message", "actionLink": "/group/chat", "groupId": ref.GroupID, "body": senderName + ": " + string(preview)})
		id := adjustmentSourceID("jazebe-group-chat-unread:" + ref.GroupID + ":" + r.id)
		_, err := db.ExecContext(ctx, `INSERT INTO notification(id,user_id,subject,content,code,sender_id,create_time)
			VALUES($1,$2,$3,$4::jsonb,1001,'00000000-0000-0000-0000-000000000000',now())
			ON CONFLICT(id) DO UPDATE SET subject=EXCLUDED.subject,content=EXCLUDED.content,create_time=EXCLUDED.create_time`,
			id, r.id, "پیام جدید در «"+r.name+"»", string(content))
		if err != nil {
			return err
		}
	}
	return nil
}
