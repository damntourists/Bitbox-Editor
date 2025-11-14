package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/wav"
)

// WaveFile represents a WAV file with metadata and sample data
type WaveFile struct {
	// Path is the path to the file
	Path string

	// Name is the name of the file, without path or extension
	Name string

	// SampleRate is the sample rate of the file
	SampleRate int

	// BitDepth is the bit depth of the file
	BitDepth int

	// NumSamples is the number of samples in the file
	NumSamples int

	// Channels is the number of channels in the file
	Channels [][]float32

	// Downsamples contains downsampled versions of the file for larger waveform component
	Downsamples []Downsample

	// MiniDownsamples contains downsampled versions of the file for smaller waveform component
	MiniDownsamples []Downsample

	// MinY and MaxY are the minimum and maximum amplitude values in the file
	MinY, MaxY float32

	// StartMarker and EndMarker are position markers for playback or processing
	StartMarker, EndMarker int

	// metadataLoaded indicates if header info metadata is loaded
	metadataLoaded bool

	// samplesLoaded indicates if full samples & downsamples are loaded
	samplesLoaded bool

	// loadErr is the error encountered during loading
	loadErr error

	// metadataErr is the error from metadata loading
	metadataErr error

	// samplesErr is the error from samples loading
	samplesErr error

	// miniErr is the error from mini downsamples loading
	miniErr error

	// progressStreamer is a streamer that wraps the original streamer and adds progress tracking
	progressStreamer *ProgressStreamer

	// format is the audio format information
	format beep.Format

	// mutex protects access to the WaveFile during concurrent operations
	mutex sync.RWMutex

	// loadMetadataOnce ensures metadata loading happens exactly once
	loadMetadataOnce sync.Once

	// loadSamplesOnce ensures sample loading happens exactly once
	loadSamplesOnce sync.Once

	// loadMiniDownsamplesOnce ensures mini downsamples loading happens exactly once
	loadMiniDownsamplesOnce sync.Once
}

// NewWaveFile creates a new WaveFile. Use Load() to load the file
func NewWaveFile(path string) *WaveFile {
	w := &WaveFile{
		Name:           filepath.Base(path),
		Path:           path,
		StartMarker:    0,
		EndMarker:      0,
		metadataLoaded: false,
		samplesLoaded:  false,
	}
	return w
}

func (w *WaveFile) IsMetadataLoaded() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.metadataLoaded
}

// Streamer creates a new audio streamer using StartMarker as the beginning
func (w *WaveFile) Streamer(loop bool) (beep.Streamer, error) {
	return w.StreamerWithSeek(loop, -1)
}

// StreamerWithSeek creates a new audio streamer for this wave file
func (w *WaveFile) StreamerWithSeek(loop bool, seekPosition int) (beep.Streamer, error) {
	cache := GetGlobalAsyncCache()
	snapshotPtr := cache.GetSnapshot(w.Path)

	if snapshotPtr == nil {
		// No cached snapshot available - file is still loading or not yet requested
		return nil, fmt.Errorf("cannot create streamer: file data not yet available (still loading)")
	}

	snapshot := *snapshotPtr
	metadataLoaded := snapshot.MetadataLoaded
	loadErr := snapshot.LoadErr

	startMarker := 0
	endMarker := snapshot.NumSamples

	if seekPosition >= 0 {
		startMarker = seekPosition
	}

	if !metadataLoaded {
		if loadErr != nil {
			return nil, fmt.Errorf("cannot create streamer: metadata load previously failed: %w", loadErr)
		}
		return nil, fmt.Errorf("cannot create streamer: metadata not loaded (caller must load metadata first)")
	}

	// Opens the file, returns an error if it fails
	f, err := os.Open(w.Path)
	if err != nil {
		return nil, err
	}

	// Decode the file, returns an error if it fails
	streamer, _, err := wav.Decode(f)
	if err != nil {
		f.Close()
		return nil, err
	}

	// Check start and end markers to see if we have samples to play, return error if not
	samplesToPlay := endMarker - startMarker
	if samplesToPlay <= 0 {
		streamer.Close()
		f.Close()
		return nil, fmt.Errorf("no samples to play (start: %d, end: %d)", startMarker, endMarker)
	}

	// Seek to start marker
	if startMarker > 0 {
		if err := streamer.Seek(startMarker); err != nil {
			streamer.Close()
			f.Close()
			return nil, fmt.Errorf("failed to seek to position %d: %w", startMarker, err)
		}
	}

	// Wrap in looping streamer if requested
	var baseStreamer beep.Streamer
	if loop {
		loopingStreamer, err := beep.Loop2(
			streamer,
			beep.LoopBetween(startMarker, endMarker),
		)
		if err != nil {
			streamer.Close()
			f.Close()
			return nil, err
		}
		baseStreamer = loopingStreamer
	} else {
		baseStreamer = beep.Take(samplesToPlay, streamer)
	}

	// Wrap in progress streamer
	progressStreamer := NewProgressStreamer(baseStreamer, startMarker, samplesToPlay)

	func() {
		w.mutex.Lock()
		defer w.mutex.Unlock()
		w.progressStreamer = progressStreamer
	}()

	finalStreamer := beep.Seq(progressStreamer, beep.Callback(func() {
		// Close the original decoder stream
		streamer.Close()

		// Close the file
		f.Close()

		// Clear the progress streamer when playback finishes
		func() {
			w.mutex.Lock()
			defer w.mutex.Unlock()
			w.progressStreamer = nil
		}()
	}))

	return finalStreamer, nil
}

// Progress returns the current playback progress
func (w *WaveFile) Progress() float64 {
	ps := w.progressStreamer
	if ps == nil {
		return 0.0
	}
	return ps.Progress()
}

// Position returns the current position in samples (relative to where playback started)
func (w *WaveFile) Position() int {
	ps := w.progressStreamer
	if ps == nil {
		return 0
	}
	return ps.Position()
}

// AbsolutePosition returns the absolute position in the original audio file
func (w *WaveFile) AbsolutePosition() int {
	ps := w.progressStreamer
	if ps == nil {
		return 0
	}
	return ps.AbsolutePosition()
}

// PositionSeconds returns the current position in seconds
func (w *WaveFile) PositionSeconds() float64 {
	ps := w.progressStreamer
	sr := w.SampleRate

	if ps == nil || sr == 0 {
		return 0
	}
	return float64(ps.Position()) / float64(sr)
}

// GetMiniDownsamples returns the downsampled data for the first channel
func (w *WaveFile) GetMiniDownsamples() []Downsample {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if w.MiniDownsamples == nil {
		return nil
	}
	dsCopy := make([]Downsample, len(w.MiniDownsamples))
	copy(dsCopy, w.MiniDownsamples)
	return dsCopy
}

// IsSamplesLoaded returns true if the full samples are loaded
func (w *WaveFile) IsSamplesLoaded() bool {
	w.mutex.RLock()
	loaded := w.samplesLoaded
	w.mutex.RUnlock()
	return loaded
}

// HasMiniDownsamples returns true if the mini downsamples are loaded
func (w *WaveFile) HasMiniDownsamples() bool {
	w.mutex.RLock()
	has := len(w.MiniDownsamples) > 0 && len(w.MiniDownsamples[0].Mins) > 0
	w.mutex.RUnlock()
	return has
}

// LoadError returns the error that caused the metadata to fail to load, if any
func (w *WaveFile) LoadError() error {
	w.mutex.RLock()
	err := w.loadErr
	w.mutex.RUnlock()
	return err
}

// LoadMetadata loads the metadata from the file.
func (w *WaveFile) LoadMetadata() error {
	w.loadMetadataOnce.Do(func() {
		w.loadMetadataImpl()
	})

	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.metadataErr
}

// loadMetadataImpl loads the metadata from the file.
func (w *WaveFile) loadMetadataImpl() {
	// Check if already loaded
	w.mutex.RLock()
	if w.metadataLoaded {
		w.mutex.RUnlock()
		return
	}
	path := w.Path
	w.mutex.RUnlock()

	f, err := os.Open(path)
	if err != nil {
		w.mutex.Lock()
		w.metadataErr = fmt.Errorf("metadata open failed: %w", err)
		w.loadErr = w.metadataErr
		w.mutex.Unlock()
		return
	}

	// Decode reads the header to determine format and length
	streamer, format, err := wav.Decode(f)
	if err != nil {
		f.Close()
		w.mutex.Lock()
		w.metadataErr = fmt.Errorf("metadata decode failed: %w", err)
		w.loadErr = w.metadataErr
		w.mutex.Unlock()
		return
	}

	// Close the streamer, we don't need it yet
	streamer.Close()

	f.Close()

	// Calculate metadata
	name := filepath.Base(path)
	sampleRate := int(format.SampleRate)
	bitDepth := int(format.Precision)
	numSamples := streamer.Len()

	w.mutex.Lock()
	w.format = format
	w.Name = name
	w.SampleRate = sampleRate
	w.BitDepth = bitDepth
	w.NumSamples = numSamples

	// Set default markers if not already set
	if w.EndMarker <= 0 || w.EndMarker > w.NumSamples {
		w.EndMarker = w.NumSamples
	}
	if w.StartMarker < 0 || w.StartMarker >= w.NumSamples {
		w.StartMarker = 0
	}

	// Clear error on success
	w.metadataLoaded = true
	w.metadataErr = nil
	w.mutex.Unlock()

	// Store snapshot after successful metadata load
	cache := GetGlobalAsyncCache()
	snapshot := w.Snapshot()
	cache.UpdateSnapshot(w.Path, &snapshot)
}

// GetDownsamples returns the downsampled data for the first channel
func (w *WaveFile) GetDownsamples() []Downsample {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	if !w.samplesLoaded || w.Downsamples == nil {
		return nil
	}
	dsCopy := make([]Downsample, len(w.Downsamples))
	copy(dsCopy, w.Downsamples)
	return dsCopy
}

// GetStartMarker returns the start marker position
func (w *WaveFile) GetStartMarker() int {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.StartMarker
}

// GetEndMarker returns the end marker position
func (w *WaveFile) GetEndMarker() int {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.EndMarker
}

// GetNumSamples returns the number of samples in the file
func (w *WaveFile) GetNumSamples() int {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.NumSamples
}

// GetFormat returns the format of the file
func (w *WaveFile) GetFormat() beep.Format {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.format
}

// LoadSamplesAndDownsample loads the full samples and downsamples them
func (w *WaveFile) LoadSamplesAndDownsample() error {
	w.loadSamplesOnce.Do(func() {
		w.loadSamplesAndDownsampleImpl()
	})

	w.mutex.RLock()
	defer w.mutex.RUnlock()
	if !w.samplesLoaded {
		if w.samplesErr != nil {
			return w.samplesErr
		}
		return fmt.Errorf("cannot load samples: metadata not loaded")
	}
	return w.samplesErr
}

// loadSamplesAndDownsampleImpl loads the full samples and downsamples them
func (w *WaveFile) loadSamplesAndDownsampleImpl() {
	// Check if already loaded
	w.mutex.RLock()
	if w.samplesLoaded {
		w.mutex.RUnlock()
		return
	}

	if !w.metadataLoaded || w.loadErr != nil {
		w.mutex.RUnlock()
		return
	}

	path := w.Path
	numSamples := w.NumSamples

	w.mutex.RUnlock()

	f, err := os.Open(path)
	if err != nil {
		w.mutex.Lock()
		w.samplesErr = fmt.Errorf("samples open failed: %w", err)
		w.loadErr = w.samplesErr
		w.mutex.Unlock()
		return
	}
	defer f.Close()

	streamer, format, err := wav.Decode(f)
	if err != nil {
		w.mutex.Lock()
		w.samplesErr = fmt.Errorf("samples decode failed: %w", err)
		w.loadErr = w.samplesErr
		w.mutex.Unlock()
		return
	}
	defer streamer.Close()

	// Read all samples
	numChannels := format.NumChannels
	if numChannels <= 0 {
		numChannels = 1
	}

	channels := make([][]float32, numChannels)

	for c := 0; c < numChannels; c++ {
		channels[c] = make([]float32, 0, numSamples)
	}

	// Read samples in chunks of 4096 samples
	const bufSize = 4096
	buf := make([][2]float64, bufSize)
	minY, maxY := float32(1.0), float32(-1.0)
	samplesRead := 0

	for {
		n, ok := streamer.Stream(buf)
		if !ok {
			break
		}
		if n == 0 {
			continue
		}

		currentBuf := buf[:n]
		samplesRead += n

		if numChannels == 1 {
			ch0 := channels[0]
			for i := range currentBuf {
				sample := float32(currentBuf[i][0])
				ch0 = append(ch0, sample)
				if sample < minY {
					minY = sample
				}
				if sample > maxY {
					maxY = sample
				}
			}
			channels[0] = ch0
		} else {
			ch0 := channels[0]
			ch1 := channels[1]
			for i := range currentBuf {
				left := float32(currentBuf[i][0])
				right := float32(currentBuf[i][1])
				ch0 = append(ch0, left)
				ch1 = append(ch1, right)
				if left < minY {
					minY = left
				}
				if left > maxY {
					maxY = left
				}
				if right < minY {
					minY = right
				}
				if right > maxY {
					maxY = right
				}
			}
			channels[0] = ch0
			channels[1] = ch1
		}
	}

	var loadErr error
	if samplesRead != numSamples {
		loadErr = fmt.Errorf("samples read (%d) != expected length (%d)", samplesRead, numSamples)
		numSamples = samplesRead
		if len(channels) > 0 {
			numSamples = len(channels[0])
		}
	} else if streamErr := streamer.Err(); streamErr != nil {
		// Check for any stream errors
		loadErr = streamErr
	}

	var downsamples []Downsample
	var miniDownsamples []Downsample

	if samplesRead > 0 {
		// Only downsample if we actually read something
		downsamples = make([]Downsample, len(channels))
		for c := 0; c < len(channels); c++ {
			ys := channels[c]
			mins, maxs := DownsampleMinMax(ys, 512)
			downsamples[c] = Downsample{Mins: mins, Maxs: maxs}
		}

		miniDownsamples = make([]Downsample, len(channels))
		for c := 0; c < len(channels); c++ {
			ys := channels[c]
			mins, maxs := DownsampleMinMax(ys, 120)
			miniDownsamples[c] = Downsample{Mins: mins, Maxs: maxs}
		}
	} else {
		if loadErr == nil {
			// No other error, mark as error due to empty
			loadErr = fmt.Errorf("no valid samples read from file")
		}
	}

	w.mutex.Lock()
	w.Channels = channels
	w.Downsamples = downsamples
	w.MiniDownsamples = miniDownsamples
	w.MinY = minY
	w.MaxY = maxY
	w.NumSamples = numSamples
	w.samplesErr = loadErr
	w.loadErr = loadErr
	w.samplesLoaded = true
	w.mutex.Unlock()

	// Store snapshot after successful load
	cache := GetGlobalAsyncCache()
	snapshot := w.Snapshot()
	cache.UpdateSnapshot(w.Path, &snapshot)
}

// LoadMiniDownsamplesOnly loads only the minimal downsampled data needed for UI preview
func (w *WaveFile) LoadMiniDownsamplesOnly() error {
	w.loadMiniDownsamplesOnce.Do(func() {
		w.loadMiniDownsamplesOnlyImpl()
	})

	w.mutex.RLock()
	defer w.mutex.RUnlock()
	if w.samplesLoaded {
		// Full samples loaded is even better
		return nil
	}
	if len(w.MiniDownsamples) > 0 {
		// Mini downsamples loaded successfully
		return nil
	}
	if w.miniErr != nil {
		// Mini downsamples failed to load
		return w.miniErr
	}

	return fmt.Errorf("failed to load mini downsamples")
}

func (w *WaveFile) loadMiniDownsamplesOnlyImpl() {
	w.mutex.RLock()
	if !w.metadataLoaded {
		w.mutex.RUnlock()
		return
	}
	if w.samplesLoaded {
		// Already have full samples, nothing to do
		w.mutex.RUnlock()
		return
	}

	path := w.Path
	numSamples := w.NumSamples
	format := w.format
	w.mutex.RUnlock()

	if numSamples == 0 {
		w.mutex.Lock()
		w.miniErr = fmt.Errorf("no samples to load")
		w.mutex.Unlock()
		return
	}

	f, err := os.Open(path)
	if err != nil {
		w.mutex.Lock()
		w.miniErr = fmt.Errorf("mini-downsample open failed: %w", err)
		w.mutex.Unlock()
		return
	}

	streamer, _, err := wav.Decode(f)
	if err != nil {
		f.Close()
		w.mutex.Lock()
		w.miniErr = fmt.Errorf("mini-downsample decode failed: %w", err)
		w.mutex.Unlock()
		return
	}
	defer streamer.Close()
	defer f.Close()

	// Calculate how many samples to skip for ~1000 downsamples
	targetPoints := 1000
	skipInterval := numSamples / targetPoints
	if skipInterval < 1 {
		skipInterval = 1
	}

	// Sample the audio at intervals
	numChannels := format.NumChannels
	miniMins := make([]float32, 0, targetPoints)
	miniMaxs := make([]float32, 0, targetPoints)

	buf := make([][2]float64, skipInterval)
	position := 0

	for position < numSamples {
		n, ok := streamer.Stream(buf)
		if !ok {
			break
		}

		// Find min/max in this chunk
		var chunkMin, chunkMax float32 = 1.0, -1.0
		for i := 0; i < n; i++ {
			// Use left channel or mono
			sample := float32(buf[i][0])
			if numChannels > 1 {
				// Also check right channel
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

	miniDS := []Downsample{{Mins: miniMins, Maxs: miniMaxs}}

	w.mutex.Lock()
	w.MiniDownsamples = miniDS
	w.miniErr = nil
	w.mutex.Unlock()

	// Store snapshot after successful load
	cache := GetGlobalAsyncCache()
	snapshot := w.Snapshot()
	cache.UpdateSnapshot(w.Path, &snapshot)
}

// Duration returns the duration of the audio file in seconds
func (w *WaveFile) Duration() float64 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	if !w.metadataLoaded || w.SampleRate == 0 {
		return 0
	}
	return float64(w.NumSamples) / float64(w.SampleRate)
}

// Snapshot creates a view model copy of the WaveFile's current state
func (w *WaveFile) Snapshot() WaveFileSnapshot {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	snapshot := WaveFileSnapshot{
		Path:           w.Path,
		Name:           w.Name,
		SampleRate:     w.SampleRate,
		BitDepth:       w.BitDepth,
		NumSamples:     w.NumSamples,
		MinY:           w.MinY,
		MaxY:           w.MaxY,
		MetadataLoaded: w.metadataLoaded,
		SamplesLoaded:  w.samplesLoaded,
		LoadErr:        w.loadErr,
	}

	// Copy downsamples if they exist
	if len(w.Downsamples) > 0 {
		snapshot.Downsamples = make([]Downsample, len(w.Downsamples))
		copy(snapshot.Downsamples, w.Downsamples)
	}
	if len(w.MiniDownsamples) > 0 {
		snapshot.MiniDownsamples = make([]Downsample, len(w.MiniDownsamples))
		copy(snapshot.MiniDownsamples, w.MiniDownsamples)
	}

	ps := w.progressStreamer
	if ps != nil {
		snapshot.Progress = ps.Progress()
		snapshot.Position = ps.Position()
		snapshot.AbsolutePosition = ps.AbsolutePosition()
	}

	return snapshot
}

// WaveFileSnapshot is an immutable view of WaveFile state
type WaveFileSnapshot struct {
	Path       string
	Name       string
	SampleRate int
	BitDepth   int
	NumSamples int

	Downsamples     []Downsample
	MiniDownsamples []Downsample
	MinY, MaxY      float32

	MetadataLoaded bool
	SamplesLoaded  bool
	LoadErr        error

	Progress         float64
	Position         int
	AbsolutePosition int
}
