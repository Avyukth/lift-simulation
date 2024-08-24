package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/Avyukth/lift-simulation/internal/application/services"
)

// SystemHandler handles HTTP requests related to the overall lift system
type SystemHandler struct {
	systemService *services.SystemService
}

// NewSystemHandler creates a new SystemHandler instance
func NewSystemHandler(systemService *services.SystemService) *SystemHandler {
	return &SystemHandler{
		systemService: systemService,
	}
}

// ConfigureSystem handles POST requests to configure the lift system
func (h *SystemHandler) ConfigureSystem(c *fiber.Ctx) error {
	var config struct {
		Floors int `json:"floors"`
		Lifts  int `json:"lifts"`
	}

	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.systemService.ConfigureSystem(c.Context(), config.Floors, config.Lifts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to configure system",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "System configured successfully",
	})
}

// GetSystemConfiguration handles GET requests to retrieve the current system configuration
func (h *SystemHandler) GetSystemConfiguration(c *fiber.Ctx) error {
	config, err := h.systemService.GetSystemConfiguration(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve system configuration",
		})
	}

	return c.JSON(config)
}

// GetSystemStatus handles GET requests to retrieve the overall system status
func (h *SystemHandler) GetSystemStatus(c *fiber.Ctx) error {
	status, err := h.systemService.GetSystemStatus(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve system status",
		})
	}

	return c.JSON(status)
}

// ResetSystem handles POST requests to reset the entire lift system
func (h *SystemHandler) ResetSystem(c *fiber.Ctx) error {
	err := h.systemService.ResetSystem(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reset system",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "System reset successfully",
	})
}

// GetSystemMetrics handles GET requests to retrieve system performance metrics
func (h *SystemHandler) GetSystemMetrics(c *fiber.Ctx) error {
	metrics, err := h.systemService.GetSystemMetrics(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve system metrics",
		})
	}

	return c.JSON(metrics)
}

// SimulateTraffic handles POST requests to simulate lift traffic in the system
func (h *SystemHandler) SimulateTraffic(c *fiber.Ctx) error {
	var request struct {
		Duration  int    `json:"duration"`
		Intensity string `json:"intensity"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.systemService.SimulateTraffic(c.Context(), request.Duration, request.Intensity)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to simulate traffic",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Traffic simulation started",
	})
}
