package events

// MidiEventType defines the type of MIDI message.
type MidiEventType int

// Midi Event Enums
const (
	MidiNoteOnEvent MidiEventType = iota
	MidiNoteOffEvent
	MidiControlChangeEvent
	MidiPitchBendEvent
	MidiProgramChangeEvent
	MidiAfterTouchEvent
	MidiPolyAfterTouchEvent
)

// Midi Event Keys
const (
	MidiNoteOnKey         = "midi.note_on"
	MidiNoteOffKey        = "midi.note_off"
	MidiControlChangeKey  = "midi.control_change"
	MidiPitchBendKey      = "midi.pitch_bend"
	MidiProgramChangeKey  = "midi.program_change"
	MidiAfterTouchKey     = "midi.after_touch"
	MidiPolyAfterTouchKey = "midi.poly_after_touch"
	MidiUnknownKey        = "midi.unknown"
)

// MidiEventRecord holds parsed data for a single MIDI message
type MidiEventRecord struct {
	// EventType is the enum value (e.g., MidiNoteOnEvent).
	EventType MidiEventType
	// Timestamp is the timestamp from the MIDI driver.
	Timestamp int32
	// PortID is the name of the port this message came from.
	PortID string

	// Channel is the MIDI channel (0-15).
	Channel uint8

	// For NoteOn/NoteOff/PolyAfterTouch
	Key      uint8
	Velocity uint8

	// For ControlChange
	Controller uint8
	Value      uint8

	// For PitchBend (14-bit value, 0-16383)
	Value14 uint16
}

// Type implements the events.Event interface
func (e MidiEventRecord) Type() string {
	switch e.EventType {
	case MidiNoteOnEvent:
		return MidiNoteOnKey
	case MidiNoteOffEvent:
		return MidiNoteOffKey
	case MidiControlChangeEvent:
		return MidiControlChangeKey
	case MidiPitchBendEvent:
		return MidiPitchBendKey
	case MidiProgramChangeEvent:
		return MidiProgramChangeKey
	case MidiAfterTouchEvent:
		return MidiAfterTouchKey
	case MidiPolyAfterTouchEvent:
		return MidiPolyAfterTouchKey
	default:
		return MidiUnknownKey
	}
}
