package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"
	"time"
)

func tgz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(body))}); err != nil {
			t.Fatal(err)
		}
		_, _ = tw.Write([]byte(body))
	}
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

func TestReadManifestAndSemverOrder(t *testing.T) {
	m, err := readManifest(tgz(t, map[string]string{"package/README.md": "x", "package/package.json": `{"name":"com.example.sdk","version":"0.6.0"}`}))
	if err != nil || m["name"] != "com.example.sdk" {
		t.Fatalf("manifest = %v %v", m, err)
	}
	if _, err := readManifest([]byte("not gzip")); err == nil {
		t.Fatal("garbage accepted")
	}
	if _, err := readManifest(tgz(t, map[string]string{"package/other.json": "{}"})); err == nil {
		t.Fatal("missing package.json accepted")
	}
	for _, v := range []string{"1.2", "01.2.3", "1.2.3-", "v1.2.3"} {
		if _, ok := parseSemver(v); ok {
			t.Fatalf("%s accepted", v)
		}
	}
	order := []string{"0.6.0-beta.1", "0.6.0", "0.6.1", "0.10.0", "1.0.0-rc.1"}
	for i := 1; i < len(order); i++ {
		a, _ := parseSemver(order[i-1])
		b, _ := parseSemver(order[i])
		if !a.less(b) || b.less(a) {
			t.Fatalf("%s should sort before %s", order[i-1], order[i])
		}
	}
	now := time.Now()
	list := []packageVersion{{Name: "com.example.sdk", Version: "0.6.0", Manifest: []byte(`{"name":"com.example.sdk","version":"0.6.0","description":"d"}`), Shasum: "s", Integrity: "i", CreatedAt: now},
		{Name: "com.example.sdk", Version: "1.0.0-rc.1", Manifest: []byte(`{"name":"com.example.sdk","version":"1.0.0-rc.1"}`), CreatedAt: now.Add(time.Minute)}}
	list[0].fillFromManifest()
	doc := packument("com.example.sdk", list, "https://api.example.com/upm")
	if doc["dist-tags"].(map[string]string)["latest"] != "0.6.0" {
		t.Fatalf("latest must skip pre-releases: %v", doc["dist-tags"])
	}
	v := doc["versions"].(map[string]any)["0.6.0"].(map[string]any)
	if v["dist"].(map[string]any)["tarball"] != "https://api.example.com/upm/com.example.sdk/-/com.example.sdk-0.6.0.tgz" || v["_id"] != "com.example.sdk@0.6.0" {
		t.Fatalf("version doc = %v", v)
	}
}
