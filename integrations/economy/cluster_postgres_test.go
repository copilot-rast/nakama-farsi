package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestEconomyReconcilerOnManyNodes(t *testing.T) {
	h := newHarness(t)
	h.mustAPI(201, http.MethodPost, "/currencies", map[string]any{"key": "coin", "name": map[string]string{"fa": "سکه"}, "type": "soft"})
	prize := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "c10", "name": map[string]string{"fa": "ده"}, "items": []map[string]any{{"type": "currency", "key": "coin", "action": "add", "value": 10}}})
	product := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "gw", "isActive": true, "name": map[string]string{"fa": "بسته"}, "prizeId": prize["id"],
		"paymentMethods": []string{"zibal"}, "priceToman": 1000})
	h.mustAPI(200, http.MethodPut, "/payment-methods/zibal", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true}})

	var mu sync.Mutex
	inquiries, verifies := map[string]int{}, map[string]int{}
	var track atomic.Int64
	track.Store(time.Now().UnixNano() % 1e9)
	zibal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/request" {
			_, _ = fmt.Fprintf(w, `{"result":100,"trackId":%d}`, track.Add(1))
			return
		}
		raw := make([]byte, 512)
		n, _ := r.Body.Read(raw)
		body := string(raw[:n])
		time.Sleep(20 * time.Millisecond)
		mu.Lock()
		defer mu.Unlock()
		switch r.URL.Path {
		case "/v1/inquiry":
			inquiries[body]++
			_, _ = w.Write([]byte(`{"result":100,"status":1,"amount":10000}`))
		case "/v1/verify":
			verifies[body]++
			_, _ = w.Write([]byte(`{"result":100,"amount":10000,"refNumber":7}`))
		}
	}))
	defer zibal.Close()
	swap(t, &zibalBase, zibal.URL)

	const payments = 12
	ids := make([]string, payments)
	users := make([]string, payments)
	for i := range ids {
		users[i] = h.user()
		ids[i] = h.mustRPC(users[i], "economy_purchase_start", map[string]any{"productId": product["id"], "method": "zibal"})["paymentId"].(string)
	}
	if _, err := h.s.db.Exec(`UPDATE economy_payment SET created_at = now() - interval '10 minutes' WHERE id = ANY($1::uuid[])`, "{"+joinIDs(ids)+"}"); err != nil {
		t.Fatal(err)
	}

	const nodes = 4
	var wg sync.WaitGroup
	var checked atomic.Int64
	for i := 0; i < nodes; i++ {
		node := testService()
		node.db, node.nk, node.store = h.s.db, h.nk, h.s.store
		wg.Add(1)
		go func() {
			defer wg.Done()
			for round := 0; round < 2; round++ {
				n, err := node.reconcileBatch(h.ctx)
				if err != nil {
					t.Error(err)
				}
				checked.Add(int64(n))
			}
		}()
	}
	wg.Wait()

	if checked.Load() != payments {
		t.Fatalf("nodes checked %d payments in total, want %d", checked.Load(), payments)
	}
	for i, id := range ids {
		p := h.payment(id)
		var count int
		_ = h.s.db.QueryRow(`SELECT inquiry_count FROM economy_payment WHERE id=$1`, id).Scan(&count)
		if p.Status != "success" || !p.RewardGranted || count != 1 {
			t.Fatalf("payment %s = %s granted=%v inquiries=%d", id, p.Status, p.RewardGranted, count)
		}
		if got := h.wallet(users[i])["coin"]; got != 10 {
			t.Fatalf("user %d coin = %d", i, got)
		}
	}
	if len(inquiries) != payments || len(verifies) != payments {
		t.Fatalf("gateway saw %d tracks inquired and %d verified, want %d", len(inquiries), len(verifies), payments)
	}
	for k, n := range verifies {
		if n != 1 || inquiries[k] != 1 {
			t.Fatalf("track %s inquired %d and verified %d times", k, inquiries[k], n)
		}
	}
}

func joinIDs(ids []string) string {
	out := ""
	for i, id := range ids {
		if i > 0 {
			out += ","
		}
		out += id
	}
	return out
}
