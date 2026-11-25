package settings

import (
	"context"
	"database/sql"
	"regexp"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/eventcontext"
	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

type Sessions struct {
	Source         string `json:"source"`
	PingSeconds    int    `json:"pingSeconds"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
	MaxSeconds     int    `json:"maxSeconds"`
	StartsPerHour  int    `json:"startsPerHour"`
	CloseOtherOpen bool   `json:"closeOtherOpen"`
	KeepDays       int    `json:"keepDays"`
}

type Retention struct {
	Days       []int  `json:"days"`
	Mode       string `json:"mode"`
	CohortRows int    `json:"cohortRows"`
}

type Heatmap struct {
	Enabled        bool    `json:"enabled"`
	UnknownMaps    string  `json:"unknownMaps"`
	CellSize       float64 `json:"cellSize"`
	MaxPoints      int     `json:"maxPoints"`
	MaxBytes       int     `json:"maxBytes"`
	UploadsPerHour int     `json:"uploadsPerHour"`
	MaxKmh         float64 `json:"maxKmh"`
	KeepRawDays    int     `json:"keepRawDays"`
	KeepCellsDays  int     `json:"keepCellsDays"`
	MaxTracePoints int     `json:"maxTracePoints"`
	ImportMaxMB    int     `json:"importMaxMB"`
	ImportMaxRows  int     `json:"importMaxRows"`
	ImportMaxFiles int     `json:"importMaxFiles"`
}

type Events struct {
	Enabled          bool    `json:"enabled"`
	MaxBatch         int     `json:"maxBatch"`
	MaxTags          int     `json:"maxTags"`
	MaxNameLength    int     `json:"maxNameLength"`
	MaxFields        int     `json:"maxFields"`
	MaxStringLength  int     `json:"maxStringLength"`
	ValueMin         float64 `json:"valueMin"`
	ValueMax         float64 `json:"valueMax"`
	BatchesPerMinute int     `json:"batchesPerMinute"`
	EventsPerMinute  int     `json:"eventsPerMinute"`
	PastHours        int     `json:"pastHours"`
	FutureMinutes    int     `json:"futureMinutes"`
	ReceiptDays      int     `json:"receiptDays"`
	UnknownEvents    string  `json:"unknownEvents"`
	UnknownTagGroups string  `json:"unknownTagGroups"`
	UnknownTagValues string  `json:"unknownTagValues"`
	ValueOutOfRange  string  `json:"valueOutOfRange"`
	GameEvents       bool    `json:"gameEvents"`
	SessionEvents    bool    `json:"sessionEvents"`
	FlushSeconds     int     `json:"flushSeconds"`
	MaxQueue         int     `json:"maxQueue"`
}

type Storage struct {
	KeepDays         int `json:"keepDays"`
	BatchSize        int `json:"batchSize"`
	MaxPendingEvents int `json:"maxPendingEvents"`
	DeadAfter        int `json:"deadAfter"`
}

type All struct {
	Timezone  string              `json:"timezone"`
	Sessions  Sessions            `json:"sessions"`
	Retention Retention           `json:"retention"`
	Heatmap   Heatmap             `json:"heatmap"`
	Events    Events              `json:"events"`
	Storage   Storage             `json:"storage"`
	Context   eventcontext.Policy `json:"context"`
}

const (
	SourceClient = "client"
	SourceSocket = "socket"

	ModeClassic = "classic"
	ModeRolling = "rolling"

	MapsAccept = "accept"
	MapsReject = "reject"

	PolicyAccept = "accept"
	PolicyFlag   = "flag"
	PolicyReject = "reject"
)

func DefaultEvents() Events {
	return Events{Enabled: true, MaxBatch: 100, MaxTags: 8, MaxNameLength: 64, MaxFields: 16, MaxStringLength: 256, ValueMin: -1e12, ValueMax: 1e12,
		BatchesPerMinute: 60, EventsPerMinute: 1200, PastHours: 72, FutureMinutes: 5, ReceiptDays: 7,
		UnknownEvents: PolicyAccept, UnknownTagGroups: PolicyAccept, UnknownTagValues: PolicyFlag, ValueOutOfRange: PolicyReject,
		GameEvents: true, SessionEvents: true, FlushSeconds: 30, MaxQueue: 1000}
}

func DefaultContext() eventcontext.Policy {
	return eventcontext.Policy{Enabled: true, Conflict: eventcontext.ConflictOverride, FlagMismatch: true, KeepClientOutside: true, GraceSeconds: 10,
		Fills: []eventcontext.Fill{
			{Group: "mission", Source: "campaign.stage", Enabled: true},
			{Group: "car", Source: "campaign.item", Enabled: true},
			{Group: "chapter", Source: "campaign.chapter", Enabled: false},
		}}
}

var tagGroup = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}$`)

func ValidateContext(c eventcontext.Policy) (string, string) {
	if f, code := c.Validate("context"); code != "" {
		return f, code
	}
	for _, f := range c.Fills {
		if !tagGroup.MatchString(f.Group) {
			return "context.fills.group", "INVALID_TAG_GROUP"
		}
	}
	return "", ""
}

func DefaultStorage() Storage {
	return Storage{KeepDays: 400, BatchSize: 50, MaxPendingEvents: 2000000, DeadAfter: 50}
}

func policyIn(v string, allowed ...string) bool {
	for _, a := range allowed {
		if v == a {
			return true
		}
	}
	return false
}

func (e Events) Validate() (string, string) {
	switch {
	case !between(e.MaxBatch, 1, 500):
		return "events.maxBatch", "VALUE_OUT_OF_RANGE"
	case !between(e.MaxTags, 1, 16):
		return "events.maxTags", "VALUE_OUT_OF_RANGE"
	case !between(e.MaxNameLength, 8, 128):
		return "events.maxNameLength", "VALUE_OUT_OF_RANGE"
	case !between(e.MaxFields, 0, 64):
		return "events.maxFields", "VALUE_OUT_OF_RANGE"
	case !between(e.MaxStringLength, 16, 2048):
		return "events.maxStringLength", "VALUE_OUT_OF_RANGE"
	case e.ValueMin >= e.ValueMax || e.ValueMin < -1e15 || e.ValueMax > 1e15:
		return "events.valueMax", "VALUE_OUT_OF_RANGE"
	case !between(e.BatchesPerMinute, 1, 10000):
		return "events.batchesPerMinute", "VALUE_OUT_OF_RANGE"
	case !between(e.EventsPerMinute, 1, 100000):
		return "events.eventsPerMinute", "VALUE_OUT_OF_RANGE"
	case !between(e.PastHours, 1, 720):
		return "events.pastHours", "VALUE_OUT_OF_RANGE"
	case !between(e.FutureMinutes, 0, 1440):
		return "events.futureMinutes", "VALUE_OUT_OF_RANGE"
	case !between(e.ReceiptDays, 1, 90):
		return "events.receiptDays", "VALUE_OUT_OF_RANGE"
	case !policyIn(e.UnknownEvents, PolicyAccept, PolicyFlag, PolicyReject):
		return "events.unknownEvents", "INVALID_POLICY"
	case !policyIn(e.UnknownTagGroups, PolicyAccept, PolicyFlag, PolicyReject):
		return "events.unknownTagGroups", "INVALID_POLICY"
	case !policyIn(e.UnknownTagValues, PolicyFlag, PolicyReject):
		return "events.unknownTagValues", "INVALID_POLICY"
	case !policyIn(e.ValueOutOfRange, PolicyFlag, PolicyReject):
		return "events.valueOutOfRange", "INVALID_POLICY"
	case !between(e.FlushSeconds, 5, 3600):
		return "events.flushSeconds", "VALUE_OUT_OF_RANGE"
	case !between(e.MaxQueue, 10, 100000):
		return "events.maxQueue", "VALUE_OUT_OF_RANGE"
	}
	return "", ""
}

func (s Storage) Validate() (string, string) {
	switch {
	case !between(s.KeepDays, 0, 3650):
		return "storage.keepDays", "VALUE_OUT_OF_RANGE"
	case !between(s.BatchSize, 1, 500):
		return "storage.batchSize", "VALUE_OUT_OF_RANGE"
	case !between(s.MaxPendingEvents, 1000, 100000000):
		return "storage.maxPendingEvents", "VALUE_OUT_OF_RANGE"
	case !between(s.DeadAfter, 1, 1000):
		return "storage.deadAfter", "VALUE_OUT_OF_RANGE"
	}
	return "", ""
}

func Defaults() All {
	return All{
		Timezone:  "Asia/Tehran",
		Sessions:  Sessions{Source: SourceClient, PingSeconds: 60, TimeoutSeconds: 300, MaxSeconds: 6 * 3600, StartsPerHour: 120, CloseOtherOpen: true, KeepDays: 400},
		Retention: Retention{Days: []int{1, 7, 30}, Mode: ModeClassic, CohortRows: 30},
		Heatmap:   Heatmap{Enabled: true, UnknownMaps: MapsAccept, CellSize: 5, MaxPoints: 20000, MaxBytes: 512 * 1024, UploadsPerHour: 120, MaxKmh: 600, KeepRawDays: 14, KeepCellsDays: 365, MaxTracePoints: 20000, ImportMaxMB: 64, ImportMaxRows: 2000000, ImportMaxFiles: 100},
		Events:    DefaultEvents(),
		Storage:   DefaultStorage(),
		Context:   DefaultContext(),
	}
}

func between(v, lo, hi int) bool { return v >= lo && v <= hi }

func (a All) Location() *time.Location {
	if loc, err := time.LoadLocation(a.Timezone); err == nil {
		return loc
	}
	return time.UTC
}

func (a All) Validate() (string, string) {
	s, r, h := a.Sessions, a.Retention, a.Heatmap
	if _, err := time.LoadLocation(a.Timezone); err != nil || a.Timezone == "" {
		return "timezone", "INVALID_TIMEZONE"
	}
	switch {
	case s.Source != SourceClient && s.Source != SourceSocket:
		return "sessions.source", "INVALID_POLICY"
	case !between(s.PingSeconds, 10, 3600):
		return "sessions.pingSeconds", "VALUE_OUT_OF_RANGE"
	case !between(s.TimeoutSeconds, 30, 86400) || s.TimeoutSeconds < s.PingSeconds:
		return "sessions.timeoutSeconds", "VALUE_OUT_OF_RANGE"
	case !between(s.MaxSeconds, 60, 7*86400):
		return "sessions.maxSeconds", "VALUE_OUT_OF_RANGE"
	case !between(s.StartsPerHour, 1, 10000):
		return "sessions.startsPerHour", "VALUE_OUT_OF_RANGE"
	case !between(s.KeepDays, 0, 3650):
		return "sessions.keepDays", "VALUE_OUT_OF_RANGE"
	case len(r.Days) == 0 || len(r.Days) > 8:
		return "retention.days", "VALUE_OUT_OF_RANGE"
	case r.Mode != ModeClassic && r.Mode != ModeRolling:
		return "retention.mode", "INVALID_POLICY"
	case !between(r.CohortRows, 1, 120):
		return "retention.cohortRows", "VALUE_OUT_OF_RANGE"
	case h.UnknownMaps != MapsAccept && h.UnknownMaps != MapsReject:
		return "heatmap.unknownMaps", "INVALID_POLICY"
	case h.CellSize < 0.1 || h.CellSize > 1000:
		return "heatmap.cellSize", "VALUE_OUT_OF_RANGE"
	case !between(h.MaxPoints, 10, 200000):
		return "heatmap.maxPoints", "VALUE_OUT_OF_RANGE"
	case !between(h.MaxBytes, 4*1024, 8<<20):
		return "heatmap.maxBytes", "VALUE_OUT_OF_RANGE"
	case !between(h.UploadsPerHour, 1, 10000):
		return "heatmap.uploadsPerHour", "VALUE_OUT_OF_RANGE"
	case h.MaxKmh < 1 || h.MaxKmh > 100000:
		return "heatmap.maxKmh", "VALUE_OUT_OF_RANGE"
	case !between(h.KeepRawDays, 0, 3650):
		return "heatmap.keepRawDays", "VALUE_OUT_OF_RANGE"
	case !between(h.KeepCellsDays, 0, 3650):
		return "heatmap.keepCellsDays", "VALUE_OUT_OF_RANGE"
	case !between(h.MaxTracePoints, 100, 500000):
		return "heatmap.maxTracePoints", "VALUE_OUT_OF_RANGE"
	case !between(h.ImportMaxMB, 1, 2048):
		return "heatmap.importMaxMB", "VALUE_OUT_OF_RANGE"
	case !between(h.ImportMaxRows, 100, 50000000):
		return "heatmap.importMaxRows", "VALUE_OUT_OF_RANGE"
	case !between(h.ImportMaxFiles, 1, 2000):
		return "heatmap.importMaxFiles", "VALUE_OUT_OF_RANGE"
	}
	if f, c := a.Events.Validate(); c != "" {
		return f, c
	}
	if f, c := a.Storage.Validate(); c != "" {
		return f, c
	}
	if f, c := ValidateContext(a.Context); c != "" {
		return f, c
	}
	seen := map[int]bool{}
	for _, d := range r.Days {
		if !between(d, 1, 365) || seen[d] {
			return "retention.days", "VALUE_OUT_OF_RANGE"
		}
		seen[d] = true
	}
	return "", ""
}

var store = settingsrow.New("insights_settings", Defaults)

func Load(ctx context.Context, db *sql.DB) (All, error) { return store.Load(ctx, db) }

func Invalidate() { store.Invalidate() }

func Read(ctx context.Context, db *sql.DB) (All, settingsrow.Meta, error) { return store.Read(ctx, db) }

func Save(ctx context.Context, db *sql.DB, in All, by string) error {
	return store.Save(ctx, db, in, by)
}
