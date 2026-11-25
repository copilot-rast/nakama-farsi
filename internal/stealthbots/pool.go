package stealthbots

import (
	"math"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"time"
)

func faNames() []string {
	return []string{"علی", "رضا", "محمد", "امیر", "حسین", "مهدی", "سینا", "آرش", "کیان", "پارسا", "ماهان", "سپهر", "نیما", "پویا", "بردیا", "عرفان", "یاسین", "امید", "حامد", "سعید",
		"مهرداد", "فرهاد", "بهراد", "آرمان", "دانیال", "ایلیا", "متین", "کسری", "سامان", "شایان", "زهرا", "سارا", "مریم", "نگار", "ترانه", "هستی", "آیدا", "پریا", "یلدا", "نازنین"}
}

func enNames() []string {
	return []string{"ali", "reza", "mamad", "amir", "hossein", "mahdi", "sina", "arash", "kian", "parsa", "mahan", "sepehr", "nima", "pouya", "bardia", "erfan", "yasin", "omid", "hamed", "saeed",
		"mehrdad", "farhad", "behrad", "arman", "danial", "ilia", "matin", "kasra", "saman", "shayan", "zahra", "sara", "maryam", "negar", "taraneh", "hasti", "aida", "pariya", "yalda", "nazanin"}
}

func nameTags() []string {
	return []string{"gol", "king", "pro", "star", "_fc", "10", "7", "_ir", "shoot", "striker", "goalie", "ace", "_x", "boss", "tiger", "eagle", "_92", "_1380", "_77", "_99"}
}

func Identity(r *rand.Rand) (string, string) {
	fa, en, tags := faNames(), enNames(), nameTags()
	i := r.IntN(len(fa))
	tag := tags[r.IntN(len(tags))]
	h := r.IntN(1 << 30)
	var name string
	switch r.IntN(5) {
	case 0:
		name = en[i] + tag
	case 1:
		name = en[i] + "_" + strconv.Itoa(1360+h%45)
	case 2:
		name = en[i] + tag + strconv.Itoa(h%900)
	case 3:
		name = fa[i] + "_" + strconv.Itoa(10+h%989)
	default:
		name = en[i] + "." + tag
	}
	name = strings.ReplaceAll(strings.ReplaceAll(name, "__", "_"), "._", ".")
	display := fa[i]
	if r.IntN(3) == 0 {
		display = ""
	}
	return name, display
}

func GrowLevel(r *rand.Rand, level, spread, minLevel int) int {
	lv := level
	if lv <= 0 {
		lv = 1 + int(math.Floor(29*math.Pow(r.Float64(), 1.6)))
	} else if spread > 0 {
		lv += r.IntN(2*spread+1) - spread
	}
	return max(lv, minLevel, 1)
}

func Age(r *rand.Rand) (time.Duration, time.Duration) {
	created := time.Duration(20+r.IntN(400))*24*time.Hour + time.Duration(r.IntN(86400))*time.Second
	updated := time.Duration(r.IntN(15*24)) * time.Hour
	return created, updated
}

type CardStats struct {
	Games, Wins, Losses int
	PlaytimeSeconds     int64
	LastPlayedHours     int
	Trophies, Best      int
}

func SeedStats(r *rand.Rand, level int) CardStats {
	lv := max(level, 1)
	games := 3 + lv*(3+r.IntN(5))
	rate := 0.38 + r.Float64()*0.24
	wins := int(math.Floor(float64(games) * rate))
	t := max(0, lv*22+r.IntN(70)-20)
	return CardStats{Games: games, Wins: wins, Losses: games - wins, PlaytimeSeconds: int64(games * (180 + r.IntN(120))), LastPlayedHours: r.IntN(200), Trophies: t, Best: t + r.IntN(40)}
}

type LookItem struct {
	Key      string
	Rarity   string
	Currency string
	Price    int64
	Unlock   int
	Starter  bool
}

func rarityRank(r string) int {
	switch strings.ToLower(r) {
	case "uncommon":
		return 1
	case "rare":
		return 2
	case "epic":
		return 3
	case "legendary":
		return 4
	}
	return 0
}

func RankLook(items []LookItem) []LookItem {
	byCur := map[string][]int64{}
	for _, it := range items {
		byCur[it.Currency] = append(byCur[it.Currency], it.Price)
	}
	for k := range byCur {
		sort.Slice(byCur[k], func(i, j int) bool { return byCur[k][i] < byCur[k][j] })
	}
	pct := func(it LookItem) float64 {
		list := byCur[it.Currency]
		if len(list) <= 1 {
			return 0
		}
		return float64(sort.Search(len(list), func(i int) bool { return list[i] >= it.Price })) / float64(len(list)-1)
	}
	out := append([]LookItem(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Starter != b.Starter {
			return a.Starter
		}
		if ra, rb := rarityRank(a.Rarity), rarityRank(b.Rarity); ra != rb {
			return ra < rb
		}
		if pa, pb := pct(a), pct(b); pa != pb {
			return pa < pb
		}
		return a.Key < b.Key
	})
	return out
}

func PickLook(items []LookItem, level, maxLevel, spread int, r *rand.Rand) string {
	if len(items) == 0 {
		return ""
	}
	ranked := RankLook(items)
	if maxLevel < 2 {
		maxLevel = 30
	}
	f := float64(clamp(level, 1, maxLevel)-1) / float64(maxLevel-1)
	if spread > 0 {
		f += (r.Float64()*2 - 1) * float64(spread) / 100
	}
	f = math.Max(0, math.Min(1, f))
	want := int(math.Round(f * float64(len(ranked)-1)))
	best, gap := "", len(ranked)+1
	for i, it := range ranked {
		if it.Unlock > 0 && it.Unlock > level {
			continue
		}
		d := i - want
		if d < 0 {
			d = -d
		}
		if d < gap {
			best, gap = it.Key, d
		}
	}
	if best == "" {
		best = ranked[0].Key
	}
	return best
}
