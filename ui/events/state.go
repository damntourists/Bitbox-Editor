package events

type State int32

func (s State) HasState(state State) bool { return s&state == state }

const (
	StateNone      State = 0
	StateHovered   State = 1
	StateHoverIn   State = 2
	StateHoverOut  State = 4
	StateActive    State = 8
	StateActiveIn  State = 16
	StateActiveOut State = 32
	StateFocused   State = 64
	StateFocusIn   State = 128
	StateFocusOut  State = 256
)
