package main

import "testing"

func TestVersionsAndRanges(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want int
	}{{"1.10", "1.9", 1}, {"2", "2.0", 0}, {"1.0.1", "1.0", 1}, {"1.0-beta", "1.0", -1}, {"1.2.3", "1.2.3", 0}} {
		if got := compareVersions(c.a, c.b); got != c.want {
			t.Fatalf("%s vs %s = %d", c.a, c.b, got)
		}
	}
	if !inRange("1.5", "1.0", "2.0") || inRange("2.1", "1.0", "2.0") || inRange("0.9", "1.0", "") || !inRange("", "1.0", "1.1") {
		t.Fatal("range")
	}
	for in, want := range map[string]string{`Android\a\.\b.bundle`: "a/b.bundle", "/x/y": "x/y", "Android/catalog.json": "catalog.json"} {
		if got, ok := cleanPackPath(in, "Android"); !ok || got != want {
			t.Fatalf("%q = %q %v", in, got, ok)
		}
	}
	for _, bad := range []string{"../a", "a/../../b", "", "/"} {
		if _, ok := cleanPackPath(bad, "Android"); ok {
			t.Fatalf("%q accepted", bad)
		}
	}
	for in, want := range map[string]string{`\Cars\BMW/`: "Cars/BMW", "": "", "/": ""} {
		if got, ok := cleanFolder(in); !ok || got != want {
			t.Fatalf("folder %q = %q", in, got)
		}
	}
	if _, ok := cleanFolder("a//b"); ok {
		t.Fatal("empty segment")
	}
}
