//go:build linux

package trafficcontrol

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

type namespaceExecutor struct{ namespace string }

func (e namespaceExecutor) Run(ctx context.Context, name string, args []string, input []byte) (CommandResult, error) {
	all := append([]string{"netns", "exec", e.namespace, name}, args...)
	return (processExecutor{}).Run(ctx, "ip", all, input)
}

func TestPrivilegedNetworkNamespaceCapabilityProbe(t *testing.T) {
	if os.Getenv("CATX_TC_NETNS_TEST") != "1" {
		t.Skip("set CATX_TC_NETNS_TEST=1 to run privileged network namespace coverage")
	}
	if os.Geteuid() != 0 {
		t.Skip("requires root/CAP_NET_ADMIN")
	}
	for _, name := range []string{"ip", "tc", "nft"} {
		if _, err := exec.LookPath(name); err != nil {
			t.Skipf("%s unavailable", name)
		}
	}
	ns := fmt.Sprintf("catx-tc-%d", os.Getpid())
	if _, err := (processExecutor{}).Run(context.Background(), "ip", []string{"netns", "add", ns}, nil); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = (processExecutor{}).Run(context.Background(), "ip", []string{"netns", "delete", ns}, nil)
	}()
	b := NewBackend(namespaceExecutor{namespace: ns}, "linux", []string{"lo"})
	cap := b.Capabilities(context.Background())
	if cap.State != State(stateReady) {
		t.Fatalf("namespace capability probe: %+v", cap)
	}
}
