package events

// Audio Load Event Keys
const (
	AudioMetadataLoadedKey = "audio.load.metadata"
	AudioSamplesLoadedKey  = "audio.load.samples"
	AudioLoadFailedKey     = "audio.load.failed"
)

/*
╭─────────────╮
│AudioLoadBase│
╰─────────────╯
*/

// AudioLoadBase contains common fields shared across audio load events.
type AudioLoadBase struct {
	// Path is the file path for the audio file
	Path string
}

// GetPath returns the file path for this audio load event.
func (b AudioLoadBase) GetPath() string {
	return b.Path
}

/*
╭────────────────────────╮
│AudioMetadataLoadedEvent│
╰────────────────────────╯
*/

// AudioMetadataLoadedEvent is published when audio file metadata has been loaded.
type AudioMetadataLoadedEvent struct {
	AudioLoadBase
}

// NewAudioMetadataLoadedEvent creates an AudioMetadataLoadedEvent with the given path.
func NewAudioMetadataLoadedEvent(path string) AudioMetadataLoadedEvent {
	return AudioMetadataLoadedEvent{
		AudioLoadBase: AudioLoadBase{Path: path},
	}
}

// Type implements the events.Event interface
func (e AudioMetadataLoadedEvent) Type() string {
	return AudioMetadataLoadedKey
}

/*
╭───────────────────────╮
│AudioSamplesLoadedEvent│
╰───────────────────────╯
*/

// AudioSamplesLoadedEvent is published when audio sample data has been loaded.
type AudioSamplesLoadedEvent struct {
	AudioLoadBase
}

// NewAudioSamplesLoadedEvent creates an AudioSamplesLoadedEvent with the given path.
func NewAudioSamplesLoadedEvent(path string) AudioSamplesLoadedEvent {
	return AudioSamplesLoadedEvent{
		AudioLoadBase: AudioLoadBase{Path: path},
	}
}

// Type implements the events.Event interface
func (e AudioSamplesLoadedEvent) Type() string {
	return AudioSamplesLoadedKey
}

/*
╭────────────────────╮
│AudioLoadFailedEvent│
╰────────────────────╯
*/

// AudioLoadFailedEvent is published when audio loading fails.
type AudioLoadFailedEvent struct {
	AudioLoadBase
	// Error is the cause of failure
	Error error
}

// NewAudioLoadFailedEvent creates an AudioLoadFailedEvent with the given path and error.
func NewAudioLoadFailedEvent(path string, err error) AudioLoadFailedEvent {
	return AudioLoadFailedEvent{
		AudioLoadBase: AudioLoadBase{Path: path},
		Error:         err,
	}
}

// Type implements the events.Event interface
func (e AudioLoadFailedEvent) Type() string {
	return AudioLoadFailedKey
}
