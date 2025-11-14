package events

type LibraryScanEvent int32

// Library Scan Event Enums
const (
	LibraryScanStartedEvent LibraryScanEvent = iota
	LibraryScanProgressEvent
	LibraryScanCompletedEvent
	LibraryScanFailedEvent
)

// Library Scan Event Keys
const (
	LibraryScanStartedKey   = "library.scan.started"
	LibraryScanProgressKey  = "library.scan.progress"
	LibraryScanCompletedKey = "library.scan.completed"
	LibraryScanFailedKey    = "library.scan.failed"
)

// LibraryScanEventRecord holds data for library scan operations.
// It reports start, progress, completion, or failure.
type LibraryScanEventRecord struct {
	// EventType is the enum value (e.g., LibraryScanStartedEvent).
	EventType LibraryScanEvent
	// Path is the directory path being scanned.
	Path string
	// FileCount is the total number of files found (in a Completed event).
	FileCount int
	// Progress is the normalized scan progress (0.0 to 1.0).
	Progress float64
	// Error contains the specific error if the scan operation failed.
	Error error
}

// Type implements the events.Event interface
func (e LibraryScanEventRecord) Type() string {
	switch e.EventType {
	case LibraryScanStartedEvent:
		return LibraryScanStartedKey
	case LibraryScanProgressEvent:
		return LibraryScanProgressKey
	case LibraryScanCompletedEvent:
		return LibraryScanCompletedKey
	case LibraryScanFailedEvent:
		return LibraryScanFailedKey
	default:
		return "library.scan.unknown"
	}
}
