package main

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/heroiclabs/nakama/v3/integrations/campaign/internal/settings"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

func (f *fixture) attempts(user string) int {
	var n int
	_ = f.db.QueryRow(`SELECT count(*) FROM campaign_attempt WHERE user_id=$1`, user).Scan(&n)
	return n
}

func TestPostgresPlayCostWalletFallback(t *testing.T) {
	f := newFixture(t)
	f.settings(func(a *settings.All) {
		a.Stars.MissingField = settings.MissingPass
		a.Attempts.CloseOthers = false
		a.Play.CostCurrency = f.coin
	})
	ch := f.chapter(map[string]any{})
	st := f.stage(ch, map[string]any{"mode": "repeatable", "minSeconds": 0})
	user := f.user()
	f.wallet.Give(t, user, map[string]int64{f.coin: 2}, map[string]any{"reason": "test"})
	out, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+st+`"}`)
	if err != nil {
		t.Fatal(err)
	}
	pc := roundTrip(out)["playCost"].(map[string]any)
	if pc["spent"] != float64(1) || pc["via"] != "wallet" || plugintest.Balance(t, f.db, user)[f.coin] != 1 {
		t.Fatalf("start cost = %v %v", pc, plugintest.Balance(t, f.db, user))
	}
	won := roundTrip(out)["attemptId"].(string)
	lost := f.start(user, st)
	if plugintest.Balance(t, f.db, user)[f.coin] != 0 {
		t.Fatalf("second start = %v", plugintest.Balance(t, f.db, user))
	}
	_, err = f.s.rpcStart(context.Background(), user, `{"stageKey":"`+st+`"}`)
	if e, ok := apierr.As(err); !ok || e.Code != "NOT_ENOUGH_PLAY_COST" || e.Meta["currencyKey"] != f.coin {
		t.Fatalf("broke start = %v", err)
	}
	if f.attempts(user) != 2 {
		t.Fatalf("refused start made an attempt: %d", f.attempts(user))
	}
	f.advance(10 * time.Second)
	res, err := f.complete(user, won, true, 5, 90)
	if err != nil {
		t.Fatal(err)
	}
	if r, _ := res["playCostRefund"].(map[string]any); r == nil || r["amount"] != float64(1) || plugintest.Balance(t, f.db, user)[f.coin] != 1 {
		t.Fatalf("win refund = %v %v", res["playCostRefund"], plugintest.Balance(t, f.db, user))
	}
	if again, err := f.complete(user, won, true, 5, 90); err != nil || again["replayed"] != true || plugintest.Balance(t, f.db, user)[f.coin] != 1 {
		t.Fatalf("resend refunded twice: %v %v", err, plugintest.Balance(t, f.db, user))
	}
	res, err = f.complete(user, lost, false, 5, 0)
	if err != nil || res["playCostRefund"] != nil || plugintest.Balance(t, f.db, user)[f.coin] != 1 {
		t.Fatalf("loss refunded: %v %v %v", res, err, plugintest.Balance(t, f.db, user))
	}
	var refunds int
	_ = f.db.QueryRow(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1 AND metadata->>'reason'='play_cost_refund'`, user).Scan(&refunds)
	if refunds != 1 {
		t.Fatalf("refund rows = %d", refunds)
	}

	free := f.stage(ch, map[string]any{"mode": "repeatable", "metadata": map[string]any{"playCost": 0}})
	if _, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+free+`"}`); err != nil {
		t.Fatalf("stage with playCost 0 = %v", err)
	}
	pricey := f.stage(ch, map[string]any{"mode": "repeatable", "metadata": map[string]any{"playCost": 3}})
	if _, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+pricey+`"}`); code(err) != "NOT_ENOUGH_PLAY_COST" {
		t.Fatalf("stage with playCost 3 = %v", err)
	}

	f.settings(func(a *settings.All) {
		a.Stars.MissingField = settings.MissingPass
		a.Attempts.CloseOthers = false
		a.Play.CostCurrency = f.coin
		a.Play.RefundOnWin = false
	})
	a := f.start(user, st)
	f.advance(10 * time.Second)
	if res, err := f.complete(user, a, true, 5, 90); err != nil || res["playCostRefund"] != nil || plugintest.Balance(t, f.db, user)[f.coin] != 0 {
		t.Fatalf("refundOnWin off = %v %v %v", res, err, plugintest.Balance(t, f.db, user))
	}
	f.settings(func(a *settings.All) { a.Stars.MissingField = settings.MissingPass })
	if _, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+st+`"}`); err != nil {
		t.Fatalf("cost off with an empty wallet = %v", err)
	}
	if code, out := f.console(http.MethodPut, f.s.settingsPut, "", map[string]any{"play": map[string]any{"costCurrency": "Bad Key", "costAmount": 1}}); code != 400 {
		t.Fatalf("bad currency setting = %d %v", code, out)
	}
}

func TestPostgresPlayCostThroughRegenHooks(t *testing.T) {
	f := newFixture(t)
	f.settings(func(a *settings.All) {
		a.Stars.MissingField = settings.MissingPass
		a.Play.CostCurrency = "heart"
	})
	var mu sync.Mutex
	hearts := map[string]int64{}
	spent := map[string]int64{}
	refunds := map[string]int{}
	next := time.Now().Add(10 * time.Minute).UTC()
	f.s.spend = func(_ context.Context, user, currency string, amount int64, source string) (economyhooks.RegenSpend, error) {
		mu.Lock()
		defer mu.Unlock()
		if currency != "heart" {
			t.Errorf("currency = %s", currency)
		}
		if _, ok := spent[source]; ok {
			return economyhooks.RegenSpend{Spent: amount, Replayed: true}, nil
		}
		if hearts[user] < amount {
			return economyhooks.RegenSpend{}, &economyhooks.Refusal{Status: 409, Code: "NOT_ENOUGH_CURRENCY", Meta: map[string]any{"nextAt": next, "value": hearts[user]}}
		}
		hearts[user] -= amount
		spent[source] = amount
		return economyhooks.RegenSpend{Spent: amount, View: economyhooks.Regen{CurrencyKey: "heart", Value: hearts[user], Regen: true, Max: 5}}, nil
	}
	f.s.refund = func(_ context.Context, user, currency, source string) (economyhooks.RegenRefund, error) {
		mu.Lock()
		defer mu.Unlock()
		refunds[source]++
		if refunds[source] == 1 {
			hearts[user] += spent[source]
		}
		return economyhooks.RegenRefund{Found: true, Refunded: spent[source], Replayed: refunds[source] > 1}, nil
	}
	ch := f.chapter(map[string]any{})
	st := f.stage(ch, map[string]any{"mode": "repeatable", "minSeconds": 0})
	user := f.user()
	hearts[user] = 1
	out, err := f.s.rpcStart(context.Background(), user, `{"stageKey":"`+st+`"}`)
	if err != nil {
		t.Fatal(err)
	}
	m := roundTrip(out)
	if m["regen"] == nil || m["playCost"].(map[string]any)["via"] != "economy" {
		t.Fatalf("start = %v", m)
	}
	_, err = f.s.rpcStart(context.Background(), user, `{"stageKey":"`+st+`"}`)
	e, ok := apierr.As(err)
	if !ok || e.Code != "NOT_ENOUGH_PLAY_COST" || e.Meta["nextAt"] == nil || e.Status != 409 {
		t.Fatalf("refusal = %v", err)
	}
	if f.attempts(user) != 1 {
		t.Fatalf("attempts = %d", f.attempts(user))
	}
	attempt := m["attemptId"].(string)
	f.advance(10 * time.Second)
	if _, err := f.complete(user, attempt, true, 5, 90); err != nil {
		t.Fatal(err)
	}
	if _, err := f.complete(user, attempt, true, 5, 90); err != nil {
		t.Fatal(err)
	}
	if refunds["campaign:"+attempt] != 1 || hearts[user] != 1 {
		t.Fatalf("refunds = %v hearts = %v", refunds, hearts)
	}
}
