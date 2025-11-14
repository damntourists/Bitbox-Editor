package audio

import (
	"math/cmplx"
	"time"

	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
)

func startFFTProcessor(ab *AudioBuffer, fftChan chan<- []float64, stopChan <-chan struct{}) {
	// Pre-allocate the buffer for audio chunk
	audioChunk := make([]float64, DefaultChunkSize)

	// Pre-compute Hann window for FFT. This will smooth out the edges of the audio chunk.
	hannWindow := window.Hann(DefaultChunkSize)

	// Ticker to run at ~60 FPS
	ticker := time.NewTicker(time.Second / 60)

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			// Read the latest data from the shared buffer
			ab.Read(audioChunk)

			// Apply Hann window
			for i := 0; i < DefaultChunkSize; i++ {
				audioChunk[i] *= hannWindow[i]
			}

			// Calculate FFT
			fftResult := fft.FFTReal(audioChunk)

			// Calculate magnitudes
			magnitudes := make([]float64, DefaultChunkSize/2)
			for i := 0; i < DefaultChunkSize/2; i++ {
				magnitudes[i] = cmplx.Abs(fftResult[i])
			}

			// Send the final data to the UI thread
			fftChan <- magnitudes
		}
	}
}
