package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama/v3/internal/objectstore"
)

func s3TestEnv(t *testing.T) map[string]string {
	t.Helper()
	endpoint := os.Getenv("OBJECTSTORE_TEST_S3_ENDPOINT")
	if endpoint == "" {
		t.Skip("OBJECTSTORE_TEST_S3_ENDPOINT not set; S3 NOT_RUN")
	}
	return map[string]string{
		objectstore.DriverEnv: objectstore.DriverS3, objectstore.S3EndpointEnv: endpoint, objectstore.S3BucketEnv: os.Getenv("OBJECTSTORE_TEST_S3_BUCKET"),
		objectstore.S3AccessKeyEnv: os.Getenv("OBJECTSTORE_TEST_S3_ACCESS_KEY"), objectstore.S3SecretKeyEnv: os.Getenv("OBJECTSTORE_TEST_S3_SECRET_KEY"),
		objectstore.S3PrefixEnv: "economy-test",
	}
}

func uploadPNG(t *testing.T, h *harness, payload []byte) map[string]string {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "a.png")
	_, _ = fw.Write(payload)
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/v2/console/economy/uploads", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	h.init.console.ServeHTTP(rec, req)
	var up map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &up)
	if rec.Code != 201 || up["url"] != "http://127.0.0.1:7450/v2/economy/files/"+up["name"] {
		t.Fatalf("upload = %d %s", rec.Code, rec.Body.String())
	}
	return up
}

func TestEconomyUploadsOnS3(t *testing.T) {
	env := s3TestEnv(t)
	h := newHarness(t)
	cfg, err := objectstore.ConfigFromEnv(env, objectstore.Namespace{LocalDir: t.TempDir(), Prefix: "economy/"})
	if err != nil {
		t.Fatal(err)
	}
	other, _ := objectstore.New(cfg, nil)
	h.s.store = other
	payload := append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{7}, 4096)...)
	up := uploadPNG(t, h, payload)
	t.Cleanup(func() { _ = other.Delete(h.ctx, up["name"]) })

	second := newHarness(t)
	node2, _ := objectstore.New(cfg, nil)
	second.s.store = node2
	rec := httptest.NewRecorder()
	second.init.public.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v2/economy/files/"+up["name"], nil))
	got, _ := io.ReadAll(rec.Body)
	if rec.Code != 200 || !bytes.Equal(got, payload) || rec.Header().Get("Content-Type") != "image/png" || !strings.Contains(rec.Header().Get("Content-Security-Policy"), "sandbox") {
		t.Fatalf("other node file = %d %d bytes %v", rec.Code, len(got), rec.Header())
	}
	ranged := httptest.NewRequest(http.MethodGet, "/v2/economy/files/"+up["name"], nil)
	ranged.Header.Set("Range", "bytes=0-7")
	rec = httptest.NewRecorder()
	second.init.public.ServeHTTP(rec, ranged)
	if rec.Code != 206 || rec.Body.String() != "\x89PNG\r\n\x1a\n" {
		t.Fatalf("range = %d %q", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	second.init.public.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v2/economy/files/"+strings.Repeat("0", 32)+".png", nil))
	if rec.Code != 404 {
		t.Fatalf("missing = %d", rec.Code)
	}

	env[objectstore.S3PublicBaseEnv] = "https://cdn.example.com/media"
	cfg, _ = objectstore.ConfigFromEnv(env, objectstore.Namespace{LocalDir: t.TempDir(), Prefix: "economy/"})
	second.s.store, _ = objectstore.New(cfg, nil)
	rec = httptest.NewRecorder()
	second.init.public.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v2/economy/files/"+up["name"], nil))
	if rec.Code != 302 || rec.Header().Get("Location") != "https://cdn.example.com/media/economy-test/economy/"+up["name"] {
		t.Fatalf("redirect = %d %s", rec.Code, rec.Header().Get("Location"))
	}
}
