package handoffkit

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"io/fs"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const ConsolePlaceholder = "{{CONSOLE_URL}}"

type Options struct {
	ConsoleURL  string
	Commit      string
	GeneratedAt time.Time
}

var (
	hrefRe   = regexp.MustCompile(`href="([^"]*)"`)
	headerRe = regexp.MustCompile(`(?s)^\s*<header>.*?</header>\s*`)
	preRe    = regexp.MustCompile(`(?s)<pre(\s[^>]*)?><code>(.*?)</code></pre>`)
	csRe     = regexp.MustCompile(`(?m)^\s*(using |public |private |var |await |if \(|[A-Z]\w*(\.\w+)+\()`)
	shellRe  = regexp.MustCompile(`^(git|ssh|cat|type|curl|\$)\b`)
)

func guessLang(code string) string {
	c := strings.TrimSpace(code)
	switch {
	case strings.HasPrefix(c, "{") || strings.HasPrefix(c, "[") || strings.HasPrefix(c, "\""):
		return "JSON"
	case strings.HasPrefix(c, "<"):
		return "XML"
	case csRe.MatchString(c) || strings.HasSuffix(c, ";") || strings.HasSuffix(c, "}"):
		return "C#"
	case shellRe.MatchString(c):
		return "Shell"
	case strings.HasPrefix(c, "http://") || strings.HasPrefix(c, "https://"):
		return "URL"
	case strings.Contains(c, "/") && !strings.Contains(c, " "):
		return "Path"
	}
	return "Text"
}

func FileName(b *Bundle, at time.Time) string {
	name := "general"
	if b.Project != nil && b.Project.Meta.Name != "" {
		name = idCleanRe.ReplaceAllString(strings.ToLower(b.Project.Meta.Name), "-")
	}
	return "handoff-" + name + "-" + at.UTC().Format("2006-01-02") + ".html"
}

func (b *Bundle) resolver(from *Section) linkResolver {
	return func(target string) (string, bool) {
		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "mailto:") {
			return "", false
		}
		file, frag, _ := strings.Cut(target, "#")
		if f, err := url.PathUnescape(frag); err == nil {
			frag = f
		}
		if file == "" {
			if frag == "" {
				return "", false
			}
			return "#" + from.ID + "-" + frag, true
		}
		p := path.Clean(path.Join(path.Dir(from.Path), file))
		var hit *Section
		for _, s := range b.Sections() {
			if s.Path == p && (hit == nil || s.Scope == from.Scope) {
				hit = s
			}
		}
		if hit == nil {
			return "", false
		}
		if frag == "" {
			return "#" + hit.ID, true
		}
		return "#" + hit.ID + "-" + frag, true
	}
}

func (b *Bundle) prepare(consoleURL string) {
	console := "#"
	if consoleURL != "" {
		console = html.EscapeString(strings.TrimRight(consoleURL, "/"))
	}
	for _, s := range b.Sections() {
		if s.kind == ".md" {
			title, body := renderMarkdown(s.body, s.ID, b.resolver(s))
			if s.Title == "" {
				s.Title = title
			}
			s.body = strings.ReplaceAll(body, ConsolePlaceholder, console)
			s.kind = "rendered"
			continue
		}
		if s.kind == ".html" {
			body := headerRe.ReplaceAllString(s.body, "")
			res := b.resolver(s)
			body = hrefRe.ReplaceAllStringFunc(body, func(v string) string {
				t := html.UnescapeString(hrefRe.FindStringSubmatch(v)[1])
				if h, ok := res(t); ok {
					return `href="` + html.EscapeString(h) + `"`
				}
				return v
			})
			body = strings.ReplaceAll(body, ConsolePlaceholder, console)
			s.body = preRe.ReplaceAllStringFunc(body, func(v string) string {
				code := preRe.FindStringSubmatch(v)[2]
				cls := ""
				if !strings.Contains(strings.TrimSpace(code), "\n") {
					cls = ` class="oneline"`
				}
				return `<div class="codeblock"><div class="code-head"><span class="lang">` + guessLang(html.UnescapeString(code)) + `</span><button class="copy" type="button">Copy</button></div><pre` + cls + `><code>` + code + `</code></pre></div>`
			})
			s.kind = "rendered"
		}
		if s.Title == "" {
			s.Title = s.Path
		}
	}
}

func Render(assets fs.FS, b *Bundle, o Options) ([]byte, error) {
	b.prepare(o.ConsoleURL)
	css, err := fs.ReadFile(assets, "assets/style.css")
	if err != nil {
		return nil, err
	}
	js, err := fs.ReadFile(assets, "assets/script.js")
	if err != nil {
		return nil, err
	}
	var fonts strings.Builder
	for _, f := range []struct{ file, weight string }{{"Vazirmatn-FD-NL-Regular.woff2", "400"}, {"Vazirmatn-FD-NL-Bold.woff2", "700"}} {
		raw, err := fs.ReadFile(assets, "assets/fonts/"+f.file)
		if err != nil {
			return nil, err
		}
		fonts.WriteString("@font-face{font-family:Vazirmatn;font-weight:" + f.weight + ";font-display:swap;src:url(data:font/woff2;base64," + base64.StdEncoding.EncodeToString(raw) + ") format('woff2')}\n")
	}
	projectTitle := "بخش عمومی"
	if b.Project != nil {
		projectTitle = b.Project.Meta.Title
	}
	var w bytes.Buffer
	esc := html.EscapeString
	w.WriteString("<!doctype html>\n<html lang=\"fa\" dir=\"rtl\">\n<head>\n<meta charset=\"utf-8\">\n<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	w.WriteString("<title>راهنمای تیم فرانت · " + esc(projectTitle) + "</title>\n<style>\n")
	w.WriteString(fonts.String())
	w.Write(css)
	w.WriteString("</style>\n</head>\n<body>\n")
	w.WriteString(`<div class="wrap">` + "\n<header>\n")
	w.WriteString(`<button class="theme" type="button" data-theme-toggle aria-label="روشن یا تیره">تیره / روشن</button>` + "\n")
	w.WriteString("<h1>راهنمای تیم فرانت: " + esc(projectTitle) + "</h1>\n")
	meta := "ساخته‌شده در " + Jalali(o.GeneratedAt)
	if o.Commit != "" {
		meta += ` · نسخهٔ سرور <code>` + esc(shortCommit(o.Commit)) + `</code>`
	}
	w.WriteString("<p>" + meta + "</p>\n")
	w.WriteString("<p>این فایل دو بخش دارد: <b>بخش عمومی</b> برای همهٔ پروژه‌ها، و <b>بخش اختصاصی</b> همین پروژه. فهرست تغییرها اول آمده است.</p>\n")
	w.WriteString("</header>\n")
	w.WriteString(`<nav class="toc" aria-label="فهرست"><h2>فهرست</h2>` + "\n<ol>\n")
	w.WriteString(`<li><a href="#changes">تازه‌ها</a></li>` + "\n")
	w.WriteString(`<li><a href="#part-general">بخش عمومی</a>` + tocList(b.General.Sections) + "</li>\n")
	if b.Project != nil {
		w.WriteString(`<li><a href="#part-project">بخش اختصاصی: ` + esc(b.Project.Meta.Title) + `</a>` + tocList(b.Project.Sections) + "</li>\n")
	}
	w.WriteString("</ol>\n</nav>\n")
	w.WriteString(`<section class="step" id="changes"><div class="step-head"><div class="num">۰</div><div><h2>تازه‌ها</h2><div class="sub">جدیدترین تغییر بالاست</div></div></div>` + "\n")
	w.WriteString(`<div class="tablewrap"><table class="changes"><thead><tr><th>تاریخ</th><th>تغییر</th><th>بخش</th></tr></thead><tbody>` + "\n")
	for _, e := range b.Changelog() {
		title := esc(e.Title)
		if e.Anchor != "" {
			title = `<a href="#` + esc(e.Anchor) + `">` + title + `</a>`
		}
		badge := `<span class="badge general">عمومی</span>`
		if e.Scope == ScopeProject {
			badge = `<span class="badge project">اختصاصی</span>`
		}
		w.WriteString(`<tr data-scope="` + e.Scope + `"><td class="nowrap">` + jalaliString(e.Date) + "</td><td>" + title + "</td><td>" + badge + "</td></tr>\n")
	}
	w.WriteString("</tbody></table></div>\n</section>\n")
	writePart(&w, "part-general", "بخش عمومی", b.General.Meta.Summary, b.General.Sections, 1)
	if b.Project != nil {
		writePart(&w, "part-project", "بخش اختصاصی: "+b.Project.Meta.Title, b.Project.Meta.Summary, b.Project.Sections, len(b.General.Sections)+1)
	} else {
		w.WriteString(`<h2 class="part" id="part-project">بخش اختصاصی</h2>` + "\n")
		w.WriteString(`<div class="note" data-no-project>این پروژه هنوز بخش اختصاصی ندارد. هر چه لازم دارید فعلاً در بخش عمومی است.</div>` + "\n")
	}
	if o.ConsoleURL != "" {
		w.WriteString(`<div class="warn">کلیدها و رمزها داخل این فایل نیستند. کلید سرور را از مدیر سرور بگیرید؛ راه تنظیمش در <a href="` + esc(strings.TrimRight(o.ConsoleURL, "/")) + `/#/system/sdk" target="_blank" rel="noopener">صفحهٔ SDK کنسول</a> آمده است.</div>` + "\n")
	}
	w.WriteString("<footer>این فایل را کنسول مدیریت ساخته است. نسخهٔ تازه را از همان‌جا بگیرید.</footer>\n</div>\n<script>\n")
	w.Write(js)
	w.WriteString("</script>\n</body>\n</html>\n")
	return w.Bytes(), nil
}

func tocList(sections []*Section) string {
	if len(sections) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("<ol>")
	for _, s := range sections {
		b.WriteString(`<li><a href="#` + html.EscapeString(s.ID) + `">` + html.EscapeString(s.Title) + "</a></li>")
	}
	b.WriteString("</ol>")
	return b.String()
}

func writePart(w *bytes.Buffer, id, title, summary string, sections []*Section, start int) {
	w.WriteString(`<h2 class="part" id="` + id + `">` + html.EscapeString(title) + "</h2>\n")
	if summary != "" {
		w.WriteString(`<p class="part-sub">` + html.EscapeString(summary) + "</p>\n")
	}
	for i, s := range sections {
		cls := "step doc"
		if strings.Contains(s.body, `class="step"`) {
			cls = "doc doc-group"
		}
		w.WriteString(`<section class="` + cls + `" id="` + html.EscapeString(s.ID) + `" data-scope="` + s.Scope + `"><div class="step-head"><div class="num">` + FaDigits(strconv.Itoa(start+i)) + `</div><div><h2>` + html.EscapeString(s.Title) + "</h2>")
		if s.Summary != "" {
			w.WriteString(`<div class="sub">` + html.EscapeString(s.Summary) + "</div>")
		}
		w.WriteString("</div></div>\n")
		w.WriteString(s.body)
		w.WriteString("\n</section>\n")
	}
}

func shortCommit(c string) string {
	if len(c) > 8 {
		return c[:8]
	}
	return c
}

func FaDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune('۰' + (r - '0'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

var jalaliMonths = []string{"فروردین", "اردیبهشت", "خرداد", "تیر", "مرداد", "شهریور", "مهر", "آبان", "آذر", "دی", "بهمن", "اسفند"}

func jalaliString(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return html.EscapeString(date)
	}
	return Jalali(t)
}

func Jalali(t time.Time) string {
	y, m, d := ToJalali(t.Year(), int(t.Month()), t.Day())
	return FaDigits(fmt.Sprintf("%d %s %d", d, jalaliMonths[m-1], y))
}

func ToJalali(gy, gm, gd int) (int, int, int) {
	gdm := []int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	gy2 := gy
	if gm <= 2 {
		gy2 = gy - 1
	}
	days := 355666 + 365*gy + (gy2+3)/4 - (gy2+99)/100 + (gy2+399)/400 + gd + gdm[gm-1]
	jy := -1595 + 33*(days/12053)
	days %= 12053
	jy += 4 * (days / 1461)
	days %= 1461
	if days > 365 {
		jy += (days - 1) / 365
		days = (days - 1) % 365
	}
	var jm, jd int
	if days < 186 {
		jm = 1 + days/31
		jd = 1 + days%31
	} else {
		jm = 7 + (days-186)/30
		jd = 1 + (days-186)%30
	}
	return jy, jm, jd
}
