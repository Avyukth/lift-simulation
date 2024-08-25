package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Avyukth/lift-simulation/internal/application/ports"
	"github.com/Avyukth/lift-simulation/internal/domain"
	"github.com/Avyukth/lift-simulation/pkg/logger"
)

// FloorService handles the business logic for floor operations
type FloorService struct {
	repo     ports.Repository
	eventBus ports.EventBus
	log      *logger.Logger // Dependency on LiftService for assigning lifts
}

// NewFloorService creates a new instance of FloorService
func NewFloorService(repo ports.Repository, eventBus ports.EventBus, log *logger.Logger) *FloorService {
	return &FloorService{
		repo:     repo,
		eventBus: eventBus,
		log:      log,
	}
}

func (s *FloorService) CallLift(ctx context.Context, floorNum int, direction domain.Direction) (*domain.Lift, error) {
	floor, err := s.repo.GetFloorByNumber(ctx, floorNum)
	if err != nil {
		if errors.Is(err, domain.ErrFloorNotFound) {
			return nil, domain.ErrFloorNotFound
		}
		return nil, fmt.Errorf("failed to get floor %d: %w", floorNum, err)
	}

	// Find an available lift, preferably on the ground floor
	lift, err := s.findAvailableLift(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find available lift: %w", err)
	}

	// Start asynchronous process to move the lift
	go s.moveLiftToFloor(context.Background(), lift, floor, direction)

	// Publish an event about the lift call
	event := domain.NewLiftCalledEvent(floor.ID, floorNum, direction)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to publish lift called event: %w", err)
	}

	return lift, nil
}

func (s *FloorService) findAvailableLift(ctx context.Context) (*domain.Lift, error) {
	lifts, err := s.repo.GetAllLifts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get lifts: %w", err)
	}

	var closestLift *domain.Lift
	minDistance := int(^uint(0) >> 1) // Max int value

	for _, lift := range lifts {
		if lift.IsAvailable() {
			distance := abs(lift.CurrentFloor)
			if distance < minDistance {
				minDistance = distance
				closestLift = lift
			}
		}
	}

	if closestLift == nil {
		return nil, fmt.Errorf("no available lift found")
	}

	return closestLift, nil
}

func (s *FloorService) moveLiftToFloor(ctx context.Context, lift *domain.Lift, targetFloor *domain.Floor, direction domain.Direction) {
	// Assign lift to the target floor
	if err := lift.AssignToFloor(targetFloor.Number); err != nil {
		s.log.Error(ctx, "Failed to assign lift to floor", "error", err)
		return
	}
	s.repo.UpdateLift(ctx, lift)

	// Move lift to the target floor
	if err := lift.MoveTo(targetFloor.Number); err != nil {
		s.log.Error(ctx, "Failed to move lift to floor", "error", err)
		return
	}
	s.repo.UpdateLift(ctx, lift)

	// Simulate doors opening and closing
	time.Sleep(2500 * time.Millisecond) // Doors opening
	time.Sleep(2500 * time.Millisecond) // Doors closing

	// Set lift back to available
	lift.SetAvailable()
	s.repo.UpdateLift(ctx, lift)

	// Publish event that lift has arrived
	event := domain.NewLiftArrivedEvent(lift.ID, targetFloor.ID, targetFloor.Number)
	s.eventBus.Publish(ctx, event)
}

// GetFloorStatus retrieves the current status of a floor
func (s *FloorService) GetFloorStatus(ctx context.Context, floorNum int) (*domain.Floor, error) {
	floor, err := s.repo.GetFloorByNumber(ctx, floorNum)
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
	floor, err := s.repo.GetFloorByNumber(ctx, floorNum)
	if err != nil {
		return fmt.Errorf("failed to get floor %d: %w", floorNum, err)
	}

	floor.ResetButtons()

	if err := s.repo.UpdateFloor(ctx, floor); err != nil {
		return fmt.Errorf("failed to update floor %d: %w", floorNum, err)
	}

	// Publish an event about the floor buttons being reset
	event := domain.NewFloorButtonsResetEvent(floor.ID, floorNum)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		return fmt.Errorf("failed to publish floor buttons reset event: %w", err)
	}

	return nil
}

// GetActiveFloorCalls retrieves the numbers of floors with active calls
func (s *FloorService) GetActiveFloorCalls(ctx context.Context) ([]int, error) {
	floors, err := s.repo.ListFloors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list floors: %w", err)
	}

	activeCalls := []int{}
	for _, floor := range floors {
		if floor.HasActiveCall() {
			activeCalls = append(activeCalls, floor.Number)
		}
	}

	return activeCalls, nil
}

// GetFloorByNumber retrieves a floor by its number
func (s *FloorService) GetFloorByNumber(ctx context.Context, floorNum int) (*domain.Floor, error) {
	floor, err := s.repo.GetFloorByNumber(ctx, floorNum)
	if err != nil {
		return nil, fmt.Errorf("failed to get floor %d: %w", floorNum, err)
	}
	return floor, nil
}
