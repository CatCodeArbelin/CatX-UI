package controller

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/forkext"
	"github.com/mhsanaei/3x-ui/v3/internal/web/global"
	"github.com/mhsanaei/3x-ui/v3/internal/web/locale"
)

// M00-T05 exercises the actual restartPanel controller/service boundary. The
// global restart hook is the production OS/process handoff seam: the test
// applies the already-prepared configuration only when that asynchronous
// handoff fires, so the HTTP response cannot accidentally prove activation.
func TestM00T05RestartPanelReportsPendingThenApplies(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catx-m00-restart-panel.db")
	if err := database.InitDB(path); err != nil {
		if strings.Contains(err.Error(), "CGO_ENABLED=0") || strings.Contains(err.Error(), "requires cgo") {
			t.Skip("BLOCKED: local restartPanel acceptance requires a CGO-enabled Go toolchain")
		}
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() {
		global.SetRestartHook(nil)
		_ = database.CloseDB()
	})

	db := database.GetDB()
	if err := forkext.NewSettings(db).Set(forkext.FlagAnalytics, true); err != nil {
		t.Fatalf("persist analytics desired state: %v", err)
	}
	if err := forkext.PrepareRuntimeFromSettings(db); err != nil {
		t.Fatalf("prepare analytics desired state: %v", err)
	}

	applyResult := make(chan error, 1)
	global.SetRestartHook(func() {
		applyResult <- forkext.ReloadRuntimeFromSettings(database.GetDB())
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("I18n", func(_ locale.I18nType, key string, _ ...string) string { return key })
		c.Next()
	})
	settingController := &SettingController{}
	router.POST("/panel/api/setting/restartPanel", settingController.restartPanel)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/panel/api/setting/restartPanel", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"success":true`) || !strings.Contains(response.Body.String(), `"state":"pending"`) {
		t.Fatalf("restartPanel response = %d %s, want successful pending acknowledgement", response.Code, response.Body.String())
	}

	items, err := forkext.NewSettings(db).FeatureFlags()
	if err != nil {
		t.Fatalf("read pending feature state: %v", err)
	}
	for _, item := range items {
		if item.Key == forkext.FlagAnalytics && (item.Active || item.State == forkext.RuntimeStateActive) {
			t.Fatalf("analytics became active before asynchronous restart apply: %+v", item)
		}
	}

	if err := <-applyResult; err != nil {
		t.Fatalf("restart hook runtime apply: %v", err)
	}
	items, err = forkext.NewSettings(db).FeatureFlags()
	if err != nil {
		t.Fatalf("read final feature state: %v", err)
	}
	for _, item := range items {
		if item.Key == forkext.FlagAnalytics {
			if !item.Active || item.State != forkext.RuntimeStateActive || item.RestartRequired {
				t.Fatalf("analytics final restart state = %+v, want active", item)
			}
			return
		}
	}
	t.Fatal("analytics flag missing from final Feature Settings response")
}
