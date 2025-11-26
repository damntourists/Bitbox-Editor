package events

const (
	ComboboxSelectionChangeEventKey = "combobox.selectionchange"
)

// ComboboxSelectionChangeEvent is published when a combobox selection changes.
type ComboboxSelectionChangeEvent struct {
	// UUID is the unique ID of the component that sent the event.
	UUID string
	// Selected is the data of the item that was selected (e.g., a string).
	Selected interface{}
}

// Type implements the events.Event interface.
func (e ComboboxSelectionChangeEvent) Type() string {
	return ComboboxSelectionChangeEventKey
}
