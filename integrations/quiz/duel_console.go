package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

var colorPattern = regexp.MustCompile(`^(#[0-9A-Fa-f]{6}|)$`)

func (s *service) duelRoutes() []struct {
	path, method string
	handler      http.HandlerFunc
} {
	p := consolePrefix
	return []struct {
		path, method string
		handler      http.HandlerFunc
	}{
		{p + "/duel-categories", http.MethodGet, s.duelCategoryList},
		{p + "/duel-categories", http.MethodPost, s.duelCategorySave},
		{p + "/duel-categories/{id}", http.MethodPut, s.duelCategorySave},
		{p + "/duel-categories/{id}", http.MethodDelete, s.duelCategoryDelete},
		{p + "/duels", http.MethodGet, s.duelList},
		{p + "/duels/{id}", http.MethodGet, s.duelGet},
		{p + "/duels/{id}/cancel", http.MethodPost, s.duelCancel},
		{p + "/duel-stats", http.MethodGet, s.duelStats},
	}
}

type categoryBody struct {
	Key         string            `json:"key"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	IconURL     string            `json:"iconUrl"`
	Color       string            `json:"color"`
	Sort        int               `json:"sort"`
	Active      *bool             `json:"isActive"`
	Duel        *bool             `json:"duel"`
}

func (s *service) duelCategoryList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	args := &pluginkit.Args{}
	where := `true`
	if v := strings.TrimSpace(q.Get("q")); v != "" {
		ph := args.Add(pluginkit.LikePattern(v))
		where += ` AND (c.key ILIKE ` + ph + ` OR c.name::text ILIKE ` + ph + `)`
	}
	if v := q.Get("active"); v == "true" || v == "false" {
		where += ` AND c.is_active=` + args.Add(v == "true")
	}
	page, limit := pluginkit.Page(r, 50, 200)
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM quiz_category c WHERE `+where, args.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	cfg := s.settings(ctx)
	rules := rulesOf(cfg.Duel)
	lim, off := args.Add(limit), args.Add((page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT c.id::text, c.key, c.name, c.description, c.icon_url, c.color, c.sort, c.is_active, c.duel, c.updated_by, c.updated_at,
		(SELECT count(*) FROM quiz_question q WHERE q.category=c.key AND q.is_active),
		(SELECT count(*) FROM quiz_question q WHERE q.category=c.key AND `+s.eligibleSQL(rules)+`),
		(SELECT count(*) FROM quiz_duel_round d WHERE d.category=c.key)
		FROM quiz_category c WHERE `+where+` ORDER BY c.sort, c.key LIMIT `+lim+` OFFSET `+off, args.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, key, icon, color, by string
		var name, desc []byte
		var sort, questions, eligible, rounds int
		var active, duel bool
		var at sql.NullTime
		if err := rows.Scan(&id, &key, &name, &desc, &icon, &color, &sort, &active, &duel, &by, &at, &questions, &eligible, &rounds); err != nil {
			s.fail(w, err)
			return
		}
		n, dd := map[string]string{}, map[string]string{}
		unmarshal(name, &n)
		unmarshal(desc, &dd)
		items = append(items, map[string]any{"id": id, "key": key, "name": n, "description": dd, "iconUrl": icon, "color": color, "sort": sort, "isActive": active, "duel": duel,
			"updatedBy": by, "updatedAt": stamp(at), "questions": questions, "eligible": eligible, "playable": eligible >= rules.PerRound, "rounds": rounds})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit, "perRound": rules.PerRound})
}

func (s *service) duelCategorySave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	create := id == ""
	var in categoryBody
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if !create {
		if _, err := uuid.FromString(id); err != nil {
			s.fail(w, pluginkit.Missing("NOT_FOUND"))
			return
		}
	}
	in.Key = strings.TrimSpace(in.Key)
	if create && !settings.KeyPattern.MatchString(in.Key) {
		s.fail(w, bad("INVALID_KEY", "key"))
		return
	}
	name, err := s.cleanText(ctx, "name", in.Name, true, 80)
	if err != nil {
		s.fail(w, err)
		return
	}
	desc, err := s.cleanText(ctx, "description", in.Description, false, 500)
	if err != nil {
		s.fail(w, err)
		return
	}
	icon, err := cleanURL("iconUrl", in.IconURL)
	if err != nil {
		s.fail(w, err)
		return
	}
	in.Color = strings.TrimSpace(in.Color)
	if !colorPattern.MatchString(in.Color) {
		s.fail(w, bad("INVALID_COLOR", "color"))
		return
	}
	if in.Sort < -100000 || in.Sort > 100000 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "sort"))
		return
	}
	active, duel := true, true
	if in.Active != nil {
		active = *in.Active
	}
	if in.Duel != nil {
		duel = *in.Duel
	}
	by := pluginkit.ConsoleUser(r)
	if create {
		id = uuid.Must(uuid.NewV4()).String()
		_, err = s.db.ExecContext(ctx, `INSERT INTO quiz_category (id, key, name, description, icon_url, color, sort, is_active, duel, updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			id, in.Key, jsonOf(name), jsonOf(desc), icon, in.Color, in.Sort, active, duel, by)
		if isUnique(err) {
			s.fail(w, conflict("KEY_EXISTS", "key"))
			return
		}
	} else {
		var res sql.Result
		res, err = s.db.ExecContext(ctx, `UPDATE quiz_category SET name=$2, description=$3, icon_url=$4, color=$5, sort=$6, is_active=$7, duel=$8, updated_by=$9, updated_at=now() WHERE id=$1`,
			id, jsonOf(name), jsonOf(desc), icon, in.Color, in.Sort, active, duel, by)
		if err == nil {
			if n, _ := res.RowsAffected(); n == 0 {
				s.fail(w, pluginkit.Missing("NOT_FOUND"))
				return
			}
		}
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	status := http.StatusOK
	if create {
		status = http.StatusCreated
	}
	pluginkit.WriteJSON(w, status, map[string]any{"id": id})
}

func (s *service) duelCategoryDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	var key string
	if err := s.db.QueryRowContext(ctx, `SELECT key FROM quiz_category WHERE id=$1`, id).Scan(&key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.fail(w, pluginkit.Missing("NOT_FOUND"))
			return
		}
		s.fail(w, err)
		return
	}
	if r.URL.Query().Get("force") != "true" {
		var used bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM quiz_question WHERE category=$1) OR EXISTS (SELECT 1 FROM quiz_duel_round WHERE category=$1)`, key).Scan(&used); err != nil {
			s.fail(w, err)
			return
		}
		if used {
			s.fail(w, conflict("IN_USE", ""))
			return
		}
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM quiz_category WHERE id=$1`, id); err != nil {
		s.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *service) duelSummary(ctx context.Context, d *duelRow, ps map[string]playerInfo) map[string]any {
	pv := func(id string, score int) any {
		if id == "" {
			return nil
		}
		p := ps[id]
		return map[string]any{"userId": id, "username": p.Username, "displayName": p.DisplayName, "score": score}
	}
	return map[string]any{"id": d.ID, "mode": d.Mode, "status": d.Status, "playerA": pv(d.A, d.ScoreA), "playerB": pv(d.B, d.ScoreB), "turnUser": d.Turn,
		"turnDeadline": iso(d.TurnDeadline), "expiresAt": iso(d.Expires), "winner": d.Winner, "reason": d.Reason, "feeA": d.FeeA, "feeB": d.FeeB,
		"paidA": d.PaidA, "paidB": d.PaidB, "payError": d.PayError, "rules": d.Rules, "createdAt": d.Created.UTC().Format(time.RFC3339),
		"acceptedAt": iso(d.Accepted), "finishedAt": iso(d.Finished), "updatedAt": d.Updated.UTC().Format(time.RFC3339)}
}

func (s *service) duelList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	args := &pluginkit.Args{}
	where := `true`
	switch v := q.Get("status"); v {
	case "", "all":
	case "open":
		where += ` AND d.status IN ('waiting','invited','active')`
	case "waiting", "invited", "active", "finished", "expired", "declined", "cancelled":
		where += ` AND d.status=` + args.Add(v)
	default:
		s.fail(w, bad("INVALID_STATUS", "status"))
		return
	}
	if v := q.Get("mode"); v == "random" || v == "friend" {
		where += ` AND d.mode=` + args.Add(v)
	}
	if v := strings.TrimSpace(q.Get("player")); v != "" {
		if _, err := uuid.FromString(v); err == nil {
			ph := args.Add(v)
			where += ` AND (d.player_a=` + ph + `::uuid OR d.player_b=` + ph + `::uuid)`
		} else {
			ph := args.Add(pluginkit.LikePattern(v))
			where += ` AND EXISTS (SELECT 1 FROM users u WHERE (u.id=d.player_a OR u.id=d.player_b) AND (u.username ILIKE ` + ph + ` OR u.display_name ILIKE ` + ph + `))`
		}
	}
	page, limit := pluginkit.Page(r, 20, 100)
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM quiz_duel d WHERE `+where, args.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	lim, off := args.Add(limit), args.Add((page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT `+duelCols+` FROM quiz_duel d WHERE `+where+` ORDER BY d.updated_at DESC, d.id LIMIT `+lim+` OFFSET `+off, args.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	var list []*duelRow
	ids := []string{}
	for rows.Next() {
		d, err := scanDuel(rows)
		if err != nil {
			rows.Close()
			s.fail(w, err)
			return
		}
		list = append(list, d)
		ids = append(ids, d.A, d.B)
	}
	rows.Close()
	ps := s.players(ctx, ids...)
	items := []map[string]any{}
	for _, d := range list {
		items = append(items, s.duelSummary(ctx, d, ps))
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) duelGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	d, err := loadDuel(ctx, s.db, pluginkit.PathVar(r, "id"), false)
	if err != nil {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	st, err := loadState(ctx, s.db, d)
	if err != nil {
		s.fail(w, err)
		return
	}
	langs, lang, _ := localized.ForUser(ctx, s.db, "", r.URL.Query().Get("lang"))
	qids := []string{}
	keys := []string{}
	for _, rr := range st.rounds {
		keys = append(keys, rr.Category)
		keys = append(keys, rr.Offers...)
		for _, p := range rr.Questions {
			qids = append(qids, p.ID)
		}
	}
	qs, err := listQuestions(ctx, s.db, `q.id::text = ANY($1::text[])`, qids)
	if err != nil {
		s.fail(w, err)
		return
	}
	byID := map[string]*question{}
	for _, q := range qs {
		byID[q.ID] = q
	}
	cats := s.categoryMap(ctx, keys)
	ans := func(user, qid string) any {
		a := st.answer(user, qid)
		if a == nil {
			return nil
		}
		return map[string]any{"outcome": a.Outcome, "optionId": a.Option, "servedAt": a.Served.UTC().Format(time.RFC3339Nano), "answeredAt": iso(a.Answered), "seconds": a.Seconds}
	}
	rounds := []any{}
	for i := 1; i <= d.Rules.Rounds; i++ {
		rr := st.rounds[i]
		if rr == nil {
			rounds = append(rounds, map[string]any{"round": i, "picker": d.picker(i), "offers": []any{}, "category": nil, "questions": []any{}})
			continue
		}
		offers := []any{}
		for _, k := range rr.Offers {
			offers = append(offers, categoryView(langs, lang, cats, k))
		}
		var cat any
		if rr.Category != "" {
			cat = categoryView(langs, lang, cats, rr.Category)
		}
		list := []any{}
		for _, p := range rr.Questions {
			row := map[string]any{"questionId": p.ID, "text": "", "options": []any{}, "a": ans(d.A, p.ID), "b": ans(d.B, p.ID)}
			if q := byID[p.ID]; q != nil {
				row["text"] = text(langs, lang, q.Question)
				opts := []any{}
				for _, o := range q.Options {
					opts = append(opts, map[string]any{"id": o.ID, "text": text(langs, lang, o.Text), "correct": o.Correct})
				}
				row["options"] = opts
			}
			list = append(list, row)
		}
		rounds = append(rounds, map[string]any{"round": i, "picker": rr.Picker, "offers": offers, "category": cat, "pickedAt": iso(rr.PickedAt), "questions": list})
	}
	out := s.duelSummary(ctx, d, s.players(ctx, d.A, d.B))
	out["rounds"] = rounds
	out["stepA"], out["stepB"] = st.step(d.A).Kind, st.step(d.B).Kind
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) duelCancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	if _, err := uuid.FromString(id); err != nil {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	_, err := s.mutate(ctx, id, func(tx *sql.Tx, st *duelState, x *duelTx) error {
		if !st.d.open() {
			return conflict("QUIZ_DUEL_OVER", "")
		}
		s.endWithout(st.d, "cancelled", "admin", x, s.now())
		x.note(st.d.A, "cancelled", "", true)
		x.note(st.d.B, "cancelled", "", true)
		return nil
	})
	if err != nil {
		var ce *pluginkit.ConsoleError
		if errors.As(err, &ce) {
			s.fail(w, err)
			return
		}
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	s.duelGet(w, r)
}

func (s *service) duelStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days < 1 || days > 365 {
		days = 30
	}
	var started, finished, open, waiting, timeouts, forfeits, draws, players int
	var avgMinutes float64
	if err := s.db.QueryRowContext(ctx, `SELECT count(*), count(*) FILTER (WHERE status='finished'), count(*) FILTER (WHERE status IN ('invited','active')), count(*) FILTER (WHERE status='waiting'),
		count(*) FILTER (WHERE reason='timeout'), count(*) FILTER (WHERE reason='forfeit'), count(*) FILTER (WHERE status='finished' AND winner IS NULL),
		COALESCE(AVG(EXTRACT(EPOCH FROM (finished_at - created_at))/60) FILTER (WHERE status='finished'),0)::float8
		FROM quiz_duel WHERE created_at > now() - make_interval(days => $1)`, days).Scan(&started, &finished, &open, &waiting, &timeouts, &forfeits, &draws, &avgMinutes); err != nil {
		s.fail(w, err)
		return
	}
	_ = s.db.QueryRowContext(ctx, `SELECT count(DISTINCT u) FROM (SELECT player_a AS u FROM quiz_duel WHERE created_at > now() - make_interval(days => $1) UNION SELECT player_b FROM quiz_duel WHERE player_b IS NOT NULL AND created_at > now() - make_interval(days => $1)) x`, days).Scan(&players)
	langs, lang, _ := localized.ForUser(ctx, s.db, "", r.URL.Query().Get("lang"))
	cats := []map[string]any{}
	rows, err := s.db.QueryContext(ctx, `SELECT r.category, count(*), COALESCE(c.name::text,'{}') FROM quiz_duel_round r JOIN quiz_duel d ON d.id=r.duel_id LEFT JOIN quiz_category c ON c.key=r.category
		WHERE r.category<>'' AND d.created_at > now() - make_interval(days => $1) GROUP BY r.category, c.name::text ORDER BY count(*) DESC LIMIT 10`, days)
	if err == nil {
		for rows.Next() {
			var key, name string
			var n int
			if rows.Scan(&key, &n, &name) == nil {
				cats = append(cats, map[string]any{"key": key, "name": resolveName(langs, lang, name, key), "picks": n})
			}
		}
		rows.Close()
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"days": days, "started": started, "finished": finished, "open": open, "waiting": waiting, "timeouts": timeouts,
		"forfeits": forfeits, "draws": draws, "players": players, "avgMinutes": avgMinutes, "categories": cats})
}
