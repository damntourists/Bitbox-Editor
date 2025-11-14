package tree

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/component/text"
	"bitbox-editor/internal/app/dragdrop"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
)

type treeRowBuildData struct {
	Path         string
	Name         string
	Icon         string
	IsDir        bool
	IsAudio      bool
	DurationText string
	SizeText     string
	DragDropType string
	Children     []treeRowBuildData
}

type TreeRowData = treeRowBuildData

type TreeRowComponent struct {
	*component.Component[*TreeRowComponent]
	label    string
	flags    imgui.TreeNodeFlags
	layout   component.Layout
	children []*TreeRowComponent

	childrenData []treeRowBuildData
	factoryFn    func(data []TreeRowData) []*TreeRowComponent
	isPopulated  bool

	dragDropData      interface{}
	dragDropType      string
	dragDropTooltipFn dragdrop.TooltipFunc
}

func NewTreeRow(id imgui.ID, components ...component.ComponentType) *TreeRowComponent {
	layout := make([]component.ComponentType, 0)
	var label string

	for _, c := range components {
		if textComp, ok := c.(*text.TextComponent); ok {
			if label == "" {
				label = textComp.Text()
			}
		}
		layout = append(layout, c)
	}

	cmp := &TreeRowComponent{
		label:    label,
		flags:    0,
		layout:   layout,
		children: make([]*TreeRowComponent, 0),
	}
	cmp.Component = component.NewComponent[*TreeRowComponent](id, nil)
	cmp.Component.SetLayoutBuilder(cmp)
	return cmp
}

func (trc *TreeRowComponent) SetChildrenData(
	data []TreeRowData,
	factory func(data []TreeRowData) []*TreeRowComponent,
) *TreeRowComponent {
	trc.childrenData = data
	trc.factoryFn = factory
	trc.isPopulated = false
	trc.children = nil
	return trc
}
func (trc *TreeRowComponent) SetDragDropData(dataType string, data interface{}) *TreeRowComponent {
	trc.Component.SetDragDropData(dataType, data)
	trc.dragDropData = data
	trc.dragDropType = dataType
	return trc
}

func (trc *TreeRowComponent) SetDragDropTooltipFn(fn dragdrop.TooltipFunc) *TreeRowComponent {
	trc.Component.SetDragDropTooltipFn(fn)
	trc.dragDropTooltipFn = fn
	return trc
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
	trc.Component.ProcessUpdates()

	imgui.PushIDInt(int32(trc.Component.ID()))
	defer imgui.PopID()

	imgui.TableNextColumn()

	var labelText string
	if len(trc.layout) > 0 {
		if textComp, ok := trc.layout[0].(*text.TextComponent); ok {
			textComp.ProcessUpdates()
			labelText = textComp.Text()
		}
	}

	if labelText == "" {
		labelText = trc.label
	}
	open := false
	isLeaf := len(trc.children) == 0 && trc.childrenData == nil

	if isLeaf {
		flags := trc.flags | imgui.TreeNodeFlagsLeaf |
			imgui.TreeNodeFlagsNoTreePushOnOpen |
			imgui.TreeNodeFlagsFramePadding
		imgui.TreeNodeExStrV(labelText, flags)
	} else {
		open = imgui.TreeNodeExStrV(
			labelText,
			imgui.TreeNodeFlagsFramePadding|trc.flags,
		)
	}

	if imgui.IsItemHovered() {
		imgui.SetMouseCursor(imgui.MouseCursorHand)
	}

	if trc.dragDropData != nil {
		if imgui.BeginDragDropSource() {

			sourceID := trc.Component.ID()
			dragdrop.SetData(sourceID, trc.dragDropData)

			idPayload := sourceID

			imgui.SetDragDropPayload(
				trc.dragDropType,
				uintptr(unsafe.Pointer(&idPayload)),
				uint64(unsafe.Sizeof(idPayload)),
			)

			if trc.dragDropTooltipFn != nil {
				trc.dragDropTooltipFn()
			} else {
				imgui.Text(labelText)
			}

			imgui.EndDragDropSource()

			if !imgui.IsMouseDragging(imgui.MouseButtonLeft) {
				dragdrop.ClearData(sourceID)
			}
		}
	}

	for i := 1; i < len(trc.layout); i++ {
		imgui.TableNextColumn()
		trc.layout[i].Build()
	}

	if !isLeaf && open {
		if !trc.isPopulated && trc.factoryFn != nil {
			trc.children = trc.factoryFn(trc.childrenData)
			trc.isPopulated = true
			trc.childrenData = nil
			trc.factoryFn = nil
		}

		for _, child := range trc.children {
			if child != nil {
				imgui.TableNextRowV(0, 0)
				child.Build()
			}
		}
		imgui.TreePop()
	}
}

// Destroy cleans up the component and all its children
func (trc *TreeRowComponent) Destroy() {
	for _, c := range trc.layout {
		if c != nil {
			c.Destroy()
		}
	}

	for _, c := range trc.children {
		if c != nil {
			c.Destroy()
		}
	}

	trc.layout = nil
	trc.children = nil

	trc.Component.Destroy()
}
