package audio

import "sync"

// globalWaveCache is a singleton cache shared across the entire application
var globalWaveCache sync.Map

// GetOrCreateWaveFile atomically gets or creates a WaveFile for the given path
func GetOrCreateWaveFile(path string) *WaveFile {
	// Try to load existing
	if cached, ok := globalWaveCache.Load(path); ok {
		return cached.(*WaveFile)
	}

	// Create a new WaveFile
	wav := NewWaveFile(path)

	// Store atomically - if another goroutine beat us, use theirs
	actual, _ := globalWaveCache.LoadOrStore(path, wav)
	return actual.(*WaveFile)
}

// GetCachedWaveFile retrieves a cached WaveFile if it exists
func GetCachedWaveFile(path string) (*WaveFile, bool) {
	if cached, ok := globalWaveCache.Load(path); ok {
		return cached.(*WaveFile), true
	}
	return nil, false
}
