package domain

type EventType int

const (
	LiftRequested EventType = iota
	LiftArrived
	LiftAssigned
	FloorButtonPressed
	FloorAtCapacity
)

func (e EventType) String() string {
	return [...]string{"LiftRequested", "LiftArrived", "LiftAssigned", "FloorButtonPressed", "FloorAtCapacity"}[e]
}

type Event interface {
	Type() EventType
}

type LiftRequestedEvent struct {
	FloorNumber int
	Direction   Direction
}

func (e LiftRequestedEvent) Type() EventType {
	return LiftRequested
}

type LiftArrivedEvent struct {
	LiftID      string
	FloorNumber int
}

func (e LiftArrivedEvent) Type() EventType {
	return LiftArrived
}

type LiftAssignedEvent struct {
	LiftID      string
	FloorNumber int
	Direction   Direction
}

func (e LiftAssignedEvent) Type() EventType {
	return LiftAssigned
}

type FloorAtCapacityEvent struct {
	FloorNumber int
}

func (e FloorAtCapacityEvent) Type() EventType {
	return FloorAtCapacity
}
