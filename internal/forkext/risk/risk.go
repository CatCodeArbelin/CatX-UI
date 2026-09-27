package risk

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mhsanaei/3x-ui/v3/internal/analytics"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var state struct {
	sync.RWMutex
	db      *gorm.DB
	enabled bool
}

const (
	retentionIPKey    = "fork.security_anomaly.retention.ip_days"
	retentionEventKey = "fork.security_anomaly.retention.event_days"
)

func Configure(db *gorm.DB, enabled bool) {
	state.Lock()
	state.db, state.enabled = db, enabled && db != nil
	state.Unlock()
}

func CurrentRetention() RetentionSettings {
	result := RetentionSettings{IPDays: 30, EventDays: 90}
	db, ok := enabledDB()
	if !ok {
		return result
	}
	for key, target := range map[string]*int{retentionIPKey: &result.IPDays, retentionEventKey: &result.EventDays} {
		var row model.Setting
		if db.Where("key = ?", key).First(&row).Error != nil {
			continue
		}
		var value int
		if _, err := fmt.Sscan(row.Value, &value); err == nil && value >= 1 && value <= 3650 {
			*target = value
		}
	}
	return result
}

func SaveRetention(input RetentionSettings) error {
	db, ok := enabledDB()
	if !ok {
		return nil
	}
	if input.IPDays < 1 || input.IPDays > 3650 || input.EventDays < 1 || input.EventDays > 3650 {
		return fmt.Errorf("retention must be between 1 and 3650 days")
	}
	for key, value := range map[string]int{retentionIPKey: input.IPDays, retentionEventKey: input.EventDays} {
		var row model.Setting
		err := db.Where("key = ?", key).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&model.Setting{Key: key, Value: fmt.Sprint(value)}).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if err := db.Model(&row).Update("value", fmt.Sprint(value)).Error; err != nil {
			return err
		}
	}
	return nil
}

func enabledDB() (*gorm.DB, bool) {
	state.RLock()
	defer state.RUnlock()
	return state.db, state.enabled
}

func Migrate(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	return db.AutoMigrate(&IPHistory{}, &Event{}, &Score{}, &Suppression{})
}

func hash(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// RecordIPHistory is called by the existing CheckClientIpJob and node
// attribution merge path. It is deliberately not a collector.
func RecordIPHistory(ctx context.Context, nodeGuid string, observations map[string][]model.ClientIpEntry, source string) error {
	db, ok := enabledDB()
	if !ok || len(observations) == 0 {
		return nil
	}
	now := time.Now().UnixMilli()
	emails := make([]string, 0, len(observations))
	for email := range observations {
		if email != "" {
			emails = append(emails, email)
		}
	}
	sort.Strings(emails)
	for _, email := range emails {
		entries := observations[email]
		for _, entry := range entries {
			if entry.IP == "" {
				continue
			}
			observed := entry.Timestamp * 1000
			if observed <= 0 {
				observed = now
			}
			asn, country, provider, confidence, metadataState := lookupMetadata(ctx, entry.IP)
			row := IPHistory{ClientEmail: email, IP: entry.IP, NodeGuid: nodeGuid, ObservedAt: observed, IngestedAt: now, ASN: asn, Country: country, MetadataSource: provider, MetadataConfidence: confidence, MetadataState: metadataState, Source: source, DedupeKey: hash(email, nodeGuid, entry.IP, fmt.Sprintf("%d", observed/60000))}
			if err := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
				return err
			}
		}
		if err := Recompute(ctx, email); err != nil {
			return err
		}
	}
	return nil
}

func lookupMetadata(ctx context.Context, raw string) (uint32, string, string, float64, string) {
	ip := net.ParseIP(strings.TrimSpace(raw))
	if ip == nil {
		return 0, "", "", 0, StateDegraded
	}
	enricher := analytics.CurrentEnricher()
	if enricher == nil {
		return 0, "", "", 0, StateDegraded
	}
	result := enricher.Enrich(ctx, analytics.EnrichmentInput{ObservedAt: time.Now(), DestinationIP: ip.String()})
	if result.ASN == 0 && result.Country == "" {
		return 0, "", result.Source, 0, StateDegraded
	}
	if result.Confidence <= 0 {
		return result.ASN, result.Country, result.Source, result.Confidence, StateDegraded
	}
	return result.ASN, result.Country, result.Source, result.Confidence, StateHealthy
}

type evidence struct {
	Kind       string  `json:"kind"`
	Value      any     `json:"value"`
	Confidence float64 `json:"confidence"`
}

func Recompute(ctx context.Context, email string) error {
	db, ok := enabledDB()
	if !ok || email == "" {
		return nil
	}
	now := time.Now().UnixMilli()
	var rows []IPHistory
	retention := CurrentRetention()
	if err := db.WithContext(ctx).Where("client_email = ? AND observed_at >= ?", email, now-int64(time.Duration(retention.IPDays)*24*time.Hour/time.Millisecond)).Order("observed_at asc, id asc").Find(&rows).Error; err != nil {
		return err
	}
	var sessions int64
	db.WithContext(ctx).Model(&analytics.NetworkSession{}).Where("client_email = ? AND last_seen >= ?", email, now-10*60*1000).Count(&sessions)
	activeCutoff := now - 10*60*1000
	activeIPs, activeNodes, activeCountries, activeASNs := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[uint32]bool{}
	knownMetadata := 0
	countryChanges, asnChanges := 0, 0
	lastCountry, lastASN := "", uint32(0)
	for _, row := range rows {
		if row.MetadataState == StateHealthy {
			knownMetadata++
		}
		if row.Country != "" && lastCountry != "" && row.Country != lastCountry {
			countryChanges++
		}
		if row.ASN != 0 && lastASN != 0 && row.ASN != lastASN {
			asnChanges++
		}
		if row.Country != "" {
			lastCountry = row.Country
		}
		if row.ASN != 0 {
			lastASN = row.ASN
		}
		if row.ObservedAt < activeCutoff || row.ObservedAt > now+5*60*1000 {
			continue
		}
		activeIPs[row.IP] = true
		if row.NodeGuid != "" {
			activeNodes[row.NodeGuid] = true
		}
		if row.Country != "" {
			activeCountries[row.Country] = true
		}
		if row.ASN != 0 {
			activeASNs[row.ASN] = true
		}
	}
	contrib := map[string]int{}
	var ev []evidence
	if len(activeIPs) >= 2 && sessions > 0 {
		contrib[SignalConcurrentIPs] = min(35, 15+len(activeIPs)*5)
		ev = append(ev, evidence{SignalConcurrentIPs, len(activeIPs), .85})
	}
	if len(activeNodes) >= 2 && sessions > 0 {
		contrib[SignalConcurrentNodes] = 15
		ev = append(ev, evidence{SignalConcurrentNodes, len(activeNodes), .8})
	}
	if len(activeCountries) >= 2 && sessions > 0 {
		contrib[SignalCountryDivergence] = 20
		ev = append(ev, evidence{SignalCountryDivergence, len(activeCountries), .75})
	}
	if len(activeASNs) >= 2 && sessions > 0 {
		contrib[SignalASNDivergence] = 15
		ev = append(ev, evidence{SignalASNDivergence, len(activeASNs), .7})
	}
	if countryChanges >= 2 && len(activeCountries) < 2 && sessions > 0 {
		contrib[SignalCountryDivergence] = 10
		ev = append(ev, evidence{SignalCountryDivergence, countryChanges, .6})
	}
	if asnChanges >= 2 && len(activeASNs) < 2 && sessions > 0 {
		contrib[SignalASNDivergence] = 8
		ev = append(ev, evidence{SignalASNDivergence, asnChanges, .55})
	}
	unique24 := map[string]bool{}
	for _, row := range rows {
		if row.ObservedAt >= now-24*60*60*1000 {
			unique24[row.IP] = true
		}
	}
	if len(unique24) >= 3 && sessions >= 2 {
		contrib[SignalIPChurn] = 15
		ev = append(ev, evidence{SignalIPChurn, len(unique24), .65})
	}
	var dns []analytics.DNSObservation
	db.WithContext(ctx).Where("client_email = ? AND observed_at >= ?", email, now-60*60*1000).Find(&dns)
	uniqueDomains := map[string]bool{}
	for _, row := range dns {
		uniqueDomains[row.Domain] = true
	}
	if len(dns) >= 100 || len(uniqueDomains) >= 50 {
		contrib[SignalDNSAnomaly] = 8
		ev = append(ev, evidence{SignalDNSAnomaly, len(uniqueDomains), .45})
	}
	for kind := range contrib {
		if !suppressed(db, email, kind, now) {
			continue
		}
		delete(contrib, kind)
		filtered := ev[:0]
		for _, item := range ev {
			if item.Kind != kind {
				filtered = append(filtered, item)
			}
		}
		ev = filtered
	}

	// A missing/stale provider never becomes a positive location assertion.
	confidence := 0.0
	if len(ev) > 0 {
		for _, item := range ev {
			confidence += item.Confidence
		}
		confidence /= float64(len(ev))
	}
	if len(rows) > 0 && knownMetadata < len(rows) {
		confidence *= .75
	}
	if len(rows) == 0 {
		confidence = 0
	}
	score := 0
	for _, value := range contrib {
		score += value
	}
	if score > 100 {
		score = 100
	}
	stateName := StateHealthy
	if len(rows) == 0 || len(ev) == 0 {
		stateName = StateInsufficient
	}
	if knownMetadata < len(rows) && stateName == StateHealthy {
		stateName = StateDegraded
	}
	if len(rows) > 0 && rows[len(rows)-1].ObservedAt < now-15*60*1000 {
		stateName = StateStale
	}
	if confidence < .5 && stateName == StateHealthy {
		stateName = StateDegraded
	}
	for kind, value := range contrib {
		payload, _ := json.Marshal(ev)
		event := Event{ClientEmail: email, Kind: kind, ObservedAt: now, ScoreContribution: value, Confidence: confidence, State: stateName, Evidence: string(payload), SourceNode: sourceNode(activeNodes), DedupeKey: hash(email, kind, fmt.Sprintf("%d", now/(60*60*1000)))}
		_ = db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&event).Error
	}
	row := Score{ClientEmail: email, Score: score, Confidence: confidence, State: stateName, Evidence: len(ev), CalculatedAt: now, Version: 1}
	return db.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "client_email"}}, DoUpdates: clause.AssignmentColumns([]string{"score", "confidence", "state", "evidence", "calculated_at", "version"})}).Create(&row).Error
}

func sourceNode(nodes map[string]bool) string {
	values := make([]string, 0, len(nodes))
	for node := range nodes {
		values = append(values, node)
	}
	sort.Strings(values)
	if len(values) == 0 {
		return ""
	}
	if len(values) == 1 {
		return values[0]
	}
	return "multiple"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func suppressed(db *gorm.DB, email, kind string, now int64) bool {
	var count int64
	db.Where("client_email = ? AND (kind = ? OR kind = ?) AND (expires_at = 0 OR expires_at > ?)", email, kind, "all", now).Model(&Suppression{}).Count(&count)
	return count > 0
}

func SummaryFor(ctx context.Context, email string) (Summary, error) {
	db, ok := enabledDB()
	if !ok {
		return Summary{Enabled: false, ClientEmail: email, State: StateInsufficient, Events: []Event{}, IPs: []IPHistory{}}, nil
	}
	_ = Recompute(ctx, email)
	result := Summary{Enabled: true, ClientEmail: email, State: StateInsufficient, Events: []Event{}, IPs: []IPHistory{}}
	var score Score
	db.WithContext(ctx).Where("client_email = ?", email).First(&score)
	result.Score, result.Confidence, result.State, result.Evidence, result.CalculatedAt = score.Score, score.Confidence, score.State, score.Evidence, score.CalculatedAt
	db.WithContext(ctx).Where("client_email = ?", email).Order("observed_at desc").Limit(100).Find(&result.Events)
	db.WithContext(ctx).Where("client_email = ?", email).Order("observed_at desc").Limit(200).Find(&result.IPs)
	return result, nil
}

func Prune(ctx context.Context, now time.Time) error {
	db, ok := enabledDB()
	if !ok {
		return nil
	}
	retention := CurrentRetention()
	if err := db.WithContext(ctx).Where("observed_at < ?", now.Add(-time.Duration(retention.IPDays)*24*time.Hour).UnixMilli()).Delete(&IPHistory{}).Error; err != nil {
		return err
	}
	return db.WithContext(ctx).Where("observed_at < ?", now.Add(-time.Duration(retention.EventDays)*24*time.Hour).UnixMilli()).Delete(&Event{}).Error
}

func RecomputeAll(ctx context.Context) error {
	db, ok := enabledDB()
	if !ok {
		return nil
	}
	var emails []string
	if err := db.WithContext(ctx).Model(&IPHistory{}).Distinct("client_email").Pluck("client_email", &emails).Error; err != nil {
		return err
	}
	for _, email := range emails {
		if err := Recompute(ctx, email); err != nil {
			return err
		}
	}
	return nil
}

func DeleteHistory(ctx context.Context, email string) error {
	db, ok := enabledDB()
	if !ok {
		return nil
	}
	return db.WithContext(ctx).Where("client_email = ?", email).Delete(&IPHistory{}).Error
}
func DeleteEvents(ctx context.Context, email string) error {
	db, ok := enabledDB()
	if !ok {
		return nil
	}
	return db.WithContext(ctx).Where("client_email = ?", email).Delete(&Event{}).Error
}
func Acknowledge(ctx context.Context, email string, id uint) error {
	db, ok := enabledDB()
	if !ok {
		return nil
	}
	return db.WithContext(ctx).Model(&Event{}).Where("id = ? AND client_email = ?", id, email).Update("acknowledged_at", time.Now().UnixMilli()).Error
}
func Suppress(ctx context.Context, email string, kind string, expiresAt int64, reason string) error {
	db, ok := enabledDB()
	if !ok {
		return nil
	}
	if !knownSignal(kind) || len(reason) > 500 {
		return fmt.Errorf("invalid suppression")
	}
	return db.WithContext(ctx).Create(&Suppression{ClientEmail: email, Kind: kind, ExpiresAt: expiresAt, Reason: reason, CreatedAt: time.Now().UnixMilli()}).Error
}

func knownSignal(kind string) bool {
	switch kind {
	case SignalConcurrentIPs, SignalConcurrentNodes, SignalCountryDivergence, SignalASNDivergence, SignalIPChurn, SignalDNSAnomaly, "all":
		return true
	default:
		return false
	}
}

func RegisterJobs(scheduler *cron.Cron) {
	if scheduler == nil {
		return
	}
	_, _ = scheduler.AddFunc("@every 15m", func() { _ = RecomputeAll(context.Background()) })
	_, _ = scheduler.AddFunc("@every 24h", func() { _ = Prune(context.Background(), time.Now()) })
}

func RegisterRoutes(api *gin.RouterGroup) {
	if api == nil {
		return
	}
	api.GET("/risk/clients/:email", func(c *gin.Context) {
		result, err := SummaryFor(c.Request.Context(), c.Param("email"))
		if err != nil {
			c.JSON(500, gin.H{"success": false, "msg": "risk unavailable"})
			return
		}
		c.JSON(200, gin.H{"success": true, "msg": "", "obj": result})
	})
	api.POST("/risk/clients/:email/events/:id/ack", func(c *gin.Context) {
		var id uint
		if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
			c.JSON(400, gin.H{"success": false, "msg": "invalid event"})
			return
		}
		if err := Acknowledge(c.Request.Context(), c.Param("email"), id); err != nil {
			c.JSON(500, gin.H{"success": false, "msg": "acknowledgement failed"})
			return
		}
		c.JSON(200, gin.H{"success": true, "msg": "", "obj": gin.H{"acknowledged": true}})
	})
	api.POST("/risk/clients/:email/suppress", func(c *gin.Context) {
		var input struct {
			Kind      string `json:"kind"`
			ExpiresAt int64  `json:"expiresAt"`
			Reason    string `json:"reason"`
		}
		if c.ShouldBindJSON(&input) != nil || input.Kind == "" {
			c.JSON(400, gin.H{"success": false, "msg": "invalid suppression"})
			return
		}
		if err := Suppress(c.Request.Context(), c.Param("email"), input.Kind, input.ExpiresAt, input.Reason); err != nil {
			c.JSON(500, gin.H{"success": false, "msg": "suppression failed"})
			return
		}
		c.JSON(200, gin.H{"success": true, "msg": "", "obj": gin.H{"saved": true}})
	})
	api.DELETE("/risk/clients/:email/ip-history", func(c *gin.Context) {
		if err := DeleteHistory(c.Request.Context(), c.Param("email")); err != nil {
			c.JSON(500, gin.H{"success": false, "msg": "history deletion failed"})
			return
		}
		c.JSON(200, gin.H{"success": true, "msg": "", "obj": gin.H{"deleted": true}})
	})
	api.DELETE("/risk/clients/:email/events", func(c *gin.Context) {
		if err := DeleteEvents(c.Request.Context(), c.Param("email")); err != nil {
			c.JSON(500, gin.H{"success": false, "msg": "event deletion failed"})
			return
		}
		c.JSON(200, gin.H{"success": true, "msg": "", "obj": gin.H{"deleted": true}})
	})
	api.GET("/risk/settings", func(c *gin.Context) { c.JSON(200, gin.H{"success": true, "msg": "", "obj": CurrentRetention()}) })
	api.POST("/risk/settings", func(c *gin.Context) {
		var input RetentionSettings
		if c.ShouldBindJSON(&input) != nil {
			c.JSON(400, gin.H{"success": false, "msg": "invalid retention"})
			return
		}
		if err := SaveRetention(input); err != nil {
			c.JSON(400, gin.H{"success": false, "msg": err.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true, "msg": "", "obj": CurrentRetention()})
	})
}
