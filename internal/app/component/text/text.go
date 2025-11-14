package text

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/logging"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

var log = logging.NewLogger("text")

type TextComponent struct {
	*component.Component[*TextComponent]

	font       *imgui.Font
	wrapped    bool
	selectable bool
}

func NewText(text string) *TextComponent {
	cmp := &TextComponent{
		font:       nil,
		wrapped:    false,
		selectable: false,
	}

	cmp.Component = component.NewComponent[*TextComponent](imgui.IDStr(text), cmp.handleUpdate)
	cmp.SetText(text)
	cmp.SetSelected(false)

	cmp.Component.SetLayoutBuilder(cmp)

	return cmp
}

func NewTextWithID(id imgui.ID, text string) *TextComponent {
	cmp := &TextComponent{
		font:       nil,
		wrapped:    false,
		selectable: false,
	}

	cmp.Component = component.NewComponent[*TextComponent](id, cmp.handleUpdate)
	cmp.SetText(text)
	cmp.SetSelected(false)

	cmp.Component.SetLayoutBuilder(cmp)

	return cmp
}

func (tc *TextComponent) handleUpdate(cmd component.UpdateCmd) {
	if tc.Component.HandleGlobalUpdate(cmd) {
		// Handled by base (e.g., CmdSetText, CmdSetSelected)
		return
	}

	switch cmd.Type {
	case cmdSetTextFont:
		// Allow nil font
		if cmd.Data == nil {
			tc.font = nil
		} else if font, ok := cmd.Data.(*imgui.Font); ok {
			tc.font = font
		} else {
			log.Warn("Invalid data type for cmdSetTextFont", zap.Any("data", cmd.Data))
		}

	case cmdSetTextWrapped:
		if wrap, ok := cmd.Data.(bool); ok {
			tc.wrapped = wrap
		}

	case cmdSetTextSelectable:
		if sel, ok := cmd.Data.(bool); ok {
			tc.selectable = sel
		}

	default:
		log.Warn(
			"TextComponent unhandled update",
			zap.String("id", tc.IDStr()),
			zap.Any("cmd", cmd),
		)
	}
}

func (tc *TextComponent) SetWrapped(wrap bool) *TextComponent {
	cmd := component.UpdateCmd{Type: cmdSetTextWrapped, Data: wrap}
	tc.Component.SendUpdate(cmd)
	return tc
}

func (tc *TextComponent) SetFont(font *imgui.Font) *TextComponent {
	cmd := component.UpdateCmd{Type: cmdSetTextFont, Data: font}
	tc.Component.SendUpdate(cmd)
	return tc
}

func (tc *TextComponent) SetSelected(selected bool) *TextComponent {
	tc.Component.SetSelected(selected)
	return tc
}

func (tc *TextComponent) SetSelectable(selectable bool) *TextComponent {
	cmd := component.UpdateCmd{Type: cmdSetTextSelectable, Data: selectable}
	tc.Component.SendUpdate(cmd)
	return tc
}

func (tc *TextComponent) DisableHoverAnimations() *TextComponent {
	// This is now handled by the base component, but we can keep
	// this setter for a clean API.
	// (Assumes CmdSetHoverAnimationsDisabled exists in base component)
	// tc.Component.SendUpdate(component.UpdateCmd{Type: component.CmdSetHoverAnimationsDisabled, Data: true})
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
		imgui.SelectableBoolV(text, selected, flags, imgui.Vec2{})

	} else {
		imgui.TextUnformatted(text)
	}
}
