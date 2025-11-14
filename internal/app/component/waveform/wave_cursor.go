package waveform

/*
┏━━━━━━━━━━━━┓
┃ WaveCursor ┃
┗━━━━━━━━━━━━┛
*/

import (
	"bitbox-editor/internal/app/animation"
	"time"
)

type WaveCursor struct {
	position      float64
	hoverPosition float64
	isHovering    bool

	Active  bool
	Hovered bool
	Held    bool

	// Animation state for smooth cursor movement
	animating      bool
	animStartPos   float64
	animTargetPos  float64
	animStartTime  time.Time
	animDuration   time.Duration
	animEasingFunc animation.EasingFunc
}

func (wc *WaveCursor) SetHoverPosition(pos float64, hovering bool) {
	wc.hoverPosition = pos
	wc.isHovering = hovering
}

func (wc *WaveCursor) GetHoverPosition() (float64, bool) {
	return wc.hoverPosition, wc.isHovering
}

// AnimateToPosition starts an animation to smoothly move the cursor to the target position
func (wc *WaveCursor) AnimateToPosition(targetPos float64, duration time.Duration, easingFunc animation.EasingFunc) {
	if wc.position == targetPos {
		// Already at target
		return
	}

	wc.animating = true
	wc.animStartPos = wc.position
	wc.animTargetPos = targetPos
	wc.animStartTime = time.Now()
	wc.animDuration = duration
	if easingFunc == nil {
		easingFunc = animation.EaseOutCubic
	}
	wc.animEasingFunc = easingFunc
}

// UpdateAnimation updates the cursor position based on the current animation state
// Returns true if still animating, false if animation is complete
func (wc *WaveCursor) UpdateAnimation() bool {
	if !wc.animating {
		return false
	}

	elapsed := time.Since(wc.animStartTime)
	if elapsed >= wc.animDuration {
		// Animation complete
		wc.position = wc.animTargetPos
		wc.animating = false
		return false
	}

	// Calculate progress (0.0 to 1.0)
	t := float64(elapsed) / float64(wc.animDuration)

	// Apply easing function
	easedT := wc.animEasingFunc(t)

	// Lerp position
	wc.position = animation.LerpFloat64(wc.animStartPos, wc.animTargetPos, easedT)

	return true
}

// SetPositionImmediate sets the cursor position without animation
func (wc *WaveCursor) SetPositionImmediate(pos float64) {
	wc.position = pos
	wc.animating = false
}
