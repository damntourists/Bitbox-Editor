package audio

import (
	"math"
)

type PeakPosition struct {
	Sample    int
	Amplitude float32
}

// DetectPeaksFromDownsample analyzes downsampled data and detects peaks
func DetectPeaksFromDownsample(downsamples []Downsample, threshold float32, minPeakDistance int) []float64 {
	if len(downsamples) == 0 || threshold <= 0 || threshold > 1.0 {
		return nil
	}

	if len(downsamples[0].Maxs) == 0 {
		return nil
	}

	maxs := downsamples[0].Maxs
	mins := downsamples[0].Mins

	// Find the maximum absolute amplitude
	maxAmplitude := float32(0)
	for i := 0; i < len(maxs); i++ {
		absMax := float32(math.Abs(float64(maxs[i])))
		absMin := float32(math.Abs(float64(mins[i])))
		if absMax > maxAmplitude {
			maxAmplitude = absMax
		}
		if absMin > maxAmplitude {
			maxAmplitude = absMin
		}
	}

	if maxAmplitude == 0 {
		return nil
	}

	// Calculate the absolute threshold value
	thresholdValue := maxAmplitude * threshold

	peaks := make([]float64, 0)
	lastPeakIndex := -minPeakDistance
	for i := 0; i < len(maxs); i++ {
		absMax := float32(math.Abs(float64(maxs[i])))
		absMin := float32(math.Abs(float64(mins[i])))
		binAmplitude := absMax
		if absMin > absMax {
			binAmplitude = absMin
		}

		// Check if this bin is above the threshold and far enough from the last peak
		if binAmplitude >= thresholdValue && (i-lastPeakIndex) >= minPeakDistance {
			isPeak := true
			windowSize := minPeakDistance / 4
			if windowSize < 1 {
				windowSize = 1
			}

			// Check preceding bins
			for j := i - windowSize; j < i && j >= 0; j++ {
				jAbsMax := float32(math.Abs(float64(maxs[j])))
				jAbsMin := float32(math.Abs(float64(mins[j])))
				jAmplitude := jAbsMax
				if jAbsMin > jAbsMax {
					jAmplitude = jAbsMin
				}
				if jAmplitude > binAmplitude {
					isPeak = false
					break
				}
			}

			if isPeak {
				for j := i + 1; j <= i+windowSize && j < len(maxs); j++ {
					jAbsMax := float32(math.Abs(float64(maxs[j])))
					jAbsMin := float32(math.Abs(float64(mins[j])))
					jAmplitude := jAbsMax
					if jAbsMin > jAbsMax {
						jAmplitude = jAbsMin
					}
					if jAmplitude > binAmplitude {
						isPeak = false
						break
					}
				}
			}

			if isPeak {
				peaks = append(peaks, float64(i))
				lastPeakIndex = i
			}
		}
	}

	return peaks
}
