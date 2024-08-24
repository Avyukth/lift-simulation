package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/Avyukth/lift-simulation/internal/application/ports"
	"github.com/Avyukth/lift-simulation/internal/domain"
	"github.com/redis/go-redis/v9"
)

// RedisPubSub implements the EventBus interface using Redis Pub/Sub
type RedisPubSub struct {
	client       *redis.Client
	subscribers  map[domain.EventType][]ports.EventHandler
	subscriberMu sync.RWMutex
}

// NewRedisPubSub creates a new instance of RedisPubSub
func NewRedisPubSub(redisURL string) (*RedisPubSub, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	pubsub := &RedisPubSub{
		client:      client,
		subscribers: make(map[domain.EventType][]ports.EventHandler),
	}

	go pubsub.subscribe()

	return pubsub, nil
}

// Publish publishes an event to Redis
func (r *RedisPubSub) Publish(ctx context.Context, event domain.Event) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = r.client.Publish(ctx, string(event.Type), eventJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// Subscribe registers an event handler for a specific event type
func (r *RedisPubSub) Subscribe(eventType domain.EventType, handler ports.EventHandler) error {
	r.subscriberMu.Lock()
	defer r.subscriberMu.Unlock()

	r.subscribers[eventType] = append(r.subscribers[eventType], handler)
	return nil
}

// Unsubscribe removes an event handler for a specific event type
func (r *RedisPubSub) Unsubscribe(eventType domain.EventType, handler ports.EventHandler) error {
	r.subscriberMu.Lock()
	defer r.subscriberMu.Unlock()

	handlers := r.subscribers[eventType]
	for i, h := range handlers {
		if h == handler {
			r.subscribers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
	return nil
}

// subscribe listens for messages on all event types
func (r *RedisPubSub) subscribe() {
	pubsub := r.client.PSubscribe(context.Background(), "*")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		var event domain.Event
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}

		r.subscriberMu.RLock()
		handlers := r.subscribers[event.Type]
		r.subscriberMu.RUnlock()

		for _, handler := range handlers {
			go func(h ports.EventHandler) {
				if err := h(context.Background(), event); err != nil {
					log.Printf("Error handling event: %v", err)
				}
			}(handler)
		}
	}
}

// Close closes the Redis connection
func (r *RedisPubSub) Close() error {
	return r.client.Close()
}

// Ensure RedisPubSub implements ports.EventBus interface
var _ ports.EventBus = (*RedisPubSub)(nil)
