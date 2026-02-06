package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"subdock/internal/model"
	"subdock/internal/service"
)

// CreateSubscriptionRequest 创建订阅请求
type CreateSubscriptionRequest struct {
	Name       string  `json:"name" binding:"required"`
	Amount     float64 `json:"amount" binding:"gte=0"`
	Currency   string  `json:"currency"`
	StartDate  string  `json:"start_date" binding:"required"`
	CycleValue int     `json:"cycle_value" binding:"required,gt=0"`
	CycleUnit  string  `json:"cycle_unit" binding:"required,oneof=day month quarter half_year year"`
	ExpireDate string  `json:"expire_date"`
	AutoRenew  bool    `json:"auto_renew"`
	RemindDays int     `json:"remind_days"`
	Remark     string  `json:"remark"`
}

// UpdateSubscriptionRequest 更新订阅请求
type UpdateSubscriptionRequest struct {
	Name       string   `json:"name"`
	Amount     *float64 `json:"amount"`
	Currency   string   `json:"currency"`
	StartDate  string   `json:"start_date"`
	CycleValue int      `json:"cycle_value"`
	CycleUnit  string   `json:"cycle_unit"`
	ExpireDate string   `json:"expire_date"`
	AutoRenew  *bool    `json:"auto_renew"`
	RemindDays int      `json:"remind_days"`
	Remark     string   `json:"remark"`
}

// ListSubscriptions 获取订阅列表
func ListSubscriptions(c *gin.Context) {
	var subscriptions []model.Subscription
	if err := model.GetDB().Order("expire_date asc").Find(&subscriptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订阅列表失败"})
		return
	}
	c.JSON(http.StatusOK, subscriptions)
}

// GetSubscription 获取单个订阅
func GetSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	var subscription model.Subscription
	if err := model.GetDB().First(&subscription, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订阅不存在"})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// CreateSubscription 创建订阅
func CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式错误，应为 YYYY-MM-DD"})
		return
	}

	currency := req.Currency
	if currency == "" {
		currency = "CNY"
	}

	remindDays := req.RemindDays
	if remindDays <= 0 {
		remindDays = 3
	}

	subscription := &model.Subscription{
		Name:       req.Name,
		Amount:     req.Amount,
		Currency:   currency,
		StartDate:  startDate,
		CycleValue: req.CycleValue,
		CycleUnit:  model.CycleUnit(req.CycleUnit),
		AutoRenew:  req.AutoRenew,
		RemindDays: remindDays,
		Remark:     req.Remark,
	}

	// 计算到期日期
	if req.ExpireDate != "" {
		expireDate, err := time.Parse("2006-01-02", req.ExpireDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "到期日期格式错误，应为 YYYY-MM-DD"})
			return
		}
		subscription.ExpireDate = expireDate
	} else {
		subscription.ExpireDate = subscription.CalculateExpireDate()
	}

	if err := model.GetDB().Create(subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订阅失败"})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// UpdateSubscription 更新订阅
func UpdateSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	var subscription model.Subscription
	if err := model.GetDB().First(&subscription, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订阅不存在"})
		return
	}

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	updates := make(map[string]interface{})
	cycleRelatedChanged := false

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Amount != nil {
		if *req.Amount < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "金额不能小于 0"})
			return
		}
		updates["amount"] = *req.Amount
	}
	if req.Currency != "" {
		updates["currency"] = req.Currency
	}
	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式错误"})
			return
		}
		updates["start_date"] = startDate
		subscription.StartDate = startDate
		cycleRelatedChanged = true
	}
	if req.CycleValue > 0 {
		updates["cycle_value"] = req.CycleValue
		subscription.CycleValue = req.CycleValue
		cycleRelatedChanged = true
	}
	if req.CycleUnit != "" {
		updates["cycle_unit"] = req.CycleUnit
		subscription.CycleUnit = model.CycleUnit(req.CycleUnit)
		cycleRelatedChanged = true
	}
	if req.AutoRenew != nil {
		updates["auto_renew"] = *req.AutoRenew
	}
	if req.RemindDays > 0 {
		updates["remind_days"] = req.RemindDays
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}

	// 处理到期日期
	if req.ExpireDate != "" {
		expireDate, err := time.Parse("2006-01-02", req.ExpireDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "到期日期格式错误"})
			return
		}
		if cycleRelatedChanged && expireDate.Equal(subscription.ExpireDate) {
			updates["expire_date"] = subscription.CalculateExpireDate()
		} else {
			updates["expire_date"] = expireDate
		}
	} else if cycleRelatedChanged {
		// 如果修改了开始日期或周期，重新计算到期日期
		updates["expire_date"] = subscription.CalculateExpireDate()
	}

	if err := model.GetDB().Model(&subscription).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新订阅失败"})
		return
	}

	model.GetDB().First(&subscription, id)
	c.JSON(http.StatusOK, subscription)
}

// RenewSubscription 手动续订一次
func RenewSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	tx := model.GetDB().Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "续订失败"})
		return
	}

	var subscription model.Subscription
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&subscription, id).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "订阅不存在"})
		return
	}

	oldExpireDate := subscription.ExpireDate
	base := subscription.ExpireDate
	if subscription.CycleValue <= 0 {
		subscription.CycleValue = 1
	}
	newExpireDate := subscription.CalculateExpireDateFrom(base)
	newRenewCount := subscription.RenewCount + 1

	if err := tx.Model(&subscription).Updates(map[string]interface{}{
		"expire_date": newExpireDate,
		"renew_count": newRenewCount,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "续订失败"})
		return
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "续订记录写入失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "续订失败"})
		return
	}

	model.GetDB().First(&subscription, id)
	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription 删除订阅
func DeleteSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	if err := model.GetDB().Delete(&model.Subscription{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除订阅失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// TestSubscriptionNotify 测试订阅通知
func TestSubscriptionNotify(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return
	}

	var subscription model.Subscription
	if err := model.GetDB().First(&subscription, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订阅不存在"})
		return
	}

	var telegramBotToken, telegramChatID, barkURL string
	model.GetDB().Model(&model.Setting{}).Where("key = ?", "telegram_bot_token").Pluck("value", &telegramBotToken)
	model.GetDB().Model(&model.Setting{}).Where("key = ?", "telegram_chat_id").Pluck("value", &telegramChatID)
	model.GetDB().Model(&model.Setting{}).Where("key = ?", "bark_url").Pluck("value", &barkURL)

	msg := formatSubscriptionNotification(&subscription)

	notifier := service.NewNotifier()
	var sent bool
	var errMsg string

	if telegramBotToken != "" && telegramChatID != "" {
		if err := notifier.SendTelegram(telegramBotToken, telegramChatID, msg); err != nil {
			errMsg += "Telegram: " + err.Error() + "; "
		} else {
			sent = true
		}
	}

	if barkURL != "" {
		if err := notifier.SendBark(barkURL, "SubDock 订阅提醒", msg); err != nil {
			errMsg += "Bark: " + err.Error()
		} else {
			sent = true
		}
	}

	if !sent {
		if errMsg == "" {
			errMsg = "未配置任何通知渠道"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "通知发送成功"})
}

// formatSubscriptionNotification 格式化订阅通知消息
func formatSubscriptionNotification(sub *model.Subscription) string {
	return "📋 订阅提醒测试\n\n" +
		"名称：" + sub.Name + "\n" +
		"金额：" + sub.Currency + " " + formatFloat(sub.Amount) + "\n" +
		"开始日期：" + sub.StartDate.Format("2006-01-02") + "\n" +
		"到期日期：" + sub.ExpireDate.Format("2006-01-02") + "\n" +
		"备注：" + sub.Remark
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
