package trafficcontrol

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
)

type fakeCommand struct {
	name  string
	args  []string
	input string
}
type fakeExecutor struct {
	mu           sync.Mutex
	calls        []fakeCommand
	failNftApply bool
	qdiscChanged bool
}

func (f *fakeExecutor) Run(_ context.Context, name string, args []string, input []byte) (CommandResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, fakeCommand{name: name, args: append([]string(nil), args...), input: string(input)})
	if name == "tc" && len(args) >= 2 && args[0] == "qdisc" && args[1] == "replace" {
		f.qdiscChanged = true
	}
	if f.failNftApply && name == "nft" && len(args) > 0 && args[0] == "-f" {
		return CommandResult{}, errors.New("injected failure")
	}
	if name == "tc" && len(args) >= 3 && args[0] == "qdisc" && args[1] == "show" {
		if f.qdiscChanged {
			return CommandResult{Stdout: []byte("qdisc htb 1: root")}, nil
		}
		return CommandResult{Stdout: []byte("qdisc noqueue 0: root")}, nil
	}
	if name == "nft" && len(args) >= 2 && args[0] == "list" {
		return CommandResult{}, errors.New("table does not exist")
	}
	return CommandResult{}, nil
}

func TestAllocateMarkStableAndCollisionSafe(t *testing.T) {
	mark, err := AllocateMark("node-a", "client-a", nil)
	if err != nil {
		t.Fatal(err)
	}
	if mark == 0 || mark > 65534 {
		t.Fatalf("mark=%d outside tc minor range", mark)
	}
	again, err := AllocateMark("node-a", "client-a", map[uint32]string{mark: "node-a\x00client-a"})
	if err != nil || again != mark {
		t.Fatalf("stable mark: %d/%v", again, err)
	}
	if _, err := AllocateMark("node-a", "client-a", map[uint32]string{mark: "other"}); err == nil {
		t.Fatal("occupied mark must be rejected")
	}
}

func TestUnsupportedPlatformIsExactNoOp(t *testing.T) {
	f := &fakeExecutor{}
	b := NewBackend(f, "windows", []string{"eth0"})
	cap := b.Capabilities(context.Background())
	if cap.State != State(stateUnsupported) {
		t.Fatalf("state=%s", cap.State)
	}
	if len(f.calls) != 0 {
		t.Fatalf("unsupported platform ran commands: %+v", f.calls)
	}
	status, err := b.Reconcile(context.Background(), nil)
	if err != nil || status.State != State(stateUnsupported) {
		t.Fatalf("reconcile=%+v err=%v", status, err)
	}
}

type qdiscExecutor struct {
	calls  []fakeCommand
	stdout string
}

func (e *qdiscExecutor) Run(_ context.Context, name string, args []string, input []byte) (CommandResult, error) {
	e.calls = append(e.calls, fakeCommand{name: name, args: args, input: string(input)})
	if name == "tc" && len(args) > 1 && args[0] == "qdisc" && args[1] == "show" {
		return CommandResult{Stdout: []byte(e.stdout)}, nil
	}
	if name == "nft" && len(args) > 1 && args[0] == "list" {
		return CommandResult{}, errors.New("missing")
	}
	return CommandResult{}, nil
}

func TestBackendRejectsAdminOwnedQdisc(t *testing.T) {
	e := &qdiscExecutor{stdout: "qdisc cake 8000: root"}
	b := NewBackend(e, "linux", []string{"eth0"})
	status, err := b.Reconcile(context.Background(), []DesiredRule{{NodeKey: "n", ClientKey: "c", Interface: "eth0", Mark: 10, UploadRateBps: 8000, DownloadRateBps: 8000, Selectors: []string{"192.0.2.1/32"}}})
	if err == nil || status.State != State(stateDegraded) {
		t.Fatalf("expected degraded refusal: %+v err=%v", status, err)
	}
	for _, call := range e.calls {
		if call.name == "tc" && len(call.args) > 1 && call.args[0] == "qdisc" && call.args[1] == "replace" {
			t.Fatal("admin qdisc was replaced")
		}
	}
}

func TestBackendApplyUsesStructuredCommandsAndRollback(t *testing.T) {
	f := &fakeExecutor{failNftApply: true}
	b := NewBackend(f, "linux", []string{"eth0"})
	rule := DesiredRule{NodeKey: "n", ClientKey: "c", Interface: "eth0", Mark: 10, UploadRateBps: 8000, DownloadRateBps: 16000, Selectors: []string{"192.0.2.1/32", "2001:db8::1/128"}}
	status, err := b.Reconcile(context.Background(), []DesiredRule{rule})
	if err == nil || status.State != State(stateDegraded) {
		t.Fatalf("expected rollback failure: %+v err=%v", status, err)
	}
	seenRollback := false
	for _, call := range f.calls {
		if call.name == "tc" && len(call.args) >= 3 && call.args[0] == "qdisc" && call.args[1] == "del" {
			seenRollback = true
		}
		for _, arg := range call.args {
			if strings.ContainsAny(arg, "|&$`") {
				t.Fatalf("shell metacharacter in structured argument %q", arg)
			}
		}
	}
	if !seenRollback {
		t.Fatal("apply failure did not attempt rollback")
	}
}

type lifecycleExecutor struct {
	qdisc bool
	nft   bool
}

func (e *lifecycleExecutor) Run(_ context.Context, name string, args []string, _ []byte) (CommandResult, error) {
	if name == "tc" && len(args) >= 2 && args[0] == "qdisc" && args[1] == "show" {
		if e.qdisc {
			return CommandResult{Stdout: []byte("qdisc htb 1: root")}, nil
		}
		return CommandResult{Stdout: []byte("qdisc noqueue 0: root")}, nil
	}
	if name == "tc" && len(args) >= 2 && args[0] == "qdisc" && args[1] == "replace" {
		e.qdisc = true
	}
	if name == "tc" && len(args) >= 2 && args[0] == "qdisc" && args[1] == "del" {
		e.qdisc = false
	}
	if name == "ip" && len(args) >= 3 && args[0] == "link" && args[1] == "show" {
		return CommandResult{}, nil
	}
	if name == "nft" && len(args) >= 2 && args[0] == "list" {
		if e.nft {
			return CommandResult{Stdout: []byte("table inet catx_traffic_control { comment \"catx-managed-v1\"; }")}, nil
		}
		return CommandResult{}, errors.New("table does not exist")
	}
	if name == "nft" && len(args) >= 1 && args[0] == "-f" {
		e.nft = true
	}
	if name == "nft" && len(args) >= 2 && args[0] == "delete" {
		e.nft = false
	}
	return CommandResult{}, nil
}

func TestConfigureDisabledCleansOwnedState(t *testing.T) {
	e := &lifecycleExecutor{}
	b := NewBackend(e, "linux", []string{"eth0"})
	rule := DesiredRule{NodeKey: "n", ClientKey: "c", Interface: "eth0", Mark: 10, UploadRateBps: 8000, DownloadRateBps: 8000, Selectors: []string{"192.0.2.1/32"}}
	if _, err := b.Reconcile(context.Background(), []DesiredRule{rule}); err != nil {
		t.Fatalf("apply owned state: %v", err)
	}
	SetBackendForTests(b)
	t.Cleanup(func() { Configure(false) })
	Configure(false)
	if e.qdisc || e.nft {
		t.Fatalf("disable left owned state behind: qdisc=%v nft=%v", e.qdisc, e.nft)
	}
}

func TestIFBNameStaysWithinLinuxInterfaceLimit(t *testing.T) {
	name := ifbName("catxrc6dummy0")
	if len(name) > 15 {
		t.Fatalf("IFB name %q is %d characters; Linux limit is 15", name, len(name))
	}
	if name != ifbName("catxrc6dummy0") {
		t.Fatalf("IFB name is not deterministic: %q", name)
	}
}

type fakeRemote struct {
	capability []byte
	response   []byte
	calls      int
}

func (r *fakeRemote) TrafficControlCapabilities(context.Context) (json.RawMessage, error) {
	r.calls++
	return r.capability, nil
}

func (r *fakeRemote) TrafficControlReconcile(context.Context, json.RawMessage) (json.RawMessage, error) {
	r.calls++
	return r.response, nil
}

func TestRemoteReconcileUsesCapabilityBoundary(t *testing.T) {
	r := &fakeRemote{capability: []byte(`{"state":"ready","platform":"linux"}`), response: []byte(`{"status":{"state":"ready","generation":2}}`)}
	status, err := ReconcileRemote(context.Background(), r, []DesiredRule{{NodeKey: "n", ClientKey: "c"}})
	if err != nil || status.Generation != 2 || r.calls != 2 {
		t.Fatalf("status=%+v err=%v calls=%d", status, err, r.calls)
	}
}
