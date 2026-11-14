package main

import "testing"

func TestRelationComparators(t *testing.T) {
	cases := []struct {
		op      string
		cur, th float64
		want    bool
		why     string
	}{
		{"gt", 10, 9.9999, true, "gt above"},
		{"gt", 10, 10, false, "gt equal"},
		{"gte", 10, 10, true, "gte equal"},
		{"gte", 9.9999, 10, false, "gte below"},
		{"lt", -1, 0, true, "lt negative"},
		{"lt", 0, 0, false, "lt equal"},
		{"lte", 0, 0, true, "lte equal"},
		{"lte", 0.0001, 0, false, "lte above"},
		{"eq", 0.1 + 0.2, 0.3, true, "eq at 4 decimals"},
		{"eq", 5.00004, 5, true, "eq rounds the fifth decimal"},
		{"eq", 5.0001, 5, false, "eq keeps the fourth decimal"},
		{"bogus", 1, 1, false, "unknown operator never holds"},
	}
	for _, c := range cases {
		if got := relationHolds(c.op, c.cur, c.th); got != c.want {
			t.Errorf("%s: relationHolds(%s, %v, %v) = %v", c.why, c.op, c.cur, c.th, got)
		}
	}
}

func TestRelationDue(t *testing.T) {
	once := relationDef{Repeat: "once"}
	rearm := relationDef{Repeat: "rearm"}
	every := relationDef{Repeat: "every", MaxFires: 2}
	type res struct{ fire, rearm bool }
	check := func(why string, r relationDef, holds bool, fires int, armed bool, want res) {
		t.Helper()
		f, a := relationDue(r, holds, fires, armed)
		if f != want.fire || a != want.rearm {
			t.Errorf("%s: got fire=%v rearm=%v, want %+v", why, f, a, want)
		}
	}
	check("once fires the first time", once, true, 0, true, res{true, false})
	check("once never fires again", once, true, 1, false, res{false, false})
	check("once ignores a false condition", once, false, 1, false, res{false, false})
	check("rearm fires while armed", rearm, true, 3, true, res{true, false})
	check("rearm waits while disarmed", rearm, true, 3, false, res{false, false})
	check("rearm re-arms when the condition stops holding", rearm, false, 3, false, res{false, true})
	check("rearm already armed stays", rearm, false, 3, true, res{false, false})
	check("every fires each time", every, true, 1, false, res{true, false})
	check("every stops at the cap", every, true, 2, false, res{false, false})
	check("rearm respects the cap", relationDef{Repeat: "rearm", MaxFires: 1}, true, 1, true, res{false, false})
}
