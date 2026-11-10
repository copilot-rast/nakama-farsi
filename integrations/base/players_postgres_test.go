package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/restrict"
)

func (h *harness) fake() *fakeNakama { return h.s.nk.(*fakeNakama) }

func noticeOf(t *testing.T, err error) restrict.Notice {
	t.Helper()
	var rt *runtime.Error
	if !errors.As(err, &rt) {
		t.Fatalf("want a runtime error with a notice, got %v", err)
	}
	var n restrict.Notice
	if json.Unmarshal([]byte(rt.Message), &n) != nil {
		t.Fatalf("message is not a notice: %q", rt.Message)
	}
	return n
}

func until(t *testing.T, iso string) time.Duration {
	t.Helper()
	at, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		t.Fatal(err)
	}
	return time.Until(at)
}

func TestRestrictionTiming(t *testing.T) {
	h := newHarness(t)
	h.s.db.Exec(`DELETE FROM base_restriction`)
	u := h.user("fa")
	reason := map[string]string{"fa": "به خاطر فحاشی", "en": "For swearing"}

	h.expectError(400, "END_REQUIRED", "POST", "/players/"+u+"/restrictions", map[string]any{"kind": "suspend", "reason": reason})
	h.expectError(400, "END_IN_PAST", "POST", "/players/"+u+"/restrictions", map[string]any{"kind": "suspend", "endsAt": time.Now().Add(-time.Hour).Format(time.RFC3339), "reason": reason})
	h.expectError(400, "END_TOO_FAR", "POST", "/players/"+u+"/restrictions", map[string]any{"kind": "suspend", "endsAt": time.Now().AddDate(11, 0, 0).Format(time.RFC3339), "reason": reason})
	h.expectError(400, "DEFAULT_LANGUAGE_REQUIRED", "POST", "/players/"+u+"/restrictions", map[string]any{"kind": "mute", "minutes": 5, "reason": map[string]string{"en": "x"}})
	h.expectError(400, "INVALID_RESTRICTION_KIND", "POST", "/players/"+u+"/restrictions", map[string]any{"kind": "ban", "minutes": 5, "reason": reason})
	h.expectError(404, "USER_NOT_FOUND", "POST", "/players/"+uuid.Must(uuid.NewV4()).String()+"/restrictions", map[string]any{"kind": "mute", "minutes": 5, "reason": reason})

	sus := h.mustAPI(201, "POST", "/players/"+u+"/restrictions", map[string]any{"kind": "suspend", "minutes": 60, "reason": reason, "note": "internal"})
	if sus["status"] != "active" || sus["createdBy"] != "" || sus["note"] != "internal" {
		t.Fatalf("suspension %v", sus)
	}
	if d := until(t, sus["endsAt"].(string)); d < 59*time.Minute || d > 61*time.Minute {
		t.Fatalf("ends in %v", d)
	}
	f := h.fake()
	if len(f.notes) != 1 || f.notes[0].code != restrict.NotifySuspended || f.notes[0].content["reason"] != "به خاطر فحاشی" || f.notes[0].persistent {
		t.Fatalf("suspend notice %+v", f.notes)
	}
	if len(f.logouts) != 1 || f.logouts[0] != u {
		t.Fatalf("logouts %v", f.logouts)
	}
	n := noticeOf(t, h.s.refuseSuspended(h.ctx, u))
	if n.Code != restrict.CodeSuspended || n.Reason != "به خاطر فحاشی" || n.Lang != "fa" {
		t.Fatalf("notice %+v", n)
	}
	if d := until(t, n.Until); d < 59*time.Minute {
		t.Fatalf("notice until %v", d)
	}
	if _, err := h.s.db.Exec(`UPDATE users SET lang_tag='en' WHERE id=$1`, u); err != nil {
		t.Fatal(err)
	}
	if n := noticeOf(t, h.s.refuseSuspended(h.ctx, u)); n.Reason != "For swearing" || n.Lang != "en" {
		t.Fatalf("english notice %+v", n)
	}

	m1 := h.mustAPI(201, "POST", "/players/"+u+"/restrictions", map[string]any{"kind": "mute", "endsAt": time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339), "reason": reason})
	m2 := h.mustAPI(201, "POST", "/players/"+u+"/restrictions", map[string]any{"kind": "mute", "minutes": 30, "reason": reason})
	old := h.mustAPI(200, "GET", "/restrictions/"+m1["id"].(string), nil)
	if old["status"] != "lifted" || old["liftNote"] != "replaced" {
		t.Fatalf("a new mute must replace the old one: %v", old)
	}
	if r, _ := restrict.Active(h.ctx, h.s.db, u, restrict.KindMute); r == nil || r.ID != m2["id"] {
		t.Fatalf("active mute %v", r)
	}
	err := h.s.chatGate(h.ctx, u)
	if n := noticeOf(t, err); n.Code != restrict.CodeMuted {
		t.Fatalf("a suspended or muted player cannot chat: %+v", n)
	}

	warn := h.mustAPI(201, "POST", "/players/"+u+"/restrictions", map[string]any{"kind": "warn", "reason": reason})
	if warn["status"] != "expired" {
		t.Fatalf("warn %v", warn)
	}
	last := f.notes[len(f.notes)-1]
	if last.code != restrict.NotifyWarned || !last.persistent {
		t.Fatalf("warn notice %+v", last)
	}

	if _, err := h.s.db.Exec(`UPDATE base_restriction SET ends_at = now() - interval '1 second', starts_at = now() - interval '1 hour' WHERE user_id=$1 AND kind='suspend' AND status='active'`, u); err != nil {
		t.Fatal(err)
	}
	if err := h.s.refuseSuspended(h.ctx, u); err != nil {
		t.Fatalf("a suspension past its end must not refuse sign-in, even before the worker runs: %v", err)
	}
	list := h.mustAPI(200, "GET", "/restrictions?userId="+u+"&kind=suspend", nil)
	if items(list)[0]["status"] != "expired" {
		t.Fatalf("effective status %v", items(list)[0])
	}
	n2, err := h.s.expireRestrictions(h.ctx)
	if err != nil || n2 < 1 {
		t.Fatalf("expired %d %v", n2, err)
	}
	h.expectError(409, "NOT_ACTIVE", "POST", "/restrictions/"+sus["id"].(string)+"/lift", map[string]any{})

	lifted := h.mustAPI(200, "POST", "/restrictions/"+m2["id"].(string)+"/lift", map[string]any{"note": "sorry"})
	if lifted["status"] != "lifted" || lifted["liftNote"] != "sorry" {
		t.Fatalf("lift %v", lifted)
	}
	if err := h.s.chatGate(h.ctx, u); err != nil {
		t.Fatalf("lifted mute still blocks: %v", err)
	}
	if f.notes[len(f.notes)-1].code != restrict.NotifyLifted {
		t.Fatalf("lift notice %+v", f.notes[len(f.notes)-1])
	}
	status := h.mustRPC(u, "base_my_status", map[string]any{"lang": "fa"})
	if status["suspension"] != nil || status["mute"] != nil || len(status["warnings"].([]any)) != 1 {
		t.Fatalf("status %v", status)
	}
	player := h.mustAPI(200, "GET", "/players/"+u, nil)
	if len(player["history"].([]any)) != 4 {
		t.Fatalf("history %v", player["history"])
	}
}

func TestPlayerTags(t *testing.T) {
	h := newHarness(t)
	u1, u2, u3 := h.user("fa"), h.user("en"), h.user("fa")
	public := h.mustAPI(201, "POST", "/tags", map[string]any{"key": "vip", "keyType": "user", "name": map[string]string{"fa": "ویژه", "en": "VIP"}, "isPublic": true})
	hidden := h.mustAPI(201, "POST", "/tags", map[string]any{"key": "watch", "keyType": "user", "name": map[string]string{"fa": "زیر نظر"}})
	general := h.mustAPI(201, "POST", "/tags", map[string]any{"key": "gen", "keyType": "general", "name": map[string]string{"fa": "عمومی"}, "isPublic": true})
	if general["isPublic"] != false || public["isPublic"] != true {
		t.Fatalf("only player tags can be public: %v %v", general, public)
	}
	pid, hid := public["id"].(string), hidden["id"].(string)

	h.expectError(400, "TAG_KIND_MISMATCH", "PUT", "/players/"+u1+"/tags", map[string]any{"tagIds": []string{general["id"].(string)}})
	out := h.mustAPI(200, "PUT", "/players/"+u1+"/tags", map[string]any{"tagIds": []string{pid, hid}})
	if len(out["tags"].([]any)) != 2 {
		t.Fatalf("tags %v", out)
	}
	mine := h.mustRPC(u1, "base_my_tags", map[string]any{"lang": "en"})
	if got := items(mine); len(got) != 1 || got[0]["key"] != "vip" || got[0]["name"] != "VIP" {
		t.Fatalf("a player sees only public tags: %v", got)
	}
	if got := items(h.mustRPC(u1, "base_tags", map[string]any{})); len(got) != 1 || got[0]["key"] != "gen" {
		t.Fatalf("player tags stay out of base_tags: %v", got)
	}

	res := h.mustAPI(200, "POST", "/players/tags", map[string]any{"userIds": []string{u1, u2, u3}, "tagIds": []string{pid}, "op": "add"})
	if res["changed"].(float64) != 2 {
		t.Fatalf("bulk add %v", res)
	}
	if got := h.mustAPI(200, "GET", "/players?tagId="+pid, nil); got["total"].(float64) != 3 {
		t.Fatalf("filter %v", got["total"])
	}
	if got := h.mustAPI(200, "GET", "/tags/"+pid, nil); got["players"].(float64) != 3 {
		t.Fatalf("player count %v", got)
	}
	h.expectError(400, "IN_USE", "PUT", "/tags/"+pid, map[string]any{"key": "vip", "keyType": "general", "name": map[string]string{"fa": "ویژه"}})
	res = h.mustAPI(200, "POST", "/players/tags", map[string]any{"userIds": []string{u2, u3}, "tagIds": []string{pid}, "op": "remove"})
	if res["changed"].(float64) != 2 {
		t.Fatalf("bulk remove %v", res)
	}
	h.expectError(400, "USERS_NOT_FOUND", "POST", "/players/tags", map[string]any{"userIds": []string{uuid.Must(uuid.NewV4()).String()}, "tagIds": []string{pid}, "op": "add"})
	h.expectError(400, "INVALID_OP", "POST", "/players/tags", map[string]any{"userIds": []string{u1}, "tagIds": []string{pid}, "op": "toggle"})
	h.mustAPI(200, "DELETE", "/tags/"+hid, nil)
	if got := h.mustAPI(200, "GET", "/players/"+u1, nil); len(got["tags"].([]any)) != 1 {
		t.Fatalf("deleting a tag removes it from players: %v", got["tags"])
	}
}

func (h *harness) message(room, sender, username, text string, at time.Time) string {
	id := uuid.Must(uuid.NewV4()).String()
	content, _ := json.Marshal(map[string]string{"text": text})
	if _, err := h.s.db.Exec(`INSERT INTO message (id, code, sender_id, username, stream_mode, stream_subject, stream_descriptor, stream_label, content, create_time, update_time)
		VALUES ($1, 0, $2, $3, 2, '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000', $4, $5, $6, $6)`, id, sender, username, room, string(content), at); err != nil {
		h.t.Fatal(err)
	}
	return id
}

func TestReportFlow(t *testing.T) {
	h := newHarness(t)
	room := "unit-" + strings.ReplaceAll(uuid.Must(uuid.NewV4()).String()[:8], "-", "")
	t.Cleanup(func() {
		h.s.db.Exec(`DELETE FROM base_chat_report WHERE channel_label=$1`, room)
		h.s.db.Exec(`DELETE FROM message WHERE stream_label=$1`, room)
	})
	sender, reporter, other := h.user("fa"), h.user("fa"), h.user("en")
	t0 := time.Now().Add(-time.Minute)
	h.message(room, reporter, "rep", "سلام", t0)
	bad := h.message(room, sender, "snd", "حرف زشت", t0.Add(time.Second))
	h.message(room, reporter, "rep", "چرا؟", t0.Add(2*time.Second))

	if code := h.rpcCode(reporter, "base_report_message", map[string]any{"messageId": bad, "titleKey": "nope"}); code != "REPORT_TITLE_NOT_FOUND" {
		t.Fatal(code)
	}
	if code := h.rpcCode(sender, "base_report_message", map[string]any{"messageId": bad, "titleKey": "abuse"}); code != "CANNOT_REPORT_OWN" {
		t.Fatal(code)
	}
	if code := h.rpcCode(reporter, "base_report_message", map[string]any{"messageId": uuid.Must(uuid.NewV4()).String(), "titleKey": "abuse"}); code != "MESSAGE_NOT_FOUND" {
		t.Fatal(code)
	}
	first := h.mustRPC(reporter, "base_report_message", map[string]any{"messageId": bad, "titleKey": "abuse", "note": "بد گفت"})
	if first["duplicate"] != false {
		t.Fatal(first)
	}
	if again := h.mustRPC(reporter, "base_report_message", map[string]any{"messageId": bad, "titleKey": "spam"}); again["duplicate"] != true || again["reportId"] != first["reportId"] {
		t.Fatalf("a player reports a message once: %v", again)
	}
	second := h.mustRPC(other, "base_report_message", map[string]any{"messageId": bad, "titleKey": "harassment"})
	titles := h.mustRPC(other, "base_report_titles", map[string]any{})
	if got := items(titles); len(got) < 6 || got[0]["key"] != "abuse" || got[0]["name"] != "Insults or swearing" {
		t.Fatalf("titles in the player's language: %v", got)
	}

	rid := first["reportId"].(string)
	detail := h.mustAPI(200, "GET", "/reports/"+rid, nil)
	ctx := detail["context"].([]any)
	if len(ctx) != 3 || ctx[1].(map[string]any)["reported"] != true || ctx[1].(map[string]any)["content"] != "حرف زشت" || ctx[0].(map[string]any)["content"] != "سلام" {
		t.Fatalf("context %v", ctx)
	}
	if detail["reports"].(float64) != 2 || len(detail["others"].([]any)) != 1 || detail["title"].(map[string]any)["fa"] != "توهین یا فحاشی" {
		t.Fatalf("detail %v", detail)
	}
	if detail["sender"].(map[string]any)["id"] != sender {
		t.Fatalf("sender %v", detail["sender"])
	}
	if got := h.mustAPI(200, "GET", "/reports/summary", nil); got["open"].(float64) < 2 {
		t.Fatalf("summary %v", got)
	}

	h.expectError(400, "END_REQUIRED", "POST", "/reports/"+rid+"/resolve", map[string]any{"action": "mute", "reason": map[string]string{"fa": "آرام"}})
	out := h.mustAPI(200, "POST", "/reports/"+rid+"/resolve", map[string]any{"action": "mute", "minutes": 60, "reason": map[string]string{"fa": "آرام"}, "deleteMessage": true, "note": "ok"})
	if out["status"] != "resolved" || out["action"] != "mute" || out["messageDeleted"] != true || out["restrictionId"] == nil {
		t.Fatalf("resolve %v", out)
	}
	if r := h.mustAPI(200, "GET", "/reports/"+second["reportId"].(string), nil); r["status"] != "resolved" {
		t.Fatalf("every open report on the message closes together: %v", r)
	}
	if len(h.fake().removed) != 1 || !strings.HasPrefix(h.fake().removed[0], "2...") {
		t.Fatalf("removed %v", h.fake().removed)
	}
	if n := noticeOf(t, h.s.chatGate(h.ctx, sender)); n.Reason != "آرام" {
		t.Fatalf("mute from the report %+v", n)
	}
	h.expectError(409, "REPORT_CLOSED", "POST", "/reports/"+rid+"/resolve", map[string]any{"action": "dismiss"})

	other2 := h.message(room, sender, "snd", "دوباره", time.Now())
	r3 := h.mustRPC(reporter, "base_report_message", map[string]any{"messageId": other2, "titleKey": "other"})
	if out := h.mustAPI(200, "POST", "/reports/"+r3["reportId"].(string)+"/resolve", map[string]any{"action": "dismiss"}); out["status"] != "dismissed" {
		t.Fatal(out)
	}
	if got := h.mustAPI(200, "GET", "/reports?status=open&senderId="+sender, nil); got["total"].(float64) != 0 {
		t.Fatalf("open %v", got)
	}
}

func TestChatModerationContent(t *testing.T) {
	h := newHarness(t)
	t.Cleanup(func() { h.s.db.Exec(`DELETE FROM base_chat_word WHERE word LIKE 'unit%'`) })
	u := h.user("fa")
	out, err := h.s.moderateContent(h.ctx, u, `{"text":"تو احمــق هستی","meta":{"to":"كثافت"}}`)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	_ = json.Unmarshal([]byte(out), &doc)
	if doc["text"] != "تو **** هستی" || doc["meta"].(map[string]any)["to"] != "*****" {
		t.Fatalf("masked %v", doc)
	}
	if _, err := h.s.moderateContent(h.ctx, u, `{"text":"ح ر ا م ز ا د ه"}`); err == nil || err.Error() != CodeChatBlocked {
		t.Fatalf("blocked %v", err)
	}
	same := `{"text":"سلام دوستان"}`
	if out, err := h.s.moderateContent(h.ctx, u, same); err != nil || out != same {
		t.Fatalf("clean text must stay byte for byte: %q %v", out, err)
	}
	h.mustAPI(201, "POST", "/chat-words", map[string]any{"word": "unitword", "action": "block"})
	h.expectError(409, "ALREADY_EXISTS", "POST", "/chat-words", map[string]any{"word": "UNITWOOORD", "action": "mask"})
	h.expectError(400, "INVALID_WORD", "POST", "/chat-words", map[string]any{"word": "۱۲۳", "action": "mask"})
	h.expectError(400, "INVALID_MOD_ACTION", "POST", "/chat-words", map[string]any{"word": "unitx", "action": "ban"})
	test := h.mustAPI(200, "POST", "/chat-words/test", map[string]any{"text": "u n i t w o r d"})
	if test["blocked"] != true || test["normalized"] != "unitword" {
		t.Fatalf("test %v", test)
	}
}
