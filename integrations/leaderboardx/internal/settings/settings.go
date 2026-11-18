package settings

import (
	"context"
	"database/sql"
	"regexp"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

type Display struct {
	TopLimit int `json:"topLimit"`

	MaxLimit int `json:"maxLimit"`

	AroundLimit int `json:"aroundLimit"`

	ShowDisplayName bool `json:"showDisplayName"`

	ShowUpcoming bool `json:"showUpcoming"`

	ShowEndedDays int `json:"showEndedDays"`
}

type Rewards struct {
	PayMode string `json:"payMode"`

	ClaimDays int `json:"claimDays"`

	Stack string `json:"stack"`

	PrizeGroupMode string `json:"prizeGroupMode"`

	MaxTries int `json:"maxTries"`

	CatchUpDays int `json:"catchUpDays"`
}

type Scores struct {
	RequireRequestID bool `json:"requireRequestId"`

	PlayerPerMinute int `json:"playerPerMinute"`

	ClientMaxPerWrite int64 `json:"clientMaxPerWrite"`

	LogDays int `json:"logDays"`
}

type History struct {
	KeepDays int `json:"keepDays"`

	PlayerLimit int `json:"playerLimit"`

	WinnersLimit int `json:"winnersLimit"`
}

type Events struct {
	EconomyEvent string `json:"economyEvent"`

	AnalyticsEvent string `json:"analyticsEvent"`

	Notify bool `json:"notify"`

	NotifyCode int `json:"notifyCode"`
}

type All struct {
	Display Display `json:"display"`
	Rewards Rewards `json:"rewards"`
	Scores  Scores  `json:"scores"`
	History History `json:"history"`
	Events  Events  `json:"events"`
}

const (
	PayAuto  = "auto"
	PayClaim = "claim"

	StackFirst = "first"
	StackAll   = "all"

	GroupDraw = "draw"
	GroupAll  = "all"

	DefaultClientMaxPerWrite = 1000000
	DefaultNotifyCode        = 2301
)

var (
	KeyPattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,79}$`)
	EventPattern = regexp.MustCompile(`^[a-z][a-z0-9_.]{0,63}$`)
)

func Defaults() All {
	return All{
		Display: Display{TopLimit: 50, MaxLimit: 200, AroundLimit: 5, ShowDisplayName: true, ShowUpcoming: true, ShowEndedDays: 3},
		Rewards: Rewards{PayMode: PayAuto, ClaimDays: 7, Stack: StackFirst, PrizeGroupMode: GroupDraw, MaxTries: 8, CatchUpDays: 7},
		Scores:  Scores{RequireRequestID: true, PlayerPerMinute: 60, ClientMaxPerWrite: DefaultClientMaxPerWrite, LogDays: 30},
		History: History{KeepDays: 365, PlayerLimit: 20, WinnersLimit: 100},
		Events:  Events{EconomyEvent: "leaderboard_ranked", AnalyticsEvent: "leaderboard_ranked", Notify: true, NotifyCode: DefaultNotifyCode},
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

func (e Events) Code() int {
	if e.NotifyCode == 0 {
		return DefaultNotifyCode
	}
	return e.NotifyCode
}

func (a All) Validate() (string, string) {
	switch {
	case !between(a.Display.TopLimit, 1, 200):
		return "display.topLimit", "VALUE_OUT_OF_RANGE"
	case !between(a.Display.MaxLimit, 1, 500) || a.Display.MaxLimit < a.Display.TopLimit:
		return "display.maxLimit", "VALUE_OUT_OF_RANGE"
	case !between(a.Display.AroundLimit, 0, 50):
		return "display.aroundLimit", "VALUE_OUT_OF_RANGE"
	case !between(a.Display.ShowEndedDays, 0, 365):
		return "display.showEndedDays", "VALUE_OUT_OF_RANGE"
	case !oneOf(a.Rewards.PayMode, PayAuto, PayClaim):
		return "rewards.payMode", "INVALID_POLICY"
	case !between(a.Rewards.ClaimDays, 1, 365):
		return "rewards.claimDays", "VALUE_OUT_OF_RANGE"
	case !oneOf(a.Rewards.Stack, StackFirst, StackAll):
		return "rewards.stack", "INVALID_POLICY"
	case !oneOf(a.Rewards.PrizeGroupMode, GroupDraw, GroupAll):
		return "rewards.prizeGroupMode", "INVALID_POLICY"
	case !between(a.Rewards.MaxTries, 1, 100):
		return "rewards.maxTries", "VALUE_OUT_OF_RANGE"
	case !between(a.Rewards.CatchUpDays, 0, 365):
		return "rewards.catchUpDays", "VALUE_OUT_OF_RANGE"
	case !between(a.Scores.PlayerPerMinute, 0, 10000):
		return "scores.playerPerMinute", "VALUE_OUT_OF_RANGE"
	case a.Scores.ClientMaxPerWrite < 0 || a.Scores.ClientMaxPerWrite > 1e15:
		return "scores.clientMaxPerWrite", "VALUE_OUT_OF_RANGE"
	case !between(a.Scores.LogDays, 1, 365):
		return "scores.logDays", "VALUE_OUT_OF_RANGE"
	case !between(a.History.KeepDays, 0, 3650):
		return "history.keepDays", "VALUE_OUT_OF_RANGE"
	case !between(a.History.PlayerLimit, 1, 200):
		return "history.playerLimit", "VALUE_OUT_OF_RANGE"
	case !between(a.History.WinnersLimit, 1, 500):
		return "history.winnersLimit", "VALUE_OUT_OF_RANGE"
	case a.Events.EconomyEvent != "" && !EventPattern.MatchString(a.Events.EconomyEvent):
		return "events.economyEvent", "INVALID_KEY"
	case a.Events.AnalyticsEvent != "" && !EventPattern.MatchString(a.Events.AnalyticsEvent):
		return "events.analyticsEvent", "INVALID_KEY"
	case !between(a.Events.NotifyCode, 0, 1<<30):
		return "events.notifyCode", "VALUE_OUT_OF_RANGE"
	}
	return "", ""
}

var store = settingsrow.New("leaderboardx_settings", Defaults)

func Load(ctx context.Context, db *sql.DB) (All, error) { return store.Load(ctx, db) }

func Invalidate() { store.Invalidate() }

func Read(ctx context.Context, db *sql.DB) (All, settingsrow.Meta, error) { return store.Read(ctx, db) }

func Save(ctx context.Context, db *sql.DB, in All, by string) error {
	return store.Save(ctx, db, in, by)
}
