package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

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
		{p + "/list", http.MethodGet, s.clanList},
		{p + "/list/{id}", http.MethodGet, s.clanGet},
		{p + "/list/{id}", http.MethodPut, s.clanPut},
		{p + "/list/{id}", http.MethodDelete, s.clanDelete},
		{p + "/list/{id}/xp", http.MethodPost, s.clanXP},
		{p + "/list/{id}/members/{userId}", http.MethodPut, s.memberRole},
		{p + "/list/{id}/members/{userId}", http.MethodDelete, s.memberKick},
		{p + "/levels", http.MethodGet, s.levelList},
		{p + "/levels", http.MethodPut, s.levelPut},
		{p + "/payouts/retry", http.MethodPost, s.payoutRetry},
	}
	for _, r := range s.warRoutes() {
		routes = append(routes, r)
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) settingsGet(w http.ResponseWriter, r *http.Request) {
	cfg, meta, err := settings.Read(r.Context(), s.db)
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"settings": cfg, "defaults": settings.Defaults(), "updatedBy": meta.UpdatedBy, "updatedAt": meta.UpdatedAt,
		"prizes": economyhooks.PrizesAvailable()})
}

func (s *service) settingsPut(w http.ResponseWriter, r *http.Request) {
	in := settings.Defaults()
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	in.Normalize()
	if field, code := in.Validate(); code != "" {
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
	out := map[string]any{}
	for name, q := range map[string]string{
		"prizes":        `SELECT key, name::text FROM economy_prize ORDER BY key LIMIT 2000`,
		"currencies":    `SELECT key, name::text FROM economy_currency ORDER BY key`,
		"economyEvents": `SELECT key, name::text FROM economy_event_type ORDER BY key LIMIT 2000`,
	} {
		list := []map[string]string{}
		rows, err := s.db.QueryContext(ctx, q)
		if err == nil {
			for rows.Next() {
				var key, raw string
				if rows.Scan(&key, &raw) == nil {
					name := key
					if langs != nil {
						if v := langs.Resolve(localized.Parse(raw), lang); v != "" {
							name = v
						}
					}
					list = append(list, map[string]string{"key": key, "name": name})
				}
			}
			rows.Close()
		}
		out[name] = list
	}
	out["gameEvents"] = []string{"task_completed", "trophy_changed", "clan_war_won"}
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

func (s *service) consoleFail(w http.ResponseWriter, err error) {
	if e, ok := apierr.As(err); ok && e.Status < 500 {
		pluginkit.WriteJSON(w, e.Status, map[string]any{"error": e.Code, "message": e.Code, "field": e.Field, "meta": e.Meta})
		return
	}
	s.fail(w, err)
}

func (s *service) clanList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, limit := pluginkit.Page(r, 25, 100)
	args := &pluginkit.Args{}
	where := []string{"true"}
	if q := strings.TrimSpace(r.URL.Query().Get("q")); q != "" {
		if uuidPattern.MatchString(q) {
			where = append(where, "c.group_id::text = "+args.Add(strings.ToLower(q)))
		} else {
			p := args.Add(pluginkit.LikePattern(q))
			where = append(where, "(g.name ILIKE "+p+" OR c.tag ILIKE "+p+")")
		}
	}
	if jt := r.URL.Query().Get("joinType"); jt != "" {
		where = append(where, "c.join_type = "+args.Add(jt))
	}
	order := "c.trophies DESC"
	switch r.URL.Query().Get("sort") {
	case "level":
		order = "c.level DESC, c.xp DESC"
	case "members":
		order = "g.edge_count DESC"
	case "new":
		order = "c.created_at DESC"
	case "wars":
		order = "c.war_wins DESC"
	}
	cond := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM clan c JOIN groups g ON g.id=c.group_id WHERE `+cond, args.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+clanCols+` FROM clan c JOIN groups g ON g.id=c.group_id WHERE `+cond+` ORDER BY `+order+`, c.group_id LIMIT `+args.Add(limit)+` OFFSET `+args.Add((page-1)*limit), args.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	list := []*clanRow{}
	for rows.Next() {
		c, err := scanClan(rows)
		if err != nil {
			rows.Close()
			s.fail(w, err)
			return
		}
		list = append(list, c)
	}
	rows.Close()
	items := make([]map[string]any, 0, len(list))
	for _, c := range list {
		items = append(items, s.clanView(ctx, c, "", ""))
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) clanGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	c, err := s.clanByID(ctx, s.db, pluginkit.PathVar(r, "id"))
	if err != nil {
		s.consoleFail(w, err)
		return
	}
	members, err := s.members(ctx, c.ID)
	if err != nil {
		s.fail(w, err)
		return
	}
	xp := []map[string]any{}
	rows, err := s.db.QueryContext(ctx, `SELECT l.source, l.amount, COALESCE(l.user_id::text,''), COALESCE(u.username,''), l.created_at FROM clan_xp_log l LEFT JOIN users u ON u.id=l.user_id WHERE l.group_id::text=$1 ORDER BY l.id DESC LIMIT 50`, c.ID)
	if err == nil {
		for rows.Next() {
			var src, uid, uname string
			var amount int64
			var at time.Time
			if rows.Scan(&src, &amount, &uid, &uname, &at) == nil {
				xp = append(xp, map[string]any{"source": src, "amount": amount, "userId": uid, "username": uname, "createdAt": at.UTC().Format(time.RFC3339)})
			}
		}
		rows.Close()
	}
	requests := 0
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM group_edge WHERE source_id::text=$1 AND state=3`, c.ID).Scan(&requests)
	wars := []map[string]any{}
	wrows, err := s.db.QueryContext(ctx, `SELECT w.id::text, r.config_key, r.mode, w.status, w.result, w.clan_a::text, COALESCE(w.clan_b::text,''), w.score_a, w.score_b, r.ends_at FROM clan_war w JOIN clan_war_round r ON r.id=w.round_id
		WHERE w.clan_a::text=$1 OR w.clan_b::text=$1 ORDER BY r.starts_at DESC LIMIT 20`, c.ID)
	if err == nil {
		for wrows.Next() {
			var id, key, mode, status, result, a, b string
			var sa, sb int64
			var ends time.Time
			if wrows.Scan(&id, &key, &mode, &status, &result, &a, &b, &sa, &sb, &ends) == nil {
				wars = append(wars, map[string]any{"warId": id, "configKey": key, "mode": mode, "status": status, "outcome": outcomeFor(result, a == c.ID, "none"), "clanA": a, "clanB": b,
					"scoreA": sa, "scoreB": sb, "endsAt": ends.UTC().Format(time.RFC3339)})
			}
		}
		wrows.Close()
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"clan": s.clanView(ctx, c, "", ""), "members": members, "xp": xp, "requests": requests, "wars": wars})
}

func (s *service) clanPut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in profileInput
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	cfg := s.settings(ctx)
	if err := in.check(cfg); err != nil {
		s.consoleFail(w, err)
		return
	}
	c, err := s.clanByID(ctx, s.db, pluginkit.PathVar(r, "id"))
	if err != nil {
		s.consoleFail(w, err)
		return
	}
	if err := s.applyProfile(ctx, cfg, c, in); err != nil {
		s.consoleFail(w, err)
		return
	}
	s.clanGet(w, r)
}

func (s *service) clanDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	c, err := s.clanByID(ctx, s.db, pluginkit.PathVar(r, "id"))
	if err != nil {
		s.consoleFail(w, err)
		return
	}
	var busy bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM clan_war WHERE (clan_a::text=$1 OR clan_b::text=$1) AND status='active')`, c.ID).Scan(&busy)
	if busy {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "CLAN_IN_WAR"})
		return
	}
	members := s.memberIDs(ctx, c.ID)
	if err := s.groups.Delete(ctx, c.ID); err != nil && !errors.Is(err, runtime.ErrGroupNotFound) {
		s.fail(w, err)
		return
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM clan WHERE group_id::text=$1`, c.ID); err != nil {
		s.fail(w, err)
		return
	}
	s.notify(ctx, members, "clan_deleted", map[string]any{"clanId": c.ID}, true)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) clanXP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in struct {
		Amount int64 `json:"amount"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if in.Amount <= 0 || in.Amount > 100000000 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "amount"))
		return
	}
	c, err := s.clanByID(ctx, s.db, pluginkit.PathVar(r, "id"))
	if err != nil {
		s.consoleFail(w, err)
		return
	}
	if _, err := s.addXP(ctx, c.ID, "", "admin:"+uuid.Must(uuid.NewV4()).String(), "admin", in.Amount); err != nil {
		s.fail(w, err)
		return
	}
	s.clanGet(w, r)
}

func (s *service) memberRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in struct {
		Role string `json:"role"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if _, ok := roleRank[in.Role]; !ok {
		s.fail(w, bad("INVALID_POLICY", "role"))
		return
	}
	clanID, userID := strings.ToLower(pluginkit.PathVar(r, "id")), strings.ToLower(pluginkit.PathVar(r, "userId"))
	m, err := s.memberOf(ctx, s.db, userID)
	if err != nil {
		s.fail(w, err)
		return
	}
	if m == nil || m.ClanID != clanID {
		s.fail(w, pluginkit.Missing("CLAN_MEMBER_NOT_FOUND"))
		return
	}
	if err := s.forceRole(ctx, clanID, userID, m.Role, in.Role); err != nil {
		s.fail(w, err)
		return
	}
	s.notify(ctx, []string{userID}, "clan_role_changed", map[string]any{"clanId": clanID, "role": in.Role, "from": m.Role}, true)
	s.clanGet(w, r)
}

func (s *service) forceRole(ctx context.Context, clanID, userID, from, to string) error {
	if from == to {
		return nil
	}
	if to == settings.RoleLeader {
		old := s.withRoles(ctx, clanID, []string{settings.RoleLeader})
		if err := s.makeLeader(ctx, clanID, userID); err != nil {
			return err
		}
		for _, id := range old {
			if err := s.groups.Demote(ctx, clanID, []string{id}); err != nil {
				return err
			}
			if _, err := s.db.ExecContext(ctx, `UPDATE clan_member SET role='co_leader' WHERE group_id::text=$1 AND user_id::text=$2`, clanID, id); err != nil {
				return err
			}
		}
		return nil
	}
	if from == settings.RoleLeader {
		return &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "CLAN_LEADER_MUST_TRANSFER"}
	}
	for cur, want := stateOf(from), stateOf(to); cur != want; {
		var err error
		if cur > want {
			err = s.groups.Promote(ctx, clanID, []string{userID})
			cur--
		} else {
			err = s.groups.Demote(ctx, clanID, []string{userID})
			cur++
		}
		if err != nil {
			return err
		}
	}
	_, err := s.db.ExecContext(ctx, `UPDATE clan_member SET role=$3 WHERE group_id::text=$1 AND user_id::text=$2`, clanID, userID, to)
	return err
}

func (s *service) memberKick(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clanID, userID := strings.ToLower(pluginkit.PathVar(r, "id")), strings.ToLower(pluginkit.PathVar(r, "userId"))
	m, err := s.memberOf(ctx, s.db, userID)
	if err != nil {
		s.fail(w, err)
		return
	}
	if m == nil || m.ClanID != clanID {
		s.fail(w, pluginkit.Missing("CLAN_MEMBER_NOT_FOUND"))
		return
	}
	if m.Role == settings.RoleLeader {
		if next := s.successor(ctx, clanID, userID); next != "" {
			if err := s.makeLeader(ctx, clanID, next); err != nil {
				s.fail(w, err)
				return
			}
		}
	}
	if err := s.groups.Kick(ctx, clanID, []string{userID}); err != nil {
		s.fail(w, err)
		return
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM clan_member WHERE group_id::text=$1 AND user_id::text=$2`, clanID, userID)
	_, _ = s.db.ExecContext(ctx, `INSERT INTO clan_cooldown (user_id, group_id, reason, created_at) VALUES ($1,$2,'kicked',now()) ON CONFLICT (user_id, group_id) DO UPDATE SET reason='kicked', created_at=now()`, userID, clanID)
	s.refreshTrophies(ctx, clanID)
	s.notify(ctx, []string{userID}, "clan_kicked", map[string]any{"clanId": clanID}, true)
	s.clanGet(w, r)
}

func (s *service) levelList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := s.levels(ctx)
	if err != nil {
		s.fail(w, err)
		return
	}
	counts := map[int]int{}
	rows, err := s.db.QueryContext(ctx, `SELECT level, count(*) FROM clan GROUP BY 1`)
	if err == nil {
		for rows.Next() {
			var lv, n int
			if rows.Scan(&lv, &n) == nil {
				counts[lv] = n
			}
		}
		rows.Close()
	}
	items := make([]map[string]any, 0, len(list))
	for _, l := range list {
		items = append(items, map[string]any{"level": l.Level, "xp": l.XP, "name": l.Name, "iconUrl": l.IconURL, "memberCap": l.MemberCap, "prizeKey": l.PrizeKey, "perks": l.Perks,
			"clans": counts[l.Level], "updatedBy": l.UpdatedBy})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *service) levelPut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in struct {
		Items []level `json:"items"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if len(in.Items) == 0 || len(in.Items) > 500 {
		s.fail(w, bad("CLAN_LEVELS_REQUIRED", "items"))
		return
	}
	list := newRoad(in.Items)
	if list[0].Level != 1 || list[0].XP != 0 {
		s.fail(w, bad("CLAN_LEVEL_ONE_REQUIRED", "items"))
		return
	}
	for i, l := range list {
		if i > 0 && (l.Level == list[i-1].Level || l.XP <= list[i-1].XP) {
			s.fail(w, bad("CLAN_LEVEL_ORDER", "items"))
			return
		}
		if l.MemberCap < 1 || l.MemberCap > 1000 || len(l.IconURL) > 500 {
			s.fail(w, bad("VALUE_OUT_OF_RANGE", "items"))
			return
		}
		if l.PrizeKey != "" {
			var ok bool
			_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_prize WHERE key=$1)`, l.PrizeKey).Scan(&ok)
			if !ok {
				s.fail(w, bad("PRIZE_NOT_FOUND", "items"))
				return
			}
		}
		name, err := s.cleanText(ctx, "name", l.Name, false, 80)
		if err != nil {
			s.fail(w, err)
			return
		}
		list[i].Name = name
		if list[i].Perks == nil {
			list[i].Perks = map[string]any{}
		}
	}
	by := pluginkit.ConsoleUser(r)
	err := pluginkit.InTx(ctx, s.db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM clan_level`); err != nil {
			return err
		}
		for _, l := range list {
			name, _ := json.Marshal(l.Name)
			perks, _ := json.Marshal(l.Perks)
			if _, err := tx.ExecContext(ctx, `INSERT INTO clan_level (level, xp, name, icon_url, member_cap, prize_key, perks, updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
				l.Level, l.XP, name, l.IconURL, l.MemberCap, l.PrizeKey, perks, by); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	s.invalidate()
	go s.reconcileCaps(context.Background())
	s.levelList(w, r)
}

func (s *service) payoutRetry(w http.ResponseWriter, r *http.Request) {
	res, err := s.db.ExecContext(r.Context(), `UPDATE clan_payout SET status='pending', tries=0 WHERE status='failed'`)
	if err != nil {
		s.fail(w, err)
		return
	}
	n, _ := res.RowsAffected()
	s.payPending(r.Context(), "", 500)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"retried": n})
}
