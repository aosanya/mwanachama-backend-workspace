package mwanachamaworkspace_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	workspace "github.com/aosanya/mwanachama-backend-workspace"
)

func newTestManager(t *testing.T) workspace.WorkspaceManager {
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
	mgr, err := workspace.NewWorkspaceManager(db, s)
	if err != nil {
		t.Fatalf("NewWorkspaceManager: %v", err)
	}
	return mgr
}
