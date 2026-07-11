package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"

	"subdock/internal/model"
	"subdock/internal/service"
)

// NotificationSender 定义调度器发送通知所需的最小渠道能力。
type NotificationSender interface {
	SendTelegram(botToken, chatID, message string) error
	SendBark(barkURL, title, message string) error
}

// Scheduler 每小时独立执行自动续订，并在配置时段发送到期提醒。
type Scheduler struct {
	cron     *cron.Cron
	db       *gorm.DB
	settings *service.SettingService
	renewals *service.RenewalService
	notifier NotificationSender
	now      func() time.Time
}

// New 创建使用默认通知客户端和系统时钟的调度器。
func New(db *gorm.DB) *Scheduler {
	return NewWithDependencies(db, service.NewNotifier(), time.Now)
}

// NewWithDependencies 创建可注入通知发送器和时钟的调度器，便于隔离测试。
func NewWithDependencies(db *gorm.DB, notifier NotificationSender, now func() time.Time) *Scheduler {
	if now == nil {
		now = time.Now
	}
	return &Scheduler{
		cron: cron.New(
			cron.WithLocation(time.Local),
			cron.WithChain(
				cron.Recover(cron.DefaultLogger),
				cron.SkipIfStillRunning(cron.DefaultLogger),
			),
		),
		db:       db,
		settings: service.NewSettingService(db),
		renewals: service.NewRenewalServiceWithClock(db, now),
		notifier: notifier,
		now:      now,
	}
}

// Start 注册每小时任务并启动调度器，注册失败时返回错误。
func (s *Scheduler) Start() error {
	if _, err := s.cron.AddFunc("0 * * * *", s.checkAndNotify); err != nil {
		return fmt.Errorf("注册订阅调度任务失败: %w", err)
	}
	s.cron.Start()
	log.Println("调度器已启动")
	return nil
}

// Stop 停止调度器并等待正在执行的任务退出。
func (s *Scheduler) Stop() {
	<-s.cron.Stop().Done()
}

// checkAndNotify 先处理全部到期自动续订，再根据当前小时决定是否发送提醒。
func (s *Scheduler) checkAndNotify() {
	ctx := context.Background()
	now := s.now()

	var subscriptions []model.Subscription
	if err := s.db.WithContext(ctx).Find(&subscriptions).Error; err != nil {
		log.Printf("获取订阅列表失败: %v", err)
		return
	}

	for index := range subscriptions {
		if !subscriptions[index].AutoRenew {
			continue
		}
		updated, renewed, err := s.renewals.AutoRenewIfDue(ctx, subscriptions[index].ID)
		if err != nil {
			log.Printf("自动续订失败(订阅ID=%d): %v", subscriptions[index].ID, err)
			continue
		}
		if renewed {
			subscriptions[index] = *updated
		}
	}

	settings, err := s.settings.GetNotificationSettings(ctx)
	if err != nil {
		log.Printf("读取通知设置失败: %v", err)
		return
	}
	hours, err := service.ParseNotifyHours(settings.NotifyHours)
	if err != nil {
		log.Printf("通知时段配置无效: %v", err)
		return
	}
	if !containsHour(hours, now.Hour()) {
		return
	}

	for _, subscription := range subscriptions {
		if subscription.ShouldRemindAt(now) {
			s.sendNotification(subscription, settings, now)
		}
	}
}

// sendNotification 将一条订阅提醒发送到全部已配置渠道。
func (s *Scheduler) sendNotification(subscription model.Subscription, settings service.NotificationSettings, now time.Time) {
	today := model.DateOnlyIn(now, now.Location())
	expireDate := model.DateOnlyIn(subscription.ExpireDate, now.Location())
	daysLeft := int(expireDate.Sub(today).Hours() / 24)
	message := fmt.Sprintf(
		"📢 订阅到期提醒\n\n订阅名称: %s\n金额: %.2f %s\n到期日期: %s\n剩余天数: %d 天",
		subscription.Name,
		subscription.Amount,
		subscription.Currency,
		subscription.ExpireDate.Format("2006-01-02"),
		daysLeft,
	)

	if settings.TelegramBotToken != "" && settings.TelegramChatID != "" {
		if err := s.notifier.SendTelegram(settings.TelegramBotToken, settings.TelegramChatID, message); err != nil {
			log.Printf("发送 Telegram 通知失败(订阅ID=%d): %v", subscription.ID, err)
		}
	}
	if settings.BarkURL != "" {
		if err := s.notifier.SendBark(settings.BarkURL, "订阅到期提醒", message); err != nil {
			log.Printf("发送 Bark 通知失败(订阅ID=%d): %v", subscription.ID, err)
		}
	}
}

// containsHour 判断当前小时是否在已去重的通知时段中。
func containsHour(hours []int, currentHour int) bool {
	for _, hour := range hours {
		if hour == currentHour {
			return true
		}
	}
	return false
}
