package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/internal/campaignhooks"
	"github.com/heroiclabs/nakama/v3/internal/eventcontext"
)

func TestPostgresContextSpansFollowTheAttempts(t *testing.T) {
	f := newFixture(t)
	ch := f.chapter(map[string]any{})
	st := f.stage(ch, map[string]any{"mode": "repeatable"})
	user := f.user()
	var chKey string
	if err := f.db.QueryRow(`SELECT key FROM campaign_chapter WHERE id=$1`, ch).Scan(&chKey); err != nil {
		t.Fatal(err)
	}
	campaignhooks.SetAttemptItem(func(_ context.Context, userID, itemKey string) (campaignhooks.ItemInfo, error) {
		if itemKey == "" {
			itemKey = "equipped-car"
		}
		return campaignhooks.ItemInfo{Key: itemKey, Values: map[string]string{"gameId": "G-" + itemKey}}, nil
	}, "gameId")
	eventcontext.Register(eventcontext.Provider{Name: "campaign", Sources: contextSources, Spans: f.s.contextSpans})
	t.Cleanup(func() {
		campaignhooks.SetAttemptItem(nil)
		eventcontext.Unregister("campaign")
	})
	keys := map[string]bool{}
	for _, s := range eventcontext.Sources() {
		keys[s.Key] = true
	}
	if !keys[SourceStage] || !keys[SourceItem] || !keys[SourceItem+".gameId"] || !keys[SourceChapter] {
		t.Fatalf("sources = %v", keys)
	}
	t0 := f.s.now()
	out, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+st+`","itemKey":"tiggo8"}`)
	if err != nil {
		t.Fatal(err)
	}
	attempt := out.(map[string]any)["attemptId"].(string)
	f.advance(30 * time.Second)
	raw, _ := json.Marshal(map[string]any{"attemptId": attempt, "success": true, "elapsedSeconds": 25, "healthPercent": 100})
	if _, err := f.s.rpcComplete(context.Background(), user, string(raw)); err != nil {
		t.Fatal(err)
	}
	f.advance(time.Minute)
	t1 := f.s.now()
	if _, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+st+`"}`); err != nil {
		t.Fatal(err)
	}
	tl, err := eventcontext.Load(context.Background(), user, t0.Add(-time.Hour), t1.Add(time.Hour), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(tl.Spans) != 2 {
		t.Fatalf("spans = %+v", tl.Spans)
	}
	v := tl.At(t0.Add(10 * time.Second))
	if v[SourceStage] != st || v[SourceChapter] != chKey || v[SourceItem] != "tiggo8" || v[SourceItem+".gameId"] != "G-tiggo8" || v[SourceAttempt] != attempt {
		t.Fatalf("first attempt = %v", v)
	}
	if v := tl.At(t0.Add(50 * time.Second)); v != nil {
		t.Fatalf("between attempts = %v", v)
	}
	if v := tl.At(t1.Add(5 * time.Minute)); v[SourceItem] != "equipped-car" || v[SourceItem+".gameId"] != "G-equipped-car" {
		t.Fatalf("open attempt with the fallback item = %v", v)
	}
	other := f.user()
	if tl, err := eventcontext.Load(context.Background(), other, t0, t1, 0); err != nil || len(tl.Spans) != 0 {
		t.Fatalf("other player = %+v %v", tl, err)
	}
}
