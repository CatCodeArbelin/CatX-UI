package policysim

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mhsanaei/3x-ui/v3/internal/policy"
)

func TestDisabledSimulationExposesTypedFeatureState(t *testing.T) {
	policy.Configure(nil, false)
	t.Cleanup(func() { policy.Configure(nil, false) })
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/panel/api"))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panel/api/policies/simulate?clientEmail=alice", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"feature_off"`) || !strings.Contains(response.Body.String(), `"featureDisabled":true`) {
		t.Fatalf("simulation response = %d %s", response.Code, response.Body.String())
	}
}

func TestRouteMatchesDomainAndCIDR(t *testing.T) {
	if !routeMatches("example.test", Input{Domain: "sub.example.test"}) {
		t.Fatal("subdomain should match domain target")
	}
	if !routeMatches("ip:10.0.0.0/8", Input{IP: "10.2.3.4"}) {
		t.Fatal("IP should match CIDR target")
	}
	if routeMatches("ip:10.0.0.0/8", Input{IP: "192.0.2.1"}) {
		t.Fatal("unrelated IP matched CIDR")
	}
}
