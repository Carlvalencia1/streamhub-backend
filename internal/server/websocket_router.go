package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	chatApp "github.com/Carlvalencia1/streamhub-backend/internal/chat/application"
	chatInfra "github.com/Carlvalencia1/streamhub-backend/internal/chat/infrastructure"
	chatWS "github.com/Carlvalencia1/streamhub-backend/internal/chat/interfaces/ws"
	ws "github.com/Carlvalencia1/streamhub-backend/internal/platform/websocket"
	"github.com/Carlvalencia1/streamhub-backend/internal/platform/webrtc"
)

func RegisterWebSocketRoutes(
	r *gin.Engine,
	manager *ws.Manager,
	db *sql.DB,
	authMiddleware gin.HandlerFunc,
) {

	chatRepo := chatInfra.NewMySQLRepository(db)
	sendUC := chatApp.NewSendMessage(chatRepo)

	chatHandler := chatWS.NewChatWSHandler(manager, sendUC)
	webrtcHandler := chatWS.NewWebRTCHandler(webrtc.NewSignalingServer())

	wsGroup := r.Group("/ws")

	// 🔥 IMPORTANTE — proteger websocket con JWT
	wsGroup.Use(authMiddleware)

	// Chat WebSocket
	wsGroup.GET("/chat/:stream_id", chatHandler.Handle)

	// WebRTC Signaling - Transmisor (Broadcaster)
	wsGroup.GET("/broadcast/:stream_id", webrtcHandler.HandleBroadcaster)

	// WebRTC Signaling - Espectador (Viewer)
	wsGroup.GET("/watch/:stream_id", webrtcHandler.HandleViewer)
}