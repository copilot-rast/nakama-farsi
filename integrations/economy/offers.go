package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"time"
)

// Personal offers: an admin pool of offer templates over shop products. Daily
// templates are drawn per player per period; entry templates fire once per
// player when a trigger holds and run on a personal timer. Pure rules live in
// this file; offers_engine.go holds the database side.

const offerBoughtEvent = "Offer:Bought"

type offerTagCount struct {
	TagID string `json:"tagId"`
	Count int    `json:"count"`
}

type offerSettings struct {
	DailyEnabled        bool            `json:"dailyEnabled"`
	Timezone            string          `json:"timezone"`
	ResetHour           int             `json:"resetHour"`
	ResetMinute         int             `json:"resetMinute"`
	PeriodDays          int             `json:"periodDays"`
	Count               int             `json:"count"`
	CountByTag          []offerTagCount `json:"countByTag"`
	Weighting           string          `json:"weighting"`
	NoRepeatDays        int             `json:"noRepeatDays"`
	FillWithRepeats     bool            `json:"fillWithRepeats"`
	BoughtCooldownDays  int             `json:"boughtCooldownDays"`
	ExcludeOwned        bool            `json:"excludeOwned"`
	RespectProductRules bool            `json:"respectProductRules"`
	InactiveProducts    string          `json:"inactiveProducts"`
	ShowSoldOut         bool            `json:"showSoldOut"`
	GiftCodes           bool            `json:"giftCodes"`
	RerollEnabled       bool            `json:"rerollEnabled"`
	RerollLimit         int             `json:"rerollLimit"`
	RerollCost          *coinPrice      `json:"rerollCost"`
	RerollCostStep      int64           `json:"rerollCostStep"`
	RerollAvoidCurrent  bool            `json:"rerollAvoidCurrent"`
	EntryEnabled        bool            `json:"entryEnabled"`
	EntryMaxActive      int             `json:"entryMaxActive"`
	EntryTimerStart     string          `json:"entryTimerStart"`
	UpdatedAt           string          `json:"updatedAt,omitempty"`
	loc                 *time.Location
}

func defaultOfferSettings() offerSettings {
	return offerSettings{
		DailyEnabled: true, Timezone: defaultTimezone, PeriodDays: 1, Count: 3, CountByTag: []offerTagCount{}, Weighting: "weighted",
		NoRepeatDays: 1, FillWithRepeats: true, BoughtCooldownDays: 7, ExcludeOwned: true, RespectProductRules: true, InactiveProducts: "allow",
		ShowSoldOut: true, RerollEnabled: true, RerollLimit: 1, RerollAvoidCurrent: true,
		EntryEnabled: true, EntryMaxActive: 1, EntryTimerStart: "list",
	}
}

func loadOfferSettings(ctx context.Context, q queryer) (offerSettings, error) {
	st := defaultOfferSettings()
	var raw string
	var updated time.Time
	err := q.QueryRowContext(ctx, `SELECT settings::text, updated_at FROM economy_offer_settings WHERE id=1`).Scan(&raw, &updated)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return st, err
	}
	if err == nil {
		_ = json.Unmarshal([]byte(raw), &st)
		st.UpdatedAt = updated.UTC().Format(time.RFC3339)
	}
	if st.CountByTag == nil {
		st.CountByTag = []offerTagCount{}
	}
	st.loc = loadZone(st.Timezone)
	return st, nil
}

func loadZone(name string) *time.Location {
	if loc, err := time.LoadLocation(name); err == nil && name != "" {
		return loc
	}
	loc, _ := time.LoadLocation(defaultTimezone)
	if loc == nil {
		return time.UTC
	}
	return loc
}

// offerPeriod returns the daily-offer period that contains now: periods last
// PeriodDays days, start at the reset time in the admin time zone and are
// aligned on 1970-01-01 so every server agrees.
func (st offerSettings) offerPeriod(now time.Time) (time.Time, time.Time) {
	loc := st.loc
	if loc == nil {
		loc = loadZone(st.Timezone)
	}
	days := st.PeriodDays
	if days < 1 {
		days = 1
	}
	local := now.In(loc)
	y, m, d := local.Date()
	reset := time.Date(y, m, d, st.ResetHour, st.ResetMinute, 0, 0, loc)
	if local.Before(reset) {
		y, m, d = time.Date(y, m, d-1, 12, 0, 0, 0, loc).Date()
	}
	civil := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	idx := int(floorDiv(int64(civil.Sub(epochDate).Hours()/24), int64(days)))
	sy, sm, sd := epochDate.AddDate(0, 0, idx*days).Date()
	start := time.Date(sy, sm, sd, st.ResetHour, st.ResetMinute, 0, 0, loc)
	end := time.Date(sy, sm, sd+days, st.ResetHour, st.ResetMinute, 0, 0, loc)
	return start, end
}

// dailyCount is the number of daily offers a player gets: the base count, or
// the largest count of a player tag the admin listed.
func (st offerSettings) dailyCount(tags map[string]bool) int {
	n := st.Count
	for _, tc := range st.CountByTag {
		if tags[tc.TagID] && tc.Count > n {
			n = tc.Count
		}
	}
	return n
}

// rerollPrice is the cost of the next reroll after `done` rerolls this period.
func (st offerSettings) rerollPrice(done int) *coinPrice {
	if st.RerollCost == nil || st.RerollCost.CurrencyKey == "" {
		return nil
	}
	amount := st.RerollCost.Amount + st.RerollCostStep*int64(done)
	if amount <= 0 {
		return nil
	}
	return &coinPrice{CurrencyKey: st.RerollCost.CurrencyKey, Amount: amount}
}

type offerMetaCond struct {
	MetaID   string  `json:"metaId"`
	Operator string  `json:"operator"`
	Value    float64 `json:"value"`
}

type offerTemplate struct {
	ID              string            `json:"id"`
	Key             string            `json:"key"`
	Kind            string            `json:"kind"`
	Name            i18n              `json:"name"`
	Description     i18n              `json:"description"`
	Badge           i18n              `json:"badge"`
	ProductID       string            `json:"productId"`
	PriceMode       string            `json:"priceMode"`
	DiscountPercent int               `json:"discountPercent"`
	PriceToman      *int64            `json:"priceToman"`
	CoinPrice       *coinPrice        `json:"coinPrice"`
	SKU             map[string]string `json:"sku"`
	PurchaseLimit   int               `json:"purchaseLimit"`
	Weight          int               `json:"weight"`
	AudienceTagIDs  []string          `json:"audienceTagIds"`
	ExcludeTagIDs   []string          `json:"excludeTagIds"`
	MinLevel        int               `json:"minLevel"`
	MaxLevel        *int              `json:"maxLevel"`
	LevelTableID    *string           `json:"levelTableId"`
	MetaConditions  []offerMetaCond   `json:"metaConditions"`
	MinAccountDays  *int              `json:"minAccountDays"`
	MaxAccountDays  *int              `json:"maxAccountDays"`
	TriggerType     string            `json:"triggerType"`
	TriggerValue    *float64          `json:"triggerValue"`
	TriggerRef      string            `json:"triggerRef"`
	DurationMinutes int               `json:"durationMinutes"`
	Repeat          string            `json:"repeat"`
	CooldownHours   int               `json:"cooldownHours"`
	MaxTriggers     *int              `json:"maxTriggers"`
	StartsAt        *time.Time        `json:"startsAt"`
	EndsAt          *time.Time        `json:"endsAt"`
	IsActive        bool              `json:"isActive"`
	Sort            int               `json:"sort"`
}

func (o offerTemplate) live(now time.Time) bool {
	return o.IsActive && (o.StartsAt == nil || !o.StartsAt.After(now)) && (o.EndsAt == nil || o.EndsAt.After(now))
}

func (o offerTemplate) levelTable() string {
	if o.LevelTableID != nil {
		return *o.LevelTableID
	}
	return ""
}

// offerFacts is what the rules know about one player. Level and Meta are
// lazy so a pool without level or meta rules costs no queries.
type offerFacts struct {
	Tags       map[string]bool
	Level      func(table string) int
	Meta       func(metaID string) float64
	AccountAge time.Duration
	LastLogin  *time.Time
	PrevLogin  *time.Time
	Now        time.Time
}

// offerBlock explains why a template is not offered to a player; "" means it
// passes the template's own audience, level, meta and account-age rules.
func offerBlock(o offerTemplate, f offerFacts) string {
	if !o.live(f.Now) {
		return "inactive"
	}
	if len(o.AudienceTagIDs) > 0 {
		ok := false
		for _, id := range o.AudienceTagIDs {
			if f.Tags[id] {
				ok = true
				break
			}
		}
		if !ok {
			return "audience"
		}
	}
	for _, id := range o.ExcludeTagIDs {
		if f.Tags[id] {
			return "excluded_tag"
		}
	}
	days := int(f.AccountAge.Hours() / 24)
	if o.MinAccountDays != nil && days < *o.MinAccountDays {
		return "account_new"
	}
	if o.MaxAccountDays != nil && days > *o.MaxAccountDays {
		return "account_old"
	}
	if o.MinLevel > 0 || o.MaxLevel != nil {
		lv := f.Level(o.levelTable())
		if lv < o.MinLevel {
			return "level_low"
		}
		if o.MaxLevel != nil && lv > *o.MaxLevel {
			return "level_high"
		}
	}
	for _, c := range o.MetaConditions {
		if !relationHolds(c.Operator, f.Meta(c.MetaID), c.Value) {
			return "meta"
		}
	}
	return ""
}

// entryTriggerHolds checks the entry template's own trigger. A lapsed return
// holds when the player's previous login is at least TriggerValue days older
// than the current one.
func entryTriggerHolds(o offerTemplate, f offerFacts) bool {
	v := 0.0
	if o.TriggerValue != nil {
		v = *o.TriggerValue
	}
	switch o.TriggerType {
	case "", "first_entry":
		return true
	case "level":
		return float64(f.Level(o.TriggerRef)) >= v
	case "meta":
		return f.Meta(o.TriggerRef) >= v
	case "tag":
		return f.Tags[o.TriggerRef]
	case "account_days":
		return f.AccountAge.Hours()/24 >= v
	case "lapsed":
		if f.LastLogin == nil || f.PrevLogin == nil {
			return false
		}
		return f.LastLogin.Sub(*f.PrevLogin).Hours()/24 >= v
	}
	return false
}

// entryHistory is a player's past with one entry template.
type entryHistory struct {
	Count         int
	Open          bool
	LastEnd       *time.Time
	LastTriggered *time.Time
}

// entryDue decides whether an entry template fires now for a player: never
// while a copy is open, at most MaxTriggers times, once unless the repeat is
// a cooldown, and for a lapsed return only on a return newer than the last
// time it fired.
func entryDue(o offerTemplate, h entryHistory, holds bool, f offerFacts) bool {
	if !holds || h.Open {
		return false
	}
	if o.MaxTriggers != nil && h.Count >= *o.MaxTriggers {
		return false
	}
	if h.Count > 0 {
		if o.Repeat != "cooldown" {
			return false
		}
		if h.LastEnd != nil && f.Now.Before(h.LastEnd.Add(time.Duration(o.CooldownHours)*time.Hour)) {
			return false
		}
		if o.TriggerType == "lapsed" && h.LastTriggered != nil && f.LastLogin != nil && !f.LastLogin.After(*h.LastTriggered) {
			return false
		}
	}
	return true
}

type offerCandidate struct {
	ID     string
	Weight int
	Sort   int
}

type offerPick struct {
	Count     int
	Mode      string
	Recent    map[string]bool
	Exclude   map[string]bool
	FillAgain bool
	Seed      int64
}

func offerSeed(parts ...string) int64 {
	h := fnv.New64a()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{0})
	}
	return int64(h.Sum64() & math.MaxInt64)
}

// pickOffers draws up to Count templates: first among the ones not seen in
// the no-repeat window, then (when FillAgain) among the recent ones. Weight 0
// never draws. Weighted draws without replacement by weight, uniform ignores
// weight, priority takes the heaviest first and breaks ties at random. The
// seed makes the draw repeatable, so a preview matches what the player gets.
func pickOffers(cands []offerCandidate, in offerPick) []string {
	rng := rand.New(rand.NewSource(in.Seed))
	sorted := append([]offerCandidate(nil), cands...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Sort != sorted[j].Sort {
			return sorted[i].Sort < sorted[j].Sort
		}
		return sorted[i].ID < sorted[j].ID
	})
	var fresh, again []offerCandidate
	for _, c := range sorted {
		if c.Weight <= 0 || in.Exclude[c.ID] {
			continue
		}
		if in.Recent[c.ID] {
			again = append(again, c)
		} else {
			fresh = append(fresh, c)
		}
	}
	out := drawOffers(fresh, in.Count, in.Mode, rng)
	if len(out) < in.Count && in.FillAgain {
		out = append(out, drawOffers(again, in.Count-len(out), in.Mode, rng)...)
	}
	return out
}

func drawOffers(pool []offerCandidate, n int, mode string, rng *rand.Rand) []string {
	out := []string{}
	if n <= 0 || len(pool) == 0 {
		return out
	}
	pool = append([]offerCandidate(nil), pool...)
	if mode == "priority" {
		keys := make([]float64, len(pool))
		for i := range pool {
			keys[i] = rng.Float64()
		}
		idx := make([]int, len(pool))
		for i := range idx {
			idx[i] = i
		}
		sort.SliceStable(idx, func(a, b int) bool {
			if pool[idx[a]].Weight != pool[idx[b]].Weight {
				return pool[idx[a]].Weight > pool[idx[b]].Weight
			}
			return keys[idx[a]] < keys[idx[b]]
		})
		for _, i := range idx {
			if len(out) == n {
				break
			}
			out = append(out, pool[i].ID)
		}
		return out
	}
	for len(out) < n && len(pool) > 0 {
		total := 0
		for _, c := range pool {
			if mode == "uniform" {
				total++
			} else {
				total += c.Weight
			}
		}
		r := rng.Intn(total)
		pick := 0
		for i, c := range pool {
			w := c.Weight
			if mode == "uniform" {
				w = 1
			}
			if r < w {
				pick = i
				break
			}
			r -= w
		}
		out = append(out, pool[pick].ID)
		pool = append(pool[:pick], pool[pick+1:]...)
	}
	return out
}

// offerPrice is the frozen price of one personal offer.
type offerPrice struct {
	OriginalToman   *int64            `json:"originalToman"`
	FinalToman      *int64            `json:"finalToman"`
	CoinCurrency    string            `json:"coinCurrency,omitempty"`
	CoinAmount      *int64            `json:"coinAmount"`
	CoinOriginal    *int64            `json:"coinOriginal"`
	DiscountPercent int               `json:"discountPercent"`
	SKU             map[string]string `json:"sku"`
}

func i64(v int64) *int64 { return &v }

// priceOffer applies a template to its product. In discount mode the offer
// percent replaces the product's own discount (0 keeps the product price);
// in price mode the special Toman and coin prices replace the product's, and
// an empty one keeps the product's. Stores sell an offer only with an offer
// SKU or when the Toman price did not change, since the store charges the
// SKU's own price.
func priceOffer(o offerTemplate, p productView) offerPrice {
	out := offerPrice{SKU: map[string]string{}}
	var productFinal *int64
	if p.PriceToman != nil {
		out.OriginalToman = i64(*p.PriceToman)
		productFinal = i64(discounted(*p.PriceToman, p.DiscountPercent))
	}
	var coinBase *coinPrice
	if p.CoinPrice != nil && p.CoinPrice.CurrencyKey != "" {
		coinBase = p.CoinPrice
	}
	if o.PriceMode == "price" {
		if o.PriceToman != nil {
			out.FinalToman = i64(*o.PriceToman)
		} else {
			out.FinalToman = productFinal
		}
		if o.CoinPrice != nil && o.CoinPrice.CurrencyKey != "" {
			out.CoinCurrency, out.CoinAmount = o.CoinPrice.CurrencyKey, i64(o.CoinPrice.Amount)
			if coinBase != nil && coinBase.CurrencyKey == o.CoinPrice.CurrencyKey {
				out.CoinOriginal = i64(coinBase.Amount)
			}
		} else if coinBase != nil {
			out.CoinCurrency, out.CoinAmount, out.CoinOriginal = coinBase.CurrencyKey, i64(coinBase.Amount), i64(coinBase.Amount)
		}
		if out.OriginalToman == nil && out.FinalToman != nil {
			out.OriginalToman = i64(*out.FinalToman)
		}
		if out.OriginalToman != nil && out.FinalToman != nil && *out.OriginalToman > 0 && *out.FinalToman < *out.OriginalToman {
			out.DiscountPercent = int((*out.OriginalToman - *out.FinalToman) * 100 / *out.OriginalToman)
		} else if out.CoinOriginal != nil && out.CoinAmount != nil && *out.CoinOriginal > 0 && *out.CoinAmount < *out.CoinOriginal {
			out.DiscountPercent = int((*out.CoinOriginal - *out.CoinAmount) * 100 / *out.CoinOriginal)
		}
	} else {
		pct := o.DiscountPercent
		if pct == 0 {
			pct = p.DiscountPercent
		}
		out.DiscountPercent = pct
		if out.OriginalToman != nil {
			out.FinalToman = i64(discounted(*out.OriginalToman, pct))
		}
		if coinBase != nil {
			out.CoinCurrency, out.CoinOriginal = coinBase.CurrencyKey, i64(coinBase.Amount)
			out.CoinAmount = i64(discounted(coinBase.Amount, o.DiscountPercent))
		}
	}
	unchanged := out.FinalToman != nil && productFinal != nil && *out.FinalToman == *productFinal
	for _, m := range []string{"bazaar", "myket"} {
		if sku := o.SKU[m]; sku != "" {
			out.SKU[m] = sku
		} else if unchanged && p.SKU[m] != "" {
			out.SKU[m] = p.SKU[m]
		}
	}
	return out
}

// offerMethods lists the product's methods an offer can use with its frozen
// price; active filters to the methods switched on.
func offerMethods(p productView, price offerPrice, active map[string]paymentMethod) []string {
	out := []string{}
	for _, m := range p.PaymentMethods {
		if active != nil {
			if _, ok := active[m]; !ok {
				continue
			}
		}
		switch {
		case m == "in_app":
			if price.CoinAmount == nil || price.CoinCurrency == "" {
				continue
			}
		case isStore(m):
			if price.SKU[m] == "" {
				continue
			}
		case isGateway(m):
			if price.FinalToman == nil || (*price.FinalToman != 0 && *price.FinalToman < 100) {
				continue
			}
		}
		out = append(out, m)
	}
	return out
}

func secondsLeft(now time.Time, until *time.Time) int64 {
	if until == nil {
		return 0
	}
	d := until.Sub(now)
	if d <= 0 {
		return 0
	}
	return int64(math.Ceil(d.Seconds()))
}

func periodKey(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}
