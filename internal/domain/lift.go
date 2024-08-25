package domain

import (
	"errors"
	"time"
)

// Direction represents the direction of lift movement
type Direction int

const (
	Up Direction = iota
	Down
	Idle
)

// LiftStatus represents the current status of a lift
type LiftStatus int

const (
	Available LiftStatus = iota
	Occupied
	OutOfService
)

// Lift represents a lift in the system
type Lift struct {
	ID           string     `json:"id"`
	CurrentFloor int        `json:"current_floor"`
	TargetFloor  int        `json:"target_floor"`
	Direction    Direction  `json:"direction"`
	Status       LiftStatus `json:"status"`
	Capacity     int        `json:"capacity"`
	Passengers   int        `json:"passengers"`
	LastMoveTime time.Time  `json:"last_move_time"`
}

// NewLift creates a new Lift instance
func NewLift(id string) *Lift {
	return &Lift{
		ID:           id,
		CurrentFloor: 1,
		Direction:    Idle,
		Status:       Available,
		Capacity:     10,
		LastMoveTime: time.Now(),
	}
}

func (l *Lift) SetCurrentFloor(floor int) {
	l.CurrentFloor = floor
}

func (l *Lift) IsAvailable() bool {
	return l.Status == Available
}

func (l *Lift) SetStatus(status LiftStatus) {
	l.Status = status
}

func (l *Lift) SetCapacity(capacity int) {
	l.Capacity = capacity
}

func (l *Lift) MoveTo(floor int) error {
	if floor == l.CurrentFloor {
		return errors.New("lift is already on the requested floor")
	}

	l.TargetFloor = floor
	if floor > l.CurrentFloor {
		l.Direction = Up
	} else {
		l.Direction = Down
	}
	l.Status = Occupied

	// Simulate movement time (2 seconds per floor)
	time.Sleep(time.Duration(abs(floor-l.CurrentFloor)) * 2 * time.Second)

	l.CurrentFloor = floor
	l.Direction = Idle
	l.Status = Available
	l.LastMoveTime = time.Now()

	return nil
}

func (l *Lift) AddPassengers(count int) error {
	if l.Passengers+count > l.Capacity {
		return errors.New("exceeds lift capacity")
	}
	l.Passengers += count
	return nil
}

func (l *Lift) RemovePassengers(count int) error {
	if l.Passengers-count < 0 {
		return errors.New("invalid passenger count")
	}
	l.Passengers -= count
	return nil
}

func (l *Lift) SetOutOfService() {
	l.Status = OutOfService
}

func (l *Lift) SetAvailable() {
	l.Status = Available
}

func (l *Lift) GetLastMoveTime() time.Time {
	return l.LastMoveTime
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (l *Lift) AssignToFloor(floor int) error {
	l.TargetFloor = floor
	l.Status = Occupied
	return nil
}

func (l *Lift) Reset() {
	l.CurrentFloor = 1
	l.TargetFloor = 1
	l.Direction = Idle
	l.Status = Available
	l.Passengers = 0
}

func StringToLiftStatus(status string) LiftStatus {
	switch status {
	case "Available":
		return Available
	case "Occupied":
		return Occupied
	case "OutOfService":
		return OutOfService
	default:
		return Available // Default to Available if unknown status
	}
}

func LiftStatusToString(status LiftStatus) string {
	switch status {
	case Available:
		return "Available"
	case Occupied:
		return "Occupied"
	case OutOfService:
		return "OutOfService"
	default:
		return "Unknown"
	}
}
