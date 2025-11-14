package text

import (
	"bitbox-editor/internal/app/component"

	"github.com/AllenDang/cimgui-go/imgui"
)

// TODO: Add documentation

type DynamicTextComponent struct {
	*component.Component[*DynamicTextComponent]
	textGetter func() string
	font       *imgui.Font
	wrapped    bool
}

func NewDynamicText(textGetter func() string) *DynamicTextComponent {
	cmp := &DynamicTextComponent{
		textGetter: textGetter,
		font:       nil,
		wrapped:    false,
	}

	cmp.Component = component.NewComponent[*DynamicTextComponent](imgui.ID(0), cmp.handleUpdate)
	cmp.Component.SetLayoutBuilder(cmp)

	return cmp
}

func (dtc *DynamicTextComponent) handleUpdate(cmd component.UpdateCmd) {
	if dtc.Component.HandleGlobalUpdate(cmd) {
		return
	}
	// No specific update handling needed for now
}

func (dtc *DynamicTextComponent) SetFont(font *imgui.Font) *DynamicTextComponent {
	dtc.font = font
	return dtc
}

func (dtc *DynamicTextComponent) SetWrapped(wrap bool) *DynamicTextComponent {
	dtc.wrapped = wrap
	return dtc
}

func (dtc *DynamicTextComponent) Layout() {
	dtc.Component.ProcessUpdates()

	text := dtc.textGetter()

	if dtc.wrapped {
		imgui.PushTextWrapPos()
		defer imgui.PopTextWrapPos()
	}

	if dtc.font != nil {
		imgui.PushFont(dtc.font, 1.0)
		defer imgui.PopFont()
	}

	imgui.TextUnformatted(text)
}
