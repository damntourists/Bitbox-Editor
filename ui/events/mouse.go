package events

type MouseButton int32

func (b MouseButton) Int() int32 {
	return int32(b)
}

const (
	MouseButtonLeft   MouseButton = 0
	MouseButtonRight  MouseButton = 1
	MouseButtonMiddle MouseButton = 2
	//MouseButtonCOUNT  MouseButton = 5
)

type ClickType int32

func (c ClickType) Int() int32 {
	return int32(c)
}

const (
	Clicked       ClickType = 0
	DoubleClicked ClickType = 1
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
	MouseSourceCOUNT       MouseSource = 3
)

type MouseEventRecord struct {
	ID     string
	Type   ClickType
	Button MouseButton
	State  State
	GUID   string
	Data   interface{}
}
