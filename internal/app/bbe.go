package app

import (
	"bitbox-editor/internal/app/component/button"
	"bitbox-editor/internal/app/component/canvas"
	"bitbox-editor/internal/app/component/spectrum"
	"bitbox-editor/internal/app/component/volume"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/app/font"
	"bitbox-editor/internal/app/theme"
	"bitbox-editor/internal/app/window/console"
	"bitbox-editor/internal/app/window/library"
	"bitbox-editor/internal/app/window/presetedit"
	"bitbox-editor/internal/app/window/presetlist"
	"bitbox-editor/internal/app/window/settings"
	"bitbox-editor/internal/app/window/storage"
	"bitbox-editor/internal/audio"
	"bitbox-editor/internal/config"
	"bitbox-editor/internal/logging"
	"bitbox-editor/internal/preset"
	"bitbox-editor/internal/util"
	"fmt"
	"image"
	_ "image/png"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var log = logging.NewLogger("bbe")
var windowClass *imgui.WindowClass

func init() {
	windowClass = imgui.NewWindowClass()
	windowClass.SetViewportFlagsOverrideSet(imgui.ViewportFlagsNoAutoMerge)
}

// BitboxEditor holds the main application state
type BitboxEditor struct {
	dockspaceID imgui.ID
	uuid        string

	Window struct {
		Settings *settings.SettingsWindow
		Console  *console.ConsoleWindow
		Storage  *storage.StorageWindow
		Presets  *presetlist.PresetListWindow
		Library  *library.LibraryWindow
		Editors  []*presetedit.PresetEditWindow
	}

	Modal struct{}

	spectrumAnalyzer *spectrum.SpectrumAnalyzerComponent
	volumeControl    *volume.VolumeControlComponent

	// Toolbar buttons
	storageButton *button.Button
	presetsButton *button.Button
	consoleButton *button.Button
	libraryButton *button.Button

	canvas *canvas.RenderPrimitive

	colormap implot.Colormap

	backend backend.Backend[glfwbackend.GLFWWindowFlags]

	initialized bool

	updates  chan UpdateCmd
	handler  UpdateHandlerFunc
	eventSub chan events.Event
}

func NewBitboxEditor() *BitboxEditor {
	app := &BitboxEditor{
		uuid:        uuid.NewString(),
		initialized: false,
		updates:     make(chan UpdateCmd, 50),
		eventSub:    make(chan events.Event, 100),
	}
	app.handler = app.handleUpdate
	app.setup()

	font.InitAndRebuildFonts(app.backend)
	font.SetGlobalScale(1)
	//font.RebuildFonts()

	return app
}

// drainEvents translates global bus events into local commands
func (b *BitboxEditor) drainEvents() {
	for {
		select {
		case event := <-b.eventSub:
			// Translate *all* public events into a command
			// The event struct itself is the "Type"
			b.SendUpdate(UpdateCmd{Type: event, Data: event})
		default:
			// No more events
			return
		}
	}
}

// handleUpdate is the main command processor for the app
func (b *BitboxEditor) handleUpdate(cmd UpdateCmd) {
	switch c := cmd.Type.(type) {

	case localCommand:
		switch c {
		case cmdEditorCreate:
			if payload, ok := cmd.Data.(editorCreatePayload); ok && payload.Preset != nil {
				p := payload.Preset
				for _, win := range b.Window.Editors {
					if win != nil && win.Preset() == p {
						imgui.SetWindowFocusStr(win.Title())
						return
					}
				}
				audioMgr := audio.GetAudioManager()
				editWindow := presetedit.NewPresetEditWindow(p, audioMgr)
				b.Window.Editors = append(b.Window.Editors, editWindow)
			}

		case cmdEditorAdd:
			if payload, ok := cmd.Data.(editorAddPayload); ok && payload.Editor != nil {
				b.Window.Editors = append(b.Window.Editors, payload.Editor)
			}

		case cmdEditorRemove:
			if payload, ok := cmd.Data.(editorRemovePayload); ok && payload.Editor != nil {
				newEditors := b.Window.Editors[:0]
				removed := false
				for _, editor := range b.Window.Editors {
					if editor != payload.Editor {
						newEditors = append(newEditors, editor)
					} else {
						editor.Destroy()
						removed = true
					}
				}
				if removed {
					b.Window.Editors = newEditors
				}
			}
		}
		return

	case events.PresetEventRecord:
		if c.EventType == events.PresetLoadEvent {
			if p, ok := c.Data.(*preset.Preset); ok && p != nil {
				log.Debug("App received LoadPreset event, creating editor", zap.String("preset", p.Name))
				b.SendUpdate(UpdateCmd{
					Type: cmdEditorCreate,
					Data: editorCreatePayload{Preset: p},
				})
			}
		}
		return

	case events.StorageEventRecord:
		if c.EventType == events.StorageActivatedEvent {
			if loc, ok := c.Data.(*storage.StorageLocation); ok {
				log.Debug("App received StorageActivated event", zap.String("path", loc.Path))
				b.Window.Presets.SetPresetLocation(loc)
				b.Window.Library.SetStorageLocation(loc)
			}
		}
		return

	case events.AudioVolumeEventRecord:
		if b.volumeControl != nil {
			b.volumeControl.SetVolume(float32(c.Volume))
		}
		return

	case events.WindowEventRecord:
		if c.EventType == events.WindowCloseEvent || c.EventType == events.WindowDestroyEvent {
			var editorToRemove *presetedit.PresetEditWindow
			for _, editor := range b.Window.Editors {
				if editor.UUID() == c.WindowID {
					editorToRemove = editor
					break
				}
			}
			if editorToRemove != nil {
				b.SendUpdate(UpdateCmd{
					Type: cmdEditorRemove,
					Data: editorRemovePayload{Editor: editorToRemove},
				})
			}
		}
		return

	default:
		log.Warn("BitboxEditor unhandled update type", zap.Any("type", fmt.Sprintf("%T", cmd.Type)))
	}
}

func (b *BitboxEditor) afterCreateContext() {
	implot.CreateContext()
}

func (b *BitboxEditor) beforeDestroyContext() {
	implot.DestroyContext()
}

func (b *BitboxEditor) beforeRender() {}

func (b *BitboxEditor) afterRender() {}

func (b *BitboxEditor) setup() {
	var icons []image.Image
	iconNames := []string{
		"icon16.png",
		"icon24.png",
		"icon32.png",
		"icon64.png",
		"icon128.png",
		"icon256.png",
		"icon512.png",
	}

	for _, iconName := range iconNames {
		icon, err := backend.LoadImage(fmt.Sprintf("./resources/icons/%s", iconName))
		util.PanicOnError(err)
		icons = append(icons, icon)
	}

	b.colormap = theme.GetCurrentColormap()

	glfwbackend.ForceX11()
	be, err := backend.CreateBackend[glfwbackend.GLFWWindowFlags](glfwbackend.NewGLFWBackend())
	if err != nil {
		panic(fmt.Errorf("failed to create GLFW backend: %w", err))
	}
	b.backend = be

	b.backend.SetBeforeRenderHook(b.beforeRender)
	b.backend.SetAfterRenderHook(b.afterRender)
	b.backend.SetAfterCreateContextHook(b.afterCreateContext)
	b.backend.SetBeforeDestroyContextHook(b.beforeDestroyContext)

	b.backend.CreateWindow("Bitbox Editor v0.0.0", 1400, 1000)
	b.backend.SetIcons(icons...)

	b.backend.SetCloseCallback(b.close)
	b.backend.SetDropCallback(b.onDrop)

	io := imgui.CurrentIO()
	io.SetIniFilename("app.ini")
	io.SetConfigFlags(imgui.ConfigFlagsDockingEnable)
}

func (b *BitboxEditor) initWindows() {

	b.canvas = canvas.NewRenderPrimitive(
		imgui.IDStr(fmt.Sprintf("##canvas:%s", b.uuid)),
		"sphere",
	)

	audioMgr := audio.GetAudioManager()
	// TODO: Finish building out midi manager.
	//midiMgr := midi.GetMidiManager()
	//log.Info("midi ports:", zap.Any("ports", midiMgr.ListPorts()))

	b.Window.Settings = settings.NewSettingsWindow()
	b.Window.Console = console.NewConsoleWindow()
	b.Window.Storage = storage.NewStorageWindow()
	b.Window.Presets = presetlist.NewPresetListWindow()
	b.Window.Library = library.NewLibraryWindow()

	b.Window.Editors = make([]*presetedit.PresetEditWindow, 0)

	toolbarSize := config.GetToolbarSize()

	b.spectrumAnalyzer = spectrum.NewSpectrumAnalyzer(imgui.IDStr("toolbar_spectrum"), audioMgr).
		SetHeight(toolbarSize - 16).
		SetPadding(2)

	b.volumeControl = volume.NewVolumeControlWithID(imgui.IDStr("toolbar_volume")).
		SetWidth(120).
		SetHeight(8).
		SetRadius(4).
		SetVolume(float32(audioMgr.GetVolume()))

	buttonSize := toolbarSize - 8 // 8 pixels padding

	b.storageButton = button.NewButtonWithID(imgui.IDStr("toolbar_storage"), b.Window.Storage.Icon()).
		SetFixedSize(buttonSize, buttonSize).
		SetPadding(theme.GetCurrentTheme().Style.FramePadding[0]).
		SetRounding(theme.GetCurrentTheme().Style.FrameRounding * 1.9).
		SetToggledColor(theme.GetCurrentTheme().Style.Colors.TabHovered.Vec4).
		SetOnClick(func() { b.Window.Storage.ToggleOpen() })

	b.presetsButton = button.NewButtonWithID(imgui.IDStr("toolbar_presets"), b.Window.Presets.Icon()).
		SetFixedSize(buttonSize, buttonSize).
		SetPadding(theme.GetCurrentTheme().Style.FramePadding[0]).
		SetRounding(theme.GetCurrentTheme().Style.FrameRounding * 1.9).
		SetOnClick(func() { b.Window.Presets.ToggleOpen() })

	b.consoleButton = button.NewButtonWithID(imgui.IDStr("toolbar_console"), b.Window.Console.Icon()).
		SetFixedSize(buttonSize, buttonSize).
		SetPadding(theme.GetCurrentTheme().Style.FramePadding[0]).
		SetRounding(theme.GetCurrentTheme().Style.FrameRounding * 1.9).
		SetOnClick(func() { b.Window.Console.ToggleOpen() })

	b.libraryButton = button.NewButtonWithID(imgui.IDStr("toolbar_library"), b.Window.Library.Icon()).
		SetFixedSize(buttonSize, buttonSize).
		SetPadding(theme.GetCurrentTheme().Style.FramePadding[0]).
		SetRounding(theme.GetCurrentTheme().Style.FrameRounding * 1.9).
		SetOnClick(func() { b.Window.Library.ToggleOpen() })

	b.volumeControl.SetOnVolumeChange(func(volume float32) {
		audioMgr.SetVolume(float64(volume))
	})

	eventbus.Bus.Subscribe(events.AudioVolumeChangedKey, b.uuid, b.eventSub)
	eventbus.Bus.Subscribe(events.StorageActivatedEventKey, b.uuid, b.eventSub)
	eventbus.Bus.Subscribe(events.PresetLoadEventKey, b.uuid, b.eventSub)
	eventbus.Bus.Subscribe(events.WindowCloseEventKey, b.uuid, b.eventSub)
	eventbus.Bus.Subscribe(events.WindowDestroyEventKey, b.uuid, b.eventSub)
}

func (b *BitboxEditor) menu() {
	if imgui.BeginMainMenuBar() {
		if imgui.BeginMenu("File") {
			if imgui.MenuItemBoolV("Exit", "CTRL+Q", false, true) {
				b.close()
			}
			imgui.EndMenu()
		}
		if imgui.BeginMenu("Edit") {
			if imgui.MenuItemBoolV("Settings", "CTRL+ALT+S", false, true) {
				b.Window.Settings.SetOpen()
			}
			imgui.EndMenu()
		}
		if imgui.BeginMenu("View") {
			imgui.EndMenu()
		}
		if len(b.Window.Editors) > 0 {
			if imgui.BeginMenu("Window") {
				for i, win := range b.Window.Editors {
					var shortcut string
					if i < 8 {
						shortcut = fmt.Sprintf("ALT+%d", i+1)
					} else if i == 9 {
						shortcut = "ALT+0"
					}
					if imgui.MenuItemBoolV(win.Title(), shortcut, false, true) {
						imgui.SetWindowFocusStr(win.Title())
					}
				}
				imgui.EndMenu()
			}
		}
		if imgui.BeginMenu("Help") {
			imgui.EndMenu()
		}
		imgui.EndMainMenuBar()
	}
}

func (b *BitboxEditor) close() {
	b.backend.SetShouldClose(true)
}

func (b *BitboxEditor) onDrop(files []string) {
	fmt.Println("Dropped files: ", files)
}

func (b *BitboxEditor) toolbar() {
	viewport := imgui.MainViewport()
	viewportPos := viewport.Pos()
	viewportSize := viewport.Size()

	toolbarFlags := imgui.WindowFlagsNoTitleBar |
		imgui.WindowFlagsNoResize |
		imgui.WindowFlagsNoMove |
		imgui.WindowFlagsNoDocking |
		imgui.WindowFlagsNoCollapse |
		imgui.WindowFlagsNoScrollbar

	mainMenuHeight := imgui.FrameHeight()
	toolbarSize := config.GetToolbarSize()
	//toolbarPadding := config.GetToolbarPadding()
	toolbarPlacement := config.GetToolbarPlacement()
	toolbarMargin := config.GetToolbarMargin()

	// Calculate toolbar position and size based on placement
	var toolbarPos imgui.Vec2
	var toolbarWindowSize imgui.Vec2
	var isVertical bool

	switch toolbarPlacement {
	case "left":
		toolbarPos = imgui.Vec2{
			X: viewportPos.X + toolbarMargin,
			Y: viewportPos.Y + mainMenuHeight + toolbarMargin,
		}
		toolbarWindowSize = imgui.Vec2{
			X: toolbarSize,
			Y: viewportSize.Y - mainMenuHeight - (toolbarMargin * 2),
		}
		isVertical = true
	case "right":
		toolbarPos = imgui.Vec2{
			X: viewportPos.X + viewportSize.X - toolbarSize - toolbarMargin,
			Y: viewportPos.Y + mainMenuHeight + toolbarMargin,
		}
		toolbarWindowSize = imgui.Vec2{
			X: toolbarSize,
			Y: viewportSize.Y - mainMenuHeight - (toolbarMargin * 2),
		}
		isVertical = true
	case "bottom":
		toolbarPos = imgui.Vec2{
			X: viewportPos.X + toolbarMargin,
			Y: viewportPos.Y + viewportSize.Y - toolbarSize - toolbarMargin,
		}
		toolbarWindowSize = imgui.Vec2{
			X: viewportSize.X - (toolbarMargin * 2),
			Y: toolbarSize,
		}
		isVertical = false
	default: // "top"
		toolbarPos = imgui.Vec2{
			X: viewportPos.X + toolbarMargin,
			Y: viewportPos.Y + mainMenuHeight + toolbarMargin,
		}
		toolbarWindowSize = imgui.Vec2{
			X: viewportSize.X - (toolbarMargin * 2),
			Y: toolbarSize,
		}
		isVertical = false
	}

	imgui.SetNextWindowPos(toolbarPos)
	imgui.SetNextWindowSize(toolbarWindowSize)

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 8, Y: 4})
	defer imgui.PopStyleVar()

	imgui.BeginV("toolbar", nil, toolbarFlags)
	defer imgui.End()

	// Update button toggle states to reflect window visibility
	b.storageButton.SetToggled(b.Window.Storage.IsOpen())
	b.presetsButton.SetToggled(b.Window.Presets.IsOpen())
	b.consoleButton.SetToggled(b.Window.Console.IsOpen())
	b.libraryButton.SetToggled(b.Window.Library.IsOpen())

	if isVertical {
		// Vertical toolbar layout (left/right)
		buttonSize := toolbarSize - 8
		buttonPadding := (toolbarSize-buttonSize)/2 - 4
		imgui.SetCursorPosX(imgui.CursorPosX() + buttonPadding)

		b.storageButton.Build()
		b.presetsButton.Build()
		b.consoleButton.Build()
		b.libraryButton.Build()

		if b.spectrumAnalyzer != nil && b.volumeControl != nil {
			availHeight := imgui.ContentRegionAvail().Y
			spectrumHeight := float32(365.0)

			if availHeight > spectrumHeight+10 {
				imgui.SetCursorPosY(imgui.CursorPosY() + availHeight - spectrumHeight)
			}

			imgui.BeginGroup()
			imgui.Dummy(imgui.Vec2{X: (toolbarSize - 30) / 2, Y: 1.0})
			imgui.SameLine()
			b.volumeControl.Build()
			imgui.EndGroup()

			imgui.Dummy(imgui.Vec2{X: 1, Y: 32.0})

			b.spectrumAnalyzer.Build()
		}
	} else {
		// Horizontal toolbar layout (top/bottom)
		buttonSize := toolbarSize - 8
		buttonPadding := (toolbarSize-buttonSize)/2 - 4
		imgui.SetCursorPosY(imgui.CursorPosY() + buttonPadding)

		b.storageButton.Build()
		imgui.SameLine()
		b.presetsButton.Build()
		imgui.SameLine()
		b.consoleButton.Build()
		imgui.SameLine()
		b.libraryButton.Build()
		imgui.SameLine()

		if b.spectrumAnalyzer != nil && b.volumeControl != nil {
			availWidth := imgui.ContentRegionAvail().X
			spectrumWidth := float32(365.0)
			totalWidth := spectrumWidth

			if availWidth > totalWidth+10 {
				imgui.SetCursorPosX(imgui.CursorPosX() + availWidth - totalWidth)
			}

			imgui.BeginGroup()
			imgui.Dummy(imgui.Vec2{X: 1.0, Y: (toolbarSize - 30) / 2})
			b.volumeControl.Build()
			imgui.EndGroup()
			imgui.SameLine()

			imgui.Dummy(imgui.Vec2{X: 32.0, Y: 1})
			imgui.SameLine()

			b.spectrumAnalyzer.Build()
		}
	}
}

func (b *BitboxEditor) dockspace() {
	viewport := imgui.MainViewport()

	viewportPos := viewport.Pos()
	viewportSize := viewport.Size()

	b.dockspaceID = imgui.IDStr("dockspace")
	dockSpaceFlags := imgui.DockNodeFlagsPassthruCentralNode

	mainMenuHeight := imgui.FrameHeight()
	toolbarSize := config.GetToolbarSize()
	toolbarPadding := config.GetToolbarPadding()
	toolbarPlacement := config.GetToolbarPlacement()
	toolbarMargin := config.GetToolbarMargin()

	// Calculate dockspace position and size based on toolbar placement
	var dockspacePos imgui.Vec2
	var dockspaceSize imgui.Vec2

	switch toolbarPlacement {
	case "left":
		dockspacePos = imgui.Vec2{
			X: viewportPos.X + toolbarMargin + toolbarSize + toolbarPadding,
			Y: viewportPos.Y + mainMenuHeight + toolbarMargin,
		}
		dockspaceSize = imgui.Vec2{
			X: viewportSize.X - (toolbarMargin * 2) - toolbarSize - toolbarPadding,
			Y: viewportSize.Y - mainMenuHeight - (toolbarMargin * 2),
		}
	case "right":
		dockspacePos = imgui.Vec2{
			X: viewportPos.X + toolbarMargin,
			Y: viewportPos.Y + mainMenuHeight + toolbarMargin,
		}
		dockspaceSize = imgui.Vec2{
			X: viewportSize.X - (toolbarMargin * 2) - toolbarSize - toolbarPadding,
			Y: viewportSize.Y - mainMenuHeight - (toolbarMargin * 2),
		}
	case "bottom":
		dockspacePos = imgui.Vec2{
			X: viewportPos.X + toolbarMargin,
			Y: viewportPos.Y + mainMenuHeight + toolbarMargin,
		}
		dockspaceSize = imgui.Vec2{
			X: viewportSize.X - (toolbarMargin * 2),
			Y: viewportSize.Y - mainMenuHeight - (toolbarMargin * 2) - toolbarSize - toolbarPadding,
		}
	default: // "top"
		dockspacePos = imgui.Vec2{
			X: viewportPos.X + toolbarMargin,
			Y: viewportPos.Y + mainMenuHeight + toolbarMargin + toolbarSize + toolbarPadding,
		}
		dockspaceSize = imgui.Vec2{
			X: viewportSize.X - (toolbarMargin * 2),
			Y: viewportSize.Y - mainMenuHeight - (toolbarMargin * 2) - toolbarSize - toolbarPadding,
		}
	}

	imgui.SetNextWindowPos(dockspacePos)
	imgui.SetNextWindowSize(dockspaceSize)
	imgui.SetNextWindowBgAlpha(0)

	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})

	imgui.BeginV(
		"dockspace-window",
		nil,
		imgui.WindowFlagsNoTitleBar|
			imgui.WindowFlagsNoDocking|
			imgui.WindowFlagsNoResize|
			imgui.WindowFlagsNoMove|
			imgui.WindowFlagsNoCollapse,
	)

	imgui.DockSpaceV(b.dockspaceID, imgui.Vec2{X: 0, Y: 0}, dockSpaceFlags, windowClass)

	imgui.End()
	imgui.PopStyleVar()
}

func (b *BitboxEditor) loop() {
	b.drainEvents()
	b.ProcessUpdates()

	currentEditors := append([]*presetedit.PresetEditWindow(nil), b.Window.Editors...)

	if !font.FontsInitialized {
		return
	}

	if !b.initialized {
		b.initWindows()
		b.initialized = true
	}

	currentTheme := theme.GetCurrentTheme()
	themeFin := currentTheme.Apply()

	b.backend.SetBgColor(currentTheme.Style.Colors.ChildBg.Vec4)

	if b.canvas != nil {
		b.canvas.Build()
	}

	defer themeFin()

	b.menu()
	b.toolbar()
	b.dockspace()

	// Layout window
	if b.Window.Console.IsOpen() {
		b.Window.Console.Build()
	}
	if b.Window.Settings.IsOpen() {
		b.Window.Settings.Build()
	}
	if b.Window.Storage.IsOpen() {
		b.Window.Storage.Build()
	}
	if b.Window.Presets.IsOpen() {
		b.Window.Presets.Build()
	}
	if b.Window.Library.IsOpen() {
		b.Window.Library.Build()
	}

	for _, editWindow := range currentEditors {
		if editWindow != nil {
			imgui.SetNextWindowDockIDV(b.dockspaceID, imgui.CondOnce)
			editWindow.Build()
		}
	}

}

func (b *BitboxEditor) SendUpdate(cmd UpdateCmd) {
	select {
	case b.updates <- cmd:
	default:
		log.Warn("BitboxEditor update channel full, dropping command")
	}
}

func (b *BitboxEditor) ProcessUpdates() {
	if b.handler == nil {
		for {
			select {
			case <-b.updates:
			default:
				return
			}
		}
	}

	const maxMessagesPerFrame = 100
	for i := 0; i < maxMessagesPerFrame; i++ {
		select {
		case cmd := <-b.updates:
			b.handler(cmd)
		default:
			return
		}
	}

	if len(b.updates) > 0 {
		log.Warn("BitboxEditor ProcessUpdates hit message limit")
	}
}

func (b *BitboxEditor) UpdateChannel() chan<- UpdateCmd {
	return b.updates
}

func (b *BitboxEditor) Run() {
	if b.backend == nil {
		panic("backend is nil: setup() did not initialize the backend")
	}

	b.backend.Run(b.loop)
}
