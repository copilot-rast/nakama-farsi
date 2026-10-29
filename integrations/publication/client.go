package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
)

const commentsPerMinute = 5

func (s *service) clientRPCs() map[string]clientRPC {
	return map[string]clientRPC{
		"publication_faq_list":      s.rpcFaqList,
		"publication_news_list":     s.rpcNewsList,
		"publication_news_get":      s.rpcNewsGet,
		"publication_news_seen":     s.rpcNewsSeen,
		"publication_news_comments": s.rpcNewsComments,
		"publication_news_comment":  s.rpcNewsComment,
		"publication_inbox":         s.rpcInbox,
		"publication_inbox_seen":    s.rpcInboxSeen,
	}
}

func readCursor(cursor string) (int, error) {
	cursor = strings.TrimSpace(cursor)
	if cursor == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(cursor)
	if err != nil || n < 0 || n > 1000000 {
		return 0, badRequest("INVALID_CURSOR", "cursor", "cursor is not valid")
	}
	return n, nil
}

func clampLimit(n, def, max int) int {
	if n <= 0 {
		return def
	}
	if n > max {
		return max
	}
	return n
}

func nextCursor(offset, limit, got int) string {
	if got < limit {
		return ""
	}
	return strconv.Itoa(offset + limit)
}

func parseI18n(raw string) i18n {
	out := i18n{}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func (s *service) rpcFaqList(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		CategoryKey string `json:"categoryKey"`
		Q           string `json:"q"`
		Lang        string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	lang := cleanLang(in.Lang)
	categories := []map[string]any{}
	rows, err := s.db.QueryContext(ctx, `SELECT c.key, c.name::text FROM economy_category c WHERE c.kind='faq'
		AND EXISTS (SELECT 1 FROM publication_faq f WHERE f.category_id=c.id AND f.is_active) ORDER BY c.sort, c.key`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var key, name string
		if err := rows.Scan(&key, &name); err != nil {
			rows.Close()
			return nil, err
		}
		categories = append(categories, map[string]any{"key": key, "name": parseI18n(name).pick(lang)})
	}
	rows.Close()
	where := &sqlWhere{}
	where.add("f.is_active")
	if k := strings.TrimSpace(in.CategoryKey); k != "" {
		where.add("c.key = ?", k)
	}
	if q := strings.TrimSpace(in.Q); q != "" {
		if len([]rune(q)) > 100 {
			return nil, badRequest("QUERY_TOO_LONG", "q", "search text is too long")
		}
		n := where.next()
		where.args = append(where.args, likePattern(q))
		where.parts = append(where.parts, "(f.question->>'fa' ILIKE "+n+" OR COALESCE(f.question->>'en','') ILIKE "+n+" OR f.answer->>'fa' ILIKE "+n+" OR COALESCE(f.answer->>'en','') ILIKE "+n+")")
	}
	rows, err = s.db.QueryContext(ctx, `SELECT f.id::text, f.question::text, f.answer::text, COALESCE(c.key,'') FROM publication_faq f LEFT JOIN economy_category c ON c.id=f.category_id`+
		where.sql()+` ORDER BY f.sort, f.created_at DESC, f.id LIMIT 500`, where.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, q, a, key string
		if err := rows.Scan(&id, &q, &a, &key); err != nil {
			return nil, err
		}
		answer := parseI18n(a).pick(lang)
		var cat any
		if key != "" {
			cat = key
		}
		items = append(items, map[string]any{"id": id, "question": parseI18n(q).pick(lang), "answer": answer, "answerHtml": renderMarkdown(answer), "categoryKey": cat})
	}
	return map[string]any{"categories": categories, "items": items}, rows.Err()
}

const newsVisible = `n.is_active AND n.published_at <= now()`

func (s *service) rpcNewsList(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Limit       int    `json:"limit"`
		Cursor      string `json:"cursor"`
		CategoryKey string `json:"categoryKey"`
		Lang        string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	lang := cleanLang(in.Lang)
	limit := clampLimit(in.Limit, 20, 50)
	offset, err := readCursor(in.Cursor)
	if err != nil {
		return nil, err
	}
	where := &sqlWhere{args: []any{userID}}
	where.add(newsVisible)
	if k := strings.TrimSpace(in.CategoryKey); k != "" {
		where.add("n.category_id IN (SELECT id FROM economy_category WHERE kind='news' AND key = ?)", k)
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(ctx, `SELECT n.id::text, n.title::text, n.summary::text, n.cover_url, n.published_at, n.is_pinned, COALESCE(c.key,''),
		EXISTS (SELECT 1 FROM publication_news_seen ns WHERE ns.news_id=n.id AND ns.user_id=$1::uuid),
		(SELECT count(*) FROM publication_news_comment nc WHERE nc.news_id=n.id AND nc.is_accepted)
		FROM publication_news n LEFT JOIN economy_category c ON c.id=n.category_id`+where.sql()+`
		ORDER BY n.is_pinned DESC, n.published_at DESC, n.id DESC LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2), append(where.args, limit, offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, title, summary, cover, key string
		var published time.Time
		var pinned, seen bool
		var comments int64
		if err := rows.Scan(&id, &title, &summary, &cover, &published, &pinned, &key, &seen, &comments); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"id": id, "title": parseI18n(title).pick(lang), "summary": parseI18n(summary).pick(lang), "coverUrl": cover,
			"publishedAt": published.UTC(), "isPinned": pinned, "categoryKey": nullable(key), "seen": seen, "commentCount": comments})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return map[string]any{"items": items, "cursor": nextCursor(offset, limit, len(items))}, nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

type newsRow struct {
	ID, Title, Summary, Body, Cover, CategoryKey, Actions string
	Published                                             time.Time
	Pinned, AllowComment, FreeComment, Seen               bool
	Comments                                              int64
}

func (s *service) visibleNews(ctx context.Context, q queryer, userID, id string) (newsRow, error) {
	var row newsRow
	if _, err := uuid.FromString(id); err != nil {
		return row, notFound("NEWS_NOT_FOUND")
	}
	err := q.QueryRowContext(ctx, `SELECT n.id::text, n.title::text, n.summary::text, n.body::text, n.cover_url, COALESCE(c.key,''), n.actions::text, n.published_at,
		n.is_pinned, n.allow_comment, n.free_comment,
		EXISTS (SELECT 1 FROM publication_news_seen ns WHERE ns.news_id=n.id AND ns.user_id=$2::uuid),
		(SELECT count(*) FROM publication_news_comment nc WHERE nc.news_id=n.id AND nc.is_accepted)
		FROM publication_news n LEFT JOIN economy_category c ON c.id=n.category_id WHERE n.id=$1 AND `+newsVisible, id, userID).
		Scan(&row.ID, &row.Title, &row.Summary, &row.Body, &row.Cover, &row.CategoryKey, &row.Actions, &row.Published, &row.Pinned, &row.AllowComment, &row.FreeComment, &row.Seen, &row.Comments)
	if errors.Is(err, sql.ErrNoRows) {
		return row, notFound("NEWS_NOT_FOUND")
	}
	return row, err
}

func (s *service) rpcNewsGet(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID   string `json:"id"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	lang := cleanLang(in.Lang)
	row, err := s.visibleNews(ctx, s.db, userID, in.ID)
	if err != nil {
		return nil, err
	}
	var actions []action
	_ = json.Unmarshal([]byte(row.Actions), &actions)
	outActions := []map[string]any{}
	for _, a := range actions {
		outActions = append(outActions, map[string]any{"type": a.Type, "label": a.Label.pick(lang), "value": a.Value})
	}
	body := parseI18n(row.Body).pick(lang)
	return map[string]any{"id": row.ID, "title": parseI18n(row.Title).pick(lang), "summary": parseI18n(row.Summary).pick(lang), "body": body, "bodyHtml": renderMarkdown(body),
		"coverUrl": row.Cover, "categoryKey": nullable(row.CategoryKey), "publishedAt": row.Published.UTC(), "isPinned": row.Pinned, "allowComment": row.AllowComment,
		"actions": outActions, "seen": row.Seen, "commentCount": row.Comments}, nil
}

func (s *service) rpcNewsSeen(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID string `json:"id"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if _, err := s.visibleNews(ctx, s.db, userID, in.ID); err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO publication_news_seen (news_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, in.ID, userID); err != nil {
		return nil, err
	}
	return map[string]any{}, nil
}

const commentClientJSON = `jsonb_build_object('id',m.id,'userId',m.user_id,'username',u.username,'displayName',u.display_name,'avatarUrl',u.avatar_url,'message',m.message,
	'parentId',m.parent_id,'isAccepted',m.is_accepted,'mine',m.user_id=$2::uuid,'createdAt',m.created_at)`

func (s *service) rpcNewsComments(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		NewsID string `json:"newsId"`
		Limit  int    `json:"limit"`
		Cursor string `json:"cursor"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if _, err := s.visibleNews(ctx, s.db, userID, in.NewsID); err != nil {
		return nil, err
	}
	limit := clampLimit(in.Limit, 20, 50)
	offset, err := readCursor(in.Cursor)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+commentClientJSON+`::text FROM publication_news_comment m JOIN users u ON u.id=m.user_id
		WHERE m.news_id=$1 AND m.parent_id IS NULL AND (m.is_accepted OR m.user_id=$2::uuid)
		ORDER BY m.created_at DESC, m.id DESC LIMIT $3 OFFSET $4`, in.NewsID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	roots := []map[string]any{}
	ids := []string{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			rows.Close()
			return nil, err
		}
		item := map[string]any{}
		_ = json.Unmarshal([]byte(raw), &item)
		item["replies"] = []map[string]any{}
		roots = append(roots, item)
		ids = append(ids, item["id"].(string))
	}
	rows.Close()
	if len(ids) > 0 {
		rows, err := s.db.QueryContext(ctx, `SELECT `+commentClientJSON+`::text FROM publication_news_comment m JOIN users u ON u.id=m.user_id
			WHERE m.news_id=$1 AND m.parent_id = ANY($3::uuid[]) AND (m.is_accepted OR m.user_id=$2::uuid)
			ORDER BY m.created_at, m.id LIMIT 2000`, in.NewsID, userID, "{"+strings.Join(ids, ",")+"}")
		if err != nil {
			return nil, err
		}
		byID := map[string]map[string]any{}
		for _, r := range roots {
			byID[r["id"].(string)] = r
		}
		for rows.Next() {
			var raw string
			if err := rows.Scan(&raw); err != nil {
				rows.Close()
				return nil, err
			}
			item := map[string]any{}
			_ = json.Unmarshal([]byte(raw), &item)
			parent, _ := item["parentId"].(string)
			if root := byID[parent]; root != nil {
				root["replies"] = append(root["replies"].([]map[string]any), item)
			}
		}
		rows.Close()
	}
	return map[string]any{"items": roots, "cursor": nextCursor(offset, limit, len(roots))}, nil
}

func (s *service) rpcNewsComment(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		NewsID   string `json:"newsId"`
		Message  string `json:"message"`
		ParentID string `json:"parentId"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	msg, err := cleanMessage(in.Message)
	if err != nil {
		return nil, err
	}
	parentID, err := cleanUUID("parentId", in.ParentID, false)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		row, err := s.visibleNews(ctx, tx, userID, in.NewsID)
		if err != nil {
			return err
		}
		if !row.AllowComment {
			return badRequest("COMMENT_NOT_ALLOWED", "newsId", "comments are closed on this news")
		}
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('publication-comment:' || $1))`, userID); err != nil {
			return err
		}
		var recent int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM publication_news_comment WHERE user_id=$1 AND created_at > now() - interval '1 minute'`, userID).Scan(&recent); err != nil {
			return err
		}
		if recent >= commentsPerMinute {
			return fail(http.StatusTooManyRequests, "COMMENT_RATE_LIMITED", "", "too many comments; wait a minute")
		}
		var parent *string
		if parentID != nil {
			var root sql.NullString
			var accepted bool
			var owner string
			err := tx.QueryRowContext(ctx, `SELECT parent_id::text, is_accepted, user_id::text FROM publication_news_comment WHERE id=$1 AND news_id=$2`, *parentID, row.ID).Scan(&root, &accepted, &owner)
			if errors.Is(err, sql.ErrNoRows) || (err == nil && !accepted && owner != userID) {
				return badRequest("PARENT_COMMENT_NOT_FOUND", "parentId", "parent comment not found on this news")
			}
			if err != nil {
				return err
			}
			p := *parentID
			if root.Valid {
				p = root.String
			}
			parent = &p
		}
		id := uuid.Must(uuid.NewV4()).String()
		if _, err := tx.ExecContext(ctx, `INSERT INTO publication_news_comment (id, news_id, user_id, parent_id, message, is_accepted) VALUES ($1,$2,$3,$4,$5,$6)`,
			id, row.ID, userID, parent, msg, row.FreeComment); err != nil {
			return err
		}
		var raw string
		if err := tx.QueryRowContext(ctx, `SELECT `+commentClientJSON+`::text FROM publication_news_comment m JOIN users u ON u.id=m.user_id WHERE m.id=$1`, id, userID).Scan(&raw); err != nil {
			return err
		}
		out = map[string]any{}
		_ = json.Unmarshal([]byte(raw), &out)
		out["replies"] = []map[string]any{}
		return nil
	})
	return out, err
}

func (s *service) unreadCount(ctx context.Context, userID string) (int64, error) {
	var n int64
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM publication_notification_recipient r JOIN publication_notification n ON n.id=r.notification_id
		WHERE r.user_id=$1 AND n.channel='in_app' AND r.status='sent' AND r.seen_at IS NULL`, userID).Scan(&n)
	return n, err
}

func (s *service) rpcInbox(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Limit  int    `json:"limit"`
		Cursor string `json:"cursor"`
		Lang   string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	lang := cleanLang(in.Lang)
	limit := clampLimit(in.Limit, 20, 50)
	offset, err := readCursor(in.Cursor)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT n.id::text, n.title::text, n.body::text, COALESCE(n.action::text,'null'), n.created_at, r.seen_at IS NOT NULL
		FROM publication_notification_recipient r JOIN publication_notification n ON n.id=r.notification_id
		WHERE r.user_id=$1 AND n.channel='in_app' AND r.status='sent'
		ORDER BY n.created_at DESC, n.id DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, title, body, act string
		var created time.Time
		var seen bool
		if err := rows.Scan(&id, &title, &body, &act, &created, &seen); err != nil {
			return nil, err
		}
		text := parseI18n(body).pick(lang)
		items = append(items, map[string]any{"id": id, "title": parseI18n(title).pick(lang), "body": text, "bodyHtml": renderMarkdown(text),
			"action": json.RawMessage(act), "createdAt": created.UTC(), "seen": seen})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	unread, err := s.unreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items": items, "unread": unread, "cursor": nextCursor(offset, limit, len(items))}, nil
}

func (s *service) rpcInboxSeen(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		IDs []string `json:"ids"`
		All bool     `json:"all"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	switch {
	case in.All:
		if _, err := s.db.ExecContext(ctx, `UPDATE publication_notification_recipient r SET seen_at=now() FROM publication_notification n
			WHERE n.id=r.notification_id AND n.channel='in_app' AND r.user_id=$1 AND r.status='sent' AND r.seen_at IS NULL`, userID); err != nil {
			return nil, err
		}
	case len(in.IDs) > 0:
		if len(in.IDs) > 200 {
			return nil, badRequest("TOO_MANY_IDS", "ids", "at most 200 ids")
		}
		ids := make([]string, 0, len(in.IDs))
		for _, raw := range in.IDs {
			id, err := cleanUUID("ids", raw, true)
			if err != nil {
				return nil, err
			}
			ids = append(ids, *id)
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE publication_notification_recipient SET seen_at=now()
			WHERE user_id=$1 AND notification_id = ANY($2::uuid[]) AND status='sent' AND seen_at IS NULL`, userID, "{"+strings.Join(ids, ",")+"}"); err != nil {
			return nil, err
		}
	default:
		return nil, badRequest("IDS_REQUIRED", "ids", "pass ids or all:true")
	}
	unread, err := s.unreadCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"unread": unread}, nil
}
