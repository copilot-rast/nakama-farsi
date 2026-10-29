package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/gorilla/mux"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type fakeNakama struct {
	runtime.NakamaModule
	db      *sql.DB
	mu      sync.Mutex
	storage map[string]map[string]string
	calls   int
}

func (n *fakeNakama) MultiUpdate(ctx context.Context, _ []*runtime.AccountUpdate, writes []*runtime.StorageWrite, _ []*runtime.StorageDelete, wallets []*runtime.WalletUpdate, ledger bool) ([]*api.StorageObjectAck, []*runtime.WalletUpdateResult, error) {
	n.mu.Lock()
	n.calls++
	n.mu.Unlock()
	tx, err := n.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()
	for _, w := range wallets {
		var raw string
		if err := tx.QueryRowContext(ctx, `SELECT wallet::text FROM users WHERE id=$1 FOR UPDATE`, w.UserID).Scan(&raw); err != nil {
			return nil, nil, err
		}
		wallet := map[string]int64{}
		_ = json.Unmarshal([]byte(raw), &wallet)
		for k, v := range w.Changeset {
			if wallet[k]+v < 0 {
				return nil, nil, &runtime.WalletNegativeError{UserID: w.UserID, Path: k, Current: wallet[k], Amount: v}
			}
			wallet[k] += v
		}
		out, _ := json.Marshal(wallet)
		if _, err := tx.ExecContext(ctx, `UPDATE users SET wallet=$2 WHERE id=$1`, w.UserID, string(out)); err != nil {
			return nil, nil, err
		}
		if ledger {
			cs, _ := json.Marshal(w.Changeset)
			md, _ := json.Marshal(w.Metadata)
			if _, err := tx.ExecContext(ctx, `INSERT INTO wallet_ledger (id, user_id, changeset, metadata) VALUES ($1,$2,$3,$4)`, uuid.Must(uuid.NewV4()).String(), w.UserID, string(cs), string(md)); err != nil {
				return nil, nil, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, w := range writes {
		if n.storage[w.UserID] == nil {
			n.storage[w.UserID] = map[string]string{}
		}
		n.storage[w.UserID][w.Key] = w.Value
	}
	return nil, nil, nil
}

func (n *fakeNakama) WalletUpdate(ctx context.Context, userID string, changeset map[string]int64, metadata map[string]any, ledger bool) (map[string]int64, map[string]int64, error) {
	_, _, err := n.MultiUpdate(ctx, nil, nil, nil, []*runtime.WalletUpdate{{UserID: userID, Changeset: changeset, Metadata: metadata}}, ledger)
	return nil, nil, err
}

func (n *fakeNakama) AccountGetId(ctx context.Context, userID string) (*api.Account, error) {
	var wallet string
	if err := n.db.QueryRowContext(ctx, `SELECT wallet::text FROM users WHERE id=$1`, userID).Scan(&wallet); err != nil {
		return nil, err
	}
	return &api.Account{Wallet: wallet}, nil
}

func (n *fakeNakama) StorageList(_ context.Context, _, userID, _ string, _ int, _ string) ([]*api.StorageObject, string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	var out []*api.StorageObject
	for k, v := range n.storage[userID] {
		out = append(out, &api.StorageObject{Key: k, Value: v, UpdateTime: timestamppb.Now()})
	}
	return out, "", nil
}

type fakeInit struct {
	runtime.Initializer
	console *mux.Router
	public  *mux.Router
}

func (f *fakeInit) RegisterConsoleHttp(path string, h func(http.ResponseWriter, *http.Request), methods ...string) error {
	f.console.HandleFunc(path, h).Methods(methods...)
	return nil
}

func (f *fakeInit) RegisterHttp(path string, h func(http.ResponseWriter, *http.Request), methods ...string) error {
	f.public.HandleFunc(path, h).Methods(methods...)
	return nil
}

type harness struct {
	t    *testing.T
	s    *service
	nk   *fakeNakama
	init *fakeInit
	rpcs map[string]clientRPC
	ctx  context.Context
}

func newHarness(t *testing.T) *harness {
	db := openDisposableTestDB(t)
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `TRUNCATE economy_payment_log, economy_payment, economy_gift_use, economy_gift, economy_product, economy_user_prize, economy_prize,
		economy_user_meta, economy_meta, economy_user_currency_init, economy_currency, economy_tag, economy_category CASCADE`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE economy_payment_method SET is_active=false, config=''`); err != nil {
		t.Fatal(err)
	}
	s := testService()
	s.cfg.uploadDir = t.TempDir()
	nk := &fakeNakama{db: db, storage: map[string]map[string]string{}}
	s.db, s.nk = db, nk
	init := &fakeInit{console: mux.NewRouter(), public: mux.NewRouter()}
	if err := s.registerConsoleHTTP(init); err != nil {
		t.Fatal(err)
	}
	if err := s.registerPublicHTTP(init); err != nil {
		t.Fatal(err)
	}
	return &harness{t: t, s: s, nk: nk, init: init, rpcs: s.clientRPCs(), ctx: ctx}
}

func (h *harness) user() string {
	id := uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.ExecContext(h.ctx, `INSERT INTO users (id, username) VALUES ($1,$2)`, id, "u"+strings.ReplaceAll(id[:12], "-", "")); err != nil {
		h.t.Fatal(err)
	}
	return id
}

func (h *harness) api(method, path string, body any) (int, map[string]any) {
	h.t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, "/v2/console/economy"+path, reader)
	rec := httptest.NewRecorder()
	h.init.console.ServeHTTP(rec, req)
	out := map[string]any{}
	if strings.HasPrefix(strings.TrimSpace(rec.Body.String()), "{") {
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
	} else {
		var list []any
		_ = json.Unmarshal(rec.Body.Bytes(), &list)
		out["list"] = list
	}
	return rec.Code, out
}

func (h *harness) mustAPI(want int, method, path string, body any) map[string]any {
	h.t.Helper()
	code, out := h.api(method, path, body)
	if code != want {
		h.t.Fatalf("%s %s = %d %v, want %d", method, path, code, out, want)
	}
	return out
}

func (h *harness) rpc(user, name string, payload any) (map[string]any, error) {
	h.t.Helper()
	raw, _ := json.Marshal(payload)
	out, err := h.rpcs[name](h.ctx, user, string(raw))
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(out)
	m := map[string]any{}
	_ = json.Unmarshal(b, &m)
	return m, nil
}

func (h *harness) mustRPC(user, name string, payload any) map[string]any {
	h.t.Helper()
	out, err := h.rpc(user, name, payload)
	if err != nil {
		h.t.Fatalf("%s: %v", name, err)
	}
	return out
}

func (h *harness) rpcCode(user, name string, payload any) string {
	h.t.Helper()
	_, err := h.rpc(user, name, payload)
	if err == nil {
		h.t.Fatalf("%s should fail", name)
	}
	if app, ok := asApp(err); ok {
		return app.Code
	}
	return err.Error()
}

func (h *harness) wallet(user string) map[string]int64 {
	w, err := h.s.userWallet(h.ctx, user)
	if err != nil {
		h.t.Fatal(err)
	}
	return w
}

func (h *harness) payment(id string) payment {
	p, err := h.s.loadPayment(h.ctx, h.s.db, id)
	if err != nil {
		h.t.Fatal(err)
	}
	return p
}

func TestEconomyPostgresFlows(t *testing.T) {
	h := newHarness(t)

	code, _ := h.api(http.MethodPost, "/currencies", map[string]any{"key": "Coin", "name": map[string]string{"fa": "سکه"}, "type": "soft"})
	if code != 400 {
		t.Fatalf("bad key accepted: %d", code)
	}
	code, out := h.api(http.MethodPost, "/currencies", map[string]any{"key": "coin", "name": map[string]string{"en": "Coin"}, "type": "soft"})
	if code != 400 || out["error"] != "FA_REQUIRED" || out["field"] != "name" {
		t.Fatalf("fa validation = %d %v", code, out)
	}
	coin := h.mustAPI(201, http.MethodPost, "/currencies", map[string]any{"key": "coin", "name": map[string]string{"fa": "سکه", "en": "Coin"}, "type": "soft", "defaultValue": 50})
	gem := h.mustAPI(201, http.MethodPost, "/currencies", map[string]any{"key": "gem", "name": map[string]string{"fa": "الماس"}, "type": "hard"})
	if code, out := h.api(http.MethodPost, "/currencies", map[string]any{"key": "coin", "name": map[string]string{"fa": "x"}, "type": "soft"}); code != 409 || out["error"] != "ALREADY_EXISTS" {
		t.Fatalf("duplicate key = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPut, "/currencies/"+gem["id"].(string), map[string]any{"key": "gems", "name": map[string]string{"fa": "الماس"}, "type": "hard"}); code != 400 || out["error"] != "KEY_IMMUTABLE" {
		t.Fatalf("currency key change = %d %v", code, out)
	}
	updated := h.mustAPI(200, http.MethodPut, "/currencies/"+gem["id"].(string), map[string]any{"key": "gem", "name": map[string]string{"fa": "الماس", "en": "Gem"}, "type": "hard", "sort": 2})
	if updated["sort"].(float64) != 2 {
		t.Fatalf("update = %v", updated)
	}
	h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": "xp", "name": map[string]string{"fa": "تجربه"}, "defaultValue": 10})
	h.mustAPI(201, http.MethodPost, "/metas", map[string]any{"key": "secret_score", "name": map[string]string{"fa": "امتیاز"}, "serverOnly": true})
	cat := h.mustAPI(201, http.MethodPost, "/categories", map[string]any{"kind": "product", "key": "packs", "name": map[string]string{"fa": "بسته"}})
	prizeCat := h.mustAPI(201, http.MethodPost, "/categories", map[string]any{"kind": "prize", "key": "rewards", "name": map[string]string{"fa": "جایزه"}})
	tag := h.mustAPI(201, http.MethodPost, "/tags", map[string]any{"key": "hot", "name": map[string]string{"fa": "داغ"}})

	if code, out := h.api(http.MethodPost, "/prizes", map[string]any{"key": "bad", "name": map[string]string{"fa": "x"}, "items": []map[string]any{{"type": "currency", "key": "nope", "action": "add", "value": 1}}}); code != 400 || out["error"] != "CURRENCY_NOT_FOUND" {
		t.Fatalf("unknown currency in prize = %d %v", code, out)
	}
	if code, out := h.api(http.MethodPost, "/prizes", map[string]any{"key": "bad2", "categoryId": cat["id"], "name": map[string]string{"fa": "x"}, "items": []map[string]any{{"type": "currency", "key": "coin", "action": "add", "value": 1}}}); code != 400 || out["error"] != "CATEGORY_KIND_MISMATCH" {
		t.Fatalf("category kind = %d %v", code, out)
	}
	prize := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "coin100", "categoryId": prizeCat["id"], "tagIds": []string{tag["id"].(string)}, "name": map[string]string{"fa": "صد سکه"},
		"items": []map[string]any{{"type": "currency", "key": "coin", "action": "add", "value": 100}, {"type": "meta", "key": "xp", "action": "add", "value": 2.5}, {"type": "storage", "key": "badge", "action": "give", "value": map[string]any{"level": 1}}}})
	once := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "starter", "isRepeatable": false, "name": map[string]string{"fa": "شروع"},
		"items": []map[string]any{{"type": "currency", "key": "gem", "action": "add", "value": 5}}})

	if code, out := h.api(http.MethodPost, "/products", map[string]any{"key": "p0", "name": map[string]string{"fa": "x"}, "prizeId": prize["id"], "paymentMethods": []string{"zibal"}, "priceToman": 50}); code != 400 || out["error"] != "PRICE_REQUIRED" {
		t.Fatalf("product validation = %d %v", code, out)
	}
	inApp := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "coins-for-gems", "isActive": true, "name": map[string]string{"fa": "بسته"}, "prizeId": prize["id"], "categoryId": cat["id"], "tagIds": []string{tag["id"].(string)},
		"paymentMethods": []string{"in_app"}, "coinPrice": map[string]any{"currencyKey": "gem", "amount": 3}})
	gateway := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "pack-gw", "isActive": true, "name": map[string]string{"fa": "بسته درگاه"}, "prizeId": prize["id"],
		"paymentMethods": []string{"zibal", "zarinpal"}, "priceToman": 20000, "discountPercent": 10})
	store := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "pack-store", "isActive": true, "name": map[string]string{"fa": "بسته بازار"}, "prizeId": prize["id"],
		"paymentMethods": []string{"bazaar", "myket"}, "priceToman": 20000, "sku": map[string]string{"bazaar": "pack.store", "myket": "pack.store.m"}})
	onceProduct := h.mustAPI(201, http.MethodPost, "/products", map[string]any{"key": "starter-pack", "isActive": true, "name": map[string]string{"fa": "شروع"}, "prizeId": once["id"],
		"paymentMethods": []string{"in_app"}, "coinPrice": map[string]any{"currencyKey": "coin", "amount": 1}})

	list := h.mustAPI(200, http.MethodGet, "/products?q=%D8%AF%D8%B1%DA%AF%D8%A7%D9%87&limit=5", nil)
	if list["total"].(float64) != 1 || len(list["items"].([]any)) != 1 {
		t.Fatalf("search = %v", list)
	}
	if l := h.mustAPI(200, http.MethodGet, "/products?isActive=true&categoryId="+cat["id"].(string), nil); l["total"].(float64) != 1 {
		t.Fatalf("filter = %v", l)
	}
	if l := h.mustAPI(200, http.MethodGet, "/products?page=2&limit=3", nil); l["total"].(float64) != 4 || len(l["items"].([]any)) != 1 || l["page"].(float64) != 2 {
		t.Fatalf("paging = %v", l)
	}
	if code, out := h.api(http.MethodDelete, "/prizes/"+prize["id"].(string), nil); code != 409 || out["error"] != "IN_USE" {
		t.Fatalf("delete used prize = %d %v", code, out)
	}
	if code, _ := h.api(http.MethodDelete, "/currencies/"+coin["id"].(string), nil); code != 409 {
		t.Fatal("used currency deleted")
	}
	if code, _ := h.api(http.MethodDelete, "/categories/"+cat["id"].(string), nil); code != 409 {
		t.Fatal("used category deleted")
	}
	h.mustAPI(200, http.MethodDelete, "/tags/"+tag["id"].(string), nil)
	if p := h.mustAPI(200, http.MethodGet, "/products/"+inApp["id"].(string), nil); len(p["tagIds"].([]any)) != 0 {
		t.Fatalf("tag not stripped: %v", p["tagIds"])
	}
	h.mustAPI(404, http.MethodGet, "/products/"+uuid.Must(uuid.NewV4()).String(), nil)

	if code, out := h.api(http.MethodPut, "/payment-methods/zibal", map[string]any{"isActive": true, "config": map[string]any{"sandbox": false}}); code != 400 || out["error"] != "PAYMENT_METHOD_NOT_CONFIGURED" {
		t.Fatalf("activation without merchant = %d %v", code, out)
	}
	pmView := h.mustAPI(200, http.MethodPut, "/payment-methods/zarinpal", map[string]any{"isActive": true, "config": map[string]any{"merchantId": "zp-secret", "sandbox": true}})
	if pmView["config"].(map[string]any)["merchantId"] != secretMask {
		t.Fatalf("secret not masked: %v", pmView)
	}
	pmView = h.mustAPI(200, http.MethodPut, "/payment-methods/zarinpal", map[string]any{"isActive": true, "config": map[string]any{"merchantId": secretMask, "sandbox": true}})
	if pm, _ := h.s.loadMethod(h.ctx, h.s.db, "zarinpal"); pm.Config.str("merchantId") != "zp-secret" {
		t.Fatalf("mask did not keep secret: %v", pm.Config)
	}
	var sealed string
	_ = h.s.db.QueryRow(`SELECT config FROM economy_payment_method WHERE method='zarinpal'`).Scan(&sealed)
	if strings.Contains(sealed, "zp-secret") {
		t.Fatal("secret stored in plaintext")
	}
	h.mustAPI(200, http.MethodPut, "/payment-methods/zibal", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true}})
	h.mustAPI(200, http.MethodPut, "/payment-methods/bazaar", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true, "packageName": "ir.game"}})
	h.mustAPI(200, http.MethodPut, "/payment-methods/myket", map[string]any{"isActive": true, "config": map[string]any{"sandbox": true, "packageName": "ir.game"}})
	h.mustAPI(200, http.MethodPut, "/payment-methods/in_app", map[string]any{"isActive": true})
	if l := h.mustAPI(200, http.MethodGet, "/payment-methods", nil)["list"].([]any); len(l) != 5 {
		t.Fatalf("methods = %v", l)
	}

	alice, bob := h.user(), h.user()
	catalog := h.mustRPC(alice, "economy_catalog", map[string]any{"lang": "en"})
	if len(catalog["products"].([]any)) != 4 || len(catalog["currencies"].([]any)) != 2 {
		t.Fatalf("catalog = %v", catalog)
	}
	for _, raw := range catalog["products"].([]any) {
		p := raw.(map[string]any)
		if p["key"] == "pack-gw" && (p["finalToman"].(float64) != 18000 || p["name"] != "بسته درگاه") {
			t.Fatalf("gateway product = %v", p)
		}
	}
	if c := h.mustRPC(alice, "economy_catalog", map[string]any{"categoryKey": "packs"}); len(c["products"].([]any)) != 1 {
		t.Fatalf("category filter = %v", c)
	}
	if w := h.wallet(alice); w["coin"] != 50 {
		t.Fatalf("default not granted: %v", w)
	}
	h.mustRPC(alice, "economy_balances", nil)
	if w := h.wallet(alice); w["coin"] != 50 {
		t.Fatalf("default granted twice: %v", w)
	}

	if c := h.rpcCode(alice, "economy_purchase_start", map[string]any{"productId": inApp["id"], "method": "in_app"}); c != "INSUFFICIENT_BALANCE" {
		t.Fatalf("insufficient = %s", c)
	}
	if w := h.wallet(alice); w["coin"] != 50 || w["gem"] != 0 {
		t.Fatalf("wallet changed on refusal: %v", w)
	}
	h.mustAPI(200, http.MethodPost, "/users/"+alice+"/grant", map[string]any{"prizeId": once["id"]})
	if w := h.wallet(alice); w["gem"] != 5 {
		t.Fatalf("admin grant: %v", w)
	}
	if code, out := h.api(http.MethodPost, "/users/"+alice+"/grant", map[string]any{"prizeId": once["id"]}); code != 409 || out["error"] != "PRIZE_ALREADY_CLAIMED" {
		t.Fatalf("non-repeatable twice = %d %v", code, out)
	}
	if c := h.rpcCode(alice, "economy_purchase_start", map[string]any{"productId": onceProduct["id"], "method": "in_app"}); c != "PRIZE_ALREADY_CLAIMED" {
		t.Fatalf("owned product = %s", c)
	}
	res := h.mustRPC(alice, "economy_purchase_start", map[string]any{"productId": inApp["id"], "method": "in_app"})
	if res["status"] != "success" || len(res["rewards"].([]any)) != 3 {
		t.Fatalf("in_app = %v", res)
	}
	if w := h.wallet(alice); w["gem"] != 2 || w["coin"] != 150 {
		t.Fatalf("in_app wallet: %v", w)
	}
	bal := h.mustRPC(alice, "economy_balances", nil)
	metas := bal["metas"].(map[string]any)
	if metas["xp"] != 12.5 || metas["secret_score"] != nil || len(bal["items"].([]any)) != 1 {
		t.Fatalf("balances = %v", bal)
	}
	inAppPay := h.payment(res["paymentId"].(string))
	if inAppPay.Status != "success" || !inAppPay.RewardGranted {
		t.Fatalf("in_app payment = %+v", inAppPay)
	}

	h.mustAPI(201, http.MethodPost, "/gifts", map[string]any{"name": "welcome", "type": "reward", "code": "welcome-1", "prizeId": prize["id"], "perUserLimit": 1, "maxTotalUses": 1})
	if code, out := h.api(http.MethodPost, "/gifts", map[string]any{"name": "dup", "type": "reward", "code": "WELCOME-1", "prizeId": prize["id"]}); code != 409 || out["field"] != "code" {
		t.Fatalf("duplicate code = %d %v", code, out)
	}
	red := h.mustRPC(alice, "economy_gift_redeem", map[string]any{"code": "welcome-1"})
	if len(red["rewards"].([]any)) != 3 || h.wallet(alice)["coin"] != 250 {
		t.Fatalf("redeem = %v %v", red, h.wallet(alice))
	}
	if c := h.rpcCode(alice, "economy_gift_redeem", map[string]any{"code": "WELCOME-1"}); c != "GIFT_ALREADY_USED" {
		t.Fatalf("second redeem = %s", c)
	}
	if c := h.rpcCode(bob, "economy_gift_redeem", map[string]any{"code": "WELCOME-1"}); c != "GIFT_EXHAUSTED" {
		t.Fatalf("exhausted = %s", c)
	}
	if c := h.rpcCode(bob, "economy_gift_redeem", map[string]any{"code": "NOPE-1"}); c != "GIFT_NOT_FOUND" {
		t.Fatalf("unknown code = %s", c)
	}
	h.mustAPI(201, http.MethodPost, "/gifts", map[string]any{"name": "half", "type": "discount", "code": "HALF", "discountPercent": 50, "perUserLimit": 1})
	if c := h.rpcCode(bob, "economy_gift_redeem", map[string]any{"code": "HALF"}); c != "GIFT_NOT_REDEEMABLE" {
		t.Fatalf("discount redeem = %s", c)
	}
	if c := h.rpcCode(bob, "economy_purchase_start", map[string]any{"productId": store["id"], "method": "bazaar", "giftCode": "HALF"}); c != "GIFT_NOT_APPLICABLE" {
		t.Fatalf("store gift = %s", c)
	}

	var zibalReply struct {
		sync.Mutex
		inquiry, verify string
	}
	zibal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zibalReply.Lock()
		defer zibalReply.Unlock()
		switch r.URL.Path {
		case "/v1/request":
			_, _ = w.Write([]byte(`{"result":100,"trackId":` + fmt.Sprint(time.Now().UnixNano()%1e12) + `}`))
		case "/v1/inquiry":
			_, _ = w.Write([]byte(zibalReply.inquiry))
		case "/v1/verify":
			_, _ = w.Write([]byte(zibalReply.verify))
		}
	}))
	defer zibal.Close()
	swap(t, &zibalBase, zibal.URL)

	start := h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": gateway["id"], "method": "zibal", "giftCode": "half"})
	payID := start["paymentId"].(string)
	p := h.payment(payID)
	if start["status"] != "redirect" || !strings.HasPrefix(start["redirectUrl"].(string), zibal.URL+"/start/") || p.FinalToman != 9000 || p.AmountRial != 90000 || !p.TrackID.Valid {
		t.Fatalf("zibal start = %v %+v", start, p)
	}
	if c := h.rpcCode(bob, "economy_purchase_start", map[string]any{"productId": gateway["id"], "method": "zibal", "giftCode": "HALF"}); c != "GIFT_ALREADY_USED" {
		t.Fatalf("reserved code reused = %s", c)
	}
	callback := func(ticket string, q url.Values) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/v2/economy/payments/callback/"+ticket+"?"+q.Encode(), nil)
		rec := httptest.NewRecorder()
		h.init.public.ServeHTTP(rec, req)
		return rec
	}
	if rec := callback("forged.ticket", url.Values{"trackId": {p.TrackID.String}, "success": {"1"}}); rec.Code != 400 {
		t.Fatalf("forged ticket = %d", rec.Code)
	}
	otherTicket := signTicket([]byte("not-the-economy-key-not-the-key!"), payID, time.Now().Add(time.Hour))
	if rec := callback(otherTicket, url.Values{"trackId": {p.TrackID.String}, "success": {"1"}}); rec.Code != 400 {
		t.Fatalf("foreign-key ticket = %d", rec.Code)
	}
	ticket := signTicket(h.s.cfg.ticketKey, payID, time.Now().Add(time.Hour))
	zibalReply.verify = `{"result":100,"amount":90000,"refNumber":5555,"cardNumber":"6037****"}`
	rec := callback(ticket, url.Values{"trackId": {"999"}, "success": {"1"}})
	if rec.Code != 302 || !strings.Contains(rec.Header().Get("Location"), "success=false") || h.payment(payID).Status != "pending" {
		t.Fatalf("mismatched track = %d %s %s", rec.Code, rec.Header().Get("Location"), h.payment(payID).Status)
	}
	coinBefore := h.wallet(bob)["coin"]
	rec = callback(ticket, url.Values{"trackId": {p.TrackID.String}, "success": {"1"}})
	if rec.Code != 302 || rec.Header().Get("Location") != "nakama://payment?success=true&paymentId="+payID || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("paid callback = %d %v", rec.Code, rec.Header())
	}
	p = h.payment(payID)
	if p.Status != "success" || !p.RewardGranted || h.wallet(bob)["coin"] != coinBefore+100 {
		t.Fatalf("paid payment = %+v wallet %v", p, h.wallet(bob))
	}
	var used string
	_ = h.s.db.QueryRow(`SELECT status FROM economy_gift_use WHERE payment_id=$1`, payID).Scan(&used)
	if used != "used" {
		t.Fatalf("gift use = %s", used)
	}
	callback(ticket, url.Values{"trackId": {p.TrackID.String}, "success": {"1"}})
	if h.wallet(bob)["coin"] != coinBefore+100 {
		t.Fatal("second callback granted again")
	}
	st := h.mustRPC(bob, "economy_purchase_status", map[string]any{"paymentId": payID})
	if st["status"] != "success" || st["rewardGranted"] != true {
		t.Fatalf("status = %v", st)
	}
	if c := h.rpcCode(alice, "economy_purchase_status", map[string]any{"paymentId": payID}); c != "PAYMENT_NOT_FOUND" {
		t.Fatalf("foreign status = %s", c)
	}

	start = h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": gateway["id"], "method": "zibal"})
	mismatchID := start["paymentId"].(string)
	zibalReply.verify = `{"result":100,"amount":100}`
	callback(signTicket(h.s.cfg.ticketKey, mismatchID, time.Now().Add(time.Hour)), url.Values{"trackId": {h.payment(mismatchID).TrackID.String}, "success": {"1"}})
	if p := h.payment(mismatchID); p.Status != "failed" || p.Error.String != "AMOUNT_MISMATCH" || p.RewardGranted {
		t.Fatalf("amount mismatch = %+v", p)
	}

	start = h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": gateway["id"], "method": "zibal"})
	missingID := start["paymentId"].(string)
	zibalReply.verify = `{"result":100}`
	zibalReply.inquiry = `{"result":100,"status":-1}`
	callback(signTicket(h.s.cfg.ticketKey, missingID, time.Now().Add(time.Hour)), url.Values{"trackId": {h.payment(missingID).TrackID.String}, "success": {"1"}})
	if p := h.payment(missingID); p.Status != "pending" || p.Error.String != "AMOUNT_UNKNOWN" {
		t.Fatalf("missing amount = %+v", p)
	}
	_, _ = h.s.db.Exec(`UPDATE economy_payment SET created_at=now()-interval '10 minutes' WHERE id=$1`, missingID)
	zibalReply.inquiry = `{"result":100,"status":1,"amount":180000}`
	zibalReply.verify = `{"result":201,"amount":180000,"refNumber":42}`
	if n, err := h.s.reconcileBatch(h.ctx); err != nil || n < 1 {
		t.Fatalf("reconcile = %d %v", n, err)
	}
	if p := h.payment(missingID); p.Status != "success" || !p.RewardGranted {
		t.Fatalf("reconciled = %+v", p)
	}

	h.mustAPI(201, http.MethodPost, "/gifts", map[string]any{"name": "half2", "type": "discount", "code": "HALF2", "discountPercent": 50, "perUserLimit": 1, "maxTotalUses": 1})
	start = h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": gateway["id"], "method": "zibal", "giftCode": "HALF2"})
	if c := h.rpcCode(alice, "economy_purchase_start", map[string]any{"productId": gateway["id"], "method": "zibal", "giftCode": "HALF2"}); c != "GIFT_EXHAUSTED" {
		t.Fatalf("reserved total = %s", c)
	}
	if _, err := h.rpc(alice, "economy_purchase_start", map[string]any{"productId": gateway["id"], "method": "zibal", "giftCode": "HALF"}); err != nil {
		t.Fatalf("other user code: %v", err)
	}
	stale := start["paymentId"].(string)
	_, _ = h.s.db.Exec(`UPDATE economy_payment SET created_at=now()-interval '2 hours' WHERE id=$1`, stale)
	zibalReply.inquiry = `{"result":100,"status":-1}`
	if _, err := h.s.reconcileOne(h.ctx, stale); err != nil {
		t.Fatal(err)
	}
	var released string
	_ = h.s.db.QueryRow(`SELECT status FROM economy_gift_use WHERE payment_id=$1`, stale).Scan(&released)
	if p := h.payment(stale); p.Status != "expired" || released != "released" {
		t.Fatalf("expired = %+v gift %s", p, released)
	}

	storeStart := h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": store["id"], "method": "bazaar"})
	sp := storeStart["paymentId"].(string)
	if storeStart["status"] != "store" || storeStart["sku"] != "pack.store" || storeStart["developerPayload"] != sp {
		t.Fatalf("store start = %v", storeStart)
	}
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": sp, "token": "sandbox-" + uuid.Must(uuid.NewV4()).String() + "-x"}); c != "DEVELOPER_PAYLOAD_MISMATCH" {
		t.Fatalf("payload mismatch = %s", c)
	}
	if h.payment(sp).Status != "pending" {
		t.Fatal("mismatch changed state")
	}
	token := "sandbox-" + sp + "-abc"
	coinBefore = h.wallet(bob)["coin"]
	sub := h.mustRPC(bob, "economy_purchase_submit", map[string]any{"paymentId": sp, "token": token})
	if sub["status"] != "success" || h.wallet(bob)["coin"] != coinBefore+100 {
		t.Fatalf("submit = %v", sub)
	}
	again := h.mustRPC(bob, "economy_purchase_submit", map[string]any{"paymentId": sp, "token": token})
	if again["status"] != "success" || h.wallet(bob)["coin"] != coinBefore+100 {
		t.Fatalf("resubmit = %v", again)
	}
	second := h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": store["id"], "method": "bazaar"})["paymentId"].(string)
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": second, "token": token}); c != "DEVELOPER_PAYLOAD_MISMATCH" {
		t.Fatalf("token for other payment = %s", c)
	}
	if _, err := h.s.db.Exec(`UPDATE economy_payment SET store_token=NULL WHERE id=$1`, second); err != nil {
		t.Fatal(err)
	}
	_, err := h.s.markSuccess(h.ctx, h.payment(second), successFields{StoreToken: token, Outcome: "paid"})
	if _, unique := isUniqueViolation(err); !unique {
		t.Fatalf("duplicate token must hit the unique index: %v", err)
	}
	if c := h.rpcCode(alice, "economy_purchase_review", map[string]any{"method": "bazaar", "sku": "pack.store", "token": token}); c != "TOKEN_ALREADY_USED" {
		t.Fatalf("other user review = %s", c)
	}
	rv := h.mustRPC(bob, "economy_purchase_review", map[string]any{"method": "bazaar", "sku": "pack.store", "token": token})
	if rv["status"] != "success" || rv["paymentId"] != sp {
		t.Fatalf("review existing = %v", rv)
	}
	lost := h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": store["id"], "method": "myket"})["paymentId"].(string)
	rv = h.mustRPC(bob, "economy_purchase_review", map[string]any{"method": "myket", "sku": "pack.store.m", "token": "sandbox-" + lost + "-r"})
	if rv["status"] != "success" || rv["paymentId"] != lost {
		t.Fatalf("review lost = %v", rv)
	}
	if c := h.rpcCode(bob, "economy_purchase_submit", map[string]any{"paymentId": second, "token": "real-token"}); c != "PURCHASE_INVALID" {
		t.Fatalf("invalid sandbox token = %s", c)
	}
	if h.payment(second).Status != "failed" {
		t.Fatal("rejected token must fail the payment")
	}
	sc := h.mustRPC(bob, "economy_store_config", map[string]any{"method": "bazaar"})
	if sc["packageName"] != "ir.game" || sc["sandbox"] != true {
		t.Fatalf("store config = %v", sc)
	}

	mine := h.mustRPC(bob, "economy_payments_mine", map[string]any{"limit": 3})
	if len(mine["items"].([]any)) != 3 || mine["cursor"] == "" {
		t.Fatalf("mine = %v", mine)
	}
	seen := 3
	cursor := mine["cursor"].(string)
	for cursor != "" {
		page := h.mustRPC(bob, "economy_payments_mine", map[string]any{"limit": 3, "cursor": cursor})
		seen += len(page["items"].([]any))
		cursor, _ = page["cursor"].(string)
	}
	var bobCount int
	_ = h.s.db.QueryRow(`SELECT count(*) FROM economy_payment WHERE user_id=$1`, bob).Scan(&bobCount)
	if seen != bobCount {
		t.Fatalf("paged %d of %d", seen, bobCount)
	}

	free := h.mustAPI(201, http.MethodPost, "/gifts", map[string]any{"name": "full", "type": "discount", "code": "FREE100", "discountPercent": 100})
	_ = free
	fr := h.mustRPC(alice, "economy_purchase_start", map[string]any{"productId": gateway["id"], "method": "zarinpal", "giftCode": "FREE100"})
	if fr["status"] != "success" || h.payment(fr["paymentId"].(string)).FinalToman != 0 {
		t.Fatalf("free = %v", fr)
	}

	_, _ = h.s.db.Exec(`UPDATE economy_payment SET reward_granted=false WHERE id=$1`, payID)
	if code, out := h.api(http.MethodPost, "/payments/"+payID+"/retry-grant", nil); code != 200 || out["rewardGranted"] != true {
		t.Fatalf("retry grant = %d %v", code, out)
	}
	if h.wallet(bob)["coin"] < 100 {
		t.Fatal("wallet")
	}
	detail := h.mustAPI(200, http.MethodGet, "/payments/"+payID, nil)
	if len(detail["logs"].([]any)) < 2 || detail["product"] == nil || detail["user"] == nil {
		t.Fatalf("detail = %v", detail)
	}
	logs, _ := json.Marshal(detail["logs"])
	if strings.Contains(string(logs), ticket) {
		t.Fatal("ticket stored in logs")
	}
	pl := h.mustAPI(200, http.MethodGet, "/payments?status=success&userId="+bob, nil)
	if pl["total"].(float64) < 3 || pl["items"].([]any)[0].(map[string]any)["username"] == nil {
		t.Fatalf("payments list = %v", pl)
	}
	ug := h.mustAPI(200, http.MethodGet, "/users/"+alice, nil)
	if ug["metas"].(map[string]any)["secret_score"] == nil || len(ug["prizes"].([]any)) < 3 {
		t.Fatalf("user = %v", ug)
	}
	sum := h.mustAPI(200, http.MethodGet, "/summary", nil)
	if sum["counts"].(map[string]any)["products"].(float64) != 4 {
		t.Fatalf("summary = %v", sum)
	}
	pending := h.mustRPC(bob, "economy_purchase_start", map[string]any{"productId": gateway["id"], "method": "zibal"})["paymentId"].(string)
	zibalReply.inquiry = `{"result":100,"status":3}`
	if out := h.mustAPI(200, http.MethodPost, "/payments/"+pending+"/inquire", nil); out["status"] != "failed" {
		t.Fatalf("inquire = %v", out)
	}
	if code, _ := h.api(http.MethodPost, "/payments/"+pending+"/inquire", nil); code != 409 {
		t.Fatal("inquire on closed payment")
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "a.png")
	_, _ = fw.Write(append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 64)...))
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/v2/console/economy/uploads", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	urec := httptest.NewRecorder()
	h.init.console.ServeHTTP(urec, req)
	var up map[string]string
	_ = json.Unmarshal(urec.Body.Bytes(), &up)
	if urec.Code != 201 || !strings.HasSuffix(up["url"], ".png") {
		t.Fatalf("upload = %d %s", urec.Code, urec.Body.String())
	}
	frec := httptest.NewRecorder()
	h.init.public.ServeHTTP(frec, httptest.NewRequest(http.MethodGet, "/v2/economy/files/"+up["name"], nil))
	if frec.Code != 200 || frec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("file = %d %v", frec.Code, frec.Header())
	}
	frec = httptest.NewRecorder()
	h.init.public.ServeHTTP(frec, httptest.NewRequest(http.MethodGet, "/v2/economy/files/..%2f..%2fetc%2fpasswd", nil))
	if frec.Code == 200 {
		t.Fatalf("traversal = %d", frec.Code)
	}
	frec = httptest.NewRecorder()
	h.init.public.ServeHTTP(frec, httptest.NewRequest(http.MethodGet, "/v2/economy/files/passwd", nil))
	if frec.Code != 404 {
		t.Fatalf("unknown name = %d", frec.Code)
	}
}

func TestEconomyPostgresConcurrentGrant(t *testing.T) {
	h := newHarness(t)
	h.mustAPI(201, http.MethodPost, "/currencies", map[string]any{"key": "coin", "name": map[string]string{"fa": "سکه"}, "type": "soft"})
	prize := h.mustAPI(201, http.MethodPost, "/prizes", map[string]any{"key": "p", "name": map[string]string{"fa": "x"}, "items": []map[string]any{{"type": "currency", "key": "coin", "action": "add", "value": 7}}})
	user := h.user()
	source := uuid.Must(uuid.NewV4()).String()
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := h.s.grant(h.ctx, grantRequest{UserID: user, PrizeID: prize["id"].(string), SourceType: "admin", SourceID: source, Reason: "test"})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
	if w := h.wallet(user); w["coin"] != 7 {
		t.Fatalf("concurrent grant applied %d", w["coin"])
	}

	if _, err := h.s.db.Exec(`INSERT INTO economy_user_prize (id,user_id,prize_id,source_type,source_id) VALUES ($1,$2,$3,'admin','crash')`, uuid.Must(uuid.NewV4()).String(), user, prize["id"]); err != nil {
		t.Fatal(err)
	}
	var rowID string
	_ = h.s.db.QueryRow(`SELECT id::text FROM economy_user_prize WHERE source_id='crash'`).Scan(&rowID)
	if _, err := h.s.db.Exec(`UPDATE users SET wallet = jsonb_set(wallet, '{coin}', to_jsonb((wallet->>'coin')::bigint + 7)) WHERE id=$1`, user); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.Exec(`INSERT INTO wallet_ledger (id,user_id,changeset,metadata) VALUES ($1,$2,'{"coin":7}',$3)`, uuid.Must(uuid.NewV4()).String(), user, `{"grantId":"`+rowID+`"}`); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.grant(h.ctx, grantRequest{UserID: user, PrizeID: prize["id"].(string), SourceType: "admin", SourceID: "crash", Reason: "test"}); err != nil {
		t.Fatal(err)
	}
	if w := h.wallet(user); w["coin"] != 14 {
		t.Fatalf("retry after crash applied the wallet twice: %d", w["coin"])
	}
}
