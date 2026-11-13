package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
)

func (s *service) providerPushTest(w http.ResponseWriter, r *http.Request, p provider, to, title, message string) {
	who := strings.TrimSpace(to)
	if who == "" {
		s.writeError(w, badRequest("USER_NOT_FOUND", "to", "give a player's username or ID"))
		return
	}
	var uid string
	err := s.db.QueryRowContext(r.Context(), `SELECT id::text FROM users WHERE (id::text = $1 OR username = $1) AND id <> '`+systemUserID+`'::uuid LIMIT 1`, who).Scan(&uid)
	if errors.Is(err, sql.ErrNoRows) {
		s.writeError(w, badRequest("USER_NOT_FOUND", "to", "no player with this username or ID"))
		return
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	if !providerConfigured(p) {
		s.writeError(w, badRequest("PROVIDER_NOT_CONFIGURED", "config", "fill the required fields first"))
		return
	}
	devices, err := s.userDevices(r.Context(), []string{uid})
	if err != nil {
		s.writeError(w, err)
		return
	}
	devs := devices[uid]
	if len(devs) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"ok": false, "sandbox": p.Sandbox, "error": "this player has no registered device"})
		return
	}
	if strings.TrimSpace(title) == "" {
		title = "Test"
	}
	sent, dead, failed := 0, 0, []string{}
	for _, d := range devs {
		res, err := s.sendPush(r.Context(), p, pushMessage{Token: d.Token, Platform: d.Platform, Title: title, Body: message, Data: map[string]string{"test": "1"}})
		switch {
		case err == nil:
			sent++
		case res.Dead:
			dead++
			_, _ = s.db.ExecContext(r.Context(), `UPDATE publication_push_device SET is_active=false, disabled_at=now(), last_error=$2, updated_at=now() WHERE id=$1`, d.ID, truncate(err.Error(), 300))
		default:
			failed = append(failed, err.Error())
		}
	}
	reply := map[string]any{"ok": sent > 0, "sandbox": p.Sandbox || p.Kind == "log", "devices": len(devs), "sent": sent, "dead": dead}
	if p.Sandbox || p.Kind == "log" {
		reply["reply"] = "sandbox: nothing was sent"
	}
	if len(failed) > 0 {
		reply["error"] = failed[0]
	}
	writeJSON(w, http.StatusOK, reply)
}

func (s *service) consolePushDevices(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	q := r.URL.Query()
	where := &sqlWhere{}
	if v := strings.TrimSpace(q.Get("userId")); v != "" {
		id, err := cleanUUID("userId", v, true)
		if err != nil {
			s.writeError(w, err)
			return
		}
		where.add("d.user_id = ?::uuid", *id)
	}
	if v := q.Get("platform"); v != "" {
		where.add("d.platform = ?", v)
	}
	if v := q.Get("active"); v == "true" || v == "false" {
		where.add("d.is_active = ?::boolean", v)
	}
	var total int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM publication_push_device d`+where.sql(), where.args...).Scan(&total); err != nil {
		s.writeError(w, err)
		return
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(r.Context(), `SELECT jsonb_build_object('id',d.id,'userId',d.user_id,'username',u.username,'platform',d.platform,'appVersion',d.app_version,
		'token',left(d.token,6)||'…'||right(d.token,4),'isActive',d.is_active,'failCount',d.fail_count,'lastError',d.last_error,'disabledAt',d.disabled_at,'lastSeenAt',d.last_seen_at,'createdAt',d.created_at)::text
		FROM publication_push_device d LEFT JOIN users u ON u.id=d.user_id`+where.sql()+` ORDER BY d.last_seen_at DESC, d.id LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2),
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
