package audio

import (
	"math"

	"github.com/gopxl/beep/v2"
)

// VolumeStreamer wraps a streamer and applies volume control
type VolumeStreamer struct {
	streamer beep.Streamer
	manager  *AudioManager
}

func NewVolumeStreamer(s beep.Streamer, manager *AudioManager) *VolumeStreamer {
	return &VolumeStreamer{
		streamer: s,
		manager:  manager,
	}
}

// Stream applies volume scaling to the audio samples
func (v *VolumeStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = v.streamer.Stream(samples)

	volume := v.manager.volume

	var gain float64
	if volume <= 0 {
		// Complete silence
		gain = 0
	} else if volume >= 1 {
		gain = 1
	} else {
		// Logarithmic scale: gain = 10^((volume-1) * 60 / 20)
		// Maps 0.0 -> -60dB (near silence), 1.0 -> 0dB (no change)
		dbValue := (volume - 1.0) * 60.0
		gain = math.Pow(10, dbValue/20.0)
	}

	// Apply gain to all samples
	for i := range samples[:n] {
		samples[i][0] *= gain
		samples[i][1] *= gain
	}

	return n, ok
}

// Err returns any error from the wrapped streamer
func (v *VolumeStreamer) Err() error {
	return v.streamer.Err()
}
