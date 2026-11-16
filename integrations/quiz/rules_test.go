package main

import (
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/quiz/internal/settings"
)

func TestPassNeed(t *testing.T) {
	cfg := settings.Defaults()
	k := &pack{}
	if n := passNeed(k, cfg, 10); n != 10 {
		t.Fatalf("legacy default 100%% of 10 = %d", n)
	}
	pct := 75.0
	k.PassPercent = &pct
	if n := passNeed(k, cfg, 10); n != 8 {
		t.Fatalf("75%% of 10 rounds up to 8, got %d", n)
	}
	nine := 9
	k.PassCorrect = &nine
	if n := passNeed(k, cfg, 10); n != 9 {
		t.Fatalf("pass correct wins: %d", n)
	}
	if n := passNeed(k, cfg, 5); n != 5 {
		t.Fatalf("pass correct is capped by the questions asked: %d", n)
	}
	if n := passNeed(k, cfg, 0); n != 0 {
		t.Fatalf("no questions: %d", n)
	}
	if s := scoreOf(9, 10); s != 90 {
		t.Fatalf("score %v", s)
	}
	if s := scoreOf(1, 3); s != 33.33 {
		t.Fatalf("score %v", s)
	}
}

func TestNormalizeAndTextAnswers(t *testing.T) {
	cases := map[string]string{
		"  Pacific   Ocean ": "pacific ocean",
		"علي":                "علی",
		"كتاب":               "کتاب",
		"۱۲۳":                "123",
		"٤٥":                 "45",
		"می‌روم":             "می روم",
		"تهران!":             "تهران",
		"اَلف":               "الف",
	}
	for in, want := range cases {
		if got := normalizeAnswer(in, settings.MatchNormalized); got != want {
			t.Errorf("normalize(%q) = %q, want %q", in, got, want)
		}
	}
	if got := normalizeAnswer(" ABC ", settings.MatchExact); got != "abc" {
		t.Fatalf("exact keeps letters, lower case: %q", got)
	}
	q := &question{Type: "text", Answers: map[string][]string{"en": {"Pacific", "Pacific Ocean"}, "fa": {"آرام", "اقیانوس آرام"}}}
	for _, ok := range []string{"pacific", "PACIFIC  ocean", "اقيانوس آرام", "ارام"} {
		if !checkText(q, ok, settings.MatchNormalized) {
			t.Errorf("%q should match", ok)
		}
	}
	for _, no := range []string{"", "Atlantic", "pacificocean"} {
		if checkText(q, no, settings.MatchNormalized) {
			t.Errorf("%q should not match", no)
		}
	}
}

func TestCheckOptions(t *testing.T) {
	q := &question{Type: "options", Options: []Option{{ID: 0}, {ID: 1, Correct: true}, {ID: 2}, {ID: 3, Correct: true}}}
	if !checkOptions(q, []int{1}, settings.MultiAny) || !checkOptions(q, []int{3}, settings.MultiAny) {
		t.Fatal("any: one right option is enough")
	}
	if checkOptions(q, []int{1, 2}, settings.MultiAny) {
		t.Fatal("any: a wrong option in the choice fails")
	}
	if checkOptions(q, []int{1}, settings.MultiAll) || !checkOptions(q, []int{3, 1}, settings.MultiAll) {
		t.Fatal("all: exactly the right set")
	}
	if checkOptions(q, nil, settings.MultiAny) || checkOptions(q, []int{9}, settings.MultiAny) {
		t.Fatal("empty or unknown choice")
	}
	if ids := correctIDs(q); len(ids) != 2 || ids[0] != 1 || ids[1] != 3 {
		t.Fatalf("correct ids %v", ids)
	}
}

func TestValidOptions(t *testing.T) {
	txt := map[string]string{"fa": "x"}
	if code := validOptions([]Option{{ID: 0, Text: txt, Correct: true}}); code != "QUIZ_OPTIONS_REQUIRED" {
		t.Fatal(code)
	}
	if code := validOptions([]Option{{ID: 0, Text: txt}, {ID: 1, Text: txt}}); code != "QUIZ_CORRECT_OPTION_REQUIRED" {
		t.Fatal(code)
	}
	if code := validOptions([]Option{{ID: 0, Text: txt, Correct: true}, {ID: 0, Text: txt}}); code != "DUPLICATE_OPTION" {
		t.Fatal(code)
	}
	if code := validOptions([]Option{{ID: 0, Text: txt, Correct: true}, {ID: 1}}); code != "QUIZ_OPTION_TEXT_REQUIRED" {
		t.Fatal(code)
	}
	if code := validOptions([]Option{{ID: 0, Text: txt, Correct: true}, {ID: 1, MediaURL: "https://x/y.png"}}); code != "" {
		t.Fatal(code)
	}
}

func TestPickQuestions(t *testing.T) {
	pool := []*question{}
	for i := 0; i < 20; i++ {
		pool = append(pool, &question{ID: string(rune('a' + i)), Sort: i, Type: "options", Options: []Option{{ID: 0}, {ID: 1}, {ID: 2}, {ID: 3}}})
	}
	rng := rand.New(rand.NewSource(1))
	all := pickQuestions(pool, 0, false, false, rng)
	if len(all) != 20 || all[0].ID != "a" || all[19].ID != "t" || all[0].Order[0] != 0 {
		t.Fatalf("all in order: %+v", all[:2])
	}
	some := pickQuestions(pool, 10, false, true, rng)
	if len(some) != 10 {
		t.Fatalf("count %d", len(some))
	}
	seen := map[string]bool{}
	for i, p := range some {
		if seen[p.ID] {
			t.Fatal("a question was drawn twice")
		}
		seen[p.ID] = true
		if i > 0 && some[i-1].ID > p.ID {
			t.Fatal("without question shuffle a drawn subset keeps the pack order")
		}
		if len(p.Order) != 4 {
			t.Fatal("option order missing")
		}
	}
	text := pickQuestions([]*question{{ID: "x", Type: "text"}}, 0, true, true, rng)
	if len(text) != 1 || text[0].Order != nil {
		t.Fatalf("text question has no option order: %+v", text)
	}
}

func TestTimingHelpers(t *testing.T) {
	cfg := settings.Defaults()
	k := &pack{}
	if questionSeconds(nil, k, cfg) != 0 {
		t.Fatal("no time by default")
	}
	cfg.Timing.QuestionSeconds = 20
	if questionSeconds(nil, k, cfg) != 20 {
		t.Fatal("settings time")
	}
	k.QuestionSeconds = 30
	if questionSeconds(nil, k, cfg) != 30 {
		t.Fatal("pack time wins")
	}
	five := 5
	if questionSeconds(&question{TimeSeconds: &five}, k, cfg) != 5 {
		t.Fatal("question time wins")
	}
	start := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	if d := questionDeadline(30, start); d == nil || !d.Equal(start.Add(30*time.Second)) {
		t.Fatalf("deadline %v", d)
	}
	if questionDeadline(0, start) != nil {
		t.Fatal("no deadline without time")
	}
	sess := &sessionRow{Started: start}
	if !clockStart(cfg, sess).Equal(start) {
		t.Fatal("first question starts with the session")
	}
	later := start.Add(time.Minute)
	sess.LastAnswer = &later
	if !clockStart(cfg, sess).Equal(later) {
		t.Fatal("all at once: the clock restarts at the previous answer")
	}
	served := start.Add(2 * time.Minute)
	sess.Served = &served
	cfg.Play.Delivery = settings.DeliveryOneByOne
	if !clockStart(cfg, sess).Equal(served) {
		t.Fatal("one by one: the clock starts when the question was sent")
	}
	tehran, _ := time.LoadLocation("Asia/Tehran")
	if dayKey(time.Date(2026, 1, 1, 21, 0, 0, 0, time.UTC), tehran) != "2026-01-02" {
		t.Fatal("day in zone")
	}
}

func TestRenderTemplate(t *testing.T) {
	out := render("سلام {user.username}، نتیجه: {status} ({score}) {unknown} {", map[string]string{"user.username": "ali", "status": "قبول", "score": "90"})
	if out != "سلام ali، نتیجه: قبول (90) {unknown} {" {
		t.Fatalf("render: %q", out)
	}
	if render("{ score }", map[string]string{"score": "1"}) != "1" {
		t.Fatal("spaces inside the braces")
	}
}

func TestWebhookGuard(t *testing.T) {
	for _, ok := range []string{"https://example.com/hook", "http://8.8.8.8:8080/x"} {
		if !webhookAllowed(ok) {
			t.Errorf("%s should be allowed", ok)
		}
	}
	for _, no := range []string{"", "ftp://example.com", "http://localhost/x", "http://127.0.0.1/", "http://10.1.2.3/", "http://192.168.1.1", "http://172.20.0.1",
		"http://169.254.169.254/latest", "http://[::1]/", "http://svc.internal/", "http://printer.local/", "http://100.64.0.1/", "http://user:pw@example.com/"} {
		if webhookAllowed(no) {
			t.Errorf("%s should be refused", no)
		}
	}
	if !privateIP(net.ParseIP("0.0.0.0")) || privateIP(net.ParseIP("1.1.1.1")) {
		t.Fatal("private ip")
	}
}

func TestHintPriceAndSettings(t *testing.T) {
	cfg := settings.Defaults()
	h := &hint{Price: settings.Price{CurrencyKey: "gem", Amount: 5}}
	if p := hintPrice(cfg, h); p.Amount != 5 {
		t.Fatal("per hint")
	}
	cfg.Hints.Pricing = settings.PriceFixed
	cfg.Hints.FixedPrice = settings.Price{CurrencyKey: "coin", Amount: 100}
	if p := hintPrice(cfg, h); p.CurrencyKey != "coin" || p.Amount != 100 {
		t.Fatal("fixed")
	}
	cfg.Hints.Pricing = settings.PriceFree
	if p := hintPrice(cfg, h); p.Amount != 0 {
		t.Fatal("free")
	}
	if f, code := settings.Defaults().Validate(); code != "" {
		t.Fatalf("defaults invalid: %s %s", f, code)
	}
	bad := settings.Defaults()
	bad.Play.Delivery = "x"
	if f, _ := bad.Validate(); f != "play.delivery" {
		t.Fatal(f)
	}
	bad = settings.Defaults()
	bad.Rewards.DayZone = "Mars/Base"
	if f, _ := bad.Validate(); f != "rewards.dayZone" {
		t.Fatal(f)
	}
	bad = settings.Defaults()
	bad.Hints.Pricing, bad.Hints.FixedPrice = settings.PriceFixed, settings.Price{Amount: 3}
	if f, _ := bad.Validate(); f != "hints.fixedPrice.currencyKey" {
		t.Fatal(f)
	}
}

func TestRowQuestion(t *testing.T) {
	q, _, code := rowQuestion(map[string]string{"Pack Key": "p1", "Question": "Capital of France?", "Option 1": "London", "Option 2": "Paris", "Option 3": "Rome", "Correct Options": "2", "Time": "30"}, "en")
	if code != "" || q.Type != "options" || len(q.Options) != 3 || !q.Options[1].Correct || q.Options[0].Correct || q.PackKey != "p1" || *q.TimeSeconds != 30 {
		t.Fatalf("options row: %+v %s", q, code)
	}
	q, _, code = rowQuestion(map[string]string{"Type": "text", "Question": "Largest ocean?", "Accepted Answers": "Pacific|Pacific Ocean", "Answer Language": "en"}, "fa")
	if code != "" || q.Type != "text" || len(q.Answers["en"]) != 2 || q.Question["fa"] == "" {
		t.Fatalf("text row: %+v %s", q, code)
	}
	if _, field, code := rowQuestion(map[string]string{"Question": "x", "Option 1": "a", "Option 2": "b", "Correct Options": "5"}, "fa"); code == "" || field != "Correct Options" {
		t.Fatalf("out of range: %s %s", field, code)
	}
}
