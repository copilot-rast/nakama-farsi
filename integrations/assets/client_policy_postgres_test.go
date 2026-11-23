package main

import (
	"context"
	"testing"

	"github.com/heroiclabs/nakama/v3/integrations/assets/internal/settings"
)

func TestManifestCarriesTheOnlinePolicy(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	folder := f.prefix + "/Cars"
	m, err := f.s.rpcManifest(ctx, "", `{"folder":"`+folder+`"}`)
	if err != nil {
		t.Fatal(err)
	}
	c := m.(map[string]any)["client"].(map[string]any)
	if c["requireOnline"] != false || c["maxCacheAgeHours"] != 0 || m.(map[string]any)["serverTime"] == "" {
		t.Fatalf("default policy = %v", m)
	}
	f.settings(func(a *settings.All) { a.Client = settings.Client{RequireOnline: true, MaxCacheAgeHours: 48} })
	settings.Invalidate()
	m, _ = f.s.rpcManifest(ctx, "", `{"folder":"`+folder+`"}`)
	tag := m.(map[string]any)["etag"].(string)
	nm, _ := f.s.rpcManifest(ctx, "", `{"folder":"`+folder+`","etag":`+jsonString(tag)+`}`)
	for _, out := range []any{m, nm} {
		c := out.(map[string]any)["client"].(map[string]any)
		if c["requireOnline"] != true || c["maxCacheAgeHours"] != 48 {
			t.Fatalf("policy = %v", out)
		}
	}
	bad := settings.Defaults()
	bad.Client.MaxCacheAgeHours = -1
	if field, code := bad.Validate(); field != "client.maxCacheAgeHours" || code != "VALUE_OUT_OF_RANGE" {
		t.Fatalf("validate = %s %s", field, code)
	}
}
