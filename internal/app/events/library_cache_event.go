package events

type LibraryCacheEvent int32

// Library Cache Events
const (
	LibraryCacheBuiltEvent LibraryCacheEvent = iota
	LibraryCacheFailedEvent
)

// Library Cache Event Keys
const (
	LibraryCacheBuiltKey  = "library.cache.built"
	LibraryCacheFailedKey = "library.cache.failed"
)

// LibraryCacheEventRecord holds data for library cache build events.
type LibraryCacheEventRecord struct {
	// EventType is the enum value (e.g., LibraryCacheBuiltEvent).
	EventType LibraryCacheEvent
	// FilePath is the path to the cache file that was built or failed to build.
	FilePath string
	// Error contains the specific error if the cache build failed.
	Error error
}

// Type implements the events.Event interface
func (e LibraryCacheEventRecord) Type() string {
	switch e.EventType {
	case LibraryCacheBuiltEvent:
		return LibraryCacheBuiltKey
	case LibraryCacheFailedEvent:
		return LibraryCacheFailedKey
	default:
		return "library.cache.unknown"
	}
}
