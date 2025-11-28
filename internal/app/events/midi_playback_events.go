package events

// MIDI Playback Event Keys
const (
	MidiPlaybackStartedKey = "midi.playback.started"
	MidiPlaybackStoppedKey = "midi.playback.stopped"
	MidiPlaybackNoteOnKey  = "midi.playback.noteon"
	MidiPlaybackNoteOffKey = "midi.playback.noteoff"
	MidiPlaybackCCKey      = "midi.playback.cc"
)

/*
╭────────────────╮
│MidiPlaybackBase│
╰────────────────╯
*/

// MidiPlaybackBase contains common fields shared across MIDI playback events.
type MidiPlaybackBase struct {
	// OwnerID identifies which window/component initiated this MIDI playback
	OwnerID string
}

/*
╭───────────────────────╮
│MidiPlaybackMessageBase│
╰───────────────────────╯
*/

// MidiPlaybackMessageBase contains common fields for MIDI message playback events.
type MidiPlaybackMessageBase struct {
	MidiPlaybackBase
	// Channel is the MIDI channel
	Channel int
}

/*
╭────────────────────────╮
│MidiPlaybackStartedEvent│
╰────────────────────────╯
*/

// MidiPlaybackStartedEvent is published when MIDI playback starts.
type MidiPlaybackStartedEvent struct {
	MidiPlaybackBase
}

// NewMidiPlaybackStartedEvent creates a MidiPlaybackStartedEvent with the given owner ID.
func NewMidiPlaybackStartedEvent(ownerID string) MidiPlaybackStartedEvent {
	return MidiPlaybackStartedEvent{
		MidiPlaybackBase: MidiPlaybackBase{OwnerID: ownerID},
	}
}

// Type implements the events.Event interface
func (e MidiPlaybackStartedEvent) Type() string {
	return MidiPlaybackStartedKey
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e MidiPlaybackStartedEvent) GetOwnerID() string {
	return e.OwnerID
}

/*
╭────────────────────────╮
│MidiPlaybackStoppedEvent│
╰────────────────────────╯
*/

// MidiPlaybackStoppedEvent is published when MIDI playback stops.
type MidiPlaybackStoppedEvent struct {
	MidiPlaybackBase
}

// NewMidiPlaybackStoppedEvent creates a MidiPlaybackStoppedEvent with the given owner ID.
func NewMidiPlaybackStoppedEvent(ownerID string) MidiPlaybackStoppedEvent {
	return MidiPlaybackStoppedEvent{
		MidiPlaybackBase: MidiPlaybackBase{OwnerID: ownerID},
	}
}

// Type implements the events.Event interface
func (e MidiPlaybackStoppedEvent) Type() string {
	return MidiPlaybackStoppedKey
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e MidiPlaybackStoppedEvent) GetOwnerID() string {
	return e.OwnerID
}

/*
╭───────────────────────╮
│MidiPlaybackNoteOnEvent│
╰───────────────────────╯
*/

// MidiPlaybackNoteOnEvent is published when a MIDI note is played.
type MidiPlaybackNoteOnEvent struct {
	MidiPlaybackMessageBase
	// Note is the MIDI note number
	Note int
	// Velocity is the MIDI velocity
	Velocity int
}

// NewMidiPlaybackNoteOnEvent creates a MidiPlaybackNoteOnEvent with the given parameters.
func NewMidiPlaybackNoteOnEvent(ownerID string, channel, note, velocity int) MidiPlaybackNoteOnEvent {
	return MidiPlaybackNoteOnEvent{
		MidiPlaybackMessageBase: MidiPlaybackMessageBase{
			MidiPlaybackBase: MidiPlaybackBase{OwnerID: ownerID},
			Channel:          channel,
		},
		Note:     note,
		Velocity: velocity,
	}
}

// Type implements the events.Event interface
func (e MidiPlaybackNoteOnEvent) Type() string {
	return MidiPlaybackNoteOnKey
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e MidiPlaybackNoteOnEvent) GetOwnerID() string {
	return e.OwnerID
}

/*
╭────────────────────────╮
│MidiPlaybackNoteOffEvent│
╰────────────────────────╯
*/

// MidiPlaybackNoteOffEvent is published when a MIDI note is released.
type MidiPlaybackNoteOffEvent struct {
	MidiPlaybackMessageBase
	// Note is the MIDI note number
	Note int
	// Velocity is the MIDI velocity
	Velocity int
}

// NewMidiPlaybackNoteOffEvent creates a MidiPlaybackNoteOffEvent with the given parameters.
func NewMidiPlaybackNoteOffEvent(ownerID string, channel, note, velocity int) MidiPlaybackNoteOffEvent {
	return MidiPlaybackNoteOffEvent{
		MidiPlaybackMessageBase: MidiPlaybackMessageBase{
			MidiPlaybackBase: MidiPlaybackBase{OwnerID: ownerID},
			Channel:          channel,
		},
		Note:     note,
		Velocity: velocity,
	}
}

// Type implements the events.Event interface
func (e MidiPlaybackNoteOffEvent) Type() string {
	return MidiPlaybackNoteOffKey
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e MidiPlaybackNoteOffEvent) GetOwnerID() string {
	return e.OwnerID
}

/*
╭───────────────────╮
│MidiPlaybackCCEvent│
╰───────────────────╯
*/

// MidiPlaybackCCEvent is published when a MIDI control change is sent.
type MidiPlaybackCCEvent struct {
	MidiPlaybackMessageBase
	// CC is the MIDI CC number
	CC int
	// Value is the MIDI CC value
	Value int
}

// NewMidiPlaybackCCEvent creates a MidiPlaybackCCEvent with the given parameters.
func NewMidiPlaybackCCEvent(ownerID string, channel, cc, value int) MidiPlaybackCCEvent {
	return MidiPlaybackCCEvent{
		MidiPlaybackMessageBase: MidiPlaybackMessageBase{
			MidiPlaybackBase: MidiPlaybackBase{OwnerID: ownerID},
			Channel:          channel,
		},
		CC:    cc,
		Value: value,
	}
}

// Type implements the events.Event interface
func (e MidiPlaybackCCEvent) Type() string {
	return MidiPlaybackCCKey
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e MidiPlaybackCCEvent) GetOwnerID() string {
	return e.OwnerID
}
