package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

type noteText struct {
	title, body string
}

func (n notificationRow) textFor(set *localized.Set, lang string, vars map[string]string) noteText {
	return noteText{title: renderVars(set.Resolve(n.Title, lang), vars), body: renderVars(set.Resolve(n.Body, lang), vars)}
}

func (s *service) markSkipped(ctx context.Context, c claimed, note string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE publication_notification_recipient SET status='skipped', error=$4
		WHERE notification_id=$1 AND user_id=$2 AND status='pending' AND attempts=$3`, c.NotificationID, c.UserID, c.Attempts, note)
	return err
}

func (s *service) markSentPush(ctx context.Context, c claimed, note, lang, message string, sent, failed int) error {
	var errText, msgText *string
	if note != "" {
		errText = &note
	}
	if message != "" {
		m := truncate(message, 1000)
		msgText = &m
	}
	_, err := s.db.ExecContext(ctx, `UPDATE publication_notification_recipient SET status='sent', sent_at=now(), error=$4, lang=NULLIF($5,''), message=$6, push_sent=push_sent+$7, push_failed=push_failed+$8
		WHERE notification_id=$1 AND user_id=$2 AND status='pending' AND attempts=$3`, c.NotificationID, c.UserID, c.Attempts, errText, lang, msgText, sent, failed)
	return err
}

func (s *service) inAppSends(n notificationRow, batch []claimed, set *localized.Set, langs map[string]string) []*runtime.NotificationSend {
	var act *action
	_ = json.Unmarshal(n.Action, &act)
	sends := make([]*runtime.NotificationSend, 0, len(batch))
	for _, c := range batch {
		lang := langs[c.UserID]
		if lang == "" {
			lang = set.Default
		}
		txt := n.textFor(set, lang, c.Vars)
		content := map[string]any{"notificationId": n.ID, "title": txt.title, "body": txt.body, "action": resolveAction(set, act, lang), "lang": lang}
		if n.Image != "" {
			content["imageUrl"] = n.Image
		}
		if n.Category != "" {
			content["category"] = n.Category
		}
		sends = append(sends, &runtime.NotificationSend{UserID: c.UserID, Subject: txt.title, Code: inAppCode, Persistent: true, Content: content})
	}
	return sends
}

func (s *service) sendPushBatch(ctx context.Context, n notificationRow, batch []claimed, set *localized.Set, langs map[string]string) error {
	p, perr := s.loadProvider(ctx, s.db, n.Provider, false)
	if perr == nil && !p.Active {
		perr = transientErr("push provider %s is not active", p.Kind)
	} else if perr != nil {
		if app, ok := asApp(perr); ok && app.Status == 404 {
			perr = permanentErr("push provider was deleted")
		}
	}
	withInbox := n.Channel == "push_in_app"
	if withInbox {
		if err := s.nk.NotificationsSend(ctx, s.inAppSends(n, batch, set, langs)); err != nil {
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
	}
	ids := make([]string, len(batch))
	for i, c := range batch {
		ids[i] = c.UserID
	}
	devices, err := s.userDevices(ctx, ids)
	if err != nil {
		return err
	}
	var act *action
	_ = json.Unmarshal(n.Action, &act)
	for _, c := range batch {
		lang := langs[c.UserID]
		if lang == "" {
			lang = set.Default
		}
		txt := n.textFor(set, lang, c.Vars)
		devs := devices[c.UserID]
		if perr != nil || len(devs) == 0 {
			if withInbox {
				note := "no device"
				if perr != nil {
					note = "push: " + perr.Error()
				}
				if err := s.markSentPush(ctx, c, note, lang, "", 0, 0); err != nil {
					return err
				}
				continue
			}
			if perr != nil {
				if err := s.markFailure(ctx, c, perr); err != nil {
					return err
				}
				continue
			}
			if err := s.markSkipped(ctx, c, "no device"); err != nil {
				return err
			}
			continue
		}
		if !p.Sandbox && p.Kind != "log" {
			mine, err := s.renewClaim(ctx, c)
			if err != nil {
				return err
			}
			if !mine {
				continue
			}
		}
		data := map[string]string{"notificationId": n.ID, "lang": lang}
		link := ""
		if a := resolveAction(set, act, lang); a != nil {
			m := a.(map[string]any)
			data["actionType"], _ = m["type"].(string)
			data["actionValue"], _ = m["value"].(string)
			if data["actionType"] == "url" {
				link = data["actionValue"]
			}
		}
		if n.Category != "" {
			data["category"] = n.Category
		}
		sent, failed, transient := 0, 0, false
		var lastErr error
		for _, d := range devs {
			ok, err := s.claimDelivery(ctx, n.ID, d, p.ID)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			sendCtx, cancel := context.WithTimeout(ctx, s.cfg.sendTimeout)
			res, err := s.sendPush(sendCtx, p, pushMessage{Token: d.Token, Platform: d.Platform, Title: txt.title, Body: txt.body, Image: n.Image, Link: link, Data: data})
			cancel()
			switch {
			case err == nil:
				sent++
				if err := s.finishDelivery(ctx, n.ID, d, "sent", res.Reply); err != nil {
					return err
				}
			case res.Dead:
				failed++
				s.logger.Info("Publication push token %s of %s is dead (%s); the device was switched off", maskToken(d.Token), d.UserID, res.Reply)
				if err := s.finishDelivery(ctx, n.ID, d, "dead", err.Error()); err != nil {
					return err
				}
			default:
				failed++
				lastErr = err
				if !isPermanent(err) {
					transient = true
				}
				s.logger.Warn("Publication push via %s to %s failed: %v", p.Kind, maskToken(d.Token), err)
				if err := s.finishDelivery(ctx, n.ID, d, "failed", err.Error()); err != nil {
					return err
				}
			}
		}
		note, preview := "", ""
		if p.Sandbox || p.Kind == "log" {
			note, preview = "sandbox", txt.title+"\n"+txt.body
		}
		switch {
		case sent > 0 || withInbox || (failed == 0 && !transient):
			if err := s.markSentPush(ctx, c, note, lang, preview, sent, failed); err != nil {
				return err
			}
		case transient:
			if err := s.markFailure(ctx, c, transientErr("push failed: %v", lastErr)); err != nil {
				return err
			}
		default:
			if err := s.markFailure(ctx, c, permanentErr("no device took the push")); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *service) pushDeliveryCounts(ctx context.Context, id string) (map[string]int64, error) {
	out := map[string]int64{"sent": 0, "failed": 0, "dead": 0, "sending": 0}
	rows, err := s.db.QueryContext(ctx, `SELECT status, count(*) FROM publication_push_delivery WHERE notification_id=$1 GROUP BY status`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var st string
		var n int64
		if err := rows.Scan(&st, &n); err != nil {
			return nil, err
		}
		out[st] = n
	}
	return out, rows.Err()
}
