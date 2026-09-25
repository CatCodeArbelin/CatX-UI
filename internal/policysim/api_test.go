package policysim

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v3/internal/policy"
)

func TestDisabledSimulationIsReadOnlyNoOp(t *testing.T) {
	policy.Configure(nil, false)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/panel/api"))
	req := httptest.NewRequest(http.MethodGet, "/panel/api/policies/simulate?clientEmail=alice%40example.test&domain=example.test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK || res.Body.String() == "" {
		t.Fatalf("disabled simulation response: %d %s", res.Code, res.Body.String())
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
