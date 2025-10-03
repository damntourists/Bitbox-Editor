package component

import (
	"github.com/AllenDang/cimgui-go/imgui"
)

type TextComponent struct {
	*Component
	text       string
	font       *imgui.Font
	wrapped    bool
	selectable bool
	selected   bool
}

func Text(text string) *TextComponent {
	c := NewComponent(imgui.IDStr(text))

	cmp := &TextComponent{
		Component:  c,
		text:       text,
		wrapped:    false,
		selectable: false,
		selected:   false,
	}

	cmp.Component.layoutBuilder = cmp
	return cmp
}

func (tc *TextComponent) Wrapped(wrap bool) *TextComponent {
	tc.wrapped = wrap
	return tc
}

func (tc *TextComponent) Font(font *imgui.Font) *TextComponent {
	tc.font = font
	return tc
}

func (tc *TextComponent) SetSelected(selected bool) *TextComponent {
	tc.selected = selected
	return tc
}

func (tc *TextComponent) Selectable(selectable bool) *TextComponent {
	tc.selectable = selectable
	return tc
}

func (tc *TextComponent) Selected() bool {
	return tc.selected
}

func (tc *TextComponent) Layout() {
	if tc.wrapped {
		imgui.PushTextWrapPos()
		defer imgui.PopTextWrapPos()
	}

	if tc.font != nil {
		imgui.PushFont(tc.font, 1.0)
		defer imgui.PopFont()
	}

	if tc.selectable {
		flags := imgui.SelectableFlagsSpanAllColumns
		if tc.selected {
			flags |= imgui.SelectableFlagsHighlight |
				imgui.SelectableFlagsAllowDoubleClick
		}
		imgui.SelectableBoolV(tc.text, tc.selected, flags, imgui.Vec2{})
	} else {
		imgui.TextUnformatted(tc.text)
	}

	tc.handleMouseEvents()
}
