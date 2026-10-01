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
	failCommand  string
	qdiscChanged bool
}

func (f *fakeExecutor) Run(_ context.Context, name string, args []string, input []byte) (CommandResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, fakeCommand{name: name, args: append([]string(nil), args...), input: string(input)})
	if f.failCommand == name+" "+strings.Join(args, " ") {
		return CommandResult{}, errors.New("injected command failure")
	}
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
	if name == "ip" && len(args) >= 4 && args[0] == "link" && args[1] == "show" && args[3] != "eth0" {
		return CommandResult{}, errors.New("does not exist")
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
	if name == "ip" && len(args) >= 4 && args[0] == "link" && args[1] == "show" && args[3] != "eth0" {
		return CommandResult{}, errors.New("does not exist")
	}
	return CommandResult{}, nil
}

func assertNoDestructiveCommands(t *testing.T, calls []fakeCommand) {
	t.Helper()
	for _, call := range calls {
		if call.name == "ip" && len(call.args) >= 2 && call.args[0] == "link" && call.args[1] == "del" {
			t.Fatalf("destructive IP command ran: %+v", call)
		}
		if call.name == "nft" && len(call.args) >= 1 && call.args[0] == "delete" {
			t.Fatalf("destructive nft command ran: %+v", call)
		}
		if call.name == "tc" && len(call.args) >= 2 && call.args[0] == "qdisc" && (call.args[1] == "replace" || call.args[1] == "del") {
			t.Fatalf("destructive tc qdisc command ran: %+v", call)
		}
	}
}

func TestBackendRejectsAdminOwnedQdisc(t *testing.T) {
	e := &qdiscExecutor{stdout: "qdisc cake 8000: root"}
	b := NewBackend(e, "linux", []string{"eth0"})
	status, err := b.Reconcile(context.Background(), []DesiredRule{{NodeKey: "n", ClientKey: "c", Interface: "eth0", Mark: 10, UploadRateBps: 8000, DownloadRateBps: 8000, Selectors: []string{"192.0.2.1/32"}}})
	if err == nil || status.State != State(stateDegraded) {
		t.Fatalf("expected degraded refusal: %+v err=%v", status, err)
	}
	assertNoDestructiveCommands(t, e.calls)
}

func TestBackendRejectsAdminOwnedIngressQdisc(t *testing.T) {
	e := &qdiscExecutor{stdout: "qdisc noqueue 0: root\nqdisc ingress ffff: parent ffff:fff1"}
	b := NewBackend(e, "linux", []string{"eth0"})
	status, err := b.Reconcile(context.Background(), []DesiredRule{{NodeKey: "n", ClientKey: "c", Interface: "eth0", Mark: 10, UploadRateBps: 8000, DownloadRateBps: 8000, Selectors: []string{"192.0.2.1/32"}}})
	if err == nil || status.State != State(stateDegraded) {
		t.Fatalf("expected ingress refusal: %+v err=%v", status, err)
	}
	assertNoDestructiveCommands(t, e.calls)
}

func TestBackendRejectsAdminOwnedClsactWithoutMutation(t *testing.T) {
	e := &qdiscExecutor{stdout: "qdisc noqueue 0: root\nqdisc clsact ffff: parent ffff:fff1"}
	b := NewBackend(e, "linux", []string{"eth0"})
	status, err := b.Reconcile(context.Background(), []DesiredRule{{NodeKey: "n", ClientKey: "c", Interface: "eth0", Mark: 10, UploadRateBps: 8000, DownloadRateBps: 8000, Selectors: []string{"192.0.2.1/32"}}})
	if err == nil || status.State != State(stateDegraded) {
		t.Fatalf("expected clsact refusal: %+v err=%v", status, err)
	}
	assertNoDestructiveCommands(t, e.calls)
}

func TestBackendCreatesIngressQdiscForRedirect(t *testing.T) {
	e := &qdiscExecutor{stdout: "qdisc noqueue 0: root"}
	b := NewBackend(e, "linux", []string{"eth0"})
	_, err := b.Reconcile(context.Background(), []DesiredRule{{NodeKey: "n", ClientKey: "c", Interface: "eth0", Mark: 10, UploadRateBps: 8000, DownloadRateBps: 8000, Selectors: []string{"192.0.2.1/32"}}})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	ingressIndex, redirectIndex := -1, -1
	for index, call := range e.calls {
		if call.name == "tc" && len(call.args) >= 7 && call.args[0] == "qdisc" && call.args[1] == "replace" && call.args[2] == "dev" && call.args[3] == "eth0" && call.args[4] == "handle" && call.args[5] == "ffff:" && call.args[6] == "ingress" {
			ingressIndex = index
		}
		if call.name == "tc" && len(call.args) >= 4 && call.args[0] == "filter" && call.args[1] == "replace" && call.args[2] == "dev" && call.args[3] == "eth0" && strings.Contains(strings.Join(call.args, " "), " redirect dev ") {
			redirectIndex = index
		}
	}
	if ingressIndex < 0 {
		t.Fatal("reconcile did not create the ingress qdisc required by the redirect filter")
	}
	if redirectIndex < 0 || ingressIndex > redirectIndex {
		t.Fatalf("ingress qdisc must exist before redirect filter: ingress=%d redirect=%d", ingressIndex, redirectIndex)
	}
	redirect := e.calls[redirectIndex]
	if !strings.Contains(strings.Join(redirect.args, " "), " handle 1 ") {
		t.Fatalf("redirect filter must use a stable handle for idempotent replacement: %v", redirect.args)
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

func TestBackendPartialApplyRollsBackCreatedRootQdisc(t *testing.T) {
	ifb := ifbName("eth0")
	f := &fakeExecutor{failCommand: "tc class replace dev " + ifb + " parent 1: classid 1:1 htb rate 1gbit"}
	b := NewBackend(f, "linux", []string{"eth0"})
	_, err := b.Reconcile(context.Background(), []DesiredRule{{NodeKey: "n", ClientKey: "c", Interface: "eth0", Mark: 10, UploadRateBps: 8000, DownloadRateBps: 8000, Selectors: []string{"192.0.2.1/32"}}})
	if err == nil {
		t.Fatal("partial apply should fail")
	}
	for _, call := range f.calls {
		if call.name == "tc" && len(call.args) >= 5 && call.args[0] == "qdisc" && call.args[1] == "del" && call.args[3] == "eth0" && call.args[4] == "root" {
			return
		}
	}
	t.Fatal("partial apply did not roll back the CatX-created root qdisc")
}

func TestBackendPartialApplyRollsBackCreatedIngressQdisc(t *testing.T) {
	f := &fakeExecutor{failCommand: "tc filter replace dev eth0 parent 1: protocol all pref 100 handle 10 fw flowid 1:10"}
	b := NewBackend(f, "linux", []string{"eth0"})
	_, err := b.Reconcile(context.Background(), []DesiredRule{{NodeKey: "n", ClientKey: "c", Interface: "eth0", Mark: 10, UploadRateBps: 8000, DownloadRateBps: 8000, Selectors: []string{"192.0.2.1/32"}}})
	if err == nil {
		t.Fatal("partial apply should fail")
	}
	var ingressDeleted, rootDeleted bool
	for _, call := range f.calls {
		if call.name != "tc" || len(call.args) < 5 || call.args[0] != "qdisc" || call.args[1] != "del" {
			continue
		}
		if call.args[3] == "eth0" && call.args[4] == "ingress" {
			ingressDeleted = true
		}
		if call.args[3] == "eth0" && call.args[4] == "root" {
			rootDeleted = true
		}
	}
	if !ingressDeleted || !rootDeleted {
		t.Fatalf("partial apply rollback deleted ingress=%v root=%v", ingressDeleted, rootDeleted)
	}
}

type lifecycleExecutor struct {
	qdisc bool
	ifb   bool
	nft   bool
}

func (e *lifecycleExecutor) Run(_ context.Context, name string, args []string, _ []byte) (CommandResult, error) {
	if name == "tc" && len(args) >= 2 && args[0] == "qdisc" && args[1] == "show" {
		if len(args) >= 4 && args[3] == ifbName("eth0") {
			if e.ifb {
				return CommandResult{Stdout: []byte("qdisc htb 1: root")}, nil
			}
			return CommandResult{}, errors.New("does not exist")
		}
		if e.qdisc {
			return CommandResult{Stdout: []byte("qdisc htb 1: root\nqdisc ingress ffff: parent ffff:fff1")}, nil
		}
		return CommandResult{Stdout: []byte("qdisc noqueue 0: root")}, nil
	}
	if name == "ip" && len(args) >= 4 && args[0] == "link" && args[1] == "show" && args[3] == ifbName("eth0") {
		if e.ifb {
			return CommandResult{}, nil
		}
		return CommandResult{}, errors.New("does not exist")
	}
	if name == "tc" && len(args) >= 2 && args[0] == "qdisc" && args[1] == "replace" {
		if len(args) >= 4 && args[3] == ifbName("eth0") {
			e.ifb = true
		} else {
			e.qdisc = true
		}
	}
	if name == "tc" && len(args) >= 2 && args[0] == "qdisc" && args[1] == "del" {
		if len(args) >= 4 && args[3] == ifbName("eth0") {
			e.ifb = false
		} else {
			e.qdisc = false
		}
	}
	if name == "ip" && len(args) >= 3 && args[0] == "link" && args[1] == "add" {
		e.ifb = true
	}
	if name == "ip" && len(args) >= 3 && args[0] == "link" && args[1] == "del" {
		e.ifb = false
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
	if _, err := b.Reconcile(context.Background(), []DesiredRule{rule}); err != nil {
		t.Fatalf("second reconcile should be idempotent: %v", err)
	}
	SetBackendForTests(b)
	t.Cleanup(func() { Configure(false) })
	Configure(false)
	if e.qdisc || e.ifb || e.nft {
		t.Fatalf("disable left owned state behind: qdisc=%v ifb=%v nft=%v", e.qdisc, e.ifb, e.nft)
	}
}

func TestConfigureEnabledPreservesRollbackOwnershipAcrossPanelRestart(t *testing.T) {
	e := &lifecycleExecutor{}
	b := NewBackend(e, "linux", []string{"eth0"})
	rule := DesiredRule{NodeKey: "n", ClientKey: "c", Interface: "eth0", Mark: 10, UploadRateBps: 8000, DownloadRateBps: 8000, Selectors: []string{"192.0.2.1/32"}}
	if _, err := b.Reconcile(context.Background(), []DesiredRule{rule}); err != nil {
		t.Fatalf("apply owned state: %v", err)
	}
	SetBackendForTests(b)
	t.Cleanup(func() { Configure(false) })

	Configure(true)
	Configure(false)
	if e.qdisc || e.ifb || e.nft {
		t.Fatalf("enabled restart lost rollback ownership: qdisc=%v ifb=%v nft=%v", e.qdisc, e.ifb, e.nft)
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
	capability   []byte
	response     []byte
	capErr       error
	reconcileErr error
	calls        int
}

func (r *fakeRemote) TrafficControlCapabilities(context.Context) (json.RawMessage, error) {
	r.calls++
	return r.capability, r.capErr
}

func (r *fakeRemote) TrafficControlReconcile(context.Context, json.RawMessage) (json.RawMessage, error) {
	r.calls++
	return r.response, r.reconcileErr
}

func TestRemoteReconcileRefusesUnsupportedAndDegradedCapabilitiesWithoutMutation(t *testing.T) {
	tests := []struct {
		name       string
		remote     *fakeRemote
		wantState  State
		wantReason string
	}{
		{
			name:       "old node without endpoint",
			remote:     &fakeRemote{capErr: errors.New("remote request failed: HTTP 404")},
			wantState:  State(stateUnsupported),
			wantReason: "remote node does not support shaping",
		},
		{
			name:       "capability probe failure",
			remote:     &fakeRemote{capErr: errors.New("connection reset")},
			wantState:  State(stateDegraded),
			wantReason: "remote capability probe failed",
		},
		{
			name:       "node reports degraded",
			remote:     &fakeRemote{capability: []byte(`{"state":"degraded","reason":"missing CAP_NET_ADMIN"}`)},
			wantState:  State(stateDegraded),
			wantReason: "missing CAP_NET_ADMIN",
		},
		{
			name:       "node reports unsupported",
			remote:     &fakeRemote{capability: []byte(`{"state":"unsupported","reason":"platform is not linux"}`)},
			wantState:  State(stateUnsupported),
			wantReason: "platform is not linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := ReconcileRemote(context.Background(), tt.remote, []DesiredRule{{NodeKey: "n", ClientKey: "c"}})
			if err == nil {
				t.Fatal("ReconcileRemote error = nil, want explicit refusal")
			}
			if status.Capabilities.State != tt.wantState || status.Capabilities.Reason != tt.wantReason {
				t.Fatalf("status=%+v, want state=%q reason=%q", status, tt.wantState, tt.wantReason)
			}
			if tt.remote.calls != 1 {
				t.Fatalf("remote calls=%d, want capability probe only and no reconcile mutation", tt.remote.calls)
			}
		})
	}
}

func TestRemoteReconcileUsesCapabilityBoundary(t *testing.T) {
	r := &fakeRemote{capability: []byte(`{"state":"ready","platform":"linux"}`), response: []byte(`{"status":{"state":"ready","generation":2}}`)}
	status, err := ReconcileRemote(context.Background(), r, []DesiredRule{{NodeKey: "n", ClientKey: "c"}})
	if err != nil || status.Generation != 2 || r.calls != 2 {
		t.Fatalf("status=%+v err=%v calls=%d", status, err, r.calls)
	}
}
