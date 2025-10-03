package preset

import (
	"bitbox-editor/lib/audio"
	"bitbox-editor/lib/logging"
	"bitbox-editor/lib/parsing/ableton"
	"bitbox-editor/lib/parsing/bitbox"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	log = logging.NewLogger("preset")
}

type Preset struct {
	Name string
	Path string

	bitboxConfig  *bitbox.Document
	abletonConfig *ableton.Ableton

	wavs []*audio.WaveFile

	errored bool

	mu   sync.RWMutex
	once sync.Once
}

func (p *Preset) IsLoading() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, wav := range p.wavs {
		println(wav.IsLoading())
		if !wav.IsLoaded() {
			return true
		}
	}
	return false
}

func (p *Preset) IsErrored() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.errored
}

func (p *Preset) BitboxConfig() *bitbox.Document {
	return p.bitboxConfig
}

func (p *Preset) AbletonConfig() *ableton.Ableton {
	return p.abletonConfig
}

func (p *Preset) resolveWavFiles() {
	if p.bitboxConfig == nil || p.abletonConfig == nil {
		log.Debug("no config to load wavs")
		return
	}

	for _, cell := range p.bitboxConfig.Session.Cells {
		if cell.Filename == "" {
			continue
		}

		path, err := p.ResolveFile(cell.Filename)
		if err != nil {
			continue
		}

		wav := audio.NewWaveFileLazy(filepath.Base(path), path)
		p.wavs = append(p.wavs, wav)
	}
}

func (p *Preset) Wavs() []*audio.WaveFile {
	return p.wavs
}

func (p *Preset) ResolveFile(inputPath string) (string, error) {
	dir, err := filepath.Abs(p.Path)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Base path %s does not exist.", p.Path))
	}

	fixedPath := strings.TrimSpace(inputPath)
	fixedPath = strings.Trim(fixedPath, `"'`)

	fixedPath = strings.ReplaceAll(fixedPath, "\\", "/")
	fixedPath = strings.TrimLeft(fixedPath, "/")

	rel := filepath.FromSlash(fixedPath)

	for {
		candidate := filepath.Join(dir, rel)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

func (p *Preset) loadAbletonConfig(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to open ableton preset %s - %s", path, err))
	}
	defer f.Close()

	bytes, err := os.ReadFile(path)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to read ableton preset %s - %s", path, err))
	}

	var config ableton.Ableton
	err = xml.Unmarshal(bytes, &config)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to parse ableton preset %s - %s", path, err))
	}

	p.abletonConfig = &config
	return nil
}

func (p *Preset) loadBitboxConfig(path string) error {
	f, err := os.Open(path)
	if err != nil {

		return errors.New(fmt.Sprintf("failed to open bitbox preset %s - %s", path, err))
	}
	defer f.Close()

	bytes, err := os.ReadFile(path)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to read bitbox preset %s - %s", path, err))
	}

	var config bitbox.Document
	err = xml.Unmarshal(bytes, &config)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to parse bitbox preset %s - %s", path, err))
	}

	p.bitboxConfig = &config
	return nil
}

func NewPreset(name string, path string) *Preset {
	p := &Preset{
		Name:    name,
		Path:    path,
		errored: false,
	}

	go p.once.Do(func() {
		err := p.loadBitboxConfig(path + "/preset.xml")
		if err != nil {
			p.errored = true
			log.Error("failed to load bitbox preset", zap.Error(err))
		}
		err = p.loadAbletonConfig(path + "/preset.als")
		if err != nil {
			p.errored = true
			log.Error("failed to load ableton preset", zap.Error(err))
		}
		p.resolveWavFiles()
	})

	return p
}
