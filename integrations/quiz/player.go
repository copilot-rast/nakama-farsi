package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/localized"
)

func iso(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}

func text(langs *localized.Set, lang string, m map[string]string) string {
	if langs == nil {
		if v, ok := m[lang]; ok {
			return v
		}
		for _, v := range m {
			return v
		}
		return ""
	}
	return langs.Resolve(localized.Text(m), lang)
}

func rewardList(cur map[string]int64) []map[string]any {
	out := []map[string]any{}
	keys := make([]string, 0, len(cur))
	for k := range cur {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if cur[k] > 0 {
			out = append(out, map[string]any{"currencyKey": k, "amount": cur[k]})
		}
	}
	return out
}

func rewardView(r Reward) map[string]any {
	return map[string]any{"currencies": rewardList(r.Currencies), "prizeKey": r.PrizeKey, "prizeGroupKey": r.PrizeGroupKey}
}

type world struct {
	missions map[string]*mission
	byKey    map[string]*mission
	packs    []*pack
	packKey  map[string]*pack
}

func loadWorld(ctx context.Context, q queryer) (*world, error) {
	w := &world{missions: map[string]*mission{}, byKey: map[string]*mission{}, packKey: map[string]*pack{}}
	rows, err := q.QueryContext(ctx, `SELECT `+missionCols+missionFrom+` WHERE m.is_active ORDER BY m.sort, m.key`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		m, err := scanMission(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		w.missions[m.ID], w.byKey[m.Key] = m, m
	}
	rows.Close()
	rows, err = q.QueryContext(ctx, `SELECT `+packCols+packFrom+` WHERE k.is_active ORDER BY k.sort, k.key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		k, err := scanPack(rows)
		if err != nil {
			return nil, err
		}
		w.packs = append(w.packs, k)
		w.packKey[k.Key] = k
	}
	return w, rows.Err()
}

func (w *world) packsOf(missionID string) []*pack {
	out := []*pack{}
	for _, k := range w.packs {
		if k.MissionID == missionID {
			out = append(out, k)
		}
	}
	return out
}

func (w *world) missionsOf(gameID string) []*mission {
	out := []*mission{}
	for _, m := range w.missions {
		if m.ParentID == gameID && m.Kind == "mission" {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Sort != out[j].Sort {
			return out[i].Sort < out[j].Sort
		}
		return out[i].Key < out[j].Key
	})
	return out
}

type userPack struct {
	Plays, Wins, Rewards, LastCorrect, LastTotal int
	Best, Last                                   float64
	FirstPass, LastPlayed                        *time.Time
}

type userMission struct {
	Started            time.Time
	Completed          *time.Time
	Rewarded           bool
	hasRow             bool
	missionID, missKey string
}

type progress struct {
	packs    map[string]userPack
	missions map[string]userMission
}

func loadProgress(ctx context.Context, q queryer, userID string) (*progress, error) {
	p := &progress{packs: map[string]userPack{}, missions: map[string]userMission{}}
	rows, err := q.QueryContext(ctx, `SELECT pack_id::text, plays, wins, rewards, best_score::float8, last_score::float8, last_correct, last_total, first_pass_at, last_played_at FROM quiz_user_pack WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id string
		var u userPack
		var fp, lp sql.NullTime
		if err := rows.Scan(&id, &u.Plays, &u.Wins, &u.Rewards, &u.Best, &u.Last, &u.LastCorrect, &u.LastTotal, &fp, &lp); err != nil {
			rows.Close()
			return nil, err
		}
		if fp.Valid {
			v := fp.Time
			u.FirstPass = &v
		}
		if lp.Valid {
			v := lp.Time
			u.LastPlayed = &v
		}
		p.packs[id] = u
	}
	rows.Close()
	rows, err = q.QueryContext(ctx, `SELECT mission_id::text, started_at, completed_at, rewarded FROM quiz_user_mission WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var u userMission
		var c sql.NullTime
		if err := rows.Scan(&id, &u.Started, &c, &u.Rewarded); err != nil {
			return nil, err
		}
		if c.Valid {
			v := c.Time
			u.Completed = &v
		}
		u.hasRow = true
		p.missions[id] = u
	}
	return p, rows.Err()
}

func (p *progress) passed(packID string) bool { return p.packs[packID].FirstPass != nil }

func (p *progress) completed(missionID string) bool { return p.missions[missionID].Completed != nil }

func missionExpiry(m *mission, u userMission) *time.Time {
	if m.TimeSeconds <= 0 || !u.hasRow {
		return nil
	}
	t := u.Started.Add(time.Duration(m.TimeSeconds) * time.Second)
	return &t
}

func packLock(w *world, p *progress, k *pack, now time.Time) *apierr.Error {
	m := w.missions[k.MissionID]
	if m == nil {
		return nil
	}
	if g := w.missions[m.ParentID]; g != nil && g.Kind == "game" {
		if exp := missionExpiry(g, p.missions[g.ID]); exp != nil && now.After(*exp) && !p.completed(g.ID) {
			return apierr.Conflict("QUIZ_GAME_TIME_OVER").With("gameKey", g.Key)
		}
		if g.OrderMode == "sequential" {
			for _, c := range w.missionsOf(g.ID) {
				if c.ID == m.ID {
					break
				}
				if !p.completed(c.ID) {
					return apierr.Forbidden("QUIZ_GAME_MISSION_LOCKED").With("gameKey", g.Key).With("missionKey", c.Key)
				}
			}
		}
	}
	if exp := missionExpiry(m, p.missions[m.ID]); exp != nil && now.After(*exp) && !p.completed(m.ID) {
		return apierr.Conflict("QUIZ_MISSION_TIME_OVER").With("missionKey", m.Key)
	}
	if m.OrderMode == "sequential" {
		for _, c := range w.packsOf(m.ID) {
			if c.ID == k.ID {
				break
			}
			if !p.passed(c.ID) {
				return apierr.Forbidden("QUIZ_MISSION_PACK_LOCKED").With("missionKey", m.Key).With("packKey", c.Key)
			}
		}
	}
	return nil
}

func sendAnswers(cfg settings.All, w *world, k *pack) bool {
	if !cfg.Reveal.AllowBeforeAnswer {
		return false
	}
	if w != nil {
		if m := w.missions[k.MissionID]; m != nil {
			return m.SendAnswers
		}
	}
	return k.SendAnswers
}

func askedCount(k *pack) int {
	if k.QuestionCount > 0 && k.QuestionCount < k.Questions {
		return k.QuestionCount
	}
	return k.Questions
}

func (s *service) packView(langs *localized.Set, lang string, cfg settings.All, w *world, p *progress, k *pack, open map[string]string, now time.Time) map[string]any {
	u := p.packs[k.ID]
	asked := askedCount(k)
	var lock any
	if e := packLock(w, p, k, now); e != nil {
		lock = map[string]any{"code": e.Code, "meta": e.Meta}
	}
	return map[string]any{
		"key": k.Key, "name": text(langs, lang, k.Name), "description": text(langs, lang, k.Description), "iconUrl": k.IconURL, "sort": k.Sort,
		"category": k.Category, "tags": k.Tags, "language": k.Language, "difficulty": k.Difficulty, "missionKey": k.MissionKey,
		"timeSeconds": k.TimeSeconds, "questionSeconds": questionSeconds(nil, k, cfg), "questionCount": asked, "passCorrect": passNeed(k, cfg, asked),
		"reward": rewardView(k.Reward), "campaignStageKey": k.CampaignStageKey, "metadata": k.Metadata,
		"plays": u.Plays, "wins": u.Wins, "bestScore": u.Best, "lastScore": u.Last, "passed": u.FirstPass != nil, "rewarded": u.Rewards > 0,
		"locked": lock, "openSessionId": nilIfEmpty(open[k.ID]),
	}
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func (s *service) openSessions(ctx context.Context, userID string) map[string]string {
	out := map[string]string{}
	rows, err := s.db.QueryContext(ctx, `SELECT pack_id::text, id::text FROM quiz_session WHERE user_id=$1 AND status='open' AND expires_at > now()`, userID)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var p, id string
		if rows.Scan(&p, &id) == nil {
			out[p] = id
		}
	}
	return out
}

func (s *service) rpcPacks(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang       string `json:"lang"`
		MissionKey string `json:"missionKey"`
		Category   string `json:"category"`
		Tag        string `json:"tag"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
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
	open := s.openSessions(ctx, userID)
	now := s.now()
	out := []map[string]any{}
	for _, k := range w.packs {
		if in.MissionKey != "" && k.MissionKey != in.MissionKey || in.Category != "" && k.Category != in.Category || in.Tag != "" && !contains(k.Tags, in.Tag) {
			continue
		}
		if k.Questions == 0 {
			continue
		}
		out = append(out, s.packView(langs, lang, cfg, w, p, k, open, now))
	}
	return map[string]any{"lang": lang, "serverTime": now.UTC().Format(time.RFC3339), "packs": out}, nil
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func (s *service) rpcRandom(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang       string   `json:"lang"`
		Categories []string `json:"categories"`
		Tags       []string `json:"tags"`
		ForceNew   bool     `json:"forceNew"`
		PreferNew  bool     `json:"preferNew"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	w, err := loadWorld(ctx, s.db)
	if err != nil {
		return nil, err
	}
	p, err := loadProgress(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	var all, fresh []*pack
	for _, k := range w.packs {
		if k.Questions == 0 || packLock(w, p, k, now) != nil {
			continue
		}
		if len(in.Categories) > 0 && !contains(in.Categories, k.Category) {
			continue
		}
		if len(in.Tags) > 0 {
			hit := false
			for _, t := range in.Tags {
				hit = hit || contains(k.Tags, t)
			}
			if !hit {
				continue
			}
		}
		all = append(all, k)
		if p.packs[k.ID].Plays == 0 {
			fresh = append(fresh, k)
		}
	}
	pool := all
	if in.ForceNew || (in.PreferNew && len(fresh) > 0) {
		pool = fresh
	}
	if len(pool) == 0 {
		return nil, apierr.NotFound("QUIZ_NO_PACK_FOUND")
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	k := pool[rand.Intn(len(pool))]
	return s.packView(langs, lang, cfg, w, p, k, s.openSessions(ctx, userID), now), nil
}

type sessionRow struct {
	ID, UserID, PackID, Status, CampaignAttempt string
	Started, Expires                            time.Time
	Deadline, Served, LastAnswer                *time.Time
	Questions                                   []picked
	Current, Answered, Correct                  int
}

const sessionCols = `id::text, user_id::text, pack_id::text, status, campaign_attempt, started_at, expires_at, deadline_at, served_at, last_answer_at, questions, current, answered, correct`

func scanSession(sc scanner) (*sessionRow, error) {
	r := &sessionRow{}
	var dl, sv, la sql.NullTime
	var qs []byte
	if err := sc.Scan(&r.ID, &r.UserID, &r.PackID, &r.Status, &r.CampaignAttempt, &r.Started, &r.Expires, &dl, &sv, &la, &qs, &r.Current, &r.Answered, &r.Correct); err != nil {
		return nil, err
	}
	for _, p := range []struct {
		n   sql.NullTime
		dst **time.Time
	}{{dl, &r.Deadline}, {sv, &r.Served}, {la, &r.LastAnswer}} {
		if p.n.Valid {
			v := p.n.Time
			*p.dst = &v
		}
	}
	unmarshal(qs, &r.Questions)
	return r, nil
}

func (r *sessionRow) index(questionID string) int {
	for i, p := range r.Questions {
		if p.ID == questionID {
			return i
		}
	}
	return -1
}

type answerRow struct {
	QuestionID string
	OptionIDs  []int
	Text       string
	Outcome    string
	At         time.Time
}

func loadAnswers(ctx context.Context, q queryer, sessionID string) (map[string]answerRow, error) {
	rows, err := q.QueryContext(ctx, `SELECT question_id::text, array_to_json(option_ids)::text, text, outcome, answered_at FROM quiz_answer WHERE session_id=$1`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]answerRow{}
	for rows.Next() {
		var a answerRow
		var ids []byte
		if err := rows.Scan(&a.QuestionID, &ids, &a.Text, &a.Outcome, &a.At); err != nil {
			return nil, err
		}
		unmarshal(ids, &a.OptionIDs)
		if a.OptionIDs == nil {
			a.OptionIDs = []int{}
		}
		out[a.QuestionID] = a
	}
	return out, rows.Err()
}

func (s *service) sessionQuestions(ctx context.Context, q queryer, sess *sessionRow) (map[string]*question, error) {
	ids := make([]string, 0, len(sess.Questions))
	for _, p := range sess.Questions {
		ids = append(ids, p.ID)
	}
	list, err := listQuestions(ctx, q, `q.id::text = ANY($1::text[])`, ids)
	if err != nil {
		return nil, err
	}
	out := map[string]*question{}
	for _, x := range list {
		out[x.ID] = x
	}
	return out, nil
}

func reveal(langs *localized.Set, lang string, cfg settings.All, q *question) map[string]any {
	out := map[string]any{}
	if q.Type == "options" {
		out["correctOptionIds"] = correctIDs(q)
	} else {
		list := q.Answers[lang]
		if len(list) == 0 {
			for _, v := range q.Answers {
				list = v
				break
			}
		}
		out["acceptedAnswers"] = list
	}
	if cfg.Reveal.Explanation {
		out["explanation"] = text(langs, lang, q.Explanation)
	}
	return out
}

func clockStart(cfg settings.All, sess *sessionRow) time.Time {
	if cfg.Play.Delivery == settings.DeliveryOneByOne && sess.Served != nil {
		return *sess.Served
	}
	if sess.LastAnswer != nil {
		return *sess.LastAnswer
	}
	return sess.Started
}

func (s *service) view(ctx context.Context, userID string, sess *sessionRow, k *pack, w *world, cfg settings.All, langs *localized.Set, lang string) (map[string]any, error) {
	qs, err := s.sessionQuestions(ctx, s.db, sess)
	if err != nil {
		return nil, err
	}
	answers, err := loadAnswers(ctx, s.db, sess.ID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(qs))
	for id := range qs {
		ids = append(ids, id)
	}
	hints, err := listHints(ctx, s.db, ids, true)
	if err != nil {
		return nil, err
	}
	owned := s.ownedHints(ctx, userID, sess.ID, cfg)
	send := sendAnswers(cfg, w, k)
	oneByOne := cfg.Play.Delivery == settings.DeliveryOneByOne
	list := []map[string]any{}
	for i, p := range sess.Questions {
		q := qs[p.ID]
		if q == nil {
			continue
		}
		a, done := answers[q.ID]
		if oneByOne && !done && i != sess.Current {
			continue
		}
		list = append(list, s.questionView(langs, lang, cfg, k, sess, q, p, i, a, done, send, hints[q.ID], owned))
	}
	return map[string]any{
		"sessionId": sess.ID, "packKey": k.Key, "status": sess.Status, "startedAt": sess.Started.UTC().Format(time.RFC3339Nano), "deadlineAt": iso(sess.Deadline),
		"expiresAt": sess.Expires.UTC().Format(time.RFC3339), "serverTime": s.now().UTC().Format(time.RFC3339Nano), "delivery": cfg.Play.Delivery,
		"total": len(sess.Questions), "answered": sess.Answered, "correct": sess.Correct, "current": sess.Current, "passCorrect": passNeed(k, cfg, len(sess.Questions)),
		"questions": list, "campaignAttemptId": nilIfEmpty(sess.CampaignAttempt),
	}, nil
}

func (s *service) questionView(langs *localized.Set, lang string, cfg settings.All, k *pack, sess *sessionRow, q *question, p picked, i int, a answerRow, done, send bool, hints []*hint, owned map[string]bool) map[string]any {
	sec := questionSeconds(q, k, cfg)
	v := map[string]any{"id": q.ID, "index": i, "type": q.Type, "text": text(langs, lang, q.Question), "mediaUrl": q.MediaURL, "difficulty": q.Difficulty,
		"timeSeconds": sec, "answered": nil}
	if !done && (cfg.Play.Delivery == settings.DeliveryOneByOne || i == sess.Current) {
		v["deadlineAt"] = iso(questionDeadline(sec, clockStart(cfg, sess)))
	}
	if q.Type == "options" {
		byID := map[int]Option{}
		for _, o := range q.Options {
			byID[o.ID] = o
		}
		order := p.Order
		if len(order) == 0 {
			for _, o := range q.Options {
				order = append(order, o.ID)
			}
		}
		opts := []map[string]any{}
		for _, id := range order {
			o, ok := byID[id]
			if !ok {
				continue
			}
			ov := map[string]any{"id": o.ID, "text": text(langs, lang, o.Text), "mediaUrl": o.MediaURL}
			if send {
				ov["correct"] = o.Correct
			}
			opts = append(opts, ov)
		}
		v["options"] = opts
	} else if send {
		v["acceptedAnswers"] = q.Answers
	}
	if done {
		av := map[string]any{"outcome": a.Outcome, "optionIds": a.OptionIDs, "text": a.Text}
		if cfg.Reveal.Mode == settings.RevealAnswer || send {
			for k2, x := range reveal(langs, lang, cfg, q) {
				av[k2] = x
			}
		}
		v["answered"] = av
	}
	hl := []map[string]any{}
	for _, h := range hints {
		own := owned[h.ID]
		hv := map[string]any{"id": h.ID, "name": text(langs, lang, h.Name), "price": hintPrice(cfg, h), "owned": own}
		if own {
			hv["text"], hv["mediaUrl"] = text(langs, lang, h.Text), h.MediaURL
		}
		hl = append(hl, hv)
	}
	v["hints"] = hl
	return v
}

func (s *service) ownedHints(ctx context.Context, userID, sessionID string, cfg settings.All) map[string]bool {
	out := map[string]bool{}
	scope := ""
	if cfg.Hints.Keep == settings.KeepSession {
		scope = sessionID
	}
	rows, err := s.db.QueryContext(ctx, `SELECT hint_id::text FROM quiz_user_hint WHERE user_id=$1 AND scope=$2 AND status='done'`, userID, scope)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			out[id] = true
		}
	}
	return out
}

func lockDay(ctx context.Context, tx *sql.Tx, userID, day string) (plays, rewarded, reports int, err error) {
	if _, err = tx.ExecContext(ctx, `INSERT INTO quiz_user_day (user_id, day) VALUES ($1,$2) ON CONFLICT DO NOTHING`, userID, day); err != nil {
		return
	}
	err = tx.QueryRowContext(ctx, `SELECT plays, rewarded_passes, reports FROM quiz_user_day WHERE user_id=$1 AND day=$2 FOR UPDATE`, userID, day).Scan(&plays, &rewarded, &reports)
	return
}

func (s *service) rpcPlay(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		PackKey  string         `json:"packKey"`
		Lang     string         `json:"lang"`
		Metadata map[string]any `json:"metadata"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	in.PackKey = strings.TrimSpace(in.PackKey)
	if in.PackKey == "" {
		return nil, apierr.BadRequest("QUIZ_PACK_KEY_REQUIRED").WithField("packKey")
	}
	if raw, _ := json.Marshal(in.Metadata); len(raw) > 4096 {
		return nil, apierr.BadRequest("QUIZ_METADATA_TOO_LARGE").WithField("metadata")
	}
	return s.play(ctx, userID, in.PackKey, in.Lang, in.Metadata)
}

func (s *service) play(ctx context.Context, userID, packKey, reqLang string, metadata map[string]any) (any, error) {
	cfg := s.settings(ctx)
	w, err := loadWorld(ctx, s.db)
	if err != nil {
		return nil, err
	}
	k := w.packKey[packKey]
	if k == nil {
		return nil, apierr.NotFound("QUIZ_PACK_NOT_FOUND").WithField("packKey")
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, reqLang)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC()

	if old, err := scanSession(s.db.QueryRowContext(ctx, `SELECT `+sessionCols+` FROM quiz_session WHERE user_id=$1 AND pack_id=$2 AND status='open'`, userID, k.ID)); err == nil {
		switch {
		case now.After(old.Expires) || old.Deadline != nil && now.After(old.Deadline.Add(time.Duration(cfg.Timing.GraceSeconds)*time.Second)):
			if _, err := s.finish(ctx, old.ID, true); err != nil {
				return nil, err
			}
		case cfg.Play.OnRestart == settings.RestartResume:
			out, err := s.view(ctx, userID, old, k, w, cfg, langs, lang)
			if err == nil {
				out["resumed"] = true
			}
			return out, err
		default:
			if _, err := s.db.ExecContext(ctx, `UPDATE quiz_session SET status='closed', completed_at=$2 WHERE id=$1 AND status='open'`, old.ID, now); err != nil {
				return nil, err
			}
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	p, err := loadProgress(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}
	if e := packLock(w, p, k, now); e != nil {
		return nil, e
	}
	u := p.packs[k.ID]
	if !cfg.Limits.ReplayAfterPass && u.FirstPass != nil {
		return nil, apierr.Conflict("QUIZ_PACK_ALREADY_PASSED").WithField("packKey")
	}
	if cfg.Limits.MaxPlaysPerPack > 0 && u.Plays >= cfg.Limits.MaxPlaysPerPack {
		return nil, apierr.Conflict("QUIZ_PACK_PLAY_LIMIT").With("limit", cfg.Limits.MaxPlaysPerPack)
	}
	if cfg.Limits.RetryCooldownSeconds > 0 && u.LastPlayed != nil {
		if next := u.LastPlayed.Add(time.Duration(cfg.Limits.RetryCooldownSeconds) * time.Second); now.Before(next) {
			return nil, apierr.TooManyRequests("QUIZ_PACK_COOLDOWN", int(next.Sub(now).Seconds())+1).With("nextPlayAt", next.Format(time.RFC3339))
		}
	}
	pool, err := poolOf(ctx, s.db, k)
	if err != nil {
		return nil, err
	}
	if len(pool) == 0 {
		return nil, apierr.Conflict("QUIZ_PACK_EMPTY").WithField("packKey")
	}
	shuffleQ, shuffleO := cfg.Play.ShuffleQuestions, cfg.Play.ShuffleOptions
	if k.ShuffleQuestions != nil {
		shuffleQ = *k.ShuffleQuestions
	}
	if k.ShuffleOptions != nil {
		shuffleO = *k.ShuffleOptions
	}
	rng := rand.New(rand.NewSource(now.UnixNano()))
	picks := pickQuestions(pool, k.QuestionCount, shuffleQ, shuffleO, rng)
	attempt := ""
	if k.CampaignStageKey != "" && cfg.Campaign.Enabled && s.campaignOn() {
		if attempt, err = s.campaignStart(ctx, userID, k.CampaignStageKey); err != nil {
			return nil, err
		}
		now = s.now().UTC()
	}
	id := uuid.Must(uuid.NewV4()).String()
	var deadline *time.Time
	expires := now.Add(time.Duration(cfg.Play.SessionHours) * time.Hour)
	if k.TimeSeconds > 0 {
		d := now.Add(time.Duration(k.TimeSeconds) * time.Second)
		deadline = &d
		if e := d.Add(time.Duration(cfg.Timing.GraceSeconds+60) * time.Second); e.Before(expires) {
			expires = e
		}
	}
	day := dayKey(now, s.zone(cfg))
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		plays, _, _, err := lockDay(ctx, tx, userID, day)
		if err != nil {
			return err
		}
		if cfg.Limits.DailyPlays > 0 && plays >= cfg.Limits.DailyPlays {
			return apierr.TooManyRequests("QUIZ_DAILY_PLAY_LIMIT", secondsToNextDay(now, s.zone(cfg))).With("limit", cfg.Limits.DailyPlays)
		}
		if m := w.missions[k.MissionID]; m != nil {
			for _, x := range []*mission{w.missions[m.ParentID], m} {
				if x == nil {
					continue
				}
				if _, err := tx.ExecContext(ctx, `INSERT INTO quiz_user_mission (user_id, mission_id, started_at) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, userID, x.ID, now); err != nil {
					return err
				}
			}
		}
		md := metadata
		if md == nil {
			md = map[string]any{}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO quiz_session (id, user_id, pack_id, started_at, expires_at, deadline_at, questions, served_at, campaign_attempt, metadata) VALUES ($1,$2,$3,$4,$5,$6,$7,$4,$8,$9)`,
			id, userID, k.ID, now, expires, deadline, jsonOf(picks), attempt, jsonOf(md)); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO quiz_user_pack (user_id, pack_id, plays, last_played_at) VALUES ($1,$2,1,$3)
			ON CONFLICT (user_id, pack_id) DO UPDATE SET plays=quiz_user_pack.plays+1, last_played_at=$3`, userID, k.ID, now); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `UPDATE quiz_user_day SET plays=plays+1 WHERE user_id=$1 AND day=$2`, userID, day)
		return err
	})
	if err != nil {
		if isUnique(err) {

			return s.play(ctx, userID, packKey, reqLang, metadata)
		}
		return nil, err
	}
	sess, err := scanSession(s.db.QueryRowContext(ctx, `SELECT `+sessionCols+` FROM quiz_session WHERE id=$1`, id))
	if err != nil {
		return nil, err
	}
	out, err := s.view(ctx, userID, sess, k, w, cfg, langs, lang)
	if err == nil {
		out["resumed"] = false
	}
	return out, err
}

func secondsToNextDay(now time.Time, loc *time.Location) int {
	t := now.In(loc)
	next := time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, loc)
	return int(next.Sub(t).Seconds()) + 1
}

func (s *service) ownSession(ctx context.Context, userID, sessionID string) (*sessionRow, *pack, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, nil, apierr.BadRequest("QUIZ_SESSION_REQUIRED").WithField("sessionId")
	}
	if _, err := uuid.FromString(sessionID); err != nil {
		return nil, nil, apierr.NotFound("QUIZ_SESSION_NOT_FOUND").WithField("sessionId")
	}
	sess, err := scanSession(s.db.QueryRowContext(ctx, `SELECT `+sessionCols+` FROM quiz_session WHERE id=$1 AND user_id=$2`, sessionID, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, apierr.NotFound("QUIZ_SESSION_NOT_FOUND").WithField("sessionId")
	}
	if err != nil {
		return nil, nil, err
	}
	k, err := loadPack(ctx, s.db, `k.id=$1`, sess.PackID)
	if err != nil {
		return nil, nil, err
	}
	return sess, k, nil
}

func (s *service) rpcSession(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		SessionID string `json:"sessionId"`
		PackKey   string `json:"packKey"`
		Lang      string `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	if in.SessionID == "" && in.PackKey != "" {
		_ = s.db.QueryRowContext(ctx, `SELECT s.id::text FROM quiz_session s JOIN quiz_pack k ON k.id=s.pack_id WHERE s.user_id=$1 AND k.key=$2 ORDER BY (s.status='open') DESC, s.started_at DESC LIMIT 1`, userID, strings.TrimSpace(in.PackKey)).Scan(&in.SessionID)
		if in.SessionID == "" {
			return nil, apierr.NotFound("QUIZ_SESSION_NOT_FOUND").WithField("packKey")
		}
	}
	sess, k, err := s.ownSession(ctx, userID, in.SessionID)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	w, err := loadWorld(ctx, s.db)
	if err != nil {
		return nil, err
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, err
	}
	out, err := s.view(ctx, userID, sess, k, w, cfg, langs, lang)
	if err != nil {
		return nil, err
	}
	if sess.Status != "open" {
		var raw []byte
		if s.db.QueryRowContext(ctx, `SELECT result FROM quiz_session WHERE id=$1`, sess.ID).Scan(&raw) == nil && raw != nil {
			res := map[string]any{}
			unmarshal(raw, &res)
			out["result"] = res
		}
	}
	return out, nil
}

type answerIn struct {
	SessionID  string `json:"sessionId"`
	QuestionID string `json:"questionId"`
	OptionID   *int   `json:"optionId"`
	OptionIDs  []int  `json:"optionIds"`
	Text       string `json:"text"`
	Lang       string `json:"lang"`
}

func (s *service) rpcAnswer(ctx context.Context, userID, payload string) (any, error) {
	var in answerIn
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	out, finished, err := s.answer(ctx, userID, in, cfg, cfg.Play.AutoFinish)
	if err != nil {
		return nil, err
	}
	if finished {
		res, err := s.finish(ctx, in.SessionID, false)
		if err != nil {
			return nil, err
		}
		out["result"] = res
	}
	return out, nil
}

func (s *service) rpcAnswerAll(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		SessionID string     `json:"sessionId"`
		Answers   []answerIn `json:"answers"`
		Lang      string     `json:"lang"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	if len(in.Answers) > 200 {
		return nil, apierr.BadRequest("QUIZ_TOO_MANY_ANSWERS").WithField("answers")
	}
	sess, _, err := s.ownSession(ctx, userID, in.SessionID)
	if err != nil {
		return nil, err
	}
	cfg := s.settings(ctx)
	list := []any{}
	for _, a := range in.Answers {
		a.SessionID, a.Lang = sess.ID, in.Lang
		out, _, err := s.answer(ctx, userID, a, cfg, false)
		if err != nil {
			if e, ok := apierr.As(err); ok && e.Status < 500 {
				list = append(list, map[string]any{"questionId": a.QuestionID, "error": e.Code})
				if e.Code == "QUIZ_TIME_OVER" || e.Code == "QUIZ_SESSION_CLOSED" {
					break
				}
				continue
			}
			return nil, err
		}
		list = append(list, out)
	}
	res, err := s.finish(ctx, sess.ID, false)
	if err != nil {
		return nil, err
	}
	return map[string]any{"sessionId": sess.ID, "answers": list, "result": res}, nil
}

func (s *service) answer(ctx context.Context, userID string, in answerIn, cfg settings.All, autoFinish bool) (map[string]any, bool, error) {
	sess, k, err := s.ownSession(ctx, userID, in.SessionID)
	if err != nil {
		return nil, false, err
	}
	in.QuestionID = strings.TrimSpace(in.QuestionID)
	if sess.index(in.QuestionID) < 0 {
		return nil, false, apierr.NotFound("QUIZ_QUESTION_NOT_IN_SESSION").WithField("questionId")
	}
	chosen := append([]int{}, in.OptionIDs...)
	if in.OptionID != nil {
		chosen = append(chosen, *in.OptionID)
	}
	if len(chosen) > 12 || len(in.Text) > 500 {
		return nil, false, apierr.BadRequest("QUIZ_ANSWER_INVALID").WithField("optionIds")
	}
	qs, err := s.sessionQuestions(ctx, s.db, sess)
	if err != nil {
		return nil, false, err
	}
	q := qs[in.QuestionID]
	if q == nil {
		return nil, false, apierr.NotFound("QUIZ_QUESTION_NOT_IN_SESSION").WithField("questionId")
	}
	if q.Type == "options" {
		known := map[int]bool{}
		for _, o := range q.Options {
			known[o.ID] = true
		}
		for _, id := range chosen {
			if !known[id] {
				return nil, false, apierr.BadRequest("QUIZ_OPTION_NOT_FOUND").WithField("optionId").With("optionId", id)
			}
		}
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, in.Lang)
	if err != nil {
		return nil, false, err
	}
	var outcome string
	replayed, late, closed, overtime := false, false, false, false
	var nextIdx int
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		replayed, late, closed, overtime = false, false, false, false
		cur, err := scanSession(tx.QueryRowContext(ctx, `SELECT `+sessionCols+` FROM quiz_session WHERE id=$1 FOR UPDATE`, sess.ID))
		if err != nil {
			return err
		}
		sess = cur
		var stored string
		err = tx.QueryRowContext(ctx, `SELECT outcome FROM quiz_answer WHERE session_id=$1 AND question_id=$2`, sess.ID, q.ID).Scan(&stored)
		if err == nil {
			outcome, replayed = stored, true
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if sess.Status != "open" {
			closed = true
			return nil
		}
		now := s.now().UTC()
		grace := time.Duration(cfg.Timing.GraceSeconds) * time.Second
		if now.After(sess.Expires) || sess.Deadline != nil && now.After(sess.Deadline.Add(grace)) {
			overtime = true
			return nil
		}
		idx := sess.index(q.ID)
		if cfg.Play.Delivery == settings.DeliveryOneByOne && idx != sess.Current {
			return apierr.Conflict("QUIZ_QUESTION_NOT_CURRENT").WithField("questionId").With("currentQuestionId", sess.Questions[minInt(sess.Current, len(sess.Questions)-1)].ID)
		}
		if dl := questionDeadline(questionSeconds(q, k, cfg), clockStart(cfg, sess)); dl != nil && now.After(dl.Add(grace)) {
			late = true
		}
		skipped := len(chosen) == 0 && strings.TrimSpace(in.Text) == ""
		switch {
		case late:
			outcome = "late"
		case skipped:
			outcome = "skip"
		case q.Type == "options" && checkOptions(q, chosen, cfg.Scoring.MultiCorrect):
			outcome = "correct"
		case q.Type == "text" && checkText(q, in.Text, cfg.Scoring.TextMatch):
			outcome = "correct"
		default:
			outcome = "wrong"
		}
		ids := make([]int64, len(chosen))
		for i, v := range chosen {
			ids[i] = int64(v)
		}
		seconds := now.Sub(clockStart(cfg, sess)).Seconds()
		if _, err := tx.ExecContext(ctx, `INSERT INTO quiz_answer (session_id, question_id, user_id, option_ids, text, outcome, seconds, answered_at) VALUES ($1,$2,$3,$4::int[],$5,$6,$7,$8)`,
			sess.ID, q.ID, userID, ids, strings.TrimSpace(in.Text), outcome, seconds, now); err != nil {
			return err
		}
		answered := map[string]bool{q.ID: true}
		rows, err := tx.QueryContext(ctx, `SELECT question_id::text FROM quiz_answer WHERE session_id=$1`, sess.ID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id string
			if rows.Scan(&id) == nil {
				answered[id] = true
			}
		}
		rows.Close()
		nextIdx = len(sess.Questions)
		for i, p := range sess.Questions {
			if !answered[p.ID] {
				nextIdx = i
				break
			}
		}
		add := 0
		if outcome == "correct" {
			add = 1
		}
		_, err = tx.ExecContext(ctx, `UPDATE quiz_session SET answered=answered+1, correct=correct+$2, last_answer_at=$3, current=$4, served_at=$3 WHERE id=$1`, sess.ID, add, now, nextIdx)
		sess.Answered++
		sess.Correct += add
		sess.Current = nextIdx
		sess.LastAnswer, sess.Served = &now, &now
		return err
	})
	if err != nil {
		return nil, false, err
	}
	if overtime {
		if _, err := s.finish(ctx, sess.ID, true); err != nil {
			return nil, false, err
		}
		return nil, false, apierr.Conflict("QUIZ_TIME_OVER").With("sessionId", sess.ID)
	}
	if closed {
		return nil, false, apierr.Conflict("QUIZ_SESSION_CLOSED").With("sessionId", sess.ID)
	}
	if late && !replayed && cfg.Timing.LateAnswer == settings.LateReject {
		return nil, false, apierr.Conflict("QUIZ_QUESTION_TIME_OVER").WithField("questionId")
	}
	out := map[string]any{"sessionId": sess.ID, "questionId": q.ID, "outcome": outcome, "correct": outcome == "correct", "replayed": replayed,
		"answered": sess.Answered, "correctCount": sess.Correct, "total": len(sess.Questions)}
	if cfg.Reveal.Mode == settings.RevealAnswer {
		for key, v := range reveal(langs, lang, cfg, q) {
			out[key] = v
		}
	}
	if !replayed && cfg.Play.Delivery == settings.DeliveryOneByOne && nextIdx < len(sess.Questions) {
		w, err := loadWorld(ctx, s.db)
		if err != nil {
			return nil, false, err
		}
		nq := qs[sess.Questions[nextIdx].ID]
		if nq != nil {
			hints, _ := listHints(ctx, s.db, []string{nq.ID}, true)
			out["next"] = s.questionView(langs, lang, cfg, k, sess, nq, sess.Questions[nextIdx], nextIdx, answerRow{}, false, sendAnswers(cfg, w, k), hints[nq.ID], s.ownedHints(ctx, userID, sess.ID, cfg))
		}
	}
	return out, autoFinish && !replayed && sess.Answered >= len(sess.Questions), nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	if b < 0 {
		return 0
	}
	return b
}

func (s *service) rpcFinish(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		SessionID string `json:"sessionId"`
	}
	if err := apierr.Decode(payload, &in); err != nil {
		return nil, err
	}
	sess, _, err := s.ownSession(ctx, userID, in.SessionID)
	if err != nil {
		return nil, err
	}
	if sess.Status == "closed" {
		return nil, apierr.Conflict("QUIZ_SESSION_CLOSED").With("sessionId", sess.ID)
	}
	return s.finish(ctx, sess.ID, false)
}
