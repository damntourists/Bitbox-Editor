package events

// Audio Volume Event Keys
const (
	AudioVolumeChangedKey = "audio.volume.changed"
)

// AudioVolumeChangedEvent is published when the audio volume changes.
type AudioVolumeChangedEvent struct {
	// Volume is the new volume level, normalized between 0.0 (silent) and 1.0 (full).
	Volume float64
}

func (e AudioVolumeChangedEvent) Type() string {
	return AudioVolumeChangedKey
}
