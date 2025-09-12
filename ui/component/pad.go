package component

import (
	"bitbox-editor/ui/theme"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
)

type Pad struct {
	*Component

	size     float32 // width=height (square)
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
}

func (p *Pad) Size(px float32) *Pad {
	p.size = px
	return p
}

func (p *Pad) Padding(px float32) *Pad {
	p.padding = px
	return p
}

func (p *Pad) Rounding(px float32) *Pad {
	p.rounding = px
	return p
}

func (p *Pad) Border(px float32) *Pad {
	p.border = px
	return p
}

func (p *Pad) Lines(l1, l2, l3 string) *Pad {
	p.line1, p.line2, p.line3 = l1, l2, l3
	return p
}

func (p *Pad) SetSelected(sel bool) *Pad {
	p.selected = sel
	return p
}

func (p *Pad) Selected() bool { return p.selected }

func (p *Pad) Bg(color imgui.Vec4) *Pad {
	p.bg = color
	return p
}
func (p *Pad) BgHovered(color imgui.Vec4) *Pad {
	p.bgHovered = color
	return p
}
func (p *Pad) BgActive(color imgui.Vec4) *Pad {
	p.bgActive = color
	return p
}
func (p *Pad) BgSelected(color imgui.Vec4) *Pad {
	p.bgSelected = color
	return p
}
func (p *Pad) TextColor(color imgui.Vec4) *Pad {
	p.textColor = color
	return p
}
func (p *Pad) BorderColor(color imgui.Vec4) *Pad {
	p.borderColor = color
	return p
}

// Layout renders the card and emits MouseEvents based on the InvisibleButton
// (the button must be the last item to keep ImGui’s item context).
func (p *Pad) Layout() {
	size := imgui.Vec2{X: p.size, Y: p.size}
	pos := imgui.CursorScreenPos()

	// Hit target
	imgui.InvisibleButton(fmt.Sprintf("##pad-%s", p.Component.IDStr()), size)

	hovered := imgui.IsItemHovered()
	active := imgui.IsItemActive()
	clicked := imgui.IsItemClicked()

	// Toggle selection (or handle externally via MouseEvents listener)
	if clicked {
		p.selected = !p.selected
	}

	// Emit mouse state for listeners
	p.handleMouseEvents()

	// Draw visuals
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

	// Text layout (3 lines max)
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
}

// addTextClipped renders a single-line text clipped to maxWidth with a basic ellipsis.
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

func NewPad(id imgui.ID, l1, l2, l3 string) *Pad {
	cmp := &Pad{
		Component: NewComponent(id),

		size:     72,
		padding:  8,
		rounding: 4,
		border:   1,

		line1: l1,
		line2: l2,
		line3: l3,

		selected: false,
	}

	t := theme.GetCurrentTheme()

	cmp.bg = t.Style.Colors.FrameBg.Vec4
	cmp.bgHovered = t.Style.Colors.HeaderHovered.Vec4
	cmp.bgActive = t.Style.Colors.HeaderActive.Vec4
	cmp.bgSelected = t.Style.Colors.Header.Vec4
	cmp.textColor = t.Style.Colors.Text.Vec4
	cmp.borderColor = t.Style.Colors.Border.Vec4

	cmp.Component.layoutBuilder = cmp
	return cmp
}
