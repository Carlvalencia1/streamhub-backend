package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	ws "github.com/Carlvalencia1/streamhub-backend/internal/platform/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func RegisterWebSocketRoutes(r *gin.Engine, hub *ws.Hub) {

	r.GET("/ws", func(c *gin.Context) {

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := &ws.Client{
			ID:   c.Query("user_id"),
			Conn: conn,
			Send: make(chan ws.Message),
		}

		hub.Register <- client

		go readPump(client, hub)
		go writePump(client)
	})
}

func readPump(client *ws.Client, hub *ws.Hub) {

	defer func() {
		hub.Unregister <- client
		client.Conn.Close()
	}()

	for {
		var msg ws.Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}

		hub.Broadcast <- msg
	}
}

func writePump(client *ws.Client) {

	for msg := range client.Send {
		client.Conn.WriteJSON(msg)
	}
}
