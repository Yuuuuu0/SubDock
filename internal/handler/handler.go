package handler

import (
	"gorm.io/gorm"

	"subdock/internal/config"
	"subdock/internal/service"
)

// NotificationSender 定义 Handler 发送 Telegram 和 Bark 测试通知所需的最小能力。
type NotificationSender interface {
	SendTelegram(botToken, chatID, message string) error
	SendBark(barkURL, title, message string) error
}

// Handler 聚合 HTTP 处理器依赖，避免业务逻辑直接读取全局数据库。
type Handler struct {
	db       *gorm.DB
	config   *config.Config
	settings *service.SettingService
	renewals *service.RenewalService
	notifier NotificationSender
}

// New 创建默认 HTTP Handler，并复用统一的设置、续订和通知服务。
func New(db *gorm.DB, cfg *config.Config) *Handler {
	return NewWithNotifier(db, cfg, service.NewNotifier())
}

// NewWithNotifier 创建可注入通知发送器的 HTTP Handler，便于隔离外部网络测试。
func NewWithNotifier(db *gorm.DB, cfg *config.Config, notifier NotificationSender) *Handler {
	return &Handler{
		db:       db,
		config:   cfg,
		settings: service.NewSettingService(db),
		renewals: service.NewRenewalService(db),
		notifier: notifier,
	}
}
