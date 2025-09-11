package audio

import (
	"errors"
	"math"
	"os"

	"github.com/go-audio/wav"
)

type (
	Downsample struct {
		Mins, Maxs []int64
	}

	WaveFile struct {
		Path        string
		SampleRate  int
		BitDepth    int
		Channels    [][]int64
		Downsamples []Downsample

		MinY, MaxY int64

		Len int

		Loading bool
	}

	WavLoadResult struct {
		Wav *WaveFile
		Err error
	}
)

// LoadWavAsync starts decoding in the background
func LoadWavAsync(path string) <-chan WavLoadResult {
	ch := make(chan WavLoadResult, 1)
	go func() {
		wf, err := LoadWAVFile(path)
		ch <- WavLoadResult{Wav: wf, Err: err}
		close(ch)
	}()
	return ch
}

// LoadWAVFile decodes a WAV file into raw
func LoadWAVFile(path string) (*WaveFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := wav.NewDecoder(f)
	if !dec.IsValidFile() {
		return nil, errors.New("invalid wav file")
	}

	intBuf, err := dec.FullPCMBuffer()
	if err != nil {
		return nil, err
	}
	if intBuf == nil || intBuf.Data == nil || intBuf.Format == nil {
		return nil, errors.New("empty PCM buffer")
	}

	ch := intBuf.Format.NumChannels
	sr := intBuf.Format.SampleRate
	n := len(intBuf.Data) / ch
	if n == 0 {
		return nil, errors.New("no frames")
	}

	chans := make([][]int64, ch)
	downsamples := make([]Downsample, ch)
	for c := 0; c < ch; c++ {
		chans[c] = make([]int64, n)
	}

	// Deinterleave into int64
	for i := 0; i < n; i++ {
		base := i * ch
		for c := 0; c < ch; c++ {
			chans[c][i] = int64(intBuf.Data[base+c])
		}
	}

	var (
		minY int64
		maxY int64
		set  bool
	)

	// Record downsamples
	for c := 0; c < len(chans); c++ {
		ys := chans[c]

		mins, maxs := downsampleMinMaxInt64(ys, 2000)
		downsamples[c] = Downsample{Mins: mins, Maxs: maxs}
		for _, v := range ys {
			if !set {
				minY, maxY, set = v, v, true
				continue
			}
			if v < minY {
				minY = v
			}
			if v > maxY {
				maxY = v
			}
		}
	}

	return &WaveFile{
		Path:        path,
		SampleRate:  sr,
		BitDepth:    int(dec.BitDepth),
		Channels:    chans,
		Downsamples: downsamples,
		Len:         n,
		MinY:        minY,
		MaxY:        maxY,
	}, nil
}

// downsampleMinMaxInt64 creates min/max envelopes per bin
func downsampleMinMaxInt64(y []int64, bins int) (mins, maxs []int64) {
	if bins <= 0 || len(y) == 0 {
		return nil, nil
	}
	if bins >= len(y) {
		mins = append([]int64(nil), y...)
		maxs = append([]int64(nil), y...)
		return
	}
	mins = make([]int64, bins)
	maxs = make([]int64, bins)
	binSize := float64(len(y)) / float64(bins)
	for i := 0; i < bins; i++ {
		start := int(math.Floor(float64(i) * binSize))
		end := int(math.Floor(float64(i+1) * binSize))
		if end <= start {
			end = start + 1
			if end > len(y) {
				end = len(y)
			}
		}
		minv, maxv := y[start], y[start]
		for j := start + 1; j < end; j++ {
			if y[j] < minv {
				minv = y[j]
			}
			if y[j] > maxv {
				maxv = y[j]
			}
		}
		mins[i] = minv
		maxs[i] = maxv
	}
	return
}
