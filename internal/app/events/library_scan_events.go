package events

// Library Scan Event Keys
const (
	LibraryScanStartedKey   = "library.scan.started"
	LibraryScanProgressKey  = "library.scan.progress"
	LibraryScanCompletedKey = "library.scan.completed"
	LibraryScanFailedKey    = "library.scan.failed"
)

// LibraryScanStartedEvent is published when a library scan begins.
type LibraryScanStartedEvent struct {
	Path string
}

func (e LibraryScanStartedEvent) Type() string {
	return LibraryScanStartedKey
}

// LibraryScanProgressEvent is published during library scanning to report progress.
type LibraryScanProgressEvent struct {
	Path     string
	Progress float64 // 0.0 to 1.0
}

func (e LibraryScanProgressEvent) Type() string {
	return LibraryScanProgressKey
}

// LibraryScanCompletedEvent is published when a library scan completes successfully.
type LibraryScanCompletedEvent struct {
	Path      string
	FileCount int
}

func (e LibraryScanCompletedEvent) Type() string {
	return LibraryScanCompletedKey
}

// LibraryScanFailedEvent is published when a library scan fails.
type LibraryScanFailedEvent struct {
	Path  string
	Error error
}

func (e LibraryScanFailedEvent) Type() string {
	return LibraryScanFailedKey
}
