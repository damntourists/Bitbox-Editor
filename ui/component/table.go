package component

import (
	"github.com/AllenDang/cimgui-go/imgui"
)

type SortDirection byte

const (
	SortAscending  SortDirection = 1
	SortDescending SortDirection = 2
)

type TableComponent struct {
	*Component

	flags imgui.TableFlags

	size         imgui.Vec2
	innerWidth   float64
	rows         []*TableRowComponent
	columns      []*TableColumnComponent
	fastMode     bool
	freezeRow    int
	freezeColumn int
	noHeader     bool
}

func (t *TableComponent) FastMode(b bool) *TableComponent {
	t.fastMode = b
	return t
}

func (t *TableComponent) NoHeader(b bool) *TableComponent {
	t.noHeader = b
	return t
}

func (t *TableComponent) Freeze(col, row int) *TableComponent {
	t.freezeColumn = col
	t.freezeRow = row
	return t
}

func (t *TableComponent) Columns(cols ...*TableColumnComponent) *TableComponent {
	t.columns = cols
	return t
}

func (t *TableComponent) Rows(rows ...*TableRowComponent) *TableComponent {
	t.rows = rows
	return t
}

func (t *TableComponent) Size(width, height float32) *TableComponent {
	t.size = imgui.Vec2{X: width, Y: height}
	return t
}

func (t *TableComponent) InnerWidth(width float64) *TableComponent {
	t.innerWidth = width
	return t
}

func (t *TableComponent) Flags(flags imgui.TableFlags) *TableComponent {
	t.flags = flags
	return t
}

func (t *TableComponent) colCount() int {
	colCount := len(t.columns)
	if colCount == 0 {
		if len(t.rows) > 0 {
			return len(t.rows[0].layout)
		}
		return 1
	}
	return colCount
}

func (t *TableComponent) handleSort() {
	if specs := imgui.TableGetSortSpecs(); specs != nil {
		if specs.SpecsDirty() {
			cs := specs.Specs()
			colIdx := cs.ColumnIndex()
			sortDir := cs.SortDirection()

			if col := t.columns[colIdx]; col.sortFn != nil {
				col.sortFn(SortDirection(sortDir))
			}

			specs.SetSpecsDirty(false)
		}
	}
}

func (t *TableComponent) Layout() {
	if imgui.BeginTableV(t.Component.IDStr(), int32(t.colCount()), t.flags, t.size, float32(t.innerWidth)) {
		if t.freezeColumn >= 0 && t.freezeRow >= 0 {
			imgui.TableSetupScrollFreeze(int32(t.freezeColumn), int32(t.freezeRow))
		}

		if len(t.columns) > 0 {
			for _, col := range t.columns {
				col.Layout()
			}

			if !t.noHeader {
				imgui.TableHeadersRow()
			}

			if t.flags&imgui.TableFlagsSortable != 0 {
				t.handleSort()
			}
		}

		if t.fastMode {
			clipper := imgui.NewListClipper()
			defer clipper.Destroy()

			clipper.Begin(int32(len(t.rows)))

			for clipper.Step() {
				for i := clipper.DisplayStart(); i < clipper.DisplayEnd(); i++ {
					row := t.rows[i]
					row.Layout()
				}
			}

			clipper.End()
		} else {
			for _, row := range t.rows {
				row.Layout()
			}
		}

		imgui.EndTable()
	}
}

func NewTableComponent(id imgui.ID) *TableComponent {
	tc := &TableComponent{
		Component: NewComponent(id),
		flags: imgui.TableFlagsResizable |
			imgui.TableFlagsScrollY |
			imgui.TableFlagsScrollY |
			imgui.TableFlagsNoBordersInBody |
			imgui.TableFlagsNoPadOuterX,
		rows:         nil,
		columns:      nil,
		fastMode:     false,
		freezeRow:    -1,
		freezeColumn: -1,
		noHeader:     false,
	}
	tc.Component.layoutBuilder = tc
	return tc
}
