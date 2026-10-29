package main

import (
	"html"
	"regexp"
	"strings"
)

var orderedItem = regexp.MustCompile(`^[0-9]{1,3}[.)] `)

type mdBlock struct {
	kind  string
	lines []string
}

func parseBlocks(src string) []mdBlock {
	src = strings.ReplaceAll(strings.ReplaceAll(src, "\r\n", "\n"), "\r", "\n")
	var blocks []mdBlock
	var cur *mdBlock
	flush := func() {
		if cur != nil && len(cur.lines) > 0 {
			blocks = append(blocks, *cur)
		}
		cur = nil
	}
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			flush()
		case strings.HasPrefix(trimmed, "### "):
			flush()
			blocks = append(blocks, mdBlock{kind: "h3", lines: []string{strings.TrimSpace(trimmed[4:])}})
		case strings.HasPrefix(trimmed, "## "):
			flush()
			blocks = append(blocks, mdBlock{kind: "h2", lines: []string{strings.TrimSpace(trimmed[3:])}})
		case strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* "):
			if cur == nil || cur.kind != "ul" {
				flush()
				cur = &mdBlock{kind: "ul"}
			}
			cur.lines = append(cur.lines, strings.TrimSpace(trimmed[2:]))
		case orderedItem.MatchString(trimmed):
			if cur == nil || cur.kind != "ol" {
				flush()
				cur = &mdBlock{kind: "ol"}
			}
			cur.lines = append(cur.lines, strings.TrimSpace(orderedItem.ReplaceAllString(trimmed, "")))
		default:
			if cur == nil || cur.kind != "p" {
				flush()
				cur = &mdBlock{kind: "p"}
			}
			cur.lines = append(cur.lines, trimmed)
		}
	}
	flush()
	return blocks
}

func renderMarkdown(src string) string {
	var b strings.Builder
	for _, blk := range parseBlocks(src) {
		switch blk.kind {
		case "h2", "h3":
			b.WriteString("<" + blk.kind + ">" + renderInline(blk.lines[0], true) + "</" + blk.kind + ">")
		case "ul", "ol":
			b.WriteString("<" + blk.kind + ">")
			for _, item := range blk.lines {
				b.WriteString("<li>" + renderInline(item, true) + "</li>")
			}
			b.WriteString("</" + blk.kind + ">")
		default:
			parts := make([]string, len(blk.lines))
			for i, l := range blk.lines {
				parts[i] = renderInline(l, true)
			}
			b.WriteString("<p>" + strings.Join(parts, "<br>") + "</p>")
		}
	}
	return b.String()
}

func markdownText(src string) string {
	var out []string
	for _, blk := range parseBlocks(src) {
		switch blk.kind {
		case "ul", "ol":
			items := make([]string, len(blk.lines))
			for i, l := range blk.lines {
				prefix := "- "
				if blk.kind == "ol" {
					prefix = itoa(i+1) + ". "
				}
				items[i] = prefix + inlineText(l)
			}
			out = append(out, strings.Join(items, "\n"))
		default:
			lines := make([]string, len(blk.lines))
			for i, l := range blk.lines {
				lines[i] = inlineText(l)
			}
			out = append(out, strings.Join(lines, "\n"))
		}
	}
	return strings.Join(out, "\n\n")
}

type linkMatch struct {
	text, url string
	end       int
}

func matchLink(s string, start int) (linkMatch, bool) {
	depth := 0
	i := start + 1
	for ; i < len(s); i++ {
		if s[i] == '[' {
			depth++
		} else if s[i] == ']' {
			if depth == 0 {
				break
			}
			depth--
		}
	}
	if i >= len(s)-1 || s[i+1] != '(' {
		return linkMatch{}, false
	}
	close := strings.IndexByte(s[i+2:], ')')
	if close < 0 {
		return linkMatch{}, false
	}
	return linkMatch{text: s[start+1 : i], url: strings.TrimSpace(s[i+2 : i+2+close]), end: i + 2 + close + 1}, true
}

func renderInline(s string, links bool) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if strings.HasPrefix(s[i:], "![") {
			if m, ok := matchLink(s, i+1); ok && safeURL(m.url) {
				b.WriteString(`<img src="` + html.EscapeString(m.url) + `" alt="` + html.EscapeString(m.text) + `">`)
				i = m.end
				continue
			}
		}
		if s[i] == '[' && links {
			if m, ok := matchLink(s, i); ok && safeURL(m.url) {
				b.WriteString(`<a href="` + html.EscapeString(m.url) + `" rel="noopener noreferrer" target="_blank">` + renderInline(m.text, false) + `</a>`)
				i = m.end
				continue
			}
		}
		if strings.HasPrefix(s[i:], "**") {
			if end := strings.Index(s[i+2:], "**"); end > 0 {
				b.WriteString("<strong>" + renderInline(s[i+2:i+2+end], links) + "</strong>")
				i = i + 2 + end + 2
				continue
			}
		}
		if s[i] == '*' && i+1 < len(s) && s[i+1] != ' ' && s[i+1] != '*' {
			if end := strings.IndexByte(s[i+1:], '*'); end > 0 {
				b.WriteString("<em>" + renderInline(s[i+1:i+1+end], links) + "</em>")
				i = i + 1 + end + 1
				continue
			}
		}
		j := i + 1
		for j < len(s) && s[j] != '*' && s[j] != '[' && s[j] != '!' {
			j++
		}
		b.WriteString(html.EscapeString(s[i:j]))
		i = j
	}
	return b.String()
}

func inlineText(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if strings.HasPrefix(s[i:], "![") {
			if m, ok := matchLink(s, i+1); ok && safeURL(m.url) {
				if m.text != "" {
					b.WriteString(m.text + " ")
				}
				b.WriteString("(" + m.url + ")")
				i = m.end
				continue
			}
		}
		if s[i] == '[' {
			if m, ok := matchLink(s, i); ok && safeURL(m.url) {
				b.WriteString(inlineText(m.text) + " (" + m.url + ")")
				i = m.end
				continue
			}
		}
		if strings.HasPrefix(s[i:], "**") {
			if end := strings.Index(s[i+2:], "**"); end > 0 {
				b.WriteString(inlineText(s[i+2 : i+2+end]))
				i = i + 2 + end + 2
				continue
			}
		}
		if s[i] == '*' && i+1 < len(s) && s[i+1] != ' ' && s[i+1] != '*' {
			if end := strings.IndexByte(s[i+1:], '*'); end > 0 {
				b.WriteString(inlineText(s[i+1 : i+1+end]))
				i = i + 1 + end + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func emailHTML(subject, bodyHTML, lang string) string {
	dir := "rtl"
	if lang == "en" {
		dir = "ltr"
	}
	return `<!DOCTYPE html><html lang="` + lang + `" dir="` + dir + `"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>` +
		html.EscapeString(subject) + `</title></head><body style="font-family:Tahoma,Arial,sans-serif;line-height:1.7;color:#1f2328;background:#ffffff;margin:0;padding:24px">` +
		`<div style="max-width:600px;margin:0 auto">` + bodyHTML + `</div></body></html>`
}
