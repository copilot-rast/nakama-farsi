package main

import "testing"

func TestLocalStoreURLsOnlyLoopback(t *testing.T) {
	saved := []string{stripeAPI, googlePlayBase, appleAPIProduction, appleAPISandbox, appleReceiptProduction, appleReceiptSandbox}
	defer func() {
		stripeAPI, googlePlayBase, appleAPIProduction, appleAPISandbox, appleReceiptProduction, appleReceiptSandbox = saved[0], saved[1], saved[2], saved[3], saved[4], saved[5]
	}()
	if err := applyLocalStoreURLs(map[string]string{"ECONOMY_LOCAL_STRIPE_URL": "https://api.example.com"}); err == nil {
		t.Fatal("a public host must be refused")
	}
	if err := applyLocalStoreURLs(map[string]string{}); err != nil || stripeAPI != saved[0] {
		t.Fatalf("no env changes nothing: %v %s", err, stripeAPI)
	}
	err := applyLocalStoreURLs(map[string]string{"ECONOMY_LOCAL_STRIPE_URL": "http://127.0.0.1:7599/", "ECONOMY_LOCAL_GOOGLE_PLAY_URL": "http://localhost:7599", "ECONOMY_LOCAL_APPLE_URL": "http://127.0.0.1:7599"})
	if err != nil {
		t.Fatal(err)
	}
	if stripeAPI != "http://127.0.0.1:7599" || googlePlayBase != "http://localhost:7599/androidpublisher/v3" || appleReceiptSandbox != "http://127.0.0.1:7599/verifyReceipt" || appleAPISandbox != "http://127.0.0.1:7599" {
		t.Fatalf("urls: %s %s %s %s", stripeAPI, googlePlayBase, appleReceiptSandbox, appleAPISandbox)
	}
}
