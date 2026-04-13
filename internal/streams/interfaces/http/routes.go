package http

import (
	"github.com/gin-gonic/gin"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, handler *Handler, validationHandler *StreamValidationHandler) {

	streams := r.Group("/streams")

	streams.GET("/", handler.GetAll)
	streams.GET("/:id", handler.GetByID)

	// NGINX RTMP Webhooks (no auth required)
	streams.POST("/validate-key", validationHandler.ValidateKey)
	streams.POST("/stop", validationHandler.StopStream)
	streams.GET("/health", validationHandler.HealthCheck)

	protected := streams.Group("/")
	protected.Use(middleware.AuthMiddleware())

	protected.POST("/", handler.Create)
	protected.PUT("/:id/start", handler.Start)
	protected.PUT("/:id/stop", handler.Stop)
	protected.POST("/:id/join", handler.Join)
	protected.GET("/:id/playback", handler.GetPlayback)
}