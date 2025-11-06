package events

type AudioLoadEvent int32

// Audio Load Event Enums
const (
	AudioMetadataLoadedEvent AudioLoadEvent = iota
	AudioSamplesLoadedEvent
	AudioLoadFailedEvent
)

// Audio Load Event Keys
const (
	AudioMetadataLoadedKey = "audio.load.metadata"
	AudioSamplesLoadedKey  = "audio.load.samples"
	AudioLoadFailedKey     = "audio.load.failed"
)

// AudioLoadEventRecord holds data for audio file loading events.
// It reports the status of metadata and sample data loading.
type AudioLoadEventRecord struct {
	// EventType is the enum value (e.g., AudioMetadataLoadedEvent).
	EventType AudioLoadEvent
	// Path is the file path of the audio file being loaded.
	Path string
	// MetadataLoaded is true if the file's metadata (sample rate, etc.) is loaded.
	MetadataLoaded bool
	// SamplesLoaded is true if the full audio sample data is loaded.
	SamplesLoaded bool
	// Failed is true if any part of the loading process failed.
	Failed bool
	// Error contains the specific error if loading failed.
	Error error
}

// Type implements the events.Event interface
func (e AudioLoadEventRecord) Type() string {
	switch e.EventType {
	case AudioMetadataLoadedEvent:
		return AudioMetadataLoadedKey
	case AudioSamplesLoadedEvent:
		return AudioSamplesLoadedKey
	case AudioLoadFailedEvent:
		return AudioLoadFailedKey
	default:
		return "audio.load.unknown"
	}
}
