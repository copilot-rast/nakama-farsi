package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/itemkinds"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

func fakeSaver(ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key         string          `json:"key"`
		TypeID      string          `json:"typeId"`
		Name        json.RawMessage `json:"name"`
		Price       json.RawMessage `json:"price"`
		UnlockLevel *int            `json:"unlockLevel"`
		ParentKey   string          `json:"parentKey"`
		RequiresKey string          `json:"requiresKey"`
		IsActive    *bool           `json:"isActive"`
		Metadata    json.RawMessage `json:"metadata"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return err
	}
	for _, ref := range []string{in.ParentKey, in.RequiresKey} {
		if ref == "" {
			continue
		}
		var ok bool
		_ = tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_item WHERE key=$1)`, ref).Scan(&ok)
		if !ok {
			return &economyhooks.Refusal{Status: 400, Code: "ITEM_NOT_FOUND", Message: "item " + ref + " does not exist", Meta: map[string]any{"field": "parentKey"}}
		}
	}
	var price any
	if len(in.Price) > 0 && string(in.Price) != "null" {
		price = string(in.Price)
	}
	null := func(s string) any {
		if s == "" {
			return nil
		}
		return s
	}
	meta := string(in.Metadata)
	if meta == "" || meta == "null" {
		meta = "{}"
	}
	active := in.IsActive == nil || *in.IsActive
	if create {
		_, err := tx.ExecContext(ctx, `INSERT INTO economy_item (id, key, type_id, name, price, unlock_level, parent_key, requires_key, is_active, metadata) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			id, in.Key, in.TypeID, string(in.Name), price, in.UnlockLevel, null(in.ParentKey), null(in.RequiresKey), active, meta)
		return err
	}
	_, err := tx.ExecContext(ctx, `UPDATE economy_item SET name=$2, price=$3, unlock_level=$4, parent_key=$5, requires_key=$6, is_active=$7, metadata=$8, updated_at=now() WHERE id=$1`,
		id, string(in.Name), price, in.UnlockLevel, null(in.ParentKey), null(in.RequiresKey), active, meta)
	return err
}

type fixture struct {
	t        *testing.T
	db       *sql.DB
	s        *service
	kind     string
	itemType string
	coin     string
	entries  []string
}

func newFixture(t *testing.T) *fixture {
	db := plugintest.OpenDB(t)
	f := &fixture{t: t, db: db, s: &service{db: db, logger: plugintest.Logger{}}, itemType: plugintest.Key("gadget"), coin: plugintest.Key("coin")}
	f.kind = "g" + hashShort(f.itemType)
	economyhooks.SetItemSaver(fakeSaver)
	plugintest.Exec(t, db, `INSERT INTO economy_currency (id, key, name, type, metadata) VALUES (gen_random_uuid(), $1, '{"fa":"سکه"}', 'soft', jsonb_build_object('cardId', $1::text || '-card'))`, f.coin)
	coin := f.coin
	must(t, itemkinds.Install(context.Background(), db, itemkinds.Kind{
		Key: f.kind, ItemType: f.itemType, Label: itemkinds.Text{"fa": "ابزار", "en": "Gadget"}, Order: 10,
		Fields:    []itemkinds.Field{{Key: "gameId", Type: "text", ReadOnly: true}, {Key: "power", Type: "number"}, {Key: "note", Type: "text"}},
		Ownership: []itemkinds.Ownable{{Field: "price", Default: "console"}, {Field: "data.note", Default: "console"}},
		Routes: []itemkinds.Route{
			{SourceType: "Test.GadgetCard", Role: itemkinds.RoleItem, Origin: f.kind + ":card"},
			{SourceType: "Test.GadgetConfig", Role: itemkinds.RoleDependent, Origin: f.kind + ":config"},
		},
		KeyOf: func(e *itemkinds.Entry) string { return "gadget-" + itemkinds.Str(e.Data["Id"]) },
		Map: func(ctx context.Context, m *itemkinds.Mapper, e *itemkinds.Entry) (*itemkinds.Mapped, error) {
			out := &itemkinds.Mapped{Key: "gadget-" + itemkinds.Str(e.Data["Id"]), MatchMeta: map[string]string{"gameId": itemkinds.Str(e.Data["Id"])},
				Name: itemkinds.Text{"fa": itemkinds.Str(e.Data["name"])}, Metadata: map[string]any{"gameId": itemkinds.Str(e.Data["Id"])},
				Data: map[string]any{"gameId": itemkinds.Str(e.Data["Id"]), "note": "from unity"}}
			if n, ok := itemkinds.Int(e.Data["coin"]); ok {
				out.Price = &itemkinds.Price{CurrencyKey: coin, Amount: n}
			}
			cfg, missing := m.Ref(e.Data["config"])
			if cfg == nil {
				out.Pending = "config " + missing + " is not on the server"
				return out, nil
			}
			out.Data["power"] = cfg.Data["power"]
			return out, nil
		},
		CheckBuy: func(ctx context.Context, in itemkinds.BuyInput) error {
			if p, _ := itemkinds.Num(in.Profile.Data["power"]); p > 100 {
				return &itemkinds.Refusal{Code: "TOO_STRONG", Message: "too strong"}
			}
			return nil
		},
		Slot: func(p itemkinds.Profile) string { return "gadget:" + p.ItemKey },
	}))
	t.Cleanup(func() {
		ctx := context.Background()
		for _, id := range f.entries {
			_, _ = db.ExecContext(ctx, `DELETE FROM assets_entry WHERE asset_id=$1`, id)
		}
		_, _ = db.ExecContext(ctx, `DELETE FROM itemkind_profile WHERE kind=$1`, f.kind)
		_, _ = db.ExecContext(ctx, `DELETE FROM economy_item WHERE type_id=(SELECT id FROM base_category WHERE key_type='item_type' AND key=$1)`, f.itemType)
		_, _ = db.ExecContext(ctx, `DELETE FROM base_category WHERE key_type='item_type' AND key=$1`, f.itemType)
		_, _ = db.ExecContext(ctx, `DELETE FROM economy_currency WHERE key=$1`, f.coin)
		_, _ = db.ExecContext(ctx, `DELETE FROM itemkind_route WHERE origin LIKE $1`, f.kind+":%")
		economyhooks.SetItemSaver(nil)
	})
	return f
}

func hashShort(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:10]
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func (f *fixture) put(id, typ string, data map[string]any) {
	f.t.Helper()
	raw, _ := json.Marshal(data)
	sum := sha256.Sum256(raw)
	h := hex.EncodeToString(sum[:])
	plugintest.Exec(f.t, f.db, `INSERT INTO assets_entry (asset_id, type, name, data, hash) VALUES ($1,$2,$1,$3,$4)
		ON CONFLICT (asset_id) DO UPDATE SET data=EXCLUDED.data, hash=EXCLUDED.hash, version=assets_entry.version+1`, id, typ, string(raw), h)
	f.entries = append(f.entries, id)
}

func (f *fixture) run(in itemkinds.RunInput) itemkinds.RunOut {
	f.t.Helper()
	ctx := context.Background()
	tx, err := f.db.BeginTx(ctx, nil)
	must(f.t, err)
	defer func() { _ = tx.Rollback() }()
	out, err := itemkinds.Run(ctx, tx, itemkinds.Defaults(), in)
	must(f.t, err)
	if !in.DryRun {
		must(f.t, tx.Commit())
	}
	return out
}

func rowFor(out itemkinds.RunOut, id string) itemkinds.Row {
	for _, r := range out.Rows {
		if r.AssetID == id {
			return r
		}
	}
	return itemkinds.Row{}
}

func (f *fixture) item(key string) (string, string, bool, string) {
	var id, price, meta string
	var active bool
	err := f.db.QueryRow(`SELECT id::text, COALESCE(price::text,''), is_active, metadata::text FROM economy_item WHERE key=$1`, key).Scan(&id, &price, &active, &meta)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", false, ""
	}
	must(f.t, err)
	return id, price, active, meta
}

func TestRouterCreatesLinksAndKeepsConsoleFields(t *testing.T) {
	f := newFixture(t)
	p := plugintest.Key("p")
	card, cfg := p+"-card", p+"-cfg"
	f.put(cfg, "Test.GadgetConfig", map[string]any{"power": 7})
	f.put(card, "Test.GadgetCard", map[string]any{"Id": p, "name": "چکش", "coin": 500, "config": map[string]any{"$asset": cfg}})

	dry := f.run(itemkinds.RunInput{Changed: []string{card, cfg}, DryRun: true, Actor: "test"})
	if r := rowFor(dry, card); r.Action != itemkinds.ActCreated {
		t.Fatalf("dry run row = %+v", r)
	}
	if id, _, _, _ := f.item("gadget-" + p); id != "" {
		t.Fatal("dry run wrote the item")
	}
	out := f.run(itemkinds.RunInput{Changed: []string{card, cfg}, Actor: "test"})
	if r := rowFor(out, card); r.Action != itemkinds.ActCreated || r.ItemKey != "gadget-"+p {
		t.Fatalf("create row = %+v", r)
	}
	if r := rowFor(out, cfg); r.Action != itemkinds.ActDependency || len(r.Parents) != 1 {
		t.Fatalf("dependency row = %+v", r)
	}
	id, price, active, _ := f.item("gadget-" + p)
	if id == "" || !active || price == "" {
		t.Fatalf("item not created: %q %q %v", id, price, active)
	}
	again := f.run(itemkinds.RunInput{Changed: []string{card}, Actor: "test"})
	if r := rowFor(again, card); r.Action != itemkinds.ActUnchanged {
		t.Fatalf("same hash must be unchanged: %+v", r)
	}

	plugintest.Exec(t, f.db, `UPDATE economy_item SET price=jsonb_build_object('currencyKey',$2::text,'amount',999) WHERE id=$1`, id, f.coin)
	plugintest.Exec(t, f.db, `UPDATE itemkind_profile SET data=data || '{"note":"admin"}' WHERE item_id=$1`, id)
	f.put(cfg, "Test.GadgetConfig", map[string]any{"power": 9})
	f.put(card, "Test.GadgetCard", map[string]any{"Id": p, "name": "چکش بزرگ", "coin": 700, "config": map[string]any{"$asset": cfg}})
	upd := f.run(itemkinds.RunInput{Changed: []string{cfg}, Actor: "test"})
	if r := rowFor(upd, card); r.Action != itemkinds.ActUpdated {
		t.Fatalf("dependency change must re-map the card: %+v", upd.Rows)
	}
	var amount int64
	var power float64
	var note string
	must(t, f.db.QueryRow(`SELECT (price->>'amount')::bigint FROM economy_item WHERE id=$1`, id).Scan(&amount))
	must(t, f.db.QueryRow(`SELECT (data->>'power')::float, data->>'note' FROM itemkind_profile WHERE item_id=$1`, id).Scan(&power, &note))
	if amount != 999 || power != 9 || note != "admin" {
		t.Fatalf("console must keep price and note, Unity owns power: amount=%d power=%v note=%q", amount, power, note)
	}

	ctx := context.Background()
	tx, err := f.db.BeginTx(ctx, nil)
	must(t, err)
	err = f.s.checkBuy(ctx, tx, economyhooks.ItemBuy{ItemID: id, ItemKey: "gadget-" + p})
	_ = tx.Rollback()
	if err != nil {
		t.Fatalf("buy check refused a weak gadget: %v", err)
	}
	if got := f.s.slot(ctx, id, f.itemType); got != "gadget:gadget-"+p {
		t.Fatalf("slot = %q", got)
	}
	if st, _ := f.s.itemStates(ctx, "", "fa", []string{id}); st[id].Fields != nil {
		t.Fatal("profile must not be sent unless asked")
	}
	st, err := f.s.itemStates(economyhooks.WithOptions(ctx, "profile"), "", "fa", []string{id})
	must(t, err)
	if st[id].Fields["kind"] != f.kind || st[id].Fields["profile"] == nil {
		t.Fatalf("item state = %+v", st[id])
	}

	plugintest.Exec(t, f.db, `DELETE FROM assets_entry WHERE asset_id=$1`, card)
	del := f.run(itemkinds.RunInput{Deleted: []string{card}, Actor: "test"})
	if r := rowFor(del, card); r.Action != itemkinds.ActDeactivated {
		t.Fatalf("delete row = %+v", r)
	}
	if _, _, active, _ := f.item("gadget-" + p); active {
		t.Fatal("deleted card must switch the item off")
	}
}

func TestRouterPendingAndLinkingSeededItems(t *testing.T) {
	f := newFixture(t)
	p := plugintest.Key("q")
	var typeID string
	must(t, f.db.QueryRow(`SELECT id::text FROM base_category WHERE key_type='item_type' AND key=$1`, f.itemType).Scan(&typeID))
	seeded := "seed-" + p
	plugintest.Exec(t, f.db, `INSERT INTO economy_item (id, key, type_id, name, price, metadata) VALUES (gen_random_uuid(), $1, $2, '{"fa":"قدیمی"}', jsonb_build_object('currencyKey',$3::text,'amount',10), jsonb_build_object('gameId',$4::text))`, seeded, typeID, f.coin, p)
	oldID, _, _, _ := f.item(seeded)
	card := p + "-card"
	f.put(card, "Test.GadgetCard", map[string]any{"Id": p, "name": "تازه", "coin": 40, "config": map[string]any{"$asset": p + "-missing"}})
	out := f.run(itemkinds.RunInput{Changed: []string{card}, Actor: "test"})
	if r := rowFor(out, card); r.Action != itemkinds.ActPending || r.ItemKey != seeded {
		t.Fatalf("missing ref must keep the linked item pending: %+v", r)
	}
	if _, _, active, _ := f.item(seeded); !active {
		t.Fatal("a pending link must not switch off a live seeded item")
	}
	cfg := p + "-missing"
	f.put(cfg, "Test.GadgetConfig", map[string]any{"power": 3})
	out = f.run(itemkinds.RunInput{Changed: []string{cfg}, Actor: "test"})
	r := rowFor(out, card)
	if r.Action != itemkinds.ActLinked || r.ItemKey != seeded {
		t.Fatalf("arriving ref must link the seeded item: %+v", out.Rows)
	}
	id, price, _, _ := f.item(seeded)
	if id != oldID {
		t.Fatal("linking must keep the item id")
	}
	var amount int64
	must(t, f.db.QueryRow(`SELECT (price->>'amount')::bigint FROM economy_item WHERE id=$1`, id).Scan(&amount))
	if amount != 40 {
		t.Fatalf("first link writes the Unity price, got %s", price)
	}

	all := f.run(itemkinds.RunInput{All: true, Kind: f.kind, Force: true, Actor: "test"})
	if rr := rowFor(all, card); rr.Action != itemkinds.ActUnchanged && rr.Action != itemkinds.ActUpdated {
		t.Fatalf("forced re-route row = %+v", rr)
	}
}

func TestPendingItemRetriedWhenTheItemItWaitsForArrives(t *testing.T) {
	f := newFixture(t)
	p := plugintest.Key("w")
	waiting, later, cfg := p+"-a", p+"-b", p+"-cfg"
	f.put(cfg, "Test.GadgetConfig", map[string]any{"power": 1})
	f.put(waiting, "Test.GadgetCard", map[string]any{"Id": p + "a", "name": "الف", "coin": 5, "config": map[string]any{"$asset": later}})
	out := f.run(itemkinds.RunInput{Changed: []string{waiting}, Actor: "test"})
	if r := rowFor(out, waiting); r.Action != itemkinds.ActPending {
		t.Fatalf("first push = %+v", r)
	}
	f.put(later, "Test.GadgetCard", map[string]any{"Id": p + "b", "name": "ب", "coin": 6, "power": 4, "config": map[string]any{"$asset": cfg}})
	out = f.run(itemkinds.RunInput{Changed: []string{later}, Actor: "test"})
	if r := rowFor(out, waiting); r.Action != itemkinds.ActLinked && r.Action != itemkinds.ActUpdated {
		t.Fatalf("the pending card must be retried when its reference arrives: %+v", out.Rows)
	}
	if _, _, active, _ := f.item("gadget-" + p + "a"); !active {
		t.Fatal("the retried item must be switched on")
	}
}
