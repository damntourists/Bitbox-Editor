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
	cmdHandleAudioStartStop
	cmdHandleAudioLoad
)

type activeWavePayload struct {
	Path        string
	DisplayData audio.WaveDisplayData
}
