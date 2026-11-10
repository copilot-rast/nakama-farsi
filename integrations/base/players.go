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
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/restrict"
)

const (
	systemUserID     = "00000000-0000-0000-0000-000000000000"
	maxBulkPlayers   = 1000
	maxRestrictDays  = 3650
	kickWindow       = 2 * time.Minute
	expireLease      = "base-restriction-expire"
	restrictionLimit = 50
)

const playerTagsJSON = `COALESCE((SELECT jsonb_agg(jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'isPublic',t.is_public) ORDER BY t.sort, t.key) FROM base_user_tag ut JOIN base_tag t ON t.id=ut.tag_id WHERE ut.user_id=u.id),'[]'::jsonb)`

const activeRestrictionJSON = `(SELECT jsonb_build_object('id',r.id,'endsAt',r.ends_at,'reason',r.reason) FROM base_restriction r WHERE r.user_id=u.id AND r.kind=%s AND r.status='active' AND r.ends_at > now() ORDER BY r.ends_at DESC LIMIT 1)`

var playerJSON = `jsonb_build_object('id',u.id,'username',u.username,'displayName',u.display_name,'avatarUrl',u.avatar_url,'createTime',u.create_time,'banned',u.disable_time > '1970-01-01 00:00:00+00','tags',` + playerTagsJSON +
	`,'suspension',` + strings.Replace(activeRestrictionJSON, "%s", "'suspend'", 1) + `,'mute',` + strings.Replace(activeRestrictionJSON, "%s", "'mute'", 1) + `)`

const restrictionJSON = `jsonb_build_object('id',r.id,'userId',r.user_id,'username',COALESCE(u.username,''),'displayName',COALESCE(u.display_name,''),'kind',r.kind,
	'status',CASE WHEN r.status='active' AND r.ends_at <= now() THEN 'expired' ELSE r.status END,'startsAt',r.starts_at,'endsAt',r.ends_at,'reason',r.reason,'note',r.note,
	'reportId',r.report_id,'createdBy',r.created_by,'createdAt',r.created_at,'liftedBy',r.lifted_by,'liftedAt',r.lifted_at,'liftNote',r.lift_note,'endedAt',r.ended_at)`

func (s *service) playerList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	where.add("u.id <> ?::uuid", systemUserID)
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		cond := "(u.username ILIKE " + n + " OR u.display_name ILIKE " + n
		if _, err := uuid.FromString(p.Q); err == nil {
			where.args = append(where.args, p.Q)
			cond += " OR u.id = $" + itoa(len(where.args)) + "::uuid"
		}
		where.parts = append(where.parts, cond+")")
	}
	q := r.URL.Query()
	if v := strings.TrimSpace(q.Get("tagId")); v != "" {
		if _, err := uuid.FromString(v); err != nil {
			s.writeError(w, badRequest("INVALID_FILTER", "tagId", "tag id is not valid"))
			return
		}
		where.add("EXISTS (SELECT 1 FROM base_user_tag ut WHERE ut.user_id=u.id AND ut.tag_id=?::uuid)", v)
	}
	switch q.Get("status") {
	case "suspended":
		where.parts = append(where.parts, "EXISTS (SELECT 1 FROM base_restriction r WHERE r.user_id=u.id AND r.kind='suspend' AND r.status='active' AND r.ends_at > now())")
	case "muted":
		where.parts = append(where.parts, "EXISTS (SELECT 1 FROM base_restriction r WHERE r.user_id=u.id AND r.kind='mute' AND r.status='active' AND r.ends_at > now())")
	case "banned":
		where.parts = append(where.parts, "u.disable_time > '1970-01-01 00:00:00+00'")
	case "tagged":
		where.parts = append(where.parts, "EXISTS (SELECT 1 FROM base_user_tag ut WHERE ut.user_id=u.id)")
	case "":
	default:
		s.writeError(w, badRequest("INVALID_FILTER", "status", "status must be suspended, muted, banned or tagged"))
		return
	}
	out, err := s.listRaw(r, ` FROM users u`, playerJSON, "u.create_time DESC, u.id", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) userExists(ctx context.Context, q queryer, id string) error {
	if _, err := uuid.FromString(id); err != nil || id == systemUserID {
		return notFound("USER_NOT_FOUND")
	}
	var found bool
	if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id=$1)`, id).Scan(&found); err != nil {
		return err
	}
	if !found {
		return notFound("USER_NOT_FOUND")
	}
	return nil
}

func (s *service) playerView(ctx context.Context, q queryer, id string) (map[string]any, error) {
	if err := s.userExists(ctx, q, id); err != nil {
		return nil, err
	}
	var raw string
	if err := q.QueryRowContext(ctx, `SELECT `+playerJSON+`::text FROM users u WHERE u.id=$1`, id).Scan(&raw); err != nil {
		return nil, err
	}
	out := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *service) playerGet(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	out, err := s.playerView(r.Context(), s.db, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+restrictionJSON+`::text FROM base_restriction r LEFT JOIN users u ON u.id=r.user_id WHERE r.user_id=$1 ORDER BY r.created_at DESC LIMIT $2`, id, restrictionLimit)
	if err != nil {
		s.writeError(w, err)
		return
	}
	history, err := rawRows(rows)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var against, by, open int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FILTER (WHERE sender_id=$1), count(*) FILTER (WHERE reporter_id=$1), count(*) FILTER (WHERE sender_id=$1 AND status='open') FROM base_chat_report WHERE sender_id=$1 OR reporter_id=$1`, id).Scan(&against, &by, &open); err != nil {
		s.writeError(w, err)
		return
	}
	out["history"] = history
	out["reports"] = map[string]int64{"against": against, "by": by, "open": open}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) checkPlayerTags(ctx context.Context, tx *sql.Tx, ids []string) ([]string, error) {
	out := []string{}
	seen := map[string]bool{}
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if _, err := uuid.FromString(id); err != nil {
			return nil, badRequest("TAG_NOT_FOUND", "tagIds", "tag does not exist")
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		var kt string
		err := tx.QueryRowContext(ctx, `SELECT key_type FROM base_tag WHERE id=$1`, id).Scan(&kt)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, badRequest("TAG_NOT_FOUND", "tagIds", "tag does not exist")
		}
		if err != nil {
			return nil, err
		}
		if kt != "user" {
			return nil, badRequest("TAG_KIND_MISMATCH", "tagIds", "only player tags can be given to players")
		}
		out = append(out, id)
	}
	if len(out) > 50 {
		return nil, badRequest("TOO_MANY_TAGS", "tagIds", "at most 50 tags")
	}
	return out, nil
}

func (s *service) playerTagsPut(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	var in struct {
		TagIDs []string `json:"tagIds"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	var out map[string]any
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		if err := s.userExists(r.Context(), tx, id); err != nil {
			return err
		}
		tags, err := s.checkPlayerTags(r.Context(), tx, in.TagIDs)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(r.Context(), `DELETE FROM base_user_tag WHERE user_id=$1 AND NOT (tag_id::text = ANY($2::text[]))`, id, tags); err != nil {
			return err
		}
		if _, err := tx.ExecContext(r.Context(), `INSERT INTO base_user_tag (user_id, tag_id, created_by) SELECT $1, t::uuid, $3 FROM unnest($2::text[]) t ON CONFLICT DO NOTHING`, id, tags, consoleUser(r)); err != nil {
			return err
		}
		out, err = s.playerView(r.Context(), tx, id)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) playerTagsBulk(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserIDs []string `json:"userIds"`
		TagIDs  []string `json:"tagIds"`
		Op      string   `json:"op"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if in.Op != "add" && in.Op != "remove" {
		s.writeError(w, badRequest("INVALID_OP", "op", "op must be add or remove"))
		return
	}
	users := []string{}
	seen := map[string]bool{}
	for _, raw := range in.UserIDs {
		id := strings.TrimSpace(raw)
		if _, err := uuid.FromString(id); err != nil || id == systemUserID {
			s.writeError(w, badRequest("USERS_NOT_FOUND", "userIds", "a player id is not valid"))
			return
		}
		if !seen[id] {
			seen[id] = true
			users = append(users, id)
		}
	}
	if len(users) == 0 {
		s.writeError(w, badRequest("USERS_REQUIRED", "userIds", "pick at least one player"))
		return
	}
	if len(users) > maxBulkPlayers {
		s.writeError(w, badRequest("TOO_MANY_USERS", "userIds", "at most 1000 players at once"))
		return
	}
	var changed int64
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		tags, err := s.checkPlayerTags(r.Context(), tx, in.TagIDs)
		if err != nil {
			return err
		}
		if len(tags) == 0 {
			return badRequest("TAGS_REQUIRED", "tagIds", "pick at least one tag")
		}
		var found int
		if err := tx.QueryRowContext(r.Context(), `SELECT count(*) FROM users WHERE id::text = ANY($1::text[])`, users).Scan(&found); err != nil {
			return err
		}
		if found != len(users) {
			return badRequest("USERS_NOT_FOUND", "userIds", strconv.Itoa(len(users)-found)+" of the players do not exist")
		}
		var res sql.Result
		if in.Op == "add" {
			res, err = tx.ExecContext(r.Context(), `INSERT INTO base_user_tag (user_id, tag_id, created_by) SELECT u::uuid, t::uuid, $3 FROM unnest($1::text[]) u CROSS JOIN unnest($2::text[]) t ON CONFLICT DO NOTHING`, users, tags, consoleUser(r))
		} else {
			res, err = tx.ExecContext(r.Context(), `DELETE FROM base_user_tag WHERE user_id::text = ANY($1::text[]) AND tag_id::text = ANY($2::text[])`, users, tags)
		}
		if err != nil {
			return err
		}
		changed, err = res.RowsAffected()
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"changed": changed, "players": len(users)})
}

type restrictionInput struct {
	Kind    string         `json:"kind"`
	EndsAt  string         `json:"endsAt"`
	Minutes int            `json:"minutes"`
	Reason  localized.Text `json:"reason"`
	Note    string         `json:"note"`
}

type restrictionPlan struct {
	userID   string
	kind     string
	endsAt   time.Time
	reason   localized.Text
	note     string
	by       string
	reportID string
}

func (s *service) planRestriction(ctx context.Context, q localized.Queryer, userID string, in restrictionInput, by string) (restrictionPlan, error) {
	plan := restrictionPlan{userID: userID, kind: strings.TrimSpace(in.Kind), by: by}
	switch plan.kind {
	case restrict.KindSuspend, restrict.KindMute, restrict.KindWarn:
	default:
		return plan, badRequest("INVALID_RESTRICTION_KIND", "kind", "kind must be suspend, mute or warn")
	}
	now := s.now()
	if plan.kind != restrict.KindWarn {
		switch {
		case strings.TrimSpace(in.EndsAt) != "":
			t, err := time.Parse(time.RFC3339, strings.TrimSpace(in.EndsAt))
			if err != nil {
				return plan, badRequest("INVALID_END", "endsAt", "use an ISO date and time")
			}
			plan.endsAt = t
		case in.Minutes > 0:
			plan.endsAt = now.Add(time.Duration(in.Minutes) * time.Minute)
		default:
			return plan, badRequest("END_REQUIRED", "endsAt", "pick when it ends")
		}
		if !plan.endsAt.After(now.Add(30 * time.Second)) {
			return plan, badRequest("END_IN_PAST", "endsAt", "the end must be in the future")
		}
		if plan.endsAt.After(now.AddDate(0, 0, maxRestrictDays)) {
			return plan, badRequest("END_TOO_FAR", "endsAt", "at most ten years; use a ban for longer")
		}
	} else {
		plan.endsAt = now
	}
	langs, err := s.languages(ctx, q)
	if err != nil {
		return plan, err
	}
	plan.reason, err = cleanText(langs, "reason", in.Reason, true, 500)
	if err != nil {
		return plan, err
	}
	plan.note = strings.TrimSpace(in.Note)
	if len([]rune(plan.note)) > 1000 {
		return plan, badRequest("TEXT_TOO_LONG", "note", "at most 1000 letters")
	}
	return plan, nil
}

func (s *service) applyRestriction(ctx context.Context, tx *sql.Tx, plan restrictionPlan) (string, error) {
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('base-restrict:' || $1))`, plan.userID); err != nil {
		return "", err
	}
	if err := s.userExists(ctx, tx, plan.userID); err != nil {
		return "", err
	}
	id := uuid.Must(uuid.NewV4()).String()
	var report any
	if plan.reportID != "" {
		report = plan.reportID
	}
	if plan.kind == restrict.KindWarn {
		_, err := tx.ExecContext(ctx, `INSERT INTO base_restriction (id, user_id, kind, status, starts_at, ends_at, reason, note, report_id, created_by, ended_at)
			VALUES ($1, $2, 'warn', 'expired', now(), now(), $3, $4, $5, $6, now())`, id, plan.userID, plan.reason.JSON(), plan.note, report, plan.by)
		return id, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE base_restriction SET status = CASE WHEN ends_at <= now() THEN 'expired' ELSE 'lifted' END,
		lifted_by = CASE WHEN ends_at <= now() THEN lifted_by ELSE $3 END, lifted_at = CASE WHEN ends_at <= now() THEN lifted_at ELSE now() END,
		lift_note = CASE WHEN ends_at <= now() THEN lift_note ELSE 'replaced' END, ended_at = LEAST(ends_at, now())
		WHERE user_id=$1 AND kind=$2 AND status='active'`, plan.userID, plan.kind, plan.by); err != nil {
		return "", err
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO base_restriction (id, user_id, kind, status, starts_at, ends_at, reason, note, report_id, created_by)
		VALUES ($1, $2, $3, 'active', now(), $4, $5, $6, $7, $8)`, id, plan.userID, plan.kind, plan.endsAt.UTC(), plan.reason.JSON(), plan.note, report, plan.by)
	return id, err
}

func (s *service) restrictionView(ctx context.Context, q queryer, id string) (json.RawMessage, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("RESTRICTION_NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+restrictionJSON+`::text FROM base_restriction r LEFT JOIN users u ON u.id=r.user_id WHERE r.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("RESTRICTION_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func (s *service) announceRestriction(ctx context.Context, userID, id string) {
	var kind, reason string
	var ends time.Time
	if err := s.db.QueryRowContext(ctx, `SELECT kind, ends_at, reason::text FROM base_restriction WHERE id=$1`, id).Scan(&kind, &ends, &reason); err != nil {
		s.logger.Warn("Base restriction notice skipped: %v", err)
		return
	}
	r := &restrict.Restriction{ID: id, UserID: userID, Kind: kind, EndsAt: ends, Reason: localized.Parse(reason)}
	notice := r.Notice(ctx, s.db, "")
	code, persistent := restrict.NotifyMuted, false
	switch kind {
	case restrict.KindSuspend:
		code = restrict.NotifySuspended
	case restrict.KindWarn:
		code, persistent = restrict.NotifyWarned, true
		notice.Code = "WARNED"
		notice.Until = ""
	}
	content := notice.Map()
	content["restrictionId"] = id
	if err := s.nk.NotificationSend(ctx, userID, notice.Reason, content, code, "", persistent); err != nil {
		s.logger.Warn("Base restriction notice failed: %v", err)
	}
	if kind == restrict.KindSuspend {
		if err := s.nk.SessionLogout(userID, "", ""); err != nil {
			s.logger.Warn("Base suspension logout failed for %s: %v", userID, err)
		}
		go func() {
			time.Sleep(500 * time.Millisecond)
			s.disconnectPlayer(context.Background(), userID)
		}()
	}
}

func (s *service) kickPlayer(ctx context.Context, userID string) {
	if err := s.nk.SessionLogout(userID, "", ""); err != nil {
		s.logger.Warn("Base suspension logout failed for %s: %v", userID, err)
	}
	s.disconnectPlayer(ctx, userID)
}

func (s *service) disconnectPlayer(ctx context.Context, userID string) {
	presences, err := s.nk.StreamUserList(0, userID, "", "", true, true)
	if err != nil {
		s.logger.Warn("Base suspension presence list failed for %s: %v", userID, err)
		return
	}
	for _, p := range presences {
		if err := s.nk.SessionDisconnect(ctx, p.GetSessionId()); err != nil {
			s.logger.Warn("Base suspension disconnect failed for %s: %v", userID, err)
		}
	}
}

func (s *service) restrictionCreate(w http.ResponseWriter, r *http.Request) {
	userID := pathVar(r, "id")
	var in restrictionInput
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	plan, err := s.planRestriction(r.Context(), s.db, userID, in, consoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	var out json.RawMessage
	var id string
	err = s.inTx(r.Context(), func(tx *sql.Tx) error {
		var err error
		if id, err = s.applyRestriction(r.Context(), tx, plan); err != nil {
			return err
		}
		out, err = s.restrictionView(r.Context(), tx, id)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.announceRestriction(r.Context(), userID, id)
	writeRaw(w, http.StatusCreated, out)
}

func (s *service) restrictionList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	q := r.URL.Query()
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(u.username ILIKE "+n+" OR u.display_name ILIKE "+n+" OR r.note ILIKE "+n+" OR "+translatedSearch("r.reason", n)+")")
	}
	if v := q.Get("kind"); v != "" {
		if v != restrict.KindSuspend && v != restrict.KindMute && v != restrict.KindWarn {
			s.writeError(w, badRequest("INVALID_RESTRICTION_KIND", "kind", "kind must be suspend, mute or warn"))
			return
		}
		where.add("r.kind = ?", v)
	}
	switch q.Get("status") {
	case "active":
		where.parts = append(where.parts, "r.status='active' AND r.ends_at > now()")
	case "expired":
		where.parts = append(where.parts, "(r.status='expired' OR (r.status='active' AND r.ends_at <= now()))")
	case "lifted":
		where.parts = append(where.parts, "r.status='lifted'")
	case "":
	default:
		s.writeError(w, badRequest("INVALID_FILTER", "status", "status must be active, expired or lifted"))
		return
	}
	if v := strings.TrimSpace(q.Get("userId")); v != "" {
		if _, err := uuid.FromString(v); err != nil {
			s.writeError(w, badRequest("USER_NOT_FOUND", "userId", "player id is not valid"))
			return
		}
		where.add("r.user_id = ?::uuid", v)
	}
	out, err := s.listRaw(r, ` FROM base_restriction r LEFT JOIN users u ON u.id=r.user_id`, restrictionJSON, "r.created_at DESC, r.id", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var active int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM base_restriction WHERE status='active' AND ends_at > now() AND kind <> 'warn'`).Scan(&active); err != nil {
		s.writeError(w, err)
		return
	}
	out["activeTotal"] = active
	writeJSON(w, http.StatusOK, out)
}

func (s *service) restrictionGet(w http.ResponseWriter, r *http.Request) {
	out, err := s.restrictionView(r.Context(), s.db, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, out)
}

func (s *service) restrictionLift(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	var in struct {
		Note string `json:"note"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	note := strings.TrimSpace(in.Note)
	if len([]rune(note)) > 1000 {
		s.writeError(w, badRequest("TEXT_TOO_LONG", "note", "at most 1000 letters"))
		return
	}
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("RESTRICTION_NOT_FOUND"))
		return
	}
	var userID, kind string
	var out json.RawMessage
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		err := tx.QueryRowContext(r.Context(), `UPDATE base_restriction SET status='lifted', lifted_by=$2, lifted_at=now(), lift_note=$3, ended_at=now()
			WHERE id=$1 AND status='active' AND ends_at > now() AND kind <> 'warn' RETURNING user_id::text, kind`, id, consoleUser(r), note).Scan(&userID, &kind)
		if errors.Is(err, sql.ErrNoRows) {
			if _, err := s.restrictionView(r.Context(), tx, id); err != nil {
				return err
			}
			return conflict("NOT_ACTIVE", "this restriction has already ended")
		}
		if err != nil {
			return err
		}
		out, err = s.restrictionView(r.Context(), tx, id)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	if err := s.nk.NotificationSend(r.Context(), userID, kind, map[string]any{"code": "LIFTED", "kind": kind, "restrictionId": id}, restrict.NotifyLifted, "", false); err != nil {
		s.logger.Warn("Base lift notice failed: %v", err)
	}
	writeRaw(w, http.StatusOK, out)
}

func (s *service) restrictionLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	tick := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		tick++
		s.kickFresh(ctx)
		if tick%6 != 0 {
			continue
		}
		ok, err := clusterkit.AcquireLease(ctx, s.db, expireLease, holder, 25*time.Second)
		if err != nil || !ok {
			continue
		}
		if _, err := s.expireRestrictions(ctx); err != nil && ctx.Err() == nil {
			s.logger.Error("Base restriction expiry failed: %v", err)
		}
	}
}

func (s *service) expireRestrictions(ctx context.Context) (int64, error) {
	res, err := s.db.ExecContext(ctx, `UPDATE base_restriction SET status='expired', ended_at=ends_at WHERE status='active' AND ends_at <= now()`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *service) kickFresh(ctx context.Context) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT user_id::text FROM base_restriction WHERE kind='suspend' AND status='active' AND ends_at > now() AND created_at > now() - $1::INT8 * INTERVAL '1 second'`, int64(kickWindow.Seconds()))
	if err != nil {
		if ctx.Err() == nil {
			s.logger.Warn("Base suspension scan failed: %v", err)
		}
		return
	}
	users := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			users = append(users, id)
		}
	}
	rows.Close()
	for _, id := range users {
		s.kickPlayer(ctx, id)
	}
}
