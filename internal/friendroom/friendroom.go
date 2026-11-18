package friendroom

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	ModeMatchmaking = "matchmaking"
	ModeFriend      = "friend"
	ModeBot         = "bot"

	StatusWaiting  = "waiting"
	StatusPlaying  = "playing"
	StatusFinished = "finished"

	ActionCancel  = "cancel"
	ActionDecline = "decline"

	CodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	CodeLength   = 6
)

type Params struct {
	Game      string
	Room      string
	Mode      string
	Host      string
	Code      string
	Invited   string
	Users     []string
	ExpiresAt time.Time
}

func FromParams(in map[string]any) Params {
	str := func(k string) string { v, _ := in[k].(string); return strings.TrimSpace(v) }
	p := Params{Game: str("game"), Room: str("room"), Mode: str("mode"), Host: str("host"), Code: str("code"), Invited: str("invited")}
	if p.Mode == "" {
		switch {
		case str("friend") == "1":
			p.Mode = ModeFriend
		case str("bot") == "1":
			p.Mode = ModeBot
		default:
			p.Mode = ModeMatchmaking
		}
	}
	if users := str("users"); users != "" {
		for _, u := range strings.Split(users, ",") {
			if u = strings.TrimSpace(u); u != "" {
				p.Users = append(p.Users, u)
			}
		}
	}
	if ms, err := strconv.ParseInt(str("expiresAt"), 10, 64); err == nil && ms > 0 {
		p.ExpiresAt = time.UnixMilli(ms).UTC()
	}
	return p
}

func (p Params) Map() map[string]any {
	out := map[string]any{"game": p.Game, "room": p.Room, "mode": p.Mode}
	if p.Host != "" {
		out["host"] = p.Host
	}
	if p.Code != "" {
		out["code"] = p.Code
	}
	if p.Invited != "" {
		out["invited"] = p.Invited
	}
	if len(p.Users) > 0 {
		out["users"] = strings.Join(p.Users, ",")
	}
	if !p.ExpiresAt.IsZero() {
		out["expiresAt"] = strconv.FormatInt(p.ExpiresAt.UnixMilli(), 10)
	}
	return out
}

func (p Params) Label(status string, extra map[string]any) string {
	l := map[string]any{"game": p.Game, "room": p.Room, "mode": p.Mode, "status": status, "open": status == StatusWaiting}
	if p.Mode == ModeFriend {
		l["host"], l["code"] = p.Host, p.Code
		if p.Invited != "" {
			l["invited"] = p.Invited
		}
		if !p.ExpiresAt.IsZero() {
			l["expiresAt"] = float64(p.ExpiresAt.UnixMilli()) / 1000
		}
	}
	for k, v := range extra {
		l[k] = v
	}
	raw, _ := json.Marshal(l)
	return string(raw)
}

func (p Params) MayJoin(userID string) bool {
	switch p.Mode {
	case ModeFriend:
		return userID == p.Host || p.Invited == "" || userID == p.Invited
	case ModeBot:
		return len(p.Users) == 0 || userID == p.Users[0]
	default:
		if len(p.Users) == 0 {
			return true
		}
		for _, u := range p.Users {
			if u == userID {
				return true
			}
		}
		return false
	}
}

type Signal struct {
	Action string `json:"action"`
	UserID string `json:"userId"`
}

func ParseSignal(data string) (Signal, bool) {
	var s Signal
	if err := json.Unmarshal([]byte(data), &s); err != nil || s.Action == "" {
		return s, false
	}
	return s, true
}

func SignalData(action, userID string) string {
	raw, _ := json.Marshal(Signal{Action: action, UserID: userID})
	return string(raw)
}

func Answer(ok bool, reason string) string {
	out := map[string]any{"ok": ok}
	if reason != "" {
		out["reason"] = reason
	}
	raw, _ := json.Marshal(out)
	return string(raw)
}

func ParseAnswer(raw string) (bool, string) {
	var out struct {
		OK     bool   `json:"ok"`
		Reason string `json:"reason"`
	}
	_ = json.Unmarshal([]byte(raw), &out)
	return out.OK, out.Reason
}

func NewCode() string {
	b := make([]byte, CodeLength)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(CodeAlphabet))))
		b[i] = CodeAlphabet[n.Int64()]
	}
	return string(b)
}

func ValidCode(code string) bool {
	return len(code) == CodeLength && strings.Trim(code, CodeAlphabet) == ""
}

var recent = struct {
	sync.Mutex
	ids map[string]string
}{ids: map[string]string{}}

func Remember(game, code, host, id string) {
	recent.Lock()
	defer recent.Unlock()
	if len(recent.ids) > 20000 {
		recent.ids = map[string]string{}
	}
	recent.ids["code:"+code] = id
	recent.ids["host:"+game+":"+host] = id
}

func waiting(m *api.Match) (map[string]any, bool) {
	if m == nil {
		return nil, false
	}
	label := map[string]any{}
	if json.Unmarshal([]byte(m.GetLabel().GetValue()), &label) != nil {
		return nil, false
	}
	return label, label["mode"] == ModeFriend && label["status"] == StatusWaiting
}

func Open(ctx context.Context, nk runtime.NakamaModule, game, field, value string) (string, map[string]any) {
	key := field + ":" + value
	if field == "host" {
		key = "host:" + game + ":" + value
	}
	recent.Lock()
	id := recent.ids[key]
	recent.Unlock()
	if id != "" {
		if m, err := nk.MatchGet(ctx, id); err == nil {
			if label, ok := waiting(m); ok && labelMatches(label, game, field, value) {
				return id, label
			}
		}
	}
	query := "+label.mode:" + ModeFriend + " +label.status:" + StatusWaiting + " +label." + field + ":" + value
	if game != "" {
		query += " +label.game:" + game
	}
	min := 0
	matches, err := nk.MatchList(ctx, 5, true, "", &min, nil, query)
	if err != nil {
		return "", nil
	}
	for _, found := range matches {
		m, err := nk.MatchGet(ctx, found.GetMatchId())
		if err != nil {
			continue
		}
		if label, ok := waiting(m); ok && labelMatches(label, game, field, value) {
			return m.GetMatchId(), label
		}
	}
	return "", nil
}

func labelMatches(label map[string]any, game, field, value string) bool {
	if game != "" && label["game"] != game {
		return false
	}
	v, _ := label[field].(string)
	return v == value
}

func SecondsLeft(label map[string]any, now time.Time, fallback int) int {
	at, ok := label["expiresAt"].(float64)
	if !ok || at <= 0 {
		return fallback
	}
	left := int(at - float64(now.UnixMilli())/1000 + 0.999)
	if left < 0 {
		return 0
	}
	return left
}

func IsFriend(ctx context.Context, nk runtime.NakamaModule, userID, other string) bool {
	state := 0
	cursor := ""
	for page := 0; page < 10; page++ {
		friends, next, err := nk.FriendsList(ctx, userID, 100, &state, cursor)
		if err != nil {
			return false
		}
		for _, f := range friends {
			if f.GetUser().GetId() == other {
				return true
			}
		}
		if next == "" {
			return false
		}
		cursor = next
	}
	return false
}
