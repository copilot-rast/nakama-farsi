package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
	"github.com/heroiclabs/nakama/v3/internal/insightsapi"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
	"github.com/heroiclabs/nakama/v3/internal/secretbox"
)

func (f *fixture) cleanupEvents() {
	g := f.prefix
	_, _ = f.s.db.Exec(`DELETE FROM insights_outbox`)
	f.t.Cleanup(func() {
		db := f.s.db
		_, _ = db.Exec(`DELETE FROM insights_event WHERE game=$1`, g)
		_, _ = db.Exec(`DELETE FROM insights_event_user WHERE game=$1`, g)
		_, _ = db.Exec(`DELETE FROM insights_event_user_day WHERE game=$1`, g)
		_, _ = db.Exec(`DELETE FROM insights_tag_group WHERE game=$1`, g)
		_, _ = db.Exec(`DELETE FROM insights_event_def WHERE game=$1`, g)
		_, _ = db.Exec(`DELETE FROM insights_funnel WHERE game=$1`, g)
		_, _ = db.Exec(`DELETE FROM insights_outbox`)
		_, _ = db.Exec(`DELETE FROM insights_storage`)
		_, _ = db.Exec(`DELETE FROM cluster_rate_limit WHERE key LIKE 'ins:%'`)
	})
}

func (f *fixture) drain() int {
	f.t.Helper()
	total := 0
	for i := 0; i < 50; i++ {
		n, err := f.s.drainOnce(context.Background())
		if err != nil {
			f.t.Fatal(err)
		}
		if n == 0 {
			return total
		}
		total += n
	}
	return total
}

func (f *fixture) batch(user, id string, events string) (*sendOut, error) {
	return f.s.send(context.Background(), user, mustSendIn(f.t, `{"batchId":"`+id+`","sessionId":"sess-1","platform":"Android","appVersion":"2.0","game":"`+f.prefix+`","events":`+events+`}`))
}

func mustSendIn(t *testing.T, raw string) sendIn {
	t.Helper()
	var in sendIn
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		t.Fatal(err)
	}
	return in
}

func call(t *testing.T, h http.HandlerFunc, method, target string, body any, vars map[string]string) (int, map[string]any) {
	t.Helper()
	var rd *bytes.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		rd = bytes.NewReader(raw)
	} else {
		rd = bytes.NewReader(nil)
	}
	r := httptest.NewRequest(method, target, rd)
	r.Header.Set("X-Nakama-Console-Username", "tester")
	for k, v := range vars {
		r.SetPathValue(k, v)
	}
	w := httptest.NewRecorder()
	h(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func okc(t *testing.T, h http.HandlerFunc, method, target string, body any, vars map[string]string) map[string]any {
	t.Helper()
	code, out := call(t, h, method, target, body, vars)
	if code != 200 {
		t.Fatalf("status %d %v", code, out)
	}
	return out
}

func TestSendValidatesStoresOnceAndDrains(t *testing.T) {
	f := newFixture(t)
	f.cleanupEvents()
	u := plugintest.NewUser(t, f.s.db)
	out, err := f.batch(u, "batch-0001", `[{"name":"DistanceOverall","tags":[{"tagGroup":"mission","id":"3434343"},{"tagGroup":"car","id":"345345"}],"value":150},
		{"path":"DistanceOverall:111:222","value":5},{"name":"9bad"},{"type":"progression","status":"Start","p1":"w1"}]`)
	if err != nil {
		t.Fatal(err)
	}
	if out.Received != 4 || out.Accepted != 2 || out.Flagged != 1 || out.Rejected != 1 || out.Results[2].Code != "ANALYTICS_INVALID_EVENT_NAME" || out.Results[1].Code != "ANALYTICS_PATH_GROUPS" {
		t.Fatalf("send %+v", out)
	}
	dup, err := f.batch(u, "batch-0001", `[{"name":"Other"}]`)
	if err != nil || !dup.Duplicate || dup.Accepted != 2 {
		t.Fatalf("duplicate %+v %v", dup, err)
	}
	var rows int
	_ = f.s.db.QueryRow(`SELECT count(*) FROM insights_outbox`).Scan(&rows)
	if rows != 1 {
		t.Fatalf("outbox rows %d", rows)
	}
	if n := f.drain(); n != 3 {
		t.Fatalf("drained %d", n)
	}
	var name, platform, version, session, tags string
	var value float64
	if err := f.s.db.QueryRow(`SELECT name, platform, app_version, session_id, tags::text, value FROM insights_event WHERE game=$1 AND user_id=$2 AND value=150`, f.prefix, u).Scan(&name, &platform, &version, &session, &tags, &value); err != nil {
		t.Fatal(err)
	}
	if name != "DistanceOverall" || platform != "android" || version != "2.0" || session != "sess-1" || !strings.Contains(tags, `"car": "345345"`) {
		t.Fatalf("stored %s %s %s %s %s", name, platform, version, session, tags)
	}
	var flagged int
	_ = f.s.db.QueryRow(`SELECT count(*) FROM insights_event WHERE game=$1 AND flagged AND flag_reason='path_groups' AND tags ? 'p1'`, f.prefix).Scan(&flagged)
	if flagged != 1 {
		t.Fatalf("flagged path %d", flagged)
	}
	for _, bad := range map[string]string{"": "ANALYTICS_BATCH_ID_REQUIRED", "short": "ANALYTICS_INVALID_BATCH_ID"} {
		_ = bad
	}
	if _, err := f.batch(u, "", `[{"name":"A"}]`); errCode(err) != "ANALYTICS_BATCH_ID_REQUIRED" {
		t.Fatalf("no batch id %v", err)
	}
	if _, err := f.batch(u, "batch-0002", `[]`); errCode(err) != "ANALYTICS_BATCH_EMPTY" {
		t.Fatalf("empty %v", err)
	}
	f.settings(func(c *settings.All) { c.Events.MaxBatch = 1; c.Events.BatchesPerMinute = 3 })
	if _, err := f.batch(u, "batch-0003", `[{"name":"A"},{"name":"B"}]`); errCode(err) != "ANALYTICS_BATCH_TOO_LARGE" {
		t.Fatalf("too large %v", err)
	}
	_, e1 := f.batch(u, "batch-0004", `[{"name":"A"}]`)
	_, e2 := f.batch(u, "batch-0005", `[{"name":"A"}]`)
	if _, err := f.batch(u, "batch-0006", `[{"name":"A"}]`); e1 != nil || e2 != nil || errCode(err) != "ANALYTICS_RATE_LIMITED" {
		t.Fatalf("rate limit %v %v %v", e1, e2, err)
	}
	f.settings(func(c *settings.All) {})
	f.s.pending.Store(int64(settings.DefaultStorage().MaxPendingEvents))
	if _, err := f.batch(u, "batch-0007", `[{"name":"A"}]`); errCode(err) != "ANALYTICS_BACKLOG_FULL" {
		t.Fatalf("backlog %v", err)
	}
	f.s.pending.Store(0)
	f.settings(func(c *settings.All) { c.Events.Enabled = false })
	if _, err := f.batch(u, "batch-0008", `[{"name":"A"}]`); errCode(err) != "ANALYTICS_DISABLED" {
		t.Fatalf("disabled %v", err)
	}
}

func TestCatalogDrivesPathsAndPolicies(t *testing.T) {
	f := newFixture(t)
	f.cleanupEvents()
	u := plugintest.NewUser(t, f.s.db)
	q := "?game=" + f.prefix
	okc(t, f.s.groupPut, "PUT", "/x"+q, map[string]any{"nameFa": "ماشین", "nameEn": "Car", "closed": true, "enabled": true, "values": []map[string]string{{"id": "345345", "labelFa": "سانتافه"}}}, map[string]string{"key": "car"})
	okc(t, f.s.groupPut, "PUT", "/x"+q, map[string]any{"nameEn": "Mission", "enabled": true}, map[string]string{"key": "mission"})
	okc(t, f.s.eventDefPut, "PUT", "/x"+q, map[string]any{"labelEn": "Distance", "enabled": true, "pathGroups": []string{"car", "mission"}, "value": map[string]any{"min": 0, "max": 1000}}, map[string]string{"name": "DistanceOverall"})
	okc(t, f.s.eventDefPut, "PUT", "/x"+q, map[string]any{"enabled": false}, map[string]string{"name": "Muted"})
	if code, _ := call(t, f.s.groupPut, "PUT", "/x"+q, map[string]any{"values": []map[string]string{{"id": "a:b"}}}, map[string]string{"key": "bad"}); code != 400 {
		t.Fatalf("bad value code %d", code)
	}
	cat := okc(t, f.s.catalogGet, "GET", "/x"+q, nil, nil)
	if len(cat["groups"].([]any)) != 2 || len(cat["events"].([]any)) != 2 {
		t.Fatalf("catalog %v", cat)
	}
	out, err := f.batch(u, "batch-cat-1", `[{"path":"DistanceOverall:345345:m9","value":10},{"path":"DistanceOverall:999:m9","value":10},{"name":"DistanceOverall","value":5000},{"name":"Muted"}]`)
	if err != nil {
		t.Fatal(err)
	}
	if out.Results[0].Status != "stored" || out.Results[1].Code != "ANALYTICS_UNKNOWN_TAG_VALUE" || out.Results[2].Code != "ANALYTICS_VALUE_OUT_OF_RANGE" || out.Results[3].Status != "dropped" {
		t.Fatalf("catalog send %+v", out.Results)
	}
	f.drain()
	var mission string
	_ = f.s.db.QueryRow(`SELECT tags->>'mission' FROM insights_event WHERE game=$1 AND tags->>'car'='345345'`, f.prefix).Scan(&mission)
	if mission != "m9" {
		t.Fatalf("path mapped mission %q", mission)
	}
	cfgOut, err := f.s.rpcConfig(context.Background(), u, `{"game":"`+f.prefix+`"}`)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(cfgOut)
	if !strings.Contains(string(raw), `"pathGroups":["car","mission"]`) || !strings.Contains(string(raw), `"maxBatch":100`) {
		t.Fatalf("config %s", raw)
	}
	okc(t, f.s.groupDelete, "DELETE", "/x"+q, nil, map[string]string{"key": "car"})
	okc(t, f.s.eventDefDelete, "DELETE", "/x"+q, nil, map[string]string{"name": "Muted"})
}

func TestServerHelperAndGameEventsAreIdempotent(t *testing.T) {
	f := newFixture(t)
	f.cleanupEvents()
	u := plugintest.NewUser(t, f.s.db)
	insightsapi.Register(f.s.serverSink)
	gameevents.Subscribe("insights", f.s.gameEvent)
	defer insightsapi.Register(nil)
	defer gameevents.Subscribe("insights", nil)
	ctx := context.Background()
	e := insightsapi.Event{Key: "match:" + f.prefix, Game: f.prefix, UserID: u, Name: "MatchWon", Tags: insightsapi.Tags("mode", "ranked"), Value: insightsapi.Num(1)}
	for i := 0; i < 2; i++ {
		if r, err := insightsapi.Send(ctx, e); err != nil || r.Accepted != 1 {
			t.Fatalf("send %+v %v", r, err)
		}
	}
	if r, _ := insightsapi.Send(ctx, insightsapi.Event{Game: f.prefix, Name: "bad name"}); r.Rejected != 1 {
		t.Fatalf("bad accepted %+v", r)
	}
	score := int64(7)
	gameevents.Publish(ctx, gameevents.Event{ID: "ge-" + f.prefix, UserID: u, Name: "task_completed", Context: "daily", Fields: map[string]any{"amount": 3.0, "kind": "x"}, Source: "economy",
		Typed: []gameevents.Typed{gameevents.Progression("Complete", "c1", "s1", "", &score)}})
	gameevents.Publish(ctx, gameevents.Event{ID: "ge-" + f.prefix, UserID: u, Name: "task_completed", Context: "daily", Fields: map[string]any{"amount": 3.0}, Source: "economy"})
	f.drain()
	var n int
	_ = f.s.db.QueryRow(`SELECT count(*) FROM insights_event WHERE game=$1 AND name='MatchWon'`, f.prefix).Scan(&n)
	if n != 1 {
		t.Fatalf("server event stored %d", n)
	}
	var value float64
	var source, ctxTag, kind string
	if err := f.s.db.QueryRow(`SELECT value, source, tags->>'context', fields->>'kind' FROM insights_event WHERE user_id=$1 AND name='task_completed'`, u).Scan(&value, &source, &ctxTag, &kind); err != nil {
		t.Fatal(err)
	}
	if value != 3 || source != "economy" || ctxTag != "daily" || kind != "x" {
		t.Fatalf("game event %v %s %s %s", value, source, ctxTag, kind)
	}
	var prog float64
	_ = f.s.db.QueryRow(`SELECT value FROM insights_event WHERE user_id=$1 AND name='progression' AND tags->>'status'='complete'`, u).Scan(&prog)
	if prog != 7 {
		t.Fatalf("typed game event %v", prog)
	}
	_, _ = f.s.db.Exec(`DELETE FROM insights_event WHERE user_id=$1`, u)
	_, _ = f.s.db.Exec(`DELETE FROM insights_event_user WHERE user_id=$1`, u)
	_, _ = f.s.db.Exec(`DELETE FROM insights_event_user_day WHERE user_id=$1`, u)
}

func TestOutboxKeepsEventsWhileTheStoreIsDown(t *testing.T) {
	f := newFixture(t)
	f.cleanupEvents()
	u := plugintest.NewUser(t, f.s.db)
	key, _ := storageBox.DeriveKey([]byte(strings.Repeat("k", 40)), "insights-secrets")
	f.s.secretKey = key
	f.settings(func(c *settings.All) { c.Storage.DeadAfter = 2 })
	down := map[string]any{"driver": "postgres", "config": map[string]any{"postgres": map[string]any{"host": "127.0.0.1", "port": 1, "database": "nowhere_disposable", "user": "x", "sslMode": "disable"}}, "secret": "pw"}
	view := okc(t, f.s.storagePut, "PUT", "/x", down, nil)
	if view["driver"] != "postgres" || view["hasSecret"] != true {
		t.Fatalf("storage view %v", view)
	}
	if _, has := view["secret"]; has {
		t.Fatal("secret returned")
	}
	if _, err := f.batch(u, "batch-down-1", `[{"name":"A","value":1},{"name":"B"}]`); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.drainOnce(context.Background()); err == nil {
		t.Fatal("drain to a down store should fail")
	}
	var attempts int
	var dead bool
	var last string
	_ = f.s.db.QueryRow(`SELECT attempts, dead, last_error FROM insights_outbox`).Scan(&attempts, &dead, &last)
	if attempts != 1 || dead || last == "" {
		t.Fatalf("after failure %d %v %q", attempts, dead, last)
	}
	_, _ = f.s.db.Exec(`UPDATE insights_outbox SET next_at = now()`)
	_, _ = f.s.drainOnce(context.Background())
	box := okc(t, f.s.outboxGet, "GET", "/x", nil, nil)
	if box["dead"].(float64) != 1 || box["deadEvents"].(float64) != 2 {
		t.Fatalf("dead %v", box)
	}
	test := okc(t, f.s.storageTest, "POST", "/x", down, nil)
	if test["ok"] != false {
		t.Fatalf("test down %v", test)
	}
	dsn, _ := url.Parse(os.Getenv(plugintest.DBEnv))
	pass, _ := dsn.User.Password()
	schema := "ins_sw_" + strings.TrimPrefix(f.prefix, "it_")
	t.Cleanup(func() { _, _ = f.s.db.Exec(`DROP SCHEMA IF EXISTS "` + schema + `" CASCADE`) })
	up := map[string]any{"driver": "postgres", "config": map[string]any{"postgres": map[string]any{"host": dsn.Hostname(), "port": atoiOr(dsn.Port(), 5432), "database": strings.TrimPrefix(dsn.Path, "/"), "user": dsn.User.Username(), "sslMode": "disable", "schema": schema}}, "secret": pass}
	if test := okc(t, f.s.storageTest, "POST", "/x", up, nil); test["ok"] != true {
		t.Fatalf("test up %v", test)
	}
	okc(t, f.s.storagePut, "PUT", "/x", up, nil)
	okc(t, f.s.outboxRetry, "POST", "/x", nil, nil)
	if n := f.drain(); n != 2 {
		t.Fatalf("redelivered %d", n)
	}
	var ext int
	if err := f.s.db.QueryRow(`SELECT count(*) FROM "`+schema+`".insights_event WHERE game=$1`, f.prefix).Scan(&ext); err != nil || ext != 2 {
		t.Fatalf("external rows %d %v", ext, err)
	}
	cfg := f.s.settings(context.Background())
	rq := httptest.NewRequest("GET", "/x?game="+f.prefix+"&from=2026-06-01&to=2026-06-01", nil)
	fl, _, err := f.s.queryFilter(rq, cfg, 30)
	if err != nil {
		t.Fatal(err)
	}
	st, _ := f.s.eventStore(context.Background())
	if st.Kind() != "postgres" {
		t.Fatalf("store kind %s", st.Kind())
	}
	if a, err := st.Activity(context.Background(), fl); err != nil || a.DAU != 1 {
		t.Fatalf("external activity %+v %v", a, err)
	}
	keep := okc(t, f.s.storagePut, "PUT", "/x", map[string]any{"driver": "postgres", "config": up["config"]}, nil)
	if keep["hasSecret"] != true {
		t.Fatal("secret kept when not sent")
	}
	okc(t, f.s.storagePut, "PUT", "/x", map[string]any{"driver": "nakama"}, nil)
	if code, _ := call(t, f.s.storagePut, "PUT", "/x", map[string]any{"driver": "oracle"}, nil); code != 400 {
		t.Fatal("unknown driver accepted")
	}
	f.s.secretKey = nil
	if code, out := call(t, f.s.storagePut, "PUT", "/x", map[string]any{"driver": "postgres", "config": up["config"], "secret": "x"}, nil); code != 400 || out["error"] != "SECRET_KEY_MISSING" {
		t.Fatalf("no key %d %v", code, out)
	}
	_ = secretbox.SecretEnv
}

func atoiOr(s string, d int) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return d
		}
		n = n*10 + int(c-'0')
	}
	if n == 0 {
		return d
	}
	return n
}

func TestDashboardsOverTheConsoleAPI(t *testing.T) {
	f := newFixture(t)
	f.cleanupEvents()
	ctx := context.Background()
	u1, u2 := plugintest.NewUser(t, f.s.db), plugintest.NewUser(t, f.s.db)
	f.at = time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC)
	if _, err := f.s.session(ctx, u1, sessionIn{Event: "start", SessionID: "sess-1", Platform: "android", AppVersion: "2.0", Game: f.prefix}, settings.SourceClient); err != nil {
		t.Fatal(err)
	}
	if _, err := f.batch(u1, "dash-0001", `[{"name":"LevelStart","tags":{"level":"1"}},{"name":"DistanceOverall","tags":{"car":"c1","mission":"m1"},"value":100}]`); err != nil {
		t.Fatal(err)
	}
	f.at = f.at.Add(10 * time.Minute)
	if _, err := f.batch(u1, "dash-0002", `[{"name":"LevelEnd","tags":{"level":"1"}},{"name":"DistanceOverall","tags":{"car":"c2","mission":"m1"},"value":40}]`); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.session(ctx, u1, sessionIn{Event: "end", SessionID: "sess-1", Game: f.prefix}, settings.SourceClient); err != nil {
		t.Fatal(err)
	}
	if _, err := f.batch(u2, "dash-0003", `[{"name":"LevelStart","tags":{"level":"1"}},{"name":"DistanceOverall","tags":{"car":"c1","mission":"m2"},"value":60}]`); err != nil {
		t.Fatal(err)
	}
	f.at = f.at.Add(24 * time.Hour)
	if _, err := f.s.send(ctx, u1, mustSendIn(t, `{"batchId":"dash-0004","sessionId":"sess-2","game":"`+f.prefix+`","events":[{"name":"DistanceOverall","tags":{"car":"c1","mission":"m2"},"value":5}]}`)); err != nil {
		t.Fatal(err)
	}
	f.drain()
	q := "&game=" + f.prefix + "&from=2026-06-01&to=2026-06-02"
	ex := okc(t, f.s.explore, "GET", "/x?name=DistanceOverall&metric=sum&groupBy=tag:car,tag:mission"+q, nil, nil)
	rows := ex["rows"].([]any)
	if len(rows) != 3 || rows[0].(map[string]any)["value"].(float64) != 100 {
		t.Fatalf("explore %v", ex)
	}
	tot := okc(t, f.s.explore, "GET", "/x?name=DistanceOverall&metric=sum&tag.car=c1"+q, nil, nil)
	if tot["rows"].([]any)[0].(map[string]any)["value"].(float64) != 165 {
		t.Fatalf("filtered total %v", tot)
	}
	if code, _ := call(t, f.s.explore, "GET", "/x?metric=nope"+q, nil, nil); code != 400 {
		t.Fatalf("bad metric code %d", code)
	}
	act := okc(t, f.s.activity, "GET", "/x?"+q[1:], nil, nil)
	a := act["activity"].(map[string]any)
	if a["newUsers"].(float64) != 2 || a["dau"].(float64) != 1 || a["wau"].(float64) != 2 || len(act["topEvents"].([]any)) == 0 {
		t.Fatalf("activity %v", act)
	}
	ret := okc(t, f.s.retentionQuery, "GET", "/x?days=1"+q, nil, nil)
	sum := ret["summary"].([]any)[0].(map[string]any)
	if sum["cohort"].(float64) != 2 || sum["returned"].(float64) != 1 {
		t.Fatalf("retention %v", ret)
	}
	fn := okc(t, f.s.funnelSave, "POST", "/x", map[string]any{"name": "Level 1", "game": f.prefix, "windowSeconds": 3600, "steps": []map[string]any{{"name": "LevelStart", "tags": map[string][]string{"level": {"1"}}}, {"name": "LevelEnd"}}}, nil)
	id := fn["id"].(string)
	run := okc(t, f.s.funnelRun, "GET", "/x?from=2026-06-01&to=2026-06-02", nil, map[string]string{"id": id})
	steps := run["steps"].([]any)
	s0, s1 := steps[0].(map[string]any), steps[1].(map[string]any)
	if s0["users"].(float64) != 2 || s1["users"].(float64) != 1 || s1["fromFirst"].(float64) != 0.5 || s1["dropOff"].(float64) != 1 || s1["medianSeconds"].(float64) != 600 {
		t.Fatalf("funnel %v", run)
	}
	if code, _ := call(t, f.s.funnelSave, "POST", "/x", map[string]any{"name": "x", "steps": []map[string]any{{"name": "A"}}}, nil); code != 400 {
		t.Fatal("one step funnel accepted")
	}
	list := okc(t, f.s.funnelList, "GET", "/x", nil, nil)
	if len(list["items"].([]any)) < 1 {
		t.Fatal("funnel list")
	}
	var sid string
	_ = f.s.db.QueryRow(`SELECT id::text FROM insights_session WHERE user_id=$1 AND client_id='sess-1'`, u1).Scan(&sid)
	tl := okc(t, f.s.sessionTimeline, "GET", "/x", nil, map[string]string{"id": sid})
	evs := tl["events"].([]any)
	var sg string
	_ = f.s.db.QueryRow(`SELECT game FROM insights_event WHERE user_id=$1 AND name='session_start'`, u1).Scan(&sg)
	if sg != f.prefix {
		t.Fatalf("session event game %q", sg)
	}
	if tl["matchedBy"] != "session" || len(evs) != 6 || evs[0].(map[string]any)["name"] != "session_start" || evs[5].(map[string]any)["name"] != "session_end" {
		t.Fatalf("timeline %d %v", len(evs), tl)
	}
	vals := okc(t, f.s.valueList, "GET", "/x?kind=tag:car"+q, nil, nil)
	if len(vals["items"].([]any)) != 2 {
		t.Fatalf("values %v", vals)
	}
	evl := okc(t, f.s.eventList, "GET", "/x?limit=2&userId="+u1+q, nil, nil)
	if len(evl["items"].([]any)) != 2 || evl["next"] == "" {
		t.Fatalf("event list %v", evl)
	}
	next := okc(t, f.s.eventList, "GET", "/x?limit=50&userId="+u1+"&cursor="+url.QueryEscape(evl["next"].(string))+q, nil, nil)
	if len(next["items"].([]any)) != 5 {
		t.Fatalf("next page %d", len(next["items"].([]any)))
	}
	okc(t, f.s.funnelDelete, "DELETE", "/x", nil, map[string]string{"id": id})
	set := okc(t, f.s.eventSettingsPut, "PUT", "/x", map[string]any{"events": settings.DefaultEvents(), "storage": map[string]any{"keepDays": 30, "batchSize": 10, "maxPendingEvents": 5000, "deadAfter": 3}}, nil)
	if set["storage"].(map[string]any)["keepDays"].(float64) != 30 {
		t.Fatalf("settings %v", set)
	}
	if code, _ := call(t, f.s.eventSettingsPut, "PUT", "/x", map[string]any{"storage": map[string]any{"keepDays": 30, "batchSize": 0, "maxPendingEvents": 5000, "deadAfter": 3}}, nil); code != 400 {
		t.Fatal("bad settings accepted")
	}
}
