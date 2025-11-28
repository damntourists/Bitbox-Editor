package events

import (
	"testing"
)

func TestOwnedEvent_AudioPlaybackEventRecord_GetOwnerID(t *testing.T) {
	event := AudioPlaybackStartedEvent{
		AudioPlaybackProgressBase: AudioPlaybackProgressBase{
			AudioPlaybackBase: AudioPlaybackBase{
				Path:    "/test/audio.wav",
				OwnerID: "window-123",
			},
		},
	}

	ownerID := event.GetOwnerID()
	if ownerID != "window-123" {
		t.Errorf("Expected OwnerID='window-123', got '%s'", ownerID)
	}
}

func TestOwnedEvent_AudioPlaybackEventRecord_EmptyOwnerID(t *testing.T) {
	event := AudioPlaybackStartedEvent{
		AudioPlaybackProgressBase: AudioPlaybackProgressBase{
			AudioPlaybackBase: AudioPlaybackBase{
				Path:    "/test/audio.wav",
				OwnerID: "",
			},
		},
	}

	ownerID := event.GetOwnerID()
	if ownerID != "" {
		t.Errorf("Expected empty OwnerID, got '%s'", ownerID)
	}
}

func TestOwnedEvent_MidiPlaybackEventRecord_GetOwnerID(t *testing.T) {
	event := MidiPlaybackNoteOnEvent{
		MidiPlaybackMessageBase: MidiPlaybackMessageBase{
			MidiPlaybackBase: MidiPlaybackBase{
				OwnerID: "midi-controller-456",
			},
			Channel: 0,
		},
		Note:     60,
		Velocity: 100,
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
		name     string
		event    Event
		expected string
	}{
		{
			"AudioPlaybackStarted",
			AudioPlaybackStartedEvent{
				AudioPlaybackProgressBase: AudioPlaybackProgressBase{
					AudioPlaybackBase: AudioPlaybackBase{
						Path:    "/test/audio.wav",
						OwnerID: "test",
					},
				},
			},
			AudioPlaybackStartedKey,
		},
		{
			"AudioPlaybackProgress",
			AudioPlaybackProgressEvent{
				AudioPlaybackProgressBase: AudioPlaybackProgressBase{
					AudioPlaybackBase: AudioPlaybackBase{
						Path:    "/test/audio.wav",
						OwnerID: "test",
					},
					Progress: 0.5,
				},
			},
			AudioPlaybackProgressKey,
		},
		{
			"AudioPlaybackStopped",
			AudioPlaybackStoppedEvent{
				AudioPlaybackProgressBase: AudioPlaybackProgressBase{
					AudioPlaybackBase: AudioPlaybackBase{
						Path:    "/test/audio.wav",
						OwnerID: "test",
					},
				},
			},
			AudioPlaybackStoppedKey,
		},
		{
			"AudioPlaybackPaused",
			AudioPlaybackPausedEvent{
				AudioPlaybackProgressBase: AudioPlaybackProgressBase{
					AudioPlaybackBase: AudioPlaybackBase{
						Path:    "/test/audio.wav",
						OwnerID: "test",
					},
				},
			},
			AudioPlaybackPausedKey,
		},
		{
			"AudioPlaybackResumed",
			AudioPlaybackResumedEvent{
				AudioPlaybackProgressBase: AudioPlaybackProgressBase{
					AudioPlaybackBase: AudioPlaybackBase{
						Path:    "/test/audio.wav",
						OwnerID: "test",
					},
				},
			},
			AudioPlaybackResumedKey,
		},
		{
			"AudioPlaybackFinished",
			AudioPlaybackFinishedEvent{
				AudioPlaybackBase: AudioPlaybackBase{
					Path:    "/test/audio.wav",
					OwnerID: "test",
				},
			},
			AudioPlaybackFinishedKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.event.Type() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.event.Type())
			}
		})
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
		name     string
		event    Event
		expected string
	}{
		{
			"MidiPlaybackStarted",
			MidiPlaybackStartedEvent{
				MidiPlaybackBase: MidiPlaybackBase{
					OwnerID: "test",
				},
			},
			MidiPlaybackStartedKey,
		},
		{
			"MidiPlaybackStopped",
			MidiPlaybackStoppedEvent{
				MidiPlaybackBase: MidiPlaybackBase{
					OwnerID: "test",
				},
			},
			MidiPlaybackStoppedKey,
		},
		{
			"MidiPlaybackNoteOn",
			MidiPlaybackNoteOnEvent{
				MidiPlaybackMessageBase: MidiPlaybackMessageBase{
					MidiPlaybackBase: MidiPlaybackBase{
						OwnerID: "test",
					},
					Channel: 0,
				},
				Note:     60,
				Velocity: 100,
			},
			MidiPlaybackNoteOnKey,
		},
		{
			"MidiPlaybackNoteOff",
			MidiPlaybackNoteOffEvent{
				MidiPlaybackMessageBase: MidiPlaybackMessageBase{
					MidiPlaybackBase: MidiPlaybackBase{
						OwnerID: "test",
					},
					Channel: 0,
				},
				Note:     60,
				Velocity: 0,
			},
			MidiPlaybackNoteOffKey,
		},
		{
			"MidiPlaybackCC",
			MidiPlaybackCCEvent{
				MidiPlaybackMessageBase: MidiPlaybackMessageBase{
					MidiPlaybackBase: MidiPlaybackBase{
						OwnerID: "test",
					},
					Channel: 0,
				},
				CC:    7,
				Value: 127,
			},
			MidiPlaybackCCKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.event.Type() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.event.Type())
			}
		})
	}
}

func TestEventTypes_AudioLoadEventRecord_Type(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected string
	}{
		{
			"AudioMetadataLoaded",
			AudioMetadataLoadedEvent{
				AudioLoadBase: AudioLoadBase{
					Path: "/test/audio.wav",
				},
			},
			AudioMetadataLoadedKey,
		},
		{
			"AudioSamplesLoaded",
			AudioSamplesLoadedEvent{
				AudioLoadBase: AudioLoadBase{
					Path: "/test/audio.wav",
				},
			},
			AudioSamplesLoadedKey,
		},
		{
			"AudioLoadFailed",
			AudioLoadFailedEvent{
				AudioLoadBase: AudioLoadBase{
					Path: "/test/audio.wav",
				},
				Error: nil,
			},
			AudioLoadFailedKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.event.Type() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.event.Type())
			}
		})
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
