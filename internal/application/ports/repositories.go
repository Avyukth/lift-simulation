package ports

import (
	"context"

	"github.com/Avyukth/lift-simulation/internal/domain"
)

// LiftRepository defines the interface for lift persistence operations
type LiftRepository interface {
	GetLift(ctx context.Context, id string) (*domain.Lift, error)
	ListLifts(ctx context.Context) ([]*domain.Lift, error)
	SaveLift(ctx context.Context, lift *domain.Lift) error
	UpdateLift(ctx context.Context, lift *domain.Lift) error
	DeleteLift(ctx context.Context, id string) error
}

// FloorRepository defines the interface for floor persistence operations
type FloorRepository interface {
	GetFloor(ctx context.Context, floorNum int) (*domain.Floor, error)
	ListFloors(ctx context.Context) ([]*domain.Floor, error)
	SaveFloor(ctx context.Context, floor *domain.Floor) error
	UpdateFloor(ctx context.Context, floor *domain.Floor) error
}

// SystemRepository defines the interface for system-wide persistence operations
type SystemRepository interface {
	GetSystem(ctx context.Context) (*domain.System, error)
	SaveSystem(ctx context.Context, system *domain.System) error
	UpdateSystem(ctx context.Context, system *domain.System) error
	GetAllLifts(ctx context.Context) ([]*domain.Lift, error)
	GetAllFloors(ctx context.Context) ([]*domain.Floor, error)
}

// Repository combines all repository interfaces
type Repository interface {
	LiftRepository
	FloorRepository
	SystemRepository
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
