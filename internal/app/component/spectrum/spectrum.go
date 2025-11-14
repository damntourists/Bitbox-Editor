package spectrum

/*
╭────────────────────╮
│ Spectrum Component │
╰────────────────────╯
*/

import (
	"bitbox-editor/internal/app/animation"
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/audio"
	"bitbox-editor/internal/config"
	"math"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
)

/*
╭─────────────╮
│ Color Modes │
╰─────────────╯
*/

type SpectrumColorMode int

const (
	// ColorModeHeight - Color based on bar height
	ColorModeHeight SpectrumColorMode = iota
	// ColorModeGradient - Color gradient from left to right
	ColorModeGradient
	// ColorModeStatic - Single static color
	ColorModeStatic
)

type SpectrumAnalyzerComponent struct {
	*component.Component[*SpectrumAnalyzerComponent]

	audioManager *audio.AudioManager

	magnitudes []float64
	barValues  []float32
	barTargets []float32

	numBars    int
	attackTime int64
	decayTime  int64

	noiseGate      float32
	smoothedMaxMag float64

	peakFallSpeed float32
	peakHoldTimes []int64
	peakHoldTime  int64
	peakValues    []float32

	attackEasing animation.EasingFunc
	decayEasing  animation.EasingFunc

	colorMode SpectrumColorMode

	staticColorIdx int32
}

func NewSpectrumAnalyzer(id imgui.ID, audioManager *audio.AudioManager) *SpectrumAnalyzerComponent {
	// Load settings from config
	attackTime := config.GetSpectrumAttackTime()
	attackEasing := getEasingFunc(config.GetSpectrumAttackEasing())
	decayTime := config.GetSpectrumDecayTime()
	decayEasing := getEasingFunc(config.GetSpectrumDecayEasing())
	peakHoldTime := config.GetSpectrumPeakHoldTime()
	peakFallSpeed := config.GetSpectrumPeakFallSpeed()
	noiseGate := config.GetSpectrumNoiseGate()
	colorMode := getColorMode(config.GetSpectrumColorMode())
	staticColorIdx := config.GetSpectrumStaticColorIdx()

	cmp := &SpectrumAnalyzerComponent{
		audioManager:   audioManager,
		magnitudes:     make([]float64, 0),
		barValues:      nil,
		barTargets:     nil,
		peakValues:     nil,
		peakHoldTimes:  nil,
		numBars:        50,
		attackEasing:   attackEasing,
		decayEasing:    decayEasing,
		attackTime:     int64(attackTime),
		decayTime:      int64(decayTime),
		peakFallSpeed:  float32(peakFallSpeed),
		peakHoldTime:   int64(peakHoldTime),
		noiseGate:      float32(noiseGate),
		colorMode:      colorMode,
		staticColorIdx: int32(staticColorIdx),
	}

	cmp.Component = component.NewComponent[*SpectrumAnalyzerComponent](id, cmp.handleUpdate)
	cmp.Component.SetLayoutBuilder(cmp)

	// Set default dimensions
	cmp.Component.SetWidth(200.0)
	cmp.Component.SetHeight(40.0)

	return cmp
}

func (sa *SpectrumAnalyzerComponent) handleUpdate(cmd component.UpdateCmd) {
	if sa.Component.HandleGlobalUpdate(cmd) {
		return
	}
}

func (sa *SpectrumAnalyzerComponent) updateBars() {
	startFreqBin := 3
	maxFreqBin := len(sa.magnitudes)
	currentTime := time.Now().UnixMilli()

	// Find the maximum magnitude for normalization
	currentMaxMag := 0.0
	for i := startFreqBin; i < maxFreqBin; i++ {
		if sa.magnitudes[i] > currentMaxMag {
			currentMaxMag = sa.magnitudes[i]
		}
	}

	// Smooth the maximum magnitude over time
	if sa.smoothedMaxMag < 1e-10 {
		sa.smoothedMaxMag = currentMaxMag
	} else {
		// Fast rise, slow fall
		if currentMaxMag > sa.smoothedMaxMag {
			// Fast attack
			sa.smoothedMaxMag += (currentMaxMag - sa.smoothedMaxMag) * 0.8
		} else {
			// Slow decay
			sa.smoothedMaxMag += (currentMaxMag - sa.smoothedMaxMag) * 0.1
		}
	}

	// Use the smoothed maximum
	maxMag := sa.smoothedMaxMag
	if maxMag < 1e-10 {
		maxMag = 1e-10
	}

	// If smoothed max is very small (silence), use current max to respond faster
	if maxMag < 0.001 && currentMaxMag > maxMag {
		maxMag = currentMaxMag
	}

	for i := 0; i < sa.numBars; i++ {
		// Logarithmic frequency distribution. Maps bar index to frequency bin logarithmically.
		t := float64(i) / float64(sa.numBars-1)
		logT := math.Pow(t, 2.0)

		binIndex := startFreqBin + int(logT*float64(maxFreqBin-startFreqBin))
		if binIndex >= maxFreqBin {
			binIndex = maxFreqBin - 1
		}

		// Get magnitude at this frequency
		mag := sa.magnitudes[binIndex]

		// Normalize to 0-1 based on current maximum
		normalized := mag / maxMag

		// Apply slight compression to make quieter sounds more visible
		normalized = math.Sqrt(normalized)

		if normalized < 0 {
			normalized = 0
		}

		if normalized > 1.0 {
			normalized = 1.0
		}

		// Apply noise gate - values below threshold become 0
		noiseGateThreshold := float64(sa.noiseGate)
		if normalized < noiseGateThreshold {
			normalized = 0
		} else {
			// Scale remaining range to 0-1
			normalized = (normalized - noiseGateThreshold) / (1.0 - noiseGateThreshold)
		}

		targetValue := float32(normalized)

		// Use per-frame interpolation instead of time-based animation for more responsiveness
		// Calculate interpolation factors based on attack/decay times
		// Convert ms to a per-frame factor (~60fps = 16.67ms per frame)
		const frameTime = 16.67
		attackFactor := frameTime / float32(sa.attackTime)
		decayFactor := frameTime / float32(sa.decayTime)

		// Clamp factors
		if attackFactor > 1.0 {
			attackFactor = 1.0
		}
		if decayFactor > 1.0 {
			decayFactor = 1.0
		}

		// Interpolate towards target
		if targetValue > sa.barValues[i] {
			// Rising (attack)
			diff := targetValue - sa.barValues[i]
			sa.barValues[i] += diff * attackFactor
		} else if targetValue < sa.barValues[i] {
			// Falling (decay)
			diff := sa.barValues[i] - targetValue
			sa.barValues[i] -= diff * decayFactor
		}

		// Clamp to valid range
		if sa.barValues[i] < 0 {
			sa.barValues[i] = 0
		}
		if sa.barValues[i] > 1.0 {
			sa.barValues[i] = 1.0
		}

		// Update targets for next frame
		sa.barTargets[i] = targetValue

		// Update peak values
		if sa.barValues[i] > sa.peakValues[i] {
			sa.peakValues[i] = sa.barValues[i]
			sa.peakHoldTimes[i] = currentTime
		} else {
			// Check if peak hold time has expired
			if currentTime-sa.peakHoldTimes[i] > sa.peakHoldTime {
				// Let peak fall
				sa.peakValues[i] -= sa.peakFallSpeed
				if sa.peakValues[i] < sa.barValues[i] {
					sa.peakValues[i] = sa.barValues[i]
				}
				if sa.peakValues[i] < 0 {
					sa.peakValues[i] = 0
				}
			}
		}
	}
}

func (sa *SpectrumAnalyzerComponent) renderSpectrum() {
	t := theme.GetCurrentTheme()

	width := sa.Component.Width()
	if width <= 0 {
		width = 200.0
	}
	height := sa.Component.Height()
	if height <= 0 {
		height = 40.0
	}

	size := imgui.Vec2{X: width, Y: height}
	pos := imgui.CursorScreenPos()

	imgui.InvisibleButtonV(sa.IDStr(), size, imgui.ButtonFlagsNone)

	dl := imgui.WindowDrawList()

	// Draw background
	bgColor := imgui.ColorU32Vec4(t.Style.Colors.Button.Vec4)
	dl.AddRectFilledV(pos, imgui.Vec2{X: pos.X + size.X, Y: pos.Y + size.Y}, bgColor, 4.0, imgui.DrawFlagsNone)

	// Draw border
	borderColor := imgui.ColorU32Vec4(t.Style.Colors.Border.Vec4)
	dl.AddRectV(pos, imgui.Vec2{X: pos.X + size.X, Y: pos.Y + size.Y}, borderColor, 4.0, imgui.DrawFlagsNone, 1.0)

	if len(sa.barValues) == 0 {
		return
	}

	// Calculate bar dimensions
	padding := float32(4.0)
	innerWidth := width - (2 * padding)
	innerHeight := height - (2 * padding)
	barWidth := innerWidth / float32(sa.numBars)
	barGap := float32(1.0)
	actualBarWidth := barWidth - barGap

	if actualBarWidth < 1.0 {
		actualBarWidth = 1.0
	}

	// Draw bars and peaks
	for i := 0; i < sa.numBars; i++ {
		barValue := sa.barValues[i]
		peakValue := sa.peakValues[i]

		// Calculate bar height
		barHeight := barValue * innerHeight
		if barHeight < 0.5 {
			barHeight = 0.5
		}

		// Calculate bar position
		barX := pos.X + padding + float32(i)*barWidth
		barY := pos.Y + padding + innerHeight - barHeight

		// Determine color based on color mode
		var color imgui.Vec4
		switch sa.colorMode {
		case ColorModeHeight:
			// Color based on bar height
			color = implot.SampleColormapV(barValue, theme.GetCurrentColormap())

		case ColorModeGradient:
			// Color gradient from left to right
			t := float32(i) / float32(sa.numBars-1)
			color = implot.SampleColormapV(t, theme.GetCurrentColormap())

		case ColorModeStatic:
			// Single static color from colormap
			color = implot.SampleColormapV(float32(sa.staticColorIdx)/255.0, theme.GetCurrentColormap())
		}

		barColor := imgui.ColorU32Vec4(color)

		// Draw bar
		dl.AddRectFilledV(
			imgui.Vec2{X: barX, Y: barY},
			imgui.Vec2{X: barX + actualBarWidth, Y: pos.Y + padding + innerHeight},
			barColor,
			1.0,
			imgui.DrawFlagsNone,
		)

		// Draw peak indicator
		if peakValue > 0.01 {
			peakY := pos.Y + padding + innerHeight - (peakValue * innerHeight)
			peakHeight := float32(2.0)

			// Peak uses same color logic as bar but brighter
			var peakColor imgui.Vec4
			switch sa.colorMode {
			case ColorModeHeight:
				peakColor = implot.SampleColormapV(peakValue, theme.GetCurrentColormap())
			case ColorModeGradient:
				t := float32(i) / float32(sa.numBars-1)
				peakColor = implot.SampleColormapV(t, theme.GetCurrentColormap())
			case ColorModeStatic:
				peakColor = implot.SampleColormapV(float32(sa.staticColorIdx)/255.0, theme.GetCurrentColormap())
			}

			// Make peak brighter by increasing RGB values
			peakColor.X = peakColor.X*0.5 + 0.5
			peakColor.Y = peakColor.Y*0.5 + 0.5
			peakColor.Z = peakColor.Z*0.5 + 0.5
			peakColor.W = 0.95

			peakColorU32 := imgui.ColorU32Vec4(peakColor)
			dl.AddRectFilledV(
				imgui.Vec2{X: barX, Y: peakY},
				imgui.Vec2{X: barX + actualBarWidth, Y: peakY + peakHeight},
				peakColorU32,
				0.5,
				imgui.DrawFlagsNone,
			)
		}
	}
}

// SetWidth sets the width of the spectrum analyzer
func (sa *SpectrumAnalyzerComponent) SetWidth(width float32) *SpectrumAnalyzerComponent {
	sa.Component.SetWidth(width)
	return sa
}

// SetHeight sets the height of the spectrum analyzer
func (sa *SpectrumAnalyzerComponent) SetHeight(height float32) *SpectrumAnalyzerComponent {
	sa.Component.SetHeight(height)
	return sa
}

// SetHeight sets the height of the spectrum analyzer
func (sa *SpectrumAnalyzerComponent) SetPadding(padding float32) *SpectrumAnalyzerComponent {
	sa.Component.SetPadding(padding)
	return sa
}

// SetNumBars sets the number of frequency bars to display
func (sa *SpectrumAnalyzerComponent) SetNumBars(numBars int) *SpectrumAnalyzerComponent {
	sa.numBars = numBars
	// Reset bar arrays to trigger reinitialization
	sa.barValues = nil
	sa.barTargets = nil
	sa.peakValues = nil
	sa.peakHoldTimes = nil
	return sa
}

// SetAttackTime sets the attack animation time in milliseconds
func (sa *SpectrumAnalyzerComponent) SetAttackTime(ms int64) *SpectrumAnalyzerComponent {
	sa.attackTime = ms
	return sa
}

// SetDecayTime sets the decay animation time in milliseconds
func (sa *SpectrumAnalyzerComponent) SetDecayTime(ms int64) *SpectrumAnalyzerComponent {
	sa.decayTime = ms
	return sa
}

// SetPeakHoldTime sets how long to hold peak indicators in milliseconds
func (sa *SpectrumAnalyzerComponent) SetPeakHoldTime(ms int64) *SpectrumAnalyzerComponent {
	sa.peakHoldTime = ms
	return sa
}

// SetPeakFallSpeed sets the speed at which peaks fall (units per frame)
func (sa *SpectrumAnalyzerComponent) SetPeakFallSpeed(speed float32) *SpectrumAnalyzerComponent {
	sa.peakFallSpeed = speed
	return sa
}

// SetNoiseGate sets the noise gate threshold (0.0 to 1.0)
func (sa *SpectrumAnalyzerComponent) SetNoiseGate(threshold float32) *SpectrumAnalyzerComponent {
	sa.noiseGate = threshold
	return sa
}

// SetColorMode sets the color mode for the spectrum bars
func (sa *SpectrumAnalyzerComponent) SetColorMode(mode SpectrumColorMode) *SpectrumAnalyzerComponent {
	sa.colorMode = mode
	return sa
}

// SetStaticColorIdx sets the colormap index for static color mode (0-255)
func (sa *SpectrumAnalyzerComponent) SetStaticColorIdx(idx int32) *SpectrumAnalyzerComponent {
	sa.staticColorIdx = idx
	return sa
}

func (sa *SpectrumAnalyzerComponent) ProcessUpdates() {
	sa.Component.ProcessUpdates()
}

func (sa *SpectrumAnalyzerComponent) Layout() {
	// Check if audio is playing
	isPlaying := sa.audioManager.IsPlaying()

	// Try to read latest FFT data
	select {
	case data := <-sa.audioManager.AnalyzerData:
		sa.magnitudes = data
	default:
		// No new data, use previous magnitudes
	}

	if sa.barValues == nil || len(sa.barValues) != sa.numBars {
		sa.barValues = make([]float32, sa.numBars)
		sa.barTargets = make([]float32, sa.numBars)
		sa.peakValues = make([]float32, sa.numBars)
		sa.peakHoldTimes = make([]int64, sa.numBars)
	}

	if !isPlaying {
		// Audio stopped - animate bars down to zero
		for i := 0; i < sa.numBars; i++ {
			if sa.barValues[i] > 0 {
				// Fast decay to zero
				// ms per frame at 60fps
				const frameTime = 16.67
				decayFactor := frameTime / float32(sa.decayTime)
				if decayFactor > 1.0 {
					decayFactor = 1.0
				}
				sa.barValues[i] -= sa.barValues[i] * decayFactor
				if sa.barValues[i] < 0.001 {
					sa.barValues[i] = 0
				}
			}
			if sa.peakValues[i] > 0 {
				// Fast peak decay when stopped
				sa.peakValues[i] -= sa.peakValues[i] * 0.3
				if sa.peakValues[i] < 0.001 {
					sa.peakValues[i] = 0
				}
			}
		}
	} else if len(sa.magnitudes) > 0 {
		sa.updateBars()
	}

	sa.renderSpectrum()
}

func getEasingFunc(name string) animation.EasingFunc {
	switch name {
	case "EaseLinear":
		return animation.EaseLinear
	case "EaseInQuad":
		return animation.EaseInQuad
	case "EaseOutQuad":
		return animation.EaseOutQuad
	case "EaseInOutQuad":
		return animation.EaseInOutQuad
	case "EaseInCubic":
		return animation.EaseInCubic
	case "EaseOutCubic":
		return animation.EaseOutCubic
	case "EaseInOutCubic":
		return animation.EaseInOutCubic
	case "EaseInExpo":
		return animation.EaseInExpo
	case "EaseOutExpo":
		return animation.EaseOutExpo
	case "EaseInOutExpo":
		return animation.EaseInOutExpo
	case "EaseInSine":
		return animation.EaseInSine
	case "EaseOutSine":
		return animation.EaseOutSine
	case "EaseInOutSine":
		return animation.EaseInOutSine
	default:
		return animation.EaseLinear
	}
}

func getColorMode(mode string) SpectrumColorMode {
	switch mode {
	case "height":
		return ColorModeHeight
	case "gradient":
		return ColorModeGradient
	case "static":
		return ColorModeStatic
	default:
		return ColorModeHeight
	}
}
