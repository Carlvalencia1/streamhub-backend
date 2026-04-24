package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Carlvalencia1/streamhub-backend/internal/notifications/application"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
)

type Handler struct {
	registerUC       *application.RegisterFcmToken
	removeUC         *application.RemoveFcmToken
	notifyStreamLive *application.NotifyStreamLive
}

func NewHandler(
	registerUC *application.RegisterFcmToken,
	removeUC *application.RemoveFcmToken,
	notifyStreamLive *application.NotifyStreamLive,
) *Handler {
	return &Handler{
		registerUC:       registerUC,
		removeUC:         removeUC,
		notifyStreamLive: notifyStreamLive,
	}
}

// RegisterFCMToken maneja el registro de tokens FCM
func (h *Handler) RegisterFCMToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Warn("user_id not found in context for RegisterFCMToken")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req RegisterFCMTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("invalid request body for RegisterFCMToken: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if req.Token == "" {
		logger.Warn("empty token provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "token cannot be empty"})
		return
	}

	input := application.RegisterFcmTokenInput{
		UserID:     userID.(string),
		Token:      req.Token,
		Platform:   req.Platform,
		DeviceID:   req.DeviceID,
		AppVersion: req.AppVersion,
	}

	if err := h.registerUC.Execute(c, input); err != nil {
		logger.Error("failed to register FCM token: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Token registrado exitosamente"})
}

// RemoveFCMToken maneja la eliminación de tokens FCM
func (h *Handler) RemoveFCMToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Warn("user_id not found in context for RemoveFCMToken")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req RemoveFCMTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("invalid request body for RemoveFCMToken: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if req.Token == "" {
		logger.Warn("empty token provided for removal")
		c.JSON(http.StatusBadRequest, gin.H{"error": "token cannot be empty"})
		return
	}

	input := application.RemoveFcmTokenInput{
		UserID: userID.(string),
		Token:  req.Token,
	}

	if err := h.removeUC.Execute(c, input); err != nil {
		logger.Error("failed to remove FCM token: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Token eliminado exitosamente"})
}

// NotifyStreamLive maneja el envío de notificaciones de stream en vivo
func (h *Handler) NotifyStreamLive(c *gin.Context) {
	var req NotifyStreamLiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("invalid request body for NotifyStreamLive: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	if req.StreamID == "" || req.StreamTitle == "" || req.OwnerUserID == "" {
		logger.Warn("missing required fields for NotifyStreamLive")
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream_id, stream_title and owner_user_id are required"})
		return
	}

	input := application.NotifyStreamLiveInput{
		StreamID:    req.StreamID,
		StreamTitle: req.StreamTitle,
		OwnerUserID: req.OwnerUserID,
	}

	if err := h.notifyStreamLive.Execute(c, input); err != nil {
		logger.Error("failed to send stream live notifications: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Notificaciones enviadas a los seguidores"})
}

// Request/Response DTOs

type RegisterFCMTokenRequest struct {
	Token      string `json:"token" binding:"required"`
	Platform   string `json:"platform" binding:"required,oneof=android ios"`
	DeviceID   string `json:"device_id"`
	AppVersion string `json:"app_version"`
}

type RemoveFCMTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type NotifyStreamLiveRequest struct {
	StreamID    string `json:"stream_id" binding:"required"`
	StreamTitle string `json:"stream_title" binding:"required"`
	OwnerUserID string `json:"owner_user_id" binding:"required"`
}