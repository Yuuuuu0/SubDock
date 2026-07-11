package router

import (
	"github.com/gin-gonic/gin"

	"subdock/internal/config"
	"subdock/internal/handler"
	"subdock/internal/middleware"
	"subdock/internal/model"
)

// Setup 创建完整的 Gin 路由并注入数据库、配置和业务服务。
func Setup() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	cfg := config.Get()
	h := handler.New(model.GetDB(), cfg)

	api := r.Group("/api")
	{
		api.GET("/config", h.GetPublicConfig)
		api.POST("/login", h.Login)

		auth := api.Group("")
		auth.Use(middleware.AuthRequired(cfg.JWTSecret))
		{
			auth.POST("/change-password", h.ChangePassword)

			auth.GET("/subscriptions", h.ListSubscriptions)
			auth.POST("/subscriptions", h.CreateSubscription)
			auth.GET("/subscriptions/:id", h.GetSubscription)
			auth.PUT("/subscriptions/:id", h.UpdateSubscription)
			auth.POST("/subscriptions/:id/renew", h.RenewSubscription)
			auth.GET("/subscriptions/:id/renewals", h.ListSubscriptionRenewals)
			auth.DELETE("/subscriptions/:id", h.DeleteSubscription)
			auth.POST("/subscriptions/:id/test-notify", h.TestSubscriptionNotify)

			auth.GET("/settings", h.GetSettings)
			auth.PUT("/settings", h.UpdateSettings)
			auth.POST("/settings/test-notify", h.TestNotify)
		}
	}

	serveStatic(r)

	return r
}
