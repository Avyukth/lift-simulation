package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Avyukth/lift-simulation/internal/application/ports"
	"github.com/Avyukth/lift-simulation/internal/domain"
	ws "github.com/Avyukth/lift-simulation/internal/infrastructure/fiber/websockets"
	"github.com/Avyukth/lift-simulation/pkg/logger"
)

// LiftService handles the business logic for lift operations
type LiftService struct {
	repo  ports.LiftRepository
	wsHub *ws.WebSocketHub
	log   *logger.Logger
}

// NewLiftService creates a new instance of LiftService
func NewLiftService(repo ports.LiftRepository, wsHub *ws.WebSocketHub, log *logger.Logger) *LiftService {
	return &LiftService{
		repo:  repo,
		wsHub: wsHub,
		log:   log,
	}
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
		s.log.Info(ctx, "Lift is already on the target floor", "lift_id", liftID, "current_floor", lift.CurrentFloor)
		return fmt.Errorf("lift is already on floor %d", targetFloor)
	}

	if err := lift.MoveTo(targetFloor); err != nil {
		s.log.Error(ctx, "Failed to move lift", "lift_id", liftID, "target_floor", targetFloor, "error", err)
		return fmt.Errorf("failed to move lift: %w", err)
	}

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
func (s *LiftService) findNearestAvailableLift(lifts []*domain.Lift, floorNum int, direction domain.Direction) *domain.Lift {
	var nearestLift *domain.Lift
	minDistance := int(^uint(0) >> 1) // Max int

	for _, lift := range lifts {
		if lift.IsAvailable() {
			distance := abs(lift.CurrentFloor - floorNum)
			if distance < minDistance || (distance == minDistance && lift.Direction == direction) {
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
