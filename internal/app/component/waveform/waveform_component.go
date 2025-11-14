package waveform

import (
	"bitbox-editor/internal/app/animation"
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/app/font"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/audio"
	"bitbox-editor/internal/logging"
	"bitbox-editor/internal/util"
	"math"
	"sort"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
	"github.com/AllenDang/cimgui-go/utils"
	"go.uber.org/zap"
)

var log = logging.NewLogger("waveform")

type WaveComponent struct {
	*component.Component[*WaveComponent]

	displayData audio.WaveDisplayData

	cursor *WaveCursor
	slices []*WaveMarker

	boundsStart       float64
	boundsEnd         float64
	boundsMarker      *WaveBoundsMarker
	boundsInitialized bool

	plotFlags  implot.Flags
	axisXFlags implot.AxisFlags
	axisYFlags implot.AxisFlags

	samplesPerBin float64

	// repeatMode 0 = off, 1 = repeat all, 2 = repeat one slice
	repeatMode     int
	repeatSliceIdx int

	emptyText string
	eventSub  chan events.Event
}

// NewWaveformComponent constructor
func NewWaveformComponent(id imgui.ID) *WaveComponent {
	cmp := &WaveComponent{
		displayData:  audio.WaveDisplayData{},
		cursor:       &WaveCursor{position: 0.0},
		slices:       make([]*WaveMarker, 0),
		boundsStart:  0.0,
		boundsEnd:    0.0,
		boundsMarker: nil,
		emptyText:    "No wav loaded . . .",
		plotFlags: implot.FlagsNoMenus |
			implot.FlagsNoLegend |
			implot.FlagsNoTitle |
			implot.FlagsNoFrame |
			implot.FlagsNoMouseText,
		axisXFlags: implot.AxisFlagsNoMenus |
			implot.AxisFlagsNoLabel |
			implot.AxisFlagsNoSideSwitch,
		axisYFlags: implot.AxisFlagsNoMenus |
			implot.AxisFlagsNoDecorations |
			implot.AxisFlagsNoHighlight |
			implot.AxisFlagsNoSideSwitch |
			implot.AxisFlagsAutoFit,
		eventSub: make(chan events.Event, 50),
	}

	cmp.Component = component.NewComponent[*WaveComponent](id, cmp.handleUpdate)
	cmp.Component.SetLayoutBuilder(cmp)

	// Subscribe to playback progress events
	bus := eventbus.Bus
	uuid := cmp.UUID()
	bus.Subscribe(events.AudioPlaybackProgressKey, uuid, cmp.eventSub)
	bus.Subscribe(events.AudioPlaybackStartedKey, uuid, cmp.eventSub)
	bus.Subscribe(events.AudioPlaybackStoppedKey, uuid, cmp.eventSub)
	bus.Subscribe(events.AudioPlaybackFinishedKey, uuid, cmp.eventSub)

	return cmp
}

// drainEvents translates global bus events into local commands
func (wc *WaveComponent) drainEvents() {
	for {
		select {
		case event := <-wc.eventSub:
			if e, ok := event.(events.AudioPlaybackEventRecord); ok {
				if wc.displayData.Path == "" || e.Path != wc.displayData.Path {
					continue
				}

				// Translate to a local command
				cmd := component.UpdateCmd{
					Type: cmdUpdatePlaybackProgress,
					Data: PlaybackProgressUpdate{
						IsPlaying:       e.IsPlaying,
						Progress:        e.Progress,
						PositionSeconds: float64(e.PositionSamples) / float64(wc.displayData.SampleRate),
					},
				}
				wc.SendUpdate(cmd)
			}
		default:
			// No more events
			return
		}
	}
}

func (wc *WaveComponent) handleUpdate(cmd component.UpdateCmd) {
	if wc.Component.HandleGlobalUpdate(cmd) {
		if cmd.Type == component.CmdSetLoading {
			wc.displayData.IsLoading = cmd.Data.(bool)
		}
		return
	}

	switch cmd.Type {
	case cmdSetWaveDisplayData:
		if data, ok := cmd.Data.(audio.WaveDisplayData); ok {
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

	case cmdUpdatePlaybackProgress:
		if update, ok := cmd.Data.(PlaybackProgressUpdate); ok {
			wc.displayData.IsPlaying = update.IsPlaying
			wc.displayData.Progress = update.Progress
			wc.displayData.PositionSeconds = update.PositionSeconds
		}

	case cmdSetWaveSlices:
		if s, ok := cmd.Data.([]*WaveMarker); ok {
			wc.slices = s
		} else if cmd.Data == nil {
			wc.slices = nil
		}

	case cmdSetWaveBounds:
		if payload, ok := cmd.Data.(WaveBoundsPayload); ok {
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
					minPos := newPositions[i-1] + minGap
					if newPositions[i] < minPos {
						newPositions[i] = minPos
						if newPositions[i] > wc.boundsEnd {
							newPositions[i] = -1
						}
					}
				}
			}

			// Third pass: apply the new positions
			for i, slice := range wc.slices {
				if slice == nil || newPositions[i] < 0 {
					continue
				}

				if newPositions[i] != slice.start {
					slice.start = newPositions[i]
				}
			}

			// Remove slices that are out of bounds
			validSlices := make([]*WaveMarker, 0, len(wc.slices))
			for i, slice := range wc.slices {
				if slice == nil {
					continue
				}

				if slice.start < wc.boundsStart || slice.start > wc.boundsEnd {
					log.Warn("Removing slice marker outside bounds",
						zap.Int("index", i),
						zap.Float64("position", slice.start))
					continue
				}

				canFit := true
				if i > 0 && len(validSlices) > 0 {
					prevSlice := validSlices[len(validSlices)-1]
					if slice.start < prevSlice.start+minGap {
						canFit = false
						log.Warn("Removing slice marker - insufficient space",
							zap.Int("index", i),
							zap.Float64("position", slice.start))
					}
				}

				if canFit {
					validSlices = append(validSlices, slice)
				}
			}

			if len(validSlices) != len(wc.slices) {
				wc.slices = validSlices
			}

			if wc.boundsMarker == nil {
				wc.boundsMarker = NewWaveBoundsMarker()
			}
		} else if cmd.Data == nil {
			wc.boundsStart = 0.0

			if wc.displayData.XLimitMax > 0 {
				wc.boundsEnd = wc.displayData.XLimitMax
			} else {
				wc.boundsEnd = 0.0
			}

		}

	case cmdSetWaveBoundsFromSamples:
		if payload, ok := cmd.Data.(WaveBoundsSamplesPayload); ok {
			if payload.StartSample == 0 && payload.EndSample == 0 {
				break
			}

			startBin := 0.0
			endBin := wc.displayData.XLimitMax
			if wc.samplesPerBin > 0 {
				startBin = float64(payload.StartSample) / wc.samplesPerBin
				endBin = float64(payload.EndSample) / wc.samplesPerBin
			}

			tolerance := 0.1
			if startBin >= 0 && endBin > startBin && endBin <= wc.displayData.XLimitMax+tolerance {
				wc.boundsStart = startBin
				wc.boundsEnd = endBin

				if wc.boundsMarker == nil {
					wc.boundsMarker = NewWaveBoundsMarker()
				}
			} else {
				log.Warn("Invalid bounds calculated from samples, skipping update",
					zap.Float64("startBin", startBin),
					zap.Float64("endBin", endBin),
					zap.Float64("xLimitMax", wc.displayData.XLimitMax))
			}
		}

	case cmdSetWaveCursor:
		if c, ok := cmd.Data.(*WaveCursor); ok {
			wc.cursor = c
		}

	case cmdSetWavePlotFlags:
		if f, ok := cmd.Data.(implot.Flags); ok {
			wc.plotFlags = f
		}

	case cmdSetWaveAxisXFlags:
		if f, ok := cmd.Data.(implot.AxisFlags); ok {
			wc.axisXFlags = f
		}

	case cmdSetWaveAxisYFlags:
		if f, ok := cmd.Data.(implot.AxisFlags); ok {
			wc.axisYFlags = f
		}

	case cmdAddWaveSlice:
		if marker, ok := cmd.Data.(*WaveMarker); ok {
			wc.slices = append(wc.slices, marker)
		}

	case cmdUpdateWaveSlicePosition:
		if payload, ok := cmd.Data.(WaveSlicePositionPayload); ok {
			if payload.Index >= 0 && payload.Index < len(wc.slices) && wc.slices[payload.Index] != nil {
				newStart := payload.NewStart
				minBound := wc.boundsStart
				maxBound := wc.boundsEnd

				if newStart < minBound {
					newStart = minBound
				}

				if newStart > maxBound {
					newStart = maxBound
				}

				minGapBins := 1.0
				if payload.Index > 0 && wc.slices[payload.Index-1] != nil {
					prevEnd := wc.slices[payload.Index-1].start + minGapBins
					if newStart < prevEnd {
						newStart = prevEnd
					}
				}

				if payload.Index < len(wc.slices)-1 && wc.slices[payload.Index+1] != nil {
					nextStart := wc.slices[payload.Index+1].start - minGapBins
					if newStart > nextStart {
						newStart = nextStart
					}
				}
				wc.slices[payload.Index].start = newStart
			}
		} else {
			log.Warn("Invalid data type for cmdUpdateWaveSlicePosition", zap.Any("data", cmd.Data))
		}
	default:
		log.Warn(
			"WaveComponent unhandled update",
			zap.String("id", wc.IDStr()),
			zap.Any("cmd", cmd),
		)
	}
}

func (wc *WaveComponent) handleUserInteraction(xMin, xMax float64) {
	minBound := xMin
	maxBound := xMax
	if wc.boundsMarker != nil {
		minBound = wc.boundsStart
		maxBound = wc.boundsEnd
	}

	// Update hover cursor
	if implot.IsPlotHovered() {
		mp := implot.GetPlotMousePos()
		x := mp.X

		if x < minBound {
			x = minBound
		}

		if x > maxBound {
			x = maxBound
		}

		if wc.cursor != nil {
			wc.cursor.SetHoverPosition(x, true)
		}
	} else {
		if wc.cursor != nil {
			wc.cursor.SetHoverPosition(0, false)
		}
	}

	isDragging := imgui.IsMouseDragging(imgui.MouseButtonLeft) ||
		imgui.IsMouseDragging(imgui.MouseButtonRight) ||
		imgui.IsMouseDragging(imgui.MouseButtonMiddle)

	if implot.IsPlotHovered() && imgui.IsMouseClickedBool(imgui.MouseButtonLeft) && !isDragging {
		mp := implot.GetPlotMousePos()
		x := mp.X

		if x < minBound {
			x = minBound
		}

		if x > maxBound {
			x = maxBound
		}

		snapThreshold := 5.0
		var nearestSlice *WaveMarker
		minDistance := snapThreshold
		for _, slice := range wc.slices {
			if slice != nil {
				sliceDistance := math.Abs(slice.start - x)
				if sliceDistance < minDistance {
					minDistance = sliceDistance
					nearestSlice = slice
				}
			}
		}

		if nearestSlice != nil {
			x = nearestSlice.start
		}

		boundsRange := maxBound - minBound
		seekPosition := 0.0

		if boundsRange > 0 {
			seekPosition = (x - minBound) / boundsRange
		}

		if wc.cursor != nil {
			wc.cursor.SetPositionImmediate(x)
		}

		eventbus.Bus.Publish(events.MouseEventRecord{
			EventType: events.ComponentClickedEvent,
			ImguiID:   wc.ID(),
			UUID:      wc.UUID(),
			Button:    events.MouseButtonLeft,
			State:     wc.State(),
			Data: map[string]interface{}{
				"seek":     true,
				"position": seekPosition,
				"path":     wc.displayData.Path,
			},
		})
	}

	// Middle-click to add slice markers - only if not dragging
	if implot.IsPlotHovered() && imgui.IsMouseClickedBool(imgui.MouseButtonMiddle) && !isDragging {
		mp := implot.GetPlotMousePos()
		x := math.Round(mp.X)
		if x < minBound {
			x = minBound
		}

		if x > maxBound {
			x = maxBound
		}

		if x >= minBound && x <= maxBound {
			marker := NewWaveMarker(x)
			wc.AddSlice(marker)
		}
	}
}

func (wc *WaveComponent) drawOutOfBounds(boundsStartValue, boundsEndValue float64, yMinPlot float64) {
	plotPos := implot.GetPlotPos()
	plotSize := implot.GetPlotSize()
	plotLeftPx := plotPos.X
	plotRightPx := plotPos.X + plotSize.X
	boundsStartPxPos := implot.PlotToPixelsdoubleV(boundsStartValue, yMinPlot, implot.AxisX1, implot.AxisY1)
	boundsEndPxPos := implot.PlotToPixelsdoubleV(boundsEndValue, yMinPlot, implot.AxisX1, implot.AxisY1)
	boundsStartPx := boundsStartPxPos.X
	boundsEndPx := boundsEndPxPos.X

	if boundsStartPx < plotLeftPx {
		boundsStartPx = plotLeftPx
	}

	if boundsStartPx > plotRightPx {
		boundsStartPx = plotRightPx
	}

	if boundsEndPx < plotLeftPx {
		boundsEndPx = plotLeftPx
	}

	if boundsEndPx > plotRightPx {
		boundsEndPx = plotRightPx
	}

	draw := implot.GetPlotDrawList()
	shade := imgui.NewColor(0, 0, 0, 0.25).Pack()
	if boundsStartPx > plotLeftPx {
		draw.AddRectFilled(
			imgui.Vec2{X: plotLeftPx, Y: plotPos.Y},
			imgui.Vec2{X: boundsStartPx, Y: plotPos.Y + plotSize.Y},
			shade,
		)
	}

	if boundsEndPx < plotRightPx {
		draw.AddRectFilled(
			imgui.Vec2{X: boundsEndPx, Y: plotPos.Y},
			imgui.Vec2{X: plotRightPx, Y: plotPos.Y + plotSize.Y},
			shade,
		)
	}
}

func (wc *WaveComponent) drawBoundsMarker(samplesPerBin float64, sampleRate int, xMin, xMax, yMin, yMax float64) {
	interactionMarker := wc.boundsMarker

	if interactionMarker == nil {
		return
	}

	currentStart := wc.boundsStart
	currentEnd := wc.boundsEnd
	newStart, newEnd, changed := interactionMarker.DrawInteract(
		currentStart, currentEnd,
		samplesPerBin, sampleRate,
		xMin, xMax, yMin, yMax,
	)

	if changed {
		wc.SetBounds(newStart, newEnd)
	}
}

func (wc *WaveComponent) drawSliceRegionShading(slices []*WaveMarker, cursorBin float64, boundsStart, boundsEnd float64, yMin, yMax float64, isPlaying bool, isPaused bool) {
	if len(slices) == 0 {
		return
	}

	currentSliceRegion := 0

	for i, slice := range slices {
		if slice != nil && cursorBin >= slice.start {
			currentSliceRegion = i + 1
		}
	}

	var regionStart, regionEnd float64
	shouldShade := false
	for i, slice := range slices {
		if slice != nil && slice.Held {
			regionStart = slice.start
			regionEnd = boundsEnd

			if i < len(slices)-1 && slices[i+1] != nil {
				regionEnd = slices[i+1].start
			}

			shouldShade = true
			break
		}
	}
	isStopped := !isPlaying && !isPaused
	shouldShowRepeatSlice := wc.repeatMode == 2 && isStopped
	if !shouldShade && (isPlaying || isPaused || shouldShowRepeatSlice) {
		var targetSliceRegion int
		if wc.repeatMode == 2 || isPaused {
			targetSliceRegion = wc.repeatSliceIdx
		} else {
			targetSliceRegion = currentSliceRegion
		}

		if targetSliceRegion >= 0 {
			if targetSliceRegion > 0 && targetSliceRegion-1 < len(slices) && slices[targetSliceRegion-1] != nil {
				regionStart = slices[targetSliceRegion-1].start
			} else {
				regionStart = boundsStart
			}
			regionEnd = boundsEnd
			if targetSliceRegion < len(slices) && slices[targetSliceRegion] != nil {
				regionEnd = slices[targetSliceRegion].start
			}
			shouldShade = true
		}
	}

	if shouldShade {
		pMin := implot.PlotToPixelsdoubleV(regionStart, yMin, implot.AxisX1, implot.AxisY1)
		pMax := implot.PlotToPixelsdoubleV(regionEnd, yMax, implot.AxisX1, implot.AxisY1)
		drawList := implot.GetPlotDrawList()
		fillCol := imgui.NewColor(0.3, 0.6, 0.8, 0.15).Pack()
		drawList.AddRectFilled(
			imgui.Vec2{X: pMin.X, Y: pMin.Y},
			imgui.Vec2{X: pMax.X, Y: pMax.Y},
			fillCol,
		)
	}
}

func (wc *WaveComponent) drawSliceMarkers(slices []*WaveMarker, cursor *WaveCursor, boundsStartValue, boundsEndValue float64, xMin, xMax, samplesPerBin float64, sampleRate int, yMin, yMax float64) {
	if len(slices) == 0 {
		return
	}

	minGapSec := 0.010
	minGapBins := 0.0

	if samplesPerBin > 0 {
		minGapBins = math.Ceil((minGapSec * float64(sampleRate)) / samplesPerBin)
	}

	sort.Slice(slices, func(i, j int) bool { return slices[i].start < slices[j].start })

	var currentSliceIdx int = -1

	isPaused := !wc.displayData.IsPlaying && wc.displayData.Progress > 0

	if wc.repeatMode == 2 {
		currentSliceIdx = wc.repeatSliceIdx
	} else if wc.displayData.IsPlaying || isPaused {
		var currentBin float64

		if isPaused && cursor != nil {
			currentBin = cursor.position
		} else {
			boundsRange := boundsEndValue - boundsStartValue
			currentBin = wc.displayData.Progress * boundsRange
		}

		for i, slice := range slices {
			if slice != nil && currentBin >= slice.start {
				currentSliceIdx = i + 1
			}
		}
	}

	needsFilter := false
	var updatesToSend []component.UpdateCmd
	for idx, mk := range slices {
		left, right := mk.Bounds(idx, slices, xMin, xMax, minGapBins, boundsStartValue, boundsEndValue)
		nextStart := math.Inf(1)

		if idx < len(slices)-1 {
			nextStart = slices[idx+1].start
		}

		isCurrentSlice := currentSliceIdx == idx+1
		newStartPos, removalRequested, positionChanged := mk.DrawInteract(
			idx, mk.start,
			left, right, nextStart, minGapBins,
			samplesPerBin, sampleRate, yMin, yMax,
			isCurrentSlice,
		)

		if removalRequested {
			needsFilter = true
			mk.WantRemove = true
		} else if positionChanged {
			updatePayload := WaveSlicePositionPayload{
				Index:    idx,
				NewStart: newStartPos,
			}
			cmd := component.UpdateCmd{Type: cmdUpdateWaveSlicePosition, Data: updatePayload}
			updatesToSend = append(updatesToSend, cmd)
		}
	}

	for _, cmd := range updatesToSend {
		wc.SendUpdate(cmd)
	}

	if needsFilter {
		dst := slices[:0]
		minBound := boundsStartValue
		maxBound := boundsEndValue

		for _, m := range slices {
			if m != nil && !m.WantRemove && m.start >= minBound && m.start <= maxBound {
				dst = append(dst, m)
			}
		}

		wc.SetSlices(dst)
	}
}

func (wc *WaveComponent) drawWaveform(downsamples []audio.Downsample) {
	if len(downsamples) == 0 {
		return
	}

	ds := downsamples[0]
	numSamples := len(ds.Mins)

	if len(ds.Maxs) < numSamples {
		numSamples = len(ds.Maxs)
	}

	if numSamples == 0 {
		return
	}

	xs := make([]float32, numSamples)

	for i := range xs {
		xs[i] = float32(i)
	}

	var minsPtr, maxsPtr *float32
	if numSamples > 0 {
		minsPtr = &ds.Mins[0]
		maxsPtr = &ds.Maxs[0]
	}

	implot.PlotShadedFloatPtrFloatPtrFloatPtr("ch_fill_0", &xs[0], minsPtr, maxsPtr, int32(numSamples))
	implot.PlotLineFloatPtrFloatPtr("ch_lines_min_0", &xs[0], minsPtr, int32(numSamples))
	implot.PlotLineFloatPtrFloatPtr("ch_lines_max_0", &xs[0], maxsPtr, int32(numSamples))
}

func (wc *WaveComponent) drawTimeTicks(sampleRate int, totalSamples int, xLimitMax float64, samplesPerBin float64) {
	xCount := int(xLimitMax) + 1

	if sampleRate <= 0 || totalSamples <= 0 || xCount <= 0 || samplesPerBin <= 0 {
		return
	}

	totalSec := float64(totalSamples) / float64(sampleRate)
	targetTicks := 10.0
	rawStep := totalSec / targetTicks
	niceSteps := []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 15, 30, 60}
	step := niceSteps[len(niceSteps)-1]

	for _, s := range niceSteps {
		if s >= rawStep {
			step = s
			break
		}
	}

	count := int(math.Floor(totalSec/step)) + 1

	if count <= 0 {
		return
	}

	positions := make([]float64, count)
	labels := make([]string, count)

	for i := 0; i < count; i++ {
		tsec := float64(i) * step
		positions[i] = (tsec * float64(sampleRate)) / samplesPerBin
		labels[i] = util.SecondsLabel(tsec)
	}

	var posPtr *float64

	if len(positions) > 0 {
		posPtr = &positions[0]
	}

	implot.SetupAxisTicksdoublePtrV(implot.AxisX1, posPtr, int32(len(positions)), labels, false)
}

func (wc *WaveComponent) drawLoadingState() {
	dummySlice := []int64{0}
	cur := imgui.CursorPos()
	avail := imgui.ContentRegionAvail()
	tempXMin, tempXMax := 0.0, 100.0
	tempYMin, tempYMax := -1.0, 1.0

	if implot.BeginPlotV("dummyPlot", imgui.Vec2{X: -1, Y: -1}, wc.plotFlags) {
		implot.SetupAxesV("Time", "Amplitude", wc.axisXFlags, wc.axisYFlags)
		implot.SetupAxisLimitsV(implot.AxisX1, tempXMin, tempXMax, implot.CondOnce)
		implot.SetupAxisLimitsV(implot.AxisY1, tempYMin, tempYMax, implot.CondOnce)
		implot.PlotShadedS64PtrInt("ch_min_dummy", utils.SliceToPtr(dummySlice), 0)
		implot.EndPlot()
	}

	style := imgui.CurrentStyle()
	radius := float32(24.0)
	spinnerSize := imgui.Vec2{
		X: radius*2 + style.FramePadding().X*2,
		Y: radius*2 + style.FramePadding().Y*2,
	}

	centerLocal := imgui.Vec2{
		X: cur.X + (avail.X-spinnerSize.X)/2,
		Y: cur.Y + (avail.Y-spinnerSize.Y)/2,
	}

	imgui.SetNextItemAllowOverlap()
	imgui.SetCursorPos(centerLocal)

	component.SpinnerBarChartAdvSineFade(
		"SpinnerBarChartAdvSineFade",
		radius,
		8,
		imgui.NewColor(1, 1, 1, 1).Pack(),
		5.8)
}

func (wc *WaveComponent) drawEmptyState() {
	imgui.PushFont(font.FontNabla, 96)
	defer imgui.PopFont()

	msgSize := imgui.CalcTextSize(wc.emptyText)
	cur := imgui.CursorPos()
	avail := imgui.ContentRegionAvail()

	centerLocal := imgui.Vec2{
		X: cur.X + (avail.X-msgSize.X-8)/2,
		Y: cur.Y + (avail.Y-msgSize.Y-8)/2,
	}

	imgui.SetCursorPos(centerLocal)

	component.WavyText(
		wc.emptyText,
		theme.GetCurrentColormap(),
		true,
		imgui.ColorU32Vec4(imgui.Vec4{X: 0, Y: 0, Z: 0, W: 1}),
		imgui.NewVec2(3.0, 3.0),
		imgui.NewVec2(3.2, 1.0),
		imgui.NewVec2(4.0, 8.0),
		0.2,
	)
}

func (wc *WaveComponent) drawCursor(isPlaying bool, progress float64, xLimitMax float64, cursor *WaveCursor, boundsStart, boundsEnd float64) {
	if cursor == nil {
		return
	}

	cursor.UpdateAnimation()
	if !isPlaying {
		hoverPos, isHovering := cursor.GetHoverPosition()
		if isHovering {
			implot.PushStyleVarFloat(implot.StyleVarLineWeight, 2.0)
			hoverColor := imgui.Vec4{X: 1.0, Y: 1.0, Z: 1.0, W: 0.3}
			implot.PushStyleColorVec4(implot.ColLine, hoverColor)
			implot.PlotInfLinesdoublePtr("hover cursor", &hoverPos, 1)
			implot.PopStyleColor()
			implot.PopStyleVar()
		}
	}

	var cursorBin float64
	if isPlaying {
		var targetPos float64
		hasPlaybackBounds := wc.displayData.PlaybackEndMarker > 0

		if wc.samplesPerBin > 0 {
			var effectiveStartMarker, effectiveEndMarker int

			if hasPlaybackBounds {
				effectiveStartMarker = wc.displayData.PlaybackStartMarker
				effectiveEndMarker = wc.displayData.PlaybackEndMarker
			} else {
				// Use component's bounds (convert from bin to sample coordinates)
				effectiveStartMarker = int(wc.boundsStart * wc.samplesPerBin)
				effectiveEndMarker = int(wc.boundsEnd * wc.samplesPerBin)
			}

			playbackRangeSamples := float64(effectiveEndMarker - effectiveStartMarker)
			currentSample := float64(effectiveStartMarker) + (progress * playbackRangeSamples)
			targetPos = currentSample / wc.samplesPerBin
		} else {
			targetPos = boundsStart
		}

		if targetPos < boundsStart {
			targetPos = boundsStart
		}

		if targetPos > boundsEnd {
			targetPos = boundsEnd
		}

		distanceFromTarget := math.Abs(cursor.position - targetPos)
		visibleRange := boundsEnd - boundsStart

		if distanceFromTarget > visibleRange*0.05 {
			cursor.SetPositionImmediate(targetPos)
		} else {
			// ~60fps
			animDuration := 16 * time.Millisecond
			cursor.AnimateToPosition(targetPos, animDuration, animation.EaseLinear)
		}

		cursorBin = cursor.position
	} else {
		if progress == 0 {
			cursorBin = boundsStart
			cursor.SetPositionImmediate(cursorBin)
		} else {
			cursorBin = cursor.position

			if cursorBin < boundsStart || cursorBin > boundsEnd {
				cursorBin = boundsStart
				cursor.SetPositionImmediate(cursorBin)
			}
		}
	}

	implot.PushStyleVarFloat(implot.StyleVarLineWeight, 3.0)
	cursorColor := imgui.Vec4{X: 1.0, Y: 0.3, Z: 0.3, W: 0.8}
	implot.PushStyleColorVec4(implot.ColLine, cursorColor)
	implot.PlotInfLinesdoublePtr("playback cursor", &cursorBin, 1)
	implot.PopStyleColor()
	implot.PopStyleVar()
}

func (wc *WaveComponent) SetWaveDisplayData(data audio.WaveDisplayData) *WaveComponent {
	cmd := component.UpdateCmd{Type: cmdSetWaveDisplayData, Data: data}
	wc.SendUpdate(cmd)
	return wc
}

func (wc *WaveComponent) ResetForNewWave() *WaveComponent {
	wc.boundsInitialized = false
	return wc
}

// GetWaveDisplayData returns the current display data
func (wc *WaveComponent) GetWaveDisplayData() audio.WaveDisplayData {
	return wc.displayData
}

// GetBoundsAndSlices returns the current bounds in sample positions and slice positions in bins
func (wc *WaveComponent) GetBoundsAndSlices() (boundsStartSample, boundsEndSample int, slicePositions []float64) {
	boundsStartSample = int(wc.boundsStart * wc.samplesPerBin)
	boundsEndSample = int(wc.boundsEnd * wc.samplesPerBin)
	slicePositions = make([]float64, len(wc.slices))

	for i, slice := range wc.slices {
		if slice != nil {
			slicePositions[i] = slice.start
		}
	}

	return boundsStartSample, boundsEndSample, slicePositions
}

// GetSamplesPerBin returns the samplesPerBin value
func (wc *WaveComponent) GetSamplesPerBin() float64 {
	return wc.samplesPerBin
}

// GetCursorPosition returns the current cursor position in bins
func (wc *WaveComponent) GetCursorPosition() float64 {
	if wc.cursor != nil {
		return wc.cursor.position
	}
	return 0.0
}

// SetRepeatMode updates the repeat mode
func (wc *WaveComponent) SetRepeatMode(mode int, sliceIdx int) {
	wc.repeatMode = mode
	wc.repeatSliceIdx = sliceIdx
}

func (wc *WaveComponent) SetBounds(start, end float64) *WaveComponent {
	payload := WaveBoundsPayload{Start: start, End: end}
	cmd := component.UpdateCmd{Type: cmdSetWaveBounds, Data: payload}
	wc.SendUpdate(cmd)
	return wc
}

// SetBoundsFromSamples sets bounds using sample positions
func (wc *WaveComponent) SetBoundsFromSamples(startSample, endSample int) *WaveComponent {
	payload := WaveBoundsSamplesPayload{
		StartSample: startSample,
		EndSample:   endSample,
	}

	cmd := component.UpdateCmd{Type: cmdSetWaveBoundsFromSamples, Data: payload}
	wc.SendUpdate(cmd)

	return wc
}

func (wc *WaveComponent) SetSlices(slices []*WaveMarker) *WaveComponent {
	slicesCopy := make([]*WaveMarker, len(slices))
	copy(slicesCopy, slices)
	cmd := component.UpdateCmd{Type: cmdSetWaveSlices, Data: slicesCopy}
	wc.SendUpdate(cmd)
	return wc
}

func (wc *WaveComponent) ClearSlices() {
	cmd := component.UpdateCmd{Type: cmdSetWaveSlices, Data: nil}
	wc.SendUpdate(cmd)
}

func (wc *WaveComponent) AddSlice(marker *WaveMarker) {
	cmd := component.UpdateCmd{Type: cmdAddWaveSlice, Data: marker}
	wc.SendUpdate(cmd)
}

func (wc *WaveComponent) SetPlotFlags(flags implot.Flags) *WaveComponent {
	cmd := component.UpdateCmd{Type: cmdSetWavePlotFlags, Data: flags}
	wc.SendUpdate(cmd)
	return wc
}

func (wc *WaveComponent) SetAxisXFlags(flags implot.AxisFlags) *WaveComponent {
	cmd := component.UpdateCmd{Type: cmdSetWaveAxisXFlags, Data: flags}
	wc.SendUpdate(cmd)
	return wc
}

func (wc *WaveComponent) SetAxisYFlags(flags implot.AxisFlags) *WaveComponent {
	cmd := component.UpdateCmd{Type: cmdSetWaveAxisYFlags, Data: flags}
	wc.SendUpdate(cmd)
	return wc
}

func (wc *WaveComponent) Layout() {
	wc.drainEvents()
	wc.Component.ProcessUpdates()

	displayData := wc.displayData
	slices := wc.slices
	cursor := wc.cursor
	samplesPerBin := wc.samplesPerBin

	if displayData.IsLoading {
		cache := audio.GetGlobalAsyncCache()
		snapshot := cache.GetSnapshot(displayData.Path)
		if snapshot != nil {
			if snapshot.SamplesLoaded {
				am := audio.GetAudioManager()
				if updatedData, err := am.GetWaveDisplayData(displayData.Path); err == nil && updatedData.IsReady {
					wc.SetWaveDisplayData(updatedData)
				}
			}
		}
		wc.drawLoadingState()
		return
	}

	if displayData.Downsamples == nil ||
		len(displayData.Downsamples) == 0 ||
		len(displayData.Downsamples[0].Mins) == 0 {

		wc.drawEmptyState()
		return
	}

	// Setup plot style
	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: 0, Y: 0})
	defer imgui.PopStyleVar()
	implot.PushStyleVarVec2(implot.StyleVarPlotPadding, imgui.Vec2{X: 0, Y: 0})
	defer implot.PopStyleVar()
	implot.PushStyleVarVec2(implot.StyleVarLabelPadding, imgui.Vec2{X: 0, Y: 0})
	defer implot.PopStyleVar()

	// Set the colormap for this plot
	implot.PushColormapPlotColormap(theme.GetCurrentColormap())
	defer implot.PopColormap()

	if implot.BeginPlotV(wc.Component.IDStr(), imgui.Vec2{X: -1, Y: -1}, wc.plotFlags) {
		defer implot.EndPlot()

		// Setup Axis
		xMin, xMax := 0.0, displayData.XLimitMax
		yMin, yMax := float64(displayData.MinY), float64(displayData.MaxY)
		implot.SetupAxesV("Time", "Amplitude", wc.axisXFlags, wc.axisYFlags)
		implot.SetupAxisLimitsV(implot.AxisX1, xMin, xMax, implot.CondOnce)
		implot.SetupAxisLimitsV(implot.AxisY1, yMin, yMax, implot.CondOnce)

		boundsStartValue := wc.boundsStart
		boundsEndValue := wc.boundsEnd

		// Draw elements using ViewModel data
		wc.drawTimeTicks(displayData.SampleRate, displayData.NumSamples, xMax, samplesPerBin)
		wc.handleUserInteraction(xMin, xMax)
		wc.drawWaveform(displayData.Downsamples)
		wc.drawOutOfBounds(boundsStartValue, boundsEndValue, yMin)

		var cursorBin float64
		isPaused := !displayData.IsPlaying && displayData.Progress > 0
		if isPaused && cursor != nil {
			cursorBin = cursor.position
		} else {
			boundsRange := boundsEndValue - boundsStartValue
			cursorBin = boundsStartValue
			if boundsRange > 0 {
				cursorBin = boundsStartValue + (displayData.Progress * boundsRange)
			}
		}

		wc.drawSliceRegionShading(slices, cursorBin, boundsStartValue, boundsEndValue, yMin, yMax, displayData.IsPlaying, isPaused)
		wc.drawCursor(displayData.IsPlaying, displayData.Progress, xMax, cursor, boundsStartValue, boundsEndValue)
		wc.drawBoundsMarker(
			samplesPerBin,
			displayData.SampleRate,
			xMin, xMax,
			yMin, yMax,
		)
		wc.drawSliceMarkers(slices, cursor, boundsStartValue, boundsEndValue, xMin, xMax, samplesPerBin, displayData.SampleRate, yMin, yMax)
	}
}

// Destroy cleans up the component
func (wc *WaveComponent) Destroy() {
	// Unsubscribe from all events
	bus := eventbus.Bus
	uuid := wc.UUID()
	bus.Unsubscribe(events.AudioPlaybackProgressKey, uuid)
	bus.Unsubscribe(events.AudioPlaybackStartedKey, uuid)
	bus.Unsubscribe(events.AudioPlaybackStoppedKey, uuid)
	bus.Unsubscribe(events.AudioPlaybackFinishedKey, uuid)

	// Call the base component's destroy method
	wc.Component.Destroy()
}
