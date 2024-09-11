package handlers

import (
	"errors"

	"github.com/Avyukth/lift-simulation/internal/application/services"
	"github.com/Avyukth/lift-simulation/internal/domain"
	"github.com/gofiber/fiber/v2"
)

// FloorHandler handles HTTP requests related to floors
type FloorHandler struct {
	floorService *services.FloorService
}

// NewFloorHandler creates a new FloorHandler instance
func NewFloorHandler(floorService *services.FloorService) *FloorHandler {
	return &FloorHandler{
		floorService: floorService,
	}
}

// ListFloors handles GET requests to list all floors
func (h *FloorHandler) ListFloors(c *fiber.Ctx) error {
	floors, err := h.floorService.ListFloors(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve floors",
		})
	}

	return c.JSON(floors)
}

// GetFloorStatus handles GET requests to retrieve the status of a specific floor
func (h *FloorHandler) GetFloorStatus(c *fiber.Ctx) error {
	floorNum, err := c.ParamsInt("floorNum")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid floor number",
		})
	}

	floor, err := h.floorService.GetFloorStatus(c.Context(), floorNum)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Floor not found",
		})
	}

	return c.JSON(floor)
}

// CallLift handles POST requests to call a lift to a specific floor
func (h *FloorHandler) CallLift(c *fiber.Ctx) error {
	floorNum, err := c.ParamsInt("floorNum")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid floor number",
		})
	}

	var request struct {
		Direction domain.Direction `json:"direction"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if request.Direction < domain.Up || request.Direction > domain.Down {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid direction. Must be 0 (Up), 1 (Down)",
		})
	}

	err = h.floorService.CallLift(c.Context(), floorNum, request.Direction)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrFloorNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Floor not found",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to call lift",
				"details": err.Error(),
			})
		}
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "Lift call accepted. Lift is on its way",
	})
}

// ResetFloorButtons handles POST requests to reset the call buttons on a floor
func (h *FloorHandler) ResetFloorButtons(c *fiber.Ctx) error {
	floorNum, err := c.ParamsInt("floorNum")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid floor number",
		})
	}

	err = h.floorService.ResetFloorButtons(c.Context(), floorNum)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reset floor buttons",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

// GetActiveFloorCalls handles GET requests to retrieve all active floor calls
func (h *FloorHandler) GetActiveFloorCalls(c *fiber.Ctx) error {
	activeFloorCalls, err := h.floorService.GetActiveFloorCalls(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve active floor calls",
		})
	}

	return c.JSON(activeFloorCalls)
}
