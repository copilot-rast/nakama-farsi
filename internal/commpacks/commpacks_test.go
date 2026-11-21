package commpacks

import (
	"testing"
	"time"
)

func TestTakeFollowsGapWindowAndLock(t *testing.T) {
	st := Defaults()
	start := time.Unix(1000, 0)
	key := "m|u"
	defer delete(flows, key)
	if r := take(key, st, start); r != nil {
		t.Fatalf("first send refused: %v", r)
	}
	r := take(key, st, start.Add(time.Second))
	if r == nil || r.Code != CodeCooldown || r.RetryAfterMs != 3000 {
		t.Fatalf("gap: %+v", r)
	}
	at := start
	for i := 1; i < st.MaxCount; i++ {
		at = at.Add(4 * time.Second)
		if r := take(key, st, at); r != nil {
			t.Fatalf("send %d refused: %v", i, r)
		}
	}
	at = at.Add(4 * time.Second)
	r = take(key, st, at)
	if r == nil || r.Code != CodeSpamLocked || r.RetryAfterMs != 30000 {
		t.Fatalf("lock: %+v", r)
	}
	if r := take(key, st, at.Add(29*time.Second)); r == nil || r.Code != CodeSpamLocked {
		t.Fatalf("still locked: %+v", r)
	}
	if r := take(key, st, at.Add(31*time.Second)); r != nil {
		t.Fatalf("after lock: %v", r)
	}
}

func TestSettingsCheck(t *testing.T) {
	if f, _ := Defaults().Check(); f != "" {
		t.Fatalf("defaults invalid: %s", f)
	}
	s := Defaults()
	s.MaxCount = 10
	if f, _ := s.Check(); f != "windowSeconds" {
		t.Fatalf("want window error, got %q", f)
	}
	s = Defaults()
	s.MinIntervalSeconds = 0
	s.MaxCount = 1000
	if f, _ := s.Check(); f != "" {
		t.Fatalf("interval off must allow many: %q", f)
	}
}
