package gamecatalog

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

const testSchema = `{"type":"object","additionalProperties":false,"required":["roundSeconds"],"properties":{
	"roundSeconds":{"type":"integer","minimum":10,"maximum":120,"default":30,"title":{"fa":"زمان دور","en":"Round time"}},
	"mode":{"type":"string","enum":["easy","hard"],"default":"easy"},
	"rush":{"type":"boolean"},
	"label":{"type":"string","maxLength":4}}}`

func TestRegisterValidatesKeyAndSchema(t *testing.T) {
	defer Unregister("tst_game")
	if err := Register(Definition{Key: "Bad Key"}); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("err=%v", err)
	}
	if err := Register(Definition{Key: "tst_game", Schema: json.RawMessage(`{"type":"array"}`)}); err == nil {
		t.Fatal("array schema must be refused")
	}
	if err := Register(Definition{Key: "tst_game", Schema: json.RawMessage(`{"type":"object","properties":{"a":{"type":"integer","minimum":5,"default":1}}}`)}); err == nil {
		t.Fatal("default below minimum must be refused")
	}
	if err := Register(Definition{Key: "tst_game", Realtime: true, Schema: json.RawMessage(testSchema)}); err != nil {
		t.Fatal(err)
	}
	d, ok := Lookup("tst_game")
	if !ok || !d.Realtime {
		t.Fatalf("lookup %v %v", d, ok)
	}
	found := false
	for _, r := range Registered() {
		found = found || r.Key == "tst_game"
	}
	if !found {
		t.Fatal("registered list misses the game")
	}
}

func TestSchemaApplyFillsDefaultsAndRefusesBadValues(t *testing.T) {
	s, err := ParseSchema(json.RawMessage(testSchema))
	if err != nil {
		t.Fatal(err)
	}
	if d := s.Defaults(); d["roundSeconds"] != float64(30) || d["mode"] != "easy" {
		t.Fatalf("defaults %v", d)
	}
	out, fe := s.Apply(map[string]any{"roundSeconds": float64(45), "rush": true})
	if fe != nil || out["roundSeconds"] != float64(45) || out["rush"] != true || out["mode"] != "easy" {
		t.Fatalf("out=%v fe=%v", out, fe)
	}
	for _, tc := range []struct {
		in   map[string]any
		code string
	}{
		{map[string]any{"roundSeconds": float64(5)}, "INVALID_SETTING"},
		{map[string]any{"roundSeconds": 12.5}, "INVALID_SETTING"},
		{map[string]any{"mode": "wild"}, "INVALID_SETTING"},
		{map[string]any{"rush": "yes"}, "INVALID_SETTING"},
		{map[string]any{"label": "toolong"}, "INVALID_SETTING"},
		{map[string]any{"other": 1.0}, "UNKNOWN_SETTING"},
	} {
		if _, fe := s.Apply(tc.in); fe == nil || fe.Code != tc.code {
			t.Fatalf("%v: %v", tc.in, fe)
		}
	}
	strict, _ := ParseSchema(json.RawMessage(`{"type":"object","required":["a"],"properties":{"a":{"type":"number"}}}`))
	if _, fe := strict.Apply(map[string]any{}); fe == nil || fe.Code != "SETTING_REQUIRED" {
		t.Fatalf("fe=%v", fe)
	}
}

func TestHelpersNeedTheProvider(t *testing.T) {
	SetProvider(nil)
	if _, err := Settings(context.Background(), "x"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err=%v", err)
	}
	if _, err := Enter(context.Background(), Entry{}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err=%v", err)
	}
	g := &Game{Rooms: []Room{{Key: "bronze"}}, Settings: map[string]any{"n": float64(7), "b": true, "s": "x"}}
	if r, ok := g.Room("bronze"); !ok || r.Key != "bronze" {
		t.Fatal("room lookup")
	}
	if g.Int("n", 1) != 7 || g.Int("m", 3) != 3 || !g.Bool("b", false) || g.String("s", "") != "x" {
		t.Fatal("setting readers")
	}
}

func TestRequirementsPathsAndFlags(t *testing.T) {
	var none *Requirements
	if !none.Empty() || none.AppliesTo(PathBot) || !none.Inherits() || none.Consumes() {
		t.Fatal("a nil set asks for nothing and inherits")
	}
	r := &Requirements{ItemIDs: []string{"a"}}
	if r.Empty() || !r.AppliesTo("") || !r.AppliesTo(PathFriend) || r.Consumes() {
		t.Fatal("a set without paths applies everywhere and only owning is not consuming")
	}
	r.Paths = []string{PathBot}
	if r.AppliesTo("") || r.AppliesTo(PathMatchmaking) || !r.AppliesTo(PathBot) {
		t.Fatal("paths limit where the set applies")
	}
	r.ConsumeItems = true
	if !r.Consumes() {
		t.Fatal("consume items")
	}
	off := false
	r2 := &Requirements{Metas: []MetaCondition{{MetaID: "m", Operator: "gte", Value: 1, Consume: 1}}, Inherit: &off}
	if !r2.Consumes() || r2.Inherits() {
		t.Fatal("ticket meta consumes and inherit off is kept")
	}
	if (&Requirements{Paths: []string{PathBot}}).AppliesTo(PathBot) {
		t.Fatal("an empty set never applies")
	}
	for _, p := range []string{"", PathMatchmaking, PathBot, PathFriend} {
		if !ValidPath(p) {
			t.Fatal(p)
		}
	}
	if ValidPath("arena") {
		t.Fatal("unknown path")
	}
}

type checkProvider struct{ Provider }

func (checkProvider) Check(context.Context, Entry) error {
	return &Error{Code: "REQUIREMENTS_NOT_MET"}
}

func TestCheckUsesOptionalChecker(t *testing.T) {
	SetProvider(nil)
	if err := Check(context.Background(), Entry{}); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
	SetProvider(checkProvider{})
	defer SetProvider(nil)
	if Code(Check(context.Background(), Entry{Game: "g"})) != "REQUIREMENTS_NOT_MET" {
		t.Fatal("checker not used")
	}
}
