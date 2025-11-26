package table

import (
	"bitbox-editor/internal/app/component"
	"image/color"

	"github.com/AllenDang/cimgui-go/imgui"
)

type TableRowComponent struct {
	*component.Component[*TableRowComponent]
	component.CommandRouter

	flags        imgui.TableRowFlags
	minRowHeight float64
	layout       component.Layout
}

func NewTableRow(id imgui.ID, components ...component.ComponentType) *TableRowComponent {
	cmp := &TableRowComponent{
		flags:        imgui.TableRowFlagsNone,
		minRowHeight: 0,
	}

	cmp.Component = component.NewComponent[*TableRowComponent](id)
	cmp.SetLayout(components...)
	cmp.Component.SetLayoutBuilder(cmp)

	cmp.CommandRouter.Init(cmp.Component)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTableRowFlags, cmp.onSetFlags)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTableRowMinHeight, cmp.onSetMinHeight)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTableRowLayout, cmp.onSetLayout)

	return cmp
}

func (tr *TableRowComponent) onSetFlags(flags imgui.TableRowFlags) {
	tr.flags = flags
}

func (tr *TableRowComponent) onSetMinHeight(h float64) {
	tr.minRowHeight = h
}

func (tr *TableRowComponent) onSetLayout(l component.Layout) {
	tr.layout = l
}

func (tr *TableRowComponent) SetBgColor(c color.Color) *TableRowComponent {
	vec4Color := component.ToVec4Color(c)
	tr.Component.SetBgColor(vec4Color)
	return tr
}

func (tr *TableRowComponent) SetFlags(flags imgui.TableRowFlags) *TableRowComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableRowFlags, Data: flags}
	tr.SendUpdate(cmd)
	return tr
}

func (tr *TableRowComponent) SetMinHeight(height float64) *TableRowComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableRowMinHeight, Data: height}
	tr.SendUpdate(cmd)
	return tr
}

func (tr *TableRowComponent) SetLayout(components ...component.ComponentType) *TableRowComponent {
	layoutCopy := make(component.Layout, len(components))
	copy(layoutCopy, components)
	cmd := component.UpdateCmd{Type: cmdSetTableRowLayout, Data: layoutCopy}
	tr.SendUpdate(cmd)
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
