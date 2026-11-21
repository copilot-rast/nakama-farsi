package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/commpacks"
)

func resetComms(h *harness) {
	h.t.Helper()
	resetStickers(h)
	for _, q := range []string{`DELETE FROM base_comm_settings`, `DELETE FROM base_comm_channel`, `DELETE FROM base_comm_favorite`} {
		if _, err := h.s.db.Exec(q); err != nil {
			h.t.Fatal(err)
		}
	}
	h.t.Cleanup(func() {
		_, _ = h.s.db.Exec(`DELETE FROM base_comm_settings`)
		_, _ = h.s.db.Exec(`DELETE FROM base_comm_channel`)
	})
}

func refusal(t *testing.T, err error) string {
	t.Helper()
	r, ok := commpacks.AsRefusal(err)
	if !ok {
		t.Fatalf("not a refusal: %v", err)
	}
	return r.Code
}

func TestCommPacksConsoleRPCAndMatchSend(t *testing.T) {
	h := newHarness(t)
	resetComms(h)
	item := h.stickerItem("comm_vip_item")

	h.expectError(400, "INVALID_KIND", "POST", "/stickers/packs", map[string]any{"key": "hi", "kind": "voice", "name": map[string]string{"fa": "سلام"}, "stickers": []map[string]any{{"key": "a", "imageUrl": "/u/a.png"}}})
	h.expectError(400, "PLACE_REQUIRED", "POST", "/stickers/packs", map[string]any{"key": "hi", "kind": "text", "inChat": false, "name": map[string]string{"fa": "سلام"}, "stickers": []map[string]any{{"key": "a", "text": map[string]string{"fa": "سلام"}}}})
	h.expectError(400, "DEFAULT_LANGUAGE_REQUIRED", "POST", "/stickers/packs", map[string]any{"key": "hi", "kind": "text", "inMatch": true, "name": map[string]string{"fa": "سلام"}, "stickers": []map[string]any{{"key": "a"}}})
	h.expectError(400, "INVALID_ASSET_ID", "POST", "/stickers/packs", map[string]any{"key": "hi", "kind": "emoji", "inMatch": true, "name": map[string]string{"fa": "سلام"}, "stickers": []map[string]any{{"key": "a", "assetId": "Bad Id"}}})

	text := h.mustAPI(201, "POST", "/stickers/packs", map[string]any{"key": "hello", "kind": "text", "inChat": false, "inMatch": true, "sort": -1, "name": map[string]string{"fa": "سلام‌ها", "en": "Hellos"},
		"stickers": []map[string]any{{"key": "hi", "text": map[string]string{"fa": "سلام!", "en": "Hi!"}, "imageUrl": "/u/x.png"}, {"key": "gg", "text": map[string]string{"fa": "بازی خوبی بود", "en": "Good game"}}}})
	if text["kind"] != "text" || text["inChat"] != false || text["inMatch"] != true {
		t.Fatalf("text pack %v", text)
	}
	first := text["stickers"].([]any)[0].(map[string]any)
	if first["imageUrl"] != "" || first["text"].(map[string]any)["en"] != "Hi!" {
		t.Fatalf("text item %v", first)
	}
	emoji := h.mustAPI(201, "POST", "/stickers/packs", map[string]any{"key": "faces", "kind": "emoji", "inMatch": true, "access": "item", "itemKey": "comm_vip_item", "name": map[string]string{"fa": "چهره"},
		"stickers": []map[string]any{{"key": "angry", "assetId": "emoji.angry"}}})
	h.expectError(400, "KIND_LOCKED", "PUT", "/stickers/packs/"+emoji["id"].(string), map[string]any{"key": "faces", "kind": "gif", "inMatch": true, "access": "item", "itemKey": "comm_vip_item", "name": map[string]string{"fa": "چهره"}, "stickers": []map[string]any{{"key": "angry", "assetId": "emoji.angry"}}})
	h.mustAPI(201, "POST", "/stickers/packs", map[string]any{"key": "chatonly", "name": map[string]string{"fa": "چت"}, "stickers": []map[string]any{{"key": "a", "imageUrl": "/u/a.png"}}})
	if got := items(h.mustAPI(200, "GET", "/stickers/packs?place=match", nil)); len(got) != 2 {
		t.Fatalf("match packs %v", got)
	}

	a, b := h.user("en"), h.user("fa")
	out := h.mustRPC(a, "base_comm_packs", map[string]any{})
	packs := out["packs"].([]any)
	if len(packs) != 2 || out["cooldown"].(map[string]any)["maxCount"] != float64(6) {
		t.Fatalf("packs %v", out)
	}
	hello := packs[0].(map[string]any)
	if hello["key"] != "hello" || hello["starter"] != true || hello["owned"] != true || hello["items"].([]any)[1].(map[string]any)["text"] != "Good game" {
		t.Fatalf("hello %v", hello)
	}
	faces := packs[1].(map[string]any)
	if faces["owned"] != false || faces["item"].(map[string]any)["key"] != "comm_vip_item" {
		t.Fatalf("faces %v", faces)
	}
	if mine := h.mustRPC(a, "base_comm_packs", map[string]any{"mine": true})["packs"].([]any); len(mine) != 1 {
		t.Fatalf("mine %v", mine)
	}
	if chat := h.mustRPC(a, "base_stickers", map[string]any{})["packs"].([]any); len(chat) != 2 || chat[0].(map[string]any)["key"] != "chatonly" || chat[1].(map[string]any)["kind"] != "emoji" {
		t.Fatalf("chat packs %v", chat)
	}

	if code := h.rpcCode(a, "base_comm_favorite", map[string]any{"pack": "faces", "item": "angry"}); code != "PACK_NOT_OWNED" {
		t.Fatal(code)
	}
	if code := h.rpcCode(a, "base_comm_favorite", map[string]any{"pack": "hello", "item": "nope"}); code != "ITEM_NOT_AVAILABLE" {
		t.Fatal(code)
	}
	h.mustRPC(a, "base_comm_favorite", map[string]any{"pack": "hello", "item": "gg"})
	h.mustRPC(a, "base_comm_favorite", map[string]any{"pack": "hello", "item": "gg"})
	favs := items(h.mustRPC(a, "base_comm_favorites", map[string]any{}))
	if len(favs) != 1 || favs[0]["ref"] != "hello/gg" || favs[0]["text"] != "Good game" {
		t.Fatalf("favorites %v", favs)
	}
	h.mustAPI(200, "PUT", "/comms/settings", map[string]any{"enabled": true, "minIntervalSeconds": 4, "maxCount": 6, "windowSeconds": 30, "lockSeconds": 30, "channelHours": 2, "maxFavorites": 1})
	if code := h.rpcCode(a, "base_comm_favorite", map[string]any{"pack": "hello", "item": "hi"}); code != "FAVORITES_FULL" {
		t.Fatal(code)
	}
	h.mustRPC(a, "base_comm_favorite", map[string]any{"pack": "hello", "item": "gg", "favorite": false})
	if favs := items(h.mustRPC(a, "base_comm_favorites", map[string]any{})); len(favs) != 0 {
		t.Fatalf("cleared %v", favs)
	}
	h.expectError(400, "INVALID_SETTINGS", "PUT", "/comms/settings", map[string]any{"enabled": true, "minIntervalSeconds": 5, "maxCount": 6, "windowSeconds": 30, "lockSeconds": 30, "channelHours": 2})
	h.mustAPI(200, "PUT", "/comms/settings", map[string]any{"enabled": true, "minIntervalSeconds": 0, "maxCount": 2, "windowSeconds": 30, "lockSeconds": 30, "channelHours": 2})

	match := uuid.Must(uuid.NewV4()).String() + ".node1"
	if err := commpacks.Open(h.ctx, h.s.db, match, []string{a, b}); err != nil {
		t.Fatal(err)
	}
	msg, err := commpacks.Send(h.ctx, h.s.db, match, a, commpacks.Input{Pack: "hello", Item: "gg", ClientMessageID: "c1"})
	if err != nil || msg.PackKind != "text" || msg.Text["fa"] != "بازی خوبی بود" || msg.SenderID != a {
		t.Fatalf("send %+v %v", msg, err)
	}
	again, err := commpacks.Send(h.ctx, h.s.db, match, a, commpacks.Input{Pack: "hello", Item: "gg", ClientMessageID: "c1"})
	if err != nil || again.MessageID != msg.MessageID {
		t.Fatalf("dedupe %+v %v", again, err)
	}
	if _, err := commpacks.Send(h.ctx, h.s.db, match, a, commpacks.Input{Pack: "faces", Item: "angry"}); refusal(t, err) != "PACK_NOT_OWNED" {
		t.Fatal(err)
	}
	if _, err := commpacks.Send(h.ctx, h.s.db, match, a, commpacks.Input{Pack: "chatonly", Item: "a"}); refusal(t, err) != "PACK_NOT_AVAILABLE" {
		t.Fatal(err)
	}
	if _, err := commpacks.Send(h.ctx, h.s.db, match, a, commpacks.Input{Pack: "hello", Item: "zz"}); refusal(t, err) != "ITEM_NOT_AVAILABLE" {
		t.Fatal(err)
	}
	h.giveItem(a, item)
	if m, err := commpacks.Send(h.ctx, h.s.db, match, a, commpacks.Input{Pack: "faces", Item: "angry"}); err != nil || m.AssetID != "emoji.angry" {
		t.Fatalf("owned emoji %+v %v", m, err)
	}
	if _, err := commpacks.Send(h.ctx, h.s.db, match, a, commpacks.Input{Pack: "hello", Item: "hi"}); refusal(t, err) != "SPAM_LOCKED" {
		t.Fatal(err)
	}

	ch := h.mustAPI(200, "GET", "/comms/channels/"+match, nil)
	if ch["status"] != "open" || ch["messages"] != float64(2) || len(ch["members"].([]any)) != 2 {
		t.Fatalf("channel %v", ch)
	}
	if list := items(h.mustAPI(200, "GET", "/comms/channels", nil)); len(list) != 1 {
		t.Fatalf("open list %v", list)
	}
	h.mustAPI(200, "POST", "/comms/channels/"+match+"/can-send", map[string]any{"userId": b, "canSend": false})
	if _, err := commpacks.Send(h.ctx, h.s.db, match, b, commpacks.Input{Pack: "hello", Item: "hi"}); refusal(t, err) != "SEND_NOT_ALLOWED" {
		t.Fatal(err)
	}
	h.expectError(404, "NOT_A_MEMBER", "POST", "/comms/channels/"+match+"/can-send", map[string]any{"userId": uuid.Must(uuid.NewV4()).String(), "canSend": true})
	h.mustAPI(200, "POST", "/comms/channels/"+match+"/can-send", map[string]any{"userId": b, "canSend": true})
	closed := h.mustAPI(200, "POST", "/comms/channels/"+match+"/close", nil)
	if closed["status"] != "closed" || closed["closeReason"] != "staff" {
		t.Fatalf("closed %v", closed)
	}
	if _, err := commpacks.Send(h.ctx, h.s.db, match, b, commpacks.Input{Pack: "hello", Item: "hi"}); refusal(t, err) != "CHANNEL_CLOSED" {
		t.Fatal(err)
	}
	h.mustAPI(200, "POST", "/comms/channels/"+match+"/open", nil)
	if _, err := commpacks.Send(h.ctx, h.s.db, match, b, commpacks.Input{Pack: "hello", Item: "hi"}); err != nil {
		t.Fatal(err)
	}
	if err := commpacks.Close(h.ctx, h.s.db, match, "ended"); err != nil {
		t.Fatal(err)
	}
	if list := items(h.mustAPI(200, "GET", "/comms/channels?status=closed", nil)); len(list) != 1 || list[0]["closeReason"] != "ended" {
		t.Fatalf("closed list %v", list)
	}
	h.mustAPI(200, "PUT", "/comms/settings", map[string]any{"enabled": false, "minIntervalSeconds": 4, "maxCount": 6, "windowSeconds": 30, "lockSeconds": 30, "channelHours": 2})
	if _, err := commpacks.Send(h.ctx, h.s.db, uuid.Must(uuid.NewV4()).String()+".n", a, commpacks.Input{Pack: "hello", Item: "hi"}); refusal(t, err) != "COMMS_DISABLED" {
		t.Fatal(err)
	}
	if out := h.mustRPC(a, "base_comm_packs", map[string]any{}); out["enabled"] != false || len(out["packs"].([]any)) != 0 {
		t.Fatalf("off %v", out)
	}
}

func TestChatStickerCarriesTextAndAsset(t *testing.T) {
	h := newHarness(t)
	resetComms(h)
	h.mustAPI(201, "POST", "/stickers/packs", map[string]any{"key": "words", "kind": "text", "inMatch": true, "name": map[string]string{"fa": "کلمه"}, "stickers": []map[string]any{{"key": "hi", "text": map[string]string{"fa": "سلام"}}}})
	h.mustAPI(201, "POST", "/stickers/packs", map[string]any{"key": "gameonly", "kind": "emoji", "inChat": false, "inMatch": true, "name": map[string]string{"fa": "بازی"}, "stickers": []map[string]any{{"key": "a", "assetId": "e.a"}}})
	a, b := h.user("fa"), h.user("fa")
	dm := dmChannel(a, b)
	out, e, err := h.send(a, dm, `{"sticker":"words/hi"}`)
	if err != nil || e != nil || !strings.Contains(out, `"stickerKind":"text"`) || !strings.Contains(out, `"stickerText":{"fa":"سلام"}`) {
		t.Fatalf("text sticker %q %v %v", out, e, err)
	}
	if _, e, _ := h.send(a, dm, `{"sticker":"gameonly/a"}`); e == nil || e.Code != CodeStickerNotFound {
		t.Fatalf("match-only pack in chat %v", e)
	}
}

func TestVoiceRoomsMuteThroughLiveKit(t *testing.T) {
	h, _ := voiceHarness(t)
	var calls []string
	lk := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		calls = append(calls, r.URL.Path+" "+string(raw))
		switch r.URL.Path {
		case "/twirp/livekit.RoomService/UpdateParticipant":
			_, _ = w.Write([]byte(`{}`))
		case "/twirp/livekit.RoomService/ListParticipants":
			_, _ = w.Write([]byte(`{"participants":[]}`))
		}
	}))
	defer lk.Close()
	t.Cleanup(func() {
		h.s.db.Exec(`DELETE FROM base_voice_room WHERE room LIKE 'room.vm%'`)
		h.s.db.Exec(`DELETE FROM base_voice_mute WHERE room LIKE 'room.vm%'`)
	})
	key := "vm" + strings.ReplaceAll(uuid.Must(uuid.NewV4()).String()[:8], "-", "")
	h.game(key, true)
	h.mustAPI(200, "PUT", "/voice/settings", map[string]any{"enabled": true, "provider": "livekit", "ttlSeconds": 600, "canPublish": true, "openRooms": true,
		"config": map[string]string{"url": lk.URL, "apiKey": "APIk"}, "secrets": map[string]string{"apiSecret": lkSecret}})
	a, b := h.user("fa"), h.user("fa")
	tok := h.mustRPC(a, "voice_token", map[string]any{"room": "lobby", "gameKey": key})
	if tok["canPublish"] != true || tok["staffMuted"] != false {
		t.Fatalf("token %v", tok)
	}
	room := tok["room"].(string)
	h.mustRPC(b, "voice_token", map[string]any{"room": "lobby", "gameKey": key})
	list := items(h.mustAPI(200, "GET", "/voice/rooms?q="+key, nil))
	if len(list) != 1 || list[0]["room"] != room || len(list[0]["members"].([]any)) != 2 || list[0]["kind"] != "room" {
		t.Fatalf("rooms %v", list)
	}
	h.expectError(404, "VOICE_ROOM_NOT_FOUND", "POST", "/voice/rooms/room.none.x/mute", map[string]any{"userId": a})
	h.expectError(404, "VOICE_PARTICIPANT_NOT_FOUND", "POST", "/voice/rooms/"+room+"/mute", map[string]any{"userId": uuid.Must(uuid.NewV4()).String()})
	res := h.mustAPI(200, "POST", "/voice/rooms/"+room+"/mute", map[string]any{"userId": a})
	if res["muted"] != true || res["applied"] != true || !strings.Contains(calls[len(calls)-1], `"canPublish":false`) || !strings.Contains(calls[len(calls)-1], a) {
		t.Fatalf("mute %v %v", res, calls)
	}
	again := h.mustRPC(a, "voice_token", map[string]any{"room": "lobby", "gameKey": key})
	if again["canPublish"] != false || again["staffMuted"] != true {
		t.Fatalf("muted token %v", again)
	}
	claims := jwtClaims(t, again["token"].(string), lkSecret)
	if claims["video"].(map[string]any)["canPublish"] != false {
		t.Fatalf("claims %v", claims)
	}
	members := items(h.mustAPI(200, "GET", "/voice/rooms?q="+key+"&live=true", nil))[0]["members"].([]any)
	if members[0].(map[string]any)["muted"] != true || members[0].(map[string]any)["connected"] != false {
		t.Fatalf("members %v", members)
	}
	h.mustAPI(200, "POST", "/voice/rooms/"+room+"/unmute", map[string]any{"userId": a})
	if !strings.Contains(calls[len(calls)-1], `"canPublish":true`) {
		t.Fatalf("unmute call %v", calls[len(calls)-1])
	}
	if tok := h.mustRPC(a, "voice_token", map[string]any{"room": "lobby", "gameKey": key}); tok["canPublish"] != true {
		t.Fatalf("unmuted token %v", tok)
	}
	h.s.db.Exec(`UPDATE base_voice_room SET expires_at=now() - INTERVAL '1 minute' WHERE room=$1`, room)
	if list := items(h.mustAPI(200, "GET", "/voice/rooms?q="+key, nil)); len(list) != 0 {
		t.Fatalf("expired room listed %v", list)
	}
	_ = time.Now
}
