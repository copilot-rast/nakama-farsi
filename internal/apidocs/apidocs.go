package apidocs

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strings"
	"unicode"

	"github.com/heroiclabs/nakama/v3/internal/settingsrow"
)

const (
	RPCPathPrefix = "/v2/rpc/"
	FileName      = "openapi.json"
)

type Settings struct {
	Enabled      bool `json:"enabled"`
	ConsoleLogin bool `json:"consoleLogin"`
}

func DefaultSettings() Settings {
	return Settings{Enabled: true}
}

type rulesRow struct {
	Docs Settings `json:"docs"`
}

var store = settingsrow.New("base_server_rules", func() rulesRow { return rulesRow{Docs: DefaultSettings()} })

func Load(ctx context.Context, db *sql.DB) Settings {
	v, err := store.Load(ctx, db)
	if err != nil {
		return DefaultSettings()
	}
	return v.Docs
}

func Invalidate() {
	store.Invalidate()
}

func PascalCase(id string) string {
	var b strings.Builder
	up := true
	for _, r := range id {
		if r == '_' || r == '-' || r == '.' {
			up = true
			continue
		}
		if up {
			b.WriteRune(unicode.ToUpper(r))
			up = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func CamelCase(id string) string {
	p := PascalCase(id)
	if p == "" {
		return p
	}
	return strings.ToLower(p[:1]) + p[1:]
}

func RequestName(id string) string  { return PascalCase(id) + "Request" }
func ResponseName(id string) string { return PascalCase(id) + "Response" }

func ref(name string) map[string]any {
	return map[string]any{"$ref": "#/components/schemas/" + name}
}

func Operation(id, module, errorFormat string, errs []string) map[string]any {
	desc := "RPC " + id + ". Socket: rpc(\"" + id + "\", payload)."
	if len(errs) > 0 {
		desc += " Error codes: " + strings.Join(errs, ", ") + "."
	}
	codes := make([]any, 0, len(errs))
	for _, e := range errs {
		codes = append(codes, e)
	}
	op := map[string]any{
		"operationId": CamelCase(id),
		"summary":     id,
		"description": desc,
		"tags":        []any{module},
		"security":    []any{map[string]any{"BearerAuth": []any{}}},
		"parameters":  []any{map[string]any{"$ref": "#/components/parameters/Unwrap"}},
		"requestBody": map[string]any{
			"required": true,
			"content":  map[string]any{"application/json": map[string]any{"schema": ref(RequestName(id))}},
		},
		"responses": map[string]any{
			"200": map[string]any{
				"description": "The RPC result.",
				"content":     map[string]any{"application/json": map[string]any{"schema": ref(ResponseName(id))}},
			},
			"4XX":     map[string]any{"$ref": "#/components/responses/RpcError"},
			"default": map[string]any{"$ref": "#/components/responses/RpcError"},
		},
		"x-rpc-id":      id,
		"x-error-codes": codes,
	}
	if errorFormat != "" {
		op["x-error-format"] = errorFormat
	}
	return op
}

func Skeleton() map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Nakama API",
			"version":     "1",
			"license":     map[string]any{"name": "Apache-2.0", "url": "https://www.apache.org/licenses/LICENSE-2.0"},
			"description": "Client API of this server: login and session endpoints over HTTP, and every RPC registered in this build as POST /v2/rpc/{id}?unwrap with the request object as the raw JSON body. The same RPCs work over the socket (rpc with the same payload). Errors: the body is NakamaError; for RPCs with x-error-format envelope its message is a JSON RpcErrorEnvelope, map code to text with devkit_error_catalog.",
		},
		"servers": []any{map[string]any{"url": "/"}},
		"tags":    []any{},
		"paths":   map[string]any{},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"BearerAuth": map[string]any{"type": "http", "scheme": "bearer", "bearerFormat": "JWT", "description": "Session token from an authenticate call."},
				"BasicAuth":  map[string]any{"type": "http", "scheme": "basic", "description": "Server key as the username and an empty password; only for authenticate and session refresh."},
			},
			"parameters": map[string]any{
				"Unwrap": map[string]any{"name": "unwrap", "in": "query", "required": true, "description": "Send the request object as the raw JSON body and get the result unwrapped.", "schema": map[string]any{"type": "boolean", "enum": []any{true}, "default": true}},
			},
			"schemas": map[string]any{
				"NakamaError": map[string]any{
					"type":     "object",
					"required": []any{"code", "message"},
					"properties": map[string]any{
						"code":    map[string]any{"type": "integer", "description": "gRPC status code (3 invalid argument, 5 not found, 7 permission denied, 9 failed precondition, 16 unauthenticated, ...)."},
						"message": map[string]any{"type": "string", "description": "For envelope RPCs a JSON RpcErrorEnvelope; in code mode the bare error code."},
						"error":   map[string]any{"description": "Same as message."},
					},
				},
				"RpcErrorEnvelope": map[string]any{
					"type":     "object",
					"required": []any{"code"},
					"properties": map[string]any{
						"code":              ref("RpcErrorCode"),
						"status":            map[string]any{"type": "integer", "description": "HTTP-like status of the error."},
						"message":           map[string]any{"type": "string", "description": "Localized text for the player."},
						"meta":              map[string]any{"type": "object", "additionalProperties": true},
						"retryAfterSeconds": map[string]any{"type": "integer"},
						"lang":              map[string]any{"type": "string"},
						"field":             map[string]any{"type": "string"},
					},
				},
				"RpcErrorCode": map[string]any{"type": "string", "description": "Error code of a refused RPC; text per language comes from devkit_error_catalog."},
			},
			"responses": map[string]any{
				"RpcError": map[string]any{
					"description": "Refused call. The message carries the error code (see the operation's x-error-codes).",
					"content":     map[string]any{"application/json": map[string]any{"schema": ref("NakamaError")}},
				},
				"NakamaError": map[string]any{
					"description": "Refused call.",
					"content":     map[string]any{"application/json": map[string]any{"schema": ref("NakamaError")}},
				},
			},
		},
	}
}

func Document(raw []byte, rpcIDs []string, serverURL string) ([]byte, error) {
	doc := Skeleton()
	if len(raw) > 0 {
		var parsed map[string]any
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return nil, err
		}
		doc = parsed
	}
	paths, _ := doc["paths"].(map[string]any)
	if paths == nil {
		paths = map[string]any{}
		doc["paths"] = paths
	}
	comps, _ := doc["components"].(map[string]any)
	schemas, _ := comps["schemas"].(map[string]any)
	if schemas == nil {
		schemas = map[string]any{}
		comps["schemas"] = schemas
	}
	registered := map[string]bool{}
	for _, id := range rpcIDs {
		registered[id] = true
	}
	for p, item := range paths {
		id, ok := strings.CutPrefix(p, RPCPathPrefix)
		if !ok || registered[id] {
			continue
		}
		delete(paths, p)
		if op, ok := item.(map[string]any)["post"].(map[string]any); ok {
			if rid, ok := op["x-rpc-id"].(string); ok {
				id = rid
			}
		}
		delete(schemas, RequestName(id))
		delete(schemas, ResponseName(id))
	}
	for _, id := range rpcIDs {
		p := RPCPathPrefix + id
		if _, ok := paths[p]; ok {
			continue
		}
		op := Operation(id, "other", "", nil)
		op["description"] = op["description"].(string) + " No schema was generated for this RPC."
		paths[p] = map[string]any{"post": op}
		schemas[RequestName(id)] = map[string]any{"type": "object", "additionalProperties": true}
		schemas[ResponseName(id)] = map[string]any{}
	}
	used := map[string]bool{}
	for _, item := range paths {
		for _, o := range item.(map[string]any) {
			if op, ok := o.(map[string]any); ok {
				for _, t := range anyList(op["tags"]) {
					if s, ok := t.(string); ok {
						used[s] = true
					}
				}
			}
		}
	}
	tags := []any{}
	known := map[string]bool{}
	for _, t := range anyList(doc["tags"]) {
		if m, ok := t.(map[string]any); ok {
			name, _ := m["name"].(string)
			if used[name] {
				tags = append(tags, m)
				known[name] = true
			}
		}
	}
	extra := []string{}
	for t := range used {
		if !known[t] {
			extra = append(extra, t)
		}
	}
	sort.Strings(extra)
	for _, t := range extra {
		tags = append(tags, map[string]any{"name": t, "description": "RPCs registered in this build without a generated schema."})
	}
	doc["tags"] = tags
	if serverURL != "" {
		doc["servers"] = []any{map[string]any{"url": serverURL}}
	}
	return json.Marshal(doc)
}

func anyList(v any) []any {
	l, _ := v.([]any)
	return l
}
