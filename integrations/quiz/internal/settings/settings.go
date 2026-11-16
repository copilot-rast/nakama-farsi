package settings

import (
	"context"
	"database/sql"
	"regexp"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

type Play struct {
	Delivery string `json:"delivery"`

	ShuffleQuestions bool `json:"shuffleQuestions"`
	ShuffleOptions   bool `json:"shuffleOptions"`

	OnRestart string `json:"onRestart"`

	SessionHours int `json:"sessionHours"`

	AutoFinish bool `json:"autoFinish"`
}

type Timing struct {
	QuestionSeconds int `json:"questionSeconds"`

	GraceSeconds int `json:"graceSeconds"`

	LateAnswer string `json:"lateAnswer"`
}

type Scoring struct {
	PassPercent float64 `json:"passPercent"`

	MultiCorrect string `json:"multiCorrect"`

	TextMatch string `json:"textMatch"`
}

type Reveal struct {
	Mode string `json:"mode"`

	AllowBeforeAnswer bool `json:"allowBeforeAnswer"`

	Explanation bool `json:"explanation"`
}

type Rewards struct {
	PackMode string `json:"packMode"`

	QuestionMode   string `json:"questionMode"`
	PrizeGroupMode string `json:"prizeGroupMode"`

	DayZone string `json:"dayZone"`
}

type Limits struct {
	DailyPlays int `json:"dailyPlays"`

	DailyRewardedPasses int `json:"dailyRewardedPasses"`

	MaxPlaysPerPack int `json:"maxPlaysPerPack"`

	ReplayAfterPass bool `json:"replayAfterPass"`

	RetryCooldownSeconds int `json:"retryCooldownSeconds"`
}

type Price struct {
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
}

type Hints struct {
	Pricing    string `json:"pricing"`
	FixedPrice Price  `json:"fixedPrice"`

	Keep string `json:"keep"`

	MaxPerQuestion int `json:"maxPerQuestion"`
}

type Reports struct {
	Enabled bool `json:"enabled"`

	DailyLimit int `json:"dailyLimit"`

	RequireSeen  bool `json:"requireSeen"`
	NoteRequired bool `json:"noteRequired"`

	AutoDisableAt int `json:"autoDisableAt"`
}

type Events struct {
	Completed string `json:"completed"`
	Passed    string `json:"passed"`
	Correct   string `json:"correct"`

	Analytics string `json:"analytics"`
	OnFailure bool   `json:"onFailure"`
}

type Campaign struct {
	Enabled bool `json:"enabled"`

	SendFailures bool `json:"sendFailures"`
}

type Retention struct {
	SessionDays int `json:"sessionDays"`

	DeliveryDays int `json:"deliveryDays"`

	WebhookAttempts int `json:"webhookAttempts"`
}

type DuelEvents struct {
	Played  string `json:"played"`
	Won     string `json:"won"`
	Correct string `json:"correct"`
}

type Duel struct {
	Enabled           bool       `json:"enabled"`
	Rounds            int        `json:"rounds"`
	QuestionsPerRound int        `json:"questionsPerRound"`
	CategoryChoices   int        `json:"categoryChoices"`
	QuestionSeconds   int        `json:"questionSeconds"`
	GraceSeconds      int        `json:"graceSeconds"`
	QuestionSource    string     `json:"questionSource"`
	AvoidRepeatDays   int        `json:"avoidRepeatDays"`
	ShuffleOptions    bool       `json:"shuffleOptions"`
	ServeNextOnAnswer bool       `json:"serveNextOnAnswer"`
	TurnHours         int        `json:"turnHours"`
	TimeoutResult     string     `json:"timeoutResult"`
	WaitingHours      int        `json:"waitingHours"`
	InviteHours       int        `json:"inviteHours"`
	FriendsOnly       bool       `json:"friendsOnly"`
	MaxActive         int        `json:"maxActive"`
	DailyStarts       int        `json:"dailyStarts"`
	EntryFee          Price      `json:"entryFee"`
	WinPrize          string     `json:"winPrize"`
	LosePrize         string     `json:"losePrize"`
	DrawPrize         string     `json:"drawPrize"`
	PerCorrect        Price      `json:"perCorrect"`
	Events            DuelEvents `json:"events"`
	Notify            bool       `json:"notify"`
	HistoryDays       int        `json:"historyDays"`
}

type All struct {
	Duel      Duel      `json:"duel"`
	Play      Play      `json:"play"`
	Timing    Timing    `json:"timing"`
	Scoring   Scoring   `json:"scoring"`
	Reveal    Reveal    `json:"reveal"`
	Rewards   Rewards   `json:"rewards"`
	Limits    Limits    `json:"limits"`
	Hints     Hints     `json:"hints"`
	Reports   Reports   `json:"reports"`
	Events    Events    `json:"events"`
	Campaign  Campaign  `json:"campaign"`
	Retention Retention `json:"retention"`
}

const (
	DeliveryAll      = "all"
	DeliveryOneByOne = "one_by_one"
	RestartResume    = "resume"
	RestartDiscard   = "discard"
	LateWrong        = "wrong"
	LateReject       = "reject"
	MultiAny         = "any"
	MultiAll         = "all"
	MatchNormalized  = "normalized"
	MatchExact       = "exact"
	RevealAnswer     = "after_answer"
	RevealFinish     = "after_finish"
	RevealNever      = "never"
	RewardFirst      = "first_pass"
	RewardEvery      = "every_pass"
	RewardNone       = "none"
	QuestionFirst    = "first_correct"
	QuestionEvery    = "every_correct"
	PricePerHint     = "per_hint"
	PriceFixed       = "fixed"
	PriceFree        = "free"
	KeepForever      = "forever"
	KeepSession      = "session"
	GroupDraw        = "draw"
	GroupAll         = "all"
	SourceAny        = "any"
	SourceBank       = "bank"
	TimeoutOpponent  = "opponent_wins"
	TimeoutScore     = "score"
)

var (
	KeyPattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,79}$`)
	EventPattern = regexp.MustCompile(`^[a-z][a-z0-9_.]{0,63}$`)
)

func Defaults() All {
	return All{
		Play:      Play{Delivery: DeliveryAll, ShuffleQuestions: true, ShuffleOptions: true, OnRestart: RestartResume, SessionHours: 24, AutoFinish: true},
		Timing:    Timing{QuestionSeconds: 0, GraceSeconds: 2, LateAnswer: LateWrong},
		Scoring:   Scoring{PassPercent: 100, MultiCorrect: MultiAny, TextMatch: MatchNormalized},
		Reveal:    Reveal{Mode: RevealAnswer, AllowBeforeAnswer: false, Explanation: true},
		Rewards:   Rewards{PackMode: RewardFirst, QuestionMode: QuestionFirst, PrizeGroupMode: GroupDraw, DayZone: "UTC"},
		Limits:    Limits{ReplayAfterPass: true},
		Hints:     Hints{Pricing: PricePerHint, Keep: KeepForever},
		Reports:   Reports{Enabled: true, DailyLimit: 20, RequireSeen: true},
		Events:    Events{Completed: "quiz_completed", Passed: "quiz_passed", Correct: "quiz_correct", Analytics: "quiz_completed", OnFailure: true},
		Campaign:  Campaign{Enabled: true, SendFailures: true},
		Retention: Retention{SessionDays: 180, DeliveryDays: 30, WebhookAttempts: 5},
		Duel: Duel{Enabled: true, Rounds: 6, QuestionsPerRound: 3, CategoryChoices: 3, QuestionSeconds: 20, GraceSeconds: 2, QuestionSource: SourceAny,
			AvoidRepeatDays: 14, ShuffleOptions: true, TurnHours: 24, TimeoutResult: TimeoutOpponent, WaitingHours: 48, InviteHours: 24, FriendsOnly: true,
			MaxActive: 10, Events: DuelEvents{Played: "quiz_duel_played", Won: "quiz_duel_won", Correct: "quiz_duel_correct"}, Notify: true, HistoryDays: 180},
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

func (a All) Validate() (string, string) {
	switch {
	case !oneOf(a.Play.Delivery, DeliveryAll, DeliveryOneByOne):
		return "play.delivery", "INVALID_POLICY"
	case !oneOf(a.Play.OnRestart, RestartResume, RestartDiscard):
		return "play.onRestart", "INVALID_POLICY"
	case !between(a.Play.SessionHours, 1, 720):
		return "play.sessionHours", "VALUE_OUT_OF_RANGE"
	case !between(a.Timing.QuestionSeconds, 0, 3600):
		return "timing.questionSeconds", "VALUE_OUT_OF_RANGE"
	case !between(a.Timing.GraceSeconds, 0, 60):
		return "timing.graceSeconds", "VALUE_OUT_OF_RANGE"
	case !oneOf(a.Timing.LateAnswer, LateWrong, LateReject):
		return "timing.lateAnswer", "INVALID_POLICY"
	case a.Scoring.PassPercent < 0 || a.Scoring.PassPercent > 100:
		return "scoring.passPercent", "VALUE_OUT_OF_RANGE"
	case !oneOf(a.Scoring.MultiCorrect, MultiAny, MultiAll):
		return "scoring.multiCorrect", "INVALID_POLICY"
	case !oneOf(a.Scoring.TextMatch, MatchNormalized, MatchExact):
		return "scoring.textMatch", "INVALID_POLICY"
	case !oneOf(a.Reveal.Mode, RevealAnswer, RevealFinish, RevealNever):
		return "reveal.mode", "INVALID_POLICY"
	case !oneOf(a.Rewards.PackMode, RewardFirst, RewardEvery, RewardNone):
		return "rewards.packMode", "INVALID_POLICY"
	case !oneOf(a.Rewards.QuestionMode, QuestionFirst, QuestionEvery, RewardNone):
		return "rewards.questionMode", "INVALID_POLICY"
	case !oneOf(a.Rewards.PrizeGroupMode, GroupDraw, GroupAll):
		return "rewards.prizeGroupMode", "INVALID_POLICY"
	case !between(a.Limits.DailyPlays, 0, 100000):
		return "limits.dailyPlays", "VALUE_OUT_OF_RANGE"
	case !between(a.Limits.DailyRewardedPasses, 0, 100000):
		return "limits.dailyRewardedPasses", "VALUE_OUT_OF_RANGE"
	case !between(a.Limits.MaxPlaysPerPack, 0, 100000):
		return "limits.maxPlaysPerPack", "VALUE_OUT_OF_RANGE"
	case !between(a.Limits.RetryCooldownSeconds, 0, 30*86400):
		return "limits.retryCooldownSeconds", "VALUE_OUT_OF_RANGE"
	case !oneOf(a.Hints.Pricing, PricePerHint, PriceFixed, PriceFree):
		return "hints.pricing", "INVALID_POLICY"
	case a.Hints.Pricing == PriceFixed && a.Hints.FixedPrice.Amount > 0 && !KeyPattern.MatchString(a.Hints.FixedPrice.CurrencyKey):
		return "hints.fixedPrice.currencyKey", "INVALID_KEY"
	case a.Hints.FixedPrice.Amount < 0 || a.Hints.FixedPrice.Amount > 1e12:
		return "hints.fixedPrice.amount", "VALUE_OUT_OF_RANGE"
	case !oneOf(a.Hints.Keep, KeepForever, KeepSession):
		return "hints.keep", "INVALID_POLICY"
	case !between(a.Hints.MaxPerQuestion, 0, 50):
		return "hints.maxPerQuestion", "VALUE_OUT_OF_RANGE"
	case !between(a.Reports.DailyLimit, 0, 10000):
		return "reports.dailyLimit", "VALUE_OUT_OF_RANGE"
	case !between(a.Reports.AutoDisableAt, 0, 10000):
		return "reports.autoDisableAt", "VALUE_OUT_OF_RANGE"
	case a.Events.Completed != "" && !EventPattern.MatchString(a.Events.Completed):
		return "events.completed", "INVALID_KEY"
	case a.Events.Passed != "" && !EventPattern.MatchString(a.Events.Passed):
		return "events.passed", "INVALID_KEY"
	case a.Events.Correct != "" && !EventPattern.MatchString(a.Events.Correct):
		return "events.correct", "INVALID_KEY"
	case a.Events.Analytics != "" && !EventPattern.MatchString(a.Events.Analytics):
		return "events.analytics", "INVALID_KEY"
	case !between(a.Retention.SessionDays, 0, 3650):
		return "retention.sessionDays", "VALUE_OUT_OF_RANGE"
	case !between(a.Retention.DeliveryDays, 0, 3650):
		return "retention.deliveryDays", "VALUE_OUT_OF_RANGE"
	case !between(a.Retention.WebhookAttempts, 1, 20):
		return "retention.webhookAttempts", "VALUE_OUT_OF_RANGE"
	}
	if f, c := a.Duel.validate(); f != "" {
		return f, c
	}
	if _, err := time.LoadLocation(a.Rewards.DayZone); err != nil || a.Rewards.DayZone == "" {
		return "rewards.dayZone", "INVALID_TIMEZONE"
	}
	return "", ""
}

var store = settingsrow.New("quiz_settings", Defaults)

func Load(ctx context.Context, db *sql.DB) (All, error) { return store.Load(ctx, db) }

func Invalidate() { store.Invalidate() }

func Read(ctx context.Context, db *sql.DB) (All, settingsrow.Meta, error) { return store.Read(ctx, db) }

func Save(ctx context.Context, db *sql.DB, in All, by string) error {
	return store.Save(ctx, db, in, by)
}

func validPrice(p Price) bool {
	return p.Amount == 0 || (p.Amount > 0 && p.Amount <= 1e12 && KeyPattern.MatchString(p.CurrencyKey))
}

func (d Duel) validate() (string, string) {
	switch {
	case !between(d.Rounds, 1, 12):
		return "duel.rounds", "VALUE_OUT_OF_RANGE"
	case !between(d.QuestionsPerRound, 1, 10):
		return "duel.questionsPerRound", "VALUE_OUT_OF_RANGE"
	case !between(d.CategoryChoices, 1, 8):
		return "duel.categoryChoices", "VALUE_OUT_OF_RANGE"
	case !between(d.QuestionSeconds, 5, 300):
		return "duel.questionSeconds", "VALUE_OUT_OF_RANGE"
	case !between(d.GraceSeconds, 0, 30):
		return "duel.graceSeconds", "VALUE_OUT_OF_RANGE"
	case !oneOf(d.QuestionSource, SourceAny, SourceBank):
		return "duel.questionSource", "INVALID_POLICY"
	case !between(d.AvoidRepeatDays, 0, 365):
		return "duel.avoidRepeatDays", "VALUE_OUT_OF_RANGE"
	case !between(d.TurnHours, 1, 720):
		return "duel.turnHours", "VALUE_OUT_OF_RANGE"
	case !oneOf(d.TimeoutResult, TimeoutOpponent, TimeoutScore):
		return "duel.timeoutResult", "INVALID_POLICY"
	case !between(d.WaitingHours, 1, 720):
		return "duel.waitingHours", "VALUE_OUT_OF_RANGE"
	case !between(d.InviteHours, 1, 720):
		return "duel.inviteHours", "VALUE_OUT_OF_RANGE"
	case !between(d.MaxActive, 1, 100):
		return "duel.maxActive", "VALUE_OUT_OF_RANGE"
	case !between(d.DailyStarts, 0, 10000):
		return "duel.dailyStarts", "VALUE_OUT_OF_RANGE"
	case !validPrice(d.EntryFee):
		return "duel.entryFee", "VALUE_OUT_OF_RANGE"
	case !validPrice(d.PerCorrect):
		return "duel.perCorrect", "VALUE_OUT_OF_RANGE"
	case d.WinPrize != "" && !KeyPattern.MatchString(d.WinPrize):
		return "duel.winPrize", "INVALID_KEY"
	case d.LosePrize != "" && !KeyPattern.MatchString(d.LosePrize):
		return "duel.losePrize", "INVALID_KEY"
	case d.DrawPrize != "" && !KeyPattern.MatchString(d.DrawPrize):
		return "duel.drawPrize", "INVALID_KEY"
	case d.Events.Played != "" && !EventPattern.MatchString(d.Events.Played):
		return "duel.events.played", "INVALID_KEY"
	case d.Events.Won != "" && !EventPattern.MatchString(d.Events.Won):
		return "duel.events.won", "INVALID_KEY"
	case d.Events.Correct != "" && !EventPattern.MatchString(d.Events.Correct):
		return "duel.events.correct", "INVALID_KEY"
	case !between(d.HistoryDays, 0, 3650):
		return "duel.historyDays", "VALUE_OUT_OF_RANGE"
	}
	return "", ""
}
