package windows

type PresetWindow struct {
	*Window
}

func (w *PresetWindow) Menu() {}

func (w *PresetWindow) Layout() {

}

func NewPresetWindow() *PresetWindow {
	sw := &PresetWindow{
		Window: NewWindow("Presets", "ListMusic", NewWindowConfig()),
	}

	// 			"/home/brett/Downloads/micro_bundle-v3/Soundopolis/Sci Fi/Texture_Alien_Bugs_002.wav",

	sw.Window.layoutBuilder = sw

	return sw
}
