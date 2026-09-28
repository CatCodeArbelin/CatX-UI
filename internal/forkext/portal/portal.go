package portal

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/audit"
)

const (
	credentialVersionKey = "portal_credential_version"
	portalClientIDKey    = "portal_client_id"
	portalCSRFKey        = "portal_csrf"
	portalCookie         = "catx-portal"
	maxPageSize          = 100
)

var (
	configured        atomicConfig
	loginMu           sync.Mutex
	loginHits         = map[string]rateWindow{}
	revokeDeviceOwned func(string, int) error
)

type atomicConfig struct {
	mu sync.RWMutex
	db *gorm.DB
	on bool
}

type rateWindow struct {
	started time.Time
	count   int
}

func Configure(db *gorm.DB, on bool) {
	configured.mu.Lock()
	configured.db, configured.on = db, on && db != nil
	configured.mu.Unlock()
}

func db() (*gorm.DB, bool) {
	configured.mu.RLock()
	defer configured.mu.RUnlock()
	return configured.db, configured.on
}
func Enabled() bool { _, on := db(); return on }

// SetDeviceRevoker installs the upstream ownership-safe device deletion
// service without importing internal/web/service into this fork package.
func SetDeviceRevoker(fn func(email string, deviceID int) error) { revokeDeviceOwned = fn }

func Migrate(database *gorm.DB) error { return database.AutoMigrate(&Credential{}, &HostGrant{}) }

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func nowMillis() int64 { return time.Now().UnixMilli() }

func credentialView(row Credential) CredentialView {
	return CredentialView{ID: row.ID, ClientID: row.ClientID, Enabled: row.Enabled, ExpiresAt: row.ExpiresAt, LastUsed: row.LastUsed, CreatedAt: row.CreatedAt.UnixMilli()}
}

func clientExists(tx *gorm.DB, id int) (bool, error) {
	var n int64
	err := tx.Model(&model.ClientRecord{}).Where("id = ?", id).Count(&n).Error
	return n == 1, err
}

func Issue(clientID int, expiresAt int64) (IssuedToken, error) {
	db, on := db()
	if !on {
		return IssuedToken{}, errors.New("self-service is disabled")
	}
	if clientID <= 0 {
		return IssuedToken{}, errors.New("clientId is required")
	}
	ok, err := clientExists(db, clientID)
	if err != nil {
		return IssuedToken{}, err
	}
	if !ok {
		return IssuedToken{}, gorm.ErrRecordNotFound
	}
	token, err := randomToken()
	if err != nil {
		return IssuedToken{}, err
	}
	row := Credential{ClientID: clientID, TokenHash: hashToken(token), Version: 1, Enabled: true, ExpiresAt: expiresAt}
	var existing Credential
	if err := db.Where("client_id = ?", clientID).First(&existing).Error; err == nil {
		row.ID, row.Version, row.CreatedAt = existing.ID, existing.Version+1, existing.CreatedAt
		if err := db.Model(&Credential{}).Where("id = ?", existing.ID).Updates(map[string]any{"token_hash": row.TokenHash, "version": row.Version, "enabled": true, "expires_at": expiresAt, "last_used": 0}).Error; err != nil {
			return IssuedToken{}, err
		}
		if err := db.First(&row, existing.ID).Error; err != nil {
			return IssuedToken{}, err
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Create(&row).Error; err != nil {
			return IssuedToken{}, err
		}
	} else {
		return IssuedToken{}, err
	}
	return IssuedToken{CredentialView: credentialView(row), Token: token}, nil
}

func Revoke(clientID int) error {
	db, on := db()
	if !on {
		return errors.New("self-service is disabled")
	}
	result := db.Model(&Credential{}).Where("client_id = ?", clientID).Updates(map[string]any{"enabled": false, "version": gorm.Expr("version + 1")})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func ListCredentials() ([]CredentialView, error) {
	db, on := db()
	if !on {
		return nil, errors.New("self-service is disabled")
	}
	var rows []Credential
	if err := db.Order("client_id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]CredentialView, 0, len(rows))
	for _, row := range rows {
		out = append(out, credentialView(row))
	}
	return out, nil
}

func grantSubjectExists(tx *gorm.DB, subjectType string, subjectID int) error {
	switch subjectType {
	case "client":
		ok, err := clientExists(tx, subjectID)
		if err != nil {
			return err
		}
		if !ok {
			return gorm.ErrRecordNotFound
		}
	case "group":
		var n int64
		if err := tx.Model(&model.ClientGroup{}).Where("id = ?", subjectID).Count(&n).Error; err != nil {
			return err
		}
		if n != 1 {
			return gorm.ErrRecordNotFound
		}
	default:
		return errors.New("subjectType must be client or group")
	}
	return nil
}

func GrantHost(subjectType string, subjectID, hostID int) (HostGrant, error) {
	db, on := db()
	if !on {
		return HostGrant{}, errors.New("self-service is disabled")
	}
	if err := grantSubjectExists(db, subjectType, subjectID); err != nil {
		return HostGrant{}, err
	}
	var n int64
	if err := db.Model(&model.Host{}).Where("id = ?", hostID).Count(&n).Error; err != nil {
		return HostGrant{}, err
	}
	if n != 1 {
		return HostGrant{}, gorm.ErrRecordNotFound
	}
	row := HostGrant{SubjectType: subjectType, SubjectID: subjectID, HostID: hostID, Enabled: true}
	err := db.Where("subject_type = ? AND subject_id = ? AND host_id = ?", subjectType, subjectID, hostID).FirstOrCreate(&row).Error
	return row, err
}

func DeleteGrant(id uint) error {
	db, on := db()
	if !on {
		return errors.New("self-service is disabled")
	}
	return db.Delete(&HostGrant{}, id).Error
}

func ListGrants() ([]HostGrant, error) {
	db, on := db()
	if !on {
		return nil, errors.New("self-service is disabled")
	}
	var rows []HostGrant
	err := db.Order("id asc").Find(&rows).Error
	return rows, err
}

func allowLogin(key string) bool {
	loginMu.Lock()
	defer loginMu.Unlock()
	now := time.Now()
	if len(loginHits) > 10000 {
		for k, v := range loginHits {
			if now.Sub(v.started) >= time.Minute {
				delete(loginHits, k)
			}
		}
	}
	w := loginHits[key]
	if now.Sub(w.started) >= time.Minute {
		w = rateWindow{started: now}
	}
	if w.count >= 10 {
		loginHits[key] = w
		return false
	}
	w.count++
	loginHits[key] = w
	return true
}

func SetPortalActor(c *gin.Context, client *model.ClientRecord) {
	c.Set("catx_portal_actor", true)
	c.Set("catx_actor_id", stringID(client.Id))
	c.Set("catx_actor_name", sanitizedName(client.Email))
	c.Set("catx_auth_method", "portal-token")
}
func stringID(id int) string { return strconv.Itoa(id) }
func sanitizedName(v string) string {
	v = strings.TrimSpace(v)
	if len(v) > 120 {
		return v[:120]
	}
	return v
}

func portalSession(c *gin.Context) sessions.Session { return sessions.Default(c) }

func PortalMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, on := db()
		if !on {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		s := portalSession(c)
		rawID, rawVersion := s.Get(portalClientIDKey), s.Get(credentialVersionKey)
		clientID, ok1 := sessionInt(rawID)
		version, ok2 := sessionInt64(rawVersion)
		if !ok1 || !ok2 || clientID <= 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		db, _ := db()
		var cred Credential
		if err := db.Where("client_id = ?", clientID).First(&cred).Error; err != nil || !cred.Enabled || cred.Version != version || (cred.ExpiresAt > 0 && cred.ExpiresAt < nowMillis()) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		var client model.ClientRecord
		if err := db.First(&client, clientID).Error; err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("portal_client", &client)
		SetPortalActor(c, &client)
		c.Next()
	}
}

func sessionInt(v any) (int, bool) { n, ok := sessionInt64(v); return int(n), ok && n > 0 }
func sessionInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
	case int32:
		return int64(n), true
	case float64:
		return int64(n), n == float64(int64(n))
	default:
		return 0, false
	}
}

func csrfMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		if !sameOrigin(c, scheme) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if c.GetHeader("X-CSRF-Token") == "" || c.GetHeader("X-CSRF-Token") != portalSession(c).Get(portalCSRFKey) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

func sameOrigin(c *gin.Context, scheme string) bool {
	origin := c.GetHeader("Origin")
	return origin == "" || origin == scheme+"://"+c.Request.Host
}

func authenticate(c *gin.Context) {
	db, on := db()
	if !on {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "msg": "portal disabled"})
		return
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if !sameOrigin(c, scheme) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	if !allowLogin(clientIP(c)) {
		c.JSON(http.StatusTooManyRequests, gin.H{"success": false, "msg": "too many attempts"})
		return
	}
	var req struct {
		Token string `json:"token"`
	}
	if c.ShouldBindJSON(&req) != nil || len(req.Token) < 40 || len(req.Token) > 64 {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "msg": "invalid portal token"})
		return
	}
	var cred Credential
	if err := db.Where("token_hash = ? AND enabled = ?", hashToken(req.Token), true).First(&cred).Error; err != nil || (cred.ExpiresAt > 0 && cred.ExpiresAt < nowMillis()) {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "msg": "invalid portal token"})
		return
	}
	var client model.ClientRecord
	if err := db.First(&client, cred.ClientID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "msg": "client unavailable"})
		return
	}
	s := portalSession(c)
	s.Clear()
	s.Set(portalClientIDKey, client.Id)
	s.Set(credentialVersionKey, cred.Version)
	csrf, _ := randomToken()
	s.Set(portalCSRFKey, csrf[:32])
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "msg": "session unavailable"})
		return
	}
	db.Model(&Credential{}).Where("id = ?", cred.ID).Update("last_used", nowMillis())
	_ = audit.Record(c.Request.Context(), &audit.AuditEvent{EventType: "portal.auth.login", Outcome: audit.EventSuccess, ActorType: "client_portal", ActorID: stringID(client.Id), ActorName: sanitizedName(client.Email), AuthMethod: "portal-token", TargetType: "client", TargetRef: stringID(client.Id), SourceIP: clientIP(c), Metadata: "{}"})
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{"csrfToken": csrf[:32]}})
}

func clientIP(c *gin.Context) string {
	host := c.Request.RemoteAddr
	if i := strings.LastIndex(host, ":"); i > 0 {
		return host[:i]
	}
	return host
}
