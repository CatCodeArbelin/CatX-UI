package analytics

// This file contains the metadata-only access.log pipeline. It intentionally
// does not share the web log viewer's scanner: ingestion needs durable offsets,
// rotation semantics, bounded backpressure, and transactional commits.

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

const (
	defaultPollInterval = 250 * time.Millisecond
	defaultQueueSize    = 256
	defaultBatchSize    = 64
	maxPartialLine      = 1 << 20
)

// AccessLogEntry is the safe, structured subset of an Xray access-log line.
// It contains connection metadata only; the log line is never persisted.
type AccessLogEntry struct {
	ObservedAt  time.Time
	From        string
	Destination string
	Inbound     string
	Outbound    string
	Email       string
}

// ParseAccessLog parses Xray's accepted-connection format defensively.
func ParseAccessLog(line string) (AccessLogEntry, bool) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return AccessLogEntry{}, false
	}
	ts, err := time.ParseInLocation("2006/01/02 15:04:05.999999", parts[0]+" "+parts[1], time.Local)
	if err != nil {
		return AccessLogEntry{}, false
	}
	var e AccessLogEntry
	e.ObservedAt = ts.UTC()
	for i, part := range parts[2:] {
		i += 2
		switch {
		case part == "from" && i+1 < len(parts):
			e.From = strings.TrimLeft(parts[i+1], "/")
		case part == "accepted" && i+1 < len(parts):
			e.Destination = strings.TrimLeft(parts[i+1], "/")
		case strings.HasPrefix(part, "["):
			e.Inbound = strings.TrimPrefix(part, "[")
		case strings.HasSuffix(part, "]"):
			e.Outbound = strings.TrimSuffix(part, "]")
		case part == "email:" && i+1 < len(parts):
			e.Email = parts[i+1]
		}
	}
	if e.Destination == "" {
		return AccessLogEntry{}, false
	}
	return e, true
}

type lineRecord struct {
	identity, line string
	end            int64
	generation     int64
	key            string
}

// TailerConfig controls resource bounds and timing. The repository is the
// existing analytics repository; no second persistence path is introduced.
type TailerConfig struct {
	Path                 string
	Repo                 Repository
	PollInterval         time.Duration
	QueueSize, BatchSize int
	Logger               *log.Logger
}

// Tailer exposes observable ingestion accounting without making it a source
// of online state. Counters are safe to read while Run is active.
type Tailer struct {
	cfg                            TailerConfig
	mu                             sync.Mutex
	deferred, committed, malformed uint64
}

func NewTailer(cfg TailerConfig) *Tailer {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = defaultPollInterval
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = defaultQueueSize
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultBatchSize
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	return &Tailer{cfg: cfg}
}

func (t *Tailer) Stats() (deferred, committed, malformed uint64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.deferred, t.committed, t.malformed
}
func (t *Tailer) addDeferred()          { t.mu.Lock(); t.deferred++; t.mu.Unlock() }
func (t *Tailer) addCommitted(n uint64) { t.mu.Lock(); t.committed += n; t.mu.Unlock() }
func (t *Tailer) addMalformed()         { t.mu.Lock(); t.malformed++; t.mu.Unlock() }

// Run exits on context cancellation. File and database errors are retried and
// reported, never returned as a panel/Xray fatal error.
func (t *Tailer) Run(ctx context.Context) error {
	if t == nil || t.cfg.Repo == nil || t.cfg.Path == "" {
		return nil
	}
	path, err := filepath.Abs(t.cfg.Path)
	if err != nil {
		return nil
	}
	cursor, err := loadCursor(ctx, t.cfg.Repo, path)
	if err != nil {
		t.cfg.Logger.Printf("analytics access.log cursor load: %v", err)
		cursor = AccessLogCursor{CursorKey: path}
	}
	q := make(chan lineRecord, t.cfg.QueueSize)
	workerDone := make(chan struct{})
	workerCtx, workerCancel := context.WithTimeout(context.Background(), 10*time.Second)
	go func() { defer close(workerDone); defer workerCancel(); t.persist(workerCtx, q, path) }()
	defer func() { close(q); <-workerDone }()

	identity, offset, generation := cursor.FileIdentity, cursor.Offset, cursor.Generation
	var pending string
	var active *os.File
	var activeIdentity string
	defer func() {
		if active != nil {
			_ = active.Close()
		}
	}()
	ticker := time.NewTicker(t.cfg.PollInterval)
	defer ticker.Stop()
	for {
		if active == nil {
			f, id, size, openErr := openCurrent(path)
			if openErr != nil {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					continue
				}
			}
			if identity != id || offset > size {
				if identity == id && offset > size {
					generation++
				} else {
					generation = 0
				}
				identity, offset, pending = id, 0, ""
			}
			active, activeIdentity = f, id
		}
		read, next, partial, readErr := readLines(active, offset, pending)
		if !errors.Is(readErr, io.EOF) {
			_ = active.Close()
			active = nil
			continue
		}
		pending = partial
		for _, item := range read {
			item.identity = activeIdentity
			item.generation = generation
			item.key = digest(activeIdentity + ":" + strconv.FormatInt(generation, 10) + ":" + strconv.FormatInt(item.end, 10) + ":" + item.line)
			select {
			case q <- item:
				offset = item.end
				next = offset
			case <-ctx.Done():
				return nil
			default:
				t.addDeferred()
				select {
				case q <- item:
					offset = item.end
					next = offset
				case <-ctx.Done():
					return nil
				}
			}
		}
		offset = next
		if pending == "" {
			if rotated, err := currentChanged(path, activeIdentity, offset); err == nil && rotated {
				_ = active.Close()
				active = nil
				identity = activeIdentity
				continue
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func readLines(f *os.File, offset int64, pending string) ([]lineRecord, int64, string, error) {
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return nil, offset, pending, err
	}
	r := bufio.NewReaderSize(f, 32*1024)
	next := offset
	data := pending
	pendingLen := len([]byte(pending))
	var out []lineRecord
	for {
		part, err := r.ReadString('\n')
		data += part
		if len(data) > maxPartialLine {
			if i := strings.IndexByte(data, '\n'); i >= 0 {
				data = data[i+1:]
			} else {
				data = data[:0]
			}
		}
		if i := strings.IndexByte(data, '\n'); i >= 0 {
			line := strings.TrimSuffix(strings.TrimSuffix(data[:i], "\r"), "\n")
			prefixBytes := len([]byte(data[:i+1]))
			fromFile := prefixBytes - pendingLen
			if fromFile < 0 {
				fromFile = 0
			}
			next += int64(fromFile)
			pendingLen -= prefixBytes
			if pendingLen < 0 {
				pendingLen = 0
			}
			out = append(out, lineRecord{line: line, end: next})
			data = data[i+1:]
			continue
		}
		if err != nil {
			return out, next, data, err
		}
	}
}

func currentChanged(path, identity string, offset int64) (bool, error) {
	f, id, size, err := openCurrent(path)
	if err != nil {
		return false, err
	}
	_ = f.Close()
	return id != identity || size < offset, nil
}

func openCurrent(path string) (*os.File, string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", 0, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, "", 0, err
	}
	return f, handleIdentity(f, info), info.Size(), nil
}

func digest(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }

func loadCursor(ctx context.Context, repo Repository, key string) (AccessLogCursor, error) {
	return repo.LoadAccessLogCursor(ctx, key)
}

func (t *Tailer) persist(ctx context.Context, q <-chan lineRecord, path string) {
	batch := make([]lineRecord, 0, t.cfg.BatchSize)
	flush := func() bool {
		if len(batch) == 0 {
			return true
		}
		events, sessions := make([]MetadataEvent, 0, len(batch)), make(map[string]NetworkSession)
		evidence := NewEvidenceService(t.cfg.Repo, EvidenceEnabled())
		for _, record := range batch {
			e, ok := normalize(record)
			if !ok {
				t.addMalformed()
				continue
			}
			events = append(events, e)
			if _, err := evidence.Correlate(ctx, DestinationInput{ObservedAt: time.UnixMilli(e.ObservedAt), ClientEmail: e.ClientEmail, NodeID: e.NodeID, InboundID: e.InboundID, DestinationIP: e.DestinationIP, DirectDomain: e.Domain, SessionKey: e.SessionKey, Source: e.Source, EventKey: e.EventKey}); err != nil {
				t.cfg.Logger.Printf("analytics access.log evidence deferred: %v", err)
			}
			s := sessionFor(e)
			if old, exists := sessions[s.SessionKey]; !exists || s.LastSeen > old.LastSeen {
				if exists && old.FirstSeen < s.FirstSeen {
					s.FirstSeen = old.FirstSeen
				}
				sessions[s.SessionKey] = s
			}
		}
		if len(batch) > 0 {
			last := batch[len(batch)-1]
			cursor := AccessLogCursor{CursorKey: path, FileIdentity: last.identity, Offset: last.end, Generation: last.generation, UpdatedAt: time.Now().UnixMilli()}
			if err := t.cfg.Repo.CommitAccessLogBatch(ctx, events, mapsToSlice(sessions), cursor); err != nil {
				t.cfg.Logger.Printf("analytics access.log batch deferred: %v", err)
				return false
			}
			t.addCommitted(uint64(len(events)))
		}
		batch = batch[:0]
		return true
	}
	for {
		select {
		case record, ok := <-q:
			if !ok {
				for !flush() {
					select {
					case <-ctx.Done():
						return
					case <-time.After(100 * time.Millisecond):
					}
				}
				return
			}
			batch = append(batch, record)
			if len(batch) >= t.cfg.BatchSize {
				for !flush() {
					select {
					case <-ctx.Done():
						return
					case <-time.After(100 * time.Millisecond):
					}
				}
			}
		case <-ctx.Done():
			flush()
			return
		}
	}
}

func mapsToSlice(m map[string]NetworkSession) []NetworkSession {
	out := make([]NetworkSession, 0, len(m))
	for _, s := range m {
		out = append(out, s)
	}
	return out
}

func normalize(r lineRecord) (MetadataEvent, bool) {
	e, ok := ParseAccessLog(r.line)
	if !ok {
		return MetadataEvent{}, false
	}
	domain, ip, port := destination(e.Destination)
	if domain == "" && ip == "" {
		return MetadataEvent{}, false
	}
	return MetadataEvent{ObservedAt: e.ObservedAt.UnixMilli(), ClientEmail: e.Email, Domain: domain, DestinationIP: ip, Port: port, Protocol: protocol(e.Destination), SessionKey: sessionKey(e), Source: SourceAccessLog, Provenance: ProvenanceObserved, Confidence: 1, EventKey: r.key}, true
}

func destination(raw string) (string, string, int) {
	raw = strings.TrimPrefix(raw, "tcp:")
	host, p, err := net.SplitHostPort(raw)
	if err != nil {
		return strings.Trim(raw, "[]"), "", 0
	}
	port, _ := strconv.Atoi(p)
	if net.ParseIP(host) != nil {
		return "", host, port
	}
	return strings.Trim(host, "[]"), "", port
}

func protocol(raw string) string {
	if i := strings.IndexByte(raw, ':'); i > 0 {
		return strings.ToLower(raw[:i])
	}
	return ""
}

func sessionKey(e AccessLogEntry) string {
	return digest(strings.Join([]string{e.Email, e.Inbound, e.From, protocol(e.Destination), e.ObservedAt.UTC().Format("200601021504")}, "|"))
}

func sessionFor(e MetadataEvent) NetworkSession {
	return NetworkSession{SessionKey: e.SessionKey, ClientEmail: e.ClientEmail, InboundID: e.InboundID, FirstSeen: e.ObservedAt, LastSeen: e.ObservedAt, Protocol: e.Protocol, Source: SourceAccessLog, Provenance: ProvenanceCorrelated, Confidence: 0.8}
}

var configured struct {
	sync.RWMutex
	repo            Repository
	enabled         bool
	evidenceEnabled bool
}

// Configure installs the analytics repository after migrations. A disabled or
// failed configuration is represented by the no-op repository.
func Configure(repo Repository, enabled bool) {
	configured.Lock()
	defer configured.Unlock()
	if repo == nil {
		repo = NoopRepository{}
	}
	configured.repo, configured.enabled, configured.evidenceEnabled = repo, enabled, false
}

// SetEvidenceEnabled applies the independent DNS intelligence flag after the
// analytics repository has been configured. It never enables analytics itself.
func SetEvidenceEnabled(enabled bool) {
	configured.Lock()
	defer configured.Unlock()
	configured.evidenceEnabled = configured.enabled && enabled
}

func EvidenceEnabled() bool {
	configured.RLock()
	defer configured.RUnlock()
	return configured.enabled && configured.evidenceEnabled
}

// Start starts the collector only when analytics was explicitly enabled.
func Start(ctx context.Context) (func(), error) {
	configured.RLock()
	repo, enabled := configured.repo, configured.enabled
	configured.RUnlock()
	if !enabled || repo == nil {
		return func() {}, nil
	}
	path, err := xray.GetAccessLogPath()
	if err != nil || path == "" || path == "none" {
		return func() {}, nil
	}
	child, cancel := context.WithCancel(ctx)
	tailer := NewTailer(TailerConfig{Path: path, Repo: repo})
	go func() { _ = tailer.Run(child) }()
	return cancel, nil
}
