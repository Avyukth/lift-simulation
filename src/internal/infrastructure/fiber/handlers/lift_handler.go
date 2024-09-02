package handlers

import (
	"errors"
	"fmt"

	"github.com/Avyukth/lift-simulation/internal/application/services"
	"github.com/Avyukth/lift-simulation/internal/domain"
	"github.com/gofiber/fiber/v2"
)

// LiftHandler handles HTTP requests related to lifts
type LiftHandler struct {
	liftService *services.LiftService
}

// NewLiftHandler creates a new LiftHandler instance
func NewLiftHandler(liftService *services.LiftService) *LiftHandler {
	return &LiftHandler{
		liftService: liftService,
	}
}

// GetLift handles GET requests to retrieve a specific lift
func (h *LiftHandler) GetLift(c *fiber.Ctx) error {
	liftID := c.Params("id")

	lift, err := h.liftService.GetLiftStatus(c.Context(), liftID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Lift not found",
		})
	}

	return c.JSON(lift)
}

// ListLifts handles GET requests to list all lifts
func (h *LiftHandler) ListLifts(c *fiber.Ctx) error {
	lifts, err := h.liftService.ListLifts(c.Context())
	if err != nil {
		fmt.Println("Failed to retrieve lifts", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve lifts",
		})
	}

	return c.JSON(lifts)
}

// MoveLift handles POST requests to move a lift
func (h *LiftHandler) MoveLift(c *fiber.Ctx) error {
	liftID := c.Params("id")

	var request struct {
		TargetFloor int `json:"targetFloor"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.liftService.MoveLift(c.Context(), liftID, request.TargetFloor)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusAccepted)
}

// SetLiftStatus handles PUT requests to set a lift's status
func (h *LiftHandler) SetLiftStatus(c *fiber.Ctx) error {
	liftID := c.Params("id")

	var request struct {
		Status domain.LiftStatus `json:"status"`
	}

	if err := c.BodyParser(&request); err != nil {
		fmt.Println("Failed to parse request body", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.liftService.SetLiftStatus(c.Context(), liftID, request.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to set lift status",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (h *LiftHandler) ResetLift(c *fiber.Ctx) error {
	liftID := c.Params("id")

	err := h.liftService.ResetLift(c.Context(), liftID)
	if err != nil {

		// Check for specific error types and return appropriate status codes
		switch {
		case errors.Is(err, domain.ErrLiftNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Lift not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to reset lift",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Lift reset successfully",
	})
}

func (h *LiftHandler) ResetLifts(c *fiber.Ctx) error {
	err := h.liftService.ResetLifts(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reset all lifts",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "All lifts reset successfully",
	})
}
