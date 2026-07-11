package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"subdock/internal/service"
)

// SettingsResponse 表示前端可编辑的通知设置。
type SettingsResponse struct {
	NotifyHours      string `json:"notify_hours"`
	TelegramBotToken string `json:"telegram_bot_token"`
	TelegramChatID   string `json:"telegram_chat_id"`
	BarkURL          string `json:"bark_url"`
}

// UpdateSettingsRequest 使用指针区分未提交字段和显式清空字段。
type UpdateSettingsRequest struct {
	NotifyHours      *string `json:"notify_hours"`
	TelegramBotToken *string `json:"telegram_bot_token"`
	TelegramChatID   *string `json:"telegram_chat_id"`
	BarkURL          *string `json:"bark_url"`
}

// TestNotifyRequest 表示指定通知渠道的测试请求。
type TestNotifyRequest struct {
	Type string `json:"type" binding:"required,oneof=telegram bark"`
}

// GetSettings 返回全部通知设置，并区分配置缺失与数据库故障。
func (h *Handler) GetSettings(c *gin.Context) {
	settings, err := h.settings.GetNotificationSettings(c.Request.Context())
	if err != nil {
		log.Printf("读取通知设置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取设置失败"})
		return
	}
	c.JSON(http.StatusOK, SettingsResponse{
		NotifyHours:      settings.NotifyHours,
		TelegramBotToken: settings.TelegramBotToken,
		TelegramChatID:   settings.TelegramChatID,
		BarkURL:          settings.BarkURL,
	})
}

// UpdateSettings 原子更新请求中出现的设置，空字符串会清除对应配置。
func (h *Handler) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	updates := make(map[string]string)
	if req.NotifyHours != nil {
		normalized, err := service.NormalizeNotifyHours(*req.NotifyHours)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updates[service.SettingNotifyHours] = normalized
	}
	if req.TelegramBotToken != nil {
		updates[service.SettingTelegramBotToken] = *req.TelegramBotToken
	}
	if req.TelegramChatID != nil {
		updates[service.SettingTelegramChatID] = *req.TelegramChatID
	}
	if req.BarkURL != nil {
		updates[service.SettingBarkURL] = *req.BarkURL
	}

	if err := h.settings.UpdateMany(c.Request.Context(), updates); err != nil {
		log.Printf("更新通知设置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新设置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "设置更新成功"})
}

// TestNotify 使用已保存的渠道配置发送一条测试通知。
func (h *Handler) TestNotify(c *gin.Context) {
	var req TestNotifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	settings, err := h.settings.GetNotificationSettings(c.Request.Context())
	if err != nil {
		log.Printf("读取测试通知设置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取通知配置失败"})
		return
	}

	testMessage := "SubDock 通知测试 - 如果你看到这条消息，说明通知配置正确！"
	switch req.Type {
	case "telegram":
		if settings.TelegramBotToken == "" || settings.TelegramChatID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置 Telegram Bot Token 和 Chat ID"})
			return
		}
		err = h.notifier.SendTelegram(settings.TelegramBotToken, settings.TelegramChatID, testMessage)
	case "bark":
		if settings.BarkURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置 Bark URL"})
			return
		}
		err = h.notifier.SendBark(settings.BarkURL, "SubDock 通知测试", testMessage)
	}
	if err != nil {
		log.Printf("发送 %s 测试通知失败: %v", req.Type, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "发送通知失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "测试通知已发送"})
}
