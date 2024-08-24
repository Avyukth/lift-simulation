package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/Avyukth/lift-simulation/internal/application/ports"
	"github.com/Avyukth/lift-simulation/internal/domain"
)

// Repository implements the Repository interface using in-memory storage
type Repository struct {
	lifts  map[string]*domain.Lift
	floors map[int]*domain.Floor
	system *domain.System
	mu     sync.RWMutex
}

// NewRepository creates a new instance of the in-memory repository
func NewRepository() *Repository {
	return &Repository{
		lifts:  make(map[string]*domain.Lift),
		floors: make(map[int]*domain.Floor),
	}
}

// Lift Repository Methods

func (r *Repository) GetLift(ctx context.Context, id string) (*domain.Lift, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lift, ok := r.lifts[id]
	if !ok {
		return nil, fmt.Errorf("lift not found: %s", id)
	}
	return lift, nil
}

func (r *Repository) ListLifts(ctx context.Context) ([]*domain.Lift, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lifts := make([]*domain.Lift, 0, len(r.lifts))
	for _, lift := range r.lifts {
		lifts = append(lifts, lift)
	}
	return lifts, nil
}

func (r *Repository) SaveLift(ctx context.Context, lift *domain.Lift) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lifts[lift.ID()] = lift
	return nil
}

func (r *Repository) UpdateLift(ctx context.Context, lift *domain.Lift) error {
	return r.SaveLift(ctx, lift)
}

func (r *Repository) DeleteLift(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.lifts, id)
	return nil
}

// Floor Repository Methods

func (r *Repository) GetFloor(ctx context.Context, floorNum int) (*domain.Floor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	floor, ok := r.floors[floorNum]
	if !ok {
		return nil, fmt.Errorf("floor not found: %d", floorNum)
	}
	return floor, nil
}

func (r *Repository) ListFloors(ctx context.Context) ([]*domain.Floor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	floors := make([]*domain.Floor, 0, len(r.floors))
	for _, floor := range r.floors {
		floors = append(floors, floor)
	}
	return floors, nil
}

func (r *Repository) SaveFloor(ctx context.Context, floor *domain.Floor) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.floors[floor.Number()] = floor
	return nil
}

func (r *Repository) UpdateFloor(ctx context.Context, floor *domain.Floor) error {
	return r.SaveFloor(ctx, floor)
}

// System Repository Methods

func (r *Repository) GetSystem(ctx context.Context) (*domain.System, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.system == nil {
		return nil, fmt.Errorf("system not initialized")
	}
	return r.system, nil
}

func (r *Repository) SaveSystem(ctx context.Context, system *domain.System) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.system = system
	return nil
}

func (r *Repository) UpdateSystem(ctx context.Context, system *domain.System) error {
	return r.SaveSystem(ctx, system)
}

func (r *Repository) GetAllLifts(ctx context.Context) ([]*domain.Lift, error) {
	return r.ListLifts(ctx)
}

func (r *Repository) GetAllFloors(ctx context.Context) ([]*domain.Floor, error) {
	return r.ListFloors(ctx)
}

// Ensure Repository implements ports.Repository interface
var _ ports.Repository = (*Repository)(nil)
