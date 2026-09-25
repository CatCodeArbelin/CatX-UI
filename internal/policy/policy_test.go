package policy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:policy-test-" + strings.ReplaceAll(t.Name(), "/", "-") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("SQLite unavailable: %v", err)
	}
	if err := db.AutoMigrate(&model.ClientRecord{}, &model.ClientGroup{}); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.ClientRecord{Email: "alice@example.test", Group: "staff"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.ClientGroup{Name: "staff"}).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPolicyCRUDAndAssignments(t *testing.T) {
	db := testDB(t)
	r := NewRepository(db)
	ctx := context.Background()
	p := Policy{Name: "baseline", Spec: `{"action":"allow","categories":["social"]}`, Enabled: true}
	if err := r.CreatePolicy(ctx, &p); err != nil {
		t.Fatal(err)
	}
	if err := r.CreateAssignment(ctx, &PolicyAssignment{PolicyID: p.ID, TargetType: TargetClient, TargetRef: "alice@example.test", Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if err := r.CreateAssignment(ctx, &PolicyAssignment{PolicyID: p.ID, TargetType: TargetClient, TargetRef: "alice@example.test", Enabled: true}); err == nil {
		t.Fatal("duplicate assignment accepted")
	}
	if err := r.CreateAssignment(ctx, &PolicyAssignment{PolicyID: p.ID, TargetType: "node", TargetRef: "1", Enabled: true}); err == nil {
		t.Fatal("invalid target accepted")
	}
	if err := r.CreateOverride(ctx, &PolicyOverride{PolicyID: p.ID, TargetType: TargetTypeForTest(), TargetRef: "alice@example.test", Scope: ScopeDomain, Value: `{"domains":["example.test"]}`, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	resolved, err := r.Resolve(ctx, "alice@example.test", "staff", time.Now().UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved.Items) != 2 || resolved.Items[0].Source != "override" {
		t.Fatalf("unexpected precedence: %+v", resolved.Items)
	}
	decisions, err := r.ResolveDecisions(ctx, []string{"alice@example.test"}, time.Now().UnixMilli())
	if err != nil || len(decisions) != 1 || decisions[0].OverrideScope != ScopeDomain || len(decisions[0].Destinations) != 1 || decisions[0].Destinations[0] != "example.test" {
		t.Fatalf("domain override was not scoped: %+v, %v", decisions, err)
	}
	var before int64
	if err := db.Model(&Policy{}).Count(&before).Error; err != nil {
		t.Fatal(err)
	}
	hypothetical, err := r.ResolveDecision(ctx, "alice@example.test", "staff", time.Now().UnixMilli())
	if err != nil || hypothetical == nil || hypothetical.PolicyID != decisions[0].PolicyID {
		t.Fatalf("hypothetical resolution diverged: %+v, %v", hypothetical, err)
	}
	var after int64
	if err := db.Model(&Policy{}).Count(&after).Error; err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatal("hypothetical resolution mutated policy storage")
	}
	if _, err := json.Marshal(hypothetical); err != nil {
		t.Fatalf("decision explanation is not serializable: %v", err)
	}
}

func TargetTypeForTest() string { return TargetClient }

func TestTemporaryOverrideExpiryAndPersistence(t *testing.T) {
	db := testDB(t)
	r := NewRepository(db)
	ctx := context.Background()
	now := time.Now().UnixMilli()
	p := Policy{Name: "temporary", Spec: `{}`, Enabled: true}
	if err := r.CreatePolicy(ctx, &p); err != nil {
		t.Fatal(err)
	}
	tmp := TemporaryOverride{PolicyID: p.ID, TargetType: TargetGroup, TargetRef: "staff", Scope: ScopePolicy, Value: `{}`, StartsAt: now - 1000, ExpiresAt: now + 100000, Enabled: true}
	if err := r.CreateTemporaryOverride(ctx, &tmp); err != nil {
		t.Fatal(err)
	}
	active, err := r.Resolve(ctx, "alice@example.test", "staff", now)
	if err != nil || len(active.Items) != 1 {
		t.Fatalf("active override missing: %+v %v", active, err)
	}
	expired, err := r.Resolve(ctx, "alice@example.test", "staff", now+200000)
	if err != nil || len(expired.Items) != 0 {
		t.Fatalf("expired override remained active: %+v %v", expired, err)
	}
	var persisted TemporaryOverride
	if err := db.First(&persisted, tmp.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.ExpiresAt != tmp.ExpiresAt {
		t.Fatal("temporary override was not persisted")
	}
}

func TestValidation(t *testing.T) {
	db := testDB(t)
	r := NewRepository(db)
	ctx := context.Background()
	for _, p := range []Policy{{Name: "bad", Spec: `[]`}, {Name: "bad-action", Spec: `{"action":"block"}`}} {
		if err := r.CreatePolicy(ctx, &p); err == nil {
			t.Fatal("invalid policy accepted")
		}
	}
	if err := r.CreatePolicy(ctx, &Policy{Name: "valid", Spec: `{}`}); err != nil {
		t.Fatal(err)
	}
}

func TestDisabledNoOpAPI(t *testing.T) {
	Configure(nil, false)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/panel/api"))
	req := httptest.NewRequest(http.MethodGet, "/panel/api/policies", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != 200 || Enabled() {
		t.Fatalf("disabled API response: status=%d body=%s", res.Code, res.Body.String())
	}
}

func TestPrecedenceTieBreakIsDeterministic(t *testing.T) {
	items := []resolvedCandidate{
		{Source: "assignment", TargetType: TargetClient, Priority: 4, CreatedAt: 2, ID: 2},
		{Source: "temporary", TargetType: TargetGroup, Priority: 0, CreatedAt: 3, ID: 3},
		{Source: "override", TargetType: TargetClient, Priority: 1, CreatedAt: 1, ID: 1},
		{Source: "override", TargetType: TargetClient, Priority: 1, CreatedAt: 1, ID: 2},
	}
	sortCandidates(items)
	if items[0].Source != "temporary" || items[1].ID != 1 || items[2].ID != 2 || items[3].Source != "assignment" {
		t.Fatalf("unexpected deterministic order: %+v", items)
	}
}
