package economyhooks

import (
	"context"
	"errors"
	"testing"
)

func TestPurchaseChecksRunInOwnerOrder(t *testing.T) {
	var calls []string
	SetPurchaseCheck("b", func(context.Context, Purchase) error {
		calls = append(calls, "b")
		return &Refusal{Status: 403, Code: "NO"}
	})
	SetPurchaseCheck("a", func(context.Context, Purchase) error { calls = append(calls, "a"); return nil })
	defer func() { SetPurchaseCheck("a", nil); SetPurchaseCheck("b", nil) }()
	err := CheckPurchase(context.Background(), Purchase{Method: "bazaar"})
	r, ok := AsRefusal(err)
	if !ok || r.Code != "NO" || len(calls) != 2 || calls[0] != "a" {
		t.Fatalf("err=%v calls=%v", err, calls)
	}
}

func TestMergeProductsKeepsExistingFields(t *testing.T) {
	SetProductFields("x", func(_ context.Context, _, _ string, ids []string) (map[string]map[string]any, error) {
		return map[string]map[string]any{ids[0]: {"name": "hook", "assets": []string{"a"}}}, nil
	})
	SetProductFields("y", func(context.Context, string, string, []string) (map[string]map[string]any, error) {
		return nil, errors.New("down")
	})
	defer func() { SetProductFields("x", nil); SetProductFields("y", nil) }()
	list := []map[string]any{{"id": "p1", "name": "economy"}}
	MergeProducts(context.Background(), "u", "fa", list)
	if list[0]["name"] != "economy" || list[0]["assets"] == nil {
		t.Fatalf("merged = %v", list[0])
	}
}

func TestItemsUnlockedWhenAnyHookSays(t *testing.T) {
	SetItemStates("x", func(context.Context, string, string, []string) (map[string]ItemState, error) {
		return map[string]ItemState{"i": {LockMeta: map[string]any{"a": 1}}}, nil
	})
	SetItemStates("y", func(context.Context, string, string, []string) (map[string]ItemState, error) {
		return map[string]ItemState{"i": {Unlocked: true, LockMeta: map[string]any{"a": 2, "b": 3}}}, nil
	})
	defer func() { SetItemStates("x", nil); SetItemStates("y", nil) }()
	st := Items(context.Background(), "u", "", []string{"i"})["i"]
	if !st.Unlocked || st.LockMeta["a"] != 1 || st.LockMeta["b"] != 3 {
		t.Fatalf("state = %+v", st)
	}
}

func TestReadersUnavailableWithoutEconomy(t *testing.T) {
	if _, err := Level(context.Background(), "u", ""); !errors.Is(err, ErrNoLevels) {
		t.Fatalf("level err = %v", err)
	}
	if _, err := Methods(context.Background()); !errors.Is(err, ErrNoMethods) {
		t.Fatalf("methods err = %v", err)
	}
}

func TestGrantPrizeAndSettleMetasNeedEconomy(t *testing.T) {
	SetPrizeGranter(nil)
	SetMetaSettler(nil)
	if PrizesAvailable() {
		t.Fatal("no granter yet")
	}
	if _, err := GrantPrize(context.Background(), PrizeGrant{UserID: "u", PrizeKey: "p"}); !errors.Is(err, ErrNoPrizes) {
		t.Fatalf("err=%v", err)
	}
	if out, err := SettleMetas(context.Background(), "u", []string{"xp"}); out != nil || err != nil {
		t.Fatalf("settle without economy = %v %v", out, err)
	}
	var got PrizeGrant
	SetPrizeGranter(func(_ context.Context, g PrizeGrant) (PrizeResult, error) {
		got = g
		return PrizeResult{Prizes: []map[string]any{{"key": g.PrizeKey}}}, nil
	})
	SetMetaSettler(func(_ context.Context, _ string, keys []string) (map[string]any, error) {
		return map[string]any{"keys": keys}, nil
	})
	defer func() { SetPrizeGranter(nil); SetMetaSettler(nil) }()
	res, err := GrantPrize(context.Background(), PrizeGrant{UserID: "u", PrizeKey: "box", SourceType: "stage", SourceID: "a1"})
	if err != nil || len(res.Prizes) != 1 || got.SourceType != "stage" {
		t.Fatalf("res=%v err=%v got=%+v", res, err, got)
	}
	if out, _ := SettleMetas(context.Background(), "u", []string{"xp"}); out == nil {
		t.Fatal("settler not called")
	}
}
