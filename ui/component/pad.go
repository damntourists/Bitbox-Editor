package component

import (
	"bitbox-editor/lib/audio"
	"bitbox-editor/lib/parsing/bitbox"
	"bitbox-editor/ui/events"
	"bitbox-editor/ui/fonts"
	"bitbox-editor/ui/theme"
	"bitbox-editor/ui/utils"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maniartech/signals"
)

type PadComponent struct {
	*Component

	row, col int

	size     float32
	padding  float32
	rounding float32
	border   float32

	line1 string
	line2 string
	line3 string

	bg              imgui.Vec4
	bgHovered       imgui.Vec4
	bgActive        imgui.Vec4
	bgSelected      imgui.Vec4
	textColor       imgui.Vec4
	borderColor     imgui.Vec4
	borderColorEdit imgui.Vec4

	playProgress       float32
	playProgressHeight float32
	playProgressBg     imgui.Vec4
	playProgressFg     imgui.Vec4

	wave *audio.WaveFile
	cell *bitbox.Cell

	selected bool
	editing  bool

	Events signals.Signal[events.PadEventRecord]
}

func (p *PadComponent) Wav() *audio.WaveFile {
	return p.wave
}
func (p *PadComponent) Selected() bool { return p.selected }
func (p *PadComponent) Progress() float32 {
	return p.playProgress
}
func (p *PadComponent) Row() int { return p.row }
func (p *PadComponent) Col() int { return p.col }

func (p *PadComponent) SetBg(color imgui.Vec4) *PadComponent {
	p.bg = color
	return p
}
func (p *PadComponent) SetBgEdit(color imgui.Vec4) *PadComponent {
	p.borderColorEdit = color
	return p
}
func (p *PadComponent) SetEditing(editing bool) *PadComponent {
	p.editing = editing
	return p
}
func (p *PadComponent) SetProgress(v float32) *PadComponent {
	if v < 0 {
		v = 0
	} else if v > 1 {
		v = 1
	}
	p.playProgress = v
	return p
}
func (p *PadComponent) SetProgressHeight(height float32) *PadComponent {
	p.playProgressHeight = height
	return p
}
func (p *PadComponent) SetCell(cell *bitbox.Cell) *PadComponent {
	p.cell = cell
	return p
}
func (p *PadComponent) SetWave(wave *audio.WaveFile) *PadComponent {
	p.wave = wave
	go func() {
		err := p.wave.Load()
		if err != nil {
			panic(err)

			//log.Error(err.Error())
		}
	}()
	return p
}

func (p *PadComponent) SetProgressBg(bg imgui.Vec4) *PadComponent {
	p.playProgressBg = bg
	return p
}
func (p *PadComponent) SetProgressFg(fg imgui.Vec4) *PadComponent {
	p.playProgressFg = fg
	return p
}
func (p *PadComponent) SetSize(px float32) *PadComponent {
	p.size = px
	return p
}
func (p *PadComponent) SetPadding(px float32) *PadComponent {
	p.padding = px
	return p
}
func (p *PadComponent) SetRounding(px float32) *PadComponent {
	p.rounding = px
	return p
}
func (p *PadComponent) SetBorder(px float32) *PadComponent {
	p.border = px
	return p
}
func (p *PadComponent) SetLines(l1, l2, l3 string) *PadComponent {
	p.line1, p.line2, p.line3 = l1, l2, l3
	return p
}
func (p *PadComponent) SetSelected(sel bool) *PadComponent {
	p.selected = sel
	return p
}
func (p *PadComponent) SetBgHovered(color imgui.Vec4) *PadComponent {
	p.bgHovered = color
	return p
}
func (p *PadComponent) SetBgActive(color imgui.Vec4) *PadComponent {
	p.bgActive = color
	return p
}
func (p *PadComponent) SetBgSelected(color imgui.Vec4) *PadComponent {
	p.bgSelected = color
	return p
}
func (p *PadComponent) SetTextColor(color imgui.Vec4) *PadComponent {
	p.textColor = color
	return p
}
func (p *PadComponent) SetBorderColor(color imgui.Vec4) *PadComponent {
	p.borderColor = color
	return p
}

func (p *PadComponent) Layout() {
	t := theme.GetCurrentTheme()

	size := imgui.Vec2{X: p.size, Y: p.size}
	pos := imgui.CursorScreenPos()

	imgui.SetNextItemAllowOverlap()
	imgui.InvisibleButton(fmt.Sprintf("##pad-%s", p.Component.IDStr()), size)

	hoveredPad := imgui.IsItemHovered()
	active := imgui.IsItemActive()
	clicked := imgui.IsItemClicked()

	borderColor := p.borderColor

	if clicked {
		p.selected = !p.selected
		if p.wave != nil && p.wave.IsLoaded() {
			audio.GetAudioManager().PlayWave(p.wave)
		}
		p.handleMouseEvents()
	}

	if p.wave != nil && p.wave.IsPlaying() {
		p.playProgress = float32(p.wave.Progress())
	} else {
		p.playProgress = 0.0
	}

	draw := imgui.WindowDrawList()

	// Top left & Bottom right corners
	padSizeMin := pos
	padSizeMax := imgui.Vec2{
		X: pos.X + size.X,
		Y: pos.Y + size.Y,
	}

	bg := p.bg
	switch {
	case active:
		bg = p.bgActive
	case hoveredPad:
		bg = p.bgHovered
	}

	if p.wave != nil {
		p.line1 = p.wave.Name
	}

	p.line3 = ""

	if p.Wav() != nil {
		if p.Wav().IsLoading() {
			p.line3 = "loading..."
		}
	}

	draw.AddRectFilledV(padSizeMin, padSizeMax, imgui.ColorU32Vec4(bg), p.rounding, imgui.DrawFlagsNone)
	if p.border > 0 {
		if p.editing {
			borderColor = p.borderColorEdit
		}
		draw.AddRectV(padSizeMin, padSizeMax, imgui.ColorU32Vec4(borderColor), p.rounding, imgui.DrawFlagsNone, p.border)
	}

	icon := fonts.Icon("Pencil")

	iconSize := imgui.CalcTextSizeV(icon, false, 0)
	padding := t.Style.FramePadding
	btnSize := imgui.Vec2{X: iconSize.X + padding[0]*2, Y: iconSize.Y + padding[0]*2}
	btnMin := imgui.Vec2{X: padSizeMax.X - btnSize.X, Y: padSizeMin.Y}
	btnMax := imgui.Vec2{X: btnMin.X + btnSize.X, Y: btnMin.Y + btnSize.Y}

	prev := imgui.CursorScreenPos()
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
				icon = fonts.Icon("PencilOff")
			}
		}

		draw.AddRectFilled(btnMin, btnMax, imgui.ColorU32Vec4(colBg))
		draw.AddRectV(btnMin, btnMax, imgui.ColorU32Vec4(borderColor), 0, imgui.DrawFlagsNone, p.border)

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

	innerX := padSizeMin.X + p.padding
	innerY := padSizeMin.Y + p.padding
	innerW := size.X - 2*p.padding
	lineH := imgui.FontSize()
	lineGap := float32(2)

	draw.PushClipRectV(padSizeMin, padSizeMax, true)
	if p.line1 != "" {
		utils.AddTextClipped(draw, imgui.Vec2{X: innerX, Y: innerY}, innerW, p.line1, p.textColor)
		innerY += lineH + lineGap
	}
	if p.line2 != "" {
		utils.AddTextClipped(draw, imgui.Vec2{X: innerX, Y: innerY}, innerW, p.line2, p.textColor)
		innerY += lineH + lineGap
	}
	if p.line3 != "" {
		utils.AddTextClipped(draw, imgui.Vec2{X: innerX, Y: innerY}, innerW, p.line3, p.textColor)
	}
	draw.PopClipRect()
	if p.playProgressHeight > 0 && p.playProgress > 0 {
		barTop := padSizeMax.Y - p.padding - p.playProgressHeight
		barBottom := padSizeMax.Y - p.padding
		if p.playProgressBg.W > 0 {
			draw.AddRectFilledV(
				imgui.Vec2{X: innerX, Y: barTop},
				imgui.Vec2{X: innerX + innerW, Y: barBottom},
				imgui.ColorU32Vec4(p.playProgressBg), 0, imgui.DrawFlagsNone,
			)
		}
		fillW := innerW * p.playProgress
		if fillW < 1 {
			fillW = 1
		}
		draw.AddRectFilledV(
			imgui.Vec2{X: innerX, Y: barTop},
			imgui.Vec2{X: innerX + fillW, Y: barBottom},
			imgui.ColorU32Vec4(p.playProgressFg), 0, imgui.DrawFlagsNone,
		)
	}

}

func NewPadComponent(id imgui.ID, row, col int, l1, l2, l3 string, size float32) *PadComponent {
	t := theme.GetCurrentTheme()

	cmp := &PadComponent{
		Component: NewComponent(id),

		row: row,
		col: col,

		size:     size,
		padding:  4,
		rounding: 0,
		border:   1,

		line1: l1,
		line2: l2,
		line3: l3,

		playProgress:       0.4,
		playProgressHeight: 4,
		playProgressBg:     t.Style.Colors.ScrollbarBg.Vec4,
		playProgressFg:     t.Style.Colors.HeaderActive.Vec4,

		bg:              t.Style.Colors.FrameBg.Vec4,
		bgHovered:       t.Style.Colors.HeaderHovered.Vec4,
		bgActive:        t.Style.Colors.HeaderActive.Vec4,
		bgSelected:      t.Style.Colors.HeaderActive.Vec4,
		textColor:       t.Style.Colors.Text.Vec4,
		borderColor:     t.Style.Colors.Border.Vec4,
		borderColorEdit: t.Style.Colors.Text.Vec4,

		selected: false,
		editing:  false,
	}

	cmp.Component.layoutBuilder = cmp
	return cmp
}
