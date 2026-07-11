package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// PublicConfig 表示无需认证即可读取的前端展示配置。
type PublicConfig struct {
	WebsiteTitle string `json:"website_title"`
}

// GetPublicConfig 返回网站标题等公开配置。
func (h *Handler) GetPublicConfig(c *gin.Context) {
	c.JSON(http.StatusOK, PublicConfig{
		WebsiteTitle: h.config.WebsiteTitle,
	})
}
