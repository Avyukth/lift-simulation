package domain

import "errors"

// System represents the entire lift system
type System struct {
	ID          string
	TotalFloors int
	TotalLifts  int
}

// NewSystem creates a new System instance
func NewSystem(systemID string, floors, lifts int) (*System, error) {
	if floors < 2 {
		return nil, errors.New("system must have at least 2 floors")
	}
	if lifts < 1 {
		return nil, errors.New("system must have at least 1 lift")
	}
	return &System{
		ID:          systemID,
		TotalFloors: floors,
		TotalLifts:  lifts,
	}, nil
}

// GetTotalFloors returns the total number of floors in the system
func (s *System) GetTotalFloors() int {
	return s.TotalFloors
}

// GetTotalLifts returns the total number of lifts in the system
func (s *System) GetTotalLifts() int {
	return s.TotalLifts
}

// SystemStatus represents the current status of the lift system
type SystemStatus struct {
	SystemID         string
	TotalFloors      int
	TotalLifts       int
	OperationalLifts int
	ActiveFloorCalls int
}

// NewSystemStatus creates a new SystemStatus instance
func NewSystemStatus(system *System, operationalLifts, activeFloorCalls int) *SystemStatus {
	return &SystemStatus{
		SystemID:         system.ID,
		TotalFloors:      system.TotalFloors,
		TotalLifts:       system.TotalLifts,
		OperationalLifts: operationalLifts,
		ActiveFloorCalls: activeFloorCalls,
	}
}
