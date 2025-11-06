package audio

import (
	"bitbox-editor/internal/logging"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
)

const (
	DefaultSampleRate = beep.SampleRate(44100)
	DefaultChunkSize  = 1024
)

var (
	log                = logging.NewLogger("audio")
	globalAudioManager *AudioManager
)

func init() {

	// Initialize the speaker
	err := speaker.Init(DefaultSampleRate, DefaultSampleRate.N(time.Second/10))
	if err != nil {
		panic(err)
		return
	}

	globalAudioManager = &AudioManager{
		AnalyzerData:   make(chan []float64, 1),
		analyzerBuffer: NewAudioBuffer(DefaultChunkSize),
		stopChan:       make(chan struct{}),
		waveCache:      sync.Map{},
		volume:         0.8,
		commands:       make(chan audioCommand, 100),
	}

	// Initialize cached state
	globalAudioManager.cachedCurrentWavePath.Store("")
	globalAudioManager.cachedIsPlaying.Store(false)
	globalAudioManager.cachedProgressStreamPtr.Store(0)
	globalAudioManager.cachedStartMarker.Store(0)
	globalAudioManager.cachedEndMarker.Store(0)

	// Start the FFT processing goroutine
	go startFFTProcessor(
		globalAudioManager.analyzerBuffer,
		globalAudioManager.AnalyzerData,
		globalAudioManager.stopChan,
	)

	// Start the audio command processor goroutine
	go globalAudioManager.processCommands()
}

func IsAudioFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".wav" || ext == ".WAV"
}
