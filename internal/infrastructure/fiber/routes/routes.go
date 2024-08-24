package routes

import (
	"github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/handlers"
	ws "github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/websockets"

	"github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/middleware"
	"github.com/Avyukth/lift-simulation/pkg/logger"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// SetupRoutes configures all the routes for the lift simulation API
func SetupRoutes(app *fiber.App, liftHandler *handlers.LiftHandler, floorHandler *handlers.FloorHandler, systemHandler *handlers.SystemHandler, hub *ws.WebSocketHub, fiberLog *logger.FiberLogger) {
	// Middleware
	authConfig := middleware.Config{
		JWTSecret: "your-jwt-secret", // In production, use a secure method to store this
	}
	authMiddleware := middleware.New(authConfig)

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// API version group
	api := app.Group("/api/v1")

	// Public routes
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// System routes
	system := api.Group("/system")
	system.Post("/configure", authMiddleware, systemHandler.ConfigureSystem)
	system.Get("/configuration", systemHandler.GetSystemConfiguration)
	system.Get("/status", systemHandler.GetSystemStatus)
	system.Post("/reset", authMiddleware, systemHandler.ResetSystem)
	system.Get("/metrics", authMiddleware, systemHandler.GetSystemMetrics)
	system.Post("/simulate-traffic", authMiddleware, systemHandler.SimulateTraffic)

	// Lift routes
	lifts := api.Group("/lifts")
	lifts.Get("/", liftHandler.ListLifts)
	lifts.Get("/:id", liftHandler.GetLift)
	lifts.Post("/:id/move", authMiddleware, liftHandler.MoveLift)
	lifts.Post("/assign", authMiddleware, liftHandler.AssignLiftToFloor)
	lifts.Put("/:id/status", authMiddleware, liftHandler.SetLiftStatus)

	// Floor routes
	floors := api.Group("/floors")
	floors.Get("/", floorHandler.ListFloors)
	floors.Get("/:floorNum", floorHandler.GetFloorStatus)
	floors.Post("/:floorNum/call", floorHandler.CallLift)
	floors.Post("/:floorNum/reset", authMiddleware, floorHandler.ResetFloorButtons)
	floors.Get("/active-calls", floorHandler.GetActiveFloorCalls)

	// WebSocket route for real-time updates
	app.Get("/ws", ws.WebSocketHandler)
	app.Get("/ws/connect", ws.WebSocketUpgradeHandler(hub))

	// 404 Handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Endpoint not found",
		})
	})
}

// WebSocketHandler handles WebSocket connections for real-time updates
func WebSocketHandler() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// This handler will be called when a client attempts to establish a WebSocket connection
		// We don't need to do anything here, as the actual connection handling is done in WebSocketConnectHandler
	})
}

// WebSocketConnectHandler handles the actual WebSocket connection and communication
func WebSocketConnectHandler(hub *ws.WebSocketHub) func(*websocket.Conn) {
	return func(c *websocket.Conn) {
		// Create a new WebSocketClient for this connection
		client := &ws.WebSocketClient{Conn: c}

		// Register this client with the hub
		hub.Register <- client

		defer func() {
			hub.Unregister <- client
		}()

		// Handle incoming messages
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					// Log the error
				}
				break
			}

			if messageType == websocket.TextMessage {
				// Handle the message if needed
				// For now, we'll just echo it back
				if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
					// Log the error
					break
				}
			}
		}
	}
}
