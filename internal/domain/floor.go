package domain

import "errors"

// Floor represents a floor in the lift system
type Floor struct {
	number         int
	upButtonActive bool
	downButtonActive bool
}

// NewFloor creates a new Floor instance
func NewFloor(number int) *Floor {
	return &Floor{
		number: number,
	}
}

// Number returns the floor number
func (f *Floor) Number() int {
	return f.number
}
func (f *Floor) SetUpButtonActive(active bool){
	f.upButtonActive = active
}
func (f *Floor) SetDownButtonActive(active bool){
	f.downButtonActive = active
}
func (f *Floor) GetUpButtonActive()bool{
	return f.upButtonActive
}
func (f *Floor) GetDownButtonActive()bool{
	return f.downButtonActive
}



// RequestLift activates the appropriate call button
func (f *Floor) RequestLift(direction Direction) error {
	switch direction {
	case Up:
		f.upButtonActive = true
	case Down:
		f.downButtonActive = true
	default:
		return errors.New("invalid direction")
	}
	return nil
}

// CancelRequest deactivates the specified call button
func (f *Floor) CancelRequest(direction Direction) error {
	switch direction {
	case Up:
		f.upButtonActive = false
	case Down:
		f.downButtonActive = false
	default:
		return errors.New("invalid direction")
	}
	return nil
}

// HasActiveCall checks if there's an active call on this floor
func (f *Floor) HasActiveCall() bool {
	return f.upButtonActive || f.downButtonActive
}

// IsUpButtonActive checks if the up button is active
func (f *Floor) IsUpButtonActive() bool {
	return f.upButtonActive
}

// IsDownButtonActive checks if the down button is active
func (f *Floor) IsDownButtonActive() bool {
	return f.downButtonActive
}

// ResetButtons resets both call buttons
func (f *Floor) ResetButtons() {
	f.upButtonActive = false
	f.downButtonActive = false
}

func NewFloorButtonsResetEvent(floorNum int) Event {
    return Event{
        Type: FloorButtonsReset,
        Payload: struct {
            FloorNumber int
        }{
            FloorNumber: floorNum,
        },
    }
}

func NewLiftAssignedEvent(liftID string, floorNum int) Event {
    return Event{
        Type: LiftAssigned,
        Payload: struct {
            LiftID      string
            FloorNumber int
        }{
            LiftID:      liftID,
            FloorNumber: floorNum,
        },
    }
}
// Add this setter method
func (f *Floor) SetNumber(number int) {
    f.number = number
}
