package audio

import "github.com/gopxl/beep/v2"

type AudioMonitor struct {
	streamer   beep.Streamer
	buffer     *AudioBuffer
	monoBuffer []float64
}

func (am *AudioMonitor) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = am.streamer.Stream(samples)
	if !ok {
		return n, false
	}

	for i := 0; i < n; i++ {
		am.monoBuffer[i] = (samples[i][0] + samples[i][1]) * 0.5
	}

	am.buffer.Write(am.monoBuffer)

	return n, ok
}

func (am *AudioMonitor) Err() error {
	return am.streamer.Err()
}

func NewAudioMonitor(source beep.Streamer, buffer *AudioBuffer) *AudioMonitor {
	return &AudioMonitor{
		streamer:   source,
		buffer:     buffer,
		monoBuffer: make([]float64, len(buffer.buffer)),
	}
}
