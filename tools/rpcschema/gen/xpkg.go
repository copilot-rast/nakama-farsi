package gen

import (
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

var (
	localMu   sync.Mutex
	localPkgs = map[string]*pkg{}
)

func modulePath(root string) string {
	raw, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func (p *pkg) local(path string) *pkg {
	mod := modulePath(p.root)
	if mod == "" || !strings.HasPrefix(path, mod+"/") {
		return nil
	}
	rel := strings.TrimPrefix(path, mod+"/")
	key := p.root + "|" + rel
	localMu.Lock()
	if got, ok := localPkgs[key]; ok {
		localMu.Unlock()
		return got
	}
	localMu.Unlock()
	lp, err := loadPkgAs(p.root, Module{Name: p.module.Name, Dir: rel}, path)
	if err != nil {
		lp = nil
	}
	localMu.Lock()
	localPkgs[key] = lp
	localMu.Unlock()
	return lp
}

func (sc *scope) resolve(fun ast.Expr) (*pkg, *ast.FuncDecl) {
	if sc.p.info != nil {
		var id *ast.Ident
		switch f := fun.(type) {
		case *ast.Ident:
			id = f
		case *ast.SelectorExpr:
			id = f.Sel
		}
		if id != nil {
			if fn, ok := sc.p.info.Uses[id].(*types.Func); ok {
				recv := ""
				if sig, ok := fn.Type().(*types.Signature); ok && sig.Recv() != nil {
					if n, ok := derefType(sig.Recv().Type()).(*types.Named); ok {
						recv = n.Obj().Name()
					}
				}
				if fn.Pkg() == nil {
					return nil, nil
				}
				target := sc.p
				if fn.Pkg().Path() != sc.p.path {
					target = sc.p.local(fn.Pkg().Path())
				}
				if target == nil {
					return nil, nil
				}
				for _, d := range target.funcs[fn.Name()] {
					got := ""
					if d.Recv != nil && len(d.Recv.List) > 0 {
						got = recvTypeName(d.Recv.List[0].Type)
					}
					if got == recv {
						return target, d
					}
				}
				return nil, nil
			}
		}
	}
	return sc.p, sc.callee(fun)
}

func (sc *scope) assigned(name string, depth int) Schema {
	var out Schema
	ast.Inspect(sc.decl.Body, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		as, ok := n.(*ast.AssignStmt)
		if !ok || as.Tok != token.ASSIGN || len(as.Lhs) != len(as.Rhs) {
			return true
		}
		for i, l := range as.Lhs {
			if id, ok := l.(*ast.Ident); ok && id.Name == name {
				got := sc.expr(as.Rhs[i], depth+1)
				if len(got) > 0 && !isNullSchema(got) {
					out = merge(out, got)
				}
			}
		}
		return true
	})
	return out
}

func isNullSchema(s Schema) bool {
	return len(s) == 1 && s["type"] == "null"
}

func hasHoles(s Schema) bool {
	if len(s) == 0 {
		return true
	}
	if props, ok := s["properties"].(Schema); ok {
		for _, v := range props {
			if vs, ok := v.(Schema); ok && hasHoles(vs) {
				return true
			}
		}
	}
	if it, ok := s["items"].(Schema); ok && hasHoles(it) {
		return true
	}
	if ap, ok := s["additionalProperties"].(Schema); ok && len(ap) > 0 && hasHoles(ap) {
		return true
	}
	return false
}

func fillHoles(base, extra Schema) Schema {
	if len(extra) == 0 || isNullSchema(extra) {
		return base
	}
	if len(base) == 0 {
		return extra
	}
	out := Schema{}
	for k, v := range base {
		out[k] = v
	}
	bp, bok := base["properties"].(Schema)
	ep, eok := extra["properties"].(Schema)
	switch {
	case bok && eok:
		props := Schema{}
		for k, v := range bp {
			vs, _ := v.(Schema)
			if es, ok := ep[k].(Schema); ok {
				props[k] = fillHoles(vs, es)
			} else {
				props[k] = vs
			}
		}
		out["properties"] = props
	case !bok && eok && base["type"] == "object" && base["additionalProperties"] == nil:
		return extra
	}
	if bi, ok := base["items"].(Schema); ok {
		if ei, ok := extra["items"].(Schema); ok {
			out["items"] = fillHoles(bi, ei)
		}
	}
	if ba, ok := base["additionalProperties"].(Schema); ok {
		if ea, ok := extra["additionalProperties"].(Schema); ok {
			out["additionalProperties"] = fillHoles(ba, ea)
		}
	}
	return out
}

func (sc *scope) fillComposite(lit *ast.CompositeLit, s Schema, depth int) Schema {
	if sc.p.info == nil {
		return s
	}
	tv, ok := sc.p.info.Types[lit]
	if !ok || tv.Type == nil {
		return s
	}
	st, ok := derefType(tv.Type).Underlying().(*types.Struct)
	if !ok {
		return s
	}
	props, ok := s["properties"].(Schema)
	if !ok {
		return s
	}
	jsonOf := map[string]string{}
	for i := 0; i < st.NumFields(); i++ {
		name, _, _ := strings.Cut(reflect.StructTag(st.Tag(i)).Get("json"), ",")
		if name == "" {
			name = st.Field(i).Name()
		}
		jsonOf[st.Field(i).Name()] = name
	}
	extra := Schema{}
	for _, el := range lit.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		k, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		jn := jsonOf[k.Name]
		if cur, ok := props[jn].(Schema); ok && hasHoles(cur) {
			extra[jn] = sc.expr(kv.Value, depth+1)
		}
	}
	if len(extra) == 0 {
		return s
	}
	return fillHoles(s, Schema{"type": "object", "properties": extra})
}
