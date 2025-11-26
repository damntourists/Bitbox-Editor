package events

import "github.com/AllenDang/cimgui-go/imgui"

const (
	ComponentClickEventKey    = "component.mouse.click"
	ComponentHoverInEventKey  = "component.mouse.hover.in"
	ComponentHoverOutEventKey = "component.mouse.hover.out"
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

// ComponentClickEvent is published when a component is clicked or double-clicked.
type ComponentClickEvent struct {
	// IsDoubleClick indicates if this was a double-click event
	IsDoubleClick bool
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
func (e ComponentClickEvent) Type() string {
	return ComponentClickEventKey
}

// ComponentHoverInEvent is published when a component is hovered.
type ComponentHoverInEvent struct {
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
func (e ComponentHoverInEvent) Type() string {
	return ComponentHoverInEventKey
}

// ComponentHoverOutEvent is published when a component is no longer hovered.
type ComponentHoverOutEvent struct {
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
func (e ComponentHoverOutEvent) Type() string {
	return ComponentHoverOutEventKey
}
