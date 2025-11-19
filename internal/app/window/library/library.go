package library

/*
┍━━━━━━━━━━━━━━━╳┑
│ Library Window │
└────────────────┘
*/

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/component/miniwave"
	"bitbox-editor/internal/app/component/table"
	"bitbox-editor/internal/app/component/text"
	"bitbox-editor/internal/app/component/tree"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/app/font"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/app/window"
	"bitbox-editor/internal/app/window/storage"
	"bitbox-editor/internal/audio"
	"bitbox-editor/internal/io"
	"bitbox-editor/internal/logging"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

type UpdateCmd = component.UpdateCmd
type UpdateCmdType = component.UpdateCmdType
type localCommand int

const (
	cmdLibSetStorageLoc localCommand = iota
	cmdLibSetScanning
	cmdLibSetFSTree
	cmdLibSetTreeRows
	cmdLibSetSearchQuery
	cmdHandleScanEvent
)

var log = logging.NewLogger("library")

// hashString returns a simple hash of a string for use as an ImGui ID
func hashString(s string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		hash ^= uint32(s[i])
		hash *= 16777619
	}
	return hash
}

// LibraryWindow is a window that displays a library of audio files.
type LibraryWindow struct {
	*window.Window[*LibraryWindow]

	storageLoc    *storage.StorageLocation
	fsTree        *io.FSTree
	searchQuery   string
	searchResults []*io.FSEntry
	isScanning    bool

	// child component
	Components struct {
		Tree *tree.TreeComponent
	}

	filteredEventSub *eventbus.FilteredSubscription
}

func (w *LibraryWindow) createTreeRowsFromData(data []tree.TreeRowData) []*tree.TreeRowComponent {
	rows := make([]*tree.TreeRowComponent, len(data))

	for i, rowData := range data {
		displayName := rowData.Icon + " " + rowData.Name
		textHash := imgui.ID(hashString(rowData.Path + ":text"))
		textComp := text.NewTextWithID(textHash, displayName)
		layoutComponents := []component.ComponentType{textComp}

		if rowData.IsAudio {
			pathHash := imgui.ID(hashString(rowData.Path + ":waveform"))
			waveform := miniwave.NewMiniWaveform(pathHash, rowData.Path)
			layoutComponents = append(layoutComponents, waveform)
		} else if !rowData.IsDir {
			layoutComponents = append(layoutComponents, component.NewDummy())
		} else {
			layoutComponents = append(layoutComponents, component.NewDummy())
		}

		if rowData.IsAudio {
			durationGetter := func() string {
				cache := audio.GetGlobalAsyncCache()
				snapshot := cache.GetSnapshot(rowData.Path)
				if snapshot != nil && snapshot.MetadataLoaded && snapshot.SampleRate > 0 {
					duration := float64(snapshot.NumSamples) / float64(snapshot.SampleRate)
					if duration > 0 {
						return formatDuration(duration)
					}
				}
				if snapshot == nil {
					cache.RequestLoad(rowData.Path, audio.LoadMetadataOnly)
				}
				return ""
			}
			layoutComponents = append(layoutComponents, text.NewDynamicText(durationGetter))
		} else {
			layoutComponents = append(layoutComponents, text.NewText(rowData.DurationText))
		}

		layoutComponents = append(layoutComponents, text.NewText(rowData.SizeText))

		row := tree.NewTreeRow(
			imgui.ID(hashString(rowData.Path)),
			layoutComponents...,
		).Flags(
			imgui.TreeNodeFlagsSpanAllColumns |
				imgui.TreeNodeFlagsDrawLinesToNodes |
				imgui.TreeNodeFlagsLabelSpanAllColumns,
		)

		if rowData.DragDropType != "" {
			row.SetDragDropData(rowData.DragDropType, rowData.Path)
			tooltipName := rowData.Name
			tooltipIcon := rowData.Icon
			row.SetDragDropTooltipFn(func() {
				imgui.Text(fmt.Sprintf("%s %s", tooltipIcon, tooltipName))
			})
		}

		if len(rowData.Children) > 0 {
			childData := make([]tree.TreeRowData, len(rowData.Children))
			for j, child := range rowData.Children {
				childData[j] = child
			}

			row.SetChildrenData(
				childData,
				w.createTreeRowsFromData,
			)
		}

		rows[i] = row
	}
	return rows
}

func (w *LibraryWindow) createInitialTopLevelRows(rowData []tree.TreeRowData) []*tree.TreeRowComponent {
	initialData := make([]tree.TreeRowData, len(rowData))
	for i, data := range rowData {
		initialData[i] = data
	}
	return w.createTreeRowsFromData(initialData)
}

// drainEvents translates global bus events into local commands
func (w *LibraryWindow) drainEvents() {
	if w.filteredEventSub != nil {
		for {
			select {
			case event := <-w.filteredEventSub.Events():
				if e, ok := event.(events.LibraryScanEventRecord); ok {
					w.SendUpdate(UpdateCmd{Type: cmdHandleScanEvent, Data: e})
				}
			default:
				return
			}
		}
	}
}

func (w *LibraryWindow) handleUpdate(cmd component.UpdateCmd) {
	if w.Window.HandleGlobalUpdate(cmd) {
		return
	}

	switch cmd.Type {
	case cmdLibSetStorageLoc:
		if loc, ok := cmd.Data.(*storage.StorageLocation); ok {
			if w.storageLoc == nil || w.storageLoc.Path != loc.Path {
				w.storageLoc = loc
				w.fsTree = nil
				w.searchResults = nil
				w.searchQuery = ""
				w.startScan()
			}
		}

	case cmdLibSetScanning:
		if scanning, ok := cmd.Data.(bool); ok {
			w.isScanning = scanning
		}

	case cmdLibSetFSTree:
		if tree, ok := cmd.Data.(*io.FSTree); ok {
			w.fsTree = tree
			go w.buildAndSetTreeRowsInBackground()
		}

	case cmdLibSetTreeRows:
		if rowData, ok := cmd.Data.([]tree.TreeRowData); ok {
			if w.Components.Tree != nil {
				componentRows := w.createInitialTopLevelRows(rowData)
				w.Components.Tree.Rows(componentRows...)
			}
		}

	case cmdLibSetSearchQuery:
		if query, ok := cmd.Data.(string); ok {
			if w.searchQuery != query {
				w.searchQuery = query
				go w.performSearchAndBuildRowsInBackground()
			}
		}

	case cmdHandleScanEvent:
		if event, ok := cmd.Data.(events.LibraryScanEventRecord); ok {
			switch event.EventType {
			case events.LibraryScanStartedEvent:
				log.Debug("Library scan started", zap.String("path", event.Path))
			case events.LibraryScanProgressEvent:
				log.Debug("Library scan progress", zap.Float64("progress", event.Progress))
			case events.LibraryScanCompletedEvent:
				log.Debug("Library scan completed",
					zap.String("path", event.Path),
					zap.Int("fileCount", event.FileCount))
			case events.LibraryScanFailedEvent:
				log.Error("Library scan failed",
					zap.String("path", event.Path),
					zap.Error(event.Error))
			}
		}

	default:
		log.Warn("LibraryWindow unhandled update", zap.Any("cmd", cmd))
	}
}

func (w *LibraryWindow) buildAndSetTreeRowsInBackground() {
	fstree := w.fsTree
	search := w.searchResults

	var rowData []tree.TreeRowData
	if fstree == nil || fstree.Root == nil {
		rowData = []tree.TreeRowData{}
	} else if search != nil {
		matchedPaths := make(map[string]bool)
		for _, entry := range search {
			matchedPaths[entry.Path] = true
			current := entry.Parent
			for current != nil {
				matchedPaths[current.Path] = true
				current = current.Parent
			}
		}
		rowData = w.buildRowDataFromEntry(fstree.Root, matchedPaths)
	} else {
		rowData = w.buildRowDataFromEntry(fstree.Root, nil)
	}

	cmd := component.UpdateCmd{Type: cmdLibSetTreeRows, Data: rowData}
	w.SendUpdate(cmd)
}

func (w *LibraryWindow) startScan() {
	if w.storageLoc == nil {
		log.Warn("startScan called with nil storageLoc")
		return
	}

	cmd := UpdateCmd{Type: cmdLibSetScanning, Data: true}
	w.SendUpdate(cmd)

	path := w.storageLoc.Path

	eventbus.Bus.Publish(events.LibraryScanEventRecord{
		EventType: events.LibraryScanStartedEvent,
		Path:      path,
	})

	go func() {
		tree := io.NewFSTree(path)
		err := tree.ScanDirectory(path, ".wav", ".WAV")

		if err != nil {
			log.Error("Failed to scan directory", zap.Error(err), zap.String("path", path))

			eventbus.Bus.Publish(events.LibraryScanEventRecord{
				EventType: events.LibraryScanFailedEvent,
				Path:      path,
				Error:     err,
			})

			doneCmd := UpdateCmd{Type: cmdLibSetScanning, Data: false}
			w.SendUpdate(doneCmd)
		} else {
			log.Debug(
				"Finished directory scan",
				zap.String("path", path),
				zap.Int("fileCount", tree.GetFileCount()),
			)

			eventbus.Bus.Publish(
				events.LibraryScanEventRecord{
					EventType: events.LibraryScanCompletedEvent,
					Path:      path,
					FileCount: tree.GetFileCount(),
					Progress:  1.0,
				},
			)

			treeCmd := UpdateCmd{Type: cmdLibSetFSTree, Data: tree}
			w.SendUpdate(treeCmd)

			doneCmd := UpdateCmd{Type: cmdLibSetScanning, Data: false}
			w.SendUpdate(doneCmd)
		}
	}()
}

func formatFileSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	kb := float64(bytes) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.1f KB", kb)
	}
	mb := kb / 1024
	return fmt.Sprintf("%.1f MB", mb)
}

func formatDuration(seconds float64) string {
	minutes := int(seconds / 60)
	secs := seconds - float64(minutes*60)
	return fmt.Sprintf("%02d:%05.2f", minutes, secs)
}

func (w *LibraryWindow) buildRowDataFromEntry(entry *io.FSEntry, matchedPaths map[string]bool) []tree.TreeRowData {
	rowData := make([]tree.TreeRowData, 0, len(entry.Children))

	for _, child := range entry.Children {
		if matchedPaths != nil && !matchedPaths[child.Path] {
			continue
		}

		var icon string = font.Icon("FileAudio")
		if child.IsDir {
			icon = font.Icon("Folder")
		}

		sizeText := ""
		if !child.IsDir && child.Size > 0 {
			sizeText = formatFileSize(child.Size)
		}

		durationText := ""
		isAudio := !child.IsDir && audio.IsAudioFile(child.Path)
		if isAudio {
			cache := audio.GetGlobalAsyncCache()
			snapshot := cache.GetSnapshot(child.Path)
			if snapshot != nil && snapshot.MetadataLoaded && snapshot.SampleRate > 0 {
				duration := float64(snapshot.NumSamples) / float64(snapshot.SampleRate)
				if duration > 0 {
					durationText = formatDuration(duration)
				}
			}
		}

		dragDropType := ""
		if isAudio {
			dragDropType = "audio/wav-path"
		} else if child.IsDir {
			dragDropType = "file/folder-path"
		}

		data := tree.TreeRowData{
			Path:         child.Path,
			Name:         child.Name,
			Icon:         icon,
			IsDir:        child.IsDir,
			IsAudio:      isAudio,
			DurationText: durationText,
			SizeText:     sizeText,
			DragDropType: dragDropType,
		}

		if child.IsDir {
			childData := w.buildRowDataFromEntry(child, matchedPaths)
			if len(childData) > 0 {
				data.Children = childData
			} else if matchedPaths == nil {
				continue
			} else {
				continue
			}
		}

		rowData = append(rowData, data)
	}
	return rowData
}

func (w *LibraryWindow) performSearchAndBuildRowsInBackground() {
	fstree := w.fsTree
	query := w.searchQuery

	if fstree == nil {
		return
	}

	var results []*io.FSEntry
	if query == "" {
		results = nil
	} else {
		results = fstree.Search(query)
	}

	var rowData []tree.TreeRowData
	if results != nil {
		matchedPaths := make(map[string]bool)
		for _, entry := range results {
			matchedPaths[entry.Path] = true
			current := entry.Parent
			for current != nil {
				matchedPaths[current.Path] = true
				current = current.Parent
			}
		}
		rowData = w.buildRowDataFromEntry(fstree.Root, matchedPaths)
	} else {
		rowData = w.buildRowDataFromEntry(fstree.Root, nil)
	}

	cmd := component.UpdateCmd{Type: cmdLibSetTreeRows, Data: rowData}
	w.SendUpdate(cmd)
}

func (w *LibraryWindow) hasMatchingDescendants(entry *io.FSEntry, matchedPaths map[string]bool) bool {
	for _, child := range entry.Children {
		if !child.IsDir && matchedPaths[child.Path] {
			return true
		}
		if child.IsDir && w.hasMatchingDescendants(child, matchedPaths) {
			return true
		}
	}
	return false
}

func (w *LibraryWindow) SearchFiles(query string) []*io.FSEntry {
	if w.fsTree == nil || w.fsTree.Root == nil {
		return nil
	}
	return w.fsTree.Search(query)
}

func (w *LibraryWindow) SetStorageLocation(loc *storage.StorageLocation) *LibraryWindow {
	cmd := component.UpdateCmd{Type: cmdLibSetStorageLoc, Data: loc}
	w.SendUpdate(cmd)
	return w
}

func (w *LibraryWindow) Menu() {
	if imgui.BeginMenuBar() {
		if imgui.Button(font.Icon("RefreshCw")) {
			if w.storageLoc != nil {
				cmd := component.UpdateCmd{Type: cmdLibSetStorageLoc, Data: w.storageLoc}
				w.SendUpdate(cmd)
			}
		}
		imgui.EndMenuBar()
	}
}

func (w *LibraryWindow) Layout() {
	w.drainEvents()
	w.Window.ProcessUpdates()

	storageLoc := w.storageLoc
	isScanning := w.isScanning
	fsTree := w.fsTree
	searchQuery := w.searchQuery

	if storageLoc == nil {
		imgui.Text("No storage location selected")
		return
	}

	if isScanning {
		component.WavyText(
			"Scanning...",
			theme.GetCurrentColormap(),
			true,
			imgui.ColorU32Vec4(imgui.Vec4{X: 0, Y: 0, Z: 0, W: 1}),
			imgui.NewVec2(2.2, 1.0),
			imgui.NewVec2(1.0, 1.0),
			imgui.NewVec2(8.0, 8.0),
			0.2,
		)
		return
	}

	if w.fsTree != nil {
		imgui.Text(fmt.Sprintf("Library: %s", w.storageLoc.Path))
		imgui.Text(fmt.Sprintf("Files: %d", w.fsTree.GetFileCount()))

		totalSize := w.fsTree.GetTotalSize()
		if totalSize > 0 {
			sizeMB := float64(totalSize) / (1024 * 1024)
			imgui.SameLine()
			imgui.Text(fmt.Sprintf("| Total: %.1f MB", sizeMB))
		}

		imgui.Separator()
	}

	currentSearchQuery := searchQuery
	imgui.PushItemWidth(-100)
	if imgui.InputTextWithHint(
		"##search", "search",
		&currentSearchQuery,
		imgui.InputTextFlagsEnterReturnsTrue, nil) {
		if currentSearchQuery != searchQuery {
			cmd := component.UpdateCmd{Type: cmdLibSetSearchQuery, Data: currentSearchQuery}
			w.SendUpdate(cmd)
		}
	}
	imgui.PopItemWidth()
	imgui.SameLine()
	if imgui.Button(font.Icon("Search")) {
		cmd := component.UpdateCmd{Type: cmdLibSetSearchQuery, Data: currentSearchQuery}
		w.SendUpdate(cmd)
	}

	if searchQuery != "" {
		imgui.Text(fmt.Sprintf("Searching for '%s'...", searchQuery))
	}

	imgui.Separator()

	if w.Components.Tree != nil && fsTree != nil {
		w.Components.Tree.Build()
	} else if fsTree == nil && !isScanning {
		imgui.Text("No files found or directory empty")
	}
}

// Destroy cleans up the window and its subscriptions
func (w *LibraryWindow) Destroy() {
	// Unsubscribe from filtered subscriptions (handles all event types)
	if w.filteredEventSub != nil {
		w.filteredEventSub.Unsubscribe()
	}

	// Destroy child components
	if w.Components.Tree != nil {
		w.Components.Tree.Destroy()
	}

	// Call base destroy
	w.Window.Destroy()
}

func NewLibraryWindow() *LibraryWindow {
	w := &LibraryWindow{
		isScanning: false,
	}
	w.Window = window.NewWindow[*LibraryWindow]("Library", "LibraryBig", w.handleUpdate)

	w.Components.Tree = tree.NewTree("library-tree").
		Columns(
			table.NewTableColumn("Name").
				SetFlags(imgui.TableColumnFlagsWidthStretch),
			table.NewTableColumn("Preview").
				SetFlags(imgui.TableColumnFlagsWidthFixed).
				SetInnerWidthOrWeight(120),
			table.NewTableColumn("Duration").
				SetFlags(imgui.TableColumnFlagsWidthFixed).
				SetInnerWidthOrWeight(60),
			table.NewTableColumn("Size").
				SetFlags(imgui.TableColumnFlagsWidthFixed).
				SetInnerWidthOrWeight(80),
		).
		Flags(
			imgui.TableFlagsResizable|
				imgui.TableFlagsScrollY|
				imgui.TableFlagsNoBordersInBody,
		).
		Size(0, -1)

	w.Window.SetLayoutBuilder(w)

	bus := eventbus.Bus
	uuid := w.UUID()

	// Create filtered subscription for library scan events
	w.filteredEventSub = eventbus.NewFilteredSubscription(uuid, 50)
	w.filteredEventSub.SubscribeMultiple(
		bus,
		events.LibraryScanStartedKey,
		events.LibraryScanProgressKey,
		events.LibraryScanCompletedKey,
		events.LibraryScanFailedKey,
	)

	return w
}
