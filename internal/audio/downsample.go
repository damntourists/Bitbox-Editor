package audio

import "math"

type SampleGetter func(i int) float64

type Downsample struct {
	Mins, Maxs []float32
}

func downsampleMinMaxCore(lenY, bins int, get SampleGetter) (mins, maxs []float64) {
	if bins <= 0 || lenY == 0 {
		return nil, nil
	}
	if bins >= lenY {
		mins = make([]float64, lenY)
		maxs = make([]float64, lenY)
		for i := 0; i < lenY; i++ {
			v := get(i)
			mins[i], maxs[i] = v, v
		}
		return
	}
	mins = make([]float64, bins)
	maxs = make([]float64, bins)
	binSize := float64(lenY) / float64(bins)
	for i := 0; i < bins; i++ {
		start := int(math.Floor(float64(i) * binSize))
		end := int(math.Floor(float64(i+1) * binSize))
		if end <= start {
			end = start + 1
			if end > lenY {
				end = lenY
			}
		}
		minv, maxv := get(start), get(start)
		for j := start + 1; j < end; j++ {
			v := get(j)
			if v < minv {
				minv = v
			}
			if v > maxv {
				maxv = v
			}
		}
		mins[i], maxs[i] = minv, maxv
	}
	return
}

// DownsampleMinMax downsamples float32 samples into min/max bins
func DownsampleMinMax(y []float32, bins int) (mins, maxs []float32) {
	fmins, fmaxs := downsampleMinMaxCore(len(y), bins, func(i int) float64 {
		return float64(y[i])
	})
	if fmins == nil {
		return nil, nil
	}
	mins = make([]float32, len(fmins))
	maxs = make([]float32, len(fmins))
	for i := range fmins {
		mins[i] = float32(fmins[i])
		maxs[i] = float32(fmaxs[i])
	}
	return
}
