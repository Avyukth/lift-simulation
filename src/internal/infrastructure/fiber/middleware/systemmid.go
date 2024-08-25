package middleware

import (
	"encoding/json"

	"github.com/Avyukth/lift-simulation/internal/application/ports"
	"github.com/Avyukth/lift-simulation/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type SystemVerificationMiddleware struct {
	repo ports.Repository
	log  *logger.FiberLogger
}

type MoveLiftPayload struct {
	TargetFloor int `json:"targetFloor"`
}

func NewSystemVerificationMiddleware(repo ports.Repository, log *logger.FiberLogger) *SystemVerificationMiddleware {
	return &SystemVerificationMiddleware{
		repo: repo,
		log:  log,
	}
}

func (m *SystemVerificationMiddleware) VerifySystem() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		system, err := m.repo.GetSystem(ctx)
		if err != nil {
			m.log.Error(ctx, "Failed to retrieve system configuration", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to verify system configuration",
			})
		}

		if system == nil || system.TotalFloors == 0 || system.TotalLifts == 0 {
			m.log.Warn(ctx, "System not properly configured")
			return c.Status(fiber.StatusPreconditionFailed).JSON(fiber.Map{
				"error": "System not properly configured",
			})
		}
		return c.Next()
	}
}

func (m *SystemVerificationMiddleware) VerifyLiftMove() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		liftID := c.Params("id")

		// Verify lift exists
		lift, err := m.repo.GetLift(ctx, liftID)
		if err != nil {
			m.log.Error(ctx, "Failed to retrieve lift", "lift_id", liftID, "error", err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Lift not found",
			})
		}

		if lift == nil {
			m.log.Warn(ctx, "Lift not found", "lift_id", liftID)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Lift not found",
			})
		}

		// Verify and parse payload
		var payload MoveLiftPayload
		if err := json.Unmarshal(c.Body(), &payload); err != nil {
			m.log.Error(ctx, "Failed to parse move lift payload", "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid payload",
			})
		}

		// Verify target floor is valid
		system, err := m.repo.GetSystem(ctx)
		if err != nil {
			m.log.Error(ctx, "Failed to retrieve system configuration", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to verify system configuration",
			})
		}

		if payload.TargetFloor < 0 || payload.TargetFloor >= system.TotalFloors {
			m.log.Warn(ctx, "Invalid target floor", "target_floor", payload.TargetFloor, "total_floors", system.TotalFloors)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid target floor",
			})
		}

		// If everything is valid, add the parsed payload to the context for the next handler
		c.Locals("moveLiftPayload", payload)

		// Continue to the next middleware/handler
		return c.Next()
	}
}
