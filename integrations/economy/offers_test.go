package main

import (
	"testing"
	"time"
)

func mustZone(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func TestOfferPeriodTehranMidnight(t *testing.T) {
	st := defaultOfferSettings()
	st.loc = mustZone(t, "Asia/Tehran")
	teh := st.loc
	cases := []struct {
		now, start, end time.Time
	}{
		{time.Date(2026, 11, 14, 10, 0, 0, 0, teh), time.Date(2026, 11, 14, 0, 0, 0, 0, teh), time.Date(2026, 11, 15, 0, 0, 0, 0, teh)},
		{time.Date(2026, 11, 14, 0, 0, 0, 0, teh), time.Date(2026, 11, 14, 0, 0, 0, 0, teh), time.Date(2026, 11, 15, 0, 0, 0, 0, teh)},
		{time.Date(2026, 11, 13, 23, 59, 59, 0, teh), time.Date(2026, 11, 13, 0, 0, 0, 0, teh), time.Date(2026, 11, 14, 0, 0, 0, 0, teh)},
		// 21:00 UTC is already the next day in Tehran (+03:30).
		{time.Date(2026, 11, 14, 21, 0, 0, 0, time.UTC), time.Date(2026, 11, 15, 0, 0, 0, 0, teh), time.Date(2026, 11, 16, 0, 0, 0, 0, teh)},
	}
	for _, c := range cases {
		s, e := st.offerPeriod(c.now)
		if !s.Equal(c.start) || !e.Equal(c.end) {
			t.Fatalf("%v: got %v – %v, want %v – %v", c.now, s, e, c.start, c.end)
		}
	}
}

func TestOfferPeriodResetHourAndLength(t *testing.T) {
	st := defaultOfferSettings()
	st.Timezone, st.loc, st.ResetHour, st.ResetMinute = "UTC", time.UTC, 5, 30
	s, e := st.offerPeriod(time.Date(2026, 11, 14, 5, 29, 0, 0, time.UTC))
	if !s.Equal(time.Date(2026, 11, 13, 5, 30, 0, 0, time.UTC)) || !e.Equal(time.Date(2026, 11, 14, 5, 30, 0, 0, time.UTC)) {
		t.Fatalf("before reset: %v – %v", s, e)
	}
	s, _ = st.offerPeriod(time.Date(2026, 11, 14, 5, 30, 0, 0, time.UTC))
	if !s.Equal(time.Date(2026, 11, 14, 5, 30, 0, 0, time.UTC)) {
		t.Fatalf("at reset: %v", s)
	}
	st.ResetHour, st.ResetMinute, st.PeriodDays = 0, 0, 3
	seen := map[time.Time]int{}
	for d := 0; d < 9; d++ {
		s, e := st.offerPeriod(time.Date(2026, 11, 10+d, 12, 0, 0, 0, time.UTC))
		if e.Sub(s) != 72*time.Hour {
			t.Fatalf("3-day period is %v", e.Sub(s))
		}
		if days := int(s.Sub(epochDate).Hours() / 24); days%3 != 0 {
			t.Fatalf("period not aligned on the epoch: %v", s)
		}
		now := time.Date(2026, 11, 10+d, 12, 0, 0, 0, time.UTC)
		if now.Before(s) || !now.Before(e) {
			t.Fatalf("%v is not inside %v – %v", now, s, e)
		}
		seen[s]++
	}
	if len(seen) < 3 || len(seen) > 4 {
		t.Fatalf("nine days in three-day periods gave %d periods", len(seen))
	}
}

func TestOfferPeriodAcrossDST(t *testing.T) {
	st := defaultOfferSettings()
	st.Timezone, st.loc = "Europe/Berlin", mustZone(t, "Europe/Berlin")
	s, e := st.offerPeriod(time.Date(2026, 3, 29, 12, 0, 0, 0, st.loc))
	if s.Hour() != 0 || e.Hour() != 0 || e.Sub(s) != 23*time.Hour {
		t.Fatalf("spring-forward day: %v – %v (%v)", s, e, e.Sub(s))
	}
}

func TestOfferDailyCountAndRerollPrice(t *testing.T) {
	st := defaultOfferSettings()
	st.CountByTag = []offerTagCount{{TagID: "vip", Count: 5}, {TagID: "new", Count: 2}}
	if n := st.dailyCount(map[string]bool{}); n != 3 {
		t.Fatalf("base count %d", n)
	}
	if n := st.dailyCount(map[string]bool{"vip": true, "new": true}); n != 5 {
		t.Fatalf("largest tag count %d", n)
	}
	if n := st.dailyCount(map[string]bool{"new": true}); n != 3 {
		t.Fatalf("a smaller tag count must not lower the base: %d", n)
	}
	if st.rerollPrice(0) != nil {
		t.Fatal("default reroll is free")
	}
	st.RerollCost, st.RerollCostStep = &coinPrice{CurrencyKey: "gem", Amount: 10}, 5
	if p := st.rerollPrice(2); p == nil || p.Amount != 20 || p.CurrencyKey != "gem" {
		t.Fatalf("third reroll costs %+v", p)
	}
}

func TestPickOffersDeterministicAndRules(t *testing.T) {
	cands := []offerCandidate{{"a", 10, 0}, {"b", 10, 1}, {"c", 10, 2}, {"d", 0, 3}, {"e", 10, 4}}
	in := offerPick{Count: 3, Mode: "weighted", Seed: offerSeed("user", "p1", "0")}
	first := pickOffers(cands, in)
	again := pickOffers(cands, in)
	if len(first) != 3 || len(again) != 3 {
		t.Fatalf("picks %v", first)
	}
	for i := range first {
		if first[i] != again[i] {
			t.Fatalf("same seed, different draw: %v vs %v", first, again)
		}
	}
	seen := map[string]bool{}
	for _, id := range first {
		if id == "d" {
			t.Fatal("weight 0 must never be drawn")
		}
		if seen[id] {
			t.Fatalf("duplicate pick in %v", first)
		}
		seen[id] = true
	}
	// Exclusions are hard, recent ones only fill when allowed.
	in.Exclude = map[string]bool{"a": true}
	in.Recent = map[string]bool{"b": true, "c": true}
	in.FillAgain = true
	got := pickOffers(cands, in)
	if len(got) != 3 || got[0] != "e" {
		t.Fatalf("fresh first then recent: %v", got)
	}
	for _, id := range got {
		if id == "a" {
			t.Fatal("excluded offer drawn")
		}
	}
	in.FillAgain = false
	in.Recent = map[string]bool{"b": true, "c": true}
	if got := pickOffers(cands, in); len(got) != 1 || got[0] != "e" {
		t.Fatalf("no refill: %v", got)
	}
	in.FillAgain = true
	in.Count = 10
	if got := pickOffers(cands, in); len(got) != 3 {
		t.Fatalf("count above the pool returns the pool: %v", got)
	}
}

func TestPickOffersWeightsAndModes(t *testing.T) {
	cands := []offerCandidate{{"heavy", 90, 0}, {"light", 10, 1}}
	heavy := 0
	for i := 0; i < 2000; i++ {
		got := pickOffers(cands, offerPick{Count: 1, Mode: "weighted", Seed: int64(i)})
		if got[0] == "heavy" {
			heavy++
		}
	}
	if heavy < 1650 || heavy > 1950 {
		t.Fatalf("weight 90/10 drew heavy %d of 2000", heavy)
	}
	heavy = 0
	for i := 0; i < 2000; i++ {
		if pickOffers(cands, offerPick{Count: 1, Mode: "uniform", Seed: int64(i)})[0] == "heavy" {
			heavy++
		}
	}
	if heavy < 850 || heavy > 1150 {
		t.Fatalf("uniform drew heavy %d of 2000", heavy)
	}
	prio := []offerCandidate{{"x", 5, 0}, {"y", 50, 1}, {"z", 50, 2}, {"w", 1, 3}}
	for i := 0; i < 50; i++ {
		got := pickOffers(prio, offerPick{Count: 3, Mode: "priority", Seed: int64(i)})
		if len(got) != 3 || got[2] != "x" || (got[0] != "y" && got[0] != "z") {
			t.Fatalf("priority order %v", got)
		}
	}
	if offerSeed("u", "1", "0") == offerSeed("u", "1", "1") || offerSeed("u1", "0") == offerSeed("u", "10") {
		t.Fatal("seeds must differ per reroll and part boundary")
	}
}

func testFacts(now time.Time) offerFacts {
	return offerFacts{Tags: map[string]bool{"vip": true}, Now: now, AccountAge: 10 * 24 * time.Hour,
		Level: func(table string) int {
			if table == "t2" {
				return 2
			}
			return 7
		},
		Meta: func(id string) float64 { return map[string]float64{"wins": 12}[id] }}
}

func intp(v int) *int { return &v }

func TestOfferBlock(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	f := testFacts(now)
	base := offerTemplate{IsActive: true}
	check := func(name string, o offerTemplate, want string) {
		t.Helper()
		if got := offerBlock(o, f); got != want {
			t.Fatalf("%s: got %q want %q", name, got, want)
		}
	}
	check("plain", base, "")
	o := base
	o.IsActive = false
	check("inactive", o, "inactive")
	o = base
	later := now.Add(time.Hour)
	o.StartsAt = &later
	check("not started", o, "inactive")
	o = base
	o.AudienceTagIDs = []string{"whale"}
	check("audience", o, "audience")
	o.AudienceTagIDs = []string{"whale", "vip"}
	check("audience any", o, "")
	o = base
	o.ExcludeTagIDs = []string{"vip"}
	check("excluded", o, "excluded_tag")
	o = base
	o.MinLevel = 8
	check("level low", o, "level_low")
	o.MinLevel, o.MaxLevel = 1, intp(6)
	check("level high", o, "level_high")
	t2 := "t2"
	o.LevelTableID, o.MaxLevel = &t2, intp(3)
	check("level of another table", o, "")
	o = base
	o.MetaConditions = []offerMetaCond{{MetaID: "wins", Operator: "gte", Value: 10}, {MetaID: "wins", Operator: "lt", Value: 12}}
	check("meta", o, "meta")
	o.MetaConditions = o.MetaConditions[:1]
	check("meta ok", o, "")
	o = base
	o.MinAccountDays = intp(11)
	check("too new", o, "account_new")
	o.MinAccountDays, o.MaxAccountDays = nil, intp(9)
	check("too old", o, "account_old")
}

func TestEntryTriggersAndDue(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	f := testFacts(now)
	v := func(x float64) *float64 { return &x }
	cases := []struct {
		o    offerTemplate
		want bool
	}{
		{offerTemplate{TriggerType: "first_entry"}, true},
		{offerTemplate{TriggerType: "level", TriggerValue: v(7)}, true},
		{offerTemplate{TriggerType: "level", TriggerValue: v(3), TriggerRef: "t2"}, false},
		{offerTemplate{TriggerType: "meta", TriggerRef: "wins", TriggerValue: v(12)}, true},
		{offerTemplate{TriggerType: "meta", TriggerRef: "wins", TriggerValue: v(13)}, false},
		{offerTemplate{TriggerType: "tag", TriggerRef: "vip"}, true},
		{offerTemplate{TriggerType: "tag", TriggerRef: "whale"}, false},
		{offerTemplate{TriggerType: "account_days", TriggerValue: v(10)}, true},
		{offerTemplate{TriggerType: "account_days", TriggerValue: v(11)}, false},
		{offerTemplate{TriggerType: "lapsed", TriggerValue: v(3)}, false},
	}
	for i, c := range cases {
		if got := entryTriggerHolds(c.o, f); got != c.want {
			t.Fatalf("case %d (%s): got %v", i, c.o.TriggerType, got)
		}
	}
	last, prev := now.Add(-time.Hour), now.Add(-5*24*time.Hour)
	f.LastLogin, f.PrevLogin = &last, &prev
	lapsed := offerTemplate{TriggerType: "lapsed", TriggerValue: v(3), Repeat: "cooldown", CooldownHours: 24}
	if !entryTriggerHolds(lapsed, f) {
		t.Fatal("a return after 5 days away is a lapse of 3")
	}
	lapsed.TriggerValue = v(6)
	if entryTriggerHolds(lapsed, f) {
		t.Fatal("5 days away is not a lapse of 6")
	}
	lapsed.TriggerValue = v(3)

	once := offerTemplate{Repeat: "once"}
	if !entryDue(once, entryHistory{}, true, f) {
		t.Fatal("first time fires")
	}
	if entryDue(once, entryHistory{}, false, f) {
		t.Fatal("no trigger, no offer")
	}
	ended := now.Add(-2 * time.Hour)
	if entryDue(once, entryHistory{Count: 1, LastEnd: &ended}, true, f) {
		t.Fatal("once never fires again")
	}
	cool := offerTemplate{Repeat: "cooldown", CooldownHours: 3}
	if entryDue(cool, entryHistory{Count: 1, LastEnd: &ended}, true, f) {
		t.Fatal("still in cooldown")
	}
	if !entryDue(cool, entryHistory{Count: 1, LastEnd: &ended}, true, offerFacts{Now: now.Add(2 * time.Hour)}) {
		t.Fatal("after cooldown it fires again")
	}
	if entryDue(cool, entryHistory{Count: 1, Open: true}, true, f) {
		t.Fatal("never while a copy is open")
	}
	cool.MaxTriggers = intp(1)
	if entryDue(cool, entryHistory{Count: 1, LastEnd: &ended}, true, offerFacts{Now: now.Add(time.Hour * 48)}) {
		t.Fatal("max triggers reached")
	}
	trig := last.Add(time.Minute)
	if entryDue(lapsed, entryHistory{Count: 1, LastEnd: &ended, LastTriggered: &trig}, true, offerFacts{Now: now.Add(48 * time.Hour), LastLogin: &last, PrevLogin: &prev}) {
		t.Fatal("a lapse fires once per return")
	}
	newer := now.Add(47 * time.Hour)
	if !entryDue(lapsed, entryHistory{Count: 1, LastEnd: &ended, LastTriggered: &trig}, true, offerFacts{Now: now.Add(48 * time.Hour), LastLogin: &newer, PrevLogin: &prev}) {
		t.Fatal("a newer return fires again after the cooldown")
	}
}

func TestPriceOffer(t *testing.T) {
	p := productView{PriceToman: i64(100000), DiscountPercent: 10, CoinPrice: &coinPrice{CurrencyKey: "gem", Amount: 50}, SKU: map[string]string{"bazaar": "pack1"},
		PaymentMethods: []string{"zarinpal", "bazaar", "in_app", "myket"}}
	o := offerTemplate{PriceMode: "discount", DiscountPercent: 40}
	got := priceOffer(o, p)
	if *got.OriginalToman != 100000 || *got.FinalToman != 60000 || got.DiscountPercent != 40 || *got.CoinAmount != 30 || *got.CoinOriginal != 50 {
		t.Fatalf("discount mode %+v", got)
	}
	if len(got.SKU) != 0 {
		t.Fatalf("a changed Toman price cannot use the product SKU: %v", got.SKU)
	}
	if m := offerMethods(p, got, nil); len(m) != 2 || m[0] != "zarinpal" || m[1] != "in_app" {
		t.Fatalf("methods %v", m)
	}
	o.SKU = map[string]string{"bazaar": "pack1_offer"}
	got = priceOffer(o, p)
	if got.SKU["bazaar"] != "pack1_offer" {
		t.Fatalf("offer SKU %v", got.SKU)
	}
	o = offerTemplate{PriceMode: "discount"}
	got = priceOffer(o, p)
	if *got.FinalToman != 90000 || got.DiscountPercent != 10 || got.SKU["bazaar"] != "pack1" || *got.CoinAmount != 50 {
		t.Fatalf("0%% offer keeps the product price %+v", got)
	}
	o = offerTemplate{PriceMode: "price", PriceToman: i64(25000)}
	got = priceOffer(o, p)
	if *got.FinalToman != 25000 || got.DiscountPercent != 75 || *got.CoinAmount != 50 {
		t.Fatalf("special Toman price %+v", got)
	}
	o = offerTemplate{PriceMode: "price", CoinPrice: &coinPrice{CurrencyKey: "gem", Amount: 20}}
	got = priceOffer(o, p)
	if *got.FinalToman != 90000 || *got.CoinAmount != 20 || *got.CoinOriginal != 50 || got.DiscountPercent != 10 {
		t.Fatalf("special coin price %+v", got)
	}
	coinOnly := productView{CoinPrice: &coinPrice{CurrencyKey: "coin", Amount: 400}, PaymentMethods: []string{"in_app", "zibal"}}
	got = priceOffer(offerTemplate{PriceMode: "price", CoinPrice: &coinPrice{CurrencyKey: "coin", Amount: 100}}, coinOnly)
	if got.DiscountPercent != 75 || got.FinalToman != nil {
		t.Fatalf("coin-only discount %+v", got)
	}
	if m := offerMethods(coinOnly, got, nil); len(m) != 1 || m[0] != "in_app" {
		t.Fatalf("no Toman price, no gateway: %v", m)
	}
	free := priceOffer(offerTemplate{PriceMode: "discount", DiscountPercent: 100}, p)
	if *free.FinalToman != 0 {
		t.Fatalf("100%% is free %+v", free)
	}
	if m := offerMethods(p, free, map[string]paymentMethod{"zarinpal": {}}); len(m) != 1 || m[0] != "zarinpal" {
		t.Fatalf("free offer goes through the gateway as free: %v", m)
	}
}

func TestSecondsLeft(t *testing.T) {
	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	later := now.Add(90*time.Second + time.Millisecond)
	if n := secondsLeft(now, &later); n != 91 {
		t.Fatalf("rounds up: %d", n)
	}
	past := now.Add(-time.Second)
	if secondsLeft(now, &past) != 0 || secondsLeft(now, nil) != 0 {
		t.Fatal("past and open-ended are 0")
	}
}
