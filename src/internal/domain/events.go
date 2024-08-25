// File: internal/domain/events.go

package domain

// Event types for the lift system
type EventType string

const (
	LiftMoved                EventType = "LIFT_MOVED"
	LiftCalled               EventType = "LIFT_CALLED"
	SystemConfigured         EventType = "SYSTEM_CONFIGURED"
	SystemReset              EventType = "SYSTEM_RESET"
	FloorButtonsReset        EventType = "FLOOR_BUTTONS_RESET"
	LiftAssigned             EventType = "LIFT_ASSIGNED"
	LiftArrived              EventType = "LIFT_ARRIVED"
	TrafficSimulationStarted EventType = "TRAFFIC_SIMULATION_STARTED"
)

// Event represents a domain event in the lift system
type Event struct {
	Type    EventType
	Payload interface{}
}

// TrafficSimulationEvent represents the event of starting a traffic simulation
type TrafficSimulationEvent struct {
	SystemID  string
	Intensity string
	Duration  int
}

// NewTrafficSimulationEvent creates a new TrafficSimulationEvent
func NewTrafficSimulationEvent(systemID string, intensity string, duration int) Event {
	return Event{
		Type: TrafficSimulationStarted,
		Payload: TrafficSimulationEvent{
			SystemID:  systemID,
			Intensity: intensity,
			Duration:  duration,
		},
	}
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
func NewLiftCalledEvent(floorID string, floor int, direction Direction) Event {
	return Event{
		Type: LiftCalled,
		Payload: struct {
			FloorID   string
			Floor     int
			Direction Direction
		}{
			FloorID:   floorID,
			Floor:     floor,
			Direction: direction,
		},
	}
}

// NewSystemConfiguredEvent creates a new SystemConfigured event
func NewSystemConfiguredEvent(systemID string, floors, lifts int) Event {
	return Event{
		Type: SystemConfigured,
		Payload: struct {
			SystemID string
			Floors   int
			Lifts    int
		}{
			SystemID: systemID,
			Floors:   floors,
			Lifts:    lifts,
		},
	}
}

// NewSystemResetEvent creates a new SystemReset event
func NewSystemResetEvent(systemID string) Event {
	return Event{
		Type: SystemReset,
		Payload: struct {
			SystemID string
		}{
			SystemID: systemID,
		},
	}
}

func NewLiftArrivedEvent(liftID, floorID string, floorNumber int) Event {
	return Event{
		Type: LiftArrived,
		Payload: struct {
			LiftID      string
			FloorID     string
			FloorNumber int
		}{
			LiftID:      liftID,
			FloorID:     floorID,
			FloorNumber: floorNumber,
		},
	}
}
