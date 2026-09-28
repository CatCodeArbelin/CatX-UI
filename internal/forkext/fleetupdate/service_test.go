package fleetupdate

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

type testResolver struct{}

func (testResolver) Resolve(context.Context, string) (ReleaseSnapshot, error) {
	return ReleaseSnapshot{Channel: "stable", Tag: "v9.9.9", APIURL: "https://api.example/releases/1", HTMLURL: "https://example/releases/1", ChecksumVerified: true}, nil
}

func testService(t *testing.T) (*Service, *FakeExecutor) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:fleet-update-"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = db.AutoMigrate(&model.Node{}, &Campaign{}, &Target{}); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 4; i++ {
		n := model.Node{Id: i, Name: "node-" + string(rune('0'+i)), Enable: true, Status: "online", LastHeartbeat: time.Now().Unix(), PanelVersion: "v9.9." + string(rune('0'+i)), XrayState: "running"}
		if err = db.Create(&n).Error; err != nil {
			t.Fatal(err)
		}
	}
	f := &FakeExecutor{Nodes: map[int]ConvergenceResult{}, DispatchErr: map[int]error{}}
	s := New(db, true)
	s.SetMutationEnabled(true)
	s.SetResolver(testResolver{})
	s.SetExecutor(f)
	return s, f
}

func TestPlanSnapshotsAndBlocks(t *testing.T) {
	s, _ := testService(t)
	p, err := s.Plan(context.Background(), PlanRequest{Channel: "stable", NodeIDs: []int{1, 2}, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if p.Campaign.State != StateReady || len(p.Targets) != 2 {
		t.Fatalf("plan=%+v", p)
	}
	if p.Targets[0].InitialVersion == "" {
		t.Fatal("missing immutable version")
	}
}

func TestCanaryBatchAndParallelism(t *testing.T) {
	s, f := testService(t)
	p, err := s.Plan(context.Background(), PlanRequest{NodeIDs: []int{1, 2, 3, 4}, CanaryCount: 1, BatchSize: 2, MaxParallel: 2, ConfirmProduction: true})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Reconcile(context.Background(), p.Campaign.ID, "test"); err != nil {
		t.Fatal(err)
	}
	if len(f.Dispatches) != 1 {
		t.Fatalf("canary dispatches=%v targets=%+v", f.Dispatches, p.Targets)
	}
}

func TestFeatureDisabledIsNoop(t *testing.T) {
	s, _ := testService(t)
	s.enabled = false
	if _, err := s.Plan(context.Background(), PlanRequest{}); !errors.Is(err, ErrDisabled) {
		t.Fatalf("err=%v", err)
	}
}

func TestProductionMutationRequiresSettingAndConfirmation(t *testing.T) {
	s, _ := testService(t)
	s.SetMutationEnabled(false)
	if _, err := s.Plan(context.Background(), PlanRequest{NodeIDs: []int{1}, ConfirmProduction: true}); !errors.Is(err, ErrMutationDisabled) {
		t.Fatalf("without mutation setting error = %v, want gate", err)
	}
	s.SetMutationEnabled(true)
	if _, err := s.Plan(context.Background(), PlanRequest{NodeIDs: []int{1}}); !errors.Is(err, ErrMutationDisabled) {
		t.Fatalf("without confirmation error = %v, want gate", err)
	}
}

func TestDryRunNeverDispatches(t *testing.T) {
	s, f := testService(t)
	p, err := s.Plan(context.Background(), PlanRequest{NodeIDs: []int{1}, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Reconcile(context.Background(), p.Campaign.ID, "test"); err != nil {
		t.Fatal(err)
	}
	if len(f.Dispatches) != 0 {
		t.Fatalf("dry-run dispatches = %v, want none", f.Dispatches)
	}
}

func TestAmbiguousDispatchIsNotRedispatched(t *testing.T) {
	s, f := testService(t)
	f.DispatchErr[1] = ErrAmbiguousDispatch
	p, err := s.Plan(context.Background(), PlanRequest{NodeIDs: []int{1}, ConfirmProduction: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Reconcile(context.Background(), p.Campaign.ID, "test"); err != nil {
		t.Fatal(err)
	}
	if err := s.Reconcile(context.Background(), p.Campaign.ID, "test"); err != nil {
		t.Fatal(err)
	}
	if len(f.Dispatches) != 1 {
		t.Fatalf("ambiguous dispatches = %v, want exactly one", f.Dispatches)
	}
}

func TestDispatchWithoutRunIDIsAmbiguousAndNotRedispatched(t *testing.T) {
	s, f := testService(t)
	p, err := s.Plan(context.Background(), PlanRequest{NodeIDs: []int{1}, ConfirmProduction: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Reconcile(context.Background(), p.Campaign.ID, "test"); err != nil {
		t.Fatal(err)
	}
	if err := s.Reconcile(context.Background(), p.Campaign.ID, "test"); err != nil {
		t.Fatal(err)
	}
	if len(f.Dispatches) != 1 {
		t.Fatalf("missing-run-id dispatches = %v, want exactly one", f.Dispatches)
	}
	got, err := s.Get(context.Background(), p.Campaign.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Targets[0].State != StateUnknown || got.Targets[0].RunID != "" || got.Targets[0].DispatchStatus != "ambiguous" {
		t.Fatalf("target = %+v, want unknown ambiguous target without persisted run ID", got.Targets[0])
	}
}

func TestAbortAndRetry(t *testing.T) {
	s, _ := testService(t)
	p, err := s.Plan(context.Background(), PlanRequest{NodeIDs: []int{1}, ConfirmProduction: true})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Abort(context.Background(), p.Campaign.ID); err != nil {
		t.Fatal(err)
	}
	if err = s.Retry(context.Background(), p.Campaign.ID); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Get(context.Background(), p.Campaign.ID)
	if got.Campaign.State != StateReady {
		t.Fatalf("state=%s", got.Campaign.State)
	}
}
