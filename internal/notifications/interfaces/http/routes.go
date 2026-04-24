package http

import (
	"github.com/gin-gonic/gin"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {
	notificationsGroup := api.Group("/notifications")
	notificationsGroup.Use(middleware.AuthMiddleware())
	{
		notificationsGroup.POST("/fcm-token", handler.RegisterFCMToken)
		notificationsGroup.DELETE("/fcm-token", handler.RemoveFCMToken)
		notificationsGroup.POST("/stream-live", handler.NotifyStreamLive)
	}
}
