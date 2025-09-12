package windows

import (
	"bitbox-editor/lib/audio"
	"bitbox-editor/ui/component"
	"fmt"

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

	rows, cols int

	pads [][]*Pad

	loading bool
}

func (w *PadWindow) Menu() {}
func (w *PadWindow) Layout() {
	for i := 0; i < w.rows; i++ {
		for j := 0; j < w.cols; j++ {
			p := component.NewPad(
				imgui.IDStr(fmt.Sprintf("pad-%dx%d", i, j)),
				"row1",
				"row2",
				"row3",
			)
			p.Layout()
			imgui.SameLine()
		}
		imgui.NewLine()
	}

}

func NewPadWindow() *PadWindow {
	cfg := NewWindowConfig()
	cfg.SetNoMenu(true)

	w := &PadWindow{
		Window: NewWindow("Pads", "Grid3x2", cfg),
		rows:   2,
		cols:   4,
	}
	w.loading = true
	w.Window.loading = true
	w.Window.layoutBuilder = w
	return w
}
