package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Notifier 通知服务
type Notifier struct {
	client *http.Client
}

// NewNotifier 创建通知服务实例
func NewNotifier() *Notifier {
	return &Notifier{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// SendTelegram 发送 Telegram 消息
func (n *Notifier) SendTelegram(botToken, chatID, message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := map[string]string{
		"chat_id": chatID,
		"text":    message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	resp, err := n.client.Post(apiURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram API 返回错误: %d", resp.StatusCode)
	}

	return nil
}

// SendBark 发送 Bark 通知
func (n *Notifier) SendBark(barkURL, title, message string) error {
	fullURL := fmt.Sprintf("%s/%s/%s", barkURL, url.PathEscape(title), url.PathEscape(message))

	resp, err := n.client.Get(fullURL)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Bark API 返回错误: %d", resp.StatusCode)
	}

	return nil
}
