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
	Window struct {
		open      bool
		title     string
		subtitle  string
		icon      string
		noClose   bool
		collapsed bool

		WindowEvents signals.Signal[events.WindowEventRecord]

		flags imgui.WindowFlags

		layoutBuilder types.WindowLayoutBuilder

		loading bool
	}
)

func (w *Window) Flags() imgui.WindowFlags { return w.flags }
func (w *Window) SetFlags(flags imgui.WindowFlags) *Window {
	w.flags = flags
	return w
}

func (w *Window) Title() string { return fmt.Sprintf("%s %s", w.Icon(), w.title) }
func (w *Window) Icon() string  { return fonts.Icon(w.icon) }
func (w *Window) Close() {
	w.open = false
	w.WindowEvents.Emit(context.Background(), events.WindowEventRecord{
		Type:        events.WindowClose,
		WindowTitle: w.Title(),
	})
}
func (w *Window) SetLayoutBuilder(layout types.WindowLayoutBuilder) *Window {
	w.layoutBuilder = layout
	return w
}

func (w *Window) Destroy() {
	if w.open {
		w.Close()
	}
	w.WindowEvents.Emit(context.Background(), events.WindowEventRecord{
		Type:        events.WindowDestroy,
		WindowTitle: w.Title(),
	})
	w.WindowEvents.Reset()
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

	if imgui.BeginV(w.Title(), &w.open, w.flags) {
		if w.layoutBuilder != nil {
			w.layoutBuilder.Menu()
		} else {
			w.Menu()
		}

		if w.layoutBuilder != nil {
			w.layoutBuilder.Layout()
		} else {
			w.Layout()
		}

	}
	imgui.End()

	if !w.open {
		w.WindowEvents.Emit(context.Background(), events.WindowEventRecord{
			Type:        events.WindowClose,
			WindowTitle: w.Title(),
		})
	}
}

func (w *Window) Menu() {}
func (w *Window) Layout() {
	panic("Layout not implemented. Please check that layoutBuilder is set.")
}

func NewWindow(title, icon string) *Window {
	//if config == nil {
	//	config = NewWindowConfig()
	//}

	return &Window{
		noClose:   false,
		collapsed: false,
		open:      true,
		title:     title,
		icon:      icon,
		//config:       config,
		flags:        imgui.WindowFlagsNone,
		WindowEvents: signals.New[events.WindowEventRecord](),
	}
}
