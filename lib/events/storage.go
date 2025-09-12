package events

type Storage int32

const (
	StorageMountedEvent   Storage = 0
	StorageUnmountedEvent Storage = 1
	StorageActivatedEvent Storage = 2
)

type StorageEventRecord struct {
	Type Storage
	Data interface{}
}
