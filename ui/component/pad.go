package component

import (
	"bitbox-editor/lib/audio"
	"bitbox-editor/ui/theme"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
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

	selected bool

	bg          imgui.Vec4
	bgHovered   imgui.Vec4
	bgActive    imgui.Vec4
	bgSelected  imgui.Vec4
	textColor   imgui.Vec4
	borderColor imgui.Vec4

	playProgress       float32
	playProgressHeight float32
	playProgressBg     imgui.Vec4
	playProgressFg     imgui.Vec4

	wave *audio.WaveFile

	loading bool
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

func (p *PadComponent) SetWave(wave *audio.WaveFile) *PadComponent {
	p.wave = wave
	return p
}

func (p *PadComponent) SetLoading(loading bool) *PadComponent {
	p.loading = loading
	return p
}

func (p *PadComponent) Progress() float32 {
	return p.playProgress
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

func (p *PadComponent) Selected() bool { return p.selected }

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
	size := imgui.Vec2{X: p.size, Y: p.size}
	pos := imgui.CursorScreenPos()

	imgui.InvisibleButton(fmt.Sprintf("##pad-%s", p.Component.IDStr()), size)

	hovered := imgui.IsItemHovered()
	active := imgui.IsItemActive()
	clicked := imgui.IsItemClicked()

	if clicked {
		p.selected = !p.selected
	}
	p.handleMouseEvents()

	draw := imgui.WindowDrawList()
	min := pos
	max := imgui.Vec2{X: pos.X + size.X, Y: pos.Y + size.Y}

	bg := p.bg
	switch {
	case active:
		bg = p.bgActive
	case hovered:
		bg = p.bgHovered
	}
	if p.selected {
		bg = p.bgSelected
	}
	draw.AddRectFilledV(min, max, imgui.ColorU32Vec4(bg), p.rounding, imgui.DrawFlagsNone)
	if p.border > 0 {
		draw.AddRectV(min, max, imgui.ColorU32Vec4(p.borderColor), p.rounding, imgui.DrawFlagsNone, p.border)
	}

	innerX := min.X + p.padding
	innerY := min.Y + p.padding
	innerW := size.X - 2*p.padding
	lineH := imgui.FontSize()
	lineGap := float32(2)

	draw.PushClipRectV(min, max, true)
	if p.line1 != "" {
		addTextClipped(draw, imgui.Vec2{X: innerX, Y: innerY}, innerW, p.line1, p.textColor)
		innerY += lineH + lineGap
	}
	if p.line2 != "" {
		addTextClipped(draw, imgui.Vec2{X: innerX, Y: innerY}, innerW, p.line2, p.textColor)
		innerY += lineH + lineGap
	}
	if p.line3 != "" {
		addTextClipped(draw, imgui.Vec2{X: innerX, Y: innerY}, innerW, p.line3, p.textColor)
	}
	draw.PopClipRect()
	if p.playProgressHeight > 0 && p.playProgress > 0 {
		barTop := max.Y - p.padding - p.playProgressHeight
		barBottom := max.Y - p.padding
		// Track (background)
		if p.playProgressBg.W > 0 {
			draw.AddRectFilledV(
				imgui.Vec2{X: innerX, Y: barTop},
				imgui.Vec2{X: innerX + innerW, Y: barBottom},
				imgui.ColorU32Vec4(p.playProgressBg), 0, imgui.DrawFlagsNone,
			)
		}
		// Filled portion
		fillW := innerW * p.playProgress
		if fillW < 1 {
			fillW = 1 // ensure at least 1px when >0
		}
		draw.AddRectFilledV(
			imgui.Vec2{X: innerX, Y: barTop},
			imgui.Vec2{X: innerX + fillW, Y: barBottom},
			imgui.ColorU32Vec4(p.playProgressFg), 0, imgui.DrawFlagsNone,
		)
	}

}

// addTextClipped adds text with truncation.
func addTextClipped(draw *imgui.DrawList, pos imgui.Vec2, maxWidth float32, text string, color imgui.Vec4) {
	if text == "" {
		return
	}
	ts := imgui.CalcTextSizeV(text, false, 0)
	if ts.X <= maxWidth {

		draw.AddTextVec2(pos, imgui.ColorU32Vec4(color), text)
		return
	}
	ellipsis := "..."
	ellipsisW := imgui.CalcTextSizeV(ellipsis, false, 0).X

	left, right := 0, len(text)
	for left < right {
		mid := (left + right) / 2
		sub := text[:mid]
		w := imgui.CalcTextSizeV(sub, false, 0).X + ellipsisW
		if w <= maxWidth {
			left = mid + 1
		} else {
			right = mid
		}
	}
	if right <= 0 {
		draw.AddTextVec2(pos, imgui.ColorU32Vec4(color), ellipsis)
		return
	}
	draw.AddTextVec2(pos, imgui.ColorU32Vec4(color), text[:right-1]+ellipsis)
}

func NewPadComponent(id imgui.ID, row, col int, l1, l2, l3 string) *PadComponent {
	t := theme.GetCurrentTheme()

	cmp := &PadComponent{
		Component: NewComponent(id),

		row: row,
		col: col,

		size:     72,
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

		bg:          t.Style.Colors.FrameBg.Vec4,
		bgHovered:   t.Style.Colors.HeaderHovered.Vec4,
		bgActive:    t.Style.Colors.HeaderActive.Vec4,
		bgSelected:  t.Style.Colors.Header.Vec4,
		textColor:   t.Style.Colors.Text.Vec4,
		borderColor: t.Style.Colors.Border.Vec4,

		selected: false,
	}

	cmp.Component.layoutBuilder = cmp
	return cmp
}
