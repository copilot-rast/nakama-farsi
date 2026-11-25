package stealthbots

import (
	"bytes"
	"encoding/json"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
)

const (
	FriendDecline = "decline"
	FriendPending = "pending"
)

type Settings struct {
	Enabled           bool `json:"enabled"`
	PoolBots          bool `json:"poolBots"`
	SyntheticFallback bool `json:"syntheticFallback"`
	LevelRange        int  `json:"levelRange"`

	ArriveMinSeconds int  `json:"arriveMinSeconds"`
	ArriveMaxSeconds int  `json:"arriveMaxSeconds"`
	ThinkMinMs       int  `json:"thinkMinMs"`
	ThinkMaxMs       int  `json:"thinkMaxMs"`
	PausePercent     int  `json:"pausePercent"`
	PauseMinMs       int  `json:"pauseMinMs"`
	PauseMaxMs       int  `json:"pauseMaxMs"`
	RandomSide       bool `json:"randomSide"`

	OnlineAfterMinMinutes int `json:"onlineAfterMinMinutes"`
	OnlineAfterMaxMinutes int `json:"onlineAfterMaxMinutes"`
	OnlineIdlePercent     int `json:"onlineIdlePercent"`

	Trophies    bool `json:"trophies"`
	FullPrize   bool `json:"fullPrize"`
	Leaderboard bool `json:"leaderboard"`
	Progress    bool `json:"progress"`

	Guard          bool `json:"guard"`
	GuardDailyWins int  `json:"guardDailyWins"`
	GuardBoost     int  `json:"guardBoost"`

	AutoGrow      bool `json:"autoGrow"`
	PoolMax       int  `json:"poolMax"`
	GrowNearLevel bool `json:"growNearLevel"`
	GrowStats     bool `json:"growStats"`
	GhostDays     int  `json:"ghostDays"`

	LookByLevel bool `json:"lookByLevel"`
	LookSpread  int  `json:"lookSpread"`

	OneLiveMatch bool `json:"oneLiveMatch"`
	Presences    bool `json:"presences"`
}

func Defaults() Settings {
	return Settings{
		PoolBots: true, SyntheticFallback: true, LevelRange: 2,
		ArriveMinSeconds: 4, ArriveMaxSeconds: 14,
		ThinkMinMs: 700, ThinkMaxMs: 2600, PausePercent: 6, PauseMinMs: 2500, PauseMaxMs: 6000, RandomSide: true,
		OnlineAfterMinMinutes: 3, OnlineAfterMaxMinutes: 25, OnlineIdlePercent: 4,
		Trophies: true, FullPrize: true, Leaderboard: true, Progress: true,
		GuardDailyWins: 10, GuardBoost: 1,
		AutoGrow: true, PoolMax: 2000, GrowNearLevel: true, GrowStats: true, GhostDays: 90,
		LookByLevel: true, LookSpread: 15,
		OneLiveMatch: true, Presences: true,
	}
}

func Parse(raw []byte) Settings {
	s := Defaults()
	raw = bytes.TrimSpace(raw)
	if len(raw) > 0 && string(raw) != "null" {
		_ = json.Unmarshal(raw, &s)
	}
	return s.Clean()
}

func Of(g *gamecatalog.Game) Settings {
	if g == nil {
		return Defaults()
	}
	return Parse(g.Stealth)
}

func On(g *gamecatalog.Game) bool {
	return Of(g).Enabled
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

func pair(a, b *int, lo, hi int) {
	*a, *b = clamp(*a, lo, hi), clamp(*b, lo, hi)
	if *b < *a {
		*a, *b = *b, *a
	}
}

func (s Settings) Clean() Settings {
	s.LevelRange = clamp(s.LevelRange, 0, 100)
	pair(&s.ArriveMinSeconds, &s.ArriveMaxSeconds, 0, 300)
	pair(&s.ThinkMinMs, &s.ThinkMaxMs, 0, 60000)
	s.PausePercent = clamp(s.PausePercent, 0, 100)
	pair(&s.PauseMinMs, &s.PauseMaxMs, 0, 60000)
	pair(&s.OnlineAfterMinMinutes, &s.OnlineAfterMaxMinutes, 0, 24*60)
	s.OnlineIdlePercent = clamp(s.OnlineIdlePercent, 0, 100)
	s.GuardDailyWins = clamp(s.GuardDailyWins, 0, 10000)
	s.GuardBoost = clamp(s.GuardBoost, 0, 10)
	s.PoolMax = clamp(s.PoolMax, 0, 100000)
	s.GhostDays = clamp(s.GhostDays, 1, 3650)
	s.LookSpread = clamp(s.LookSpread, 0, 100)
	return s
}

func Between(lo, hi int) int {
	if hi <= lo {
		return lo
	}
	return lo + rand.IntN(hi-lo+1)
}

func (s Settings) ArriveDelay() time.Duration {
	return time.Duration(Between(s.ArriveMinSeconds*1000, s.ArriveMaxSeconds*1000)) * time.Millisecond
}

func (s Settings) ThinkDelay() time.Duration {
	d := Between(s.ThinkMinMs, s.ThinkMaxMs)
	if s.PausePercent > 0 && rand.IntN(100) < s.PausePercent {
		d += Between(s.PauseMinMs, s.PauseMaxMs)
	}
	return time.Duration(d) * time.Millisecond
}

func (s Settings) OnlineAfter() time.Duration {
	return time.Duration(Between(s.OnlineAfterMinMinutes*60, s.OnlineAfterMaxMinutes*60)) * time.Second
}

func (s Settings) Side(sides int) int {
	if !s.RandomSide || sides <= 1 {
		return sides - 1
	}
	return rand.IntN(sides)
}

func (s Settings) Boost(winsToday int) int {
	if !s.Enabled || !s.Guard || s.GuardBoost <= 0 || winsToday < s.GuardDailyWins {
		return 0
	}
	return s.GuardBoost
}

func hiddenName(l string) bool {
	switch l {
	case "botsallowed", "humanseats", "botfillseconds", "botafterseconds", "bottimeoutseconds", "bottemplate", "botspay", "botprize", "botpoints", "stealth", "autobot":
		return true
	}
	return false
}

func botPrefixed(k string) bool {
	if len(k) < 3 || !strings.EqualFold(k[:3], "bot") {
		return false
	}
	if len(k) == 3 {
		return true
	}
	c := k[3]
	return c == '_' || c == '-' || (c >= 'A' && c <= 'Z') || (c == 's' && (len(k) == 4 || k[4] == '_' || (k[4] >= 'A' && k[4] <= 'Z')))
}

func hiddenKey(k string) bool {
	return hiddenName(strings.ToLower(k)) || botPrefixed(k)
}

func Scrub(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, x := range t {
			if hiddenKey(k) {
				delete(t, k)
				continue
			}
			t[k] = Scrub(x)
		}
		return t
	case []any:
		for i := range t {
			t[i] = Scrub(t[i])
		}
		return t
	case []map[string]any:
		for i := range t {
			Scrub(t[i])
		}
		return t
	}
	return v
}

func WithoutStealth(settings map[string]any) map[string]any {
	if _, ok := settings["stealth"]; !ok {
		return settings
	}
	out := make(map[string]any, len(settings))
	for k, v := range settings {
		if k != "stealth" {
			out[k] = v
		}
	}
	return out
}

func markerName(l string) bool {
	switch l {
	case "isbot", "bot", "vsbot", "trophyfree", "opponentisbot", "botprize", "botpoints", "botafterseconds", "botfillseconds", "humanseats",
		"botsallowed", "bottimeoutseconds", "botid", "bots", "stealth", "botfillat":
		return true
	}
	return false
}

func markerValue(key, s string) bool {
	l := strings.ToLower(strings.TrimSpace(s))
	switch l {
	case "bot", "online_bot", "robot", "vs_bot", "vsbot":
		return true
	}
	if strings.HasPrefix(l, "bot:") || strings.HasPrefix(l, "bot-") || strings.HasPrefix(l, "bot_") {
		return true
	}
	k := strings.ToLower(key)
	if (k == "mode" || k == "kind" || k == "type" || k == "path") && strings.Contains(l, "bot") {
		return true
	}
	return false
}

func Markers(raw []byte) []string {
	var v any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil
	}
	out := []string{}
	var walk func(path, key string, v any)
	walk = func(path, key string, v any) {
		switch t := v.(type) {
		case map[string]any:
			for k, x := range t {
				l := strings.ToLower(k)
				if markerName(l) || botPrefixed(k) {
					out = append(out, path+"."+k)
				}
				walk(path+"."+k, k, x)
			}
		case []any:
			for _, x := range t {
				walk(path+"[]", key, x)
			}
		case string:
			if markerValue(key, t) {
				out = append(out, path+"="+t)
			}
			trimmed := strings.TrimSpace(t)
			if len(trimmed) > 1 && (trimmed[0] == '{' || trimmed[0] == '[') && json.Valid([]byte(trimmed)) {
				for _, m := range Markers([]byte(trimmed)) {
					out = append(out, path+">"+m)
				}
			}
		}
	}
	walk("$", "", v)
	return out
}
