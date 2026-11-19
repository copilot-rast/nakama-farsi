package handoffkit

import (
	"html"
	"io/fs"
	"regexp"
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
	b.prepare("", Connection{})
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

func TestShippedContent(t *testing.T) {
	b, err := Load(content.FS)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(content.FS, b, Options{ConsoleURL: "https://console.example", GeneratedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	h := string(out)
	if m := regexp.MustCompile(`href="[^"#:]*\.(md|html)(#[^"]*)?"`).FindAllString(h, 5); len(m) > 0 {
		t.Errorf("links to source files left: %v", m)
	}
	if m := regexp.MustCompile(`[0-9a-f]{64}`).FindString(h); m != "" {
		t.Errorf("a 64-hex key is in the file: %s…", m[:8])
	}
	if strings.Contains(h, ConsolePlaceholder) || strings.Contains(h, "{{SERVER_") {
		t.Error("placeholder left")
	}
	ids := map[string]bool{}
	for _, m := range regexp.MustCompile(` id="([^"]+)"`).FindAllStringSubmatch(h, -1) {
		ids[m[1]] = true
	}
	for _, m := range regexp.MustCompile(`href="#([^"]+)"`).FindAllStringSubmatch(h, -1) {
		if !ids[html.UnescapeString(m[1])] && !ids[m[1]] {
			t.Errorf("anchor #%s has no target", m[1])
		}
	}
}

const liveKey = "4f1c9a0e7b2d4c6e8a1f3b5d7c9e0a2b4d6f8a0c2e4b6d8f0a1c3e5b7d9f1a3c"

func TestSourcesKeepPlaceholdersOnly(t *testing.T) {
	hex := regexp.MustCompile(`[0-9a-f]{64}`)
	seen := false
	err := fs.WalkDir(content.FS, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !(strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".json")) {
			return err
		}
		raw, err := fs.ReadFile(content.FS, name)
		if err != nil {
			return err
		}
		if hex.Match(raw) {
			t.Errorf("%s holds a 64-hex key", name)
		}
		if strings.Contains(string(raw), DefaultServerKey) && !strings.HasSuffix(name, "PROGRESS.md") {
			t.Errorf("%s holds the default key", name)
		}
		seen = seen || strings.Contains(string(raw), KeyPlaceholder)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !seen {
		t.Error("no source uses the server key placeholder")
	}
}

func renderShipped(t *testing.T, c Connection) string {
	t.Helper()
	b, err := Load(content.FS)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(content.FS, b, Options{ConsoleURL: "https://console.example", GeneratedAt: time.Now(), Connection: c})
	if err != nil {
		t.Fatal(err)
	}
	h := string(out)
	if strings.Contains(h, "{{SERVER_") {
		t.Error("placeholder left")
	}
	return h
}

func TestRenderShowsLiveKey(t *testing.T) {
	set := DefaultSettings()
	c := Resolve(Server{ServerKey: liveKey, Env: map[string]string{EnvPublicHost: "api.game.test"}}, set)
	h := renderShipped(t, c)
	for _, want := range []string{
		`data-connection data-key-shown="true" data-host-set="true"`,
		`<td data-server-key>`,
		"new Client(&#39;" + liveKey + "&#39;, &#39;api.game.test&#39;, &#39;443&#39;, true)",
		"<code>https://api.game.test</code>",
		`<a href="#connection">`,
	} {
		if !strings.Contains(h, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(h, "data-default-key") || strings.Contains(h, codeAskKey) {
		t.Error("default-key warning or ask marker with a live key")
	}
}

func TestRenderHidesKeyWhenOff(t *testing.T) {
	set := DefaultSettings()
	set.ShowServerKey = false
	h := renderShipped(t, Resolve(Server{ServerKey: liveKey}, set))
	if strings.Contains(h, liveKey) || strings.Contains(h, liveKey[:16]) {
		t.Fatal("key in the file while showServerKey is off")
	}
	for _, want := range []string{`data-key-hidden`, codeAskKey, codeAskHost, "data-host-missing", "HANDOFF_PUBLIC_HOST", textAskKey} {
		if !strings.Contains(h, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestRenderWarnsOnDefaultKey(t *testing.T) {
	h := renderShipped(t, Resolve(Server{ServerKey: DefaultServerKey}, DefaultSettings()))
	if !strings.Contains(h, "data-default-key") || !strings.Contains(h, "<code>defaultkey</code>") {
		t.Fatal("default key warning missing")
	}
	set := DefaultSettings()
	set.ShowServerKey = false
	if strings.Contains(renderShipped(t, Resolve(Server{ServerKey: DefaultServerKey}, set)), "data-default-key") {
		t.Fatal("default key named while hidden")
	}
}

func TestResolve(t *testing.T) {
	cases := []struct {
		env  map[string]string
		url  string
		host string
		port int
		ssl  bool
	}{
		{map[string]string{EnvPublicHost: "api.game.test"}, "https://api.game.test", "api.game.test", 443, true},
		{map[string]string{EnvPublicHost: "api.game.test", EnvPublicSSL: "false"}, "http://api.game.test", "api.game.test", 80, false},
		{map[string]string{EnvPublicHost: "api.game.test", EnvPublicPort: "7350", EnvPublicSSL: "0"}, "http://api.game.test:7350", "api.game.test", 7350, false},
		{map[string]string{EnvPublicHost: "https://api.game.test:8443/"}, "https://api.game.test:8443", "api.game.test", 8443, true},
		{map[string]string{EnvPublicBase: "http://100.106.1.79:7450"}, "http://100.106.1.79:7450", "100.106.1.79", 7450, false},
		{map[string]string{EnvPublicBase: "https://api.example.com"}, "", "", 0, false},
		{map[string]string{EnvPublicHost: "bad host<x>"}, "", "", 0, false},
	}
	for _, c := range cases {
		got := Resolve(Server{ServerKey: "k", Env: c.env}, DefaultSettings())
		if got.URL != c.url || got.Host != c.host || got.Port != c.port || got.SSL != c.ssl || got.HostSet != (c.host != "") {
			t.Errorf("%v: got %+v", c.env, got)
		}
		if got.ServerKey != "k" || !got.KeyShown {
			t.Errorf("%v: key not shown", c.env)
		}
	}
	off := DefaultSettings()
	off.ShowServerKey = false
	if c := Resolve(Server{ServerKey: "k"}, off); c.ServerKey != "" || c.KeyShown {
		t.Fatal("key kept while off")
	}
}

func TestApplyEscapes(t *testing.T) {
	c := Connection{ServerKey: `a"<b>`, KeyShown: true}
	out := c.Apply(`<p>{{SERVER_KEY}} <code>{{SERVER_HOST}}</code> {{SERVER_HOST}}</p>`)
	if !strings.Contains(out, "a&#34;&lt;b&gt;") || !strings.Contains(out, "<code>"+codeAskHost+"</code>") || !strings.Contains(out, textAskHost) {
		t.Fatal(out)
	}
}
