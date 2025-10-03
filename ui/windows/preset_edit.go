package windows

import (
	libEvents "bitbox-editor/lib/events"
	"bitbox-editor/lib/preset"
	"bitbox-editor/ui/component"
	uiEvents "bitbox-editor/ui/events"
	"bitbox-editor/ui/fonts"
	"bitbox-editor/ui/theme"
	"context"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maniartech/signals"
)

type PresetEditWindow struct {
	*Window

	Components struct {
		Wave *component.WaveComponent
		// TODO: Wav controls
		PadGrid           *component.PadGridComponent
		PadConfig         *component.PadConfigComponent
		PadGridSizeSelect *component.ComboBoxComponent
	}

	preset *preset.Preset

	loading bool

	Events signals.Signal[libEvents.PresetEditEventRecord]
}

func (w *PresetEditWindow) Menu() {}
func (w *PresetEditWindow) Layout() {
	t := theme.GetCurrentTheme()

	if w.preset == nil {
		imgui.Text("No preset selected")
		return
	}

	imgui.BeginChildStrV(
		"top-controls",
		imgui.Vec2{X: 0, Y: 0},
		imgui.ChildFlagsResizeY|imgui.ChildFlagsBorders,
		imgui.WindowFlagsNone|imgui.WindowFlagsMenuBar,
	)

	if imgui.BeginMenuBar() {

		component.NewLabelComponent("Wave").
			SetRounding(2).
			Layout()

		imgui.SameLine()

		imgui.EndMenuBar()
	}

	imgui.BeginChildStrV(
		"top-controls",
		imgui.Vec2{X: 0, Y: imgui.FrameHeight() + (t.Style.FramePadding[1] * 2)},
		imgui.ChildFlagsBorders|imgui.ChildFlagsFrameStyle,
		imgui.WindowFlagsNone,
	)

	if w.Components.Wave.Wav() == nil {
		imgui.Text("No wav loaded")
	} else {

		if w.Components.Wave.Wav().IsPlaying() {
			if imgui.Button(fonts.Icon("Square")) {
				w.Components.Wave.Wav().Stop()
			}
			imgui.SameLine()
			imgui.Text(fmt.Sprintf("%.2f", w.Components.Wave.Wav().PositionSeconds()))
		} else {
			if imgui.Button(fonts.Icon("Play")) {
				w.Components.Wave.Wav().Play()
			}
			imgui.SameLine()
		}
	}

	imgui.EndChild()

	imgui.BeginChildStrV(
		"top-waveform",
		imgui.Vec2{X: 0, Y: 0},
		imgui.ChildFlagsNone,
		imgui.WindowFlagsNone,
	)
	w.Components.Wave.Layout()
	imgui.EndChild()

	imgui.EndChild()

	imgui.BeginChildStrV(
		"split-left-3",
		imgui.Vec2{X: 0, Y: 0},
		imgui.ChildFlagsNone,
		imgui.WindowFlagsNone,
	)

	imgui.BeginChildStrV(
		"split-left-3-left",
		imgui.Vec2{X: imgui.ContentRegionAvail().X * 0.66, Y: 0},
		imgui.ChildFlagsBorders|imgui.ChildFlagsResizeX,
		imgui.WindowFlagsNone|imgui.WindowFlagsMenuBar,
	)

	// Pad settings
	if imgui.BeginMenuBar() {
		component.NewLabelComponent("Configuration").
			SetRounding(2).
			Layout()
		imgui.EndMenuBar()
	}

	w.Components.PadConfig.Layout()

	imgui.EndChild()
	imgui.SameLine()
	imgui.BeginChildStrV(
		"split-left-3-right",
		imgui.Vec2{X: 0, Y: 0},
		imgui.ChildFlagsBorders,
		imgui.WindowFlagsNone|imgui.WindowFlagsMenuBar,
	)

	// Pad grid
	if imgui.BeginMenuBar() {

		component.NewLabelComponent("Pads").
			SetRounding(2).
			Layout()

		w.Components.PadGridSizeSelect.Layout()

		imgui.SameLine()

		imgui.EndMenuBar()
	}

	w.Components.PadGrid.Layout()

	imgui.EndChild()

	imgui.EndChild()
}

func (w *PresetEditWindow) Preset() *preset.Preset {
	return w.preset
}

func (w *PresetEditWindow) SetPreset(preset *preset.Preset) {
	w.preset = preset
	w.Components.PadGrid.SetPreset(preset)

}

func NewPresetEditWindow(p *preset.Preset) *PresetEditWindow {
	w := &PresetEditWindow{
		Window: NewWindow(fmt.Sprintf("Preset: %s", p.Name), "FileMusic").
			SetFlags(imgui.WindowFlagsNoScrollbar),
		preset: p,
		Events: signals.New[libEvents.PresetEditEventRecord](),
	}

	w.Components.Wave = component.NewWaveformComponent(imgui.IDStr("wav"))

	w.Components.PadGrid = component.NewPadGrid(
		imgui.IDStr(fmt.Sprintf("%s-pad-grid", w.Title())),
		2,
		4,
		100,
	)
	w.Components.PadGrid.Events.AddListener(func(ctx context.Context, record uiEvents.PadEventRecord) {
		switch record.Type {
		case uiEvents.PadActivated:
			pc := record.Data.(*component.PadComponent)
			go w.Components.PadConfig.SetPad(pc)
			go w.Components.Wave.SetWav(pc.Wav())
		}
	},
		fmt.Sprintf("%s-pad-grid-events", w.Title()))

	w.Components.PadConfig = component.NewPadConfigComponent(imgui.IDStr("pad-config"), p)

	w.Components.PadGridSizeSelect = component.NewComboBoxComponent(
		imgui.IDStr("pad-grid-combo"),
		"##grid-config",
	).
		SetPreview(fonts.Icon("Grid3x2")).
		SetFlags(imgui.ComboFlagsWidthFitPreview).
		SetItems([]string{"4x2", "4x4"})
	w.Components.PadGridSizeSelect.Events.AddListener(func(ctx context.Context, record uiEvents.ComboBoxEventRecord) {
		switch record.Type {
		case uiEvents.SelectionChanged:
			layout := record.Data.(string)
			switch layout {
			case "4x2":
				w.Components.PadGrid.SetRows(2)
			case "4x4":
				w.Components.PadGrid.SetRows(4)
			}
		}
	})

	w.Window.layoutBuilder = w
	return w
}
