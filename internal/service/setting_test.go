package service

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"subdock/internal/model"
)

// TestParseNotifyHoursAcceptsZeroToTwentyThree 验证 0 点、23 点、去重和排序语义。
func TestParseNotifyHoursAcceptsZeroToTwentyThree(t *testing.T) {
	hours, err := ParseNotifyHours("23,0,9,0")
	if err != nil {
		t.Fatalf("解析通知时段失败: %v", err)
	}
	want := []int{0, 9, 23}
	if !reflect.DeepEqual(hours, want) {
		t.Fatalf("通知时段为 %v，期望 %v", hours, want)
	}

	normalized, err := NormalizeNotifyHours("23,0,9,0")
	if err != nil {
		t.Fatalf("规范化通知时段失败: %v", err)
	}
	if normalized != "0,9,23" {
		t.Fatalf("规范化结果为 %q，期望 0,9,23", normalized)
	}
}

// TestParseNotifyHoursRejectsOutOfRange 验证 24 点和负数配置会被拒绝。
func TestParseNotifyHoursRejectsOutOfRange(t *testing.T) {
	for _, value := range []string{"24", "-1", "invalid"} {
		if _, err := ParseNotifyHours(value); err == nil {
			t.Fatalf("通知时段 %q 应返回错误", value)
		}
	}
}

// TestSettingServiceAllowsClearingValue 验证空字符串会覆盖旧配置而非被忽略。
func TestSettingServiceAllowsClearingValue(t *testing.T) {
	db := openServiceTestDB(t)
	settings := NewSettingService(db)
	ctx := context.Background()

	if err := settings.UpdateMany(ctx, map[string]string{SettingBarkURL: "https://example.com/key"}); err != nil {
		t.Fatalf("写入初始设置失败: %v", err)
	}
	if err := settings.UpdateMany(ctx, map[string]string{SettingBarkURL: ""}); err != nil {
		t.Fatalf("清空设置失败: %v", err)
	}
	value, err := settings.Get(ctx, SettingBarkURL, "default")
	if err != nil {
		t.Fatalf("读取设置失败: %v", err)
	}
	if value != "" {
		t.Fatalf("清空后的值为 %q，期望空字符串", value)
	}
}

// openServiceTestDB 创建使用临时 SQLite 文件的服务层测试数据库。
func openServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "service.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}, &model.Subscription{}, &model.SubscriptionRenewal{}); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}
	return db
}
