package windows

import (
	"bitbox-editor/ui/events"
	"bitbox-editor/ui/fonts"
	"bitbox-editor/ui/types"
	"context"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maniartech/signals"
)

type (

	// Window represents a customizable Backend windows
	Window struct {
		open      bool
		title     string
		subtitle  string
		icon      string
		noClose   bool
		collapsed bool

		WindowEvents signals.Signal[events.WindowEventRecord]

		config *WindowConfig

		layoutBuilder types.LayoutBuilder
	}
)

func (w *Window) Title() string { return fmt.Sprintf("%s %s", fonts.Icon(w.icon), w.title) }
func (w *Window) Icon() string  { return w.icon }
func (w *Window) Close() {
	w.open = false
	w.WindowEvents.Emit(context.Background(), events.WindowEventRecord{
		Type:        events.WindowClose,
		WindowTitle: w.Title(),
	})
}
func (w *Window) Open() {
	w.open = true
	w.WindowEvents.Emit(context.Background(), events.WindowEventRecord{
		Type:        events.WindowOpen,
		WindowTitle: w.Title(),
	})

}
func (w *Window) IsOpen() bool   { return w.open }
func (w *Window) IsClosed() bool { return !w.open }

func (w *Window) Style() func() {
	return func() {}
}

func (w *Window) Build() {
	if !w.open {
		return
	}

	var styleFin func()
	if w.layoutBuilder != nil {
		styleFin = w.layoutBuilder.Style()
	} else {
		styleFin = w.Style()
	}
	defer styleFin()

	if imgui.BeginV(w.Title(), &w.open, w.config.Combined()) {
		// Run layout builder's Menu function
		if w.layoutBuilder != nil {
			w.layoutBuilder.Menu()
		} else {
			w.Menu()
		}

		// Run layout builder's Layout function
		if w.layoutBuilder != nil {
			w.layoutBuilder.Layout()
		} else {
			w.Layout()
		}

	}
	imgui.End()
}

func (w *Window) Menu() {}
func (w *Window) Layout() {
	panic("Layout not implemented. Please check that layoutBuilder is set.")
}

// NewWindow creates a new Window instance with the provided title and icon
func NewWindow(title, icon string, config *WindowConfig) *Window {
	if config == nil {
		config = NewWindowConfig()
	}

	return &Window{
		noClose:      false,
		collapsed:    false,
		open:         true,
		title:        title,
		icon:         icon,
		config:       config,
		WindowEvents: signals.New[events.WindowEventRecord](),
	}
}
