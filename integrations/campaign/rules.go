package main

import (
	"math"
	"sort"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/campaign/internal/settings"
)

type runResult struct {
	Success bool
	Elapsed float64
	Health  *float64
	Fields  map[string]float64
}

func resultValues(r runResult, timeLimit *float64) map[string]float64 {
	out := map[string]float64{}
	for k, v := range r.Fields {
		out[k] = v
	}
	elapsed := r.Elapsed
	if timeLimit != nil && *timeLimit > 0 && elapsed > *timeLimit {
		elapsed = *timeLimit
	}
	out["elapsedSeconds"] = elapsed
	if timeLimit != nil && *timeLimit > 0 {
		out["timeSavedPercent"] = (*timeLimit - elapsed) / *timeLimit * 100
	} else {
		out["timeSavedPercent"] = 100
	}
	delete(out, "healthPercent")
	if r.Health != nil {
		out["healthPercent"] = *r.Health
	}
	return out
}

func condHolds(c settings.Cond, vals map[string]float64, missingPass bool) bool {
	v, ok := vals[c.Field]
	if !ok {
		switch c.Missing {
		case settings.MissingPass:
			return true
		case settings.MissingFail:
			return false
		}
		return missingPass
	}
	switch c.Op {
	case ">=":
		return v >= c.Value
	case ">":
		return v > c.Value
	case "<=":
		return v <= c.Value
	case "<":
		return v < c.Value
	case "==":
		return v == c.Value
	case "!=":
		return v != c.Value
	}
	return false
}

func computeStars(success bool, vals map[string]float64, rules []settings.StarRule, maxStars, minOnSuccess int, missingPass bool) int {
	if !success {
		return 0
	}
	sorted := append([]settings.StarRule(nil), rules...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Stars > sorted[j].Stars })
	for _, r := range sorted {
		ok := true
		for _, c := range r.Conditions {
			if !condHolds(c, vals, missingPass) {
				ok = false
				break
			}
		}
		if ok {
			return clamp(r.Stars, 1, maxStars)
		}
	}
	return clamp(minOnSuccess, 0, maxStars)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func multiplier(stars int, table []float64) float64 {
	if stars <= 0 || len(table) == 0 {
		return 0
	}
	if stars > len(table) {
		stars = len(table)
	}
	return table[stars-1]
}

func scaled(full int64, mult float64) int64 {
	if full <= 0 || mult <= 0 {
		return 0
	}
	return int64(math.Floor(float64(full)*mult + 1e-9))
}

func payout(full int64, mult float64, claimed int64) int64 {
	v := scaled(full, mult) - claimed
	if v < 0 {
		return 0
	}
	return v
}

type groupState struct {
	LastAt *time.Time
	Day    string
	Count  int
}

func dayKey(now time.Time, loc *time.Location) string {
	return now.In(loc).Format("2006-01-02")
}

func limitCheck(st groupState, cooldown, daily int, now time.Time, loc *time.Location) string {
	if cooldown > 0 && st.LastAt != nil && now.Sub(*st.LastAt) < time.Duration(cooldown)*time.Second {
		return "cooldown"
	}
	if daily > 0 && st.Day == dayKey(now, loc) && st.Count >= daily {
		return "daily_limit"
	}
	return ""
}

func nextRewardAt(st groupState, cooldown int, now time.Time) *time.Time {
	if cooldown <= 0 || st.LastAt == nil {
		return nil
	}
	next := st.LastAt.Add(time.Duration(cooldown) * time.Second)
	if !next.After(now) {
		return nil
	}
	return &next
}

func rewardsLeft(st groupState, daily int, now time.Time, loc *time.Location) *int {
	if daily <= 0 {
		return nil
	}
	used := 0
	if st.Day == dayKey(now, loc) {
		used = st.Count
	}
	left := daily - used
	if left < 0 {
		left = 0
	}
	return &left
}

func counted(mode string, modes []string) bool {
	for _, m := range modes {
		if m == mode {
			return true
		}
	}
	return false
}

func chapterPercent(stars, stages, maxStars int) float64 {
	if stages == 0 || maxStars == 0 {
		return 0
	}
	return math.Min(100, float64(stars)*100/float64(stages*maxStars))
}

type criterion struct {
	Field string   `json:"field"`
	Value *float64 `json:"value"`
	Stars int      `json:"stars"`
}

func criteriaTiers(success bool, vals map[string]float64, rules []settings.StarRule, maxStars, minOnSuccess int, missingPass bool) []criterion {
	sorted := append([]settings.StarRule(nil), rules...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Stars > sorted[j].Stars })
	var order []string
	seen := map[string]bool{}
	for _, r := range sorted {
		for _, c := range r.Conditions {
			if !seen[c.Field] {
				seen[c.Field] = true
				order = append(order, c.Field)
			}
		}
	}
	out := []criterion{}
	for _, f := range order {
		tier := 0
		if success {
			tier = clamp(minOnSuccess, 0, maxStars)
			for _, r := range sorted {
				ok, any := true, false
				for _, c := range r.Conditions {
					if c.Field != f {
						continue
					}
					any = true
					if !condHolds(c, vals, missingPass) {
						ok = false
						break
					}
				}
				if any && ok {
					if s := clamp(r.Stars, 1, maxStars); s > tier {
						tier = s
					}
					break
				}
			}
		}
		c := criterion{Field: f, Stars: tier}
		if v, ok := vals[f]; ok {
			r := math.Round(v*100) / 100
			c.Value = &r
		}
		out = append(out, c)
	}
	return out
}

const percentTolerance = 0.005

func reachedPercent(have, need float64) bool {
	return have+percentTolerance+1e-9 >= need
}

func milestoneID(m Milestone) string {
	if m.ID != "" {
		return m.ID
	}
	return ftoa(m.Percent)
}

func milestonePaid(paid []any, m Milestone) bool {
	for _, v := range paid {
		switch x := v.(type) {
		case string:
			if m.ID != "" && x == m.ID {
				return true
			}
		case float64:
			if math.Abs(x-m.Percent) < 1e-6 {
				return true
			}
		}
	}
	return false
}

func paidEntry(m Milestone) any {
	if m.ID != "" {
		return m.ID
	}
	return m.Percent
}
