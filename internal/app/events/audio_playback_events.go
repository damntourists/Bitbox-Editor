package events

// Audio Playback Event Keys
const (
	AudioPlaybackStartedKey  = "audio.playback.started"
	AudioPlaybackProgressKey = "audio.playback.progress"
	AudioPlaybackStoppedKey  = "audio.playback.stopped"
	AudioPlaybackPausedKey   = "audio.playback.paused"
	AudioPlaybackResumedKey  = "audio.playback.resumed"
	AudioPlaybackFinishedKey = "audio.playback.finished"
)

/*
╭─────────────────╮
│AudioPlaybackBase│
╰─────────────────╯
*/

// AudioPlaybackBase contains common fields shared across audio playback events.
type AudioPlaybackBase struct {
	// Path is the file path for the audio file
	Path string
	// OwnerID is the window/component identifier who initiated this playback
	OwnerID string
}

// GetPath returns the file path for this audio playback event.
func (b AudioPlaybackBase) GetPath() string {
	return b.Path
}

/*
╭─────────────────────────╮
│AudioPlaybackProgressBase│
╰─────────────────────────╯
*/

// AudioPlaybackProgressBase contains common fields for events that track playback progress.
type AudioPlaybackProgressBase struct {
	AudioPlaybackBase
	// Progress is the normalized playback progress (0.0 to 1.0)
	Progress float64
	// PositionSamples is current playback position in audio samples
	PositionSamples int
	// DurationSamples is the total duration of the audio stream in samples
	DurationSamples int
	// WaveID is the unique identifier for the specific playback instance
	WaveID uintptr
}

// GetProgress returns the normalized playback progress.
func (b AudioPlaybackProgressBase) GetProgress() float64 {
	return b.Progress
}

// GetPositionSamples returns the current playback position in audio samples.
func (b AudioPlaybackProgressBase) GetPositionSamples() int {
	return b.PositionSamples
}

/*
╭─────────────────────────╮
│AudioPlaybackStartedEvent│
╰─────────────────────────╯
*/

// AudioPlaybackStartedEvent is published when audio playback starts.
type AudioPlaybackStartedEvent struct {
	AudioPlaybackProgressBase
	// LoopEnabled true if loop playback is enabled
	LoopEnabled bool
}

// NewAudioPlaybackStartedEvent creates an AudioPlaybackStartedEvent with the given parameters.
func NewAudioPlaybackStartedEvent(
	path,
	ownerID string,
	progress float64,
	positionSamples,
	durationSamples int,
	loopEnabled bool,
) AudioPlaybackStartedEvent {
	return AudioPlaybackStartedEvent{
		AudioPlaybackProgressBase: AudioPlaybackProgressBase{
			AudioPlaybackBase: AudioPlaybackBase{
				Path:    path,
				OwnerID: ownerID,
			},
			Progress:        progress,
			PositionSamples: positionSamples,
			DurationSamples: durationSamples,
		},
		LoopEnabled: loopEnabled,
	}
}

// Type implements the events.Event interface
func (e AudioPlaybackStartedEvent) Type() string {
	return AudioPlaybackStartedKey
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e AudioPlaybackStartedEvent) GetOwnerID() string {
	return e.OwnerID
}

/*
╭──────────────────────────╮
│AudioPlaybackProgressEvent│
╰──────────────────────────╯
*/

// AudioPlaybackProgressEvent is published during audio playback to report progress.
type AudioPlaybackProgressEvent struct {
	AudioPlaybackProgressBase
}

// NewAudioPlaybackProgressEvent creates an AudioPlaybackProgressEvent with the given parameters.
func NewAudioPlaybackProgressEvent(
	path,
	ownerID string,
	progress float64,
	positionSamples,
	durationSamples int,
) AudioPlaybackProgressEvent {
	return AudioPlaybackProgressEvent{
		AudioPlaybackProgressBase: AudioPlaybackProgressBase{
			AudioPlaybackBase: AudioPlaybackBase{
				Path:    path,
				OwnerID: ownerID,
			},
			Progress:        progress,
			PositionSamples: positionSamples,
			DurationSamples: durationSamples,
		},
	}
}

// Type implements the events.Event interface
func (e AudioPlaybackProgressEvent) Type() string {
	return AudioPlaybackProgressKey
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e AudioPlaybackProgressEvent) GetOwnerID() string {
	return e.OwnerID
}

/*
╭─────────────────────────╮
│AudioPlaybackStoppedEvent│
╰─────────────────────────╯
*/

// AudioPlaybackStoppedEvent is published when audio playback stops.
type AudioPlaybackStoppedEvent struct {
	AudioPlaybackProgressBase
}

// NewAudioPlaybackStoppedEvent creates an AudioPlaybackStoppedEvent with the given parameters.
func NewAudioPlaybackStoppedEvent(
	path,
	ownerID string,
	progress float64,
	positionSamples,
	durationSamples int,
) AudioPlaybackStoppedEvent {
	return AudioPlaybackStoppedEvent{
		AudioPlaybackProgressBase: AudioPlaybackProgressBase{
			AudioPlaybackBase: AudioPlaybackBase{
				Path:    path,
				OwnerID: ownerID,
			},
			Progress:        progress,
			PositionSamples: positionSamples,
			DurationSamples: durationSamples,
		},
	}
}

// Type implements the events.Event interface
func (e AudioPlaybackStoppedEvent) Type() string {
	return AudioPlaybackStoppedKey
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e AudioPlaybackStoppedEvent) GetOwnerID() string {
	return e.OwnerID
}

/*
╭────────────────────────╮
│AudioPlaybackPausedEvent│
╰────────────────────────╯
*/

// AudioPlaybackPausedEvent is published when audio playback is paused.
type AudioPlaybackPausedEvent struct {
	AudioPlaybackProgressBase
}

// NewAudioPlaybackPausedEvent creates an AudioPlaybackPausedEvent with the given parameters.
func NewAudioPlaybackPausedEvent(
	path,
	ownerID string,
	progress float64,
	positionSamples,
	durationSamples int,
) AudioPlaybackPausedEvent {
	return AudioPlaybackPausedEvent{
		AudioPlaybackProgressBase: AudioPlaybackProgressBase{
			AudioPlaybackBase: AudioPlaybackBase{
				Path:    path,
				OwnerID: ownerID,
			},
			Progress:        progress,
			PositionSamples: positionSamples,
			DurationSamples: durationSamples,
		},
	}
}

// Type implements the events.Event interface
func (e AudioPlaybackPausedEvent) Type() string {
	return AudioPlaybackPausedKey
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e AudioPlaybackPausedEvent) GetOwnerID() string {
	return e.OwnerID
}

/*
╭─────────────────────────╮
│AudioPlaybackResumedEvent│
╰─────────────────────────╯
*/

// AudioPlaybackResumedEvent is published when audio playback is resumed.
type AudioPlaybackResumedEvent struct {
	AudioPlaybackProgressBase
}

// NewAudioPlaybackResumedEvent creates an AudioPlaybackResumedEvent with the given parameters.
func NewAudioPlaybackResumedEvent(
	path,
	ownerID string,
	progress float64,
	positionSamples,
	durationSamples int,
) AudioPlaybackResumedEvent {
	return AudioPlaybackResumedEvent{
		AudioPlaybackProgressBase: AudioPlaybackProgressBase{
			AudioPlaybackBase: AudioPlaybackBase{
				Path:    path,
				OwnerID: ownerID,
			},
			Progress:        progress,
			PositionSamples: positionSamples,
			DurationSamples: durationSamples,
		},
	}
}

// Type implements the events.Event interface
func (e AudioPlaybackResumedEvent) Type() string {
	return AudioPlaybackResumedKey
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e AudioPlaybackResumedEvent) GetOwnerID() string {
	return e.OwnerID
}

/*
╭──────────────────────────╮
│AudioPlaybackFinishedEvent│
╰──────────────────────────╯
*/

// AudioPlaybackFinishedEvent is published when audio playback finishes naturally.
type AudioPlaybackFinishedEvent struct {
	AudioPlaybackBase
	// LoopEnabled true if loop playback was enabled
	LoopEnabled bool
}

// NewAudioPlaybackFinishedEvent creates an AudioPlaybackFinishedEvent with the given parameters.
func NewAudioPlaybackFinishedEvent(path, ownerID string, loopEnabled bool) AudioPlaybackFinishedEvent {
	return AudioPlaybackFinishedEvent{
		AudioPlaybackBase: AudioPlaybackBase{
			Path:    path,
			OwnerID: ownerID,
		},
		LoopEnabled: loopEnabled,
	}
}

// Type implements the events.Event interface
func (e AudioPlaybackFinishedEvent) Type() string {
	return AudioPlaybackFinishedKey
}

// GetOwnerID implements the OwnedEvent interface for automatic filtering
func (e AudioPlaybackFinishedEvent) GetOwnerID() string {
	return e.OwnerID
}
