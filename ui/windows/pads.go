package windows

import (
	"bitbox-editor/lib/audio"
	"bitbox-editor/lib/preset"
	"bitbox-editor/ui/component"

	"github.com/AllenDang/cimgui-go/imgui"
)

type Pad struct {
	name string
	file string

	wave *audio.WaveFile

	loading bool
}

type PadWindow struct {
	*Window

	padGrid *component.PadGridComponent

	preset  *preset.Preset
	loading bool
}

func (w *PadWindow) Menu() {}
func (w *PadWindow) Layout() {
	w.padGrid.Layout()
}

func (w *PadWindow) SetPreset(preset *preset.Preset) {
	w.preset = preset
	w.padGrid.SetPreset(preset)
}

func NewPadWindow() *PadWindow {
	cfg := NewWindowConfig()
	cfg.SetNoMenu(true)

	w := &PadWindow{
		Window: NewWindow("Pads", "Grid3x2", cfg),
		padGrid: component.NewPadGrid(
			imgui.IDStr("pad-grid"),
			2,
			4,
			82,
		),
	}

	w.Window.layoutBuilder = w
	return w
}
