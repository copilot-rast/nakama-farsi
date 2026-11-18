package handoffkit

import (
	"strings"
	"testing"
	"testing/fstest"
	"time"

	content "github.com/heroiclabs/nakama/v3/docs/handoff"
)

func TestJalali(t *testing.T) {
	y, m, d := ToJalali(2026, 11, 18)
	if y != 1405 || m != 8 || d != 27 {
		t.Fatalf("got %d-%d-%d", y, m, d)
	}
	y, m, d = ToJalali(2025, 3, 21)
	if y != 1404 || m != 1 || d != 1 {
		t.Fatalf("nowruz: got %d-%d-%d", y, m, d)
	}
}

func TestMarkdown(t *testing.T) {
	src := "# Title\n\nIntro with `code <x>` and **bold** and [next](02-b.md#part-two).\n\n* one\n  * nested `a|b`\n* two\n\n  more of two\n\n1. first\n2. second\n\n| a | b |\n| --- | --- |\n| `x\\|y` | <b> |\n\n```json\n{\"k\": 1}\n```\n\n## Part two\n\n- [x] done\n"
	files := fstest.MapFS{
		"general/handoff.json": {Data: []byte(`{"title":"G","changelog":[]}`)},
		"general/01-a.md":      {Data: []byte(src)},
		"general/02-b.md":      {Data: []byte("---\ntitle: B\n---\n## Part two\ntext\n")},
	}
	b, err := Load(files)
	if err != nil {
		t.Fatal(err)
	}
	b.prepare("")
	s := b.General.Sections[0]
	if s.Title != "Title" {
		t.Fatalf("title %q", s.Title)
	}
	for _, want := range []string{
		"<code>code &lt;x&gt;</code>",
		"<b>bold</b>",
		`<a href="#g-02-b-part-two">next</a>`,
		"<li>nested <code>a|b</code></li>",
		"<p>more of two</p>",
		"<ol>\n<li>first</li>",
		"<td><code>x|y</code></td><td>&lt;b&gt;</td>",
		`<span class="lang">JSON</span>`,
		`<h3 id="g-01-a-part-two">Part two</h3>`,
		`<input type="checkbox" disabled checked> done`,
	} {
		if !strings.Contains(s.body, want) {
			t.Errorf("missing %q in\n%s", want, s.body)
		}
	}
}

func TestBundleAndChangelog(t *testing.T) {
	files := fstest.MapFS{
		"general/handoff.json":  {Data: []byte(`{"title":"G","changelog":[{"date":"2026-11-10","title":"old general","section":"01-a.md"},{"date":"2026-11-20","title":"new general"}]}`)},
		"general/01-a.md":       {Data: []byte("# A\ntext\n")},
		"project/handoff.json":  {Data: []byte(`{"name":"Demo Game","title":"دمو","changelog":[{"date":"2026-11-15","title":"project entry","section":"guide/01.html"}]}`)},
		"project/guide/01.html": {Data: []byte("---\ntitle: P\n---\n<header><h1>x</h1></header>\n<p><a href=\"{{CONSOLE_URL}}/#/system/sdk\">sdk</a> <a href=\"../01-a.md\">a</a> <a href=\"https://e.x\">e</a></p>\n<pre><code>var a = 1;</code></pre>\n")},
	}
	b, err := Load(files)
	if err != nil {
		t.Fatal(err)
	}
	log := b.Changelog()
	if len(log) != 3 || log[0].Title != "new general" || log[1].Scope != ScopeProject || log[2].Anchor != "g-01-a" {
		t.Fatalf("changelog %+v", log)
	}
	out, err := Render(content.FS, b, Options{ConsoleURL: "https://console.example", GeneratedAt: time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	h := string(out)
	for _, want := range []string{`id="part-general"`, `id="part-project"`, "بخش اختصاصی: دمو", `id="p-guide-01"`, `href="#p-guide-01"`, "<span class=\"lang\">C#</span>", "https://console.example/#/system/sdk", `<a href="https://console.example/#/system/sdk">sdk</a>`, "@font-face", "۲۷ آبان ۱۴۰۵"} {
		if !strings.Contains(h, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(h, "<header><h1>x</h1>") {
		t.Error("fragment header kept")
	}
	if FileName(b, time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC)) != "handoff-demo-game-2026-11-18.html" {
		t.Error(FileName(b, time.Now()))
	}
}

func TestNoProject(t *testing.T) {
	b, err := Load(content.FS)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(content.FS, b, Options{GeneratedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if b.Project == nil && !strings.Contains(string(out), "data-no-project") {
		t.Error("no-project note missing")
	}
	if len(b.General.Sections) == 0 || len(b.Changelog()) == 0 {
		t.Error("general part is empty")
	}
}

func TestBadChangelog(t *testing.T) {
	for _, raw := range []string{
		`{"changelog":[{"date":"18-11-2026","title":"x"}]}`,
		`{"changelog":[{"date":"2026-11-18","title":""}]}`,
		`{"changelog":[{"date":"2026-11-18","title":"x","section":"missing.md"}]}`,
	} {
		if _, err := Load(fstest.MapFS{"general/handoff.json": {Data: []byte(raw)}}); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
}

func TestSettings(t *testing.T) {
	if !DefaultSettings().Allows(false) {
		t.Fatal("default must allow every console user")
	}
	if (Settings{Enabled: true, Audience: AudienceAdmins}).Allows(false) {
		t.Fatal("admins only let a non-admin in")
	}
	if (Settings{Enabled: false, Audience: AudienceAll}).Allows(true) {
		t.Fatal("disabled still allowed")
	}
	if _, ok := (Settings{Audience: "everyone"}).Clean(); ok {
		t.Fatal("bad audience accepted")
	}
}
