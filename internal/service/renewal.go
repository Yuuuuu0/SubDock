package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"subdock/internal/model"
)

var (
	ErrSubscriptionNotFound = errors.New("订阅不存在")
	ErrRenewalConflict      = errors.New("订阅已被其他请求更新，请重试")
	ErrInvalidCycle         = errors.New("订阅周期配置无效")
)

// RenewalService 统一处理手动续订、自动续订和续订历史查询。
type RenewalService struct {
	db  *gorm.DB
	now func() time.Time
}

// NewRenewalService 创建使用指定数据库和系统时钟的续订服务。
func NewRenewalService(db *gorm.DB) *RenewalService {
	return NewRenewalServiceWithClock(db, time.Now)
}

// NewRenewalServiceWithClock 创建可注入时钟的续订服务，主要用于调度和测试。
func NewRenewalServiceWithClock(db *gorm.DB, now func() time.Time) *RenewalService {
	if now == nil {
		now = time.Now
	}
	return &RenewalService{db: db, now: now}
}

// Renew 手动续订一次；逾期订阅从当前自然日开始推进，返回更新后的订阅。
func (s *RenewalService) Renew(ctx context.Context, subscriptionID uint) (*model.Subscription, error) {
	updated, _, err := s.renew(ctx, subscriptionID, false)
	return updated, err
}

// AutoRenewIfDue 在订阅启用自动续订且已到期时续订一次，未到期时返回 renewed=false。
func (s *RenewalService) AutoRenewIfDue(ctx context.Context, subscriptionID uint) (*model.Subscription, bool, error) {
	return s.renew(ctx, subscriptionID, true)
}

// ListHistory 返回指定订阅的续订历史，最新记录排在前面。
func (s *RenewalService) ListHistory(ctx context.Context, subscriptionID uint) ([]model.SubscriptionRenewal, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.Subscription{}).Where("id = ?", subscriptionID).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("检查订阅失败: %w", err)
	}
	if count == 0 {
		return nil, ErrSubscriptionNotFound
	}

	var history []model.SubscriptionRenewal
	if err := s.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("renewed_at desc, id desc").
		Find(&history).Error; err != nil {
		return nil, fmt.Errorf("查询续订历史失败: %w", err)
	}
	return history, nil
}

// renew 在事务内执行共享续订逻辑，通过旧值条件更新避免 SQLite 并发丢更新。
func (s *RenewalService) renew(ctx context.Context, subscriptionID uint, onlyIfDue bool) (*model.Subscription, bool, error) {
	now := s.now()
	var updated model.Subscription
	renewed := false

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var subscription model.Subscription
		err := tx.First(&subscription, subscriptionID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSubscriptionNotFound
		}
		if err != nil {
			return fmt.Errorf("查询订阅失败: %w", err)
		}
		if err := subscription.ValidateCycle(); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidCycle, err)
		}

		today := model.DateOnlyIn(now, now.Location())
		expireDate := model.DateOnlyIn(subscription.ExpireDate, now.Location())
		if onlyIfDue && (!subscription.AutoRenew || expireDate.After(today)) {
			updated = subscription
			return nil
		}

		oldExpireDate := subscription.ExpireDate
		oldRenewCount := subscription.RenewCount
		newRenewCount := oldRenewCount + 1
		scheduledExpireDate := subscription.CalculateExpireDate()
		var newExpireDate time.Time
		if !expireDate.Before(today) && sameLogicalDate(oldExpireDate, scheduledExpireDate) {
			// 未逾期且仍使用计划到期日时，从开始日期和新续订次数重算，以保留月末账单日锚点。
			subscription.RenewCount = newRenewCount
			newExpireDate = subscription.CalculateExpireDate()
		} else {
			base := oldExpireDate
			if expireDate.Before(today) {
				base = time.Date(
					today.Year(),
					today.Month(),
					today.Day(),
					oldExpireDate.Hour(),
					oldExpireDate.Minute(),
					oldExpireDate.Second(),
					oldExpireDate.Nanosecond(),
					oldExpireDate.Location(),
				)
			}
			newExpireDate = subscription.CalculateExpireDateFrom(base)
		}

		result := tx.Model(&model.Subscription{}).
			Where("id = ? AND expire_date = ? AND renew_count = ?", subscription.ID, oldExpireDate, oldRenewCount).
			Updates(map[string]interface{}{
				"expire_date": newExpireDate,
				"renew_count": newRenewCount,
			})
		if result.Error != nil {
			return fmt.Errorf("更新订阅续订信息失败: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return ErrRenewalConflict
		}

		renewal := model.SubscriptionRenewal{
			SubscriptionID: subscription.ID,
			RenewedAt:      now,
			OldExpireDate:  oldExpireDate,
			NewExpireDate:  newExpireDate,
			RenewCount:     newRenewCount,
		}
		if err := tx.Create(&renewal).Error; err != nil {
			return fmt.Errorf("写入续订历史失败: %w", err)
		}

		subscription.ExpireDate = newExpireDate
		subscription.RenewCount = newRenewCount
		updated = subscription
		renewed = true
		return nil
	})
	if err != nil {
		if isSQLiteWriteConflict(err) {
			return nil, false, fmt.Errorf("%w: %v", ErrRenewalConflict, err)
		}
		return nil, false, err
	}
	return &updated, renewed, nil
}

// isSQLiteWriteConflict 识别 SQLite 并发写锁错误并将其转换为可重试冲突。
func isSQLiteWriteConflict(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "database is locked") || strings.Contains(message, "database table is locked")
}

// sameLogicalDate 判断两个持久化时间是否表示相同的年月日，不受时区偏移影响。
func sameLogicalDate(left, right time.Time) bool {
	return left.Year() == right.Year() && left.Month() == right.Month() && left.Day() == right.Day()
}
