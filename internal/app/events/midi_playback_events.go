package events

type MidiPlaybackEvent int32

// MIDI Playback Event Enums
const (
	MidiPlaybackStartedEvent MidiPlaybackEvent = iota
	MidiPlaybackStoppedEvent
	MidiPlaybackNoteOnEvent
	MidiPlaybackNoteOffEvent
	MidiPlaybackCCEvent
)

// MIDI Playback Event Keys
const (
	MidiPlaybackStartedKey = "midi.playback.started"
	MidiPlaybackStoppedKey = "midi.playback.stopped"
	MidiPlaybackNoteOnKey  = "midi.playback.noteon"
	MidiPlaybackNoteOffKey = "midi.playback.noteoff"
	MidiPlaybackCCKey      = "midi.playback.cc"
)

// MidiPlaybackEventRecord holds data for MIDI playback events
type MidiPlaybackEventRecord struct {
	// EventType is the enum value (e.g., MidiPlaybackStartedEvent)
	EventType MidiPlaybackEvent
	// Note is the MIDI note number
	Note int
	// Velocity is the MIDI velocity
	Velocity int
	// CC is the MIDI CC number
	CC int
	// CCValue is the MIDI CC value
	CCValue int
	// Channel is the MIDI channel
	Channel int
	// OwnerID identifies which window/component initiated this MIDI playback
	OwnerID string
}

// Type implements the events.Event interface
func (e MidiPlaybackEventRecord) Type() string {
	switch e.EventType {
	case MidiPlaybackStartedEvent:
		return MidiPlaybackStartedKey
	case MidiPlaybackStoppedEvent:
		return MidiPlaybackStoppedKey
	case MidiPlaybackNoteOnEvent:
		return MidiPlaybackNoteOnKey
	case MidiPlaybackNoteOffEvent:
		return MidiPlaybackNoteOffKey
	case MidiPlaybackCCEvent:
		return MidiPlaybackCCKey
	default:
		return "midi.playback.unknown"
	}
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e MidiPlaybackEventRecord) GetOwnerID() string {
	return e.OwnerID
}
