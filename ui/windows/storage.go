package windows

import (
	"bitbox-editor/lib/events"
	"bitbox-editor/lib/io"
	"bitbox-editor/lib/io/drive/detect"
	"bitbox-editor/lib/logging"
	"bitbox-editor/ui/component"
	uiEvents "bitbox-editor/ui/events"
	"bitbox-editor/ui/fonts"
	"context"
	"os"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maniartech/signals"
)

var log = logging.NewLogger("presets")

type StorageType int32

const (
	RemovableStorage StorageType = 0
	FixedStorage     StorageType = 1
)

type StorageLocation struct {
	storageType StorageType

	name      string
	path      string
	collapsed bool
	stale     bool
}

type StorageWindow struct {
	*Window

	driveMonitor bool

	driveTable *component.TableComponent

	selectedLocation *StorageLocation
	driveLocations   []*StorageLocation
	customLocations  []*StorageLocation

	Events signals.Signal[events.StorageEventRecord]
}

func (w *StorageWindow) Menu() {
	scanIcon := fonts.Icon("ScanEye")
	if !w.driveMonitor {
		scanIcon = fonts.Icon("Scan")
	}

	if imgui.BeginMenuBar() {
		if imgui.Button(fonts.Icon("FolderPlus")) {

		}
		imgui.SameLine()

		if imgui.Button(fonts.Icon("FolderOpen")) {

		}
		imgui.SameLine()

		if imgui.Button(scanIcon) {
			w.driveMonitor = !w.driveMonitor
		}

		imgui.EndMenuBar()
	}
}

func (w *StorageWindow) Layout() {
	if len(w.driveLocations) == 0 && len(w.customLocations) == 0 {
		imgui.Text("No presets locations found")
		return
	}

	var rows = make([]*component.TableRowComponent, 0)
	for _, p := range w.driveLocations {
		isSelected := p == w.selectedLocation

		selectable := component.Text(p.path).
			Selectable(true).
			SetSelected(isSelected)
		selectable.SetData(p)
		selectable.MouseEvents.AddListener(
			func(ctx context.Context, e uiEvents.MouseEventRecord) {
				switch e.Type {
				case uiEvents.Clicked:
					w.selectedLocation = p
					w.Events.Emit(
						context.Background(),
						events.StorageEventRecord{
							Type: events.StorageActivatedEvent,
							Data: p,
						},
					)
				}
			},
			selectable.IDStr())

		tr := component.TableRow(
			imgui.IDStr(p.path),
			component.Text(fonts.Icon("HardDrive")),
			selectable,
		)

		rows = append(rows, tr)
	}

	for _, p := range w.customLocations {
		isSelected := p == w.selectedLocation

		selectable := component.Text(p.path).
			Selectable(true).
			SetSelected(isSelected)
		selectable.SetData(p)
		selectable.MouseEvents.AddListener(
			func(ctx context.Context, e uiEvents.MouseEventRecord) {
				switch e.Type {
				case uiEvents.Clicked:
					w.selectedLocation = p
					w.Events.Emit(
						context.Background(),
						events.StorageEventRecord{
							Type: events.StorageActivatedEvent,
							Data: p,
						},
					)
				}
			},
			selectable.IDStr())

		tr := component.TableRow(
			imgui.IDStr(p.path),
			component.Text(fonts.Icon("Folder")),
			selectable,
		)

		rows = append(rows, tr)
	}

	w.driveTable.
		Columns(
			component.TableColumn("##icon").
				Flags(
					imgui.TableColumnFlagsWidthFixed,
				),
			component.TableColumn("Name").
				Flags(
					imgui.TableColumnFlagsWidthStretch,
				),
		).
		Rows(rows...).
		Layout()

}

func NewStorageWindow() *StorageWindow {
	sw := &StorageWindow{
		Window:       NewWindow("Storage", "HardDrive", NewWindowConfig()),
		driveMonitor: true,
		driveTable: component.Table(imgui.IDStr("storage-table")).
			NoHeader(true),
		driveLocations:  make([]*StorageLocation, 0),
		customLocations: make([]*StorageLocation, 0),
		Events:          signals.New[events.StorageEventRecord](),
	}
	sw.Window.layoutBuilder = sw

	// Start drive detect goroutine
	go func() {
		log.Debug("Starting drive detection")
		for {
			if sw.driveMonitor {
				if drives, err := detect.Detect(); err == nil {
					for _, d := range drives {
						presetsExist, _ := io.Exists(d + "/Presets")
						if !presetsExist {
							continue
						}

						// Check if we have new drives
						exists := false
						for _, sd := range sw.driveLocations {
							if sd.path == d {
								exists = true
							}

							// also check if stale
							_, err := os.Stat(sd.path)
							if err != nil {
								if !sd.stale {

									sw.Events.Emit(
										context.Background(),
										events.StorageEventRecord{
											Type: events.StorageUnmountedEvent,
											Data: sd,
										},
									)

									sd.stale = true
								}
							} else {
								if sd.stale {
									sw.Events.Emit(
										context.Background(),
										events.StorageEventRecord{
											Type: events.StorageMountedEvent,
											Data: sd,
										},
									)
								}
								sd.stale = false
							}
						}

						if !exists {
							// Create new location
							sl := &StorageLocation{
								storageType: RemovableStorage,
								name:        d,
								path:        d,
								collapsed:   false,
								stale:       false,
							}
							sw.driveLocations = append(sw.driveLocations, sl)

							sw.Events.Emit(
								context.Background(),
								events.StorageEventRecord{
									Type: events.StorageMountedEvent,
									Data: sl,
								},
							)
						}
					}

				} else {
					//log.Debug(err.Error())
				}
			}

			time.Sleep(5 * time.Second)
		}
	}()

	return sw
}
