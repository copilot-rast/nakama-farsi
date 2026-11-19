package economyhooks

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"sync"
	"time"
)

type Refusal struct {
	Status  int
	Code    string
	Message string
	Meta    map[string]any
}

func (r *Refusal) Error() string {
	if r.Message != "" {
		return r.Code + ": " + r.Message
	}
	return r.Code
}

func AsRefusal(err error) (*Refusal, bool) {
	var r *Refusal
	if errors.As(err, &r) {
		return r, true
	}
	return nil, false
}

type Purchase struct {
	UserID     string
	ProductID  string
	ProductKey string
	Method     string
	Sandbox    bool
	Restore    bool
}

type ItemState struct {
	Unlocked bool

	LockMeta map[string]any

	Fields map[string]any
}

type (
	PurchaseCheck func(ctx context.Context, p Purchase) error
	ProductFields func(ctx context.Context, userID, lang string, productIDs []string) (map[string]map[string]any, error)
	ItemStates    func(ctx context.Context, userID, lang string, itemIDs []string) (map[string]ItemState, error)

	LevelReader func(ctx context.Context, userID, tableID string) (int, error)

	MethodReader func(ctx context.Context) (map[string]bool, error)
)

var (
	ErrNoLevels  = errors.New("player levels are not available: the economy plugin is not loaded")
	ErrNoMethods = errors.New("payment methods are not available: the economy plugin is not loaded")

	mu       sync.RWMutex
	checks   = map[string]PurchaseCheck{}
	products = map[string]ProductFields{}
	items    = map[string]ItemStates{}
	levels   LevelReader
	methods  MethodReader
)

func SetPurchaseCheck(owner string, fn PurchaseCheck) { set(checks, owner, fn) }
func SetProductFields(owner string, fn ProductFields) { set(products, owner, fn) }
func SetItemStates(owner string, fn ItemStates)       { set(items, owner, fn) }

func SetLevelReader(fn LevelReader) {
	mu.Lock()
	defer mu.Unlock()
	levels = fn
}

func SetMethodReader(fn MethodReader) {
	mu.Lock()
	defer mu.Unlock()
	methods = fn
}

func set[T any](m map[string]T, owner string, fn T) {
	mu.Lock()
	defer mu.Unlock()
	m[owner] = fn
}

func sorted[T any](m map[string]T) []T {
	mu.RLock()
	defer mu.RUnlock()
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]T, 0, len(keys))
	for _, k := range keys {
		out = append(out, m[k])
	}
	return out
}

func CheckPurchase(ctx context.Context, p Purchase) error {
	for _, fn := range sorted(checks) {
		if fn == nil {
			continue
		}
		if err := fn(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

func MergeProducts(ctx context.Context, userID, lang string, list []map[string]any) {
	fns := sorted(products)
	if len(fns) == 0 || len(list) == 0 {
		return
	}
	ids := make([]string, 0, len(list))
	for _, p := range list {
		if id, ok := p["id"].(string); ok {
			ids = append(ids, id)
		}
	}
	for _, fn := range fns {
		if fn == nil {
			continue
		}
		extra, err := fn(ctx, userID, lang, ids)
		if err != nil {
			continue
		}
		for _, p := range list {
			id, _ := p["id"].(string)
			for k, v := range extra[id] {
				if _, ok := p[k]; !ok {
					p[k] = v
				}
			}
		}
	}
}

func Items(ctx context.Context, userID, lang string, itemIDs []string) map[string]ItemState {
	out := map[string]ItemState{}
	for _, fn := range sorted(items) {
		if fn == nil || len(itemIDs) == 0 {
			continue
		}
		got, err := fn(ctx, userID, lang, itemIDs)
		if err != nil {
			continue
		}
		for id, st := range got {
			cur := out[id]
			cur.Unlocked = cur.Unlocked || st.Unlocked
			cur.LockMeta = mergeMap(cur.LockMeta, st.LockMeta)
			cur.Fields = mergeMap(cur.Fields, st.Fields)
			out[id] = cur
		}
	}
	return out
}

func Fill(v map[string]any, fields map[string]any) {
	for k, x := range fields {
		if _, ok := v[k]; !ok {
			v[k] = x
		}
	}
}

func mergeMap(dst, src map[string]any) map[string]any {
	if len(src) == 0 {
		return dst
	}
	if dst == nil {
		dst = map[string]any{}
	}
	Fill(dst, src)
	return dst
}

func Level(ctx context.Context, userID, tableID string) (int, error) {
	mu.RLock()
	fn := levels
	mu.RUnlock()
	if fn == nil {
		return 0, ErrNoLevels
	}
	return fn(ctx, userID, tableID)
}

func Methods(ctx context.Context) (map[string]bool, error) {
	mu.RLock()
	fn := methods
	mu.RUnlock()
	if fn == nil {
		return nil, ErrNoMethods
	}
	return fn(ctx)
}

type PrizeGrant struct {
	UserID     string
	PrizeKey   string
	GroupKey   string
	GroupMode  string
	SourceType string
	SourceID   string
	Reason     string
}

const (
	GroupDraw = "draw"

	GroupAll = "all"
)

type PrizeResult struct {
	Prizes []map[string]any `json:"prizes"`
}

type (
	PrizeGranter func(ctx context.Context, g PrizeGrant) (PrizeResult, error)

	MetaSettler func(ctx context.Context, userID string, metaKeys []string) (map[string]any, error)
)

var (
	ErrNoPrizes = errors.New("economy prizes are not available: the economy plugin is not loaded")

	prizeGranter PrizeGranter
	metaSettler  MetaSettler
)

func SetPrizeGranter(fn PrizeGranter) {
	mu.Lock()
	defer mu.Unlock()
	prizeGranter = fn
}

func SetMetaSettler(fn MetaSettler) {
	mu.Lock()
	defer mu.Unlock()
	metaSettler = fn
}

func PrizesAvailable() bool {
	mu.RLock()
	defer mu.RUnlock()
	return prizeGranter != nil
}

func GrantPrize(ctx context.Context, g PrizeGrant) (PrizeResult, error) {
	mu.RLock()
	fn := prizeGranter
	mu.RUnlock()
	if fn == nil {
		return PrizeResult{}, ErrNoPrizes
	}
	return fn(ctx, g)
}

func SettleMetas(ctx context.Context, userID string, metaKeys []string) (map[string]any, error) {
	mu.RLock()
	fn := metaSettler
	mu.RUnlock()
	if fn == nil || len(metaKeys) == 0 {
		return nil, nil
	}
	return fn(ctx, userID, metaKeys)
}

type ItemSaver func(ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error

var (
	ErrNoItemSaver = errors.New("economy items are not available: the economy plugin is not loaded")

	itemSaver ItemSaver
)

func SetItemSaver(fn ItemSaver) {
	mu.Lock()
	defer mu.Unlock()
	itemSaver = fn
}

func SaveItem(ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	mu.RLock()
	fn := itemSaver
	mu.RUnlock()
	if fn == nil {
		return ErrNoItemSaver
	}
	return fn(ctx, tx, id, create, raw)
}

type RegenPrice struct {
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
}

type Regen struct {
	CurrencyKey     string      `json:"currencyKey"`
	Value           int64       `json:"value"`
	Regen           bool        `json:"regen"`
	Max             int64       `json:"max"`
	IntervalSeconds int64       `json:"intervalSeconds"`
	NextAt          *time.Time  `json:"nextAt"`
	FullAt          *time.Time  `json:"fullAt"`
	Unlimited       bool        `json:"unlimited"`
	UnlimitedUntil  *time.Time  `json:"unlimitedUntil"`
	RefillPrice     *RegenPrice `json:"refillPrice"`
}

type RegenSpend struct {
	View     Regen `json:"view"`
	Spent    int64 `json:"spent"`
	Free     bool  `json:"free"`
	Replayed bool  `json:"replayed"`
}

type RegenRefund struct {
	View     Regen `json:"view"`
	Refunded int64 `json:"refunded"`
	Replayed bool  `json:"replayed"`
	Found    bool  `json:"found"`
}

type (
	RegenViewer   func(ctx context.Context, userID, currency string) (Regen, error)
	RegenSpender  func(ctx context.Context, userID, currency string, amount int64, sourceID string) (RegenSpend, error)
	RegenRefunder func(ctx context.Context, userID, currency, sourceID string) (RegenRefund, error)
)

var (
	ErrNoRegen = errors.New("regenerating currencies are not available: the economy plugin is not loaded")

	regenView   RegenViewer
	regenSpend  RegenSpender
	regenRefund RegenRefunder
)

func SetRegen(view RegenViewer, spend RegenSpender, refund RegenRefunder) {
	mu.Lock()
	defer mu.Unlock()
	regenView, regenSpend, regenRefund = view, spend, refund
}

func RegenAvailable() bool {
	mu.RLock()
	defer mu.RUnlock()
	return regenSpend != nil
}

func RegenView(ctx context.Context, userID, currency string) (Regen, error) {
	mu.RLock()
	fn := regenView
	mu.RUnlock()
	if fn == nil {
		return Regen{}, ErrNoRegen
	}
	return fn(ctx, userID, currency)
}

func SpendRegen(ctx context.Context, userID, currency string, amount int64, sourceID string) (RegenSpend, error) {
	mu.RLock()
	fn := regenSpend
	mu.RUnlock()
	if fn == nil {
		return RegenSpend{}, ErrNoRegen
	}
	return fn(ctx, userID, currency, amount, sourceID)
}

func RefundRegen(ctx context.Context, userID, currency, sourceID string) (RegenRefund, error) {
	mu.RLock()
	fn := regenRefund
	mu.RUnlock()
	if fn == nil {
		return RegenRefund{}, ErrNoRegen
	}
	return fn(ctx, userID, currency, sourceID)
}

type (
	CatalogReader  func(ctx context.Context, userID, lang string) (map[string]any, error)
	BalancesReader func(ctx context.Context, userID string) (map[string]any, error)
)

var (
	ErrNoCatalog = errors.New("the economy catalog is not available: the economy plugin is not loaded")

	catalogReader  CatalogReader
	balancesReader BalancesReader
)

func SetCatalog(catalog CatalogReader, balances BalancesReader) {
	mu.Lock()
	defer mu.Unlock()
	catalogReader, balancesReader = catalog, balances
}

func Catalog(ctx context.Context, userID, lang string) (map[string]any, error) {
	mu.RLock()
	fn := catalogReader
	mu.RUnlock()
	if fn == nil {
		return nil, ErrNoCatalog
	}
	return fn(ctx, userID, lang)
}

func Balances(ctx context.Context, userID string) (map[string]any, error) {
	mu.RLock()
	fn := balancesReader
	mu.RUnlock()
	if fn == nil {
		return nil, ErrNoCatalog
	}
	return fn(ctx, userID)
}

type LevelProgress struct {
	TableKey   string
	Level      int
	MaxLevel   int
	Value      float64
	LevelStart float64
	Next       *float64
}

type LevelProgressReader func(ctx context.Context, userID, tableKey string) (LevelProgress, error)

var (
	ErrNoLevelTable = errors.New("level table not found or inactive")

	progressReader LevelProgressReader
)

func SetLevelProgressReader(fn LevelProgressReader) {
	mu.Lock()
	defer mu.Unlock()
	progressReader = fn
}

func LevelProgressOf(ctx context.Context, userID, tableKey string) (LevelProgress, error) {
	mu.RLock()
	fn := progressReader
	mu.RUnlock()
	if fn == nil {
		return LevelProgress{}, ErrNoLevels
	}
	return fn(ctx, userID, tableKey)
}
