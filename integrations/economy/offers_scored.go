package main

import (
	"encoding/json"
	"math"
	"sort"
	"time"
	"unicode/utf16"
)

type offerScore struct {
	Affordable  float64            `json:"affordable"`
	Rarity      map[string]float64 `json:"rarity"`
	RarityUnset float64            `json:"rarityUnset"`
	Recency     float64            `json:"recency"`
	Newness     float64            `json:"newness"`
	RecencyDays int                `json:"recencyDays"`
	NewDays     int                `json:"newDays"`
	Floor       float64            `json:"floor"`
}

func defaultOfferScore() offerScore {
	return offerScore{Affordable: 20, Rarity: map[string]float64{"common": 10, "uncommon": 10, "rare": 13, "epic": 13}, RarityUnset: 0,
		Recency: 18, Newness: 16, RecencyDays: 7, NewDays: 7, Floor: 1}
}

func (sc offerScore) valid() (string, bool) {
	in := func(v, lo, hi float64) bool { return !math.IsNaN(v) && v >= lo && v <= hi }
	switch {
	case !in(sc.Affordable, 0, 1000):
		return "score.affordable", false
	case !in(sc.RarityUnset, 0, 1000):
		return "score.rarityUnset", false
	case !in(sc.Recency, 0, 1000):
		return "score.recency", false
	case !in(sc.Newness, 0, 1000):
		return "score.newness", false
	case sc.RecencyDays < 1 || sc.RecencyDays > 365:
		return "score.recencyDays", false
	case sc.NewDays < 0 || sc.NewDays > 90:
		return "score.newDays", false
	case !in(sc.Floor, 1, 1000):
		return "score.floor", false
	case len(sc.Rarity) > 20:
		return "score.rarity", false
	}
	for k, v := range sc.Rarity {
		if k == "" || len(k) > 32 || !in(v, 0, 1000) {
			return "score.rarity", false
		}
	}
	return "", true
}

type scoredCandidate struct {
	ID          string
	Key         string
	Currency    string
	Price       int64
	HasPrice    bool
	Rarity      string
	CreatedAt   time.Time
	LastOffered *time.Time
}

func legacyHash(s string) uint32 {
	h := uint32(0x811c9dc5)
	for _, c := range utf16.Encode([]rune(s)) {
		h ^= uint32(c)
		h *= 0x01000193
	}
	return h
}

func mulberry32(seed uint32) func() float64 {
	a := seed
	return func() float64 {
		a += 0x6d2b79f5
		t := a
		t = (t ^ (t >> 15)) * (t | 1)
		t ^= t + (t^(t>>7))*(t|61)
		return float64(t^(t>>14)) / 4294967296
	}
}

func scoreOffer(c scoredCandidate, wallet map[string]int64, periodStart, now time.Time, sc offerScore) float64 {
	affordable := 0.0
	if c.HasPrice && c.Price <= wallet[c.Currency] {
		affordable = 1
	}
	rarity := 0.0
	if c.Rarity == "" {
		rarity = sc.RarityUnset
	} else {
		rarity = sc.Rarity[c.Rarity]
	}
	recency := 1.0
	if c.LastOffered != nil {
		days := math.Round(periodStart.Sub(*c.LastOffered).Hours()) / 24
		recency = math.Min(days/float64(sc.RecencyDays), 1)
	}
	newness := 0.0
	if sc.NewDays > 0 && now.Sub(c.CreatedAt) <= time.Duration(sc.NewDays)*24*time.Hour {
		newness = 1
	}
	raw := sc.Affordable*affordable + rarity + sc.Recency*recency + sc.Newness*newness
	return math.Max(raw, sc.Floor)
}

func pickScored(cands []scoredCandidate, count int, seed string, wallet map[string]int64, periodStart, now time.Time, sc offerScore) []string {
	pool := append([]scoredCandidate(nil), cands...)
	sort.SliceStable(pool, func(i, j int) bool { return utf16Less(pool[i].Key, pool[j].Key) })
	scores := make([]float64, len(pool))
	for i, c := range pool {
		scores[i] = scoreOffer(c, wallet, periodStart, now, sc)
	}
	rng := mulberry32(legacyHash(seed))
	take := count
	if take > len(pool) {
		take = len(pool)
	}
	out := []string{}
	for n := 0; n < take; n++ {
		total := 0.0
		for _, s := range scores {
			total += s
		}
		cursor := rng() * total
		idx := len(scores) - 1
		for i, s := range scores {
			cursor -= s
			if cursor <= 0 {
				idx = i
				break
			}
		}
		out = append(out, pool[idx].ID)
		pool = append(pool[:idx], pool[idx+1:]...)
		scores = append(scores[:idx], scores[idx+1:]...)
	}
	return out
}

func utf16Less(a, b string) bool {
	x, y := utf16.Encode([]rune(a)), utf16.Encode([]rune(b))
	for i := 0; i < len(x) && i < len(y); i++ {
		if x[i] != y[i] {
			return x[i] < y[i]
		}
	}
	return len(x) < len(y)
}

func offerRarity(meta json.RawMessage) string {
	var m map[string]any
	if json.Unmarshal(meta, &m) != nil {
		return ""
	}
	if v, ok := m["rarity"].(string); ok {
		return v
	}
	return ""
}

func offerCreatedAt(meta json.RawMessage, fallback time.Time) time.Time {
	var m map[string]any
	if json.Unmarshal(meta, &m) == nil {
		if v, ok := m["createdAt"].(string); ok {
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t
			}
		}
	}
	return fallback
}
