package component

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"bitbox-editor/lib/audio"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
	"github.com/AllenDang/cimgui-go/utils"
)

type WaveComponent struct {
	*Component

	label string

	wav *audio.WaveFile

	cursor *WaveMarker

	slices []*WaveMarker

	plotFlags  implot.Flags
	axisXFlags implot.AxisFlags
	axisYFlags implot.AxisFlags

	xLimitMin, xLimitMax float64
	yLimitMin, yLimitMax float64

	samplesPerBin float64

	loading bool
}

func (wc *WaveComponent) LoadWav(path string) {
	wc.loading = true
	go func() {
		var err error
		wc.wav, err = audio.LoadWAVFile(path)
		if err != nil {
			wc.loading = false
			panic(err)
		}

		wc.calculatePlotLimits()
		wc.calculateSamplesPerBin()

		wc.loading = false
	}()
}

func (wc *WaveComponent) SetWav(wav *audio.WaveFile) {
	wc.loading = true
	wc.wav = wav
	wc.calculatePlotLimits()
	wc.calculateSamplesPerBin()
	wc.loading = false
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

	// Data range in bins
	dataStart := 0.0
	dataEnd := wc.xLimitMax

	// Convert dataStart/dataEnd
	yMin := float64(wc.wav.MinY)
	pStart := implot.PlotToPixelsdoubleV(dataStart, yMin, implot.AxisX1, implot.AxisY1)
	pEnd := implot.PlotToPixelsdoubleV(dataEnd, yMin, implot.AxisX1, implot.AxisY1)

	// Clamp to plot horizontal bounds
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

	// Draw semi-transparent overlays on left and right outside the data range
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

func (wc *WaveComponent) Layout() {
	if wc.wav == nil && !wc.loading {
		imgui.Text("No wav loaded")
		return
	}

	if wc.loading {
		// Create an empty plot and display spinner over it
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

		// Center within the remaining content.
		centerLocal := imgui.Vec2{
			X: cur.X + (avail.X-spinnerSize.X)/2,
			Y: cur.Y + (avail.Y-spinnerSize.Y)/2,
		}

		// Display loading spinner overlapping dummy plot
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

	// Create actual plot once wav is fully loaded
	if implot.BeginPlotV(wc.label, imgui.Vec2{X: -1, Y: -1}, wc.plotFlags) {
		implot.SetupAxesV("Time", "Amplitude", wc.axisXFlags, wc.axisYFlags)
		implot.SetupAxisLimitsV(implot.AxisX1, wc.xLimitMin, wc.xLimitMax, implot.CondOnce)
		implot.SetupAxisLimitsV(implot.AxisY1, wc.yLimitMin, wc.yLimitMax, implot.CondOnce)

		// Add time tick marks
		setupTimeTicks(wc.wav.SampleRate, wc.wav.Len, int(wc.xLimitMax)+1)

		if implot.IsPlotHovered() && imgui.IsMouseClickedBool(imgui.MouseButtonMiddle) {
			mp := implot.GetPlotMousePos()

			// Snap to nearest bin and clamp
			x := math.Round(mp.X)
			if x < 0 {
				x = 0
			}

			if x > wc.xLimitMax {
				x = wc.xLimitMax
			}

			// Create a slice marker at the clicked bin
			mk := &WaveMarker{
				start: x,
			}
			wc.slices = append(wc.slices, mk)
		}

		// TODO: Make colormap configurable
		implot.PushColormapPlotColormap(implot.ColormapCool)
		defer implot.PopColormap()

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

		// Shade outside the WAV range to make start/end obvious
		wc.drawOutOfBounds()

		// Minimum gap between markers
		minGapSec := 0.010
		minGapBins := math.Ceil((minGapSec * float64(wc.wav.SampleRate)) / wc.samplesPerBin)

		if len(wc.slices) > 0 {
			sort.Slice(wc.slices, func(i, j int) bool { return wc.slices[i].start < wc.slices[j].start })

			for idx, mk := range wc.slices {
				left, right := mk.Bounds(idx, wc.slices, wc.xLimitMin, wc.xLimitMax, minGapBins)
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

			// Remove markers that request removal
			if func() bool {
				removed := false
				dst := wc.slices[:0]
				for _, m := range wc.slices {
					if m.WantRemove {
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
				// other actions to perform after removal here
			}
		}

		implot.EndPlot()
	}
}

func NewWaveformComponent(label string) *WaveComponent {
	cmp := &WaveComponent{
		Component: NewComponent(ID(label)),
		label:     label,
		cursor:    &WaveMarker{start: 0.0},

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

	defer cmp.LoadWav("/home/brett/Downloads/micro_bundle-v3/Soundopolis/Sci Fi/Texture_Alien_Bugs_002.wav") //filepath)

	return cmp
}

func secondsLabel(sec float64) string {
	// mm:ss.mmm
	d := time.Duration(sec * float64(time.Second))
	minutes := int(d / time.Minute)
	rem := d % time.Minute
	return fmt.Sprintf("%02d:%06.2f", minutes, rem.Seconds())
}

func setupTimeTicks(sampleRate int, totalSamples int, xCount int) {
	if sampleRate <= 0 || totalSamples <= 0 || xCount <= 0 {
		return
	}
	totalSec := float64(totalSamples) / float64(sampleRate)
	samplesPerBin := float64(totalSamples) / float64(xCount)

	// Choose a step that gives ~8–12 ticks across the whole file (tune as needed)
	targetTicks := 10.0
	rawStep := totalSec / targetTicks
	// Snap to “nice” steps: 0.1, 0.2, 0.5, 1, 2, 5, 10, ...
	niceSteps := []float64{0.1, 0.2, 0.5, 1, 2, 5, 10, 15, 30, 60}
	step := niceSteps[len(niceSteps)-1]
	for _, s := range niceSteps {
		if s >= rawStep {
			step = s
			break
		}
	}

	// Positions are bin indices corresponding to each tick time:
	// index = (tsec * sampleRate) / samplesPerBin
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
