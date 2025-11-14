package audio

import (
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
	"go.uber.org/zap"
)

// TODO: Finish documenting

type AudioManager struct {
	currentWave           *WaveFile
	currentProgressStream *ProgressStreamer
	originalStartMarker   int
	originalEndMarker     int
	isSeeking             bool
	// volume is a value from 0.0 to 1.0
	volume float64
	// currentOwnerID is the window/component UUID that initiated current playback
	currentOwnerID string

	cachedCurrentWavePath atomic.Value
	cachedIsPlaying       atomic.Bool
	// cachedProgressStreamPtr is a uintptr of *ProgressStreamer for monitor goroutine
	cachedProgressStreamPtr atomic.Uintptr
	cachedStartMarker       atomic.Int64
	cachedEndMarker         atomic.Int64
	cachedOwnerID           atomic.Value

	analyzerBuffer *AudioBuffer
	AnalyzerData   chan []float64
	stopChan       chan struct{}

	// cursorPositions for tracking cursor positions for stopped wavs (map[string]int)
	cursorPositions sync.Map

	// playbackRegions tracks playback regions for each wav (map[string]*PlaybackRegion)
	playbackRegions sync.Map

	// playbackStates tracks playback state (map[string]*PlaybackState)
	playbackStates sync.Map

	// commands is a channel for all state changes
	commands chan audioCommand

	waveCache sync.Map
}

// getOrLoadWave retrieves a *WaveFile from global cache or loads a new wav file
func (am *AudioManager) getOrLoadWave(path string) (*WaveFile, error) {
	wav := GetOrCreateWaveFile(path)

	if wav.IsMetadataLoaded() && wav.IsSamplesLoaded() {
		return wav, wav.LoadError()
	}

	if !wav.IsMetadataLoaded() {
		err := wav.LoadMetadata()
		if err != nil {
			return nil, fmt.Errorf("failed to load metadata for %s: %w", path, err)
		}
	}

	if !wav.IsSamplesLoaded() {
		err := wav.LoadSamplesAndDownsample()
		if err != nil {
			return wav, fmt.Errorf("failed to load samples for %s (metadata OK): %w", path, err)
		}
	}

	return wav, nil
}

func (am *AudioManager) PlayWaveByPath(path string, loop bool, startMarker, endMarker int) error {
	cache := GetGlobalAsyncCache()
	snapshot := cache.GetSnapshot(path)

	if snapshot == nil || !snapshot.MetadataLoaded {
		cache.RequestLoad(path, LoadMetadataOnly)
		return fmt.Errorf("audio file not yet loaded, loading in background...")
	}

	dummyWave := &WaveFile{Path: path}

	return am.PlayWave(dummyWave, loop, startMarker, endMarker)
}

func (am *AudioManager) PlayWave(wave *WaveFile, loop bool, startMarker, endMarker int) error {
	cache := GetGlobalAsyncCache()
	snapshot := cache.GetSnapshot(wave.Path)
	if snapshot == nil || !snapshot.MetadataLoaded {
		cache.RequestLoad(wave.Path, LoadMetadataOnly)
		return fmt.Errorf("audio file not yet loaded, loading in background...")
	}

	waveStartMarker := startMarker
	waveEndMarker := endMarker
	seekPosition := -1
	if cursorPos, ok := am.cursorPositions.Load(wave.Path); ok {
		cursorSample := cursorPos.(int)
		if cursorSample >= waveStartMarker && cursorSample <= waveEndMarker {
			seekPosition = cursorSample
		} else {
			log.Warn("Stored cursor position outside current bounds, ignoring",
				zap.String("path", wave.Path),
				zap.Int("cursorSample", cursorSample),
				zap.Int("boundsStart", waveStartMarker),
				zap.Int("boundsEnd", waveEndMarker))
		}
	}

	am.setOriginalBounds(waveStartMarker, waveEndMarker)
	am.setProgressStream(nil)
	am.setCurrentWave(wave)

	am.clearSpeaker()

	region := am.GetPlaybackRegion(wave.Path)
	if region != nil && seekPosition >= region.EndSample {
		seekPosition = region.StartSample
		am.cursorPositions.Store(wave.Path, seekPosition)
	}

	rawStreamer, err := StreamerFromSnapshot(snapshot, loop, seekPosition, waveStartMarker, waveEndMarker)
	if err != nil {
		log.Error("Failed to create streamer",
			zap.String("path", wave.Path),
			zap.Error(err))
		return err
	}

	monitorStreamer := NewAudioMonitor(rawStreamer, am.analyzerBuffer)
	volumeStreamer := NewVolumeStreamer(monitorStreamer, am)

	var progressStreamer *ProgressStreamer
	var finalStreamer beep.Streamer
	region = am.GetPlaybackRegion(wave.Path)
	if region != nil {
		regionLength := region.EndSample - region.StartSample
		if region.LoopEnabled {
			remainingSamples := regionLength
			if seekPosition > region.StartSample {
				remainingSamples = region.EndSample - seekPosition
			}
			limitedStreamer := beep.Take(remainingSamples, volumeStreamer)
			totalSamples := waveEndMarker - waveStartMarker
			progressStreamer = NewProgressStreamerWithRegion(
				limitedStreamer,
				waveStartMarker,
				totalSamples,
				region,
				am,
				wave.Path,
				seekPosition,
				true,
			)
			finalStreamer = progressStreamer
		} else {
			remainingSamples := regionLength

			if seekPosition > region.StartSample {
				remainingSamples = region.EndSample - seekPosition
			}

			limitedStreamer := beep.Take(remainingSamples, volumeStreamer)
			totalSamples := waveEndMarker - waveStartMarker
			progressStreamer = NewProgressStreamerWithRegion(
				limitedStreamer,
				waveStartMarker,
				totalSamples,
				region,
				am,
				wave.Path,
				seekPosition,
				false,
			)
			finalStreamer = progressStreamer
		}
	} else {
		totalSamples := waveEndMarker - waveStartMarker
		progressStreamer = NewProgressStreamerWithRegion(
			volumeStreamer,
			waveStartMarker,
			totalSamples,
			nil,
			am,
			wave.Path,
			seekPosition, false,
		)
		finalStreamer = progressStreamer
	}

	wavePath := wave.Path
	waveID := uintptr(unsafe.Pointer(wave))
	finished := beep.Callback(func() {
		loopEnabled := false

		// First check playbackStates
		if stateInterface, ok := am.playbackStates.Load(wavePath); ok {
			if state, ok := stateInterface.(*PlaybackState); ok {
				_, _, loopEnabled = state.GetPlaybackRegion()
			}
		}

		if !loopEnabled {
			region := am.GetPlaybackRegion(wavePath)
			loopEnabled = region != nil && region.LoopEnabled
		}

		select {
		case am.commands <- audioCommand{
			Type: cmdPlaybackFinished,
			Data: playbackFinishedCommand{
				Path:        wavePath,
				WaveID:      waveID,
				LoopEnabled: loopEnabled,
			},
		}:
		default:
			log.Warn("Audio command channel full, dropping PlaybackFinished command")
		}
	})

	am.setProgressStream(progressStreamer)
	am.playStreamer(beep.Seq(finalStreamer, finished))

	// Get owner ID from cached value (set by PlayWithState or defaults to empty)
	ownerID := ""
	if ownerVal := am.cachedOwnerID.Load(); ownerVal != nil {
		if ownerStr, ok := ownerVal.(string); ok {
			ownerID = ownerStr
		}
	}

	eventbus.Bus.Publish(events.AudioPlaybackEventRecord{
		EventType:       events.AudioPlaybackStartedEvent,
		Path:            wave.Path,
		Progress:        0.0,
		PositionSamples: seekPosition,
		DurationSamples: waveEndMarker - waveStartMarker,
		IsPlaying:       true,
		IsPaused:        false,
		OwnerID:         ownerID,
	})

	go am.monitorPlaybackProgress(wave.Path, progressStreamer)

	return nil
}

// monitorPlaybackProgress emits progress events periodically while audio is playing
func (am *AudioManager) monitorPlaybackProgress(path string, progressStream *ProgressStreamer) {
	// ~60fps checks
	//ticker := time.NewTicker(16 * time.Millisecond)
	ticker := time.NewTicker(12 * time.Millisecond)
	defer ticker.Stop()

	// Get owner ID once at start
	ownerID := ""
	if ownerVal := am.cachedOwnerID.Load(); ownerVal != nil {
		if ownerStr, ok := ownerVal.(string); ok {
			ownerID = ownerStr
		}
	}

	lastPublishTime := time.Now()

	for {
		<-ticker.C

		var currentStream *ProgressStreamer

		if ptr := am.cachedProgressStreamPtr.Load(); ptr != 0 {
			currentStream = (*ProgressStreamer)(unsafe.Pointer(ptr))
		}

		stillPlaying := currentStream == progressStream && currentStream != nil

		if !stillPlaying {
			eventbus.Bus.Publish(events.AudioPlaybackEventRecord{
				EventType: events.AudioPlaybackStoppedEvent,
				Path:      path,
				IsPlaying: false,
				OwnerID:   ownerID,
			})
			return
		}

		position := progressStream.Position()
		totalSamples := progressStream.TotalSamples()
		progress := 0.0
		if totalSamples > 0 {
			progress = float64(position) / float64(totalSamples)
		}

		if progress >= 0.999 {
			am.setProgressStream(nil)

			region := am.GetPlaybackRegion(path)
			loopEnabled := region != nil && region.LoopEnabled

			eventbus.Bus.Publish(events.AudioPlaybackEventRecord{
				EventType:   events.AudioPlaybackFinishedEvent,
				Path:        path,
				LoopEnabled: loopEnabled,
				IsPlaying:   false,
				OwnerID:     ownerID,
			})

			return
		}

		// Throttle event publishing to ~30fps
		now := time.Now()
		if now.Sub(lastPublishTime) >= 33*time.Millisecond {
			lastPublishTime = now
			eventbus.Bus.Publish(events.AudioPlaybackEventRecord{
				EventType:       events.AudioPlaybackProgressEvent,
				Path:            path,
				Progress:        progress,
				PositionSamples: position,
				DurationSamples: totalSamples,
				IsPlaying:       true,
				IsPaused:        false,
				OwnerID:         ownerID,
			})
		}
	}
}

func (am *AudioManager) StopCurrent() {
	// Get path from cached value
	pathVal := am.cachedCurrentWavePath.Load()
	path, ok := pathVal.(string)
	if !ok || path == "" {
		log.Debug("Cannot stop: no active wave path")
		return
	}

	log.Debug("Stopping current wave", zap.String("path", path))

	// Update cached flags so the monitor stops
	am.cachedIsPlaying.Store(false)
	am.cachedCurrentWavePath.Store("")
	am.cachedProgressStreamPtr.Store(0)

	// Clear the wave and speaker
	am.setCurrentWave(nil)
	am.clearSpeaker()
	am.setProgressStream(nil)

	// Get owner ID from cached value
	ownerID := ""
	if ownerVal := am.cachedOwnerID.Load(); ownerVal != nil {
		if ownerStr, ok := ownerVal.(string); ok {
			ownerID = ownerStr
		}
	}

	// Emit stopped event
	eventbus.Bus.Publish(events.AudioPlaybackEventRecord{
		EventType:       events.AudioPlaybackStoppedEvent,
		Path:            path,
		IsPlaying:       false,
		IsPaused:        false,
		Progress:        0,
		PositionSamples: 0,
		DurationSamples: 0,
		OwnerID:         ownerID,
	})
}

func (am *AudioManager) PauseCurrent() {
	// Get current position from progress stream before clearing
	currentPos := am.GetCurrentAbsolutePosition()

	// Get path from cached value
	pathVal := am.cachedCurrentWavePath.Load()
	path, ok := pathVal.(string)
	if !ok || path == "" {
		log.Warn("Cannot pause: no active wave path")
		return
	}

	// Store cursor position if we got a valid one
	if currentPos >= 0 {
		am.cursorPositions.Store(path, currentPos)
	}

	// Calculate progress for the stopped event (so waveform keeps cursor visible at paused position)
	boundsStart := int(am.cachedStartMarker.Load())
	boundsEnd := int(am.cachedEndMarker.Load())
	progress := 0.0
	if boundsEnd > boundsStart && currentPos >= boundsStart {
		progress = float64(currentPos-boundsStart) / float64(boundsEnd-boundsStart)
		// Clamp to [0, 1]
		if progress < 0 {
			progress = 0
		}
		if progress > 1 {
			progress = 1
		}
	}

	log.Debug("Calculated pause progress",
		zap.Int("boundsStart", boundsStart),
		zap.Int("boundsEnd", boundsEnd),
		zap.Float64("progress", progress))

	// Update cached flags so monitor stops
	am.cachedIsPlaying.Store(false)
	am.cachedCurrentWavePath.Store("")

	// Clear progress stream pointer so monitor exits immediately
	am.cachedProgressStreamPtr.Store(0)

	// Clear the wave and speaker
	am.setCurrentWave(nil)
	am.clearSpeaker()

	// Clear progress stream  too
	am.setProgressStream(nil)

	// Get owner ID from cached value
	ownerID := ""
	if ownerVal := am.cachedOwnerID.Load(); ownerVal != nil {
		if ownerStr, ok := ownerVal.(string); ok {
			ownerID = ownerStr
		}
	}

	// Emit paused event
	eventbus.Bus.Publish(events.AudioPlaybackEventRecord{
		EventType:       events.AudioPlaybackPausedEvent,
		Path:            path,
		IsPlaying:       false,
		IsPaused:        true,
		Progress:        progress,
		PositionSamples: currentPos,
		DurationSamples: boundsEnd - boundsStart,
		OwnerID:         ownerID,
	})
}

func (am *AudioManager) CurrentWave() *WaveFile {
	if !am.cachedIsPlaying.Load() {
		return nil
	}

	response := make(chan interface{}, 1)

	select {
	case am.commands <- audioCommand{
		Type:     cmdGetCurrentWave,
		Response: response,
	}:
		result := <-response
		if resp, ok := result.(getCurrentWaveResponse); ok {
			if wave, ok := resp.Wave.(*WaveFile); ok {
				return wave
			}
		}
	default:
		log.Warn("Command channel full in CurrentWave")
	}
	return nil
}

func (am *AudioManager) GetCurrentAbsolutePosition() int {
	// Use cached progress stream pointer
	ptr := am.cachedProgressStreamPtr.Load()
	if ptr == 0 {
		return -1
	}
	ps := (*ProgressStreamer)(unsafe.Pointer(ptr))
	return ps.AbsolutePosition()
}

func (am *AudioManager) GetAdjustedProgress() float64 {
	if !am.cachedIsPlaying.Load() {
		return 0.0
	}
	originalStart := int(am.cachedStartMarker.Load())
	originalEnd := int(am.cachedEndMarker.Load())
	if result, ok := am.queryCommand(cmdGetCurrentWave, nil, 10*time.Millisecond); ok {
		if waveResp, ok := result.(getCurrentWaveResponse); ok {
			if wave, ok := waveResp.Wave.(*WaveFile); ok && wave != nil {
				currentPosition := wave.AbsolutePosition()
				originalRange := originalEnd - originalStart
				if originalRange == 0 {
					return 0.0
				}
				adjustedProgress := float64(currentPosition-originalStart) / float64(originalRange)
				return adjustedProgress
			}
		}
	}
	return 0.0
}

func (am *AudioManager) IsPlaying() bool {
	return am.cachedIsPlaying.Load()
}

func (am *AudioManager) GetPlaybackBounds() (int, int, bool) {
	startMarker := int(am.cachedStartMarker.Load())
	endMarker := int(am.cachedEndMarker.Load())
	hasPlayback := am.cachedIsPlaying.Load()
	return startMarker, endMarker, hasPlayback
}

func (am *AudioManager) SeekToPosition(position float64) error {
	isPlaying := am.cachedIsPlaying.Load()

	if !isPlaying {
		return am.SetCursorPosition(position)
	}

	if position < 0 {
		position = 0
	}

	if position > 1 {
		position = 1
	}

	originalStart := int(am.cachedStartMarker.Load())
	originalEnd := int(am.cachedEndMarker.Load())
	totalSamples := originalEnd - originalStart
	seekSample := originalStart + int(float64(totalSamples)*position)

	if result, ok := am.queryCommand(cmdGetCurrentWave, nil, 10*time.Millisecond); ok {
		if waveResp, ok := result.(getCurrentWaveResponse); ok {
			if wave, ok := waveResp.Wave.(*WaveFile); ok && wave != nil {
				path := wave.Path
				am.StopCurrent()
				am.cursorPositions.Store(path, seekSample)
				err := am.PlayWave(wave, false, originalStart, originalEnd)
				if err != nil {
					log.Error(
						"Failed to seek",
						zap.String("path", path),
						zap.Float64("position", position),
						zap.Error(err),
					)
					return err
				}
				return nil
			}
		}
	}
	return fmt.Errorf("failed to seek: unable to get current wave")
}

func (am *AudioManager) SetCursorPosition(position float64) error {
	if position < 0 {
		position = 0
	}
	if position > 1 {
		position = 1
	}

	// Get bounds from cached atomic values
	originalStart := int(am.cachedStartMarker.Load())
	originalEnd := int(am.cachedEndMarker.Load())

	// Calculate the sample position based on bounds
	totalSamples := originalEnd - originalStart
	if totalSamples <= 0 {
		return fmt.Errorf("no wave loaded or invalid bounds")
	}

	cursorSample := originalStart + int(float64(totalSamples)*position)

	// We need to find and update the wave file that corresponds to these bounds
	var foundWave *WaveFile
	am.waveCache.Range(func(key, value interface{}) bool {
		// Search the cache for a wave with matching originalStartMarker and originalEndMarker
		wave := value.(*WaveFile)
		wave.mutex.RLock()

		startMatches := wave.StartMarker == originalStart
		endMatches := wave.EndMarker == originalEnd
		wave.mutex.RUnlock()

		if startMatches && endMatches {
			foundWave = wave
			return false
		}
		return true
	})

	if foundWave != nil {
		// Update the wave's StartMarker so playback will begin from cursor position
		foundWave.mutex.Lock()
		foundWave.StartMarker = cursorSample
		foundWave.mutex.Unlock()
	} else {
		log.Warn("Could not find wave to update cursor position")
	}

	return nil
}

func (am *AudioManager) SetCursorPositionByPath(path string, position float64, startMarker, endMarker int) error {
	if position < 0 {
		position = 0
	}
	if position > 1 {
		position = 1
	}

	// Get snapshot from async cache
	cache := GetGlobalAsyncCache()
	snapshot := cache.GetSnapshot(path)
	if snapshot == nil || !snapshot.MetadataLoaded {
		// Request load and return error
		cache.RequestLoad(path, LoadMetadataOnly)
		return fmt.Errorf("failed to get wave for cursor position: not yet loaded")
	}

	// Calculate the absolute sample position based on bounds
	totalSamples := endMarker - startMarker
	if totalSamples <= 0 {
		return fmt.Errorf("invalid bounds for cursor position")
	}

	// Calculate absolute cursor position within current bounds
	cursorSample := startMarker + int(float64(totalSamples)*position)

	// Store the absolute cursor position for this wave
	am.cursorPositions.Store(path, cursorSample)

	return nil
}

func (am *AudioManager) ClearCursorPosition(path string) {
	am.cursorPositions.Delete(path)
}

func (am *AudioManager) GetCursorProgress(path string) float64 {
	// Check if currently playing via cached state
	if am.cachedIsPlaying.Load() {
		// Check if this is the currently playing path
		if cachedPath, ok := am.cachedCurrentWavePath.Load().(string); ok && cachedPath == path {
			// Get progress stream from cached atomic pointer
			if ptr := am.cachedProgressStreamPtr.Load(); ptr != 0 {
				progressStream := (*ProgressStreamer)(unsafe.Pointer(ptr))

				// Wave is playing - use position from progress streamer
				absolutePosition := progressStream.AbsolutePosition()

				startMarker := int(am.cachedStartMarker.Load())
				endMarker := int(am.cachedEndMarker.Load())

				totalSamples := endMarker - startMarker
				if totalSamples <= 0 {
					return -1
				}

				progress := float64(absolutePosition-startMarker) / float64(totalSamples)

				if progress < 0 {
					progress = 0
				}
				if progress > 1 {
					progress = 1
				}

				return progress
			}
		}
	}

	// Wave is not playing
	return -1
}

func (am *AudioManager) SetPlaybackRegion(path string, startSample, endSample int, loopEnabled bool) {
	region := &PlaybackRegion{
		StartSample: startSample,
		EndSample:   endSample,
		LoopEnabled: loopEnabled,
	}

	am.playbackRegions.Store(path, region)

	if am.cachedIsPlaying.Load() {
		if cachedPath, ok := am.cachedCurrentWavePath.Load().(string); ok && cachedPath == path {
			// Get the wave file from cache
			currentWave := GetOrCreateWaveFile(path)
			if currentWave != nil && currentWave.IsMetadataLoaded() {
				// Store the current absolute position before restarting
				currentPos := am.GetCurrentAbsolutePosition()
				if currentPos < 0 {
					currentPos = startSample
				}

				// Clamp position to new region
				if currentPos < startSample {
					currentPos = startSample
				} else if currentPos >= endSample {
					currentPos = startSample
				}

				// Store as cursor position so PlayWave will use it
				am.cursorPositions.Store(path, currentPos)

				// Get current bounds from cached values
				boundsStart := int(am.cachedStartMarker.Load())
				boundsEnd := int(am.cachedEndMarker.Load())

				// Update stored PlaybackState if one exists for slice looping
				if stateInterface, ok := am.playbackStates.Load(path); ok {
					if state, ok := stateInterface.(*PlaybackState); ok {
						// Make a copy and update cursor
						updatedState := state.Copy()
						updatedState.CursorPosition = currentPos
						am.playbackStates.Store(path, updatedState)
					}
				}

				// Restart playback from the current position when region changes
				_ = am.PlayWave(currentWave, false, boundsStart, boundsEnd)
			}
		}
	}
}

func (am *AudioManager) GetPlaybackRegion(path string) *PlaybackRegion {
	if region, ok := am.playbackRegions.Load(path); ok {
		return region.(*PlaybackRegion)
	}
	return nil
}

func (am *AudioManager) ClearPlaybackRegion(path string) {
	am.playbackRegions.Delete(path)
	log.Debug("Cleared playback region", zap.String("path", filepath.Base(path)))
}

func (am *AudioManager) GetWaveDisplayData(path string) (WaveDisplayData, error) {
	// Check async cache
	cache := GetGlobalAsyncCache()
	snapshot := cache.GetSnapshot(path)

	if snapshot != nil {
		// We have a snapshot, check if it has full samples loaded
		if !snapshot.SamplesLoaded && snapshot.LoadErr == nil {
			// Metadata loaded but samples not loaded yet.
			cache.RequestLoad(path, LoadFullSamples)
		}
		// Return current state (may be loading)
		return am.buildDisplayDataFromSnapshot(snapshot, path)
	}

	// No snapshot yet, request async load and return placeholder
	cache.RequestLoad(path, LoadFullSamples)

	return WaveDisplayData{
		Name:       filepath.Base(path),
		Path:       path,
		IsLoading:  true,
		IsReady:    false,
		LoadFailed: false,
	}, nil
}

func (am *AudioManager) buildDisplayDataFromSnapshot(snapshot *WaveFileSnapshot, path string) (WaveDisplayData, error) {
	// Calculate derived values from snapshot
	xLimitMax := 0.0
	if len(snapshot.Downsamples) > 0 && len(snapshot.Downsamples[0].Mins) > 0 {
		xLimitMax = float64(len(snapshot.Downsamples[0].Mins) - 1)
	}

	duration := 0.0
	if snapshot.SampleRate > 0 {
		duration = float64(snapshot.NumSamples) / float64(snapshot.SampleRate)
	}

	positionSeconds := 0.0
	if snapshot.SampleRate > 0 {
		positionSeconds = float64(snapshot.Position) / float64(snapshot.SampleRate)
	}

	// Get playback state from AudioManager using cached state
	isPlaying := false
	playbackStartMarker := 0
	playbackEndMarker := 0

	if am.cachedIsPlaying.Load() {
		if cachedPath, ok := am.cachedCurrentWavePath.Load().(string); ok && cachedPath == path {
			isPlaying = true
			// Read bounds from cached atomic values
			playbackStartMarker = int(am.cachedStartMarker.Load())
			playbackEndMarker = int(am.cachedEndMarker.Load())
		}
	}

	// Build display data from snapshot
	data := WaveDisplayData{
		Name:                snapshot.Name,
		Path:                snapshot.Path,
		IsLoading:           !snapshot.SamplesLoaded && snapshot.LoadErr == nil,
		IsReady:             snapshot.SamplesLoaded && snapshot.LoadErr == nil,
		LoadFailed:          snapshot.LoadErr != nil,
		IsPlaying:           isPlaying,
		Progress:            snapshot.Progress,
		Downsamples:         snapshot.Downsamples,
		MinY:                snapshot.MinY,
		MaxY:                snapshot.MaxY,
		SampleRate:          snapshot.SampleRate,
		NumSamples:          snapshot.NumSamples,
		XLimitMax:           xLimitMax,
		PositionSeconds:     positionSeconds,
		DurationSeconds:     duration,
		AbsolutePosition:    snapshot.AbsolutePosition,
		PlaybackStartMarker: playbackStartMarker,
		PlaybackEndMarker:   playbackEndMarker,
	}

	var err error
	if snapshot.LoadErr != nil {
		err = snapshot.LoadErr
	}

	return data, err
}

func GetAudioManager() *AudioManager {
	return globalAudioManager
}

func (am *AudioManager) processCommands() {
	for cmd := range am.commands {
		switch cmd.Type {
		case cmdSetVolume:
			if volumeCmd, ok := cmd.Data.(volumeCommand); ok {
				am.volume = volumeCmd.Volume
				eventbus.Bus.Publish(events.AudioVolumeEventRecord{
					EventType: events.AudioVolumeChangedEvent,
					Volume:    volumeCmd.Volume,
				})
			} else {
				log.Error("Invalid VolumeCommand data type")
			}

		case cmdGetVolume:
			if cmd.Response != nil {
				select {
				case cmd.Response <- am.volume:
				default:
					log.Warn("Failed to send GetVolume response")
				}
			}

		case cmdClearSpeaker:
			speaker.Clear()

		case cmdPlayStreamer:
			if playCmd, ok := cmd.Data.(playStreamerCommand); ok {
				streamer := playCmd.Streamer
				if streamer != nil {
					speaker.Play(streamer)
				} else {
					log.Error("Invalid streamer type in PlayStreamerCommand or nil streamer")
				}
			} else {
				log.Error("Invalid PlayStreamerCommand data type")
			}

		case cmdSetCurrentWave:
			if waveCmd, ok := cmd.Data.(setCurrentWaveCommand); ok {
				wave := waveCmd.Wave
				if wave != nil {
					// Valid wave file
					am.currentWave = wave
					am.cachedCurrentWavePath.Store(wave.Path)
					am.cachedIsPlaying.Store(true)
				} else {
					// nil or invalid type - clear state
					am.currentWave = nil
					am.cachedCurrentWavePath.Store("")
					am.cachedIsPlaying.Store(false)
				}
			}

		case cmdGetCurrentWave:
			if cmd.Response != nil {
				select {
				case cmd.Response <- getCurrentWaveResponse{Wave: am.currentWave}:
				default:
					log.Warn("Failed to send GetCurrentWave response")
				}
			}

		case cmdSetProgressStream:
			if streamCmd, ok := cmd.Data.(setProgressStreamCommand); ok {
				stream := streamCmd.Stream
				if stream != nil {
					am.currentProgressStream = stream
					am.cachedProgressStreamPtr.Store(uintptr(unsafe.Pointer(stream)))
				} else {
					am.currentProgressStream = nil
					am.cachedProgressStreamPtr.Store(0)
				}
			}

		case cmdGetProgressStream:
			if cmd.Response != nil {
				select {
				case cmd.Response <- getProgressStreamResponse{Stream: am.currentProgressStream}:
				default:
					log.Warn("Failed to send GetProgressStream response")
				}
			}

		case cmdSetOriginalBounds:
			if boundsCmd, ok := cmd.Data.(setOriginalBoundsCommand); ok {
				am.originalStartMarker = boundsCmd.StartMarker
				am.originalEndMarker = boundsCmd.EndMarker
				am.cachedStartMarker.Store(int64(boundsCmd.StartMarker))
				am.cachedEndMarker.Store(int64(boundsCmd.EndMarker))
			}

		case cmdGetOriginalBounds:
			if cmd.Response != nil {
				response := getOriginalBoundsResponse{
					StartMarker: am.originalStartMarker,
					EndMarker:   am.originalEndMarker,
					HasPlayback: am.currentWave != nil,
				}
				select {
				case cmd.Response <- response:
				default:
					log.Warn("Failed to send GetOriginalBounds response")
				}
			}

		case cmdSetSeekingFlag:
			if seekCmd, ok := cmd.Data.(setSeekingFlagCommand); ok {
				am.isSeeking = seekCmd.IsSeeking
			}

		case cmdPlaybackFinished:
			if event, ok := cmd.Data.(playbackFinishedCommand); ok {
				// Handle looping
				if event.LoopEnabled {
					// Get the stored playback state which has the correct region/slice info
					stateInterface, ok := am.playbackStates.Load(event.Path)
					if !ok {
						log.Warn("Cannot loop: no playback state found",
							zap.String("path", filepath.Base(event.Path)))
						continue
					}

					state, ok := stateInterface.(*PlaybackState)
					if !ok || state == nil {
						log.Warn("Cannot loop: invalid playback state",
							zap.String("path", filepath.Base(event.Path)))
						continue
					}

					// Make a copy to avoid modifying the stored state
					stateCopy := state.Copy()

					// Get the playback region to reset cursor
					regionStart, _, _ := stateCopy.GetPlaybackRegion()
					am.cursorPositions.Store(event.Path, regionStart)

					// Update state's cursor to region start for clean loop
					stateCopy.CursorPosition = regionStart

					// Restart playback using PlayWithState
					_ = am.PlayWithState(stateCopy)
				} else {
					// Not looping. Check if it's still the current wave.
					if am.currentWave != nil && uintptr(unsafe.Pointer(am.currentWave)) == event.WaveID {
						am.setCurrentWave(nil)
						am.setProgressStream(nil)
					}
				}
			}

		default:
			log.Warn("Unknown audio command type", zap.Int("type", int(cmd.Type)))
		}
	}
}

// SetVolume sets the playback volume (0.0 to 1.0)
func (am *AudioManager) SetVolume(volume float64) {
	if volume < 0 {
		volume = 0
	} else if volume > 1 {
		volume = 1
	}

	select {
	case am.commands <- audioCommand{
		Type: cmdSetVolume,
		Data: volumeCommand{Volume: volume},
	}:
	default:
		log.Warn("Audio command channel full, dropping SetVolume command",
			zap.Float64("volume", volume))
	}
}

// GetVolume returns the current volume level (0.0 to 1.0)
func (am *AudioManager) GetVolume() float64 {
	response := make(chan interface{}, 1)
	select {
	case am.commands <- audioCommand{
		Type:     cmdGetVolume,
		Response: response,
	}:
		result := <-response
		if vol, ok := result.(float64); ok {
			return vol
		}
		log.Error("GetVolume received invalid response type")
		return 0.8
	default:
		log.Warn("Audio command channel full, returning default volume")
		return 0.8
	}
}

// volumeToGain converts a volume level (0.0 to 1.0) to a gain value in decibels
func (am *AudioManager) volumeToGain(volume float64) float64 {
	if volume <= 0 {
		return -10
	}
	return (volume - 1.0) * 60.0
}

// queryCommand sends a command and waits for response with timeout
func (am *AudioManager) queryCommand(
	cmdType audioCommandType,
	data interface{},
	timeout time.Duration) (interface{}, bool) {

	response := make(chan interface{}, 1)
	select {
	case am.commands <- audioCommand{
		Type:     cmdType,
		Data:     data,
		Response: response,
	}:
		select {
		case result := <-response:
			return result, true
		case <-time.After(timeout):
			log.Warn(
				"Timeout waiting for command response",
				zap.String("cmdType", fmt.Sprintf("%v", cmdType)),
			)
			return nil, false
		}
	default:
		log.Warn("Command channel full", zap.String("cmdType", fmt.Sprintf("%v", cmdType)))
		return nil, false
	}
}

// setCurrentWave sets the current playing wave
func (am *AudioManager) setCurrentWave(wave *WaveFile) {
	select {
	case am.commands <- audioCommand{
		Type: cmdSetCurrentWave,
		Data: setCurrentWaveCommand{Wave: wave},
	}:
	default:
		log.Error("Audio command channel full when setting current wave!")
	}
}

// setProgressStream sets the current progress stream
func (am *AudioManager) setProgressStream(stream *ProgressStreamer) {
	select {
	case am.commands <- audioCommand{
		Type: cmdSetProgressStream,
		Data: setProgressStreamCommand{Stream: stream},
	}:
	default:
		log.Error("Audio command channel full when setting progress stream!")
	}
}

// setOriginalBounds sets the original playback bounds
func (am *AudioManager) setOriginalBounds(startMarker, endMarker int) {
	select {
	case am.commands <- audioCommand{
		Type: cmdSetOriginalBounds,
		Data: setOriginalBoundsCommand{
			StartMarker: startMarker,
			EndMarker:   endMarker,
		},
	}:
	default:
		log.Error("Audio command channel full when setting original bounds!")
	}
}

// setSeekingFlag sets the seeking flag
func (am *AudioManager) setSeekingFlag(isSeeking bool) {
	select {
	case am.commands <- audioCommand{
		Type: cmdSetSeekingFlag,
		Data: setSeekingFlagCommand{IsSeeking: isSeeking},
	}:
	default:
		log.Error("Audio command channel full when setting seeking flag!")
	}
}

// clearSpeaker stops all playback by clearing the speaker
func (am *AudioManager) clearSpeaker() {
	select {
	case am.commands <- audioCommand{Type: cmdClearSpeaker}:
	default:
		log.Error("Audio command channel full when clearing speaker!")
	}
}

// playStreamer plays a streamer on the speaker
func (am *AudioManager) playStreamer(streamer beep.Streamer) {
	select {
	case am.commands <- audioCommand{
		Type: cmdPlayStreamer,
		Data: playStreamerCommand{Streamer: streamer},
	}:
	default:
		log.Warn("Audio command channel full, dropping PlayStreamer command")
	}
}

// DetectPeaksForWave detects peaks in the waveform based on threshold (0.0 to 1.0)
func (am *AudioManager) DetectPeaksForWave(path string, threshold float32) ([]float64, error) {
	cache := GetGlobalAsyncCache()
	snapshot := cache.GetSnapshot(path)

	if snapshot == nil {
		return nil, fmt.Errorf("wave not loaded: %s", path)
	}

	if !snapshot.SamplesLoaded {
		return nil, fmt.Errorf("wave samples not fully loaded: %s", path)
	}

	if len(snapshot.Downsamples) == 0 {
		return nil, fmt.Errorf("no downsample data available for: %s", path)
	}

	// Use a minimum peak distance of 20 bins (about 0.5 seconds)
	minPeakDistance := 20

	peaks := DetectPeaksFromDownsample(snapshot.Downsamples, threshold, minPeakDistance)

	return peaks, nil
}

// PlayWithState plays audio using a PlaybackState
func (am *AudioManager) PlayWithState(state *PlaybackState) error {
	if err := state.Validate(); err != nil {
		return err
	}

	// Get the playback region from state
	regionStart, regionEnd, loopEnabled := state.GetPlaybackRegion()

	am.cachedOwnerID.Store(state.OwnerID)
	am.cursorPositions.Store(state.Path, state.CursorPosition)
	am.playbackStates.Store(state.Path, state.Copy())

	if loopEnabled || state.RepeatMode != RepeatModeOff {
		am.playbackRegions.Store(state.Path, &PlaybackRegion{
			StartSample: regionStart,
			EndSample:   regionEnd,
			LoopEnabled: loopEnabled,
		})
	} else {
		// Clear region for normal playback
		am.playbackRegions.Delete(state.Path)
	}

	// Call existing PlayWave with bounds
	cache := GetGlobalAsyncCache()
	snapshot := cache.GetSnapshot(state.Path)
	if snapshot == nil || !snapshot.MetadataLoaded {
		cache.RequestLoad(state.Path, LoadMetadataOnly)
		return fmt.Errorf("audio file not yet loaded, loading in background...")
	}

	dummyWave := &WaveFile{Path: state.Path}
	return am.PlayWave(dummyWave, false, state.BoundsStart, state.BoundsEnd)
}

// GetPlaybackState returns a read-only copy of the current playback state
// Returns nil if no playback is active
func (am *AudioManager) GetPlaybackState() *PlaybackState {
	if !am.cachedIsPlaying.Load() {
		return nil
	}

	pathInterface := am.cachedCurrentWavePath.Load()
	if pathInterface == nil {
		return nil
	}

	path, ok := pathInterface.(string)
	if !ok || path == "" {
		return nil
	}

	// Get current position
	currentPos := am.GetCurrentAbsolutePosition()
	if currentPos < 0 {
		currentPos = 0
	}

	// Get bounds
	boundsStart := int(am.cachedStartMarker.Load())
	boundsEnd := int(am.cachedEndMarker.Load())

	// Create state with current values
	state := NewPlaybackState(path, boundsStart, boundsEnd)
	state.CursorPosition = currentPos
	state.IsPlaying = true

	// Try to get region info (legacy) TODO: Revisit this.
	if region, ok := am.playbackRegions.Load(path); ok {
		if r, ok := region.(*PlaybackRegion); ok {
			if r.LoopEnabled {
				if r.StartSample == boundsStart && r.EndSample == boundsEnd {
					state.RepeatMode = RepeatModeAll
				} else {
					state.RepeatMode = RepeatModeSlice
				}
			}
		}
	}

	return state
}

// StorePlaybackState stores a playback state in the audio manager's state map
func (am *AudioManager) StorePlaybackState(state *PlaybackState) {
	if state == nil {
		return
	}
	am.playbackStates.Store(state.Path, state.Copy())
}

// GetPlaybackStateByPath retrieves a stored playback state by path
func (am *AudioManager) GetPlaybackStateByPath(path string) (*PlaybackState, bool) {
	stateInterface, ok := am.playbackStates.Load(path)
	if !ok {
		return nil, false
	}
	state, ok := stateInterface.(*PlaybackState)
	if !ok {
		return nil, false
	}
	return state.Copy(), true
}
