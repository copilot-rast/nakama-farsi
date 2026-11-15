package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama/v3/internal/localized"
)

var newsStatuses = []string{"draft", "pending_review", "published", "archived", "rejected"}
var newsNotifyModes = []string{"none", "in_app", "push", "both"}

const newsBodyMax = 50000
const newsRichMax = 200000

type consoleUserKey struct{}

func withConsoleUser(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, consoleUserKey{}, consoleUser(r))
}

func ctxConsoleUser(ctx context.Context) string {
	v, _ := ctx.Value(consoleUserKey{}).(string)
	return v
}

type newsSettings struct {
	CommentsPerMinute  int        `json:"commentsPerMinute"`
	ReactionsPerMinute int        `json:"reactionsPerMinute"`
	UnreadScope        string     `json:"unreadScope"`
	ShowExpired        bool       `json:"showExpired"`
	FillDefaultLang    bool       `json:"fillDefaultLang"`
	RealtimeOnPublish  bool       `json:"realtimeOnPublish"`
	NotifyCategoryKey  string     `json:"notifyCategoryKey"`
	LogRetentionDays   int        `json:"logRetentionDays"`
	LogPayloadBytes    int        `json:"logPayloadBytes"`
	UpdatedBy          string     `json:"updatedBy"`
	UpdatedAt          *time.Time `json:"updatedAt,omitempty"`
}

func defaultNewsSettings() newsSettings {
	return newsSettings{CommentsPerMinute: commentsPerMinute, ReactionsPerMinute: 30, UnreadScope: "signup", FillDefaultLang: true, RealtimeOnPublish: true, NotifyCategoryKey: "news", LogRetentionDays: 30, LogPayloadBytes: 16384}
}

func loadNewsSettings(ctx context.Context, q queryer) (newsSettings, error) {
	st := defaultNewsSettings()
	var at time.Time
	err := q.QueryRowContext(ctx, `SELECT comments_per_minute, reactions_per_minute, unread_scope, show_expired, fill_default_lang, realtime_on_publish, notify_category_key, log_retention_days, log_payload_bytes, updated_by, updated_at
		FROM publication_news_settings WHERE id=1`).Scan(&st.CommentsPerMinute, &st.ReactionsPerMinute, &st.UnreadScope, &st.ShowExpired, &st.FillDefaultLang, &st.RealtimeOnPublish, &st.NotifyCategoryKey, &st.LogRetentionDays, &st.LogPayloadBytes, &st.UpdatedBy, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return defaultNewsSettings(), nil
	}
	if err != nil {
		return st, err
	}
	st.UpdatedAt = &at
	return st, nil
}

func (s *service) consoleNewsSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := loadNewsSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *service) consoleNewsSettingsPut(w http.ResponseWriter, r *http.Request) {
	in := defaultNewsSettings()
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	for _, c := range []struct {
		f         string
		v, lo, hi int
	}{{"commentsPerMinute", in.CommentsPerMinute, 1, 600}, {"reactionsPerMinute", in.ReactionsPerMinute, 1, 600}, {"logRetentionDays", in.LogRetentionDays, 1, 365}, {"logPayloadBytes", in.LogPayloadBytes, 0, 262144}} {
		if c.v < c.lo || c.v > c.hi {
			s.writeError(w, badRequest("OUT_OF_RANGE", c.f, "value is out of range"))
			return
		}
	}
	if in.UnreadScope != "signup" && in.UnreadScope != "all" {
		s.writeError(w, badRequest("INVALID_VALUE", "unreadScope", "must be signup or all"))
		return
	}
	in.NotifyCategoryKey = strings.TrimSpace(in.NotifyCategoryKey)
	var ok bool
	if err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM publication_notify_category WHERE key=$1)`, in.NotifyCategoryKey).Scan(&ok); err != nil {
		s.writeError(w, err)
		return
	}
	if !ok {
		s.writeError(w, badRequest("CATEGORY_NOT_FOUND", "notifyCategoryKey", "notification category does not exist"))
		return
	}
	_, err := s.db.ExecContext(r.Context(), `INSERT INTO publication_news_settings (id, comments_per_minute, reactions_per_minute, unread_scope, show_expired, fill_default_lang, realtime_on_publish, notify_category_key, log_retention_days, log_payload_bytes, updated_by, updated_at)
		VALUES (1,$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,now()) ON CONFLICT (id) DO UPDATE SET comments_per_minute=EXCLUDED.comments_per_minute, reactions_per_minute=EXCLUDED.reactions_per_minute,
		unread_scope=EXCLUDED.unread_scope, show_expired=EXCLUDED.show_expired, fill_default_lang=EXCLUDED.fill_default_lang, realtime_on_publish=EXCLUDED.realtime_on_publish,
		notify_category_key=EXCLUDED.notify_category_key, log_retention_days=EXCLUDED.log_retention_days, log_payload_bytes=EXCLUDED.log_payload_bytes, updated_by=EXCLUDED.updated_by, updated_at=now()`,
		in.CommentsPerMinute, in.ReactionsPerMinute, in.UnreadScope, in.ShowExpired, in.FillDefaultLang, in.RealtimeOnPublish, in.NotifyCategoryKey, in.LogRetentionDays, in.LogPayloadBytes, consoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.consoleNewsSettingsGet(w, r)
}

func cleanNewsBody(langs *localized.Set, format string, in i18n, required bool) (i18n, error) {
	if format == "markdown" {
		return cleanI18n(langs, "body", in, required, newsBodyMax)
	}
	clean := i18n{}
	for k, v := range in {
		if utf8.RuneCountInString(v) > newsRichMax {
			return nil, badRequest("TEXT_TOO_LONG", "body", "text is too long (max 200000)")
		}
		if out := sanitizeRichHTML(v); out != "" && strings.TrimSpace(richText(out)) != "" || strings.Contains(out, "<table") {
			clean[k] = out
		}
	}
	return cleanI18n(langs, "body", clean, required, newsRichMax)
}

func saveNews(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Title          i18n            `json:"title"`
		Summary        i18n            `json:"summary"`
		Body           i18n            `json:"body"`
		BodyFormat     string          `json:"bodyFormat"`
		CoverURL       string          `json:"coverUrl"`
		CategoryID     string          `json:"categoryId"`
		TagIDs         []string        `json:"tagIds"`
		Status         string          `json:"status"`
		IsActive       *bool           `json:"isActive"`
		IsPinned       bool            `json:"isPinned"`
		PublishedAt    string          `json:"publishedAt"`
		ExpiresAt      string          `json:"expiresAt"`
		AllowComment   *bool           `json:"allowComment"`
		FreeComment    bool            `json:"freeComment"`
		AllowReaction  *bool           `json:"allowReaction"`
		NotifyMode     string          `json:"notifyMode"`
		NotifyAudience *audienceInput  `json:"notifyAudience"`
		RejectReason   string          `json:"rejectReason"`
		Actions        []action        `json:"actions"`
		Metadata       json.RawMessage `json:"metadata"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := localized.Load(ctx, tx)
	if err != nil {
		return err
	}
	var oldStatus string
	var oldPublished time.Time
	if !create {
		err := tx.QueryRowContext(ctx, `SELECT status, published_at FROM publication_news WHERE id=$1 FOR UPDATE`, id).Scan(&oldStatus, &oldPublished)
		if errors.Is(err, sql.ErrNoRows) {
			return notFound("NOT_FOUND")
		}
		if err != nil {
			return err
		}
	}
	status := strings.TrimSpace(in.Status)
	if status == "" {
		switch {
		case in.IsActive != nil && *in.IsActive:
			status = "published"
		case in.IsActive != nil:
			status = "draft"
		case create:
			status = "published"
		default:
			status = oldStatus
		}
	}
	if !containsStr(newsStatuses, status) {
		return badRequest("INVALID_STATUS", "status", "status must be draft, pending_review, published, archived or rejected")
	}
	format := strings.TrimSpace(in.BodyFormat)
	if format == "" {
		format = "markdown"
	}
	if format != "markdown" && format != "html" {
		return badRequest("INVALID_BODY_FORMAT", "bodyFormat", "bodyFormat must be markdown or html")
	}
	title, err := cleanI18n(langs, "title", in.Title, true, 200)
	if err != nil {
		return err
	}
	summary, err := cleanI18n(langs, "summary", in.Summary, false, 1000)
	if err != nil {
		return err
	}
	body, err := cleanNewsBody(langs, format, in.Body, true)
	if err != nil {
		return err
	}
	cover, err := cleanURL("coverUrl", in.CoverURL)
	if err != nil {
		return err
	}
	cat, err := cleanUUID("categoryId", in.CategoryID, false)
	if err != nil {
		return err
	}
	if err := checkCategory(ctx, tx, cat, "news"); err != nil {
		return err
	}
	tags, err := cleanTagIDs(ctx, tx, "tagIds", in.TagIDs, "news", "general")
	if err != nil {
		return err
	}
	published, err := cleanTime("publishedAt", in.PublishedAt)
	if err != nil {
		return err
	}
	if published == nil {
		now := time.Now().UTC()
		if !create {
			now = oldPublished
		}
		published = &now
	}
	expires, err := cleanTime("expiresAt", in.ExpiresAt)
	if err != nil {
		return err
	}
	if expires != nil && !expires.After(*published) {
		return badRequest("EXPIRES_BEFORE_PUBLISH", "expiresAt", "the end must be after the publish time")
	}
	mode := strings.TrimSpace(in.NotifyMode)
	if mode == "" {
		mode = "none"
	}
	if !containsStr(newsNotifyModes, mode) {
		return badRequest("INVALID_NOTIFY_MODE", "notifyMode", "notifyMode must be none, in_app, push or both")
	}
	audJSON := "{}"
	if in.NotifyAudience != nil {
		a, err := cleanAudience(ctx, tx, *in.NotifyAudience)
		if err != nil {
			return err
		}
		if !a.empty() {
			b, _ := json.Marshal(a.input())
			audJSON = string(b)
		}
	}
	if len(in.Actions) > 5 {
		return badRequest("TOO_MANY_ACTIONS", "actions", "at most 5 actions")
	}
	actions := make([]action, 0, len(in.Actions))
	for _, a := range in.Actions {
		clean, err := cleanAction(langs, "actions", a, true)
		if err != nil {
			return err
		}
		actions = append(actions, clean)
	}
	actionsJSON, _ := json.Marshal(actions)
	metadata, err := cleanObject("metadata", in.Metadata)
	if err != nil {
		return err
	}
	reason := strings.TrimSpace(in.RejectReason)
	if utf8.RuneCountInString(reason) > 500 {
		return badRequest("TEXT_TOO_LONG", "rejectReason", "reason is too long (max 500)")
	}
	if status != "rejected" {
		reason = ""
	}
	who := ctxConsoleUser(ctx)
	cols := []string{"title", "summary", "body", "body_format", "cover_url", "category_id", "tag_ids", "status", "is_active", "is_pinned", "published_at", "expires_at", "allow_comment", "free_comment",
		"allow_reaction", "notify_mode", "notify_audience", "reject_reason", "actions", "metadata", "updated_by"}
	vals := []any{i18nJSON(title), i18nJSON(summary), i18nJSON(body), format, cover, cat, "{" + strings.Join(tags, ",") + "}", status, status == "published", in.IsPinned, *published, expires,
		boolOr(in.AllowComment, true), in.FreeComment, boolOr(in.AllowReaction, true), mode, audJSON, reason, string(actionsJSON), metadata, who}
	if create {
		cols = append(cols, "created_by", "notified_at")
		var notified any
		vals = append(vals, who, notified)
	}
	if err := execSave(ctx, tx, "publication_news", id, create, cols, vals); err != nil {
		return err
	}
	if status == "published" && oldStatus != "published" {
		_, err := tx.ExecContext(ctx, `UPDATE publication_news SET approved_by=$2, approved_at=now() WHERE id=$1`, id, who)
		return err
	}
	return nil
}

func (s *service) consoleNewsReview(w http.ResponseWriter, r *http.Request) {
	var in struct {
		IDs    []string `json:"ids"`
		Action string   `json:"action"`
		Reason string   `json:"reason"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if len(in.IDs) == 0 {
		s.writeError(w, badRequest("IDS_REQUIRED", "ids", "pick at least one news"))
		return
	}
	if len(in.IDs) > 200 {
		s.writeError(w, badRequest("TOO_MANY_IDS", "ids", "at most 200 ids"))
		return
	}
	ids := make([]string, 0, len(in.IDs))
	for _, raw := range in.IDs {
		id, err := cleanUUID("ids", raw, true)
		if err != nil {
			s.writeError(w, err)
			return
		}
		ids = append(ids, *id)
	}
	reason := strings.TrimSpace(in.Reason)
	if utf8.RuneCountInString(reason) > 500 {
		s.writeError(w, badRequest("TEXT_TOO_LONG", "reason", "reason is too long (max 500)"))
		return
	}
	who := consoleUser(r)
	var res sql.Result
	var err error
	switch in.Action {
	case "approve":
		res, err = s.db.ExecContext(r.Context(), `UPDATE publication_news SET status='published', is_active=true, approved_by=$2, approved_at=now(), reject_reason='', updated_by=$2, updated_at=now()
			WHERE id = ANY($1::uuid[]) AND status IN ('pending_review','draft','rejected')`, uuidArray(ids), who)
	case "reject":
		res, err = s.db.ExecContext(r.Context(), `UPDATE publication_news SET status='rejected', is_active=false, reject_reason=$3, updated_by=$2, updated_at=now()
			WHERE id = ANY($1::uuid[]) AND status IN ('pending_review','draft')`, uuidArray(ids), who, reason)
	case "archive":
		res, err = s.db.ExecContext(r.Context(), `UPDATE publication_news SET status='archived', is_active=false, updated_by=$2, updated_at=now()
			WHERE id = ANY($1::uuid[]) AND status <> 'archived'`, uuidArray(ids), who)
	default:
		s.writeError(w, badRequest("INVALID_ACTION", "action", "action must be approve, reject or archive"))
		return
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	n, _ := res.RowsAffected()
	writeJSON(w, http.StatusOK, map[string]any{"changed": n, "skipped": int64(len(ids)) - n})
}

const reactionTypeJSON = `jsonb_build_object('id',t.id,'key',t.key,'icon',t.icon,'name',t.name,'sort',t.sort,'isActive',t.is_active,
	'uses',(SELECT count(*) FROM publication_news_reaction r WHERE r.reaction=t.key),'createdAt',t.created_at,'updatedAt',t.updated_at)`

func reactionTypeResource() resource {
	return resource{
		name:    "news-reactions",
		table:   "publication_news_reaction_type",
		json:    reactionTypeJSON,
		from:    ` FROM publication_news_reaction_type t`,
		search:  []string{"t.key", translatedSearch("t.name")},
		filters: map[string]string{"isActive": "t.is_active = ?::boolean"},
		order:   "t.sort, t.key, t.id",
		save:    saveReactionType,
		remove: func(s *service, ctx context.Context, tx *sql.Tx, id string) error {
			var used bool
			if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_news_reaction r JOIN publication_news_reaction_type t ON t.key=r.reaction WHERE t.id=$1)`, id).Scan(&used); err != nil {
				return err
			}
			if used {
				return conflict("IN_USE", "players used this reaction; switch it off instead")
			}
			return nil
		},
	}
}

func saveReactionType(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key      string `json:"key"`
		Icon     string `json:"icon"`
		Name     i18n   `json:"name"`
		Sort     int    `json:"sort"`
		IsActive *bool  `json:"isActive"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	langs, err := localized.Load(ctx, tx)
	if err != nil {
		return err
	}
	key, err := cleanBannerKey(ctx, tx, "publication_news_reaction_type", id, create, in.Key)
	if err != nil {
		return err
	}
	icon := strings.TrimSpace(in.Icon)
	if icon == "" {
		return badRequest("REQUIRED", "icon", "pick an emoji or a picture")
	}
	if strings.HasPrefix(icon, "/") || strings.HasPrefix(strings.ToLower(icon), "http") {
		if icon, err = cleanURL("icon", icon); err != nil {
			return err
		}
	} else if utf8.RuneCountInString(icon) > 16 || strings.ContainsAny(icon, "<>\"'`") {
		return badRequest("INVALID_ICON", "icon", "use one emoji (max 16 characters) or an image URL")
	}
	name, err := cleanI18n(langs, "name", in.Name, true, 40)
	if err != nil {
		return err
	}
	return execSave(ctx, tx, "publication_news_reaction_type", id, create, []string{"key", "icon", "name", "sort", "is_active"}, []any{key, icon, i18nJSON(name), in.Sort, boolOr(in.IsActive, true)})
}
