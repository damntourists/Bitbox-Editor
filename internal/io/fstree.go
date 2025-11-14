package io

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FSTree struct {
	Root    *FSEntry
	pathMap map[string]*FSEntry
}

func NewFSTree(rootPath string) *FSTree {
	root := &FSEntry{
		Name:     filepath.Base(rootPath),
		Path:     rootPath,
		IsDir:    true,
		Children: make([]*FSEntry, 0),
		Parent:   nil,
	}

	return &FSTree{
		Root:    root,
		pathMap: make(map[string]*FSEntry),
	}
}

func (t *FSTree) AddEntry(path string, isDir bool, size int64) *FSEntry {
	path = filepath.Clean(path)

	if existing, exists := t.pathMap[path]; exists {
		return existing
	}

	entry := &FSEntry{
		Name:     filepath.Base(path),
		Path:     path,
		IsDir:    isDir,
		Children: make([]*FSEntry, 0),
		Size:     size,
	}

	t.pathMap[path] = entry

	parentPath := filepath.Dir(path)
	if parentPath != path {
		parent := t.findOrCreateParent(parentPath)
		entry.Parent = parent
		parent.Children = append(parent.Children, entry)

		sort.Slice(parent.Children, func(i, j int) bool {
			if parent.Children[i].IsDir != parent.Children[j].IsDir {
				return parent.Children[i].IsDir
			}
			return strings.ToLower(parent.Children[i].Name) < strings.ToLower(parent.Children[j].Name)
		})
	}

	return entry
}

func (t *FSTree) findOrCreateParent(path string) *FSEntry {
	if existing, exists := t.pathMap[path]; exists {
		return existing
	}

	if path == t.Root.Path {
		return t.Root
	}

	return t.AddEntry(path, true, 0)
}

func (t *FSTree) ScanDirectory(rootPath string, extensions ...string) error {
	if len(extensions) == 0 {
		extensions = []string{".wav"}
	}

	extMap := make(map[string]bool)
	for _, ext := range extensions {
		extMap[strings.ToLower(ext)] = true
	}

	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == rootPath {
			return nil
		}

		// Skip hidden files and directories
		name := info.Name()
		if strings.HasPrefix(name, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			t.AddEntry(path, true, 0)
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if extMap[ext] {
			t.AddEntry(path, false, info.Size())
		}
		return nil
	})
}

func (t *FSTree) FindEntry(path string) (*FSEntry, bool) {
	entry, exists := t.pathMap[filepath.Clean(path)]
	return entry, exists
}

func (t *FSTree) GetAllFiles() []*FSEntry {
	files := make([]*FSEntry, 0)
	t.collectFiles(t.Root, &files)
	return files
}

func (t *FSTree) collectFiles(entry *FSEntry, files *[]*FSEntry) {
	if !entry.IsDir {
		*files = append(*files, entry)
		return
	} else {
		for _, child := range entry.Children {
			t.collectFiles(child, files)
		}
	}
}

func (t *FSTree) GetFileCount() int {
	return len(t.GetAllFiles())
}

func (t *FSTree) GetTotalSize() int64 {
	var total int64
	for _, entry := range t.GetAllFiles() {
		total += entry.Size
	}
	return total
}

func (t *FSTree) Filter(predicate func(*FSEntry) bool) []*FSEntry {
	results := make([]*FSEntry, 0)
	for _, child := range t.Root.Children {
		t.filterRecursive(child, predicate, &results)
	}
	return results
}

func (t *FSTree) filterRecursive(entry *FSEntry, predicate func(*FSEntry) bool, results *[]*FSEntry) {
	if predicate(entry) {
		*results = append(*results, entry)
	}
	for _, child := range entry.Children {
		t.filterRecursive(child, predicate, results)
	}
}

func (t *FSTree) Search(query string) []*FSEntry {
	query = strings.ToLower(query)
	return t.Filter(func(entry *FSEntry) bool {
		return strings.Contains(strings.ToLower(entry.Name), query)
	})
}

func (t *FSTree) SearchPath(query string) []*FSEntry {
	query = strings.ToLower(query)
	return t.Filter(func(entry *FSEntry) bool {
		return !entry.IsDir && strings.Contains(strings.ToLower(entry.Path), query)
	})
}

func (t *FSTree) GetByExtension(ext string) []*FSEntry {
	ext = strings.ToLower(ext)
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return t.Filter(func(e *FSEntry) bool {
		return !e.IsDir && strings.ToLower(filepath.Ext(e.Path)) == ext
	})
}
func (t *FSTree) GetFilesByPattern(pattern string) []*FSEntry {
	return t.Filter(func(e *FSEntry) bool {
		if e.IsDir {
			return false
		}
		matched, err := filepath.Match(pattern, e.Name)
		return err == nil && matched
	})
}

func (t *FSTree) GetDirectories() []*FSEntry {
	return t.Filter(func(e *FSEntry) bool {
		return e.IsDir
	})
}

func (t *FSTree) GetFilesOnly() []*FSEntry {
	return t.Filter(func(e *FSEntry) bool {
		return !e.IsDir
	})
}

func (t *FSTree) GetEntriesByDepth(depth int) []*FSEntry {
	return t.Filter(func(e *FSEntry) bool {
		return e.GetDepth() == depth
	})
}

func (t *FSTree) GetLargestFiles(n int) []*FSEntry {
	files := t.GetFilesOnly()

	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})

	if n > len(files) {
		n = len(files)
	}
	return files[:n]
}
