package main

import (
	"html"
	"regexp"
	"strconv"
	"strings"

	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var (
	richDropped = map[string]bool{"script": true, "style": true, "iframe": true, "object": true, "embed": true, "noscript": true, "template": true, "svg": true, "math": true,
		"form": true, "textarea": true, "select": true, "button": true, "head": true, "title": true, "meta": true, "link": true, "frame": true, "frameset": true, "applet": true, "audio": true, "video": true, "canvas": true, "img": true, "picture": true, "source": true}
	richRename = map[string]string{"b": "strong", "i": "em", "del": "s", "strike": "s", "ins": "u"}
	richBlocks = map[string]bool{"p": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true, "ul": true, "ol": true, "li": true, "blockquote": true, "pre": true,
		"hr": true, "table": true, "colgroup": true, "col": true, "thead": true, "tbody": true, "tr": true, "th": true, "td": true, "div": true}
	richInline      = map[string]bool{"strong": true, "em": true, "u": true, "s": true, "code": true, "br": true, "a": true, "label": true, "input": true, "span": true}
	richTransparent = map[string]bool{"section": true, "article": true, "header": true, "footer": true, "main": true, "figure": true, "figcaption": true, "aside": true, "nav": true,
		"body": true, "html": true, "center": true, "font": true, "small": true, "big": true, "sup": true, "sub": true, "mark": true, "abbr": true, "cite": true, "q": true, "time": true, "dl": true, "dt": true, "dd": true, "address": true, "caption": true, "tfoot": true}
	alignStyle = regexp.MustCompile(`(?i)text-align\s*:\s*(left|right|center|justify)`)
	widthList  = regexp.MustCompile(`^[0-9]{1,4}(,[0-9]{1,4}){0,20}$`)
	tagLike    = regexp.MustCompile(`(?i)</?[a-z][a-z0-9]*(\s[^<>]*)?/?>`)
)

func looksLikeHTML(s string) bool {
	return tagLike.MatchString(s)
}

func richLinkOK(href string) bool {
	href = strings.TrimSpace(href)
	lower := strings.ToLower(href)
	if strings.ContainsAny(href, " \t\r\n\"'<>`") || href == "" || len(href) > 2048 {
		return false
	}
	for _, p := range []string{"http://", "https://", "mailto:", "tel:"} {
		if strings.HasPrefix(lower, p) {
			return len(href) > len(p)
		}
	}
	return strings.HasPrefix(href, "/") && !strings.HasPrefix(href, "//") && !strings.HasPrefix(href, `/\`)
}

type richWriter struct {
	b       strings.Builder
	inlineP bool
	depth   int
}

func attr(n *xhtml.Node, key string) (string, bool) {
	for _, a := range n.Attr {
		if a.Namespace == "" && strings.EqualFold(a.Key, key) {
			return a.Val, true
		}
	}
	return "", false
}

func smallInt(v string, max int) (int, bool) {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	return n, err == nil && n >= 1 && n <= max
}

func (w *richWriter) open(name string, n *xhtml.Node, parent string) {
	var attrs []string
	add := func(k, v string) { attrs = append(attrs, k+`="`+html.EscapeString(v)+`"`) }
	switch name {
	case "p", "h1", "h2", "h3", "h4", "h5", "h6":
		if st, ok := attr(n, "style"); ok {
			if m := alignStyle.FindStringSubmatch(st); m != nil {
				add("style", "text-align: "+strings.ToLower(m[1]))
			}
		}
	case "a":
		if href, ok := attr(n, "href"); ok && richLinkOK(href) {
			add("href", strings.TrimSpace(href))
		}
		add("target", "_blank")
		add("rel", "noopener noreferrer nofollow")
	case "ol":
		if v, ok := attr(n, "start"); ok {
			if s, ok := smallInt(v, 100000); ok {
				add("start", strconv.Itoa(s))
			}
		}
	case "ul":
		if v, _ := attr(n, "data-type"); v == "taskList" {
			add("data-type", "taskList")
		}
	case "li":
		if v, _ := attr(n, "data-type"); v == "taskItem" && parent == "ul" {
			add("data-type", "taskItem")
			c, _ := attr(n, "data-checked")
			if c != "true" {
				c = "false"
			}
			add("data-checked", c)
		}
	case "th", "td":
		for _, k := range []string{"colspan", "rowspan"} {
			if v, ok := attr(n, k); ok {
				if s, ok := smallInt(v, 100); ok {
					add(k, strconv.Itoa(s))
				}
			}
		}
		if v, ok := attr(n, "colwidth"); ok && widthList.MatchString(strings.TrimSpace(v)) {
			add("colwidth", strings.TrimSpace(v))
		}
	case "input":
		add("type", "checkbox")
		if _, ok := attr(n, "checked"); ok {
			attrs = append(attrs, `checked="checked"`)
		}
		attrs = append(attrs, `disabled="disabled"`)
	}
	w.b.WriteString("<" + name)
	for _, a := range attrs {
		w.b.WriteString(" " + a)
	}
	w.b.WriteString(">")
}

func voidTag(name string) bool {
	return name == "br" || name == "hr" || name == "col" || name == "input"
}

func (w *richWriter) startP() {
	if !w.inlineP {
		w.b.WriteString("<p>")
		w.inlineP = true
	}
}

func (w *richWriter) endP() {
	if w.inlineP {
		w.b.WriteString("</p>")
		w.inlineP = false
	}
}

func (w *richWriter) walk(n *xhtml.Node, parent string, top bool) {
	w.depth++
	defer func() { w.depth-- }()
	if w.depth > 200 {
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case xhtml.TextNode:
			text := c.Data
			if top && !w.inlineP && strings.TrimSpace(text) == "" {
				continue
			}
			if parent == "table" || parent == "tbody" || parent == "thead" || parent == "tr" || parent == "ul" || parent == "ol" || parent == "colgroup" {
				if strings.TrimSpace(text) == "" {
					continue
				}
			}
			if top {
				w.startP()
			}
			w.b.WriteString(html.EscapeString(text))
		case xhtml.ElementNode:
			name := strings.ToLower(c.Data)
			if c.DataAtom != 0 {
				name = c.DataAtom.String()
			}
			if r, ok := richRename[name]; ok {
				name = r
			}
			if richDropped[name] {
				continue
			}
			if name == "input" {
				if tp, _ := attr(c, "type"); !strings.EqualFold(tp, "checkbox") {
					continue
				}
			}
			if richTransparent[name] || (!richBlocks[name] && !richInline[name]) {
				w.walk(c, parent, top)
				continue
			}
			block := richBlocks[name]
			if top {
				if block {
					w.endP()
				} else {
					w.startP()
				}
			}
			w.open(name, c, parent)
			if voidTag(name) {
				continue
			}
			w.walk(c, name, false)
			w.b.WriteString("</" + name + ">")
		}
	}
}

func sanitizeRichHTML(src string) string {
	nodes, err := xhtml.ParseFragment(strings.NewReader(src), &xhtml.Node{Type: xhtml.ElementNode, Data: "body", DataAtom: atom.Body})
	if err != nil {
		return ""
	}
	root := &xhtml.Node{Type: xhtml.ElementNode, Data: "div", DataAtom: atom.Div}
	for _, n := range nodes {
		root.AppendChild(n)
	}
	w := &richWriter{}
	w.walk(root, "", true)
	w.endP()
	out := strings.TrimSpace(w.b.String())
	for _, empty := range []string{"<p></p>"} {
		if out == empty {
			return ""
		}
	}
	return out
}

func richText(src string) string {
	nodes, err := xhtml.ParseFragment(strings.NewReader(src), &xhtml.Node{Type: xhtml.ElementNode, Data: "body", DataAtom: atom.Body})
	if err != nil {
		return ""
	}
	var b strings.Builder
	var walk func(n *xhtml.Node)
	walk = func(n *xhtml.Node) {
		switch n.Type {
		case xhtml.TextNode:
			b.WriteString(n.Data)
		case xhtml.ElementNode:
			name := strings.ToLower(n.Data)
			if richDropped[name] {
				return
			}
			if name == "br" {
				b.WriteString("\n")
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
			if richBlocks[name] && name != "col" && name != "colgroup" {
				b.WriteString("\n")
			} else if name == "td" || name == "th" {
				b.WriteString(" ")
			}
		}
	}
	for _, n := range nodes {
		walk(n)
	}
	lines := strings.Split(b.String(), "\n")
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.Join(strings.Fields(l), " ")
		if l != "" {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

func textToRich(src string) string {
	var b strings.Builder
	for _, para := range strings.Split(strings.ReplaceAll(strings.TrimSpace(src), "\r\n", "\n"), "\n\n") {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		lines := strings.Split(para, "\n")
		for i, l := range lines {
			lines[i] = html.EscapeString(strings.TrimSpace(l))
		}
		b.WriteString("<p>" + strings.Join(lines, "<br>") + "</p>")
	}
	return b.String()
}

func newsBodyHTML(format, body string) string {
	if format == "html" {
		return body
	}
	return renderMarkdown(body)
}

func newsBodyText(format, body string) string {
	if format == "html" {
		return richText(body)
	}
	return markdownText(body)
}
