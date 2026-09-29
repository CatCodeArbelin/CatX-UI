package fleetupdate

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	g := api.Group("/fleet-updates")
	g.POST("/campaigns", create)
	g.GET("/campaigns", list)
	g.GET("/campaigns/:id", get)
	g.POST("/campaigns/:id/reconcile", reconcile)
	g.POST("/campaigns/:id/abort", abort)
	g.POST("/campaigns/:id/retry", retry)
}

func svc(c *gin.Context) (*Service, bool) {
	s := Current()
	if s == nil || !s.Enabled() {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "fleet updates disabled", "featureDisabled": true})
		return nil, false
	}
	return s, true
}

func id(c *gin.Context) (uint, error) {
	v, e := strconv.ParseUint(c.Param("id"), 10, 32)
	return uint(v), e
}

func create(c *gin.Context) {
	s, ok := svc(c)
	if !ok {
		return
	}
	var r PlanRequest
	if c.ShouldBindJSON(&r) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid campaign"})
		return
	}
	p, e := s.Plan(c.Request.Context(), r)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": e.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": p})
}

func list(c *gin.Context) {
	s, ok := svc(c)
	if !ok {
		return
	}
	v, e := s.List(c.Request.Context())
	if e != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": e.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": v})
}

func get(c *gin.Context) {
	s, ok := svc(c)
	if !ok {
		return
	}
	n, e := id(c)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid campaign id"})
		return
	}
	v, e := s.Get(c.Request.Context(), n)
	if e != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "campaign not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": v})
}

func reconcile(c *gin.Context) {
	s, ok := svc(c)
	if !ok {
		return
	}
	n, e := id(c)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid campaign id"})
		return
	}
	if e = s.Reconcile(c.Request.Context(), n, "api"); e != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "msg": e.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func abort(c *gin.Context) {
	s, ok := svc(c)
	if !ok {
		return
	}
	n, e := id(c)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid campaign id"})
		return
	}
	if e = s.Abort(c.Request.Context(), n); e != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "msg": e.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func retry(c *gin.Context) {
	s, ok := svc(c)
	if !ok {
		return
	}
	n, e := id(c)
	if e != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid campaign id"})
		return
	}
	if e = s.Retry(c.Request.Context(), n); e != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "msg": e.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
