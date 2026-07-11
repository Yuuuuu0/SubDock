package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"subdock/internal/config"
	"subdock/internal/model"
)

// TestUnknownAPIRouteReturnsJSONNotFound 验证未知 API 不会被 SPA 首页回退为 200。
func TestUnknownAPIRouteReturnsJSONNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	serveStatic(router)

	request := httptest.NewRequest(http.MethodGet, "/api/not-exists", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("未知 API 返回状态 %d，期望 404", response.Code)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("未知 API Content-Type 为 %q，期望 JSON", contentType)
	}
}

// TestSubscriptionReadRoutesRegistered 验证单条订阅和续订历史查询路由均已注册。
func TestSubscriptionReadRoutesRegistered(t *testing.T) {
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	if _, err := config.Load(); err != nil {
		t.Fatalf("加载测试配置失败: %v", err)
	}
	db, err := model.InitDB()
	if err != nil {
		t.Fatalf("初始化测试数据库失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取测试数据库连接失败: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	routes := Setup().Routes()
	wanted := map[string]bool{
		"GET /api/subscriptions/:id":          false,
		"GET /api/subscriptions/:id/renewals": false,
	}
	for _, route := range routes {
		key := route.Method + " " + route.Path
		if _, exists := wanted[key]; exists {
			wanted[key] = true
		}
	}
	for route, found := range wanted {
		if !found {
			t.Fatalf("未注册路由 %s", route)
		}
	}
}
