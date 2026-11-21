package gamecatalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Definition struct {
	Key      string            `json:"key"`
	Name     map[string]string `json:"name,omitempty"`
	Plugin   string            `json:"plugin,omitempty"`
	Realtime bool              `json:"realtime"`
	Schema   json.RawMessage   `json:"schema,omitempty"`
}

type Fee struct {
	CurrencyKey string `json:"currencyKey"`
	Amount      int64  `json:"amount"`
}

type Room struct {
	Key          string            `json:"key"`
	Name         map[string]string `json:"name,omitempty"`
	EntryFee     *Fee              `json:"entryFee,omitempty"`
	PrizeID      string            `json:"prizeId,omitempty"`
	PrizeGroupID string            `json:"prizeGroupId,omitempty"`
	MinLevel     int               `json:"minLevel,omitempty"`
	Requirements *Requirements     `json:"requirements,omitempty"`

	Players        int    `json:"players,omitempty"`
	HumanSeats     int    `json:"humanSeats,omitempty"`
	BotFillSeconds int    `json:"botFillSeconds,omitempty"`
	BotTemplate    string `json:"botTemplate,omitempty"`
	BotsPay        bool   `json:"botsPay,omitempty"`
	Payout         string `json:"payout,omitempty"`
	PayoutPercent  int    `json:"payoutPercent,omitempty"`
	LoserPrizeID   string `json:"loserPrizeId,omitempty"`
	WinnerXP       int    `json:"winnerXp,omitempty"`
	LoserXP        int    `json:"loserXp,omitempty"`
	Overdraft      bool   `json:"overdraft,omitempty"`
	FeeAt          string `json:"feeAt,omitempty"`
	NewPlayersOnly bool   `json:"newPlayersOnly,omitempty"`
	Replaces       string `json:"replaces,omitempty"`
	Disabled       bool   `json:"disabled,omitempty"`
}

const (
	PayoutPrize = "prize"
	PayoutPot   = "pot"
	FeeAtSeat   = "seat"
	FeeAtStart  = "start"
)

func (r *Room) Seats(g *Game) int {
	if r != nil && r.Players > 0 {
		return r.Players
	}
	if g != nil && g.MaxPlayers > 0 {
		return g.MaxPlayers
	}
	return 2
}

func (r *Room) Humans(g *Game) int {
	seats := r.Seats(g)
	if r == nil || r.HumanSeats <= 0 || r.HumanSeats > seats {
		return seats
	}
	return r.HumanSeats
}

func (r *Room) ChargesAtStart() bool {
	return r != nil && r.FeeAt == FeeAtStart
}

func (r *Room) PotMode() bool {
	return r != nil && r.Payout == PayoutPot
}

// Entry paths. A requirement set applies to the paths it lists (none listed = every path).
const (
	PathMatchmaking = "matchmaking"
	PathBot         = "bot"
	PathFriend      = "friend"
)

// Paths lists the entry paths in their display order.
var Paths = []string{PathMatchmaking, PathBot, PathFriend}

// ValidPath reports whether p is a known entry path ("" means matchmaking).
func ValidPath(p string) bool {
	return p == "" || p == PathMatchmaking || p == PathBot || p == PathFriend
}

// MetaCondition compares a player meta with a value; Consume > 0 takes that amount on entry (a ticket meta).
type MetaCondition struct {
	MetaID   string  `json:"metaId"`
	Operator string  `json:"operator"`
	Value    float64 `json:"value"`
	Consume  float64 `json:"consume,omitempty"`
}

// Requirements are what a player needs to enter a game or a room. IDs are economy item, player tag,
// meta and product ids. Modes: ItemsMode/TagsMode "any" (default) or "all"; Refund "return" (default:
// consumed tickets come back when the player leaves before the start or the match is cancelled) or
// "keep"; Lock "show" (default: the client sees the locked room and why) or "hide"; Inherit (rooms
// only, default true) also applies the game's own requirements.
type Requirements struct {
	ItemIDs       []string          `json:"itemIds,omitempty"`
	ItemsMode     string            `json:"itemsMode,omitempty"`
	ConsumeItems  bool              `json:"consumeItems,omitempty"`
	TagIDs        []string          `json:"tagIds,omitempty"`
	TagsMode      string            `json:"tagsMode,omitempty"`
	ExcludeTagIDs []string          `json:"excludeTagIds,omitempty"`
	MinLevel      int               `json:"minLevel,omitempty"`
	Metas         []MetaCondition   `json:"metas,omitempty"`
	ProductIDs    []string          `json:"productIds,omitempty"`
	Paths         []string          `json:"paths,omitempty"`
	Refund        string            `json:"refund,omitempty"`
	Lock          string            `json:"lock,omitempty"`
	Inherit       *bool             `json:"inherit,omitempty"`
	Message       map[string]string `json:"message,omitempty"`
}

// Empty reports whether the set asks for nothing.
func (r *Requirements) Empty() bool {
	return r == nil || (len(r.ItemIDs) == 0 && len(r.TagIDs) == 0 && len(r.ExcludeTagIDs) == 0 && r.MinLevel <= 0 && len(r.Metas) == 0)
}

// AppliesTo reports whether the set is checked on this entry path.
func (r *Requirements) AppliesTo(path string) bool {
	if r.Empty() {
		return false
	}
	if path == "" {
		path = PathMatchmaking
	}
	if len(r.Paths) == 0 {
		return true
	}
	for _, p := range r.Paths {
		if p == path {
			return true
		}
	}
	return false
}

// Inherits reports whether a room set also applies the game's set.
func (r *Requirements) Inherits() bool {
	return r == nil || r.Inherit == nil || *r.Inherit
}

// Consumes reports whether entering takes anything from the player.
func (r *Requirements) Consumes() bool {
	if r.Empty() {
		return false
	}
	if r.ConsumeItems && len(r.ItemIDs) > 0 {
		return true
	}
	for _, m := range r.Metas {
		if m.Consume > 0 {
			return true
		}
	}
	return false
}

type Game struct {
	ID            string            `json:"id"`
	Key           string            `json:"key"`
	Active        bool              `json:"isActive"`
	Name          map[string]string `json:"name,omitempty"`
	MinPlayers    int               `json:"minPlayers"`
	MaxPlayers    int               `json:"maxPlayers"`
	TurnSeconds   int               `json:"turnSeconds,omitempty"`
	MatchSeconds  int               `json:"matchSeconds,omitempty"`
	BotsAllowed   bool              `json:"botsAllowed"`
	LeaderboardID string            `json:"leaderboardId,omitempty"`
	OneMatch      bool              `json:"oneMatchPerPlayer"`
	Rooms         []Room            `json:"rooms"`
	Requirements  *Requirements     `json:"requirements,omitempty"`
	Settings      map[string]any    `json:"settings"`
	Version       int64             `json:"version"`
}

func (g *Game) Room(key string) (*Room, bool) {
	if g == nil {
		return nil, false
	}
	for i := range g.Rooms {
		if g.Rooms[i].Key == key {
			return &g.Rooms[i], true
		}
	}
	return nil, false
}

func (g *Game) Int(name string, fallback int) int {
	if g == nil {
		return fallback
	}
	switch v := g.Settings[name].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		if n, err := v.Int64(); err == nil {
			return int(n)
		}
	}
	return fallback
}

func (g *Game) Bool(name string, fallback bool) bool {
	if g == nil {
		return fallback
	}
	if v, ok := g.Settings[name].(bool); ok {
		return v
	}
	return fallback
}

func (g *Game) String(name, fallback string) string {
	if g == nil {
		return fallback
	}
	if v, ok := g.Settings[name].(string); ok {
		return v
	}
	return fallback
}

// Entry is one player entering a match. Path is matchmaking (default), bot or friend; the game or
// room requirements that list other paths are not checked.
type Entry struct {
	MatchID string
	Game    string
	Room    string
	UserID  string
	Path    string
	Bot     bool
}

type EntryResult struct {
	Status   string         `json:"status"`
	Charged  bool           `json:"charged"`
	Fee      *Fee           `json:"fee,omitempty"`
	Players  int            `json:"players"`
	Rejoined bool           `json:"rejoined"`
	Tickets  map[string]any `json:"tickets,omitempty"`
}

type Payout struct {
	UserID   string           `json:"userId"`
	Status   string           `json:"status,omitempty"`
	Rank     int              `json:"rank,omitempty"`
	PrizeID  string           `json:"prizeId,omitempty"`
	Amount   *Fee             `json:"amount,omitempty"`
	Refunded *Fee             `json:"refunded,omitempty"`
	Rewards  []map[string]any `json:"rewards"`
	Error    string           `json:"error,omitempty"`
}

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Code + ": " + e.Message
	}
	return e.Code
}

func Code(err error) string {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

type Provider interface {
	Game(ctx context.Context, key string) (*Game, error)
	Enter(ctx context.Context, e Entry) (EntryResult, error)
	Leave(ctx context.Context, matchID, userID string) (bool, error)
	Start(ctx context.Context, matchID string) error
	Cancel(ctx context.Context, matchID string) (int, error)
	Finish(ctx context.Context, matchID string, winners []string) ([]Payout, error)
	Settle(ctx context.Context, st Settlement) ([]Payout, error)
	Void(ctx context.Context, matchID, reason, by string) (int, error)
	Release(ctx context.Context, matchID, userID string) error
	Active(ctx context.Context, userID string) (*ActiveMatch, error)
	Bots(ctx context.Context, q BotQuery) ([]Bot, error)
	Log(ctx context.Context, matchID string, actions []Action) error
}

// Checker is implemented by providers that can tell whether a player may enter a room without
// entering (bot and friend room RPCs call Check before they create a match).
type Checker interface {
	Check(ctx context.Context, e Entry) error
}

var (
	ErrUnavailable = errors.New("game settings are not available: the economy plugin is not loaded")
	ErrNotFound    = &Error{Code: "GAME_NOT_FOUND", Message: "this game is not configured in the console"}
	ErrInvalidKey  = errors.New("a game key is 2 to 64 lower-case letters, digits or underscores and starts with a letter")
	keyPattern     = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)
	camelPattern   = regexp.MustCompile(`^[a-z][a-zA-Z0-9_]{0,63}$`)
	mu             sync.RWMutex
	defs           = map[string]Definition{}
	provider       Provider
)

func ValidKey(key string) bool {
	return keyPattern.MatchString(key)
}

func Register(d Definition) error {
	d.Key = strings.TrimSpace(d.Key)
	if !ValidKey(d.Key) {
		return fmt.Errorf("game %q: %w", d.Key, ErrInvalidKey)
	}
	if len(d.Schema) > 0 {
		if _, err := ParseSchema(d.Schema); err != nil {
			return fmt.Errorf("game %s schema: %w", d.Key, err)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	defs[d.Key] = d
	return nil
}

func Unregister(key string) {
	mu.Lock()
	defer mu.Unlock()
	delete(defs, key)
}

func Registered() []Definition {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Definition, 0, len(defs))
	for _, d := range defs {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func Lookup(key string) (Definition, bool) {
	mu.RLock()
	defer mu.RUnlock()
	d, ok := defs[key]
	return d, ok
}

func SetProvider(p Provider) {
	mu.Lock()
	defer mu.Unlock()
	provider = p
}

func current() (Provider, error) {
	mu.RLock()
	defer mu.RUnlock()
	if provider == nil {
		return nil, ErrUnavailable
	}
	return provider, nil
}

func Settings(ctx context.Context, key string) (*Game, error) {
	p, err := current()
	if err != nil {
		return nil, err
	}
	return p.Game(ctx, key)
}

// Check answers nil when the player meets the game and room requirements for e.Path, or an *Error
// (REQUIREMENTS_NOT_MET, LEVEL_TOO_LOW, ...) with the missing requirements and products in Meta.
// MatchID is not needed.
func Check(ctx context.Context, e Entry) error {
	p, err := current()
	if err != nil {
		return err
	}
	c, ok := p.(Checker)
	if !ok {
		return nil
	}
	return c.Check(ctx, e)
}

func Enter(ctx context.Context, e Entry) (EntryResult, error) {
	p, err := current()
	if err != nil {
		return EntryResult{}, err
	}
	return p.Enter(ctx, e)
}

func Leave(ctx context.Context, matchID, userID string) (bool, error) {
	p, err := current()
	if err != nil {
		return false, err
	}
	return p.Leave(ctx, matchID, userID)
}

func Start(ctx context.Context, matchID string) error {
	p, err := current()
	if err != nil {
		return err
	}
	return p.Start(ctx, matchID)
}

func Cancel(ctx context.Context, matchID string) (int, error) {
	p, err := current()
	if err != nil {
		return 0, err
	}
	return p.Cancel(ctx, matchID)
}

func Finish(ctx context.Context, matchID string, winners []string) ([]Payout, error) {
	p, err := current()
	if err != nil {
		return nil, err
	}
	return p.Finish(ctx, matchID, winners)
}
