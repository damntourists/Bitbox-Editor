package component

import (
	"github.com/AllenDang/cimgui-go/imgui"
)

type TreeRowComponent struct {
	*Component
	label    string
	flags    imgui.TreeNodeFlags
	layout   Layout
	children []*TreeRowComponent
}

func (trc *TreeRowComponent) Children(children ...*TreeRowComponent) *TreeRowComponent {
	trc.children = children
	return trc
}

func (trc *TreeRowComponent) Flags(flags imgui.TreeNodeFlags) *TreeRowComponent {
	trc.flags = flags
	return trc
}

func (trc *TreeRowComponent) Layout() {
	imgui.TableNextRowV(0, 0)
	imgui.TableNextColumn()

	var labelText string
	if len(trc.layout) > 0 {
		if textComp, ok := trc.layout[0].(*TextComponent); ok {
			labelText = textComp.text
		}
	}

	if labelText == "" {
		labelText = trc.label
	}

	open := false
	if len(trc.children) > 0 {
		open = imgui.TreeNodeExStrV(labelText, trc.flags)
	} else {
		flags := trc.flags | imgui.TreeNodeFlagsLeaf | imgui.TreeNodeFlagsNoTreePushOnOpen
		imgui.TreeNodeExStrV(labelText, flags)
	}

	for i := 1; i < len(trc.layout); i++ {
		imgui.TableNextColumn()
		trc.layout[i].Layout()
	}

	if len(trc.children) > 0 && open {
		for _, child := range trc.children {
			child.Layout()
		}
		imgui.TreePop()
	}
}

func TreeRow(label string, components ...ComponentType) *TreeRowComponent {
	layout := make([]ComponentType, 0)

	for _, c := range components {
		if textComp, ok := c.(*TextComponent); ok {
			if label == "" {
				label = textComp.text
			}
		}
		layout = append(layout, c)
	}

	return &TreeRowComponent{
		Component: NewComponent(imgui.IDStr(label)),
		label:     label,
		flags:     0,
		layout:    layout,
		children:  make([]*TreeRowComponent, 0),
	}
}
