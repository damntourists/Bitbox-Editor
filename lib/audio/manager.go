package audio

import "sync"

type AudioManager struct {
	currentWave *WaveFile
	mu          sync.Mutex
}

var globalAudioManager = &AudioManager{}

func GetAudioManager() *AudioManager {
	return globalAudioManager
}

func (a *AudioManager) PlayWave(wave *WaveFile) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentWave != nil && a.currentWave.IsPlaying() {
		a.currentWave.Stop()
	}

	a.currentWave = wave
	if wave != nil {
		wave.Play()
	}
}

func (a *AudioManager) StopCurrent() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentWave != nil && a.currentWave.IsPlaying() {
		a.currentWave.Stop()
		a.currentWave = nil
	}
}

func (a *AudioManager) CurrentWave() *WaveFile {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentWave
}
