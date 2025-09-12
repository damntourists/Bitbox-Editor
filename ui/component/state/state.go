package state

type ItemState int32

func (s ItemState) HasState(state ItemState) bool { return s&state == state }

const (
	ItemStateNone      ItemState = 0
	ItemStateHovered   ItemState = 1
	ItemStateHoverIn   ItemState = 2
	ItemStateHoverOut  ItemState = 4
	ItemStateActive    ItemState = 8
	ItemStateActiveIn  ItemState = 16
	ItemStateActiveOut ItemState = 32
	ItemStateFocused   ItemState = 64
	ItemStateFocusIn   ItemState = 128
	ItemStateFocusOut  ItemState = 256
)
