package pad

type localCommand int

const (
	cmdSetPadTextLines localCommand = iota
	cmdSetPadWaveDisplayData
	cmdSetPadCellDisplayData
)
