package routes

import (
	"net/http"

	"github.com/Avyukth/lift-simulation/internal/config"
	ws "github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/websockets"

	"github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/middleware"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/swagger"
)

// SetupRoutes configures all the routes for the lift simulation API

func SetupRoutes(config config.RouteConfig) {
	// Middleware
	app := config.App
	liftHandler := config.LiftHandler
	floorHandler := config.FloorHandler
	systemHandler := config.SystemHandler
	hub := config.Hub
	fiberLog := config.FiberLog
	repo := config.Repo

	authConfig := middleware.Config{
		JWTSecret: "your-jwt-secret",
	}
	systemVerification := middleware.NewSystemVerificationMiddleware(repo, fiberLog)
	_ = middleware.New(authConfig)
	app.Use(cors.New())

	// Swagger documentation
	app.Use("/docs", filesystem.New(filesystem.Config{
		Root: http.Dir("./docs"),
	}))

	// Setup Swagger UI
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:         "/docs/swagger.json",
		DeepLinking: true,
	}))

	// API version group
	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	system := api.Group("/system")

	system.Post("/configure", systemHandler.ConfigureSystem)

	system.Get("/configuration", systemHandler.GetSystemConfiguration)
	system.Get("/status", systemHandler.GetSystemStatus)
	system.Post("/reset", systemHandler.ResetSystem)
	system.Get("/metrics", systemHandler.GetSystemMetrics)
	system.Post("/simulate-traffic", systemHandler.SimulateTraffic)

	// Lift routes
	lifts := api.Group("/lifts")
	lifts.Get("/", liftHandler.ListLifts)
	lifts.Put("/reset", liftHandler.ResetLifts)
	lifts.Get("/:id", liftHandler.GetLift)
	lifts.Post("/:id/move", systemVerification.VerifyLiftMove(), liftHandler.MoveLift)
	lifts.Put("/:id/reset", liftHandler.ResetLift)
	lifts.Put("/:id/status", liftHandler.SetLiftStatus)

	// Floor routes
	floors := api.Group("/floors")
	floors.Get("/", floorHandler.ListFloors)
	floors.Get("/active-calls", floorHandler.GetActiveFloorCalls)
	floors.Get("/:floorNum", floorHandler.GetFloorStatus)
	floors.Post("/:floorNum/call", floorHandler.CallLift)
	floors.Post("/:floorNum/reset", floorHandler.ResetFloorButtons)

	// WebSocket route for real-time updates
	// WIP  websocket for emergency call and lift status
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
				if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
					break
				}
			}
		}
	}
}
