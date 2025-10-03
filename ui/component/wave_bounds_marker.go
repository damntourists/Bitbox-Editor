package component

import (
	"bitbox-editor/ui/fonts"
	"bitbox-editor/ui/theme"
	"fmt"
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
)

type WaveBoundsMarker struct {
	Start float64
	End   float64

	StartHovered bool
	StartHeld    bool

	EndHovered bool
	EndHeld    bool
}

func (m *WaveBoundsMarker) StartBin() float64 { return m.Start }
func (m *WaveBoundsMarker) EndBin() float64   { return m.End }

func (m *WaveBoundsMarker) StartSeconds(sampleRate int, samplesPerBin float64) float64 {
	return (m.Start * samplesPerBin) / float64(sampleRate)
}

func (m *WaveBoundsMarker) EndSeconds(sampleRate int, samplesPerBin float64) float64 {
	return (m.End * samplesPerBin) / float64(sampleRate)
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

func (m *WaveBoundsMarker) drawOutOfBoundsShading(xMin, xMax, yMin, yMax float64) {
	t := theme.GetCurrentTheme()
	plotPos := implot.GetPlotPos()
	plotSize := implot.GetPlotSize()

	pStart := implot.PlotToPixelsdoubleV(m.Start, yMin, implot.AxisX1, implot.AxisY1)
	pEnd := implot.PlotToPixelsdoubleV(m.End, yMin, implot.AxisX1, implot.AxisY1)
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
	samplesPerBin float64,
	sampleRate int,
	xMin, xMax float64,
	yMin, yMax float64,
) {
	m.drawOutOfBoundsShading(xMin, xMax, yMin, yMax)

	t := theme.GetCurrentTheme()

	startColor := t.Style.Colors.TextSelectedBg.Vec4
	endColor := t.Style.Colors.TextSelectedBg.Vec4
	flags := implot.DragToolFlagsNone

	// Draw start
	var startClicked = false
	if implot.DragLineXV(
		-1,
		&m.Start,
		startColor,
		3.0,
		flags,
		&startClicked,
		&m.StartHovered,
		&m.StartHeld) {

		if m.Start < xMin {
			m.Start = xMin
		}
		if m.Start > m.End {
			m.Start = m.End
		}

		// Show timestamp while dragging
		sec := (m.Start * samplesPerBin) / float64(sampleRate)
		implot.AnnotationStr(
			m.Start,
			0,
			startColor,
			imgui.Vec2{X: 0, Y: 0},
			true,
			fmt.Sprintf("Start: %s", secondsLabel(sec)),
		)
	}

	var endClicked = false
	if implot.DragLineXV(
		-2,
		&m.End,
		endColor,
		3.0,
		flags,
		&endClicked,
		&m.EndHovered,
		&m.EndHeld) {

		if m.End > xMax {
			m.End = xMax
		}
		if m.End < m.Start {
			m.End = m.Start
		}

		// Show timestamp while dragging
		sec := (m.End * samplesPerBin) / float64(sampleRate)
		implot.AnnotationStr(
			m.End,
			0,
			endColor,
			imgui.Vec2{X: 0, Y: 0},
			true,
			fmt.Sprintf("End: %s", secondsLabel(sec)),
		)
	}

	// Shade the region between start and end markers
	if m.StartHeld || m.EndHeld || m.StartHovered || m.EndHovered {
		pMin := implot.PlotToPixelsdoubleV(m.Start, yMin, implot.AxisX1, implot.AxisY1)
		pMax := implot.PlotToPixelsdoubleV(m.End, yMax, implot.AxisX1, implot.AxisY1)

		drawList := implot.GetPlotDrawList()
		fillCol := imgui.NewColor(0.5, 0.5, 0.5, 0.15).Pack()
		drawList.AddRectFilled(
			imgui.Vec2{X: pMin.X, Y: pMin.Y},
			imgui.Vec2{X: pMax.X, Y: pMax.Y},
			fillCol,
		)
	}

	// Draw labels at the top
	implot.AnnotationStr(
		m.Start,
		yMax,
		startColor,
		imgui.Vec2{X: 0, Y: -10},
		true,
		fonts.Icon("ArrowRightFromLine"),
	)

	implot.AnnotationStr(
		m.End,
		yMax,
		endColor,
		imgui.Vec2{X: 0, Y: -10},
		true,
		fonts.Icon("ArrowLeftFromLine"),
	)
}

func NewWaveBoundsMarker(start, end float64) *WaveBoundsMarker {
	return &WaveBoundsMarker{
		Start: start,
		End:   end,
	}
}
