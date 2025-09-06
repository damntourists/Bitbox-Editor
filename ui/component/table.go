package component

import (
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
)

type TableColumnComponent struct {
	*Component
	label              string
	flags              imgui.TableColumnFlags
	innerWidthOrWeight float32
	userID             int32
}

func TableColumn(label string) *TableColumnComponent {
	cmp := NewComponent(
		ID(fmt.Sprintf("table-column::%s", label)),
	)

	return &TableColumnComponent{
		Component:          cmp,
		label:              label,
		flags:              0,
		innerWidthOrWeight: 0,
		userID:             0,
	}
}

func (tcc *TableColumnComponent) Flags(flags imgui.TableColumnFlags) *TableColumnComponent {
	tcc.flags = flags
	return tcc
}

func (tcc *TableColumnComponent) InnerWidthOrWeight(w float32) *TableColumnComponent {
	tcc.innerWidthOrWeight = w
	return tcc
}

func (tcc *TableColumnComponent) UserID(id int32) *TableColumnComponent {
	tcc.userID = id
	return tcc
}

func (tcc *TableColumnComponent) Layout() {
	imgui.TableSetupColumnV(
		tcc.label,
		tcc.flags,
		tcc.innerWidthOrWeight,
		imgui.IDInt(tcc.userID),
	)
}
