package ports

import (
	"context"

	"github.com/Avyukth/lift-simulation/internal/domain"
)

// LiftRepository defines the interface for lift persistence operations
type LiftRepository interface {
	GetLift(ctx context.Context, id string) (*domain.Lift, error)
	ListLifts(ctx context.Context) ([]*domain.Lift, error)
	UpdateLift(ctx context.Context, lift *domain.Lift) error
	DeleteLift(ctx context.Context, id string) error
}

// FloorRepository defines the interface for floor persistence operations
type FloorRepository interface {
	GetFloor(ctx context.Context, id string) (*domain.Floor, error)
	GetFloorByNumber(ctx context.Context, floorNum int) (*domain.Floor, error)
	ListFloors(ctx context.Context) ([]*domain.Floor, error)
	UpdateFloor(ctx context.Context, floor *domain.Floor) error
}

// SystemRepository defines the interface for system-wide persistence operations
type SystemRepository interface {
	GetSystem(ctx context.Context) (*domain.System, error)
	SaveSystem(ctx context.Context, system *domain.System) error
	UpdateSystem(ctx context.Context, system *domain.System) error
	SaveLift(ctx context.Context, lift *domain.Lift) error
	SaveFloor(ctx context.Context, floor *domain.Floor) error
}

type LiftFloorManager interface {
	GetLift(ctx context.Context, id string) (*domain.Lift, error)
	GetFloor(ctx context.Context, id string) (*domain.Floor, error)
	GetAllLifts(ctx context.Context) ([]*domain.Lift, error)
	GetAllFloors(ctx context.Context) ([]*domain.Floor, error)
	GetFloorByNumber(ctx context.Context, floorNum int) (*domain.Floor, error)
	UnassignLiftFromFloor(ctx context.Context, liftID string, floorID string) error
	AssignLiftToFloor(ctx context.Context, liftID string, floorID string, floorNumber int) error
	GetAssignedLiftsForFloor(ctx context.Context, floorID string) ([]*domain.Lift, error)
}

// Repository combines all repository interfaces
type Repository interface {
	LiftRepository
	FloorRepository
	SystemRepository
	LiftFloorManager
}

type LiftOperations interface {
	LiftRepository
	LiftFloorManager
	GetSystem(ctx context.Context) (*domain.System, error)
}

type FloorOperations interface {
	FloorRepository
	LiftFloorManager
	GetSystem(ctx context.Context) (*domain.System, error)
}

// Transaction defines the interface for database transactions
type Transaction interface {
	Commit() error
	Rollback() error
}

// TransactionalRepository extends Repository with transaction support
type TransactionalRepository interface {
	Repository
	BeginTx(ctx context.Context) (Transaction, error)
	WithTx(tx Transaction) Repository
}
