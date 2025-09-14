package windows

import (
	"bitbox-editor/lib/preset"
	"bitbox-editor/ui/component"
	"bitbox-editor/ui/events"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maniartech/signals"
)

type PadConfigWindow struct {
	*Window

	pad *component.PadComponent

	Events signals.Signal[events.PadEventRecord]

	preset *preset.Preset
}

func (w *PadConfigWindow) SetPad(pad *component.PadComponent) {
	w.pad = pad
}

func (w *PadConfigWindow) SetPreset(preset *preset.Preset) {
	w.preset = preset

	println(preset.Name)

}

func (w *PadConfigWindow) Menu() {}
func (w *PadConfigWindow) Layout() {
	if w.pad == nil {
		imgui.Text("No pad selected")
		return
	}

	if w.preset == nil {
		imgui.Text("No preset selected")
		return
	}

}

func NewPadConfigWindow() *PadConfigWindow {
	w := &PadConfigWindow{
		Window: NewWindow("Pad Config", "PadConfig", NewWindowConfig()),
		pad:    nil,
		preset: nil,
	}
	w.Window.layoutBuilder = w
	return w
}
