package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

func (f *fixture) art(kind, key string, minLevel int, price int64, def bool) {
	f.t.Helper()
	cur := ""
	if price > 0 {
		cur = "coin"
	}
	plugintest.Exec(f.t, f.s.db, `INSERT INTO clan_art (id, kind, key, name, image_url, min_level, price_currency, price_amount, is_default) VALUES ($1,$2,$3,'{"en":"x"}',$4,$5,$6,$7,$8)`,
		uuid.Must(uuid.NewV4()).String(), kind, key, "https://img.test/"+key+".png", minLevel, cur, price, def)
	f.t.Cleanup(func() { _, _ = f.s.db.Exec(`DELETE FROM clan_art WHERE key=$1`, key) })
}

func (f *fixture) look(change func(*settings.All)) {
	f.t.Helper()
	cfg := f.s.settings(context.Background())
	change(&cfg)
	if err := settings.Save(context.Background(), f.s.db, cfg, "test"); err != nil {
		f.t.Fatal(err)
	}
}

func TestClanLookGalleryLocksPricesAndCooldown(t *testing.T) {
	f := newFixture(t)
	k := plugintest.Key("lk")
	f.art(artEmblem, k+"_def", 1, 0, true)
	f.art(artEmblem, k+"_gold", 2, 0, false)
	f.art(artCover, k+"_paid", 1, 100, false)
	f.look(func(c *settings.All) {
		c.Profile.Tags = []settings.Tag{{Key: "casual", Name: map[string]string{"en": "Casual"}}, {Key: "pro"}}
		c.Profile.MaxTags = 1
		c.Profile.BlockedWords = []string{"badword"}
		c.Profile.ChangeCooldownMinutes = 10
	})
	debits := []string{}
	f.s.debit = func(_ context.Context, _ string, grant string, amounts map[string]int64) error {
		if len(amounts) > 0 {
			debits = append(debits, grant)
		}
		return nil
	}
	leader := f.user()
	if _, err := f.call(f.s.rpcCreate, leader, map[string]any{"name": "Mybadword " + k}); code(err) != "CLAN_NAME_BLOCKED" {
		t.Fatalf("blocked word in the name: %v", err)
	}
	if _, err := f.call(f.s.rpcCreate, leader, map[string]any{"name": "Look " + k, "avatarUrl": "https://evil.test/a.png"}); code(err) != "CLAN_AVATAR_USE_GALLERY" {
		t.Fatalf("players cannot set any avatar url: %v", err)
	}
	out := f.must(f.s.rpcCreate, leader, map[string]any{"name": "Look " + k, "motto": "  با هم  می‌بریم ", "color": "#ff8800", "tags": []string{"casual"}})
	c := out["clan"].(map[string]any)
	f.clans = append(f.clans, c["clanId"].(string))
	if c["emblemKey"] != k+"_def" || c["emblemUrl"] != "https://img.test/"+k+"_def.png" || c["motto"] != "با هم می‌بریم" || c["color"] != "#FF8800" {
		t.Fatalf("default emblem and look on create: %v", c)
	}
	other := f.user()
	if _, err := f.call(f.s.rpcCreate, other, map[string]any{"name": "LOOK " + k}); code(err) != "CLAN_NAME_TAKEN" {
		t.Fatalf("names are unique without case: %v", err)
	}
	if _, err := f.call(f.s.rpcUpdate, leader, map[string]any{"emblemKey": k + "_gold"}); code(err) != "CLAN_ART_LOCKED" {
		t.Fatalf("level 2 emblem is locked at level 1: %v", err)
	}
	if _, err := f.call(f.s.rpcUpdate, leader, map[string]any{"tags": []string{"casual", "pro"}}); code(err) != "CLAN_TOO_MANY_TAGS" {
		t.Fatalf("tag cap: %v", err)
	}
	if _, err := f.call(f.s.rpcUpdate, leader, map[string]any{"tags": []string{"hardcore"}}); code(err) != "CLAN_TAG_INVALID" {
		t.Fatalf("only admin tags: %v", err)
	}
	up := f.must(f.s.rpcUpdate, leader, map[string]any{"coverKey": k + "_paid"})
	if up["clan"].(map[string]any)["coverUrl"] != "https://img.test/"+k+"_paid.png" || len(debits) != 1 {
		t.Fatalf("paid cover bought once: %v %v", up["clan"], debits)
	}
	if _, err := f.call(f.s.rpcUpdate, leader, map[string]any{"motto": "again"}); code(err) != "CLAN_CHANGE_COOLDOWN" {
		t.Fatalf("changes wait for the cooldown: %v", err)
	}
	plugintest.Exec(t, f.s.db, `UPDATE clan SET look_changed_at=now() - interval '1 hour', level=2 WHERE group_id::text=$1`, c["clanId"])
	f.must(f.s.rpcUpdate, leader, map[string]any{"coverKey": ""})
	plugintest.Exec(t, f.s.db, `UPDATE clan SET look_changed_at=NULL WHERE group_id::text=$1`, c["clanId"])
	f.must(f.s.rpcUpdate, leader, map[string]any{"coverKey": k + "_paid", "emblemKey": k + "_gold"})
	if len(debits) != 1 {
		t.Fatalf("an owned cover is not paid again: %v", debits)
	}
	var meta string
	_ = f.s.db.QueryRow(`SELECT metadata::text FROM groups WHERE id::text=$1`, c["clanId"]).Scan(&meta)
	var m map[string]any
	_ = json.Unmarshal([]byte(meta), &m)
	if m["cover"] != "https://img.test/"+k+"_paid.png" || m["clan"] != true {
		t.Fatalf("cover kept in group metadata: %v", m)
	}
	opts := f.must(f.s.rpcOptions, leader, nil)
	if len(opts["emblems"].([]any)) < 2 || len(opts["tags"].([]any)) != 2 {
		t.Fatalf("options: %v", opts)
	}
	member := f.user()
	f.must(f.s.rpcJoin, member, map[string]any{"clanId": c["clanId"]})
	plugintest.Exec(t, f.s.db, `UPDATE clan SET look_changed_at=NULL WHERE group_id::text=$1`, c["clanId"])
	if _, err := f.call(f.s.rpcUpdate, member, map[string]any{"motto": "x"}); code(err) != "CLAN_ROLE_NOT_ALLOWED" {
		t.Fatalf("members cannot change the look: %v", err)
	}
}

func TestClanImageUploadWaitsForApproval(t *testing.T) {
	f := newFixture(t)
	dir := t.TempDir()
	store, err := objectstore.New(objectstore.Config{Driver: objectstore.DriverLocal, LocalDir: dir}, nil)
	if err != nil {
		t.Fatal(err)
	}
	f.s.store, f.s.publicBase = store, "http://api.test"
	leader := f.user()
	id := f.create(leader, "Up "+plugintest.Key("u"), settings.JoinOpen)
	var buf bytes.Buffer
	_ = png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 8, 8)))
	data := base64.StdEncoding.EncodeToString(buf.Bytes())
	if _, err := f.call(f.s.rpcImageUpload, leader, map[string]any{"kind": "emblem", "data": data}); code(err) != "CLAN_UPLOADS_DISABLED" {
		t.Fatalf("uploads are off by default: %v", err)
	}
	f.look(func(c *settings.All) { c.Uploads.Enabled, c.Uploads.MaxBytes = true, 4096 })
	big := base64.StdEncoding.EncodeToString(append(buf.Bytes(), make([]byte, 5000)...))
	if _, err := f.call(f.s.rpcImageUpload, leader, map[string]any{"kind": "emblem", "data": big}); code(err) != "CLAN_UPLOAD_TOO_LARGE" {
		t.Fatalf("size limit: %v", err)
	}
	if _, err := f.call(f.s.rpcImageUpload, leader, map[string]any{"kind": "emblem", "data": base64.StdEncoding.EncodeToString([]byte("GIF89a not an allowed image"))}); code(err) != "CLAN_UPLOAD_TYPE" {
		t.Fatalf("type limit: %v", err)
	}
	out := f.must(f.s.rpcImageUpload, leader, map[string]any{"kind": "emblem", "data": data})
	if out["status"] != "pending" || out["url"] != nil {
		t.Fatalf("pending until approved: %v", out)
	}
	var file string
	_ = f.s.db.QueryRow(`SELECT file_name FROM clan_upload WHERE id::text=$1`, out["uploadId"]).Scan(&file)
	r := httptest.NewRequest(http.MethodGet, "/v2/clan/files/"+file, nil)
	r.SetPathValue("name", file)
	w := httptest.NewRecorder()
	f.s.filesHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("a pending file is not public: %d", w.Code)
	}
	code2, _ := f.consoleCall(http.MethodPost, f.s.uploadApprove, map[string]string{"id": out["uploadId"].(string)}, map[string]any{})
	if code2 != http.StatusOK {
		t.Fatalf("approve %d", code2)
	}
	w = httptest.NewRecorder()
	f.s.filesHTTP(w, r)
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("approved file is public: %d %s", w.Code, w.Header().Get("Content-Type"))
	}
	var avatar string
	_ = f.s.db.QueryRow(`SELECT avatar_url FROM groups WHERE id::text=$1`, id).Scan(&avatar)
	if avatar != "http://api.test/v2/clan/files/"+file {
		t.Fatalf("approved upload becomes the clan avatar: %s", avatar)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal(err)
	}
	second := f.must(f.s.rpcImageUpload, f.s.withRoles(context.Background(), id, []string{"leader"})[0], map[string]any{"kind": "cover", "data": data})
	if code3, _ := f.consoleCall(http.MethodPost, f.s.uploadReject, map[string]string{"id": second["uploadId"].(string)}, map[string]any{"reason": "no"}); code3 != http.StatusOK {
		t.Fatal("reject")
	}
	var st string
	_ = f.s.db.QueryRow(`SELECT status FROM clan_upload WHERE id::text=$1`, second["uploadId"]).Scan(&st)
	if st != "rejected" {
		t.Fatal(st)
	}
}

func (f *fixture) consoleCall(method string, h http.HandlerFunc, vars map[string]string, body any) (int, map[string]any) {
	raw, _ := json.Marshal(body)
	r := httptest.NewRequest(method, "/x", bytes.NewReader(raw))
	for k, v := range vars {
		r.SetPathValue(k, v)
	}
	w := httptest.NewRecorder()
	h(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}
