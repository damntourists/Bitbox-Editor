package audio

import (
	"sync"
	"time"
)

// StateChangeDebouncer debounce rapid state changes (like bounds dragging)
type StateChangeDebouncer struct {
	mu sync.Mutex

	// Delay is how long to wait before committing changes
	delay time.Duration

	timer        *time.Timer
	pendingState *PlaybackState
	hasChanges   bool
	// commitFunc is called when changes are committed
	commitFunc    func(*PlaybackState)
	lastCommitted *PlaybackState
}

// NewStateChangeDebouncer creates a new debouncer
func NewStateChangeDebouncer(delay time.Duration, commitFunc func(*PlaybackState)) *StateChangeDebouncer {
	return &StateChangeDebouncer{
		delay:      delay,
		commitFunc: commitFunc,
	}
}

// Update queues a state change.
func (d *StateChangeDebouncer) Update(newState *PlaybackState) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if newState == nil {
		return
	}

	if d.lastCommitted != nil && d.statesEqual(d.lastCommitted, newState) {
		// No actual change, ignore
		return
	}

	d.pendingState = newState.Copy()
	d.hasChanges = true

	// Reset or create a timer
	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.delay, func() {
		d.commit()
	})
}

// Flush immediately commits any pending changes without waiting for the timer
func (d *StateChangeDebouncer) Flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}

	if d.hasChanges && d.pendingState != nil {
		d.commitLocked()
	}
}

// commit is called by the timer when the delay has elapsed
func (d *StateChangeDebouncer) commit() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.hasChanges && d.pendingState != nil {
		d.commitLocked()
	}
}

// commitLocked performs the actual commit
func (d *StateChangeDebouncer) commitLocked() {
	// Update last committed state
	d.lastCommitted = d.pendingState.Copy()

	// Call the commit function
	if d.commitFunc != nil {
		d.commitFunc(d.pendingState)
	}

	// Clear pending changes
	d.hasChanges = false
	d.timer = nil
}

// statesEqual checks if two states are equal
func (d *StateChangeDebouncer) statesEqual(a, b *PlaybackState) bool {
	if a.BoundsStart != b.BoundsStart || a.BoundsEnd != b.BoundsEnd {
		return false
	}

	if a.RepeatMode != b.RepeatMode {
		return false
	}

	if a.RepeatMode == RepeatModeSlice {
		// For slice mode, check if slices have changed
		if len(a.SlicePositions) != len(b.SlicePositions) {
			return false
		}
		for i := range a.SlicePositions {
			if a.SlicePositions[i] != b.SlicePositions[i] {
				return false
			}
		}
		if a.SliceIdx != b.SliceIdx {
			return false
		}
	}

	return true
}

// Cancel stops any pending commits
func (d *StateChangeDebouncer) Cancel() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}

	d.hasChanges = false
	d.pendingState = nil
}
