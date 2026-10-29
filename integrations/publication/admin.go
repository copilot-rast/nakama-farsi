package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
)

const consolePrefix = "/v2/console/publication"

func (s *service) registerConsoleHTTP(initializer runtime.Initializer) error {
	type route struct {
		path, method string
		handler      http.HandlerFunc
	}
	routes := []route{
		{consolePrefix + "/summary", http.MethodGet, s.consoleSummary},
		{consolePrefix + "/comments", http.MethodGet, s.consoleCommentsList},
		{consolePrefix + "/comments/{id}", http.MethodPut, s.consoleCommentUpdate},
		{consolePrefix + "/comments/{id}", http.MethodDelete, s.consoleCommentDelete},
		{consolePrefix + "/notifications", http.MethodGet, s.consoleNotificationsList},
		{consolePrefix + "/notifications", http.MethodPost, s.consoleNotificationCreate},
		{consolePrefix + "/notifications/preview-target", http.MethodGet, s.consolePreviewTarget},
		{consolePrefix + "/notifications/{id}", http.MethodGet, s.consoleNotificationGet},
		{consolePrefix + "/providers", http.MethodGet, s.consoleProvidersList},
		{consolePrefix + "/providers", http.MethodPost, s.consoleProviderCreate},
		{consolePrefix + "/providers/{id}", http.MethodPut, s.consoleProviderUpdate},
		{consolePrefix + "/providers/{id}", http.MethodDelete, s.consoleProviderDelete},
		{consolePrefix + "/providers/{id}/test", http.MethodPost, s.consoleProviderTest},
		{consolePrefix + "/users/search", http.MethodGet, s.consoleUsersSearch},
	}
	for _, res := range resources() {
		base := consolePrefix + "/" + res.name
		routes = append(routes,
			route{base, http.MethodGet, s.resourceList(res)},
			route{base, http.MethodPost, s.resourceWrite(res, true)},
			route{base + "/{id}", http.MethodGet, s.resourceGet(res)},
			route{base + "/{id}", http.MethodPut, s.resourceWrite(res, false)},
			route{base + "/{id}", http.MethodDelete, s.resourceDelete(res)},
		)
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

const commentAdminJSON = `jsonb_build_object('id',m.id,'newsId',m.news_id,'newsTitle',n.title,'userId',m.user_id,'username',u.username,'displayName',u.display_name,
	'parentId',m.parent_id,'message',m.message,'isAccepted',m.is_accepted,'replyCount',(SELECT count(*) FROM publication_news_comment rc WHERE rc.parent_id=m.id),
	'createdAt',m.created_at,'updatedAt',m.updated_at)`

const commentAdminFrom = ` FROM publication_news_comment m JOIN publication_news n ON n.id=m.news_id LEFT JOIN users u ON u.id=m.user_id`

func (s *service) consoleCommentsList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	q := r.URL.Query()
	where := &sqlWhere{}
	if v := q.Get("newsId"); v != "" {
		if _, err := uuid.FromString(v); err != nil {
			s.writeError(w, badRequest("INVALID_FILTER", "newsId", "must be a UUID"))
			return
		}
		where.add("m.news_id = ?::uuid", v)
	}
	switch q.Get("status") {
	case "":
	case "pending":
		where.add("NOT m.is_accepted")
	case "accepted":
		where.add("m.is_accepted")
	default:
		s.writeError(w, badRequest("INVALID_FILTER", "status", "must be pending or accepted"))
		return
	}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(m.message ILIKE "+n+" OR COALESCE(u.username,'') ILIKE "+n+" OR COALESCE(n.title->>'fa','') ILIKE "+n+")")
	}
	var total int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*)`+commentAdminFrom+where.sql(), where.args...).Scan(&total); err != nil {
		s.writeError(w, err)
		return
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+commentAdminJSON+`::text`+commentAdminFrom+where.sql()+` ORDER BY m.created_at DESC, m.id DESC LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2),
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

func (s *service) consoleCommentUpdate(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("COMMENT_NOT_FOUND"))
		return
	}
	var in struct {
		IsAccepted *bool `json:"isAccepted"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if in.IsAccepted == nil {
		s.writeError(w, badRequest("REQUIRED", "isAccepted", "isAccepted is required"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `UPDATE publication_news_comment SET is_accepted=$2, updated_at=now() WHERE id=$1`, id, *in.IsAccepted)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.writeError(w, notFound("COMMENT_NOT_FOUND"))
		return
	}
	var raw string
	err = s.db.QueryRowContext(r.Context(), `SELECT `+commentAdminJSON+`::text`+commentAdminFrom+` WHERE m.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		s.writeError(w, notFound("COMMENT_NOT_FOUND"))
		return
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, []byte(raw))
}

func (s *service) consoleCommentDelete(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("COMMENT_NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM publication_news_comment WHERE id=$1`, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.writeError(w, notFound("COMMENT_NOT_FOUND"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) consoleUsersSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	if len([]rune(q)) > 100 {
		s.writeError(w, badRequest("QUERY_TOO_LONG", "q", "search text is too long"))
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT jsonb_build_object('id',id,'username',username,'displayName',display_name)::text FROM users
		WHERE id <> '`+systemUserID+`'::uuid AND (username ILIKE $1 OR COALESCE(display_name,'') ILIKE $1 OR id::text = $2)
		ORDER BY (id::text = $2) DESC, (username ILIKE $3) DESC, username LIMIT 20`, likePattern(q), strings.ToLower(q), strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(q)+"%")
	if err != nil {
		s.writeError(w, err)
		return
	}
	items, err := rawRows(rows)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *service) consoleSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	counts := map[string]int64{}
	for name, query := range map[string]string{
		"faqs":            `SELECT count(*) FROM publication_faq`,
		"activeFaqs":      `SELECT count(*) FROM publication_faq WHERE is_active`,
		"news":            `SELECT count(*) FROM publication_news`,
		"publishedNews":   `SELECT count(*) FROM publication_news WHERE is_active AND published_at <= now()`,
		"scheduledNews":   `SELECT count(*) FROM publication_news WHERE is_active AND published_at > now()`,
		"comments":        `SELECT count(*) FROM publication_news_comment`,
		"pendingComments": `SELECT count(*) FROM publication_news_comment WHERE NOT is_accepted`,
		"notifications":   `SELECT count(*) FROM publication_notification`,
		"notifications7d": `SELECT count(*) FROM publication_notification WHERE created_at >= now() - interval '7 days'`,
		"providers":       `SELECT count(*) FROM publication_provider`,
		"activeProviders": `SELECT count(*) FROM publication_provider WHERE is_active`,
	} {
		var n int64
		if err := s.db.QueryRowContext(ctx, query).Scan(&n); err != nil {
			s.writeError(w, err)
			return
		}
		counts[name] = n
	}
	byStatus := map[string]int64{"queued": 0, "sending": 0, "done": 0, "failed": 0}
	byChannel := map[string]int64{"in_app": 0, "sms": 0, "email": 0}
	rows, err := s.db.QueryContext(ctx, `SELECT status, channel, count(*) FROM publication_notification GROUP BY status, channel`)
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var st, ch string
		var n int64
		if err := rows.Scan(&st, &ch, &n); err != nil {
			s.writeError(w, err)
			return
		}
		byStatus[st] += n
		byChannel[ch] += n
	}
	var providers []json.RawMessage
	prow, err := s.db.QueryContext(ctx, `SELECT jsonb_build_object('id',id,'channel',channel,'kind',kind,'isActive',is_active,'isDefault',is_default,'sandbox',sandbox)::text FROM publication_provider ORDER BY channel, kind`)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if providers, err = rawRows(prow); err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"counts": counts, "notificationsByStatus": byStatus, "notificationsByChannel": byChannel, "providers": providers})
}
