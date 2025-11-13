package events

import "github.com/AllenDang/cimgui-go/imgui"

const (
	ComponentClickEventKey    = "component.mouse.click"
	ComponentHoverInEventKey  = "component.mouse.hover.in"
	ComponentHoverOutEventKey = "component.mouse.hover.out"
)

// ComponentEventType is the enum for all component-level UI interactions.
type ComponentEventType int32

const (
	ComponentClickedEvent ComponentEventType = iota
	ComponentDoubleClickedEvent
	ComponentHoverInEvent
	ComponentHoverOutEvent
)

// ItemState is a bitmask for a component UI state
type ItemState uint32

const (
	ItemStateNone ItemState = 1 << iota
	ItemStateHovered
	ItemStateActive
	ItemStateFocused
	ItemStateHoverIn
	ItemStateHoverOut
	ItemStateActiveIn
	ItemStateActiveOut
	ItemStateFocusIn
	ItemStateFocusOut
)

// HasState checks if the given state bit is set
func (is ItemState) HasState(state ItemState) bool {
	return is&state != 0
}

type MouseButton int32

const (
	MouseButtonNone MouseButton = iota
	MouseButtonLeft
	MouseButtonRight
	MouseButtonMiddle
)

// MouseEventRecord is published by a component when a mouse event occurs.
type MouseEventRecord struct {
	// EventType is the enum value (e.g., ComponentClickedEvent).
	EventType ComponentEventType

	// ImguiID is the internal ImGui ID of the component.
	ImguiID imgui.ID

	// UUID is the event-bus-safe unique ID of the component.
	UUID string

	// The mouse button that was used for this event
	Button MouseButton

	// The state of the component when the event fired
	State ItemState

	// The component's drag-drop data, if any
	Data interface{}
}

// Type implements the events.Event interface
func (e MouseEventRecord) Type() string {
	// It switches on the *enum*, not the string
	switch e.EventType {
	case ComponentClickedEvent,
		ComponentDoubleClickedEvent:
		return ComponentClickEventKey
	case ComponentHoverInEvent:
		return ComponentHoverInEventKey
	case ComponentHoverOutEvent:
		return ComponentHoverOutEventKey
	// TODO: Add more event types. (e.g., drag, active)
	default:
		return "component.mouse.unknown"
	}
}
