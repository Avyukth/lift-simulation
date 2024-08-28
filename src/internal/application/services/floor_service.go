package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Avyukth/lift-simulation/internal/application/events"
	"github.com/Avyukth/lift-simulation/internal/application/ports"
	"github.com/Avyukth/lift-simulation/internal/domain"
	"github.com/Avyukth/lift-simulation/pkg/logger"
)

// FloorService handles the business logic for floor operations
type FloorService struct {
	repo     ports.FloorOperations
	eventBus events.EventBus
	log      *logger.Logger
}

type LiftAssignedHandler struct {
	service *FloorService
}
type LiftArrivedHandler struct {
	service *FloorService
}

// NewFloorService creates a new instance of FloorService
func NewFloorService(repo ports.FloorOperations, eventBus events.EventBus, log *logger.Logger) *FloorService {
	service := &FloorService{
		repo:     repo,
		eventBus: eventBus,
		log:      log,
	}
	eventBus.Subscribe(domain.LiftAssigned, &LiftAssignedHandler{service: service})
	eventBus.Subscribe(domain.LiftArrived, &LiftArrivedHandler{service: service})
	return service

}
func (h *LiftAssignedHandler) Handle(event domain.Event) {
	if liftAssignedEvent, ok := event.(domain.LiftAssignedEvent); ok {
		h.service.handleLiftAssigned(context.Background(), liftAssignedEvent.FloorNumber, liftAssignedEvent.LiftID)
	}
}

func (h *LiftArrivedHandler) Handle(event domain.Event) {
	if liftArrivedEvent, ok := event.(domain.LiftArrivedEvent); ok {
		h.service.handleLiftArrived(context.Background(), liftArrivedEvent.FloorNumber, liftArrivedEvent.LiftID)
	}
}

func (s *FloorService) handleLiftAssigned(ctx context.Context, floorNum int, liftID string) {
	s.log.Info(ctx, "Lift assigned to floor", "floor", floorNum, "lift_id", liftID)
	if err := s.updateFloorDisplay(ctx, floorNum, liftID); err != nil {
		s.log.Error(ctx, "Failed to update floor display", "floor", floorNum, "lift_id", liftID, "error", err)
	}
}

func (s *FloorService) handleLiftArrived(ctx context.Context, floorNum int, liftID string) {
	s.log.Info(ctx, "Lift arrived at floor", "floor", floorNum, "lift_id", liftID)
	if err := s.ResetFloorButtons(ctx, floorNum); err != nil {
		s.log.Error(ctx, "Failed to reset floor buttons", "floor", floorNum, "error", err)
	}
	if err := s.updateFloorDisplay(ctx, floorNum, liftID); err != nil {
		s.log.Error(ctx, "Failed to update floor display", "floor", floorNum, "lift_id", liftID, "error", err)
	}
}

func (s *FloorService) updateFloorDisplay(ctx context.Context, floorNum int, liftID string) error {
	// For now, we'll just log the information
	s.log.Info(ctx, "Floor display updated", "floor", floorNum, "assigned_lift", liftID)
	return nil
}

func (s *FloorService) CallLift(ctx context.Context, floorNum int, direction domain.Direction) error {
	floor, err := s.repo.GetFloorByNumber(ctx, floorNum)
	if err != nil {
		if errors.Is(err, domain.ErrFloorNotFound) {
			return domain.ErrFloorNotFound
		}
		return fmt.Errorf("failed to get floor %d: %w", floorNum, err)
	}

	// Publish a LiftRequested event
	event := domain.LiftRequestedEvent{
		FloorNumber: floorNum,
		Direction:   direction,
	}
	s.eventBus.Publish(event)

	s.log.Info(ctx, "Lift requested", "floor", floor.Number, "direction", direction)
	return nil
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
