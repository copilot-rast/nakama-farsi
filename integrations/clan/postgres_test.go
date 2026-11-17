package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/playerevents"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeGroups struct {
	db   *sql.DB
	mu   sync.Mutex
	said []map[string]any
}

func (g *fakeGroups) Create(ctx context.Context, ownerID, name, lang, description, avatar string, open bool, maxCount int) (string, error) {
	id := uuid.Must(uuid.NewV4()).String()
	state := 1
	if open {
		state = 0
	}
	if lang == "" {
		lang = "en"
	}
	if _, err := g.db.ExecContext(ctx, `INSERT INTO groups (id, creator_id, name, description, avatar_url, lang_tag, state, edge_count, max_count) VALUES ($1,$2,$3,$4,$5,$6,$7,1,$8)`,
		id, ownerID, name, description, avatar, lang, state, maxCount); err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) && pg.Code == "23505" {
			return "", runtime.ErrGroupNameInUse
		}
		return "", err
	}
	_, err := g.db.ExecContext(ctx, `INSERT INTO group_edge (source_id, position, destination_id, state) VALUES ($1,$2,$3,0)`, id, time.Now().UnixNano(), ownerID)
	return id, err
}

func (g *fakeGroups) Update(ctx context.Context, id, name, lang, description, avatar string, open bool, maxCount int) error {
	state := 1
	if open {
		state = 0
	}
	_, err := g.db.ExecContext(ctx, `UPDATE groups SET name=COALESCE(NULLIF($2,''),name), state=$3, max_count=CASE WHEN $4 > 0 THEN $4 ELSE max_count END WHERE id::text=$1`, id, name, state, maxCount)
	return err
}

func (g *fakeGroups) Delete(ctx context.Context, id string) error {
	_, _ = g.db.ExecContext(ctx, `DELETE FROM group_edge WHERE source_id::text=$1`, id)
	_, err := g.db.ExecContext(ctx, `DELETE FROM groups WHERE id::text=$1`, id)
	return err
}

func (g *fakeGroups) edge(ctx context.Context, id, user string) int {
	st := -1
	_ = g.db.QueryRowContext(ctx, `SELECT state FROM group_edge WHERE source_id::text=$1 AND destination_id::text=$2`, id, user).Scan(&st)
	return st
}

func (g *fakeGroups) addState(ctx context.Context, id, user string, state int) error {
	if state < 3 {
		var count, max int
		_ = g.db.QueryRowContext(ctx, `SELECT edge_count, max_count FROM groups WHERE id::text=$1`, id).Scan(&count, &max)
		if count >= max {
			return runtime.ErrGroupFull
		}
		if _, err := g.db.ExecContext(ctx, `UPDATE groups SET edge_count=edge_count+1 WHERE id::text=$1`, id); err != nil {
			return err
		}
	}
	if cur := g.edge(ctx, id, user); cur >= 0 {
		_, err := g.db.ExecContext(ctx, `UPDATE group_edge SET state=$3 WHERE source_id::text=$1 AND destination_id::text=$2`, id, user, state)
		return err
	}
	_, err := g.db.ExecContext(ctx, `INSERT INTO group_edge (source_id, position, destination_id, state) VALUES ($1,$2,$3,$4)`, id, time.Now().UnixNano(), user, state)
	return err
}

func (g *fakeGroups) Join(ctx context.Context, id, userID, _ string) error {
	var state int
	_ = g.db.QueryRowContext(ctx, `SELECT state FROM groups WHERE id::text=$1`, id).Scan(&state)
	if state == 0 {
		return g.addState(ctx, id, userID, 2)
	}
	return g.addState(ctx, id, userID, 3)
}

func (g *fakeGroups) remove(ctx context.Context, id, user string) error {
	st := g.edge(ctx, id, user)
	if st < 0 {
		return runtime.ErrGroupUserNotFound
	}
	if st < 3 {
		_, _ = g.db.ExecContext(ctx, `UPDATE groups SET edge_count=edge_count-1 WHERE id::text=$1`, id)
	}
	_, err := g.db.ExecContext(ctx, `DELETE FROM group_edge WHERE source_id::text=$1 AND destination_id::text=$2`, id, user)
	return err
}

func (g *fakeGroups) Leave(ctx context.Context, id, userID, _ string) error { return g.remove(ctx, id, userID) }

func (g *fakeGroups) Add(ctx context.Context, id string, userIDs []string) error {
	for _, u := range userIDs {
		if err := g.addState(ctx, id, u, 2); err != nil {
			return err
		}
	}
	return nil
}

func (g *fakeGroups) Kick(ctx context.Context, id string, userIDs []string) error {
	for _, u := range userIDs {
		_ = g.remove(ctx, id, u)
	}
	return nil
}

func (g *fakeGroups) shift(ctx context.Context, id string, userIDs []string, by int) error {
	for _, u := range userIDs {
		if _, err := g.db.ExecContext(ctx, `UPDATE group_edge SET state=GREATEST(0, LEAST(2, state+$3)) WHERE source_id::text=$1 AND destination_id::text=$2`, id, u, by); err != nil {
			return err
		}
	}
	return nil
}

func (g *fakeGroups) Promote(ctx context.Context, id string, userIDs []string) error {
	return g.shift(ctx, id, userIDs, -1)
}

func (g *fakeGroups) Demote(ctx context.Context, id string, userIDs []string) error {
	return g.shift(ctx, id, userIDs, 1)
}

func (g *fakeGroups) Say(_ context.Context, id string, content map[string]any) error {
	g.mu.Lock()
	g.said = append(g.said, content)
	g.mu.Unlock()
	return nil
}

type fixture struct {
	t      *testing.T
	s      *service
	g      *fakeGroups
	mu     sync.Mutex
	grants []economyhooks.PrizeGrant
	notes  map[string][]string
	clans  []string
}

func newFixture(t *testing.T) *fixture {
	db := plugintest.OpenDB(t)
	f := &fixture{t: t, g: &fakeGroups{db: db}, notes: map[string][]string{}}
	s := newService(db, nil, plugintest.Logger{})
	s.groups = f.g
	s.debit = func(context.Context, string, string, map[string]int64) error { return nil }
	s.prize = func(_ context.Context, g economyhooks.PrizeGrant) (economyhooks.PrizeResult, error) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.grants = append(f.grants, g)
		return economyhooks.PrizeResult{}, nil
	}
	s.notify = func(_ context.Context, ids []string, subject string, _ map[string]any, _ bool) {
		f.mu.Lock()
		defer f.mu.Unlock()
		for _, id := range ids {
			f.notes[id] = append(f.notes[id], subject)
		}
	}
	f.s = s
	cfg := settings.Defaults()
	cfg.Cooldowns.JoinAfterLeaveMinutes = 0
	cfg.XP.Sources = []settings.XPSource{{Type: "economy_event", Event: "cl_won", Amount: 30}}
	if err := settings.Save(context.Background(), db, cfg, "test"); err != nil {
		t.Fatal(err)
	}
	plugintest.Exec(t, db, `DELETE FROM clan_level`)
	plugintest.Exec(t, db, `INSERT INTO clan_level (level, xp, name, member_cap) VALUES (1,0,'{}',3), (2,50,'{}',5)`)
	plugintest.Exec(t, db, `UPDATE clan_level SET prize_key='cl_lv2' WHERE level=2`)
	s.invalidate()
	t.Cleanup(func() {
		ctx := context.Background()
		for _, id := range f.clans {
			_, _ = db.ExecContext(ctx, `DELETE FROM clan_payout WHERE group_id::text=$1`, id)
			_, _ = db.ExecContext(ctx, `DELETE FROM clan_xp_log WHERE group_id::text=$1`, id)
			_, _ = db.ExecContext(ctx, `DELETE FROM clan_war WHERE clan_a::text=$1 OR clan_b::text=$1`, id)
			_, _ = db.ExecContext(ctx, `DELETE FROM clan WHERE group_id::text=$1`, id)
			_ = f.g.Delete(ctx, id)
		}
		_, _ = db.ExecContext(ctx, `DELETE FROM clan_war_round WHERE config_key LIKE 'clt_%'`)
		_, _ = db.ExecContext(ctx, `DELETE FROM clan_war_config WHERE key LIKE 'clt_%'`)
		_, _ = db.ExecContext(ctx, `DELETE FROM clan_level`)
		_, _ = db.ExecContext(ctx, `INSERT INTO clan_level (level, xp, name, member_cap) VALUES (1,0,'{"fa":"سطح ۱","en":"Level 1"}',30)`)
		_ = settings.Save(ctx, db, settings.Defaults(), "")
	})
	return f
}

func (f *fixture) user() string {
	id := plugintest.NewUser(f.t, f.s.db)
	f.t.Cleanup(func() {
		for _, q := range []string{`DELETE FROM clan_member WHERE user_id=$1`, `DELETE FROM clan_cooldown WHERE user_id=$1`, `DELETE FROM clan_request WHERE user_id=$1`, `DELETE FROM group_edge WHERE destination_id=$1`} {
			_, _ = f.s.db.Exec(q, id)
		}
	})
	return id
}

func (f *fixture) call(fn apierr.Handler, user string, in map[string]any) (map[string]any, error) {
	raw, _ := json.Marshal(in)
	out, err := fn(context.Background(), user, string(raw))
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(out)
	m := map[string]any{}
	_ = json.Unmarshal(b, &m)
	return m, nil
}

func (f *fixture) must(fn apierr.Handler, user string, in map[string]any) map[string]any {
	f.t.Helper()
	out, err := f.call(fn, user, in)
	if err != nil {
		f.t.Fatalf("call failed: %v", err)
	}
	return out
}

func code(err error) string {
	if e, ok := apierr.As(err); ok {
		return e.Code
	}
	if err == nil {
		return ""
	}
	return err.Error()
}

func (f *fixture) create(leader, name, join string) string {
	out := f.must(f.s.rpcCreate, leader, map[string]any{"name": name, "joinType": join, "badge": map[string]any{"emblem": "star"}})
	id := out["clan"].(map[string]any)["clanId"].(string)
	f.clans = append(f.clans, id)
	return id
}

func (f *fixture) role(user string) string {
	m, _ := f.s.memberOf(context.Background(), f.s.db, user)
	if m == nil {
		return ""
	}
	return m.Role
}

func TestClanMembershipFlow(t *testing.T) {
	f := newFixture(t)
	leader, a, b, c := f.user(), f.user(), f.user(), f.user()
	open := f.create(leader, "Open "+plugintest.Key("c"), settings.JoinOpen)
	if _, err := f.call(f.s.rpcCreate, leader, map[string]any{"name": "Other " + plugintest.Key("c")}); code(err) != "CLAN_ALREADY_MEMBER" {
		t.Fatalf("one clan per player: %v", err)
	}
	if out := f.must(f.s.rpcJoin, a, map[string]any{"clanId": open}); out["status"] != "joined" {
		t.Fatalf("open join %v", out)
	}
	closedLeader := f.user()
	closed := f.create(closedLeader, "Closed "+plugintest.Key("c"), settings.JoinClosed)
	if out := f.must(f.s.rpcJoin, b, map[string]any{"clanId": closed}); out["status"] != "requested" {
		t.Fatalf("closed join %v", out)
	}
	if f.role(b) != "" || f.g.edge(context.Background(), closed, b) != 3 {
		t.Fatal("a request is a native join request, not a member")
	}
	if _, err := f.call(f.s.rpcAccept, leader, map[string]any{"userId": b}); code(err) != "CLAN_REQUEST_NOT_FOUND" {
		t.Fatalf("a member of another clan sees no request: %v", err)
	}
	reqs := f.must(f.s.rpcRequests, closedLeader, nil)
	if items := reqs["items"].([]any); len(items) != 1 {
		t.Fatalf("requests %v", reqs)
	}
	f.must(f.s.rpcAccept, closedLeader, map[string]any{"userId": b})
	if f.role(b) != settings.RoleMember || f.g.edge(context.Background(), closed, b) != 2 {
		t.Fatalf("accepted member: %s", f.role(b))
	}
	if !contains(f.notes[b], "clan_request_accepted") {
		t.Fatalf("accepted player is told: %v", f.notes[b])
	}
	f.must(f.s.rpcPromote, leader, map[string]any{"userId": a})
	if f.role(a) != settings.RoleElder {
		t.Fatalf("elder %s", f.role(a))
	}
	f.must(f.s.rpcPromote, leader, map[string]any{"userId": a})
	if f.role(a) != settings.RoleCoLeader || f.g.edge(context.Background(), open, a) != 1 {
		t.Fatalf("co-leader is native admin: %s %d", f.role(a), f.g.edge(context.Background(), open, a))
	}
	f.must(f.s.rpcJoin, c, map[string]any{"clanId": open})
	if _, err := f.call(f.s.rpcKick, c, map[string]any{"userId": a}); code(err) != "CLAN_ROLE_NOT_ALLOWED" {
		t.Fatalf("members cannot kick: %v", err)
	}
	f.must(f.s.rpcKick, a, map[string]any{"userId": c})
	if f.role(c) != "" {
		t.Fatal("kicked")
	}
	if _, err := f.call(f.s.rpcJoin, c, map[string]any{"clanId": open}); code(err) != "CLAN_KICK_COOLDOWN" {
		t.Fatalf("rejoin after a kick waits: %v", err)
	}
	if _, err := f.call(f.s.rpcKick, a, map[string]any{"userId": leader}); code(err) != "CLAN_ROLE_TOO_LOW" {
		t.Fatalf("co-leader cannot kick the leader: %v", err)
	}
	f.must(f.s.rpcLeave, leader, nil)
	if f.role(a) != settings.RoleLeader || f.g.edge(context.Background(), open, a) != 0 {
		t.Fatalf("leader leaving hands the clan to the co-leader: %s", f.role(a))
	}
	inviteOnly := f.create(f.user(), "Invite "+plugintest.Key("c"), settings.JoinInvite)
	d := f.user()
	if _, err := f.call(f.s.rpcJoin, d, map[string]any{"clanId": inviteOnly}); code(err) != "CLAN_INVITE_ONLY" {
		t.Fatalf("invite only: %v", err)
	}
	_ = inviteOnly
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func TestClanCapAndInvites(t *testing.T) {
	f := newFixture(t)
	leader := f.user()
	id := f.create(leader, "Cap "+plugintest.Key("c"), settings.JoinOpen)
	f.must(f.s.rpcJoin, f.user(), map[string]any{"clanId": id})
	f.must(f.s.rpcJoin, f.user(), map[string]any{"clanId": id})
	if _, err := f.call(f.s.rpcJoin, f.user(), map[string]any{"clanId": id}); code(err) != "CLAN_FULL" {
		t.Fatalf("level 1 holds 3 members: %v", err)
	}
	invited := f.user()
	other := f.create(f.user(), "Inv "+plugintest.Key("c"), settings.JoinInvite)
	owner := f.s.withRoles(context.Background(), other, []string{settings.RoleLeader})[0]
	f.must(f.s.rpcInvite, owner, map[string]any{"userId": invited})
	inv := f.must(f.s.rpcInvites, invited, nil)
	if len(inv["items"].([]any)) != 1 {
		t.Fatalf("invites %v", inv)
	}
	if out := f.must(f.s.rpcInviteAccept, invited, map[string]any{"clanId": other}); out["status"] != "joined" {
		t.Fatalf("invite accept %v", out)
	}
}

func TestClanXPLevelUpRaisesCapAndPaysMembers(t *testing.T) {
	f := newFixture(t)
	leader, a := f.user(), f.user()
	id := f.create(leader, "XP "+plugintest.Key("c"), settings.JoinOpen)
	f.must(f.s.rpcJoin, a, map[string]any{"clanId": id})
	ev := playerevents.Recorded{ID: uuid.Must(uuid.NewV4()).String(), UserID: a, Key: "cl_won", Value: 1}
	f.s.onEconomyEvent(context.Background(), ev)
	f.s.onEconomyEvent(context.Background(), ev)
	var xp int64
	var lv, cap int
	_ = f.s.db.QueryRow(`SELECT c.xp, c.level, g.max_count FROM clan c JOIN groups g ON g.id=c.group_id WHERE c.group_id::text=$1`, id).Scan(&xp, &lv, &cap)
	if xp != 30 || lv != 1 {
		t.Fatalf("one event counts once: xp %d level %d", xp, lv)
	}
	f.s.onEconomyEvent(context.Background(), playerevents.Recorded{ID: uuid.Must(uuid.NewV4()).String(), UserID: leader, Key: "cl_won", Value: 1})
	_ = f.s.db.QueryRow(`SELECT c.xp, c.level, g.max_count FROM clan c JOIN groups g ON g.id=c.group_id WHERE c.group_id::text=$1`, id).Scan(&xp, &lv, &cap)
	if xp != 60 || lv != 2 || cap != 5 {
		t.Fatalf("level 2 at 50 xp with a cap of 5: xp %d level %d cap %d", xp, lv, cap)
	}
	paid := 0
	for _, g := range f.grants {
		if g.PrizeKey == "cl_lv2" {
			paid++
		}
	}
	if paid != 2 || !contains(f.notes[a], "clan_level_up") {
		t.Fatalf("both members get the level prize once: %d %v", paid, f.notes[a])
	}
	levels := f.must(f.s.rpcLevels, a, nil)
	if lv := levels["clan"].(map[string]any)["level"].(float64); lv != 2 {
		t.Fatalf("levels view %v", levels)
	}
}

func (f *fixture) war(mode string, rules WarRules) *warConfig {
	f.t.Helper()
	rules.normalize(mode)
	raw, _ := json.Marshal(rules)
	key := strings.ReplaceAll(plugintest.Key("clt"), "-", "_")
	plugintest.Exec(f.t, f.s.db, `INSERT INTO clan_war_config (id, key, name, mode, duration_minutes, rules) VALUES ($1,$2,'{"en":"War"}',$3,60,$4)`, uuid.Must(uuid.NewV4()).String(), key, mode, raw)
	c, err := scanWarConfig(f.s.db.QueryRow(`SELECT `+warConfigCols+` FROM clan_war_config WHERE key=$1`, key))
	if err != nil {
		f.t.Fatal(err)
	}
	return c
}

func (f *fixture) startWar(c *warConfig) string {
	f.t.Helper()
	ctx := context.Background()
	plugintest.Exec(f.t, f.s.db, `UPDATE clan_war_round SET status='cancelled' WHERE status IN ('signup','active')`)
	id, err := f.s.createRound(ctx, c, time.Now().Add(-time.Second), 60, "test")
	if err != nil {
		f.t.Fatal(err)
	}
	if err := f.s.startRound(ctx, id); err != nil {
		f.t.Fatal(err)
	}
	return id
}

func (f *fixture) ev(user, key, match string, fields map[string]any) {
	m := map[string]any{"matchId": match}
	for k, v := range fields {
		m[k] = v
	}
	f.s.onEconomyEvent(context.Background(), playerevents.Recorded{ID: uuid.Must(uuid.NewV4()).String(), UserID: user, Key: key, Value: 1, Metadata: m})
}

func (f *fixture) warOf(round string) (string, string, string, int64, int64, int, int) {
	var id, a, b string
	var sa, sb int64
	var wa, wb int
	_ = f.s.db.QueryRow(`SELECT id::text, clan_a::text, COALESCE(clan_b::text,''), score_a, score_b, wins_a, wins_b FROM clan_war WHERE round_id::text=$1`, round).Scan(&id, &a, &b, &sa, &sb, &wa, &wb)
	return id, a, b, sa, sb, wa, wb
}

func TestPointsWarScoresOncePerMatchAndSettles(t *testing.T) {
	f := newFixture(t)
	l1, m1, l2 := f.user(), f.user(), f.user()
	c1 := f.create(l1, "WarA "+plugintest.Key("c"), settings.JoinOpen)
	f.must(f.s.rpcJoin, m1, map[string]any{"clanId": c1})
	c2 := f.create(l2, "WarB "+plugintest.Key("c"), settings.JoinOpen)
	rules := defaultRules(modePoints)
	rules.Points = []PointRule{{Type: "economy_event", Event: "clw_won", Points: 3}, {Type: "economy_event", Event: "clw_played", Points: 1}}
	rules.AttacksPerMember = 2
	rules.Prizes = Prizes{Win: Prize{PrizeKey: "clw_win"}, Lose: Prize{PrizeKey: "clw_lose"}, To: "participants"}
	plugintest.Exec(t, f.s.db, `UPDATE clan SET trophies=-1 WHERE group_id::text NOT IN ($1,$2)`, c1, c2)
	round := f.startWar(f.war(modePoints, rules))
	var wars int
	_ = f.s.db.QueryRow(`SELECT count(*) FROM clan_war w WHERE round_id::text=$1 AND (clan_a::text IN ($2,$3) OR clan_b::text IN ($2,$3))`, round, c1, c2).Scan(&wars)
	if wars == 0 {
		t.Fatal("both clans are paired")
	}
	var mine string
	_ = f.s.db.QueryRow(`SELECT id::text FROM clan_war WHERE round_id::text=$1 AND (clan_a::text=$2 OR clan_b::text=$2)`, round, c1).Scan(&mine)
	var rival string
	_ = f.s.db.QueryRow(`SELECT CASE WHEN clan_a::text=$2 THEN COALESCE(clan_b::text,'') ELSE clan_a::text END FROM clan_war WHERE id::text=$1`, mine, c1).Scan(&rival)
	if rival == "" {
		t.Skip("another clan in the shared test database took the pair; pairing is covered by the rule test")
	}
	mA := uuid.Must(uuid.NewV4()).String()
	f.ev(m1, "clw_won", mA, nil)
	f.ev(m1, "clw_played", mA, nil)
	f.ev(m1, "clw_won", mA, nil)
	f.ev(m1, "clw_won", uuid.Must(uuid.NewV4()).String(), nil)
	f.ev(m1, "clw_won", uuid.Must(uuid.NewV4()).String(), nil)
	var points int64
	var attacks int
	_ = f.s.db.QueryRow(`SELECT points, attacks FROM clan_war_member WHERE war_id::text=$1 AND user_id::text=$2`, mine, m1).Scan(&points, &attacks)
	if points != 7 || attacks != 2 {
		t.Fatalf("3+1 for match A, a resent event is ignored, 3 for match B, the third match is over the limit: points %d attacks %d", points, attacks)
	}
	if err := f.s.settleRound(context.Background(), round); err != nil {
		t.Fatal(err)
	}
	if err := f.s.settleRound(context.Background(), round); err != nil {
		t.Fatal(err)
	}
	var wins int
	var xp int64
	_ = f.s.db.QueryRow(`SELECT war_wins, xp FROM clan WHERE group_id::text=$1`, c1).Scan(&wins, &xp)
	if wins != 1 || xp != settings.Defaults().XP.WarWin {
		t.Fatalf("war win and its XP once: wins %d xp %d", wins, xp)
	}
	win, lose := 0, 0
	for _, g := range f.grants {
		switch {
		case g.PrizeKey == "clw_win" && g.UserID == m1:
			win++
		case g.PrizeKey == "clw_win" && g.UserID == l1:
			t.Fatal("a member who did not play gets no prize with prizes.to=participants")
		case g.PrizeKey == "clw_lose":
			lose++
		}
	}
	if win != 1 || lose != 0 {
		t.Fatalf("prizes: win %d lose %d", win, lose)
	}
}

func TestDuelWarCountsEachMatchOnceForTheWinner(t *testing.T) {
	f := newFixture(t)
	a1, a2, b1 := f.user(), f.user(), f.user()
	ca := f.create(a1, "DuelA "+plugintest.Key("c"), settings.JoinOpen)
	f.must(f.s.rpcJoin, a2, map[string]any{"clanId": ca})
	cb := f.create(b1, "DuelB "+plugintest.Key("c"), settings.JoinOpen)
	rules := defaultRules(modeDuel)
	rules.Duel.Event = "cld_played"
	rules.Duel.WinPoints = 1
	rules.AttacksPerMember = 5
	plugintest.Exec(t, f.s.db, `UPDATE clan SET trophies=-1 WHERE group_id::text NOT IN ($1,$2)`, ca, cb)
	plugintest.Exec(t, f.s.db, `UPDATE clan SET trophies=100 WHERE group_id::text IN ($1,$2)`, ca, cb)
	cfg := f.war(modeDuel, rules)
	cfg.Pairing = "trophies"
	plugintest.Exec(t, f.s.db, `UPDATE clan_war_config SET pairing='trophies' WHERE id::text=$1`, cfg.ID)
	round := f.startWar(cfg)
	var war string
	_ = f.s.db.QueryRow(`SELECT id::text FROM clan_war WHERE round_id::text=$1 AND ((clan_a::text=$2 AND clan_b::text=$3) OR (clan_a::text=$3 AND clan_b::text=$2))`, round, ca, cb).Scan(&war)
	if war == "" {
		t.Skip("the two test clans were not paired together in the shared database")
	}
	if r, ok := f.s.rivalOf(context.Background(), a1, b1); !ok || r.WarID != war || r.Mode != modeDuel {
		t.Fatalf("rival check %+v %v", r, ok)
	}
	if _, ok := f.s.rivalOf(context.Background(), a1, a2); ok {
		t.Fatal("clan mates are not rivals")
	}
	m := uuid.Must(uuid.NewV4()).String()
	f.ev(a1, "cld_played", m, map[string]any{"won": true, "opponentId": b1})
	f.ev(b1, "cld_played", m, map[string]any{"won": false, "opponentId": a1})
	f.ev(a1, "cld_played", m, map[string]any{"won": true, "opponentId": b1})
	d := uuid.Must(uuid.NewV4()).String()
	f.ev(a2, "cld_played", d, map[string]any{"won": false, "draw": true, "opponentId": b1})
	f.ev(b1, "cld_played", d, map[string]any{"won": false, "draw": true, "opponentId": a2})
	f.ev(a2, "cld_played", uuid.Must(uuid.NewV4()).String(), map[string]any{"won": true, "opponentId": a1})
	var sa, sb int64
	var wa, wb int
	var clanA string
	_ = f.s.db.QueryRow(`SELECT clan_a::text, score_a, score_b, wins_a, wins_b FROM clan_war WHERE id::text=$1`, war).Scan(&clanA, &sa, &sb, &wa, &wb)
	if clanA != ca {
		sa, sb, wa, wb = sb, sa, wb, wa
	}
	if sa != 1 || wa != 1 || sb != 0 || wb != 0 {
		t.Fatalf("one win for clan A, the draw scores 0, a match against a clan mate is ignored: %d-%d wins %d-%d", sa, sb, wa, wb)
	}
	var att int
	_ = f.s.db.QueryRow(`SELECT attacks FROM clan_war_member WHERE war_id::text=$1 AND user_id::text=$2`, war, b1).Scan(&att)
	if att != 2 {
		t.Fatalf("b1 played two counted duels: %d", att)
	}
	if err := f.s.settleRound(context.Background(), round); err != nil {
		t.Fatal(err)
	}
	var res string
	_ = f.s.db.QueryRow(`SELECT result FROM clan_war WHERE id::text=$1`, war).Scan(&res)
	if (clanA == ca && res != "a") || (clanA != ca && res != "b") {
		t.Fatalf("clan A wins the duel war: %s", res)
	}
}
