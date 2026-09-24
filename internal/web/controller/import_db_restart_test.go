package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/web/global"
	webservice "github.com/mhsanaei/3x-ui/v3/internal/web/service"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"

	"github.com/gin-gonic/gin"
)

// A successful import must schedule the panel restart itself: the browser's
// restartPanel follow-up can 401 once the imported users table lands (#6446).
func TestImportDBSchedulesPanelRestart(t *testing.T) {
	uploadPath := filepath.Join(t.TempDir(), "x-ui.db")
	if err := database.InitDB(uploadPath); err != nil {
		t.Fatalf("InitDB(upload): %v", err)
	}
	if err := database.CloseDB(); err != nil {
		t.Fatalf("CloseDB(upload): %v", err)
	}
	upload, err := os.ReadFile(uploadPath)
	if err != nil {
		t.Fatalf("read upload: %v", err)
	}

	newHostTestDB(t)
	binDir := t.TempDir()
	t.Setenv("XUI_BIN_FOLDER", binDir)
	t.Setenv("XUI_LOG_FOLDER", t.TempDir())
	restoreRecovery := webservice.SetXrayRecoveryHooksForTest(
		func(context.Context, *xray.Config) error { return nil },
		func(context.Context, *xray.Process) error { return nil },
	)
	restoreStart := webservice.SetXrayStartHookForTest(func(*xray.Process) error { return nil })
	restoreProcess := webservice.SetXrayProcessForTest(nil)
	t.Cleanup(restoreRecovery)
	t.Cleanup(restoreStart)
	t.Cleanup(restoreProcess)

	restarts := make(chan struct{}, 1)
	global.SetRestartHook(func() {
		select {
		case restarts <- struct{}{}:
		default:
		}
	})
	t.Cleanup(func() { global.SetRestartHook(func() {}) })

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("db", "x-ui.db")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(upload); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	a := &ServerController{}
	engine := gin.New()
	engine.POST("/panel/api/server/importDB", a.importDB)
	req := httptest.NewRequest(http.MethodPost, "/panel/api/server/importDB", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	var env hostEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, w.Body.String())
	}
	if !env.Success {
		t.Fatalf("importDB failed: %s", env.Msg)
	}

	select {
	case <-restarts:
	case <-time.After(6 * time.Second):
		t.Fatal("importDB succeeded but no panel restart was scheduled within 6s")
	}
}
