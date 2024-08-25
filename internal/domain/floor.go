package domain

import (
	"errors"
)

var (
	ErrFloorNotFound = errors.New("floor not found")
	ErrLiftNotFound  = errors.New("lift not found")
)

// Floor represents a floor in the lift system
type Floor struct {
	ID               string
	Number           int
	UpButtonActive   bool
	DownButtonActive bool
}

// NewFloor creates a new Floor instance
func NewFloor(id string, number int) *Floor {
	return &Floor{
		ID:     id,
		Number: number,
	}
}

// GetNumber returns the floor number
func (f *Floor) GetNumber() int {
	return f.Number
}

func (f *Floor) SetUpButtonActive(active bool) {
	f.UpButtonActive = active
}

func (f *Floor) SetDownButtonActive(active bool) {
	f.DownButtonActive = active
}

func (f *Floor) GetUpButtonActive() bool {
	return f.UpButtonActive
}

func (f *Floor) GetDownButtonActive() bool {
	return f.DownButtonActive
}

// RequestLift activates the appropriate call button
func (f *Floor) RequestLift(direction Direction) error {
	switch direction {
	case Up:
		f.UpButtonActive = true
	case Down:
		f.DownButtonActive = true
	default:
		return errors.New("invalid direction")
	}
	return nil
}

// CancelRequest deactivates the specified call button
func (f *Floor) CancelRequest(direction Direction) error {
	switch direction {
	case Up:
		f.UpButtonActive = false
	case Down:
		f.DownButtonActive = false
	default:
		return errors.New("invalid direction")
	}
	return nil
}

// HasActiveCall checks if there's an active call on this floor
func (f *Floor) HasActiveCall() bool {
	return f.UpButtonActive || f.DownButtonActive
}

// IsUpButtonActive checks if the up button is active
func (f *Floor) IsUpButtonActive() bool {
	return f.UpButtonActive
}

// IsDownButtonActive checks if the down button is active
func (f *Floor) IsDownButtonActive() bool {
	return f.DownButtonActive
}

// ResetButtons resets both call buttons
func (f *Floor) ResetButtons() {
	f.UpButtonActive = false
	f.DownButtonActive = false
}

func NewFloorButtonsResetEvent(floorID string, floorNum int) Event {
	return Event{
		Type: FloorButtonsReset,
		Payload: struct {
			FloorID     string
			FloorNumber int
		}{
			FloorID:     floorID,
			FloorNumber: floorNum,
		},
	}
}

func NewLiftAssignedEvent(liftID string, floorID string, floorNum int) Event {
	return Event{
		Type: LiftAssigned,
		Payload: struct {
			LiftID      string
			FloorID     string
			FloorNumber int
		}{
			LiftID:      liftID,
			FloorID:     floorID,
			FloorNumber: floorNum,
		},
	}
}

// SetNumber sets the floor number
func (f *Floor) SetNumber(number int) {
	f.Number = number
}
