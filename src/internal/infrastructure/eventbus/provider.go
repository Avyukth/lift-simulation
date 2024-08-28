package eventbus

import (
	"github.com/Avyukth/lift-simulation/internal/application/events"
)

func ProvideEventBus() events.EventBus {
	return events.NewInMemoryEventBus()
}
