package portal

import (
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:portal-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.ClientRecord{}, &model.ClientGroup{}, &model.Host{}, &Credential{}, &HostGrant{}); err != nil {
		t.Fatal(err)
	}
	Configure(db, true)
	t.Cleanup(func() { Configure(nil, false) })
	return db
}

func TestIssueRotationStoresOnlyHashAndRevocationInvalidatesVersion(t *testing.T) {
	db := testDB(t)
	client := model.ClientRecord{Email: "portal@example.test", SubID: "sub-1", Enable: true}
	if err := db.Create(&client).Error; err != nil {
		t.Fatal(err)
	}
	first, err := Issue(client.Id, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Token) < 40 {
		t.Fatalf("unexpected token shape")
	}
	var stored Credential
	if err := db.First(&stored, first.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.TokenHash == first.Token || len(stored.TokenHash) != 64 {
		t.Fatalf("plaintext token persisted")
	}
	second, err := Issue(client.Id, 0)
	if err != nil {
		t.Fatal(err)
	}
	var rotated Credential
	if err := db.First(&rotated, second.ID).Error; err != nil {
		t.Fatal(err)
	}
	if second.Token == first.Token || rotated.Version <= stored.Version {
		t.Fatalf("rotation did not replace token/version")
	}
	if err := Revoke(client.Id); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&stored, first.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Enabled || stored.Version <= rotated.Version {
		t.Fatalf("revoke did not disable and increment version")
	}
}

func TestGrantHostRequiresKnownSubject(t *testing.T) {
	db := testDB(t)
	client := model.ClientRecord{Email: "portal@example.test", SubID: "sub-1"}
	host := model.Host{InboundId: 1, Address: "edge.example.test", Port: 443}
	if err := db.Create(&client).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&host).Error; err != nil {
		t.Fatal(err)
	}
	grant, err := GrantHost("client", client.Id, host.Id)
	if err != nil {
		t.Fatal(err)
	}
	if grant.HostID != host.Id {
		t.Fatalf("grant host id = %d, want %d", grant.HostID, host.Id)
	}
	if _, err := GrantHost("client", 999, host.Id); err == nil || strings.Contains(err.Error(), "token") {
		t.Fatalf("unknown subject accepted or sensitive error: %v", err)
	}
}
