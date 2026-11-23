package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/content/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/objectstore"
	"github.com/heroiclabs/nakama/v3/internal/objectstore/s3test"
)

func randomData(t *testing.T, n int) []byte {
	t.Helper()
	out := make([]byte, n)
	if _, err := rand.Read(out); err != nil {
		t.Fatal(err)
	}
	return out
}

func shaOf(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func (f *fixture) createBig(version string, data []byte) (string, missingFile) {
	f.t.Helper()
	u := upload{files: map[string][]byte{"catalog.json": []byte(`{"v":"` + version + `"}`), "map.bundle": data}}
	out, err := f.s.createPack(context.Background(), packCreateIn{Platform: "Android", Version: version, Files: u.list()}, "t")
	if err != nil {
		f.t.Fatal(err)
	}
	id := out["pack"].(pack).ID
	var big missingFile
	for _, m := range out["missing"].([]missingFile) {
		if m.SHA256 == shaOf(data) {
			big = m
		}
		if m.Path == "catalog.json" {
			if _, err := f.s.putPart(context.Background(), id, m.SHA256, 0, bytes.NewReader(u.files["catalog.json"])); err != nil {
				f.t.Fatal(err)
			}
		}
	}
	if big.Parts < 2 {
		f.t.Fatalf("big file has %d parts", big.Parts)
	}
	return id, big
}

func (f *fixture) sendParts(s *service, id string, m missingFile, data []byte, skip map[int]bool) map[string]any {
	f.t.Helper()
	var last map[string]any
	for n := 0; n < m.Parts; n++ {
		if skip[n] {
			continue
		}
		end := min((n+1)*m.PartBytes, len(data))
		res, err := s.putPart(context.Background(), id, m.SHA256, n, bytes.NewReader(data[n*m.PartBytes:end]))
		if err != nil {
			f.t.Fatal(err)
		}
		last = res
	}
	return last
}

func (f *fixture) download(id, path string) []byte {
	f.t.Helper()
	r := httptest.NewRequest(http.MethodGet, "/v2/content/files/"+id+"/"+path, nil)
	r.SetPathValue("pack", id)
	r.SetPathValue("path", path)
	w := httptest.NewRecorder()
	f.s.httpContentFile(w, r)
	if w.Code != 200 {
		f.t.Fatalf("download %s = %d", path, w.Code)
	}
	return w.Body.Bytes()
}

func TestLargeBlobAssemblesOffTheRequest(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 256 * 1024; c.Packs.RedirectToStorage = false })
	data := randomData(t, 22<<20+4321)
	id, big := f.createBig(f.prefix+"-big", data)
	if big.Parts != 89 || big.State != StateUploading {
		t.Fatalf("plan = %+v", big)
	}
	release := make(chan struct{})
	started := make(chan string, 1)
	f.s.assembleHook = func(sha string) {
		started <- sha
		<-release
	}
	res := f.sendParts(f.s, id, big, data, map[int]bool{40: true})
	if res["state"] != StateUploading {
		t.Fatalf("before last = %v", res)
	}
	at := time.Now()
	res = f.sendParts(f.s, id, big, data, map[int]bool{})
	if res["state"] != StateAssembling || res["stored"] != false || int(res["received"].(int)) != big.Parts {
		t.Fatalf("last part = %v", res)
	}
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("assembly did not start")
	}
	if time.Since(at) > 5*time.Second {
		t.Fatalf("parts took %s", time.Since(at))
	}
	_, err := f.s.completePack(ctx, id, true, "t")
	if errCode(err) != "PACK_ASSEMBLING" {
		t.Fatalf("complete while assembling = %v", err)
	}
	w, out := call(t, f.s.packComplete, http.MethodPost, "/complete", []byte(`{}`), map[string]string{"id": id})
	if w.Code != http.StatusConflict || out["error"] != "PACK_ASSEMBLING" {
		t.Fatalf("http complete = %d %s", w.Code, w.Body.String())
	}
	w, out = call(t, f.s.packGet, http.MethodGet, "/packs/"+id, nil, map[string]string{"id": id})
	miss := out["missing"].([]any)
	if w.Code != 200 || len(miss) != 1 || miss[0].(map[string]any)["state"] != StateAssembling {
		t.Fatalf("pack get = %s", w.Body.String())
	}
	close(release)
	f.s.waitAssembly()
	p, err := f.s.completePack(ctx, id, true, "t")
	if err != nil || p.Status != "ready" {
		t.Fatalf("complete = %+v %v", p, err)
	}
	if got := f.download(id, "map.bundle"); shaOf(got) != big.SHA256 || len(got) != len(data) {
		t.Fatalf("download differs: %d bytes", len(got))
	}
	if _, err := f.s.store.Get(ctx, partKey(big.SHA256, 0)); err == nil {
		t.Fatal("parts not removed")
	}
	_, _ = f.s.db.Exec(`DELETE FROM content_blob WHERE sha256=$1`, big.SHA256)
}

func TestAssemblyRecoversAfterNodeDies(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 64 * 1024; c.Packs.RedirectToStorage = false })
	data := randomData(t, 3<<20+17)
	id, big := f.createBig(f.prefix+"-dead", data)
	release := make(chan struct{})
	started := make(chan struct{}, 1)
	f.s.assembleHook = func(string) {
		started <- struct{}{}
		<-release
	}
	defer close(release)
	f.sendParts(f.s, id, big, data, nil)
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("node a did not start")
	}
	nodeB := newService(f.s.db, f.s.logger, f.s.store, "http://cdn.test")
	nodeB.holder = "node-b"
	if err := nodeB.resumeAssemblies(ctx); err != nil {
		t.Fatal(err)
	}
	nodeB.waitAssembly()
	var stored bool
	var holder string
	_ = f.s.db.QueryRow(`SELECT stored, assemble_holder FROM content_blob WHERE sha256=$1`, big.SHA256).Scan(&stored, &holder)
	if stored || holder != f.s.holder {
		t.Fatalf("live lease taken: stored=%v holder=%q", stored, holder)
	}
	if _, err := f.s.db.Exec(`UPDATE content_blob SET assemble_until=now() - interval '1 second' WHERE sha256=$1`, big.SHA256); err != nil {
		t.Fatal(err)
	}
	if _, err := nodeB.completePack(ctx, id, false, "t"); errCode(err) != "PACK_ASSEMBLING" {
		t.Fatalf("complete on node b = %v", err)
	}
	nodeB.waitAssembly()
	p, err := nodeB.completePack(ctx, id, false, "t")
	if err != nil || p.Status != "ready" {
		t.Fatalf("complete after recovery = %+v %v", p, err)
	}
	var attempts int
	_ = f.s.db.QueryRow(`SELECT stored, assemble_attempts FROM content_blob WHERE sha256=$1`, big.SHA256).Scan(&stored, &attempts)
	if !stored || attempts != 2 {
		t.Fatalf("stored=%v attempts=%d", stored, attempts)
	}
	if got := f.downloadFrom(nodeB, id, "map.bundle"); !bytes.Equal(got, data) {
		t.Fatal("recovered file differs")
	}
	_, _ = f.s.db.Exec(`DELETE FROM content_blob WHERE sha256=$1`, big.SHA256)
}

func (f *fixture) downloadFrom(s *service, id, path string) []byte {
	saved := f.s
	f.s = s
	defer func() { f.s = saved }()
	return f.download(id, path)
}

func TestAssemblyLostPartAsksForItAgain(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 64 * 1024 })
	data := randomData(t, 640*1024)
	id, big := f.createBig(f.prefix+"-lost", data)
	f.s.assembleHook = func(sha string) { _ = f.s.store.Delete(ctx, partKey(sha, 3)) }
	f.sendParts(f.s, id, big, data, nil)
	f.s.waitAssembly()
	f.s.assembleHook = nil
	missing, _ := missingBlobs(ctx, f.s.db, id)
	if len(missing) != 1 || missing[0].Error != assemblePartLost || len(missing[0].Received) != big.Parts-1 {
		t.Fatalf("missing = %+v", missing)
	}
	for _, n := range missing[0].Received {
		if n == 3 {
			t.Fatal("lost part still counted")
		}
	}
	res, err := f.s.putPart(ctx, id, big.SHA256, 3, bytes.NewReader(data[3*64*1024:4*64*1024]))
	if err != nil || res["state"] != StateAssembling {
		t.Fatalf("resend = %v %v", res, err)
	}
	f.s.waitAssembly()
	if _, err := f.s.completePack(ctx, id, false, "t"); err != nil {
		t.Fatal(err)
	}
	_, _ = f.s.db.Exec(`DELETE FROM content_blob WHERE sha256=$1`, big.SHA256)
}

func TestLargeBlobOnFakeS3Multipart(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	fake := s3test.New()
	srv := httptest.NewServer(fake)
	defer srv.Close()
	store, err := objectstore.New(objectstore.Config{Driver: objectstore.DriverS3, Endpoint: srv.URL, Bucket: "game", Region: "us-east-1", AccessKey: "a", SecretKey: "b", PathStyle: true, Prefix: "content/"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	f.s.store = store
	f.settings(func(c *settings.All) { c.Packs.PartBytes = 1 << 20; c.Packs.RedirectToStorage = false })
	data := randomData(t, 21<<20+99)
	id, big := f.createBig(f.prefix+"-s3", data)
	res := f.sendParts(f.s, id, big, data, nil)
	if res["state"] != StateAssembling {
		t.Fatalf("last part = %v", res)
	}
	f.s.waitAssembly()
	if _, err := f.s.completePack(ctx, id, false, "t"); err != nil {
		t.Fatal(err)
	}
	completed, _, aborted, open, sizes := fake.Snapshot()
	if completed != 1 || aborted != 0 || open != 0 || len(sizes) != 3 || sizes[0] != objectstore.DefaultMultipartBytes || sizes[1] != objectstore.DefaultMultipartBytes {
		t.Fatalf("multipart: completed=%d aborted=%d open=%d sizes=%v", completed, aborted, open, sizes)
	}
	if got := f.download(id, "map.bundle"); !bytes.Equal(got, data) {
		t.Fatal("s3 file differs")
	}
	for n := 0; n < big.Parts; n++ {
		if _, ok := fake.Object(fmt.Sprintf("/game/content/%s", partKey(big.SHA256, n))); ok {
			t.Fatalf("part %d left on s3", n)
		}
	}
	_, _ = f.s.db.Exec(`DELETE FROM content_blob WHERE sha256=$1`, big.SHA256)
}
