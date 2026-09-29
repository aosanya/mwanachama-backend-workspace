package mwanachamaworkspace_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	workspace "github.com/aosanya/mwanachama-backend-workspace"
	"github.com/aosanya/mwanachama-backend-workspace/models"
)

func newLayoutManager(t *testing.T) workspace.LayoutManager {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	s, err := workspace.LoadSpec("spec/examples/wakala.workspace.json")
	if err != nil {
		t.Fatalf("LoadSpec: %v", err)
	}
	if err := workspace.Migrate(db, s); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	lm, err := workspace.NewLayoutManager(db, s)
	if err != nil {
		t.Fatalf("NewLayoutManager: %v", err)
	}
	return lm
}

const scope = "instance:ag_7f3kq2"

const publishedBody = `{
  "schema": 1,
  "regions": {
    "left":  {"size": 340, "blocks": ["pending", "shortcuts"]},
    "main":  {"blocks": ["pins"]},
    "right": {"size": 380, "blocks": ["chat"]}
  },
  "blocks": {
    "pending": {"type": "queue", "action": "taskmanager.task.list", "params": {"status": {"fixed": "pending"}}},
    "pins": {"type": "pins", "items": {"approve": {"action": "taskmanager.task.update", "label": "Approve", "hotkey": "a"}}, "order": ["approve"]},
    "shortcuts": {"type": "shortcuts"},
    "chat": {"type": "chat", "dock": "right"}
  },
  "display": {"density": "comfortable"},
  "locks": ["/regions/right", "/blocks/pending/params/status", "/blocks/pins/items/approve"]
}`

func seedPublished(t *testing.T, lm workspace.LayoutManager, ctx context.Context) models.Layout {
	t.Helper()
	l, err := lm.CreateLayout(ctx, models.Layout{
		Scope: scope, Owner: scope, Title: "Approvals",
		Form: models.LayoutFormWhole, Body: json.RawMessage(publishedBody),
	})
	if err != nil {
		t.Fatalf("CreateLayout: %v", err)
	}
	l, err = lm.PublishLayout(ctx, l.ID, ada)
	if err != nil {
		t.Fatalf("PublishLayout: %v", err)
	}
	return l
}

func seedPatch(t *testing.T, lm workspace.LayoutManager, ctx context.Context, base models.Layout, owner string) models.Layout {
	t.Helper()
	l, err := lm.CreateLayout(ctx, models.Layout{
		Scope: scope, Owner: owner, Title: "Approvals",
		Form: models.LayoutFormPatch, BaseID: base.ID, Body: json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatalf("CreateLayout(patch): %v", err)
	}
	return l
}

func at(t *testing.T, raw json.RawMessage, path string) any {
	t.Helper()
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	cur := doc
	for _, token := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[token]
	}
	return cur
}

func patchOf(t *testing.T, lm workspace.LayoutManager, ctx context.Context, id, patch, subject string) models.LayoutApplied {
	t.Helper()
	out, err := lm.PatchLayout(ctx, id, models.LayoutProposal{
		Patch: json.RawMessage(patch), Via: models.LayoutViaChat, Subject: subject,
	})
	if err != nil {
		t.Fatalf("PatchLayout(%s): %v", patch, err)
	}
	return out
}

func TestAPatchOverAPublishedLayoutResolvesWithoutChangingIt(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)

	patchOf(t, lm, ctx, mine.ID, `{"regions":{"left":{"size":420}},"display":{"density":"compact"}}`, ada)

	got, err := lm.ReadLayout(ctx, mine.ID, ada)
	if err != nil {
		t.Fatalf("ReadLayout: %v", err)
	}
	if v := at(t, got.Body, "/regions/left/size"); v != float64(420) {
		t.Errorf("the person's own size did not win: %v", v)
	}
	if v := at(t, got.Body, "/display/density"); v != "compact" {
		t.Errorf("the person's own density did not win: %v", v)
	}
	if v := at(t, got.Body, "/blocks/chat/dock"); v != "right" {
		t.Errorf("the base's own blocks did not come through: %v", v)
	}

	published, err := lm.ReadLayout(ctx, base.ID, ada)
	if err != nil {
		t.Fatalf("ReadLayout(base): %v", err)
	}
	if v := at(t, published.Body, "/regions/left/size"); v != float64(340) {
		t.Errorf("one person's patch reached the published layout: %v", v)
	}
}

func TestAPatchUnderALockIsSetAsideRatherThanApplied(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)

	applied := patchOf(t, lm, ctx, mine.ID,
		`{"blocks":{"pending":{"params":{"status":{"fixed":"declined"}},"sort":{"field":"created_at"}}}}`, ada)

	if v := at(t, applied.Body, "/blocks/pending/params/status/fixed"); v != "pending" {
		t.Errorf("a locked parameter was overwritten: %v", v)
	}
	if v := at(t, applied.Body, "/blocks/pending/sort/field"); v != "created_at" {
		t.Errorf("the unlocked part of the same patch was dropped too: %v", v)
	}
	if len(applied.SetAside) != 1 || applied.SetAside[0] != "/blocks/pending/params/status" {
		t.Errorf("the locked path was not set aside: %+v", applied.SetAside)
	}
	for _, line := range applied.Lines {
		if line.Path == "/blocks/pending/params/status" && line.Held {
			return
		}
	}
	t.Errorf("no held line reported the lock: %+v", applied.Lines)
}

func TestSettingAsideSurvivesTheLockBeingLifted(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)
	patchOf(t, lm, ctx, mine.ID, `{"blocks":{"pending":{"params":{"status":{"fixed":"declined"}}}}}`, ada)

	held, err := lm.ReadLayout(ctx, mine.ID, ada)
	if err != nil {
		t.Fatalf("ReadLayout: %v", err)
	}
	if len(held.SetAside) != 1 {
		t.Fatalf("the locked path was not set aside to begin with: %+v", held.SetAside)
	}

	patchOf(t, lm, ctx, base.ID,
		`{"locks":["/regions/right","/blocks/pins/items/approve"]}`, grace)

	got, err := lm.ReadLayout(ctx, mine.ID, ada)
	if err != nil {
		t.Fatalf("ReadLayout: %v", err)
	}
	if v := at(t, got.Body, "/blocks/pending/params/status/fixed"); v != "declined" {
		t.Errorf("the lock was lifted and the person's value did not come back: %v", v)
	}
	if len(got.SetAside) != 0 {
		t.Errorf("the path is applied again and should no longer be set aside: %+v", got.SetAside)
	}
}

func TestProposeAnswersADiffAndStoresNothing(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)

	diff, err := lm.ProposeLayout(ctx, mine.ID, models.LayoutProposal{
		Patch: json.RawMessage(`{"regions":{"left":{"size":420}},"blocks":{"pending":{"params":{"status":{"fixed":"declined"}}}}}`),
	})
	if err != nil {
		t.Fatalf("ProposeLayout: %v", err)
	}
	var changed, held int
	for _, line := range diff.Lines {
		if line.Held {
			held++
			continue
		}
		changed++
		if line.Path == "/regions/left/size" && string(line.From) != "340" {
			t.Errorf("the diff did not say what it was coming from: %s", line.From)
		}
	}
	if changed != 1 || held != 1 {
		t.Errorf("expected one change and one held line, got %d and %d: %+v", changed, held, diff.Lines)
	}

	after, err := lm.ReadLayout(ctx, mine.ID, ada)
	if err != nil {
		t.Fatalf("ReadLayout: %v", err)
	}
	if v := at(t, after.Body, "/regions/left/size"); v != float64(340) {
		t.Errorf("proposing changed the layout: %v", v)
	}
	changes, err := lm.ListLayoutChanges(ctx, mine.ID, models.LayoutChangeFilter{})
	if err != nil {
		t.Fatalf("ListLayoutChanges: %v", err)
	}
	if len(changes) != 0 {
		t.Errorf("proposing left %d changes behind", len(changes))
	}
}

func TestEveryAppliedChangeCanBeUndone(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)

	patchOf(t, lm, ctx, mine.ID, `{"regions":{"left":{"size":420}}}`, ada)
	patchOf(t, lm, ctx, mine.ID, `{"display":{"theme":"dark"}}`, ada)

	undone, err := lm.RevertLayout(ctx, mine.ID, models.LayoutRevert{Subject: ada})
	if err != nil {
		t.Fatalf("RevertLayout: %v", err)
	}
	if v := at(t, undone.Body, "/display/theme"); v != nil {
		t.Errorf("the last change was not undone: %v", v)
	}
	if v := at(t, undone.Body, "/regions/left/size"); v != float64(420) {
		t.Errorf("undoing the last change also undid the one before it: %v", v)
	}

	again, err := lm.RevertLayout(ctx, mine.ID, models.LayoutRevert{Subject: ada})
	if err != nil {
		t.Fatalf("RevertLayout(second): %v", err)
	}
	if v := at(t, again.Body, "/regions/left/size"); v != float64(340) {
		t.Errorf("the change before it was not undone: %v", v)
	}
}

func TestUndoingTwiceDoesNotUndoTheUndo(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)

	patchOf(t, lm, ctx, mine.ID, `{"display":{"theme":"dark"}}`, ada)
	if _, err := lm.RevertLayout(ctx, mine.ID, models.LayoutRevert{Subject: ada}); err != nil {
		t.Fatalf("RevertLayout: %v", err)
	}
	_, err := lm.RevertLayout(ctx, mine.ID, models.LayoutRevert{Subject: ada})
	if !errors.Is(err, workspace.ErrLayoutChangeNotFound) {
		t.Fatalf("expected nothing left to undo, got %v", err)
	}
}

func TestOnePathCanBeResetOnItsOwn(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)
	patchOf(t, lm, ctx, mine.ID, `{"regions":{"left":{"size":420}},"display":{"theme":"dark"}}`, ada)

	out, err := lm.RevertLayout(ctx, mine.ID, models.LayoutRevert{Path: "/regions/left/size", Subject: ada})
	if err != nil {
		t.Fatalf("RevertLayout: %v", err)
	}
	if v := at(t, out.Body, "/regions/left/size"); v != float64(340) {
		t.Errorf("the one path did not go back to the base: %v", v)
	}
	if v := at(t, out.Body, "/display/theme"); v != "dark" {
		t.Errorf("resetting one path cleared another: %v", v)
	}
}

func TestClearingEverythingGoesBackToTheBase(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)
	patchOf(t, lm, ctx, mine.ID, `{"regions":{"left":{"size":420}},"display":{"theme":"dark"}}`, ada)

	out, err := lm.RevertLayout(ctx, mine.ID, models.LayoutRevert{All: true, Subject: ada})
	if err != nil {
		t.Fatalf("RevertLayout: %v", err)
	}
	if v := at(t, out.Body, "/regions/left/size"); v != float64(340) {
		t.Errorf("clearing did not go back to the base: %v", v)
	}
	if v := at(t, out.Body, "/display/theme"); v != nil {
		t.Errorf("clearing left a value behind: %v", v)
	}
}

func TestPublishingTwiceIsRefused(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)

	if _, err := lm.PublishLayout(ctx, base.ID, ada); !errors.Is(err, workspace.ErrLayoutPublished) {
		t.Fatalf("expected a second publish to be refused, got %v", err)
	}
}

func TestOnlyAWholeLayoutCanBePublished(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)

	if _, err := lm.PublishLayout(ctx, mine.ID, ada); !errors.Is(err, workspace.ErrLayoutNotWhole) {
		t.Fatalf("expected one person's patch to be unpublishable, got %v", err)
	}
}

func TestAnAdminsChangeToAPublishedLayoutReachesEverybodyOnTheirNextRead(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)
	patchOf(t, lm, ctx, mine.ID, `{"display":{"theme":"dark"}}`, ada)

	patchOf(t, lm, ctx, base.ID, `{"regions":{"left":{"size":300}},"blocks":{"pending":{"params":{"status":{"fixed":"new"}}}}}`, grace)

	got, err := lm.ReadLayout(ctx, mine.ID, ada)
	if err != nil {
		t.Fatalf("ReadLayout: %v", err)
	}
	if v := at(t, got.Body, "/regions/left/size"); v != float64(300) {
		t.Errorf("the agency's change did not reach a person who had not overridden it: %v", v)
	}
	if v := at(t, got.Body, "/blocks/pending/params/status/fixed"); v != "new" {
		t.Errorf("a locked path an admin changed did not move: %v", v)
	}
	if v := at(t, got.Body, "/display/theme"); v != "dark" {
		t.Errorf("the agency's change overwrote what the person had chosen: %v", v)
	}
}

func TestAnAdminIsNotStoppedByTheirOwnLocks(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)

	applied := patchOf(t, lm, ctx, base.ID, `{"blocks":{"pins":{"items":{"approve":{"label":"Approve it"}}}}}`, grace)
	if v := at(t, applied.Body, "/blocks/pins/items/approve/label"); v != "Approve it" {
		t.Errorf("an admin could not change a block their own layout locks: %v", v)
	}
	if len(applied.SetAside) != 0 {
		t.Errorf("an admin's own change was set aside: %+v", applied.SetAside)
	}
}

func TestReplacingADraftRefusesAVersionTheCallerHasNotSeen(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	draft, err := lm.CreateLayout(ctx, models.Layout{
		Scope: scope, Owner: ada, Title: "Intake", Body: json.RawMessage(`{"schema":1}`),
	})
	if err != nil {
		t.Fatalf("CreateLayout: %v", err)
	}
	if _, err := lm.UpdateLayout(ctx, models.Layout{ID: draft.ID, Version: 1, Title: "Intake", Body: json.RawMessage(`{"schema":1,"display":{}}`)}); err != nil {
		t.Fatalf("UpdateLayout: %v", err)
	}
	_, err = lm.UpdateLayout(ctx, models.Layout{ID: draft.ID, Version: 1, Title: "Intake", Body: json.RawMessage(`{"schema":1}`)})
	if !errors.Is(err, workspace.ErrLayoutMoved) {
		t.Fatalf("expected a stale version to be refused, got %v", err)
	}
}

func TestAPatchNoticesItsBaseMovingOn(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)

	got, err := lm.ReadLayout(ctx, mine.ID, ada)
	if err != nil {
		t.Fatalf("ReadLayout: %v", err)
	}
	if got.Stale {
		t.Fatal("a freshly made patch was already stale")
	}

	if err := bumpBase(lm, ctx, base); err != nil {
		t.Fatalf("bumpBase: %v", err)
	}
	got, err = lm.ReadLayout(ctx, mine.ID, grace)
	if err != nil {
		t.Fatalf("ReadLayout: %v", err)
	}
	if !got.Stale {
		t.Error("the base moved on and the patch did not say so")
	}

	settled, err := lm.ReadLayout(ctx, mine.ID, ada)
	if err != nil {
		t.Fatalf("ReadLayout(owner): %v", err)
	}
	if !settled.Stale {
		t.Error("the owner should still be told once")
	}
	after, err := lm.ReadLayout(ctx, mine.ID, ada)
	if err != nil {
		t.Fatalf("ReadLayout(again): %v", err)
	}
	if after.Stale {
		t.Error("the owner was told twice about the same move")
	}
}

func bumpBase(lm workspace.LayoutManager, ctx context.Context, base models.Layout) error {
	held, err := lm.ListLayouts(ctx, base.Scope, models.LayoutFilter{State: models.LayoutStatePublished})
	if err != nil {
		return err
	}
	if len(held) == 0 {
		return errors.New("no published layout")
	}
	_, err = lm.PatchLayout(ctx, base.ID, models.LayoutProposal{Patch: json.RawMessage(`{"display":{"typeScale":1}}`)})
	if errors.Is(err, workspace.ErrLayoutPublished) {
		return nil
	}
	return err
}

func TestACopyIsASnapshotAndNotALink(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)
	patchOf(t, lm, ctx, mine.ID, `{"display":{"theme":"dark"}}`, ada)

	copied, err := lm.CopyLayout(ctx, mine.ID, models.LayoutCopy{To: grace, Title: "Ada's approvals", Subject: ada})
	if err != nil {
		t.Fatalf("CopyLayout: %v", err)
	}
	if copied.Owner != grace || copied.Form != models.LayoutFormWhole || copied.State != models.LayoutStateDraft {
		t.Fatalf("a copy should be a whole draft owned by whoever got it: %+v", copied)
	}
	if v := at(t, copied.Body, "/display/theme"); v != "dark" {
		t.Errorf("the copy did not carry the resolved arrangement: %v", v)
	}

	patchOf(t, lm, ctx, mine.ID, `{"display":{"theme":"light"}}`, ada)
	after, err := lm.ReadLayout(ctx, copied.ID, grace)
	if err != nil {
		t.Fatalf("ReadLayout(copy): %v", err)
	}
	if v := at(t, after.Body, "/display/theme"); v != "dark" {
		t.Errorf("a later change followed the copy: %v", v)
	}
}

func TestTwoPeopleChooseTheirOwnLayoutInOneWorkbook(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)

	if _, err := lm.ChooseLayout(ctx, models.LayoutChoice{WorkbookID: "wb1", Subject: ada, LayoutID: mine.ID}); err != nil {
		t.Fatalf("ChooseLayout(ada): %v", err)
	}
	if _, err := lm.ChooseLayout(ctx, models.LayoutChoice{WorkbookID: "wb1", Subject: grace, LayoutID: base.ID}); err != nil {
		t.Fatalf("ChooseLayout(grace): %v", err)
	}

	hers, err := lm.ChosenLayout(ctx, "wb1", ada)
	if err != nil {
		t.Fatalf("ChosenLayout(ada): %v", err)
	}
	if hers.LayoutID != mine.ID {
		t.Errorf("Ada is at %s, expected %s", hers.LayoutID, mine.ID)
	}
	theirs, err := lm.ChosenLayout(ctx, "wb1", grace)
	if err != nil {
		t.Fatalf("ChosenLayout(grace): %v", err)
	}
	if theirs.LayoutID != base.ID {
		t.Errorf("Grace is at %s, expected %s", theirs.LayoutID, base.ID)
	}

	if _, err := lm.ChooseLayout(ctx, models.LayoutChoice{WorkbookID: "wb1", Subject: ada, LayoutID: base.ID}); err != nil {
		t.Fatalf("ChooseLayout(ada, again): %v", err)
	}
	hers, err = lm.ChosenLayout(ctx, "wb1", ada)
	if err != nil {
		t.Fatalf("ChosenLayout(ada, again): %v", err)
	}
	if hers.LayoutID != base.ID {
		t.Errorf("choosing again left Ada at %s", hers.LayoutID)
	}
}

func TestAListNeverLeavesItsScope(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	seedPublished(t, lm, ctx)
	if _, err := lm.CreateLayout(ctx, models.Layout{
		Scope: "instance:other", Owner: grace, Title: "Elsewhere", Body: json.RawMessage(`{}`),
	}); err != nil {
		t.Fatalf("CreateLayout: %v", err)
	}

	held, err := lm.ListLayouts(ctx, scope, models.LayoutFilter{})
	if err != nil {
		t.Fatalf("ListLayouts: %v", err)
	}
	for _, l := range held {
		if l.Scope != scope {
			t.Errorf("a layout from %q came back in %q's list", l.Scope, scope)
		}
	}
	if _, err := lm.ListLayouts(ctx, "", models.LayoutFilter{}); !errors.Is(err, workspace.ErrInvalidLayout) {
		t.Errorf("an empty scope should be refused rather than listing everyone's: %v", err)
	}
}

func TestOnlyTheLastTwentyChangesAreKept(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)

	for i := 0; i < 25; i++ {
		patchOf(t, lm, ctx, mine.ID, `{"display":{"typeScale":`+string(rune('1'+i%9))+`}}`, ada)
	}
	held, err := lm.ListLayoutChanges(ctx, mine.ID, models.LayoutChangeFilter{})
	if err != nil {
		t.Fatalf("ListLayoutChanges: %v", err)
	}
	if len(held) > 20 {
		t.Errorf("kept %d changes, expected at most 20", len(held))
	}
}

func TestAPatchIsRefusedWhenItIsNotAJSONObject(t *testing.T) {
	lm := newLayoutManager(t)
	ctx := context.Background()
	base := seedPublished(t, lm, ctx)
	mine := seedPatch(t, lm, ctx, base, ada)

	for _, bad := range []string{`[]`, `"left"`, `7`} {
		_, err := lm.ProposeLayout(ctx, mine.ID, models.LayoutProposal{Patch: json.RawMessage(bad)})
		if !errors.Is(err, workspace.ErrInvalidPatch) {
			t.Errorf("%s was accepted as a patch: %v", bad, err)
		}
	}
}

func TestADomainThatDeclaresNoLayoutsStillBuildsAWorkspaceManager(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	s, err := workspace.ParseSpec([]byte(`{
      "module": "workspace", "domain": "plain", "instance": "plain",
      "objects": [
        {"role": "workbook", "name": "workbook", "table": "workbooks", "description": "A container of work."},
        {"role": "reference", "name": "reference", "table": "references", "description": "Something it touched."},
        {"role": "use", "name": "use", "table": "uses", "description": "An act performed in it."},
        {"role": "shortcut", "name": "shortcut", "table": "shortcuts", "description": "An act worth offering again."},
        {"role": "share", "name": "share", "table": "shares", "description": "A key that opens it."}
      ]
    }`))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	if err := workspace.Migrate(db, s); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if _, err := workspace.NewWorkspaceManager(db, s); err != nil {
		t.Fatalf("a domain with no layouts could not build a workspace manager: %v", err)
	}
	if workspace.DeclaresLayouts(s) {
		t.Error("DeclaresLayouts said yes about a domain that fills none of the three roles")
	}
	if _, err := workspace.NewLayoutManager(db, s); !errors.Is(err, workspace.ErrLayoutsNotDeclared) {
		t.Errorf("expected a plain refusal, got %v", err)
	}
}
