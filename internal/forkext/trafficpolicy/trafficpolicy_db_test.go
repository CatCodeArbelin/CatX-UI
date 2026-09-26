package trafficpolicy

import (
	"os"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTrafficPolicySchemaPostgres(t *testing.T) {
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
	if !db.Migrator().HasTable(&Policy{}) || !db.Migrator().HasTable(&State{}) {
		t.Fatal("traffic policy tables were not migrated")
	}
}
