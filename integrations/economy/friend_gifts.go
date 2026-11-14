package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/heroiclabs/nakama/v3/internal/chatfilter"
)

// Friend gifts: a player sends a friend an item or currency from their own
// inventory, or buys a shop product for the friend. Pure rules and the admin
// settings live in this file; friend_gifts_engine.go holds the database side.
// "Gift codes" (economy_gift, the «هدیه‌ها» page) are a different feature.

const (
	giftSentEvent     = "Gift:Sent"
	giftReceivedEvent = "Gift:Received"
	giftNotifyCode    = 1310
)

type giftItemType struct {
	TypeID   string `json:"typeId"`
	Giftable bool   `json:"giftable"`
}

type giftCurrency struct {
	CurrencyKey string `json:"currencyKey"`
	Min         int64  `json:"min"`
	Max         int64  `json:"max"`
}

type friendGiftSettings struct {
	Enabled                 bool           `json:"enabled"`
	Recipients              string         `json:"recipients"`
	MinFriendDays           int            `json:"minFriendDays"`
	MinMutualFriends        int            `json:"minMutualFriends"`
	MinSenderLevel          int            `json:"minSenderLevel"`
	LevelTableID            string         `json:"levelTableId"`
	MinSenderAccountDays    int            `json:"minSenderAccountDays"`
	MinRecipientAccountDays int            `json:"minRecipientAccountDays"`
	InventoryEnabled        bool           `json:"inventoryEnabled"`
	ItemsDefault            bool           `json:"itemsDefault"`
	ItemTypes               []giftItemType `json:"itemTypes"`
	Equipped                string         `json:"equipped"`
	UpgradeLevel            string         `json:"upgradeLevel"`
	Regift                  string         `json:"regift"`
	ItemRequirements        string         `json:"itemRequirements"`
	CurrencyEnabled         bool           `json:"currencyEnabled"`
	Currencies              []giftCurrency `json:"currencies"`
	ShopEnabled             bool           `json:"shopEnabled"`
	ProductsDefault         bool           `json:"productsDefault"`
	Methods                 []string       `json:"methods"`
	StackableProducts       string         `json:"stackableProducts"`
	PartialOwned            string         `json:"partialOwned"`
	ProductRules            string         `json:"productRules"`
	DiscountCodes           bool           `json:"discountCodes"`
	PerSenderDaily          int            `json:"perSenderDaily"`
	PerRecipientDaily       int            `json:"perRecipientDaily"`
	PerPairDaily            int            `json:"perPairDaily"`
	MatchPolicy             string         `json:"matchPolicy"`
	SendBonusMeta           string         `json:"sendBonusMeta"`
	ReceiveBonusMeta        string         `json:"receiveBonusMeta"`
	Delivery                string         `json:"delivery"`
	ExpiryHours             int            `json:"expiryHours"`
	OnExpiry                string         `json:"onExpiry"`
	CoinReturn              string         `json:"coinReturn"`
	MoneyReturn             string         `json:"moneyReturn"`
	SenderCancel            bool           `json:"senderCancel"`
	MessageEnabled          bool           `json:"messageEnabled"`
	MessageMaxLength        int            `json:"messageMaxLength"`
	MessageFilter           string         `json:"messageFilter"`
	NotifyChannel           string         `json:"notifyChannel"`
	SocketNotify            bool           `json:"socketNotify"`
	NotifySender            bool           `json:"notifySender"`
	UpdatedAt               string         `json:"updatedAt,omitempty"`
}

var giftMethodKinds = []string{"in_app", "gateway", "store"}

func defaultFriendGiftSettings() friendGiftSettings {
	return friendGiftSettings{
		Enabled: true, Recipients: "friends",
		InventoryEnabled: true, ItemsDefault: true, ItemTypes: []giftItemType{}, Equipped: "block", UpgradeLevel: "keep", Regift: "block", ItemRequirements: "check",
		Currencies:  []giftCurrency{},
		ShopEnabled: true, ProductsDefault: true, Methods: append([]string(nil), giftMethodKinds...), StackableProducts: "allow", PartialOwned: "block", ProductRules: "recipient",
		PerSenderDaily: 5, PerRecipientDaily: 10, PerPairDaily: 2, MatchPolicy: "priority",
		Delivery: "instant", ExpiryHours: 72, OnExpiry: "return", CoinReturn: "refund", MoneyReturn: "sender", SenderCancel: true,
		MessageEnabled: true, MessageMaxLength: 120, MessageFilter: "rules",
		NotifyChannel: "in_app", SocketNotify: true, NotifySender: true,
	}
}

func loadFriendGiftSettings(ctx context.Context, q queryer) (friendGiftSettings, error) {
	st := defaultFriendGiftSettings()
	var raw string
	var updated time.Time
	err := q.QueryRowContext(ctx, `SELECT settings::text, updated_at FROM economy_friend_gift_settings WHERE id=1`).Scan(&raw, &updated)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return st, err
	}
	if err == nil {
		_ = json.Unmarshal([]byte(raw), &st)
		st.UpdatedAt = updated.UTC().Format(time.RFC3339)
	}
	if st.ItemTypes == nil {
		st.ItemTypes = []giftItemType{}
	}
	if st.Currencies == nil {
		st.Currencies = []giftCurrency{}
	}
	if st.Methods == nil {
		st.Methods = []string{}
	}
	return st, nil
}

// methodKind groups a payment method into the kinds an admin allows for
// friend gifts.
func methodKind(m string) string {
	switch {
	case m == "in_app":
		return "in_app"
	case isStore(m):
		return "store"
	case isGateway(m):
		return "gateway"
	}
	return ""
}

func (st friendGiftSettings) methodAllowed(m string) bool {
	k := methodKind(m)
	for _, x := range st.Methods {
		if x == k {
			return true
		}
	}
	return false
}

// itemGiftable resolves an item's own flag, then its type default, then the
// project default.
func (st friendGiftSettings) itemGiftable(own *bool, typeID string) bool {
	if own != nil {
		return *own
	}
	for _, t := range st.ItemTypes {
		if t.TypeID == typeID {
			return t.Giftable
		}
	}
	return st.ItemsDefault
}

func (st friendGiftSettings) productGiftable(own *bool) bool {
	if own != nil {
		return *own
	}
	return st.ProductsDefault
}

func (st friendGiftSettings) currency(key string) (giftCurrency, bool) {
	if !st.CurrencyEnabled {
		return giftCurrency{}, false
	}
	for _, c := range st.Currencies {
		if c.CurrencyKey == key {
			return c, true
		}
	}
	return giftCurrency{}, false
}

// giftLimits are the daily limits that apply to one send. 0 means no limit.
type giftLimits struct {
	Sender         int    `json:"sender"`
	Recipient      int    `json:"recipient"`
	Pair           int    `json:"pair"`
	SenderBonus    int    `json:"senderBonus"`
	RecipientBonus int    `json:"recipientBonus"`
	CanSend        bool   `json:"canSend"`
	RuleID         string `json:"ruleId,omitempty"`
	RuleName       string `json:"ruleName,omitempty"`
}

type giftRule struct {
	ID        string
	Name      string
	Sender    sql.NullInt64
	Recipient sql.NullInt64
	Pair      sql.NullInt64
	CanSend   bool
}

func generousGiftLimit(cur int, v int64) int {
	if cur == 0 || v == 0 {
		return 0
	}
	return max(cur, int(v))
}

// applyGiftRules folds the tag rules a player matches into the project
// limits. Rules come sorted by priority. "priority" takes the first rule's
// values; "generous" takes the largest of each limit (0 = no limit wins) and
// lets the player send when any rule allows it.
func applyGiftRules(base giftLimits, rules []giftRule, policy string) giftLimits {
	out := base
	out.CanSend = true
	if len(rules) == 0 {
		return out
	}
	out.RuleID, out.RuleName = rules[0].ID, rules[0].Name
	if policy != "generous" {
		r := rules[0]
		if r.Sender.Valid {
			out.Sender = int(r.Sender.Int64)
		}
		if r.Recipient.Valid {
			out.Recipient = int(r.Recipient.Int64)
		}
		if r.Pair.Valid {
			out.Pair = int(r.Pair.Int64)
		}
		out.CanSend = r.CanSend
		return out
	}
	setS, setR, setP := false, false, false
	out.CanSend = false
	for _, r := range rules {
		if r.CanSend {
			out.CanSend = true
		}
		if r.Sender.Valid {
			if !setS {
				out.Sender, setS = int(r.Sender.Int64), true
			} else {
				out.Sender = generousGiftLimit(out.Sender, r.Sender.Int64)
			}
		}
		if r.Recipient.Valid {
			if !setR {
				out.Recipient, setR = int(r.Recipient.Int64), true
			} else {
				out.Recipient = generousGiftLimit(out.Recipient, r.Recipient.Int64)
			}
		}
		if r.Pair.Valid {
			if !setP {
				out.Pair, setP = int(r.Pair.Int64), true
			} else {
				out.Pair = generousGiftLimit(out.Pair, r.Pair.Int64)
			}
		}
	}
	return out
}

// giftUsage is how many gifts already count against today's limits.
type giftUsage struct {
	Sender    int `json:"sender"`
	Recipient int `json:"recipient"`
	Pair      int `json:"pair"`
}

// limitBlock returns the code of the first daily limit the next gift would
// break, or "".
func limitBlock(l giftLimits, u giftUsage) string {
	switch {
	case l.Sender > 0 && u.Sender >= l.Sender:
		return "GIFT_SENDER_DAILY_LIMIT"
	case l.Recipient > 0 && u.Recipient >= l.Recipient:
		return "GIFT_RECIPIENT_DAILY_LIMIT"
	case l.Pair > 0 && u.Pair >= l.Pair:
		return "GIFT_PAIR_DAILY_LIMIT"
	}
	return ""
}

func remaining(limit, used int) *int {
	if limit <= 0 {
		return nil
	}
	n := limit - used
	if n < 0 {
		n = 0
	}
	return &n
}

// giftRelation is what the server knows about the two players.
type giftRelation struct {
	Mine          int // the sender's edge to the recipient (-1 none)
	Theirs        int // the recipient's edge to the sender (-1 none)
	FriendsSince  *time.Time
	Mutual        int
	SenderAge     time.Duration
	RecipientAge  time.Duration
	SenderLevel   int
	RecipientGone bool
}

// relationBlock checks who may receive a gift from whom.
func relationBlock(st friendGiftSettings, r giftRelation, now time.Time) string {
	if r.RecipientGone {
		return "RECIPIENT_NOT_FOUND"
	}
	if r.Mine == edgeBlockedState || r.Theirs == edgeBlockedState {
		return "GIFT_BLOCKED"
	}
	friends := r.Mine == 0 && r.Theirs == 0
	if st.Recipients != "any" {
		if !friends {
			return "NOT_FRIENDS"
		}
		if st.MinFriendDays > 0 && (r.FriendsSince == nil || now.Sub(*r.FriendsSince) < time.Duration(st.MinFriendDays)*24*time.Hour) {
			return "FRIENDSHIP_TOO_NEW"
		}
	}
	if st.MinMutualFriends > 0 && r.Mutual < st.MinMutualFriends {
		return "NOT_ENOUGH_MUTUAL_FRIENDS"
	}
	if st.MinSenderAccountDays > 0 && r.SenderAge < time.Duration(st.MinSenderAccountDays)*24*time.Hour {
		return "SENDER_ACCOUNT_TOO_NEW"
	}
	if st.MinRecipientAccountDays > 0 && r.RecipientAge < time.Duration(st.MinRecipientAccountDays)*24*time.Hour {
		return "RECIPIENT_ACCOUNT_TOO_NEW"
	}
	if st.MinSenderLevel > 0 && r.SenderLevel < st.MinSenderLevel {
		return "SENDER_LEVEL_TOO_LOW"
	}
	return ""
}

const edgeBlockedState = 3

// cleanGiftMessage trims and checks the optional note. The chat word list is
// reused: "rules" applies each word's own action (mask, block or flag),
// "block" refuses any hit and "off" skips the list.
func cleanGiftMessage(st friendGiftSettings, raw string, filter *chatfilter.Filter) (string, bool, error) {
	msg := strings.TrimSpace(raw)
	if msg == "" {
		return "", false, nil
	}
	if !st.MessageEnabled {
		return "", false, badRequest("GIFT_MESSAGES_OFF", "message", "gift messages are turned off")
	}
	if utf8.RuneCountInString(msg) > st.MessageMaxLength {
		return "", false, badRequest("MESSAGE_TOO_LONG", "message", "the message is too long")
	}
	if st.MessageFilter == "off" || filter == nil || filter.Len() == 0 {
		return msg, false, nil
	}
	res := filter.Apply(msg)
	if len(res.Hits) == 0 {
		return msg, false, nil
	}
	if st.MessageFilter == "block" || res.Blocked {
		return "", false, badRequest("MESSAGE_BLOCKED", "message", "the message has a word that is not allowed")
	}
	return res.Text, res.Flagged, nil
}

// productHolding is what the recipient already has of a product.
type productHolding struct {
	PrizeClaimed bool // a non-repeatable prize the recipient already got
	OneTimeHad   bool // a one-time product the recipient bought or was given
	ItemKeys     int  // item keys the prize gives
	OwnedKeys    int  // of those, the ones the recipient owns
	PendingSame  bool // a pending gift of the same product to the recipient
}

// productOwnBlock decides whether a recipient may receive a product. A product
// without items is "stackable" (currency, metas); the admin decides whether
// those can be gifted. A product whose items the recipient owns in part is
// refused unless the admin allows partial gifts.
func productOwnBlock(st friendGiftSettings, h productHolding) string {
	if h.ItemKeys == 0 && st.StackableProducts == "block" {
		return "PRODUCT_NOT_GIFTABLE"
	}
	if h.PrizeClaimed || h.OneTimeHad {
		return "RECIPIENT_OWNS_PRODUCT"
	}
	if h.ItemKeys > 0 {
		if h.OwnedKeys == h.ItemKeys {
			return "RECIPIENT_OWNS_PRODUCT"
		}
		if h.OwnedKeys > 0 && st.PartialOwned != "allow" {
			return "RECIPIENT_OWNS_PRODUCT"
		}
	}
	if h.PendingSame && h.ItemKeys > 0 {
		return "GIFT_ALREADY_PENDING"
	}
	return ""
}

// returnAction picks what happens to a declined, expired or cancelled gift.
// Items and currency always go back to the sender. A coin purchase follows
// coinReturn; a real-money purchase follows moneyReturn.
func returnAction(st friendGiftSettings, kind, method string) string {
	if kind != "product" {
		return "returned"
	}
	if method == "in_app" {
		switch st.CoinReturn {
		case "sender":
			return "sender_got"
		case "none":
			return "kept"
		}
		return "refunded"
	}
	switch st.MoneyReturn {
	case "refund_due":
		return "refund_due"
	case "none":
		return "kept"
	}
	return "sender_got"
}
