package config

import (
	"fmt"
	"os"
	"time"

	"github.com/Avyukth/lift-simulation/internal/application/ports"
	"github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/handlers"
	ws "github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/websockets"
	"github.com/Avyukth/lift-simulation/pkg/logger"
	"github.com/ardanlabs/conf/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

// Config holds all configuration for our program
type Config struct {
	conf.Version
	Web struct {
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:10s"`
		IdleTimeout     time.Duration `conf:"default:120s"`
		ShutdownTimeout time.Duration `conf:"default:20s"`
		APIHost         string        `conf:"default:0.0.0.0:4000"`
		DebugHost       string        `conf:"default:0.0.0.0:5000"`
	}
	Auth struct {
		KeysFolder string `conf:"default:zarf/keys/"`
		ActiveKID  string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
		Issuer     string `conf:"default:lift simulation project"`
		JWTSecret  string
	}
	DB struct {
		Path         string        `conf:"default:./db/lift_simulation.sqlite"`
		MaxOpenConns int           `conf:"default:10"`
		MaxIdleConns int           `conf:"default:5"`
		MaxLifetime  time.Duration `conf:"default:1h"`
	}
	Redis struct {
		Host     string `conf:"default:localhost"`
		Port     int    `conf:"default:6379"`
		Password string
		DB       int `conf:"default:0"`
		PoolSize int `conf:"default:10"`
	}
	Lift struct {
		MaxFloors     int `conf:"default:50"`
		MaxLifts      int `conf:"default:10"`
		FloorTripTime int `conf:"default:2"`
	}
	API struct {
		Port   int
		Secret string
	}
	LogLevel string `conf:"default:info"`
}

type RouteConfig struct {
	App           *fiber.App
	LiftHandler   *handlers.LiftHandler
	FloorHandler  *handlers.FloorHandler
	SystemHandler *handlers.SystemHandler
	Hub           *ws.WebSocketHub
	FiberLog      *logger.FiberLogger
	Repo          ports.Repository
}

// LoadConfig reads configuration from environment variables and .env file.
func LoadConfig(build string) (Config, error) {
	cfg := Config{
		Version: conf.Version{
			Build: build,
			Desc:  "Lift Simulation",
		},
	}

	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}

	viper.SetConfigFile(fmt.Sprintf("src/.env.%s", env))
	viper.AutomaticEnv()
	// Load .env file
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("reading .env file: %w", err)
	}

	// Override config with values from .env
	cfg.Auth.JWTSecret = viper.GetString("JWT_SECRET")
	cfg.Redis.Password = viper.GetString("REDIS_PASSWORD")
	cfg.API.Port = viper.GetInt("API_PORT")
	cfg.API.Secret = viper.GetString("API_SECRET")
	cfg.Redis.Host = viper.GetString("REDIS_HOST")
	cfg.Redis.Port = viper.GetInt("REDIS_PORT")
	cfg.LogLevel = viper.GetString("LOG_LEVEL")
	cfg.DB.Path = viper.GetString("DB_PATH")

	// Parse the rest of the configuration
	const prefix = "LIFT"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if err == conf.ErrHelpWanted {
			fmt.Println(help)
			return cfg, nil
		}
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

// String returns the configuration as a string.
func (cfg *Config) String() (string, error) {
	out, err := conf.String(cfg)
	if err != nil {
		return "", fmt.Errorf("generating config for output: %w", err)
	}
	return out, nil
}
