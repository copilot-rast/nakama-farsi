package main

import (
	"strings"
	"testing"
)

func TestCleanReportContext(t *testing.T) {
	if v, err := cleanContext("matchId", "  abc.node1 ", 128); err != nil || v != "abc.node1" {
		t.Fatalf("%q %v", v, err)
	}
	for _, bad := range []string{strings.Repeat("x", 129), "a\nb", "a\x00b"} {
		_, err := cleanContext("matchId", bad, 128)
		if app, ok := asApp(err); !ok || app.Code != "INVALID_CONTEXT" || app.Field != "matchId" {
			t.Fatalf("%q: %v", bad, err)
		}
	}
	if len(playerReportStatuses) != 4 || validKeyType(playerReportActions, "delete") {
		t.Fatal("player reports have four statuses and no message deletion")
	}
}
