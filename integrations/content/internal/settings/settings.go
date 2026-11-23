package settings

import (
	"context"
	"database/sql"
	"net/url"
	"regexp"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

type Packs struct {
	Platforms         []string `json:"platforms"`
	PartBytes         int      `json:"partBytes"`
	MaxFileBytes      int64    `json:"maxFileBytes"`
	MaxPackBytes      int64    `json:"maxPackBytes"`
	MaxFiles          int      `json:"maxFiles"`
	UploadHours       int      `json:"uploadHours"`
	KeepInactive      int      `json:"keepInactive"`
	SingleActive      bool     `json:"singleActive"`
	RequireAppVersion bool     `json:"requireAppVersion"`
	PublicIndex       bool     `json:"publicIndex"`
	IndexCacheSeconds int      `json:"indexCacheSeconds"`
	BaseURL           string   `json:"baseUrl"`
	RedirectToStorage bool     `json:"redirectToStorage"`
	CatalogPattern    string   `json:"catalogPattern"`
	RequireHash       bool     `json:"requireHash"`
	UploadParallel    int      `json:"uploadParallel"`
	UploadRetries     int      `json:"uploadRetries"`
	AssembleWorkers   int      `json:"assembleWorkers"`
	AssembleLeaseSec  int      `json:"assembleLeaseSeconds"`
	AssembleSweepSec  int      `json:"assembleSweepSeconds"`
}

type All struct {
	Packs Packs `json:"packs"`
}

var PlatformPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,31}$`)

const MaxPartBytes = 3 << 20

func Defaults() All {
	return All{
		Packs: Packs{
			Platforms:    []string{"Android", "iOS", "StandaloneWindows64", "StandaloneOSX", "StandaloneLinux64", "WebGL"},
			PartBytes:    1 << 20,
			MaxFileBytes: 256 << 20, MaxPackBytes: 2 << 30, MaxFiles: 5000,
			UploadHours: 24, KeepInactive: 5, SingleActive: false, RequireAppVersion: false,
			PublicIndex: true, IndexCacheSeconds: 60, RedirectToStorage: true,
			CatalogPattern: `^catalog[^/]*\.(json|bin)$`, RequireHash: false,
			UploadParallel: 4, UploadRetries: 5,
			AssembleWorkers: 2, AssembleLeaseSec: 120, AssembleSweepSec: 30,
		},
	}
}

func (a All) Validate() (string, string) {
	p := a.Packs
	switch {
	case len(p.Platforms) > 32:
		return "packs.platforms", "VALUE_OUT_OF_RANGE"
	case p.PartBytes < 64*1024 || p.PartBytes > MaxPartBytes:
		return "packs.partBytes", "VALUE_OUT_OF_RANGE"
	case p.MaxFileBytes < 1024 || p.MaxFileBytes > 4<<30:
		return "packs.maxFileBytes", "VALUE_OUT_OF_RANGE"
	case p.MaxPackBytes < p.MaxFileBytes || p.MaxPackBytes > 64<<30:
		return "packs.maxPackBytes", "VALUE_OUT_OF_RANGE"
	case p.MaxFiles < 1 || p.MaxFiles > 100000:
		return "packs.maxFiles", "VALUE_OUT_OF_RANGE"
	case p.UploadHours < 1 || p.UploadHours > 720:
		return "packs.uploadHours", "VALUE_OUT_OF_RANGE"
	case p.KeepInactive < 0 || p.KeepInactive > 1000:
		return "packs.keepInactive", "VALUE_OUT_OF_RANGE"
	case p.IndexCacheSeconds < 0 || p.IndexCacheSeconds > 86400:
		return "packs.indexCacheSeconds", "VALUE_OUT_OF_RANGE"
	case p.UploadParallel < 1 || p.UploadParallel > 8:
		return "packs.uploadParallel", "VALUE_OUT_OF_RANGE"
	case p.UploadRetries < 0 || p.UploadRetries > 20:
		return "packs.uploadRetries", "VALUE_OUT_OF_RANGE"
	case p.AssembleWorkers < 1 || p.AssembleWorkers > 16:
		return "packs.assembleWorkers", "VALUE_OUT_OF_RANGE"
	case p.AssembleLeaseSec < 30 || p.AssembleLeaseSec > 3600:
		return "packs.assembleLeaseSeconds", "VALUE_OUT_OF_RANGE"
	case p.AssembleSweepSec < 5 || p.AssembleSweepSec > 3600:
		return "packs.assembleSweepSeconds", "VALUE_OUT_OF_RANGE"
	}
	seen := map[string]bool{}
	for _, pl := range p.Platforms {
		if !PlatformPattern.MatchString(pl) || seen[pl] {
			return "packs.platforms", "INVALID_PLATFORM"
		}
		seen[pl] = true
	}
	if p.BaseURL != "" {
		u, err := url.Parse(p.BaseURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.RawQuery != "" {
			return "packs.baseUrl", "INVALID_URL"
		}
	}
	if _, err := regexp.Compile(p.CatalogPattern); err != nil || p.CatalogPattern == "" || len(p.CatalogPattern) > 200 {
		return "packs.catalogPattern", "INVALID_PATTERN"
	}
	return "", ""
}

func (p Packs) PlatformAllowed(name string) bool {
	if !PlatformPattern.MatchString(name) {
		return false
	}
	if len(p.Platforms) == 0 {
		return true
	}
	for _, x := range p.Platforms {
		if x == name {
			return true
		}
	}
	return false
}

var store = settingsrow.New("content_settings", Defaults)

func Load(ctx context.Context, db *sql.DB) (All, error) { return store.Load(ctx, db) }

func Invalidate() { store.Invalidate() }

func Read(ctx context.Context, db *sql.DB) (All, settingsrow.Meta, error) { return store.Read(ctx, db) }

func Save(ctx context.Context, db *sql.DB, in All, by string) error {
	return store.Save(ctx, db, in, by)
}
