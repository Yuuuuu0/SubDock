package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"subdock/internal/model"
)

// TestRenewUsesSharedOverduePolicyAndWritesHistory 验证逾期续订从当前自然日推进并原子记录历史。
func TestRenewUsesSharedOverduePolicyAndWritesHistory(t *testing.T) {
	db := openServiceTestDB(t)
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("加载时区失败: %v", err)
	}
	now := time.Date(2026, time.July, 12, 9, 0, 0, 0, location)
	subscription := model.Subscription{
		Name:       "测试订阅",
		Amount:     10,
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

	renewals := NewRenewalServiceWithClock(db, func() time.Time { return now })
	updated, err := renewals.Renew(context.Background(), subscription.ID)
	if err != nil {
		t.Fatalf("手动续订失败: %v", err)
	}
	wantExpireDate := time.Date(2026, time.August, 12, 0, 0, 0, 0, time.UTC)
	if !updated.ExpireDate.Equal(wantExpireDate) {
		t.Fatalf("续订到期日为 %s，期望 %s", updated.ExpireDate, wantExpireDate)
	}
	if updated.RenewCount != 1 {
		t.Fatalf("续订次数为 %d，期望 1", updated.RenewCount)
	}

	history, err := renewals.ListHistory(context.Background(), subscription.ID)
	if err != nil {
		t.Fatalf("查询续订历史失败: %v", err)
	}
	if len(history) != 1 || history[0].RenewCount != 1 {
		t.Fatalf("续订历史为 %+v，期望一条续订记录", history)
	}
}

// TestAutoRenewIfDueSkipsFutureSubscription 验证自动续订不会提前推进尚未到期的订阅。
func TestAutoRenewIfDueSkipsFutureSubscription(t *testing.T) {
	db := openServiceTestDB(t)
	now := time.Date(2026, time.July, 12, 9, 0, 0, 0, time.UTC)
	subscription := model.Subscription{
		Name:       "未来订阅",
		Currency:   "CNY",
		StartDate:  now,
		CycleValue: 1,
		CycleUnit:  model.CycleUnitMonth,
		ExpireDate: time.Date(2026, time.August, 12, 0, 0, 0, 0, time.UTC),
		AutoRenew:  true,
	}
	if err := db.Create(&subscription).Error; err != nil {
		t.Fatalf("创建测试订阅失败: %v", err)
	}

	renewals := NewRenewalServiceWithClock(db, func() time.Time { return now })
	_, renewed, err := renewals.AutoRenewIfDue(context.Background(), subscription.ID)
	if err != nil {
		t.Fatalf("检查自动续订失败: %v", err)
	}
	if renewed {
		t.Fatal("未到期订阅不应自动续订")
	}
}

// TestRenewKeepsMonthEndAnchorForScheduledExpiry 验证计划到期日续订会从原始开始日恢复月末锚点。
func TestRenewKeepsMonthEndAnchorForScheduledExpiry(t *testing.T) {
	db := openServiceTestDB(t)
	now := time.Date(2025, time.February, 10, 9, 0, 0, 0, time.UTC)
	subscription := model.Subscription{
		Name:       "月末订阅",
		Currency:   "CNY",
		StartDate:  time.Date(2025, time.January, 31, 0, 0, 0, 0, time.UTC),
		CycleValue: 1,
		CycleUnit:  model.CycleUnitMonth,
		ExpireDate: time.Date(2025, time.February, 28, 0, 0, 0, 0, time.UTC),
	}
	if err := db.Create(&subscription).Error; err != nil {
		t.Fatalf("创建月末订阅失败: %v", err)
	}

	updated, err := NewRenewalServiceWithClock(db, func() time.Time { return now }).Renew(context.Background(), subscription.ID)
	if err != nil {
		t.Fatalf("续订月末订阅失败: %v", err)
	}
	want := time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC)
	if !updated.ExpireDate.Equal(want) {
		t.Fatalf("续订到期日为 %s，期望恢复到 %s", updated.ExpireDate, want)
	}
}

// TestRenewAdvancesFromCustomExpiry 验证自定义到期日不会被强制改回开始日期锚点。
func TestRenewAdvancesFromCustomExpiry(t *testing.T) {
	db := openServiceTestDB(t)
	now := time.Date(2025, time.February, 10, 9, 0, 0, 0, time.UTC)
	subscription := model.Subscription{
		Name:       "自定义到期日",
		Currency:   "CNY",
		StartDate:  time.Date(2025, time.January, 31, 0, 0, 0, 0, time.UTC),
		CycleValue: 1,
		CycleUnit:  model.CycleUnitMonth,
		ExpireDate: time.Date(2025, time.February, 27, 0, 0, 0, 0, time.UTC),
	}
	if err := db.Create(&subscription).Error; err != nil {
		t.Fatalf("创建自定义到期日订阅失败: %v", err)
	}

	updated, err := NewRenewalServiceWithClock(db, func() time.Time { return now }).Renew(context.Background(), subscription.ID)
	if err != nil {
		t.Fatalf("续订自定义到期日订阅失败: %v", err)
	}
	want := time.Date(2025, time.March, 27, 0, 0, 0, 0, time.UTC)
	if !updated.ExpireDate.Equal(want) {
		t.Fatalf("续订到期日为 %s，期望 %s", updated.ExpireDate, want)
	}
}

// TestRenewRejectsInvalidCycle 验证历史异常周期不会产生到期日不前进的续订记录。
func TestRenewRejectsInvalidCycle(t *testing.T) {
	db := openServiceTestDB(t)
	subscription := model.Subscription{
		Name:       "异常订阅",
		Currency:   "CNY",
		StartDate:  time.Now(),
		CycleValue: 0,
		CycleUnit:  model.CycleUnitMonth,
		ExpireDate: time.Now(),
	}
	if err := db.Create(&subscription).Error; err != nil {
		t.Fatalf("创建测试订阅失败: %v", err)
	}
	if err := db.Model(&subscription).UpdateColumn("cycle_value", 0).Error; err != nil {
		t.Fatalf("写入异常周期值失败: %v", err)
	}

	_, err := NewRenewalService(db).Renew(context.Background(), subscription.ID)
	if !errors.Is(err, ErrInvalidCycle) {
		t.Fatalf("错误为 %v，期望 ErrInvalidCycle", err)
	}
}

// TestConcurrentRenewDoesNotLoseUpdate 验证并发续订只会成功提交完整事务或返回可重试冲突。
func TestConcurrentRenewDoesNotLoseUpdate(t *testing.T) {
	db := openServiceTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取测试数据库连接失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(5)
	now := time.Date(2026, time.July, 12, 9, 0, 0, 0, time.UTC)
	subscription := model.Subscription{
		Name:       "并发续订",
		Currency:   "CNY",
		StartDate:  now,
		CycleValue: 1,
		CycleUnit:  model.CycleUnitMonth,
		ExpireDate: time.Date(2026, time.August, 12, 0, 0, 0, 0, time.UTC),
	}
	if err := db.Create(&subscription).Error; err != nil {
		t.Fatalf("创建测试订阅失败: %v", err)
	}

	renewals := NewRenewalServiceWithClock(db, func() time.Time { return now })
	start := make(chan struct{})
	errorsCh := make(chan error, 2)
	var waitGroup sync.WaitGroup
	for range 2 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			_, err := renewals.Renew(context.Background(), subscription.ID)
			errorsCh <- err
		}()
	}
	close(start)
	waitGroup.Wait()
	close(errorsCh)

	successes := 0
	for err := range errorsCh {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, ErrRenewalConflict):
			// 并发冲突是明确且可重试的预期结果。
		default:
			t.Fatalf("并发续订返回非预期错误: %v", err)
		}
	}
	if successes == 0 {
		t.Fatal("并发续订至少应有一个请求成功")
	}

	var updated model.Subscription
	if err := db.First(&updated, subscription.ID).Error; err != nil {
		t.Fatalf("查询并发续订结果失败: %v", err)
	}
	var historyCount int64
	if err := db.Model(&model.SubscriptionRenewal{}).Where("subscription_id = ?", subscription.ID).Count(&historyCount).Error; err != nil {
		t.Fatalf("统计续订历史失败: %v", err)
	}
	if updated.RenewCount != successes || historyCount != int64(successes) {
		t.Fatalf("成功数=%d，订阅续订次数=%d，历史数=%d", successes, updated.RenewCount, historyCount)
	}
}
