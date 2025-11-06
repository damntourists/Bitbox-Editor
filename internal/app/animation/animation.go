package animation

/*
╭───────────╮
│ Animation │
╰───────────╯
*/
import (
	"math"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
)

const DefaultColorFadeDuration = 350 * time.Millisecond

// EasingFunc is a function that takes a normalized time value (0.0 to 1.0)
// and returns an eased value (typically 0.0 to 1.0)
type EasingFunc func(t float64) float64

func EaseLinear(t float64) float64 {
	return t
}

func EaseInQuad(t float64) float64 {
	return t * t
}

func EaseOutQuad(t float64) float64 {
	return 1 - (1-t)*(1-t)
}

func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

func EaseInCubic(t float64) float64 {
	return t * t * t
}

func EaseOutCubic(t float64) float64 {
	ft := t - 1
	return ft*ft*ft + 1
}

func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	ft := 2*t - 2
	return 1 + ft*ft*ft/2
}

func EaseInExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	return math.Pow(2, 10*(t-1))
}

func EaseOutExpo(t float64) float64 {
	if t == 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

func EaseInOutExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	if t < 0.5 {
		return 0.5 * math.Pow(2, 20*t-10)
	}
	return 0.5 * (2 - math.Pow(2, -20*t+10))
}

func EaseInSine(t float64) float64 {
	return 1 - math.Cos(t*math.Pi/2)
}

func EaseOutSine(t float64) float64 {
	return math.Sin(t * math.Pi / 2)
}

func EaseInOutSine(t float64) float64 {
	return 0.5 * (1 - math.Cos(math.Pi*t))
}

func LerpVec4(a, b imgui.Vec4, t float32) imgui.Vec4 {
	return imgui.Vec4{
		X: a.X + (b.X-a.X)*t,
		Y: a.Y + (b.Y-a.Y)*t,
		Z: a.Z + (b.Z-a.Z)*t,
		W: a.W + (b.W-a.W)*t,
	}
}

func LerpFloat64(a, b, t float64) float64 {
	return a + (b-a)*t
}
