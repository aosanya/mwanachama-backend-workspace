package routes_test

import (
	"context"

	"github.com/aosanya/mwanachama-backend-workspace/models"
)

type stubLayouts struct{}

func (*stubLayouts) ListLayouts(context.Context, string, models.LayoutFilter) ([]models.Layout, error) {
	panic("unused")
}
func (*stubLayouts) ReadLayout(context.Context, string, string) (models.ResolvedLayout, error) {
	panic("unused")
}
func (*stubLayouts) CreateLayout(context.Context, models.Layout) (models.Layout, error) {
	panic("unused")
}
func (*stubLayouts) UpdateLayout(context.Context, models.Layout) (models.Layout, error) {
	panic("unused")
}
func (*stubLayouts) PublishLayout(context.Context, string, string) (models.Layout, error) {
	panic("unused")
}
func (*stubLayouts) ProposeLayout(context.Context, string, models.LayoutProposal) (models.LayoutDiff, error) {
	panic("unused")
}
func (*stubLayouts) PatchLayout(context.Context, string, models.LayoutProposal) (models.LayoutApplied, error) {
	panic("unused")
}
func (*stubLayouts) RevertLayout(context.Context, string, models.LayoutRevert) (models.LayoutApplied, error) {
	panic("unused")
}
func (*stubLayouts) CopyLayout(context.Context, string, models.LayoutCopy) (models.Layout, error) {
	panic("unused")
}
func (*stubLayouts) ListLayoutChanges(context.Context, string, models.LayoutChangeFilter) ([]models.LayoutChange, error) {
	panic("unused")
}
func (*stubLayouts) ChooseLayout(context.Context, models.LayoutChoice) (models.LayoutChoice, error) {
	panic("unused")
}
func (*stubLayouts) ChosenLayout(context.Context, string, string) (models.LayoutChoice, error) {
	panic("unused")
}
