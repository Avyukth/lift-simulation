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
	db  *sql.DB
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
		r.log.Error(ctx, "Failed to get lift", "error", err)
		return nil, fmt.Errorf("lift not found: %s", id)
	}
	if err != nil {
		r.log.Error(ctx, "Failed to get lift", "error", err)
		return nil, fmt.Errorf("failed to get lift: %w", err)
	}
	lift.SetCurrentFloor(currentFloor)
	lift.SetStatus(status)
	lift.SetCapacity(capacity)
	return &lift, nil
}

func (r *Repository) ListLifts(ctx context.Context) ([]*domain.Lift, error) {
	r.log.Info(ctx, "Listing all lifts")

	query := `SELECT id, current_floor, status, capacity FROM lifts`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "Failed to query lifts", "error", err)
		return nil, fmt.Errorf("failed to list lifts: %w", err)
	}
	defer rows.Close()

	var lifts []*domain.Lift
	for rows.Next() {
		var id string
		var currentFloor int
		var statusStr string
		var capacity int

		if err := rows.Scan(&id, &currentFloor, &statusStr, &capacity); err != nil {
			r.log.Error(ctx, "Failed to scan lift row", "error", err)
			return nil, fmt.Errorf("failed to scan lift: %w", err)
		}

		lift := domain.NewLift(id)
		lift.SetCurrentFloor(currentFloor)
		lift.SetStatus(domain.StringToLiftStatus(statusStr))
		lift.SetCapacity(capacity)

		lifts = append(lifts, lift)

		r.log.Info(ctx, "Scanned lift", "id", id, "current_floor", currentFloor, "status", statusStr, "capacity", capacity)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(ctx, "Error after scanning all rows", "error", err)
		return nil, fmt.Errorf("error after scanning lifts: %w", err)
	}

	r.log.Info(ctx, "Successfully listed all lifts", "count", lifts)
	return lifts, nil
}

func (r *Repository) SaveLift(ctx context.Context, lift *domain.Lift) error {
	// r.log.Info(ctx, "Saving lift", "lift_id", lift.ID())

	stmt, err := r.db.PrepareContext(ctx, `
		INSERT OR REPLACE INTO lifts (id, current_floor, status, capacity)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		r.log.Error(ctx, "Failed to prepare statement", "error", err)
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	statusStr := domain.LiftStatusToString(lift.Status)

	r.log.Debug(ctx, "Executing SQL",
		"lift_id", lift.ID,
		"current_floor", lift.CurrentFloor,
		"status", statusStr,
		"capacity", lift.Capacity)

	_, err = stmt.ExecContext(ctx,
		lift.ID,
		lift.CurrentFloor,
		statusStr,
		lift.Capacity)
	if err != nil {
		r.log.Error(ctx, "Failed to save lift", "error", err)
		return fmt.Errorf("failed to save lift: %w", err)
	}

	// r.log.Info(ctx, "Lift saved successfully", "lift_id", lift.ID())
	return nil
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
	// r.log.Info(ctx, "Saving floor", "floor_number", floor.Number())

	stmt, err := r.db.PrepareContext(ctx, `
		INSERT OR REPLACE INTO floors (floor_number, up_button_active, down_button_active)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		r.log.Error(ctx, "Failed to prepare statement", "error", err)
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	r.log.Debug(ctx, "Executing SQL",
		"floor_number", floor.Number(),
		"up_button", floor.GetUpButtonActive(),
		"down_button", floor.GetDownButtonActive())

	_, err = stmt.ExecContext(ctx,
		floor.Number(),
		floor.GetUpButtonActive(),
		floor.GetDownButtonActive())
	if err != nil {
		r.log.Error(ctx, "Failed to save floor", "error", err)
		return fmt.Errorf("failed to save floor: %w", err)
	}

	// r.log.Info(ctx, "Floor saved successfully", "floor_number", floor.Number())
	return nil
}

func (r *Repository) UpdateFloor(ctx context.Context, floor *domain.Floor) error {
	return r.SaveFloor(ctx, floor)
}

// System Repository Methods

func (r *Repository) GetSystem(ctx context.Context) (*domain.System, error) {
	r.log.Info(ctx, "Getting system configuration")

	query := `SELECT total_floors, total_lifts FROM system WHERE id = 1`
	var totalFloors, totalLifts int
	err := r.db.QueryRowContext(ctx, query).Scan(&totalFloors, &totalLifts)
	if err == sql.ErrNoRows {
		r.log.Error(ctx, "System configuration not found")
		return nil, fmt.Errorf("system configuration not found")
	}
	if err != nil {
		r.log.Error(ctx, "Failed to get system configuration", "error", err)
		return nil, fmt.Errorf("failed to get system configuration: %w", err)
	}

	system, err := domain.NewSystem(totalFloors, totalLifts)
	if err != nil {
		r.log.Error(ctx, "Failed to create new system from configuration", "error", err)
		return nil, fmt.Errorf("Failed to create new system from configuration: %w", err)
	}
	r.log.Info(ctx, "Successfully retrieved system configuration",
		"total_floors", system.TotalFloors(),
		"total_lifts", system.TotalLifts())
	return system, nil
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

func (r *Repository) UpdateLift(ctx context.Context, lift *domain.Lift) error {
	r.log.Info(ctx, "Updating lift", "lift_id", lift.ID)

	query := `
		UPDATE lifts
		SET current_floor = ?, status = ?, capacity = ?
		WHERE id = ?
	`

	statusStr := domain.LiftStatusToString(lift.Status)

	r.log.Debug(ctx, "Executing SQL",
		"lift_id", lift.ID,
		"current_floor", lift.CurrentFloor,
		"status", statusStr,
		"capacity", lift.Capacity)

	result, err := r.db.ExecContext(ctx, query,
		lift.CurrentFloor,
		statusStr,
		lift.Capacity,
		lift.ID)
	if err != nil {
		r.log.Error(ctx, "Failed to update lift", "error", err)
		return fmt.Errorf("failed to update lift: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error(ctx, "Failed to get rows affected", "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.log.Warn(ctx, "No lift updated", "lift_id", lift.ID)
		return fmt.Errorf("lift not found: %s", lift.ID)
	}

	r.log.Info(ctx, "Lift updated successfully", "lift_id", lift.ID)
	return nil
}

// Ensure Repository implements ports.Repository interface
var _ ports.Repository = (*Repository)(nil)
