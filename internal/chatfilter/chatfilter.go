package chatfilter

import (
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const (
	ActionMask  = "mask"
	ActionBlock = "block"
	ActionFlag  = "flag"
	MatchWord   = "word"
	MatchAny    = "contains"
)

type Rule struct {
	ID     string
	Word   string
	Action string
	Match  string
}

type Hit struct {
	RuleID string `json:"id"`
	Word   string `json:"word"`
	Action string `json:"action"`
}

type Result struct {
	Text    string
	Changed bool
	Blocked bool
	Flagged bool
	Hits    []Hit
}

type unit struct {
	r      rune
	src    int
	letter bool
	space  bool
}

type run struct {
	r     rune
	n     int
	first int
	last  int
}

type token struct {
	runs []run
}

var removed = map[rune]bool{
	0x200B: true, 0x200C: true, 0x200D: true, 0x200E: true, 0x200F: true, 0x2060: true, 0x2061: true,
	0x2062: true, 0x2063: true, 0x2064: true, 0xFEFF: true, 0x00AD: true, 0x0640: true, 0x034F: true,
	0x061C: true, 0x202A: true, 0x202B: true, 0x202C: true, 0x202D: true, 0x202E: true, 0x2066: true,
	0x2067: true, 0x2068: true, 0x2069: true,
}

var letterMap = map[rune]rune{
	'ي': 'ی', 'ى': 'ی', 'ې': 'ی', 'ۍ': 'ی', 'ے': 'ی',
	'ك': 'ک', 'ڪ': 'ک',
	'ة': 'ه', 'ە': 'ه', 'ھ': 'ه', 'ہ': 'ه', 'ۃ': 'ه',
	'ٱ': 'ا',
	'а': 'a', 'е': 'e', 'о': 'o', 'р': 'p', 'с': 'c', 'у': 'y', 'х': 'x', 'к': 'k', 'м': 'm', 'т': 't', 'і': 'i', 'ѕ': 's', 'д': 'd',
	'α': 'a', 'ο': 'o', 'ε': 'e', 'ι': 'i', 'κ': 'k', 'ν': 'v', 'τ': 't', 'υ': 'u',
}

var leet = map[rune]rune{'0': 'o', '1': 'i', '3': 'e', '4': 'a', '5': 's', '7': 't', '8': 'b', '@': 'a', '$': 's', '!': 'i', '|': 'l', '+': 't'}

var suffixes = map[string]bool{
	"ی": true, "ای": true, "ها": true, "های": true, "ان": true, "ت": true, "ش": true, "م": true,
	"مان": true, "تان": true, "شان": true, "تر": true, "ترین": true, "ه": true, "هایت": true, "هایش": true,
	"s": true, "es": true, "ed": true, "er": true, "ers": true, "ing": true, "y": true, "ish": true,
}

func mapRune(r rune) rune {
	switch {
	case r >= 0x06F0 && r <= 0x06F9:
		return '0' + (r - 0x06F0)
	case r >= 0x0660 && r <= 0x0669:
		return '0' + (r - 0x0660)
	}
	if m, ok := letterMap[r]; ok {
		return m
	}
	return unicode.ToLower(r)
}

func isLatin(r rune) bool {
	return r < 0x250 && unicode.IsLetter(r)
}

func units(src []rune) []unit {
	out := make([]unit, 0, len(src))
	var buf [8]rune
	for i, r := range src {
		if removed[r] || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) {
			continue
		}
		parts := buf[:0]
		for _, d := range norm.NFKD.String(string(r)) {
			parts = append(parts, d)
		}
		for _, d := range parts {
			if removed[d] || unicode.Is(unicode.Mn, d) || unicode.Is(unicode.Me, d) || unicode.Is(unicode.Cf, d) {
				continue
			}
			m := mapRune(d)
			out = append(out, unit{r: m, src: i, letter: unicode.IsLetter(m), space: unicode.IsSpace(m)})
		}
	}
	for i := len(out) - 2; i >= 0; i-- {
		if to, ok := leet[out[i].r]; ok && isLatin(out[i+1].r) {
			out[i].r = to
			out[i].letter = true
		}
	}
	return out
}

func appendRun(runs []run, u unit) []run {
	if n := len(runs); n > 0 && runs[n-1].r == u.r {
		runs[n-1].n++
		runs[n-1].last = u.src
		return runs
	}
	return append(runs, run{r: u.r, n: 1, first: u.src, last: u.src})
}

func tokenize(us []unit, byPunct bool) []token {
	var toks []token
	var cur []run
	letters := 0
	flush := func() {
		if len(cur) > 0 {
			toks = append(toks, token{runs: cur})
		}
		cur = nil
		letters = 0
	}
	for _, u := range us {
		if u.space || (byPunct && !u.letter) {
			flush()
			continue
		}
		if !u.letter {
			continue
		}
		cur = appendRun(cur, u)
		letters++
	}
	flush()
	out := make([]token, 0, len(toks))
	for i := 0; i < len(toks); {
		j := i
		for j < len(toks) && len(toks[j].runs) == 1 && toks[j].runs[0].n == 1 {
			j++
		}
		if j-i >= 2 {
			var merged []run
			for _, t := range toks[i:j] {
				for _, r := range t.runs {
					merged = appendRun(merged, unit{r: r.r, src: r.first})
				}
			}
			out = append(out, token{runs: merged})
			i = j
			continue
		}
		out = append(out, toks[i])
		i++
	}
	return out
}

func joined(toks []token) []run {
	var out []run
	for _, t := range toks {
		for _, r := range t.runs {
			if n := len(out); n > 0 && out[n-1].r == r.r {
				out[n-1].n += r.n
				out[n-1].last = r.last
				continue
			}
			out = append(out, r)
		}
	}
	return out
}

func collapsed(runs []run) string {
	var b strings.Builder
	for _, r := range runs {
		b.WriteRune(r.r)
	}
	return b.String()
}

func Normalize(s string) string {
	toks := tokenize(units([]rune(s)), false)
	parts := make([]string, 0, len(toks))
	for _, t := range toks {
		parts = append(parts, collapsed(t.runs))
	}
	return strings.Join(parts, " ")
}

type compiled struct {
	rule   Rule
	tokens []token
	all    []run
}

type Filter struct {
	rules []compiled
}

func Compile(rules []Rule) *Filter {
	f := &Filter{}
	for _, r := range rules {
		toks := tokenize(units([]rune(r.Word)), true)
		if len(toks) == 0 {
			continue
		}
		if r.Match != MatchAny {
			r.Match = MatchWord
		}
		f.rules = append(f.rules, compiled{rule: r, tokens: toks, all: joined(toks)})
	}
	return f
}

func (f *Filter) Len() int { return len(f.rules) }

func runsMatch(text, word []run) bool {
	if len(text) < len(word) {
		return false
	}
	for i, w := range word {
		if text[i].r != w.r || text[i].n < w.n {
			return false
		}
	}
	return true
}

func tokenMatch(text, word token, last bool) bool {
	if !runsMatch(text.runs, word.runs) {
		return false
	}
	rest := text.runs[len(word.runs):]
	if len(rest) == 0 {
		return true
	}
	if !last {
		return false
	}
	return suffixes[collapsed(rest)]
}

type span struct{ from, to int }

func (c compiled) find(views [][]token, all []run) []span {
	var out []span
	if c.rule.Match == MatchAny {
		w := c.all
		for i := 0; i+len(w) <= len(all); i++ {
			ok := true
			for j, wr := range w {
				tr := all[i+j]
				if tr.r != wr.r || tr.n < wr.n {
					ok = false
					break
				}
			}
			if ok {
				out = append(out, span{all[i].first, all[i+len(w)-1].last})
				i += len(w) - 1
			}
		}
		return out
	}
	k := len(c.tokens)
	if k == 1 {
		for _, toks := range views {
			for i := 0; i+1 < len(toks); i++ {
				pair := token{runs: joined(toks[i : i+2])}
				if tokenMatch(pair, c.tokens[0], true) && !tokenMatch(toks[i], c.tokens[0], true) {
					last := toks[i+1].runs
					out = append(out, span{toks[i].runs[0].first, last[len(last)-1].last})
					i++
				}
			}
		}
	}
	for _, toks := range views {
		for i := 0; i+k <= len(toks); i++ {
			ok := true
			for j := 0; j < k; j++ {
				if !tokenMatch(toks[i+j], c.tokens[j], j == k-1) {
					ok = false
					break
				}
			}
			if ok {
				last := toks[i+k-1].runs
				out = append(out, span{toks[i].runs[0].first, last[len(last)-1].last})
				i += k - 1
			}
		}
	}
	return out
}

func (f *Filter) Apply(text string) Result {
	res := Result{Text: text}
	if f == nil || len(f.rules) == 0 || text == "" {
		return res
	}
	src := []rune(text)
	us := units(src)
	views := [][]token{tokenize(us, true), tokenize(us, false)}
	all := joined(views[0])
	var masks []span
	seen := map[string]bool{}
	for _, c := range f.rules {
		spans := c.find(views, all)
		if len(spans) == 0 {
			continue
		}
		if !seen[c.rule.ID+"|"+c.rule.Word] {
			seen[c.rule.ID+"|"+c.rule.Word] = true
			res.Hits = append(res.Hits, Hit{RuleID: c.rule.ID, Word: c.rule.Word, Action: c.rule.Action})
		}
		switch c.rule.Action {
		case ActionBlock:
			res.Blocked = true
		case ActionFlag:
			res.Flagged = true
		default:
			masks = append(masks, spans...)
		}
	}
	if len(masks) == 0 {
		return res
	}
	sort.Slice(masks, func(i, j int) bool { return masks[i].from < masks[j].from })
	hide := make([]bool, len(src))
	for _, m := range masks {
		for i := m.from; i <= m.to && i < len(src); i++ {
			hide[i] = true
		}
	}
	var b strings.Builder
	for i, r := range src {
		if !hide[i] {
			b.WriteRune(r)
			continue
		}
		if removed[r] || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) {
			continue
		}
		if unicode.IsSpace(r) || (unicode.IsPunct(r) && leet[r] == 0) {
			b.WriteRune(r)
			continue
		}
		b.WriteRune('*')
	}
	res.Text = b.String()
	res.Changed = res.Text != text
	return res
}
