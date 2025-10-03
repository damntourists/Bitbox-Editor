package audio

import (
	"bitbox-editor/lib/logging"
	"math"

	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	log = logging.NewLogger("audio")
}

func DownsampleMinMaxInt64(y []int64, bins int) (mins, maxs []int64) {
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
