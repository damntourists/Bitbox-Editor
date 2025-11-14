package padgrid

type localCommand int

const (
	cmdSetPadGridPreset localCommand = iota
	cmdSetPadGridConfig
	cmdSetPadGridSelectedPad
	cmdSetPadGridPadSize
	cmdHandlePadClick // For translating events
)
