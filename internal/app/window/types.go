package window

import (
	cmp "bitbox-editor/internal/app/component"
)

type UpdateCmd = cmp.UpdateCmd
type UpdateCmdType = cmp.UpdateCmdType
type UpdateHandlerFunc = cmp.UpdateHandlerFunc

type WindowType interface {
	Menu()
	Layout()
	Build()
	Destroy()
}

type WindowLayoutBuilder interface {
	Menu()
	Layout()
	Style() func()
	Destroy()
}
