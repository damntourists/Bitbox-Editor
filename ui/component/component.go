package component

import (
	"bitbox-editor/lib/logging"
	"bitbox-editor/ui/component/state"
	"bitbox-editor/ui/events"
	"bitbox-editor/ui/types"
	"context"
	"strconv"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/maniartech/signals"
	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	log = logging.NewLogger("cmp")
}

type ComponentType interface {
	Layout()
}

type Component struct {
	id imgui.ID

	state         state.ItemState
	previousState state.ItemState

	data interface{}

	mouseEvents signals.Signal[events.MouseEventRecord]

	layoutBuilder types.ComponentLayoutBuilder
}

func (c *Component) MouseEvents() signals.Signal[events.MouseEventRecord] {
	return c.mouseEvents
}

func (c *Component) SetData(data interface{}) *Component {
	c.data = data
	return c
}

func (c *Component) SetLayoutBuilder(layoutBuilder types.ComponentLayoutBuilder) *Component {
	c.layoutBuilder = layoutBuilder
	return c
}

func (c *Component) Data() interface{} {
	return c.data
}

func (c *Component) ID() imgui.ID {
	return c.id
}

func (c *Component) IDStr() string {
	return strconv.Itoa(int(c.id))
}

func (c *Component) Layout() {
	panic("not implemented")
}

func (c *Component) handleMouseEvents() {
	var (
		click  events.ClickType
		button events.MouseButton

		clickSet  bool
		buttonSet bool
	)

	// store previous state
	c.previousState = c.state

	// clear current state
	c.state = state.ItemStateNone

	// Check if the item is hovered, active, or focused and set state accordingly
	if imgui.IsItemHovered() {
		c.state |= state.ItemStateHovered
		if !c.previousState.HasState(state.ItemStateHovered) {
			c.state |= state.ItemStateHoverIn
		}
	} else {
		if c.previousState.HasState(state.ItemStateHovered) {
			c.state |= state.ItemStateHoverOut
		}
	}

	if imgui.IsItemActive() {
		c.state |= state.ItemStateActive
		if !c.previousState.HasState(state.ItemStateActive) {
			c.state |= state.ItemStateActiveIn
		}
	} else {
		if c.previousState.HasState(state.ItemStateActive) {
			c.state |= state.ItemStateActiveOut
		}
	}

	if imgui.IsItemFocused() {
		c.state |= state.ItemStateFocused
		if !c.previousState.HasState(state.ItemStateFocused) {
			c.state |= state.ItemStateFocusIn
		}
	} else {
		if c.previousState.HasState(state.ItemStateFocused) {
			c.state |= state.ItemStateFocusOut
		}
	}

	// Iterate through mouse buttons and update button/click
	buttonTypes := []imgui.MouseButton{
		imgui.MouseButtonLeft,
		imgui.MouseButtonRight,
		imgui.MouseButtonMiddle,
	}
	for _, bt := range buttonTypes {
		if imgui.IsItemClickedV(bt) {
			button = events.MouseButton(bt)
			buttonSet = true

			click = events.Clicked
			clickSet = true

			if imgui.IsMouseDoubleClicked(bt) {
				click = events.DoubleClicked
				clickSet = true
			}
		}
	}

	// Check if any of state, click, or button were set
	if c.previousState != c.state || buttonSet || clickSet {
		// Create a new event record
		event := events.MouseEventRecord{
			ID:     c.ID(),
			Type:   click,
			State:  c.state,
			Button: button,
			Data:   c.Data,
		}

		// Emit the event because it's different
		c.mouseEvents.Emit(context.Background(), event)
	}
}

func NewComponent(id imgui.ID) *Component {
	return &Component{
		id:            id,
		state:         state.ItemStateNone,
		previousState: state.ItemStateNone,
		data:          nil,
		mouseEvents:   signals.New[events.MouseEventRecord](),
	}
}
