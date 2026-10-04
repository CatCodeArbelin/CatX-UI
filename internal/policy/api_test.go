package policy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

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
