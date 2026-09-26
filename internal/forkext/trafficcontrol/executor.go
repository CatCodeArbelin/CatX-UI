package trafficcontrol

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type CommandResult struct {
	Stdout []byte
	Stderr []byte
}

type Executor interface {
	Run(context.Context, string, []string, []byte) (CommandResult, error)
}

type processExecutor struct{}

func (processExecutor) Run(ctx context.Context, name string, args []string, stdin []byte) (CommandResult, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = bytes.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if err != nil {
		return CommandResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, fmt.Errorf("%s failed: %w: %s", name, err, stderr.String())
	}
	return CommandResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}, nil
}
