package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
)

type incoming struct {
	Type    string
	ID      string
	UserID  string
	Name    string
	Context string
	Value   float64
	Fields  map[string]any
}

func (s *service) onEconomyEvent(ctx context.Context, ev playerevents.Recorded) {
	s.handle(ctx, incoming{Type: "economy_event", ID: "ev:" + ev.ID, UserID: ev.UserID, Name: ev.Key, Value: ev.Value, Fields: ev.Metadata})
}

func (s *service) onGameEvent(ctx context.Context, ev gameevents.Event) (any, error) {
	if strings.HasPrefix(ev.ID, owner+":") {
		return nil, nil
	}
	if ev.Name == "trophy_changed" {
		if m, _ := s.memberOf(ctx, s.db, ev.UserID); m != nil {
			s.refreshTrophies(ctx, m.ClanID)
		}
	}
	value := 1.0
	if v, ok := number(ev.Fields["value"]); ok {
		value = v
	}
	s.handle(ctx, incoming{Type: "game_event", ID: "ge:" + ev.ID, UserID: ev.UserID, Name: ev.Name, Context: ev.Context, Value: value, Fields: ev.Fields})
	return nil, nil
}

func (s *service) handle(ctx context.Context, in incoming) {
	if in.UserID == "" || in.Name == "" {
		return
	}
	if in.Fields == nil {
		in.Fields = map[string]any{}
	}
	cfg := s.settings(ctx)
	if !cfg.Enabled {
		return
	}
	m, err := s.memberOf(ctx, s.db, in.UserID)
	if err != nil || m == nil {
		return
	}
	for i, src := range cfg.XP.Sources {
		if src.Type != in.Type || src.Event != in.Name || (src.Context != "" && src.Context != in.Context) {
			continue
		}
		amount := xpFor(src, in.Value, in.Fields)
		if amount <= 0 {
			continue
		}
		if _, err := s.addXP(ctx, m.ClanID, in.UserID, fmt.Sprintf("%s:%d", in.ID, i), in.Name, amount); err != nil {
			s.logger.Warn("Clan XP for %s (%s) not added: %v", m.ClanID, in.ID, err)
		}
	}
	s.scoreWars(ctx, in)
}

type xpResult struct {
	Added     int64
	XP        int64
	Level     int
	LevelFrom int
	Duplicate bool
	Capped    bool
}

func (s *service) addXP(ctx context.Context, clanID, userID, sourceID, source string, amount int64) (*xpResult, error) {
	cfg := s.settings(ctx)
	r := s.road(ctx)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	out, err := s.addXPTx(ctx, tx, cfg, r, clanID, userID, sourceID, source, amount)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.afterXP(ctx, cfg, r, clanID, userID, out)
	return out, nil
}

func (s *service) addXPTx(ctx context.Context, tx *sql.Tx, cfg settings.All, r road, clanID, userID, sourceID, source string, amount int64) (*xpResult, error) {
	var xp int64
	var lv int
	if err := tx.QueryRowContext(ctx, `SELECT xp, level FROM clan WHERE group_id::text=$1 FOR UPDATE`, clanID).Scan(&xp, &lv); err != nil {
		return nil, err
	}
	if userID != "" && cfg.XP.MemberDailyCap > 0 {
		var today int64
		var day sql.NullTime
		_ = tx.QueryRowContext(ctx, `SELECT xp_day, xp_today FROM clan_member WHERE group_id::text=$1 AND user_id::text=$2`, clanID, userID).Scan(&day, &today)
		if !day.Valid || day.Time.Format("2006-01-02") != s.now().UTC().Format("2006-01-02") {
			today = 0
		}
		amount = clipDaily(cfg.XP.MemberDailyCap, today, amount)
		if amount <= 0 {
			return &xpResult{XP: xp, Level: lv, LevelFrom: lv, Capped: true}, nil
		}
	}
	var uid any
	if userID != "" {
		uid = userID
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO clan_xp_log (group_id, user_id, source_id, source, amount) VALUES ($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, clanID, uid, sourceID, source, amount)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &xpResult{XP: xp, Level: lv, LevelFrom: lv, Duplicate: true}, nil
	}
	if userID != "" {
		if _, err := tx.ExecContext(ctx, `UPDATE clan_member SET xp=xp+$3, xp_today=CASE WHEN xp_day = (now() AT TIME ZONE 'UTC')::date THEN xp_today+$3 ELSE $3 END,
			xp_day=(now() AT TIME ZONE 'UTC')::date WHERE group_id::text=$1 AND user_id::text=$2`, clanID, userID, amount); err != nil {
			return nil, err
		}
	}
	newXP := xp + amount
	newLv := r.levelFor(newXP)
	if newLv < lv {
		newLv = lv
	}
	if _, err := tx.ExecContext(ctx, `UPDATE clan SET xp=$2, level=$3, updated_at=now() WHERE group_id::text=$1`, clanID, newXP, newLv); err != nil {
		return nil, err
	}
	out := &xpResult{Added: amount, XP: newXP, Level: newLv, LevelFrom: lv}
	if newLv <= lv {
		return out, nil
	}
	members := []string{}
	rows, err := tx.QueryContext(ctx, `SELECT user_id::text FROM clan_member WHERE group_id::text=$1`, clanID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			members = append(members, id)
		}
	}
	rows.Close()
	for _, l := range r {
		if l.Level <= lv || l.Level > newLv {
			continue
		}
		res, err := tx.ExecContext(ctx, `INSERT INTO clan_level_reached (group_id, level, members) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, clanID, l.Level, len(members))
		if err != nil {
			return nil, err
		}
		if n, _ := res.RowsAffected(); n == 0 || l.PrizeKey == "" {
			continue
		}
		for _, m := range members {
			if _, err := tx.ExecContext(ctx, `INSERT INTO clan_payout (id, source_id, kind, group_id, user_id, ref, prize_key) VALUES ($1,$2,'level',$3,$4,$5,$6) ON CONFLICT DO NOTHING`,
				uuid.Must(uuid.NewV4()).String(), fmt.Sprintf("level:%s:%d:%s", clanID, l.Level, m), clanID, m, fmt.Sprint(l.Level), l.PrizeKey); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func (s *service) afterXP(ctx context.Context, cfg settings.All, r road, clanID, userID string, out *xpResult) {
	if out == nil || out.Added <= 0 {
		return
	}
	if cfg.Events.Publish {
		gameevents.Publish(ctx, gameevents.Event{ID: owner + ":x:" + uuid.Must(uuid.NewV4()).String(), UserID: clanID, Name: "clan_xp", Source: owner,
			Fields: map[string]any{"xp": float64(out.XP), "delta": float64(out.Added), "level": float64(out.Level)}})
		if userID != "" {
			gameevents.Publish(ctx, gameevents.Event{ID: owner + ":mx:" + uuid.Must(uuid.NewV4()).String(), UserID: userID, Name: "clan_member_xp", Context: clanID, Source: owner,
				Fields: map[string]any{"xp": float64(out.Added), "clanId": clanID}})
		}
	}
	if out.Level > out.LevelFrom {
		s.levelUp(ctx, cfg, r, clanID, out.LevelFrom, out.Level)
	}
}

func (s *service) levelUp(ctx context.Context, cfg settings.All, r road, clanID string, from, to int) {
	cap := r.memberCap(to, 0)
	if cap > 0 {
		var join string
		_ = s.db.QueryRowContext(ctx, `SELECT join_type FROM clan WHERE group_id::text=$1`, clanID).Scan(&join)
		if err := s.groups.Update(ctx, clanID, "", "", "", "", join == settings.JoinOpen, cap, nil); err != nil {
			s.logger.Warn("Clan %s member cap not raised to %d: %v", clanID, cap, err)
		}
	}
	var perks map[string]any
	prize := ""
	if l := r.get(to); l != nil {
		perks, prize = l.Perks, l.PrizeKey
	}
	members := s.memberIDs(ctx, clanID)
	s.notify(ctx, members, "clan_level_up", map[string]any{"clanId": clanID, "level": to, "from": from, "memberCap": cap, "perks": perks, "prizeKey": prize}, true)
	s.say(ctx, cfg, clanID, "level", "", fmt.Sprint(to))
	s.payPending(ctx, clanID, len(members)*(to-from)+10)
}

func (s *service) payPending(ctx context.Context, clanID string, limit int) {
	for i := 0; i < limit; i++ {
		if !s.payOne(ctx, clanID) {
			return
		}
	}
}

func (s *service) payOne(ctx context.Context, clanID string) bool {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false
	}
	defer func() { _ = tx.Rollback() }()
	q := `SELECT id::text, source_id, kind, user_id::text, ref, prize_key, tries FROM clan_payout WHERE status='pending'`
	args := []any{}
	if clanID != "" {
		q += ` AND group_id::text=$1`
		args = append(args, clanID)
	}
	var id, source, kind, uid, ref, prize string
	var tries int
	if err := tx.QueryRowContext(ctx, q+` ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED`, args...).Scan(&id, &source, &kind, &uid, &ref, &prize, &tries); err != nil {
		return false
	}
	res, gerr := s.prize(ctx, economyhooks.PrizeGrant{UserID: uid, PrizeKey: prize, SourceType: sourceType, SourceID: source, Reason: "clan " + kind + " " + ref})
	if gerr != nil {
		status := "pending"
		if tries+1 >= 8 {
			status = "failed"
		}
		_, _ = tx.ExecContext(ctx, `UPDATE clan_payout SET tries=tries+1, last_error=$2, status=$3 WHERE id::text=$1`, id, truncate(gerr.Error(), 500), status)
		_ = tx.Commit()
		return false
	}
	paid, _ := json.Marshal(res.Prizes)
	if _, err := tx.ExecContext(ctx, `UPDATE clan_payout SET status='paid', tries=tries+1, paid=$2, paid_at=now(), last_error='' WHERE id::text=$1`, id, paid); err != nil {
		return false
	}
	return tx.Commit() == nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func iso(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}
