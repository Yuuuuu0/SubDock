package scheduler

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm/clause"

	"subdock/internal/model"
	"subdock/internal/service"
)

// Scheduler 定时任务调度器
type Scheduler struct {
	cron     *cron.Cron
	notifier *service.Notifier
}

// New 创建调度器
func New() *Scheduler {
	return &Scheduler{
		cron:     cron.New(),
		notifier: service.NewNotifier(),
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	// 每小时检查一次
	s.cron.AddFunc("0 * * * *", s.checkAndNotify)
	s.cron.Start()
	log.Println("调度器已启动")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// checkAndNotify 检查并发送到期提醒
func (s *Scheduler) checkAndNotify() {
	currentHour := time.Now().Hour()

	// 获取通知时段配置
	notifyHours := getSetting("notify_hours", "9")
	hours := parseNotifyHours(notifyHours)

	// 检查当前小时是否在通知时段内
	shouldNotify := false
	for _, h := range hours {
		if h == currentHour {
			shouldNotify = true
			break
		}
	}

	if !shouldNotify {
		return
	}

	// 获取需要提醒的订阅
	var subscriptions []model.Subscription
	if err := model.GetDB().Find(&subscriptions).Error; err != nil {
		log.Printf("获取订阅列表失败: %v", err)
		return
	}

	for _, sub := range subscriptions {
		if sub.AutoRenew {
			renewed, err := s.autoRenewIfNeeded(sub.ID)
			if err != nil {
				log.Printf("自动续订失败(订阅ID=%d): %v", sub.ID, err)
			} else if renewed {
				if err := model.GetDB().First(&sub, sub.ID).Error; err != nil {
					log.Printf("自动续订后刷新订阅失败(订阅ID=%d): %v", sub.ID, err)
				}
			}
		}

		if sub.ShouldRemindToday() {
			s.sendNotification(sub)
		}
	}
}

// autoRenewIfNeeded 当启用自动续订且已到期时自动续订 1 次
func (s *Scheduler) autoRenewIfNeeded(subscriptionID uint) (bool, error) {
	tx := model.GetDB().Begin()
	if tx.Error != nil {
		return false, tx.Error
	}

	var subscription model.Subscription
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&subscription, subscriptionID).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	expire := subscription.ExpireDate.Truncate(24 * time.Hour)
	if !subscription.AutoRenew || expire.After(today) {
		tx.Rollback()
		return false, nil
	}

	oldExpireDate := subscription.ExpireDate
	base := subscription.ExpireDate
	if base.Before(today) {
		base = today
	}
	newExpireDate := subscription.CalculateExpireDateFrom(base)
	newRenewCount := subscription.RenewCount + 1

	if err := tx.Model(&subscription).Updates(map[string]interface{}{
		"expire_date": newExpireDate,
		"renew_count": newRenewCount,
	}).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	renewal := &model.SubscriptionRenewal{
		SubscriptionID: subscription.ID,
		RenewedAt:      time.Now(),
		OldExpireDate:  oldExpireDate,
		NewExpireDate:  newExpireDate,
		RenewCount:     newRenewCount,
	}
	if err := tx.Create(renewal).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	if err := tx.Commit().Error; err != nil {
		return false, err
	}

	return true, nil
}

// sendNotification 发送订阅到期提醒
func (s *Scheduler) sendNotification(sub model.Subscription) {
	daysLeft := int(time.Until(sub.ExpireDate).Hours() / 24)
	message := fmt.Sprintf("📢 订阅到期提醒\n\n订阅名称: %s\n金额: %.2f %s\n到期日期: %s\n剩余天数: %d 天",
		sub.Name, sub.Amount, sub.Currency, sub.ExpireDate.Format("2006-01-02"), daysLeft)

	// 尝试 Telegram 通知
	telegramToken := getSetting("telegram_bot_token", "")
	telegramChatID := getSetting("telegram_chat_id", "")
	if telegramToken != "" && telegramChatID != "" {
		if err := s.notifier.SendTelegram(telegramToken, telegramChatID, message); err != nil {
			log.Printf("发送 Telegram 通知失败: %v", err)
		}
	}

	// 尝试 Bark 通知
	barkURL := getSetting("bark_url", "")
	if barkURL != "" {
		if err := s.notifier.SendBark(barkURL, "订阅到期提醒", message); err != nil {
			log.Printf("发送 Bark 通知失败: %v", err)
		}
	}
}

// parseNotifyHours 解析通知时段配置
func parseNotifyHours(s string) []int {
	var hours []int
	parts := strings.Split(s, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if h, err := strconv.Atoi(p); err == nil && h >= 1 && h <= 24 {
			hours = append(hours, h%24)
		}
	}
	if len(hours) == 0 {
		hours = []int{9}
	}
	return hours
}

func getSetting(key, defaultVal string) string {
	var setting model.Setting
	if err := model.GetDB().Where("key = ?", key).First(&setting).Error; err != nil {
		return defaultVal
	}
	return setting.Value
}
