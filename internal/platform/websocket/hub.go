package websocket
import "fmt"
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	StreamID   string
}

func NewHub(streamID string) *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		StreamID:   streamID,
	}
}

func (h *Hub) Run() {
	for {
		select {

		case client := <-h.Register:
			h.Clients[client] = true
			h.sendViewerCount()

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				h.sendViewerCount()
			}

		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

func (h *Hub) sendViewerCount() {
	count := len(h.Clients)

	msg := []byte(`{"type":"viewer_count","count":` + itoa(count) + `}`)

	for client := range h.Clients {
		client.Send <- msg
	}
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}