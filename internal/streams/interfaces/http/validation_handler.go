package http

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Carlvalencia1/streamhub-backend/internal/streams/domain"
	"github.com/Carlvalencia1/streamhub-backend/pkg/response"
)

// StreamValidationHandler handles requests from NGINX RTMP module
type StreamValidationHandler struct {
	streamRepository domain.StreamRepository
}

// NewStreamValidationHandler creates a new handler
func NewStreamValidationHandler(streamRepository domain.StreamRepository) *StreamValidationHandler {
	return &StreamValidationHandler{
		streamRepository: streamRepository,
	}
}

// ValidateStreamKeyResponse is the response for NGINX
type ValidateStreamKeyResponse struct {
	Valid bool `json:"valid"`
}

// ValidateKey validates stream_key from SRS on_publish event
// SRS sends POST with JSON body; NGINX RTMP uses query params
// URL: POST /api/streams/validate-key
func (h *StreamValidationHandler) ValidateKey(c *gin.Context) {
	// Try query params first (NGINX RTMP style)
	streamKey := c.Query("name")
	app := c.Query("app")

	// If not in query params, try SRS JSON body
	if streamKey == "" {
		var body struct {
			Stream string `json:"stream"`
			App    string `json:"app"`
		}
		if err := c.ShouldBindJSON(&body); err == nil {
			streamKey = body.Stream
			if app == "" {
				app = body.App
			}
		}
	}

	if app == "" {
		app = "live"
	}

	log.Printf("[StreamValidation] Received validation request: app=%s, stream=%s", app, streamKey)

	if streamKey == "" {
		log.Printf("[StreamValidation] Missing stream key parameter")
		response.Error(c, http.StatusBadRequest, "Missing stream key")
		return
	}

	log.Printf("[StreamValidation] Validating stream key: %s (app: %s)", streamKey, app)

	// Check if stream exists with this stream_key
	stream, err := h.streamRepository.GetByStreamKey(c.Request.Context(), streamKey)
	if err != nil || stream == nil {
		log.Printf("[StreamValidation] Stream key not found or invalid: %s (error: %v)", streamKey, err)
		response.Error(c, http.StatusUnauthorized, "Invalid stream key")
		return
	}

	// Check if stream is in the correct state (should be created, not already streaming)
	if stream.IsLive {
		log.Printf("[StreamValidation] Stream already active: %s", stream.ID)
		// Allow re-connection (overwrite previous stream)
		// If you want to prevent this, uncomment below:
		// response.Error(c, http.StatusConflict, "Stream already active")
		// return
	}

	// Mark stream as active/live
	stream.IsLive = true
	if err := h.streamRepository.Update(c.Request.Context(), stream); err != nil {
		log.Printf("[StreamValidation] Failed to update stream: %v", err)
		response.Error(c, http.StatusInternalServerError, "Failed to mark stream as active")
		return
	}

	log.Printf("[StreamValidation] ✓ Stream key validated successfully: %s", streamKey)
	response.JSON(c, http.StatusOK, ValidateStreamKeyResponse{Valid: true})
}

// StopStream handles SRS on_unpublish event
// Called when a stream publisher disconnects
// URL: POST /api/streams/stop
func (h *StreamValidationHandler) StopStream(c *gin.Context) {
	// Try query params first (NGINX RTMP style)
	streamKey := c.Query("name")
	app := c.Query("app")

	// If not in query params, try SRS JSON body
	if streamKey == "" {
		var body struct {
			Stream string `json:"stream"`
			App    string `json:"app"`
		}
		if err := c.ShouldBindJSON(&body); err == nil {
			streamKey = body.Stream
			if app == "" {
				app = body.App
			}
		}
	}

	if app == "" {
		app = "live"
	}

	log.Printf("[StreamStop] Received stop request: app=%s, stream=%s", app, streamKey)

	if streamKey == "" {
		log.Printf("[StreamStop] Missing stream key parameter")
		response.JSON(c, http.StatusOK, gin.H{"message": "Stream stopped (no key)"})
		return
	}

	log.Printf("[StreamStop] Stopping stream: %s (app: %s)", streamKey, app)

	// Get stream by stream_key
	stream, err := h.streamRepository.GetByStreamKey(c.Request.Context(), streamKey)
	if err != nil || stream == nil {
		log.Printf("[StreamStop] Stream not found: %s (error: %v)", streamKey, err)
		// Return 200 OK even if stream not found (idempotent)
		response.JSON(c, http.StatusOK, gin.H{"message": "Stream stopped (not found)"})
		return
	}

	// Mark stream as inactive
	stream.IsLive = false
	if err := h.streamRepository.Update(c.Request.Context(), stream); err != nil {
		log.Printf("[StreamStop] Failed to update stream: %v", err)
		// Still return 200 to prevent NGINX errors
		response.JSON(c, http.StatusOK, gin.H{
			"message": "Stream marked as stopped",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("[StreamStop] ✓ Stream stopped successfully: %s", streamKey)
	response.JSON(c, http.StatusOK, gin.H{
		"message":   "Stream stopped successfully",
		"stream_id": stream.ID,
	})
}

// HealthCheck for NGINX upstream
func (h *StreamValidationHandler) HealthCheck(c *gin.Context) {
	log.Printf("[HealthCheck] Backend is healthy")
	response.JSON(c, http.StatusOK, gin.H{"status": "healthy"})
}
