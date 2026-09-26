package trafficpolicy

import (
	"context"
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/forkext/trafficcontrol"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

type testAttributionProvider struct {
	identity Attribution
}

func (p testAttributionProvider) Resolve(context.Context, string) (Attribution, error) {
	return p.identity, nil
}

type testShaper struct {
	calls []trafficcontrol.DesiredRule
}

func (s *testShaper) Capabilities(context.Context) trafficcontrol.Capabilities {
	return trafficcontrol.Capabilities{State: "ready", UserAttribution: true}
}

func (s *testShaper) ApplyClientLimit(context.Context, trafficcontrol.DesiredRule) error { return nil }

func (s *testShaper) RemoveClientLimit(context.Context, string, string) error { return nil }

func (s *testShaper) Reconcile(_ context.Context, rules []trafficcontrol.DesiredRule) (trafficcontrol.Status, error) {
	s.calls = append([]trafficcontrol.DesiredRule(nil), rules...)
	return trafficcontrol.Status{Capabilities: s.Capabilities(context.Background())}, nil
}

func TestStageBAttributionTransitionsAndCleanup(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:traffic-policy-stage-b?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Policy{}, &State{}, &xray.ClientTraffic{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&xray.ClientTraffic{Email: "stage-b", Enable: true}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&Policy{ClientEmail: "stage-b", Enabled: true, WindowSeconds: 3600, QuotaBytes: 100, ActiveUploadBps: 1000, ActiveDownloadBps: 2000, ThrottleUploadBps: 100, ThrottleDownloadBps: 200}).Error; err != nil {
		t.Fatal(err)
	}
	Configure(db, true)
	t.Cleanup(func() {
		ConfigureAttributionProvider(nil)
		Configure(nil, false)
	})
	if err := ApplyDeltas(db, []*xray.ClientTraffic{{Email: "stage-b"}}); err != nil {
		t.Fatal(err)
	}
	provider := testAttributionProvider{identity: Attribution{Supported: true, Stable: true, NodeKey: "local", Interface: "eth0", Mark: 17, Selectors: []string{"192.0.2.17/32"}}}
	shaper := &testShaper{}
	freshShaper := &testShaper{}
	if status, err := ReconcileEnforcement(context.Background(), testAttributionProvider{identity: Attribution{Supported: false}}, freshShaper); err != nil || len(freshShaper.calls) != 0 || status.State != StateUnsupported {
		t.Fatalf("unsupported attribution must be an exact no-op: status=%+v calls=%+v err=%v", status, freshShaper.calls, err)
	}
	status, err := ReconcileEnforcement(context.Background(), provider, shaper)
	if err != nil || len(shaper.calls) != 1 || shaper.calls[0].UploadRateBps != 1000 || status.State != "ready" {
		t.Fatalf("active reconcile: status=%+v calls=%+v err=%v", status, shaper.calls, err)
	}
	if err := db.Model(&State{}).Where("client_email = ?", "stage-b").Updates(map[string]any{"lifecycle": StateThrottled, "owner": OwnerSoftQuota}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := ReconcileEnforcement(context.Background(), provider, shaper); err != nil || shaper.calls[0].UploadRateBps != 100 {
		t.Fatalf("throttled reconcile: calls=%+v err=%v", shaper.calls, err)
	}
	unsupported := testAttributionProvider{identity: Attribution{Supported: false}}
	if status, err := ReconcileEnforcement(context.Background(), unsupported, shaper); err != nil || len(shaper.calls) != 0 || status.State != StateUnsupported {
		t.Fatalf("unsupported cleanup: status=%+v calls=%+v err=%v", status, shaper.calls, err)
	}
	if status, err := ReconcileEnforcement(context.Background(), nil, shaper); err != nil || len(shaper.calls) != 0 || status.State != StateUnsupported {
		t.Fatalf("empty production provider must be no-op: status=%+v calls=%+v err=%v", status, shaper.calls, err)
	}
}

func TestFixedWindowAndSoftThrottleLifecycle(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:traffic-policy-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Policy{}, &State{}, &xray.ClientTraffic{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&xray.ClientTraffic{Email: "alice", Enable: true, Up: 100, Down: 100}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&Policy{ClientEmail: "alice", Enabled: true, WindowSeconds: 3600, QuotaBytes: 15}).Error; err != nil {
		t.Fatal(err)
	}
	Configure(db, true)
	t.Cleanup(func() { Configure(nil, false) })

	row := &xray.ClientTraffic{Email: "alice", Up: 10, Down: 10}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&xray.ClientTraffic{}).Where("email = ?", "alice").Updates(map[string]any{"up": 110, "down": 110}).Error; err != nil {
			return err
		}
		return ApplyDeltas(tx, []*xray.ClientTraffic{row})
	}); err != nil {
		t.Fatal(err)
	}
	v, err := Get("alice")
	if err != nil {
		t.Fatal(err)
	}
	if v.Lifecycle != StateThrottled || v.UsedBytes != 20 || v.Owner != OwnerSoftQuota {
		t.Fatalf("unexpected throttle state: %+v", v)
	}
	if err := Reset("alice"); err != nil {
		t.Fatal(err)
	}
	v, err = Get("alice")
	if err != nil || v.Lifecycle != StateActive || v.UsedBytes != 0 {
		t.Fatalf("reset did not release soft throttle: %+v, %v", v, err)
	}
	if err := db.Model(&xray.ClientTraffic{}).Where("email = ?", "alice").Update("enable", false).Error; err != nil {
		t.Fatal(err)
	}
	if err := ApplyDeltas(db, []*xray.ClientTraffic{{Email: "alice"}}); err != nil {
		t.Fatal(err)
	}
	if err := Reset("alice"); err != nil {
		t.Fatal(err)
	}
	v, err = Get("alice")
	if err != nil || v.Lifecycle != StateDisabled {
		t.Fatalf("reset changed an upstream disable: %+v, %v", v, err)
	}
}

func TestGenericUserMarkPathRemainsUnsupported(t *testing.T) {
	trafficcontrol.Configure(false)
	if _, err := DesiredRuleFor(View{}); !errors.Is(err, trafficcontrol.ErrUnsupported) {
		t.Fatalf("DesiredRuleFor error = %v", err)
	}
	if _, err := ReconcileRemote(context.TODO(), nil, View{}); !errors.Is(err, trafficcontrol.ErrUnsupported) {
		t.Fatalf("ReconcileRemote error = %v", err)
	}
}

func TestFixedWindowMath(t *testing.T) {
	start, end := window(3601, 3600)
	if start != 3600 || end != 7200 {
		t.Fatalf("window = %d..%d", start, end)
	}
	if got := usage(xray.ClientTraffic{Up: 100, Down: 50}, State{BaselineUp: 90, BaselineDown: 75}); got != 10 {
		t.Fatalf("usage after counter decrease = %d", got)
	}
}
