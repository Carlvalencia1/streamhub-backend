package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/Carlvalencia1/streamhub-backend/internal/chat/application"
	wsManager "github.com/Carlvalencia1/streamhub-backend/internal/platform/websocket"
)


type ChatWSHandler struct {
	manager        *wsManager.Manager
	sendMessageUse *application.SendMessage
}

func NewChatWSHandler(
	manager *wsManager.Manager,
	send *application.SendMessage,
) *ChatWSHandler {
	return &ChatWSHandler{
		manager:        manager,
		sendMessageUse: send,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type IncomingMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

func (h *ChatWSHandler) Handle(c *gin.Context) {

	streamID := c.Param("stream_id")

	userID := c.GetString("user_id")
	username := c.GetString("username")

	// 🔥 Validar auth
	if userID == "" {
		log.Println("WS ERROR: user_id missing in context")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	log.Println("WS CONNECT:", userID, "stream:", streamID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WS UPGRADE ERROR:", err)
		return
	}

	hub := h.manager.GetHub(streamID)

	client := &wsManager.Client{
		ID:       userID,
		Username: username,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	hub.Register <- client

	// 🔥 Writer goroutine
	go client.WritePump()

	// 🔥 Reader
	client.ReadPump(func(data []byte) {

		log.Println("WS MESSAGE RECEIVED:", string(data))

		var msg IncomingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Println("JSON ERROR:", err)
			return
		}

		if msg.Type != "send_message" {
			log.Println("IGNORED MESSAGE TYPE:", msg.Type)
			return
		}

		if msg.Content == "" {
			log.Println("EMPTY MESSAGE")
			return
		}

		// 🔥 Context correcto del request
		ctx := c.Request.Context()

		// 🔥 Guardar en DB
		saved, err := h.sendMessageUse.Execute(
			ctx,
			streamID,
			userID,
			msg.Content,
		)

		if err != nil {
			log.Println("ERROR SAVING MESSAGE:", err)
			return
		}

		log.Println("MESSAGE SAVED:", saved.ID)

		// 🔥 Broadcast a la sala
		out := map[string]interface{}{
			"type":       "message",
			"id":         saved.ID,
			"user_id":    userID,
			"username":   username,
			"content":    saved.Content,
			"created_at": saved.CreatedAt,
		}

		jsonMsg, err := json.Marshal(out)
		if err != nil {
			log.Println("JSON MARSHAL ERROR:", err)
			return
		}

		hub.Broadcast <- jsonMsg
	})

	// 🔥 Cuando se desconecta
	defer func() {
		hub.Unregister <- client
		conn.Close()
		log.Println("WS DISCONNECT:", userID)
	}()
}