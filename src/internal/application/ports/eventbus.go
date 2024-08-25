package ports

import (
	"context"

	"github.com/Avyukth/lift-simulation/internal/domain"
)

// EventHandler is a function that handles a specific event
type EventHandler func(ctx context.Context, event domain.Event) error

// EventBus defines the interface for event publication and subscription
type EventBus interface {
	// Publish publishes an event to the event bus
	Publish(ctx context.Context, event domain.Event) error

	// Subscribe registers an event handler for a specific event type
	Subscribe(eventType domain.EventType, handler EventHandler) error

	// Unsubscribe removes an event handler for a specific event type
	Unsubscribe(eventType domain.EventType, handler EventHandler) error
}

// AsyncEventBus is an extension of EventBus that supports asynchronous event handling
type AsyncEventBus interface {
	EventBus

	// PublishAsync publishes an event asynchronously
	PublishAsync(ctx context.Context, event domain.Event) error

	// SubscribeAsync registers an asynchronous event handler for a specific event type
	SubscribeAsync(eventType domain.EventType, handler EventHandler) error
}

// EventBusWithRetry is an extension of EventBus that supports retry mechanisms
type EventBusWithRetry interface {
	EventBus

	// PublishWithRetry publishes an event with retry mechanism
	PublishWithRetry(ctx context.Context, event domain.Event, maxRetries int) error
}

// EventStore defines the interface for persisting and retrieving events
type EventStore interface {
	// SaveEvent persists an event to the event store
	SaveEvent(ctx context.Context, event domain.Event) error

	// GetEvents retrieves events from the event store based on criteria
	GetEvents(ctx context.Context, criteria EventCriteria) ([]domain.Event, error)
}

// EventCriteria defines the criteria for retrieving events
type EventCriteria struct {
	EventTypes []domain.EventType
	FromTime   int64
	ToTime     int64
	Limit      int
}

// ComprehensiveEventBus combines all event bus functionalities
type ComprehensiveEventBus interface {
	AsyncEventBus
	EventBusWithRetry
	EventStore
}
