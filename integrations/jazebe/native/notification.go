package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// Persist a Nakama notification in the caller's transaction. The stable ID
// prevents a retried reward or adjustment from creating duplicate notices.
// Nakama's existing notification list reads this row without a parallel store.
func persistJazebeNotification(ctx context.Context, tx *sql.Tx, userID, sourceKey, subject string, content map[string]any, at time.Time) error {
	// A notice the recipient muted is not delivered in-app at all (source
	// resolveChannels); critical notices ignore preferences.
	delivered, err := notificationDelivered(ctx, tx, userID, content)
	if err != nil || !delivered {
		return err
	}
	encoded, err := json.Marshal(content)
	if err != nil {
		return err
	}
	id := jazebeNotificationID(userID, sourceKey)
	_, err = tx.ExecContext(ctx, `INSERT INTO notification (id,user_id,subject,content,code,sender_id,create_time)
		VALUES ($1,$2,$3,$4::jsonb,1001,'00000000-0000-0000-0000-000000000000',$5)
		ON CONFLICT (id) DO NOTHING`, id, userID, subject, string(encoded), at)
	return err
}

func jazebeNotificationID(userID, sourceKey string) string {
	return adjustmentSourceID("jazebe-notification:" + userID + ":" + sourceKey)
}
