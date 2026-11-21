package settings

import (
	"context"
	"database/sql"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

type Unlock struct {
	PriceMode        string `json:"priceMode"`
	OwnFreeItems     bool   `json:"ownFreeItems"`
	CheckRequirement bool   `json:"checkRequirement"`
	RequireRequestID bool   `json:"requireRequestId"`
}

type Upgrade struct {
	RequireFromLevel bool  `json:"requireFromLevel"`
	RequireRequestID bool  `json:"requireRequestId"`
	MaxTracks        int   `json:"maxTracks"`
	MaxLevels        int   `json:"maxLevels"`
	StoppedItems     *bool `json:"stoppedItems"`
}

type Import struct {
	Mode          string `json:"mode"`
	MissingFields string `json:"missingFields"`
	DefaultType   string `json:"defaultType"`
	MaxItems      int    `json:"maxItems"`
	KeepReports   int    `json:"keepReports"`
}

type All struct {
	Unlock       Unlock  `json:"unlock"`
	Upgrade      Upgrade `json:"upgrade"`
	RequestHours int     `json:"requestHours"`
	Import       Import  `json:"import"`
}

const (
	PriceAll    = "all"
	PriceChoice = "choice"
)

func (u Upgrade) AllowsStopped() bool { return u.StoppedItems == nil || *u.StoppedItems }

var on = true

func Defaults() All {
	return All{
		Unlock:       Unlock{PriceMode: PriceAll, OwnFreeItems: true},
		Upgrade:      Upgrade{MaxTracks: 20, MaxLevels: 50, StoppedItems: &on},
		RequestHours: 24,
		Import:       Import{Mode: "atomic", MissingFields: "keep", MaxItems: 1000, KeepReports: 50},
	}
}

func (a All) Validate() (string, string) {
	switch {
	case a.Unlock.PriceMode != PriceAll && a.Unlock.PriceMode != PriceChoice:
		return "unlock.priceMode", "INVALID_POLICY"
	case a.Import.Mode != "atomic" && a.Import.Mode != "partial":
		return "import.mode", "INVALID_POLICY"
	case a.Import.MissingFields != "keep" && a.Import.MissingFields != "reset":
		return "import.missingFields", "INVALID_POLICY"
	case len(a.Import.DefaultType) > 64:
		return "import.defaultType", "INVALID_KEY"
	case a.Import.MaxItems < 1 || a.Import.MaxItems > 5000:
		return "import.maxItems", "VALUE_OUT_OF_RANGE"
	case a.Import.KeepReports < 1 || a.Import.KeepReports > 1000:
		return "import.keepReports", "VALUE_OUT_OF_RANGE"
	case a.Upgrade.MaxTracks < 1 || a.Upgrade.MaxTracks > 50:
		return "upgrade.maxTracks", "VALUE_OUT_OF_RANGE"
	case a.Upgrade.MaxLevels < 1 || a.Upgrade.MaxLevels > 200:
		return "upgrade.maxLevels", "VALUE_OUT_OF_RANGE"
	case a.RequestHours < 1 || a.RequestHours > 720:
		return "requestHours", "VALUE_OUT_OF_RANGE"
	}
	return "", ""
}

var store = settingsrow.New("inventoryx_settings", Defaults)

func Load(ctx context.Context, db *sql.DB) (All, error) { return store.Load(ctx, db) }

func Invalidate() { store.Invalidate() }

func Read(ctx context.Context, db *sql.DB) (All, settingsrow.Meta, error) { return store.Read(ctx, db) }

func Save(ctx context.Context, db *sql.DB, in All, by string) error {
	return store.Save(ctx, db, in, by)
}
