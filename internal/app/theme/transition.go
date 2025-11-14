package theme

import (
	"bitbox-editor/internal/app/animation"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
)

var (
	transitionState = &ThemeTransition{
		isTransitioning: atomic.Bool{},
		easingFunc:      animation.EaseInOutCubic,
	}
)

// ThemeTransition manages smooth transitions between themes
type ThemeTransition struct {
	mu              sync.RWMutex
	isTransitioning atomic.Bool
	fromTheme       *Theme
	toTheme         *Theme
	startTime       time.Time
	duration        time.Duration
	progress        float32
	easingFunc      animation.EasingFunc
	cancelChan      chan struct{}
}

// lerpFloat32 linearly interpolates between two float32 values
func lerpFloat32(a, b, t float32) float32 {
	return a + (b-a)*t
}

// lerpVec2 linearly interpolates between two Vec2 values (stored as []float32 slices)
func lerpVec2(a, b []float32, t float32) []float32 {
	if len(a) < 2 || len(b) < 2 {
		return a
	}
	return []float32{
		lerpFloat32(a[0], b[0], t),
		lerpFloat32(a[1], b[1], t),
	}
}

// lerpVec4 linearly interpolates between two Vec4 values (colors)
func lerpVec4(a, b imgui.Vec4, t float32) imgui.Vec4 {
	return imgui.Vec4{
		X: lerpFloat32(a.X, b.X, t),
		Y: lerpFloat32(a.Y, b.Y, t),
		Z: lerpFloat32(a.Z, b.Z, t),
		W: lerpFloat32(a.W, b.W, t),
	}
}

// lerpStyle interpolates between two Style structs
func lerpStyle(from, to Style, t float32, easingFunc animation.EasingFunc) Style {
	// Apply easing function
	easedT := float32(easingFunc(float64(t)))

	result := Style{
		Alpha:                     lerpFloat32(from.Alpha, to.Alpha, easedT),
		DisabledAlpha:             lerpFloat32(from.DisabledAlpha, to.DisabledAlpha, easedT),
		WindowPadding:             lerpVec2(from.WindowPadding, to.WindowPadding, easedT),
		WindowRounding:            lerpFloat32(from.WindowRounding, to.WindowRounding, easedT),
		WindowBorderSize:          lerpFloat32(from.WindowBorderSize, to.WindowBorderSize, easedT),
		WindowMinSize:             lerpVec2(from.WindowMinSize, to.WindowMinSize, easedT),
		WindowTitleAlign:          lerpVec2(from.WindowTitleAlign, to.WindowTitleAlign, easedT),
		WindowMenuButtonPosition:  to.WindowMenuButtonPosition, // Can't interpolate string
		ChildRounding:             lerpFloat32(from.ChildRounding, to.ChildRounding, easedT),
		ChildBorderSize:           lerpFloat32(from.ChildBorderSize, to.ChildBorderSize, easedT),
		PopupRounding:             lerpFloat32(from.PopupRounding, to.PopupRounding, easedT),
		PopupBorderSize:           lerpFloat32(from.PopupBorderSize, to.PopupBorderSize, easedT),
		FramePadding:              lerpVec2(from.FramePadding, to.FramePadding, easedT),
		FrameRounding:             lerpFloat32(from.FrameRounding, to.FrameRounding, easedT),
		FrameBorderSize:           lerpFloat32(from.FrameBorderSize, to.FrameBorderSize, easedT),
		ItemSpacing:               lerpVec2(from.ItemSpacing, to.ItemSpacing, easedT),
		ItemInnerSpacing:          lerpVec2(from.ItemInnerSpacing, to.ItemInnerSpacing, easedT),
		CellPadding:               lerpVec2(from.CellPadding, to.CellPadding, easedT),
		IndentSpacing:             lerpFloat32(from.IndentSpacing, to.IndentSpacing, easedT),
		ColumnsMinSpacing:         lerpFloat32(from.ColumnsMinSpacing, to.ColumnsMinSpacing, easedT),
		ScrollbarSize:             lerpFloat32(from.ScrollbarSize, to.ScrollbarSize, easedT),
		ScrollbarRounding:         lerpFloat32(from.ScrollbarRounding, to.ScrollbarRounding, easedT),
		GrabMinSize:               lerpFloat32(from.GrabMinSize, to.GrabMinSize, easedT),
		GrabRounding:              lerpFloat32(from.GrabRounding, to.GrabRounding, easedT),
		TabRounding:               lerpFloat32(from.TabRounding, to.TabRounding, easedT),
		TabBorderSize:             lerpFloat32(from.TabBorderSize, to.TabBorderSize, easedT),
		TabMinWidthForCloseButton: lerpFloat32(from.TabMinWidthForCloseButton, to.TabMinWidthForCloseButton, easedT),
		ColorButtonPosition:       to.ColorButtonPosition, // Can't interpolate string
		ButtonTextAlign:           lerpVec2(from.ButtonTextAlign, to.ButtonTextAlign, easedT),
		SelectableTextAlign:       lerpVec2(from.SelectableTextAlign, to.SelectableTextAlign, easedT),
		Colors:                    lerpStyleColors(from.Colors, to.Colors, easedT),
	}

	return result
}

// lerpStyleColors interpolates between two StyleColor structs
func lerpStyleColors(from, to StyleColor, t float32) StyleColor {

	return StyleColor{
		Text:                  rgba{lerpVec4(from.Text.Vec4, to.Text.Vec4, t)},
		TextDisabled:          rgba{lerpVec4(from.TextDisabled.Vec4, to.TextDisabled.Vec4, t)},
		WindowBg:              rgba{lerpVec4(from.WindowBg.Vec4, to.WindowBg.Vec4, t)},
		ChildBg:               rgba{lerpVec4(from.ChildBg.Vec4, to.ChildBg.Vec4, t)},
		PopupBg:               rgba{lerpVec4(from.PopupBg.Vec4, to.PopupBg.Vec4, t)},
		Border:                rgba{lerpVec4(from.Border.Vec4, to.Border.Vec4, t)},
		BorderShadow:          rgba{lerpVec4(from.BorderShadow.Vec4, to.BorderShadow.Vec4, t)},
		FrameBg:               rgba{lerpVec4(from.FrameBg.Vec4, to.FrameBg.Vec4, t)},
		FrameBgHovered:        rgba{lerpVec4(from.FrameBgHovered.Vec4, to.FrameBgHovered.Vec4, t)},
		FrameBgActive:         rgba{lerpVec4(from.FrameBgActive.Vec4, to.FrameBgActive.Vec4, t)},
		TitleBg:               rgba{lerpVec4(from.TitleBg.Vec4, to.TitleBg.Vec4, t)},
		TitleBgActive:         rgba{lerpVec4(from.TitleBgActive.Vec4, to.TitleBgActive.Vec4, t)},
		TitleBgCollapsed:      rgba{lerpVec4(from.TitleBgCollapsed.Vec4, to.TitleBgCollapsed.Vec4, t)},
		MenuBarBg:             rgba{lerpVec4(from.MenuBarBg.Vec4, to.MenuBarBg.Vec4, t)},
		ScrollbarBg:           rgba{lerpVec4(from.ScrollbarBg.Vec4, to.ScrollbarBg.Vec4, t)},
		ScrollbarGrab:         rgba{lerpVec4(from.ScrollbarGrab.Vec4, to.ScrollbarGrab.Vec4, t)},
		ScrollbarGrabHovered:  rgba{lerpVec4(from.ScrollbarGrabHovered.Vec4, to.ScrollbarGrabHovered.Vec4, t)},
		ScrollbarGrabActive:   rgba{lerpVec4(from.ScrollbarGrabActive.Vec4, to.ScrollbarGrabActive.Vec4, t)},
		CheckMark:             rgba{lerpVec4(from.CheckMark.Vec4, to.CheckMark.Vec4, t)},
		SliderGrab:            rgba{lerpVec4(from.SliderGrab.Vec4, to.SliderGrab.Vec4, t)},
		SliderGrabActive:      rgba{lerpVec4(from.SliderGrabActive.Vec4, to.SliderGrabActive.Vec4, t)},
		Button:                rgba{lerpVec4(from.Button.Vec4, to.Button.Vec4, t)},
		ButtonHovered:         rgba{lerpVec4(from.ButtonHovered.Vec4, to.ButtonHovered.Vec4, t)},
		ButtonActive:          rgba{lerpVec4(from.ButtonActive.Vec4, to.ButtonActive.Vec4, t)},
		Header:                rgba{lerpVec4(from.Header.Vec4, to.Header.Vec4, t)},
		HeaderHovered:         rgba{lerpVec4(from.HeaderHovered.Vec4, to.HeaderHovered.Vec4, t)},
		HeaderActive:          rgba{lerpVec4(from.HeaderActive.Vec4, to.HeaderActive.Vec4, t)},
		Separator:             rgba{lerpVec4(from.Separator.Vec4, to.Separator.Vec4, t)},
		SeparatorHovered:      rgba{lerpVec4(from.SeparatorHovered.Vec4, to.SeparatorHovered.Vec4, t)},
		SeparatorActive:       rgba{lerpVec4(from.SeparatorActive.Vec4, to.SeparatorActive.Vec4, t)},
		ResizeGrip:            rgba{lerpVec4(from.ResizeGrip.Vec4, to.ResizeGrip.Vec4, t)},
		ResizeGripHovered:     rgba{lerpVec4(from.ResizeGripHovered.Vec4, to.ResizeGripHovered.Vec4, t)},
		ResizeGripActive:      rgba{lerpVec4(from.ResizeGripActive.Vec4, to.ResizeGripActive.Vec4, t)},
		Tab:                   rgba{lerpVec4(from.Tab.Vec4, to.Tab.Vec4, t)},
		TabHovered:            rgba{lerpVec4(from.TabHovered.Vec4, to.TabHovered.Vec4, t)},
		TabActive:             rgba{lerpVec4(from.TabActive.Vec4, to.TabActive.Vec4, t)},
		TabUnfocused:          rgba{lerpVec4(from.TabUnfocused.Vec4, to.TabUnfocused.Vec4, t)},
		TabUnfocusedActive:    rgba{lerpVec4(from.TabUnfocusedActive.Vec4, to.TabUnfocusedActive.Vec4, t)},
		DockingPreview:        rgba{lerpVec4(from.DockingPreview.Vec4, to.DockingPreview.Vec4, t)},
		DockingEmptyBg:        rgba{lerpVec4(from.DockingEmptyBg.Vec4, to.DockingEmptyBg.Vec4, t)},
		PlotLines:             rgba{lerpVec4(from.PlotLines.Vec4, to.PlotLines.Vec4, t)},
		PlotLinesHovered:      rgba{lerpVec4(from.PlotLinesHovered.Vec4, to.PlotLinesHovered.Vec4, t)},
		PlotHistogram:         rgba{lerpVec4(from.PlotHistogram.Vec4, to.PlotHistogram.Vec4, t)},
		PlotHistogramHovered:  rgba{lerpVec4(from.PlotHistogramHovered.Vec4, to.PlotHistogramHovered.Vec4, t)},
		TableHeaderBg:         rgba{lerpVec4(from.TableHeaderBg.Vec4, to.TableHeaderBg.Vec4, t)},
		TableBorderStrong:     rgba{lerpVec4(from.TableBorderStrong.Vec4, to.TableBorderStrong.Vec4, t)},
		TableBorderLight:      rgba{lerpVec4(from.TableBorderLight.Vec4, to.TableBorderLight.Vec4, t)},
		TableRowBg:            rgba{lerpVec4(from.TableRowBg.Vec4, to.TableRowBg.Vec4, t)},
		TableRowBgAlt:         rgba{lerpVec4(from.TableRowBgAlt.Vec4, to.TableRowBgAlt.Vec4, t)},
		TextSelectedBg:        rgba{lerpVec4(from.TextSelectedBg.Vec4, to.TextSelectedBg.Vec4, t)},
		DragDropTarget:        rgba{lerpVec4(from.DragDropTarget.Vec4, to.DragDropTarget.Vec4, t)},
		NavHighlight:          rgba{lerpVec4(from.NavHighlight.Vec4, to.NavHighlight.Vec4, t)},
		NavWindowingHighlight: rgba{lerpVec4(from.NavWindowingHighlight.Vec4, to.NavWindowingHighlight.Vec4, t)},
		NavWindowingDimBg:     rgba{lerpVec4(from.NavWindowingDimBg.Vec4, to.NavWindowingDimBg.Vec4, t)},
		ModalWindowDimBg:      rgba{lerpVec4(from.ModalWindowDimBg.Vec4, to.ModalWindowDimBg.Vec4, t)},
	}
}

// TransitionToTheme starts a smooth transition from the current theme to a new theme
func TransitionToTheme(toTheme *Theme, durationMs int) {
	TransitionToThemeWithEasing(toTheme, durationMs, animation.EaseInOutCubic)
}

// TransitionToThemeWithEasing starts a smooth transition with a custom easing function
func TransitionToThemeWithEasing(toTheme *Theme, durationMs int, easingFunc animation.EasingFunc) {
	transitionState.mu.Lock()
	defer transitionState.mu.Unlock()

	// Cancel any existing transition
	if transitionState.isTransitioning.Load() {
		close(transitionState.cancelChan)
	}

	// Set up new transition
	transitionState.fromTheme = GetCurrentTheme()
	transitionState.toTheme = toTheme
	transitionState.startTime = time.Now()
	transitionState.duration = time.Duration(durationMs) * time.Millisecond
	transitionState.progress = 0
	transitionState.easingFunc = easingFunc
	transitionState.cancelChan = make(chan struct{})
	transitionState.isTransitioning.Store(true)

	// Start transition goroutine
	go runTransition()
}

// runTransition executes the theme transition animation
func runTransition() {
	// ~60fps
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-transitionState.cancelChan:
			return
		case <-ticker.C:
			transitionState.mu.Lock()
			elapsed := time.Since(transitionState.startTime)
			progress := float32(elapsed) / float32(transitionState.duration)

			if progress >= 1.0 {
				// Transition complete
				SetCurrentTheme(transitionState.toTheme)
				transitionState.isTransitioning.Store(false)
				transitionState.mu.Unlock()
				return
			}

			// Update progress
			transitionState.progress = progress

			// Create interpolated theme with easing function
			interpolatedTheme := &Theme{
				Name:        transitionState.toTheme.Name,
				Author:      transitionState.toTheme.Author,
				Description: transitionState.toTheme.Description,
				Tags:        transitionState.toTheme.Tags,
				Date:        transitionState.toTheme.Date,
				Style: lerpStyle(
					transitionState.fromTheme.Style,
					transitionState.toTheme.Style,
					progress, transitionState.easingFunc,
				),
			}

			// Apply interpolated theme
			SetCurrentTheme(interpolatedTheme)
			transitionState.mu.Unlock()
		}
	}
}

// IsTransitioning returns true if a theme transition is currently in progress
func IsTransitioning() bool {
	return transitionState.isTransitioning.Load()
}

// GetTransitionProgress returns the current progress of the transition (0.0 to 1.0)
func GetTransitionProgress() float32 {
	transitionState.mu.RLock()
	defer transitionState.mu.RUnlock()
	return transitionState.progress
}

// CancelTransition cancels any ongoing theme transition
func CancelTransition() {
	transitionState.mu.Lock()
	defer transitionState.mu.Unlock()

	if transitionState.isTransitioning.Load() {
		close(transitionState.cancelChan)
		SetCurrentTheme(transitionState.toTheme)
		transitionState.isTransitioning.Store(false)
	}
}
