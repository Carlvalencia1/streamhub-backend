package http

import (
	"github.com/gin-gonic/gin"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {
	g := api.Group("/upload")
	g.Use(middleware.AuthMiddleware())
	g.POST("", handler.Upload)
}
