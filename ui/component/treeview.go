package component

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
)

type TreeRowComponent struct {
	*Component
	label    ID
	flags    imgui.TreeNodeFlags
	layout   Layout
	children []*TreeRowComponent
}

func TreeRow(label string, components ...ComponentType) *TreeRowComponent {
	cmp := NewComponent(ID(fmt.Sprintf("tree-row::%s", label)))

	return &TreeRowComponent{
		Component: cmp,
		label:     ID(label),
		layout:    components,
	}
}

func (trc *TreeRowComponent) Children(rows ...*TreeRowComponent) *TreeRowComponent {
	trc.children = rows
	return trc
}

func (trc *TreeRowComponent) Flags(flags imgui.TreeNodeFlags) *TreeRowComponent {
	trc.flags = flags
	return trc
}

func (trc *TreeRowComponent) Layout() {
	imgui.TableNextRowV(0, 0)
	imgui.TableNextColumn()

	open := false
	if len(trc.children) > 0 {
		open = imgui.TreeNodeExStrV(trc.label.String(), trc.flags)
	} else {
		trc.flags |= imgui.TreeNodeFlagsLeaf | imgui.TreeNodeFlagsNoTreePushOnOpen
		if imgui.TreeNodeExStrV(trc.label.String(), trc.flags) {
			trc.HandleMouseEvents()
		}
	}

	for _, c := range trc.layout {
		switch c.(type) {
		// noop for different types like tooltips or context menus or popup modals
		default:
			imgui.TableNextColumn()
		}

		c.Layout()
	}

	if len(trc.children) > 0 && open {
		for _, child := range trc.children {
			child.Layout()
		}
		imgui.TreePop()
	}
}

var _ ComponentType = &TreeComponent{}

type TreeComponent struct {
	*Component
	flags        imgui.TableFlags
	size         imgui.Vec2
	columns      []*TableColumnComponent
	rows         []*TreeRowComponent
	freezeRow    int
	freezeColumn int
}

func Tree(id string) *TreeComponent {
	cmp := NewComponent(ID(id))

	return &TreeComponent{
		Component: cmp,
		flags: imgui.TableFlagsBordersV |
			imgui.TableFlagsBordersOuterH |
			imgui.TableFlagsResizable |
			imgui.TableFlagsNoBordersInBody,
		rows:    nil,
		columns: nil,
	}
}

// Freeze columns/rows so they stay visible when scrolled.
func (tt *TreeComponent) Freeze(col, row int) *TreeComponent {
	tt.freezeColumn = col
	tt.freezeRow = row

	return tt
}

// Size sets size of the table.
func (tt *TreeComponent) Size(width, height float32) *TreeComponent {
	tt.size = imgui.Vec2{X: width, Y: height}
	return tt
}

// Flags sets table flags.
func (tt *TreeComponent) Flags(flags imgui.TableFlags) *TreeComponent {
	tt.flags = flags
	return tt
}

// Columns sets table's columns.
func (tt *TreeComponent) Columns(cols ...*TableColumnComponent) *TreeComponent {
	tt.columns = cols
	return tt
}

// Rows sets TreeTable rows.
func (tt *TreeComponent) Rows(rows ...*TreeRowComponent) *TreeComponent {
	tt.rows = rows
	return tt
}

// Build implements Widget interface.
func (tt *TreeComponent) Layout() {
	if len(tt.rows) == 0 {
		return
	}

	colCount := len(tt.columns)
	if colCount == 0 {
		colCount = len(tt.rows[0].layout) + 1
	}

	if imgui.BeginTableV(
		tt.id.String(),
		int32(colCount),
		tt.flags,
		tt.size,
		0,
	) {
		if tt.freezeColumn >= 0 && tt.freezeRow >= 0 {
			imgui.TableSetupScrollFreeze(int32(tt.freezeColumn), int32(tt.freezeRow))
		}

		if len(tt.columns) > 0 {
			for _, col := range tt.columns {
				col.Layout()
			}

			imgui.TableHeadersRow()
		}

		for _, row := range tt.rows {

			row.Layout()
		}

		imgui.EndTable()
	}
}
