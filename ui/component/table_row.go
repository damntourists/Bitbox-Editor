package component

import (
	"image/color"

	"github.com/AllenDang/cimgui-go/imgui"
)

type TableRowComponent struct {
	*Component
	flags        imgui.TableRowFlags
	minRowHeight float64
	layout       Layout
	bgColor      color.Color
}

func (tr *TableRowComponent) BgColor(c color.Color) *TableRowComponent {
	tr.bgColor = c
	return tr
}

func (tr *TableRowComponent) Flags(flags imgui.TableRowFlags) *TableRowComponent {
	tr.flags = flags
	return tr
}

func (tr *TableRowComponent) MinHeight(height float64) *TableRowComponent {
	tr.minRowHeight = height
	return tr
}

func (tr *TableRowComponent) Layout() {
	imgui.TableNextRowV(tr.flags, float32(tr.minRowHeight))
	for i, c := range tr.layout {
		switch c.(type) {
		//case *ToolTipComponent, *ContextMenuComponent, *PopupModalComponent:
		//noop
		default:
			imgui.TableSetColumnIndex(int32(i))
		}

		c.Layout()
	}

	if tr.bgColor != nil {
		imgui.TableSetBgColorV(
			imgui.TableBgTargetRowBg0,
			imgui.ColorU32Vec4(ToVec4Color(tr.bgColor)),
			-1,
		)
	}

}

func TableRow(id imgui.ID, components ...ComponentType) *TableRowComponent {
	trc := &TableRowComponent{
		Component:    NewComponent(id),
		flags:        imgui.TableRowFlagsNone,
		minRowHeight: 0,
		layout:       components,
		bgColor:      nil,
	}
	//trc.Component.layoutBuilder = trc
	return trc
}
