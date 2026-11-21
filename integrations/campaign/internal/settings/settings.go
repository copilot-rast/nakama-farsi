package settings

import (
	"context"
	"database/sql"
	"regexp"
	"strconv"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

type Cond struct {
	Field   string  `json:"field"`
	Op      string  `json:"op"`
	Value   float64 `json:"value"`
	Missing string  `json:"missing,omitempty"`
}

type StarRule struct {
	Stars      int    `json:"stars"`
	Conditions []Cond `json:"conditions"`
}

type Attempts struct {
	Hours              int     `json:"hours"`
	GraceSeconds       int     `json:"graceSeconds"`
	DefaultMinSeconds  float64 `json:"defaultMinSeconds"`
	CloseOthers        bool    `json:"closeOthers"`
	ServerClosesOthers bool    `json:"serverClosesOthers"`
	KeepDays           int     `json:"keepDays"`
}

type Template struct {
	Key   string            `json:"key"`
	Name  map[string]string `json:"name"`
	Rules []StarRule        `json:"rules"`
}

type Stars struct {
	Rules                 []StarRule `json:"rules"`
	Templates             []Template `json:"templates"`
	Multipliers           []float64  `json:"multipliers"`
	RepeatableMultipliers []float64  `json:"repeatableMultipliers"`
	MaxStars              int        `json:"maxStars"`
	MinOnSuccess          int        `json:"minOnSuccess"`
	MissingField          string     `json:"missingField"`
}

type Unlock struct {
	StageDefault   string `json:"stageDefault"`
	ChapterDefault string `json:"chapterDefault"`
	MinStars       int    `json:"minStars"`
	ShowLocked     bool   `json:"showLocked"`
}

type Rewards struct {
	ReplayMode     string   `json:"replayMode"`
	DayZone        string   `json:"dayZone"`
	PrizeGroupMode string   `json:"prizeGroupMode"`
	MilestoneModes []string `json:"milestoneModes"`
}

type XP struct {
	MetaKey       string `json:"metaKey"`
	LevelTableKey string `json:"levelTableKey"`
}

type Events struct {
	EconomyEvent   string `json:"economyEvent"`
	AnalyticsEvent string `json:"analyticsEvent"`
	OnFailure      bool   `json:"onFailure"`
}

type Freeride struct {
	TableEnabled     bool `json:"tableEnabled"`
	RequireRequestID bool `json:"requireRequestId"`
	RequestHours     int  `json:"requestHours"`
}

type FieldDef struct {
	Key      string            `json:"key"`
	Label    map[string]string `json:"label"`
	Hint     map[string]string `json:"hint,omitempty"`
	Type     string            `json:"type"`
	Default  any               `json:"default,omitempty"`
	Min      *float64          `json:"min,omitempty"`
	Max      *float64          `json:"max,omitempty"`
	Options  []string          `json:"options,omitempty"`
	Fields   []FieldDef        `json:"fields,omitempty"`
	MaxItems int               `json:"maxItems,omitempty"`
	Chapters []string          `json:"chapters,omitempty"`
}

type Fields struct {
	Stage   []FieldDef `json:"stage"`
	Chapter []FieldDef `json:"chapter"`
}

type Play struct {
	CostCurrency string `json:"costCurrency"`
	CostAmount   int64  `json:"costAmount"`
	RefundOnWin  bool   `json:"refundOnWin"`
}

type All struct {
	Play     Play     `json:"play"`
	Attempts Attempts `json:"attempts"`
	Stars    Stars    `json:"stars"`
	Unlock   Unlock   `json:"unlock"`
	Rewards  Rewards  `json:"rewards"`
	XP       XP       `json:"xp"`
	Events   Events   `json:"events"`
	Freeride Freeride `json:"freeride"`
	Fields   Fields   `json:"fields"`
}

const (
	ModeOrdered    = "ordered"
	ModeRepeatable = "repeatable"
	ModeOnce       = "once"

	UnlockAllPrevious = "all_previous"
	UnlockPrevious    = "previous"
	UnlockNone        = "none"

	ReplayDifference = "difference"
	ReplayNone       = "none"
	ReplayFull       = "full"

	MissingPass = "pass"
	MissingFail = "fail"

	GroupDraw = "draw"
	GroupAll  = "all"

	FieldNumber  = "number"
	FieldInteger = "integer"
	FieldBool    = "bool"
	FieldText    = "text"
	FieldI18n    = "i18n"
	FieldSelect  = "select"
	FieldTags    = "tags"
	FieldList    = "list"

	MaxFieldDepth = 3
)

var FieldTypes = map[string]bool{FieldNumber: true, FieldInteger: true, FieldBool: true, FieldText: true, FieldI18n: true, FieldSelect: true, FieldTags: true, FieldList: true}

func baseRules(time, health bool) []StarRule {
	three := []Cond{}
	two := []Cond{}
	if time {
		three = append(three, Cond{Field: "timeSavedPercent", Op: ">=", Value: 30})
		two = append(two, Cond{Field: "timeSavedPercent", Op: ">=", Value: 10})
	}
	if health {
		three = append(three, Cond{Field: "healthPercent", Op: ">=", Value: 75})
		two = append(two, Cond{Field: "healthPercent", Op: ">=", Value: 50})
	}
	three = append(three, Cond{Field: "timesSeen", Op: "<=", Value: 0, Missing: MissingPass})
	two = append(two, Cond{Field: "timesSeen", Op: "<=", Value: 2, Missing: MissingPass})
	return []StarRule{{Stars: 3, Conditions: three}, {Stars: 2, Conditions: two}}
}

func DefaultTemplates() []Template {
	return []Template{
		{Key: "base", Name: map[string]string{"fa": "پایه", "en": "Base"}, Rules: baseRules(true, true)},
		{Key: "race", Name: map[string]string{"fa": "مسابقه", "en": "Race"}, Rules: baseRules(true, false)},
		{Key: "stealth", Name: map[string]string{"fa": "مخفی", "en": "Stealth"}, Rules: baseRules(false, true)},
		{Key: "quest", Name: map[string]string{"fa": "کوئست", "en": "Quest"}, Rules: baseRules(true, true)},
	}
}

func (s Stars) Template(key string) *Template {
	for i := range s.Templates {
		if s.Templates[i].Key == key {
			return &s.Templates[i]
		}
	}
	return nil
}

var (
	KeyPattern      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,79}$`)
	FieldPattern    = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,39}$`)
	EventPattern    = regexp.MustCompile(`^[a-z][a-z0-9_.]{0,63}$`)
	CurrencyPattern = regexp.MustCompile(`^[a-z0-9_]{2,32}$`)
	Ops             = map[string]bool{">=": true, ">": true, "<=": true, "<": true, "==": true, "!=": true}
)

func Defaults() All {
	return All{
		Play:     Play{CostAmount: 1, RefundOnWin: true},
		Attempts: Attempts{Hours: 24, GraceSeconds: 3, DefaultMinSeconds: 5, CloseOthers: true, KeepDays: 90},
		Stars: Stars{
			Rules:                 baseRules(true, true),
			Templates:             DefaultTemplates(),
			Multipliers:           []float64{0.2, 0.5, 1},
			RepeatableMultipliers: []float64{1, 1, 1},
			MaxStars:              3,
			MinOnSuccess:          1,
			MissingField:          MissingFail,
		},
		Unlock:   Unlock{StageDefault: UnlockAllPrevious, ChapterDefault: UnlockNone, MinStars: 1, ShowLocked: true},
		Rewards:  Rewards{ReplayMode: ReplayDifference, DayZone: "UTC", PrizeGroupMode: GroupDraw, MilestoneModes: []string{ModeOrdered}},
		XP:       XP{MetaKey: "xp"},
		Events:   Events{EconomyEvent: "stage_completed", AnalyticsEvent: "stage_completed", OnFailure: true},
		Freeride: Freeride{RequireRequestID: true, RequestHours: 24},
		Fields:   Fields{Stage: []FieldDef{}, Chapter: []FieldDef{}},
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

func ValidateRules(rules []StarRule, maxStars int) (string, string) {
	if len(rules) > 10 {
		return "", "TOO_MANY_RULES"
	}
	for i, r := range rules {
		if r.Stars < 1 || r.Stars > maxStars {
			return "[" + strconv.Itoa(i) + "].stars", "VALUE_OUT_OF_RANGE"
		}
		if len(r.Conditions) > 10 {
			return "[" + strconv.Itoa(i) + "].conditions", "TOO_MANY_CONDITIONS"
		}
		for j, c := range r.Conditions {
			if !FieldPattern.MatchString(c.Field) {
				return "[" + strconv.Itoa(i) + "].conditions[" + strconv.Itoa(j) + "].field", "INVALID_FIELD"
			}
			if !Ops[c.Op] {
				return "[" + strconv.Itoa(i) + "].conditions[" + strconv.Itoa(j) + "].op", "INVALID_OPERATOR"
			}
			if c.Missing != "" && !oneOf(c.Missing, MissingPass, MissingFail) {
				return "[" + strconv.Itoa(i) + "].conditions[" + strconv.Itoa(j) + "].missing", "INVALID_POLICY"
			}
		}
	}
	return "", ""
}

func ValidateMultipliers(m []float64, maxStars int) bool {
	if len(m) != maxStars {
		return false
	}
	for _, v := range m {
		if v < 0 || v > 10 {
			return false
		}
	}
	return true
}

func (a All) Validate() (string, string) {
	switch {
	case a.Play.CostCurrency != "" && !CurrencyPattern.MatchString(a.Play.CostCurrency):
		return "play.costCurrency", "INVALID_KEY"
	case a.Play.CostAmount < 1 || a.Play.CostAmount > 1000000:
		return "play.costAmount", "VALUE_OUT_OF_RANGE"
	case a.Attempts.Hours < 1 || a.Attempts.Hours > 168:
		return "attempts.hours", "VALUE_OUT_OF_RANGE"
	case a.Attempts.DefaultMinSeconds < 0 || a.Attempts.DefaultMinSeconds > 3600:
		return "attempts.defaultMinSeconds", "VALUE_OUT_OF_RANGE"
	case a.Attempts.GraceSeconds < 0 || a.Attempts.GraceSeconds > 120:
		return "attempts.graceSeconds", "VALUE_OUT_OF_RANGE"
	case a.Attempts.KeepDays < 0 || a.Attempts.KeepDays > 3650:
		return "attempts.keepDays", "VALUE_OUT_OF_RANGE"
	case a.Stars.MaxStars < 1 || a.Stars.MaxStars > 5:
		return "stars.maxStars", "VALUE_OUT_OF_RANGE"
	case a.Stars.MinOnSuccess < 0 || a.Stars.MinOnSuccess > a.Stars.MaxStars:
		return "stars.minOnSuccess", "VALUE_OUT_OF_RANGE"
	case !oneOf(a.Stars.MissingField, MissingPass, MissingFail):
		return "stars.missingField", "INVALID_POLICY"
	case !ValidateMultipliers(a.Stars.Multipliers, a.Stars.MaxStars):
		return "stars.multipliers", "INVALID_MULTIPLIERS"
	case !ValidateMultipliers(a.Stars.RepeatableMultipliers, a.Stars.MaxStars):
		return "stars.repeatableMultipliers", "INVALID_MULTIPLIERS"
	case !oneOf(a.Unlock.StageDefault, UnlockAllPrevious, UnlockPrevious, UnlockNone):
		return "unlock.stageDefault", "INVALID_POLICY"
	case !oneOf(a.Unlock.ChapterDefault, UnlockNone, UnlockPrevious):
		return "unlock.chapterDefault", "INVALID_POLICY"
	case a.Unlock.MinStars < 1 || a.Unlock.MinStars > a.Stars.MaxStars:
		return "unlock.minStars", "VALUE_OUT_OF_RANGE"
	case !oneOf(a.Rewards.ReplayMode, ReplayDifference, ReplayNone, ReplayFull):
		return "rewards.replayMode", "INVALID_POLICY"
	case !oneOf(a.Rewards.PrizeGroupMode, GroupDraw, GroupAll):
		return "rewards.prizeGroupMode", "INVALID_POLICY"
	case a.XP.MetaKey != "" && !KeyPattern.MatchString(a.XP.MetaKey):
		return "xp.metaKey", "INVALID_KEY"
	case a.XP.LevelTableKey != "" && !KeyPattern.MatchString(a.XP.LevelTableKey):
		return "xp.levelTableKey", "INVALID_KEY"
	case a.Events.EconomyEvent != "" && !EventPattern.MatchString(a.Events.EconomyEvent):
		return "events.economyEvent", "INVALID_KEY"
	case a.Events.AnalyticsEvent != "" && !EventPattern.MatchString(a.Events.AnalyticsEvent):
		return "events.analyticsEvent", "INVALID_KEY"
	case a.Freeride.RequestHours < 1 || a.Freeride.RequestHours > 720:
		return "freeride.requestHours", "VALUE_OUT_OF_RANGE"
	}
	if _, err := time.LoadLocation(a.Rewards.DayZone); err != nil || a.Rewards.DayZone == "" {
		return "rewards.dayZone", "INVALID_TIMEZONE"
	}
	for _, m := range a.Rewards.MilestoneModes {
		if !oneOf(m, ModeOrdered, ModeRepeatable, ModeOnce) {
			return "rewards.milestoneModes", "INVALID_POLICY"
		}
	}
	if f, code := ValidateRules(a.Stars.Rules, a.Stars.MaxStars); code != "" {
		return "stars.rules" + f, code
	}
	if len(a.Stars.Templates) > 20 {
		return "stars.templates", "TOO_MANY_TEMPLATES"
	}
	seen := map[string]bool{}
	for i, t := range a.Stars.Templates {
		f := "stars.templates[" + strconv.Itoa(i) + "]"
		if !KeyPattern.MatchString(t.Key) {
			return f + ".key", "INVALID_KEY"
		}
		if seen[t.Key] {
			return f + ".key", "DUPLICATE_KEY"
		}
		seen[t.Key] = true
		if sub, code := ValidateRules(t.Rules, a.Stars.MaxStars); code != "" {
			return f + ".rules" + sub, code
		}
	}
	if f, code := ValidateFields(a.Fields.Stage, 1); code != "" {
		return "fields.stage" + f, code
	}
	if f, code := ValidateFields(a.Fields.Chapter, 1); code != "" {
		return "fields.chapter" + f, code
	}
	return "", ""
}

func ValidateFields(defs []FieldDef, depth int) (string, string) {
	if depth > MaxFieldDepth {
		return "", "FIELDS_TOO_DEEP"
	}
	if len(defs) > 60 {
		return "", "TOO_MANY_FIELDS"
	}
	seen := map[string]bool{}
	for i, d := range defs {
		f := "[" + strconv.Itoa(i) + "]"
		if !FieldPattern.MatchString(d.Key) {
			return f + ".key", "INVALID_FIELD"
		}
		if seen[d.Key] {
			return f + ".key", "DUPLICATE_KEY"
		}
		seen[d.Key] = true
		if !FieldTypes[d.Type] {
			return f + ".type", "INVALID_FIELD_TYPE"
		}
		if d.Type == FieldSelect && len(d.Options) == 0 {
			return f + ".options", "OPTIONS_REQUIRED"
		}
		if len(d.Options) > 50 {
			return f + ".options", "TOO_MANY_OPTIONS"
		}
		if d.Min != nil && d.Max != nil && *d.Min > *d.Max {
			return f + ".min", "VALUE_OUT_OF_RANGE"
		}
		if d.MaxItems < 0 || d.MaxItems > 100 {
			return f + ".maxItems", "VALUE_OUT_OF_RANGE"
		}
		for _, c := range d.Chapters {
			if !KeyPattern.MatchString(c) {
				return f + ".chapters", "INVALID_KEY"
			}
		}
		if d.Type == FieldList {
			if len(d.Fields) == 0 {
				return f + ".fields", "FIELDS_REQUIRED"
			}
			if sub, code := ValidateFields(d.Fields, depth+1); code != "" {
				return f + ".fields" + sub, code
			}
		} else if len(d.Fields) > 0 {
			return f + ".fields", "INVALID_FIELD_TYPE"
		}
		if d.Default != nil {
			if _, code := CleanValue(d, d.Default, depth); code != "" {
				return f + ".default", code
			}
		}
	}
	return "", ""
}

var store = settingsrow.New("campaign_settings", Defaults)

func Load(ctx context.Context, db *sql.DB) (All, error) { return store.Load(ctx, db) }

func Invalidate() { store.Invalidate() }

func Read(ctx context.Context, db *sql.DB) (All, settingsrow.Meta, error) { return store.Read(ctx, db) }

func Save(ctx context.Context, db *sql.DB, in All, by string) error {
	return store.Save(ctx, db, in, by)
}
