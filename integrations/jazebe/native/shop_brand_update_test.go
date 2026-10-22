package main

import (
	"database/sql"
	"encoding/json"
	"testing"
)

func TestPatchShopFields(t *testing.T) {
	text := sql.NullString{String: "previous", Valid: true}
	if err := patchShopText(map[string]json.RawMessage{"shortDescription": json.RawMessage(`null`)}, "shortDescription", &text, false); err != nil || text.Valid {
		t.Fatalf("null %+v: %v", text, err)
	}
	if err := patchShopText(map[string]json.RawMessage{"name": json.RawMessage(`null`)}, "name", &text, true); err == nil {
		t.Fatal("required name accepted null")
	}
	price := sql.NullInt64{Int64: 10, Valid: true}
	if err := patchShopInt(map[string]json.RawMessage{"pricePoints": json.RawMessage(`0`)}, "pricePoints", &price, true); err == nil {
		t.Fatal("zero price accepted")
	}
	if err := patchShopInt(map[string]json.RawMessage{"stock": json.RawMessage(`null`)}, "stock", &price, false); err != nil || price.Valid {
		t.Fatalf("stock %+v: %v", price, err)
	}
}
