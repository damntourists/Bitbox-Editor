package table

import (
	"bitbox-editor/internal/app/component"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

type TableColumnComponent struct {
	*component.Component[*TableColumnComponent]
	label              string
	flags              imgui.TableColumnFlags
	innerWidthOrWeight float32
	userID             int32
	sortFn             SortFunc
}

func NewTableColumn(label string) *TableColumnComponent {
	cmp := &TableColumnComponent{
		label:              label,
		flags:              0,
		innerWidthOrWeight: 0,
		userID:             0,
		sortFn:             nil,
	}

	id := imgui.IDStr(fmt.Sprintf("table-column::%s", label))
	cmp.Component = component.NewComponent[*TableColumnComponent](id, cmp.handleUpdate)
	cmp.Component.SetLayoutBuilder(cmp)

	return cmp
}

func (tcc *TableColumnComponent) handleUpdate(cmd component.UpdateCmd) {
	if tcc.Component.HandleGlobalUpdate(cmd) {
		return
	}

	switch cmd.Type {
	case cmdSetTableColumnFlags:
		if flags, ok := cmd.Data.(imgui.TableColumnFlags); ok {
			tcc.flags = flags
		}

	case cmdSetTableColumnWidthOrWeight:
		if w, ok := cmd.Data.(float32); ok {
			tcc.innerWidthOrWeight = w
		}

	case cmdSetTableColumnUserID:
		if id, ok := cmd.Data.(int32); ok {
			tcc.userID = id
		}

	case cmdSetTableColumnSortFn:
		if fn, ok := cmd.Data.(SortFunc); ok {
			tcc.sortFn = fn
		} else if cmd.Data == nil {
			tcc.sortFn = nil
		}

	default:
		log.Warn(
			"TableColumnComponent unhandled update",
			zap.String("id", tcc.IDStr()),
			zap.Any("cmd", cmd),
		)
	}
}

func (tcc *TableColumnComponent) SetFlags(flags imgui.TableColumnFlags) *TableColumnComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableColumnFlags, Data: flags}
	tcc.Component.SendUpdate(cmd)
	return tcc
}

func (tcc *TableColumnComponent) SetInnerWidthOrWeight(w float32) *TableColumnComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableColumnWidthOrWeight, Data: w}
	tcc.Component.SendUpdate(cmd)
	return tcc
}

func (tcc *TableColumnComponent) SetUserID(id int32) *TableColumnComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableColumnUserID, Data: id}
	tcc.Component.SendUpdate(cmd)
	return tcc
}

func (tcc *TableColumnComponent) SetSortFn(fn SortFunc) *TableColumnComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableColumnSortFn, Data: fn}
	tcc.Component.SendUpdate(cmd)
	return tcc
}
func (tcc *TableColumnComponent) Layout() {
	tcc.Component.ProcessUpdates()

	label := tcc.label
	flags := tcc.flags
	widthOrWeight := tcc.innerWidthOrWeight
	userID := tcc.userID

	imgui.TableSetupColumnV(label, flags, widthOrWeight, imgui.IDInt(userID))
}

// Destroy cleans up the component
func (tcc *TableColumnComponent) Destroy() {
	tcc.Component.Destroy()
}
