package windows

import (
	"bitbox-editor/ui/component"

	"github.com/AllenDang/cimgui-go/imgui"
)

type WavPlotWindow struct {
	*Window

	waveform *component.WaveComponent
}

func (w *WavPlotWindow) Menu() {}

func (w *WavPlotWindow) Layout() {
	w.waveform.Layout()
}

func NewWavPlotWindow() *WavPlotWindow {
	sw := &WavPlotWindow{
		Window: NewWindow("Wave", "AudioWaveform"),
		waveform: component.NewWaveformComponent(
			imgui.IDStr("waveform"),
		),
	}

	sw.Window.layoutBuilder = sw

	return sw
}
