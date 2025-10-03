package component

import (
	"github.com/AllenDang/cimgui-go/imgui"
)

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
	cmp := &TreeComponent{
		Component: NewComponent(imgui.IDStr(id)),
		flags: imgui.TableFlagsBordersV |
			imgui.TableFlagsBordersOuterH |
			imgui.TableFlagsResizable |
			imgui.TableFlagsNoBordersInBody,
		rows:    nil,
		columns: nil,
	}
	cmp.Component.layoutBuilder = cmp
	return cmp
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

func (tt *TreeComponent) Layout() {
	if len(tt.rows) == 0 {
		return
	}

	colCount := len(tt.columns)
	if colCount == 0 {
		colCount = len(tt.rows[0].layout) + 1
	}

	if imgui.BeginTableV(
		tt.Component.IDStr(),
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
