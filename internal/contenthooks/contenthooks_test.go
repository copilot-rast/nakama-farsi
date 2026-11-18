package contenthooks

import (
	"context"
	"errors"
	"testing"
)

func TestCatalogWithoutAndWithReader(t *testing.T) {
	SetCatalog(nil)
	if _, err := Catalog(context.Background(), "Web", ""); !errors.Is(err, ErrUnavailable) || Available() {
		t.Fatalf("want ErrUnavailable, got %v", err)
	}
	SetCatalog(func(_ context.Context, platform, appVersion string) (map[string]any, error) {
		return map[string]any{"platform": platform, "appVersion": appVersion}, nil
	})
	defer SetCatalog(nil)
	out, err := Catalog(context.Background(), "Web", "1.0.0")
	if err != nil || out["platform"] != "Web" || out["appVersion"] != "1.0.0" || !Available() {
		t.Fatalf("got %v %v", out, err)
	}
}
