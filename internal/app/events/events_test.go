package events

import (
	"testing"
)

func TestOwnedEvent_AudioPlaybackEventRecord_GetOwnerID(t *testing.T) {
	event := AudioPlaybackEventRecord{
		EventType: AudioPlaybackStartedEvent,
		Path:      "/test/audio.wav",
		OwnerID:   "window-123",
	}

	ownerID := event.GetOwnerID()
	if ownerID != "window-123" {
		t.Errorf("Expected OwnerID='window-123', got '%s'", ownerID)
	}
}

func TestOwnedEvent_AudioPlaybackEventRecord_EmptyOwnerID(t *testing.T) {
	event := AudioPlaybackEventRecord{
		EventType: AudioPlaybackStartedEvent,
		Path:      "/test/audio.wav",
		OwnerID:   "",
	}

	ownerID := event.GetOwnerID()
	if ownerID != "" {
		t.Errorf("Expected empty OwnerID, got '%s'", ownerID)
	}
}

func TestOwnedEvent_MidiPlaybackEventRecord_GetOwnerID(t *testing.T) {
	event := MidiPlaybackEventRecord{
		EventType: MidiPlaybackNoteOnEvent,
		OwnerID:   "midi-controller-456",
	}

	ownerID := event.GetOwnerID()
	if ownerID != "midi-controller-456" {
		t.Errorf("Expected OwnerID='midi-controller-456', got '%s'", ownerID)
	}
}

func TestOwnedEvent_PadGridSelectEvent_GetOwnerID(t *testing.T) {
	event := PadGridSelectEvent{
		Pad:     nil,
		OwnerID: "preset-window-789",
	}

	ownerID := event.GetOwnerID()
	if ownerID != "preset-window-789" {
		t.Errorf("Expected OwnerID='preset-window-789', got '%s'", ownerID)
	}
}

func TestEventTypes_AudioPlaybackEventRecord_Type(t *testing.T) {
	tests := []struct {
		eventType AudioPlaybackEvent
		expected  string
	}{
		{AudioPlaybackStartedEvent, AudioPlaybackStartedKey},
		{AudioPlaybackProgressEvent, AudioPlaybackProgressKey},
		{AudioPlaybackStoppedEvent, AudioPlaybackStoppedKey},
		{AudioPlaybackPausedEvent, AudioPlaybackPausedKey},
		{AudioPlaybackResumedEvent, AudioPlaybackResumedKey},
		{AudioPlaybackFinishedEvent, AudioPlaybackFinishedKey},
	}

	for _, tt := range tests {
		event := AudioPlaybackEventRecord{EventType: tt.eventType}
		if event.Type() != tt.expected {
			t.Errorf("Event type %d: expected %s, got %s",
				tt.eventType, tt.expected, event.Type())
		}
	}
}

func TestEventTypes_AudioPlaybackEventRecord_UnknownType(t *testing.T) {
	event := AudioPlaybackEventRecord{EventType: AudioPlaybackEvent(999)}
	eventType := event.Type()
	if eventType != "audio.playback.unknown" {
		t.Errorf("Expected 'audio.playback.unknown', got '%s'", eventType)
	}
}

func TestEventTypes_PadGridSelectEvent_Type(t *testing.T) {
	event := PadGridSelectEvent{}
	eventType := event.Type()
	if eventType != PadGridSelectKey {
		t.Errorf("Expected %s, got %s", PadGridSelectKey, eventType)
	}
}

func TestEventTypes_MidiPlaybackEventRecord_Type(t *testing.T) {
	tests := []struct {
		eventType MidiPlaybackEvent
		expected  string
	}{
		{MidiPlaybackNoteOnEvent, MidiPlaybackNoteOnKey},
		{MidiPlaybackNoteOffEvent, MidiPlaybackNoteOffKey},
		{MidiPlaybackCCEvent, MidiPlaybackCCKey},
		{MidiPlaybackCCEvent, MidiPlaybackCCKey},
	}

	for _, tt := range tests {
		event := MidiPlaybackEventRecord{EventType: tt.eventType}
		if event.Type() != tt.expected {
			t.Errorf("MIDI event type %d: expected %s, got %s",
				tt.eventType, tt.expected, event.Type())
		}
	}
}

func TestEventTypes_AudioLoadEventRecord_Type(t *testing.T) {
	tests := []struct {
		eventType AudioLoadEvent
		expected  string
	}{
		{AudioMetadataLoadedEvent, AudioMetadataLoadedKey},
		{AudioSamplesLoadedEvent, AudioSamplesLoadedKey},
		{AudioLoadFailedEvent, AudioLoadFailedKey},
	}

	for _, tt := range tests {
		event := AudioLoadEventRecord{EventType: tt.eventType}
		if event.Type() != tt.expected {
			t.Errorf("Audio load event type %d: expected %s, got %s",
				tt.eventType, tt.expected, event.Type())
		}
	}
}

func TestEventTypes_LibraryScanEvents_Type(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected string
	}{
		{"LibraryScanStartedEvent", LibraryScanStartedEvent{}, LibraryScanStartedKey},
		{"LibraryScanProgressEvent", LibraryScanProgressEvent{}, LibraryScanProgressKey},
		{"LibraryScanCompletedEvent", LibraryScanCompletedEvent{}, LibraryScanCompletedKey},
		{"LibraryScanFailedEvent", LibraryScanFailedEvent{}, LibraryScanFailedKey},
	}

	for _, tt := range tests {
		if tt.event.Type() != tt.expected {
			t.Errorf("%s: expected %s, got %s",
				tt.name, tt.expected, tt.event.Type())
		}
	}
}

func TestEventTypes_StorageEvents_Type(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected string
	}{
		{"StorageActivatedEvent", StorageActivatedEvent{}, StorageActivatedEventKey},
		{"StorageDeselectedEvent", StorageDeselectedEvent{}, StorageDeselectedEventKey},
		{"StorageMountedEvent", StorageMountedEvent{}, StorageMountedEventKey},
		{"StorageUnmountedEvent", StorageUnmountedEvent{}, StorageUnmountedEventKey},
	}

	for _, tt := range tests {
		if tt.event.Type() != tt.expected {
			t.Errorf("%s: expected %s, got %s",
				tt.name, tt.expected, tt.event.Type())
		}
	}
}

func TestEventTypes_PresetEvents_Type(t *testing.T) {
	event := PresetLoadEvent{}
	if event.Type() != PresetLoadEventKey {
		t.Errorf("Expected %s, got %s", PresetLoadEventKey, event.Type())
	}
}

func TestEventTypes_WindowEvents_Type(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected string
	}{
		{"WindowOpenEvent", WindowOpenEvent{}, WindowOpenEventKey},
		{"WindowCloseEvent", WindowCloseEvent{}, WindowCloseEventKey},
		{"WindowDestroyEvent", WindowDestroyEvent{}, WindowDestroyEventKey},
	}

	for _, tt := range tests {
		if tt.event.Type() != tt.expected {
			t.Errorf("%s: expected %s, got %s",
				tt.name, tt.expected, tt.event.Type())
		}
	}
}

func TestEventTypes_ComboboxSelectionChangeEvent_Type(t *testing.T) {
	event := ComboboxSelectionChangeEvent{}
	eventType := event.Type()
	if eventType != ComboboxSelectionChangeEventKey {
		t.Errorf("Expected %s, got %s", ComboboxSelectionChangeEventKey, eventType)
	}
}

func TestEventTypes_AudioVolumeEvents_Type(t *testing.T) {
	event := AudioVolumeChangedEvent{}
	if event.Type() != AudioVolumeChangedKey {
		t.Errorf("Expected %s, got %s", AudioVolumeChangedKey, event.Type())
	}
}

func TestEventTypes_ComponentEvents_Type(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected string
	}{
		{"ComponentClickEvent", ComponentClickEvent{}, ComponentClickEventKey},
		{"ComponentHoverInEvent", ComponentHoverInEvent{}, ComponentHoverInEventKey},
		{"ComponentHoverOutEvent", ComponentHoverOutEvent{}, ComponentHoverOutEventKey},
	}

	for _, tt := range tests {
		if tt.event.Type() != tt.expected {
			t.Errorf("%s: expected %s, got %s",
				tt.name, tt.expected, tt.event.Type())
		}
	}
}

func TestEventTypes_AllEventKeys_NotEmpty(t *testing.T) {
	// Verify all event keys are defined and non-empty
	keys := []string{
		AudioPlaybackStartedKey,
		AudioPlaybackProgressKey,
		AudioPlaybackStoppedKey,
		AudioMetadataLoadedKey,
		AudioSamplesLoadedKey,
		AudioLoadFailedKey,
		LibraryScanStartedKey,
		LibraryScanProgressKey,
		LibraryScanCompletedKey,
		PadGridSelectKey,
		PresetLoadEventKey,
		StorageActivatedEventKey,
		WindowOpenEventKey,
		WindowCloseEventKey,
		WindowDestroyEventKey,
		ComponentClickEventKey,
		AudioVolumeChangedKey,
		ComboboxSelectionChangeEventKey,
	}

	for _, key := range keys {
		if key == "" {
			t.Errorf("Found empty event key")
		}
	}
}

func TestEventTypes_AudioPlaybackKeys_Unique(t *testing.T) {
	keys := []string{
		AudioPlaybackStartedKey,
		AudioPlaybackProgressKey,
		AudioPlaybackStoppedKey,
		AudioPlaybackPausedKey,
		AudioPlaybackResumedKey,
		AudioPlaybackFinishedKey,
	}

	seen := make(map[string]bool)
	for _, key := range keys {
		if seen[key] {
			t.Errorf("Duplicate event key: %s", key)
		}
		seen[key] = true
	}
}
