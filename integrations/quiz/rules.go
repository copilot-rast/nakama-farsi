package main

import (
	"math"
	"math/rand"
	"net"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
)

func passNeed(k *pack, cfg settings.All, total int) int {
	if total <= 0 {
		return 0
	}
	if k.PassCorrect != nil {
		if *k.PassCorrect > total {
			return total
		}
		return *k.PassCorrect
	}
	pct := cfg.Scoring.PassPercent
	if k.PassPercent != nil {
		pct = *k.PassPercent
	}
	n := int(math.Ceil(pct*float64(total)/100 - 1e-9))
	if n > total {
		n = total
	}
	return n
}

func scoreOf(correct, total int) float64 {
	if total <= 0 {
		return 0
	}
	return math.Round(float64(correct)/float64(total)*10000) / 100
}

var letterForms = map[rune]rune{'ي': 'ی', 'ى': 'ی', 'ئ': 'ی', 'ك': 'ک', 'ة': 'ه', 'ۀ': 'ه', 'أ': 'ا', 'إ': 'ا', 'آ': 'ا', 'ٱ': 'ا', 'ؤ': 'و'}

func normalizeAnswer(s string, mode string) string {
	s = strings.TrimSpace(s)
	if mode == settings.MatchExact {
		return strings.ToLower(s)
	}
	var b strings.Builder
	space := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= '۰' && r <= '۹':
			r = '0' + (r - '۰')
		case r >= '٠' && r <= '٩':
			r = '0' + (r - '٠')
		case r == '‌' || r == '‍' || r == '-' || r == '_':
			r = ' '
		case r == 'ـ' || unicode.Is(unicode.Mn, r):
			continue
		}
		if v, ok := letterForms[r]; ok {
			r = v
		}
		if unicode.IsSpace(r) {
			space = true
			continue
		}
		if unicode.IsPunct(r) {
			continue
		}
		if space && b.Len() > 0 {
			b.WriteByte(' ')
		}
		space = false
		b.WriteRune(r)
	}
	return b.String()
}

func checkText(q *question, text, mode string) bool {
	needle := normalizeAnswer(text, mode)
	if needle == "" {
		return false
	}
	for _, list := range q.Answers {
		for _, a := range list {
			if normalizeAnswer(a, mode) == needle {
				return true
			}
		}
	}
	return false
}

func checkOptions(q *question, chosen []int, mode string) bool {
	if len(chosen) == 0 {
		return false
	}
	right := map[int]bool{}
	known := map[int]bool{}
	for _, o := range q.Options {
		known[o.ID] = true
		if o.Correct {
			right[o.ID] = true
		}
	}
	seen := map[int]bool{}
	for _, id := range chosen {
		if !known[id] || !right[id] {
			return false
		}
		seen[id] = true
	}
	if mode == settings.MultiAll {
		return len(seen) == len(right)
	}
	return true
}

func correctIDs(q *question) []int {
	out := []int{}
	for _, o := range q.Options {
		if o.Correct {
			out = append(out, o.ID)
		}
	}
	return out
}

func validOptions(opts []Option) string {
	if len(opts) < 2 {
		return "QUIZ_OPTIONS_REQUIRED"
	}
	if len(opts) > 12 {
		return "TOO_MANY_OPTIONS"
	}
	ids, right := map[int]bool{}, 0
	for _, o := range opts {
		if ids[o.ID] || o.ID < 0 {
			return "DUPLICATE_OPTION"
		}
		ids[o.ID] = true
		if o.Correct {
			right++
		}
		if len(o.Text) == 0 && o.MediaURL == "" {
			return "QUIZ_OPTION_TEXT_REQUIRED"
		}
	}
	if right == 0 {
		return "QUIZ_CORRECT_OPTION_REQUIRED"
	}
	return ""
}

func pickQuestions(pool []*question, count int, shuffleQ, shuffleO bool, rng *rand.Rand) []picked {
	list := append([]*question{}, pool...)
	if shuffleQ || (count > 0 && count < len(list)) {
		rng.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })
	}
	if count > 0 && count < len(list) {
		list = list[:count]
		if !shuffleQ {
			sort.SliceStable(list, func(i, j int) bool {
				if list[i].Sort != list[j].Sort {
					return list[i].Sort < list[j].Sort
				}
				return list[i].ID < list[j].ID
			})
		}
	}
	out := make([]picked, 0, len(list))
	for _, q := range list {
		p := picked{ID: q.ID}
		if q.Type == "options" {
			for _, o := range q.Options {
				p.Order = append(p.Order, o.ID)
			}
			if shuffleO {
				rng.Shuffle(len(p.Order), func(i, j int) { p.Order[i], p.Order[j] = p.Order[j], p.Order[i] })
			}
		}
		out = append(out, p)
	}
	return out
}

func questionSeconds(q *question, k *pack, cfg settings.All) int {
	if q != nil && q.TimeSeconds != nil {
		return *q.TimeSeconds
	}
	if k.QuestionSeconds > 0 {
		return k.QuestionSeconds
	}
	return cfg.Timing.QuestionSeconds
}

func questionDeadline(seconds int, clockStart time.Time) *time.Time {
	if seconds <= 0 {
		return nil
	}
	t := clockStart.Add(time.Duration(seconds) * time.Second)
	return &t
}

func dayKey(t time.Time, loc *time.Location) string { return t.In(loc).Format("2006-01-02") }

func render(body string, values map[string]string) string {
	var b strings.Builder
	for {
		open := strings.IndexByte(body, '{')
		if open < 0 {
			b.WriteString(body)
			break
		}
		end := strings.IndexByte(body[open+1:], '}')
		if end < 0 {
			b.WriteString(body)
			break
		}
		end += open + 1
		b.WriteString(body[:open])
		token := strings.TrimSpace(body[open+1 : end])
		if v, ok := values[token]; ok {
			b.WriteString(v)
		} else {
			b.WriteString(body[open : end+1])
		}
		body = body[end+1:]
	}
	return b.String()
}

var placeholders = []map[string]string{
	{"key": "status", "group": "result"}, {"key": "score", "group": "result"},
	{"key": "user.id", "group": "user"}, {"key": "user.username", "group": "user"}, {"key": "user.displayName", "group": "user"},
	{"key": "mission.name", "group": "mission"}, {"key": "mission.description", "group": "mission"}, {"key": "mission.time", "group": "mission"},
	{"key": "mission.startedAt", "group": "mission"}, {"key": "mission.completedAt", "group": "mission"},
	{"key": "missions.total", "group": "mission"}, {"key": "missions.completed", "group": "mission"}, {"key": "missions.names", "group": "mission"},
	{"key": "pack.total", "group": "pack"}, {"key": "pack.played", "group": "pack"}, {"key": "pack.won", "group": "pack"}, {"key": "pack.names", "group": "pack"},
	{"key": "quiz.total", "group": "quiz"}, {"key": "quiz.correct", "group": "quiz"}, {"key": "quiz.wrong", "group": "quiz"},
}

func webhookAllowed(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return !privateIP(ip)
	}
	return true
}

func privateIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
		if ip4[0] == 0 || (ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127) {
			return true
		}
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast()
}

func hintPrice(cfg settings.All, h *hint) settings.Price {
	switch cfg.Hints.Pricing {
	case settings.PriceFree:
		return settings.Price{}
	case settings.PriceFixed:
		return cfg.Hints.FixedPrice
	}
	return h.Price
}

func addCurrencies(dst map[string]int64, src map[string]int64) {
	for k, v := range src {
		if v > 0 {
			dst[k] += v
		}
	}
}
