package http

import (
	"github.com/gin-gonic/gin"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {
	g := api.Group("/channel")
	g.Use(middleware.AuthMiddleware())

	g.POST("/posts", handler.CreatePost)
	g.POST("/posts/poll", handler.CreatePollPost)
	g.GET("/posts", handler.GetMyPosts)
	g.DELETE("/posts/:postId", handler.DeletePost)
	g.GET("/feed", handler.GetFeed)
	g.GET("/streamer/:streamerId/posts", handler.GetStreamerPosts)
	g.POST("/polls/:pollId/vote", handler.VotePoll)
}
