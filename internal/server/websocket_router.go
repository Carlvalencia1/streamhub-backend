package server

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	chatApp "github.com/Carlvalencia1/streamhub-backend/internal/chat/application"
	chatInfra "github.com/Carlvalencia1/streamhub-backend/internal/chat/infrastructure"
	chatWS "github.com/Carlvalencia1/streamhub-backend/internal/chat/interfaces/ws"
	ws "github.com/Carlvalencia1/streamhub-backend/internal/platform/websocket"
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

	wsGroup := r.Group("/ws")

	// 🔥 IMPORTANTE — proteger websocket con JWT
	wsGroup.Use(authMiddleware)

	wsGroup.GET("/chat/:stream_id", chatHandler.Handle)
}