package events

type AudioPlaybackEvent int32

// Audio Playback Event Enums
const (
	AudioPlaybackStartedEvent AudioPlaybackEvent = iota
	AudioPlaybackProgressEvent
	AudioPlaybackStoppedEvent
	AudioPlaybackPausedEvent
	AudioPlaybackResumedEvent
	AudioPlaybackFinishedEvent
)

// Audio Playback Event Keys
const (
	AudioPlaybackStartedKey  = "audio.playback.started"
	AudioPlaybackProgressKey = "audio.playback.progress"
	AudioPlaybackStoppedKey  = "audio.playback.stopped"
	AudioPlaybackPausedKey   = "audio.playback.paused"
	AudioPlaybackResumedKey  = "audio.playback.resumed"
	AudioPlaybackFinishedKey = "audio.playback.finished"
)

// AudioPlaybackEventRecord holds data for playback-related events, such as
// start, stop, progress, and pause.
type AudioPlaybackEventRecord struct {
	// EventType is the enum value (e.g., AudioPlaybackStartedEvent).
	EventType AudioPlaybackEvent
	// Path is the file path of the audio stream.
	Path string
	// Progress is the normalized playback progress (0.0 to 1.0).
	Progress float64
	// PositionSamples is the current playback position in audio samples.
	PositionSamples int
	// DurationSamples is the total duration of the audio stream in samples.
	DurationSamples int
	// IsPlaying is true if the audio is actively playing.
	IsPlaying bool
	// IsPaused is true if the audio is paused.
	IsPaused bool
	// WaveID is a unique identifier for the specific playback instance.
	WaveID uintptr
	// LoopEnabled is true if the playback was set to loop.
	LoopEnabled bool
	// OwnerID identifies which window/component initiated this playback.
	OwnerID string
}

// Type implements the events.Event interface
func (e AudioPlaybackEventRecord) Type() string {
	switch e.EventType {
	case AudioPlaybackStartedEvent:
		return AudioPlaybackStartedKey
	case AudioPlaybackProgressEvent:
		return AudioPlaybackProgressKey
	case AudioPlaybackStoppedEvent:
		return AudioPlaybackStoppedKey
	case AudioPlaybackPausedEvent:
		return AudioPlaybackPausedKey
	case AudioPlaybackResumedEvent:
		return AudioPlaybackResumedKey
	case AudioPlaybackFinishedEvent:
		return AudioPlaybackFinishedKey
	default:
		return "audio.playback.unknown"
	}
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e AudioPlaybackEventRecord) GetOwnerID() string {
	return e.OwnerID
}
