package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestExpiredClaimIsNotSentTwice(t *testing.T) {
	h := newHarness(t)
	var mu sync.Mutex
	delivered := map[string]int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		time.Sleep(250 * time.Millisecond)
		mu.Lock()
		delivered[r.Form.Get("receptor")]++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"return":{"status":200,"message":"ok"},"entries":[{"messageid":1,"status":1}]}`))
	}))
	defer srv.Close()
	ids := []string{}
	for i := 0; i < 4; i++ {
		ids = append(ids, h.user(map[string]any{"phone": fmt.Sprintf("0912000%04d", i)}, "", "fa"))
	}
	h.mustAPI(http.StatusCreated, http.MethodPost, "/providers", map[string]any{"channel": "sms", "kind": "kavenegar", "isActive": true, "isDefault": true, "sandbox": false, "config": map[string]any{"apiKey": "k", "sender": "1000"}})
	n := h.mustAPI(http.StatusCreated, http.MethodPost, "/notifications", map[string]any{"channel": "sms", "body": map[string]string{"fa": "x"}, "target": map[string]any{"mode": "users", "userIds": ids}})

	nodes := []*service{testService(), testService()}
	for _, node := range nodes {
		node.db, node.nk = h.s.db, h.nk
		node.cfg.claimLease = 400 * time.Millisecond
		node.cfg.sendTimeout = 300 * time.Millisecond
		node.baseURLs["kavenegar"] = srv.URL
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if _, err := nodes[0].runWorker(h.ctx); err != nil {
			t.Error(err)
		}
	}()
	go func() {
		defer wg.Done()
		time.Sleep(450 * time.Millisecond)
		for i := 0; i < 6; i++ {
			if _, err := nodes[1].runWorker(h.ctx); err != nil {
				t.Error(err)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	wg.Wait()
	for i := 0; i < 10; i++ {
		if _, err := nodes[1].runWorker(h.ctx); err != nil {
			t.Fatal(err)
		}
		mu.Lock()
		done := len(delivered) == 4
		mu.Unlock()
		if done {
			break
		}
		time.Sleep(450 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(delivered) != 4 {
		t.Fatalf("delivered to %d phones, want 4: %v", len(delivered), delivered)
	}
	for phone, c := range delivered {
		if c != 1 {
			t.Fatalf("%s got %d messages", phone, c)
		}
	}
	got := h.mustAPI(http.StatusOK, http.MethodGet, "/notifications/"+n["id"].(string), nil)
	if num(got["sentCount"]) != 4 || got["status"] != "done" {
		t.Fatalf("notification = %v", got)
	}
}
