package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

const newsPublishedCode = 1320

const newsVisibleSQL = `n.status = 'published' AND n.published_at <= now() AND (n.expires_at IS NULL OR n.expires_at > now())`

func (s *service) newsLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.workerEvery)
	defer ticker.Stop()
	sweep := time.NewTicker(30 * time.Minute)
	defer sweep.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := s.runNewsNotify(ctx); err != nil && ctx.Err() == nil {
				s.logger.Error("Publication news notify failed: %v", err)
			}
		case <-sweep.C:
			ok, err := clusterkit.AcquireLease(ctx, s.db, "publication-news-sweep", holder, 29*time.Minute)
			if err != nil || !ok {
				continue
			}
			if err := s.sweepNewsIngest(ctx); err != nil && ctx.Err() == nil {
				s.logger.Error("Publication news sweep failed: %v", err)
			}
		}
	}
}

type dueNews struct {
	ID, Mode, Audience, Title, Summary, Body, Format, Cover string
}

func (s *service) runNewsNotify(ctx context.Context) (int, error) {
	rows, err := s.db.QueryContext(ctx, `UPDATE publication_news SET notify_claimed_until = now() + interval '2 minutes'
		WHERE id IN (SELECT n.id FROM publication_news n WHERE n.notified_at IS NULL AND `+newsVisibleSQL+`
			AND (n.notify_claimed_until IS NULL OR n.notify_claimed_until < now()) ORDER BY n.published_at LIMIT 20 FOR UPDATE SKIP LOCKED)
		RETURNING id::text`)
	if err != nil {
		return 0, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	st, err := loadNewsSettings(ctx, s.db)
	if err != nil {
		return 0, err
	}
	done := 0
	for _, id := range ids {
		title, fired, err := s.notifyNews(ctx, id, st)
		if err != nil {
			s.logger.Error("Publication news %s notify failed: %v", id, err)
			continue
		}
		if fired {
			done++
			if st.RealtimeOnPublish {
				if err := s.nk.NotificationSendAll(ctx, "news", map[string]any{"event": "news_published", "newsId": id, "title": title}, newsPublishedCode, false); err != nil {
					s.logger.Warn("Publication news %s realtime notice failed: %v", id, err)
				}
			}
		}
	}
	return done, nil
}

func newsChannel(mode string) string {
	switch mode {
	case "in_app":
		return "in_app"
	case "push":
		return "push"
	case "both":
		return "push_in_app"
	}
	return ""
}

func shortI18n(langs *localized.Set, t i18n, max int) i18n {
	out := i18n{}
	for k, v := range t {
		if v = truncRunes(v, max); v != "" {
			out[k] = v
		}
	}
	if out[langs.Default] == "" {
		if v := truncRunes(langs.Resolve(t, langs.Default), max); v != "" {
			out[langs.Default] = v
		}
	}
	return out
}

func (s *service) notifyNews(ctx context.Context, id string, st newsSettings) (string, bool, error) {
	var title string
	fired := false
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		var d dueNews
		var notified sql.NullTime
		err := tx.QueryRowContext(ctx, `SELECT n.notify_mode, n.notify_audience::text, n.title::text, n.summary::text, n.body::text, n.body_format, n.cover_url, n.notified_at
			FROM publication_news n WHERE n.id=$1 AND `+newsVisibleSQL+` FOR UPDATE`, id).Scan(&d.Mode, &d.Audience, &d.Title, &d.Summary, &d.Body, &d.Format, &d.Cover, &notified)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if notified.Valid {
			return nil
		}
		langs, err := localized.Load(ctx, tx)
		if err != nil {
			return err
		}
		titles := parseI18n(d.Title)
		title = langs.Resolve(titles, langs.Default)
		var noteID any
		if ch := newsChannel(d.Mode); ch != "" {
			var catID string
			if err := tx.QueryRowContext(ctx, `SELECT id::text FROM publication_notify_category WHERE key=$1`, st.NotifyCategoryKey).Scan(&catID); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			bodies := parseI18n(d.Summary)
			if len(bodies) == 0 {
				bodies = i18n{}
				for k, v := range parseI18n(d.Body) {
					bodies[k] = newsBodyText(d.Format, v)
				}
			}
			in := notificationInput{Channel: ch, Title: shortI18n(langs, titles, 120), Body: shortI18n(langs, bodies, 500), Target: target{Mode: "all"},
				CategoryID: catID, Name: "news: " + truncRunes(title, 100), Action: &action{Type: "goto", Value: "news/" + id}}
			if d.Cover != "" && isPushChannel(ch) && safeURL(d.Cover) {
				in.ImageURL = d.Cover
			}
			if d.Audience != "" && d.Audience != "{}" {
				var aud audienceInput
				if err := json.Unmarshal([]byte(d.Audience), &aud); err == nil {
					in.Audience = &aud
				}
			}
			if _, err := tx.ExecContext(ctx, `SAVEPOINT news_note`); err != nil {
				return err
			}
			nid, err := s.createNotificationTx(ctx, tx, in, "news", notifyOpts{kind: "news"})
			if err != nil {
				s.logger.Warn("Publication news %s notification refused: %v", id, err)
				if _, err := tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT news_note`); err != nil {
					return err
				}
			} else {
				noteID = nid
				if _, err := tx.ExecContext(ctx, `UPDATE publication_notification SET news_id=$2 WHERE id=$1`, nid, id); err != nil {
					return err
				}
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE publication_news SET notified_at=now(), notification_id=$2, notify_claimed_until=NULL WHERE id=$1`, id, noteID); err != nil {
			return err
		}
		fired = true
		return nil
	})
	return title, fired, err
}
