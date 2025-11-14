package settings

/*
┍━━━━━━━━━━━━━━━━╳┑
│ Settings Window │
└─────────────────┘
*/

import (
	"bitbox-editor/internal/app/animation"
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/component/label"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/app/window"
	"bitbox-editor/internal/config"
	"bitbox-editor/internal/logging"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
	"go.uber.org/zap"
)

var log = logging.NewLogger("settings")

type spectrumSettings struct {
	attackTime     int32
	attackEasing   string
	decayTime      int32
	decayEasing    string
	peakHoldTime   int32
	peakFallSpeed  float32
	noiseGate      float32
	colorMode      string
	staticColorIdx int32
}

type SettingsWindow struct {
	*window.Window[*SettingsWindow]

	currentThemeName    string
	consoleMaxLines     int32
	consoleMaxLinesTemp int32
	currentColormap     string
	spectrumSettings    spectrumSettings
	spectrumTemp        spectrumSettings
}

func NewSettingsWindow() *SettingsWindow {
	consoleMaxLines := int32(config.GetConsoleMaxLines())
	currentColormap := theme.GetColormapName(theme.GetCurrentColormap())

	spectSettings := spectrumSettings{
		attackTime:     int32(config.GetSpectrumAttackTime()),
		attackEasing:   config.GetSpectrumAttackEasing(),
		decayTime:      int32(config.GetSpectrumDecayTime()),
		decayEasing:    config.GetSpectrumDecayEasing(),
		peakHoldTime:   int32(config.GetSpectrumPeakHoldTime()),
		peakFallSpeed:  float32(config.GetSpectrumPeakFallSpeed()),
		noiseGate:      float32(config.GetSpectrumNoiseGate()),
		colorMode:      config.GetSpectrumColorMode(),
		staticColorIdx: int32(config.GetSpectrumStaticColorIdx()),
	}

	w := &SettingsWindow{
		currentThemeName:    theme.GetCurrentTheme().Name,
		consoleMaxLines:     consoleMaxLines,
		consoleMaxLinesTemp: consoleMaxLines,
		currentColormap:     currentColormap,
		spectrumSettings:    spectSettings,
		spectrumTemp:        spectSettings,
	}

	w.Window = window.NewWindow[*SettingsWindow]("Settings", "Cog", w.handleUpdate)

	w.Window.SetLayoutBuilder(w)

	return w
}

func (w *SettingsWindow) handleUpdate(cmd component.UpdateCmd) {
	if w.Window.HandleGlobalUpdate(cmd) {
		return
	}

	switch cmd.Type {
	case cmdSettingsSetTheme:
		if themeName, ok := cmd.Data.(string); ok {
			newTheme, err := theme.GetThemeByName(themeName)
			if err != nil {
				log.Error("Failed to get theme by name", zap.String("name", themeName), zap.Error(err))
				return
			}

			theme.TransitionToThemeWithEasing(newTheme, 500, animation.EaseOutCubic)

			w.currentThemeName = newTheme.Name

			// Save theme to config
			if err := config.SetTheme(themeName); err != nil {
				log.Error("Failed to save theme to config", zap.Error(err))
			}
		}

	case cmdSettingsUpdateCurrentThemeName:
		if name, ok := cmd.Data.(string); ok {
			w.currentThemeName = name
		}

	case cmdSettingsSetConsoleMaxLines:
		if maxLines, ok := cmd.Data.(int); ok {
			if err := config.SetConsoleMaxLines(maxLines); err != nil {
				log.Error("Failed to save console max lines to config", zap.Error(err))
			} else {
				w.consoleMaxLines = int32(maxLines)
				w.consoleMaxLinesTemp = int32(maxLines)
			}
		}

	case cmdSettingsSetColormap:
		if colormapName, ok := cmd.Data.(string); ok {
			colormap := theme.GetColormapByName(colormapName)
			theme.SetCurrentColormap(colormap)
			w.currentColormap = colormapName

			// Clear implot color cache to apply new colormap
			implot.BustColorCache()

			// Save colormap to config
			if err := config.SetColormap(colormapName); err != nil {
				log.Error("Failed to save colormap to config", zap.Error(err))
			}
		}

	case cmdSettingsSetSpectrumSettings:
		if settings, ok := cmd.Data.(spectrumSettings); ok {
			// Save all spectrum settings to config
			config.SetSpectrumAttackTime(int(settings.attackTime))
			config.SetSpectrumAttackEasing(settings.attackEasing)
			config.SetSpectrumDecayTime(int(settings.decayTime))
			config.SetSpectrumDecayEasing(settings.decayEasing)
			config.SetSpectrumPeakHoldTime(int(settings.peakHoldTime))
			config.SetSpectrumPeakFallSpeed(float64(settings.peakFallSpeed))
			config.SetSpectrumNoiseGate(float64(settings.noiseGate))
			config.SetSpectrumColorMode(settings.colorMode)
			config.SetSpectrumStaticColorIdx(int(settings.staticColorIdx))

			// Update current settings
			w.spectrumSettings = settings
			w.spectrumTemp = settings
		}

	default:
		log.Warn("SettingsWindow unhandled update", zap.Any("cmd", cmd))
	}
}

func (w *SettingsWindow) Menu() {}

func (w *SettingsWindow) Layout() {
	w.Window.ProcessUpdates()

	actualCurrentTheme := theme.GetCurrentTheme()
	if w.currentThemeName != actualCurrentTheme.Name {
		cmd := component.UpdateCmd{Type: cmdSettingsUpdateCurrentThemeName, Data: actualCurrentTheme.Name}
		w.Window.SendUpdate(cmd)
		w.currentThemeName = actualCurrentTheme.Name
	}
	currentName := w.currentThemeName

	imgui.BeginChildIDV(
		imgui.IDStr("settings::container"),
		imgui.NewVec2(-1, -1),
		imgui.ChildFlagsBorders,
		imgui.WindowFlagsChildWindow,
	)

	// Theme setting
	label.NewLabel("Theme").Build()
	imgui.SameLine()

	if imgui.BeginCombo("##theme", currentName) {
		themeNames := theme.GetNames()
		for _, themeName := range themeNames {
			isSelected := themeName == currentName
			if imgui.SelectableBoolV(themeName, isSelected, imgui.SelectableFlagsNone, imgui.Vec2{}) {
				cmd := component.UpdateCmd{Type: cmdSettingsSetTheme, Data: themeName}
				w.Window.SendUpdate(cmd)
				imgui.CloseCurrentPopup()
			}
			if isSelected {
				imgui.SetItemDefaultFocus()
			}
		}
		imgui.EndCombo()
	}

	imgui.Spacing()
	imgui.Separator()
	imgui.Spacing()

	// Console max lines setting
	label.NewLabel("Console Max Lines").Build()
	imgui.SameLineV(0, 10)

	imgui.PushItemWidth(200)
	if imgui.SliderIntV("##console_max_lines", &w.consoleMaxLinesTemp, 10, 1000, "%d", imgui.SliderFlagsNone) {
		// Slider changed, update temp value
	}
	imgui.PopItemWidth()

	// Show apply button if value changed
	if w.consoleMaxLinesTemp != w.consoleMaxLines {
		imgui.SameLine()
		if imgui.Button("Apply") {
			cmd := component.UpdateCmd{Type: cmdSettingsSetConsoleMaxLines, Data: int(w.consoleMaxLinesTemp)}
			w.Window.SendUpdate(cmd)
		}
		imgui.SameLine()
		if imgui.Button("Reset") {
			w.consoleMaxLinesTemp = w.consoleMaxLines
		}
	}

	imgui.Spacing()
	imgui.Separator()
	imgui.Spacing()

	// Colormap setting
	label.NewLabel("Colormap").Build()
	imgui.SameLine()

	if imgui.BeginCombo("##colormap", w.currentColormap) {
		colormapNames := theme.GetAllColormapNames()
		for _, colormapName := range colormapNames {
			isSelected := colormapName == w.currentColormap

			// Draw the selectable with colormap preview
			if imgui.SelectableBoolV("##"+colormapName, isSelected, imgui.SelectableFlagsNone, imgui.Vec2{X: 0, Y: 20}) {
				cmd := component.UpdateCmd{Type: cmdSettingsSetColormap, Data: colormapName}
				w.Window.SendUpdate(cmd)
				imgui.CloseCurrentPopup()
			}

			// Draw colormap preview gradient on the same line
			itemRectMin := imgui.ItemRectMin()
			itemRectMax := imgui.ItemRectMax()
			drawList := imgui.WindowDrawList()

			// Draw the colormap name text
			imgui.SetCursorScreenPos(imgui.Vec2{X: itemRectMin.X + 5, Y: itemRectMin.Y + 2})
			imgui.Text(colormapName)

			// Draw gradient bar below text
			barHeight := float32(8)
			barY := itemRectMin.Y + 12
			barWidth := itemRectMax.X - itemRectMin.X - 10
			steps := 50

			colormap := theme.GetColormapByName(colormapName)
			for i := 0; i < steps; i++ {
				t := float32(i) / float32(steps-1)
				x1 := itemRectMin.X + 5 + (barWidth * float32(i) / float32(steps))
				x2 := itemRectMin.X + 5 + (barWidth * float32(i+1) / float32(steps))

				// Sample color from colormap
				col := implot.SampleColormapV(t, colormap)
				drawList.AddRectFilledV(
					imgui.Vec2{X: x1, Y: barY},
					imgui.Vec2{X: x2, Y: barY + barHeight},
					imgui.NewColor(col.X, col.Y, col.Z, col.W).Pack(),
					0, 0,
				)
			}

			if isSelected {
				imgui.SetItemDefaultFocus()
			}
		}
		imgui.EndCombo()
	}

	imgui.Spacing()
	imgui.Separator()
	imgui.Spacing()

	// Spectrum Analyzer Settings Section
	label.NewLabel("Spectrum Analyzer").Build()
	imgui.Spacing()

	// Attack Time
	label.NewLabel("Attack Time (ms)").Build()
	imgui.SameLineV(0, 10)
	imgui.PushItemWidth(200)
	imgui.SliderIntV(
		"##spectrum_attack_time",
		&w.spectrumTemp.attackTime,
		1,
		500,
		"%d",
		imgui.SliderFlagsNone,
	)
	imgui.PopItemWidth()

	// Attack Easing
	label.NewLabel("Attack Easing").Build()
	imgui.SameLineV(0, 10)
	imgui.PushItemWidth(200)
	if imgui.BeginCombo("##attack_easing", w.spectrumTemp.attackEasing) {
		easingFuncs := []string{
			"EaseLinear", "EaseInQuad", "EaseOutQuad", "EaseInOutQuad",
			"EaseInCubic", "EaseOutCubic", "EaseInOutCubic",
			"EaseInExpo", "EaseOutExpo", "EaseInOutExpo",
			"EaseInSine", "EaseOutSine", "EaseInOutSine",
		}
		for _, easing := range easingFuncs {
			if imgui.SelectableBoolV(
				easing,
				easing == w.spectrumTemp.attackEasing,
				imgui.SelectableFlagsNone,
				imgui.Vec2{},
			) {
				w.spectrumTemp.attackEasing = easing
			}
		}
		imgui.EndCombo()
	}
	imgui.PopItemWidth()

	// Decay Time
	label.NewLabel("Decay Time (ms)").Build()
	imgui.SameLineV(0, 10)
	imgui.PushItemWidth(200)
	imgui.SliderIntV(
		"##spectrum_decay_time",
		&w.spectrumTemp.decayTime,
		10,
		2000,
		"%d",
		imgui.SliderFlagsNone,
	)
	imgui.PopItemWidth()

	// Decay Easing
	label.NewLabel("Decay Easing").Build()
	imgui.SameLineV(0, 10)
	imgui.PushItemWidth(200)
	if imgui.BeginCombo("##decay_easing", w.spectrumTemp.decayEasing) {
		easingFuncs := []string{
			"EaseLinear", "EaseInQuad", "EaseOutQuad", "EaseInOutQuad",
			"EaseInCubic", "EaseOutCubic", "EaseInOutCubic",
			"EaseInExpo", "EaseOutExpo", "EaseInOutExpo",
			"EaseInSine", "EaseOutSine", "EaseInOutSine",
		}
		for _, easing := range easingFuncs {
			if imgui.SelectableBoolV(
				easing,
				easing == w.spectrumTemp.decayEasing,
				imgui.SelectableFlagsNone,
				imgui.Vec2{},
			) {
				w.spectrumTemp.decayEasing = easing
			}
		}
		imgui.EndCombo()
	}
	imgui.PopItemWidth()

	// Peak Hold Time
	label.NewLabel("Peak Hold Time (ms)").Build()
	imgui.SameLineV(0, 10)
	imgui.PushItemWidth(200)
	imgui.SliderIntV(
		"##spectrum_peak_hold",
		&w.spectrumTemp.peakHoldTime,
		100,
		1500,
		"%d",
		imgui.SliderFlagsNone,
	)
	imgui.PopItemWidth()

	// Peak Fall Speed
	label.NewLabel("Peak Fall Speed").Build()
	imgui.SameLineV(0, 10)
	imgui.PushItemWidth(200)
	peakFallDisplay := w.spectrumTemp.peakFallSpeed * 1000.0
	if imgui.SliderFloatV(
		"##spectrum_peak_fall",
		&peakFallDisplay,
		1.0,
		20.0,
		"%.1f",
		imgui.SliderFlagsNone,
	) {
		w.spectrumTemp.peakFallSpeed = peakFallDisplay / 1000.0
	}
	imgui.PopItemWidth()

	// Noise Gate
	label.NewLabel("Noise Gate").Build()
	imgui.SameLineV(0, 10)
	imgui.PushItemWidth(200)
	imgui.SliderFloatV(
		"##spectrum_noise_gate",
		&w.spectrumTemp.noiseGate,
		0.0,
		0.2,
		"%.3f",
		imgui.SliderFlagsNone,
	)
	imgui.PopItemWidth()

	// Color Mode
	label.NewLabel("Color Mode").Build()
	imgui.SameLineV(0, 10)
	if imgui.RadioButtonBool("Height", w.spectrumTemp.colorMode == "height") {
		w.spectrumTemp.colorMode = "height"
	}
	imgui.SameLine()
	if imgui.RadioButtonBool("Gradient", w.spectrumTemp.colorMode == "gradient") {
		w.spectrumTemp.colorMode = "gradient"
	}
	imgui.SameLine()
	if imgui.RadioButtonBool("Static", w.spectrumTemp.colorMode == "static") {
		w.spectrumTemp.colorMode = "static"
	}

	// Static Color Index (only show if static mode)
	if w.spectrumTemp.colorMode == "static" {
		label.NewLabel("Static Color Index").Build()
		imgui.SameLineV(0, 10)
		imgui.PushItemWidth(200)
		imgui.SliderIntV(
			"##spectrum_static_color",
			&w.spectrumTemp.staticColorIdx,
			0,
			255,
			"%d",
			imgui.SliderFlagsNone,
		)
		imgui.PopItemWidth()

		// Preview color
		imgui.SameLine()
		previewColor := imgui.ColorConvertU32ToFloat4(imgui.ColorU32Vec4(
			implot.SampleColormapV(float32(w.spectrumTemp.staticColorIdx)/255.0, theme.GetCurrentColormap()),
		))
		imgui.ColorButtonV("##color_preview", previewColor, imgui.ColorEditFlagsNone, imgui.Vec2{X: 40, Y: 20})
	}

	// Show apply/reset buttons if any spectrum setting changed
	spectrumChanged := w.spectrumTemp.attackTime != w.spectrumSettings.attackTime ||
		w.spectrumTemp.attackEasing != w.spectrumSettings.attackEasing ||
		w.spectrumTemp.decayTime != w.spectrumSettings.decayTime ||
		w.spectrumTemp.decayEasing != w.spectrumSettings.decayEasing ||
		w.spectrumTemp.peakHoldTime != w.spectrumSettings.peakHoldTime ||
		w.spectrumTemp.peakFallSpeed != w.spectrumSettings.peakFallSpeed ||
		w.spectrumTemp.noiseGate != w.spectrumSettings.noiseGate ||
		w.spectrumTemp.colorMode != w.spectrumSettings.colorMode ||
		w.spectrumTemp.staticColorIdx != w.spectrumSettings.staticColorIdx

	if spectrumChanged {
		imgui.Spacing()
		if imgui.Button("Apply Spectrum Settings") {
			cmd := component.UpdateCmd{Type: cmdSettingsSetSpectrumSettings, Data: w.spectrumTemp}
			w.Window.SendUpdate(cmd)
		}
		imgui.SameLine()
		if imgui.Button("Reset Spectrum Settings") {
			w.spectrumTemp = w.spectrumSettings
		}
	}

	imgui.EndChild()
}
