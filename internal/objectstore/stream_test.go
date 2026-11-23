package objectstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"math/rand"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama/v3/internal/objectstore/s3test"
	"time"
)

func fakeStore(t *testing.T, fake *s3test.Server) *s3Store {
	t.Helper()
	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)
	store, err := New(Config{Driver: DriverS3, Endpoint: srv.URL, Bucket: "b", Region: "us-east-1", AccessKey: "a", SecretKey: "s", PathStyle: true, Prefix: "content/", MultipartBytes: MinMultipartBytes}, nil)
	if err != nil {
		t.Fatal(err)
	}
	return store.(*s3Store)
}

func randomBytes(n int) []byte {
	out := make([]byte, n)
	rand.New(rand.NewSource(int64(n))).Read(out)
	return out
}

type failingReader struct {
	r     io.Reader
	after int
	read  int
}

func (f *failingReader) Read(p []byte) (int, error) {
	if f.read >= f.after {
		return 0, errors.New("source broke")
	}
	if len(p) > f.after-f.read {
		p = p[:f.after-f.read]
	}
	n, err := f.r.Read(p)
	f.read += n
	return n, err
}

func TestS3PutStreamMultipart(t *testing.T) {
	fake := s3test.New()
	store := fakeStore(t, fake)
	ctx := context.Background()
	data := randomBytes(21<<20 + 12345)
	if err := PutStream(ctx, store, "packs/blobs/big", io.MultiReader(bytes.NewReader(data[:3<<20]), bytes.NewReader(data[3<<20:])), "application/octet-stream"); err != nil {
		t.Fatal(err)
	}
	if fake.Completed != 1 || fake.Puts != 0 {
		t.Fatalf("completed=%d puts=%d", fake.Completed, fake.Puts)
	}
	sizes := append([]int(nil), fake.PartSizes...)
	if len(sizes) != 5 {
		t.Fatalf("parts = %v", sizes)
	}
	for _, n := range sizes[:len(sizes)-1] {
		if n != MinMultipartBytes {
			t.Fatalf("part sizes = %v", sizes)
		}
	}
	obj, err := store.Get(ctx, "packs/blobs/big")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(obj.Body)
	obj.Body.Close()
	if !bytes.Equal(got, data) {
		t.Fatalf("round trip %d != %d bytes", len(got), len(data))
	}
}

func TestS3PutStreamExactMultipleAndSmall(t *testing.T) {
	fake := s3test.New()
	store := fakeStore(t, fake)
	ctx := context.Background()
	exact := randomBytes(2 * MinMultipartBytes)
	if err := store.PutStream(ctx, "exact", bytes.NewReader(exact), ""); err != nil {
		t.Fatal(err)
	}
	if fake.Completed != 1 || len(fake.PartSizes) != 2 {
		t.Fatalf("exact: completed=%d parts=%v", fake.Completed, fake.PartSizes)
	}
	small := randomBytes(1000)
	if err := store.PutStream(ctx, "small", bytes.NewReader(small), "text/plain"); err != nil {
		t.Fatal(err)
	}
	if fake.Puts != 1 || !bytes.Equal(fake.Objects["/b/content/small"], small) {
		t.Fatalf("small: puts=%d", fake.Puts)
	}
	if err := store.PutStream(ctx, "empty", bytes.NewReader(nil), ""); err != nil {
		t.Fatal(err)
	}
	if data, ok := fake.Objects["/b/content/empty"]; !ok || len(data) != 0 {
		t.Fatal("empty object not written")
	}
}

func TestS3PutStreamAbortsOnReadError(t *testing.T) {
	fake := s3test.New()
	store := fakeStore(t, fake)
	data := randomBytes(12 << 20)
	err := store.PutStream(context.Background(), "broken", &failingReader{r: bytes.NewReader(data), after: 7 << 20}, "")
	if err == nil || !strings.Contains(err.Error(), "source broke") {
		t.Fatalf("err = %v", err)
	}
	if fake.Aborted != 1 || len(fake.Uploads) != 0 {
		t.Fatalf("aborted=%d open=%d", fake.Aborted, len(fake.Uploads))
	}
	if _, ok := fake.Objects["/b/content/broken"]; ok {
		t.Fatal("broken object written")
	}
}

func TestS3PutStreamRetriesPart(t *testing.T) {
	fake := s3test.New()
	fake.FailPart = 2
	store := fakeStore(t, fake)
	data := randomBytes(11 << 20)
	start := time.Now()
	if err := store.PutStream(context.Background(), "retry", bytes.NewReader(data), ""); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(fake.Objects["/b/content/retry"], data) || time.Since(start) > 10*time.Second {
		t.Fatal("retry did not store the object")
	}
}

func TestLocalPutStreamLarge(t *testing.T) {
	dir := t.TempDir()
	store, err := New(Config{Driver: DriverLocal, LocalDir: dir}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	data := randomBytes(24<<20 + 7)
	h := sha256.New()
	if err := PutStream(ctx, store, "packs/blobs/x", io.TeeReader(bytes.NewReader(data), h), ""); err != nil {
		t.Fatal(err)
	}
	obj, err := store.Get(ctx, "packs/blobs/x")
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(obj.Body)
	obj.Body.Close()
	want := sha256.Sum256(data)
	if !bytes.Equal(got, data) || hex.EncodeToString(h.Sum(nil)) != hex.EncodeToString(want[:]) {
		t.Fatal("local round trip failed")
	}
	if err := PutStream(ctx, store, "packs/blobs/y", &failingReader{r: bytes.NewReader(data), after: 1 << 20}, ""); err == nil {
		t.Fatal("broken stream accepted")
	}
	if _, err := store.Get(ctx, "packs/blobs/y"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("broken stream left %v", err)
	}
	_ = filepath.WalkDir(dir, func(p string, _ fs.DirEntry, _ error) error {
		if strings.HasSuffix(p, ".tmp") {
			t.Errorf("temp file left: %s", p)
		}
		return nil
	})
}
