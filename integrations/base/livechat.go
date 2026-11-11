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
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/cluster"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const (
	notifyChat         = 1211
	typingWindow       = 5 * time.Second
	playerOnlineWindow = 30 * time.Second
	maxChatBody        = 2000
	maxPollWait        = 20 * time.Second
)

type chatRow struct {
	ID, UserID, Status, Agent, Topic, Lang string
	DepartmentID                           sql.NullString
	LastSeq                                int64
}

func loadChatForUpdate(ctx context.Context, tx *sql.Tx, id string) (chatRow, error) {
	var c chatRow
	if _, err := uuid.FromString(id); err != nil {
		return c, notFound("CHAT_NOT_FOUND")
	}
	err := tx.QueryRowContext(ctx, `SELECT id::text, user_id::text, status, agent, topic, lang, department_id::text, last_seq FROM base_live_chat WHERE id=$1 FOR UPDATE`, id).
		Scan(&c.ID, &c.UserID, &c.Status, &c.Agent, &c.Topic, &c.Lang, &c.DepartmentID, &c.LastSeq)
	if errors.Is(err, sql.ErrNoRows) {
		return c, notFound("CHAT_NOT_FOUND")
	}
	return c, err
}

func addChatMessage(ctx context.Context, tx *sql.Tx, chatID, sender, author, body, event string, data map[string]any) (int64, error) {
	var seq int64
	prev := ""
	if sender != "system" {
		prev = preview(body)
	}
	if err := tx.QueryRowContext(ctx, `UPDATE base_live_chat SET last_seq=last_seq+1, last_message_at=CASE WHEN $2 <> 'system' THEN now() ELSE last_message_at END,
		last_preview=CASE WHEN $2 <> 'system' THEN $3 ELSE last_preview END, updated_at=now() WHERE id=$1 RETURNING last_seq`, chatID, sender, prev).Scan(&seq); err != nil {
		return 0, err
	}
	raw := []byte("{}")
	if data != nil {
		raw, _ = json.Marshal(data)
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO base_live_chat_message (chat_id, seq, sender, author, body, event, event_data) VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb)`, chatID, seq, sender, author, body, event, string(raw))
	return seq, err
}

func cleanChatBody(body string) (string, error) {
	body = strings.TrimSpace(strings.ReplaceAll(body, "\r\n", "\n"))
	if body == "" {
		return "", badRequest("BODY_REQUIRED", "body", "write a message")
	}
	if utf8.RuneCountInString(body) > maxChatBody {
		return "", badRequest("BODY_TOO_LONG", "body", "message is at most 2000 characters")
	}
	return body, nil
}

func chatTopic(id string) string {
	return "chat:" + id
}

func chatPollWait() time.Duration {
	if cluster.SignalsEnabled() {
		return 5 * time.Second
	}
	return 600 * time.Millisecond
}

func (s *service) pushChat(ctx context.Context, userID, chatID, event string, seq int64) {
	cluster.Signal(chatTopic(chatID))
	_ = s.nk.NotificationSend(ctx, userID, "support", map[string]any{"type": "chat", "chatId": chatID, "event": event, "seq": seq}, notifyChat, "", false)
}

func (s *service) chatMessages(ctx context.Context, chatID string, after int64, forPlayer bool) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT seq, sender, author, body, event, event_data::text, created_at FROM base_live_chat_message WHERE chat_id=$1 AND seq > $2 ORDER BY seq LIMIT 500`, chatID, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var seq int64
		var sender, author, body, event, data string
		var at time.Time
		if err := rows.Scan(&seq, &sender, &author, &body, &event, &data, &at); err != nil {
			return nil, err
		}
		m := map[string]any{"seq": seq, "sender": sender, "body": body, "createdAt": at.UTC().Format(time.RFC3339)}
		if sender == "agent" || !forPlayer {
			m["author"] = author
		}
		if event != "" {
			m["event"] = event
			m["eventData"] = json.RawMessage(data)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

const chatJSON = `jsonb_build_object('id',c.id,'userId',c.user_id,'username',u.username,'displayName',u.display_name,'status',c.status,'topic',c.topic,'agent',c.agent,'lang',c.lang,
	'departmentId',c.department_id,'department',(SELECT jsonb_build_object('id',d.id,'key',d.key,'name',d.name) FROM base_support_department d WHERE d.id=c.department_id),
	'lastSeq',c.last_seq,'unread',GREATEST(c.last_seq - c.agent_read_seq, 0),'lastMessageAt',c.last_message_at,'lastPreview',c.last_preview,
	'playerTyping',COALESCE(c.player_typing_at > now() - INTERVAL '5 seconds', false),'agentTyping',COALESCE(c.agent_typing_at > now() - INTERVAL '5 seconds', false),
	'playerOnline',(c.player_seen_at > now() - INTERVAL '30 seconds' AND c.status <> 'closed'),'playerSeenAt',c.player_seen_at,'playerReadSeq',c.player_read_seq,'agentReadSeq',c.agent_read_seq,
	'waitingSince',c.waiting_since,'waitingSeconds',CASE WHEN c.status='waiting' THEN EXTRACT(EPOCH FROM now() - c.waiting_since)::int ELSE 0 END,
	'takenAt',c.taken_at,'closedAt',c.closed_at,'closedBy',c.closed_by,'closeReason',c.close_reason,'ticketId',c.ticket_id,
	'ticketNumber',(SELECT t.number FROM base_support_ticket t WHERE t.id=c.ticket_id),'rating',c.rating,'ratingComment',c.rating_comment,'createdAt',c.created_at)`

const chatFrom = ` FROM base_live_chat c JOIN users u ON u.id=c.user_id`

func (s *service) supportChatList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	q := r.URL.Query()
	where := &sqlWhere{}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(u.username ILIKE "+n+" OR c.topic ILIKE "+n+" OR c.user_id::text ILIKE "+n+" OR c.agent ILIKE "+n+")")
	}
	order := "c.last_message_at DESC"
	switch q.Get("view") {
	case "queue":
		where.add("c.status = 'waiting'")
		order = "c.waiting_since ASC"
	case "mine":
		where.add("c.status = 'active' AND c.agent = ?", agentName(r))
	case "active":
		where.add("c.status = 'active'")
	case "closed":
		where.add("c.status = 'closed'")
	default:
		where.add("c.status <> 'closed'")
		order = "(c.status = 'waiting') DESC, c.last_message_at DESC"
	}
	if v := q.Get("userId"); v != "" {
		if _, err := uuid.FromString(v); err == nil {
			where.add("c.user_id = ?::uuid", v)
		}
	}
	out, err := s.listRaw(r, chatFrom, chatJSON, order, where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) chatState(ctx context.Context, id string) (map[string]any, error) {
	raw, err := fetchRow(ctx, s.db, `SELECT `+chatJSON+`::text`+chatFrom+` WHERE c.id=$1`, id)
	if err != nil {
		if app, ok := asApp(err); ok && app.Code == "NOT_FOUND" {
			return nil, notFound("CHAT_NOT_FOUND")
		}
		return nil, err
	}
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out, nil
}

func chatStamp(c map[string]any) string {
	b, _ := json.Marshal([]any{c["lastSeq"], c["status"], c["agent"], c["playerTyping"], c["playerOnline"], c["playerReadSeq"]})
	return string(b)
}

func (s *service) supportChatGet(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	q := r.URL.Query()
	after, _ := strconv.ParseInt(q.Get("after"), 10, 64)
	waitSec, _ := strconv.Atoi(q.Get("wait"))
	wait := time.Duration(waitSec) * time.Second
	if wait > maxPollWait {
		wait = maxPollWait
	}
	stamp := q.Get("stamp")
	_ = s.touchAgent(r.Context(), agentName(r), nil)
	deadline := time.Now().Add(wait)
	for {
		c, err := s.chatState(r.Context(), id)
		if err != nil {
			s.writeError(w, err)
			return
		}
		cur := chatStamp(c)
		if wait <= 0 || cur != stamp || time.Now().After(deadline) {
			msgs, err := s.chatMessages(r.Context(), id, after, false)
			if err != nil {
				s.writeError(w, err)
				return
			}
			c["messages"] = msgs
			c["stamp"] = cur
			if after == 0 {
				var tickets, chats int
				_ = s.db.QueryRowContext(r.Context(), `SELECT (SELECT count(*) FROM base_support_ticket WHERE user_id=$1), (SELECT count(*) FROM base_live_chat WHERE user_id=$1)`, c["userId"]).Scan(&tickets, &chats)
				c["player"] = map[string]any{"tickets": tickets, "chats": chats}
			}
			writeJSON(w, http.StatusOK, c)
			return
		}
		cluster.WaitSignal(r.Context(), chatTopic(id), chatPollWait())
		if r.Context().Err() != nil {
			return
		}
	}
}

func (s *service) supportChatTake(w http.ResponseWriter, r *http.Request) {
	me := agentName(r)
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("CHAT_NOT_FOUND"))
		return
	}
	var userID string
	var seq int64
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		err := tx.QueryRowContext(r.Context(), `UPDATE base_live_chat SET status='active', agent=$2, taken_at=now(), updated_at=now() WHERE id=$1 AND status='waiting' RETURNING user_id::text`, id, me).Scan(&userID)
		if errors.Is(err, sql.ErrNoRows) {
			var status, agent string
			if err := tx.QueryRowContext(r.Context(), `SELECT status, agent FROM base_live_chat WHERE id=$1`, id).Scan(&status, &agent); errors.Is(err, sql.ErrNoRows) {
				return notFound("CHAT_NOT_FOUND")
			} else if err != nil {
				return err
			}
			if status == "closed" {
				return conflict("CHAT_CLOSED", "this chat is closed")
			}
			return fail(http.StatusConflict, "ALREADY_TAKEN", "agent", agent)
		}
		if err != nil {
			return err
		}
		seq, err = addChatMessage(r.Context(), tx, id, "system", me, "", "accepted", map[string]any{"agent": me})
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	_ = s.touchAgent(r.Context(), me, nil)
	s.pushChat(r.Context(), userID, id, "status", seq)
	s.record(r.Context(), activity.Entry{UserID: userID, Kind: activity.KindChat, Action: "taken", RefID: id, Actor: "console:" + me, Data: map[string]any{"agent": me}})
	s.writeChat(w, r, id)
}

func (s *service) writeChat(w http.ResponseWriter, r *http.Request, id string) {
	c, err := s.chatState(r.Context(), id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	msgs, err := s.chatMessages(r.Context(), id, 0, false)
	if err != nil {
		s.writeError(w, err)
		return
	}
	c["messages"] = msgs
	c["stamp"] = chatStamp(c)
	writeJSON(w, http.StatusOK, c)
}

func ownChat(c chatRow, me string) error {
	if c.Status == "closed" {
		return conflict("CHAT_CLOSED", "this chat is closed")
	}
	if c.Status != "active" || c.Agent != me {
		return fail(http.StatusConflict, "NOT_YOUR_CHAT", "agent", c.Agent)
	}
	return nil
}

func (s *service) supportChatMessage(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Body string `json:"body"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	me := agentName(r)
	body, err := cleanChatBody(in.Body)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var c chatRow
	var seq int64
	err = s.inTx(r.Context(), func(tx *sql.Tx) error {
		var err error
		if c, err = loadChatForUpdate(r.Context(), tx, pathVar(r, "id")); err != nil {
			return err
		}
		if err := ownChat(c, me); err != nil {
			return err
		}
		if seq, err = addChatMessage(r.Context(), tx, c.ID, "agent", me, body, "", nil); err != nil {
			return err
		}
		_, err = tx.ExecContext(r.Context(), `UPDATE base_live_chat SET agent_typing_at=NULL, agent_read_seq=$2 WHERE id=$1`, c.ID, seq)
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	_ = s.touchAgent(r.Context(), me, nil)
	s.pushChat(r.Context(), c.UserID, c.ID, "message", seq)
	s.writeChat(w, r, c.ID)
}

func (s *service) supportChatTyping(w http.ResponseWriter, r *http.Request) {
	me := agentName(r)
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("CHAT_NOT_FOUND"))
		return
	}
	var userID string
	err := s.db.QueryRowContext(r.Context(), `UPDATE base_live_chat SET agent_typing_at=now() WHERE id=$1 AND status='active' AND agent=$2 RETURNING user_id::text`, id, me).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		s.writeError(w, conflict("NOT_YOUR_CHAT", "take the chat first"))
		return
	}
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.pushChat(r.Context(), userID, id, "typing", 0)
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) supportChatRead(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Seq int64 `json:"seq"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("CHAT_NOT_FOUND"))
		return
	}
	if _, err := s.db.ExecContext(r.Context(), `UPDATE base_live_chat SET agent_read_seq=GREATEST(agent_read_seq, LEAST($2, last_seq)) WHERE id=$1 AND agent=$3`, id, in.Seq, agentName(r)); err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) supportChatTransfer(w http.ResponseWriter, r *http.Request) {
	var in struct {
		To string `json:"to"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	me := agentName(r)
	to := strings.TrimSpace(in.To)
	if to == me {
		s.writeError(w, badRequest("INVALID_AGENT", "to", "pick another agent"))
		return
	}
	if to != "" {
		var known bool
		if err := s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM console_user WHERE username=$1) OR EXISTS (SELECT 1 FROM base_support_agent WHERE username=$1)`, to).Scan(&known); err != nil {
			s.writeError(w, err)
			return
		}
		if !known {
			s.writeError(w, badRequest("INVALID_AGENT", "to", "unknown agent"))
			return
		}
	}
	var c chatRow
	var seq int64
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		var err error
		if c, err = loadChatForUpdate(r.Context(), tx, pathVar(r, "id")); err != nil {
			return err
		}
		if err := ownChat(c, me); err != nil {
			return err
		}
		if to == "" {
			_, err = tx.ExecContext(r.Context(), `UPDATE base_live_chat SET status='waiting', agent='', waiting_since=now(), agent_typing_at=NULL, updated_at=now() WHERE id=$1`, c.ID)
		} else {
			_, err = tx.ExecContext(r.Context(), `UPDATE base_live_chat SET agent=$2, agent_typing_at=NULL, agent_read_seq=0, updated_at=now() WHERE id=$1`, c.ID, to)
		}
		if err != nil {
			return err
		}
		seq, err = addChatMessage(r.Context(), tx, c.ID, "system", me, "", "transferred", map[string]any{"from": me, "to": to})
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.pushChat(r.Context(), c.UserID, c.ID, "status", seq)
	s.record(r.Context(), activity.Entry{UserID: c.UserID, Kind: activity.KindChat, Action: "transferred", RefID: c.ID, Actor: "console:" + me, Data: map[string]any{"from": me, "to": to}})
	s.writeChat(w, r, c.ID)
}

var transcriptPlayer = localized.Text{"fa": "بازیکن", "en": "Player", "ar": "اللاعب"}

func (s *service) transcript(ctx context.Context, tx *sql.Tx, chatID, playerLabel string) (string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT sender, author, body FROM base_live_chat_message WHERE chat_id=$1 AND sender <> 'system' ORDER BY seq`, chatID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	lines := []string{}
	for rows.Next() {
		var sender, author, body string
		if err := rows.Scan(&sender, &author, &body); err != nil {
			return "", err
		}
		who := playerLabel
		if sender == "agent" {
			who = author
		}
		lines = append(lines, who+": "+body)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	text := strings.Join(lines, "\n")
	if utf8.RuneCountInString(text) > maxTicketBody {
		r := []rune(text)
		text = string(r[len(r)-maxTicketBody+1:])
	}
	return text, nil
}

var chatTicketSubject = localized.Text{"fa": "گفتگوی زنده", "en": "Live chat", "ar": "محادثة مباشرة"}

func (s *service) closeChat(ctx context.Context, id, by, reason string, toTicket bool, staff bool) (chatRow, string, int64, error) {
	var c chatRow
	var ticketID string
	var seq int64
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var err error
		if c, err = loadChatForUpdate(ctx, tx, id); err != nil {
			return err
		}
		if c.Status == "closed" {
			return conflict("CHAT_CLOSED", "this chat is already closed")
		}
		if staff && c.Status == "active" && c.Agent != by {
			return fail(http.StatusConflict, "NOT_YOUR_CHAT", "agent", c.Agent)
		}
		if toTicket {
			langs, err := s.languages(ctx, tx)
			if err != nil {
				return err
			}
			body, err := s.transcript(ctx, tx, c.ID, langs.Resolve(transcriptPlayer, c.Lang))
			if err != nil {
				return err
			}
			if body == "" {
				return badRequest("CHAT_EMPTY", "", "there is nothing to turn into a ticket")
			}
			subject := strings.TrimSpace(c.Topic)
			if subject == "" {
				subject = langs.Resolve(chatTicketSubject, c.Lang)
			}
			var dept *string
			if c.DepartmentID.Valid {
				d := c.DepartmentID.String
				dept = &d
			} else if dept, err = s.departmentByKeyOrID(ctx, tx, "", "", false); err != nil {
				return err
			}
			cid := c.ID
			nt := newTicket{UserID: c.UserID, Subject: subject, DepartmentID: dept, Priority: "normal", Body: body, Source: "chat", ChatID: &cid, CreatedBy: by, Lang: c.Lang}
			if staff && c.Agent == by {
				nt.StaffAuthor = ""
			}
			if ticketID, _, err = s.insertTicket(ctx, tx, nt); err != nil {
				return err
			}
			if staff {
				if _, err := tx.ExecContext(ctx, `UPDATE base_support_ticket SET assignee=$2 WHERE id=$1 AND assignee=''`, ticketID, by); err != nil {
					return err
				}
			}
			if err := addTicketEvent(ctx, tx, ticketID, "system", by, "created_from_chat", map[string]any{"chatId": c.ID}); err != nil {
				return err
			}
		}
		closer := by
		if !staff {
			closer = "player"
		}
		if _, err := tx.ExecContext(ctx, `UPDATE base_live_chat SET status='closed', closed_at=now(), closed_by=$2, close_reason=$3, ticket_id=NULLIF($4,'')::uuid, agent_typing_at=NULL, player_typing_at=NULL, updated_at=now() WHERE id=$1`,
			c.ID, closer, reason, ticketID); err != nil {
			return err
		}
		seq, err = addChatMessage(ctx, tx, c.ID, "system", closer, "", "closed", map[string]any{"by": closer, "reason": reason, "ticketId": ticketID})
		return err
	})
	return c, ticketID, seq, err
}

func (s *service) supportChatClose(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ToTicket bool `json:"toTicket"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	me := agentName(r)
	c, ticketID, seq, err := s.closeChat(r.Context(), pathVar(r, "id"), me, "agent", in.ToTicket, true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.pushChat(r.Context(), c.UserID, c.ID, "status", seq)
	data := map[string]any{"by": me}
	if ticketID != "" {
		data["ticketId"] = ticketID
	}
	s.record(r.Context(), activity.Entry{UserID: c.UserID, Kind: activity.KindChat, Action: "closed", RefID: c.ID, Actor: "console:" + me, Data: data})
	s.writeChat(w, r, c.ID)
}

func (s *service) idleCloseChats(ctx context.Context) (int, error) {
	st, err := loadSupportSettings(ctx, s.db)
	if err != nil {
		return 0, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id::text FROM base_live_chat WHERE status <> 'closed' AND last_message_at < now() - make_interval(mins => $1) AND player_seen_at < now() - make_interval(mins => $1) LIMIT 100`, st.ChatIdleMinutes)
	if err != nil {
		return 0, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	n := 0
	for _, id := range ids {
		c, _, seq, err := s.closeChat(ctx, id, "system", "idle", false, false)
		if err != nil {
			continue
		}
		n++
		s.pushChat(ctx, c.UserID, c.ID, "status", seq)
	}
	return n, nil
}

func (s *service) playerChat(ctx context.Context, userID, chatID string, after int64, touch bool) (map[string]any, error) {
	if chatID == "" {
		var id string
		err := s.db.QueryRowContext(ctx, `SELECT id::text FROM base_live_chat WHERE user_id=$1 ORDER BY (status <> 'closed') DESC, created_at DESC LIMIT 1`, userID).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			online, err := agentsOnline(ctx, s.db)
			if err != nil {
				return nil, err
			}
			return map[string]any{"chat": nil, "agentsOnline": online > 0}, nil
		}
		if err != nil {
			return nil, err
		}
		chatID = id
	}
	if _, err := uuid.FromString(chatID); err != nil {
		return nil, notFound("CHAT_NOT_FOUND")
	}
	var owner string
	if err := s.db.QueryRowContext(ctx, `SELECT user_id::text FROM base_live_chat WHERE id=$1`, chatID).Scan(&owner); err != nil || owner != userID {
		return nil, notFound("CHAT_NOT_FOUND")
	}
	if touch {
		if _, err := s.db.ExecContext(ctx, `UPDATE base_live_chat SET player_seen_at=now(), player_read_seq=last_seq WHERE id=$1 AND status <> 'closed'`, chatID); err != nil {
			return nil, err
		}
		cluster.Signal(chatTopic(chatID))
	}
	c, err := s.chatState(ctx, chatID)
	if err != nil {
		return nil, err
	}
	msgs, err := s.chatMessages(ctx, chatID, after, true)
	if err != nil {
		return nil, err
	}
	online, err := agentsOnline(ctx, s.db)
	if err != nil {
		return nil, err
	}
	view := map[string]any{"id": c["id"], "status": c["status"], "topic": c["topic"], "agent": c["agent"], "agentTyping": c["agentTyping"], "lastSeq": c["lastSeq"], "agentReadSeq": c["agentReadSeq"],
		"ticketId": c["ticketId"], "ticketNumber": c["ticketNumber"], "rating": c["rating"], "canRate": c["status"] == "closed" && c["rating"] == nil && c["takenAt"] != nil, "createdAt": c["createdAt"], "closeReason": c["closeReason"], "messages": msgs}
	return map[string]any{"chat": view, "agentsOnline": online > 0}, nil
}

func (s *service) rpcChatOpen(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Topic         string `json:"topic"`
		Body          string `json:"body"`
		DepartmentKey string `json:"departmentKey"`
		Lang          string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	st, err := loadSupportSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if !st.ChatEnabled {
		return nil, fail(http.StatusServiceUnavailable, "CHAT_DISABLED", "", "live chat is turned off")
	}
	topic := strings.TrimSpace(in.Topic)
	if utf8.RuneCountInString(topic) > 80 {
		return nil, badRequest("INVALID_TOPIC", "topic", "topic is at most 80 characters")
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	var existing string
	err = s.db.QueryRowContext(ctx, `SELECT id::text FROM base_live_chat WHERE user_id=$1 AND status <> 'closed'`, userID).Scan(&existing)
	if err == nil {
		if strings.TrimSpace(in.Body) != "" {
			return s.rpcChatSend(ctx, userID, mustJSON(map[string]string{"chatId": existing, "body": in.Body}))
		}
		return s.playerChat(ctx, userID, existing, 0, true)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err := s.allow(ctx, "sup-c:"+userID, 10, time.Hour, "TOO_MANY_CHATS"); err != nil {
		return nil, err
	}
	id := uuid.Must(uuid.NewV4()).String()
	var greet string
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		dept, err := s.departmentByKeyOrID(ctx, tx, in.DepartmentKey, "", true)
		if err != nil && in.DepartmentKey != "" {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO base_live_chat (id, user_id, topic, department_id, lang) VALUES ($1,$2,$3,$4,$5)`, id, userID, topic, dept, lang); err != nil {
			return err
		}
		greet = langs.Resolve(st.ChatGreeting, lang)
		if greet != "" {
			if _, err := addChatMessage(ctx, tx, id, "system", "", greet, "greeting", nil); err != nil {
				return err
			}
		}
		if body := strings.TrimSpace(in.Body); body != "" {
			clean, err := cleanChatBody(body)
			if err != nil {
				return err
			}
			if _, err := addChatMessage(ctx, tx, id, "player", userID, clean, "", nil); err != nil {
				return err
			}
		}
		return nil
	})
	if _, unique := isUniqueViolation(err); unique {
		return s.playerChat(ctx, userID, "", 0, true)
	}
	if err != nil {
		return nil, err
	}
	s.record(ctx, activity.FromContext(ctx, activity.Entry{UserID: userID, Kind: activity.KindChat, Action: "opened", RefID: id, Actor: "player", Data: map[string]any{"topic": topic}}))
	return s.playerChat(ctx, userID, id, 0, true)
}

func (s *service) rpcChatGet(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ChatID string `json:"chatId"`
		After  int64  `json:"after"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	return s.playerChat(ctx, userID, strings.TrimSpace(in.ChatID), in.After, true)
}

func (s *service) rpcChatSend(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ChatID string `json:"chatId"`
		Body   string `json:"body"`
		After  int64  `json:"after"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	body, err := cleanChatBody(in.Body)
	if err != nil {
		return nil, err
	}
	st, err := loadSupportSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if err := s.allow(ctx, "sup-cm:"+userID, st.ChatMessagesPerMinute, time.Minute, "TOO_MANY_MESSAGES"); err != nil {
		return nil, err
	}
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		c, err := loadChatForUpdate(ctx, tx, in.ChatID)
		if err != nil {
			return err
		}
		if c.UserID != userID {
			return notFound("CHAT_NOT_FOUND")
		}
		if c.Status == "closed" {
			return conflict("CHAT_CLOSED", "this chat is closed")
		}
		seq, err := addChatMessage(ctx, tx, c.ID, "player", userID, body, "", nil)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE base_live_chat SET player_typing_at=NULL, player_read_seq=$2, player_seen_at=now() WHERE id=$1`, c.ID, seq)
		return err
	})
	if err != nil {
		return nil, err
	}
	return s.playerChat(ctx, userID, in.ChatID, in.After, true)
}

func (s *service) rpcChatTyping(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ChatID string `json:"chatId"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if _, err := uuid.FromString(in.ChatID); err != nil {
		return nil, notFound("CHAT_NOT_FOUND")
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE base_live_chat SET player_typing_at=now(), player_seen_at=now() WHERE id=$1 AND user_id=$2 AND status <> 'closed'`, in.ChatID, userID); err != nil {
		return nil, err
	}
	cluster.Signal(chatTopic(in.ChatID))
	return map[string]any{}, nil
}

func (s *service) rpcChatClose(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ChatID   string `json:"chatId"`
		ToTicket bool   `json:"toTicket"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	if _, err := uuid.FromString(in.ChatID); err != nil {
		return nil, notFound("CHAT_NOT_FOUND")
	}
	var owner string
	if err := s.db.QueryRowContext(ctx, `SELECT user_id::text FROM base_live_chat WHERE id=$1`, in.ChatID).Scan(&owner); err != nil || owner != userID {
		return nil, notFound("CHAT_NOT_FOUND")
	}
	_, ticketID, _, err := s.closeChat(ctx, in.ChatID, userID, "player", in.ToTicket, false)
	if err != nil {
		return nil, err
	}
	data := map[string]any{"by": "player"}
	if ticketID != "" {
		data["ticketId"] = ticketID
	}
	s.record(ctx, activity.Entry{UserID: userID, Kind: activity.KindChat, Action: "closed", RefID: in.ChatID, Actor: "player", Data: data})
	cluster.Signal(chatTopic(in.ChatID))
	return s.playerChat(ctx, userID, in.ChatID, 0, false)
}

func (s *service) rpcChatRate(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		ChatID  string `json:"chatId"`
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
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
	if _, err := uuid.FromString(in.ChatID); err != nil {
		return nil, notFound("CHAT_NOT_FOUND")
	}
	res, err := s.db.ExecContext(ctx, `UPDATE base_live_chat SET rating=$3, rating_comment=$4, updated_at=now() WHERE id=$1 AND user_id=$2 AND status='closed' AND rating IS NULL AND taken_at IS NOT NULL`, in.ChatID, userID, in.Rating, comment)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, conflict("CANNOT_RATE", "this chat cannot be rated")
	}
	return s.playerChat(ctx, userID, in.ChatID, 0, false)
}
