package policy

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPolicyRuntimeStateLifecycle(t *testing.T) {
	SetRuntimeChangeCallback(nil)
	MarkRuntimeApplied()
	t.Cleanup(func() {
		SetRuntimeChangeCallback(nil)
		MarkRuntimeApplied()
	})
	if status := RuntimeStatus(); status.State != RuntimeStateActive || status.RestartRequired {
		t.Fatalf("initial policy runtime status = %+v", status)
	}
	MarkRuntimeApplyRequired()
	if status := RuntimeStatus(); status.State != RuntimeStateRestartRequired || !status.RestartRequired {
		t.Fatalf("pending policy runtime status = %+v", status)
	}
	MarkRuntimeApplyError(errors.New("candidate rejected"))
	if status := RuntimeStatus(); status.State != RuntimeStateError || status.LastError != "candidate rejected" {
		t.Fatalf("failed policy runtime status = %+v", status)
	}
	MarkRuntimeApplied()
	if status := RuntimeStatus(); status.State != RuntimeStateActive || status.RestartRequired || status.LastError != "" {
		t.Fatalf("applied policy runtime status = %+v", status)
	}
}

func TestDisabledPolicyRoutesExposeTypedFeatureState(t *testing.T) {
	Configure(nil, false)
	t.Cleanup(func() { Configure(nil, false) })
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/panel/api"))

	for _, path := range []string{"/panel/api/policies/status", "/panel/api/policies"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"feature_off"`) || !strings.Contains(response.Body.String(), `"featureDisabled":true`) {
			t.Fatalf("%s response = %d %s", path, response.Code, response.Body.String())
		}
	}
}

func TestDisabledPolicyWithPendingXrayApplyIsNotReportedAsFeatureOff(t *testing.T) {
	Configure(nil, false)
	MarkRuntimeApplyRequired()
	t.Cleanup(func() {
		Configure(nil, false)
	})
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/panel/api"))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panel/api/policies/status", nil))
	body := response.Body.String()
	if response.Code != http.StatusOK || !strings.Contains(body, `"state":"restart_required"`) || strings.Contains(body, `"featureDisabled":true`) {
		t.Fatalf("pending disabled policy response = %d %s", response.Code, body)
	}
}
