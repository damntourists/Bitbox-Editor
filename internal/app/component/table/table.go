package table

/*
╭─────────────────╮
│ Table Component │
╰─────────────────╯
*/

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/logging"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

var log = logging.NewLogger("table")

type TableComponent struct {
	*component.Component[*TableComponent]

	flags imgui.TableFlags

	innerWidth   float64
	rows         []*TableRowComponent
	columns      []*TableColumnComponent
	fastMode     bool
	freezeRow    int
	freezeColumn int
	noHeader     bool
}

func NewTableComponent(id imgui.ID) *TableComponent {
	cmp := &TableComponent{
		flags: imgui.TableFlagsResizable |
			imgui.TableFlagsBordersInnerV |
			imgui.TableFlagsScrollY |
			imgui.TableFlagsNoBordersInBody |
			imgui.TableFlagsNoPadOuterX |
			imgui.TableFlagsSortable,
		rows:         nil,
		columns:      nil,
		fastMode:     false,
		freezeRow:    0,
		freezeColumn: 0,
		noHeader:     false,
		innerWidth:   0,
	}

	cmp.Component = component.NewComponent[*TableComponent](id, cmp.handleUpdate)
	cmp.Component.SetLayoutBuilder(cmp)

	return cmp
}

func (t *TableComponent) handleUpdate(cmd component.UpdateCmd) {
	if t.Component.HandleGlobalUpdate(cmd) {
		return
	}

	switch cmd.Type {
	case cmdSetTableFlags:
		if flags, ok := cmd.Data.(imgui.TableFlags); ok {
			t.flags = flags
		}

	case cmdSetTableInnerWidth:
		if width, ok := cmd.Data.(float64); ok {
			t.innerWidth = width
		}

	case cmdSetTableRows:
		if rows, ok := cmd.Data.([]*TableRowComponent); ok {
			t.rows = rows
		} else if cmd.Data == nil {
			t.rows = nil
		}

	case cmdSetTableColumns:
		if cols, ok := cmd.Data.([]*TableColumnComponent); ok {
			t.columns = cols
		} else if cmd.Data == nil {
			t.columns = nil
		}

	case cmdSetTableFastMode:
		if mode, ok := cmd.Data.(bool); ok {
			t.fastMode = mode
		}

	case cmdSetTableFreeze:
		if payload, ok := cmd.Data.(TableFreezePayload); ok {
			t.freezeColumn = payload.Col
			t.freezeRow = payload.Row
		}

	case cmdSetTableNoHeader:
		if noHeader, ok := cmd.Data.(bool); ok {
			t.noHeader = noHeader
		}

	default:
		log.Warn("TableComponent unhandled update", zap.String("id", t.IDStr()), zap.Any("cmd", cmd))
	}
}

func (t *TableComponent) colCount() int {
	colCount := len(t.columns)
	if colCount == 0 {
		if len(t.rows) > 0 && t.rows[0] != nil {
			if t.rows[0].layout != nil {
				return len(t.rows[0].layout)
			}
		}
		return 1
	}
	return colCount
}

func (t *TableComponent) handleSort() {
	if specs := imgui.TableGetSortSpecs(); specs != nil && specs.SpecsDirty() {
		cs := specs.Specs()
		colIdx := cs.ColumnIndex()
		sortDir := cs.SortDirection()

		if colIdx >= 0 && int(colIdx) < len(t.columns) {
			col := t.columns[colIdx]
			if col != nil && col.sortFn != nil {
				col.sortFn(SortDirection(sortDir))
			}
		}
		specs.SetSpecsDirty(false)
	}
}

func (t *TableComponent) SetFastMode(b bool) *TableComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableFastMode, Data: b}
	t.Component.SendUpdate(cmd)
	return t
}

func (t *TableComponent) SetNoHeader(b bool) *TableComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableNoHeader, Data: b}
	t.Component.SendUpdate(cmd)
	return t
}

func (t *TableComponent) SetFreeze(col, row int) *TableComponent {
	payload := TableFreezePayload{Col: col, Row: row}
	cmd := component.UpdateCmd{Type: cmdSetTableFreeze, Data: payload}
	t.Component.SendUpdate(cmd)
	return t
}

func (t *TableComponent) SetColumns(cols ...*TableColumnComponent) *TableComponent {
	colsCopy := make([]*TableColumnComponent, len(cols))
	copy(colsCopy, cols)
	cmd := component.UpdateCmd{Type: cmdSetTableColumns, Data: colsCopy}
	t.Component.SendUpdate(cmd)
	return t
}

func (t *TableComponent) SetRows(rows ...*TableRowComponent) *TableComponent {
	rowsCopy := make([]*TableRowComponent, len(rows))
	copy(rowsCopy, rows)
	cmd := component.UpdateCmd{Type: cmdSetTableRows, Data: rowsCopy}
	t.Component.SendUpdate(cmd)
	return t
}

func (t *TableComponent) SetSize(width, height float32) *TableComponent {
	t.Component.SetSize(imgui.Vec2{X: width, Y: height})
	return t
}

func (t *TableComponent) SetInnerWidth(width float64) *TableComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableInnerWidth, Data: width}
	t.Component.SendUpdate(cmd)
	return t
}

func (t *TableComponent) SetFlags(flags imgui.TableFlags) *TableComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableFlags, Data: flags}
	t.Component.SendUpdate(cmd)
	return t
}

func (t *TableComponent) Layout() {
	t.Component.ProcessUpdates()

	flags := t.flags
	size := t.Component.Size()
	innerWidth := t.innerWidth
	columns := t.columns
	rows := t.rows
	freezeCol := t.freezeColumn
	freezeRow := t.freezeRow
	noHeader := t.noHeader
	fastMode := t.fastMode
	colCount := t.colCount()

	if imgui.BeginTableV(t.Component.IDStr(), int32(colCount), flags, size, float32(innerWidth)) {
		defer imgui.EndTable()

		if freezeCol >= 0 && freezeRow >= 0 {
			imgui.TableSetupScrollFreeze(int32(freezeCol), int32(freezeRow))
		}

		// Setup columns
		if len(columns) > 0 {
			for _, col := range columns {
				if col != nil {
					col.Build()
				}
			}

			if !noHeader {
				imgui.TableHeadersRow()
			}

			if flags&imgui.TableFlagsSortable != 0 {
				t.handleSort()
			}
		}

		// Render rows
		if fastMode {
			clipper := imgui.NewListClipper()
			clipper.Begin(int32(len(rows)))
			for clipper.Step() {
				for i := clipper.DisplayStart(); i < clipper.DisplayEnd(); i++ {
					if i >= 0 && int(i) < len(rows) {
						row := rows[i]
						if row != nil {
							row.Build()
						}
					}
				}
			}
			clipper.End()
			clipper.Destroy()
		} else {
			for _, row := range rows {
				if row != nil {
					row.Build()
				}
			}
		}
	}
}

// Destroy cleans up the component and all its children
func (t *TableComponent) Destroy() {
	for _, col := range t.columns {
		if col != nil {
			col.Destroy()
		}
	}
	for _, row := range t.rows {
		if row != nil {
			row.Destroy()
		}
	}
	t.columns = nil
	t.rows = nil
	t.Component.Destroy()
}
