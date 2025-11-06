package eventbus

import (
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/logging"
	"sync"
)

var log = logging.NewLogger("bus")

var Bus = NewEventBus()

type EventChannel chan events.Event

type EventBus struct {
	mu sync.RWMutex
	// Map of: eventType -> subscriberID -> channel
	subscribers map[string]map[string]EventChannel
}

// Subscribe adds a channel to the list for a specific event type
func (bus *EventBus) Subscribe(eventType string, subscriberID string, ch EventChannel) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if _, ok := bus.subscribers[eventType]; !ok {
		bus.subscribers[eventType] = make(map[string]EventChannel)
	}
	bus.subscribers[eventType][subscriberID] = ch
}

// Unsubscribe removes a specific subscriber's channel
func (bus *EventBus) Unsubscribe(eventType string, subscriberID string) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if subs, ok := bus.subscribers[eventType]; ok {
		delete(subs, subscriberID)
		if len(subs) == 0 {
			delete(bus.subscribers, eventType)
		}
	}
}

// Publish sends an event to all subscribed channels
func (bus *EventBus) Publish(event events.Event) {
	bus.mu.RLock()

	subs, ok := bus.subscribers[event.Type()]
	if !ok {
		bus.mu.RUnlock()
		return
	}

	channelsToPublish := make([]EventChannel, 0, len(subs))
	for _, ch := range subs {
		channelsToPublish = append(channelsToPublish, ch)
	}

	bus.mu.RUnlock()

	for _, ch := range channelsToPublish {
		select {
		case ch <- event:
		default:
			// Subscriber's channel is full, drop event.
		}
	}
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string]map[string]EventChannel),
	}
}
