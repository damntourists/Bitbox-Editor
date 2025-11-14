package audio

import (
	"math"
	"sync/atomic"

	"github.com/gopxl/beep/v2"
)

type ProgressStreamer struct {
	beep.Streamer
	startSample  int
	totalSamples int
	position     atomic.Int64
	progress     atomic.Uint64
	regionPtr    atomic.Pointer[PlaybackRegion]
	audioManager *AudioManager
	wavePath     string
	isLooping    bool
}

func NewProgressStreamer(s beep.Streamer, startSample, totalSamples int) *ProgressStreamer {
	ps := &ProgressStreamer{
		Streamer:     s,
		startSample:  startSample,
		totalSamples: totalSamples,
	}

	ps.position.Store(0)
	ps.progress.Store(0)
	return ps
}

// NewProgressStreamerWithRegion creates a progress streamer with playback region support
func NewProgressStreamerWithRegion(
	s beep.Streamer,
	startSample, totalSamples int,
	region *PlaybackRegion,
	am *AudioManager,
	wavePath string,
	initialPosition int,
	isLooping bool) *ProgressStreamer {

	progressVal := float64(0)
	if totalSamples > 0 {
		progressVal = float64(initialPosition) / float64(totalSamples)
		if progressVal > 1 {
			progressVal = 1
		}
	}

	ps := &ProgressStreamer{
		Streamer:     s,
		startSample:  startSample,
		totalSamples: totalSamples,
		audioManager: am,
		wavePath:     wavePath,
		isLooping:    isLooping,
	}

	ps.position.Store(int64(initialPosition))
	ps.progress.Store(math.Float64bits(progressVal))

	if region != nil {
		regionCopy := *region
		ps.regionPtr.Store(&regionCopy)
	}

	return ps
}

func (ps *ProgressStreamer) Stream(samples [][2]float64) (int, bool) {
	n, ok := ps.Streamer.Stream(samples)

	newPosition := ps.position.Add(int64(n))

	// Calculate and store progress
	var progressVal float64
	playbackRegion := ps.regionPtr.Load()
	if ps.isLooping && playbackRegion != nil {
		regionLength := playbackRegion.EndSample - playbackRegion.StartSample
		if regionLength > 0 {
			relativePos := int(newPosition) % regionLength
			// Calculate progress within the region
			absolutePos := playbackRegion.StartSample + relativePos
			if ps.totalSamples > 0 {
				progressVal = float64(absolutePos-ps.startSample) / float64(ps.totalSamples)
			}
		}
	} else {
		// Non-looping: linear progress
		if ps.totalSamples > 0 {
			progressVal = float64(newPosition) / float64(ps.totalSamples)
			if progressVal > 1 {
				progressVal = 1
			}
		}
	}

	// Store progress as bits
	ps.progress.Store(math.Float64bits(progressVal))

	return n, ok
}

func (ps *ProgressStreamer) Progress() float64 {
	// Load progress bits and convert back to float64
	bits := ps.progress.Load()
	return math.Float64frombits(bits)
}

func (ps *ProgressStreamer) TotalSamples() int {
	// Immutable after creation, safe to read directly
	return ps.totalSamples
}

func (ps *ProgressStreamer) Position() int {
	// Load position atomically (lock-free)
	return int(ps.position.Load())
}

// AbsolutePosition returns the absolute position in the original audio file
func (ps *ProgressStreamer) AbsolutePosition() int {
	pos := int(ps.position.Load())

	// For looping regions, return position within the region
	playbackRegion := ps.regionPtr.Load()
	if ps.isLooping && playbackRegion != nil {
		regionLength := playbackRegion.EndSample - playbackRegion.StartSample
		if regionLength > 0 {
			relativePos := pos % regionLength
			return playbackRegion.StartSample + relativePos
		}
	}

	return ps.startSample + pos
}

func (ps *ProgressStreamer) Reset() {
	// Reset position and progress
	ps.position.Store(0)
	ps.progress.Store(0)
}

// UpdatePlaybackRegion updates the playback region for an active streamer
func (ps *ProgressStreamer) UpdatePlaybackRegion(region *PlaybackRegion) {
	if region != nil {
		// Create a copy
		regionCopy := *region
		ps.regionPtr.Store(&regionCopy)
	} else {
		ps.regionPtr.Store(nil)
	}
}
