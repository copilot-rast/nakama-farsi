package main

import (
	"errors"
	"testing"

	"github.com/heroiclabs/nakama/v3/integrations/inventoryx/internal/settings"
	"github.com/heroiclabs/nakama/v3/internal/localized"
	"github.com/heroiclabs/nakama/v3/internal/pluginkit"
)

func consoleCode(err error) string {
	var ce *pluginkit.ConsoleError
	if errors.As(err, &ce) {
		return ce.Code + "@" + ce.Field
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func TestCleanConfig(t *testing.T) {
	cur := map[string]bool{"coin": true, "gem": true}
	langs := localized.NewSet([]localized.Language{{Code: "fa", Active: true, Default: true}, {Code: "en", Active: true}})
	cfg := settings.Defaults()
	v := 3.5
	cases := []struct {
		name string
		in   itemConfig
		want string
	}{
		{"ok", itemConfig{UnlockPrices: []price{{"gem", 40}}, Tracks: []track{{Key: "Power", CurrencyKey: "coin", Name: localized.Text{"fa": "قدرت"}, Levels: []step{{Price: 0, Value: &v}, {Price: 10, Requires: []need{{"Range", 1}}}}}, {Key: "Range", CurrencyKey: "coin", Levels: []step{{Price: 1}}}}}, ""},
		{"unknown currency", itemConfig{UnlockPrices: []price{{"ruby", 1}}}, "CURRENCY_NOT_FOUND@unlockPrices[0]"},
		{"zero unlock", itemConfig{UnlockPrices: []price{{"gem", 0}}}, "INVALID_PRICE@unlockPrices[0]"},
		{"same currency", itemConfig{UnlockPrices: []price{{"gem", 1}, {"gem", 2}}}, "DUPLICATE_CURRENCY@unlockPrices[1]"},
		{"bad key", itemConfig{Tracks: []track{{Key: "1x", CurrencyKey: "coin", Levels: []step{{}}}}}, "INVALID_KEY@tracks[0].key"},
		{"dup key", itemConfig{Tracks: []track{{Key: "A", CurrencyKey: "coin", Levels: []step{{}}}, {Key: "A", CurrencyKey: "coin", Levels: []step{{}}}}}, "DUPLICATE_TRACK@tracks[1].key"},
		{"no levels", itemConfig{Tracks: []track{{Key: "A", CurrencyKey: "coin"}}}, "LEVELS_REQUIRED@tracks[0].levels"},
		{"negative price", itemConfig{Tracks: []track{{Key: "A", CurrencyKey: "coin", Levels: []step{{Price: -1}}}}}, "INVALID_PRICE@tracks[0].levels[0].price"},
		{"self need", itemConfig{Tracks: []track{{Key: "A", CurrencyKey: "coin", Levels: []step{{Requires: []need{{"A", 1}}}}}}}, "INVALID_REQUIREMENT@tracks[0].levels[0].requires[0].track"},
		{"need too high", itemConfig{Tracks: []track{{Key: "A", CurrencyKey: "coin", Levels: []step{{Requires: []need{{"B", 2}}}}}, {Key: "B", CurrencyKey: "coin", Levels: []step{{}}}}}, "INVALID_REQUIREMENT@tracks[0].levels[0].requires[0].level"},
		{"name needs default language", itemConfig{Tracks: []track{{Key: "A", CurrencyKey: "coin", Name: localized.Text{"en": "A"}, Levels: []step{{}}}}}, "DEFAULT_LANGUAGE_REQUIRED@tracks[0].name"},
	}
	for _, c := range cases {
		_, err := cleanConfig(cfg, langs, cur, c.in)
		if got := consoleCode(err); got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
}

func TestTrackValues(t *testing.T) {
	a, b := 12.0, 15.0
	tr := track{Base: 10, Levels: []step{{Value: &a}, {}, {Value: &b}}}
	for lv, want := range []float64{10, 12, 12, 15} {
		if got := tr.valueAt(lv); got != want {
			t.Errorf("level %d = %v want %v", lv, got, want)
		}
	}
	plain := track{Base: 7, Levels: []step{{}, {}}}
	if plain.valueAt(2) != 2 {
		t.Errorf("without values the attribute is the level")
	}
}

func TestPickPrices(t *testing.T) {
	list := []price{{"gem", 5}, {"coin", 500}}
	all := settings.Defaults()
	if got, _ := pickPrices(all, list, ""); len(got) != 2 {
		t.Fatalf("all mode takes every price: %v", got)
	}
	choice := settings.Defaults()
	choice.Unlock.PriceMode = settings.PriceChoice
	if got, err := pickPrices(choice, list, "coin"); err != nil || len(got) != 1 || got[0].CurrencyKey != "coin" {
		t.Fatalf("choice = %v %v", got, err)
	}
	if got, err := pickPrices(choice, list[:1], ""); err != nil || len(got) != 1 {
		t.Fatalf("one price needs no choice: %v %v", got, err)
	}
}

func TestSettingsValidate(t *testing.T) {
	s := settings.Defaults()
	if f, c := s.Validate(); c != "" {
		t.Fatalf("defaults invalid: %s %s", f, c)
	}
	s.Unlock.PriceMode = "any"
	if f, _ := s.Validate(); f != "unlock.priceMode" {
		t.Fatalf("field = %s", f)
	}
}
