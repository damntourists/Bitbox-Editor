package events

type StorageEventType int32

const (
	StorageActivatedEvent StorageEventType = iota
	StorageDeselectedEvent
	StorageMountedEvent
	StorageUnmountedEvent
)

const (
	StorageActivatedEventKey  = "storage.activated"
	StorageDeselectedEventKey = "storage.deselected"
	StorageMountedEventKey    = "storage.mounted"
	StorageUnmountedEventKey  = "storage.unmounted"
)

// StorageEventRecord is published when a storage location is selected, deselected, mounted, or unmounted.
type StorageEventRecord struct {
	EventType StorageEventType
	// Data will be the *storage.StorageLocation struct.
	Data interface{}
}

// Type implements the events.Event interface, returning the string routing key.
func (e StorageEventRecord) Type() string {
	switch e.EventType {
	case StorageActivatedEvent:
		return StorageActivatedEventKey
	case StorageDeselectedEvent:
		return StorageDeselectedEventKey
	case StorageMountedEvent:
		return StorageMountedEventKey
	case StorageUnmountedEvent:
		return StorageUnmountedEventKey
	default:
		return "storage.unknown"
	}
}
