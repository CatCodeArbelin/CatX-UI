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
	for _, iface := range b.interfaces {
		if err := b.reconcileInterface(ctx, iface, byIface[iface]); err != nil {
			status.State, status.Reason, status.LastError = State(stateDegraded), "apply failed", err.Error()
			_ = b.rollback(ctx, iface)
			b.status = status
			return status, err
		}
		allRules = append(allRules, byIface[iface]...)
		status.AppliedInterfaces = append(status.AppliedInterfaces, iface)
	}
	if err := b.applyNft(ctx, allRules); err != nil {
		status.State, status.Reason, status.LastError = State(stateDegraded), "apply failed", err.Error()
		if rollbackErr := b.rollback(ctx, ""); rollbackErr != nil {
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

func (b *Backend) reconcileInterface(ctx context.Context, iface string, rules []DesiredRule) error {
	qdisc, err := b.exec.Run(ctx, "tc", []string{"qdisc", "show", "dev", iface}, nil)
	if err != nil {
		return err
	}
	q := string(qdisc.Stdout)
	managed := strings.Contains(q, "noqueue")
	if !managed {
		managed, _ = b.ownedTable(ctx)
	}
	if !managed {
		return fmt.Errorf("refusing admin-owned qdisc on %s", iface)
	}
	if err := b.execSimple(ctx, "tc", []string{"qdisc", "replace", "dev", iface, "root", "handle", "1:", "htb", "default", "1"}); err != nil {
		return err
	}
	if err := b.execSimple(ctx, "tc", []string{"class", "replace", "dev", iface, "parent", "1:", "classid", "1:1", "htb", "rate", "1gbit"}); err != nil {
		return err
	}
	if err := b.execSimple(ctx, "ip", []string{"link", "add", ifbName(iface), "type", "ifb"}); err != nil && !strings.Contains(err.Error(), "File exists") {
		return err
	}
	if err := b.execSimple(ctx, "ip", []string{"link", "set", "dev", ifbName(iface), "up"}); err != nil {
		return err
	}
	if err := b.execSimple(ctx, "tc", []string{"qdisc", "replace", "dev", ifbName(iface), "root", "handle", "1:", "htb", "default", "1"}); err != nil {
		return err
	}
	if err := b.execSimple(ctx, "tc", []string{"class", "replace", "dev", ifbName(iface), "parent", "1:", "classid", "1:1", "htb", "rate", "1gbit"}); err != nil {
		return err
	}
	for _, rule := range rules {
		minor := strconv.FormatUint(uint64(rule.Mark), 10)
		if err := b.execSimple(ctx, "tc", []string{"class", "replace", "dev", iface, "parent", "1:", "classid", "1:" + minor, "htb", "rate", rate(rule.UploadRateBps)}); err != nil {
			return err
		}
		if err := b.execSimple(ctx, "tc", []string{"class", "replace", "dev", ifbName(iface), "parent", "1:", "classid", "1:" + minor, "htb", "rate", rate(rule.DownloadRateBps)}); err != nil {
			return err
		}
		if err := b.execSimple(ctx, "tc", []string{"filter", "replace", "dev", iface, "parent", "1:", "protocol", "all", "pref", "100", "handle", minor, "fw", "flowid", "1:" + minor}); err != nil {
			return err
		}
		if err := b.execSimple(ctx, "tc", []string{"filter", "replace", "dev", ifbName(iface), "parent", "1:", "protocol", "all", "pref", "100", "handle", minor, "fw", "flowid", "1:" + minor}); err != nil {
			return err
		}
	}
	if err := b.execSimple(ctx, "tc", []string{"filter", "replace", "dev", iface, "ingress", "protocol", "all", "pref", "10", "flower", "action", "mirred", "egress", "redirect", "dev", ifbName(iface)}); err != nil {
		return err
	}
	return nil
}

func (b *Backend) applyNft(ctx context.Context, rules []DesiredRule) error {
	var script strings.Builder
	table, err := b.exec.Run(ctx, "nft", []string{"list", "table", nftFamily, nftTable}, nil)
	if err != nil {
		script.WriteString("add table inet " + nftTable + " { comment \"catx-managed-v1\"; }\n")
		script.WriteString("add chain inet " + nftTable + " catx_mark { type filter hook forward priority -150; policy accept; }\n")
		script.WriteString("add chain inet " + nftTable + " catx_output { type route hook output priority -150; policy accept; }\n")
	} else if !strings.Contains(string(table.Stdout), "catx-managed-v1") {
		return fmt.Errorf("refusing nft table %s without CatX ownership marker", nftTable)
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
				return fmt.Errorf("invalid selector %q: %w", selector, err)
			}
			family := "ip"
			if ip.To4() == nil {
				family = "ip6"
			}
			script.WriteString("add rule inet " + nftTable + " catx_mark iifname \"" + rule.Interface + "\" " + family + " saddr " + selector + " ct mark set " + strconv.FormatUint(uint64(rule.Mark), 10) + " meta mark set ct mark\n")
		}
	}
	return b.execSimpleInput(ctx, "nft", []string{"-f", "-"}, []byte(script.String()))
}

func (b *Backend) rollback(ctx context.Context, iface string) error {
	var first error
	interfaces := b.interfaces
	if iface != "" {
		interfaces = []string{iface}
	}
	for _, managedIface := range interfaces {
		qdisc, err := b.exec.Run(ctx, "tc", []string{"qdisc", "show", "dev", managedIface}, nil)
		if err != nil {
			if first == nil {
				first = err
			}
			continue
		}
		if !strings.Contains(string(qdisc.Stdout), "noqueue") {
			for _, cmd := range [][]string{{"tc", "qdisc", "del", "dev", managedIface, "root"}, {"tc", "qdisc", "del", "dev", ifbName(managedIface), "root"}, {"ip", "link", "del", ifbName(managedIface)}} {
				if err := b.execSimple(ctx, cmd[0], cmd[1:]); err != nil && first == nil {
					first = err
				}
			}
		}
	}
	_ = b.execSimpleInput(ctx, "nft", []string{"delete", "table", nftFamily, nftTable}, nil)
	return first
}

func (b *Backend) removeOwnedState(ctx context.Context) error {
	owned, _ := b.ownedTable(ctx)
	if !owned {
		return nil
	}
	return b.rollback(ctx, "")
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
		n = "catx-ifb" + fmt.Sprintf("%x", sum[:])[:8]
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
