package services

import (
	"context"
	"errors"

	"github.com/Avyukth/lift-simulation/internal/application/ports"
	"github.com/Avyukth/lift-simulation/internal/domain"
	ws "github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/websockets"
	"github.com/Avyukth/lift-simulation/pkg/logger"
)

// LiftService handles the business logic for lift operations
type LiftService struct {
	repo     ports.LiftRepository
	eventBus ports.EventBus
	wsHub    *ws.WebSocketHub
	log  *logger.Logger
}

// NewLiftService creates a new instance of LiftService
func NewLiftService(repo ports.LiftRepository, eventBus ports.EventBus, wsHub *ws.WebSocketHub, log *logger.Logger) *LiftService {
	return &LiftService{
		repo:     repo,
		eventBus: eventBus,
		wsHub:    wsHub,
		log:      log,
	}
}

// MoveLift moves a lift to a target floor
func (s *LiftService) MoveLift(ctx context.Context, liftID string, targetFloor int) error {
	lift, err := s.repo.GetLift(ctx, liftID)
	if err != nil {
		return err
	}

	if err := lift.MoveTo(targetFloor); err != nil {
		return err
	}

	if err := s.repo.UpdateLift(ctx, lift); err != nil {
		return err
	}

	// Publish an event about the lift movement
	event := domain.NewLiftMovedEvent(lift.ID(), lift.CurrentFloor())
	s.wsHub.BroadcastUpdate(event)
	return s.eventBus.Publish(ctx, event)
}

// GetLiftStatus retrieves the current status of a lift
func (s *LiftService) GetLiftStatus(ctx context.Context, liftID string) (*domain.Lift, error) {
	return s.repo.GetLift(ctx, liftID)
}

// ListLifts retrieves all lifts in the system
func (s *LiftService) ListLifts(ctx context.Context) ([]*domain.Lift, error) {
	return s.repo.ListLifts(ctx)
}

// AssignLiftToFloor assigns the most appropriate lift to service a floor request
func (s *LiftService) AssignLiftToFloor(ctx context.Context, floorNum int, direction domain.Direction) (*domain.Lift, error) {
	lifts, err := s.repo.ListLifts(ctx)
	if err != nil {
		return nil, err
	}

	assignedLift := s.findNearestAvailableLift(lifts, floorNum, direction)
	if assignedLift == nil {
		return nil, errors.New("no available lifts")
	}

	if err := assignedLift.AssignToFloor(floorNum); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateLift(ctx, assignedLift); err != nil {
		return nil, err
	}

	// Publish an event about the lift assignment
	event := domain.NewLiftAssignedEvent(assignedLift.ID(), floorNum)
	if err := s.eventBus.Publish(ctx, event); err != nil {
		return nil, err
	}

	return assignedLift, nil
}

// findNearestAvailableLift is a helper function to find the nearest available lift
func (s *LiftService) findNearestAvailableLift(lifts []*domain.Lift, floorNum int, direction domain.Direction) *domain.Lift {
	var nearestLift *domain.Lift
	minDistance := int(^uint(0) >> 1) // Max int

	for _, lift := range lifts {
		if lift.IsAvailable() {
			distance := abs(lift.CurrentFloor() - floorNum)
			if distance < minDistance || (distance == minDistance && lift.Direction() == direction) {
				minDistance = distance
				nearestLift = lift
			}
		}
	}

	return nearestLift
}

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
