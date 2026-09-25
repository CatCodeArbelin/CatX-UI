package policycompiler

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/mhsanaei/3x-ui/v3/internal/policy"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

const RuleTagPrefix = "catx-policy-"

// Compile is a pure Xray-side operation. It copies only the mutable routing
// section, removes prior CatX rules, then appends a deterministic replacement.
func Compile(cfg *xray.Config, decisions []policy.Decision) (*xray.Config, error) {
	if cfg == nil || len(decisions) == 0 {
		return cfg, nil
	}
	routing := map[string]any{}
	if len(cfg.RouterConfig) != 0 && string(cfg.RouterConfig) != "null" {
		if err := json.Unmarshal(cfg.RouterConfig, &routing); err != nil {
			return cfg, fmt.Errorf("policy routing decode: %w", err)
		}
	}
	if routing == nil {
		routing = map[string]any{}
	}
	rules, ok := routing["rules"].([]any)
	if !ok && routing["rules"] != nil {
		return cfg, fmt.Errorf("policy routing rules have invalid shape")
	}
	kept := make([]any, 0, len(rules)+len(decisions))
	for _, raw := range rules {
		obj, ok := raw.(map[string]any)
		if !ok {
			return cfg, fmt.Errorf("policy routing rule has invalid shape")
		}
		if tag, _ := obj["ruleTag"].(string); strings.HasPrefix(tag, RuleTagPrefix) {
			continue
		}
		kept = append(kept, raw)
	}
	ordered := append([]policy.Decision(nil), decisions...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].ClientEmail != ordered[j].ClientEmail {
			return ordered[i].ClientEmail < ordered[j].ClientEmail
		}
		return ordered[i].PolicyID < ordered[j].PolicyID
	})
	generated := make(map[string]struct{})
	for _, d := range ordered {
		if d.Action != "allow" && d.Action != "deny" {
			return cfg, fmt.Errorf("policy %d: unsupported action %q", d.PolicyID, d.Action)
		}
		outbound, err := selectOutbound(cfg.OutboundConfigs, d.Action)
		if err != nil {
			return cfg, err
		}
		destinations, err := supportedDestinations(d)
		if err != nil {
			return cfg, err
		}
		for i, destination := range destinations {
			ruleTag := fmt.Sprintf("%s%d-%s-%d", RuleTagPrefix, d.PolicyID, stableClient(d.ClientEmail), i)
			if _, exists := generated[ruleTag]; exists {
				continue
			}
			generated[ruleTag] = struct{}{}
			selector := map[string]any{"type": "field", "user": []string{d.ClientEmail}, "outboundTag": outbound, "ruleTag": ruleTag}
			if strings.HasPrefix(destination, "ip:") {
				selector["ip"] = []string{strings.TrimPrefix(destination, "ip:")}
			} else {
				selector["domain"] = []string{destination}
			}
			kept = append(kept, selector)
		}
	}
	if len(kept) == len(rules) && len(decisions) > 0 {
		return cfg, nil
	}
	routing["rules"] = kept
	raw, err := json.Marshal(routing)
	if err != nil {
		return cfg, fmt.Errorf("policy routing encode: %w", err)
	}
	copyCfg := *cfg
	copyCfg.InboundConfigs = append([]xray.InboundConfig(nil), cfg.InboundConfigs...)
	copyCfg.RouterConfig = raw
	return &copyCfg, nil
}

func stableClient(email string) string {
	return strings.NewReplacer("@", "-", ".", "-", ":", "-").Replace(email)
}

func supportedDestinations(d policy.Decision) ([]string, error) {
	set := map[string]struct{}{}
	for _, raw := range d.Destinations {
		s := strings.ToLower(strings.TrimSpace(raw))
		if s == "" {
			continue
		}
		s = strings.TrimPrefix(s, "domain:")
		if strings.HasPrefix(s, "ip:") {
			if net.ParseIP(strings.TrimPrefix(s, "ip:")) == nil {
				return nil, fmt.Errorf("policy %d: malformed IP destination", d.PolicyID)
			}
		} else if strings.Contains(s, "/") {
			if _, _, err := net.ParseCIDR(s); err != nil {
				return nil, fmt.Errorf("policy %d: malformed CIDR destination", d.PolicyID)
			}
			s = "ip:" + s
		} else if strings.ContainsAny(s, " /\\") || !strings.Contains(s, ".") {
			return nil, fmt.Errorf("policy %d: unsupported destination %q", d.PolicyID, raw)
		}
		set[s] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out, nil
}

func selectOutbound(raw []byte, action string) (string, error) {
	var outbounds []map[string]any
	if err := json.Unmarshal(raw, &outbounds); err != nil {
		return "", fmt.Errorf("policy outbounds decode: %w", err)
	}
	want := "blocked"
	if action == "allow" {
		want = "direct"
	}
	for _, o := range outbounds {
		if tag, _ := o["tag"].(string); tag == want {
			return tag, nil
		}
	}
	for _, o := range outbounds {
		tag, _ := o["tag"].(string)
		proto, _ := o["protocol"].(string)
		if action == "allow" && strings.EqualFold(proto, "freedom") {
			return tag, nil
		}
		if action == "deny" && strings.EqualFold(proto, "blackhole") {
			return tag, nil
		}
	}
	return "", fmt.Errorf("no compatible %s outbound", action)
}
