package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/chatfilter"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/hookchain"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/restrict"
	"google.golang.org/grpc/codes"
)

const (
	CodeChatBlocked = "CHAT_BLOCKED"
	maxReportNote   = 500
)

func ctxUser(ctx context.Context) string {
	id, _ := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	return id
}

func ctxLang(ctx context.Context) string {
	v, _ := ctx.Value(runtime.RUNTIME_CTX_LANG).(string)
	return v
}

func (s *service) chatGate(ctx context.Context, userID string) error {
	for _, kind := range []string{restrict.KindSuspend, restrict.KindMute} {
		r, err := restrict.Active(ctx, s.db, userID, kind)
		if err != nil {
			return s.hookFailure(err)
		}
		if r != nil {
			n := r.Notice(ctx, s.db, "")
			n.Code = restrict.CodeMuted
			return n.Error()
		}
	}
	return nil
}

func filterValue(f *chatfilter.Filter, v any, res *chatfilter.Result) any {
	switch x := v.(type) {
	case string:
		r := f.Apply(x)
		if r.Blocked {
			res.Blocked = true
		}
		if r.Flagged {
			res.Flagged = true
		}
		if r.Changed {
			res.Changed = true
		}
		res.Hits = append(res.Hits, r.Hits...)
		return r.Text
	case map[string]any:
		for k, item := range x {
			x[k] = filterValue(f, item, res)
		}
		return x
	case []any:
		for i, item := range x {
			x[i] = filterValue(f, item, res)
		}
		return x
	}
	return v
}

func filterContent(f *chatfilter.Filter, content string) (string, chatfilter.Result) {
	res := chatfilter.Result{Text: content}
	if f == nil || f.Len() == 0 || content == "" {
		return content, res
	}
	var doc any
	if err := json.Unmarshal([]byte(content), &doc); err != nil {
		return content, res
	}
	doc = filterValue(f, doc, &res)
	seen := map[string]bool{}
	hits := res.Hits[:0]
	for _, h := range res.Hits {
		if !seen[h.RuleID+"|"+h.Word] {
			seen[h.RuleID+"|"+h.Word] = true
			hits = append(hits, h)
		}
	}
	res.Hits = hits
	if !res.Changed {
		return content, res
	}
	out, err := json.Marshal(doc)
	if err != nil {
		return content, res
	}
	res.Text = string(out)
	return res.Text, res
}

func (s *service) moderateContent(ctx context.Context, userID, content string) (string, error) {
	if err := s.chatGate(ctx, userID); err != nil {
		return "", err
	}
	f, err := s.loadFilter(ctx, s.db)
	if err != nil {
		return "", s.hookFailure(err)
	}
	out, res := filterContent(f, content)
	if res.Blocked {
		return "", runtime.NewError(CodeChatBlocked, int(codes.InvalidArgument))
	}
	return out, nil
}

func (s *service) beforeChannelMessageSend(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *rtapi.Envelope) (*rtapi.Envelope, error) {
	msg := in.GetChannelMessageSend()
	if msg == nil {
		return in, nil
	}
	if err := s.dmMessageGate(ctx, ctxUser(ctx), msg.ChannelId); err != nil {
		return nil, err
	}
	content, err := s.moderateContent(ctx, ctxUser(ctx), msg.Content)
	if err != nil {
		return nil, err
	}
	content, isSticker, stickers, err := s.stickerGate(ctx, ctxUser(ctx), msg.ChannelId, content)
	if err != nil {
		return nil, err
	}
	if !isSticker || stickers.CountQuota {
		if err := s.chatQuotaGate(ctx, ctxUser(ctx), msg.ChannelId, content); err != nil {
			return nil, err
		}
	}
	msg.Content = content
	return in, nil
}

func (s *service) beforeChannelMessageUpdate(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *rtapi.Envelope) (*rtapi.Envelope, error) {
	msg := in.GetChannelMessageUpdate()
	if msg == nil {
		return in, nil
	}
	content, err := s.moderateContent(ctx, ctxUser(ctx), msg.Content)
	if err != nil {
		return nil, err
	}
	msg.Content = content
	return in, nil
}

func (s *service) afterChannelMessageSend(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out, in *rtapi.Envelope) error {
	ack := out.GetChannelMessageAck()
	msg := in.GetChannelMessageSend()
	if ack == nil || msg == nil || !ack.GetPersistent().GetValue() {
		return nil
	}
	f, err := s.loadFilter(ctx, s.db)
	if err != nil {
		return err
	}
	_, res := filterContent(f, msg.Content)
	if !res.Flagged {
		return nil
	}
	flagged := []chatfilter.Hit{}
	for _, h := range res.Hits {
		if h.Action == chatfilter.ActionFlag {
			flagged = append(flagged, h)
		}
	}
	if _, _, err := s.insertReport(ctx, "filter", ack.MessageId, "", "", "", localized.Text{}, "", flagged, 0); err != nil {
		s.logger.Warn("Base chat flag not stored: %v", err)
	}
	return nil
}

func (s *service) userBy(ctx context.Context, query string, arg string) string {
	if strings.TrimSpace(arg) == "" {
		return ""
	}
	var id string
	if err := s.db.QueryRowContext(ctx, query, arg).Scan(&id); err != nil {
		return ""
	}
	return id
}

func (s *service) refuseSuspended(ctx context.Context, userID string) error {
	if userID == "" {
		return nil
	}
	if err := restrict.RefuseSuspended(ctx, s.db, userID, ctxLang(ctx)); err != nil {
		if _, ok := err.(*runtime.Error); ok {
			return err
		}
		return s.hookFailure(err)
	}
	return nil
}

func (s *service) revokeIfSuspended(ctx context.Context, session *api.Session) {
	if session == nil {
		return
	}
	uid := restrict.TokenUserID(session.Token)
	if uid == "" {
		return
	}
	r, err := restrict.Active(ctx, s.db, uid, restrict.KindSuspend)
	if err != nil || r == nil {
		return
	}
	if err := s.nk.SessionLogout(uid, session.Token, session.RefreshToken); err != nil {
		s.logger.Warn("Base suspended sign-in not revoked: %v", err)
	}
}

func (s *service) registerModerationHooks(initializer runtime.Initializer) error {
	steps := []func() error{
		func() error {
			return initializer.RegisterBeforeRt("ChannelMessageSend", hookchain.Before("rt:ChannelMessageSend", s.beforeChannelMessageSend))
		},
		func() error {
			return initializer.RegisterBeforeRt("ChannelMessageUpdate", hookchain.Before("rt:ChannelMessageUpdate", s.beforeChannelMessageUpdate))
		},
		func() error {
			return initializer.RegisterAfterRt("ChannelMessageSend", hookchain.AfterRt("ChannelMessageSend", s.afterChannelMessageSend))
		},
		func() error {
			return initializer.RegisterBeforeSessionRefresh(hookchain.Before("SessionRefresh", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, in *api.SessionRefreshRequest) (*api.SessionRefreshRequest, error) {
				return in, s.refuseSuspended(ctx, restrict.TokenUserID(in.Token))
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateGoogle(hookchain.After("AuthenticateGoogle", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateGoogleRequest) error {
				s.revokeIfSuspended(ctx, out)
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateApple(hookchain.After("AuthenticateApple", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateAppleRequest) error {
				s.revokeIfSuspended(ctx, out)
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateFacebook(hookchain.After("AuthenticateFacebook", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateFacebookRequest) error {
				s.revokeIfSuspended(ctx, out)
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateSteam(hookchain.After("AuthenticateSteam", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateSteamRequest) error {
				s.revokeIfSuspended(ctx, out)
				return nil
			}))
		},
		func() error {
			return initializer.RegisterAfterAuthenticateGameCenter(hookchain.After("AuthenticateGameCenter", func(ctx context.Context, _ runtime.Logger, _ *sql.DB, _ runtime.NakamaModule, out *api.Session, _ *api.AuthenticateGameCenterRequest) error {
				s.revokeIfSuspended(ctx, out)
				return nil
			}))
		},
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) rpcMyTags(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT t.key, t.key_type, t.name::text, t.description::text, t.icon_url, t.metadata::text, t.sort, ut.expires_at FROM base_tag t JOIN base_user_tag ut ON ut.tag_id=t.id
		WHERE t.key_type='user' AND t.is_public AND ut.user_id=$1::uuid AND (ut.expires_at IS NULL OR ut.expires_at > now()) ORDER BY t.sort, t.key`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var key, kt, name, desc, icon, meta string
		var sort int
		var exp sql.NullTime
		if err := rows.Scan(&key, &kt, &name, &desc, &icon, &meta, &sort, &exp); err != nil {
			return nil, err
		}
		var expires any
		if exp.Valid {
			expires = exp.Time.UTC().Format(time.RFC3339)
		}
		items = append(items, map[string]any{"key": key, "keyType": kt, "name": langs.Resolve(localized.Parse(name), lang), "description": langs.Resolve(localized.Parse(desc), lang),
			"iconUrl": icon, "metadata": rawJSON(meta), "sort": sort, "expiresAt": expires})
	}
	return map[string]any{"lang": lang, "items": items}, rows.Err()
}

func (s *service) rpcMyStatus(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	suspension, err := restrictionNotice(ctx, s, userID, restrict.KindSuspend, in.Lang)
	if err != nil {
		return nil, err
	}
	mute, err := restrictionNotice(ctx, s, userID, restrict.KindMute, in.Lang)
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, reason::text, created_at FROM base_restriction WHERE user_id=$1 AND kind='warn' AND created_at > now() - INTERVAL '7 days' ORDER BY created_at DESC LIMIT 5`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	warnings := []map[string]any{}
	for rows.Next() {
		var id, reason string
		var at time.Time
		if err := rows.Scan(&id, &reason, &at); err != nil {
			return nil, err
		}
		warnings = append(warnings, map[string]any{"id": id, "reason": langs.Resolve(localized.Parse(reason), lang), "createdAt": at.UTC().Format(time.RFC3339)})
	}
	return map[string]any{"lang": lang, "suspension": suspension, "mute": mute, "warnings": warnings}, rows.Err()
}

func (s *service) rpcReportTitles(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT key, name::text, description::text FROM base_report_title WHERE is_active ORDER BY sort, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var key, name, desc string
		if err := rows.Scan(&key, &name, &desc); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"key": key, "name": langs.Resolve(localized.Parse(name), lang), "description": langs.Resolve(localized.Parse(desc), lang)})
	}
	return map[string]any{"lang": lang, "items": items}, rows.Err()
}

func (s *service) canSeeMessage(ctx context.Context, userID string, msg map[string]any) bool {
	switch msg["mode"].(int) {
	case 2:
		return true
	case 3:
		var ok bool
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM group_edge WHERE source_id=$1::uuid AND destination_id=$2::uuid AND state <= 2)`, msg["subject"], userID).Scan(&ok)
		return ok
	case 4:
		return msg["subject"] == userID || msg["descriptor"] == userID
	}
	return false
}

func (s *service) rpcReportMessage(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		MessageID string `json:"messageId"`
		TitleKey  string `json:"titleKey"`
		Note      string `json:"note"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if _, err := uuid.FromString(strings.TrimSpace(in.MessageID)); err != nil {
		return nil, badRequest("MESSAGE_NOT_FOUND", "messageId", "message does not exist")
	}
	note := strings.TrimSpace(in.Note)
	if len([]rune(note)) > maxReportNote {
		return nil, badRequest("TEXT_TOO_LONG", "note", "at most 500 letters")
	}
	var titleID, titleName string
	err := s.db.QueryRowContext(ctx, `SELECT id::text, name::text FROM base_report_title WHERE key=$1 AND is_active`, strings.TrimSpace(in.TitleKey)).Scan(&titleID, &titleName)
	if err == sql.ErrNoRows {
		return nil, badRequest("REPORT_TITLE_NOT_FOUND", "titleKey", "pick one of the report reasons")
	}
	if err != nil {
		return nil, err
	}
	msg, _, err := s.messageContext(ctx, in.MessageID, 0, 0)
	if err != nil {
		return nil, err
	}
	if msg["senderId"] == userID {
		return nil, badRequest("CANNOT_REPORT_OWN", "messageId", "you cannot report your own message")
	}
	if !s.canSeeMessage(ctx, userID, msg) {
		return nil, notFound("MESSAGE_NOT_FOUND")
	}
	d, err := clusterkit.Allow(ctx, s.db, "chat-report:"+userID, 20, time.Hour)
	if err != nil {
		return nil, err
	}
	if !d.Allowed {
		return nil, fail(429, "TOO_MANY_REPORTS", "", "too many reports, try again later")
	}
	var username string
	_ = s.db.QueryRowContext(ctx, `SELECT username FROM users WHERE id=$1`, userID).Scan(&username)
	id, dup, err := s.insertReport(ctx, "player", in.MessageID, userID, username, titleID, localized.Parse(titleName), note, nil, 5)
	if err != nil {
		return nil, err
	}
	if !dup {
		s.record(ctx, activity.Entry{UserID: userID, Kind: activity.KindReport, Action: "filed", RefID: id, Actor: "player", Data: map[string]any{"reason": in.TitleKey}})
		if sender, ok := msg["senderId"].(string); ok && sender != "" {
			s.record(ctx, activity.Entry{UserID: sender, Kind: activity.KindReport, Action: "received", RefID: id, Actor: "player", Data: map[string]any{"reason": in.TitleKey}})
		}
	}
	return map[string]any{"reportId": id, "duplicate": dup}, nil
}
