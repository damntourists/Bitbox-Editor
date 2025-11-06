package audio

import (
	"sync"
)

// AudioBuffer is a thread-safe buffer for audio data.
type AudioBuffer struct {
	mu     sync.Mutex
	buffer []float64
}

// Write appends audio data to the buffer.
func (ab *AudioBuffer) Write(data []float64) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	copy(ab.buffer, data)
}

// Read retrieves audio data from the buffer.
func (ab *AudioBuffer) Read(out []float64) {
	ab.mu.Lock()
	defer ab.mu.Unlock()
	copy(out, ab.buffer)
}

// NewAudioBuffer creates a new AudioBuffer with the specified chunk size.
func NewAudioBuffer(chunkSize int) *AudioBuffer {
	return &AudioBuffer{
		buffer: make([]float64, chunkSize),
	}
}
