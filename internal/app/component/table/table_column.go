package table

import (
	"bitbox-editor/internal/app/component"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
)

type TableColumnComponent struct {
	*component.Component[*TableColumnComponent]
	component.CommandRouter
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
	cmp.Component = component.NewComponent[*TableColumnComponent](id)
	cmp.Component.SetLayoutBuilder(cmp)

	cmp.CommandRouter.Init(cmp.Component)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTableColumnFlags, cmp.onSetFlags)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTableColumnWidthOrWeight, cmp.onSetWidthOrWeight)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTableColumnUserID, cmp.onSetUserID)
	component.OnCommandTyped(&cmp.CommandRouter, cmdSetTableColumnSortFn, cmp.onSetSortFn)

	return cmp
}

func (tcc *TableColumnComponent) onSetFlags(flags imgui.TableColumnFlags) {
	tcc.flags = flags
}

func (tcc *TableColumnComponent) onSetWidthOrWeight(w float32) {
	tcc.innerWidthOrWeight = w
}

func (tcc *TableColumnComponent) onSetUserID(id int32) {
	tcc.userID = id
}

func (tcc *TableColumnComponent) onSetSortFn(fn SortFunc) {
	tcc.sortFn = fn
}

func (tcc *TableColumnComponent) SetFlags(flags imgui.TableColumnFlags) *TableColumnComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableColumnFlags, Data: flags}
	tcc.SendUpdate(cmd)
	return tcc
}

func (tcc *TableColumnComponent) SetInnerWidthOrWeight(w float32) *TableColumnComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableColumnWidthOrWeight, Data: w}
	tcc.SendUpdate(cmd)
	return tcc
}

func (tcc *TableColumnComponent) SetUserID(id int32) *TableColumnComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableColumnUserID, Data: id}
	tcc.SendUpdate(cmd)
	return tcc
}

func (tcc *TableColumnComponent) SetSortFn(fn SortFunc) *TableColumnComponent {
	cmd := component.UpdateCmd{Type: cmdSetTableColumnSortFn, Data: fn}
	tcc.SendUpdate(cmd)
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
