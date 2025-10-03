package io

import "path/filepath"

type FSEntry struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*FSEntry
	Parent   *FSEntry
	Size     int64
}

func (e *FSEntry) GetRelativePath(root string) string {
	relPath, err := filepath.Rel(root, e.Path)
	if err != nil {
		return e.Path
	}
	return relPath
}

func (e *FSEntry) GetDepth() int {
	depth := 0
	current := e.Parent
	for current != nil {
		depth++
		current = current.Parent
	}
	return depth
}

func (e *FSEntry) HasChildren() bool {
	return e.IsDir && len(e.Children) > 0
}

func (e *FSEntry) GetFileChildren() []*FSEntry {
	files := make([]*FSEntry, 0)
	for _, child := range e.Children {
		if !child.IsDir {
			files = append(files, child)
		}
	}
	return files
}

func (e *FSEntry) GetDirChildren() []*FSEntry {
	dirs := make([]*FSEntry, 0)
	for _, child := range e.Children {
		if child.IsDir {
			dirs = append(dirs, child)
		}
	}
	return dirs
}
