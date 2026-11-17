package gen

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/heroiclabs/nakama/v3/internal/apidocs"
)

const repoRoot = "../../.."

func servedDoc(t *testing.T, ids []string) map[string]any {
	t.Helper()
	modules, err := Discover(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), apidocs.FileName)
	if err := WriteOpenAPI(repoRoot, file, modules); err != nil {
		t.Fatal(err)
	}
	h := apidocs.NewHandler(apidocs.Options{SchemaFile: file, RPCIDs: func() []string { return ids }})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://api.test/docs/openapi.json", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var doc map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestServedDocumentListsEveryRegisteredRPC(t *testing.T) {
	modules, err := Discover(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	rpcs, err := ScanRPCs(repoRoot, modules)
	if err != nil {
		t.Fatal(err)
	}
	if len(rpcs) < 100 {
		t.Fatalf("only %d RPCs scanned", len(rpcs))
	}
	ids := []string{"lua_only_rpc"}
	for _, r := range rpcs {
		ids = append(ids, r.ID)
	}
	doc := servedDoc(t, ids)
	paths := doc["paths"].(map[string]any)
	schemas := doc["components"].(map[string]any)["schemas"].(map[string]any)
	for _, id := range ids {
		item, ok := paths[apidocs.RPCPathPrefix+id].(map[string]any)
		if !ok {
			t.Errorf("%s is registered but not in the document", id)
			continue
		}
		if _, ok := item["post"]; !ok {
			t.Errorf("%s has no POST", id)
		}
		for _, n := range []string{apidocs.RequestName(id), apidocs.ResponseName(id)} {
			if _, ok := schemas[n]; !ok {
				t.Errorf("%s: schema %s missing", id, n)
			}
		}
	}
	rpcPaths := 0
	for p := range paths {
		if strings.HasPrefix(p, apidocs.RPCPathPrefix) {
			rpcPaths++
		}
	}
	if rpcPaths != len(ids) {
		t.Errorf("document has %d RPC paths, %d registered", rpcPaths, len(ids))
	}
	for _, p := range []string{"/v2/account/authenticate/device", "/v2/account/authenticate/email", "/v2/account/authenticate/custom", "/v2/account/session/refresh", "/v2/session/logout", "/v2/account"} {
		if _, ok := paths[p]; !ok {
			t.Errorf("REST path %s missing", p)
		}
	}
	if s := doc["servers"].([]any)[0].(map[string]any)["url"]; s != "http://api.test" {
		t.Errorf("servers url %v", s)
	}
	raw, _ := json.Marshal(doc)
	if strings.Contains(string(raw), "http_key") || strings.Contains(string(raw), "HttpKeyAuth") {
		t.Error("the server-to-server http_key must not be documented")
	}
	if strings.Contains(string(raw), "#/definitions/") {
		t.Error("swagger 2 refs left")
	}
	for _, m := range regexp.MustCompile(`"#/components/schemas/([A-Za-z0-9_]+)"`).FindAllStringSubmatch(string(raw), -1) {
		if _, ok := schemas[m[1]]; !ok {
			t.Errorf("dangling ref %s", m[1])
		}
	}
}

func TestUnregisteredRPCsAreLeftOut(t *testing.T) {
	doc := servedDoc(t, []string{"base_account_get"})
	paths := doc["paths"].(map[string]any)
	if _, ok := paths["/v2/rpc/base_account_get"]; !ok {
		t.Fatal("registered rpc missing")
	}
	if _, ok := paths["/v2/rpc/economy_catalog"]; ok {
		t.Fatal("rpc of a plugin that is not loaded is listed")
	}
	schemas := doc["components"].(map[string]any)["schemas"].(map[string]any)
	if _, ok := schemas["EconomyCatalogResponse"]; ok {
		t.Fatal("schema of an unregistered rpc kept")
	}
	for _, tg := range doc["tags"].([]any) {
		if tg.(map[string]any)["name"] == "economy" {
			t.Fatal("tag of an unloaded plugin kept")
		}
	}
}

func TestAccountGetIsTypedFromTheProtoMessage(t *testing.T) {
	doc := servedDoc(t, []string{"base_account_get"})
	s := doc["components"].(map[string]any)["schemas"].(map[string]any)["BaseAccountGetResponse"].(map[string]any)
	user := s["properties"].(map[string]any)["user"].(map[string]any)
	if _, ok := user["properties"].(map[string]any)["display_name"]; !ok {
		t.Fatalf("account user has no display_name: %v", user)
	}
}

func TestGeneratedFileIsNotCommitted(t *testing.T) {
	if _, err := os.Stat(filepath.Join(repoRoot, "integrations", apidocs.FileName)); err == nil {
		t.Fatal("the document is generated at build time")
	}
}

func TestNullableAndNestedShapesMatchLiveResponses(t *testing.T) {
	doc := servedDoc(t, []string{"economy_avatar_catalog", "base_my_status", "economy_chance_games"})
	schemas := doc["components"].(map[string]any)["schemas"].(map[string]any)
	at := func(s any, path ...string) map[string]any {
		cur := s.(map[string]any)
		for _, p := range path {
			if p == "[]" {
				cur = cur["items"].(map[string]any)
				continue
			}
			cur = cur["properties"].(map[string]any)[p].(map[string]any)
		}
		return cur
	}
	price := at(schemas["EconomyAvatarCatalogResponse"], "categories", "[]", "items", "[]", "price")
	if price["nullable"] != true {
		t.Errorf("avatar catalog price is not nullable: %v", price)
	}
	for _, k := range []string{"mute", "suspension"} {
		if s := at(schemas["BaseMyStatusResponse"], k); s["nullable"] != true {
			t.Errorf("my status %s is not nullable: %v", k, s)
		}
	}
	group := at(schemas["EconomyChanceGamesResponse"], "items", "[]", "prize", "[]", "group")
	for _, k := range []string{"id", "key", "name", "prizes"} {
		if _, ok := group["properties"].(map[string]any)[k]; !ok {
			t.Errorf("chance prize group has no %s: %v", k, group)
		}
	}
	prize := at(group, "prizes", "[]")
	for _, k := range []string{"id", "key", "name", "weight", "share"} {
		if _, ok := prize["properties"].(map[string]any)[k]; !ok {
			t.Errorf("chance prize group member has no %s: %v", k, prize)
		}
	}
}
