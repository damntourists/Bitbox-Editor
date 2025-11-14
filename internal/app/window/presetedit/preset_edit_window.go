package presetedit

/*
┍━━━━━━━━━━━━━━━━━━━╳┑
│ Preset Edit Window │
└────────────────────┘
*/

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/component/button"
	"bitbox-editor/internal/app/component/combobox"
	"bitbox-editor/internal/app/component/label"
	"bitbox-editor/internal/app/component/pad"
	"bitbox-editor/internal/app/component/pad_config"
	"bitbox-editor/internal/app/component/padgrid"
	"bitbox-editor/internal/app/component/waveform"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/app/font"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/app/window"
	"bitbox-editor/internal/audio"
	"bitbox-editor/internal/logging"
	"bitbox-editor/internal/preset"
	"fmt"
	"strings"
	"time"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

// TODO: This needs cleaning up.

var log = logging.NewLogger("presetedit")

// formatTimeFromSamples converts samples to time format (mm:ss.ms)
func formatTimeFromSamples(samples int, sampleRate int) string {
	if sampleRate <= 0 {
		return "00:00.00"
	}
	seconds := float64(samples) / float64(sampleRate)
	d := time.Duration(seconds * float64(time.Second))
	minutes := int(d / time.Minute)
	rem := d % time.Minute
	return fmt.Sprintf("%02d:%05.2f", minutes, rem.Seconds())
}

// WaveformState stores the user's edits to a waveform (markers, slices)
type WaveformState struct {
	BoundsStartSample int
	BoundsEndSample   int
	SlicePositions    []float64
}

type PresetEditWindow struct {
	*window.Window[*PresetEditWindow]

	Components struct {
		Wave                  *waveform.WaveComponent
		PadGrid               *padgrid.PadGridComponent
		PadConfig             *pad_config.PadConfigComponent
		PadGridSizeSelect     *combobox.ComboBoxComponent
		PlayFromStartButton   *button.Button
		PlayPauseButton       *button.Button
		StopButton            *button.Button
		SkipBackButton        *button.Button
		SkipForwardButton     *button.Button
		RepeatButton          *button.Button
		WaveLabel             *label.LabelComponent
		ConfigurationLabel    *label.LabelComponent
		PadsLabel             *label.LabelComponent
		PlaybackStatusLabel   *label.LabelComponent
		SliceInfoLabel        *label.LabelComponent
		PositionLabel         *label.LabelComponent
		CursorPositionLabel   *label.LabelComponent
		BoundsStartLabel      *label.LabelComponent
		BoundsEndLabel        *label.LabelComponent
		GeneratePeaksButton   *button.Button
		PeakThresholdSliderID imgui.ID
	}

	preset         *preset.Preset
	loading        bool
	activeWavePath string
	activeWaveData audio.WaveDisplayData
	activePadKey   string
	previousPadKey string

	waveformStates map[string]*WaveformState
	audioManager   *audio.AudioManager
	wasPlaying     bool

	lastPadProgressUpdate   time.Time
	currentlyPlayingPadPath string

	peakThreshold float32
	playbackState *audio.PlaybackState

	eventSub         chan events.Event
	filteredEventSub *eventbus.FilteredSubscription
}

func NewPresetEditWindow(p *preset.Preset, audioMgr *audio.AudioManager) *PresetEditWindow {
	t := theme.GetCurrentTheme()

	w := &PresetEditWindow{
		loading:        false,
		audioManager:   audioMgr,
		preset:         p,
		waveformStates: make(map[string]*WaveformState),
		peakThreshold:  0.5, // Default to 50% threshold
		eventSub:       make(chan events.Event, 100),
	}

	windowTitle := "Preset Editor"
	if p != nil {
		windowTitle = fmt.Sprintf("Preset: %s", p.Name)
	}

	w.Window = window.NewWindow[*PresetEditWindow](windowTitle, "FileMusic", w.handleUpdate)

	baseID := imgui.ID(uintptr(unsafe.Pointer(w)))

	w.Components.Wave = waveform.NewWaveformComponent(baseID + 1)

	w.Components.PadConfig = pad_config.NewPadConfigComponent(baseID+2, p)

	w.Components.PadGridSizeSelect = combobox.NewComboBoxComponent(baseID+3, "##grid-config").
		SetPreview(font.Icon("Grid3x2")).
		SetFlags(imgui.ComboFlagsWidthFitPreview).
		SetItems([]string{"4x2", "4x4"})

	w.Components.WaveLabel = label.NewLabelWithID(baseID+10, "Wave")

	w.Components.WaveLabel.
		SetRounding(2).
		SetBgColor(imgui.Vec4{X: 0.2, Y: 0.3, Z: 0.4, W: 1.0}).
		SetBgHoveredColor(imgui.Vec4{X: 0.3, Y: 0.4, Z: 0.5, W: 1.0})

	w.Components.ConfigurationLabel = label.NewLabelWithID(baseID+11, "Configuration")

	w.Components.PadsLabel = label.NewLabelWithID(baseID+12, "Pads")

	w.Components.PlayFromStartButton = button.NewButtonWithID(baseID+20, font.Icon("StepForward")).
		SetPadding(4).
		SetRounding(4).
		SetOnClick(func() { w.onPlayFromStart() })

	w.Components.PlayPauseButton = button.NewButtonWithID(baseID+21, font.Icon("Play")).
		SetPadding(4).
		SetRounding(4).
		SetOnClick(func() { w.onPlayPause() })

	w.Components.StopButton = button.NewButtonWithID(baseID+22, font.Icon("Square")).
		SetPadding(4).
		SetRounding(4).
		SetOnClick(func() { w.onStop() })

	w.Components.SkipBackButton = button.NewButtonWithID(baseID+23, font.Icon("ChevronLeft")).
		SetPadding(4).
		SetRounding(4).
		SetOnClick(func() { w.onSkipBack() })

	w.Components.SkipForwardButton = button.NewButtonWithID(baseID+24, font.Icon("ChevronRight")).
		SetPadding(4).
		SetRounding(4).
		SetOnClick(func() { w.onSkipForward() })

	w.Components.RepeatButton = button.NewButtonWithID(baseID+25, font.Icon("Repeat")).
		SetPadding(4).
		SetRounding(4).
		SetOnClick(func() { w.onRepeat() })

	w.Components.GeneratePeaksButton = button.NewButtonWithID(baseID+26, font.Icon("Sparkles")).
		SetPadding(4).
		SetRounding(4).
		SetOnClick(func() { w.onGeneratePeaks() })

	w.Components.PlaybackStatusLabel = label.NewLabelWithID(baseID+30, "Stopped").
		SetPadding(4).
		SetRounding(4).
		SetBgColor(imgui.Vec4{X: 0.6, Y: 0.6, Z: 0.6, W: 1.0}).
		SetFgColor(t.Style.Colors.Text.Vec4)

	w.Components.SliceInfoLabel = label.NewLabelWithID(baseID+31, "").
		SetPadding(4).
		SetRounding(4).
		SetBgColor(t.Style.Colors.FrameBg.Vec4).
		SetFgColor(t.Style.Colors.Text.Vec4)

	w.Components.PositionLabel = label.NewLabelWithID(baseID+32, "0.00s").
		SetPadding(4).
		SetRounding(4).
		SetBgColor(t.Style.Colors.FrameBg.Vec4).
		SetFgColor(t.Style.Colors.Text.Vec4)

	w.Components.CursorPositionLabel = label.NewLabelWithID(baseID+33, "[00:00.00]").
		SetPadding(4).
		SetRounding(4).
		SetBgColor(t.Style.Colors.FrameBg.Vec4).
		SetFgColor(t.Style.Colors.Text.Vec4)

	w.Components.BoundsStartLabel = label.NewLabelWithID(baseID+34, "Start: 0").
		SetPadding(4).
		SetRounding(4).
		SetBgColor(t.Style.Colors.FrameBg.Vec4).
		SetFgColor(t.Style.Colors.Text.Vec4)

	w.Components.BoundsEndLabel = label.NewLabelWithID(baseID+35, "End: 0").
		SetPadding(4).
		SetRounding(4).
		SetBgColor(t.Style.Colors.FrameBg.Vec4).
		SetFgColor(t.Style.Colors.Text.Vec4)

	w.Components.PadGrid = padgrid.NewPadGrid(
		baseID+40,
		2, 4, 100,
	)

	bus := eventbus.Bus
	uuid := w.UUID()

	w.Components.PadGrid.SetOwnerID(uuid)

	if p != nil {
		w.Components.PadGrid.SetPreset(p)
		go w.preloadPresetWavs(p)
	}

	w.Window.SetLayoutBuilder(w)

	// Create filtered subscription for owned events (audio/MIDI/pad clicks)
	w.filteredEventSub = eventbus.NewFilteredSubscription(uuid, 100)
	w.filteredEventSub.SubscribeMultiple(
		bus,
		events.AudioPlaybackProgressKey,
		events.AudioPlaybackStartedKey,
		events.AudioPlaybackPausedKey,
		events.AudioPlaybackStoppedKey,
		events.AudioPlaybackFinishedKey,
		events.PadGridSelectKey,
		// TODO: Add MIDI events here when ready:
		// events.MidiPlaybackNoteOnKey,
		// events.MidiPlaybackNoteOffKey,
		// ...
	)

	// Subscribe to non-owned events (UI interactions, file loading)
	bus.Subscribe(events.ComboboxSelectionChangeEventKey, uuid, w.eventSub)
	bus.Subscribe(events.ComponentClickEventKey, uuid, w.eventSub)
	bus.Subscribe(events.AudioMetadataLoadedKey, uuid, w.eventSub)
	bus.Subscribe(events.AudioSamplesLoadedKey, uuid, w.eventSub)
	bus.Subscribe(events.AudioLoadFailedKey, uuid, w.eventSub)

	return w
}

// drainEvents translates global bus events into local commands
func (w *PresetEditWindow) drainEvents() {
	if w.filteredEventSub != nil {
		for {
			select {
			case event := <-w.filteredEventSub.Events():
				var cmd component.UpdateCmd
				switch event.Type() {
				case events.AudioPlaybackProgressKey:
					cmd = component.UpdateCmd{Type: cmdHandleAudioProgress, Data: event}
				case events.AudioPlaybackStartedKey, events.AudioPlaybackPausedKey, events.AudioPlaybackStoppedKey, events.AudioPlaybackFinishedKey:
					cmd = component.UpdateCmd{Type: cmdHandleAudioStartStop, Data: event}
				case events.PadGridSelectKey:
					cmd = component.UpdateCmd{Type: cmdHandlePadGridClick, Data: event}
					// TODO: Add MIDI events
					// case events.MidiPlaybackNoteOnKey:
					//     cmd = UpdateCmd{Type: cmdHandleMidiNoteOn, Data: event}
					// ...
				}

				if cmd.Type != 0 {
					w.SendUpdate(cmd)
				}
			default:
				return
			}
		}
	}
}

func (w *PresetEditWindow) handleUpdate(cmd component.UpdateCmd) {
	switch c := cmd.Type.(type) {
	case component.GlobalCommand:
		w.Window.HandleGlobalUpdate(cmd)
		return

	case window.GlobalCommand:
		w.Window.HandleGlobalUpdate(cmd)
		return

	case localCommand:
		switch c {
		case cmdEditSetPreset:
			if p, ok := cmd.Data.(*preset.Preset); ok {
				w.preset = p
				if w.Components.PadGrid != nil {
					w.Components.PadGrid.SetPreset(p)
				}
				if w.Components.PadConfig != nil {
					w.Components.PadConfig.SetPreset(p)
				}
				w.activeWavePath = ""
				w.activeWaveData = audio.WaveDisplayData{}
				if w.Components.Wave != nil {
					w.Components.Wave.SetWaveDisplayData(w.activeWaveData)
				}
				go w.preloadPresetWavs(p)
			}

		case cmdUpdateCachedProgress:
			if progress, ok := cmd.Data.(float64); ok {
				if w.playbackState != nil {
					w.playbackState.SetCursorFromProgress(progress)
				}
			}

		case cmdUpdateButtonStates:
			if isPlaying, ok := cmd.Data.(bool); ok {
				w.updateButtonStates(isPlaying)
			}

		case cmdEditSetActiveWave:
			if payload, ok := cmd.Data.(activeWavePayload); ok {
				if w.previousPadKey != "" && w.Components.Wave != nil {
					boundsStartSample, boundsEndSample, slicePositions := w.Components.Wave.GetBoundsAndSlices()
					w.waveformStates[w.previousPadKey] = &WaveformState{
						BoundsStartSample: boundsStartSample,
						BoundsEndSample:   boundsEndSample,
						SlicePositions:    slicePositions,
					}
				}

				w.activeWavePath = payload.Path
				w.activeWaveData = payload.DisplayData
				if w.Components.Wave != nil {
					w.Components.Wave.ResetForNewWave()
					// Restore state for current pad (if previously saved)
					if savedState, exists := w.waveformStates[w.activePadKey]; exists {
						isFullRange := savedState.BoundsStartSample == 0 &&
							savedState.BoundsEndSample >= payload.DisplayData.NumSamples-500

						if isFullRange && len(savedState.SlicePositions) == 0 {
							cleanDisplayData := payload.DisplayData
							cleanDisplayData.PlaybackStartMarker = 0
							cleanDisplayData.PlaybackEndMarker = 0
							w.Components.Wave.SetWaveDisplayData(cleanDisplayData)
							w.Components.Wave.ClearSlices()
						} else {
							cleanDisplayData := payload.DisplayData
							cleanDisplayData.PlaybackStartMarker = 0
							cleanDisplayData.PlaybackEndMarker = 0
							w.Components.Wave.SetWaveDisplayData(cleanDisplayData)
							w.Components.Wave.SetBoundsFromSamples(savedState.BoundsStartSample, savedState.BoundsEndSample)
							if len(savedState.SlicePositions) > 0 {
								slices := make([]*waveform.WaveMarker, len(savedState.SlicePositions))
								for i, pos := range savedState.SlicePositions {
									slices[i] = waveform.NewWaveMarker(pos)
								}
								w.Components.Wave.SetSlices(slices)
							} else {
								w.Components.Wave.ClearSlices()
							}
						}
					} else {
						cleanDisplayData := payload.DisplayData
						cleanDisplayData.PlaybackStartMarker = 0
						cleanDisplayData.PlaybackEndMarker = 0
						w.Components.Wave.SetWaveDisplayData(cleanDisplayData)
						w.Components.Wave.ClearSlices()
					}
				}
			}

		case cmdHandlePadGridClick:
			if event, ok := cmd.Data.(events.PadGridEventRecord); ok {
				if pc, ok := event.Pad.(*pad.PadComponent); ok && pc != nil {
					wavePath := pc.GetWavePath()
					if wavePath != "" && w.audioManager != nil {
						w.Components.PadConfig.SetPad(pc)
						w.audioManager.ClearPlaybackRegion(wavePath)

						// Save previous pad key and set new one
						w.previousPadKey = w.activePadKey
						padKey := fmt.Sprintf("%d_%d", pc.Row(), pc.Col())
						w.activePadKey = padKey

						// Set the active wave path
						w.activeWavePath = wavePath

						// Try to get display data
						displayData, err := w.audioManager.GetWaveDisplayData(wavePath)
						if err == nil && displayData.NumSamples > 0 {
							w.SendUpdate(
								component.UpdateCmd{
									Type: cmdEditSetActiveWave,
									Data: activeWavePayload{
										Path:        wavePath,
										DisplayData: displayData,
									},
								},
							)
						}

						wasPlaying := w.audioManager != nil && w.audioManager.IsPlaying()
						isSamePad := w.activePadKey == padKey
						isFirstPadClick := w.activePadKey == ""
						shouldPlay := wasPlaying || isSamePad || isFirstPadClick

						var boundsStart, boundsEnd int
						var slicePositions []float64
						if savedState, exists := w.waveformStates[padKey]; exists {
							boundsStart = savedState.BoundsStartSample
							boundsEnd = savedState.BoundsEndSample
							slicePositions = savedState.SlicePositions
						} else {
							boundsStart = 0
							boundsEnd = displayData.NumSamples
							slicePositions = []float64{}
						}

						if shouldPlay {
							preservedRepeatMode := audio.RepeatModeOff
							if w.playbackState != nil {
								preservedRepeatMode = w.playbackState.RepeatMode
							}

							w.playbackState = audio.NewPlaybackState(wavePath, boundsStart, boundsEnd)
							w.playbackState.OwnerID = w.UUID()
							w.playbackState.SetRepeatMode(preservedRepeatMode)

							if w.Components.Wave != nil {
								samplesPerBin := w.Components.Wave.GetSamplesPerBin()
								w.playbackState.UpdateBoundsAndSlices(
									boundsStart,
									boundsEnd,
									slicePositions,
									samplesPerBin,
								)
							}

							err = w.audioManager.PlayWithState(w.playbackState)
							if err != nil && !strings.Contains(err.Error(), "not yet loaded") {
								log.Error("Failed to play wave with state on click", zap.Error(err))
							}
						}
					}
				}
			}

		case cmdHandleGridSizeChange:
			if event, ok := cmd.Data.(events.ComboboxEventRecord); ok {
				if event.UUID == w.Components.PadGridSizeSelect.UUID() {
					if layout, ok := event.Selected.(string); ok {
						rows, cols := 2, 4 // Defaults
						switch layout {
						case "4x2":
							rows, cols = 2, 4
						case "4x4":
							rows, cols = 4, 4
						}
						w.Components.PadGrid.SetRows(rows)
						w.Components.PadGrid.SetCols(cols)
					}
				}
			}

		case cmdHandleWaveformClick:
			if event, ok := cmd.Data.(events.MouseEventRecord); ok {
				if event.UUID == w.Components.Wave.UUID() {
					if event.EventType == events.ComponentClickedEvent {
						if data, ok := event.Data.(map[string]interface{}); ok {
							if isSeek, seekOk := data["seek"].(bool); seekOk && isSeek {
								if position, posOk := data["position"].(float64); posOk {
									var path string
									if pathVal, pathOk := data["path"].(string); pathOk {
										path = pathVal
									}
									if w.audioManager != nil {
										var err error
										isPlaying := w.audioManager.IsPlaying() &&
											w.audioManager.CurrentWave().Path == path

										if isPlaying {
											err = w.audioManager.SeekToPosition(position)
										} else if path != "" {
											boundsStart, boundsEnd, _ := w.Components.Wave.GetBoundsAndSlices()
											err = w.audioManager.SetCursorPositionByPath(
												path,
												position,
												boundsStart,
												boundsEnd,
											)
										} else {
											err = w.audioManager.SeekToPosition(position)
										}
										if err != nil {
											log.Error("Failed to set cursor/seek", zap.Error(err))
										}
									}
								}
							}
						}
					}
				}
			}

		case cmdHandleAudioProgress:
			if event, ok := cmd.Data.(events.AudioPlaybackEventRecord); ok {
				if event.Path == w.activeWavePath {
					w.activeWaveData.Progress = event.Progress
					w.activeWaveData.IsPlaying = event.IsPlaying
					if w.activeWaveData.SampleRate > 0 {
						w.activeWaveData.PositionSeconds = float64(event.PositionSamples) / float64(w.activeWaveData.SampleRate)
					}

					if w.Components.Wave != nil {
						w.Components.Wave.SetWaveDisplayData(w.activeWaveData)
					}

					w.SendUpdate(component.UpdateCmd{Type: cmdUpdateButtonStates, Data: event.IsPlaying})
				}

				now := time.Now()
				if now.Sub(w.lastPadProgressUpdate) >= 100*time.Millisecond {
					w.lastPadProgressUpdate = now
					if w.Components.PadGrid != nil {
						for _, p := range w.Components.PadGrid.Pads() {
							if p != nil && p.GetWavePath() == event.Path {
								displayData := p.GetWaveDisplayData()
								if displayData.Path != "" {
									displayData.Progress = event.Progress
									displayData.IsPlaying = event.IsPlaying
									p.SetWaveDisplayData(displayData)
								}
								break
							}
						}
					}
				}
			}

		case cmdHandleAudioStartStop:
			if event, ok := cmd.Data.(events.AudioPlaybackEventRecord); ok {
				if event.EventType == events.AudioPlaybackStartedEvent {
					if event.Path == w.activeWavePath {
						w.activeWaveData.IsPlaying = true
						if w.Components.Wave != nil {
							w.Components.Wave.SetWaveDisplayData(w.activeWaveData)
						}

						if w.playbackState != nil && w.playbackState.Path == w.activeWavePath {
							w.playbackState.Play()
							log.Debug("Playback state set to playing", zap.String("path", w.activeWavePath))
						}

						w.SendUpdate(component.UpdateCmd{Type: cmdUpdateButtonStates, Data: true})
					}

					if w.currentlyPlayingPadPath != "" &&
						w.currentlyPlayingPadPath != event.Path &&
						w.Components.PadGrid != nil {
						for _, pad := range w.Components.PadGrid.Pads() {
							if pad != nil && pad.GetWavePath() == w.currentlyPlayingPadPath {
								displayData := pad.GetWaveDisplayData()
								if displayData.Path != "" {
									displayData.Progress = 0
									displayData.IsPlaying = false
									pad.SetWaveDisplayData(displayData)
								}
								break
							}
						}
					}
					w.currentlyPlayingPadPath = event.Path

				} else if event.EventType == events.AudioPlaybackPausedEvent {
					if event.Path == w.activeWavePath {

						w.activeWaveData.IsPlaying = false
						if w.Components.Wave != nil {
							w.Components.Wave.SetWaveDisplayData(w.activeWaveData)
						}

						if w.playbackState != nil && w.playbackState.Path == w.activeWavePath {
							w.playbackState.IsPlaying = event.IsPlaying
							w.playbackState.IsPaused = event.IsPaused
						}

						w.SendUpdate(component.UpdateCmd{Type: cmdUpdateButtonStates, Data: false})
					}

				} else if event.EventType == events.AudioPlaybackStoppedEvent ||
					event.EventType == events.AudioPlaybackFinishedEvent {
					if event.Path == w.activeWavePath {
						isActuallyPlaying := w.audioManager != nil && w.audioManager.IsPlaying() &&
							w.audioManager.CurrentWave().Path == event.Path

						if !isActuallyPlaying {
							w.activeWaveData.IsPlaying = false
							w.activeWaveData.Progress = 0
							w.activeWaveData.PositionSeconds = 0

							if w.Components.Wave != nil {
								w.Components.Wave.SetWaveDisplayData(w.activeWaveData)
							}

							if w.playbackState != nil && w.playbackState.Path == w.activeWavePath {
								w.playbackState.IsPlaying = event.IsPlaying
								w.playbackState.IsPaused = event.IsPaused
							}
						}

						w.SendUpdate(component.UpdateCmd{Type: cmdUpdateButtonStates, Data: false})
					}

					if w.Components.PadGrid != nil {
						for _, pad := range w.Components.PadGrid.Pads() {
							if pad != nil && pad.GetWavePath() == event.Path {
								displayData := pad.GetWaveDisplayData()
								if displayData.Path != "" {
									displayData.Progress = 0
									displayData.IsPlaying = false
									pad.SetWaveDisplayData(displayData)
								}
								break
							}
						}
					}
					if w.currentlyPlayingPadPath == event.Path {
						w.currentlyPlayingPadPath = ""
					}
				}
			}

		case cmdHandleAudioLoad:
			if event, ok := cmd.Data.(events.AudioLoadEventRecord); ok {
				if event.EventType == events.AudioMetadataLoadedEvent && event.Path == w.activeWavePath {
					displayData, err := w.audioManager.GetWaveDisplayData(event.Path)
					if err == nil {
						if displayData.NumSamples > 0 {
							w.SendUpdate(component.UpdateCmd{Type: cmdEditSetActiveWave, Data: activeWavePayload{
								Path:        event.Path,
								DisplayData: displayData,
							}})
						} else {
							log.Warn("Display data has NumSamples = 0, not setting active wave",
								zap.String("path", event.Path))
						}
					} else {
						log.Error("Failed to get wave display data", zap.Error(err))
					}

					// Get bounds for active pad
					var boundsStart, boundsEnd int
					if savedState, exists := w.waveformStates[w.activePadKey]; exists {
						boundsStart = savedState.BoundsStartSample
						boundsEnd = savedState.BoundsEndSample
					} else if displayData.NumSamples > 0 {
						boundsStart = 0
						boundsEnd = displayData.NumSamples
					}

					err = w.audioManager.PlayWaveByPath(event.Path, false, boundsStart, boundsEnd)
					if err != nil {
						log.Error("Failed to play wave after metadata load", zap.Error(err))
					}
				}
			}
		}
		return

	default:
		log.Warn("PresetEditWindow unhandled update type", zap.Any("type", fmt.Sprintf("%T", cmd.Type)))
	}
}

func (w *PresetEditWindow) Menu() {}
func (w *PresetEditWindow) Layout() {
	w.drainEvents()
	w.Window.ProcessUpdates()

	currentPreset := w.preset
	isLoading := w.loading
	waveData := w.activeWaveData

	t := theme.GetCurrentTheme()

	if currentPreset == nil {
		imgui.Text("No preset selected")
		return
	}

	if isLoading {
		imgui.Text("Loading...")
		return
	}

	availHeight := imgui.ContentRegionAvail().Y
	defaultWaveformHeight := availHeight * 0.5
	if defaultWaveformHeight < 200 {
		defaultWaveformHeight = 200
	}
	if defaultWaveformHeight > availHeight-100 {
		defaultWaveformHeight = availHeight - 100
	}

	imgui.BeginChildStrV(
		"top-controls",
		imgui.Vec2{X: 0, Y: defaultWaveformHeight},
		imgui.ChildFlagsResizeY|imgui.ChildFlagsBorders,
		imgui.WindowFlagsNone|imgui.WindowFlagsMenuBar,
	)

	if imgui.BeginMenuBar() {
		w.Components.WaveLabel.Build()
		imgui.SameLine()
		imgui.EndMenuBar()
	}

	imgui.BeginChildStrV(
		"transport-controls",
		imgui.Vec2{X: 0, Y: (imgui.FrameHeight() + t.Style.FramePadding[1]*2) * 1.5},
		imgui.ChildFlagsBorders|imgui.ChildFlagsFrameStyle,
		imgui.WindowFlagsNone,
	)

	if waveData.Name != "" {
		isPlaying := waveData.IsPlaying
		w.updateButtonStates(isPlaying)

		w.Components.PlayFromStartButton.Build()

		if imgui.IsItemHovered() {
			imgui.SetTooltip("Play from beginning")
		}

		imgui.SameLine()

		w.Components.PlayPauseButton.Build()

		if isPlaying {
			if imgui.IsItemHovered() {
				imgui.SetTooltip("Pause")
			}
		} else {
			if imgui.IsItemHovered() {
				imgui.SetTooltip("Play from cursor")
			}
		}

		imgui.SameLine()

		w.Components.StopButton.Build()

		if imgui.IsItemHovered() {
			imgui.SetTooltip("Stop")
		}

		imgui.SameLine()

		w.Components.SkipBackButton.Build()

		if imgui.IsItemHovered() {
			imgui.SetTooltip("Previous slice")
		}

		imgui.SameLine()

		w.Components.SkipForwardButton.Build()

		if imgui.IsItemHovered() {
			imgui.SetTooltip("Next slice")
		}

		imgui.SameLine()

		w.Components.RepeatButton.Build()

		if imgui.IsItemHovered() {
			repeatMode := 0
			if w.playbackState != nil {
				repeatMode = int(w.playbackState.RepeatMode)
			}
			switch repeatMode {
			case 0:
				imgui.SetTooltip("Repeat: Off")
			case 1:
				imgui.SetTooltip("Repeat: All")
			case 2:
				imgui.SetTooltip("Repeat: One Slice")
			}
		}

		imgui.SameLine()

		w.Components.PlaybackStatusLabel.Build()

		imgui.SameLine()

		_, _, slicePositions := w.Components.Wave.GetBoundsAndSlices()
		hasSlices := len(slicePositions) > 0
		currentSliceIdx := 0
		if hasSlices {
			samplesPerBin := w.Components.Wave.GetSamplesPerBin()
			if samplesPerBin > 0 {
				var currentBin float64
				isPaused := !isPlaying && waveData.Progress > 0
				if isPaused {
					currentBin = w.Components.Wave.GetCursorPosition()
				} else {
					currentProgress := float64(0)
					if w.audioManager != nil {
						if isPlaying {
							currentProgress = waveData.Progress
						} else if w.playbackState != nil {
							currentProgress = w.playbackState.GetCursorProgress()
						}
					}

					boundsStartSampleCalc, boundsEndSampleCalc, _ := w.Components.Wave.GetBoundsAndSlices()
					boundsRange := float64(boundsEndSampleCalc - boundsStartSampleCalc)
					boundsStartBin := float64(0)
					if samplesPerBin > 0 {
						boundsStartBin = float64(boundsStartSampleCalc) / samplesPerBin
					}
					currentBin = boundsStartBin + (currentProgress * (boundsRange / samplesPerBin))
				}

				for i := 0; i < len(slicePositions); i++ {
					if currentBin >= slicePositions[i]-0.5 {
						currentSliceIdx = i + 1
					}
				}
			}
		}

		if hasSlices {
			w.Components.SliceInfoLabel.
				SetText(fmt.Sprintf("Slice: %d/%d", currentSliceIdx, len(slicePositions))).
				Build()

			imgui.SameLine()
		}

		currentDisplayData := w.Components.Wave.GetWaveDisplayData()
		boundsStartSampleTime, boundsEndSampleTime, _ := w.Components.Wave.GetBoundsAndSlices()
		startTime := formatTimeFromSamples(boundsStartSampleTime, currentDisplayData.SampleRate)
		endTime := formatTimeFromSamples(boundsEndSampleTime, currentDisplayData.SampleRate)

		w.Components.BoundsStartLabel.
			SetText(fmt.Sprintf("Start: %s", startTime)).
			Build()

		imgui.SameLine()

		w.Components.BoundsEndLabel.
			SetText(fmt.Sprintf("End: %s", endTime)).
			Build()

		imgui.SameLine()

		var cursorSample int
		if isPlaying {
			cursorSample = int(waveData.Progress * float64(waveData.NumSamples))
		} else if w.audioManager != nil && w.playbackState != nil {
			cursorSample = int(w.playbackState.GetCursorProgress() * float64(waveData.NumSamples))
		}

		cursorTime := formatTimeFromSamples(cursorSample, waveData.SampleRate)
		w.Components.CursorPositionLabel.SetText(fmt.Sprintf("[%s]", cursorTime))
		w.Components.CursorPositionLabel.Build()

		// Peak detection controls
		imgui.SameLine()
		imgui.SetNextItemWidth(100)
		if imgui.SliderFloatV(
			"##peakThreshold",
			&w.peakThreshold,
			0.1,
			1.0,
			"%.2f",
			imgui.SliderFlagsNone) {
			// Slider value changed
		}

		if imgui.IsItemHoveredV(imgui.HoveredFlagsNone) {
			imgui.SetTooltip("Peak detection threshold (10%% to 100%% of max volume)")
		}

		imgui.SameLine()

		w.Components.GeneratePeaksButton.Build()

		if imgui.IsItemHoveredV(imgui.HoveredFlagsNone) {
			imgui.SetTooltip("Generate slice markers from audio peaks")
		}

	} else {
		imgui.Text("No wave selected")
	}

	imgui.EndChild()

	imgui.BeginChildStrV(
		"waveform",
		imgui.Vec2{X: 0, Y: 0},
		imgui.ChildFlagsBorders,
		imgui.WindowFlagsNone,
	)

	if w.Components.Wave != nil {
		if w.playbackState != nil {
			w.Components.Wave.SetRepeatMode(int(w.playbackState.RepeatMode), w.playbackState.SliceIdx)
		}
		w.Components.Wave.Build()
	}

	imgui.EndChild()

	if waveData.Name != "" && w.audioManager != nil {
		isPlaying := waveData.IsPlaying
		w.wasPlaying = isPlaying
	}

	imgui.EndChild()

	imgui.BeginChildStrV(
		"bottom-section",
		imgui.Vec2{X: 0, Y: -1},
		imgui.ChildFlagsNone,
		imgui.WindowFlagsNone,
	)

	imgui.BeginChildStrV(
		"bottom-left",
		imgui.Vec2{X: imgui.ContentRegionAvail().X * 0.5, Y: 0},
		imgui.ChildFlagsBorders|imgui.ChildFlagsResizeX,
		imgui.WindowFlagsNone|imgui.WindowFlagsMenuBar,
	)

	if imgui.BeginMenuBar() {
		w.Components.ConfigurationLabel.Build()
		imgui.EndMenuBar()
	}
	if w.Components.PadConfig != nil {
		w.Components.PadConfig.Build()
	}

	imgui.EndChild()

	imgui.SameLine()

	imgui.BeginChildStrV(
		"bottom-right",
		imgui.Vec2{X: 0, Y: 0},
		imgui.ChildFlagsBorders,
		imgui.WindowFlagsNone|imgui.WindowFlagsMenuBar,
	)

	if imgui.BeginMenuBar() {
		w.Components.PadsLabel.Build()
		if w.Components.PadGridSizeSelect != nil {
			w.Components.PadGridSizeSelect.Build()
		}
		imgui.SameLine()
		imgui.EndMenuBar()
	}

	if w.Components.PadGrid != nil {
		w.Components.PadGrid.Build()
	}

	imgui.EndChild()

	imgui.EndChild()
}

// ensurePlaybackState ensures playback state exists with valid bounds
func (w *PresetEditWindow) ensurePlaybackState() bool {
	if w.activeWavePath == "" || w.Components.Wave == nil {
		return false
	}

	boundsStart, boundsEnd, slicePositions := w.Components.Wave.GetBoundsAndSlices()
	samplesPerBin := w.Components.Wave.GetSamplesPerBin()

	// Check if bounds are valid
	if boundsEnd <= boundsStart {
		return false
	}

	// Create new playback state if needed, or update existing one
	if w.playbackState == nil || w.playbackState.Path != w.activeWavePath {
		w.playbackState = audio.NewPlaybackState(w.activeWavePath, boundsStart, boundsEnd)
		w.playbackState.OwnerID = w.UUID() // Tag with this window's UUID
	}

	// Update state with current UI values
	w.playbackState.UpdateBoundsAndSlices(boundsStart, boundsEnd, slicePositions, samplesPerBin)

	w.syncPlaybackStateToAudioManager()

	return true
}

// syncPlaybackStateToAudioManager updates the audio manager's stored playback state
func (w *PresetEditWindow) syncPlaybackStateToAudioManager() {
	if w.audioManager == nil || w.playbackState == nil {
		return
	}

	// Only sync if audio is currently playing and matches our active wave
	if w.audioManager.IsPlaying() && w.playbackState.Path == w.activeWavePath {
		// Get the stored playback state from audio manager to compare
		storedState, exists := w.audioManager.GetPlaybackStateByPath(w.playbackState.Path)

		// Check if slices have actually changed
		slicesChanged := false
		if !exists || storedState == nil {
			slicesChanged = len(w.playbackState.SlicePositions) > 0
		} else {
			// Compare slice counts and positions
			if len(storedState.SlicePositions) != len(w.playbackState.SlicePositions) {
				slicesChanged = true
			} else {
				// Same count, check if positions differ
				for i := range w.playbackState.SlicePositions {
					if storedState.SlicePositions[i] != w.playbackState.SlicePositions[i] {
						slicesChanged = true
						break
					}
				}
			}
		}

		w.audioManager.StorePlaybackState(w.playbackState)

		if slicesChanged && w.playbackState.RepeatMode == audio.RepeatModeSlice {
			currentPos := w.audioManager.GetCurrentAbsolutePosition()
			if currentPos >= 0 {
				w.playbackState.CursorPosition = currentPos
			}

			// Restart playback with updated state
			_ = w.audioManager.PlayWithState(w.playbackState)
		}
	}
}

func (w *PresetEditWindow) onPlayFromStart() {
	if w.audioManager == nil || w.activeWavePath == "" {
		return
	}

	if !w.ensurePlaybackState() {
		return
	}

	boundsStartSample, boundsEndSample, _ := w.Components.Wave.GetBoundsAndSlices()
	w.playbackState.SliceIdx = 0

	go func(startMarker, endMarker int, path string) {
		w.audioManager.ClearCursorPosition(path)
		if err := w.audioManager.PlayWaveByPath(path, false, startMarker, endMarker); err != nil {
			log.Debug("Play wave failed (may still be loading)", zap.Error(err))
		}
	}(boundsStartSample, boundsEndSample, w.activeWavePath)
}

func (w *PresetEditWindow) onPlayPause() {
	// Logic is in updateButtonStates
}

func (w *PresetEditWindow) onStop() {
	if !w.ensurePlaybackState() {
		return
	}

	w.playbackState.Stop()

	go func(path string, state *audio.PlaybackState) {
		w.audioManager.StopCurrent()
		// Set cursor position from state
		progress := state.GetCursorProgress()
		w.audioManager.SetCursorPositionByPath(path, progress, state.BoundsStart, state.BoundsEnd)
	}(w.activeWavePath, w.playbackState)
}

func (w *PresetEditWindow) onSkipBack() {
	if !w.ensurePlaybackState() || len(w.playbackState.SlicePositions) == 0 {
		return
	}

	currentSliceStart, _, _ := w.playbackState.GetPlaybackRegion()
	atSliceStart := (w.playbackState.CursorPosition - currentSliceStart) < int(2.0*w.playbackState.SamplesPerBin)

	// Determine target slice
	var targetSliceIdx int
	if atSliceStart {
		targetSliceIdx = w.playbackState.SliceIdx - 1
		if targetSliceIdx < 0 {
			targetSliceIdx = len(w.playbackState.SlicePositions)
		}
	} else {
		// Not at start - restart current slice
		targetSliceIdx = w.playbackState.SliceIdx
	}

	if err := w.playbackState.NavigateToSlice(targetSliceIdx); err != nil {
		log.Error("Failed to navigate to slice", zap.Error(err))
		return
	}

	w.Components.Wave.SetRepeatMode(int(w.playbackState.RepeatMode), w.playbackState.SliceIdx)

	if w.audioManager != nil && w.audioManager.IsPlaying() {
		if err := w.audioManager.PlayWithState(w.playbackState); err != nil {
			log.Error("Failed to apply state after skip back", zap.Error(err))
		}
	} else {
		// Not playing - update audio manager cursor
		progress := w.playbackState.GetCursorProgress()
		w.audioManager.SetCursorPositionByPath(
			w.activeWavePath,
			progress,
			w.playbackState.BoundsStart,
			w.playbackState.BoundsEnd,
		)
	}
}

func (w *PresetEditWindow) onSkipForward() {
	if !w.ensurePlaybackState() || len(w.playbackState.SlicePositions) == 0 {
		return
	}

	nextSliceIdx := w.playbackState.SliceIdx + 1
	if nextSliceIdx > len(w.playbackState.SlicePositions) {
		nextSliceIdx = 0
	}

	if err := w.playbackState.NavigateToSlice(nextSliceIdx); err != nil {
		log.Error("Failed to navigate to slice", zap.Error(err))
		return
	}

	w.Components.Wave.SetRepeatMode(int(w.playbackState.RepeatMode), w.playbackState.SliceIdx)

	if w.audioManager != nil && w.audioManager.IsPlaying() {
		if err := w.audioManager.PlayWithState(w.playbackState); err != nil {
			log.Error("Failed to apply state after skip forward", zap.Error(err))
		}
	} else {
		// Not playing - update audio manager cursor
		progress := w.playbackState.GetCursorProgress()
		w.audioManager.SetCursorPositionByPath(
			w.activeWavePath,
			progress,
			w.playbackState.BoundsStart,
			w.playbackState.BoundsEnd,
		)
	}
}

func (w *PresetEditWindow) onRepeat() {
	if !w.ensurePlaybackState() {
		return
	}

	if w.audioManager != nil && w.audioManager.IsPlaying() {
		currentPos := w.audioManager.GetCurrentAbsolutePosition()
		if currentPos >= 0 {
			w.playbackState.CursorPosition = currentPos
		}
	}

	newMode := (w.playbackState.RepeatMode + 1) % 3
	w.playbackState.SetRepeatMode(audio.RepeatMode(newMode))

	if w.Components.Wave != nil {
		w.Components.Wave.SetRepeatMode(int(w.playbackState.RepeatMode), w.playbackState.SliceIdx)
	}

	if w.audioManager != nil && w.audioManager.IsPlaying() {
		if err := w.audioManager.PlayWithState(w.playbackState); err != nil {
			log.Error("Failed to apply new repeat mode", zap.Error(err))
		}
	}
}

// onGeneratePeaks generates slice markers based on peak detection
func (w *PresetEditWindow) onGeneratePeaks() {
	if w.activeWavePath == "" || w.audioManager == nil {
		log.Warn("Cannot generate peaks: no active wave")
		return
	}

	// Detect peaks using the current threshold
	peaks, err := w.audioManager.DetectPeaksForWave(w.activeWavePath, w.peakThreshold)
	if err != nil {
		log.Error("Failed to detect peaks", zap.Error(err))
		return
	}

	if len(peaks) == 0 {
		log.Info("No peaks detected with current threshold",
			zap.Float32("threshold", w.peakThreshold))
		return
	}

	// Convert peak positions to WaveMarker objects
	slices := make([]*waveform.WaveMarker, len(peaks))
	for i, peakPos := range peaks {
		slices[i] = waveform.NewWaveMarker(peakPos)
	}

	// Apply peaks as slice markers to the waveform
	if w.Components.Wave != nil {
		w.Components.Wave.SetSlices(slices)
	}
}

// updateButtonStates updates button text, enabled state, and colors based on playback state
func (w *PresetEditWindow) updateButtonStates(isPlaying bool) {
	// TODO: Rethink this, it's getting messy

	t := theme.GetCurrentTheme()

	if isPlaying {
		w.Components.PlayPauseButton.
			SetText(font.Icon("Pause")).
			SetOnClick(func() {
				if w.audioManager != nil && w.playbackState != nil {
					// Get current position before pausing
					currentPos := w.audioManager.GetCurrentAbsolutePosition()
					// Update playback state
					w.playbackState.Pause(currentPos)
					// Pause audio
					w.audioManager.PauseCurrent()
				}
			})
	} else {
		w.Components.PlayPauseButton.
			SetText(font.Icon("Play")).
			SetOnClick(func() {
				if w.audioManager != nil && w.activeWavePath != "" {
					// Ensure playback state is initialized with valid bounds
					if !w.ensurePlaybackState() {
						return
					}

					boundsStartSample, boundsEndSample, slicePositions := w.Components.Wave.GetBoundsAndSlices()

					if len(slicePositions) > 0 {
						samplesPerBin := w.Components.Wave.GetSamplesPerBin()

						if samplesPerBin == 0 {
							return
						}

						boundsRange := float64(boundsEndSample - boundsStartSample)
						boundsStartBin := float64(boundsStartSample) / samplesPerBin
						currentProgress := float64(0)

						if w.playbackState != nil {
							currentProgress = w.playbackState.GetCursorProgress()
						}

						isPaused := currentProgress > 0
						if isPaused {
							currentSample := currentProgress * boundsRange
							currentBin := (float64(boundsStartSample) + currentSample) / samplesPerBin
							currentSliceIdx := 0
							for i := 0; i < len(slicePositions); i++ {
								if currentBin >= slicePositions[i]-0.5 {
									currentSliceIdx = i + 1
								}
							}
							sliceIdx := 0
							if w.playbackState != nil {
								sliceIdx = w.playbackState.SliceIdx
							}
							if currentSliceIdx != sliceIdx {
								var sliceStartBin float64
								if sliceIdx > 0 && sliceIdx <= len(slicePositions) {
									sliceStartBin = slicePositions[sliceIdx-1]
								} else {
									sliceStartBin = boundsStartBin
								}
								sliceStartSample := sliceStartBin * samplesPerBin
								relativeSliceSample := sliceStartSample - float64(boundsStartSample)
								sliceProgress := relativeSliceSample / boundsRange
								w.audioManager.SetCursorPositionByPath(
									w.activeWavePath,
									sliceProgress,
									boundsStartSample,
									boundsEndSample,
								)
							}
						} else if w.playbackState != nil && w.playbackState.RepeatMode == 2 {
							var sliceStartBin float64
							boundsStartBin := float64(boundsStartSample) / samplesPerBin
							sliceIdx := w.playbackState.SliceIdx

							if sliceIdx > 0 && sliceIdx <= len(slicePositions) {
								sliceStartBin = slicePositions[sliceIdx-1]
							} else {
								sliceStartBin = boundsStartBin
							}

							sliceStartSample := sliceStartBin * samplesPerBin
							relativeSliceSample := sliceStartSample - float64(boundsStartSample)
							sliceProgress := relativeSliceSample / boundsRange
							w.audioManager.SetCursorPositionByPath(
								w.activeWavePath,
								sliceProgress,
								boundsStartSample,
								boundsEndSample,
							)
						}
					}

					go func(startMarker, endMarker int, path string) {
						if err := w.audioManager.PlayWaveByPath(path, false, startMarker, endMarker); err != nil {
							log.Warn("Play wave failed (may still be loading)", zap.Error(err))
						}
					}(boundsStartSample, boundsEndSample, w.activeWavePath)
				}
			})
	}

	// Update Stop button state
	hasCursor := w.audioManager != nil && w.playbackState != nil && w.playbackState.GetCursorProgress() >= 0
	stopEnabled := isPlaying || hasCursor
	w.Components.StopButton.SetEnabled(stopEnabled)

	// Update Skip buttons - only enabled when slices exist
	_, _, slicePositions := w.Components.Wave.GetBoundsAndSlices()
	hasSlices := len(slicePositions) > 0
	w.Components.SkipBackButton.SetEnabled(hasSlices)
	w.Components.SkipForwardButton.SetEnabled(hasSlices)

	// Update Repeat button
	repeatMode := 0
	if w.playbackState != nil {
		repeatMode = int(w.playbackState.RepeatMode)
	}
	if repeatMode > 0 {
		greenColor := imgui.Vec4{X: 0.2, Y: 0.7, Z: 0.3, W: 1.0}
		w.Components.RepeatButton.SetNormalColor(greenColor).
			SetHoveredColor(imgui.Vec4{X: 0.25, Y: 0.8, Z: 0.35, W: 1.0}).
			SetActiveColor(imgui.Vec4{X: 0.15, Y: 0.6, Z: 0.25, W: 1.0})
	} else {
		w.Components.RepeatButton.SetNormalColor(t.Style.Colors.Button.Vec4).
			SetHoveredColor(t.Style.Colors.ButtonHovered.Vec4).
			SetActiveColor(t.Style.Colors.ButtonActive.Vec4)
	}

	// Update Repeat button icon based on mode
	switch repeatMode {
	case 0:
		w.Components.RepeatButton.SetText(font.Icon("Repeat"))
	case 1:
		w.Components.RepeatButton.SetText(font.Icon("Repeat") + " All")
	case 2:
		w.Components.RepeatButton.SetText(font.Icon("Repeat") + " 1")
	}

	// Update Playback Status Label
	var statusText string
	var statusColor imgui.Vec4
	isPaused := w.playbackState != nil && w.playbackState.IsPaused
	if isPlaying {
		statusText = "Playing"
		statusColor = imgui.Vec4{X: 0.2, Y: 0.8, Z: 0.3, W: 1.0}
	} else if isPaused {
		statusText = "Paused"
		statusColor = imgui.Vec4{X: 0.9, Y: 0.7, Z: 0.2, W: 1.0}
	} else {
		statusText = "Stopped"
		statusColor = imgui.Vec4{X: 0.6, Y: 0.6, Z: 0.6, W: 1.0}
	}
	// TODO: Derive from colormap
	w.Components.PlaybackStatusLabel.
		SetText(statusText).
		AnimateToBgColor(statusColor)
}

func (w *PresetEditWindow) Preset() *preset.Preset {
	return w.preset
}

// preloadPresetWavs requests async loading of all wav files in the preset
func (w *PresetEditWindow) preloadPresetWavs(p *preset.Preset) {
	if p == nil || w.audioManager == nil {
		return
	}

	wavs := p.Wavs()

	if len(wavs) == 0 {
		return
	}

	for _, wav := range wavs {
		if wav.Path != "" {
			go func(path string) {
				_, _ = w.audioManager.GetWaveDisplayData(path)
			}(wav.Path)
		}
	}
}

func (w *PresetEditWindow) Destroy() {
	// Unsubscribe from filtered subscriptions
	if w.filteredEventSub != nil {
		w.filteredEventSub.Unsubscribe()
	}

	// Unsubscribe from legacy events (non-owned events)
	bus := eventbus.Bus
	uuid := w.UUID()
	bus.Unsubscribe(events.PadGridSelectKey, uuid)
	bus.Unsubscribe(events.ComboboxSelectionChangeEventKey, uuid)
	bus.Unsubscribe(events.ComponentClickEventKey, uuid)
	bus.Unsubscribe(events.AudioMetadataLoadedKey, uuid)
	bus.Unsubscribe(events.AudioSamplesLoadedKey, uuid)
	bus.Unsubscribe(events.AudioLoadFailedKey, uuid)

	// Destroy children
	w.Components.BoundsStartLabel.Destroy()
	w.Components.BoundsEndLabel.Destroy()
	w.Components.ConfigurationLabel.Destroy()
	w.Components.CursorPositionLabel.Destroy()
	w.Components.GeneratePeaksButton.Destroy()
	w.Components.PadConfig.Destroy()
	w.Components.PadGrid.Destroy()
	w.Components.PadGridSizeSelect.Destroy()
	w.Components.PadsLabel.Destroy()
	w.Components.PlaybackStatusLabel.Destroy()
	w.Components.PlayFromStartButton.Destroy()
	w.Components.PlayPauseButton.Destroy()
	w.Components.PositionLabel.Destroy()
	w.Components.RepeatButton.Destroy()
	w.Components.SkipBackButton.Destroy()
	w.Components.SkipForwardButton.Destroy()
	w.Components.SliceInfoLabel.Destroy()
	w.Components.StopButton.Destroy()
	w.Components.Wave.Destroy()
	w.Components.WaveLabel.Destroy()

	// Finally, call the base method
	w.Window.Destroy()
}
