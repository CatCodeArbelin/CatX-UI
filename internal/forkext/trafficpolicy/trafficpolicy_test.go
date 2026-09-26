package trafficpolicy

import (
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/forkext/trafficcontrol"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
	if _, err := DesiredRuleFor(View{}); err != trafficcontrol.ErrUnsupported {
		t.Fatalf("DesiredRuleFor error = %v", err)
	}
	if _, err := ReconcileRemote(nil, nil, View{}); err != trafficcontrol.ErrUnsupported {
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
