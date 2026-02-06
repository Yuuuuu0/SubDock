package handler

import (
	"net/http"
	"subdock/internal/config"

	"github.com/gin-gonic/gin"
)

type PublicConfig struct {
	WebsiteTitle string `json:"website_title"`
}

func GetPublicConfig(c *gin.Context) {
	cfg := config.Get()
	c.JSON(http.StatusOK, PublicConfig{
		WebsiteTitle: cfg.WebsiteTitle,
	})
}
