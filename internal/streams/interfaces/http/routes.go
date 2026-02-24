package http

import (
	"github.com/gin-gonic/gin"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, handler *Handler) {

	streams := r.Group("/streams")

	streams.GET("/", handler.GetAll)

	protected := streams.Group("/")
	protected.Use(middleware.AuthMiddleware())

	protected.POST("/", handler.Create)
	protected.PUT("/:id/start", handler.Start)
}