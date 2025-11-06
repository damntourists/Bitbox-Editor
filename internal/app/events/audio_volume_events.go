package events

type AudioVolumeEvent int32

// Audio Volume Event Enums
const (
	AudioVolumeChangedEvent AudioVolumeEvent = iota
)

// Audio Volume Event Keys
const (
	AudioVolumeChangedKey = "audio.volume.changed"
)

// AudioVolumeEventRecord holds data for a volume change event.
type AudioVolumeEventRecord struct {
	// EventType is the enum value (e.g., AudioVolumeChangedEvent).
	EventType AudioVolumeEvent
	// Volume is the new volume level, normalized between 0.0 (silent) and 1.0 (full).
	Volume float64
}

// Type implements the events.Event interface
func (e AudioVolumeEventRecord) Type() string {
	switch e.EventType {
	case AudioVolumeChangedEvent:
		return AudioVolumeChangedKey
	default:
		return "audio.volume.unknown"
	}
}
