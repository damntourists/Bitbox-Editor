package pad_config

type localCommand int

const (
	cmdSetPadConfigPad localCommand = iota
	cmdSetPadConfigPreset
)
