package gen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var discoverRoots = []string{"integrations", "samples"}

func Discover(root string) ([]Module, error) {
	out := []Module{}
	for _, top := range discoverRoots {
		base := filepath.Join(root, top)
		if _, err := os.Stat(base); err != nil {
			continue
		}
		err := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				return nil
			}
			switch d.Name() {
			case "testdata", "vendor", "node_modules", "cmd":
				return filepath.SkipDir
			}
			if p == base || !hasInitModule(p) {
				return nil
			}
			rel, _ := filepath.Rel(root, p)
			rel = filepath.ToSlash(rel)
			parts := strings.Split(rel, "/")
			out = append(out, Module{Name: parts[1], Dir: rel})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Dir < out[j].Dir })
	return out, nil
}

func hasInitModule(dir string) bool {
	files, err := goFiles(dir, false)
	if err != nil {
		return false
	}
	fset := token.NewFileSet()
	for _, f := range files {
		file, err := parser.ParseFile(fset, f, nil, parser.SkipObjectResolution)
		if err != nil || file.Name.Name != "main" {
			continue
		}
		for _, d := range file.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "InitModule" {
				return true
			}
		}
	}
	return false
}
