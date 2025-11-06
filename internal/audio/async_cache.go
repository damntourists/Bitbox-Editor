package audio

import (
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/wav"
	"go.uber.org/zap"
)

const asyncCacheWorkers = 4

// AsyncWaveCache cache for audio file data. Uses worker pool to limit concurrent file I/O.
type AsyncWaveCache struct {
	entries    sync.Map
	loadQueue  chan loadRequest
	inProgress sync.Map
}

type loadRequest struct {
	path     string
	loadType LoadType
}

type LoadType int

const (
	LoadMetadataOnly LoadType = iota
	LoadMiniDownsamples
	LoadFullSamples
)

var globalAsyncCache *AsyncWaveCache

func init() {
	globalAsyncCache = &AsyncWaveCache{
		loadQueue: make(chan loadRequest, 10000),
	}
	// Start worker pool
	for i := 0; i < asyncCacheWorkers; i++ {
		go globalAsyncCache.worker()
	}
}

func GetGlobalAsyncCache() *AsyncWaveCache {
	return globalAsyncCache
}

// worker processes load requests from queue
func (c *AsyncWaveCache) worker() {
	for req := range c.loadQueue {
		c.processLoad(req.path, req.loadType)
	}
}

// GetSnapshot returns the current snapshot for a path, or nil if not loaded yet
func (c *AsyncWaveCache) GetSnapshot(path string) *WaveFileSnapshot {
	if entry, ok := c.entries.Load(path); ok {
		ptr := entry.(*atomic.Pointer[WaveFileSnapshot])
		return ptr.Load()
	}
	return nil
}

// UpdateSnapshot atomically updates the snapshot for a path
func (c *AsyncWaveCache) UpdateSnapshot(path string, snapshot *WaveFileSnapshot) {
	entry, _ := c.entries.LoadOrStore(path, &atomic.Pointer[WaveFileSnapshot]{})
	ptr := entry.(*atomic.Pointer[WaveFileSnapshot])
	ptr.Store(snapshot)
}

// RequestLoad queues a file for loading if not already loaded/loading
func (c *AsyncWaveCache) RequestLoad(path string, loadType LoadType) {
	// Check if already loaded
	snapshot := c.GetSnapshot(path)
	if snapshot != nil {
		switch loadType {
		case LoadMetadataOnly:
			if snapshot.MetadataLoaded {
				return
			}
		case LoadMiniDownsamples:
			if len(snapshot.MiniDownsamples) > 0 {
				return
			}
		case LoadFullSamples:
			if snapshot.SamplesLoaded {
				return
			}
		}
	}

	// Check if already in progress
	key := fmt.Sprintf("%s:%d", path, loadType)
	if _, loading := c.inProgress.LoadOrStore(key, true); loading {
		//log.Debug("Load already in progress",
		//	zap.String("path", path),
		//	zap.Int("loadType", int(loadType)))
		return
	}

	// Queue for loading
	select {
	case c.loadQueue <- loadRequest{path: path, loadType: loadType}:
	default:
		// Queue full - drop request
		//log.Warn("Load queue full, dropping request",
		//	zap.String("path", path),
		//	zap.Int("loadType", int(loadType)))
		c.inProgress.Delete(key)
	}
}

// processLoad performs the actual loading
func (c *AsyncWaveCache) processLoad(path string, loadType LoadType) {
	defer func() {
		key := fmt.Sprintf("%s:%d", path, loadType)
		c.inProgress.Delete(key)
	}()

	// Get current snapshot or create new one
	snapshot := c.GetSnapshot(path)
	if snapshot == nil {
		snapshot = &WaveFileSnapshot{
			Path: path,
			Name: filepath.Base(path),
		}
	}

	switch loadType {
	case LoadMetadataOnly:
		c.loadMetadataSync(path, snapshot)
	case LoadMiniDownsamples:
		// Load metadata first if needed
		if !snapshot.MetadataLoaded {
			snapshot = c.loadMetadataSync(path, snapshot)
			if snapshot == nil || !snapshot.MetadataLoaded {
				return
			}
		}
		c.loadMiniDownsamplesSync(path, snapshot)
	case LoadFullSamples:
		// Load metadata first if needed
		if !snapshot.MetadataLoaded {
			snapshot = c.loadMetadataSync(path, snapshot)
			if snapshot == nil || !snapshot.MetadataLoaded {
				log.Warn("Metadata load failed, aborting full samples", zap.String("path", path))
				return
			}
		}
		c.loadFullSamplesSync(path, snapshot)
	}
}

func (c *AsyncWaveCache) loadMetadataSync(path string, baseSnapshot *WaveFileSnapshot) *WaveFileSnapshot {
	// Open file
	f, err := os.Open(path)
	if err != nil {
		snapshot := *baseSnapshot
		snapshot.LoadErr = err
		c.UpdateSnapshot(path, &snapshot)
		// Publish fail event
		eventbus.Bus.Publish(events.AudioLoadEventRecord{
			EventType: events.AudioLoadFailedEvent,
			Path:      path,
			Failed:    true,
			Error:     err,
		})
		return nil
	}
	defer f.Close()

	// Decode header
	streamer, format, err := wav.Decode(f)
	if err != nil {
		log.Error("Failed to decode file for metadata", zap.String("path", path), zap.Error(err))
		snapshot := *baseSnapshot
		snapshot.LoadErr = err
		c.UpdateSnapshot(path, &snapshot)
		// Publish fail event
		eventbus.Bus.Publish(events.AudioLoadEventRecord{
			EventType: events.AudioLoadFailedEvent,
			Path:      path,
			Failed:    true,
			Error:     err,
		})
		return nil
	}
	streamer.Close()

	// Create updated snapshot
	newSnapshot := *baseSnapshot
	newSnapshot.Name = filepath.Base(path)
	newSnapshot.SampleRate = int(format.SampleRate)
	newSnapshot.BitDepth = int(format.Precision)
	newSnapshot.NumSamples = streamer.Len()
	newSnapshot.MetadataLoaded = true
	newSnapshot.LoadErr = nil

	// Update cache atomically
	c.UpdateSnapshot(path, &newSnapshot)

	// Publish metadata loaded event
	eventbus.Bus.Publish(events.AudioLoadEventRecord{
		EventType:      events.AudioMetadataLoadedEvent,
		Path:           path,
		MetadataLoaded: true,
		SamplesLoaded:  false,
		Failed:         false,
	})

	return &newSnapshot
}

func (c *AsyncWaveCache) loadMiniDownsamplesSync(path string, baseSnapshot *WaveFileSnapshot) {
	if !baseSnapshot.MetadataLoaded {
		return
	}

	// Open file
	f, err := os.Open(path)
	if err != nil {
		// Don't publish failure here, as this is an auxiliary load.
		// The main metadata/sample load will report the error.
		return
	}
	defer f.Close()

	streamer, format, err := wav.Decode(f)
	if err != nil {
		return
	}
	defer streamer.Close()

	// Sample at intervals for ~1000 points
	targetPoints := 1000
	skipInterval := baseSnapshot.NumSamples / targetPoints
	if skipInterval < 1 {
		skipInterval = 1
	}

	numChannels := format.NumChannels
	miniMins := make([]float32, 0, targetPoints)
	miniMaxs := make([]float32, 0, targetPoints)

	buf := make([][2]float64, skipInterval)
	position := 0

	for position < baseSnapshot.NumSamples {
		n, ok := streamer.Stream(buf)
		if !ok {
			break
		}

		var chunkMin, chunkMax float32 = 1.0, -1.0
		for i := 0; i < n; i++ {
			sample := float32(buf[i][0])
			if numChannels > 1 {
				sample2 := float32(buf[i][1])
				if sample2 < chunkMin {
					chunkMin = sample2
				}
				if sample2 > chunkMax {
					chunkMax = sample2
				}
			}
			if sample < chunkMin {
				chunkMin = sample
			}
			if sample > chunkMax {
				chunkMax = sample
			}
		}

		miniMins = append(miniMins, chunkMin)
		miniMaxs = append(miniMaxs, chunkMax)
		position += n
	}

	// Create an updated snapshot
	newSnapshot := *baseSnapshot
	newSnapshot.MiniDownsamples = []Downsample{{Mins: miniMins, Maxs: miniMaxs}}

	// Update cache atomically
	c.UpdateSnapshot(path, &newSnapshot)
}

func (c *AsyncWaveCache) loadFullSamplesSync(path string, baseSnapshot *WaveFileSnapshot) {
	if !baseSnapshot.MetadataLoaded {
		log.Warn("Cannot load full samples - metadata not loaded", zap.String("path", path))
		return
	}

	// Open file
	f, err := os.Open(path)
	if err != nil {
		log.Error("Failed to open file for full samples",
			zap.String("path", path),
			zap.Error(err))

		eventbus.Bus.Publish(events.AudioLoadEventRecord{
			EventType: events.AudioLoadFailedEvent,
			Path:      path,
			Failed:    true,
			Error:     err,
		})
		return
	}
	defer f.Close()

	streamer, format, err := wav.Decode(f)
	if err != nil {
		log.Error("Failed to decode for full samples",
			zap.String("path", path),
			zap.Error(err))

		eventbus.Bus.Publish(events.AudioLoadEventRecord{
			EventType: events.AudioLoadFailedEvent,
			Path:      path,
			Failed:    true,
			Error:     err,
		})
		return
	}
	defer streamer.Close()

	// Load all samples into memory as mono
	numSamples := baseSnapshot.NumSamples
	samples := make([]float32, numSamples)
	buf := make([][2]float64, 512)
	position := 0
	numChannels := format.NumChannels

	for position < numSamples {
		n, ok := streamer.Stream(buf)
		if !ok {
			break
		}

		for i := 0; i < n && position < numSamples; i++ {
			// Convert to mono by averaging channels
			sample := float32(buf[i][0])
			if numChannels > 1 {
				sample = (sample + float32(buf[i][1])) / 2.0
			}
			samples[position] = sample
			position++
		}
	}

	// Downsample
	targetBins := 10000
	mins, maxs := DownsampleMinMax(samples, targetBins)

	// Calculate min/max Y for display
	var minY, maxY float32 = 1.0, -1.0
	for _, v := range samples {
		if v < minY {
			minY = v
		}
		if v > maxY {
			maxY = v
		}
	}

	// Create updated snapshot
	newSnapshot := *baseSnapshot
	newSnapshot.Downsamples = []Downsample{{Mins: mins, Maxs: maxs}}
	newSnapshot.MinY = minY
	newSnapshot.MaxY = maxY
	newSnapshot.SamplesLoaded = true
	newSnapshot.LoadErr = nil

	// Update cache atomically
	c.UpdateSnapshot(path, &newSnapshot)

	// Emit samples loaded event
	eventbus.Bus.Publish(events.AudioLoadEventRecord{
		EventType:      events.AudioSamplesLoadedEvent,
		Path:           path,
		MetadataLoaded: true,
		SamplesLoaded:  true,
		Failed:         false,
	})
}

// StreamerFromSnapshot creates a streamer directly from snapshot data
func StreamerFromSnapshot(
	snapshot *WaveFileSnapshot,
	loop bool,
	seekPosition,
	startMarker,
	endMarker int,
) (beep.Streamer, error) {
	if !snapshot.MetadataLoaded {
		return nil, fmt.Errorf("metadata not loaded")
	}

	// Open file fresh for streaming
	f, err := os.Open(snapshot.Path)
	if err != nil {
		log.Error("Failed to open file for streaming",
			zap.String("path", snapshot.Path),
			zap.Error(err))
		return nil, err
	}

	streamer, _, err := wav.Decode(f)
	if err != nil {
		log.Error("Failed to decode WAV file",
			zap.String("path", snapshot.Path),
			zap.Error(err))
		f.Close()
		return nil, err
	}

	// Use the provided bounds
	actualStartMarker := startMarker
	if seekPosition >= 0 {
		actualStartMarker = seekPosition
	}

	samplesToPlay := endMarker - actualStartMarker
	if samplesToPlay <= 0 {
		streamer.Close()
		f.Close()
		return nil, fmt.Errorf("no samples to play")
	}

	// Seek if needed
	if actualStartMarker > 0 {
		if err := streamer.Seek(actualStartMarker); err != nil {
			streamer.Close()
			f.Close()
			return nil, err
		}
	}

	var baseStreamer beep.Streamer
	if loop {
		loopingStreamer, err := beep.Loop2(streamer, beep.LoopBetween(startMarker, endMarker))
		if err != nil {
			streamer.Close()
			f.Close()
			return nil, err
		}
		baseStreamer = loopingStreamer
	} else {
		baseStreamer = beep.Take(samplesToPlay, streamer)
	}

	finalStreamer := beep.Seq(baseStreamer, beep.Callback(func() {
		streamer.Close()
		f.Close()
	}))

	return finalStreamer, nil
}
