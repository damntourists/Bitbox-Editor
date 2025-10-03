package component

import (
	"bitbox-editor/ui/theme"
	"strconv"

	"github.com/AllenDang/cimgui-go/imgui"
)

type LabelComponent struct {
	*Component

	text string

	bgColor      imgui.Vec4
	fgColor      imgui.Vec4
	outlineColor imgui.Vec4

	padding      float32
	rounding     float32
	outline      bool
	outlineWidth float32
}

func (lc *LabelComponent) SetText(text string) *LabelComponent {
	lc.text = text
	return lc
}

func (lc *LabelComponent) SetBgColor(color imgui.Vec4) *LabelComponent {
	lc.bgColor = color
	return lc
}

func (lc *LabelComponent) SetFgColor(color imgui.Vec4) *LabelComponent {
	lc.fgColor = color
	return lc
}

func (lc *LabelComponent) SetOutlineColor(color imgui.Vec4) *LabelComponent {
	lc.outlineColor = color
	return lc
}

func (lc *LabelComponent) SetPadding(padding float32) *LabelComponent {
	lc.padding = padding
	return lc
}

func (lc *LabelComponent) SetRounding(rounding float32) *LabelComponent {
	lc.rounding = rounding
	return lc
}

func (lc *LabelComponent) SetOutlineWidth(outlineWidth float32) *LabelComponent {
	lc.outlineWidth = outlineWidth
	return lc
}

func (lc *LabelComponent) SetOutline(outline bool) *LabelComponent {
	lc.outline = outline
	return lc
}

func (lc *LabelComponent) Layout() {
	textSize := imgui.CalcTextSize(lc.text)

	size := imgui.Vec2{
		X: textSize.X + lc.padding*2,
		Y: textSize.Y + lc.padding*2,
	}

	framePad := imgui.Vec2{
		X: theme.GetCurrentTheme().Style.FramePadding[0],
		Y: theme.GetCurrentTheme().Style.FramePadding[1],
	}

	pos := imgui.Vec2{
		X: imgui.CursorScreenPos().X + (framePad.X / 2),
		Y: imgui.CursorScreenPos().Y + (framePad.Y / 2),
	}

	imgui.InvisibleButtonV(strconv.Itoa(int(lc.id)), size, imgui.ButtonFlagsNone)

	dl := imgui.WindowDrawList()
	maxBounds := imgui.Vec2{
		X: pos.X + size.X + (lc.padding),
		Y: pos.Y + size.Y + (lc.padding),
	}

	if lc.rounding <= 0 {
		// Default pill shape
		lc.rounding = size.X / 2
	}

	dl.AddRectFilledV(
		pos,
		maxBounds,
		imgui.ColorU32Vec4(lc.bgColor),
		lc.rounding,
		imgui.DrawFlagsNone,
	)
	if lc.outline {
		dl.AddRectV(
			pos,
			maxBounds,
			imgui.ColorU32Vec4(lc.outlineColor),
			lc.rounding,
			imgui.DrawFlagsNone,
			lc.outlineWidth,
		)
	}

	dl.AddTextVec2(
		imgui.Vec2{
			X: pos.X + (size.X-textSize.X)/2 + (lc.padding / 2),
			Y: pos.Y + (size.Y-textSize.Y)/2 + (lc.padding / 2),
		},
		imgui.ColorU32Vec4(lc.fgColor),
		lc.text,
	)
}

func NewLabelComponent(text string) *LabelComponent {
	cmp := &LabelComponent{
		Component: NewComponent(imgui.IDStr(text)),
		text:      text,

		fgColor: theme.GetCurrentTheme().Style.Colors.Text.Vec4,
		bgColor: theme.GetCurrentTheme().Style.Colors.HeaderActive.Vec4,

		padding:      2,
		rounding:     2,
		outlineWidth: 1,
		outline:      true,
		outlineColor: theme.GetCurrentTheme().Style.Colors.Border.Vec4,
	}

	return cmp
}
