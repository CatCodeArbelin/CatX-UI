package controller

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCatXCompositionDoesNotExposeLegacySponsorRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewIndexController(router.Group("/"))

	for _, route := range router.Routes() {
		if route.Path == "/sponsors" || route.Path == "/sponsors/logo/:name" {
			t.Fatalf("legacy upstream sponsor route is still reachable: %s %s", route.Method, route.Path)
		}
	}
}
