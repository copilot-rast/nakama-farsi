package gen

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/importer"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

var (
	exportMu    sync.Mutex
	exportRoot  string
	exportFiles map[string]string
	TypesOff    bool
	BuildFlags  []string
)

func exportData(root string, dirs []string) map[string]string {
	exportMu.Lock()
	defer exportMu.Unlock()
	if TypesOff {
		return nil
	}
	if exportFiles != nil && exportRoot == root {
		return exportFiles
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		return nil
	}
	args := append([]string{"list", "-e", "-export", "-deps", "-json=ImportPath,Export"}, BuildFlags...)
	for _, d := range dirs {
		args = append(args, "./"+filepath.ToSlash(d))
	}
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	cmd.Env = os.Environ()
	if _, err := os.Stat(filepath.Join(root, "vendor")); err == nil && !strings.Contains(os.Getenv("GOFLAGS"), "-mod=") {
		cmd.Env = append(cmd.Env, "GOFLAGS=-mod=vendor")
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil && out.Len() == 0 {
		return nil
	}
	files := map[string]string{}
	dec := json.NewDecoder(&out)
	for {
		var p struct{ ImportPath, Export string }
		if err := dec.Decode(&p); err != nil {
			break
		}
		if p.Export != "" {
			files[p.ImportPath] = p.Export
		}
	}
	exportRoot, exportFiles = root, files
	return files
}

func PrepareTypes(root string, modules []Module) {
	dirs := make([]string, 0, len(modules))
	for _, m := range modules {
		if _, err := os.Stat(filepath.Join(root, m.Dir)); err == nil {
			dirs = append(dirs, m.Dir)
		}
	}
	exportData(root, dirs)
}

func (p *pkg) typecheck() {
	files := exportData(p.root, []string{p.module.Dir})
	if len(files) == 0 || len(p.asts) == 0 {
		return
	}
	imp := importer.ForCompiler(p.fset, "gc", func(path string) (io.ReadCloser, error) {
		f, ok := files[path]
		if !ok {
			return nil, os.ErrNotExist
		}
		return os.Open(f)
	})
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}, Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}}
	conf := types.Config{Importer: imp, Error: func(error) {}, FakeImportC: true}
	_, _ = conf.Check(p.path, p.fset, p.asts, info)
	p.info = info
}

func (p *pkg) typedExpr(e ast.Expr) (Schema, bool) {
	if p.info == nil || e == nil {
		return nil, false
	}
	tv, ok := p.info.Types[e]
	if !ok || tv.Type == nil {
		if id, ok := e.(*ast.Ident); ok {
			if obj := p.info.Uses[id]; obj != nil {
				if v, ok := obj.(*types.Var); ok {
					return goSchema(v.Type())
				}
			}
		}
		return nil, false
	}
	if b, ok := tv.Type.(*types.Basic); ok && b.Kind() == types.UntypedNil {
		return nil, false
	}
	if call, ok := e.(*ast.CallExpr); ok && len(call.Args) == 1 && rawResult(tv.Type) {
		if at, ok := p.info.Types[call.Args[0]]; ok && at.Type != nil && hasMethod(derefType(at.Type), "ProtoReflect") {
			return protoSchema(at.Type, map[types.Type]bool{}, 0), true
		}
	}
	return goSchema(tv.Type)
}

func derefType(t types.Type) types.Type {
	if pt, ok := t.(*types.Pointer); ok {
		return pt.Elem()
	}
	return t
}

func rawResult(t types.Type) bool {
	if tu, ok := t.(*types.Tuple); ok {
		if tu.Len() == 0 {
			return false
		}
		t = tu.At(0).Type()
	}
	if n, ok := t.(*types.Named); ok && n.Obj().Name() == "RawMessage" {
		return true
	}
	switch u := t.Underlying().(type) {
	case *types.Interface:
		return true
	case *types.Slice:
		b, ok := u.Elem().Underlying().(*types.Basic)
		return ok && b.Kind() == types.Byte
	}
	return false
}

func protoSchema(t types.Type, seen map[types.Type]bool, depth int) Schema {
	t = derefType(types.Unalias(t))
	if depth > 10 {
		return Schema{"type": "object"}
	}
	if n, ok := t.(*types.Named); ok {
		switch n.Obj().Pkg().Path() + "." + n.Obj().Name() {
		case "google.golang.org/protobuf/types/known/timestamppb.Timestamp":
			return Schema{"type": "string", "format": "date-time"}
		case "google.golang.org/protobuf/types/known/wrapperspb.StringValue":
			return Schema{"type": "string"}
		case "google.golang.org/protobuf/types/known/wrapperspb.BoolValue":
			return Schema{"type": "boolean"}
		case "google.golang.org/protobuf/types/known/wrapperspb.Int32Value", "google.golang.org/protobuf/types/known/wrapperspb.UInt32Value":
			return Schema{"type": "integer"}
		case "google.golang.org/protobuf/types/known/wrapperspb.Int64Value", "google.golang.org/protobuf/types/known/wrapperspb.UInt64Value":
			return Schema{"type": "string", "format": "int64"}
		case "google.golang.org/protobuf/types/known/wrapperspb.DoubleValue", "google.golang.org/protobuf/types/known/wrapperspb.FloatValue":
			return Schema{"type": "number"}
		case "google.golang.org/protobuf/types/known/structpb.Struct":
			return Schema{"type": "object"}
		}
		if seen[n] {
			return Schema{"type": "object"}
		}
		seen[n] = true
		defer delete(seen, n)
	}
	switch u := t.Underlying().(type) {
	case *types.Basic:
		switch {
		case u.Info()&types.IsBoolean != 0:
			return Schema{"type": "boolean"}
		case u.Kind() == types.Int64 || u.Kind() == types.Uint64:
			return Schema{"type": "string", "format": "int64"}
		case u.Info()&types.IsInteger != 0:
			return Schema{"type": "integer"}
		case u.Info()&types.IsFloat != 0:
			return Schema{"type": "number"}
		case u.Info()&types.IsString != 0:
			return Schema{"type": "string"}
		}
	case *types.Slice:
		if b, ok := u.Elem().Underlying().(*types.Basic); ok && b.Kind() == types.Byte {
			return Schema{"type": "string", "format": "byte"}
		}
		return Schema{"type": "array", "items": protoSchema(u.Elem(), seen, depth+1)}
	case *types.Map:
		return Schema{"type": "object", "additionalProperties": protoSchema(u.Elem(), seen, depth+1)}
	case *types.Struct:
		props := Schema{}
		for i := 0; i < u.NumFields(); i++ {
			f := u.Field(i)
			if !f.Exported() {
				continue
			}
			tag := reflect.StructTag(u.Tag(i)).Get("protobuf")
			name := ""
			for _, part := range strings.Split(tag, ",") {
				if v, ok := strings.CutPrefix(part, "name="); ok {
					name = v
				}
			}
			if name == "" {
				continue
			}
			props[name] = protoSchema(f.Type(), seen, depth+1)
		}
		return Schema{"type": "object", "properties": props}
	}
	return Schema{}
}

func (p *pkg) typedIdent(decl *ast.FuncDecl, name string) Schema {
	if p.info == nil || decl == nil || decl.Body == nil {
		return nil
	}
	for id, obj := range p.info.Defs {
		if obj == nil || id.Name != name || id.Pos() < decl.Pos() || id.End() > decl.End() {
			continue
		}
		if v, ok := obj.(*types.Var); ok {
			s, _ := goSchema(v.Type())
			return s
		}
	}
	return nil
}

func goSchema(t types.Type) (Schema, bool) {
	return (&goTyper{seen: map[types.Type]int{}}).schema(t, 0)
}

type goTyper struct {
	seen map[types.Type]int
}

func (g *goTyper) schema(t types.Type, depth int) (Schema, bool) {
	if depth > 14 {
		return Schema{"type": "object"}, true
	}
	switch v := t.(type) {
	case *types.Tuple:
		if v.Len() == 0 {
			return nil, false
		}
		return g.schema(v.At(0).Type(), depth)
	case *types.Alias:
		return g.schema(types.Unalias(v), depth)
	case *types.Pointer:
		s, exact := g.schema(v.Elem(), depth+1)
		if s == nil {
			return nil, false
		}
		c := Schema{}
		for k, x := range s {
			c[k] = x
		}
		if len(c) > 0 && c["$ref"] == nil {
			c["nullable"] = true
		}
		return c, exact
	case *types.Named:
		obj := v.Obj()
		path := ""
		if obj.Pkg() != nil {
			path = obj.Pkg().Path()
		}
		switch path + "." + obj.Name() {
		case "time.Time":
			return Schema{"type": "string", "format": "date-time"}, true
		case "time.Duration":
			return Schema{"type": "integer"}, true
		case "encoding/json.RawMessage":
			return Schema{"description": "Any JSON value."}, true
		case "encoding/json.Number":
			return Schema{"type": "number"}, true
		case "database/sql.NullString":
			return Schema{"type": "string", "nullable": true}, true
		case "database/sql.NullInt64", "database/sql.NullInt32", "database/sql.NullInt16":
			return Schema{"type": "integer", "nullable": true}, true
		case "database/sql.NullFloat64":
			return Schema{"type": "number", "nullable": true}, true
		case "database/sql.NullBool":
			return Schema{"type": "boolean", "nullable": true}, true
		case "database/sql.NullTime":
			return Schema{"type": "string", "format": "date-time", "nullable": true}, true
		case "github.com/gofrs/uuid/v5.UUID", "github.com/gofrs/uuid.UUID", "github.com/google/uuid.UUID":
			return Schema{"type": "string", "format": "uuid"}, true
		case "google.golang.org/protobuf/types/known/timestamppb.Timestamp":
			return Schema{"type": "object", "properties": Schema{"seconds": Schema{"type": "integer"}, "nanos": Schema{"type": "integer"}}}, true
		}
		if hasMethod(v, "MarshalJSON") {
			return Schema{}, true
		}
		if hasMethod(v, "MarshalText") {
			return Schema{"type": "string"}, true
		}
		if g.seen[v] > 0 {
			return Schema{"type": "object"}, true
		}
		g.seen[v]++
		defer func() { g.seen[v]-- }()
		return g.schema(v.Underlying(), depth+1)
	case *types.Basic:
		switch {
		case v.Info()&types.IsBoolean != 0:
			return Schema{"type": "boolean"}, true
		case v.Info()&types.IsInteger != 0:
			return Schema{"type": "integer"}, true
		case v.Info()&types.IsFloat != 0:
			return Schema{"type": "number"}, true
		case v.Info()&types.IsString != 0:
			return Schema{"type": "string"}, true
		}
		return nil, false
	case *types.Slice:
		if b, ok := v.Elem().Underlying().(*types.Basic); ok && b.Kind() == types.Byte {
			return Schema{"type": "string"}, true
		}
		it, exact := g.schema(v.Elem(), depth+1)
		if it == nil {
			it = Schema{}
		}
		return Schema{"type": "array", "items": it}, exact
	case *types.Array:
		it, exact := g.schema(v.Elem(), depth+1)
		if it == nil {
			it = Schema{}
		}
		return Schema{"type": "array", "items": it}, exact
	case *types.Map:
		val, exact := g.schema(v.Elem(), depth+1)
		if val == nil || len(val) == 0 {
			return Schema{"type": "object"}, false
		}
		return Schema{"type": "object", "additionalProperties": val}, exact
	case *types.Interface:
		return nil, false
	case *types.Struct:
		props := Schema{}
		required := []any{}
		exact := true
		g.fields(v, depth, props, &required, &exact)
		out := Schema{"type": "object", "properties": props}
		if len(required) > 0 {
			out["required"] = required
		}
		_ = exact
		return out, true
	}
	return nil, false
}

func (g *goTyper) fields(st *types.Struct, depth int, props Schema, required *[]any, exact *bool) {
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		tag := reflect.StructTag(st.Tag(i)).Get("json")
		if tag == "-" {
			continue
		}
		name, opts, _ := strings.Cut(tag, ",")
		if f.Embedded() && name == "" {
			ft := f.Type()
			if pt, ok := ft.(*types.Pointer); ok {
				ft = pt.Elem()
			}
			if est, ok := ft.Underlying().(*types.Struct); ok {
				g.fields(est, depth+1, props, required, exact)
				continue
			}
		}
		if !f.Exported() {
			continue
		}
		if name == "" {
			name = f.Name()
		}
		s, ok := g.schema(f.Type(), depth+1)
		if s == nil {
			s = Schema{}
		}
		if !ok {
			*exact = false
		}
		props[name] = s
		if !strings.Contains(opts, "omitempty") && !strings.Contains(opts, "omitzero") {
			*required = append(*required, name)
		}
	}
}

func hasMethod(t types.Type, name string) bool {
	for _, tt := range []types.Type{t, types.NewPointer(t)} {
		ms := types.NewMethodSet(tt)
		for i := 0; i < ms.Len(); i++ {
			if ms.At(i).Obj().Name() == name {
				return true
			}
		}
	}
	return false
}
