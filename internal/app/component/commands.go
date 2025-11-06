package component

// type UpdateCmdType int
type UpdateCmdType = interface{}

type UpdateHandlerFunc func(cmd UpdateCmd)

type UpdateCmd struct {
	Type UpdateCmdType
	Data interface{}
}

type GlobalCommand int

// Global component update commands
const (
	CmdSetBgActiveColor GlobalCommand = iota
	CmdSetBgColor
	CmdSetBgHoveredColor
	CmdSetBgSelectedColor
	CmdSetFgColor
	CmdSetBorderColor
	CmdSetBorderSize
	CmdSetCollapsed
	CmdSetDragDropData
	CmdSetDragDropTooltipFn
	CmdSetEditing
	CmdSetEnabled
	CmdSetOutline
	CmdSetOutlineColor
	CmdSetOutlineWidth
	CmdSetPadding
	CmdSetProgress
	CmdSetProgressBgColor
	CmdSetProgressFgColor
	CmdSetProgressHeight
	CmdSetRounding
	CmdSetSelected
	CmdSetSize
	CmdSetText
	CmdSetTextColor
	CmdSetWidth
	CmdSetLoading
	CmdSetTooltip
	CmdSetVolume
	CmdSetMuted
	CmdSetHeight
	CmdInternalSetWaveFile
	CmdInternalSetLoadFailed
	CmdInternalSetCache
)
