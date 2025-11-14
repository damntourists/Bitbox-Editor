package events

type PresetEvent int32

// PresetEvent Event Enums

const (
	PresetLoadEvent PresetEvent = iota
)

// PresetEvent Event Keys

const (
	PresetLoadEventKey = "preset.load"
)

// PresetEventRecord holds data for preset events.
type PresetEventRecord struct {
	// EventType is the enum value (e.g., PresetLoadEvent).
	EventType PresetEvent
	// TODO: Replace Data interface{} with more specific fields.
	// Data is the event-specific data.
	Data interface{}
}

// Type implements the events.Event interface
func (e PresetEventRecord) Type() string {
	switch e.EventType {
	case PresetLoadEvent:
		return PresetLoadEventKey
	default:
		return "preset.unknown"
	}
}
