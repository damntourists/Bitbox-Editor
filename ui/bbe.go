package ui

import (
	"bitbox-editor/lib/events"
	"bitbox-editor/lib/logging"
	"bitbox-editor/lib/preset"
	"bitbox-editor/lib/util"
	uiEvents "bitbox-editor/ui/events"
	"bitbox-editor/ui/fonts"
	"bitbox-editor/ui/theme"
	"bitbox-editor/ui/windows"
	"context"
	"image"
	"runtime"
	"sync"

	"fmt"
	_ "image/png"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
)

var log = logging.NewLogger("bbe")
var windowClass *imgui.WindowClass
var toolbarHeight = float32(50.0)

func init() {
	runtime.LockOSThread()

	windowClass = imgui.NewWindowClass()
	windowClass.SetViewportFlagsOverrideSet(imgui.ViewportFlagsNoAutoMerge)
}

type BitboxEditor struct {
	dockspaceID imgui.ID

	Window struct {
		Settings *windows.SettingsWindow
		Console  *windows.ConsoleWindow

		Storage *windows.StorageWindow
		Presets *windows.PresetListWindow
		Library *windows.LibraryWindow

		Editors []*windows.PresetEditWindow
	}

	Modal struct {
	}

	theme *theme.Theme

	backend backend.Backend[glfwbackend.GLFWWindowFlags]

	implotCtx *implot.Context

	initialized bool

	sync.RWMutex
}

func (b *BitboxEditor) afterCreateContext() {
	b.implotCtx = implot.CreateContext()
}

func (b *BitboxEditor) beforeDestroyContext() {
	implot.DestroyContext()
}

func (b *BitboxEditor) beforeRender() {}

func (b *BitboxEditor) afterRender() {
}

func (b *BitboxEditor) close() {
	log.Info("Closing window")
	b.backend.SetShouldClose(true)
}

func (b *BitboxEditor) onDrop(files []string) {
	fmt.Println("Dropped files: ", files)
}

func (b *BitboxEditor) setup() {
	// Load window icons
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

	// Load theme
	b.theme = theme.GetCurrentTheme()

	glfwbackend.ForceX11()

	be, err := backend.CreateBackend[glfwbackend.GLFWWindowFlags](glfwbackend.NewGLFWBackend())
	if err != nil {
		panic(fmt.Errorf("failed to create GLFW backend: %w", err))
	}

	b.backend = be

	// Register hooks
	b.backend.SetBeforeRenderHook(b.beforeRender)
	b.backend.SetAfterRenderHook(b.afterRender)
	b.backend.SetAfterCreateContextHook(b.afterCreateContext)
	b.backend.SetBeforeDestroyContextHook(b.beforeDestroyContext)

	// Create the window
	b.backend.SetBgColor(b.theme.Style.Colors.ScrollbarBg.Vec4) //imgui.NewVec4(0.45, 0.55, 0.6, 1.0))
	b.backend.CreateWindow("Bitbox Editor v0.0.0", 1400, 1000)
	b.backend.SetIcons(icons...)

	b.backend.SetCloseCallback(b.close)
	b.backend.SetDropCallback(b.onDrop)

	if err != nil {
		panic(err)
	}

	// Setup IO
	io := imgui.CurrentIO()
	io.SetIniFilename("ui.ini")
	io.SetConfigFlags(imgui.ConfigFlagsDockingEnable)
}

func (b *BitboxEditor) initWindows() {
	b.Window.Settings = windows.NewSettingsWindow()
	b.Window.Console = windows.NewConsoleWindow()
	b.Window.Storage = windows.NewStorageWindow()
	b.Window.Presets = windows.NewPresetListWindow()
	b.Window.Library = windows.NewLibraryWindow()

	b.Window.Editors = make([]*windows.PresetEditWindow, 0)

	b.Window.Settings.Close()

	// Register Storage Window events
	b.Window.Storage.Events.AddListener(
		func(ctx context.Context, record events.StorageEventRecord) {
			if record.Type == events.StorageActivatedEvent {
				b.Window.Presets.SetPresetLocation(record.Data.(*windows.StorageLocation))
				b.Window.Library.SetStorageLocation(record.Data.(*windows.StorageLocation))
			}
		},
	)

	// Register Preset Window events
	b.Window.Presets.Events.AddListener(
		func(ctx context.Context, record events.PresetEventRecord) {
			switch record.Type {
			case events.LoadPreset:
				p := record.Data.(*preset.Preset)

				// Check if we already have the window open and focus it.
				for _, win := range b.Window.Editors {
					if win.Preset() == p {
						imgui.SetWindowFocusStr(win.Title())
						return
					}
				}

				// Create  open a new Preset Edit Window
				editWindow := windows.NewPresetEditWindow(p)
				editWindow.SetPreset(p)
				editWindow.Open()

				editWindow.WindowEvents.AddListener(
					func(ctx context.Context, record uiEvents.WindowEventRecord) {
						switch record.Type {
						case uiEvents.WindowClose:
							println("WINDOW CLOSE")
							b.Window.Editors = append(b.Window.Editors[:len(b.Window.Editors)-1])
						}
					},
					fmt.Sprintf("%s-window-events", editWindow.Title()),
				)

				b.Window.Editors = append(b.Window.Editors, editWindow)
			}
		})

}

func (b *BitboxEditor) menu() {
	if imgui.BeginMainMenuBar() {

		if imgui.BeginMenu("File") {
			if imgui.MenuItemBoolV(
				"Exit",
				"CTRL+Q",
				false,
				true) {

				b.close()

			}

			imgui.EndMenu()
		}

		if imgui.BeginMenu("Edit") {
			if imgui.MenuItemBoolV(
				"Settings",
				"CTRL+ALT+S",
				false,
				true) {

				b.Window.Settings.Open()
			}

			imgui.EndMenu()
		}

		if imgui.BeginMenu("Help") {
			imgui.EndMenu()
		}

		imgui.EndMainMenuBar()
	}
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

	imgui.SetNextWindowPos(imgui.Vec2{X: viewportPos.X, Y: viewportPos.Y + mainMenuHeight})

	imgui.SetNextWindowSize(imgui.Vec2{X: viewportSize.X, Y: toolbarHeight})

	imgui.BeginV("toolbar", nil, toolbarFlags)
	if imgui.Button(b.Window.Storage.Icon()) {
		log.Debug("btn 1 clicked")
	}
	imgui.SameLine()
	if imgui.Button(b.Window.Presets.Icon()) {
		log.Debug("btn 2 clicked")
	}
	imgui.SameLine()
	if imgui.Button(b.Window.Console.Icon()) {
		log.Debug("btn 3 clicked")
	}
	imgui.SameLine()
	if imgui.Button(b.Window.Library.Icon()) {
		log.Debug("btn 4 clicked")
	}
	imgui.End()
}

func (b *BitboxEditor) dockspace() {
	viewport := imgui.MainViewport()
	viewportPos := viewport.Pos()
	viewportSize := viewport.Size()

	b.dockspaceID = imgui.IDStr("dockspace")

	dockSpaceFlags := imgui.DockNodeFlagsPassthruCentralNode

	dockspacePos := imgui.Vec2{
		X: viewportPos.X,
		Y: viewportPos.Y + imgui.FrameHeight() + toolbarHeight,
	}
	dockspaceSize := imgui.Vec2{
		X: viewportSize.X,
		Y: viewportSize.Y - imgui.FrameHeight() - toolbarHeight,
	}
	imgui.SetNextWindowPos(dockspacePos)
	imgui.SetNextWindowSize(dockspaceSize)
	imgui.SetNextWindowBgAlpha(0)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})

	defer func() {
		imgui.PopStyleVar()
	}()

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
}

func (b *BitboxEditor) loop() {
	imgui.ClearSizeCallbackPool()

	if !fonts.FontsInitialized {
		return
	}

	if !b.initialized {
		b.initWindows()

		b.initialized = true
	}

	themeFin := b.theme.Apply()

	defer themeFin()

	b.menu()
	b.toolbar()
	b.dockspace()

	// Layout windows
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

	for _, editWindow := range b.Window.Editors {
		imgui.SetNextWindowDockIDV(b.dockspaceID, imgui.CondOnce)
		editWindow.Build()
	}

}

func (b *BitboxEditor) Run() {
	if b.backend == nil {
		panic("backend is nil: setup() did not initialize the backend")
	}
	switch v := any(b.backend).(type) {
	case *glfwbackend.GLFWBackend:
		if v == nil {
			panic("backend is a typed-nil *GLFWBackend")
		}
	}

	b.backend.Run(b.loop)
}

func NewBitboxEditor() *BitboxEditor {
	app := &BitboxEditor{
		initialized: false,
	}

	app.setup()

	fonts.RebuildFonts(app.backend)

	return app
}
