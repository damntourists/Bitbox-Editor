package component

import (
	"fmt"
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
)

type WaveMarker struct {
	start float64
	end   float64

	// Interaction
	Active     bool
	Hovered    bool
	Held       bool
	WantRemove bool
}

func (m *WaveMarker) StartBin() float64 { return m.start }
func (m *WaveMarker) EndBin() float64   { return m.end }
func (m *WaveMarker) StartSeconds(sampleRate int, samplesPerBin float64) float64 {
	return (m.start * samplesPerBin) / float64(sampleRate)
}
func (m *WaveMarker) EndSeconds(sampleRate int, samplesPerBin float64) float64 {
	return (m.end * samplesPerBin) / float64(sampleRate)
}

// DrawInteract Handles marker interactions from user
func (m *WaveMarker) DrawInteract(
	idx int,
	leftBound,
	rightBound,
	nextStart,
	minGapBins,
	samplesPerBin float64,
	sampleRate int,
	yMin,
	yMax float64,
) {

	col := imgui.NewColor(1, 0, 1, 1).FieldValue
	flags := implot.DragToolFlagsNone

	var oc, oh, held = false, false, false
	if implot.DragLineXV(int32(idx+1), &m.start, col, 3.0, flags, &oc, &oh, &held) {
		// Clamp into neighbor-aware bounds while dragging
		if m.start < leftBound {
			m.start = leftBound
		}
		if m.start > rightBound {
			m.start = rightBound
		}

		// Timestamp label near cursor while dragging
		sec := (m.start * samplesPerBin) / float64(sampleRate)
		implot.AnnotationStr(
			m.start,
			0,
			imgui.NewVec4(1, 0.2, 1, 0.50),
			imgui.Vec2{X: 0, Y: 0},
			true,
			secondsLabel(sec),
		)
	}

	// Persist interaction
	m.Hovered = oh
	m.Held = held

	// Middle-click on the marker to remove it
	if oh && !m.Held && imgui.IsMouseClickedBool(imgui.MouseButtonMiddle) {
		m.WantRemove = true
	}

	// Open the popup based on the marker hover
	ctxID := fmt.Sprintf("marker_ctx_%p", m)
	if oh && imgui.IsMouseClickedBool(imgui.MouseButtonRight) {
		m.Active = true
		imgui.OpenPopupStr(ctxID)
	}

	// Ensure clamping even if no drag this frame
	if m.start < leftBound {
		m.start = leftBound
	}
	if m.start > rightBound {
		m.start = rightBound
	}

	// Update endpoint
	endBin := rightBound
	if !math.IsInf(nextStart, 1) {
		endBin = nextStart - minGapBins
		if endBin < m.start {
			endBin = m.start
		}
	}
	m.end = endBin

	// Shade while dragging
	if m.Held {
		x1, x2 := m.start, m.end
		if x2 < x1 {
			x1, x2 = x2, x1
		}
		pMin := implot.PlotToPixelsdoubleV(x1, yMin, implot.AxisX1, implot.AxisY1)
		pMax := implot.PlotToPixelsdoubleV(x2, yMax, implot.AxisX1, implot.AxisY1)

		drawList := implot.GetPlotDrawList()
		fillCol := imgui.NewColor(1.0, 0.2, 1.0, 0.20).Pack()
		drawList.AddRectFilled(
			imgui.Vec2{X: pMin.X, Y: pMin.Y},
			imgui.Vec2{X: pMax.X, Y: pMax.Y},
			fillCol,
		)
	}

	// Render popup context menu
	if imgui.BeginPopup(ctxID) {
		if imgui.MenuItemBool("Remove") {
			m.WantRemove = true
			m.Active = false
		}
		imgui.EndPopup()
	} else if m.Active && !imgui.IsPopupOpenStrV(ctxID, imgui.PopupFlagsNone) {
		// Clear Active when context menu is closed
		m.Active = false
	}

	// Draw label above marker with marker's identifier
	implot.AnnotationStr(
		m.start,
		yMax,
		imgui.NewVec4(1, 0.2, 1, 1),
		imgui.Vec2{X: 0, Y: 1},
		true,
		fmt.Sprintf("%02d", idx+1),
	)

	return
}

// Bounds computes neighbor-aware clamping limits
func (m *WaveMarker) Bounds(idx int, slices []*WaveMarker, xMin, xMax, minGapBins float64) (left, right float64) {
	left = xMin
	right = xMax
	if idx > 0 {
		left = math.Max(left, slices[idx-1].start+minGapBins)
	}
	if idx < len(slices)-1 {
		right = math.Min(right, slices[idx+1].start-minGapBins)
	}
	if right < left {
		right = left
	}
	return
}
