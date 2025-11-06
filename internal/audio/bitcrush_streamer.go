package audio

import (
	"math"

	"github.com/gopxl/beep/v2"
)

/*
TODO: Add once effects processing is implemented

// Defaults
crushBits := 6
crushDownsample := 4
crushMix := 0.5

// Bitcrush streamer
bitcrush := NewBitcrushStreamer(inputStream, crushBits, crushDownsample, crushMix)

speaker.Play(bitcrush)
*/

// BitcrushStreamer - A Bitcrusher effect
type BitcrushStreamer struct {
	streamer beep.Streamer
	// Number of bits to reduce to (1-16)
	bits int
	// downsample factor (1 = no downsampling)
	downsample int
	// mix (0=dry, 1=wet)
	mix float64
	// Quantization levels
	qLevels       float64
	sampleCounter int
	heldSample    [2]float64
}

// NewBitcrushStreamer creates a new bitcrush streamer.
func NewBitcrushStreamer(s beep.Streamer, bits, downsample int, mix float64) *BitcrushStreamer {
	if bits < 1 {
		bits = 1
	}
	if bits > 16 {
		bits = 16
	}
	if downsample < 1 {
		downsample = 1
	}

	// Calculate quantization levels.
	qLevels := float64(int(1) << (bits - 1))

	return &BitcrushStreamer{
		streamer:      s,
		bits:          bits,
		downsample:    downsample,
		mix:           mix,
		qLevels:       qLevels,
		sampleCounter: 0,
		heldSample:    [2]float64{0, 0},
	}
}

// Stream processes the audio by applying quantization and downsampling
func (b *BitcrushStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = b.streamer.Stream(samples)
	if !ok {
		return n, false
	}

	for i := range samples[:n] {
		drySample := samples[i]

		// Bit-depth Reduction (Quantization)
		quantizedSample := [2]float64{
			math.Floor(drySample[0]*b.qLevels) / b.qLevels,
			math.Floor(drySample[1]*b.qLevels) / b.qLevels,
		}

		// Downsampling (Sample and Hold)
		if b.sampleCounter%b.downsample == 0 {
			b.heldSample = quantizedSample
		}

		// The final wet sample is always the held sample.
		wetSample := b.heldSample

		b.sampleCounter++

		// Mix the original dry signal with the processed wet signal
		mixedSample := [2]float64{
			(drySample[0] * (1 - b.mix)) + (wetSample[0] * b.mix),
			(drySample[1] * (1 - b.mix)) + (wetSample[1] * b.mix),
		}

		samples[i] = mixedSample
	}

	return n, true
}

// Err propagates errors from the underlying streamer.
func (b *BitcrushStreamer) Err() error {
	return b.streamer.Err()
}
