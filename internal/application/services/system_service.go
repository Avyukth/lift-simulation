package services

import (
	"context"
	"fmt"

	"github.com/Avyukth/lift-simulation/internal/domain"
	"github.com/Avyukth/lift-simulation/internal/application/ports"
)

// SystemService handles the business logic for overall system operations
type SystemService struct {
	repo     ports.SystemRepository
	eventBus ports.EventBus
}

// NewSystemService creates a new instance of SystemService
func NewSystemService(repo ports.SystemRepository, eventBus ports.EventBus) *SystemService {
	return &SystemService{
		repo:     repo,
		eventBus: eventBus,
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
		return fmt.Errorf("invalid number of lifts: must be less than no of floor")
	}

	system, err := domain.NewSystem(floors, lifts)
	if err != nil{
		return fmt.Errorf("failed to create system configuration: %w", err)
	}

	if err := s.repo.SaveSystem(ctx, system); err != nil {
		return fmt.Errorf("failed to save system configuration: %w", err)
	}

	// Initialize floors
	for i := 1; i <= floors; i++ {
		floor := domain.NewFloor(i)
		if err := s.repo.SaveFloor(ctx, floor); err != nil {
			return fmt.Errorf("failed to save floor %d: %w", i, err)
		}
	}

	// Initialize lifts
	for i := 1; i <= lifts; i++ {
		lift := domain.NewLift(fmt.Sprintf("L%d", i))
		if err := s.repo.SaveLift(ctx, lift); err != nil {
			return fmt.Errorf("failed to save lift %d: %w", i, err)
		}
	}

	// Publish an event about the system configuration
	event := domain.NewSystemConfiguredEvent(floors, lifts)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		return fmt.Errorf("failed to publish system configured event: %w", err)
	}

	return nil
}

// GetSystemConfiguration retrieves the current system configuration
func (s *SystemService) GetSystemConfiguration(ctx context.Context) (*domain.System, error) {
	system, err := s.repo.GetSystem(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system configuration: %w", err)
	}
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
		TotalFloors:    system.TotalFloors(),
		TotalLifts:     system.TotalLifts(),
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

	// Reset all lifts to ground floor
	for i := 1; i <= system.TotalLifts(); i++ {
		lift, err := s.repo.GetLift(ctx, fmt.Sprintf("L%d", i))
		if err != nil {
			return fmt.Errorf("failed to get lift %d: %w", i, err)
		}
		lift.Reset()
		if err := s.repo.SaveLift(ctx, lift); err != nil {
			return fmt.Errorf("failed to save reset lift %d: %w", i, err)
		}
	}

	// Reset all floor buttons
	for i := 1; i <= system.TotalFloors(); i++ {
		floor, err := s.repo.GetFloor(ctx, i)
		if err != nil {
			return fmt.Errorf("failed to get floor %d: %w", i, err)
		}
		floor.ResetButtons()
		if err := s.repo.SaveFloor(ctx, floor); err != nil {
			return fmt.Errorf("failed to save reset floor %d: %w", i, err)
		}
	}

	// Publish an event about the system reset
	event := domain.NewSystemResetEvent()
	if err := s.eventBus.Publish(ctx, event); err != nil {
		return fmt.Errorf("failed to publish system reset event: %w", err)
	}

	return nil
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
