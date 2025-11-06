package app

import (
	"bitbox-editor/internal/app/component"
	"bitbox-editor/internal/app/window/presetedit"
	"bitbox-editor/internal/preset"
)

type UpdateCmd = component.UpdateCmd
type UpdateCmdType = component.UpdateCmdType
type UpdateHandlerFunc = component.UpdateHandlerFunc

type localCommand int

const (
	cmdEditorCreate localCommand = iota
	cmdEditorAdd
	cmdEditorRemove
)

type editorCreatePayload struct {
	Preset *preset.Preset
}

type editorAddPayload struct {
	Editor *presetedit.PresetEditWindow
}

type editorRemovePayload struct {
	Editor *presetedit.PresetEditWindow
}
