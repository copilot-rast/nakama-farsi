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
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/heroiclabs/nakama/v3/integrations/assets/internal/settings"
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
	f := &fixture{t: t, s: newService(db, plugintest.Logger{}, store, "http://cdn.test"), prefix: plugintest.Key("pt")}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM assets_entry WHERE asset_id LIKE $1`, f.prefix+"%")
		_, _ = db.Exec(`DELETE FROM assets_entry_history WHERE asset_id LIKE $1`, f.prefix+"%")
		_, _ = db.Exec(`DELETE FROM assets_pack WHERE version LIKE $1`, f.prefix+"%")
		_, _ = db.Exec(`DELETE FROM assets_pack_event WHERE version LIKE $1`, f.prefix+"%")
		_, _ = db.Exec(`UPDATE assets_settings SET value='{}' WHERE id=1`)
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

func (f *fixture) id(name string) string { return f.prefix + "_" + name }

func item(id, folder string, data string, base *int) entryIn {
	return entryIn{AssetID: id, SourceGUID: "g-" + id, Type: "Card", Folder: folder, Name: id, Data: json.RawMessage(data), BaseVersion: base}
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

func TestPushVersionsAndConflicts(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	a, b := f.id("a"), f.id("b")
	out, err := f.s.push(ctx, pushIn{Items: []entryIn{item(a, "Cars/BMW", `{"speed":1.50}`, nil), item(b, "Cars", `{"speed":2}`, nil)}}, "t", true)
	if err != nil || out.Created != 2 {
		t.Fatalf("create = %+v %v", out, err)
	}
	// The same content again is never a conflict, even without a base version.
	out, err = f.s.push(ctx, pushIn{Items: []entryIn{item(a, "Cars/BMW", `{ "speed" : 1.50 }`, nil)}}, "t", true)
	if err != nil || out.Unchanged != 1 {
		t.Fatalf("same = %+v %v", out, err)
	}
	// A change without the current version is refused as a whole (default reject mode).
	_, err = f.s.push(ctx, pushIn{Items: []entryIn{item(a, "Cars/BMW", `{"speed":3}`, intp(0)), item(b, "Cars", `{"speed":5}`, intp(1))}}, "t", true)
	if errCode(err) != "ASSET_VERSION_CONFLICT" {
		t.Fatalf("stale = %v", err)
	}
	e, _ := apierr.As(err)
	if c := e.Meta["conflicts"].([]conflict); len(c) != 1 || c[0].AssetID != a || *c[0].Version != 1 {
		t.Fatalf("conflicts = %+v", e.Meta)
	}
	if got := f.version(b); got != 1 {
		t.Fatalf("b changed in a refused push: %d", got)
	}
	// Skip mode writes the others.
	f.settings(func(c *settings.All) { c.Entries.ConflictMode = settings.ConflictSkip })
	out, err = f.s.push(ctx, pushIn{Items: []entryIn{item(a, "Cars/BMW", `{"speed":3}`, intp(0)), item(b, "Cars", `{"speed":5}`, intp(1))}}, "t", true)
	if err != nil || out.Updated != 1 || len(out.Conflicts) != 1 || f.version(b) != 2 {
		t.Fatalf("skip = %+v %v", out, err)
	}
	// Dry run reports and writes nothing.
	out, err = f.s.push(ctx, pushIn{DryRun: true, Items: []entryIn{item(a, "Cars/BMW", `{"speed":9}`, intp(1))}}, "t", true)
	if err != nil || out.Updated != 1 || f.version(a) != 1 {
		t.Fatalf("dry run = %+v %v v%d", out, err, f.version(a))
	}
	// Deletes need the current version and the delete right.
	if _, err := f.s.push(ctx, pushIn{Deletes: []deleteIn{{AssetID: b, BaseVersion: intp(2)}}}, "t", false); errCode(err) != "ASSET_DELETE_NOT_ALLOWED" {
		t.Fatalf("delete right = %v", err)
	}
	out, err = f.s.push(ctx, pushIn{Deletes: []deleteIn{{AssetID: b, BaseVersion: intp(2)}}}, "t", true)
	if err != nil || out.Deleted != 1 {
		t.Fatalf("delete = %+v %v", out, err)
	}
	// Recreated ids continue their version numbers, so client caches see a change.
	out, err = f.s.push(ctx, pushIn{Items: []entryIn{item(b, "Cars", `{"speed":5}`, nil)}}, "t", true)
	if err != nil || out.Created != 1 || *out.Rows[0].Version != 3 {
		t.Fatalf("recreate = %+v %v", out, err)
	}
}

func TestPushRefusesDuplicatesAndSourceMoves(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	a, b := f.id("a"), f.id("b")
	x := item(a, "", `{}`, nil)
	if _, err := f.s.push(ctx, pushIn{Items: []entryIn{x, x}}, "t", true); errCode(err) != "ASSET_ID_DUPLICATED" {
		t.Fatalf("dup id = %v", err)
	}
	y := item(b, "", `{}`, nil)
	y.SourceGUID = x.SourceGUID
	if _, err := f.s.push(ctx, pushIn{Items: []entryIn{x, y}}, "t", true); errCode(err) != "ASSET_SOURCE_DUPLICATED" {
		t.Fatalf("dup source = %v", err)
	}
	if _, err := f.s.push(ctx, pushIn{Items: []entryIn{x}}, "t", true); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.push(ctx, pushIn{Items: []entryIn{y}}, "t", true); errCode(err) != "ASSET_SOURCE_ID_CHANGED" {
		t.Fatalf("renamed id = %v", err)
	}
	// Renaming works when the old id is deleted in the same push.
	if _, err := f.s.push(ctx, pushIn{Items: []entryIn{y}, Deletes: []deleteIn{{AssetID: a, BaseVersion: intp(1)}}}, "t", true); err != nil {
		t.Fatalf("rename with delete = %v", err)
	}
	bad := item(f.id("c"), "a/../b", `{}`, nil)
	if _, err := f.s.push(ctx, pushIn{Items: []entryIn{bad}}, "t", true); errCode(err) != "ASSET_INVALID" {
		t.Fatalf("folder = %v", err)
	}
	f.settings(func(c *settings.All) { c.Entries.MaxEntryBytes = 1024 })
	big := item(f.id("d"), "", `{"x":"`+strings.Repeat("a", 2000)+`"}`, nil)
	if _, err := f.s.push(ctx, pushIn{Items: []entryIn{big}}, "t", true); errCode(err) != "ASSET_TOO_LARGE" {
		t.Fatalf("size = %v", err)
	}
}

func (f *fixture) version(id string) int {
	var v int
	_ = f.s.db.QueryRow(`SELECT version FROM assets_entry WHERE asset_id=$1`, id).Scan(&v)
	return v
}

func TestParallelPushesWithSameBaseVersion(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	a := f.id("race")
	if _, err := f.s.push(ctx, pushIn{Items: []entryIn{item(a, "", `{"n":0}`, nil)}}, "t", true); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make([]error, 8)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, results[i] = f.s.push(ctx, pushIn{Items: []entryIn{item(a, "", fmt.Sprintf(`{"n":%d}`, i+1), intp(1))}}, "t", true)
		}(i)
	}
	wg.Wait()
	ok, conflicts := 0, 0
	for _, err := range results {
		switch errCode(err) {
		case "":
			ok++
		case "ASSET_VERSION_CONFLICT":
			conflicts++
		default:
			t.Fatalf("unexpected %v", err)
		}
	}
	if ok != 1 || conflicts != 7 || f.version(a) != 2 {
		t.Fatalf("ok %d conflicts %d version %d", ok, conflicts, f.version(a))
	}
	// Parallel creation of the same new id: one wins.
	n := f.id("new")
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, results[i] = f.s.push(ctx, pushIn{Items: []entryIn{item(n, "", fmt.Sprintf(`{"n":%d}`, i), nil)}}, "t", true)
		}(i)
	}
	wg.Wait()
	ok = 0
	for _, err := range results {
		if err == nil {
			ok++
		} else if errCode(err) != "ASSET_VERSION_CONFLICT" {
			t.Fatalf("create race: %v", err)
		}
	}
	if ok < 1 || f.version(n) != 1 {
		t.Fatalf("create race ok=%d v=%d", ok, f.version(n))
	}
}

func TestDeltaFetch(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	folder := f.prefix + "/Cars"
	ids := []string{f.id("1"), f.id("2"), f.id("3")}
	items := []entryIn{}
	for _, id := range ids {
		items = append(items, item(id, folder, `{"v":1}`, nil))
	}
	items = append(items, item(f.id("other"), f.prefix+"/Guns", `{}`, nil))
	if _, err := f.s.push(ctx, pushIn{Items: items}, "t", true); err != nil {
		t.Fatal(err)
	}
	first, err := f.s.fetch(ctx, fetchIn{scope: scope{Folder: folder}})
	if err != nil || len(first.Items) != 3 || first.Total != 3 {
		t.Fatalf("first = %+v %v", first, err)
	}
	have := map[string]int{}
	for _, e := range first.Items {
		have[e.AssetID] = e.Version
	}
	have[f.id("gone")] = 4
	if _, err := f.s.push(ctx, pushIn{Items: []entryIn{item(ids[1], folder, `{"v":2}`, intp(1))}}, "t", true); err != nil {
		t.Fatal(err)
	}
	d, err := f.s.fetch(ctx, fetchIn{scope: scope{Folder: folder}, Have: have})
	if err != nil || len(d.Items) != 1 || d.Items[0].AssetID != ids[1] || string(d.Items[0].Data) != `{"v":2}` || d.Unchanged != 2 || len(d.Removed) != 1 || d.Removed[0] != f.id("gone") {
		t.Fatalf("delta = %+v %v", d, err)
	}
	// Paging by limit: the client sends its new versions and gets the rest.
	page, _ := f.s.fetch(ctx, fetchIn{scope: scope{Folder: f.prefix}, Limit: 2})
	if len(page.Items) != 2 || !page.More {
		t.Fatalf("page = %+v", page)
	}
	// Manifest ETag changes with a version.
	m1, _ := f.s.rpcManifest(ctx, "", `{"folder":"`+folder+`"}`)
	tag := m1.(map[string]any)["etag"].(string)
	m2, _ := f.s.rpcManifest(ctx, "", `{"folder":"`+folder+`","etag":`+jsonString(tag)+`}`)
	if m2.(map[string]any)["notModified"] != true {
		t.Fatalf("etag = %v", m2)
	}
	if _, err := f.s.push(ctx, pushIn{Items: []entryIn{item(ids[0], folder, `{"v":3}`, intp(1))}}, "t", true); err != nil {
		t.Fatal(err)
	}
	m3, _ := f.s.rpcManifest(ctx, "", `{"folder":"`+folder+`","etag":`+jsonString(tag)+`}`)
	if m3.(map[string]any)["notModified"] != false {
		t.Fatalf("etag after change = %v", m3)
	}
	// Non-recursive scope.
	flat, _ := f.s.manifest(ctx, scope{Folder: f.prefix, Recursive: new(bool)}, false)
	if len(flat) != 0 {
		t.Fatalf("flat = %+v", flat)
	}
}

func jsonString(s string) string { b, _ := json.Marshal(s); return string(b) }

func TestHistoryAndRollback(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	a := f.id("h")
	for i, data := range []string{`{"v":1}`, `{"v":2}`, `{"v":3}`} {
		var base *int
		if i > 0 {
			base = intp(i)
		}
		if _, err := f.s.push(ctx, pushIn{Items: []entryIn{item(a, "", data, base)}}, "t", true); err != nil {
			t.Fatal(err)
		}
	}
	hist, err := f.s.history(ctx, a, 10)
	if err != nil || len(hist) != 3 {
		t.Fatalf("history = %+v %v", hist, err)
	}
	first := hist[2].ID
	if _, err := f.s.rollback(ctx, a, first, intp(2), "admin"); errCode(err) != "ASSET_VERSION_CONFLICT" {
		t.Fatalf("stale rollback = %v", err)
	}
	e, err := f.s.rollback(ctx, a, first, intp(3), "admin")
	if err != nil || e.Version != 4 || string(e.Data) != `{"v":1}` {
		t.Fatalf("rollback = %+v %v", e, err)
	}
	// A deleted entry comes back from its history.
	if _, err := f.s.push(ctx, pushIn{Deletes: []deleteIn{{AssetID: a, BaseVersion: intp(4)}}}, "t", true); err != nil {
		t.Fatal(err)
	}
	hist, _ = f.s.history(ctx, a, 1)
	if hist[0].Action != "deleted" {
		t.Fatalf("last = %+v", hist[0])
	}
	e, err = f.s.rollback(ctx, a, hist[0].ID, nil, "admin")
	if err != nil || e.Version != 5 || string(e.Data) != `{"v":1}` {
		t.Fatalf("restore = %+v %v", e, err)
	}
	// Trimming keeps the newest states.
	f.settings(func(c *settings.All) { c.History.KeepVersions = 2 })
	f.s.trimHistory(ctx, f.s.settings(ctx))
	hist, _ = f.s.history(ctx, a, 10)
	if len(hist) != 2 || hist[0].Version != 5 {
		t.Fatalf("trimmed = %+v", hist)
	}
}

// ---- packs ----

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
	ctx := context.Background()
	out, err := f.s.createPack(ctx, packCreateIn{Platform: platform, Version: version, MinAppVersion: min, MaxAppVersion: max, Files: u.list()}, "t")
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
		for n := m.Parts - 1; n >= 0; n-- { // out of order on purpose
			end := (n + 1) * m.PartBytes
			if end > len(data) {
				end = len(data)
			}
			if _, err := f.s.putPart(ctx, p.ID, m.SHA256, n, bytes.NewReader(data[n*m.PartBytes:end])); err != nil {
				f.t.Fatal(err)
			}
		}
	}
	p, err = f.s.completePack(ctx, p.ID, activate, "t")
	if err != nil {
		f.t.Fatal(err)
	}
	return p
}

func TestPackUploadActivationAndIndex(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 16 * 1024 })
	bundle := bytes.Repeat([]byte("bundle-data-"), 5000) // 60 KB → 4 parts
	u := upload{files: map[string][]byte{"Android/catalog_1.json": []byte(`{"c":1}`), "Android/catalog_1.hash": []byte("h1"), "Android/x.bundle": bundle}}
	p1 := f.uploadPack("Android", f.prefix+"-1", "1.0", "1.9", u, true)
	if p1.Status != "ready" || !p1.Active || p1.CatalogPath != "catalog_1.json" || p1.CatalogHash != "h1" || p1.FileCount != 3 {
		t.Fatalf("p1 = %+v", p1)
	}
	// A second pack reuses stored files: only the new catalog is missing.
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
	// Rollback: activating the older pack again makes it win where both match.
	if _, err := f.s.activate(ctx, p1.ID, true, "admin"); err != nil {
		t.Fatal(err)
	}
	if idx("1.6") != f.prefix+"-1" {
		t.Fatalf("rollback = %q", idx("1.6"))
	}
	// Single active per range: the overlapping pack is switched off.
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 16 * 1024; c.Packs.SingleActive = true })
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
	// Files are served from the pack folder.
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
	if view["catalogUrl"] != "http://cdn.test/v2/assets/content/"+p2.ID+"/catalog_2.json" {
		t.Fatalf("view = %+v", view)
	}
}

func TestPackPartChecksAndRetention(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 16 * 1024; c.Packs.KeepInactive = 1 })
	data := bytes.Repeat([]byte("z"), 20*1024)
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
	wrong := bytes.Repeat([]byte("y"), 16*1024)
	if _, err := f.s.putPart(ctx, id, bsha, 0, bytes.NewReader(wrong)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.putPart(ctx, id, bsha, 1, bytes.NewReader(data[16*1024:])); errCode(err) != "PACK_FILE_HASH_MISMATCH" {
		t.Fatalf("hash = %v", err)
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
	// Retention: of three inactive ready packs one is kept; expired uploads go.
	_, _ = f.s.db.Exec(`DELETE FROM assets_pack WHERE id=$1`, id)
	for i := 0; i < 3; i++ {
		f.uploadPack("iOS", fmt.Sprintf("%s-r%d", f.prefix, i), "", "", upload{files: map[string][]byte{"catalog.json": []byte(fmt.Sprintf(`{"r":%d}`, i))}}, false)
	}
	_, _ = f.s.db.Exec(`DELETE FROM assets_pack WHERE platform='iOS' AND version NOT LIKE $1`, f.prefix+"%")
	if err := f.s.sweepPacks(ctx, f.s.settings(ctx)); err != nil {
		t.Fatal(err)
	}
	var n int
	_ = f.s.db.QueryRow(`SELECT count(*) FROM assets_pack WHERE version LIKE $1`, f.prefix+"%").Scan(&n)
	if n != 1 {
		t.Fatalf("kept %d packs", n)
	}
	if err := f.s.collectBlobs(ctx, 0); err != nil {
		t.Fatal(err)
	}
	var orphan bool
	_ = f.s.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM assets_blob WHERE sha256=$1)`, bsha).Scan(&orphan)
	if orphan {
		t.Fatal("unused blob kept")
	}
	if _, err := f.s.store.Get(ctx, partKey(bsha, 0)); !errors.Is(err, objectstore.ErrNotFound) {
		t.Fatalf("part left: %v", err)
	}
}

// TestPackUploadOnS3 runs a part upload against a real S3 store (MinIO) when OBJECTSTORE_TEST_S3_* is set.
func TestPackUploadOnS3(t *testing.T) {
	endpoint := os.Getenv("OBJECTSTORE_TEST_S3_ENDPOINT")
	if endpoint == "" {
		t.Skip("OBJECTSTORE_TEST_S3_ENDPOINT not set; S3 NOT_RUN")
	}
	f := newFixture(t)
	store, err := objectstore.FromEnv(map[string]string{objectstore.DriverEnv: objectstore.DriverS3, objectstore.S3EndpointEnv: endpoint, objectstore.S3BucketEnv: os.Getenv("OBJECTSTORE_TEST_S3_BUCKET"),
		objectstore.S3AccessKeyEnv: os.Getenv("OBJECTSTORE_TEST_S3_ACCESS_KEY"), objectstore.S3SecretKeyEnv: os.Getenv("OBJECTSTORE_TEST_S3_SECRET_KEY")}, objectstore.Namespace{LocalDir: t.TempDir(), Prefix: "assets-test/"})
	if err != nil {
		t.Fatal(err)
	}
	f.s.store = store
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 16 * 1024 })
	data := bytes.Repeat([]byte(f.prefix), 3000)
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
