package http

import (
	"github.com/gin-gonic/gin"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/middleware"
)

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {
	g := api.Group("/communities")
	g.Use(middleware.AuthMiddleware())

	g.POST("", handler.CreateCommunity)
	g.GET("", handler.GetMyCommunities)
	g.GET("/:id", handler.GetCommunity)
	g.PUT("/:id", handler.UpdateCommunity)
	g.DELETE("/:id", handler.DeleteCommunity)
	g.POST("/join/:code", handler.JoinByInvite)
	g.DELETE("/:id/leave", handler.LeaveCommunity)
	g.POST("/:id/channels", handler.CreateChannel)
	g.DELETE("/:id/channels/:channelId", handler.DeleteChannel)
	g.DELETE("/:id/members/:memberId", handler.RemoveMember)

	g.GET("/:id/channels/:channelId/messages", handler.GetMessages)
	g.POST("/:id/channels/:channelId/messages", handler.SendMessage)
	g.POST("/:id/channels/:channelId/messages/poll", handler.SendPollMessage)
	g.DELETE("/:id/channels/:channelId/messages/:messageId", handler.DeleteMessage)
	g.POST("/:id/channels/:channelId/messages/:messageId/react", handler.ReactToMessage)
	g.PUT("/:id/channels/:channelId/settings", handler.SetDisappearing)
	g.POST("/polls/:pollId/vote", handler.VotePoll)
}
