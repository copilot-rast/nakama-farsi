package main

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

func (s *service) rpcHintBuy(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		HintID    string `json:"hintId"`
		SessionID string `json:"sessionId"`
		Lang      string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	in.HintID = strings.TrimSpace(in.HintID)
	if _, err := uuid.FromString(in.HintID); err != nil {
		return nil, apierr.NotFound("QUIZ_HINT_NOT_FOUND").WithField("hintId")
	}
	cfg := s.settings(ctx)
	h, err := scanHint(s.db.QueryRowContext(ctx, `SELECT `+hintCols+` FROM quiz_hint h JOIN quiz_question q ON q.id=h.question_id WHERE h.id=$1 AND h.is_active AND q.is_active`, in.HintID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apierr.NotFound("QUIZ_HINT_NOT_FOUND").WithField("hintId")
	}
	if err != nil {
		return nil, err
	}
	scope := ""
	if cfg.Hints.Keep == settings.KeepSession || strings.TrimSpace(in.SessionID) != "" {
		sess, _, err := s.ownSession(ctx, userID, in.SessionID)
		if err != nil {
			return nil, err
		}
		if sess.index(h.QuestionID) < 0 {
			return nil, apierr.BadRequest("QUIZ_QUESTION_NOT_IN_SESSION").WithField("hintId")
		}
		if cfg.Hints.Keep == settings.KeepSession {
			if sess.Status != "open" {
				return nil, apierr.Conflict("QUIZ_SESSION_CLOSED").With("sessionId", sess.ID)
			}
			scope = sess.ID
		}
	}
	price := hintPrice(cfg, h)
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	answer := func(replayed bool) map[string]any {
		return map[string]any{"hintId": h.ID, "questionId": h.QuestionID, "name": text(langs, lang, h.Name), "text": text(langs, lang, h.Text), "mediaUrl": h.MediaURL,
			"price": price, "owned": true, "replayed": replayed}
	}
	var status string
	err = s.db.QueryRowContext(ctx, `SELECT status FROM quiz_user_hint WHERE user_id=$1 AND hint_id=$2 AND scope=$3`, userID, h.ID, scope).Scan(&status)
	if err == nil && status == "done" {
		return answer(true), nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		err = s.inTx(ctx, func(tx *sql.Tx) error {
			if cfg.Hints.MaxPerQuestion > 0 {
				var n int
				if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM quiz_user_hint u JOIN quiz_hint x ON x.id=u.hint_id WHERE u.user_id=$1 AND u.scope=$2 AND x.question_id=$3`,
					userID, scope, h.QuestionID).Scan(&n); err != nil {
					return err
				}
				if n >= cfg.Hints.MaxPerQuestion {
					return apierr.Conflict("QUIZ_HINT_LIMIT").With("limit", cfg.Hints.MaxPerQuestion)
				}
			}
			_, err := tx.ExecContext(ctx, `INSERT INTO quiz_user_hint (user_id, hint_id, scope, price, status) VALUES ($1,$2,$3,$4,'paying') ON CONFLICT DO NOTHING`, userID, h.ID, scope, jsonOf(price))
			return err
		})
		if err != nil {
			return nil, err
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if err := tx.QueryRowContext(ctx, `SELECT status FROM quiz_user_hint WHERE user_id=$1 AND hint_id=$2 AND scope=$3 FOR UPDATE`, userID, h.ID, scope).Scan(&status); err != nil {
		return nil, err
	}
	if status == "done" {
		return answer(true), nil
	}
	refuse := func(e error) (any, error) {
		_, _ = tx.ExecContext(ctx, `DELETE FROM quiz_user_hint WHERE user_id=$1 AND hint_id=$2 AND scope=$3`, userID, h.ID, scope)
		_ = tx.Commit()
		return nil, e
	}
	if price.Amount > 0 {
		if price.CurrencyKey == "" {
			return refuse(apierr.Conflict("QUIZ_HINT_PRICE_INVALID"))
		}
		grant := "quiz:hint:" + userID + ":" + h.ID + ":" + scope
		err := pluginkit.Debit(ctx, s.nk, s.db, userID, grant, map[string]int64{price.CurrencyKey: price.Amount},
			map[string]any{"module": "quiz", "sourceType": sourceType, "sourceId": h.ID, "reason": "quiz_hint"})
		if err != nil {
			var short *pluginkit.ErrInsufficient
			if errors.As(err, &short) {
				return refuse(apierr.BadRequest("QUIZ_INSUFFICIENT_CURRENCY_FOR_HINT").With("currencyKey", price.CurrencyKey).With("amount", price.Amount))
			}
			return nil, err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE quiz_user_hint SET status='done' WHERE user_id=$1 AND hint_id=$2 AND scope=$3`, userID, h.ID, scope); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return answer(false), nil
}

func (s *service) missionView(langs *localized.Set, lang string, w *world, p *progress, m *mission, now time.Time) map[string]any {
	u := p.missions[m.ID]
	items := []map[string]any{}
	unlocked := true
	if m.Kind == "game" {
		for _, c := range w.missionsOf(m.ID) {
			done := p.completed(c.ID)
			items = append(items, map[string]any{"key": c.Key, "name": text(langs, lang, c.Name), "sort": c.Sort, "completed": done, "playable": m.OrderMode != "sequential" || unlocked})
			unlocked = unlocked && done
		}
	} else {
		for _, k := range w.packsOf(m.ID) {
			won := p.passed(k.ID)
			items = append(items, map[string]any{"key": k.Key, "name": text(langs, lang, k.Name), "sort": k.Sort, "passed": won, "playable": m.OrderMode != "sequential" || unlocked})
			unlocked = unlocked && won
		}
	}
	var started any
	if u.hasRow {
		started = u.Started.UTC().Format(time.RFC3339)
	}
	exp := missionExpiry(m, u)
	return map[string]any{"key": m.Key, "kind": m.Kind, "parentKey": m.ParentKey, "name": text(langs, lang, m.Name), "description": text(langs, lang, m.Description),
		"iconUrl": m.IconURL, "sort": m.Sort, "orderMode": m.OrderMode, "timeSeconds": m.TimeSeconds, "reward": rewardView(m.Reward), "metadata": m.Metadata,
		"startedAt": started, "expiresAt": iso(exp), "timeOver": exp != nil && now.After(*exp) && u.Completed == nil, "completedAt": iso(u.Completed), "rewarded": u.Rewarded, "items": items}
}

func (s *service) rpcMissions(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
		Kind string `json:"kind"`
		Key  string `json:"key"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	w, err := loadWorld(ctx, s.db)
	if err != nil {
		return nil, err
	}
	p, err := loadProgress(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	now := s.now()
	if in.Key != "" {
		m := w.byKey[in.Key]
		if m == nil {
			return nil, apierr.NotFound("QUIZ_MISSION_NOT_FOUND").WithField("key")
		}
		return s.missionView(langs, lang, w, p, m, now), nil
	}
	list := []*mission{}
	for _, m := range w.missions {
		if in.Kind == "" || m.Kind == in.Kind {
			list = append(list, m)
		}
	}
	sortMissions(list)
	out := []map[string]any{}
	for _, m := range list {
		out = append(out, s.missionView(langs, lang, w, p, m, now))
	}
	return map[string]any{"lang": lang, "serverTime": now.UTC().Format(time.RFC3339), "missions": out}, nil
}

func sortMissions(list []*mission) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && (list[j].Sort < list[j-1].Sort || list[j].Sort == list[j-1].Sort && list[j].Key < list[j-1].Key); j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}

func (s *service) rpcMissionStart(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Key  string `json:"key"`
		Lang string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	m, err := loadMission(ctx, s.db, `m.key=$1 AND m.is_active`, strings.TrimSpace(in.Key))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apierr.NotFound("QUIZ_MISSION_NOT_FOUND").WithField("key")
	}
	if err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO quiz_user_mission (user_id, mission_id, started_at) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, userID, m.ID, s.now().UTC()); err != nil {
		return nil, err
	}
	raw := `{"key":` + strconv.Quote(m.Key) + `,"lang":` + strconv.Quote(in.Lang) + `}`
	return s.rpcMissions(ctx, userID, raw)
}

func (s *service) rpcStats(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		PackKey string `json:"packKey"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	if in.PackKey != "" {
		var plays, wins int
		var best, last float64
		err := s.db.QueryRowContext(ctx, `SELECT COALESCE(u.plays,0), COALESCE(u.wins,0), COALESCE(u.best_score,0)::float8, COALESCE(u.last_score,0)::float8 FROM quiz_pack k
			LEFT JOIN quiz_user_pack u ON u.pack_id=k.id AND u.user_id=$1 WHERE k.key=$2`, userID, in.PackKey).Scan(&plays, &wins, &best, &last)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierr.NotFound("QUIZ_PACK_NOT_FOUND").WithField("packKey")
		}
		if err != nil {
			return nil, err
		}
		return map[string]any{"packKey": in.PackKey, "plays": plays, "wins": wins, "bestScore": best, "lastScore": last}, nil
	}
	var plays, wins, passed, correct, answered int
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(sum(plays),0), COALESCE(sum(wins),0), count(first_pass_at) FROM quiz_user_pack WHERE user_id=$1`, userID).Scan(&plays, &wins, &passed); err != nil {
		return nil, err
	}
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE outcome='correct'), count(*) FROM quiz_answer WHERE user_id=$1`, userID).Scan(&correct, &answered)
	return map[string]any{"plays": plays, "wins": wins, "packsPassed": passed, "answers": answered, "correctAnswers": correct}, nil
}

func (s *service) rpcReportReasons(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT key, name FROM quiz_report_reason WHERE is_active ORDER BY sort, key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var key string
		var raw []byte
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		name := map[string]string{}
		unmarshal(raw, &name)
		out = append(out, map[string]any{"key": key, "name": text(langs, lang, name)})
	}
	return map[string]any{"lang": lang, "enabled": s.settings(ctx).Reports.Enabled, "reasons": out}, rows.Err()
}

func (s *service) rpcReportQuestion(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		QuestionID string `json:"questionId"`
		ReasonKey  string `json:"reasonKey"`
		Note       string `json:"note"`
		SessionID  string `json:"sessionId"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	if !cfg.Reports.Enabled {
		return nil, apierr.Forbidden("QUIZ_REPORTS_CLOSED")
	}
	in.QuestionID, in.ReasonKey, in.Note = strings.TrimSpace(in.QuestionID), strings.TrimSpace(in.ReasonKey), strings.TrimSpace(in.Note)
	if _, err := uuid.FromString(in.QuestionID); err != nil {
		return nil, apierr.NotFound("QUIZ_QUESTION_NOT_FOUND").WithField("questionId")
	}
	var exists bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM quiz_question WHERE id=$1)`, in.QuestionID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, apierr.NotFound("QUIZ_QUESTION_NOT_FOUND").WithField("questionId")
	}
	var reasonOK bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM quiz_report_reason WHERE key=$1 AND is_active)`, in.ReasonKey).Scan(&reasonOK); err != nil {
		return nil, err
	}
	if !reasonOK {
		return nil, apierr.BadRequest("QUIZ_REPORT_REASON_NOT_FOUND").WithField("reasonKey")
	}
	if utf8.RuneCountInString(in.Note) > 1000 {
		return nil, apierr.BadRequest("QUIZ_REPORT_NOTE_TOO_LONG").WithField("note")
	}
	if cfg.Reports.NoteRequired && in.Note == "" {
		return nil, apierr.BadRequest("QUIZ_REPORT_NOTE_REQUIRED").WithField("note")
	}
	if cfg.Reports.RequireSeen {
		var seen bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM quiz_session WHERE user_id=$1 AND questions @> $2::jsonb)`, userID, `[{"id":"`+in.QuestionID+`"}]`).Scan(&seen); err != nil {
			return nil, err
		}
		if !seen {
			return nil, apierr.Forbidden("QUIZ_REPORT_NOT_SEEN").WithField("questionId")
		}
	}
	var session any
	if _, err := uuid.FromString(strings.TrimSpace(in.SessionID)); err == nil {
		session = strings.TrimSpace(in.SessionID)
	}
	id := uuid.Must(uuid.NewV4()).String()
	dup := false
	now := s.now().UTC()
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		dup = false
		var existing string
		if err := tx.QueryRowContext(ctx, `SELECT id::text FROM quiz_question_report WHERE user_id=$1 AND question_id=$2 AND status='open'`, userID, in.QuestionID).Scan(&existing); err == nil {
			id, dup = existing, true
			return nil
		}
		day := dayKey(now, s.zone(cfg))
		_, _, reports, err := lockDay(ctx, tx, userID, day)
		if err != nil {
			return err
		}
		if cfg.Reports.DailyLimit > 0 && reports >= cfg.Reports.DailyLimit {
			return apierr.TooManyRequests("QUIZ_REPORT_LIMIT", secondsToNextDay(now, s.zone(cfg))).With("limit", cfg.Reports.DailyLimit)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO quiz_question_report (id, user_id, question_id, reason_key, note, session_id, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			id, userID, in.QuestionID, in.ReasonKey, in.Note, session, now); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE quiz_user_day SET reports=reports+1 WHERE user_id=$1 AND day=$2`, userID, day); err != nil {
			return err
		}
		if cfg.Reports.AutoDisableAt > 0 {
			var open int
			if err := tx.QueryRowContext(ctx, `SELECT count(DISTINCT user_id) FROM quiz_question_report WHERE question_id=$1 AND status='open'`, in.QuestionID).Scan(&open); err != nil {
				return err
			}
			if open >= cfg.Reports.AutoDisableAt {
				if _, err := tx.ExecContext(ctx, `UPDATE quiz_question SET is_active=false, updated_by='auto:reports', updated_at=now() WHERE id=$1`, in.QuestionID); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		if isUnique(err) {
			return map[string]any{"reportId": id, "duplicate": true}, nil
		}
		return nil, err
	}
	return map[string]any{"reportId": id, "duplicate": dup}, nil
}
