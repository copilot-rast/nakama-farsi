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
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const agentOnlineWindow = 75 * time.Second

type supportSettings struct {
	MaxOpenTickets        int            `json:"maxOpenTickets"`
	TicketsPerHour        int            `json:"ticketsPerHour"`
	MessagesPerHour       int            `json:"messagesPerHour"`
	MaxAttachments        int            `json:"maxAttachments"`
	MaxAttachmentKB       int            `json:"maxAttachmentKb"`
	AutoCloseDays         int            `json:"autoCloseDays"`
	ChatEnabled           bool           `json:"chatEnabled"`
	ChatGreeting          localized.Text `json:"chatGreeting"`
	ChatOffline           localized.Text `json:"chatOffline"`
	ChatIdleMinutes       int            `json:"chatIdleMinutes"`
	ChatMessagesPerMinute int            `json:"chatMessagesPerMinute"`
	ActivityRetentionDays int            `json:"activityRetentionDays"`
	UpdatedAt             string         `json:"updatedAt,omitempty"`
}

func loadSupportSettings(ctx context.Context, q queryer) (supportSettings, error) {
	st := supportSettings{MaxOpenTickets: 10, TicketsPerHour: 6, MessagesPerHour: 40, MaxAttachments: 5, MaxAttachmentKB: 2048, AutoCloseDays: 7, ChatEnabled: true, ChatIdleMinutes: 30, ChatMessagesPerMinute: 12, ActivityRetentionDays: 180, ChatGreeting: localized.Text{}, ChatOffline: localized.Text{}}
	var greeting, offline string
	var updated time.Time
	err := q.QueryRowContext(ctx, `SELECT max_open_tickets, tickets_per_hour, messages_per_hour, max_attachments, max_attachment_kb, auto_close_days, chat_enabled, chat_greeting::text, chat_offline::text, chat_idle_minutes, chat_messages_per_minute, activity_retention_days, updated_at FROM base_support_settings WHERE id=1`).
		Scan(&st.MaxOpenTickets, &st.TicketsPerHour, &st.MessagesPerHour, &st.MaxAttachments, &st.MaxAttachmentKB, &st.AutoCloseDays, &st.ChatEnabled, &greeting, &offline, &st.ChatIdleMinutes, &st.ChatMessagesPerMinute, &st.ActivityRetentionDays, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	st.ChatGreeting = localized.Parse(greeting)
	st.ChatOffline = localized.Parse(offline)
	st.UpdatedAt = updated.UTC().Format(time.RFC3339)
	return st, nil
}

func (s *service) supportSettingsGet(w http.ResponseWriter, r *http.Request) {
	st, err := loadSupportSettings(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func between(field string, v, lo, hi int) error {
	if v < lo || v > hi {
		return badRequest("INVALID_VALUE", field, field+" is out of range")
	}
	return nil
}

func (s *service) supportSettingsPut(w http.ResponseWriter, r *http.Request) {
	var in supportSettings
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	langs, err := s.languages(r.Context(), s.db)
	if err != nil {
		s.writeError(w, err)
		return
	}
	checks := []error{
		between("maxOpenTickets", in.MaxOpenTickets, 1, 100),
		between("ticketsPerHour", in.TicketsPerHour, 1, 1000),
		between("messagesPerHour", in.MessagesPerHour, 1, 10000),
		between("maxAttachments", in.MaxAttachments, 0, 10),
		between("maxAttachmentKb", in.MaxAttachmentKB, 16, 10240),
		between("autoCloseDays", in.AutoCloseDays, 1, 365),
		between("chatIdleMinutes", in.ChatIdleMinutes, 5, 1440),
		between("chatMessagesPerMinute", in.ChatMessagesPerMinute, 1, 600),
		between("activityRetentionDays", in.ActivityRetentionDays, 7, 3650),
	}
	for _, c := range checks {
		if c != nil {
			s.writeError(w, c)
			return
		}
	}
	greeting, err := cleanText(langs, "chatGreeting", in.ChatGreeting, false, 500)
	if err != nil {
		s.writeError(w, err)
		return
	}
	offline, err := cleanText(langs, "chatOffline", in.ChatOffline, false, 500)
	if err != nil {
		s.writeError(w, err)
		return
	}
	_, err = s.db.ExecContext(r.Context(), `INSERT INTO base_support_settings (id, max_open_tickets, tickets_per_hour, messages_per_hour, max_attachments, max_attachment_kb, auto_close_days, chat_enabled, chat_greeting, chat_offline, chat_idle_minutes, chat_messages_per_minute, activity_retention_days, updated_at)
		VALUES (1,$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,now())
		ON CONFLICT (id) DO UPDATE SET max_open_tickets=EXCLUDED.max_open_tickets, tickets_per_hour=EXCLUDED.tickets_per_hour, messages_per_hour=EXCLUDED.messages_per_hour, max_attachments=EXCLUDED.max_attachments,
		max_attachment_kb=EXCLUDED.max_attachment_kb, auto_close_days=EXCLUDED.auto_close_days, chat_enabled=EXCLUDED.chat_enabled, chat_greeting=EXCLUDED.chat_greeting, chat_offline=EXCLUDED.chat_offline,
		chat_idle_minutes=EXCLUDED.chat_idle_minutes, chat_messages_per_minute=EXCLUDED.chat_messages_per_minute, activity_retention_days=EXCLUDED.activity_retention_days, updated_at=now()`,
		in.MaxOpenTickets, in.TicketsPerHour, in.MessagesPerHour, in.MaxAttachments, in.MaxAttachmentKB, in.AutoCloseDays, in.ChatEnabled, greeting.JSON(), offline.JSON(), in.ChatIdleMinutes, in.ChatMessagesPerMinute, in.ActivityRetentionDays)
	if err != nil {
		s.writeError(w, err)
		return
	}
	s.supportSettingsGet(w, r)
}

const departmentJSON = `jsonb_build_object('id',d.id,'key',d.key,'name',d.name,'description',d.description,'firstResponseHours',d.first_response_hours,'agents',to_jsonb(d.agents),'isActive',d.is_active,'sort',d.sort,
	'open',(SELECT count(*) FROM base_support_ticket t WHERE t.department_id=d.id AND t.status NOT IN ('resolved','closed')),'createdAt',d.created_at,'updatedAt',d.updated_at)`

const cannedJSON = `jsonb_build_object('id',c.id,'key',c.key,'title',c.title,'body',c.body,'departmentId',c.department_id,'department',(SELECT jsonb_build_object('id',d.id,'key',d.key,'name',d.name) FROM base_support_department d WHERE d.id=c.department_id),'isActive',c.is_active,'sort',c.sort,'createdAt',c.created_at,'updatedAt',c.updated_at)`

func activeFilter(r *http.Request, where *sqlWhere, col string) {
	if v := r.URL.Query().Get("isActive"); v == "true" || v == "false" {
		where.add(col+" = ?", v == "true")
	}
}

func (s *service) departmentList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(d.key ILIKE "+n+" OR "+translatedSearch("d.name", n)+")")
	}
	activeFilter(r, where, "d.is_active")
	out, err := s.listRaw(r, ` FROM base_support_department d`, departmentJSON, "d.sort, d.key", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func fetchRow(ctx context.Context, q queryer, sqlText, id string) (json.RawMessage, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOT_FOUND")
	}
	var raw string
	err := q.QueryRowContext(ctx, sqlText, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func (s *service) departmentGet(w http.ResponseWriter, r *http.Request) {
	raw, err := fetchRow(r.Context(), s.db, `SELECT `+departmentJSON+`::text FROM base_support_department d WHERE d.id=$1`, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, raw)
}

func cleanAgentList(in []string) ([]string, error) {
	seen := map[string]bool{}
	out := []string{}
	for _, a := range in {
		a = strings.TrimSpace(a)
		if a == "" || seen[a] {
			continue
		}
		if len(a) > 128 {
			return nil, badRequest("INVALID_AGENT", "agents", "agent names are at most 128 characters")
		}
		seen[a] = true
		out = append(out, a)
	}
	if len(out) > 50 {
		return nil, badRequest("TOO_MANY_AGENTS", "agents", "at most 50 agents per department")
	}
	return out, nil
}

func (s *service) departmentWrite(create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Key                string         `json:"key"`
			Name               localized.Text `json:"name"`
			Description        localized.Text `json:"description"`
			FirstResponseHours *int           `json:"firstResponseHours"`
			Agents             []string       `json:"agents"`
			IsActive           *bool          `json:"isActive"`
			Sort               *int           `json:"sort"`
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
			hours := 24
			if in.FirstResponseHours != nil {
				hours = *in.FirstResponseHours
			}
			if err := between("firstResponseHours", hours, 1, 720); err != nil {
				return err
			}
			agents, err := cleanAgentList(in.Agents)
			if err != nil {
				return err
			}
			active := in.IsActive == nil || *in.IsActive
			sortN := 100
			if in.Sort != nil {
				sortN = *in.Sort
			}
			if create {
				_, err = tx.ExecContext(r.Context(), `INSERT INTO base_support_department (id, key, name, description, first_response_hours, agents, is_active, sort) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
					id, key, name.JSON(), desc.JSON(), hours, agents, active, sortN)
			} else {
				var res sql.Result
				res, err = tx.ExecContext(r.Context(), `UPDATE base_support_department SET key=$2, name=$3, description=$4, first_response_hours=$5, agents=$6, is_active=$7, sort=$8, updated_at=now() WHERE id=$1`,
					id, key, name.JSON(), desc.JSON(), hours, agents, active, sortN)
				if err == nil {
					if n, _ := res.RowsAffected(); n == 0 {
						return notFound("NOT_FOUND")
					}
				}
			}
			if err != nil {
				return err
			}
			out, err = fetchRow(r.Context(), tx, `SELECT `+departmentJSON+`::text FROM base_support_department d WHERE d.id=$1`, id)
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

func (s *service) departmentDelete(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		var used bool
		if err := tx.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM base_support_ticket WHERE department_id=$1 AND status NOT IN ('resolved','closed'))`, id).Scan(&used); err != nil {
			return err
		}
		if used {
			return conflict("IN_USE", "this department still has open tickets")
		}
		res, err := tx.ExecContext(r.Context(), `DELETE FROM base_support_department WHERE id=$1`, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return notFound("NOT_FOUND")
		}
		return nil
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (s *service) cannedList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	where := &sqlWhere{}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(c.key ILIKE "+n+" OR "+translatedSearch("c.title", n)+" OR "+translatedSearch("c.body", n)+")")
	}
	activeFilter(r, where, "c.is_active")
	if v := r.URL.Query().Get("departmentId"); v != "" {
		if _, err := uuid.FromString(v); err == nil {
			where.add("(c.department_id = ?::uuid OR c.department_id IS NULL)", v)
		}
	}
	out, err := s.listRaw(r, ` FROM base_support_canned c`, cannedJSON, "c.sort, c.key", where, p)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) cannedGet(w http.ResponseWriter, r *http.Request) {
	raw, err := fetchRow(r.Context(), s.db, `SELECT `+cannedJSON+`::text FROM base_support_canned c WHERE c.id=$1`, pathVar(r, "id"))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeRaw(w, http.StatusOK, raw)
}

func optionalUUID(field, v string) (*string, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	if _, err := uuid.FromString(v); err != nil {
		return nil, badRequest("INVALID_ID", field, field+" must be a UUID")
	}
	return &v, nil
}

func (s *service) cannedWrite(create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Key          string         `json:"key"`
			Title        localized.Text `json:"title"`
			Body         localized.Text `json:"body"`
			DepartmentID string         `json:"departmentId"`
			IsActive     *bool          `json:"isActive"`
			Sort         *int           `json:"sort"`
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
			title, err := cleanText(langs, "title", in.Title, true, 100)
			if err != nil {
				return err
			}
			body, err := cleanText(langs, "body", in.Body, true, 3000)
			if err != nil {
				return err
			}
			dept, err := optionalUUID("departmentId", in.DepartmentID)
			if err != nil {
				return err
			}
			if dept != nil {
				var ok bool
				if err := tx.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM base_support_department WHERE id=$1)`, *dept).Scan(&ok); err != nil {
					return err
				}
				if !ok {
					return badRequest("DEPARTMENT_NOT_FOUND", "departmentId", "department does not exist")
				}
			}
			active := in.IsActive == nil || *in.IsActive
			sortN := 100
			if in.Sort != nil {
				sortN = *in.Sort
			}
			if create {
				_, err = tx.ExecContext(r.Context(), `INSERT INTO base_support_canned (id, key, title, body, department_id, is_active, sort) VALUES ($1,$2,$3,$4,$5,$6,$7)`, id, key, title.JSON(), body.JSON(), dept, active, sortN)
			} else {
				var res sql.Result
				res, err = tx.ExecContext(r.Context(), `UPDATE base_support_canned SET key=$2, title=$3, body=$4, department_id=$5, is_active=$6, sort=$7, updated_at=now() WHERE id=$1`, id, key, title.JSON(), body.JSON(), dept, active, sortN)
				if err == nil {
					if n, _ := res.RowsAffected(); n == 0 {
						return notFound("NOT_FOUND")
					}
				}
			}
			if err != nil {
				return err
			}
			out, err = fetchRow(r.Context(), tx, `SELECT `+cannedJSON+`::text FROM base_support_canned c WHERE c.id=$1`, id)
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

func (s *service) cannedDelete(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM base_support_canned WHERE id=$1`, id)
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

func agentName(r *http.Request) string {
	if u := consoleUser(r); u != "" {
		return u
	}
	return "admin"
}

func (s *service) touchAgent(ctx context.Context, name string, away *bool) error {
	if away == nil {
		_, err := s.db.ExecContext(ctx, `INSERT INTO base_support_agent (username, last_seen_at) VALUES ($1, now()) ON CONFLICT (username) DO UPDATE SET last_seen_at=now()`, name)
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO base_support_agent (username, last_seen_at, away) VALUES ($1, now(), $2) ON CONFLICT (username) DO UPDATE SET last_seen_at=now(), away=EXCLUDED.away`, name, *away)
	return err
}

func agentsOnline(ctx context.Context, q queryer) (int, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT count(*) FROM base_support_agent WHERE NOT away AND last_seen_at > now() - $1::INT8 * INTERVAL '1 second'`, int64(agentOnlineWindow.Seconds())).Scan(&n)
	return n, err
}

func (s *service) supportAgents(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `WITH names AS (
			SELECT username::text AS username FROM console_user
			UNION SELECT username FROM base_support_agent
			UNION SELECT $1::text)
		SELECT n.username, COALESCE(a.last_seen_at > now() - $2::INT8 * INTERVAL '1 second' AND NOT a.away, false), COALESCE(a.away, false), a.last_seen_at,
			(SELECT count(*) FROM base_support_ticket t WHERE t.assignee=n.username AND t.status NOT IN ('resolved','closed')),
			(SELECT count(*) FROM base_live_chat c WHERE c.agent=n.username AND c.status='active')
		FROM names n LEFT JOIN base_support_agent a ON a.username=n.username ORDER BY 2 DESC, n.username`, agentName(r), int64(agentOnlineWindow.Seconds()))
	if err != nil {
		s.writeError(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var name string
		var online, away bool
		var seen sql.NullTime
		var tickets, chats int
		if err := rows.Scan(&name, &online, &away, &seen, &tickets, &chats); err != nil {
			s.writeError(w, err)
			return
		}
		it := map[string]any{"username": name, "online": online, "away": away, "openTickets": tickets, "activeChats": chats, "me": name == agentName(r)}
		if seen.Valid {
			it["lastSeenAt"] = seen.Time.UTC().Format(time.RFC3339)
		}
		items = append(items, it)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "me": agentName(r)})
}

func (s *service) supportSummaryData(ctx context.Context, me string) (map[string]any, error) {
	var awaiting, open, overdue, unassigned, mine, waiting, active, myChats int
	var avgFirst, avgRating sql.NullFloat64
	var rated int
	err := s.db.QueryRowContext(ctx, `SELECT
			count(*) FILTER (WHERE status IN ('open','awaiting_staff')),
			count(*) FILTER (WHERE status NOT IN ('resolved','closed')),
			count(*) FILTER (WHERE status IN ('open','awaiting_staff') AND due_at < now()),
			count(*) FILTER (WHERE status NOT IN ('resolved','closed') AND assignee=''),
			count(*) FILTER (WHERE status IN ('open','awaiting_staff') AND assignee=$1),
			avg(EXTRACT(EPOCH FROM first_response_at - created_at) / 60) FILTER (WHERE first_response_at IS NOT NULL AND created_at > now() - INTERVAL '30 days'),
			avg(rating) FILTER (WHERE rating IS NOT NULL),
			count(*) FILTER (WHERE rating IS NOT NULL)
		FROM base_support_ticket`, me).Scan(&awaiting, &open, &overdue, &unassigned, &mine, &avgFirst, &avgRating, &rated)
	if err != nil {
		return nil, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE status='waiting'), count(*) FILTER (WHERE status='active'), count(*) FILTER (WHERE status='active' AND agent=$1) FROM base_live_chat WHERE status <> 'closed'`, me).Scan(&waiting, &active, &myChats); err != nil {
		return nil, err
	}
	online, err := agentsOnline(ctx, s.db)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"awaitingStaff": awaiting, "open": open, "overdue": overdue, "unassigned": unassigned, "mine": mine, "waitingChats": waiting, "activeChats": active, "myChats": myChats, "agentsOnline": online, "rated": rated, "me": me}
	if avgFirst.Valid {
		out["avgFirstResponseMinutes"] = int(avgFirst.Float64 + 0.5)
	}
	if avgRating.Valid {
		out["avgRating"] = float64(int(avgRating.Float64*10+0.5)) / 10
	}
	return out, nil
}

func (s *service) supportSummary(w http.ResponseWriter, r *http.Request) {
	out, err := s.supportSummaryData(r.Context(), agentName(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) supportHeartbeat(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Away *bool `json:"away"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	if err := s.touchAgent(r.Context(), agentName(r), in.Away); err != nil {
		s.writeError(w, err)
		return
	}
	s.supportSummary(w, r)
}
