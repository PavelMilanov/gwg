package buildtests

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Project commands have one literal/factory per root.go, located under their
// exact CLI name. Cobra-generated help/completion commands are library-owned.
func TestCommandArchitecture(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	composition := map[string]string{
		"internal/gwucli": "gwu", "internal/gwrcli": "gwr",
	}
	checked := 0
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		alias := ""
		for _, imp := range file.Imports {
			if imp.Path.Value == `"github.com/spf13/cobra"` {
				alias = "cobra"
				if imp.Name != nil {
					alias = imp.Name.Name
				}
			}
		}
		if alias == "" {
			return nil
		}
		var commands []*ast.CompositeLit
		ast.Inspect(file, func(node ast.Node) bool {
			lit, ok := node.(*ast.CompositeLit)
			if !ok {
				return true
			}
			sel, ok := lit.Type.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			name, ok := sel.X.(*ast.Ident)
			if ok && name.Name == alias && sel.Sel.Name == "Command" {
				commands = append(commands, lit)
			}
			return true
		})
		if len(commands) == 0 {
			return nil
		}
		checked++
		if filepath.Base(path) != "root.go" || len(commands) != 1 {
			t.Errorf("%s: expected one command in root.go, got %d", path, len(commands))
			return nil
		}
		factory := false
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "NewCommand" {
				factory = true
			}
		}
		if !factory {
			t.Errorf("%s: missing NewCommand", path)
		}
		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		expected := filepath.Base(filepath.Dir(path))
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) == 3 && parts[0] == "modules" && parts[2] == "command" {
			expected = parts[1]
		}
		if name, ok := composition[filepath.ToSlash(rel)]; ok {
			expected = name
		}
		use := ""
		for _, elt := range commands[0].Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != "Use" {
				continue
			}
			if value, ok := kv.Value.(*ast.BasicLit); ok {
				use, _ = strconv.Unquote(value.Value)
			}
		}
		fields := strings.Fields(use)
		if len(fields) == 0 || fields[0] != expected {
			t.Errorf("%s: command Use %q must match folder %q", path, use, expected)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatal("no command sources checked")
	}
}
