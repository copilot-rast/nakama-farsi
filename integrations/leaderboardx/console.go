package main

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/leaderboardx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/activity"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

func (s *service) registerConsoleHTTP(initializer runtime.Initializer) error {
	type route struct {
		path, method string
		handler      http.HandlerFunc
	}
	p := consolePrefix
	routes := []route{
		{p + "/settings", http.MethodGet, s.settingsGet},
		{p + "/settings", http.MethodPut, s.settingsPut},
		{p + "/refs", http.MethodGet, s.refs},
		{p + "/boards", http.MethodGet, s.boardList},
		{p + "/boards", http.MethodPost, s.boardSave},
		{p + "/boards/{id}", http.MethodGet, s.boardGet},
		{p + "/boards/{id}", http.MethodPut, s.boardSave},
		{p + "/boards/{id}", http.MethodDelete, s.boardDelete},
		{p + "/boards/{id}/standings", http.MethodGet, s.standings},
		{p + "/boards/{id}/records/{userId}", http.MethodPut, s.recordSet},
		{p + "/boards/{id}/records/{userId}", http.MethodDelete, s.recordDelete},
		{p + "/boards/{id}/close", http.MethodPost, s.boardClose},
		{p + "/periods", http.MethodGet, s.periodList},
		{p + "/periods/{id}", http.MethodGet, s.periodGet},
		{p + "/periods/{id}/retry", http.MethodPost, s.periodRetry},
		{p + "/players/{id}", http.MethodGet, s.playerGet},
		{p + "/report", http.MethodGet, s.report},
	}
	for _, r := range routes {
		if err := initializer.RegisterConsoleHttp(r.path, r.handler, r.method); err != nil {
			return err
		}
	}
	return nil
}

func bad(code, field string) error { return pluginkit.Bad(code, field, "") }

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
		"currencies":    `SELECT key, name::text FROM economy_currency ORDER BY key`,
		"prizes":        `SELECT key, name::text FROM economy_prize ORDER BY key LIMIT 2000`,
		"prizeGroups":   `SELECT key, name::text FROM economy_prize_group ORDER BY key`,
		"economyEvents": `SELECT key, name::text FROM economy_event_type ORDER BY key LIMIT 2000`,
		"coreBoards":    `SELECT id, '{}' FROM leaderboard WHERE duration = 0 AND id NOT IN (SELECT key FROM leaderboardx_board) ORDER BY id LIMIT 500`,
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

var (
	sourceTypes = map[string]bool{"economy_event": true, "game_event": true, "client": true, "server": true}
	operators   = map[string]bool{"best": true, "set": true, "incr": true, "decr": true}
)

func (s *service) cleanBoard(ctx context.Context, b *board) error {
	var err error
	b.Key = strings.TrimSpace(b.Key)
	if !settings.KeyPattern.MatchString(b.Key) {
		return bad("INVALID_KEY", "key")
	}
	if b.Scope == "" {
		b.Scope = scopePlayer
	}
	if b.Scope != scopePlayer && b.Scope != scopeClan {
		return bad("INVALID_POLICY", "scope")
	}
	if b.Scope == scopeClan && len(b.Rewards) > 0 {
		return bad("LEADERBOARD_CLAN_NO_REWARDS", "rewards")
	}
	if b.Scope == scopeClan && b.Limits.ClientSubmit {
		return bad("LEADERBOARD_CLAN_NO_SUBMIT", "limits.clientSubmit")
	}
	if b.Name, err = s.cleanText(ctx, "name", b.Name, true, 120); err != nil {
		return err
	}
	if b.Description, err = s.cleanText(ctx, "description", b.Description, false, 1000); err != nil {
		return err
	}
	b.IconURL = strings.TrimSpace(b.IconURL)
	if len(b.IconURL) > 500 {
		return bad("VALUE_TOO_LONG", "iconUrl")
	}
	if b.SortOrder != "desc" && b.SortOrder != "asc" {
		return bad("INVALID_POLICY", "sortOrder")
	}
	if !operators[b.Operator] {
		return bad("INVALID_POLICY", "operator")
	}
	cron, ok := schedule(b.Period, b.ResetSchedule)
	if !ok {
		return bad("LEADERBOARD_INVALID_SCHEDULE", "resetSchedule")
	}
	b.ResetSchedule = cron
	if b.StartsAt != nil && b.EndsAt != nil && !b.EndsAt.After(*b.StartsAt) {
		return bad("LEADERBOARD_INVALID_WINDOW", "endsAt")
	}
	if b.MaxPeriods < 0 || b.MaxPeriods > 100000 {
		return bad("VALUE_OUT_OF_RANGE", "maxPeriods")
	}
	src := &b.Source
	src.Event, src.Context, src.Field = strings.TrimSpace(src.Event), strings.TrimSpace(src.Context), strings.TrimSpace(src.Field)
	if !sourceTypes[src.Type] {
		return bad("INVALID_POLICY", "source.type")
	}
	if src.Calc != "value" && src.Calc != "count" {
		return bad("INVALID_POLICY", "source.calc")
	}
	if src.Calc == "count" && b.Operator != "incr" {
		return bad("LEADERBOARD_COUNT_NEEDS_INCR", "source.calc")
	}
	if src.Multiplier <= 0 || src.Multiplier > 1e6 || math.IsNaN(src.Multiplier) {
		return bad("VALUE_OUT_OF_RANGE", "source.multiplier")
	}
	switch src.Type {
	case "economy_event", "game_event":
		if !eventPattern.MatchString(src.Event) {
			return bad("INVALID_KEY", "source.event")
		}
	default:
		src.Event, src.Context, src.Field = "", "", ""
	}
	if src.Type != "game_event" {
		src.Context, src.Field = "", ""
	}
	if src.Field != "" && !fieldPattern.MatchString(src.Field) {
		return bad("INVALID_FIELD", "source.field")
	}
	if len(src.Context) > 80 {
		return bad("VALUE_TOO_LONG", "source.context")
	}
	l := &b.Limits
	if src.Type == "client" {
		l.ClientSubmit = true
	}
	if l.MaxPerWrite < 0 || l.MaxPerWrite > 1e15 {
		return bad("VALUE_OUT_OF_RANGE", "limits.maxPerWrite")
	}
	if l.PerMinute < 0 || l.PerMinute > 10000 {
		return bad("VALUE_OUT_OF_RANGE", "limits.perMinute")
	}
	if l.OnExceed != "reject" && l.OnExceed != "clamp" {
		return bad("INVALID_POLICY", "limits.onExceed")
	}
	switch b.TierBasis {
	case "none":
		b.Tiers = []Tier{}
	case "score", "rank", "percent":
	default:
		return bad("INVALID_POLICY", "tierBasis")
	}
	if len(b.Tiers) > 20 {
		return bad("TOO_MANY_ROWS", "tiers")
	}
	seen := map[string]bool{}
	for i := range b.Tiers {
		t := &b.Tiers[i]
		f := "tiers[" + strconv.Itoa(i) + "]"
		t.Key, t.IconURL = strings.TrimSpace(t.Key), strings.TrimSpace(t.IconURL)
		if !settings.KeyPattern.MatchString(t.Key) {
			return bad("INVALID_KEY", f+".key")
		}
		if seen[t.Key] {
			return bad("KEY_EXISTS", f+".key")
		}
		seen[t.Key] = true
		if t.Name, err = s.cleanText(ctx, f+".name", t.Name, false, 80); err != nil {
			return err
		}
		if math.IsNaN(t.Threshold) || math.IsInf(t.Threshold, 0) {
			return bad("VALUE_OUT_OF_RANGE", f+".threshold")
		}
		switch b.TierBasis {
		case "rank":
			if t.Threshold < 1 || t.Threshold != math.Trunc(t.Threshold) || (i > 0 && t.Threshold <= b.Tiers[i-1].Threshold) {
				return bad("LEADERBOARD_TIER_ORDER", f+".threshold")
			}
		case "percent":
			if t.Threshold <= 0 || t.Threshold > 100 || (i > 0 && t.Threshold <= b.Tiers[i-1].Threshold) {
				return bad("LEADERBOARD_TIER_ORDER", f+".threshold")
			}
		case "score":
			if t.Threshold < 0 || (i > 0 && !b.better(b.Tiers[i-1].Threshold, t.Threshold)) {
				return bad("LEADERBOARD_TIER_ORDER", f+".threshold")
			}
		}
	}
	if len(b.Rewards) > 50 {
		return bad("TOO_MANY_ROWS", "rewards")
	}
	for i := range b.Rewards {
		rw := &b.Rewards[i]
		f := "rewards[" + strconv.Itoa(i) + "]"
		rw.Tier = strings.TrimSpace(rw.Tier)
		if rw.Tier != "" && !seen[rw.Tier] {
			return bad("LEADERBOARD_TIER_NOT_FOUND", f+".tier")
		}
		if rw.FromRank < 1 {
			rw.FromRank = 1
		}
		if rw.ToRank < 0 || (rw.ToRank > 0 && rw.ToRank < rw.FromRank) || (rw.ToRank == 0 && rw.Tier == "") || rw.ToRank > 1000000 {
			return bad("LEADERBOARD_INVALID_RANKS", f+".toRank")
		}
		clean := map[string]int64{}
		for k, v := range rw.Currencies {
			k = strings.TrimSpace(k)
			if !settings.KeyPattern.MatchString(k) {
				return bad("INVALID_KEY", f+".currencies")
			}
			if v < 0 || v > 1e12 {
				return bad("VALUE_OUT_OF_RANGE", f+".currencies."+k)
			}
			if v > 0 {
				clean[k] = v
			}
		}
		if len(clean) > 20 {
			return bad("TOO_MANY_CURRENCIES", f+".currencies")
		}
		rw.Currencies = clean
		rw.PrizeKey, rw.PrizeGroupKey = strings.TrimSpace(rw.PrizeKey), strings.TrimSpace(rw.PrizeGroupKey)
		if rw.PrizeKey != "" && !s.keyExists(ctx, "economy_prize", rw.PrizeKey) {
			return bad("LEADERBOARD_PRIZE_NOT_FOUND", f+".prizeKey")
		}
		if rw.PrizeGroupKey != "" && !s.keyExists(ctx, "economy_prize_group", rw.PrizeGroupKey) {
			return bad("LEADERBOARD_PRIZE_NOT_FOUND", f+".prizeGroupKey")
		}
		if rw.empty() {
			return bad("LEADERBOARD_EMPTY_REWARD", f)
		}
	}
	if b.RewardMinScore < 0 {
		return bad("VALUE_OUT_OF_RANGE", "rewardMinScore")
	}
	if b.PayMode != "" && b.PayMode != settings.PayAuto && b.PayMode != settings.PayClaim {
		return bad("INVALID_POLICY", "payMode")
	}
	if b.Metadata == nil {
		b.Metadata = map[string]any{}
	}
	if len(mustJSON(b.Metadata)) > 8000 {
		return bad("VALUE_TOO_LONG", "metadata")
	}
	return nil
}

func (s *service) keyExists(ctx context.Context, table, key string) bool {
	var ok bool
	if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM `+table+` WHERE key=$1)`, key).Scan(&ok); err != nil {

		return true
	}
	return ok
}

type coreRow struct {
	SortOrder string
	Operator  string
	Schedule  string
	Auth      bool
	Tourney   bool
}

func (s *service) coreState(ctx context.Context, key string) (*coreRow, error) {
	var so, op int
	var c coreRow
	var dur int
	err := s.db.QueryRowContext(ctx, `SELECT sort_order, operator, COALESCE(reset_schedule,''), authoritative, duration FROM leaderboard WHERE id=$1`, key).Scan(&so, &op, &c.Schedule, &c.Auth, &dur)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.SortOrder = map[int]string{0: "asc", 1: "desc"}[so]
	c.Operator = map[int]string{0: "best", 1: "set", 2: "incr", 3: "decr"}[op]
	c.Tourney = dur > 0
	return &c, nil
}

func (s *service) currentRecords(ctx context.Context, key string) int {
	var n int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM leaderboard_record WHERE leaderboard_id=$1 AND (expiry_time = $2 OR expiry_time > now())`, key, epoch).Scan(&n)
	return n
}

func (s *service) applyCore(ctx context.Context, b *board, restart, payOnRestart bool, by string) error {
	cur, err := s.coreState(ctx, b.Key)
	if err != nil {
		return err
	}
	meta := map[string]any{"module": owner}
	if cur == nil {
		return s.core.Create(ctx, b.Key, b.SortOrder, b.Operator, b.ResetSchedule, meta)
	}
	if cur.Tourney {
		return bad("LEADERBOARD_KEY_IS_TOURNAMENT", "key")
	}
	if cur.SortOrder == b.SortOrder && cur.Operator == b.Operator && cur.Schedule == b.ResetSchedule && cur.Auth {
		return nil
	}
	if n := s.currentRecords(ctx, b.Key); n > 0 {
		if !restart {
			return &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "LEADERBOARD_RESTART_REQUIRED", Field: "resetSchedule", Meta: map[string]any{"players": n}}
		}
		if b.ID != "" {
			_ = s.closeDue(ctx)
			old := *b
			old.SortOrder, old.Operator, old.ResetSchedule = cur.SortOrder, cur.Operator, cur.Schedule
			now := s.now().UTC()
			start, _ := periodBounds(cur.Schedule, now)
			if _, err := s.closePeriod(ctx, &old, expiry(cur.Schedule, now), now, start, "restart", by, payOnRestart); err != nil {
				return err
			}
		}
	} else {
		_ = s.closeDue(ctx)
	}
	if err := s.core.Delete(ctx, b.Key); err != nil {
		return err
	}
	return s.core.Create(ctx, b.Key, b.SortOrder, b.Operator, b.ResetSchedule, meta)
}

func (s *service) boardList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := s.boards(ctx, "")
	if err != nil {
		s.fail(w, err)
		return
	}
	now := s.now()
	items := make([]map[string]any, 0, len(list))
	for _, b := range list {
		items = append(items, s.boardOut(ctx, b, now))
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "serverTime": now.UTC().Format(time.RFC3339)})
}

func (s *service) boardOut(ctx context.Context, b *board, now time.Time) map[string]any {
	start, end := periodBounds(b.ResetSchedule, now)
	var last any
	var pid string
	var pend time.Time
	var players, rewarded int
	if err := s.db.QueryRowContext(ctx, `SELECT id::text, period_end, players, rewarded FROM leaderboardx_period WHERE board_key=$1 ORDER BY period_end DESC LIMIT 1`, b.Key).Scan(&pid, &pend, &players, &rewarded); err == nil {
		last = map[string]any{"periodId": pid, "periodEnd": pend.UTC().Format(time.RFC3339), "players": players, "rewarded": rewarded}
	}
	core, _ := s.coreState(ctx, b.Key)
	return map[string]any{"board": b, "status": b.status(now), "periodStart": iso(start), "periodEnd": iso(end),
		"players": countPlayers(ctx, s.db, b.Key, expiry(b.ResetSchedule, now)), "lastPeriod": last, "coreFound": core != nil}
}

func (s *service) boardGet(w http.ResponseWriter, r *http.Request) {
	b, err := s.boardByID(r.Context(), pluginkit.PathVar(r, "id"))
	if errors.Is(err, errNoBoard) {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, s.boardOut(r.Context(), b, s.now()))
}

func (s *service) boardSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var in struct {
		board
		Restart      bool `json:"restart"`
		PayOnRestart bool `json:"payOnRestart"`
	}
	in.board = board{Active: true, SortOrder: "desc", Operator: "incr", Period: "none", TierBasis: "none", Source: Source{Type: "server", Calc: "value", Multiplier: 1}, Limits: Limits{OnExceed: "reject"}}
	if err := pluginkit.ReadBody(r, &in, 256*1024); err != nil {
		s.fail(w, err)
		return
	}
	b := &in.board
	id := pluginkit.PathVar(r, "id")
	creating := id == ""
	var prev *board
	if !creating {
		var err error
		if prev, err = s.boardByID(ctx, id); errors.Is(err, errNoBoard) {
			s.fail(w, pluginkit.Missing("NOT_FOUND"))
			return
		} else if err != nil {
			s.fail(w, err)
			return
		}
		b.Key, b.ID = prev.Key, prev.ID
	}
	b.normalize()
	if err := s.cleanBoard(ctx, b); err != nil {
		s.fail(w, err)
		return
	}
	by := pluginkit.ConsoleUser(r)
	if creating {
		var exists bool
		_ = s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM leaderboardx_board WHERE key=$1)`, b.Key).Scan(&exists)
		if exists {
			s.fail(w, bad("KEY_EXISTS", "key"))
			return
		}
	}
	if err := s.applyCore(ctx, b, in.Restart, in.PayOnRestart, by); err != nil {
		s.fail(w, err)
		return
	}
	args := []any{b.Key, mustJSON(b.Name), mustJSON(b.Description), b.IconURL, b.Sort, b.Active, b.SortOrder, b.Operator, b.Period, b.ResetSchedule,
		b.StartsAt, b.EndsAt, b.MaxPeriods, mustJSON(b.Source), mustJSON(b.Limits), b.TierBasis, mustJSON(b.Tiers), mustJSON(b.Rewards), b.RewardMinScore, b.PayMode, mustJSON(b.Metadata), by}
	var err error
	if creating {
		b.ID = uuid.Must(uuid.NewV4()).String()
		_, err = s.db.ExecContext(ctx, `INSERT INTO leaderboardx_board (key, name, description, icon_url, sort, is_active, sort_order, operator, period, reset_schedule,
			starts_at, ends_at, max_periods, source, limits, tier_basis, tiers, rewards, reward_min_score, pay_mode, metadata, updated_by, id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`, append(args, b.ID)...)
	} else {

		_, err = s.db.ExecContext(ctx, `UPDATE leaderboardx_board SET key=$1, name=$2, description=$3, icon_url=$4, sort=$5, is_active=$6, sort_order=$7, operator=$8, period=$9, reset_schedule=$10,
			starts_at=$11, ends_at=$12, max_periods=$13, source=$14, limits=$15, tier_basis=$16, tiers=$17, rewards=$18, reward_min_score=$19, pay_mode=$20, metadata=$21, updated_by=$22, updated_at=now(),
			finished_at = CASE WHEN ($12::timestamptz IS NULL OR $12::timestamptz > now()) AND ($13 = 0 OR periods_closed < $13) THEN NULL ELSE finished_at END
			WHERE id=$23`, append(args, b.ID)...)
	}
	if err != nil {
		var pg interface{ SQLState() string }
		if errors.As(err, &pg) && pg.SQLState() == "23505" {
			s.fail(w, bad("KEY_EXISTS", "key"))
			return
		}
		s.fail(w, err)
		return
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE leaderboardx_board SET scope=$2 WHERE id=$1`, b.ID, b.Scope); err != nil {
		s.fail(w, err)
		return
	}
	invalidateBoards()
	saved, err := s.boardByID(ctx, b.ID)
	if err != nil {
		s.fail(w, err)
		return
	}
	status := http.StatusOK
	if creating {
		status = http.StatusCreated
	}
	pluginkit.WriteJSON(w, status, s.boardOut(ctx, saved, s.now()))
}

func (s *service) boardDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	b, err := s.boardByID(ctx, pluginkit.PathVar(r, "id"))
	if errors.Is(err, errNoBoard) {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	var open int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM leaderboardx_result r JOIN leaderboardx_period p ON p.id=r.period_id WHERE p.board_key=$1 AND r.status IN ('pending','claimable')`, b.Key).Scan(&open)
	if open > 0 {
		s.fail(w, &pluginkit.ConsoleError{Status: http.StatusConflict, Code: "IN_USE", Message: "rewards of this board are not settled", Meta: map[string]any{"open": open}})
		return
	}
	if r.URL.Query().Get("core") != "keep" {
		if err := s.core.Delete(ctx, b.Key); err != nil {
			s.logger.Warn("Leaderboard %s: core leaderboard not deleted: %v", b.Key, err)
		}
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM leaderboardx_board WHERE id=$1`, b.ID); err != nil {
		s.fail(w, err)
		return
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM leaderboardx_score_log WHERE board_id=$1`, b.ID)
	invalidateBoards()
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": b.Key})
}

func (s *service) standings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	b, err := s.boardByID(ctx, pluginkit.PathVar(r, "id"))
	if errors.Is(err, errNoBoard) {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	now := s.now()
	exp := expiry(b.ResetSchedule, now)
	total := countPlayers(ctx, s.db, b.Key, exp)
	cfg := s.settings(ctx)
	cfg.Display.ShowDisplayName = true
	out := map[string]any{"board": b, "players": total, "status": b.status(now), "items": []any{}, "nextCursor": ""}
	start, end := periodBounds(b.ResetSchedule, now)
	out["periodStart"], out["periodEnd"] = iso(start), iso(end)
	if find := strings.TrimSpace(q.Get("q")); find != "" {
		var uid string
		err := s.db.QueryRowContext(ctx, `SELECT id::text FROM users WHERE id::text=$1 OR lower(username)=lower($1) LIMIT 1`, find).Scan(&uid)
		if err != nil {
			pluginkit.WriteJSON(w, http.StatusOK, out)
			return
		}
		hay, err := s.core.Haystack(ctx, b.Key, uid, 5)
		if err == nil {
			out["items"] = s.rows(ctx, b, hay, uid, total, cfg)
		}
		pluginkit.WriteJSON(w, http.StatusOK, out)
		return
	}
	recs, _, next, err := s.core.List(ctx, b.Key, nil, limit, q.Get("cursor"))
	if err != nil {
		s.fail(w, bad("LEADERBOARD_INVALID_CURSOR", "cursor"))
		return
	}
	out["items"], out["nextCursor"] = s.rows(ctx, b, recs, "", total, cfg), next
	pluginkit.WriteJSON(w, http.StatusOK, out)
}

func (s *service) recordSet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	b, err := s.boardByID(ctx, pluginkit.PathVar(r, "id"))
	if errors.Is(err, errNoBoard) {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	uid := pluginkit.PathVar(r, "userId")
	var in struct {
		Score    int64  `json:"score"`
		Subscore int64  `json:"subscore"`
		Mode     string `json:"mode"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	if !uuidPattern.MatchString(uid) {
		s.fail(w, bad("INVALID_ID", "userId"))
		return
	}
	var username string
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(username,'') FROM users WHERE id::text=$1`, uid).Scan(&username); err != nil {
		s.fail(w, pluginkit.Missing("USER_NOT_FOUND"))
		return
	}
	if in.Score < 0 || in.Subscore < 0 || in.Score > 1e15 {
		s.fail(w, bad("VALUE_OUT_OF_RANGE", "score"))
		return
	}
	op := int(2)
	if in.Mode == "add" {
		op = 3
	} else if in.Mode != "" && in.Mode != "set" {
		s.fail(w, bad("INVALID_POLICY", "mode"))
		return
	}
	rec, err := s.core.Write(ctx, b.Key, uid, username, in.Score, in.Subscore, &op)
	if err != nil {
		s.fail(w, err)
		return
	}
	by := pluginkit.ConsoleUser(r)
	_, _ = s.db.ExecContext(ctx, `INSERT INTO leaderboardx_score_log (board_id, source_id, user_id, source, value, delta, outcome) VALUES ($1,$2,$3,'admin',$4,$4,$5)`,
		b.ID, "admin:"+uuid.Must(uuid.NewV4()).String(), uid, in.Score, in.Mode)
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: uid, Kind: activity.KindConfig, Action: "leaderboard_score_set", RefID: b.Key, Actor: by,
		Data: map[string]any{"boardKey": b.Key, "score": in.Score, "mode": in.Mode}})
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"score": rec.GetScore(), "subscore": rec.GetSubscore(), "rank": rec.GetRank()})
}

func (s *service) recordDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	b, err := s.boardByID(ctx, pluginkit.PathVar(r, "id"))
	if errors.Is(err, errNoBoard) {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	uid := pluginkit.PathVar(r, "userId")
	if !uuidPattern.MatchString(uid) {
		s.fail(w, bad("INVALID_ID", "userId"))
		return
	}
	if err := s.core.RecordDelete(ctx, b.Key, uid); err != nil {
		s.fail(w, err)
		return
	}
	activity.Safe(ctx, s.db, s.logger, activity.Entry{UserID: uid, Kind: activity.KindConfig, Action: "leaderboard_score_removed", RefID: b.Key, Actor: pluginkit.ConsoleUser(r),
		Data: map[string]any{"boardKey": b.Key}})
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

func (s *service) boardClose(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	b, err := s.boardByID(ctx, pluginkit.PathVar(r, "id"))
	if errors.Is(err, errNoBoard) {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	var in struct {
		Pay bool `json:"pay"`
	}
	if err := pluginkit.ReadBody(r, &in, 0); err != nil {
		s.fail(w, err)
		return
	}
	_ = s.closeDue(ctx)
	now := s.now().UTC()
	start, _ := periodBounds(b.ResetSchedule, now)
	var last time.Time
	if err := s.db.QueryRowContext(ctx, `SELECT period_end FROM leaderboardx_period WHERE board_key=$1 ORDER BY period_end DESC LIMIT 1`, b.Key).Scan(&last); err == nil && (start == nil || last.After(*start)) {
		start = &last
	}
	pid, err := s.closePeriod(ctx, b, expiry(b.ResetSchedule, now), now, start, "manual", pluginkit.ConsoleUser(r), in.Pay)
	if err != nil {
		s.fail(w, err)
		return
	}
	if err := s.core.Delete(ctx, b.Key); err != nil {
		s.fail(w, err)
		return
	}
	if err := s.core.Create(ctx, b.Key, b.SortOrder, b.Operator, b.ResetSchedule, map[string]any{"module": owner}); err != nil {
		s.fail(w, err)
		return
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"periodId": pid})
}

func (s *service) periodList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, limit := pluginkit.Page(r, 20, 100)
	var a pluginkit.Args
	where := []string{"true"}
	if k := strings.TrimSpace(r.URL.Query().Get("boardKey")); k != "" {
		where = append(where, "p.board_key="+a.Add(k))
	}
	w2 := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM leaderboardx_period p WHERE `+w2, a.List...).Scan(&total); err != nil {
		s.fail(w, err)
		return
	}
	lim, off := a.Add(limit), a.Add((page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT p.id::text, p.board_key, p.board_name::text, p.period_start, p.period_end, p.reason, p.players, p.rewarded, p.pay_mode, p.created_by, p.created_at,
		count(*) FILTER (WHERE r.status='paid'), count(*) FILTER (WHERE r.status='pending'), count(*) FILTER (WHERE r.status='claimable'),
		count(*) FILTER (WHERE r.status='failed'), count(*) FILTER (WHERE r.status='expired')
		FROM leaderboardx_period p LEFT JOIN leaderboardx_result r ON r.period_id=p.id AND r.status <> 'none'
		WHERE `+w2+` GROUP BY p.id ORDER BY p.period_end DESC LIMIT `+lim+` OFFSET `+off, a.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var id, key, rawName, reason, mode, by string
		var start *time.Time
		var end, created time.Time
		var players, rewarded, paid, pending, claimable, failed, expired int
		if err := rows.Scan(&id, &key, &rawName, &start, &end, &reason, &players, &rewarded, &mode, &by, &created, &paid, &pending, &claimable, &failed, &expired); err != nil {
			s.fail(w, err)
			return
		}
		name := map[string]string{}
		unmarshal([]byte(rawName), &name)
		items = append(items, map[string]any{"id": id, "boardKey": key, "boardName": name, "periodStart": iso(start), "periodEnd": iso(&end), "reason": reason, "players": players,
			"rewarded": rewarded, "payMode": mode, "createdBy": by, "createdAt": created.UTC().Format(time.RFC3339),
			"counts": map[string]int{"paid": paid, "pending": pending, "claimable": claimable, "failed": failed, "expired": expired}})
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) periodGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	if !uuidPattern.MatchString(id) {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	var key, rawName, reason, mode, rawConf string
	var start *time.Time
	var end time.Time
	var players, rewarded int
	err := s.db.QueryRowContext(ctx, `SELECT board_key, board_name::text, period_start, period_end, reason, players, rewarded, pay_mode, config::text FROM leaderboardx_period WHERE id::text=$1`, id).
		Scan(&key, &rawName, &start, &end, &reason, &players, &rewarded, &mode, &rawConf)
	if errors.Is(err, sql.ErrNoRows) {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	if err != nil {
		s.fail(w, err)
		return
	}
	page, limit := pluginkit.Page(r, 50, 500)
	var a pluginkit.Args
	where := []string{"r.period_id=" + a.Add(id)}
	q := r.URL.Query()
	switch st := q.Get("status"); st {
	case "":
	case "rewarded":
		where = append(where, "r.status <> 'none'")
	default:
		where = append(where, "r.status="+a.Add(st))
	}
	if find := strings.TrimSpace(q.Get("q")); find != "" {
		where = append(where, "(r.user_id::text="+a.Add(find)+" OR r.username ILIKE "+a.Add(pluginkit.LikePattern(find))+")")
	}
	w2 := strings.Join(where, " AND ")
	var total int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM leaderboardx_result r WHERE `+w2, a.List...).Scan(&total)
	lim, off := a.Add(limit), a.Add((page-1)*limit)
	rows, err := s.db.QueryContext(ctx, `SELECT r.user_id::text, r.rank, r.tier_rank, r.tier, r.score, r.subscore, COALESCE(NULLIF(u.username,''), r.username), r.status, r.plan, r.paid, r.claim_expires_at, r.paid_at, r.last_error, r.pay_tries
		FROM leaderboardx_result r LEFT JOIN users u ON u.id=r.user_id WHERE `+w2+` ORDER BY r.rank LIMIT `+lim+` OFFSET `+off, a.List...)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var uid, tier, uname, status, lastErr string
		var rank, tierRank, tries int
		var score, sub int64
		var rawPlan, rawPaid []byte
		var claimUntil, paidAt *time.Time
		if err := rows.Scan(&uid, &rank, &tierRank, &tier, &score, &sub, &uname, &status, &rawPlan, &rawPaid, &claimUntil, &paidAt, &lastErr, &tries); err != nil {
			s.fail(w, err)
			return
		}
		items = append(items, resultRow(id, uid, rank, tierRank, tier, score, sub, uname, status, rawPlan, rawPaid, claimUntil, paidAt, lastErr, tries))
	}
	name := map[string]string{}
	unmarshal([]byte(rawName), &name)
	conf := map[string]any{}
	unmarshal([]byte(rawConf), &conf)
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"period": map[string]any{"id": id, "boardKey": key, "boardName": name, "periodStart": iso(start), "periodEnd": iso(&end),
		"reason": reason, "players": players, "rewarded": rewarded, "payMode": mode, "config": conf}, "items": items, "total": total, "page": page, "limit": limit})
}

func (s *service) periodRetry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := pluginkit.PathVar(r, "id")
	if !uuidPattern.MatchString(id) {
		s.fail(w, pluginkit.Missing("NOT_FOUND"))
		return
	}
	res, err := s.db.ExecContext(ctx, `UPDATE leaderboardx_result SET status='pending', pay_tries=0 WHERE period_id::text=$1 AND status='failed'`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	n, _ := res.RowsAffected()
	rows, err := s.db.QueryContext(ctx, `SELECT user_id::text FROM leaderboardx_result WHERE period_id::text=$1 AND status='pending' LIMIT 1000`, id)
	if err != nil {
		s.fail(w, err)
		return
	}
	var users []string
	for rows.Next() {
		var u string
		if rows.Scan(&u) == nil {
			users = append(users, u)
		}
	}
	rows.Close()
	paid := 0
	for _, u := range users {
		if out, err := s.payResult(ctx, id, u, false); err == nil && out["status"] == "paid" {
			paid++
		}
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"requeued": n, "paid": paid})
}

func (s *service) playerGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uid := pluginkit.PathVar(r, "id")
	if !uuidPattern.MatchString(uid) {
		s.fail(w, bad("INVALID_ID", "id"))
		return
	}
	list, err := s.boards(ctx, "")
	if err != nil {
		s.fail(w, err)
		return
	}
	now := s.now()
	cfg := s.settings(ctx)
	current := []map[string]any{}
	for _, b := range list {
		_, own, _, err := s.core.List(ctx, b.Key, []string{uid}, 1, "")
		if err != nil || len(own) == 0 {
			continue
		}
		me := s.meRow(ctx, b, own[0])
		me["boardId"], me["boardKey"], me["name"], me["status"] = b.ID, b.Key, b.Name, b.status(now)
		current = append(current, me)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.id::text, p.board_key, p.board_name::text, p.period_end, r.rank, r.tier_rank, r.tier, r.score, r.subscore, r.username, r.status, r.plan, r.paid, r.claim_expires_at, r.paid_at, r.last_error, r.pay_tries
		FROM leaderboardx_result r JOIN leaderboardx_period p ON p.id=r.period_id WHERE r.user_id::text=$1 ORDER BY p.period_end DESC LIMIT 30`, uid)
	if err != nil {
		s.fail(w, err)
		return
	}
	defer rows.Close()
	history := []map[string]any{}
	for rows.Next() {
		var pid, key, rawName, tier, uname, status, lastErr string
		var end time.Time
		var rank, tierRank, tries int
		var score, sub int64
		var rawPlan, rawPaid []byte
		var claimUntil, paidAt *time.Time
		if err := rows.Scan(&pid, &key, &rawName, &end, &rank, &tierRank, &tier, &score, &sub, &uname, &status, &rawPlan, &rawPaid, &claimUntil, &paidAt, &lastErr, &tries); err != nil {
			s.fail(w, err)
			return
		}
		row := resultRow(pid, uid, rank, tierRank, tier, score, sub, uname, status, rawPlan, rawPaid, claimUntil, paidAt, lastErr, tries)
		name := map[string]string{}
		unmarshal([]byte(rawName), &name)
		row["boardKey"], row["boardName"], row["periodEnd"] = key, name, end.UTC().Format(time.RFC3339)
		history = append(history, row)
	}
	var refused int
	_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM leaderboardx_score_log WHERE user_id::text=$1 AND outcome IN ('too_high','rate_limited') AND created_at > now() - interval '30 days'`, uid).Scan(&refused)
	_ = cfg
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"current": current, "history": history, "refused30d": refused})
}

func (s *service) report(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days != 7 && days != 30 && days != 90 {
		days = 30
	}
	list, err := s.boards(ctx, "")
	if err != nil {
		s.fail(w, err)
		return
	}
	now := s.now()
	type agg struct{ writes, refused, writers int }
	byBoard := map[string]agg{}
	rows, err := s.db.QueryContext(ctx, `SELECT board_id::text, count(*) FILTER (WHERE outcome IN ('','clamped','set','add')), count(*) FILTER (WHERE outcome IN ('too_high','rate_limited')), count(DISTINCT user_id)
		FROM leaderboardx_score_log WHERE created_at > now() - make_interval(days => $1) GROUP BY 1`, days)
	if err == nil {
		for rows.Next() {
			var id string
			var a agg
			if rows.Scan(&id, &a.writes, &a.refused, &a.writers) == nil {
				byBoard[id] = a
			}
		}
		rows.Close()
	}
	type rw struct{ paid, pending, claimable, failed, expired int }
	rewards := map[string]rw{}
	rows, err = s.db.QueryContext(ctx, `SELECT p.board_key, count(*) FILTER (WHERE r.status='paid'), count(*) FILTER (WHERE r.status='pending'), count(*) FILTER (WHERE r.status='claimable'),
		count(*) FILTER (WHERE r.status='failed'), count(*) FILTER (WHERE r.status='expired')
		FROM leaderboardx_result r JOIN leaderboardx_period p ON p.id=r.period_id WHERE p.period_end > now() - make_interval(days => $1) AND r.status <> 'none' GROUP BY 1`, days)
	if err == nil {
		for rows.Next() {
			var k string
			var x rw
			if rows.Scan(&k, &x.paid, &x.pending, &x.claimable, &x.failed, &x.expired) == nil {
				rewards[k] = x
			}
		}
		rows.Close()
	}
	items := []map[string]any{}
	totals := map[string]int{"boards": 0, "players": 0, "writes": 0, "refused": 0, "paid": 0, "open": 0, "failed": 0}
	for _, b := range list {
		a, x := byBoard[b.ID], rewards[b.Key]
		players := countPlayers(ctx, s.db, b.Key, expiry(b.ResetSchedule, now))
		var periods int
		_ = s.db.QueryRowContext(ctx, `SELECT count(*) FROM leaderboardx_period WHERE board_key=$1 AND period_end > now() - make_interval(days => $2)`, b.Key, days).Scan(&periods)
		items = append(items, map[string]any{"id": b.ID, "key": b.Key, "name": b.Name, "status": b.status(now), "players": players, "writes": a.writes, "refused": a.refused, "writers": a.writers,
			"periods": periods, "rewards": map[string]int{"paid": x.paid, "pending": x.pending, "claimable": x.claimable, "failed": x.failed, "expired": x.expired}})
		if b.status(now) == "active" {
			totals["boards"]++
		}
		totals["players"] += players
		totals["writes"] += a.writes
		totals["refused"] += a.refused
		totals["paid"] += x.paid
		totals["open"] += x.pending + x.claimable
		totals["failed"] += x.failed
	}
	daily := []map[string]any{}
	rows, err = s.db.QueryContext(ctx, `SELECT to_char(date_trunc('day', created_at AT TIME ZONE 'UTC'), 'YYYY-MM-DD'), count(*) FILTER (WHERE outcome IN ('','clamped','set','add')), count(*) FILTER (WHERE outcome IN ('too_high','rate_limited'))
		FROM leaderboardx_score_log WHERE created_at > now() - make_interval(days => $1) GROUP BY 1 ORDER BY 1`, days)
	if err == nil {
		for rows.Next() {
			var d string
			var wr, rf int
			if rows.Scan(&d, &wr, &rf) == nil {
				daily = append(daily, map[string]any{"day": d, "writes": wr, "refused": rf})
			}
		}
		rows.Close()
	}
	pluginkit.WriteJSON(w, http.StatusOK, map[string]any{"days": days, "items": items, "totals": totals, "daily": daily})
}

var _ = apierr.BadRequest
