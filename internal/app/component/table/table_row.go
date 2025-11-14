package table

import (
	"bitbox-editor/internal/app/component"
	"image/color"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

type TableRowComponent struct {
	*component.Component[*TableRowComponent]

	flags        imgui.TableRowFlags
	minRowHeight float64
	layout       component.Layout
}

func NewTableRow(id imgui.ID, components ...component.ComponentType) *TableRowComponent {
	cmp := &TableRowComponent{
		flags:        imgui.TableRowFlagsNone,
		minRowHeight: 0,
	}

	cmp.Component = component.NewComponent[*TableRowComponent](id, cmp.handleUpdate)
	cmp.SetLayout(components...)
	cmp.Component.SetLayoutBuilder(cmp)

	return cmp
}

func (tr *TableRowComponent) handleUpdate(cmd component.UpdateCmd) {
	if tr.Component.HandleGlobalUpdate(cmd) {
		return
	}

	switch cmd.Type {
	case cmdSetTableRowFlags:
		if flags, ok := cmd.Data.(imgui.TableRowFlags); ok {
			tr.flags = flags
		}

	case cmdSetTableRowMinHeight:
		if h, ok := cmd.Data.(float64); ok {
			tr.minRowHeight = h
		}

	case cmdSetTableRowLayout:
		if l, ok := cmd.Data.(component.Layout); ok {
			tr.layout = l
		} else if cmd.Data == nil {
			tr.layout = nil
		}

	default:
		log.Warn(
			"TableRowComponent unhandled update",
			zap.String("id", tr.IDStr()),
			zap.Any("cmd", cmd),
		)
	}
}

func (tr *TableRowComponent) SetBgColor(c color.Color) *TableRowComponent {
	vec4Color := component.ToVec4Color(c)
	tr.Component.SetBgColor(vec4Color)
	return tr
}

func (tr *TableRowComponent) SetFlags(flags imgui.TableRowFlags) *TableRowComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableRowFlags, Data: flags}
	tr.Component.SendUpdate(cmd)
	return tr
}

func (tr *TableRowComponent) SetMinHeight(height float64) *TableRowComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableRowMinHeight, Data: height}
	tr.Component.SendUpdate(cmd)
	return tr
}

func (tr *TableRowComponent) SetLayout(components ...component.ComponentType) *TableRowComponent {
	layoutCopy := make(component.Layout, len(components))
	copy(layoutCopy, components)
	cmd := component.UpdateCmd{Type: cmdSetTableRowLayout, Data: layoutCopy}
	tr.Component.SendUpdate(cmd)
	return tr
}

func (tr *TableRowComponent) Layout() {
	tr.Component.ProcessUpdates()
	flags := tr.flags
	minHeight := tr.minRowHeight
	layout := tr.layout
	bgColor := tr.Component.GetAnimatedBgColor()

	imgui.TableNextRowV(flags, float32(minHeight))

	for i, c := range layout {
		if c == nil {
			continue
		}

		isOverlayComponent := false

		if !isOverlayComponent {
			imgui.TableSetColumnIndex(int32(i))
		}

		c.Build()
	}

	if bgColor.W > 0 {
		imgui.TableSetBgColorV(
			imgui.TableBgTargetRowBg0,
			imgui.ColorU32Vec4(bgColor),
			-1,
		)
	}

}

// Destroy cleans up the component and all its children
func (tr *TableRowComponent) Destroy() {
	for _, c := range tr.layout {
		if c != nil {
			c.Destroy()
		}
	}
	tr.layout = nil
	tr.Component.Destroy()
}
