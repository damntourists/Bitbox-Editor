package audio

import "github.com/gopxl/beep/v2"

// TODO: Finish documenting

type audioCommandType int

const (
	// Basic Playback
	cmdSetVolume audioCommandType = iota
	cmdGetVolume
	cmdClearSpeaker
	cmdPlayStreamer

	// State Management
	cmdSetCurrentWave
	cmdGetCurrentWave
	cmdSetProgressStream
	cmdGetProgressStream
	cmdSetOriginalBounds
	cmdGetOriginalBounds
	cmdSetSeekingFlag

	cmdPlaybackFinished
)

type audioCommand struct {
	Type     audioCommandType
	Data     interface{}
	Response chan interface{}
}

type volumeCommand struct {
	Volume float64
}

type playStreamerCommand struct {
	Streamer beep.Streamer
}

type setCurrentWaveCommand struct {
	Wave *WaveFile
}

type getCurrentWaveResponse struct {
	Wave interface{} // *WaveFile
}

type setProgressStreamCommand struct {
	Stream *ProgressStreamer
}

type getProgressStreamResponse struct {
	Stream interface{} // *ProgressStreamer
}

type setOriginalBoundsCommand struct {
	StartMarker int
	EndMarker   int
}

type getOriginalBoundsResponse struct {
	StartMarker int
	EndMarker   int
	HasPlayback bool
}

type setSeekingFlagCommand struct {
	IsSeeking bool
}

// playbackFinishedCommand is the payload for the finished event
type playbackFinishedCommand struct {
	Path        string
	WaveID      uintptr
	LoopEnabled bool
}
