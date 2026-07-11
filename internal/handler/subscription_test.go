package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"subdock/internal/config"
	"subdock/internal/model"
)

// fakeNotifier 为 Handler 测试隔离外部通知网络调用。
type fakeNotifier struct{}

// SendTelegram 模拟成功发送 Telegram 消息。
func (fakeNotifier) SendTelegram(_, _, _ string) error { return nil }

// SendBark 模拟成功发送 Bark 消息。
func (fakeNotifier) SendBark(_, _, _ string) error { return nil }

// TestUpdateSubscriptionSupportsExplicitZeroAndClearing 验证更新可清空备注、设为当天提醒并覆盖到期日。
func TestUpdateSubscriptionSupportsExplicitZeroAndClearing(t *testing.T) {
	handler, db, subscription := newSubscriptionTestHandler(t)
	router := gin.New()
	router.PUT("/subscriptions/:id", handler.UpdateSubscription)

	body := map[string]interface{}{
		"remark":      "",
		"remind_days": 0,
		"expire_date": "2026-08-31",
	}
	response := performJSONRequest(t, router, http.MethodPut, "/subscriptions/"+strconv.Itoa(int(subscription.ID)), body)
	if response.Code != http.StatusOK {
		t.Fatalf("更新返回状态 %d，响应 %s", response.Code, response.Body.String())
	}

	var updated model.Subscription
	if err := db.First(&updated, subscription.ID).Error; err != nil {
		t.Fatalf("查询更新后的订阅失败: %v", err)
	}
	if updated.Remark != "" {
		t.Fatalf("备注为 %q，期望空字符串", updated.Remark)
	}
	if updated.RemindDays != 0 {
		t.Fatalf("提醒天数为 %d，期望 0", updated.RemindDays)
	}
	if got := updated.ExpireDate.Format(dateLayout); got != "2026-08-31" {
		t.Fatalf("到期日期为 %s，期望 2026-08-31", got)
	}
}

// TestUpdateSubscriptionRejectsInvalidCycleUnit 验证更新接口不会写入非法周期单位。
func TestUpdateSubscriptionRejectsInvalidCycleUnit(t *testing.T) {
	handler, db, subscription := newSubscriptionTestHandler(t)
	router := gin.New()
	router.PUT("/subscriptions/:id", handler.UpdateSubscription)

	response := performJSONRequest(t, router, http.MethodPut, "/subscriptions/"+strconv.Itoa(int(subscription.ID)), map[string]interface{}{
		"cycle_unit": "weekly",
	})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("非法周期返回状态 %d，期望 400", response.Code)
	}

	var unchanged model.Subscription
	if err := db.First(&unchanged, subscription.ID).Error; err != nil {
		t.Fatalf("查询订阅失败: %v", err)
	}
	if unchanged.CycleUnit != model.CycleUnitMonth {
		t.Fatalf("周期单位被意外更新为 %s", unchanged.CycleUnit)
	}
}

// TestDeleteSubscriptionReturnsNotFound 验证删除不存在记录时返回 404。
func TestDeleteSubscriptionReturnsNotFound(t *testing.T) {
	handler, _, _ := newSubscriptionTestHandler(t)
	router := gin.New()
	router.DELETE("/subscriptions/:id", handler.DeleteSubscription)

	request := httptest.NewRequest(http.MethodDelete, "/subscriptions/99999", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("删除不存在订阅返回状态 %d，期望 404", response.Code)
	}
}

// newSubscriptionTestHandler 创建带有一条订阅的临时 Handler 和数据库。
func newSubscriptionTestHandler(t *testing.T) (*Handler, *gorm.DB, model.Subscription) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "handler.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.Subscription{}, &model.SubscriptionRenewal{}, &model.Setting{}); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}
	subscription := model.Subscription{
		Name:       "测试订阅",
		Amount:     18,
		Currency:   "CNY",
		StartDate:  time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
		CycleValue: 1,
		CycleUnit:  model.CycleUnitMonth,
		ExpireDate: time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
		RemindDays: 3,
		Remark:     "待清空",
	}
	if err := db.Create(&subscription).Error; err != nil {
		t.Fatalf("创建测试订阅失败: %v", err)
	}
	cfg := &config.Config{JWTSecret: "0123456789abcdef0123456789abcdef", WebsiteTitle: "SubDock"}
	return NewWithNotifier(db, cfg, fakeNotifier{}), db, subscription
}

// performJSONRequest 编码 JSON 请求并通过指定 Gin 路由执行。
func performJSONRequest(t *testing.T, router http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("编码请求失败: %v", err)
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
