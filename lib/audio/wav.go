package audio

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
	beepWav "github.com/gopxl/beep/v2/wav"
)

var sampleRate = beep.SampleRate(44100)

func init() {
	err := speaker.Init(sampleRate, sampleRate.N(time.Second/10))
	if err != nil {
		panic(err)
		return
	}
}

type (
	Downsample struct {
		Mins, Maxs []int64
	}

	WaveFile struct {
		decoder *wav.Decoder
		pcm     *audio.IntBuffer
		file    *os.File

		Path        string
		Name        string
		SampleRate  int
		BitDepth    int
		Channels    [][]int64
		Downsamples []Downsample

		MinY, MaxY int64

		Len int

		loading bool
		loaded  bool
		playing bool

		streamer         beep.StreamSeeker
		progressStreamer *progressStreamer
		format           beep.Format

		bgOnce sync.Once
		mu     sync.Mutex
	}
)

func (w *WaveFile) IsLoading() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.loading
}

func (w *WaveFile) IsLoaded() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.loaded
}

func (w *WaveFile) IsPlaying() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.playing
}

func (w *WaveFile) Progress() float64 {
	if w.progressStreamer == nil {
		return 0.0
	}
	return w.progressStreamer.Progress()
}

func (w *WaveFile) Position() int {
	if w.progressStreamer == nil {
		return 0
	}
	return w.progressStreamer.Position()
}

func (w *WaveFile) PositionSeconds() float64 {
	if w.SampleRate == 0 {
		return 0.0
	}
	return float64(w.Position()) / float64(w.SampleRate)
}

func (w *WaveFile) setPlaying(v bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.playing = v
}

func (w *WaveFile) SetLoading(v bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.loading = v
}

func (w *WaveFile) Play() {
	log.Debug(fmt.Sprintf("Playing %s", w.Path))

	if w.progressStreamer == nil {
		log.Error("No progress streamer")
		return
	}

	err := w.progressStreamer.Seek(0)
	if err != nil {
		log.Error(err.Error())
		return
	}

	w.setPlaying(true)

	// Resample to 44100
	resampler := beep.Resample(4, w.format.SampleRate, sampleRate, w.progressStreamer)

	speaker.Play(
		beep.Seq(resampler, beep.Callback(func() {
			log.Debug(fmt.Sprintf("Finished playing %s", w.Path))
			w.setPlaying(false)
			w.progressStreamer.Reset()
		})),
	)
}

func (w *WaveFile) Stop() {
	if w.IsPlaying() {
		log.Debug(fmt.Sprintf("Stopping %s", w.Path))
		speaker.Clear()
		w.setPlaying(false)
		if w.progressStreamer != nil {
			w.progressStreamer.Reset()
		}
	}
}

func (w *WaveFile) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *WaveFile) Load() error {
	if w.loading || w.loaded {
		return nil
	}

	w.SetLoading(true)
	defer w.SetLoading(false)

	f, err := os.Open(w.Path)
	if err != nil {
		return err
	}

	if w.streamer == nil {
		w.streamer, w.format, err = beepWav.Decode(f)

		if err != nil {
			return err
		}

		totalSamples := w.streamer.Len()

		w.progressStreamer = &progressStreamer{
			StreamSeeker: w.streamer,
			totalSamples: totalSamples,
			position:     0,
			progress:     0,
		}

		w.file = f
	}

	f.Seek(0, io.SeekStart)

	if w.decoder == nil {
		w.decoder = wav.NewDecoder(f)
		if !w.decoder.IsValidFile() {
			return errors.New("invalid wav file")
		}
	}

	if w.pcm == nil {
		w.pcm, err = w.decoder.FullPCMBuffer()
		if err != nil {
			return err
		}
	}

	if w.pcm == nil || w.pcm.Data == nil || w.pcm.Format == nil {
		return errors.New("empty PCM buffer")
	}

	ch := w.pcm.Format.NumChannels
	sr := w.pcm.Format.SampleRate
	n := len(w.pcm.Data) / ch
	if n == 0 {
		return errors.New("no frames")
	}

	chans := make([][]int64, ch)
	downsamples := make([]Downsample, ch)
	for c := 0; c < ch; c++ {
		chans[c] = make([]int64, n)
	}

	for i := 0; i < n; i++ {
		base := i * ch
		for c := 0; c < ch; c++ {
			chans[c][i] = int64(w.pcm.Data[base+c])
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

		mins, maxs := DownsampleMinMaxInt64(ys, 2000)
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

	w.SampleRate = sr
	w.BitDepth = int(w.decoder.BitDepth)
	w.Channels = chans
	w.Downsamples = downsamples
	w.Len = n
	w.MinY = minY
	w.MaxY = maxY

	w.loaded = true

	return nil
}

func NewWaveFileLazy(name, path string) *WaveFile {
	w := &WaveFile{
		Name:    name,
		Path:    path,
		loaded:  false,
		loading: false,
	}
	return w
}
