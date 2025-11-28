package events

// Midi Event Keys
const (
	MidiNoteOnKey         = "midi.note_on"
	MidiNoteOffKey        = "midi.note_off"
	MidiControlChangeKey  = "midi.control_change"
	MidiPitchBendKey      = "midi.pitch_bend"
	MidiProgramChangeKey  = "midi.program_change"
	MidiAfterTouchKey     = "midi.after_touch"
	MidiPolyAfterTouchKey = "midi.poly_after_touch"
)

/*
╭───────────────╮
│MidiMessageBase│
╰───────────────╯
*/

// MidiMessageBase contains common fields shared across MIDI events.
type MidiMessageBase struct {
	// Timestamp from MIDI driver
	Timestamp int32
	// PortID is the name of the port this message came from
	PortID string
	// Channel is the MIDI channel (0-15)
	Channel uint8
}

/*
╭───────────────╮
│MidiNoteOnEvent│
╰───────────────╯
*/

// MidiNoteOnEvent is published when a MIDI Note On message is received.
type MidiNoteOnEvent struct {
	MidiMessageBase
	// Key is the note key
	Key uint8
	// Velocity is the note velocity
	Velocity uint8
}

// NewMidiNoteOnEvent creates a MidiNoteOnEvent with the given parameters.
func NewMidiNoteOnEvent(timestamp int32, portID string, channel, key, velocity uint8) MidiNoteOnEvent {
	return MidiNoteOnEvent{
		MidiMessageBase: MidiMessageBase{
			Timestamp: timestamp,
			PortID:    portID,
			Channel:   channel,
		},
		Key:      key,
		Velocity: velocity,
	}
}

// Type implements the events.Event interface
func (e MidiNoteOnEvent) Type() string {
	return MidiNoteOnKey
}

/*
╭────────────────╮
│MidiNoteOffEvent│
╰────────────────╯
*/

// MidiNoteOffEvent is published when a MIDI Note Off message is received.
type MidiNoteOffEvent struct {
	MidiMessageBase
	// Key is the note key
	Key uint8
}

// NewMidiNoteOffEvent creates a MidiNoteOffEvent with the given parameters.
func NewMidiNoteOffEvent(timestamp int32, portID string, channel, key uint8) MidiNoteOffEvent {
	return MidiNoteOffEvent{
		MidiMessageBase: MidiMessageBase{
			Timestamp: timestamp,
			PortID:    portID,
			Channel:   channel,
		},
		Key: key,
	}
}

// Type implements the events.Event interface
func (e MidiNoteOffEvent) Type() string {
	return MidiNoteOffKey
}

/*
╭──────────────────────╮
│MidiControlChangeEvent│
╰──────────────────────╯
*/

// MidiControlChangeEvent is published when a MIDI Control Change message is received.
type MidiControlChangeEvent struct {
	MidiMessageBase
	// Controller is the controller number
	Controller uint8
	// Value is the controller value
	Value uint8
}

// NewMidiControlChangeEvent creates a MidiControlChangeEvent with the given parameters.
func NewMidiControlChangeEvent(timestamp int32, portID string, channel, controller, value uint8) MidiControlChangeEvent {
	return MidiControlChangeEvent{
		MidiMessageBase: MidiMessageBase{
			Timestamp: timestamp,
			PortID:    portID,
			Channel:   channel,
		},
		Controller: controller,
		Value:      value,
	}
}

// Type implements the events.Event interface
func (e MidiControlChangeEvent) Type() string {
	return MidiControlChangeKey
}

/*
╭──────────────────╮
│MidiPitchBendEvent│
╰──────────────────╯
*/

// MidiPitchBendEvent is published when a MIDI Pitch Bend message is received.
type MidiPitchBendEvent struct {
	MidiMessageBase
	// Value14 is the 14-bit pitch bend value (0-16383)
	Value14 uint16
}

// NewMidiPitchBendEvent creates a MidiPitchBendEvent with the given parameters.
func NewMidiPitchBendEvent(timestamp int32, portID string, channel uint8, value14 uint16) MidiPitchBendEvent {
	return MidiPitchBendEvent{
		MidiMessageBase: MidiMessageBase{
			Timestamp: timestamp,
			PortID:    portID,
			Channel:   channel,
		},
		Value14: value14,
	}
}

// Type implements the events.Event interface
func (e MidiPitchBendEvent) Type() string {
	return MidiPitchBendKey
}

/*
╭──────────────────────╮
│MidiProgramChangeEvent│
╰──────────────────────╯
*/

// MidiProgramChangeEvent is published when a MIDI Program Change message is received.
type MidiProgramChangeEvent struct {
	MidiMessageBase
	// Program is the program number
	Program uint8
}

// NewMidiProgramChangeEvent creates a MidiProgramChangeEvent with the given parameters.
func NewMidiProgramChangeEvent(timestamp int32, portID string, channel, program uint8) MidiProgramChangeEvent {
	return MidiProgramChangeEvent{
		MidiMessageBase: MidiMessageBase{
			Timestamp: timestamp,
			PortID:    portID,
			Channel:   channel,
		},
		Program: program,
	}
}

// Type implements the events.Event interface
func (e MidiProgramChangeEvent) Type() string {
	return MidiProgramChangeKey
}

/*
╭───────────────────╮
│MidiAfterTouchEvent│
╰───────────────────╯
*/

// MidiAfterTouchEvent is published when a MIDI Channel Aftertouch message is received.
type MidiAfterTouchEvent struct {
	MidiMessageBase
	// Pressure is the pressure value
	Pressure uint8
}

// NewMidiAfterTouchEvent creates a MidiAfterTouchEvent with the given parameters.
func NewMidiAfterTouchEvent(timestamp int32, portID string, channel, pressure uint8) MidiAfterTouchEvent {
	return MidiAfterTouchEvent{
		MidiMessageBase: MidiMessageBase{
			Timestamp: timestamp,
			PortID:    portID,
			Channel:   channel,
		},
		Pressure: pressure,
	}
}

// Type implements the events.Event interface
func (e MidiAfterTouchEvent) Type() string {
	return MidiAfterTouchKey
}

/*
╭───────────────────────╮
│MidiPolyAfterTouchEvent│
╰───────────────────────╯
*/

// MidiPolyAfterTouchEvent is published when a MIDI Polyphonic Aftertouch message is received.
type MidiPolyAfterTouchEvent struct {
	MidiMessageBase
	// Key is the note key
	Key uint8
	// Pressure is the pressure value
	Pressure uint8
}

// NewMidiPolyAfterTouchEvent creates a MidiPolyAfterTouchEvent with the given parameters.
func NewMidiPolyAfterTouchEvent(timestamp int32, portID string, channel, key, pressure uint8) MidiPolyAfterTouchEvent {
	return MidiPolyAfterTouchEvent{
		MidiMessageBase: MidiMessageBase{
			Timestamp: timestamp,
			PortID:    portID,
			Channel:   channel,
		},
		Key:      key,
		Pressure: pressure,
	}
}

// Type implements the events.Event interface
func (e MidiPolyAfterTouchEvent) Type() string {
	return MidiPolyAfterTouchKey
}
