package preset

import (
	"bitbox-editor/lib/audio"
	"bitbox-editor/lib/logging"
	"bitbox-editor/lib/parsing/ableton"
	"bitbox-editor/lib/parsing/bitbox"
	"encoding/xml"
	"os"

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

	loading bool
}

func (p *Preset) BitboxConfig() *bitbox.Document {
	return p.bitboxConfig
}

func (p *Preset) AbletonConfig() *ableton.Ableton {
	return p.abletonConfig
}

func (p *Preset) loadAbletonConfig(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Error("failed to open ableton preset", zap.String("path", path), zap.Error(err))
	}
	defer f.Close()

	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Error("failed to read ableton preset", zap.String("path", path), zap.Error(err))
	}

	var config ableton.Ableton
	err = xml.Unmarshal(bytes, &config)
	if err != nil {
		log.Error("failed to parse ableton preset", zap.String("path", path), zap.Error(err))
	}

	p.abletonConfig = &config
}

func (p *Preset) loadBitboxConfig(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Error("failed to open bitbox preset", zap.String("path", path), zap.Error(err))
	}
	defer f.Close()

	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Error("failed to read bitbox preset", zap.String("path", path), zap.Error(err))
	}

	var config bitbox.Document
	err = xml.Unmarshal(bytes, &config)
	if err != nil {
		log.Error("failed to parse bitbox preset", zap.String("path", path), zap.Error(err))
	}

	p.bitboxConfig = &config
}

func NewPreset(name string, path string) *Preset {
	p := &Preset{
		Name:    name,
		Path:    path,
		loading: true,
	}
	go p.loadBitboxConfig(path + "/preset.xml")
	go p.loadAbletonConfig(path + "/preset.als")
	return p
}
