package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Admin 管理员账号
type Admin struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Username     string         `gorm:"uniqueIndex;size:64;not null" json:"username"`
	PasswordHash string         `gorm:"size:256;not null" json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// CycleUnit 周期单位
type CycleUnit string

const (
	CycleUnitDay      CycleUnit = "day"
	CycleUnitMonth    CycleUnit = "month"
	CycleUnitQuarter  CycleUnit = "quarter"
	CycleUnitHalfYear CycleUnit = "half_year"
	CycleUnitYear     CycleUnit = "year"
)

// IsValid 判断周期单位是否为系统支持的值。
func (u CycleUnit) IsValid() bool {
	switch u {
	case CycleUnitDay, CycleUnitMonth, CycleUnitQuarter, CycleUnitHalfYear, CycleUnitYear:
		return true
	default:
		return false
	}
}

// ParseCycleUnit 解析并校验字符串周期单位。
func ParseCycleUnit(value string) (CycleUnit, error) {
	unit := CycleUnit(value)
	if !unit.IsValid() {
		return "", fmt.Errorf("不支持的周期单位: %s", value)
	}
	return unit, nil
}

// Subscription 订阅
type Subscription struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	Name       string         `gorm:"size:128;not null" json:"name"`
	Amount     float64        `gorm:"not null" json:"amount"`
	Currency   string         `gorm:"size:8;default:CNY" json:"currency"`
	StartDate  time.Time      `gorm:"not null" json:"start_date"`
	CycleValue int            `gorm:"not null;default:1" json:"cycle_value"`
	CycleUnit  CycleUnit      `gorm:"size:16;not null;default:month" json:"cycle_unit"`
	ExpireDate time.Time      `gorm:"not null" json:"expire_date"`
	AutoRenew  bool           `gorm:"not null;default:false" json:"auto_renew"`
	RenewCount int            `gorm:"not null;default:0" json:"renew_count"`
	RemindDays int            `gorm:"not null;default:3" json:"remind_days"`
	Remark     string         `gorm:"size:512" json:"remark"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// SubscriptionRenewal 订阅续订记录
type SubscriptionRenewal struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	SubscriptionID uint      `gorm:"index;not null" json:"subscription_id"`
	RenewedAt      time.Time `gorm:"not null" json:"renewed_at"`
	OldExpireDate  time.Time `gorm:"not null" json:"old_expire_date"`
	NewExpireDate  time.Time `gorm:"not null" json:"new_expire_date"`
	RenewCount     int       `gorm:"not null" json:"renew_count"`
}

// CalculateExpireDate 根据开始日期、周期和续订次数计算到期日期，月末日期会钳制到目标月最后一天。
func (s *Subscription) CalculateExpireDate() time.Time {
	totalCycles := s.RenewCount + 1
	return s.calculateExpireDateFrom(s.StartDate, totalCycles)
}

// CalculateExpireDateFrom 根据给定基准日期推进一个订阅周期，并正确处理月末和闰年。
func (s *Subscription) CalculateExpireDateFrom(base time.Time) time.Time {
	return s.calculateExpireDateFrom(base, 1)
}

// ValidateCycle 校验订阅周期值和单位，避免无效周期产生不前进的续订。
func (s *Subscription) ValidateCycle() error {
	if s.CycleValue <= 0 {
		return fmt.Errorf("周期数值必须大于 0")
	}
	if !s.CycleUnit.IsValid() {
		return fmt.Errorf("不支持的周期单位: %s", s.CycleUnit)
	}
	return nil
}

// calculateExpireDateFrom 从基准日期推进指定周期数，月份类周期保持原始账单日锚点。
func (s *Subscription) calculateExpireDateFrom(base time.Time, cycles int) time.Time {
	if cycles <= 0 {
		return base
	}
	switch s.CycleUnit {
	case CycleUnitDay:
		return base.AddDate(0, 0, s.CycleValue*cycles)
	case CycleUnitMonth:
		return addMonthsClamped(base, s.CycleValue*cycles)
	case CycleUnitQuarter:
		return addMonthsClamped(base, s.CycleValue*3*cycles)
	case CycleUnitHalfYear:
		return addMonthsClamped(base, s.CycleValue*6*cycles)
	case CycleUnitYear:
		return addMonthsClamped(base, s.CycleValue*12*cycles)
	default:
		return base
	}
}

// addMonthsClamped 按月份推进日期，并将超出目标月的日期钳制到月末。
func addMonthsClamped(base time.Time, months int) time.Time {
	targetFirstDay := time.Date(
		base.Year(),
		base.Month()+time.Month(months),
		1,
		base.Hour(),
		base.Minute(),
		base.Second(),
		base.Nanosecond(),
		base.Location(),
	)
	lastDay := time.Date(
		targetFirstDay.Year(),
		targetFirstDay.Month()+1,
		0,
		base.Hour(),
		base.Minute(),
		base.Second(),
		base.Nanosecond(),
		base.Location(),
	).Day()
	day := base.Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(
		targetFirstDay.Year(),
		targetFirstDay.Month(),
		day,
		base.Hour(),
		base.Minute(),
		base.Second(),
		base.Nanosecond(),
		base.Location(),
	)
}

// DateOnlyIn 将时间的年月日解释为指定时区中的本地自然日零点。
func DateOnlyIn(value time.Time, location *time.Location) time.Time {
	if location == nil {
		location = time.Local
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, location)
}

// ShouldRemindAt 判断指定时刻所在自然日是否落在订阅提醒窗口内。
func (s *Subscription) ShouldRemindAt(now time.Time) bool {
	today := DateOnlyIn(now, now.Location())
	expireDate := DateOnlyIn(s.ExpireDate, now.Location())
	remindDate := expireDate.AddDate(0, 0, -s.RemindDays)
	return !today.Before(remindDate) && !today.After(expireDate)
}

// ShouldRemindToday 判断当前本地自然日是否应该提醒。
func (s *Subscription) ShouldRemindToday() bool {
	return s.ShouldRemindAt(time.Now())
}

// Setting 系统设置
type Setting struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	Key   string `gorm:"uniqueIndex;size:64;not null" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}
