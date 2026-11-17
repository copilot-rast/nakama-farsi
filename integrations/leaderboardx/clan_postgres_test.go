package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/internal/clanhooks"
	"github.com/heroiclabs/nakama/v3/internal/gameevents"
)

func TestClanBoardsAndInClanStandings(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	clanA, clanB := uuid.Must(uuid.NewV4()).String(), uuid.Must(uuid.NewV4()).String()
	a1, a2, b1 := f.user(), f.user(), f.user()
	members := map[string][]string{clanA: {a1, a2}, clanB: {b1}}
	clanhooks.Set(func(_ context.Context, u string) (clanhooks.Membership, bool) {
		for c, list := range members {
			for _, x := range list {
				if x == u {
					return clanhooks.Membership{ClanID: c, Role: "member"}, true
				}
			}
		}
		return clanhooks.Membership{}, false
	}, func(_ context.Context, c string) []string { return members[c] }, nil)
	defer clanhooks.Set(nil, nil, nil)

	if code, out := f.console(http.MethodPost, f.s.boardSave, nil, map[string]any{"key": "pt_clan_bad", "name": map[string]string{"fa": "ب", "en": "b"}, "scope": "clan",
		"source": map[string]any{"type": "game_event", "event": "clan_xp", "calc": "value", "field": "delta"}, "rewards": []map[string]any{{"fromRank": 1, "toRank": 1, "currencies": map[string]int{f.coin: 5}}}}); code != http.StatusBadRequest || out["error"] != "LEADERBOARD_CLAN_NO_REWARDS" {
		t.Fatalf("clan boards pay no rewards: %d %v", code, out)
	}
	_, clanKey := f.board(map[string]any{"scope": "clan", "operator": "incr", "source": map[string]any{"type": "game_event", "event": "clan_xp", "calc": "value", "field": "delta"}})
	_, memberKey := f.board(map[string]any{"operator": "set", "source": map[string]any{"type": "game_event", "event": "trophy_changed", "calc": "value", "field": "trophies"}})
	for i, e := range []struct {
		user, name string
		field      string
		v          float64
	}{{clanA, "clan_xp", "delta", 40}, {clanB, "clan_xp", "delta", 70}, {clanA, "clan_xp", "delta", 50}, {a1, "trophy_changed", "trophies", 120}, {a2, "trophy_changed", "trophies", 300}, {b1, "trophy_changed", "trophies", 999}} {
		_, _ = f.s.onGameEvent(ctx, gameevents.Event{ID: "t" + uuid.Must(uuid.NewV4()).String() + string(rune('a'+i)), UserID: e.user, Name: e.name, Fields: map[string]any{e.field: e.v}})
	}
	out, err := f.rpc(f.s.rpcBoard, a1, map[string]any{"key": clanKey})
	if err != nil {
		t.Fatal(err)
	}
	top := out["top"].([]any)
	if len(top) != 2 || top[0].(map[string]any)["clanId"] != clanA || top[0].(map[string]any)["score"].(float64) != 90 || out["scope"] != "clan" {
		t.Fatalf("clan board ranks clans: %v", out)
	}
	if me := out["me"].(map[string]any); me["rank"].(float64) != 1 {
		t.Fatalf("me is my clan: %v", me)
	}
	out, err = f.rpc(f.s.rpcBoard, a1, map[string]any{"key": memberKey, "clan": true})
	if err != nil {
		t.Fatal(err)
	}
	inner := out["top"].([]any)
	if len(inner) != 2 || inner[0].(map[string]any)["userId"] != a2 || inner[0].(map[string]any)["rank"].(float64) != 1 || inner[1].(map[string]any)["boardRank"].(float64) != 3 {
		t.Fatalf("in-clan standings rank only clan members: %v", inner)
	}
	if me := out["me"].(map[string]any); me["rank"].(float64) != 2 {
		t.Fatalf("my in-clan rank: %v", me)
	}
	if _, err := f.rpc(f.s.rpcBoard, f.user(), map[string]any{"key": memberKey, "clan": true}); err == nil {
		t.Fatal("a player without a clan has no in-clan board")
	}
}
