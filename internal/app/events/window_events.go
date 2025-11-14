package events

type WindowEvent int32

// Window Event Enums
const (
	WindowOpenEvent WindowEvent = iota
	WindowCloseEvent
	WindowDestroyEvent
)

// Window Event Keys
const (
	WindowOpenEventKey    = "window.open"
	WindowCloseEventKey   = "window.close"
	WindowDestroyEventKey = "window.destroy"
)

// WindowEventRecord holds data for a window event.
type WindowEventRecord struct {
	// EventType is the enum value (e.g., WindowOpenEvent).
	EventType WindowEvent

	// WindowTitle is the source window title which produced this event.
	WindowTitle string

	// WindowID is the source window ID which produced this event.
	WindowID string
}

// Type returns the event type key.
func (e WindowEventRecord) Type() string {
	switch e.EventType {
	case WindowOpenEvent:
		return WindowOpenEventKey
	case WindowCloseEvent:
		return WindowCloseEventKey
	case WindowDestroyEvent:
		return WindowDestroyEventKey
	default:
		return "window.unknown"
	}
}
