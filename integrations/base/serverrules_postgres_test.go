package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/heroiclabs/nakama/v3/internal/rpcgate"
	"google.golang.org/grpc/codes"
)

func (n *fakeNakama) AccountDeleteId(_ context.Context, userID string, _ bool) error {
	_, err := n.db.Exec(`DELETE FROM users WHERE id=$1`, userID)
	return err
}

func resetServerRules(h *harness) {
	h.t.Helper()
	for _, q := range []string{`DELETE FROM base_server_rules`, `DELETE FROM base_account_deletion`} {
		if _, err := h.s.db.Exec(q); err != nil {
			h.t.Fatal(err)
		}
	}
	serverRulesStore.Invalidate()
	h.t.Cleanup(func() {
		_, _ = h.s.db.Exec(`DELETE FROM base_server_rules`)
		_, _ = h.s.db.Exec(`DELETE FROM base_account_deletion`)
		serverRulesStore.Invalidate()
		rpcgate.Set(nil)
	})
}

func (h *harness) rules(patch map[string]any) map[string]any {
	h.t.Helper()
	body := map[string]any{
		"maintenance":   map[string]any{"enabled": false, "message": map[string]string{}, "allowRpcs": []string{}, "allowTagIds": []string{}},
		"storage":       map[string]any{"blockWrites": false, "blockDeletes": false, "allowCollections": []string{}},
		"accountDelete": map[string]any{"enabled": true, "mode": "grace", "graceDays": 14},
	}
	for k, v := range patch {
		body[k] = v
	}
	out := h.mustAPI(200, "PUT", "/server-rules", body)
	serverRulesStore.Invalidate()
	return out
}

func gateCode(err error) string {
	var rt *runtime.Error
	if !errors.As(err, &rt) {
		return ""
	}
	var m map[string]any
	_ = json.Unmarshal([]byte(rt.Message), &m)
	code, _ := m["code"].(string)
	return code
}

func TestServerRulesMaintenance(t *testing.T) {
	h := newHarness(t)
	resetServerRules(h)
	rpcgate.Set(h.s.maintenanceGate)
	u := h.user("fa")
	ctx := context.Background()
	if _, err := rpcgate.Check(ctx, "economy_catalog", u); err != nil {
		t.Fatal("maintenance off must not block")
	}
	out := h.rules(map[string]any{"maintenance": map[string]any{"enabled": true, "message": map[string]string{"fa": "تا ساعت ۵ تعمیر داریم"}, "allowRpcs": []string{" Economy_Catalog "}, "allowTagIds": []string{}}})
	if m := out["maintenance"].(map[string]any); m["allowRpcs"].([]any)[0] != "economy_catalog" {
		t.Fatalf("allow list not cleaned: %v", m)
	}
	code, err := rpcgate.Check(ctx, "economy_items", u)
	if err == nil || code != codes.Unavailable || gateCode(err) != CodeMaintenance {
		t.Fatalf("blocked rpc: %v %v", code, err)
	}
	var rt *runtime.Error
	errors.As(err, &rt)
	if !json.Valid([]byte(rt.Message)) || !contains(rt.Message, "تا ساعت ۵ تعمیر داریم") {
		t.Fatalf("message: %s", rt.Message)
	}
	for _, id := range []string{"economy_catalog", "base_app_version", "base_server_status"} {
		if _, err := rpcgate.Check(ctx, id, u); err != nil {
			t.Fatalf("%s must stay open: %v", id, err)
		}
	}
	if _, err := rpcgate.Check(ctx, "economy_items", ""); err != nil {
		t.Fatal("server calls without a user pass")
	}
	st := h.mustRPC(u, "base_server_status", map[string]any{})
	if m := st["maintenance"].(map[string]any); m["on"] != true || m["bypass"] != false {
		t.Fatalf("status: %v", st)
	}

	tag := uuid.Must(uuid.NewV4()).String()
	if _, err := h.s.db.Exec(`INSERT INTO base_tag (id, key_type, key, name) VALUES ($1,'user','testers','{"en":"Testers"}'::jsonb)`, tag); err != nil {
		t.Fatal(err)
	}
	if _, err := h.s.db.Exec(`INSERT INTO base_user_tag (user_id, tag_id) VALUES ($1,$2)`, u, tag); err != nil {
		t.Fatal(err)
	}
	h.rules(map[string]any{"maintenance": map[string]any{"enabled": true, "message": map[string]string{}, "allowRpcs": []string{}, "allowTagIds": []string{tag}}})
	if _, err := rpcgate.Check(ctx, "economy_items", u); err != nil {
		t.Fatalf("tester bypass: %v", err)
	}
	other := h.user("en")
	_, err = rpcgate.Check(ctx, "economy_items", other)
	if gateCode(err) != CodeMaintenance || !contains(err.Error(), "maintenance") {
		t.Fatalf("english default text: %v", err)
	}
	past := time.Now().Add(-time.Minute).UTC().Format(time.RFC3339)
	h.rules(map[string]any{"maintenance": map[string]any{"enabled": true, "until": past, "message": map[string]string{}, "allowRpcs": []string{}, "allowTagIds": []string{}}})
	if _, err := rpcgate.Check(ctx, "economy_items", other); err != nil {
		t.Fatal("maintenance ends at until")
	}
	h.expectError(400, "INVALID_RULE", "PUT", "/server-rules", map[string]any{"accountDelete": map[string]any{"enabled": true, "mode": "later", "graceDays": 3}})
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}

func TestServerRulesStorageLock(t *testing.T) {
	h := newHarness(t)
	resetServerRules(h)
	u := h.user("fa")
	ctx := context.WithValue(context.Background(), runtime.RUNTIME_CTX_USER_ID, u)
	write := &api.WriteStorageObjectsRequest{Objects: []*api.WriteStorageObject{{Collection: "saves", Key: "a", Value: "{}"}}}
	del := &api.DeleteStorageObjectsRequest{ObjectIds: []*api.DeleteStorageObjectId{{Collection: "saves", Key: "a"}}}
	if _, err := h.s.beforeStorageWrite(ctx, nil, nil, nil, write); err != nil {
		t.Fatal("open by default")
	}
	h.rules(map[string]any{"storage": map[string]any{"blockWrites": true, "blockDeletes": true, "allowCollections": []string{"settings"}}})
	if _, err := h.s.beforeStorageWrite(ctx, nil, nil, nil, write); gateCode(err) != CodeStorageWriteBlocked {
		t.Fatalf("write: %v", err)
	}
	if _, err := h.s.beforeStorageDelete(ctx, nil, nil, nil, del); gateCode(err) != CodeStorageDeleteBlocked {
		t.Fatalf("delete: %v", err)
	}
	ok := &api.WriteStorageObjectsRequest{Objects: []*api.WriteStorageObject{{Collection: "settings", Key: "a", Value: "{}"}}}
	if _, err := h.s.beforeStorageWrite(ctx, nil, nil, nil, ok); err != nil {
		t.Fatalf("allowed collection: %v", err)
	}
	if _, err := h.s.beforeStorageWrite(context.Background(), nil, nil, nil, write); err != nil {
		t.Fatal("server writes pass")
	}
}

func TestServerRulesAccountDelete(t *testing.T) {
	h := newHarness(t)
	resetServerRules(h)
	u := h.user("fa")
	if code := h.rpcCode(u, "base_account_delete", map[string]any{}); code != CodeConfirmRequired {
		t.Fatalf("confirm: %s", code)
	}
	out := h.mustRPC(u, "base_account_delete", map[string]any{"confirm": true})
	if out["deleted"] != false || out["graceDays"] != float64(14) {
		t.Fatalf("grace: %v", out)
	}
	if st := h.mustRPC(u, "base_server_status", map[string]any{}); st["deletionDueAt"] == nil {
		t.Fatal("status shows the pending deletion")
	}
	h.mustRPC(u, "base_account_delete_cancel", map[string]any{})
	if code := h.rpcCode(u, "base_account_delete_cancel", map[string]any{}); code != CodeNoPendingDeletion {
		t.Fatalf("second cancel: %s", code)
	}
	h.mustRPC(u, "base_account_delete", map[string]any{"confirm": true})
	if _, err := h.s.db.Exec(`UPDATE base_account_deletion SET due_at=now() - interval '1 minute' WHERE user_id=$1`, u); err != nil {
		t.Fatal(err)
	}
	if n := h.s.deleteDueAccounts(context.Background()); n != 1 {
		t.Fatalf("due deletions: %d", n)
	}
	var left int
	_ = h.s.db.QueryRow(`SELECT count(*) FROM users WHERE id=$1`, u).Scan(&left)
	if left != 0 {
		t.Fatal("account still there")
	}
	v := h.user("en")
	h.rules(map[string]any{"accountDelete": map[string]any{"enabled": true, "mode": "immediate", "graceDays": 14}})
	if out := h.mustRPC(v, "base_account_delete", map[string]any{"confirm": true}); out["deleted"] != true {
		t.Fatalf("immediate: %v", out)
	}
	w := h.user("en")
	h.rules(map[string]any{"accountDelete": map[string]any{"enabled": false, "mode": "grace", "graceDays": 14}})
	if code := h.rpcCode(w, "base_account_delete", map[string]any{"confirm": true}); code != CodeAccountDeleteOff {
		t.Fatalf("off: %s", code)
	}
}
