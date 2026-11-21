package importer

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func serviceAccountJSON(t *testing.T, tokenURL string) (string, *rsa.PublicKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalPKCS8PrivateKey(key)
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	raw, _ := json.Marshal(map[string]string{"type": "service_account", "client_email": "sa@test.iam", "private_key": string(pemKey), "token_uri": tokenURL})
	return string(raw), &key.PublicKey
}

func tokenHandler(t *testing.T, pub *rsa.PublicKey, wantScope string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		parts := strings.Split(r.Form.Get("assertion"), ".")
		if r.Form.Get("grant_type") != "urn:ietf:params:oauth:grant-type:jwt-bearer" || len(parts) != 3 {
			w.WriteHeader(400)
			return
		}
		sig, _ := base64.RawURLEncoding.DecodeString(parts[2])
		sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
		if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, sum[:], sig); err != nil {
			w.WriteHeader(401)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
			return
		}
		claims, _ := base64.RawURLEncoding.DecodeString(parts[1])
		var c map[string]any
		_ = json.Unmarshal(claims, &c)
		if c["scope"] != wantScope || c["iss"] != "sa@test.iam" {
			w.WriteHeader(403)
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"tok-1","expires_in":3600}`))
	}
}

type collector struct {
	mu    sync.Mutex
	items []Item
	fail  int
}

func (c *collector) emit(_ context.Context, items []Item) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fail > 0 {
		c.fail--
		return errors.New("outbox down")
	}
	c.items = append(c.items, items...)
	return nil
}

func row(vals ...any) map[string]any {
	f := []map[string]any{}
	for _, v := range vals {
		f = append(f, map[string]any{"v": v})
	}
	return map[string]any{"f": f}
}

func TestGA4BigQueryPagesDaysAndMaps(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	secret, pub := serviceAccountJSON(t, srv.URL+"/token")
	mux.HandleFunc("/token", tokenHandler(t, pub, scopeBigQuery))
	queries := []string{}
	params := `[{"key":"ga_session_id","value":{"int_value":"77"}},{"key":"car","value":{"string_value":"SantaFe"}},{"key":"distance","value":{"double_value":12.5}}]`
	mux.HandleFunc("/bigquery/v2/projects/proj-1/queries", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok-1" {
			w.WriteHeader(401)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		q := body["query"].(string)
		queries = append(queries, q)
		if strings.Contains(q, "events_20261102") {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"error":{"code":404,"message":"Not found: Table"}}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jobComplete": true, "jobReference": map[string]string{"jobId": "job-1", "location": "EU"}, "pageToken": "p2",
			"rows": []any{row("1793500000000000", "level_up", "", "pseudo-1", "ANDROID", "1.4.0", "5", "77", params)}})
	})
	mux.HandleFunc("/bigquery/v2/projects/proj-1/queries/job-1", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("pageToken") != "p2" || r.URL.Query().Get("location") != "EU" {
			w.WriteHeader(400)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"jobComplete": true, "rows": []any{row("1793500001000000", "first_open", "user-9", "pseudo-2", "IOS", "1.4.1", "6", nil, `[]`), row("bad", "x", "", "", "", "", "", nil, "[]")}})
	})
	cfg := json.RawMessage(`{"projectId":"proj-1","dataset":"analytics_1","startDate":"2026-11-01","daysPerRun":2,"lagDays":1,"paramTags":["car"],"valueParam":"distance","nameMap":{"level_up":"LevelUp"},"apiBase":"` + srv.URL + `","tokenUrl":"` + srv.URL + `/token"}`)
	imp := All["ga4_bigquery"]
	if _, err := imp.Check(cfg, secret); err != nil {
		t.Fatal(err)
	}
	if _, err := imp.Check(json.RawMessage(`{"projectId":"p","dataset":"d","apiBase":"http://example.com"}`), ""); err == nil {
		t.Fatal("plain http to a remote host accepted")
	}
	col := &collector{}
	now := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	out, err := imp.Pull(context.Background(), Env{HTTP: srv.Client(), Now: now, Config: cfg, Secret: secret, Emit: col.emit})
	if err != nil {
		t.Fatal(err)
	}
	if out.Read != 3 || out.Skipped != 1 || len(col.items) != 2 || !out.More {
		t.Fatalf("outcome %+v items %d", out, len(col.items))
	}
	var cur bqCursor
	_ = json.Unmarshal(out.Cursor, &cur)
	if cur.Day != "2026-11-03" || cur.JobID != "" {
		t.Fatalf("cursor %+v", cur)
	}
	a := col.items[0]
	if a.Raw.Name != "LevelUp" || a.UserID != "ga4:pseudo-1" || a.SessionID != "77" || a.Platform != "android" || a.AppVersion != "1.4.0" || a.Raw.Value != 12.5 || !strings.Contains(string(a.Raw.Tags), `"car":"SantaFe"`) {
		t.Fatalf("item %+v tags %s", a, a.Raw.Tags)
	}
	if !a.At.Equal(time.UnixMicro(1793500000000000)) || col.items[1].UserID != "user-9" || col.items[1].Raw.Name != "first_open" {
		t.Fatalf("time/user %+v", col.items[1])
	}
	if len(queries) != 2 || !strings.Contains(queries[0], "`proj-1.analytics_1.events_20261101`") {
		t.Fatalf("queries %v", queries)
	}
	bad := json.RawMessage(strings.Replace(string(cfg), srv.URL+"/token", srv.URL+"/nope", 1))
	if _, err := imp.Pull(context.Background(), Env{HTTP: srv.Client(), Now: now, Config: bad, Secret: secret, Emit: col.emit}); err == nil {
		t.Fatal("token failure not reported")
	}
}

func TestGA4DataAPIAggregates(t *testing.T) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	secret, pub := serviceAccountJSON(t, srv.URL+"/token")
	mux.HandleFunc("/token", tokenHandler(t, pub, scopeAnalytics))
	mux.HandleFunc("/v1beta/properties/123:runReport", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		day := body["dateRanges"].([]any)[0].(map[string]any)["startDate"].(string)
		_ = json.NewEncoder(w).Encode(map[string]any{"rowCount": 1, "rows": []any{map[string]any{
			"dimensionValues": []any{map[string]string{"value": strings.ReplaceAll(day, "-", "")}, map[string]string{"value": "purchase"}, map[string]string{"value": "Android"}},
			"metricValues":    []any{map[string]string{"value": "42"}}}}})
	})
	cfg := json.RawMessage(`{"propertyId":"123","startDate":"2026-11-18","daysPerRun":5,"dimensions":["platform"],"apiBase":"` + srv.URL + `","tokenUrl":"` + srv.URL + `/token"}`)
	col := &collector{}
	out, err := All["ga4_data"].Pull(context.Background(), Env{HTTP: srv.Client(), Now: time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC), Config: cfg, Secret: secret, Emit: col.emit})
	if err != nil {
		t.Fatal(err)
	}
	if len(col.items) != 3 || out.More || col.items[0].Raw.Value != 42.0 || col.items[0].UserID != "ga4:aggregate" || !strings.Contains(string(col.items[0].Raw.Tags), `"platform":"Android"`) {
		t.Fatalf("data %+v %+v", out, col.items)
	}
	if col.items[0].Key == col.items[1].Key {
		t.Fatal("keys must differ per day")
	}
}

func gz(t *testing.T, lines ...string) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = io.WriteString(w, strings.Join(lines, "\n")+"\n")
	_ = w.Close()
	return b.Bytes()
}

func TestGameAnalyticsRawExportFromS3(t *testing.T) {
	files := map[string][]byte{
		"exports/2026/11/20/a.json.gz": gz(t,
			`{"data":{"category":"design","event_id":"Vehicle:Distance:SUV","value":120,"user_id":"u1","session_id":"s1","platform":"android","build":"1.0","client_ts":1793500000}}`,
			`{"data":{"category":"progression","event_id":"Complete:World1:Level2","score":300,"user_id":"u1","client_ts":1793500010}}`,
			`{"data":{"category":"resource","event_id":"Sink:gems:Shop:Car","amount":20,"user_id":"u1","client_ts":1793500020}}`,
			`{"data":{"category":"business","event_id":"Coins:Pack1","amount":99,"currency":"USD","cart_type":"shop","user_id":"u1","client_ts":1793500030}}`,
			`{"data":{"category":"error","severity":"error","message":"boom","user_id":"u1","client_ts":1793500040}}`,
			`{"data":{"category":"user","user_id":"u1","session_id":"s1","client_ts":1793500000}}`,
			`{"data":{"category":"session_end","length":95,"user_id":"u1","session_id":"s1","client_ts":1793500100}}`,
			`{"data":{"category":"ads","user_id":"u1","client_ts":1793500100}}`,
			`not json`),
		"exports/2026/11/21/b.json.gz": gz(t, `{"data":{"category":"design","event_id":"Boost","user_id":"u2","client_ts":1793600000}}`),
	}
	var signed int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=AKIA1/") {
			signed++
		} else {
			w.WriteHeader(403)
			return
		}
		if r.URL.Path == "/ga-bucket" {
			after := r.URL.Query().Get("start-after")
			var b strings.Builder
			b.WriteString(`<?xml version="1.0"?><ListBucketResult>`)
			for _, k := range []string{"exports/2026/11/20/a.json.gz", "exports/2026/11/21/b.json.gz"} {
				if k > after {
					fmt.Fprintf(&b, "<Contents><Key>%s</Key><Size>%d</Size></Contents>", k, len(files[k]))
				}
			}
			b.WriteString(`<IsTruncated>false</IsTruncated></ListBucketResult>`)
			_, _ = w.Write([]byte(b.String()))
			return
		}
		data, ok := files[strings.TrimPrefix(r.URL.Path, "/ga-bucket/")]
		if !ok {
			w.WriteHeader(404)
			return
		}
		_, _ = w.Write(data)
	}))
	defer srv.Close()
	cfg := json.RawMessage(`{"endpoint":"` + srv.URL + `","bucket":"ga-bucket","prefix":"exports/","accessKey":"AKIA1"}`)
	imp := All["gameanalytics"]
	if _, err := imp.Check(cfg, ""); err != nil {
		t.Fatal(err)
	}
	col := &collector{fail: 1}
	now := time.Date(2026, 11, 21, 10, 0, 0, 0, time.UTC)
	out, err := imp.Pull(context.Background(), Env{HTTP: srv.Client(), Now: now, Config: cfg, Secret: "s3cret", Emit: col.emit})
	if err == nil {
		t.Fatal("emit failure not reported")
	}
	out, err = imp.Pull(context.Background(), Env{HTTP: srv.Client(), Now: now, Config: cfg, Secret: "s3cret", Cursor: out.Cursor, Emit: col.emit})
	if err != nil {
		t.Fatal(err)
	}
	if len(col.items) != 8 || out.Skipped != 0 && out.Skipped != 2 {
		t.Fatalf("items %d outcome %+v", len(col.items), out)
	}
	byType := map[string]Item{}
	for _, it := range col.items {
		k := it.Raw.Type
		if k == "" {
			k = it.Raw.Name
		}
		byType[k] = it
	}
	if d := byType["design"]; d.Raw.ID != "Boost" && d.Raw.ID != "Vehicle:Distance:SUV" || d.UserID == "" || !strings.HasPrefix(d.UserID, "ga:") {
		t.Fatalf("design %+v", d)
	}
	if p := byType["progression"]; p.Raw.Status != "Complete" || p.Raw.P2 != "Level2" || p.Raw.Score != 300.0 {
		t.Fatalf("progression %+v", p)
	}
	if r := byType["resource"]; r.Raw.Flow != "Sink" || r.Raw.Currency != "gems" || r.Raw.Item != "Car" {
		t.Fatalf("resource %+v", r)
	}
	if b := byType["business"]; b.Raw.Currency != "USD" || b.Raw.ItemType != "Coins" || b.Raw.CartType != "shop" {
		t.Fatalf("business %+v", b)
	}
	if e := byType["session_end"]; e.Raw.Value != 95.0 || e.SessionID != "s1" {
		t.Fatalf("session end %+v", e)
	}
	var cur gaCursor
	_ = json.Unmarshal(out.Cursor, &cur)
	if cur.After != "exports/2026/11/21/b.json.gz" || cur.Line != 0 {
		t.Fatalf("cursor %+v", cur)
	}
	again, _ := imp.Pull(context.Background(), Env{HTTP: srv.Client(), Now: now, Config: cfg, Secret: "s3cret", Cursor: out.Cursor, Emit: col.emit})
	if again.Read != 0 || len(col.items) != 8 {
		t.Fatalf("second pull read %d", again.Read)
	}
	if signed == 0 {
		t.Fatal("requests not signed")
	}
}

func TestNameHelpers(t *testing.T) {
	if GroupKey("Car Class") != "car_class" || GroupKey("1x") != "p_1x" || EventName("level up!") != "level_up_" || EventName("1st") != "e_1st" || TagID("a:b") != "a_b" {
		t.Fatal("helpers")
	}
}
