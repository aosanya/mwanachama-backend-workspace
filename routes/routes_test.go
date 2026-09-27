package routes_test

import (
	"strings"
	"testing"

	workspace "github.com/aosanya/mwanachama-backend-workspace"
	"github.com/aosanya/mwanachama-backend-workspace/routes"
)

func TestBuild_EveryDeclaredOperationMatchesTheManager(t *testing.T) {
	built, err := routes.Build((*stubManager)(nil))
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(built) == 0 {
		t.Fatal("no routes were generated")
	}
	seen := map[string]bool{}
	for _, r := range built {
		if r.Action == "" {
			t.Errorf("route %s %s carries no action, so nothing can gate it", r.Method, r.Path)
		}
		if !strings.HasPrefix(r.Action, "workspace.") {
			t.Errorf("action %q is not namespaced to this module", r.Action)
		}
		key := r.Method + " " + r.Path
		if seen[key] {
			t.Errorf("two operations claim %s", key)
		}
		seen[key] = true
	}
}

func TestTools_AreBuiltAndTheSharedOpenIsNotAmongThem(t *testing.T) {
	tools, err := routes.BuildTools((*stubManager)(nil))
	if err != nil {
		t.Fatalf("BuildTools: %v", err)
	}
	if len(tools) == 0 {
		t.Fatal("no tools were generated")
	}
	for _, tool := range tools {
		if tool.Action == "workspace.workbook.open" {
			t.Error("opening with a key is the one anonymous operation; it has no place in an operator's toolbox")
		}
		if tool.Description == "" {
			t.Errorf("tool %q has no description", tool.Name)
		}
	}
}

func TestFields_DescribeEveryDeclaredField(t *testing.T) {
	fields, err := routes.Fields()
	if err != nil {
		t.Fatalf("Fields: %v", err)
	}
	b, err := workspace.Blueprint()
	if err != nil {
		t.Fatalf("Blueprint: %v", err)
	}
	for _, o := range b.Objects {
		for _, f := range o.Fields {
			found := false
			for key := range fields {
				if strings.HasSuffix(key, "."+f.Name) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("no field doc reaches %s.%s, so a tool's schema would describe it as nothing", o.Role, f.Name)
			}
		}
	}
}

func TestAnonymousActions_IsOnlyTheKeyedOpen(t *testing.T) {
	if len(routes.AnonymousActions) != 1 || routes.AnonymousActions[0] != "workspace.workbook.open" {
		t.Errorf("AnonymousActions = %v; only opening with a key has no caller to authorize", routes.AnonymousActions)
	}
}
