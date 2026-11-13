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

var campaignActions = []string{"screen", "goto", "product", "achievement", "url"}

type notifyOpts struct {
	kind    string
	ruleID  string
	vars    map[string]map[string]string
	trusted bool
}

func isPushChannel(ch string) bool { return ch == "push" || ch == "push_in_app" }

func cleanCampaignAction(ctx context.Context, q queryer, langs *localized.Set, field string, in action) (action, error) {
	in.Type = strings.TrimSpace(in.Type)
	in.Value = strings.TrimSpace(in.Value)
	switch in.Type {
	case "url", "goto":
		return cleanAction(langs, field, in, false)
	case "screen":
		if !bannerScreenPattern.MatchString(in.Value) {
			return in, badRequest("INVALID_ACTION", field+".value", "screen names use a-z, 0-9, _ and - (max 40)")
		}
	case "product", "achievement":
		table := "economy_product"
		if in.Type == "achievement" {
			table = "economy_achievement"
		}
		var ok bool
		if err := q.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM `+table+` WHERE key=$1)`, in.Value).Scan(&ok); err != nil {
			return in, err
		}
		if !ok {
			return in, badRequest("ACTION_TARGET_NOT_FOUND", field+".value", in.Type+" does not exist")
		}
	default:
		return in, badRequest("INVALID_ACTION", field+".type", "type must be screen, product, achievement or url")
	}
	label, err := cleanI18n(langs, field+".label", in.Label, false, 64)
	if err != nil {
		return in, err
	}
	if len(label) == 0 {
		label = nil
	}
	in.Label = label
	return in, nil
}

func (in notificationInput) isCampaign() bool {
	return isPushChannel(in.Channel) || strings.TrimSpace(in.CategoryID) != "" || strings.TrimSpace(in.ScheduledAt) != "" || in.Audience != nil
}

func (s *service) createNotificationTx(ctx context.Context, tx *sql.Tx, in notificationInput, createdBy string, opts notifyOpts) (string, error) {
	langs, err := localized.Load(ctx, tx)
	if err != nil {
		return "", err
	}
	in, err = cleanNotification(langs, in)
	if err != nil {
		return "", err
	}
	if opts.kind == "" {
		opts.kind = "message"
		if in.isCampaign() {
			opts.kind = "campaign"
		}
	}
	if in.Action != nil && (isPushChannel(in.Channel) || opts.kind != "message") {
		a, err := cleanCampaignAction(ctx, tx, langs, "action", *in.Action)
		if err != nil {
			return "", err
		}
		in.Action = &a
	}
	name := strings.TrimSpace(in.Name)
	if len([]rune(name)) > 120 {
		return "", badRequest("TEXT_TOO_LONG", "name", "name is too long (max 120)")
	}
	image, err := cleanURL("imageUrl", in.ImageURL)
	if err != nil {
		return "", err
	}
	if image != "" && (in.Channel == "sms" || in.Channel == "email") {
		return "", badRequest("IMAGE_NOT_ALLOWED", "imageUrl", "sms and email notifications have no picture")
	}
	cat, err := cleanCategoryID(ctx, tx, in.CategoryID, isPushChannel(in.Channel) && opts.kind != "test")
	if err != nil {
		return "", err
	}
	var aud *audienceRule
	if in.Audience != nil && opts.kind != "test" {
		if in.Channel == "sms" || in.Channel == "email" {
			return "", badRequest("AUDIENCE_NOT_ALLOWED", "audience", "player filters work with in-app and push notifications")
		}
		a, err := cleanAudience(ctx, tx, *in.Audience)
		if err != nil {
			return "", err
		}
		if !a.empty() {
			aud = &a
		}
	}
	scheduled, err := cleanTime("scheduledAt", in.ScheduledAt)
	if err != nil {
		return "", err
	}
	if scheduled != nil {
		if opts.kind == "test" || opts.kind == "auto" {
			scheduled = nil
		} else if scheduled.Before(s.now().Add(-time.Minute)) {
			return "", badRequest("SCHEDULE_IN_PAST", "scheduledAt", "pick a time in the future")
		} else if scheduled.After(s.now().Add(366 * 24 * time.Hour)) {
			return "", badRequest("SCHEDULE_TOO_FAR", "scheduledAt", "at most a year ahead")
		} else if !scheduled.After(s.now()) {
			scheduled = nil
		}
	}
	t, err := s.cleanTarget(ctx, tx, in.Target)
	if err != nil {
		return "", err
	}
	var providerID *string
	if in.Channel != "in_app" {
		pch := in.Channel
		if pch == "push_in_app" {
			pch = "push"
		}
		var p provider
		if strings.TrimSpace(in.ProviderID) != "" {
			if p, err = s.loadProvider(ctx, tx, strings.TrimSpace(in.ProviderID), false); err != nil {
				if app, ok := asApp(err); ok && app.Status == http.StatusNotFound {
					return "", badRequest("PROVIDER_NOT_FOUND", "providerId", "provider does not exist")
				}
				return "", err
			}
			if p.Channel != pch {
				return "", badRequest("PROVIDER_CHANNEL_MISMATCH", "providerId", "provider channel must be "+pch)
			}
			if !p.Active {
				return "", badRequest("PROVIDER_INACTIVE", "providerId", "provider is not active")
			}
		} else if p, err = s.defaultProvider(ctx, tx, pch); err != nil {
			return "", err
		}
		if !providerConfigured(p) {
			return "", badRequest("PROVIDER_NOT_CONFIGURED", "providerId", "provider is missing required fields")
		}
		providerID = &p.ID
	}
	id := uuid.Must(uuid.NewV4()).String()
	targetJSON, _ := json.Marshal(t)
	var actionJSON *string
	if in.Action != nil {
		b, _ := json.Marshal(in.Action)
		v := string(b)
		actionJSON = &v
	}
	audJSON := "{}"
	if aud != nil {
		b, _ := json.Marshal(aud.input())
		audJSON = string(b)
	}
	status := "queued"
	if scheduled != nil {
		status = "scheduled"
	}
	var ruleID *string
	if opts.ruleID != "" {
		ruleID = &opts.ruleID
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO publication_notification (id, channel, title, body, action, target, provider, status, created_by, kind, name, image_url, category_id, audience, scheduled_at, rule_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14::jsonb,$15,$16)`,
		id, in.Channel, i18nJSON(in.Title), i18nJSON(in.Body), actionJSON, string(targetJSON), providerID, status, createdBy, opts.kind, name, image, cat, audJSON, scheduled, ruleID); err != nil {
		return "", err
	}
	if scheduled != nil {
		return id, nil
	}
	if err := s.materialize(ctx, tx, id, in.Channel, t, aud, deref(cat), opts); err != nil {
		return "", err
	}
	return id, refreshCounts(ctx, tx, id)
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (s *service) createNotification(ctx context.Context, in notificationInput, createdBy string) (string, error) {
	var id string
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var err error
		id, err = s.createNotificationTx(ctx, tx, in, createdBy, notifyOpts{})
		return err
	})
	return id, err
}

type audienceView struct {
	tags    map[string]map[string]bool
	devices map[string][]pushDevice
	country map[string]string
	lang    map[string]string
	created map[string]time.Time
	last    map[string]time.Time
}

func (s *service) loadAudienceView(ctx context.Context, q txQueryer, ids []string, needLast bool) (*audienceView, error) {
	v := &audienceView{tags: map[string]map[string]bool{}, country: map[string]string{}, lang: map[string]string{}, created: map[string]time.Time{}, last: map[string]time.Time{}}
	arr := uuidArray(ids)
	rows, err := q.QueryContext(ctx, `SELECT user_id::text, tag_id::text FROM base_user_tag WHERE user_id = ANY($1::uuid[])`, arr)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var u, t string
		if err := rows.Scan(&u, &t); err != nil {
			rows.Close()
			return nil, err
		}
		if v.tags[u] == nil {
			v.tags[u] = map[string]bool{}
		}
		v.tags[u][t] = true
	}
	rows.Close()
	set, err := localized.Load(ctx, q)
	if err != nil {
		return nil, err
	}
	rows, err = q.QueryContext(ctx, `SELECT id::text, COALESCE(metadata->>'country',''), COALESCE(lang_tag,''), create_time FROM users WHERE id = ANY($1::uuid[])`, arr)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var u, c, l string
		var ct time.Time
		if err := rows.Scan(&u, &c, &l, &ct); err != nil {
			rows.Close()
			return nil, err
		}
		v.country[u] = strings.ToUpper(strings.TrimSpace(c))
		v.lang[u] = set.Choose(l)
		v.created[u] = ct
	}
	rows.Close()
	rows, err = q.QueryContext(ctx, `SELECT id::text, user_id::text, token, platform, app_version FROM publication_push_device WHERE user_id = ANY($1::uuid[]) AND is_active`, arr)
	if err != nil {
		return nil, err
	}
	v.devices = map[string][]pushDevice{}
	for rows.Next() {
		var d pushDevice
		if err := rows.Scan(&d.ID, &d.UserID, &d.Token, &d.Platform, &d.AppVersion); err != nil {
			rows.Close()
			return nil, err
		}
		v.devices[d.UserID] = append(v.devices[d.UserID], d)
	}
	rows.Close()
	if needLast {
		rows, err = q.QueryContext(ctx, `SELECT tot.user_id::text, tot.last_at FROM economy_player_event_total tot JOIN economy_event_type t ON t.id=tot.type_id AND t.key='login' WHERE tot.user_id = ANY($1::uuid[])`, arr)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var u string
			var at time.Time
			if err := rows.Scan(&u, &at); err != nil {
				rows.Close()
				return nil, err
			}
			v.last[u] = at
		}
		rows.Close()
	}
	return v, nil
}

func (s *service) audienceMatch(ctx context.Context, q queryer, a *audienceRule, view *audienceView, userID string, now time.Time, zone *time.Location) (string, error) {
	if a == nil {
		return "", nil
	}
	v := &viewer{UserID: userID, Country: view.country[userID], Lang: view.lang[userID], Tags: view.tags[userID], Created: view.created[userID], Now: now, Zone: zone, q: q}
	if v.Tags == nil {
		v.Tags = map[string]bool{}
	}
	if a.InactiveDays > 0 {
		last := view.created[userID]
		if l, ok := view.last[userID]; ok && l.After(last) {
			last = l
		}
		if now.Sub(last) < time.Duration(a.InactiveDays)*24*time.Hour {
			return "active", nil
		}
	}
	if len(a.Platforms) == 0 && a.MinVersion == "" && a.MaxVersion == "" {
		return a.match(ctx, v)
	}
	reason := "platform"
	for _, d := range view.devices[userID] {
		v.Platform, v.Version = d.Platform, d.AppVersion
		r, err := a.match(ctx, v)
		if err != nil {
			return "", err
		}
		if r == "" {
			return "", nil
		}
		reason = r
	}
	return reason, nil
}

func (s *service) materialize(ctx context.Context, tx *sql.Tx, id, channel string, t target, aud *audienceRule, categoryID string, opts notifyOpts) error {
	type rec struct {
		U string            `json:"u"`
		S string            `json:"s"`
		A *string           `json:"a"`
		E *string           `json:"e"`
		V map[string]string `json:"v"`
	}
	now := s.now()
	zone := projectZone(ctx, tx)
	noAddress, optedOut, noDevice := "no address", "opted_out", "no device"
	return s.eachTargetBatch(ctx, tx, t, func(batch []targetUser) error {
		ids := make([]string, len(batch))
		for i, u := range batch {
			ids[i] = u.ID
		}
		var view *audienceView
		if aud != nil || channel == "push" {
			var err error
			if view, err = s.loadAudienceView(ctx, tx, ids, aud != nil && aud.InactiveDays > 0); err != nil {
				return err
			}
		}
		out := map[string]bool{}
		if categoryID != "" && !opts.trusted {
			rows, err := tx.QueryContext(ctx, `SELECT u::text FROM unnest($2::uuid[]) u WHERE NOT `+categoryAllowsSQL("u", "$1::uuid"), categoryID, uuidArray(ids))
			if err != nil {
				return err
			}
			for rows.Next() {
				var uid string
				if err := rows.Scan(&uid); err != nil {
					rows.Close()
					return err
				}
				out[uid] = true
			}
			rows.Close()
		}
		var names map[string]string
		if opts.vars == nil && channel != "sms" && channel != "email" {
			var err error
			if names, err = playerNames(ctx, tx, ids); err != nil {
				return err
			}
		}
		recs := make([]rec, 0, len(batch))
		for _, u := range batch {
			if aud != nil {
				reason, err := s.audienceMatch(ctx, tx, aud, view, u.ID, now, zone)
				if err != nil {
					return err
				}
				if reason != "" {
					continue
				}
			}
			r := rec{U: u.ID, S: "pending", V: opts.vars[u.ID]}
			if names != nil {
				r.V = map[string]string{"player": names[u.ID]}
			}
			switch {
			case out[u.ID]:
				r.S, r.E = "skipped", &optedOut
			case channel == "sms" || channel == "email":
				if addr, ok := addressFor(channel, u); ok {
					a := addr
					r.A = &a
				} else {
					r.S, r.E = "skipped", &noAddress
				}
			case channel == "push" && len(view.devices[u.ID]) == 0:
				r.S, r.E = "skipped", &noDevice
			}
			recs = append(recs, r)
		}
		if len(recs) == 0 {
			return nil
		}
		raw, _ := json.Marshal(recs)
		_, err := tx.ExecContext(ctx, `INSERT INTO publication_notification_recipient (notification_id, user_id, status, address, error, vars)
			SELECT $1, x.u, x.s, x.a, x.e, COALESCE(x.v, '{}'::jsonb) FROM jsonb_to_recordset($2::jsonb) AS x(u uuid, s text, a text, e text, v jsonb) ON CONFLICT DO NOTHING`, id, string(raw))
		return err
	})
}

func (s *service) startScheduled(ctx context.Context) (int, error) {
	started := 0
	for i := 0; i < 5; i++ {
		var got bool
		err := s.inTx(ctx, func(tx *sql.Tx) error {
			var id, channel, targetRaw, audRaw string
			var cat sql.NullString
			err := tx.QueryRowContext(ctx, `SELECT id::text, channel, target::text, audience::text, category_id::text FROM publication_notification
				WHERE status='scheduled' AND scheduled_at <= now() ORDER BY scheduled_at, id LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&id, &channel, &targetRaw, &audRaw, &cat)
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			if err != nil {
				return err
			}
			got = true
			var t target
			_ = json.Unmarshal([]byte(targetRaw), &t)
			var aud *audienceRule
			var ai audienceInput
			if json.Unmarshal([]byte(audRaw), &ai) == nil {
				if a, err := cleanAudience(ctx, tx, ai); err == nil && !a.empty() {
					aud = &a
				} else if err != nil {
					s.logger.Warn("Publication campaign %s: player filter no longer valid (%v); sending without it is refused", id, err)
					_, err2 := tx.ExecContext(ctx, `UPDATE publication_notification SET status='failed', updated_at=now() WHERE id=$1`, id)
					return err2
				}
			}
			if err := s.materialize(ctx, tx, id, channel, t, aud, cat.String, notifyOpts{}); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `UPDATE publication_notification SET status='queued', updated_at=now() WHERE id=$1`, id); err != nil {
				return err
			}
			return refreshCounts(ctx, tx, id)
		})
		if err != nil {
			return started, err
		}
		if !got {
			break
		}
		started++
	}
	return started, nil
}

func (s *service) consoleNotificationCancel(w http.ResponseWriter, r *http.Request) {
	id := pathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.writeError(w, notFound("NOTIFICATION_NOT_FOUND"))
		return
	}
	who := consoleUser(r)
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		var status string
		err := tx.QueryRowContext(r.Context(), `SELECT status FROM publication_notification WHERE id=$1 FOR UPDATE`, id).Scan(&status)
		if errors.Is(err, sql.ErrNoRows) {
			return notFound("NOTIFICATION_NOT_FOUND")
		}
		if err != nil {
			return err
		}
		if status != "scheduled" && status != "queued" && status != "sending" {
			return conflict("NOT_CANCELLABLE", "only scheduled or sending notifications can be cancelled")
		}
		if _, err := tx.ExecContext(r.Context(), `UPDATE publication_notification_recipient SET status='skipped', error='cancelled' WHERE notification_id=$1 AND status='pending'`, id); err != nil {
			return err
		}
		if _, err := tx.ExecContext(r.Context(), `UPDATE publication_notification SET status='cancelled', cancelled_by=$2, cancelled_at=now(), updated_at=now() WHERE id=$1`, id, who); err != nil {
			return err
		}
		return refreshCounts(r.Context(), tx, id)
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	out, err := s.notificationDetail(r.Context(), id)
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *service) consoleNotificationTest(w http.ResponseWriter, r *http.Request) {
	var in struct {
		notificationInput
		UserID string `json:"userId"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	uid, err := cleanUUID("userId", in.UserID, true)
	if err != nil {
		s.writeError(w, err)
		return
	}
	n := in.notificationInput
	n.Target = target{Mode: "users", UserIDs: []string{*uid}}
	n.ScheduledAt = ""
	n.Audience = nil
	names, err := playerNames(r.Context(), s.db, []string{*uid})
	if err != nil {
		s.writeError(w, err)
		return
	}
	var id string
	err = s.inTx(r.Context(), func(tx *sql.Tx) error {
		var err error
		id, err = s.createNotificationTx(r.Context(), tx, n, consoleUser(r), notifyOpts{kind: "test", trusted: true, vars: map[string]map[string]string{*uid: {"player": names[*uid]}}})
		return err
	})
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

func (s *service) consoleNotificationPreview(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Channel    string         `json:"channel"`
		Target     target         `json:"target"`
		Audience   *audienceInput `json:"audience"`
		CategoryID string         `json:"categoryId"`
	}
	if err := readBody(r, &in); err != nil {
		s.writeError(w, err)
		return
	}
	var out map[string]any
	err := s.inTx(r.Context(), func(tx *sql.Tx) error {
		t, err := s.cleanTarget(r.Context(), tx, in.Target)
		if err != nil {
			return err
		}
		var aud *audienceRule
		if in.Audience != nil {
			a, err := cleanAudience(r.Context(), tx, *in.Audience)
			if err != nil {
				return err
			}
			if !a.empty() {
				aud = &a
			}
		}
		cat, err := cleanCategoryID(r.Context(), tx, in.CategoryID, false)
		if err != nil {
			return err
		}
		now := s.now()
		zone := projectZone(r.Context(), tx)
		var count, targeted, opted, withPush, withPhone, withEmail int64
		err = s.eachTargetBatch(r.Context(), tx, t, func(batch []targetUser) error {
			ids := make([]string, len(batch))
			for i, u := range batch {
				ids[i] = u.ID
			}
			view, err := s.loadAudienceView(r.Context(), tx, ids, aud != nil && aud.InactiveDays > 0)
			if err != nil {
				return err
			}
			off := map[string]bool{}
			if cat != nil {
				rows, err := tx.QueryContext(r.Context(), `SELECT u::text FROM unnest($2::uuid[]) u WHERE NOT `+categoryAllowsSQL("u", "$1::uuid"), *cat, uuidArray(ids))
				if err != nil {
					return err
				}
				for rows.Next() {
					var id string
					if err := rows.Scan(&id); err != nil {
						rows.Close()
						return err
					}
					off[id] = true
				}
				rows.Close()
			}
			for _, u := range batch {
				targeted++
				reason, err := s.audienceMatch(r.Context(), tx, aud, view, u.ID, now, zone)
				if err != nil {
					return err
				}
				if reason != "" {
					continue
				}
				count++
				if off[u.ID] {
					opted++
					continue
				}
				if len(view.devices[u.ID]) > 0 {
					withPush++
				}
				if _, ok := u.phone(); ok {
					withPhone++
				}
				if _, ok := u.email(); ok {
					withEmail++
				}
			}
			return nil
		})
		out = map[string]any{"count": count, "targeted": targeted, "optedOut": opted, "withPush": withPush, "withPhone": withPhone, "withEmail": withEmail}
		return err
	})
	if err != nil {
		s.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
