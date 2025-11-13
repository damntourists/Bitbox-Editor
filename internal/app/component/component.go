/*
Package component provides component implementations and base types for building UI components.
*/

package component

import (
	"bitbox-editor/internal/app/animation"
	"bitbox-editor/internal/app/dragdrop"
	"bitbox-editor/internal/app/eventbus"
	"bitbox-editor/internal/app/events"
	"bitbox-editor/internal/logging"
	"fmt"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/AllenDang/cimgui-go/implot"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TODO: Finish documenting this

const maxMessagesPerFrame = 100

var log *zap.Logger

func init() {
	log = logging.NewLogger("cmp")
}

type animStateData struct {
	startTime   time.Time
	duration    time.Duration
	startVal    imgui.Vec4
	endVal      imgui.Vec4
	currentVal  imgui.Vec4
	isAnimating bool
	easingFn    animation.EasingFunc
}

type Component[T ComponentType] struct {
	uuid string
	id   imgui.ID

	// State
	state         events.ItemState
	previousState events.ItemState

	// Drag drop
	dragDropData      interface{}
	dragDropType      string
	dragDropTooltipFn dragdrop.TooltipFunc

	animState map[UpdateCmdType]*animStateData
	animMutex sync.RWMutex

	// UI Update request channel
	updates chan UpdateCmd
	handler UpdateHandlerFunc

	// Command handler registry for custom commands
	commandHandlers map[any]func(UpdateCmd)

	// common properties
	text string
	icon string
	// style
	bgActiveColor   imgui.Vec4
	bgHoverColor    imgui.Vec4
	bgSelectedColor imgui.Vec4
	fgColor         imgui.Vec4
	bgColor         imgui.Vec4
	textColor       imgui.Vec4
	outlineColor    imgui.Vec4
	borderColor     imgui.Vec4
	// sizing
	size         imgui.Vec2
	width        float32
	height       float32
	borderSize   float32
	rounding     float32
	padding      float32
	outlineWidth float32
	// progress related
	progress       float32
	progressBg     imgui.Vec4
	progressFg     imgui.Vec4
	progressHeight float32
	// toggles
	editing   bool
	enabled   bool
	selected  bool
	collapsed bool
	outline   bool
	loading   bool
	// animation control
	hoverAnimationsDisabled bool
	// interactivity
	clickable bool
	// misc
	layoutBuilder LayoutBuilderType
	colormap      implot.Colormap
}

func NewComponent[T ComponentType](id imgui.ID, handler UpdateHandlerFunc) *Component[T] {
	return &Component[T]{
		id:              id,
		uuid:            uuid.NewString(),
		state:           events.ItemStateNone,
		previousState:   events.ItemStateNone,
		dragDropData:    nil,
		updates:         make(chan UpdateCmd, 500),
		handler:         handler,
		commandHandlers: make(map[any]func(UpdateCmd)),
		animState:       make(map[UpdateCmdType]*animStateData),
		enabled:         true,
	}
}

func (c *Component[T]) Width() float32 {
	return c.width
}

func (c *Component[T]) Height() float32 { // <-- ADDED
	return c.height
}

func (c *Component[T]) Size() imgui.Vec2 {
	return c.size
}

func (c *Component[T]) Padding() float32 {
	return c.padding
}

func (c *Component[T]) Rounding() float32 {
	return c.rounding
}

func (c *Component[T]) Outline() bool {
	return c.outline
}

func (c *Component[T]) OutlineWidth() float32 {
	return c.outlineWidth
}

func (c *Component[T]) OutlineColor() imgui.Vec4 {
	return c.outlineColor
}

func (c *Component[T]) BorderSize() float32 {
	return c.borderSize
}

func (c *Component[T]) BgHoverColor() imgui.Vec4 {
	return c.bgHoverColor
}

func (c *Component[T]) BgActiveColor() imgui.Vec4 {
	return c.bgActiveColor
}

func (c *Component[T]) BgSelectedColor() imgui.Vec4 {
	return c.bgSelectedColor
}

func (c *Component[T]) TextColor() imgui.Vec4 {
	return c.textColor
}

func (c *Component[T]) BorderColor() imgui.Vec4 {
	return c.borderColor
}

func (c *Component[T]) Progress() float32 {
	return c.progress
}

func (c *Component[T]) ProgressHeight() float32 {
	return c.progressHeight
}

func (c *Component[T]) ProgressBg() imgui.Vec4 {
	return c.progressBg
}

func (c *Component[T]) ProgressFg() imgui.Vec4 {
	return c.progressFg
}

func (c *Component[T]) Selected() bool {
	return c.selected
}

func (c *Component[T]) Editing() bool {
	return c.editing
}

func (c *Component[T]) Data() interface{} {
	return c.dragDropData
}

func (c *Component[T]) ID() imgui.ID {
	return c.id
}

func (c *Component[T]) IDStr() string {
	return strconv.Itoa(int(c.id))
}

func (c *Component[T]) UUID() string {
	return c.uuid
}

func (c *Component[T]) Text() string { return c.text }

func (c *Component[T]) Enabled() bool {
	return c.enabled
}

func (c *Component[T]) State() events.ItemState {
	return c.state
}

// RegisterCommandHandler registers a handler for a specific command type.
func (c *Component[T]) RegisterCommandHandler(cmdType any, handler func(UpdateCmd)) *Component[T] {
	if c.commandHandlers == nil {
		c.commandHandlers = make(map[any]func(UpdateCmd))
	}
	c.commandHandlers[cmdType] = handler
	return c
}

// UnregisterCommandHandler removes a handler for a specific command type.
func (c *Component[T]) UnregisterCommandHandler(cmdType any) *Component[T] {
	if c.commandHandlers != nil {
		delete(c.commandHandlers, cmdType)
	}
	return c
}

// HasCommandHandler checks if a handler is registered for a command type.
func (c *Component[T]) HasCommandHandler(cmdType any) bool {
	if c.commandHandlers == nil {
		return false
	}
	_, ok := c.commandHandlers[cmdType]
	return ok
}

// SendUpdate provides a way to send a command to the update channel.
func (c *Component[T]) SendUpdate(cmd UpdateCmd) {
	select {
	case c.updates <- cmd:
	default:
		log.Warn("Component update channel full, dropping command",
			zap.String("uuid", c.uuid),
			zap.Any("cmdType", cmd.Type))
	}
}

// ProcessUpdates drains the component's update channel and calls its handler.
func (c *Component[T]) ProcessUpdates() {
	if c.handler == nil && len(c.commandHandlers) == 0 {
		for {
			select {
			case <-c.updates:
			default:
				return
			}
		}
	}

	// Limit the number of messages processed per frame
	for i := 0; i < maxMessagesPerFrame; i++ {
		select {
		case cmd := <-c.updates:
			// Try registered handlers first
			if handler, ok := c.commandHandlers[cmd.Type]; ok {
				handler(cmd)
			} else if c.handler != nil {
				// TODO: Update all other components to use new method above, method below is legacy
				c.handler(cmd)
			}
		default:
			// Channel is empty, stop processing
			return
		}
	}

	if len(c.updates) > 0 {
		log.Warn("Component ProcessUpdates hit message limit", zap.String("uuid", c.uuid))
	}
}

func (c *Component[T]) ProcessAnimations() {
	c.animMutex.Lock()
	defer c.animMutex.Unlock()

	now := time.Now()
	for _, state := range c.animState {
		if !state.isAnimating {
			continue
		}

		elapsed := now.Sub(state.startTime)

		t := float64(elapsed) / float64(state.duration)

		if t >= 1.0 {
			// Animation finished, Snap to final value
			state.currentVal = state.endVal
			state.isAnimating = false
		} else {
			// Animation in progress.
			easingFn := state.easingFn
			if easingFn == nil {
				easingFn = animation.EaseOutQuad
			}

			// Apply easing
			easedT := easingFn(t)

			// TODO: Update to handle different value types. Currently only Vec4
			state.currentVal = animation.LerpVec4(state.startVal, state.endVal, float32(easedT))
		}
	}
}

// GetAnimatedBgColor returns the current animated background color, or base color if not animating
func (c *Component[T]) GetAnimatedBgColor() imgui.Vec4 {
	c.animMutex.RLock()
	defer c.animMutex.RUnlock()

	if state, exists := c.animState[CmdSetBgColor]; exists {
		return state.currentVal
	}
	return c.bgColor
}

// StartVec4Animation begins a color/vec4 animation
func (c *Component[T]) StartVec4Animation(
	cmdType UpdateCmdType,
	targetValue imgui.Vec4,
	duration time.Duration,
	easeFn animation.EasingFunc) {
	c.animMutex.Lock()
	defer c.animMutex.Unlock()

	if easeFn == nil {
		easeFn = animation.EaseOutQuad
	}

	state, exists := c.animState[cmdType]
	if !exists {
		// Initialize new animation state with current value
		state = &animStateData{
			currentVal: targetValue,
			easingFn:   easeFn,
		}
		c.animState[cmdType] = state
	}

	// If already animating to this value, do nothing
	if state.isAnimating && state.endVal == targetValue {
		return
	}

	// Start a new animation
	state.startTime = time.Now()
	state.duration = duration
	state.startVal = state.currentVal
	state.endVal = targetValue
	state.isAnimating = true

	if state.easingFn == nil {
		state.easingFn = easeFn
	}
}

func (c *Component[T]) UpdateChannel() chan<- UpdateCmd {
	return c.updates
}

// SetEasing configures the easing function for a specific animation type
func (c *Component[T]) SetEasing(cmdType UpdateCmdType, easingFn animation.EasingFunc) {
	// TODO: Maybe not needed. Easing function is set per animation in StartVec4Animation
	c.animMutex.Lock()
	defer c.animMutex.Unlock()

	state, exists := c.animState[cmdType]
	if !exists {
		// Create a state if it doesn't exist
		state = &animStateData{
			easingFn: easingFn,
		}
		c.animState[cmdType] = state
	} else {
		state.easingFn = easingFn
	}
}

func (c *Component[T]) SetBgActiveColor(color imgui.Vec4) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetBgActiveColor, Data: color})
	return c
}

func (c *Component[T]) SetBgColor(color imgui.Vec4) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetBgColor, Data: color})
	return c
}

func (c *Component[T]) SetBgHoveredColor(color imgui.Vec4) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetBgHoveredColor, Data: color})
	return c
}

func (c *Component[T]) SetBgSelectedColor(color imgui.Vec4) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetBgSelectedColor, Data: color})
	return c
}

func (c *Component[T]) SetFgColor(color imgui.Vec4) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetFgColor, Data: color})
	return c
}

func (c *Component[T]) SetBorderColor(color imgui.Vec4) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetBorderColor, Data: color})
	return c
}

func (c *Component[T]) SetBorderSize(size float32) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetBorderSize, Data: size})
	return c
}

func (c *Component[T]) SetCollapsed(collapsed bool) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetCollapsed, Data: collapsed})
	return c
}

func (c *Component[T]) SetHeight(height float32) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetHeight, Data: height})
	return c
}

func (c *Component[T]) SetProgress(progress float32) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetProgress, Data: progress})
	return c
}

func (c *Component[T]) SetProgressBgColor(color imgui.Vec4) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetProgressBgColor, Data: color})
	return c
}

func (c *Component[T]) SetProgressFgColor(color imgui.Vec4) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetProgressFgColor, Data: color})
	return c
}

func (c *Component[T]) SetProgressHeight(height float32) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetProgressHeight, Data: height})
	return c
}

func (c *Component[T]) SetRounding(rounding float32) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetRounding, Data: rounding})
	return c
}

func (c *Component[T]) SetPadding(padding float32) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetPadding, Data: padding})
	return c
}

func (c *Component[T]) SetWidth(width float32) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetWidth, Data: width})
	return c
}

func (c *Component[T]) SetSize(size imgui.Vec2) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetSize, Data: size})
	return c
}

func (c *Component[T]) SetEditing(editing bool) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetEditing, Data: editing})
	return c
}

func (c *Component[T]) SetEnabled(enabled bool) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetEnabled, Data: enabled})
	return c
}

func (c *Component[T]) SetSelected(selected bool) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetSelected, Data: selected})
	return c
}

func (c *Component[T]) SetOutline(outline bool) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetOutline, Data: outline})
	return c
}

func (c *Component[T]) SetOutlineColor(color imgui.Vec4) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetOutlineColor, Data: color})
	return c
}

func (c *Component[T]) SetOutlineWidth(width float32) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetOutlineWidth, Data: width})
	return c
}

func (c *Component[T]) SetText(text string) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetText, Data: text})
	return c
}

func (c *Component[T]) SetTextColor(color imgui.Vec4) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetTextColor, Data: color})
	return c
}

func (c *Component[T]) SetDragDropData(dataType string, data interface{}) *Component[T] {
	payload := dragdrop.DataPayload{
		Type: dataType,
		Data: data,
	}
	c.SendUpdate(UpdateCmd{Type: CmdSetDragDropData, Data: payload})
	return c
}

func (c *Component[T]) SetDragDropTooltipFn(tooltipFn dragdrop.TooltipFunc) *Component[T] {
	c.SendUpdate(UpdateCmd{Type: CmdSetDragDropTooltipFn, Data: tooltipFn})
	return c
}

// SetLayoutBuilder does not run in the update channel. It is intended to be called from the constructor.
func (c *Component[T]) SetLayoutBuilder(layoutBuilder LayoutBuilderType) *Component[T] {
	c.layoutBuilder = layoutBuilder
	return c
}

func (c *Component[T]) SetClickable(clickable bool) *Component[T] {
	c.clickable = clickable
	return c
}

func (c *Component[T]) handleMouseEvents() {
	var (
		button   events.MouseButton
		clickSet bool
	)

	c.previousState = c.state
	c.state = events.ItemStateNone

	if imgui.IsItemHovered() {
		c.state |= events.ItemStateHovered
		if !c.previousState.HasState(events.ItemStateHovered) {
			c.state |= events.ItemStateHoverIn
		}
	} else {
		if c.previousState.HasState(events.ItemStateHovered) {
			c.state |= events.ItemStateHoverOut
		}
	}

	if imgui.IsItemActive() {
		c.state |= events.ItemStateActive
		if !c.previousState.HasState(events.ItemStateActive) {
			c.state |= events.ItemStateActiveIn
		}
	} else {
		if c.previousState.HasState(events.ItemStateActive) {
			c.state |= events.ItemStateActiveOut
		}
	}

	if imgui.IsItemFocused() {
		c.state |= events.ItemStateFocused
		if !c.previousState.HasState(events.ItemStateFocused) {
			c.state |= events.ItemStateFocusIn
		}
	} else {
		if c.previousState.HasState(events.ItemStateFocused) {
			c.state |= events.ItemStateFocusOut
		}
	}

	buttonTypes := map[imgui.MouseButton]events.MouseButton{
		imgui.MouseButtonLeft:   events.MouseButtonLeft,
		imgui.MouseButtonRight:  events.MouseButtonRight,
		imgui.MouseButtonMiddle: events.MouseButtonMiddle,
	}
	var clickType events.ComponentEventType
	for ibt, ebt := range buttonTypes {
		if imgui.IsItemClickedV(ibt) {
			button = ebt
			clickType = events.ComponentClickedEvent
			clickSet = true

			if imgui.IsMouseDoubleClicked(ibt) {
				clickType = events.ComponentDoubleClickedEvent
				clickSet = true
			}
		}
	}

	if !c.hoverAnimationsDisabled {
		if c.state.HasState(events.ItemStateHoverIn) {
			if c.bgHoverColor.X != 0 ||
				c.bgHoverColor.Y != 0 ||
				c.bgHoverColor.Z != 0 ||
				c.bgHoverColor.W != 0 {

				c.StartVec4Animation(
					CmdSetBgColor,
					c.bgHoverColor,
					animation.DefaultColorFadeDuration,
					animation.DefaultEasingFunction,
				)
			}
		}
		if c.state.HasState(events.ItemStateHoverOut) {
			c.StartVec4Animation(
				CmdSetBgColor,
				c.bgColor,
				animation.DefaultColorFadeDuration,
				animation.DefaultEasingFunction)
		}
	}

	if c.state.HasState(events.ItemStateHovered) {
		if c.clickable {
			imgui.SetMouseCursor(imgui.MouseCursorHand)
		} else {
			hasHoverColor := c.bgHoverColor.X != 0 ||
				c.bgHoverColor.Y != 0 ||
				c.bgHoverColor.Z != 0 ||
				c.bgHoverColor.W != 0

			if hasHoverColor && !c.hoverAnimationsDisabled {
				imgui.SetMouseCursor(imgui.MouseCursorHand)
			}
		}
	}

	// Check for a click event
	if clickSet {
		eventbus.Bus.Publish(events.MouseEventRecord{
			EventType: clickType,
			ImguiID:   c.ID(),
			UUID:      c.UUID(),
			Button:    button,
			State:     c.state,
			Data:      c.Data(),
		})
	}

	// Check for a hover-in event
	if c.state.HasState(events.ItemStateHoverIn) {
		eventbus.Bus.Publish(events.MouseEventRecord{
			EventType: events.ComponentHoverInEvent,
			ImguiID:   c.ID(),
			UUID:      c.UUID(),
			State:     c.state,
			Data:      c.Data(),
		})
	}

	// Check for a hover-out event
	if c.state.HasState(events.ItemStateHoverOut) {
		eventbus.Bus.Publish(events.MouseEventRecord{
			EventType: events.ComponentHoverOutEvent,
			ImguiID:   c.ID(),
			UUID:      c.UUID(),
			State:     c.state,
			Data:      c.Data(),
		})
	}
}

// HandleDropTarget is a helper for child components to use inside imgui.BeginDragDropTarget
func (c *Component[T]) HandleDropTarget(dropDataType string) (data interface{}, ok bool) {
	payload := imgui.AcceptDragDropPayload(dropDataType)

	// Check for nil payload first
	if payload.CData == nil {
		return nil, false
	}

	expectedSize := int32(unsafe.Sizeof(imgui.ID(0)))
	if payload.DataSize() != expectedSize {
		return nil, false
	}

	idPtr := (*imgui.ID)(unsafe.Pointer(payload.Data()))

	if idPtr == nil {
		// Payload data is invalid
		log.Error(
			"DragDrop payload data pointer was nil after cast",
			zap.String("expectedType", dropDataType),
		)
		return nil, false
	}

	sourceID := *idPtr

	return dragdrop.GetData(sourceID)
}

func (c *Component[T]) HandleGlobalUpdate(cmd UpdateCmd) bool {
	handled := true

	switch cmd.Type {
	case CmdSetBgActiveColor:
		c.bgActiveColor = cmd.Data.(imgui.Vec4)

	case CmdSetBgColor:
		if color, ok := cmd.Data.(imgui.Vec4); ok {
			c.bgColor = color
			if _, exists := c.animState[CmdSetBgColor]; !exists {
				c.animState[CmdSetBgColor] = &animStateData{
					currentVal:  color,
					isAnimating: false,
					easingFn:    animation.EaseOutQuad,
				}
			}
		} else {
			handled = false
			log.Warn("Invalid data type for CmdSetBgColor",
				zap.String("expectedType", "imgui.Vec4"),
				zap.String("actualType", fmt.Sprintf("%T", cmd.Data)),
			)
		}

	case CmdSetBgHoveredColor:
		c.bgHoverColor = cmd.Data.(imgui.Vec4)
	case CmdSetBgSelectedColor:
		c.bgSelectedColor = cmd.Data.(imgui.Vec4)
	case CmdSetFgColor:
		c.fgColor = cmd.Data.(imgui.Vec4)
		c.textColor = cmd.Data.(imgui.Vec4)
	case CmdSetBorderColor:
		c.borderColor = cmd.Data.(imgui.Vec4)
	case CmdSetBorderSize:
		c.borderSize = cmd.Data.(float32)
	case CmdSetCollapsed:
		c.collapsed = cmd.Data.(bool)
	case CmdSetProgress:
		v := cmd.Data.(float32)
		if v < 0 {
			v = 0
		} else if v > 1 {
			v = 1
		}
		c.progress = v
	case CmdSetProgressBgColor:
		c.progressBg = cmd.Data.(imgui.Vec4)
	case CmdSetProgressFgColor:
		c.progressFg = cmd.Data.(imgui.Vec4)
	case CmdSetProgressHeight:
		c.progressHeight = cmd.Data.(float32)
	case CmdSetRounding:
		c.rounding = cmd.Data.(float32)
	case CmdSetPadding:
		c.padding = cmd.Data.(float32)
	case CmdSetWidth:
		c.width = cmd.Data.(float32)
	case CmdSetHeight:
		c.height = cmd.Data.(float32)
	case CmdSetSize:
		c.size = cmd.Data.(imgui.Vec2)
	case CmdSetEditing:
		c.editing = cmd.Data.(bool)
	case CmdSetEnabled:
		c.enabled = cmd.Data.(bool)
	case CmdSetSelected:
		c.selected = cmd.Data.(bool)
	case CmdSetText:
		c.text = cmd.Data.(string)
	case CmdSetTextColor:
		c.textColor = cmd.Data.(imgui.Vec4)
	case CmdSetOutline:
		c.outline = cmd.Data.(bool)
	case CmdSetOutlineColor:
		c.outlineColor = cmd.Data.(imgui.Vec4)
	case CmdSetOutlineWidth:
		c.outlineWidth = cmd.Data.(float32)

	case CmdSetLoading:
		c.loading = cmd.Data.(bool)
	case CmdSetDragDropData:
		payload := cmd.Data.(dragdrop.DataPayload)
		c.dragDropType = payload.Type
		c.dragDropData = payload.Data
	case CmdSetDragDropTooltipFn:
		if cmd.Data == nil {
			c.dragDropTooltipFn = nil
		} else {
			c.dragDropTooltipFn = cmd.Data.(dragdrop.TooltipFunc)
		}
	default:
		handled = false
	}
	return handled
}

// Destroy is the base cleanup method.
func (c *Component[T]) Destroy() {
	// *crickets*
}

func (c *Component[T]) Layout() {
	panic("not implemented")
}

func (c *Component[T]) Build() {
	// Apply state changes / start animations
	c.ProcessUpdates()

	// Calculate current interpolated values
	c.ProcessAnimations()

	if c.layoutBuilder != nil {
		c.layoutBuilder.Layout()
	} else {
		c.Layout()
	}

	c.handleMouseEvents()
}
