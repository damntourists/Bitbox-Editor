package windows

import (
	"bitbox-editor/ui/component"
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
		Window: NewWindow("Wave Test", "AudioWaveform", NewWindowConfig()),
		waveform: component.NewWaveformComponent(
			"waveform",
		),
	}

	// 			"/home/brett/Downloads/micro_bundle-v3/Soundopolis/Sci Fi/Texture_Alien_Bugs_002.wav",

	sw.Window.layoutBuilder = sw

	return sw
}
