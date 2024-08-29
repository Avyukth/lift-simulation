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
			name TEXT UNIQUE,
			current_floor INTEGER,
			status TEXT,
			capacity INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS floors (
			id TEXT PRIMARY KEY,
			floor_number INTEGER UNIQUE,
			up_button_active BOOLEAN,
			down_button_active BOOLEAN
		)`,
		`CREATE TABLE IF NOT EXISTS system (
			id TEXT PRIMARY KEY,
			total_floors INTEGER,
			total_lifts INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS floor_lift_assignments (
            floor_id TEXT,
            lift_id TEXT,
			floor_number INTEGER,
            PRIMARY KEY (floor_id, lift_id),
            FOREIGN KEY (floor_id) REFERENCES floors(id),
            FOREIGN KEY (lift_id) REFERENCES lifts(id)
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
	r.log.Info(ctx, "Getting lift", "lift_id", id)

	query := `SELECT id, name, current_floor, status, capacity FROM lifts WHERE id = ?`
	var liftID, name string
	var currentFloor int
	var statusStr string
	var capacity int

	err := r.db.QueryRowContext(ctx, query, id).Scan(&liftID, &name, &currentFloor, &statusStr, &capacity)
	if err == sql.ErrNoRows {
		r.log.Error(ctx, "Lift not found", "lift_id", id)
		return nil, fmt.Errorf("lift not found: %s", id)
	}
	if err != nil {
		r.log.Error(ctx, "Failed to get lift", "lift_id", id, "error", err)
		return nil, fmt.Errorf("failed to get lift: %w", err)
	}

	lift := domain.NewLift(liftID, name)
	lift.SetCurrentFloor(currentFloor)
	lift.SetStatus(domain.StringToLiftStatus(statusStr))
	lift.SetCapacity(capacity)

	r.log.Info(ctx, "Successfully retrieved lift", "lift_id", id, "current_floor", currentFloor, "status", statusStr, "capacity", capacity)
	return lift, nil
}

func (r *Repository) ListLifts(ctx context.Context) ([]*domain.Lift, error) {
	r.log.Info(ctx, "Listing all lifts")

	query := `SELECT id, name, current_floor, status, capacity FROM lifts`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.log.Error(ctx, "Failed to query lifts", "error", err)
		return nil, fmt.Errorf("failed to list lifts: %w", err)
	}
	defer rows.Close()

	var lifts []*domain.Lift
	for rows.Next() {
		var id, name string
		var currentFloor int
		var statusStr string
		var capacity int

		if err := rows.Scan(&id, &name, &currentFloor, &statusStr, &capacity); err != nil {
			r.log.Error(ctx, "Failed to scan lift row", "error", err)
			return nil, fmt.Errorf("failed to scan lift: %w", err)
		}

		lift := domain.NewLift(id, name)
		lift.SetCurrentFloor(currentFloor)
		lift.SetStatus(domain.StringToLiftStatus(statusStr))
		lift.SetCapacity(capacity)

		lifts = append(lifts, lift)

		r.log.Info(ctx, "Scanned lift", "id", id, "name", name, "current_floor", currentFloor, "status", statusStr, "capacity", capacity)
	}

	if err = rows.Err(); err != nil {
		r.log.Error(ctx, "Error after scanning all rows", "error", err)
		return nil, fmt.Errorf("error after scanning lifts: %w", err)
	}

	r.log.Info(ctx, "Successfully listed all lifts", "count", len(lifts))
	return lifts, nil
}

func (r *Repository) SaveLift(ctx context.Context, lift *domain.Lift) error {
	stmt, err := r.db.PrepareContext(ctx, `
		INSERT OR REPLACE INTO lifts (id, name, current_floor, status, capacity)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		r.log.Error(ctx, "Failed to prepare statement", "error", err)
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	statusStr := domain.LiftStatusToString(lift.Status)

	r.log.Debug(ctx, "Executing SQL",
		"lift_id", lift.ID,
		"name", lift.Name,
		"current_floor", lift.CurrentFloor,
		"status", statusStr,
		"capacity", lift.Capacity)

	_, err = stmt.ExecContext(ctx,
		lift.ID,
		lift.Name,
		lift.CurrentFloor,
		statusStr,
		lift.Capacity)
	if err != nil {
		r.log.Error(ctx, "Failed to save lift", "error", err)
		return fmt.Errorf("failed to save lift: %w", err)
	}

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
func (r *Repository) GetFloor(ctx context.Context, id string) (*domain.Floor, error) {
	query := `SELECT id, floor_number, up_button_active, down_button_active FROM floors WHERE id = ?`
	var floor domain.Floor
	var floorID string
	var number int
	var upButtonActive bool
	var downButtonActive bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&floorID, &number, &upButtonActive, &downButtonActive)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("floor not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get floor: %w", err)
	}
	floor = *domain.NewFloor(floorID, number)
	floor.SetUpButtonActive(upButtonActive)
	floor.SetDownButtonActive(downButtonActive)
	return &floor, nil
}

func (r *Repository) ListFloors(ctx context.Context) ([]*domain.Floor, error) {
	query := `SELECT id, floor_number, up_button_active, down_button_active FROM floors`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list floors: %w", err)
	}
	defer rows.Close()

	var floors []*domain.Floor
	for rows.Next() {
		var floorID string
		var number int
		var upButtonActive bool
		var downButtonActive bool
		if err := rows.Scan(&floorID, &number, &upButtonActive, &downButtonActive); err != nil {
			return nil, fmt.Errorf("failed to scan floor: %w", err)
		}
		floor := domain.NewFloor(floorID, number)
		floor.SetUpButtonActive(upButtonActive)
		floor.SetDownButtonActive(downButtonActive)
		floors = append(floors, floor)
	}
	return floors, nil
}

func (r *Repository) SaveFloor(ctx context.Context, floor *domain.Floor) error {
	stmt, err := r.db.PrepareContext(ctx, `
		INSERT OR REPLACE INTO floors (id, floor_number, up_button_active, down_button_active)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		r.log.Error(ctx, "Failed to prepare statement", "error", err)
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	r.log.Debug(ctx, "Executing SQL",
		"floor_id", floor.ID,
		"floor_number", floor.Number,
		"up_button", floor.GetUpButtonActive(),
		"down_button", floor.GetDownButtonActive())

	_, err = stmt.ExecContext(ctx,
		floor.ID,
		floor.Number,
		floor.GetUpButtonActive(),
		floor.GetDownButtonActive())
	if err != nil {
		r.log.Error(ctx, "Failed to save floor", "error", err)
		return fmt.Errorf("failed to save floor: %w", err)
	}

	return nil
}

func (r *Repository) UpdateFloor(ctx context.Context, floor *domain.Floor) error {
	return r.SaveFloor(ctx, floor)
}

// System Repository Methods

func (r *Repository) GetSystem(ctx context.Context) (*domain.System, error) {
	r.log.Info(ctx, "Getting system configuration")

	query := `SELECT id, total_floors, total_lifts FROM system LIMIT 1`
	var systemID string
	var totalFloors, totalLifts int
	err := r.db.QueryRowContext(ctx, query).Scan(&systemID, &totalFloors, &totalLifts)
	if err == sql.ErrNoRows {
		r.log.Error(ctx, "System configuration not found")
		return nil, fmt.Errorf("system configuration not found")
	}
	if err != nil {
		r.log.Error(ctx, "Failed to get system configuration", "error", err)
		return nil, fmt.Errorf("failed to get system configuration: %w", err)
	}

	system, err := domain.NewSystem(systemID, totalFloors, totalLifts)
	if err != nil {
		r.log.Error(ctx, "Failed to create new system from configuration", "error", err)
		return nil, fmt.Errorf("failed to create new system from configuration: %w", err)
	}
	r.log.Info(ctx, "Successfully retrieved system configuration",
		"system_id", system.ID,
		"total_floors", system.TotalFloors,
		"total_lifts", system.TotalLifts)
	return system, nil
}

func (r *Repository) SaveSystem(ctx context.Context, system *domain.System) error {
	query := `
		INSERT OR REPLACE INTO system (id, total_floors, total_lifts)
		VALUES (?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, system.ID, system.TotalFloors, system.TotalLifts)
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
		SET name = ?, current_floor = ?, status = ?, capacity = ?
		WHERE id = ?
	`

	statusStr := domain.LiftStatusToString(lift.Status)

	r.log.Debug(ctx, "Executing SQL",
		"lift_id", lift.ID,
		"name", lift.Name,
		"current_floor", lift.CurrentFloor,
		"status", statusStr,
		"capacity", lift.Capacity)

	result, err := r.db.ExecContext(ctx, query,
		lift.Name,
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
func (r *Repository) GetFloorByNumber(ctx context.Context, floorNum int) (*domain.Floor, error) {
	query := `SELECT id FROM floors WHERE floor_number = ?`
	var floorID string
	err := r.db.QueryRowContext(ctx, query, floorNum).Scan(&floorID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("floor not found: %d", floorNum)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get floor: %w", err)
	}
	return domain.NewFloor(floorID, floorNum), nil
}

func (r *Repository) AssignLiftToFloor(ctx context.Context, liftID string, floorID string, floorNumber int) error {
	query := `INSERT OR REPLACE INTO floor_lift_assignments (floor_id, lift_id, floor_number) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, floorID, liftID, floorNumber)
	if err != nil {
		return fmt.Errorf("failed to assign lift to floor: %w", err)
	}
	return nil
}

func (r *Repository) UnassignLiftFromFloor(ctx context.Context, liftID string, floorID string) error {
	query := `DELETE FROM floor_lift_assignments WHERE floor_id = ? AND lift_id = ?`
	_, err := r.db.ExecContext(ctx, query, floorID, liftID)
	if err != nil {
		return fmt.Errorf("failed to unassign lift from floor: %w", err)
	}
	return nil
}

func (r *Repository) GetAssignedLiftsForFloor(ctx context.Context, floorID string) ([]*domain.Lift, error) {
	query := `
        SELECT l.id, l.name, l.current_floor, l.status, l.capacity
        FROM lifts l
        JOIN floor_lift_assignments fla ON l.id = fla.lift_id
        WHERE fla.floor_id = ?
    `
	rows, err := r.db.QueryContext(ctx, query, floorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get assigned lifts: %w", err)
	}
	defer rows.Close()

	var lifts []*domain.Lift
	for rows.Next() {
		var lift domain.Lift
		var statusStr string
		err := rows.Scan(&lift.ID, &lift.Name, &lift.CurrentFloor, &statusStr, &lift.Capacity)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lift row: %w", err)
		}
		lift.Status = domain.StringToLiftStatus(statusStr)
		lifts = append(lifts, &lift)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after scanning all rows: %w", err)
	}

	return lifts, nil
}

func (r *Repository) UnassignBulk(ctx context.Context) error {
	query := `DELETE FROM floor_lift_assignments`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete all records from floor_lift_assignments: %w", err)
	}
	return nil

}

// Ensure Repository implements ports.Repository interface
var _ ports.Repository = (*Repository)(nil)
