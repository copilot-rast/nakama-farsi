package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
)

func resetStickers(h *harness) {
	h.t.Helper()
	for _, q := range []string{`DELETE FROM base_sticker_pack`, `DELETE FROM base_sticker_settings`, `DELETE FROM base_chat_quota_settings`} {
		if _, err := h.s.db.Exec(q); err != nil {
			h.t.Fatal(err)
		}
	}
	h.t.Cleanup(func() {
		_, _ = h.s.db.Exec(`DELETE FROM base_sticker_pack`)
		_, _ = h.s.db.Exec(`DELETE FROM base_sticker_settings`)
		_, _ = h.s.db.Exec(`DELETE FROM base_chat_quota_settings`)
	})
}

func (h *harness) stickerItem(key string) string {
	h.t.Helper()
	typeID, itemID := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.Exec(`INSERT INTO base_category (id, key_type, key, name) VALUES ($1,'item_type',$2,'{"en":"Stickers"}'::jsonb)`, typeID, key+"_type"); err != nil {
		h.t.Fatal(err)
	}
	if _, err := h.s.db.Exec(`INSERT INTO economy_item (id, key, type_id, name) VALUES ($1,$2,$3,'{"fa":"بستهٔ استیکر","en":"Sticker pack"}'::jsonb)`, itemID, key, typeID); err != nil {
		h.t.Fatal(err)
	}
	h.t.Cleanup(func() { _, _ = h.s.db.Exec(`DELETE FROM economy_item WHERE id=$1`, itemID) })
	return itemID
}

func (h *harness) giveItem(user, itemID string) {
	h.t.Helper()
	if _, err := h.s.db.Exec(`INSERT INTO economy_user_item (user_id, item_id, source) VALUES ($1,$2,'admin')`, user, itemID); err != nil {
		h.t.Fatal(err)
	}
}

func (h *harness) send(user, channel, content string) (string, *stickerError, error) {
	h.t.Helper()
	ctx := asUser(user)
	out, isSticker, st, err := h.s.stickerGate(ctx, user, channel, content)
	if err == nil && (!isSticker || st.CountQuota) {
		err = h.s.chatQuotaGate(ctx, user, channel, out)
	}
	if err == nil {
		return out, nil, nil
	}
	var rt *runtime.Error
	if errors.As(err, &rt) {
		e := &stickerError{}
		if json.Unmarshal([]byte(rt.Message), e) == nil && e.Code != "" {
			return "", e, nil
		}
	}
	return "", nil, err
}

func TestStickerPacksConsoleAndRPC(t *testing.T) {
	h := newHarness(t)
	resetStickers(h)
	h.stickerItem("e2e_sticker_item")

	h.expectError(400, "INVALID_STICKERS", "POST", "/stickers/packs", map[string]any{"key": "happy", "name": map[string]string{"fa": "شاد"}})
	h.expectError(400, "IMAGE_REQUIRED", "POST", "/stickers/packs", map[string]any{"key": "happy", "name": map[string]string{"fa": "شاد"}, "stickers": []map[string]any{{"key": "a"}}})
	h.expectError(400, "ITEM_REQUIRED", "POST", "/stickers/packs", map[string]any{"key": "vip", "name": map[string]string{"fa": "ویژه"}, "access": "item", "stickers": []map[string]any{{"key": "a", "imageUrl": "/u/a.png"}}})
	h.expectError(400, "ITEM_NOT_FOUND", "POST", "/stickers/packs", map[string]any{"key": "vip", "name": map[string]string{"fa": "ویژه"}, "access": "item", "itemKey": "nope", "stickers": []map[string]any{{"key": "a", "imageUrl": "/u/a.png"}}})
	h.expectError(400, "DUPLICATE_KEY", "POST", "/stickers/packs", map[string]any{"key": "happy", "name": map[string]string{"fa": "شاد"}, "stickers": []map[string]any{{"key": "a", "imageUrl": "/u/a.png"}, {"key": "a", "imageUrl": "/u/b.png"}}})

	free := h.mustAPI(201, "POST", "/stickers/packs", map[string]any{"key": "happy", "name": map[string]string{"fa": "شاد", "en": "Happy"}, "sort": 1,
		"stickers": []map[string]any{{"key": "smile", "name": map[string]string{"fa": "لبخند"}, "imageUrl": "/u/smile.png"}, {"key": "laugh", "imageUrl": "https://cdn.example.com/laugh.png"}}})
	if free["count"] != float64(2) || free["access"] != "free" {
		t.Fatalf("free pack: %v", free)
	}
	h.expectError(409, "KEY_EXISTS", "POST", "/stickers/packs", map[string]any{"key": "happy", "name": map[string]string{"fa": "شاد"}, "stickers": []map[string]any{{"key": "a", "imageUrl": "/u/a.png"}}})
	vip := h.mustAPI(201, "POST", "/stickers/packs", map[string]any{"key": "vip", "name": map[string]string{"fa": "ویژه"}, "access": "item", "itemKey": "e2e_sticker_item", "sort": 2,
		"stickers": []map[string]any{{"key": "crown", "imageUrl": "/u/crown.png"}}})
	if vip["item"] == nil {
		t.Fatalf("vip pack has no item: %v", vip)
	}
	list := h.mustAPI(200, "GET", "/stickers/packs?access=item", nil)
	if list["total"] != float64(1) {
		t.Fatalf("filter by access: %v", list)
	}

	edited := h.mustAPI(200, "PUT", "/stickers/packs/"+free["id"].(string), map[string]any{"key": "ignored", "name": map[string]string{"fa": "شاد"}, "stickers": []map[string]any{{"key": "laugh", "imageUrl": "/u/laugh2.png"}}})
	if edited["key"] != "happy" || edited["count"] != float64(1) {
		t.Fatalf("edit keeps the key and replaces stickers: %v", edited)
	}

	u := h.user("fa")
	out := h.mustRPC(u, "base_stickers", map[string]any{})
	packs := out["packs"].([]any)
	if len(packs) != 2 || out["enabled"] != true {
		t.Fatalf("rpc: %v", out)
	}
	p0, p1 := packs[0].(map[string]any), packs[1].(map[string]any)
	if p0["key"] != "happy" || p0["owned"] != true || p0["name"] != "شاد" || p1["owned"] != false || p1["itemKey"] != "e2e_sticker_item" {
		t.Fatalf("packs: %v %v", p0, p1)
	}
	if s := p0["stickers"].([]any)[0].(map[string]any); s["ref"] != "happy/laugh" || s["imageUrl"] != "/u/laugh2.png" {
		t.Fatalf("sticker: %v", s)
	}

	h.mustAPI(200, "PUT", "/stickers/settings", map[string]any{"enabled": false, "direct": true, "groups": true, "rooms": true, "countQuota": true})
	if out := h.mustRPC(u, "base_stickers", map[string]any{}); out["enabled"] != false || len(out["packs"].([]any)) != 0 {
		t.Fatalf("disabled: %v", out)
	}
	h.mustAPI(200, "DELETE", "/stickers/packs/"+vip["id"].(string), nil)
	h.expectError(404, "NOT_FOUND", "GET", "/stickers/packs/"+vip["id"].(string), nil)
}

func TestStickerGate(t *testing.T) {
	h := newHarness(t)
	resetStickers(h)
	item := h.stickerItem("e2e_sticker_item2")
	h.mustAPI(201, "POST", "/stickers/packs", map[string]any{"key": "happy", "name": map[string]string{"fa": "شاد"}, "stickers": []map[string]any{{"key": "smile", "imageUrl": "/u/smile.png"}, {"key": "off", "imageUrl": "/u/off.png", "isActive": false}}})
	h.mustAPI(201, "POST", "/stickers/packs", map[string]any{"key": "vip", "name": map[string]string{"fa": "ویژه"}, "access": "item", "itemKey": "e2e_sticker_item2", "stickers": []map[string]any{{"key": "crown", "imageUrl": "/u/crown.png"}}})
	a, b := h.user("fa"), h.user("en")
	dm := dmChannel(a, b)

	out, e, err := h.send(a, dm, `{"text":"hi"}`)
	if err != nil || e != nil || out != `{"text":"hi"}` {
		t.Fatalf("text passes untouched: %q %v %v", out, e, err)
	}
	out, e, err = h.send(a, dm, `{"sticker":"happy/smile"}`)
	if err != nil || e != nil {
		t.Fatalf("free sticker: %v %v", e, err)
	}
	var doc map[string]any
	_ = json.Unmarshal([]byte(out), &doc)
	if doc["stickerUrl"] != "/u/smile.png" || doc["stickerPack"] != "happy" || doc["sticker"] != "happy/smile" {
		t.Fatalf("enriched: %v", doc)
	}
	for _, c := range []struct{ content, code string }{
		{`{"sticker":"happy/off"}`, CodeStickerNotFound},
		{`{"sticker":"happy/none"}`, CodeStickerNotFound},
		{`{"sticker":"bad"}`, CodeStickerNotFound},
		{`{"sticker":"vip/crown"}`, CodeStickerLocked},
	} {
		if _, e, err := h.send(a, dm, c.content); err != nil || e == nil || e.Code != c.code {
			t.Fatalf("%s: want %s, got %v %v", c.content, c.code, e, err)
		}
	}
	_, e, _ = h.send(a, dm, `{"sticker":"vip/crown"}`)
	if e.Lang != "fa" || !strings.Contains(e.Message, "استیکر") || e.Meta["itemKey"] != "e2e_sticker_item2" {
		t.Fatalf("locked refusal: %+v", e)
	}
	h.giveItem(a, item)
	if _, e, err := h.send(a, dm, `{"sticker":"vip/crown"}`); err != nil || e != nil {
		t.Fatalf("owned pack: %v %v", e, err)
	}

	h.mustAPI(200, "PUT", "/stickers/settings", map[string]any{"enabled": true, "direct": false, "groups": true, "rooms": true, "countQuota": true})
	if _, e, _ := h.send(a, dm, `{"sticker":"happy/smile"}`); e == nil || e.Code != CodeStickerNotAllowedHere || e.Meta["chat"] != "dm" {
		t.Fatalf("direct off: %+v", e)
	}
	if _, e, _ := h.send(a, "2...lobby", `{"sticker":"happy/smile"}`); e != nil {
		t.Fatalf("room still allowed: %+v", e)
	}

	h.mustAPI(200, "PUT", "/stickers/settings", map[string]any{"enabled": true, "direct": true, "groups": true, "rooms": true, "countQuota": false})
	h.chatSettings(map[string]any{"dailyMessages": 1, "countMode": "messages"})
	if _, e, err := h.send(b, dm, `{"text":"one"}`); e != nil || err != nil {
		t.Fatalf("first text: %v %v", e, err)
	}
	if _, e, err := h.send(b, dm, `{"text":"two"}`); err == nil && e == nil {
		t.Fatal("second text should hit the quota")
	}
	if _, e, err := h.send(b, dm, `{"sticker":"happy/smile"}`); e != nil || err != nil {
		t.Fatalf("sticker outside the quota: %v %v", e, err)
	}
	h.mustAPI(200, "PUT", "/stickers/settings", map[string]any{"enabled": true, "direct": true, "groups": true, "rooms": true, "countQuota": true})
	if _, e, err := h.send(b, dm, `{"sticker":"happy/smile"}`); err == nil && e == nil {
		t.Fatal("sticker should cost quota when countQuota is on")
	}

	h.chatSettings(map[string]any{"dailyMessages": 1, "countMode": "messages", "countDirect": false, "countRooms": true})
	if _, e, err := h.send(b, dm, `{"text":"free now"}`); e != nil || err != nil {
		t.Fatalf("direct not counted: %v %v", e, err)
	}
	if st := h.mustRPC(b, "base_chat_quota", map[string]any{}); st["counts"].(map[string]any)["direct"] != false {
		t.Fatalf("status counts: %v", st["counts"])
	}
}
