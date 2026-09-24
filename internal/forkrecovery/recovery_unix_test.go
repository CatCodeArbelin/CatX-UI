//go:build !windows

package forkrecovery

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

func TestValidateXrayCandidateUsesInstalledBinaryWithoutReplacingKnownGood(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XUI_BIN_FOLDER", dir)
	knownGood := []byte(`{"known":"good"}`)
	if err := os.WriteFile(xray.GetConfigPath(), knownGood, 0o600); err != nil {
		t.Fatalf("write known-good config: %v", err)
	}
	argsPath := filepath.Join(dir, "validator.args")
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" > \"$XRAY_TEST_ARGS\"\n" +
		"echo 'synthetic core rejection' >&2\n" +
		"exit 23\n"
	if err := os.WriteFile(xray.GetBinaryPath(), []byte(script), 0o755); err != nil {
		t.Fatalf("write validator stub: %v", err)
	}
	t.Setenv("XRAY_TEST_ARGS", argsPath)

	err := ValidateXrayCandidate(context.Background(), &xray.Config{})
	if err == nil || !strings.Contains(err.Error(), "synthetic core rejection") {
		t.Fatalf("ValidateXrayCandidate error = %v, want installed-core rejection", err)
	}
	args, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read validator args: %v", err)
	}
	if !strings.Contains(string(args), "run -test -c ") || !strings.Contains(string(args), ".xray-candidate-") {
		t.Fatalf("validator args = %q, want run -test -c <private candidate>", args)
	}
	got, err := os.ReadFile(xray.GetConfigPath())
	if err != nil {
		t.Fatalf("read known-good config: %v", err)
	}
	if string(got) != string(knownGood) {
		t.Fatalf("known-good config changed to %q", got)
	}
	matches, err := filepath.Glob(filepath.Join(dir, ".xray-candidate-*.json"))
	if err != nil || len(matches) != 0 {
		t.Fatalf("candidate temp files after validation = %v, err=%v", matches, err)
	}
}
