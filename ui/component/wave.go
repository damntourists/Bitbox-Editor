package component

import (
	"bitbox-editor/lib/parsing/bitbox"
	"bitbox-editor/ui/theme"
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"

	"bitbox-editor/lib/audio"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
	"github.com/AllenDang/cimgui-go/utils"
)

type WaveComponent struct {
	*Component

	wav  *audio.WaveFile
	cell *bitbox.Cell

	cursor *WaveCursor

	slices []*WaveMarker

	bounds *WaveBoundsMarker

	plotFlags  implot.Flags
	axisXFlags implot.AxisFlags
	axisYFlags implot.AxisFlags

	xLimitMin, xLimitMax float64
	yLimitMin, yLimitMax float64

	samplesPerBin float64

	mu   sync.Mutex
	once sync.Once
}

func (wc *WaveComponent) Bounds() *WaveBoundsMarker {
	return wc.bounds
}

func (wc *WaveComponent) SetBounds(start, end float64) {
	if wc.bounds == nil {
		wc.bounds = NewWaveBoundsMarker(start, end)
	} else {
		wc.bounds.Start = start
		wc.bounds.End = end
	}
}

func (wc *WaveComponent) Cursor() *WaveCursor {
	return wc.cursor
}

func (wc *WaveComponent) Slices() []*WaveMarker {
	return wc.slices
}

func (wc *WaveComponent) Wav() *audio.WaveFile {
	return wc.wav
}

func (wc *WaveComponent) SetCursor(cursor *WaveCursor) {
	wc.cursor = cursor
}

func (wc *WaveComponent) SetSlices(slices []*WaveMarker) {
	wc.slices = slices
}

func (wc *WaveComponent) ClearSlices() {
	wc.slices = make([]*WaveMarker, 0)
}

func (wc *WaveComponent) IsLoading() bool {
	return wc.wav.IsLoading()
}

func (wc *WaveComponent) SetCell(cell *bitbox.Cell) {
	wc.cell = cell
}

func (wc *WaveComponent) SetWav(wav *audio.WaveFile) {
	wc.ClearSlices()

	wc.wav = wav

	err := wav.Load()
	if err != nil {
		log.Error(err.Error())
	}
	wc.calculatePlotLimits()
	wc.calculateSamplesPerBin()
	wc.SetBounds(0, wc.xLimitMax)
}

func (wc *WaveComponent) calculatePlotLimits() {
	xCount := 0
	useDown := false
	if len(wc.wav.Channels) > 0 {
		ys0 := wc.wav.Channels[0]
		useDown = len(ys0) > 2000
		if useDown {
			xCount = len(wc.wav.Downsamples[0].Mins)
		} else {
			xCount = len(ys0)
		}
	}

	if xCount <= 0 {
		xCount = 1
	}

	wc.xLimitMin = 0
	wc.xLimitMax = float64(xCount - 1)

	wc.yLimitMin = float64(wc.wav.MinY)
	wc.yLimitMax = float64(wc.wav.MaxY)
}

func (wc *WaveComponent) calculateSamplesPerBin() {
	wc.samplesPerBin = float64(wc.wav.Len) / float64(wc.xLimitMax+1)
}

func (wc *WaveComponent) drawOutOfBounds() {
	plotPos := implot.GetPlotPos()
	plotSize := implot.GetPlotSize()
	plotLeftPx := plotPos.X
	plotRightPx := plotPos.X + plotSize.X

	dataStart := 0.0
	dataEnd := wc.xLimitMax

	yMin := float64(wc.wav.MinY)
	pStart := implot.PlotToPixelsdoubleV(dataStart, yMin, implot.AxisX1, implot.AxisY1)
	pEnd := implot.PlotToPixelsdoubleV(dataEnd, yMin, implot.AxisX1, implot.AxisY1)

	startPx := pStart.X
	endPx := pEnd.X
	if startPx < plotLeftPx {
		startPx = plotLeftPx
	}
	if startPx > plotRightPx {
		startPx = plotRightPx
	}
	if endPx < plotLeftPx {
		endPx = plotLeftPx
	}
	if endPx > plotRightPx {
		endPx = plotRightPx
	}

	draw := implot.GetPlotDrawList()
	shade := imgui.NewColor(0, 0, 0, 0.25).Pack()

	// Left
	if startPx > plotLeftPx {
		draw.AddRectFilled(
			imgui.Vec2{X: plotLeftPx, Y: plotPos.Y},
			imgui.Vec2{X: startPx, Y: plotPos.Y + plotSize.Y},
			shade,
		)
	}

	// Right
	if endPx < plotRightPx {
		draw.AddRectFilled(
			imgui.Vec2{X: endPx, Y: plotPos.Y},
			imgui.Vec2{X: plotRightPx, Y: plotPos.Y + plotSize.Y},
			shade,
		)
	}

}

func (wc *WaveComponent) drawBoundsMarker() {
	if wc.bounds != nil {
		wc.bounds.DrawInteract(
			wc.samplesPerBin,
			wc.wav.SampleRate,
			wc.xLimitMin,
			wc.xLimitMax,
			float64(wc.wav.MinY),
			float64(wc.wav.MaxY),
		)
	}
}

func (wc *WaveComponent) drawCursor() {
	if wc.wav.IsPlaying() {
		t := theme.GetCurrentTheme()

		progress := wc.wav.Progress()
		cursorBin := progress * wc.xLimitMax
		implot.PushStyleColorVec4(implot.ColLine, t.Style.Colors.NavHighlight.Vec4)
		implot.PlotInfLinesdoublePtr(
			"playback cursor", &cursorBin, 1,
		)
		implot.PopStyleColor()

		if wc.cursor != nil {
			wc.cursor.position = cursorBin
		}
	}
}

func (wc *WaveComponent) drawSliceMarkers() {
	minGapSec := 0.010
	minGapBins := math.Ceil((minGapSec * float64(wc.wav.SampleRate)) / wc.samplesPerBin)

	if len(wc.slices) > 0 {
		sort.Slice(wc.slices, func(i, j int) bool { return wc.slices[i].start < wc.slices[j].start })

		for idx, mk := range wc.slices {
			left, right := mk.Bounds(
				idx,
				wc.slices,
				wc.xLimitMin,
				wc.xLimitMax,
				minGapBins,
				wc.bounds,
			)
			nextStart := math.Inf(1)
			if idx < len(wc.slices)-1 {
				nextStart = wc.slices[idx+1].start
			}

			mk.DrawInteract(
				idx,
				left, right,
				nextStart,
				minGapBins,
				wc.samplesPerBin,
				wc.wav.SampleRate,
				float64(wc.wav.MinY),
				float64(wc.wav.MaxY),
			)
		}

		if func() bool {
			removed := false
			dst := wc.slices[:0]

			minBound := 0.0
			maxBound := wc.xLimitMax
			if wc.bounds != nil {
				minBound = wc.bounds.Start
				maxBound = wc.bounds.End
			}

			for _, m := range wc.slices {
				if m.WantRemove || m.start < minBound || m.start > maxBound {
					removed = true
					continue
				}
				dst = append(dst, m)
			}
			if removed {
				wc.slices = dst
			}
			return removed
		}() {
			// other actions to perform after removal
		}
	}
}

func (wc *WaveComponent) drawWaveform() {

	for c := 0; c < len(wc.wav.Channels); c++ {

		ys := wc.wav.Channels[c]

		if len(ys) > 2000 {
			// Plot downsampled data
			implot.PlotShadedS64PtrInt(
				"ch_min_"+strconv.Itoa(c),
				utils.SliceToPtr(wc.wav.Downsamples[c].Mins),
				int32(len(wc.wav.Downsamples[c].Mins)),
			)
			implot.PlotShadedS64PtrInt(
				"ch_max_"+strconv.Itoa(c),
				utils.SliceToPtr(wc.wav.Downsamples[c].Maxs),
				int32(len(wc.wav.Downsamples[c].Maxs)),
			)
		} else {
			// Plot standard data
			implot.PlotShadedS64PtrInt(
				"ch_"+strconv.Itoa(c),
				utils.SliceToPtr(ys),
				int32(len(ys)),
			)
		}
	}
}

func (wc *WaveComponent) handleUserInteraction() {
	if implot.IsPlotHovered() &&
		imgui.IsMouseClickedBool(imgui.MouseButtonMiddle) {
		mp := implot.GetPlotMousePos()

		x := math.Round(mp.X)

		minBound := 0.0
		maxBound := wc.xLimitMax

		if wc.bounds != nil {
			minBound = wc.bounds.Start
			maxBound = wc.bounds.End
		}

		if x < minBound {
			x = minBound
		}

		if x > maxBound {
			x = maxBound
		}

		if x >= minBound && x <= maxBound {
			mk := &WaveMarker{
				start: x,
			}
			wc.slices = append(wc.slices, mk)

		}
	}
}

func (wc *WaveComponent) drawTimeTicks() {

	sampleRate := wc.wav.SampleRate
	totalSamples := wc.wav.Len
	xCount := int(wc.xLimitMax) + 1

	if sampleRate <= 0 || totalSamples <= 0 || xCount <= 0 {
		return
	}
	totalSec := float64(totalSamples) / float64(sampleRate)
	samplesPerBin := float64(totalSamples) / float64(xCount)

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
	positions := make([]float64, count)
	labels := make([]string, count)
	for i := 0; i < count; i++ {
		tsec := float64(i) * step
		positions[i] = (tsec * float64(sampleRate)) / samplesPerBin // bin index
		labels[i] = secondsLabel(tsec)
	}

	implot.SetupAxisTicksdoublePtrV(
		implot.AxisX1,
		utils.SliceToPtr(positions),
		int32(len(positions)),
		labels,
		false)
}

func (wc *WaveComponent) Layout() {
	if wc.Wav() == nil {
		imgui.Text("No wav loaded")
		return
	}

	if wc.IsLoading() {
		dummySlice := []int64{0}

		cur := imgui.CursorPos()
		avail := imgui.ContentRegionAvail()

		if implot.BeginPlotV("dummyPlot", imgui.Vec2{X: -1, Y: -1}, wc.plotFlags) {
			implot.SetupAxesV("Time", "Amplitude", wc.axisXFlags, wc.axisYFlags)
			implot.SetupAxisLimitsV(implot.AxisX1, wc.xLimitMin, wc.xLimitMax, implot.CondOnce)
			implot.SetupAxisLimitsV(implot.AxisY1, wc.yLimitMin, wc.yLimitMax, implot.CondOnce)
			implot.PlotShadedS64PtrInt(
				"ch_min_dummy",
				utils.SliceToPtr(dummySlice),
				0,
			)
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

		SpinnerBarChartAdvSineFade(
			"SpinnerBarChartAdvSineFade",
			radius,
			8,
			imgui.NewColor(1, 1, 1, 1).Pack(),
			5.8)

		return
	}

	imgui.PushStyleVarVec2(imgui.StyleVarFramePadding, imgui.Vec2{X: 0, Y: 0})
	defer imgui.PopStyleVar()

	implot.PushStyleVarVec2(implot.StyleVarPlotPadding, imgui.Vec2{X: 0, Y: 0})
	defer implot.PopStyleVar()

	implot.PushStyleVarVec2(implot.StyleVarLabelPadding, imgui.Vec2{X: 0, Y: 0})
	defer implot.PopStyleVar()

	// TODO: Make colormap configurable
	implot.PushColormapPlotColormap(implot.ColormapCool)
	defer implot.PopColormap()

	if implot.BeginPlotV(wc.Component.IDStr(), imgui.Vec2{X: -1, Y: -1}, wc.plotFlags) {
		implot.SetupAxesV("Time", "Amplitude", wc.axisXFlags, wc.axisYFlags)
		implot.SetupAxisLimitsV(implot.AxisX1, wc.xLimitMin, wc.xLimitMax, implot.CondOnce)
		implot.SetupAxisLimitsV(implot.AxisY1, wc.yLimitMin, wc.yLimitMax, implot.CondOnce)

		wc.drawTimeTicks()

		wc.handleUserInteraction()

		wc.drawWaveform()

		wc.drawOutOfBounds()

		wc.drawCursor()

		wc.drawBoundsMarker()

		wc.drawSliceMarkers()

		implot.EndPlot()
	}
}

func secondsLabel(sec float64) string {
	// mm:ss.mmm
	d := time.Duration(sec * float64(time.Second))
	minutes := int(d / time.Minute)
	rem := d % time.Minute
	return fmt.Sprintf("%02d:%06.2f", minutes, rem.Seconds())
}

func NewWaveformComponent(id imgui.ID) *WaveComponent {
	cmp := &WaveComponent{
		Component: NewComponent(id),
		wav:       nil,
		cursor:    &WaveCursor{position: 0.0},
		slices:    make([]*WaveMarker, 0),

		plotFlags: implot.FlagsNoMenus |
			implot.FlagsCrosshairs |
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
	}

	cmp.Component.layoutBuilder = cmp

	return cmp
}
