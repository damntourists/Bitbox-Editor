package events

// Preset Event Keys
const (
	PresetLoadEventKey = "preset.load"
)

// PresetLoadEvent is published when a preset should be loaded.
type PresetLoadEvent struct {
	// Preset is the preset to load.
	// Type: *preset.Preset
	Preset interface{}
}

func (e PresetLoadEvent) Type() string {
	return PresetLoadEventKey
}
