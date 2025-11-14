package events

// TODO: Come back to this, it's not ready for implementation yet.

type PresetEditEvent int32

const (
	Updated PresetEditEvent = iota
)

type PresetEditEventRecord struct {
	EventType PresetEditEvent
	Data      interface{}
}
