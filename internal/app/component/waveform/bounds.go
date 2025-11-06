package waveform

import (
	"bitbox-editor/internal/app/font"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/util"
	"fmt"
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
)

// TODO: Finish docstrings

type WaveBoundsMarker struct {
	StartHovered bool
	StartHeld    bool
	EndHovered   bool
	EndHeld      bool
}

func NewWaveBoundsMarker() *WaveBoundsMarker {
	return &WaveBoundsMarker{}
}

func CalculateStartSeconds(startBin, samplesPerBin float64, sampleRate int) float64 {
	if sampleRate == 0 {
		return 0
	}
	return (startBin * samplesPerBin) / float64(sampleRate)
}

func CalculateEndSeconds(endBin, samplesPerBin float64, sampleRate int) float64 {
	if sampleRate == 0 {
		return 0
	}
	return (endBin * samplesPerBin) / float64(sampleRate)
}

func (m *WaveBoundsMarker) drawDiagonalStripes(
	drawList *imgui.DrawList,
	pMin, pMax imgui.Vec2,
	stripeColor uint32,
	spacing float32,
	thickness float32,
) {
	width := pMax.X - pMin.X
	height := pMax.Y - pMin.Y

	startOffset := -height
	endOffset := width

	for offset := startOffset; offset < endOffset; offset += spacing {
		x1 := pMin.X + offset
		y1 := pMin.Y
		x2 := pMin.X + offset + height
		y2 := pMax.Y

		if x1 < pMin.X {
			diff := pMin.X - x1
			x1 = pMin.X
			y1 = pMin.Y + diff
		}
		if x2 > pMax.X {
			diff := x2 - pMax.X
			x2 = pMax.X
			y2 = pMax.Y - diff
		}

		if x1 <= pMax.X && x2 >= pMin.X && y1 <= pMax.Y && y2 >= pMin.Y {
			drawList.AddLineV(
				imgui.Vec2{X: x1, Y: y1},
				imgui.Vec2{X: x2, Y: y2},
				stripeColor,
				thickness,
			)
		}
	}
}

func (m *WaveBoundsMarker) drawOutOfBoundsShading(start, end, xMin, xMax, yMin, yMax float64) {
	t := theme.GetCurrentTheme()

	plotPos := implot.GetPlotPos()
	plotSize := implot.GetPlotSize()

	pStart := implot.PlotToPixelsdoubleV(start, yMin, implot.AxisX1, implot.AxisY1)
	pEnd := implot.PlotToPixelsdoubleV(end, yMin, implot.AxisX1, implot.AxisY1)
	pPlotMin := implot.PlotToPixelsdoubleV(xMin, yMin, implot.AxisX1, implot.AxisY1)
	pPlotMax := implot.PlotToPixelsdoubleV(xMax, yMax, implot.AxisX1, implot.AxisY1)

	drawList := implot.GetPlotDrawList()

	baseFillColor := t.Style.Colors.ChildBg
	stripeColor := t.Style.Colors.ButtonActive

	plotLeft := plotPos.X
	plotRight := plotPos.X + plotSize.X
	plotTop := plotPos.Y
	plotBottom := plotPos.Y + plotSize.Y

	baseOpacity := float32(0.6)
	stripeOpacity := float32(0.9)
	spacing := float32(12)
	thickness := float32(4)

	if pStart.X > plotLeft {
		leftStart := math.Max(float64(plotLeft), float64(pPlotMin.X))
		leftEnd := math.Min(float64(pStart.X), float64(plotRight))

		if leftEnd > leftStart {
			drawList.AddRectFilled(
				imgui.Vec2{X: float32(leftStart), Y: plotTop},
				imgui.Vec2{X: float32(leftEnd), Y: plotBottom},
				imgui.NewColor(baseFillColor.X, baseFillColor.Y, baseFillColor.Z, baseOpacity).Pack(),
			)

			m.drawDiagonalStripes(
				drawList,
				imgui.Vec2{X: float32(leftStart), Y: plotTop},
				imgui.Vec2{X: float32(leftEnd), Y: plotBottom},
				imgui.NewColor(stripeColor.X, stripeColor.Y, stripeColor.Z, stripeOpacity).Pack(),
				spacing,
				thickness,
			)
		}
	}

	if pEnd.X < plotRight {
		rightStart := math.Max(float64(pEnd.X), float64(plotLeft))
		rightEnd := math.Min(float64(plotRight), float64(pPlotMax.X))

		if rightEnd > rightStart {
			drawList.AddRectFilled(
				imgui.Vec2{X: float32(rightStart), Y: plotTop},
				imgui.Vec2{X: float32(rightEnd), Y: plotBottom},
				imgui.NewColor(baseFillColor.X, baseFillColor.Y, baseFillColor.Z, baseOpacity).Pack(),
			)

			m.drawDiagonalStripes(
				drawList,
				imgui.Vec2{X: float32(rightStart), Y: plotTop},
				imgui.Vec2{X: float32(rightEnd), Y: plotBottom},
				imgui.NewColor(stripeColor.X, stripeColor.Y, stripeColor.Z, stripeOpacity).Pack(),
				spacing,
				thickness,
			)
		}
	}
}

func (m *WaveBoundsMarker) DrawInteract(
	currentStart, currentEnd float64,
	samplesPerBin float64,
	sampleRate int,
	xMin, xMax float64,
	yMin, yMax float64,
) (newStart, newEnd float64, changed bool) {

	newStart = currentStart
	newEnd = currentEnd
	changed = false

	m.drawOutOfBoundsShading(newStart, newEnd, xMin, xMax, yMin, yMax)

	t := theme.GetCurrentTheme()
	startColor := t.Style.Colors.TextSelectedBg.Vec4
	endColor := t.Style.Colors.TextSelectedBg.Vec4
	flags := implot.DragToolFlagsNone

	// Draw start
	var startClicked = false
	tempStart := newStart
	minGap := 1.0
	if implot.DragLineXV(
		-1,
		&tempStart,
		startColor,
		3.0,
		flags,
		&startClicked,
		&m.StartHovered,
		&m.StartHeld) {

		// Clamp
		if tempStart < xMin {
			tempStart = xMin
		}
		if tempStart > newEnd-minGap {
			tempStart = newEnd - minGap
		}
		if tempStart != newStart {
			newStart = tempStart
			changed = true
		}

		// Show timestamp while dragging
		sec := CalculateStartSeconds(newStart, samplesPerBin, sampleRate)
		implot.AnnotationStr(
			newStart, 0, startColor, imgui.Vec2{X: 0, Y: 0}, true,
			fmt.Sprintf("Start: %s", util.SecondsLabel(sec)), // Ensure secondsLabel exists
		)
	} else {
		clampedStart := newStart
		if clampedStart < xMin {
			clampedStart = xMin
		}

		if clampedStart > newEnd-minGap {
			clampedStart = newEnd - minGap
		}

		if clampedStart != newStart {
			newStart = clampedStart
			changed = true
		}
	}

	var endClicked = false
	tempEnd := newEnd
	if implot.DragLineXV(
		-2,
		&tempEnd,
		endColor,
		3.0,
		flags,
		&endClicked,
		&m.EndHovered,
		&m.EndHeld) {

		if tempEnd > xMax {
			tempEnd = xMax
		}

		if tempEnd < newStart+minGap {
			tempEnd = newStart + minGap
		}

		if tempEnd != newEnd {
			newEnd = tempEnd
			changed = true
		}

		sec := CalculateEndSeconds(newEnd, samplesPerBin, sampleRate)
		implot.AnnotationStr(
			newEnd, 0, endColor, imgui.Vec2{X: 0, Y: 0}, true,
			fmt.Sprintf("End: %s", util.SecondsLabel(sec)),
		)
	} else {
		clampedEnd := newEnd
		if clampedEnd > xMax {
			clampedEnd = xMax
		}

		if clampedEnd < newStart+minGap {
			clampedEnd = newStart + minGap
		}
		if clampedEnd != newEnd {
			newEnd = clampedEnd
			changed = true
		}
	}

	// Draw a fill if the bounds are being dragged
	if m.StartHeld || m.EndHeld || m.StartHovered || m.EndHovered {
		pMin := implot.PlotToPixelsdoubleV(newStart, yMin, implot.AxisX1, implot.AxisY1)
		pMax := implot.PlotToPixelsdoubleV(newEnd, yMax, implot.AxisX1, implot.AxisY1)
		drawList := implot.GetPlotDrawList()
		fillCol := imgui.NewColor(0.5, 0.5, 0.5, 0.15).Pack()
		drawList.AddRectFilled(
			imgui.Vec2{X: pMin.X, Y: pMin.Y},
			imgui.Vec2{X: pMax.X, Y: pMax.Y},
			fillCol,
		)
	}

	// Draw labels
	implot.AnnotationStr(
		newStart,
		yMax,
		startColor,
		imgui.Vec2{X: 0, Y: -1},
		true,
		font.Icon("ArrowRightFromLine"),
	)
	implot.AnnotationStr(
		newEnd,
		yMax,
		endColor,
		imgui.Vec2{X: 0, Y: -1},
		true,
		font.Icon("ArrowLeftFromLine"),
	)

	return newStart, newEnd, changed
}
