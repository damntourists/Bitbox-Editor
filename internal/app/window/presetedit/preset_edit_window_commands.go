package presetedit

import (
	"bitbox-editor/internal/audio"
)

type localCommand int

const (
	cmdEditSetPreset localCommand = iota
	cmdUpdateCachedProgress
	cmdUpdateButtonStates
	cmdEditSetActiveWave
	cmdHandlePadGridClick
	cmdHandleGridSizeChange
	cmdHandleWaveformClick
	cmdHandleAudioProgress
	cmdHandleAudioPlaybackStarted
	cmdHandleAudioPlaybackPaused
	cmdHandleAudioPlaybackStopped
	cmdHandleAudioPlaybackFinished
	cmdHandleAudioMetadataLoaded
)

type activeWavePayload struct {
	Path        string
	DisplayData audio.WaveDisplayData
}
