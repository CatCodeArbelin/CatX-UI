package policysim

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v3/internal/policy"
	"github.com/mhsanaei/3x-ui/v3/internal/policycompiler"
)

type Input struct {
	ClientEmail string `json:"clientEmail" form:"clientEmail"`
	GroupName   string `json:"groupName,omitempty" form:"groupName"`
	Domain      string `json:"domain,omitempty" form:"domain"`
	IP          string `json:"ip,omitempty" form:"ip"`
	Category    string `json:"category,omitempty" form:"category"`
	Service     string `json:"service,omitempty" form:"service"`
	At          int64  `json:"at" form:"at"`
}

type Result struct {
	Enabled    bool                         `json:"enabled"`
	Input      Input                        `json:"input"`
	Decision   *policy.Decision             `json:"decision,omitempty"`
	Route      []policycompiler.RulePreview `json:"route"`
	NoOpReason string                       `json:"noOpReason,omitempty"`
}

func RegisterRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	api.GET("/policies/simulate", simulate)
	api.GET("/policies/explain", simulate)
	api.GET("/policies/explain-route", simulate)
}

func parseInput(c *gin.Context) (Input, error) {
	in := Input{ClientEmail: strings.TrimSpace(c.Query("clientEmail")), GroupName: strings.TrimSpace(c.Query("groupName")), Domain: strings.ToLower(strings.TrimSpace(c.Query("domain"))), IP: strings.TrimSpace(c.Query("ip")), Category: strings.TrimSpace(c.Query("category")), Service: strings.TrimSpace(c.Query("service"))}
	if in.ClientEmail == "" {
		return in, fmt.Errorf("clientEmail is required")
	}
	in.At = time.Now().UnixMilli()
	if raw := c.Query("at"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return in, fmt.Errorf("invalid at")
		}
		in.At = parsed
	}
	return in, nil
}

func simulate(c *gin.Context) {
	in, err := parseInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}
	if !policy.Enabled() {
		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "", "obj": Result{Enabled: false, Input: in, Route: []policycompiler.RulePreview{}, NoOpReason: "policies are disabled"}})
		return
	}
	decision, err := policy.ResolveCurrentDecision(c.Request.Context(), in.ClientEmail, in.GroupName, in.At)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "policy simulation unavailable"})
		return
	}
	result := Result{Enabled: true, Input: in, Decision: decision, Route: []policycompiler.RulePreview{}}
	if decision == nil {
		result.NoOpReason = "no effective policy"
		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "", "obj": result})
		return
	}
	preview, err := policycompiler.Preview([]policy.Decision{*decision})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": err.Error()})
		return
	}
	for i := range preview {
		preview[i].Matched = routeMatches(preview[i].Destination, in)
	}
	result.Route = preview
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "", "obj": result})
}

func routeMatches(destination string, in Input) bool {
	if destination == "" {
		return false
	}
	if strings.HasPrefix(destination, "ip:") {
		target := strings.TrimPrefix(destination, "ip:")
		if in.IP == "" {
			return false
		}
		if strings.Contains(target, "/") {
			_, network, err := net.ParseCIDR(target)
			parsed := net.ParseIP(in.IP)
			return err == nil && parsed != nil && network.Contains(parsed)
		}
		return net.ParseIP(target) != nil && target == in.IP
	}
	if in.Domain == "" {
		return false
	}
	domain := strings.TrimPrefix(destination, "domain:")
	return in.Domain == domain || strings.HasSuffix(in.Domain, "."+domain)
}
