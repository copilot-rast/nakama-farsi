package main

import "testing"

func TestNotificationIDScopedToRecipient(t *testing.T) {
	a := jazebeNotificationID("user-a", "lottery:draw-1")
	b := jazebeNotificationID("user-b", "lottery:draw-1")
	if a == b {
		t.Fatal("different recipients share one notification ID")
	}
	if a != jazebeNotificationID("user-a", "lottery:draw-1") {
		t.Fatal("notification retry is not stable")
	}
}
