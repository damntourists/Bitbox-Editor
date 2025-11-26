package events

// Library Cache Event Keys
const (
	LibraryCacheBuiltKey  = "library.cache.built"
	LibraryCacheFailedKey = "library.cache.failed"
)

// LibraryCacheBuiltEvent is published when a library cache is successfully built.
type LibraryCacheBuiltEvent struct {
	// FilePath is the path to the cache file that was built.
	FilePath string
}

// Type implements the events.Event interface
func (e LibraryCacheBuiltEvent) Type() string {
	return LibraryCacheBuiltKey
}

// LibraryCacheFailedEvent is published when a library cache build fails.
type LibraryCacheFailedEvent struct {
	// FilePath is the path to the cache file that failed to build.
	FilePath string
	// Error contains the specific error if the cache build failed.
	Error error
}

// Type implements the events.Event interface
func (e LibraryCacheFailedEvent) Type() string {
	return LibraryCacheFailedKey
}
