package main

import (
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestConsoleModulesClean(t *testing.T) {
	v, err := consoleModules{Disabled: []string{" quiz ", "cups", "quiz", ""}}.clean()
	if err != nil || !reflect.DeepEqual(v.Disabled, []string{"cups", "quiz"}) {
		t.Fatalf("ids must be trimmed, deduped and sorted, got %v %v", v.Disabled, err)
	}
	for _, bad := range []string{"Cups", "cups!", "1quiz", strings.Repeat("a", 49)} {
		if _, err := (consoleModules{Disabled: []string{bad}}).clean(); err == nil {
			t.Fatalf("%q must be rejected", bad)
		}
	}
	many := make([]string, 0, maxConsoleModules+1)
	for i := 0; i <= maxConsoleModules; i++ {
		many = append(many, "m"+strings.Repeat("a", i%40)+string(rune('a'+i%26))+string(rune('a'+i/26)))
	}
	if _, err := (consoleModules{Disabled: many}).clean(); err == nil {
		t.Fatal("more than the limit must be rejected")
	}
	if d := defaultConsoleModules(); d.Disabled == nil || len(d.Disabled) != 0 {
		t.Fatal("default must be an empty list: every module on")
	}
}

func TestConsoleModulesRoundTrip(t *testing.T) {
	h := newHarness(t)
	if _, err := h.s.db.ExecContext(h.ctx, `DELETE FROM base_console_modules`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = h.s.db.Exec(`DELETE FROM base_console_modules`)
		consoleModulesStore.Invalidate()
	})
	got := h.mustAPI(http.StatusOK, http.MethodGet, "/console-modules", nil)
	if list, ok := got["disabled"].([]any); !ok || len(list) != 0 {
		t.Fatalf("fresh install must have every module on, got %v", got)
	}
	got = h.mustAPI(http.StatusOK, http.MethodPut, "/console-modules", map[string]any{"disabled": []string{"quiz", "cups", "cups"}, "updatedBy": "x"})
	if !reflect.DeepEqual(got["disabled"], []any{"cups", "quiz"}) {
		t.Fatalf("saved list = %v", got["disabled"])
	}
	got = h.mustAPI(http.StatusOK, http.MethodGet, "/console-modules", nil)
	if !reflect.DeepEqual(got["disabled"], []any{"cups", "quiz"}) {
		t.Fatalf("read back = %v", got)
	}
	h.expectError(http.StatusBadRequest, "INVALID_MODULE", http.MethodPut, "/console-modules", map[string]any{"disabled": []string{"Bad Id"}})
	h.expectError(http.StatusBadRequest, "INVALID_JSON", http.MethodPut, "/console-modules", map[string]any{"disabled": "cups"})
	got = h.mustAPI(http.StatusOK, http.MethodPut, "/console-modules", map[string]any{})
	if list, ok := got["disabled"].([]any); !ok || len(list) != 0 {
		t.Fatalf("an empty body turns every module back on, got %v", got)
	}
}
