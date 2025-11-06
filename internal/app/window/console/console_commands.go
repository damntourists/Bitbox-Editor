package console

import "bitbox-editor/internal/app/window"

type UpdateCmd = window.UpdateCmd
type UpdateCmdType = window.UpdateCmdType

type localCommand int

const (
	cmdConsoleAddLog localCommand = iota
)
