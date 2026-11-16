package settings

import (
	"context"
	"database/sql"
	"time"

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
}

type All struct {
	Timezone  string    `json:"timezone"`
	Sessions  Sessions  `json:"sessions"`
	Retention Retention `json:"retention"`
	Heatmap   Heatmap   `json:"heatmap"`
}

const (
	SourceClient = "client"
	SourceSocket = "socket"

	ModeClassic = "classic"
	ModeRolling = "rolling"

	MapsAccept = "accept"
	MapsReject = "reject"
)

func Defaults() All {
	return All{
		Timezone:  "Asia/Tehran",
		Sessions:  Sessions{Source: SourceClient, PingSeconds: 60, TimeoutSeconds: 300, MaxSeconds: 6 * 3600, StartsPerHour: 120, CloseOtherOpen: true, KeepDays: 400},
		Retention: Retention{Days: []int{1, 7, 30}, Mode: ModeClassic, CohortRows: 30},
		Heatmap:   Heatmap{Enabled: true, UnknownMaps: MapsAccept, CellSize: 5, MaxPoints: 20000, MaxBytes: 512 * 1024, UploadsPerHour: 120, MaxKmh: 600, KeepRawDays: 14, KeepCellsDays: 365, MaxTracePoints: 20000},
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
