package groupquota

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newQuotaTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:groupquota_test_"+strings.ReplaceAll(t.Name(), "/", "_")+"_"+time.Now().Format("150405.000000000")+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		if strings.Contains(err.Error(), "CGO_ENABLED=0") {
			t.Skip("SQLite coverage requires cgo")
		}
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&State{}, &Membership{}, &model.ClientRecord{}, &xray.ClientTraffic{}); err != nil {
		t.Fatal(err)
	}
	Configure(db, true)
	t.Cleanup(func() { Configure(nil, false) })
	return db
}

func TestGroupQuotaSchemaPostgres(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("XUI_TEST_PG_DSN"))
	if dsn == "" {
		t.Skip("set XUI_TEST_PG_DSN to a reachable Postgres to run this test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db); err != nil {
		t.Fatal(err)
	}
	if !db.Migrator().HasTable(&State{}) || !db.Migrator().HasTable(&Membership{}) {
		t.Fatal("group quota tables were not migrated")
	}
}

func seedQuotaClient(t *testing.T, db *gorm.DB, up, down int64) {
	t.Helper()
	if err := db.Create(&model.ClientRecord{Email: "a@example", Group: "g", Enable: true}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&xray.ClientTraffic{Email: "a@example", Enable: true, Up: up, Down: down}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&State{GroupName: "g", QuotaBytes: 10_000, ActiveMultiplierPPM: Scale, PendingMultiplierPPM: Scale}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&Membership{GroupName: "g", ClientEmail: "a@example"}).Error; err != nil {
		t.Fatal(err)
	}
}

func TestPreserveResetCarriesUsageAndRebasesToPostResetCounter(t *testing.T) {
	db := newQuotaTestDB(t)
	seedQuotaClient(t, db, 100, 50)
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := PreserveReset(tx, "a@example", 0, 0); err != nil {
			return err
		}
		return tx.Model(&xray.ClientTraffic{}).Where("email = ?", "a@example").Updates(map[string]any{"up": 0, "down": 0}).Error
	}); err != nil {
		t.Fatal(err)
	}
	v, err := View(db, "g")
	if err != nil {
		t.Fatal(err)
	}
	if v.UsedBytes != 150 {
		t.Fatalf("used after reset = %d, want 150", v.UsedBytes)
	}
	if err := db.Model(&xray.ClientTraffic{}).Where("email = ?", "a@example").Updates(map[string]any{"up": 20, "down": 10}).Error; err != nil {
		t.Fatal(err)
	}
	v, err = View(db, "g")
	if err != nil {
		t.Fatal(err)
	}
	if v.UsedBytes != 180 {
		t.Fatalf("used after post-reset traffic = %d, want 180", v.UsedBytes)
	}
	var m Membership
	if err := db.Where("group_name = ? AND client_email = ?", "g", "a@example").First(&m).Error; err != nil {
		t.Fatal(err)
	}
	if m.BaseUp != 0 || m.BaseDown != 0 {
		t.Fatalf("membership base = (%d,%d), want post-reset (0,0)", m.BaseUp, m.BaseDown)
	}
}

func TestMultiplierChangeIsPendingUntilReset(t *testing.T) {
	db := newQuotaTestDB(t)
	seedQuotaClient(t, db, 3, 0)
	if err := db.Transaction(func(tx *gorm.DB) error {
		return UpsertConfig(tx, "g", 10_000, 2*Scale, periodNever, 1)
	}); err != nil {
		t.Fatal(err)
	}
	v, err := View(db, "g")
	if err != nil {
		t.Fatal(err)
	}
	if v.UsedBytes != 3 || v.ActiveMultiplierPPM != Scale || v.PendingMultiplierPPM != 2*Scale {
		t.Fatalf("pending multiplier repriced current period: %+v", v)
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		_, err := Reset(tx, "g")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	v, err = View(db, "g")
	if err != nil {
		t.Fatal(err)
	}
	if v.UsedBytes != 0 || v.ActiveMultiplierPPM != 2*Scale {
		t.Fatalf("reset did not activate pending multiplier: %+v", v)
	}
}

func TestConcurrentDeltasDoNotLoseQuotaUsage(t *testing.T) {
	db := newQuotaTestDB(t)
	seedQuotaClient(t, db, 0, 0)
	const workers = 4
	const updatesPerWorker = 8
	var wg sync.WaitGroup
	errs := make(chan error, workers*updatesPerWorker)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < updatesPerWorker; j++ {
				var err error
				for attempt := 0; attempt < 5; attempt++ {
					err = db.Transaction(func(tx *gorm.DB) error {
						if err := tx.Model(&xray.ClientTraffic{}).Where("email = ?", "a@example").UpdateColumn("up", gorm.Expr("up + ?", 1)).Error; err != nil {
							return err
						}
						return ApplyClientDeltas(tx, []*xray.ClientTraffic{{Email: "a@example", Up: 1}})
					})
					if err == nil {
						break
					}
					time.Sleep(time.Millisecond)
				}
				if err != nil {
					errs <- err
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	v, err := View(db, "g")
	if err != nil {
		t.Fatal(err)
	}
	if v.UsedBytes != workers*updatesPerWorker {
		t.Fatalf("concurrent used = %d, want %d", v.UsedBytes, workers*updatesPerWorker)
	}
}
