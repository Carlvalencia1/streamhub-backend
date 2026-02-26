package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/security"
	streamsApp "github.com/Carlvalencia1/streamhub-backend/internal/streams/application"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type StreamHandler struct {
	createUC *streamsApp.CreateStream
	getUC    *streamsApp.GetStreams
}

func NewStreamHandler(
	createUC *streamsApp.CreateStream,
	getUC *streamsApp.GetStreams,
) *StreamHandler {
	return &StreamHandler{
		createUC: createUC,
		getUC:    getUC,
	}
}

type StreamMessage struct {
	Action      string `json:"action"`
	StreamID    string `json:"stream_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Category    string `json:"category"`
}

func (h *StreamHandler) Handle(c *gin.Context) {

	token := c.Query("token")

	claims, err := security.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	userID := claims.UserID

	for {

		var msg StreamMessage

		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		switch msg.Action {
		case "create":
			// Crear stream
			err := h.createUC.Execute(c, msg.Title, msg.Description, msg.Thumbnail, msg.Category, userID)
			if err != nil {
				conn.WriteJSON(gin.H{"error": "failed to create stream"})
				continue
			}
			conn.WriteJSON(gin.H{"message": "stream created successfully"})

		case "get":
			// Obtener streams
			streams, err := h.getUC.Execute(c)
			if err != nil {
				conn.WriteJSON(gin.H{"error": "failed to get streams"})
				continue
			}
			conn.WriteJSON(streams)

		default:
			conn.WriteJSON(gin.H{"error": "unknown action"})
		}
	}
}
