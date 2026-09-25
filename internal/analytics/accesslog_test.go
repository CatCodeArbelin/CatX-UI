package analytics

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

type captureRepo struct {
	NoopRepository
	mu       sync.Mutex
	cursor   AccessLogCursor
	events   []MetadataEvent
	sessions []NetworkSession
	dns      []DNSObservation
}

func (r *captureRepo) LoadAccessLogCursor(_ context.Context, key string) (AccessLogCursor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cursor.CursorKey == "" {
		return AccessLogCursor{CursorKey: key}, nil
	}
	return r.cursor, nil
}

func (r *captureRepo) CommitAccessLogBatch(_ context.Context, events []MetadataEvent, sessions []NetworkSession, cursor AccessLogCursor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, events...)
	r.sessions = append(r.sessions, sessions...)
	r.cursor = cursor
	return nil
}

func (r *captureRepo) ListDestinationPage(_ context.Context, _ string, _, _ int64, limit, offset int) (DestinationPage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	end := offset + limit
	if end > len(r.events) {
		end = len(r.events)
	}
	if offset > len(r.events) {
		offset = len(r.events)
	}
	items := make([]DestinationObservation, 0, end-offset)
	for _, event := range r.events[offset:end] {
		items = append(items, event.Observation())
	}
	return DestinationPage{Items: items, Total: int64(len(r.events))}, nil
}

func (r *captureRepo) ListSessionPage(_ context.Context, _ string, _, _ int64, limit, offset int) (SessionPage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	end := offset + limit
	if end > len(r.sessions) {
		end = len(r.sessions)
	}
	if offset > len(r.sessions) {
		offset = len(r.sessions)
	}
	items := append([]NetworkSession(nil), r.sessions[offset:end]...)
	return SessionPage{Items: items, Total: int64(len(r.sessions))}, nil
}

func (r *captureRepo) ListDNSPage(_ context.Context, _ string, _, _ int64, limit, offset int) (DNSPage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	end := offset + limit
	if end > len(r.dns) {
		end = len(r.dns)
	}
	if offset > len(r.dns) {
		offset = len(r.dns)
	}
	return DNSPage{Items: append([]DNSObservation(nil), r.dns[offset:end]...), Total: int64(len(r.dns))}, nil
}

func (r *captureRepo) snapshot() (int, AccessLogCursor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.events), r.cursor
}

const testLine = "2026/09/24 12:00:00.123456 from 203.0.113.4:555 accepted tcp:example.com:443 [in >> out] email: alice@example.com\n"

func TestParseAccessLogMetadataOnly(t *testing.T) {
	e, ok := ParseAccessLog(strings.TrimSpace(testLine))
	if !ok {
		t.Fatal("valid line rejected")
	}
	if e.Email != "alice@example.com" || e.Destination != "tcp:example.com:443" {
		t.Fatalf("unexpected entry: %+v", e)
	}
	if _, ok := ParseAccessLog("not a valid line"); ok {
		t.Fatal("malformed line accepted")
	}
}

func FuzzParseAccessLog(f *testing.F) {
	f.Add(testLine)
	f.Add("")
	f.Add("2026/01/01 00:00:00 accepted tcp:[::1]:443")
	f.Fuzz(func(t *testing.T, line string) { _, _ = ParseAccessLog(line) })
}

func TestTailerPartialLineAndResume(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "access.log")
	if err := os.WriteFile(path, []byte(strings.TrimSuffix(testLine, "\n")), 0o600); err != nil {
		t.Fatal(err)
	}
	repo := &captureRepo{}
	tailer := NewTailer(TailerConfig{Path: path, Repo: repo, PollInterval: 5 * time.Millisecond, BatchSize: 1, QueueSize: 2})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = tailer.Run(ctx); close(done) }()
	time.Sleep(20 * time.Millisecond)
	if n, _ := repo.snapshot(); n != 0 {
		t.Fatalf("partial line committed: %d", n)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString("\n")
	_ = f.Close()
	waitEvents(t, repo, 1)
	cancel()
	<-done

	if n, c := repo.snapshot(); n != 1 || c.Offset == 0 {
		t.Fatalf("resume state = events %d cursor %+v", n, c)
	}
	// Restart at the committed cursor must not replay the event.
	ctx2, cancel2 := context.WithCancel(context.Background())
	done2 := make(chan struct{})
	go func() { _ = tailer.Run(ctx2); close(done2) }()
	time.Sleep(20 * time.Millisecond)
	cancel2()
	<-done2
	if n, _ := repo.snapshot(); n != 1 {
		t.Fatalf("restart replayed event: %d", n)
	}
}

func TestTailerTruncationAndRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "access.log")
	if err := os.WriteFile(path, []byte(testLine), 0o600); err != nil {
		t.Fatal(err)
	}
	repo := &captureRepo{}
	tailer := NewTailer(TailerConfig{Path: path, Repo: repo, PollInterval: 5 * time.Millisecond, BatchSize: 1})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = tailer.Run(ctx); close(done) }()
	waitEvents(t, repo, 1)
	if err := os.Truncate(path, 0); err != nil {
		t.Fatal(err)
	}
	// Give the live tailer a poll opportunity to observe a true zero-length
	// truncation before the replacement content is written.
	time.Sleep(20 * time.Millisecond)
	if err := os.WriteFile(path, []byte(testLine), 0o600); err != nil {
		t.Fatal(err)
	}
	rotated := filepath.Join(dir, "access.log.1")
	if runtime.GOOS == "windows" {
		cancel()
		<-done
		if err := os.Rename(path, rotated); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(testLine), 0o600); err != nil {
			t.Fatal(err)
		}
		ctx, cancel = context.WithCancel(context.Background())
		done = make(chan struct{})
		go func() { _ = tailer.Run(ctx); close(done) }()
	} else if err := os.Rename(path, rotated); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if err := os.WriteFile(path, []byte(testLine), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	waitEvents(t, repo, 2)
	cancel()
	<-done
	if n, _ := repo.snapshot(); n < 2 {
		t.Fatalf("rotation/truncation events = %d", n)
	}
}

func TestTailerStatsConcurrentWithShutdown(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "access.log")
	if err := os.WriteFile(path, []byte(testLine), 0o600); err != nil {
		t.Fatal(err)
	}
	repo := &captureRepo{}
	tailer := NewTailer(TailerConfig{Path: path, Repo: repo, PollInterval: time.Millisecond, BatchSize: 1, QueueSize: 1})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = tailer.Run(ctx); close(done) }()
	stopReaders := make(chan struct{})
	var readers sync.WaitGroup
	for i := 0; i < 4; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stopReaders:
					return
				default:
					_, _, _ = tailer.Stats()
				}
			}
		}()
	}
	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done
	close(stopReaders)
	readers.Wait()
}

func waitEvents(t *testing.T, repo *captureRepo, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if n, _ := repo.snapshot(); n >= want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	n, _ := repo.snapshot()
	t.Fatalf("timed out waiting for %d events, got %d", want, n)
}
