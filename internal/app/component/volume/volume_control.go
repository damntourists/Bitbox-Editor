package volume

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/logging"
	"strconv"

	"github.com/AllenDang/cimgui-go/imgui"
	"go.uber.org/zap"
)

var log = logging.NewLogger("volume")

// VolumeControlComponent provides a slider to control audio playback volume
type VolumeControlComponent struct {
	*component.Component[*VolumeControlComponent]

	onVolumeChange func(volume float32)
	onMuteToggle   func(muted bool)

	showPercent  bool
	handleRadius float32
}

// NewVolumeControl creates a new volume control component with default settings
func NewVolumeControl() *VolumeControlComponent {
	return NewVolumeControlWithID(imgui.IDStr("##volume_control"))
}

// NewVolumeControlWithID creates a new volume control with a specific ID
func NewVolumeControlWithID(id imgui.ID) *VolumeControlComponent {
	t := theme.GetCurrentTheme()

	vc := &VolumeControlComponent{
		showPercent:  false,
		handleRadius: 4,
	}

	vc.Component = component.NewComponent[*VolumeControlComponent](id, vc.HandleUpdate)
	vc.Component.SetLayoutBuilder(vc)

	vc.SetVolume(0.8)
	vc.SetMuted(false)
	vc.SetWidth(200)
	vc.SetHeight(16)
	vc.SetRounding(4)

	vc.SetSliderColor(t.Style.Colors.ButtonActive.Vec4)
	vc.SetHandleColor(t.Style.Colors.Text.Vec4)
	vc.SetTrackColor(t.Style.Colors.FrameBg.Vec4)

	return vc
}

// SetVolume sets the volume level (0.0 to 1.0)
func (vc *VolumeControlComponent) SetVolume(volume float32) *VolumeControlComponent {
	vc.Component.SetProgress(volume)
	return vc
}

// SetMuted sets the muted state
func (vc *VolumeControlComponent) SetMuted(muted bool) *VolumeControlComponent {
	vc.Component.SetEnabled(!muted)
	return vc
}

// SetWidth sets the width of the volume control
func (vc *VolumeControlComponent) SetWidth(width float32) *VolumeControlComponent {
	vc.Component.SetWidth(width)
	return vc
}

// SetHeight sets the height of the slider track
func (vc *VolumeControlComponent) SetHeight(height float32) *VolumeControlComponent {
	vc.Component.SetHeight(height)
	return vc
}

// SetSliderColor sets the color of the filled portion of the slider
func (vc *VolumeControlComponent) SetSliderColor(color imgui.Vec4) *VolumeControlComponent {
	vc.Component.SetProgressFgColor(color)
	return vc
}

// SetHandleColor sets the color of the slider handle (thumb)
func (vc *VolumeControlComponent) SetHandleColor(color imgui.Vec4) *VolumeControlComponent {
	vc.Component.SetTextColor(color)
	return vc
}

// SetTrackColor sets the color of the slider track (background)
func (vc *VolumeControlComponent) SetTrackColor(color imgui.Vec4) *VolumeControlComponent {
	vc.Component.SetProgressBgColor(color)
	return vc
}

// SetOnVolumeChange sets a callback that's called when the volume changes
func (vc *VolumeControlComponent) SetOnVolumeChange(callback func(volume float32)) *VolumeControlComponent {
	vc.onVolumeChange = callback
	return vc
}

// SetOnMuteToggle sets a callback that's called when mute is toggled
func (vc *VolumeControlComponent) SetOnMuteToggle(callback func(muted bool)) *VolumeControlComponent {
	vc.onMuteToggle = callback
	return vc
}

// SetRadius sets the radius of the slider handle
func (vc *VolumeControlComponent) SetRadius(radius float32) *VolumeControlComponent {
	vc.handleRadius = radius
	return vc
}

// GetVolume returns the current volume level
func (vc *VolumeControlComponent) GetVolume() float32 {
	return vc.Component.Progress()
}

// IsMuted returns whether the volume is muted
func (vc *VolumeControlComponent) IsMuted() bool {
	return !vc.Component.Enabled()
}

// ToggleMute toggles the muted state
func (vc *VolumeControlComponent) ToggleMute() {
	muted := vc.IsMuted()
	newMuted := !muted
	vc.SetMuted(newMuted)

	if vc.onMuteToggle != nil {
		vc.onMuteToggle(newMuted)
	}
}

func (vc *VolumeControlComponent) HandleUpdate(cmd component.UpdateCmd) {
	if vc.Component.HandleGlobalUpdate(cmd) {
		return
	}

	log.Warn("VolumeControlComponent unhandled update", zap.String("id", vc.IDStr()), zap.Any("cmd", cmd))
}

func (vc *VolumeControlComponent) Layout() {
	vc.Component.ProcessUpdates()

	t := theme.GetCurrentTheme()

	volume := vc.Component.Progress()
	muted := !vc.Component.Enabled()
	width := vc.Component.Width()
	height := vc.Component.Height()
	sliderColor := vc.Component.ProgressFg()
	handleColor := vc.Component.TextColor()
	trackColor := vc.Component.ProgressBg()
	rounding := vc.Component.Rounding()
	handleRadius := vc.handleRadius

	// Get the cursor position where the control will be placed
	pos := imgui.CursorScreenPos()
	sliderSize := imgui.Vec2{X: width, Y: height}

	buttonID := "##volume_slider_" + strconv.Itoa(int(vc.ID()))
	clicked := imgui.InvisibleButtonV(buttonID, sliderSize, imgui.ButtonFlagsNone)

	isHovered := imgui.IsItemHovered()
	isActive := imgui.IsItemActive()

	if isActive && imgui.IsMouseDraggingV(imgui.MouseButtonLeft, 0) {
		mousePos := imgui.MousePos()
		relativeX := mousePos.X - pos.X
		newVolume := relativeX / width

		if newVolume < 0 {
			newVolume = 0
		} else if newVolume > 1 {
			newVolume = 1
		}

		if newVolume != volume {
			vc.Component.SetProgress(newVolume)

			if vc.onVolumeChange != nil {
				vc.onVolumeChange(newVolume)
			}
		}
	}

	// Handle click to set volume at position
	if clicked {
		mousePos := imgui.MousePos()
		relativeX := mousePos.X - pos.X
		newVolume := relativeX / width

		if newVolume < 0 {
			newVolume = 0
		} else if newVolume > 1 {
			newVolume = 1
		}

		if newVolume != volume {
			vc.Component.SetProgress(newVolume)

			if vc.onVolumeChange != nil {
				vc.onVolumeChange(newVolume)
			}
		}
	}

	dl := imgui.WindowDrawList()
	maxBounds := imgui.Vec2{X: pos.X + width, Y: pos.Y + height}

	// Draw track (background)
	currentTrackColor := trackColor
	if muted {
		// Dim the track when muted
		currentTrackColor = imgui.Vec4{
			X: trackColor.X * 0.5,
			Y: trackColor.Y * 0.5,
			Z: trackColor.Z * 0.5,
			W: trackColor.W,
		}
	}

	dl.AddRectFilledV(
		pos,
		maxBounds,
		imgui.ColorU32Vec4(currentTrackColor),
		rounding,
		imgui.DrawFlagsNone,
	)

	// Draw filled portion (slider color)
	if volume > 0 && !muted {
		filledMax := imgui.Vec2{
			X: pos.X + (width * volume),
			Y: pos.Y + height,
		}

		currentSliderColor := sliderColor
		if isActive {
			currentSliderColor = imgui.Vec4{
				X: sliderColor.X * 1.2,
				Y: sliderColor.Y * 1.2,
				Z: sliderColor.Z * 1.2,
				W: sliderColor.W,
			}
		} else if isHovered {
			currentSliderColor = imgui.Vec4{
				X: sliderColor.X * 1.1,
				Y: sliderColor.Y * 1.1,
				Z: sliderColor.Z * 1.1,
				W: sliderColor.W,
			}
		}

		dl.AddRectFilledV(
			pos,
			filledMax,
			imgui.ColorU32Vec4(currentSliderColor),
			rounding,
			imgui.DrawFlagsNone,
		)
	}

	// Draw handle at current volume position
	handleX := pos.X + (width * volume)
	handleY := pos.Y + (height * 0.5)
	handleCenter := imgui.Vec2{X: handleX, Y: handleY}

	currentHandleColor := handleColor
	if muted {
		currentHandleColor = imgui.Vec4{
			X: handleColor.X * 0.5,
			Y: handleColor.Y * 0.5,
			Z: handleColor.Z * 0.5,
			W: handleColor.W,
		}
	} else if isActive {
		currentHandleColor = imgui.Vec4{
			X: handleColor.X * 1.3,
			Y: handleColor.Y * 1.3,
			Z: handleColor.Z * 1.3,
			W: handleColor.W,
		}
	} else if isHovered {
		currentHandleColor = imgui.Vec4{
			X: handleColor.X * 1.15,
			Y: handleColor.Y * 1.15,
			Z: handleColor.Z * 1.15,
			W: handleColor.W,
		}
	}

	dl.AddCircleFilledV(
		handleCenter,
		handleRadius,
		imgui.ColorU32Vec4(currentHandleColor),
		16,
	)

	// Draw border around handle
	borderColor := t.Style.Colors.Border.Vec4
	dl.AddCircleV(
		handleCenter,
		handleRadius,
		imgui.ColorU32Vec4(borderColor),
		16,
		1.5,
	)

	// Show hand cursor when hovering
	if isHovered {
		imgui.SetMouseCursor(imgui.MouseCursorHand)
	}
}

// Destroy cleans up the component
func (vc *VolumeControlComponent) Destroy() {
	vc.Component.Destroy()
}
