package component

import "github.com/AllenDang/cimgui-go/imgui"

// TODO: Add documentation

type DummyComponent struct {
	*Component[*DummyComponent]
	size imgui.Vec2
}

func (dc *DummyComponent) handleUpdate(cmd UpdateCmd) {
	dc.Component.HandleGlobalUpdate(cmd)
}

func (dc *DummyComponent) Layout() {
	dc.Component.ProcessUpdates()
	imgui.Dummy(dc.Component.size)
}

func NewDummy() *DummyComponent {
	cmp := &DummyComponent{}
	cmp.Component = NewComponent[*DummyComponent](imgui.IDStr("##dummy"), cmp.handleUpdate)
	cmp.Component.SetLayoutBuilder(cmp)
	return cmp
}
