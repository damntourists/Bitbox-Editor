package tree

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/component/table"
	"bitbox-editor/internal/logging"

	"github.com/AllenDang/cimgui-go/imgui"
)

var log = logging.NewLogger("tree")

type TreeComponent struct {
	*component.Component[*TreeComponent]
	component.CommandRouter

	flags        imgui.TableFlags
	rows         []*TreeRowComponent
	columns      []*table.TableColumnComponent
	freezeRow    int
	freezeColumn int
}

func NewTree(id string) *TreeComponent {
	cmp := &TreeComponent{
		flags: imgui.TableFlagsBordersV |
			imgui.TableFlagsBordersOuterH |
			imgui.TableFlagsResizable |
			imgui.TableFlagsNoBordersInBody,
		rows:         nil,
		columns:      nil,
		freezeRow:    0,
		freezeColumn: 0,
	}

	cmp.Component = component.NewComponent[*TreeComponent](imgui.IDStr(id))
	cmp.Component.SetLayoutBuilder(cmp)

	cmp.CommandRouter.Init(cmp.Component)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTreeFlags, cmp.onSetFlags)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTreeColumns, cmp.onSetColumns)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTreeRows, cmp.onSetRows)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTreeFreeze, cmp.onSetFreeze)

	return cmp
}

func (tt *TreeComponent) onSetFlags(flags imgui.TableFlags) {
	tt.flags = flags
}

func (tt *TreeComponent) onSetColumns(cols []*table.TableColumnComponent) {
	tt.columns = cols
}

func (tt *TreeComponent) onSetRows(rows []*TreeRowComponent) {
	tt.rows = rows
}

func (tt *TreeComponent) onSetFreeze(payload table.TableFreezePayload) {
	tt.freezeColumn = payload.Col
	tt.freezeRow = payload.Row
}

func (tt *TreeComponent) Freeze(col, row int) *TreeComponent {
	payload := table.TableFreezePayload{Col: col, Row: row}
	cmd := component.UpdateCmd{Type: cmdSetTreeFreeze, Data: payload}
	tt.SendUpdate(cmd)
	return tt
}

func (tt *TreeComponent) Size(width, height float32) *TreeComponent {
	tt.Component.SetSize(imgui.Vec2{X: width, Y: height})
	return tt
}

func (tt *TreeComponent) Flags(flags imgui.TableFlags) *TreeComponent {
	cmd := component.UpdateCmd{Type: cmdSetTreeFlags, Data: flags}
	tt.SendUpdate(cmd)
	return tt
}

func (tt *TreeComponent) Columns(cols ...*table.TableColumnComponent) *TreeComponent {
	colsCopy := make([]*table.TableColumnComponent, len(cols))
	copy(colsCopy, cols)
	cmd := component.UpdateCmd{Type: cmdSetTreeColumns, Data: colsCopy}
	tt.SendUpdate(cmd)
	return tt
}

func (tt *TreeComponent) Rows(rows ...*TreeRowComponent) *TreeComponent {
	rowsCopy := make([]*TreeRowComponent, len(rows))
	copy(rowsCopy, rows)
	cmd := component.UpdateCmd{Type: cmdSetTreeRows, Data: rowsCopy}
	tt.SendUpdate(cmd)
	return tt
}

func (tt *TreeComponent) Layout() {
	tt.Component.ProcessUpdates()

	rows := tt.rows
	size := tt.Component.Size()
	columns := tt.columns
	flags := tt.flags
	freezeCol := tt.freezeColumn
	freezeRow := tt.freezeRow
	colCount := len(columns)
	if colCount == 0 {
		colCount = 1
	}

	if len(rows) == 0 {
		if flags&imgui.TableFlagsScrollY != 0 {
			imgui.BeginChildStrV(tt.IDStr()+"_empty_scroll", size, 0, 0)
			imgui.Text("Empty")
			imgui.EndChild()
		} else {
			imgui.Text("Empty")
		}
		return
	}

	if imgui.BeginTableV(tt.Component.IDStr(), int32(colCount), flags, size, 0) {
		defer imgui.EndTable()

		if freezeCol >= 0 && freezeRow >= 0 {
			imgui.TableSetupScrollFreeze(int32(freezeCol), int32(freezeRow))
		}
		if len(columns) > 0 {
			for _, col := range columns {
				if col != nil {
					col.Build()
				}
			}
			imgui.TableHeadersRow()
		}

		for _, row := range rows {
			if row != nil {
				row.Build()
			}
		}

	}
}

// Destroy cleans up the component and all its children
func (tt *TreeComponent) Destroy() {
	for _, col := range tt.columns {
		if col != nil {
			col.Destroy()
		}
	}
	for _, row := range tt.rows {
		if row != nil {
			row.Destroy()
		}
	}
	tt.columns = nil
	tt.rows = nil
	tt.Component.Destroy()
}
