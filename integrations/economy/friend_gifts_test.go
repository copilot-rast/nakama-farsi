package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/chatfilter"
)

func ni(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }

func TestGiftRulesPriorityAndGenerous(t *testing.T) {
	base := giftLimits{Sender: 5, Recipient: 10, Pair: 2}
	if got := applyGiftRules(base, nil, "priority"); got.Sender != 5 || got.Recipient != 10 || got.Pair != 2 || !got.CanSend {
		t.Fatalf("no rules %+v", got)
	}
	rules := []giftRule{
		{ID: "vip", Name: "VIP", Sender: ni(20), CanSend: true},
		{ID: "new", Name: "New", Sender: ni(1), Pair: ni(0), CanSend: false},
	}
	p := applyGiftRules(base, rules, "priority")
	if p.Sender != 20 || p.Pair != 2 || p.Recipient != 10 || !p.CanSend || p.RuleID != "vip" {
		t.Fatalf("priority %+v", p)
	}
	g := applyGiftRules(base, rules, "generous")
	if g.Sender != 20 || g.Pair != 0 || !g.CanSend {
		t.Fatalf("generous takes the largest and 0 = no limit wins: %+v", g)
	}
	blocked := applyGiftRules(base, []giftRule{{ID: "x", CanSend: false}}, "generous")
	if blocked.CanSend || blocked.Sender != 5 {
		t.Fatalf("a single blocking rule %+v", blocked)
	}
	only := applyGiftRules(base, []giftRule{{ID: "x", CanSend: false, Recipient: ni(3)}}, "priority")
	if only.CanSend || only.Recipient != 3 {
		t.Fatalf("priority block %+v", only)
	}
}

func TestGiftLimitBlock(t *testing.T) {
	l := giftLimits{Sender: 2, Recipient: 3, Pair: 1}
	cases := []struct {
		u    giftUsage
		want string
	}{
		{giftUsage{}, ""},
		{giftUsage{Sender: 2}, "GIFT_SENDER_DAILY_LIMIT"},
		{giftUsage{Sender: 1, Recipient: 3}, "GIFT_RECIPIENT_DAILY_LIMIT"},
		{giftUsage{Sender: 1, Recipient: 1, Pair: 1}, "GIFT_PAIR_DAILY_LIMIT"},
	}
	for _, c := range cases {
		if got := limitBlock(l, c.u); got != c.want {
			t.Fatalf("%+v = %q want %q", c.u, got, c.want)
		}
	}
	if limitBlock(giftLimits{}, giftUsage{Sender: 999, Recipient: 999, Pair: 999}) != "" {
		t.Fatal("0 means no limit")
	}
	if r := remaining(0, 5); r != nil {
		t.Fatal("no limit has no remaining count")
	}
	if r := remaining(3, 5); r == nil || *r != 0 {
		t.Fatal("remaining never goes below zero")
	}
}

func TestGiftRelationBlock(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	since := now.Add(-48 * time.Hour)
	friends := giftRelation{Mine: 0, Theirs: 0, FriendsSince: &since, SenderAge: 100 * 24 * time.Hour, RecipientAge: 100 * 24 * time.Hour, SenderLevel: 5, Mutual: 2}
	st := defaultFriendGiftSettings()
	check := func(name string, st friendGiftSettings, r giftRelation, want string) {
		t.Helper()
		if got := relationBlock(st, r, now); got != want {
			t.Fatalf("%s = %q want %q", name, got, want)
		}
	}
	check("friends", st, friends, "")
	half := friends
	half.Theirs = -1
	check("one-sided request", st, half, "NOT_FRIENDS")
	blocked := friends
	blocked.Theirs = 3
	check("blocked by the recipient", st, blocked, "GIFT_BLOCKED")
	anyone := st
	anyone.Recipients = "any"
	stranger := giftRelation{Mine: -1, Theirs: -1, SenderAge: friends.SenderAge, RecipientAge: friends.RecipientAge}
	check("any player", anyone, stranger, "")
	strangerBlocked := stranger
	strangerBlocked.Mine = 3
	check("any player but blocked", anyone, strangerBlocked, "GIFT_BLOCKED")
	days := st
	days.MinFriendDays = 3
	check("friends too new", days, friends, "FRIENDSHIP_TOO_NEW")
	days.MinFriendDays = 2
	check("friends long enough", days, friends, "")
	mutual := st
	mutual.MinMutualFriends = 3
	check("mutual friends", mutual, friends, "NOT_ENOUGH_MUTUAL_FRIENDS")
	age := st
	age.MinSenderAccountDays = 101
	check("sender too new", age, friends, "SENDER_ACCOUNT_TOO_NEW")
	age = st
	age.MinRecipientAccountDays = 101
	check("recipient too new", age, friends, "RECIPIENT_ACCOUNT_TOO_NEW")
	lv := st
	lv.MinSenderLevel = 6
	check("sender level", lv, friends, "SENDER_LEVEL_TOO_LOW")
	check("gone", st, giftRelation{RecipientGone: true}, "RECIPIENT_NOT_FOUND")
}

func TestGiftMessage(t *testing.T) {
	st := defaultFriendGiftSettings()
	f := chatfilter.Compile([]chatfilter.Rule{{ID: "1", Word: "badword", Action: chatfilter.ActionMask, Match: chatfilter.MatchWord}, {ID: "2", Word: "never", Action: chatfilter.ActionBlock, Match: chatfilter.MatchWord},
		{ID: "3", Word: "watch", Action: chatfilter.ActionFlag, Match: chatfilter.MatchWord}})
	if msg, flagged, err := cleanGiftMessage(st, "  hello  ", f); err != nil || msg != "hello" || flagged {
		t.Fatalf("plain %q %v %v", msg, flagged, err)
	}
	if msg, _, err := cleanGiftMessage(st, "you badword", f); err != nil || msg == "you badword" {
		t.Fatalf("masked %q %v", msg, err)
	}
	if _, _, err := cleanGiftMessage(st, "never do it", f); err == nil {
		t.Fatal("a block word refuses the message")
	}
	if msg, flagged, err := cleanGiftMessage(st, "watch this", f); err != nil || !flagged || msg != "watch this" {
		t.Fatalf("flag %q %v %v", msg, flagged, err)
	}
	strict := st
	strict.MessageFilter = "block"
	if _, _, err := cleanGiftMessage(strict, "you badword", f); err == nil {
		t.Fatal("block mode refuses any listed word")
	}
	off := st
	off.MessageFilter = "off"
	if msg, _, err := cleanGiftMessage(off, "never do it", f); err != nil || msg != "never do it" {
		t.Fatalf("off %q %v", msg, err)
	}
	short := st
	short.MessageMaxLength = 3
	if _, _, err := cleanGiftMessage(short, "سلام", f); err == nil {
		t.Fatal("length counts letters")
	}
	if _, _, err := cleanGiftMessage(short, "سلا", f); err != nil {
		t.Fatalf("three letters fit: %v", err)
	}
	none := st
	none.MessageEnabled = false
	if _, _, err := cleanGiftMessage(none, "hi", f); err == nil {
		t.Fatal("messages off")
	}
	if msg, _, err := cleanGiftMessage(none, "   ", f); err != nil || msg != "" {
		t.Fatal("an empty message is fine when messages are off")
	}
}

func TestGiftProductOwnership(t *testing.T) {
	st := defaultFriendGiftSettings()
	cases := []struct {
		name string
		st   func(*friendGiftSettings)
		h    productHolding
		want string
	}{
		{"fresh", nil, productHolding{ItemKeys: 2}, ""},
		{"owns all", nil, productHolding{ItemKeys: 2, OwnedKeys: 2}, "RECIPIENT_OWNS_PRODUCT"},
		{"owns part, strict", nil, productHolding{ItemKeys: 2, OwnedKeys: 1}, "RECIPIENT_OWNS_PRODUCT"},
		{"owns part, allowed", func(s *friendGiftSettings) { s.PartialOwned = "allow" }, productHolding{ItemKeys: 2, OwnedKeys: 1}, ""},
		{"prize once", nil, productHolding{PrizeClaimed: true}, "RECIPIENT_OWNS_PRODUCT"},
		{"one-time had", nil, productHolding{OneTimeHad: true}, "RECIPIENT_OWNS_PRODUCT"},
		{"stackable", nil, productHolding{}, ""},
		{"stackable blocked", func(s *friendGiftSettings) { s.StackableProducts = "block" }, productHolding{}, "PRODUCT_NOT_GIFTABLE"},
		{"stackable pending is fine", nil, productHolding{PendingSame: true}, ""},
		{"items pending", nil, productHolding{ItemKeys: 1, PendingSame: true}, "GIFT_ALREADY_PENDING"},
	}
	for _, c := range cases {
		s := st
		if c.st != nil {
			c.st(&s)
		}
		if got := productOwnBlock(s, c.h); got != c.want {
			t.Fatalf("%s = %q want %q", c.name, got, c.want)
		}
	}
}

func TestGiftReturnActionAndGiftable(t *testing.T) {
	st := defaultFriendGiftSettings()
	if returnAction(st, "item", "") != "returned" || returnAction(st, "currency", "") != "returned" {
		t.Fatal("inventory gifts always go back")
	}
	if returnAction(st, "product", "in_app") != "refunded" || returnAction(st, "product", "zibal") != "sender_got" {
		t.Fatal("defaults: coins refunded, money buys go to the sender")
	}
	st.CoinReturn, st.MoneyReturn = "sender", "refund_due"
	if returnAction(st, "product", "in_app") != "sender_got" || returnAction(st, "product", "bazaar") != "refund_due" {
		t.Fatal("custom return")
	}
	st.CoinReturn, st.MoneyReturn = "none", "none"
	if returnAction(st, "product", "in_app") != "kept" || returnAction(st, "product", "zarinpal") != "kept" {
		t.Fatal("none")
	}
	st = defaultFriendGiftSettings()
	yes, no := true, false
	st.ItemTypes = []giftItemType{{TypeID: "skin", Giftable: false}}
	if st.itemGiftable(nil, "skin") || !st.itemGiftable(nil, "car") || !st.itemGiftable(&yes, "skin") || st.itemGiftable(&no, "car") {
		t.Fatal("item flag beats type default beats project default")
	}
	st.ItemsDefault = false
	if st.itemGiftable(nil, "car") {
		t.Fatal("project default")
	}
	if !st.productGiftable(nil) || st.productGiftable(&no) {
		t.Fatal("product flag")
	}
	st.Methods = []string{"in_app", "store"}
	if !st.methodAllowed("in_app") || !st.methodAllowed("myket") || st.methodAllowed("zibal") || st.methodAllowed("nope") {
		t.Fatal("method kinds")
	}
	if _, ok := st.currency("coin"); ok {
		t.Fatal("currency gifts are off by default")
	}
	st.CurrencyEnabled, st.Currencies = true, []giftCurrency{{CurrencyKey: "coin", Min: 1, Max: 10}}
	if c, ok := st.currency("coin"); !ok || c.Max != 10 {
		t.Fatal("listed currency")
	}
}

func TestGiftSendInputKind(t *testing.T) {
	if (giftSendInput{ItemKey: "a"}).kind() != "item" || (giftSendInput{CurrencyKey: "c"}).kind() != "currency" || (giftSendInput{ProductID: "p"}).kind() != "product" {
		t.Fatal("kinds")
	}
	if (giftSendInput{}).kind() != "" || (giftSendInput{ItemKey: "a", ProductID: "p"}).kind() != "" {
		t.Fatal("exactly one source")
	}
}
