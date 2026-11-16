package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/insights/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

type fixture struct {
	t      *testing.T
	s      *service
	prefix string
	at     time.Time
}

func newFixture(t *testing.T) *fixture {
	db := plugintest.OpenDB(t)
	store, err := objectstore.New(objectstore.Config{Driver: objectstore.DriverLocal, LocalDir: t.TempDir()}, nil)
	if err != nil {
		t.Fatal(err)
	}
	f := &fixture{t: t, s: &service{db: db, logger: plugintest.Logger{}, store: store}, prefix: plugintest.Key("it"), at: time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)}
	f.s.now = func() time.Time { return f.at }
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM insights_heatmap_map WHERE key LIKE $1`, f.prefix+"%")
		_, _ = db.Exec(`DELETE FROM insights_settings WHERE id=1`)
		settings.Invalidate()
	})
	f.settings(func(*settings.All) {})
	return f
}

func (f *fixture) settings(change func(*settings.All)) {
	f.t.Helper()
	cfg := settings.Defaults()
	change(&cfg)
	if err := settings.Save(context.Background(), f.s.db, cfg, "test"); err != nil {
		f.t.Fatal(err)
	}
}

func errCode(err error) string {
	if e, ok := apierr.As(err); ok {
		return e.Code
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func (f *fixture) send(user, event, sid string, extra string) map[string]any {
	f.t.Helper()
	out, err := f.s.rpcSession(context.Background(), user, `{"event":"`+event+`","sessionId":"`+sid+`","platform":"`+f.prefix+`","appVersion":"1.4.0"`+extra+`}`)
	if err != nil {
		f.t.Fatalf("%s %s: %v", event, sid, err)
	}
	return out.(map[string]any)
}

func get(t *testing.T, h http.HandlerFunc, target string, vars map[string]string) map[string]any {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, target, nil)
	for k, v := range vars {
		r.SetPathValue(k, v)
	}
	w := httptest.NewRecorder()
	h(w, r)
	if w.Code != 200 {
		t.Fatalf("%s = %d %s", target, w.Code, w.Body.String())
	}
	out := map[string]any{}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestSessionLifecycle(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	u := plugintest.NewUser(t, f.s.db)
	out := f.send(u, "start", "s1", "")
	if out["created"] != true || out["status"] != "open" {
		t.Fatalf("start = %+v", out)
	}
	f.at = f.at.Add(time.Minute)
	if out := f.send(u, "ping", "s1", ""); out["created"] != false {
		t.Fatalf("ping = %+v", out)
	}
	f.at = f.at.Add(2 * time.Minute)
	out = f.send(u, "end", "s1", `,"seconds":150`)
	if out["status"] != "ended" || out["seconds"] != int64(150) {
		t.Fatalf("end = %+v", out)
	}
	if out := f.send(u, "end", "s1", ""); out["duplicate"] != true || out["seconds"] != int64(150) {
		t.Fatalf("duplicate end = %+v", out)
	}
	out = f.send(u, "end", "s2", `,"seconds":99999999`)
	if out["seconds"] != int64(6*3600) {
		t.Fatalf("end without start is capped = %+v", out)
	}
	if _, err := f.s.rpcSession(ctx, u, `{"event":"jump","sessionId":"s3"}`); errCode(err) != "ANALYTICS_INVALID_SESSION" {
		t.Fatalf("bad event = %v", err)
	}
	if _, err := f.s.rpcSession(ctx, u, `{"event":"start","sessionId":"bad id!"}`); errCode(err) != "ANALYTICS_INVALID_SESSION" {
		t.Fatalf("bad id = %v", err)
	}
	f.send(u, "start", "s4", "")
	f.at = f.at.Add(30 * time.Second)
	f.send(u, "start", "s5", "")
	var ended bool
	_ = f.s.db.QueryRow(`SELECT ended_at IS NOT NULL FROM insights_session WHERE user_id=$1 AND client_id='s4'`, u).Scan(&ended)
	if !ended {
		t.Fatal("older open session not closed by a new start")
	}
	f.at = f.at.Add(10 * time.Minute)
	if n, err := f.s.closeIdle(ctx, f.s.settings(ctx)); err != nil || n < 1 {
		t.Fatalf("idle close = %d %v", n, err)
	}
	var secs int64
	_ = f.s.db.QueryRow(`SELECT seconds FROM insights_session WHERE user_id=$1 AND client_id='s5'`, u).Scan(&secs)
	if secs != 0 {
		t.Fatalf("idle session seconds = %d", secs)
	}
	var daySessions, daySeconds int64
	_ = f.s.db.QueryRow(`SELECT sessions, seconds FROM insights_player_day WHERE user_id=$1`, u).Scan(&daySessions, &daySeconds)
	if daySessions != 4 || daySeconds != 150+6*3600 {
		t.Fatalf("day = %d sessions %d seconds", daySessions, daySeconds)
	}
	f.settings(func(c *settings.All) { c.Sessions.StartsPerHour = 3 })
	if _, err := f.s.rpcSession(ctx, u, `{"event":"start","sessionId":"s6"}`); errCode(err) != "ANALYTICS_RATE_LIMITED" {
		t.Fatalf("rate = %v", err)
	}
	f.settings(func(c *settings.All) { c.Sessions.Source = settings.SourceSocket })
	if out := f.send(u, "start", "s7", ""); out["ignored"] != true {
		t.Fatalf("socket mode = %+v", out)
	}
}

func TestRetentionAndOverview(t *testing.T) {
	f := newFixture(t)
	day0 := f.at
	users := []string{plugintest.NewUser(t, f.s.db), plugintest.NewUser(t, f.s.db), plugintest.NewUser(t, f.s.db)}
	play := func(u string, day int, n int) {
		f.at = day0.AddDate(0, 0, day).Add(time.Duration(n) * time.Hour)
		sid := fmt.Sprintf("d%d-%d", day, n)
		f.send(u, "start", sid, "")
		f.at = f.at.Add(4 * time.Minute)
		f.send(u, "end", sid, "")
	}
	for _, u := range users {
		play(u, 0, 0)
	}
	play(users[0], 1, 0)
	play(users[0], 7, 0)
	play(users[1], 7, 1)
	play(users[2], 3, 0)
	f.at = day0.AddDate(0, 0, 8)
	q := "/overview?from=2026-06-01&to=2026-06-08&platform=" + f.prefix
	out := get(t, f.s.overview, q, nil)
	ret := out["retention"].([]any)
	d1, d7 := ret[0].(map[string]any), ret[1].(map[string]any)
	if d1["cohort"] != 3.0 || d1["returned"] != 1.0 || d7["returned"] != 2.0 {
		t.Fatalf("classic retention = %v", ret)
	}
	tt := out["totals"].(map[string]any)
	if tt["sessions"] != 7.0 || tt["players"] != 3.0 || tt["newPlayers"] != 3.0 || tt["avgSeconds"] != 240.0 {
		t.Fatalf("totals = %v", tt)
	}
	series := out["series"].([]any)
	if series[0].(map[string]any)["dau"] != 3.0 || series[7].(map[string]any)["dau"] != 2.0 || series[6].(map[string]any)["dau"] != 0.0 {
		t.Fatalf("series = %v", series)
	}
	act := out["active"].(map[string]any)
	if act["mau"] != 3.0 || act["wau"] != 3.0 || act["dau"] != 2.0 {
		t.Fatalf("active = %v", act)
	}
	cohorts := out["cohorts"].([]any)
	if len(cohorts) != 1 || cohorts[0].(map[string]any)["size"] != 3.0 {
		t.Fatalf("cohorts = %v", cohorts)
	}
	cells := cohorts[0].(map[string]any)["cells"].([]any)
	if cells[2].(map[string]any)["returned"] != nil {
		t.Fatalf("D30 must be empty before day 30: %v", cells)
	}
	f.settings(func(c *settings.All) { c.Retention.Mode = settings.ModeRolling })
	out = get(t, f.s.overview, q, nil)
	if r := out["retention"].([]any)[0].(map[string]any); r["returned"] != 3.0 {
		t.Fatalf("rolling D1 = %v", r)
	}
	list := get(t, f.s.sessionList, "/sessions?platform="+f.prefix+"&limit=5", nil)
	if list["total"] != 7.0 || len(list["items"].([]any)) != 5 {
		t.Fatalf("sessions = %v", list["total"])
	}
}

func gz(s string) string {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write([]byte(s))
	_ = w.Close()
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

func TestHeatmapUploadCellsAndTrace(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	u := plugintest.NewUser(t, f.s.db)
	m := f.prefix + "_desert"
	var csv strings.Builder
	csv.WriteString("t,x,z,player,kmh\n")
	for i := 0; i < 40; i++ {
		fmt.Fprintf(&csv, "%.1f,%d,%d,%s,%d\n", float64(i)*0.5, i, i*2, []string{"me", "bot1"}[i%2], 100+i)
	}
	csv.WriteString("1,2,3,me,9999\nbroken line\n")
	payload, _ := json.Marshal(map[string]any{"map": m, "uploadId": "race-1", "format": "csv.gz", "data": gz(csv.String())})
	out, err := f.s.rpcHeatmap(ctx, u, string(payload))
	if err != nil {
		t.Fatal(err)
	}
	res := out.(map[string]any)
	if res["accepted"] != 40 || res["dropped"] != 2 || res["players"] != 2 {
		t.Fatalf("upload = %+v", res)
	}
	if out, _ := f.s.rpcHeatmap(ctx, u, string(payload)); out.(map[string]any)["duplicate"] != true {
		t.Fatalf("resend = %+v", out)
	}
	part2, _ := json.Marshal(map[string]any{"map": m, "uploadId": "race-1", "part": 1, "points": [][]any{{20, 1, 1, "me", 50}, {20.5, 2, 1, "me", 60}}})
	if _, err := f.s.rpcHeatmap(ctx, u, string(part2)); err != nil {
		t.Fatal(err)
	}
	cells := get(t, f.s.mapCells, "/cells?from=2026-06-01&to=2026-06-01", map[string]string{"key": m})
	tt := cells["totals"].(map[string]any)
	if tt["points"] != 42.0 || tt["uploads"] != 2.0 || cells["size"] != 5.0 {
		t.Fatalf("cells = %v", tt)
	}
	first := cells["cells"].([]any)[0].([]any)
	if first[0] != 0.0 || first[1] != 0.0 || first[2] != 5.0 {
		t.Fatalf("top cell = %v", first)
	}
	ups := get(t, f.s.mapUploads, "/uploads", map[string]string{"key": m})
	items := ups["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("uploads = %v", items)
	}
	var id string
	for _, it := range items {
		if it.(map[string]any)["part"] == 0.0 {
			id = it.(map[string]any)["id"].(string)
		}
	}
	tr := get(t, f.s.uploadTrace, "/trace", map[string]string{"id": id})
	if tr["points"] != 40.0 || len(tr["players"].([]any)) != 2 {
		t.Fatalf("trace = %v", tr["points"])
	}
	f.settings(func(c *settings.All) { c.Heatmap.UnknownMaps = settings.MapsReject; c.Heatmap.MaxPoints = 10 })
	other, _ := json.Marshal(map[string]any{"map": f.prefix + "_x", "uploadId": "r2", "data": "1,2,3,me,4"})
	if _, err := f.s.rpcHeatmap(ctx, u, string(other)); errCode(err) != "HEATMAP_MAP_UNKNOWN" {
		t.Fatalf("unknown map = %v", err)
	}
	big, _ := json.Marshal(map[string]any{"map": m, "uploadId": "r3", "data": csv.String()})
	if _, err := f.s.rpcHeatmap(ctx, u, string(big)); errCode(err) != "HEATMAP_TOO_LARGE" {
		t.Fatalf("too many = %v", err)
	}
	plugintest.Exec(t, f.s.db, `UPDATE insights_heatmap_map SET enabled=false WHERE key=$1`, m)
	small, _ := json.Marshal(map[string]any{"map": m, "uploadId": "r4", "data": "1,2,3,me,4"})
	if _, err := f.s.rpcHeatmap(ctx, u, string(small)); errCode(err) != "HEATMAP_MAP_DISABLED" {
		t.Fatalf("disabled map = %v", err)
	}
	f.settings(func(c *settings.All) { c.Heatmap.Enabled = false })
	if _, err := f.s.rpcHeatmap(ctx, u, string(small)); errCode(err) != "HEATMAP_DISABLED" {
		t.Fatalf("disabled = %v", err)
	}
	cfg, _ := f.s.rpcConfig(ctx, u, "")
	if cfg.(map[string]any)["heatmap"].(map[string]any)["enabled"] != false {
		t.Fatalf("config = %v", cfg)
	}
}
