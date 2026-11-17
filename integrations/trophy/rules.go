package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/heroiclabs/nakama/v3/integrations/trophy/internal/settings"
)

type league struct {
	ID          string            `json:"id"`
	Key         string            `json:"key"`
	Name        map[string]string `json:"name"`
	Description map[string]string `json:"description"`
	IconURL     string            `json:"iconUrl"`
	Color       string            `json:"color"`
	MinTrophies int               `json:"minTrophies"`
	Floor       bool              `json:"floor"`
	ResetTo     *int              `json:"resetTo"`
	PrizeKey    string            `json:"prizeKey"`
	Active      bool              `json:"isActive"`
	Metadata    map[string]any    `json:"metadata"`
	UpdatedBy   string            `json:"updatedBy"`
}

type ladder []league

func newLadder(list []league) ladder {
	out := make(ladder, 0, len(list))
	for _, l := range list {
		if l.Active {
			out = append(out, l)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].MinTrophies < out[j].MinTrophies })
	return out
}

func (l ladder) index(trophies int) int {
	idx := -1
	for i, x := range l {
		if trophies >= x.MinTrophies {
			idx = i
		}
	}
	return idx
}

func (l ladder) at(trophies int) *league {
	if i := l.index(trophies); i >= 0 {
		return &l[i]
	}
	return nil
}

func (l ladder) next(trophies int) *league {
	i := l.index(trophies)
	if i+1 < len(l) {
		return &l[i+1]
	}
	return nil
}

const (
	outcomeWin  = "win"
	outcomeLoss = "loss"
	outcomeDraw = "draw"
)

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func matchDelta(a settings.Amounts, outcome string, mine int, opponent *int) int {
	shift := 0
	if a.Scale && opponent != nil && a.ScalePer > 0 {
		shift = int(math.Round(float64(*opponent-mine) / float64(a.ScalePer) * float64(a.ScaleStep)))
	}
	switch outcome {
	case outcomeWin:
		if !a.Scale {
			return a.Win
		}
		return clamp(a.Win+shift, a.MinWin, a.MaxWin)
	case outcomeLoss:
		if !a.Scale {
			return -a.Loss
		}
		return -clamp(a.Loss-shift, a.MinLoss, a.MaxLoss)
	}
	return a.Draw
}

func applyDelta(l ladder, floors bool, before, delta int) int {
	after := before + delta
	if after < 0 {
		after = 0
	}
	if delta < 0 && floors {
		if cur := l.at(before); cur != nil && cur.Floor && after < cur.MinTrophies {
			after = cur.MinTrophies
		}
	}
	return after
}

func seasonReset(cfg settings.Season, l ladder, trophies int) int {
	switch cfg.Mode {
	case settings.ResetZero:
		return 0
	case settings.ResetLeague:
		if cur := l.at(trophies); cur != nil && cur.ResetTo != nil && trophies > *cur.ResetTo {
			return *cur.ResetTo
		}
		return trophies
	}
	if trophies <= cfg.Above {
		return trophies
	}
	return cfg.Above + (trophies-cfg.Above)*cfg.KeepPercent/100
}

func claimKey(cfg settings.Prizes, leagueKey string, season int) string {
	if cfg.OncePer == settings.OnceSeason {
		return fmt.Sprintf("%s:s%d", leagueKey, season)
	}
	return leagueKey
}

func truthy(v any) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		s := strings.ToLower(strings.TrimSpace(x))
		return s == "1" || s == "true" || s == "yes"
	case float64:
		return x != 0
	case int:
		return x != 0
	case int64:
		return x != 0
	case json.Number:
		f, _ := x.Float64()
		return f != 0
	}
	return false
}

func str(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case fmt.Stringer:
		return strings.TrimSpace(x.String())
	case float64:
		return fmt.Sprint(int64(x))
	case int, int64:
		return fmt.Sprint(x)
	}
	return ""
}

type matchFacts struct {
	Outcome  string
	MatchID  string
	Opponent string
}

func readMatch(m settings.Match, fields map[string]any) (matchFacts, bool) {
	if m.SkipField != "" && truthy(fields[m.SkipField]) {
		return matchFacts{}, false
	}
	if len(m.Modes) > 0 {
		mode := str(fields[m.ModeField])
		ok := false
		for _, x := range m.Modes {
			if x == mode {
				ok = true
				break
			}
		}
		if !ok {
			return matchFacts{}, false
		}
	}
	f := matchFacts{Outcome: outcomeLoss}
	if m.DrawField != "" && truthy(fields[m.DrawField]) {
		f.Outcome = outcomeDraw
	} else if truthy(fields[m.WonField]) {
		f.Outcome = outcomeWin
	}
	if m.MatchField != "" {
		f.MatchID = str(fields[m.MatchField])
	}
	if m.OpponentField != "" {
		f.Opponent = str(fields[m.OpponentField])
	}
	return f, true
}
