package contentimport

import (
	"context"
	"testing"
	"time"
)

func TestRegisterImportAwait(t *testing.T) {
	ctx := context.Background()
	if _, err := Import(ctx, "t", nil, ""); err != ErrUnknown {
		t.Fatalf("want ErrUnknown, got %v", err)
	}
	go func() {
		time.Sleep(300 * time.Millisecond)
		Register("t", func(_ context.Context, doc []byte, by string) (map[string]any, error) {
			return map[string]any{"len": len(doc), "by": by}, nil
		})
	}()
	if !Await(ctx, "t", 3*time.Second) {
		t.Fatal("importer not seen")
	}
	out, err := Import(ctx, "t", []byte("{}"), "seed")
	if err != nil || out["len"] != 2 || out["by"] != "seed" {
		t.Fatalf("%v %v", out, err)
	}
	if len(Names()) != 1 {
		t.Fatal(Names())
	}
	Register("t", nil)
	if Await(ctx, "t", 10*time.Millisecond) {
		t.Fatal("removed importer still seen")
	}
}
