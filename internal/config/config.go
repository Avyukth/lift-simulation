package config

import (
	"fmt"
	"time"

	"github.com/ardanlabs/conf/v3"
)

// Config holds all configuration for our program
type Config struct {
	conf.Version
	Web struct {
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:10s"`
		IdleTimeout     time.Duration `conf:"default:120s"`
		ShutdownTimeout time.Duration `conf:"default:20s"`
		APIHost         string        `conf:"default:0.0.0.0:3000"`
		DebugHost       string        `conf:"default:0.0.0.0:4000"`
	}
	Auth struct {
		KeysFolder string `conf:"default:zarf/keys/"`
		ActiveKID  string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
		Issuer     string `conf:"default:lift simulation project"`
	}
	DB struct {
		Path          string        `conf:"default:./db/lift_simulation.sqlite"`
		MaxOpenConns  int           `conf:"default:10"`
		MaxIdleConns  int           `conf:"default:5"`
		MaxLifetime   time.Duration `conf:"default:1h"`
	}
	Redis struct {
		URL      string        `conf:"default:redis://localhost:6379,mask"`
		Password string        `conf:"default:redispassword,mask"`
		DB       int           `conf:"default:0"`
		PoolSize int           `conf:"default:10"`
	}
	Lift struct {
		MaxFloors     int `conf:"default:50"`
		MaxLifts      int `conf:"default:10"`
		FloorTripTime int `conf:"default:2"`
	}
}

// LoadConfig reads configuration from environment variables.
func LoadConfig(build string) (Config, error) {
	cfg := Config{
		Version: conf.Version{
			Build: build,
			Desc:  "Lift Simulation",
		},
	}

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
