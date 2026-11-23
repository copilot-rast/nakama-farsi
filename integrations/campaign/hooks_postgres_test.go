package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/campaignhooks"
)

func TestPostgresStartGateAndResultObserversSeeTheItem(t *testing.T) {
	f := newFixture(t)
	ch := f.chapter(map[string]any{})
	st := f.stage(ch, map[string]any{"mode": "repeatable"})
	user := f.user()
	var gate []campaignhooks.StartCheck
	var seen []campaignhooks.Result
	campaignhooks.SetStartGate("test", func(_ context.Context, c campaignhooks.StartCheck) error {
		gate = append(gate, c)
		if c.ItemKey == "broken-car" {
			return apierr.Forbidden("CAR_NEEDS_SERVICE")
		}
		return nil
	})
	campaignhooks.SetResultObserver("test", func(_ context.Context, r campaignhooks.Result) (any, error) {
		seen = append(seen, r)
		return map[string]any{"item": r.ItemKey}, nil
	})
	t.Cleanup(func() {
		campaignhooks.SetStartGate("test", nil)
		campaignhooks.SetResultObserver("test", nil)
	})

	if _, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+st+`","vehicleKey":"broken-car"}`); code(err) != "CAR_NEEDS_SERVICE" {
		t.Fatalf("gate = %v", err)
	}
	out, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+st+`","vehicleKey":"audi"}`)
	if err != nil {
		t.Fatal(err)
	}
	attempt := out.(map[string]any)["attemptId"].(string)
	if len(gate) != 2 || gate[1].ItemKey != "audi" || gate[1].StageKey != st || gate[1].Server {
		t.Fatalf("gate calls = %+v", gate)
	}
	var stored string
	if err := f.db.QueryRow(`SELECT item_key FROM campaign_attempt WHERE id=$1`, attempt).Scan(&stored); err != nil || stored != "audi" {
		t.Fatalf("item_key = %q %v", stored, err)
	}
	f.advance(20 * time.Second)
	raw, _ := json.Marshal(map[string]any{"attemptId": attempt, "success": true, "elapsedSeconds": 15, "healthPercent": 40, "fields": map[string]any{"distanceKm": 2.5}})
	res, err := f.s.rpcComplete(context.Background(), user, string(raw))
	if err != nil {
		t.Fatal(err)
	}
	m := roundTrip(res)
	ext, _ := m["extensions"].(map[string]any)
	if ext == nil || ext["test"].(map[string]any)["item"] != "audi" {
		t.Fatalf("extensions = %v", m["extensions"])
	}
	if len(seen) != 1 || seen[0].ItemKey != "audi" || seen[0].StageKey != st || seen[0].HealthPercent == nil || *seen[0].HealthPercent != 40 || seen[0].Fields["distanceKm"] != 2.5 || seen[0].Replayed || seen[0].ElapsedSeconds != 15 {
		t.Fatalf("observer = %+v", seen)
	}
	if _, err := f.s.rpcComplete(context.Background(), user, string(raw)); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 2 || !seen[1].Replayed || seen[1].AttemptID != attempt {
		t.Fatalf("replay observer = %+v", seen)
	}
}
