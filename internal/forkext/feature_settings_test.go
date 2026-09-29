package forkext

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func TestFeatureSettingsRoutesReadAndWrite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:forkext-feature-settings-route?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate settings: %v", err)
	}
	setSettingsDB(db)
	defer setSettingsDB(nil)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/panel/api")
	registerFeatureSettingsRoutes(api)

	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/panel/api/fork/settings/features", nil))
	if get.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", get.Code)
	}
	var envelope struct {
		Success bool `json:"success"`
		Obj     struct {
			Items []FeatureFlagInfo `json:"items"`
		} `json:"obj"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode GET: %v", err)
	}
	if !envelope.Success || len(envelope.Obj.Items) != 9 {
		t.Fatalf("unexpected GET envelope: %+v", envelope)
	}

	put := httptest.NewRecorder()
	body := `{"flags":{"analytics.enabled":true,"dns_intelligence.enabled":true}}`
	req := httptest.NewRequest(http.MethodPut, "/panel/api/fork/settings/features", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(put, req)
	if put.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, body = %s", put.Code, put.Body.String())
	}
	if enabled, _ := NewSettings(db).Enabled(FlagDNSIntelligence); !enabled {
		t.Fatal("PUT did not persist DNS intelligence")
	}
}
