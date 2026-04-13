package ws

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/logger"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/webrtc"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/websocket"
)

// WebRTCHandler maneja las conexiones WebRTC via WebSocket
type WebRTCHandler struct {
	signalingServer *webrtc.SignalingServer
	upgrader        websocket.Upgrader
}

// NewWebRTCHandler crea un nuevo handler WebRTC
func NewWebRTCHandler(signalingServer *webrtc.SignalingServer) *WebRTCHandler {
	return &WebRTCHandler{
		signalingServer: signalingServer,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Permitir CORS
			},
		},
	}
}

// HandleBroadcaster maneja la conexión del transmisor
func (h *WebRTCHandler) HandleBroadcaster(c *gin.Context) {
	streamID := c.Param("stream_id")
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("user_id not found in context for broadcaster")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userIDStr := userID.(string)

	// Upgrade HTTP connection a WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.ErrorWithContext("WebRTCHandler", "failed to upgrade connection", err)
		return
	}
	defer conn.Close()

	// Iniciar sesión de broadcast
	if err := h.signalingServer.StartBroadcast(streamID, userIDStr); err != nil {
		logger.ErrorWithContext("WebRTCHandler", "failed to start broadcast", err)
		conn.WriteJSON(gin.H{"error": "failed to start broadcast"})
		return
	}

	logger.StreamEvent("BROADCASTER_CONNECTED", streamID, "User: "+userIDStr)

	// Enviar confirmación al transmisor
	conn.WriteJSON(webrtc.SignalingMessage{
		Type:   webrtc.TypeBroadcasterReady,
		StreamID: streamID,
	})

	// Leer mensajes del transmisor
	for {
		var msg webrtc.SignalingMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.ErrorWithContext("WebRTCHandler", "websocket error", err)
			}
			break
		}

		msg.StreamID = streamID
		msg.FromUserID = userIDStr

		// Procesar mensaje de signaling
		if err := h.signalingServer.ProcessSignalingMessage(&msg); err != nil {
			logger.ErrorWithContext("WebRTCHandler", "failed to process signaling message", err)
			continue
		}

		// Enviar a todos los espectadores (implementar en siguiente paso)
		logger.Debug(fmt.Sprintf("Broadcasting message type %s for stream %s", msg.Type, streamID))
	}

	// Detener broadcast cuando se desconecta
	h.signalingServer.StopBroadcast(streamID)
	logger.StreamEvent("BROADCASTER_DISCONNECTED", streamID, "User: "+userIDStr)
}

// HandleViewer maneja la conexión de un espectador
func (h *WebRTCHandler) HandleViewer(c *gin.Context) {
	streamID := c.Param("stream_id")
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("user_id not found in context for viewer")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userIDStr := userID.(string)

	// Verificar que el stream esté en vivo
	if !h.signalingServer.IsBroadcasting(streamID) {
		logger.Warn("Stream not broadcasting: " + streamID)
		c.JSON(http.StatusNotFound, gin.H{"error": "stream not broadcasting"})
		return
	}

	// Upgrade HTTP connection a WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.ErrorWithContext("WebRTCHandler", "failed to upgrade connection", err)
		return
	}
	defer conn.Close()

	// Registrar espectador
	if err := h.signalingServer.RegisterViewer(streamID, userIDStr); err != nil {
		logger.ErrorWithContext("WebRTCHandler", "failed to register viewer", err)
		conn.WriteJSON(gin.H{"error": "failed to register viewer"})
		return
	}

	logger.StreamEvent("VIEWER_CONNECTED", streamID, fmt.Sprintf("User: %s | Total viewers: %d", userIDStr, h.signalingServer.GetViewerCount(streamID)))

	// Enviar confirmación al espectador
	broadcasterID, _ := h.signalingServer.GetBroadcaster(streamID)
	conn.WriteJSON(gin.H{
		"type":          "viewer_ready",
		"stream_id":     streamID,
		"broadcaster_id": broadcasterID,
	})

	// Leer mensajes del espectador
	for {
		var msg webrtc.SignalingMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.ErrorWithContext("WebRTCHandler", "websocket error", err)
			}
			break
		}

		msg.StreamID = streamID
		msg.FromUserID = userIDStr

		// Procesar mensaje de signaling
		if err := h.signalingServer.ProcessSignalingMessage(&msg); err != nil {
			logger.ErrorWithContext("WebRTCHandler", "failed to process signaling message", err)
			continue
		}

		logger.Debug(fmt.Sprintf("Received signaling message from viewer: %s", msg.Type))
	}

	// Desregistrar espectador
	h.signalingServer.UnregisterViewer(streamID, userIDStr)
	logger.StreamEvent("VIEWER_DISCONNECTED", streamID, fmt.Sprintf("User: %s | Remaining viewers: %d", userIDStr, h.signalingServer.GetViewerCount(streamID)))
}

// HandleSignaling maneja (alternativa) ambos transmisor y espectador en misma conexión
// NO RECOMENDADO para este proyecto, pero útil si necesitas simplificar
func (h *WebRTCHandler) HandleSignaling(c *gin.Context) {
	streamID := c.Param("stream_id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userIDStr := userID.(string)
	isBroadcaster := strings.ToLower(c.Query("type")) == "broadcaster"

	if isBroadcaster {
		h.HandleBroadcaster(c)
	} else {
		h.HandleViewer(c)
	}
}
