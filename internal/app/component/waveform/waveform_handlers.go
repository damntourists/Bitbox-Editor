package waveform

import (
	"bitbox-editor/internal/audio"
	"bitbox-editor/internal/logging"

	"github.com/AllenDang/cimgui-go/implot"
	"go.uber.org/zap"
)

func (wc *WaveComponent) onSetWaveDisplayData(data audio.WaveDisplayData) {
	wc.displayData = data

	// Recalculate samplesPerBin based on new data
	if wc.displayData.NumSamples > 0 && wc.displayData.XLimitMax > 0 {
		wc.samplesPerBin = float64(wc.displayData.NumSamples) / (wc.displayData.XLimitMax + 1)
	} else {
		wc.samplesPerBin = 1.0
	}

	// Only initialize bounds on first load or when explicitly cleared
	if !wc.boundsInitialized && wc.displayData.XLimitMax > 0 {
		wc.boundsStart = 0.0
		wc.boundsEnd = wc.displayData.XLimitMax
		log.Debug("Initialized bounds to full range",
			zap.String("id", wc.IDStr()),
			zap.Float64("boundsEnd", wc.boundsEnd))

		if wc.boundsMarker == nil {
			wc.boundsMarker = NewWaveBoundsMarker()
		}

		wc.boundsInitialized = true
	}
}

func (wc *WaveComponent) onUpdatePlaybackProgress(update PlaybackProgressUpdate) {
	wc.displayData.IsPlaying = update.IsPlaying
	wc.displayData.Progress = update.Progress
	wc.displayData.PositionSeconds = update.PositionSeconds
}

func (wc *WaveComponent) onSetWaveSlices(slices []*WaveMarker) {
	wc.slices = slices
}

func (wc *WaveComponent) onSetWaveBounds(payload WaveBoundsPayload) {
	wc.boundsStart = payload.Start
	wc.boundsEnd = payload.End

	// Adjust cursor position if it's now outside the new bounds
	if wc.cursor != nil {
		cursorPos := wc.cursor.position
		if cursorPos < payload.Start || cursorPos > payload.End {
			if cursorPos < payload.Start {
				wc.cursor.SetPositionImmediate(payload.Start)
			} else if cursorPos > payload.End {
				wc.cursor.SetPositionImmediate(payload.End)
			}
			log.Debug("Cursor position adjusted due to bounds change",
				zap.Float64("oldCursorPos", cursorPos),
				zap.Float64("newCursorPos", wc.cursor.position))
		}
	}

	// Adjust slice markers to stay within new bounds
	minGap := 1.0
	newPositions := make([]float64, len(wc.slices))
	for i, slice := range wc.slices {
		if slice == nil {
			newPositions[i] = -1
			continue
		}

		newPos := slice.start
		if newPos < wc.boundsStart {
			newPos = wc.boundsStart
		}

		if newPos > wc.boundsEnd {
			newPos = wc.boundsEnd
		}

		newPositions[i] = newPos
	}

	// Second pass: adjust positions to maintain gaps. Left to right.
	for i := 0; i < len(wc.slices); i++ {
		if wc.slices[i] == nil || newPositions[i] < 0 {
			continue
		}

		if i > 0 && wc.slices[i-1] != nil && newPositions[i-1] >= 0 {
			if newPositions[i] < newPositions[i-1]+minGap {
				newPositions[i] = newPositions[i-1] + minGap
			}
		}

		if newPositions[i] > wc.boundsEnd {
			newPositions[i] = wc.boundsEnd
		}
	}

	// Third pass: right to left to push left if needed
	for i := len(wc.slices) - 1; i >= 0; i-- {
		if wc.slices[i] == nil || newPositions[i] < 0 {
			continue
		}

		if i < len(wc.slices)-1 && wc.slices[i+1] != nil && newPositions[i+1] >= 0 {
			if newPositions[i] > newPositions[i+1]-minGap {
				newPositions[i] = newPositions[i+1] - minGap
			}
		}

		if newPositions[i] < wc.boundsStart {
			newPositions[i] = wc.boundsStart
		}
	}

	// Apply the new positions
	for i, slice := range wc.slices {
		if slice != nil && newPositions[i] >= 0 {
			slice.start = newPositions[i]
		}
	}
}

func (wc *WaveComponent) onSetWaveBoundsFromSamples(payload WaveBoundsSamplesPayload) {
	if wc.displayData.SampleRate == 0 {
		logging.NewLogger("waveform").Warn("Cannot set bounds from samples: sample rate is 0")
		return
	}

	// Convert sample positions to display positions (seconds)
	startSeconds := float64(payload.StartSample) / float64(wc.displayData.SampleRate)
	endSeconds := float64(payload.EndSample) / float64(wc.displayData.SampleRate)

	// Use the existing onSetWaveBounds handler
	wc.onSetWaveBounds(WaveBoundsPayload{
		Start: startSeconds,
		End:   endSeconds,
	})
}

func (wc *WaveComponent) onSetWaveCursor(position float64) {
	if wc.cursor != nil {
		wc.cursor.SetPositionImmediate(position)
	}
}

func (wc *WaveComponent) onSetWavePlotFlags(flags implot.Flags) {
	wc.plotFlags = flags
}

func (wc *WaveComponent) onSetWaveAxisXFlags(flags implot.AxisFlags) {
	wc.axisXFlags = flags
}

func (wc *WaveComponent) onSetWaveAxisYFlags(flags implot.AxisFlags) {
	wc.axisYFlags = flags
}

func (wc *WaveComponent) onAddWaveSlice(position float64) {
	marker := NewWaveMarker(position)
	wc.slices = append(wc.slices, marker)
}

func (wc *WaveComponent) onUpdateWaveSlicePosition(payload WaveSlicePositionPayload) {
	if payload.Index >= 0 && payload.Index < len(wc.slices) && wc.slices[payload.Index] != nil {
		wc.slices[payload.Index].start = payload.NewStart
	}
}
