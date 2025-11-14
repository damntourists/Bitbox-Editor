package eventbus

import (
	"bitbox-editor/internal/app/events"
	"reflect"
	"sync"
	"time"
)

// OwnedEvent is an interface that events can implement to support owner-based filtering
type OwnedEvent interface {
	events.Event
	GetOwnerID() string
}

// FilteredSubscription manages a subscription with automatic owner-based filtering
type FilteredSubscription struct {
	ownerID       string
	eventChan     chan events.Event
	internalChans []chan events.Event // All internal channels merged into one output
	unsubFuncs    []func()
	mu            sync.Mutex
	started       bool
}

// NewFilteredSubscription creates a new filtered subscription for the given owner
// The owner will only receive events that either:
//  1. Implement OwnedEvent and have a matching OwnerID
//  2. Don't implement OwnedEvent (broadcast events)
func NewFilteredSubscription(ownerID string, bufferSize int) *FilteredSubscription {
	return &FilteredSubscription{
		ownerID:       ownerID,
		eventChan:     make(chan events.Event, bufferSize),
		internalChans: make([]chan events.Event, 0),
		unsubFuncs:    make([]func(), 0),
		started:       false,
	}
}

// Subscribe registers interest in an event type with automatic filtering
func (fs *FilteredSubscription) Subscribe(bus *EventBus, eventType string) {
	internalChan := make(chan events.Event, cap(fs.eventChan))

	bus.Subscribe(eventType, fs.ownerID, internalChan)

	fs.mu.Lock()

	fs.internalChans = append(fs.internalChans, internalChan)
	fs.unsubFuncs = append(fs.unsubFuncs, func() {
		bus.Unsubscribe(eventType, fs.ownerID)
		close(internalChan)
	})

	// Start the filtering goroutine once when first subscription is made
	if !fs.started {
		fs.started = true
		go fs.filterAllEvents()
	}
	fs.mu.Unlock()
}

// filterAllEvents multiplexes all internal channels and filters events
func (fs *FilteredSubscription) filterAllEvents() {
	fs.mu.Lock()

	cases := make([]reflect.SelectCase, len(fs.internalChans))
	for i, ch := range fs.internalChans {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
	}

	fs.mu.Unlock()

	// If no channels, exit
	if len(cases) == 0 {
		return
	}

	// Multiplex all channels and filter events
	for {
		// Wait for any channel to receive an event
		chosen, value, ok := reflect.Select(cases)

		if !ok {
			// Channel closed - remove it from cases
			cases[chosen] = cases[len(cases)-1]
			cases = cases[:len(cases)-1]

			// If no channels left, exit
			if len(cases) == 0 {
				return
			}
			continue
		}

		// Extract the event
		event, ok := value.Interface().(events.Event)
		if !ok {
			// Not an event, skip
			continue
		}

		// Apply filtering and forward if appropriate
		fs.filterAndForward(event)
	}
}

// filterAndForward applies owner filtering logic and forwards matching events
func (fs *FilteredSubscription) filterAndForward(event events.Event) {
	// Check if event implements OwnedEvent
	if ownedEvent, ok := event.(OwnedEvent); ok {
		// Only forward if owner matches or if empty (broadcast)
		ownerID := ownedEvent.GetOwnerID()
		shouldForward := ownerID == "" || ownerID == fs.ownerID

		if shouldForward {
			select {
			case fs.eventChan <- event:
			case <-time.After(10 * time.Millisecond):
				// Channel full or slow receiver - drop event with timeout
			}
		}
		// Silently drop events for other owners
	} else {
		// Not an owned event - broadcast
		select {
		case fs.eventChan <- event:
		case <-time.After(10 * time.Millisecond):
			// Channel full - drop event with timeout
		}
	}
}

// Events returns the channel that receives filtered events
func (fs *FilteredSubscription) Events() <-chan events.Event {
	return fs.eventChan
}

// Unsubscribe removes all subscriptions and closes channels
func (fs *FilteredSubscription) Unsubscribe() {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Call all unsubscribe functions
	for _, unsub := range fs.unsubFuncs {
		unsub()
	}
	fs.unsubFuncs = nil

	// Close the output channel
	close(fs.eventChan)
}

// SubscribeMultiple subscribes to multiple event types at once
func (fs *FilteredSubscription) SubscribeMultiple(bus *EventBus, eventTypes ...string) {
	for _, eventType := range eventTypes {
		fs.Subscribe(bus, eventType)
	}
}
