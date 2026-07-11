package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"subdock/internal/model"
)

const (
	SettingNotifyHours      = "notify_hours"
	SettingTelegramBotToken = "telegram_bot_token"
	SettingTelegramChatID   = "telegram_chat_id"
	SettingBarkURL          = "bark_url"
)

// NotificationSettings 表示持久化的通知渠道和发送时段配置。
type NotificationSettings struct {
	NotifyHours      string
	TelegramBotToken string
	TelegramChatID   string
	BarkURL          string
}

// SettingService 提供可复用且错误可感知的设置读写能力。
type SettingService struct {
	db *gorm.DB
}

// NewSettingService 创建使用指定数据库的设置服务。
func NewSettingService(db *gorm.DB) *SettingService {
	return &SettingService{db: db}
}

// Get 获取单项设置；键不存在时返回默认值，数据库故障会原样返回。
func (s *SettingService) Get(ctx context.Context, key, defaultValue string) (string, error) {
	var setting model.Setting
	err := s.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return defaultValue, nil
	}
	if err != nil {
		return "", fmt.Errorf("读取设置 %s 失败: %w", key, err)
	}
	return setting.Value, nil
}

// UpdateMany 在单个事务中新增或更新设置，空字符串会被正常持久化用于清除配置。
func (s *SettingService) UpdateMany(ctx context.Context, values map[string]string) error {
	if len(values) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for key, value := range values {
			setting := model.Setting{Key: key, Value: value}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.Assignments(map[string]interface{}{"value": value}),
			}).Create(&setting).Error; err != nil {
				return fmt.Errorf("保存设置 %s 失败: %w", key, err)
			}
		}
		return nil
	})
}

// GetNotificationSettings 一次读取全部通知配置，避免调度循环重复查询数据库。
func (s *SettingService) GetNotificationSettings(ctx context.Context) (NotificationSettings, error) {
	keys := []string{
		SettingNotifyHours,
		SettingTelegramBotToken,
		SettingTelegramChatID,
		SettingBarkURL,
	}
	var rows []model.Setting
	if err := s.db.WithContext(ctx).Where("key IN ?", keys).Find(&rows).Error; err != nil {
		return NotificationSettings{}, fmt.Errorf("读取通知设置失败: %w", err)
	}
	values := map[string]string{SettingNotifyHours: "9"}
	for _, row := range rows {
		values[row.Key] = row.Value
	}
	return NotificationSettings{
		NotifyHours:      values[SettingNotifyHours],
		TelegramBotToken: values[SettingTelegramBotToken],
		TelegramChatID:   values[SettingTelegramChatID],
		BarkURL:          values[SettingBarkURL],
	}, nil
}

// ParseNotifyHours 解析逗号分隔的 0 到 23 点配置，空字符串表示关闭定时通知。
func ParseNotifyHours(value string) ([]int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return []int{}, nil
	}

	seen := make(map[int]struct{})
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		hour, err := strconv.Atoi(part)
		if err != nil || hour < 0 || hour > 23 {
			return nil, fmt.Errorf("通知小时必须是 0 到 23 的整数: %s", part)
		}
		seen[hour] = struct{}{}
	}

	hours := make([]int, 0, len(seen))
	for hour := range seen {
		hours = append(hours, hour)
	}
	sort.Ints(hours)
	return hours, nil
}

// NormalizeNotifyHours 校验通知时段并返回去重、升序的持久化字符串。
func NormalizeNotifyHours(value string) (string, error) {
	hours, err := ParseNotifyHours(value)
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(hours))
	for _, hour := range hours {
		parts = append(parts, strconv.Itoa(hour))
	}
	return strings.Join(parts, ","), nil
}
