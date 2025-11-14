package window

/*
┍━━━━━━━━━━━━━━╳┑
│ Window (Base) │
└───────────────┘
*/

import (
	cmp "bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/app/font"
	"bitbox-editor/internal/logging"
	"fmt"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var log = logging.NewLogger("window")

var WindowEventChannel = make(chan events.WindowEventRecord, 100)

type (
	Window[T WindowType] struct {
		uuid      string
		collapsed bool
		focused   bool
		loading   bool
		noClose   bool
		open      bool
		destroyed bool

		icon     string
		title    string
		subtitle string
		suffix   string

		flags imgui.WindowFlags

		layoutBuilder WindowLayoutBuilder

		updates chan UpdateCmd
		handler UpdateHandlerFunc

		// Command handler registry for custom commands
		commandHandlers map[any]func(UpdateCmd)
	}
)

func NewWindow[T WindowType](title, icon string, handler UpdateHandlerFunc) *Window[T] {
	return &Window[T]{
		uuid:            uuid.NewString(),
		noClose:         false,
		collapsed:       false,
		open:            true,
		title:           title,
		icon:            icon,
		flags:           imgui.WindowFlagsNone,
		updates:         make(chan cmp.UpdateCmd, 100),
		handler:         handler,
		commandHandlers: make(map[any]func(UpdateCmd)),
	}
}

func (w *Window[T]) ProcessUpdates() {
	if w.handler == nil && len(w.commandHandlers) == 0 {
		for {
			select {
			case <-w.updates:
				// discard update
			default:
				return
			}
		}
	}

	const maxMessagesPerFrame = 100
	processedCount := 0

	for processedCount < maxMessagesPerFrame {
		select {
		case cmd := <-w.updates:
			// Try registered handlers first
			if handler, ok := w.commandHandlers[cmd.Type]; ok {
				handler(cmd)
			} else if w.handler != nil {
				// Fall back to the legacy handler for backward compatibility
				w.handler(cmd)
			}
			processedCount++
		default:
			return
		}
	}

	if len(w.updates) > 0 {
		log.Warn("Window ProcessUpdates hit message limit", zap.String("windowTitle", w.title))
	}
}

func (w *Window[T]) HandleGlobalUpdate(cmd UpdateCmd) bool {
	switch cmd.Type {

	case CmdSetLoading:
		w.loading = cmd.Data.(bool)
		return true

	case CmdWinSetTitle:
		w.title = cmd.Data.(string)
		return true

	case CmdWinSetIcon:
		w.icon = cmd.Data.(string)
		return true

	case CmdWinSetSuffix:
		w.suffix = cmd.Data.(string)
		return true

	case CmdWinSetNoClose:
		w.noClose = cmd.Data.(bool)
		return true

	case CmdWinSetCollapsed:
		w.collapsed = cmd.Data.(bool)
		return true

	case CmdWinSetFlags:
		w.flags = cmd.Data.(imgui.WindowFlags)
		return true

	case CmdWinSetOpen:
		w.open = cmd.Data.(bool)
		record := events.WindowEventRecord{
			EventType:   events.WindowOpenEvent,
			WindowTitle: w.Title(),
			WindowID:    w.uuid,
		}
		if !w.open {
			record.EventType = events.WindowCloseEvent
		}

		eventbus.Bus.Publish(record)

		return true

	case CmdWinDestroy:
		w.destroyed = cmd.Data.(bool)
		record := events.WindowEventRecord{
			EventType:   events.WindowDestroyEvent,
			WindowTitle: w.Title(),
		}
		select {
		case WindowEventChannel <- record:
		default:
			// Channel full, drop event
		}

		return true

	default:
		return false
	}
}

func (w *Window[T]) UpdateChannel() chan<- cmp.UpdateCmd { // Add UpdateChannel
	return w.updates
}

// RegisterCommandHandler registers a handler for a specific command type.
// Returns the window for method chaining.
func (w *Window[T]) RegisterCommandHandler(cmdType any, handler func(UpdateCmd)) *Window[T] {
	if w.commandHandlers == nil {
		w.commandHandlers = make(map[any]func(UpdateCmd))
	}
	w.commandHandlers[cmdType] = handler
	return w
}

// UnregisterCommandHandler removes a handler for a specific command type.
// Returns the window for method chaining.
func (w *Window[T]) UnregisterCommandHandler(cmdType any) *Window[T] {
	if w.commandHandlers != nil {
		delete(w.commandHandlers, cmdType)
	}
	return w
}

// HasCommandHandler checks if a handler is registered for a command type.
func (w *Window[T]) HasCommandHandler(cmdType any) bool {
	if w.commandHandlers == nil {
		return false
	}
	_, ok := w.commandHandlers[cmdType]
	return ok
}

func (w *Window[T]) SendUpdate(cmd UpdateCmd) {
	select {
	case w.updates <- cmd:
		// Command sent successfully
	default:
		// Channel is full, drop the command to avoid blocking.
		log.Warn("Window update channel full, dropping command",
			zap.String("windowTitle", w.title),
			zap.Any("cmdType", cmd.Type))
	}
}

func (w *Window[T]) UUID() string { return w.uuid }

func (w *Window[T]) Collapsed() bool { return w.collapsed }

func (w *Window[T]) Focused() bool { return w.focused }

func (w *Window[T]) Loading() bool { return w.loading }

func (w *Window[T]) NoClose() bool { return w.noClose }

func (w *Window[T]) Flags() imgui.WindowFlags { return w.flags }

func (w *Window[T]) Title() string {
	if w.suffix != "" {
		return fmt.Sprintf("%s %s %s", w.Icon(), w.title, w.suffix)
	}
	return fmt.Sprintf("%s %s", w.Icon(), w.title)
}

func (w *Window[T]) Icon() string { return font.Icon(w.icon) }

func (w *Window[T]) Destroy() {
	w.open = false
	w.destroyed = true
}

func (w *Window[T]) IsOpen() bool { return w.open }

func (w *Window[T]) IsClosed() bool { return !w.open }

func (w *Window[T]) SetTitle(title string) *Window[T] {
	w.SendUpdate(UpdateCmd{Type: CmdWinSetTitle, Data: title})
	return w
}

func (w *Window[T]) SetIcon(icon string) *Window[T] {
	w.SendUpdate(UpdateCmd{Type: CmdWinSetIcon, Data: icon})
	return w
}

func (w *Window[T]) SetNoClose(v bool) *Window[T] {
	w.SendUpdate(UpdateCmd{Type: CmdWinSetNoClose, Data: v})
	return w
}

func (w *Window[T]) SetCollapsed(v bool) *Window[T] {
	w.SendUpdate(UpdateCmd{Type: CmdWinSetCollapsed, Data: v})
	return w
}

func (w *Window[T]) SetFlags(flags imgui.WindowFlags) *Window[T] {
	w.SendUpdate(UpdateCmd{Type: CmdWinSetFlags, Data: flags})
	return w
}

func (w *Window[T]) SetOpen() {
	w.SendUpdate(UpdateCmd{Type: CmdWinSetOpen, Data: true})
}

func (w *Window[T]) SetClose() {
	w.SendUpdate(UpdateCmd{Type: CmdWinSetOpen, Data: false})
}

func (w *Window[T]) ToggleOpen() {
	w.open = !w.open
	w.SendUpdate(UpdateCmd{Type: CmdWinSetOpen, Data: w.open})
}

func (w *Window[T]) SetDestroyed() {
	w.SendUpdate(UpdateCmd{Type: CmdWinDestroy, Data: true})
}

func (w *Window[T]) SetLayoutBuilder(layout WindowLayoutBuilder) *Window[T] {
	w.layoutBuilder = layout
	return w
}

func (w *Window[T]) Build() {
	w.ProcessUpdates()

	open := w.open
	flags := w.flags
	title := w.Title()

	if !open {
		w.focused = false
		return
	}

	// Apply layout builder's style if it exists, otherwise use the default style
	var styleFin func()
	if w.layoutBuilder != nil {
		styleFin = w.layoutBuilder.Style()
	} else {
		styleFin = w.Style()
	}
	defer styleFin()

	isOpenPtr := &open
	if imgui.BeginV(title, isOpenPtr, flags) {
		w.focused = imgui.IsWindowFocused()
		w.collapsed = imgui.IsWindowCollapsed()

		// Build layout builder's menu if it exists, otherwise use the default menu
		if w.layoutBuilder != nil {
			w.layoutBuilder.Menu()
		} else {
			w.Menu()
		}

		// Build layout builder's layout if it exists, otherwise use the default menu
		if w.layoutBuilder != nil {
			w.layoutBuilder.Layout()
		} else {
			w.Layout()
		}
	} else {
		// If BeginV returns false, the window is not being rendered (minimized/closed)
		// Definitely not focused
		w.focused = false
	}

	imgui.End()

	if !*isOpenPtr && w.open {
		w.SetClose()
	}
}

func (w *Window[T]) Style() func() {
	return func() { /* *crickets* */ }
}

func (w *Window[T]) Menu() { /* *crickets* */ }

func (w *Window[T]) Layout() {
	panic("Layout() not implemented. Did you run SetLayoutBuilder()?")
}
