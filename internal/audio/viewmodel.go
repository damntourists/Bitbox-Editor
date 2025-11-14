package audio

// PlaybackRegion defines the boundaries and loop behavior
type PlaybackRegion struct {
	StartSample int
	EndSample   int
	LoopEnabled bool
}

// WaveDisplayData view model
type WaveDisplayData struct {
	Name             string
	Path             string
	IsLoading        bool
	LoadFailed       bool
	IsReady          bool
	IsPlaying        bool
	Progress         float64
	Downsamples      []Downsample
	MiniDownsamples  []Downsample
	MinY, MaxY       float32
	SampleRate       int
	NumSamples       int
	XLimitMax        float64
	PositionSeconds  float64
	DurationSeconds  float64
	AbsolutePosition int
	SlicePositions   []float64

	PlaybackStartMarker int
	PlaybackEndMarker   int
}
