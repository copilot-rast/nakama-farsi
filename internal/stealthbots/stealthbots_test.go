package stealthbots

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/gamecatalog"
)

func TestDefaultsAreOffAndParseKeepsDefaults(t *testing.T) {
	if Of(nil).Enabled || Of(&gamecatalog.Game{}).Enabled {
		t.Fatal("stealth must be off by default")
	}
	s := Parse([]byte(`{"enabled":true,"arriveMinSeconds":20,"arriveMaxSeconds":5,"levelRange":-3}`))
	if !s.Enabled || !s.PoolBots || !s.Trophies || !s.FullPrize || !s.Leaderboard || !s.Progress {
		t.Fatalf("%+v", s)
	}
	if s.ArriveMinSeconds != 5 || s.ArriveMaxSeconds != 20 || s.LevelRange != 0 {
		t.Fatalf("%+v", s)
	}
	if On(&gamecatalog.Game{Stealth: json.RawMessage(`{"enabled":true}`)}) != true {
		t.Fatal("On")
	}
}

func TestTimingStaysInRange(t *testing.T) {
	s := Parse([]byte(`{"enabled":true,"arriveMinSeconds":3,"arriveMaxSeconds":9,"thinkMinMs":500,"thinkMaxMs":900,"pausePercent":50,"pauseMinMs":1000,"pauseMaxMs":2000}`))
	seen := map[time.Duration]bool{}
	paused := 0
	for i := 0; i < 2000; i++ {
		a := s.ArriveDelay()
		if a < 3*time.Second || a > 9*time.Second {
			t.Fatal(a)
		}
		seen[a.Truncate(time.Second)] = true
		d := s.ThinkDelay()
		if d < 500*time.Millisecond || d > 2900*time.Millisecond {
			t.Fatal(d)
		}
		if d > 900*time.Millisecond {
			paused++
		}
	}
	if len(seen) < 5 || paused < 700 || paused > 1300 {
		t.Fatalf("arrivals %d paused %d", len(seen), paused)
	}
	sides := map[int]int{}
	for i := 0; i < 400; i++ {
		sides[s.Side(2)]++
	}
	if sides[0] < 100 || sides[1] < 100 {
		t.Fatal(sides)
	}
	fixed := Parse([]byte(`{"randomSide":false}`))
	if fixed.Side(2) != 1 {
		t.Fatal("fixed side")
	}
}

func TestBoost(t *testing.T) {
	s := Parse([]byte(`{"enabled":true,"guard":true,"guardDailyWins":3,"guardBoost":2}`))
	if s.Boost(2) != 0 || s.Boost(3) != 2 {
		t.Fatal(s.Boost(2), s.Boost(3))
	}
	if Parse([]byte(`{"enabled":true,"guardDailyWins":3}`)).Boost(9) != 0 {
		t.Fatal("guard off")
	}
}

func TestScrubAndMarkers(t *testing.T) {
	view := map[string]any{"key": "g", "botsAllowed": true, "bottom": 1, "settings": map[string]any{"botNames": []any{"a"}, "botPrizeMode": "full", "roundSeconds": 3, "stealth": map[string]any{}},
		"rooms": []map[string]any{{"key": "r", "humanSeats": 1, "botFillSeconds": 5, "players": 2}}}
	Scrub(view)
	raw, _ := json.Marshal(view)
	if m := Markers(raw); len(m) != 0 {
		t.Fatalf("%s %v", raw, m)
	}
	if view["bottom"] != 1 || view["settings"].(map[string]any)["roundSeconds"] != 3 {
		t.Fatalf("%s", raw)
	}
	bad := []string{
		`{"players":[{"id":"x","isBot":true}]}`,
		`{"mode":"online_bot"}`,
		`{"a":{"kind":"bot"}}`,
		`{"opponentId":"bot:hard"}`,
		`{"vsBot":false}`,
		`{"botPrize":{"mode":"full"}}`,
		`{"payload":"{\"trophyFree\":true}"}`,
		`{"metadata":{"bot":true}}`,
		`{"mode":"bot"}`,
	}
	for _, b := range bad {
		if len(Markers([]byte(b))) == 0 {
			t.Fatalf("no marker in %s", b)
		}
	}
	good := `{"mode":"online","players":[{"id":"6c1a","username":"abbot","bottom":2}],"label":"{\"mode\":\"matchmaking\"}"}`
	if m := Markers([]byte(good)); len(m) != 0 {
		t.Fatal(m)
	}
}
