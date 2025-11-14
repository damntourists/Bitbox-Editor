package presetlist

type localCommand int

const (
	cmdPresetListSetLocation localCommand = iota
	cmdPresetListSetLoading
	cmdPresetListUpdateList
	cmdPresetListSetSelected
	cmdHandleRowClick
)
