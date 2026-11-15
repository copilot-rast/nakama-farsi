package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitizeRichHTMLAllowlist(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{`<p style="text-align: center; color: red" onclick="x()">Hi <b>there</b> <i>you</i></p>`, `<p style="text-align: center">Hi <strong>there</strong> <em>you</em></p>`},
		{`<script>alert(1)</script><p>ok</p><style>p{}</style><iframe src="x"></iframe>`, `<p>ok</p>`},
		{`<a href="javascript:alert(1)">x</a>`, `<p><a target="_blank" rel="noopener noreferrer nofollow">x</a></p>`},
		{`<a href="https://e.com/a?b=1&c=2" onmouseover="x">y</a>`, `<p><a href="https://e.com/a?b=1&amp;c=2" target="_blank" rel="noopener noreferrer nofollow">y</a></p>`},
		{`<a href="mailto:a@b.c">m</a>`, `<p><a href="mailto:a@b.c" target="_blank" rel="noopener noreferrer nofollow">m</a></p>`},
		{`plain <u>under</u> <s>gone</s> <strike>old</strike>`, `<p>plain <u>under</u> <s>gone</s> <s>old</s></p>`},
		{`<h2 style="text-align:right">T</h2><h7>x</h7>`, `<h2 style="text-align: right">T</h2><p>x</p>`},
		{`<table><tbody><tr><th colspan="2" style="x">A</th></tr><tr><td rowspan="999">B</td><td colwidth="120">C</td></tr></tbody></table>`,
			`<table><tbody><tr><th colspan="2">A</th></tr><tr><td>B</td><td colwidth="120">C</td></tr></tbody></table>`},
		{`<ul data-type="taskList"><li data-type="taskItem" data-checked="true"><label><input type="checkbox" checked><span></span></label><div><p>Done</p></div></li></ul>`,
			`<ul data-type="taskList"><li data-type="taskItem" data-checked="true"><label><input type="checkbox" checked="checked" disabled="disabled"><span></span></label><div><p>Done</p></div></li></ul>`},
		{`<input type="text" value="x"><img src="https://e.com/a.png" onerror="x">`, ``},
		{`<div><section><p>Deep</p></section></div>`, `<div><p>Deep</p></div>`},
		{`<blockquote>q</blockquote><pre><code>a &lt; b</code></pre><hr>`, `<blockquote>q</blockquote><pre><code>a &lt; b</code></pre><hr>`},
		{`<svg><script>x</script></svg><p>a</p><!-- c -->`, `<p>a</p>`},
		{`<ol start="3"><li>x</li></ol>`, `<ol start="3"><li>x</li></ol>`},
		{`<p></p>`, ``},
	} {
		if got := sanitizeRichHTML(tc.in); got != tc.want {
			t.Errorf("sanitize(%q)\n got  %q\n want %q", tc.in, got, tc.want)
		}
	}
	if again := sanitizeRichHTML(sanitizeRichHTML(`<p style="text-align: justify">x <a href="/news/1">y</a></p>`)); again != `<p style="text-align: justify">x <a href="/news/1" target="_blank" rel="noopener noreferrer nofollow">y</a></p>` {
		t.Fatalf("not idempotent: %q", again)
	}
}

func TestRichTextAndMarkdownConversion(t *testing.T) {
	if got := richText(`<h2>Title</h2><p>a <b>b</b></p><ul><li>one</li><li>two</li></ul>`); got != "Title\na b\none\ntwo" {
		t.Fatalf("richText = %q", got)
	}
	if got := sanitizeRichHTML(renderMarkdown("## Head\n\n**bold** [l](https://e.com)\n\n- a\n- b")); got != `<h2>Head</h2><p><strong>bold</strong> <a href="https://e.com" target="_blank" rel="noopener noreferrer nofollow">l</a></p><ul><li>a</li><li>b</li></ul>` {
		t.Fatalf("markdown = %q", got)
	}
	if got := textToRich("line <1>\nline 2\n\npara"); got != `<p>line &lt;1&gt;<br>line 2</p><p>para</p>` {
		t.Fatalf("text = %q", got)
	}
	if !looksLikeHTML("<p>x</p>") || looksLikeHTML("a < b and c > d") {
		t.Fatal("looksLikeHTML")
	}
}

func TestJSONPathLookup(t *testing.T) {
	var doc any
	_ = json.Unmarshal([]byte(`{"data":{"items":[{"title":"A","tags":["x"]},{"title":"B"}]},"meta-info":{"lang":"en"}}`), &doc)
	for path, want := range map[string]any{"data.items[1].title": "B", "$.data.items[0].title": "A", "meta-info.lang": "en", "data.items[0].tags[0]": "x"} {
		if got, ok := lookupJSONPath(doc, path); !ok || got != want {
			t.Errorf("%s = %v %v", path, got, ok)
		}
	}
	for _, path := range []string{"data.items[5].title", "data.nope", "data.items.title"} {
		if _, ok := lookupJSONPath(doc, path); ok {
			t.Errorf("%s should be missing", path)
		}
	}
	for _, bad := range []string{"a..b", "a[x]", "a b", "[", strings.Repeat("a.", 120)} {
		if validJSONPath(bad) {
			t.Errorf("%q should be invalid", bad)
		}
	}
	if !validJSONPath("") || !validJSONPath("$") || !validJSONPath("items[0]") {
		t.Fatal("valid paths refused")
	}
}

func TestMapIngestItem(t *testing.T) {
	src := newsSource{BodyFormat: "auto", Mapping: map[string]string{}}
	for k, v := range defaultMapping {
		src.Mapping[k] = v
	}
	src.Mapping["cover"] = "media.cover"
	var item any
	dec := json.NewDecoder(strings.NewReader(`{"id":42,"title":"<b>Big</b> news","summary":"short","body":"<p onclick=1>Hello <script>x</script>world</p>","media":{"cover":"https://e.com/c.png"},
		"tags":"a, b","publishedAt":1700000000,"expiresAt":"bad","url":"javascript:x","lang":"en-US"}`))
	dec.UseNumber()
	_ = dec.Decode(&item)
	m, notes, err := mapIngestItem(src, item)
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "Big news" || m.Body != "<p>Hello world</p>" || m.Cover != "https://e.com/c.png" || m.ExternalID != "42" || len(m.Tags) != 2 || m.Tags[1] != "b" || m.Lang != "en-US" {
		t.Fatalf("mapped = %+v", m)
	}
	if m.PublishedAt == nil || m.PublishedAt.Unix() != 1700000000 || m.URL != "" {
		t.Fatalf("time/url = %+v", m)
	}
	if strings.Join(notes, ",") != "SOURCE_URL_IGNORED,EXPIRES_AT_IGNORED" {
		t.Fatalf("notes = %v", notes)
	}
	if _, _, err := mapIngestItem(src, map[string]any{"body": "x"}); err == nil {
		t.Fatal("title required")
	}
	m2, _, _ := mapIngestItem(src, map[string]any{"title": "T", "body": "**md** text"})
	if m2.Body != "<p><strong>md</strong> text</p>" {
		t.Fatalf("markdown body = %q", m2.Body)
	}
	src.BodyFormat = "text"
	m3, _, _ := mapIngestItem(src, map[string]any{"title": "T", "body": "<b>raw</b>"})
	if m3.Body != "<p>&lt;b&gt;raw&lt;/b&gt;</p>" || m3.hash() == m2.hash() {
		t.Fatalf("text body = %q", m3.Body)
	}
}

func TestSignIngestMatchesDocumentedFormat(t *testing.T) {
	if got := signIngest("secret", "1700000000", []byte(`{"a":1}`)); got != "49f24e537407743fa4a0242bb63b94b9a47ee99cbbe071ccd8a22550ae411686" {
		t.Fatalf("sign = %q", got)
	}
	if signIngest("secret", "1", []byte("x")) == signIngest("secret", "2", []byte("x")) {
		t.Fatal("timestamp must be signed")
	}
}

func TestSlugify(t *testing.T) {
	for _, tc := range []struct{ in, style, want string }{
		{"خبر  مهم: به‌روزرسانی ۲!", "native", "خبر-مهم-به-روزرسانی-2"},
		{"Hello, World — 2026", "native", "hello-world-2026"},
		{"خبر مهم شب یلدا", "latin", "khbr-mhm-shb-ilda"},
		{"Big NEWS!", "latin", "big-news"},
		{"anything", "id", ""},
		{"!!!", "native", ""},
	} {
		if got := slugify(tc.in, tc.style); got != tc.want {
			t.Errorf("slugify(%q,%s) = %q want %q", tc.in, tc.style, got, tc.want)
		}
	}
	if got := slugify(strings.Repeat("کلمه ", 40), "native"); len([]rune(got)) > 80 || strings.HasSuffix(got, "-") {
		t.Fatalf("long slug = %q", got)
	}
	if s, err := cleanSlug("  Yalda-شب "); err != nil || s != "yalda-شب" {
		t.Fatalf("cleanSlug = %q %v", s, err)
	}
	for _, bad := range []string{"a b", "a--b", "-a", "a/b", "<x>"} {
		if _, err := cleanSlug(bad); err == nil {
			t.Errorf("%q should be refused", bad)
		}
	}
}
