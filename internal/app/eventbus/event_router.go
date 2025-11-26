package eventbus

import (
	"bitbox-editor/internal/app/events"
)

// EventHandlerFunc is a function that handles a specific event type
type EventHandlerFunc func(event events.Event)

type EventRouter struct {
	ownerID  string
	sub      *FilteredSubscription
	handlers map[string]EventHandlerFunc
}

// Init initializes the event handler with an owner ID for event filtering.
func (eh *EventRouter) Init(ownerID string) {
	eh.ownerID = ownerID
	eh.handlers = make(map[string]EventHandlerFunc)
}

// OnEvent registers a handler function for a specific event type.
func (eh *EventRouter) OnEvent(eventKey string, handler EventHandlerFunc) {
	// Create subscription on first event registration
	if eh.sub == nil {
		eh.sub = NewFilteredSubscription(eh.ownerID, 10)
	}

	// Register handler
	eh.handlers[eventKey] = handler

	// Subscribe to this event type
	eh.sub.Subscribe(Bus, eventKey)
}

// ProcessEvents drains all pending events and routes them to registered handlers.
func (eh *EventRouter) ProcessEvents() {
	if eh.sub == nil {
		return
	}

	// Drain all available events from the channel
	for {
		select {
		case event := <-eh.sub.Events():
			if handler, ok := eh.handlers[event.Type()]; ok {
				handler(event)
			}
		default:
			// No more events available
			return
		}
	}
}

// Destroy unsubscribes from all events and cleans up resources.
func (eh *EventRouter) Destroy() {
	if eh.sub != nil {
		eh.sub.Unsubscribe()
		eh.sub = nil
	}
	eh.handlers = nil
}

// OwnerID returns the component's owner ID used for event filtering
func (eh *EventRouter) OwnerID() string {
	return eh.ownerID
}
