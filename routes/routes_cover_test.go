package routes_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	workspace "github.com/aosanya/mwanachama-backend-workspace"
	"github.com/aosanya/mwanachama-backend-workspace/routes"
)

func actionsOf(rs []routes.Route) []string {
	out := make([]string, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.Action)
	}
	return out
}

func TestSplit_OnlyTheKeyedOpenIsAnonymousAndEverythingElseIsGated(t *testing.T) {
	split := routes.Split((*stubManager)(nil))
	if got := actionsOf(split.Anonymous); len(got) != 1 || got[0] != "workspace.workbook.open" {
		t.Fatalf("anonymous routes = %v, want only workspace.workbook.open", got)
	}
	if len(split.Gated) == 0 {
		t.Fatal("no gated routes: every address would be public")
	}
	for _, a := range actionsOf(split.Gated) {
		if a == "workspace.workbook.open" {
			t.Error("the keyed open appears in the gated half as well")
		}
	}
	built, err := routes.Build((*stubManager)(nil))
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(split.Anonymous)+len(split.Gated) != len(built) {
		t.Errorf("split holds %d routes, Build made %d", len(split.Anonymous)+len(split.Gated), len(built))
	}
}

func TestPublicAndOwnerRoutes_AgreeWithSplit(t *testing.T) {
	m := (*stubManager)(nil)
	if got := actionsOf(routes.PublicRoutes(m)); len(got) != 1 || got[0] != "workspace.workbook.open" {
		t.Errorf("PublicRoutes = %v", got)
	}
	owner := routes.OwnerRoutes(m)
	if len(owner) != len(routes.Split(m).Gated) {
		t.Errorf("OwnerRoutes has %d routes, the gated half has %d", len(owner), len(routes.Split(m).Gated))
	}
	if len(routes.Routes(m)) != len(owner)+1 {
		t.Errorf("Routes = %d, want owner routes plus the one public open", len(routes.Routes(m)))
	}
}

func TestOwnerRoutesWith_DeniedActionAnswers403BeforeTouchingTheManager(t *testing.T) {
	var asked []string
	deny := func(_ context.Context, action string) error {
		asked = append(asked, action)
		return errors.New("no grant")
	}
	owner := routes.OwnerRoutesWith((*stubManager)(nil), deny)
	if len(owner) == 0 {
		t.Fatal("no owner routes")
	}
	first := owner[0]
	rec := httptest.NewRecorder()
	first.Handler(rec, httptest.NewRequest(first.Method, "/x", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "no grant") {
		t.Errorf("the authorizer's own reason reached the caller: %s", rec.Body.String())
	}
	if len(asked) != 1 || asked[0] != first.Action {
		t.Errorf("authorizer was asked %v, want [%s]", asked, first.Action)
	}
}

func TestRoutesWith_EveryRouteAsksTheAuthorizerForItsOwnAction(t *testing.T) {
	for _, r := range routes.RoutesWith((*stubManager)(nil), func(_ context.Context, action string) error {
		return errors.New("deny " + action)
	}) {
		rec := httptest.NewRecorder()
		r.Handler(rec, httptest.NewRequest(r.Method, "/x", nil))
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s answered %d under a denying authorizer, want 403", r.Action, rec.Code)
		}
	}
}

func TestToolsWith_DeniedToolRefusesAndTheOpenIsNotOffered(t *testing.T) {
	deny := func(context.Context, string) error { return errors.New("no") }
	tools := routes.ToolsWith((*stubManager)(nil), deny)
	if len(tools) == 0 {
		t.Fatal("no tools")
	}
	for _, tool := range tools {
		if tool.Action == "workspace.workbook.open" {
			t.Errorf("tool %s is the anonymous open", tool.Name)
		}
	}
	if len(routes.Tools((*stubManager)(nil))) != len(tools) {
		t.Error("an authorizer changed which tools are offered")
	}
}

func TestLayoutRoutesWith_DeniedActionAnswers403(t *testing.T) {
	deny := func(context.Context, string) error { return errors.New("no") }
	rs := routes.LayoutRoutesWith(&stubLayouts{}, deny)
	if len(rs) != len(routes.LayoutRoutes(&stubLayouts{})) || len(rs) == 0 {
		t.Fatalf("LayoutRoutesWith made %d routes", len(rs))
	}
	for _, r := range rs {
		rec := httptest.NewRecorder()
		r.Handler(rec, httptest.NewRequest(r.Method, "/x", nil))
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s answered %d, want 403", r.Action, rec.Code)
		}
	}
}

func TestLayoutTools_MatchTheLayoutRoutes(t *testing.T) {
	tools := routes.LayoutTools(&stubLayouts{})
	withAuth := routes.LayoutToolsWith(&stubLayouts{}, func(context.Context, string) error { return nil })
	if len(tools) == 0 || len(tools) != len(withAuth) {
		t.Fatalf("tools %d, with authorizer %d", len(tools), len(withAuth))
	}
	if len(tools) != len(routes.LayoutRoutes(&stubLayouts{})) {
		t.Errorf("%d tools for %d routes: the two surfaces disagree", len(tools), len(routes.LayoutRoutes(&stubLayouts{})))
	}
}

var _ workspace.WorkspaceManager = (*stubManager)(nil)
