package waveform

type localCommand int

const (
	cmdSetWaveDisplayData localCommand = iota
	cmdUpdatePlaybackProgress
	cmdSetWaveSlices
	cmdSetWaveBounds
	cmdSetWaveBoundsFromSamples
	cmdSetWaveCursor
	cmdSetWavePlotFlags
	cmdSetWaveAxisXFlags
	cmdSetWaveAxisYFlags
	cmdAddWaveSlice
	cmdUpdateWaveSlicePosition
)

type WaveBoundsPayload struct {
	Start float64
	End   float64
}

type WaveBoundsSamplesPayload struct {
	StartSample int
	EndSample   int
}

type WaveSlicePositionPayload struct {
	Index    int
	NewStart float64
}

type PlaybackProgressUpdate struct {
	IsPlaying       bool
	Progress        float64
	PositionSeconds float64
}
