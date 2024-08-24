package domain

import "errors"

// System represents the entire lift system
type System struct {
	totalFloors int
	totalLifts  int
}

// NewSystem creates a new System instance
func NewSystem(floors, lifts int) (*System, error) {
	if floors < 2 {
		return nil, errors.New("system must have at least 2 floors")
	}
	if lifts < 1 {
		return nil, errors.New("system must have at least 1 lift")
	}
	return &System{
		totalFloors: floors,
		totalLifts:  lifts,
	}, nil
}

// TotalFloors returns the total number of floors in the system
func (s *System) TotalFloors() int {
	return s.totalFloors
}

// TotalLifts returns the total number of lifts in the system
func (s *System) TotalLifts() int {
	return s.totalLifts
}

// SystemStatus represents the current status of the lift system
type SystemStatus struct {
	TotalFloors       int
	TotalLifts        int
	OperationalLifts  int
	ActiveFloorCalls  int
}

// NewSystemStatus creates a new SystemStatus instance
func NewSystemStatus(system *System, operationalLifts, activeFloorCalls int) *SystemStatus {
	return &SystemStatus{
		TotalFloors:      system.totalFloors,
		TotalLifts:       system.totalLifts,
		OperationalLifts: operationalLifts,
		ActiveFloorCalls: activeFloorCalls,
	}
}

// Event types for the lift system
type EventType string

const (
	LiftMoved          EventType = "LIFT_MOVED"
	LiftCalled         EventType = "LIFT_CALLED"
	SystemConfigured   EventType = "SYSTEM_CONFIGURED"
	SystemReset        EventType = "SYSTEM_RESET"
	FloorButtonsReset  EventType = "FLOOR_BUTTONS_RESET"
	LiftAssigned       EventType = "LIFT_ASSIGNED"
)

// Event represents a domain event in the lift system
type Event struct {
	Type    EventType
	Payload interface{}
}

// NewLiftMovedEvent creates a new LiftMoved event
func NewLiftMovedEvent(liftID string, floor int) Event {
	return Event{
		Type: LiftMoved,
		Payload: struct {
			LiftID string
			Floor  int
		}{
			LiftID: liftID,
			Floor:  floor,
		},
	}
}

// NewLiftCalledEvent creates a new LiftCalled event
func NewLiftCalledEvent(floor int, direction Direction) Event {
	return Event{
		Type: LiftCalled,
		Payload: struct {
			Floor     int
			Direction Direction
		}{
			Floor:     floor,
			Direction: direction,
		},
	}
}

// NewSystemConfiguredEvent creates a new SystemConfigured event
func NewSystemConfiguredEvent(floors, lifts int) Event {
	return Event{
		Type: SystemConfigured,
		Payload: struct {
			Floors int
			Lifts  int
		}{
			Floors: floors,
			Lifts:  lifts,
		},
	}
}

// NewSystemResetEvent creates a new SystemReset event
func NewSystemResetEvent() Event {
	return Event{
		Type:    SystemReset,
		Payload: nil,
	}
}
