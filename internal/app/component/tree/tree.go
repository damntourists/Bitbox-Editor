package tree

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/component/table"
	"bitbox-editor/internal/logging"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

var log = logging.NewLogger("tree")

type TreeComponent struct {
	*component.Component[*TreeComponent]

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

	cmp.Component = component.NewComponent[*TreeComponent](imgui.IDStr(id), cmp.handleUpdate)
	cmp.Component.SetLayoutBuilder(cmp)

	return cmp
}

func (tt *TreeComponent) handleUpdate(cmd component.UpdateCmd) {
	if tt.Component.HandleGlobalUpdate(cmd) {
		return
	}

	switch cmd.Type {
	case cmdSetTreeFlags:
		if flags, ok := cmd.Data.(imgui.TableFlags); ok {
			tt.flags = flags
		}
	case cmdSetTreeColumns:
		if cmd.Data == nil {
			tt.columns = nil
		} else if cols, ok := cmd.Data.([]*table.TableColumnComponent); ok {
			tt.columns = cols
		}
	case cmdSetTreeRows:
		if cmd.Data == nil {
			tt.rows = nil
		} else if rows, ok := cmd.Data.([]*TreeRowComponent); ok {
			tt.rows = rows
		}
	case cmdSetTreeFreeze:
		if payload, ok := cmd.Data.(table.TableFreezePayload); ok {
			tt.freezeColumn = payload.Col
			tt.freezeRow = payload.Row
		}
	default:
		log.Warn("TreeComponent unhandled update", zap.String("id", tt.IDStr()), zap.Any("cmd", cmd))
	}
}

func (tt *TreeComponent) Freeze(col, row int) *TreeComponent {
	payload := table.TableFreezePayload{Col: col, Row: row} // Reuse payload type
	cmd := component.UpdateCmd{Type: cmdSetTreeFreeze, Data: payload}
	tt.Component.SendUpdate(cmd)
	return tt
}

func (tt *TreeComponent) Size(width, height float32) *TreeComponent {
	tt.Component.SetSize(imgui.Vec2{X: width, Y: height})
	return tt
}

func (tt *TreeComponent) Flags(flags imgui.TableFlags) *TreeComponent {
	cmd := component.UpdateCmd{Type: cmdSetTreeFlags, Data: flags}
	tt.Component.SendUpdate(cmd)
	return tt
}

func (tt *TreeComponent) Columns(cols ...*table.TableColumnComponent) *TreeComponent {
	colsCopy := make([]*table.TableColumnComponent, len(cols))
	copy(colsCopy, cols)
	cmd := component.UpdateCmd{Type: cmdSetTreeColumns, Data: colsCopy}
	tt.Component.SendUpdate(cmd)
	return tt
}

func (tt *TreeComponent) Rows(rows ...*TreeRowComponent) *TreeComponent {
	rowsCopy := make([]*TreeRowComponent, len(rows))
	copy(rowsCopy, rows)
	cmd := component.UpdateCmd{Type: cmdSetTreeRows, Data: rowsCopy}
	tt.Component.SendUpdate(cmd)
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
