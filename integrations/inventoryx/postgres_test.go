package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama/v3/integrations/devkit/apierr"
	"github.com/heroiclabs/nakama/v3/integrations/inventoryx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

type fixture struct {
	t      *testing.T
	db     *sql.DB
	s      *service
	wallet *plugintest.Wallet
	coin   string
	gem    string
	typeID string
	levels map[string]int
	mu     sync.Mutex
}

func newFixture(t *testing.T) *fixture {
	db := plugintest.OpenDB(t)
	f := &fixture{t: t, db: db, wallet: &plugintest.Wallet{DB: db}, levels: map[string]int{}}
	f.s = &service{db: db, nk: f.wallet, logger: plugintest.Logger{}}
	f.s.level = func(_ context.Context, userID, _ string) (int, error) {
		f.mu.Lock()
		defer f.mu.Unlock()
		return f.levels[userID], nil
	}
	f.coin, f.gem = plugintest.Key("coin"), plugintest.Key("gem")
	f.typeID = uuid.Must(uuid.NewV4()).String()
	plugintest.Exec(t, db, `INSERT INTO economy_currency (id, key, name, type) VALUES (gen_random_uuid(), $1, '{"fa":"سکه"}', 'soft'), (gen_random_uuid(), $2, '{"fa":"جم"}', 'hard')`, f.coin, f.gem)
	plugintest.Exec(t, db, `INSERT INTO base_category (id, key_type, key, name) VALUES ($1, 'item_type', $2, '{"fa":"تفنگ"}')`, f.typeID, plugintest.Key("gun"))
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM economy_item WHERE type_id=$1`, f.typeID)
		_, _ = db.Exec(`DELETE FROM base_category WHERE id=$1`, f.typeID)
		_, _ = db.Exec(`DELETE FROM economy_currency WHERE key IN ($1,$2)`, f.coin, f.gem)
		_, _ = db.Exec(`DELETE FROM inventoryx_settings`)
		_, _ = db.Exec(`INSERT INTO inventoryx_settings (id) VALUES (1) ON CONFLICT DO NOTHING`)
		settings.Invalidate()
	})
	return f
}

func (f *fixture) settings(change func(*settings.All)) {
	f.t.Helper()
	cfg := settings.Defaults()
	change(&cfg)
	if err := settings.Save(context.Background(), f.db, cfg, "test"); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) item(starter bool, unlockLevel int, price *int64, cfg itemConfig) (string, string) {
	f.t.Helper()
	id := uuid.Must(uuid.NewV4()).String()
	key := plugintest.Key("it")
	var p any
	if price != nil {
		b, _ := json.Marshal(map[string]any{"currencyKey": f.coin, "amount": *price})
		p = string(b)
	}
	var lv any
	if unlockLevel > 0 {
		lv = unlockLevel
	}
	plugintest.Exec(f.t, f.db, `INSERT INTO economy_item (id, key, type_id, name, is_starter, unlock_level, price) VALUES ($1,$2,$3,'{"fa":"آیتم"}',$4,$5,$6)`, id, key, f.typeID, starter, lv, p)
	if !cfg.empty() {
		if err := saveConfig(context.Background(), f.db, id, cfg, "test"); err != nil {
			f.t.Fatal(err)
		}
	}
	return id, key
}

func code(err error) string {
	if e, ok := apierr.As(err); ok {
		return e.Code
	}
	if err == nil {
		return ""
	}
	return err.Error()
}

func (f *fixture) count(q string, args ...any) int {
	f.t.Helper()
	var n int
	if err := f.db.QueryRow(q, args...).Scan(&n); err != nil {
		f.t.Fatal(err)
	}
	return n
}

func TestUnlockEarlyIsPaidOnceAndOpensBuying(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	user := plugintest.NewUser(t, f.db)
	f.levels[user] = 1
	cost := int64(100)
	itemID, key := f.item(false, 5, &cost, itemConfig{UnlockPrices: []price{{CurrencyKey: f.gem, Amount: 40}}})
	f.wallet.Give(t, user, map[string]int64{f.gem: 50}, map[string]any{"reason": "test"})

	r, err := f.s.unlock(ctx, user, key, "req-1", "")
	if err != nil || r.Replayed || len(r.Paid) != 1 || r.Paid[0].Amount != 40 || r.Owned {
		t.Fatalf("unlock = %+v %v", r, err)
	}
	again, err := f.s.unlock(ctx, user, key, "req-1", "")
	if err != nil || !again.Replayed || again.Paid[0].Amount != 40 {
		t.Fatalf("replay = %+v %v", again, err)
	}
	if _, err := f.s.unlock(ctx, user, key, "req-2", ""); code(err) != "ALREADY_UNLOCKED" {
		t.Fatalf("second request = %v", err)
	}
	if _, err := f.s.unlock(ctx, user, key, "req-1", "x"); code(err) != "REQUEST_ID_REUSED" {
		t.Fatalf("reused id = %v", err)
	}
	if got := plugintest.Balance(t, f.db, user)[f.gem]; got != 10 {
		t.Fatalf("gems = %d, want 10", got)
	}
	states, err := f.s.itemStates(ctx, user, "en", []string{itemID})
	if err != nil || !states[itemID].Unlocked || states[itemID].Fields["unlockedEarly"] != true || states[itemID].Fields["canUnlockEarly"] != false {
		t.Fatalf("state = %+v %v", states[itemID], err)
	}

	poor := plugintest.NewUser(t, f.db)
	if _, err := f.s.unlock(ctx, poor, key, "p-1", ""); code(err) != "INSUFFICIENT_BALANCE" {
		t.Fatalf("poor = %v", err)
	}
	if n := f.count(`SELECT count(*) FROM inventoryx_user_unlock WHERE user_id=$1`, poor); n != 0 {
		t.Fatalf("a failed unlock left %d rows", n)
	}
	high := plugintest.NewUser(t, f.db)
	f.levels[high] = 9
	if _, err := f.s.unlock(ctx, high, key, "h-1", ""); code(err) != "ALREADY_UNLOCKED" {
		t.Fatalf("level reached = %v", err)
	}
}

func TestUnlockRaceTakesOnePayment(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	user := plugintest.NewUser(t, f.db)
	_, key := f.item(false, 3, nil, itemConfig{UnlockPrices: []price{{CurrencyKey: f.gem, Amount: 5}}})
	f.wallet.Give(t, user, map[string]int64{f.gem: 100}, map[string]any{"reason": "test"})
	var wg sync.WaitGroup
	var mu sync.Mutex
	codes := map[string]int{}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := "race-" + string(rune('a'+i))
			if i%2 == 1 {
				id = "race-same"
			}
			r, err := f.s.unlock(ctx, user, key, id, "")
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err != nil:
				codes[code(err)]++
			case r.Replayed:
				codes["replayed"]++
			default:
				codes["ok"]++
			}
		}(i)
	}
	wg.Wait()
	if codes["ok"] != 1 {
		t.Fatalf("codes = %v", codes)
	}
	if got := plugintest.Balance(t, f.db, user)[f.gem]; got != 95 {
		t.Fatalf("gems = %d, want 95 (codes %v)", got, codes)
	}
	if n := f.count(`SELECT count(*) FROM economy_user_item u JOIN economy_item i ON i.id=u.item_id WHERE u.user_id=$1 AND i.key=$2 AND u.source='unlock'`, user, key); n != 1 {
		t.Fatalf("free item owned %d times", n)
	}
}

func TestUnlockPriceChoice(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.settings(func(c *settings.All) { c.Unlock.PriceMode = settings.PriceChoice; c.Unlock.OwnFreeItems = false })
	user := plugintest.NewUser(t, f.db)
	_, key := f.item(false, 3, nil, itemConfig{UnlockPrices: []price{{CurrencyKey: f.gem, Amount: 5}, {CurrencyKey: f.coin, Amount: 500}}})
	f.wallet.Give(t, user, map[string]int64{f.gem: 10, f.coin: 1000}, map[string]any{"reason": "test"})
	if _, err := f.s.unlock(ctx, user, key, "c-1", ""); code(err) != "CURRENCY_REQUIRED" {
		t.Fatalf("no currency = %v", err)
	}
	if _, err := f.s.unlock(ctx, user, key, "c-2", "ruby"); code(err) != "CURRENCY_NOT_ACCEPTED" {
		t.Fatalf("bad currency = %v", err)
	}
	r, err := f.s.unlock(ctx, user, key, "c-3", f.coin)
	if err != nil || r.Owned {
		t.Fatalf("choice = %+v %v", r, err)
	}
	b := plugintest.Balance(t, f.db, user)
	if b[f.coin] != 500 || b[f.gem] != 10 {
		t.Fatalf("balance = %v", b)
	}
}

func gunConfig(coin string) itemConfig {
	v12, v15, v20 := 12.0, 15.0, 20.0
	return itemConfig{Tracks: []track{
		{Key: "Power", CurrencyKey: coin, Base: 10, Levels: []step{{Price: 500, Value: &v12}, {Price: 1000, Value: &v15}, {Price: 1500, Value: &v20, Requires: []need{{Track: "MagazineCapacity", Level: 1}}}}},
		{Key: "MagazineCapacity", CurrencyKey: coin, Levels: []step{{Price: 300}, {Price: 600, MinPlayerLevel: 4}}},
	}}
}

func TestUpgradeTracksFromLevelAndRequirements(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	user := plugintest.NewUser(t, f.db)
	f.levels[user] = 1
	itemID, key := f.item(true, 0, nil, gunConfig(f.coin))
	f.wallet.Give(t, user, map[string]int64{f.coin: 10000}, map[string]any{"reason": "test"})
	zero, one := 0, 1

	r, err := f.s.upgrade(ctx, user, key, "Power", &zero, "u-1")
	if err != nil || r.Level != 1 || r.Value != 12 || r.Paid[0].Amount != 500 {
		t.Fatalf("first = %+v %v", r, err)
	}
	if again, err := f.s.upgrade(ctx, user, key, "Power", &zero, "u-1"); err != nil || !again.Replayed || again.Level != 1 {
		t.Fatalf("replay = %+v %v", again, err)
	}
	if _, err := f.s.upgrade(ctx, user, key, "Power", &zero, "u-2"); code(err) != "UPGRADE_CHANGED" {
		t.Fatalf("stale fromLevel = %v", err)
	}
	if _, err := f.s.upgrade(ctx, user, key, "Power", &one, "u-3"); err != nil {
		t.Fatal(err)
	}
	two := 2
	if _, err := f.s.upgrade(ctx, user, key, "Power", &two, "u-4"); code(err) != "TRACK_REQUIREMENT" {
		t.Fatalf("needs magazine = %v", err)
	}
	if _, err := f.s.upgrade(ctx, user, key, "MagazineCapacity", &zero, "u-5"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.s.upgrade(ctx, user, key, "MagazineCapacity", &one, "u-6"); code(err) != "LEVEL_TOO_LOW" {
		t.Fatalf("player level = %v", err)
	}
	if r, err := f.s.upgrade(ctx, user, key, "Power", &two, "u-7"); err != nil || r.Value != 20 {
		t.Fatalf("third = %+v %v", r, err)
	}
	three := 3
	if _, err := f.s.upgrade(ctx, user, key, "Power", &three, "u-8"); code(err) != "UPGRADE_MAXED" {
		t.Fatalf("max = %v", err)
	}
	if _, err := f.s.upgrade(ctx, user, key, "Range", nil, "u-9"); code(err) != "TRACK_NOT_FOUND" {
		t.Fatalf("unknown track = %v", err)
	}
	if got := plugintest.Balance(t, f.db, user)[f.coin]; got != 10000-500-1000-300-1500 {
		t.Fatalf("coins = %d", got)
	}
	states, err := f.s.itemStates(ctx, user, "en", []string{itemID})
	if err != nil {
		t.Fatal(err)
	}
	attrs := states[itemID].Fields["attributes"].(map[string]any)
	if attrs["Power"] != 20.0 || attrs["MagazineCapacity"] != 1.0 {
		t.Fatalf("attributes = %v", attrs)
	}

	other := plugintest.NewUser(t, f.db)
	_, notMine := f.item(false, 0, nil, gunConfig(f.coin))
	if _, err := f.s.upgrade(ctx, other, notMine, "Power", nil, "o-1"); code(err) != "NOT_OWNED" {
		t.Fatalf("not owned = %v", err)
	}
	poor := plugintest.NewUser(t, f.db)
	if _, err := f.s.upgrade(ctx, poor, key, "Power", &zero, "p-1"); code(err) != "INSUFFICIENT_BALANCE" {
		t.Fatalf("poor = %v", err)
	}
	if n := f.count(`SELECT COALESCE(max(level),0) FROM inventoryx_user_track WHERE user_id=$1`, poor); n != 0 {
		t.Fatalf("a failed upgrade raised the level to %d", n)
	}
}

func TestUpgradeRaceBuysOneLevel(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	user := plugintest.NewUser(t, f.db)
	_, key := f.item(true, 0, nil, gunConfig(f.coin))
	f.wallet.Give(t, user, map[string]int64{f.coin: 10000}, map[string]any{"reason": "test"})
	var wg sync.WaitGroup
	var mu sync.Mutex
	got := map[string]int{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			zero := 0
			_, err := f.s.upgrade(ctx, user, key, "Power", &zero, "race-"+strings.Repeat("x", i+1))
			mu.Lock()
			got[code(err)]++
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	if got[""] != 1 || got["UPGRADE_CHANGED"] != 9 {
		t.Fatalf("results = %v", got)
	}
	if b := plugintest.Balance(t, f.db, user)[f.coin]; b != 9500 {
		t.Fatalf("coins = %d", b)
	}
}

func TestRetryAfterWalletCommitDoesNotPayTwice(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	user := plugintest.NewUser(t, f.db)
	_, key := f.item(true, 0, nil, gunConfig(f.coin))
	f.wallet.Give(t, user, map[string]int64{f.coin: 1000}, map[string]any{"reason": "test"})
	gid := grantID(user, "lost-1")
	f.wallet.Give(t, user, map[string]int64{f.coin: -500}, map[string]any{"grantId": gid})
	zero := 0
	r, err := f.s.upgrade(ctx, user, key, "Power", &zero, "lost-1")
	if err != nil || r.Level != 1 {
		t.Fatalf("retry = %+v %v", r, err)
	}
	if b := plugintest.Balance(t, f.db, user)[f.coin]; b != 500 {
		t.Fatalf("coins = %d, want 500 (paid once)", b)
	}
}

func TestImportLegacyRows(t *testing.T) {
	f := newFixture(t)
	_, key := f.item(true, 0, nil, itemConfig{})
	row := importRow{}
	raw := `{"key":"` + key + `","unlockPrices":[{"currencyId":"` + f.gem + `","amount":40}],"upgrades":[{"key":"Power","currencyId":"` + f.coin + `","prices":[500,1000]},{"key":"ReloadTime","currencyId":"` + f.coin + `","prices":[500]}]}`
	if err := json.Unmarshal([]byte(raw), &row); err != nil {
		t.Fatal(err)
	}
	cur, err := currencyKeys(context.Background(), f.db)
	if err != nil {
		t.Fatal(err)
	}
	c, err := cleanConfig(settings.Defaults(), nil, cur, row.config("fa"))
	if err != nil {
		t.Fatal(err)
	}
	if len(c.UnlockPrices) != 1 || c.UnlockPrices[0].Amount != 40 || len(c.Tracks) != 2 || len(c.Tracks[0].Levels) != 2 || c.Tracks[0].Levels[1].Price != 1000 {
		t.Fatalf("config = %+v", c)
	}
}

func TestStoppedItemUpgradeFollowsTheSetting(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	user := plugintest.NewUser(t, f.db)
	f.levels[user] = 1
	itemID, key := f.item(true, 0, nil, gunConfig(f.coin))
	f.wallet.Give(t, user, map[string]int64{f.coin: 10000}, map[string]any{"reason": "test"})
	plugintest.Exec(t, f.db, `UPDATE economy_item SET is_active=false WHERE id=$1`, itemID)
	zero, one := 0, 1
	if r, err := f.s.upgrade(ctx, user, key, "Power", &zero, "s-1"); err != nil || r.Level != 1 {
		t.Fatalf("upgrade of a stopped item by default = %+v %v", r, err)
	}
	f.settings(func(a *settings.All) { off := false; a.Upgrade.StoppedItems = &off })
	if _, err := f.s.upgrade(ctx, user, key, "Power", &one, "s-2"); code(err) != "ITEM_SALE_STOPPED" {
		t.Fatalf("setting off = %v", err)
	}
	if _, err := f.s.unlock(ctx, user, key, "s-3", ""); code(err) != "ITEM_SALE_STOPPED" {
		t.Fatalf("unlock of a stopped item = %v", err)
	}
}
