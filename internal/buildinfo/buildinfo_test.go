package buildinfo

import (
	"encoding/base64"
	"testing"
)

func TestFrom(t *testing.T) {
	up := base64.StdEncoding.EncodeToString([]byte(`{"repo":"heroiclabs/nakama","branch":"master","commit":"37618AD4","date":"2026-11-12T15:32:42Z"}`))
	info := From("abc123", "dev", "2026-11-13T08:00:00Z", "true", "191", up)
	if !info.Known || !info.Dirty || info.OwnAhead != 191 || info.Branch != "dev" || info.Date != "2026-11-13T08:00:00Z" {
		t.Fatalf("%+v", info)
	}
	if info.Parent == nil || info.Parent.Repo != "heroiclabs/nakama" || info.Parent.Commit != "37618ad4" || info.Parent.Provider != "github" {
		t.Fatalf("parent %+v", info.Parent)
	}
	empty := From("", "", "not a date", "", "x", "%%%")
	if empty.Known || empty.Date != "" || empty.Parent != nil || empty.OwnAhead != 0 {
		t.Fatalf("%+v", empty)
	}
	if p := ParseParent(`{"repo":"a/b","commit":"c"}`); p == nil || p.Repo != "a/b" {
		t.Fatal("plain JSON not read")
	}
	if ParseParent(`{"repo":"a/b"}`) != nil {
		t.Fatal("parent without a commit accepted")
	}
}
