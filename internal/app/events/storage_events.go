package events

// Storage Event Keys
const (
	StorageActivatedEventKey  = "storage.activated"
	StorageDeselectedEventKey = "storage.deselected"
	StorageMountedEventKey    = "storage.mounted"
	StorageUnmountedEventKey  = "storage.unmounted"
)

// StorageActivatedEvent is published when a storage location is selected/activated.
type StorageActivatedEvent struct {
	// Location is the storage location that was activated.
	// Type: *storage.StorageLocation
	Location interface{}
}

func (e StorageActivatedEvent) Type() string {
	return StorageActivatedEventKey
}

// StorageDeselectedEvent is published when a storage location is deselected.
type StorageDeselectedEvent struct {
	// No additional data needed
}

func (e StorageDeselectedEvent) Type() string {
	return StorageDeselectedEventKey
}

// StorageMountedEvent is published when a storage device is mounted.
type StorageMountedEvent struct {
	// Location is the storage location that was mounted.
	// Type: *storage.StorageLocation
	Location interface{}
}

func (e StorageMountedEvent) Type() string {
	return StorageMountedEventKey
}

// StorageUnmountedEvent is published when a storage device is unmounted.
type StorageUnmountedEvent struct {
	// Location is the storage location that was unmounted.
	// Type: *storage.StorageLocation
	Location interface{}
}

func (e StorageUnmountedEvent) Type() string {
	return StorageUnmountedEventKey
}
