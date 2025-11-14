package audio

import (
	"errors"

	"go.uber.org/zap"
)

var (
	ErrInvalidBounds = errors.New("invalid playback bounds")
)

// RepeatMode defines playback repeat behavior
type RepeatMode int

const (
	RepeatModeOff   RepeatMode = 0
	RepeatModeAll   RepeatMode = 1
	RepeatModeSlice RepeatMode = 2
)

// PlaybackState is a centralized container for all playback-related state
type PlaybackState struct {
	Path      string
	IsPlaying bool
	IsPaused  bool

	BoundsStart int
	BoundsEnd   int

	RepeatMode     RepeatMode
	SliceIdx       int
	SlicePositions []float64
	SamplesPerBin  float64

	CursorPosition int

	OwnerID string
}

func NewPlaybackState(path string, boundsStart, boundsEnd int) *PlaybackState {
	return &PlaybackState{
		Path:           path,
		BoundsStart:    boundsStart,
		BoundsEnd:      boundsEnd,
		RepeatMode:     RepeatModeOff,
		SliceIdx:       0,
		SlicePositions: []float64{},
		CursorPosition: boundsStart,
		IsPlaying:      false,
		IsPaused:       false,
	}
}

// Copy creates a deep copy of the playback state
func (ps *PlaybackState) Copy() *PlaybackState {
	slicesCopy := make([]float64, len(ps.SlicePositions))
	copy(slicesCopy, ps.SlicePositions)

	return &PlaybackState{
		Path:           ps.Path,
		IsPlaying:      ps.IsPlaying,
		IsPaused:       ps.IsPaused,
		BoundsStart:    ps.BoundsStart,
		BoundsEnd:      ps.BoundsEnd,
		RepeatMode:     ps.RepeatMode,
		SliceIdx:       ps.SliceIdx,
		SlicePositions: slicesCopy,
		SamplesPerBin:  ps.SamplesPerBin,
		CursorPosition: ps.CursorPosition,
		OwnerID:        ps.OwnerID,
	}
}

// GetPlaybackRegion computes the active playback region based on repeat mode
func (ps *PlaybackState) GetPlaybackRegion() (start, end int, loop bool) {
	switch ps.RepeatMode {
	case RepeatModeOff:
		// No repeat - play bounds once
		return ps.BoundsStart, ps.BoundsEnd, false

	case RepeatModeAll:
		// Repeat entire bounds
		return ps.BoundsStart, ps.BoundsEnd, true

	case RepeatModeSlice:
		// Repeat specific slice/region
		if len(ps.SlicePositions) == 0 {
			// No slices defined - loop the full file as the only region
			return ps.BoundsStart, ps.BoundsEnd, true
		}

		start, end := ps.calculateSliceRegion()
		return start, end, true

	default:
		return ps.BoundsStart, ps.BoundsEnd, false
	}
}

// calculateSliceRegion computes the start/end samples for the current slice
func (ps *PlaybackState) calculateSliceRegion() (start, end int) {
	if ps.SamplesPerBin == 0 {
		log.Warn("calculateSliceRegion called with SamplesPerBin=0, returning full bounds")
		return ps.BoundsStart, ps.BoundsEnd
	}

	if len(ps.SlicePositions) == 0 {
		return ps.BoundsStart, ps.BoundsEnd
	}

	if ps.SliceIdx == 0 {
		// First slice: from bounds start to first marker
		start = ps.BoundsStart
		end = int(ps.SlicePositions[0] * ps.SamplesPerBin)
	} else if ps.SliceIdx <= len(ps.SlicePositions) {
		// Middle or last slice
		start = int(ps.SlicePositions[ps.SliceIdx-1] * ps.SamplesPerBin)
		if ps.SliceIdx < len(ps.SlicePositions) {
			end = int(ps.SlicePositions[ps.SliceIdx] * ps.SamplesPerBin)
		} else {
			end = ps.BoundsEnd
		}
	} else {
		// Invalid slice index - return last slice
		start = int(ps.SlicePositions[len(ps.SlicePositions)-1] * ps.SamplesPerBin)
		end = ps.BoundsEnd
	}

	return start, end
}

// DetermineSliceAtPosition finds which slice contains the given sample position
func (ps *PlaybackState) DetermineSliceAtPosition(samplePosition int) int {
	if len(ps.SlicePositions) == 0 || ps.SamplesPerBin == 0 {
		return 0
	}

	sliceIdx := 0
	for i := 0; i < len(ps.SlicePositions); i++ {
		sliceBoundary := int(ps.SlicePositions[i] * ps.SamplesPerBin)
		if samplePosition >= sliceBoundary {
			sliceIdx = i + 1
		} else {
			break
		}
	}

	return sliceIdx
}

// UpdateSliceIdx updates the slice index based on current cursor position
func (ps *PlaybackState) UpdateSliceIdx() {
	ps.SliceIdx = ps.DetermineSliceAtPosition(ps.CursorPosition)
}

// ClampCursorToRegion ensures cursor is within the active playback region
func (ps *PlaybackState) ClampCursorToRegion() {
	start, end, _ := ps.GetPlaybackRegion()

	if ps.CursorPosition < start {
		ps.CursorPosition = start
	} else if ps.CursorPosition >= end {
		ps.CursorPosition = start
	}
}

// Validate checks if the playback state is valid
func (ps *PlaybackState) Validate() error {
	if ps.BoundsEnd <= ps.BoundsStart {
		log.Error("Invalid bounds: end <= start",
			zap.Int("boundsStart", ps.BoundsStart),
			zap.Int("boundsEnd", ps.BoundsEnd))
		return ErrInvalidBounds
	}

	if ps.CursorPosition < 0 {
		log.Warn("Negative cursor position, resetting to bounds start",
			zap.Int("cursorPosition", ps.CursorPosition))
		ps.CursorPosition = ps.BoundsStart
	}

	return nil
}

// SetRepeatMode changes the repeat mode
func (ps *PlaybackState) SetRepeatMode(mode RepeatMode) {
	oldMode := ps.RepeatMode
	ps.RepeatMode = mode

	// When switching to slice mode, determine which slice we're in
	if mode == RepeatModeSlice && oldMode != RepeatModeSlice {
		ps.UpdateSliceIdx()
		ps.ClampCursorToRegion()
	}

	// When switching away from slice mode, reset to bounds
	if mode != RepeatModeSlice && oldMode == RepeatModeSlice {
		ps.SliceIdx = 0
	}
}

// Stop transitions to stopped state and preserves cursor position appropriately
func (ps *PlaybackState) Stop() {
	ps.IsPlaying = false
	ps.IsPaused = false

	if ps.RepeatMode == RepeatModeSlice {
		start, _, _ := ps.GetPlaybackRegion()
		ps.CursorPosition = start
	}
}

// Play transitions to playing state
func (ps *PlaybackState) Play() {
	ps.IsPlaying = true
	ps.IsPaused = false

	// Ensure cursor is valid for current repeat mode
	ps.ClampCursorToRegion()
}

// Pause transitions to paused state
func (ps *PlaybackState) Pause(currentPosition int) {
	ps.IsPlaying = false
	ps.IsPaused = true
	ps.CursorPosition = currentPosition
}

// NavigateToSlice changes the current slice and moves cursor to slice start
func (ps *PlaybackState) NavigateToSlice(sliceIdx int) error {
	// Validate slice index
	maxSliceIdx := len(ps.SlicePositions)
	if sliceIdx < 0 || sliceIdx > maxSliceIdx {
		return errors.New("invalid slice index")
	}

	ps.SliceIdx = sliceIdx

	// Move cursor to the start of the new slice
	start, _, _ := ps.GetPlaybackRegion()
	ps.CursorPosition = start

	return nil
}

// UpdateBoundsAndSlices updates bounds and slice positions, adjusting cursor if needed
func (ps *PlaybackState) UpdateBoundsAndSlices(boundsStart, boundsEnd int, slicePositions []float64, samplesPerBin float64) {
	ps.BoundsStart = boundsStart
	ps.BoundsEnd = boundsEnd
	ps.SlicePositions = slicePositions
	ps.SamplesPerBin = samplesPerBin

	// Ensure cursor and slice index are still valid
	if ps.RepeatMode == RepeatModeSlice {
		if ps.SliceIdx > len(ps.SlicePositions) {
			ps.SliceIdx = len(ps.SlicePositions)
		}
	}

	ps.ClampCursorToRegion()
}

// GetCursorProgress returns cursor position as a progress value (0.0 to 1.0) relative to bounds
func (ps *PlaybackState) GetCursorProgress() float64 {
	boundsRange := float64(ps.BoundsEnd - ps.BoundsStart)
	if boundsRange <= 0 {
		return 0.0
	}

	relativePos := float64(ps.CursorPosition - ps.BoundsStart)
	progress := relativePos / boundsRange

	// Clamp to 0-1
	if progress < 0 {
		return 0.0
	}
	if progress > 1.0 {
		return 1.0
	}

	return progress
}

// SetCursorFromProgress sets cursor position from a progress value (0.0 to 1.0) relative to bounds
func (ps *PlaybackState) SetCursorFromProgress(progress float64) {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	boundsRange := float64(ps.BoundsEnd - ps.BoundsStart)
	ps.CursorPosition = ps.BoundsStart + int(progress*boundsRange)
}
