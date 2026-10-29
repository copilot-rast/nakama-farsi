package main

import (
	"strings"
	"testing"
)

func TestRenderMarkdownSubset(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"hello", "<p>hello</p>"},
		{"**bold** and *it*", "<p><strong>bold</strong> and <em>it</em></p>"},
		{"## Title\n\ntext", "<h2>Title</h2><p>text</p>"},
		{"### Sub", "<h3>Sub</h3>"},
		{"- a\n- b\n\n1. x\n2. y", "<ul><li>a</li><li>b</li></ul><ol><li>x</li><li>y</li></ol>"},
		{"line1\nline2", "<p>line1<br>line2</p>"},
		{"[site](https://example.com/a?b=1&c=2)", `<p><a href="https://example.com/a?b=1&amp;c=2" rel="noopener noreferrer" target="_blank">site</a></p>`},
		{"![cover](/v2/economy/files/x.png)", `<p><img src="/v2/economy/files/x.png" alt="cover"></p>`},
		{"[**bold link**](https://e.com)", `<p><a href="https://e.com" rel="noopener noreferrer" target="_blank"><strong>bold link</strong></a></p>`},
		{"سلام **دنیا**", "<p>سلام <strong>دنیا</strong></p>"},
		{"2 * 3 * 4", "<p>2 * 3 * 4</p>"},
		{"a ** b", "<p>a ** b</p>"},
		{"wow!", "<p>wow!</p>"},
	} {
		if got := renderMarkdown(tc.in); got != tc.want {
			t.Errorf("renderMarkdown(%q)\n got %s\nwant %s", tc.in, got, tc.want)
		}
	}
}

func TestRenderMarkdownEscapesHTML(t *testing.T) {
	for _, in := range []string{
		`<script>alert(1)</script>`,
		`<img src=x onerror=alert(1)>`,
		`[x](javascript:alert(1))`,
		`[x](JaVaScRiPt:alert(1))`,
		`[x](data:text/html;base64,PHNjcmlwdD4=)`,
		`![x](javascript:alert(1))`,
		`[x](https://e.com/"onmouseover="alert(1))`,
		`![a" onerror="alert(1)](https://e.com/i.png)`,
		`[x](//evil.com)`,
		`**<b>**`,
		`## <iframe src=x>`,
		`- <svg onload=alert(1)>`,
	} {
		out := renderMarkdown(in)
		lower := strings.ToLower(out)
		for _, bad := range []string{"<script", "<img src=x", `href="javascript`, `src="javascript`, `href="data`, `src="data`, "<iframe", "<svg", `" onerror=`, `"onmouseover=`, `href="//`, "<b>", "<a "} {
			if strings.Contains(lower, bad) {
				t.Errorf("renderMarkdown(%q) = %s contains %q", in, out, bad)
			}
		}
	}
	if got := renderMarkdown(`![a" onerror="alert(1)](https://e.com/i.png)`); !strings.Contains(got, `alt="a&#34; onerror=&#34;alert(1)"`) {
		t.Errorf("alt text must be attribute-escaped: %s", got)
	}
}

func TestMarkdownText(t *testing.T) {
	got := markdownText("## Hi\n\n**bold** [site](https://e.com)\n\n- a\n- b")
	want := "Hi\n\nbold site (https://e.com)\n\n- a\n- b"
	if got != want {
		t.Errorf("markdownText = %q, want %q", got, want)
	}
}
