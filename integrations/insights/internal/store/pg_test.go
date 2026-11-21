package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

func gameKey() string {
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	return "t" + hex.EncodeToString(b)
}

func ev(game, user, name string, day string, hour int, value *float64, tags map[string]string) Event {
	t, _ := time.Parse(time.DateOnly, day)
	at := t.Add(time.Duration(hour) * time.Hour)
	return Event{ID: uuid.Must(uuid.NewV4()).String(), Game: game, Name: name, Value: value, Tags: tags, UserID: user, SessionID: "s-" + user + "-" + day,
		At: at, Day: day, ServerAt: at, Platform: "android", AppVersion: "1.0", Source: "client"}
}

func num(v float64) *float64 { return &v }

func storeContract(t *testing.T, st Store) {
	ctx := context.Background()
	g := gameKey()
	gp := &g
	day1, day2, day8 := "2026-06-01", "2026-06-02", "2026-06-08"
	list := []Event{
		ev(g, "u1", "DistanceOverall", day1, 1, num(150), map[string]string{"mission": "m1", "car": "c1"}),
		ev(g, "u1", "DistanceOverall", day1, 2, num(50), map[string]string{"mission": "m2", "car": "c1"}),
		ev(g, "u2", "DistanceOverall", day1, 3, num(100), map[string]string{"mission": "m1", "car": "c2"}),
		ev(g, "u2", "DistanceOverall", day2, 3, num(25), map[string]string{"mission": "m1", "car": "c2"}),
		ev(g, "u3", "DistanceOverall", day2, 4, num(10), map[string]string{"mission": "m1"}),
		ev(g, "u1", "LevelStart", day1, 1, nil, map[string]string{"level": "1"}),
		ev(g, "u1", "LevelEnd", day1, 2, nil, map[string]string{"level": "1"}),
		ev(g, "u2", "LevelStart", day1, 3, nil, map[string]string{"level": "1"}),
		ev(g, "u2", "LevelEnd", day2, 9, nil, map[string]string{"level": "1"}),
		ev(g, "u1", "LevelStart", day8, 1, nil, nil),
	}
	flagged := ev(g, "u9", "DistanceOverall", day1, 1, num(1e6), map[string]string{"car": "c1"})
	flagged.Flagged, flagged.FlagReason = true, "value_range"
	list = append(list, flagged)
	if err := st.Write(ctx, list); err != nil {
		t.Fatal(err)
	}
	if err := st.Write(ctx, list[:3]); err != nil {
		t.Fatal("rewrite", err)
	}
	base := Filter{From: day1, To: day8, Game: gp, Names: []string{"DistanceOverall"}}
	agg := func(metric string, groups []string, f Filter) []AggRow {
		t.Helper()
		rows, err := st.Aggregate(ctx, AggQuery{Filter: f, Metric: metric, GroupBy: groups})
		if err != nil {
			t.Fatal(err)
		}
		return rows
	}
	if r := agg(MetricSum, nil, base); len(r) != 1 || r[0].Value != 335 || r[0].Count != 5 || r[0].Users != 3 {
		t.Fatalf("total %+v", r)
	}
	withFlagged := base
	withFlagged.IncludeFlagged = true
	if r := agg(MetricSum, nil, withFlagged); r[0].Value != 1e6+335 {
		t.Fatalf("flagged %+v", r)
	}
	byUser := agg(MetricSum, []string{"user"}, base)
	if len(byUser) != 3 || byUser[0].Keys[0] != "u1" || byUser[0].Value != 200 {
		t.Fatalf("by user %+v", byUser)
	}
	byCar := agg(MetricSum, []string{"tag:car"}, base)
	if len(byCar) != 2 || byCar[0].Keys[0] != "c1" || byCar[0].Value != 200 || byCar[1].Value != 125 {
		t.Fatalf("by car %+v", byCar)
	}
	byMission := agg(MetricSum, []string{"tag:mission"}, base)
	if len(byMission) != 2 || byMission[0].Keys[0] != "m1" || byMission[0].Value != 285 {
		t.Fatalf("by mission %+v", byMission)
	}
	combo := agg(MetricSum, []string{"tag:car", "tag:mission", "user"}, base)
	if len(combo) != 3 || combo[0].Keys[0] != "c1" || combo[0].Keys[1] != "m1" || combo[0].Keys[2] != "u1" || combo[0].Value != 150 {
		t.Fatalf("combo %+v", combo)
	}
	filtered := base
	filtered.Tags = map[string][]string{"car": {"c2"}}
	if r := agg(MetricSum, nil, filtered); r[0].Value != 125 {
		t.Fatalf("filter %+v", r)
	}
	daily := agg(MetricCount, nil, base)
	_ = daily
	rows, err := st.Aggregate(ctx, AggQuery{Filter: base, Metric: MetricSum, Bucket: BucketDay})
	if err != nil || len(rows) != 2 || rows[0].Bucket != day1 || rows[0].Value != 300 {
		t.Fatalf("bucket %+v %v", rows, err)
	}
	if r := agg(MetricAvg, nil, base); r[0].Value != 67 {
		t.Fatalf("avg %+v", r)
	}
	if r := agg(MetricUsers, nil, base); r[0].Value != 3 {
		t.Fatalf("users %+v", r)
	}
	all := Filter{From: day1, To: day8, Game: gp}
	act, err := st.Activity(ctx, all)
	if err != nil {
		t.Fatal(err)
	}
	if act.Days[0].DAU != 2 || act.Days[0].NewUsers != 2 || act.Days[1].DAU != 2 || act.Days[1].NewUsers != 1 || act.Days[1].Returning != 1 || act.DAU != 1 || act.WAU != 3 || act.MAU != 3 || act.NewUsers != 3 {
		t.Fatalf("activity %+v", act)
	}
	ret, err := st.Retention(ctx, RetentionQuery{Filter: all, Days: []int{1, 7}, Today: "2026-07-01"})
	if err != nil {
		t.Fatal(err)
	}
	if ret.Summary[0].Cohort != 3 || ret.Summary[0].Returned != 1 || ret.Summary[1].Returned != 1 || len(ret.Cohorts) != 2 || ret.Cohorts[1].Size != 2 || *ret.Cohorts[1].Cells[0].Returned != 1 {
		t.Fatalf("retention %+v", ret)
	}
	notYet, err := st.Retention(ctx, RetentionQuery{Filter: all, Days: []int{7}, Today: day2})
	if err != nil || notYet.Cohorts[0].Cells[0].Returned != nil {
		t.Fatalf("future cell %+v %v", notYet, err)
	}
	rolling, _ := st.Retention(ctx, RetentionQuery{Filter: all, Days: []int{1}, Mode: ModeRolling, Today: "2026-07-01"})
	if rolling.Summary[0].Returned != 2 {
		t.Fatalf("rolling %+v", rolling.Summary)
	}
	steps, err := st.Funnel(ctx, FunnelQuery{Filter: Filter{From: day1, To: day8, Game: gp}, Steps: []Step{{Name: "LevelStart", Tags: map[string][]string{"level": {"1"}}}, {Name: "LevelEnd"}}, Window: 24 * time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	if steps[0].Users != 2 || steps[1].Users != 1 || steps[1].MedianSeconds == nil || *steps[1].MedianSeconds != 3600 {
		t.Fatalf("funnel %+v", steps)
	}
	wide, _ := st.Funnel(ctx, FunnelQuery{Filter: Filter{From: day1, To: day1, Game: gp}, Steps: []Step{{Name: "LevelStart"}, {Name: "LevelEnd"}}, Window: 48 * time.Hour})
	if wide[1].Users != 2 {
		t.Fatalf("wide window %+v", wide)
	}
	evs, err := st.Events(ctx, EventQuery{Filter: Filter{From: day1, To: day8, Game: gp, UserID: "u1"}, Limit: 3})
	if err != nil || len(evs) != 3 || evs[0].Day != day8 || evs[0].Name != "LevelStart" {
		t.Fatalf("events %+v %v", evs, err)
	}
	more, err := st.Events(ctx, EventQuery{Filter: Filter{From: day1, To: day8, Game: gp, UserID: "u1"}, Limit: 10, BeforeAt: &evs[2].At, BeforeID: evs[2].ID})
	if err != nil || len(more) != 2 {
		t.Fatalf("cursor %+v %v", more, err)
	}
	asc, _ := st.Events(ctx, EventQuery{Filter: Filter{From: day1, To: day8, Game: gp, SessionID: "s-u1-" + day1}, Asc: true})
	if len(asc) != 4 || asc[3].At.Before(asc[0].At) {
		t.Fatalf("session events %+v", asc)
	}
	vals, err := st.Values(ctx, ValuesQuery{Filter: Filter{From: day1, To: day8, Game: gp}, Kind: "tag:mission"})
	if err != nil || len(vals) != 2 || vals[0].Value != "m1" || vals[0].Count != 4 {
		t.Fatalf("values %+v %v", vals, err)
	}
	names, _ := st.Values(ctx, ValuesQuery{Filter: Filter{From: day1, To: day8, Game: gp}, Kind: "name"})
	groups, _ := st.Values(ctx, ValuesQuery{Filter: Filter{From: day1, To: day8, Game: gp}, Kind: "tags"})
	if len(names) != 3 || len(groups) != 3 {
		t.Fatalf("names %+v groups %+v", names, groups)
	}
	if _, err := st.Aggregate(ctx, AggQuery{Filter: base, Metric: "median"}); err != ErrBadQuery {
		t.Fatalf("bad metric %v", err)
	}
	if _, err := st.Aggregate(ctx, AggQuery{Filter: base, GroupBy: []string{"tag:Car"}}); err != ErrBadQuery {
		t.Fatalf("bad group %v", err)
	}
	n, err := st.Sweep(ctx, "1990-01-01")
	if err != nil || n != 0 {
		t.Fatalf("sweep none %d %v", n, err)
	}
	old := ev(g, "u7", "Old", "2001-01-01", 1, nil, nil)
	if err := st.Write(ctx, []Event{old}); err != nil {
		t.Fatal(err)
	}
	if n, err := st.Sweep(ctx, "2001-01-02"); err != nil || n < 1 {
		t.Fatalf("sweep %d %v", n, err)
	}
}

func TestNakamaPostgresStore(t *testing.T) {
	db := plugintest.OpenDB(t)
	st := NewNakamaPG(db)
	if err := st.Ensure(context.Background()); err != nil {
		t.Fatal(err)
	}
	storeContract(t, st)
}

func TestExternalPostgresStoreCreatesItsSchema(t *testing.T) {
	db := plugintest.OpenDB(t)
	schema := "ins_ext_" + gameKey()
	t.Cleanup(func() { _, _ = db.Exec(`DROP SCHEMA IF EXISTS "` + schema + `" CASCADE`) })
	st, err := NewExternalPG(db, schema)
	if err != nil {
		t.Fatal(err)
	}
	st.owned = false
	if err := st.Ensure(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := st.Ensure(context.Background()); err != nil {
		t.Fatal("second ensure", err)
	}
	if err := st.Ping(context.Background()); err == nil {
		t.Fatal("empty table should answer no rows")
	}
	storeContract(t, st)
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM "` + schema + `".insights_event`).Scan(&n); err != nil || n < 10 {
		t.Fatalf("rows in external schema: %d %v", n, err)
	}
	if _, err := NewExternalPG(db, "Bad-Schema"); err == nil {
		t.Fatal("bad schema accepted")
	}
}
