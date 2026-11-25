package main

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/eventcontext"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

func (f *fixture) attempts(user string, spans ...eventcontext.Span) {
	eventcontext.Register(eventcontext.Provider{Name: "test", Spans: func(_ context.Context, userID string, from, to time.Time) ([]eventcontext.Span, error) {
		if userID != user {
			return nil, nil
		}
		return spans, nil
	}})
	f.t.Cleanup(func() { eventcontext.Unregister("test") })
}

func (f *fixture) tagsOf(user, name string) []string {
	f.t.Helper()
	rows, err := f.s.db.Query(`SELECT tags::text || '|' || flagged::text || '|' || flag_reason || '|' || fields::text FROM insights_event WHERE game=$1 AND user_id=$2 AND name=$3 ORDER BY at, id`, f.prefix, user, name)
	if err != nil {
		f.t.Fatal(err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			f.t.Fatal(err)
		}
		out = append(out, s)
	}
	return out
}

func ms(t time.Time) string { return strconv.FormatInt(t.UnixMilli(), 10) }

func TestServerFillsMissionAndCarFromTheOpenAttempt(t *testing.T) {
	f := newFixture(t)
	f.cleanupEvents()
	u := plugintest.NewUser(t, f.s.db)
	start := f.at.Add(-30 * time.Minute)
	f.attempts(u, eventcontext.Span{ID: "a1", From: start, To: start.Add(20 * time.Minute), Values: map[string]string{"campaign.stage": "chapter-1-level-01", "campaign.chapter": "chapter-1", "campaign.item": "tiggo8"}})
	inside, after := start.Add(5*time.Minute), start.Add(25*time.Minute)
	out, err := f.batch(u, "ctx-00001", `[{"name":"DistanceOverall","value":150,"time":`+ms(inside)+`},
		{"name":"DistanceOverall","tags":[{"tagGroup":"mission","id":"3434343"},{"tagGroup":"car","id":"345345"}],"value":7,"time":`+ms(inside)+`},
		{"name":"DistanceOverall","tags":{"car":"audiq2"},"value":3,"time":`+ms(after)+`},
		{"name":"DistanceOverall","tags":{"car":"TIGGO8"},"value":1,"time":`+ms(inside)+`}]`)
	if err != nil {
		t.Fatal(err)
	}
	if out.Accepted != 3 || out.Flagged != 1 || out.Results[1].Code != "ANALYTICS_CONTEXT_MISMATCH" {
		t.Fatalf("send %+v", out)
	}
	dup, err := f.batch(u, "ctx-00001", `[{"name":"DistanceOverall","value":1}]`)
	if err != nil || !dup.Duplicate || dup.Accepted != 3 {
		t.Fatalf("duplicate %+v %v", dup, err)
	}
	f.drain()
	got := f.tagsOf(u, "DistanceOverall")
	if len(got) != 4 {
		t.Fatalf("rows %v", got)
	}
	want := []func(string) bool{
		func(s string) bool {
			return strings.Contains(s, `"car": "tiggo8"`) && strings.Contains(s, `"mission": "chapter-1-level-01"`) && !strings.Contains(s, "chapter\":") && strings.Contains(s, "|false|")
		},
		func(s string) bool {
			return strings.Contains(s, `"car": "tiggo8"`) && strings.Contains(s, `"mission": "chapter-1-level-01"`) && strings.Contains(s, "|true|context_mismatch|") && strings.Contains(s, `"clientTag.mission": "3434343"`) && strings.Contains(s, `"clientTag.car": "345345"`)
		},
		func(s string) bool { return strings.Contains(s, `"car": "tiggo8"`) && strings.Contains(s, "|false|") },
		func(s string) bool { return strings.HasPrefix(s, `{"car": "audiq2"}|false|`) },
	}
	for _, w := range want {
		ok := false
		for _, g := range got {
			if w(g) {
				ok = true
			}
		}
		if !ok {
			t.Fatalf("stored rows %v", got)
		}
	}
	q := "&game=" + f.prefix + "&from=2026-06-01&to=2026-06-01&includeFlagged=true"
	ex := okc(t, f.s.explore, "GET", "/x?name=DistanceOverall&metric=sum&groupBy=tag:car,tag:mission"+q, nil, nil)
	sums := map[string]float64{}
	for _, r := range ex["rows"].([]any) {
		m := r.(map[string]any)
		keys := []string{}
		for _, k := range m["keys"].([]any) {
			keys = append(keys, k.(string))
		}
		sums[strings.Join(keys, "/")] = m["value"].(float64)
	}
	if sums["tiggo8/chapter-1-level-01"] != 158 || len(sums) != 1 {
		t.Fatalf("split by tag %v", sums)
	}
	byCar := okc(t, f.s.explore, "GET", "/x?name=DistanceOverall&metric=sum&groupBy=tag:car"+q, nil, nil)
	cars := map[string]float64{}
	for _, r := range byCar["rows"].([]any) {
		m := r.(map[string]any)
		cars[m["keys"].([]any)[0].(string)] = m["value"].(float64)
	}
	if cars["tiggo8"] != 158 || cars["audiq2"] != 3 || len(cars) != 2 {
		t.Fatalf("split by car %v", cars)
	}
}

func TestContextPoliciesKeepRejectAndOutside(t *testing.T) {
	f := newFixture(t)
	f.cleanupEvents()
	u := plugintest.NewUser(t, f.s.db)
	start := f.at.Add(-10 * time.Minute)
	f.attempts(u, eventcontext.Span{ID: "a1", From: start, Values: map[string]string{"campaign.stage": "s-key", "campaign.item": ""}})
	f.settings(func(c *settings.All) { c.Context.Conflict = "keep_client" })
	out, err := f.batch(u, "ctx-00002", `[{"name":"Shot","tags":{"mission":"m-client","car":"c-client"},"value":1}]`)
	if err != nil || out.Flagged != 1 {
		t.Fatalf("keep %+v %v", out, err)
	}
	f.settings(func(c *settings.All) { c.Context.Conflict = "reject"; c.Context.FlagMismatch = false })
	out, err = f.batch(u, "ctx-00003", `[{"name":"Shot","tags":{"mission":"m-client"},"value":1},{"name":"Shot","tags":{"mission":"s-key"},"value":1},{"name":"Shot","value":1}]`)
	if err != nil || out.Rejected != 1 || out.Accepted != 2 || out.Results[0].Code != "ANALYTICS_CONTEXT_MISMATCH" || out.Results[0].Problems[0] != "mission" {
		t.Fatalf("reject %+v %v", out, err)
	}
	f.settings(func(c *settings.All) { c.Context.KeepClientOutside = false })
	out, err = f.batch(u, "ctx-00004", `[{"name":"Shop","tags":{"car":"audiq2","item":"x"},"value":1,"time":`+ms(start.Add(-time.Hour))+`}]`)
	if err != nil || out.Accepted != 1 {
		t.Fatalf("outside %+v %v", out, err)
	}
	f.settings(func(c *settings.All) { c.Context.Enabled = false })
	if _, err := f.batch(u, "ctx-00005", `[{"name":"Off","value":1}]`); err != nil {
		t.Fatal(err)
	}
	f.drain()
	g := f.tagsOf(u, "Shot")
	keep, filled := 0, 0
	for _, r := range g {
		if strings.Contains(r, `"mission": "m-client"`) && strings.Contains(r, `"serverTag.mission": "s-key"`) && !strings.Contains(r, "serverTag.car") && strings.Contains(r, "|true|context_mismatch|") {
			keep++
		}
		if strings.HasPrefix(r, `{"mission": "s-key"}|false|`) {
			filled++
		}
	}
	if len(g) != 3 || keep != 1 || filled != 2 {
		t.Fatalf("keep client rows %v", g)
	}
	if g := f.tagsOf(u, "Shop"); len(g) != 1 || !strings.HasPrefix(g[0], `{"item": "x"}|`) {
		t.Fatalf("outside dropped %v", g)
	}
	if g := f.tagsOf(u, "Off"); len(g) != 1 || !strings.HasPrefix(g[0], `{}|`) {
		t.Fatalf("disabled %v", g)
	}
}

func TestNoProviderLeavesClientTagsAlone(t *testing.T) {
	f := newFixture(t)
	f.cleanupEvents()
	u := plugintest.NewUser(t, f.s.db)
	f.settings(func(c *settings.All) {
		c.Context.Conflict = "reject"
		c.Context.KeepClientOutside = false
	})
	out, err := f.batch(u, "ctx-00006", `[{"name":"DistanceOverall","tags":{"mission":"m1","car":"c1"},"value":2}]`)
	if err != nil || out.Accepted != 1 {
		t.Fatalf("send %+v %v", out, err)
	}
	f.drain()
	if g := f.tagsOf(u, "DistanceOverall"); len(g) != 1 || !strings.HasPrefix(g[0], `{"car": "c1", "mission": "m1"}|false|`) {
		t.Fatalf("rows %v", g)
	}
	cfg := settings.Defaults()
	cfg.Context.Fills = append(cfg.Context.Fills, eventcontext.Fill{Group: "Car", Source: "campaign.item"})
	if f, c := cfg.Validate(); c != "INVALID_TAG_GROUP" || f != "context.fills.group" {
		t.Fatalf("validate %s %s", f, c)
	}
	d := okc(t, f.s.eventSettingsGet, "GET", "/x", nil, nil)
	if d["context"].(map[string]any)["conflict"] != "reject" || d["defaults"].(map[string]any)["context"].(map[string]any)["conflict"] != "override" {
		t.Fatalf("settings get %v", d)
	}
	put := okc(t, f.s.eventSettingsPut, "PUT", "/x", map[string]any{"context": map[string]any{"enabled": true, "conflict": "override", "flagMismatch": true, "keepClientOutside": true, "graceSeconds": 0,
		"fills": []map[string]any{{"group": "mission", "source": "campaign.stage", "enabled": true}}}}, nil)
	if fills := put["context"].(map[string]any)["fills"].([]any); len(fills) != 1 {
		t.Fatalf("settings put %v", put)
	}
	if code, _ := call(t, f.s.eventSettingsPut, "PUT", "/x", map[string]any{"context": map[string]any{"conflict": "nope", "fills": []any{}}}, nil); code != 400 {
		t.Fatalf("bad conflict %d", code)
	}
}
