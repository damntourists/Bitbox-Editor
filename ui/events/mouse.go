package events

import (
	"bitbox-editor/ui/component/state"

	"github.com/AllenDang/cimgui-go/imgui"
)

type MouseButton int32

func (b MouseButton) Int() int32 {
	return int32(b)
}

const (
	MouseButtonNone   MouseButton = 0
	MouseButtonLeft   MouseButton = 1
	MouseButtonRight  MouseButton = 2
	MouseButtonMiddle MouseButton = 3
)

type ClickType int32

func (c ClickType) Int() int32 {
	return int32(c)
}

const (
	NoClick       ClickType = 0
	Clicked       ClickType = 1
	DoubleClicked ClickType = 2
)

// MouseSource - Enumeration for AddMouseSourceEvent() actual source of Mouse Input data.
type MouseSource int32

func (s MouseSource) Int() int32 {
	return int32(s)
}

const (
	MouseSourceMouse       MouseSource = 0
	MouseSourceTouchScreen MouseSource = 1
	MouseSourcePen         MouseSource = 2
)

type MouseEventRecord struct {
	ID     imgui.ID
	Type   ClickType
	Button MouseButton
	State  state.ItemState
	Data   interface{}
}
