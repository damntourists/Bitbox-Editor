package events

type ComboboxEvent int32

const (
	ComboboxSelectionChangeEvent ComboboxEvent = iota
)
const (
	ComboboxSelectionChangeEventKey = "combobox.selectionchange"
)

// ComboboxEventRecord is published when a selection changes.
type ComboboxEventRecord struct {
	EventType ComboboxEvent
	// UUID is the unique ID of the component that sent the event.
	UUID string
	// Selected is the data of the item that was selected (e.g., a string).
	Selected interface{}
}

// Type implements the events.Event interface.
func (e ComboboxEventRecord) Type() string {
	switch e.EventType {
	case ComboboxSelectionChangeEvent:
		return ComboboxSelectionChangeEventKey
	default:
		return "combobox.unknown"
	}
}
