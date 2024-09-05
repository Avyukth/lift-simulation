package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/Avyukth/lift-simulation/internal/application/events"
	"github.com/Avyukth/lift-simulation/internal/application/ports"
	"github.com/Avyukth/lift-simulation/internal/domain"
	ws "github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/websockets"
	"github.com/Avyukth/lift-simulation/pkg/logger"
)

// LiftService handles the business logic for lift operations
type LiftService struct {
	repo     ports.LiftOperations
	eventBus events.EventBus
	wsHub    *ws.WebSocketHub
	mu       sync.RWMutex
	log      *logger.Logger
}

type LiftRequestedHandler struct {
	service *LiftService
}

func (h *LiftRequestedHandler) Handle(event domain.Event) {
	if liftRequestedEvent, ok := event.(domain.LiftRequestedEvent); ok {
		h.service.processLiftRequest(context.Background(), liftRequestedEvent.FloorNumber, liftRequestedEvent.Direction)
	}
}

// NewLiftService creates a new instance of LiftService
func NewLiftService(repo ports.LiftOperations, eventBus events.EventBus, wsHub *ws.WebSocketHub, log *logger.Logger) *LiftService {
	service := &LiftService{
		repo:     repo,
		eventBus: eventBus,
		wsHub:    wsHub,
		mu:       sync.RWMutex{},
		log:      log,
	}

	// Subscribe to LiftRequested events
	eventBus.Subscribe(domain.LiftRequested, &LiftRequestedHandler{service: service})

	return service
}

// MoveLift moves a lift to a target floor
func (s *LiftService) MoveLift(ctx context.Context, liftID string, targetFloor int) error {
	s.log.Info(ctx, "Moving lift", "lift_id", liftID, "target_floor", targetFloor)

	lift, err := s.repo.GetLift(ctx, liftID)
	if err != nil {
		s.log.Error(ctx, "Failed to retrieve lift", "lift_id", liftID, "error", err)
		return fmt.Errorf("failed to retrieve lift: %w", err)
	}

	// Check if the lift is already on the target floor
	if lift.CurrentFloor == targetFloor {
		s.log.Error(ctx, "Lift is already on the target floor", "lift_id", liftID, "current_floor", lift.CurrentFloor)
		return fmt.Errorf("lift is already on floor %d", targetFloor)
	}

	// Unassign the lift from its current floor
	currentFloor, err := s.repo.GetFloorByNumber(ctx, lift.CurrentFloor)
	if err != nil {
		s.log.Error(ctx, "Failed to get current floor", "floor_number", lift.CurrentFloor, "error", err)
		return fmt.Errorf("failed to get current floor: %w", err)
	}
	err = s.UnassignLiftFromFloor(ctx, liftID, currentFloor.ID)
	if err != nil {
		s.log.Error(ctx, "Failed to unassign lift from current floor", "lift_id", liftID, "floor_id", currentFloor.ID, "error", err)
		return fmt.Errorf("failed to unassign lift from current floor: %w", err)
	}

	if err := lift.MoveTo(targetFloor); err != nil {
		s.log.Error(ctx, "Failed to move lift", "lift_id", liftID, "target_floor", targetFloor, "error", err)
		return fmt.Errorf("failed to move lift: %w", err)
	}

	// Assign the lift to the target floor
	floor, err := s.repo.GetFloorByNumber(ctx, targetFloor)
	if err != nil {
		s.log.Error(ctx, "Failed to get target floor", "floor_number", targetFloor, "error", err)
		return fmt.Errorf("failed to get target floor: %w", err)
	}
	err = s.AssignLiftToFloor(ctx, liftID, floor.ID, floor.Number)
	if err != nil {
		s.log.Error(ctx, "Failed to assign lift to target floor", "lift_id", liftID, "floor_id", floor.ID, "error", err)
		return fmt.Errorf("failed to assign lift to target floor: %w", err)
	}
	s.log.Info(ctx, "Failed to Lift--service-----------------------")

	if err := s.repo.UpdateLift(ctx, lift); err != nil {
		s.log.Error(ctx, "Failed to update lift after move", "lift_id", liftID, "error", err)
		return fmt.Errorf("failed to update lift after move: %w", err)
	}

	s.log.Info(ctx, "Successfully moved lift", "lift_id", liftID, "target_floor", targetFloor)
	return nil
}

// GetLiftStatus retrieves the current status of a lift
func (s *LiftService) GetLiftStatus(ctx context.Context, liftID string) (*domain.Lift, error) {
	return s.repo.GetLift(ctx, liftID)
}

// ListLifts retrieves all lifts in the system
func (s *LiftService) ListLifts(ctx context.Context) ([]*domain.Lift, error) {
	return s.repo.ListLifts(ctx)
}

// findNearestAvailableLift is a helper function to find the nearest available lift
// func (s *LiftService) findNearestAvailableLift(lifts []*domain.Lift, floorNum int, direction domain.Direction) *domain.Lift {
// 	var nearestLift *domain.Lift
// 	minDistance := int(^uint(0) >> 1) // Max int

// 	for _, lift := range lifts {
// 		if lift.IsAvailable() {
// 			distance := abs(lift.CurrentFloor - floorNum)
// 			if distance < minDistance || (distance == minDistance && lift.Direction == direction) {
// 				minDistance = distance
// 				nearestLift = lift
// 			}
// 		}
// 	}

// 	return nearestLift
// }

// abs returns the absolute value of an int
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// SetLiftStatus sets the status of a lift
func (s *LiftService) SetLiftStatus(ctx context.Context, liftID string, status domain.LiftStatus) error {
	lift, err := s.repo.GetLift(ctx, liftID)
	if err != nil {
		return err
	}

	lift.SetStatus(status)
	if err := s.repo.UpdateLift(ctx, lift); err != nil {
		return err
	}

	return nil
}

// AssignLiftToFloor assigns a lift to a floor
func (s *LiftService) AssignLiftToFloor(ctx context.Context, liftID, floorID string, floorNum int) error {
	s.log.Info(ctx, "Attempting to assign lift to floor", "lift_id", liftID, "floor_id", floorID, "floor_num", floorNum)

	// Use a mutex to prevent race conditions
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if there are already two lifts assigned to this floor
	assignedLifts, err := s.repo.GetAssignedLiftsForFloor(ctx, floorID)
	if err != nil {
		s.log.Error(ctx, "Failed to get assigned lifts", "error", err, "floor_id", floorID)
		return fmt.Errorf("failed to get assigned lifts: %w", err)
	}

	if len(assignedLifts) >= 2 {
		s.log.Warn(ctx, "Floor already has maximum lifts assigned", "floor_id", floorID)
		return fmt.Errorf("lift capacity exceeded")
	}

	// Assign the lift to the floor
	err = s.repo.AssignLiftToFloor(ctx, liftID, floorID, floorNum)
	if err != nil {
		s.log.Error(ctx, "Failed to assign lift to floor", "error", err, "lift_id", liftID, "floor_id", floorID)
		return fmt.Errorf("failed to assign lift to floor: %w", err)
	}

	s.log.Info(ctx, "Lift successfully assigned to floor", "lift_id", liftID, "floor_id", floorID)
	// time.Sleep(60*time.Second)
	return nil
}

// UnassignLiftFromFloor removes a lift assignment from a floor
func (s *LiftService) UnassignLiftFromFloor(ctx context.Context, liftID string, floorID string) error {
	err := s.repo.UnassignLiftFromFloor(ctx, liftID, floorID)
	if err != nil {
		s.log.Error(ctx, "failed to unassign lift from floor: %w", err, "lift_id", liftID, "floor_id", floorID)
		return fmt.Errorf("failed to unassign lift from floor: %w", err)
	}
	s.log.Info(ctx, "Lift unassigned from floor=======LiftService", "lift_id", liftID, "floor_id", floorID)

	s.log.Info(ctx, "Lift unassigned from floor", "lift_id", liftID, "floor_id", floorID)
	return nil
}

// GetAssignedLiftsForFloor retrieves the lifts assigned to a specific floor
func (s *LiftService) GetAssignedLiftsForFloor(ctx context.Context, floorID string) ([]*domain.Lift, error) {
	return s.repo.GetAssignedLiftsForFloor(ctx, floorID)
}

func (s *LiftService) findAvailableLift(ctx context.Context) (*domain.Lift, error) {
	lifts, err := s.repo.GetAllLifts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get lifts: %w", err)
	}

	var closestLift *domain.Lift
	minDistance := int(^uint(0) >> 1)

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

func (s *LiftService) processLiftRequest(ctx context.Context, floorNum int, direction domain.Direction) {
	system, err := s.repo.GetSystem(ctx)
	if err != nil {
		s.log.Error(ctx, "Failed to get system information", "error", err)
		return
	}

	maxLiftsPerFloor := max(int(math.Ceil(float64(system.TotalLifts)*0.1)), 2)
	floor, err := s.repo.GetFloorByNumber(ctx, floorNum)

	if err != nil {
		s.log.Error(ctx, "Failed to get floor information: %w", err, "floor_num", floorNum)
		return
	}

	assignedLifts, err := s.repo.GetAssignedLiftsForFloor(ctx, floor.ID)
	if err != nil {
		s.log.Error(ctx, "Failed to get assigned lifts for floor", "floor", floorNum, "error", err)
		return
	}

	if len(assignedLifts) >= maxLiftsPerFloor {
		s.log.Warn(ctx, "Floor has reached maximum lift capacity", "floor", floorNum, "max_capacity", maxLiftsPerFloor)
		// Publish an event or notify the requester that the floor is at capacity
		s.eventBus.Publish(domain.FloorAtCapacityEvent{FloorNumber: floorNum})
		return
	}

	lift, err := s.findAvailableLift(ctx)
	if err != nil {
		s.log.Error(ctx, "Failed to find available lift", "error", err)
		return
	}

	s.log.Info(ctx, "Lift is Moving", "lift_id", lift.ID, "target_floor", floorNum, "direction", direction)

	err = s.MoveLift(ctx, lift.ID, floorNum)
	if err != nil {
		s.log.Error(ctx, "Failed to move lift", "lift_id", lift.ID, "target_floor", floorNum, "error", err)
		return
	}

	s.log.Info(ctx, "Lift arrived at requested floor", "lift_id", lift.ID, "floor", floorNum)

	// Publish a LiftArrived event
	s.eventBus.Publish(domain.LiftArrivedEvent{
		LiftID:      lift.ID,
		FloorNumber: floorNum,
	})
}

func (s *LiftService) ResetLift(ctx context.Context, liftID string) error {
	lift, err := s.repo.GetLift(ctx, liftID)
	if err != nil {
		s.log.Error(ctx, "Failed to reset lift", "lift", liftID, "error", err)
		if errors.Is(err, domain.ErrLiftNotFound) {
			return domain.ErrLiftNotFound
		}
		return fmt.Errorf("failed to get lift: %w", err)
	}

	// Reset lift properties
	lift.CurrentFloor = 0
	lift.Status = domain.LiftStatus(domain.Idle)

	err = s.repo.UpdateLift(ctx, lift)
	if err != nil {
		return fmt.Errorf("failed to update lift: %w", err)
	}

	return nil
}

func (s *LiftService) ResetLifts(ctx context.Context) error {
	// Get all lifts
	lifts, err := s.repo.GetAllLifts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all lifts: %w", err)
	}

	// Reset each lift
	for _, lift := range lifts {
		// Create a new lift with the same ID and name, but reset all other properties
		resetLift := domain.NewLift(lift.ID, lift.Name)

		// Update the lift in the repository
		err = s.repo.UpdateLift(ctx, resetLift)
		if err != nil {
			return fmt.Errorf("failed to update lift %s: %w", lift.ID, err)
		}
	}

	return nil
}
