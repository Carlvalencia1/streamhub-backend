package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Carlvalencia1/streamhub-backend/internal/notifications/application"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
	"github.com/Carlvalencia1/streamhub-backend/pkg/response"
)

type Handler struct {
	registerUC *application.RegisterFcmToken
	removeUC   *application.RemoveFcmToken
}

func NewHandler(
	registerUC *application.RegisterFcmToken,
	removeUC *application.RemoveFcmToken,
) *Handler {
	return &Handler{
		registerUC: registerUC,
		removeUC:   removeUC,
	}
}

// RegisterFCMToken godoc
// @Summary Registrar token FCM
// @Description Registra o actualiza un token FCM para el dispositivo del usuario autenticado
// @Tags Notifications
// @Accept json
// @Produce json
// @Param request body RegisterFCMTokenRequest true "FCM Token Request"
// @Security Bearer
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/fcm-token [post]
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

	// Validación
	if req.Token == "" {
		logger.Warn("empty token provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "token cannot be empty"})
		return
	}

	// Ejecutar usecase
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

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "Token registrado exitosamente",
	})
}

// RemoveFCMToken godoc
// @Summary Eliminar token FCM
// @Description Elimina un token FCM específico del usuario autenticado
// @Tags Notifications
// @Accept json
// @Produce json
// @Param request body RemoveFCMTokenRequest true "FCM Token to Remove"
// @Security Bearer
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /notifications/fcm-token [delete]
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

	// Validación
	if req.Token == "" {
		logger.Warn("empty token provided for removal")
		c.JSON(http.StatusBadRequest, gin.H{"error": "token cannot be empty"})
		return
	}

	// Ejecutar usecase
	input := application.RemoveFcmTokenInput{
		UserID: userID.(string),
		Token:  req.Token,
	}

	if err := h.removeUC.Execute(c, input); err != nil {
		logger.Error("failed to remove FCM token: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove token"})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Success: true,
		Message: "Token eliminado exitosamente",
	})
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
