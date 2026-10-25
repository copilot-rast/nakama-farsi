package main

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func chatMember(ctx context.Context, tx *sql.Tx, ref groupRef, user string, active bool) (string, error) {
	var role, state, planetState, name string
	var brandActive, planetMemberActive bool
	err := tx.QueryRowContext(ctx, `SELECT m.role,g.status,p.status,g.name,b.is_active,COALESCE(pm.is_active,true)
		FROM jazebe_group g JOIN jazebe_planet p ON p.id=g.planet_id JOIN jazebe_brand b ON b.id=p.brand_id
		JOIN jazebe_group_membership m ON m.group_id=g.id AND m.user_id=$3 AND m.status='active'
		LEFT JOIN jazebe_planet_membership pm ON pm.planet_id=p.id AND pm.user_id=$3
		WHERE g.id=$1 AND g.planet_id=$2 FOR SHARE OF g`, ref.GroupID, ref.PlanetID, user).
		Scan(&role, &state, &planetState, &name, &brandActive, &planetMemberActive)
	if errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.PermissionDenied, "active group membership required")
	}
	if err != nil {
		return "", status.Error(codes.Internal, "group chat access unavailable")
	}
	if planetState != "active" || !brandActive || !planetMemberActive || state == "disbanded" {
		return "", status.Error(codes.PermissionDenied, "group unavailable")
	}
	if active && state != "active" {
		return "", status.Error(codes.FailedPrecondition, "group suspended")
	}
	return role, nil
}

type chatView struct {
	Seq       int64     `json:"seq"`
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Sender    any       `json:"sender"`
	Mine      bool      `json:"mine"`
	Text      any       `json:"text"`
	ReplyTo   any       `json:"replyTo"`
	Deleted   any       `json:"deleted"`
	CreatedAt time.Time `json:"createdAt"`
}

func chatRows(ctx context.Context, tx *sql.Tx, groupID, viewer string, before int64, limit int) ([]chatView, error) {
	rows, err := tx.QueryContext(ctx, `SELECT m.seq,m.id,m.kind,m.sender_user_id,m.body,m.deleted_at,m.deleted_by,m.deleted_as_admin,m.created_at,
		COALESCE(NULLIF(trim(s.display_name),''),'کاربر ناشناس'),COALESCE(NULLIF(trim(d.display_name),''),'کاربر ناشناس'),
		q.seq,q.body,q.deleted_at,q.sender_user_id,COALESCE(NULLIF(trim(qs.display_name),''),'کاربر ناشناس')
		FROM jazebe_group_message m LEFT JOIN users s ON s.id=m.sender_user_id
		LEFT JOIN users d ON d.id=m.deleted_by LEFT JOIN jazebe_group_message q ON q.id=m.reply_to_message_id
		LEFT JOIN users qs ON qs.id=q.sender_user_id
		WHERE m.group_id=$1 AND ($2::bigint=0 OR m.seq<$2)
		ORDER BY m.seq DESC LIMIT $3`, groupID, before, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, "group messages unavailable")
	}
	defer rows.Close()
	items := make([]chatView, 0)
	for rows.Next() {
		var v chatView
		var sender, body, deletedBy, quotedBody, quotedUser sql.NullString
		var deleted, quotedDeleted sql.NullTime
		var quotedSeq sql.NullInt64
		var asAdmin bool
		var senderName, deletedName, quotedName string
		if err := rows.Scan(&v.Seq, &v.ID, &v.Kind, &sender, &body, &deleted, &deletedBy, &asAdmin, &v.CreatedAt,
			&senderName, &deletedName, &quotedSeq, &quotedBody, &quotedDeleted, &quotedUser, &quotedName); err != nil {
			return nil, status.Error(codes.Internal, "group messages unavailable")
		}
		if sender.Valid {
			v.Sender = map[string]any{"id": sender.String, "name": senderName, "initial": groupInitial(senderName), "color": groupMemberColor(sender.String)}
		}
		v.Mine = sender.Valid && sender.String == viewer
		if !deleted.Valid && body.Valid {
			v.Text = body.String
		}
		if quotedSeq.Valid {
			name := "سیستم"
			if quotedUser.Valid {
				name = quotedName
			}
			text := quotedBody.String
			if quotedDeleted.Valid {
				text = "پیام حذف شده"
			}
			v.ReplyTo = map[string]any{"seq": quotedSeq.Int64, "sender": name, "text": text}
		}
		if deleted.Valid {
			var by any
			if deletedBy.Valid {
				by = deletedName
			}
			v.Deleted = map[string]any{"asAdmin": asAdmin, "byName": by, "at": deleted.Time}
		}
		items = append(items, v)
	}
	if rows.Err() != nil {
		return nil, status.Error(codes.Internal, "group messages unavailable")
	}
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
	return items, nil
}

func groupChatList(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
		GroupID  string `json:"groupId"`
		Before   *int64 `json:"before"`
		Limit    *int   `json:"limit"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	ref, err := readGroupRef(jsonString(groupRef{PlanetID: input.PlanetID, GroupID: input.GroupID}))
	if err != nil {
		return "", err
	}
	if (input.Before != nil && *input.Before < 1) || (input.Limit != nil && (*input.Limit < 1 || *input.Limit > 100)) {
		return "", status.Error(codes.InvalidArgument, "invalid message page")
	}
	limit := 50
	if input.Limit != nil {
		limit = *input.Limit
	}
	before := int64(0)
	if input.Before != nil {
		before = *input.Before
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "chat unavailable")
	}
	defer tx.Rollback()
	if _, err := chatMember(ctx, tx, ref, user, false); err != nil {
		return "", err
	}
	items, err := chatRows(ctx, tx, ref.GroupID, user, before, limit+1)
	if err != nil {
		return "", err
	}
	hasMore := len(items) > limit
	if hasMore {
		items = items[1:]
	}
	var lastRead int64
	err = tx.QueryRowContext(ctx, `SELECT last_read_seq FROM jazebe_group_message_read WHERE group_id=$1 AND user_id=$2`, ref.GroupID, user).Scan(&lastRead)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", status.Error(codes.Internal, "message cursor unavailable")
	}
	var oldest, newest any
	if len(items) > 0 {
		oldest = items[0].Seq
		newest = items[len(items)-1].Seq
	}
	return encode(map[string]any{"items": items, "lastReadSeq": lastRead, "hasMore": hasMore, "oldestSeq": oldest, "newestSeq": newest})
}

func validateChatText(value string) (string, error) {
	value = strings.TrimSpace(value)
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) < 1 || utf8.RuneCountInString(value) > 2000 {
		return "", status.Error(codes.InvalidArgument, "message text must contain 1 to 2000 characters")
	}
	return value, nil
}

func groupChatSend(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID string `json:"planetId"`
		GroupID  string `json:"groupId"`
		Text     string `json:"text"`
		ReplyTo  *struct {
			Seq    int64  `json:"seq"`
			Sender string `json:"sender"`
			Text   string `json:"text"`
		} `json:"replyTo"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	ref, err := readGroupRef(jsonString(groupRef{PlanetID: input.PlanetID, GroupID: input.GroupID}))
	if err != nil {
		return "", err
	}
	input.Text, err = validateChatText(input.Text)
	if err != nil {
		return "", err
	}
	if input.ReplyTo != nil {
		if input.ReplyTo.Seq < 0 || !utf8.ValidString(input.ReplyTo.Sender) || utf8.RuneCountInString(input.ReplyTo.Sender) > 64 ||
			!utf8.ValidString(input.ReplyTo.Text) || utf8.RuneCountInString(input.ReplyTo.Text) > 2000 {
			return "", status.Error(codes.InvalidArgument, "invalid quoted message")
		}
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "chat unavailable")
	}
	defer tx.Rollback()
	if _, err := chatMember(ctx, tx, ref, user, true); err != nil {
		return "", err
	}
	if err := lockXact(ctx, tx, "774104:"+ref.GroupID+user); err != nil {
		return "", status.Error(codes.Internal, "chat rate check failed")
	}
	var recent int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM jazebe_group_message WHERE group_id=$1 AND sender_user_id=$2
		AND kind='text' AND created_at>now()-interval '1 minute'`, ref.GroupID, user).Scan(&recent); err != nil {
		return "", status.Error(codes.Internal, "chat rate check failed")
	}
	if recent >= 20 {
		return "", status.Error(codes.ResourceExhausted, "message rate limit reached")
	}
	var quoted any
	if input.ReplyTo != nil && input.ReplyTo.Seq > 0 {
		var quotedID string
		err = tx.QueryRowContext(ctx, `SELECT id FROM jazebe_group_message WHERE group_id=$1 AND seq=$2`, ref.GroupID, input.ReplyTo.Seq).Scan(&quotedID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", status.Error(codes.NotFound, "quoted message not found")
		}
		if err != nil {
			return "", status.Error(codes.Internal, "quoted message unavailable")
		}
		quoted = quotedID
	}
	var seq int64
	err = tx.QueryRowContext(ctx, `INSERT INTO jazebe_group_message(id,planet_id,group_id,sender_user_id,kind,body,reply_to_message_id)
		VALUES($1,$2,$3,$4,'text',$5,$6) RETURNING seq`, uuid.Must(uuid.NewV4()).String(), ref.PlanetID, ref.GroupID, user, input.Text, quoted).Scan(&seq)
	if err != nil {
		return "", status.Error(codes.Internal, "message send failed")
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO jazebe_group_message_read(planet_id,group_id,user_id,last_read_seq)
		VALUES($1,$2,$3,$4) ON CONFLICT(group_id,user_id) DO UPDATE
		SET last_read_seq=GREATEST(jazebe_group_message_read.last_read_seq,EXCLUDED.last_read_seq)`, ref.PlanetID, ref.GroupID, user, seq); err != nil {
		return "", status.Error(codes.Internal, "message cursor failed")
	}
	items, err := chatRows(ctx, tx, ref.GroupID, user, seq+1, 1)
	if err != nil {
		return "", err
	}
	if len(items) != 1 || items[0].Seq != seq {
		return "", status.Error(codes.Internal, "message projection failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "message send failed")
	}
	items[0].Mine = false
	if err := groupChatBroadcast(ctx, db, nk, ref, "message.created", items[0]); err != nil {
		logger.Warn("Jazebe group chat realtime delivery failed: %v", err)
	}
	senderName := "کاربر ناشناس"
	if sender, ok := items[0].Sender.(map[string]any); ok {
		if name, ok := sender["name"].(string); ok {
			senderName = name
		}
	}
	if err := groupChatNotifyAbsent(ctx, db, nk, ref, user, senderName, input.Text); err != nil {
		logger.Warn("Jazebe group chat unread notices failed: %v", err)
	}
	items[0].Mine = true
	return encode(items[0])
}

func groupChatMarkRead(ctx context.Context, _ runtime.Logger, db *sql.DB, _ runtime.NakamaModule, payload string) (string, error) {
	user, err := userID(ctx)
	if err != nil {
		return "", err
	}
	var input struct {
		PlanetID    string `json:"planetId"`
		GroupID     string `json:"groupId"`
		LastReadSeq int64  `json:"lastReadSeq"`
	}
	if err := decode(payload, &input); err != nil {
		return "", err
	}
	ref, err := readGroupRef(jsonString(groupRef{PlanetID: input.PlanetID, GroupID: input.GroupID}))
	if err != nil {
		return "", err
	}
	if input.LastReadSeq < 0 {
		return "", status.Error(codes.InvalidArgument, "invalid read cursor")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", status.Error(codes.Internal, "chat unavailable")
	}
	defer tx.Rollback()
	if _, err := chatMember(ctx, tx, ref, user, false); err != nil {
		return "", err
	}
	var last int64
	err = tx.QueryRowContext(ctx, `INSERT INTO jazebe_group_message_read(planet_id,group_id,user_id,last_read_seq)
		VALUES($1,$2,$3,$4) ON CONFLICT(group_id,user_id) DO UPDATE
		SET last_read_seq=GREATEST(jazebe_group_message_read.last_read_seq,EXCLUDED.last_read_seq)
		RETURNING last_read_seq`, ref.PlanetID, ref.GroupID, user, input.LastReadSeq).Scan(&last)
	if err != nil {
		return "", status.Error(codes.Internal, "read cursor update failed")
	}
	if err := tx.Commit(); err != nil {
		return "", status.Error(codes.Internal, "read cursor update failed")
	}
	return encode(map[string]any{"lastReadSeq": last})
}
