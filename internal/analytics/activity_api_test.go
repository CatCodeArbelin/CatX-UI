package analytics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestActivityRoutesExposeReadOnlyPagedMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &captureRepo{events: []MetadataEvent{{ObservedAt: 1000, ClientEmail: "alice", Domain: "example.com", Port: 443, Protocol: "tcp", Source: SourceAccessLog, Provenance: ProvenanceObserved, Confidence: 1}}}
	Configure(repo, true)
	t.Cleanup(func() { Configure(NoopRepository{}, false) })
	router := gin.New()
	RegisterActivityRoutes(router.Group("/panel/api"))
	req := httptest.NewRequest(http.MethodGet, "/panel/api/analytics/clients/alice/activity?from=1&to=2000&page=1&pageSize=1", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", resp.Code, resp.Body.String())
	}
	if got := resp.Body.String(); !strings.Contains(got, "example.com") || !strings.Contains(got, `"total":1`) {
		t.Fatalf("unexpected response: %s", got)
	}
}

func TestActivityRoutesAreExactNoOpWhenDisabled(t *testing.T) {
	Configure(NoopRepository{}, false)
	router := gin.New()
	RegisterActivityRoutes(router.Group("/panel/api"))
	req := httptest.NewRequest(http.MethodGet, "/panel/api/analytics/clients/alice/activity", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), `"enabled":false`) {
		t.Fatalf("disabled response = %d %s", resp.Code, resp.Body.String())
	}
}

func TestDNSActivityRouteExposesMetadataOnlyPage(t *testing.T) {
	repo := &captureRepo{dns: []DNSObservation{{ID: 4, Domain: "example.com", ResolvedIP: "192.0.2.4", Source: SourceDNSObserver, Provenance: ProvenanceObserved, Confidence: .9}}}
	Configure(repo, true)
	SetEvidenceEnabled(true)
	t.Cleanup(func() { Configure(NoopRepository{}, false) })
	router := gin.New()
	RegisterActivityRoutes(router.Group("/panel/api"))
	req := httptest.NewRequest(http.MethodGet, "/panel/api/analytics/clients/alice/dns?from=1&to=2000", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), "example.com") || !strings.Contains(resp.Body.String(), "192.0.2.4") {
		t.Fatalf("unexpected DNS response: %d %s", resp.Code, resp.Body.String())
	}
}

func TestTrafficHistoryRouteReturnsBoundedHistoryShape(t *testing.T) {
	Configure(NoopRepository{}, true)
	t.Cleanup(func() { Configure(NoopRepository{}, false) })
	router := gin.New()
	RegisterActivityRoutes(router.Group("/panel/api"))
	req := httptest.NewRequest(http.MethodGet, "/panel/api/analytics/clients/alice/traffic?from=1&to=2000", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), `"serviceBreakdown":[]`) || !strings.Contains(resp.Body.String(), `"categoryBreakdown":[]`) {
		t.Fatalf("unexpected traffic response: %d %s", resp.Code, resp.Body.String())
	}
}

func TestActivityQueryRejectsUnboundedRange(t *testing.T) {
	Configure(NoopRepository{}, false)
	router := gin.New()
	RegisterActivityRoutes(router.Group("/panel/api"))
	req := httptest.NewRequest(http.MethodGet, "/panel/api/analytics/traffic?from=1&to=9999999999999", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("unbounded range status = %d, body = %s", resp.Code, resp.Body.String())
	}
}

var _ Repository = (*captureRepo)(nil)
