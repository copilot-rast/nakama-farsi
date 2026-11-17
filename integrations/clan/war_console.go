package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
	"github.com/jackc/pgx/v5/pgconn"
)

type consoleRoute = struct {
	path, method string
	handler      http.HandlerFunc
}

func (s *service) warRoutes() []consoleRoute {
	p := consolePrefix + "/wars"
	return []consoleRoute{
		{p + "/configs", http.MethodGet, s.warConfigList},
		{p + "/configs", http.MethodPost, s.warConfigSave},
		{p + "/configs/{id}", http.MethodPut, s.warConfigSave},
		{p + "/configs/{id}", http.MethodDelete, s.warConfigDelete},
		{p + "/rounds", http.MethodGet, s.roundList},
		{p + "/rounds", http.MethodPost, s.roundCreate},
		{p + "/rounds/{id}", http.MethodGet, s.roundGet},
		{p + "/rounds/{id}/start", http.MethodPost, s.roundStart},
		{p + "/rounds/{id}/settle", http.MethodPost, s.roundSettle},
		{p + "/rounds/{id}/cancel", http.MethodPost, s.roundCancel},
	}
}

func (s *service) warConfigList(w http.ResponseWriter, r *http.Request) {
	list, err := s.warConfigs(r.Context(), false)
	if err != nil {
		s.fail(w, err)
		return
	}
	now := s.now()
	items := make([]map[string]any, 0, len(list))
	for _, c := range list {
		raw, _ := json.Marshal(c)
		m := map[string]any{}
		_ = json.Unmarshal(raw, &m)
		m["nextStart"] = iso(cronNext(c.Schedule, now))
		items = append(items, m)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "defaults": map[string]any{"points": defaultRules(modePoints), "duel": defaultRules(modeDuel)}})
}

func (s *service) warConfigSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var in warConfig
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	in.Key = strings.TrimSpace(in.Key)
	if !settings.KeyPattern.MatchString(in.Key) {
		s.fail(w, bad("INVALID_KEY", "key"))
		return
	}
	if in.Mode != modePoints && in.Mode != modeDuel {
		s.fail(w, bad("INVALID_POLICY", "mode"))
		return
	}
	var err error
	if in.Name, err = s.cleanText(ctx, "name", in.Name, true, 80); err != nil {
		s.fail(w, err)
		return
	}
	if in.Description, err = s.cleanText(ctx, "description", in.Description, false, 500); err != nil {
		s.fail(w, err)
		return
	}
	in.Schedule = strings.TrimSpace(in.Schedule)
	if in.Schedule != "" && cronNext(in.Schedule, s.now()) == nil {
		s.fail(w, bad("INVALID_SCHEDULE", "schedule"))
		return
	}
	switch {
	case in.DurationMinutes < 1 || in.DurationMinutes > 43200:
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "durationMinutes"))
		return
	case in.Signup != "auto" && in.Signup != "leader":
		s.fail(w, bad("INVALID_POLICY", "signup"))
		return
	case in.SignupMinutes < 0 || in.SignupMinutes > 43200:
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "signupMinutes"))
		return
	case in.Pairing != "random" && in.Pairing != "level" && in.Pairing != "trophies":
		s.fail(w, bad("INVALID_POLICY", "pairing"))
		return
	}
	in.Rules.normalize(in.Mode)
	if field, code := in.Rules.validate(in.Mode); code != "" {
		s.fail(w, bad(code, field))
		return
	}
	for _, key := range []string{in.Rules.Prizes.Win.PrizeKey, in.Rules.Prizes.Lose.PrizeKey, in.Rules.Prizes.Draw.PrizeKey, in.Rules.Prizes.Bye.PrizeKey} {
		if key == "" {
			continue
		}
		var ok bool
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_prize WHERE key=$1)`, key).Scan(&ok)
		if !ok {
			s.fail(w, bad("PRIZE_NOT_FOUND", "rules.prizes"))
			return
		}
	}
	name, _ := json.Marshal(in.Name)
	desc, _ := json.Marshal(in.Description)
	rules, _ := json.Marshal(in.Rules)
	by := pluginkit.ConsoleUser(r)
	var res sql.Result
	if id == "" {
		res, err = s.db.ExecContext(ctx, `INSERT INTO clan_war_config (id, key, name, description, mode, is_active, schedule, duration_minutes, signup, signup_minutes, pairing, rules, updated_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, uuid.Must(uuid.NewV4()).String(), in.Key, name, desc, in.Mode, in.Active, in.Schedule, in.DurationMinutes, in.Signup, in.SignupMinutes, in.Pairing, rules, by)
	} else {
		res, err = s.db.ExecContext(ctx, `UPDATE clan_war_config SET key=$2, name=$3, description=$4, mode=$5, is_active=$6, schedule=$7, duration_minutes=$8, signup=$9, signup_minutes=$10, pairing=$11,
			rules=$12, updated_by=$13, updated_at=now() WHERE id::text=$1`, id, in.Key, name, desc, in.Mode, in.Active, in.Schedule, in.DurationMinutes, in.Signup, in.SignupMinutes, in.Pairing, rules, by)
	}
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "CLAN_WAR_CONFIG_EXISTS", Field: "key"})
			return
		}
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("CLAN_WAR_CONFIG_NOT_FOUND"))
		return
	}
	s.wake()
	s.warConfigList(w, r)
}

func (s *service) warConfigDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	var busy bool
	_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM clan_war_round WHERE config_id::text=$1 AND status IN ('signup','active'))`, id).Scan(&busy)
	if busy {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "IN_USE"})
		return
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM clan_war_config WHERE id::text=$1`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, pluginkit.Missing("CLAN_WAR_CONFIG_NOT_FOUND"))
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func roundJSON(rd *roundRow) map[string]any {
	return map[string]any{"id": rd.ID, "configKey": rd.ConfigKey, "mode": rd.Mode, "status": rd.Status, "signupAt": iso(rd.SignupAt), "startsAt": rd.StartsAt.UTC().Format(time.RFC3339),
		"endsAt": rd.EndsAt.UTC().Format(time.RFC3339), "name": rd.Name, "pairing": rd.Pairing, "signup": rd.Signup, "wars": rd.Wars, "actor": rd.Actor,
		"createdAt": rd.CreatedAt.UTC().Format(time.RFC3339), "settledAt": iso(rd.SettledAt), "rules": rd.Rules}
}

func (s *service) roundList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, limit := pluginkit.Page(r, 25, 100)
	args := &pluginkit.Args{}
	where := []string{"true"}
	if st := r.URL.Query().Get("status"); st != "" {
		where = append(where, "status = "+args.Add(st))
	}
	if k := r.URL.Query().Get("configKey"); k != "" {
		where = append(where, "config_key = "+args.Add(k))
	}
	cond := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM clan_war_round WHERE `+cond, args.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+roundCols+`, (SELECT count(*) FROM clan_war_signup x WHERE x.round_id=clan_war_round.id) FROM clan_war_round WHERE `+cond+` ORDER BY starts_at DESC LIMIT `+args.Add(limit)+` OFFSET `+args.Add((page-1)*limit), args.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var signups int
		rd := &roundRow{}
		var rules, name string
		if err := rows.Scan(&rd.ID, &rd.ConfigKey, &rd.Mode, &rd.Status, &rd.SignupAt, &rd.StartsAt, &rd.EndsAt, &rules, &name, &rd.Pairing, &rd.Signup, &rd.Wars, &rd.Actor, &rd.CreatedAt, &rd.SettledAt, &signups); err != nil {
			s.fail(w, err)
			return
		}
		_ = json.Unmarshal([]byte(name), &rd.Name)
		rd.Rules = defaultRules(rd.Mode)
		_ = json.Unmarshal([]byte(rules), &rd.Rules)
		m := roundJSON(rd)
		m["signups"] = signups
		delete(m, "rules")
		items = append(items, m)
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) roundCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in struct {
		ConfigKey       string     `json:"configKey"`
		StartsAt        *time.Time `json:"startsAt"`
		DurationMinutes int        `json:"durationMinutes"`
		StartNow        bool       `json:"startNow"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	c, err := scanWarConfig(s.db.QueryRowContext(ctx, `SELECT `+warConfigCols+` FROM clan_war_config WHERE key=$1`, strings.TrimSpace(in.ConfigKey)))
	if errors.Is(err, sql.ErrNoRows) {
		s.fail(w, pluginkit.Missing("CLAN_WAR_CONFIG_NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	if in.DurationMinutes < 0 || in.DurationMinutes > 43200 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "durationMinutes"))
		return
	}
	start := s.now().UTC().Truncate(time.Second)
	if in.StartsAt != nil && !in.StartNow {
		start = in.StartsAt.UTC()
	}
	id, err := s.createRound(ctx, c, start, in.DurationMinutes, pluginkit.ConsoleUser(r))
	if err != nil {
		s.fail(w, err)
		return
	}
	if in.StartNow {
		if err := s.startRound(ctx, id); err != nil {
			s.fail(w, err)
			return
		}
	}
	r.SetPathValue("id", id)
	s.roundGet(w, r)
}

func (s *service) roundGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	rd, err := scanRound(s.db.QueryRowContext(ctx, `SELECT `+roundCols+` FROM clan_war_round WHERE id::text=$1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		s.fail(w, pluginkit.Missing("CLAN_WAR_ROUND_NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	rows, err := s.db.QueryContext(ctx, `SELECT w.id::text, w.clan_a::text, COALESCE(ga.name,''), COALESCE(w.clan_b::text,''), COALESCE(gb.name,''), w.score_a, w.score_b, w.wins_a, w.wins_b, w.status, w.result,
		(SELECT count(*) FROM clan_war_member m WHERE m.war_id=w.id AND m.group_id=w.clan_a), (SELECT count(*) FROM clan_war_member m WHERE m.war_id=w.id AND m.group_id=w.clan_b),
		(SELECT count(*) FROM clan_war_event e WHERE e.war_id=w.id)
		FROM clan_war w LEFT JOIN groups ga ON ga.id=w.clan_a LEFT JOIN groups gb ON gb.id=w.clan_b WHERE w.round_id::text=$1 ORDER BY w.score_a + w.score_b DESC`, rd.ID)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	wars := []map[string]any{}
	for rows.Next() {
		var wid, a, an, b, bn, status, result string
		var sa, sb int64
		var wa, wb, ma, mb, events int
		if err := rows.Scan(&wid, &a, &an, &b, &bn, &sa, &sb, &wa, &wb, &status, &result, &ma, &mb, &events); err != nil {
			s.fail(w, err)
			return
		}
		wars = append(wars, map[string]any{"warId": wid, "clanA": a, "nameA": an, "clanB": b, "nameB": bn, "scoreA": sa, "scoreB": sb, "winsA": wa, "winsB": wb, "status": status,
			"result": result, "membersA": ma, "membersB": mb, "events": events})
	}
	signups := []map[string]any{}
	srows, err := s.db.QueryContext(ctx, `SELECT x.group_id::text, COALESCE(g.name,''), x.created_at FROM clan_war_signup x LEFT JOIN groups g ON g.id=x.group_id WHERE x.round_id::text=$1 ORDER BY x.created_at`, rd.ID)
	if err == nil {
		for srows.Next() {
			var gid, name string
			var at time.Time
			if srows.Scan(&gid, &name, &at) == nil {
				signups = append(signups, map[string]any{"clanId": gid, "name": name, "createdAt": at.UTC().Format(time.RFC3339)})
			}
		}
		srows.Close()
	}
	out := roundJSON(rd)
	out["items"] = wars
	out["signupList"] = signups
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) roundStart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	res, err := s.db.ExecContext(ctx, `UPDATE clan_war_round SET starts_at=LEAST(starts_at, now()), ends_at=GREATEST(ends_at, now() + (ends_at - starts_at)) WHERE id::text=$1 AND status='signup'`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "CLAN_WAR_ROUND_NOT_OPEN"})
		return
	}
	if err := s.startRound(ctx, id); err != nil {
		s.fail(w, err)
		return
	}
	s.roundGet(w, r)
}

func (s *service) roundSettle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	res, err := s.db.ExecContext(ctx, `UPDATE clan_war_round SET ends_at=LEAST(ends_at, now()) WHERE id::text=$1 AND status='active'`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "CLAN_WAR_ROUND_NOT_ACTIVE"})
		return
	}
	if err := s.settleRound(ctx, id); err != nil {
		s.fail(w, err)
		return
	}
	s.roundGet(w, r)
}

func (s *service) roundCancel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	err := pluginkit.InTx(ctx, s.db, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `UPDATE clan_war_round SET status='cancelled', settled_at=now() WHERE id::text=$1 AND status IN ('signup','active')`, id)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "CLAN_WAR_ROUND_NOT_OPEN"}
		}
		_, err = tx.ExecContext(ctx, `UPDATE clan_war SET status='settled', result='', settled_at=now() WHERE round_id::text=$1 AND status='active'`, id)
		return err
	})
	if err != nil {
		s.fail(w, err)
		return
	}
	s.roundGet(w, r)
}
