package component

import (
	"github.com/AllenDang/cimgui-go/imgui"
)

type TextComponent struct {
	*Component
	text    string
	font    *imgui.Font
	wrapped bool
}

func Text(text string) *TextComponent {
	cmp := NewComponent(ID(text))

	return &TextComponent{
		Component: cmp,
		text:      text,
		wrapped:   false,
	}
}

func (tc *TextComponent) Wrapped(wrap bool) *TextComponent {
	tc.wrapped = wrap
	return tc
}

func (tc *TextComponent) Font(font *imgui.Font) *TextComponent {
	tc.font = font
	return tc
}

func (tc *TextComponent) Layout() {
	if tc.wrapped {
		imgui.PushTextWrapPos()
		defer imgui.PopTextWrapPos()
	}

	if tc.font != nil {
		imgui.PushFont(tc.font)
		defer imgui.PopFont()
	}

	imgui.TextUnformatted(tc.text)
}
