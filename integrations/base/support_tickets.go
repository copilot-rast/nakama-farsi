package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/inapp"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	notifySupport    = 1210
	ticketAlphabet   = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	maxTicketBody    = 5000
	maxSubject       = 150
	autoCloseLease   = "base-support-autoclose"
	supportCodeShort = "support"
)

var (
	ticketStatuses   = []string{"open", "awaiting_staff", "awaiting_player", "resolved", "closed"}
	ticketPriorities = []string{"low", "normal", "high", "urgent"}
)

func oneOfList(v string, list []string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func finished(status string) bool { return status == "resolved" || status == "closed" }

func (s *service) allow(ctx context.Context, key string, limit int, window time.Duration, code string) error {
	d, err := clusterkit.Allow(ctx, s.db, key, int64(limit), window)
	if err != nil {
		return err
	}
	if !d.Allowed {
		return fail(http.StatusTooManyRequests, code, "", "too many requests, try again later")
	}
	return nil
}

func ticketNumber() string {
	b := make([]byte, 6)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(ticketAlphabet))))
		b[i] = ticketAlphabet[n.Int64()]
	}
	return "T-" + string(b)
}

func preview(body string) string {
	body = strings.Join(strings.Fields(body), " ")
	if utf8.RuneCountInString(body) <= 140 {
		return body
	}
	r := []rune(body)
	return string(r[:139]) + "…"
}

func cleanBody(body string, required bool) (string, error) {
	body = strings.TrimSpace(strings.ReplaceAll(body, "\r\n", "\n"))
	if body == "" && required {
		return "", badRequest("BODY_REQUIRED", "body", "write a message")
	}
	if utf8.RuneCountInString(body) > maxTicketBody {
		return "", badRequest("BODY_TOO_LONG", "body", "message is at most 5000 characters")
	}
	return body, nil
}

const ticketJSON = `jsonb_build_object('id',t.id,'number',t.number,'userId',t.user_id,'username',u.username,'displayName',u.display_name,'subject',t.subject,
	'departmentId',t.department_id,'department',(SELECT jsonb_build_object('id',d.id,'key',d.key,'name',d.name) FROM base_support_department d WHERE d.id=t.department_id),
	'priority',t.priority,'status',t.status,'assignee',t.assignee,'source',t.source,'chatId',t.chat_id,'lang',t.lang,'messageCount',t.message_count,
	'lastMessageAt',t.last_message_at,'lastMessageBy',t.last_message_by,'lastPreview',t.last_preview,'lastPlayerMessageAt',t.last_player_message_at,'lastStaffMessageAt',t.last_staff_message_at,
	'playerSeenAt',t.player_seen_at,'dueAt',t.due_at,'overdue',(t.status IN ('open','awaiting_staff') AND t.due_at < now()),'firstResponseAt',t.first_response_at,'resolvedAt',t.resolved_at,'closedAt',t.closed_at,
	'rating',t.rating,'ratingComment',t.rating_comment,'ratedAt',t.rated_at,'createdBy',t.created_by,'createdAt',t.created_at,'updatedAt',t.updated_at)`

const ticketFrom = ` FROM base_support_ticket t JOIN users u ON u.id=t.user_id`

type ticketRow struct {
	ID, Number, UserID, Subject, Status, Priority, Assignee, Lang string
	DepartmentID                                                  sql.NullString
	FirstResponse                                                 sql.NullTime
	Rating                                                        sql.NullInt64
}

func loadTicketForUpdate(ctx context.Context, tx *sql.Tx, id string) (ticketRow, error) {
	var t ticketRow
	if _, err := uuid.FromString(id); err != nil {
		return t, notFound("TICKET_NOT_FOUND")
	}
	err := tx.QueryRowContext(ctx, `SELECT id::text, number, user_id::text, subject, status, priority, assignee, lang, department_id::text, first_response_at, rating FROM base_support_ticket WHERE id=$1 FOR UPDATE`, id).
		Scan(&t.ID, &t.Number, &t.UserID, &t.Subject, &t.Status, &t.Priority, &t.Assignee, &t.Lang, &t.DepartmentID, &t.FirstResponse, &t.Rating)
	if errors.Is(err, sql.ErrNoRows) {
		return t, notFound("TICKET_NOT_FOUND")
	}
	return t, err
}

func addTicketEvent(ctx context.Context, tx *sql.Tx, ticketID, authorType, author, event string, data map[string]any) error {
	raw, _ := json.Marshal(data)
	if data == nil {
		raw = []byte("{}")
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO base_support_message (id, ticket_id, kind, author_type, author, event, event_data) VALUES ($1,$2,'event',$3,$4,$5,$6::jsonb)`,
		uuid.Must(uuid.NewV4()).String(), ticketID, authorType, author, event, string(raw))
	return err
}

func refreshDue(ctx context.Context, tx *sql.Tx, ticketID string) error {
	_, err := tx.ExecContext(ctx, `UPDATE base_support_ticket t SET due_at = CASE WHEN t.status IN ('open','awaiting_staff')
		THEN COALESCE(t.last_player_message_at, t.created_at) + make_interval(hours => COALESCE((SELECT d.first_response_hours FROM base_support_department d WHERE d.id=t.department_id), 24)) ELSE NULL END
		WHERE t.id=$1`, ticketID)
	return err
}

type newTicket struct {
	UserID       string
	Subject      string
	DepartmentID *string
	Priority     string
	Body         string
	Attachments  []attachment
	Source       string
	ChatID       *string
	CreatedBy    string
	Lang         string
	StaffAuthor  string
}

func (s *service) insertTicket(ctx context.Context, tx *sql.Tx, in newTicket) (string, string, error) {
	id := uuid.Must(uuid.NewV4()).String()
	status, by := "open", "player"
	if in.Source == "staff" {
		status, by = "awaiting_player", "staff"
	}
	assignee := ""
	if in.StaffAuthor != "" {
		assignee = in.StaffAuthor
	} else if in.DepartmentID != nil {
		_ = tx.QueryRowContext(ctx, `SELECT a FROM base_support_department d, unnest(d.agents) a WHERE d.id=$1
			ORDER BY (SELECT count(*) FROM base_support_ticket t WHERE t.assignee=a AND t.status NOT IN ('resolved','closed')), a LIMIT 1`, *in.DepartmentID).Scan(&assignee)
	}
	var number string
	for attempt := 0; ; attempt++ {
		number = ticketNumber()
		_, err := tx.ExecContext(ctx, `SAVEPOINT ticket_number`)
		if err != nil {
			return "", "", err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO base_support_ticket (id, number, user_id, subject, department_id, priority, status, assignee, source, chat_id, lang, created_by, message_count, last_message_at, last_message_by, last_preview, last_player_message_at, last_staff_message_at, first_response_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,1,now(),$13,$14,CASE WHEN $13='player' THEN now() END,CASE WHEN $13='staff' THEN now() END,CASE WHEN $13='staff' THEN now() END)`,
			id, number, in.UserID, in.Subject, in.DepartmentID, in.Priority, status, assignee, in.Source, in.ChatID, in.Lang, in.CreatedBy, by, preview(in.Body))
		if c, unique := isUniqueViolation(err); unique && strings.Contains(c, "number") && attempt < 5 {
			if _, err := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT ticket_number`); err != nil {
				return "", "", err
			}
			continue
		}
		if isForeignKeyViolation(err) {
			return "", "", notFound("USER_NOT_FOUND")
		}
		if err != nil {
			return "", "", err
		}
		break
	}
	author := in.UserID
	if by == "staff" {
		author = in.StaffAuthor
	}
	if in.Attachments == nil {
		in.Attachments = []attachment{}
	}
	raw, _ := json.Marshal(in.Attachments)
	if _, err := tx.ExecContext(ctx, `INSERT INTO base_support_message (id, ticket_id, kind, author_type, author, body, attachments) VALUES ($1,$2,'message',$3,$4,$5,$6::jsonb)`,
		uuid.Must(uuid.NewV4()).String(), id, by, author, in.Body, string(raw)); err != nil {
		return "", "", err
	}
	if assignee != "" && in.StaffAuthor == "" {
		if err := addTicketEvent(ctx, tx, id, "system", "", "assigned", map[string]any{"to": assignee, "auto": true}); err != nil {
			return "", "", err
		}
	}
	if err := refreshDue(ctx, tx, id); err != nil {
		return "", "", err
	}
	return id, number, nil
}

func (s *service) departmentByKeyOrID(ctx context.Context, q queryer, key, id string, activeOnly bool) (*string, error) {
	key, id = strings.TrimSpace(key), strings.TrimSpace(id)
	if key == "" && id == "" {
		var first string
		err := q.QueryRowContext(ctx, `SELECT id::text FROM base_support_department WHERE is_active ORDER BY sort, key LIMIT 1`).Scan(&first)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return &first, nil
	}
	var got string
	var active bool
	var err error
	if id != "" {
		if _, perr := uuid.FromString(id); perr != nil {
			return nil, badRequest("DEPARTMENT_NOT_FOUND", "departmentId", "department does not exist")
		}
		err = q.QueryRowContext(ctx, `SELECT id::text, is_active FROM base_support_department WHERE id=$1`, id).Scan(&got, &active)
	} else {
		err = q.QueryRowContext(ctx, `SELECT id::text, is_active FROM base_support_department WHERE key=$1`, key).Scan(&got, &active)
	}
	if errors.Is(err, sql.ErrNoRows) || (err == nil && activeOnly && !active) {
		return nil, badRequest("DEPARTMENT_NOT_FOUND", "departmentId", "department does not exist")
	}
	if err != nil {
		return nil, err
	}
	return &got, nil
}

var supportTexts = map[string]localized.Text{
	"reply.title":    {"fa": "پاسخ تازه از پشتیبانی", "en": "New reply from support", "ar": "رد جديد من الدعم"},
	"reply.body":     {"fa": "به تیکت «{subject}» ({number}) پاسخ داده شد.", "en": "Your ticket “{subject}” ({number}) has a new reply.", "ar": "تذكرتك «{subject}» ({number}) لديها رد جديد."},
	"resolved.title": {"fa": "تیکت شما حل شد", "en": "Your ticket is solved", "ar": "تم حل تذكرتك"},
	"resolved.body":  {"fa": "تیکت «{subject}» ({number}) حل شد. لطفاً به پاسخ پشتیبانی امتیاز بدهید.", "en": "Ticket “{subject}” ({number}) is solved. Please rate the answer.", "ar": "تم حل التذكرة «{subject}» ({number}). يرجى تقييم الرد."},
	"closed.title":   {"fa": "تیکت شما بسته شد", "en": "Your ticket is closed", "ar": "تم إغلاق تذكرتك"},
	"closed.body":    {"fa": "تیکت «{subject}» ({number}) بسته شد.", "en": "Ticket “{subject}” ({number}) is closed.", "ar": "تم إغلاق التذكرة «{subject}» ({number})."},
	"open.label":     {"fa": "دیدن تیکت", "en": "Open ticket", "ar": "عرض التذكرة"},
}

func fill(t localized.Text, vars map[string]string) map[string]string {
	out := map[string]string{}
	for lang, v := range t {
		for k, val := range vars {
			v = strings.ReplaceAll(v, "{"+k+"}", val)
		}
		out[lang] = v
	}
	return out
}

func (s *service) notifyTicket(ctx context.Context, t ticketRow, event, by string) {
	_ = s.nk.NotificationSend(ctx, t.UserID, supportCodeShort, map[string]any{"type": "ticket", "ticketId": t.ID, "number": t.Number, "event": event}, notifySupport, "", false)
	key := map[string]string{"reply": "reply", "resolved": "resolved", "closed": "closed"}[event]
	if key == "" || !inapp.Available() {
		return
	}
	vars := map[string]string{"subject": t.Subject, "number": t.Number}
	_, err := inapp.Send(ctx, inapp.Message{UserIDs: []string{t.UserID}, Title: fill(supportTexts[key+".title"], vars), Body: fill(supportTexts[key+".body"], vars),
		Action: &inapp.Action{Type: "goto", Value: "support/ticket/" + t.ID, Label: supportTexts["open.label"]}, CreatedBy: "support:" + by})
	if err != nil {
		s.logger.Warn("Support notification for ticket %s not queued: %v", t.ID, err)
	}
}

func (s *service) record(ctx context.Context, e activity.Entry) {
	activity.Safe(ctx, s.db, s.logger, e)
}

func (s *service) supportTicketList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	q := r.URL.Query()
	where := &sqlWhere{}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(t.number ILIKE "+n+" OR t.subject ILIKE "+n+" OR u.username ILIKE "+n+" OR t.user_id::text ILIKE "+n+")")
	}
	switch v := q.Get("status"); v {
	case "":
	case "active":
		where.add("t.status NOT IN ('resolved','closed')")
	case "waiting":
		where.add("t.status IN ('open','awaiting_staff')")
	default:
		if oneOfList(v, ticketStatuses) {
			where.add("t.status = ?", v)
		}
	}
	if v := q.Get("priority"); oneOfList(v, ticketPriorities) {
		where.add("t.priority = ?", v)
	}
	if v := q.Get("departmentId"); v != "" {
		if _, err := uuid.FromString(v); err == nil {
			where.add("t.department_id = ?::uuid", v)
		}
	}
	switch v := strings.TrimSpace(q.Get("assignee")); v {
	case "":
	case "me":
		where.add("t.assignee = ?", agentName(r))
	case "none":
		where.add("t.assignee = ''")
	default:
		where.add("t.assignee = ?", v)
	}
	if v := q.Get("userId"); v != "" {
		if _, err := uuid.FromString(v); err == nil {
			where.add("t.user_id = ?::uuid", v)
		}
	}
	if q.Get("overdue") == "true" {
		where.add("t.status IN ('open','awaiting_staff') AND t.due_at < now()")
	}
	order := "t.last_message_at DESC"
	if q.Get("sort") == "due" {
		order = "t.due_at ASC NULLS LAST, t.last_message_at DESC"
	} else if q.Get("sort") == "priority" {
		order = "array_position(ARRAY['urgent','high','normal','low'], t.priority), t.last_message_at DESC"
	}
	out, err := s.listRaw(r, ticketFrom, ticketJSON, order, where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) ticketMessages(ctx context.Context, ticketID string, forPlayer bool) ([]map[string]any, error) {
	cond := ""
	if forPlayer {
		cond = ` AND (m.kind='message' OR (m.kind='event' AND m.event IN ('status_changed','rated','created_from_chat')))`
	}
	rows, err := s.db.QueryContext(ctx, `SELECT m.id::text, m.kind, m.author_type, m.author, m.body, m.attachments::text, m.event, m.event_data::text, m.created_at FROM base_support_message m WHERE m.ticket_id=$1`+cond+` ORDER BY m.created_at, m.id`, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, kind, authorType, author, body, atts, event, data string
		var at time.Time
		if err := rows.Scan(&id, &kind, &authorType, &author, &body, &atts, &event, &data, &at); err != nil {
			return nil, err
		}
		list := []attachment{}
		_ = json.Unmarshal([]byte(atts), &list)
		m := map[string]any{"id": id, "kind": kind, "authorType": authorType, "body": body, "attachments": s.signAttachments(list), "createdAt": at.UTC().Format(time.RFC3339)}
		if event != "" {
			m["event"] = event
			m["eventData"] = json.RawMessage(data)
		}
		if !forPlayer || authorType == "player" {
			m["author"] = author
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *service) ticketDetail(ctx context.Context, id string) (map[string]any, error) {
	raw, err := fetchRow(ctx, s.db, `SELECT `+ticketJSON+`::text`+ticketFrom+` WHERE t.id=$1`, id)
	if err != nil {
		if app, ok := asApp(err); ok && app.Code == "NOT_FOUND" {
			return nil, notFound("TICKET_NOT_FOUND")
		}
		return nil, err
	}
	var t map[string]any
	_ = json.Unmarshal(raw, &t)
	msgs, err := s.ticketMessages(ctx, id, false)
	if err != nil {
		return nil, err
	}
	t["messages"] = msgs
	var tickets, chats int
	_ = s.db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM base_support_ticket WHERE user_id=$1), (SELECT count(*) FROM base_live_chat WHERE user_id=$1)`, t["userId"]).Scan(&tickets, &chats)
	t["player"] = map[string]any{"tickets": tickets, "chats": chats}
	return t, nil
}

func (s *service) writeTicket(w http.ResponseWriter, r *http.Request, status int, id string) {
	t, err := s.ticketDetail(r.Context(), id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, status, t)
}

func (s *service) supportTicketGet(w http.ResponseWriter, r *http.Request) {
	s.writeTicket(w, r, http.StatusOK, pathVar(r, "id"))
}

func (s *service) supportTicketCreate(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID       string   `json:"userId"`
		Subject      string   `json:"subject"`
		DepartmentID string   `json:"departmentId"`
		Priority     string   `json:"priority"`
		Body         string   `json:"body"`
		Attachments  []string `json:"attachments"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	var id string
	var tr ticketRow
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		if _, err := uuid.FromString(in.UserID); err != nil {
			return badRequest("USER_NOT_FOUND", "userId", "player does not exist")
		}
		subject := strings.TrimSpace(in.Subject)
		if subject == "" || utf8.RuneCountInString(subject) > maxSubject {
			return badRequest("INVALID_SUBJECT", "subject", "subject is 1 to 150 characters")
		}
		body, err := cleanBody(in.Body, true)
		if err != nil {
			return err
		}
		dept, err := s.departmentByKeyOrID(r.Context(), tx, "", in.DepartmentID, false)
		if err != nil {
			return err
		}
		pr := in.Priority
		if pr == "" {
			pr = "normal"
		}
		if !oneOfList(pr, ticketPriorities) {
			return badRequest("INVALID_PRIORITY", "priority", "unknown priority")
		}
		st, err := loadSupportSettings(r.Context(), tx)
		if err != nil {
			return err
		}
		atts, err := s.claimAttachments(r.Context(), tx, in.Attachments, "", true, st.MaxAttachments)
		if err != nil {
			return err
		}
		lang := localized.UserLangTag(r.Context(), tx, in.UserID)
		var number string
		id, number, err = s.insertTicket(r.Context(), tx, newTicket{UserID: in.UserID, Subject: subject, DepartmentID: dept, Priority: pr, Body: body, Attachments: atts, Source: "staff", CreatedBy: agentName(r), Lang: lang, StaffAuthor: agentName(r)})
		tr = ticketRow{ID: id, Number: number, UserID: in.UserID, Subject: subject}
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.record(r.Context(), activity.Entry{UserID: tr.UserID, Kind: activity.KindTicket, Action: "created_by_staff", RefID: id, Actor: "console:" + agentName(r), Data: map[string]any{"number": tr.Number, "subject": tr.Subject}})
	s.notifyTicket(r.Context(), tr, "reply", agentName(r))
	s.writeTicket(w, r, http.StatusCreated, id)
}

func (s *service) supportTicketMessage(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Body        string   `json:"body"`
		Internal    bool     `json:"internal"`
		Status      string   `json:"status"`
		Attachments []string `json:"attachments"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	me := agentName(r)
	var t ticketRow
	var newStatus string
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		var err error
		if t, err = loadTicketForUpdate(r.Context(), tx, pathVar(r, "id")); err != nil {
			return err
		}
		st, err := loadSupportSettings(r.Context(), tx)
		if err != nil {
			return err
		}
		atts, err := s.claimAttachments(r.Context(), tx, in.Attachments, me, true, st.MaxAttachments)
		if err != nil {
			return err
		}
		body, err := cleanBody(in.Body, len(atts) == 0)
		if err != nil {
			return err
		}
		raw, _ := json.Marshal(atts)
		kind := "message"
		if in.Internal {
			kind = "note"
		}
		if _, err := tx.ExecContext(r.Context(), `INSERT INTO base_support_message (id, ticket_id, kind, author_type, author, body, attachments) VALUES ($1,$2,$3,'staff',$4,$5,$6::jsonb)`,
			uuid.Must(uuid.NewV4()).String(), t.ID, kind, me, body, string(raw)); err != nil {
			return err
		}
		if in.Internal {
			_, err := tx.ExecContext(r.Context(), `UPDATE base_support_ticket SET updated_at=now() WHERE id=$1`, t.ID)
			return err
		}
		if t.Status == "closed" {
			return conflict("TICKET_CLOSED", "this ticket is closed")
		}
		newStatus = in.Status
		if newStatus == "" {
			newStatus = "awaiting_player"
		}
		if newStatus != "awaiting_player" && newStatus != "resolved" && newStatus != "closed" {
			return badRequest("INVALID_STATUS", "status", "status after a reply must be awaiting_player, resolved or closed")
		}
		assignee := t.Assignee
		if assignee == "" {
			assignee = me
			if err := addTicketEvent(r.Context(), tx, t.ID, "staff", me, "assigned", map[string]any{"to": me, "auto": true}); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(r.Context(), `UPDATE base_support_ticket SET status=$2, assignee=$3, message_count=message_count+1, last_message_at=now(), last_message_by='staff', last_preview=$4,
			last_staff_message_at=now(), first_response_at=COALESCE(first_response_at, now()), resolved_at=CASE WHEN $2 IN ('resolved','closed') THEN COALESCE(resolved_at, now()) ELSE resolved_at END,
			closed_at=CASE WHEN $2='closed' THEN now() ELSE closed_at END, updated_at=now() WHERE id=$1`, t.ID, newStatus, assignee, preview(body)); err != nil {
			return err
		}
		if newStatus != t.Status {
			if err := addTicketEvent(r.Context(), tx, t.ID, "staff", me, "status_changed", map[string]any{"from": t.Status, "to": newStatus}); err != nil {
				return err
			}
		}
		return refreshDue(r.Context(), tx, t.ID)
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	if !in.Internal {
		s.record(r.Context(), activity.Entry{UserID: t.UserID, Kind: activity.KindTicket, Action: "staff_reply", RefID: t.ID, Actor: "console:" + me, Data: map[string]any{"number": t.Number, "status": newStatus}})
		ev := "reply"
		if newStatus == "resolved" || newStatus == "closed" {
			ev = newStatus
		}
		s.notifyTicket(r.Context(), t, ev, me)
	}
	s.supportTicketGet(w, r)
}

func (s *service) supportTicketStatus(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Status string `json:"status"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	me := agentName(r)
	var t ticketRow
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		var err error
		if t, err = loadTicketForUpdate(r.Context(), tx, pathVar(r, "id")); err != nil {
			return err
		}
		if !oneOfList(in.Status, ticketStatuses) || in.Status == "open" {
			return badRequest("INVALID_STATUS", "status", "unknown status")
		}
		if in.Status == t.Status {
			return nil
		}
		switch in.Status {
		case "resolved", "closed":
			if t.Status == "closed" {
				return conflict("INVALID_TRANSITION", "a closed ticket cannot change")
			}
		case "awaiting_staff":
			if !finished(t.Status) {
				return conflict("INVALID_TRANSITION", "only a finished ticket can be reopened")
			}
		case "awaiting_player":
			if t.Status != "open" && t.Status != "awaiting_staff" {
				return conflict("INVALID_TRANSITION", "only an open ticket can wait for the player")
			}
		}
		if _, err := tx.ExecContext(r.Context(), `UPDATE base_support_ticket SET status=$2, resolved_at=CASE WHEN $2 IN ('resolved','closed') THEN COALESCE(resolved_at, now()) WHEN $2='awaiting_staff' THEN NULL ELSE resolved_at END,
			closed_at=CASE WHEN $2='closed' THEN now() WHEN $2='awaiting_staff' THEN NULL ELSE closed_at END, updated_at=now() WHERE id=$1`, t.ID, in.Status); err != nil {
			return err
		}
		if err := addTicketEvent(r.Context(), tx, t.ID, "staff", me, "status_changed", map[string]any{"from": t.Status, "to": in.Status}); err != nil {
			return err
		}
		return refreshDue(r.Context(), tx, t.ID)
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	if in.Status != t.Status && finished(in.Status) {
		s.record(r.Context(), activity.Entry{UserID: t.UserID, Kind: activity.KindTicket, Action: in.Status, RefID: t.ID, Actor: "console:" + me, Data: map[string]any{"number": t.Number}})
		s.notifyTicket(r.Context(), t, in.Status, me)
	}
	s.supportTicketGet(w, r)
}

func (s *service) supportTicketAssign(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Assignee     *string `json:"assignee"`
		DepartmentID *string `json:"departmentId"`
		Priority     *string `json:"priority"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	me := agentName(r)
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		t, err := loadTicketForUpdate(r.Context(), tx, pathVar(r, "id"))
		if err != nil {
			return err
		}
		if in.Assignee != nil {
			a := strings.TrimSpace(*in.Assignee)
			if a == "me" {
				a = me
			}
			if len(a) > 128 {
				return badRequest("INVALID_AGENT", "assignee", "unknown agent")
			}
			if a != t.Assignee {
				if _, err := tx.ExecContext(r.Context(), `UPDATE base_support_ticket SET assignee=$2, updated_at=now() WHERE id=$1`, t.ID, a); err != nil {
					return err
				}
				if err := addTicketEvent(r.Context(), tx, t.ID, "staff", me, "assigned", map[string]any{"from": t.Assignee, "to": a}); err != nil {
					return err
				}
			}
		}
		if in.DepartmentID != nil {
			dept, err := s.departmentByKeyOrID(r.Context(), tx, "", *in.DepartmentID, false)
			if err != nil {
				return err
			}
			if dept != nil && (!t.DepartmentID.Valid || t.DepartmentID.String != *dept) {
				if _, err := tx.ExecContext(r.Context(), `UPDATE base_support_ticket SET department_id=$2, updated_at=now() WHERE id=$1`, t.ID, *dept); err != nil {
					return err
				}
				if err := addTicketEvent(r.Context(), tx, t.ID, "staff", me, "department_changed", map[string]any{"from": t.DepartmentID.String, "to": *dept}); err != nil {
					return err
				}
				if err := refreshDue(r.Context(), tx, t.ID); err != nil {
					return err
				}
			}
		}
		if in.Priority != nil && *in.Priority != t.Priority {
			if !oneOfList(*in.Priority, ticketPriorities) {
				return badRequest("INVALID_PRIORITY", "priority", "unknown priority")
			}
			if _, err := tx.ExecContext(r.Context(), `UPDATE base_support_ticket SET priority=$2, updated_at=now() WHERE id=$1`, t.ID, *in.Priority); err != nil {
				return err
			}
			if err := addTicketEvent(r.Context(), tx, t.ID, "staff", me, "priority_changed", map[string]any{"from": t.Priority, "to": *in.Priority}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.supportTicketGet(w, r)
}

func (s *service) autoCloseTickets(ctx context.Context) (int64, error) {
	st, err := loadSupportSettings(ctx, s.db)
	if err != nil {
		return 0, err
	}
	rows, err := s.db.QueryContext(ctx, `UPDATE base_support_ticket SET status='closed', closed_at=now(), resolved_at=COALESCE(resolved_at, now()), due_at=NULL, updated_at=now()
		WHERE (status='awaiting_player' AND last_message_at < now() - make_interval(days => $1)) OR (status='resolved' AND resolved_at < now() - make_interval(days => $1))
		RETURNING id::text, status`, st.AutoCloseDays)
	if err != nil {
		return 0, err
	}
	ids := []string{}
	for rows.Next() {
		var id, status string
		if rows.Scan(&id, &status) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		_ = s.inTx(ctx, func(tx *sql.Tx) error {
			return addTicketEvent(ctx, tx, id, "system", "", "status_changed", map[string]any{"to": "closed", "auto": true})
		})
	}
	return int64(len(ids)), nil
}

func (s *service) supportLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
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
		ok, err := clusterkit.AcquireLease(ctx, s.db, autoCloseLease, holder, 25*time.Second)
		if err != nil || !ok {
			continue
		}
		if _, err := s.idleCloseChats(ctx); err != nil && ctx.Err() == nil {
			s.logger.Warn("Live chat idle close failed: %v", err)
		}
		if tick%20 != 1 {
			continue
		}
		if _, err := s.autoCloseTickets(ctx); err != nil && ctx.Err() == nil {
			s.logger.Warn("Support auto-close failed: %v", err)
		}
		if _, err := s.sweepSupportFiles(ctx); err != nil && ctx.Err() == nil {
			s.logger.Warn("Support file sweep failed: %v", err)
		}
		if _, err := s.sweepActivity(ctx); err != nil && ctx.Err() == nil {
			s.logger.Warn("Activity sweep failed: %v", err)
		}
	}
}

func (s *service) rpcSupportInfo(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	st, err := loadSupportSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, key, name::text, description::text FROM base_support_department WHERE is_active ORDER BY sort, key`)
	if err != nil {
		return nil, err
	}
	depts := []map[string]any{}
	for rows.Next() {
		var id, key, name, desc string
		if err := rows.Scan(&id, &key, &name, &desc); err != nil {
			rows.Close()
			return nil, err
		}
		depts = append(depts, map[string]any{"id": id, "key": key, "name": langs.Resolve(localized.Parse(name), lang), "description": langs.Resolve(localized.Parse(desc), lang)})
	}
	rows.Close()
	online, err := agentsOnline(ctx, s.db)
	if err != nil {
		return nil, err
	}
	var unread, open int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE last_staff_message_at IS NOT NULL AND (player_seen_at IS NULL OR player_seen_at < last_staff_message_at)), count(*) FILTER (WHERE status NOT IN ('resolved','closed')) FROM base_support_ticket WHERE user_id=$1`, userID).Scan(&unread, &open); err != nil {
		return nil, err
	}
	var chatID sql.NullString
	_ = s.db.QueryRowContext(ctx, `SELECT id::text FROM base_live_chat WHERE user_id=$1 AND status <> 'closed'`, userID).Scan(&chatID)
	out := map[string]any{"lang": lang, "departments": depts, "maxAttachments": st.MaxAttachments, "maxAttachmentKb": st.MaxAttachmentKB, "fileTypes": []string{"png", "jpg", "webp", "pdf"},
		"attachments": s.files != nil && st.MaxAttachments > 0, "chatEnabled": st.ChatEnabled, "agentsOnline": online > 0, "greeting": langs.Resolve(st.ChatGreeting, lang), "offline": langs.Resolve(st.ChatOffline, lang),
		"unread": unread, "openTickets": open, "chatId": nil}
	if chatID.Valid {
		out["chatId"] = chatID.String
	}
	return out, nil
}

const playerTicketJSON = `jsonb_build_object('id',t.id,'number',t.number,'subject',t.subject,'departmentKey',(SELECT d.key FROM base_support_department d WHERE d.id=t.department_id),
	'departmentName',(SELECT d.name FROM base_support_department d WHERE d.id=t.department_id),'status',t.status,'priority',t.priority,'messageCount',t.message_count,
	'lastMessageAt',t.last_message_at,'lastMessageBy',t.last_message_by,'lastPreview',t.last_preview,'unread',(t.last_staff_message_at IS NOT NULL AND (t.player_seen_at IS NULL OR t.player_seen_at < t.last_staff_message_at)),
	'canRate',(t.status IN ('resolved','closed') AND t.rating IS NULL),'rating',t.rating,'ratingComment',t.rating_comment,'createdAt',t.created_at,'resolvedAt',t.resolved_at,'closedAt',t.closed_at)`

func (s *service) playerTicket(ctx context.Context, userID, id string, langs *localized.Set, lang string) (map[string]any, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("TICKET_NOT_FOUND")
	}
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT `+playerTicketJSON+`::text FROM base_support_ticket t WHERE t.id=$1 AND t.user_id=$2`, id, userID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("TICKET_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	_ = json.Unmarshal([]byte(raw), &out)
	if dn, ok := out["departmentName"].(map[string]any); ok {
		txt := localized.Text{}
		for k, v := range dn {
			if sv, ok := v.(string); ok {
				txt[k] = sv
			}
		}
		out["departmentName"] = langs.Resolve(txt, lang)
	}
	return out, nil
}

func (s *service) rpcTicketList(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT t.id::text FROM base_support_ticket t WHERE t.user_id=$1 ORDER BY (t.status IN ('resolved','closed')), t.last_message_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	items := []map[string]any{}
	for _, id := range ids {
		it, err := s.playerTicket(ctx, userID, id, langs, lang)
		if err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return map[string]any{"lang": lang, "items": items}, nil
}

func (s *service) rpcTicketGet(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID   string `json:"id"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	t, err := s.playerTicket(ctx, userID, in.ID, langs, lang)
	if err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE base_support_ticket SET player_seen_at=now() WHERE id=$1`, in.ID); err != nil {
		return nil, err
	}
	t["unread"] = false
	msgs, err := s.ticketMessages(ctx, in.ID, true)
	if err != nil {
		return nil, err
	}
	t["messages"] = msgs
	t["lang"] = lang
	return t, nil
}

func (s *service) rpcTicketCreate(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Subject       string   `json:"subject"`
		DepartmentKey string   `json:"departmentKey"`
		DepartmentID  string   `json:"departmentId"`
		Body          string   `json:"body"`
		Attachments   []string `json:"attachments"`
		Lang          string   `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	subject := strings.TrimSpace(in.Subject)
	if subject == "" || utf8.RuneCountInString(subject) > maxSubject {
		return nil, badRequest("INVALID_SUBJECT", "subject", "subject is 1 to 150 characters")
	}
	body, err := cleanBody(in.Body, true)
	if err != nil {
		return nil, err
	}
	st, err := loadSupportSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	var open int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM base_support_ticket WHERE user_id=$1 AND status NOT IN ('resolved','closed')`, userID).Scan(&open); err != nil {
		return nil, err
	}
	if open >= st.MaxOpenTickets {
		return nil, conflict("TOO_MANY_OPEN_TICKETS", "close an open ticket first")
	}
	if err := s.allow(ctx, "sup-t:"+userID, st.TicketsPerHour, time.Hour, "TOO_MANY_TICKETS"); err != nil {
		return nil, err
	}
	_, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	var id, number string
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		dept, err := s.departmentByKeyOrID(ctx, tx, in.DepartmentKey, in.DepartmentID, true)
		if err != nil {
			return err
		}
		atts, err := s.claimAttachments(ctx, tx, in.Attachments, userID, false, st.MaxAttachments)
		if err != nil {
			return err
		}
		id, number, err = s.insertTicket(ctx, tx, newTicket{UserID: userID, Subject: subject, DepartmentID: dept, Priority: "normal", Body: body, Attachments: atts, Source: "player", Lang: lang})
		return err
	})
	if err != nil {
		return nil, err
	}
	s.record(ctx, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindTicket, Action: "created", RefID: id, Actor: "player", Data: map[string]any{"number": number, "subject": subject}}))
	return s.rpcTicketGet(ctx, userID, `{"id":"`+id+`","lang":"`+lang+`"}`)
}

func (s *service) rpcTicketReply(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID          string   `json:"id"`
		Body        string   `json:"body"`
		Attachments []string `json:"attachments"`
		Lang        string   `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	st, err := loadSupportSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if err := s.allow(ctx, "sup-m:"+userID, st.MessagesPerHour, time.Hour, "TOO_MANY_MESSAGES"); err != nil {
		return nil, err
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		t, err := loadTicketForUpdate(ctx, tx, in.ID)
		if err != nil {
			return err
		}
		if t.UserID != userID {
			return notFound("TICKET_NOT_FOUND")
		}
		if t.Status == "closed" {
			return conflict("TICKET_CLOSED", "this ticket is closed; open a new one")
		}
		atts, err := s.claimAttachments(ctx, tx, in.Attachments, userID, false, st.MaxAttachments)
		if err != nil {
			return err
		}
		body, err := cleanBody(in.Body, len(atts) == 0)
		if err != nil {
			return err
		}
		raw, _ := json.Marshal(atts)
		if _, err := tx.ExecContext(ctx, `INSERT INTO base_support_message (id, ticket_id, kind, author_type, author, body, attachments) VALUES ($1,$2,'message','player',$3,$4,$5::jsonb)`,
			uuid.Must(uuid.NewV4()).String(), t.ID, userID, body, string(raw)); err != nil {
			return err
		}
		next := "awaiting_staff"
		if t.Status == "open" {
			next = "open"
		}
		if _, err := tx.ExecContext(ctx, `UPDATE base_support_ticket SET status=$2, message_count=message_count+1, last_message_at=now(), last_message_by='player', last_preview=$3, last_player_message_at=now(), player_seen_at=now(),
			resolved_at=CASE WHEN $2='awaiting_staff' AND status='resolved' THEN NULL ELSE resolved_at END, updated_at=now() WHERE id=$1`, t.ID, next, preview(body)); err != nil {
			return err
		}
		if next != t.Status {
			if err := addTicketEvent(ctx, tx, t.ID, "player", userID, "status_changed", map[string]any{"from": t.Status, "to": next}); err != nil {
				return err
			}
		}
		return refreshDue(ctx, tx, t.ID)
	})
	if err != nil {
		return nil, err
	}
	return s.rpcTicketGet(ctx, userID, mustJSON(map[string]string{"id": in.ID, "lang": in.Lang}))
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func (s *service) rpcTicketResolve(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID   string `json:"id"`
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	var number string
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		t, err := loadTicketForUpdate(ctx, tx, in.ID)
		if err != nil {
			return err
		}
		if t.UserID != userID {
			return notFound("TICKET_NOT_FOUND")
		}
		number = t.Number
		if finished(t.Status) {
			return nil
		}
		if _, err := tx.ExecContext(ctx, `UPDATE base_support_ticket SET status='resolved', resolved_at=now(), due_at=NULL, updated_at=now() WHERE id=$1`, t.ID); err != nil {
			return err
		}
		return addTicketEvent(ctx, tx, t.ID, "player", userID, "status_changed", map[string]any{"from": t.Status, "to": "resolved"})
	})
	if err != nil {
		return nil, err
	}
	s.record(ctx, activity.Entry{UserID: userID, Kind: activity.KindTicket, Action: "resolved_by_player", RefID: in.ID, Actor: "player", Data: map[string]any{"number": number}})
	return s.rpcTicketGet(ctx, userID, mustJSON(in))
}

func (s *service) rpcTicketRate(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ID      string `json:"id"`
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
		Lang    string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if in.Rating < 1 || in.Rating > 5 {
		return nil, badRequest("INVALID_RATING", "rating", "rating is 1 to 5")
	}
	comment := strings.TrimSpace(in.Comment)
	if utf8.RuneCountInString(comment) > 500 {
		return nil, badRequest("COMMENT_TOO_LONG", "comment", "comment is at most 500 characters")
	}
	var number string
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		t, err := loadTicketForUpdate(ctx, tx, in.ID)
		if err != nil {
			return err
		}
		if t.UserID != userID {
			return notFound("TICKET_NOT_FOUND")
		}
		number = t.Number
		if !finished(t.Status) {
			return conflict("TICKET_NOT_FINISHED", "rate a ticket after it is solved")
		}
		if t.Rating.Valid {
			return conflict("ALREADY_RATED", "this ticket is already rated")
		}
		if _, err := tx.ExecContext(ctx, `UPDATE base_support_ticket SET rating=$2, rating_comment=$3, rated_at=now(), updated_at=now() WHERE id=$1`, t.ID, in.Rating, comment); err != nil {
			return err
		}
		return addTicketEvent(ctx, tx, t.ID, "player", userID, "rated", map[string]any{"rating": in.Rating})
	})
	if err != nil {
		return nil, err
	}
	s.record(ctx, activity.Entry{UserID: userID, Kind: activity.KindTicket, Action: "rated", RefID: in.ID, Actor: "player", Data: map[string]any{"number": number, "rating": in.Rating}})
	return s.rpcTicketGet(ctx, userID, mustJSON(map[string]string{"id": in.ID, "lang": in.Lang}))
}
