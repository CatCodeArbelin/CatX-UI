package forkrecovery

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/util/json_util"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

// TestValidateXrayCandidate_E2E exercises the exact installed-binary boundary.
// It is opt-in for local/CI smoke jobs that provide the pinned core binary.
func TestValidateXrayCandidate_E2E(t *testing.T) {
	source := os.Getenv("XRAY_E2E_BINARY")
	if source == "" {
		t.Skip("set XRAY_E2E_BINARY to run installed-Xray candidate validation")
	}
	dir := t.TempDir()
	t.Setenv("XUI_BIN_FOLDER", dir)
	data, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read Xray binary: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, xray.GetBinaryName()), data, 0o755); err != nil {
		t.Fatalf("stage Xray binary: %v", err)
	}

	valid := &xray.Config{
		LogConfig:       json_util.RawMessage(`{"loglevel":"warning"}`),
		RouterConfig:    json_util.RawMessage(`{"rules":[]}`),
		InboundConfigs:  []xray.InboundConfig{},
		OutboundConfigs: json_util.RawMessage(`[{"tag":"direct","protocol":"freedom"}]`),
		Policy:          json_util.RawMessage(`{}`),
		Stats:           json_util.RawMessage(`{}`),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := ValidateXrayCandidate(ctx, valid); err != nil {
		t.Fatalf("installed Xray rejected valid candidate: %v", err)
	}

	invalid, err := CloneXrayConfig(valid)
	if err != nil {
		t.Fatalf("clone valid candidate: %v", err)
	}
	invalid.OutboundConfigs = json_util.RawMessage(`[{"tag":"bad","protocol":"not-a-protocol"}]`)
	if err := ValidateXrayCandidate(ctx, invalid); err == nil {
		t.Fatal("installed Xray accepted an invalid outbound protocol")
	}
}
