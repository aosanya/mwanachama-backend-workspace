package mwanachamaworkspace_test

import (
	"context"
	"errors"
	"testing"

	workspace "github.com/aosanya/mwanachama-backend-workspace"
)

const (
	ada   = "0f3d1b8e-5a21-4c77-9f4e-2b6c0a11d3e9"
	grace = "7c2a9f41-63be-4d0a-8e15-4f9ab2c77d10"
)

func seedWorkbook(t *testing.T, mgr workspace.WorkspaceManager, ctx context.Context, owner string) workspace.Workbook {
	t.Helper()
	w, err := mgr.CreateWorkbook(ctx, workspace.Workbook{Owner: owner, Title: "Q4 planning"})
	if err != nil {
		t.Fatalf("CreateWorkbook: %v", err)
	}
	return w
}

func TestListWorkbooks_NeverReturnsSomebodyElses(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()
	seedWorkbook(t, mgr, ctx, ada)
	seedWorkbook(t, mgr, ctx, grace)

	mine, err := mgr.ListWorkbooks(ctx, ada, workspace.WorkbookFilter{})
	if err != nil {
		t.Fatalf("ListWorkbooks: %v", err)
	}
	if len(mine) != 1 || mine[0].Owner != ada {
		t.Fatalf("expected only Ada's, got %+v", mine)
	}
}

func TestListWorkbooks_RefusesAnEmptyOwnerRatherThanListingEveryone(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()
	seedWorkbook(t, mgr, ctx, ada)
	seedWorkbook(t, mgr, ctx, grace)

	if _, err := mgr.ListWorkbooks(ctx, "", workspace.WorkbookFilter{}); !errors.Is(err, workspace.ErrInvalidWorkbook) {
		t.Errorf("an empty owner must be refused, not treated as a wildcard; got %v", err)
	}
}

func TestDeleteWorkbook_LeavesItOutOfEveryListing(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()
	w := seedWorkbook(t, mgr, ctx, ada)

	if err := mgr.DeleteWorkbook(ctx, w.ID); err != nil {
		t.Fatalf("DeleteWorkbook: %v", err)
	}
	mine, err := mgr.ListWorkbooks(ctx, ada, workspace.WorkbookFilter{})
	if err != nil {
		t.Fatalf("ListWorkbooks: %v", err)
	}
	if len(mine) != 0 {
		t.Errorf("a retired workbook should not be listed, got %+v", mine)
	}
}

func TestTouch_IsOneRowPerThingAndMovesItsLastSeen(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()
	w := seedWorkbook(t, mgr, ctx, ada)

	ref := workspace.Reference{
		WorkbookID: w.ID, Module: "catalog", Item: "wakalaagency", Label: "Mwanachama-Wakala",
	}
	first, err := mgr.Touch(ctx, ref)
	if err != nil {
		t.Fatalf("Touch: %v", err)
	}
	again, err := mgr.Touch(ctx, ref)
	if err != nil {
		t.Fatalf("Touch again: %v", err)
	}
	if again.ID != first.ID {
		t.Errorf("touching the same thing twice made two rows: %s then %s", first.ID, again.ID)
	}

	held, err := mgr.ListReferences(ctx, w.ID, workspace.ReferenceFilter{})
	if err != nil {
		t.Fatalf("ListReferences: %v", err)
	}
	if len(held) != 1 {
		t.Fatalf("expected one reference, got %+v", held)
	}
}

func TestRecord_EarnsAShortcutAndCountsIt(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()
	w := seedWorkbook(t, mgr, ctx, ada)

	shortcuts, err := mgr.ListShortcuts(ctx, w.ID)
	if err != nil {
		t.Fatalf("ListShortcuts: %v", err)
	}
	if len(shortcuts) != 0 {
		t.Fatalf("a new workbook offers nothing until something is done in it, got %+v", shortcuts)
	}

	for i := 0; i < 3; i++ {
		if _, err := mgr.Record(ctx, workspace.Use{
			WorkbookID: w.ID, Action: "agency.goal.create", Subject: ada,
		}); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}

	shortcuts, err = mgr.ListShortcuts(ctx, w.ID)
	if err != nil {
		t.Fatalf("ListShortcuts: %v", err)
	}
	if len(shortcuts) != 1 {
		t.Fatalf("expected one earned shortcut, got %+v", shortcuts)
	}
	if shortcuts[0].Uses != 3 {
		t.Errorf("uses = %d, want 3", shortcuts[0].Uses)
	}
	if shortcuts[0].Action != "agency.goal.create" {
		t.Errorf("action = %q", shortcuts[0].Action)
	}
	if shortcuts[0].Label != "Create goal" {
		t.Errorf("label = %q, want a label built from the action's own words", shortcuts[0].Label)
	}
}

func TestRecord_RefusesAMalformedAction(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()
	w := seedWorkbook(t, mgr, ctx, ada)

	for _, bad := range []string{"", "nodots", "Agency.Goal.Create", "agency..create", "agency.goal.create; drop table uses"} {
		if _, err := mgr.Record(ctx, workspace.Use{WorkbookID: w.ID, Action: bad}); !errors.Is(err, workspace.ErrInvalidAction) {
			t.Errorf("Record(%q) error = %v, want ErrInvalidAction", bad, err)
		}
	}
}

func TestShare_OpensOnlyWithALiveKey(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()
	w := seedWorkbook(t, mgr, ctx, ada)

	share, key, err := mgr.IssueShare(ctx, w.ID, "for Grace", ada)
	if err != nil {
		t.Fatalf("IssueShare: %v", err)
	}
	if key == "" {
		t.Fatal("a share must return its key once")
	}
	if share.KeyHash == key {
		t.Error("the stored value must be a hash, not the key itself")
	}

	opened, err := mgr.OpenWithKey(ctx, w.ID, key)
	if err != nil || opened.ID != w.ID {
		t.Fatalf("a live key should open the workbook, got %+v / %v", opened, err)
	}
	if _, err := mgr.OpenWithKey(ctx, w.ID, "not-the-key"); !errors.Is(err, workspace.ErrWorkbookNotFound) {
		t.Errorf("a wrong key must answer the same as an unknown workbook, got %v", err)
	}

	if err := mgr.RetireShare(ctx, share.ID); err != nil {
		t.Fatalf("RetireShare: %v", err)
	}
	if _, err := mgr.OpenWithKey(ctx, w.ID, key); !errors.Is(err, workspace.ErrWorkbookNotFound) {
		t.Errorf("a retired key must stop opening it, got %v", err)
	}

	held, err := mgr.ListShares(ctx, w.ID)
	if err != nil {
		t.Fatalf("ListShares: %v", err)
	}
	if len(held) != 1 || held[0].Status != workspace.ShareStatusRetired {
		t.Errorf("a retired key is kept for the trail, got %+v", held)
	}
}
