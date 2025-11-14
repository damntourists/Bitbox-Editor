package component

import (
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
)

// TODO: Move this to https://github.com/damntourists/imspinner-go
// TODO: Move damped & ease functions to an animation package

func SpinnerBegin(label string, radius float32) (imgui.Vec2, imgui.Vec2, imgui.Vec2, bool) {
	style := imgui.CurrentStyle()
	size := imgui.Vec2{X: radius*2 + style.FramePadding().X*2, Y: radius*2 + style.FramePadding().Y*2}
	pos := imgui.CursorScreenPos()
	imgui.InvisibleButton(label, size)
	if !imgui.IsItemVisible() {
		return imgui.Vec2{}, imgui.Vec2{}, imgui.Vec2{}, false
	}

	center := imgui.Vec2{
		X: pos.X + radius + style.FramePadding().X,
		Y: pos.Y + radius + style.FramePadding().Y,
	}

	return pos, size, center, true
}

func dampedSpring(mass, stiffness, damping, t, a, b float32) float32 {
	omega := float32(math.Sqrt(float64(stiffness / mass)))
	alpha := damping / (2 * mass)
	expTerm := float32(math.Exp(float64(-alpha * t)))
	under := float32(1 - alpha*alpha)
	if under < 0 {
		under = 0
	}
	cosTerm := float32(math.Cos(float64(omega * float32(math.Sqrt(float64(under))) * t)))
	result := expTerm * cosTerm
	return result*a + b
}
func dampedGravity(limtime float32) float32 {
	var time float32 = 0
	var initialHeight float32 = 10
	height := initialHeight
	var velocity float32 = 0
	_ = velocity
	var prtime float32 = 0
	for height >= 0 {
		if prtime >= limtime {
			return height / 10
		}
		time += 0.01
		prtime += 0.01
		height = initialHeight - 0.5*9.81*time*time
		if height < 0 {
			initialHeight = 0
			time = 0
		}
	}
	return 0
}

func dampedTrifolium(limtime, a, b float32) float32 {
	return a*float32(math.Sin(float64(limtime))) - b*float32(math.Sin(float64(3*limtime)))
}

func dampedInOutElastic(t, amplitude, period float32) float32 {
	if t == 0 {
		return 0
	}
	t *= 2
	if t == 2 {
		return 1
	}
	var s float32
	if amplitude < 1 {
		amplitude = 1
		s = period / 4
	} else {
		s = period / (2 * math.Pi) * float32(math.Asin(float64(1/amplitude)))
	}
	if t < 1 {
		return -0.5 * (amplitude * float32(math.Pow(2, float64(10*(t-1)))) * float32(math.Sin(float64((t-1-s)*(2*math.Pi)/period))))
	}
	return amplitude*float32(math.Pow(2, float64(-10*(t-1))))*float32(math.Sin(float64((t-1-s)*(2*math.Pi)/period)))*0.5 + 1
}

func dampedInfinity(t, a float32) (float32, float32) {
	s := float32(math.Sin(float64(t)))
	c := float32(math.Cos(float64(t)))
	den := 1 + s*s
	return (a * c) / den, (a * s * c) / den
}

func easeInQuad(t float32) float32 { return t * t }

//func easeOutQuad(t float32) float32 { return t * (2 - t) }

func easeInOutQuadFloat32(t float32) float32 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

func easeOutCubic(t float32) float32 { ft := t - 1; return ft*ft*ft + 1 }

func easeInExpo(t float32) float32 {
	if t == 0 {
		return 0
	}
	return float32(math.Pow(2, float64(10*(t-1))))
}

func easeInOutExpoFloat32(t float32) float32 {
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	if t < 0.5 {
		return 0.5 * float32(math.Pow(2, float64(20*t-10)))
	}
	return 0.5 * (2 - float32(math.Pow(2, float64(-20*t+10))))
}

func easeInOutQuadP(p []float32) float32 {
	tr := float32(math.Max(float64(float32(math.Sin(float64(p[0])))-0.5), 0)) * (p[1] * 0.5)
	return easeInOutQuadFloat32(tr)
}

func easeInOutExpoP(p []float32) float32 {
	tr := float32(math.Max(float64(float32(math.Sin(float64(p[0])))-0.5), 0)) * (p[1] * 0.4)
	return easeInOutExpoFloat32(tr) * (p[1] * 0.3)
}

func easeSpringP(p []float32) float32 {
	// p = [time, period, a, b]
	return dampedSpring(1, 10, 1, float32(math.Sin(float64(float32(math.Mod(float64(p[0]), float64(p[1])))))), p[2], p[3])
}

func easeGravityP(p []float32) float32 { return dampedGravity(p[0]) }

func easeInfinityP(p []float32) float32 { _, y := dampedInfinity(p[0], p[1]); return y }

func easeElasticP(p []float32) float32 { return dampedInOutElastic(p[1], p[2], p[3]) }

func easeSineP(p []float32) float32 { return 0.5 * (1 - float32(math.Cos(float64(p[0]*math.Pi)))) }

func easeDampingP(p []float32) float32 {
	const A float32 = 3.14 * 2
	const ma float32 = 5.0
	const k float32 = 2.1
	const b float32 = 0.09
	const theta float32 = 0.0
	w := float32(math.Sqrt(float64(k / ma)))
	t := float32(math.Mod(float64(p[0]), 25))
	x := A * float32(math.Exp(float64(-b*t))) * float32(math.Cos(float64(w*t-theta)))
	return x
}

type SpinnerEaseMode int

const (
	SpinnerEaseNone SpinnerEaseMode = iota
	SpinnerEaseInOutQuad
	SpinnerEaseInOutExpo
	SpinnerEaseSpring
	SpinnerEaseGravity
	SpinnerEaseInfinity
	SpinnerEaseElastic
	SpinnerEaseSine
	SpinnerEaseDamping
)

func SpinnerEase(mode SpinnerEaseMode, params ...float32) float32 {
	switch mode {
	case SpinnerEaseInOutQuad:
		return easeInOutQuadP(params)
	case SpinnerEaseInOutExpo:
		return easeInOutExpoP(params)
	case SpinnerEaseSpring:
		return easeSpringP(params)
	case SpinnerEaseGravity:
		return easeGravityP(params)
	case SpinnerEaseInfinity:
		return easeInfinityP(params)
	case SpinnerEaseElastic:
		return easeElasticP(params)
	case SpinnerEaseSine:
		return easeSineP(params)
	case SpinnerEaseDamping:
		return easeDampingP(params)
	case SpinnerEaseNone:
		return 0
	}
	return 0
}

// Spinner draws a circular animated spinner.
func Spinner(label string, radius, thickness float32, color uint32) bool {
	_, size, center, ok := SpinnerBegin(label, radius)

	if !ok {
		return false
	}

	imgui.InvisibleButton(label, size)
	if !imgui.IsItemVisible() {
		return false
	}

	// Draw
	dl := imgui.WindowDrawList()
	dl.PathClear()

	numSegments := 15
	t := imgui.Time()
	start := math.Abs(math.Abs(math.Sin(t)) * float64(numSegments-5))

	aMin := float32(math.Pi * 2.0 * float64(start) / float64(numSegments))
	aMax := float32(math.Pi * 2.0 * float64(numSegments-3) / float64(numSegments))

	for i := 0; i < numSegments; i++ {
		a := aMin + (float32(i)/float32(numSegments))*(aMax-aMin)
		x := center.X + float32(math.Cos(float64(a)+t*8.0))*radius
		y := center.Y + float32(math.Sin(float64(a)+t*8.0))*radius
		dl.PathLineTo(imgui.Vec2{X: x, Y: y})
	}

	dl.PathStrokeV(color, imgui.DrawFlagsNone, thickness)
	return true
}

func SpinnerDnaDots(label string, radius, thickness float32, color uint32, speed float32, dot_segments int32, delta float32, mode bool) {
	_, size, centre, ok := SpinnerBegin(label, radius)

	if !ok {
		return
	}

	const nextItemCoeff float32 = 2.5
	dots := int(size.X / (thickness * nextItemCoeff))
	if dots <= 0 {
		return
	}

	start := float32(math.Mod(imgui.Time()*float64(speed), 2*math.Pi))
	base := imgui.ColorConvertU32ToFloat4(color)

	var h, s, v float32
	imgui.ColorConvertRGBtoHSV(base.X, base.Y, base.Z, &h, &s, &v)

	dl := imgui.WindowDrawList()

	drawPoint := func(angle float32, i int) imgui.Vec2 {
		a := angle + start + float32(math.Pi-float64(i)*math.Pi/float64(dots))
		thK := 1.0 + float32(math.Sin(float64(a)+math.Pi*0.5))*0.5

		var pp float32
		if mode {
			pp = centre.X + float32(math.Sin(float64(a)))*size.X*delta
		} else {
			pp = centre.Y + float32(math.Sin(float64(a)))*size.Y*delta
		}

		hue := h + float32(i)*(1.0/float32(dots)*2.0)

		var h, s, v float32
		imgui.ColorConvertHSVtoRGB(hue, s, v, &h, &s, &v)

		dotCol := imgui.NewColor(1, 1, 1, 1).Pack()

		var p imgui.Vec2
		if mode {
			p = imgui.Vec2{
				X: pp,
				Y: centre.Y - (size.Y * 0.5) + float32(i)*thickness*nextItemCoeff,
			}
		} else {
			p = imgui.Vec2{
				X: centre.X - (size.X * 0.5) + float32(i)*thickness*nextItemCoeff,
				Y: pp,
			}
		}

		dl.AddCircleFilledV(p, thickness*thK, dotCol, dot_segments)
		return p
	}

	for i := 0; i < dots; i++ {
		p1 := drawPoint(0, i)
		p2 := drawPoint(float32(math.Pi), i)
		dl.AddLineV(p1, p2, color, thickness*0.5)
	}
}

func SpinnerFluidPoints(label string, radius, thickness float32, color uint32, speed float32, dots int, delta float32) {
	pos, size, centre, ok := SpinnerBegin(label, radius)
	if !ok {
		return
	}

	const r0 float32 = 0.033
	const r2 float32 = 0.8

	numSegments := 15

	jk := radius * 2.0 / float32(numSegments)

	base := imgui.ColorConvertU32ToFloat4(color)

	var h, s, v float32
	imgui.ColorConvertRGBtoHSV(base.X, base.Y, base.Z, &h, &s, &v)

	if dots <= 0 {
		dots = 6
	}

	dl := imgui.WindowDrawList()

	t := float32(imgui.Time())
	for j := 0; j < numSegments; j++ {
		hcol := (0.6 + delta*float32(math.Sin(float64(t*(speed*r2*2.0)+2.0*r0*float32(j)*jk)))) * (radius * 2.0 * r2)

		for i := 0; i < dots; i++ {
			hue := h - float32(i)*0.1

			var r, g, b float32
			imgui.ColorConvertHSVtoRGB(hue, s, v, &r, &g, &b)

			dotCol := imgui.ColorConvertFloat4ToU32(imgui.Vec4{X: r, Y: g, Z: b, W: 1})

			x := pos.X + imgui.CurrentStyle().FramePadding().X + float32(j)*jk
			y := centre.Y + size.Y/2.0 - (hcol/float32(dots))*float32(i)

			dl.AddCircleFilledV(imgui.Vec2{X: x, Y: y}, thickness, dotCol, 0)
		}
	}
}

func SpinnerHerbertBalls3D(label string, radius, thickness float32, color uint32, speed float32) {
	_, _, centre, ok := SpinnerBegin(label, radius)
	if !ok {
		return
	}
	dl := imgui.WindowDrawList()

	rstart := float32(math.Mod(imgui.Time()*float64(speed), 2*math.Pi))
	radius1 := 0.3 * radius
	radius2 := 0.8 * radius
	balls := 2
	angleOffset := float32(2 * math.Pi / float64(balls))

	var frontpos, backpos imgui.Vec2
	for i := 0; i < balls; i++ {
		a := rstart + float32(i)*angleOffset
		t := thickness
		if i == 1 {
			t = 0.7 * thickness
		}
		p := imgui.Vec2{
			X: centre.X + float32(math.Cos(float64(a)))*radius1,
			Y: centre.Y + float32(math.Sin(float64(a)))*radius1,
		}
		dl.AddCircleFilledV(p, t, color, 0)
		if i == 0 {
			frontpos = p
		} else {
			backpos = p
		}
	}

	var lastpos imgui.Vec2
	steps := balls * 2
	for i := 0; i <= steps; i++ {
		a := -rstart + float32(i)*(angleOffset/2.0)
		p := imgui.Vec2{
			X: centre.X + float32(math.Cos(float64(a)))*radius2,
			Y: centre.Y + float32(math.Sin(float64(a)))*radius2,
		}
		dx := p.X - frontpos.X
		dy := p.Y - frontpos.Y
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))
		t := (dist / (radius * 1.0)) * thickness

		dl.AddCircleFilledV(p, t, color, 0)

		lineT := thickness / 2.0
		if lineT < 1.0 {
			lineT = 1.0
		}

		stColor := imgui.ColorConvertU32ToFloat4(color)
		stColor.W = 0.5
		stColorFinal := imgui.NewColor(stColor.X, stColor.Y, stColor.Z, stColor.W).Pack()

		dl.AddLineV(p, backpos, stColorFinal, lineT)

		if i > 0 {
			dl.AddLineV(p, lastpos, color, lineT)
		}

		dl.AddLineV(p, frontpos, color, lineT)

		lastpos = p
	}
}

func SpinnerBarChartAdvSineFade(label string, radius, thickness float32, color uint32, speed float32) {
	_, _, centre, ok := SpinnerBegin(label, radius)
	if !ok {
		return
	}

	dl := imgui.WindowDrawList()
	start := float32(imgui.Time()) * speed

	bars := int((radius * 2.0) / thickness)
	if bars <= 0 {
		return
	}

	baseX := centre.X - radius

	offset := float32((math.Pi / 2.0) / float64(bars))
	for i := 0; i < bars; i++ {
		a := start - float32(i)*offset
		halfsy := float32(math.Max(0.1, math.Cos(float64(a))+1.0)) * radius * 0.5

		x := baseX + float32(i)*thickness
		min := imgui.Vec2{X: x - thickness/2.0, Y: centre.Y + halfsy}
		max := imgui.Vec2{X: x + thickness/2.0, Y: centre.Y - halfsy}

		alpha := float32(math.Max(0.1, float64(halfsy/radius)))

		newColV4 := imgui.ColorConvertU32ToFloat4(color)
		newColV4.W = alpha
		newCol := imgui.NewColor(newColV4.X, newColV4.Y, newColV4.Z, newColV4.W).Pack()

		dl.AddRectFilledV(min, max, newCol, 0, 0)
	}
}

func SpinnerTwinAng180(label string, radius1, radius2, thickness float32, color1, color2 uint32, speed, angle float32, mode int) {
	radius := radius1
	if radius2 > radius {
		radius = radius2
	}
	_, _, centre, ok := SpinnerBegin(label, radius)
	if !ok {
		return
	}
	dl := imgui.WindowDrawList()

	numSegments := 15
	numSegments *= 8

	start := float32(math.Mod(imgui.Time()*float64(speed), 2*math.Pi))
	aoffset := float32(math.Mod(imgui.Time(), 2*math.Pi))
	bofsset := aoffset
	if aoffset > math.Pi {
		bofsset = math.Pi
	}
	angleOffset := float32((2 * math.Pi) / float64(numSegments))

	var aredMin float32
	if aoffset > math.Pi {
		aredMin = aoffset - math.Pi
	} else {
		aredMin = 0
	}

	radiusMode := func(a, r, f float32) float32 {
		switch mode {
		case 2:
			// damped_trifolium(a, 0, f) * r
			return dampedTrifolium(a, 0, f) * r
		default:
			return r
		}
	}

	var ared float32
	dl.PathClear()
	for i := 0; i <= numSegments/2+1; i++ {
		ared = start + float32(i)*angleOffset
		switch mode {
		case 1:
			_, y := dampedInfinity(start, angle)
			ared += y
		}

		iAngle := float32(i) * angleOffset
		if iAngle < aredMin {
			continue
		}
		if iAngle > bofsset {
			break
		}

		x := centre.X + float32(math.Cos(float64(ared)))*radiusMode(ared, radius2, -1.1)
		y := centre.Y + float32(math.Sin(float64(ared)))*radiusMode(ared, radius2, -1.1)
		dl.PathLineTo(imgui.Vec2{X: x, Y: y})
	}
	dl.PathStrokeV(color2, imgui.DrawFlagsNone, thickness)

	dl.PathClear()
	for i := 0; i <= 2*numSegments+1; i++ {
		a := ared + aredMin + float32(i)*angleOffset

		iAngle := float32(i) * angleOffset
		if iAngle < aredMin {
			continue
		}
		if iAngle > bofsset {
			break
		}

		x := centre.X + float32(math.Cos(float64(a)))*radiusMode(a, radius1, 1.0)
		y := centre.Y + float32(math.Sin(float64(a)))*radiusMode(a, radius1, 1.0)
		dl.PathLineTo(imgui.Vec2{X: x, Y: y})
	}
	dl.PathStrokeV(color1, imgui.DrawFlagsNone, thickness)
}
