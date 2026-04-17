package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Carlvalencia1/streamhub-backend/internal/followers/application"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
)

type Handler struct {
	followUC        *application.Follow
	unfollowUC      *application.Unfollow
	getStatusUC     *application.GetFollowerStatus
	getFollowingUC  *application.GetFollowing
}

func NewHandler(
	followUC *application.Follow,
	unfollowUC *application.Unfollow,
	getStatusUC *application.GetFollowerStatus,
	getFollowingUC *application.GetFollowing,
) *Handler {
	return &Handler{
		followUC:       followUC,
		unfollowUC:     unfollowUC,
		getStatusUC:    getStatusUC,
		getFollowingUC: getFollowingUC,
	}
}

func (h *Handler) Follow(c *gin.Context) {
	followerID := c.GetString("user_id")
	streamerID := strings.TrimSpace(c.Param("streamerId"))
	if streamerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "streamer id required"})
		return
	}
	if err := h.followUC.Execute(c, followerID, streamerID); err != nil {
		logger.Error("follow failed: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "followed"})
}

func (h *Handler) Unfollow(c *gin.Context) {
	followerID := c.GetString("user_id")
	streamerID := strings.TrimSpace(c.Param("streamerId"))
	if streamerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "streamer id required"})
		return
	}
	if err := h.unfollowUC.Execute(c, followerID, streamerID); err != nil {
		logger.Error("unfollow failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "unfollowed"})
}

func (h *Handler) GetStatus(c *gin.Context) {
	followerID := c.GetString("user_id")
	streamerID := strings.TrimSpace(c.Param("streamerId"))
	if streamerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "streamer id required"})
		return
	}
	isFollowing, count, err := h.getStatusUC.Execute(c, followerID, streamerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"is_following":   isFollowing,
		"follower_count": count,
	})
}

func (h *Handler) GetFollowing(c *gin.Context) {
	followerID := c.GetString("user_id")
	ids, err := h.getFollowingUC.Execute(c, followerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ids == nil {
		ids = []string{}
	}
	c.JSON(http.StatusOK, gin.H{"streamer_ids": ids})
}
