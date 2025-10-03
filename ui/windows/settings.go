package windows

import (
	"bitbox-editor/lib/util"
	"bitbox-editor/ui/theme"

	"github.com/AllenDang/cimgui-go/imgui"
)

type SettingsWindow struct {
	*Window
}

func (w *SettingsWindow) Menu() {}

func (w *SettingsWindow) Layout() {
	imgui.BeginChildIDV(
		imgui.IDStr("settings::theme"),
		imgui.NewVec2(-1, -1),
		imgui.ChildFlagsBorders,
		imgui.WindowFlagsChildWindow,
	)

	imgui.Text("Theme")
	imgui.SameLine()

	if imgui.BeginComboV(
		"##theme",
		theme.GetCurrentTheme().Name,
		imgui.ComboFlagsNone) {

		for _, t := range theme.GetNames() {
			if imgui.SelectableBool(t) {
				tt, err := theme.GetThemeByName(t)
				util.PanicOnError(err)
				theme.SetCurrentTheme(tt)
			}

		}

		imgui.EndCombo()
	}

	imgui.EndChild()
}

func NewSettingsWindow() *SettingsWindow {
	sw := &SettingsWindow{
		Window: NewWindow("Settings", "Cog"),
	}
	sw.Window.layoutBuilder = sw
	return sw
}
