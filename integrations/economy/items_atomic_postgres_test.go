package main

import (
	"testing"
)

func TestItemBuyAndUpgradeDebitIsAtomic(t *testing.T) {
	h, skin := newItemHarness(t)
	h.item(map[string]any{"key": "hat", "typeId": skin, "price": map[string]any{"currencyKey": "coin", "amount": 10},
		"upgrades": []map[string]any{{"currencyKey": "coin", "amount": 5}}})
	user := h.user()
	h.give(user, 100)
	for _, q := range []string{
		`CREATE OR REPLACE FUNCTION test_fail_item_request() RETURNS trigger AS $$ BEGIN IF NEW.request_id LIKE 'boom-%' THEN RAISE EXCEPTION 'forced failure'; END IF; RETURN NEW; END $$ LANGUAGE plpgsql`,
		`DROP TRIGGER IF EXISTS test_fail_item_request ON economy_item_request`,
		`CREATE TRIGGER test_fail_item_request BEFORE UPDATE ON economy_item_request FOR EACH ROW EXECUTE FUNCTION test_fail_item_request()`,
	} {
		if _, err := h.s.db.ExecContext(h.ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_, _ = h.s.db.Exec(`DROP TRIGGER IF EXISTS test_fail_item_request ON economy_item_request`)
	})
	if _, err := h.rpc(user, "economy_item_buy", map[string]any{"key": "hat", "requestId": "boom-buy"}); err == nil {
		t.Fatal("forced failure did not fail")
	}
	if h.wallet(user)["coin"] != 100 {
		t.Fatalf("coins taken by a failed buy: %v", h.wallet(user))
	}
	if h.inventory(user)["hat"]["owned"] == true {
		t.Fatal("item granted by a failed buy")
	}
	if _, err := h.s.db.ExecContext(h.ctx, `DROP TRIGGER test_fail_item_request ON economy_item_request`); err != nil {
		t.Fatal(err)
	}
	h.mustRPC(user, "economy_item_buy", map[string]any{"key": "hat", "requestId": "boom-buy"})
	h.mustRPC(user, "economy_item_buy", map[string]any{"key": "hat", "requestId": "boom-buy"})
	if h.wallet(user)["coin"] != 90 {
		t.Fatalf("retry with the same requestId charged %v", h.wallet(user))
	}
	if _, err := h.s.db.ExecContext(h.ctx, `CREATE TRIGGER test_fail_item_request BEFORE UPDATE ON economy_item_request FOR EACH ROW EXECUTE FUNCTION test_fail_item_request()`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.rpc(user, "economy_item_upgrade", map[string]any{"key": "hat", "requestId": "boom-up"}); err == nil {
		t.Fatal("forced upgrade failure did not fail")
	}
	if h.wallet(user)["coin"] != 90 {
		t.Fatalf("coins taken by a failed upgrade: %v", h.wallet(user))
	}
	var n int
	if err := h.s.db.QueryRow(`SELECT count(*) FROM wallet_ledger WHERE user_id=$1 AND metadata->>'reason' IN ('item_buy','item_upgrade')`, user).Scan(&n); err != nil || n != 1 {
		t.Fatalf("ledger rows = %d %v", n, err)
	}
}
