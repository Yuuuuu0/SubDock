package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"subdock/internal/model"
	"subdock/internal/service"
)

// SettingsResponse 设置响应
type SettingsResponse struct {
	NotifyHours    string `json:"notify_hours"`
	TelegramBotToken string `json:"telegram_bot_token"`
	TelegramChatID   string `json:"telegram_chat_id"`
	BarkURL          string `json:"bark_url"`
}

// UpdateSettingsRequest 更新设置请求
type UpdateSettingsRequest struct {
	NotifyHours      string `json:"notify_hours"`
	TelegramBotToken string `json:"telegram_bot_token"`
	TelegramChatID   string `json:"telegram_chat_id"`
	BarkURL          string `json:"bark_url"`
}

// TestNotifyRequest 测试通知请求
type TestNotifyRequest struct {
	Type string `json:"type" binding:"required,oneof=telegram bark"`
}

// GetSettings 获取设置
func GetSettings(c *gin.Context) {
	settings := SettingsResponse{
		NotifyHours:      getSetting("notify_hours", "9"),
		TelegramBotToken: getSetting("telegram_bot_token", ""),
		TelegramChatID:   getSetting("telegram_chat_id", ""),
		BarkURL:          getSetting("bark_url", ""),
	}
	
	c.JSON(http.StatusOK, settings)
}

// UpdateSettings 更新设置
func UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if req.NotifyHours != "" {
		setSetting("notify_hours", req.NotifyHours)
	}
	if req.TelegramBotToken != "" {
		setSetting("telegram_bot_token", req.TelegramBotToken)
	}
	if req.TelegramChatID != "" {
		setSetting("telegram_chat_id", req.TelegramChatID)
	}
	if req.BarkURL != "" {
		setSetting("bark_url", req.BarkURL)
	}

	c.JSON(http.StatusOK, gin.H{"message": "设置更新成功"})
}

// TestNotify 测试通知
func TestNotify(c *gin.Context) {
	var req TestNotifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	notifier := service.NewNotifier()
	
	testMsg := "SubDock 通知测试 - 如果你看到这条消息，说明通知配置正确！"
	var err error

	switch req.Type {
	case "telegram":
		token := getSetting("telegram_bot_token", "")
		chatID := getSetting("telegram_chat_id", "")
		if token == "" || chatID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置 Telegram Bot Token 和 Chat ID"})
			return
		}
		err = notifier.SendTelegram(token, chatID, testMsg)
	case "bark":
		barkURL := getSetting("bark_url", "")
		if barkURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请先配置 Bark URL"})
			return
		}
		err = notifier.SendBark(barkURL, "SubDock 通知测试", testMsg)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发送通知失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "测试通知已发送"})
}

func getSetting(key, defaultVal string) string {
	var setting model.Setting
	if err := model.GetDB().Where("key = ?", key).First(&setting).Error; err != nil {
		return defaultVal
	}
	return setting.Value
}

func setSetting(key, value string) {
	var setting model.Setting
	result := model.GetDB().Where("key = ?", key).First(&setting)
	if result.Error != nil {
		model.GetDB().Create(&model.Setting{Key: key, Value: value})
	} else {
		model.GetDB().Model(&setting).Update("value", value)
	}
}
