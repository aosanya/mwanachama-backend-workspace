package mwanachamaworkspace_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	workspace "github.com/aosanya/mwanachama-backend-workspace"
)

var domainWords = []string{
	"agency", "agencies", "studio", "operator", "catalog", "archetype",
	"clinician", "patient", "chart", "ward", "handover",
	"tenant", "employee", "member",
}

func carriesDomainWord(name string) (string, bool) {
	lower := strings.ToLower(name)
	for _, w := range domainWords {
		if strings.Contains(lower, w) {
			return w, true
		}
	}
	return "", false
}

func TestNoDomainWordsInIdentifiers(t *testing.T) {
	for _, dir := range []string{"models", "routes", "."} {
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
			return !strings.HasSuffix(fi.Name(), "_test.go")
		}, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", dir, err)
		}
		seen := map[string]bool{}
		for _, pkg := range pkgs {
			for path, file := range pkg.Files {
				ast.Inspect(file, func(n ast.Node) bool {
					id, ok := n.(*ast.Ident)
					if !ok {
						return true
					}
					w, bad := carriesDomainWord(id.Name)
					if !bad {
						return true
					}
					key := filepath.Base(path) + ":" + id.Name
					if seen[key] {
						return true
					}
					seen[key] = true
					t.Errorf("%s: identifier %q carries the domain word %q — a workbook remembers what it touched without learning what it was",
						filepath.Base(path), id.Name, w)
					return true
				})
			}
		}
	}
}

func TestNoDomainWordsInTheBlueprint(t *testing.T) {
	b, err := workspace.Blueprint()
	if err != nil {
		t.Fatalf("blueprint: %v", err)
	}
	report := func(kind, name string) {
		if w, bad := carriesDomainWord(name); bad {
			t.Errorf("blueprint %s %q carries the domain word %q — it belongs in a domain's own spec", kind, name, w)
		}
	}
	for _, o := range b.Objects {
		report("role", o.Role)
		if o.Name != "" || o.Table != "" {
			t.Errorf("blueprint role %q names a table, which is the domain's to choose", o.Role)
		}
		for _, f := range o.Fields {
			report("field", o.Role+"."+f.Name)
			for _, v := range f.Values {
				report("stored value", o.Role+"."+f.Name+"="+v)
			}
		}
		for _, idx := range o.Indexes {
			report("index", o.Role+"."+idx.Name)
		}
	}
}

func TestBothExampleDomainsLoadAndDisagreeAboutEveryNoun(t *testing.T) {
	a, err := workspace.LoadSpec("spec/examples/wakala.workspace.json")
	if err != nil {
		t.Fatalf("LoadSpec(wakala): %v", err)
	}
	b, err := workspace.LoadSpec("spec/examples/clinic.workspace.json")
	if err != nil {
		t.Fatalf("LoadSpec(clinic): %v", err)
	}
	for _, role := range []string{"workbook", "reference", "use", "shortcut", "share", "layout", "layout_change", "layout_choice"} {
		x, ok := a.ByRole(role)
		if !ok {
			t.Fatalf("wakala fills no %q", role)
		}
		y, ok := b.ByRole(role)
		if !ok {
			t.Fatalf("clinic fills no %q", role)
		}
		if x.Name == y.Name {
			t.Errorf("role %q is called %q in both domains — the example demonstrates nothing", role, x.Name)
		}
		if a.TableFor(x) == b.TableFor(y) {
			t.Errorf("role %q lands in the same table for both domains (%s)", role, a.TableFor(x))
		}
	}
}

func TestEveryEmittedNameFitsPostgres(t *testing.T) {
	for _, path := range []string{"spec/examples/wakala.workspace.json", "spec/examples/clinic.workspace.json"} {
		s, err := workspace.LoadSpec(path)
		if err != nil {
			t.Fatalf("LoadSpec(%s): %v", path, err)
		}
		for _, stmt := range s.DDL("postgres") {
			for _, word := range strings.Fields(strings.ReplaceAll(stmt, "(", " ")) {
				if strings.HasPrefix(word, s.Instance+"_") && len(word) > 63 {
					t.Errorf("%s emits %q at %d bytes; Postgres truncates at 63", path, word, len(word))
				}
			}
		}
	}
}
