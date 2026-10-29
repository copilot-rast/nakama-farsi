package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const systemUserID = "00000000-0000-0000-0000-000000000000"

const maxTargetUsers = 1000

type target struct {
	Mode    string   `json:"mode"`
	UserIDs []string `json:"userIds,omitempty"`
	GroupID string   `json:"groupId,omitempty"`
}

func (s *service) cleanTarget(ctx context.Context, q queryer, t target) (target, error) {
	out := target{Mode: strings.TrimSpace(t.Mode)}
	switch out.Mode {
	case "all":
	case "users":
		seen := map[string]bool{}
		for _, raw := range t.UserIDs {
			id, err := cleanUUID("target.userIds", raw, true)
			if err != nil {
				return out, err
			}
			if *id == systemUserID {
				return out, badRequest("USERS_NOT_FOUND", "target.userIds", "the system user cannot be targeted")
			}
			if !seen[*id] {
				seen[*id] = true
				out.UserIDs = append(out.UserIDs, *id)
			}
		}
		if len(out.UserIDs) == 0 {
			return out, badRequest("TARGET_USERS_REQUIRED", "target.userIds", "pick at least one user")
		}
		if len(out.UserIDs) > maxTargetUsers {
			return out, badRequest("TOO_MANY_USERS", "target.userIds", "at most 1000 users")
		}
		var found int
		if err := q.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE id = ANY($1::uuid[])`, uuidArray(out.UserIDs)).Scan(&found); err != nil {
			return out, err
		}
		if found != len(out.UserIDs) {
			return out, badRequest("USERS_NOT_FOUND", "target.userIds", strconv.Itoa(len(out.UserIDs)-found)+" of the users do not exist")
		}
	case "group":
		id, err := cleanUUID("target.groupId", t.GroupID, true)
		if err != nil {
			return out, err
		}
		var exists bool
		if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM groups WHERE id=$1 AND disable_time='1970-01-01 00:00:00+00')`, *id).Scan(&exists); err != nil {
			return out, err
		}
		if !exists {
			return out, badRequest("GROUP_NOT_FOUND", "target.groupId", "group does not exist")
		}
		out.GroupID = *id
	default:
		return out, badRequest("INVALID_TARGET", "target.mode", "mode must be all, users or group")
	}
	return out, nil
}

func uuidArray(ids []string) string {
	return "{" + strings.Join(ids, ",") + "}"
}

type targetUser struct {
	ID       string
	Email    string
	Metadata string
}

func (t targetUser) phone() (string, bool) {
	var md map[string]any
	_ = json.Unmarshal([]byte(t.Metadata), &md)
	switch v := md["phone"].(type) {
	case string:
		return normalizePhone(v)
	case float64:
		return normalizePhone("0" + strconv.FormatFloat(v, 'f', 0, 64))
	}
	return "", false
}

func (t targetUser) email() (string, bool) {
	if e, ok := normalizeEmail(t.Email); ok {
		return e, true
	}
	var md map[string]any
	_ = json.Unmarshal([]byte(t.Metadata), &md)
	if v, ok := md["email"].(string); ok {
		return normalizeEmail(v)
	}
	return "", false
}

func (s *service) eachTargetBatch(ctx context.Context, q interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}, t target, fn func([]targetUser) error) error {
	var base string
	args := []any{}
	switch t.Mode {
	case "all":
		base = `FROM users u WHERE u.id <> '` + systemUserID + `'::uuid AND u.disable_time = '1970-01-01 00:00:00+00'`
	case "users":
		base = `FROM users u WHERE u.id = ANY($1::uuid[])`
		args = append(args, uuidArray(t.UserIDs))
	case "group":
		base = `FROM users u JOIN group_edge ge ON ge.destination_id=u.id WHERE ge.source_id=$1 AND ge.state <= 2 AND u.disable_time = '1970-01-01 00:00:00+00'`
		args = append(args, t.GroupID)
	}
	last := systemUserID
	n := len(args)
	for {
		rows, err := q.QueryContext(ctx, `SELECT u.id::text, COALESCE(u.email,''), u.metadata::text `+base+` AND u.id > $`+itoa(n+1)+`::uuid ORDER BY u.id LIMIT 1000`, append(append([]any{}, args...), last)...)
		if err != nil {
			return err
		}
		batch := []targetUser{}
		for rows.Next() {
			var u targetUser
			if err := rows.Scan(&u.ID, &u.Email, &u.Metadata); err != nil {
				rows.Close()
				return err
			}
			batch = append(batch, u)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}
		if err := fn(batch); err != nil {
			return err
		}
		if len(batch) < 1000 {
			return nil
		}
		last = batch[len(batch)-1].ID
	}
}

func addressFor(channel string, u targetUser) (string, bool) {
	switch channel {
	case "sms":
		return u.phone()
	case "email":
		return u.email()
	}
	return "", true
}

type notificationInput struct {
	Channel    string  `json:"channel"`
	Title      i18n    `json:"title"`
	Body       i18n    `json:"body"`
	Action     *action `json:"action"`
	Target     target  `json:"target"`
	ProviderID string  `json:"providerId"`
}

func cleanNotification(langs *localized.Set, in notificationInput) (notificationInput, error) {
	var err error
	titleMax, bodyMax := 200, 2000
	switch in.Channel {
	case "in_app":
	case "sms":
		bodyMax = 600
	case "email":
		bodyMax = 20000
	default:
		return in, badRequest("INVALID_CHANNEL", "channel", "channel must be in_app, sms or email")
	}
	if in.Title, err = cleanI18n(langs, "title", in.Title, in.Channel != "sms", titleMax); err != nil {
		return in, err
	}
	if in.Body, err = cleanI18n(langs, "body", in.Body, true, bodyMax); err != nil {
		return in, err
	}
	if in.Action != nil && (in.Action.Type != "" || in.Action.Value != "") {
		a, err := cleanAction(langs, "action", *in.Action, false)
		if err != nil {
			return in, err
		}
		in.Action = &a
	} else {
		in.Action = nil
	}
	if in.Channel == "in_app" && strings.TrimSpace(in.ProviderID) != "" {
		return in, badRequest("PROVIDER_NOT_ALLOWED", "providerId", "in_app notifications do not use a provider")
	}
	return in, nil
}

const notificationJSON = `jsonb_build_object('id',n.id,'channel',n.channel,'title',n.title,'body',n.body,'action',n.action,'target',n.target,'providerId',n.provider,'providerKind',p.kind,
	'status',n.status,'recipientCount',n.recipient_count,'sentCount',n.sent_count,'failedCount',n.failed_count,'skippedCount',n.skipped_count,
	'pendingCount',GREATEST(n.recipient_count-n.sent_count-n.failed_count-n.skipped_count,0),'createdBy',n.created_by,'createdAt',n.created_at,'updatedAt',n.updated_at)`

const notificationFrom = ` FROM publication_notification n LEFT JOIN publication_provider p ON p.id::text = n.provider`

func (s *service) createNotification(ctx context.Context, in notificationInput, createdBy string) (string, error) {
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		return "", err
	}
	in, err = cleanNotification(langs, in)
	if err != nil {
		return "", err
	}
	id := uuid.Must(uuid.NewV4()).String()
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		t, err := s.cleanTarget(ctx, tx, in.Target)
		if err != nil {
			return err
		}
		var providerID *string
		if in.Channel != "in_app" {
			var p provider
			if strings.TrimSpace(in.ProviderID) != "" {
				if p, err = s.loadProvider(ctx, tx, strings.TrimSpace(in.ProviderID), false); err != nil {
					if app, ok := asApp(err); ok && app.Status == http.StatusNotFound {
						return badRequest("PROVIDER_NOT_FOUND", "providerId", "provider does not exist")
					}
					return err
				}
				if p.Channel != in.Channel {
					return badRequest("PROVIDER_CHANNEL_MISMATCH", "providerId", "provider channel must be "+in.Channel)
				}
				if !p.Active {
					return badRequest("PROVIDER_INACTIVE", "providerId", "provider is not active")
				}
			} else if p, err = s.defaultProvider(ctx, tx, in.Channel); err != nil {
				return err
			}
			if !providerConfigured(p) {
				return badRequest("PROVIDER_NOT_CONFIGURED", "providerId", "provider is missing required fields")
			}
			providerID = &p.ID
		}
		targetJSON, _ := json.Marshal(t)
		var actionJSON *string
		if in.Action != nil {
			b, _ := json.Marshal(in.Action)
			v := string(b)
			actionJSON = &v
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO publication_notification (id, channel, title, body, action, target, provider, status, created_by) VALUES ($1,$2,$3,$4,$5,$6,$7,'queued',$8)`,
			id, in.Channel, i18nJSON(in.Title), i18nJSON(in.Body), actionJSON, string(targetJSON), providerID, createdBy); err != nil {
			return err
		}
		type rec struct {
			U string  `json:"u"`
			S string  `json:"s"`
			A *string `json:"a"`
			E *string `json:"e"`
		}
		noAddress := "no address"
		err = s.eachTargetBatch(ctx, tx, t, func(batch []targetUser) error {
			recs := make([]rec, 0, len(batch))
			for _, u := range batch {
				r := rec{U: u.ID, S: "pending"}
				if in.Channel != "in_app" {
					if addr, ok := addressFor(in.Channel, u); ok {
						a := addr
						r.A = &a
					} else {
						r.S = "skipped"
						r.E = &noAddress
					}
				}
				recs = append(recs, r)
			}
			raw, _ := json.Marshal(recs)
			_, err := tx.ExecContext(ctx, `INSERT INTO publication_notification_recipient (notification_id, user_id, status, address, error)
				SELECT $1, x.u, x.s, x.a, x.e FROM jsonb_to_recordset($2::jsonb) AS x(u uuid, s text, a text, e text) ON CONFLICT DO NOTHING`, id, string(raw))
			return err
		})
		if err != nil {
			return err
		}
		return refreshCounts(ctx, tx, id)
	})
	return id, err
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func refreshCounts(ctx context.Context, q execer, id string) error {
	_, err := q.ExecContext(ctx, `UPDATE publication_notification n SET recipient_count=c.total, sent_count=c.sent, failed_count=c.failed, skipped_count=c.skipped,
		status = CASE WHEN c.pending > 0 THEN (CASE WHEN c.sent + c.failed > 0 OR n.status='sending' THEN 'sending' ELSE 'queued' END)
			WHEN c.sent = 0 AND c.failed > 0 THEN 'failed' ELSE 'done' END,
		updated_at=now()
		FROM (SELECT count(*) AS total, count(*) FILTER (WHERE status='sent') AS sent, count(*) FILTER (WHERE status='failed') AS failed,
			count(*) FILTER (WHERE status='skipped') AS skipped, count(*) FILTER (WHERE status='pending') AS pending
			FROM publication_notification_recipient WHERE notification_id=$1) c
		WHERE n.id=$1`, id)
	return err
}

func (s *service) notificationDetail(ctx context.Context, id string) (map[string]any, error) {
	if _, err := uuid.FromString(id); err != nil {
		return nil, notFound("NOTIFICATION_NOT_FOUND")
	}
	var raw string
	err := s.db.QueryRowContext(ctx, `SELECT `+notificationJSON+`::text`+notificationFrom+` WHERE n.id=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, notFound("NOTIFICATION_NOT_FOUND")
	}
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	err = json.Unmarshal([]byte(raw), &out)
	return out, err
}

func (s *service) consoleNotificationCreate(w http.ResponseWriter, r *http.Request) {
	var in notificationInput
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	id, err := s.createNotification(r.Context(), in, consoleUser(r))
	if err != nil {
		s.writeError(w, err)
		return
	}
	out, err := s.notificationDetail(r.Context(), id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func consoleUser(r *http.Request) string {
	name := strings.TrimSpace(r.Header.Get("X-Nakama-Console-Username"))
	if utf8.RuneCountInString(name) > 128 {
		name = string([]rune(name)[:128])
	}
	return name
}

func (s *service) consoleNotificationsList(w http.ResponseWriter, r *http.Request) {
	p := readPage(r)
	q := r.URL.Query()
	where := &sqlWhere{}
	if v := q.Get("channel"); v != "" {
		where.add("n.channel = ?", v)
	}
	if v := q.Get("status"); v != "" {
		where.add("n.status = ?", v)
	}
	if p.Q != "" {
		n := where.next()
		where.args = append(where.args, likePattern(p.Q))
		where.parts = append(where.parts, "(COALESCE("+translatedSearch("n.title")+",'') ILIKE "+n+" OR COALESCE("+translatedSearch("n.body")+",'') ILIKE "+n+" OR n.id::text ILIKE "+n+")")
	}
	var total int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*)`+notificationFrom+where.sql(), where.args...).Scan(&total); err != nil {
		s.writeError(w, err)
		return
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(r.Context(), `SELECT `+notificationJSON+`::text`+notificationFrom+where.sql()+` ORDER BY n.created_at DESC, n.id DESC LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2),
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

func (s *service) consoleNotificationGet(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	out, err := s.notificationDetail(r.Context(), id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("rPage"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("rLimit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	where := &sqlWhere{args: []any{id}, parts: []string{"r.notification_id = $1"}}
	if v := q.Get("rStatus"); v != "" {
		where.add("r.status = ?", v)
	}
	var total int64
	if err := s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM publication_notification_recipient r`+where.sql(), where.args...).Scan(&total); err != nil {
		s.writeError(w, err)
		return
	}
	n := len(where.args)
	rows, err := s.db.QueryContext(r.Context(), `SELECT jsonb_build_object('userId',r.user_id,'username',u.username,'displayName',u.display_name,'status',r.status,'address',r.address,
		'error',r.error,'lang',r.lang,'message',r.message,'attempts',r.attempts,'nextAttemptAt',CASE WHEN r.status='pending' THEN r.next_attempt_at END,'seenAt',r.seen_at,'sentAt',r.sent_at)::text
		FROM publication_notification_recipient r LEFT JOIN users u ON u.id=r.user_id`+where.sql()+` ORDER BY u.username, r.user_id LIMIT $`+itoa(n+1)+` OFFSET $`+itoa(n+2),
		append(append([]any{}, where.args...), limit, (page-1)*limit)...)
	if err != nil {
		s.writeError(w, err)
		return
	}
	items, err := rawRows(rows)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var seen int64
	_ = s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM publication_notification_recipient WHERE notification_id=$1 AND seen_at IS NOT NULL`, id).Scan(&seen)
	out["seenCount"] = seen
	out["recipients"] = map[string]any{"items": items, "total": total, "page": page, "limit": limit}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) consolePreviewTarget(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	t := target{Mode: q.Get("mode"), GroupID: q.Get("groupId")}
	for _, part := range strings.Split(q.Get("userIds"), ",") {
		if part = strings.TrimSpace(part); part != "" {
			t.UserIDs = append(t.UserIDs, part)
		}
	}
	clean, err := s.cleanTarget(r.Context(), s.db, t)
	if err != nil {
		s.writeError(w, err)
		return
	}
	var count, withPhone, withEmail int64
	err = s.eachTargetBatch(r.Context(), s.db, clean, func(batch []targetUser) error {
		for _, u := range batch {
			count++
			if _, ok := u.phone(); ok {
				withPhone++
			}
			if _, ok := u.email(); ok {
				withEmail++
			}
		}
		return nil
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": count, "withPhone": withPhone, "withEmail": withEmail})
}
