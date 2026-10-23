package main

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

var preparableSQL = regexp.MustCompile(`(?is)^\s*(WITH|SELECT|INSERT|UPDATE|DELETE)\b.*\$\d`)

// TestJazebeSQLPreparesPostgres asks PostgreSQL to parse and type-check every
// complete parameterized SQL literal in this package against the migrated
// schema. It catches untyped parameters (for example bare $N arguments to the
// VARIADIC "any" jsonb_build_object), conflicting parameter types and schema
// drift that pure unit tests cannot see. Statements are prepared, never run.
// Literals joined with + are runtime-assembled fragments and are skipped.
func TestJazebeSQLPreparesPostgres(t *testing.T) {
	db := openDisposableTestDB(t)
	defer db.Close()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	checked := 0
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		fragments := map[*ast.BasicLit]bool{}
		extended := map[string]bool{}
		ast.Inspect(file, func(n ast.Node) bool {
			if as, ok := n.(*ast.AssignStmt); ok && as.Tok == token.ADD_ASSIGN {
				for _, lhs := range as.Lhs {
					if id, ok := lhs.(*ast.Ident); ok {
						extended[id.Name] = true
					}
				}
			}
			return true
		})
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.BinaryExpr:
				if node.Op == token.ADD {
					for _, side := range []ast.Expr{node.X, node.Y} {
						if lit, ok := side.(*ast.BasicLit); ok {
							fragments[lit] = true
						}
					}
				}
			case *ast.AssignStmt:
				// A literal later extended with += is only a query prefix.
				for i, lhs := range node.Lhs {
					id, ok := lhs.(*ast.Ident)
					if !ok || !extended[id.Name] || i >= len(node.Rhs) {
						continue
					}
					if lit, ok := node.Rhs[i].(*ast.BasicLit); ok {
						fragments[lit] = true
					}
				}
			}
			return true
		})
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING || fragments[lit] {
				return true
			}
			query, err := strconv.Unquote(lit.Value)
			if err != nil || !preparableSQL.MatchString(query) {
				return true
			}
			checked++
			// PREPARE runs on the server without executing the statement.
			stmt, err := conn.PrepareContext(ctx, query)
			if err != nil {
				t.Errorf("%s: %v", fset.Position(lit.Pos()), err)
				return true
			}
			stmt.Close()
			return true
		})
	}
	if checked < 100 {
		t.Fatalf("only %d SQL statements found; extraction is broken", checked)
	}
	t.Logf("prepared %d SQL statements", checked)
}
