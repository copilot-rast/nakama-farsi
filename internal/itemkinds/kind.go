package itemkinds

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Text = map[string]string

type Field struct {
	Key      string   `json:"key"`
	Label    Text     `json:"label"`
	Type     string   `json:"type"`
	Options  []string `json:"options,omitempty"`
	Fields   []Field  `json:"fields,omitempty"`
	Hint     Text     `json:"hint,omitempty"`
	Min      *float64 `json:"min,omitempty"`
	Max      *float64 `json:"max,omitempty"`
	ReadOnly bool     `json:"readOnly,omitempty"`
	Default  any      `json:"default,omitempty"`
	MaxItems int      `json:"maxItems,omitempty"`
}

type Column struct {
	Key   string `json:"key"`
	Label Text   `json:"label"`
	Path  string `json:"path"`
}

type Ownable struct {
	Field   string `json:"field"`
	Label   Text   `json:"label"`
	Default string `json:"default"`
}

type Route struct {
	ID         int64          `json:"id"`
	SourceType string         `json:"sourceType"`
	Match      map[string]any `json:"match"`
	Kind       string         `json:"kind"`
	Role       string         `json:"role"`
	Priority   int            `json:"priority"`
	Enabled    bool           `json:"enabled"`
	Note       string         `json:"note"`
	Origin     string         `json:"origin"`
	UpdatedBy  string         `json:"updatedBy"`
	UpdatedAt  string         `json:"updatedAt,omitempty"`
}

type Price struct {
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
}

type Mapped struct {
	Key          string
	MatchMeta    map[string]string
	Name         Text
	Description  Text
	Price        *Price
	NoPrice      bool
	UnlockLevel  *int
	ParentKey    *string
	RequiresKey  *string
	Active       *bool
	Sort         *int
	Metadata     map[string]any
	UnlockPrices []Price
	HasUnlock    bool
	Data         map[string]any
	Pending      string
	Skip         string
	Rank         int
}

type Profile struct {
	ItemID        string         `json:"itemId"`
	ItemKey       string         `json:"itemKey"`
	ParentKey     string         `json:"parentKey,omitempty"`
	Kind          string         `json:"kind"`
	Data          map[string]any `json:"data"`
	Status        string         `json:"status"`
	Reason        string         `json:"reason,omitempty"`
	SchemaVersion int            `json:"schemaVersion"`
	SourceAssetID string         `json:"sourceAssetId,omitempty"`
	SourceType    string         `json:"sourceType,omitempty"`
	SourceHash    string         `json:"sourceHash,omitempty"`
	SourceVersion int            `json:"sourceVersion"`
	LinkedAt      string         `json:"linkedAt,omitempty"`
	UpdatedBy     string         `json:"updatedBy,omitempty"`
	UpdatedAt     string         `json:"updatedAt,omitempty"`
}

type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type ViewInput struct {
	Q        Querier
	UserID   string
	Profiles []Profile
	Config   map[string]any
}

type BuyInput struct {
	Tx      *sql.Tx
	UserID  string
	ItemKey string
	Profile Profile
	Config  map[string]any
}

type Refusal struct {
	Code    string
	Message string
	Meta    map[string]any
}

func (r *Refusal) Error() string { return r.Code + ": " + r.Message }

type Kind struct {
	Key           string
	ItemType      string
	Label         Text
	Icon          string
	Order         int
	SchemaVersion int
	Fields        []Field
	Columns       []Column
	Settings      []Field
	Defaults      map[string]any
	Ownership     []Ownable
	Routes        []Route
	KeyOf         func(e *Entry) string
	Map           func(ctx context.Context, m *Mapper, e *Entry) (*Mapped, error)
	Validate      func(data map[string]any) error
	View          func(ctx context.Context, in ViewInput) (map[string]map[string]any, error)
	CheckBuy      func(ctx context.Context, in BuyInput) error
	Slot          func(p Profile) string
}

var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

var (
	regMu sync.RWMutex
	kinds = map[string]*Kind{}
)

func Register(k Kind) error {
	if !keyPattern.MatchString(k.Key) {
		return errors.New("itemkinds: invalid kind key " + k.Key)
	}
	if strings.TrimSpace(k.ItemType) == "" {
		return errors.New("itemkinds: kind " + k.Key + " needs an item type")
	}
	if k.SchemaVersion < 1 {
		k.SchemaVersion = 1
	}
	if k.Defaults == nil {
		k.Defaults = map[string]any{}
	}
	regMu.Lock()
	defer regMu.Unlock()
	kk := k
	kinds[k.Key] = &kk
	return nil
}

func Get(key string) *Kind {
	regMu.RLock()
	defer regMu.RUnlock()
	return kinds[key]
}

func List() []*Kind {
	regMu.RLock()
	out := make([]*Kind, 0, len(kinds))
	for _, k := range kinds {
		out = append(out, k)
	}
	regMu.RUnlock()
	sort.Slice(out, func(i, j int) bool {
		if out[i].Order != out[j].Order {
			return out[i].Order < out[j].Order
		}
		return out[i].Key < out[j].Key
	})
	return out
}

func Install(ctx context.Context, db *sql.DB, k Kind) error {
	if err := Register(k); err != nil {
		return err
	}
	if db == nil {
		return nil
	}
	name, _ := json.Marshal(k.Label)
	if _, err := db.ExecContext(ctx, `INSERT INTO base_category (id, key_type, key, name, sort) VALUES (gen_random_uuid(), 'item_type', $1, $2, 50) ON CONFLICT (key_type, key) DO NOTHING`, k.ItemType, string(name)); err != nil {
		return err
	}
	for _, r := range k.Routes {
		origin := r.Origin
		if origin == "" {
			origin = k.Key + ":" + r.SourceType
		}
		match, _ := json.Marshal(nonNilMap(r.Match))
		kind := r.Kind
		if kind == "" && r.Role == RoleItem {
			kind = k.Key
		}
		prio := r.Priority
		if prio == 0 {
			prio = 100
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO itemkind_route (source_type, match, kind, role, priority, enabled, note, origin, updated_by)
			VALUES ($1,$2,$3,$4,$5,true,$6,$7,'kind:'||$8) ON CONFLICT (origin) WHERE origin <> '' DO NOTHING`,
			r.SourceType, string(match), kind, r.Role, prio, r.Note, origin, k.Key); err != nil {
			return err
		}
	}
	return nil
}

const (
	RoleItem      = "item"
	RoleDependent = "dependent"
	RoleIgnore    = "ignore"
)

func nonNilMap(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return m
}
