package trafficcontrol

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	nftFamily = "inet"
	nftTable  = "catx_traffic_control"
)

type Backend struct {
	mu         sync.Mutex
	exec       Executor
	platform   string
	interfaces []string
	status     Status
	desired    map[string]DesiredRule
}

type interfacePlan struct {
	rootOwned    bool
	ingressOwned bool
	ifbOwned     bool
	ifbPresent   bool
}

type interfaceMutation struct {
	iface          string
	ifb            string
	mutated        bool
	rootCreated    bool
	ingressCreated bool
	ifbCreated     bool
	ifbRootCreated bool
}

func (m interfaceMutation) changed() bool {
	return m.mutated || m.rootCreated || m.ingressCreated || m.ifbCreated || m.ifbRootCreated
}

func NewBackend(e Executor, platform string, interfaces []string) *Backend {
	if e == nil {
		e = processExecutor{}
	}
	if platform == "" {
		platform = runtime.GOOS
	}
	return &Backend{exec: e, platform: platform, interfaces: cleanInterfaces(interfaces), desired: map[string]DesiredRule{}}
}

func cleanInterfaces(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v != "" && len(v) <= 15 && validInterfaceName(v) && !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	sort.Strings(out)
	return out
}

func validInterfaceName(name string) bool {
	for _, r := range name {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' && r != '-' && r != '.' && r != ':' {
			return false
		}
	}
	return true
}

func (b *Backend) Capabilities(ctx context.Context) Capabilities {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.capabilitiesLocked(ctx)
}

func (b *Backend) capabilitiesLocked(ctx context.Context) Capabilities {
	cap := Capabilities{Platform: b.platform, Interfaces: append([]string(nil), b.interfaces...)}
	if b.platform != "linux" {
		cap.State, cap.Reason = State(stateUnsupported), "Linux is required"
		return cap
	}
	if len(b.interfaces) == 0 {
		cap.State, cap.Reason = State(stateUnsupported), "no explicit managed interface configured"
		return cap
	}
	for _, check := range []struct {
		name string
		args []string
		dst  *bool
	}{
		{"tc", []string{"-V"}, &cap.Tc}, {"nft", []string{"--version"}, &cap.Nftables},
	} {
		if _, err := b.exec.Run(ctx, check.name, check.args, nil); err == nil {
			*check.dst = true
		}
	}
	for _, iface := range b.interfaces {
		if _, err := b.exec.Run(ctx, "ip", []string{"link", "show", "dev", iface}, nil); err != nil {
			cap.State, cap.Reason = State(stateUnsupported), "managed interface unavailable: "+iface
			return cap
		}
	}
	cap.NetAdmin = b.probeNetAdmin(ctx)
	cap.ConntrackMarks, cap.IFB = cap.Nftables && cap.Tc, cap.Nftables && cap.Tc
	if !cap.Tc || !cap.Nftables || !cap.NetAdmin {
		cap.State, cap.Reason = State(stateUnsupported), "tc, nftables, and CAP_NET_ADMIN are required"
		return cap
	}
	cap.State = State(stateReady)
	return cap
}

func (b *Backend) probeNetAdmin(ctx context.Context) bool {
	_, err := b.exec.Run(ctx, "tc", []string{"qdisc", "show", "dev", b.interfaces[0]}, nil)
	return err == nil
}

func (b *Backend) Reconcile(ctx context.Context, rules []DesiredRule) (Status, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if rules == nil {
		rules = make([]DesiredRule, 0, len(b.desired))
		for _, rule := range b.desired {
			rules = append(rules, rule)
		}
	}
	cap := b.capabilitiesLocked(ctx)
	status := Status{Capabilities: cap, Generation: b.status.Generation}
	if cap.State != State(stateReady) {
		b.status = status
		return status, nil
	}
	if len(rules) == 0 {
		if err := b.removeOwnedState(ctx); err != nil {
			status.State, status.Reason, status.LastError = State(stateDegraded), "cleanup failed", err.Error()
			b.status = status
			return status, err
		}
		b.desired = map[string]DesiredRule{}
		status.Generation++
		b.status = status
		return status, nil
	}
	byIface := map[string][]DesiredRule{}
	marks := map[uint32]string{}
	for _, rule := range rules {
		if !contains(b.interfaces, rule.Interface) {
			status.State, status.Reason = State(stateUnsupported), "rule targets an unmanaged interface"
			b.status = status
			return status, ErrUnsupported
		}
		if rule.Mark == 0 || rule.Mark > 65534 || rule.UploadRateBps == 0 || rule.DownloadRateBps == 0 {
			status.State, status.Reason = State(stateUnsupported), "invalid substrate rule"
			b.status = status
			return status, ErrUnsupported
		}
		owner := rule.NodeKey + "\x00" + rule.ClientKey
		if previous, ok := marks[rule.Mark]; ok && previous != owner {
			status.State, status.Reason = State(stateUnsupported), "mark collision"
			b.status = status
			return status, ErrUnsupported
		}
		marks[rule.Mark] = owner
		byIface[rule.Interface] = append(byIface[rule.Interface], rule)
	}
	allRules := make([]DesiredRule, 0, len(rules))
	mutations := make([]interfaceMutation, 0, len(b.interfaces))
	for _, iface := range b.interfaces {
		mutation, err := b.reconcileInterface(ctx, iface, byIface[iface])
		if mutation.changed() {
			mutations = append(mutations, mutation)
		}
		if err != nil {
			status.State, status.Reason, status.LastError = State(stateDegraded), "apply failed", err.Error()
			if rollbackErr := b.rollbackMutations(ctx, mutations, false); rollbackErr != nil {
				status.LastError += "; rollback: " + rollbackErr.Error()
			}
			b.status = status
			return status, err
		}
		allRules = append(allRules, byIface[iface]...)
		status.AppliedInterfaces = append(status.AppliedInterfaces, iface)
	}
	nftCreated, err := b.applyNft(ctx, allRules)
	if err != nil {
		status.State, status.Reason, status.LastError = State(stateDegraded), "apply failed", err.Error()
		if rollbackErr := b.rollbackMutations(ctx, mutations, nftCreated); rollbackErr != nil {
			status.LastError += "; rollback: " + rollbackErr.Error()
		}
		b.status = status
		return status, err
	}
	status.Generation++
	b.desired = make(map[string]DesiredRule, len(rules))
	for _, rule := range rules {
		b.desired[rule.NodeKey+"\x00"+rule.ClientKey] = rule
	}
	b.status = status
	return status, nil
}

func (b *Backend) ApplyClientLimit(ctx context.Context, rule DesiredRule) error {
	b.mu.Lock()
	key := rule.NodeKey + "\x00" + rule.ClientKey
	rules := make([]DesiredRule, 0, len(b.desired))
	for _, item := range b.desired {
		rules = append(rules, item)
	}
	var replaced bool
	for i := range rules {
		if rules[i].NodeKey+"\x00"+rules[i].ClientKey == key {
			rules[i] = rule
			replaced = true
		}
	}
	if !replaced {
		rules = append(rules, rule)
	}
	b.mu.Unlock()
	_, err := b.Reconcile(ctx, rules)
	return err
}

func (b *Backend) RemoveClientLimit(ctx context.Context, nodeKey, clientKey string) error {
	b.mu.Lock()
	rules := make([]DesiredRule, 0, len(b.desired))
	for _, item := range b.desired {
		if item.NodeKey+"\x00"+item.ClientKey != nodeKey+"\x00"+clientKey {
			rules = append(rules, item)
		}
	}
	b.mu.Unlock()
	_, err := b.Reconcile(ctx, rules)
	return err
}

func (b *Backend) reconcileInterface(ctx context.Context, iface string, rules []DesiredRule) (interfaceMutation, error) {
	mutation := interfaceMutation{iface: iface, ifb: ifbName(iface)}
	plan, err := b.preflightInterface(ctx, iface)
	if err != nil {
		return mutation, err
	}
	apply := func(name string, args []string) error {
		// Mark before execution: a command can fail after changing kernel state.
		mutation.mutated = true
		return b.execSimple(ctx, name, args)
	}

	if !plan.rootOwned {
		mutation.rootCreated = true
		if err := apply("tc", []string{"qdisc", "replace", "dev", iface, "root", "handle", "1:", "htb", "default", "1"}); err != nil {
			return mutation, err
		}
	}
	if err := apply("tc", []string{"class", "replace", "dev", iface, "parent", "1:", "classid", "1:1", "htb", "rate", "1gbit"}); err != nil {
		return mutation, err
	}
	if !plan.ifbPresent {
		mutation.ifbCreated = true
		if err := apply("ip", []string{"link", "add", mutation.ifb, "type", "ifb"}); err != nil {
			if strings.Contains(err.Error(), "File exists") {
				mutation.ifbCreated = false
			}
			return mutation, err
		}
	}
	if err := apply("ip", []string{"link", "set", "dev", mutation.ifb, "up"}); err != nil {
		return mutation, err
	}
	if !plan.ifbOwned {
		mutation.ifbRootCreated = true
		if err := apply("tc", []string{"qdisc", "replace", "dev", mutation.ifb, "root", "handle", "1:", "htb", "default", "1"}); err != nil {
			return mutation, err
		}
	}
	if err := apply("tc", []string{"class", "replace", "dev", mutation.ifb, "parent", "1:", "classid", "1:1", "htb", "rate", "1gbit"}); err != nil {
		return mutation, err
	}
	if !plan.ingressOwned {
		mutation.ingressCreated = true
		if err := apply("tc", []string{"qdisc", "replace", "dev", iface, "handle", "ffff:", "ingress"}); err != nil {
			return mutation, err
		}
	}
	for _, rule := range rules {
		minor := strconv.FormatUint(uint64(rule.Mark), 10)
		if err := apply("tc", []string{"class", "replace", "dev", iface, "parent", "1:", "classid", "1:" + minor, "htb", "rate", rate(rule.UploadRateBps)}); err != nil {
			return mutation, err
		}
		if err := apply("tc", []string{"class", "replace", "dev", mutation.ifb, "parent", "1:", "classid", "1:" + minor, "htb", "rate", rate(rule.DownloadRateBps)}); err != nil {
			return mutation, err
		}
		if err := apply("tc", []string{"filter", "replace", "dev", iface, "parent", "1:", "protocol", "all", "pref", "100", "handle", minor, "fw", "flowid", "1:" + minor}); err != nil {
			return mutation, err
		}
		if err := apply("tc", []string{"filter", "replace", "dev", mutation.ifb, "parent", "1:", "protocol", "all", "pref", "100", "handle", minor, "fw", "flowid", "1:" + minor}); err != nil {
			return mutation, err
		}
	}
	if err := apply("tc", []string{"filter", "replace", "dev", iface, "ingress", "protocol", "all", "pref", "10", "handle", "1", "flower", "action", "mirred", "egress", "redirect", "dev", mutation.ifb}); err != nil {
		return mutation, err
	}
	return mutation, nil
}

func (b *Backend) preflightInterface(ctx context.Context, iface string) (interfacePlan, error) {
	qdisc, err := b.exec.Run(ctx, "tc", []string{"qdisc", "show", "dev", iface}, nil)
	if err != nil {
		return interfacePlan{}, err
	}
	q := string(qdisc.Stdout)
	owned, _ := b.ownedTable(ctx)
	rootNoqueue := hasRootQdisc(q, "noqueue")
	rootOwned := owned && isCatXRootQdisc(q)
	if hasQdiscKind(q, "clsact") {
		return interfacePlan{}, fmt.Errorf("refusing admin-owned clsact qdisc on %s", iface)
	}
	if rootNoqueue {
		if hasQdiscKind(q, "ingress") {
			return interfacePlan{}, fmt.Errorf("refusing admin-owned ingress qdisc on %s", iface)
		}
	} else if !rootOwned {
		return interfacePlan{}, fmt.Errorf("refusing admin-owned qdisc on %s", iface)
	}

	plan := interfacePlan{rootOwned: rootOwned}
	if hasQdiscKind(q, "ingress") {
		if !rootOwned || !hasQdisc(q, "ingress", "ffff:") {
			return interfacePlan{}, fmt.Errorf("refusing admin-owned ingress qdisc on %s", iface)
		}
		plan.ingressOwned = true
	}

	ifb := ifbName(iface)
	_, err = b.exec.Run(ctx, "ip", []string{"link", "show", "dev", ifb}, nil)
	if err == nil {
		plan.ifbPresent = true
		if !rootOwned || !owned {
			return interfacePlan{}, fmt.Errorf("refusing pre-existing IFB %s on %s", ifb, iface)
		}
		ifbQdisc, err := b.exec.Run(ctx, "tc", []string{"qdisc", "show", "dev", ifb}, nil)
		if err != nil {
			return interfacePlan{}, err
		}
		if !isCatXRootQdisc(string(ifbQdisc.Stdout)) {
			return interfacePlan{}, fmt.Errorf("refusing non-CatX IFB %s", ifb)
		}
		plan.ifbOwned = true
	} else {
		failure := strings.ToLower(err.Error())
		if !strings.Contains(failure, "does not exist") && !strings.Contains(failure, "cannot find device") && !strings.Contains(failure, "not found") {
			return interfacePlan{}, fmt.Errorf("checking IFB %s failed: %w", ifb, err)
		}
	}
	return plan, nil
}

func hasQdisc(qdisc, kind, handle string) bool {
	for _, line := range strings.Split(qdisc, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[0] == "qdisc" && fields[1] == kind && fields[2] == handle {
			return true
		}
	}
	return false
}

func hasQdiscKind(qdisc, kind string) bool {
	for _, line := range strings.Split(qdisc, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "qdisc" && fields[1] == kind {
			return true
		}
	}
	return false
}

func hasRootQdisc(qdisc, kind string) bool {
	for _, line := range strings.Split(qdisc, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[0] == "qdisc" && fields[1] == kind && contains(fields[3:], "root") {
			return true
		}
	}
	return false
}

func isCatXRootQdisc(qdisc string) bool {
	return hasQdisc(qdisc, "htb", "1:") && hasRootQdisc(qdisc, "htb")
}

func (b *Backend) rollbackMutations(ctx context.Context, mutations []interfaceMutation, nftCreated bool) error {
	var first error
	for _, mutation := range mutations {
		commands := make([][]string, 0, 4)
		if mutation.ingressCreated {
			commands = append(commands, []string{"tc", "qdisc", "del", "dev", mutation.iface, "ingress"})
		}
		if mutation.rootCreated {
			commands = append(commands, []string{"tc", "qdisc", "del", "dev", mutation.iface, "root"})
		}
		if mutation.ifbRootCreated {
			commands = append(commands, []string{"tc", "qdisc", "del", "dev", mutation.ifb, "root"})
		}
		if mutation.ifbCreated {
			commands = append(commands, []string{"ip", "link", "del", mutation.ifb})
		}
		for _, command := range commands {
			if err := b.execSimple(ctx, command[0], command[1:]); err != nil && first == nil {
				first = err
			}
		}
	}
	if nftCreated {
		if err := b.deleteOwnedNftTable(ctx); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (b *Backend) deleteOwnedNftTable(ctx context.Context) error {
	owned, err := b.ownedTable(ctx)
	if err != nil || !owned {
		return nil
	}
	return b.execSimpleInput(ctx, "nft", []string{"delete", "table", nftFamily, nftTable}, nil)
}

func (b *Backend) applyNft(ctx context.Context, rules []DesiredRule) (bool, error) {
	var script strings.Builder
	table, err := b.exec.Run(ctx, "nft", []string{"list", "table", nftFamily, nftTable}, nil)
	nftCreated := err != nil
	if err != nil {
		script.WriteString("add table inet " + nftTable + " { comment \"catx-managed-v1\"; }\n")
		script.WriteString("add chain inet " + nftTable + " catx_mark { type filter hook forward priority -150; policy accept; }\n")
		script.WriteString("add chain inet " + nftTable + " catx_output { type route hook output priority -150; policy accept; }\n")
	} else if !strings.Contains(string(table.Stdout), "catx-managed-v1") {
		return false, fmt.Errorf("refusing nft table %s without CatX ownership marker", nftTable)
	} else if !strings.Contains(string(table.Stdout), "catx_mark") || !strings.Contains(string(table.Stdout), "catx_output") {
		script.WriteString("add chain inet " + nftTable + " catx_mark { type filter hook forward priority -150; policy accept; }\n")
		script.WriteString("add chain inet " + nftTable + " catx_output { type route hook output priority -150; policy accept; }\n")
	}
	script.WriteString("flush chain inet " + nftTable + " catx_mark\nflush chain inet " + nftTable + " catx_output\n")
	script.WriteString("add rule inet " + nftTable + " catx_mark ct mark != 0 meta mark set ct mark\n")
	script.WriteString("add rule inet " + nftTable + " catx_output ct mark != 0 meta mark set ct mark\n")
	for _, rule := range rules {
		for _, selector := range rule.Selectors {
			ip, _, err := net.ParseCIDR(selector)
			if err != nil {
				return nftCreated, fmt.Errorf("invalid selector %q: %w", selector, err)
			}
			family := "ip"
			if ip.To4() == nil {
				family = "ip6"
			}
			script.WriteString("add rule inet " + nftTable + " catx_mark iifname \"" + rule.Interface + "\" " + family + " saddr " + selector + " ct mark set " + strconv.FormatUint(uint64(rule.Mark), 10) + " meta mark set ct mark\n")
		}
	}
	return nftCreated, b.execSimpleInput(ctx, "nft", []string{"-f", "-"}, []byte(script.String()))
}

func (b *Backend) removeOwnedState(ctx context.Context) error {
	owned, err := b.ownedTable(ctx)
	if err != nil || !owned {
		return nil
	}
	var first error
	for _, iface := range b.interfaces {
		plan, err := b.preflightInterface(ctx, iface)
		if err != nil {
			// Refusal is deliberately non-destructive. The nft marker may be
			// stale, but an administrator-owned qdisc must remain untouched.
			continue
		}
		for _, command := range []struct {
			args  []string
			owned bool
		}{
			{[]string{"tc", "qdisc", "del", "dev", iface, "ingress"}, plan.ingressOwned},
			{[]string{"tc", "qdisc", "del", "dev", iface, "root"}, plan.rootOwned},
		} {
			if !command.owned {
				continue
			}
			if err := b.execSimple(ctx, command.args[0], command.args[1:]); err != nil && first == nil {
				first = err
			}
		}
		if plan.ifbOwned {
			if err := b.execSimple(ctx, "tc", []string{"qdisc", "del", "dev", ifbName(iface), "root"}); err != nil && first == nil {
				first = err
			}
			if err := b.execSimple(ctx, "ip", []string{"link", "del", ifbName(iface)}); err != nil && first == nil {
				first = err
			}
		}
	}
	if err := b.deleteOwnedNftTable(ctx); err != nil && first == nil {
		first = err
	}
	return first
}

func (b *Backend) ownedTable(ctx context.Context) (bool, error) {
	result, err := b.exec.Run(ctx, "nft", []string{"list", "table", nftFamily, nftTable}, nil)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(result.Stdout), "catx-managed-v1"), nil
}

func (b *Backend) execSimple(ctx context.Context, name string, args []string) error {
	_, err := b.exec.Run(ctx, name, args, nil)
	return err
}

func (b *Backend) execSimpleInput(ctx context.Context, name string, args []string, input []byte) error {
	_, err := b.exec.Run(ctx, name, args, input)
	return err
}

func ifbName(iface string) string {
	n := "catx-" + iface
	if len(n) > 15 {
		sum := sha256.Sum256([]byte(iface))
		// Linux IFNAMSIZ allows at most 15 visible characters. Keep the
		// deterministic collision-resistant suffix within that limit.
		n = "catx-ifb" + fmt.Sprintf("%x", sum[:])[:7]
	}
	return n
}

func rate(bps uint64) string {
	if bps < 8 {
		return "1bit"
	}
	return strconv.FormatUint((bps+7)/8, 10) + "bit"
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

var _ Shaper = (*Backend)(nil)
