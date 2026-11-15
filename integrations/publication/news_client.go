package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const newsLiveSQL = `n.status = 'published' AND n.published_at <= now()`
const newsNotExpiredSQL = `(n.expires_at IS NULL OR n.expires_at > now())`

type newsRow struct {
	ID, Title, Summary, Body, Format, Cover, CategoryKey, Actions, SourceURL string
	Tags                                                                     []string
	Published                                                                time.Time
	Expires                                                                  *time.Time
	Pinned, AllowComment, FreeComment, AllowReaction, Seen                   bool
	Comments                                                                 int64
	Seo                                                                      seoRow
}

func (r newsRow) expired(now time.Time) bool {
	return r.Expires != nil && !r.Expires.After(now)
}

func newsRowCols(u string) string {
	return `n.id::text, n.title::text, n.summary::text, n.body::text, n.body_format, n.cover_url, COALESCE(c.key,''), n.actions::text, n.source_url, to_jsonb(n.tag_ids)::text,
	n.published_at, n.expires_at, n.is_pinned, n.allow_comment, n.free_comment, n.allow_reaction,
	EXISTS (SELECT 1 FROM publication_news_seen ns WHERE ns.news_id=n.id AND ns.user_id=` + u + `::uuid),
	(SELECT count(*) FROM publication_news_comment nc WHERE nc.news_id=n.id AND nc.is_accepted),
	n.slug, n.meta_title::text, n.meta_description::text, to_jsonb(n.keywords)::text, n.canonical_url, n.og_image, n.noindex`
}

func scanNewsRow(row interface{ Scan(...any) error }) (newsRow, error) {
	var r newsRow
	var tags, mt, md, kw string
	var exp sql.NullTime
	err := row.Scan(&r.ID, &r.Title, &r.Summary, &r.Body, &r.Format, &r.Cover, &r.CategoryKey, &r.Actions, &r.SourceURL, &tags, &r.Published, &exp, &r.Pinned, &r.AllowComment,
		&r.FreeComment, &r.AllowReaction, &r.Seen, &r.Comments, &r.Seo.Slug, &mt, &md, &kw, &r.Seo.Canonical, &r.Seo.OgImage, &r.Seo.Noindex)
	if err != nil {
		return r, err
	}
	r.Seo.MetaTitle, r.Seo.MetaDesc = parseI18n(mt), parseI18n(md)
	_ = json.Unmarshal([]byte(kw), &r.Seo.Keywords)
	_ = json.Unmarshal([]byte(tags), &r.Tags)
	if exp.Valid {
		t := exp.Time
		r.Expires = &t
	}
	return r, nil
}

func (s *service) visibleNews(ctx context.Context, q queryer, userID, id string, allowExpired bool) (newsRow, error) {
	var row newsRow
	if _, err := uuid.FromString(id); err != nil {
		return row, notFound("NEWS_NOT_FOUND")
	}
	return s.visibleNewsWhere(ctx, q, userID, "n.id=$1", id, allowExpired)
}

func (s *service) visibleNewsWhere(ctx context.Context, q queryer, userID, cond, arg string, allowExpired bool) (newsRow, error) {
	row, err := scanNewsRow(q.QueryRowContext(ctx, `SELECT `+newsRowCols("$2")+` FROM publication_news n LEFT JOIN base_category c ON c.id=n.category_id WHERE `+cond+` AND `+newsLiveSQL, arg, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return row, notFound("NEWS_NOT_FOUND")
	}
	if err != nil {
		return row, err
	}
	if row.expired(time.Now()) {
		if !allowExpired {
			return row, fail(http.StatusConflict, "NEWS_EXPIRED", "newsId", "this news has ended")
		}
		st, err := loadNewsSettings(ctx, q)
		if err != nil {
			return row, err
		}
		if !st.ShowExpired {
			return row, notFound("NEWS_NOT_FOUND")
		}
	}
	return row, nil
}

type newsExtras struct {
	tags      map[string]map[string]any
	reactions map[string]map[string]int64
	mine      map[string]string
	types     []map[string]any
}

func (s *service) loadNewsExtras(ctx context.Context, rows []newsRow, userID string, langs *localized.Set, lang string, withTags, withReactions bool) (*newsExtras, error) {
	ex := &newsExtras{tags: map[string]map[string]any{}, reactions: map[string]map[string]int64{}, mine: map[string]string{}}
	if withTags {
		var ids []string
		for _, r := range rows {
			ids = append(ids, r.Tags...)
		}
		if len(ids) > 0 {
			q, err := s.db.QueryContext(ctx, `SELECT id::text, key, name::text, icon_url FROM base_tag WHERE id = ANY($1::uuid[])`, uuidArray(ids))
			if err != nil {
				return nil, err
			}
			for q.Next() {
				var id, key, name, icon string
				if err := q.Scan(&id, &key, &name, &icon); err != nil {
					q.Close()
					return nil, err
				}
				t := map[string]any{"key": key, "name": langs.Resolve(parseI18n(name), lang)}
				if icon != "" {
					t["iconUrl"] = icon
				}
				ex.tags[id] = t
			}
			q.Close()
		}
	}
	if withReactions {
		types, err := s.reactionTypes(ctx, langs, lang)
		if err != nil {
			return nil, err
		}
		ex.types = types
		ids := make([]string, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		if len(ids) > 0 {
			q, err := s.db.QueryContext(ctx, `SELECT r.news_id::text, r.reaction, count(*), bool_or(r.user_id=$2::uuid) FROM publication_news_reaction r
				JOIN publication_news_reaction_type t ON t.key=r.reaction AND t.is_active WHERE r.news_id = ANY($1::uuid[]) GROUP BY r.news_id, r.reaction`, uuidArray(ids), userID)
			if err != nil {
				return nil, err
			}
			for q.Next() {
				var nid, key string
				var n int64
				var mine bool
				if err := q.Scan(&nid, &key, &n, &mine); err != nil {
					q.Close()
					return nil, err
				}
				if ex.reactions[nid] == nil {
					ex.reactions[nid] = map[string]int64{}
				}
				ex.reactions[nid][key] = n
				if mine {
					ex.mine[nid] = key
				}
			}
			q.Close()
		}
	}
	return ex, nil
}

func (ex *newsExtras) fill(item map[string]any, r newsRow, withTags, withReactions bool) {
	if withTags {
		tags := []map[string]any{}
		for _, id := range r.Tags {
			if t := ex.tags[id]; t != nil {
				tags = append(tags, t)
			}
		}
		item["tags"] = tags
	}
	if withReactions {
		counts := ex.reactions[r.ID]
		if counts == nil {
			counts = map[string]int64{}
		}
		var total int64
		for _, n := range counts {
			total += n
		}
		item["reactions"] = counts
		item["reactionCount"] = total
		item["myReaction"] = nullable(ex.mine[r.ID])
		item["allowReaction"] = r.AllowReaction
	}
}

func (s *service) reactionTypes(ctx context.Context, langs *localized.Set, lang string) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, icon, name::text FROM publication_news_reaction_type WHERE is_active ORDER BY sort, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var key, icon, name string
		if err := rows.Scan(&key, &icon, &name); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"key": key, "icon": icon, "name": langs.Resolve(parseI18n(name), lang)})
	}
	return out, rows.Err()
}

func boolDefault(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func cleanKeys(field string, in []string, max int) ([]string, error) {
	out := []string{}
	for _, k := range in {
		k = strings.ToLower(strings.TrimSpace(k))
		if k == "" {
			continue
		}
		if len(k) > 64 {
			return nil, badRequest("INVALID_KEY", field, "keys are at most 64 characters")
		}
		if !containsStr(out, k) {
			out = append(out, k)
		}
	}
	if len(out) > max {
		return nil, badRequest("TOO_MANY_KEYS", field, "at most "+strconv.Itoa(max)+" keys")
	}
	return out, nil
}

type newsFilter struct {
	CategoryKey string   `json:"categoryKey"`
	TagKey      string   `json:"tagKey"`
	TagKeys     []string `json:"tagKeys"`
	TagMode     string   `json:"tagMode"`
}

func (f newsFilter) apply(where *sqlWhere) error {
	if k := strings.TrimSpace(f.CategoryKey); k != "" {
		where.add("n.category_id IN (SELECT id FROM base_category WHERE key_type='news' AND key = ?)", k)
	}
	keys := f.TagKeys
	if strings.TrimSpace(f.TagKey) != "" {
		keys = append(keys, f.TagKey)
	}
	keys, err := cleanKeys("tagKeys", keys, 20)
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		switch f.TagMode {
		case "", "any":
			where.add("n.tag_ids && ARRAY(SELECT id::text FROM base_tag WHERE key_type IN ('news','general') AND lower(key) = ANY(?::text[]))", "{"+quoteArray(keys)+"}")
		case "all":
			for _, k := range keys {
				where.add("EXISTS (SELECT 1 FROM base_tag bt WHERE bt.key_type IN ('news','general') AND lower(bt.key) = ? AND bt.id::text = ANY(n.tag_ids))", k)
			}
		default:
			return badRequest("INVALID_TAG_MODE", "tagMode", "tagMode must be any or all")
		}
	}
	return nil
}

func (s *service) rpcNewsList(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		newsFilter
		Limit            int    `json:"limit"`
		Cursor           string `json:"cursor"`
		Lang             string `json:"lang"`
		Expired          string `json:"expired"`
		UnreadOnly       bool   `json:"unreadOnly"`
		IncludeTags      *bool  `json:"includeTags"`
		IncludeReactions *bool  `json:"includeReactions"`
		IncludeSeo       bool   `json:"includeSeo"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := s.clientLang(ctx, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	st, err := loadNewsSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	limit := clampLimit(in.Limit, 20, 50)
	offset, err := readCursor(in.Cursor)
	if err != nil {
		return nil, err
	}
	where := &sqlWhere{args: []any{userID}}
	where.add(newsLiveSQL)
	expired := strings.TrimSpace(in.Expired)
	switch expired {
	case "", "exclude":
		expired = "exclude"
	case "include", "only":
	default:
		return nil, badRequest("INVALID_EXPIRED", "expired", "expired must be exclude, include or only")
	}
	if !st.ShowExpired {
		expired = "exclude"
	}
	switch expired {
	case "exclude":
		where.add(newsNotExpiredSQL)
	case "only":
		where.add("NOT " + newsNotExpiredSQL)
	}
	if err := in.newsFilter.apply(where); err != nil {
		return nil, err
	}
	if in.UnreadOnly {
		where.add("NOT EXISTS (SELECT 1 FROM publication_news_seen ns WHERE ns.news_id=n.id AND ns.user_id=$1::uuid)")
	}
	n := len(where.args)
	args := append(append([]any{}, where.args...), limit, offset)
	rows, err := s.db.QueryContext(ctx, `SELECT `+newsRowCols("$1")+` FROM publication_news n LEFT JOIN base_category c ON c.id=n.category_id`+where.sql()+`
		ORDER BY n.is_pinned DESC, n.published_at DESC, n.id DESC LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2), args...)
	if err != nil {
		return nil, err
	}
	var list []newsRow
	for rows.Next() {
		r, err := scanNewsRow(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		list = append(list, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	withTags, withReactions := boolDefault(in.IncludeTags, true), boolDefault(in.IncludeReactions, true)
	ex, err := s.loadNewsExtras(ctx, list, userID, langs, lang, withTags, withReactions)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	items := []map[string]any{}
	for _, r := range list {
		item := map[string]any{"id": r.ID, "title": langs.Resolve(parseI18n(r.Title), lang), "summary": langs.Resolve(parseI18n(r.Summary), lang), "coverUrl": r.Cover,
			"publishedAt": r.Published.UTC(), "expiresAt": r.Expires, "expired": r.expired(now), "isPinned": r.Pinned, "categoryKey": nullable(r.CategoryKey), "seen": r.Seen,
			"commentCount": r.Comments, "allowComment": r.AllowComment, "sourceUrl": nullable(r.SourceURL), "slug": r.Seo.Slug}
		if in.IncludeSeo {
			title, summary := item["title"].(string), item["summary"].(string)
			item["seo"] = seoView(langs, lang, r.Seo, title, summary, newsBodyText(r.Format, langs.Resolve(parseI18n(r.Body), lang)), r.Cover)
		}
		ex.fill(item, r, withTags, withReactions)
		items = append(items, item)
	}
	out := map[string]any{"lang": lang, "items": items, "cursor": nextCursor(offset, limit, len(items))}
	if withReactions {
		out["reactionTypes"] = ex.types
	}
	return out, nil
}

func (s *service) rpcNewsGet(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := s.clientLang(ctx, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	var row newsRow
	if strings.TrimSpace(in.ID) == "" && strings.TrimSpace(in.Slug) != "" {
		slug, _ := cleanSlug(in.Slug)
		if slug == "" {
			return nil, notFound("NEWS_NOT_FOUND")
		}
		row, err = s.visibleNewsWhere(ctx, s.db, userID, "n.slug=$1", slug, true)
	} else {
		row, err = s.visibleNews(ctx, s.db, userID, in.ID, true)
	}
	if err != nil {
		return nil, err
	}
	var actions []action
	_ = json.Unmarshal([]byte(row.Actions), &actions)
	outActions := []map[string]any{}
	for _, a := range actions {
		outActions = append(outActions, map[string]any{"type": a.Type, "label": langs.Resolve(a.Label, lang), "value": a.Value})
	}
	body := langs.Resolve(parseI18n(row.Body), lang)
	ex, err := s.loadNewsExtras(ctx, []newsRow{row}, userID, langs, lang, true, true)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"lang": lang, "id": row.ID, "title": langs.Resolve(parseI18n(row.Title), lang), "summary": langs.Resolve(parseI18n(row.Summary), lang), "body": body,
		"bodyFormat": row.Format, "bodyHtml": newsBodyHTML(row.Format, body), "coverUrl": row.Cover, "categoryKey": nullable(row.CategoryKey), "publishedAt": row.Published.UTC(),
		"expiresAt": row.Expires, "expired": row.expired(time.Now()), "isPinned": row.Pinned, "allowComment": row.AllowComment, "actions": outActions, "seen": row.Seen,
		"commentCount": row.Comments, "sourceUrl": nullable(row.SourceURL), "reactionTypes": ex.types, "slug": row.Seo.Slug}
	out["seo"] = seoView(langs, lang, row.Seo, out["title"].(string), out["summary"].(string), newsBodyText(row.Format, body), row.Cover)
	ex.fill(out, row, true, true)
	return out, nil
}

func (s *service) newsUnread(ctx context.Context, userID string, f newsFilter) (int64, error) {
	st, err := loadNewsSettings(ctx, s.db)
	if err != nil {
		return 0, err
	}
	where := &sqlWhere{args: []any{userID}}
	where.add(newsLiveSQL)
	where.add(newsNotExpiredSQL)
	where.add("NOT EXISTS (SELECT 1 FROM publication_news_seen ns WHERE ns.news_id=n.id AND ns.user_id=$1::uuid)")
	if st.UnreadScope == "signup" {
		where.add("n.published_at >= (SELECT create_time FROM users WHERE id=$1::uuid)")
	}
	if err := f.apply(where); err != nil {
		return 0, err
	}
	var n int64
	err = s.db.QueryRowContext(ctx, `SELECT count(*) FROM publication_news n`+where.sql(), where.args...).Scan(&n)
	return n, err
}

func (s *service) rpcNewsUnread(ctx context.Context, userID, payload string) (any, error) {
	var in newsFilter
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	n, err := s.newsUnread(ctx, userID, in)
	if err != nil {
		return nil, err
	}
	return map[string]any{"unread": n}, nil
}

func (s *service) rpcNewsSeen(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID  string   `json:"id"`
		IDs []string `json:"ids"`
		All bool     `json:"all"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	switch {
	case in.All:
		if _, err := s.db.ExecContext(ctx, `INSERT INTO publication_news_seen (news_id, user_id) SELECT n.id, $1 FROM publication_news n WHERE `+newsLiveSQL+` ON CONFLICT DO NOTHING`, userID); err != nil {
			return nil, err
		}
	case in.ID != "" || len(in.IDs) > 0:
		ids := in.IDs
		if in.ID != "" {
			ids = append(ids, in.ID)
		}
		if len(ids) > 200 {
			return nil, badRequest("TOO_MANY_IDS", "ids", "at most 200 ids")
		}
		clean := make([]string, 0, len(ids))
		for _, raw := range ids {
			id, err := cleanUUID("ids", raw, true)
			if err != nil {
				return nil, notFound("NEWS_NOT_FOUND")
			}
			clean = append(clean, *id)
		}
		res, err := s.db.ExecContext(ctx, `INSERT INTO publication_news_seen (news_id, user_id) SELECT n.id, $2 FROM publication_news n WHERE n.id = ANY($1::uuid[]) AND `+newsLiveSQL+` ON CONFLICT DO NOTHING`,
			uuidArray(clean), userID)
		if err != nil {
			return nil, err
		}
		if in.ID != "" && len(in.IDs) == 0 {
			if n, _ := res.RowsAffected(); n == 0 {
				if _, err := s.visibleNews(ctx, s.db, userID, in.ID, true); err != nil {
					return nil, err
				}
			}
		}
	default:
		return nil, badRequest("IDS_REQUIRED", "id", "pass id, ids or all:true")
	}
	n, err := s.newsUnread(ctx, userID, newsFilter{})
	if err != nil {
		return nil, err
	}
	return map[string]any{"unread": n}, nil
}

func (s *service) rpcNewsReactionTypes(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := s.clientLang(ctx, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	types, err := s.reactionTypes(ctx, langs, lang)
	if err != nil {
		return nil, err
	}
	return map[string]any{"lang": lang, "items": types}, nil
}

func (s *service) rpcNewsReact(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		NewsID   string  `json:"newsId"`
		Reaction *string `json:"reaction"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	st, err := loadNewsSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	d, err := clusterkit.Allow(ctx, s.db, "news-react:"+userID, int64(st.ReactionsPerMinute), time.Minute)
	if err != nil {
		return nil, err
	}
	if !d.Allowed {
		return nil, fail(http.StatusTooManyRequests, "REACTION_RATE_LIMITED", "", "too many reactions; wait "+strconv.Itoa(int(math.Ceil(d.RetryAfter.Seconds())))+" seconds")
	}
	key := ""
	if in.Reaction != nil {
		key = strings.TrimSpace(*in.Reaction)
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		row, err := s.visibleNews(ctx, tx, userID, in.NewsID, false)
		if err != nil {
			return err
		}
		if !row.AllowReaction {
			return badRequest("REACTION_NOT_ALLOWED", "newsId", "reactions are closed on this news")
		}
		if key == "" {
			_, err := tx.ExecContext(ctx, `DELETE FROM publication_news_reaction WHERE news_id=$1 AND user_id=$2`, row.ID, userID)
			return err
		}
		var ok bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM publication_news_reaction_type WHERE key=$1 AND is_active)`, key).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return badRequest("REACTION_NOT_FOUND", "reaction", "this reaction does not exist")
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO publication_news_reaction (news_id, user_id, reaction) VALUES ($1,$2,$3)
			ON CONFLICT (news_id, user_id) DO UPDATE SET reaction=EXCLUDED.reaction, updated_at=now()`, row.ID, userID, key)
		return err
	})
	if err != nil {
		return nil, err
	}
	langs, lang, err := s.clientLang(ctx, userID, "")
	if err != nil {
		return nil, err
	}
	ex, err := s.loadNewsExtras(ctx, []newsRow{{ID: in.NewsID, AllowReaction: true}}, userID, langs, lang, false, true)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"newsId": in.NewsID}
	ex.fill(out, newsRow{ID: in.NewsID, AllowReaction: true}, false, true)
	delete(out, "allowReaction")
	return out, nil
}

func (s *service) rpcNewsCommentDelete(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID string `json:"id"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if _, err := uuid.FromString(in.ID); err != nil {
		return nil, notFound("COMMENT_NOT_FOUND")
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM publication_news_comment WHERE id=$1 AND user_id=$2`, in.ID, userID)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, notFound("COMMENT_NOT_FOUND")
	}
	return map[string]any{"deleted": true}, nil
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
