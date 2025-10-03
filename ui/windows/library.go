package windows

import (
	"bitbox-editor/lib/io"
	"bitbox-editor/ui/component"
	"bitbox-editor/ui/fonts"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
)

type LibraryWindow struct {
	*Window

	storageLoc *StorageLocation

	fsTree *io.FSTree

	searchQuery   string
	searchResults []*io.FSEntry

	Components struct {
		Tree *component.TreeComponent
	}

	isScanning   bool
	needsRebuild bool
}

func (w *LibraryWindow) SetStorageLocation(loc *StorageLocation) *LibraryWindow {
	w.storageLoc = loc

	w.isScanning = true

	go func() {
		w.fsTree = io.NewFSTree(loc.path)

		err := w.fsTree.ScanDirectory(loc.path, ".wav", ".WAV")
		if err != nil {
			log.Error(fmt.Sprintf("Failed to scan directory: %s", err.Error()))
			w.isScanning = false
			return
		}

		w.isScanning = false
		w.needsRebuild = true
	}()

	return w
}

func (w *LibraryWindow) buildTreeRows() {
	if w.fsTree == nil || w.fsTree.Root == nil {
		return
	}

	rows := w.buildRowsFromEntry(w.fsTree.Root)
	w.Components.Tree.Rows(rows...)
}

func (w *LibraryWindow) buildRowsFromEntry(entry *io.FSEntry) []*component.TreeRowComponent {
	rows := make([]*component.TreeRowComponent, 0)

	for _, child := range entry.Children {
		var icon string
		if child.IsDir {
			icon = fonts.Icon("Folder")
		} else {
			icon = fonts.Icon("FileAudio")
		}

		displayName := icon + " " + child.Name

		var sizeText string
		if !child.IsDir && child.Size > 0 {
			sizeKB := float64(child.Size) / 1024.0
			if sizeKB < 1024 {
				sizeText = fmt.Sprintf("%.1f KB", sizeKB)
			} else {
				sizeMB := sizeKB / 1024.0
				sizeText = fmt.Sprintf("%.1f MB", sizeMB)
			}
		} else {
			sizeText = ""
		}

		row := component.TreeRow(
			child.Path,
			component.Text(displayName),
			component.Text(sizeText),
		).Flags(imgui.TreeNodeFlagsSpanAvailWidth)

		if child.IsDir {
			childRows := w.buildRowsFromEntry(child)

			if len(childRows) == 0 {
				// Skip empty directories
				continue
			}

			row.Children(childRows...)
		}

		rows = append(rows, row)
	}

	return rows
}

func (w *LibraryWindow) SearchFiles(query string) []*io.FSEntry {
	if w.fsTree == nil || w.fsTree.Root == nil {
		return nil
	}
	return w.fsTree.Search(query)
}

func (w *LibraryWindow) Menu() {
	if imgui.BeginMenuBar() {
		if imgui.Button(fonts.Icon("RefreshCw")) {
			// Rescan the directory
			if w.storageLoc != nil {
				w.SetStorageLocation(w.storageLoc)
			}
		}

		imgui.EndMenuBar()
	}
}

func (w *LibraryWindow) Layout() {
	if w.storageLoc == nil {
		imgui.Text("No storage location selected")
		return
	}

	if w.isScanning {
		imgui.Text("Scanning...")
		return
	}

	if w.needsRebuild && w.fsTree != nil {
		w.buildTreeRows()
		w.needsRebuild = false
	}

	if w.fsTree != nil {
		imgui.Text(fmt.Sprintf("Library: %s", w.storageLoc.path))
		imgui.Text(fmt.Sprintf("Files: %d", w.fsTree.GetFileCount()))

		totalSize := w.fsTree.GetTotalSize()
		if totalSize > 0 {
			sizeMB := float64(totalSize) / (1024 * 1024)
			imgui.SameLine()
			imgui.Text(fmt.Sprintf("| Total: %.1f MB", sizeMB))
		}

		imgui.Separator()
	}

	if imgui.InputTextWithHint(
		"##search",
		"search",
		&w.searchQuery,
		imgui.InputTextFlagsEnterReturnsTrue,
		nil) {
		w.performSearch()
	}
	imgui.SameLine()
	if imgui.Button(fonts.Icon("Search")) {
		w.performSearch()
	}

	if w.searchQuery != "" && len(w.searchResults) > 0 {
		imgui.Text(fmt.Sprintf("Found %d matches", len(w.searchResults)))
	}

	imgui.Separator()

	if w.Components.Tree != nil && w.fsTree != nil {
		w.Components.Tree.Layout()
	} else if w.fsTree == nil {
		imgui.Text("No files found")
	}
}

func (w *LibraryWindow) performSearch() {
	if w.fsTree == nil {
		return
	}

	if w.searchQuery == "" {
		w.searchResults = nil
		w.buildTreeRows()
		return
	}

	w.searchResults = w.fsTree.Search(w.searchQuery)

	w.buildSearchResultRows()
}

func (w *LibraryWindow) buildSearchResultRows() {
	if len(w.searchResults) == 0 {
		w.Components.Tree.Rows()
		return
	}

	matchedPaths := make(map[string]bool)
	for _, entry := range w.searchResults {
		matchedPaths[entry.Path] = true

		current := entry.Parent
		for current != nil {
			matchedPaths[current.Path] = true
			current = current.Parent
		}
	}

	rows := w.buildFilteredRowsFromEntry(w.fsTree.Root, matchedPaths)
	w.Components.Tree.Rows(rows...)
}

func (w *LibraryWindow) buildFilteredRowsFromEntry(
	entry *io.FSEntry,
	matchedPaths map[string]bool) []*component.TreeRowComponent {

	rows := make([]*component.TreeRowComponent, 0)

	for _, child := range entry.Children {
		if !matchedPaths[child.Path] {
			continue
		}

		var icon string
		if child.IsDir {
			icon = fonts.Icon("Folder")
		} else {
			icon = fonts.Icon("FileAudio")
		}

		displayName := icon + " " + child.Name

		var sizeText string
		if !child.IsDir && child.Size > 0 {
			sizeKB := float64(child.Size) / 1024.0
			if sizeKB < 1024 {
				sizeText = fmt.Sprintf("%.1f KB", sizeKB)
			} else {
				sizeMB := sizeKB / 1024.0
				sizeText = fmt.Sprintf("%.1f MB", sizeMB)
			}
		} else {
			// Don't show size for directories
			sizeText = ""
		}

		row := component.TreeRow(
			child.Path,
			component.Text(displayName),
			component.Text(sizeText),
		).Flags(imgui.TreeNodeFlagsSpanAvailWidth)

		if child.IsDir {
			childRows := w.buildFilteredRowsFromEntry(child, matchedPaths)

			if len(childRows) == 0 {
				// Skip empty directories
				continue
			}

			row.Children(childRows...)
		}

		rows = append(rows, row)
	}

	return rows
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

func NewLibraryWindow() *LibraryWindow {
	lw := &LibraryWindow{
		Window:     NewWindow("Library", "LibraryBig"),
		isScanning: false,
	}

	lw.Components.Tree = component.Tree("library-tree").
		Columns(
			component.NewTableColumn("Name").
				Flags(imgui.TableColumnFlagsWidthStretch),
			component.NewTableColumn("Size").
				Flags(imgui.TableColumnFlagsWidthFixed).
				InnerWidthOrWeight(100),
		).
		Flags(
			imgui.TableFlagsBordersV|
				imgui.TableFlagsResizable|
				imgui.TableFlagsRowBg|
				imgui.TableFlagsScrollY|
				imgui.TableFlagsNoBordersInBody,
		).
		Size(0, 0)

	lw.Window.layoutBuilder = lw
	return lw
}
