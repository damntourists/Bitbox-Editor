package events

type Preset int32

const (
	LoadPreset Preset = 0
)

type PresetEventRecord struct {
	Type Preset
	Data interface{}
}
