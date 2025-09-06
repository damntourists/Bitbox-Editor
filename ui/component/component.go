package component

import (
	events2 "bitbox-editor/ui/events"
	"context"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maniartech/signals"
)

const (
	// Auto is used to widget.Size to indicate height or width to occupy available spaces.
	Auto float32 = -1
)

var ComponentRegistry = make(map[ID]*Component)

type ID string

func (i ID) String() string {
	return string(i)
}

type ComponentType interface {
	Layout()
}

// Component - Base Backend component
type Component struct {
	id    ID
	srcID ID

	state         events2.State
	previousState events2.State

	MouseEvents signals.Signal[events2.MouseEventRecord]
}

func (c *Component) Layout() {
	panic("implement me")
}
func (c *Component) SourceID(src ID) *Component {
	c.srcID = src
	return c
}
func (c *Component) ID() ID {
	return c.id
}

func (c *Component) HandleMouseEvents() {
	var (
		click  events2.ClickType
		button events2.MouseButton

		clickSet  bool
		buttonSet bool
	)

	// store previous state
	c.previousState = c.state

	// clear current state
	c.state = events2.StateNone

	// Check if the item is hovered, active, or focused and set state accordingly
	if imgui.IsItemHovered() {
		c.state |= events2.StateHovered
		if !c.previousState.HasState(events2.StateHovered) {
			c.state |= events2.StateHoverIn
		}
	} else {
		if c.previousState.HasState(events2.StateHovered) {
			c.state |= events2.StateHoverOut
		}
	}

	if imgui.IsItemActive() {
		c.state |= events2.StateActive
		if !c.previousState.HasState(events2.StateActive) {
			c.state |= events2.StateActiveIn
		}
	} else {
		if c.previousState.HasState(events2.StateActive) {
			c.state |= events2.StateActiveOut
		}
	}

	if imgui.IsItemFocused() {
		c.state |= events2.StateFocused
		if !c.previousState.HasState(events2.StateFocused) {
			c.state |= events2.StateFocusIn
		}
	} else {
		if c.previousState.HasState(events2.StateFocused) {
			c.state |= events2.StateFocusOut
		}
	}

	buttonTypes := map[imgui.MouseButton]events2.MouseButton{
		imgui.MouseButtonLeft:   events2.MouseButtonLeft,
		imgui.MouseButtonRight:  events2.MouseButtonRight,
		imgui.MouseButtonMiddle: events2.MouseButtonMiddle,
	}
	for bt, ebt := range buttonTypes {
		if imgui.IsItemClickedV(bt) {
			button = ebt
			buttonSet = true

			click = events2.Clicked
			clickSet = true

			if imgui.IsMouseDoubleClicked(bt) {
				click = events2.DoubleClicked
				clickSet = true
			}
		}
	}

	// Check if any of state, click, or button were set
	if c.previousState != c.state || buttonSet || clickSet {
		// Create a new event record
		event := events2.MouseEventRecord{
			ID:     c.id.String(),
			Type:   click,
			State:  c.state,
			Button: button,
			GUID:   "src id here",
		}

		// Emit the event because it's different
		c.MouseEvents.Emit(context.Background(), event)
	}
}

func NewComponent(id ID) *Component {
	if _, ok := ComponentRegistry[id]; !ok {
		ComponentRegistry[id] = &Component{
			id:            id,
			state:         events2.StateNone,
			previousState: events2.StateNone,
			MouseEvents:   signals.New[events2.MouseEventRecord](), //events.ComponentMouseEvent,
		}
	}

	return ComponentRegistry[id]
}
