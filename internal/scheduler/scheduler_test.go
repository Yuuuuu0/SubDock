package scheduler

import (
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"subdock/internal/model"
)

// countingNotifier 记录调度测试中实际触发的通知次数。
type countingNotifier struct {
	count int
}

// SendTelegram 记录一次 Telegram 通知。
func (n *countingNotifier) SendTelegram(_, _, _ string) error {
	n.count++
	return nil
}

// SendBark 记录一次 Bark 通知。
func (n *countingNotifier) SendBark(_, _, _ string) error {
	n.count++
	return nil
}

// TestCheckAndNotifyRenewsOutsideNotifyHours 验证自动续订不再依赖通知小时。
func TestCheckAndNotifyRenewsOutsideNotifyHours(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "scheduler.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.Subscription{}, &model.SubscriptionRenewal{}, &model.Setting{}); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}
	now := time.Date(2026, time.July, 12, 5, 0, 0, 0, time.UTC)
	subscription := model.Subscription{
		Name:       "自动续订",
		Currency:   "CNY",
		StartDate:  time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
		CycleValue: 1,
		CycleUnit:  model.CycleUnitMonth,
		ExpireDate: time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
		AutoRenew:  true,
		RemindDays: 3,
	}
	if err := db.Create(&subscription).Error; err != nil {
		t.Fatalf("创建测试订阅失败: %v", err)
	}

	notifier := &countingNotifier{}
	scheduler := NewWithDependencies(db, notifier, func() time.Time { return now })
	scheduler.checkAndNotify()

	var updated model.Subscription
	if err := db.First(&updated, subscription.ID).Error; err != nil {
		t.Fatalf("查询自动续订结果失败: %v", err)
	}
	if updated.RenewCount != 1 {
		t.Fatalf("非通知时段续订次数为 %d，期望 1", updated.RenewCount)
	}
	if notifier.count != 0 {
		t.Fatalf("非通知时段发送了 %d 条通知，期望 0", notifier.count)
	}
}
