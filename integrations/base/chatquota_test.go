package main

import (
	"database/sql"
	"testing"
	"time"
)

func TestQuotaPeriod(t *testing.T) {
	teh, err := time.LoadLocation("Asia/Tehran")
	if err != nil {
		t.Skip("no tz data")
	}
	cases := []struct {
		now          string
		hh, mm       int
		day, startAt string
		endAt        string
	}{
		{"2026-11-14T10:00:00Z", 0, 0, "2026-11-14", "2026-11-13T20:30:00Z", "2026-11-14T20:30:00Z"},
		{"2026-11-14T20:29:59Z", 0, 0, "2026-11-14", "2026-11-13T20:30:00Z", "2026-11-14T20:30:00Z"},
		{"2026-11-14T20:30:00Z", 0, 0, "2026-11-15", "2026-11-14T20:30:00Z", "2026-11-15T20:30:00Z"},
		{"2026-11-14T02:00:00Z", 6, 0, "2026-11-13", "2026-11-13T02:30:00Z", "2026-11-14T02:30:00Z"},
		{"2026-11-14T03:00:00Z", 6, 0, "2026-11-14", "2026-11-14T02:30:00Z", "2026-11-15T02:30:00Z"},
		{"2026-12-31T21:00:00Z", 0, 0, "2027-01-01", "2026-12-31T20:30:00Z", "2027-01-01T20:30:00Z"},
	}
	for _, c := range cases {
		now, _ := time.Parse(time.RFC3339, c.now)
		day, start, end := quotaPeriod(now, teh, c.hh, c.mm)
		if day != c.day || start.UTC().Format(time.RFC3339) != c.startAt || end.UTC().Format(time.RFC3339) != c.endAt {
			t.Errorf("%s %02d:%02d = %s %s %s", c.now, c.hh, c.mm, day, start.UTC().Format(time.RFC3339), end.UTC().Format(time.RFC3339))
		}
	}
	day, _, _ := quotaPeriod(time.Date(2026, 3, 1, 23, 59, 0, 0, time.UTC), time.UTC, 0, 0)
	if day != "2026-03-01" {
		t.Fatal(day)
	}
}

func TestParseChatChannel(t *testing.T) {
	a, b := "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"
	if c, ok := parseChatChannel("4."+a+"."+b+".", a); !ok || c.Kind != "dm" || c.Peer != b || c.Key != "u:"+b || c.Mode != 4 {
		t.Fatalf("dm as a: %+v %v", c, ok)
	}
	if c, ok := parseChatChannel("4."+a+"."+b+".", b); !ok || c.Peer != a {
		t.Fatalf("dm as b: %+v", c)
	}
	if _, ok := parseChatChannel("4."+a+"."+b+".", "33333333-3333-3333-3333-333333333333"); ok {
		t.Fatal("stranger in a dm")
	}
	if _, ok := parseChatChannel("4."+a+"."+a+".", a); ok {
		t.Fatal("dm to self")
	}
	if c, ok := parseChatChannel("3."+a+"..", b); !ok || c.Kind != "group" || c.Key != "g:"+a {
		t.Fatalf("group: %+v", c)
	}
	if c, ok := parseChatChannel("2...lobby", b); !ok || c.Kind != "room" || c.Key != "r:lobby" || c.Label != "lobby" {
		t.Fatalf("room: %+v", c)
	}
	for _, bad := range []string{"", "4.x.y.", "2.a..lobby", "3..", "9...x", "4." + a + "." + b + ".z"} {
		if _, ok := parseChatChannel(bad, a); ok {
			t.Fatalf("%q parsed", bad)
		}
	}
	st := defaultChatQuotaSettings()
	if !st.counts(chatTarget{Kind: "dm"}) || st.counts(chatTarget{Kind: "group"}) || st.counts(chatTarget{Kind: "room"}) {
		t.Fatal("default scope is direct messages only")
	}
	st.CountGroups, st.CountRooms = true, true
	if !st.counts(chatTarget{Kind: "group"}) || !st.counts(chatTarget{Kind: "room"}) {
		t.Fatal("groups and rooms on")
	}
}

func n(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }

func TestMergeChatRules(t *testing.T) {
	st := defaultChatQuotaSettings()
	base := mergeChatRules(st, nil)
	if base.Messages != 20 || base.Conversations != 5 || base.Unlimited || base.RuleID != "" {
		t.Fatalf("defaults: %+v", base)
	}
	rules := []chatRuleRow{
		{id: "a", name: "A", msgs: n(10)},
		{id: "b", name: "B", msgs: n(100), convs: n(50)},
		{id: "c", name: "C", unlimited: true},
	}
	got := mergeChatRules(st, rules)
	if got.Messages != 10 || got.Conversations != 5 || got.Unlimited || got.RuleID != "a" {
		t.Fatalf("priority takes the first: %+v", got)
	}
	st.MatchPolicy = "generous"
	got = mergeChatRules(st, rules)
	if !got.Unlimited || got.UnlimitedBy != "rule" || got.RuleID != "c" || got.Messages != 100 || got.Conversations != 50 {
		t.Fatalf("generous: %+v", got)
	}
	got = mergeChatRules(st, rules[:2])
	if got.Unlimited || got.Messages != 100 || got.RuleID != "a" {
		t.Fatalf("generous without unlimited: %+v", got)
	}
	st.MatchPolicy = "priority"
	got = mergeChatRules(st, []chatRuleRow{{id: "c", unlimited: true}, {id: "a", msgs: n(1)}})
	if !got.Unlimited {
		t.Fatal("first rule unlimited")
	}
}

func TestDecideQuota(t *testing.T) {
	type c struct {
		name string
		in   quotaInput
		ok   bool
		code string
		free int
		conv int
		pack int
	}
	cases := []c{
		{"free left", quotaInput{Mode: "messages", MsgLimit: 2, MsgUsed: 1}, true, "", 1, 0, 0},
		{"out, no packs", quotaInput{Mode: "messages", MsgLimit: 2, MsgUsed: 2}, false, CodeChatQuota, 0, 0, 0},
		{"out, packs", quotaInput{Mode: "messages", MsgLimit: 2, MsgUsed: 2, Packs: 3}, true, "", 0, 0, 1},
		{"packs first", quotaInput{Mode: "messages", PacksFirst: true, MsgLimit: 2, MsgUsed: 0, Packs: 3}, true, "", 0, 0, 1},
		{"packs first, none left", quotaInput{Mode: "messages", PacksFirst: true, MsgLimit: 2, MsgUsed: 0}, true, "", 1, 0, 0},
		{"zero daily, packs only", quotaInput{Mode: "messages", MsgLimit: 0, Packs: 1}, true, "", 0, 0, 1},
		{"unlimited", quotaInput{Mode: "messages", Unlimited: true, MsgLimit: 0}, true, "", 0, 0, 0},
		{"free reply", quotaInput{Mode: "messages", Free: true, MsgLimit: 0}, true, "", 0, 0, 0},
		{"conversations: open chat costs nothing", quotaInput{Mode: "conversations", ConvLimit: 1, ConvUsed: 1}, true, "", 0, 0, 0},
		{"conversations: new one counted", quotaInput{Mode: "conversations", NewConv: true, ConvLimit: 2, ConvUsed: 1}, true, "", 0, 1, 0},
		{"conversations: out", quotaInput{Mode: "conversations", NewConv: true, ConvLimit: 1, ConvUsed: 1}, false, CodeChatQuota, 0, 0, 0},
		{"conversations: pack", quotaInput{Mode: "conversations", NewConv: true, ConvLimit: 1, ConvUsed: 1, Packs: 1}, true, "", 0, 0, 1},
		{"both: conversation cap is hard", quotaInput{Mode: "both", NewConv: true, ConvLimit: 1, ConvUsed: 1, MsgLimit: 9, Packs: 5}, false, CodeChatConversationLimit, 0, 0, 0},
		{"both: new conversation counts both", quotaInput{Mode: "both", NewConv: true, ConvLimit: 2, MsgLimit: 9}, true, "", 1, 1, 0},
		{"both: old conversation counts messages", quotaInput{Mode: "both", ConvLimit: 0, ConvUsed: 0, MsgLimit: 1, MsgUsed: 1, Packs: 1}, true, "", 0, 0, 1},
	}
	for _, x := range cases {
		r := decideQuota(x.in)
		if r.OK != x.ok || r.Code != x.code || r.FreeMsg != x.free || r.FreeConv != x.conv || r.Pack != x.pack {
			t.Errorf("%s: %+v", x.name, r)
		}
	}
}

func TestReplyAndConversation(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	none := chatTargetRow{}
	at := func(first, last time.Duration) chatTargetRow {
		return chatTargetRow{found: true, first: t0.Add(first), last: t0.Add(last)}
	}
	if replyIsFree("off", none, at(0, 0)) || replyIsFree("started", none, none) {
		t.Fatal("off or nothing received")
	}
	if !replyIsFree("started", none, at(0, 0)) || !replyIsFree("started", at(time.Hour, 2*time.Hour), at(0, 0)) || replyIsFree("started", at(0, 0), at(time.Hour, time.Hour)) {
		t.Fatal("started: free when they wrote first")
	}
	if !replyIsFree("reply", at(0, time.Hour), at(0, 2*time.Hour)) || replyIsFree("reply", at(0, 3*time.Hour), at(0, 2*time.Hour)) {
		t.Fatal("reply: free only right after their message")
	}
	start := t0.Add(5 * time.Hour)
	if !isNewConversation("day", none, start) || !isNewConversation("day", at(0, time.Hour), start) || isNewConversation("day", at(0, 6*time.Hour), start) {
		t.Fatal("day scope")
	}
	if isNewConversation("ever", at(0, time.Hour), start) || !isNewConversation("ever", none, start) {
		t.Fatal("ever scope")
	}
}

func TestChatQuotaSettingsClean(t *testing.T) {
	st, err := chatQuotaSettings{}.clean()
	if err != nil || st.CountMode != "messages" || st.TimeZone != "Asia/Tehran" || st.ResetTime != "00:00" || st.Order != "free_first" || st.Notice != "none" {
		t.Fatalf("empty fills defaults: %+v %v", st, err)
	}
	bad := []func(*chatQuotaSettings){
		func(s *chatQuotaSettings) { s.CountMode = "words" },
		func(s *chatQuotaSettings) { s.DailyMessages = -1 },
		func(s *chatQuotaSettings) { s.DailyMessages = 100001 },
		func(s *chatQuotaSettings) { s.ResetTime = "24:00" },
		func(s *chatQuotaSettings) { s.ResetTime = "7:00" },
		func(s *chatQuotaSettings) { s.TimeZone = "Mars/Base" },
		func(s *chatQuotaSettings) { s.FriendPolicy = "x" },
		func(s *chatQuotaSettings) { s.ReplyFree = "x" },
		func(s *chatQuotaSettings) { s.Order = "x" },
		func(s *chatQuotaSettings) { s.PackExpiryDays = 4000 },
		func(s *chatQuotaSettings) { s.UnlimitedMetaMode = "x" },
		func(s *chatQuotaSettings) { s.Notice = "sms" },
		func(s *chatQuotaSettings) { s.Message = map[string]string{"fa": string(make([]rune, 301))} },
	}
	for i, f := range bad {
		s := defaultChatQuotaSettings()
		f(&s)
		if _, err := s.clean(); err == nil {
			t.Errorf("bad setting %d accepted", i)
		}
	}
	s := defaultChatQuotaSettings()
	s.PackProducts = []string{" a ", "a", "", "b"}
	s.Message = map[string]string{"fa": " ", "en": "x"}
	out, err := s.clean()
	if err != nil || len(out.PackProducts) != 2 || len(out.Message) != 1 {
		t.Fatalf("keys and texts trimmed: %+v %v", out, err)
	}
	now := time.Unix(1000, 0)
	if !metaUnlimited("positive", 1, now) || metaUnlimited("positive", 0, now) || !metaUnlimited("until", 1001, now) || metaUnlimited("until", 999, now) {
		t.Fatal("meta unlimited modes")
	}
	if !validChatContent(`{"text":"hi"}`) || validChatContent(`"hi"`) || validChatContent(`{bad`) || validChatContent(``) {
		t.Fatal("content check")
	}
}
