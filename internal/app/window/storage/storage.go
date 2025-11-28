package storage

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/component/table"
	"bitbox-editor/internal/app/component/text"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/app/font"
	"bitbox-editor/internal/app/window"
	"bitbox-editor/internal/io"
	"bitbox-editor/internal/io/drive/detect"
	"bitbox-editor/internal/logging"
	"sync"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

var log = logging.NewLogger("storage")

type StorageType int32

const (
	RemovableStorage StorageType = 0
	FixedStorage     StorageType = 1
)

type StorageLocation struct {
	StorageType StorageType
	Name        string
	Path        string
	Collapsed   bool
	Stale       bool
}

// StorageLocationsPayload is the message from the monitor
type StorageLocationsPayload struct {
	Drives []*StorageLocation
}

// StorageWindow is a window that displays available drives/storage.
type StorageWindow struct {
	*window.Window[*StorageWindow]
	eventbus.EventRouter
	component.CommandRouter

	// Child Components
	driveTable *table.TableComponent

	// Internal State
	driveMonitor     bool
	selectedLocation *StorageLocation
	driveLocations   []*StorageLocation
	customLocations  []*StorageLocation
	needsRebuild     bool

	// Background task control
	stopMonitor chan struct{}
	monitorWG   sync.WaitGroup
}

func NewStorageWindow() *StorageWindow {
	w := &StorageWindow{
		driveMonitor:     true,
		driveLocations:   make([]*StorageLocation, 0),
		customLocations:  make([]*StorageLocation, 0),
		selectedLocation: nil,
		needsRebuild:     true,
		stopMonitor:      nil,
	}

	w.Window = window.NewWindow[*StorageWindow]("Storage", "HardDrive")
	w.SetFlags(imgui.WindowFlagsMenuBar)

	w.driveTable = table.NewTableComponent(imgui.IDStr("storage-table")).
		SetNoHeader(true).
		SetFlags(imgui.TableFlagsScrollY|imgui.TableFlagsRowBg).
		SetColumns(
			table.NewTableColumn("##icon").SetFlags(imgui.TableColumnFlagsWidthFixed).SetInnerWidthOrWeight(25),
			table.NewTableColumn("Path").SetFlags(imgui.TableColumnFlagsWidthStretch),
		)

	w.Window.SetLayoutBuilder(w)

	uuid := w.UUID()

	w.EventRouter.Init(uuid)
	w.EventRouter.OnEvent(events.ComponentClickEventKey, w.onClick)

	w.CommandRouter.Init(w.Window)
	component.OnCommandTyped(&w.CommandRouter, cmdStorageSetLocations, w.onSetLocations)
	component.OnCommandTyped(&w.CommandRouter, cmdStorageHandleClick, w.onHandleClick)
	component.OnCommandTyped(&w.CommandRouter, cmdStorageSetSelected, w.onSetSelected)
	component.OnCommandTyped(&w.CommandRouter, cmdStorageSetMonitor, w.onSetMonitor)

	if w.driveMonitor {
		w.stopMonitor = make(chan struct{})
		w.startDriveMonitor()
	}

	return w
}

func (w *StorageWindow) onSetLocations(payload StorageLocationsPayload) {
	w.driveLocations = payload.Drives
	foundSelected := false
	currentSelection := w.selectedLocation
	if currentSelection != nil {
		for _, loc := range w.driveLocations {
			if loc.Path == currentSelection.Path {
				foundSelected = true
				break
			}
		}
		if !foundSelected {
			for _, loc := range w.customLocations {
				if loc.Path == currentSelection.Path {
					foundSelected = true
					break
				}
			}
		}
	}
	if !foundSelected && w.selectedLocation != nil {
		w.selectedLocation = nil
		eventbus.Bus.Publish(events.StorageDeselectedEvent{})
	}
	w.needsRebuild = true
}

func (w *StorageWindow) onHandleClick(event events.ComponentClickEvent) {
	if loc, ok := event.Data.(*StorageLocation); ok {
		w.SendUpdate(component.UpdateCmd{Type: cmdStorageSetSelected, Data: loc})
	}
}

func (w *StorageWindow) onSetSelected(loc *StorageLocation) {
	if w.selectedLocation != loc {
		w.selectedLocation = loc
		w.needsRebuild = true

		if loc != nil {
			eventbus.Bus.Publish(events.StorageActivatedEvent{
				Location: loc,
			})
		} else {
			eventbus.Bus.Publish(events.StorageDeselectedEvent{})
		}
	}
}

func (w *StorageWindow) onSetMonitor(monitor bool) {
	if w.driveMonitor != monitor {
		w.driveMonitor = monitor
		if !monitor {
			if w.stopMonitor != nil {
				select {
				case <-w.stopMonitor:
				default:
					close(w.stopMonitor)
				}
				w.stopMonitor = nil
			}
		} else {
			if w.stopMonitor == nil {
				w.stopMonitor = make(chan struct{})
				w.startDriveMonitor()
			}
		}
	}
}

// startDriveMonitor is a background task to detect drives
func (w *StorageWindow) startDriveMonitor() {
	log.Debug("Starting drive detection ...")
	w.monitorWG.Add(1)
	currentStopChan := w.stopMonitor

	go func() {
		defer w.monitorWG.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		forceCheck := true

		for {
			select {
			case <-ticker.C:
				forceCheck = true
			case <-currentStopChan:
				log.Debug("Stopping drive detection goroutine")
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}

			if !forceCheck {
				continue
			}
			forceCheck = false

			detectedDrives, err := detect.Detect()
			if err != nil {
				log.Error("Drive detection failed", zap.Error(err))
				continue
			}

			newDriveLocations := make([]*StorageLocation, 0)
			for _, d := range detectedDrives {
				presetsExist, _ := io.Exists(d + "/Presets")
				if !presetsExist {
					continue
				}
				sl := &StorageLocation{
					StorageType: RemovableStorage,
					Name:        d,
					Path:        d,
					Stale:       false,
				}
				newDriveLocations = append(newDriveLocations, sl)
			}

			// Send Update Command (Drives Only)
			payload := StorageLocationsPayload{
				Drives: newDriveLocations,
			}
			cmd := component.UpdateCmd{Type: cmdStorageSetLocations, Data: payload}
			w.SendUpdate(cmd)
		}
	}()
}

// rebuildTableRows creates row components
func (w *StorageWindow) rebuildTableRows() {
	if w.driveTable == nil {
		log.Error("rebuildTableRows called with nil driveTable")
		return
	}

	drives := w.driveLocations
	customs := w.customLocations
	selected := w.selectedLocation
	totalRows := len(drives) + len(customs)
	rows := make([]*table.TableRowComponent, 0, totalRows)

	createRow := func(loc *StorageLocation, iconKey string) *table.TableRowComponent {
		location := loc
		isSelected := selected != nil && selected.Path == location.Path

		selectableText := text.NewText(location.Path).
			SetSelectable(true)

		selectableText.SetDragDropData("", location)
		selectableText.SetSelected(isSelected)

		tr := table.NewTableRow(
			imgui.IDStr(location.Path),
			text.NewText(font.Icon(iconKey)),
			selectableText,
		)
		return tr
	}

	for _, loc := range drives {
		rows = append(rows, createRow(loc, "HardDrive"))
	}

	for _, loc := range customs {
		rows = append(rows, createRow(loc, "Folder"))
	}

	w.driveTable.SetRows(rows...)
}

func (w *StorageWindow) Menu() {
	monitor := w.driveMonitor

	scanIcon := font.Icon("Scan")
	if monitor {
		scanIcon = font.Icon("ScanEye")
	}

	if imgui.BeginMenuBar() {
		if imgui.Button(font.Icon("FolderPlus")) {
			// TODO: Open dialog to select folder, then send CmdStorageAddCustom
		}
		imgui.SameLine()
		if imgui.Button(font.Icon("FolderOpen")) {
			// TODO: Open selected location in OS file explorer?
		}
		imgui.SameLine()

		if imgui.Button(scanIcon) {
			// Send command to toggle state
			cmd := component.UpdateCmd{Type: cmdStorageSetMonitor, Data: !monitor}
			w.SendUpdate(cmd)
		}
		imgui.EndMenuBar()
	}
}

func (w *StorageWindow) onClick(event events.Event) {
	w.SendUpdate(component.UpdateCmd{Type: cmdStorageHandleClick, Data: event})
}

func (w *StorageWindow) Layout() {
	w.EventRouter.ProcessEvents()
	w.Window.ProcessUpdates()

	isLoading := w.Loading()
	drives := w.driveLocations
	customs := w.customLocations
	needsRebuild := w.needsRebuild

	if needsRebuild {
		w.rebuildTableRows()
		w.needsRebuild = false
	}

	if isLoading {
		imgui.Text("Scanning drives...")
		return
	}

	if len(drives) == 0 && len(customs) == 0 {
		imgui.Text("No preset locations found.")
		return
	}

	if w.driveTable != nil {
		w.driveTable.Build()
	} else {
		imgui.Text("Drive table error.")
	}

}

func (w *StorageWindow) Destroy() {
	// Signal monitor to stop and wait for it
	if w.stopMonitor != nil {
		select {
		case <-w.stopMonitor:
		default:
			close(w.stopMonitor)
		}
	}

	log.Debug("Waiting for drive monitor to stop...")

	w.monitorWG.Wait()

	log.Debug("Drive monitor stopped.")

	// Unsubscribe from all events
	w.EventRouter.Destroy()

	// Destroy child components
	if w.driveTable != nil {
		w.driveTable.Destroy()
	}
}
