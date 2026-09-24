// Package forkrecovery contains the downstream safety boundary for risky
// runtime and data-store activation.  It deliberately depends only on the
// public Xray process/config surface so upstream lifecycle code stays in
// control of how Xray is built and run.
package forkrecovery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

// Failure reports both the candidate failure and the outcome of automatic
// rollback. RollbackErr being nil means the previous known-good state was
// restored and passed its healthcheck.
type Failure struct {
	Operation   string
	Candidate   error
	RollbackErr error
}

func (e *Failure) Error() string {
	operation := strings.TrimSpace(e.Operation)
	if operation == "" {
		operation = "candidate activation"
	}
	if e.RollbackErr != nil {
		return fmt.Sprintf("%s failed: %v; automatic rollback failed: %v", operation, e.Candidate, e.RollbackErr)
	}
	return fmt.Sprintf("%s failed: %v; automatic rollback succeeded", operation, e.Candidate)
}

func (e *Failure) Unwrap() error { return e.Candidate }

// RollbackFailed makes rollback failure machine-distinguishable without
// callers having to parse operator-facing error text.
func (e *Failure) RollbackFailed() bool { return e.RollbackErr != nil }

// NewFailure records a candidate failure and the result of restoring and
// healthchecking the previous known-good state.
func NewFailure(operation string, candidateErr, rollbackErr error) error {
	if candidateErr == nil {
		candidateErr = errors.New("unknown candidate failure")
	}
	return &Failure{Operation: operation, Candidate: candidateErr, RollbackErr: rollbackErr}
}

// XraySnapshot is the minimum state needed to reconstruct the previously
// working runtime and restore its exact on-disk configuration.
type XraySnapshot struct {
	Config       *xray.Config
	ConfigFile   []byte
	ConfigExists bool
}

// CaptureXraySnapshot copies the running process configuration when present
// and also preserves config.json exactly. The latter covers panel startup,
// where no in-memory process exists yet but the last working file does.
func CaptureXraySnapshot(process *xray.Process) (*XraySnapshot, error) {
	snapshot := &XraySnapshot{}
	if process != nil && process.GetConfig() != nil {
		cfg, err := CloneXrayConfig(process.GetConfig())
		if err != nil {
			return nil, fmt.Errorf("copy running Xray config: %w", err)
		}
		snapshot.Config = cfg
	}

	data, err := os.ReadFile(xray.GetConfigPath())
	if err == nil {
		snapshot.ConfigFile = data
		snapshot.ConfigExists = true
		if snapshot.Config == nil {
			var cfg xray.Config
			if unmarshalErr := json.Unmarshal(data, &cfg); unmarshalErr == nil {
				snapshot.Config = &cfg
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) && snapshot.Config == nil {
		return nil, fmt.Errorf("read known-good Xray config: %w", err)
	}
	return snapshot, nil
}

// CloneXrayConfig returns a deep copy without introducing a second config
// builder. JSON is also the exact representation handed to Xray.
func CloneXrayConfig(cfg *xray.Config) (*xray.Config, error) {
	if cfg == nil {
		return nil, nil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	var clone xray.Config
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil, err
	}
	return &clone, nil
}

// RestoreConfigFile restores the exact pre-activation config.json. When no
// previous file existed it leaves the file produced by a successfully restored
// runtime in place.
func (s *XraySnapshot) RestoreConfigFile() error {
	if s == nil || !s.ConfigExists {
		return nil
	}
	return writeFileAtomic(xray.GetConfigPath(), s.ConfigFile, 0o600)
}

// ValidateXrayCandidate asks the installed Xray binary to parse and build the
// candidate from a private temporary file. It never replaces config.json.
func ValidateXrayCandidate(ctx context.Context, cfg *xray.Config) error {
	if cfg == nil {
		return errors.New("Xray candidate is nil")
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal Xray candidate: %w", err)
	}
	dir := filepath.Dir(xray.GetConfigPath())
	file, err := os.CreateTemp(dir, ".xray-candidate-*.json")
	if err != nil {
		return fmt.Errorf("create Xray candidate file: %w", err)
	}
	path := file.Name()
	defer os.Remove(path)
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return fmt.Errorf("protect Xray candidate file: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return fmt.Errorf("write Xray candidate file: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync Xray candidate file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close Xray candidate file: %w", err)
	}

	cmd := exec.CommandContext(ctx, xray.GetBinaryPath(), "run", "-test", "-c", path)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(output.String())
		if message == "" {
			message = err.Error()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("installed Xray validation timed out: %w", ctx.Err())
		}
		return fmt.Errorf("installed Xray rejected candidate: %s", message)
	}
	return nil
}

// HealthcheckXray waits until the process remains alive and its local API port
// accepts a connection. The caller owns the deadline.
func HealthcheckXray(ctx context.Context, process *xray.Process) error {
	if process == nil {
		return errors.New("Xray process is nil")
	}
	port := process.GetAPIPort()
	if port <= 0 {
		return errors.New("Xray API port is unavailable")
	}
	address := fmt.Sprintf("127.0.0.1:%d", port)
	dialer := net.Dialer{Timeout: 250 * time.Millisecond}
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	var lastErr error
	for {
		if !process.IsRunning() {
			result := strings.TrimSpace(process.GetResult())
			if result == "" {
				result = "process exited"
			}
			return fmt.Errorf("Xray stopped during healthcheck: %s", result)
		}
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err == nil {
			_ = conn.Close()
			if !process.IsRunning() {
				return errors.New("Xray stopped immediately after its API became reachable")
			}
			return nil
		}
		lastErr = err
		select {
		case <-ctx.Done():
			return fmt.Errorf("Xray API healthcheck failed: %w", errors.Join(lastErr, ctx.Err()))
		case <-ticker.C:
		}
	}
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) (err error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".xray-recovery-*.tmp")
	if err != nil {
		return err
	}
	tempPath := file.Name()
	defer func() {
		_ = file.Close()
		if err != nil {
			_ = os.Remove(tempPath)
		}
	}()
	if mode == 0 {
		mode = 0o600
	}
	if err = file.Chmod(mode); err != nil {
		return err
	}
	if _, err = file.Write(data); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err = replaceFile(tempPath, path); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return nil
	}
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	err = directory.Sync()
	_ = directory.Close()
	return err
}
