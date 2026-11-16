package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/clusterkit"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"google.golang.org/grpc/codes"
)

const (
	CodeChatQuota             = "DM_QUOTA_EXCEEDED"
	CodeChatConversationLimit = "DM_CONVERSATION_LIMIT"

	notifyChatQuota   = 1120
	chatQuotaLease    = "base.chatquota.sweep"
	chatQuotaLeaseTTL = 2 * time.Minute
	chatQuotaTick     = time.Minute
	chatQuotaMaxLimit = 100000
)

var (
	chatCountModes     = []string{"messages", "conversations", "both"}
	chatConvScopes     = []string{"day", "ever"}
	chatFriendPolicies = []string{"same", "separate", "free"}
	chatReplyModes     = []string{"off", "started", "reply"}
	chatOrders         = []string{"free_first", "packs_first"}
	chatMetaModes      = []string{"positive", "until"}
	chatNoticeModes    = []string{"none", "inbox"}
)

type chatQuotaSettings struct {
	Enabled                  bool           `json:"enabled"`
	CountMode                string         `json:"countMode"`
	DailyMessages            int            `json:"dailyMessages"`
	DailyConversations       int            `json:"dailyConversations"`
	ConversationScope        string         `json:"conversationScope"`
	FriendPolicy             string         `json:"friendPolicy"`
	FriendDailyMessages      int            `json:"friendDailyMessages"`
	FriendDailyConversations int            `json:"friendDailyConversations"`
	ReplyFree                string         `json:"replyFree"`
	CountDirect              bool           `json:"countDirect"`
	CountGroups              bool           `json:"countGroups"`
	CountRooms               bool           `json:"countRooms"`
	ResetTime                string         `json:"resetTime"`
	TimeZone                 string         `json:"timeZone"`
	Order                    string         `json:"order"`
	PackMeta                 string         `json:"packMeta"`
	PackExpiryDays           int            `json:"packExpiryDays"`
	PackProducts             []string       `json:"packProducts"`
	BonusMeta                string         `json:"bonusMeta"`
	MatchPolicy              string         `json:"matchPolicy"`
	UnlimitedItems           []string       `json:"unlimitedItems"`
	UnlimitedMeta            string         `json:"unlimitedMeta"`
	UnlimitedMetaMode        string         `json:"unlimitedMetaMode"`
	Notice                   string         `json:"notice"`
	Message                  localized.Text `json:"message"`
	UpdatedBy                string         `json:"updatedBy,omitempty"`
	UpdatedAt                time.Time      `json:"updatedAt"`
}

func defaultChatQuotaSettings() chatQuotaSettings {
	return chatQuotaSettings{
		CountMode: "messages", CountDirect: true, DailyMessages: 20, DailyConversations: 5, ConversationScope: "day",
		FriendPolicy: "same", FriendDailyMessages: 50, FriendDailyConversations: 10, ReplyFree: "off",
		ResetTime: "00:00", TimeZone: "Asia/Tehran", Order: "free_first", MatchPolicy: "priority",
		UnlimitedMetaMode: "positive", Notice: "none", PackProducts: []string{}, UnlimitedItems: []string{}, Message: localized.Text{},
	}
}

func loadChatQuotaSettings(ctx context.Context, q queryer) (chatQuotaSettings, error) {
	st := defaultChatQuotaSettings()
	var raw string
	var by string
	var at time.Time
	err := q.QueryRowContext(ctx, `SELECT settings::text, updated_by, updated_at FROM base_chat_quota_settings WHERE id=1`).Scan(&raw, &by, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		return defaultChatQuotaSettings(), nil
	}
	st, _ = st.clean()
	st.UpdatedBy, st.UpdatedAt = by, at
	return st, nil
}

func cleanKeys(in []string, max int) ([]string, bool) {
	out := []string{}
	seen := map[string]bool{}
	for _, k := range in {
		k = strings.TrimSpace(k)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, k)
	}
	return out, len(out) <= max
}

func parseClock(v string) (int, int, bool) {
	parts := strings.Split(strings.TrimSpace(v), ":")
	if len(parts) != 2 || len(parts[0]) != 2 || len(parts[1]) != 2 {
		return 0, 0, false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, false
	}
	return h, m, true
}

func (st chatQuotaSettings) clean() (chatQuotaSettings, error) {
	d := defaultChatQuotaSettings()
	pick := func(v *string, def string) {
		*v = strings.TrimSpace(*v)
		if *v == "" {
			*v = def
		}
	}
	pick(&st.CountMode, d.CountMode)
	pick(&st.ConversationScope, d.ConversationScope)
	pick(&st.FriendPolicy, d.FriendPolicy)
	pick(&st.ReplyFree, d.ReplyFree)
	pick(&st.ResetTime, d.ResetTime)
	pick(&st.TimeZone, d.TimeZone)
	pick(&st.Order, d.Order)
	pick(&st.MatchPolicy, d.MatchPolicy)
	pick(&st.UnlimitedMetaMode, d.UnlimitedMetaMode)
	pick(&st.Notice, d.Notice)
	st.PackMeta = strings.TrimSpace(st.PackMeta)
	st.BonusMeta = strings.TrimSpace(st.BonusMeta)
	st.UnlimitedMeta = strings.TrimSpace(st.UnlimitedMeta)
	if st.Message == nil {
		st.Message = localized.Text{}
	}
	for k, v := range st.Message {
		if strings.TrimSpace(v) == "" {
			delete(st.Message, k)
		}
	}
	var okP, okI bool
	st.PackProducts, okP = cleanKeys(st.PackProducts, 20)
	st.UnlimitedItems, okI = cleanKeys(st.UnlimitedItems, 50)
	inRange := func(v int) bool { return v >= 0 && v <= chatQuotaMaxLimit }
	switch {
	case !validKeyType(chatCountModes, st.CountMode):
		return st, badRequest("INVALID_SETTING", "countMode", "messages, conversations or both")
	case !inRange(st.DailyMessages):
		return st, badRequest("INVALID_SETTING", "dailyMessages", "between 0 and 100000")
	case !inRange(st.DailyConversations):
		return st, badRequest("INVALID_SETTING", "dailyConversations", "between 0 and 100000")
	case !validKeyType(chatConvScopes, st.ConversationScope):
		return st, badRequest("INVALID_SETTING", "conversationScope", "day or ever")
	case !validKeyType(chatFriendPolicies, st.FriendPolicy):
		return st, badRequest("INVALID_SETTING", "friendPolicy", "same, separate or free")
	case !inRange(st.FriendDailyMessages):
		return st, badRequest("INVALID_SETTING", "friendDailyMessages", "between 0 and 100000")
	case !inRange(st.FriendDailyConversations):
		return st, badRequest("INVALID_SETTING", "friendDailyConversations", "between 0 and 100000")
	case !validKeyType(chatReplyModes, st.ReplyFree):
		return st, badRequest("INVALID_SETTING", "replyFree", "off, started or reply")
	case !validKeyType(chatOrders, st.Order):
		return st, badRequest("INVALID_SETTING", "order", "free_first or packs_first")
	case !validKeyType(friendMatchPolicies, st.MatchPolicy):
		return st, badRequest("INVALID_SETTING", "matchPolicy", "priority or generous")
	case !validKeyType(chatMetaModes, st.UnlimitedMetaMode):
		return st, badRequest("INVALID_SETTING", "unlimitedMetaMode", "positive or until")
	case !validKeyType(chatNoticeModes, st.Notice):
		return st, badRequest("INVALID_SETTING", "notice", "none or inbox")
	case st.PackExpiryDays < 0 || st.PackExpiryDays > 3650:
		return st, badRequest("INVALID_SETTING", "packExpiryDays", "between 0 and 3650")
	case !okP:
		return st, badRequest("TOO_MANY_PRODUCTS", "packProducts", "at most 20 products")
	case !okI:
		return st, badRequest("TOO_MANY_ITEMS", "unlimitedItems", "at most 50 items")
	}
	if _, _, ok := parseClock(st.ResetTime); !ok {
		return st, badRequest("INVALID_SETTING", "resetTime", "HH:MM")
	}
	if _, err := time.LoadLocation(st.TimeZone); err != nil || st.TimeZone == "Local" {
		return st, badRequest("INVALID_TIMEZONE", "timeZone", "unknown time zone")
	}
	for k, v := range st.Message {
		if len([]rune(v)) > 300 {
			return st, badRequest("TEXT_TOO_LONG", "message."+k, "at most 300 letters")
		}
	}
	return st, nil
}

func (st chatQuotaSettings) zone() (*time.Location, int, int) {
	loc, err := time.LoadLocation(st.TimeZone)
	if err != nil {
		loc = time.UTC
	}
	h, m, _ := parseClock(st.ResetTime)
	return loc, h, m
}

func quotaPeriod(now time.Time, zone *time.Location, hh, mm int) (string, time.Time, time.Time) {
	local := now.In(zone)
	start := time.Date(local.Year(), local.Month(), local.Day(), hh, mm, 0, 0, zone)
	if local.Before(start) {
		y, mo, d := local.AddDate(0, 0, -1).Date()
		start = time.Date(y, mo, d, hh, mm, 0, 0, zone)
	}
	y, mo, d := start.Date()
	end := time.Date(y, mo, d+1, hh, mm, 0, 0, zone)
	return start.Format("2006-01-02"), start, end
}

func (s *service) chatPeriod(st chatQuotaSettings) (string, time.Time, time.Time) {
	zone, h, m := st.zone()
	return quotaPeriod(s.now(), zone, h, m)
}

type chatTarget struct {
	Kind       string
	Key        string
	Peer       string
	Mode       uint8
	Subject    string
	Subcontext string
	Label      string
}

func parseChatChannel(id, userID string) (chatTarget, bool) {
	parts := strings.SplitN(id, ".", 4)
	if len(parts) != 4 {
		return chatTarget{}, false
	}
	t := chatTarget{Subject: parts[1], Subcontext: parts[2], Label: parts[3]}
	switch parts[0] {
	case "2":
		if parts[1] != "" || parts[2] != "" || parts[3] == "" || len(parts[3]) > 64 {
			return t, false
		}
		t.Kind, t.Mode, t.Key = "room", 2, "r:"+parts[3]
	case "3":
		if parts[2] != "" || parts[3] != "" {
			return t, false
		}
		if _, err := uuid.FromString(parts[1]); err != nil {
			return t, false
		}
		t.Kind, t.Mode, t.Key = "group", 3, "g:"+parts[1]
	case "4":
		if parts[3] != "" {
			return t, false
		}
		if _, err := uuid.FromString(parts[1]); err != nil {
			return t, false
		}
		if _, err := uuid.FromString(parts[2]); err != nil {
			return t, false
		}
		switch userID {
		case parts[1]:
			t.Peer = parts[2]
		case parts[2]:
			t.Peer = parts[1]
		default:
			return t, false
		}
		if t.Peer == userID {
			return t, false
		}
		t.Kind, t.Mode, t.Key = "dm", 4, "u:"+t.Peer
	default:
		return t, false
	}
	return t, true
}

func (st chatQuotaSettings) counts(t chatTarget) bool {
	switch t.Kind {
	case "dm":
		return st.CountDirect
	case "group":
		return st.CountGroups
	case "room":
		return st.CountRooms
	}
	return false
}

type chatLimits struct {
	Unlimited           bool
	UnlimitedBy         string
	Messages            int
	Conversations       int
	FriendMessages      int
	FriendConversations int
	Bonus               int
	RuleID              string
	RuleName            string
}

type chatRuleRow struct {
	id, name            string
	unlimited           bool
	msgs, convs, fm, fc sql.NullInt64
}

func mergeChatRules(st chatQuotaSettings, rules []chatRuleRow) chatLimits {
	out := chatLimits{Messages: st.DailyMessages, Conversations: st.DailyConversations, FriendMessages: st.FriendDailyMessages, FriendConversations: st.FriendDailyConversations}
	if len(rules) == 0 {
		return out
	}
	out.RuleID, out.RuleName = rules[0].id, rules[0].name
	if st.MatchPolicy != "generous" {
		r := rules[0]
		if r.unlimited {
			out.Unlimited, out.UnlimitedBy = true, "rule"
		}
		for _, p := range []struct {
			v   sql.NullInt64
			dst *int
		}{{r.msgs, &out.Messages}, {r.convs, &out.Conversations}, {r.fm, &out.FriendMessages}, {r.fc, &out.FriendConversations}} {
			if p.v.Valid {
				*p.dst = int(p.v.Int64)
			}
		}
		return out
	}
	set := [4]bool{}
	for _, r := range rules {
		if r.unlimited && !out.Unlimited {
			out.Unlimited, out.UnlimitedBy = true, "rule"
			out.RuleID, out.RuleName = r.id, r.name
		}
		for i, p := range []struct {
			v   sql.NullInt64
			dst *int
		}{{r.msgs, &out.Messages}, {r.convs, &out.Conversations}, {r.fm, &out.FriendMessages}, {r.fc, &out.FriendConversations}} {
			if !p.v.Valid {
				continue
			}
			if !set[i] {
				*p.dst, set[i] = int(p.v.Int64), true
			} else {
				*p.dst = max(*p.dst, int(p.v.Int64))
			}
		}
	}
	return out
}

func (s *service) metaValue(ctx context.Context, q queryer, userID, key string) (float64, bool, error) {
	if key == "" {
		return 0, false, nil
	}
	var v sql.NullFloat64
	err := q.QueryRowContext(ctx, `SELECT COALESCE(um.value, m.default_value)::float8 FROM economy_meta m LEFT JOIN economy_user_meta um ON um.meta_id=m.id AND um.user_id=$1::uuid WHERE m.key=$2`, userID, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !v.Valid) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return v.Float64, true, nil
}

func (s *service) effectiveChatLimits(ctx context.Context, st chatQuotaSettings, userID string) (chatLimits, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT r.id::text, r.name, r.unlimited, r.daily_messages, r.daily_conversations, r.friend_daily_messages, r.friend_daily_conversations FROM base_chat_quota_rule r
		WHERE r.is_active AND EXISTS (SELECT 1 FROM base_user_tag ut WHERE ut.user_id=$1::uuid AND ut.tag_id::text = ANY(r.tag_ids) AND (ut.expires_at IS NULL OR ut.expires_at > now()))
		ORDER BY r.priority DESC, r.created_at, r.id`, userID)
	if err != nil {
		return chatLimits{}, err
	}
	rules := []chatRuleRow{}
	for rows.Next() {
		var r chatRuleRow
		if err := rows.Scan(&r.id, &r.name, &r.unlimited, &r.msgs, &r.convs, &r.fm, &r.fc); err != nil {
			rows.Close()
			return chatLimits{}, err
		}
		rules = append(rules, r)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return chatLimits{}, err
	}
	out := mergeChatRules(st, rules)
	if !out.Unlimited && len(st.UnlimitedItems) > 0 {
		var owns bool
		if err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_user_item ui JOIN economy_item i ON i.id=ui.item_id WHERE ui.user_id=$1::uuid AND i.key = ANY($2::text[]))`, userID, st.UnlimitedItems).Scan(&owns); err != nil {
			return out, err
		}
		if owns {
			out.Unlimited, out.UnlimitedBy = true, "item"
		}
	}
	if !out.Unlimited && st.UnlimitedMeta != "" {
		v, ok, err := s.metaValue(ctx, s.db, userID, st.UnlimitedMeta)
		if err != nil {
			return out, err
		}
		if ok && metaUnlimited(st.UnlimitedMetaMode, v, s.now()) {
			out.Unlimited, out.UnlimitedBy = true, "meta"
		}
	}
	if st.BonusMeta != "" {
		v, ok, err := s.metaValue(ctx, s.db, userID, st.BonusMeta)
		if err != nil {
			return out, err
		}
		if ok && v > 0 && !math.IsNaN(v) {
			out.Bonus = int(math.Min(math.Floor(v), chatQuotaMaxLimit))
		}
	}
	if st.CountMode == "conversations" {
		out.Conversations += out.Bonus
		out.FriendConversations += out.Bonus
	} else {
		out.Messages += out.Bonus
		out.FriendMessages += out.Bonus
	}
	return out, nil
}

func metaUnlimited(mode string, v float64, now time.Time) bool {
	if math.IsNaN(v) {
		return false
	}
	if mode == "until" {
		return v > float64(now.Unix())
	}
	return v > 0
}

type quotaInput struct {
	Mode       string
	PacksFirst bool
	Unlimited  bool
	Free       bool
	NewConv    bool
	MsgLimit   int
	ConvLimit  int
	MsgUsed    int
	ConvUsed   int
	Packs      int
}

type quotaResult struct {
	OK       bool
	Code     string
	Unit     string
	FreeMsg  int
	FreeConv int
	Pack     int
	Reason   string
}

func decideQuota(in quotaInput) quotaResult {
	if in.Unlimited {
		return quotaResult{OK: true, Reason: "unlimited"}
	}
	if in.Free {
		return quotaResult{OK: true, Reason: "free"}
	}
	unit := func(used, limit int, kind string) quotaResult {
		left := max(limit-used, 0)
		switch {
		case in.PacksFirst && in.Packs > 0:
			return quotaResult{OK: true, Pack: 1, Reason: "pack", Unit: kind}
		case left > 0:
			r := quotaResult{OK: true, Reason: "daily", Unit: kind}
			if kind == "conversations" {
				r.FreeConv = 1
			} else {
				r.FreeMsg = 1
			}
			return r
		case in.Packs > 0:
			return quotaResult{OK: true, Pack: 1, Reason: "pack", Unit: kind}
		}
		return quotaResult{Code: CodeChatQuota, Unit: kind}
	}
	switch in.Mode {
	case "conversations":
		if !in.NewConv {
			return quotaResult{OK: true, Reason: "open"}
		}
		return unit(in.ConvUsed, in.ConvLimit, "conversations")
	case "both":
		if in.NewConv && in.ConvUsed >= in.ConvLimit {
			return quotaResult{Code: CodeChatConversationLimit, Unit: "conversations"}
		}
		r := unit(in.MsgUsed, in.MsgLimit, "messages")
		if r.OK && in.NewConv {
			r.FreeConv = 1
		}
		return r
	}
	return unit(in.MsgUsed, in.MsgLimit, "messages")
}

type chatDay struct {
	Messages, Conversations, FriendMessages, FriendConversations int
	PackUsed, FreeSent, UnlimitedSent, Total, Blocked            int
	Notified                                                     bool
}

const chatDayCols = `messages, conversations, friend_messages, friend_conversations, pack_used, free_sent, unlimited_sent, total, blocked, notified`

func (d *chatDay) scan(row interface{ Scan(...any) error }) error {
	return row.Scan(&d.Messages, &d.Conversations, &d.FriendMessages, &d.FriendConversations, &d.PackUsed, &d.FreeSent, &d.UnlimitedSent, &d.Total, &d.Blocked, &d.Notified)
}

type execQueryer interface {
	queryer
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func absorbPacks(ctx context.Context, q execQueryer, st chatQuotaSettings, userID string) error {
	if st.PackMeta == "" {
		return nil
	}
	var user any
	if userID != "" {
		user = userID
	}
	_, err := q.ExecContext(ctx, `WITH old AS (
			SELECT um.user_id, um.meta_id, um.value, um.updated_at FROM economy_user_meta um JOIN economy_meta m ON m.id=um.meta_id
			WHERE m.key=$1 AND um.value >= 1 AND ($2::uuid IS NULL OR um.user_id=$2::uuid) AND EXISTS (SELECT 1 FROM users u WHERE u.id=um.user_id)
			FOR UPDATE OF um),
		moved AS (
			UPDATE economy_user_meta um SET value = um.value - floor(old.value), updated_at = now() FROM old
			WHERE um.user_id=old.user_id AND um.meta_id=old.meta_id
			RETURNING old.user_id, LEAST(floor(old.value), 2000000000)::int AS amount, old.updated_at)
		INSERT INTO base_chat_quota_lot (id, user_id, amount, remaining, source, expires_at, created_at)
		SELECT gen_random_uuid(), user_id, amount, amount, 'meta', CASE WHEN $3::int > 0 THEN updated_at + $3::int * INTERVAL '1 day' END, updated_at FROM moved WHERE amount > 0`,
		st.PackMeta, user, st.PackExpiryDays)
	return err
}

func packsLeft(ctx context.Context, q queryer, userID string) (int, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT COALESCE(sum(remaining),0)::int FROM base_chat_quota_lot WHERE user_id=$1::uuid AND remaining > 0 AND (expires_at IS NULL OR expires_at > now())`, userID).Scan(&n)
	return n, err
}

type chatTargetRow struct {
	found       bool
	first, last time.Time
}

func loadTargetRow(ctx context.Context, q queryer, userID, target string) (chatTargetRow, error) {
	var r chatTargetRow
	err := q.QueryRowContext(ctx, `SELECT first_at, last_at FROM base_chat_quota_target WHERE user_id=$1::uuid AND target=$2`, userID, target).Scan(&r.first, &r.last)
	if errors.Is(err, sql.ErrNoRows) {
		return r, nil
	}
	r.found = err == nil
	return r, err
}

func replyIsFree(mode string, mine, theirs chatTargetRow) bool {
	if !theirs.found {
		return false
	}
	switch mode {
	case "started":
		return !mine.found || theirs.first.Before(mine.first)
	case "reply":
		return !mine.found || theirs.last.After(mine.last)
	}
	return false
}

func isNewConversation(scope string, mine chatTargetRow, start time.Time) bool {
	if !mine.found {
		return true
	}
	return scope == "day" && mine.last.Before(start)
}

type chatSpend struct {
	Result   quotaResult
	Limits   chatLimits
	Friend   bool
	Day      string
	Reset    time.Time
	Limit    int
	Used     int
	Packs    int
	Notify   bool
	Settings chatQuotaSettings
}

func (s *service) areFriends(ctx context.Context, a, b string) (bool, error) {
	var ok bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM user_edge WHERE source_id=$1::uuid AND destination_id=$2::uuid AND state=0)`, a, b).Scan(&ok)
	return ok, err
}

func (s *service) spendChatQuota(ctx context.Context, st chatQuotaSettings, userID string, t chatTarget) (chatSpend, error) {
	out := chatSpend{Settings: st}
	lim, err := s.effectiveChatLimits(ctx, st, userID)
	if err != nil {
		return out, err
	}
	out.Limits = lim
	if t.Kind == "dm" && st.FriendPolicy != "same" && !lim.Unlimited {
		if out.Friend, err = s.areFriends(ctx, userID, t.Peer); err != nil {
			return out, err
		}
	}
	day, start, reset := s.chatPeriod(st)
	out.Day, out.Reset = day, reset
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO base_chat_quota_day (user_id, day) VALUES ($1::uuid, $2::date) ON CONFLICT (user_id, day) DO NOTHING`, userID, day); err != nil {
			return err
		}
		var d chatDay
		if err := d.scan(tx.QueryRowContext(ctx, `SELECT `+chatDayCols+` FROM base_chat_quota_day WHERE user_id=$1::uuid AND day=$2::date FOR UPDATE`, userID, day)); err != nil {
			return err
		}
		mine, err := loadTargetRow(ctx, tx, userID, t.Key)
		if err != nil {
			return err
		}
		in := quotaInput{Mode: st.CountMode, PacksFirst: st.Order == "packs_first", Unlimited: lim.Unlimited, NewConv: isNewConversation(st.ConversationScope, mine, start)}
		if out.Friend && st.FriendPolicy == "free" {
			in.Free = true
		}
		if !in.Unlimited && !in.Free && t.Kind == "dm" && st.ReplyFree != "off" {
			theirs, err := loadTargetRow(ctx, tx, t.Peer, "u:"+userID)
			if err != nil {
				return err
			}
			in.Free = replyIsFree(st.ReplyFree, mine, theirs)
		}
		separate := out.Friend && st.FriendPolicy == "separate"
		if separate {
			in.MsgLimit, in.ConvLimit, in.MsgUsed, in.ConvUsed = lim.FriendMessages, lim.FriendConversations, d.FriendMessages, d.FriendConversations
		} else {
			in.MsgLimit, in.ConvLimit, in.MsgUsed, in.ConvUsed = lim.Messages, lim.Conversations, d.Messages, d.Conversations
		}
		if !in.Unlimited && !in.Free {
			if err := absorbPacks(ctx, tx, st, userID); err != nil {
				return err
			}
			if in.Packs, err = packsLeft(ctx, tx, userID); err != nil {
				return err
			}
		}
		res := decideQuota(in)
		out.Result, out.Packs = res, in.Packs
		if res.Unit == "conversations" {
			out.Limit, out.Used = in.ConvLimit, in.ConvUsed
		} else {
			out.Limit, out.Used = in.MsgLimit, in.MsgUsed
		}
		if !res.OK {
			out.Notify = st.Notice == "inbox" && !d.Notified
			_, err := tx.ExecContext(ctx, `UPDATE base_chat_quota_day SET blocked = blocked + 1, notified = notified OR $3, updated_at = now() WHERE user_id=$1::uuid AND day=$2::date`, userID, day, out.Notify)
			return err
		}
		if res.Pack > 0 {
			r, err := tx.ExecContext(ctx, `UPDATE base_chat_quota_lot SET remaining = remaining - 1, updated_at = now() WHERE id = (
				SELECT id FROM base_chat_quota_lot WHERE user_id=$1::uuid AND remaining > 0 AND (expires_at IS NULL OR expires_at > now())
				ORDER BY expires_at NULLS LAST, created_at, id LIMIT 1 FOR UPDATE)`, userID)
			if err != nil {
				return err
			}
			if n, _ := r.RowsAffected(); n == 0 {
				return errors.New("chat quota pack vanished")
			}
		}
		msgCol, convCol := "messages", "conversations"
		if separate {
			msgCol, convCol = "friend_messages", "friend_conversations"
		}
		free, unl := 0, 0
		if res.Reason == "free" {
			free = 1
		}
		if res.Reason == "unlimited" {
			unl = 1
		}
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`UPDATE base_chat_quota_day SET %s = %s + $3, %s = %s + $4, pack_used = pack_used + $5, free_sent = free_sent + $6, unlimited_sent = unlimited_sent + $7, total = total + 1, updated_at = now()
			WHERE user_id=$1::uuid AND day=$2::date`, msgCol, msgCol, convCol, convCol), userID, day, res.FreeMsg, res.FreeConv, res.Pack, free, unl); err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO base_chat_quota_target (user_id, target, first_at, last_at) VALUES ($1::uuid, $2, clock_timestamp(), clock_timestamp())
			ON CONFLICT (user_id, target) DO UPDATE SET last_at = clock_timestamp()`, userID, t.Key)
		return err
	})
	return out, err
}

type chatQuotaProduct struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

func (s *service) chatPackProducts(ctx context.Context, st chatQuotaSettings, langs *localized.Set, lang string, activeOnly bool) ([]chatQuotaProduct, error) {
	out := []chatQuotaProduct{}
	if st.PackMeta == "" && len(st.PackProducts) == 0 {
		return out, nil
	}
	rows, err := s.db.QueryContext(ctx, `SELECT p.id::text, p.key, p.name::text,
			COALESCE((SELECT sum(CASE WHEN jsonb_typeof(it->'value')='number' THEN (it->>'value')::numeric ELSE 0 END) FROM jsonb_array_elements(z.items) it
				WHERE it->>'type'='meta' AND it->>'key'=$1 AND COALESCE(it->>'action','add') <> 'set'), 0)::int
		FROM economy_product p JOIN economy_prize z ON z.id=p.prize_id
		WHERE (NOT $3 OR p.is_active) AND ((cardinality($2::text[]) > 0 AND p.key = ANY($2::text[]))
			OR (cardinality($2::text[]) = 0 AND $1 <> '' AND z.items @> jsonb_build_array(jsonb_build_object('type','meta','key',$1::text))))
		ORDER BY p.sort, p.key LIMIT 20`, st.PackMeta, st.PackProducts, activeOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p chatQuotaProduct
		var name string
		if err := rows.Scan(&p.ID, &p.Key, &name, &p.Amount); err != nil {
			return nil, err
		}
		if langs != nil {
			p.Name = langs.Resolve(localized.Parse(name), lang)
		} else {
			p.Name = localized.Parse(name)["en"]
		}
		if p.Name == "" {
			p.Name = p.Key
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

var chatQuotaMessages = map[string]localized.Text{
	CodeChatQuota: {
		"fa": "پیام‌های خصوصی امروزت تمام شد. از {reset} دوباره می‌توانی پیام بدهی، یا یک بستهٔ پیام بخر.",
		"en": "You have used up today's private messages. You can send more after {reset}, or buy a message pack.",
		"ar": "لقد استنفدت رسائلك الخاصة لهذا اليوم. يمكنك الإرسال مجدداً بعد {reset} أو شراء حزمة رسائل.",
	},
	CodeChatConversationLimit: {
		"fa": "امروز بیشتر از {limit} گفتگوی خصوصی تازه نمی‌توانی شروع کنی. از {reset} دوباره می‌توانی.",
		"en": "You can start at most {limit} new private chats a day. You can start more after {reset}.",
		"ar": "يمكنك بدء {limit} محادثة خاصة جديدة كحد أقصى في اليوم. يمكنك البدء مجدداً بعد {reset}.",
	},
}

type chatQuotaError struct {
	Code       string             `json:"code"`
	Message    string             `json:"message"`
	Lang       string             `json:"lang"`
	Unit       string             `json:"unit"`
	Limit      int                `json:"limit"`
	Used       int                `json:"used"`
	Remaining  int                `json:"remaining"`
	PacksLeft  int                `json:"packsLeft"`
	ResetsAt   string             `json:"resetsAt"`
	Friend     bool               `json:"friend"`
	Rule       map[string]any     `json:"rule"`
	Products   []chatQuotaProduct `json:"products"`
	ProductIDs []string           `json:"productIds"`
}

func (e *chatQuotaError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func (s *service) chatQuotaFailure(ctx context.Context, userID string, sp chatSpend) (*chatQuotaError, error) {
	st := sp.Settings
	e := &chatQuotaError{Code: sp.Result.Code, Unit: sp.Result.Unit, Limit: sp.Limit, Used: sp.Used, Remaining: 0, PacksLeft: sp.Packs, ResetsAt: sp.Reset.UTC().Format(time.RFC3339), Friend: sp.Friend}
	if sp.Limits.RuleID != "" {
		e.Rule = map[string]any{"id": sp.Limits.RuleID, "name": sp.Limits.RuleName}
	}
	langs, lang, err := localized.ForUser(ctx, s.db, userID, "")
	if err != nil {
		langs, lang = nil, "en"
	}
	e.Lang = lang
	if e.Products, err = s.chatPackProducts(ctx, st, langs, lang, true); err != nil {
		return nil, err
	}
	e.ProductIDs = make([]string, 0, len(e.Products))
	for _, p := range e.Products {
		e.ProductIDs = append(e.ProductIDs, p.ID)
	}
	text := ""
	if e.Code == CodeChatQuota && len(st.Message) > 0 {
		if langs != nil {
			text = langs.Resolve(st.Message, lang)
		}
		if text == "" {
			text = st.Message["en"]
		}
	}
	if text == "" {
		if langs != nil {
			text = langs.Resolve(chatQuotaMessages[e.Code], lang)
		}
		if text == "" {
			text = chatQuotaMessages[e.Code]["en"]
		}
	}
	zone, _, _ := st.zone()
	text = strings.ReplaceAll(text, "{limit}", localDigits(lang, strconv.Itoa(e.Limit)))
	text = strings.ReplaceAll(text, "{reset}", localDigits(lang, sp.Reset.In(zone).Format("2006-01-02 15:04")))
	e.Message = text
	return e, nil
}

func (s *service) chatQuotaNotice(ctx context.Context, userID string, e *chatQuotaError) {
	content := map[string]any{"type": "chat_quota", "code": e.Code, "unit": e.Unit, "resetsAt": e.ResetsAt, "productIds": e.ProductIDs, "message": e.Message}
	if err := s.nk.NotificationSend(ctx, userID, e.Message, content, notifyChatQuota, "", true); err != nil {
		s.logger.Warn("Base chat quota notice not sent: %v", err)
	}
}

func validChatContent(content string) bool {
	b := []byte(strings.TrimSpace(content))
	return len(b) > 0 && b[0] == '{' && json.Valid(b)
}

func (s *service) chatQuotaGate(ctx context.Context, userID, channelID, content string) error {
	if userID == "" {
		return nil
	}
	st, err := loadChatQuotaSettings(ctx, s.db)
	if err != nil {
		return s.hookFailure(err)
	}
	if !st.Enabled {
		return nil
	}
	t, ok := parseChatChannel(channelID, userID)
	if !ok || !st.counts(t) || !validChatContent(content) {
		return nil
	}
	if sid, _ := ctx.Value(runtime.RUNTIME_CTX_SESSION_ID).(string); sid != "" {
		meta, err := s.nk.StreamUserGet(t.Mode, t.Subject, t.Subcontext, t.Label, userID, sid)
		if err == nil && meta == nil {
			return nil
		}
	}
	sp, err := s.spendChatQuota(ctx, st, userID, t)
	if err != nil {
		return s.hookFailure(err)
	}
	if sp.Result.OK {
		return nil
	}
	e, err := s.chatQuotaFailure(ctx, userID, sp)
	if err != nil {
		return s.hookFailure(err)
	}
	if sp.Notify {
		s.chatQuotaNotice(ctx, userID, e)
	}
	return runtime.NewError(e.Error(), int(codes.ResourceExhausted))
}

type chatLot struct {
	ID        string     `json:"id"`
	Amount    int        `json:"amount"`
	Remaining int        `json:"remaining"`
	Source    string     `json:"source"`
	Note      string     `json:"note,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt"`
	CreatedBy string     `json:"createdBy,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	Active    bool       `json:"active"`
}

func (s *service) chatLots(ctx context.Context, userID string, limit int) ([]chatLot, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id::text, amount, remaining, source, note, expires_at, created_by, created_at, remaining > 0 AND (expires_at IS NULL OR expires_at > now())
		FROM base_chat_quota_lot WHERE user_id=$1::uuid ORDER BY (remaining > 0 AND (expires_at IS NULL OR expires_at > now())) DESC, expires_at NULLS LAST, created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []chatLot{}
	for rows.Next() {
		var l chatLot
		var exp sql.NullTime
		if err := rows.Scan(&l.ID, &l.Amount, &l.Remaining, &l.Source, &l.Note, &exp, &l.CreatedBy, &l.CreatedAt, &l.Active); err != nil {
			return nil, err
		}
		if exp.Valid {
			t := exp.Time.UTC()
			l.ExpiresAt = &t
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func pool(limit, used int) map[string]any {
	return map[string]any{"limit": limit, "used": used, "left": max(limit-used, 0)}
}

func (s *service) chatQuotaStatus(ctx context.Context, userID, lang string, lots int) (map[string]any, error) {
	st, err := loadChatQuotaSettings(ctx, s.db)
	if err != nil {
		return nil, err
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error { return absorbPacks(ctx, tx, st, userID) }); err != nil {
		return nil, err
	}
	lim, err := s.effectiveChatLimits(ctx, st, userID)
	if err != nil {
		return nil, err
	}
	day, _, reset := s.chatPeriod(st)
	var d chatDay
	err = d.scan(s.db.QueryRowContext(ctx, `SELECT `+chatDayCols+` FROM base_chat_quota_day WHERE user_id=$1::uuid AND day=$2::date`, userID, day))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	packs, err := packsLeft(ctx, s.db, userID)
	if err != nil {
		return nil, err
	}
	langs, resolved, err := localized.ForUser(ctx, s.db, userID, lang)
	if err != nil {
		return nil, err
	}
	products, err := s.chatPackProducts(ctx, st, langs, resolved, true)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(products))
	for _, p := range products {
		ids = append(ids, p.ID)
	}
	var nextExpiry any
	var exp sql.NullTime
	if err := s.db.QueryRowContext(ctx, `SELECT min(expires_at) FROM base_chat_quota_lot WHERE user_id=$1::uuid AND remaining > 0 AND expires_at > now()`, userID).Scan(&exp); err != nil {
		return nil, err
	}
	if exp.Valid {
		nextExpiry = exp.Time.UTC().Format(time.RFC3339)
	}
	var rule any
	if lim.RuleID != "" {
		rule = map[string]any{"id": lim.RuleID, "name": lim.RuleName}
	}
	freeLeft := max(lim.Messages-d.Messages, 0)
	if st.CountMode == "conversations" {
		freeLeft = max(lim.Conversations-d.Conversations, 0)
	}
	out := map[string]any{
		"lang":          resolved,
		"enabled":       st.Enabled,
		"countMode":     st.CountMode,
		"unlimited":     lim.Unlimited,
		"unlimitedBy":   lim.UnlimitedBy,
		"rule":          rule,
		"freeLeft":      freeLeft,
		"packsLeft":     packs,
		"packsExpireAt": nextExpiry,
		"order":         st.Order,
		"day":           day,
		"resetsAt":      reset.UTC().Format(time.RFC3339),
		"timeZone":      st.TimeZone,
		"messages":      pool(lim.Messages, d.Messages),
		"conversations": pool(lim.Conversations, d.Conversations),
		"friendPolicy":  st.FriendPolicy,
		"replyFree":     st.ReplyFree,
		"bonus":         lim.Bonus,
		"today":         map[string]any{"sent": d.Total, "fromPacks": d.PackUsed, "free": d.FreeSent, "unlimited": d.UnlimitedSent, "blocked": d.Blocked},
		"products":      products,
		"productIds":    ids,
		"counts":        map[string]bool{"direct": st.CountDirect, "groups": st.CountGroups, "rooms": st.CountRooms},
	}
	if st.FriendPolicy == "separate" {
		out["friends"] = map[string]any{"messages": pool(lim.FriendMessages, d.FriendMessages), "conversations": pool(lim.FriendConversations, d.FriendConversations)}
	}
	if lots > 0 {
		list, err := s.chatLots(ctx, userID, lots)
		if err != nil {
			return nil, err
		}
		out["lots"] = list
	}
	return out, nil
}

func (s *service) rpcChatQuota(ctx context.Context, userID, payload string) (any, error) {
	var in struct {
		Lang string `json:"lang"`
	}
	if err := decodeStrict(payload, &in); err != nil {
		return nil, err
	}
	return s.chatQuotaStatus(ctx, userID, in.Lang, 0)
}

func (s *service) chatQuotaLoop(ctx context.Context) {
	ticker := time.NewTicker(chatQuotaTick)
	defer ticker.Stop()
	holder := clusterkit.Holder(ctx)
	for {
		select {
		case <-ctx.Done():
			_ = clusterkit.ReleaseLease(context.Background(), s.db, chatQuotaLease, holder)
			return
		case <-ticker.C:
		}
		if err := s.chatQuotaSweep(ctx, holder); err != nil && ctx.Err() == nil {
			s.logger.Warn("Base chat quota sweep failed: %v", err)
		}
	}
}

func (s *service) chatQuotaSweep(ctx context.Context, holder string) error {
	ok, err := clusterkit.AcquireLease(ctx, s.db, chatQuotaLease, holder, chatQuotaLeaseTTL)
	if err != nil || !ok {
		return err
	}
	st, err := loadChatQuotaSettings(ctx, s.db)
	if err != nil {
		return err
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error { return absorbPacks(ctx, tx, st, "") }); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM base_chat_quota_day WHERE day < (now() - INTERVAL '400 days')::date`); err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `DELETE FROM base_chat_quota_lot WHERE updated_at < now() - INTERVAL '180 days' AND (remaining = 0 OR expires_at < now())`)
	return err
}
