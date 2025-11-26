package events

// Window Event Keys
const (
	WindowOpenEventKey    = "window.open"
	WindowCloseEventKey   = "window.close"
	WindowDestroyEventKey = "window.destroy"
)

// WindowOpenEvent is published when a window is opened.
type WindowOpenEvent struct {
	// WindowTitle is the title of the window that was opened.
	WindowTitle string
	// WindowID is the unique identifier for the window.
	WindowID string
}

func (e WindowOpenEvent) Type() string {
	return WindowOpenEventKey
}

// WindowCloseEvent is published when a window is closed.
type WindowCloseEvent struct {
	// WindowTitle is the title of the window that was closed.
	WindowTitle string
	// WindowID is the unique identifier for the window.
	WindowID string
}

func (e WindowCloseEvent) Type() string {
	return WindowCloseEventKey
}

// WindowDestroyEvent is published when a window is destroyed.
type WindowDestroyEvent struct {
	// WindowTitle is the title of the window that was destroyed.
	WindowTitle string
	// WindowID is the unique identifier for the window.
	WindowID string
}

func (e WindowDestroyEvent) Type() string {
	return WindowDestroyEventKey
}
