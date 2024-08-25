package services

import (
	"context"
	"fmt"

	"github.com/Avyukth/lift-simulation/internal/application/ports"
	"github.com/Avyukth/lift-simulation/internal/domain"
	"github.com/Avyukth/lift-simulation/pkg/logger"
	"github.com/google/uuid"
)

// SystemService handles the business logic for overall system operations
type SystemService struct {
	repo     ports.SystemRepository
	eventBus ports.EventBus
	log      *logger.Logger
}

// NewSystemService creates a new instance of SystemService
func NewSystemService(repo ports.SystemRepository, eventBus ports.EventBus, log *logger.Logger) *SystemService {
	return &SystemService{
		repo:     repo,
		eventBus: eventBus,
		log:      log,
	}
}

// ConfigureSystem sets up the lift system with the specified number of floors and lifts
func (s *SystemService) ConfigureSystem(ctx context.Context, floors, lifts int) error {
	if floors < 2 {
		return fmt.Errorf("invalid number of floors: must be at least 2")
	}
	if lifts < 1 {
		return fmt.Errorf("invalid number of lifts: must be at least 1")
	}
	if lifts > floors {
		return fmt.Errorf("invalid number of lifts: must be less than number of floors")
	}
	s.log.Info(ctx, "Configuring system", "total_floors", floors, "total_lifts", lifts)

	// Create a new system
	systemID := uuid.New().String()
	system, err := domain.NewSystem(systemID, floors, lifts)
	if err != nil {
		s.log.Error(ctx, "Failed to create system configuration", "error", err)
		return fmt.Errorf("failed to create system configuration: %w", err)
	}

	// Save the system configuration
	if err := s.repo.SaveSystem(ctx, system); err != nil {
		s.log.Error(ctx, "Failed to save system configuration", "error", err)
		return fmt.Errorf("failed to save system configuration: %w", err)
	}

	// Initialize floors (0-based)
	for i := 0; i < floors; i++ {
		floorID := uuid.New().String()
		floor := domain.NewFloor(floorID, i)
		if err := s.repo.SaveFloor(ctx, floor); err != nil {
			s.log.Error(ctx, "Failed to save floor", "floor_number", i, "error", err)
			return fmt.Errorf("failed to save floor %d: %w", i, err)
		}
		s.log.Debug(ctx, "Floor created", "floor_id", floorID, "floor_number", i)
	}

	// Initialize lifts
	for i := 1; i <= lifts; i++ {
		liftID := uuid.New().String()
		liftName := fmt.Sprintf("L%d", i)
		lift := domain.NewLift(liftID, liftName)
		if err := s.repo.SaveLift(ctx, lift); err != nil {
			s.log.Error(ctx, "Failed to save lift", "lift_name", liftName, "error", err)
			return fmt.Errorf("failed to save lift %s: %w", liftName, err)
		}
		s.log.Debug(ctx, "Lift created", "lift_id", liftID, "lift_name", liftName)
	}

	// Publish an event about the system configuration
	event := domain.NewSystemConfiguredEvent(systemID, floors, lifts)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		s.log.Error(ctx, "Failed to publish system configured event", "error", err)
		return fmt.Errorf("failed to publish system configured event: %w", err)
	}

	s.log.Info(ctx, "System configuration completed successfully",
		"system_id", systemID,
		"total_floors", floors,
		"total_lifts", lifts)
	return nil
}

// GetSystemConfiguration retrieves the current system configuration
func (s *SystemService) GetSystemConfiguration(ctx context.Context) (*domain.System, error) {
	s.log.Info(ctx, "Getting system configuration")

	system, err := s.repo.GetSystem(ctx)
	if err != nil {
		s.log.Error(ctx, "Failed to get system configuration", "error", err)
		return nil, fmt.Errorf("failed to get system configuration: %w", err)
	}

	s.log.Info(ctx, "Retrieved system configuration",
		"system_id", system.ID,
		"total_floors", system.TotalFloors,
		"total_lifts", system.TotalLifts)
	return system, nil
}

// GetSystemStatus retrieves the overall status of the lift system
func (s *SystemService) GetSystemStatus(ctx context.Context) (*domain.SystemStatus, error) {
	system, err := s.repo.GetSystem(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system configuration: %w", err)
	}

	lifts, err := s.repo.GetAllLifts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get lifts: %w", err)
	}

	floors, err := s.repo.GetAllFloors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get floors: %w", err)
	}

	status := &domain.SystemStatus{
		SystemID:         system.ID,
		TotalFloors:      system.TotalFloors,
		TotalLifts:       system.TotalLifts,
		OperationalLifts: len(lifts),
		ActiveFloorCalls: countActiveFloorCalls(floors),
	}

	return status, nil
}

// ResetSystem resets the entire lift system to its initial state
func (s *SystemService) ResetSystem(ctx context.Context) error {
	system, err := s.repo.GetSystem(ctx)
	if err != nil {
		return fmt.Errorf("failed to get system configuration: %w", err)
	}

	// Reset all lifts to ground floor (now floor 0)
	lifts, err := s.repo.GetAllLifts(ctx)
	s.log.Info(ctx, "Resetting system", "lifts", lifts)
	if err != nil {
		return fmt.Errorf("failed to get lifts: %w", err)
	}

	for _, lift := range lifts {
		lift.Reset() // This should now set CurrentFloor to 0
		if err := s.repo.SaveLift(ctx, lift); err != nil {
			return fmt.Errorf("failed to save reset lift %s: %w", lift.ID, err)
		}
	}

	// Reset all floor buttons
	// for i := 0; i < system.TotalFloors; i++ {
	// 	floor, err := s.repo.GetFloorByNumber(ctx, i)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to get floor %d: %w", i, err)
	// 	}
	// 	floor.ResetButtons()
	// 	if err := s.repo.SaveFloor(ctx, floor); err != nil {
	// 		return fmt.Errorf("failed to save reset floor %d: %w", i, err)
	// 	}
	// }

	// Publish an event about the system reset
	event := domain.NewSystemResetEvent(system.ID)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		return fmt.Errorf("failed to publish system reset event: %w", err)
	}

	return nil
}

// GetSystemMetrics retrieves various metrics about the system
func (s *SystemService) GetSystemMetrics(ctx context.Context) (map[string]interface{}, error) {
	system, err := s.repo.GetSystem(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system: %w", err)
	}

	lifts, err := s.repo.GetAllLifts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get lifts: %w", err)
	}

	metrics := map[string]interface{}{
		"systemID":          system.ID,
		"totalFloors":       system.TotalFloors,
		"totalLifts":        system.TotalLifts,
		"availableLifts":    countAvailableLifts(lifts),
		"occupiedLifts":     countOccupiedLifts(lifts),
		"outOfServiceLifts": countOutOfServiceLifts(lifts),
	}

	return metrics, nil
}

// SimulateTraffic simulates traffic in the system based on the given intensity and duration
func (s *SystemService) SimulateTraffic(ctx context.Context, duration int, intensity string) error {
	system, err := s.repo.GetSystem(ctx)
	if err != nil {
		return fmt.Errorf("failed to get system: %w", err)
	}

	switch intensity {
	case "low", "medium", "high":
		// Simulate traffic based on intensity and duration
		event := domain.NewTrafficSimulationEvent(system.ID, intensity, duration)
		if err := s.eventBus.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish traffic simulation event: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("invalid traffic intensity: %s", intensity)
	}
}

// Helper function to count active floor calls
func countActiveFloorCalls(floors []*domain.Floor) int {
	count := 0
	for _, floor := range floors {
		if floor.HasActiveCall() {
			count++
		}
	}
	return count
}

// Helper functions for GetSystemMetrics
func countAvailableLifts(lifts []*domain.Lift) int {
	count := 0
	for _, lift := range lifts {
		if lift.Status == domain.Available {
			count++
		}
	}
	return count
}

func countOccupiedLifts(lifts []*domain.Lift) int {
	count := 0
	for _, lift := range lifts {
		if lift.Status == domain.Occupied {
			count++
		}
	}
	return count
}

func countOutOfServiceLifts(lifts []*domain.Lift) int {
	count := 0
	for _, lift := range lifts {
		if lift.Status == domain.OutOfService {
			count++
		}
	}
	return count
}
