package main

import (
	"encoding/json"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/leaderboardx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/cronexpr"
)

var (
	uuidPattern  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	eventPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.:]{0,63}$`)
	fieldPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,39}$`)
)

func itoa(n int) string { return strconv.Itoa(n) }

type Source struct {
	Type       string  `json:"type"`
	Event      string  `json:"event,omitempty"`
	Context    string  `json:"context,omitempty"`
	Field      string  `json:"field,omitempty"`
	Calc       string  `json:"calc"`
	Multiplier float64 `json:"multiplier"`
}

type Limits struct {
	ClientSubmit bool `json:"clientSubmit"`

	MaxPerWrite int64 `json:"maxPerWrite"`

	PerMinute int `json:"perMinute"`

	OnExceed string `json:"onExceed"`
}

type Tier struct {
	Key       string            `json:"key"`
	Name      map[string]string `json:"name"`
	IconURL   string            `json:"iconUrl,omitempty"`
	Threshold float64           `json:"threshold"`
}

type RewardRow struct {
	Tier          string           `json:"tier,omitempty"`
	FromRank      int              `json:"fromRank"`
	ToRank        int              `json:"toRank"`
	Currencies    map[string]int64 `json:"currencies,omitempty"`
	PrizeKey      string           `json:"prizeKey,omitempty"`
	PrizeGroupKey string           `json:"prizeGroupKey,omitempty"`
}

func (r RewardRow) empty() bool {
	for _, v := range r.Currencies {
		if v > 0 {
			return false
		}
	}
	return r.PrizeKey == "" && r.PrizeGroupKey == ""
}

type board struct {
	ID             string            `json:"id"`
	Key            string            `json:"key"`
	Name           map[string]string `json:"name"`
	Description    map[string]string `json:"description"`
	IconURL        string            `json:"iconUrl"`
	Sort           int               `json:"sort"`
	Active         bool              `json:"isActive"`
	SortOrder      string            `json:"sortOrder"`
	Operator       string            `json:"operator"`
	Period         string            `json:"period"`
	ResetSchedule  string            `json:"resetSchedule"`
	StartsAt       *time.Time        `json:"startsAt"`
	EndsAt         *time.Time        `json:"endsAt"`
	MaxPeriods     int               `json:"maxPeriods"`
	Source         Source            `json:"source"`
	Limits         Limits            `json:"limits"`
	TierBasis      string            `json:"tierBasis"`
	Tiers          []Tier            `json:"tiers"`
	Rewards        []RewardRow       `json:"rewards"`
	RewardMinScore int64             `json:"rewardMinScore"`
	PayMode        string            `json:"payMode"`
	PeriodsClosed  int               `json:"periodsClosed"`
	FinishedAt     *time.Time        `json:"finishedAt"`
	Metadata       map[string]any    `json:"metadata"`
	Scope          string            `json:"scope"`
	UpdatedBy      string            `json:"updatedBy"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

const boardCols = `id::text, key, name::text, description::text, icon_url, sort, is_active, sort_order, operator, period, reset_schedule,
	starts_at, ends_at, max_periods, source::text, limits::text, tier_basis, tiers::text, rewards::text, reward_min_score, pay_mode,
	periods_closed, finished_at, metadata::text, scope, updated_by, created_at, updated_at`

type scanner interface{ Scan(dest ...any) error }

func scanBoard(row scanner) (*board, error) {
	b := &board{}
	var name, desc, src, lim, tiers, rewards, meta string
	if err := row.Scan(&b.ID, &b.Key, &name, &desc, &b.IconURL, &b.Sort, &b.Active, &b.SortOrder, &b.Operator, &b.Period, &b.ResetSchedule,
		&b.StartsAt, &b.EndsAt, &b.MaxPeriods, &src, &lim, &b.TierBasis, &tiers, &rewards, &b.RewardMinScore, &b.PayMode,
		&b.PeriodsClosed, &b.FinishedAt, &meta, &b.Scope, &b.UpdatedBy, &b.CreatedAt, &b.UpdatedAt); err != nil {
		return nil, err
	}
	unmarshal([]byte(name), &b.Name)
	unmarshal([]byte(desc), &b.Description)
	unmarshal([]byte(src), &b.Source)
	unmarshal([]byte(lim), &b.Limits)
	unmarshal([]byte(tiers), &b.Tiers)
	unmarshal([]byte(rewards), &b.Rewards)
	unmarshal([]byte(meta), &b.Metadata)
	b.normalize()
	return b, nil
}

const (
	scopePlayer = "player"
	scopeClan   = "clan"
)

func (b *board) normalize() {
	if b.Scope == "" {
		b.Scope = scopePlayer
	}
	if b.Name == nil {
		b.Name = map[string]string{}
	}
	if b.Description == nil {
		b.Description = map[string]string{}
	}
	if b.Tiers == nil {
		b.Tiers = []Tier{}
	}
	if b.Rewards == nil {
		b.Rewards = []RewardRow{}
	}
	if b.Metadata == nil {
		b.Metadata = map[string]any{}
	}
	if b.Source.Calc == "" {
		b.Source.Calc = "value"
	}
	if b.Source.Multiplier == 0 {
		b.Source.Multiplier = 1
	}
	if b.Limits.OnExceed == "" {
		b.Limits.OnExceed = "reject"
	}
}

func unmarshal(raw []byte, out any) {
	if len(raw) == 0 {
		return
	}
	_ = json.Unmarshal(raw, out)
}

func mustJSON(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}

func (b *board) payMode(cfg settings.All) string {
	if b.PayMode != "" {
		return b.PayMode
	}
	return cfg.Rewards.PayMode
}

var presets = map[string]string{"none": "", "daily": "0 0 * * *", "weekly": "0 0 * * 1", "monthly": "0 0 1 * *"}

func schedule(period, custom string) (string, bool) {
	if period == "custom" {
		c := strings.TrimSpace(custom)
		if c == "" {
			return "", false
		}
		if _, err := cronexpr.Parse(c); err != nil {
			return "", false
		}
		return c, true
	}
	s, ok := presets[period]
	return s, ok
}

func periodBounds(cron string, now time.Time) (*time.Time, *time.Time) {
	if cron == "" {
		return nil, nil
	}
	expr, err := cronexpr.Parse(cron)
	if err != nil {
		return nil, nil
	}
	now = now.UTC()
	next := expr.Next(now)
	var start *time.Time
	if last := expr.Last(now); !last.IsZero() {
		l := last.UTC()
		start = &l
	}
	if next.IsZero() {
		return start, nil
	}
	n := next.UTC()
	return start, &n
}

func (b *board) status(now time.Time) string {
	switch {
	case !b.Active:
		return "inactive"
	case b.FinishedAt != nil:
		return "ended"
	case b.StartsAt != nil && now.Before(*b.StartsAt):
		return "upcoming"
	case b.EndsAt != nil && !now.Before(*b.EndsAt):
		return "ended"
	}
	return "active"
}

func (b *board) better(a, c float64) bool {
	if b.SortOrder == "asc" {
		return a < c
	}
	return a > c
}

func tierIndex(b *board, rank int, score int64, total int) int {
	if b.TierBasis == "none" || len(b.Tiers) == 0 || rank < 1 {
		return -1
	}
	for i, t := range b.Tiers {
		switch b.TierBasis {
		case "score":
			s := float64(score)
			if s == t.Threshold || b.better(s, t.Threshold) {
				return i
			}
		case "rank":
			if float64(rank) <= t.Threshold {
				return i
			}
		case "percent":
			if total > 0 && rank <= percentRank(t.Threshold, total) {
				return i
			}
		}
	}
	return -1
}

func percentRank(pct float64, total int) int {
	if pct <= 0 || total <= 0 {
		return 0
	}
	n := int(math.Ceil(pct / 100 * float64(total)))
	if n < 1 {
		n = 1
	}
	if n > total {
		n = total
	}
	return n
}

func nextTier(b *board, idx, rank int, score int64, total int) map[string]any {
	if b.TierBasis == "none" || len(b.Tiers) == 0 {
		return nil
	}
	target := len(b.Tiers) - 1
	if idx >= 0 {
		target = idx - 1
	}
	if target < 0 {
		return nil
	}
	t := b.Tiers[target]
	out := map[string]any{"key": t.Key, "index": target}
	switch b.TierBasis {
	case "score":
		need := t.Threshold - float64(score)
		if b.SortOrder == "asc" {
			need = float64(score) - t.Threshold
		}
		out["score"] = t.Threshold
		out["need"] = math.Max(0, need)
	case "rank":
		out["rank"] = int(t.Threshold)
		out["need"] = max(0, rank-int(t.Threshold))
	case "percent":
		r := percentRank(t.Threshold, total)
		out["rank"] = r
		out["need"] = max(0, rank-r)
	}
	return out
}

type placed struct {
	UserID   string
	Username string
	Rank     int
	Score    int64
	Subscore int64
	Tier     string
	TierRank int
}

type prizeRef struct {
	PrizeKey      string `json:"prizeKey,omitempty"`
	PrizeGroupKey string `json:"prizeGroupKey,omitempty"`
	SourceID      string `json:"sourceId"`
	Paid          bool   `json:"paid"`
	Error         string `json:"error,omitempty"`
}

type plan struct {
	Currencies map[string]int64 `json:"currencies,omitempty"`
	Prizes     []prizeRef       `json:"prizes,omitempty"`
	Rows       []int            `json:"rows"`
	WalletDone bool             `json:"walletDone,omitempty"`
	EventsSent bool             `json:"eventsSent,omitempty"`
}

func (p *plan) empty() bool { return p == nil || (len(p.Currencies) == 0 && len(p.Prizes) == 0) }

func assignTiers(b *board, list []placed) {
	counts := map[string]int{}
	for i := range list {
		idx := tierIndex(b, list[i].Rank, list[i].Score, len(list))
		if idx < 0 {
			continue
		}
		k := b.Tiers[idx].Key
		counts[k]++
		list[i].Tier, list[i].TierRank = k, counts[k]
	}
}

func planFor(b *board, p placed, stack, sourcePrefix string) *plan {
	if p.Rank < 1 {
		return nil
	}
	if b.RewardMinScore > 0 {
		if b.SortOrder == "asc" {
			if p.Score > b.RewardMinScore {
				return nil
			}
		} else if p.Score < b.RewardMinScore {
			return nil
		}
	}
	out := &plan{Currencies: map[string]int64{}, Rows: []int{}}
	for i, r := range b.Rewards {
		if r.empty() {
			continue
		}
		rank := p.Rank
		if r.Tier != "" {
			if p.Tier != r.Tier {
				continue
			}
			rank = p.TierRank
		}
		if r.FromRank > 0 && rank < r.FromRank {
			continue
		}
		if r.ToRank > 0 && rank > r.ToRank {
			continue
		}
		out.Rows = append(out.Rows, i)
		for k, v := range r.Currencies {
			if v > 0 {
				out.Currencies[k] += v
			}
		}
		if r.PrizeKey != "" || r.PrizeGroupKey != "" {
			out.Prizes = append(out.Prizes, prizeRef{PrizeKey: r.PrizeKey, PrizeGroupKey: r.PrizeGroupKey, SourceID: sourcePrefix + ":r" + strconv.Itoa(i)})
		}
		if stack == settings.StackFirst {
			break
		}
	}
	if len(out.Currencies) == 0 {
		out.Currencies = nil
	}
	if out.empty() {
		return nil
	}
	return out
}

func rewardPreview(b *board, lang func(map[string]string, string) string) []map[string]any {
	out := make([]map[string]any, 0, len(b.Rewards))
	for _, r := range b.Rewards {
		if r.empty() {
			continue
		}
		cur := []map[string]any{}
		keys := make([]string, 0, len(r.Currencies))
		for k := range r.Currencies {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if r.Currencies[k] > 0 {
				cur = append(cur, map[string]any{"currencyKey": k, "amount": r.Currencies[k]})
			}
		}
		out = append(out, map[string]any{"tier": r.Tier, "fromRank": r.FromRank, "toRank": r.ToRank, "currencies": cur, "prizeKey": r.PrizeKey, "prizeGroupKey": r.PrizeGroupKey})
	}
	return out
}

func currencyList(m map[string]int64) []map[string]any {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]map[string]any, 0, len(keys))
	for _, k := range keys {
		out = append(out, map[string]any{"currencyKey": k, "amount": m[k]})
	}
	return out
}

func scoreDelta(src Source, value float64) int64 {
	if src.Calc == "count" {
		return 1
	}
	m := src.Multiplier
	if m == 0 {
		m = 1
	}
	v := math.Round(value * m)
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	if v > 1e15 {
		v = 1e15
	}
	if v < -1e15 {
		v = -1e15
	}
	return int64(v)
}

func clampWrite(l Limits, v int64) (int64, string) {
	if l.MaxPerWrite <= 0 || v <= l.MaxPerWrite {
		return v, ""
	}
	if l.OnExceed == "clamp" {
		return l.MaxPerWrite, "clamped"
	}
	return v, "too_high"
}
