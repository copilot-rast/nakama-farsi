package gen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var codePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,79}$`)

type Module struct {
	Name string
	Dir  string
}

var DefaultModules = []Module{
	{Name: "base", Dir: "integrations/base"},
	{Name: "economy", Dir: "integrations/economy"},
	{Name: "publication", Dir: "integrations/publication"},
	{Name: "leaderboardx", Dir: "integrations/leaderboardx"},
	{Name: "content", Dir: "integrations/content"},
	{Name: "insights", Dir: "integrations/insights"},
	{Name: "quiz", Dir: "integrations/quiz"},
	{Name: "trophy", Dir: "integrations/trophy"},
	{Name: "clan", Dir: "integrations/clan"},
	{Name: "cup", Dir: "integrations/cup"},
}

type CodeInfo struct {
	Status  int      `json:"status"`
	Modules []string `json:"modules"`
}

var fixedStatus = map[string]int{
	"badRequest": http.StatusBadRequest, "BadRequest": http.StatusBadRequest,
	"notFound": http.StatusNotFound, "NotFound": http.StatusNotFound,
	"conflict": http.StatusConflict, "Conflict": http.StatusConflict,
	"Unauthorized": http.StatusUnauthorized, "Forbidden": http.StatusForbidden,
	"Unavailable": http.StatusServiceUnavailable, "TooManyRequests": http.StatusTooManyRequests,
}

var statusFirst = map[string]bool{"fail": true, "New": true}

var grpcByName = map[string]grpcCode{
	"InvalidArgument": codeInvalidArgument, "NotFound": codeNotFound, "AlreadyExists": codeAlreadyExists,
	"PermissionDenied": codePermissionDenied, "Unauthenticated": codeUnauthenticated, "FailedPrecondition": codeFailedPrecondition,
	"ResourceExhausted": codeResourceExhausted, "Unavailable": codeUnavailable, "Internal": codeInternal,
	"DeadlineExceeded": codeDeadlineExceeded, "Canceled": codeCanceled, "Aborted": codeAborted, "OutOfRange": codeOutOfRange,
	"Unimplemented": codeUnimplemented, "Unknown": codeUnknown,
}

func calleeName(fn ast.Expr) string {
	switch f := fn.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		return f.Sel.Name
	}
	return ""
}

func stringLit(e ast.Expr) (string, bool) {
	bl, ok := e.(*ast.BasicLit)
	if !ok || bl.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(bl.Value)
	return s, err == nil
}

func httpStatusExpr(e ast.Expr) int {
	switch v := e.(type) {
	case *ast.BasicLit:
		n, _ := strconv.Atoi(v.Value)
		return n
	case *ast.SelectorExpr:
		if id, ok := v.X.(*ast.Ident); ok && id.Name == "http" {
			return httpStatusByName[v.Sel.Name]
		}
	}
	return 0
}

func grpcExpr(e ast.Expr) (grpcCode, bool) {
	switch v := e.(type) {
	case *ast.CallExpr:
		if len(v.Args) == 1 {
			return grpcExpr(v.Args[0])
		}
	case *ast.SelectorExpr:
		if c, ok := grpcByName[v.Sel.Name]; ok {
			return c, true
		}
	case *ast.BasicLit:
		n, err := strconv.Atoi(v.Value)
		return grpcCode(n), err == nil
	}
	return 0, false
}

func CodeStatus(call *ast.CallExpr) (string, int, bool) {
	name := calleeName(call.Fun)
	if name == "" || len(call.Args) == 0 {
		return "", 0, false
	}
	if st, ok := fixedStatus[name]; ok {
		code, ok := stringLit(call.Args[0])
		if !ok || !codePattern.MatchString(code) {
			return "", 0, false
		}
		return code, st, true
	}
	if statusFirst[name] && len(call.Args) >= 2 {
		code, ok := stringLit(call.Args[1])
		if !ok || !codePattern.MatchString(code) {
			return "", 0, false
		}
		return code, httpStatusExpr(call.Args[0]), true
	}
	if name == "NewError" && len(call.Args) == 2 {
		code, ok := stringLit(call.Args[0])
		if !ok || !codePattern.MatchString(code) {
			return "", 0, false
		}
		st := 0
		if c, ok := grpcExpr(call.Args[1]); ok {
			st = statusFromGRPC(c)
		}
		return code, st, true
	}
	return "", 0, false
}

func goFiles(dir string, recursive bool) ([]string, error) {
	var out []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != dir && (!recursive || d.Name() == "cmd" || d.Name() == "testdata" || d.Name() == "gen") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			out = append(out, p)
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

func ScanCodes(root string, modules []Module) (map[string]CodeInfo, error) {
	type tally struct {
		statuses map[int]int
		modules  map[string]bool
	}
	all := map[string]*tally{}
	fset := token.NewFileSet()
	for _, m := range modules {
		files, err := goFiles(filepath.Join(root, m.Dir), true)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, f := range files {
			file, err := parser.ParseFile(fset, f, nil, 0)
			if err != nil {
				return nil, err
			}
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				code, st, ok := CodeStatus(call)
				if !ok {
					return true
				}
				t := all[code]
				if t == nil {
					t = &tally{statuses: map[int]int{}, modules: map[string]bool{}}
					all[code] = t
				}
				if st > 0 {
					t.statuses[st]++
				}
				t.modules[m.Name] = true
				return true
			})
		}
	}
	out := map[string]CodeInfo{}
	for code, t := range all {
		best, bestN := http.StatusBadRequest, 0
		for st, n := range t.statuses {
			if n > bestN || (n == bestN && st < best) {
				best, bestN = st, n
			}
		}
		mods := []string{}
		for m := range t.modules {
			mods = append(mods, m)
		}
		sort.Strings(mods)
		out[code] = CodeInfo{Status: best, Modules: mods}
	}
	return out, nil
}

var httpStatusByName = map[string]int{
	"StatusBadRequest": 400, "StatusUnauthorized": 401, "StatusPaymentRequired": 402, "StatusForbidden": 403,
	"StatusNotFound": 404, "StatusMethodNotAllowed": 405, "StatusRequestTimeout": 408, "StatusConflict": 409,
	"StatusGone": 410, "StatusPreconditionFailed": 412, "StatusRequestEntityTooLarge": 413, "StatusUnsupportedMediaType": 415,
	"StatusUnprocessableEntity": 422, "StatusLocked": 423, "StatusTooManyRequests": 429, "StatusInternalServerError": 500,
	"StatusNotImplemented": 501, "StatusBadGateway": 502, "StatusServiceUnavailable": 503, "StatusGatewayTimeout": 504,
}

type grpcCode int

const (
	codeCanceled           grpcCode = 1
	codeUnknown            grpcCode = 2
	codeInvalidArgument    grpcCode = 3
	codeDeadlineExceeded   grpcCode = 4
	codeNotFound           grpcCode = 5
	codeAlreadyExists      grpcCode = 6
	codePermissionDenied   grpcCode = 7
	codeResourceExhausted  grpcCode = 8
	codeFailedPrecondition grpcCode = 9
	codeAborted            grpcCode = 10
	codeOutOfRange         grpcCode = 11
	codeUnimplemented      grpcCode = 12
	codeInternal           grpcCode = 13
	codeUnavailable        grpcCode = 14
	codeUnauthenticated    grpcCode = 16
)

func statusFromGRPC(c grpcCode) int {
	switch c {
	case codeInvalidArgument, codeOutOfRange, codeFailedPrecondition:
		return http.StatusBadRequest
	case codeUnauthenticated:
		return http.StatusUnauthorized
	case codePermissionDenied:
		return http.StatusForbidden
	case codeNotFound:
		return http.StatusNotFound
	case codeAlreadyExists, codeAborted:
		return http.StatusConflict
	case codeResourceExhausted:
		return http.StatusTooManyRequests
	case codeUnavailable:
		return http.StatusServiceUnavailable
	case codeDeadlineExceeded:
		return http.StatusGatewayTimeout
	case codeCanceled:
		return 499
	case codeUnimplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}
