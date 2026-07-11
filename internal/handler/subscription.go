package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"subdock/internal/model"
	"subdock/internal/service"
)

const dateLayout = "2006-01-02"

// CreateSubscriptionRequest 表示创建订阅时允许提交的字段。
type CreateSubscriptionRequest struct {
	Name       string  `json:"name" binding:"required"`
	Amount     float64 `json:"amount" binding:"gte=0"`
	Currency   string  `json:"currency"`
	StartDate  string  `json:"start_date" binding:"required"`
	CycleValue int     `json:"cycle_value" binding:"required,gt=0"`
	CycleUnit  string  `json:"cycle_unit" binding:"required"`
	ExpireDate string  `json:"expire_date"`
	AutoRenew  bool    `json:"auto_renew"`
	RemindDays *int    `json:"remind_days" binding:"omitempty,gte=0"`
	Remark     string  `json:"remark"`
}

// UpdateSubscriptionRequest 使用指针保留部分更新语义，并支持清空备注和将提醒天数设为零。
type UpdateSubscriptionRequest struct {
	Name       *string  `json:"name"`
	Amount     *float64 `json:"amount"`
	Currency   *string  `json:"currency"`
	StartDate  *string  `json:"start_date"`
	CycleValue *int     `json:"cycle_value"`
	CycleUnit  *string  `json:"cycle_unit"`
	ExpireDate *string  `json:"expire_date"`
	AutoRenew  *bool    `json:"auto_renew"`
	RemindDays *int     `json:"remind_days"`
	Remark     *string  `json:"remark"`
}

// ListSubscriptions 返回按到期日期升序排列的订阅列表。
func (h *Handler) ListSubscriptions(c *gin.Context) {
	var subscriptions []model.Subscription
	if err := h.db.WithContext(c.Request.Context()).Order("expire_date asc").Find(&subscriptions).Error; err != nil {
		log.Printf("获取订阅列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订阅列表失败"})
		return
	}
	c.JSON(http.StatusOK, subscriptions)
}

// GetSubscription 返回指定 ID 的单个订阅。
func (h *Handler) GetSubscription(c *gin.Context) {
	id, ok := parseSubscriptionID(c)
	if !ok {
		return
	}

	subscription, err := h.findSubscription(c, id)
	if err != nil {
		writeSubscriptionLookupError(c, err)
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// CreateSubscription 校验输入、计算到期日期并创建订阅。
func (h *Handler) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	name := strings.TrimSpace(req.Name)
	if err := validateTextLength("订阅名称", name, 1, 128); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateTextLength("备注", req.Remark, 0, 512); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式错误，应为 YYYY-MM-DD"})
		return
	}
	cycleUnit, err := model.ParseCycleUnit(req.CycleUnit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	currency := strings.TrimSpace(req.Currency)
	if currency == "" {
		currency = "CNY"
	}
	if err := validateTextLength("币种", currency, 1, 8); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	remindDays := 3
	if req.RemindDays != nil {
		remindDays = *req.RemindDays
	}

	subscription := &model.Subscription{
		Name:       name,
		Amount:     req.Amount,
		Currency:   currency,
		StartDate:  startDate,
		CycleValue: req.CycleValue,
		CycleUnit:  cycleUnit,
		AutoRenew:  req.AutoRenew,
		RemindDays: remindDays,
		Remark:     req.Remark,
	}
	if req.ExpireDate != "" {
		expireDate, err := parseDate(req.ExpireDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "到期日期格式错误，应为 YYYY-MM-DD"})
			return
		}
		subscription.ExpireDate = expireDate
	} else {
		subscription.ExpireDate = subscription.CalculateExpireDate()
	}

	if err := h.db.WithContext(c.Request.Context()).Create(subscription).Error; err != nil {
		log.Printf("创建订阅失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订阅失败"})
		return
	}
	c.JSON(http.StatusCreated, subscription)
}

// UpdateSubscription 更新请求中出现的字段，周期变更时自动重算到期日，显式到期日优先。
func (h *Handler) UpdateSubscription(c *gin.Context) {
	id, ok := parseSubscriptionID(c)
	if !ok {
		return
	}
	subscription, err := h.findSubscription(c, id)
	if err != nil {
		writeSubscriptionLookupError(c, err)
		return
	}

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	updates := make(map[string]interface{})
	cycleRelatedChanged := false
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if err := validateTextLength("订阅名称", name, 1, 128); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updates["name"] = name
	}
	if req.Amount != nil {
		if *req.Amount < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "金额不能小于 0"})
			return
		}
		updates["amount"] = *req.Amount
	}
	if req.Currency != nil {
		currency := strings.TrimSpace(*req.Currency)
		if err := validateTextLength("币种", currency, 1, 8); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updates["currency"] = currency
	}
	if req.StartDate != nil {
		startDate, err := parseDate(*req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式错误，应为 YYYY-MM-DD"})
			return
		}
		updates["start_date"] = startDate
		subscription.StartDate = startDate
		cycleRelatedChanged = true
	}
	if req.CycleValue != nil {
		if *req.CycleValue <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "周期数值必须大于 0"})
			return
		}
		updates["cycle_value"] = *req.CycleValue
		subscription.CycleValue = *req.CycleValue
		cycleRelatedChanged = true
	}
	if req.CycleUnit != nil {
		cycleUnit, err := model.ParseCycleUnit(*req.CycleUnit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updates["cycle_unit"] = cycleUnit
		subscription.CycleUnit = cycleUnit
		cycleRelatedChanged = true
	}
	if req.AutoRenew != nil {
		updates["auto_renew"] = *req.AutoRenew
	}
	if req.RemindDays != nil {
		if *req.RemindDays < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "提醒天数不能小于 0"})
			return
		}
		updates["remind_days"] = *req.RemindDays
	}
	if req.Remark != nil {
		if err := validateTextLength("备注", *req.Remark, 0, 512); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updates["remark"] = *req.Remark
	}

	explicitExpireDate := false
	if req.ExpireDate != nil {
		if strings.TrimSpace(*req.ExpireDate) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "到期日期不能为空"})
			return
		}
		expireDate, err := parseDate(*req.ExpireDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "到期日期格式错误，应为 YYYY-MM-DD"})
			return
		}
		updates["expire_date"] = expireDate
		explicitExpireDate = true
	}
	if cycleRelatedChanged && !explicitExpireDate {
		if err := subscription.ValidateCycle(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updates["expire_date"] = subscription.CalculateExpireDate()
	}

	if len(updates) > 0 {
		if err := h.db.WithContext(c.Request.Context()).Model(subscription).Updates(updates).Error; err != nil {
			log.Printf("更新订阅失败(ID=%d): %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新订阅失败"})
			return
		}
	}
	if err := h.db.WithContext(c.Request.Context()).First(subscription, id).Error; err != nil {
		log.Printf("刷新更新后的订阅失败(ID=%d): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取更新结果失败"})
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// RenewSubscription 手动续订一次并返回更新后的订阅。
func (h *Handler) RenewSubscription(c *gin.Context) {
	id, ok := parseSubscriptionID(c)
	if !ok {
		return
	}
	subscription, err := h.renewals.Renew(c.Request.Context(), id)
	if err != nil {
		writeRenewalError(c, err)
		return
	}
	c.JSON(http.StatusOK, subscription)
}

// ListSubscriptionRenewals 返回指定订阅的全部续订历史。
func (h *Handler) ListSubscriptionRenewals(c *gin.Context) {
	id, ok := parseSubscriptionID(c)
	if !ok {
		return
	}
	history, err := h.renewals.ListHistory(c.Request.Context(), id)
	if err != nil {
		writeRenewalError(c, err)
		return
	}
	c.JSON(http.StatusOK, history)
}

// DeleteSubscription 软删除指定订阅，不存在时返回 404。
func (h *Handler) DeleteSubscription(c *gin.Context) {
	id, ok := parseSubscriptionID(c)
	if !ok {
		return
	}
	result := h.db.WithContext(c.Request.Context()).Delete(&model.Subscription{}, id)
	if result.Error != nil {
		log.Printf("删除订阅失败(ID=%d): %v", id, result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除订阅失败"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "订阅不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// TestSubscriptionNotify 使用全部已配置渠道发送指定订阅的测试提醒。
func (h *Handler) TestSubscriptionNotify(c *gin.Context) {
	id, ok := parseSubscriptionID(c)
	if !ok {
		return
	}
	subscription, err := h.findSubscription(c, id)
	if err != nil {
		writeSubscriptionLookupError(c, err)
		return
	}
	settings, err := h.settings.GetNotificationSettings(c.Request.Context())
	if err != nil {
		log.Printf("读取订阅测试通知设置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取通知配置失败"})
		return
	}

	message := formatSubscriptionNotification(subscription)
	sent := false
	if settings.TelegramBotToken != "" && settings.TelegramChatID != "" {
		if err := h.notifier.SendTelegram(settings.TelegramBotToken, settings.TelegramChatID, message); err != nil {
			log.Printf("发送 Telegram 订阅测试通知失败: %v", err)
		} else {
			sent = true
		}
	}
	if settings.BarkURL != "" {
		if err := h.notifier.SendBark(settings.BarkURL, "SubDock 订阅提醒", message); err != nil {
			log.Printf("发送 Bark 订阅测试通知失败: %v", err)
		} else {
			sent = true
		}
	}
	if !sent {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未配置可用通知渠道或发送失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "通知发送成功"})
}

// findSubscription 查询订阅并保留可区分的数据库错误。
func (h *Handler) findSubscription(c *gin.Context, id uint) (*model.Subscription, error) {
	var subscription model.Subscription
	if err := h.db.WithContext(c.Request.Context()).First(&subscription, id).Error; err != nil {
		return nil, err
	}
	return &subscription, nil
}

// parseSubscriptionID 解析路由中的订阅 ID，失败时直接写入 400 响应。
func parseSubscriptionID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 ID"})
		return 0, false
	}
	return uint(id), true
}

// parseDate 按稳定的日期格式解析逻辑日期，保持现有数据库和 JSON 兼容。
func parseDate(value string) (time.Time, error) {
	return time.Parse(dateLayout, value)
}

// validateTextLength 校验文本的 Unicode 字符数，弥补 SQLite 不强制 VARCHAR 长度的限制。
func validateTextLength(field, value string, minLength, maxLength int) error {
	length := utf8.RuneCountInString(value)
	if length < minLength {
		return errors.New(field + "不能为空")
	}
	if length > maxLength {
		return errors.New(field + "长度不能超过 " + strconv.Itoa(maxLength) + " 个字符")
	}
	return nil
}

// writeSubscriptionLookupError 将订阅查询错误映射为稳定的 HTTP 响应。
func writeSubscriptionLookupError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "订阅不存在"})
		return
	}
	log.Printf("查询订阅失败: %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "查询订阅失败"})
}

// writeRenewalError 将续订领域错误映射为对应 HTTP 状态码。
func writeRenewalError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSubscriptionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": service.ErrSubscriptionNotFound.Error()})
	case errors.Is(err, service.ErrRenewalConflict):
		c.JSON(http.StatusConflict, gin.H{"error": service.ErrRenewalConflict.Error()})
	case errors.Is(err, service.ErrInvalidCycle):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		log.Printf("处理续订失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "续订失败"})
	}
}

// formatSubscriptionNotification 格式化订阅测试通知消息。
func formatSubscriptionNotification(subscription *model.Subscription) string {
	return "📋 订阅提醒测试\n\n" +
		"名称：" + subscription.Name + "\n" +
		"金额：" + subscription.Currency + " " + formatFloat(subscription.Amount) + "\n" +
		"开始日期：" + subscription.StartDate.Format(dateLayout) + "\n" +
		"到期日期：" + subscription.ExpireDate.Format(dateLayout) + "\n" +
		"备注：" + subscription.Remark
}

// formatFloat 将金额格式化为两位小数。
func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}
