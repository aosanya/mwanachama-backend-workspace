package mwanachamaworkspace_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/aosanya/mwanachama-backend-shared/specstore"
	workspace "github.com/aosanya/mwanachama-backend-workspace"
	"github.com/aosanya/mwanachama-backend-workspace/models"
)

func newLayoutStore(t *testing.T, specPath string) *specstore.Store {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	s, err := workspace.LoadSpec(specPath)
	if err != nil {
		t.Fatalf("LoadSpec: %v", err)
	}
	if err := workspace.Migrate(db, s); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	st, err := specstore.New(db, s, map[string]any{
		workspace.RoleLayout:       workspace.Layout{},
		workspace.RoleLayoutChange: workspace.LayoutChange{},
		workspace.RoleLayoutChoice: workspace.LayoutChoice{},
	})
	if err != nil {
		t.Fatalf("the layout types disagree with the blueprint: %v", err)
	}
	return st
}

func sameJSON(t *testing.T, got, want json.RawMessage) {
	t.Helper()
	var a, b bytes.Buffer
	if err := json.Compact(&a, got); err != nil {
		t.Fatalf("stored body is not JSON: %v (%s)", err, got)
	}
	if err := json.Compact(&b, want); err != nil {
		t.Fatalf("want: %v", err)
	}
	var x, y any
	_ = json.Unmarshal(a.Bytes(), &x)
	_ = json.Unmarshal(b.Bytes(), &y)
	xs, _ := json.Marshal(x)
	ys, _ := json.Marshal(y)
	if !bytes.Equal(xs, ys) {
		t.Fatalf("body changed on the way through storage:\n got %s\nwant %s", xs, ys)
	}
}

func TestLayoutTypesAgreeWithTheBlueprintInBothDomains(t *testing.T) {
	newLayoutStore(t, "spec/examples/wakala.workspace.json")
	newLayoutStore(t, "spec/examples/clinic.workspace.json")
}

func TestALayoutBodyComesBackAsItWentIn(t *testing.T) {
	st := newLayoutStore(t, "spec/examples/wakala.workspace.json")
	ctx := context.Background()
	body := json.RawMessage(`{"schema":1,"regions":{"left":{"size":340,"blocks":["pending"]}},"blocks":{"pending":{"type":"queue","action":"taskmanager.task.list","params":{"status":{"fixed":"pending"}},"row":[["id",{"age":"created_at"}],["applicant"]]}},"locks":["/blocks/pending/params/status"]}`)
	in := workspace.Layout{
		ID: specstore.NewID(), Scope: "workbook:wb1", Owner: ada, Title: "Approvals — week 40",
		Form: models.LayoutFormWhole, State: models.LayoutStateDraft, Version: 1, Body: body,
		CreatedAt: models.NowRFC3339(), UpdatedAt: models.NowRFC3339(),
	}
	if err := st.Insert(ctx, workspace.RoleLayout, in); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	var out workspace.Layout
	q := st.Query(ctx, workspace.RoleLayout).Where("id = ?", in.ID)
	if err := st.Take(q, workspace.RoleLayout, &out, gorm.ErrRecordNotFound); err != nil {
		t.Fatalf("Take: %v", err)
	}
	sameJSON(t, out.Body, body)
	if out.Form != models.LayoutFormWhole || out.State != models.LayoutStateDraft || out.Version != 1 {
		t.Fatalf("scalars changed: %+v", out)
	}
	if len(out.SetAside) != 0 {
		t.Fatalf("an empty set_aside came back as %s", out.SetAside)
	}
}

func TestAChangeKeepsItsInverse(t *testing.T) {
	st := newLayoutStore(t, "spec/examples/wakala.workspace.json")
	ctx := context.Background()
	patch := json.RawMessage(`{"blocks":{"shortcuts":{"collapsed":true}}}`)
	inverse := json.RawMessage(`{"blocks":{"shortcuts":{"collapsed":null}}}`)
	in := workspace.LayoutChange{ID: specstore.NewID(), LayoutID: "l1", Subject: ada, Via: models.LayoutViaChat, Patch: patch, Inverse: inverse, At: models.NowRFC3339()}
	if err := st.Insert(ctx, workspace.RoleLayoutChange, in); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	got, err := specstore.List[workspace.LayoutChange](st, st.Query(ctx, workspace.RoleLayoutChange).Where("layout_id = ?", "l1"), workspace.RoleLayoutChange)
	if err != nil || len(got) != 1 {
		t.Fatalf("List: %v %+v", err, got)
	}
	sameJSON(t, got[0].Patch, patch)
	sameJSON(t, got[0].Inverse, inverse)
}

func TestOnePersonHasOneLayoutChoicePerWorkbook(t *testing.T) {
	st := newLayoutStore(t, "spec/examples/wakala.workspace.json")
	ctx := context.Background()
	first := workspace.LayoutChoice{WorkbookID: "wb1", Subject: ada, LayoutID: "l1", ChosenAt: models.NowRFC3339()}
	if err := st.Insert(ctx, workspace.RoleLayoutChoice, first); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	second := first
	second.LayoutID = "l2"
	if err := st.Insert(ctx, workspace.RoleLayoutChoice, second); err == nil {
		t.Fatal("a second choice for the same person and workbook was stored beside the first")
	}
	other := first
	other.Subject = grace
	if err := st.Insert(ctx, workspace.RoleLayoutChoice, other); err != nil {
		t.Fatalf("Grace could not choose a layout in the same workbook: %v", err)
	}
}
