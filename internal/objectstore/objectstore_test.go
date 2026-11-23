package objectstore

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	vectorAccess = "AKIAIOSFODNN7EXAMPLE"
	vectorSecret = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
)

var vectorTime = time.Date(2013, 5, 24, 0, 0, 0, 0, time.UTC)

func signature(t *testing.T, req *http.Request, payload string) string {
	t.Helper()
	SignV4(req, payload, vectorAccess, vectorSecret, "us-east-1", "s3", vectorTime)
	auth := req.Header.Get("Authorization")
	i := strings.Index(auth, "Signature=")
	if i < 0 {
		t.Fatalf("no signature in %q", auth)
	}
	return auth[i+len("Signature="):]
}

func TestSignV4MatchesAWSExamples(t *testing.T) {
	get, _ := http.NewRequest(http.MethodGet, "https://examplebucket.s3.amazonaws.com/test.txt", nil)
	get.Header.Set("Range", "bytes=0-9")
	if got := signature(t, get, emptySHA256Hex); got != "f0e8bdb87c964420e857bd35b5d6ed310bd44f0170aba48dd91039c6036bdb41" {
		t.Errorf("GET object signature = %s", got)
	}
	if !strings.Contains(get.Header.Get("Authorization"), "Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host;range;x-amz-content-sha256;x-amz-date,") {
		t.Errorf("authorization = %s", get.Header.Get("Authorization"))
	}

	put, _ := http.NewRequest(http.MethodPut, "https://examplebucket.s3.amazonaws.com/test$file.text", strings.NewReader("Welcome to Amazon S3."))
	put.Header.Set("Date", "Fri, 24 May 2013 00:00:00 GMT")
	put.Header.Set("X-Amz-Storage-Class", "REDUCED_REDUNDANCY")
	if got := signature(t, put, "44ce7dd67c959e0d3524ffac1771dfbba87d2b6b4b4e99e42034a8b803f8b072"); got != "98ad721746da40c64f1a55b78f14c238d841ea1380cd77a1b5971af0ece108bd" {
		t.Errorf("PUT object signature = %s", got)
	}

	list, _ := http.NewRequest(http.MethodGet, "https://examplebucket.s3.amazonaws.com/?max-keys=2&prefix=J", nil)
	if got := signature(t, list, emptySHA256Hex); got != "34b48302e7b5fa45bde8084f4b7868a86f0a534bc59db6670ed5711ef69dc6f7" {
		t.Errorf("list objects signature = %s", got)
	}
}

func TestConfigFromEnv(t *testing.T) {
	ns := Namespace{LocalDir: "/tmp/x", Prefix: "economy/"}
	cfg, err := ConfigFromEnv(map[string]string{}, ns)
	if err != nil || cfg.Driver != DriverLocal || cfg.LocalDir != "/tmp/x" {
		t.Fatalf("default = %+v %v", cfg, err)
	}
	full := map[string]string{
		DriverEnv: "S3", S3EndpointEnv: "http://127.0.0.1:9000/", S3BucketEnv: "game", S3AccessKeyEnv: "a", S3SecretKeyEnv: "b",
		S3PrefixEnv: "/prod/", S3PublicBaseEnv: "https://cdn.example.com/game/",
	}
	cfg, err = ConfigFromEnv(full, ns)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Endpoint != "http://127.0.0.1:9000" || cfg.Region != "us-east-1" || !cfg.PathStyle || cfg.Prefix != "prod/economy/" || cfg.PublicBaseURL != "https://cdn.example.com/game" {
		t.Fatalf("s3 = %+v", cfg)
	}
	for k, v := range map[string]string{
		DriverEnv: "ftp", S3EndpointEnv: "127.0.0.1:9000", S3BucketEnv: "", S3AccessKeyEnv: "", S3PathStyleEnv: "maybe", S3PublicBaseEnv: "cdn",
	} {
		bad := map[string]string{}
		for a, b := range full {
			bad[a] = b
		}
		bad[k] = v
		if _, err := ConfigFromEnv(bad, ns); err == nil {
			t.Errorf("%s=%q accepted", k, v)
		}
	}
}

func TestObjectURLs(t *testing.T) {
	s := &s3Store{cfg: Config{Endpoint: "https://s3.example.com", Bucket: "b", Prefix: "economy/", PublicBaseURL: "https://cdn.example.com"}}
	s.cfg.PathStyle = true
	u, _ := s.objectURL("a b.png")
	if u.String() != "https://s3.example.com/b/economy/a%20b.png" {
		t.Errorf("path style = %s", u)
	}
	s.cfg.PathStyle = false
	u, _ = s.objectURL("x.png")
	if u.String() != "https://b.s3.example.com/economy/x.png" {
		t.Errorf("virtual host = %s", u)
	}
	if got := s.PublicURL("x.png"); got != "https://cdn.example.com/economy/x.png" {
		t.Errorf("public = %s", got)
	}
	for _, key := range []string{"", "../x", "a//b", "/abs", `a\b`} {
		if _, err := s.objectURL(key); err == nil {
			t.Errorf("key %q accepted", key)
		}
	}
}

func exercise(t *testing.T, store Store) {
	t.Helper()
	ctx := context.Background()
	raw := make([]byte, 8)
	_, _ = rand.Read(raw)
	key := "objectstore-test-" + hex.EncodeToString(raw) + ".png"
	data := bytes.Repeat([]byte{0x89, 'P', 'N', 'G'}, 300)
	if _, err := store.Get(ctx, key); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing get = %v", err)
	}
	if err := store.Put(ctx, key, data, "image/png"); err != nil {
		t.Fatal(err)
	}
	obj, err := store.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(obj.Body)
	obj.Body.Close()
	if !bytes.Equal(got, data) || obj.Size != int64(len(data)) {
		t.Fatalf("round trip: %d bytes, size %d", len(got), obj.Size)
	}
	if err := store.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(ctx, key); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted get = %v", err)
	}
	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("second delete = %v", err)
	}
}

func TestLocalStore(t *testing.T) {
	dir := t.TempDir()
	store, err := FromEnv(map[string]string{}, Namespace{LocalDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	exercise(t, store)
	if err := store.Put(context.Background(), "../escape.png", []byte("x"), ""); err == nil {
		t.Fatal("path escape accepted")
	}
	if _, err := store.Get(context.Background(), "../../etc/passwd"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("escape get = %v", err)
	}
}

func TestS3StoreAgainstRealServer(t *testing.T) {
	endpoint := os.Getenv("OBJECTSTORE_TEST_S3_ENDPOINT")
	if endpoint == "" {
		t.Skip("OBJECTSTORE_TEST_S3_ENDPOINT not set; S3 NOT_RUN")
	}
	env := map[string]string{
		DriverEnv: DriverS3, S3EndpointEnv: endpoint, S3BucketEnv: os.Getenv("OBJECTSTORE_TEST_S3_BUCKET"),
		S3AccessKeyEnv: os.Getenv("OBJECTSTORE_TEST_S3_ACCESS_KEY"), S3SecretKeyEnv: os.Getenv("OBJECTSTORE_TEST_S3_SECRET_KEY"),
	}
	store, err := FromEnv(env, Namespace{LocalDir: t.TempDir(), Prefix: "objectstore-test/"})
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("OBJECTSTORE_TEST_S3_CREATE_BUCKET") == "true" {
		cfg := store.(*s3Store).cfg
		req, _ := http.NewRequest(http.MethodPut, cfg.Endpoint+"/"+cfg.Bucket, nil)
		SignV4(req, emptySHA256Hex, cfg.AccessKey, cfg.SecretKey, cfg.Region, "s3", time.Now())
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
			t.Fatalf("create bucket = %d", resp.StatusCode)
		}
	}
	exercise(t, store)
	big := bytes.Repeat([]byte("multipart-"), 1300000)
	if err := PutStream(context.Background(), store, "stream-test.bin", bytes.NewReader(big), "application/octet-stream"); err != nil {
		t.Fatal(err)
	}
	obj, err := store.Get(context.Background(), "stream-test.bin")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(obj.Body)
	obj.Body.Close()
	_ = store.Delete(context.Background(), "stream-test.bin")
	if !bytes.Equal(got, big) {
		t.Fatalf("multipart round trip: %d of %d bytes", len(got), len(big))
	}
	bad := &s3Store{cfg: store.(*s3Store).cfg, client: http.DefaultClient, now: time.Now}
	bad.cfg.SecretKey = "wrong"
	if err := bad.Put(context.Background(), "denied.png", []byte("x"), "image/png"); err == nil || !strings.Contains(err.Error(), "SignatureDoesNotMatch") {
		t.Fatalf("wrong secret = %v", err)
	}
}
