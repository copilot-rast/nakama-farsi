package main

import (
	"reflect"
	"testing"
	"time"
)

func TestScoredPicksMatchLegacyDailyOffers(t *testing.T) {
	at := func(s string) time.Time {
		v, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	period := func(d string) time.Time { return at(d + "T01:30:00Z") }
	old := at("2026-08-24T14:49:00Z")
	rows := []scoredCandidate{
		{Key: "pawn-star", Currency: "coin", Price: 3000, Rarity: "common", CreatedAt: old},
		{Key: "avatar-nurse", Currency: "coin", Price: 3000, Rarity: "common", CreatedAt: old},
		{Key: "pawn-shield", Currency: "coin", Price: 6000, Rarity: "uncommon", CreatedAt: old},
		{Key: "avatar-brute", Currency: "gem", Price: 100, Rarity: "uncommon", CreatedAt: old},
		{Key: "dice-yummy", Currency: "coin", Price: 10000, Rarity: "rare", CreatedAt: at("2026-11-18T10:00:00Z")},
		{Key: "pawn-heart", Currency: "gem", Price: 250, Rarity: "rare", CreatedAt: old},
		{Key: "avatar-dragon", Currency: "gem", Price: 350, Rarity: "epic", CreatedAt: old},
		{Key: "banner-zucchini", Currency: "gem", Price: 350, CreatedAt: old},
		{Key: "pawn-viking", Currency: "coin", Price: 16000, Rarity: "epic", CreatedAt: at("2026-11-20T10:00:00Z")},
	}
	last := map[string]time.Time{"pawn-star": period("2026-11-20"), "avatar-dragon": period("2026-11-16"), "pawn-viking": period("2026-11-10")}
	for i := range rows {
		rows[i].ID, rows[i].HasPrice = rows[i].Key, true
		if v, ok := last[rows[i].Key]; ok {
			x := v
			rows[i].LastOffered = &x
		}
	}
	want := []struct {
		user, date string
		picks      []string
	}{
		{"8f2d6c1e-0b7a-4c39-9e5f-1a2b3c4d5e6f", "2026-11-21", []string{"pawn-viking", "dice-yummy", "pawn-heart"}},
		{"8f2d6c1e-0b7a-4c39-9e5f-1a2b3c4d5e6f", "2026-11-22", []string{"dice-yummy", "pawn-heart", "pawn-star"}},
		{"8f2d6c1e-0b7a-4c39-9e5f-1a2b3c4d5e6f", "2026-12-31", []string{"pawn-star", "dice-yummy", "avatar-dragon"}},
		{"00000000-0000-0000-0000-000000000001", "2026-11-21", []string{"pawn-star", "pawn-viking", "avatar-nurse"}},
		{"00000000-0000-0000-0000-000000000001", "2026-11-22", []string{"avatar-brute", "pawn-viking", "pawn-star"}},
		{"00000000-0000-0000-0000-000000000001", "2026-12-31", []string{"pawn-heart", "avatar-dragon", "avatar-nurse"}},
		{"c0ffee00-1234-4abc-8def-0123456789ab", "2026-11-21", []string{"pawn-heart", "pawn-shield", "avatar-brute"}},
		{"c0ffee00-1234-4abc-8def-0123456789ab", "2026-11-22", []string{"avatar-brute", "avatar-dragon", "pawn-heart"}},
		{"c0ffee00-1234-4abc-8def-0123456789ab", "2026-12-31", []string{"avatar-brute", "pawn-shield", "avatar-nurse"}},
	}
	wallet := map[string]int64{"coin": 7000, "gem": 260}
	for _, w := range want {
		got := pickScored(rows, 3, w.user+":"+w.date, wallet, period(w.date), at(w.date+"T09:00:00Z"), defaultOfferScore())
		if !reflect.DeepEqual(got, w.picks) {
			t.Fatalf("%s %s = %v, legacy %v", w.user, w.date, got, w.picks)
		}
	}
}

func TestLegacyHashAndRandom(t *testing.T) {
	if legacyHash("") != 0x811c9dc5 || legacyHash("a") != 0xe40c292c {
		t.Fatalf("fnv1a = %x %x", legacyHash(""), legacyHash("a"))
	}
	r := mulberry32(1)
	for i := 0; i < 100; i++ {
		if v := r(); v < 0 || v >= 1 {
			t.Fatalf("out of range %v", v)
		}
	}
}

func TestOfferScoreValidation(t *testing.T) {
	if _, ok := defaultOfferScore().valid(); !ok {
		t.Fatal("defaults must be valid")
	}
	bad := defaultOfferScore()
	bad.Floor = 0
	if f, ok := bad.valid(); ok || f != "score.floor" {
		t.Fatalf("floor 0 = %s %v", f, ok)
	}
}
