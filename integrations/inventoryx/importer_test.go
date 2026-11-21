package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/heroiclabs/nakama/v3/integrations/inventoryx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/economyhooks"
	"github.com/heroiclabs/nakama/v3/internal/plugintest"
)

// fakeSaver stands in for economy's item editor: it writes the columns the import sets and
// refuses unknown currencies like economy does.
func fakeSaver(ctx context.Context, tx *sql.Tx, id string, create bool, raw []byte) error {
	var in struct {
		Key       string            `json:"key"`
		TypeID    string            `json:"typeId"`
		Name      map[string]string `json:"name"`
		Price     *price            `json:"price"`
		Sort      int               `json:"sort"`
		IsActive  *bool             `json:"isActive"`
		ParentKey string            `json:"parentKey"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		return err
	}
	var p any
	if in.Price != nil {
		var ok bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_currency WHERE key=$1)`, in.Price.CurrencyKey).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return &economyhooks.Refusal{Status: 400, Code: "CURRENCY_NOT_FOUND", Meta: map[string]any{"field": "price"}}
		}
		b, _ := json.Marshal(in.Price)
		p = string(b)
	}
	var parent any
	if in.ParentKey != "" {
		var ok bool
		_ = tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM economy_item WHERE key=$1)`, in.ParentKey).Scan(&ok)
		if !ok {
			return &economyhooks.Refusal{Status: 400, Code: "ITEM_NOT_FOUND", Meta: map[string]any{"field": "parentKey"}}
		}
		parent = in.ParentKey
	}
	active := in.IsActive == nil || *in.IsActive
	name, _ := json.Marshal(in.Name)
	if create {
		_, err := tx.ExecContext(ctx, `INSERT INTO economy_item (id, key, type_id, name, price, sort, is_active, parent_key) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, id, in.Key, in.TypeID, string(name), p, in.Sort, active, parent)
		return err
	}
	_, err := tx.ExecContext(ctx, `UPDATE economy_item SET type_id=$2, name=$3, price=$4, sort=$5, is_active=$6, parent_key=$7 WHERE id=$1`, id, in.TypeID, string(name), p, in.Sort, active, parent)
	return err
}

func TestImportUpsertsItemsByKey(t *testing.T) {
	f := newFixture(t)
	economyhooks.SetItemSaver(fakeSaver)
	t.Cleanup(func() { economyhooks.SetItemSaver(nil) })
	ctx := context.Background()
	var typeKey string
	_ = f.db.QueryRow(`SELECT key FROM base_category WHERE id=$1`, f.typeID).Scan(&typeKey)
	_, existing := f.item(false, 0, nil, itemConfig{})
	plugintest.Exec(t, f.db, `UPDATE economy_item SET sort=7 WHERE key=$1`, existing)
	child, parent := plugintest.Key("child"), plugintest.Key("parent")
	body := `{"items":[
		{"key":"` + child + `","type":"` + typeKey + `","name":"بچه","parentKey":"` + parent + `","prices":[{"currencyId":"` + f.coin + `","amount":50}]},
		{"key":"` + parent + `","type":"` + typeKey + `","name":{"fa":"پدر","en":"Parent"},"order":3,"upgrades":[{"key":"Power","currencyId":"` + f.coin + `","prices":[500,1000]}]},
		{"key":"` + existing + `","name":"نام تازه","unlockPrices":[{"currencyId":"` + f.gem + `","amount":40}]}
	],"dryRun":true}`
	in, err := parseImport([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	out, err := f.s.runImport(ctx, in, "t")
	if err != nil {
		t.Fatal(err)
	}
	rows := out["rows"].([]importReport)
	if rows[0].Action != "created" || rows[1].Action != "created" || rows[2].Action != "updated" || out["applied"] != false {
		t.Fatalf("dry run = %+v", out)
	}
	if f.count(`SELECT count(*) FROM economy_item WHERE key IN ($1,$2)`, child, parent) != 0 {
		t.Fatal("dry run wrote items")
	}
	in.DryRun = false
	out, err = f.s.runImport(ctx, in, "t")
	if err != nil || out["applied"] != true {
		t.Fatalf("import = %+v %v", out, err)
	}
	var name string
	var sort int
	_ = f.db.QueryRow(`SELECT name->>'fa', sort FROM economy_item WHERE key=$1`, existing).Scan(&name, &sort)
	if name != "نام تازه" || sort != 7 {
		t.Fatalf("kept fields: %q sort %d", name, sort)
	}
	if f.count(`SELECT count(*) FROM inventoryx_item x JOIN economy_item i ON i.id=x.item_id WHERE i.key=$1 AND jsonb_array_length(x.tracks)=1`, parent) != 1 {
		t.Fatal("tracks not saved")
	}
	if f.count(`SELECT count(*) FROM economy_item WHERE key=$1 AND parent_key=$2 AND price->>'amount'='50'`, child, parent) != 1 {
		t.Fatal("child not saved after its parent")
	}
	// Again: nothing changes.
	out, _ = f.s.runImport(ctx, in, "t")
	if s := out["summary"].(map[string]int); s["unchanged"] != 3 {
		t.Fatalf("again = %+v", out["rows"])
	}
	// Atomic mode: one bad row writes nothing; partial mode writes the good ones.
	bad := `[{"key":"` + existing + `","name":"x2"},{"key":"` + plugintest.Key("nope") + `","name":"x"},{"key":"` + child + `","prices":[{"currencyId":"nocur","amount":5}]}]`
	in2, _ := parseImport([]byte(bad))
	out, err = f.s.runImport(ctx, in2, "t")
	if err != nil || out["applied"] != false {
		t.Fatalf("atomic = %+v %v", out, err)
	}
	rows = out["rows"].([]importReport)
	if rows[1].Code != "ITEM_TYPE_REQUIRED" || rows[2].Code != "CURRENCY_NOT_FOUND" || rows[2].Field != "price" {
		t.Fatalf("codes = %+v", rows)
	}
	_ = f.db.QueryRow(`SELECT name->>'fa' FROM economy_item WHERE key=$1`, existing).Scan(&name)
	if name != "نام تازه" {
		t.Fatal("atomic import wrote a row")
	}
	f.settings(func(c *settings.All) { c.Import.Mode = "partial" })
	out, _ = f.s.runImport(ctx, in2, "t")
	_ = f.db.QueryRow(`SELECT name->>'fa' FROM economy_item WHERE key=$1`, existing).Scan(&name)
	if out["applied"] != true || name != "x2" {
		t.Fatalf("partial = %+v %q", out["summary"], name)
	}
	// Too many prices for one economy item.
	in3, _ := parseImport([]byte(`[{"key":"` + child + `","prices":[{"currencyId":"` + f.coin + `","amount":1},{"currencyId":"` + f.gem + `","amount":1}]}]`))
	out, _ = f.s.runImport(ctx, in3, "t")
	if r := out["rows"].([]importReport); r[0].Code != "TOO_MANY_PRICES" {
		t.Fatalf("prices = %+v", r)
	}
	_, _ = f.db.Exec(`DELETE FROM economy_item WHERE key=$1`, child)
}
