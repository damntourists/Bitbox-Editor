package events

type WindowEventType int32

const (
	WindowOpen    WindowEventType = 0
	WindowClose   WindowEventType = 1
	WindowDestroy WindowEventType = 1
)

type WindowEventRecord struct {
	Type        WindowEventType
	WindowTitle string
}
