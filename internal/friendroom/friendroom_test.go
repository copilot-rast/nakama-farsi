package friendroom

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParamsRoundTripAndJoinRules(t *testing.T) {
	at := time.Now().Add(time.Minute).Truncate(time.Millisecond).UTC()
	p := Params{Game: "g", Room: "r", Mode: ModeFriend, Host: "h", Code: NewCode(), Invited: "f", ExpiresAt: at}
	back := FromParams(p.Map())
	if back.Mode != ModeFriend || back.Host != "h" || back.Invited != "f" || !back.ExpiresAt.Equal(at) || !ValidCode(back.Code) {
		t.Fatalf("%+v", back)
	}
	if !back.MayJoin("h") || !back.MayJoin("f") || back.MayJoin("x") {
		t.Fatal("friend room join rules")
	}
	mm := FromParams(map[string]any{"game": "g", "users": "a,b"})
	if mm.Mode != ModeMatchmaking || !mm.MayJoin("b") || mm.MayJoin("c") {
		t.Fatalf("%+v", mm)
	}
	legacy := FromParams(map[string]any{"friend": "1", "host": "h"})
	if legacy.Mode != ModeFriend || !legacy.MayJoin("anyone") {
		t.Fatalf("%+v", legacy)
	}
	bot := FromParams(map[string]any{"bot": "1", "users": "u"})
	if bot.Mode != ModeBot || !bot.MayJoin("u") || bot.MayJoin("v") {
		t.Fatalf("%+v", bot)
	}
	label := map[string]any{}
	_ = json.Unmarshal([]byte(p.Label(StatusWaiting, map[string]any{"players": 1})), &label)
	if label["code"] != p.Code || label["open"] != true || label["players"] != float64(1) || SecondsLeft(label, time.Now(), 0) < 59 {
		t.Fatalf("%v", label)
	}
	if s, ok := ParseSignal(SignalData(ActionCancel, "h")); !ok || s.Action != ActionCancel || s.UserID != "h" {
		t.Fatalf("%+v", s)
	}
	if ok, reason := ParseAnswer(Answer(false, "STARTED")); ok || reason != "STARTED" {
		t.Fatal("answer")
	}
	if ValidCode("ABC") || ValidCode("ABCDE0") {
		t.Fatal("code check")
	}
}
