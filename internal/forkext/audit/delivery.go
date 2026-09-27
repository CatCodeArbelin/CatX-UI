package audit

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/eventbus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

const (
	maxWebhookBody       = 64 * 1024
	maxWebhookResponse   = 4 * 1024
	webhookTimeout       = 5 * time.Second
	maxWebhookAttempts   = 10
	webhookLeaseDuration = 30 * time.Second
)

type deliveryWorker struct {
	db   *gorm.DB
	wake chan struct{}
	stop chan struct{}
	done chan struct{}
}

var (
	workerMu sync.Mutex
	worker   *deliveryWorker
	metrics  = newMetrics()
)

func Start(ctx context.Context) (func(), error) {
	if !Enabled() {
		return func() {}, nil
	}
	workerMu.Lock()
	if worker != nil {
		workerMu.Unlock()
		return func() {}, nil
	}
	w := &deliveryWorker{db: database.Load(), wake: make(chan struct{}, 1), stop: make(chan struct{}), done: make(chan struct{})}
	worker = w
	workerMu.Unlock()
	go w.run(ctx)
	return func() { stopWorker(w) }, nil
}

func Stop() {
	workerMu.Lock()
	w := worker
	worker = nil
	workerMu.Unlock()
	if w != nil {
		stopWorker(w)
	}
}

func stopWorker(w *deliveryWorker) {
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
	select {
	case <-w.done:
	case <-time.After(10 * time.Second):
	}
}

func Wake() {
	workerMu.Lock()
	w := worker
	workerMu.Unlock()
	if w != nil {
		select {
		case w.wake <- struct{}{}:
		default:
		}
	}
}

func RegisterEventSubscribers(bus *eventbus.Bus) {
	if bus == nil || !Enabled() {
		return
	}
	bus.Subscribe("fork-audit-webhook-wake", func(eventbus.Event) { Wake() })
}

func RegisterJobs(scheduler *cron.Cron) {
	if scheduler == nil || !Enabled() {
		return
	}
	_, _ = scheduler.AddFunc("@every 1h", func() { Prune(context.Background()) })
}

func Prune(ctx context.Context) {
	if !Enabled() {
		return
	}
	var settings RetentionSettingsRow
	if err := database.Load().WithContext(ctx).First(&settings, 1).Error; err != nil {
		return
	}
	now := time.Now().UTC()
	_ = database.Load().WithContext(ctx).Where("created_at < ?", now.Add(-time.Duration(settings.AuditDays)*24*time.Hour)).Delete(&AuditEvent{}).Error
	_ = database.Load().WithContext(ctx).Where("status = ? AND delivered_at < ?", DeliverySuccess, now.Add(-time.Duration(settings.DeliveryDays)*24*time.Hour)).Delete(&WebhookDelivery{}).Error
	_ = database.Load().WithContext(ctx).Where("status = ? AND created_at < ?", DeliveryDead, now.Add(-time.Duration(settings.DeadLetterDays)*24*time.Hour)).Delete(&WebhookDelivery{}).Error
}

func (w *deliveryWorker) run(ctx context.Context) {
	defer close(w.done)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		w.process(ctx)
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case <-w.wake:
		case <-ticker.C:
		}
	}
}

func (w *deliveryWorker) process(ctx context.Context) {
	for i := 0; i < 20; i++ {
		delivery, endpoint, event, ok := w.claim(ctx)
		if !ok {
			return
		}
		if err := deliver(ctx, endpoint, event, delivery); err != nil {
			w.fail(delivery, err)
		} else {
			w.succeed(delivery)
		}
	}
}

func (w *deliveryWorker) claim(ctx context.Context) (*WebhookDelivery, *WebhookEndpoint, *AuditEvent, bool) {
	var result WebhookDelivery
	now := time.Now().UTC()
	err := w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("(status = ? AND next_attempt_at <= ?) OR (status = ? AND lease_until < ?)", DeliveryPending, now, DeliveryLeased, now).
			Order("next_attempt_at asc").First(&result).Error; err != nil {
			return err
		}
		lease := now.Add(webhookLeaseDuration)
		updated := tx.Model(&WebhookDelivery{}).Where("id = ? AND (status = ? OR (status = ? AND lease_until < ?))", result.ID, DeliveryPending, DeliveryLeased, now).
			Updates(map[string]any{"status": DeliveryLeased, "lease_until": lease, "attempt": gorm.Expr("attempt + 1")})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		result.Status = DeliveryLeased
		result.Attempt++
		result.LeaseUntil = &lease
		return nil
	})
	if err != nil {
		return nil, nil, nil, false
	}
	var endpoint WebhookEndpoint
	var event AuditEvent
	if w.db.WithContext(ctx).First(&endpoint, "id = ?", result.EndpointID).Error != nil || w.db.WithContext(ctx).First(&event, "id = ?", result.EventID).Error != nil {
		w.fail(&result, errors.New("delivery target no longer exists"))
		return nil, nil, nil, false
	}
	return &result, &endpoint, &event, true
}

func deliver(ctx context.Context, endpoint *WebhookEndpoint, event *AuditEvent, delivery *WebhookDelivery) error {
	if err := validateURL(ctx, endpoint.URL); err != nil {
		return err
	}
	secret, err := decryptSecret(endpoint.SecretCipher)
	if err != nil {
		return err
	}
	body, err := json.Marshal(event)
	if err != nil || len(body) > maxWebhookBody {
		return errors.New("webhook payload exceeds size limit")
	}
	u, err := url.Parse(endpoint.URL)
	if err != nil {
		return errInvalidWebhookURL
	}
	timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	signature := sign(secret, timestamp, body)
	reqCtx, cancel := context.WithTimeout(ctx, webhookTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpoint.URL, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "CatX-UI-Webhook/1")
	req.Header.Set("X-CatX-Event-ID", event.ID)
	req.Header.Set("X-CatX-Delivery-ID", delivery.ID)
	req.Header.Set("X-CatX-Timestamp", timestamp)
	req.Header.Set("X-CatX-Signature", "sha256="+signature)
	transport := &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}, DisableKeepAlives: true}
	transport.DialContext = publicDialContext
	client := &http.Client{Transport: transport, Timeout: webhookTimeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	read, copyErr := io.CopyN(io.Discard, resp.Body, maxWebhookResponse+1)
	if copyErr == nil || read > maxWebhookResponse {
		return errors.New("webhook response exceeds size limit")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	_ = u
	return nil
}

func publicDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, errPrivateWebhookAddress
	}
	for _, ip := range ips {
		if !isPublicIP(ip) {
			continue
		}
		conn, err := (&net.Dialer{}).DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
	}
	return nil, errPrivateWebhookAddress
}

func sign(secret, timestamp string, body []byte) string {
	// HMAC implementation is kept in signing.go to make canonicalization easy to test.
	return hmacSignature(secret, timestamp, body)
}

func (w *deliveryWorker) succeed(delivery *WebhookDelivery) {
	now := time.Now().UTC()
	w.db.Model(&WebhookDelivery{}).Where("id = ?", delivery.ID).Updates(map[string]any{"status": DeliverySuccess, "lease_until": nil, "delivered_at": now, "last_error": ""})
	metrics.deliveries.WithLabelValues("success").Inc()
}

func (w *deliveryWorker) fail(delivery *WebhookDelivery, deliveryErr error) {
	now := time.Now().UTC()
	if delivery.Attempt >= maxWebhookAttempts {
		w.db.Model(&WebhookDelivery{}).Where("id = ?", delivery.ID).Updates(map[string]any{"status": DeliveryDead, "lease_until": nil, "last_error": safeError(deliveryErr)})
		metrics.deliveries.WithLabelValues("dead").Inc()
		return
	}
	delay := time.Duration(1<<min(delivery.Attempt, 8)) * time.Second
	next := now.Add(delay)
	w.db.Model(&WebhookDelivery{}).Where("id = ?", delivery.ID).Updates(map[string]any{"status": DeliveryPending, "lease_until": nil, "next_attempt_at": next, "last_error": safeError(deliveryErr)})
	metrics.deliveries.WithLabelValues("retry").Inc()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func safeError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) > 500 {
		message = message[:500]
	}
	return message
}

type metricsSet struct {
	deliveries  *prometheus.CounterVec
	auditWrites *prometheus.CounterVec
	pending     prometheus.Gauge
	dead        prometheus.Gauge
	registry    *prometheus.Registry
}

func newMetrics() *metricsSet {
	registry := prometheus.NewRegistry()
	m := &metricsSet{
		registry:    registry,
		deliveries:  prometheus.NewCounterVec(prometheus.CounterOpts{Name: "catx_webhook_deliveries_total", Help: "Webhook delivery outcomes."}, []string{"outcome"}),
		auditWrites: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "catx_audit_events_total", Help: "Durable audit events written."}, []string{"outcome"}),
		pending:     prometheus.NewGauge(prometheus.GaugeOpts{Name: "catx_webhook_pending_deliveries", Help: "Webhook deliveries awaiting work."}),
		dead:        prometheus.NewGauge(prometheus.GaugeOpts{Name: "catx_webhook_dead_letters", Help: "Webhook deliveries in the dead-letter state."}),
	}
	registry.MustRegister(m.deliveries)
	registry.MustRegister(m.auditWrites, m.pending, m.dead)
	return m
}

func MetricsHandler() http.Handler {
	if db := database.Load(); db != nil {
		var pending, dead int64
		db.Model(&WebhookDelivery{}).Where("status IN ?", []string{DeliveryPending, DeliveryLeased}).Count(&pending)
		db.Model(&WebhookDelivery{}).Where("status = ?", DeliveryDead).Count(&dead)
		metrics.pending.Set(float64(pending))
		metrics.dead.Set(float64(dead))
	}
	return promhttp.HandlerFor(metrics.registry, promhttp.HandlerOpts{ErrorHandling: promhttp.HTTPErrorOnError})
}
