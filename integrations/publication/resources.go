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

	"github.com/gofrs/uuid/v5"
)

type resource struct {
	name    string
	table   string
	json    string
	from    string
	search  []string
	filters map[string]string
	order   string
	save    func(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error
}

const faqJSON = `jsonb_build_object('id',t.id,'categoryId',t.category_id,'categoryKey',c.key,'categoryName',c.name,'question',t.question,'answer',t.answer,'isActive',t.is_active,'sort',t.sort,'createdAt',t.created_at,'updatedAt',t.updated_at)`

const newsJSON = `jsonb_build_object('id',t.id,'title',t.title,'summary',t.summary,'body',t.body,'coverUrl',t.cover_url,'categoryId',t.category_id,'categoryKey',c.key,'categoryName',c.name,
	'isActive',t.is_active,'isPinned',t.is_pinned,'publishedAt',t.published_at,'allowComment',t.allow_comment,'freeComment',t.free_comment,'actions',t.actions,'metadata',t.metadata,
	'commentCount',(SELECT count(*) FROM publication_news_comment nc WHERE nc.news_id=t.id),
	'pendingComments',(SELECT count(*) FROM publication_news_comment nc WHERE nc.news_id=t.id AND NOT nc.is_accepted),
	'seenCount',(SELECT count(*) FROM publication_news_seen ns WHERE ns.news_id=t.id),
	'createdAt',t.created_at,'updatedAt',t.updated_at)`

func resources() []resource {
	return []resource{
		{
			name:    "faqs",
			table:   "publication_faq",
			json:    faqJSON,
			from:    ` FROM publication_faq t LEFT JOIN economy_category c ON c.id=t.category_id`,
			search:  []string{"t.question->>'fa'", "t.question->>'en'", "t.answer->>'fa'", "t.answer->>'en'"},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "categoryId": "t.category_id = ?::uuid"},
			order:   "t.sort, t.created_at DESC, t.id",
			save:    saveFaq,
		},
		{
			name:    "news",
			table:   "publication_news",
			json:    newsJSON,
			from:    ` FROM publication_news t LEFT JOIN economy_category c ON c.id=t.category_id`,
			search:  []string{"t.title->>'fa'", "t.title->>'en'", "t.summary->>'fa'", "t.summary->>'en'"},
			filters: map[string]string{"isActive": "t.is_active = ?::boolean", "categoryId": "t.category_id = ?::uuid", "isPinned": "t.is_pinned = ?::boolean"},
			order:   "t.is_pinned DESC, t.published_at DESC, t.id DESC",
			save:    saveNews,
		},
	}
}

func (s *service) resourceList(res resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := readPage(r)
		where := &sqlWhere{}
		if p.Q != "" {
			parts := make([]string, 0, len(res.search))
			n := where.next()
			for _, col := range res.search {
				parts = append(parts, "COALESCE("+col+",'') ILIKE "+n)
			}
			where.args = append(where.args, likePattern(p.Q))
			where.parts = append(where.parts, "("+strings.Join(parts, " OR ")+")")
		}
		for param, expr := range res.filters {
			if v := strings.TrimSpace(r.URL.Query().Get(param)); v != "" {
				if strings.Contains(expr, "boolean") && v != "true" && v != "false" {
					s.writeError(w, badRequest("INVALID_FILTER", param, "must be true or false"))
					return
				}
				if strings.Contains(expr, "uuid") {
					if _, err := uuid.FromString(v); err != nil {
						s.writeError(w, badRequest("INVALID_FILTER", param, "must be a UUID"))
						return
					}
				}
				where.add(expr, v)
			}
		}
		var total int64
		if err := s.db.QueryRowContext(r.Context(), `SELECT count(*)`+res.from+where.sql(), where.args...).Scan(&total); err != nil {
			s.writeError(w, err)
			return
		}
		n := len(where.args)
		rows, err := s.db.QueryContext(r.Context(), `SELECT `+res.json+`::text`+res.from+where.sql()+` ORDER BY `+res.order+` LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2),
			append(append([]any{}, where.args...), p.Limit, p.Offset)...)
		if err != nil {
			s.writeError(w, err)
			return
		}
		items, err := rawRows(rows)
		if err != nil {
			s.writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": p.Page, "limit": p.Limit})
	}
}

func rawRows(rows *sql.Rows) ([]json.RawMessage, error) {
	defer rows.Close()
	items := []json.RawMessage{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		items = append(items, json.RawMessage(raw))
	}
	return items, rows.Err()
}

type queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (s *service) fetchResource(ctx context.Context, q queryer, res resource, id string) (json.RawMessage, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+res.json+`::text`+res.from+` WHERE t.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func (s *service) resourceGet(res resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, err := s.fetchResource(r.Context(), s.db, res, pathVar(r, "id"))
		if err != nil {
			s.writeError(w, err)
			return
		}
		writeRaw(w, http.StatusOK, raw)
	}
}

func (s *service) resourceWrite(res resource, create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body json.RawMessage
		if err := readBody(r, &body); err != nil {
			s.writeError(w, err)
			return
		}
		id := pathVar(r, "id")
		if create {
			id = uuid.Must(uuid.NewV4()).String()
		} else if _, err := uuid.FromString(id); err != nil {
			s.writeError(w, notFound("NOT_FOUND"))
			return
		}
		var out json.RawMessage
		err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			if err := res.save(s, r.Context(), tx, id, create, body); err != nil {
				return err
			}
			var err error
			out, err = s.fetchResource(r.Context(), tx, res, id)
			return err
		})
		if err != nil {
			if isForeignKeyViolation(err) {
				err = badRequest("REFERENCE_NOT_FOUND", "", "a referenced record does not exist")
			}
			s.writeError(w, err)
			return
		}
		status := http.StatusOK
		if create {
			status = http.StatusCreated
		}
		writeRaw(w, status, out)
	}
}

func (s *service) resourceDelete(res resource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := pathVar(r, "id")
		if _, err := uuid.FromString(id); err != nil {
			s.writeError(w, notFound("NOT_FOUND"))
			return
		}
		result, err := s.db.ExecContext(r.Context(), `DELETE FROM `+res.table+` WHERE id=$1`, id)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if n, _ := result.RowsAffected(); n == 0 {
			s.writeError(w, notFound("NOT_FOUND"))
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{})
	}
}

func (s *service) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func execSave(ctx context.Context, tx *sql.Tx, table, id string, create bool, cols []string, vals []any) error {
	if create {
		ph := make([]string, len(cols))
		for i := range cols {
			ph[i] = "$" + itoa(i+2)
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO `+table+` (id,`+strings.Join(cols, ",")+`) VALUES ($1,`+strings.Join(ph, ",")+`)`, append([]any{id}, vals...)...)
		return err
	}
	sets := make([]string, len(cols))
	for i, c := range cols {
		sets[i] = c + "=$" + itoa(i+2)
	}
	res, err := tx.ExecContext(ctx, `UPDATE `+table+` SET `+strings.Join(sets, ",")+`, updated_at=now() WHERE id=$1`, append([]any{id}, vals...)...)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return notFound("NOT_FOUND")
	}
	return nil
}

func checkCategory(ctx context.Context, tx *sql.Tx, id *string, kind string) error {
	if id == nil {
		return nil
	}
	var k string
	err := tx.QueryRowContext(ctx, `SELECT kind FROM economy_category WHERE id=$1`, *id).Scan(&k)
	if errors.Is(err, sql.ErrNoRows) {
		return badRequest("CATEGORY_NOT_FOUND", "categoryId", "category does not exist")
	}
	if err != nil {
		return err
	}
	if k != kind {
		return badRequest("CATEGORY_KIND_MISMATCH", "categoryId", "category kind must be "+kind)
	}
	return nil
}

func boolOr(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

func saveFaq(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		CategoryID string `json:"categoryId"`
		Question   i18n   `json:"question"`
		Answer     i18n   `json:"answer"`
		IsActive   *bool  `json:"isActive"`
		Sort       int    `json:"sort"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	cat, err := cleanUUID("categoryId", in.CategoryID, false)
	if err != nil {
		return err
	}
	if err := checkCategory(ctx, tx, cat, "faq"); err != nil {
		return err
	}
	question, err := cleanI18n("question", in.Question, true, 500)
	if err != nil {
		return err
	}
	answer, err := cleanI18n("answer", in.Answer, true, 10000)
	if err != nil {
		return err
	}
	return execSave(ctx, tx, "publication_faq", id, create, []string{"category_id", "question", "answer", "is_active", "sort"},
		[]any{cat, i18nJSON(question), i18nJSON(answer), boolOr(in.IsActive, true), in.Sort})
}

func saveNews(s *service, ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Title        i18n            `json:"title"`
		Summary      i18n            `json:"summary"`
		Body         i18n            `json:"body"`
		CoverURL     string          `json:"coverUrl"`
		CategoryID   string          `json:"categoryId"`
		IsActive     *bool           `json:"isActive"`
		IsPinned     bool            `json:"isPinned"`
		PublishedAt  string          `json:"publishedAt"`
		AllowComment *bool           `json:"allowComment"`
		FreeComment  bool            `json:"freeComment"`
		Actions      []action        `json:"actions"`
		Metadata     json.RawMessage `json:"metadata"`
	}
	if err := decodeInto(raw, &in); err != nil {
		return err
	}
	title, err := cleanI18n("title", in.Title, true, 200)
	if err != nil {
		return err
	}
	summary, err := cleanI18n("summary", in.Summary, false, 1000)
	if err != nil {
		return err
	}
	body, err := cleanI18n("body", in.Body, true, 50000)
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
	published, err := cleanTime("publishedAt", in.PublishedAt)
	if err != nil {
		return err
	}
	if published == nil {
		now := time.Now().UTC()
		if !create {
			var old time.Time
			if err := tx.QueryRowContext(ctx, `SELECT published_at FROM publication_news WHERE id=$1`, id).Scan(&old); err == nil {
				now = old
			}
		}
		published = &now
	}
	if len(in.Actions) > 5 {
		return badRequest("TOO_MANY_ACTIONS", "actions", "at most 5 actions")
	}
	actions := make([]action, 0, len(in.Actions))
	for _, a := range in.Actions {
		clean, err := cleanAction("actions", a, true)
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
	return execSave(ctx, tx, "publication_news", id, create,
		[]string{"title", "summary", "body", "cover_url", "category_id", "is_active", "is_pinned", "published_at", "allow_comment", "free_comment", "actions", "metadata"},
		[]any{i18nJSON(title), i18nJSON(summary), i18nJSON(body), cover, cat, boolOr(in.IsActive, true), in.IsPinned, *published, boolOr(in.AllowComment, true), in.FreeComment, string(actionsJSON), metadata})
}

func cleanMessage(raw string) (string, error) {
	msg := strings.TrimSpace(raw)
	if msg == "" {
		return "", badRequest("COMMENT_MESSAGE_REQUIRED", "message", "message is required")
	}
	if !utf8.ValidString(msg) || utf8.RuneCountInString(msg) > 1000 {
		return "", badRequest("COMMENT_TOO_LONG", "message", "message is too long (max 1000)")
	}
	return msg, nil
}
