package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Avyukth/lift-simulation/internal/application/ports"
	"github.com/Avyukth/lift-simulation/internal/domain"
	"github.com/Avyukth/lift-simulation/pkg/logger"
	_ "github.com/mattn/go-sqlite3"
)

// Repository implements the Repository interface using SQLite
type Repository struct {
	db *sql.DB
	log *logger.Logger
}

// NewRepository creates a new instance of the SQLite repository
func NewRepository(dbPath string, log *logger.Logger) (*Repository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return &Repository{db: db, log: log}, nil
}

// createTables creates the necessary tables if they don't exist
func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS lifts (
			id TEXT PRIMARY KEY,
			current_floor INTEGER,
			status TEXT,
			capacity INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS floors (
			floor_number INTEGER PRIMARY KEY,
			up_button_active BOOLEAN,
			down_button_active BOOLEAN
		)`,
		`CREATE TABLE IF NOT EXISTS system (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			total_floors INTEGER,
			total_lifts INTEGER
		)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return err
		}
	}

	return nil
}

// Lift Repository Methods

func (r *Repository) GetLift(ctx context.Context, id string) (*domain.Lift, error) {
	query := `SELECT id, current_floor, status, capacity FROM lifts WHERE id = ?`
	var lift domain.Lift
	var currentFloor int
	var status domain.LiftStatus
	var capacity int
	err := r.db.QueryRowContext(ctx, query, id).Scan(&currentFloor, &status, &capacity)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lift not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lift: %w", err)
	}
	lift.SetCurrentFloor(currentFloor)
	lift.SetStatus(status)
	lift.SetCapacity(capacity)
	return &lift, nil
}

func (r *Repository) ListLifts(ctx context.Context) ([]*domain.Lift, error) {
	query := `SELECT id, current_floor, status, capacity FROM lifts`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list lifts: %w", err)
	}
	defer rows.Close()

	var lifts []*domain.Lift
	for rows.Next() {
		var lift domain.Lift
		var currentFloor int
		var status domain.LiftStatus
		var capacity int
		if err := rows.Scan(&currentFloor, &status, &capacity); err != nil {
			return nil, fmt.Errorf("failed to scan lift: %w", err)
		}
		lift.SetCurrentFloor(currentFloor)
		lift.SetStatus(status)
		lift.SetCapacity(capacity)
		lifts = append(lifts, &lift)
	}
	return lifts, nil
}

func (r *Repository) SaveLift(ctx context.Context, lift *domain.Lift) error {
	stmt, err := r.db.PrepareContext(ctx, `
		INSERT OR REPLACE INTO lifts (id, current_floor, status, capacity)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var id string
	var currentFloor int
	var status domain.LiftStatus
	var capacity int

	err = stmt.QueryRow(
		lift.ID(),
		lift.CurrentFloor(),
		lift.Status(),
		lift.Capacity(),
	).Scan(&id, &currentFloor, &status, &capacity)

	if err != nil {
		return fmt.Errorf("failed to save lift: %w", err)
	}

	lift.SetCurrentFloor(currentFloor)
	lift.SetStatus(status)
	lift.SetCapacity(capacity)

	return nil
}

func (r *Repository) UpdateLift(ctx context.Context, lift *domain.Lift) error {
	return r.SaveLift(ctx, lift)
}

func (r *Repository) DeleteLift(ctx context.Context, id string) error {
	query := `DELETE FROM lifts WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete lift: %w", err)
	}
	return nil
}

// Floor Repository Methods

func (r *Repository) GetFloor(ctx context.Context, floorNum int) (*domain.Floor, error) {
	query := `SELECT floor_number, up_button_active, down_button_active FROM floors WHERE floor_number = ?`
	var floor domain.Floor
	var number int
	var upButtonActive bool
	var downButtonActive bool
	err := r.db.QueryRowContext(ctx, query, floorNum).Scan(&number, &upButtonActive, &downButtonActive)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("floor not found: %d", floorNum)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get floor: %w", err)
	}
	floor.SetNumber(number)
	floor.SetUpButtonActive(upButtonActive)
	floor.SetDownButtonActive(downButtonActive)
	return &floor, nil
}

func (r *Repository) ListFloors(ctx context.Context) ([]*domain.Floor, error) {
	query := `SELECT floor_number, up_button_active, down_button_active FROM floors`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list floors: %w", err)
	}
	defer rows.Close()

	var floors []*domain.Floor
	for rows.Next() {
		var floor domain.Floor
		var number int
		var upButtonActive bool
		var downButtonActive bool
		if err := rows.Scan(&number, &upButtonActive, &downButtonActive); err != nil {
			return nil, fmt.Errorf("failed to scan floor: %w", err)
		}
		floor.SetNumber(number)
		floor.SetUpButtonActive(upButtonActive)
		floor.SetDownButtonActive(downButtonActive)
		floors = append(floors, &floor)
	}
	return floors, nil
}

func (r *Repository) SaveFloor(ctx context.Context, floor *domain.Floor) error {
	stmt, err := r.db.PrepareContext(ctx, `
		INSERT OR REPLACE INTO floors (floor_number, up_button_active, down_button_active)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var number int
	err = stmt.QueryRow(floor.Number()).Scan(&number)
	if err != nil {
		return fmt.Errorf("failed to save floor: %w", err)
	}
	floor.SetNumber(number)
	return nil
}

func (r *Repository) UpdateFloor(ctx context.Context, floor *domain.Floor) error {
	return r.SaveFloor(ctx, floor)
}

// System Repository Methods

func (r *Repository) GetSystem(ctx context.Context) (*domain.System, error) {
	query := `SELECT total_floors, total_lifts FROM system WHERE id = 1`
	var system domain.System
	err := r.db.QueryRowContext(ctx, query).Scan(system.GetTotalFloors(), system.TotalLifts())
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("system configuration not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get system configuration: %w", err)
	}
	return &system, nil
}

func (r *Repository) SaveSystem(ctx context.Context, system *domain.System) error {
	query := `
		INSERT OR REPLACE INTO system (id, total_floors, total_lifts)
		VALUES (1, ?, ?)
	`
	totalFloors := system.TotalFloors()
	totalLifts := system.TotalLifts()
	_, err := r.db.ExecContext(ctx, query, totalFloors, totalLifts)
	if err != nil {
		return fmt.Errorf("failed to save system configuration: %w", err)
	}
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

// Close closes the database connection
func (r *Repository) Close() error {
	return r.db.Close()
}

// Ensure Repository implements ports.Repository interface
var _ ports.Repository = (*Repository)(nil)
