package commpacks

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	CodeDisabled       = "COMMS_DISABLED"
	CodeChannelClosed  = "CHANNEL_CLOSED"
	CodeSendNotAllowed = "SEND_NOT_ALLOWED"
	CodePackNotFound   = "PACK_NOT_AVAILABLE"
	CodeItemNotFound   = "ITEM_NOT_AVAILABLE"
	CodePackNotOwned   = "PACK_NOT_OWNED"
	CodeCooldown       = "COOLDOWN_ACTIVE"
	CodeSpamLocked     = "SPAM_LOCKED"
	CodeInvalid        = "INVALID_COMM_MESSAGE"

	dedupeFor = time.Minute
)

var keyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_.-]{0,99}$`)

type Settings struct {
	Enabled            bool `json:"enabled"`
	MinIntervalSeconds int  `json:"minIntervalSeconds"`
	MaxCount           int  `json:"maxCount"`
	WindowSeconds      int  `json:"windowSeconds"`
	LockSeconds        int  `json:"lockSeconds"`
	ChannelHours       int  `json:"channelHours"`
	MaxFavorites       int  `json:"maxFavorites"`
}

func Defaults() Settings {
	return Settings{Enabled: true, MinIntervalSeconds: 4, MaxCount: 6, WindowSeconds: 30, LockSeconds: 30, ChannelHours: 2, MaxFavorites: 0}
}

func (s Settings) Check() (string, string) {
	switch {
	case s.MinIntervalSeconds < 0 || s.MinIntervalSeconds > 60:
		return "minIntervalSeconds", "between 0 and 60"
	case s.MaxCount < 1 || s.MaxCount > 1000:
		return "maxCount", "between 1 and 1000"
	case s.WindowSeconds < 1 || s.WindowSeconds > 600:
		return "windowSeconds", "between 1 and 600"
	case s.LockSeconds < 1 || s.LockSeconds > 3600:
		return "lockSeconds", "between 1 and 3600"
	case s.ChannelHours < 1 || s.ChannelHours > 48:
		return "channelHours", "between 1 and 48"
	case s.MaxFavorites < 0 || s.MaxFavorites > 1000:
		return "maxFavorites", "between 0 and 1000"
	case s.MaxCount*s.MinIntervalSeconds >= s.WindowSeconds:
		return "windowSeconds", "must be longer than maxCount × minIntervalSeconds"
	}
	return "", ""
}

type Queryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func Load(ctx context.Context, q Queryer) (Settings, error) {
	st := Defaults()
	var raw string
	err := q.QueryRowContext(ctx, `SELECT settings::text FROM base_comm_settings WHERE id=1`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return st, nil
	}
	if err != nil {
		return st, err
	}
	if json.Unmarshal([]byte(raw), &st) != nil {
		return Defaults(), nil
	}
	if f, _ := st.Check(); f != "" {
		return Defaults(), nil
	}
	return st, nil
}

func (s Settings) Rule() map[string]int {
	return map[string]int{"minIntervalSeconds": s.MinIntervalSeconds, "maxCount": s.MaxCount, "windowSeconds": s.WindowSeconds, "lockSeconds": s.LockSeconds}
}

type Input struct {
	Pack            string `json:"pack"`
	Item            string `json:"item"`
	ClientMessageID string `json:"clientMessageId"`
}

type Message struct {
	MatchID         string            `json:"matchId"`
	MessageID       string            `json:"messageId"`
	SenderID        string            `json:"senderId"`
	Pack            string            `json:"pack"`
	PackKind        string            `json:"packKind"`
	Item            string            `json:"item"`
	Text            map[string]string `json:"text,omitempty"`
	AssetID         string            `json:"assetId,omitempty"`
	ImageURL        string            `json:"imageUrl,omitempty"`
	SentAt          time.Time         `json:"sentAt"`
	ClientMessageID string            `json:"clientMessageId,omitempty"`
}

type Refusal struct {
	Code         string `json:"code"`
	RetryAfterMs int64  `json:"retryAfterMs,omitempty"`
	Pack         string `json:"pack,omitempty"`
	Item         string `json:"item,omitempty"`
}

func (r *Refusal) Error() string { return r.Code }

type flow struct {
	lockUntil time.Time
	gapUntil  time.Time
	window    []time.Time
	seen      time.Time
}

type dedupe struct {
	msg Message
	at  time.Time
}

var (
	mu    sync.Mutex
	flows = map[string]*flow{}
	dups  = map[string]dedupe{}
	now   = time.Now
)

func take(key string, st Settings, at time.Time) *Refusal {
	f := flows[key]
	if f == nil {
		f = &flow{}
		flows[key] = f
	}
	f.seen = at
	if at.Before(f.lockUntil) {
		return &Refusal{Code: CodeSpamLocked, RetryAfterMs: ceilMs(f.lockUntil.Sub(at))}
	}
	if at.Before(f.gapUntil) {
		return &Refusal{Code: CodeCooldown, RetryAfterMs: ceilMs(f.gapUntil.Sub(at))}
	}
	cut := at.Add(-time.Duration(st.WindowSeconds) * time.Second)
	kept := f.window[:0]
	for _, t := range f.window {
		if t.After(cut) {
			kept = append(kept, t)
		}
	}
	f.window = kept
	if len(f.window) >= st.MaxCount {
		f.lockUntil = at.Add(time.Duration(st.LockSeconds) * time.Second)
		f.window = nil
		return &Refusal{Code: CodeSpamLocked, RetryAfterMs: ceilMs(f.lockUntil.Sub(at))}
	}
	f.window = append(f.window, at)
	f.gapUntil = at.Add(time.Duration(st.MinIntervalSeconds) * time.Second)
	return nil
}

func ceilMs(d time.Duration) int64 {
	ms := int64(d / time.Millisecond)
	if d%time.Millisecond > 0 {
		ms++
	}
	if ms < 1 {
		ms = 1
	}
	return ms
}

func sweep(at time.Time) {
	for k, f := range flows {
		if at.Sub(f.seen) > 2*time.Hour {
			delete(flows, k)
		}
	}
	for k, d := range dups {
		if at.Sub(d.at) > dedupeFor {
			delete(dups, k)
		}
	}
}

type DB interface {
	Queryer
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func Open(ctx context.Context, db DB, matchID string, userIDs []string) error {
	st, err := Load(ctx, db)
	if err != nil {
		return err
	}
	if err := ensureChannel(ctx, db, matchID, st); err != nil {
		return err
	}
	for _, id := range userIDs {
		if _, err := db.ExecContext(ctx, `INSERT INTO base_comm_member (match_id, user_id) VALUES ($1,$2::uuid) ON CONFLICT DO NOTHING`, matchID, id); err != nil {
			return err
		}
	}
	return nil
}

func ensureChannel(ctx context.Context, db DB, matchID string, st Settings) error {
	_, err := db.ExecContext(ctx, `INSERT INTO base_comm_channel (match_id, expires_at) VALUES ($1, now() + $2::INT * INTERVAL '1 hour') ON CONFLICT (match_id) DO NOTHING`, matchID, st.ChannelHours)
	return err
}

func Close(ctx context.Context, db DB, matchID, reason string) error {
	mu.Lock()
	prefix := matchID + "|"
	for k := range flows {
		if strings.HasPrefix(k, prefix) {
			delete(flows, k)
		}
	}
	for k := range dups {
		if strings.HasPrefix(k, prefix) {
			delete(dups, k)
		}
	}
	mu.Unlock()
	if reason == "" {
		reason = "ended"
	}
	_, err := db.ExecContext(ctx, `UPDATE base_comm_channel SET status='closed', close_reason=$2, closed_by='match', closed_at=now() WHERE match_id=$1 AND status='open'`, matchID, reason)
	return err
}

func newID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func Send(ctx context.Context, db DB, matchID, userID string, in Input) (*Message, error) {
	in.Pack, in.Item, in.ClientMessageID = strings.TrimSpace(in.Pack), strings.TrimSpace(in.Item), strings.TrimSpace(in.ClientMessageID)
	if matchID == "" || userID == "" || !keyPattern.MatchString(in.Pack) || !keyPattern.MatchString(in.Item) || len(in.ClientMessageID) > 64 {
		return nil, &Refusal{Code: CodeInvalid}
	}
	st, err := Load(ctx, db)
	if err != nil {
		return nil, err
	}
	if !st.Enabled {
		return nil, &Refusal{Code: CodeDisabled}
	}
	key := matchID + "|" + userID
	if in.ClientMessageID != "" {
		mu.Lock()
		d, ok := dups[key+"|"+in.ClientMessageID]
		mu.Unlock()
		if ok && now().Sub(d.at) < dedupeFor {
			m := d.msg
			return &m, nil
		}
	}
	if err := ensureChannel(ctx, db, matchID, st); err != nil {
		return nil, err
	}
	var status string
	var expired bool
	if err := db.QueryRowContext(ctx, `SELECT status, expires_at <= now() FROM base_comm_channel WHERE match_id=$1`, matchID).Scan(&status, &expired); err != nil {
		return nil, err
	}
	if status != "open" || expired {
		return nil, &Refusal{Code: CodeChannelClosed}
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO base_comm_member (match_id, user_id) VALUES ($1,$2::uuid) ON CONFLICT DO NOTHING`, matchID, userID); err != nil {
		return nil, err
	}
	var canSend bool
	if err := db.QueryRowContext(ctx, `SELECT can_send FROM base_comm_member WHERE match_id=$1 AND user_id=$2::uuid`, matchID, userID).Scan(&canSend); err != nil {
		return nil, err
	}
	if !canSend {
		return nil, &Refusal{Code: CodeSendNotAllowed}
	}
	msg := Message{MatchID: matchID, SenderID: userID, Pack: in.Pack, Item: in.Item, ClientMessageID: in.ClientMessageID}
	var access, itemKey string
	err = db.QueryRowContext(ctx, `SELECT kind, access, item_key FROM base_sticker_pack WHERE key=$1 AND is_active AND in_match`, in.Pack).Scan(&msg.PackKind, &access, &itemKey)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &Refusal{Code: CodePackNotFound, Pack: in.Pack}
	}
	if err != nil {
		return nil, err
	}
	var text string
	err = db.QueryRowContext(ctx, `SELECT s.text::text, s.asset_id, s.image_url FROM base_sticker s JOIN base_sticker_pack p ON p.id=s.pack_id WHERE p.key=$1 AND s.key=$2 AND s.is_active`, in.Pack, in.Item).Scan(&text, &msg.AssetID, &msg.ImageURL)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &Refusal{Code: CodeItemNotFound, Pack: in.Pack, Item: in.Item}
	}
	if err != nil {
		return nil, err
	}
	if access == "item" {
		var owns bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_user_item ui JOIN economy_item i ON i.id=ui.item_id WHERE ui.user_id=$1::uuid AND i.key=$2)`, userID, itemKey).Scan(&owns); err != nil {
			return nil, err
		}
		if !owns {
			return nil, &Refusal{Code: CodePackNotOwned, Pack: in.Pack}
		}
	}
	_ = json.Unmarshal([]byte(text), &msg.Text)
	if len(msg.Text) == 0 {
		msg.Text = nil
	}
	at := now()
	mu.Lock()
	sweep(at)
	refusal := take(key, st, at)
	mu.Unlock()
	if refusal != nil {
		return nil, refusal
	}
	msg.MessageID = newID()
	msg.SentAt = at.UTC()
	if in.ClientMessageID != "" {
		mu.Lock()
		dups[key+"|"+in.ClientMessageID] = dedupe{msg: msg, at: at}
		mu.Unlock()
	}
	if _, err := db.ExecContext(ctx, `UPDATE base_comm_member SET sent=sent+1 WHERE match_id=$1 AND user_id=$2::uuid`, matchID, userID); err != nil {
		return nil, err
	}
	if _, err := db.ExecContext(ctx, `UPDATE base_comm_channel SET messages=messages+1, last_at=now() WHERE match_id=$1`, matchID); err != nil {
		return nil, err
	}
	return &msg, nil
}

func AsRefusal(err error) (*Refusal, bool) {
	var r *Refusal
	if errors.As(err, &r) {
		return r, true
	}
	return nil, false
}
