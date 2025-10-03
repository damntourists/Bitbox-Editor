package windows

import (
	"bitbox-editor/lib/events"
	"bitbox-editor/lib/preset"
	"bitbox-editor/ui/component"
	uiEvents "bitbox-editor/ui/events"
	"bitbox-editor/ui/fonts"
	"context"
	"os"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maniartech/signals"
)

type PresetListWindow struct {
	*Window

	presetLocation *StorageLocation

	table *component.TableComponent

	selectedPreset *preset.Preset
	presets        []*preset.Preset

	rowCache map[string]*component.TableRowComponent

	loading bool

	Events signals.Signal[events.PresetEventRecord]
}

func (w *PresetListWindow) SetPresetLocation(location *StorageLocation) {
	w.loading = true
	w.presets = make([]*preset.Preset, 0)
	w.presetLocation = location
	go func() {
		entries, err := os.ReadDir(w.presetLocation.path + "/Presets/")
		if err != nil {
			w.loading = false
			return
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			p := preset.NewPreset(
				entry.Name(),
				w.presetLocation.path+"/Presets/"+entry.Name(),
			)

			w.presets = append(w.presets, p)
		}
	}()

}

func (w *PresetListWindow) Menu() {
	if imgui.BeginMenuBar() {
		if imgui.Button(fonts.Icon("ListPlus")) {

		}

		imgui.EndMenuBar()
	}
}

func (w *PresetListWindow) Layout() {
	if w.presetLocation == nil {
		return
	}

	rows := make([]*component.TableRowComponent, 0)

	for _, p := range w.presets {
		isSelected := p == w.selectedPreset

		selectable := component.Text(p.Name).
			Selectable(true).
			SetSelected(isSelected)
		selectable.SetData(p)

		selectable.MouseEvents().AddListener(
			func(ctx context.Context, e uiEvents.MouseEventRecord) {
				switch e.Type {
				case uiEvents.DoubleClicked:
					w.selectedPreset = p
					w.Events.Emit(
						context.Background(),
						events.PresetEventRecord{
							Type: events.LoadPreset,
							Data: p,
						},
					)
				}
			},
			selectable.IDStr(),
		)

		tr := component.TableRow(
			imgui.IDStr(p.Path),
			selectable,
		)

		rows = append(rows, tr)
	}

	w.table.
		Columns(
			component.NewTableColumn("name").
				Flags(
					imgui.TableColumnFlagsWidthStretch,
				),
		).
		Rows(rows...).
		Layout()

}

func NewPresetListWindow() *PresetListWindow {
	sw := &PresetListWindow{
		Window: NewWindow("Presets", "ListMusic"),
		table: component.
			NewTableComponent(
				imgui.IDStr("presets-table"),
			).
			NoHeader(true),
		Events: signals.New[events.PresetEventRecord](),
	}

	sw.Window.layoutBuilder = sw

	return sw
}
