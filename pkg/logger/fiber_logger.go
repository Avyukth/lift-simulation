package logger

import (
	"github.com/gofiber/fiber/v2"
)

// FiberLogger wraps the Logger to provide Fiber-specific logging convenience
type FiberLogger struct {
	*Logger
}

// NewFiberLogger creates a new FiberLogger
func NewFiberLogger(logger *Logger) *FiberLogger {
	return &FiberLogger{Logger: logger}
}

// LogWithFiberContext logs a message with the given level and Fiber context
func (fl *FiberLogger) LogWithFiberContext(c *fiber.Ctx, level Level, msg string, args ...any) {
	ctx := c.Context()
	
	// Add Fiber-specific information to the log
	args = append(args, 
		"path", c.Path(),
		"method", c.Method(),
		"ip", c.IP(),
		"user_agent", c.Get("User-Agent"),
	)
	
	switch level {
	case LevelDebug:
		fl.Debug(ctx, msg, args...)
	case LevelInfo:
		fl.Info(ctx, msg, args...)
	case LevelWarn:
		fl.Warn(ctx, msg, args...)
	case LevelError:
		fl.Error(ctx, msg, args...)
	}
}

// DebugFiber logs at LevelDebug with Fiber context
func (fl *FiberLogger) DebugFiber(c *fiber.Ctx, msg string, args ...any) {
	fl.LogWithFiberContext(c, LevelDebug, msg, args...)
}

// InfoFiber logs at LevelInfo with Fiber context
func (fl *FiberLogger) InfoFiber(c *fiber.Ctx, msg string, args ...any) {
	fl.LogWithFiberContext(c, LevelInfo, msg, args...)
}

// WarnFiber logs at LevelWarn with Fiber context
func (fl *FiberLogger) WarnFiber(c *fiber.Ctx, msg string, args ...any) {
	fl.LogWithFiberContext(c, LevelWarn, msg, args...)
}

// ErrorFiber logs at LevelError with Fiber context
func (fl *FiberLogger) ErrorFiber(c *fiber.Ctx, msg string, args ...any) {
	fl.LogWithFiberContext(c, LevelError, msg, args...)
}
