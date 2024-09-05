package websockets

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"

	"github.com/Avyukth/lift-simulation/pkg/logger"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

// StatusUpdate represents a status update for a floor or lift
type StatusUpdate struct {
	Type         string `json:"type"` // "floor" or "lift"
	ID           string `json:"id"`   // Floor number or lift ID
	Status       string `json:"status"`
	CurrentFloor int    `json:"currentFloor,omitempty"` // Only for lifts
}

// WebSocketClient represents a WebSocket client connection
type WebSocketClient struct {
	Conn     *websocket.Conn
	Mu       sync.Mutex
	FloorSub string // Subscribed floor number
	LiftSub  string // Subscribed lift ID
}

// WebSocketHub maintains the set of active clients and broadcasts messages to the clients
type WebSocketHub struct {
	Clients    map[*WebSocketClient]bool
	Register   chan *WebSocketClient
	Unregister chan *WebSocketClient
	Broadcast  chan StatusUpdate
	Mu         sync.Mutex
	Log        *logger.Logger
}

// NewWebSocketHub creates a new WebSocketHub
func NewWebSocketHub(log *logger.Logger) *WebSocketHub {
	return &WebSocketHub{
		Clients:    make(map[*WebSocketClient]bool),
		Register:   make(chan *WebSocketClient),
		Unregister: make(chan *WebSocketClient),
		Broadcast:  make(chan StatusUpdate),
		Log:        log,
	}
}

// Run starts the WebSocketHub
func (h *WebSocketHub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.Log.Info(ctx, "WebSocketHub shutting down")
			return
		case client := <-h.Register:
			h.Mu.Lock()
			h.Clients[client] = true
			h.Log.Info(ctx, "New client registered", "total_clients", len(h.Clients))
			h.Mu.Unlock()
		case client := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Conn.Close()
				h.Log.Info(ctx, "Client unregistered", "total_clients", len(h.Clients))
			}
			h.Mu.Unlock()
		case update := <-h.Broadcast:
			h.Mu.Lock()
			for client := range h.Clients {
				if (update.Type == "floor" && client.FloorSub == update.ID) ||
					(update.Type == "lift" && client.LiftSub == update.ID) {
					client.Mu.Lock()
					data, err := json.Marshal(update)
					if err != nil {
						h.Log.Error(ctx, "Error marshaling update", "error", err)
						client.Mu.Unlock()
						continue
					}
					if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
						h.Log.Error(ctx, "Error writing message to client", "error", err)
						client.Conn.Close()
						delete(h.Clients, client)
					}
					client.Mu.Unlock()
				}
			}
			h.Mu.Unlock()
		}
	}
}

// WebSocketHandler handles WebSocket connections for real-time updates
func WebSocketHandler(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// WebSocketUpgradeHandler handles the WebSocket upgrade
func WebSocketUpgradeHandler(hub *WebSocketHub) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// Create a new client
		client := &WebSocketClient{Conn: c}

		// Register the client
		hub.Register <- client

		// Ensure the client is unregistered when the function returns
		defer func() {
			hub.Unregister <- client
		}()

		for {
			// Read message from the WebSocket connection
			_, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					hub.Log.Error(context.Background(), "websocket_error", "Error writing message to client", "error", err)
					c.Locals("websocket_error", err.Error())
				}
				break
			}

			// Parse the subscription message
			var subscription struct {
				Type string      `json:"type"`
				ID   interface{} `json:"id"`
			}
			if err := json.Unmarshal(message, &subscription); err != nil {
				hub.Log.Error(context.Background(), "websocket_error", "Error unmarshaling subscription: "+err.Error())
				continue
			}

			// Handle the subscription
			switch subscription.Type {
			case "floor":
				// Parse floor ID as an integer
				floorID, ok := subscription.ID.(float64)
				if !ok {
					c.Locals("websocket_error", "Invalid floor ID format")
					hub.Log.Error(context.Background(), "websocket_error", "Issue with FloorID ", floorID)

					continue
				}
				client.FloorSub = strconv.Itoa(int(floorID))
				hub.Log.Info(context.Background(), "websocket_info", "Client subscribed to floor: "+client.FloorSub)

			case "lift":
				// Keep lift ID as a string
				liftID, ok := subscription.ID.(string)
				if !ok {
					c.Locals("websocket_error", "Invalid lift ID format")
					hub.Log.Error(context.Background(), "websocket_error", "Issue with liftID ", liftID)

					continue
				}
				client.LiftSub = liftID
				c.Locals("websocket_info", "Client subscribed to lift: "+client.LiftSub)
				hub.Log.Info(context.Background(), "websocket_info", "Client subscribed to floor: "+client.LiftSub)

			default:
				c.Locals("websocket_warning", "Unknown subscription type: "+subscription.Type)
				hub.Log.Info(context.Background(), "websocket_warning", "Unknown subscription type:  "+subscription.Type)
			}
		}
	})
}

// BroadcastUpdate sends an update to all relevant WebSocket clients
func (h *WebSocketHub) BroadcastUpdate(update StatusUpdate) {
	h.Broadcast <- update
}
