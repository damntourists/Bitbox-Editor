package windows

import (
	"bitbox-editor/lib/io/drive/detect"
	"bitbox-editor/lib/logging"
	"bitbox-editor/ui/fonts"
	"fmt"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
)

var log = logging.NewLogger("presets")

type StorageLocation struct {
	name      string
	path      string
	collapsed bool
	stale     bool
}

type StorageWindow struct {
	*Window

	driveMonitor bool

	driveLocations  []StorageLocation
	customLocations []StorageLocation
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

	for _, p := range w.driveLocations {
		imgui.Text(fonts.Icon("HardDrive") + p.name)
	}
	for _, p := range w.customLocations {
		imgui.Text(fonts.Icon("Folder") + p.name)
	}
}

func NewStorageWindow() *StorageWindow {
	sw := &StorageWindow{
		Window:          NewWindow("Storage", "HardDrive", NewWindowConfig()),
		driveMonitor:    true,
		driveLocations:  []StorageLocation{},
		customLocations: []StorageLocation{},
	}
	sw.Window.layoutBuilder = sw

	// Start drive detect goroutine
	go func() {
		log.Debug("Starting drive detection")

		for {
			if sw.driveMonitor {
				if drives, err := detect.Detect(); err == nil {
					log.Debug(fmt.Sprintf("%d USB Devices Found", len(drives)))
					for _, d := range drives {
						log.Debug(d)
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
