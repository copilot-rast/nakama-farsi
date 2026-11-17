package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/heroiclabs/nakama/v3/integrations/clan/internal/settings"
)

var roleRank = map[string]int{settings.RoleLeader: 0, settings.RoleCoLeader: 1, settings.RoleElder: 2, settings.RoleMember: 3}

func allowed(list []string, role string) bool {
	for _, r := range list {
		if r == role {
			return true
		}
	}
	return false
}

func outranks(actor, target string) bool {
	a, ok1 := roleRank[actor]
	b, ok2 := roleRank[target]
	return ok1 && ok2 && a < b
}

func promoteTo(role string) string {
	switch role {
	case settings.RoleMember:
		return settings.RoleElder
	case settings.RoleElder:
		return settings.RoleCoLeader
	}
	return ""
}

func demoteTo(role string) string {
	switch role {
	case settings.RoleCoLeader:
		return settings.RoleElder
	case settings.RoleElder:
		return settings.RoleMember
	}
	return ""
}

func canPromote(actor, target string) (string, bool) {
	to := promoteTo(target)
	if to == "" || !outranks(actor, target) {
		return "", false
	}
	if roleRank[to] <= roleRank[actor] && actor != settings.RoleLeader {
		return "", false
	}
	return to, true
}

func stateOf(role string) int {
	switch role {
	case settings.RoleLeader:
		return 0
	case settings.RoleCoLeader:
		return 1
	}
	return 2
}

type level struct {
	Level     int               `json:"level"`
	XP        int64             `json:"xp"`
	Name      map[string]string `json:"name"`
	IconURL   string            `json:"iconUrl"`
	MemberCap int               `json:"memberCap"`
	PrizeKey  string            `json:"prizeKey"`
	Perks     map[string]any    `json:"perks"`
	UpdatedBy string            `json:"updatedBy"`
}

type road []level

func newRoad(list []level) road {
	out := append(road{}, list...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Level < out[j].Level })
	return out
}

func (r road) levelFor(xp int64) int {
	lv := 1
	for _, l := range r {
		if xp >= l.XP && l.Level > lv {
			lv = l.Level
		}
	}
	return lv
}

func (r road) get(lv int) *level {
	for i := range r {
		if r[i].Level == lv {
			return &r[i]
		}
	}
	return nil
}

func (r road) next(lv int) *level {
	for i := range r {
		if r[i].Level > lv {
			return &r[i]
		}
	}
	return nil
}

func (r road) memberCap(lv int, fallback int) int {
	cap := fallback
	for _, l := range r {
		if l.Level <= lv && l.MemberCap > 0 {
			cap = l.MemberCap
		}
	}
	return cap
}

type PointRule struct {
	Type     string `json:"type"`
	Event    string `json:"event"`
	Field    string `json:"field,omitempty"`
	Equals   string `json:"equals,omitempty"`
	Points   int64  `json:"points"`
	PerValue bool   `json:"perValue,omitempty"`
}

type DuelRule struct {
	Type          string `json:"type"`
	Event         string `json:"event"`
	WonField      string `json:"wonField"`
	DrawField     string `json:"drawField"`
	OpponentField string `json:"opponentField"`
	WinPoints     int64  `json:"winPoints"`
	DrawPoints    int64  `json:"drawPoints"`
}

type Eligibility struct {
	MinMembers   int `json:"minMembers"`
	MaxMembers   int `json:"maxMembers"`
	MinClanLevel int `json:"minClanLevel"`
	MinTrophies  int `json:"minTrophies"`
	JoinedHours  int `json:"joinedHours"`
}

type Prize struct {
	PrizeKey string `json:"prizeKey"`
}

type Prizes struct {
	Win  Prize  `json:"win"`
	Lose Prize  `json:"lose"`
	Draw Prize  `json:"draw"`
	Bye  Prize  `json:"bye"`
	To   string `json:"to"`
}

type WarRules struct {
	Points           []PointRule `json:"points"`
	Duel             DuelRule    `json:"duel"`
	MatchField       string      `json:"matchField"`
	AttacksPerMember int         `json:"attacksPerMember"`
	Eligibility      Eligibility `json:"eligibility"`
	Prizes           Prizes      `json:"prizes"`
	ByeResult        string      `json:"byeResult"`
}

func defaultRules(mode string) WarRules {
	r := WarRules{Points: []PointRule{}, MatchField: "matchId", AttacksPerMember: 0, Eligibility: Eligibility{MinMembers: 1}, Prizes: Prizes{To: "participants"}, ByeResult: "none"}
	if mode == modeDuel {
		r.Duel = DuelRule{Type: "economy_event", WonField: "won", DrawField: "draw", OpponentField: "opponentId", WinPoints: 1}
		r.AttacksPerMember = 3
	}
	return r
}

const (
	modePoints = "points"
	modeDuel   = "duel"
)

func (r *WarRules) normalize(mode string) {
	if r.Points == nil {
		r.Points = []PointRule{}
	}
	if r.MatchField == "" {
		r.MatchField = "matchId"
	}
	if r.Prizes.To == "" {
		r.Prizes.To = "participants"
	}
	if r.ByeResult == "" {
		r.ByeResult = "none"
	}
	if r.Eligibility.MinMembers < 1 {
		r.Eligibility.MinMembers = 1
	}
	if mode == modeDuel {
		if r.Duel.Type == "" {
			r.Duel.Type = "economy_event"
		}
		if r.Duel.WonField == "" {
			r.Duel.WonField = "won"
		}
		if r.Duel.OpponentField == "" {
			r.Duel.OpponentField = "opponentId"
		}
	}
}

func (r WarRules) validate(mode string) (string, string) {
	if mode == modePoints {
		if len(r.Points) == 0 {
			return "rules.points", "CLAN_WAR_POINTS_REQUIRED"
		}
		for _, p := range r.Points {
			switch {
			case p.Type != "economy_event" && p.Type != "game_event":
				return "rules.points", "INVALID_POLICY"
			case !settings.EventPattern.MatchString(p.Event):
				return "rules.points", "INVALID_KEY"
			case p.Field != "" && !settings.FieldPattern.MatchString(p.Field):
				return "rules.points", "INVALID_KEY"
			case p.Points < -1000000 || p.Points > 1000000:
				return "rules.points", "VALUE_OUT_OF_RANGE"
			}
		}
	}
	if mode == modeDuel {
		d := r.Duel
		switch {
		case d.Type != "economy_event" && d.Type != "game_event":
			return "rules.duel.type", "INVALID_POLICY"
		case !settings.EventPattern.MatchString(d.Event):
			return "rules.duel.event", "INVALID_KEY"
		case !settings.FieldPattern.MatchString(d.WonField):
			return "rules.duel.wonField", "INVALID_KEY"
		case d.DrawField != "" && !settings.FieldPattern.MatchString(d.DrawField):
			return "rules.duel.drawField", "INVALID_KEY"
		case !settings.FieldPattern.MatchString(d.OpponentField):
			return "rules.duel.opponentField", "INVALID_KEY"
		case d.WinPoints < 0 || d.WinPoints > 1000000 || d.DrawPoints < 0 || d.DrawPoints > 1000000:
			return "rules.duel.winPoints", "VALUE_OUT_OF_RANGE"
		}
	}
	e := r.Eligibility
	switch {
	case !settings.FieldPattern.MatchString(r.MatchField):
		return "rules.matchField", "INVALID_KEY"
	case r.AttacksPerMember < 0 || r.AttacksPerMember > 10000:
		return "rules.attacksPerMember", "VALUE_OUT_OF_RANGE"
	case e.MinMembers < 1 || e.MinMembers > 1000:
		return "rules.eligibility.minMembers", "VALUE_OUT_OF_RANGE"
	case e.MaxMembers < 0 || e.MaxMembers > 1000 || (e.MaxMembers > 0 && e.MaxMembers < e.MinMembers):
		return "rules.eligibility.maxMembers", "VALUE_OUT_OF_RANGE"
	case e.MinClanLevel < 0 || e.MinClanLevel > 10000:
		return "rules.eligibility.minClanLevel", "VALUE_OUT_OF_RANGE"
	case e.MinTrophies < 0 || e.MinTrophies > 10000000:
		return "rules.eligibility.minTrophies", "VALUE_OUT_OF_RANGE"
	case e.JoinedHours < 0 || e.JoinedHours > 8760:
		return "rules.eligibility.joinedHours", "VALUE_OUT_OF_RANGE"
	case r.Prizes.To != "participants" && r.Prizes.To != "members":
		return "rules.prizes.to", "INVALID_POLICY"
	case r.ByeResult != "none" && r.ByeResult != "win":
		return "rules.byeResult", "INVALID_POLICY"
	}
	return "", ""
}

type entrant struct {
	ID       string
	Level    int
	Trophies int64
	XP       int64
}

type pair struct {
	A, B string
}

func pairClans(list []entrant, mode string, rnd *rand.Rand) []pair {
	cp := append([]entrant{}, list...)
	switch mode {
	case "level":
		sort.SliceStable(cp, func(i, j int) bool {
			if cp[i].Level != cp[j].Level {
				return cp[i].Level > cp[j].Level
			}
			return cp[i].XP > cp[j].XP
		})
	case "trophies":
		sort.SliceStable(cp, func(i, j int) bool { return cp[i].Trophies > cp[j].Trophies })
	default:
		rnd.Shuffle(len(cp), func(i, j int) { cp[i], cp[j] = cp[j], cp[i] })
	}
	out := make([]pair, 0, (len(cp)+1)/2)
	for i := 0; i+1 < len(cp); i += 2 {
		out = append(out, pair{A: cp[i].ID, B: cp[i+1].ID})
	}
	if len(cp)%2 == 1 {
		out = append(out, pair{A: cp[len(cp)-1].ID})
	}
	return out
}

func warResult(scoreA, scoreB int64, bye bool) string {
	switch {
	case bye:
		return "bye"
	case scoreA > scoreB:
		return "a"
	case scoreB > scoreA:
		return "b"
	}
	return "draw"
}

func outcomeFor(result string, sideA bool, byeResult string) string {
	switch result {
	case "bye":
		if byeResult == "win" {
			return "win"
		}
		return "bye"
	case "draw":
		return "draw"
	case "a":
		if sideA {
			return "win"
		}
		return "lose"
	case "b":
		if sideA {
			return "lose"
		}
		return "win"
	}
	return ""
}

func (p Prizes) keyFor(outcome string) string {
	switch outcome {
	case "win":
		return p.Win.PrizeKey
	case "lose":
		return p.Lose.PrizeKey
	case "draw":
		return p.Draw.PrizeKey
	case "bye":
		return p.Bye.PrizeKey
	}
	return ""
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
	case float64:
		return fmt.Sprint(int64(x))
	case int, int64:
		return fmt.Sprint(x)
	case bool:
		return fmt.Sprint(x)
	}
	return ""
}

func number(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	}
	return 0, false
}

func matches(field, equals string, fields map[string]any) bool {
	if field == "" || equals == "" {
		return true
	}
	return str(fields[field]) == equals
}

func xpFor(src settings.XPSource, value float64, fields map[string]any) int64 {
	if !matches(src.Field, src.Equals, fields) {
		return 0
	}
	if !src.PerValue {
		return src.Amount
	}
	v := value
	if src.Field != "" && src.Equals == "" {
		if n, ok := number(fields[src.Field]); ok {
			v = n
		}
	}
	if v <= 0 {
		return 0
	}
	return int64(float64(src.Amount) * v)
}

func pointsFor(p PointRule, value float64, fields map[string]any) int64 {
	if !matches(p.Field, p.Equals, fields) {
		return 0
	}
	if !p.PerValue {
		return p.Points
	}
	v := value
	if p.Field != "" && p.Equals == "" {
		if n, ok := number(fields[p.Field]); ok {
			v = n
		}
	}
	return int64(float64(p.Points) * v)
}

func clipDaily(cap, today, amount int64) int64 {
	if cap <= 0 {
		return amount
	}
	left := cap - today
	if left <= 0 {
		return 0
	}
	if amount > left {
		return left
	}
	return amount
}
