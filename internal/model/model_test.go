package model

import (
	"testing"
	"time"
)

// TestCalculateExpireDateClampsMonthEnd 验证月末和闰年账期不会跨入错误月份。
func TestCalculateExpireDateClampsMonthEnd(t *testing.T) {
	tests := []struct {
		name  string
		start time.Time
		unit  CycleUnit
		want  time.Time
	}{
		{
			name:  "平年一月末",
			start: time.Date(2025, time.January, 31, 0, 0, 0, 0, time.UTC),
			unit:  CycleUnitMonth,
			want:  time.Date(2025, time.February, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "闰年一月末",
			start: time.Date(2024, time.January, 31, 0, 0, 0, 0, time.UTC),
			unit:  CycleUnitMonth,
			want:  time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "闰日按年续订",
			start: time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC),
			unit:  CycleUnitYear,
			want:  time.Date(2025, time.February, 28, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			subscription := Subscription{StartDate: test.start, CycleValue: 1, CycleUnit: test.unit}
			if got := subscription.CalculateExpireDate(); !got.Equal(test.want) {
				t.Fatalf("到期日为 %s，期望 %s", got, test.want)
			}
		})
	}
}

// TestCalculateExpireDateKeepsOriginalBillingAnchor 验证多周期计算仍以原始账单日为锚点。
func TestCalculateExpireDateKeepsOriginalBillingAnchor(t *testing.T) {
	subscription := Subscription{
		StartDate:  time.Date(2025, time.January, 31, 0, 0, 0, 0, time.UTC),
		CycleValue: 1,
		CycleUnit:  CycleUnitMonth,
		RenewCount: 1,
	}
	want := time.Date(2025, time.March, 31, 0, 0, 0, 0, time.UTC)
	if got := subscription.CalculateExpireDate(); !got.Equal(want) {
		t.Fatalf("两周期到期日为 %s，期望 %s", got, want)
	}
}

// TestShouldRemindAtUsesLocalCalendarDay 验证提醒判断按本地自然日而非 24 小时截断计算。
func TestShouldRemindAtUsesLocalCalendarDay(t *testing.T) {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("加载时区失败: %v", err)
	}
	subscription := Subscription{
		ExpireDate: time.Date(2026, time.July, 13, 0, 0, 0, 0, time.UTC),
		RemindDays: 0,
	}

	if subscription.ShouldRemindAt(time.Date(2026, time.July, 12, 23, 59, 0, 0, location)) {
		t.Fatal("到期日前一天不应提醒")
	}
	if !subscription.ShouldRemindAt(time.Date(2026, time.July, 13, 0, 1, 0, 0, location)) {
		t.Fatal("到期日凌晨应提醒")
	}
	if subscription.ShouldRemindAt(time.Date(2026, time.July, 14, 0, 1, 0, 0, location)) {
		t.Fatal("到期日之后不应提醒")
	}
}

// TestParseCycleUnitRejectsInvalidValue 验证非法周期单位不会静默回退为月。
func TestParseCycleUnitRejectsInvalidValue(t *testing.T) {
	if _, err := ParseCycleUnit("weekly"); err == nil {
		t.Fatal("非法周期单位应返回错误")
	}
}
