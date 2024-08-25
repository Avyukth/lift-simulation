package websockets

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/Avyukth/lift-simulation/internal/domain"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// WebSocketHandler handles WebSocket connections for real-time updates
func WebSocketHandler(c *fiber.Ctx) error {
	// IsWebSocketUpgrade returns true if the client
	// requested upgrade to the WebSocket protocol.
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// WebSocketClient represents a WebSocket client connection
type WebSocketClient struct {
	Conn *websocket.Conn
	Mu   sync.Mutex
}

// WebSocketHub maintains the set of active clients and broadcasts messages to the clients
type WebSocketHub struct {
	Clients    map[*WebSocketClient]bool
	Register   chan *WebSocketClient
	Unregister chan *WebSocketClient
	Broadcast  chan []byte
	Mu         sync.Mutex
}

// NewWebSocketHub creates a new WebSocketHub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		Clients:    make(map[*WebSocketClient]bool),
		Register:   make(chan *WebSocketClient),
		Unregister: make(chan *WebSocketClient),
		Broadcast:  make(chan []byte),
	}
}

// Run starts the WebSocketHub
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()
			h.Clients[client] = true
			h.Mu.Unlock()
		case client := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Conn.Close()
			}
			h.Mu.Unlock()
		case message := <-h.Broadcast:
			h.Mu.Lock()
			for client := range h.Clients {
				client.Mu.Lock()
				if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("error: %v", err)
					client.Conn.Close()
					delete(h.Clients, client)
				}
				client.Mu.Unlock()
			}
			h.Mu.Unlock()
		}
	}
}

// BroadcastUpdate sends an update to all connected WebSocket clients
func (h *WebSocketHub) BroadcastUpdate(event domain.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("error marshaling event: %v", err)
		return
	}
	h.Broadcast <- data
}

// WebSocketUpgradeHandler handles the WebSocket upgrade
func WebSocketUpgradeHandler(hub *WebSocketHub) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		client := &WebSocketClient{Conn: c}
		hub.Register <- client
		defer func() {
			hub.Unregister <- client
		}()

		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			}

			if messageType == websocket.TextMessage {
				// Handle incoming messages if needed
				log.Printf("received message: %s", message)
			}
		}
	})
}
