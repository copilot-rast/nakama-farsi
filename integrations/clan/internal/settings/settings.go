package settings

import (
	"context"
	"database/sql"
	"regexp"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

type Create struct {
	Enabled bool `json:"enabled"`

	Cost map[string]int64 `json:"cost"`

	MinTrophies int `json:"minTrophies"`

	NameMin int `json:"nameMin"`

	NameMax int `json:"nameMax"`

	DescriptionMax int `json:"descriptionMax"`

	DefaultJoinType string `json:"defaultJoinType"`

	MaxMinTrophies int `json:"maxMinTrophies"`
}

type Badge struct {
	Emblems []string `json:"emblems"`

	Colors []string `json:"colors"`

	Frames []string `json:"frames"`
}

type Roles struct {
	Accept []string `json:"accept"`

	Invite []string `json:"invite"`

	Kick []string `json:"kick"`

	Promote []string `json:"promote"`

	Edit []string `json:"edit"`

	War []string `json:"war"`

	MaxCoLeaders int `json:"maxCoLeaders"`

	LeaderLeave string `json:"leaderLeave"`
}

type Cooldowns struct {
	JoinAfterLeaveMinutes int `json:"joinAfterLeaveMinutes"`

	RejoinAfterKickHours int `json:"rejoinAfterKickHours"`

	LeaveAfterJoinMinutes int `json:"leaveAfterJoinMinutes"`

	InviteDays int `json:"inviteDays"`

	RequestsPerUser int `json:"requestsPerUser"`
}

type XPSource struct {
	Type string `json:"type"`

	Event string `json:"event"`

	Context string `json:"context,omitempty"`

	Field string `json:"field,omitempty"`

	Equals string `json:"equals,omitempty"`

	Amount int64 `json:"amount"`

	PerValue bool `json:"perValue,omitempty"`
}

type XP struct {
	Sources []XPSource `json:"sources"`

	MemberDailyCap int64 `json:"memberDailyCap"`

	WarWin int64 `json:"warWin"`

	WarDraw int64 `json:"warDraw"`

	WarLoss int64 `json:"warLoss"`
}

type Chat struct {
	SystemMessages bool `json:"systemMessages"`
}

type Display struct {
	SearchLimit int `json:"searchLimit"`

	HistoryLimit int `json:"historyLimit"`
}

type Events struct {
	Publish bool `json:"publish"`

	Notify bool `json:"notify"`

	Code int `json:"code"`
}

type All struct {
	Enabled   bool      `json:"enabled"`
	Create    Create    `json:"create"`
	Badge     Badge     `json:"badge"`
	Roles     Roles     `json:"roles"`
	Cooldowns Cooldowns `json:"cooldowns"`
	XP        XP        `json:"xp"`
	Chat      Chat      `json:"chat"`
	Display   Display   `json:"display"`
	Events    Events    `json:"events"`
}

const (
	RoleLeader   = "leader"
	RoleCoLeader = "co_leader"
	RoleElder    = "elder"
	RoleMember   = "member"

	JoinOpen   = "open"
	JoinClosed = "closed"
	JoinInvite = "invite"

	LeaveTransfer = "transfer"
	LeaveBlock    = "block"
)

var (
	Roles4       = []string{RoleLeader, RoleCoLeader, RoleElder, RoleMember}
	KeyPattern   = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{0,63}$`)
	EventPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.:]{0,63}$`)
	FieldPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,39}$`)
	ShortPattern = regexp.MustCompile(`^[A-Za-z0-9_#-]{1,32}$`)
)

func Defaults() All {
	managers := []string{RoleLeader, RoleCoLeader}
	return All{
		Enabled: true,
		Create:  Create{Enabled: true, Cost: map[string]int64{}, NameMin: 3, NameMax: 24, DescriptionMax: 250, DefaultJoinType: JoinOpen, MaxMinTrophies: 100000},
		Badge:   Badge{Emblems: []string{}, Colors: []string{}, Frames: []string{}},
		Roles: Roles{Accept: []string{RoleLeader, RoleCoLeader, RoleElder}, Invite: []string{RoleLeader, RoleCoLeader, RoleElder}, Kick: managers, Promote: managers,
			Edit: managers, War: managers, MaxCoLeaders: 3, LeaderLeave: LeaveTransfer},
		Cooldowns: Cooldowns{JoinAfterLeaveMinutes: 60, RejoinAfterKickHours: 24, LeaveAfterJoinMinutes: 0, InviteDays: 7, RequestsPerUser: 5},
		XP:        XP{Sources: []XPSource{{Type: "game_event", Event: "task_completed", Amount: 5}}, MemberDailyCap: 0, WarWin: 100, WarDraw: 40, WarLoss: 20},
		Chat:      Chat{SystemMessages: true},
		Display:   Display{SearchLimit: 30, HistoryLimit: 20},
		Events:    Events{Publish: true, Notify: true, Code: 2410},
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

func (a *All) Normalize() {
	if a.Create.Cost == nil {
		a.Create.Cost = map[string]int64{}
	}
	for _, p := range []*[]string{&a.Badge.Emblems, &a.Badge.Colors, &a.Badge.Frames, &a.Roles.Accept, &a.Roles.Invite, &a.Roles.Kick, &a.Roles.Promote, &a.Roles.Edit, &a.Roles.War} {
		if *p == nil {
			*p = []string{}
		}
	}
	if a.XP.Sources == nil {
		a.XP.Sources = []XPSource{}
	}
}

func rolesOK(list []string) bool {
	for _, r := range list {
		if !oneOf(r, Roles4...) {
			return false
		}
	}
	return true
}

func (a All) Validate() (string, string) {
	for k, v := range a.Create.Cost {
		if !KeyPattern.MatchString(k) || v < 0 || v > 1e12 {
			return "create.cost", "VALUE_OUT_OF_RANGE"
		}
	}
	for _, list := range [][]string{a.Badge.Emblems, a.Badge.Colors, a.Badge.Frames} {
		if len(list) > 500 {
			return "badge", "VALUE_OUT_OF_RANGE"
		}
		for _, v := range list {
			if !ShortPattern.MatchString(v) {
				return "badge", "INVALID_KEY"
			}
		}
	}
	for name, list := range map[string][]string{"roles.accept": a.Roles.Accept, "roles.invite": a.Roles.Invite, "roles.kick": a.Roles.Kick, "roles.promote": a.Roles.Promote, "roles.edit": a.Roles.Edit, "roles.war": a.Roles.War} {
		if !rolesOK(list) {
			return name, "INVALID_POLICY"
		}
	}
	if len(a.XP.Sources) > 50 {
		return "xp.sources", "VALUE_OUT_OF_RANGE"
	}
	for _, src := range a.XP.Sources {
		switch {
		case !oneOf(src.Type, "economy_event", "game_event"):
			return "xp.sources", "INVALID_POLICY"
		case !EventPattern.MatchString(src.Event):
			return "xp.sources", "INVALID_KEY"
		case src.Field != "" && !FieldPattern.MatchString(src.Field):
			return "xp.sources", "INVALID_KEY"
		case src.Amount < 0 || src.Amount > 1000000:
			return "xp.sources", "VALUE_OUT_OF_RANGE"
		}
	}
	switch {
	case !between(a.Create.MinTrophies, 0, 10000000):
		return "create.minTrophies", "VALUE_OUT_OF_RANGE"
	case !between(a.Create.NameMin, 1, 64) || a.Create.NameMax < a.Create.NameMin:
		return "create.nameMin", "VALUE_OUT_OF_RANGE"
	case !between(a.Create.NameMax, 1, 64):
		return "create.nameMax", "VALUE_OUT_OF_RANGE"
	case !between(a.Create.DescriptionMax, 0, 2000):
		return "create.descriptionMax", "VALUE_OUT_OF_RANGE"
	case !oneOf(a.Create.DefaultJoinType, JoinOpen, JoinClosed, JoinInvite):
		return "create.defaultJoinType", "INVALID_POLICY"
	case !between(a.Create.MaxMinTrophies, 0, 10000000):
		return "create.maxMinTrophies", "VALUE_OUT_OF_RANGE"
	case !between(a.Roles.MaxCoLeaders, 0, 100):
		return "roles.maxCoLeaders", "VALUE_OUT_OF_RANGE"
	case !oneOf(a.Roles.LeaderLeave, LeaveTransfer, LeaveBlock):
		return "roles.leaderLeave", "INVALID_POLICY"
	case !between(a.Cooldowns.JoinAfterLeaveMinutes, 0, 43200):
		return "cooldowns.joinAfterLeaveMinutes", "VALUE_OUT_OF_RANGE"
	case !between(a.Cooldowns.RejoinAfterKickHours, 0, 8760):
		return "cooldowns.rejoinAfterKickHours", "VALUE_OUT_OF_RANGE"
	case !between(a.Cooldowns.LeaveAfterJoinMinutes, 0, 43200):
		return "cooldowns.leaveAfterJoinMinutes", "VALUE_OUT_OF_RANGE"
	case !between(a.Cooldowns.InviteDays, 1, 365):
		return "cooldowns.inviteDays", "VALUE_OUT_OF_RANGE"
	case !between(a.Cooldowns.RequestsPerUser, 1, 100):
		return "cooldowns.requestsPerUser", "VALUE_OUT_OF_RANGE"
	case a.XP.MemberDailyCap < 0 || a.XP.MemberDailyCap > 1e9:
		return "xp.memberDailyCap", "VALUE_OUT_OF_RANGE"
	case a.XP.WarWin < 0 || a.XP.WarWin > 1e9:
		return "xp.warWin", "VALUE_OUT_OF_RANGE"
	case a.XP.WarDraw < 0 || a.XP.WarDraw > 1e9:
		return "xp.warDraw", "VALUE_OUT_OF_RANGE"
	case a.XP.WarLoss < 0 || a.XP.WarLoss > 1e9:
		return "xp.warLoss", "VALUE_OUT_OF_RANGE"
	case !between(a.Display.SearchLimit, 1, 100):
		return "display.searchLimit", "VALUE_OUT_OF_RANGE"
	case !between(a.Display.HistoryLimit, 1, 100):
		return "display.historyLimit", "VALUE_OUT_OF_RANGE"
	case !between(a.Events.Code, 0, 1<<30):
		return "events.code", "VALUE_OUT_OF_RANGE"
	}
	return "", ""
}

var store = settingsrow.New("clan_settings", Defaults)

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
