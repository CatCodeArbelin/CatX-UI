package runtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemotePanelUpdateStartAndStatusAreTypedAndCorrelated(t *testing.T) {
	var postCount, statusCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/panel/api/server/updatePanel":
			postCount++
			_, _ = w.Write([]byte(`{"success":true,"obj":{"runId":"run-123"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/panel/api/server/getUpdateStatus":
			statusCount++
			_, _ = w.Write([]byte(`{"success":true,"obj":{"runId":"run-123","state":"success","exitCode":0,"finishedAt":1735689600,"rolledBack":false,"rollbackHealthy":false}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	r := NewRemote(nodeForPlainServer(t, srv, "verify", "tok"), nil)
	started, err := r.StartUpdate(context.Background(), false)
	if err != nil {
		t.Fatalf("StartUpdate: %v", err)
	}
	if started.RunID != "run-123" {
		t.Fatalf("run ID = %q, want run-123", started.RunID)
	}
	status, err := r.GetUpdateStatus(context.Background())
	if err != nil {
		t.Fatalf("GetUpdateStatus: %v", err)
	}
	if status.RunID != started.RunID || status.State != "success" || status.ExitCode != 0 {
		t.Fatalf("status = %+v, want matching successful evidence", status)
	}
	if postCount != 1 || statusCount != 1 {
		t.Fatalf("requests = post %d/status %d, want 1/1", postCount, statusCount)
	}
}

func TestRemotePanelUpdateStartRejectsMissingRunID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"obj":{"runId":""}}`))
	}))
	defer srv.Close()
	r := NewRemote(nodeForPlainServer(t, srv, "verify", "tok"), nil)
	if _, err := r.StartUpdate(context.Background(), false); err == nil {
		t.Fatal("StartUpdate accepted an empty run ID")
	}
}
