package windows

import (
	"bitbox-editor/lib/preset"
	"bitbox-editor/ui/component"
	"bitbox-editor/ui/events"
	"context"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maniartech/signals"
)

//type Pad struct {
//	name string
//	file string
//
//	wave *audio.WaveFile
//
//	loading bool
//}

type PadGridWindow struct {
	*Window

	padGrid *component.PadGridComponent

	preset  *preset.Preset
	loading bool

	Events signals.Signal[events.PadEventRecord]
}

func (w *PadGridWindow) Menu() {}
func (w *PadGridWindow) Layout() {
	w.padGrid.Layout()
}

func (w *PadGridWindow) SetPreset(preset *preset.Preset) {
	w.preset = preset
	w.padGrid.SetPreset(preset)
}

func NewPadWindow() *PadGridWindow {
	cfg := NewWindowConfig()
	cfg.SetNoMenu(true)

	w := &PadGridWindow{
		Window: NewWindow("Pads", "Grid3x2", cfg),
		padGrid: component.NewPadGrid(
			imgui.IDStr("pad-grid"),
			2,
			4,
			72,
		),
		Events: signals.New[events.PadEventRecord](),
	}

	w.padGrid.Events.AddListener(
		func(ctx context.Context, record events.PadEventRecord) {
			// Bubble up event
			w.Events.Emit(ctx, record)
		},
		"pad-grid-events",
	)

	w.Window.layoutBuilder = w
	return w
}
