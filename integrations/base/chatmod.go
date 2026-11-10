package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/chatfilter"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/restrict"
)

const wordJSON = `jsonb_build_object('id',w.id,'word',w.word,'action',w.action,'match',w.match,'isActive',w.is_active,'note',w.note,'createdBy',w.created_by,'createdAt',w.created_at,'updatedAt',w.updated_at)`

const reportTitleJSON = `jsonb_build_object('id',t.id,'key',t.key,'name',t.name,'description',t.description,'isActive',t.is_active,'sort',t.sort,'used',(SELECT count(*) FROM base_chat_report c WHERE c.title_id=t.id),'createdAt',t.created_at,'updatedAt',t.updated_at)`

const reportJSON = `jsonb_build_object('id',c.id,'source',c.source,'status',c.status,'messageId',c.message_id,'channelId',c.channel_id,'channelLabel',c.channel_label,
	'senderId',c.sender_id,'senderUsername',c.sender_username,'senderName',COALESCE(su.display_name,''),'content',c.content,'messageAt',c.message_at,
	'reporterId',c.reporter_id,'reporterUsername',c.reporter_username,'reporterName',COALESCE(ru.display_name,''),'titleId',c.title_id,'titleKey',COALESCE(t.key,''),'title',c.title,'note',c.note,'matched',c.matched,
	'action',c.action,'restrictionId',c.restriction_id,'messageDeleted',c.message_deleted,'resolvedBy',c.resolved_by,'resolvedAt',c.resolved_at,'resolutionNote',c.resolution_note,
	'reports',(SELECT count(*) FROM base_chat_report o WHERE o.message_id=c.message_id),'createdAt',c.created_at)`

const reportFrom = ` FROM base_chat_report c LEFT JOIN users su ON su.id=c.sender_id LEFT JOIN users ru ON ru.id=c.reporter_id LEFT JOIN base_report_title t ON t.id=c.title_id`

var (
	wordActions = []string{chatfilter.ActionMask, chatfilter.ActionBlock, chatfilter.ActionFlag}
	wordMatches = []string{chatfilter.MatchWord, chatfilter.MatchAny}
)

func withNormalized(raw json.RawMessage) json.RawMessage {
	var m map[string]any
	if json.Unmarshal(raw, &m) != nil {
		return raw
	}
	word, _ := m["word"].(string)
	m["normalized"] = chatfilter.Normalize(word)
	out, _ := json.Marshal(m)
	return out
}

func (s *service) wordList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(w.word ILIKE "+n+" OR w.note ILIKE "+n+")")
	}
	q := r.URL.Query()
	if v := q.Get("action"); v != "" {
		if !validKeyType(wordActions, v) {
			s.writeError(w, badRequest("INVALID_MOD_ACTION", "action", "action must be mask, block or flag"))
			return
		}
		where.add("w.action = ?", v)
	}
	if v := q.Get("isActive"); v == "true" || v == "false" {
		where.add("w.is_active = ?", v == "true")
	}
	out, err := s.listRaw(r, ` FROM base_chat_word w`, wordJSON, "w.created_at DESC, w.id", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if items, ok := out["items"].([]json.RawMessage); ok {
		for i, it := range items {
			items[i] = withNormalized(it)
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func fetchWord(ctx context.Context, q queryer, id string) (json.RawMessage, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+wordJSON+`::text FROM base_chat_word w WHERE w.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return withNormalized(json.RawMessage(raw)), nil
}

func (s *service) wordGet(w http.ResponseWriter, r *http.Request) {
	raw, err := fetchWord(r.Context(), s.db, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, raw)
}

func (s *service) wordWrite(create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Word     string `json:"word"`
			Action   string `json:"action"`
			Match    string `json:"match"`
			IsActive *bool  `json:"isActive"`
			Note     string `json:"note"`
		}
		if err := readBody(r, &in); err != nil {
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
		word := strings.Join(strings.Fields(in.Word), " ")
		if n := len([]rune(word)); n == 0 || n > 64 {
			s.writeError(w, badRequest("INVALID_WORD", "word", "a word is 1 to 64 letters"))
			return
		}
		if chatfilter.Compile([]chatfilter.Rule{{Word: word}}).Len() == 0 {
			s.writeError(w, badRequest("INVALID_WORD", "word", "the word needs at least one letter"))
			return
		}
		if !validKeyType(wordActions, in.Action) {
			s.writeError(w, badRequest("INVALID_MOD_ACTION", "action", "action must be mask, block or flag"))
			return
		}
		if in.Match == "" {
			in.Match = chatfilter.MatchWord
		}
		if !validKeyType(wordMatches, in.Match) {
			s.writeError(w, badRequest("INVALID_MATCH", "match", "match must be word or contains"))
			return
		}
		note := strings.TrimSpace(in.Note)
		if len([]rune(note)) > 500 {
			s.writeError(w, badRequest("TEXT_TOO_LONG", "note", "at most 500 letters"))
			return
		}
		active := in.IsActive == nil || *in.IsActive
		norm := chatfilter.Normalize(word)
		var out json.RawMessage
		err := s.inTx(r.Context(), func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(r.Context(), `SELECT pg_advisory_xact_lock(hashtext('base-chat-words'))`); err != nil {
				return err
			}
			rows, err := tx.QueryContext(r.Context(), `SELECT id::text, word FROM base_chat_word WHERE id <> $1`, id)
			if err != nil {
				return err
			}
			dup := false
			for rows.Next() {
				var other, text string
				if err := rows.Scan(&other, &text); err != nil {
					rows.Close()
					return err
				}
				if chatfilter.Normalize(text) == norm {
					dup = true
				}
			}
			rows.Close()
			if dup {
				return fail(http.StatusConflict, "ALREADY_EXISTS", "word", "this word is already in the list")
			}
			if create {
				_, err = tx.ExecContext(r.Context(), `INSERT INTO base_chat_word (id, word, action, match, is_active, note, created_by) VALUES ($1,$2,$3,$4,$5,$6,$7)`, id, word, in.Action, in.Match, active, note, consoleUser(r))
			} else {
				var res sql.Result
				res, err = tx.ExecContext(r.Context(), `UPDATE base_chat_word SET word=$2, action=$3, match=$4, is_active=$5, note=$6, updated_at=now() WHERE id=$1`, id, word, in.Action, in.Match, active, note)
				if err == nil {
					if n, _ := res.RowsAffected(); n == 0 {
						return notFound("NOT_FOUND")
					}
				}
			}
			if err != nil {
				return err
			}
			out, err = fetchWord(r.Context(), tx, id)
			return err
		})
		if err != nil {
			if _, ok := isUniqueViolation(err); ok {
				err = fail(http.StatusConflict, "ALREADY_EXISTS", "word", "this word is already in the list")
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

func (s *service) wordDelete(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM base_chat_word WHERE id=$1`, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) loadFilter(ctx context.Context, q interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}) (*chatfilter.Filter, error) {
	rows, err := q.QueryContext(ctx, `SELECT id::text, word, action, match FROM base_chat_word WHERE is_active ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rules := []chatfilter.Rule{}
	for rows.Next() {
		var r chatfilter.Rule
		if err := rows.Scan(&r.ID, &r.Word, &r.Action, &r.Match); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return chatfilter.Compile(rules), rows.Err()
}

func (s *service) wordTest(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Text string `json:"text"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if len([]rune(in.Text)) > 2000 {
		s.writeError(w, badRequest("TEXT_TOO_LONG", "text", "at most 2000 letters"))
		return
	}
	f, err := s.loadFilter(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	res := f.Apply(in.Text)
	hits := res.Hits
	if hits == nil {
		hits = []chatfilter.Hit{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"text": res.Text, "normalized": chatfilter.Normalize(in.Text), "changed": res.Changed, "blocked": res.Blocked, "flagged": res.Flagged, "hits": hits})
}

func (s *service) reportTitleList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(t.key ILIKE "+n+" OR "+translatedSearch("t.name", n)+")")
	}
	if v := r.URL.Query().Get("isActive"); v == "true" || v == "false" {
		where.add("t.is_active = ?", v == "true")
	}
	out, err := s.listRaw(r, ` FROM base_report_title t`, reportTitleJSON, "t.sort, t.key", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func fetchReportTitle(ctx context.Context, q queryer, id string) (json.RawMessage, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+reportTitleJSON+`::text FROM base_report_title t WHERE t.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func (s *service) reportTitleGet(w http.ResponseWriter, r *http.Request) {
	raw, err := fetchReportTitle(r.Context(), s.db, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, raw)
}

func (s *service) reportTitleWrite(create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Key         string         `json:"key"`
			Name        localized.Text `json:"name"`
			Description localized.Text `json:"description"`
			IsActive    *bool          `json:"isActive"`
			Sort        *int           `json:"sort"`
		}
		if err := readBody(r, &in); err != nil {
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
			key, err := cleanKey("key", in.Key)
			if err != nil {
				return err
			}
			langs, err := s.languages(r.Context(), tx)
			if err != nil {
				return err
			}
			name, err := cleanText(langs, "name", in.Name, true, 100)
			if err != nil {
				return err
			}
			desc, err := cleanText(langs, "description", in.Description, false, 300)
			if err != nil {
				return err
			}
			active := in.IsActive == nil || *in.IsActive
			sortN := 100
			if in.Sort != nil {
				sortN = *in.Sort
			}
			if create {
				_, err = tx.ExecContext(r.Context(), `INSERT INTO base_report_title (id, key, name, description, is_active, sort) VALUES ($1,$2,$3,$4,$5,$6)`, id, key, name.JSON(), desc.JSON(), active, sortN)
			} else {
				var res sql.Result
				res, err = tx.ExecContext(r.Context(), `UPDATE base_report_title SET key=$2, name=$3, description=$4, is_active=$5, sort=$6, updated_at=now() WHERE id=$1`, id, key, name.JSON(), desc.JSON(), active, sortN)
				if err == nil {
					if n, _ := res.RowsAffected(); n == 0 {
						return notFound("NOT_FOUND")
					}
				}
			}
			if err != nil {
				return err
			}
			out, err = fetchReportTitle(r.Context(), tx, id)
			return err
		})
		if err != nil {
			s.writeError(w, mapWriteError(err))
			return
		}
		status := http.StatusOK
		if create {
			status = http.StatusCreated
		}
		writeRaw(w, status, out)
	}
}

func (s *service) reportTitleDelete(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM base_report_title WHERE id=$1`, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) reportList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	q := r.URL.Query()
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(c.content ILIKE "+n+" OR c.sender_username ILIKE "+n+" OR c.reporter_username ILIKE "+n+" OR c.note ILIKE "+n+")")
	}
	if v := q.Get("status"); v != "" {
		if v != "open" && v != "dismissed" && v != "resolved" {
			s.writeError(w, badRequest("INVALID_FILTER", "status", "status must be open, dismissed or resolved"))
			return
		}
		where.add("c.status = ?", v)
	}
	if v := q.Get("source"); v != "" {
		if v != "player" && v != "filter" {
			s.writeError(w, badRequest("INVALID_FILTER", "source", "source must be player or filter"))
			return
		}
		where.add("c.source = ?", v)
	}
	if v := strings.TrimSpace(q.Get("titleId")); v != "" {
		if _, err := uuid.FromString(v); err != nil {
			s.writeError(w, badRequest("INVALID_FILTER", "titleId", "report title is not valid"))
			return
		}
		where.add("c.title_id = ?::uuid", v)
	}
	if v := strings.TrimSpace(q.Get("senderId")); v != "" {
		if _, err := uuid.FromString(v); err != nil {
			s.writeError(w, badRequest("USER_NOT_FOUND", "senderId", "player id is not valid"))
			return
		}
		where.add("c.sender_id = ?::uuid", v)
	}
	out, err := s.listRaw(r, reportFrom, reportJSON, "c.created_at DESC, c.id", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var open int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM base_chat_report WHERE status='open'`).Scan(&open); err != nil {
		s.writeError(w, err)
		return
	}
	out["openTotal"] = open
	writeJSON(w, http.StatusOK, out)
}

func (s *service) reportSummary(w http.ResponseWriter, r *http.Request) {
	var open, player, filter int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*), count(*) FILTER (WHERE source='player'), count(*) FILTER (WHERE source='filter') FROM base_chat_report WHERE status='open'`).Scan(&open, &player, &filter); err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"open": open, "player": player, "filter": filter})
}

func fetchReport(ctx context.Context, q queryer, id string) (map[string]any, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("REPORT_NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, `SELECT `+reportJSON+`::text`+reportFrom+` WHERE c.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("REPORT_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	return out, json.Unmarshal([]byte(raw), &out)
}

func (s *service) reportGet(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	out, err := fetchReport(r.Context(), s.db, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var ctxRaw string
	if err := s.db.QueryRowContext(r.Context(), `SELECT context::text FROM base_chat_report WHERE id=$1`, id).Scan(&ctxRaw); err != nil {
		s.writeError(w, err)
		return
	}
	out["context"] = json.RawMessage(ctxRaw)
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+reportJSON+`::text`+reportFrom+` WHERE c.message_id=$1::uuid AND c.id <> $2 ORDER BY c.created_at DESC LIMIT 50`, out["messageId"], id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	others, err := rawRows(rows)
	if err != nil {
		s.writeError(w, err)
		return
	}
	out["others"] = others
	if sender, _ := out["senderId"].(string); sender != "" {
		player, err := s.playerView(r.Context(), s.db, sender)
		if err == nil {
			var warns, mutes, suspensions, reports int64
			_ = s.db.QueryRowContext(r.Context(), `SELECT count(*) FILTER (WHERE kind='warn'), count(*) FILTER (WHERE kind='mute'), count(*) FILTER (WHERE kind='suspend') FROM base_restriction WHERE user_id=$1`, sender).Scan(&warns, &mutes, &suspensions)
			_ = s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM base_chat_report WHERE sender_id=$1`, sender).Scan(&reports)
			player["counts"] = map[string]int64{"warn": warns, "mute": mutes, "suspend": suspensions, "reports": reports}
			out["sender"] = player
		}
	}
	writeJSON(w, http.StatusOK, out)
}

var reportActions = []string{"dismiss", "delete", "warn", "mute", "suspend"}

func (s *service) reportResolve(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	var in struct {
		Action        string         `json:"action"`
		DeleteMessage bool           `json:"deleteMessage"`
		EndsAt        string         `json:"endsAt"`
		Minutes       int            `json:"minutes"`
		Reason        localized.Text `json:"reason"`
		Note          string         `json:"note"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if !validKeyType(reportActions, in.Action) {
		s.writeError(w, badRequest("INVALID_MOD_ACTION", "action", "action must be dismiss, delete, warn, mute or suspend"))
		return
	}
	note := strings.TrimSpace(in.Note)
	if len([]rune(note)) > 1000 {
		s.writeError(w, badRequest("TEXT_TOO_LONG", "note", "at most 1000 letters"))
		return
	}
	report, err := fetchReport(r.Context(), s.db, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	if report["status"] != "open" {
		s.writeError(w, conflict("REPORT_CLOSED", "this report was already handled"))
		return
	}
	sender, _ := report["senderId"].(string)
	messageID, _ := report["messageId"].(string)
	by := consoleUser(r)
	var plan restrictionPlan
	restricting := in.Action == "warn" || in.Action == "mute" || in.Action == "suspend"
	if restricting {
		if sender == "" {
			s.writeError(w, badRequest("USER_NOT_FOUND", "action", "the sender no longer exists"))
			return
		}
		plan, err = s.planRestriction(r.Context(), s.db, sender, restrictionInput{Kind: in.Action, EndsAt: in.EndsAt, Minutes: in.Minutes, Reason: in.Reason, Note: note}, by)
		if err != nil {
			s.writeError(w, err)
			return
		}
		plan.reportID = id
	}
	deleting := in.Action == "delete" || (restricting && in.DeleteMessage)
	deleted := false
	if deleting {
		deleted, err = s.removeMessage(r.Context(), messageID)
		if err != nil {
			s.writeError(w, err)
			return
		}
	}
	status := "resolved"
	if in.Action == "dismiss" {
		status = "dismissed"
	}
	var restrictionID string
	err = s.inTx(r.Context(), func(tx *sql.Tx) error {
		var locked string
		if err := tx.QueryRowContext(r.Context(), `SELECT status FROM base_chat_report WHERE id=$1 FOR UPDATE`, id).Scan(&locked); err != nil {
			return err
		}
		if locked != "open" {
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
		_, err := tx.ExecContext(r.Context(), `UPDATE base_chat_report SET status=$3, action=$4, restriction_id=$5, message_deleted = message_deleted OR $6, resolved_by=$7, resolved_at=now(), resolution_note=$8, updated_at=now()
			WHERE (id=$1 OR message_id=$2::uuid) AND status='open'`, id, messageID, status, in.Action, rid, deleted, by, note)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	if restricting {
		s.announceRestriction(r.Context(), sender, restrictionID)
	}
	out, err := fetchReport(r.Context(), s.db, id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) removeMessage(ctx context.Context, messageID string) (bool, error) {
	var mode int
	var subject, descriptor, label, sender, username string
	err := s.db.QueryRowContext(ctx, `SELECT stream_mode, stream_subject::text, stream_descriptor::text, stream_label, sender_id::text, username FROM message WHERE id=$1`, messageID).
		Scan(&mode, &subject, &descriptor, &label, &sender, &username)
	if errors.Is(err, sql.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	channelID := channelIDOf(mode, subject, descriptor, label)
	if _, err := s.nk.ChannelMessageRemove(ctx, channelID, messageID, sender, username, true); err != nil {
		s.logger.Warn("Base report message removal failed: %v", err)
		res, err := s.db.ExecContext(ctx, `DELETE FROM message WHERE id=$1`, messageID)
		if err != nil {
			return false, err
		}
		n, _ := res.RowsAffected()
		return n > 0, nil
	}
	return true, nil
}

func channelIDOf(mode int, subject, descriptor, label string) string {
	if subject == systemUserID {
		subject = ""
	}
	if descriptor == systemUserID {
		descriptor = ""
	}
	return itoa(mode) + "." + subject + "." + descriptor + "." + label
}

type contextLine struct {
	ID        string    `json:"id"`
	SenderID  string    `json:"senderId"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	Reported  bool      `json:"reported,omitempty"`
}

func (s *service) messageContext(ctx context.Context, messageID string, before, after int) (map[string]any, []contextLine, error) {
	var mode int
	var subject, descriptor, label, sender, username, content string
	var created time.Time
	err := s.db.QueryRowContext(ctx, `SELECT stream_mode, stream_subject::text, stream_descriptor::text, stream_label, sender_id::text, username, content::text, create_time FROM message WHERE id=$1 AND code=0`, messageID).
		Scan(&mode, &subject, &descriptor, &label, &sender, &username, &content, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, notFound("MESSAGE_NOT_FOUND")
	}
	if err != nil {
		return nil, nil, err
	}
	msg := map[string]any{"mode": mode, "subject": subject, "descriptor": descriptor, "label": label, "senderId": sender, "username": username, "content": content, "createdAt": created, "channelId": channelIDOf(mode, subject, descriptor, label)}
	lines := []contextLine{}
	collect := func(query string, n int) error {
		if n <= 0 {
			return nil
		}
		rows, err := s.db.QueryContext(ctx, query, mode, subject, descriptor, label, created, messageID, n)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var l contextLine
			if err := rows.Scan(&l.ID, &l.SenderID, &l.Username, &l.Content, &l.CreatedAt); err != nil {
				return err
			}
			lines = append(lines, l)
		}
		return rows.Err()
	}
	if err := collect(`SELECT id::text, sender_id::text, username, content::text, create_time FROM (SELECT * FROM message WHERE stream_mode=$1 AND stream_subject=$2::uuid AND stream_descriptor=$3::uuid AND stream_label=$4 AND code=0 AND (create_time, id) < ($5, $6::uuid) ORDER BY create_time DESC, id DESC LIMIT $7) m ORDER BY create_time, id`, before); err != nil {
		return nil, nil, err
	}
	lines = append(lines, contextLine{ID: messageID, SenderID: sender, Username: username, Content: content, CreatedAt: created, Reported: true})
	if err := collect(`SELECT id::text, sender_id::text, username, content::text, create_time FROM message WHERE stream_mode=$1 AND stream_subject=$2::uuid AND stream_descriptor=$3::uuid AND stream_label=$4 AND code=0 AND (create_time, id) > ($5, $6::uuid) ORDER BY create_time, id LIMIT $7`, after); err != nil {
		return nil, nil, err
	}
	return msg, lines, nil
}

func messageText(content string) string {
	var m map[string]any
	if json.Unmarshal([]byte(content), &m) == nil {
		for _, k := range []string{"text", "message", "body", "content"} {
			if v, ok := m[k].(string); ok {
				return v
			}
		}
	}
	return content
}

func (s *service) insertReport(ctx context.Context, source, messageID, reporter, reporterName, titleID string, title localized.Text, note string, matched []chatfilter.Hit, contextAfter int) (string, bool, error) {
	msg, lines, err := s.messageContext(ctx, messageID, 5, contextAfter)
	if err != nil {
		return "", false, err
	}
	for i := range lines {
		lines[i].Content = messageText(lines[i].Content)
	}
	ctxJSON, _ := json.Marshal(lines)
	if matched == nil {
		matched = []chatfilter.Hit{}
	}
	matchedJSON, _ := json.Marshal(matched)
	id := uuid.Must(uuid.NewV4()).String()
	var rep, tid any
	if reporter != "" {
		rep = reporter
	}
	if titleID != "" {
		tid = titleID
	}
	conflictTarget := `(message_id) WHERE source = 'filter'`
	if source == "player" {
		conflictTarget = `(message_id, reporter_id) WHERE source = 'player'`
	}
	res, err := s.db.ExecContext(ctx, `INSERT INTO base_chat_report (id, source, message_id, channel_id, channel_label, sender_id, sender_username, content, message_at, context, reporter_id, reporter_username, title_id, title, note, matched)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) ON CONFLICT `+conflictTarget+` DO NOTHING`,
		id, source, messageID, msg["channelId"], msg["label"], msg["senderId"], msg["username"], messageText(msg["content"].(string)), msg["createdAt"], string(ctxJSON), rep, reporterName, tid, title.JSON(), note, string(matchedJSON))
	if err != nil {
		return "", false, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		var existing string
		_ = s.db.QueryRowContext(ctx, `SELECT id::text FROM base_chat_report WHERE message_id=$1 AND source=$2 AND ($3::uuid IS NULL OR reporter_id=$3::uuid) ORDER BY created_at LIMIT 1`, messageID, source, rep).Scan(&existing)
		return existing, true, nil
	}
	return id, false, nil
}

func restrictionNotice(ctx context.Context, s *service, userID, kind, lang string) (any, error) {
	r, err := restrict.Active(ctx, s.db, userID, kind)
	if err != nil || r == nil {
		return nil, err
	}
	return r.Notice(ctx, s.db, lang).Map(), nil
}
