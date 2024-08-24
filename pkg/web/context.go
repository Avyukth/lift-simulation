package web

import (
	"context"
	"io"
	"time"

	"github.com/Avyukth/lift-simulation/pkg/logger"
)

type ctxKey int

const (
	key ctxKey = iota
	loggerKey
)

// Values represent state for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

// GetValues returns the values from the context.
func GetValues(ctx context.Context) *Values {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return &Values{
			TraceID: "00000000-0000-0000-0000-000000000000",
			Now:     time.Now(),
		}
	}

	return v
}

// GetTraceID returns the trace id from the context.
func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}

	return v.TraceID
}

// GetTime returns the time from the context.
func GetTime(ctx context.Context) time.Time {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return time.Now()
	}

	return v.Now
}

func setStatusCode(ctx context.Context, statusCode int) {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return
	}

	v.StatusCode = statusCode
}

func setValues(ctx context.Context, v *Values) context.Context {
	return context.WithValue(ctx, key, v)
}


func SetLogger(ctx context.Context, log *logger.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

// GetLogger retrieves the logger from the context.
func GetLogger(ctx context.Context) *logger.Logger {
	if log, ok := ctx.Value(loggerKey).(*logger.Logger); ok {
		return log
	}
	// Return a no-op logger if not found
	return logger.New(io.Discard, logger.LevelInfo, "NO-OP", nil)
}
