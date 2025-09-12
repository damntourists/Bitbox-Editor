package io

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func IsFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

func FindByExt(root string, exts ...string) ([]string, error) {
	normalized := make(map[string]struct{}, len(exts))
	for _, e := range exts {
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		normalized[strings.ToLower(e)] = struct{}{}
	}

	matches := []string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			name := d.Name()
			if name == "Presets" ||
				strings.HasPrefix(name, ".") {
				return fs.SkipDir
			}
			return nil
		}

		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if _, ok := normalized[ext]; ok {
			matches = append(matches, path)
		}
		return nil
	})
	return matches, err
}
