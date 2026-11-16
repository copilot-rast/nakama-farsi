package main

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/heroiclabs/nakama-common/runtime"
)

func resetAvatar(h *harness) {
	h.t.Helper()
	for _, q := range []string{`DELETE FROM base_avatar_part`, `DELETE FROM base_avatar_settings`, `DELETE FROM base_avatar_look`} {
		if _, err := h.s.db.Exec(q); err != nil {
			h.t.Fatal(err)
		}
	}
	h.t.Cleanup(func() {
		_, _ = h.s.db.Exec(`DELETE FROM base_avatar_part`)
		_, _ = h.s.db.Exec(`DELETE FROM base_avatar_settings`)
	})
}

func avatarCode(t *testing.T, err error) (string, map[string]any) {
	t.Helper()
	var rt *runtime.Error
	if !errors.As(err, &rt) {
		t.Fatalf("not a runtime error: %v", err)
	}
	var body struct {
		Code string         `json:"code"`
		Meta map[string]any `json:"meta"`
	}
	if json.Unmarshal([]byte(rt.Message), &body) != nil {
		return rt.Message, nil
	}
	return body.Code, body.Meta
}

func TestAvatarPartsConsole(t *testing.T) {
	h := newHarness(t)
	resetAvatar(h)
	h.stickerItem("e2e_avatar_hat_item")
	opt := func(key string, extra map[string]any) map[string]any {
		o := map[string]any{"key": key, "imageUrl": "/a/" + key + ".png"}
		for k, v := range extra {
			o[k] = v
		}
		return o
	}
	h.expectError(400, "INVALID_OPTIONS", "POST", "/avatar/parts", map[string]any{"key": "hat", "name": map[string]string{"fa": "کلاه"}})
	h.expectError(400, "OPTION_LOOK_REQUIRED", "POST", "/avatar/parts", map[string]any{"key": "hat", "name": map[string]string{"fa": "کلاه"}, "options": []map[string]any{{"key": "a"}}})
	h.expectError(400, "DEFAULT_MUST_BE_FREE", "POST", "/avatar/parts", map[string]any{"key": "hat", "name": map[string]string{"fa": "کلاه"}, "options": []map[string]any{opt("cap", map[string]any{"access": "item", "itemKey": "e2e_avatar_hat_item", "isDefault": true})}})
	h.expectError(400, "REQUIRED_NEEDS_FREE", "POST", "/avatar/parts", map[string]any{"key": "hat", "name": map[string]string{"fa": "کلاه"}, "required": true, "options": []map[string]any{opt("cap", map[string]any{"access": "item", "itemKey": "e2e_avatar_hat_item"})}})
	h.expectError(400, "ONE_DEFAULT", "POST", "/avatar/parts", map[string]any{"key": "hat", "name": map[string]string{"fa": "کلاه"}, "options": []map[string]any{opt("a", map[string]any{"isDefault": true}), opt("b", map[string]any{"isDefault": true})}})
	h.expectError(400, "ITEM_NOT_FOUND", "POST", "/avatar/parts", map[string]any{"key": "hat", "name": map[string]string{"fa": "کلاه"}, "options": []map[string]any{opt("cap", map[string]any{"access": "item", "itemKey": "nope"})}})

	skin := h.mustAPI(201, "POST", "/avatar/parts", map[string]any{"key": "skin", "name": map[string]string{"fa": "رنگ پوست"}, "layer": 0, "required": true,
		"options": []map[string]any{{"key": "light", "value": "#F1C27D", "isDefault": true}, {"key": "dark", "value": "#8D5524"}}})
	if skin["count"] != float64(2) || skin["required"] != true {
		t.Fatalf("skin: %v", skin)
	}
	h.expectError(409, "KEY_EXISTS", "POST", "/avatar/parts", map[string]any{"key": "skin", "name": map[string]string{"fa": "پوست"}, "options": []map[string]any{{"key": "x", "value": "#000"}}})
	hat := h.mustAPI(201, "POST", "/avatar/parts", map[string]any{"key": "hat", "name": map[string]string{"fa": "کلاه"}, "layer": 5,
		"options": []map[string]any{opt("cap", nil), opt("crown", map[string]any{"access": "item", "itemKey": "e2e_avatar_hat_item"})}})
	edited := h.mustAPI(200, "PUT", "/avatar/parts/"+hat["id"].(string), map[string]any{"key": "other", "name": map[string]string{"fa": "کلاه"}, "layer": 5,
		"options": []map[string]any{opt("crown", map[string]any{"access": "item", "itemKey": "e2e_avatar_hat_item"})}})
	if edited["key"] != "hat" || edited["count"] != float64(1) {
		t.Fatalf("edit keeps key and replaces options: %v", edited)
	}
	if l := h.mustAPI(200, "GET", "/avatar/parts", nil); l["total"] != float64(2) {
		t.Fatalf("list: %v", l)
	}
	h.mustAPI(200, "DELETE", "/avatar/parts/"+hat["id"].(string), nil)
	h.expectError(404, "NOT_FOUND", "GET", "/avatar/parts/"+hat["id"].(string), nil)
}

func TestAvatarLookRPCs(t *testing.T) {
	h := newHarness(t)
	resetAvatar(h)
	item := h.stickerItem("e2e_avatar_crown")
	h.mustAPI(201, "POST", "/avatar/parts", map[string]any{"key": "skin", "name": map[string]string{"fa": "رنگ پوست", "en": "Skin"}, "layer": 0, "required": true,
		"options": []map[string]any{{"key": "light", "value": "#F1C27D", "isDefault": true}, {"key": "dark", "value": "#8D5524"}}})
	h.mustAPI(201, "POST", "/avatar/parts", map[string]any{"key": "hat", "name": map[string]string{"fa": "کلاه"}, "layer": 5,
		"options": []map[string]any{{"key": "cap", "imageUrl": "/a/cap.png"}, {"key": "crown", "imageUrl": "/a/crown.png", "access": "item", "itemKey": "e2e_avatar_crown"}}})
	h.mustAPI(201, "POST", "/avatar/parts", map[string]any{"key": "eyes", "name": map[string]string{"fa": "چشم"}, "layer": 2, "required": true,
		"options": []map[string]any{{"key": "round", "imageUrl": "/a/round.png"}}})
	a, b := h.user("fa"), h.user("en")

	out := h.mustRPC(a, "base_avatar", map[string]any{})
	parts := out["parts"].([]any)
	if len(parts) != 3 || out["enabled"] != true {
		t.Fatalf("parts: %v", out)
	}
	hat := parts[1].(map[string]any)
	if hat["key"] != "hat" {
		hat = parts[2].(map[string]any)
	}
	crown := hat["options"].([]any)[1].(map[string]any)
	if crown["owned"] != false || crown["itemKey"] != "e2e_avatar_crown" || crown["item"] == nil {
		t.Fatalf("crown: %v", crown)
	}
	look := out["look"].(map[string]any)
	if p := look["parts"].(map[string]any); p["skin"] != "light" || p["eyes"] != "round" || p["hat"] != nil {
		t.Fatalf("default look: %v", look)
	}
	layers := look["layers"].([]any)
	if len(layers) != 2 || layers[0].(map[string]any)["part"] != "skin" || layers[1].(map[string]any)["part"] != "eyes" {
		t.Fatalf("layers in order: %v", layers)
	}

	for _, c := range []struct {
		look map[string]string
		code string
	}{
		{map[string]string{"wings": "x"}, "AVATAR_PART_NOT_FOUND"},
		{map[string]string{"hat": "nope"}, "AVATAR_OPTION_NOT_FOUND"},
		{map[string]string{"hat": "crown"}, "AVATAR_OPTION_LOCKED"},
		{map[string]string{"skin": ""}, "AVATAR_PART_REQUIRED"},
	} {
		_, err := h.rpc(a, "base_avatar_save", map[string]any{"look": c.look})
		if code, meta := avatarCode(t, err); code != c.code || meta == nil {
			t.Fatalf("%v: want %s, got %s %v", c.look, c.code, code, meta)
		}
	}
	saved := h.mustRPC(a, "base_avatar_save", map[string]any{"look": map[string]string{"skin": "dark", "hat": "cap"}})
	if p := saved["look"].(map[string]any)["parts"].(map[string]any); p["skin"] != "dark" || p["hat"] != "cap" || p["eyes"] != "round" {
		t.Fatalf("saved: %v", saved)
	}
	h.giveItem(a, item)
	h.mustRPC(a, "base_avatar_save", map[string]any{"look": map[string]string{"skin": "dark", "hat": "crown"}})

	got := h.mustRPC(b, "base_avatar_get", map[string]any{"userIds": []string{a, b, "bot-1"}})
	looks := got["looks"].(map[string]any)
	if p := looks[a].(map[string]any)["parts"].(map[string]any); p["hat"] != "crown" || p["skin"] != "dark" {
		t.Fatalf("other player's look: %v", looks[a])
	}
	if p := looks["bot-1"].(map[string]any)["parts"].(map[string]any); p["skin"] != "light" || p["hat"] != nil {
		t.Fatalf("bot gets the default look: %v", looks["bot-1"])
	}
	if _, err := h.s.db.Exec(`DELETE FROM economy_user_item WHERE user_id=$1`, a); err != nil {
		t.Fatal(err)
	}
	if p := h.mustRPC(b, "base_avatar_get", map[string]any{"userIds": []string{a}})["looks"].(map[string]any)[a].(map[string]any)["parts"].(map[string]any); p["hat"] != nil {
		t.Fatalf("a lost item falls back: %v", p)
	}
	h.mustRPC(a, "base_avatar_save", map[string]any{"look": map[string]string{"hat": ""}})
	if card := h.mustAPI(200, "GET", "/players/"+a+"/avatar", nil); card["saved"] != true {
		t.Fatalf("player card: %v", card)
	}
	if code := h.rpcCode(b, "base_avatar_get", map[string]any{"userIds": []string{}}); code != "INVALID_USER_IDS" {
		t.Fatalf("empty ids: %s", code)
	}

	h.mustAPI(200, "PUT", "/avatar/settings", map[string]any{"enabled": false})
	if out := h.mustRPC(a, "base_avatar", map[string]any{}); out["enabled"] != false || len(out["parts"].([]any)) != 0 {
		t.Fatalf("disabled: %v", out)
	}
	_, err := h.rpc(a, "base_avatar_save", map[string]any{"look": map[string]string{}})
	if code, _ := avatarCode(t, err); code != "AVATAR_DISABLED" {
		t.Fatalf("disabled save: %s", code)
	}
}
