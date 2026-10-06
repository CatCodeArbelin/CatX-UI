package forkext

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

type featureSettingsEnvelope struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Obj     struct {
		Items           []FeatureFlagInfo `json:"items"`
		RestartRequired bool              `json:"restartRequired"`
	} `json:"obj"`
}

func newFeatureSettingsTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	disableRuntime()
	setSettingsDB(newSettingsTestDB(t))
	t.Cleanup(func() {
		setSettingsDB(nil)
		disableRuntime()
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerFeatureSettingsRoutes(router.Group("/panel/api"))
	return router
}

func decodeFeatureSettings(t *testing.T, recorder *httptest.ResponseRecorder) featureSettingsEnvelope {
	t.Helper()
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want application/json; charset=utf-8", contentType)
	}
	var response featureSettingsEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, recorder.Body.String())
	}
	return response
}

func TestFeatureSettingsRoutesExposeJSONContractAndPersistFlags(t *testing.T) {
	router := newFeatureSettingsTestRouter(t)

	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/panel/api/fork/settings/features", nil))
	if get.Code != http.StatusOK {
		t.Fatalf("GET status = %d; body=%s", get.Code, get.Body.String())
	}
	loaded := decodeFeatureSettings(t, get)
	if !loaded.Success {
		t.Fatalf("GET success=%v msg=%q", loaded.Success, loaded.Msg)
	}
	if loaded.Obj.RestartRequired {
		t.Fatal("GET restartRequired = true, want false for the disabled runtime")
	}
	if len(loaded.Obj.Items) != len(managedFeatureFlags) {
		t.Fatalf("GET item count = %d, want %d", len(loaded.Obj.Items), len(managedFeatureFlags))
	}
	for _, item := range loaded.Obj.Items {
		if item.Enabled {
			t.Errorf("default %q = true, want false", item.Key)
		}
		if item.Active || item.State != RuntimeStateFeatureOff || item.RestartRequired {
			t.Errorf("default %q runtime metadata = active:%v state:%q restart:%v, want off", item.Key, item.Active, item.State, item.RestartRequired)
		}
		if item.Key == FlagWebhooks || item.Key == FlagMetrics {
			t.Errorf("reserved flag %q leaked into managed settings", item.Key)
		}
	}

	payload, err := json.Marshal(map[string]map[string]bool{
		"flags": {
			string(FlagAnalytics):       true,
			string(FlagDNSIntelligence): true,
		},
	})
	if err != nil {
		t.Fatalf("marshal PUT payload: %v", err)
	}
	put := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/panel/api/fork/settings/features", bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(put, request)
	if put.Code != http.StatusOK {
		t.Fatalf("PUT status = %d; body=%s", put.Code, put.Body.String())
	}
	saved := decodeFeatureSettings(t, put)
	if !saved.Success || !saved.Obj.RestartRequired {
		t.Fatalf("PUT envelope = %+v, want success with restart metadata", saved)
	}
	for _, item := range saved.Obj.Items {
		if item.Key == FlagAnalytics || item.Key == FlagDNSIntelligence {
			if item.Active || item.State != RuntimeStateRestartRequired || !item.RestartRequired {
				t.Errorf("saved %q runtime metadata = active:%v state:%q restart:%v, want restart-required", item.Key, item.Active, item.State, item.RestartRequired)
			}
		}
	}
	if savedItems := boolFromItems(saved.Obj.Items); !savedItems[FlagAnalytics] || !savedItems[FlagDNSIntelligence] {
		t.Fatalf("PUT response omitted enabled dependency state: %+v", savedItems)
	}
}

func boolFromItems(items []FeatureFlagInfo) map[Flag]bool {
	values := make(map[Flag]bool, len(items))
	for _, item := range items {
		values[item.Key] = item.Enabled
	}
	return values
}

func TestFeatureSettingsRoutesRejectMalformedJSONAndInvalidDependenciesAtomically(t *testing.T) {
	router := newFeatureSettingsTestRouter(t)

	malformed := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPut, "/panel/api/fork/settings/features", bytes.NewBufferString(`{"flags":`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(malformed, request)
	if malformed.Code != http.StatusBadRequest {
		t.Fatalf("malformed JSON PUT status = %d; body=%s", malformed.Code, malformed.Body.String())
	}

	invalid := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodPut, "/panel/api/fork/settings/features", bytes.NewBufferString(`{"flags":{"dns_intelligence.enabled":true}}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(invalid, request)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid dependency PUT status = %d; body=%s", invalid.Code, invalid.Body.String())
	}
	response := decodeFeatureSettings(t, invalid)
	if response.Success {
		t.Fatal("invalid dependency PUT returned success")
	}

	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/panel/api/fork/settings/features", nil))
	if get.Code != http.StatusOK {
		t.Fatalf("GET after rejected PUT status = %d; body=%s", get.Code, get.Body.String())
	}
	values := boolFromItems(decodeFeatureSettings(t, get).Obj.Items)
	if values[FlagAnalytics] || values[FlagDNSIntelligence] {
		t.Fatalf("rejected dependency update changed persisted state: %+v", values)
	}
}

func TestFeatureSettingsExposeMalformedPersistedFlagAsRuntimeError(t *testing.T) {
	router := newFeatureSettingsTestRouter(t)
	db := currentSettingsDB()
	if err := db.Create(&model.Setting{Key: settingKey(FlagAnalytics), Value: "not-a-boolean"}).Error; err != nil {
		t.Fatalf("create malformed feature setting: %v", err)
	}

	get := httptest.NewRecorder()
	router.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/panel/api/fork/settings/features", nil))
	if get.Code != http.StatusOK {
		t.Fatalf("GET status = %d; body=%s", get.Code, get.Body.String())
	}
	response := decodeFeatureSettings(t, get)
	if !response.Success {
		t.Fatalf("GET success=%v msg=%q", response.Success, response.Msg)
	}
	for _, item := range response.Obj.Items {
		if item.Key == FlagAnalytics {
			if item.Enabled || item.Active || item.State != RuntimeStateError || item.RestartRequired {
				t.Fatalf("malformed analytics runtime metadata = %+v, want inactive/error", item)
			}
			return
		}
	}
	t.Fatalf("malformed analytics flag missing from response: %+v", response.Obj.Items)
}

func TestFeatureSettingsCanRepairMalformedPersistedFlag(t *testing.T) {
	router := newFeatureSettingsTestRouter(t)
	db := currentSettingsDB()
	if err := db.Create(&model.Setting{Key: settingKey(FlagAnalytics), Value: "not-a-boolean"}).Error; err != nil {
		t.Fatalf("create malformed feature setting: %v", err)
	}

	request := httptest.NewRequest(http.MethodPut, "/panel/api/fork/settings/features", bytes.NewBufferString(`{"flags":{"analytics.enabled":false}}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("repair PUT status = %d; body=%s", response.Code, response.Body.String())
	}
	if saved := decodeFeatureSettings(t, response); !saved.Success {
		t.Fatalf("repair PUT response = %+v", saved)
	}
	value, err := NewSettings(db).Enabled(FlagAnalytics)
	if err != nil || value {
		t.Fatalf("repaired analytics flag = %v, err=%v; want false without error", value, err)
	}
}
