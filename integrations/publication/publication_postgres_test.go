package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
	"github.com/heroiclabs/nakama-common/runtime"
)

type fakeNakama struct {
	runtime.NakamaModule
	mu        sync.Mutex
	sent      []*runtime.NotificationSend
	failNext  int
	failAll   bool
	sendDelay time.Duration
}

func (n *fakeNakama) NotificationsSend(ctx context.Context, sends []*runtime.NotificationSend) error {
	if n.sendDelay > 0 {
		time.Sleep(n.sendDelay)
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.failAll {
		return errors.New("socket registry down")
	}
	if n.failNext > 0 {
		n.failNext--
		return errors.New("temporary failure")
	}
	n.sent = append(n.sent, sends...)
	return nil
}

func (n *fakeNakama) sentTo() map[string]int {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := map[string]int{}
	for _, s := range n.sent {
		out[s.UserID]++
	}
	return out
}

type fakeInit struct {
	runtime.Initializer
	console *mux.Router
}

func (f *fakeInit) RegisterConsoleHttp(path string, h func(http.ResponseWriter, *http.Request), methods ...string) error {
	f.console.HandleFunc(path, h).Methods(methods...)
	return nil
}

type harness struct {
	t    *testing.T
	s    *service
	nk   *fakeNakama
	init *fakeInit
	ctx  context.Context
}

func newHarness(t *testing.T) *harness {
	db := openDisposableTestDB(t)
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `TRUNCATE publication_notification_recipient, publication_notification, publication_provider, publication_news_comment, publication_news_seen,
		publication_news, publication_faq, base_category CASCADE`); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`DELETE FROM base_language WHERE code NOT IN ('fa','en')`,
		`UPDATE base_language SET is_default = false WHERE code = 'en'`,
		`UPDATE base_language SET is_default = (code = 'fa'), is_active = true`,
	} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM group_edge; DELETE FROM groups; DELETE FROM users WHERE id <> '00000000-0000-0000-0000-000000000000'`); err != nil {
		t.Fatal(err)
	}
	s := testService()
	nk := &fakeNakama{}
	s.db, s.nk = db, nk
	for k := range s.baseURLs {
		s.baseURLs[k] = "http://127.0.0.1:1"
	}
	init := &fakeInit{console: mux.NewRouter()}
	if err := s.registerConsoleHTTP(init); err != nil {
		t.Fatal(err)
	}
	return &harness{t: t, s: s, nk: nk, init: init, ctx: ctx}
}

func (h *harness) user(metadata map[string]any, email, lang string) string {
	h.t.Helper()
	id := uuid.Must(uuid.NewV4()).String()
	md, _ := json.Marshal(metadata)
	if metadata == nil {
		md = []byte("{}")
	}
	var e any
	if email != "" {
		e = email
	}
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO users (id, username, metadata, email, lang_tag) VALUES ($1,$2,$3,$4,$5)`,
		id, "u"+strings.ReplaceAll(id[:13], "-", ""), string(md), e, lang); err != nil {
		h.t.Fatal(err)
	}
	return id
}

func (h *harness) group(members map[string]int) string {
	h.t.Helper()
	id := uuid.Must(uuid.NewV4()).String()
	var creator string
	for u := range members {
		creator = u
		break
	}
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO groups (id, creator_id, name, edge_count, max_count) VALUES ($1,$2,$3,$4,100)`, id, creator, "g"+id[:8], len(members)); err != nil {
		h.t.Fatal(err)
	}
	pos := 0
	for u, state := range members {
		pos++
		if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO group_edge (source_id, position, destination_id, state) VALUES ($1,$2,$3,$4)`, id, pos, u, state); err != nil {
			h.t.Fatal(err)
		}
	}
	return id
}

func (h *harness) api(method, path string, body any) (int, map[string]any) {
	h.t.Helper()
	var raw []byte
	if body != nil {
		raw, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, consolePrefix+path, bytes.NewReader(raw))
	req.Header.Set("X-Nakama-Console-Username", "admin")
	rec := httptest.NewRecorder()
	h.init.console.ServeHTTP(rec, req)
	out := map[string]any{}
	if strings.HasPrefix(strings.TrimSpace(rec.Body.String()), "{") {
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
	} else {
		var list []any
		_ = json.Unmarshal(rec.Body.Bytes(), &list)
		out["list"] = list
	}
	return rec.Code, out
}

func (h *harness) mustAPI(want int, method, path string, body any) map[string]any {
	h.t.Helper()
	code, out := h.api(method, path, body)
	if code != want {
		h.t.Fatalf("%s %s = %d %v, want %d", method, path, code, out, want)
	}
	return out
}

func (h *harness) rpc(user, name string, payload any) (map[string]any, error) {
	h.t.Helper()
	raw, _ := json.Marshal(payload)
	out, err := h.s.clientRPCs()[name](h.ctx, user, string(raw))
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(out)
	m := map[string]any{}
	_ = json.Unmarshal(b, &m)
	return m, nil
}

func (h *harness) mustRPC(user, name string, payload any) map[string]any {
	h.t.Helper()
	out, err := h.rpc(user, name, payload)
	if err != nil {
		h.t.Fatalf("%s: %v", name, err)
	}
	return out
}

func (h *harness) rpcCode(user, name string, payload any) string {
	h.t.Helper()
	_, err := h.rpc(user, name, payload)
	if app, ok := asApp(err); ok {
		return app.Code
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func (h *harness) category(kind, key string) string {
	h.t.Helper()
	id := uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO base_category (id, key_type, key, name) VALUES ($1,$2,$3,'{"fa":"دسته","en":"Cat"}')`, id, kind, key); err != nil {
		h.t.Fatal(err)
	}
	return id
}

func (h *harness) news(body map[string]any) string {
	h.t.Helper()
	out := h.mustAPI(http.StatusCreated, http.MethodPost, "/news", body)
	return out["id"].(string)
}

func (h *harness) drain() {
	h.t.Helper()
	for i := 0; i < 50; i++ {
		n, err := h.s.runWorker(h.ctx)
		if err != nil {
			h.t.Fatal(err)
		}
		if n == 0 {
			return
		}
	}
	h.t.Fatal("worker did not drain")
}

func (h *harness) dueNow() {
	h.t.Helper()
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_notification_recipient SET next_attempt_at=now() WHERE status='pending'`); err != nil {
		h.t.Fatal(err)
	}
}

func items(m map[string]any) []any {
	v, _ := m["items"].([]any)
	return v
}

func num(v any) int {
	f, _ := v.(float64)
	return int(f)
}

func TestCommentsModerationAndReplies(t *testing.T) {
	h := newHarness(t)
	a, b := h.user(nil, "", "fa"), h.user(nil, "", "en")
	cat := h.category("news", "updates")
	newsID := h.news(map[string]any{"title": map[string]string{"fa": "خبر"}, "body": map[string]string{"fa": "متن"}, "categoryId": cat})
	if code, _ := h.api(http.MethodPost, "/news", map[string]any{"title": map[string]string{"fa": "x"}, "body": map[string]string{"fa": "y"}, "categoryId": h.category("faq", "f")}); code != http.StatusBadRequest {
		t.Fatalf("faq category on news = %d", code)
	}

	root := h.mustRPC(a, "publication_news_comment", map[string]any{"newsId": newsID, "message": "  first  "})
	if root["isAccepted"] != false || root["message"] != "first" || root["mine"] != true {
		t.Fatalf("root = %v", root)
	}
	rootID := root["id"].(string)
	if got := items(h.mustRPC(b, "publication_news_comments", map[string]any{"newsId": newsID})); len(got) != 0 {
		t.Fatalf("pending comment leaked to another user: %v", got)
	}
	if got := items(h.mustRPC(a, "publication_news_comments", map[string]any{"newsId": newsID})); len(got) != 1 {
		t.Fatalf("author must see own pending comment: %v", got)
	}
	if code := h.rpcCode(b, "publication_news_comment", map[string]any{"newsId": newsID, "message": "x", "parentId": rootID}); code != "PARENT_COMMENT_NOT_FOUND" {
		t.Fatalf("reply to hidden comment = %s", code)
	}
	pending := h.mustAPI(http.StatusOK, http.MethodGet, "/comments?status=pending&newsId="+newsID, nil)
	if num(pending["total"]) != 1 {
		t.Fatalf("pending = %v", pending)
	}
	accepted := h.mustAPI(http.StatusOK, http.MethodPut, "/comments/"+rootID, map[string]any{"isAccepted": true})
	if accepted["isAccepted"] != true || accepted["newsTitle"].(map[string]any)["fa"] != "خبر" {
		t.Fatalf("accept = %v", accepted)
	}
	reply := h.mustRPC(b, "publication_news_comment", map[string]any{"newsId": newsID, "message": "reply", "parentId": rootID})
	h.mustAPI(http.StatusOK, http.MethodPut, "/comments/"+reply["id"].(string), map[string]any{"isAccepted": true})
	nested := h.mustRPC(a, "publication_news_comment", map[string]any{"newsId": newsID, "message": "nested", "parentId": reply["id"]})
	if nested["parentId"] != rootID {
		t.Fatalf("reply to a reply must attach to the root: %v", nested)
	}
	list := items(h.mustRPC(b, "publication_news_comments", map[string]any{"newsId": newsID}))
	if len(list) != 1 {
		t.Fatalf("roots = %v", list)
	}
	replies := list[0].(map[string]any)["replies"].([]any)
	if len(replies) != 1 || replies[0].(map[string]any)["message"] != "reply" {
		t.Fatalf("b must see only the accepted reply: %v", replies)
	}
	replies = items(h.mustRPC(a, "publication_news_comments", map[string]any{"newsId": newsID}))[0].(map[string]any)["replies"].([]any)
	if len(replies) != 2 {
		t.Fatalf("a must see accepted + own pending reply: %v", replies)
	}
	got := h.mustRPC(a, "publication_news_get", map[string]any{"id": newsID})
	if num(got["commentCount"]) != 2 {
		t.Fatalf("commentCount = %v", got["commentCount"])
	}

	if code := h.rpcCode(a, "publication_news_comment", map[string]any{"newsId": newsID, "message": strings.Repeat("ب", 1001)}); code != "COMMENT_TOO_LONG" {
		t.Fatalf("long comment = %s", code)
	}
	for i := 0; i < 3; i++ {
		h.mustRPC(a, "publication_news_comment", map[string]any{"newsId": newsID, "message": "spam"})
	}
	if code := h.rpcCode(a, "publication_news_comment", map[string]any{"newsId": newsID, "message": "spam"}); code != "COMMENT_RATE_LIMITED" {
		t.Fatalf("6th comment in a minute = %s", code)
	}

	h.mustAPI(http.StatusOK, http.MethodDelete, "/comments/"+rootID, nil)
	var left int
	_ = h.s.db.QueryRow(`SELECT count(*) FROM publication_news_comment WHERE id=$1 OR parent_id=$1`, rootID).Scan(&left)
	if left != 0 {
		t.Fatalf("deleting a root must delete its replies, %d left", left)
	}
	h.mustAPI(http.StatusNotFound, http.MethodDelete, "/comments/"+rootID, nil)

	free := h.news(map[string]any{"title": map[string]string{"fa": "آزاد"}, "body": map[string]string{"fa": "x"}, "freeComment": true})
	if c := h.mustRPC(b, "publication_news_comment", map[string]any{"newsId": free, "message": "hi"}); c["isAccepted"] != true {
		t.Fatalf("free comment must be accepted: %v", c)
	}
	closed := h.news(map[string]any{"title": map[string]string{"fa": "بسته"}, "body": map[string]string{"fa": "x"}, "allowComment": false})
	if code := h.rpcCode(b, "publication_news_comment", map[string]any{"newsId": closed, "message": "hi"}); code != "COMMENT_NOT_ALLOWED" {
		t.Fatalf("closed = %s", code)
	}
	if _, err := h.s.db.Exec(`DELETE FROM base_category WHERE id=$1`, cat); err == nil {
		t.Fatal("a used news category must not be deletable")
	}
}

func TestNewsListSeenAndFaq(t *testing.T) {
	h := newHarness(t)
	u := h.user(nil, "", "en")
	cat := h.category("news", "events")
	old := h.news(map[string]any{"title": map[string]string{"fa": "قدیمی", "en": "Old"}, "body": map[string]string{"fa": "**متن**"}, "publishedAt": time.Now().Add(-48 * time.Hour).Format(time.RFC3339)})
	pinned := h.news(map[string]any{"title": map[string]string{"fa": "سنجاق"}, "body": map[string]string{"fa": "x"}, "isPinned": true, "publishedAt": time.Now().Add(-72 * time.Hour).Format(time.RFC3339), "categoryId": cat,
		"actions": []map[string]any{{"type": "url", "label": map[string]string{"fa": "باز کن", "en": "Open"}, "value": "https://example.com"}}})
	fresh := h.news(map[string]any{"title": map[string]string{"fa": "تازه"}, "body": map[string]string{"fa": "x"}})
	h.news(map[string]any{"title": map[string]string{"fa": "آینده"}, "body": map[string]string{"fa": "x"}, "publishedAt": time.Now().Add(time.Hour).Format(time.RFC3339)})
	h.news(map[string]any{"title": map[string]string{"fa": "خاموش"}, "body": map[string]string{"fa": "x"}, "isActive": false})

	list := items(h.mustRPC(u, "publication_news_list", map[string]any{"lang": "en"}))
	if len(list) != 3 || list[0].(map[string]any)["id"] != pinned || list[1].(map[string]any)["id"] != fresh || list[2].(map[string]any)["id"] != old {
		t.Fatalf("order = %v", list)
	}
	if list[2].(map[string]any)["title"] != "Old" || list[1].(map[string]any)["title"] != "تازه" {
		t.Fatalf("lang fallback: %v", list)
	}
	page1 := h.mustRPC(u, "publication_news_list", map[string]any{"limit": 2})
	if len(items(page1)) != 2 || page1["cursor"] != "2" {
		t.Fatalf("page1 = %v", page1)
	}
	page2 := h.mustRPC(u, "publication_news_list", map[string]any{"limit": 2, "cursor": page1["cursor"]})
	if len(items(page2)) != 1 || page2["cursor"] != "" {
		t.Fatalf("page2 = %v", page2)
	}
	if got := items(h.mustRPC(u, "publication_news_list", map[string]any{"categoryKey": "events"})); len(got) != 1 {
		t.Fatalf("category filter = %v", got)
	}
	for i := 0; i < 2; i++ {
		h.mustRPC(u, "publication_news_seen", map[string]any{"id": old})
	}
	var seenRows int
	_ = h.s.db.QueryRow(`SELECT count(*) FROM publication_news_seen WHERE news_id=$1`, old).Scan(&seenRows)
	if seenRows != 1 {
		t.Fatalf("seen rows = %d", seenRows)
	}
	got := h.mustRPC(u, "publication_news_get", map[string]any{"id": old, "lang": "en"})
	if got["seen"] != true || got["bodyHtml"] != "<p><strong>متن</strong></p>" {
		t.Fatalf("get = %v", got)
	}
	got = h.mustRPC(u, "publication_news_get", map[string]any{"id": pinned, "lang": "en"})
	acts := got["actions"].([]any)
	if len(acts) != 1 || acts[0].(map[string]any)["label"] != "Open" {
		t.Fatalf("actions = %v", got["actions"])
	}
	future := h.mustAPI(http.StatusOK, http.MethodGet, "/news?q=آینده", nil)
	if num(future["total"]) != 1 {
		t.Fatalf("console search = %v", future)
	}
	if code := h.rpcCode(u, "publication_news_get", map[string]any{"id": items(future)[0].(map[string]any)["id"]}); code != "NEWS_NOT_FOUND" {
		t.Fatalf("future news = %s", code)
	}
	if code := h.rpcCode(u, "publication_news_seen", map[string]any{"id": "nope"}); code != "NEWS_NOT_FOUND" {
		t.Fatalf("bad id = %s", code)
	}

	faqCat := h.category("faq", "account")
	h.mustAPI(http.StatusCreated, http.MethodPost, "/faqs", map[string]any{"question": map[string]string{"fa": "رمز؟", "en": "Password?"}, "answer": map[string]string{"fa": "[اینجا](https://e.com)"}, "categoryId": faqCat, "sort": 2})
	h.mustAPI(http.StatusCreated, http.MethodPost, "/faqs", map[string]any{"question": map[string]string{"fa": "خرید؟"}, "answer": map[string]string{"fa": "ب"}, "sort": 1})
	h.mustAPI(http.StatusCreated, http.MethodPost, "/faqs", map[string]any{"question": map[string]string{"fa": "پنهان"}, "answer": map[string]string{"fa": "x"}, "isActive": false})
	if code, out := h.api(http.MethodPost, "/faqs", map[string]any{"question": map[string]string{"en": "x"}, "answer": map[string]string{"fa": "x"}}); code != http.StatusBadRequest || out["error"] != "DEFAULT_LANGUAGE_REQUIRED" {
		t.Fatalf("fa required = %d %v", code, out)
	}
	faq := h.mustRPC(u, "publication_faq_list", map[string]any{"lang": "en"})
	fi := items(faq)
	if len(fi) != 2 || fi[0].(map[string]any)["question"] != "خرید؟" || fi[1].(map[string]any)["question"] != "Password?" {
		t.Fatalf("faq = %v", fi)
	}
	if !strings.Contains(fi[1].(map[string]any)["answerHtml"].(string), `<a href="https://e.com"`) || fi[1].(map[string]any)["categoryKey"] != "account" {
		t.Fatalf("faq item = %v", fi[1])
	}
	if cats := faq["categories"].([]any); len(cats) != 1 || cats[0].(map[string]any)["name"] != "Cat" {
		t.Fatalf("faq categories = %v", cats)
	}
	if got := items(h.mustRPC(u, "publication_faq_list", map[string]any{"categoryKey": "account"})); len(got) != 1 {
		t.Fatalf("faq category filter = %v", got)
	}
	if got := items(h.mustRPC(u, "publication_faq_list", map[string]any{"q": "خرید"})); len(got) != 1 {
		t.Fatalf("faq search = %v", got)
	}
	listed := h.mustAPI(http.StatusOK, http.MethodGet, "/faqs?isActive=false", nil)
	if num(listed["total"]) != 1 {
		t.Fatalf("faq isActive filter = %v", listed)
	}
}

func TestNotificationTargetsAndAddresses(t *testing.T) {
	h := newHarness(t)
	withPhone := h.user(map[string]any{"phone": "09121234567"}, "", "fa")
	withIntl := h.user(map[string]any{"phone": "+989351112233", "email": "meta@example.com"}, "", "en")
	noPhone := h.user(nil, "col@example.com", "fa")
	disabled := h.user(map[string]any{"phone": "09120000000"}, "", "fa")
	if _, err := h.s.db.Exec(`UPDATE users SET disable_time=now() WHERE id=$1`, disabled); err != nil {
		t.Fatal(err)
	}
	joinRequest := h.user(map[string]any{"phone": "09129999999"}, "", "fa")
	group := h.group(map[string]int{withPhone: 0, withIntl: 2, joinRequest: 3})

	preview := h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/preview-target?mode=all", nil)
	if num(preview["count"]) != 4 || num(preview["withPhone"]) != 3 || num(preview["withEmail"]) != 2 {
		t.Fatalf("preview all = %v", preview)
	}
	preview = h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/preview-target?mode=group&groupId="+group, nil)
	if num(preview["count"]) != 2 {
		t.Fatalf("preview group must skip join requests = %v", preview)
	}
	preview = h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/preview-target?mode=users&userIds="+withPhone+","+noPhone, nil)
	if num(preview["count"]) != 2 || num(preview["withPhone"]) != 1 {
		t.Fatalf("preview users = %v", preview)
	}
	if code, out := h.api(http.MethodGet, "/notifications/preview-target?mode=users&userIds="+uuid.Must(uuid.NewV4()).String(), nil); code != http.StatusBadRequest || out["error"] != "USERS_NOT_FOUND" {
		t.Fatalf("missing user = %d %v", code, out)
	}
	if code, out := h.api(http.MethodGet, "/notifications/preview-target?mode=group&groupId="+uuid.Must(uuid.NewV4()).String(), nil); code != http.StatusBadRequest || out["error"] != "GROUP_NOT_FOUND" {
		t.Fatalf("missing group = %d %v", code, out)
	}

	if code, out := h.api(http.MethodPost, "/notifications", map[string]any{"channel": "sms", "body": map[string]string{"fa": "x"}, "target": map[string]any{"mode": "all"}}); code != http.StatusBadRequest || out["error"] != "PROVIDER_NOT_CONFIGURED" {
		t.Fatalf("sms without provider = %d %v", code, out)
	}
	sms := h.mustAPI(http.StatusCreated, http.MethodPost, "/providers", map[string]any{"channel": "sms", "kind": "smsir", "isActive": true, "isDefault": true, "config": map[string]any{"apiKey": "k", "lineNumber": "3000"}})
	if sms["sandbox"] != true {
		t.Fatalf("new providers must start in sandbox: %v", sms)
	}
	n := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "sms", "body": map[string]string{"fa": "کد ۱۲۳"}, "target": map[string]any{"mode": "all"}})
	if num(n["recipientCount"]) != 4 || num(n["skippedCount"]) != 1 || n["status"] != "queued" || n["providerId"] != sms["id"] || n["createdBy"] != "admin" {
		t.Fatalf("sms all = %v", n)
	}
	rows := map[string]string{}
	r, _ := h.s.db.Query(`SELECT user_id::text, COALESCE(address,'')||'|'||status FROM publication_notification_recipient WHERE notification_id=$1`, n["id"])
	for r.Next() {
		var id, v string
		_ = r.Scan(&id, &v)
		rows[id] = v
	}
	r.Close()
	if rows[withPhone] != "+989121234567|pending" || rows[withIntl] != "+989351112233|pending" || rows[noPhone] != "|skipped" || rows[disabled] != "" {
		t.Fatalf("recipients = %v", rows)
	}

	h.mustAPI(http.StatusCreated, http.MethodPost, "/providers", map[string]any{"channel": "email", "kind": "smtp", "isActive": true, "config": map[string]any{"host": "smtp.example.com", "from": "a@example.com"}})
	e := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "email", "title": map[string]string{"fa": "سلام"}, "body": map[string]string{"fa": "x"},
		"target": map[string]any{"mode": "users", "userIds": []string{withIntl, noPhone, withPhone}}})
	if num(e["recipientCount"]) != 3 || num(e["skippedCount"]) != 1 {
		t.Fatalf("email users = %v", e)
	}
	var addr string
	_ = h.s.db.QueryRow(`SELECT address FROM publication_notification_recipient WHERE notification_id=$1 AND user_id=$2`, e["id"], noPhone).Scan(&addr)
	if addr != "col@example.com" {
		t.Fatalf("account email must win: %q", addr)
	}
	_ = h.s.db.QueryRow(`SELECT address FROM publication_notification_recipient WHERE notification_id=$1 AND user_id=$2`, e["id"], withIntl).Scan(&addr)
	if addr != "meta@example.com" {
		t.Fatalf("metadata email fallback: %q", addr)
	}
	g := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "in_app", "title": map[string]string{"fa": "گروه"}, "body": map[string]string{"fa": "x"},
		"target": map[string]any{"mode": "group", "groupId": group}})
	if num(g["recipientCount"]) != 2 || num(g["skippedCount"]) != 0 {
		t.Fatalf("group = %v", g)
	}
	empty := h.group(map[string]int{joinRequest: 3})
	g2 := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "in_app", "title": map[string]string{"fa": "خالی"}, "body": map[string]string{"fa": "x"},
		"target": map[string]any{"mode": "group", "groupId": empty}})
	if num(g2["recipientCount"]) != 0 || g2["status"] != "done" {
		t.Fatalf("empty group must be done at once: %v", g2)
	}

	h.drain()
	detail := h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/"+n["id"].(string)+"?rStatus=sent", nil)
	if detail["status"] != "done" || num(detail["sentCount"]) != 3 || num(detail["skippedCount"]) != 1 || num(detail["recipients"].(map[string]any)["total"]) != 3 {
		t.Fatalf("sms detail = %v", detail)
	}
	for _, it := range detail["recipients"].(map[string]any)["items"].([]any) {
		if it.(map[string]any)["error"] != "sandbox" {
			t.Fatalf("sandbox recipient = %v", it)
		}
	}
	list := h.mustAPI(http.StatusOK, http.MethodGet, "/notifications?channel=sms", nil)
	if num(list["total"]) != 1 {
		t.Fatalf("list filter = %v", list)
	}
	summary := h.mustAPI(http.StatusOK, http.MethodGet, "/summary", nil)
	if num(summary["counts"].(map[string]any)["notifications"]) != 4 {
		t.Fatalf("summary = %v", summary)
	}
	found := h.mustAPI(http.StatusOK, http.MethodGet, "/users/search?q="+withIntl, nil)["list"].([]any)
	if len(found) != 1 || found[0].(map[string]any)["id"] != withIntl {
		t.Fatalf("users search by id = %v", found)
	}
}

func TestInAppWorkerInboxAndRetries(t *testing.T) {
	h := newHarness(t)
	fa, en := h.user(nil, "", "fa"), h.user(nil, "", "en")
	n := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "in_app", "title": map[string]string{"fa": "سلام", "en": "Hello"},
		"body": map[string]string{"fa": "متن", "en": "Text"}, "action": map[string]any{"type": "goto", "value": "shop"}, "target": map[string]any{"mode": "all"}})
	id := n["id"].(string)

	h.nk.failNext = 1
	h.drain()
	var attempts, pending int
	_ = h.s.db.QueryRow(`SELECT max(attempts), count(*) FILTER (WHERE status='pending') FROM publication_notification_recipient WHERE notification_id=$1`, id).Scan(&attempts, &pending)
	if attempts != 1 || pending != 2 {
		t.Fatalf("after a failed attempt: attempts=%d pending=%d", attempts, pending)
	}
	if got := h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/"+id, nil); got["status"] != "sending" {
		t.Fatalf("status = %v", got["status"])
	}
	h.dueNow()
	h.drain()
	got := h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/"+id, nil)
	if got["status"] != "done" || num(got["sentCount"]) != 2 {
		t.Fatalf("after retry = %v", got)
	}
	subjects := map[string]string{}
	for _, s := range h.nk.sent {
		subjects[s.UserID] = s.Subject
		if s.Code != inAppCode || !s.Persistent || s.Content["notificationId"] != id || s.Content["action"].(map[string]any)["value"] != "shop" {
			t.Fatalf("send = %+v", s)
		}
	}
	if subjects[fa] != "سلام" || subjects[en] != "Hello" {
		t.Fatalf("subjects = %v", subjects)
	}

	inbox := h.mustRPC(en, "publication_inbox", map[string]any{"lang": "en"})
	if num(inbox["unread"]) != 1 || len(items(inbox)) != 1 || items(inbox)[0].(map[string]any)["title"] != "Hello" || items(inbox)[0].(map[string]any)["seen"] != false {
		t.Fatalf("inbox = %v", inbox)
	}
	h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "in_app", "title": map[string]string{"fa": "دوم"}, "body": map[string]string{"fa": "x"}, "target": map[string]any{"mode": "users", "userIds": []string{en}}})
	h.drain()
	if out := h.mustRPC(en, "publication_inbox_seen", map[string]any{"ids": []string{id}}); num(out["unread"]) != 1 {
		t.Fatalf("seen ids = %v", out)
	}
	if out := h.mustRPC(en, "publication_inbox_seen", map[string]any{"ids": []string{id}}); num(out["unread"]) != 1 {
		t.Fatalf("seen is idempotent = %v", out)
	}
	if out := h.mustRPC(en, "publication_inbox_seen", map[string]any{"all": true}); num(out["unread"]) != 0 {
		t.Fatalf("seen all = %v", out)
	}
	if out := h.mustRPC(fa, "publication_inbox", map[string]any{}); num(out["unread"]) != 1 {
		t.Fatalf("other users unaffected = %v", out)
	}
	if code := h.rpcCode(en, "publication_inbox_seen", map[string]any{}); code != "IDS_REQUIRED" {
		t.Fatalf("empty seen = %s", code)
	}

	h.nk.failAll = true
	dead := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "in_app", "title": map[string]string{"fa": "x"}, "body": map[string]string{"fa": "x"}, "target": map[string]any{"mode": "users", "userIds": []string{fa}}})
	for i := 0; i < 4; i++ {
		h.drain()
		h.dueNow()
	}
	got = h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/"+dead["id"].(string), nil)
	r := got["recipients"].(map[string]any)["items"].([]any)[0].(map[string]any)
	if got["status"] != "failed" || num(got["failedCount"]) != 1 || num(r["attempts"]) != 3 || !strings.Contains(r["error"].(string), "socket registry down") {
		t.Fatalf("exhausted = %v", got)
	}
	h.nk.failAll = false
}

func TestProviderFailuresUseRetries(t *testing.T) {
	h := newHarness(t)
	u := h.user(map[string]any{"phone": "09121234567"}, "", "fa")
	p := h.mustAPI(http.StatusCreated, http.MethodPost, "/providers", map[string]any{"channel": "sms", "kind": "kavenegar", "isActive": true, "config": map[string]any{"apiKey": "k"}})
	n := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "sms", "body": map[string]string{"fa": "x"}, "target": map[string]any{"mode": "users", "userIds": []string{u}}})
	h.mustAPI(http.StatusOK, http.MethodPut, "/providers/"+p["id"].(string), map[string]any{"isActive": false})
	for i := 0; i < 4; i++ {
		h.drain()
		h.dueNow()
	}
	got := h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/"+n["id"].(string), nil)
	r := got["recipients"].(map[string]any)["items"].([]any)[0].(map[string]any)
	if got["status"] != "failed" || num(r["attempts"]) != 3 || !strings.Contains(r["error"].(string), "not active") {
		t.Fatalf("inactive provider = %v", got)
	}
	h.mustAPI(http.StatusOK, http.MethodPut, "/providers/"+p["id"].(string), map[string]any{"isActive": true})
	n2 := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "sms", "body": map[string]string{"fa": "x"}, "target": map[string]any{"mode": "users", "userIds": []string{u}}})
	if code, out := h.api(http.MethodDelete, "/providers/"+p["id"].(string), nil); code != http.StatusConflict || out["error"] != "IN_USE" {
		t.Fatalf("delete busy provider = %d %v", code, out)
	}
	if _, err := h.s.db.Exec(`DELETE FROM publication_provider WHERE id=$1`, p["id"]); err != nil {
		t.Fatal(err)
	}
	h.drain()
	got = h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/"+n2["id"].(string), nil)
	r = got["recipients"].(map[string]any)["items"].([]any)[0].(map[string]any)
	if got["status"] != "failed" || num(r["attempts"]) != 1 {
		t.Fatalf("deleted provider must fail at once = %v", got)
	}
}

func TestConcurrentWorkersDoNotDoubleSend(t *testing.T) {
	h := newHarness(t)
	users := map[string]bool{}
	for i := 0; i < 130; i++ {
		users[h.user(nil, "", "fa")] = true
	}
	h.nk.sendDelay = 5 * time.Millisecond
	n := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "in_app", "title": map[string]string{"fa": "همه"}, "body": map[string]string{"fa": "x"}, "target": map[string]any{"mode": "all"}})
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for w := 0; w < 6; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				if _, err := h.s.runWorker(h.ctx); err != nil {
					errs <- err
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	sent := h.nk.sentTo()
	if len(sent) != 130 {
		t.Fatalf("sent to %d users, want 130", len(sent))
	}
	for u, c := range sent {
		if c != 1 || !users[u] {
			t.Fatalf("user %s got %d sends", u, c)
		}
	}
	got := h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/"+n["id"].(string), nil)
	if got["status"] != "done" || num(got["sentCount"]) != 130 || num(got["recipientCount"]) != 130 {
		t.Fatalf("counts = %v", got)
	}
}

func TestProvidersConsole(t *testing.T) {
	h := newHarness(t)
	k := h.mustAPI(http.StatusCreated, http.MethodPost, "/providers", map[string]any{"channel": "sms", "kind": "kavenegar", "config": map[string]any{"apiKey": "secret-kave", "sender": "1000"}})
	if k["config"].(map[string]any)["apiKey"] != secretMask || k["config"].(map[string]any)["sender"] != "1000" || k["sandbox"] != true || k["isActive"] != false {
		t.Fatalf("create = %v", k)
	}
	id := k["id"].(string)
	h.mustAPI(http.StatusConflict, http.MethodPost, "/providers", map[string]any{"channel": "sms", "kind": "kavenegar"})
	h.mustAPI(http.StatusBadRequest, http.MethodPost, "/providers", map[string]any{"channel": "sms", "kind": "smtp"})
	upd := h.mustAPI(http.StatusOK, http.MethodPut, "/providers/"+id, map[string]any{"isActive": true, "isDefault": true, "config": map[string]any{"apiKey": secretMask, "sender": "2000"}})
	if upd["isDefault"] != true || upd["config"].(map[string]any)["sender"] != "2000" {
		t.Fatalf("update = %v", upd)
	}
	p, err := h.s.loadProvider(h.ctx, h.s.db, id, false)
	if err != nil || p.Config["apiKey"] != "secret-kave" {
		t.Fatalf("masked round trip lost the secret: %v %v", p.Config, err)
	}
	var sealed string
	_ = h.s.db.QueryRow(`SELECT config FROM publication_provider WHERE id=$1`, id).Scan(&sealed)
	if strings.Contains(sealed, "secret-kave") {
		t.Fatal("config must be encrypted at rest")
	}
	tw := h.mustAPI(http.StatusCreated, http.MethodPost, "/providers", map[string]any{"channel": "sms", "kind": "twilio", "isActive": true, "isDefault": true,
		"config": map[string]any{"accountSid": "AC1", "authToken": "t", "from": "+14155550100"}})
	if tw["isDefault"] != true {
		t.Fatalf("twilio = %v", tw)
	}
	list := h.mustAPI(http.StatusOK, http.MethodGet, "/providers", nil)
	defaults := 0
	for _, it := range items(list) {
		if it.(map[string]any)["isDefault"] == true {
			defaults++
		}
	}
	if defaults != 1 || len(list["catalog"].([]any)) != 5 {
		t.Fatalf("list = %v", list)
	}
	if code, out := h.api(http.MethodPut, "/providers/"+id, map[string]any{"kind": "smsir"}); code != http.StatusBadRequest || out["error"] != "PROVIDER_KIND_FIXED" {
		t.Fatalf("kind change = %d %v", code, out)
	}
	test := h.mustAPI(http.StatusOK, http.MethodPost, "/providers/"+id+"/test", map[string]any{"to": "09121234567", "message": "test"})
	if test["ok"] != true || test["sandbox"] != true {
		t.Fatalf("sandbox test = %v", test)
	}
	h.mustAPI(http.StatusBadRequest, http.MethodPost, "/providers/"+id+"/test", map[string]any{"to": "12", "message": "test"})
	smtp := h.mustAPI(http.StatusCreated, http.MethodPost, "/providers", map[string]any{"channel": "email", "kind": "smtp", "config": map[string]any{"host": "smtp.example.com", "port": 465, "security": "tls", "from": "a@example.com", "password": "pw"}})
	if smtp["config"].(map[string]any)["port"] != "465" || smtp["config"].(map[string]any)["password"] != secretMask {
		t.Fatalf("smtp = %v", smtp)
	}
	h.mustAPI(http.StatusBadRequest, http.MethodPost, "/providers/"+smtp["id"].(string)+"/test", map[string]any{"to": "not-mail", "message": "x"})
	if out := h.mustAPI(http.StatusOK, http.MethodPost, "/providers/"+smtp["id"].(string)+"/test", map[string]any{"to": "u@example.com", "message": "x"}); out["sandbox"] != true {
		t.Fatalf("smtp sandbox test = %v", out)
	}
	h.mustAPI(http.StatusOK, http.MethodDelete, "/providers/"+smtp["id"].(string), nil)
	h.mustAPI(http.StatusNotFound, http.MethodDelete, "/providers/"+smtp["id"].(string), nil)
}

func TestMultilingualDeliveryAndContent(t *testing.T) {
	h := newHarness(t)
	if _, err := h.s.db.Exec(`INSERT INTO base_language (code, name, native_name, direction, sort) VALUES ('ar','{"fa":"عربی"}','العربية','rtl',3)`); err != nil {
		t.Fatal(err)
	}
	fa := h.user(map[string]any{"phone": "09121111111"}, "fa@example.com", "fa")
	en := h.user(map[string]any{"phone": "09122222222"}, "en@example.com", "en")
	ar := h.user(map[string]any{"phone": "09123333333"}, "ar@example.com", "ar")
	odd := h.user(map[string]any{"phone": "09124444444"}, "zz@example.com", "zz")
	expect := map[string]string{fa: "fa", en: "en", ar: "ar", odd: "fa"}

	title := map[string]string{"fa": "سلام", "en": "Hello", "ar": "مرحبا"}
	body := map[string]string{"fa": "متن", "en": "Text", "ar": "نص"}
	n := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "in_app", "title": title, "body": body,
		"action": map[string]any{"type": "goto", "value": "shop", "label": map[string]string{"fa": "فروشگاه", "ar": "متجر"}}, "target": map[string]any{"mode": "all"}})
	if n["title"].(map[string]any)["ar"] != "مرحبا" || n["action"].(map[string]any)["label"].(map[string]any)["ar"] != "متجر" {
		t.Fatalf("console keeps full maps: %v", n)
	}
	if code, out := h.api(http.MethodPost, "/notifications", map[string]any{"channel": "in_app", "title": map[string]string{"fa": "x", "tr": "y"}, "body": body, "target": map[string]any{"mode": "all"}}); code != http.StatusBadRequest || out["error"] != "INVALID_LANGUAGE" {
		t.Fatalf("unknown language = %d %v", code, out)
	}
	h.drain()
	sent := map[string]map[string]any{}
	for _, s := range h.nk.sent {
		sent[s.UserID] = s.Content
		if s.Subject != title[expect[s.UserID]] {
			t.Errorf("subject for %s = %q", expect[s.UserID], s.Subject)
		}
	}
	for user, lang := range expect {
		c := sent[user]
		if c["lang"] != lang || c["title"] != title[lang] || c["body"] != body[lang] {
			t.Errorf("in_app content for %s = %v", lang, c)
		}
		inbox := items(h.mustRPC(user, "publication_inbox", map[string]any{}))
		if len(inbox) != 1 || inbox[0].(map[string]any)["title"] != title[lang] {
			t.Errorf("inbox for %s = %v", lang, inbox)
		}
	}
	if label := sent[ar]["action"].(map[string]any)["label"]; label != "متجر" {
		t.Errorf("ar action label = %v", label)
	}
	if label := sent[en]["action"].(map[string]any)["label"]; label != "فروشگاه" {
		t.Errorf("en action label falls back to default = %v", label)
	}
	if inbox := items(h.mustRPC(fa, "publication_inbox", map[string]any{"lang": "ar"})); inbox[0].(map[string]any)["title"] != "مرحبا" || inbox[0].(map[string]any)["action"].(map[string]any)["label"] != "متجر" {
		t.Errorf("requested lang wins = %v", inbox)
	}
	var langs map[string]string = map[string]string{}
	rows, _ := h.s.db.Query(`SELECT user_id::text, COALESCE(lang,'') FROM publication_notification_recipient WHERE notification_id=$1`, n["id"])
	for rows.Next() {
		var id, l string
		_ = rows.Scan(&id, &l)
		langs[id] = l
	}
	rows.Close()
	for user, lang := range expect {
		if langs[user] != lang {
			t.Errorf("recipient lang for %s = %q", lang, langs[user])
		}
	}

	h.mustAPI(http.StatusCreated, http.MethodPost, "/providers", map[string]any{"channel": "sms", "kind": "smsir", "isActive": true, "isDefault": true, "config": map[string]any{"apiKey": "k", "lineNumber": "3000"}})
	h.mustAPI(http.StatusCreated, http.MethodPost, "/providers", map[string]any{"channel": "email", "kind": "smtp", "isActive": true, "isDefault": true, "config": map[string]any{"host": "smtp.example.com", "from": "a@example.com"}})
	sms := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "sms", "body": map[string]string{"fa": "کد", "en": "Code", "ar": "رمز"}, "target": map[string]any{"mode": "all"}})
	email := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "email", "title": title, "body": map[string]string{"fa": "**متن**", "ar": "**نص**"}, "target": map[string]any{"mode": "all"}})
	h.drain()
	smsText := map[string]string{"fa": "کد", "en": "Code", "ar": "رمز"}
	for _, it := range h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/"+sms["id"].(string), nil)["recipients"].(map[string]any)["items"].([]any) {
		r := it.(map[string]any)
		lang := expect[r["userId"].(string)]
		if r["status"] != "sent" || r["error"] != "sandbox" || r["lang"] != lang || r["message"] != smsText[lang] {
			t.Errorf("sms recipient = %v", r)
		}
	}
	emailText := map[string]string{"fa": "سلام\nمتن", "en": "Hello\nمتن", "ar": "مرحبا\nنص"}
	for _, it := range h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/"+email["id"].(string), nil)["recipients"].(map[string]any)["items"].([]any) {
		r := it.(map[string]any)
		lang := expect[r["userId"].(string)]
		if r["lang"] != lang || r["message"] != emailText[lang] {
			t.Errorf("email recipient = %v", r)
		}
	}

	cat := h.category("faq", "account")
	if _, err := h.s.db.Exec(`UPDATE base_category SET name='{"fa":"حساب","ar":"الحساب"}' WHERE id=$1`, cat); err != nil {
		t.Fatal(err)
	}
	h.mustAPI(http.StatusCreated, http.MethodPost, "/faqs", map[string]any{"question": map[string]string{"fa": "رمز؟", "ar": "كلمة المرور؟"}, "answer": map[string]string{"fa": "پاسخ", "ar": "**جواب**"}, "categoryId": cat})
	faq := h.mustRPC(fa, "publication_faq_list", map[string]any{"lang": "ar"})
	item := items(faq)[0].(map[string]any)
	if faq["lang"] != "ar" || item["question"] != "كلمة المرور؟" || item["answer"] != "**جواب**" || !strings.Contains(item["answerHtml"].(string), "<strong>جواب</strong>") || faq["categories"].([]any)[0].(map[string]any)["name"] != "الحساب" {
		t.Fatalf("faq ar = %v", faq)
	}
	if found := items(h.mustRPC(fa, "publication_faq_list", map[string]any{"q": "المرور"})); len(found) != 1 {
		t.Fatalf("faq search in ar = %v", found)
	}
	newsID := h.news(map[string]any{"title": map[string]string{"fa": "خبر", "ar": "أخبار"}, "summary": map[string]string{"fa": "خلاصه"}, "body": map[string]string{"fa": "متن", "ar": "نص الخبر"},
		"actions": []map[string]any{{"type": "url", "label": map[string]string{"fa": "باز کن", "ar": "افتح"}, "value": "https://example.com"}}})
	list := h.mustRPC(ar, "publication_news_list", map[string]any{})
	if list["lang"] != "ar" || items(list)[0].(map[string]any)["title"] != "أخبار" || items(list)[0].(map[string]any)["summary"] != "خلاصه" {
		t.Fatalf("news list ar = %v", list)
	}
	got := h.mustRPC(en, "publication_news_get", map[string]any{"id": newsID, "lang": "ar"})
	if got["body"] != "نص الخبر" || got["actions"].([]any)[0].(map[string]any)["label"] != "افتح" {
		t.Fatalf("news get ar = %v", got)
	}
	if got := h.mustRPC(en, "publication_news_get", map[string]any{"id": newsID}); got["lang"] != "en" || got["title"] != "خبر" {
		t.Fatalf("news en falls back to default = %v", got)
	}
}
