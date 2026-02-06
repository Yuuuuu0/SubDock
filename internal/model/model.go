package model

import (
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
	CycleUnitDay     CycleUnit = "day"
	CycleUnitMonth   CycleUnit = "month"
	CycleUnitQuarter CycleUnit = "quarter"
	CycleUnitHalfYear CycleUnit = "half_year"
	CycleUnitYear    CycleUnit = "year"
)

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
	RemindDays int            `gorm:"not null;default:3" json:"remind_days"`
	Remark     string         `gorm:"size:512" json:"remark"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// CalculateExpireDate 根据开始日期和周期计算到期日期
func (s *Subscription) CalculateExpireDate() time.Time {
	switch s.CycleUnit {
	case CycleUnitDay:
		return s.StartDate.AddDate(0, 0, s.CycleValue)
	case CycleUnitMonth:
		return s.StartDate.AddDate(0, s.CycleValue, 0)
	case CycleUnitQuarter:
		return s.StartDate.AddDate(0, s.CycleValue*3, 0)
	case CycleUnitHalfYear:
		return s.StartDate.AddDate(0, s.CycleValue*6, 0)
	case CycleUnitYear:
		return s.StartDate.AddDate(s.CycleValue, 0, 0)
	default:
		return s.StartDate.AddDate(0, s.CycleValue, 0)
	}
}

// ShouldRemindToday 判断今天是否应该提醒
func (s *Subscription) ShouldRemindToday() bool {
	today := time.Now().Truncate(24 * time.Hour)
	remindDate := s.ExpireDate.AddDate(0, 0, -s.RemindDays).Truncate(24 * time.Hour)
	expireDate := s.ExpireDate.Truncate(24 * time.Hour)
	return !today.Before(remindDate) && !today.After(expireDate)
}

// Setting 系统设置
type Setting struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	Key   string `gorm:"uniqueIndex;size:64;not null" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}
