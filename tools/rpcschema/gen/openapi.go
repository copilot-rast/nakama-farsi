package gen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/heroiclabs/nakama/v3/internal/apidocs"
)

const SwaggerFile = "apigrpc/apigrpc.swagger.json"

type restOp struct {
	Path   string
	Method string
	ID     string
	Name   string
}

var RestOps = []restOp{
	{"/v2/account/authenticate/device", "post", "Nakama_AuthenticateDevice", "authenticateDevice"},
	{"/v2/account/authenticate/email", "post", "Nakama_AuthenticateEmail", "authenticateEmail"},
	{"/v2/account/authenticate/custom", "post", "Nakama_AuthenticateCustom", "authenticateCustom"},
	{"/v2/account/session/refresh", "post", "Nakama_SessionRefresh", "sessionRefresh"},
	{"/v2/session/logout", "post", "Nakama_SessionLogout", "sessionLogout"},
	{"/v2/account", "get", "Nakama_GetAccount", "getAccount"},
	{"/v2/account", "put", "Nakama_UpdateAccount", "updateAccount"},
}

func BuildOpenAPI(root string, modules []Module) ([]byte, error) {
	rpcs, err := ScanRPCs(root, modules)
	if err != nil {
		return nil, err
	}
	codes, err := ScanCodes(root, modules)
	if err != nil {
		return nil, err
	}
	doc := apidocs.Skeleton()
	comps := doc["components"].(map[string]any)
	schemas := comps["schemas"].(map[string]any)
	paths := doc["paths"].(map[string]any)
	tags := map[string]bool{}
	if err := addRest(root, paths, schemas); err != nil {
		return nil, err
	}
	if len(RestOps) > 0 {
		tags["nakama"] = true
	}
	list := make([]string, 0, len(codes))
	for c := range codes {
		list = append(list, c)
	}
	sort.Strings(list)
	schemas["RpcErrorCode"].(map[string]any)["enum"] = stringsAny(list)
	for _, r := range rpcs {
		tags[r.Module] = true
		schemas[apidocs.RequestName(r.ID)] = stripRequired(CleanSchema(r.Request))
		schemas[apidocs.ResponseName(r.ID)] = CleanSchema(r.Response)
		paths[apidocs.RPCPathPrefix+r.ID] = map[string]any{"post": apidocs.Operation(r.ID, r.Module, r.ErrorFormat, r.Errors)}
	}
	names := make([]string, 0, len(tags))
	for t := range tags {
		names = append(names, t)
	}
	sort.Strings(names)
	tagList := []any{}
	for _, t := range names {
		desc := "RPCs of the " + t + " plugin."
		if t == "nakama" {
			desc = "Built-in Nakama endpoints: login, session and account."
		}
		tagList = append(tagList, map[string]any{"name": t, "description": desc})
	}
	doc["tags"] = tagList
	return marshal(doc)
}

func WriteOpenAPI(root, file string, modules []Module) error {
	raw, err := BuildOpenAPI(root, modules)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	return os.WriteFile(file, raw, 0o644)
}

func stringsAny(in []string) []any {
	out := make([]any, 0, len(in))
	for _, s := range in {
		out = append(out, s)
	}
	return out
}

func CleanSchema(v any) any {
	switch s := v.(type) {
	case Schema:
		return CleanSchema(map[string]any(s))
	case map[string]any:
		out := map[string]any{}
		for k, x := range s {
			switch k {
			case "properties":
				props := map[string]any{}
				if pm, ok := asMap(x); ok {
					for pk, pv := range pm {
						props[pk] = CleanSchema(pv)
					}
				}
				out[k] = props
			case "items", "additionalProperties":
				if b, ok := x.(bool); ok {
					out[k] = b
				} else {
					out[k] = CleanSchema(x)
				}
			case "type":
				if x == "null" {
					out["nullable"] = true
				} else {
					out[k] = x
				}
			default:
				out[k] = x
			}
		}
		if out["type"] == "array" {
			if _, ok := out["items"]; !ok {
				out["items"] = map[string]any{}
			}
		}
		if _, ok := out["nullable"]; ok {
			if _, typed := out["type"]; !typed {
				if _, ref := out["$ref"]; !ref {
					delete(out, "nullable")
				}
			}
		}
		return out
	}
	return v
}

func asMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	case Schema:
		return map[string]any(m), true
	}
	return nil, false
}

func addRest(root string, paths, schemas map[string]any) error {
	raw, err := os.ReadFile(filepath.Join(root, SwaggerFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var sw map[string]any
	if err := json.Unmarshal(raw, &sw); err != nil {
		return err
	}
	swPaths, _ := sw["paths"].(map[string]any)
	defs, _ := sw["definitions"].(map[string]any)
	needed := map[string]bool{}
	for _, ro := range RestOps {
		item, _ := swPaths[ro.Path].(map[string]any)
		src, _ := item[ro.Method].(map[string]any)
		if src == nil {
			return fmt.Errorf("swagger has no %s %s", ro.Method, ro.Path)
		}
		op := map[string]any{"operationId": ro.Name, "tags": []any{"nakama"}}
		if s, ok := src["summary"].(string); ok {
			op["summary"] = s
		}
		if sec, ok := src["security"].([]any); ok && len(sec) > 0 {
			op["security"] = []any{map[string]any{"BasicAuth": []any{}}}
		} else {
			op["security"] = []any{map[string]any{"BearerAuth": []any{}}}
		}
		params := []any{}
		for _, p := range listOf(src["parameters"]) {
			pm := p.(map[string]any)
			if pm["in"] == "body" {
				schema := convertRef(pm["schema"], needed)
				op["requestBody"] = map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": schema}}}
				continue
			}
			np := map[string]any{"name": pm["name"], "in": pm["in"], "required": pm["required"] == true}
			if d, ok := pm["description"].(string); ok {
				np["description"] = d
			}
			ps := map[string]any{"type": pm["type"]}
			if f, ok := pm["format"]; ok {
				ps["format"] = f
			}
			np["schema"] = ps
			params = append(params, np)
		}
		if len(params) > 0 {
			op["parameters"] = params
		}
		responses := map[string]any{}
		if r200 := mapOf(mapOf(src["responses"])["200"]); r200 != nil {
			resp := map[string]any{"description": strOr(r200["description"], "A successful response.")}
			if sch, ok := r200["schema"]; ok {
				conv := convertRef(sch, needed)
				if cm, ok := conv.(map[string]any); ok && cm["type"] == "object" && cm["properties"] == nil && cm["$ref"] == nil {
					conv = map[string]any{"type": "object"}
				}
				resp["content"] = map[string]any{"application/json": map[string]any{"schema": conv}}
			}
			responses["200"] = resp
		}
		responses["4XX"] = map[string]any{"$ref": "#/components/responses/NakamaError"}
		responses["default"] = map[string]any{"$ref": "#/components/responses/NakamaError"}
		op["responses"] = responses
		pi, _ := paths[ro.Path].(map[string]any)
		if pi == nil {
			pi = map[string]any{}
			paths[ro.Path] = pi
		}
		pi[ro.Method] = op
	}
	done := map[string]bool{}
	for len(done) < len(needed) {
		for name := range needed {
			if done[name] {
				continue
			}
			done[name] = true
			if d, ok := defs[name]; ok {
				schemas[name] = convertSchema(d, needed)
			}
		}
	}
	return nil
}

func listOf(v any) []any {
	l, _ := v.([]any)
	return l
}

func mapOf(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func strOr(v any, def string) string {
	if s, ok := v.(string); ok && s != "" {
		return s
	}
	return def
}

func convertRef(v any, needed map[string]bool) any {
	return convertSchema(v, needed)
}

func convertSchema(v any, needed map[string]bool) any {
	switch s := v.(type) {
	case map[string]any:
		out := map[string]any{}
		for k, x := range s {
			switch k {
			case "$ref":
				ref, _ := x.(string)
				name := strings.TrimPrefix(ref, "#/definitions/")
				needed[name] = true
				out[k] = "#/components/schemas/" + name
			case "properties":
				props := map[string]any{}
				for pk, pv := range mapOf(x) {
					props[ProtoName(pk)] = convertSchema(pv, needed)
				}
				out[k] = props
			case "required":
				req := []any{}
				for _, r := range listOf(x) {
					if rs, ok := r.(string); ok {
						req = append(req, ProtoName(rs))
					}
				}
				if len(req) > 0 {
					out[k] = req
				}
			case "items", "additionalProperties":
				if b, ok := x.(bool); ok {
					out[k] = b
				} else {
					out[k] = convertSchema(x, needed)
				}
			case "title", "x-nullable", "readOnly":
			default:
				out[k] = x
			}
		}
		if _, ok := out["enum"]; ok && out["type"] == "string" {
			delete(out, "enum")
			out["type"] = "integer"
			out["description"] = strings.TrimSpace(strOr(out["description"], "") + " Sent as the enum number.")
			delete(out, "default")
		}
		if r, ok := out["$ref"]; ok {
			return map[string]any{"$ref": r}
		}
		return out
	}
	return v
}

func ProtoName(camel string) string {
	var b strings.Builder
	for i, r := range camel {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func stripRequired(v any) any {
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}
	delete(m, "required")
	if props, ok := m["properties"].(map[string]any); ok {
		for k, x := range props {
			props[k] = stripRequired(x)
		}
	}
	for _, k := range []string{"items", "additionalProperties"} {
		if x, ok := m[k].(map[string]any); ok {
			m[k] = stripRequired(x)
		}
	}
	return m
}
