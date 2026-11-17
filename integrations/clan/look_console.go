package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *service) lookRoutes() []consoleRoute {
	p := consolePrefix
	return []consoleRoute{
		{p + "/art", http.MethodGet, s.artList},
		{p + "/art", http.MethodPost, s.artSave},
		{p + "/art/{id}", http.MethodPut, s.artSave},
		{p + "/art/{id}", http.MethodDelete, s.artDelete},
		{p + "/uploads", http.MethodGet, s.uploadList},
		{p + "/uploads/{id}/file", http.MethodGet, s.uploadFile},
		{p + "/uploads/{id}/approve", http.MethodPost, s.uploadApprove},
		{p + "/uploads/{id}/reject", http.MethodPost, s.uploadReject},
		{p + "/list/{id}/reset-look", http.MethodPost, s.resetLook},
	}
}

func (s *service) artList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := s.arts(ctx, false)
	if err != nil {
		s.fail(w, err)
		return
	}
	used := map[string]int{}
	rows, err := s.db.QueryContext(ctx, `SELECT 'emblem:'||emblem_key, count(*) FROM clan WHERE emblem_key<>'' GROUP BY 1 UNION ALL SELECT 'cover:'||cover_key, count(*) FROM clan WHERE cover_key<>'' GROUP BY 1`)
	if err == nil {
		for rows.Next() {
			var k string
			var n int
			if rows.Scan(&k, &n) == nil {
				used[k] = n
			}
		}
		rows.Close()
	}
	items := make([]map[string]any, 0, len(list))
	for _, a := range list {
		raw, _ := json.Marshal(a)
		m := map[string]any{}
		_ = json.Unmarshal(raw, &m)
		m["clans"] = used[a.Kind+":"+a.Key]
		items = append(items, m)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *service) artSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in art
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	in.Key = strings.TrimSpace(in.Key)
	in.ImageURL = strings.TrimSpace(in.ImageURL)
	switch {
	case in.Kind != artEmblem && in.Kind != artCover:
		s.fail(w, bad("INVALID_POLICY", "kind"))
		return
	case !settings.KeyPattern.MatchString(in.Key):
		s.fail(w, bad("INVALID_KEY", "key"))
		return
	case in.ImageURL == "" || len(in.ImageURL) > 500 || !(strings.HasPrefix(in.ImageURL, "https://") || strings.HasPrefix(in.ImageURL, "http://") || strings.HasPrefix(in.ImageURL, "/")):
		s.fail(w, bad("CLAN_ART_IMAGE_REQUIRED", "imageUrl"))
		return
	case in.MinLevel < 1 || in.MinLevel > 10000:
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "minLevel"))
		return
	case in.Amount < 0 || in.Amount > 1e12 || (in.Amount > 0 && in.Currency == ""):
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "priceAmount"))
		return
	}
	if in.Currency != "" {
		var ok bool
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, in.Currency).Scan(&ok)
		if !ok {
			s.fail(w, bad("CURRENCY_NOT_FOUND", "priceCurrency"))
			return
		}
	}
	var err error
	if in.Name, err = s.cleanText(ctx, "name", in.Name, false, 80); err != nil {
		s.fail(w, err)
		return
	}
	name, _ := json.Marshal(in.Name)
	by := pluginkit.ConsoleUser(r)
	err = pluginkit.InTx(ctx, s.db, func(tx *sql.Tx) error {
		if in.Default {
			if _, err := tx.ExecContext(ctx, `UPDATE clan_art SET is_default=false WHERE kind=$1`, in.Kind); err != nil {
				return err
			}
		}
		var res sql.Result
		var err error
		if id == "" {
			res, err = tx.ExecContext(ctx, `INSERT INTO clan_art (id, kind, key, name, image_url, sort, is_active, min_level, price_currency, price_amount, is_default, updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
				uuid.Must(uuid.NewV4()).String(), in.Kind, in.Key, name, in.ImageURL, in.Sort, in.Active, in.MinLevel, in.Currency, in.Amount, in.Default, by)
		} else {
			res, err = tx.ExecContext(ctx, `UPDATE clan_art SET kind=$2, key=$3, name=$4, image_url=$5, sort=$6, is_active=$7, min_level=$8, price_currency=$9, price_amount=$10, is_default=$11, updated_by=$12, updated_at=now() WHERE id::text=$1`,
				id, in.Kind, in.Key, name, in.ImageURL, in.Sort, in.Active, in.MinLevel, in.Currency, in.Amount, in.Default, by)
		}
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return pluginkit.Missing("CLAN_ART_NOT_FOUND")
		}
		if id != "" {
			col := map[string]string{artEmblem: "emblem_key", artCover: "cover_key"}[in.Kind]
			if in.Kind == artCover {
				_, err = tx.ExecContext(ctx, `UPDATE clan SET cover_url=$2 WHERE cover_key=$1`, in.Key, in.ImageURL)
			} else {
				_, err = tx.ExecContext(ctx, `UPDATE groups g SET avatar_url=$2 FROM clan c WHERE c.group_id=g.id AND c.`+col+`=$1`, in.Key, in.ImageURL)
			}
		}
		return err
	})
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "CLAN_ART_EXISTS", Field: "key"})
			return
		}
		s.fail(w, err)
		return
	}
	s.artList(w, r)
}

func (s *service) artDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var kind, key string
	if err := s.db.QueryRowContext(ctx, `SELECT kind, key FROM clan_art WHERE id::text=$1`, id).Scan(&kind, &key); err != nil {
		s.fail(w, pluginkit.Missing("CLAN_ART_NOT_FOUND"))
		return
	}
	var n int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM clan WHERE (emblem_key=$1 AND $2='emblem') OR (cover_key=$1 AND $2='cover')`, key, kind).Scan(&n)
	if n > 0 {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "IN_USE", Meta: map[string]any{"clans": n}})
		return
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM clan_art WHERE id::text=$1`, id); err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) uploadList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, limit := pluginkit.Page(r, 25, 100)
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}
	var total int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM clan_upload WHERE status=$1 OR $1='all'`, status).Scan(&total)
	rows, err := s.db.QueryContext(ctx, `SELECT u.id::text, u.group_id::text, COALESCE(g.name,''), u.kind, u.file_name, u.content_type, u.size, u.status, COALESCE(u.uploaded_by::text,''), COALESCE(us.username,''),
		u.reviewed_by, u.reason, u.created_at, u.reviewed_at FROM clan_upload u LEFT JOIN groups g ON g.id=u.group_id LEFT JOIN users us ON us.id=u.uploaded_by
		WHERE u.status=$1 OR $1='all' ORDER BY u.created_at DESC LIMIT $2 OFFSET $3`, status, limit, (page-1)*limit)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, gid, gname, kind, file, ctype, st, uid, uname, rby, reason string
		var size int
		var at time.Time
		var rat *time.Time
		if err := rows.Scan(&id, &gid, &gname, &kind, &file, &ctype, &size, &st, &uid, &uname, &rby, &reason, &at, &rat); err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, map[string]any{"id": id, "clanId": gid, "clanName": gname, "kind": kind, "fileName": file, "contentType": ctype, "size": size, "status": st,
			"userId": uid, "username": uname, "reviewedBy": rby, "reason": reason, "createdAt": at.UTC().Format(time.RFC3339), "reviewedAt": iso(rat), "url": s.fileURL(file)})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) uploadRow(r *http.Request) (string, string, string, string, error) {
	var gid, kind, file, status string
	err := s.db.QueryRowContext(r.Context(), `SELECT group_id::text, kind, file_name, status FROM clan_upload WHERE id::text=$1`, pluginkit.PathVar(r, "id")).Scan(&gid, &kind, &file, &status)
	if err != nil {
		return "", "", "", "", pluginkit.Missing("CLAN_UPLOAD_NOT_FOUND")
	}
	return gid, kind, file, status, nil
}

func (s *service) uploadFile(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		http.NotFound(w, r)
		return
	}
	_, _, file, _, err := s.uploadRow(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	s.serveFile(w, r, file, "private, no-store")
}

func (s *service) uploadApprove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gid, kind, file, status, err := s.uploadRow(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	if status != "pending" {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "CLAN_UPLOAD_NOT_PENDING"})
		return
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE clan_upload SET status='approved', reviewed_by=$2, reviewed_at=now() WHERE id::text=$1`, pluginkit.PathVar(r, "id"), pluginkit.ConsoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	if err := s.useUpload(ctx, gid, kind, file, true); err != nil {
		s.fail(w, err)
		return
	}
	cfg := s.settings(ctx)
	s.notify(ctx, s.withRoles(ctx, gid, cfg.Roles.Edit), "clan_upload_reviewed", map[string]any{"clanId": gid, "kind": kind, "status": "approved", "url": s.fileURL(file)}, true)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"approved": true})
}

func (s *service) uploadReject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in struct {
		Reason string `json:"reason"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	gid, kind, file, status, err := s.uploadRow(r)
	if err != nil {
		s.fail(w, err)
		return
	}
	if status != "pending" {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "CLAN_UPLOAD_NOT_PENDING"})
		return
	}
	reason := strings.TrimSpace(in.Reason)
	if len(reason) > 300 {
		reason = reason[:300]
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE clan_upload SET status='rejected', reason=$3, reviewed_by=$2, reviewed_at=now() WHERE id::text=$1`, pluginkit.PathVar(r, "id"), pluginkit.ConsoleUser(r), reason); err != nil {
		s.fail(w, err)
		return
	}
	if s.store != nil {
		_ = s.store.Delete(ctx, file)
	}
	cfg := s.settings(ctx)
	s.notify(ctx, s.withRoles(ctx, gid, cfg.Roles.Edit), "clan_upload_reviewed", map[string]any{"clanId": gid, "kind": kind, "status": "rejected", "reason": reason}, true)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"rejected": true})
}

func (s *service) resetLook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	c, err := s.clanByID(ctx, s.db, pluginkit.PathVar(r, "id"))
	if err != nil {
		s.consoleFail(w, err)
		return
	}
	empty := ""
	p := &lookPlan{tags: c.Tags, country: c.Country, changed: true}
	eKey, cKey, eURL, cURL := empty, empty, empty, empty
	if d := s.defaultArt(ctx, artEmblem); d != nil {
		eKey, eURL = d.Key, d.ImageURL
	}
	if d := s.defaultArt(ctx, artCover); d != nil {
		cKey, cURL = d.Key, d.ImageURL
	}
	p.emblemKey, p.avatar, p.coverKey, p.cover = &eKey, &eURL, &cKey, &cURL
	if err := s.applyPlan(ctx, c, c.JoinType, p, editor{admin: true}); err != nil {
		s.fail(w, err)
		return
	}
	_, _ = s.db.ExecContext(ctx, `UPDATE clan_upload SET status='replaced', reviewed_at=now() WHERE group_id::text=$1 AND status='pending'`, c.ID)
	s.clanGet(w, r)
}
