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

	// Child Components
	Components struct {
		PresetTable *table.TableComponent
	}

	// Internal State
	presets        []*preset.Preset
	selectedPreset *preset.Preset
	presetLocation *storage.StorageLocation
	loading        bool

	filteredEventSub *eventbus.FilteredSubscription
}

func NewPresetListWindow() *PresetListWindow {
	w := &PresetListWindow{
		presets:        make([]*preset.Preset, 0),
		presetLocation: nil,
		selectedPreset: nil,
		loading:        false,
	}

	w.Window = window.NewWindow[*PresetListWindow]("Presets", "ListMusic", w.handleUpdate)
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

	bus := eventbus.Bus
	uuid := w.UUID()

	// Create filtered subscription for all events
	w.filteredEventSub = eventbus.NewFilteredSubscription(uuid, 20)
	w.filteredEventSub.SubscribeMultiple(
		bus,
		events.StorageActivatedEventKey,
		events.ComponentClickEventKey,
	)

	return w
}

// drainEvents translates global bus events into local commands
func (w *PresetListWindow) drainEvents() {
	if w.filteredEventSub != nil {
		for {
			select {
			case event := <-w.filteredEventSub.Events():
				var cmd component.UpdateCmd
				switch event.Type() {
				case events.StorageActivatedEventKey:
					cmd = component.UpdateCmd{Type: cmdPresetListSetLocation, Data: event}
				case events.ComponentClickEventKey:
					cmd = component.UpdateCmd{Type: cmdHandleRowClick, Data: event}
				}
				if cmd.Type != 0 {
					w.SendUpdate(cmd)
				}
			default:
				// No more events
				return
			}
		}
	}
}

func (w *PresetListWindow) handleUpdate(cmd component.UpdateCmd) {
	if w.Window.HandleGlobalUpdate(cmd) {
		return
	}

	switch cmd.Type {
	case cmdPresetListSetLocation:
		if event, ok := cmd.Data.(events.StorageEventRecord); ok {
			if loc, ok := event.Data.(*storage.StorageLocation); ok {
				if w.presetLocation == nil || w.presetLocation.Path != loc.Path {
					w.presetLocation = loc
					w.presets = nil
					w.selectedPreset = nil
					w.rebuildTableRows()
					w.startScan()
				}
			}
		}

	case cmdPresetListSetLoading:
		if isLoading, ok := cmd.Data.(bool); ok {
			w.loading = isLoading
		}

	case cmdPresetListUpdateList:
		if newList, ok := cmd.Data.([]*preset.Preset); ok {
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

	case cmdPresetListSetSelected:
		if p, ok := cmd.Data.(*preset.Preset); ok {
			if w.selectedPreset != p {
				w.selectedPreset = p
				w.rebuildTableRows()
			}
		} else if cmd.Data == nil {
			if w.selectedPreset != nil {
				w.selectedPreset = nil
				w.rebuildTableRows()
			}
		}

	case cmdHandleRowClick:
		if event, ok := cmd.Data.(events.MouseEventRecord); ok {
			if p, ok := event.Data.(*preset.Preset); ok {
				w.SendUpdate(component.UpdateCmd{Type: cmdPresetListSetSelected, Data: p})
				if event.EventType == events.ComponentDoubleClickedEvent {
					eventbus.Bus.Publish(events.PresetEventRecord{
						EventType: events.PresetLoadEvent,
						Data:      p,
					})
				}
			}
		}

	default:
		log.Warn("PresetListWindow unhandled update", zap.Any("cmd", cmd))
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
	w.drainEvents()
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
	// Unsubscribe from filtered subscriptions
	if w.filteredEventSub != nil {
		w.filteredEventSub.Unsubscribe()
	}

	// Destroy child components
	if w.Components.PresetTable != nil {
		w.Components.PresetTable.Destroy()
	}

	// Call base destroy
	w.Window.Destroy()
}
