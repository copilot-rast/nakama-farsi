package main

import (
	"context"
	"database/sql"
	"errors"
	"math"

	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/campaignhooks"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

const maxPayTries = 8

type plan struct {
	GrantID    string           `json:"grantId"`
	Currencies map[string]int64 `json:"currencies"`
	Prizes     []prizeRef       `json:"prizes"`
	SourceID   string           `json:"sourceId"`
	Event      *eventRef        `json:"event,omitempty"`
	Campaign   *campaignRef     `json:"campaign,omitempty"`
	Activity   bool             `json:"activity"`
}

type prizeRef struct {
	PrizeKey      string `json:"prizeKey,omitempty"`
	PrizeGroupKey string `json:"prizeGroupKey,omitempty"`
	SourceID      string `json:"sourceId"`
	For           string `json:"for"`
	Paid          bool   `json:"paid"`
	Error         string `json:"error,omitempty"`
}

type eventRef struct {
	Completed string         `json:"completed"`
	Passed    string         `json:"passed"`
	Correct   string         `json:"correct"`
	Analytics string         `json:"analytics"`
	Context   string         `json:"context"`
	Score     float64        `json:"score"`
	Right     int            `json:"right"`
	Pass      bool           `json:"pass"`
	Fields    map[string]any `json:"fields"`
	Sent      bool           `json:"sent"`
}

type campaignRef struct {
	AttemptID string  `json:"attemptId"`
	StageKey  string  `json:"stageKey"`
	Success   bool    `json:"success"`
	Elapsed   float64 `json:"elapsed"`
	Done      bool    `json:"done"`
	Error     string  `json:"error,omitempty"`
}

func (s *service) finish(ctx context.Context, sessionID string, expired bool) (map[string]any, error) {
	cfg := s.settings(ctx)
	replayed := false
	err := s.inTx(ctx, func(tx *sql.Tx) error {
		replayed = false
		sess, err := scanSession(tx.QueryRowContext(ctx, `SELECT `+sessionCols+` FROM quiz_session WHERE id=$1 FOR UPDATE`, sessionID))
		if err != nil {
			return err
		}
		switch sess.Status {
		case "paying", "done":
			replayed = true
			return nil
		case "closed":
			return apierr.Conflict("QUIZ_SESSION_CLOSED").With("sessionId", sessionID)
		}
		k, err := loadPack(ctx, tx, `k.id=$1`, sess.PackID)
		if err != nil {
			return err
		}
		res, pl, err := s.score(ctx, tx, cfg, sess, k, expired)
		if err != nil {
			return err
		}
		pl.GrantID = "quiz:" + sess.ID
		_, err = tx.ExecContext(ctx, `UPDATE quiz_session SET status='paying', completed_at=$2, score=$3, passed=$4, correct=$5, result=$6, plan=$7 WHERE id=$1`,
			sess.ID, s.now().UTC(), res["score"], res["passed"], res["correct"], jsonOf(res), jsonOf(pl))
		return err
	})
	if err != nil {
		return nil, err
	}
	res, err := s.pay(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if replayed {
		res["replayed"] = true
	}
	return res, nil
}

func (s *service) score(ctx context.Context, tx *sql.Tx, cfg settings.All, sess *sessionRow, k *pack, expired bool) (map[string]any, plan, error) {
	now := s.now().UTC()
	answers, err := loadAnswers(ctx, tx, sess.ID)
	if err != nil {
		return nil, plan{}, err
	}
	qs, err := s.sessionQuestions(ctx, tx, sess)
	if err != nil {
		return nil, plan{}, err
	}
	total := len(sess.Questions)
	correct := 0
	for _, a := range answers {
		if a.Outcome == "correct" {
			correct++
		}
	}
	need := passNeed(k, cfg, total)
	passed := total > 0 && correct >= need
	score := scoreOf(correct, total)
	userID := sess.UserID
	pay := map[string]int64{}
	var prizes []prizeRef
	limited := ""

	if _, err := tx.ExecContext(ctx, `INSERT INTO quiz_user_pack (user_id, pack_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, userID, k.ID); err != nil {
		return nil, plan{}, err
	}
	var rewards int
	var firstPass sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT rewards, first_pass_at FROM quiz_user_pack WHERE user_id=$1 AND pack_id=$2 FOR UPDATE`, userID, k.ID).Scan(&rewards, &firstPass); err != nil {
		return nil, plan{}, err
	}
	packPaid := false
	if passed && !k.Reward.empty() && (cfg.Rewards.PackMode == settings.RewardEvery || cfg.Rewards.PackMode == settings.RewardFirst && rewards == 0) {
		day := dayKey(now, s.zone(cfg))
		_, rewarded, _, err := lockDay(ctx, tx, userID, day)
		if err != nil {
			return nil, plan{}, err
		}
		if cfg.Limits.DailyRewardedPasses > 0 && rewarded >= cfg.Limits.DailyRewardedPasses {
			limited = "daily_limit"
		} else {
			if _, err := tx.ExecContext(ctx, `UPDATE quiz_user_day SET rewarded_passes=rewarded_passes+1 WHERE user_id=$1 AND day=$2`, userID, day); err != nil {
				return nil, plan{}, err
			}
			packPaid = true
			addCurrencies(pay, k.Reward.Currencies)
			if k.Reward.PrizeKey != "" || k.Reward.PrizeGroupKey != "" {
				src := "pack:" + userID + ":" + k.ID
				if cfg.Rewards.PackMode == settings.RewardEvery {
					src = "session:" + sess.ID
				}
				prizes = append(prizes, prizeRef{PrizeKey: k.Reward.PrizeKey, PrizeGroupKey: k.Reward.PrizeGroupKey, SourceID: src, For: "pack"})
			}
		}
	}
	inc := 0
	if packPaid {
		inc = 1
	}
	won := 0
	if passed {
		won = 1
	}
	if _, err := tx.ExecContext(ctx, `UPDATE quiz_user_pack SET wins=wins+$3, rewards=rewards+$4, best_score=GREATEST(best_score,$5), last_score=$5, last_correct=$6, last_total=$7,
		first_pass_at=CASE WHEN $8 AND first_pass_at IS NULL THEN $9 ELSE first_pass_at END WHERE user_id=$1 AND pack_id=$2`,
		userID, k.ID, won, inc, score, correct, total, passed, now); err != nil {
		return nil, plan{}, err
	}

	questionPay := map[string]int64{}
	if cfg.Rewards.QuestionMode != settings.RewardNone {
		for _, p := range sess.Questions {
			q := qs[p.ID]
			if q == nil || q.Reward.empty() || answers[q.ID].Outcome != "correct" {
				continue
			}
			src := "session:" + sess.ID + ":q:" + q.ID
			if cfg.Rewards.QuestionMode == settings.QuestionFirst {
				res, err := tx.ExecContext(ctx, `INSERT INTO quiz_user_question (user_id, question_id, rewarded_at) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, userID, q.ID, now)
				if err != nil {
					return nil, plan{}, err
				}
				if n, _ := res.RowsAffected(); n == 0 {
					continue
				}
				src = "question:" + userID + ":" + q.ID
			}
			addCurrencies(questionPay, q.Reward.Currencies)
			if q.Reward.PrizeKey != "" || q.Reward.PrizeGroupKey != "" {
				prizes = append(prizes, prizeRef{PrizeKey: q.Reward.PrizeKey, PrizeGroupKey: q.Reward.PrizeGroupKey, SourceID: src, For: "question"})
			}
		}
	}

	completed := []string{}
	missionPay := map[string]int64{}
	if passed && k.MissionID != "" {
		w, err := loadWorld(ctx, tx)
		if err != nil {
			return nil, plan{}, err
		}
		p, err := loadProgress(ctx, tx, userID)
		if err != nil {
			return nil, plan{}, err
		}
		u := p.packs[k.ID]
		if u.FirstPass == nil {
			u.FirstPass = &now
			p.packs[k.ID] = u
		}
		m := w.missions[k.MissionID]
		for m != nil {
			done := true
			if m.Kind == "game" {
				for _, c := range w.missionsOf(m.ID) {
					done = done && p.completed(c.ID)
				}
			} else {
				list := w.packsOf(m.ID)
				done = len(list) > 0
				for _, c := range list {
					done = done && p.passed(c.ID)
				}
			}
			if !done {
				break
			}
			var rewarded bool
			if _, err := tx.ExecContext(ctx, `INSERT INTO quiz_user_mission (user_id, mission_id, started_at) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, userID, m.ID, now); err != nil {
				return nil, plan{}, err
			}
			if err := tx.QueryRowContext(ctx, `SELECT rewarded FROM quiz_user_mission WHERE user_id=$1 AND mission_id=$2 FOR UPDATE`, userID, m.ID).Scan(&rewarded); err != nil {
				return nil, plan{}, err
			}
			if !rewarded {
				if _, err := tx.ExecContext(ctx, `UPDATE quiz_user_mission SET completed_at=COALESCE(completed_at,$3), rewarded=true WHERE user_id=$1 AND mission_id=$2`, userID, m.ID, now); err != nil {
					return nil, plan{}, err
				}
				completed = append(completed, m.Key)
				addCurrencies(missionPay, m.Reward.Currencies)
				if m.Reward.PrizeKey != "" || m.Reward.PrizeGroupKey != "" {
					prizes = append(prizes, prizeRef{PrizeKey: m.Reward.PrizeKey, PrizeGroupKey: m.Reward.PrizeGroupKey, SourceID: "mission:" + userID + ":" + m.ID, For: m.Kind})
				}
			}
			um := p.missions[m.ID]
			um.Completed, um.hasRow = &now, true
			p.missions[m.ID] = um
			m = w.missions[m.ParentID]
		}
	}
	totalPay := map[string]int64{}
	addCurrencies(totalPay, pay)
	addCurrencies(totalPay, questionPay)
	addCurrencies(totalPay, missionPay)
	serverSeconds := now.Sub(sess.Started).Seconds()
	pl := plan{Currencies: totalPay, Prizes: prizes, SourceID: k.Key, Activity: true}
	fields := map[string]any{"packKey": k.Key, "missionKey": k.MissionKey, "category": k.Category, "score": score, "correct": correct, "total": total, "passed": passed,
		"seconds": math.Round(serverSeconds*100) / 100}
	if (passed || cfg.Events.OnFailure) && (cfg.Events.Completed != "" || cfg.Events.Passed != "" || cfg.Events.Correct != "" || cfg.Events.Analytics != "") {
		pl.Event = &eventRef{Completed: cfg.Events.Completed, Passed: cfg.Events.Passed, Correct: cfg.Events.Correct, Analytics: cfg.Events.Analytics, Context: k.Key,
			Score: score, Right: correct, Pass: passed, Fields: fields}
	}
	if sess.CampaignAttempt != "" && (passed || cfg.Campaign.SendFailures) {
		pl.Campaign = &campaignRef{AttemptID: sess.CampaignAttempt, StageKey: k.CampaignStageKey, Success: passed, Elapsed: math.Max(0, serverSeconds-1)}
	}
	res := map[string]any{
		"sessionId": sess.ID, "packKey": k.Key, "passed": passed, "score": score, "correct": correct, "total": total, "passCorrect": need,
		"answered": len(answers), "expired": expired, "rewards": rewardList(pay), "questionRewards": rewardList(questionPay), "missionRewards": rewardList(missionPay),
		"missionsCompleted": completed, "rewardLimited": nilIfEmpty(limited), "serverSeconds": math.Round(serverSeconds*100) / 100, "replayed": false,
	}
	if cfg.Reveal.Mode != settings.RevealNever {
		langs, lang, _ := localized.ForUser(ctx, tx, userID, "")
		review := []map[string]any{}
		for _, p := range sess.Questions {
			q := qs[p.ID]
			if q == nil {
				continue
			}
			a, ok := answers[q.ID]
			item := map[string]any{"questionId": q.ID, "outcome": "unanswered"}
			if ok {
				item["outcome"], item["optionIds"], item["text"] = a.Outcome, a.OptionIDs, a.Text

				for key, v := range reveal(langs, lang, cfg, q) {
					item[key] = v
				}
			}
			review = append(review, item)
		}
		res["review"] = review
	}
	return res, pl, nil
}

func (s *service) pay(ctx context.Context, sessionID string) (map[string]any, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	var userID, status string
	var rawRes, rawPlan []byte
	var tries int
	if err := tx.QueryRowContext(ctx, `SELECT user_id::text, status, result, plan, pay_tries FROM quiz_session WHERE id=$1 FOR UPDATE`, sessionID).Scan(&userID, &status, &rawRes, &rawPlan, &tries); err != nil {
		return nil, err
	}
	res := map[string]any{}
	unmarshal(rawRes, &res)
	if status != "paying" {
		return res, nil
	}
	var pl plan
	unmarshal(rawPlan, &pl)
	ok := s.payPlan(ctx, userID, &pl, res)
	next := "done"
	if !ok && tries+1 < maxPayTries {
		next = "paying"
	}
	if _, err := tx.ExecContext(ctx, `UPDATE quiz_session SET status=$2, result=$3, plan=$4, pay_tries=pay_tries+1 WHERE id=$1`, sessionID, next, jsonOf(res), jsonOf(pl)); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *service) payPlan(ctx context.Context, userID string, pl *plan, res map[string]any) bool {
	cfg := s.settings(ctx)
	ok := true
	changes := map[string]int64{}
	for k, v := range pl.Currencies {
		if v > 0 {
			changes[k] = v
		}
	}
	if len(changes) > 0 {
		applied, err := pluginkit.WalletApplied(ctx, s.db, userID, pl.GrantID)
		if err == nil && !applied {
			md := map[string]any{"grantId": pl.GrantID, "module": "quiz", "sourceType": sourceType, "sourceId": pl.SourceID, "reason": "quiz_reward"}
			_, _, err = s.nk.MultiUpdate(ctx, nil, nil, nil, []*runtime.WalletUpdate{{UserID: userID, Changeset: changes, Metadata: md}}, true)
		}
		if err != nil {
			s.logger.Warn("Quiz wallet payment %s failed: %v", pl.GrantID, err)
			return false
		}
	}
	paid := []any{}
	if prev, found := res["prizes"].([]any); found {
		paid = append(paid, prev...)
	}
	for i := range pl.Prizes {
		pr := &pl.Prizes[i]
		if pr.Paid {
			continue
		}
		out, err := s.prize(ctx, economyhooks.PrizeGrant{UserID: userID, PrizeKey: pr.PrizeKey, GroupKey: pr.PrizeGroupKey, GroupMode: cfg.Rewards.PrizeGroupMode,
			SourceType: sourceType, SourceID: pr.SourceID, Reason: "quiz_" + pr.For})
		if err != nil {
			pr.Error = err.Error()
			ok = false
			continue
		}
		pr.Paid, pr.Error = true, ""
		for _, p := range out.Prizes {
			p["for"] = pr.For
			paid = append(paid, p)
		}
	}
	res["prizes"] = paid
	pending := []map[string]any{}
	for _, pr := range pl.Prizes {
		if !pr.Paid {
			pending = append(pending, map[string]any{"prizeKey": pr.PrizeKey, "prizeGroupKey": pr.PrizeGroupKey, "for": pr.For})
		}
	}
	res["pendingPrizes"] = pending
	if c := pl.Campaign; c != nil && !c.Done {
		out, err := s.campaignComplete(ctx, userID, campaignhooks.Completion{AttemptID: c.AttemptID, StageKey: c.StageKey, Success: c.Success, ElapsedSeconds: c.Elapsed,
			Fields: map[string]any{"quizScore": res["score"], "quizCorrect": res["correct"]}})
		switch {
		case err == nil:
			c.Done, c.Error = true, ""
			res["campaign"] = out
		default:
			c.Error = err.Error()
			if e, isAPI := apierr.As(err); isAPI && e.Status < 500 && e.Code != "STAGE_BUSY" {

				c.Done = true
				res["campaign"] = map[string]any{"error": e.Code}
			} else {
				ok = false
			}
		}
	}
	if pl.Event != nil && !pl.Event.Sent {
		pl.Event.Sent = true
		ev := pl.Event
		if ev.Analytics != "" {
			answers, errs := gameevents.Publish(ctx, gameevents.Event{ID: pl.GrantID, UserID: userID, Name: ev.Analytics, Context: ev.Context, Fields: ev.Fields})
			if a, found := answers["analytics"]; found {
				res["trophies"] = a
			}
			for owner, err := range errs {
				s.logger.Debug("Quiz event to %s not recorded: %v", owner, err)
			}
		}
		if playerevents.Available() {
			record := func(key string, value float64) {
				if key == "" {
					return
				}
				if _, err := playerevents.Record(ctx, playerevents.Event{UserID: userID, Key: key, Value: value, Metadata: ev.Fields}); err != nil {
					s.logger.Debug("Quiz economy event %s not recorded: %v", key, err)
				}
			}
			record(ev.Completed, ev.Score)
			if ev.Pass {
				record(ev.Passed, 1)
			}
			if ev.Right > 0 {
				record(ev.Correct, float64(ev.Right))
			}
		}
	}
	if pl.Activity {
		pl.Activity = false
		activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: userID, Kind: activity.KindEvent, Action: "quiz_completed", RefID: pl.SourceID, Actor: "player",
			Data: map[string]any{"packKey": pl.SourceID, "score": res["score"], "passed": res["passed"], "rewards": res["rewards"]}})
	}
	return ok
}

func (s *service) closeExpired(ctx context.Context, limit int) {
	cfg := s.settings(ctx)
	rows, err := s.db.QueryContext(ctx, `SELECT id::text FROM quiz_session WHERE status='open' AND (expires_at < now() OR deadline_at < now() - make_interval(secs => $2)) ORDER BY started_at LIMIT $1`, limit, cfg.Timing.GraceSeconds+5)
	if err != nil {
		return
	}
	var ids []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		if _, err := s.finish(ctx, id, true); err != nil {
			var e *apierr.Error
			if !errors.As(err, &e) {
				s.logger.Warn("Quiz could not close session %s: %v", id, err)
			}
		}
	}
}

func (s *service) retryPayments(ctx context.Context, limit int) {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text FROM quiz_session WHERE status='paying' AND completed_at < now() - interval '30 seconds' ORDER BY completed_at LIMIT $1`, limit)
	if err != nil {
		return
	}
	var ids []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		if _, err := s.pay(ctx, id); err != nil {
			s.logger.Warn("Quiz could not pay session %s: %v", id, err)
		}
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM quiz_user_hint WHERE status='paying' AND created_at < now() - interval '10 minutes'`)
}
