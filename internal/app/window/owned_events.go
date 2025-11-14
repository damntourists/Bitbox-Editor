package window

import (
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
)

// OwnedEventsMixin provides methods for windows to work with owned events
type OwnedEventsMixin struct {
	filteredSub *eventbus.FilteredSubscription
	ownerID     string
}

// InitOwnedEvents initializes the owned events for a window. Should be called in window constructor.
func (m *OwnedEventsMixin) InitOwnedEvents(ownerID string, bufferSize int) {
	m.ownerID = ownerID
	m.filteredSub = eventbus.NewFilteredSubscription(ownerID, bufferSize)
}

// SubscribeToAudioEvents subscribes to all audio playback events
func (m *OwnedEventsMixin) SubscribeToAudioEvents(bus *eventbus.EventBus) {
	if m.filteredSub == nil {
		panic("OwnedEventsMixin: must call InitOwnedEvents first")
	}
	m.filteredSub.SubscribeMultiple(bus,
		events.AudioPlaybackProgressKey,
		events.AudioPlaybackStartedKey,
		events.AudioPlaybackPausedKey,
		events.AudioPlaybackStoppedKey,
		events.AudioPlaybackFinishedKey,
	)
}

// SubscribeToMidiEvents subscribes to all MIDI playback events
func (m *OwnedEventsMixin) SubscribeToMidiEvents(bus *eventbus.EventBus) {
	if m.filteredSub == nil {
		panic("OwnedEventsMixin: must call InitOwnedEvents first")
	}
	m.filteredSub.SubscribeMultiple(bus,
		events.MidiPlaybackNoteOnKey,
		events.MidiPlaybackNoteOffKey,
		events.MidiPlaybackCCKey,
	)
}

// SubscribeToCustomEvents subscribes to custom event types
func (m *OwnedEventsMixin) SubscribeToCustomEvents(bus *eventbus.EventBus, eventTypes ...string) {
	if m.filteredSub == nil {
		panic("OwnedEventsMixin: must call InitOwnedEvents first")
	}
	m.filteredSub.SubscribeMultiple(bus, eventTypes...)
}

// FilteredEvents returns the channel to receive filtered events
func (m *OwnedEventsMixin) FilteredEvents() <-chan events.Event {
	if m.filteredSub == nil {
		panic("OwnedEventsMixin: must call InitOwnedEvents first")
	}
	return m.filteredSub.Events()
}

// CleanupOwnedEvents unsubscribes and cleans up resources
func (m *OwnedEventsMixin) CleanupOwnedEvents() {
	if m.filteredSub != nil {
		m.filteredSub.Unsubscribe()
		m.filteredSub = nil
	}
}

// OwnerID returns the owner ID for this window
func (m *OwnedEventsMixin) OwnerID() string {
	return m.ownerID
}
