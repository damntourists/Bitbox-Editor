package window

type GlobalCommand int

const (
	// Window commands
	CmdWinSetTitle GlobalCommand = iota
	CmdWinSetIcon
	CmdWinSetSuffix
	CmdWinSetNoClose
	CmdWinSetCollapsed
	CmdWinSetFlags
	CmdWinSetOpen
	CmdWinDestroy
	CmdSetLoading
)
