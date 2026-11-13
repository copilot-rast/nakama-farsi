package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
)

func (h *harness) notifyReset() {
	h.t.Helper()
	for _, q := range []string{
		`TRUNCATE publication_push_delivery, publication_auto_fire, publication_push_device, publication_notify_pref CASCADE`,
		`DELETE FROM publication_auto_rule`,
		`DELETE FROM cluster_rate_limit WHERE key LIKE 'push:%'`,
		`UPDATE publication_auto_state SET event_cursor = now() - interval '1 hour'`,
		`DELETE FROM publication_notification WHERE category_id IN (SELECT id FROM publication_notify_category WHERE key LIKE 'nt\_%')`,
		`DELETE FROM publication_notify_category WHERE key LIKE 'nt\_%'`,
		`DELETE FROM base_tag WHERE key LIKE 'nt\_%'`,
		`DELETE FROM economy_chance_game WHERE key LIKE 'nt\_%'`,
		`DELETE FROM economy_streak WHERE key LIKE 'nt\_%'`,
		`UPDATE economy_event_settings SET timezone='Asia/Tehran' WHERE id=1`,
		`INSERT INTO publication_provider (id, channel, kind, is_active, is_default, sandbox, config) SELECT gen_random_uuid(), 'push', 'log', true, true, true, '' WHERE NOT EXISTS (SELECT 1 FROM publication_provider WHERE channel='push' AND kind='log')`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			h.t.Fatal(q, err)
		}
	}
}

func (h *harness) notifyCategory(key string, canOut bool) string {
	h.t.Helper()
	id := uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO publication_notify_category (id, key, name, can_opt_out) VALUES ($1,$2,'{"fa":"دسته","en":"Cat"}',$3)`, id, key, canOut); err != nil {
		h.t.Fatal(err)
	}
	return id
}

func (h *harness) device(user, token, platform, version string) {
	h.t.Helper()
	h.mustRPC(user, "publication_push_register", map[string]any{"token": token, "platform": platform, "appVersion": version})
}

func (h *harness) count(q string, args ...any) int {
	h.t.Helper()
	var n int
	if err := h.s.db.QueryRowContext(h.ctx, q, args...).Scan(&n); err != nil {
		h.t.Fatal(q, err)
	}
	return n
}

func TestPushDevicesRegisterMoveAndLimit(t *testing.T) {
	h := newHarness(t)
	h.notifyReset()
	a, b := h.user(nil, "", "fa"), h.user(nil, "", "en")
	if code := h.rpcCode(a, "publication_push_register", map[string]any{"token": "short", "platform": "web"}); code != "INVALID_TOKEN" {
		t.Fatalf("short token = %s", code)
	}
	if code := h.rpcCode(a, "publication_push_register", map[string]any{"token": "sandbox-web-aaaaaaaaaaaaaaaa", "platform": "symbian"}); code != "INVALID_PLATFORM" {
		t.Fatalf("platform = %s", code)
	}
	for i := 0; i < 12; i++ {
		h.device(a, fmt.Sprintf("sandbox-tok-%02d-aaaaaaaaaaaa", i), "android", "1.4.0")
		time.Sleep(2 * time.Millisecond)
	}
	if n := h.count(`SELECT count(*) FROM publication_push_device WHERE user_id=$1 AND is_active`, a); n != pushDevicesPerUser {
		t.Fatalf("active devices = %d", n)
	}
	h.device(b, "sandbox-tok-11-aaaaaaaaaaaa", "ios", "2.0.0")
	if n := h.count(`SELECT count(*) FROM publication_push_device WHERE token='sandbox-tok-11-aaaaaaaaaaaa' AND user_id=$1 AND platform='ios'`, b); n != 1 {
		t.Fatalf("token did not move to the new player")
	}
	list := h.mustRPC(a, "publication_push_devices", map[string]any{})
	first := items(list)[0].(map[string]any)
	if strings.Contains(first["token"].(string), "aaaaaaaaaaaa") {
		t.Fatalf("token not masked: %v", first)
	}
	out := h.mustRPC(a, "publication_push_unregister", map[string]any{"token": "sandbox-tok-10-aaaaaaaaaaaa"})
	if out["removed"] != true {
		t.Fatalf("unregister = %v", out)
	}
	if out := h.mustRPC(a, "publication_push_unregister", map[string]any{"token": "sandbox-tok-11-aaaaaaaaaaaa"}); out["removed"] != false {
		t.Fatalf("unregistered someone else's token: %v", out)
	}
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = h.rpc(b, "publication_push_register", map[string]any{"token": "sandbox-same-token-bbbbbbbbbb", "platform": "web"})
		}()
	}
	wg.Wait()
	if n := h.count(`SELECT count(*) FROM publication_push_device WHERE token='sandbox-same-token-bbbbbbbbbb'`); n != 1 {
		t.Fatalf("parallel register made %d rows", n)
	}
}

func TestNotifyPrefsAndOptOut(t *testing.T) {
	h := newHarness(t)
	h.notifyReset()
	u := h.user(nil, "", "en")
	h.notifyCategory("nt_offers", true)
	h.notifyCategory("nt_system", false)
	got := h.mustRPC(u, "publication_notify_prefs", map[string]any{"lang": "en"})
	byKey := map[string]map[string]any{}
	for _, it := range items(got) {
		m := it.(map[string]any)
		byKey[m["key"].(string)] = m
	}
	if byKey["nt_offers"]["enabled"] != true || byKey["nt_system"]["canOptOut"] != false {
		t.Fatalf("prefs = %v", got)
	}
	if code := h.rpcCode(u, "publication_notify_prefs_set", map[string]any{"prefs": map[string]bool{"nt_system": false}}); code != "CATEGORY_LOCKED" {
		t.Fatalf("locked = %s", code)
	}
	if code := h.rpcCode(u, "publication_notify_prefs_set", map[string]any{"prefs": map[string]bool{"nope": false}}); code != "CATEGORY_NOT_FOUND" {
		t.Fatalf("unknown = %s", code)
	}
	set := h.mustRPC(u, "publication_notify_prefs_set", map[string]any{"prefs": map[string]bool{"nt_offers": false}})
	for _, it := range items(set) {
		m := it.(map[string]any)
		if m["key"] == "nt_offers" && m["enabled"] != false {
			t.Fatalf("not saved: %v", m)
		}
	}
}

func TestPushCampaignAudienceOptOutOpenedAndDeadToken(t *testing.T) {
	h := newHarness(t)
	h.notifyReset()
	cat := h.notifyCategory("nt_offers", true)
	vip := h.tag("user", "nt_vip")
	a, b, c, d := h.user(nil, "", "fa"), h.user(nil, "", "en"), h.user(nil, "", "fa"), h.user(nil, "", "fa")
	for _, u := range []string{a, b, c, d} {
		if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO base_user_tag (user_id, tag_id) VALUES ($1,$2)`, u, vip); err != nil {
			t.Fatal(err)
		}
	}
	h.device(a, "sandbox-web-a-aaaaaaaaaaaa", "web", "1.4.0")
	h.device(b, "sandbox-dead-b-bbbbbbbbbbb", "android", "1.4.0")
	h.device(c, "sandbox-web-c-cccccccccccc", "web", "1.0.0")
	h.mustRPC(d, "publication_notify_prefs_set", map[string]any{"prefs": map[string]bool{"nt_offers": false}})
	h.device(d, "sandbox-web-d-dddddddddddd", "web", "1.4.0")

	if code, out := h.api(http.MethodPost, "/notifications", map[string]any{"channel": "push", "title": map[string]string{"fa": "t"}, "body": map[string]string{"fa": "b"}, "target": map[string]any{"mode": "all"}}); code != 400 || out["error"] != "CATEGORY_REQUIRED" {
		t.Fatalf("no category = %d %v", code, out)
	}
	prev := h.mustAPI(200, http.MethodPost, "/notifications/preview", map[string]any{"channel": "push_in_app", "target": map[string]any{"mode": "tags", "tagIds": []string{vip}},
		"audience": map[string]any{"minVersion": "1.2.0"}, "categoryId": cat})
	if num(prev["count"]) != 3 || num(prev["optedOut"]) != 1 || num(prev["withPush"]) != 2 {
		t.Fatalf("preview = %v", prev)
	}
	n := h.mustAPI(201, http.MethodPost, "/notifications", map[string]any{"channel": "push_in_app", "name": "VIP sale", "title": map[string]string{"fa": "حراج {player}", "en": "Sale"}, "body": map[string]string{"fa": "بیا", "en": "Come"},
		"imageUrl": "https://cdn.example.com/sale.png", "categoryId": cat, "action": map[string]any{"type": "screen", "value": "shop"},
		"target": map[string]any{"mode": "tags", "tagIds": []string{vip}}, "audience": map[string]any{"minVersion": "1.2.0"}})
	id := n["id"].(string)
	if n["kind"] != "campaign" || num(n["recipientCount"]) != 3 || num(n["skippedCount"]) != 1 {
		t.Fatalf("campaign = %v", n)
	}
	h.drain()
	got := h.mustAPI(200, http.MethodGet, "/notifications/"+id, nil)
	if got["status"] != "done" || num(got["sentCount"]) != 2 || num(got["pushSent"]) != 1 || num(got["pushFailed"]) != 1 {
		t.Fatalf("after send = %v", got)
	}
	if h.count(`SELECT count(*) FROM publication_push_device WHERE token LIKE 'sandbox-dead-%' AND NOT is_active`) != 1 {
		t.Fatal("dead token still active")
	}
	if h.count(`SELECT count(*) FROM publication_notification_recipient WHERE notification_id=$1 AND status='skipped' AND error='opted_out'`, id) != 1 {
		t.Fatal("opted out player not skipped")
	}
	inbox := h.mustRPC(a, "publication_inbox", map[string]any{"lang": "fa"})
	it := items(inbox)[0].(map[string]any)
	if !strings.HasPrefix(it["title"].(string), "حراج ") || it["imageUrl"] != "https://cdn.example.com/sale.png" || it["category"].(map[string]any)["key"] != "nt_offers" {
		t.Fatalf("inbox item = %v", it)
	}
	if h.count(`SELECT count(*) FROM publication_notification_recipient WHERE notification_id=$1 AND user_id=$2 AND message LIKE 'حراج %'`, id, a) != 1 {
		t.Fatal("sandbox message not stored")
	}
	for i := 0; i < 3; i++ {
		out := h.mustRPC(a, "publication_notification_opened", map[string]any{"notificationId": id})
		if (out["counted"] == true) != (i == 0) {
			t.Fatalf("opened %d = %v", i, out)
		}
	}
	if code := h.rpcCode(d, "publication_notification_opened", map[string]any{"notificationId": id}); code != "NOTIFICATION_NOT_FOUND" {
		t.Fatalf("opened by skipped player = %s", code)
	}
	if got := h.mustAPI(200, http.MethodGet, "/notifications/"+id, nil); num(got["openedCount"]) != 1 {
		t.Fatalf("opened count = %v", got["openedCount"])
	}
	test := h.mustAPI(201, http.MethodPost, "/notifications/test", map[string]any{"channel": "push", "title": map[string]string{"fa": "آزمایش"}, "body": map[string]string{"fa": "x"}, "categoryId": cat, "userId": d})
	if test["kind"] != "test" || num(test["recipientCount"]) != 1 || num(test["skippedCount"]) != 0 {
		t.Fatalf("test send = %v", test)
	}
}

func TestScheduledCampaignStartsOnceOnManyNodesAndCancel(t *testing.T) {
	h := newHarness(t)
	h.notifyReset()
	cat := h.notifyCategory("nt_news", true)
	users := []string{}
	for i := 0; i < 5; i++ {
		u := h.user(nil, "", "fa")
		users = append(users, u)
		h.device(u, fmt.Sprintf("sandbox-web-%d-aaaaaaaaaaa", i), "web", "1.0.0")
	}
	at := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	n := h.mustAPI(201, http.MethodPost, "/notifications", map[string]any{"channel": "push", "title": map[string]string{"fa": "بعدا"}, "body": map[string]string{"fa": "x"}, "categoryId": cat,
		"target": map[string]any{"mode": "users", "userIds": users}, "scheduledAt": at})
	id := n["id"].(string)
	if n["status"] != "scheduled" || num(n["recipientCount"]) != 0 {
		t.Fatalf("scheduled = %v", n)
	}
	if code, out := h.api(http.MethodPost, "/notifications", map[string]any{"channel": "push", "title": map[string]string{"fa": "x"}, "body": map[string]string{"fa": "x"}, "categoryId": cat,
		"target": map[string]any{"mode": "all"}, "scheduledAt": "2020-01-01T00:00:00Z"}); code != 400 || out["error"] != "SCHEDULE_IN_PAST" {
		t.Fatalf("past = %d %v", code, out)
	}
	h.drain()
	if h.count(`SELECT count(*) FROM publication_notification_recipient WHERE notification_id=$1`, id) != 0 {
		t.Fatal("scheduled campaign started early")
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_notification SET scheduled_at = now() - interval '1 second' WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	nodes := []*service{testService(), testService(), testService()}
	var wg sync.WaitGroup
	for _, node := range nodes {
		node.db, node.nk = h.s.db, h.nk
		wg.Add(1)
		go func(s *service) {
			defer wg.Done()
			for i := 0; i < 5; i++ {
				if _, err := s.runWorker(h.ctx); err != nil {
					t.Error(err)
				}
			}
		}(node)
	}
	wg.Wait()
	h.drain()
	got := h.mustAPI(200, http.MethodGet, "/notifications/"+id, nil)
	if got["status"] != "done" || num(got["recipientCount"]) != 5 || num(got["pushSent"]) != 5 {
		t.Fatalf("after start = %v", got)
	}
	if h.count(`SELECT count(*) FROM publication_push_delivery WHERE notification_id=$1`, id) != 5 {
		t.Fatal("a device got the push twice or not at all")
	}
	later := h.mustAPI(201, http.MethodPost, "/notifications", map[string]any{"channel": "push", "title": map[string]string{"fa": "لغو"}, "body": map[string]string{"fa": "x"}, "categoryId": cat,
		"target": map[string]any{"mode": "users", "userIds": users}, "scheduledAt": at})
	c := h.mustAPI(200, http.MethodPost, "/notifications/"+later["id"].(string)+"/cancel", nil)
	if c["status"] != "cancelled" || c["cancelledBy"] != "admin" {
		t.Fatalf("cancel = %v", c)
	}
	if code, out := h.api(http.MethodPost, "/notifications/"+later["id"].(string)+"/cancel", nil); code != 409 || out["error"] != "NOT_CANCELLABLE" {
		t.Fatalf("cancel twice = %d %v", code, out)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_notification SET scheduled_at = now() - interval '1 second' WHERE id=$1`, later["id"]); err != nil {
		t.Fatal(err)
	}
	h.drain()
	if h.count(`SELECT count(*) FROM publication_notification_recipient WHERE notification_id=$1`, later["id"]) != 0 {
		t.Fatal("cancelled campaign was sent")
	}
}

func testServiceAccount(t *testing.T, project string) string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, _ := x509.MarshalPKCS8PrivateKey(key)
	pemKey := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	raw, _ := json.MarshalIndent(map[string]string{"type": "service_account", "project_id": project, "private_key_id": "kid1", "private_key": pemKey, "client_email": "push@" + project + ".iam.gserviceaccount.com"}, "", "  ")
	return string(raw)
}

func TestFCMProviderTokenSendAndDeadTokens(t *testing.T) {
	h := newHarness(t)
	h.notifyReset()
	var mu sync.Mutex
	tokenCalls, sends := 0, map[string]int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/token" {
			_ = r.ParseForm()
			parts := strings.Split(r.Form.Get("assertion"), ".")
			claims, _ := base64.RawURLEncoding.DecodeString(parts[1])
			if r.Form.Get("grant_type") != "urn:ietf:params:oauth:grant-type:jwt-bearer" || !strings.Contains(string(claims), "firebase.messaging") || len(parts) != 3 {
				w.WriteHeader(400)
				_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
				return
			}
			mu.Lock()
			tokenCalls++
			mu.Unlock()
			_, _ = w.Write([]byte(`{"access_token":"ya29.test","expires_in":3600,"token_type":"Bearer"}`))
			return
		}
		if r.URL.Path != "/v1/projects/demo-proj/messages:send" || r.Header.Get("Authorization") != "Bearer ya29.test" {
			w.WriteHeader(401)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var in struct {
			Message struct {
				Token        string            `json:"token"`
				Notification map[string]string `json:"notification"`
				Data         map[string]string `json:"data"`
			} `json:"message"`
		}
		_ = json.Unmarshal(body, &in)
		mu.Lock()
		sends[in.Message.Token]++
		mu.Unlock()
		switch {
		case strings.HasPrefix(in.Message.Token, "fcm-gone"):
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"error":{"code":404,"message":"Requested entity was not found.","status":"NOT_FOUND","details":[{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":"UNREGISTERED"}]}}`))
		case strings.HasPrefix(in.Message.Token, "fcm-busy"):
			w.WriteHeader(503)
			_, _ = w.Write([]byte(`{"error":{"code":503,"status":"UNAVAILABLE"}}`))
		default:
			if in.Message.Notification["title"] == "" || in.Message.Data["notificationId"] == "" {
				w.WriteHeader(400)
				return
			}
			_, _ = w.Write([]byte(`{"name":"projects/demo-proj/messages/1"}`))
		}
	}))
	defer srv.Close()
	h.s.baseURLs["fcm"] = srv.URL
	h.s.baseURLs["fcmauth"] = srv.URL + "/token"
	sa := testServiceAccount(t, "demo-proj")
	if code, out := h.api(http.MethodPost, "/providers", map[string]any{"channel": "push", "kind": "fcm", "config": map[string]any{"serviceAccount": `{"type":"user"}`}}); code != 400 || out["error"] != "INVALID_SERVICE_ACCOUNT" {
		t.Fatalf("bad key = %d %v", code, out)
	}
	p := h.mustAPI(201, http.MethodPost, "/providers", map[string]any{"channel": "push", "kind": "fcm", "isActive": true, "isDefault": true, "sandbox": false, "config": map[string]any{"serviceAccount": sa}})
	raw, _ := json.Marshal(p)
	if strings.Contains(string(raw), "PRIVATE KEY") || p["config"].(map[string]any)["serviceAccount"] != secretMask || p["info"].(map[string]any)["projectId"] != "demo-proj" {
		t.Fatalf("provider view leaks or misses info: %v", p)
	}
	cat := h.notifyCategory("nt_offers", true)
	ok, gone, busy := h.user(nil, "", "fa"), h.user(nil, "", "fa"), h.user(nil, "", "fa")
	h.device(ok, "fcm-ok-aaaaaaaaaaaaaaaa", "android", "1.0.0")
	h.device(gone, "fcm-gone-aaaaaaaaaaaaaa", "android", "1.0.0")
	h.device(busy, "fcm-busy-aaaaaaaaaaaaaa", "ios", "1.0.0")
	n := h.mustAPI(201, http.MethodPost, "/notifications", map[string]any{"channel": "push", "title": map[string]string{"fa": "سلام"}, "body": map[string]string{"fa": "x"}, "categoryId": cat,
		"target": map[string]any{"mode": "users", "userIds": []string{ok, gone, busy}}})
	id := n["id"].(string)
	h.drain()
	if h.count(`SELECT count(*) FROM publication_push_device WHERE token LIKE 'fcm-gone%' AND NOT is_active`) != 1 {
		t.Fatal("UNREGISTERED token not switched off")
	}
	if h.count(`SELECT count(*) FROM publication_notification_recipient WHERE notification_id=$1 AND user_id=$2 AND status='sent' AND push_sent=1`, id, ok) != 1 {
		t.Fatal("ok device not sent")
	}
	if h.count(`SELECT count(*) FROM publication_notification_recipient WHERE notification_id=$1 AND user_id=$2 AND status='pending' AND attempts=1`, id, busy) != 1 {
		t.Fatal("503 should retry later")
	}
	h.dueNow()
	h.drain()
	mu.Lock()
	if sends["fcm-ok-aaaaaaaaaaaaaaaa"] != 1 || sends["fcm-gone-aaaaaaaaaaaaaa"] != 1 || sends["fcm-busy-aaaaaaaaaaaaaa"] < 2 || tokenCalls != 1 {
		t.Fatalf("sends = %v, token calls = %d", sends, tokenCalls)
	}
	mu.Unlock()
	if _, err := h.s.db.ExecContext(h.ctx, `DELETE FROM publication_provider WHERE kind='fcm'`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_provider SET is_default=true WHERE kind='log'`); err != nil {
		t.Fatal(err)
	}
}

func (h *harness) eventType(key string) string {
	h.t.Helper()
	var id string
	err := h.s.db.QueryRowContext(h.ctx, `INSERT INTO economy_event_type (id, key, name, client_allowed) VALUES (gen_random_uuid(), $1, '{"fa":"رویداد"}', true)
		ON CONFLICT (key) DO UPDATE SET key=EXCLUDED.key RETURNING id::text`, key).Scan(&id)
	if err != nil {
		h.t.Fatal(err)
	}
	return id
}

func (h *harness) event(user, typeID string, meta map[string]any) {
	h.t.Helper()
	raw, _ := json.Marshal(meta)
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_player_event (id, user_id, type_id, value, metadata, source) VALUES (gen_random_uuid(), $1, $2, 5, $3::jsonb, 'server')`, user, typeID, string(raw)); err != nil {
		h.t.Fatal(err)
	}
}

func TestAutoRulesFireOncePerOccurrenceOnManyNodes(t *testing.T) {
	h := newHarness(t)
	h.notifyReset()
	cat := h.notifyCategory("nt_rewards", true)
	et := h.eventType("nt_badge")
	a, b, c := h.user(nil, "", "fa"), h.user(nil, "", "fa"), h.user(nil, "", "fa")
	h.device(a, "sandbox-web-a-aaaaaaaaaaaa", "web", "1.0.0")
	h.mustRPC(c, "publication_notify_prefs_set", map[string]any{"prefs": map[string]bool{"nt_rewards": false}})
	if code, out := h.api(http.MethodPost, "/auto-rules", map[string]any{"key": "nt_rule", "trigger": "event", "eventKey": "missing", "categoryId": cat, "title": map[string]string{"fa": "x"}, "body": map[string]string{"fa": "x"}}); code != 400 || out["error"] != "EVENT_TYPE_NOT_FOUND" {
		t.Fatalf("missing event = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/auto-rules", map[string]any{"key": "nt_rule", "trigger": "event", "eventKey": "nt_badge", "categoryId": cat, "title": map[string]string{"fa": "x"}, "body": map[string]string{"fa": "x"}, "quietStart": "22:00"}); code != 400 || out["error"] != "INVALID_QUIET_HOURS" {
		t.Fatalf("half quiet = %d %v", code, out)
	}
	rule := h.mustAPI(201, http.MethodPost, "/auto-rules", map[string]any{"key": "nt_rule", "trigger": "event", "eventKey": "nt_badge", "matchField": "key", "matchValue": "gold",
		"channel": "push_in_app", "categoryId": cat, "title": map[string]string{"fa": "{player} طلا گرفت"}, "body": map[string]string{"fa": "امتیاز {value}"}, "cooldownMinutes": 60})
	h.event(a, et, map[string]any{"key": "gold"})
	h.event(a, et, map[string]any{"key": "gold"})
	h.event(a, et, map[string]any{"key": "silver"})
	h.event(b, et, map[string]any{"key": "gold"})
	h.event(c, et, map[string]any{"key": "gold"})
	nodes := []*service{testService(), testService(), testService()}
	var wg sync.WaitGroup
	for _, node := range nodes {
		node.db, node.nk = h.s.db, h.nk
		wg.Add(1)
		go func(s *service) {
			defer wg.Done()
			for i := 0; i < 3; i++ {
				if err := s.runAuto(h.ctx); err != nil {
					t.Error(err)
				}
			}
		}(node)
	}
	wg.Wait()
	if n := h.count(`SELECT count(*) FROM publication_auto_fire WHERE rule_id=$1`, rule["id"]); n != 4 {
		t.Fatalf("fires = %d, want 4 (one per matching event)", n)
	}
	if n := h.count(`SELECT count(*) FROM publication_auto_fire WHERE rule_id=$1 AND status='sent'`, rule["id"]); n != 2 {
		t.Fatalf("sent fires = %d, want 2", n)
	}
	if n := h.count(`SELECT count(*) FROM publication_auto_fire WHERE rule_id=$1 AND reason='cooldown'`, rule["id"]); n != 1 {
		t.Fatalf("cooldown fires = %d", n)
	}
	if n := h.count(`SELECT count(*) FROM publication_auto_fire WHERE rule_id=$1 AND reason='opted_out' AND user_id=$2`, rule["id"], c); n != 1 {
		t.Fatalf("opted out fires = %d", n)
	}
	h.drain()
	inbox := h.mustRPC(a, "publication_inbox", map[string]any{"lang": "fa"})
	it := items(inbox)[0].(map[string]any)
	if !strings.HasSuffix(it["title"].(string), " طلا گرفت") || it["body"] != "امتیاز 5" {
		t.Fatalf("rendered = %v", it)
	}
	if n := h.count(`SELECT count(*) FROM publication_push_delivery d JOIN publication_notification n ON n.id=d.notification_id WHERE n.rule_id=$1 AND d.status='sent'`, rule["id"]); n != 1 {
		t.Fatalf("pushes = %d", n)
	}
	for _, node := range nodes {
		if err := node.runAuto(h.ctx); err != nil {
			t.Fatal(err)
		}
	}
	if n := h.count(`SELECT count(*) FROM publication_auto_fire WHERE rule_id=$1`, rule["id"]); n != 4 {
		t.Fatalf("rescan added fires: %d", n)
	}
	fires := h.mustAPI(200, http.MethodGet, "/auto-rules/"+rule["id"].(string)+"/fires?status=skipped", nil)
	if num(fires["total"]) != 2 {
		t.Fatalf("fires list = %v", fires)
	}
	tst := h.mustAPI(201, http.MethodPost, "/auto-rules/"+rule["id"].(string)+"/test", map[string]any{"userId": c})
	if tst["kind"] != "test" || num(tst["recipientCount"]) != 1 {
		t.Fatalf("rule test = %v", tst)
	}
}

func TestAutoRulesQuietHoursAndTimeTriggers(t *testing.T) {
	h := newHarness(t)
	h.notifyReset()
	cat := h.notifyCategory("nt_reminders", true)
	zone, _ := time.LoadLocation("Asia/Tehran")
	now := time.Now().In(zone)
	start := (now.Hour()*60 + now.Minute() + 1440 - 30) % 1440
	end := (now.Hour()*60 + now.Minute() + 30) % 1440
	hm := func(m int) string { return fmt.Sprintf("%02d:%02d", m/60, m%60) }
	et := h.eventType("nt_ping")
	u := h.user(nil, "", "fa")
	quiet := h.mustAPI(201, http.MethodPost, "/auto-rules", map[string]any{"key": "nt_quiet", "trigger": "event", "eventKey": "nt_ping", "channel": "in_app", "categoryId": cat,
		"title": map[string]string{"fa": "x"}, "body": map[string]string{"fa": "y"}, "cooldownMinutes": 0, "quietStart": hm(start), "quietEnd": hm(end)})
	h.event(u, et, nil)
	if err := h.s.runAuto(h.ctx); err != nil {
		t.Fatal(err)
	}
	var due time.Time
	if err := h.s.db.QueryRowContext(h.ctx, `SELECT due_at FROM publication_auto_fire WHERE rule_id=$1 AND status='pending'`, quiet["id"]).Scan(&due); err != nil {
		t.Fatalf("quiet fire not pending: %v", err)
	}
	if d := due.Sub(time.Now()); d < 25*time.Minute || d > 31*time.Minute {
		t.Fatalf("due in %v, want about 30 minutes", d)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_auto_rule SET is_active=false WHERE id=$1`, quiet["id"]); err != nil {
		t.Fatal(err)
	}

	var gameID string
	if err := h.s.db.QueryRowContext(h.ctx, `INSERT INTO economy_chance_game (id, key, name, interval_time, prize) VALUES (gen_random_uuid(), 'nt_wheel', '{"fa":"گردونه"}', 60, '[]') ON CONFLICT (key) DO UPDATE SET interval_time=60 RETURNING id::text`).Scan(&gameID); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_chance_play (id, game_id, user_id, mode, status, roll, created_at) VALUES (gen_random_uuid(), $1, $2, 'free', 'done', 1, now() - interval '61 minutes')`, gameID, u); err != nil {
		t.Fatal(err)
	}
	ready := h.mustAPI(201, http.MethodPost, "/auto-rules", map[string]any{"key": "nt_ready", "trigger": "chance_ready", "params": map[string]any{"gameKey": "nt_wheel"}, "channel": "in_app", "categoryId": cat,
		"title": map[string]string{"fa": "{game} آماده است"}, "body": map[string]string{"fa": "بچرخان"}, "cooldownMinutes": 0})
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE users SET create_time = now() - interval '4 days' WHERE id=$1`, u); err != nil {
		t.Fatal(err)
	}
	away := h.mustAPI(201, http.MethodPost, "/auto-rules", map[string]any{"key": "nt_away", "trigger": "inactive", "params": map[string]any{"days": 3}, "channel": "in_app", "categoryId": cat,
		"title": map[string]string{"fa": "{days} روز"}, "body": map[string]string{"fa": "برگرد"}, "cooldownMinutes": 0})
	var streakID string
	if err := h.s.db.QueryRowContext(h.ctx, `INSERT INTO economy_streak (id, key, name, event_type_id, period_days, grace_periods) VALUES (gen_random_uuid(), 'nt_daily', '{"fa":"ورود روزانه"}', $1, 1, 0) RETURNING id::text`, et).Scan(&streakID); err != nil {
		t.Fatal(err)
	}
	idx := civilDayIndex(time.Now(), zone)
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO economy_user_streak (streak_id, user_id, count, last_period) VALUES ($1,$2,4,$3)`, streakID, u, idx-1); err != nil {
		t.Fatal(err)
	}
	risk := h.mustAPI(201, http.MethodPost, "/auto-rules", map[string]any{"key": "nt_risk", "trigger": "streak_at_risk", "params": map[string]any{"streakKey": "nt_daily", "hoursBefore": 23}, "channel": "in_app", "categoryId": cat,
		"title": map[string]string{"fa": "{count} روز {streak}"}, "body": map[string]string{"fa": "{hours} ساعت"}, "cooldownMinutes": 0})
	if code, out := h.api(http.MethodPost, "/auto-rules", map[string]any{"key": "nt_risk2", "trigger": "streak_at_risk", "params": map[string]any{"streakKey": "nope"}, "categoryId": cat,
		"title": map[string]string{"fa": "x"}, "body": map[string]string{"fa": "x"}}); code != 400 || out["error"] != "STREAK_NOT_FOUND" {
		t.Fatalf("missing streak = %d %v", code, out)
	}
	for i := 0; i < 2; i++ {
		if err := h.s.runAuto(h.ctx); err != nil {
			t.Fatal(err)
		}
	}
	if now.Hour() >= 1 {
		if n := h.count(`SELECT count(*) FROM publication_auto_fire WHERE rule_id=$1 AND status='sent' AND vars->>'count'='4' AND vars->>'streak'='ورود روزانه'`, risk["id"]); n != 1 {
			t.Fatalf("streak fires = %d", n)
		}
	}
	if _, err := h.s.db.ExecContext(h.ctx, `UPDATE publication_auto_rule SET is_active=false WHERE id=$1`, risk["id"]); err != nil {
		t.Fatal(err)
	}
	if n := h.count(`SELECT count(*) FROM publication_auto_fire WHERE rule_id=$1 AND status='sent'`, ready["id"]); n != 1 {
		t.Fatalf("chance ready fires = %d", n)
	}
	if n := h.count(`SELECT count(*) FROM publication_auto_fire WHERE rule_id=$1 AND status='sent' AND user_id=$2`, away["id"], u); n != 1 {
		t.Fatalf("inactive fires = %d", n)
	}
	h.drain()
	titles := map[string]bool{}
	for _, it := range items(h.mustRPC(u, "publication_inbox", map[string]any{"lang": "fa"})) {
		titles[it.(map[string]any)["title"].(string)] = true
	}
	if !titles["3 روز"] || !titles["گردونه آماده است"] || (now.Hour() >= 1 && !titles["4 روز ورود روزانه"]) {
		t.Fatalf("inbox titles = %v", titles)
	}
}
