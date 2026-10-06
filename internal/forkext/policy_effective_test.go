package forkext

import (
	"context"
	"encoding/json"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/policy"
	"github.com/mhsanaei/3x-ui/v3/internal/policycompiler"
	"github.com/mhsanaei/3x-ui/v3/internal/util/json_util"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

func effectivePolicyConfig() *xray.Config {
	return &xray.Config{
		InboundConfigs: []xray.InboundConfig{{
			Tag:      "inbound",
			Settings: json_util.RawMessage(`{"clients":[{"email":"alice@example.test"}]}`),
		}},
		OutboundConfigs: json_util.RawMessage(`[{"protocol":"freedom","tag":"direct"},{"protocol":"blackhole","tag":"blocked"}]`),
		RouterConfig:    json_util.RawMessage(`{"rules":[{"ruleTag":"upstream-rule","outboundTag":"direct"}]}`),
	}
}

func effectivePolicyRules(t *testing.T, cfg *xray.Config) []map[string]any {
	t.Helper()
	var routing struct {
		Rules []map[string]any `json:"rules"`
	}
	if err := json.Unmarshal(cfg.RouterConfig, &routing); err != nil {
		t.Fatal(err)
	}
	return routing.Rules
}

func TestDecorateXrayConfigUsesEffectivePolicyAndRemovesStaleDelta(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:forkext-effective-policy?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.ClientRecord{}); err != nil {
		t.Fatalf("migrate client record: %v", err)
	}
	if err := policy.Migrate(db); err != nil {
		t.Fatalf("migrate policy schema: %v", err)
	}
	if err := db.Create(&model.ClientRecord{Email: "alice@example.test", Group: "staff"}).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}
	policy.Configure(db, true)
	t.Cleanup(func() { policy.Configure(nil, false) })
	repo := policy.NewRepository(db)
	ctx := context.Background()
	definition := policy.Policy{Name: "effective", Spec: `{"action":"deny","destinations":["example.com"]}`, Enabled: true}
	if err := repo.CreatePolicy(ctx, &definition); err != nil {
		t.Fatalf("create policy: %v", err)
	}
	assignment := policy.PolicyAssignment{PolicyID: definition.ID, TargetType: policy.TargetClient, TargetRef: "alice@example.test", Enabled: true}
	if err := repo.CreateAssignment(ctx, &assignment); err != nil {
		t.Fatalf("create assignment: %v", err)
	}

	denied, err := DecorateXrayConfig(ctx, effectivePolicyConfig())
	if err != nil {
		t.Fatalf("decorate deny config: %v", err)
	}
	if !hasPolicyOutbound(effectivePolicyRules(t, denied), "blocked") {
		t.Fatalf("effective deny rule missing: %#v", effectivePolicyRules(t, denied))
	}

	definition.Spec = `{"action":"allow","destinations":["example.com"]}`
	if err := repo.UpdatePolicy(ctx, &definition); err != nil {
		t.Fatalf("update policy: %v", err)
	}
	allowed, err := DecorateXrayConfig(ctx, denied)
	if err != nil {
		t.Fatalf("decorate updated config: %v", err)
	}
	if !hasPolicyOutbound(effectivePolicyRules(t, allowed), "direct") || hasPolicyOutbound(effectivePolicyRules(t, allowed), "blocked") {
		t.Fatalf("effective update did not replace deny rule: %#v", effectivePolicyRules(t, allowed))
	}

	if err := repo.DeleteAssignment(ctx, assignment.ID); err != nil {
		t.Fatalf("delete assignment: %v", err)
	}
	stripped, err := DecorateXrayConfig(ctx, allowed)
	if err != nil {
		t.Fatalf("decorate after policy removal: %v", err)
	}
	assertNoPolicyRules(t, stripped)

	policy.Configure(db, false)
	disabled, err := DecorateXrayConfig(ctx, allowed)
	if err != nil {
		t.Fatalf("decorate disabled policy: %v", err)
	}
	assertNoPolicyRules(t, disabled)
}

func hasPolicyOutbound(rules []map[string]any, outbound string) bool {
	for _, rule := range rules {
		if tag, _ := rule["ruleTag"].(string); len(tag) >= len(policycompiler.RuleTagPrefix) && tag[:len(policycompiler.RuleTagPrefix)] == policycompiler.RuleTagPrefix {
			if rule["outboundTag"] == outbound {
				return true
			}
		}
	}
	return false
}

func assertNoPolicyRules(t *testing.T, cfg *xray.Config) {
	t.Helper()
	for _, rule := range effectivePolicyRules(t, cfg) {
		tag, _ := rule["ruleTag"].(string)
		if len(tag) >= len(policycompiler.RuleTagPrefix) && tag[:len(policycompiler.RuleTagPrefix)] == policycompiler.RuleTagPrefix {
			t.Fatalf("stale CatX policy rule remained: %#v", rule)
		}
	}
	if len(effectivePolicyRules(t, cfg)) != 1 {
		t.Fatalf("upstream routing rule was not preserved: %#v", effectivePolicyRules(t, cfg))
	}
}
