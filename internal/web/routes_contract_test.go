package web

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/dbtest"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext/audit"
	"github.com/mhsanaei/3x-ui/v3/internal/web/global"
)

/*
frontend/src/pages/api-docs/endpoints.ts is a hand-maintained registry: an
API route omitted there silently vanishes from the generated OpenAPI docs,
and an entry for a removed route documents an endpoint that 404s. This test
constructs the real router and diffs it against the registry both ways.

Scope: everything under /panel/api/ plus the session-auth surface the
registry also documents (/login, /logout, /csrf-token,
/getTwoFactorEnable, /ws). SPA page routes are UI, not API, and stay out;
registry paths that start with "/{" describe the standalone subscription
server, which this engine does not serve.
*/

var contractExtraRoutes = map[string]bool{
	"POST /login":              true,
	"POST /logout":             true,
	"GET /csrf-token":          true,
	"POST /getTwoFactorEnable": true,
	"GET /ws":                  true,
}

func inContractScope(method, path string) bool {
	return strings.HasPrefix(path, "/panel/api/") || contractExtraRoutes[method+" "+path]
}

func registeredContractRoutes(t *testing.T) map[string]bool {
	t.Helper()
	dbtest.InitDB(t, filepath.Join(t.TempDir(), "x-ui.db"))
	// The fork audit routes are intentionally absent while the feature is
	// disabled. Exercise the documented fork surface in its enabled state so
	// this contract test does not mistake the feature-off no-op for stale API
	// documentation.
	if err := forkext.NewSettings(database.GetDB()).Set(forkext.FlagAudit, true); err != nil {
		t.Fatalf("enable audit for route contract: %v", err)
	}
	audit.Configure(database.GetDB(), true)
	t.Cleanup(func() { audit.Configure(nil, false) })

	previous := global.GetWebServer()
	s := NewServer()
	s.cron = cron.New(cron.WithLocation(time.Local), cron.WithSeconds())
	global.SetWebServer(s)
	t.Cleanup(func() {
		s.cancel()
		global.SetWebServer(previous)
	})

	engine, err := s.initRouter()
	if err != nil {
		t.Fatalf("init router: %v", err)
	}
	routes := make(map[string]bool)
	for _, r := range engine.Routes() {
		routes[r.Method+" "+r.Path] = true
	}
	if len(routes) == 0 {
		t.Fatal("no routes registered; router construction is broken")
	}
	return routes
}

func documentedContractRoutes(t *testing.T) map[string]bool {
	t.Helper()
	paths := []string{
		filepath.Join("..", "..", "frontend", "src", "pages", "api-docs", "endpoints.ts"),
		filepath.Join("..", "..", "frontend", "src", "forkext", "registry.ts"),
	}
	var sources []string
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		sources = append(sources, string(source))
	}
	text := strings.Join(sources, "\n")
	methodRe := regexp.MustCompile(`method:\s*'(GET|POST|PUT|DELETE|PATCH|HEAD|WS)'`)
	pathRe := regexp.MustCompile(`path:\s*'([^']+)'`)
	methods := methodRe.FindAllStringSubmatchIndex(text, -1)
	if declared := strings.Count(text, "method: '"); len(methods) != declared {
		t.Fatalf("parsed %d method fields but API registries declare %d — the parser regex no longer matches the file shape", len(methods), declared)
	}
	docs := make(map[string]bool)
	for i, m := range methods {
		segmentEnd := len(text)
		if i+1 < len(methods) {
			segmentEnd = methods[i+1][0]
		}
		pathMatch := pathRe.FindStringSubmatch(text[m[1]:segmentEnd])
		if pathMatch == nil {
			t.Fatalf("entry %d in API registries has a method but no path before the next entry — the parser cannot pair it", i)
		}
		method := text[m[2]:m[3]]
		if strings.HasPrefix(pathMatch[1], "/{") || !strings.HasPrefix(pathMatch[1], "/") {
			continue
		}
		docs[method+" "+pathMatch[1]] = true
	}
	if len(docs) == 0 {
		t.Fatal("no entries parsed from API registries; the parser regex is broken")
	}
	return docs
}

func TestRouteRegistryContract(t *testing.T) {
	registered := registeredContractRoutes(t)
	documented := documentedContractRoutes(t)

	t.Run("every API route is documented", func(t *testing.T) {
		var missing []string
		for route := range registered {
			fields := strings.Fields(route)
			if inContractScope(fields[0], fields[1]) && !documented[route] {
				missing = append(missing, route)
			}
		}
		sort.Strings(missing)
		for _, route := range missing {
			t.Error(fmt.Errorf("route %s is registered but absent from endpoints.ts — add an entry or it vanishes from the API docs", route))
		}
	})

	t.Run("every documented route is registered", func(t *testing.T) {
		var stale []string
		for route := range documented {
			if !registered[route] {
				stale = append(stale, route)
			}
		}
		sort.Strings(stale)
		for _, route := range stale {
			t.Error(fmt.Errorf("endpoints.ts documents %s but the server does not register it — remove or fix the entry", route))
		}
	})
}
