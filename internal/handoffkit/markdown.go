package handoffkit

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

type linkResolver func(target string) (string, bool)

type mdRenderer struct {
	prefix  string
	resolve linkResolver
	ids     map[string]int
}

var (
	fenceRe     = regexp.MustCompile("^(\\s*)(```+|~~~+)\\s*([A-Za-z0-9_+-]*)\\s*$")
	headingRe   = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*#*\s*$`)
	bulletRe    = regexp.MustCompile(`^(\s*)([-*+])\s+(.*)$`)
	orderedRe   = regexp.MustCompile(`^(\s*)(\d{1,3})[.)]\s+(.*)$`)
	hrRe        = regexp.MustCompile(`^\s{0,3}(-\s*-\s*-[-\s]*|\*\s*\*\s*\*[*\s]*|_\s*_\s*_[_\s]*)$`)
	tableSepRe  = regexp.MustCompile(`^\s*\|?\s*:?-{2,}:?\s*(\|\s*:?-{2,}:?\s*)*\|?\s*$`)
	linkRe      = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)
	boldRe      = regexp.MustCompile(`\*\*([^*]+?)\*\*`)
	italicRe    = regexp.MustCompile(`(^|[^*\w])\*([^*\s][^*]*?)\*([^*\w]|$)`)
	strikeRe    = regexp.MustCompile(`~~([^~]+?)~~`)
	taskRe      = regexp.MustCompile(`^\[([ xX])\]\s+(.*)$`)
	placeholder = "\x00%d\x00"
)

func Slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r == ' ':
			b.WriteRune('-')
		case r == '-' || r == '_':
			b.WriteRune(r)
		case unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.Is(unicode.M, r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (m *mdRenderer) headingID(text string) string {
	base := Slug(plainText(text))
	n := m.ids[base]
	m.ids[base] = n + 1
	if n > 0 {
		base = base + "-" + itoa(n)
	}
	return m.prefix + "-" + base
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}

func plainText(s string) string {
	s = strings.ReplaceAll(s, "`", "")
	s = linkRe.ReplaceAllString(s, "$1")
	s = strings.ReplaceAll(s, "**", "")
	return s
}

func renderMarkdown(src, prefix string, resolve linkResolver) (string, string) {
	m := &mdRenderer{prefix: prefix, resolve: resolve, ids: map[string]int{}}
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	title := ""
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if h := headingRe.FindStringSubmatch(l); h != nil && len(h[1]) == 1 {
			title = h[2]
			lines = append(lines[:i:i], lines[i+1:]...)
		}
		break
	}
	var b strings.Builder
	m.blocks(&b, lines)
	return title, b.String()
}

func indentOf(l string) int {
	n := 0
	for _, r := range l {
		if r == ' ' {
			n++
		} else if r == '\t' {
			n += 4
		} else {
			break
		}
	}
	return n
}

func dedent(l string, n int) string {
	i := 0
	for i < len(l) && n > 0 && (l[i] == ' ' || l[i] == '\t') {
		if l[i] == '\t' {
			n -= 4
		} else {
			n--
		}
		i++
	}
	return l[i:]
}

func (m *mdRenderer) blocks(b *strings.Builder, lines []string) {
	i := 0
	for i < len(lines) {
		l := lines[i]
		t := strings.TrimSpace(l)
		switch {
		case t == "":
			i++
		case fenceRe.MatchString(l):
			i = m.fence(b, lines, i)
		case headingRe.MatchString(t):
			h := headingRe.FindStringSubmatch(t)
			level := len(h[1]) + 1
			if level > 6 {
				level = 6
			}
			tag := "h" + itoa(level)
			b.WriteString("<" + tag + ` id="` + html.EscapeString(m.headingID(h[2])) + `">` + m.inline(h[2]) + "</" + tag + ">\n")
			i++
		case hrRe.MatchString(l):
			b.WriteString("<hr>\n")
			i++
		case strings.HasPrefix(t, "|") && i+1 < len(lines) && tableSepRe.MatchString(lines[i+1]):
			i = m.table(b, lines, i)
		case bulletRe.MatchString(l) || orderedRe.MatchString(l):
			i = m.list(b, lines, i)
		case strings.HasPrefix(t, ">"):
			var q []string
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), ">") {
				q = append(q, strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(lines[i]), ">"), " "))
				i++
			}
			b.WriteString(`<div class="note">`)
			m.blocks(b, q)
			b.WriteString("</div>\n")
		default:
			var p []string
			for i < len(lines) {
				c := lines[i]
				ct := strings.TrimSpace(c)
				if ct == "" || fenceRe.MatchString(c) || headingRe.MatchString(ct) || hrRe.MatchString(c) || bulletRe.MatchString(c) || orderedRe.MatchString(c) || (strings.HasPrefix(ct, "|") && i+1 < len(lines) && tableSepRe.MatchString(lines[i+1])) {
					break
				}
				p = append(p, ct)
				i++
			}
			b.WriteString("<p>" + m.inline(strings.Join(p, " ")) + "</p>\n")
		}
	}
}

func (m *mdRenderer) fence(b *strings.Builder, lines []string, i int) int {
	f := fenceRe.FindStringSubmatch(lines[i])
	ind, marker, lang := len(f[1]), f[2], f[3]
	i++
	var code []string
	for i < len(lines) {
		if strings.TrimSpace(lines[i]) == marker || (strings.HasPrefix(strings.TrimSpace(lines[i]), marker) && strings.Trim(strings.TrimSpace(lines[i]), string(marker[0])) == "") {
			i++
			break
		}
		code = append(code, dedent(lines[i], ind))
		i++
	}
	b.WriteString(codeBlock(strings.Join(code, "\n"), lang))
	return i
}

var langNames = map[string]string{"ts": "TypeScript", "typescript": "TypeScript", "js": "JavaScript", "javascript": "JavaScript", "json": "JSON", "bash": "Shell", "sh": "Shell", "shell": "Shell", "csharp": "C#", "cs": "C#", "env": "env", "toml": "TOML", "yaml": "YAML", "yml": "YAML", "sql": "SQL", "html": "HTML", "xml": "XML", "go": "Go", "http": "HTTP", "text": "Text"}

func codeBlock(code, lang string) string {
	name := langNames[strings.ToLower(lang)]
	if name == "" {
		name = "Text"
		if lang != "" {
			name = lang
		}
	}
	cls := ""
	if !strings.Contains(strings.TrimSpace(code), "\n") {
		cls = ` class="oneline"`
	}
	return `<div class="codeblock"><div class="code-head"><span class="lang">` + html.EscapeString(name) + `</span><button class="copy" type="button">Copy</button></div><pre` + cls + `><code>` + html.EscapeString(strings.TrimRight(code, "\n ")) + "</code></pre></div>\n"
}

func splitRow(l string) []string {
	l = strings.TrimSpace(l)
	l = strings.TrimPrefix(l, "|")
	if strings.HasSuffix(l, "|") && !strings.HasSuffix(l, `\|`) {
		l = l[:len(l)-1]
	}
	var cells []string
	var cur strings.Builder
	inCode := false
	for i := 0; i < len(l); i++ {
		c := l[i]
		switch {
		case c == '\\' && i+1 < len(l) && l[i+1] == '|':
			cur.WriteByte('|')
			i++
		case c == '`':
			inCode = !inCode
			cur.WriteByte(c)
		case c == '|' && !inCode:
			cells = append(cells, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	return append(cells, strings.TrimSpace(cur.String()))
}

func (m *mdRenderer) table(b *strings.Builder, lines []string, i int) int {
	head := splitRow(lines[i])
	i += 2
	b.WriteString(`<div class="tablewrap"><table><thead><tr>`)
	for _, c := range head {
		b.WriteString("<th>" + m.inline(c) + "</th>")
	}
	b.WriteString("</tr></thead><tbody>\n")
	for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), "|") {
		b.WriteString("<tr>")
		for _, c := range splitRow(lines[i]) {
			b.WriteString("<td>" + m.inline(c) + "</td>")
		}
		b.WriteString("</tr>\n")
		i++
	}
	b.WriteString("</tbody></table></div>\n")
	return i
}

func listMarker(l string) (indent int, ordered bool, content string, width int, ok bool) {
	if g := bulletRe.FindStringSubmatch(l); g != nil && !hrRe.MatchString(l) {
		ind := indentOf(g[1])
		return ind, false, g[3], ind + len(g[2]) + 1, true
	}
	if g := orderedRe.FindStringSubmatch(l); g != nil {
		ind := indentOf(g[1])
		return ind, true, g[3], ind + len(g[2]) + 2, true
	}
	return 0, false, "", 0, false
}

func (m *mdRenderer) list(b *strings.Builder, lines []string, i int) int {
	base, ordered, _, _, _ := listMarker(lines[i])
	tag := "ul"
	if ordered {
		tag = "ol"
	}
	b.WriteString("<" + tag + ">\n")
	for i < len(lines) {
		ind, ord, content, width, ok := listMarker(lines[i])
		if !ok || ind != base || ord != ordered {
			break
		}
		item := []string{content}
		i++
		for i < len(lines) {
			l := lines[i]
			if strings.TrimSpace(l) == "" {
				j := i + 1
				for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
					j++
				}
				if j < len(lines) && indentOf(lines[j]) > base {
					for ; i < j; i++ {
						item = append(item, "")
					}
					continue
				}
				break
			}
			li, _, _, _, isItem := listMarker(l)
			if isItem && li <= base {
				break
			}
			if !isItem && indentOf(l) <= base && (headingRe.MatchString(strings.TrimSpace(l)) || strings.HasPrefix(strings.TrimSpace(l), "|") || fenceRe.MatchString(l)) {
				break
			}
			item = append(item, dedent(l, width))
			i++
		}
		b.WriteString("<li>")
		first := item[0]
		check := ""
		if t := taskRe.FindStringSubmatch(first); t != nil {
			check = `<input type="checkbox" disabled`
			if t[1] != " " {
				check += " checked"
			}
			check += "> "
			first = t[2]
		}
		rest := item[1:]
		var para []string
		para = append(para, strings.TrimSpace(first))
		k := 0
		for k < len(rest) && strings.TrimSpace(rest[k]) != "" && !fenceRe.MatchString(rest[k]) && !bulletRe.MatchString(rest[k]) && !orderedRe.MatchString(rest[k]) && !strings.HasPrefix(strings.TrimSpace(rest[k]), "|") {
			para = append(para, strings.TrimSpace(rest[k]))
			k++
		}
		b.WriteString(check + m.inline(strings.Join(para, " ")))
		if k < len(rest) {
			b.WriteString("\n")
			m.blocks(b, rest[k:])
		}
		b.WriteString("</li>\n")
	}
	b.WriteString("</" + tag + ">\n")
	return i
}

func (m *mdRenderer) inline(s string) string {
	var saved []string
	keep := func(v string) string {
		saved = append(saved, v)
		return strings.Replace(placeholder, "%d", itoa(len(saved)-1), 1)
	}
	var out strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '`' {
			n := 0
			for i+n < len(s) && s[i+n] == '`' {
				n++
			}
			ticks := strings.Repeat("`", n)
			if end := strings.Index(s[i+n:], ticks); end >= 0 {
				code := strings.TrimSpace(s[i+n : i+n+end])
				out.WriteString(keep("<code>" + html.EscapeString(code) + "</code>"))
				i += n + end + n
				continue
			}
			out.WriteString(ticks)
			i += n
			continue
		}
		out.WriteByte(s[i])
		i++
	}
	t := out.String()
	t = linkRe.ReplaceAllStringFunc(t, func(v string) string {
		g := linkRe.FindStringSubmatch(v)
		text, target := g[1], g[2]
		href, internal := target, false
		if m.resolve != nil {
			if h, ok := m.resolve(target); ok {
				href, internal = h, true
			}
		}
		attrs := ` href="` + html.EscapeString(href) + `"`
		if !internal && (strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://")) {
			attrs += ` target="_blank" rel="noopener"`
		}
		return keep("<a" + attrs + ">" + m.emphasis(html.EscapeString(text), saved) + "</a>")
	})
	t = html.EscapeString(t)
	t = m.emphasis(t, saved)
	for k := len(saved) - 1; k >= 0; k-- {
		t = strings.ReplaceAll(t, strings.Replace(placeholder, "%d", itoa(k), 1), saved[k])
	}
	return t
}

func (m *mdRenderer) emphasis(t string, saved []string) string {
	t = boldRe.ReplaceAllString(t, "<b>$1</b>")
	t = strikeRe.ReplaceAllString(t, "<s>$1</s>")
	t = italicRe.ReplaceAllString(t, "$1<i>$2</i>$3")
	return t
}
