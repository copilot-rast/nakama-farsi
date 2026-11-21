// Package settings holds the admin policies of the assets module (one JSON row in assets_settings).
package settings

import (
	"context"
	"database/sql"
	"net/url"
	"regexp"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

type Entries struct {
	// MaxEntryBytes: largest data document (compact JSON) of one entry.
	MaxEntryBytes int `json:"maxEntryBytes"`
	// MaxPushItems: items + deletes in one push.
	MaxPushItems int `json:"maxPushItems"`
	// MaxFetchItems: entries returned by one fetch (the answer says when more are left).
	MaxFetchItems int `json:"maxFetchItems"`
	// MaxEntries: entries the store may hold (0 = no limit).
	MaxEntries int `json:"maxEntries"`
	// ConflictMode: "reject" refuses the whole push when one item is stale (nothing is written),
	// "skip" writes the others and reports the stale ones as conflicts.
	ConflictMode string `json:"conflictMode"`
	// AllowForce lets a push with force=true overwrite stale items.
	AllowForce bool `json:"allowForce"`
	// Access to the game reads: "session" (RPCs with a player session) or "public"
	// (also GET /v2/assets/manifest and POST /v2/assets/fetch without a session).
	Access string `json:"access"`
	// CacheSeconds of the public manifest answer (it always carries an ETag).
	CacheSeconds int `json:"cacheSeconds"`
}

type History struct {
	// KeepVersions per entry (0 = keep all).
	KeepVersions int `json:"keepVersions"`
	// KeepDays: older states are removed, except the newest KeepVersions (0 = no age limit).
	KeepDays int `json:"keepDays"`
	// KeepDeletedDays: how long a deleted entry can be restored (0 = forever).
	KeepDeletedDays int `json:"keepDeletedDays"`
}

type Packs struct {
	// Platforms a pack may name (Addressables build targets); empty = any name.
	Platforms []string `json:"platforms"`
	// PartBytes: size of one upload part; keep it below the server's socket.max_request_size_bytes.
	PartBytes int `json:"partBytes"`
	// MaxFileBytes / MaxPackBytes / MaxFiles limit one upload.
	MaxFileBytes int64 `json:"maxFileBytes"`
	MaxPackBytes int64 `json:"maxPackBytes"`
	MaxFiles     int   `json:"maxFiles"`
	// UploadHours: a pack that is not completed in time is removed.
	UploadHours int `json:"uploadHours"`
	// KeepInactive: ready packs that are not active kept per platform (older ones and their files are removed; 0 = keep all).
	KeepInactive int `json:"keepInactive"`
	// ActivateFromEditor: an upload may ask to activate its pack at once (else only the console activates).
	ActivateFromEditor bool `json:"activateFromEditor"`
	// SingleActive: activating a pack deactivates the other packs of its platform whose app range overlaps.
	SingleActive bool `json:"singleActive"`
	// PublicIndex: GET /v2/assets/content-index answers without a session (Addressables start before login).
	PublicIndex bool `json:"publicIndex"`
	// IndexCacheSeconds of the content index answer.
	IndexCacheSeconds int `json:"indexCacheSeconds"`
	// BaseURL where the game downloads pack files (a CDN in front of the server); empty = the server's own URL.
	BaseURL string `json:"baseUrl"`
	// RedirectToStorage: with an S3 store that has a public URL, file requests are redirected there.
	RedirectToStorage bool `json:"redirectToStorage"`
	// CatalogPattern finds the Addressables catalog among the files (the .hash next to it is optional).
	CatalogPattern string `json:"catalogPattern"`
	// RequireHash: a pack without the catalog's .hash file is refused.
	RequireHash bool `json:"requireHash"`
}

type All struct {
	Entries Entries `json:"entries"`
	History History `json:"history"`
	Packs   Packs   `json:"packs"`
}

const (
	ConflictReject = "reject"
	ConflictSkip   = "skip"
	AccessSession  = "session"
	AccessPublic   = "public"
)

var (
	PlatformPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,31}$`)
)

func Defaults() All {
	return All{
		Entries: Entries{MaxEntryBytes: 1 << 20, MaxPushItems: 500, MaxFetchItems: 200, ConflictMode: ConflictReject, Access: AccessSession, CacheSeconds: 0},
		History: History{KeepVersions: 20, KeepDays: 0, KeepDeletedDays: 90},
		Packs: Packs{
			Platforms:    []string{"Android", "iOS", "StandaloneWindows64", "StandaloneOSX", "StandaloneLinux64", "WebGL"},
			PartBytes:    192 * 1024,
			MaxFileBytes: 256 << 20, MaxPackBytes: 2 << 30, MaxFiles: 5000,
			UploadHours: 24, KeepInactive: 5, ActivateFromEditor: true, SingleActive: false,
			PublicIndex: true, IndexCacheSeconds: 60, RedirectToStorage: true,
			CatalogPattern: `^catalog[^/]*\.(json|bin)$`, RequireHash: false,
		},
	}
}

// Validate returns the failing field and an error code, or two empty strings.
func (a All) Validate() (string, string) {
	e, h, p := a.Entries, a.History, a.Packs
	switch {
	case e.MaxEntryBytes < 1024 || e.MaxEntryBytes > 16<<20:
		return "entries.maxEntryBytes", "VALUE_OUT_OF_RANGE"
	case e.MaxPushItems < 1 || e.MaxPushItems > 5000:
		return "entries.maxPushItems", "VALUE_OUT_OF_RANGE"
	case e.MaxFetchItems < 1 || e.MaxFetchItems > 2000:
		return "entries.maxFetchItems", "VALUE_OUT_OF_RANGE"
	case e.MaxEntries < 0 || e.MaxEntries > 10_000_000:
		return "entries.maxEntries", "VALUE_OUT_OF_RANGE"
	case e.ConflictMode != ConflictReject && e.ConflictMode != ConflictSkip:
		return "entries.conflictMode", "INVALID_POLICY"
	case e.Access != AccessSession && e.Access != AccessPublic:
		return "entries.access", "INVALID_POLICY"
	case e.CacheSeconds < 0 || e.CacheSeconds > 86400:
		return "entries.cacheSeconds", "VALUE_OUT_OF_RANGE"
	case h.KeepVersions < 0 || h.KeepVersions > 1000:
		return "history.keepVersions", "VALUE_OUT_OF_RANGE"
	case h.KeepDays < 0 || h.KeepDays > 3650:
		return "history.keepDays", "VALUE_OUT_OF_RANGE"
	case h.KeepDeletedDays < 0 || h.KeepDeletedDays > 3650:
		return "history.keepDeletedDays", "VALUE_OUT_OF_RANGE"
	case len(p.Platforms) > 32:
		return "packs.platforms", "VALUE_OUT_OF_RANGE"
	case p.PartBytes < 16*1024 || p.PartBytes > 64<<20:
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
	}
	for _, pl := range p.Platforms {
		if !PlatformPattern.MatchString(pl) {
			return "packs.platforms", "INVALID_PLATFORM"
		}
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

// PlatformAllowed reports whether a pack may use this platform name.
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

var store = settingsrow.New("assets_settings", Defaults)

func Load(ctx context.Context, db *sql.DB) (All, error) { return store.Load(ctx, db) }

func Invalidate() { store.Invalidate() }

func Read(ctx context.Context, db *sql.DB) (All, settingsrow.Meta, error) { return store.Read(ctx, db) }

func Save(ctx context.Context, db *sql.DB, in All, by string) error {
	return store.Save(ctx, db, in, by)
}
