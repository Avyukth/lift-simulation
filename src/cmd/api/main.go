package main

import (
	"context"
	"crypto/tls"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"

	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/Avyukth/lift-simulation/internal/application/services"
	"github.com/Avyukth/lift-simulation/internal/config"
	"github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/handlers"
	"github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/routes"
	ws "github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/websockets"
	"github.com/Avyukth/lift-simulation/internal/infrastructure/persistence/sqlite"
	"github.com/Avyukth/lift-simulation/pkg/logger"
	"github.com/Avyukth/lift-simulation/pkg/web"
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
		return web.GetTraceID(ctx)
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
	// maxProcs := 4
	// runtime.GOMAXPROCS(maxProcs)

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))
	// Log current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Error(ctx, "getting current working directory==========================================", "error", err)
	} else {
		log.Info(ctx, "current working directory==========================================", "path", cwd)
	}

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

	repo, err := sqlite.NewRepository(cfg.DB.Path, log)
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer repo.Close()

	// -------------------------------------------------------------------------
	// Event Bus Support
	log.Info(ctx, "startup", "status", "initializing event bus")

	// -------------------------------------------------------------------------
	// Initialize WebSocket hub

	hub := ws.NewWebSocketHub()
	go hub.Run()

	// -------------------------------------------------------------------------
	// Initialize Services

	liftService := services.NewLiftService(repo, hub, log)
	floorService := services.NewFloorService(repo, log)
	systemService := services.NewSystemService(repo, log)

	liftHandler := handlers.NewLiftHandler(liftService)
	floorHandler := handlers.NewFloorHandler(floorService)

	systemHandler := handlers.NewSystemHandler(systemService)

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Info(ctx, "startup", "status", "debug router started", "host", cfg.Web.DebugHostPort)

		if err := http.ListenAndServe(cfg.Web.DebugHostPort, http.DefaultServeMux); err != nil {
			log.Error(ctx, "shutdown", "status", "debug router closed", "host", cfg.Web.DebugHostPort, "msg", err)
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
	app.Use(cors.New())
	// cors.Config{
	// 	AllowOrigins: "http://localhost:6000", // Replace with your web app's origin
	// 	AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	// }))
	routeConfig := config.RouteConfig{
		App:           app,
		LiftHandler:   liftHandler,
		FloorHandler:  floorHandler,
		SystemHandler: systemHandler,
		Hub:           hub,
		FiberLog:      fiberLog,
		Repo:          repo,
	}

	routes.SetupRoutes(routeConfig)
	// Add a test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("API is working")
	})

	// -------------------------------------------------------------------------
	// Start API Service

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "http router started", "host", cfg.Web.HTTPHostPort)
		err := app.Listen(cfg.Web.HTTPHostPort)
		if err != nil {
			serverErrors <- errors.Wrap(err, "http server error")
		}
	}()

	// Start HTTPS server if TLS is configured
	if cfg.Web.CertFile != "" && cfg.Web.KeyFile != "" {
		go func() {
			log.Info(ctx, "startup", "status", "https router started", "host", cfg.Web.HTTPSHostPort)
			cert, err := tls.LoadX509KeyPair(cfg.Web.CertFile, cfg.Web.KeyFile)
			if err != nil {
				serverErrors <- errors.Wrap(err, "loading ssl certificates")
				return
			}

			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
				MinVersion:   tls.VersionTLS12,
			}

			ln, err := tls.Listen("tcp", cfg.Web.HTTPSHostPort, tlsConfig)
			if err != nil {
				serverErrors <- errors.Wrap(err, "creating https listener")
				return
			}

			err = app.Listener(ln)
			if err != nil {
				serverErrors <- errors.Wrap(err, "https server error")
			}
		}()
	}

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
