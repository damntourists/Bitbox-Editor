package presetlist

/*
┍━━━━━━━━━━━━━━━━━━━╳┑
│ Preset List Window │
└────────────────────┘
*/

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/component/table"
	"bitbox-editor/internal/app/component/text"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/app/font"
	"bitbox-editor/internal/app/window"
	"bitbox-editor/internal/app/window/storage"
	"bitbox-editor/internal/logging"
	"bitbox-editor/internal/preset"
	"os"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

var log = logging.NewLogger("presetlist")

type PresetListWindow struct {
	*window.Window[*PresetListWindow]
	eventbus.EventRouter
	component.CommandRouter

	// Child Components
	Components struct {
		PresetTable *table.TableComponent
	}

	// Internal State
	presets        []*preset.Preset
	selectedPreset *preset.Preset
	presetLocation *storage.StorageLocation
	loading        bool
}

func NewPresetListWindow() *PresetListWindow {
	w := &PresetListWindow{
		presets:        make([]*preset.Preset, 0),
		presetLocation: nil,
		selectedPreset: nil,
		loading:        false,
	}

	w.Window = window.NewWindow[*PresetListWindow]("Presets", "ListMusic")
	w.SetFlags(imgui.WindowFlagsMenuBar)
	w.Window.SetLayoutBuilder(w)

	w.Components.PresetTable = table.NewTableComponent(imgui.IDStr("presets-table")).
		SetNoHeader(true).
		SetFlags(
			imgui.TableFlagsScrollY |
				//imgui.TableFlagsRowBg |
				imgui.TableFlagsSortable,
		).
		SetColumns(
			table.NewTableColumn("Name").
				SetFlags(
					imgui.TableColumnFlagsWidthStretch |
						imgui.TableColumnFlagsDefaultSort,
				),
		)

	uuid := w.UUID()

	w.EventRouter.Init(uuid)
	w.OnEvent(events.StorageActivatedEventKey, w.onStorageActivated)
	w.OnEvent(events.ComponentClickEventKey, w.onRowClick)

	w.CommandRouter.Init(w.Window)
	component.OnCommandTyped(&w.CommandRouter, cmdPresetListSetLocation, w.onSetLocation)
	component.OnCommandTyped(&w.CommandRouter, cmdPresetListSetLoading, w.onSetLoading)
	component.OnCommandTyped(&w.CommandRouter, cmdPresetListUpdateList, w.onUpdateList)
	component.OnCommandTyped(&w.CommandRouter, cmdPresetListSetSelected, w.onSetSelected)
	component.OnCommandTyped(&w.CommandRouter, cmdHandleRowClick, w.onHandleRowClick)

	return w
}

func (w *PresetListWindow) onStorageActivated(event events.Event) {
	w.SendUpdate(component.UpdateCmd{Type: cmdPresetListSetLocation, Data: event})
}

func (w *PresetListWindow) onRowClick(event events.Event) {
	w.SendUpdate(component.UpdateCmd{Type: cmdHandleRowClick, Data: event})
}

func (w *PresetListWindow) onSetLocation(event events.StorageActivatedEvent) {
	if loc, ok := event.Location.(*storage.StorageLocation); ok {
		if w.presetLocation == nil || w.presetLocation.Path != loc.Path {
			w.presetLocation = loc
			w.presets = nil
			w.selectedPreset = nil
			w.rebuildTableRows()
			w.startScan()
		}
	}
}

func (w *PresetListWindow) onSetLoading(isLoading bool) {
	w.loading = isLoading
}

func (w *PresetListWindow) onUpdateList(newList []*preset.Preset) {
	w.presets = newList
	foundSelected := false
	if w.selectedPreset != nil {
		for _, p := range w.presets {
			if p == w.selectedPreset {
				foundSelected = true
				break
			}
		}
	}
	if !foundSelected {
		w.selectedPreset = nil
	}
	w.rebuildTableRows()
}

func (w *PresetListWindow) onSetSelected(p *preset.Preset) {
	if w.selectedPreset != p {
		w.selectedPreset = p
		w.rebuildTableRows()
	}
}

func (w *PresetListWindow) onHandleRowClick(event events.ComponentClickEvent) {
	if p, ok := event.Data.(*preset.Preset); ok {
		w.SendUpdate(component.UpdateCmd{Type: cmdPresetListSetSelected, Data: p})
		if event.IsDoubleClick {
			eventbus.Bus.Publish(events.PresetLoadEvent{
				Preset: p,
			})
		}
	}
}

func (w *PresetListWindow) rebuildTableRows() {
	if w.Components.PresetTable == nil {
		return
	}

	rows := make([]*table.TableRowComponent, 0, len(w.presets))

	for _, p := range w.presets {
		preset := p

		isSelected := preset == w.selectedPreset

		selectableText := text.NewText(preset.Name).
			SetSelectable(true).
			SetSelected(isSelected).
			DisableHoverAnimations()

		selectableText.SetDragDropData("", preset)

		tr := table.NewTableRow(
			imgui.IDStr(preset.Path),
			selectableText,
		)
		rows = append(rows, tr)
	}

	w.Components.PresetTable.SetRows(rows...)
}

func (w *PresetListWindow) startScan() {
	if w.presetLocation == nil {
		return
	}

	w.SendUpdate(component.UpdateCmd{Type: cmdPresetListSetLoading, Data: true})

	path := w.presetLocation.Path

	go func() {
		presetList := make([]*preset.Preset, 0)
		scanPath := path + "/Presets/"
		entries, err := os.ReadDir(scanPath)

		if err != nil {
			log.Error("Failed to read preset directory", zap.Error(err), zap.String("path", scanPath))
			// Send empty list
			w.SendUpdate(component.UpdateCmd{Type: cmdPresetListUpdateList, Data: presetList})
		} else {
			for _, entry := range entries {
				if entry.IsDir() {
					p := preset.NewPreset(
						entry.Name(),
						scanPath+entry.Name(),
					)
					presetList = append(presetList, p)
				}
			}
			// Send updated list
			w.SendUpdate(component.UpdateCmd{Type: cmdPresetListUpdateList, Data: presetList})
		}

		w.SendUpdate(component.UpdateCmd{Type: cmdPresetListSetLoading, Data: false})
	}()
}

func (w *PresetListWindow) SetPresetLocation(location *storage.StorageLocation) *PresetListWindow {
	// Create a *copy* of the location to be safe
	locCopy := *location
	w.SendUpdate(component.UpdateCmd{Type: cmdPresetListSetLocation, Data: &locCopy})
	return w
}

func (w *PresetListWindow) Menu() {
	if imgui.BeginMenuBar() {
		if imgui.Button(font.Icon("ListPlus")) {
			// TODO: Create new preset
		}
		imgui.EndMenuBar()
	}
}

func (w *PresetListWindow) Layout() {
	w.EventRouter.ProcessEvents()
	w.Window.ProcessUpdates()

	isLoading := w.loading
	presetLoc := w.presetLocation

	if presetLoc == nil {
		imgui.Text("No preset location set.")
		return
	}
	if isLoading {
		imgui.Text("Loading presets...")
		return
	}

	if w.Components.PresetTable != nil {
		w.Components.PresetTable.Build()
	} else {
		imgui.Text("Preset table not initialized.")
	}
}

func (w *PresetListWindow) Destroy() {
	// Unsubscribe from all events
	w.EventRouter.Destroy()

	// Destroy child components
	if w.Components.PresetTable != nil {
		w.Components.PresetTable.Destroy()
	}

	// Call base destroy
	w.Window.Destroy()
}
