package main

import (
	"bitbox-editor/internal/app"
	"runtime"

	_ "github.com/silbinarywolf/preferdiscretegpu"
)

func init() {
	runtime.LockOSThread()
}

// Main entrypoint for the BitBox Editor.
func main() {

	editor := app.NewBitboxEditor()
	editor.Run()
}
