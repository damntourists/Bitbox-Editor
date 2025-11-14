package pad

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/app/font"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/audio"
	"bitbox-editor/internal/logging"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

// TODO: Move to commands file
type localCommand int // This type is private to the 'pad' package

const (
	cmdSetPadTextLines localCommand = iota
	cmdSetPadWaveDisplayData
	cmdSetPadCellDisplayData
)

var log = logging.NewLogger("pad")

type PadCellDisplayData struct {
	// TODO: cell data
	Param1 string
	Param2 int
}

type PadComponent struct {
	*component.Component[*PadComponent]

	row, col int
	line1    string
	line2    string
	line3    string

	waveDisplayData audio.WaveDisplayData
	cellDisplayData PadCellDisplayData

	editing bool
}

func NewPad(id imgui.ID, row, col int, size float32) *PadComponent {
	t := theme.GetCurrentTheme()

	cmp := &PadComponent{
		row:             row,
		col:             col,
		editing:         false,
		waveDisplayData: audio.WaveDisplayData{},
		cellDisplayData: PadCellDisplayData{},
	}
	cmp.Component = component.NewComponent[*PadComponent](id, cmp.handleUpdate)
	cmp.SetWidth(size)
	// TODO: Update base padding to use vec2
	cmp.SetPadding(t.Style.FramePadding[0])
	cmp.SetRounding(t.Style.FrameRounding)
	cmp.SetBorderSize(1)

	cmp.SetBgColor(t.Style.Colors.FrameBg.Vec4)
	cmp.SetBgHoveredColor(t.Style.Colors.HeaderHovered.Vec4)
	cmp.SetBgActiveColor(t.Style.Colors.HeaderActive.Vec4)
	cmp.SetBgSelectedColor(t.Style.Colors.HeaderActive.Vec4)
	cmp.SetTextColor(t.Style.Colors.Text.Vec4)
	cmp.SetBorderColor(t.Style.Colors.Border.Vec4)

	cmp.SetProgressHeight(4)
	cmp.SetProgressBgColor(t.Style.Colors.ScrollbarBg.Vec4)
	cmp.SetProgressFgColor(t.Style.Colors.HeaderActive.Vec4)

	cmp.SetSelected(false)
	cmp.SetEditing(false)
	cmp.SetClickable(true)

	cmp.Component.SetLayoutBuilder(cmp)
	return cmp
}

func (p *PadComponent) handleUpdate(cmd component.UpdateCmd) {
	handled := p.Component.HandleGlobalUpdate(cmd)

	switch cmd.Type {
	case cmdSetPadTextLines:
		if lines, ok := cmd.Data.([3]string); ok {
			p.line1 = lines[0]
			p.line2 = lines[1]
			p.line3 = lines[2]
		}
	case cmdSetPadWaveDisplayData:
		if data, ok := cmd.Data.(audio.WaveDisplayData); ok {
			//log.Debug("Pad received wave display data update",
			//	zap.String("padID", p.IDStr()),
			//	zap.String("path", data.Path),
			//	zap.String("filename", data.Name),
			//	zap.Bool("isReady", data.IsReady))
			p.waveDisplayData = data
		}
	case cmdSetPadCellDisplayData:
		if data, ok := cmd.Data.(PadCellDisplayData); ok {
			p.cellDisplayData = data
		}
	default:
		if !handled {
			log.Warn(
				"PadComponent unhandled update",
				zap.String("id", p.IDStr()),
				zap.Any("cmd", cmd),
			)
		}
	}
}

func (p *PadComponent) Row() int {
	return p.row
}

func (p *PadComponent) Col() int {
	return p.col
}

func (p *PadComponent) IsEditing() bool {
	return p.editing
}

func (p *PadComponent) GetWavePath() string {
	path := p.waveDisplayData.Path
	if path == "" {
		log.Warn("GetWavePath called but path is empty",
			zap.String("padID", p.IDStr()),
			zap.String("name", p.waveDisplayData.Name),
			zap.Bool("isReady", p.waveDisplayData.IsReady))
	}
	return path
}
func (p *PadComponent) SetTextLines(l1, l2, l3 string) *PadComponent {
	lines := [3]string{l1, l2, l3}
	p.Component.SendUpdate(component.UpdateCmd{Type: cmdSetPadTextLines, Data: lines})
	return p
}

func (p *PadComponent) SetWaveDisplayData(data audio.WaveDisplayData) *PadComponent {
	p.Component.SendUpdate(component.UpdateCmd{Type: cmdSetPadWaveDisplayData, Data: data})
	return p
}

func (p *PadComponent) GetWaveDisplayData() audio.WaveDisplayData {
	return p.waveDisplayData
}

func (p *PadComponent) SetCellDisplayData(data PadCellDisplayData) *PadComponent {
	p.Component.SendUpdate(component.UpdateCmd{Type: cmdSetPadCellDisplayData, Data: data})
	return p
}

func (p *PadComponent) Layout() {
	p.Component.ProcessUpdates()

	t := theme.GetCurrentTheme()

	var size imgui.Vec2
	if p.Component.Width() > 0 {
		size = imgui.Vec2{X: p.Component.Width(), Y: p.Component.Width()}
	} else {
		size = p.Component.Size()
	}

	padding := p.Component.Padding()
	rounding := p.Component.Rounding()
	borderSize := p.Component.BorderSize()
	// Use the animated color
	bgColor := p.Component.GetAnimatedBgColor()
	bgHoveredColor := p.Component.BgHoverColor()
	bgActiveColor := p.Component.BgActiveColor()
	bgSelectedColor := p.Component.BgSelectedColor()
	textColor := p.Component.TextColor()
	borderColor := p.Component.BorderColor()
	progress := p.Component.Progress()
	progressHeight := p.Component.ProgressHeight()
	progressBg := p.Component.ProgressBg()
	progressFg := p.Component.ProgressFg()
	baseSelected := p.Component.Selected()
	pos := imgui.CursorScreenPos()

	imgui.SetNextItemAllowOverlap()
	imgui.InvisibleButton(fmt.Sprintf("##pad-%s", p.Component.IDStr()), size)

	hoveredPad := imgui.IsItemHovered()
	active := imgui.IsItemActive()
	clicked := imgui.IsItemClicked()

	// Set hand cursor when hovering over the pad
	if hoveredPad {
		imgui.SetMouseCursor(imgui.MouseCursorHand)
	}

	if clicked {
		eventbus.Bus.Publish(events.MouseEventRecord{
			EventType: events.ComponentClickedEvent,
			ImguiID:   p.ID(),
			UUID:      p.UUID(),
			Button:    events.MouseButtonLeft,
			State:     p.State(),
			// Send a reference to this pad
			Data: p,
		})
	}

	line1 := p.line1
	line2 := p.line2
	line3 := p.line3
	if p.waveDisplayData.IsReady {
		line1 = p.waveDisplayData.Name
		if p.waveDisplayData.IsLoading {
			line3 = "loading..."
		} else if p.waveDisplayData.IsPlaying {
			line3 = fmt.Sprintf("%.2f%%", p.waveDisplayData.Progress*100)
		}
	} else {
		// clear lines if no wav data
		line1 = ""
		line3 = ""
	}

	draw := imgui.WindowDrawList()

	// Top left & Bottom right corners
	padSizeMin := pos
	padSizeMax := imgui.Vec2{
		X: pos.X + size.X,
		Y: pos.Y + size.Y,
	}

	bg := bgColor
	if baseSelected {
		bg = bgSelectedColor
	}
	if active {
		bg = bgActiveColor
	} else if hoveredPad {
		bg = bgHoveredColor
	}
	p.line3 = ""

	draw.AddRectFilledV(padSizeMin, padSizeMax, imgui.ColorU32Vec4(bg), rounding, imgui.DrawFlagsNone)

	currentBorderColor := borderColor
	if p.editing {
		currentBorderColor = t.Style.Colors.Text.Vec4
	}

	if borderSize > 0 {
		draw.AddRectV(
			padSizeMin,
			padSizeMax,
			imgui.ColorU32Vec4(currentBorderColor),
			rounding,
			imgui.DrawFlagsNone,
			borderSize,
		)
	}

	icon := font.Icon("Pencil")
	prev := imgui.CursorScreenPos()

	iconSize := imgui.CalcTextSizeV(icon, false, 0)
	btnSize := imgui.Vec2{X: iconSize.X + padding*2, Y: iconSize.Y + padding*2}
	btnMin := imgui.Vec2{X: padSizeMax.X - btnSize.X, Y: padSizeMin.Y}
	btnMax := imgui.Vec2{X: btnMin.X + btnSize.X, Y: btnMin.Y + btnSize.Y}

	imgui.SetNextItemAllowOverlap()
	imgui.SetCursorScreenPos(btnMin)
	cornerClicked := imgui.InvisibleButton(fmt.Sprintf("##pad-corner-%s", p.Component.IDStr()), btnSize)
	cornerHovered := imgui.IsItemHovered()
	imgui.SetCursorScreenPos(prev)

	// Corner "Edit" button
	if cornerClicked {
		p.editing = !p.editing
	}

	if hoveredPad || cornerHovered || p.editing {

		colBg := t.Style.Colors.PopupBg.Vec4
		colBg.W = 0.7

		colIcon := t.Style.Colors.Text.Vec4
		colIcon.W = 0.9
		if cornerHovered {
			colIcon.W = 1.0
			colBg = t.Style.Colors.HeaderActive.Vec4
			if p.editing {
				icon = font.Icon("PencilOff")
			}
		}

		draw.AddRectFilled(btnMin, btnMax, imgui.ColorU32Vec4(colBg))
		draw.AddRectV(btnMin, btnMax, imgui.ColorU32Vec4(borderColor), 0, imgui.DrawFlagsNone, borderSize)

		center := imgui.Vec2{
			X: (btnMin.X + btnMax.X) * 0.5,
			Y: (btnMin.Y + btnMax.Y) * 0.5,
		}
		iconPos := imgui.Vec2{
			X: center.X - iconSize.X*0.5,
			Y: center.Y - iconSize.Y*0.5,
		}
		draw.AddTextVec2(iconPos, imgui.ColorU32Vec4(colIcon), icon)
	}

	innerX := padSizeMin.X + padding
	innerY := padSizeMin.Y + padding
	innerW := size.X - 2*padding
	lineH := imgui.FontSize()
	lineGap := float32(2)

	draw.PushClipRectV(padSizeMin, padSizeMax, true)
	if line1 != "" {
		component.AddTextClipped(draw, imgui.Vec2{X: innerX, Y: innerY}, innerW, line1, textColor)
		innerY += lineH + lineGap
	}
	if line2 != "" {
		component.AddTextClipped(draw, imgui.Vec2{X: innerX, Y: innerY}, innerW, line2, textColor)
		innerY += lineH + lineGap
	}
	if line3 != "" {
		component.AddTextClipped(draw, imgui.Vec2{X: innerX, Y: innerY}, innerW, line3, textColor)
	}
	draw.PopClipRect()

	// TODO: Fix progress bar.
	// Draw progress bar
	if progressHeight > 0 && progress > 0 {
		barTop := padSizeMax.Y - padding - progressHeight
		barBottom := padSizeMax.Y - padding
		if progressBg.W > 0 {
			draw.AddRectFilledV(
				imgui.Vec2{X: innerX, Y: barTop},
				imgui.Vec2{X: innerX + innerW, Y: barBottom},
				imgui.ColorU32Vec4(progressBg), 0, imgui.DrawFlagsNone,
			)
		}
		fillW := innerW * progress
		if fillW < 1 {
			fillW = 1
		}
		draw.AddRectFilledV(
			imgui.Vec2{X: innerX, Y: barTop},
			imgui.Vec2{X: innerX + fillW, Y: barBottom},
			imgui.ColorU32Vec4(progressFg), 0, imgui.DrawFlagsNone,
		)
	}

	if imgui.BeginDragDropTarget() {
		defer imgui.EndDragDropTarget()

		// TODO: Come up with a better dropDataType
		//draggedData, success := p.Component.HandleDropTarget("bitbox_wave")

		// TODO: Emit event or send command to handle the drop data
		//if success {
		// eventbus.Bus.Publish(events.PadDropEvent{...})
		//}
	}

}
