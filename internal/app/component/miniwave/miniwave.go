package miniwave

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/audio"
	"bitbox-editor/internal/logging"
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
	"go.uber.org/zap"
)

var log = logging.NewLogger("miniwave")

// WaveformLine is a vertical line for rendering the waveform
type WaveformLine struct {
	X     float32
	YMin  float32
	YMax  float32
	Color uint32
}

// MiniWaveformComponent displays a small, non-interactive waveform
type MiniWaveformComponent struct {
	*component.Component[*MiniWaveformComponent]

	path string // The file path this component represents

	isReady              bool
	loadFailed           bool
	cacheValid           bool
	cachedLines          []WaveformLine
	lastCachedWidth      float32
	lastCachedColormap   implot.Colormap
	normalizeFactor      float32
	lastCacheRequestTime float64

	filteredEventSub *eventbus.FilteredSubscription
}

// NewMiniWaveform creates a new waveform component
func NewMiniWaveform(id imgui.ID, path string) *MiniWaveformComponent {
	cmp := &MiniWaveformComponent{
		path:                 path,
		isReady:              false,
		loadFailed:           false,
		cacheValid:           false,
		cachedLines:          nil,
		lastCachedWidth:      -1,
		lastCachedColormap:   theme.GetCurrentColormap(),
		normalizeFactor:      1.0,
		lastCacheRequestTime: 0.0,
	}

	cmp.Component = component.NewComponent[*MiniWaveformComponent](id, cmp.handleUpdate)
	cmp.Component.SetLayoutBuilder(cmp)

	// Subscribe to all load events using filtered subscription
	bus := eventbus.Bus
	uuid := cmp.UUID()
	cmp.filteredEventSub = eventbus.NewFilteredSubscription(uuid, 10)
	cmp.filteredEventSub.SubscribeMultiple(
		bus,
		events.AudioMetadataLoadedKey,
		events.AudioSamplesLoadedKey,
		events.AudioLoadFailedKey,
	)

	return cmp
}

// drainEvents reads from the event bus subscription channel and translates relevant events into local commands.
func (mw *MiniWaveformComponent) drainEvents() {
	if mw.filteredEventSub != nil {
		for {
			select {
			case event := <-mw.filteredEventSub.Events():
				// Check if it's an audio load event
				if e, ok := event.(events.AudioLoadEventRecord); ok {
					// Check if it's for the file this component cares about
					if e.Path == mw.path {
						mw.SendUpdate(component.UpdateCmd{Type: cmdUpdateStateFromCache})
					}
				}
			default:
				// No more events
				return
			}
		}
	}
}

// updateStateFromCache is called from handleUpdate
func (mw *MiniWaveformComponent) updateStateFromCache() {
	if mw.path == "" {
		return
	}

	cache := audio.GetGlobalAsyncCache()
	snapshot := cache.GetSnapshot(mw.path)

	if snapshot == nil {
		// No data yet. Request load if not already requested.
		cache.RequestLoad(mw.path, audio.LoadMiniDownsamples)
		return
	}

	// We have a snapshot - check if mini downsamples are available
	if len(snapshot.MiniDownsamples) > 0 && len(snapshot.MiniDownsamples[0].Mins) > 0 {
		mw.isReady = true
		mw.loadFailed = snapshot.LoadErr != nil
		// Mark cache as invalid so Layout() will rebuild it
		mw.cacheValid = false
	} else if snapshot.LoadErr != nil {
		mw.loadFailed = true
		mw.isReady = false
	} else {
		// Snapshot exists but has no mini downsamples yet, keep waiting
	}
}

func (mw *MiniWaveformComponent) ProcessUpdates() {
	mw.Component.ProcessUpdates()
}

func (mw *MiniWaveformComponent) handleUpdate(cmd component.UpdateCmd) {
	if mw.Component.HandleGlobalUpdate(cmd) {
		return
	}

	switch cmd.Type {
	case cmdUpdateStateFromCache:
		// This is triggered by an event from the event bus
		mw.updateStateFromCache()

	default:
		log.Warn("Unhandled update", zap.String("id", mw.IDStr()), zap.Any("cmd", cmd))
	}
}

// buildCache generates WaveformLines from the downsample data
func (mw *MiniWaveformComponent) buildCache(width float32, samples audio.Downsample) {
	numPixels := int(width)
	if numPixels < 1 {
		numPixels = 1
	}

	numSamples := len(samples.Mins)
	if len(samples.Maxs) < numSamples {
		numSamples = len(samples.Maxs)
	}

	if numSamples == 0 {
		mw.cachedLines = make([]WaveformLine, 0)
		mw.cacheValid = true
		return
	}

	samplesPerPixel := float32(numSamples) / float32(numPixels)

	// Find peak
	peak := float32(0.001)
	for i := 0; i < numSamples; i++ {
		absMin := samples.Mins[i]
		if absMin < 0 {
			absMin = -absMin
		}
		absMax := samples.Maxs[i]
		if absMax < 0 {
			absMax = -absMax
		}
		if absMin > peak {
			peak = absMin
		}
		if absMax > peak {
			peak = absMax
		}
	}

	// Normalization factor
	normFactor := float32(0.95) / peak
	if normFactor > 20.0 {
		normFactor = 20.0
	}

	// Build cache
	newCache := make([]WaveformLine, numPixels)
	for x := 0; x < numPixels; x++ {
		startIdx := int(float32(x) * samplesPerPixel)
		endIdx := int(float32(x+1) * samplesPerPixel)

		// Clamp
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx >= numSamples {
			startIdx = numSamples - 1
		}
		if endIdx > numSamples {
			endIdx = numSamples
		}
		if endIdx <= startIdx {
			endIdx = startIdx + 1
		}

		// Find min/max
		minVal := samples.Mins[startIdx]
		maxVal := samples.Maxs[startIdx]
		for i := startIdx + 1; i < endIdx && i < numSamples; i++ {
			if samples.Mins[i] < minVal {
				minVal = samples.Mins[i]
			}
			if samples.Maxs[i] > maxVal {
				maxVal = samples.Maxs[i]
			}
		}

		// Normalize and clamp
		minVal *= normFactor
		maxVal *= normFactor
		if minVal < -1.0 {
			minVal = -1.0
		}
		if minVal > 1.0 {
			minVal = 1.0
		}
		if maxVal < -1.0 {
			maxVal = -1.0
		}
		if maxVal > 1.0 {
			maxVal = 1.0
		}

		// Calculate color using current colormap from theme
		color := imgui.ColorU32Vec4(
			implot.SampleColormapV((maxVal+0.5)/2.0, theme.GetCurrentColormap()),
		)

		newCache[x] = WaveformLine{
			X:     float32(x),
			YMin:  minVal,
			YMax:  maxVal,
			Color: color,
		}
	}

	mw.cachedLines = newCache
	mw.cacheValid = true
	mw.normalizeFactor = normFactor
	mw.lastCachedColormap = theme.GetCurrentColormap()
}

// renderFromCache draws the cached lines
func (mw *MiniWaveformComponent) renderFromCache(width, height float32) {
	if len(mw.cachedLines) == 0 {
		mw.renderPlaceholder(width, height)
		return
	}

	t := theme.GetCurrentTheme()
	size := imgui.Vec2{X: width, Y: height}
	pos := imgui.CursorScreenPos()
	centerY := pos.Y + (size.Y * 0.5)
	scaleY := size.Y * 0.5
	padding := float32(4)
	paddedPos := imgui.Vec2{X: pos.X + padding, Y: pos.Y + padding}
	paddedSize := imgui.Vec2{X: size.X - (2 * padding), Y: size.Y - (2 * padding)}

	widthFactor := paddedSize.X / float32(len(mw.cachedLines))
	if len(mw.cachedLines) == 0 {
		widthFactor = 0
	}

	lineThickness := float32(1.0)
	if widthFactor >= 1.0 {
		lineThickness = widthFactor
	}

	imgui.InvisibleButtonV(mw.IDStr(), size, imgui.ButtonFlagsNone)
	dl := imgui.WindowDrawList()
	dl.AddRectFilledV(
		pos,
		imgui.Vec2{X: pos.X + size.X, Y: pos.Y + size.Y},
		imgui.ColorU32Vec4(t.Style.Colors.Button.Vec4),
		4.0,
		imgui.DrawFlagsNone,
	)
	dl.AddRectV(
		pos,
		imgui.Vec2{X: pos.X + size.X, Y: pos.Y + size.Y},
		imgui.ColorU32Vec4(t.Style.Colors.Border.Vec4),
		4.0,
		imgui.DrawFlagsNone,
		1.0,
	)

	for i, line := range mw.cachedLines {
		yMin := centerY - (line.YMin * scaleY)
		yMax := centerY - (line.YMax * scaleY)
		clampedYMin := math.Max(float64(paddedPos.Y), math.Min(float64(paddedPos.Y+paddedSize.Y), float64(yMin)))
		clampedYMax := math.Max(float64(paddedPos.Y), math.Min(float64(paddedPos.Y+paddedSize.Y), float64(yMax)))

		xPos := paddedPos.X + (float32(i) * widthFactor)

		dl.AddLineV(
			imgui.Vec2{X: xPos, Y: float32(clampedYMin)},
			imgui.Vec2{X: xPos, Y: float32(clampedYMax)},
			line.Color,
			lineThickness,
		)
	}
}

func (mw *MiniWaveformComponent) Layout() {
	mw.drainEvents()
	mw.ProcessUpdates()

	// On first render, check cache and request load if needed
	if !mw.isReady && !mw.loadFailed {
		mw.updateStateFromCache()
	}

	cellPaddingY := imgui.CurrentStyle().CellPadding().Y
	height := imgui.TextLineHeightWithSpacing() + cellPaddingY
	width := imgui.ContentRegionAvail().X
	if width < 10.0 {
		width = 10.0
	}
	currentWidth := width
	mw.Component.SetWidth(width)

	layoutChanged := mw.lastCachedWidth != currentWidth
	widthDiff := currentWidth - mw.lastCachedWidth
	if widthDiff < 0 {
		widthDiff = -widthDiff
	}
	significantWidthChange := widthDiff > 20.0

	// Check if colormap changed
	colormapChanged := mw.lastCachedColormap != theme.GetCurrentColormap()

	needsRebuild := false
	if layoutChanged && significantWidthChange {
		needsRebuild = true
	}
	if colormapChanged {
		needsRebuild = true
	}
	if !mw.cacheValid {
		needsRebuild = true
	}

	// Rebuild cache if needed
	if mw.isReady && !mw.loadFailed && needsRebuild {
		currentTime := imgui.Time()
		timeSinceLastRequest := currentTime - mw.lastCacheRequestTime

		// 150ms debounce
		if timeSinceLastRequest >= 0.15 {
			cache := audio.GetGlobalAsyncCache()
			snapshot := cache.GetSnapshot(mw.path)

			if snapshot != nil && len(snapshot.MiniDownsamples) > 0 && len(snapshot.MiniDownsamples[0].Mins) > 0 {
				padding := float32(4)
				paddedWidth := currentWidth - (2 * padding)
				if paddedWidth < 1 {
					paddedWidth = 1
				}

				mw.buildCache(paddedWidth, snapshot.MiniDownsamples[0])
				mw.lastCachedWidth = currentWidth
				mw.lastCacheRequestTime = currentTime
			}
		}
	}

	if mw.loadFailed {
		mw.renderPlaceholder(currentWidth, height)
		return
	}

	if mw.cachedLines != nil && len(mw.cachedLines) > 0 {
		mw.renderFromCache(currentWidth, height)
	} else {
		mw.renderPlaceholder(currentWidth, height)
	}
}

// renderPlaceholder draws a simple box when no data is available
func (mw *MiniWaveformComponent) renderPlaceholder(width, height float32) {
	t := theme.GetCurrentTheme()
	size := imgui.Vec2{X: width, Y: height}
	pos := imgui.CursorScreenPos()

	imgui.InvisibleButtonV(mw.IDStr(), size, imgui.ButtonFlagsNone)

	dl := imgui.WindowDrawList()
	dl.AddRectFilledV(
		pos,
		imgui.Vec2{X: pos.X + size.X, Y: pos.Y + size.Y},
		imgui.ColorU32Vec4(t.Style.Colors.Button.Vec4),
		4.0,
		imgui.DrawFlagsNone,
	)
	dl.AddRectV(
		pos,
		imgui.Vec2{X: pos.X + size.X, Y: pos.Y + size.Y},
		imgui.ColorU32Vec4(t.Style.Colors.Border.Vec4),
		4.0,
		imgui.DrawFlagsNone,
		1.0,
	)
}

// Destroy cleans up the component
func (mw *MiniWaveformComponent) Destroy() {
	// Unsubscribe from filtered subscriptions (handles all event types)
	if mw.filteredEventSub != nil {
		mw.filteredEventSub.Unsubscribe()
	}

	// Call the base component's destroy method
	mw.Component.Destroy()
}
