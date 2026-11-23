package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama/v3/integrations/content/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

type fixture struct {
	t      *testing.T
	s      *service
	prefix string
}

func newFixture(t *testing.T) *fixture {
	db := plugintest.OpenDB(t)
	store, err := objectstore.New(objectstore.Config{Driver: objectstore.DriverLocal, LocalDir: t.TempDir()}, nil)
	if err != nil {
		t.Fatal(err)
	}
	f := &fixture{t: t, s: newService(db, plugintest.Logger{}, store, "http://cdn.test"), prefix: plugintest.Key("ct")}
	unused := `DELETE FROM content_blob b WHERE NOT EXISTS (SELECT 1 FROM content_pack_file f WHERE f.sha256=b.sha256)`
	_, _ = db.Exec(unused)
	t.Cleanup(func() {
		f.s.waitAssembly()
		_, _ = db.Exec(`DELETE FROM content_pack WHERE version LIKE $1`, f.prefix+"%")
		_, _ = db.Exec(`DELETE FROM content_pack_event WHERE version LIKE $1`, f.prefix+"%")
		_, _ = db.Exec(unused)
		_, _ = db.Exec(`DELETE FROM content_settings WHERE id=1`)
		settings.Invalidate()
	})
	f.settings(func(*settings.All) {})
	return f
}

func (f *fixture) settings(change func(*settings.All)) {
	f.t.Helper()
	cfg := settings.Defaults()
	change(&cfg)
	if err := settings.Save(context.Background(), f.s.db, cfg, "test"); err != nil {
		f.t.Fatal(err)
	}
}

func errCode(err error) string {
	if e, ok := apierr.As(err); ok {
		return e.Code
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

type upload struct {
	files map[string][]byte
}

func (u upload) list() []packFileIn {
	out := []packFileIn{}
	for p, b := range u.files {
		sum := sha256.Sum256(b)
		out = append(out, packFileIn{Path: p, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(b))})
	}
	return out
}

func (f *fixture) uploadPack(platform, version, min, max string, u upload, activate bool) pack {
	f.t.Helper()
	return f.uploadKind("", platform, version, min, max, u, activate)
}

func (f *fixture) uploadKind(kind, platform, version, min, max string, u upload, activate bool) pack {
	f.t.Helper()
	ctx := context.Background()
	out, err := f.s.createPack(ctx, packCreateIn{Kind: kind, Platform: platform, Version: version, MinAppVersion: min, MaxAppVersion: max, Files: u.list()}, "t")
	if err != nil {
		f.t.Fatal(err)
	}
	p := out["pack"].(pack)
	for _, m := range out["missing"].([]missingFile) {
		var data []byte
		for _, x := range u.list() {
			if x.SHA256 == m.SHA256 {
				data = u.files[x.Path]
			}
		}
		for n := m.Parts - 1; n >= 0; n-- {
			end := (n + 1) * m.PartBytes
			if end > len(data) {
				end = len(data)
			}
			if _, err := f.s.putPart(ctx, p.ID, m.SHA256, n, bytes.NewReader(data[n*m.PartBytes:end])); err != nil {
				f.t.Fatal(err)
			}
		}
	}
	f.s.waitAssembly()
	p, err = f.s.completePack(ctx, p.ID, activate, "t")
	if err != nil {
		f.t.Fatal(err)
	}
	return p
}

func TestPackUploadActivationAndIndex(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 64 * 1024 })
	bundle := bytes.Repeat([]byte("bundle-data-"), 20000)
	u := upload{files: map[string][]byte{"Android/catalog_1.json": []byte(`{"c":1}`), "Android/catalog_1.hash": []byte("h1"), "Android/x.bundle": bundle}}
	p1 := f.uploadPack("Android", f.prefix+"-1", "1.0", "1.9", u, true)
	if p1.Status != "ready" || !p1.Active || p1.CatalogPath != "catalog_1.json" || p1.CatalogHash != "h1" || p1.FileCount != 3 {
		t.Fatalf("p1 = %+v", p1)
	}
	u2 := upload{files: map[string][]byte{"catalog_2.json": []byte(`{"c":2}`), "catalog_2.hash": []byte("h2"), "x.bundle": bundle}}
	out, err := f.s.createPack(ctx, packCreateIn{Platform: "Android", Version: f.prefix + "-2", MinAppVersion: "1.5", Files: u2.list()}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if m := out["missing"].([]missingFile); len(m) != 2 {
		t.Fatalf("missing = %+v", m)
	}
	if _, err := f.s.completePack(ctx, out["pack"].(pack).ID, false, "t"); errCode(err) != "PACK_FILES_MISSING" {
		t.Fatalf("complete early = %v", err)
	}
	if _, err := f.s.createPack(ctx, packCreateIn{Platform: "Android", Version: f.prefix + "-1", Files: u.list()}, "t"); errCode(err) != "PACK_VERSION_EXISTS" {
		t.Fatalf("same version = %v", err)
	}
	p2 := f.uploadPack("Android", f.prefix+"-2", "1.5", "", u2, false)
	idx := func(v string) string {
		p, err := f.s.latestPack(ctx, "Android", v)
		if err != nil {
			t.Fatal(err)
		}
		if p == nil {
			return ""
		}
		return p.Version
	}
	if got := idx("1.6"); got != f.prefix+"-1" {
		t.Fatalf("before activation = %q", got)
	}
	if _, err := f.s.activate(ctx, p2.ID, true, "admin"); err != nil {
		t.Fatal(err)
	}
	if idx("1.6") != f.prefix+"-2" || idx("1.2") != f.prefix+"-1" || idx("0.9") != "" || idx("2.10") != f.prefix+"-2" {
		t.Fatalf("ranges: 1.6=%q 1.2=%q 0.9=%q 2.10=%q", idx("1.6"), idx("1.2"), idx("0.9"), idx("2.10"))
	}
	if _, err := f.s.activate(ctx, p1.ID, true, "admin"); err != nil {
		t.Fatal(err)
	}
	if idx("1.6") != f.prefix+"-1" {
		t.Fatalf("rollback = %q", idx("1.6"))
	}
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 64 * 1024; c.Packs.SingleActive = true })
	if _, err := f.s.activate(ctx, p2.ID, true, "admin"); err != nil {
		t.Fatal(err)
	}
	cur, _ := f.s.loadPack(ctx, f.s.db, p1.ID, false)
	if cur.Active {
		t.Fatal("overlapping pack still active")
	}
	if err := f.s.deletePack(ctx, p2.ID, "admin"); errCode(err) != "PACK_ACTIVE" {
		t.Fatalf("delete active = %v", err)
	}
	files, err := f.s.packFiles(ctx, p2.ID)
	if err != nil || files["x.bundle"].size != int64(len(bundle)) {
		t.Fatalf("files = %+v %v", files, err)
	}
	obj, err := f.s.store.Get(ctx, blobKey(files["x.bundle"].sha))
	if err != nil {
		t.Fatal(err)
	}
	defer obj.Body.Close()
	view := indexView(&p2, "http://cdn.test").(map[string]any)
	if view["catalogUrl"] != "http://cdn.test/v2/content/files/"+p2.ID+"/catalog_2.json" {
		t.Fatalf("view = %+v", view)
	}
}

func TestPackPartChecksAndRetention(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 64 * 1024; c.Packs.KeepInactive = 1 })
	data := bytes.Repeat([]byte("z"), 80*1024)
	u := upload{files: map[string][]byte{"catalog.json": []byte(`{}`), "b.bundle": data}}
	out, err := f.s.createPack(ctx, packCreateIn{Platform: "iOS", Version: f.prefix + "-a", Files: u.list()}, "t")
	if err != nil {
		t.Fatal(err)
	}
	id := out["pack"].(pack).ID
	var bsha string
	for _, m := range out["missing"].([]missingFile) {
		if m.Path == "b.bundle" {
			bsha = m.SHA256
		}
	}
	if _, err := f.s.putPart(ctx, id, bsha, 0, bytes.NewReader(data[:100])); errCode(err) != "PACK_PART_INVALID" {
		t.Fatalf("short part = %v", err)
	}
	if _, err := f.s.putPart(ctx, id, bsha, 5, bytes.NewReader(data[:100])); errCode(err) != "PACK_PART_INVALID" {
		t.Fatalf("part index = %v", err)
	}
	wrong := bytes.Repeat([]byte("y"), 64*1024)
	if _, err := f.s.putPart(ctx, id, bsha, 0, bytes.NewReader(wrong)); err != nil {
		t.Fatal(err)
	}
	res, err := f.s.putPart(ctx, id, bsha, 1, bytes.NewReader(data[64*1024:]))
	if err != nil || res["state"] != StateAssembling {
		t.Fatalf("last part = %v %v", res, err)
	}
	f.s.waitAssembly()
	if _, err := f.s.completePack(ctx, id, false, "t"); errCode(err) != "PACK_FILES_MISSING" {
		t.Fatalf("complete after bad file = %v", err)
	}
	missing, err := missingBlobs(ctx, f.s.db, id)
	if err != nil {
		t.Fatal(err)
	}
	rejected := false
	for _, m := range missing {
		if m.SHA256 == bsha {
			rejected = len(m.Received) == 0 && m.Error == "PACK_FILE_HASH_MISMATCH" && m.State == StateUploading
		}
	}
	if !rejected {
		t.Fatalf("missing after hash mismatch = %+v", missing)
	}
	evs, _ := f.s.events(ctx, id, 10)
	if len(evs) == 0 || evs[0]["action"] != "file_rejected" || evs[0]["detail"].(map[string]any)["path"] != "b.bundle" {
		t.Fatalf("events = %+v", evs)
	}
	if _, err := f.s.createPack(ctx, packCreateIn{Platform: "Switch2", Version: f.prefix + "-x", Files: u.list()}, "t"); errCode(err) != "PACK_PLATFORM_NOT_ALLOWED" {
		t.Fatalf("platform = %v", err)
	}
	if _, err := f.s.createPack(ctx, packCreateIn{Platform: "iOS", Version: f.prefix + "-y", Files: []packFileIn{{Path: "../x", SHA256: strings.Repeat("a", 64)}}}, "t"); errCode(err) != "PACK_INVALID" {
		t.Fatalf("path = %v", err)
	}
	if _, err := f.s.createPack(ctx, packCreateIn{Platform: "iOS", Version: f.prefix + "-z", Files: []packFileIn{{Path: "a.bundle", SHA256: strings.Repeat("a", 64)}}}, "t"); errCode(err) != "PACK_CATALOG_MISSING" {
		t.Fatalf("catalog = %v", err)
	}
	_, _ = f.s.db.Exec(`DELETE FROM content_pack WHERE id=$1`, id)
	for i := 0; i < 3; i++ {
		f.uploadPack("iOS", fmt.Sprintf("%s-r%d", f.prefix, i), "", "", upload{files: map[string][]byte{"catalog.json": []byte(fmt.Sprintf(`{"r":%d}`, i))}}, false)
	}
	_, _ = f.s.db.Exec(`DELETE FROM content_pack WHERE platform='iOS' AND version NOT LIKE $1`, f.prefix+"%")
	if err := f.s.sweepPacks(ctx, f.s.settings(ctx)); err != nil {
		t.Fatal(err)
	}
	var n int
	_ = f.s.db.QueryRow(`SELECT count(*) FROM content_pack WHERE version LIKE $1`, f.prefix+"%").Scan(&n)
	if n != 1 {
		t.Fatalf("kept %d packs", n)
	}
	if err := f.s.collectBlobs(ctx, 0); err != nil {
		t.Fatal(err)
	}
	var orphan bool
	_ = f.s.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM content_blob WHERE sha256=$1)`, bsha).Scan(&orphan)
	if orphan {
		t.Fatal("unused blob kept")
	}
	if _, err := f.s.store.Get(ctx, partKey(bsha, 0)); !errors.Is(err, objectstore.ErrNotFound) {
		t.Fatalf("part left: %v", err)
	}
}

func TestPackUploadOnS3(t *testing.T) {
	endpoint := os.Getenv("OBJECTSTORE_TEST_S3_ENDPOINT")
	if endpoint == "" {
		t.Skip("OBJECTSTORE_TEST_S3_ENDPOINT not set; S3 NOT_RUN")
	}
	f := newFixture(t)
	store, err := objectstore.FromEnv(map[string]string{objectstore.DriverEnv: objectstore.DriverS3, objectstore.S3EndpointEnv: endpoint, objectstore.S3BucketEnv: os.Getenv("OBJECTSTORE_TEST_S3_BUCKET"),
		objectstore.S3AccessKeyEnv: os.Getenv("OBJECTSTORE_TEST_S3_ACCESS_KEY"), objectstore.S3SecretKeyEnv: os.Getenv("OBJECTSTORE_TEST_S3_SECRET_KEY")}, objectstore.Namespace{LocalDir: t.TempDir(), Prefix: "content-test/"})
	if err != nil {
		t.Fatal(err)
	}
	f.s.store = store
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 64 * 1024 })
	data := bytes.Repeat([]byte(f.prefix), 12000)
	p := f.uploadPack("Android", f.prefix+"-s3", "", "", upload{files: map[string][]byte{"catalog.json": []byte(`{}`), "b.bundle": data}}, false)
	files, err := f.s.packFiles(context.Background(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	obj, err := store.Get(context.Background(), blobKey(files["b.bundle"].sha))
	if err != nil {
		t.Fatal(err)
	}
	defer obj.Body.Close()
	got, _ := io.ReadAll(obj.Body)
	if !bytes.Equal(got, data) {
		t.Fatal("stored file differs")
	}
	if err := f.s.deletePack(context.Background(), p.ID, "t"); err != nil {
		t.Fatal(err)
	}
	if err := f.s.collectBlobs(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(context.Background(), blobKey(files["b.bundle"].sha)); !errors.Is(err, objectstore.ErrNotFound) {
		t.Fatalf("blob left on S3: %v", err)
	}
}

func call(t *testing.T, h http.HandlerFunc, method, target string, body []byte, vars map[string]string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	r := httptest.NewRequest(method, target, bytes.NewReader(body))
	for k, v := range vars {
		r.SetPathValue(k, v)
	}
	w := httptest.NewRecorder()
	h(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w, out
}

func TestConsoleUploadCatalogRPCAndFiles(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 64 * 1024; c.Packs.Platforms = []string{"Android"} })
	bundle := bytes.Repeat([]byte("road-storm-"), 9000)
	u := upload{files: map[string][]byte{"ServerData/Android/catalog_2026.json": []byte(`{"m_LocatorId":"x"}`), "Android/catalog_2026.hash": []byte("abc123\n"), "Android/cars_assets_all.bundle": bundle}}
	body, _ := json.Marshal(packCreateIn{Platform: "Android", Version: f.prefix + "-c1", MinAppVersion: "1.2.0", Note: "first", Files: u.list()})
	w, out := call(t, f.s.packCreate, http.MethodPost, "/packs", body, nil)
	if w.Code != 200 {
		t.Fatalf("create = %d %s", w.Code, w.Body.String())
	}
	id := out["pack"].(map[string]any)["id"].(string)
	missing := out["missing"].([]any)
	if len(missing) != 3 {
		t.Fatalf("missing = %v", missing)
	}
	for _, raw := range missing {
		m := raw.(map[string]any)
		sha, parts, pb := m["sha256"].(string), int(m["parts"].(float64)), int(m["partBytes"].(float64))
		var data []byte
		for _, x := range u.list() {
			if x.SHA256 == sha {
				data = u.files[x.Path]
			}
		}
		for n := 0; n < parts; n++ {
			end := min((n+1)*pb, len(data))
			w, _ := call(t, f.s.packPart, http.MethodPut, "/part", data[n*pb:end], map[string]string{"id": id, "sha": sha, "part": fmt.Sprint(n)})
			if w.Code != 200 {
				t.Fatalf("part = %d %s", w.Code, w.Body.String())
			}
		}
	}
	f.s.waitAssembly()
	w, out = call(t, f.s.packComplete, http.MethodPost, "/complete", []byte(`{"activate":true}`), map[string]string{"id": id})
	item := out["item"].(map[string]any)
	if w.Code != 200 || item["status"] != "ready" || item["active"] != true || item["catalogHash"] != "abc123" || item["catalogPath"] != "catalog_2026.json" {
		t.Fatalf("complete = %d %s", w.Code, w.Body.String())
	}
	got, err := f.s.rpcCatalog(ctx, "", `{"platform":"Android","appVersion":"1.3.1"}`)
	if err != nil {
		t.Fatal(err)
	}
	pk := got.(map[string]any)["pack"].(map[string]any)
	if pk["catalogUrl"] != "http://cdn.test/v2/content/files/"+id+"/catalog_2026.json" || pk["hashUrl"] != "http://cdn.test/v2/content/files/"+id+"/catalog_2026.hash" || pk["baseUrl"] != "http://cdn.test/v2/content/files/"+id {
		t.Fatalf("catalog = %+v", pk)
	}
	if got, _ := f.s.rpcCatalog(ctx, "", `{"platform":"Android","appVersion":"1.1"}`); got.(map[string]any)["pack"] != nil {
		t.Fatalf("old app got a pack: %+v", got)
	}
	if _, err := f.s.rpcCatalog(ctx, "", `{"platform":""}`); errCode(err) != "PLATFORM_REQUIRED" {
		t.Fatalf("platform = %v", err)
	}
	f.settings(func(c *settings.All) { c.Packs.RequireAppVersion = true })
	if _, err := f.s.rpcCatalog(ctx, "", `{"platform":"Android"}`); errCode(err) != "APP_VERSION_REQUIRED" {
		t.Fatalf("app version = %v", err)
	}
	r := httptest.NewRequest(http.MethodGet, "/v2/content/files/"+id+"/cars_assets_all.bundle", nil)
	r.SetPathValue("pack", id)
	r.SetPathValue("path", "cars_assets_all.bundle")
	fw := httptest.NewRecorder()
	f.s.httpContentFile(fw, r)
	if fw.Code != 200 || !bytes.Equal(fw.Body.Bytes(), bundle) || !strings.Contains(fw.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("file = %d %d %q", fw.Code, fw.Body.Len(), fw.Header().Get("Cache-Control"))
	}
	r = httptest.NewRequest(http.MethodGet, "/v2/content/files/"+id+"/nope.bundle", nil)
	r.SetPathValue("pack", id)
	r.SetPathValue("path", "nope.bundle")
	fw = httptest.NewRecorder()
	f.s.httpContentFile(fw, r)
	if fw.Code != 404 {
		t.Fatalf("missing file = %d", fw.Code)
	}
	w, out = call(t, f.s.packList, http.MethodGet, "/packs", nil, nil)
	if w.Code != 200 || out["current"].(map[string]any)["Android"] != id {
		t.Fatalf("list = %s", w.Body.String())
	}
	w, _ = call(t, f.s.packDelete, http.MethodDelete, "/packs", nil, map[string]string{"id": id})
	if w.Code != 409 {
		t.Fatalf("delete active = %d %s", w.Code, w.Body.String())
	}
	w, _ = call(t, f.s.settingsPut, http.MethodPut, "/settings", []byte(`{"packs":{"partBytes":10}}`), nil)
	if w.Code != 400 {
		t.Fatalf("bad settings = %d %s", w.Code, w.Body.String())
	}
}
