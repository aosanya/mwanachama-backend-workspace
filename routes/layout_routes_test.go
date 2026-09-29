package routes_test

import (
	"sort"
	"strings"
	"testing"

	workspace "github.com/aosanya/mwanachama-backend-workspace"
	"github.com/aosanya/mwanachama-backend-workspace/routes"
)

func TestEveryDeclaredLayoutOperationReachesAManagerMethod(t *testing.T) {
	built, err := routes.BuildLayoutRoutes(&stubLayouts{}, nil)
	if err != nil {
		t.Fatalf("BuildLayoutRoutes: %v", err)
	}
	want := []string{
		"workspace.layout.choose",
		"workspace.layout.chosen",
		"workspace.layout.create",
		"workspace.layout.history",
		"workspace.layout.list",
		"workspace.layout.patch",
		"workspace.layout.propose",
		"workspace.layout.publish",
		"workspace.layout.read",
		"workspace.layout.revert",
		"workspace.layout.share",
		"workspace.layout.update",
	}
	got := make([]string, 0, len(built))
	for _, rt := range built {
		got = append(got, rt.Action)
	}
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("the layout operations dispatch to:\n got %v\nwant %v", got, want)
	}
}

func TestLayoutOperationsPublishAsTools(t *testing.T) {
	tools, err := routes.BuildLayoutTools(&stubLayouts{}, nil)
	if err != nil {
		t.Fatalf("BuildLayoutTools: %v", err)
	}
	if len(tools) == 0 {
		t.Fatal("no layout tools were published")
	}
	for _, tool := range tools {
		if tool.Description == "" {
			t.Errorf("tool %q has nothing to say for itself", tool.Action)
		}
	}
}

func TestTheLayoutRoutesDoNotCollideWithTheWorkbookOnes(t *testing.T) {
	plain, err := routes.Build(&stubManager{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	layout, err := routes.BuildLayoutRoutes(&stubLayouts{}, nil)
	if err != nil {
		t.Fatalf("BuildLayoutRoutes: %v", err)
	}
	held := map[string]bool{}
	for _, rt := range plain {
		held[rt.Method+" "+rt.Path] = true
	}
	for _, rt := range layout {
		if held[rt.Method+" "+rt.Path] {
			t.Errorf("%s %s is answered twice", rt.Method, rt.Path)
		}
	}
}

func TestTheTwoSpecsClaimNoActionTwice(t *testing.T) {
	plain, err := routes.Build(&stubManager{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	layout, err := routes.BuildLayoutRoutes(&stubLayouts{}, nil)
	if err != nil {
		t.Fatalf("BuildLayoutRoutes: %v", err)
	}
	held := map[string]bool{}
	for _, rt := range plain {
		held[rt.Action] = true
	}
	for _, rt := range layout {
		if held[rt.Action] {
			t.Errorf("%s is claimed by both specs", rt.Action)
		}
	}
}

var _ workspace.LayoutManager = (*stubLayouts)(nil)
