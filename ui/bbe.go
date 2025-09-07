package ui

import (
	"bitbox-editor/lib/io/drive/detect"
	"bitbox-editor/lib/logging"
	"bitbox-editor/lib/util"
	"bitbox-editor/ui/fonts"
	"bitbox-editor/ui/theme"
	"bitbox-editor/ui/windows"
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
	Window struct {
		SettingsWindow *windows.SettingsWindow
		ConsoleWindow  *windows.ConsoleWindow
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
	b.backend.Refresh()
	//imgui.UpdatePlatformWindows()
	//imgui.RenderPlatformWindowsDefault()
}
func (b *BitboxEditor) close() {
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

	// Create GLFW backend (be explicit with the generic parameter)
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
	b.backend.SetBgColor(imgui.NewVec4(0.45, 0.55, 0.6, 1.0))
	b.backend.CreateWindow("Bitbox Editor", 800, 600)

	b.backend.SetIcons(icons...)

	// Register window callbacks (window must exist)
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
	b.Window.SettingsWindow = windows.NewSettingsWindow()
	b.Window.ConsoleWindow = windows.NewConsoleWindow()

	//b.Window.Viewers = make([]*windows.ViewerWindow, 0)
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

				b.Window.SettingsWindow.Open()
			}

			imgui.EndMenu()
		}

		if imgui.BeginMenu("Window") {

			if imgui.MenuItemBoolV(
				"NewConsoleWindow",
				"ALT+6",
				b.Window.ConsoleWindow.IsOpen(),
				true) {

				b.Window.ConsoleWindow.Open()
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
	if imgui.Button("btn1") {
		log.Debug("btn 1 clicked")
	}
	imgui.SameLine()
	if imgui.Button("btn2") {
		log.Debug("btn 2 clicked")
	}
	imgui.End()
}

func (b *BitboxEditor) dockspace() {
	viewport := imgui.MainViewport()
	viewportPos := viewport.Pos()
	viewportSize := viewport.Size()

	dockSpaceFlags := imgui.DockNodeFlagsPassthruCentralNode

	dockSpace := imgui.IDStr("dockspace")

	dockspacePos := imgui.Vec2{X: viewportPos.X, Y: viewportPos.Y + imgui.FrameHeight() + toolbarHeight}
	dockspaceSize := imgui.Vec2{X: viewportSize.X, Y: viewportSize.Y - imgui.FrameHeight() - toolbarHeight}

	imgui.SetNextWindowPos(dockspacePos)
	imgui.SetNextWindowSize(dockspaceSize)
	imgui.SetNextWindowBgAlpha(0)
	imgui.BeginV(
		"dockspace-window",
		nil,
		imgui.WindowFlagsNoTitleBar|
			imgui.WindowFlagsNoDocking|
			imgui.WindowFlagsNoResize|
			imgui.WindowFlagsNoMove|
			imgui.WindowFlagsNoCollapse,
	)

	imgui.DockSpaceV(dockSpace, imgui.Vec2{X: 0, Y: 0}, dockSpaceFlags, windowClass)
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

	// Set theme and pop theme after loop
	themeFin := b.theme.Apply()
	defer themeFin()

	// Main app framework
	b.menu()
	b.toolbar()
	b.dockspace()

	// Layout windows
	b.Window.ConsoleWindow.Build()

	if b.Window.SettingsWindow.IsOpen() {
		b.Window.SettingsWindow.Build()
	}
}

func (b *BitboxEditor) Run() {
	// Guard both nil interface and typed-nil concrete
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

	// Start drive detect goroutine
	go func() {
		if drives, err := detect.Detect(); err == nil {
			log.Debug(fmt.Sprintf("%d USB Devices Found", len(drives)))
			for _, d := range drives {
				log.Debug(d)
			}
		} else {
			//log.Debug(err.Error())
		}
	}()

	fonts.RebuildFonts(app.backend)

	return app
}
