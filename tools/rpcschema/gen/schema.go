package gen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Schema map[string]any

type RPC struct {
	ID          string   `json:"id"`
	Module      string   `json:"module"`
	Handler     string   `json:"handler"`
	Source      string   `json:"source"`
	Auth        string   `json:"auth"`
	ErrorFormat string   `json:"errorFormat"`
	CodeMode    string   `json:"codeModeFormat,omitempty"`
	Request     Schema   `json:"request"`
	Response    Schema   `json:"response"`
	Errors      []string `json:"errors"`
}

type pkg struct {
	module  Module
	root    string
	funcs   map[string][]*ast.FuncDecl
	types   map[string]ast.Expr
	files   map[*ast.FuncDecl]string
	formats map[string]string
	fset    *token.FileSet
	path    string
	asts    []*ast.File
	info    *types.Info
}

var wrapperFormat = map[string]string{
	"rpc":       "code",
	"rpcSized":  "code",
	"chanceRPC": "json",
	"RPC":       "envelope",
	"wrapRPC":   "envelope",
}

func loadPkg(root string, m Module) (*pkg, error) {
	return loadPkgAs(root, m, "main")
}

func loadPkgAs(root string, m Module, path string) (*pkg, error) {
	files, err := goFiles(filepath.Join(root, m.Dir), false)
	if err != nil {
		return nil, err
	}
	p := &pkg{module: m, root: root, path: path, funcs: map[string][]*ast.FuncDecl{}, types: map[string]ast.Expr{}, files: map[*ast.FuncDecl]string{}, formats: map[string]string{}}
	fset := token.NewFileSet()
	p.fset = fset
	for _, f := range files {
		file, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			return nil, err
		}
		p.asts = append(p.asts, file)
		rel, _ := filepath.Rel(root, f)
		for _, d := range file.Decls {
			switch decl := d.(type) {
			case *ast.FuncDecl:
				p.funcs[decl.Name.Name] = append(p.funcs[decl.Name.Name], decl)
				p.files[decl] = filepath.ToSlash(rel)
			case *ast.GenDecl:
				for _, s := range decl.Specs {
					if ts, ok := s.(*ast.TypeSpec); ok {
						p.types[ts.Name.Name] = ts.Type
					}
				}
			}
		}
	}
	p.scanWrappers()
	p.typecheck()
	return p, nil
}

func (p *pkg) fn(name string) *ast.FuncDecl {
	if list := p.funcs[name]; len(list) > 0 {
		return list[0]
	}
	return nil
}

func recvTypeName(t ast.Expr) string {
	switch v := t.(type) {
	case *ast.StarExpr:
		return recvTypeName(v.X)
	case *ast.Ident:
		return v.Name
	}
	return ""
}

func (p *pkg) method(name, recv string) *ast.FuncDecl {
	list := p.funcs[name]
	for _, d := range list {
		got := ""
		if d.Recv != nil && len(d.Recv.List) > 0 {
			got = recvTypeName(d.Recv.List[0].Type)
		}
		if got == recv {
			return d
		}
	}
	if len(list) > 0 {
		return list[0]
	}
	return nil
}

func (sc *scope) selectorRecv(x ast.Expr) string {
	id, ok := x.(*ast.Ident)
	if !ok {
		return ""
	}
	if t := sc.paramType(id.Name); t != nil {
		return recvTypeName(t)
	}
	if t := sc.varType(id.Name); t != nil {
		return recvTypeName(t)
	}
	if call, ok := sc.varInit(id.Name).(*ast.CallExpr); ok {
		if d := sc.callee(call.Fun); d != nil && d != sc.decl && d.Type.Results != nil && len(d.Type.Results.List) > 0 {
			return recvTypeName(d.Type.Results.List[0].Type)
		}
	}
	return ""
}

func (sc *scope) callee(fun ast.Expr) *ast.FuncDecl {
	switch v := fun.(type) {
	case *ast.SelectorExpr:
		return sc.p.method(v.Sel.Name, sc.selectorRecv(v.X))
	case *ast.Ident:
		if lit, ok := sc.varInit(v.Name).(*ast.FuncLit); ok {
			return &ast.FuncDecl{Recv: sc.decl.Recv, Name: v, Type: lit.Type, Body: lit.Body}
		}
		return sc.p.method(v.Name, "")
	}
	return nil
}

func wrapperName(e ast.Expr) string {
	if call, ok := e.(*ast.CallExpr); ok {
		return calleeName(call.Fun)
	}
	return ""
}

func (p *pkg) scanWrappers() {
	init := p.fn("InitModule")
	if init == nil {
		return
	}
	merged := map[string][]string{}
	varSource := map[string]string{}
	ast.Inspect(init.Body, func(n ast.Node) bool {
		switch st := n.(type) {
		case *ast.AssignStmt:
			if len(st.Lhs) == 1 && len(st.Rhs) == 1 {
				if id, ok := st.Lhs[0].(*ast.Ident); ok {
					if call, ok := st.Rhs[0].(*ast.CallExpr); ok && isTableFunc(calleeName(call.Fun)) {
						varSource[id.Name] = calleeName(call.Fun)
					}
				}
			}
		case *ast.RangeStmt:
			src := ""
			switch x := st.X.(type) {
			case *ast.CallExpr:
				src = calleeName(x.Fun)
			case *ast.Ident:
				src = "$" + x.Name
			}
			if src == "" {
				return true
			}
			ast.Inspect(st.Body, func(m ast.Node) bool {
				switch b := m.(type) {
				case *ast.CallExpr:
					if calleeName(b.Fun) == "RegisterRpc" && len(b.Args) == 2 {
						p.formats[src] = wrapperFormat[wrapperName(b.Args[1])]
					}
				case *ast.AssignStmt:
					if ix, ok := b.Lhs[0].(*ast.IndexExpr); ok {
						if id, ok := ix.X.(*ast.Ident); ok {
							merged["$"+id.Name] = append(merged["$"+id.Name], src)
						}
					}
				}
				return true
			})
		}
		return true
	})
	for v, src := range varSource {
		if f, ok := p.formats["$"+v]; ok {
			p.formats[src] = f
			for _, other := range merged["$"+v] {
				p.formats[other] = f
			}
		}
	}
}

var rpcIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func isTableFunc(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), "rpcs")
}

func ScanRPCs(root string, modules []Module) ([]RPC, error) {
	out := []RPC{}
	seen := map[string]bool{}
	PrepareTypes(root, modules)
	for _, m := range modules {
		p, err := loadPkg(root, m)
		if err != nil {
			return nil, err
		}
		names := make([]string, 0, len(p.funcs))
		for n := range p.funcs {
			names = append(names, n)
		}
		sort.Strings(names)
		add := func(lit *ast.CompositeLit, format string) {
			if format == "" {
				format = "code"
			}
			for _, el := range lit.Elts {
				kv, ok := el.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				id, ok := stringLit(kv.Key)
				if !ok || !rpcIDPattern.MatchString(id) || seen[id] {
					continue
				}
				handler := handlerName(kv.Value)
				if handler == "" {
					continue
				}
				seen[id] = true
				out = append(out, p.describe(id, handler, format))
			}
		}
		for _, name := range names {
			for _, decl := range p.funcs[name] {
				if decl.Body == nil {
					continue
				}
				if isTableFunc(name) {
					format := p.formats[name]
					ast.Inspect(decl.Body, func(n ast.Node) bool {
						lit, ok := n.(*ast.CompositeLit)
						if !ok {
							return true
						}
						if _, ok := lit.Type.(*ast.MapType); !ok {
							return true
						}
						add(lit, format)
						return false
					})
					continue
				}
				ast.Inspect(decl.Body, func(n ast.Node) bool {
					if c, ok := n.(*ast.CallExpr); ok && calleeName(c.Fun) == "RegisterRpc" && len(c.Args) == 2 {
						if id, ok := stringLit(c.Args[0]); ok && rpcIDPattern.MatchString(id) && !seen[id] {
							h, format := c.Args[1], "code"
							if w, ok := h.(*ast.CallExpr); ok && len(w.Args) >= 1 {
								if f := wrapperFormat[calleeName(w.Fun)]; f != "" {
									format = f
								}
								h = w.Args[len(w.Args)-1]
							}
							if name := handlerName(h); name != "" {
								seen[id] = true
								out = append(out, p.describe(id, name, format))
							}
						}
						return true
					}
					rs, ok := n.(*ast.RangeStmt)
					if !ok {
						return true
					}
					lit, ok := rs.X.(*ast.CompositeLit)
					if !ok {
						return true
					}
					if _, ok := lit.Type.(*ast.MapType); !ok {
						return true
					}
					format := ""
					ast.Inspect(rs.Body, func(m ast.Node) bool {
						if c, ok := m.(*ast.CallExpr); ok && calleeName(c.Fun) == "RegisterRpc" && len(c.Args) == 2 {
							format = wrapperFormat[wrapperName(c.Args[1])]
							if format == "" {
								format = "code"
							}
						}
						return true
					})
					if format != "" {
						add(lit, format)
					}
					return true
				})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func handlerName(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return v.Sel.Name
	case *ast.CallExpr:
		return handlerName(v.Fun)
	}
	return ""
}

func (p *pkg) handlerDecl(name string) *ast.FuncDecl {
	decl := p.fn(name)
	if decl == nil || decl.Body == nil {
		return decl
	}
	if res := decl.Type.Results; res != nil && len(res.List) == 1 {
		for _, st := range decl.Body.List {
			if r, ok := st.(*ast.ReturnStmt); ok && len(r.Results) == 1 {
				if lit, ok := r.Results[0].(*ast.FuncLit); ok {
					return &ast.FuncDecl{Recv: decl.Recv, Name: decl.Name, Type: lit.Type, Body: lit.Body}
				}
			}
		}
	}
	return decl
}

func (p *pkg) describe(id, handler, format string) RPC {
	codeMode := ""
	if format != "envelope" {

		format, codeMode = "envelope", format
	}
	r := RPC{ID: id, Module: p.module.Name, Handler: handler, Auth: "session", ErrorFormat: format, CodeMode: codeMode, Request: Schema{"type": "object", "properties": Schema{}}, Response: Schema{}, Errors: []string{}}
	decl := p.handlerDecl(handler)
	if decl == nil {
		return r
	}
	r.Source = p.files[p.fn(handler)]
	sc := &scope{p: p, decl: decl, depth: 0}
	if req := sc.request(); req != nil {
		r.Request = req
	}
	r.Response = sc.returns(0)
	if len(r.Response) == 0 {
		r.Response = Schema{}
	}
	r.Errors = p.errorCodes(decl, 0, map[*ast.FuncDecl]bool{})
	return r
}

func (p *pkg) errorCodes(decl *ast.FuncDecl, depth int, seen map[*ast.FuncDecl]bool) []string {
	set := map[string]bool{}
	var walk func(d *ast.FuncDecl, depth int)
	walk = func(d *ast.FuncDecl, depth int) {
		if d == nil || d.Body == nil || seen[d] {
			return
		}
		seen[d] = true
		ast.Inspect(d.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if code, _, ok := CodeStatus(call); ok {
				set[code] = true
			}
			if depth < 2 {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == "s" {
						walk(p.fn(sel.Sel.Name), depth+1)
					}
				}
			}
			return true
		})
	}
	walk(decl, depth)
	out := make([]string, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

type scope struct {
	p     *pkg
	decl  *ast.FuncDecl
	depth int
	args  map[string]boundArg
}

type boundArg struct {
	expr ast.Expr
	sc   *scope
}

func (sc *scope) request() Schema {
	var target string
	ast.Inspect(sc.decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || target != "" {
			return target == ""
		}
		name := calleeName(call.Fun)
		if (name == "decodeStrict" || name == "Decode" || name == "Unmarshal") && len(call.Args) == 2 {
			if u, ok := call.Args[1].(*ast.UnaryExpr); ok && u.Op == token.AND {
				if id, ok := u.X.(*ast.Ident); ok {
					if name == "Unmarshal" && !mentionsPayload(call.Args[0]) {
						return true
					}
					target = id.Name
				}
			}
		}
		return true
	})
	if target == "" {
		return sc.delegatedRequest()
	}
	if s := sc.p.typedIdent(sc.decl, target); s["type"] == "object" {
		return s
	}
	if t := sc.varType(target); t != nil {
		s := sc.p.typeSchema(t, 0)
		if s["type"] == "object" {
			return s
		}
	}
	return nil
}

func (sc *scope) delegatedRequest() Schema {
	if sc.depth >= 5 {
		return nil
	}
	var out Schema
	ast.Inspect(sc.decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || out != nil {
			return out == nil
		}
		passes := false
		for _, a := range call.Args {
			if id, ok := a.(*ast.Ident); ok && id.Name == "payload" {
				passes = true
			}
		}
		if !passes {
			return true
		}
		if d := sc.callee(call.Fun); d != nil && d != sc.decl && d.Body != nil {
			out = (&scope{p: sc.p, decl: d, depth: sc.depth + 1}).request()
		}
		return true
	})
	return out
}

func mentionsPayload(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == "payload" {
			found = true
		}
		return true
	})
	return found
}

func (sc *scope) varType(name string) ast.Expr {
	var out ast.Expr
	ast.Inspect(sc.decl.Body, func(n ast.Node) bool {
		if out != nil {
			return false
		}
		switch d := n.(type) {
		case *ast.ValueSpec:
			for _, id := range d.Names {
				if id.Name == name && d.Type != nil {
					out = d.Type
				}
			}
		}
		return true
	})
	return out
}

func (sc *scope) varInit(name string) ast.Expr {
	var out ast.Expr
	ast.Inspect(sc.decl.Body, func(n ast.Node) bool {
		if out != nil {
			return false
		}
		switch d := n.(type) {
		case *ast.AssignStmt:
			if d.Tok == token.DEFINE && len(d.Lhs) == len(d.Rhs) {
				for i, l := range d.Lhs {
					if id, ok := l.(*ast.Ident); ok && id.Name == name {
						out = d.Rhs[i]
					}
				}
			} else if d.Tok == token.DEFINE && len(d.Rhs) == 1 && len(d.Lhs) > 1 {
				for i, l := range d.Lhs {
					if id, ok := l.(*ast.Ident); ok && id.Name == name {
						if i == 0 {
							out = d.Rhs[0]
						} else {
							out = &multiResult{call: d.Rhs[0], index: i}
						}
					}
				}
			}
		case *ast.ValueSpec:
			for i, id := range d.Names {
				if id.Name == name && i < len(d.Values) {
					out = d.Values[i]
				}
			}
		}
		return true
	})
	return out
}

func (sc *scope) indexAssigns(name string) Schema {
	props := Schema{}
	ast.Inspect(sc.decl.Body, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || len(as.Lhs) != len(as.Rhs) {
			return true
		}
		for i, l := range as.Lhs {
			ix, ok := l.(*ast.IndexExpr)
			if !ok {
				continue
			}
			if id, ok := ix.X.(*ast.Ident); ok && id.Name == name {
				if key, ok := stringLit(ix.Index); ok {
					got := sc.expr(as.Rhs[i], 0)
					if cur, ok := props[key].(Schema); ok {
						got = merge(cur, got)
					}
					props[key] = got
				}
			}
		}
		return true
	})
	return props
}

func (sc *scope) appended(name string) Schema {
	var item Schema
	ast.Inspect(sc.decl.Body, func(n ast.Node) bool {
		if item != nil {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || calleeName(call.Fun) != "append" || len(call.Args) < 2 {
			return true
		}
		if id, ok := call.Args[0].(*ast.Ident); ok && id.Name == name {
			if call.Ellipsis.IsValid() {
				return true
			}
			item = sc.expr(call.Args[1], 1)
		}
		return true
	})
	return item
}

func orOperands(e ast.Expr) []ast.Expr {
	if b, ok := e.(*ast.BinaryExpr); ok && b.Op == token.LOR {
		return append(orOperands(b.X), orOperands(b.Y)...)
	}
	if p, ok := e.(*ast.ParenExpr); ok {
		return orOperands(p.X)
	}
	return []ast.Expr{e}
}

func orGuardedReturns(body *ast.BlockStmt) map[*ast.ReturnStmt]bool {
	out := map[*ast.ReturnStmt]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		ifs, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		ops := orOperands(ifs.Cond)
		if len(ops) < 2 {
			return true
		}
		for _, st := range ifs.Body.List {
			r, ok := st.(*ast.ReturnStmt)
			if !ok || len(r.Results) < 2 {
				continue
			}
			last, ok := r.Results[len(r.Results)-1].(*ast.Ident)
			if !ok {
				continue
			}
			for _, op := range ops {
				if b, ok := op.(*ast.BinaryExpr); ok && b.Op == token.NEQ && isNilIdent(b.Y) {
					if id, ok := b.X.(*ast.Ident); ok && id.Name == last.Name {
						out[r] = true
					}
				}
			}
		}
		return true
	})
	return out
}

func (sc *scope) returns(depth int) Schema {
	var out Schema
	sawNil := false
	orGuarded := orGuardedReturns(sc.decl.Body)
	var walk func(n ast.Node) bool
	walk = func(n ast.Node) bool {
		switch st := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ReturnStmt:
			if len(st.Results) == 0 {
				return true
			}
			first := st.Results[0]
			if id, ok := first.(*ast.Ident); ok && id.Name == "nil" {
				if len(st.Results) == 1 || isNilIdent(st.Results[len(st.Results)-1]) || orGuarded[st] {
					sawNil = true
				}
				return true
			}
			s := sc.expr(first, depth)
			out = merge(out, s)
		}
		return true
	}
	ast.Inspect(sc.decl.Body, walk)
	if sawNil && len(out) > 0 {
		out = nullable(out)
	}
	if out == nil {
		if res := sc.decl.Type.Results; res != nil && len(res.List) > 0 {
			if t := res.List[0].Type; !isAnyType(t) {
				return sc.p.typeSchema(t, 0)
			}
		}
	}
	return out
}

func isAnyType(t ast.Expr) bool {
	switch v := t.(type) {
	case *ast.Ident:
		return v.Name == "any"
	case *ast.InterfaceType:
		return true
	case *ast.MapType:
		return isAnyType(v.Value)
	}
	return false
}

func merge(a, b Schema) Schema {
	if a == nil {
		return b
	}
	if b == nil || len(b) == 0 {
		return a
	}
	if isNullSchema(a) && !isNullSchema(b) {
		c := Schema{}
		for k, v := range b {
			c[k] = v
		}
		c["nullable"] = true
		return c
	}
	if len(a) == 0 {
		return b
	}
	if a["type"] == "object" && b["type"] == "object" {
		pa, _ := a["properties"].(Schema)
		pb, _ := b["properties"].(Schema)
		if pa == nil || pb == nil {
			return a
		}
		out := Schema{}
		for k, v := range pa {
			out[k] = v
		}
		for k, v := range pb {
			if cur, ok := out[k]; ok {
				cs, _ := cur.(Schema)
				vs, _ := v.(Schema)
				out[k] = merge(cs, vs)
			} else {
				out[k] = v
			}
		}
		res := Schema{"type": "object", "properties": out}
		if a["nullable"] == true || b["nullable"] == true {
			res["nullable"] = true
		}
		return res
	}
	if a["type"] == "array" && b["type"] == "array" {
		ia, _ := a["items"].(Schema)
		ib, _ := b["items"].(Schema)
		res := Schema{"type": "array", "items": merge(ia, ib)}
		if a["nullable"] == true || b["nullable"] == true {
			res["nullable"] = true
		}
		return res
	}
	if reflect.DeepEqual(a, b) {
		return a
	}
	if b["type"] == "null" {
		c := Schema{}
		for k, v := range a {
			c[k] = v
		}
		c["nullable"] = true
		return c
	}
	return a
}

func (sc *scope) expr(e ast.Expr, depth int) Schema {
	if depth > 12 {
		return Schema{}
	}
	s, exact := sc.p.typedExpr(e)
	if s != nil && exact && !hasHoles(s) {
		return s
	}
	u := sc.untypedExpr(e, depth)
	if lit, ok := e.(*ast.CompositeLit); ok && s != nil {
		s = sc.fillComposite(lit, s, depth)
	}
	switch {
	case s == nil:
		return u
	case exact:
		return fillHoles(s, u)
	case len(u) > 0:
		return fillHoles(u, s)
	}
	return s
}

func (sc *scope) untypedExpr(e ast.Expr, depth int) Schema {
	switch v := e.(type) {
	case *multiResult:
		return sc.multi(v)
	case *ast.ParenExpr:
		return sc.expr(v.X, depth)
	case *ast.BasicLit:
		switch v.Kind {
		case token.STRING:
			return Schema{"type": "string"}
		case token.INT:
			return Schema{"type": "integer"}
		case token.FLOAT:
			return Schema{"type": "number"}
		}
	case *ast.Ident:
		switch v.Name {
		case "true", "false":
			return Schema{"type": "boolean"}
		case "nil":
			return Schema{"type": "null"}
		}
		if b, ok := sc.args[v.Name]; ok && b.sc != nil && sc.varInit(v.Name) == nil && sc.varType(v.Name) == nil {
			if got := b.sc.expr(b.expr, depth+1); len(got) > 0 {
				return got
			}
		}
		if vt := sc.varType(v.Name); vt == nil || isAnyType(vt) {
			if got := sc.assigned(v.Name, depth); len(got) > 0 {
				init := sc.varInit(v.Name)
				var is Schema
				if init != nil {
					is = sc.expr(init, depth+1)
				}
				if init == nil || isNullSchema(is) {
					return nullable(got)
				}
				if len(is) == 0 || (is["type"] == "object" && is["properties"] == nil && got["properties"] != nil) {
					return got
				}
			}
			if vt != nil {
				if init := sc.varInit(v.Name); init != nil {
					if is := sc.expr(init, depth+1); len(is) > 0 && !isNullSchema(is) {
						if sc.assignsNil(v.Name) {
							return nullable(is)
						}
						return is
					}
				}
			}
		}
		if t := sc.varType(v.Name); t != nil {
			s := sc.p.typeSchema(t, 0)
			return sc.enrich(v.Name, s, depth)
		}
		if init := sc.varInit(v.Name); init != nil {
			s := sc.expr(init, depth+1)
			return sc.enrich(v.Name, s, depth)
		}
		if t := sc.paramType(v.Name); t != nil {
			return sc.p.typeSchema(t, 0)
		}
	case *ast.CompositeLit:
		return sc.composite(v, depth)
	case *ast.UnaryExpr:
		if v.Op == token.AND {
			return sc.expr(v.X, depth)
		}
		if v.Op == token.NOT {
			return Schema{"type": "boolean"}
		}
		return sc.expr(v.X, depth)
	case *ast.StarExpr:
		return sc.expr(v.X, depth)
	case *ast.BinaryExpr:
		switch v.Op {
		case token.EQL, token.NEQ, token.LSS, token.GTR, token.LEQ, token.GEQ, token.LAND, token.LOR:
			return Schema{"type": "boolean"}
		case token.ADD:
			l := sc.expr(v.X, depth+1)
			if l["type"] == "string" {
				return l
			}
			r := sc.expr(v.Y, depth+1)
			if r["type"] == "string" {
				return r
			}
			return numeric(l, r)
		case token.SUB, token.MUL, token.QUO, token.REM:
			return numeric(sc.expr(v.X, depth+1), sc.expr(v.Y, depth+1))
		}
	case *ast.SelectorExpr:
		return sc.selector(v, depth)
	case *ast.IndexExpr:
		base := sc.expr(v.X, depth+1)
		if props, ok := base["properties"].(Schema); ok {
			if key, ok := stringLit(v.Index); ok {
				if ps, ok := props[key].(Schema); ok {
					return ps
				}
			}
		}
		if base["type"] == "array" {
			if it, ok := base["items"].(Schema); ok {
				return it
			}
		}
		if ap, ok := base["additionalProperties"].(Schema); ok {
			return ap
		}
	case *ast.CallExpr:
		return sc.call(v, depth)
	}
	return Schema{}
}

func numeric(l, r Schema) Schema {
	if l["type"] == "number" || r["type"] == "number" {
		return Schema{"type": "number"}
	}
	if l["type"] == "integer" || r["type"] == "integer" {
		return Schema{"type": "integer"}
	}
	return Schema{"type": "number"}
}

func (sc *scope) enrich(name string, s Schema, depth int) Schema {
	if s == nil {
		s = Schema{}
	}
	if s["type"] == "object" {
		if extra := sc.allIndexAssigns(name, depth); len(extra) > 0 {
			props, _ := s["properties"].(Schema)
			merged := Schema{}
			for k, v := range props {
				merged[k] = v
			}
			for k, v := range extra {
				if cur, ok := merged[k].(Schema); ok && isNullSchema(cur) {
					if vs, ok := v.(Schema); ok {
						v = nullable(vs)
					}
				}
				merged[k] = v
			}
			c := Schema{"type": "object", "properties": merged}
			return c
		}
	}
	if s["type"] == "array" {
		if it, ok := s["items"].(Schema); !ok || len(it) == 0 || (it["type"] == "object" && it["properties"] == nil) {
			if item := sc.appended(name); item != nil {
				return Schema{"type": "array", "items": item}
			}
		}
	}
	return s
}

type multiResult struct {
	ast.Expr
	call  ast.Expr
	index int
}

var externalResults = map[string][]string{
	"ForUser": {"", "string", ""},
	"Load":    {"", ""},
}

func (sc *scope) multi(m *multiResult) Schema {
	call, ok := m.call.(*ast.CallExpr)
	if !ok {
		return Schema{}
	}
	name := calleeName(call.Fun)
	if d := sc.p.fn(name); d != nil && d.Type.Results != nil {
		i := 0
		for _, f := range d.Type.Results.List {
			n := len(f.Names)
			if n == 0 {
				n = 1
			}
			if m.index < i+n {
				return sc.p.typeSchema(f.Type, 0)
			}
			i += n
		}
	}
	if r, ok := externalResults[name]; ok && m.index < len(r) && r[m.index] != "" {
		return Schema{"type": r[m.index]}
	}
	return Schema{}
}

func (sc *scope) paramType(name string) ast.Expr {
	if sc.decl.Recv != nil {
		for _, f := range sc.decl.Recv.List {
			for _, n := range f.Names {
				if n.Name == name {
					return f.Type
				}
			}
		}
	}
	for _, f := range sc.decl.Type.Params.List {
		for _, n := range f.Names {
			if n.Name == name {
				return f.Type
			}
		}
	}
	return nil
}

func (sc *scope) composite(v *ast.CompositeLit, depth int) Schema {
	switch t := v.Type.(type) {
	case *ast.MapType:
		props := Schema{}
		keyed := false
		for _, el := range v.Elts {
			kv, ok := el.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if k, ok := stringLit(kv.Key); ok {
				props[k] = sc.expr(kv.Value, depth+1)
				keyed = true
			}
		}
		if keyed || len(v.Elts) == 0 && isAnyType(t.Value) {
			return Schema{"type": "object", "properties": props}
		}
		return Schema{"type": "object", "additionalProperties": sc.p.typeSchema(t.Value, 0)}
	case *ast.ArrayType:
		if len(v.Elts) > 0 {
			return Schema{"type": "array", "items": sc.expr(v.Elts[0], depth+1)}
		}
		return Schema{"type": "array", "items": sc.p.typeSchema(t.Elt, 0)}
	case *ast.Ident, *ast.StructType, *ast.SelectorExpr:
		return sc.p.typeSchema(t, 0)
	}
	return Schema{}
}

func (sc *scope) selector(v *ast.SelectorExpr, depth int) Schema {
	base := sc.expr(v.X, depth+1)
	if props, ok := base["properties"].(Schema); ok {
		for k, val := range props {
			if strings.EqualFold(k, v.Sel.Name) {
				if s, ok := val.(Schema); ok {
					return s
				}
			}
		}
	}
	if id, ok := v.X.(*ast.Ident); ok {
		if t := sc.varType(id.Name); t != nil {
			if f := sc.p.fieldType(t, v.Sel.Name); f != nil {
				return sc.p.typeSchema(f, 0)
			}
		}
		if t := sc.paramType(id.Name); t != nil {
			if f := sc.p.fieldType(t, v.Sel.Name); f != nil {
				return sc.p.typeSchema(f, 0)
			}
		}
	}
	return Schema{}
}

var stringCalls = map[string]bool{"Itoa": true, "Sprintf": true, "Sprint": true, "Format": true, "FormatInt": true, "FormatFloat": true, "Resolve": true, "TrimSpace": true, "ToLower": true, "ToUpper": true, "Join": true, "String": true, "Replace": true, "ReplaceAll": true, "EncodeToString": true, "Trim": true}

func (sc *scope) call(v *ast.CallExpr, depth int) Schema {
	name := calleeName(v.Fun)
	switch name {
	case "len", "cap", "Unix", "UnixMilli", "Count", "Atoi":
		return Schema{"type": "integer"}
	case "string":
		return Schema{"type": "string"}
	case "int", "int32", "int64", "uint", "uint32", "uint64":
		return Schema{"type": "integer"}
	case "float32", "float64", "Round", "Floor", "Ceil":
		return Schema{"type": "number"}
	case "bool", "HasPrefix", "Contains", "Before", "After", "Equal", "IsZero":
		return Schema{"type": "boolean"}
	case "append":
		if len(v.Args) > 0 {
			return sc.expr(v.Args[0], depth+1)
		}
	case "make":
		if len(v.Args) > 0 {
			return sc.p.typeSchema(v.Args[0], 0)
		}
	}
	if stringCalls[name] {
		return Schema{"type": "string"}
	}
	if sel, ok := v.Fun.(*ast.SelectorExpr); ok {
		if pkgID, ok := sel.X.(*ast.Ident); ok && (pkgID.Name == "time" || pkgID.Name == "strconv" || pkgID.Name == "strings" || pkgID.Name == "fmt") {
			return Schema{}
		}
		if p2, d := sc.resolve(sel); d != nil && depth < 8 && sc.depth < 6 && d != sc.decl {
			return sc.callResultIn(p2, d, v.Args, depth)
		}
	} else if id, ok := v.Fun.(*ast.Ident); ok {
		if p2, d := sc.resolve(id); d != nil && depth < 8 && sc.depth < 6 && d != sc.decl {
			return sc.callResultIn(p2, d, v.Args, depth)
		}
		if t, ok := sc.p.types[id.Name]; ok && len(v.Args) == 1 {
			return sc.p.typeSchema(t, 0)
		}
	}
	return Schema{}
}

func (sc *scope) callResultIn(p2 *pkg, d *ast.FuncDecl, args []ast.Expr, depth int) Schema {
	if p2 == sc.p && len(args) == 0 {
		return sc.callResult(d, depth)
	}
	if res := d.Type.Results; res != nil && len(res.List) > 0 {
		if t := res.List[0].Type; !isAnyType(t) {
			if s := p2.typeSchema(t, 0); len(s) > 0 && !(s["type"] == "array" && isEmptyItems(s)) && !hasHoles(s) {
				return s
			}
		}
	}
	if d.Body == nil {
		return Schema{}
	}
	inner := &scope{p: p2, decl: d, depth: sc.depth + 1, args: sc.bind(d, args)}
	return inner.returns(depth + 1)
}

func (sc *scope) bind(d *ast.FuncDecl, args []ast.Expr) map[string]boundArg {
	out := map[string]boundArg{}
	i := 0
	for _, f := range d.Type.Params.List {
		names := f.Names
		if len(names) == 0 {
			i++
			continue
		}
		for _, n := range names {
			if i < len(args) {
				out[n.Name] = boundArg{expr: args[i], sc: sc}
			}
			i++
		}
	}
	return out
}

func (sc *scope) callResult(d *ast.FuncDecl, depth int) Schema {
	if res := d.Type.Results; res != nil && len(res.List) > 0 {
		if t := res.List[0].Type; !isAnyType(t) {
			if s := sc.p.typeSchema(t, 0); len(s) > 0 && !(s["type"] == "array" && isEmptyItems(s)) {
				return s
			}
		}
	}
	if d.Body == nil {
		return Schema{}
	}
	inner := &scope{p: sc.p, decl: d, depth: sc.depth + 1}
	return inner.returns(depth + 1)
}

func isEmptyItems(s Schema) bool {
	it, ok := s["items"].(Schema)
	return !ok || len(it) == 0
}

func (p *pkg) fieldType(t ast.Expr, name string) ast.Expr {
	switch v := t.(type) {
	case *ast.StarExpr:
		return p.fieldType(v.X, name)
	case *ast.Ident:
		if def, ok := p.types[v.Name]; ok {
			return p.fieldType(def, name)
		}
	case *ast.StructType:
		for _, f := range v.Fields.List {
			for _, n := range f.Names {
				if n.Name == name {
					return f.Type
				}
			}
		}
	}
	return nil
}

func (p *pkg) typeSchema(t ast.Expr, depth int) Schema {
	if depth > 12 {
		return Schema{}
	}
	switch v := t.(type) {
	case *ast.Ident:
		switch v.Name {
		case "string":
			return Schema{"type": "string"}
		case "bool":
			return Schema{"type": "boolean"}
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			return Schema{"type": "integer"}
		case "float32", "float64":
			return Schema{"type": "number"}
		case "any", "error":
			return Schema{}
		}
		if def, ok := p.types[v.Name]; ok {
			return p.typeSchema(def, depth+1)
		}
	case *ast.StarExpr:
		s := p.typeSchema(v.X, depth+1)
		if len(s) > 0 {
			c := Schema{}
			for k, x := range s {
				c[k] = x
			}
			c["nullable"] = true
			return c
		}
		return s
	case *ast.ArrayType:
		if id, ok := v.Elt.(*ast.Ident); ok && id.Name == "byte" {
			return Schema{"type": "string"}
		}
		return Schema{"type": "array", "items": p.typeSchema(v.Elt, depth+1)}
	case *ast.MapType:
		if isAnyType(v.Value) {
			return Schema{"type": "object"}
		}
		return Schema{"type": "object", "additionalProperties": p.typeSchema(v.Value, depth+1)}
	case *ast.InterfaceType:
		return Schema{}
	case *ast.SelectorExpr:
		switch v.Sel.Name {
		case "Time":
			return Schema{"type": "string", "format": "date-time"}
		case "Text":
			return Schema{"type": "object", "additionalProperties": Schema{"type": "string"}}
		case "NullString":
			return Schema{"type": "string", "nullable": true}
		case "NullInt64", "NullInt32":
			return Schema{"type": "integer", "nullable": true}
		case "RawMessage":
			return Schema{"description": "Any JSON value."}
		}
		return Schema{}
	case *ast.StructType:
		props := Schema{}
		required := []string{}
		for _, f := range v.Fields.List {
			name, omit, skip := jsonName(f)
			if skip {
				continue
			}
			if len(f.Names) == 0 {
				emb := p.typeSchema(f.Type, depth+1)
				if ep, ok := emb["properties"].(Schema); ok {
					for k, x := range ep {
						props[k] = x
					}
				}
				continue
			}
			for _, n := range f.Names {
				if !ast.IsExported(n.Name) {
					continue
				}
				key := name
				if key == "" {
					key = n.Name
				}
				props[key] = p.typeSchema(f.Type, depth+1)
				if !omit {
					required = append(required, key)
				}
			}
		}
		out := Schema{"type": "object", "properties": props}
		_ = required
		return out
	}
	return Schema{}
}

func jsonName(f *ast.Field) (string, bool, bool) {
	if f.Tag == nil {
		return "", false, false
	}
	raw, err := strconv.Unquote(f.Tag.Value)
	if err != nil {
		return "", false, false
	}
	tag := reflect.StructTag(raw).Get("json")
	if tag == "-" {
		return "", false, true
	}
	parts := strings.Split(tag, ",")
	omit := false
	for _, p := range parts[1:] {
		if p == "omitempty" || p == "omitzero" {
			omit = true
		}
	}
	return parts[0], omit, false
}

func isNilIdent(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "nil"
}

func nullable(s Schema) Schema {
	if len(s) == 0 || s["$ref"] != nil {
		return s
	}
	c := Schema{}
	for k, v := range s {
		c[k] = v
	}
	c["nullable"] = true
	return c
}

func (sc *scope) assignsNil(name string) bool {
	found := false
	ast.Inspect(sc.decl.Body, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || as.Tok != token.ASSIGN || len(as.Lhs) != len(as.Rhs) {
			return !found
		}
		for i, l := range as.Lhs {
			if id, ok := l.(*ast.Ident); ok && id.Name == name && isNilIdent(as.Rhs[i]) {
				found = true
			}
		}
		return !found
	})
	return found
}

type assignKey struct {
	decl *ast.FuncDecl
	name string
}

var assignCache = map[assignKey]Schema{}

func (sc *scope) allIndexAssigns(name string, depth int) Schema {
	key := assignKey{sc.decl, name}
	if got, ok := assignCache[key]; ok {
		return got
	}
	assignCache[key] = Schema{}
	out := sc.allIndexAssignsUncached(name, depth)
	assignCache[key] = out
	return out
}

func (sc *scope) allIndexAssignsUncached(name string, depth int) Schema {
	props := sc.indexAssigns(name)
	if depth > 8 || sc.depth > 3 {
		return props
	}
	ast.Inspect(sc.decl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		at := -1
		for i, a := range call.Args {
			if id, ok := a.(*ast.Ident); ok && id.Name == name {
				at = i
			}
		}
		if at < 0 {
			return true
		}
		p2, d := sc.resolve(call.Fun)
		if d == nil || d == sc.decl || d.Body == nil {
			return true
		}
		param := ""
		i := 0
		for _, f := range d.Type.Params.List {
			for _, n := range f.Names {
				if i == at {
					param = n.Name
				}
				i++
			}
			if len(f.Names) == 0 {
				i++
			}
		}
		if param == "" {
			return true
		}
		inner := &scope{p: p2, decl: d, depth: sc.depth + 1, args: sc.bind(d, call.Args)}
		for k, v := range inner.allIndexAssigns(param, depth+1) {
			if _, ok := props[k]; !ok {
				props[k] = v
			}
		}
		return true
	})
	return props
}
