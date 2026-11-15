package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

type broadcast struct {
	subject string
	content map[string]any
	code    int
}

var broadcastMu sync.Mutex
var broadcasts []broadcast

func (n *fakeNakama) NotificationSendAll(ctx context.Context, subject string, content map[string]interface{}, code int, persistent bool) error {
	broadcastMu.Lock()
	defer broadcastMu.Unlock()
	broadcasts = append(broadcasts, broadcast{subject, content, code})
	return nil
}

func resetBroadcasts() int {
	broadcastMu.Lock()
	defer broadcastMu.Unlock()
	n := len(broadcasts)
	broadcasts = nil
	return n
}

func newsHarness(t *testing.T) *harness {
	h := newHarness(t)
	for _, q := range []string{
		`TRUNCATE publication_news_ingest_log, publication_news_replay, publication_news_reaction, publication_news_source CASCADE`,
		`DELETE FROM publication_news_reaction_type WHERE key NOT IN ('like','love','haha','wow','sad')`,
		`UPDATE publication_news_reaction_type SET is_active=true`,
		`DELETE FROM base_tag WHERE key LIKE 'nt_%'`,
		`DELETE FROM cluster_rate_limit WHERE key LIKE 'news-%'`,
		`INSERT INTO publication_news_settings (id) VALUES (1) ON CONFLICT (id) DO UPDATE SET comments_per_minute=5, reactions_per_minute=30, unread_scope='signup', show_expired=false,
			fill_default_lang=true, realtime_on_publish=true, notify_category_key='news', log_retention_days=30, log_payload_bytes=16384`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	resetBroadcasts()
	return h
}

func (h *harness) newsTag(kind, key string) string {
	h.t.Helper()
	var id string
	if err := h.s.db.QueryRowContext(h.ctx, `INSERT INTO base_tag (id, key_type, key, name) VALUES (gen_random_uuid(),$1,$2,'{"fa":"برچسب","en":"Tag"}') RETURNING id::text`, kind, key).Scan(&id); err != nil {
		h.t.Fatal(err)
	}
	return id
}

func (h *harness) ingest(key string, headers map[string]string, body string) (int, map[string]any) {
	h.t.Helper()
	r := mux.NewRouter()
	r.HandleFunc(ingestPrefix+"{source}", h.s.ingestHTTP).Methods(http.MethodPost)
	req := httptest.NewRequest(http.MethodPost, ingestPrefix+key, strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func (h *harness) status(id string) (string, bool) {
	h.t.Helper()
	var st string
	var notified bool
	if err := h.s.db.QueryRowContext(h.ctx, `SELECT status, notified_at IS NOT NULL FROM publication_news WHERE id=$1`, id).Scan(&st, &notified); err != nil {
		h.t.Fatal(err)
	}
	return st, notified
}

func TestNewsStatusTagsReactionsUnread(t *testing.T) {
	h := newsHarness(t)
	u := h.user(nil, "", "fa")
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE users SET create_time = now() - interval '10 days' WHERE id=$1`, u); err != nil {
		t.Fatal(err)
	}
	tg := h.newsTag("news", "nt_update")
	other := h.newsTag("product", "nt_shop")
	if code, out := h.api(http.MethodPost, "/news", map[string]any{"title": map[string]string{"fa": "x"}, "body": map[string]string{"fa": "x"}, "tagIds": []string{other}}); code != 400 || out["error"] != "TAG_KIND_MISMATCH" {
		t.Fatalf("tag kind = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/news", map[string]any{"title": map[string]string{"fa": "x"}, "body": map[string]string{"fa": "x"}, "status": "nope"}); code != 400 || out["error"] != "INVALID_STATUS" {
		t.Fatalf("status = %d %v", code, out)
	}
	past := time.Now().Add(-time.Hour).Format(time.RFC3339)
	if code, out := h.api(http.MethodPost, "/news", map[string]any{"title": map[string]string{"fa": "x"}, "body": map[string]string{"fa": "x"}, "publishedAt": past, "expiresAt": past}); code != 400 || out["error"] != "EXPIRES_BEFORE_PUBLISH" {
		t.Fatalf("expires = %d %v", code, out)
	}
	rich := h.news(map[string]any{"title": map[string]string{"fa": "غنی"}, "bodyFormat": "html", "body": map[string]string{"fa": `<p style="text-align: right">سلام <script>x</script><u>زیر</u></p><table><tbody><tr><td>۱</td></tr></tbody></table>`},
		"tagIds": []string{tg}, "status": "published"})
	draft := h.news(map[string]any{"title": map[string]string{"fa": "پیش‌نویس"}, "body": map[string]string{"fa": "x"}, "status": "draft"})
	pending := h.news(map[string]any{"title": map[string]string{"fa": "در انتظار"}, "body": map[string]string{"fa": "x"}, "status": "pending_review"})
	expiring := h.news(map[string]any{"title": map[string]string{"fa": "تمام‌شده"}, "body": map[string]string{"fa": "x"}, "publishedAt": time.Now().Add(-2 * time.Hour).Format(time.RFC3339), "expiresAt": time.Now().Add(time.Hour).Format(time.RFC3339)})
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_news SET expires_at=now() - interval '1 minute' WHERE id=$1`, expiring); err != nil {
		t.Fatal(err)
	}
	console := h.mustAPI(http.StatusOK, http.MethodGet, "/news/"+rich, nil)
	if console["status"] != "published" || console["approvedBy"] != "admin" || console["createdBy"] != "admin" || console["bodyFormat"] != "html" {
		t.Fatalf("console news = %v", console)
	}
	if body := console["body"].(map[string]any)["fa"]; body != `<p style="text-align: right">سلام <u>زیر</u></p><table><tbody><tr><td>۱</td></tr></tbody></table>` {
		t.Fatalf("sanitized body = %v", body)
	}
	if got := h.mustAPI(http.StatusOK, http.MethodGet, "/news?status=pending_review", nil); num(got["total"]) != 1 {
		t.Fatalf("status filter = %v", got)
	}
	if got := h.mustAPI(http.StatusOK, http.MethodGet, "/news?tagId="+tg, nil); num(got["total"]) != 1 {
		t.Fatalf("tag filter = %v", got)
	}
	if got := h.mustAPI(http.StatusOK, http.MethodGet, "/news?expired=true", nil); num(got["total"]) != 1 {
		t.Fatalf("expired filter = %v", got)
	}

	list := items(h.mustRPC(u, "publication_news_list", map[string]any{}))
	if len(list) != 1 || list[0].(map[string]any)["id"] != rich {
		t.Fatalf("visible = %v", list)
	}
	first := list[0].(map[string]any)
	if tags := first["tags"].([]any); len(tags) != 1 || tags[0].(map[string]any)["key"] != "nt_update" {
		t.Fatalf("tags = %v", first)
	}
	if got := items(h.mustRPC(u, "publication_news_list", map[string]any{"tagKeys": []string{"NT_UPDATE"}})); len(got) != 1 {
		t.Fatalf("tag filter = %v", got)
	}
	if got := items(h.mustRPC(u, "publication_news_list", map[string]any{"tagKey": "nt_missing"})); len(got) != 0 {
		t.Fatalf("missing tag = %v", got)
	}
	if got := items(h.mustRPC(u, "publication_news_list", map[string]any{"expired": "include"})); len(got) != 1 {
		t.Fatalf("expired hidden while the setting is off = %v", got)
	}
	h.mustAPI(http.StatusOK, http.MethodPut, "/news-settings", map[string]any{"commentsPerMinute": 5, "reactionsPerMinute": 4, "unreadScope": "signup", "showExpired": true, "fillDefaultLang": true,
		"realtimeOnPublish": true, "notifyCategoryKey": "news", "logRetentionDays": 30, "logPayloadBytes": 1000})
	if got := items(h.mustRPC(u, "publication_news_list", map[string]any{"expired": "only"})); len(got) != 1 || got[0].(map[string]any)["expired"] != true {
		t.Fatalf("expired only = %v", got)
	}
	if code := h.rpcCode(u, "publication_news_react", map[string]any{"newsId": expiring, "reaction": "like"}); code != "NEWS_EXPIRED" {
		t.Fatalf("react expired = %s", code)
	}
	get := h.mustRPC(u, "publication_news_get", map[string]any{"id": rich})
	if !strings.Contains(get["bodyHtml"].(string), "<table>") || get["myReaction"] != nil || len(get["reactionTypes"].([]any)) != 5 {
		t.Fatalf("get = %v", get)
	}
	for _, id := range []string{draft, pending} {
		if code := h.rpcCode(u, "publication_news_get", map[string]any{"id": id}); code != "NEWS_NOT_FOUND" {
			t.Fatalf("hidden %s = %s", id, code)
		}
	}

	if got := h.mustRPC(u, "publication_news_unread", map[string]any{}); num(got["unread"]) != 1 {
		t.Fatalf("unread = %v", got)
	}
	r := h.mustRPC(u, "publication_news_react", map[string]any{"newsId": rich, "reaction": "like"})
	if r["myReaction"] != "like" || num(r["reactionCount"]) != 1 {
		t.Fatalf("react = %v", r)
	}
	r = h.mustRPC(u, "publication_news_react", map[string]any{"newsId": rich, "reaction": "love"})
	if r["myReaction"] != "love" || num(r["reactionCount"]) != 1 || num(r["reactions"].(map[string]any)["love"]) != 1 {
		t.Fatalf("change = %v", r)
	}
	r = h.mustRPC(u, "publication_news_react", map[string]any{"newsId": rich, "reaction": ""})
	if r["myReaction"] != nil || num(r["reactionCount"]) != 0 {
		t.Fatalf("remove = %v", r)
	}
	if code := h.rpcCode(u, "publication_news_react", map[string]any{"newsId": rich, "reaction": "like"}); code != "REACTION_RATE_LIMITED" {
		t.Fatalf("rate limit = %s", code)
	}
	h.s.db.ExecContext(h.ctx, `DELETE FROM cluster_rate_limit WHERE key LIKE 'news-%'`)
	if code := h.rpcCode(u, "publication_news_react", map[string]any{"newsId": rich, "reaction": "nope"}); code != "REACTION_NOT_FOUND" {
		t.Fatalf("unknown = %s", code)
	}
	h.mustAPI(http.StatusOK, http.MethodPut, "/news/"+rich, map[string]any{"title": map[string]string{"fa": "غنی"}, "bodyFormat": "html", "body": map[string]string{"fa": "<p>x</p>"}, "status": "published", "allowReaction": false})
	if code := h.rpcCode(u, "publication_news_react", map[string]any{"newsId": rich, "reaction": "like"}); code != "REACTION_NOT_ALLOWED" {
		t.Fatalf("closed = %s", code)
	}
	if got := h.mustRPC(u, "publication_news_seen", map[string]any{"all": true}); num(got["unread"]) != 0 {
		t.Fatalf("seen all = %v", got)
	}

	review := h.mustAPI(http.StatusOK, http.MethodPost, "/news-review", map[string]any{"ids": []string{pending, draft, rich}, "action": "approve"})
	if num(review["changed"]) != 2 || num(review["skipped"]) != 1 {
		t.Fatalf("review = %v", review)
	}
	if st, _ := h.status(pending); st != "published" {
		t.Fatalf("approved = %s", st)
	}
	if got := h.mustRPC(u, "publication_news_unread", map[string]any{}); num(got["unread"]) != 2 {
		t.Fatalf("unread after approve = %v", got)
	}
	h.mustAPI(http.StatusOK, http.MethodPost, "/news-review", map[string]any{"ids": []string{pending}, "action": "archive"})
	if code, out := h.api(http.MethodPost, "/news-review", map[string]any{"ids": []string{pending}, "action": "boom"}); code != 400 || out["error"] != "INVALID_ACTION" {
		t.Fatalf("bad action = %d %v", code, out)
	}
	used := h.mustAPI(http.StatusOK, http.MethodGet, "/news-reactions", nil)
	if num(used["total"]) != 5 {
		t.Fatalf("reaction types = %v", used)
	}
	h.mustAPI(http.StatusCreated, http.MethodPost, "/news-reactions", map[string]any{"key": "fire", "icon": "🔥", "name": map[string]string{"fa": "داغ"}, "sort": 9})
	if code, out := h.api(http.MethodPost, "/news-reactions", map[string]any{"key": "bad", "icon": "<b>", "name": map[string]string{"fa": "x"}}); code != 400 || out["error"] != "INVALID_ICON" {
		t.Fatalf("icon = %d %v", code, out)
	}
}

func TestNewsIngestTokenHMACDedupeAndNotifyOnce(t *testing.T) {
	h := newsHarness(t)
	h.user(nil, "", "fa")
	h.category("news", "sports")
	tg := h.newsTag("news", "nt_hot")
	src := h.mustAPI(http.StatusCreated, http.MethodPost, "/news-sources", map[string]any{"key": "wire", "name": "Wire", "publishMode": "review", "defaultLang": "fa",
		"mapping": map[string]string{"externalId": "guid", "cover": ""}, "itemsPath": "items", "notifyMode": "in_app"})
	srcID := src["id"].(string)
	if src["secretSet"] != true || src["endpoint"] != ingestPrefix+"wire" {
		t.Fatalf("source = %v", src)
	}
	secret := h.mustAPI(http.StatusOK, http.MethodPost, "/news-sources/"+srcID+"/secret/reveal", nil)["secret"].(string)
	body := `{"items":[{"guid":"a1","title":"Goal","body":"<p>Big <b>goal</b></p><script>x</script>","category":"sports","tags":["nt_hot","unknown"],"lang":"en"},{"guid":"a2","title":"Second","body":"text"}]}`
	if code, out := h.ingest("wire", map[string]string{"X-News-Token": "wrong"}, body); code != 401 || out["error"] != "INVALID_TOKEN" {
		t.Fatalf("bad token = %d %v", code, out)
	}
	code, out := h.ingest("wire", map[string]string{"X-News-Token": secret}, body)
	if code != 200 || num(out["created"]) != 2 {
		t.Fatalf("ingest = %d %v", code, out)
	}
	res := out["results"].([]any)[0].(map[string]any)
	newsID := res["newsId"].(string)
	if !strings.Contains(strings.Join(toStrings(res["notes"]), ","), "TAGS_UNKNOWN") {
		t.Fatalf("notes = %v", res)
	}
	var title, bodyJSON, lang, cat, status string
	var tags []byte
	if err := h.s.db.QueryRowContext(h.ctx, `SELECT n.title::text, n.body::text, n.lang, c.key, n.status, array_to_json(n.tag_ids)::text FROM publication_news n JOIN base_category c ON c.id=n.category_id WHERE n.id=$1`, newsID).
		Scan(&title, &bodyJSON, &lang, &cat, &status, &tags); err != nil {
		t.Fatal(err)
	}
	if lang != "en" || cat != "sports" || status != "pending_review" || !strings.Contains(string(tags), tg) || !strings.Contains(title, `"fa": "Goal"`) || strings.Contains(bodyJSON, "script") {
		t.Fatalf("stored = %s %s %s %s %s %s", title, bodyJSON, lang, cat, status, tags)
	}
	code, out = h.ingest("wire", map[string]string{"X-News-Token": secret}, body)
	if code != 200 || num(out["unchanged"]) != 2 {
		t.Fatalf("dedupe = %d %v", code, out)
	}
	changed := strings.Replace(body, "Big", "Huge", 1)
	if _, out = h.ingest("wire", map[string]string{"X-News-Token": secret}, changed); num(out["updated"]) != 1 || num(out["unchanged"]) != 1 {
		t.Fatalf("update = %v", out)
	}
	if n, _ := h.s.runNewsNotify(h.ctx); n != 0 || resetBroadcasts() != 0 {
		t.Fatal("pending news must not notify")
	}
	h.mustAPI(http.StatusOK, http.MethodPost, "/news-review", map[string]any{"ids": []string{newsID}, "action": "approve"})
	var wg sync.WaitGroup
	total := make(chan int, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := h.s.runNewsNotify(h.ctx)
			if err != nil {
				t.Error(err)
			}
			total <- n
		}()
	}
	wg.Wait()
	close(total)
	sum := 0
	for n := range total {
		sum += n
	}
	if sum != 1 || resetBroadcasts() != 1 {
		t.Fatalf("notify once: %d", sum)
	}
	var notes int
	h.s.db.QueryRowContext(h.ctx, `SELECT count(*) FROM publication_notification WHERE news_id=$1 AND kind='news' AND channel='in_app'`, newsID).Scan(&notes)
	if notes != 1 {
		t.Fatalf("notifications = %d", notes)
	}
	if n, _ := h.s.runNewsNotify(h.ctx); n != 0 {
		t.Fatal("second run notified again")
	}
	if _, notified := h.status(newsID); !notified {
		t.Fatal("notified_at not set")
	}

	h.mustAPI(http.StatusOK, http.MethodPut, "/news-sources/"+srcID, map[string]any{"key": "wire", "authMode": "hmac", "publishMode": "auto", "itemsPath": "items", "mapping": map[string]string{"externalId": "guid"}})
	rotated := h.mustAPI(http.StatusOK, http.MethodPost, "/news-sources/"+srcID+"/secret/rotate", nil)["secret"].(string)
	if rotated == secret {
		t.Fatal("rotate kept the secret")
	}
	one := `{"items":[{"guid":"h1","title":"Signed","body":"## Head\n\n**md**","publishedAt":"2020-01-01T00:00:00Z"}]}`
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := signIngest(rotated, ts, []byte(one))
	if code, out := h.ingest("wire", map[string]string{"X-News-Timestamp": ts, "X-News-Signature": "sha256=" + signIngest(secret, ts, []byte(one))}, one); code != 401 || out["error"] != "INVALID_SIGNATURE" {
		t.Fatalf("old secret = %d %v", code, out)
	}
	old := strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10)
	if code, out := h.ingest("wire", map[string]string{"X-News-Timestamp": old, "X-News-Signature": signIngest(rotated, old, []byte(one))}, one); code != 401 || out["error"] != "TIMESTAMP_OUT_OF_WINDOW" {
		t.Fatalf("old ts = %d %v", code, out)
	}
	code, out = h.ingest("wire", map[string]string{"X-News-Timestamp": ts, "X-News-Signature": "sha256=" + sig}, one)
	if code != 200 || num(out["created"]) != 1 {
		t.Fatalf("hmac = %d %v", code, out)
	}
	if code, out := h.ingest("wire", map[string]string{"X-News-Timestamp": ts, "X-News-Signature": sig}, one); code != 409 || out["error"] != "REPLAYED" {
		t.Fatalf("replay = %d %v", code, out)
	}
	auto := out["results"].([]any)[0].(map[string]any)["newsId"].(string)
	if st, _ := h.status(auto); st != "published" {
		t.Fatalf("auto publish = %s", st)
	}
	if n, _ := h.s.runNewsNotify(h.ctx); n != 1 {
		t.Fatal("auto-published news should notify once")
	}

	preview := h.mustAPI(http.StatusOK, http.MethodPost, "/news-sources/"+srcID+"/test", map[string]any{"payload": `{"items":[{"guid":"p1","title":"Preview"}]}`})
	sm := preview["summary"].(map[string]any)
	if preview["saved"] != false || num(sm["created"]) != 1 || sm["results"].([]any)[0].(map[string]any)["preview"] == nil {
		t.Fatalf("preview = %v", preview)
	}
	var exists bool
	h.s.db.QueryRowContext(h.ctx, `SELECT EXISTS (SELECT 1 FROM publication_news WHERE external_id='p1')`).Scan(&exists)
	if exists {
		t.Fatal("preview saved news")
	}
	bad := h.mustAPI(http.StatusOK, http.MethodPost, "/news-sources/"+srcID+"/test", map[string]any{"payload": `nope`})
	if bad["ok"] != false || bad["code"] != "INVALID_JSON" {
		t.Fatalf("bad test = %v", bad)
	}
	logs := h.mustAPI(http.StatusOK, http.MethodGet, "/news-ingest-log?sourceId="+srcID, nil)
	if num(logs["total"]) < 8 {
		t.Fatalf("log = %v", logs)
	}
	firstLog := items(logs)[0].(map[string]any)["id"].(string)
	detail := h.mustAPI(http.StatusOK, http.MethodGet, "/news-ingest-log/"+firstLog, nil)
	if detail["payload"] != "nope" || detail["isTest"] != true {
		t.Fatalf("log detail = %v", detail)
	}
	if code, _ := h.ingest("nosuch", nil, "{}"); code != 404 {
		t.Fatalf("unknown source = %d", code)
	}
	h.mustAPI(http.StatusOK, http.MethodPut, "/news-sources/"+srcID, map[string]any{"key": "wire", "rateLimitPerMinute": 1, "itemsPath": "items"})
	h.s.db.ExecContext(h.ctx, `DELETE FROM cluster_rate_limit WHERE key LIKE 'news-%'`)
	h.ingest("wire", nil, "{}")
	if code, out := h.ingest("wire", nil, "{}"); code != 429 || out["error"] != "RATE_LIMITED" {
		t.Fatalf("rate = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPut, "/news-sources/"+srcID, map[string]any{"key": "other"}); code != 400 || out["error"] != "KEY_LOCKED" {
		t.Fatalf("key lock = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/news-sources", map[string]any{"key": "bad path", "mapping": map[string]string{"title": "a..b"}}); code != 400 || out["error"] != "INVALID_KEY" {
		t.Fatalf("bad key = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/news-sources", map[string]any{"key": "pathy", "mapping": map[string]string{"title": "a..b"}}); code != 400 || out["field"] != "mapping.title" {
		t.Fatalf("bad path = %d %v", code, out)
	}
}

func TestNewsRemoteImageCopy(t *testing.T) {
	h := newsHarness(t)
	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\rIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82")
	img := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/a.png" {
			_, _ = w.Write(png)
			return
		}
		_, _ = w.Write([]byte("<html>no</html>"))
	}))
	defer img.Close()
	dir := t.TempDir()
	h.s.setupImageStore(map[string]string{"ECONOMY_PUBLIC_BASE_URL": "https://cdn.example", "ECONOMY_UPLOAD_DIR": dir})
	if _, err := h.s.copyRemoteImage(h.ctx, img.URL+"/a.png"); err == nil || !strings.Contains(err.Error(), "private") {
		t.Fatalf("private address must be refused: %v", err)
	}
	h.s.fetchPrivate = true
	got, err := h.s.copyRemoteImage(h.ctx, img.URL+"/a.png")
	if err != nil || !strings.HasPrefix(got, "https://cdn.example/v2/economy/files/") || !strings.HasSuffix(got, ".png") {
		t.Fatalf("copy = %q %v", got, err)
	}
	if _, err := h.s.copyRemoteImage(h.ctx, img.URL+"/page"); err == nil {
		t.Fatal("html must be refused")
	}
	src := h.mustAPI(http.StatusCreated, http.MethodPost, "/news-sources", map[string]any{"key": "imgs", "publishMode": "auto", "copyImages": true})
	secret := h.mustAPI(http.StatusOK, http.MethodPost, "/news-sources/"+src["id"].(string)+"/secret/reveal", nil)["secret"].(string)
	_, out := h.ingest("imgs", map[string]string{"X-News-Token": secret}, `{"id":"i1","title":"Pic","image":"`+img.URL+`/a.png"}`)
	id := out["results"].([]any)[0].(map[string]any)["newsId"].(string)
	var cover string
	h.s.db.QueryRowContext(h.ctx, `SELECT cover_url FROM publication_news WHERE id=$1`, id).Scan(&cover)
	if !strings.HasPrefix(cover, "https://cdn.example/v2/economy/files/") {
		t.Fatalf("cover = %s", cover)
	}
}

func toStrings(v any) []string {
	arr, _ := v.([]any)
	out := []string{}
	for _, x := range arr {
		s, _ := x.(string)
		out = append(out, s)
	}
	return out
}
