package router

import (
	"subdock/internal/handler"
	"subdock/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/config", handler.GetPublicConfig)
		api.POST("/login", handler.Login)

		auth := api.Group("")
		auth.Use(middleware.AuthRequired())
		{
			auth.POST("/change-password", handler.ChangePassword)

			auth.GET("/subscriptions", handler.ListSubscriptions)
			auth.POST("/subscriptions", handler.CreateSubscription)
			auth.PUT("/subscriptions/:id", handler.UpdateSubscription)
			auth.DELETE("/subscriptions/:id", handler.DeleteSubscription)
			auth.POST("/subscriptions/:id/test-notify", handler.TestSubscriptionNotify)

			auth.GET("/settings", handler.GetSettings)
			auth.PUT("/settings", handler.UpdateSettings)
			auth.POST("/settings/test-notify", handler.TestNotify)
		}
	}

	serveStatic(r)

	return r
}
