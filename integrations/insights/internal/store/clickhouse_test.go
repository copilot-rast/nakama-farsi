package store

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

type chCall struct {
	query  string
	params map[string]string
	body   string
	user   string
	key    string
}

type fakeCH struct {
	mu     sync.Mutex
	calls  []chCall
	answer func(q string) string
}

var placeholder = regexp.MustCompile(`\{(p\d+):[^}]+\}`)

func (f *fakeCH) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		c := chCall{params: map[string]string{}, user: r.Header.Get("X-ClickHouse-User"), key: r.Header.Get("X-ClickHouse-Key")}
		for k, v := range r.URL.Query() {
			c.params[k] = v[0]
		}
		if q := r.URL.Query().Get("query"); q != "" {
			c.query, c.body = q, string(raw)
		} else {
			c.query = string(raw)
		}
		for _, m := range placeholder.FindAllStringSubmatch(c.query, -1) {
			if _, ok := c.params["param_"+m[1]]; !ok {
				w.WriteHeader(400)
				_, _ = w.Write([]byte("Code: 456. Substitution " + m[1] + " is not set"))
				return
			}
		}
		f.mu.Lock()
		f.calls = append(f.calls, c)
		f.mu.Unlock()
		if strings.Contains(c.query, "boom") {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("Code: 60. Table does not exist"))
			return
		}
		_, _ = w.Write([]byte(f.answer(c.query)))
	}
}

func (f *fakeCH) last() chCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[len(f.calls)-1]
}

func TestClickHouseDriverAgainstAFakeServer(t *testing.T) {
	fake := &fakeCH{answer: func(q string) string {
		switch {
		case strings.Contains(q, "windowFunnel"):
			return `{"data":[[0,4],[1,3],[2,1]]}`
		case strings.Contains(q, "WITH u AS"):
			return `{"data":[["2026-06-02",2,1],["2026-06-01",3,2]]}`
		case strings.Contains(q, "SELECT toString(first)"):
			return `{"data":[["2026-06-01",3]]}`
		case strings.Contains(q, "uniqExactIf"):
			return `{"data":[[1,3,5,5]]}`
		case strings.Contains(q, "SELECT toString(day)"):
			return `{"data":[["2026-06-01",3,9]]}`
		case strings.Contains(q, "SELECT toString(id)"):
			return `{"data":[["0b945420-b1ee-5114-af35-3554bcd8706d","g","DistanceOverall",150,{"car":"c1"},"{\"weather\":\"rain\"}","u1","s1","2026-06-01 08:00:00.000","2026-06-01","2026-06-01 08:00:01.000","android","1.0","client",0,""]]}`
		case strings.Contains(q, "AS v, count() AS n"):
			return `{"data":[["c1",4],["c2",1]]}`
		case strings.Contains(q, "AS v, count(), uniqExact"):
			return `{"data":[["","c1","m1",150,1,1],["","c2","m1",100,2,1]]}`
		}
		return ""
	}}
	srv := httptest.NewServer(fake.handler(t))
	defer srv.Close()
	if _, err := NewClickHouse("ftp://x", "", "", "", "", nil); err == nil {
		t.Fatal("bad url accepted")
	}
	if _, err := NewClickHouse(srv.URL, "bad-db", "", "", "", nil); err != ErrBadQuery {
		t.Fatal("bad database accepted")
	}
	ch, err := NewClickHouse(srv.URL, "analytics", "", "reader", "pw", srv.Client())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := ch.Ensure(ctx); err != nil {
		t.Fatal(err)
	}
	if c := fake.last(); !strings.Contains(c.query, "ReplacingMergeTree") || !strings.Contains(c.query, "`analytics`.`insights_event`") || c.user != "reader" || c.key != "pw" || c.params["database"] != "analytics" {
		t.Fatalf("ensure %+v", c)
	}
	at := time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC)
	e := Event{ID: "0b945420-b1ee-5114-af35-3554bcd8706d", Name: "DistanceOverall", Value: num(150), Tags: map[string]string{"car": "c1"}, UserID: "u1", At: at, Day: "2026-06-01", ServerAt: at, Source: "client", Flagged: true}
	if err := ch.Write(ctx, []Event{e, e}); err != nil {
		t.Fatal(err)
	}
	w := fake.last()
	if !strings.HasPrefix(w.query, "INSERT INTO `analytics`.`insights_event` FORMAT JSONEachRow") || strings.Count(w.body, "\n") != 2 || !strings.Contains(w.body, `"tags":{"car":"c1"}`) || !strings.Contains(w.body, `"flagged":1`) || w.params["insert_deduplication_token"] == "" {
		t.Fatalf("write %+v", w)
	}
	g := "g"
	f := Filter{From: "2026-06-01", To: "2026-06-08", Game: &g, Names: []string{"DistanceOverall"}, Tags: map[string][]string{"mission": {"m1", "o'k"}}}
	rows, err := ch.Aggregate(ctx, AggQuery{Filter: f, Metric: MetricSum, GroupBy: []string{"tag:car", "tag:mission"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Keys[0] != "c1" || rows[0].Value != 150 || rows[1].Count != 2 {
		t.Fatalf("rows %+v", rows)
	}
	a := fake.last()
	if !strings.Contains(a.query, "FINAL") || !strings.Contains(a.query, "sum(ifNull(value, 0))") || !strings.Contains(a.query, "flagged = 0") || !strings.Contains(a.query, "mapContains(tags,") {
		t.Fatalf("aggregate sql %s", a.query)
	}
	foundArr := false
	for k, v := range a.params {
		if strings.HasPrefix(k, "param_") && v == `['m1','o\'k']` {
			foundArr = true
		}
	}
	if !foundArr {
		t.Fatalf("array param not escaped: %v", a.params)
	}
	if _, err := ch.Aggregate(ctx, AggQuery{Filter: f, Bucket: BucketHour}); err != nil || !strings.Contains(fake.last().query, "toStartOfHour(toTimeZone(at,") {
		t.Fatalf("hour bucket %v %s", err, fake.last().query)
	}
	act, err := ch.Activity(ctx, Filter{From: "2026-06-01", To: "2026-06-02", Game: &g})
	if err != nil || act.Days[0].DAU != 3 || act.Days[0].NewUsers != 3 || act.MAU != 5 || act.DAU != 1 || act.Events != 9 {
		t.Fatalf("activity %+v %v", act, err)
	}
	ret, err := ch.Retention(ctx, RetentionQuery{Filter: Filter{From: "2026-06-01", To: "2026-06-02"}, Days: []int{1}, Today: "2026-06-02"})
	if err != nil || len(ret.Cohorts) != 2 || ret.Cohorts[0].Cells[0].Returned != nil || *ret.Cohorts[1].Cells[0].Returned != 2 || ret.Summary[0].Cohort != 3 {
		t.Fatalf("retention %+v %v", ret, err)
	}
	if !strings.Contains(fake.last().query, "a.day = addDays(u.first, 1)") {
		t.Fatalf("retention sql %s", fake.last().query)
	}
	steps, err := ch.Funnel(ctx, FunnelQuery{Filter: Filter{From: "2026-06-01", To: "2026-06-02"}, Steps: []Step{{Name: "A", Tags: map[string][]string{"level": {"1"}}}, {Name: "B"}}, Window: time.Hour})
	if err != nil || steps[0].Users != 4 || steps[1].Users != 1 {
		t.Fatalf("funnel %+v %v", steps, err)
	}
	if !strings.Contains(fake.last().query, "windowFunnel(3600)") {
		t.Fatalf("funnel sql %s", fake.last().query)
	}
	evs, err := ch.Events(ctx, EventQuery{Filter: Filter{UserID: "u1"}, BeforeAt: &at, BeforeID: e.ID})
	if err != nil || len(evs) != 1 || evs[0].Tags["car"] != "c1" || evs[0].Fields["weather"] != "rain" || *evs[0].Value != 150 || !evs[0].At.Equal(at) {
		t.Fatalf("events %+v %v", evs, err)
	}
	vals, err := ch.Values(ctx, ValuesQuery{Kind: "tag:car"})
	if err != nil || len(vals) != 2 || vals[0].Count != 4 {
		t.Fatalf("values %+v %v", vals, err)
	}
	if _, err := ch.Sweep(ctx, "2026-01-01"); err != nil || !strings.Contains(fake.last().query, "ALTER TABLE `analytics`.`insights_event` DELETE WHERE day <") {
		t.Fatalf("sweep %v", err)
	}
	if err := ch.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	bad, _ := NewClickHouse(srv.URL, "analytics", "boom", "", "", srv.Client())
	if err := bad.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	if err := bad.Ensure(ctx); err == nil || !strings.Contains(err.Error(), "HTTP 500") {
		t.Fatalf("error not surfaced: %v", err)
	}
}
