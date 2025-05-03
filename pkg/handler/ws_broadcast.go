package handler

// "sync"

// "github.com/gorilla/websocket"

func (h *Hub) BroadcastToSpace(spaceID string, message interface{}) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for _, client := range h.clients {
		if clientSpaceID, ok := client.SpaceID(); ok && clientSpaceID == spaceID {
			client.Conn.WriteJSON(message)
		}
	}
}
