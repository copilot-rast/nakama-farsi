package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/cup/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/cuphooks"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

func bad(code, field string) error { return pluginkit.Bad(code, field, "") }

func (s *service) registerConsoleHTTP(initializer runtime.Initializer) error {
	p := consolePrefix
	routes := []struct {
		path, method string
		handler      http.HandlerFunc
	}{
		{p + "/settings", http.MethodGet, s.settingsGet},
		{p + "/settings", http.MethodPut, s.settingsPut},
		{p + "/refs", http.MethodGet, s.refs},
		{p + "/list", http.MethodGet, s.cupList},
		{p + "/list", http.MethodPost, s.cupSave},
		{p + "/list/{id}", http.MethodGet, s.cupGet},
		{p + "/list/{id}", http.MethodPut, s.cupSave},
		{p + "/list/{id}", http.MethodDelete, s.cupDelete},
		{p + "/list/{id}/publish", http.MethodPost, s.cupPublish},
		{p + "/list/{id}/start", http.MethodPost, s.cupStart},
		{p + "/list/{id}/cancel", http.MethodPost, s.cupCancel},
		{p + "/list/{id}/players", http.MethodPost, s.playerAdd},
		{p + "/list/{id}/players/{userId}", http.MethodDelete, s.playerRemove},
		{p + "/list/{id}/players/{userId}/disqualify", http.MethodPost, s.playerDisqualify},
		{p + "/list/{id}/swap", http.MethodPost, s.playerSwap},
		{p + "/list/{id}/matches/{matchId}/result", http.MethodPost, s.matchResult},
		{p + "/templates", http.MethodGet, s.templateList},
		{p + "/templates", http.MethodPost, s.templateSave},
		{p + "/templates/{id}", http.MethodPut, s.templateSave},
		{p + "/templates/{id}", http.MethodDelete, s.templateDelete},
		{p + "/templates/{id}/run", http.MethodPost, s.templateRun},
		{p + "/payouts/retry", http.MethodPost, s.payoutRetry},
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) consoleFail(w http.ResponseWriter, err error) {
	if errors.Is(err, errSkip) {
		err = apierr.Conflict("CUP_BUSY")
	}
	s.fail(w, err)
}

func (s *service) settingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, meta, err := settings.Read(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"settings": cfg, "defaults": settings.Defaults(), "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt})
}

func (s *service) settingsPut(w http.ResponseWriter, r *http.Request) {
	var in settings.All
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	in.Normalize()
	if field, code := in.Validate(); field != "" {
		s.fail(w, bad(code, field))
		return
	}
	if err := settings.Save(r.Context(), s.db, in, pluginkit.ConsoleUser(r)); err != nil {
		s.fail(w, err)
		return
	}
	s.settingsGet(w, r)
}

func (s *service) refs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	langs, lang, _ := localized.ForUser(ctx, s.db, "", r.URL.Query().Get("lang"))
	name := func(raw, key string) string {
		if langs != nil {
			if v := langs.Resolve(localized.Parse(raw), lang); v != "" {
				return v
			}
		}
		return key
	}
	out := map[string]any{}
	for field, q := range map[string]string{
		"prizes":     `SELECT key, name::text FROM economy_prize ORDER BY key LIMIT 2000`,
		"currencies": `SELECT key, name::text FROM economy_currency ORDER BY key`,
		"leagues":    `SELECT key, name::text FROM trophy_league ORDER BY min_trophies`,
		"tags":       `SELECT id::text, name::text FROM base_tag ORDER BY key_type, sort, key LIMIT 2000`,
	} {
		list := []map[string]string{}
		if rows, err := s.db.QueryContext(ctx, q); err == nil {
			for rows.Next() {
				var key, raw string
				if rows.Scan(&key, &raw) == nil {
					list = append(list, map[string]string{"key": key, "name": name(raw, key)})
				}
			}
			rows.Close()
		}
		out[field] = list
	}
	games := []map[string]string{{"key": "", "name": "external"}}
	for _, g := range cuphooks.Games() {
		n := g.Name[lang]
		if n == "" {
			n = g.Key
		}
		games = append(games, map[string]string{"key": g.Key, "name": n})
	}
	out["games"] = games
	out["defaults"] = s.settings(ctx).Defaults
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) cleanText(ctx context.Context, field string, in map[string]string, required bool, max int) (map[string]string, error) {
	langs, err := localized.Load(ctx, s.db)
	if err != nil {
		return nil, err
	}
	out, err := langs.Clean(field, localized.Text(in), required, max)
	if err != nil {
		var le *localized.Error
		if errors.As(err, &le) {
			return nil, bad(le.Code, field)
		}
		return nil, err
	}
	return out, nil
}

func cupAdminJSON(c *cupRow) map[string]any {
	return map[string]any{"id": c.ID, "templateId": c.TemplateID, "templateKey": c.TemplateKey, "name": c.Name, "description": c.Description, "iconUrl": c.IconURL,
		"status": c.Status, "announceAt": iso(c.AnnounceAt), "signupAt": isoT(c.SignupAt), "startsAt": isoT(c.StartsAt), "rules": c.Rules, "size": c.Size, "rounds": c.Rounds,
		"currentRound": c.CurrentRound, "players": c.Players, "pool": c.Pool, "championId": c.Champion, "cancelReason": c.CancelReason, "createdBy": c.CreatedBy,
		"createdAt": isoT(c.CreatedAt), "updatedAt": isoT(c.UpdatedAt), "startedAt": iso(c.StartedAt), "finishedAt": iso(c.FinishedAt)}
}

func (s *service) cupList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, limit := pluginkit.Page(r, 30, 100)
	q := r.URL.Query()
	where := []string{"true"}
	args := []any{}
	if st := q.Get("status"); st != "" {
		if !strings.Contains("draft upcoming signup running finished cancelled", st) {
			s.fail(w, bad("INVALID_POLICY", "status"))
			return
		}
		args = append(args, st)
		where = append(where, "status=$"+strconv.Itoa(len(args)))
	}
	if t := q.Get("template"); t != "" {
		args = append(args, t)
		where = append(where, "template_key=$"+strconv.Itoa(len(args)))
	}
	if text := strings.TrimSpace(q.Get("q")); text != "" {
		args = append(args, pluginkit.LikePattern(text))
		where = append(where, "(name::text ILIKE $"+strconv.Itoa(len(args))+" OR template_key ILIKE $"+strconv.Itoa(len(args))+" OR id::text ILIKE $"+strconv.Itoa(len(args))+")")
	}
	cond := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM cup WHERE `+cond, args...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	args = append(args, limit, (page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT `+cupCols+`, (SELECT count(*) FROM cup_player p WHERE p.cup_id=cup.id AND p.status NOT IN ('withdrawn','removed'))
		FROM cup WHERE `+cond+` ORDER BY CASE status WHEN 'running' THEN 0 WHEN 'signup' THEN 1 WHEN 'upcoming' THEN 2 WHEN 'draft' THEN 3 ELSE 4 END, starts_at DESC
		LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		c := &cupRow{}
		var name, desc, rules, pool string
		var entrants int
		if err := rows.Scan(&c.ID, &c.TemplateID, &c.TemplateKey, &name, &desc, &c.IconURL, &c.Status, &c.AnnounceAt, &c.SignupAt, &c.StartsAt, &rules,
			&c.Size, &c.Rounds, &c.CurrentRound, &c.Players, &pool, &c.Champion, &c.CancelReason, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt, &c.StartedAt, &c.FinishedAt, &entrants); err != nil {
			s.fail(w, err)
			return
		}
		full, _ := scanCupFrom(c, name, desc, rules, pool)
		m := cupAdminJSON(full)
		m["entrants"] = entrants
		items = append(items, m)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func scanCupFrom(c *cupRow, name, desc, rules, pool string) (*cupRow, error) {
	c.Name, c.Description = textMap(name), textMap(desc)
	c.Rules = settings.DefaultRules()
	_ = jsonUnmarshal(rules, &c.Rules)
	c.Rules.Normalize()
	_ = jsonUnmarshal(pool, &c.Pool)
	return c, nil
}

func (s *service) cupGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	b, err := s.readBracket(ctx, pluginkit.PathVar(r, "id"))
	if err != nil {
		s.consoleFail(w, err)
		return
	}
	profiles := s.profiles(ctx, b.userIDs())
	players := []map[string]any{}
	for _, p := range b.order {
		pr := profiles[p.UserID]
		players = append(players, map[string]any{"userId": p.UserID, "username": pr.Username, "displayName": pr.DisplayName, "avatarUrl": pr.AvatarURL, "status": p.Status,
			"seed": p.Seed, "trophies": p.Trophies, "place": p.Place, "reached": p.Reached, "fee": p.Fee, "feeState": p.FeeState, "addedBy": p.AddedBy, "joinedAt": isoT(p.JoinedAt)})
	}
	payouts := []map[string]any{}
	if rows, err := s.db.QueryContext(ctx, `SELECT id::text, user_id::text, kind, prize_key, amounts::text, status, tries, last_error, created_at, paid_at FROM cup_payout WHERE cup_id=$1::uuid ORDER BY created_at, kind`, b.cup.ID); err == nil {
		for rows.Next() {
			var id, uid, kind, prize, amounts, status, lastErr string
			var tries int
			var created time.Time
			var paid *time.Time
			if rows.Scan(&id, &uid, &kind, &prize, &amounts, &status, &tries, &lastErr, &created, &paid) == nil {
				am := map[string]int64{}
				_ = jsonUnmarshal(amounts, &am)
				pr := profiles[uid]
				payouts = append(payouts, map[string]any{"id": id, "userId": uid, "username": pr.Username, "kind": kind, "prizeKey": prize, "amounts": am, "status": status,
					"tries": tries, "lastError": lastErr, "createdAt": isoT(created), "paidAt": iso(paid)})
			}
		}
		rows.Close()
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"cup": cupAdminJSON(b.cup), "players": players, "bracket": b.view(profiles), "payouts": payouts, "serverTime": isoT(b.now)})
}

type cupInput struct {
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	IconURL     string            `json:"iconUrl"`
	AnnounceAt  *time.Time        `json:"announceAt"`
	SignupAt    *time.Time        `json:"signupAt"`
	StartsAt    *time.Time        `json:"startsAt"`
	Rules       *settings.Rules   `json:"rules"`
	Publish     bool              `json:"publish"`
}

func (s *service) checkRules(ctx context.Context, field string, r *settings.Rules) error {
	r.Normalize()
	if f, code := r.Validate(s.settings(ctx).Limits.MaxCapacity); f != "" {
		return bad(code, field+"."+f)
	}
	if r.Game != "" {
		if _, ok := cuphooks.Lookup(r.Game); !ok {
			return bad("CUP_GAME_UNKNOWN", field+".game")
		}
	}
	return nil
}

func (s *service) cupSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in cupInput
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	name, err := s.cleanText(ctx, "name", in.Name, true, 80)
	if err != nil {
		s.fail(w, err)
		return
	}
	desc, err := s.cleanText(ctx, "description", in.Description, false, 1000)
	if err != nil {
		s.fail(w, err)
		return
	}
	if len(in.IconURL) > 500 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "iconUrl"))
		return
	}
	actor := pluginkit.ConsoleUser(r)
	if id == "" {
		if in.Rules == nil {
			d := s.settings(ctx).Defaults
			in.Rules = &d
		}
		if err := s.checkRules(ctx, "rules", in.Rules); err != nil {
			s.fail(w, err)
			return
		}
		if in.StartsAt == nil || in.SignupAt == nil {
			s.fail(w, bad("CUP_TIMES_REQUIRED", "startsAt"))
			return
		}
		if !in.SignupAt.Before(*in.StartsAt) {
			s.fail(w, bad("CUP_SIGNUP_AFTER_START", "signupAt"))
			return
		}
		if !in.StartsAt.After(s.now()) {
			s.fail(w, bad("CUP_START_IN_PAST", "startsAt"))
			return
		}
		status := "draft"
		if in.Publish {
			status = "upcoming"
			if !s.now().Before(*in.SignupAt) {
				status = "signup"
			}
		}
		id = uuid.Must(uuid.NewV4()).String()
		if _, err := s.db.ExecContext(ctx, `INSERT INTO cup (id, name, description, icon_url, status, announce_at, signup_at, starts_at, rules, created_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			id, jsonOf(name), jsonOf(desc), in.IconURL, status, in.AnnounceAt, in.SignupAt.UTC(), in.StartsAt.UTC(), jsonOf(in.Rules), actor); err != nil {
			s.fail(w, err)
			return
		}
		r.SetPathValue("id", id)
		s.cupGet(w, r)
		return
	}
	_, err = s.withCup(ctx, id, false, func(tx *sql.Tx, b *bracket) error {
		c := b.cup
		c.Name, c.Description, c.IconURL = name, desc, in.IconURL
		if c.Status == "draft" || c.Status == "upcoming" || c.Status == "signup" {
			if in.Rules != nil {
				if err := s.checkRules(ctx, "rules", in.Rules); err != nil {
					return err
				}
				if in.Rules.Capacity < c.Players {
					return bad("CUP_CAPACITY_BELOW_PLAYERS", "rules.capacity")
				}
				c.Rules = *in.Rules
			}
			if in.SignupAt != nil {
				c.SignupAt = in.SignupAt.UTC()
			}
			if in.StartsAt != nil {
				c.StartsAt = in.StartsAt.UTC()
			}
			if !c.SignupAt.Before(c.StartsAt) {
				return bad("CUP_SIGNUP_AFTER_START", "signupAt")
			}
			c.AnnounceAt = in.AnnounceAt
			if c.Status != "draft" {
				c.Status = "upcoming"
				if !b.now.Before(c.SignupAt) {
					c.Status = "signup"
				}
			}
		}
		_, err := tx.ExecContext(ctx, `UPDATE cup SET name=$2, description=$3, icon_url=$4, rules=$5, signup_at=$6, starts_at=$7, announce_at=$8 WHERE id=$1::uuid`,
			c.ID, jsonOf(c.Name), jsonOf(c.Description), c.IconURL, jsonOf(c.Rules), c.SignupAt, c.StartsAt, c.AnnounceAt)
		return err
	})
	if err != nil {
		s.consoleFail(w, err)
		return
	}
	s.wake()
	s.cupGet(w, r)
}

func (s *service) cupDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	c, err := loadCup(ctx, s.db, id, "")
	if err != nil {
		s.consoleFail(w, apierr.NotFound("CUP_NOT_FOUND"))
		return
	}
	if c.Status != "draft" && c.Status != "cancelled" {
		s.fail(w, apierr.Conflict("CUP_DELETE_ONLY_DRAFT_OR_CANCELLED"))
		return
	}
	var unpaid int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM cup_player WHERE cup_id=$1::uuid AND fee_state='paid'`, id).Scan(&unpaid)
	if unpaid > 0 {
		s.fail(w, apierr.Conflict("CUP_REFUNDS_PENDING"))
		return
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM cup WHERE id=$1::uuid`, id); err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) cupAction(w http.ResponseWriter, r *http.Request, fn func(tx *sql.Tx, b *bracket, actor string) error) {
	ctx := r.Context()
	actor := pluginkit.ConsoleUser(r)
	if _, err := s.withCup(ctx, pluginkit.PathVar(r, "id"), false, func(tx *sql.Tx, b *bracket) error { return fn(tx, b, actor) }); err != nil {
		s.consoleFail(w, err)
		return
	}
	s.wake()
	s.cupGet(w, r)
}

func (s *service) cupPublish(w http.ResponseWriter, r *http.Request) {
	s.cupAction(w, r, func(tx *sql.Tx, b *bracket, _ string) error {
		c := b.cup
		if c.Status != "draft" {
			return apierr.Conflict("CUP_NOT_DRAFT")
		}
		if !c.StartsAt.After(b.now) {
			return bad("CUP_START_IN_PAST", "startsAt")
		}
		c.Status = "upcoming"
		if !b.now.Before(c.SignupAt) {
			c.Status = "signup"
		}
		return nil
	})
}

func (s *service) cupStart(w http.ResponseWriter, r *http.Request) {
	s.cupAction(w, r, func(tx *sql.Tx, b *bracket, _ string) error {
		c := b.cup
		if c.Status != "signup" && c.Status != "upcoming" && c.Status != "draft" {
			return apierr.Conflict("CUP_NOT_STARTABLE")
		}
		c.StartsAt = b.now
		if c.SignupAt.After(b.now) {
			c.SignupAt = b.now
		}
		if _, err := tx.ExecContext(reqCtx(r), `UPDATE cup SET starts_at=$2, signup_at=$3 WHERE id=$1::uuid`, c.ID, c.StartsAt, c.SignupAt); err != nil {
			return err
		}
		return s.startLocked(reqCtx(r), tx, b)
	})
}

func reqCtx(r *http.Request) context.Context { return r.Context() }

func (s *service) cupCancel(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Reason string `json:"reason"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	s.cupAction(w, r, func(tx *sql.Tx, b *bracket, actor string) error {
		if b.cup.Status == "finished" || b.cup.Status == "cancelled" {
			return apierr.Conflict("CUP_ALREADY_OVER")
		}
		reason := strings.TrimSpace(in.Reason)
		if reason == "" {
			reason = "admin"
		}
		s.cancelLocked(b, reason)
		return nil
	})
}

func (s *service) userExists(r *http.Request, id string) bool {
	if !uuidPattern.MatchString(id) {
		return false
	}
	var ok bool
	_ = s.db.QueryRowContext(r.Context(), `SELECT EXISTS (SELECT 1 FROM users WHERE id=$1::uuid)`, id).Scan(&ok)
	return ok
}

func (s *service) playerAdd(w http.ResponseWriter, r *http.Request) {
	var in struct {
		UserID string `json:"userId"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if !s.userExists(r, in.UserID) {
		s.fail(w, bad("PLAYER_NOT_FOUND", "userId"))
		return
	}
	s.cupAction(w, r, func(tx *sql.Tx, b *bracket, actor string) error {
		c := b.cup
		if c.Status != "draft" && c.Status != "upcoming" && c.Status != "signup" {
			return apierr.Conflict("CUP_SIGNUP_CLOSED")
		}
		if p := b.players[in.UserID]; p != nil && p.Status == "registered" {
			return apierr.Conflict("CUP_ALREADY_SIGNED_UP")
		}
		if c.Players >= c.Rules.Capacity {
			return apierr.Conflict("CUP_FULL")
		}
		if _, err := tx.ExecContext(reqCtx(r), `INSERT INTO cup_player (cup_id, user_id, status, added_by, joined_at) VALUES ($1,$2,'registered',$3,$4)
			ON CONFLICT (cup_id, user_id) DO UPDATE SET status='registered', fee='{}', fee_state=CASE WHEN cup_player.fee_state='paid' THEN 'paid' ELSE 'none' END, added_by=EXCLUDED.added_by, joined_at=EXCLUDED.joined_at, updated_at=now()`,
			c.ID, in.UserID, "console:"+actor, b.now); err != nil {
			return err
		}
		c.Players++
		return nil
	})
}

func (s *service) playerRemove(w http.ResponseWriter, r *http.Request) {
	user := pluginkit.PathVar(r, "userId")
	var removed *playerRow
	s.cupAction(w, r, func(tx *sql.Tx, b *bracket, actor string) error {
		c := b.cup
		p := b.players[user]
		if p == nil || p.Status != "registered" {
			return apierr.Conflict("CUP_NOT_SIGNED_UP")
		}
		if c.Status != "draft" && c.Status != "upcoming" && c.Status != "signup" {
			return apierr.Conflict("CUP_USE_DISQUALIFY")
		}
		p.Status = "removed"
		b.touchPlayer(user)
		if c.Players > 0 {
			c.Players--
		}
		removed = p
		return nil
	})
	if removed != nil {
		_ = s.refundPlayer(r.Context(), pluginkit.PathVar(r, "id"), removed, 100)
	}
}

func (b *bracket) voidCurrentGame(m *matchRow) {
	if m.Status != stPlaying {
		return
	}
	if g := b.currentGame(m); g != nil && (g.Status == "launching" || g.Status == "playing") {
		g.Status, g.Reason, g.FinishedAt = "void", "admin", ptr(b.now)
		b.newGame = append(b.newGame, g)
	}
	m.Status, m.GameMatchID = stOpen, ""
	b.touch(m)
}

func (s *service) playerDisqualify(w http.ResponseWriter, r *http.Request) {
	user := pluginkit.PathVar(r, "userId")
	var refund *playerRow
	var in struct {
		Refund bool `json:"refund"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	s.cupAction(w, r, func(tx *sql.Tx, b *bracket, actor string) error {
		c := b.cup
		p := b.players[user]
		if p == nil || (p.Status != "registered" && p.Status != "playing") {
			return apierr.Conflict("CUP_NOT_A_PLAYER")
		}
		if c.Status == "finished" || c.Status == "cancelled" {
			return apierr.Conflict("CUP_ALREADY_OVER")
		}
		p.Status, p.Place = "disqualified", 0
		b.touchPlayer(user)
		if c.Status != "running" {
			if c.Players > 0 {
				c.Players--
			}
			if in.Refund {
				refund = p
			}
			return nil
		}
		for _, m := range b.matches {
			if (m.A != user && m.B != user) || done(m.Status) || m.Status == stPending {
				continue
			}
			b.voidCurrentGame(m)
			m.Actor = actor
			if opp := b.other(m, user); opp != "" && !b.disqualified(opp) {
				b.finish(m, opp, "disqualified")
			} else {
				b.voidMatch(m, "disqualified")
			}
		}
		b.notify([]string{user}, "cup_eliminated", map[string]any{"reason": "disqualified"}, true, false)
		b.advance()
		return nil
	})
	if refund != nil {
		_ = s.refundPlayer(r.Context(), pluginkit.PathVar(r, "id"), refund, 100)
	}
}

func (s *service) playerSwap(w http.ResponseWriter, r *http.Request) {
	var in struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if !s.userExists(r, in.To) {
		s.fail(w, bad("PLAYER_NOT_FOUND", "to"))
		return
	}
	var refund *playerRow
	s.cupAction(w, r, func(tx *sql.Tx, b *bracket, actor string) error {
		c := b.cup
		from := b.players[in.From]
		if from == nil || (from.Status != "registered" && from.Status != "playing") {
			return bad("CUP_NOT_A_PLAYER", "from")
		}
		if to := b.players[in.To]; to != nil && (to.Status == "registered" || to.Status == "playing" || to.Seed > 0) {
			return bad("CUP_ALREADY_SIGNED_UP", "to")
		}
		status, seed := "registered", 0
		if c.Status == "running" {
			m := b.activeMatchOf(in.From)
			if m == nil {
				for _, x := range b.matches {
					if (x.A == in.From || x.B == in.From) && x.Status == stPending {
						m = x
					}
				}
			}
			if m == nil || m.Games > 0 || from.Reached > 1 || m.Round > 1 {
				return apierr.Conflict("CUP_SWAP_TOO_LATE")
			}
			if m.A == in.From {
				m.A, m.ReadyA, m.ShowedA = in.To, nil, false
			} else {
				m.B, m.ReadyB, m.ShowedB = in.To, nil, false
			}
			m.Actor = actor
			b.touch(m)
			status, seed = "playing", from.Seed
			b.notify([]string{in.To}, "cup_round_started", map[string]any{"matchId": m.ID, "round": m.Round, "opponentId": b.other(m, in.To), "closesAt": iso(m.ClosesAt), "swapped": true}, true, false)
		} else if c.Status != "draft" && c.Status != "upcoming" && c.Status != "signup" {
			return apierr.Conflict("CUP_SWAP_TOO_LATE")
		}
		from.Status = "removed"
		b.touchPlayer(in.From)
		refund = from
		_, err := tx.ExecContext(reqCtx(r), `INSERT INTO cup_player (cup_id, user_id, status, seed, reached, trophies, added_by, joined_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			ON CONFLICT (cup_id, user_id) DO UPDATE SET status=EXCLUDED.status, seed=EXCLUDED.seed, reached=EXCLUDED.reached, fee='{}', fee_state=CASE WHEN cup_player.fee_state='paid' THEN 'paid' ELSE 'none' END, added_by=EXCLUDED.added_by, updated_at=now()`,
			c.ID, in.To, status, seed, from.Reached, from.Trophies, "swap:"+actor, from.JoinedAt)
		return err
	})
	if refund != nil {
		_ = s.refundPlayer(r.Context(), pluginkit.PathVar(r, "id"), refund, 100)
	}
}

func (s *service) matchResult(w http.ResponseWriter, r *http.Request) {
	var in struct {
		WinnerID string `json:"winnerId"`
		Void     bool   `json:"void"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	matchID := pluginkit.PathVar(r, "matchId")
	s.cupAction(w, r, func(tx *sql.Tx, b *bracket, actor string) error {
		if b.cup.Status != "running" {
			return apierr.Conflict("CUP_NOT_RUNNING")
		}
		m := b.matchByID(matchID)
		if m == nil {
			return apierr.NotFound("CUP_MATCH_NOT_FOUND")
		}
		if done(m.Status) {
			return apierr.Conflict("CUP_MATCH_OVER")
		}
		if !in.Void && (in.WinnerID == "" || (in.WinnerID != m.A && in.WinnerID != m.B)) {
			return bad("CUP_WINNER_NOT_IN_MATCH", "winnerId")
		}
		if m.A == "" || m.B == "" {
			return apierr.Conflict("CUP_MATCH_NOT_READY")
		}
		b.voidCurrentGame(m)
		m.Actor = actor
		if in.Void {
			b.voidMatch(m, "admin")
		} else {
			b.finish(m, in.WinnerID, "admin")
		}
		b.advance()
		return nil
	})
}

type templateInput struct {
	Key             string            `json:"key"`
	Name            map[string]string `json:"name"`
	Description     map[string]string `json:"description"`
	IconURL         string            `json:"iconUrl"`
	Active          bool              `json:"active"`
	Schedule        string            `json:"schedule"`
	SignupMinutes   int               `json:"signupMinutes"`
	AnnounceMinutes int               `json:"announceMinutes"`
	Rules           *settings.Rules   `json:"rules"`
	Sort            int               `json:"sort"`
}

func (s *service) templateList(w http.ResponseWriter, r *http.Request) {
	list, err := s.templates(r.Context(), false)
	if err != nil {
		s.fail(w, err)
		return
	}
	now := s.now()
	items := []map[string]any{}
	for _, t := range list {
		var open int
		_ = s.db.QueryRowContext(r.Context(), `SELECT count(*) FROM cup WHERE template_id=$1::uuid AND status IN ('upcoming','signup','running')`, t.ID).Scan(&open)
		items = append(items, map[string]any{"id": t.ID, "key": t.Key, "name": t.Name, "description": t.Description, "iconUrl": t.IconURL, "active": t.Active,
			"schedule": t.Schedule, "signupMinutes": t.SignupMinutes, "announceMinutes": t.AnnounceMinutes, "rules": t.Rules, "sort": t.Sort,
			"lastStartsAt": iso(t.LastStartsAt), "nextStart": iso(cronNext(t.Schedule, now)), "openCups": open, "updatedBy": t.UpdatedBy, "updatedAt": isoT(t.UpdatedAt)})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "defaults": s.settings(r.Context()).Defaults})
}

func (s *service) templateSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in templateInput
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	in.Key = strings.TrimSpace(in.Key)
	if !settings.KeyPattern.MatchString(in.Key) {
		s.fail(w, bad("INVALID_KEY", "key"))
		return
	}
	name, err := s.cleanText(ctx, "name", in.Name, true, 80)
	if err != nil {
		s.fail(w, err)
		return
	}
	desc, err := s.cleanText(ctx, "description", in.Description, false, 1000)
	if err != nil {
		s.fail(w, err)
		return
	}
	in.Schedule = strings.TrimSpace(in.Schedule)
	if in.Schedule != "" && cronNext(in.Schedule, s.now()) == nil {
		s.fail(w, bad("INVALID_SCHEDULE", "schedule"))
		return
	}
	if in.Active && in.Schedule == "" {
		s.fail(w, bad("CUP_SCHEDULE_REQUIRED", "schedule"))
		return
	}
	if in.SignupMinutes < 0 || in.SignupMinutes > 43200 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "signupMinutes"))
		return
	}
	if in.AnnounceMinutes < 0 || in.AnnounceMinutes > 43200 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "announceMinutes"))
		return
	}
	if in.Rules == nil {
		d := s.settings(ctx).Defaults
		in.Rules = &d
	}
	if err := s.checkRules(ctx, "rules", in.Rules); err != nil {
		s.fail(w, err)
		return
	}
	actor := pluginkit.ConsoleUser(r)
	if id == "" {
		id = uuid.Must(uuid.NewV4()).String()
		_, err = s.db.ExecContext(ctx, `INSERT INTO cup_template (id, key, name, description, icon_url, active, schedule, signup_minutes, announce_minutes, rules, sort, updated_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, id, in.Key, jsonOf(name), jsonOf(desc), in.IconURL, in.Active, in.Schedule, in.SignupMinutes, in.AnnounceMinutes, jsonOf(in.Rules), in.Sort, actor)
	} else {
		var res sql.Result
		res, err = s.db.ExecContext(ctx, `UPDATE cup_template SET key=$2, name=$3, description=$4, icon_url=$5, active=$6, schedule=$7, signup_minutes=$8, announce_minutes=$9, rules=$10, sort=$11,
			updated_by=$12, updated_at=now() WHERE id::text=$1`, id, in.Key, jsonOf(name), jsonOf(desc), in.IconURL, in.Active, in.Schedule, in.SignupMinutes, in.AnnounceMinutes, jsonOf(in.Rules), in.Sort, actor)
		if err == nil {
			if n, _ := res.RowsAffected(); n == 0 {
				s.fail(w, pluginkit.Missing("CUP_TEMPLATE_NOT_FOUND"))
				return
			}
		}
	}
	if err != nil {
		if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key") {
			s.fail(w, bad("KEY_TAKEN", "key"))
			return
		}
		s.fail(w, err)
		return
	}
	s.wake()
	s.templateList(w, r)
}

func (s *service) templateDelete(w http.ResponseWriter, r *http.Request) {
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM cup_template WHERE id::text=$1`, pluginkit.PathVar(r, "id"))
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("CUP_TEMPLATE_NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) templateRun(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in struct {
		StartsAt *time.Time `json:"startsAt"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	t, err := scanTemplate(s.db.QueryRowContext(ctx, `SELECT `+templateCols+` FROM cup_template WHERE id::text=$1`, pluginkit.PathVar(r, "id")))
	if err != nil {
		s.fail(w, pluginkit.Missing("CUP_TEMPLATE_NOT_FOUND"))
		return
	}
	start := s.now().Add(time.Duration(t.SignupMinutes) * time.Minute)
	if in.StartsAt != nil {
		start = *in.StartsAt
	}
	if !start.After(s.now()) {
		s.fail(w, bad("CUP_START_IN_PAST", "startsAt"))
		return
	}
	id, err := s.createFromTemplate(ctx, t, start.Truncate(time.Second), "console:"+pluginkit.ConsoleUser(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	s.wake()
	r.SetPathValue("id", id)
	s.cupGet(w, r)
}

func (s *service) payoutRetry(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CupID string `json:"cupId"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	ctx := r.Context()
	q := `UPDATE cup_payout SET status='pending', tries=0 WHERE status='failed'`
	args := []any{}
	if in.CupID != "" {
		q += ` AND cup_id::text=$1`
		args = append(args, in.CupID)
	}
	res, err := s.db.ExecContext(ctx, q, args...)
	if err != nil {
		s.fail(w, err)
		return
	}
	n, _ := res.RowsAffected()
	s.payPending(ctx, in.CupID, 500)
	s.retryRefunds(ctx, 500)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"requeued": n})
}
