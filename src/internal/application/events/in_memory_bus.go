package events

import (
	"sync"

	"github.com/Avyukth/lift-simulation/internal/domain"
)

type InMemoryEventBus struct {
	handlers map[domain.EventType][]EventHandler
	mu       sync.RWMutex
}

func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[domain.EventType][]EventHandler),
	}
}

func (b *InMemoryEventBus) Publish(event domain.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if handlers, exists := b.handlers[event.Type()]; exists {
		for _, handler := range handlers {
			go handler.Handle(event)
		}
	}
}

func (b *InMemoryEventBus) Subscribe(eventType domain.EventType, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *InMemoryEventBus) Unsubscribe(eventType domain.EventType, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if handlers, exists := b.handlers[eventType]; exists {
		for i, h := range handlers {
			if h == handler {
				b.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
}
