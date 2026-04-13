package http

import (
	"encoding/json"
	"log"
	"net/http"

	"streamhub/internal/streams/domain"
	"streamhub/pkg/response"
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

// ValidateStreamKeyRequest represents the request from NGINX
// NGINX sends: ?app=live&name={stream_key}
type ValidateStreamKeyRequest struct {
	App  string `json:"app"`
	Name string `json:"name"`
}

// ValidateStreamKeyResponse is the response for NGINX
type ValidateStreamKeyResponse struct {
	Valid bool `json:"valid"`
}

// ValidateKey validates stream_key from NGINX RTMP on_publish event
// NGINX will only allow publishing if this returns 200 OK
// URL: POST /api/streams/validate-key?app=live&name={stream_key}
func (h *StreamValidationHandler) ValidateKey(w http.ResponseWriter, r *http.Request) {
	log.Printf("[StreamValidation] Received validation request: %s %s", r.Method, r.RequestURI)

	// NGINX RTMP sends parameters in query string
	app := r.URL.Query().Get("app")
	streamKey := r.URL.Query().Get("name")

	if app == "" || streamKey == "" {
		log.Printf("[StreamValidation] Missing app or name parameter")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing app or name"})
		return
	}

	log.Printf("[StreamValidation] Validating stream key: %s (app: %s)", streamKey, app)

	// Check if stream exists with this stream_key
	stream, err := h.streamRepository.GetByStreamKey(r.Context(), streamKey)
	if err != nil || stream == nil {
		log.Printf("[StreamValidation] Stream key not found or invalid: %s (error: %v)", streamKey, err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid stream key"})
		return
	}

	// Check if stream is in the correct state (should be created, not already streaming)
	if stream.IsLive {
		log.Printf("[StreamValidation] Stream already active: %s", stream.ID)
		// Allow re-connection (overwrite previous stream)
		// If you want to prevent this, uncomment below:
		// w.WriteHeader(http.StatusConflict)
		// json.NewEncoder(w).Encode(map[string]string{"error": "Stream already active"})
		// return
	}

	// Mark stream as active/live
	stream.IsLive = true
	if err := h.streamRepository.Update(r.Context(), stream); err != nil {
		log.Printf("[StreamValidation] Failed to update stream: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to mark stream as active"})
		return
	}

	log.Printf("[StreamValidation] ✓ Stream key validated successfully: %s", streamKey)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ValidateStreamKeyResponse{Valid: true})
}

// StopStreamRequest represents the request from NGINX on_publish_done
type StopStreamRequest struct {
	App  string `json:"app"`
	Name string `json:"name"`
}

// StopStream handles NGINX RTMP on_publish_done event
// Called when a stream publisher disconnects
// URL: POST /api/streams/stop?app=live&name={stream_key}
func (h *StreamValidationHandler) StopStream(w http.ResponseWriter, r *http.Request) {
	log.Printf("[StreamStop] Received stop request: %s %s", r.Method, r.RequestURI)

	// NGINX RTMP sends parameters in query string
	app := r.URL.Query().Get("app")
	streamKey := r.URL.Query().Get("name")

	if app == "" || streamKey == "" {
		log.Printf("[StreamStop] Missing app or name parameter")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing app or name"})
		return
	}

	log.Printf("[StreamStop] Stopping stream: %s (app: %s)", streamKey, app)

	// Get stream by stream_key
	stream, err := h.streamRepository.GetByStreamKey(r.Context(), streamKey)
	if err != nil || stream == nil {
		log.Printf("[StreamStop] Stream not found: %s (error: %v)", streamKey, err)
		// Return 200 OK even if stream not found (idempotent)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Stream stopped (not found)"})
		return
	}

	// Mark stream as inactive
	stream.IsActive = false
	if err := h.streamRepository.Update(r.Context(), stream); err != nil {
		log.Printf("[StreamStop] Failed to update stream: %v", err)
		// Still return 200 to prevent NGINX errors
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Stream marked as stopped", "error": err.Error()})
		return
	}

	log.Printf("[StreamStop] ✓ Stream stopped successfully: %s", streamKey)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Stream stopped successfully",
		"stream_id": stream.ID,
	})
}

// HealthCheck for NGINX upstream
func (h *StreamValidationHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response.SuccessResponse(w, http.StatusOK, map[string]string{"status": "healthy"})
}
