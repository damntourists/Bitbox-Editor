package bitbox

import (
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Document struct {
	XMLName xml.Name `xml:"document"`
	Session *Session `xml:"session"`
}

func (d *Document) ResolveWavFiles(baseDir string) ([]WavFile, error) {
	var out []WavFile

	if baseDir != "" {
		if abs, err := filepath.Abs(baseDir); err == nil {
			baseDir = abs
		}
	}

	resolve := func(orig string, row, col, layer *int, typ string) WavFile {
		w := WavFile{
			Original: orig,
			Row:      row,
			Column:   col,
			Layer:    layer,
			Type:     typ,
		}
		if orig == "" {
			return w
		}

		normalized := normalizePath(orig)

		if p, ok := tryExists(normalized); ok {
			w.Resolved = p
			return w
		}

		if baseDir != "" {
			trimmed := strings.TrimLeft(normalized, string(filepath.Separator))
			j := filepath.Join(baseDir, trimmed)
			if p, ok := tryExists(j); ok {
				w.Resolved = p
				return w
			}
			j2 := filepath.Join(baseDir, normalized)
			if p, ok := tryExists(j2); ok {
				w.Resolved = p
				return w
			}
		}

		if baseDir != "" {
			base := filepath.Base(normalized)
			if found := findFileByName(baseDir, base, 64_000); found != "" {
				w.Resolved = found
				return w
			}
		}

		return w
	}

	for i := range d.Session.Cells {
		c := &d.Session.Cells[i]
		if c.Filename == "" {
			continue
		}
		if strings.EqualFold(filepath.Ext(c.Filename), ".wav") {
			out = append(out, resolve(c.Filename, c.Row, c.Column, c.Layer, c.Type))
		}
	}

	return out, nil
}

func normalizePath(p string) string {
	p = strings.ReplaceAll(p, "\\", string(filepath.Separator))
	return filepath.Clean(p)
}

func tryExists(p string) (string, bool) {
	if p == "" {
		return "", false
	}
	if !filepath.IsAbs(p) {
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
	}
	if st, err := os.Stat(p); err == nil && !st.IsDir() {
		return p, true
	}
	return "", false
}

func findFileByName(root, basename string, maxFiles int) string {
	if basename == "" {
		return ""
	}
	count := 0
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		count++
		if count > maxFiles {
			return io.EOF
		}
		if strings.EqualFold(d.Name(), basename) {
			if st, err := os.Stat(path); err == nil && !st.IsDir() {
				abs := path
				if a, err := filepath.Abs(path); err == nil {
					abs = a
				}
				return &foundErr{path: abs}
			}
		}
		return nil
	})
	return ""
}

type foundErr struct{ path string }

func (e *foundErr) Error() string { return e.path }
