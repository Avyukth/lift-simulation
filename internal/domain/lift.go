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
	id           string
	currentFloor int
	targetFloor  int
	direction    Direction
	status       LiftStatus
	capacity     int
	passengers   int
	lastMoveTime time.Time
}

// NewLift creates a new Lift instance
func NewLift(id string) *Lift {
	return &Lift{
		id:           id,
		currentFloor: 1,
		direction:    Idle,
		status:       Available,
		capacity:     10,
		lastMoveTime: time.Now(),
	}
}

// ID returns the lift's identifier
func (l *Lift) ID() string {
	return l.id
}

// CurrentFloor returns the current floor of the lift
func (l *Lift) CurrentFloor() int {
	return l.currentFloor
}

// Status returns the current status of the lift
func (l *Lift) Status() LiftStatus {
	return l.status
}

// IsAvailable checks if the lift is available
func (l *Lift) IsAvailable() bool {
	return l.status == Available
}

// MoveTo moves the lift to the specified floor
func (l *Lift) MoveTo(floor int) error {
	if floor == l.currentFloor {
		return errors.New("lift is already on the requested floor")
	}

	l.targetFloor = floor
	if floor > l.currentFloor {
		l.direction = Up
	} else {
		l.direction = Down
	}
	l.status = Occupied

	// Simulate movement time (2 seconds per floor)
	time.Sleep(time.Duration(abs(floor-l.currentFloor)) * 2 * time.Second)

	l.currentFloor = floor
	l.direction = Idle
	l.status = Available
	l.lastMoveTime = time.Now()

	return nil
}

// AddPassengers adds passengers to the lift
func (l *Lift) AddPassengers(count int) error {
	if l.passengers+count > l.capacity {
		return errors.New("exceeds lift capacity")
	}
	l.passengers += count
	return nil
}

// RemovePassengers removes passengers from the lift
func (l *Lift) RemovePassengers(count int) error {
	if l.passengers-count < 0 {
		return errors.New("invalid passenger count")
	}
	l.passengers -= count
	return nil
}

// SetOutOfService sets the lift's status to out of service
func (l *Lift) SetOutOfService() {
	l.status = OutOfService
}

// SetAvailable sets the lift's status to available
func (l *Lift) SetAvailable() {
	l.status = Available
}

// LastMoveTime returns the time of the lift's last movement
func (l *Lift) LastMoveTime() time.Time {
	return l.lastMoveTime
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (l *Lift) AssignToFloor(floor int) error {
    l.targetFloor = floor
    l.status = Occupied
    return nil
}

func (l *Lift) Reset() {
    l.currentFloor = 1
    l.targetFloor = 1
    l.direction = Idle
    l.status = Available
    l.passengers = 0
}

func (l *Lift) Direction() Direction {
    return l.direction
}

// Add these setter methods
func (l *Lift) SetCurrentFloor(floor int) {
    l.currentFloor = floor
}

func (l *Lift) SetStatus(status LiftStatus) {
    l.status = status
}

func (l *Lift) SetCapacity(capacity int) {
    l.capacity = capacity
}

func (l *Lift) Capacity() int {
    return l.capacity
}
