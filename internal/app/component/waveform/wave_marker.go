package waveform

import (
	"bitbox-editor/internal/util"
	"fmt"
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
)

type WaveMarker struct {
	start float64

	Active     bool
	Hovered    bool
	Held       bool
	WantRemove bool
}

func NewWaveMarker(start float64) *WaveMarker {
	return &WaveMarker{start: start}
}

func (m *WaveMarker) Bounds(
	idx int,
	slices []*WaveMarker,
	xMin, xMax, minGapBins float64,
	boundsStartValue, boundsEndValue float64,
) (left, right float64) {

	left = math.Max(xMin, boundsStartValue)
	right = math.Min(xMax, boundsEndValue)

	// Constrain by previous marker (with extra buffer to prevent overlap)
	if idx > 0 && slices[idx-1] != nil {
		left = math.Max(left, slices[idx-1].start+minGapBins)
	}

	// Constrain by next marker (with extra buffer to prevent overlap)
	if idx < len(slices)-1 && slices[idx+1] != nil {
		right = math.Min(right, slices[idx+1].start-minGapBins)
	}

	// Ensure left is not greater than right - if they would overlap, clamp to left
	if right < left {
		right = left
	}
	return
}

func (m *WaveMarker) StartBin() float64 { return m.start }

func (m *WaveMarker) StartSeconds(sampleRate int, samplesPerBin float64) float64 {
	if sampleRate == 0 {
		return 0
	}
	return (m.start * samplesPerBin) / float64(sampleRate)
}

func (m *WaveMarker) DrawInteract(
	idx int,
	currentStart float64,
	leftBound, rightBound, nextStart, minGapBins, samplesPerBin float64,
	sampleRate int,
	yMin, yMax float64,
	isCurrentSlice bool,
) (newStart float64, wantRemove bool, changed bool) {

	newStart = currentStart
	wantRemove = false
	changed = false

	var col imgui.Vec4
	var labelColor imgui.Vec4
	if isCurrentSlice {
		col = implot.GetColormapColor(0)
		labelColor = col
	} else {
		colormapSize := implot.GetColormapSize()
		col = implot.GetColormapColor(colormapSize - 1)
		labelColor = col
	}
	flags := implot.DragToolFlagsNone

	draggedStart := newStart
	var outClicked = false

	if implot.DragLineXV(
		int32(idx+1),
		&draggedStart,
		col, 3.0, flags,
		&outClicked, &m.Hovered, &m.Held,
	) {
		changed = true

		// Clamp
		if draggedStart < leftBound {
			draggedStart = leftBound
		}
		if draggedStart > rightBound {
			draggedStart = rightBound
		}

		newStart = draggedStart

		dragLabelColor := imgui.NewVec4(1, 1, 1, 0.90)
		if newStart <= leftBound || newStart >= rightBound {
			dragLabelColor = imgui.NewVec4(1, 0.5, 0, 0.90)
		}
		sec := (newStart * samplesPerBin) / float64(sampleRate)
		implot.AnnotationStr(
			newStart, 0,
			dragLabelColor, imgui.Vec2{X: 0, Y: 0}, true,
			util.SecondsLabel(sec),
		)
	} else {
		if newStart < leftBound {
			newStart = leftBound
		}
		if newStart > rightBound {
			newStart = rightBound
		}

		if newStart != currentStart && !changed {
			changed = true
		}
	}

	// Middle-click on the marker to remove it
	if m.Hovered && !m.Held && imgui.IsMouseClickedBool(imgui.MouseButtonMiddle) {
		wantRemove = true
	}

	// Context Menu
	ctxID := fmt.Sprintf("marker_ctx_%p", m)
	if m.Hovered && imgui.IsMouseClickedBool(imgui.MouseButtonRight) {
		m.Active = true
		imgui.OpenPopupStr(ctxID)
	}

	if imgui.BeginPopup(ctxID) {
		if imgui.MenuItemBool("Remove") {
			wantRemove = true
			m.Active = false
		}
		imgui.EndPopup()
	} else if m.Active && !imgui.IsPopupOpenStrV(ctxID, imgui.PopupFlagsNone) {
		m.Active = false
	}

	// End position for shading
	endBin := rightBound
	if !math.IsInf(nextStart, 1) {
		endBin = nextStart - minGapBins
		if endBin < newStart {
			endBin = newStart
		}
	}

	// Draw shaded region during drag
	if m.Held {
		x1, x2 := newStart, endBin
		if x2 < x1 {
			x1, x2 = x2, x1
		}
		pMin := implot.PlotToPixelsdoubleV(x1, yMin, implot.AxisX1, implot.AxisY1)
		pMax := implot.PlotToPixelsdoubleV(x2, yMax, implot.AxisX1, implot.AxisY1)

		drawList := implot.GetPlotDrawList()
		shadedCol := imgui.Vec4{X: col.X, Y: col.Y, Z: col.Z, W: 0.20}
		fillCol := imgui.ColorConvertFloat4ToU32(shadedCol)
		drawList.AddRectFilled(
			imgui.Vec2{X: pMin.X, Y: pMin.Y},
			imgui.Vec2{X: pMax.X, Y: pMax.Y},
			fillCol,
		)
	}

	// Draw index label
	implot.AnnotationStr(
		newStart, yMax,
		labelColor, imgui.Vec2{X: 0, Y: -1},
		true, fmt.Sprintf("%02d", idx+1),
	)

	return newStart, wantRemove, changed
}
