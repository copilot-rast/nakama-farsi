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
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	playerReportNoteMax   = 1000
	playerReportLimit     = 10
	playerReportWindow    = time.Hour
	playerReportDupWindow = 24 * time.Hour
	playerReportNoteLimit = 50
)

var (
	playerReportStatuses = []string{"open", "reviewing", "resolved", "dismissed"}
	playerReportActions  = []string{"dismiss", "warn", "mute", "suspend"}
)

const playerReportJSON = `jsonb_build_object('id',p.id,'status',p.status,
	'reporterId',p.reporter_id,'reporterUsername',p.reporter_username,'reporterName',COALESCE(ru.display_name,''),
	'targetId',p.target_id,'targetUsername',p.target_username,'targetName',COALESCE(tu.display_name,''),
	'titleId',p.title_id,'titleKey',p.title_key,'title',p.title,'note',p.note,'matchId',p.match_id,'contentId',p.content_id,'gameKey',p.game_key,
	'reviewer',p.reviewer,'reviewStartedAt',p.review_started_at,'notes',p.notes,'action',p.action,'restrictionId',p.restriction_id,
	'resolvedBy',p.resolved_by,'resolvedAt',p.resolved_at,'resolutionNote',p.resolution_note,
	'against',(SELECT count(*) FROM base_player_report o WHERE o.target_id=p.target_id),'createdAt',p.created_at,'updatedAt',p.updated_at)`

const playerReportFrom = ` FROM base_player_report p LEFT JOIN users ru ON ru.id=p.reporter_id LEFT JOIN users tu ON tu.id=p.target_id`

func cleanContext(field, v string, max int) (string, error) {
	v = strings.TrimSpace(v)
	if len(v) > max {
		return "", badRequest("INVALID_CONTEXT", field, "at most "+strconv.Itoa(max)+" characters")
	}
	for _, r := range v {
		if r < 0x20 || r == 0x7f {
			return "", badRequest("INVALID_CONTEXT", field, "control characters are not allowed")
		}
	}
	return v, nil
}

func (s *service) rpcReportPlayer(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		UserID    string `json:"userId"`
		TitleKey  string `json:"titleKey"`
		Note      string `json:"note"`
		MatchID   string `json:"matchId"`
		ContentID string `json:"contentId"`
		GameKey   string `json:"gameKey"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	target := strings.TrimSpace(in.UserID)
	if target == userID {
		return nil, badRequest("CANNOT_REPORT_SELF", "userId", "you cannot report yourself")
	}
	if err := s.userExists(ctx, s.db, target); err != nil {
		return nil, err
	}
	note := strings.TrimSpace(in.Note)
	if len([]rune(note)) > playerReportNoteMax {
		return nil, badRequest("TEXT_TOO_LONG", "note", "at most 1000 letters")
	}
	rules, err := loadReportRules(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if rules.NoteRequired && note == "" {
		return nil, badRequest("NOTE_REQUIRED", "note", "describe what happened")
	}
	if rules.MinAccountHours > 0 {
		var young bool
		if err := s.db.QueryRowContext(ctx, `SELECT create_time > now() - $2::INT * INTERVAL '1 hour' FROM users WHERE id=$1`, userID, rules.MinAccountHours).Scan(&young); err != nil {
			return nil, err
		}
		if young {
			return nil, fail(http.StatusForbidden, "ACCOUNT_TOO_NEW", "", "this account is too new to report players")
		}
	}
	limit, _, err := s.reportLimitFor(ctx, rules, userID)
	if err != nil {
		return nil, err
	}
	matchID, err := cleanContext("matchId", in.MatchID, 128)
	if err != nil {
		return nil, err
	}
	contentID, err := cleanContext("contentId", in.ContentID, 128)
	if err != nil {
		return nil, err
	}
	gameKey, err := cleanContext("gameKey", in.GameKey, 64)
	if err != nil {
		return nil, err
	}
	var titleID, titleName, titleKey string
	err = s.db.QueryRowContext(ctx, `SELECT id::text, key, name::text FROM base_report_title WHERE key=$1 AND is_active`, strings.TrimSpace(in.TitleKey)).Scan(&titleID, &titleKey, &titleName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, badRequest("REPORT_TITLE_NOT_FOUND", "titleKey", "pick one of the report reasons")
	}
	if err != nil {
		return nil, err
	}
	var id string
	dup := false
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('base-player-report:' || $1 || ':' || $2 || ':' || $3))`, userID, target, titleID); err != nil {
			return err
		}
		if rules.DuplicateHours > 0 {
			err := tx.QueryRowContext(ctx, `SELECT id::text FROM base_player_report WHERE reporter_id=$1 AND target_id=$2 AND title_id=$3 AND status IN ('open','reviewing')
				AND created_at > now() - $4::INT * INTERVAL '1 hour' ORDER BY created_at DESC LIMIT 1`, userID, target, titleID, rules.DuplicateHours).Scan(&id)
			if err == nil {
				dup = true
				return nil
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}
		if limit > 0 {
			d, err := clusterkit.Allow(ctx, tx, "player-report:"+userID, int64(limit), playerReportWindow)
			if err != nil {
				return err
			}
			if !d.Allowed {
				return fail(http.StatusTooManyRequests, "TOO_MANY_REPORTS", "", "too many reports, try again later")
			}
		}
		id = uuid.Must(uuid.NewV4()).String()
		_, err = tx.ExecContext(ctx, `INSERT INTO base_player_report (id, reporter_id, reporter_username, target_id, target_username, title_id, title_key, title, note, match_id, content_id, game_key)
			VALUES ($1, $2, COALESCE((SELECT username FROM users WHERE id=$2), ''), $3, COALESCE((SELECT username FROM users WHERE id=$3), ''), $4, $5, $6, $7, $8, $9, $10)`,
			id, userID, target, titleID, titleKey, localized.Parse(titleName).JSON(), note, matchID, contentID, gameKey)
		return err
	})
	if err != nil {
		return nil, err
	}
	if !dup {
		data := map[string]any{"reason": titleKey, "kind": "player"}
		if gameKey != "" {
			data["gameKey"] = gameKey
		}
		s.record(ctx, activity.Entry{UserID: userID, Kind: activity.KindReport, Action: "filed", RefID: id, Actor: "player", Data: data})
		s.record(ctx, activity.Entry{UserID: target, Kind: activity.KindReport, Action: "received", RefID: id, Actor: "player", Data: data})
	}
	return map[string]any{"reportId": id, "duplicate": dup}, nil
}

func (s *service) rpcMyReports(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang  string `json:"lang"`
		Page  int    `json:"page"`
		Limit int    `json:"limit"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if in.Limit < 1 || in.Limit > 50 {
		in.Limit = 20
	}
	if in.Page < 1 {
		in.Page = 1
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	var total int64
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM base_player_report WHERE reporter_id=$1`, userID).Scan(&total); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.id::text, COALESCE(p.target_id::text,''), p.target_username, COALESCE(u.display_name,''), p.title_key, COALESCE(t.name, p.title)::text,
		p.note, p.match_id, p.content_id, p.game_key, p.status, p.created_at, p.resolved_at
		FROM base_player_report p LEFT JOIN users u ON u.id=p.target_id LEFT JOIN base_report_title t ON t.id=p.title_id
		WHERE p.reporter_id=$1 ORDER BY p.created_at DESC, p.id LIMIT $2 OFFSET $3`, userID, in.Limit, (in.Page-1)*in.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, target, username, display, key, name, note, match, content, game, status string
		var created time.Time
		var resolved sql.NullTime
		if err := rows.Scan(&id, &target, &username, &display, &key, &name, &note, &match, &content, &game, &status, &created, &resolved); err != nil {
			return nil, err
		}
		item := map[string]any{"id": id, "userId": target, "username": username, "displayName": display, "titleKey": key, "title": langs.Resolve(localized.Parse(name), lang),
			"note": note, "matchId": match, "contentId": content, "gameKey": game, "status": status, "createdAt": created.UTC().Format(time.RFC3339), "closedAt": nil}
		if resolved.Valid {
			item["closedAt"] = resolved.Time.UTC().Format(time.RFC3339)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rules, err := loadReportRules(ctx, s.db)
	if err != nil {
		return nil, err
	}
	perHour, trusted, err := s.reportLimitFor(ctx, rules, userID)
	if err != nil {
		return nil, err
	}
	left, err := s.reportsLeft(ctx, userID, perHour)
	if err != nil {
		return nil, err
	}
	limits := map[string]any{"perHour": perHour, "remaining": left, "trusted": trusted, "noteRequired": rules.NoteRequired, "minAccountHours": rules.MinAccountHours, "duplicateHours": rules.DuplicateHours, "noteMax": playerReportNoteMax}
	return map[string]any{"lang": lang, "items": items, "total": total, "page": in.Page, "limit": in.Limit, "limits": limits}, nil
}

func (s *service) playerReportList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	q := r.URL.Query()
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(p.note ILIKE "+n+" OR p.target_username ILIKE "+n+" OR p.reporter_username ILIKE "+n+" OR p.match_id ILIKE "+n+" OR p.content_id ILIKE "+n+" OR p.game_key ILIKE "+n+")")
	}
	if v := q.Get("status"); v != "" {
		if v == "active" {
			where.add("p.status IN ('open','reviewing')")
		} else if !validKeyType(playerReportStatuses, v) {
			s.writeError(w, badRequest("INVALID_FILTER", "status", "status must be open, reviewing, resolved or dismissed"))
			return
		} else {
			where.add("p.status = ?", v)
		}
	}
	for _, f := range []struct{ name, col string }{{"titleId", "p.title_id"}, {"targetId", "p.target_id"}, {"reporterId", "p.reporter_id"}} {
		if v := strings.TrimSpace(q.Get(f.name)); v != "" {
			if _, err := uuid.FromString(v); err != nil {
				s.writeError(w, badRequest("INVALID_FILTER", f.name, "not a valid id"))
				return
			}
			where.add(f.col+" = ?::uuid", v)
		}
	}
	if v := strings.TrimSpace(q.Get("gameKey")); v != "" {
		where.add("p.game_key = ?", v)
	}
	out, err := s.listRaw(r, playerReportFrom, playerReportJSON, "p.created_at DESC, p.id", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var open int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM base_player_report WHERE status IN ('open','reviewing')`).Scan(&open); err != nil {
		s.writeError(w, err)
		return
	}
	out["openTotal"] = open
	writeJSON(w, http.StatusOK, out)
}

func (s *service) playerReportSummary(w http.ResponseWriter, r *http.Request) {
	var open, reviewing int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FILTER (WHERE status='open'), count(*) FILTER (WHERE status='reviewing') FROM base_player_report WHERE status IN ('open','reviewing')`).Scan(&open, &reviewing); err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"open": open, "reviewing": reviewing, "active": open + reviewing})
}

func fetchPlayerReport(ctx context.Context, q queryer, id string) (map[string]any, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("REPORT_NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+playerReportJSON+`::text`+playerReportFrom+` WHERE p.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("REPORT_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	return out, json.Unmarshal([]byte(raw), &out)
}

func (s *service) playerReportGet(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	out, err := fetchPlayerReport(r.Context(), s.db, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	others := []json.RawMessage{}
	if target, _ := out["targetId"].(string); target != "" {
		rows, err := s.db.QueryContext(r.Context(), `SELECT `+playerReportJSON+`::text`+playerReportFrom+` WHERE p.target_id=$1::uuid AND p.id <> $2 ORDER BY p.created_at DESC LIMIT 20`, target, id)
		if err != nil {
			s.writeError(w, err)
			return
		}
		if others, err = rawRows(rows); err != nil {
			s.writeError(w, err)
			return
		}
		if player, err := s.playerView(r.Context(), s.db, target); err == nil {
			var warns, mutes, suspensions, chat, reports int64
			_ = s.db.QueryRowContext(r.Context(), `SELECT count(*) FILTER (WHERE kind='warn'), count(*) FILTER (WHERE kind='mute'), count(*) FILTER (WHERE kind='suspend') FROM base_restriction WHERE user_id=$1`, target).Scan(&warns, &mutes, &suspensions)
			_ = s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM base_chat_report WHERE sender_id=$1`, target).Scan(&chat)
			_ = s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM base_player_report WHERE target_id=$1`, target).Scan(&reports)
			player["counts"] = map[string]int64{"warn": warns, "mute": mutes, "suspend": suspensions, "chatReports": chat, "reports": reports}
			out["target"] = player
		}
	}
	out["others"] = others
	writeJSON(w, http.StatusOK, out)
}

func (s *service) playerReportStatus(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	var in struct {
		Status string `json:"status"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if in.Status != "open" && in.Status != "reviewing" {
		s.writeError(w, badRequest("INVALID_REPORT_STATUS", "status", "status must be open or reviewing; use resolve to close a report"))
		return
	}
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("REPORT_NOT_FOUND"))
		return
	}
	by := consoleUser(r)
	res, err := s.db.ExecContext(r.Context(), `UPDATE base_player_report SET status=$2,
		reviewer = CASE WHEN $2='reviewing' THEN $3 ELSE '' END, review_started_at = CASE WHEN $2='reviewing' THEN now() ELSE NULL END, updated_at=now()
		WHERE id=$1 AND status IN ('open','reviewing')`, id, in.Status, by)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := fetchPlayerReport(r.Context(), s.db, id); err != nil {
			s.writeError(w, err)
			return
		}
		s.writeError(w, conflict("REPORT_CLOSED", "this report was already handled"))
		return
	}
	out, err := fetchPlayerReport(r.Context(), s.db, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) playerReportNote(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	var in struct {
		Text string `json:"text"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	text := strings.TrimSpace(in.Text)
	if text == "" {
		s.writeError(w, badRequest("TEXT_REQUIRED", "text", "write a note"))
		return
	}
	if len([]rune(text)) > 1000 {
		s.writeError(w, badRequest("TEXT_TOO_LONG", "text", "at most 1000 letters"))
		return
	}
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("REPORT_NOT_FOUND"))
		return
	}
	note, _ := json.Marshal([]map[string]any{{"by": consoleUser(r), "at": time.Now().UTC().Format(time.RFC3339), "text": text}})
	res, err := s.db.ExecContext(r.Context(), `UPDATE base_player_report SET notes = notes || $2::jsonb, updated_at=now() WHERE id=$1 AND jsonb_array_length(notes) < $3`, id, string(note), playerReportNoteLimit)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		if _, err := fetchPlayerReport(r.Context(), s.db, id); err != nil {
			s.writeError(w, err)
			return
		}
		s.writeError(w, conflict("TOO_MANY_NOTES", "a report keeps at most 50 notes"))
		return
	}
	out, err := fetchPlayerReport(r.Context(), s.db, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) playerReportResolve(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	var in struct {
		Action      string         `json:"action"`
		EndsAt      string         `json:"endsAt"`
		Minutes     int            `json:"minutes"`
		Reason      localized.Text `json:"reason"`
		Note        string         `json:"note"`
		CloseOthers bool           `json:"closeOthers"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if !validKeyType(playerReportActions, in.Action) {
		s.writeError(w, badRequest("INVALID_MOD_ACTION", "action", "action must be dismiss, warn, mute or suspend"))
		return
	}
	note := strings.TrimSpace(in.Note)
	if len([]rune(note)) > 1000 {
		s.writeError(w, badRequest("TEXT_TOO_LONG", "note", "at most 1000 letters"))
		return
	}
	report, err := fetchPlayerReport(r.Context(), s.db, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if st := report["status"]; st != "open" && st != "reviewing" {
		s.writeError(w, conflict("REPORT_CLOSED", "this report was already handled"))
		return
	}
	target, _ := report["targetId"].(string)
	by := consoleUser(r)
	restricting := in.Action != "dismiss"
	var plan restrictionPlan
	if restricting {
		if target == "" {
			s.writeError(w, badRequest("USER_NOT_FOUND", "action", "the reported player no longer exists"))
			return
		}
		plan, err = s.planRestriction(r.Context(), s.db, target, restrictionInput{Kind: in.Action, EndsAt: in.EndsAt, Minutes: in.Minutes, Reason: in.Reason, Note: note}, by)
		if err != nil {
			s.writeError(w, err)
			return
		}
		plan.reportID = id
	}
	status := "resolved"
	if in.Action == "dismiss" {
		status = "dismissed"
	}
	var restrictionID string
	err = s.inTx(r.Context(), func(tx *sql.Tx) error {
		var locked string
		if err := tx.QueryRowContext(r.Context(), `SELECT status FROM base_player_report WHERE id=$1 FOR UPDATE`, id).Scan(&locked); err != nil {
			return err
		}
		if locked != "open" && locked != "reviewing" {
			return conflict("REPORT_CLOSED", "this report was already handled")
		}
		var rid any
		if restricting {
			v, err := s.applyRestriction(r.Context(), tx, plan)
			if err != nil {
				return err
			}
			restrictionID, rid = v, v
		}
		var targetArg any
		if in.CloseOthers && target != "" {
			targetArg = target
		}
		_, err := tx.ExecContext(r.Context(), `UPDATE base_player_report SET status=$2, action=$3, restriction_id=$4, resolved_by=$5, resolved_at=now(), resolution_note=$6, updated_at=now()
			WHERE (id=$1 OR target_id=$7::uuid) AND status IN ('open','reviewing')`, id, status, in.Action, rid, by, note, targetArg)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	if restricting {
		s.announceRestriction(r.Context(), target, restrictionID)
	}
	out, err := fetchPlayerReport(r.Context(), s.db, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
