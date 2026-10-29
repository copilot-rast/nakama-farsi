package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const inAppCode = 1001

type claimed struct {
	NotificationID string
	UserID         string
	Address        string
	Attempts       int
}

type notificationRow struct {
	ID       string
	Channel  string
	Title    i18n
	Body     i18n
	Action   json.RawMessage
	Provider string
}

func (s *service) workerLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.workerEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for round := 0; round < 20; round++ {
				n, err := s.runWorker(ctx)
				if err != nil {
					if ctx.Err() == nil {
						s.logger.Error("Publication worker failed: %v", err)
					}
					break
				}
				if n < s.cfg.workerBatch {
					break
				}
			}
		}
	}
}

func (s *service) runWorker(ctx context.Context) (int, error) {
	if err := s.sweepExhausted(ctx); err != nil {
		return 0, err
	}
	batch, err := s.claim(ctx)
	if err != nil || len(batch) == 0 {
		return 0, err
	}
	groups := map[string][]claimed{}
	order := []string{}
	for _, c := range batch {
		if _, ok := groups[c.NotificationID]; !ok {
			order = append(order, c.NotificationID)
		}
		groups[c.NotificationID] = append(groups[c.NotificationID], c)
	}
	for _, id := range order {
		if err := s.processNotification(ctx, id, groups[id]); err != nil {
			return len(batch), err
		}
		if err := refreshCounts(ctx, s.db, id); err != nil {
			return len(batch), err
		}
	}
	return len(batch), nil
}

func (s *service) sweepExhausted(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `UPDATE publication_notification_recipient SET status='failed', error=COALESCE(error,'gave up after '||attempts||' attempts')
		WHERE status='pending' AND attempts >= $1 AND next_attempt_at <= now() RETURNING notification_id::text`, s.cfg.maxAttempts)
	if err != nil {
		return err
	}
	touched := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		touched[id] = true
	}
	rows.Close()
	for id := range touched {
		if err := refreshCounts(ctx, s.db, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) claim(ctx context.Context) ([]claimed, error) {
	rows, err := s.db.QueryContext(ctx, `UPDATE publication_notification_recipient r SET attempts = r.attempts + 1, next_attempt_at = now() + make_interval(secs => $2)
		FROM (SELECT r2.notification_id, r2.user_id FROM publication_notification_recipient r2 JOIN publication_notification n ON n.id = r2.notification_id
			WHERE r2.status='pending' AND r2.next_attempt_at <= now() AND r2.attempts < $3 AND n.status IN ('queued','sending')
			ORDER BY r2.next_attempt_at, r2.notification_id, r2.user_id LIMIT $1 FOR UPDATE OF r2 SKIP LOCKED) c
		WHERE r.notification_id = c.notification_id AND r.user_id = c.user_id
		RETURNING r.notification_id::text, r.user_id::text, COALESCE(r.address,''), r.attempts`, s.cfg.workerBatch, s.cfg.claimLease.Seconds(), s.cfg.maxAttempts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []claimed{}
	for rows.Next() {
		var c claimed
		if err := rows.Scan(&c.NotificationID, &c.UserID, &c.Address, &c.Attempts); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *service) loadNotification(ctx context.Context, id string) (notificationRow, error) {
	var n notificationRow
	var title, body string
	var act, prov sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT id::text, channel, title::text, body::text, action::text, provider FROM publication_notification WHERE id=$1`, id).
		Scan(&n.ID, &n.Channel, &title, &body, &act, &prov)
	if err != nil {
		return n, err
	}
	n.Title, n.Body = parseI18n(title), parseI18n(body)
	if act.Valid {
		n.Action = json.RawMessage(act.String)
	} else {
		n.Action = json.RawMessage("null")
	}
	n.Provider = prov.String
	return n, nil
}

func (s *service) userLangs(ctx context.Context, langs *localized.Set, ids []string) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, COALESCE(lang_tag,'') FROM users WHERE id = ANY($1::uuid[])`, uuidArray(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var id, lang string
		if err := rows.Scan(&id, &lang); err != nil {
			return nil, err
		}
		out[id] = langs.Choose(lang)
	}
	return out, rows.Err()
}

func (s *service) markSent(ctx context.Context, c claimed, note, lang, message string) error {
	var errText, msgText *string
	if note != "" {
		errText = &note
	}
	if message != "" {
		m := truncate(message, 1000)
		msgText = &m
	}
	_, err := s.db.ExecContext(ctx, `UPDATE publication_notification_recipient SET status='sent', sent_at=now(), error=$4, lang=NULLIF($5,''), message=$6
		WHERE notification_id=$1 AND user_id=$2 AND status='pending' AND attempts=$3`, c.NotificationID, c.UserID, c.Attempts, errText, lang, msgText)
	return err
}

func (s *service) markFailure(ctx context.Context, c claimed, cause error) error {
	msg := truncate(cause.Error(), 500)
	if isPermanent(cause) || c.Attempts >= s.cfg.maxAttempts {
		_, err := s.db.ExecContext(ctx, `UPDATE publication_notification_recipient SET status='failed', error=$4
			WHERE notification_id=$1 AND user_id=$2 AND status='pending' AND attempts=$3`, c.NotificationID, c.UserID, c.Attempts, msg)
		return err
	}
	delay := s.cfg.retryBackoff * time.Duration(1<<(c.Attempts-1))
	_, err := s.db.ExecContext(ctx, `UPDATE publication_notification_recipient SET error=$4, next_attempt_at=now() + make_interval(secs => $5)
		WHERE notification_id=$1 AND user_id=$2 AND status='pending' AND attempts=$3`, c.NotificationID, c.UserID, c.Attempts, msg, delay.Seconds())
	return err
}

func (s *service) processNotification(ctx context.Context, id string, batch []claimed) error {
	n, err := s.loadNotification(ctx, id)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE publication_notification SET status='sending', updated_at=now() WHERE id=$1 AND status='queued'`, id); err != nil {
		return err
	}
	ids := make([]string, len(batch))
	for i, c := range batch {
		ids[i] = c.UserID
	}
	set, err := localized.Load(ctx, s.db)
	if err != nil {
		return err
	}
	langs, err := s.userLangs(ctx, set, ids)
	if err != nil {
		return err
	}
	if n.Channel == "in_app" {
		return s.sendInApp(ctx, n, batch, set, langs)
	}
	p, perr := s.loadProvider(ctx, s.db, n.Provider, false)
	if perr == nil && !p.Active {
		perr = transientErr("provider %s is not active", p.Kind)
	} else if perr != nil {
		if app, ok := asApp(perr); ok && app.Status == 404 {
			perr = permanentErr("provider was deleted")
		}
	}
	for _, c := range batch {
		if perr != nil {
			if err := s.markFailure(ctx, c, perr); err != nil {
				return err
			}
			continue
		}
		lang := langs[c.UserID]
		if lang == "" {
			lang = set.Default
		}
		msg := outMessage{To: c.Address, Lang: lang, IdempotencyKey: n.ID + "/" + c.UserID}
		body := set.Resolve(n.Body, lang)
		preview := body
		if n.Channel == "email" {
			msg.Subject = set.Resolve(n.Title, lang)
			msg.HTML = emailHTML(msg.Subject, renderMarkdown(body), lang, set.Direction(lang))
			msg.Text = markdownText(body)
			preview = msg.Subject + "\n" + msg.Text
		} else {
			msg.Text = body
		}
		if p.Sandbox {
			s.logger.Info("Publication sandbox %s via %s to %s in %s (notification %s): %s", n.Channel, p.Kind, maskAddress(c.Address), lang, n.ID, truncate(msg.Text, 200))
			if err := s.markSent(ctx, c, "sandbox", lang, preview); err != nil {
				return err
			}
			continue
		}
		sendCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		reply, err := s.deliver(sendCtx, p, msg)
		cancel()
		if err != nil {
			s.logger.Warn("Publication %s via %s to %s failed (attempt %d): %v", n.Channel, p.Kind, maskAddress(c.Address), c.Attempts, err)
			if err := s.markFailure(ctx, c, err); err != nil {
				return err
			}
			continue
		}
		s.logger.Debug("Publication %s via %s to %s: %s", n.Channel, p.Kind, maskAddress(c.Address), reply)
		if err := s.markSent(ctx, c, "", lang, ""); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) sendInApp(ctx context.Context, n notificationRow, batch []claimed, set *localized.Set, langs map[string]string) error {
	var act *action
	_ = json.Unmarshal(n.Action, &act)
	sends := make([]*runtime.NotificationSend, 0, len(batch))
	for _, c := range batch {
		lang := langs[c.UserID]
		if lang == "" {
			lang = set.Default
		}
		title, body := set.Resolve(n.Title, lang), set.Resolve(n.Body, lang)
		sends = append(sends, &runtime.NotificationSend{UserID: c.UserID, Subject: title, Code: inAppCode, Persistent: true,
			Content: map[string]any{"notificationId": n.ID, "title": title, "body": body, "action": resolveAction(set, act, lang), "lang": lang}})
	}
	if err := s.nk.NotificationsSend(ctx, sends); err != nil {
		cause := transientErr("in-app send failed: %v", err)
		for _, c := range batch {
			if err := s.markFailure(ctx, c, cause); err != nil {
				return err
			}
		}
		if errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	}
	for _, c := range batch {
		if err := s.markSent(ctx, c, "", langs[c.UserID], ""); err != nil {
			return err
		}
	}
	return nil
}

func resolveAction(set *localized.Set, a *action, lang string) any {
	if a == nil || a.Type == "" {
		return nil
	}
	out := map[string]any{"type": a.Type, "value": a.Value}
	if label := set.Resolve(a.Label, lang); label != "" {
		out["label"] = label
	}
	return out
}
