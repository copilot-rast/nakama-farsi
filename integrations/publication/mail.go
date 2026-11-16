package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
)

func (s *service) rpcInboxClaim(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		NotificationID string `json:"notificationId"`
		Lang           string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	id := strings.TrimSpace(in.NotificationID)
	if _, err := uuid.FromString(id); err != nil {
		return nil, badRequest("INVALID_ID", "notificationId", "send the message id from publication_inbox")
	}
	if !economyhooks.PrizesAvailable() {
		return nil, fail(http.StatusServiceUnavailable, "PRIZES_UNAVAILABLE", "", "the economy plugin is not loaded")
	}
	var prizeKey string
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var prize sql.NullString
		var claimed sql.NullTime
		err := tx.QueryRowContext(ctx, `SELECT p.key, r.claimed_at FROM publication_notification_recipient r JOIN publication_notification n ON n.id=r.notification_id
			LEFT JOIN economy_prize p ON p.id=n.prize_id
			WHERE r.notification_id=$1 AND r.user_id=$2 AND r.status='sent' AND n.channel IN ('in_app','push_in_app') FOR UPDATE OF r`, id, userID).Scan(&prize, &claimed)
		if errors.Is(err, sql.ErrNoRows) {
			return notFound("MESSAGE_NOT_FOUND")
		}
		if err != nil {
			return err
		}
		if !prize.Valid {
			return fail(http.StatusConflict, "MESSAGE_HAS_NO_PRIZE", "", "this message has no prize")
		}
		if claimed.Valid {
			return fail(http.StatusConflict, "ALREADY_CLAIMED", "", "the prize of this message was already claimed")
		}
		prizeKey = prize.String
		_, err = tx.ExecContext(ctx, `UPDATE publication_notification_recipient SET claimed_at=now(), seen_at=coalesce(seen_at, now()) WHERE notification_id=$1 AND user_id=$2`, id, userID)
		return err
	})
	if err != nil {
		return nil, err
	}
	res, err := economyhooks.GrantPrize(ctx, economyhooks.PrizeGrant{UserID: userID, PrizeKey: prizeKey, SourceType: "mail", SourceID: id + ":" + userID, Reason: "mail"})
	if err != nil {
		_, _ = s.db.ExecContext(ctx, `UPDATE publication_notification_recipient SET claimed_at=NULL WHERE notification_id=$1 AND user_id=$2`, id, userID)
		return nil, err
	}
	activity.Safe(ctx, s.db, s.logger, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindPrize, Action: "mail_claimed", RefID: id, Actor: "player", Data: map[string]any{"prize": prizeKey}}))
	return map[string]any{"notificationId": id, "claimed": true, "prizes": res.Prizes}, nil
}
