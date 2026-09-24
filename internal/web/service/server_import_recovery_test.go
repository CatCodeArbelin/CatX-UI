package service

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/forkrecovery"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

func setRecoveryMarker(t *testing.T, value string) {
	t.Helper()
	if err := database.GetDB().Where("key = ?", "recoveryTestMarker").Assign(model.Setting{Value: value}).FirstOrCreate(&model.Setting{Key: "recoveryTestMarker"}).Error; err != nil {
		t.Fatalf("set recovery marker: %v", err)
	}
}

func recoveryMarker() string {
	var setting model.Setting
	if db := database.GetDB(); db != nil {
		_ = db.Where("key = ?", "recoveryTestMarker").First(&setting).Error
	}
	return setting.Value
}

func TestSQLiteImportInvalidXrayCandidateRestoresPreviousDatabase(t *testing.T) {
	preserveXrayTestGlobals(t)
	dir := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", dir)
	t.Setenv("XUI_BIN_FOLDER", t.TempDir())
	t.Setenv("XUI_LOG_FOLDER", t.TempDir())
	uploadPath := filepath.Join(dir, "upload.db")
	if err := database.InitDB(uploadPath); err != nil {
		t.Fatalf("InitDB(upload): %v", err)
	}
	setRecoveryMarker(t, "invalid-candidate")
	if err := database.CloseDB(); err != nil {
		t.Fatalf("CloseDB(upload): %v", err)
	}

	livePath := filepath.Join(dir, "x-ui.db")
	if err := database.InitDB(livePath); err != nil {
		t.Fatalf("InitDB(live): %v", err)
	}
	t.Cleanup(func() { _ = database.CloseDB() })
	setRecoveryMarker(t, "known-good")

	upload, err := os.Open(uploadPath)
	if err != nil {
		t.Fatalf("open upload: %v", err)
	}
	defer upload.Close()

	restoreHooks := SetXrayRecoveryHooksForTest(
		func(context.Context, *xray.Config) error {
			if recoveryMarker() == "invalid-candidate" {
				return errors.New("installed Xray rejected imported config")
			}
			return nil
		},
		func(context.Context, *xray.Process) error { return nil },
	)
	restoreStart := SetXrayStartHookForTest(func(*xray.Process) error { return nil })
	t.Cleanup(restoreHooks)
	t.Cleanup(restoreStart)

	err = (&ServerService{}).ImportDB(upload, false)
	var failure *forkrecovery.Failure
	if !errors.As(err, &failure) || failure.RollbackFailed() {
		t.Fatalf("ImportDB error = %v, want rejected candidate with successful rollback", err)
	}
	if got := recoveryMarker(); got != "known-good" {
		t.Fatalf("database marker = %q, want restored known-good", got)
	}
	if err := database.ValidateSQLiteDB(livePath); err != nil {
		t.Fatalf("restored database is invalid: %v", err)
	}
}

func TestSQLiteImportReportsRollbackFailure(t *testing.T) {
	preserveXrayTestGlobals(t)
	dir := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", dir)
	t.Setenv("XUI_BIN_FOLDER", t.TempDir())
	t.Setenv("XUI_LOG_FOLDER", t.TempDir())
	uploadPath := filepath.Join(dir, "upload.db")
	if err := database.InitDB(uploadPath); err != nil {
		t.Fatalf("InitDB(upload): %v", err)
	}
	setRecoveryMarker(t, "invalid-candidate")
	if err := database.CloseDB(); err != nil {
		t.Fatalf("CloseDB(upload): %v", err)
	}

	livePath := filepath.Join(dir, "x-ui.db")
	if err := database.InitDB(livePath); err != nil {
		t.Fatalf("InitDB(live): %v", err)
	}
	t.Cleanup(func() { _ = database.CloseDB() })
	setRecoveryMarker(t, "known-good")
	upload, err := os.Open(uploadPath)
	if err != nil {
		t.Fatalf("open upload: %v", err)
	}
	defer upload.Close()

	restoreHooks := SetXrayRecoveryHooksForTest(
		func(context.Context, *xray.Config) error {
			if recoveryMarker() != "invalid-candidate" {
				return nil
			}
			matches, globErr := filepath.Glob(livePath + ".recovery-*")
			if globErr != nil || len(matches) != 1 {
				t.Fatalf("recovery snapshots = %v, err=%v; want one", matches, globErr)
			}
			if removeErr := os.Remove(matches[0]); removeErr != nil {
				t.Fatalf("inject rollback failure: %v", removeErr)
			}
			return errors.New("candidate rejected before injected rollback failure")
		},
		func(context.Context, *xray.Process) error { return nil },
	)
	restoreStart := SetXrayStartHookForTest(func(*xray.Process) error { return nil })
	t.Cleanup(restoreHooks)
	t.Cleanup(restoreStart)

	err = (&ServerService{}).ImportDB(upload, false)
	var failure *forkrecovery.Failure
	if !errors.As(err, &failure) || !failure.RollbackFailed() {
		t.Fatalf("ImportDB error = %v, want distinguishable rollback failure", err)
	}
	if !strings.Contains(err.Error(), "restore known-good SQLite database") {
		t.Fatalf("ImportDB error = %q, want rollback cause", err)
	}
	// The rollback helper puts the candidate back when the known-good snapshot
	// is unavailable, so the DB path remains recoverable for inspection/retry.
	if err := database.InitDB(livePath); err != nil {
		t.Fatalf("reopen retained candidate after injected rollback failure: %v", err)
	}
}

func TestPostgresImportInvalidXrayCandidateRestoresPreviousDatabase(t *testing.T) {
	if os.Getenv("XUI_DB_TYPE") != "postgres" || strings.TrimSpace(os.Getenv("XUI_DB_DSN")) == "" {
		t.Skip("set XUI_DB_TYPE=postgres and XUI_DB_DSN to run PostgreSQL import recovery")
	}
	if _, err := exec.LookPath("pg_dump"); err != nil {
		t.Skip("pg_dump/pg_restore client tools are required")
	}
	preserveXrayTestGlobals(t)
	t.Setenv("XUI_BIN_FOLDER", t.TempDir())
	t.Setenv("XUI_LOG_FOLDER", t.TempDir())
	if err := database.InitDB(""); err != nil {
		t.Fatalf("InitDB(postgres): %v", err)
	}
	t.Cleanup(func() { _ = database.CloseDB() })

	setRecoveryMarker(t, "invalid-candidate")
	candidate, err := (&ServerService{}).GetDb()
	if err != nil {
		t.Fatalf("create PostgreSQL candidate dump: %v", err)
	}
	setRecoveryMarker(t, "known-good")
	uploadPath := filepath.Join(t.TempDir(), "candidate.dump")
	if err := os.WriteFile(uploadPath, candidate, 0o600); err != nil {
		t.Fatalf("write candidate dump: %v", err)
	}
	upload, err := os.Open(uploadPath)
	if err != nil {
		t.Fatalf("open candidate dump: %v", err)
	}
	defer upload.Close()

	restoreHooks := SetXrayRecoveryHooksForTest(
		func(context.Context, *xray.Config) error {
			if recoveryMarker() == "invalid-candidate" {
				return errors.New("installed Xray rejected imported PostgreSQL config")
			}
			return nil
		},
		func(context.Context, *xray.Process) error { return nil },
	)
	restoreStart := SetXrayStartHookForTest(func(*xray.Process) error { return nil })
	t.Cleanup(restoreHooks)
	t.Cleanup(restoreStart)

	err = (&ServerService{}).ImportDB(upload, false)
	var failure *forkrecovery.Failure
	if !errors.As(err, &failure) || failure.RollbackFailed() {
		t.Fatalf("ImportDB error = %v, want rejected candidate with successful PostgreSQL rollback", err)
	}
	if got := recoveryMarker(); got != "known-good" {
		t.Fatalf("PostgreSQL marker = %q, want restored known-good", got)
	}
}

func TestPostgresRollbackFailureReportsRetainedSnapshot(t *testing.T) {
	rollbackErr := errors.New("pg_restore rollback failed")
	err := retainPostgresRecoveryError("/secure/known-good.dump", rollbackErr)
	if !errors.Is(err, rollbackErr) || !strings.Contains(err.Error(), "known-good PostgreSQL snapshot retained at /secure/known-good.dump") {
		t.Fatalf("retained rollback error = %v", err)
	}
	if err := retainPostgresRecoveryError("unused", nil); err != nil {
		t.Fatalf("successful rollback returned error: %v", err)
	}
}
