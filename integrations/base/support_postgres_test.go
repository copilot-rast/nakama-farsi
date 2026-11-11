package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/inapp"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
)

func newSupportHarness(t *testing.T) *harness {
	h := newHarness(t)
	for _, q := range []string{
		`DELETE FROM base_live_chat`, `DELETE FROM base_support_ticket`, `DELETE FROM base_support_agent`, `DELETE FROM base_support_file`, `DELETE FROM base_support_upload`,
		`DELETE FROM cluster_rate_limit WHERE key LIKE 'sup-%'`,
		`UPDATE base_support_settings SET max_open_tickets=10, tickets_per_hour=6, messages_per_hour=40, max_attachments=5, max_attachment_kb=2048, chat_enabled=true, chat_idle_minutes=30, chat_messages_per_minute=12`,
		`UPDATE base_support_department SET agents='{}', is_active=true`,
	} {
		if _, err := h.s.db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	store, err := objectstore.New(objectstore.Config{Driver: objectstore.DriverLocal, LocalDir: t.TempDir()}, nil)
	if err != nil {
		t.Fatal(err)
	}
	h.s.files, h.s.fileKey, h.s.publicBase = store, []byte("0123456789abcdef0123456789abcdef"), ""
	return h
}

func pngBytes(t *testing.T) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(1, 1, color.RGBA{255, 0, 0, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func (h *harness) upload(user string, data []byte) map[string]any {
	h.t.Helper()
	begin := h.mustRPC(user, "base_support_upload_begin", map[string]any{"name": "shot.png", "size": len(data)})
	parts := int(begin["parts"].(float64))
	for i := 0; i < parts; i++ {
		end := (i + 1) * uploadChunkBytes
		if end > len(data) {
			end = len(data)
		}
		h.mustRPC(user, "base_support_upload_part", map[string]any{"uploadId": begin["uploadId"], "index": i, "data": base64.StdEncoding.EncodeToString(data[i*uploadChunkBytes : end])})
	}
	return h.mustRPC(user, "base_support_upload_finish", map[string]any{"uploadId": begin["uploadId"]})
}

func TestSupportTicketFlow(t *testing.T) {
	h := newSupportHarness(t)
	var sent []inapp.Message
	var mu sync.Mutex
	inapp.SetSender(func(_ context.Context, m inapp.Message) (string, error) {
		mu.Lock()
		sent = append(sent, m)
		mu.Unlock()
		return "n1", nil
	})
	t.Cleanup(func() { inapp.SetSender(nil) })
	player := h.user("fa")
	other := h.user("en")

	info := h.mustRPC(player, "base_support_info", map[string]any{})
	if len(info["departments"].([]any)) < 2 || info["attachments"] != true || info["chatEnabled"] != true {
		t.Fatalf("info %v", info)
	}
	if code := h.rpcCode(player, "base_support_ticket_create", map[string]any{"subject": " ", "body": "x"}); code != "INVALID_SUBJECT" {
		t.Fatal(code)
	}
	if code := h.rpcCode(player, "base_support_ticket_create", map[string]any{"subject": "x", "body": "x", "departmentKey": "nope"}); code != "DEPARTMENT_NOT_FOUND" {
		t.Fatal(code)
	}
	if code := h.rpcCode(player, "base_support_upload_begin", map[string]any{"name": "a.png", "size": 3 * 1024 * 1024}); code != "FILE_TOO_LARGE" {
		t.Fatal(code)
	}
	att := h.upload(player, pngBytes(t))
	if att["type"] != "image/png" || !strings.HasSuffix(att["key"].(string), ".png") || att["url"] == "" {
		t.Fatalf("upload %v", att)
	}
	bad := h.mustRPC(player, "base_support_upload_begin", map[string]any{"name": "x.exe", "size": 10})
	h.mustRPC(player, "base_support_upload_part", map[string]any{"uploadId": bad["uploadId"], "index": 0, "data": base64.StdEncoding.EncodeToString([]byte("MZ12345678"))})
	if code := h.rpcCode(player, "base_support_upload_finish", map[string]any{"uploadId": bad["uploadId"]}); code != "UNSUPPORTED_FILE_TYPE" {
		t.Fatal(code)
	}
	if code := h.rpcCode(other, "base_support_ticket_create", map[string]any{"subject": "steal", "body": "x", "attachments": []string{att["key"].(string)}}); code != "ATTACHMENT_NOT_FOUND" {
		t.Fatal(code)
	}

	tk := h.mustRPC(player, "base_support_ticket_create", map[string]any{"subject": "سکه‌هایم نرسید", "departmentKey": "payments", "body": "خرید کردم ولی سکه نیامد", "attachments": []string{att["key"].(string)}})
	id := tk["id"].(string)
	if tk["status"] != "open" || !strings.HasPrefix(tk["number"].(string), "T-") || len(tk["messages"].([]any)) != 1 {
		t.Fatalf("ticket %v", tk)
	}
	first := tk["messages"].([]any)[0].(map[string]any)
	atts := first["attachments"].([]any)
	if len(atts) != 1 {
		t.Fatalf("attachments %v", first)
	}
	url := atts[0].(map[string]any)["url"].(string)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.SetPathValue("key", att["key"].(string))
	h.s.fileHTTP(rec, req)
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/png" || !strings.Contains(rec.Header().Get("Content-Security-Policy"), "sandbox") {
		t.Fatalf("file %d %v", rec.Code, rec.Header())
	}
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, strings.Replace(url, "&s=", "&s=00", 1), nil)
	req.SetPathValue("key", att["key"].(string))
	h.s.fileHTTP(rec, req)
	if rec.Code != 404 {
		t.Fatalf("forged signature served %d", rec.Code)
	}
	if code := h.rpcCode(other, "base_support_ticket_get", map[string]any{"id": id}); code != "TICKET_NOT_FOUND" {
		t.Fatal(code)
	}
	detail := h.mustAPI(200, "GET", "/support/tickets/"+id, nil)
	if detail["dueAt"] == nil || detail["firstResponseAt"] != nil || detail["assignee"] != "" {
		t.Fatalf("console ticket %v", detail)
	}

	code, note, _ := h.apiAs("sara", "POST", "/support/tickets/"+id+"/messages", map[string]any{"body": "بازیکن قبلاً هم این را گفته", "internal": true})
	if code != 200 || note["status"] != "open" || note["firstResponseAt"] != nil {
		t.Fatalf("note %d %v", code, note)
	}
	code, rep, _ := h.apiAs("sara", "POST", "/support/tickets/"+id+"/messages", map[string]any{"body": "سلام، بررسی کردیم."})
	if code != 200 || rep["status"] != "awaiting_player" || rep["assignee"] != "sara" || rep["firstResponseAt"] == nil || rep["dueAt"] != nil {
		t.Fatalf("reply %d %v", code, rep)
	}
	mine := h.mustRPC(player, "base_support_ticket_get", map[string]any{"id": id})
	for _, m := range mine["messages"].([]any) {
		mm := m.(map[string]any)
		if mm["kind"] == "note" || strings.Contains(mm["body"].(string), "قبلاً") {
			t.Fatalf("player sees the internal note: %v", mm)
		}
		if mm["kind"] == "event" && mm["event"] == "assigned" {
			t.Fatalf("player sees assignment: %v", mm)
		}
	}
	if len(sent) != 1 || sent[0].UserIDs[0] != player || sent[0].Action.Value != "support/ticket/"+id || !strings.Contains(sent[0].Title["fa"], "پاسخ") {
		t.Fatalf("in-app %+v", sent)
	}
	live := 0
	for _, n := range h.fake().notes {
		if n.code == notifySupport && n.user == player && !n.persistent {
			live++
		}
	}
	if live != 1 {
		t.Fatalf("live pushes %d", live)
	}
	if code := h.rpcCode(player, "base_support_ticket_rate", map[string]any{"id": id, "rating": 5}); code != "TICKET_NOT_FINISHED" {
		t.Fatal(code)
	}
	h.mustRPC(player, "base_support_ticket_reply", map[string]any{"id": id, "body": "هنوز نیامده"})
	if ticket := h.mustAPI(200, "GET", "/support/tickets/"+id, nil); ticket["status"] != "awaiting_staff" || ticket["dueAt"] == nil {
		t.Fatalf("after player reply %v", ticket)
	}
	h.expectError(400, "INVALID_STATUS", "POST", "/support/tickets/"+id+"/status", map[string]any{"status": "lost"})
	res := h.mustAPI(200, "POST", "/support/tickets/"+id+"/status", map[string]any{"status": "resolved"})
	if res["status"] != "resolved" || res["resolvedAt"] == nil {
		t.Fatalf("resolved %v", res)
	}
	rated := h.mustRPC(player, "base_support_ticket_rate", map[string]any{"id": id, "rating": 4, "comment": "ممنون"})
	if rated["rating"] != float64(4) || rated["canRate"] != false {
		t.Fatalf("rated %v", rated)
	}
	if code := h.rpcCode(player, "base_support_ticket_rate", map[string]any{"id": id, "rating": 1}); code != "ALREADY_RATED" {
		t.Fatal(code)
	}
	reopened := h.mustRPC(player, "base_support_ticket_reply", map[string]any{"id": id, "body": "یک سؤال دیگر"})
	if reopened["status"] != "awaiting_staff" {
		t.Fatalf("reply on resolved reopens: %v", reopened)
	}
	h.mustAPI(200, "POST", "/support/tickets/"+id+"/status", map[string]any{"status": "closed"})
	h.expectError(409, "INVALID_TRANSITION", "POST", "/support/tickets/"+id+"/status", map[string]any{"status": "resolved"})
	h.expectError(409, "INVALID_TRANSITION", "POST", "/support/tickets/"+id+"/status", map[string]any{"status": "awaiting_player"})
	if code := h.rpcCode(player, "base_support_ticket_reply", map[string]any{"id": id, "body": "x"}); code != "TICKET_CLOSED" {
		t.Fatal(code)
	}
	assigned := h.mustAPI(200, "POST", "/support/tickets/"+id+"/assign", map[string]any{"assignee": "reza", "priority": "urgent"})
	if assigned["assignee"] != "reza" || assigned["priority"] != "urgent" {
		t.Fatalf("assign %v", assigned)
	}
	list := items(h.mustAPI(200, "GET", "/support/tickets?assignee=reza&priority=urgent&q="+tk["number"].(string), nil))
	if len(list) != 1 {
		t.Fatalf("filtered list %v", list)
	}
	var acts int
	h.s.db.QueryRow(`SELECT count(*) FROM base_activity WHERE user_id=$1 AND kind='ticket'`, player).Scan(&acts)
	if acts < 4 {
		t.Fatalf("ticket activity %d", acts)
	}
}

func TestSupportTicketLimitsAndAutoAssign(t *testing.T) {
	h := newSupportHarness(t)
	player := h.user("fa")
	if _, err := h.s.db.Exec(`UPDATE base_support_settings SET max_open_tickets=2, tickets_per_hour=3`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.Exec(`UPDATE base_support_department SET agents='{ali,mina}' WHERE key='general'`); err != nil {
		t.Fatal(err)
	}
	a := h.mustRPC(player, "base_support_ticket_create", map[string]any{"subject": "a", "departmentKey": "general", "body": "a"})
	b := h.mustRPC(player, "base_support_ticket_create", map[string]any{"subject": "b", "departmentKey": "general", "body": "b"})
	if code := h.rpcCode(player, "base_support_ticket_create", map[string]any{"subject": "c", "body": "c"}); code != "TOO_MANY_OPEN_TICKETS" {
		t.Fatal(code)
	}
	ta := h.mustAPI(200, "GET", "/support/tickets/"+a["id"].(string), nil)
	tb := h.mustAPI(200, "GET", "/support/tickets/"+b["id"].(string), nil)
	if ta["assignee"] == tb["assignee"] || ta["assignee"] == "" || tb["assignee"] == "" {
		t.Fatalf("auto-assign balance %v %v", ta["assignee"], tb["assignee"])
	}
	h.mustRPC(player, "base_support_ticket_resolve", map[string]any{"id": a["id"]})
	d := h.mustRPC(player, "base_support_ticket_create", map[string]any{"subject": "d", "body": "d"})
	h.mustRPC(player, "base_support_ticket_resolve", map[string]any{"id": d["id"]})
	if code := h.rpcCode(player, "base_support_ticket_create", map[string]any{"subject": "e", "body": "e"}); code != "TOO_MANY_TICKETS" {
		t.Fatalf("hourly limit not enforced: %s", code)
	}
	if _, err := h.s.db.Exec(`UPDATE base_support_ticket SET status='awaiting_player', last_message_at=now() - INTERVAL '8 days' WHERE id=$1`, b["id"]); err != nil {
		t.Fatal(err)
	}
	n, err := h.s.autoCloseTickets(h.ctx)
	if err != nil || n < 1 {
		t.Fatalf("auto close %d %v", n, err)
	}
	if st := h.mustAPI(200, "GET", "/support/tickets/"+b["id"].(string), nil)["status"]; st != "closed" {
		t.Fatalf("idle ticket %v", st)
	}
	s := h.mustAPI(200, "GET", "/support/summary", nil)
	if s["open"] == nil || s["agentsOnline"] == nil {
		t.Fatalf("summary %v", s)
	}
}

func TestLiveChatTakeAndTransferRaces(t *testing.T) {
	h := newSupportHarness(t)
	player := h.user("fa")
	const n = 8
	var wg sync.WaitGroup
	ids := make([]string, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			out, err := h.rpc(player, "base_support_chat_open", map[string]any{"topic": "خرید"})
			if err == nil {
				ids[i] = out["chat"].(map[string]any)["id"].(string)
			}
		}(i)
	}
	wg.Wait()
	var open int
	h.s.db.QueryRow(`SELECT count(*) FROM base_live_chat WHERE user_id=$1 AND status <> 'closed'`, player).Scan(&open)
	if open != 1 {
		t.Fatalf("open chats %d", open)
	}
	chat := h.mustRPC(player, "base_support_chat_send", map[string]any{"chatId": ids[0], "body": "سلام، کسی هست؟"})["chat"].(map[string]any)
	id := chat["id"].(string)
	if chat["status"] != "waiting" {
		t.Fatalf("chat %v", chat)
	}
	queue := items(h.mustAPI(200, "GET", "/support/chats?view=queue", nil))
	if len(queue) != 1 || queue[0]["id"] != id {
		t.Fatalf("queue %v", queue)
	}
	agents := []string{"a1", "a2", "a3", "a4", "a5", "a6"}
	codes := make([]int, len(agents))
	for i, a := range agents {
		wg.Add(1)
		go func(i int, a string) {
			defer wg.Done()
			codes[i], _, _ = h.apiAs(a, "POST", "/support/chats/"+id+"/take", nil)
		}(i, a)
	}
	wg.Wait()
	winner := ""
	for i, c := range codes {
		switch c {
		case 200:
			if winner != "" {
				t.Fatalf("two agents took the chat: %v", codes)
			}
			winner = agents[i]
		case 409:
		default:
			t.Fatalf("take codes %v", codes)
		}
	}
	if winner == "" {
		t.Fatalf("nobody took the chat %v", codes)
	}
	loser := "a1"
	if winner == "a1" {
		loser = "a2"
	}
	if c, out, _ := h.apiAs(loser, "POST", "/support/chats/"+id+"/messages", map[string]any{"body": "x"}); c != 409 || out["error"] != "NOT_YOUR_CHAT" {
		t.Fatalf("other agent wrote %d %v", c, out)
	}
	if c, _, _ := h.apiAs(winner, "POST", "/support/chats/"+id+"/typing", nil); c != 200 {
		t.Fatal(c)
	}
	got := h.mustRPC(player, "base_support_chat_get", map[string]any{"chatId": id})["chat"].(map[string]any)
	if got["agentTyping"] != true || got["status"] != "active" || got["agent"] != winner {
		t.Fatalf("player view %v", got)
	}
	if c, _, _ := h.apiAs(winner, "POST", "/support/chats/"+id+"/messages", map[string]any{"body": "سلام، در خدمتم"}); c != 200 {
		t.Fatal(c)
	}
	h.mustRPC(player, "base_support_chat_typing", map[string]any{"chatId": id})
	code, state, _ := h.apiAs(winner, "GET", "/support/chats/"+id+"?after=0", nil)
	if code != 200 || state["playerTyping"] != true || state["playerOnline"] != true {
		t.Fatalf("agent view %d %v", code, state)
	}
	start := time.Now()
	code, polled, _ := h.apiAs(winner, "GET", "/support/chats/"+id+"?after=100&wait=1&stamp="+jsonEscape(state["stamp"].(string)), nil)
	if code != 200 || time.Since(start) < 800*time.Millisecond {
		t.Fatalf("long poll returned early %d %v %v", code, time.Since(start), polled["stamp"])
	}
	targets := []string{"t1", "t2"}
	for _, tname := range targets {
		h.s.db.Exec(`INSERT INTO base_support_agent (username) VALUES ($1) ON CONFLICT DO NOTHING`, tname)
	}
	tcodes := make([]int, 2)
	for i, tname := range targets {
		wg.Add(1)
		go func(i int, to string) {
			defer wg.Done()
			tcodes[i], _, _ = h.apiAs(winner, "POST", "/support/chats/"+id+"/transfer", map[string]any{"to": to})
		}(i, tname)
	}
	wg.Wait()
	if !((tcodes[0] == 200 && tcodes[1] == 409) || (tcodes[0] == 409 && tcodes[1] == 200)) {
		t.Fatalf("transfer race %v", tcodes)
	}
	var holder string
	h.s.db.QueryRow(`SELECT agent FROM base_live_chat WHERE id=$1`, id).Scan(&holder)
	if holder != "t1" && holder != "t2" {
		t.Fatalf("holder %s", holder)
	}
	if c, _, _ := h.apiAs(holder, "POST", "/support/chats/"+id+"/transfer", map[string]any{"to": "nobody-here"}); c != 400 {
		t.Fatalf("unknown agent %d", c)
	}
	c, closed, _ := h.apiAs(holder, "POST", "/support/chats/"+id+"/close", map[string]any{"toTicket": true})
	if c != 200 || closed["status"] != "closed" || closed["ticketId"] == nil {
		t.Fatalf("close %d %v", c, closed)
	}
	tk := h.mustAPI(200, "GET", "/support/tickets/"+closed["ticketId"].(string), nil)
	body := tk["messages"].([]any)[0].(map[string]any)["body"].(string)
	if tk["source"] != "chat" || !strings.Contains(body, "کسی هست") || !strings.Contains(body, "در خدمتم") || tk["assignee"] != holder {
		t.Fatalf("chat ticket %v", tk)
	}
	if code := h.rpcCode(player, "base_support_chat_send", map[string]any{"chatId": id, "body": "x"}); code != "CHAT_CLOSED" {
		t.Fatal(code)
	}
	rated := h.mustRPC(player, "base_support_chat_rate", map[string]any{"chatId": id, "rating": 5})["chat"].(map[string]any)
	if rated["rating"] != float64(5) || rated["canRate"] != false {
		t.Fatalf("rate %v", rated)
	}
	if code := h.rpcCode(player, "base_support_chat_rate", map[string]any{"chatId": id, "rating": 5}); code != "CANNOT_RATE" {
		t.Fatal(code)
	}
	fresh := h.mustRPC(player, "base_support_chat_open", map[string]any{"body": "دوباره سلام"})["chat"].(map[string]any)
	if fresh["id"] == id || fresh["status"] != "waiting" {
		t.Fatalf("new chat %v", fresh)
	}
	h.s.db.Exec(`UPDATE base_live_chat SET last_message_at=now() - INTERVAL '2 hours', player_seen_at=now() - INTERVAL '2 hours' WHERE id=$1`, fresh["id"])
	if n, err := h.s.idleCloseChats(h.ctx); err != nil || n != 1 {
		t.Fatalf("idle close %d %v", n, err)
	}
}

func jsonEscape(s string) string {
	r := strings.NewReplacer(`"`, "%22", "[", "%5B", "]", "%5D", ",", "%2C", " ", "%20")
	return r.Replace(s)
}
