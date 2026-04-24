package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Carlvalencia1/streamhub-backend/internal/notifications/application"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
	"github.com/Carlvalencia1/streamhub-backend/pkg/response"
)

type Handler struct {
	registerUC      *application.RegisterFcmToken
	removeUC        *application.RemoveFcmToken
	notifyStreamLive *application.NotifyStreamLive  // 👈 Agrega esto
}

func NewHandler(
	registerUC *application.RegisterFcmToken,
	removeUC *application.RemoveFcmToken,
	notifyStreamLive *application.NotifyStreamLive, // 👈 Agrega esto
) *Handler {
	return &Handler{
		registerUC:      registerUC,
		removeUC:        removeUC,
		notifyStreamLive: notifyStreamLive, // 👈 Agrega esto
	}
}

// ... tus métodos existentes RegisterFCMToken y RemoveFCMToken ...

// NotifyStreamLiveRequest representa la solicitud para notificar inicio de stream
type NotifyStreamLiveRequest struct {
	StreamID    string `json:"stream_id" binding:"required"`
	StreamTitle string `json:"stream_title" binding:"required"`
	OwnerUserID string `json:"owner_user_id" binding:"required"`
}

// NotifyStreamLive godoc
// @Summary Notificar inicio de stream
// @Description Envía notificaciones push a todos los seguidores del streamer
// @Tags Notifications
// @Accept json
// @Produce json
// @Param request body NotifyStreamLiveRequest true "Stream Live Notification Request"
// @Security Bearer
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/stream-live [post]
func (h *Handler) NotifyStreamLive(c *gin.Context) {
	var req NotifyStreamLiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("invalid request body for NotifyStreamLive: " + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Validación
	if req.StreamID == "" || req.StreamTitle == "" || req.OwnerUserID == "" {
		logger.Warn("missing required fields for NotifyStreamLive")
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream_id, stream_title and owner_user_id are required"})
		return
	}

	// Ejecutar usecase
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

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "Notificaciones enviadas a los seguidores",
	})
}