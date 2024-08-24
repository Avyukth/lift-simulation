package services

import (
	"context"
	"fmt"

	"github.com/Avyukth/lift-simulation/internal/domain"
	"github.com/Avyukth/lift-simulation/internal/application/ports"
)

// FloorService handles the business logic for floor operations
type FloorService struct {
	repo     ports.FloorRepository
	eventBus ports.EventBus
	liftSvc  *LiftService // Dependency on LiftService for assigning lifts
}

// NewFloorService creates a new instance of FloorService
func NewFloorService(repo ports.FloorRepository, eventBus ports.EventBus, liftSvc *LiftService) *FloorService {
	return &FloorService{
		repo:     repo,
		eventBus: eventBus,
		liftSvc:  liftSvc,
	}
}

// CallLift requests a lift to a specific floor
func (s *FloorService) CallLift(ctx context.Context, floorNum int, direction domain.Direction) error {
	floor, err := s.repo.GetFloor(ctx, floorNum)
	if err != nil {
		return fmt.Errorf("failed to get floor %d: %w", floorNum, err)
	}

	if err := floor.RequestLift(direction); err != nil {
		return fmt.Errorf("failed to request lift for floor %d: %w", floorNum, err)
	}

	if err := s.repo.UpdateFloor(ctx, floor); err != nil {
		return fmt.Errorf("failed to update floor %d: %w", floorNum, err)
	}

	// Assign a lift to this floor request
	// assignedLift, err := s.liftSvc.AssignLiftToFloor(ctx, floorNum, direction)
	if err != nil {
		return fmt.Errorf("failed to assign lift to floor %d: %w", floorNum, err)
	}

	// Publish an event about the lift call
	// event := domain.NewLiftCalledEvent(floorNum, direction, assignedLift.ID())
	event := domain.NewLiftCalledEvent(floorNum, direction)

	if err := s.eventBus.Publish(ctx, event); err != nil {
		return fmt.Errorf("failed to publish lift called event: %w", err)
	}

	return nil
}

// GetFloorStatus retrieves the current status of a floor
func (s *FloorService) GetFloorStatus(ctx context.Context, floorNum int) (*domain.Floor, error) {
	floor, err := s.repo.GetFloor(ctx, floorNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get floor %d: %w", floorNum, err)
	}
	return floor, nil
}

// ListFloors retrieves all floors in the system
func (s *FloorService) ListFloors(ctx context.Context) ([]*domain.Floor, error) {
	floors, err := s.repo.ListFloors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list floors: %w", err)
	}
	return floors, nil
}

// ResetFloorButtons resets the call buttons on a floor after a lift has arrived
func (s *FloorService) ResetFloorButtons(ctx context.Context, floorNum int) error {
	floor, err := s.repo.GetFloor(ctx, floorNum)
	if err != nil {
		return fmt.Errorf("failed to get floor %d: %w", floorNum, err)
	}

	floor.ResetButtons()

	if err := s.repo.UpdateFloor(ctx, floor); err != nil {
		return fmt.Errorf("failed to update floor %d: %w", floorNum, err)
	}

	// Publish an event about the floor buttons being reset
	event := domain.NewFloorButtonsResetEvent(floorNum)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		return fmt.Errorf("failed to publish floor buttons reset event: %w", err)
	}

	return nil
}
