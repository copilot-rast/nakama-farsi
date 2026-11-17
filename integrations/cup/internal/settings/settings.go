package settings

import (
	"context"
	"database/sql"
	"regexp"
	"strconv"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

type Fee struct {
	CurrencyKey string `json:"currencyKey"`

	Amount int64 `json:"amount"`
}

type Requirements struct {
	MinTrophies int `json:"minTrophies"`

	MaxTrophies int `json:"maxTrophies"`

	Leagues []string `json:"leagues"`

	TagIDs []string `json:"tagIds"`

	TagsMode string `json:"tagsMode"`

	ExcludeTagIDs []string `json:"excludeTagIds"`
}

type Entry struct {
	Fee Fee `json:"fee"`

	Requirements Requirements `json:"requirements"`

	WithdrawRefundPercent int `json:"withdrawRefundPercent"`
}

type RoundBestOf struct {
	Round string `json:"round"`

	BestOf int `json:"bestOf"`
}

type Timing struct {
	Mode string `json:"mode"`

	WindowMinutes int `json:"windowMinutes"`

	RoundMinutes int `json:"roundMinutes"`

	CheckInMinutes int `json:"checkInMinutes"`

	NoShow string `json:"noShow"`

	Unplayed string `json:"unplayed"`

	ReadySeconds int `json:"readySeconds"`

	JoinSeconds int `json:"joinSeconds"`

	GameMinutes int `json:"gameMinutes"`
}

type Draw struct {
	Rule string `json:"rule"`

	MaxReplays int `json:"maxReplays"`

	Fallback string `json:"fallback"`

	GameDecides bool `json:"gameDecides"`
}

type Reached struct {
	Round int `json:"round"`

	PrizeKey string `json:"prizeKey"`
}

type Shares struct {
	Champion int `json:"champion"`

	RunnerUp int `json:"runnerUp"`

	Semi int `json:"semi"`

	Third int `json:"third"`

	Fourth int `json:"fourth"`
}

type Pool struct {
	Percent int `json:"percent"`

	Base int64 `json:"base"`

	Shares Shares `json:"shares"`
}

type Prizes struct {
	Champion string `json:"champion"`

	RunnerUp string `json:"runnerUp"`

	Semi string `json:"semi"`

	Third string `json:"third"`

	Fourth string `json:"fourth"`

	Reached []Reached `json:"reached"`

	Pool Pool `json:"pool"`
}

type Rules struct {
	Game string `json:"game"`

	GameOptions map[string]any `json:"gameOptions"`

	Capacity int `json:"capacity"`

	MinPlayers int `json:"minPlayers"`

	Underfill string `json:"underfill"`

	Seeding string `json:"seeding"`

	ThirdPlace bool `json:"thirdPlace"`

	BestOf int `json:"bestOf"`

	RoundBestOf []RoundBestOf `json:"roundBestOf"`

	Timing Timing `json:"timing"`

	Draw Draw `json:"draw"`

	Entry Entry `json:"entry"`

	Prizes Prizes `json:"prizes"`

	Rated bool `json:"rated"`
}

type Events struct {
	Notify bool `json:"notify"`

	Code int `json:"code"`

	InApp bool `json:"inApp"`

	Publish bool `json:"publish"`
}

type Display struct {
	ListLimit int `json:"listLimit"`

	HistoryLimit int `json:"historyLimit"`

	FinishedDays int `json:"finishedDays"`
}

type Limits struct {
	MaxCapacity int `json:"maxCapacity"`

	ActivePerPlayer int `json:"activePerPlayer"`

	TickSeconds int `json:"tickSeconds"`
}

type All struct {
	Enabled bool `json:"enabled"`

	Events Events `json:"events"`

	Display Display `json:"display"`

	Limits Limits `json:"limits"`

	Defaults Rules `json:"defaults"`
}

const (
	ModeWindow = "window"
	ModeLive   = "live"

	UnderfillShrink = "shrink"
	UnderfillCancel = "cancel"

	SeedRandom   = "random"
	SeedTrophies = "trophies"

	NoShowBothOut    = "both_out"
	NoShowHigherSeed = "higher_seed"
	NoShowRandom     = "random"

	DrawReplay     = "replay"
	DrawHigherSeed = "higher_seed"
	DrawRandom     = "random"

	TagsAny = "any"
	TagsAll = "all"
)

var (
	KeyPattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{0,63}$`)
	GamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
	IDPattern   = regexp.MustCompile(`^[0-9a-fA-F-]{36}$`)
)

func DefaultRules() Rules {
	return Rules{
		GameOptions: map[string]any{},
		Capacity:    16, MinPlayers: 4, Underfill: UnderfillShrink, Seeding: SeedRandom, BestOf: 1, RoundBestOf: []RoundBestOf{},
		Timing: Timing{Mode: ModeWindow, WindowMinutes: 120, RoundMinutes: 20, CheckInMinutes: 5, NoShow: NoShowBothOut, Unplayed: NoShowHigherSeed,
			ReadySeconds: 180, JoinSeconds: 90, GameMinutes: 30},
		Draw:   Draw{Rule: DrawReplay, MaxReplays: 3, Fallback: DrawHigherSeed, GameDecides: true},
		Entry:  Entry{Requirements: Requirements{Leagues: []string{}, TagIDs: []string{}, TagsMode: TagsAny, ExcludeTagIDs: []string{}}, WithdrawRefundPercent: 100},
		Prizes: Prizes{Reached: []Reached{}, Pool: Pool{Shares: Shares{Champion: 50, RunnerUp: 25, Semi: 12, Third: 0, Fourth: 0}}},
	}
}

func Defaults() All {
	return All{
		Enabled:  true,
		Events:   Events{Notify: true, Code: 2420, InApp: true, Publish: true},
		Display:  Display{ListLimit: 30, HistoryLimit: 20, FinishedDays: 14},
		Limits:   Limits{MaxCapacity: 256, ActivePerPlayer: 3, TickSeconds: 5},
		Defaults: DefaultRules(),
	}
}

func oneOf(v string, list ...string) bool {
	for _, x := range list {
		if v == x {
			return true
		}
	}
	return false
}

func between(v, lo, hi int) bool { return v >= lo && v <= hi }

func (r *Rules) Normalize() {
	d := DefaultRules()
	if r.GameOptions == nil {
		r.GameOptions = map[string]any{}
	}
	if r.RoundBestOf == nil {
		r.RoundBestOf = []RoundBestOf{}
	}
	if r.Prizes.Reached == nil {
		r.Prizes.Reached = []Reached{}
	}
	for _, p := range []*[]string{&r.Entry.Requirements.Leagues, &r.Entry.Requirements.TagIDs, &r.Entry.Requirements.ExcludeTagIDs} {
		if *p == nil {
			*p = []string{}
		}
	}
	if r.Entry.Requirements.TagsMode == "" {
		r.Entry.Requirements.TagsMode = TagsAny
	}
	if r.Underfill == "" {
		r.Underfill = d.Underfill
	}
	if r.Seeding == "" {
		r.Seeding = d.Seeding
	}
	if r.BestOf == 0 {
		r.BestOf = 1
	}
	if r.Timing.Mode == "" {
		r.Timing.Mode = d.Timing.Mode
	}
	if r.Timing.NoShow == "" {
		r.Timing.NoShow = d.Timing.NoShow
	}
	if r.Timing.Unplayed == "" {
		r.Timing.Unplayed = d.Timing.Unplayed
	}
	if r.Draw.Rule == "" {
		r.Draw.Rule = d.Draw.Rule
	}
	if r.Draw.Fallback == "" {
		r.Draw.Fallback = d.Draw.Fallback
	}
}

func (r Rules) BestOfFor(round, rounds int, third bool) int {
	n := r.BestOf
	for _, o := range r.RoundBestOf {
		hit := false
		switch o.Round {
		case "final":
			hit = !third && round == rounds
		case "semi":
			hit = !third && round == rounds-1
		case "third":
			hit = third
		default:
			v, err := strconv.Atoi(o.Round)
			hit = err == nil && !third && v == round
		}
		if hit {
			n = o.BestOf
		}
	}
	if n < 1 {
		n = 1
	}
	return n
}

func (r Rules) Validate(maxCapacity int) (string, string) {
	t := r.Timing
	switch {
	case r.Game != "" && !GamePattern.MatchString(r.Game):
		return "game", "INVALID_KEY"
	case !between(r.Capacity, 2, maxCapacity):
		return "capacity", "VALUE_OUT_OF_RANGE"
	case !between(r.MinPlayers, 2, r.Capacity):
		return "minPlayers", "VALUE_OUT_OF_RANGE"
	case !oneOf(r.Underfill, UnderfillShrink, UnderfillCancel):
		return "underfill", "INVALID_POLICY"
	case !oneOf(r.Seeding, SeedRandom, SeedTrophies):
		return "seeding", "INVALID_POLICY"
	case !between(r.BestOf, 1, 9) || r.BestOf%2 == 0:
		return "bestOf", "CUP_BEST_OF_ODD"
	case len(r.RoundBestOf) > 20:
		return "roundBestOf", "VALUE_OUT_OF_RANGE"
	case !oneOf(t.Mode, ModeWindow, ModeLive):
		return "timing.mode", "INVALID_POLICY"
	case !between(t.WindowMinutes, 5, 10080):
		return "timing.windowMinutes", "VALUE_OUT_OF_RANGE"
	case !between(t.RoundMinutes, 3, 1440):
		return "timing.roundMinutes", "VALUE_OUT_OF_RANGE"
	case !between(t.CheckInMinutes, 1, 120):
		return "timing.checkInMinutes", "VALUE_OUT_OF_RANGE"
	case !oneOf(t.NoShow, NoShowBothOut, NoShowHigherSeed, NoShowRandom):
		return "timing.noShow", "INVALID_POLICY"
	case !oneOf(t.Unplayed, NoShowBothOut, NoShowHigherSeed, NoShowRandom):
		return "timing.unplayed", "INVALID_POLICY"
	case !between(t.ReadySeconds, 20, 3600):
		return "timing.readySeconds", "VALUE_OUT_OF_RANGE"
	case !between(t.JoinSeconds, 15, 900):
		return "timing.joinSeconds", "VALUE_OUT_OF_RANGE"
	case !between(t.GameMinutes, 2, 240):
		return "timing.gameMinutes", "VALUE_OUT_OF_RANGE"
	case !oneOf(r.Draw.Rule, DrawReplay, DrawHigherSeed, DrawRandom):
		return "draw.rule", "INVALID_POLICY"
	case !between(r.Draw.MaxReplays, 0, 10):
		return "draw.maxReplays", "VALUE_OUT_OF_RANGE"
	case !oneOf(r.Draw.Fallback, DrawHigherSeed, DrawRandom):
		return "draw.fallback", "INVALID_POLICY"
	case r.Entry.Fee.Amount < 0 || r.Entry.Fee.Amount > 1e12:
		return "entry.fee.amount", "VALUE_OUT_OF_RANGE"
	case r.Entry.Fee.Amount > 0 && !KeyPattern.MatchString(r.Entry.Fee.CurrencyKey):
		return "entry.fee.currencyKey", "INVALID_KEY"
	case !between(r.Entry.WithdrawRefundPercent, 0, 100):
		return "entry.withdrawRefundPercent", "VALUE_OUT_OF_RANGE"
	case r.Entry.Requirements.MinTrophies < 0 || r.Entry.Requirements.MaxTrophies < 0:
		return "entry.requirements.minTrophies", "VALUE_OUT_OF_RANGE"
	case r.Entry.Requirements.MaxTrophies > 0 && r.Entry.Requirements.MaxTrophies < r.Entry.Requirements.MinTrophies:
		return "entry.requirements.maxTrophies", "VALUE_OUT_OF_RANGE"
	case !oneOf(r.Entry.Requirements.TagsMode, TagsAny, TagsAll):
		return "entry.requirements.tagsMode", "INVALID_POLICY"
	case !between(r.Prizes.Pool.Percent, 0, 100):
		return "prizes.pool.percent", "VALUE_OUT_OF_RANGE"
	case r.Prizes.Pool.Base < 0 || r.Prizes.Pool.Base > 1e12:
		return "prizes.pool.base", "VALUE_OUT_OF_RANGE"
	case r.Prizes.Pool.Base > 0 && !KeyPattern.MatchString(r.Entry.Fee.CurrencyKey):
		return "entry.fee.currencyKey", "INVALID_KEY"
	case len(r.Prizes.Reached) > 10:
		return "prizes.reached", "VALUE_OUT_OF_RANGE"
	}
	for _, o := range r.RoundBestOf {
		if !between(o.BestOf, 1, 9) || o.BestOf%2 == 0 {
			return "roundBestOf", "CUP_BEST_OF_ODD"
		}
		if !oneOf(o.Round, "final", "semi", "third") {
			if v, err := strconv.Atoi(o.Round); err != nil || !between(v, 1, 10) {
				return "roundBestOf", "INVALID_KEY"
			}
		}
	}
	for _, list := range [][]string{r.Entry.Requirements.TagIDs, r.Entry.Requirements.ExcludeTagIDs} {
		if len(list) > 50 {
			return "entry.requirements.tagIds", "VALUE_OUT_OF_RANGE"
		}
		for _, id := range list {
			if !IDPattern.MatchString(id) {
				return "entry.requirements.tagIds", "INVALID_KEY"
			}
		}
	}
	for _, k := range r.Entry.Requirements.Leagues {
		if !KeyPattern.MatchString(k) {
			return "entry.requirements.leagues", "INVALID_KEY"
		}
	}
	p := r.Prizes
	for name, k := range map[string]string{"prizes.champion": p.Champion, "prizes.runnerUp": p.RunnerUp, "prizes.semi": p.Semi, "prizes.third": p.Third, "prizes.fourth": p.Fourth} {
		if k != "" && !KeyPattern.MatchString(k) {
			return name, "INVALID_KEY"
		}
	}
	for _, x := range p.Reached {
		if !between(x.Round, 1, 10) || !KeyPattern.MatchString(x.PrizeKey) {
			return "prizes.reached", "INVALID_KEY"
		}
	}
	sh := p.Pool.Shares
	for _, v := range []int{sh.Champion, sh.RunnerUp, sh.Semi, sh.Third, sh.Fourth} {
		if !between(v, 0, 100) {
			return "prizes.pool.shares", "VALUE_OUT_OF_RANGE"
		}
	}
	if sh.Champion+sh.RunnerUp+2*sh.Semi+sh.Third+sh.Fourth > 100 {
		return "prizes.pool.shares", "CUP_SHARES_OVER_100"
	}
	return "", ""
}

func (a *All) Normalize() {
	a.Defaults.Normalize()
}

func (a All) Validate() (string, string) {
	switch {
	case !between(a.Events.Code, 0, 1<<30):
		return "events.code", "VALUE_OUT_OF_RANGE"
	case !between(a.Display.ListLimit, 1, 100):
		return "display.listLimit", "VALUE_OUT_OF_RANGE"
	case !between(a.Display.HistoryLimit, 1, 100):
		return "display.historyLimit", "VALUE_OUT_OF_RANGE"
	case !between(a.Display.FinishedDays, 0, 365):
		return "display.finishedDays", "VALUE_OUT_OF_RANGE"
	case !between(a.Limits.MaxCapacity, 2, 1024):
		return "limits.maxCapacity", "VALUE_OUT_OF_RANGE"
	case !between(a.Limits.ActivePerPlayer, 0, 50):
		return "limits.activePerPlayer", "VALUE_OUT_OF_RANGE"
	case !between(a.Limits.TickSeconds, 2, 60):
		return "limits.tickSeconds", "VALUE_OUT_OF_RANGE"
	}
	if f, c := a.Defaults.Validate(a.Limits.MaxCapacity); f != "" {
		return "defaults." + f, c
	}
	return "", ""
}

var store = settingsrow.New("cup_settings", Defaults)

func Load(ctx context.Context, db *sql.DB) (All, error) {
	a, err := store.Load(ctx, db)
	a.Normalize()
	return a, err
}

func Invalidate() { store.Invalidate() }

func Read(ctx context.Context, db *sql.DB) (All, settingsrow.Meta, error) {
	a, m, err := store.Read(ctx, db)
	a.Normalize()
	return a, m, err
}

func Save(ctx context.Context, db *sql.DB, in All, by string) error {
	in.Normalize()
	return store.Save(ctx, db, in, by)
}
