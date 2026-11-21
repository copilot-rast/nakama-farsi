package main

import (
	"context"
	"database/sql"
	"strings"

	"github.com/gofrs/uuid/v5"

	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/hookchain"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"google.golang.org/grpc/codes"
)

const (
	notifyFriendDeclined  = 1130
	notifyFriendRemoved   = 1131
	CodeFriendshipNeeded  = "FRIENDSHIP_REQUIRED"
	codeNoticeDeclined    = "FRIEND_REQUEST_DECLINED"
	codeNoticeRemoved     = "FRIEND_REMOVED"
	channelTypeDirectJoin = 2
)

var friendNoticeTexts = map[string]localized.Text{
	codeNoticeDeclined:   {"fa": "{name} درخواست دوستی‌ات را نپذیرفت.", "en": "{name} declined your friend request.", "ar": "رفض {name} طلب صداقتك."},
	codeNoticeRemoved:    {"fa": "{name} تو را از دوستانش حذف کرد.", "en": "{name} removed you as a friend.", "ar": "أزالك {name} من أصدقائه."},
	CodeFriendshipNeeded: {"fa": "پیام خصوصی فقط بین دوستان است.", "en": "Private chat is only between friends.", "ar": "الدردشة الخاصة بين الأصدقاء فقط."},
}

func (s *service) friendEdgeNotices(ctx context.Context, userID string, rejected, removed []string) {
	if len(rejected) == 0 && len(removed) == 0 {
		return
	}
	if len(rejected) > 0 {
		cycle, _ := dayBounds(s.now(), s.socialClock(ctx))
		for _, id := range rejected {
			if _, err := s.db.ExecContext(ctx, `INSERT INTO base_friend_suggest_skip (user_id, candidate_id, until_cycle) VALUES ($1,$2,$3)
				ON CONFLICT (user_id, candidate_id) DO UPDATE SET until_cycle=EXCLUDED.until_cycle`, id, userID, cycle); err != nil {
				s.logger.Warn("Base friend suggestion skip not saved: %v", err)
			}
		}
		if _, err := s.db.ExecContext(ctx, `DELETE FROM base_friend_suggest WHERE user_id = ANY($1::uuid[]) AND candidate_id=$2`, rejected, userID); err != nil {
			s.logger.Warn("Base friend suggestion not cleared: %v", err)
		}
	}
	rules, err := loadFriendRules(ctx, s.db)
	if err != nil {
		s.logger.Warn("Base friend notices skipped: %v", err)
		return
	}
	var name string
	_ = s.db.QueryRowContext(ctx, `SELECT COALESCE(NULLIF(display_name,''), username) FROM users WHERE id=$1`, userID).Scan(&name)
	send := func(to []string, code string, notify int) {
		for _, id := range to {
			subject := s.noticeText(ctx, id, code, name)
			content := map[string]any{"code": code, "userId": userID, "name": name}
			if err := s.nk.NotificationSend(ctx, id, subject, content, notify, userID, true); err != nil {
				s.logger.Warn("Base friend notice %s to %s not sent: %v", code, id, err)
			}
		}
	}
	if rules.NotifyDeclined {
		send(rejected, codeNoticeDeclined, notifyFriendDeclined)
	}
	if rules.NotifyRemoved {
		send(removed, codeNoticeRemoved, notifyFriendRemoved)
	}
}

func (s *service) noticeText(ctx context.Context, userID, code, name string) string {
	text := friendNoticeTexts[code]["en"]
	if langs, lang, err := localized.ForUser(ctx, s.db, userID, ""); err == nil && langs != nil {
		if v := langs.Resolve(friendNoticeTexts[code], lang); v != "" {
			text = v
		}
	}
	if name == "" {
		name = "?"
	}
	return strings.ReplaceAll(text, "{name}", name)
}

func (s *service) dmGate(ctx context.Context, userID, peer string) error {
	if userID == "" || peer == "" {
		return nil
	}
	rules, err := loadFriendRules(ctx, s.db)
	if err != nil {
		return s.hookFailure(err)
	}
	if !rules.DMFriendsOnly {
		return nil
	}
	ok, err := s.areFriends(ctx, userID, peer)
	if err != nil {
		return s.hookFailure(err)
	}
	if !ok {
		return s.socialRefusal(ctx, userID, CodeFriendshipNeeded, codes.PermissionDenied, friendNoticeTexts, map[string]any{"userId": peer}, nil)
	}
	return nil
}

func (s *service) beforeChannelJoin(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *rtapi.Envelope) (*rtapi.Envelope, error) {
	join := in.GetChannelJoin()
	if join == nil || join.Type != channelTypeDirectJoin {
		return in, nil
	}
	if err := s.dmGate(ctx, ctxUser(ctx), join.Target); err != nil {
		return nil, err
	}
	return in, nil
}

func (s *service) dmMessageGate(ctx context.Context, userID, channelID string) error {
	t, ok := parseChatChannel(channelID, userID)
	if !ok || t.Kind != "dm" {
		return nil
	}
	return s.dmGate(ctx, userID, t.Peer)
}

func (s *service) beforeStatusFollow(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *rtapi.Envelope) (*rtapi.Envelope, error) {
	follow := in.GetStatusFollow()
	userID := ctxUser(ctx)
	if follow == nil || userID == "" {
		return in, nil
	}
	rules, err := loadFriendRules(ctx, s.db)
	if err != nil {
		return nil, s.hookFailure(err)
	}
	if !rules.PresenceFriendsOnly {
		return in, nil
	}
	ids := []string{}
	for _, id := range follow.UserIds {
		if _, err := uuid.FromString(id); err == nil {
			ids = append(ids, id)
		}
	}
	names := append([]string{}, follow.Usernames...)
	rows, err := s.db.QueryContext(ctx, `SELECT u.id::text FROM users u JOIN user_edge e ON e.source_id=$1::uuid AND e.destination_id=u.id AND e.state=0
		WHERE u.id = ANY($2::uuid[]) OR u.username = ANY($3::text[])`, userID, ids, names)
	if err != nil {
		return nil, s.hookFailure(err)
	}
	defer rows.Close()
	keep := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, s.hookFailure(err)
		}
		keep = append(keep, id)
	}
	if err := rows.Err(); err != nil {
		return nil, s.hookFailure(err)
	}
	follow.UserIds = keep
	follow.Usernames = nil
	return in, nil
}

func (s *service) registerFriendExtraHooks(initializer runtime.Initializer) error {
	if err := initializer.RegisterBeforeRt("ChannelJoin", hookchain.Before("rt:ChannelJoin", s.beforeChannelJoin)); err != nil {
		return err
	}
	return initializer.RegisterBeforeRt("StatusFollow", hookchain.Before("rt:StatusFollow", s.beforeStatusFollow))
}
