package policycompiler

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/policy"
	"github.com/mhsanaei/3x-ui/v3/internal/util/json_util"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

func baseConfig() *xray.Config {
	return &xray.Config{
		RouterConfig:    json_util.RawMessage(`{"domainStrategy":"AsIs","rules":[{"type":"field","port":"80","outboundTag":"existing","ruleTag":"upstream-rule"}]}`),
		OutboundConfigs: json_util.RawMessage(`[{"protocol":"freedom","tag":"direct"},{"protocol":"blackhole","tag":"blocked"}]`),
		InboundConfigs:  []xray.InboundConfig{{Tag: "in", Settings: json_util.RawMessage(`{"clients":[{"email":"alice@example.test"},{"email":"bob@example.test"}]}`)}},
	}
}

func decision(email, action string, id uint, destinations ...string) policy.Decision {
	return policy.Decision{ClientEmail: email, Action: action, PolicyID: id, Destinations: destinations}
}

func rules(t *testing.T, cfg *xray.Config) []map[string]any {
	t.Helper()
	var routing struct {
		Rules []map[string]any `json:"rules"`
	}
	if err := json.Unmarshal(cfg.RouterConfig, &routing); err != nil {
		t.Fatal(err)
	}
	return routing.Rules
}

func TestCompileDisabledOrEmptyIsExactNoOp(t *testing.T) {
	cfg := baseConfig()
	if got, err := Compile(cfg, nil); err != nil || got != cfg {
		t.Fatalf("empty compile = %p, %v; want original %p", got, err, cfg)
	}
}

func TestCompilePreservesUpstreamAndIsIdempotent(t *testing.T) {
	cfg := baseConfig()
	original := *cfg
	got, err := Compile(cfg, []policy.Decision{decision("alice@example.test", "deny", 7, "example.com")})
	if err != nil {
		t.Fatal(err)
	}
	if got == cfg || !reflect.DeepEqual(cfg.InboundConfigs, original.InboundConfigs) || string(cfg.OutboundConfigs) != string(original.OutboundConfigs) {
		t.Fatal("compiler mutated upstream config")
	}
	r := rules(t, got)
	if len(r) != 2 || r[0]["ruleTag"] != "upstream-rule" || r[1]["outboundTag"] != "blocked" {
		t.Fatalf("rules = %#v", r)
	}
	again, err := Compile(got, []policy.Decision{decision("alice@example.test", "deny", 7, "example.com")})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(rules(t, got), rules(t, again)) {
		t.Fatal("repeated compilation changed rule structure")
	}
	for _, rule := range rules(t, again) {
		if tag, _ := rule["ruleTag"].(string); tag == "catx-policy-7-alice-example-test-0" {
			continue
		}
	}
}

func TestCompileDeterministicMultiClientIsolation(t *testing.T) {
	a, err := Compile(baseConfig(), []policy.Decision{decision("bob@example.test", "allow", 2, "b.example"), decision("alice@example.test", "deny", 1, "a.example")})
	if err != nil {
		t.Fatal(err)
	}
	r := rules(t, a)
	if r[1]["user"].([]any)[0] != "alice@example.test" || r[2]["user"].([]any)[0] != "bob@example.test" {
		t.Fatalf("not deterministic: %#v", r)
	}
}

func TestCompileMalformedFailsWithoutMutation(t *testing.T) {
	cfg := baseConfig()
	before := append([]byte(nil), cfg.RouterConfig...)
	if _, err := Compile(cfg, []policy.Decision{decision("alice@example.test", "deny", 1, "not-a-domain")}); err == nil {
		t.Fatal("malformed destination accepted")
	}
	if !reflect.DeepEqual(before, []byte(cfg.RouterConfig)) {
		t.Fatal("malformed compile mutated original")
	}
	if _, err := Compile(cfg, []policy.Decision{decision("alice@example.test", "unknown", 1, "example.com")}); err == nil {
		t.Fatal("malformed action accepted")
	}
}

func TestCompileCategoryOnlyIsConservativeNoOp(t *testing.T) {
	cfg := baseConfig()
	got, err := Compile(cfg, []policy.Decision{decision("alice@example.test", "deny", 1)})
	if err != nil || got != cfg {
		t.Fatalf("category-only decision must not create destructive rule: %v", err)
	}
}

func TestCompileDuplicateDecisionsDoNotDuplicateRules(t *testing.T) {
	cfg := baseConfig()
	got, err := Compile(cfg, []policy.Decision{
		decision("alice@example.test", "deny", 9, "example.com"),
		decision("alice@example.test", "deny", 9, "example.com"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotRules := rules(t, got); len(gotRules) != 2 {
		t.Fatalf("duplicate decisions emitted %d rules", len(gotRules))
	}
}

func TestCompileConcurrentCallsAreIndependent(t *testing.T) {
	decisions := []policy.Decision{decision("alice@example.test", "deny", 4, "example.com")}
	want, err := Compile(baseConfig(), decisions)
	if err != nil {
		t.Fatal(err)
	}
	wantRules := rules(t, want)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, compileErr := Compile(baseConfig(), decisions)
			if compileErr != nil || !reflect.DeepEqual(rules(t, got), wantRules) {
				t.Errorf("concurrent compile mismatch: %v", compileErr)
			}
		}()
	}
	wg.Wait()
}

func TestPreviewUsesStableCompilerRuleIdentity(t *testing.T) {
	decision := decision("alice@example.test", "allow", 12, "example.com")
	preview, err := Preview([]policy.Decision{decision})
	if err != nil || len(preview) != 1 {
		t.Fatalf("preview = %#v, err=%v", preview, err)
	}
	if preview[0].RuleTag != "catx-policy-12-alice-example-test-0" || preview[0].OutboundTag != "direct" || !preview[0].Emitted {
		t.Fatalf("unexpected preview: %#v", preview[0])
	}
}

func TestCompileQuarantineRulesPrecedeUpstreamRules(t *testing.T) {
	d := decision("alice@example.test", "deny", 21)
	d.Quarantined = true
	d.QuarantineAllowlist = []string{"support.example"}
	got, err := Compile(baseConfig(), []policy.Decision{d})
	if err != nil {
		t.Fatal(err)
	}
	r := rules(t, got)
	if r[0]["ruleTag"] != "catx-quarantine-alice-example-test-0" || r[1]["ruleTag"] != "catx-quarantine-alice-example-test" || r[2]["ruleTag"] != "upstream-rule" {
		t.Fatalf("quarantine ordering = %#v", r)
	}
	if r[1]["outboundTag"] != "blocked" {
		t.Fatalf("quarantine catch-all = %#v", r[1])
	}
}

func TestCompileManagedDNSUsesExistingDNSOutbound(t *testing.T) {
	cfg := baseConfig()
	cfg.OutboundConfigs = json_util.RawMessage(`[{"protocol":"freedom","tag":"direct"},{"protocol":"blackhole","tag":"blocked"},{"protocol":"dns","tag":"dns-out"}]`)
	d := decision("alice@example.test", "allow", 22, "example.com")
	d.ManagedDNS = true
	got, err := Compile(cfg, []policy.Decision{d})
	if err != nil {
		t.Fatal(err)
	}
	r := rules(t, got)
	if r[0]["outboundTag"] != "dns-out" || r[0]["port"] != "53" || r[0]["network"] != "tcp,udp" {
		t.Fatalf("dns rule = %#v", r[0])
	}
}

func TestCompileManagedDNSRequiresDNSOutbound(t *testing.T) {
	d := decision("alice@example.test", "allow", 24, "example.com")
	d.ManagedDNS = true
	if _, err := Compile(baseConfig(), []policy.Decision{d}); err == nil {
		t.Fatal("managed DNS compiled without a dns outbound")
	}
}

func TestCompileSafeSearchRequiresExplicitDNSOutboundProfile(t *testing.T) {
	cfg := baseConfig()
	cfg.OutboundConfigs = json_util.RawMessage(`[{"protocol":"freedom","tag":"direct"},{"protocol":"blackhole","tag":"blocked"},{"protocol":"dns","tag":"safe-dns"}]`)
	d := decision("alice@example.test", "allow", 27, "example.com")
	d.SafeSearch = true
	d.ManagedDNS = true
	if _, err := Compile(cfg, []policy.Decision{d}); err == nil {
		t.Fatal("SafeSearch compiled without explicit DNS profile")
	}
	d.DNSOutboundTag = "safe-dns"
	if _, err := Compile(cfg, []policy.Decision{d}); err != nil {
		t.Fatalf("explicit SafeSearch profile rejected: %v", err)
	}
}
