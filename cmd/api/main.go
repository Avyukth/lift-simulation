package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"

	"github.com/Avyukth/lift-simulation/internal/application/services"
	"github.com/Avyukth/lift-simulation/internal/config"
	"github.com/Avyukth/lift-simulation/internal/infrastructure/eventbus"
	"github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/handlers"
	"github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/routes"
	"github.com/Avyukth/lift-simulation/internal/infrastructure/persistence/sqlite"
	"github.com/Avyukth/lift-simulation/pkg/logger"
)

var build = "develop"

func main() {
	ctx := context.Background()

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			fmt.Println("******* SEND ALERT *******")
		},
	}

	traceIDFn := func(ctx context.Context) string {
		return "00000000-0000-0000-0000-000000000000"
	}

	log := logger.NewWithEvents(os.Stdout, logger.LevelInfo, "LIFT-SIMULATION", traceIDFn, events)
	fiberLog := logger.NewFiberLogger(log)

	if err := run(ctx, log, fiberLog); err != nil {
		log.Error(ctx, "startup", "msg", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger, fiberLog *logger.FiberLogger) error {
	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// Configuration

	cfg, err := config.LoadConfig(build)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// -------------------------------------------------------------------------
	// App Starting

	log.Info(ctx, "starting service", "version", build)
	defer log.Info(ctx, "shutdown complete")

	expvar.NewString("build").Set(build)

	// -------------------------------------------------------------------------
	// Database Support

	log.Info(ctx, "startup", "status", "initializing database support", "path", cfg.DB.Path)

	repo, err := sqlite.NewRepository(cfg.DB.Path)
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer repo.Close()

	// -------------------------------------------------------------------------
	// Event Bus Support

	log.Info(ctx, "startup", "status", "initializing event bus")

	eventBus, err := eventbus.NewRedisPubSub(cfg.Redis.URL)
	if err != nil {
		return fmt.Errorf("connecting to Redis: %w", err)
	}
	defer eventBus.Close()

	// -------------------------------------------------------------------------
	// Initialize WebSocket hub

	hub := handlers.NewWebSocketHub()
	go hub.Run()

	// -------------------------------------------------------------------------
	// Initialize Services

	liftService := services.NewLiftService(repo, eventBus)
	floorService := services.NewFloorService(repo, eventBus, liftService)
	systemService := services.NewSystemService(repo, eventBus)

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Info(ctx, "startup", "status", "debug router started", "host", cfg.Web.DebugHost)

		if err := http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux); err != nil {
			log.Error(ctx, "shutdown", "status", "debug router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start API Service

	log.Info(ctx, "startup", "status", "initializing API support")

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorHandler: customErrorHandler(fiberLog),
	})

	app.Use(recover.New())

	routes.SetupRoutes(app, liftService, floorService, systemService, hub, fiberLog)
	app.Get("/swagger/*", swagger.HandlerDefault)

	// -------------------------------------------------------------------------
	// Start API Service

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", cfg.Web.APIHost)
		serverErrors <- app.Listen(cfg.Web.APIHost)
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			app.Shutdown()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

func customErrorHandler(fiberLog *logger.FiberLogger) func(*fiber.Ctx, error) error {
	return func(c *fiber.Ctx, err error) error {
		fiberLog.ErrorFiber(c, "request error", "error", err, "path", c.Path())

		code := fiber.StatusInternalServerError
		msg := "Internal Server Error"

		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			msg = e.Message
		}

		return c.Status(code).JSON(fiber.Map{
			"error": msg,
		})
	}
}
