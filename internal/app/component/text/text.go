package text

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/logging"

	"github.com/AllenDang/cimgui-go/imgui"
)

var log = logging.NewLogger("text")

type TextComponent struct {
	*component.Component[*TextComponent]
	component.CommandRouter

	font       *imgui.Font
	wrapped    bool
	selectable bool
}

func NewText(text string) *TextComponent {
	return NewTextWithID(imgui.IDStr(text), text)
}

func NewTextWithID(id imgui.ID, text string) *TextComponent {
	cmp := &TextComponent{
		font:       nil,
		wrapped:    false,
		selectable: false,
	}

	cmp.Component = component.NewComponent[*TextComponent](id)
	cmp.SetText(text)
	cmp.SetSelected(false)

	cmp.Component.SetLayoutBuilder(cmp)

	cmp.CommandRouter.Init(cmp.Component)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTextFont, cmp.onSetFont)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTextWrapped, cmp.onSetWrapped)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTextSelectable, cmp.onSetSelectable)

	// Process initial updates immediately so properties are available for first render
	// This is needed for text components that are created and immediately rendered
	cmp.Component.ProcessUpdates()

	return cmp
}

func (tc *TextComponent) onSetFont(font *imgui.Font) {
	tc.font = font
}

func (tc *TextComponent) onSetWrapped(wrap bool) {
	tc.wrapped = wrap
}

func (tc *TextComponent) onSetSelectable(sel bool) {
	tc.selectable = sel
}

func (tc *TextComponent) SetWrapped(wrap bool) *TextComponent {
	cmd := component.UpdateCmd{Type: cmdSetTextWrapped, Data: wrap}
	tc.SendUpdate(cmd)
	return tc
}

func (tc *TextComponent) SetFont(font *imgui.Font) *TextComponent {
	cmd := component.UpdateCmd{Type: cmdSetTextFont, Data: font}
	tc.SendUpdate(cmd)
	return tc
}

func (tc *TextComponent) SetSelected(selected bool) *TextComponent {
	tc.Component.SetSelected(selected)
	return tc
}

func (tc *TextComponent) SetSelectable(selectable bool) *TextComponent {
	cmd := component.UpdateCmd{Type: cmdSetTextSelectable, Data: selectable}
	tc.SendUpdate(cmd)
	return tc
}

func (tc *TextComponent) DisableHoverAnimations() *TextComponent {
	return tc
}

func (tc *TextComponent) SetText(text string) *TextComponent {
	tc.Component.SetText(text)
	return tc
}

func (tc *TextComponent) Selected() bool {
	return tc.Component.Selected()
}

func (tc *TextComponent) Text() string {
	return tc.Component.Text()
}

func (tc *TextComponent) Layout() {
	tc.Component.ProcessUpdates()

	text := tc.Component.Text()
	selected := tc.Component.Selected()

	font := tc.font
	wrapped := tc.wrapped
	selectable := tc.selectable

	if wrapped {
		imgui.PushTextWrapPos()
		defer imgui.PopTextWrapPos()
	}

	if font != nil {
		imgui.PushFont(font, 1.0)
		defer imgui.PopFont()
	}

	if selectable {
		flags := imgui.SelectableFlagsSpanAllColumns
		if imgui.SelectableBoolV(text, selected, flags, imgui.Vec2{}) {
			tc.Component.SetSelected(!selected)
		}
	} else {
		imgui.TextUnformatted(text)
	}
}
