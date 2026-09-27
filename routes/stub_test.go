package routes_test

import (
	"context"

	"github.com/aosanya/mwanachama-backend-workspace/models"
)

type stubManager struct{}

func (*stubManager) CreateWorkbook(context.Context, models.Workbook) (models.Workbook, error) {
	panic("unused")
}
func (*stubManager) GetWorkbook(context.Context, string) (models.Workbook, error) { panic("unused") }
func (*stubManager) ListWorkbooks(context.Context, string, models.WorkbookFilter) ([]models.Workbook, error) {
	panic("unused")
}
func (*stubManager) UpdateWorkbook(context.Context, models.Workbook) (models.Workbook, error) {
	panic("unused")
}
func (*stubManager) DeleteWorkbook(context.Context, string) error { panic("unused") }
func (*stubManager) Touch(context.Context, models.Reference) (models.Reference, error) {
	panic("unused")
}
func (*stubManager) ListReferences(context.Context, string, models.ReferenceFilter) ([]models.Reference, error) {
	panic("unused")
}
func (*stubManager) DropReference(context.Context, string) error { panic("unused") }
func (*stubManager) Record(context.Context, models.Use) (models.Use, error) {
	panic("unused")
}
func (*stubManager) ListUses(context.Context, string, models.UseFilter) ([]models.Use, error) {
	panic("unused")
}
func (*stubManager) ListShortcuts(context.Context, string) ([]models.Shortcut, error) {
	panic("unused")
}
func (*stubManager) IssueShare(context.Context, string, string, string) (models.Share, string, error) {
	panic("unused")
}
func (*stubManager) ListShares(context.Context, string) ([]models.Share, error) { panic("unused") }
func (*stubManager) RetireShare(context.Context, string) error                  { panic("unused") }
func (*stubManager) OpenWithKey(context.Context, string, string) (models.Workbook, error) {
	panic("unused")
}
