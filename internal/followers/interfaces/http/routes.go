package http

import (
	"github.com/gin-gonic/gin"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {
	g := api.Group("/followers")
	g.Use(middleware.AuthMiddleware())

	g.GET("/following", handler.GetFollowing)
	g.POST("/:streamerId/follow", handler.Follow)
	g.DELETE("/:streamerId/unfollow", handler.Unfollow)
	g.GET("/:streamerId/status", handler.GetStatus)
}
