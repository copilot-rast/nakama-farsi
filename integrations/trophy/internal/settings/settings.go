package settings

import (
	"context"
	"database/sql"
	"regexp"

	"github.com/heroiclabs/nakama/v3/internal/cronexpr"
	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

type Match struct {
	Source string `json:"source"`

	Event string `json:"event"`

	WonField string `json:"wonField"`

	DrawField string `json:"drawField"`

	MatchField string `json:"matchField"`

	OpponentField string `json:"opponentField"`

	ModeField string `json:"modeField"`

	Modes []string `json:"modes"`

	SkipField string `json:"skipField"`
}

type Amounts struct {
	Start int `json:"start"`

	Win int `json:"win"`

	Loss int `json:"loss"`

	Draw int `json:"draw"`

	Scale bool `json:"scale"`

	ScalePer int `json:"scalePer"`

	ScaleStep int `json:"scaleStep"`

	MinWin int `json:"minWin"`

	MaxWin int `json:"maxWin"`

	MinLoss int `json:"minLoss"`

	MaxLoss int `json:"maxLoss"`

	Floors bool `json:"floors"`
}

type Prizes struct {
	OncePer string `json:"oncePer"`

	OnlyRising bool `json:"onlyRising"`

	MaxTries int `json:"maxTries"`
}

type Season struct {
	Enabled bool `json:"enabled"`

	Schedule string `json:"schedule"`

	Mode string `json:"mode"`

	Above int `json:"above"`

	KeepPercent int `json:"keepPercent"`
}

type Matchmaking struct {
	Enabled bool `json:"enabled"`

	Games []string `json:"games"`

	Range int `json:"range"`

	Property string `json:"property"`
}

type Display struct {
	HistoryLimit int `json:"historyLimit"`

	LogDays int `json:"logDays"`

	ShowOpponent bool `json:"showOpponent"`
}

type Events struct {
	Publish bool `json:"publish"`

	Notify bool `json:"notify"`

	NotifyCode int `json:"notifyCode"`

	LeagueCode int `json:"leagueCode"`
}

type All struct {
	Enabled     bool        `json:"enabled"`
	Match       Match       `json:"match"`
	Amounts     Amounts     `json:"amounts"`
	Prizes      Prizes      `json:"prizes"`
	Season      Season      `json:"season"`
	Matchmaking Matchmaking `json:"matchmaking"`
	Display     Display     `json:"display"`
	Events      Events      `json:"events"`
}

const (
	SourceEconomy = "economy_event"
	SourceGame    = "game_event"

	OnceEver   = "ever"
	OnceSeason = "season"

	ResetCut    = "cut"
	ResetLeague = "league"
	ResetZero   = "zero"
)

var (
	KeyPattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{0,63}$`)
	EventPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.:]{0,63}$`)
	FieldPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,39}$`)
)

func Defaults() All {
	return All{
		Enabled: true,
		Match:   Match{Source: SourceEconomy, WonField: "won", DrawField: "draw", MatchField: "matchId", OpponentField: "opponentId", ModeField: "mode", Modes: []string{}},
		Amounts: Amounts{Start: 0, Win: 30, Loss: 20, Draw: 0, ScalePer: 100, ScaleStep: 2, MinWin: 10, MaxWin: 50, MinLoss: 5, MaxLoss: 40, Floors: true},
		Prizes:  Prizes{OncePer: OnceEver, OnlyRising: true, MaxTries: 8},
		Season:  Season{Schedule: "0 0 1 * *", Mode: ResetCut, Above: 1000, KeepPercent: 50},
		Matchmaking: Matchmaking{Games: []string{}, Range: 300, Property: "trophies"},
		Display:     Display{HistoryLimit: 30, LogDays: 90, ShowOpponent: true},
		Events:      Events{Publish: true, Notify: true, NotifyCode: 2401, LeagueCode: 2402},
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

func fieldOK(v string) bool { return v == "" || FieldPattern.MatchString(v) }

func (a *All) Normalize() {
	if a.Match.Modes == nil {
		a.Match.Modes = []string{}
	}
	if a.Matchmaking.Games == nil {
		a.Matchmaking.Games = []string{}
	}
}

func (a All) Validate() (string, string) {
	for _, m := range a.Match.Modes {
		if m == "" || len(m) > 40 {
			return "match.modes", "INVALID_KEY"
		}
	}
	for _, g := range a.Matchmaking.Games {
		if !KeyPattern.MatchString(g) {
			return "matchmaking.games", "INVALID_KEY"
		}
	}
	switch {
	case !oneOf(a.Match.Source, SourceEconomy, SourceGame):
		return "match.source", "INVALID_POLICY"
	case a.Match.Event != "" && !EventPattern.MatchString(a.Match.Event):
		return "match.event", "INVALID_KEY"
	case !fieldOK(a.Match.WonField) || a.Match.WonField == "":
		return "match.wonField", "INVALID_KEY"
	case !fieldOK(a.Match.DrawField):
		return "match.drawField", "INVALID_KEY"
	case !fieldOK(a.Match.MatchField):
		return "match.matchField", "INVALID_KEY"
	case !fieldOK(a.Match.OpponentField):
		return "match.opponentField", "INVALID_KEY"
	case !fieldOK(a.Match.ModeField):
		return "match.modeField", "INVALID_KEY"
	case !fieldOK(a.Match.SkipField):
		return "match.skipField", "INVALID_KEY"
	case !between(a.Amounts.Start, 0, 1000000):
		return "amounts.start", "VALUE_OUT_OF_RANGE"
	case !between(a.Amounts.Win, 0, 100000):
		return "amounts.win", "VALUE_OUT_OF_RANGE"
	case !between(a.Amounts.Loss, 0, 100000):
		return "amounts.loss", "VALUE_OUT_OF_RANGE"
	case !between(a.Amounts.Draw, -100000, 100000):
		return "amounts.draw", "VALUE_OUT_OF_RANGE"
	case !between(a.Amounts.ScalePer, 1, 100000):
		return "amounts.scalePer", "VALUE_OUT_OF_RANGE"
	case !between(a.Amounts.ScaleStep, 0, 100000):
		return "amounts.scaleStep", "VALUE_OUT_OF_RANGE"
	case !between(a.Amounts.MinWin, 0, 100000) || a.Amounts.MaxWin < a.Amounts.MinWin:
		return "amounts.minWin", "VALUE_OUT_OF_RANGE"
	case !between(a.Amounts.MaxWin, 0, 100000):
		return "amounts.maxWin", "VALUE_OUT_OF_RANGE"
	case !between(a.Amounts.MinLoss, 0, 100000) || a.Amounts.MaxLoss < a.Amounts.MinLoss:
		return "amounts.minLoss", "VALUE_OUT_OF_RANGE"
	case !between(a.Amounts.MaxLoss, 0, 100000):
		return "amounts.maxLoss", "VALUE_OUT_OF_RANGE"
	case !oneOf(a.Prizes.OncePer, OnceEver, OnceSeason):
		return "prizes.oncePer", "INVALID_POLICY"
	case !between(a.Prizes.MaxTries, 1, 100):
		return "prizes.maxTries", "VALUE_OUT_OF_RANGE"
	case !oneOf(a.Season.Mode, ResetCut, ResetLeague, ResetZero):
		return "season.mode", "INVALID_POLICY"
	case a.Season.Enabled && a.Season.Schedule == "":
		return "season.schedule", "INVALID_SCHEDULE"
	case a.Season.Schedule != "" && !validCron(a.Season.Schedule):
		return "season.schedule", "INVALID_SCHEDULE"
	case !between(a.Season.Above, 0, 1000000):
		return "season.above", "VALUE_OUT_OF_RANGE"
	case !between(a.Season.KeepPercent, 0, 100):
		return "season.keepPercent", "VALUE_OUT_OF_RANGE"
	case !between(a.Matchmaking.Range, 1, 1000000):
		return "matchmaking.range", "VALUE_OUT_OF_RANGE"
	case !FieldPattern.MatchString(a.Matchmaking.Property):
		return "matchmaking.property", "INVALID_KEY"
	case !between(a.Display.HistoryLimit, 1, 200):
		return "display.historyLimit", "VALUE_OUT_OF_RANGE"
	case !between(a.Display.LogDays, 1, 3650):
		return "display.logDays", "VALUE_OUT_OF_RANGE"
	case !between(a.Events.NotifyCode, 0, 1<<30):
		return "events.notifyCode", "VALUE_OUT_OF_RANGE"
	case !between(a.Events.LeagueCode, 0, 1<<30):
		return "events.leagueCode", "VALUE_OUT_OF_RANGE"
	}
	return "", ""
}

func validCron(s string) bool {
	_, err := cronexpr.Parse(s)
	return err == nil
}

var store = settingsrow.New("trophy_settings", Defaults)

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
