package portal

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/audit"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
	"gorm.io/gorm"
)

func RegisterAdminRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	g := api.Group("/portal")
	g.GET("/settings", adminSettings)
	g.POST("/settings", adminSetSettings)
	g.GET("/credentials", adminCredentials)
	g.POST("/credentials", adminIssue)
	g.POST("/credentials/:clientId/rotate", adminRotate)
	g.POST("/credentials/:clientId/revoke", adminRevoke)
	g.GET("/host-grants", adminGrants)
	g.POST("/host-grants", adminGrant)
	g.DELETE("/host-grants/:id", adminDeleteGrant)
}

func RegisterPortalRoutes(g *gin.RouterGroup, secret []byte, basePath string, secure bool) {
	if g == nil {
		return
	}
	portal := g.Group("/portal")
	store := cookie.NewStore(secret)
	store.Options(sessions.Options{Path: basePath, HttpOnly: true, Secure: secure, SameSite: http.SameSiteStrictMode, MaxAge: 8 * 60 * 60})
	portal.Use(sessions.Sessions(portalCookie, store))
	portal.POST("/auth", authenticate)
	portal.GET("/csrf", func(c *gin.Context) {
		s := portalSession(c)
		token, _ := s.Get(portalCSRFKey).(string)
		if token == "" {
			token, _ = randomToken()
			token = token[:32]
			s.Set(portalCSRFKey, token)
			_ = s.Save()
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "obj": token})
	})
	portal.Use(PortalMiddleware())
	portal.Use(csrfMiddleware())
	portal.GET("/me", me)
	portal.GET("/devices", devices)
	portal.PUT("/devices/:id", renameDevice)
	portal.DELETE("/devices/:id", revokeDevice)
	portal.GET("/hosts", hosts)
	portal.GET("/traffic", traffic)
	portal.POST("/access/rotate", rotateSelf)
	portal.POST("/access/revoke", revokeSelf)
	portal.POST("/logout", logout)
}

func adminSettings(c *gin.Context) {
	_, on := db()
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": PortalSettings{Enabled: on}})
}
func adminSetSettings(c *gin.Context) {
	var req PortalSettings
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid settings"})
		return
	}
	d, _ := db()
	if d == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "database unavailable"})
		return
	}
	if err := setFlag(d, req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "settings save failed"})
		return
	}
	Configure(d, req.Enabled)
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": PortalSettings{Enabled: req.Enabled}})
}
func setFlag(d *gorm.DB, enabled bool) error {
	var row model.Setting
	err := d.Where("key = ?", "fork.self_service.enabled").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return d.Create(&model.Setting{Key: "fork.self_service.enabled", Value: strconv.FormatBool(enabled)}).Error
	}
	if err != nil {
		return err
	}
	row.Value = strconv.FormatBool(enabled)
	return d.Save(&row).Error
}

func adminCredentials(c *gin.Context) {
	rows, err := ListCredentials()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": rows})
}
func parseClientID(c *gin.Context) (int, error) { return strconv.Atoi(c.Param("clientId")) }
func adminIssue(c *gin.Context) {
	var req struct {
		ClientID  int   `json:"clientId"`
		ExpiresAt int64 `json:"expiresAt"`
	}
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid credential request"})
		return
	}
	row, err := Issue(req.ClientID, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": row})
}
func adminRotate(c *gin.Context) {
	id, err := parseClientID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid clientId"})
		return
	}
	var req struct {
		ExpiresAt int64 `json:"expiresAt"`
	}
	_ = c.ShouldBindJSON(&req)
	row, err := Issue(id, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": row})
}
func adminRevoke(c *gin.Context) {
	id, err := parseClientID(c)
	if err != nil || Revoke(id) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "credential not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
func adminGrants(c *gin.Context) {
	rows, err := ListGrants()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": rows})
}
func adminGrant(c *gin.Context) {
	var req HostGrant
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid host grant"})
		return
	}
	row, err := GrantHost(req.SubjectType, req.SubjectID, req.HostID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": row})
}
func adminDeleteGrant(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || DeleteGrant(uint(id)) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "grant not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func portalClient(c *gin.Context) *model.ClientRecord {
	v, _ := c.Get("portal_client")
	client, _ := v.(*model.ClientRecord)
	return client
}
func me(c *gin.Context) {
	client := portalClient(c)
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": clientView(client)})
}
func clientView(client *model.ClientRecord) gin.H {
	return gin.H{"id": client.Id, "email": client.Email, "group": client.Group, "enabled": client.Enable, "totalGB": client.TotalGB, "expiryTime": client.ExpiryTime, "limitHwid": client.LimitHwid}
}

func devices(c *gin.Context) {
	client := portalClient(c)
	db, _ := db()
	var rows []model.ClientHwid
	if err := db.Where("sub_id = ?", client.SubID).Order("last_seen desc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "devices unavailable"})
		return
	}
	type view struct {
		ID          int    `json:"id"`
		FirstSeen   int64  `json:"firstSeen"`
		LastSeen    int64  `json:"lastSeen"`
		DeviceName  string `json:"deviceName"`
		DeviceOS    string `json:"deviceOs"`
		OsVersion   string `json:"osVersion"`
		DeviceModel string `json:"deviceModel"`
	}
	out := make([]view, 0, len(rows))
	for _, r := range rows {
		out = append(out, view{r.Id, r.FirstSeen, r.LastSeen, r.DeviceName, r.DeviceOS, r.OsVersion, r.DeviceModel})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": out})
}
func renameDevice(c *gin.Context) {
	client := portalClient(c)
	db, _ := db()
	id, err := strconv.Atoi(c.Param("id"))
	var req struct {
		DeviceName string `json:"deviceName"`
	}
	if err != nil || c.ShouldBindJSON(&req) != nil || len([]rune(req.DeviceName)) > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid device name"})
		return
	}
	res := db.Model(&model.ClientHwid{}).Where("id = ? AND sub_id = ?", id, client.SubID).Update("device_name", strings.TrimSpace(req.DeviceName))
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "device update failed"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "device not found"})
		return
	}
	AuditPortalMutation(c, "portal.device.rename", client)
	c.JSON(http.StatusOK, gin.H{"success": true})
}
func revokeDevice(c *gin.Context) {
	client := portalClient(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": "invalid device"})
		return
	}
	if revokeDeviceOwned == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "msg": "device service unavailable"})
		return
	}
	if err := revokeDeviceOwned(client.Email, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "device not found") {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "device not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "device revoke failed"})
		return
	}
	AuditPortalMutation(c, "portal.device.revoke", client)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

type hostView struct {
	ID          int      `json:"id"`
	InboundID   int      `json:"inboundId"`
	Remark      string   `json:"remark"`
	Address     string   `json:"address"`
	Port        int      `json:"port"`
	Security    string   `json:"security"`
	SNI         string   `json:"sni,omitempty"`
	HostHeader  string   `json:"hostHeader,omitempty"`
	Path        string   `json:"path,omitempty"`
	ALPN        []string `json:"alpn,omitempty"`
	Fingerprint string   `json:"fingerprint,omitempty"`
}

func hosts(c *gin.Context) {
	client := portalClient(c)
	db, _ := db()
	var inboundIDs []int
	db.Table("client_inbounds").Where("client_id = ?", client.Id).Pluck("inbound_id", &inboundIDs)
	if len(inboundIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "obj": []hostView{}})
		return
	}
	var rows []model.Host
	q := db.Where("inbound_id IN ? AND is_disabled = ? AND is_hidden = ?", inboundIDs, false, false)
	if client.Group != "" {
		var group model.ClientGroup
		if db.Where("name = ?", client.Group).First(&group).Error == nil {
			var ids []int
			db.Model(&HostGrant{}).Where("enabled = ? AND ((subject_type = ? AND subject_id = ?) OR (subject_type = ? AND subject_id = ?))", true, "client", client.Id, "group", group.Id).Pluck("host_id", &ids)
			if len(ids) > 0 {
				q = q.Or("id IN ?", ids)
			}
		}
	}
	if err := q.Order("sort_order asc, id asc").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "hosts unavailable"})
		return
	}
	out := make([]hostView, 0, len(rows))
	for _, h := range rows {
		out = append(out, hostView{h.Id, h.InboundId, h.Remark, h.Address, h.Port, h.Security, h.Sni, h.HostHeader, h.Path, h.Alpn, h.Fingerprint})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": out})
}
func traffic(c *gin.Context) {
	client := portalClient(c)
	db, _ := db()
	var row struct {
		Up    int64
		Down  int64
		Total int64
	}
	err := db.Model(&xray.ClientTraffic{}).Select("COALESCE(SUM(up),0) AS up, COALESCE(SUM(down),0) AS down, COALESCE(MAX(total),0) AS total").Where("email = ?", client.Email).Scan(&row).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "traffic unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{"up": row.Up, "down": row.Down, "total": row.Total, "expiryTime": client.ExpiryTime}})
}
func rotateSelf(c *gin.Context) {
	client := portalClient(c)
	row, err := Issue(client.Id, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "rotation failed"})
		return
	}
	portalSession(c).Clear()
	_ = portalSession(c).Save()
	AuditPortalMutation(c, "portal.access.rotate", client)
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": row})
}
func revokeSelf(c *gin.Context) {
	client := portalClient(c)
	if err := Revoke(client.Id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "revoke failed"})
		return
	}
	portalSession(c).Clear()
	_ = portalSession(c).Save()
	AuditPortalMutation(c, "portal.access.revoke", client)
	c.JSON(http.StatusOK, gin.H{"success": true})
}
func logout(c *gin.Context) {
	s := portalSession(c)
	s.Clear()
	_ = s.Save()
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func AuditPortalMutation(c *gin.Context, event string, client *model.ClientRecord) {
	_ = audit.Record(c.Request.Context(), &audit.AuditEvent{EventType: event, Outcome: audit.EventSuccess, ActorType: "client_portal", ActorID: strconv.Itoa(client.Id), ActorName: sanitizedName(client.Email), AuthMethod: "portal-session", TargetType: "client", TargetRef: strconv.Itoa(client.Id), SourceIP: clientIP(c), Metadata: "{}"})
}
