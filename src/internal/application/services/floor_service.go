package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/Avyukth/lift-simulation/internal/application/events"
	"github.com/Avyukth/lift-simulation/internal/application/ports"
	"github.com/Avyukth/lift-simulation/internal/domain"
	"github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/websockets"
	"github.com/Avyukth/lift-simulation/pkg/logger"
)

// FloorService handles the business logic for floor operations
type FloorService struct {
	repo     ports.FloorOperations
	eventBus events.EventBus
	log      *logger.Logger
	hub      *websockets.WebSocketHub
}

type LiftAssignedHandler struct {
	service *FloorService
}
type LiftArrivedHandler struct {
	service *FloorService
}

// NewFloorService creates a new instance of FloorService
func NewFloorService(repo ports.FloorOperations, eventBus events.EventBus, log *logger.Logger, hub *websockets.WebSocketHub) *FloorService {
	service := &FloorService{
		repo:     repo,
		eventBus: eventBus,
		log:      log,
		hub:      hub,
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
		h.service.handleLiftArrived(context.Background(), liftArrivedEvent.LiftID, liftArrivedEvent.FloorNumber)
	}
}

func (s *FloorService) handleLiftAssigned(ctx context.Context, floorNum int, liftID string) {
	s.log.Info(ctx, "Lift assigned to floor", "floor", floorNum, "lift_id", liftID)
	if err := s.updateFloorDisplay(ctx, floorNum, liftID); err != nil {
		s.log.Error(ctx, "Failed to update floor display", "floor", floorNum, "lift_id", liftID, "error", err)
	}
}

func (s *FloorService) handleLiftArrived(ctx context.Context, liftID string, floorNum int) {
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
	s.sendWebSocketUpdate(ctx, "floor", strconv.Itoa(floorNum), fmt.Sprintf("display_updated:%s", liftID), floorNum)
	return nil
}

func (s *FloorService) sendWebSocketUpdate(ctx context.Context, updateType, id, status string, currentFloor int) {
	update := websockets.StatusUpdate{
		Type:         updateType,
		ID:           id,
		Status:       status,
		CurrentFloor: currentFloor,
	}

	s.hub.BroadcastUpdate(update)
	s.log.Info(ctx, "WebSocket update sent", "type", updateType, "id", id, "status", status)
}

func (s *FloorService) CallLift(ctx context.Context, floorNum int, direction domain.Direction) error {
	floor, err := s.repo.GetFloorByNumber(ctx, floorNum)
	if err != nil {
		if errors.Is(err, domain.ErrFloorNotFound) {
			return domain.ErrFloorNotFound
		}
		return fmt.Errorf("failed to get floor %d: %w", floorNum, err)
	}

	// Check floor capacity before requesting a lift
	system, err := s.repo.GetSystem(ctx)
	if err != nil {
		s.log.Error(ctx, "Failed to get system information", "error", err)
		return fmt.Errorf("failed to get system information: %w", err)
	}

	maxLiftsPerFloor := max(int(math.Ceil(float64(system.TotalLifts)*0.1)), 2)

	assignedLifts, err := s.repo.GetAssignedLiftsForFloor(ctx, floor.ID)
	s.log.Info(ctx, "Assigned lifts for floor", "floor", floorNum, "assignedLifts", assignedLifts)
	if err != nil {
		s.log.Error(ctx, "Failed to get assigned lifts for floor", "floor", floorNum, "error", err)
		return fmt.Errorf("failed to get assigned lifts for floor: %w", err)
	}

	if len(assignedLifts) >= maxLiftsPerFloor {
		s.log.Warn(ctx, "Floor has reached maximum lift capacity", "floor", floorNum, "max_capacity", maxLiftsPerFloor)
		s.eventBus.Publish(domain.FloorAtCapacityEvent{FloorNumber: floorNum})
		return fmt.Errorf("floor %d has reached maximum lift capacity", floorNum)
	}

	// If the floor hasn't reached capacity, proceed with the lift request
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
