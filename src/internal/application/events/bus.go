package events

import (
	"github.com/Avyukth/lift-simulation/internal/domain"
)

type EventHandler interface {
	Handle(event domain.Event)
}

type EventBus interface {
	Publish(event domain.Event)
	Subscribe(eventType domain.EventType, handler EventHandler)
	Unsubscribe(eventType domain.EventType, handler EventHandler)
}
