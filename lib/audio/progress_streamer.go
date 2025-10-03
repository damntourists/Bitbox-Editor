package audio

import (
	"sync"

	"github.com/gopxl/beep/v2"
)

type progressStreamer struct {
	beep.StreamSeeker
	totalSamples int
	position     int
	progress     float64
	mu           sync.RWMutex
}

func (ps *progressStreamer) Stream(samples [][2]float64) (int, bool) {
	n, ok := ps.StreamSeeker.Stream(samples)

	ps.mu.Lock()
	ps.position += n
	if ps.totalSamples > 0 {
		ps.progress = float64(ps.position) / float64(ps.totalSamples)
		if ps.progress > 1 {
			ps.progress = 1
		}
	}
	ps.mu.Unlock()

	return n, ok
}

func (ps *progressStreamer) Seek(pos int) error {
	err := ps.StreamSeeker.Seek(pos)
	if err == nil {
		ps.mu.Lock()
		ps.position = pos
		if ps.totalSamples > 0 {
			ps.progress = float64(ps.position) / float64(ps.totalSamples)
			if ps.progress > 1 {
				ps.progress = 1
			}
		}
		ps.mu.Unlock()
	}
	return err
}

func (ps *progressStreamer) Progress() float64 {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.progress
}

func (ps *progressStreamer) TotalSamples() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.totalSamples
}

func (ps *progressStreamer) Position() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.position
}

func (ps *progressStreamer) Reset() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.position = 0
	ps.progress = 0
}
